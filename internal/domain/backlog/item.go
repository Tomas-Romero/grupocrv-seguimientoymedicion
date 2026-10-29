// Package backlog contiene las reglas del Product Backlog: los items, su
// prioridad MoSCoW, su estado y sus criterios de aceptacion (US-005).
//
// Es dominio puro: no importa nada fuera de la biblioteca estandar, no genera
// IDs y no lee el reloj. Que el proyecto exista, el ID y el numero del item los
// resuelven la aplicacion y la persistencia.
package backlog

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// LargoMaximoTitulo es el maximo de caracteres del titulo recortado (RN-005-3).
const LargoMaximoTitulo = 120

// DatosItem son los datos que carga la persona al crear o editar un item.
type DatosItem struct {
	Titulo      string
	Descripcion string
	Prioridad   Prioridad
	Criterios   []string
}

// ItemBacklog es un item del Product Backlog de un proyecto.
type ItemBacklog struct {
	// ID lo genera la persistencia: esta vacio hasta que el item se registra.
	ID string
	// Numero es el correlativo dentro del proyecto (#1, #2, ...); lo asigna la
	// persistencia.
	Numero int
	// ProyectoID lo completa la persistencia al registrar el item.
	ProyectoID  string
	Titulo      string
	Descripcion string
	Prioridad   Prioridad
	Estado      Estado
	StoryPoints StoryPoints
	Criterios   []string
	CreadoEn    time.Time
	// CompletadoEn es cuando se marco como completado (US-010). El valor cero es "sin
	// fecha": solo un item completado la tiene.
	CompletadoEn time.Time
}

// NuevoItem valida los datos y arma un item pendiente y sin estimar. Si algun
// dato es invalido no devuelve item: devuelve todos los errores unidos (ver
// validarDatos).
func NuevoItem(datos DatosItem, creadoEn time.Time) (ItemBacklog, error) {
	validos, err := validarDatos(datos)
	if err != nil {
		return ItemBacklog{}, err
	}
	return ItemBacklog{
		Titulo:      validos.Titulo,
		Descripcion: validos.Descripcion,
		Prioridad:   validos.Prioridad,
		Estado:      EstadoPendiente,
		StoryPoints: SinEstimar(),
		Criterios:   validos.Criterios,
		CreadoEn:    creadoEn,
	}, nil
}

// validarDatos valida los datos que carga la persona y los devuelve recortados.
// Es la unica validacion de los datos de un item: la usan crear (US-005) y
// editar (US-006, RN-006-1). Si algun dato es invalido devuelve todos los
// errores unidos con errors.Join, en el orden de RN-005-13 (titulo, prioridad y
// criterios por posicion), para que quien carga el item vea de una vez todo lo
// que tiene que corregir.
func validarDatos(datos DatosItem) (DatosItem, error) {
	var errs []error

	// RN-005-1: se recorta antes de validar, y se guarda recortado.
	titulo := strings.TrimSpace(datos.Titulo)
	if err := validarTitulo(titulo); err != nil {
		errs = append(errs, err)
	}
	// RN-007-8: la regla de la prioridad es la misma que usa el filtro del backlog.
	if err := ValidarPrioridad(datos.Prioridad); err != nil {
		errs = append(errs, err)
	}
	criterios, errsCriterios := recortarCriterios(datos.Criterios)
	errs = append(errs, errsCriterios...)

	if err := errors.Join(errs...); err != nil {
		return DatosItem{}, err
	}
	return DatosItem{
		Titulo: titulo,
		// RN-005-4: opcional y sin largo maximo; si queda vacia, es valida.
		Descripcion: strings.TrimSpace(datos.Descripcion),
		Prioridad:   datos.Prioridad,
		Criterios:   criterios,
	}, nil
}

// validarTitulo valida el titulo ya recortado (RN-005-2 y RN-005-3).
func validarTitulo(titulo string) error {
	if titulo == "" {
		return ErrTituloVacio
	}
	// Se cuentan runas y no bytes: una tilde o una ñ es un caracter (CL-005-3).
	if n := utf8.RuneCountInString(titulo); n > LargoMaximoTitulo {
		return fmt.Errorf("%w (tiene %d)", ErrTituloMuyLargo, n)
	}
	return nil
}

// recortarCriterios devuelve una copia de los criterios recortados, en el mismo
// orden, y un CriterioVacioError por cada uno que quede vacio (RN-005-8).
func recortarCriterios(criterios []string) ([]string, []error) {
	recortados := make([]string, len(criterios))
	var errs []error
	for i, criterio := range criterios {
		recortados[i] = strings.TrimSpace(criterio)
		if recortados[i] == "" {
			errs = append(errs, CriterioVacioError{Posicion: i + 1})
		}
	}
	return recortados, errs
}
