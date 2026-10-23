// Package planningpoker contiene las reglas del Planning Poker: la sesion sobre una
// historia, sus estados, sus rondas y sus participantes (US-014 a US-018).
//
// Es dominio puro: fuera de la biblioteca estandar solo toma de backlog el error de un
// item completado y la escala de estimacion (spec US-014, seccion 5). No accede a la
// base, no genera IDs y no lee el reloj.
package planningpoker

import (
	"errors"
	"fmt"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// MinimoParticipantes es la cantidad minima de participantes de una sesion: con uno no
// hay nada que comparar (RN-014-5).
const MinimoParticipantes = 2

// Estado es el estado de una sesion (RN-014-8). US-014 la crea votando; US-016 la
// revela, US-017 la vuelve a votando en una ronda nueva y US-018 la finaliza.
type Estado string

// Los estados posibles de una sesion.
const (
	EstadoVotando    Estado = "votando"
	EstadoRevelada   Estado = "revelada"
	EstadoFinalizada Estado = "finalizada"
)

// Abierta dice si la sesion esta abierta: lo esta mientras no se finalizo (RN-014-7).
func (e Estado) Abierta() bool {
	return e != EstadoFinalizada
}

// Errores de iniciar una sesion (spec US-014, seccion 7). Son valores para compararlos
// con errors.Is, tambien cuando vienen unidos con errors.Join (RN-014-10). El de una
// historia completada es backlog.ErrItemCompletado (RN-014-2).
var (
	// ErrSesionAbierta: la historia ya tiene una sesion que no esta finalizada
	// (RN-014-7).
	ErrSesionAbierta = errors.New("la historia ya tiene una sesion de planning poker abierta")
	// ErrParticipantesInsuficientes: el proyecto tiene menos de 2 integrantes
	// (RN-014-5). Se devuelve envuelto, indicando cuantos tiene.
	ErrParticipantesInsuficientes = errors.New("se necesitan al menos 2 integrantes en el proyecto para una sesion de planning poker")
)

// SesionPlanningPoker es una sesion de Planning Poker sobre una historia del backlog.
type SesionPlanningPoker struct {
	// ID lo genera la persistencia: esta vacio hasta que la sesion se registra
	// (RN-014-12).
	ID     string
	ItemID string
	// ProyectoID es el proyecto de la historia; lo completa la persistencia (US-015 lo
	// usa para verificar que quien vota sea de ese proyecto).
	ProyectoID string
	Estado     Estado
	// Ronda empieza en 1; US-017 la aumenta.
	Ronda int
	// Participantes son los IDs de los integrantes del proyecto al iniciar, en orden
	// de alta (RN-014-4). Quien se registre despues no participa.
	Participantes []string
}

// DatosDeLaHistoria son los datos minimos de la historia que necesita iniciar la
// sesion. Es un tipo propio del paquete, como DatosDelItem en sprint (US-009).
type DatosDeLaHistoria struct {
	// Completada dice si la historia esta completada (RN-014-2).
	Completada bool
	// ConSesionAbierta dice si la historia ya tiene una sesion que no esta finalizada
	// (RN-014-7).
	ConSesionAbierta bool
}

// NuevaSesion valida que se pueda iniciar una sesion sobre la historia y la arma
// votando, en la ronda 1, con los participantes recibidos (RN-014-6). La historia
// puede estar ya estimada: re-estimar esta permitido (RN-014-3).
//
// RN-014-10: no corta en el primer error. Devuelve todos unidos con errors.Join, en
// orden fijo: historia completada, sesion abierta y participantes insuficientes. La
// usa el caso de uso con los datos que leyo y la repite el repositorio con la historia
// bloqueada (RN-014-11).
func NuevaSesion(itemID string, historia DatosDeLaHistoria, participantes []string) (SesionPlanningPoker, error) {
	var errs []error
	if historia.Completada {
		errs = append(errs, backlog.ErrItemCompletado)
	}
	if historia.ConSesionAbierta {
		errs = append(errs, ErrSesionAbierta)
	}
	if len(participantes) < MinimoParticipantes {
		errs = append(errs, fmt.Errorf("%w (hay %d)", ErrParticipantesInsuficientes, len(participantes)))
	}
	if err := errors.Join(errs...); err != nil {
		return SesionPlanningPoker{}, err
	}
	return SesionPlanningPoker{
		ItemID:        itemID,
		Estado:        EstadoVotando,
		Ronda:         1,
		Participantes: slices.Clone(participantes),
	}, nil
}
