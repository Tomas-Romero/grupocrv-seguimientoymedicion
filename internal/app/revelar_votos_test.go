package app_test

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// repositorioRevelarFalso simula la persistencia de revelar sin base de datos. Al guardar
// devuelve la revelacion que tiene armada, como el repositorio real devuelve la suya.
type repositorioRevelarFalso struct {
	sesion           planningpoker.SesionPlanningPoker
	existeSesion     bool
	existeIntegrante bool
	votos            map[string]int
	revelacion       planningpoker.Revelacion

	errSesion     error
	errIntegrante error
	errVotos      error
	errGuardar    error

	integrantesConsultados []string
	guardados              []string
}

func (r *repositorioRevelarFalso) ObtenerSesion(_ context.Context, _ string) (planningpoker.SesionPlanningPoker, bool, error) {
	return r.sesion, r.existeSesion, r.errSesion
}

func (r *repositorioRevelarFalso) ExisteIntegranteEnProyecto(_ context.Context, proyectoID, integranteID string) (bool, error) {
	r.integrantesConsultados = append(r.integrantesConsultados, proyectoID+"/"+integranteID)
	return r.existeIntegrante, r.errIntegrante
}

func (r *repositorioRevelarFalso) VotosDeLaRonda(_ context.Context, _ string) (map[string]int, error) {
	return r.votos, r.errVotos
}

func (r *repositorioRevelarFalso) GuardarRevelado(_ context.Context, sesionID, integranteID string) (planningpoker.Revelacion, error) {
	r.guardados = append(r.guardados, sesionID+"/"+integranteID)
	if r.errGuardar != nil {
		return planningpoker.Revelacion{}, r.errGuardar
	}
	return r.revelacion, nil
}

// sesionConVotos es un repositorio con una sesion votando del proyecto 1, de la que ana y
// beto participan, y en la que ana ya voto 5.
func sesionConVotos() *repositorioRevelarFalso {
	return &repositorioRevelarFalso{
		sesion: planningpoker.SesionPlanningPoker{
			ID: "sesion-1", ItemID: "item-1", ProyectoID: "proyecto-1",
			Estado: planningpoker.EstadoVotando, Ronda: 1, Participantes: []string{"ana", "beto"},
		},
		existeSesion:     true,
		existeIntegrante: true,
		votos:            map[string]int{"ana": 5},
		revelacion: planningpoker.Revelacion{
			SesionID: "sesion-1", Ronda: 1, Votos: []planningpoker.Voto{{ParticipanteID: "ana", Valor: 5}},
			Faltaron: []string{"beto"}, Minimo: 5, Maximo: 5,
		},
	}
}

func revelar(repo *repositorioRevelarFalso, integranteID string) (planningpoker.Revelacion, error) {
	return app.NuevoRevelarVotos(repo).Ejecutar(context.Background(), "sesion-1", integranteID)
}

// US-016 / CA-016-1, CA-016-3: el caso de uso verifica el integrante en el proyecto de la
// sesion, valida con el dominio, guarda el revelado una sola vez y devuelve la revelacion.
func TestRevelarVotos_Revela(t *testing.T) {
	repo := sesionConVotos()

	got, err := revelar(repo, "beto")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if !reflect.DeepEqual(got, repo.revelacion) {
		t.Errorf("revelacion = %+v, se esperaba la que devolvio el repositorio", got)
	}
	if !slices.Equal(repo.integrantesConsultados, []string{"proyecto-1/beto"}) {
		t.Errorf("se consulto %q, se esperaba el integrante en el proyecto de la sesion", repo.integrantesConsultados)
	}
	if !slices.Equal(repo.guardados, []string{"sesion-1/beto"}) {
		t.Errorf("se guardo %q, se esperaba un solo revelado de la sesion 1", repo.guardados)
	}
}

// US-016 / CA-016-8, RN-016-1, CL-016-13: si la sesion o el integrante no existen (o el
// integrante es de otro proyecto), se devuelve ese error sin validar nada mas y no se
// guarda nada. Si no existe la sesion, no se mira el integrante.
func TestRevelarVotos_LoQueNoExiste(t *testing.T) {
	casos := []struct {
		nombre      string
		romper      func(*repositorioRevelarFalso)
		errEsperado error
	}{
		{"sesion inexistente", func(r *repositorioRevelarFalso) {
			r.existeSesion = false
			r.existeIntegrante = false
		}, app.ErrSesionInexistente},
		{"integrante inexistente o de otro proyecto", func(r *repositorioRevelarFalso) { r.existeIntegrante = false }, app.ErrIntegranteInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := sesionConVotos()
			c.romper(repo)
			// Ademas nadie voto: si se validara, el error seria otro.
			repo.votos = nil

			_, err := revelar(repo, "dani")

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if errors.Is(err, planningpoker.ErrSinVotos) {
				t.Errorf("el error %v trae validaciones del dominio: la inexistencia va sola", err)
			}
			if len(repo.guardados) != 0 {
				t.Errorf("se guardo %q, no se esperaba nada", repo.guardados)
			}
		})
	}
	repo := sesionConVotos()
	repo.existeSesion, repo.existeIntegrante = false, false
	if _, err := revelar(repo, "dani"); err != nil && len(repo.integrantesConsultados) != 0 {
		t.Errorf("con la sesion inexistente se consulto el integrante: %q", repo.integrantesConsultados)
	}
}

// US-016 / CA-016-5, CA-016-6, CA-016-7: el dominio rechaza y no se guarda nada. Los errores
// del dominio se devuelven tal cual, con su mensaje.
func TestRevelarVotos_ReglasDelDominio(t *testing.T) {
	casos := []struct {
		nombre string
		romper func(*repositorioRevelarFalso)
		quien  string
		quiero error
	}{
		{"nadie voto", func(r *repositorioRevelarFalso) { r.votos = nil }, "ana", planningpoker.ErrSinVotos},
		{"quien revela no participa", func(*repositorioRevelarFalso) {}, "dani", planningpoker.ErrNoEsParticipante},
		{"la sesion ya esta revelada", func(r *repositorioRevelarFalso) { r.sesion.Estado = planningpoker.EstadoRevelada }, "ana", planningpoker.ErrVotacionCerrada},
		{"la sesion esta finalizada", func(r *repositorioRevelarFalso) { r.sesion.Estado = planningpoker.EstadoFinalizada }, "ana", planningpoker.ErrVotacionCerrada},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := sesionConVotos()
			c.romper(repo)

			got, err := revelar(repo, c.quien)

			if !errors.Is(err, c.quiero) {
				t.Fatalf("error = %v, se esperaba %v", err, c.quiero)
			}
			if !reflect.DeepEqual(got, planningpoker.Revelacion{}) {
				t.Errorf("con error la revelacion tiene que estar vacia: %+v", got)
			}
			if len(repo.guardados) != 0 {
				t.Errorf("se guardo %q, no se esperaba nada", repo.guardados)
			}
		})
	}
}

// US-016 / RN-016-8, CL-016-14: si el repositorio, con la sesion bloqueada, ve que las reglas ya
// no se cumplen (se revelo entre la lectura y la transaccion), el error del dominio llega
// tal cual.
func TestRevelarVotos_RechazosDelRepositorio(t *testing.T) {
	repo := sesionConVotos()
	repo.errGuardar = planningpoker.ErrVotacionCerrada

	_, err := revelar(repo, "ana")

	if !errors.Is(err, planningpoker.ErrVotacionCerrada) {
		t.Errorf("error = %v, se esperaba ErrVotacionCerrada", err)
	}
}

// Los fallos de infraestructura llegan envueltos con contexto (RN-016-11).
func TestRevelarVotos_FallosDeInfraestructura(t *testing.T) {
	errBase := errors.New("base caida")
	casos := []struct {
		nombre string
		romper func(*repositorioRevelarFalso)
	}{
		{"al leer la sesion", func(r *repositorioRevelarFalso) { r.errSesion = errBase }},
		{"al leer el integrante", func(r *repositorioRevelarFalso) { r.errIntegrante = errBase }},
		{"al leer los votos", func(r *repositorioRevelarFalso) { r.errVotos = errBase }},
		{"al guardar", func(r *repositorioRevelarFalso) { r.errGuardar = errBase }},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := sesionConVotos()
			c.romper(repo)

			_, err := revelar(repo, "ana")

			if !errors.Is(err, errBase) {
				t.Errorf("error = %v, se esperaba que envolviera %v", err, errBase)
			}
		})
	}
}
