package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// RepositorioBacklog cumple tambien el puerto del caso de uso de US-013.
var _ app.RepositorioEstimacion = (*RepositorioBacklog)(nil)

// ObtenerItem lee el item completo. Un ID que ningun item tiene, o que no es un
// UUID, es un item inexistente y no un error.
func (r *RepositorioBacklog) ObtenerItem(ctx context.Context, itemID string) (backlog.ItemBacklog, bool, error) {
	const consulta = `
SELECT id::text, proyecto_id::text, numero, titulo, descripcion, prioridad, estado, story_points, criterios, creado_en
FROM items_backlog WHERE id = $1`
	var (
		item        backlog.ItemBacklog
		prioridad   string
		estado      string
		storyPoints *int32
	)
	err := r.pool.QueryRow(ctx, consulta, itemID).Scan(&item.ID, &item.ProyectoID, &item.Numero, &item.Titulo,
		&item.Descripcion, &prioridad, &estado, &storyPoints, &item.Criterios, &item.CreadoEn)
	if errors.Is(err, pgx.ErrNoRows) || esIDMalFormado(err) {
		return backlog.ItemBacklog{}, false, nil
	}
	if err != nil {
		return backlog.ItemBacklog{}, false, fmt.Errorf("leer el item %s: %w", itemID, err)
	}
	item.Prioridad = backlog.Prioridad(prioridad)
	item.Estado = backlog.Estado(estado)

	// El nulo de la columna es "sin estimar" (RN-005-7). Un valor fuera de la
	// escala es un dato corrupto: se informa, no se oculta.
	item.StoryPoints = backlog.SinEstimar()
	if storyPoints != nil {
		item.StoryPoints, err = backlog.NuevosStoryPoints(int(*storyPoints))
		if err != nil {
			return backlog.ItemBacklog{}, false, fmt.Errorf("el item %s tiene una estimacion guardada invalida: %w", itemID, err)
		}
	}
	return item, true, nil
}

// GuardarEstimacion escribe solo la columna story_points con un unico UPDATE
// (RN-013-7). Sin estimar se guarda como nulo, no como 0. Si el item no existe,
// o el ID no es un UUID, devuelve app.ErrItemInexistente.
func (r *RepositorioBacklog) GuardarEstimacion(ctx context.Context, itemID string, puntos backlog.StoryPoints) error {
	var valor any // nil se guarda como NULL
	if p, estimado := puntos.Valor(); estimado {
		valor = p
	}

	resultado, err := r.pool.Exec(ctx, `UPDATE items_backlog SET story_points = $2 WHERE id = $1`, itemID, valor)
	if esIDMalFormado(err) {
		return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, itemID)
	}
	if err != nil {
		return fmt.Errorf("guardar la estimacion del item %s: %w", itemID, err)
	}
	if resultado.RowsAffected() == 0 {
		return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, itemID)
	}
	return nil
}
