package app_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// repositorioPlanningPokerFalso simula la persistencia de las sesiones sin base de
// datos. Al registrar devuelve la sesion con un ID, como la base.
type repositorioPlanningPokerFalso struct {
	existeProyecto bool
	item           backlog.ItemBacklog
	existeItem     bool
	sesionAbierta  bool
	integrantes    []string

	errProyecto      error
	errItem          error
	errSesionAbierta error
	errIntegrantes   error
	errRegistrar     error

	registradas []planningpoker.SesionPlanningPoker
}

func (r *repositorioPlanningPokerFalso) ExisteProyecto(_ context.Context, _ string) (bool, error) {
	return r.existeProyecto, r.errProyecto
}

func (r *repositorioPlanningPokerFalso) ObtenerItem(_ context.Context, _ string) (backlog.ItemBacklog, bool, error) {
	return r.item, r.existeItem, r.errItem
}

func (r *repositorioPlanningPokerFalso) TieneSesionAbierta(_ context.Context, _ string) (bool, error) {
	return r.sesionAbierta, r.errSesionAbierta
}

func (r *repositorioPlanningPokerFalso) IDsDeIntegrantes(_ context.Context, _ string) ([]string, error) {
	return r.integrantes, r.errIntegrantes
}

func (r *repositorioPlanningPokerFalso) RegistrarSesion(_ context.Context, _ string, sesion planningpoker.SesionPlanningPoker) (planningpoker.SesionPlanningPoker, error) {
	r.registradas = append(r.registradas, sesion)
	if r.errRegistrar != nil {
		return planningpoker.SesionPlanningPoker{}, r.errRegistrar
	}
	sesion.ID = "sesion-1"
	return sesion, nil
}

// listoParaIniciar es un repositorio con el proyecto, una historia pendiente del
// proyecto y tres integrantes.
func listoParaIniciar(t *testing.T) *repositorioPlanningPokerFalso {
	t.Helper()
	item := itemDelRepositorio(t, backlog.EstadoPendiente)
	return &repositorioPlanningPokerFalso{
		existeProyecto: true,
		item:           item,
		existeItem:     true,
		integrantes:    []string{"ana", "beto", "caro"},
	}
}

func iniciar(repo *repositorioPlanningPokerFalso, proyectoID, itemID string) (planningpoker.SesionPlanningPoker, error) {
	return app.NuevoIniciarSesionPlanningPoker(repo).Ejecutar(context.Background(), proyectoID, itemID)
}

// US-014 / CA-014-1, RN-014-4, RN-014-6: con todo en regla, el caso de uso registra la
// sesion votando, en la ronda 1, con los integrantes del proyecto, y la devuelve con su
// ID.
func TestIniciarSesionPlanningPoker_Inicia(t *testing.T) {
	repo := listoParaIniciar(t)

	sesion, err := iniciar(repo, "proyecto-1", "item-1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if sesion.ID != "sesion-1" || sesion.ItemID != "item-1" || sesion.Estado != planningpoker.EstadoVotando || sesion.Ronda != 1 {
		t.Errorf("sesion = %+v, se esperaba la sesion 1 del item 1, votando y en la ronda 1", sesion)
	}
	if !slices.Equal(sesion.Participantes, []string{"ana", "beto", "caro"}) {
		t.Errorf("participantes = %q, se esperaban los tres integrantes", sesion.Participantes)
	}
	if len(repo.registradas) != 1 {
		t.Errorf("se registraron %d sesiones, se esperaba 1", len(repo.registradas))
	}
}

// US-014 / CA-014-2, RN-014-3: una historia ya estimada o en progreso se pone en sesion.
func TestIniciarSesionPlanningPoker_HistoriaEstimadaOEnProgreso(t *testing.T) {
	repo := listoParaIniciar(t)
	item, err := repo.item.Estimar(5)
	if err != nil {
		t.Fatalf("Estimar: %v", err)
	}
	item.Estado = backlog.EstadoEnProgreso
	repo.item = item

	if _, err := iniciar(repo, "proyecto-1", "item-1"); err != nil {
		t.Errorf("Ejecutar: %v", err)
	}
}

// US-014 / CA-014-7, RN-014-1, CL-014-10: si el proyecto o la historia no existen, o la
// historia es de otro proyecto, se devuelve ese error sin validar nada mas y no se
// registra nada. Si no existe el proyecto, no se mira la historia.
func TestIniciarSesionPlanningPoker_LoQueNoExiste(t *testing.T) {
	casos := []struct {
		nombre      string
		romper      func(*repositorioPlanningPokerFalso)
		errEsperado error
		id          string
	}{
		{"proyecto inexistente", func(r *repositorioPlanningPokerFalso) {
			r.existeProyecto = false
			r.existeItem = false
		}, app.ErrProyectoInexistente, "proyecto-1"},
		{"historia inexistente", func(r *repositorioPlanningPokerFalso) { r.existeItem = false }, app.ErrItemInexistente, "item-1"},
		{"historia de otro proyecto", func(r *repositorioPlanningPokerFalso) { r.item.ProyectoID = "otro" }, app.ErrItemInexistente, "item-1"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := listoParaIniciar(t)
			c.romper(repo)
			// Ademas la historia esta completada y hay un solo integrante: si se validara,
			// el error seria otro.
			repo.item.Estado = backlog.EstadoCompletado
			repo.integrantes = []string{"ana"}

			_, err := iniciar(repo, "proyecto-1", "item-1")

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if errors.Is(err, backlog.ErrItemCompletado) || errors.Is(err, planningpoker.ErrParticipantesInsuficientes) {
				t.Errorf("el error %q se unio con los de validacion", err)
			}
			if !strings.Contains(err.Error(), c.id) {
				t.Errorf("el error %q no indica el ID %s", err, c.id)
			}
			if len(repo.registradas) != 0 {
				t.Errorf("se registraron %d sesiones, se esperaba ninguna", len(repo.registradas))
			}
		})
	}
}

// US-014 / CA-014-4, CA-014-5, CA-014-6, RN-014-2, RN-014-5, RN-014-7: lo que rechaza el
// dominio se devuelve tal cual y no se registra nada.
func TestIniciarSesionPlanningPoker_ReglasDelDominio(t *testing.T) {
	casos := []struct {
		nombre      string
		romper      func(*repositorioPlanningPokerFalso)
		errEsperado error
	}{
		{"historia completada", func(r *repositorioPlanningPokerFalso) { r.item.Estado = backlog.EstadoCompletado }, backlog.ErrItemCompletado},
		{"sesion abierta", func(r *repositorioPlanningPokerFalso) { r.sesionAbierta = true }, planningpoker.ErrSesionAbierta},
		{"un integrante", func(r *repositorioPlanningPokerFalso) { r.integrantes = []string{"ana"} }, planningpoker.ErrParticipantesInsuficientes},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := listoParaIniciar(t)
			c.romper(repo)

			sesion, err := iniciar(repo, "proyecto-1", "item-1")

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if sesion.ItemID != "" || len(repo.registradas) != 0 {
				t.Errorf("devolvio %+v y registro %d sesiones, se esperaba nada", sesion, len(repo.registradas))
			}
		})
	}
}

// US-014 / RN-014-11, CL-014-8, CL-014-11: lo que el repositorio rechaza dentro de la
// transaccion (otra sesion se abrio, la historia se completo o desaparecio) se devuelve
// reconocible y sin sesion.
func TestIniciarSesionPlanningPoker_RechazosDelRepositorio(t *testing.T) {
	for _, errRegistrar := range []error{planningpoker.ErrSesionAbierta, backlog.ErrItemCompletado, app.ErrItemInexistente} {
		t.Run(errRegistrar.Error(), func(t *testing.T) {
			repo := listoParaIniciar(t)
			repo.errRegistrar = errRegistrar

			sesion, err := iniciar(repo, "proyecto-1", "item-1")

			if !errors.Is(err, errRegistrar) {
				t.Errorf("error = %v, se esperaba %v", err, errRegistrar)
			}
			if sesion.ID != "" {
				t.Errorf("devolvio la sesion %+v, se esperaba el valor cero", sesion)
			}
		})
	}
}

// Un fallo del repositorio en cualquier lectura o al registrar se devuelve envuelto y no
// es un inexistente.
func TestIniciarSesionPlanningPoker_FallosDeInfraestructura(t *testing.T) {
	falla := errors.New("se corto la conexion")
	casos := map[string]func(*repositorioPlanningPokerFalso){
		"al leer el proyecto":     func(r *repositorioPlanningPokerFalso) { r.errProyecto = falla },
		"al leer la historia":     func(r *repositorioPlanningPokerFalso) { r.errItem = falla },
		"al consultar sesiones":   func(r *repositorioPlanningPokerFalso) { r.errSesionAbierta = falla },
		"al leer los integrantes": func(r *repositorioPlanningPokerFalso) { r.errIntegrantes = falla },
		"al registrar la sesion":  func(r *repositorioPlanningPokerFalso) { r.errRegistrar = falla },
	}
	for nombre, romper := range casos {
		t.Run(nombre, func(t *testing.T) {
			repo := listoParaIniciar(t)
			romper(repo)

			_, err := iniciar(repo, "proyecto-1", "item-1")

			if !errors.Is(err, falla) {
				t.Fatalf("error = %v, se esperaba el del repositorio", err)
			}
			if errors.Is(err, app.ErrProyectoInexistente) || errors.Is(err, app.ErrItemInexistente) {
				t.Errorf("un fallo del repositorio no es un inexistente: %v", err)
			}
		})
	}
}
