package app_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// Tests de los filtros de la lista de defectos (spec US-024). Se arman sobre el mismo repositorio falso de
// T-021: devuelve todos los defectos del proyecto sin ordenar y el caso de uso filtra y ordena (seccion 5).

// defectoFiltrable es un defecto listado con el sprint de deteccion y la severidad que el filtro mira. El numero
// de sprint de deteccion sigue al ID ("s1" es el Sprint 1), asi se puede comprobar que el filtro no cambia los
// datos de lectura (RN-024-6).
func defectoFiltrable(id, sprintID string, severidad defecto.Severidad, estado defecto.Estado, creadoEn time.Time) app.DefectoListado {
	l := defectoListado(id, estado, creadoEn)
	l.Defecto.SprintDeteccionID, l.Defecto.Severidad = sprintID, severidad
	l.SprintDeteccion = int(sprintID[1] - '0')
	return l
}

var baseDeFiltros = time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC)

// proyectoConDefectos es el proyecto p1 con tres sprints (s1, s2 y s3, este sin defectos) y un sprint s9 que es
// de otro proyecto. Los defectos, de detectado en: severidad, estado, cubren las cuatro severidades y los cuatro
// estados:
//
//	d1 Sprint 1: alta, abierto        d2 Sprint 1: baja, descartado     d3 Sprint 2: alta, resuelto en el 2
//	d4 Sprint 2: critica, abierto     d5 Sprint 1: media, en correccion d6 Sprint 1: critica, resuelto en el 2
//
// d6 se detecto en el Sprint 1 y se resolvio en el Sprint 2 (CL-024-3).
func proyectoConDefectos() *lecturaDeDefectosFalsa {
	resuelto := defectoFiltrable("d3", "s2", defecto.SeveridadAlta, defecto.EstadoResuelto, baseDeFiltros.Add(2*time.Minute))
	resuelto.Defecto.SprintResolucionID, resuelto.SprintResolucion = "s2", 2
	cruzado := defectoFiltrable("d6", "s1", defecto.SeveridadCritica, defecto.EstadoResuelto, baseDeFiltros.Add(5*time.Minute))
	cruzado.Defecto.SprintResolucionID, cruzado.SprintResolucion = "s2", 2
	return &lecturaDeDefectosFalsa{
		proyectos: map[string]bool{"p1": true, "vacio": true},
		sprints:   map[string]string{"s1": "p1", "s2": "p1", "s3": "p1", "s9": "p2"},
		defectos: []app.DefectoListado{
			defectoFiltrable("d1", "s1", defecto.SeveridadAlta, defecto.EstadoAbierto, baseDeFiltros),
			defectoFiltrable("d2", "s1", defecto.SeveridadBaja, defecto.EstadoDescartado, baseDeFiltros.Add(time.Minute)),
			resuelto,
			defectoFiltrable("d4", "s2", defecto.SeveridadCritica, defecto.EstadoAbierto, baseDeFiltros.Add(3*time.Minute)),
			defectoFiltrable("d5", "s1", defecto.SeveridadMedia, defecto.EstadoEnCorreccion, baseDeFiltros.Add(4*time.Minute)),
			cruzado,
		},
	}
}

// listarConFiltro corre el caso de uso sobre el proyecto de ejemplo y devuelve los IDs, del mas reciente al mas
// antiguo.
func listarConFiltro(t *testing.T, filtro app.FiltroDefectos) []string {
	t.Helper()
	got, err := app.NuevoListarDefectos(proyectoConDefectos()).Ejecutar(context.Background(), "p1", filtro)
	if err != nil {
		t.Fatalf("Ejecutar(%+v): %v", filtro, err)
	}
	return idsDeDefectos(got)
}

// US-024 / CA-024-5 / RN-024-1 / CL-024-1: sin filtros, o con los dos valores vacios, la lista es la de T-021:
// todos los defectos del proyecto, de cualquier estado, en el mismo orden.
func TestListarDefectos_SinFiltrosEsLaListaCompleta(t *testing.T) {
	esperado := []string{"d6", "d5", "d4", "d3", "d2", "d1"}

	if got := listarConFiltro(t, app.FiltroDefectos{}); !slices.Equal(got, esperado) {
		t.Errorf("sin filtros = %v, quiero %v", got, esperado)
	}
}

// US-024 / CA-024-1 / RN-024-2 / CL-024-2: el filtro por sprint deja solo los defectos detectados en ese sprint,
// de cualquier severidad y en cualquier estado.
func TestListarDefectos_FiltroPorSprint(t *testing.T) {
	casos := []struct {
		nombre string
		sprint string
		quiero []string
	}{
		{"Sprint 1", "s1", []string{"d6", "d5", "d2", "d1"}},
		{"Sprint 2", "s2", []string{"d4", "d3"}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := listarConFiltro(t, app.FiltroDefectos{SprintDeteccionID: c.sprint}); !slices.Equal(got, c.quiero) {
				t.Errorf("filtro por %s = %v, quiero %v", c.sprint, got, c.quiero)
			}
		})
	}
}

// US-024 / RN-024-2 / CL-024-3: el sprint de resolucion no cuenta. El defecto d6 se detecto en el Sprint 1 y se
// resolvio en el Sprint 2: sale al filtrar por el Sprint 1 y no al filtrar por el Sprint 2.
func TestListarDefectos_ElSprintDeResolucionNoCuenta(t *testing.T) {
	if got := listarConFiltro(t, app.FiltroDefectos{SprintDeteccionID: "s1"}); !slices.Contains(got, "d6") {
		t.Errorf("filtro por el Sprint 1 = %v, quiero que incluya d6 (detectado en el 1)", got)
	}
	if got := listarConFiltro(t, app.FiltroDefectos{SprintDeteccionID: "s2"}); slices.Contains(got, "d6") {
		t.Errorf("filtro por el Sprint 2 = %v, no tiene que incluir d6 (solo se resolvio en el 2)", got)
	}
}

// US-024 / CA-024-1 / RN-024-3 / CL-024-4: el filtro por severidad deja solo los defectos de esa severidad, de
// todos los sprints y estados. Se prueba con cada una de las cuatro.
func TestListarDefectos_FiltroPorSeveridad(t *testing.T) {
	casos := []struct {
		severidad defecto.Severidad
		quiero    []string
	}{
		{defecto.SeveridadCritica, []string{"d6", "d4"}},
		{defecto.SeveridadAlta, []string{"d3", "d1"}},
		{defecto.SeveridadMedia, []string{"d5"}},
		{defecto.SeveridadBaja, []string{"d2"}},
	}
	for _, c := range casos {
		t.Run(string(c.severidad), func(t *testing.T) {
			if got := listarConFiltro(t, app.FiltroDefectos{Severidad: c.severidad}); !slices.Equal(got, c.quiero) {
				t.Errorf("filtro por %s = %v, quiero %v", c.severidad, got, c.quiero)
			}
		})
	}
}

// US-024 / CA-024-2 / RN-024-4 / CL-024-5: los dos filtros se combinan con Y. Sprint 1 y alta deja solo d1: no
// salen ni el del Sprint 1 con otra severidad ni el del Sprint 2 con severidad alta.
func TestListarDefectos_LosFiltrosSeCombinanConY(t *testing.T) {
	casos := []struct {
		nombre string
		filtro app.FiltroDefectos
		quiero []string
	}{
		{"Sprint 1 y alta", app.FiltroDefectos{SprintDeteccionID: "s1", Severidad: defecto.SeveridadAlta}, []string{"d1"}},
		{"Sprint 2 y alta", app.FiltroDefectos{SprintDeteccionID: "s2", Severidad: defecto.SeveridadAlta}, []string{"d3"}},
		{"Sprint 1 y critica", app.FiltroDefectos{SprintDeteccionID: "s1", Severidad: defecto.SeveridadCritica}, []string{"d6"}},
		{"Sprint 2 y critica", app.FiltroDefectos{SprintDeteccionID: "s2", Severidad: defecto.SeveridadCritica}, []string{"d4"}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := listarConFiltro(t, c.filtro); !slices.Equal(got, c.quiero) {
				t.Errorf("filtro %+v = %v, quiero %v", c.filtro, got, c.quiero)
			}
		})
	}
}

// US-024 / CA-024-3 / RN-024-5 / CL-024-6 / CL-024-7 / CL-024-8 / CL-024-9: que ningun defecto cumpla el filtro no
// es un error: es una lista vacia (no nula). Vale para un sprint del proyecto sin defectos, una severidad sin
// defectos, dos filtros cuya combinacion no tiene coincidencias aunque cada uno tenga por separado, y un
// proyecto sin defectos.
func TestListarDefectos_SinCoincidenciasEsListaVacia(t *testing.T) {
	sinCritica := proyectoConDefectos()
	sinCritica.defectos = slices.DeleteFunc(sinCritica.defectos, func(l app.DefectoListado) bool {
		return l.Defecto.Severidad == defecto.SeveridadCritica
	})

	casos := []struct {
		nombre   string
		repo     *lecturaDeDefectosFalsa
		proyecto string
		filtro   app.FiltroDefectos
	}{
		{"sprint del proyecto sin defectos", proyectoConDefectos(), "p1", app.FiltroDefectos{SprintDeteccionID: "s3"}},
		{"severidad sin defectos", sinCritica, "p1", app.FiltroDefectos{Severidad: defecto.SeveridadCritica}},
		{"combinacion sin interseccion", sinCritica, "p1", app.FiltroDefectos{SprintDeteccionID: "s1", Severidad: defecto.SeveridadCritica}},
		{"proyecto sin defectos y filtros validos", &lecturaDeDefectosFalsa{
			proyectos: map[string]bool{"vacio": true}, sprints: map[string]string{"s1": "vacio"},
		}, "vacio", app.FiltroDefectos{SprintDeteccionID: "s1", Severidad: defecto.SeveridadAlta}},
		{"proyecto sin defectos y sin filtros", &lecturaDeDefectosFalsa{proyectos: map[string]bool{"vacio": true}}, "vacio", app.FiltroDefectos{}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got, err := app.NuevoListarDefectos(c.repo).Ejecutar(context.Background(), c.proyecto, c.filtro)
			if err != nil || got == nil || len(got) != 0 {
				t.Errorf("Ejecutar = %v, %v; quiero una lista vacia y sin error", got, err)
			}
		})
	}
}

// US-024 / RN-024-6 / CL-024-10: un defecto descartado que cumple el filtro sale: la lista no oculta ningun
// estado, y cambiarle el estado no lo saca de una lista filtrada porque el filtro mira su sprint de deteccion y
// su severidad.
func TestListarDefectos_ElFiltroNoOcultaEstados(t *testing.T) {
	got := listarConFiltro(t, app.FiltroDefectos{SprintDeteccionID: "s1", Severidad: defecto.SeveridadBaja})
	if !slices.Equal(got, []string{"d2"}) {
		t.Errorf("filtro del Sprint 1 y baja = %v, quiero d2 (descartado)", got)
	}
}

// US-024 / RN-024-6 / CL-024-11: el filtro solo quita defectos, no reordena: con distinta fecha de carga el mas
// reciente primero y, con la misma, por ID descendente, sin importar como los entrega el repositorio.
func TestListarDefectos_ElFiltroNoCambiaElOrden(t *testing.T) {
	repo := &lecturaDeDefectosFalsa{
		proyectos: map[string]bool{"p1": true},
		sprints:   map[string]string{"s1": "p1", "s2": "p1"},
		defectos: []app.DefectoListado{
			defectoFiltrable("a", "s1", defecto.SeveridadAlta, defecto.EstadoAbierto, baseDeFiltros),
			defectoFiltrable("x", "s2", defecto.SeveridadAlta, defecto.EstadoAbierto, baseDeFiltros.Add(time.Hour)),
			defectoFiltrable("c", "s1", defecto.SeveridadAlta, defecto.EstadoAbierto, baseDeFiltros),
			defectoFiltrable("b", "s1", defecto.SeveridadAlta, defecto.EstadoAbierto, baseDeFiltros),
			defectoFiltrable("z", "s1", defecto.SeveridadAlta, defecto.EstadoAbierto, baseDeFiltros.Add(2*time.Hour)),
		},
	}

	got, err := app.NuevoListarDefectos(repo).Ejecutar(context.Background(), "p1", app.FiltroDefectos{SprintDeteccionID: "s1"})
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if orden := idsDeDefectos(got); !slices.Equal(orden, []string{"z", "c", "b", "a"}) {
		t.Errorf("orden = %v, quiero z (el mas reciente), luego c, b, a (misma fecha, por ID descendente)", orden)
	}
}

// US-024 / RN-024-6 / CL-024-12: un defecto filtrado es el mismo DefectoListado que en la lista completa: la
// misma historia, el mismo numero de sprint de deteccion y el mismo de resolucion (o 0, sin resolver).
func TestListarDefectos_ElFiltradoEsElMismoDefectoQueEnLaListaCompleta(t *testing.T) {
	ejecutar := func(filtro app.FiltroDefectos) []app.DefectoListado {
		got, err := app.NuevoListarDefectos(proyectoConDefectos()).Ejecutar(context.Background(), "p1", filtro)
		if err != nil {
			t.Fatalf("Ejecutar(%+v): %v", filtro, err)
		}
		return got
	}
	completa := ejecutar(app.FiltroDefectos{})
	filtrada := ejecutar(app.FiltroDefectos{SprintDeteccionID: "s1"})

	if len(filtrada) == 0 || len(filtrada) >= len(completa) {
		t.Fatalf("el filtro por el Sprint 1 devolvio %d de %d defectos, quiero algunos pero no todos", len(filtrada), len(completa))
	}
	for _, f := range filtrada {
		i := slices.IndexFunc(completa, func(l app.DefectoListado) bool { return l.Defecto.ID == f.Defecto.ID })
		if i < 0 || completa[i] != f {
			t.Errorf("el defecto filtrado %+v no es igual al de la lista completa", f)
		}
	}
	if d6 := filtrada[0]; d6.Defecto.ID != "d6" || d6.SprintDeteccion != 1 || d6.SprintResolucion != 2 {
		t.Errorf("d6 = %+v, quiero detectado en el Sprint 1 y resuelto en el 2", d6)
	}
}

// US-024 / CA-024-4 / RN-024-7 / CL-024-13 / CL-024-14: un sprint que no es del proyecto, que no existe o que esta
// mal formado es ErrSprintInexistente envuelto con el ID, y no una lista vacia. Un sprint del proyecto se acepta
// cualquiera sea su estado: el caso de uso solo pregunta si es del proyecto (CL-024-15).
func TestListarDefectos_SprintInexistente(t *testing.T) {
	casos := []struct{ nombre, sprint string }{
		{"sprint de otro proyecto", "s9"},
		{"sprint que no existe", "00000000-0000-0000-0000-000000000000"},
		{"sprint mal formado", "abc"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := proyectoConDefectos()

			got, err := app.NuevoListarDefectos(repo).Ejecutar(context.Background(), "p1", app.FiltroDefectos{SprintDeteccionID: c.sprint})
			if !errors.Is(err, app.ErrSprintInexistente) || !strings.Contains(err.Error(), c.sprint) {
				t.Errorf("Ejecutar = %v, %v; quiero ErrSprintInexistente envuelto con el ID %q", got, err, c.sprint)
			}
			if got != nil {
				t.Errorf("un rechazo no devuelve lista, devolvio %v", got)
			}
			if repo.lecturas != 0 {
				t.Errorf("leyo la lista %d veces: el rechazo tiene que ocurrir antes de leer (CL-024-19)", repo.lecturas)
			}
		})
	}
}

// US-024 / CA-024-4 / RN-024-8 / CL-024-16: una severidad que no es una de las cuatro, exacta, es
// ErrSeveridadInvalida envuelto con el valor recibido. No se pasa a minuscula ni se recorta: " alta " y un solo
// espacio son invalidas porque no estan vacias. El caso de uso rechaza antes de leer la lista.
func TestListarDefectos_SeveridadInvalida(t *testing.T) {
	for _, severidad := range []defecto.Severidad{"Alta", "CRITICA", "crítica", " alta ", " ", "urgente"} {
		t.Run(string(severidad), func(t *testing.T) {
			repo := proyectoConDefectos()

			got, err := app.NuevoListarDefectos(repo).Ejecutar(context.Background(), "p1", app.FiltroDefectos{Severidad: severidad})
			if !errors.Is(err, defecto.ErrSeveridadInvalida) || !strings.Contains(err.Error(), `"`+string(severidad)+`"`) {
				t.Errorf("Ejecutar = %v, %v; quiero ErrSeveridadInvalida envuelto con %q", got, err, severidad)
			}
			if got != nil || repo.lecturas != 0 {
				t.Errorf("un rechazo no lee ni devuelve la lista: devolvio %v tras %d lecturas", got, repo.lecturas)
			}
		})
	}
}

// US-024 / CA-024-6 / RN-024-9 / CL-024-17: con el sprint inexistente y la severidad invalida a la vez se
// devuelve solo ErrSprintInexistente, sin unir los dos errores. Con el sprint corregido y la misma severidad, se
// devuelve ErrSeveridadInvalida.
func TestListarDefectos_SprintInexistenteYSeveridadInvalida(t *testing.T) {
	repo := proyectoConDefectos()
	ejecutor := app.NuevoListarDefectos(repo)

	_, err := ejecutor.Ejecutar(context.Background(), "p1", app.FiltroDefectos{SprintDeteccionID: "s9", Severidad: "urgente"})
	if !errors.Is(err, app.ErrSprintInexistente) || errors.Is(err, defecto.ErrSeveridadInvalida) {
		t.Errorf("Ejecutar = %v, quiero solo ErrSprintInexistente", err)
	}

	_, err = ejecutor.Ejecutar(context.Background(), "p1", app.FiltroDefectos{SprintDeteccionID: "s1", Severidad: "urgente"})
	if !errors.Is(err, defecto.ErrSeveridadInvalida) || errors.Is(err, app.ErrSprintInexistente) {
		t.Errorf("con el sprint corregido: Ejecutar = %v, quiero solo ErrSeveridadInvalida", err)
	}
}

// US-024 / RN-024-9 / CL-024-18: con el proyecto inexistente, un sprint tambien inexistente y una severidad
// invalida, se devuelve solo ErrProyectoInexistente, y ni siquiera se consulta el sprint.
func TestListarDefectos_ProyectoInexistenteConFiltrosInvalidos(t *testing.T) {
	repo := proyectoConDefectos()

	_, err := app.NuevoListarDefectos(repo).Ejecutar(context.Background(), "otro", app.FiltroDefectos{SprintDeteccionID: "s9", Severidad: "urgente"})
	if !errors.Is(err, app.ErrProyectoInexistente) || errors.Is(err, app.ErrSprintInexistente) || errors.Is(err, defecto.ErrSeveridadInvalida) {
		t.Errorf("Ejecutar = %v, quiero solo ErrProyectoInexistente", err)
	}
	if repo.consultasDeSprint != 0 || repo.lecturas != 0 {
		t.Errorf("consulto el sprint %d veces y leyo la lista %d: el proyecto se verifica primero y solo", repo.consultasDeSprint, repo.lecturas)
	}
}

// US-024 / RN-024-7 / CL-024-1: sin sprint informado no se verifica nada: un filtro solo por severidad no consulta
// los sprints.
func TestListarDefectos_SinSprintNoSeVerificaNada(t *testing.T) {
	repo := proyectoConDefectos()

	if _, err := app.NuevoListarDefectos(repo).Ejecutar(context.Background(), "p1", app.FiltroDefectos{Severidad: defecto.SeveridadAlta}); err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if repo.consultasDeSprint != 0 {
		t.Errorf("consulto el sprint %d veces sin que se informara ninguno", repo.consultasDeSprint)
	}
}

// US-024 / CL-024-20: una falla de la base al verificar el sprint o al listar, con o sin filtro, se devuelve
// envuelta y sin tocarla (no se confunde con un sprint inexistente), y ninguna lista parcial.
func TestListarDefectos_FallaDeLaBaseConFiltro(t *testing.T) {
	caida := errors.New("base caida")
	casos := []struct {
		nombre string
		filtro app.FiltroDefectos
		falla  func(*lecturaDeDefectosFalsa)
	}{
		{"al verificar el sprint", app.FiltroDefectos{SprintDeteccionID: "s1"}, func(l *lecturaDeDefectosFalsa) { l.errSprint = caida }},
		{"al listar con filtro de sprint", app.FiltroDefectos{SprintDeteccionID: "s1"}, func(l *lecturaDeDefectosFalsa) { l.errListar = caida }},
		{"al listar con filtro de severidad", app.FiltroDefectos{Severidad: defecto.SeveridadAlta}, func(l *lecturaDeDefectosFalsa) { l.errListar = caida }},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := proyectoConDefectos()
			c.falla(repo)

			got, err := app.NuevoListarDefectos(repo).Ejecutar(context.Background(), "p1", c.filtro)
			if !errors.Is(err, caida) || errors.Is(err, app.ErrSprintInexistente) || got != nil {
				t.Errorf("Ejecutar = %v, %v; quiero la falla de la base envuelta y ninguna lista", got, err)
			}
		})
	}
}
