// Package defecto contiene las reglas de un defecto: lo que se registra de una falla
// del producto, con su severidad, la historia donde se introdujo y el sprint en el que
// se detecto (US-022), y su ciclo de vida: los estados, las transiciones permitidas y el
// sprint en el que se resuelve (US-023).
//
// Es dominio puro: no importa nada fuera de la biblioteca estandar, no accede a la
// base, no genera IDs y no lee el reloj (la fecha de carga la pone la base). Que el
// proyecto, la historia y los sprints existan y se correspondan lo verifica la capa de
// aplicacion, que tambien traduce los sprints a su numero antes de llamar al dominio.
package defecto

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// MaxLargoDescripcion es el maximo de caracteres de la descripcion recortada
// (RN-022-1).
const MaxLargoDescripcion = 2000

// Errores del paquete (spec US-022, seccion 7). Son valores para compararlos con
// errors.Is.
var (
	// ErrDescripcionVacia: la descripcion esta vacia o tiene solo espacios (RN-022-1).
	ErrDescripcionVacia = errors.New("la descripcion es obligatoria")
	// ErrDescripcionMuyLarga: la descripcion recortada pasa los 2000 caracteres
	// (RN-022-1). Se devuelve envuelto, indicando cuantos tiene.
	ErrDescripcionMuyLarga = errors.New("la descripcion no puede tener mas de 2000 caracteres")
	// ErrSeveridadInvalida: la severidad no es critica, alta, media ni baja, o esta
	// vacia (RN-022-2). Se devuelve envuelto, indicando el valor recibido.
	ErrSeveridadInvalida = errors.New("la severidad tiene que ser critica, alta, media o baja")
)

// Severidad es la gravedad de un defecto (RN-022-2).
type Severidad string

// Valores de Severidad. Son los que viajan al dominio y se guardan: las etiquetas con
// tilde que ve la persona ("Critica") las convierte el adaptador HTTP.
const (
	SeveridadCritica Severidad = "critica"
	SeveridadAlta    Severidad = "alta"
	SeveridadMedia   Severidad = "media"
	SeveridadBaja    Severidad = "baja"
)

// Estado es el estado de un defecto en su ciclo de vida.
type Estado string

// Valores de Estado (RN-023-1). EstadoAbierto es el inicial: todo defecto nace abierto
// (RN-022-3). Los otros tres los suma US-023.
const (
	EstadoAbierto      Estado = "abierto"
	EstadoEnCorreccion Estado = "en_correccion"
	EstadoResuelto     Estado = "resuelto"
	EstadoDescartado   Estado = "descartado"
)

// Defecto es una falla registrada en un proyecto.
type Defecto struct {
	// ID esta vacio hasta que la persistencia lo asigna (RN-022-10).
	ID         string
	ProyectoID string
	// ItemID es la historia relacionada: donde se introdujo el defecto.
	ItemID string
	// SprintDeteccionID es el sprint donde se detecto, por su ID y no por su numero
	// (RN-022-6).
	SprintDeteccionID string
	Descripcion       string
	Severidad         Severidad
	Estado            Estado
	// SprintResolucionID esta vacio mientras el defecto no esta resuelto: solo el estado
	// resuelto lleva sprint de resolucion (RN-022-3, RN-023-3).
	SprintResolucionID string
}

// ValidarSeveridad devuelve nil si la severidad es una de las cuatro (RN-022-2, RN-024-8).
// TODO(US-024): esqueleto para el RED, todavia no valida nada.
func ValidarSeveridad(severidad Severidad) error { return nil }

// Nuevo arma el defecto, sin ID, siempre abierto y sin sprint de resolucion (RN-022-3).
//
// RN-022-4: valida todo y devuelve todos los errores unidos con errors.Join, en orden
// fijo: descripcion y despues severidad. De la descripcion se devuelve un solo error: el
// de vacia y el de muy larga no pueden darse a la vez. Si hay algun error no devuelve
// defecto.
//
// Los IDs se copian tal cual: que existan y sean del proyecto lo verifica la capa de
// aplicacion (RN-022-5), nunca el dominio.
func Nuevo(proyectoID, itemID, sprintDeteccionID, descripcion string, severidad Severidad) (Defecto, error) {
	var errs []error

	// RN-022-1: se recorta antes de validar y se guarda recortada. El limite se mide en
	// caracteres y no en bytes. Una descripcion de solo espacios queda vacia: es
	// ErrDescripcionVacia aunque tenga 3000, nunca ErrDescripcionMuyLarga.
	descripcion = strings.TrimSpace(descripcion)
	switch largo := utf8.RuneCountInString(descripcion); {
	case largo == 0:
		errs = append(errs, ErrDescripcionVacia)
	case largo > MaxLargoDescripcion:
		errs = append(errs, fmt.Errorf("%w (tiene %d)", ErrDescripcionMuyLarga, largo))
	}

	// RN-022-2: valores exactos; no se pasa a minuscula ni se recorta.
	switch severidad {
	case SeveridadCritica, SeveridadAlta, SeveridadMedia, SeveridadBaja:
	default:
		errs = append(errs, fmt.Errorf("%w (se recibio %q)", ErrSeveridadInvalida, string(severidad)))
	}

	if err := errors.Join(errs...); err != nil {
		return Defecto{}, err
	}
	return Defecto{
		ProyectoID:        proyectoID,
		ItemID:            itemID,
		SprintDeteccionID: sprintDeteccionID,
		Descripcion:       descripcion,
		Severidad:         severidad,
		Estado:            EstadoAbierto,
	}, nil
}
