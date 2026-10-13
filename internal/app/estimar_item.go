package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// RepositorioEstimacion es lo que el caso de uso de US-013 necesita de la
// persistencia.
type RepositorioEstimacion interface {
	// ObtenerItem devuelve el item y si existe. Un ID mal formado es un item
	// inexistente, no un error.
	ObtenerItem(ctx context.Context, itemID string) (backlog.ItemBacklog, bool, error)

	// GuardarEstimacion guarda solo los Story Points del item (RN-013-7). Si el
	// item ya no existe, devuelve ErrItemInexistente.
	GuardarEstimacion(ctx context.Context, itemID string, puntos backlog.StoryPoints) error
}

// EstimarItem es el caso de uso de US-013: asignarle Story Points de la escala
// Fibonacci a un item del Product Backlog.
type EstimarItem struct {
	repo RepositorioEstimacion
}

// NuevoEstimarItem arma el caso de uso.
func NuevoEstimarItem(repo RepositorioEstimacion) *EstimarItem {
	return &EstimarItem{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): carga el item, lo estima con
// el dominio y guarda la estimacion. Si el item no existe responde
// ErrItemInexistente sin mirar el valor (RN-013-6). Devuelve el item ya
// estimado; si algo falla no guarda nada (RN-013-8).
func (e *EstimarItem) Ejecutar(ctx context.Context, itemID string, puntos int) (backlog.ItemBacklog, error) {
	item, existe, err := e.repo.ObtenerItem(ctx, itemID)
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("consultar el item %s: %w", itemID, err)
	}
	if !existe {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", ErrItemInexistente, itemID)
	}

	estimado, err := item.Estimar(puntos)
	if err != nil {
		// Los errores del dominio se devuelven tal cual: su mensaje es el que ve
		// la persona (spec US-013, seccion 7).
		return backlog.ItemBacklog{}, err
	}

	if err := e.repo.GuardarEstimacion(ctx, itemID, estimado.StoryPoints); err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("guardar la estimacion del item %s: %w", itemID, err)
	}
	return estimado, nil
}
