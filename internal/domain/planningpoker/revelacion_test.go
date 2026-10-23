package planningpoker_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// US-016 (CA-016-1, CA-016-2, CA-016-3 y CA-016-4): al revelar salen los votos de quienes
// votaron, en el orden de los participantes; los que no votaron figuran aparte, y hay
// divergencia si los valores emitidos no son todos iguales (RN-016-5 y RN-016-6).
func TestRevelar(t *testing.T) {
	casos := []struct {
		nombre      string
		votos       map[string]int
		esperados   []planningpoker.Voto
		faltaron    []string
		divergencia bool
		minimo      int
		maximo      int
	}{
		{
			// CA-016-1 / CL-016-1
			nombre:    "todos votaron lo mismo",
			votos:     map[string]int{"ana": 5, "beto": 5, "caro": 5},
			esperados: []planningpoker.Voto{{ParticipanteID: "ana", Valor: 5}, {ParticipanteID: "beto", Valor: 5}, {ParticipanteID: "caro", Valor: 5}},
			faltaron:  []string{}, divergencia: false, minimo: 5, maximo: 5,
		},
		{
			// CA-016-2 / CL-016-2
			nombre:    "valores distintos",
			votos:     map[string]int{"caro": 8, "ana": 3, "beto": 5},
			esperados: []planningpoker.Voto{{ParticipanteID: "ana", Valor: 3}, {ParticipanteID: "beto", Valor: 5}, {ParticipanteID: "caro", Valor: 8}},
			faltaron:  []string{}, divergencia: true, minimo: 3, maximo: 8,
		},
		{
			// CA-016-4 / CL-016-3 / RN-016-6: valores contiguos de la escala
			nombre:    "valores contiguos",
			votos:     map[string]int{"ana": 5, "beto": 8, "caro": 5},
			esperados: []planningpoker.Voto{{ParticipanteID: "ana", Valor: 5}, {ParticipanteID: "beto", Valor: 8}, {ParticipanteID: "caro", Valor: 5}},
			faltaron:  []string{}, divergencia: true, minimo: 5, maximo: 8,
		},
		{
			// CA-016-3 / CL-016-5 / RN-016-5: faltan votos pero los emitidos coinciden
			nombre:    "faltan votos y los emitidos coinciden",
			votos:     map[string]int{"ana": 5, "beto": 5},
			esperados: []planningpoker.Voto{{ParticipanteID: "ana", Valor: 5}, {ParticipanteID: "beto", Valor: 5}},
			faltaron:  []string{"caro"}, divergencia: false, minimo: 5, maximo: 5,
		},
		{
			// CA-016-4 / CL-016-4: un solo voto emitido
			nombre:    "un solo voto",
			votos:     map[string]int{"beto": 13},
			esperados: []planningpoker.Voto{{ParticipanteID: "beto", Valor: 13}},
			faltaron:  []string{"ana", "caro"}, divergencia: false, minimo: 13, maximo: 13,
		},
		{
			// RN-016-6: el voto de quien no participa no cuenta, ni para la divergencia
			nombre:    "voto de alguien que no participa",
			votos:     map[string]int{"ana": 5, "dani": 13},
			esperados: []planningpoker.Voto{{ParticipanteID: "ana", Valor: 5}},
			faltaron:  []string{"beto", "caro"}, divergencia: false, minimo: 5, maximo: 5,
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			s := sesionEn(planningpoker.EstadoVotando)
			s.Ronda = 2

			r, err := planningpoker.Revelar(s, "ana", c.votos)
			if err != nil {
				t.Fatalf("Revelar: error inesperado: %v", err)
			}
			if r.SesionID != "sesion-1" || r.Ronda != 2 {
				t.Errorf("sesion %q ronda %d, quiero sesion-1 y la ronda 2 (RN-016-7)", r.SesionID, r.Ronda)
			}
			if !reflect.DeepEqual(r.Votos, c.esperados) {
				t.Errorf("Votos = %+v, quiero %+v", r.Votos, c.esperados)
			}
			if !reflect.DeepEqual(r.Faltaron, c.faltaron) {
				t.Errorf("Faltaron = %q, quiero %q", r.Faltaron, c.faltaron)
			}
			if r.Divergencia != c.divergencia || r.Minimo != c.minimo || r.Maximo != c.maximo {
				t.Errorf("divergencia %v, minimo %d, maximo %d; quiero %v, %d, %d", r.Divergencia, r.Minimo, r.Maximo, c.divergencia, c.minimo, c.maximo)
			}
		})
	}
}

// US-016 / RN-016-11: revelar no modifica los votos que recibe.
func TestRevelar_NoModificaLosVotos(t *testing.T) {
	votos := map[string]int{"ana": 5, "beto": 8}

	if _, err := planningpoker.Revelar(sesionEn(planningpoker.EstadoVotando), "ana", votos); err != nil {
		t.Fatalf("Revelar: %v", err)
	}

	if !reflect.DeepEqual(votos, map[string]int{"ana": 5, "beto": 8}) {
		t.Errorf("votos = %v: revelar los cambio", votos)
	}
}

// US-016 (CA-016-5, CA-016-6 y CA-016-7): se rechaza sin revelacion.
func TestRevelar_Rechazos(t *testing.T) {
	casos := []struct {
		nombre string
		estado planningpoker.Estado
		quien  string
		votos  map[string]int
		quiero []error
	}{
		// CA-016-5 / CL-016-6 / RN-016-4
		{"nadie voto", planningpoker.EstadoVotando, "ana", nil, []error{planningpoker.ErrSinVotos}},
		// solo votos de alguien que no participa: tampoco hay votos
		{"solo vota quien no participa", planningpoker.EstadoVotando, "ana", map[string]int{"dani": 5}, []error{planningpoker.ErrSinVotos}},
		// CA-016-6 / CL-016-12 / RN-016-3
		{"quien revela no participa", planningpoker.EstadoVotando, "dani", map[string]int{"ana": 5}, []error{planningpoker.ErrNoEsParticipante}},
		// CA-016-7 / RN-016-2: revelar dos veces no es idempotente
		{"ya esta revelada", planningpoker.EstadoRevelada, "ana", map[string]int{"ana": 5}, []error{planningpoker.ErrVotacionCerrada}},
		{"esta finalizada", planningpoker.EstadoFinalizada, "ana", map[string]int{"ana": 5}, []error{planningpoker.ErrVotacionCerrada}},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r, err := planningpoker.Revelar(sesionEn(c.estado), c.quien, c.votos)

			if err == nil {
				t.Fatal("Revelar no devolvio error")
			}
			for _, q := range c.quiero {
				if !errors.Is(err, q) {
					t.Errorf("el error %q no es %v", err, q)
				}
			}
			if !reflect.DeepEqual(r, planningpoker.Revelacion{}) {
				t.Errorf("con error la revelacion tiene que estar vacia, devolvio %+v", r)
			}
		})
	}
}

// US-016 / CL-016-8, RN-016-10: se devuelven todos los errores, unidos y en un orden fijo:
// votacion cerrada, no es participante y sin votos. Con datos reales no se da (una sesion
// revelada ya tuvo votos), pero la funcion es pura y se prueba con ese caso.
func TestRevelar_ErroresUnidosEnOrden(t *testing.T) {
	_, err := planningpoker.Revelar(sesionEn(planningpoker.EstadoFinalizada), "dani", nil)

	for _, esperado := range []error{planningpoker.ErrVotacionCerrada, planningpoker.ErrNoEsParticipante, planningpoker.ErrSinVotos} {
		if !errors.Is(err, esperado) {
			t.Fatalf("el error %q no incluye %v", err, esperado)
		}
	}
	texto := err.Error()
	desde := 0
	for _, parte := range []string{planningpoker.ErrVotacionCerrada.Error(), planningpoker.ErrNoEsParticipante.Error(), planningpoker.ErrSinVotos.Error()} {
		i := strings.Index(texto[desde:], parte)
		if i < 0 {
			t.Fatalf("el error %q no sigue el orden de RN-016-10", texto)
		}
		desde += i + len(parte)
	}
}

// US-016 / seccion 7: el mensaje del error nuevo.
func TestErrSinVotos_Mensaje(t *testing.T) {
	if got := planningpoker.ErrSinVotos.Error(); got != "no hay votos para revelar en esta ronda" {
		t.Errorf("ErrSinVotos = %q", got)
	}
}
