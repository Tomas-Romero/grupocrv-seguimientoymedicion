// Package proyecto contiene las reglas de negocio de un proyecto: su nombre,
// su descripcion y sus fechas de inicio y fin.
//
// Es dominio puro (CONTRIBUTING.md, seccion 1): no importa nada fuera de la
// biblioteca estandar, no accede a la base, no genera IDs y no lee el reloj.
package proyecto

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// MaxLargoNombre es el largo maximo del nombre, en caracteres (RN-001-2).
const MaxLargoNombre = 100

// Errores del paquete. Son valores para compararlos con errors.Is, tambien
// cuando vienen unidos con errors.Join (RN-001-7). Los mensajes son los de la
// seccion 7 de la spec.
var (
	ErrNombreVacio         = errors.New("el nombre es obligatorio")
	ErrNombreLargo         = errors.New("el nombre supera el largo maximo (100 caracteres)")
	ErrFechaInicioFaltante = errors.New("falta la fecha de inicio")
	ErrFechaFinFaltante    = errors.New("falta la fecha de fin")
	ErrFechasIncoherentes  = errors.New("fecha de fin anterior a la de inicio")
)

// Proyecto es un proyecto de software ya validado.
type Proyecto struct {
	// ID esta vacio hasta que la persistencia lo asigna (RN-001-6).
	ID          string
	Nombre      string
	Descripcion string
	FechaInicio time.Time
	FechaFin    time.Time
}

// Nuevo valida los datos y arma el Proyecto, sin ID.
//
// RN-001-7: valida todo y devuelve todos los errores unidos con errors.Join,
// en orden fijo: nombre, falta inicio, falta fin, fechas incoherentes. Si hay
// algun error no devuelve un Proyecto.
func Nuevo(nombre, descripcion string, fechaInicio, fechaFin time.Time) (Proyecto, error) {
	nombre = strings.TrimSpace(nombre)

	var errs []error

	// RN-001-1 y RN-001-2: el limite se mide en caracteres, despues de recortar.
	switch {
	case nombre == "":
		errs = append(errs, ErrNombreVacio)
	case utf8.RuneCountInString(nombre) > MaxLargoNombre:
		errs = append(errs, ErrNombreLargo)
	}

	// RN-001-3: las dos fechas son obligatorias.
	if fechaInicio.IsZero() {
		errs = append(errs, ErrFechaInicioFaltante)
	}
	if fechaFin.IsZero() {
		errs = append(errs, ErrFechaFinFaltante)
	}

	// RN-001-4 y CL-001-8: solo se compara si estan las dos. Iguales es valido.
	if !fechaInicio.IsZero() && !fechaFin.IsZero() && fechaFin.Before(fechaInicio) {
		errs = append(errs, ErrFechasIncoherentes)
	}

	if err := errors.Join(errs...); err != nil {
		return Proyecto{}, err
	}

	// RN-001-5: la descripcion se guarda tal cual, sin recortar.
	return Proyecto{
		Nombre:      nombre,
		Descripcion: descripcion,
		FechaInicio: fechaInicio,
		FechaFin:    fechaFin,
	}, nil
}
