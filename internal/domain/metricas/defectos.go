package metricas

import (
	"errors"
	"fmt"
)

var (
	// ErrSprintInvalido: un numero de sprint es menor que 1. Se devuelve envuelto, indicando el
	// sprint consultado o la posicion del defecto (RN-029-6 y RN-029-7).
	ErrSprintInvalido = errors.New("el numero de sprint tiene que ser 1 o mayor")

	// ErrResolucionAnteriorADeteccion: un defecto figura resuelto en un sprint anterior al de su
	// deteccion. Se devuelve envuelto, indicando la posicion del defecto (RN-029-7).
	ErrResolucionAnteriorADeteccion = errors.New("el defecto se resolvio antes de detectarse")
)

// DefectoRegistrado es lo unico que el calculo necesita saber de un defecto (spec US-029,
// seccion 2): en que sprint se detecto y en cual se resolvio. Los sprints son numeros, para poder
// verificar el orden; quien llama traduce el sprint de cada defecto a su numero.
type DefectoRegistrado struct {
	SprintDeteccion int
	// SprintResolucion es 0 mientras el defecto sigue sin resolver.
	SprintResolucion int
}

// ResumenDefectos son los defectos de un sprint: los que se detectaron en el y los que se
// resolvieron en el, sin importar donde se detectaron.
type ResumenDefectos struct {
	Detectados int
	Resueltos  int
}

// CalcularDefectosDelSprint cuenta los defectos detectados en el sprint (RN-029-1) y los resueltos
// en el (RN-029-2). Un defecto detectado y resuelto en el mismo sprint cuenta en las dos cifras
// (RN-029-3) y uno sin resolver, solo como detectado (RN-029-4). Una lista vacia o nil da 0 y 0, y
// no es un error (RN-029-5).
//
// Valida todo antes de contar. Si algo es invalido devuelve todos los errores unidos con
// errors.Join, en el orden de RN-029-8 (primero el sprint consultado y despues cada defecto por
// posicion, contada desde 1), y ningun resumen: una cifra calculada sobre datos corruptos seria
// peor que un error.
func CalcularDefectosDelSprint(sprint int, defectos []DefectoRegistrado) (ResumenDefectos, error) {
	var errs []error
	if sprint < 1 {
		errs = append(errs, fmt.Errorf("sprint consultado %d: %w", sprint, ErrSprintInvalido))
	}
	for i, defecto := range defectos {
		if defecto.SprintDeteccion < 1 {
			errs = append(errs, fmt.Errorf("defecto en la posicion %d (detectado en el sprint %d): %w", i+1, defecto.SprintDeteccion, ErrSprintInvalido))
		}
		if defecto.SprintResolucion != 0 && defecto.SprintResolucion < defecto.SprintDeteccion {
			errs = append(errs, fmt.Errorf("defecto en la posicion %d (detectado en el sprint %d y resuelto en el %d): %w",
				i+1, defecto.SprintDeteccion, defecto.SprintResolucion, ErrResolucionAnteriorADeteccion))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return ResumenDefectos{}, err
	}

	var resumen ResumenDefectos
	for _, defecto := range defectos {
		if defecto.SprintDeteccion == sprint {
			resumen.Detectados++
		}
		if defecto.SprintResolucion == sprint {
			resumen.Resueltos++
		}
	}
	return resumen, nil
}
