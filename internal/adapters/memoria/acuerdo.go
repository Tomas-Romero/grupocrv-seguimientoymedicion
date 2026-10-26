package memoria

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// Base cumple el puerto del caso de uso de US-018 (ObtenerSesion es el de US-015,
// ExisteIntegranteEnProyecto el de US-019 y ObtenerItem el de US-013).
var _ app.RepositorioAcuerdo = (*Base)(nil)

// GuardarAcuerdo replica lo que hace la persistencia (RN-018-9), con mu tomado en lugar del
// candado sobre la sesion: verifica que la sesion este, vuelve a verificar con
// planningpoker.Acordar con lo guardado (tambien si la historia se completo) y, si puede,
// estima la historia y finaliza la sesion con el valor. Es de todo o nada.
func (b *Base) GuardarAcuerdo(_ context.Context, sesionID, integranteID string, valor int) (planningpoker.SesionPlanningPoker, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s := b.sesionGuardada(sesionID)
	if s == nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", app.ErrSesionInexistente, sesionID)
	}
	item := b.itemGuardado(s.ItemID)
	if item == nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", app.ErrItemInexistente, s.ItemID)
	}
	finalizada, err := planningpoker.Acordar(*s, item.Estado == backlog.EstadoCompletado, integranteID, valor)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	item.StoryPoints = finalizada.ValorAcordado
	*s = finalizada
	return copiaDeSesion(finalizada), nil
}

// UltimaSesionDeLaHistoria devuelve la ultima sesion de la historia, abierta o finalizada.
// La usan los escenarios: una sesion finalizada ya no esta abierta (RN-014-7).
func (b *Base) UltimaSesionDeLaHistoria(itemID string) (planningpoker.SesionPlanningPoker, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := len(b.sesiones) - 1; i >= 0; i-- {
		if b.sesiones[i].ItemID == itemID {
			return copiaDeSesion(b.sesiones[i]), true
		}
	}
	return planningpoker.SesionPlanningPoker{}, false
}
