package memoria_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// conSesion devuelve la base con una sesion registrada sobre la historia, con tres
// participantes.
func conSesion(t *testing.T) (baseParaPoker, planningpoker.SesionPlanningPoker) {
	t.Helper()
	b := nuevaBaseParaPoker(t, "ana", "beto", "caro")
	sesion, err := b.base.RegistrarSesion(context.Background(), b.proyecto, sesionPara(t, b))
	if err != nil {
		t.Fatalf("RegistrarSesion: %v", err)
	}
	return b, sesion
}

// US-015 / RN-015-1: la sesion se lee con su proyecto, su estado, su ronda y sus
// participantes. Una que no esta, o con un ID mal formado, es inexistente.
func TestBase_ObtenerSesion(t *testing.T) {
	ctx := context.Background()
	b, registrada := conSesion(t)

	sesion, existe, err := b.base.ObtenerSesion(ctx, registrada.ID)
	if err != nil || !existe {
		t.Fatalf("ObtenerSesion = (existe %t, %v), se esperaba encontrarla", existe, err)
	}
	if sesion.ProyectoID != b.proyecto || sesion.ItemID != b.item || sesion.Estado != planningpoker.EstadoVotando ||
		sesion.Ronda != 1 || !slices.Equal(sesion.Participantes, b.integrantes) {
		t.Errorf("sesion = %+v, se esperaba la registrada con su proyecto", sesion)
	}
	for _, id := range []string{idSinDuenoAsignaciones, "no-es-un-uuid"} {
		if _, existe, err := b.base.ObtenerSesion(ctx, id); err != nil || existe {
			t.Errorf("ObtenerSesion(%q) = (existe %t, %v), se esperaba (false, nil)", id, existe, err)
		}
	}
}

// US-015 / CA-015-1, RN-015-6: el voto se guarda en la ronda actual y la votacion que se
// devuelve dice quien voto, sin dejar leer los valores.
func TestBase_RegistrarVoto(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesion(t)
	ana := b.integrantes[0]

	votacion, err := b.base.RegistrarVoto(ctx, sesion.ID, ana, 5)
	if err != nil {
		t.Fatalf("RegistrarVoto: %v", err)
	}

	if !slices.Equal(votacion.Votaron(), []string{ana}) || len(votacion.Faltan()) != 2 {
		t.Errorf("votaron %q y faltan %q, se esperaba ana y los otros dos", votacion.Votaron(), votacion.Faltan())
	}
	if _, err := votacion.Valores(); !errors.Is(err, planningpoker.ErrVotosOcultos) {
		t.Errorf("la votacion deja leer los valores: %v", err)
	}
	if valor, ok := b.base.VotoGuardado(sesion.ID, 1, ana); !ok || valor != 5 {
		t.Errorf("voto guardado = (%d, %t), se esperaba 5 en la ronda 1", valor, ok)
	}
}

// US-015 / CA-015-2, RN-015-5, CL-015-3: votar otra vez reemplaza el voto: cuenta el
// ultimo, y el participante figura una sola vez.
func TestBase_RegistrarVoto_CambiarElVoto(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesion(t)
	ana := b.integrantes[0]
	if _, err := b.base.RegistrarVoto(ctx, sesion.ID, ana, 3); err != nil {
		t.Fatalf("primer voto: %v", err)
	}

	votacion, err := b.base.RegistrarVoto(ctx, sesion.ID, ana, 8)
	if err != nil {
		t.Fatalf("segundo voto: %v", err)
	}

	if !slices.Equal(votacion.Votaron(), []string{ana}) {
		t.Errorf("votaron %q, se esperaba ana una sola vez", votacion.Votaron())
	}
	if valor, _ := b.base.VotoGuardado(sesion.ID, 1, ana); valor != 8 {
		t.Errorf("voto guardado = %d, se esperaba el ultimo (8)", valor)
	}
}

// US-015 / RN-015-6: los votos van por ronda. En una ronda nueva (US-017) la votacion
// empieza sin votos, y los de la ronda anterior quedan guardados como estaban.
func TestBase_RegistrarVoto_PorRonda(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesion(t)
	ana, beto := b.integrantes[0], b.integrantes[1]
	if _, err := b.base.RegistrarVoto(ctx, sesion.ID, ana, 3); err != nil {
		t.Fatalf("voto de la ronda 1: %v", err)
	}
	b.base.CambiarSesion(sesion.ID, planningpoker.EstadoVotando, 2)

	votacion, err := b.base.RegistrarVoto(ctx, sesion.ID, beto, 5)
	if err != nil {
		t.Fatalf("voto de la ronda 2: %v", err)
	}

	if !slices.Equal(votacion.Votaron(), []string{beto}) {
		t.Errorf("en la ronda 2 votaron %q, se esperaba solo beto", votacion.Votaron())
	}
	if valor, ok := b.base.VotoGuardado(sesion.ID, 1, ana); !ok || valor != 3 {
		t.Errorf("el voto de la ronda 1 quedo (%d, %t), se esperaba 3", valor, ok)
	}
}

// US-015 / RN-015-9, CL-015-6, CL-015-9: la base vuelve a verificar las reglas con lo
// guardado. Un integrante que no participa, un valor fuera de la escala o una sesion que
// ya no esta votando se rechazan, y no se guarda nada.
func TestBase_RegistrarVoto_Reglas(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesion(t)
	ana := b.integrantes[0]
	dani := b.registrar(t, "dani")

	if _, err := b.base.RegistrarVoto(ctx, sesion.ID, dani, 5); !errors.Is(err, planningpoker.ErrNoEsParticipante) {
		t.Errorf("no participa: error = %v, se esperaba ErrNoEsParticipante", err)
	}
	if _, err := b.base.RegistrarVoto(ctx, sesion.ID, ana, 4); !errors.Is(err, backlog.ErrEstimacionFueraDeEscala) {
		t.Errorf("fuera de escala: error = %v, se esperaba ErrEstimacionFueraDeEscala", err)
	}
	b.base.CambiarSesion(sesion.ID, planningpoker.EstadoRevelada, 1)
	if _, err := b.base.RegistrarVoto(ctx, sesion.ID, ana, 5); !errors.Is(err, planningpoker.ErrVotacionCerrada) {
		t.Errorf("sesion revelada: error = %v, se esperaba ErrVotacionCerrada", err)
	}
	for _, id := range []string{ana, dani} {
		if _, ok := b.base.VotoGuardado(sesion.ID, 1, id); ok {
			t.Errorf("se guardo un voto rechazado de %s", id)
		}
	}
}

// US-015 / CA-015-7: votar en una sesion que no esta se rechaza con ErrSesionInexistente.
func TestBase_RegistrarVoto_SesionInexistente(t *testing.T) {
	b, _ := conSesion(t)

	_, err := b.base.RegistrarVoto(context.Background(), idSinDuenoAsignaciones, b.integrantes[0], 5)

	if !errors.Is(err, app.ErrSesionInexistente) {
		t.Errorf("error = %v, se esperaba ErrSesionInexistente", err)
	}
}

// VotacionDeLaSesion devuelve la votacion de la ronda actual. La usan los escenarios para
// ver quien voto; con la sesion votando tampoco deja leer los valores.
func TestBase_VotacionDeLaSesion(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesion(t)
	if _, err := b.base.RegistrarVoto(ctx, sesion.ID, b.integrantes[1], 13); err != nil {
		t.Fatalf("RegistrarVoto: %v", err)
	}

	votacion, ok := b.base.VotacionDeLaSesion(sesion.ID)

	if !ok || !slices.Equal(votacion.Votaron(), []string{b.integrantes[1]}) {
		t.Errorf("VotacionDeLaSesion = (votaron %q, %t), se esperaba beto", votacion.Votaron(), ok)
	}
	if _, err := votacion.Valores(); !errors.Is(err, planningpoker.ErrVotosOcultos) {
		t.Errorf("la votacion deja leer los valores: %v", err)
	}
	if _, ok := b.base.VotacionDeLaSesion(idSinDuenoAsignaciones); ok {
		t.Error("una sesion que no esta tiene votacion")
	}
}
