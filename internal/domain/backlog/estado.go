package backlog

// Estado es el avance del trabajo de un item (RN-005-6). Estar en un sprint no es
// un estado: es una relacion entre el item y el sprint, y la resuelve US-009.
type Estado string

// Los estados posibles de un item. US-005 solo usa el inicial.
const (
	EstadoPendiente  Estado = "pendiente"
	EstadoEnProgreso Estado = "en_progreso"
	EstadoCompletado Estado = "completado"
)
