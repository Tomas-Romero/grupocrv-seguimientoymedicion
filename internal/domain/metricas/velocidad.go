package metricas

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	// ErrResumenInconsistente: un sprint cerrado tiene mas Story Points completados
	// que planificados, lo que no puede pasar (RN-025-4): es un dato corrupto. Se
	// devuelve envuelto, despues del nombre del sprint.
	ErrResumenInconsistente = errors.New("tiene mas Story Points completados que planificados: revisa sus datos")

	// ErrNombreVacio: un sprint cerrado no tiene nombre. Se devuelve envuelto,
	// indicando su posicion en la lista contada desde 1.
	ErrNombreVacio = errors.New("todo sprint cerrado necesita un nombre para poder mostrarlo")
)

// SprintCerrado es un sprint ya cerrado con su resumen (US-025). El nombre no
// interviene en el promedio: se lleva para poder mostrar la velocidad por sprint.
type SprintCerrado struct {
	Nombre  string
	Resumen ResumenSprint
}

// CalcularVelocidad devuelve el promedio de los Story Points completados de los
// sprints cerrados, redondeado a 1 decimal (RN-026-1 y RN-026-2). Sin sprints
// cerrados la velocidad es 0 y no es un error: es lo normal antes de cerrar el
// primero.
//
// Valida todos los sprints antes de promediar. Si alguno es invalido devuelve
// todos los errores unidos con errors.Join, en el orden de la lista (RN-026-4), y
// ninguna velocidad: una cifra calculada sobre datos corruptos seria peor que un
// error.
func CalcularVelocidad(sprints []SprintCerrado) (float64, error) {
	var errs []error
	total := 0

	for i, sprint := range sprints {
		if strings.TrimSpace(sprint.Nombre) == "" {
			errs = append(errs, fmt.Errorf("sprint en la posicion %d: %w", i+1, ErrNombreVacio))
		}
		if sprint.Resumen.Completados > sprint.Resumen.Planificados {
			errs = append(errs, fmt.Errorf("el sprint %q %w", sprint.Nombre, ErrResumenInconsistente))
		}
		total += sprint.Resumen.Completados
	}

	if err := errors.Join(errs...); err != nil {
		return 0, err
	}
	if len(sprints) == 0 {
		return 0, nil
	}

	promedio := float64(total) / float64(len(sprints))
	return math.Round(promedio*10) / 10, nil
}
