package defecto_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// US-024 / CA-024-4 / RN-024-8 / CL-024-16: ValidarSeveridad es la regla de RN-022-2 escrita una sola
// vez. Acepta exactamente las cuatro severidades y rechaza cualquier otro valor, incluida la vacia,
// sin pasarlo a minuscula ni recortarlo, con ErrSeveridadInvalida envuelto con el valor recibido.
func TestValidarSeveridad(t *testing.T) {
	casos := []struct {
		nombre    string
		severidad defecto.Severidad
		valida    bool
	}{
		{"critica", defecto.SeveridadCritica, true},
		{"alta", defecto.SeveridadAlta, true},
		{"media", defecto.SeveridadMedia, true},
		{"baja", defecto.SeveridadBaja, true},
		{"con la primera en mayuscula", "Alta", false},
		{"toda en mayuscula", "CRITICA", false},
		{"con tilde", "crítica", false},
		{"con espacios alrededor", " alta ", false},
		{"un solo espacio", " ", false},
		{"vacia", "", false},
		{"fuera de las cuatro", "urgente", false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := defecto.ValidarSeveridad(c.severidad)
			if c.valida {
				if err != nil {
					t.Fatalf("ValidarSeveridad(%q) = %v, quiero nil", c.severidad, err)
				}
				return
			}
			if !errors.Is(err, defecto.ErrSeveridadInvalida) {
				t.Fatalf("ValidarSeveridad(%q) = %v, quiero ErrSeveridadInvalida", c.severidad, err)
			}
			if esperado := `se recibio "` + string(c.severidad) + `"`; !strings.Contains(err.Error(), esperado) {
				t.Errorf("el error = %q, quiero que incluya %s", err.Error(), esperado)
			}
		})
	}
}
