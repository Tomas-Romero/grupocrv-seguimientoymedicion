// Package esfuerzo contiene las reglas de un registro de esfuerzo: quien trabajo,
// en que historia, que dia, en que actividad y cuantas horas (US-019).
//
// Es dominio puro: no importa nada fuera de la biblioteca estandar, no accede a la
// base, no genera IDs y no lee el reloj. Que el proyecto, la historia y el
// integrante existan y se correspondan lo verifica la capa de aplicacion.
package esfuerzo

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"
)

// Limites de un registro (RN-019-1, RN-019-2 y RN-019-5).
const (
	// MaxHoras es el maximo de horas de un solo registro.
	MaxHoras = 24
	// MaxLargoActividad es el maximo de caracteres de la actividad recortada.
	MaxLargoActividad = 120
	// pasoDeHoras: las horas se registran en multiplos de 0,25 (15 minutos). Es
	// una potencia de dos, asi que sus multiplos se representan exactos en
	// float64 y la comprobacion no tiene errores de redondeo.
	pasoDeHoras = 0.25
)

// Errores del paquete (spec US-019, seccion 7). Son valores para compararlos con
// errors.Is, tambien cuando vienen unidos con errors.Join (RN-019-8).
var (
	// ErrFechaVacia: la fecha es el valor cero (RN-019-4).
	ErrFechaVacia = errors.New("la fecha es obligatoria")
	// ErrFechaFutura: la fecha es posterior a hoy (RN-019-3). Se devuelve
	// envuelto, indicando la fecha recibida.
	ErrFechaFutura = errors.New("la fecha no puede ser posterior a hoy")
	// ErrActividadVacia: la actividad esta vacia o tiene solo espacios (RN-019-5).
	ErrActividadVacia = errors.New("la actividad es obligatoria")
	// ErrActividadMuyLarga: la actividad recortada pasa los 120 caracteres
	// (RN-019-5). Se devuelve envuelto, indicando cuantos tiene.
	ErrActividadMuyLarga = errors.New("la actividad no puede tener mas de 120 caracteres")
	// ErrHorasFueraDeRango: las horas no son mayores a 0 y como mucho 24, o no son
	// un numero (RN-019-1). Se devuelve envuelto, indicando el valor recibido.
	ErrHorasFueraDeRango = errors.New("las horas tienen que ser mayores a 0 y no pasar de 24")
	// ErrHorasFraccion: las horas no son multiplo de 0,25 (RN-019-2). Se devuelve
	// envuelto, indicando el valor recibido.
	ErrHorasFraccion = errors.New("las horas se registran en multiplos de 0,25")
)

// RegistroEsfuerzo son las horas que un integrante dedico a una historia un dia.
type RegistroEsfuerzo struct {
	// ID esta vacio hasta que la persistencia lo asigna (RN-019-10).
	ID           string
	ProyectoID   string
	ItemID       string
	IntegranteID string
	// Fecha es el dia del trabajo, sin hora (UTC).
	Fecha     time.Time
	Actividad string
	Horas     float64
}

// Nuevo valida los datos y arma el registro, sin ID. hoy es la fecha actual: se
// recibe por parametro porque el dominio no lee el reloj.
//
// RN-019-8: valida todo y devuelve todos los errores unidos con errors.Join, en
// orden fijo: fecha, actividad, horas. De las horas se devuelve un solo error: el
// de rango tiene prioridad sobre el de fraccion. Si hay algun error no devuelve
// registro.
//
// Los IDs se copian tal cual: que existan y sean del proyecto lo verifica la capa
// de aplicacion (RN-019-6), nunca el dominio.
func Nuevo(proyectoID, itemID, integranteID string, fecha time.Time, actividad string, horas float64, hoy time.Time) (RegistroEsfuerzo, error) {
	var errs []error

	// RN-019-3 y RN-019-4: se comparan solo el anio, el mes y el dia (en UTC).
	dia := soloElDia(fecha)
	switch {
	case fecha.IsZero():
		errs = append(errs, ErrFechaVacia)
	case dia.After(soloElDia(hoy)):
		errs = append(errs, fmt.Errorf("%w (se recibio %s)", ErrFechaFutura, dia.Format(time.DateOnly)))
	}

	// RN-019-5: se recorta antes de validar y se guarda recortada. El limite se
	// mide en caracteres y no en bytes.
	actividad = strings.TrimSpace(actividad)
	switch largo := utf8.RuneCountInString(actividad); {
	case largo == 0:
		errs = append(errs, ErrActividadVacia)
	case largo > MaxLargoActividad:
		errs = append(errs, fmt.Errorf("%w (tiene %d)", ErrActividadMuyLarga, largo))
	}

	// RN-019-1 y RN-019-2. La negacion de la comparacion rechaza tambien NaN.
	switch {
	case !(horas > 0 && horas <= MaxHoras):
		errs = append(errs, fmt.Errorf("%w (se recibio %v)", ErrHorasFueraDeRango, horas))
	case !esMultiploDelPaso(horas):
		errs = append(errs, fmt.Errorf("%w (se recibio %v)", ErrHorasFraccion, horas))
	}

	if err := errors.Join(errs...); err != nil {
		return RegistroEsfuerzo{}, err
	}
	return RegistroEsfuerzo{
		ProyectoID:   proyectoID,
		ItemID:       itemID,
		IntegranteID: integranteID,
		Fecha:        dia,
		Actividad:    actividad,
		Horas:        horas,
	}, nil
}

// soloElDia devuelve el dia de t a medianoche UTC.
func soloElDia(t time.Time) time.Time {
	anio, mes, dia := t.UTC().Date()
	return time.Date(anio, mes, dia, 0, 0, 0, 0, time.UTC)
}

// esMultiploDelPaso dice si horas es un multiplo exacto de 0,25.
func esMultiploDelPaso(horas float64) bool {
	pasos := horas / pasoDeHoras
	return pasos == math.Trunc(pasos)
}
