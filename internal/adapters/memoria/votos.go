package memoria

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// claveDeVoto identifica un voto: como mucho uno por sesion, ronda e integrante
// (RN-015-6).
type claveDeVoto struct {
	sesionID     string
	ronda        int
	integranteID string
}

// Base cumple el puerto del caso de uso de US-015 (ExisteIntegranteEnProyecto es el de
// US-019).
var _ app.RepositorioVotos = (*Base)(nil)

// ObtenerSesion busca la sesion por su ID. Cualquier ID que no este, incluido uno mal
// formado, es una sesion inexistente y no un error.
func (b *Base) ObtenerSesion(_ context.Context, sesionID string) (planningpoker.SesionPlanningPoker, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s := b.sesionGuardada(sesionID)
	if s == nil {
		return planningpoker.SesionPlanningPoker{}, false, nil
	}
	return copiaDeSesion(*s), true, nil
}

// RegistrarVoto replica lo que hace la persistencia (RN-015-9), con mu tomado en lugar del
// candado sobre la sesion: verifica que la sesion este, vuelve a verificar el voto con
// planningpoker.VerificarVoto con lo guardado y guarda el voto en la ronda actual,
// reemplazando el anterior del mismo participante (RN-015-5). Devuelve la votacion de la
// ronda. Es de todo o nada.
func (b *Base) RegistrarVoto(_ context.Context, sesionID, integranteID string, valor int) (planningpoker.Votacion, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s := b.sesionGuardada(sesionID)
	if s == nil {
		return planningpoker.Votacion{}, fmt.Errorf("%w (id %s)", app.ErrSesionInexistente, sesionID)
	}
	if err := planningpoker.VerificarVoto(*s, integranteID, valor); err != nil {
		return planningpoker.Votacion{}, err
	}

	b.votos[claveDeVoto{sesionID: sesionID, ronda: s.Ronda, integranteID: integranteID}] = valor
	return b.votacion(*s), nil
}

// VotacionDeLaSesion devuelve la votacion de la ronda actual de la sesion. La usan los
// escenarios para ver quien voto: la votacion no deja leer los valores mientras se vota.
func (b *Base) VotacionDeLaSesion(sesionID string) (planningpoker.Votacion, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s := b.sesionGuardada(sesionID)
	if s == nil {
		return planningpoker.Votacion{}, false
	}
	return b.votacion(*s), true
}

// votacion arma la votacion de la ronda actual con los votos guardados. Se llama con mu
// tomado.
func (b *Base) votacion(s planningpoker.SesionPlanningPoker) planningpoker.Votacion {
	votos := map[string]int{}
	for clave, valor := range b.votos {
		if clave.sesionID == s.ID && clave.ronda == s.Ronda {
			votos[clave.integranteID] = valor
		}
	}
	return planningpoker.NuevaVotacion(s, votos)
}

// sesionGuardada devuelve la sesion guardada, o nil si no esta. Se llama con mu tomado.
func (b *Base) sesionGuardada(sesionID string) *planningpoker.SesionPlanningPoker {
	for i := range b.sesiones {
		if b.sesiones[i].ID == sesionID {
			return &b.sesiones[i]
		}
	}
	return nil
}
