package metricas_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
)

// US-025: el resumen de un sprint suma los Story Points de todos sus items
// (planificados) y solo los de los completados (completados).
func TestCalcularResumenSprint(t *testing.T) {
	casos := []struct {
		nombre   string
		items    []metricas.ItemDelSprint
		esperado metricas.ResumenSprint
	}{
		{
			// CA-025-1 (normal) / RN-025-1 / RN-025-2
			nombre: "items completados y pendientes",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 5, Completado: true},
				{StoryPoints: 3, Completado: false},
				{StoryPoints: 8, Completado: true},
			},
			esperado: metricas.ResumenSprint{Planificados: 16, Completados: 13},
		},
		{
			// CA-025-2 (alternativo) / CL-025-3
			nombre: "todos los items completados",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 2, Completado: true},
				{StoryPoints: 5, Completado: true},
			},
			esperado: metricas.ResumenSprint{Planificados: 7, Completados: 7},
		},
		{
			// CA-025-3 (limite) / CL-025-1
			nombre:   "sprint sin items",
			items:    nil,
			esperado: metricas.ResumenSprint{Planificados: 0, Completados: 0},
		},
		{
			// CL-025-2
			nombre: "un solo item completado",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 5, Completado: true},
			},
			esperado: metricas.ResumenSprint{Planificados: 5, Completados: 5},
		},
		{
			// CL-025-4
			nombre: "ningun item completado",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 5, Completado: false},
				{StoryPoints: 3, Completado: false},
			},
			esperado: metricas.ResumenSprint{Planificados: 8, Completados: 0},
		},
		{
			// CL-025-5 / RN-025-3: un item con 0 puntos no altera ninguna suma
			nombre: "items con cero story points",
			items: []metricas.ItemDelSprint{
				{StoryPoints: 0, Completado: true},
				{StoryPoints: 0, Completado: false},
				{StoryPoints: 3, Completado: true},
			},
			esperado: metricas.ResumenSprint{Planificados: 3, Completados: 3},
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			obtenido, err := metricas.CalcularResumenSprint(c.items)
			if err != nil {
				t.Fatalf("no se esperaba error: %v", err)
			}
			if obtenido != c.esperado {
				t.Fatalf("resumen = %+v, se esperaba %+v", obtenido, c.esperado)
			}
		})
	}
}

// US-025 / CA-025-4 (error): un item con Story Points negativos devuelve
// ErrStoryPointsNegativos con la posicion del item (desde 1) y ningun resumen
// parcial.
func TestCalcularResumenSprint_StoryPointsNegativos(t *testing.T) {
	items := []metricas.ItemDelSprint{
		{StoryPoints: 5, Completado: true},
		{StoryPoints: -1, Completado: false},
		{StoryPoints: 3, Completado: true},
	}

	obtenido, err := metricas.CalcularResumenSprint(items)

	if !errors.Is(err, metricas.ErrStoryPointsNegativos) {
		t.Fatalf("error = %v, se esperaba ErrStoryPointsNegativos", err)
	}
	if !strings.Contains(err.Error(), "posicion 2") {
		t.Fatalf("el error %q no indica la posicion 2 (contada desde 1)", err)
	}
	if obtenido != (metricas.ResumenSprint{}) {
		t.Fatalf("resumen = %+v, se esperaba el valor cero (sin resumen parcial)", obtenido)
	}
}
