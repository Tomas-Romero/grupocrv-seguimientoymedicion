package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// DefectoEnProyecto devuelve el defecto guardado y dice si existe y es de ese proyecto (RN-023-7).
// Uno de otro proyecto no cuenta, y un ID vacio o mal formado es inexistente, no un error.
func (r *RepositorioDefectos) DefectoEnProyecto(ctx context.Context, proyectoID, defectoID string) (defecto.Defecto, bool, error) {
	if !formatoUUID.MatchString(proyectoID) || !formatoUUID.MatchString(defectoID) {
		return defecto.Defecto{}, false, nil
	}
	const consulta = `
SELECT id::text, proyecto_id::text, item_id::text, sprint_deteccion_id::text,
       COALESCE(sprint_resolucion_id::text, ''), descripcion, severidad, estado
FROM defectos WHERE id = $1 AND proyecto_id = $2`

	var d defecto.Defecto
	var severidad, estado string
	err := r.pool.QueryRow(ctx, consulta, defectoID, proyectoID).Scan(&d.ID, &d.ProyectoID, &d.ItemID,
		&d.SprintDeteccionID, &d.SprintResolucionID, &d.Descripcion, &severidad, &estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return defecto.Defecto{}, false, nil
	}
	if err != nil {
		return defecto.Defecto{}, false, fmt.Errorf("consultar el defecto %s del proyecto %s: %w", defectoID, proyectoID, err)
	}
	d.Severidad, d.Estado = defecto.Severidad(severidad), defecto.Estado(estado)
	return d, true, nil
}

// NumeroDeSprintEnProyecto devuelve el numero (sprint.Numero) del sprint y dice si existe y es de
// ese proyecto (RN-023-7, RN-023-8). Uno de otro proyecto no cuenta, y un ID vacio o mal formado es
// inexistente. El numero no se copia al defecto: se lee de sprints (RN-022-6).
func (r *RepositorioDefectos) NumeroDeSprintEnProyecto(ctx context.Context, proyectoID, sprintID string) (int, bool, error) {
	if !formatoUUID.MatchString(proyectoID) || !formatoUUID.MatchString(sprintID) {
		return 0, false, nil
	}
	var numero int
	err := r.pool.QueryRow(ctx, `SELECT numero FROM sprints WHERE id = $1 AND proyecto_id = $2`, sprintID, proyectoID).Scan(&numero)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("consultar el sprint %s del proyecto %s: %w", sprintID, proyectoID, err)
	}
	return numero, true, nil
}

// ActualizarEstadoDefecto guarda el estado y el sprint de resolucion con un unico UPDATE
// condicionado al estado que leyo el caso de uso (RN-023-11). Un sprint de resolucion vacio se
// guarda como NULL (RN-023-5).
func (r *RepositorioDefectos) ActualizarEstadoDefecto(ctx context.Context, d defecto.Defecto, estadoAnterior defecto.Estado) error {
	const actualizar = `
UPDATE defectos SET estado = $3, sprint_resolucion_id = $4
WHERE id = $1 AND proyecto_id = $2 AND estado = $5`

	var sprintResolucion *string
	if d.SprintResolucionID != "" {
		sprintResolucion = &d.SprintResolucionID
	}
	if _, err := r.pool.Exec(ctx, actualizar, d.ID, d.ProyectoID, string(d.Estado), sprintResolucion, string(estadoAnterior)); err != nil {
		return fmt.Errorf("actualizar el estado del defecto %s: %w", d.ID, err)
	}
	return nil
}
