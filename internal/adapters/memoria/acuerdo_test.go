package memoria_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// conSesionRevelada devuelve la base con una sesion revelada sobre la historia.
func conSesionRevelada(t *testing.T) (baseParaPoker, planningpoker.SesionPlanningPoker) {
	t.Helper()
	b, sesion := conSesion(t)
	if _, err := b.base.RegistrarVoto(context.Background(), sesion.ID, b.integrantes[0], 3); err != nil {
		t.Fatalf("RegistrarVoto: %v", err)
	}
	if _, err := b.base.GuardarRevelado(context.Background(), sesion.ID, b.integrantes[0]); err != nil {
		t.Fatalf("GuardarRevelado: %v", err)
	}
	return b, sesion
}

func puntosDeLaHistoria(t *testing.T, b baseParaPoker) (int, bool) {
	t.Helper()
	item, _, err := b.base.ObtenerItem(context.Background(), b.item)
	if err != nil {
		t.Fatalf("ObtenerItem: %v", err)
	}
	return item.StoryPoints.Valor()
}

// US-018 / CA-018-1, CA-018-2, RN-018-8: guardar el acuerdo finaliza la sesion con el valor
// (aunque nadie lo haya votado) y estima la historia con el. ObtenerSesion lee el valor.
func TestBase_GuardarAcuerdo(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesionRevelada(t)

	finalizada, err := b.base.GuardarAcuerdo(ctx, sesion.ID, b.integrantes[1], 5)
	if err != nil {
		t.Fatalf("GuardarAcuerdo: %v", err)
	}

	for nombre, s := range map[string]planningpoker.SesionPlanningPoker{"devuelta": finalizada, "guardada": leerSesionEnMemoria(t, b, sesion.ID)} {
		if puntos, ok := s.ValorAcordado.Valor(); s.Estado != planningpoker.EstadoFinalizada || !ok || puntos != 5 {
			t.Errorf("%s = (%q, %d, %t), se esperaba (finalizada, 5, true)", nombre, s.Estado, puntos, ok)
		}
	}
	if puntos, ok := puntosDeLaHistoria(t, b); !ok || puntos != 5 {
		t.Errorf("story points de la historia = (%d, %t), se esperaba 5", puntos, ok)
	}
	if valor, ok := b.base.VotoGuardado(sesion.ID, 1, b.integrantes[0]); !ok || valor != 3 {
		t.Errorf("el voto cambio: (%d, %t)", valor, ok)
	}
}

func leerSesionEnMemoria(t *testing.T, b baseParaPoker, sesionID string) planningpoker.SesionPlanningPoker {
	t.Helper()
	s, ok, err := b.base.ObtenerSesion(context.Background(), sesionID)
	if err != nil || !ok {
		t.Fatalf("ObtenerSesion = (%t, %v)", ok, err)
	}
	return s
}

// US-018 / CA-018-3, CL-018-5: una historia ya estimada queda con el valor acordado.
func TestBase_GuardarAcuerdo_HistoriaYaEstimada(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesionRevelada(t)
	tres, _ := backlog.NuevosStoryPoints(3)
	if err := b.base.GuardarEstimacion(ctx, b.item, tres); err != nil {
		t.Fatalf("GuardarEstimacion: %v", err)
	}

	if _, err := b.base.GuardarAcuerdo(ctx, sesion.ID, b.integrantes[0], 8); err != nil {
		t.Fatalf("GuardarAcuerdo: %v", err)
	}

	if puntos, _ := puntosDeLaHistoria(t, b); puntos != 8 {
		t.Errorf("story points = %d, se esperaba el valor acordado (8)", puntos)
	}
}

// US-018 / RN-018-9, CL-018-6, CL-018-7, CL-018-11: la base vuelve a verificar con lo
// guardado. Un segundo registro (la sesion ya esta finalizada), una sesion votando, un
// integrante que no participa o un valor fuera de la escala se rechazan, y no cambia nada.
func TestBase_GuardarAcuerdo_Reglas(t *testing.T) {
	ctx := context.Background()

	t.Run("segundo registro", func(t *testing.T) {
		b, sesion := conSesionRevelada(t)
		if _, err := b.base.GuardarAcuerdo(ctx, sesion.ID, b.integrantes[0], 5); err != nil {
			t.Fatalf("primer registro: %v", err)
		}
		if _, err := b.base.GuardarAcuerdo(ctx, sesion.ID, b.integrantes[0], 8); !errors.Is(err, planningpoker.ErrSesionNoRevelada) {
			t.Errorf("error = %v, se esperaba ErrSesionNoRevelada", err)
		}
		if puntos, _ := puntosDeLaHistoria(t, b); puntos != 5 {
			t.Errorf("story points = %d, se esperaba el del primer registro (5)", puntos)
		}
	})

	casos := []struct {
		nombre      string
		votando     bool
		integrante  func(baseParaPoker) string
		valor       int
		errEsperado error
	}{
		{"sesion votando", true, func(b baseParaPoker) string { return b.integrantes[0] }, 5, planningpoker.ErrSesionNoRevelada},
		{"no participa", false, func(b baseParaPoker) string { return idSinDuenoAsignaciones }, 5, planningpoker.ErrNoEsParticipante},
		{"fuera de la escala", false, func(b baseParaPoker) string { return b.integrantes[0] }, 4, backlog.ErrEstimacionFueraDeEscala},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			b, sesion := conSesionRevelada(t)
			if c.votando {
				b.base.CambiarSesion(sesion.ID, planningpoker.EstadoVotando, 1)
			}

			if _, err := b.base.GuardarAcuerdo(ctx, sesion.ID, c.integrante(b), c.valor); !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if s := leerSesionEnMemoria(t, b, sesion.ID); s.Estado == planningpoker.EstadoFinalizada {
				t.Errorf("se finalizo la sesion: %+v", s)
			}
			if _, ok := puntosDeLaHistoria(t, b); ok {
				t.Error("se estimo la historia")
			}
		})
	}
}

// US-018 / CA-018-7, RN-018-6, CL-018-8, CL-018-10: con la historia completada (aunque se
// haya completado despues de la lectura del caso de uso) se rechaza con ErrItemCompletado
// y la sesion sigue revelada.
func TestBase_GuardarAcuerdo_HistoriaCompletada(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesionRevelada(t)
	b.base.CompletarItem(b.item)

	_, err := b.base.GuardarAcuerdo(ctx, sesion.ID, b.integrantes[0], 5)

	if !errors.Is(err, backlog.ErrItemCompletado) {
		t.Fatalf("error = %v, se esperaba ErrItemCompletado", err)
	}
	if s := leerSesionEnMemoria(t, b, sesion.ID); s.Estado != planningpoker.EstadoRevelada {
		t.Errorf("estado de la sesion = %q, se esperaba revelada", s.Estado)
	}
	if _, ok := puntosDeLaHistoria(t, b); ok {
		t.Error("se estimo una historia completada")
	}
}

// US-018 / CA-018-8: registrar en una sesion que no esta se rechaza con
// ErrSesionInexistente.
func TestBase_GuardarAcuerdo_SesionInexistente(t *testing.T) {
	b, _ := conSesionRevelada(t)

	_, err := b.base.GuardarAcuerdo(context.Background(), idSinDuenoAsignaciones, b.integrantes[0], 5)

	if !errors.Is(err, app.ErrSesionInexistente) {
		t.Errorf("error = %v, se esperaba ErrSesionInexistente", err)
	}
}

// UltimaSesionDeLaHistoria devuelve la ultima sesion de la historia, abierta o finalizada.
// La usan los escenarios: una sesion finalizada ya no esta abierta (RN-014-7).
func TestBase_UltimaSesionDeLaHistoria(t *testing.T) {
	ctx := context.Background()
	b, sesion := conSesionRevelada(t)
	if _, err := b.base.GuardarAcuerdo(ctx, sesion.ID, b.integrantes[0], 5); err != nil {
		t.Fatalf("GuardarAcuerdo: %v", err)
	}
	if _, abierta := b.base.SesionAbiertaDeLaHistoria(b.item); abierta {
		t.Fatal("la sesion finalizada sigue abierta")
	}

	ultima, ok := b.base.UltimaSesionDeLaHistoria(b.item)
	if !ok || ultima.ID != sesion.ID || ultima.Estado != planningpoker.EstadoFinalizada {
		t.Errorf("UltimaSesionDeLaHistoria = (%+v, %t), se esperaba la finalizada", ultima, ok)
	}

	nueva, err := b.base.RegistrarSesion(ctx, b.proyecto, sesionPara(t, b))
	if err != nil {
		t.Fatalf("RegistrarSesion: %v", err)
	}
	if ultima, _ := b.base.UltimaSesionDeLaHistoria(b.item); ultima.ID != nueva.ID {
		t.Errorf("ultima = %s, se esperaba la sesion nueva %s", ultima.ID, nueva.ID)
	}
	if _, ok := b.base.UltimaSesionDeLaHistoria(idSinDuenoAsignaciones); ok {
		t.Error("una historia sin sesiones tiene ultima sesion")
	}
}
