package backlog

import "fmt"

// Prioridad es la prioridad MoSCoW de un item (RN-005-5).
type Prioridad string

// Los cuatro valores que acepta el dominio, exactos y en minuscula.
const (
	PrioridadMust   Prioridad = "must"
	PrioridadShould Prioridad = "should"
	PrioridadCould  Prioridad = "could"
	PrioridadWont   Prioridad = "wont"
)

// ValidarPrioridad devuelve nil si la prioridad es una de las cuatro, y ErrPrioridadInvalida envuelto
// con el valor recibido en cualquier otro caso, incluida la vacia (RN-005-5).
//
// Es la regla del alta expuesta para que el filtro del backlog (RN-007-8) y la creacion usen la
// misma, escrita una sola vez. Que en el filtro la prioridad vacia signifique "sin filtro" lo decide el
// caso de uso antes de llamarla; aca la vacia es invalida, como en el alta.
func ValidarPrioridad(prioridad Prioridad) error {
	if !prioridad.valida() {
		return fmt.Errorf("%w: se recibio %q", ErrPrioridadInvalida, string(prioridad))
	}
	return nil
}

// valida dice si p es uno de los cuatro valores exactos. No pasa a minuscula ni
// recorta: convertir el texto de un formulario es tarea del adaptador, y "MUST"
// o " must " se rechazan (CL-005-12).
func (p Prioridad) valida() bool {
	switch p {
	case PrioridadMust, PrioridadShould, PrioridadCould, PrioridadWont:
		return true
	}
	return false
}
