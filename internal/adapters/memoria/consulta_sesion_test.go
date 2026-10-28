package memoria_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// T-020: la base en memoria cumple el puerto de la lectura de la pantalla de la sesion.
var _ app.RepositorioConsultaDeSesion = (*memoria.Base)(nil)

// T-020 / CL-T020-2: cada historia del proyecto con una sesion abierta lleva a esa sesion; una
// finalizada no cuenta, y otro proyecto no tiene ninguna.
func TestBase_SesionesAbiertas(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesion(t)

	abiertas, err := b.base.SesionesAbiertas(ctx, b.proyecto)
	if err != nil || !reflect.DeepEqual(abiertas, map[string]string{b.item: sesion.ID}) {
		t.Errorf("SesionesAbiertas = %v (%v), se esperaba la sesion de la historia", abiertas, err)
	}

	b.base.CambiarSesion(sesion.ID, planningpoker.EstadoFinalizada, 1)
	if abiertas, err := b.base.SesionesAbiertas(ctx, b.proyecto); err != nil || len(abiertas) != 0 {
		t.Errorf("con la sesion finalizada: %v (%v), se esperaba ninguna", abiertas, err)
	}
	for _, id := range []string{idSinDuenoAsignaciones, "no-es-un-uuid"} {
		if abiertas, err := b.base.SesionesAbiertas(ctx, id); err != nil || len(abiertas) != 0 {
			t.Errorf("SesionesAbiertas(%q) = %v (%v), se esperaba ninguna", id, abiertas, err)
		}
	}
}
