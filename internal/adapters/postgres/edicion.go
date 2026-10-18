package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// RepositorioBacklog cumple tambien el puerto del caso de uso de US-006.
var _ app.RepositorioEdicion = (*RepositorioBacklog)(nil)

// GuardarEdicion escribe solo el titulo, la descripcion, la prioridad y los
// criterios del item con un unico UPDATE (RN-006-8). El estado, los Story Points
// y el numero no se tocan. Sin bloqueo de filas ni version: si dos ediciones
// llegan a la vez, gana la ultima (RN-006-9).
//
// El UPDATE esta condicionado a que el item no este completado: si se completo
// entre la lectura del caso de uso y este guardado, no actualiza ninguna fila.
// Entonces se relee el item para distinguir los dos motivos, igual que
// GuardarEstimacion: si no existe, o el ID no es un UUID, devuelve
// app.ErrItemInexistente; si esta completado, backlog.ErrItemCompletado.
func (r *RepositorioBacklog) GuardarEdicion(ctx context.Context, item backlog.ItemBacklog) error {
	// La columna es NOT NULL: una lista nula se guarda como arreglo vacio.
	criterios := item.Criterios
	if criterios == nil {
		criterios = []string{}
	}

	const actualizar = `
UPDATE items_backlog SET titulo = $2, descripcion = $3, prioridad = $4, criterios = $5
WHERE id = $1 AND estado <> 'completado'`
	resultado, err := r.pool.Exec(ctx, actualizar, item.ID, item.Titulo, item.Descripcion, string(item.Prioridad), criterios)
	if esIDMalFormado(err) {
		return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, item.ID)
	}
	if err != nil {
		return fmt.Errorf("guardar la edicion del item %s: %w", item.ID, err)
	}
	if resultado.RowsAffected() > 0 {
		return nil
	}

	// Ninguna fila actualizada: o el item no existe o esta completado.
	var estado string
	err = r.pool.QueryRow(ctx, `SELECT estado FROM items_backlog WHERE id = $1`, item.ID).Scan(&estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, item.ID)
	}
	if err != nil {
		return fmt.Errorf("releer el item %s: %w", item.ID, err)
	}
	if backlog.Estado(estado) == backlog.EstadoCompletado {
		return fmt.Errorf("%w (id %s)", backlog.ErrItemCompletado, item.ID)
	}
	// El item existe y no esta completado: no hay motivo para que el UPDATE no
	// haya tocado la fila. Se informa en lugar de devolver un exito falso.
	return fmt.Errorf("no se pudo guardar la edicion del item %s", item.ID)
}
