// Package sprint contiene las reglas de un Sprint: su Sprint Goal, sus fechas, su
// estado y su lugar en la secuencia de sprints de un proyecto (US-008).
//
// Es dominio puro: no importa nada fuera de la biblioteca estandar, no accede a la
// base, no genera IDs y no lee el reloj. Que el proyecto exista, que numero le toca
// al sprint y que sprints tiene ya el proyecto lo resuelven la capa de aplicacion y
// el repositorio.
package sprint

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// Limites de un sprint (RN-008-3 y RN-008-6).
const (
	// MaxLargoSprintGoal es el maximo de caracteres del Sprint Goal recortado.
	MaxLargoSprintGoal = 200
	// MaxDias es la duracion maxima de un sprint, contando el dia de inicio y el
	// de fin. Se toma 28 por ser el mes mas corto.
	MaxDias = 28
)

// Estado es el estado de un sprint (RN-008-8). Se crea abierto; cerrarlo es US-011.
type Estado string

// Los estados posibles de un sprint.
const (
	EstadoAbierto Estado = "abierto"
	EstadoCerrado Estado = "cerrado"
)

// Errores del paquete (spec US-008, seccion 7). Son valores para compararlos con
// errors.Is, tambien cuando vienen unidos con errors.Join (RN-008-10).
var (
	// ErrSprintGoalVacio: el Sprint Goal esta vacio o tiene solo espacios (RN-008-2).
	ErrSprintGoalVacio = errors.New("el Sprint Goal es obligatorio")
	// ErrSprintGoalMuyLargo: el Sprint Goal recortado pasa los 200 caracteres
	// (RN-008-3). Se devuelve envuelto, indicando cuantos tiene.
	ErrSprintGoalMuyLargo = errors.New("el Sprint Goal no puede tener mas de 200 caracteres")
	// ErrFechaInicioFaltante: no se informo la fecha de inicio (RN-008-4).
	ErrFechaInicioFaltante = errors.New("falta la fecha de inicio")
	// ErrFechaFinFaltante: no se informo la fecha de fin (RN-008-4).
	ErrFechaFinFaltante = errors.New("falta la fecha de fin")
	// ErrFechasIncoherentes: la fecha de fin es anterior a la de inicio (RN-008-5).
	ErrFechasIncoherentes = errors.New("fecha de fin anterior a la de inicio")
	// ErrDuracionExcedida: el sprint dura mas de 28 dias (RN-008-6). Se devuelve
	// envuelto, indicando cuantos dias dura.
	ErrDuracionExcedida = errors.New("el sprint no puede durar mas de 28 dias")
	// ErrFueraDelProyecto: alguna fecha queda fuera de las del proyecto (RN-008-7).
	// Se devuelve envuelto, indicando las fechas del proyecto.
	ErrFueraDelProyecto = errors.New("las fechas del sprint tienen que estar dentro de las del proyecto")
	// ErrUltimoSprintAbierto: el ultimo sprint del proyecto sigue abierto
	// (RN-008-17). Se devuelve envuelto, indicando su nombre.
	ErrUltimoSprintAbierto = errors.New("no se puede crear un sprint mientras el ultimo sigue abierto: primero hay que cerrarlo")
	// ErrInicioNoPosteriorAlUltimo: el ultimo sprint esta cerrado y el nuevo empieza
	// el dia en que termina, o antes (RN-008-13). Se devuelve envuelto, indicando el
	// nombre y el fin del ultimo.
	ErrInicioNoPosteriorAlUltimo = errors.New("el sprint tiene que empezar despues de que termine el ultimo sprint del proyecto")
)

// Sprint es un sprint de un proyecto ya validado.
type Sprint struct {
	// ID esta vacio hasta que la persistencia lo asigna (RN-008-16).
	ID string
	// Numero es el correlativo dentro del proyecto (1, 2, ...); lo asigna la
	// persistencia (RN-008-11). Cero significa "todavia sin numero".
	Numero      int
	ProyectoID  string
	SprintGoal  string
	FechaInicio time.Time
	FechaFin    time.Time
	Estado      Estado
}

// Nombre es "Sprint N", armado a partir del numero (RN-008-12): no se guarda
// aparte, asi no puede quedar distinto de el. Un sprint todavia sin numero no tiene
// nombre.
func (s Sprint) Nombre() string {
	return nombreDe(s.Numero)
}

func nombreDe(numero int) string {
	if numero <= 0 {
		return ""
	}
	return fmt.Sprintf("Sprint %d", numero)
}

// Nuevo valida los datos y arma el sprint abierto, sin ID ni numero. inicioProyecto y
// finProyecto son las fechas del proyecto: no las carga la persona, las pasa el
// caso de uso.
//
// RN-008-10: valida todo y devuelve todos los errores unidos con errors.Join, en
// orden fijo: Sprint Goal, falta inicio, falta fin y, solo si estan las dos fechas,
// fechas incoherentes, duracion y fuera del proyecto. Si hay algun error no devuelve
// sprint.
//
// Las fechas llegan normalizadas a medianoche UTC; el dominio solo las compara y
// cuenta dias.
func Nuevo(proyectoID, sprintGoal string, inicio, fin, inicioProyecto, finProyecto time.Time) (Sprint, error) {
	var errs []error

	// RN-008-1 a RN-008-3: se recorta antes de validar, y el limite se mide en
	// caracteres y no en bytes.
	sprintGoal = strings.TrimSpace(sprintGoal)
	switch largo := utf8.RuneCountInString(sprintGoal); {
	case largo == 0:
		errs = append(errs, ErrSprintGoalVacio)
	case largo > MaxLargoSprintGoal:
		errs = append(errs, fmt.Errorf("%w (tiene %d)", ErrSprintGoalMuyLargo, largo))
	}

	// RN-008-4.
	if inicio.IsZero() {
		errs = append(errs, ErrFechaInicioFaltante)
	}
	if fin.IsZero() {
		errs = append(errs, ErrFechaFinFaltante)
	}

	// RN-008-5 a RN-008-7: comparan fechas, asi que solo se evaluan si estan las dos.
	if !inicio.IsZero() && !fin.IsZero() {
		if fin.Before(inicio) {
			errs = append(errs, ErrFechasIncoherentes)
		}
		// Con las fechas invertidas la duracion da 0 o menos: no es excedida.
		if dias := duracionEnDias(inicio, fin); dias > MaxDias {
			errs = append(errs, fmt.Errorf("%w (dura %d)", ErrDuracionExcedida, dias))
		}
		if err := VerificarDentroDelProyecto(inicio, fin, inicioProyecto, finProyecto); err != nil {
			errs = append(errs, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		return Sprint{}, err
	}
	return Sprint{
		ProyectoID:  proyectoID,
		SprintGoal:  sprintGoal,
		FechaInicio: inicio,
		FechaFin:    fin,
		Estado:      EstadoAbierto,
	}, nil
}

// duracionEnDias cuenta los dias de un sprint contando el de inicio y el de fin
// (RN-008-6): un sprint de un dia dura 1.
func duracionEnDias(inicio, fin time.Time) int {
	return int(fin.Sub(inicio).Hours()/24) + 1
}

// VerificarDentroDelProyecto devuelve ErrFueraDelProyecto si alguna de las dos
// fechas queda fuera de las del proyecto, con los extremos incluidos (RN-008-7).
// Es un solo error aunque las dos esten afuera. La usa Nuevo y la repite el
// repositorio dentro de la transaccion, con las fechas del proyecto leidas con la
// fila bloqueada (RN-008-14).
func VerificarDentroDelProyecto(inicio, fin, inicioProyecto, finProyecto time.Time) error {
	if fueraDelRango(inicio, inicioProyecto, finProyecto) || fueraDelRango(fin, inicioProyecto, finProyecto) {
		return fmt.Errorf("%w (el proyecto va del %s al %s)", ErrFueraDelProyecto,
			inicioProyecto.Format(time.DateOnly), finProyecto.Format(time.DateOnly))
	}
	return nil
}

// fueraDelRango dice si la fecha queda antes del inicio o despues del fin del
// proyecto. Se mira cada fecha por separado: con las fechas invertidas, el fin
// puede quedar antes del inicio del proyecto aunque el inicio no (CL-008-12).
func fueraDelRango(fecha, desde, hasta time.Time) bool {
	return fecha.Before(desde) || fecha.After(hasta)
}

// UltimoSprint son los datos del ultimo sprint de un proyecto que necesita la
// verificacion de la transaccion: el de numero mas alto.
type UltimoSprint struct {
	Numero   int
	Estado   Estado
	FechaFin time.Time
}

// VerificarUltimoSprint comprueba las reglas que dependen del ultimo sprint del
// proyecto, con ese sprint leido con el proyecto bloqueado. ultimo es nil si el
// proyecto todavia no tiene sprints: entonces no hay nada con que comparar.
//
//   - Si el ultimo esta abierto, devuelve solo ErrUltimoSprintAbierto, sin comparar
//     fechas (RN-008-17).
//   - Si esta cerrado y el nuevo no empieza despues de su fin, devuelve
//     ErrInicioNoPosteriorAlUltimo (RN-008-13).
//
// Nunca vienen juntos.
func VerificarUltimoSprint(inicio time.Time, ultimo *UltimoSprint) error {
	if ultimo == nil {
		return nil
	}
	if ultimo.Estado == EstadoAbierto {
		return fmt.Errorf("%w (%s)", ErrUltimoSprintAbierto, nombreDe(ultimo.Numero))
	}
	if !inicio.After(ultimo.FechaFin) {
		return fmt.Errorf("%w (%s termina el %s)", ErrInicioNoPosteriorAlUltimo,
			nombreDe(ultimo.Numero), ultimo.FechaFin.Format(time.DateOnly))
	}
	return nil
}
