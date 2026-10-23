package metricas

import (
	"errors"
	"fmt"
	"math"
)

// HorasPorStoryPoint es la equivalencia que el equipo usa desde el Sprint 0:
// 1 Story Point = 1,5 horas (RN-027-1). Se exporta para que la pantalla pueda
// mostrar la que se uso.
const HorasPorStoryPoint = 1.5

// ErrHorasInvalidas se devuelve cuando una hora registrada no es un numero mayor
// a 0. Se devuelve envuelto, despues de la posicion de la hora en la lista.
var ErrHorasInvalidas = errors.New("no es un numero mayor a 0")

// ResumenHoras son las horas de un sprint: las estimadas a partir de sus Story
// Points, las trabajadas de verdad y la diferencia entre ambas. Todas en horas y
// redondeadas a 2 decimales.
type ResumenHoras struct {
	Estimadas  float64
	Reales     float64
	Desviacion float64 // Reales - Estimadas: positiva si se trabajo mas de lo estimado
}

// DesviacionPorcentual devuelve la desviacion como porcentaje de las horas
// estimadas, redondeada a 1 decimal (RN-027-4). Sin horas estimadas no hay base
// de comparacion: devuelve 0 y false, y no es un error.
func (r ResumenHoras) DesviacionPorcentual() (porcentaje float64, calculable bool) {
	if r.Estimadas <= 0 {
		return 0, false
	}
	return math.Round(r.Desviacion/r.Estimadas*1000) / 10, true
}

// CalcularHoras arma el resumen de horas de un sprint (RN-027-1 a RN-027-3): las
// estimadas son los Story Points planificados por HorasPorStoryPoint, las reales
// la suma de las horas registradas, sin filtrar nada, y la desviacion la
// diferencia. Una lista vacia o nil es un sprint sin horas, no un error.
//
// Valida todo antes de calcular. Si algo es invalido devuelve todos los errores
// unidos con errors.Join, en el orden de RN-027-7 (primero los Story Points y
// despues cada hora por posicion, contada desde 1), y ningun resumen: una cifra
// calculada sobre datos corruptos seria peor que un error.
func CalcularHoras(storyPointsPlanificados int, horasRegistradas []float64) (ResumenHoras, error) {
	var errs []error
	if storyPointsPlanificados < 0 {
		errs = append(errs, fmt.Errorf("%d story points planificados: %w", storyPointsPlanificados, ErrStoryPointsNegativos))
	}

	reales := 0.0
	for i, h := range horasRegistradas {
		if math.IsNaN(h) || math.IsInf(h, 0) || h <= 0 {
			errs = append(errs, fmt.Errorf("la hora de la posicion %d %w", i+1, ErrHorasInvalidas))
			continue
		}
		reales += h
	}
	if err := errors.Join(errs...); err != nil {
		return ResumenHoras{}, err
	}

	estimadas := redondear(float64(storyPointsPlanificados)*HorasPorStoryPoint, 100)
	reales = redondear(reales, 100)
	return ResumenHoras{
		Estimadas:  estimadas,
		Reales:     reales,
		Desviacion: redondear(reales-estimadas, 100),
	}, nil
}

// redondear redondea a la cantidad de decimales que indica el factor (100 son 2
// decimales): oculta el ruido de la coma flotante, como 7.500000000000001.
func redondear(valor, factor float64) float64 {
	return math.Round(valor*factor) / factor
}
