package planningpoker_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// US-018 / CA-018-1, CA-018-2, RN-018-2, RN-018-5, RN-018-8: con la sesion revelada, un
// participante acuerda cualquier valor de la escala (votado o no, con o sin divergencia).
// La sesion queda finalizada con ese valor, y el resto queda igual.
func TestAcordar(t *testing.T) {
	for _, valor := range backlog.EscalaFibonacci() {
		sesion := sesionEn(planningpoker.EstadoRevelada)

		finalizada, err := planningpoker.Acordar(sesion, false, "beto", valor)
		if err != nil {
			t.Fatalf("Acordar(%d): %v", valor, err)
		}

		if puntos, ok := finalizada.ValorAcordado.Valor(); finalizada.Estado != planningpoker.EstadoFinalizada || !ok || puntos != valor {
			t.Errorf("Acordar(%d) = (%q, %d, %t), se esperaba (finalizada, %d, true)", valor, finalizada.Estado, puntos, ok, valor)
		}
		if finalizada.ID != sesion.ID || finalizada.ItemID != sesion.ItemID || finalizada.Ronda != sesion.Ronda ||
			!slices.Equal(finalizada.Participantes, sesion.Participantes) {
			t.Errorf("Acordar cambio algo mas que el estado y el valor: %+v", finalizada)
		}
	}
}

// Acordar devuelve una copia: la sesion recibida no cambia y no comparte los
// participantes con la finalizada.
func TestAcordar_NoModificaLaSesion(t *testing.T) {
	sesion := sesionEn(planningpoker.EstadoRevelada)

	finalizada, err := planningpoker.Acordar(sesion, false, "ana", 5)
	if err != nil {
		t.Fatalf("Acordar: %v", err)
	}
	finalizada.Participantes[0] = "cambiado"

	if sesion.Estado != planningpoker.EstadoRevelada || sesion.ValorAcordado != backlog.SinEstimar() || sesion.Participantes[0] != "ana" {
		t.Errorf("la sesion recibida cambio: %+v", sesion)
	}
}

// US-014 / RN-014-8, US-018 / spec seccion 3: una sesion sin acordar no tiene valor.
func TestSesion_SinValorAcordado(t *testing.T) {
	sesion, err := planningpoker.NuevaSesion("item-1", planningpoker.DatosDeLaHistoria{}, []string{"ana", "beto"})
	if err != nil {
		t.Fatalf("NuevaSesion: %v", err)
	}
	if _, acordado := sesion.ValorAcordado.Valor(); acordado {
		t.Errorf("una sesion nueva tiene valor acordado: %+v", sesion.ValorAcordado)
	}
}

// US-018 / CA-018-4, CA-018-5, CA-018-6, RN-018-2 a RN-018-4, CL-018-4, CL-018-6, CL-018-7:
// una sesion que no esta revelada, un integrante que no participa o un valor fuera de la
// escala se rechazan.
func TestAcordar_Rechazos(t *testing.T) {
	casos := []struct {
		nombre      string
		estado      planningpoker.Estado
		integrante  string
		valor       int
		errEsperado error
	}{
		{"sesion votando", planningpoker.EstadoVotando, "ana", 5, planningpoker.ErrSesionNoRevelada},
		{"sesion finalizada", planningpoker.EstadoFinalizada, "ana", 5, planningpoker.ErrSesionNoRevelada},
		{"no participa", planningpoker.EstadoRevelada, "dani", 5, planningpoker.ErrNoEsParticipante},
		{"cero", planningpoker.EstadoRevelada, "ana", 0, backlog.ErrEstimacionFueraDeEscala},
		{"cuatro", planningpoker.EstadoRevelada, "ana", 4, backlog.ErrEstimacionFueraDeEscala},
		{"veintiuno", planningpoker.EstadoRevelada, "ana", 21, backlog.ErrEstimacionFueraDeEscala},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			finalizada, err := planningpoker.Acordar(sesionEn(c.estado), false, c.integrante, c.valor)

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if finalizada.ID != "" {
				t.Errorf("devolvio la sesion %+v, se esperaba el valor cero", finalizada)
			}
		})
	}
}

// US-018 / RN-018-7, CL-018-9: fuera del caso de la historia completada, los errores se
// unen en orden fijo: sesion no revelada, no es participante, valor fuera de la escala.
func TestAcordar_ErroresUnidosEnOrden(t *testing.T) {
	_, err := planningpoker.Acordar(sesionEn(planningpoker.EstadoVotando), false, "dani", 4)

	esperados := []error{planningpoker.ErrSesionNoRevelada, planningpoker.ErrNoEsParticipante, backlog.ErrEstimacionFueraDeEscala}
	posicion := -1
	for _, esperado := range esperados {
		if !errors.Is(err, esperado) {
			t.Fatalf("error = %v, se esperaba que incluya %v", err, esperado)
		}
		i := strings.Index(err.Error(), esperado.Error())
		if i < posicion {
			t.Errorf("el error %q no respeta el orden de RN-018-7", err)
		}
		posicion = i
	}
}

// US-018 / CA-018-7, RN-018-6, CL-018-8: una historia completada se valida primero y su
// error va solo, aunque la sesion no este revelada, el integrante no participe y el valor
// sea invalido (mismo criterio que RN-013-5).
func TestAcordar_HistoriaCompletada(t *testing.T) {
	casos := map[string]struct {
		estado     planningpoker.Estado
		integrante string
		valor      int
	}{
		"todo lo demas en regla": {planningpoker.EstadoRevelada, "ana", 5},
		"todo lo demas invalido": {planningpoker.EstadoVotando, "dani", 4},
	}
	for nombre, c := range casos {
		t.Run(nombre, func(t *testing.T) {
			_, err := planningpoker.Acordar(sesionEn(c.estado), true, c.integrante, c.valor)

			if !errors.Is(err, backlog.ErrItemCompletado) {
				t.Fatalf("error = %v, se esperaba ErrItemCompletado", err)
			}
			for _, otro := range []error{planningpoker.ErrSesionNoRevelada, planningpoker.ErrNoEsParticipante, backlog.ErrEstimacionFueraDeEscala} {
				if errors.Is(err, otro) {
					t.Errorf("el error %q incluye %v: con la historia completada va solo", err, otro)
				}
			}
		})
	}
}

// US-018 / spec seccion 7: el mensaje que ve la persona.
func TestErrSesionNoRevelada_Mensaje(t *testing.T) {
	const esperado = "la estimacion acordada se registra con los votos revelados y la sesion sin finalizar"
	if planningpoker.ErrSesionNoRevelada.Error() != esperado {
		t.Errorf("mensaje = %q, se esperaba %q", planningpoker.ErrSesionNoRevelada.Error(), esperado)
	}
}
