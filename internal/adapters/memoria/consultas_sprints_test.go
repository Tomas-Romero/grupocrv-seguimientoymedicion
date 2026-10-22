package memoria_test

import (
	"context"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// escenarioDeSprints arma un proyecto con un sprint y dos historias listas, una de ellas
// asignada al sprint.
type escenarioDeSprints struct {
	base       *memoria.Base
	proyectoID string
	sprint     sprint.Sprint
	asignada   backlog.ItemBacklog
	libre      backlog.ItemBacklog
}

func historiaLista(t *testing.T, base *memoria.Base, proyectoID, titulo string) backlog.ItemBacklog {
	t.Helper()
	puntos, err := backlog.NuevosStoryPoints(5)
	if err != nil {
		t.Fatalf("NuevosStoryPoints: %v", err)
	}
	item, err := base.RegistrarItem(context.Background(), proyectoID, backlog.ItemBacklog{
		Titulo: titulo, Prioridad: backlog.PrioridadMust, Estado: backlog.EstadoPendiente, StoryPoints: puntos, Criterios: []string{"un criterio"},
	})
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	return item
}

func nuevoEscenarioDeSprints(t *testing.T) escenarioDeSprints {
	t.Helper()
	ctx := context.Background()
	base := memoria.NuevaBase()
	inicio := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	p := registrarProyectoDePrueba(t, base, "Demo")
	s, err := base.RegistrarSprint(ctx, sprint.Sprint{ProyectoID: p.ID, SprintGoal: "Meta", FechaInicio: inicio, FechaFin: inicio.AddDate(0, 0, 13), Estado: sprint.EstadoAbierto})
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}
	asignada := historiaLista(t, base, p.ID, "Asignada")
	libre := historiaLista(t, base, p.ID, "Libre")
	if err := base.AsignarItemASprint(ctx, asignada.ID, s.ID); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	return escenarioDeSprints{base: base, proyectoID: p.ID, sprint: s, asignada: asignada, libre: libre}
}

// T-015 / RN-T015-2: ListarSprintsConItems devuelve cada sprint con las historias que se le asignaron.
func TestBase_ListarSprintsConItems(t *testing.T) {
	e := nuevoEscenarioDeSprints(t)

	got, err := e.base.ListarSprintsConItems(context.Background(), e.proyectoID)
	if err != nil {
		t.Fatalf("ListarSprintsConItems: %v", err)
	}
	if len(got) != 1 || got[0].Sprint.ID != e.sprint.ID {
		t.Fatalf("sprints = %+v, quiero el sprint del escenario", got)
	}
	if len(got[0].Items) != 1 || got[0].Items[0].ID != e.asignada.ID {
		t.Errorf("historias del sprint = %+v, quiero solo la asignada", got[0].Items)
	}

	vacio, err := e.base.ListarSprintsConItems(context.Background(), "no-existe")
	if err != nil || len(vacio) != 0 {
		t.Errorf("proyecto inexistente: %v, %v; quiero lista vacia y sin error", vacio, err)
	}
}

// ItemsDelSprint devuelve las historias del sprint; un sprint que no existe, una lista vacia.
func TestBase_ItemsDelSprint(t *testing.T) {
	e := nuevoEscenarioDeSprints(t)

	got, err := e.base.ItemsDelSprint(context.Background(), e.sprint.ID)
	if err != nil || len(got) != 1 || got[0].ID != e.asignada.ID {
		t.Errorf("ItemsDelSprint = %+v, %v; quiero la historia asignada", got, err)
	}
	vacio, err := e.base.ItemsDelSprint(context.Background(), "no-existe")
	if err != nil || len(vacio) != 0 {
		t.Errorf("sprint inexistente: %v, %v; quiero lista vacia y sin error", vacio, err)
	}

	got[0].Criterios[0] = "cambiado"
	if guardados, _ := e.base.ItemsDelSprint(context.Background(), e.sprint.ID); guardados[0].Criterios[0] != "un criterio" {
		t.Error("cambiar lo devuelto cambio lo guardado")
	}
}

// ListarItemsDisponibles ofrece las historias que no estan completadas ni en un sprint abierto
// (RN-T015-6, CL-T015-5).
func TestBase_ListarItemsDisponibles(t *testing.T) {
	e := nuevoEscenarioDeSprints(t)
	ctx := context.Background()
	if _, err := e.base.RegistrarItem(ctx, e.proyectoID, backlog.ItemBacklog{Titulo: "Completada", Prioridad: backlog.PrioridadMust, Estado: backlog.EstadoCompletado, CompletadoEn: time.Now()}); err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	if _, err := e.base.GuardarCambioDeEstado(ctx, e.asignada.ID, e.sprint.ID, func(i backlog.ItemBacklog) backlog.ItemBacklog { return i.Completar(time.Now()) }); err != nil {
		t.Fatalf("marcar la asignada como completada: %v", err)
	}

	got, err := e.base.ListarItemsDisponibles(ctx, e.proyectoID)
	if err != nil {
		t.Fatalf("ListarItemsDisponibles: %v", err)
	}
	if len(got) != 1 || got[0].ID != e.libre.ID {
		t.Errorf("disponibles = %+v, quiero solo la historia libre", got)
	}
	if vacio, err := e.base.ListarItemsDisponibles(ctx, "no-existe"); err != nil || len(vacio) != 0 {
		t.Errorf("proyecto inexistente: %v, %v; quiero lista vacia y sin error", vacio, err)
	}
}

var (
	_ app.RepositorioListarSprints   = (*memoria.Base)(nil)
	_ app.RepositorioConsultarSprint = (*memoria.Base)(nil)
)
