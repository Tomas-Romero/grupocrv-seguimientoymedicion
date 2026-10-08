package backlog

// Prioridad es la prioridad MoSCoW de un item (RN-005-5).
type Prioridad string

// Los cuatro valores que acepta el dominio, exactos y en minuscula.
const (
	PrioridadMust   Prioridad = "must"
	PrioridadShould Prioridad = "should"
	PrioridadCould  Prioridad = "could"
	PrioridadWont   Prioridad = "wont"
)

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
