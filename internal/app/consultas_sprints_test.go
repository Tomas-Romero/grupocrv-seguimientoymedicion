package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// lecturasDeSprintsFalsas simula las lecturas de sprints sin base de datos.
type lecturasDeSprintsFalsas struct {
	proyectos   map[string]bool
	sprints     []app.SprintConItems
	disponibles []backlog.ItemBacklog
	err         error
}

func (l *lecturasDeSprintsFalsas) ExisteProyecto(_ context.Context, id string) (bool, error) {
	return l.proyectos[id], l.err
}

func (l *lecturasDeSprintsFalsas) ListarSprintsConItems(context.Context, string) ([]app.SprintConItems, error) {
	return append([]app.SprintConItems(nil), l.sprints...), l.err
}

func (l *lecturasDeSprintsFalsas) ObtenerSprint(_ context.Context, id string) (sprint.Sprint, bool, error) {
	if l.err != nil {
		return sprint.Sprint{}, false, l.err
	}
	for _, s := range l.sprints {
		if s.Sprint.ID == id {
			return s.Sprint, true, nil
		}
	}
	return sprint.Sprint{}, false, nil
}

func (l *lecturasDeSprintsFalsas) ItemsDelSprint(_ context.Context, id string) ([]app.HistoriaAsignada, error) {
	for _, s := range l.sprints {
		if s.Sprint.ID == id {
			return append([]app.HistoriaAsignada(nil), s.Items...), l.err
		}
	}
	return nil, l.err
}

func (l *lecturasDeSprintsFalsas) ListarItemsDisponibles(context.Context, string) ([]backlog.ItemBacklog, error) {
	return append([]backlog.ItemBacklog(nil), l.disponibles...), l.err
}

// asignada es una historia en un sprint que sigue abierto para ella: su fila esta pendiente.
func asignada(item backlog.ItemBacklog) app.HistoriaAsignada {
	return app.HistoriaAsignada{Item: item, Resultado: sprint.ResultadoPendiente}
}

// cerrada es una historia de un sprint cerrado, con el resultado que quedo en su fila.
func cerrada(item backlog.ItemBacklog, resultado sprint.Resultado) app.HistoriaAsignada {
	return app.HistoriaAsignada{Item: item, Resultado: resultado}
}

func sprintDe(id string, numero int, proyectoID string) sprint.Sprint {
	return sprint.Sprint{ID: id, Numero: numero, ProyectoID: proyectoID, SprintGoal: "Objetivo " + id, Estado: sprint.EstadoAbierto,
		FechaInicio: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), FechaFin: time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC)}
}

func historiaDe(id string, numero int, prioridad backlog.Prioridad, puntos int, estado backlog.Estado) backlog.ItemBacklog {
	item := backlog.ItemBacklog{ID: id, Numero: numero, ProyectoID: "p1", Titulo: "Historia " + id, Prioridad: prioridad, Estado: estado, StoryPoints: backlog.SinEstimar()}
	if puntos > 0 {
		item.StoryPoints, _ = backlog.NuevosStoryPoints(puntos)
	}
	return item
}

// T-015 / RN-T015-3, RN-T015-8: los sprints salen por numero descendente, con la cantidad de
// historias y los Story Points planificados y completados (los calcula metricas, US-025).
func TestListarSprints(t *testing.T) {
	repo := &lecturasDeSprintsFalsas{
		proyectos: map[string]bool{"p1": true},
		sprints: []app.SprintConItems{
			{Sprint: sprintDe("s1", 1, "p1"), Items: []app.HistoriaAsignada{
				asignada(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado)),
				asignada(historiaDe("b", 2, backlog.PrioridadMust, 3, backlog.EstadoPendiente)),
			}},
			{Sprint: sprintDe("s3", 3, "p1")},
			{Sprint: sprintDe("s2", 2, "p1"), Items: []app.HistoriaAsignada{
				asignada(historiaDe("c", 3, backlog.PrioridadShould, 8, backlog.EstadoCompletado)),
			}},
		},
	}

	got, err := app.NuevoListarSprints(repo).Ejecutar(context.Background(), "p1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if len(got) != 3 || got[0].Sprint.ID != "s3" || got[1].Sprint.ID != "s2" || got[2].Sprint.ID != "s1" {
		t.Fatalf("orden = %+v, quiero s3, s2, s1", got)
	}
	vacio, medio, primero := got[0], got[1], got[2]
	if vacio.Historias != 0 || vacio.Planificados != 0 || vacio.Completados != 0 {
		t.Errorf("sprint sin historias = %+v, quiero todo en 0", vacio)
	}
	if medio.Historias != 1 || medio.Planificados != 8 || medio.Completados != 8 {
		t.Errorf("sprint 2 = %+v, quiero 1 historia, 8 planificados y 8 completados", medio)
	}
	if primero.Historias != 2 || primero.Planificados != 8 || primero.Completados != 5 {
		t.Errorf("sprint 1 = %+v, quiero 2 historias, 8 planificados y 5 completados", primero)
	}
}

// CL-T015-1, CL-T015-2: sin sprints, o con historias sin estimar, no hay error.
func TestListarSprints_VacioYSinEstimar(t *testing.T) {
	repo := &lecturasDeSprintsFalsas{proyectos: map[string]bool{"p1": true}}
	got, err := app.NuevoListarSprints(repo).Ejecutar(context.Background(), "p1")
	if err != nil || len(got) != 0 {
		t.Errorf("sin sprints: %v, %v; quiero lista vacia y sin error", got, err)
	}

	repo.sprints = []app.SprintConItems{{Sprint: sprintDe("s1", 1, "p1"), Items: []app.HistoriaAsignada{asignada(historiaDe("a", 1, backlog.PrioridadMust, 0, backlog.EstadoPendiente))}}}
	got, err = app.NuevoListarSprints(repo).Ejecutar(context.Background(), "p1")
	if err != nil || got[0].Historias != 1 || got[0].Planificados != 0 {
		t.Errorf("historia sin estimar: %+v, %v; quiero 1 historia y 0 planificados", got, err)
	}
}

// RN-T015-10: un proyecto que no existe es app.ErrProyectoInexistente.
func TestListarSprints_ProyectoInexistente(t *testing.T) {
	_, err := app.NuevoListarSprints(&lecturasDeSprintsFalsas{}).Ejecutar(context.Background(), "no-esta")
	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Errorf("Ejecutar() = %v, quiero ErrProyectoInexistente", err)
	}
}

func TestListarSprints_ErrorDeLaBase(t *testing.T) {
	errBase := errors.New("base caida")
	_, err := app.NuevoListarSprints(&lecturasDeSprintsFalsas{err: errBase}).Ejecutar(context.Background(), "p1")
	if !errors.Is(err, errBase) {
		t.Errorf("Ejecutar() = %v, quiero un error que envuelva %v", err, errBase)
	}
}

// RN-T015-4, RN-T015-8, CL-T015-9: el detalle ordena las historias por prioridad y numero, marca las
// completadas, suma los Story Points y ofrece las disponibles en el mismo orden.
func TestConsultarSprint_Detalle(t *testing.T) {
	repo := &lecturasDeSprintsFalsas{
		proyectos: map[string]bool{"p1": true},
		sprints: []app.SprintConItems{{Sprint: sprintDe("s1", 1, "p1"), Items: []app.HistoriaAsignada{
			asignada(historiaDe("a", 4, backlog.PrioridadShould, 3, backlog.EstadoPendiente)),
			asignada(historiaDe("b", 2, backlog.PrioridadMust, 5, backlog.EstadoCompletado)),
			asignada(historiaDe("c", 1, backlog.PrioridadMust, 2, backlog.EstadoPendiente)),
		}}},
		disponibles: []backlog.ItemBacklog{
			historiaDe("x", 9, backlog.PrioridadCould, 0, backlog.EstadoPendiente),
			historiaDe("y", 8, backlog.PrioridadMust, 0, backlog.EstadoPendiente),
		},
	}

	got, err := app.NuevoConsultarSprint(repo).Ejecutar(context.Background(), "p1", "s1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	quiero := []string{"c", "b", "a"}
	if obtenido := ids(got.Historias, func(h app.HistoriaDelSprint) string { return h.Item.ID }); !igual(obtenido, quiero) {
		t.Errorf("historias = %v, quiero %v", obtenido, quiero)
	}
	if got.Historias[0].Completada || !got.Historias[1].Completada {
		t.Errorf("completadas = %v %v, quiero false y true", got.Historias[0].Completada, got.Historias[1].Completada)
	}
	if got.Planificados != 10 || got.Completados != 5 {
		t.Errorf("planificados %d, completados %d; quiero 10 y 5", got.Planificados, got.Completados)
	}
	if obtenido := ids(got.Disponibles, func(i backlog.ItemBacklog) string { return i.ID }); !igual(obtenido, []string{"y", "x"}) {
		t.Errorf("disponibles = %v, quiero y, x", obtenido)
	}
	if got.Sprint.ID != "s1" {
		t.Errorf("Sprint.ID = %q, quiero s1", got.Sprint.ID)
	}
}

// RN-T015-5, CL-T015-8: un sprint solo se alcanza por la ruta de su proyecto.
func TestConsultarSprint_Inexistente(t *testing.T) {
	repo := &lecturasDeSprintsFalsas{sprints: []app.SprintConItems{{Sprint: sprintDe("s1", 1, "p1")}}}
	consultar := app.NuevoConsultarSprint(repo)

	casos := []struct{ nombre, proyectoID, sprintID string }{
		{"el sprint es de otro proyecto", "p2", "s1"},
		{"el sprint no existe", "p1", "no-esta"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if _, err := consultar.Ejecutar(context.Background(), c.proyectoID, c.sprintID); !errors.Is(err, app.ErrSprintInexistente) {
				t.Errorf("Ejecutar() = %v, quiero ErrSprintInexistente", err)
			}
		})
	}
}

func TestConsultarSprint_ErrorDeLaBase(t *testing.T) {
	errBase := errors.New("base caida")
	_, err := app.NuevoConsultarSprint(&lecturasDeSprintsFalsas{err: errBase}).Ejecutar(context.Background(), "p1", "s1")
	if !errors.Is(err, errBase) {
		t.Errorf("Ejecutar() = %v, quiero un error que envuelva %v", err, errBase)
	}
}

// DEF-007: en un sprint cerrado manda el resultado que quedo en la fila de la relacion, no el estado
// actual de la historia. La "b" no se completo en el sprint 1 y volvio al backlog; despues se completo
// en otro sprint, pero para el sprint 1 sigue siendo no completada.
func historiasDeUnSprintCerradoConUnaReasignada() []app.HistoriaAsignada {
	return []app.HistoriaAsignada{
		cerrada(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado), sprint.ResultadoCompletado),
		cerrada(historiaDe("b", 2, backlog.PrioridadMust, 3, backlog.EstadoCompletado), sprint.ResultadoNoCompletado),
		cerrada(historiaDe("c", 3, backlog.PrioridadShould, 2, backlog.EstadoPendiente), sprint.ResultadoNoCompletado),
	}
}

func TestListarSprints_SprintCerradoUsaElResultadoDeLaFila(t *testing.T) {
	cerrado := sprintDe("s1", 1, "p1")
	cerrado.Estado = sprint.EstadoCerrado
	repo := &lecturasDeSprintsFalsas{
		proyectos: map[string]bool{"p1": true},
		sprints:   []app.SprintConItems{{Sprint: cerrado, Items: historiasDeUnSprintCerradoConUnaReasignada()}},
	}

	got, err := app.NuevoListarSprints(repo).Ejecutar(context.Background(), "p1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if got[0].Planificados != 10 || got[0].Completados != 5 {
		t.Errorf("sprint cerrado = %d planificados y %d completados, quiero 10 y 5 (la \"b\" no se completo en el)", got[0].Planificados, got[0].Completados)
	}
}

func TestConsultarSprint_SprintCerradoUsaElResultadoDeLaFila(t *testing.T) {
	cerrado := sprintDe("s1", 1, "p1")
	cerrado.Estado = sprint.EstadoCerrado
	repo := &lecturasDeSprintsFalsas{
		proyectos: map[string]bool{"p1": true},
		sprints:   []app.SprintConItems{{Sprint: cerrado, Items: historiasDeUnSprintCerradoConUnaReasignada()}},
	}

	got, err := app.NuevoConsultarSprint(repo).Ejecutar(context.Background(), "p1", "s1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if got.Planificados != 10 || got.Completados != 5 {
		t.Errorf("sprint cerrado = %d planificados y %d completados, quiero 10 y 5", got.Planificados, got.Completados)
	}
	completadas := map[string]bool{}
	resultados := map[string]sprint.Resultado{}
	for _, h := range got.Historias {
		completadas[h.Item.ID] = h.Completada
		resultados[h.Item.ID] = h.Resultado
	}
	if !completadas["a"] || completadas["b"] || completadas["c"] {
		t.Errorf("completadas = %v, quiero solo la a", completadas)
	}
	if resultados["a"] != sprint.ResultadoCompletado || resultados["b"] != sprint.ResultadoNoCompletado {
		t.Errorf("resultados = %v, quiero que cada historia traiga el de su fila", resultados)
	}
}

// En un sprint abierto la fila sigue pendiente: manda el estado de la historia (US-010 la marca
// completada sin tocar la fila hasta el cierre).
func TestConsultarSprint_SprintAbiertoUsaElEstadoDeLaHistoria(t *testing.T) {
	repo := &lecturasDeSprintsFalsas{
		proyectos: map[string]bool{"p1": true},
		sprints: []app.SprintConItems{{Sprint: sprintDe("s1", 1, "p1"), Items: []app.HistoriaAsignada{
			asignada(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado)),
			asignada(historiaDe("b", 2, backlog.PrioridadMust, 3, backlog.EstadoPendiente)),
		}}},
	}

	got, err := app.NuevoConsultarSprint(repo).Ejecutar(context.Background(), "p1", "s1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if got.Completados != 5 || !got.Historias[0].Completada || got.Historias[1].Completada {
		t.Errorf("sprint abierto = %+v, quiero la a completada y 5 completados", got)
	}
}
