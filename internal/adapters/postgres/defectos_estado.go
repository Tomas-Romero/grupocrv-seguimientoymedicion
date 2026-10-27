package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// claveForaneaDefectoSprintResolucion es la clave foranea de sprint_resolucion_id (el nombre que
// Postgres genera por defecto), la que se traduce a ErrSprintInexistente.
const claveForaneaDefectoSprintResolucion = "defectos_sprint_resolucion_id_fkey"

// RepositorioDefectos cumple tambien el puerto del caso de uso de US-023 (ExisteProyecto es el de
// US-022).
var _ app.RepositorioCambioDeEstadoDefecto = (*RepositorioDefectos)(nil)

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
//
// Si no actualiza ninguna fila relee: si el defecto no esta en el proyecto devuelve
// ErrDefectoInexistente; si esta, otro cambio se adelanto y devuelve ErrDefectoModificado sin
// guardar nada. No hay transaccion ni bloqueo: con READ COMMITTED el segundo UPDATE espera la fila
// del primero, reevalua el WHERE y no la encuentra. No reintenta: con el estado cambiado, lo pedido
// quiza ya no es valido.
//
// La clave foranea es la defensa final contra un sprint de resolucion que desaparecio despues de
// que el caso de uso lo verifico: se traduce a ErrSprintInexistente. Cualquier otro error, como un
// CHECK, se devuelve envuelto y sin traducir.
func (r *RepositorioDefectos) ActualizarEstadoDefecto(ctx context.Context, d defecto.Defecto, estadoAnterior defecto.Estado) error {
	const actualizar = `
UPDATE defectos SET estado = $3, sprint_resolucion_id = $4
WHERE id = $1 AND proyecto_id = $2 AND estado = $5`

	// Un ID que no es un UUID no puede ser de ninguna fila: Postgres lo rechazaria con un error de
	// sintaxis, asi que se traduce antes.
	if !formatoUUID.MatchString(d.ID) || !formatoUUID.MatchString(d.ProyectoID) {
		return fmt.Errorf("%w (id %s, proyecto %s)", app.ErrDefectoInexistente, d.ID, d.ProyectoID)
	}
	var sprintResolucion *string
	if d.SprintResolucionID != "" {
		if !formatoUUID.MatchString(d.SprintResolucionID) {
			return fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, d.SprintResolucionID)
		}
		sprintResolucion = &d.SprintResolucionID
	}

	etiqueta, err := r.pool.Exec(ctx, actualizar, d.ID, d.ProyectoID, string(d.Estado), sprintResolucion, string(estadoAnterior))
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codigoClaveForaneaViolada && pgErr.ConstraintName == claveForaneaDefectoSprintResolucion {
		return fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, d.SprintResolucionID)
	}
	if err != nil {
		return fmt.Errorf("actualizar el estado del defecto %s: %w", d.ID, err)
	}
	if etiqueta.RowsAffected() == 1 {
		return nil
	}

	var existe bool
	const releer = `SELECT EXISTS (SELECT 1 FROM defectos WHERE id = $1 AND proyecto_id = $2)`
	if err := r.pool.QueryRow(ctx, releer, d.ID, d.ProyectoID).Scan(&existe); err != nil {
		return fmt.Errorf("releer el defecto %s: %w", d.ID, err)
	}
	if !existe {
		return fmt.Errorf("%w (id %s, proyecto %s)", app.ErrDefectoInexistente, d.ID, d.ProyectoID)
	}
	return fmt.Errorf("%w (id %s)", app.ErrDefectoModificado, d.ID)
}
