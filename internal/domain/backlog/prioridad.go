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
