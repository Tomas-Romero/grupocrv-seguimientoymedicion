package backlog

import (
	"errors"
	"fmt"
)

// Errores de validacion de un item (spec US-005, seccion 7). Son valores del
// paquete para compararlos con errors.Is, tambien cuando vienen unidos.
var (
	// ErrTituloVacio: el titulo esta vacio o tiene solo espacios (RN-005-2).
	ErrTituloVacio = errors.New("el titulo es obligatorio")
	// ErrTituloMuyLargo: el titulo recortado pasa los 120 caracteres (RN-005-3).
	// Se devuelve envuelto, indicando cuantos caracteres tiene.
	ErrTituloMuyLargo = errors.New("el titulo no puede tener mas de 120 caracteres")
	// ErrPrioridadInvalida: la prioridad no es must, should, could ni wont
	// (RN-005-5). Se devuelve envuelto, indicando el valor recibido.
	ErrPrioridadInvalida = errors.New("la prioridad tiene que ser must, should, could o wont")
	// ErrCriterioVacio: un criterio de aceptacion esta vacio o tiene solo
	// espacios (RN-005-8). Se devuelve como CriterioVacioError, con su posicion.
	ErrCriterioVacio = errors.New("el criterio de aceptacion esta vacio")
)

// CriterioVacioError es el error de un criterio vacio. Lleva su posicion,
// contada desde 1 porque el mensaje lo lee una persona (RN-005-14), y se
// reconoce con errors.Is(err, ErrCriterioVacio). Con errors.As se obtiene la
// posicion, por ejemplo para marcar el campo en el formulario.
type CriterioVacioError struct {
	Posicion int
}

func (e CriterioVacioError) Error() string {
	return fmt.Sprintf("el criterio de aceptacion de la posicion %d esta vacio", e.Posicion)
}

// Unwrap hace que errors.Is reconozca el error como ErrCriterioVacio.
func (e CriterioVacioError) Unwrap() error {
	return ErrCriterioVacio
}
