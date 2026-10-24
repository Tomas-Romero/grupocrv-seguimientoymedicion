package metricas_test

import (
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
)

// US-028: el porcentaje de historias completadas es la cantidad de historias con
// Completado = true sobre el total, por 100.
func TestCalcularPorcentajeCompletadas(t *testing.T) {
	casos := []struct {
		nombre   string
		items    []metricas.ItemDelSprint
		esperado float64
	}{
		{
			// CA-028-3 (limite) / CL-028-1 / RN-028-4: sin historias no hay error, es 0
			nombre:   "lista vacia",
			items:    nil,
			esperado: 0,
		},
		{
			// CL-028-2: todas las historias completadas
			nombre: "todas completadas",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 2, Completado: true},
				{StoryPoints: 5, Completado: true},
			},
			esperado: 100,
		},
		{
			// CL-028-3: hay historias pero ninguna completada (distinto de la lista vacia)
			nombre: "ninguna completada",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 5, Completado: false},
				{StoryPoints: 3, Completado: false},
			},
			esperado: 0,
		},
		{
			// CL-028-4: una sola historia, completada
			nombre: "una sola historia completada",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 5, Completado: true},
			},
			esperado: 100,
		},
		{
			// CL-028-4: una sola historia, sin completar
			nombre: "una sola historia sin completar",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 5, Completado: false},
			},
			esperado: 0,
		},
		{
			// CA-028-1 (normal): 3 de 4 es exacto, no necesita redondeo
			nombre: "tres de cuatro completadas",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 3, Completado: true},
				{StoryPoints: 5, Completado: true},
				{StoryPoints: 2, Completado: true},
				{StoryPoints: 8, Completado: false},
			},
			esperado: 75,
		},
		{
			// CL-028-5 / RN-028-3: 1 de 3 es 33,333...: se redondea a 1 decimal
			nombre: "una de tres completadas",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 3, Completado: true},
				{StoryPoints: 5, Completado: false},
				{StoryPoints: 2, Completado: false},
			},
			esperado: 33.3,
		},
		{
			// CA-028-4 (limite) / CL-028-5 / RN-028-3: 2 de 3 es 66,666...: se redondea, no se trunca
			nombre: "dos de tres completadas",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 3, Completado: true},
				{StoryPoints: 5, Completado: true},
				{StoryPoints: 2, Completado: false},
			},
			esperado: 66.7,
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			obtenido, err := metricas.CalcularPorcentajeCompletadas(c.items)
			if err != nil {
				t.Fatalf("no se esperaba error: %v", err)
			}
			if obtenido != c.esperado {
				t.Fatalf("porcentaje = %v, se esperaba %v", obtenido, c.esperado)
			}
		})
	}
}
