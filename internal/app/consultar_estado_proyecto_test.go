package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// estadoGeneralFalso simula las tres lecturas de US-004 sin base de datos y cuenta cuantas veces se llama a cada una
// (RN-004-8: una sola lectura de cada tipo).
type estadoGeneralFalso struct {
	proyecto proyecto.Proyecto
	items    []backlog.ItemBacklog
	sprints  []app.SprintConItems
	err      error

	lecturas map[string]int
}

func (f *estadoGeneralFalso) contar(nombre string) {
	if f.lecturas == nil {
		f.lecturas = map[string]int{}
	}
	f.lecturas[nombre]++
}

func (f *estadoGeneralFalso) BuscarProyecto(_ context.Context, id string) (proyecto.Proyecto, error) {
	f.contar("proyecto")
	if f.proyecto.ID != id {
		return proyecto.Proyecto{}, app.ErrProyectoInexistente
	}
	return f.proyecto, nil
}

func (f *estadoGeneralFalso) ListarItems(context.Context, string) ([]backlog.ItemBacklog, error) {
	f.contar("items")
	return append([]backlog.ItemBacklog(nil), f.items...), f.err
}

func (f *estadoGeneralFalso) ListarSprintsConItems(context.Context, string) ([]app.SprintConItems, error) {
	f.contar("sprints")
	return append([]app.SprintConItems(nil), f.sprints...), f.err
}

func proyectoDeEstado() proyecto.Proyecto {
	return proyecto.Proyecto{ID: "p1", Nombre: "Demo",
		FechaInicio: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), FechaFin: time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)}
}

func consultarEstado(repo *estadoGeneralFalso, hoy time.Time) (app.EstadoGeneral, error) {
	return app.NuevoConsultarEstadoDelProyecto(repo, func() time.Time { return hoy }).Ejecutar(context.Background(), "p1")
}

var seisDeOctubre = time.Date(2026, 10, 6, 15, 30, 0, 0, time.UTC)

// US-004 / CA-004-1, CA-004-2: el estado y los dias salen del dominio con el reloj del caso de uso.
func TestConsultarEstado_EstadoDeLasFechas(t *testing.T) {
	casos := []struct {
		nombre string
		hoy    time.Time
		estado proyecto.EstadoDelProyecto
		dias   int
	}{
		{"en curso: faltan 86 dias", seisDeOctubre, proyecto.EnCurso, 86},
		{"no iniciado: empieza en 12 dias", time.Date(2025, 12, 20, 0, 0, 0, 0, time.UTC), proyecto.NoIniciado, 12},
		{"finalizado: termino hace 3 dias", time.Date(2027, 1, 3, 0, 0, 0, 0, time.UTC), proyecto.Finalizado, 3},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e, err := consultarEstado(&estadoGeneralFalso{proyecto: proyectoDeEstado()}, c.hoy)
			if err != nil {
				t.Fatalf("Ejecutar: %v", err)
			}
			if e.Estado != c.estado || e.Dias != c.dias || e.Proyecto.ID != "p1" {
				t.Errorf("estado = %q con %d dias del proyecto %q, se esperaba %q con %d", e.Estado, e.Dias, e.Proyecto.ID, c.estado, c.dias)
			}
		})
	}
}

// US-004 / CA-004-3, RN-004-5, RN-004-6, RN-004-7, CL-004-9: las historias se cuentan por su estado actual (con las que no
// estan en ningun sprint), los Story Points de las sin estimar suman 0 y el sprint abierto es el unico abierto.
func TestConsultarEstado_Avance(t *testing.T) {
	cerrado := sprintDe("s1", 1, "p1")
	cerrado.Estado = sprint.EstadoCerrado
	abierto := sprintDe("s2", 2, "p1")
	repo := &estadoGeneralFalso{
		proyecto: proyectoDeEstado(),
		items: []backlog.ItemBacklog{
			historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado),
			historiaDe("b", 2, backlog.PrioridadMust, 8, backlog.EstadoEnProgreso),
			historiaDe("c", 3, backlog.PrioridadShould, 8, backlog.EstadoPendiente),
			historiaDe("d", 4, backlog.PrioridadCould, 0, backlog.EstadoCompletado), // completada y sin estimar: suma 0
		},
		sprints: []app.SprintConItems{{Sprint: cerrado}, {Sprint: abierto}},
	}

	e, err := consultarEstado(repo, seisDeOctubre)

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if e.Historias != (app.HistoriasPorEstado{Pendientes: 1, EnProgreso: 1, Completadas: 2, Total: 4}) {
		t.Errorf("historias = %+v, se esperaba 1 pendiente, 1 en progreso, 2 completadas y 4 en total", e.Historias)
	}
	if e.StoryPoints != 21 || e.Completados != 5 {
		t.Errorf("Story Points = %d completados de %d, se esperaban 5 de 21", e.Completados, e.StoryPoints)
	}
	if e.Porcentaje != 50 {
		t.Errorf("porcentaje = %v, se esperaba 50", e.Porcentaje)
	}
	if e.SprintAbierto == nil || e.SprintAbierto.ID != "s2" {
		t.Errorf("sprint abierto = %+v, se esperaba el s2", e.SprintAbierto)
	}
}

// US-004 / CA-004-3: un tercio completado se redondea a 1 decimal, como en US-028.
func TestConsultarEstado_PorcentajeConUnDecimal(t *testing.T) {
	repo := &estadoGeneralFalso{proyecto: proyectoDeEstado(), items: []backlog.ItemBacklog{
		historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado),
		historiaDe("b", 2, backlog.PrioridadMust, 3, backlog.EstadoEnProgreso),
		historiaDe("c", 3, backlog.PrioridadMust, 2, backlog.EstadoPendiente),
	}}

	e, err := consultarEstado(repo, seisDeOctubre)

	if err != nil || e.Porcentaje != 33.3 {
		t.Errorf("porcentaje = %v (%v), se esperaba 33,3", e.Porcentaje, err)
	}
}

// US-004 / CA-004-5, CL-004-7, CL-004-10: sin historias ni sprints (o con todos los sprints cerrados) todo es 0, el
// porcentaje no divide por cero y no hay sprint abierto; nada de eso es un error.
func TestConsultarEstado_ProyectoVacio(t *testing.T) {
	e, err := consultarEstado(&estadoGeneralFalso{proyecto: proyectoDeEstado()}, seisDeOctubre)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if e.Historias != (app.HistoriasPorEstado{}) || e.StoryPoints != 0 || e.Completados != 0 || e.Porcentaje != 0 || e.SprintAbierto != nil {
		t.Errorf("estado = %+v, se esperaba todo en 0 y sin sprint abierto", e)
	}

	cerrado := sprintDe("s1", 1, "p1")
	cerrado.Estado = sprint.EstadoCerrado
	e, err = consultarEstado(&estadoGeneralFalso{proyecto: proyectoDeEstado(), sprints: []app.SprintConItems{{Sprint: cerrado}}}, seisDeOctubre)
	if err != nil || e.SprintAbierto != nil {
		t.Errorf("con todos los sprints cerrados: sprint abierto = %+v (%v), se esperaba ninguno", e.SprintAbierto, err)
	}
}

// US-004 / CL-004-12: el estado depende solo de las fechas, no del avance: un proyecto finalizado con historias sin
// completar sigue finalizado y el avance dice cuanto faltó.
func TestConsultarEstado_FinalizadoConHistoriasSinCompletar(t *testing.T) {
	repo := &estadoGeneralFalso{proyecto: proyectoDeEstado(), items: []backlog.ItemBacklog{
		historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado),
		historiaDe("b", 2, backlog.PrioridadMust, 5, backlog.EstadoPendiente),
	}}

	e, err := consultarEstado(repo, time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC))

	if err != nil || e.Estado != proyecto.Finalizado || e.Porcentaje != 50 {
		t.Errorf("estado = %q, porcentaje = %v (%v), se esperaba finalizado y 50", e.Estado, e.Porcentaje, err)
	}
}

// US-004 / RN-004-4, CA-004-6, CL-004-11: un proyecto que no existe se rechaza con ErrProyectoInexistente, envuelto con
// el ID, y no se lee nada mas.
func TestConsultarEstado_ProyectoInexistente(t *testing.T) {
	repo := &estadoGeneralFalso{proyecto: proyectoDeEstado()}

	_, err := app.NuevoConsultarEstadoDelProyecto(repo, func() time.Time { return seisDeOctubre }).Ejecutar(context.Background(), "otro")

	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Fatalf("error = %v, se esperaba ErrProyectoInexistente", err)
	}
	if repo.lecturas["items"] != 0 || repo.lecturas["sprints"] != 0 {
		t.Errorf("lecturas despues de un proyecto inexistente = %v, no se esperaba ninguna", repo.lecturas)
	}
}

// US-004 / RN-004-8: una sola lectura de cada tipo, no una por historia ni por sprint.
func TestConsultarEstado_UnaLecturaDeCadaTipo(t *testing.T) {
	repo := &estadoGeneralFalso{proyecto: proyectoDeEstado(), items: []backlog.ItemBacklog{
		historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado),
		historiaDe("b", 2, backlog.PrioridadMust, 3, backlog.EstadoPendiente),
	}, sprints: []app.SprintConItems{{Sprint: sprintDe("s1", 1, "p1")}, {Sprint: sprintDe("s2", 2, "p1")}}}

	if _, err := consultarEstado(repo, seisDeOctubre); err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if repo.lecturas["proyecto"] != 1 || repo.lecturas["items"] != 1 || repo.lecturas["sprints"] != 1 {
		t.Errorf("lecturas = %v, se esperaba una de cada tipo", repo.lecturas)
	}
}

// US-004 / sección 7: una falla de la persistencia se devuelve envuelta, sin ocultarla. (Un dato corrupto que rechaza metricas, o
// sea Story Points negativos, no se puede armar desde afuera: backlog.StoryPoints no deja crearlos.)
func TestConsultarEstado_FallaDeLaPersistencia(t *testing.T) {
	caido := errors.New("la base no responde")

	_, err := consultarEstado(&estadoGeneralFalso{proyecto: proyectoDeEstado(), err: caido}, seisDeOctubre)

	if !errors.Is(err, caido) {
		t.Errorf("error = %v, se esperaba el de la persistencia envuelto", err)
	}
}
