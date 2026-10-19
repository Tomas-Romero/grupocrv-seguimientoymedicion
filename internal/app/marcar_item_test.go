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

// cambioGuardado es una llamada a GuardarCambioDeEstado que anota el repositorio falso.
type cambioGuardado struct{ itemID, sprintID string }

// repositorioMarcadoFalso simula la persistencia del marcado sin base de datos. Al
// guardar aplica el cambio al item que tiene, como el repositorio real lo aplica al
// item que lee con el sprint bloqueado.
type repositorioMarcadoFalso struct {
	item         backlog.ItemBacklog
	existeItem   bool
	sprint       sprint.Sprint
	existeSprint bool
	enElSprint   bool

	errObtenerItem   error
	errObtenerSprint error
	errEnElSprint    error
	errGuardar       error

	// enLaBase, si no es nil, es el item que el repositorio ve dentro de la
	// transaccion, cuando cambio despues de la lectura del caso de uso.
	enLaBase  *backlog.ItemBacklog
	guardados []cambioGuardado
}

func (r *repositorioMarcadoFalso) ObtenerItem(_ context.Context, _ string) (backlog.ItemBacklog, bool, error) {
	return r.item, r.existeItem, r.errObtenerItem
}

func (r *repositorioMarcadoFalso) ObtenerSprint(_ context.Context, _ string) (sprint.Sprint, bool, error) {
	return r.sprint, r.existeSprint, r.errObtenerSprint
}

func (r *repositorioMarcadoFalso) ItemEnSprint(_ context.Context, _, _ string) (bool, error) {
	return r.enElSprint, r.errEnElSprint
}

func (r *repositorioMarcadoFalso) GuardarCambioDeEstado(_ context.Context, itemID, sprintID string, cambio app.CambioDeEstado) (backlog.ItemBacklog, error) {
	r.guardados = append(r.guardados, cambioGuardado{itemID: itemID, sprintID: sprintID})
	if r.errGuardar != nil {
		return backlog.ItemBacklog{}, r.errGuardar
	}
	actual := r.item
	if r.enLaBase != nil {
		actual = *r.enLaBase
	}
	return cambio(actual), nil
}

// enElSprintAbierto es un repositorio con el item asignado al sprint abierto.
func enElSprintAbierto(t *testing.T, estado backlog.Estado) *repositorioMarcadoFalso {
	t.Helper()
	item := itemListoParaSprint(t)
	item.Estado = estado
	if estado == backlog.EstadoCompletado {
		item.CompletadoEn = time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC)
	}
	return &repositorioMarcadoFalso{
		item: item, existeItem: true,
		sprint: sprintAbiertoDelProyecto(), existeSprint: true,
		enElSprint: true,
	}
}

// US-010 / CA-010-1, RN-010-5, RN-010-8: marcar deja el item completado con la fecha
// del reloj inyectado y guarda el cambio una sola vez, en ese sprint.
func TestMarcarCompletado_Marca(t *testing.T) {
	repo := enElSprintAbierto(t, backlog.EstadoPendiente)

	item, err := app.NuevoMarcarCompletado(repo, relojFijo).Ejecutar(context.Background(), "item-1", "sprint-1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if item.Estado != backlog.EstadoCompletado || !item.CompletadoEn.Equal(ahora) {
		t.Errorf("item = (%q, %v), se esperaba (completado, %v)", item.Estado, item.CompletadoEn, ahora)
	}
	if len(repo.guardados) != 1 || repo.guardados[0] != (cambioGuardado{itemID: "item-1", sprintID: "sprint-1"}) {
		t.Errorf("se guardo %+v, se esperaba un cambio del item 1 en el sprint 1", repo.guardados)
	}
}

// US-010 / CA-010-3, RN-010-6, CL-010-2: marcar un item ya completado no es un error y
// conserva la fecha original.
func TestMarcarCompletado_YaCompletado(t *testing.T) {
	repo := enElSprintAbierto(t, backlog.EstadoCompletado)
	original := repo.item.CompletadoEn

	item, err := app.NuevoMarcarCompletado(repo, relojFijo).Ejecutar(context.Background(), "item-1", "sprint-1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if !item.CompletadoEn.Equal(original) {
		t.Errorf("completado en = %v, se esperaba la fecha original %v", item.CompletadoEn, original)
	}
}

// US-010 / CL-010-10: el item que se devuelve es el que quedo en la base. Si otro
// marcado lo completo despues de la lectura, queda la fecha del primero.
func TestMarcarCompletado_DevuelveLoQueQuedoEnLaBase(t *testing.T) {
	repo := enElSprintAbierto(t, backlog.EstadoPendiente)
	primero := repo.item.Completar(time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC))
	repo.enLaBase = &primero

	item, err := app.NuevoMarcarCompletado(repo, relojFijo).Ejecutar(context.Background(), "item-1", "sprint-1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if !item.CompletadoEn.Equal(primero.CompletadoEn) {
		t.Errorf("completado en = %v, se esperaba la del primer marcado %v", item.CompletadoEn, primero.CompletadoEn)
	}
}

// US-010 / CA-010-2, RN-010-7, CL-010-3: desmarcar deja el item pendiente y sin fecha.
func TestDesmarcarCompletado_Desmarca(t *testing.T) {
	repo := enElSprintAbierto(t, backlog.EstadoCompletado)

	item, err := app.NuevoDesmarcarCompletado(repo).Ejecutar(context.Background(), "item-1", "sprint-1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if item.Estado != backlog.EstadoPendiente || !item.CompletadoEn.IsZero() {
		t.Errorf("item = (%q, %v), se esperaba (pendiente, sin fecha)", item.Estado, item.CompletadoEn)
	}
	if len(repo.guardados) != 1 {
		t.Errorf("se guardaron %d cambios, se esperaba 1", len(repo.guardados))
	}
}

// US-010 / RN-010-7, CL-010-4: desmarcar un item que no esta completado no es un error
// y no cambia nada.
func TestDesmarcarCompletado_NoCompletado(t *testing.T) {
	repo := enElSprintAbierto(t, backlog.EstadoEnProgreso)

	item, err := app.NuevoDesmarcarCompletado(repo).Ejecutar(context.Background(), "item-1", "sprint-1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if item.Estado != backlog.EstadoEnProgreso {
		t.Errorf("estado = %q, se esperaba que siguiera en progreso", item.Estado)
	}
}

// resultadoOperacion es lo que deja marcar o desmarcar: el error y el repositorio que
// uso, para ver que guardo.
type resultadoOperacion struct {
	err  error
	repo *repositorioMarcadoFalso
}

// ejecutarLasDos corre marcar y desmarcar, cada una con un repositorio nuevo, y
// devuelve el resultado de cada una.
func ejecutarLasDos(nuevoRepo func() *repositorioMarcadoFalso) map[string]resultadoOperacion {
	marcar := nuevoRepo()
	_, errMarcar := app.NuevoMarcarCompletado(marcar, relojFijo).Ejecutar(context.Background(), "item-1", "sprint-1")
	desmarcar := nuevoRepo()
	_, errDesmarcar := app.NuevoDesmarcarCompletado(desmarcar).Ejecutar(context.Background(), "item-1", "sprint-1")
	return map[string]resultadoOperacion{
		"marcar":    {err: errMarcar, repo: marcar},
		"desmarcar": {err: errDesmarcar, repo: desmarcar},
	}
}

// US-010 / CA-010-6, RN-010-1, CL-010-8: si el item o el sprint no existen, se
// devuelve ese error sin validar nada mas y no se guarda nada. Si faltan los dos,
// solo ErrItemInexistente.
func TestMarcarYDesmarcar_LoQueNoExiste(t *testing.T) {
	casos := []struct {
		nombre       string
		existeItem   bool
		existeSprint bool
		errEsperado  error
		otro         error
	}{
		{"item inexistente", false, true, app.ErrItemInexistente, app.ErrSprintInexistente},
		{"sprint inexistente", true, false, app.ErrSprintInexistente, app.ErrItemInexistente},
		{"los dos inexistentes", false, false, app.ErrItemInexistente, app.ErrSprintInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			resultados := ejecutarLasDos(func() *repositorioMarcadoFalso {
				repo := enElSprintAbierto(t, backlog.EstadoPendiente)
				repo.existeItem, repo.existeSprint = c.existeItem, c.existeSprint
				// Ademas no esta en el sprint: si se validara, el error seria otro.
				repo.enElSprint = false
				return repo
			})
			for operacion, r := range resultados {
				if !errors.Is(r.err, c.errEsperado) {
					t.Errorf("%s: error = %v, se esperaba %v", operacion, r.err, c.errEsperado)
				}
				if errors.Is(r.err, c.otro) || errors.Is(r.err, sprint.ErrItemFueraDelSprint) {
					t.Errorf("%s: el error %q se unio con otros: la inexistencia va sola", operacion, r.err)
				}
				if len(r.repo.guardados) != 0 {
					t.Errorf("%s: se guardaron %d cambios, se esperaba ninguno", operacion, len(r.repo.guardados))
				}
			}
		})
	}
}

// US-010 / CA-010-4, CA-010-5, RN-010-2, RN-010-3, RN-010-4, RN-010-10: un item fuera
// del sprint o un sprint cerrado se rechazan con el error del dominio, sin guardar
// nada. Con el sprint cerrado, el unico error es ErrSprintCerrado.
func TestMarcarYDesmarcar_ReglasDelDominio(t *testing.T) {
	casos := []struct {
		nombre      string
		cerrado     bool
		enElSprint  bool
		errEsperado error
		noEsperado  error
	}{
		{"item fuera del sprint", false, false, sprint.ErrItemFueraDelSprint, sprint.ErrSprintCerrado},
		{"sprint cerrado", true, true, sprint.ErrSprintCerrado, sprint.ErrItemFueraDelSprint},
		{"sprint cerrado e item fuera", true, false, sprint.ErrSprintCerrado, sprint.ErrItemFueraDelSprint},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			resultados := ejecutarLasDos(func() *repositorioMarcadoFalso {
				repo := enElSprintAbierto(t, backlog.EstadoCompletado)
				repo.enElSprint = c.enElSprint
				if c.cerrado {
					repo.sprint.Estado = sprint.EstadoCerrado
				}
				return repo
			})
			for operacion, r := range resultados {
				if !errors.Is(r.err, c.errEsperado) {
					t.Errorf("%s: error = %v, se esperaba %v", operacion, r.err, c.errEsperado)
				}
				if errors.Is(r.err, c.noEsperado) {
					t.Errorf("%s: el error %q no deberia incluir %v", operacion, r.err, c.noEsperado)
				}
				if len(r.repo.guardados) != 0 {
					t.Errorf("%s: se guardaron %d cambios, se esperaba ninguno", operacion, len(r.repo.guardados))
				}
			}
		})
	}
}

// US-010 / RN-010-9, CL-010-9: lo que el repositorio rechaza dentro de la transaccion
// (el sprint se cerro, o el item o el sprint desaparecieron en el medio) se devuelve
// reconocible y sin item.
func TestMarcarYDesmarcar_RechazosDelRepositorio(t *testing.T) {
	for _, errGuardar := range []error{sprint.ErrSprintCerrado, sprint.ErrItemFueraDelSprint, app.ErrItemInexistente, app.ErrSprintInexistente} {
		t.Run(errGuardar.Error(), func(t *testing.T) {
			resultados := ejecutarLasDos(func() *repositorioMarcadoFalso {
				repo := enElSprintAbierto(t, backlog.EstadoPendiente)
				repo.errGuardar = errGuardar
				return repo
			})
			for operacion, r := range resultados {
				if !errors.Is(r.err, errGuardar) {
					t.Errorf("%s: error = %v, se esperaba %v", operacion, r.err, errGuardar)
				}
			}
		})
	}
}

// Un fallo del repositorio en cualquiera de las lecturas o al guardar se devuelve
// envuelto con el ID y no es un inexistente.
func TestMarcarYDesmarcar_FallosDeInfraestructura(t *testing.T) {
	falla := errors.New("se corto la conexion")
	casos := []struct {
		nombre string
		romper func(*repositorioMarcadoFalso)
		id     string
	}{
		{"al leer el item", func(r *repositorioMarcadoFalso) { r.errObtenerItem = falla }, "item-1"},
		{"al leer el sprint", func(r *repositorioMarcadoFalso) { r.errObtenerSprint = falla }, "sprint-1"},
		{"al consultar si esta en el sprint", func(r *repositorioMarcadoFalso) { r.errEnElSprint = falla }, "item-1"},
		{"al guardar", func(r *repositorioMarcadoFalso) { r.errGuardar = falla }, "item-1"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			resultados := ejecutarLasDos(func() *repositorioMarcadoFalso {
				repo := enElSprintAbierto(t, backlog.EstadoPendiente)
				c.romper(repo)
				return repo
			})
			for operacion, r := range resultados {
				if !errors.Is(r.err, falla) {
					t.Fatalf("%s: error = %v, se esperaba el del repositorio", operacion, r.err)
				}
				if !strings.Contains(r.err.Error(), c.id) {
					t.Errorf("%s: el error %q no indica %s", operacion, r.err, c.id)
				}
				if errors.Is(r.err, app.ErrItemInexistente) || errors.Is(r.err, app.ErrSprintInexistente) {
					t.Errorf("%s: un fallo del repositorio no es un inexistente: %v", operacion, r.err)
				}
			}
		})
	}
}
