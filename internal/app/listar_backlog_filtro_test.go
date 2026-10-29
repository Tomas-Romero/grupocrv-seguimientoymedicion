package app_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// Tests del filtro del backlog (spec US-007). Se arman sobre el mismo repositorio falso de T-011: devuelve
// todas las historias del proyecto sin ordenar y el caso de uso filtra y ordena (seccion 5).

// itemFiltrable es una historia del proyecto p1 con el estado y la prioridad que mira el filtro.
func itemFiltrable(id string, numero int, prioridad backlog.Prioridad, estado backlog.Estado) backlog.ItemBacklog {
	item := itemDe(id, numero, prioridad)
	item.Estado = estado
	return item
}

// backlogDeEjemplo es el proyecto p1 con seis historias que cubren las cuatro prioridades y los tres estados.
// Sin filtrar salen en este orden: c, e, d, f, b, a (RN-T011-4).
//
//	a #1 wont:   pendiente      b #2 could:  completado     c #3 must:   pendiente
//	d #4 should: en_progreso    e #5 must:   completado     f #6 should: pendiente
//
// "p2" existe y no tiene historias.
func backlogDeEjemplo() *lecturasFalsas {
	return &lecturasFalsas{
		proyectos: []proyecto.Proyecto{
			proyectoDe("p1", "Demo", fechaDe(2026, time.October, 1)),
			proyectoDe("p2", "Vacio", fechaDe(2026, time.October, 1)),
		},
		items: map[string][]backlog.ItemBacklog{"p1": {
			itemFiltrable("a", 1, backlog.PrioridadWont, backlog.EstadoPendiente),
			itemFiltrable("b", 2, backlog.PrioridadCould, backlog.EstadoCompletado),
			itemFiltrable("c", 3, backlog.PrioridadMust, backlog.EstadoPendiente),
			itemFiltrable("d", 4, backlog.PrioridadShould, backlog.EstadoEnProgreso),
			itemFiltrable("e", 5, backlog.PrioridadMust, backlog.EstadoCompletado),
			itemFiltrable("f", 6, backlog.PrioridadShould, backlog.EstadoPendiente),
		}},
	}
}

// listarBacklogConFiltro corre el caso de uso sobre el proyecto de ejemplo y devuelve los IDs, en el orden
// del resultado.
func listarBacklogConFiltro(t *testing.T, filtro app.FiltroBacklog) []string {
	t.Helper()
	got, err := app.NuevoListarBacklog(backlogDeEjemplo()).Ejecutar(context.Background(), "p1", filtro)
	if err != nil {
		t.Fatalf("Ejecutar(%+v): %v", filtro, err)
	}
	return ids(got, func(i backlog.ItemBacklog) string { return i.ID })
}

// US-007 / CA-007-3 / RN-007-1 / CL-007-1: sin filtros el resultado es el de T-011: todas las historias,
// en cualquier estado, por prioridad y despues por numero.
func TestListarBacklog_SinFiltrosEsElBacklogCompleto(t *testing.T) {
	quiero := []string{"c", "e", "d", "f", "b", "a"}
	if got := listarBacklogConFiltro(t, app.FiltroBacklog{}); !igual(got, quiero) {
		t.Errorf("sin filtros = %v, quiero %v", got, quiero)
	}
}

// US-007 / CA-007-1 / RN-007-2 / RN-007-6 / CL-007-2, CL-007-3, CL-007-4, CL-007-11: el filtro por estado
// deja solo las historias en ese estado, de cualquier prioridad, y no cambia el orden: sigue por
// prioridad y despues por numero.
func TestListarBacklog_FiltroPorEstado(t *testing.T) {
	casos := []struct {
		nombre string
		estado backlog.Estado
		quiero []string
	}{
		{"pendiente", backlog.EstadoPendiente, []string{"c", "f", "a"}},
		{"en progreso", backlog.EstadoEnProgreso, []string{"d"}},
		{"completado", backlog.EstadoCompletado, []string{"e", "b"}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := listarBacklogConFiltro(t, app.FiltroBacklog{Estado: c.estado}); !igual(got, c.quiero) {
				t.Errorf("estado %q = %v, quiero %v", c.estado, got, c.quiero)
			}
		})
	}
}

// US-007 / CA-007-1 / RN-007-6 / CL-007-12: el filtro solo quita historias: las que quedan tienen los
// mismos datos que sin filtrar (numero, titulo, descripcion, Story Points, criterios...).
func TestListarBacklog_FiltroNoCambiaLosDatos(t *testing.T) {
	puntos, err := backlog.NuevosStoryPoints(5)
	if err != nil {
		t.Fatalf("NuevosStoryPoints: %v", err)
	}
	completa := itemFiltrable("x", 7, backlog.PrioridadMust, backlog.EstadoCompletado)
	completa.Descripcion, completa.Criterios, completa.StoryPoints = "Con todo", []string{"Primero", "Segundo"}, puntos
	repo := backlogDeEjemplo()
	repo.items["p1"] = append(repo.items["p1"], completa)

	got, err := app.NuevoListarBacklog(repo).Ejecutar(context.Background(), "p1", app.FiltroBacklog{Estado: backlog.EstadoCompletado})
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("hay %d historias completadas, quiero 3", len(got))
	}
	// x es must #7: sale despues de e (must #5) y antes de b (could), asi que es la segunda.
	if !reflect.DeepEqual(got[1], completa) {
		t.Errorf("la historia filtrada = %+v, quiero la misma que sin filtrar (%+v)", got[1], completa)
	}
}

// US-007 / CA-007-1 / RN-007-3 / CL-007-5: el filtro por prioridad deja solo las historias de esa
// prioridad, en cualquier estado. Se prueba con cada una de las cuatro.
func TestListarBacklog_FiltroPorPrioridad(t *testing.T) {
	casos := []struct {
		nombre    string
		prioridad backlog.Prioridad
		quiero    []string
	}{
		{"must", backlog.PrioridadMust, []string{"c", "e"}},
		{"should", backlog.PrioridadShould, []string{"d", "f"}},
		{"could", backlog.PrioridadCould, []string{"b"}},
		{"wont", backlog.PrioridadWont, []string{"a"}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := listarBacklogConFiltro(t, app.FiltroBacklog{Prioridad: c.prioridad}); !igual(got, c.quiero) {
				t.Errorf("prioridad %q = %v, quiero %v", c.prioridad, got, c.quiero)
			}
		})
	}
}

// US-007 / CA-007-2 / RN-007-4 / CL-007-6, CL-007-9: los dos filtros se combinan con Y. Una historia sale
// solo si cumple los dos; la completada must no sale con pendiente, ni la pendiente should con must.
func TestListarBacklog_FiltrosCombinadosConY(t *testing.T) {
	casos := []struct {
		nombre    string
		estado    backlog.Estado
		prioridad backlog.Prioridad
		quiero    []string
	}{
		{"pendiente y must", backlog.EstadoPendiente, backlog.PrioridadMust, []string{"c"}},
		{"completado y must", backlog.EstadoCompletado, backlog.PrioridadMust, []string{"e"}},
		{"pendiente y should", backlog.EstadoPendiente, backlog.PrioridadShould, []string{"f"}},
		{"en progreso y should", backlog.EstadoEnProgreso, backlog.PrioridadShould, []string{"d"}},
		{"cada uno tiene historias pero no en comun", backlog.EstadoEnProgreso, backlog.PrioridadMust, []string{}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got := listarBacklogConFiltro(t, app.FiltroBacklog{Estado: c.estado, Prioridad: c.prioridad})
			if !igual(got, c.quiero) {
				t.Errorf("estado %q y prioridad %q = %v, quiero %v", c.estado, c.prioridad, got, c.quiero)
			}
		})
	}
}

// US-007 / CA-007-4 / RN-007-4 / CL-007-1: un solo valor vacio no filtra por esa dimension y el otro se
// aplica solo.
func TestListarBacklog_UnValorVacioNoFiltraEsaDimension(t *testing.T) {
	if got, quiero := listarBacklogConFiltro(t, app.FiltroBacklog{Prioridad: backlog.PrioridadMust}), []string{"c", "e"}; !igual(got, quiero) {
		t.Errorf("solo prioridad = %v, quiero %v", got, quiero)
	}
	if got, quiero := listarBacklogConFiltro(t, app.FiltroBacklog{Estado: backlog.EstadoPendiente}), []string{"c", "f", "a"}; !igual(got, quiero) {
		t.Errorf("solo estado = %v, quiero %v", got, quiero)
	}
}

// US-007 / CA-007-4 / RN-007-5 / CL-007-7, CL-007-8, CL-007-10: que ninguna historia cumpla el filtro no
// es un error: es una lista vacia. Vale para un estado sin historias, una prioridad sin historias y un
// proyecto sin historias, con filtros validos.
func TestListarBacklog_SinCoincidenciasEsListaVacia(t *testing.T) {
	soloPendientesMust := &lecturasFalsas{
		proyectos: []proyecto.Proyecto{proyectoDe("p1", "Demo", fechaDe(2026, time.October, 1))},
		items: map[string][]backlog.ItemBacklog{"p1": {
			itemFiltrable("a", 1, backlog.PrioridadMust, backlog.EstadoPendiente),
			itemFiltrable("b", 2, backlog.PrioridadCould, backlog.EstadoPendiente),
		}},
	}
	casos := []struct {
		nombre string
		repo   *lecturasFalsas
		id     string
		filtro app.FiltroBacklog
	}{
		{"un estado sin historias", soloPendientesMust, "p1", app.FiltroBacklog{Estado: backlog.EstadoCompletado}},
		{"una prioridad sin historias", soloPendientesMust, "p1", app.FiltroBacklog{Prioridad: backlog.PrioridadWont}},
		{"un proyecto sin historias con filtros", backlogDeEjemplo(), "p2", app.FiltroBacklog{Estado: backlog.EstadoPendiente, Prioridad: backlog.PrioridadMust}},
		{"un proyecto sin historias sin filtros", backlogDeEjemplo(), "p2", app.FiltroBacklog{}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got, err := app.NuevoListarBacklog(c.repo).Ejecutar(context.Background(), c.id, c.filtro)
			if err != nil || len(got) != 0 {
				t.Errorf("Ejecutar() = %v, %v; quiero lista vacia y sin error", got, err)
			}
		})
	}
}

// backlogContado cuenta cuantas veces se leyo la lista de historias, para comprobar que un rechazo ocurre
// antes de leerla (RN-007-9, CL-007-19).
type backlogContado struct {
	*lecturasFalsas
	lecturas int
}

func (b *backlogContado) ListarItems(ctx context.Context, proyectoID string) ([]backlog.ItemBacklog, error) {
	b.lecturas++
	return b.lecturasFalsas.ListarItems(ctx, proyectoID)
}

// US-007 / CA-007-5 / RN-007-7 / CL-007-14: un estado que no es uno de los tres se rechaza con
// ErrEstadoInvalido, envuelto con el valor, sin pasarlo a minuscula ni recortarlo, y sin leer la lista.
func TestListarBacklog_EstadoInvalido(t *testing.T) {
	invalidos := []backlog.Estado{"Pendiente", "COMPLETADO", "en progreso", " pendiente ", " ", "terminado"}
	for _, estado := range invalidos {
		t.Run(string(estado), func(t *testing.T) {
			repo := &backlogContado{lecturasFalsas: backlogDeEjemplo()}

			got, err := app.NuevoListarBacklog(repo).Ejecutar(context.Background(), "p1", app.FiltroBacklog{Estado: estado})

			if !errors.Is(err, backlog.ErrEstadoInvalido) {
				t.Fatalf("Ejecutar() = %v, %v; quiero ErrEstadoInvalido", got, err)
			}
			if errors.Is(err, backlog.ErrPrioridadInvalida) {
				t.Errorf("el error %q tambien es ErrPrioridadInvalida y la prioridad estaba vacia", err)
			}
			if esperado := `se recibio "` + string(estado) + `"`; !strings.Contains(err.Error(), esperado) {
				t.Errorf("el error = %q, quiero que incluya %s", err.Error(), esperado)
			}
			if got != nil || repo.lecturas != 0 {
				t.Errorf("resultado %v y %d lecturas; quiero ninguna lista y no leer la base", got, repo.lecturas)
			}
		})
	}
}

// US-007 / CA-007-5 / RN-007-8 / CL-007-15: una prioridad que no es una de las cuatro se rechaza con
// ErrPrioridadInvalida, envuelto con el valor, y sin leer la lista.
func TestListarBacklog_PrioridadInvalida(t *testing.T) {
	invalidas := []backlog.Prioridad{"Must", "MUST", " must ", " ", "urgente"}
	for _, prioridad := range invalidas {
		t.Run(string(prioridad), func(t *testing.T) {
			repo := &backlogContado{lecturasFalsas: backlogDeEjemplo()}

			got, err := app.NuevoListarBacklog(repo).Ejecutar(context.Background(), "p1", app.FiltroBacklog{Prioridad: prioridad})

			if !errors.Is(err, backlog.ErrPrioridadInvalida) {
				t.Fatalf("Ejecutar() = %v, %v; quiero ErrPrioridadInvalida", got, err)
			}
			if errors.Is(err, backlog.ErrEstadoInvalido) {
				t.Errorf("el error %q tambien es ErrEstadoInvalido y el estado estaba vacio", err)
			}
			if esperado := `se recibio "` + string(prioridad) + `"`; !strings.Contains(err.Error(), esperado) {
				t.Errorf("el error = %q, quiero que incluya %s", err.Error(), esperado)
			}
			if got != nil || repo.lecturas != 0 {
				t.Errorf("resultado %v y %d lecturas; quiero ninguna lista y no leer la base", got, repo.lecturas)
			}
		})
	}
}

// US-007 / CA-007-5 / RN-007-9 / CL-007-16, CL-007-17: con el estado y la prioridad invalidos a la vez se
// devuelven los dos errores unidos, el del estado primero. Con uno solo invalido, solo ese.
func TestListarBacklog_EstadoYPrioridadInvalidosSeUnen(t *testing.T) {
	_, err := app.NuevoListarBacklog(backlogDeEjemplo()).Ejecutar(context.Background(), "p1",
		app.FiltroBacklog{Estado: "terminado", Prioridad: "urgente"})

	if !errors.Is(err, backlog.ErrEstadoInvalido) || !errors.Is(err, backlog.ErrPrioridadInvalida) {
		t.Fatalf("Ejecutar() = %v; quiero ErrEstadoInvalido y ErrPrioridadInvalida unidos", err)
	}
	estado, prioridad := strings.Index(err.Error(), backlog.ErrEstadoInvalido.Error()), strings.Index(err.Error(), backlog.ErrPrioridadInvalida.Error())
	if estado < 0 || prioridad < 0 || estado > prioridad {
		t.Errorf("el mensaje %q tiene que nombrar primero el estado y despues la prioridad", err)
	}
}

// US-007 / CA-007-6 / RN-007-9 / CL-007-18, CL-007-19: un proyecto inexistente (o con un ID mal formado) se
// devuelve solo, antes de mirar los filtros, aunque los dos sean invalidos, y sin leer la lista.
func TestListarBacklog_ProyectoInexistenteVaSolo(t *testing.T) {
	for _, id := range []string{"no-esta", "abc", ""} {
		t.Run("proyecto "+id, func(t *testing.T) {
			repo := &backlogContado{lecturasFalsas: backlogDeEjemplo()}

			_, err := app.NuevoListarBacklog(repo).Ejecutar(context.Background(), id,
				app.FiltroBacklog{Estado: "terminado", Prioridad: "urgente"})

			if !errors.Is(err, app.ErrProyectoInexistente) {
				t.Fatalf("Ejecutar() = %v, quiero ErrProyectoInexistente", err)
			}
			if errors.Is(err, backlog.ErrEstadoInvalido) || errors.Is(err, backlog.ErrPrioridadInvalida) {
				t.Errorf("el error %q tiene que ir solo, sin los de los filtros", err)
			}
			if repo.lecturas != 0 {
				t.Errorf("se leyo la lista %d veces; quiero ninguna", repo.lecturas)
			}
		})
	}
}

// US-007 / RN-007-9 / CL-007-20: una falla inesperada de la base, con filtros validos, se devuelve envuelta
// con su contexto y sin ninguna lista parcial.
func TestListarBacklog_ErrorDeLaBaseConFiltro(t *testing.T) {
	errBase := errors.New("base caida")
	filtro := app.FiltroBacklog{Estado: backlog.EstadoPendiente, Prioridad: backlog.PrioridadMust}

	got, err := app.NuevoListarBacklog(&lecturasFalsas{err: errBase}).Ejecutar(context.Background(), "p1", filtro)

	if !errors.Is(err, errBase) || got != nil {
		t.Errorf("Ejecutar() = %v, %v; quiero un error que envuelva %v y ninguna lista", got, err, errBase)
	}
}
