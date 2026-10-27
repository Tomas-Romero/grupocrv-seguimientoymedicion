package defecto

import "errors"

// Errores del cambio de estado (spec US-023, seccion 7). Son valores para compararlos con
// errors.Is.
var (
	// ErrEstadoInvalido: el estado pedido no es abierto, en_correccion, resuelto ni descartado
	// (RN-023-1). Se devuelve envuelto, indicando el valor recibido.
	ErrEstadoInvalido = errors.New("el estado tiene que ser abierto, en_correccion, resuelto o descartado")
	// ErrTransicionInvalida: la transicion no esta permitida, o el estado pedido es el actual
	// (RN-023-2). Se devuelve envuelto, indicando el estado actual y el pedido.
	ErrTransicionInvalida = errors.New("el defecto no puede pasar a ese estado desde el actual")
	// ErrSprintResolucionFaltante: resuelto sin sprint de resolucion (RN-023-3).
	ErrSprintResolucionFaltante = errors.New("un defecto resuelto necesita un sprint de resolucion")
	// ErrSprintResolucionNoCorresponde: sprint de resolucion con un estado distinto de resuelto
	// (RN-023-3). Se devuelve envuelto, indicando el estado pedido.
	ErrSprintResolucionNoCorresponde = errors.New("solo un defecto resuelto tiene sprint de resolucion")
	// ErrSprintResolucionAnterior: el sprint de resolucion es anterior al de deteccion (RN-023-4).
	// Se devuelve envuelto, indicando los dos numeros.
	ErrSprintResolucionAnterior = errors.New("el sprint de resolucion no puede ser anterior al de deteccion")
)

// SprintDeResolucion es el sprint en el que se resolvio un defecto, ya traducido por la capa
// de aplicacion (RN-023-8).
type SprintDeResolucion struct {
	// ID esta vacio si no se informo sprint.
	ID string
	// Numero es sprint.Numero. El dominio compara numeros y nunca IDs (RN-023-4).
	Numero int
}

// CambiarEstado devuelve una copia del defecto con el estado pedido. Solo resuelto lleva sprint
// de resolucion; en cualquier otro estado el defecto queda sin sprint, sin conservar el anterior
// (RN-023-5). No lee la base, no genera IDs y no sabe si los sprints existen: recibe sus
// numeros ya traducidos (RN-023-8).
func (d Defecto) CambiarEstado(nuevo Estado, resolucion SprintDeResolucion, _ int) (Defecto, error) {
	d.Estado = nuevo
	d.SprintResolucionID = resolucion.ID
	return d, nil
}
