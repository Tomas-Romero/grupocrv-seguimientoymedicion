package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
)

// RepositorioPlanningPoker cumple tambien el puerto de la lectura de la pantalla de la sesion
// (T-020).
var _ app.RepositorioConsultaDeSesion = (*RepositorioPlanningPoker)(nil)

// SesionesAbiertas devuelve, para cada historia del proyecto con una sesion abierta, el ID de esa
// sesion. Un ID mal formado no es de ningun proyecto: no tiene sesiones.
func (r *RepositorioPlanningPoker) SesionesAbiertas(ctx context.Context, proyectoID string) (map[string]string, error) {
	const consulta = `
SELECT s.item_id::text, s.id::text
FROM sesiones_planning_poker s JOIN items_backlog i ON i.id = s.item_id
WHERE i.proyecto_id = $1 AND s.estado <> 'finalizada'`
	abiertas := map[string]string{}
	filas, err := r.pool.Query(ctx, consulta, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("leer las sesiones abiertas del proyecto %s: %w", proyectoID, err)
	}
	var item, sesion string
	if _, err := pgx.ForEachRow(filas, []any{&item, &sesion}, func() error {
		abiertas[item] = sesion
		return nil
	}); err != nil {
		if esIDMalFormado(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("leer las sesiones abiertas del proyecto %s: %w", proyectoID, err)
	}
	return abiertas, nil
}
