package backlog

import "fmt"

// Estado es el avance del trabajo de un item (RN-005-6). Estar en un sprint no es
// un estado: es una relacion entre el item y el sprint, y la resuelve US-009.
type Estado string

// Los estados posibles de un item. US-005 solo usa el inicial.
const (
	EstadoPendiente  Estado = "pendiente"
	EstadoEnProgreso Estado = "en_progreso"
	EstadoCompletado Estado = "completado"
)

// ValidarEstado devuelve nil si el estado es uno de los tres, y ErrEstadoInvalido envuelto con el valor
// recibido en cualquier otro caso, incluido el vacio (RN-007-7). Los valores son exactos: no se pasa a
// minuscula ni se recorta. Que en el filtro el estado vacio signifique "sin filtro" lo decide el caso de
// uso antes de llamarla.
func ValidarEstado(estado Estado) error {
	switch estado {
	case EstadoPendiente, EstadoEnProgreso, EstadoCompletado:
		return nil
	}
	return fmt.Errorf("%w: se recibio %q", ErrEstadoInvalido, string(estado))
}
