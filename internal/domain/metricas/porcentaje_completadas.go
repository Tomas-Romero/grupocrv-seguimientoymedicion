package metricas

import (
	"fmt"
	"math"
)

// CalcularPorcentajeCompletadas devuelve el porcentaje de historias completadas:
// las que tienen Completado = true sobre el total, por 100 (RN-028-1). Cuenta
// historias y no Story Points (RN-028-2). Se redondea a 1 decimal, como la
// velocidad, porque se muestra a personas (RN-028-3 / CL-028-5).
//
// Una lista vacia o nil es un sprint sin historias: devuelve 0 y no es un error
// (RN-028-4 / CL-028-1). Hay que cortarla antes de dividir, porque 0/0 daria NaN.
//
// Valida todos los items antes de contar, con la regla de US-025: si alguno tiene
// Story Points negativos devuelve ErrStoryPointsNegativos con la posicion del
// primero, contada desde 1, y ningun porcentaje (RN-028-5 / CL-028-8). Aunque no
// usa los puntos, un dato corrupto no se calcula en silencio.
func CalcularPorcentajeCompletadas(items []ItemDelSprint) (float64, error) {
	for i, item := range items {
		if item.StoryPoints < 0 {
			return 0, fmt.Errorf("item en la posicion %d: %w", i+1, ErrStoryPointsNegativos)
		}
	}

	if len(items) == 0 {
		return 0, nil
	}

	completadas := 0
	for _, item := range items {
		if item.Completado {
			completadas++
		}
	}

	porcentaje := float64(completadas) / float64(len(items)) * 100
	return math.Round(porcentaje*10) / 10, nil
}
