package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// RepositorioEdicion es lo que el caso de uso de US-006 necesita de la
// persistencia.
type RepositorioEdicion interface {
	// ObtenerItem devuelve el item y si existe. Un ID mal formado es un item
	// inexistente, no un error.
	ObtenerItem(ctx context.Context, itemID string) (backlog.ItemBacklog, bool, error)

	// GuardarEdicion guarda solo el titulo, la descripcion, la prioridad y los
	// criterios del item (RN-006-8). Si el item ya no existe, devuelve
	// ErrItemInexistente; si se completo despues de leerlo, devuelve
	// backlog.ErrItemCompletado y no lo cambia.
	GuardarEdicion(ctx context.Context, item backlog.ItemBacklog) error
}

// EditarItem es el caso de uso de US-006: editar los datos de un item del
// Product Backlog, incluida su prioridad (repriorizar).
type EditarItem struct {
	repo RepositorioEdicion
}

// NuevoEditarItem arma el caso de uso.
func NuevoEditarItem(repo RepositorioEdicion) *EditarItem {
	return &EditarItem{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): carga el item, lo edita con el
// dominio y guarda la edicion. Si el item no existe responde
// ErrItemInexistente sin validar los datos (RN-006-7). Devuelve el item ya
// editado; si algo falla no guarda nada (RN-006-10).
func (e *EditarItem) Ejecutar(ctx context.Context, itemID string, datos backlog.DatosItem) (backlog.ItemBacklog, error) {
	item, existe, err := e.repo.ObtenerItem(ctx, itemID)
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("consultar el item %s: %w", itemID, err)
	}
	if !existe {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", ErrItemInexistente, itemID)
	}

	editado, err := item.Editar(datos)
	if err != nil {
		// Los errores del dominio se devuelven tal cual: su mensaje es el que ve
		// la persona (spec US-006, seccion 7).
		return backlog.ItemBacklog{}, err
	}

	if err := e.repo.GuardarEdicion(ctx, editado); err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("guardar la edicion del item %s: %w", itemID, err)
	}
	return editado, nil
}
