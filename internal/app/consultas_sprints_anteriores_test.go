package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// cerradoDe arma un sprint cerrado con sus historias.
func cerradoDe(id string, numero int, items ...app.HistoriaAsignada) app.SprintConItems {
	s := sprintDe(id, numero, "p1")
	s.Estado = sprint.EstadoCerrado
	return app.SprintConItems{Sprint: s, Items: items}
}

// completadas arma una historia completada por cada cantidad de puntos, de un sprint cerrado.
func completadas(prefijo string, puntos ...int) []app.HistoriaAsignada {
	historias := make([]app.HistoriaAsignada, len(puntos))
	for i, p := range puntos {
		historias[i] = cerrada(historiaDe(fmt.Sprintf("%s%d", prefijo, i), i+1, backlog.PrioridadMust, p, backlog.EstadoCompletado), sprint.ResultadoCompletado)
	}
	return historias
}

func consultarAnteriores(repo *lecturasDeSprintsFalsas) (app.SprintsAnteriores, error) {
	return app.NuevoConsultarSprintsAnteriores(repo).Ejecutar(context.Background(), "p1")
}

// US-012 (CA-012-1, CA-012-2 y CA-012-3): solo los sprints cerrados, del mas reciente al mas antiguo, cada
// uno con sus Story Points y la cantidad de historias completadas y no completadas, y la
// velocidad del equipo con esos mismos sprints (RN-012-2 a RN-012-5).
func TestConsultarSprintsAnteriores(t *testing.T) {
	abierto := sprintDe("s3", 3, "p1")
	repo := &lecturasDeSprintsFalsas{
		proyectos: map[string]bool{"p1": true},
		sprints: []app.SprintConItems{
			cerradoDe("s1", 1,
				cerrada(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado), sprint.ResultadoCompletado),
				cerrada(historiaDe("b", 2, backlog.PrioridadMust, 3, backlog.EstadoPendiente), sprint.ResultadoNoCompletado)),
			{Sprint: abierto, Items: []app.HistoriaAsignada{asignada(historiaDe("d", 4, backlog.PrioridadMust, 13, backlog.EstadoCompletado))}},
			cerradoDe("s2", 2,
				cerrada(historiaDe("c", 3, backlog.PrioridadMust, 8, backlog.EstadoCompletado), sprint.ResultadoCompletado)),
		},
	}

	got, err := consultarAnteriores(repo)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if len(got.Sprints) != 2 || got.Sprints[0].Sprint.ID != "s2" || got.Sprints[1].Sprint.ID != "s1" {
		t.Fatalf("sprints = %+v, quiero s2 y s1 (el abierto no aparece, el mas reciente primero)", got.Sprints)
	}
	segundo, primero := got.Sprints[0], got.Sprints[1]
	if segundo.Planificados != 8 || segundo.Completados != 8 || segundo.HistoriasCompletadas != 1 || segundo.HistoriasNoCompletadas != 0 {
		t.Errorf("sprint 2 = %+v, quiero 8 planificados, 8 completados, 1 historia completada y 0 no completadas", segundo)
	}
	if primero.Planificados != 8 || primero.Completados != 5 || primero.HistoriasCompletadas != 1 || primero.HistoriasNoCompletadas != 1 {
		t.Errorf("sprint 1 = %+v, quiero 8 planificados, 5 completados, 1 historia completada y 1 no completada", primero)
	}
	// RN-012-5: (8 + 5) / 2, solo con los sprints cerrados; los 13 puntos del abierto no cuentan.
	if got.Velocidad != 6.5 {
		t.Errorf("Velocidad = %v, quiero 6,5", got.Velocidad)
	}
}

// US-012 / CL-012-7, RN-012-5: la velocidad sale de metricas, con su redondeo.
func TestConsultarSprintsAnteriores_Velocidad(t *testing.T) {
	repo := &lecturasDeSprintsFalsas{
		proyectos: map[string]bool{"p1": true},
		sprints: []app.SprintConItems{
			// 30 = 13 + 13 + 3 + 1 y 33 = 13 + 13 + 5 + 2: la escala no pasa de 13.
			cerradoDe("s1", 1, completadas("a", 13, 13, 3, 1)...),
			cerradoDe("s2", 2, completadas("b", 13, 13, 5, 2)...),
		},
	}

	got, err := consultarAnteriores(repo)

	if err != nil || got.Velocidad != 31.5 {
		t.Errorf("Velocidad = %v, %v; quiero 31,5", got.Velocidad, err)
	}
}

// US-012 / CA-012-5, RN-012-4: en un sprint cerrado manda el resultado de la fila y no el estado
// actual de la historia. La "b" no se completo en el sprint 1, volvio al backlog y se completo
// despues: su estado es "completado", pero en el sprint 1 sigue siendo no completada.
func TestConsultarSprintsAnteriores_ElResultadoEsElDeLaFila(t *testing.T) {
	repo := &lecturasDeSprintsFalsas{
		proyectos: map[string]bool{"p1": true},
		sprints: []app.SprintConItems{cerradoDe("s1", 1,
			cerrada(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado), sprint.ResultadoCompletado),
			cerrada(historiaDe("b", 2, backlog.PrioridadMust, 3, backlog.EstadoCompletado), sprint.ResultadoNoCompletado))},
	}

	got, err := consultarAnteriores(repo)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	s := got.Sprints[0]
	if s.Planificados != 8 || s.Completados != 5 || s.HistoriasCompletadas != 1 || s.HistoriasNoCompletadas != 1 {
		t.Errorf("sprint 1 = %+v, quiero 8 planificados, 5 completados, 1 completada y 1 no completada", s)
	}
}

// US-012 / CA-012-4, CL-012-1 a CL-012-5, CL-012-9, RN-012-6, RN-012-7: lo que no tiene resultado no es un error.
func TestConsultarSprintsAnteriores_CasosLimite(t *testing.T) {
	casos := []struct {
		nombre                                                string
		sprints                                               []app.SprintConItems
		cantidad, planificados, completados, hechas, noHechas int
	}{
		{"proyecto sin sprints", nil, 0, 0, 0, 0, 0},
		{"solo hay un sprint abierto", []app.SprintConItems{{Sprint: sprintDe("s1", 1, "p1")}}, 0, 0, 0, 0, 0},
		{"sprint cerrado sin historias", []app.SprintConItems{cerradoDe("s1", 1)}, 1, 0, 0, 0, 0},
		{"todas completadas", []app.SprintConItems{cerradoDe("s1", 1,
			cerrada(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado), sprint.ResultadoCompletado),
			cerrada(historiaDe("b", 2, backlog.PrioridadMust, 3, backlog.EstadoCompletado), sprint.ResultadoCompletado))}, 1, 8, 8, 2, 0},
		{"ninguna completada", []app.SprintConItems{cerradoDe("s1", 1,
			cerrada(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoPendiente), sprint.ResultadoNoCompletado),
			cerrada(historiaDe("b", 2, backlog.PrioridadMust, 3, backlog.EstadoPendiente), sprint.ResultadoNoCompletado))}, 1, 8, 0, 0, 2},
		{"una historia sin estimar cuenta como historia y como cero puntos", []app.SprintConItems{cerradoDe("s1", 1,
			cerrada(historiaDe("a", 1, backlog.PrioridadMust, 0, backlog.EstadoPendiente), sprint.ResultadoNoCompletado))}, 1, 0, 0, 0, 1},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := &lecturasDeSprintsFalsas{proyectos: map[string]bool{"p1": true}, sprints: c.sprints}

			got, err := consultarAnteriores(repo)

			if err != nil {
				t.Fatalf("Ejecutar: %v", err)
			}
			if len(got.Sprints) != c.cantidad {
				t.Fatalf("hay %d sprints, quiero %d", len(got.Sprints), c.cantidad)
			}
			if c.cantidad == 0 {
				if got.Velocidad != 0 {
					t.Errorf("Velocidad = %v, quiero 0", got.Velocidad)
				}
				return
			}
			s := got.Sprints[0]
			if s.Planificados != c.planificados || s.Completados != c.completados || s.HistoriasCompletadas != c.hechas || s.HistoriasNoCompletadas != c.noHechas {
				t.Errorf("sprint = %+v, quiero %d planificados, %d completados, %d completadas y %d no completadas", s, c.planificados, c.completados, c.hechas, c.noHechas)
			}
		})
	}
}

// US-012 / CA-012-6, RN-012-1, CL-012-11: un proyecto inexistente es app.ErrProyectoInexistente.
func TestConsultarSprintsAnteriores_ProyectoInexistente(t *testing.T) {
	_, err := app.NuevoConsultarSprintsAnteriores(&lecturasDeSprintsFalsas{}).Ejecutar(context.Background(), "no-esta")

	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Errorf("Ejecutar() = %v, quiero ErrProyectoInexistente", err)
	}
}

// El fallo de la base llega envuelto con contexto.
func TestConsultarSprintsAnteriores_ErrorDeLaBase(t *testing.T) {
	errBase := errors.New("base caida")

	_, err := consultarAnteriores(&lecturasDeSprintsFalsas{err: errBase})

	if !errors.Is(err, errBase) {
		t.Errorf("Ejecutar() = %v, quiero un error que envuelva %v", err, errBase)
	}
}
