package backlog

import "errors"

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
)
