package memoria

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// Base cumple el puerto del caso de uso de US-013.
var _ app.RepositorioEstimacion = (*Base)(nil)

// ObtenerItem busca el item por su ID en todos los proyectos. Cualquier ID que no
// este, incluido uno mal formado, es un item inexistente y no un error.
func (b *Base) ObtenerItem(_ context.Context, itemID string) (backlog.ItemBacklog, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, items := range b.items {
		for _, item := range items {
			if item.ID == itemID {
				return copiaDe(item), true, nil
			}
		}
	}
	return backlog.ItemBacklog{}, false, nil
}

// GuardarEstimacion cambia solo los Story Points del item. Si el item no esta en
// la base, devuelve app.ErrItemInexistente y no cambia nada.
func (b *Base) GuardarEstimacion(_ context.Context, itemID string, puntos backlog.StoryPoints) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, items := range b.items {
		for i := range items {
			if items[i].ID == itemID {
				items[i].StoryPoints = puntos
				return nil
			}
		}
	}
	return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, itemID)
}
