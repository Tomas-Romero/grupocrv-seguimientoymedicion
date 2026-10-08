package backlog

import (
	"errors"
	"fmt"
	"slices"
)

// ErrEstimacionFueraDeEscala: la estimacion no esta en la escala Fibonacci del
// proyecto (RN-013-1). Se devuelve envuelto, indicando el valor recibido.
var ErrEstimacionFueraDeEscala = errors.New("la estimacion tiene que ser 1, 2, 3, 5, 8 o 13")

// escalaFibonacci son los unicos Story Points que se pueden asignar. No incluye
// el 0 (RN-013-2) ni el 21: un item de mas de 13 puntos es demasiado grande para
// un sprint y se divide.
var escalaFibonacci = []int{1, 2, 3, 5, 8, 13}

// EscalaFibonacci devuelve una copia de la escala de estimacion, de menor a
// mayor, para que quien la lea (por ejemplo, la pantalla de Planning Poker) use
// la misma fuente que la validacion.
func EscalaFibonacci() []int {
	return slices.Clone(escalaFibonacci)
}

// NuevosStoryPoints valida que puntos pertenezca a la escala y devuelve la
// estimacion. Si no pertenece, devuelve SinEstimar() y ErrEstimacionFueraDeEscala.
func NuevosStoryPoints(puntos int) (StoryPoints, error) {
	if !slices.Contains(escalaFibonacci, puntos) {
		return SinEstimar(), fmt.Errorf("%w (se recibio %d)", ErrEstimacionFueraDeEscala, puntos)
	}
	return StoryPoints{valor: puntos, estimado: true}, nil
}
