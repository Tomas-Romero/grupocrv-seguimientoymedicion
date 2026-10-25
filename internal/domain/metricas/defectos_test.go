package metricas_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
)

// d arma un defecto registrado: detectado en un sprint y resuelto en otro (0 = sin resolver).
func d(deteccion, resolucion int) metricas.DefectoRegistrado {
	return metricas.DefectoRegistrado{SprintDeteccion: deteccion, SprintResolucion: resolucion}
}

// US-029 (CA-029-1, CA-029-2, CA-029-3 y CA-029-4): los detectados se cuentan por el sprint de
// deteccion y los resueltos por el de resolucion (RN-029-1 a RN-029-5).
func TestCalcularDefectosDelSprint(t *testing.T) {
	casos := []struct {
		nombre     string
		sprint     int
		defectos   []metricas.DefectoRegistrado
		detectados int
		resueltos  int
	}{
		{
			// CA-029-1 (normal) / RN-029-1, RN-029-2
			nombre: "tres detectados y dos resueltos en el sprint",
			sprint: 3, defectos: []metricas.DefectoRegistrado{d(3, 3), d(3, 3), d(3, 0)},
			detectados: 3, resueltos: 2,
		},
		{
			// CA-029-2 / CL-029-3: detectado en el 2, resuelto en el 4; se cuenta en cada uno
			nombre: "detectado en el sprint 2 y resuelto en el 4: se consulta el 2",
			sprint: 2, defectos: []metricas.DefectoRegistrado{d(2, 4)},
			detectados: 1, resueltos: 0,
		},
		{
			nombre: "detectado en el sprint 2 y resuelto en el 4: se consulta el 4",
			sprint: 4, defectos: []metricas.DefectoRegistrado{d(2, 4)},
			detectados: 0, resueltos: 1,
		},
		{
			nombre: "detectado en el sprint 2 y resuelto en el 4: se consulta el 3",
			sprint: 3, defectos: []metricas.DefectoRegistrado{d(2, 4)},
			detectados: 0, resueltos: 0,
		},
		{
			// CA-029-3 / CL-029-2 / RN-029-3: detectado y resuelto en el mismo sprint
			nombre: "detectado y resuelto en el mismo sprint",
			sprint: 5, defectos: []metricas.DefectoRegistrado{d(5, 5)},
			detectados: 1, resueltos: 1,
		},
		{
			// CA-029-4 (limite) / CL-029-4 / RN-029-4: sin resolver nunca cuenta como resuelto
			nombre: "defecto sin resolver: su sprint",
			sprint: 1, defectos: []metricas.DefectoRegistrado{d(1, 0)},
			detectados: 1, resueltos: 0,
		},
		{
			nombre: "defecto sin resolver: otro sprint",
			sprint: 2, defectos: []metricas.DefectoRegistrado{d(1, 0)},
			detectados: 0, resueltos: 0,
		},
		{
			// CA-029-4 / CL-029-1 / RN-029-5
			nombre: "lista vacia",
			sprint: 1, defectos: []metricas.DefectoRegistrado{},
			detectados: 0, resueltos: 0,
		},
		{
			// CL-029-12: nil es lo mismo que vacia
			nombre: "lista nil",
			sprint: 1, defectos: nil,
			detectados: 0, resueltos: 0,
		},
		{
			// CL-029-5: los de otros sprints no cuentan, ni los de sprints posteriores
			nombre: "defectos de otros sprints",
			sprint: 2, defectos: []metricas.DefectoRegistrado{d(1, 1), d(2, 2), d(3, 0), d(4, 4)},
			detectados: 1, resueltos: 1,
		},
		{
			// CL-029-6: se suman uno por uno
			nombre: "varios defectos en el mismo sprint",
			sprint: 2, defectos: []metricas.DefectoRegistrado{d(2, 0), d(2, 0), d(2, 0), d(2, 0)},
			detectados: 4, resueltos: 0,
		},
		{
			// CL-029-7: un sprint que corrige defectos de sprints anteriores
			nombre: "mas resueltos que detectados",
			sprint: 4, defectos: []metricas.DefectoRegistrado{d(4, 0), d(1, 4), d(2, 4), d(3, 4)},
			detectados: 1, resueltos: 3,
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			resumen, err := metricas.CalcularDefectosDelSprint(c.sprint, c.defectos)
			if err != nil {
				t.Fatalf("CalcularDefectosDelSprint(%d, %v): error inesperado: %v", c.sprint, c.defectos, err)
			}
			if resumen.Detectados != c.detectados || resumen.Resueltos != c.resueltos {
				t.Errorf("CalcularDefectosDelSprint(%d, %v) = %+v, quiero %d detectados y %d resueltos",
					c.sprint, c.defectos, resumen, c.detectados, c.resueltos)
			}
		})
	}
}

// US-029 (CA-029-5 y CA-029-6): los datos invalidos se rechazan, todos juntos y sin resumen (RN-029-6 a RN-029-8).
func TestCalcularDefectosDelSprint_Rechazos(t *testing.T) {
	casos := []struct {
		nombre   string
		sprint   int
		defectos []metricas.DefectoRegistrado
		quiero   []error
		mensajes []string
	}{
		{
			// CA-029-5 (error) / CL-029-8 / RN-029-7
			nombre: "resuelto antes de detectarse",
			sprint: 3, defectos: []metricas.DefectoRegistrado{d(3, 3), d(3, 2)},
			quiero:   []error{metricas.ErrResolucionAnteriorADeteccion},
			mensajes: []string{"defecto en la posicion 2"},
		},
		{
			// una resolucion negativa tambien es anterior a la deteccion
			nombre: "resolucion negativa",
			sprint: 3, defectos: []metricas.DefectoRegistrado{d(3, -1)},
			quiero:   []error{metricas.ErrResolucionAnteriorADeteccion},
			mensajes: []string{"defecto en la posicion 1"},
		},
		{
			// CL-029-9
			nombre: "sprint de deteccion en cero o negativo",
			sprint: 3, defectos: []metricas.DefectoRegistrado{d(0, 0), d(2, 0), d(-1, 0)},
			quiero:   []error{metricas.ErrSprintInvalido},
			mensajes: []string{"defecto en la posicion 1", "defecto en la posicion 3"},
		},
		{
			// CA-029-6 (error) / CL-029-10 / RN-029-6: aunque la lista este vacia
			nombre: "sprint consultado en cero",
			sprint: 0, defectos: nil,
			quiero: []error{metricas.ErrSprintInvalido},
		},
		{
			nombre: "sprint consultado negativo",
			sprint: -2, defectos: []metricas.DefectoRegistrado{d(1, 1)},
			quiero: []error{metricas.ErrSprintInvalido},
		},
		{
			// CL-029-11 / RN-029-8: el sprint consultado primero y despues cada defecto por posicion
			nombre: "sprint consultado y dos defectos invalidos",
			sprint: 0, defectos: []metricas.DefectoRegistrado{d(2, 1), d(0, 0)},
			quiero:   []error{metricas.ErrSprintInvalido, metricas.ErrResolucionAnteriorADeteccion},
			mensajes: []string{"sprint consultado 0", "defecto en la posicion 1", "defecto en la posicion 2"},
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			resumen, err := metricas.CalcularDefectosDelSprint(c.sprint, c.defectos)

			if err == nil {
				t.Fatalf("CalcularDefectosDelSprint(%d, %v) no devolvio error", c.sprint, c.defectos)
			}
			if resumen != (metricas.ResumenDefectos{}) {
				t.Errorf("con error el resumen tiene que estar vacio, devolvio %+v", resumen)
			}
			for _, q := range c.quiero {
				if !errors.Is(err, q) {
					t.Errorf("el error %q no es %v", err, q)
				}
			}
			// RN-029-8: los mensajes salen en el orden de la lista.
			desde := 0
			for _, m := range c.mensajes {
				i := strings.Index(err.Error()[desde:], m)
				if i < 0 {
					t.Fatalf("el error %q no tiene %q en el orden %q", err, m, c.mensajes)
				}
				desde += i + len(m)
			}
		})
	}
}

// US-029 / seccion 7: los mensajes de los errores nuevos.
func TestErroresDeDefectos_Mensajes(t *testing.T) {
	if got := metricas.ErrSprintInvalido.Error(); got != "el numero de sprint tiene que ser 1 o mayor" {
		t.Errorf("ErrSprintInvalido = %q", got)
	}
	if got := metricas.ErrResolucionAnteriorADeteccion.Error(); got != "el defecto se resolvio antes de detectarse" {
		t.Errorf("ErrResolucionAnteriorADeteccion = %q", got)
	}
}
