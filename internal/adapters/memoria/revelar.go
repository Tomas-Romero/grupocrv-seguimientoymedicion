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

// VotacionDeLaRonda devuelve la votacion de la ronda actual de la sesion. Mientras se vota no deja
// leer los valores (RN-015-7). Una sesion que no esta, incluido un ID mal formado, no tiene votos:
// devuelve una votacion vacia.
func (b *Base) VotacionDeLaRonda(_ context.Context, sesionID string) (planningpoker.Votacion, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s := b.sesionGuardada(sesionID)
	if s == nil {
		return planningpoker.Votacion{}, nil
	}
	return b.votacion(*s), nil
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
	revelacion, err := planningpoker.Revelar(*s, integranteID, b.votacion(*s))
	if err != nil {
		return planningpoker.Revelacion{}, err
	}
	s.Estado = planningpoker.EstadoRevelada
	return revelacion, nil
}
