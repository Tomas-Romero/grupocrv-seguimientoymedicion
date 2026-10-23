package metricas_test

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
)

// US-027 (CA-027-1, CA-027-2, CA-027-3 y CA-027-4): las horas estimadas salen de los
// Story Points planificados (1,5 horas por punto), las reales son la suma de las horas
// registradas y la desviacion es la diferencia entre ambas.
func TestCalcularHoras(t *testing.T) {
	casos := []struct {
		nombre     string
		puntos     int
		horas      []float64
		estimadas  float64
		reales     float64
		desviacion float64
		porcentaje float64
		calculable bool
	}{
		{
			// CA-027-1 (normal) / CL-027-5 / RN-027-1 a RN-027-4: 6,67 se redondea a 6,7, no se trunca
			nombre: "5 story points y 8 horas",
			puntos: 5, horas: []float64{3, 5},
			estimadas: 7.5, reales: 8, desviacion: 0.5, porcentaje: 6.7, calculable: true,
		},
		{
			// CA-027-2 (alternativo)
			nombre: "8 story points y 9 horas en tres registros",
			puntos: 8, horas: []float64{2, 3, 4},
			estimadas: 12, reales: 9, desviacion: -3, porcentaje: -25, calculable: true,
		},
		{
			// CA-027-3 (limite) / CL-027-1: nadie cargo horas
			nombre: "sin horas registradas",
			puntos: 8, horas: nil,
			estimadas: 12, reales: 0, desviacion: -12, porcentaje: -100, calculable: true,
		},
		{
			// CL-027-1: una lista vacia es lo mismo que nil
			nombre: "lista vacia de horas",
			puntos: 8, horas: []float64{},
			estimadas: 12, reales: 0, desviacion: -12, porcentaje: -100, calculable: true,
		},
		{
			// CA-027-4 (limite) / CL-027-2 / RN-027-4: sin base de comparacion
			nombre: "sin story points y con horas",
			puntos: 0, horas: []float64{4},
			estimadas: 0, reales: 4, desviacion: 4, calculable: false,
		},
		{
			// CL-027-3
			nombre: "sin nada",
			puntos: 0, horas: nil,
			estimadas: 0, reales: 0, desviacion: 0, calculable: false,
		},
		{
			// CL-027-4: coinciden
			nombre: "horas reales iguales a las estimadas",
			puntos: 5, horas: []float64{7.5},
			estimadas: 7.5, reales: 7.5, desviacion: 0, porcentaje: 0, calculable: true,
		},
		{
			// CL-027-6: el minimo de US-019
			nombre: "un solo registro de un cuarto de hora",
			puntos: 1, horas: []float64{0.25},
			estimadas: 1.5, reales: 0.25, desviacion: -1.25, porcentaje: -83.3, calculable: true,
		},
		{
			// CL-027-7 / RN: la suma de decimales no deja ruido de coma flotante
			nombre: "suma de decimales exacta",
			puntos: 1, horas: []float64{0.25, 0.25, 0.5},
			estimadas: 1.5, reales: 1, desviacion: -0.5, porcentaje: -33.3, calculable: true,
		},
		{
			// RN-027-2: no se filtra nada, y las cifras se redondean a 2 decimales
			nombre: "ruido de coma flotante",
			puntos: 5, horas: []float64{0.1, 0.2, 7.2},
			estimadas: 7.5, reales: 7.5, desviacion: 0, porcentaje: 0, calculable: true,
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			resumen, err := metricas.CalcularHoras(c.puntos, c.horas)
			if err != nil {
				t.Fatalf("CalcularHoras(%d, %v): error inesperado: %v", c.puntos, c.horas, err)
			}
			if resumen.Estimadas != c.estimadas || resumen.Reales != c.reales || resumen.Desviacion != c.desviacion {
				t.Errorf("CalcularHoras(%d, %v) = %+v, quiero estimadas %v, reales %v, desviacion %v",
					c.puntos, c.horas, resumen, c.estimadas, c.reales, c.desviacion)
			}
			porcentaje, calculable := resumen.DesviacionPorcentual()
			if calculable != c.calculable {
				t.Fatalf("DesviacionPorcentual: calculable = %v, quiero %v", calculable, c.calculable)
			}
			if calculable && math.Abs(porcentaje-c.porcentaje) > 1e-9 {
				t.Errorf("DesviacionPorcentual = %v, quiero %v", porcentaje, c.porcentaje)
			}
			if !calculable && porcentaje != 0 {
				t.Errorf("DesviacionPorcentual no calculable devolvio %v, quiero 0", porcentaje)
			}
		})
	}
}

// RN-027-1: la equivalencia es una constante del dominio que la pantalla puede mostrar.
func TestHorasPorStoryPoint(t *testing.T) {
	if metricas.HorasPorStoryPoint != 1.5 {
		t.Errorf("HorasPorStoryPoint = %v, quiero 1,5 (acta del Sprint 0)", metricas.HorasPorStoryPoint)
	}
}

// US-027 (CA-027-5 y CA-027-6): los datos invalidos se rechazan, todos juntos y sin resumen.
func TestCalcularHoras_Rechazos(t *testing.T) {
	casos := []struct {
		nombre   string
		puntos   int
		horas    []float64
		quiero   []error
		mensajes []string
	}{
		{
			// CA-027-5 (error) / RN-027-6 / CL-027-9
			nombre: "una hora en cero",
			puntos: 5, horas: []float64{2, 0, 3},
			quiero:   []error{metricas.ErrHorasInvalidas},
			mensajes: []string{"la hora de la posicion 2 no es un numero mayor a 0"},
		},
		{
			// CA-027-5 / CL-027-9
			nombre: "una hora negativa",
			puntos: 5, horas: []float64{-1},
			quiero:   []error{metricas.ErrHorasInvalidas},
			mensajes: []string{"la hora de la posicion 1 no es un numero mayor a 0"},
		},
		{
			// RN-027-6: NaN e infinito no son horas
			nombre: "NaN e infinito",
			puntos: 5, horas: []float64{math.NaN(), 2, math.Inf(1)},
			quiero:   []error{metricas.ErrHorasInvalidas},
			mensajes: []string{"la hora de la posicion 1 ", "la hora de la posicion 3 "},
		},
		{
			// CA-027-6 (error) / RN-027-5
			nombre: "story points negativos",
			puntos: -1, horas: []float64{2},
			quiero: []error{metricas.ErrStoryPointsNegativos},
		},
		{
			// CL-027-8 / RN-027-7: los tres errores, el de los story points primero
			nombre: "story points negativos y dos horas invalidas",
			puntos: -3, horas: []float64{0, 4, -2},
			quiero:   []error{metricas.ErrStoryPointsNegativos, metricas.ErrHorasInvalidas},
			mensajes: []string{"la hora de la posicion 1 ", "la hora de la posicion 3 "},
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			resumen, err := metricas.CalcularHoras(c.puntos, c.horas)
			if err == nil {
				t.Fatalf("CalcularHoras(%d, %v) no devolvio error", c.puntos, c.horas)
			}
			if resumen != (metricas.ResumenHoras{}) {
				t.Errorf("con error el resumen tiene que estar vacio, devolvio %+v", resumen)
			}
			for _, q := range c.quiero {
				if !errors.Is(err, q) {
					t.Errorf("el error %q no es %v", err, q)
				}
			}
			for _, m := range c.mensajes {
				if !strings.Contains(err.Error(), m) {
					t.Errorf("el error %q no contiene %q", err, m)
				}
			}
		})
	}

	// RN-027-7: el orden es fijo, los story points primero y despues las horas por posicion.
	_, err := metricas.CalcularHoras(-3, []float64{0, 4, -2})
	texto := err.Error()
	orden := []string{"story points", "posicion 1 ", "posicion 3 "}
	desde := 0
	for _, parte := range orden {
		i := strings.Index(texto[desde:], parte)
		if i < 0 {
			t.Fatalf("el error %q no tiene %q en el orden %v", texto, parte, orden)
		}
		desde += i + len(parte)
	}
}
