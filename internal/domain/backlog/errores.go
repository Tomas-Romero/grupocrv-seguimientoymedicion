package backlog

import "errors"

// Errores de validacion de un item (spec US-005, seccion 7). Son valores del
// paquete para compararlos con errors.Is, tambien cuando vienen unidos.
var (
	// ErrTituloVacio: el titulo esta vacio o tiene solo espacios (RN-005-2).
	ErrTituloVacio = errors.New("el titulo es obligatorio")
)
