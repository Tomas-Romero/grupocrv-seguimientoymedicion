package memoria_test

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// US-016: la base en memoria cumple el puerto del caso de uso de revelar.
var _ app.RepositorioRevelar = (*memoria.Base)(nil)

// votarTodos guarda el voto de cada integrante de la lista, en la ronda actual de la sesion.
func votarTodos(t *testing.T, b baseParaPoker, sesionID string, valores map[string]int) {
	t.Helper()
	for integrante, valor := range valores {
		if _, err := b.base.RegistrarVoto(context.Background(), sesionID, integrante, valor); err != nil {
			t.Fatalf("RegistrarVoto de %s: %v", integrante, err)
		}
	}
}

// US-016 / RN-016-7, RN-015-7: la votacion que entrega la base es la de la ronda actual (los votos de una
// ronda anterior no entran) y, mientras se vota, no deja leer los valores: el puerto nunca los entrega en
// crudo. Una sesion que no esta, o con un ID mal formado, no tiene votos.
func TestBase_VotacionDeLaRonda(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesion(t)
	ana, beto, caro := b.integrantes[0], b.integrantes[1], b.integrantes[2]
	votarTodos(t, b, sesion.ID, map[string]int{ana: 5, beto: 8})

	v, err := b.base.VotacionDeLaRonda(ctx, sesion.ID)
	if err != nil || !slices.Equal(v.Votaron(), []string{ana, beto}) || !slices.Equal(v.Faltan(), []string{caro}) {
		t.Fatalf("votaron %q y faltan %q (%v); se esperaba ana y beto, y falta caro", v.Votaron(), v.Faltan(), err)
	}
	if _, err := v.Valores(); !errors.Is(err, planningpoker.ErrVotosOcultos) {
		t.Errorf("la votacion deja leer los valores mientras se vota: %v", err)
	}

	// Revelada, los valores se pueden leer.
	b.base.CambiarSesion(sesion.ID, planningpoker.EstadoRevelada, 1)
	revelada, _ := b.base.VotacionDeLaRonda(ctx, sesion.ID)
	if valores, err := revelada.Valores(); err != nil || !reflect.DeepEqual(valores, map[string]int{ana: 5, beto: 8}) {
		t.Errorf("Valores con la sesion revelada = %v, %v; se esperaba el voto de ana y de beto", valores, err)
	}

	b.base.CambiarSesion(sesion.ID, planningpoker.EstadoVotando, 2)
	v, err = b.base.VotacionDeLaRonda(ctx, sesion.ID)
	if err != nil || len(v.Votaron()) != 0 || len(v.Faltan()) != 3 {
		t.Errorf("en la ronda 2 votaron %q y faltan %q (%v); se esperaba que no hubiera votos (los de la ronda 1 no entran)", v.Votaron(), v.Faltan(), err)
	}

	for _, id := range []string{idSinDuenoAsignaciones, "no-es-un-uuid"} {
		if v, err := b.base.VotacionDeLaRonda(ctx, id); err != nil || len(v.Votaron()) != 0 {
			t.Errorf("VotacionDeLaRonda(%q) votaron %q (%v); se esperaba una votacion vacia y sin error", id, v.Votaron(), err)
		}
	}
}

// US-016 / CA-016-1, CA-016-2, CA-016-3: revelar pasa la sesion a revelada, no cambia los votos
// y devuelve la revelacion: quienes votaron, quienes faltan y si hay divergencia.
func TestBase_GuardarRevelado(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesion(t)
	ana, beto, caro := b.integrantes[0], b.integrantes[1], b.integrantes[2]
	votarTodos(t, b, sesion.ID, map[string]int{ana: 5, beto: 8})

	r, err := b.base.GuardarRevelado(ctx, sesion.ID, caro)
	if err != nil {
		t.Fatalf("GuardarRevelado: %v", err)
	}

	quiero := planningpoker.Revelacion{
		SesionID: sesion.ID, Ronda: 1,
		Votos:    []planningpoker.Voto{{ParticipanteID: ana, Valor: 5}, {ParticipanteID: beto, Valor: 8}},
		Faltaron: []string{caro}, Divergencia: true, Minimo: 5, Maximo: 8,
	}
	if !reflect.DeepEqual(r, quiero) {
		t.Errorf("revelacion = %+v, se esperaba %+v", r, quiero)
	}
	guardada, _, _ := b.base.ObtenerSesion(ctx, sesion.ID)
	if guardada.Estado != planningpoker.EstadoRevelada || guardada.Ronda != 1 {
		t.Errorf("la sesion quedo %q en la ronda %d, se esperaba revelada en la 1", guardada.Estado, guardada.Ronda)
	}
	if valor, _ := b.base.VotoGuardado(sesion.ID, 1, ana); valor != 5 {
		t.Errorf("el voto de ana quedo en %d: revelar no tiene que cambiar los votos", valor)
	}
}

// US-016 / CA-016-5, CA-016-6, CA-016-7, RN-016-8: el repositorio vuelve a verificar con la
// misma funcion del dominio y, si algo no se cumple, no cambia nada.
func TestBase_GuardarRevelado_Rechazos(t *testing.T) {
	ctx := context.Background()
	casos := []struct {
		nombre string
		armar  func(t *testing.T, b baseParaPoker, sesionID string)
		quien  int
		quiero error
	}{
		{"nadie voto", func(*testing.T, baseParaPoker, string) {}, 0, planningpoker.ErrSinVotos},
		{"quien revela no participa", func(t *testing.T, b baseParaPoker, id string) {
			votarTodos(t, b, id, map[string]int{b.integrantes[0]: 5})
		}, -1, planningpoker.ErrNoEsParticipante},
		{"la sesion ya esta revelada", func(t *testing.T, b baseParaPoker, id string) {
			votarTodos(t, b, id, map[string]int{b.integrantes[0]: 5})
			b.base.CambiarSesion(id, planningpoker.EstadoRevelada, 1)
		}, 0, planningpoker.ErrVotacionCerrada},
		{"la sesion esta finalizada", func(t *testing.T, b baseParaPoker, id string) {
			votarTodos(t, b, id, map[string]int{b.integrantes[0]: 5})
			b.base.CambiarSesion(id, planningpoker.EstadoFinalizada, 1)
		}, 0, planningpoker.ErrVotacionCerrada},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			b, sesion := conSesion(t)
			c.armar(t, b, sesion.ID)
			antes, _, _ := b.base.ObtenerSesion(ctx, sesion.ID)
			quien := "integrante-que-no-participa"
			if c.quien >= 0 {
				quien = b.integrantes[c.quien]
			}

			r, err := b.base.GuardarRevelado(ctx, sesion.ID, quien)

			if !errors.Is(err, c.quiero) {
				t.Fatalf("error = %v, se esperaba %v", err, c.quiero)
			}
			if !reflect.DeepEqual(r, planningpoker.Revelacion{}) {
				t.Errorf("con error la revelacion tiene que estar vacia: %+v", r)
			}
			despues, _, _ := b.base.ObtenerSesion(ctx, sesion.ID)
			if despues.Estado != antes.Estado {
				t.Errorf("la sesion paso de %q a %q: un rechazo no cambia nada", antes.Estado, despues.Estado)
			}
		})
	}
}

// US-016 / RN-016-8: una sesion que no esta, o con un ID mal formado, es inexistente.
func TestBase_GuardarRevelado_SesionInexistente(t *testing.T) {
	b, _ := conSesion(t)
	for _, id := range []string{idSinDuenoAsignaciones, "no-es-un-uuid"} {
		if _, err := b.base.GuardarRevelado(context.Background(), id, b.integrantes[0]); !errors.Is(err, app.ErrSesionInexistente) {
			t.Errorf("GuardarRevelado(%q) = %v, se esperaba ErrSesionInexistente", id, err)
		}
	}
}

// US-016 / CA-016-7, CL-016-11: despues de revelar ya no se vota, y los votos revelados no
// cambian.
func TestBase_GuardarRevelado_DespuesNoSeVota(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesion(t)
	ana, beto := b.integrantes[0], b.integrantes[1]
	votarTodos(t, b, sesion.ID, map[string]int{ana: 5})
	if _, err := b.base.GuardarRevelado(ctx, sesion.ID, ana); err != nil {
		t.Fatalf("GuardarRevelado: %v", err)
	}

	_, err := b.base.RegistrarVoto(ctx, sesion.ID, beto, 8)

	if !errors.Is(err, planningpoker.ErrVotacionCerrada) {
		t.Fatalf("votar con la sesion revelada = %v, se esperaba ErrVotacionCerrada", err)
	}
	v, _ := b.base.VotacionDeLaRonda(ctx, sesion.ID)
	if !slices.Equal(v.Votaron(), []string{ana}) {
		t.Errorf("votaron %q, se esperaba solo ana", v.Votaron())
	}
}
