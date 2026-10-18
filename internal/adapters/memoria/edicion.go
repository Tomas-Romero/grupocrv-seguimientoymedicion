package memoria

import (
	"context"
	"fmt"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// Base cumple el puerto del caso de uso de US-006.
var _ app.RepositorioEdicion = (*Base)(nil)

// GuardarEdicion cambia solo el titulo, la descripcion, la prioridad y los
// criterios del item (RN-006-8): el estado, los Story Points y el numero que
// traiga el item recibido no se escriben. Si el item no esta en la base,
// devuelve app.ErrItemInexistente; si ya esta completado, devuelve
// backlog.ErrItemCompletado. En los dos casos no cambia nada.
func (b *Base) GuardarEdicion(_ context.Context, item backlog.ItemBacklog) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, items := range b.items {
		for i := range items {
			if items[i].ID != item.ID {
				continue
			}
			// Igual que el UPDATE condicionado de Postgres: un item completado
			// no se edita aunque se haya completado despues de leerlo.
			if items[i].Estado == backlog.EstadoCompletado {
				return fmt.Errorf("%w (id %s)", backlog.ErrItemCompletado, item.ID)
			}
			items[i].Titulo = item.Titulo
			items[i].Descripcion = item.Descripcion
			items[i].Prioridad = item.Prioridad
			items[i].Criterios = slices.Clone(item.Criterios)
			return nil
		}
	}
	return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, item.ID)
}
