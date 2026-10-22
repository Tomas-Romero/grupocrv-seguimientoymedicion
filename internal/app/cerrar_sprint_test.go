package app_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// repositorioCierreFalso simula la persistencia del cierre sin base de datos. Al
// guardar devuelve los resultados que calcula el dominio con sus historias, como hace
// el repositorio real con las que lee con el sprint bloqueado.
type repositorioCierreFalso struct {
	sprint       sprint.Sprint
	existeSprint bool
	historias    []sprint.HistoriaDelCierre

	errObtener   error
	errHistorias error
	errGuardar   error

	guardados []string
}

func (r *repositorioCierreFalso) ObtenerSprint(_ context.Context, _ string) (sprint.Sprint, bool, error) {
	return r.sprint, r.existeSprint, r.errObtener
}

func (r *repositorioCierreFalso) HistoriasDelSprint(_ context.Context, _ string) ([]sprint.HistoriaDelCierre, error) {
	return r.historias, r.errHistorias
}

func (r *repositorioCierreFalso) GuardarCierre(_ context.Context, sprintID string) ([]sprint.ResultadoDelCierre, error) {
	r.guardados = append(r.guardados, sprintID)
	if r.errGuardar != nil {
		return nil, r.errGuardar
	}
	return sprint.Cerrar(sprint.DatosDelSprint{Estado: r.sprint.Estado, Nombre: r.sprint.Nombre()}, r.historias)
}

func sprintParaCerrar() *repositorioCierreFalso {
	return &repositorioCierreFalso{
		sprint:       sprintAbiertoDelProyecto(),
		existeSprint: true,
		historias: []sprint.HistoriaDelCierre{
			{ItemID: "item-1", Completada: true},
			{ItemID: "item-2", Completada: false},
		},
	}
}

// US-011 / CA-011-1, RN-011-5: el caso de uso lee el sprint, valida con el dominio y
// guarda el cierre una sola vez. Devuelve el resultado de cada historia.
func TestCerrarSprint_Cierra(t *testing.T) {
	repo := sprintParaCerrar()

	resultados, err := app.NuevoCerrarSprint(repo).Ejecutar(context.Background(), "sprint-1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	esperados := []sprint.ResultadoDelCierre{
		{ItemID: "item-1", Resultado: sprint.ResultadoCompletado},
		{ItemID: "item-2", Resultado: sprint.ResultadoNoCompletado},
	}
	if !slices.Equal(resultados, esperados) {
		t.Errorf("resultados = %+v, se esperaban %+v", resultados, esperados)
	}
	if !slices.Equal(repo.guardados, []string{"sprint-1"}) {
		t.Errorf("se guardo %v, se esperaba un cierre del sprint 1", repo.guardados)
	}
}

// US-011 / CA-011-4, RN-011-4: un sprint sin historias se cierra.
func TestCerrarSprint_SinHistorias(t *testing.T) {
	repo := sprintParaCerrar()
	repo.historias = nil

	resultados, err := app.NuevoCerrarSprint(repo).Ejecutar(context.Background(), "sprint-1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if len(resultados) != 0 || len(repo.guardados) != 1 {
		t.Errorf("resultados = %+v y %d cierres, se esperaba ninguno y 1", resultados, len(repo.guardados))
	}
}

// US-011 / CA-011-6, RN-011-1, CL-011-7: si el sprint no existe, se devuelve
// ErrSprintInexistente con el ID, sin leer sus historias ni guardar nada.
func TestCerrarSprint_SprintInexistente(t *testing.T) {
	repo := sprintParaCerrar()
	repo.existeSprint = false
	repo.errHistorias = errors.New("no deberia leer las historias")

	_, err := app.NuevoCerrarSprint(repo).Ejecutar(context.Background(), "no-existe")

	if !errors.Is(err, app.ErrSprintInexistente) {
		t.Fatalf("error = %v, se esperaba ErrSprintInexistente", err)
	}
	if !strings.Contains(err.Error(), "no-existe") {
		t.Errorf("el error %q no indica el ID recibido", err)
	}
	if len(repo.guardados) != 0 {
		t.Errorf("se guardaron %d cierres, se esperaba ninguno", len(repo.guardados))
	}
}

// US-011 / CA-011-5, RN-011-2, CL-011-6: un sprint ya cerrado se rechaza con el error
// del dominio, sin guardar nada.
func TestCerrarSprint_SprintCerrado(t *testing.T) {
	repo := sprintParaCerrar()
	repo.sprint.Estado = sprint.EstadoCerrado

	_, err := app.NuevoCerrarSprint(repo).Ejecutar(context.Background(), "sprint-1")

	if !errors.Is(err, sprint.ErrSprintCerrado) {
		t.Fatalf("error = %v, se esperaba ErrSprintCerrado", err)
	}
	if len(repo.guardados) != 0 {
		t.Errorf("se guardaron %d cierres, se esperaba ninguno", len(repo.guardados))
	}
}

// US-011 / RN-011-8, RN-011-9, CL-011-10: lo que el repositorio rechaza dentro de la
// transaccion (otro cierre llego antes, o el sprint desaparecio) se devuelve
// reconocible y sin resultados.
func TestCerrarSprint_RechazosDelRepositorio(t *testing.T) {
	for _, errGuardar := range []error{sprint.ErrSprintCerrado, app.ErrSprintInexistente} {
		t.Run(errGuardar.Error(), func(t *testing.T) {
			repo := sprintParaCerrar()
			repo.errGuardar = errGuardar

			resultados, err := app.NuevoCerrarSprint(repo).Ejecutar(context.Background(), "sprint-1")

			if !errors.Is(err, errGuardar) {
				t.Errorf("error = %v, se esperaba %v", err, errGuardar)
			}
			if resultados != nil {
				t.Errorf("resultados = %+v, se esperaba ninguno", resultados)
			}
		})
	}
}

// Un fallo del repositorio al leer el sprint, sus historias o al guardar se devuelve
// envuelto con el ID del sprint y no es un sprint inexistente.
func TestCerrarSprint_FallosDeInfraestructura(t *testing.T) {
	falla := errors.New("se corto la conexion")
	casos := map[string]func(*repositorioCierreFalso){
		"al leer el sprint":     func(r *repositorioCierreFalso) { r.errObtener = falla },
		"al leer las historias": func(r *repositorioCierreFalso) { r.errHistorias = falla },
		"al guardar el cierre":  func(r *repositorioCierreFalso) { r.errGuardar = falla },
	}
	for nombre, romper := range casos {
		t.Run(nombre, func(t *testing.T) {
			repo := sprintParaCerrar()
			romper(repo)

			_, err := app.NuevoCerrarSprint(repo).Ejecutar(context.Background(), "sprint-1")

			if !errors.Is(err, falla) {
				t.Fatalf("error = %v, se esperaba el del repositorio", err)
			}
			if !strings.Contains(err.Error(), "sprint-1") {
				t.Errorf("el error %q no indica el sprint", err)
			}
			if errors.Is(err, app.ErrSprintInexistente) {
				t.Errorf("un fallo del repositorio no es un sprint inexistente: %v", err)
			}
		})
	}
}
