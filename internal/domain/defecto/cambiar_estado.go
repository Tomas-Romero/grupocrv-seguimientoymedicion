package defecto

import (
	"errors"
	"fmt"
	"slices"
)

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

// transicionesPermitidas es la tabla de RN-023-2: lo que no esta listado esta prohibido, incluido
// pasar de un estado a si mismo. Desde resuelto y descartado solo se reabre.
var transicionesPermitidas = map[Estado][]Estado{
	EstadoAbierto:      {EstadoEnCorreccion, EstadoResuelto, EstadoDescartado},
	EstadoEnCorreccion: {EstadoAbierto, EstadoResuelto, EstadoDescartado},
	EstadoResuelto:     {EstadoAbierto},
	EstadoDescartado:   {EstadoAbierto},
}

// CambiarEstado devuelve una copia del defecto con el estado pedido. Solo resuelto lleva sprint
// de resolucion; en cualquier otro estado el defecto queda sin sprint, sin conservar el anterior
// (RN-023-5). No lee la base, no genera IDs y no sabe si los sprints existen: recibe sus
// numeros ya traducidos (RN-023-8).
func (d Defecto) CambiarEstado(nuevo Estado, resolucion SprintDeResolucion, _ int) (Defecto, error) {
	// RN-023-1: valores exactos; no se recorta ni se pasa a minuscula. Con un estado invalido no
	// se sabe a que estado se pasa, asi que ninguna otra regla se puede evaluar (RN-023-6).
	switch nuevo {
	case EstadoAbierto, EstadoEnCorreccion, EstadoResuelto, EstadoDescartado:
	default:
		return Defecto{}, fmt.Errorf("%w (se recibio %q)", ErrEstadoInvalido, string(nuevo))
	}

	if !slices.Contains(transicionesPermitidas[d.Estado], nuevo) {
		return Defecto{}, fmt.Errorf("%w (de %s a %s)", ErrTransicionInvalida, d.Estado, nuevo)
	}
	// RN-023-3: solo resuelto lleva sprint de resolucion, y ahi es obligatorio. Informar uno con
	// otro estado se rechaza para no descartarlo en silencio.
	switch {
	case nuevo == EstadoResuelto && resolucion.ID == "":
		return Defecto{}, ErrSprintResolucionFaltante
	case nuevo != EstadoResuelto && resolucion.ID != "":
		return Defecto{}, fmt.Errorf("%w (se pidio el estado %s)", ErrSprintResolucionNoCorresponde, nuevo)
	}

	d.Estado = nuevo
	d.SprintResolucionID = resolucion.ID
	return d, nil
}
