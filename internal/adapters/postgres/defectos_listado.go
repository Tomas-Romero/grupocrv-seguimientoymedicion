package postgres

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// RepositorioDefectos cumple tambien el puerto de la lectura de defectos de T-021 (ExisteProyecto
// es el de US-022).
var _ app.RepositorioListarDefectos = (*RepositorioDefectos)(nil)

// ListarDefectosDelProyecto devuelve los defectos del proyecto, de cualquier estado y sin ordenar:
// el orden lo pone el caso de uso (RN-T021-4). Una sola consulta trae la historia y los numeros de
// sprint con JOIN, sin una consulta por fila (spec T-021, seccion 5). El de resolucion es un LEFT
// JOIN porque solo el defecto resuelto lo tiene; sin el, SprintResolucion queda en 0. Un
// proyecto sin defectos, o un ID mal formado, devuelve una lista vacia.
func (r *RepositorioDefectos) ListarDefectosDelProyecto(ctx context.Context, proyectoID string) ([]app.DefectoListado, error) {
	if !formatoUUID.MatchString(proyectoID) {
		return []app.DefectoListado{}, nil
	}
	const consulta = `
SELECT d.id::text, d.proyecto_id::text, d.item_id::text, d.sprint_deteccion_id::text,
       COALESCE(d.sprint_resolucion_id::text, ''), d.descripcion, d.severidad, d.estado,
       i.numero, i.titulo, sd.numero, COALESCE(sr.numero, 0), d.creado_en
FROM defectos d
JOIN items_backlog i ON i.id = d.item_id
JOIN sprints sd ON sd.id = d.sprint_deteccion_id
LEFT JOIN sprints sr ON sr.id = d.sprint_resolucion_id
WHERE d.proyecto_id = $1`
	filas, err := r.pool.Query(ctx, consulta, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar los defectos del proyecto %s: %w", proyectoID, err)
	}
	defer filas.Close()

	listados := []app.DefectoListado{}
	for filas.Next() {
		var (
			l                 app.DefectoListado
			severidad, estado string
		)
		err := filas.Scan(&l.Defecto.ID, &l.Defecto.ProyectoID, &l.Defecto.ItemID, &l.Defecto.SprintDeteccionID,
			&l.Defecto.SprintResolucionID, &l.Defecto.Descripcion, &severidad, &estado,
			&l.HistoriaNumero, &l.HistoriaTitulo, &l.SprintDeteccion, &l.SprintResolucion, &l.CreadoEn)
		if err != nil {
			return nil, fmt.Errorf("leer un defecto: %w", err)
		}
		l.Defecto.Severidad, l.Defecto.Estado = defecto.Severidad(severidad), defecto.Estado(estado)
		listados = append(listados, l)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar los defectos del proyecto %s: %w", proyectoID, err)
	}
	return listados, nil
}
