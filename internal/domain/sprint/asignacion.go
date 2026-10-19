package sprint

import (
	"errors"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// Errores de la asignacion de un item a un sprint (spec US-009, seccion 7). Son
// valores para compararlos con errors.Is, tambien cuando vienen unidos con
// errors.Join (RN-009-8). El error de un item completado es backlog.ErrItemCompletado:
// es el mismo concepto que en US-013 y US-006 (RN-009-5).
var (
	// ErrProyectoDistinto: el item y el sprint son de proyectos distintos (RN-009-3).
	ErrProyectoDistinto = errors.New("la historia y el sprint son de proyectos distintos")
	// ErrSprintCerrado: el sprint esta cerrado (RN-009-4). Se devuelve envuelto,
	// indicando el nombre del sprint. El mensaje es general a proposito: lo reusan
	// las otras operaciones sobre un sprint (US-010).
	ErrSprintCerrado = errors.New("el sprint esta cerrado")
	// ErrItemYaAsignado: el item ya esta en el sprint abierto del proyecto (RN-009-6).
	ErrItemYaAsignado = errors.New("la historia ya esta asignada a un sprint")
	// ErrItemSinEstimar: el item no tiene Story Points (RN-009-7, Definition of Ready).
	ErrItemSinEstimar = errors.New("la historia no esta estimada: estimala antes de asignarla")
	// ErrItemSinCriterios: el item no tiene criterios de aceptacion (RN-009-7,
	// Definition of Ready).
	ErrItemSinCriterios = errors.New("la historia no tiene criterios de aceptacion")
)

// DatosDelItem son los datos minimos de un item que necesita la asignacion. Es un
// tipo propio de este paquete, como ItemDelSprint en US-025: el dominio de sprint no
// usa ItemBacklog. La capa de aplicacion adapta el item a este tipo.
type DatosDelItem struct {
	ProyectoID string
	// Completado dice si el item esta en estado completado.
	Completado bool
	// Estimado dice si tiene Story Points (presentes, no ausentes): un item con 0
	// puntos esta estimado.
	Estimado bool
	// Criterios es la cantidad de criterios de aceptacion.
	Criterios int
	// EnSprintAbierto dice si el item ya esta en el sprint abierto del proyecto, su
	// sprint actual. Haber estado en sprints cerrados no cuenta (RN-009-6).
	EnSprintAbierto bool
}

// DatosDelSprint son los datos minimos de un sprint que necesita la asignacion.
type DatosDelSprint struct {
	ProyectoID string
	Nombre     string
	Estado     Estado
}

// VerificarAsignacion comprueba las reglas RN-009-3 a RN-009-7: el item y el sprint son
// del mismo proyecto, el sprint esta abierto, el item no esta completado ni ya
// asignado y cumple la Definition of Ready.
//
// RN-009-8: no corta en el primer error. Devuelve todos unidos con errors.Join, en
// orden fijo: proyecto distinto, sprint cerrado, item completado, item ya asignado,
// item sin estimar y item sin criterios. Con datos validos devuelve nil.
//
// La usa el caso de uso con los datos que leyo y la repite el repositorio dentro de la
// transaccion, con los datos leidos con los bloqueos tomados (RN-009-11).
//
// Que el item y el sprint existan no se verifica aca: lo hacen la aplicacion y el
// repositorio, antes (RN-009-1 y RN-009-2). Tampoco se valida la capacidad del sprint
// ni se compara con las fechas (RN-009-12).
func VerificarAsignacion(item DatosDelItem, sprint DatosDelSprint) error {
	var errs []error

	if item.ProyectoID != sprint.ProyectoID {
		errs = append(errs, ErrProyectoDistinto)
	}
	if sprint.Estado == EstadoCerrado {
		errs = append(errs, fmt.Errorf("%w (%s)", ErrSprintCerrado, sprint.Nombre))
	}
	if item.Completado {
		errs = append(errs, backlog.ErrItemCompletado)
	}
	if item.EnSprintAbierto {
		errs = append(errs, ErrItemYaAsignado)
	}
	if !item.Estimado {
		errs = append(errs, ErrItemSinEstimar)
	}
	if item.Criterios == 0 {
		errs = append(errs, ErrItemSinCriterios)
	}
	return errors.Join(errs...)
}
