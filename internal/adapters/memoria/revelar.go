package memoria

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// Base cumple el puerto del caso de uso de US-016 (ObtenerSesion es el de US-015 y
// ExisteIntegranteEnProyecto, el de US-019).
var _ app.RepositorioRevelar = (*Base)(nil)

// VotosDeLaRonda devuelve el voto de cada participante en la ronda actual de la sesion, por
// integrante. Es una copia. Una sesion que no esta, incluido un ID mal formado, no tiene votos.
func (b *Base) VotosDeLaRonda(_ context.Context, sesionID string) (map[string]int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.votosDeLaRonda(sesionID), nil
}

// votosDeLaRonda devuelve los votos de la ronda actual de la sesion. Se llama con mu tomado.
func (b *Base) votosDeLaRonda(sesionID string) map[string]int {
	votos := map[string]int{}
	s := b.sesionGuardada(sesionID)
	if s == nil {
		return votos
	}
	for clave, valor := range b.votos {
		if clave.sesionID == s.ID && clave.ronda == s.Ronda {
			votos[clave.integranteID] = valor
		}
	}
	return votos
}

// GuardarRevelado replica lo que hace la persistencia (RN-016-8), con mu tomado en lugar
// del candado sobre la sesion: verifica que la sesion este, vuelve a verificar con
// planningpoker.Revelar con lo guardado y, si puede, la pasa a revelada. Devuelve la
// revelacion. Es de todo o nada: si Revelar rechaza, no cambia nada.
func (b *Base) GuardarRevelado(_ context.Context, sesionID, integranteID string) (planningpoker.Revelacion, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s := b.sesionGuardada(sesionID)
	if s == nil {
		return planningpoker.Revelacion{}, fmt.Errorf("%w (id %s)", app.ErrSesionInexistente, sesionID)
	}
	revelacion, err := planningpoker.Revelar(*s, integranteID, b.votosDeLaRonda(sesionID))
	if err != nil {
		return planningpoker.Revelacion{}, err
	}
	s.Estado = planningpoker.EstadoRevelada
	return revelacion, nil
}
