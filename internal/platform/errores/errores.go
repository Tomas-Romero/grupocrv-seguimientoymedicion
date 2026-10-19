// Package errores clasifica los errores de la aplicacion para poder responderlos igual
// en todas las pantallas (spec T-005).
//
// Conoce los tipos de error pero no los errores concretos: no importa el dominio ni la
// aplicacion. La lista que une un error concreto con su tipo y su campo (el catalogo)
// se arma en el adaptador HTTP, que si conoce a los dos.
package errores

import (
	"errors"
	"strings"
)

// Tipo dice que clase de problema es un error. Cada tipo tiene su estado HTTP, que
// decide el adaptador HTTP (RN-T005-1).
type Tipo string

const (
	// Validacion: depende solo de los valores recibidos, o del reloj (RN-T005-2).
	Validacion Tipo = "validacion"
	// NoEncontrado: apunta a algo que no existe.
	NoEncontrado Tipo = "no_encontrado"
	// Conflicto: depende del estado que ya hay guardado.
	Conflicto Tipo = "conflicto"
	// Interno: todo lo demas. Nunca se muestra su detalle.
	Interno Tipo = "interno"
)

// MensajeInterno es lo unico que ve la persona cuando el error es interno (RN-T005-7).
const MensajeInterno = "ocurrio un error inesperado: proba de nuevo en unos minutos"

// ErrorDeCampo es un mensaje para la persona, con el campo del formulario al que
// pertenece. Un campo vacio es un error general.
type ErrorDeCampo struct {
	Campo   string
	Mensaje string
}

// Problema es el resultado de clasificar un error. Con Tipo vacio no hay problema.
type Problema struct {
	Tipo    Tipo
	Errores []ErrorDeCampo
}

// Regla une un error concreto con su tipo y con el campo al que pertenece (RN-T005-8).
type Regla struct {
	Err  error
	Tipo Tipo
	// Campo es el nombre del campo del formulario, o vacio si es un error general.
	Campo string
}

// Catalogo es la lista de errores conocidos. Se arma una vez al iniciar y no se modifica
// despues, asi que se puede usar desde varias goroutines (spec T-005, seccion 5).
type Catalogo struct {
	reglas []Regla
}

// NuevoCatalogo arma el catalogo. Si un error aparece mas de una vez, gana la primera regla.
func NuevoCatalogo(reglas ...Regla) *Catalogo {
	return &Catalogo{reglas: append([]Regla(nil), reglas...)}
}

// ErrorDeValidacion agrupa los campos de un formulario con un formato invalido
// (RN-T005-9). Se clasifica como validacion y conserva sus campos y mensajes.
type ErrorDeValidacion struct {
	Errores []ErrorDeCampo
}

// Error devuelve los campos mal formados, separados con punto y coma.
func (e *ErrorDeValidacion) Error() string {
	partes := make([]string, len(e.Errores))
	for i, c := range e.Errores {
		partes[i] = c.Campo + ": " + c.Mensaje
	}
	return strings.Join(partes, "; ")
}

// gravedad ordena los tipos de menor a mayor (RN-T005-6).
func gravedad(t Tipo) int {
	switch t {
	case Validacion:
		return 1
	case Conflicto:
		return 2
	case NoEncontrado:
		return 3
	case Interno:
		return 4
	default:
		return 0
	}
}

// Clasificar separa el error en sus partes (RN-T005-4), clasifica cada una y arma el
// Problema. Con nil devuelve un Problema vacio.
func (c *Catalogo) Clasificar(err error) Problema {
	if err == nil {
		return Problema{}
	}

	var problema Problema
	for _, parte := range partes(err) {
		tipo, errs := c.clasificarParte(parte)
		if gravedad(tipo) > gravedad(problema.Tipo) {
			problema.Tipo = tipo
		}
		problema.Errores = append(problema.Errores, errs...)
	}

	// RN-T005-6: con una parte interna no se muestra ninguna de las otras.
	if problema.Tipo == Interno {
		problema.Errores = []ErrorDeCampo{{Mensaje: MensajeInterno}}
	}
	return problema
}

// clasificarParte clasifica un error que ya no es una union.
func (c *Catalogo) clasificarParte(err error) (Tipo, []ErrorDeCampo) {
	var deValidacion *ErrorDeValidacion
	if errors.As(err, &deValidacion) {
		return Validacion, append([]ErrorDeCampo(nil), deValidacion.Errores...)
	}
	for _, r := range c.reglas {
		if errors.Is(err, r.Err) {
			return r.Tipo, []ErrorDeCampo{{Campo: r.Campo, Mensaje: r.Err.Error()}}
		}
	}
	return Interno, nil
}

// partes aplana las uniones (errors.Join o varios %w), tambien las anidadas, en el orden
// de lectura. Un error que no es una union se devuelve tal cual, con sus capas.
func partes(err error) []error {
	if union, ok := err.(interface{ Unwrap() []error }); ok {
		var todas []error
		for _, e := range union.Unwrap() {
			todas = append(todas, partes(e)...)
		}
		return todas
	}
	// Una capa de contexto (fmt.Errorf con un solo %w) puede envolver una union: se
	// baja hasta encontrarla. Si no hay union debajo, la capa entera es la parte.
	if envuelto := errors.Unwrap(err); envuelto != nil {
		if debajo := partes(envuelto); len(debajo) > 1 {
			return debajo
		}
	}
	return []error{err}
}
