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

// repositorioNuevaRondaFalso simula la persistencia de abrir una nueva ronda sin base de datos.
// Al guardar devuelve la sesion que tiene armada, como el repositorio real devuelve la suya.
type repositorioNuevaRondaFalso struct {
	sesion           planningpoker.SesionPlanningPoker
	existeSesion     bool
	existeIntegrante bool
	votos            map[string]int
	nueva            planningpoker.SesionPlanningPoker

	errSesion     error
	errIntegrante error
	errVotos      error
	errGuardar    error

	integrantesConsultados []string
	guardados              []string
}

func (r *repositorioNuevaRondaFalso) ObtenerSesion(_ context.Context, _ string) (planningpoker.SesionPlanningPoker, bool, error) {
	return r.sesion, r.existeSesion, r.errSesion
}

func (r *repositorioNuevaRondaFalso) ExisteIntegranteEnProyecto(_ context.Context, proyectoID, integranteID string) (bool, error) {
	r.integrantesConsultados = append(r.integrantesConsultados, proyectoID+"/"+integranteID)
	return r.existeIntegrante, r.errIntegrante
}

func (r *repositorioNuevaRondaFalso) VotacionDeLaRonda(_ context.Context, _ string) (planningpoker.Votacion, error) {
	if r.errVotos != nil {
		return planningpoker.Votacion{}, r.errVotos
	}
	return planningpoker.NuevaVotacion(r.sesion, r.votos), nil
}

func (r *repositorioNuevaRondaFalso) GuardarNuevaRonda(_ context.Context, sesionID, integranteID string) (planningpoker.SesionPlanningPoker, error) {
	r.guardados = append(r.guardados, sesionID+"/"+integranteID)
	if r.errGuardar != nil {
		return planningpoker.SesionPlanningPoker{}, r.errGuardar
	}
	return r.nueva, nil
}

// sesionParaNuevaRonda es un repositorio con una sesion revelada del proyecto 1, de la que ana y beto
// participan, con votos de 3 y 8.
func sesionParaNuevaRonda() *repositorioNuevaRondaFalso {
	sesion := planningpoker.SesionPlanningPoker{
		ID: "sesion-1", ItemID: "item-1", ProyectoID: "proyecto-1",
		Estado: planningpoker.EstadoRevelada, Ronda: 1, Participantes: []string{"ana", "beto"},
	}
	nueva := sesion
	nueva.Estado, nueva.Ronda = planningpoker.EstadoVotando, 2
	return &repositorioNuevaRondaFalso{
		sesion: sesion, existeSesion: true, existeIntegrante: true,
		votos: map[string]int{"ana": 3, "beto": 8}, nueva: nueva,
	}
}

func abrirRonda(repo *repositorioNuevaRondaFalso, integranteID string) (planningpoker.SesionPlanningPoker, error) {
	return app.NuevoAbrirNuevaRonda(repo).Ejecutar(context.Background(), "sesion-1", integranteID)
}

// US-017 / CA-017-1: el caso de uso verifica el integrante en el proyecto de la sesion, valida
// con el dominio, guarda una sola vez y devuelve la sesion en la ronda nueva.
func TestAbrirNuevaRonda_Abre(t *testing.T) {
	repo := sesionParaNuevaRonda()

	got, err := abrirRonda(repo, "beto")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if !reflect.DeepEqual(got, repo.nueva) {
		t.Errorf("sesion = %+v, se esperaba la que devolvio el repositorio", got)
	}
	if !slices.Equal(repo.integrantesConsultados, []string{"proyecto-1/beto"}) {
		t.Errorf("se consulto %q, se esperaba el integrante en el proyecto de la sesion", repo.integrantesConsultados)
	}
	if !slices.Equal(repo.guardados, []string{"sesion-1/beto"}) {
		t.Errorf("se guardo %q, se esperaba una sola ronda nueva de la sesion 1", repo.guardados)
	}
}

// US-017 / CA-017-4, RN-017-1, CL-017-12: si la sesion o el integrante no existen (o el
// integrante es de otro proyecto), se devuelve ese error sin validar nada mas y no se guarda.
func TestAbrirNuevaRonda_LoQueNoExiste(t *testing.T) {
	casos := []struct {
		nombre      string
		romper      func(*repositorioNuevaRondaFalso)
		errEsperado error
	}{
		{"sesion inexistente", func(r *repositorioNuevaRondaFalso) {
			r.existeSesion = false
			r.existeIntegrante = false
		}, app.ErrSesionInexistente},
		{"integrante inexistente o de otro proyecto", func(r *repositorioNuevaRondaFalso) { r.existeIntegrante = false }, app.ErrIntegranteInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := sesionParaNuevaRonda()
			c.romper(repo)
			// Ademas la sesion esta votando: si se validara, el error seria otro.
			repo.sesion.Estado = planningpoker.EstadoVotando

			_, err := abrirRonda(repo, "dani")

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if errors.Is(err, planningpoker.ErrNuevaRondaSinRevelar) || errors.Is(err, planningpoker.ErrNoEsParticipante) {
				t.Errorf("el error %v trae validaciones del dominio: la inexistencia va sola", err)
			}
			if len(repo.guardados) != 0 {
				t.Errorf("se guardo %q, no se esperaba nada", repo.guardados)
			}
		})
	}
	repo := sesionParaNuevaRonda()
	repo.existeSesion, repo.existeIntegrante = false, false
	if _, err := abrirRonda(repo, "dani"); err != nil && len(repo.integrantesConsultados) != 0 {
		t.Errorf("con la sesion inexistente se consulto el integrante: %q", repo.integrantesConsultados)
	}
}

// US-017 / CA-017-3, CA-017-4: el dominio rechaza y no se guarda nada. Los errores del dominio
// se devuelven tal cual, con su mensaje.
func TestAbrirNuevaRonda_ReglasDelDominio(t *testing.T) {
	casos := []struct {
		nombre string
		romper func(*repositorioNuevaRondaFalso)
		quien  string
		quiero error
	}{
		{"sin divergencia", func(r *repositorioNuevaRondaFalso) { r.votos = map[string]int{"ana": 5, "beto": 5} }, "ana", planningpoker.ErrSinDivergencia},
		{"quien abre no participa", func(*repositorioNuevaRondaFalso) {}, "dani", planningpoker.ErrNoEsParticipante},
		{"la sesion esta votando", func(r *repositorioNuevaRondaFalso) { r.sesion.Estado = planningpoker.EstadoVotando }, "ana", planningpoker.ErrNuevaRondaSinRevelar},
		{"la sesion esta finalizada", func(r *repositorioNuevaRondaFalso) { r.sesion.Estado = planningpoker.EstadoFinalizada }, "ana", planningpoker.ErrNuevaRondaSinRevelar},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := sesionParaNuevaRonda()
			c.romper(repo)

			got, err := abrirRonda(repo, c.quien)

			if !errors.Is(err, c.quiero) {
				t.Fatalf("error = %v, se esperaba %v", err, c.quiero)
			}
			if !reflect.DeepEqual(got, planningpoker.SesionPlanningPoker{}) {
				t.Errorf("con error la sesion tiene que estar vacia: %+v", got)
			}
			if len(repo.guardados) != 0 {
				t.Errorf("se guardo %q, no se esperaba nada", repo.guardados)
			}
		})
	}
}

// US-017 / RN-017-9: si el repositorio, con la sesion bloqueada, ve que las reglas ya no se
// cumplen (otro abrio la ronda en el medio), el error del dominio llega tal cual.
func TestAbrirNuevaRonda_RechazosDelRepositorio(t *testing.T) {
	repo := sesionParaNuevaRonda()
	repo.errGuardar = planningpoker.ErrNuevaRondaSinRevelar

	_, err := abrirRonda(repo, "ana")

	if !errors.Is(err, planningpoker.ErrNuevaRondaSinRevelar) {
		t.Errorf("error = %v, se esperaba ErrNuevaRondaSinRevelar", err)
	}
}

// Los fallos de infraestructura llegan envueltos con contexto.
func TestAbrirNuevaRonda_FallosDeInfraestructura(t *testing.T) {
	errBase := errors.New("base caida")
	casos := []struct {
		nombre string
		romper func(*repositorioNuevaRondaFalso)
	}{
		{"al leer la sesion", func(r *repositorioNuevaRondaFalso) { r.errSesion = errBase }},
		{"al leer el integrante", func(r *repositorioNuevaRondaFalso) { r.errIntegrante = errBase }},
		{"al leer los votos", func(r *repositorioNuevaRondaFalso) { r.errVotos = errBase }},
		{"al guardar", func(r *repositorioNuevaRondaFalso) { r.errGuardar = errBase }},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := sesionParaNuevaRonda()
			c.romper(repo)

			_, err := abrirRonda(repo, "ana")

			if !errors.Is(err, errBase) {
				t.Errorf("error = %v, se esperaba que envolviera %v", err, errBase)
			}
		})
	}
}
