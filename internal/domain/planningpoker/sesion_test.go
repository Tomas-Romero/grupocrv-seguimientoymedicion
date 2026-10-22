package planningpoker_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// historiaLista es una historia que se puede poner en sesion: sin completar y sin
// sesion abierta.
func historiaLista() planningpoker.DatosDeLaHistoria {
	return planningpoker.DatosDeLaHistoria{}
}

var tresParticipantes = []string{"ana", "beto", "caro"}

// US-014 / CA-014-1, RN-014-4, RN-014-6: la sesion se crea votando, en la ronda 1, sobre
// la historia y con los participantes recibidos, en el mismo orden. No tiene ID: lo
// genera la base (RN-014-12).
func TestNuevaSesion(t *testing.T) {
	sesion, err := planningpoker.NuevaSesion("item-1", historiaLista(), tresParticipantes)
	if err != nil {
		t.Fatalf("NuevaSesion: %v", err)
	}

	if sesion.ItemID != "item-1" || sesion.Estado != planningpoker.EstadoVotando || sesion.Ronda != 1 {
		t.Errorf("sesion = %+v, se esperaba la del item 1, votando y en la ronda 1", sesion)
	}
	if !slices.Equal(sesion.Participantes, tresParticipantes) {
		t.Errorf("participantes = %q, se esperaban %q", sesion.Participantes, tresParticipantes)
	}
	if sesion.ID != "" {
		t.Errorf("ID = %q, se esperaba vacio", sesion.ID)
	}
}

// La sesion tiene su propia lista de participantes: cambiar la recibida no la cambia.
func TestNuevaSesion_CopiaLosParticipantes(t *testing.T) {
	participantes := slices.Clone(tresParticipantes)
	sesion, err := planningpoker.NuevaSesion("item-1", historiaLista(), participantes)
	if err != nil {
		t.Fatalf("NuevaSesion: %v", err)
	}

	participantes[0] = "cambiado"

	if sesion.Participantes[0] != "ana" {
		t.Error("la sesion comparte la lista de participantes con quien la armo")
	}
}

// US-014 / CA-014-4, RN-014-5, CL-014-1, CL-014-2: hacen falta al menos 2 participantes.
func TestNuevaSesion_MinimoDeParticipantes(t *testing.T) {
	casos := []struct {
		nombre        string
		participantes []string
		valida        bool
	}{
		{"dos", []string{"ana", "beto"}, true},
		{"uno", []string{"ana"}, false},
		{"ninguno", nil, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := planningpoker.NuevaSesion("item-1", historiaLista(), c.participantes)
			if c.valida && err != nil {
				t.Errorf("error = %v, se esperaba una sesion valida", err)
			}
			if !c.valida && !errors.Is(err, planningpoker.ErrParticipantesInsuficientes) {
				t.Errorf("error = %v, se esperaba ErrParticipantesInsuficientes", err)
			}
		})
	}
}

// US-014 / RN-014-3, CL-014-3: una historia ya estimada se puede poner en sesion. El
// dominio no la distingue: no recibe la estimacion.
//
// US-014 / CA-014-5, CA-014-6, RN-014-2, RN-014-7: una historia completada o con una
// sesion abierta se rechaza.
func TestNuevaSesion_Rechazos(t *testing.T) {
	casos := []struct {
		nombre      string
		historia    planningpoker.DatosDeLaHistoria
		errEsperado error
	}{
		{"historia completada", planningpoker.DatosDeLaHistoria{Completada: true}, backlog.ErrItemCompletado},
		{"sesion abierta", planningpoker.DatosDeLaHistoria{ConSesionAbierta: true}, planningpoker.ErrSesionAbierta},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			sesion, err := planningpoker.NuevaSesion("item-1", c.historia, tresParticipantes)

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if sesion.ItemID != "" {
				t.Errorf("devolvio la sesion %+v, se esperaba el valor cero", sesion)
			}
		})
	}
}

// US-014 / RN-014-10, CL-014-9: el dominio valida todo y devuelve los errores unidos, en
// orden fijo: historia completada, sesion abierta, participantes insuficientes.
func TestNuevaSesion_ErroresUnidosEnOrden(t *testing.T) {
	historia := planningpoker.DatosDeLaHistoria{Completada: true, ConSesionAbierta: true}

	_, err := planningpoker.NuevaSesion("item-1", historia, []string{"ana"})

	esperados := []error{backlog.ErrItemCompletado, planningpoker.ErrSesionAbierta, planningpoker.ErrParticipantesInsuficientes}
	posicion := -1
	for _, esperado := range esperados {
		if !errors.Is(err, esperado) {
			t.Fatalf("error = %v, se esperaba que incluya %v", err, esperado)
		}
		i := strings.Index(err.Error(), esperado.Error())
		if i < posicion {
			t.Errorf("el error %q no respeta el orden de RN-014-10", err)
		}
		posicion = i
	}
}

// US-014 / RN-014-7, RN-014-8: una sesion esta abierta mientras no esta finalizada.
func TestEstado_Abierta(t *testing.T) {
	for estado, abierta := range map[planningpoker.Estado]bool{
		planningpoker.EstadoVotando:    true,
		planningpoker.EstadoRevelada:   true,
		planningpoker.EstadoFinalizada: false,
	} {
		if estado.Abierta() != abierta {
			t.Errorf("%q.Abierta() = %t, se esperaba %t", estado, estado.Abierta(), abierta)
		}
	}
}

// US-014 / RN-014-8: los estados son los de la tabla de la migracion.
func TestEstado_Valores(t *testing.T) {
	for estado, esperado := range map[planningpoker.Estado]string{
		planningpoker.EstadoVotando:    "votando",
		planningpoker.EstadoRevelada:   "revelada",
		planningpoker.EstadoFinalizada: "finalizada",
	} {
		if string(estado) != esperado {
			t.Errorf("estado = %q, se esperaba %q", estado, esperado)
		}
	}
}

// US-014 / spec seccion 7: los mensajes que ve la persona.
func TestErrores_Mensajes(t *testing.T) {
	for err, esperado := range map[error]string{
		planningpoker.ErrSesionAbierta:              "la historia ya tiene una sesion de planning poker abierta",
		planningpoker.ErrParticipantesInsuficientes: "se necesitan al menos 2 integrantes en el proyecto para una sesion de planning poker",
	} {
		if err.Error() != esperado {
			t.Errorf("mensaje = %q, se esperaba %q", err.Error(), esperado)
		}
	}
}
