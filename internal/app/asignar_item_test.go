package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// asignacionGuardada es una llamada a AsignarItemASprint que anota el repositorio falso.
type asignacionGuardada struct{ itemID, sprintID string }

// repositorioAsignacionesFalso simula la persistencia de las asignaciones sin base de
// datos. Responde lo que se le configura y anota lo que le piden.
type repositorioAsignacionesFalso struct {
	item          backlog.ItemBacklog
	existeItem    bool
	sprint        sprint.Sprint
	existeSprint  bool
	enAbierto     bool
	errItem       error
	errSprint     error
	errEnAbierto  error
	errAsignar    error
	llamadas      []string
	asignaciones  []asignacionGuardada
	consultadoEn  string
	idsConsultado [2]string
}

func (r *repositorioAsignacionesFalso) ObtenerItem(_ context.Context, itemID string) (backlog.ItemBacklog, bool, error) {
	r.llamadas = append(r.llamadas, "item")
	r.idsConsultado[0] = itemID
	return r.item, r.existeItem, r.errItem
}

func (r *repositorioAsignacionesFalso) ObtenerSprint(_ context.Context, sprintID string) (sprint.Sprint, bool, error) {
	r.llamadas = append(r.llamadas, "sprint")
	r.idsConsultado[1] = sprintID
	return r.sprint, r.existeSprint, r.errSprint
}

func (r *repositorioAsignacionesFalso) ItemEnSprintAbierto(_ context.Context, itemID string) (bool, error) {
	r.llamadas = append(r.llamadas, "abierto")
	r.consultadoEn = itemID
	return r.enAbierto, r.errEnAbierto
}

func (r *repositorioAsignacionesFalso) AsignarItemASprint(_ context.Context, itemID, sprintID string) error {
	r.llamadas = append(r.llamadas, "asignar")
	r.asignaciones = append(r.asignaciones, asignacionGuardada{itemID, sprintID})
	return r.errAsignar
}

// itemListoParaSprint es un item del proyecto p1 que cumple la Definition of Ready: estimado
// (5 puntos) y con un criterio.
func itemListoParaSprint(t *testing.T) backlog.ItemBacklog {
	t.Helper()
	item, err := backlog.NuevoItem(backlog.DatosItem{Titulo: "Crear proyecto", Prioridad: backlog.PrioridadMust, Criterios: []string{"Se guarda"}}, time.Now())
	if err != nil {
		t.Fatalf("NuevoItem: %v", err)
	}
	if item, err = item.Estimar(5); err != nil {
		t.Fatalf("Estimar: %v", err)
	}
	item.ID, item.ProyectoID = "item-1", "p1"
	return item
}

func sprintAbiertoDelProyecto() sprint.Sprint {
	return sprint.Sprint{ID: "sprint-1", Numero: 2, ProyectoID: "p1", Estado: sprint.EstadoAbierto}
}

// todoListo es un repositorio con el item y el sprint existentes y compatibles.
func todoListo(t *testing.T) *repositorioAsignacionesFalso {
	t.Helper()
	return &repositorioAsignacionesFalso{
		item: itemListoParaSprint(t), existeItem: true,
		sprint: sprintAbiertoDelProyecto(), existeSprint: true,
	}
}

// US-009 / CA-009-1, RN-009-1, RN-009-2, RN-009-9: con todo en regla el caso de uso lee
// el item, el sprint y si el item ya esta en un sprint abierto, valida con el
// dominio y asigna una sola vez. Devuelve la asignacion.
func TestAsignarItemASprint_Asigna(t *testing.T) {
	repo := todoListo(t)

	a, err := app.NuevoAsignarItemASprint(repo).Ejecutar(context.Background(), "item-1", "sprint-1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if a != (app.Asignacion{ItemID: "item-1", SprintID: "sprint-1"}) {
		t.Errorf("asignacion = %+v, se esperaba el item 1 en el sprint 1", a)
	}
	if esperado := "item,sprint,abierto,asignar"; strings.Join(repo.llamadas, ",") != esperado {
		t.Errorf("orden de llamadas = %v, se esperaba %s", repo.llamadas, esperado)
	}
	if len(repo.asignaciones) != 1 || repo.asignaciones[0] != (asignacionGuardada{"item-1", "sprint-1"}) {
		t.Errorf("asignaciones = %+v, se esperaba una del item 1 al sprint 1", repo.asignaciones)
	}
}

// US-009 / RN-009-1, RN-009-2, CL-009-11, CA-009-8: si el item o el sprint no existen se
// devuelve solo ese error, con el ID recibido, sin mirar el resto (aca el item no
// cumple la DoR) y sin asignar nada. El item se verifica primero: si faltan los dos,
// solo ErrItemInexistente.
func TestAsignarItemASprint_LoQueNoExiste(t *testing.T) {
	casos := []struct {
		nombre    string
		repo      func() *repositorioAsignacionesFalso
		esperado  error
		idEnMsg   string
		noEsperan []error
		llamadas  string
	}{
		{"el item no existe", func() *repositorioAsignacionesFalso {
			return &repositorioAsignacionesFalso{existeSprint: true, sprint: sprintAbiertoDelProyecto()}
		}, app.ErrItemInexistente, "item-1", []error{app.ErrSprintInexistente}, "item"},
		{"el sprint no existe", func() *repositorioAsignacionesFalso {
			item := backlog.ItemBacklog{ID: "item-1", ProyectoID: "p1"} // sin estimar y sin criterios
			return &repositorioAsignacionesFalso{existeItem: true, item: item}
		}, app.ErrSprintInexistente, "sprint-1", []error{sprint.ErrItemSinEstimar, sprint.ErrItemSinCriterios}, "item,sprint"},
		{"no existe ninguno de los dos", func() *repositorioAsignacionesFalso { return &repositorioAsignacionesFalso{} },
			app.ErrItemInexistente, "item-1", []error{app.ErrSprintInexistente}, "item"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := c.repo()

			_, err := app.NuevoAsignarItemASprint(repo).Ejecutar(context.Background(), "item-1", "sprint-1")

			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.esperado)
			}
			for _, e := range c.noEsperan {
				if errors.Is(err, e) {
					t.Errorf("el error %q se unio con %v", err, e)
				}
			}
			if !strings.Contains(err.Error(), c.idEnMsg) {
				t.Errorf("el error %q no indica el ID %s", err, c.idEnMsg)
			}
			if strings.Join(repo.llamadas, ",") != c.llamadas {
				t.Errorf("llamadas = %v, se esperaba %s", repo.llamadas, c.llamadas)
			}
			if len(repo.asignaciones) != 0 {
				t.Errorf("se asignaron %d items, se esperaba ninguno", len(repo.asignaciones))
			}
		})
	}
}

// US-009 / RN-009-3 a RN-009-8, RN-009-10: el caso de uso arma los datos minimos con lo
// que leyo y valida con el dominio: cada regla se rechaza con su error, unidos, y no
// se asigna nada.
func TestAsignarItemASprint_ReglasDelDominio(t *testing.T) {
	sinEstimarNiCriterios := backlog.ItemBacklog{ID: "item-1", ProyectoID: "p1"}
	completado := itemListoParaSprint(t)
	completado.Estado = backlog.EstadoCompletado
	deOtroProyecto := itemListoParaSprint(t)
	deOtroProyecto.ProyectoID = "otro"
	cerrado := sprintAbiertoDelProyecto()
	cerrado.Estado = sprint.EstadoCerrado

	casos := []struct {
		nombre    string
		item      backlog.ItemBacklog
		sprint    sprint.Sprint
		enAbierto bool
		esperados []error
	}{
		{"sin estimar y sin criterios", sinEstimarNiCriterios, sprintAbiertoDelProyecto(), false, []error{sprint.ErrItemSinEstimar, sprint.ErrItemSinCriterios}},
		{"item completado", completado, sprintAbiertoDelProyecto(), false, []error{backlog.ErrItemCompletado}},
		{"item ya asignado", itemListoParaSprint(t), sprintAbiertoDelProyecto(), true, []error{sprint.ErrItemYaAsignado}},
		{"proyectos distintos", deOtroProyecto, sprintAbiertoDelProyecto(), false, []error{sprint.ErrProyectoDistinto}},
		{"sprint cerrado", itemListoParaSprint(t), cerrado, false, []error{sprint.ErrSprintCerrado}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := &repositorioAsignacionesFalso{item: c.item, existeItem: true, sprint: c.sprint, existeSprint: true, enAbierto: c.enAbierto}

			_, err := app.NuevoAsignarItemASprint(repo).Ejecutar(context.Background(), "item-1", "sprint-1")

			for _, e := range c.esperados {
				if !errors.Is(err, e) {
					t.Errorf("error = %v, se esperaba que incluyera %v", err, e)
				}
			}
			if len(repo.asignaciones) != 0 {
				t.Errorf("se asignaron %d items, se esperaba ninguno", len(repo.asignaciones))
			}
		})
	}
}

// US-009 / CL-009-1: un item en progreso (no completado) se asigna: la regla es no estar
// completado, no estar pendiente.
func TestAsignarItemASprint_ItemEnProgresoSeAsigna(t *testing.T) {
	repo := todoListo(t)
	repo.item.Estado = backlog.EstadoEnProgreso

	if _, err := app.NuevoAsignarItemASprint(repo).Ejecutar(context.Background(), "item-1", "sprint-1"); err != nil {
		t.Errorf("error = %v, se esperaba que un item en progreso se asigne", err)
	}
}

// Los errores que devuelve el repositorio dentro de la transaccion (el estado cambio
// entre la lectura y el alta) llegan reconocibles con errors.Is.
func TestAsignarItemASprint_ErroresDelRepositorioAlAsignar(t *testing.T) {
	for _, e := range []error{sprint.ErrItemYaAsignado, sprint.ErrSprintCerrado, app.ErrItemInexistente, app.ErrSprintInexistente} {
		repo := todoListo(t)
		repo.errAsignar = e

		_, err := app.NuevoAsignarItemASprint(repo).Ejecutar(context.Background(), "item-1", "sprint-1")

		if !errors.Is(err, e) {
			t.Errorf("error = %v, se esperaba %v", err, e)
		}
	}
}

// Un fallo de lectura o de escritura se devuelve envuelto, con el ID del item o del sprint, y sigue
// reconociendose con errors.Is. No devuelve asignacion.
func TestAsignarItemASprint_FallosDeInfraestructura(t *testing.T) {
	falla := errors.New("se corto la conexion")
	casos := []struct {
		nombre string
		ajuste func(r *repositorioAsignacionesFalso)
	}{
		{"al leer el item", func(r *repositorioAsignacionesFalso) { r.errItem = falla }},
		{"al leer el sprint", func(r *repositorioAsignacionesFalso) { r.errSprint = falla }},
		{"al consultar si el item ya esta asignado", func(r *repositorioAsignacionesFalso) { r.errEnAbierto = falla }},
		{"al asignar", func(r *repositorioAsignacionesFalso) { r.errAsignar = falla }},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := todoListo(t)
			c.ajuste(repo)

			a, err := app.NuevoAsignarItemASprint(repo).Ejecutar(context.Background(), "item-1", "sprint-1")

			if !errors.Is(err, falla) {
				t.Fatalf("error = %v, se esperaba el del repositorio", err)
			}
			if !strings.Contains(err.Error(), "item-1") && !strings.Contains(err.Error(), "sprint-1") {
				t.Errorf("el error %q no indica el item ni el sprint", err)
			}
			if a != (app.Asignacion{}) {
				t.Errorf("devolvio la asignacion %+v, se esperaba el valor cero", a)
			}
		})
	}
}
