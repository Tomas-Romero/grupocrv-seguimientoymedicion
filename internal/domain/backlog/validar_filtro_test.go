package backlog_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// US-007 / CA-007-5 / RN-007-8 / CL-007-15: ValidarPrioridad es la regla de RN-005-5 escrita una sola
// vez. Acepta exactamente las cuatro prioridades y rechaza cualquier otro valor, incluida la vacia,
// sin pasarlo a minuscula ni recortarlo, con ErrPrioridadInvalida envuelto con el valor recibido.
func TestValidarPrioridad(t *testing.T) {
	casos := []struct {
		nombre    string
		prioridad backlog.Prioridad
		valida    bool
	}{
		{"must", backlog.PrioridadMust, true},
		{"should", backlog.PrioridadShould, true},
		{"could", backlog.PrioridadCould, true},
		{"wont", backlog.PrioridadWont, true},
		{"con la primera en mayuscula", "Must", false},
		{"toda en mayuscula", "MUST", false},
		{"con espacios alrededor", " must ", false},
		{"un solo espacio", " ", false},
		{"vacia", "", false},
		{"fuera de las cuatro", "urgente", false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := backlog.ValidarPrioridad(c.prioridad)
			if c.valida {
				if err != nil {
					t.Fatalf("ValidarPrioridad(%q) = %v, quiero nil", c.prioridad, err)
				}
				return
			}
			if !errors.Is(err, backlog.ErrPrioridadInvalida) {
				t.Fatalf("ValidarPrioridad(%q) = %v, quiero ErrPrioridadInvalida", c.prioridad, err)
			}
			if esperado := `se recibio "` + string(c.prioridad) + `"`; !strings.Contains(err.Error(), esperado) {
				t.Errorf("el error = %q, quiero que incluya %s", err.Error(), esperado)
			}
		})
	}
}

// US-007 / CA-007-5 / RN-007-7 / CL-007-14: ValidarEstado acepta exactamente los tres estados y rechaza
// cualquier otro valor, incluido el vacio, sin normalizarlo, con ErrEstadoInvalido envuelto con el valor
// recibido. Que en el filtro el estado vacio signifique "sin filtro" lo decide el caso de uso.
func TestValidarEstado(t *testing.T) {
	casos := []struct {
		nombre string
		estado backlog.Estado
		valido bool
	}{
		{"pendiente", backlog.EstadoPendiente, true},
		{"en progreso", backlog.EstadoEnProgreso, true},
		{"completado", backlog.EstadoCompletado, true},
		{"con la primera en mayuscula", "Pendiente", false},
		{"todo en mayuscula", "COMPLETADO", false},
		{"con espacio en vez de guion bajo", "en progreso", false},
		{"con espacios alrededor", " pendiente ", false},
		{"un solo espacio", " ", false},
		{"vacio", "", false},
		{"fuera de los tres", "terminado", false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := backlog.ValidarEstado(c.estado)
			if c.valido {
				if err != nil {
					t.Fatalf("ValidarEstado(%q) = %v, quiero nil", c.estado, err)
				}
				return
			}
			if !errors.Is(err, backlog.ErrEstadoInvalido) {
				t.Fatalf("ValidarEstado(%q) = %v, quiero ErrEstadoInvalido", c.estado, err)
			}
			if esperado := `se recibio "` + string(c.estado) + `"`; !strings.Contains(err.Error(), esperado) {
				t.Errorf("el error = %q, quiero que incluya %s", err.Error(), esperado)
			}
		})
	}
}

// US-007 / CA-007-5 / spec seccion 7: el mensaje de ErrEstadoInvalido es el que ve la persona.
func TestErrEstadoInvalido_Mensaje(t *testing.T) {
	quiero := "el estado tiene que ser pendiente, en_progreso o completado"
	if got := backlog.ErrEstadoInvalido.Error(); got != quiero {
		t.Errorf("ErrEstadoInvalido = %q, quiero %q", got, quiero)
	}
}
