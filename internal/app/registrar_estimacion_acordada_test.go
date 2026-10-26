package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// acuerdoGuardado es una llamada a GuardarAcuerdo que anota el repositorio falso.
type acuerdoGuardado struct {
	sesionID, integranteID string
	valor                  int
}

// repositorioAcuerdoFalso simula la persistencia del acuerdo sin base de datos. Al guardar
// devuelve la sesion finalizada que arma el dominio, como el repositorio real.
type repositorioAcuerdoFalso struct {
	sesion           planningpoker.SesionPlanningPoker
	existeSesion     bool
	existeIntegrante bool
	item             backlog.ItemBacklog
	existeItem       bool

	errSesion     error
	errIntegrante error
	errItem       error
	errGuardar    error

	itemsConsultados []string
	guardados        []acuerdoGuardado
}

func (r *repositorioAcuerdoFalso) ObtenerSesion(_ context.Context, _ string) (planningpoker.SesionPlanningPoker, bool, error) {
	return r.sesion, r.existeSesion, r.errSesion
}

func (r *repositorioAcuerdoFalso) ExisteIntegranteEnProyecto(_ context.Context, _, _ string) (bool, error) {
	return r.existeIntegrante, r.errIntegrante
}

func (r *repositorioAcuerdoFalso) ObtenerItem(_ context.Context, itemID string) (backlog.ItemBacklog, bool, error) {
	r.itemsConsultados = append(r.itemsConsultados, itemID)
	return r.item, r.existeItem, r.errItem
}

func (r *repositorioAcuerdoFalso) GuardarAcuerdo(_ context.Context, sesionID, integranteID string, valor int) (planningpoker.SesionPlanningPoker, error) {
	r.guardados = append(r.guardados, acuerdoGuardado{sesionID: sesionID, integranteID: integranteID, valor: valor})
	if r.errGuardar != nil {
		return planningpoker.SesionPlanningPoker{}, r.errGuardar
	}
	return planningpoker.Acordar(r.sesion, r.item.Estado == backlog.EstadoCompletado, integranteID, valor)
}

// sesionRevelada es un repositorio con una sesion revelada del proyecto 1, sobre el item 1
// pendiente, de la que ana participa.
func sesionRevelada(t *testing.T) *repositorioAcuerdoFalso {
	t.Helper()
	return &repositorioAcuerdoFalso{
		sesion: planningpoker.SesionPlanningPoker{
			ID: "sesion-1", ItemID: "item-1", ProyectoID: "proyecto-1",
			Estado: planningpoker.EstadoRevelada, Ronda: 1, Participantes: []string{"ana", "beto"},
		},
		existeSesion:     true,
		existeIntegrante: true,
		item:             itemDelRepositorio(t, backlog.EstadoPendiente),
		existeItem:       true,
	}
}

func acordar(repo *repositorioAcuerdoFalso, integranteID string, valor int) (planningpoker.SesionPlanningPoker, error) {
	return app.NuevoRegistrarEstimacionAcordada(repo).Ejecutar(context.Background(), "sesion-1", integranteID, valor)
}

// US-018 / CA-018-1, RN-018-8: el caso de uso verifica la sesion, el integrante y la
// historia de la sesion, valida con el dominio y guarda el acuerdo una sola vez. Devuelve
// la sesion finalizada con el valor acordado.
func TestRegistrarEstimacionAcordada_Registra(t *testing.T) {
	repo := sesionRevelada(t)

	sesion, err := acordar(repo, "ana", 5)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if puntos, ok := sesion.ValorAcordado.Valor(); sesion.Estado != planningpoker.EstadoFinalizada || !ok || puntos != 5 {
		t.Errorf("sesion = (%q, %d, %t), se esperaba (finalizada, 5, true)", sesion.Estado, puntos, ok)
	}
	if len(repo.itemsConsultados) != 1 || repo.itemsConsultados[0] != "item-1" {
		t.Errorf("se consulto %q, se esperaba la historia de la sesion", repo.itemsConsultados)
	}
	if len(repo.guardados) != 1 || repo.guardados[0] != (acuerdoGuardado{sesionID: "sesion-1", integranteID: "ana", valor: 5}) {
		t.Errorf("se guardo %+v, se esperaba el acuerdo de ana en la sesion 1", repo.guardados)
	}
}

// US-018 / CA-018-8, RN-018-1, CL-018-13: si la sesion o el integrante no existen (o el
// integrante es de otro proyecto), se devuelve ese error sin validar nada mas y no se
// guarda nada. Si no existe la sesion, no se mira el resto.
func TestRegistrarEstimacionAcordada_LoQueNoExiste(t *testing.T) {
	casos := []struct {
		nombre      string
		romper      func(*repositorioAcuerdoFalso)
		errEsperado error
		id          string
	}{
		{"sesion inexistente", func(r *repositorioAcuerdoFalso) {
			r.existeSesion = false
			r.existeIntegrante = false
		}, app.ErrSesionInexistente, "sesion-1"},
		{"integrante inexistente o de otro proyecto", func(r *repositorioAcuerdoFalso) { r.existeIntegrante = false }, app.ErrIntegranteInexistente, "dani"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := sesionRevelada(t)
			c.romper(repo)
			// Ademas la sesion esta votando y el valor esta fuera de la escala: si se
			// validara, el error seria otro.
			repo.sesion.Estado = planningpoker.EstadoVotando

			_, err := acordar(repo, "dani", 4)

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if errors.Is(err, planningpoker.ErrSesionNoRevelada) || errors.Is(err, backlog.ErrEstimacionFueraDeEscala) {
				t.Errorf("el error %q se unio con los de validacion", err)
			}
			if !strings.Contains(err.Error(), c.id) {
				t.Errorf("el error %q no indica el ID %s", err, c.id)
			}
			if len(repo.guardados) != 0 {
				t.Errorf("se guardaron %d acuerdos, se esperaba ninguno", len(repo.guardados))
			}
		})
	}
}

// US-018 / CA-018-4 a CA-018-7, RN-018-2 a RN-018-6: lo que rechaza el dominio se devuelve
// tal cual y no se guarda nada. Con la historia completada, el unico error es
// ErrItemCompletado.
func TestRegistrarEstimacionAcordada_ReglasDelDominio(t *testing.T) {
	casos := []struct {
		nombre      string
		romper      func(*repositorioAcuerdoFalso)
		integrante  string
		valor       int
		errEsperado error
	}{
		{"sesion votando", func(r *repositorioAcuerdoFalso) { r.sesion.Estado = planningpoker.EstadoVotando }, "ana", 5, planningpoker.ErrSesionNoRevelada},
		{"sesion finalizada", func(r *repositorioAcuerdoFalso) { r.sesion.Estado = planningpoker.EstadoFinalizada }, "ana", 5, planningpoker.ErrSesionNoRevelada},
		{"no participa", func(*repositorioAcuerdoFalso) {}, "dani", 5, planningpoker.ErrNoEsParticipante},
		{"fuera de la escala", func(*repositorioAcuerdoFalso) {}, "ana", 4, backlog.ErrEstimacionFueraDeEscala},
		{"historia completada", func(r *repositorioAcuerdoFalso) { r.item.Estado = backlog.EstadoCompletado }, "dani", 4, backlog.ErrItemCompletado},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := sesionRevelada(t)
			c.romper(repo)

			_, err := acordar(repo, c.integrante, c.valor)

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if len(repo.guardados) != 0 {
				t.Errorf("se guardaron %d acuerdos, se esperaba ninguno", len(repo.guardados))
			}
		})
	}
}

// US-018 / RN-018-9, RN-018-10, CL-018-10, CL-018-11: lo que el repositorio rechaza dentro
// de la transaccion (otro registro llego antes, la historia se completo en el medio o la
// sesion desaparecio) se devuelve reconocible y sin sesion.
func TestRegistrarEstimacionAcordada_RechazosDelRepositorio(t *testing.T) {
	for _, errGuardar := range []error{planningpoker.ErrSesionNoRevelada, backlog.ErrItemCompletado, app.ErrSesionInexistente} {
		t.Run(errGuardar.Error(), func(t *testing.T) {
			repo := sesionRevelada(t)
			repo.errGuardar = errGuardar

			sesion, err := acordar(repo, "ana", 5)

			if !errors.Is(err, errGuardar) {
				t.Errorf("error = %v, se esperaba %v", err, errGuardar)
			}
			if sesion.ID != "" {
				t.Errorf("devolvio la sesion %+v, se esperaba el valor cero", sesion)
			}
		})
	}
}

// Una historia de la sesion que no esta (no deberia pasar: la sesion se borra con su
// historia) se informa como item inexistente, con su ID.
func TestRegistrarEstimacionAcordada_HistoriaQueNoEsta(t *testing.T) {
	repo := sesionRevelada(t)
	repo.existeItem = false

	_, err := acordar(repo, "ana", 5)

	if !errors.Is(err, app.ErrItemInexistente) || !strings.Contains(err.Error(), "item-1") {
		t.Errorf("error = %v, se esperaba ErrItemInexistente con el ID de la historia", err)
	}
	if len(repo.guardados) != 0 {
		t.Errorf("se guardaron %d acuerdos, se esperaba ninguno", len(repo.guardados))
	}
}

// Un fallo del repositorio en cualquier lectura o al guardar se devuelve envuelto y no es
// un inexistente.
func TestRegistrarEstimacionAcordada_FallosDeInfraestructura(t *testing.T) {
	falla := errors.New("se corto la conexion")
	casos := map[string]func(*repositorioAcuerdoFalso){
		"al leer la sesion":          func(r *repositorioAcuerdoFalso) { r.errSesion = falla },
		"al consultar el integrante": func(r *repositorioAcuerdoFalso) { r.errIntegrante = falla },
		"al leer la historia":        func(r *repositorioAcuerdoFalso) { r.errItem = falla },
		"al guardar el acuerdo":      func(r *repositorioAcuerdoFalso) { r.errGuardar = falla },
	}
	for nombre, romper := range casos {
		t.Run(nombre, func(t *testing.T) {
			repo := sesionRevelada(t)
			romper(repo)

			_, err := acordar(repo, "ana", 5)

			if !errors.Is(err, falla) {
				t.Fatalf("error = %v, se esperaba el del repositorio", err)
			}
			if errors.Is(err, app.ErrSesionInexistente) || errors.Is(err, app.ErrIntegranteInexistente) || errors.Is(err, app.ErrItemInexistente) {
				t.Errorf("un fallo del repositorio no es un inexistente: %v", err)
			}
		})
	}
}
