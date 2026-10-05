// Package metricas calcula las metricas del proyecto a partir de datos que le
// pasa quien llama. Es dominio puro: solo usa la biblioteca estandar y no lee el
// reloj, ni la base, ni el entorno.
package metricas

import (
	"errors"
	"fmt"
)

// ErrStoryPointsNegativos se devuelve cuando algun item del sprint tiene Story
// Points negativos: es un dato inconsistente, no un caso a calcular.
var ErrStoryPointsNegativos = errors.New("los story points de un item del sprint no pueden ser negativos")

// ItemDelSprint es lo unico que el calculo necesita saber de un item asignado a
// un sprint. Un item sin estimar no se representa aca: la capa de aplicacion lo
// rechaza al adaptarlo, porque en el backlog "sin estimar" no es lo mismo que 0.
type ItemDelSprint struct {
	StoryPoints int
	Completado  bool
}

// ResumenSprint son los Story Points de un sprint. Completados nunca supera a
// Planificados, porque todo item completado forma parte de los planificados.
type ResumenSprint struct {
	Planificados int
	Completados  int
}

// CalcularResumenSprint suma los Story Points de todos los items (planificados)
// y solo los de los completados (completados).
//
// Si algun item tiene Story Points negativos devuelve ErrStoryPointsNegativos
// indicando su posicion en la lista, contada desde 1, y ningun resumen parcial:
// una cifra a medias seria mas confusa que un error.
func CalcularResumenSprint(items []ItemDelSprint) (ResumenSprint, error) {
	var resumen ResumenSprint

	for i, item := range items {
		if item.StoryPoints < 0 {
			return ResumenSprint{}, fmt.Errorf("item en la posicion %d: %w", i+1, ErrStoryPointsNegativos)
		}
		resumen.Planificados += item.StoryPoints
		if item.Completado {
			resumen.Completados += item.StoryPoints
		}
	}

	return resumen, nil
}
