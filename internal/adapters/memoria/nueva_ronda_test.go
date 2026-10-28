package memoria_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// US-017: la base en memoria cumple el puerto del caso de uso y la lectura de rondas anteriores.
var (
	_ app.RepositorioNuevaRonda    = (*memoria.Base)(nil)
	_ app.LectorDeRondasAnteriores = (*memoria.Base)(nil)
)

// conDivergencia es una sesion revelada en la que ana voto 3 y beto 8.
func conDivergencia(t *testing.T) (baseParaPoker, planningpoker.SesionPlanningPoker) {
	t.Helper()
	b, sesion := conSesion(t)
	votarTodos(t, b, sesion.ID, map[string]int{b.integrantes[0]: 3, b.integrantes[1]: 8})
	if _, err := b.base.GuardarRevelado(context.Background(), sesion.ID, b.integrantes[0]); err != nil {
		t.Fatalf("GuardarRevelado: %v", err)
	}
	return b, sesion
}

// US-017 / CA-017-1, CA-017-2, RN-017-5, RN-017-6, CL-017-5: abrir la ronda deja la sesion
// votando en la ronda 2 sin votos; los de la ronda 1 se conservan y se leen como ronda anterior,
// y un voto nuevo va a la ronda 2.
func TestBase_GuardarNuevaRonda(t *testing.T) {
	ctx := context.Background()
	b, sesion := conDivergencia(t)
	ana, beto := b.integrantes[0], b.integrantes[1]

	nueva, err := b.base.GuardarNuevaRonda(ctx, sesion.ID, beto)
	if err != nil {
		t.Fatalf("GuardarNuevaRonda: %v", err)
	}
	if nueva.Estado != planningpoker.EstadoVotando || nueva.Ronda != 2 {
		t.Fatalf("la sesion quedo %q en la ronda %d, se esperaba votando en la 2", nueva.Estado, nueva.Ronda)
	}
	guardada, _, _ := b.base.ObtenerSesion(ctx, sesion.ID)
	if !reflect.DeepEqual(guardada, nueva) {
		t.Errorf("la sesion guardada %+v no es la devuelta %+v", guardada, nueva)
	}
	if v, _ := b.base.VotacionDeLaRonda(ctx, sesion.ID); len(v.Votaron()) != 0 {
		t.Errorf("en la ronda nueva votaron %q, se esperaba nadie", v.Votaron())
	}

	votarTodos(t, b, sesion.ID, map[string]int{ana: 13})

	rondas, err := b.base.RondasAnteriores(ctx, sesion.ID)
	if err != nil || len(rondas) != 1 || rondas[0].Numero != 1 {
		t.Fatalf("RondasAnteriores = %+v (%v), se esperaba solo la ronda 1", rondas, err)
	}
	valores, err := rondas[0].Votacion.Valores()
	if err != nil || !reflect.DeepEqual(valores, map[string]int{ana: 3, beto: 8}) {
		t.Errorf("valores de la ronda 1 = %v (%v), se esperaban los de ana y beto sin cambios", valores, err)
	}
	if v, _ := b.base.VotacionDeLaRonda(ctx, sesion.ID); len(v.Votaron()) != 1 {
		t.Errorf("en la ronda 2 votaron %q, se esperaba solo ana", v.Votaron())
	}
	if _, err := mustVotacion(t, b, sesion.ID).Valores(); !errors.Is(err, planningpoker.ErrVotosOcultos) {
		t.Errorf("los votos de la ronda 2 se leen mientras se vota: %v", err)
	}
}

func mustVotacion(t *testing.T, b baseParaPoker, sesionID string) planningpoker.Votacion {
	t.Helper()
	v, err := b.base.VotacionDeLaRonda(context.Background(), sesionID)
	if err != nil {
		t.Fatalf("VotacionDeLaRonda: %v", err)
	}
	return v
}

// US-017 / CA-017-3, CA-017-4, CA-017-5, RN-017-8, RN-017-9: con lo guardado, vuelve a verificar
// con el dominio. Una sesion votando (tambien la que ya abrio la ronda) o sin divergencia se
// rechaza y no cambia nada.
func TestBase_GuardarNuevaRonda_Reglas(t *testing.T) {
	ctx := context.Background()

	t.Run("segundo pedido", func(t *testing.T) {
		b, sesion := conDivergencia(t)
		if _, err := b.base.GuardarNuevaRonda(ctx, sesion.ID, b.integrantes[0]); err != nil {
			t.Fatalf("primer pedido: %v", err)
		}
		_, err := b.base.GuardarNuevaRonda(ctx, sesion.ID, b.integrantes[1])
		if !errors.Is(err, planningpoker.ErrNuevaRondaSinRevelar) {
			t.Fatalf("segundo pedido = %v, se esperaba ErrNuevaRondaSinRevelar", err)
		}
		if s, _, _ := b.base.ObtenerSesion(ctx, sesion.ID); s.Ronda != 2 {
			t.Errorf("la sesion quedo en la ronda %d, se esperaba la 2: se abrio una sola", s.Ronda)
		}
	})

	t.Run("sin divergencia", func(t *testing.T) {
		b, sesion := conSesion(t)
		votarTodos(t, b, sesion.ID, map[string]int{b.integrantes[0]: 5, b.integrantes[1]: 5})
		b.base.CambiarSesion(sesion.ID, planningpoker.EstadoRevelada, 1)

		_, err := b.base.GuardarNuevaRonda(ctx, sesion.ID, b.integrantes[0])

		if !errors.Is(err, planningpoker.ErrSinDivergencia) {
			t.Fatalf("GuardarNuevaRonda = %v, se esperaba ErrSinDivergencia", err)
		}
		if s, _, _ := b.base.ObtenerSesion(ctx, sesion.ID); s.Estado != planningpoker.EstadoRevelada || s.Ronda != 1 {
			t.Errorf("la sesion cambio: %q en la ronda %d", s.Estado, s.Ronda)
		}
	})
}

// RN-017-8: una sesion que no esta, incluido un ID mal formado, responde ErrSesionInexistente.
func TestBase_GuardarNuevaRonda_SesionInexistente(t *testing.T) {
	b, _ := conDivergencia(t)
	for _, id := range []string{idSinDuenoAsignaciones, "no-es-un-uuid"} {
		if _, err := b.base.GuardarNuevaRonda(context.Background(), id, b.integrantes[0]); !errors.Is(err, app.ErrSesionInexistente) {
			t.Errorf("GuardarNuevaRonda(%q) = %v, se esperaba ErrSesionInexistente", id, err)
		}
	}
}

// RN-017-6: una sesion en la ronda 1, o que no esta, no tiene rondas anteriores.
func TestBase_RondasAnteriores_Ninguna(t *testing.T) {
	b, sesion := conDivergencia(t)
	for _, id := range []string{sesion.ID, idSinDuenoAsignaciones, "no-es-un-uuid"} {
		if rondas, err := b.base.RondasAnteriores(context.Background(), id); err != nil || len(rondas) != 0 {
			t.Errorf("RondasAnteriores(%q) = %+v (%v), se esperaba ninguna", id, rondas, err)
		}
	}
}
