// Package backlog contiene las reglas del Product Backlog: los items, su
// prioridad MoSCoW, su estado y sus criterios de aceptacion (US-005).
//
// Es dominio puro: no importa nada fuera de la biblioteca estandar, no genera
// IDs y no lee el reloj. Que el proyecto exista, el ID y el numero del item los
// resuelven la aplicacion y la persistencia.
package backlog

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

// LargoMaximoTitulo es el maximo de caracteres del titulo recortado (RN-005-3).
const LargoMaximoTitulo = 120

// DatosItem son los datos que carga la persona al crear un item.
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
}

// NuevoItem valida los datos y arma un item pendiente y sin estimar.
func NuevoItem(datos DatosItem, creadoEn time.Time) (ItemBacklog, error) {
	// RN-005-1: se recorta antes de validar, y se guarda recortado.
	titulo := strings.TrimSpace(datos.Titulo)
	if titulo == "" {
		return ItemBacklog{}, ErrTituloVacio
	}
	// Se cuentan runas y no bytes: una tilde o una ñ es un caracter (CL-005-3).
	if n := utf8.RuneCountInString(titulo); n > LargoMaximoTitulo {
		return ItemBacklog{}, fmt.Errorf("%w (tiene %d)", ErrTituloMuyLargo, n)
	}
	if !datos.Prioridad.valida() {
		return ItemBacklog{}, fmt.Errorf("%w: se recibio %q", ErrPrioridadInvalida, string(datos.Prioridad))
	}

	return ItemBacklog{
		Titulo: titulo,
		// RN-005-4: opcional y sin largo maximo; si queda vacia, es valida.
		Descripcion: strings.TrimSpace(datos.Descripcion),
		Prioridad:   datos.Prioridad,
		Estado:      EstadoPendiente,
		StoryPoints: SinEstimar(),
		Criterios:   datos.Criterios,
		CreadoEn:    creadoEn,
	}, nil
}
