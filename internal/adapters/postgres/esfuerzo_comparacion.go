package postgres

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
)

// RepositorioEsfuerzo tambien cumple el puerto del caso de uso de US-021.
var _ app.RepositorioCompararEsfuerzo = (*RepositorioEsfuerzo)(nil)

// LeerHistoriasConHoras devuelve todas las historias del proyecto con las horas de todos sus
// registros de esfuerzo, en orden de fecha. Es una sola consulta (un LEFT JOIN de historias a
// registros), asi que todo es de la misma foto y no hay una consulta por historia. Una historia
// sin registros llega con la lista de horas vacia. Un ID mal formado, o un proyecto sin historias,
// devuelve una lista vacia.
func (r *RepositorioEsfuerzo) LeerHistoriasConHoras(ctx context.Context, proyectoID string) ([]app.HistoriaConHoras, error) {
	historias := []app.HistoriaConHoras{}
	if !formatoUUID.MatchString(proyectoID) {
		return historias, nil
	}

	const consulta = `
SELECT ` + columnasDeItem + `, e.horas::float8
FROM items_backlog i
LEFT JOIN registros_esfuerzo e ON e.item_id = i.id
WHERE i.proyecto_id = $1
ORDER BY i.numero, e.fecha, e.creado_en, e.id`
	filas, err := r.pool.Query(ctx, consulta, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("leer las historias con horas del proyecto %s: %w", proyectoID, err)
	}
	defer filas.Close()

	posicion := map[string]int{}
	for filas.Next() {
		var (
			f     filaDeItem
			horas *float64
		)
		if err := filas.Scan(append(f.destinos(), &horas)...); err != nil {
			return nil, fmt.Errorf("leer una historia con horas: %w", err)
		}
		i, vista := posicion[f.item.ID]
		if !vista {
			item, err := f.armar()
			if err != nil {
				return nil, err
			}
			i = len(historias)
			posicion[item.ID] = i
			historias = append(historias, app.HistoriaConHoras{Item: item})
		}
		if horas != nil {
			historias[i].Horas = append(historias[i].Horas, *horas)
		}
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("leer las historias con horas del proyecto %s: %w", proyectoID, err)
	}
	return historias, nil
}
