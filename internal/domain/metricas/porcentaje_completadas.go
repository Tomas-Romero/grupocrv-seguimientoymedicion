package metricas

// CalcularPorcentajeCompletadas devuelve el porcentaje de historias completadas:
// las que tienen Completado = true sobre el total, por 100 (RN-028-1). Cuenta
// historias y no Story Points (RN-028-2).
//
// Una lista vacia o nil es un sprint sin historias: devuelve 0 y no es un error
// (RN-028-4 / CL-028-1). Hay que cortarla antes de dividir, porque 0/0 daria NaN.
func CalcularPorcentajeCompletadas(items []ItemDelSprint) (float64, error) {
	if len(items) == 0 {
		return 0, nil
	}

	completadas := 0
	for _, item := range items {
		if item.Completado {
			completadas++
		}
	}

	return float64(completadas) / float64(len(items)) * 100, nil
}
