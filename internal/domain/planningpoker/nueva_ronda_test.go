package planningpoker_test

import (
	"errors"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// divergente son los votos de una ronda con divergencia: 3 y 8.
var divergente = map[string]int{"ana": 3, "beto": 8}

// US-017 / CA-017-1, RN-017-5, CL-017-1, CL-017-2: con la sesion revelada y con divergencia, un
// participante abre la ronda siguiente: la sesion queda votando, con los mismos participantes.
func TestAbrirNuevaRonda(t *testing.T) {
	for nombre, votos := range map[string]map[string]int{
		"3 y 8":      divergente,
		"contiguos":  {"ana": 5, "beto": 8},
		"falta caro": {"ana": 1, "beto": 2},
	} {
		t.Run(nombre, func(t *testing.T) {
			sesion := sesionEn(planningpoker.EstadoRevelada)

			nueva, err := planningpoker.AbrirNuevaRonda(sesion, "beto", votacionDe(sesion, votos))
			if err != nil {
				t.Fatalf("AbrirNuevaRonda: %v", err)
			}

			if nueva.Estado != planningpoker.EstadoVotando || nueva.Ronda != 2 {
				t.Errorf("la sesion quedo %q en la ronda %d, se esperaba votando en la 2", nueva.Estado, nueva.Ronda)
			}
			if nueva.ID != sesion.ID || nueva.ItemID != sesion.ItemID || !slices.Equal(nueva.Participantes, sesion.Participantes) {
				t.Errorf("AbrirNuevaRonda cambio algo mas que el estado y la ronda: %+v", nueva)
			}
		})
	}
}

// RN-017-5: no hay limite de rondas.
func TestAbrirNuevaRonda_SinLimite(t *testing.T) {
	sesion := sesionEn(planningpoker.EstadoRevelada)
	sesion.Ronda = 7

	nueva, err := planningpoker.AbrirNuevaRonda(sesion, "ana", votacionDe(sesion, divergente))
	if err != nil || nueva.Ronda != 8 {
		t.Fatalf("AbrirNuevaRonda en la ronda 7 = (ronda %d, %v), se esperaba la 8", nueva.Ronda, err)
	}
}

// AbrirNuevaRonda devuelve una copia: la sesion recibida no cambia y no comparte los
// participantes con la nueva.
func TestAbrirNuevaRonda_NoModificaLaSesion(t *testing.T) {
	sesion := sesionEn(planningpoker.EstadoRevelada)

	nueva, err := planningpoker.AbrirNuevaRonda(sesion, "ana", votacionDe(sesion, divergente))
	if err != nil {
		t.Fatalf("AbrirNuevaRonda: %v", err)
	}
	nueva.Participantes[0] = "cambiado"

	if sesion.Estado != planningpoker.EstadoRevelada || sesion.Ronda != 1 || sesion.Participantes[0] != "ana" {
		t.Errorf("la sesion recibida cambio: %+v", sesion)
	}
}

// US-017 / CA-017-3, CA-017-4, RN-017-2 a RN-017-4, CL-017-3, CL-017-7, CL-017-8: una sesion que
// no esta revelada, un integrante que no participa o una ronda sin divergencia se rechazan.
func TestAbrirNuevaRonda_Rechazos(t *testing.T) {
	casos := []struct {
		nombre      string
		estado      planningpoker.Estado
		integrante  string
		votos       map[string]int
		errEsperado error
	}{
		{"sesion votando", planningpoker.EstadoVotando, "ana", divergente, planningpoker.ErrNuevaRondaSinRevelar},
		{"sesion finalizada", planningpoker.EstadoFinalizada, "ana", divergente, planningpoker.ErrNuevaRondaSinRevelar},
		{"no participa", planningpoker.EstadoRevelada, "dani", divergente, planningpoker.ErrNoEsParticipante},
		{"todos votaron lo mismo", planningpoker.EstadoRevelada, "ana", map[string]int{"ana": 5, "beto": 5, "caro": 5}, planningpoker.ErrSinDivergencia},
		{"un solo voto", planningpoker.EstadoRevelada, "ana", map[string]int{"caro": 13}, planningpoker.ErrSinDivergencia},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			sesion := sesionEn(c.estado)

			nueva, err := planningpoker.AbrirNuevaRonda(sesion, c.integrante, votacionDe(sesion, c.votos))

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("AbrirNuevaRonda = %v, se esperaba %v", err, c.errEsperado)
			}
			if !reflect.DeepEqual(nueva, planningpoker.SesionPlanningPoker{}) {
				t.Errorf("con error devolvio una sesion: %+v", nueva)
			}
		})
	}
}

// RN-017-4, CL-017-7: con la sesion votando la divergencia no se evalua, porque los valores
// estan ocultos (RN-015-7): el unico error es el de la sesion sin revelar.
func TestAbrirNuevaRonda_VotandoNoEvaluaLaDivergencia(t *testing.T) {
	sesion := sesionEn(planningpoker.EstadoVotando)

	_, err := planningpoker.AbrirNuevaRonda(sesion, "ana", votacionDe(sesion, map[string]int{"ana": 5, "beto": 5}))

	if !errors.Is(err, planningpoker.ErrNuevaRondaSinRevelar) || errors.Is(err, planningpoker.ErrSinDivergencia) {
		t.Errorf("AbrirNuevaRonda = %v, se esperaba solo ErrNuevaRondaSinRevelar", err)
	}
}

// RN-017-7, CL-017-9: los errores se unen en orden fijo: sin revelar, no participa y sin
// divergencia.
func TestAbrirNuevaRonda_ErroresUnidosEnOrden(t *testing.T) {
	casos := []struct {
		nombre    string
		estado    planningpoker.Estado
		votos     map[string]int
		esperados []error
	}{
		{"votando y no participa", planningpoker.EstadoVotando, divergente,
			[]error{planningpoker.ErrNuevaRondaSinRevelar, planningpoker.ErrNoEsParticipante}},
		{"no participa y sin divergencia", planningpoker.EstadoRevelada, map[string]int{"ana": 5},
			[]error{planningpoker.ErrNoEsParticipante, planningpoker.ErrSinDivergencia}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			sesion := sesionEn(c.estado)

			_, err := planningpoker.AbrirNuevaRonda(sesion, "dani", votacionDe(sesion, c.votos))

			if err == nil {
				t.Fatal("AbrirNuevaRonda no devolvio error")
			}
			var mensajes []string
			for _, e := range c.esperados {
				if !errors.Is(err, e) {
					t.Errorf("falta %v en %v", e, err)
				}
				mensajes = append(mensajes, e.Error())
			}
			if err.Error() != strings.Join(mensajes, "\n") {
				t.Errorf("errores = %q, se esperaban en este orden: %q", err.Error(), mensajes)
			}
		})
	}
}

// US-017 / CA-017-2, RN-017-6, CL-017-6: los valores de una ronda anterior a la actual se leen
// aunque la sesion este votando; los de la ronda actual siguen ocultos.
func TestVotacionDeRonda(t *testing.T) {
	sesion := sesionEn(planningpoker.EstadoVotando)
	sesion.Ronda = 2

	anterior := planningpoker.NuevaVotacionDeRonda(sesion, 1, divergente)
	valores, err := anterior.Valores()
	if err != nil || !reflect.DeepEqual(valores, divergente) {
		t.Errorf("valores de la ronda 1 = (%v, %v), se esperaban %v", valores, err, divergente)
	}

	actual := planningpoker.NuevaVotacionDeRonda(sesion, 2, map[string]int{"ana": 13})
	if _, err := actual.Valores(); !errors.Is(err, planningpoker.ErrVotosOcultos) {
		t.Errorf("valores de la ronda actual votando: %v, se esperaba ErrVotosOcultos", err)
	}

	// Con la sesion revelada, la ronda actual tambien se lee.
	sesion.Estado = planningpoker.EstadoRevelada
	if _, err := planningpoker.NuevaVotacionDeRonda(sesion, 2, map[string]int{"ana": 13}).Valores(); err != nil {
		t.Errorf("valores de la ronda actual revelada: %v", err)
	}
}

// RN-017-10: la votacion de una ronda anterior tampoco cuenta el voto de quien no participa.
func TestVotacionDeRonda_SoloParticipantes(t *testing.T) {
	sesion := sesionEn(planningpoker.EstadoVotando)
	sesion.Ronda = 2

	v := planningpoker.NuevaVotacionDeRonda(sesion, 1, map[string]int{"ana": 3, "dani": 8})

	if !slices.Equal(v.Votaron(), []string{"ana"}) {
		t.Errorf("votaron %v, se esperaba solo ana", v.Votaron())
	}
}

// spec US-017, seccion 5: el resultado de una ronda se arma solo si sus valores se pueden leer.
func TestResultadoDeLaRonda(t *testing.T) {
	sesion := sesionEn(planningpoker.EstadoVotando)
	sesion.Ronda = 2

	r, err := planningpoker.ResultadoDeLaRonda(planningpoker.NuevaVotacionDeRonda(sesion, 1, divergente))
	if err != nil {
		t.Fatalf("ResultadoDeLaRonda de la ronda 1: %v", err)
	}
	esperado := planningpoker.Resultado{
		Votos:       []planningpoker.Voto{{ParticipanteID: "ana", Valor: 3}, {ParticipanteID: "beto", Valor: 8}},
		Faltaron:    []string{"caro"},
		Divergencia: true, Minimo: 3, Maximo: 8,
	}
	if !reflect.DeepEqual(r, esperado) {
		t.Errorf("ResultadoDeLaRonda = %+v, se esperaba %+v", r, esperado)
	}

	if _, err := planningpoker.ResultadoDeLaRonda(votacionDe(sesion, divergente)); !errors.Is(err, planningpoker.ErrVotosOcultos) {
		t.Errorf("ResultadoDeLaRonda de la ronda actual votando: %v, se esperaba ErrVotosOcultos", err)
	}
}

// Seccion 7 de la spec: los mensajes de los errores nuevos.
func TestErroresDeNuevaRonda_Mensajes(t *testing.T) {
	casos := map[error]string{
		planningpoker.ErrNuevaRondaSinRevelar: "la nueva ronda se abre con los votos revelados y la sesion sin finalizar",
		planningpoker.ErrSinDivergencia:       "no hay divergencia: registren la estimacion acordada",
	}
	for err, mensaje := range casos {
		if err.Error() != mensaje {
			t.Errorf("mensaje = %q, se esperaba %q", err.Error(), mensaje)
		}
	}
}
