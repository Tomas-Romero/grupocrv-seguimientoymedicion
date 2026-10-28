package postgres_test

import (
	"reflect"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
)

// T-020: el repositorio cumple el puerto de la lectura de la pantalla de la sesion.
var _ app.RepositorioConsultaDeSesion = (*postgres.RepositorioPlanningPoker)(nil)

// T-020 / CL-T020-2: cada historia del proyecto con una sesion abierta lleva a esa sesion; una
// finalizada no cuenta, y otro proyecto o un ID mal formado no tienen ninguna.
func TestSesionesAbiertas(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionReveladaEnPostgres(ctx, t)

	abiertas, err := e.repo.SesionesAbiertas(ctx, e.proyecto)
	if err != nil || !reflect.DeepEqual(abiertas, map[string]string{sesion.ItemID: sesion.ID}) {
		t.Errorf("SesionesAbiertas = %v (%v), se esperaba la sesion de la historia", abiertas, err)
	}

	if _, err := e.repo.GuardarAcuerdo(ctx, sesion.ID, e.integrantes[0], 5); err != nil {
		t.Fatalf("GuardarAcuerdo: %v", err)
	}
	for _, id := range []string{e.proyecto, idQueNoParticipa, "no-es-un-uuid"} {
		if abiertas, err := e.repo.SesionesAbiertas(ctx, id); err != nil || len(abiertas) != 0 {
			t.Errorf("SesionesAbiertas(%q) = %v (%v), se esperaba ninguna", id, abiertas, err)
		}
	}
}
