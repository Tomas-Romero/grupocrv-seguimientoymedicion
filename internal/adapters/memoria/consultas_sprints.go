package memoria

import (
	"context"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Base cumple los puertos de lectura de las pantallas de sprints (T-015). ObtenerSprint y
// ExisteProyecto ya existen, de US-009 y US-005.
var (
	_ app.RepositorioListarSprints   = (*Base)(nil)
	_ app.RepositorioConsultarSprint = (*Base)(nil)
)

// itemsAsignados devuelve copias de las historias del proyecto que tienen una fila de la
// relacion con ese sprint, sin importar su resultado, cada una con el resultado de su fila. Se
// llama con mu tomado.
func (b *Base) itemsAsignados(proyectoID, sprintID string) []app.HistoriaAsignada {
	items := []app.HistoriaAsignada{}
	for _, item := range b.items[proyectoID] {
		for _, a := range b.asignaciones[item.ID] {
			if a.sprintID == sprintID {
				items = append(items, app.HistoriaAsignada{Item: copiaDe(item), Resultado: sprint.Resultado(a.resultado)})
				break
			}
		}
	}
	return items
}

// ListarSprintsConItems devuelve los sprints del proyecto con las historias que se les
// asignaron, sin orden. Un proyecto que no existe devuelve una lista vacia.
func (b *Base) ListarSprintsConItems(_ context.Context, proyectoID string) ([]app.SprintConItems, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	sprints := make([]app.SprintConItems, 0, len(b.sprints[proyectoID]))
	for _, s := range b.sprints[proyectoID] {
		sprints = append(sprints, app.SprintConItems{Sprint: s, Items: b.itemsAsignados(proyectoID, s.ID)})
	}
	return sprints, nil
}

// ItemsDelSprint devuelve las historias que se asignaron al sprint, sin orden. Un sprint que
// no existe devuelve una lista vacia.
func (b *Base) ItemsDelSprint(_ context.Context, sprintID string) ([]app.HistoriaAsignada, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s, ok := b.sprintPorID(sprintID)
	if !ok {
		return []app.HistoriaAsignada{}, nil
	}
	return b.itemsAsignados(s.ProyectoID, s.ID), nil
}

// ListarItemsDisponibles devuelve las historias del proyecto que no estan completadas y no
// estan en ningun sprint abierto, sin orden. Un proyecto que no existe devuelve una lista vacia.
func (b *Base) ListarItemsDisponibles(_ context.Context, proyectoID string) ([]backlog.ItemBacklog, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	items := []backlog.ItemBacklog{}
	for _, item := range b.items[proyectoID] {
		if item.Estado != backlog.EstadoCompletado && !b.enSprintAbierto(item.ID) {
			items = append(items, copiaDe(item))
		}
	}
	return items, nil
}
