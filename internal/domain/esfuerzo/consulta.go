package esfuerzo

import (
	"math"
	"time"
)

// RangoSprint son las fechas de un sprint del proyecto, tal como las necesita SprintDeLaFecha.
// El dominio recibe los rangos ya armados y no importa el paquete sprint.
type RangoSprint struct {
	ID     string
	Numero int
	// Inicio y Fin son dias, ambos incluidos (RN-020-3).
	Inicio time.Time
	Fin    time.Time
}

// SprintDeLaFecha devuelve el sprint cuyo rango contiene la fecha de un registro, o false si cae
// fuera de todos (RN-019-12 y RN-020-3). El dia de inicio y el de fin son del sprint, y se
// comparan solo el anio, el mes y el dia (UTC).
//
// US-008 impide que los sprints de un proyecto se superpongan, pero la funcion no lo da por
// hecho: si dos rangos contuvieran la fecha, gana el de menor numero (RN-020-13), sin importar
// el orden en que llegan.
func SprintDeLaFecha(fecha time.Time, rangos []RangoSprint) (RangoSprint, bool) {
	dia := soloElDia(fecha)

	var elegido RangoSprint
	encontrado := false
	for _, r := range rangos {
		if dia.Before(soloElDia(r.Inicio)) || dia.After(soloElDia(r.Fin)) {
			continue
		}
		if !encontrado || r.Numero < elegido.Numero {
			elegido, encontrado = r, true
		}
	}
	return elegido, encontrado
}

// TotalHoras es la suma de las horas de los registros, redondeada a 2 decimales para no mostrar
// ruido de coma flotante (RN-020-6). Sin registros es 0.
func TotalHoras(registros []RegistroEsfuerzo) float64 {
	var total float64
	for _, r := range registros {
		total += r.Horas
	}
	return math.Round(total*100) / 100
}
