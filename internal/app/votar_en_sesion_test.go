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

// votoGuardado es una llamada a RegistrarVoto que anota el repositorio falso.
type votoGuardado struct {
	sesionID, integranteID string
	valor                  int
}

// repositorioVotosFalso simula la persistencia de los votos sin base de datos. Al
// registrar devuelve la votacion con el voto nuevo, como el repositorio real.
type repositorioVotosFalso struct {
	sesion           planningpoker.SesionPlanningPoker
	existeSesion     bool
	existeIntegrante bool

	errSesion     error
	errIntegrante error
	errRegistrar  error

	integrantesConsultados []string
	guardados              []votoGuardado
}

func (r *repositorioVotosFalso) ObtenerSesion(_ context.Context, _ string) (planningpoker.SesionPlanningPoker, bool, error) {
	return r.sesion, r.existeSesion, r.errSesion
}

func (r *repositorioVotosFalso) ExisteIntegranteEnProyecto(_ context.Context, proyectoID, integranteID string) (bool, error) {
	r.integrantesConsultados = append(r.integrantesConsultados, proyectoID+"/"+integranteID)
	return r.existeIntegrante, r.errIntegrante
}

func (r *repositorioVotosFalso) RegistrarVoto(_ context.Context, sesionID, integranteID string, valor int) (planningpoker.Votacion, error) {
	r.guardados = append(r.guardados, votoGuardado{sesionID: sesionID, integranteID: integranteID, valor: valor})
	if r.errRegistrar != nil {
		return planningpoker.Votacion{}, r.errRegistrar
	}
	return planningpoker.NuevaVotacion(r.sesion, map[string]int{integranteID: valor}), nil
}

// sesionVotando es un repositorio con una sesion votando del proyecto 1, de la que ana
// participa.
func sesionVotando() *repositorioVotosFalso {
	return &repositorioVotosFalso{
		sesion: planningpoker.SesionPlanningPoker{
			ID: "sesion-1", ItemID: "item-1", ProyectoID: "proyecto-1",
			Estado: planningpoker.EstadoVotando, Ronda: 1, Participantes: []string{"ana", "beto"},
		},
		existeSesion:     true,
		existeIntegrante: true,
	}
}

func votar(repo *repositorioVotosFalso, integranteID string, valor int) (planningpoker.Votacion, error) {
	return app.NuevoVotarEnSesion(repo).Ejecutar(context.Background(), "sesion-1", integranteID, valor)
}

// US-015 / CA-015-1, RN-015-5, RN-015-6: el caso de uso verifica el integrante en el
// proyecto de la sesion, valida con el dominio y guarda el voto una sola vez. Devuelve la
// votacion, que dice quien voto y no los valores (RN-015-7).
func TestVotarEnSesion_Vota(t *testing.T) {
	repo := sesionVotando()

	votacion, err := votar(repo, "ana", 5)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if !slices.Equal(votacion.Votaron(), []string{"ana"}) || !slices.Equal(votacion.Faltan(), []string{"beto"}) {
		t.Errorf("votaron %q y faltan %q, se esperaban [ana] y [beto]", votacion.Votaron(), votacion.Faltan())
	}
	if _, err := votacion.Valores(); !errors.Is(err, planningpoker.ErrVotosOcultos) {
		t.Errorf("la votacion devuelta deja leer los valores: %v", err)
	}
	if !slices.Equal(repo.integrantesConsultados, []string{"proyecto-1/ana"}) {
		t.Errorf("se consulto %q, se esperaba el integrante en el proyecto de la sesion", repo.integrantesConsultados)
	}
	if len(repo.guardados) != 1 || repo.guardados[0] != (votoGuardado{sesionID: "sesion-1", integranteID: "ana", valor: 5}) {
		t.Errorf("se guardo %+v, se esperaba el voto de ana en la sesion 1", repo.guardados)
	}
}

// US-015 / CA-015-7, RN-015-1, CL-015-10: si la sesion o el integrante no existen (o el
// integrante es de otro proyecto), se devuelve ese error sin validar nada mas y no se
// guarda nada. Si no existe la sesion, no se mira el integrante.
func TestVotarEnSesion_LoQueNoExiste(t *testing.T) {
	casos := []struct {
		nombre      string
		romper      func(*repositorioVotosFalso)
		errEsperado error
		id          string
	}{
		{"sesion inexistente", func(r *repositorioVotosFalso) {
			r.existeSesion = false
			r.existeIntegrante = false
		}, app.ErrSesionInexistente, "sesion-1"},
		{"integrante inexistente o de otro proyecto", func(r *repositorioVotosFalso) { r.existeIntegrante = false }, app.ErrIntegranteInexistente, "dani"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := sesionVotando()
			c.romper(repo)

			// Ademas el valor esta fuera de la escala: si se validara, el error seria otro.
			_, err := votar(repo, "dani", 4)

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if errors.Is(err, backlog.ErrEstimacionFueraDeEscala) || errors.Is(err, planningpoker.ErrNoEsParticipante) {
				t.Errorf("el error %q se unio con los de validacion", err)
			}
			if !strings.Contains(err.Error(), c.id) {
				t.Errorf("el error %q no indica el ID %s", err, c.id)
			}
			if len(repo.guardados) != 0 {
				t.Errorf("se guardaron %d votos, se esperaba ninguno", len(repo.guardados))
			}
		})
	}
	repo := sesionVotando()
	repo.existeSesion = false
	_, _ = votar(repo, "ana", 5)
	if len(repo.integrantesConsultados) != 0 {
		t.Error("con la sesion inexistente se consulto el integrante")
	}
}

// US-015 / CA-015-4, CA-015-5, CA-015-6, RN-015-2 a RN-015-4: lo que rechaza el dominio
// se devuelve tal cual y no se guarda nada.
func TestVotarEnSesion_ReglasDelDominio(t *testing.T) {
	casos := []struct {
		nombre      string
		estado      planningpoker.Estado
		integrante  string
		valor       int
		errEsperado error
	}{
		{"fuera de la escala", planningpoker.EstadoVotando, "ana", 4, backlog.ErrEstimacionFueraDeEscala},
		{"no participa", planningpoker.EstadoVotando, "dani", 5, planningpoker.ErrNoEsParticipante},
		{"sesion revelada", planningpoker.EstadoRevelada, "ana", 5, planningpoker.ErrVotacionCerrada},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := sesionVotando()
			repo.sesion.Estado = c.estado

			_, err := votar(repo, c.integrante, c.valor)

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if len(repo.guardados) != 0 {
				t.Errorf("se guardaron %d votos, se esperaba ninguno", len(repo.guardados))
			}
		})
	}
}

// US-015 / RN-015-9, CL-015-9: lo que el repositorio rechaza dentro de la transaccion (la
// sesion se revelo o desaparecio en el medio) se devuelve reconocible.
func TestVotarEnSesion_RechazosDelRepositorio(t *testing.T) {
	for _, errRegistrar := range []error{planningpoker.ErrVotacionCerrada, app.ErrSesionInexistente} {
		t.Run(errRegistrar.Error(), func(t *testing.T) {
			repo := sesionVotando()
			repo.errRegistrar = errRegistrar

			if _, err := votar(repo, "ana", 5); !errors.Is(err, errRegistrar) {
				t.Errorf("error = %v, se esperaba %v", err, errRegistrar)
			}
		})
	}
}

// Un fallo del repositorio en cualquier lectura o al guardar se devuelve envuelto y no es
// un inexistente.
func TestVotarEnSesion_FallosDeInfraestructura(t *testing.T) {
	falla := errors.New("se corto la conexion")
	casos := map[string]func(*repositorioVotosFalso){
		"al leer la sesion":          func(r *repositorioVotosFalso) { r.errSesion = falla },
		"al consultar el integrante": func(r *repositorioVotosFalso) { r.errIntegrante = falla },
		"al guardar el voto":         func(r *repositorioVotosFalso) { r.errRegistrar = falla },
	}
	for nombre, romper := range casos {
		t.Run(nombre, func(t *testing.T) {
			repo := sesionVotando()
			romper(repo)

			_, err := votar(repo, "ana", 5)

			if !errors.Is(err, falla) {
				t.Fatalf("error = %v, se esperaba el del repositorio", err)
			}
			if errors.Is(err, app.ErrSesionInexistente) || errors.Is(err, app.ErrIntegranteInexistente) {
				t.Errorf("un fallo del repositorio no es un inexistente: %v", err)
			}
		})
	}
}

// US-015 / spec seccion 7: el mensaje que ve la persona.
func TestErrSesionInexistente_Mensaje(t *testing.T) {
	if app.ErrSesionInexistente.Error() != "la sesion de planning poker no existe" {
		t.Errorf("mensaje = %q", app.ErrSesionInexistente.Error())
	}
}
