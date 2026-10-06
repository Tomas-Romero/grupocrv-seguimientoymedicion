package metricas_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
)

// cerrado arma un sprint cerrado para los casos de la tabla.
func cerrado(nombre string, planificados, completados int) metricas.SprintCerrado {
	return metricas.SprintCerrado{
		Nombre:  nombre,
		Resumen: metricas.ResumenSprint{Planificados: planificados, Completados: completados},
	}
}

// US-026: la velocidad es el promedio de los Story Points completados de los
// sprints cerrados, redondeado a 1 decimal.
func TestCalcularVelocidad(t *testing.T) {
	casos := []struct {
		nombre   string
		sprints  []metricas.SprintCerrado
		esperado float64
	}{
		{
			// CA-026-1 (normal) / RN-026-1
			nombre: "varios sprints con distintos completados",
			sprints: []metricas.SprintCerrado{
				cerrado("Sprint 1", 25, 20),
				cerrado("Sprint 2", 30, 30),
				cerrado("Sprint 3", 30, 25),
			},
			esperado: 25,
		},
		{
			// CA-026-2 (alternativo) / CL-026-2
			nombre:   "un solo sprint cerrado",
			sprints:  []metricas.SprintCerrado{cerrado("Sprint 1", 20, 18)},
			esperado: 18,
		},
		{
			// CA-026-3 (limite) / CL-026-1
			nombre:   "ningun sprint cerrado",
			sprints:  nil,
			esperado: 0,
		},
		{
			// CL-026-3: hubo sprints, pero no se completo nada
			nombre: "sprints cerrados sin nada completado",
			sprints: []metricas.SprintCerrado{
				cerrado("Sprint 1", 10, 0),
				cerrado("Sprint 2", 12, 0),
			},
			esperado: 0,
		},
		{
			// CL-026-4 / RN-026-2: el promedio exacto es 31,5
			nombre: "promedio con un decimal exacto",
			sprints: []metricas.SprintCerrado{
				cerrado("Sprint 1", 30, 30),
				cerrado("Sprint 2", 33, 33),
			},
			esperado: 31.5,
		},
		{
			// CL-026-4 / RN-026-2: 17/3 = 5,666... se redondea hacia arriba
			nombre: "promedio que se redondea hacia arriba",
			sprints: []metricas.SprintCerrado{
				cerrado("Sprint 1", 5, 5),
				cerrado("Sprint 2", 6, 6),
				cerrado("Sprint 3", 6, 6),
			},
			esperado: 5.7,
		},
		{
			// CL-026-4 / RN-026-2: 16/3 = 5,333... se redondea hacia abajo
			nombre: "promedio que se redondea hacia abajo",
			sprints: []metricas.SprintCerrado{
				cerrado("Sprint 1", 5, 5),
				cerrado("Sprint 2", 5, 5),
				cerrado("Sprint 3", 6, 6),
			},
			esperado: 5.3,
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			obtenida, err := metricas.CalcularVelocidad(c.sprints)
			if err != nil {
				t.Fatalf("no se esperaba error: %v", err)
			}
			if obtenida != c.esperado {
				t.Fatalf("velocidad = %v, se esperaba %v", obtenida, c.esperado)
			}
		})
	}
}

// US-026 / CA-026-4 (error) / RN-026-3: un sprint con mas completados que
// planificados es un dato corrupto: ErrResumenInconsistente con el nombre del
// sprint y ninguna velocidad.
func TestCalcularVelocidad_ResumenInconsistente(t *testing.T) {
	sprints := []metricas.SprintCerrado{
		cerrado("Sprint 1", 20, 20),
		cerrado("Sprint 2", 5, 8),
	}

	velocidad, err := metricas.CalcularVelocidad(sprints)

	if !errors.Is(err, metricas.ErrResumenInconsistente) {
		t.Fatalf("error = %v, se esperaba ErrResumenInconsistente", err)
	}
	if !strings.Contains(err.Error(), `"Sprint 2"`) {
		t.Fatalf("el error %q no nombra al sprint inconsistente", err)
	}
	if velocidad != 0 {
		t.Fatalf("velocidad = %v, se esperaba 0 junto con el error", velocidad)
	}
}

// US-026 / seccion 7: un sprint sin nombre se rechaza con ErrNombreVacio y su
// posicion en la lista, contada desde 1.
func TestCalcularVelocidad_NombreVacio(t *testing.T) {
	sprints := []metricas.SprintCerrado{
		cerrado("Sprint 1", 20, 20),
		cerrado("   ", 10, 10),
	}

	velocidad, err := metricas.CalcularVelocidad(sprints)

	if !errors.Is(err, metricas.ErrNombreVacio) {
		t.Fatalf("error = %v, se esperaba ErrNombreVacio", err)
	}
	if !strings.Contains(err.Error(), "posicion 2") {
		t.Fatalf("el error %q no indica la posicion 2 (contada desde 1)", err)
	}
	if velocidad != 0 {
		t.Fatalf("velocidad = %v, se esperaba 0 junto con el error", velocidad)
	}
}

// US-026 / CL-026-5 / RN-026-4: con varios sprints invalidos se devuelven todos
// los errores unidos, en el orden de la lista.
func TestCalcularVelocidad_VariosErrores(t *testing.T) {
	sprints := []metricas.SprintCerrado{
		cerrado("", 10, 10),
		cerrado("Sprint 2", 20, 20),
		cerrado("Sprint 3", 5, 9),
	}

	_, err := metricas.CalcularVelocidad(sprints)

	if !errors.Is(err, metricas.ErrNombreVacio) || !errors.Is(err, metricas.ErrResumenInconsistente) {
		t.Fatalf("error = %v, se esperaban ErrNombreVacio y ErrResumenInconsistente unidos", err)
	}
	texto := err.Error()
	if strings.Index(texto, "posicion 1") > strings.Index(texto, `"Sprint 3"`) {
		t.Fatalf("los errores no estan en el orden de la lista: %q", texto)
	}
}
