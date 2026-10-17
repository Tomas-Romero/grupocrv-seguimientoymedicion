// Package integrante contiene las reglas de negocio de un integrante de un
// proyecto: su nombre, su email y su rol.
//
// Es dominio puro (CONTRIBUTING.md, seccion 1): no importa nada fuera de la
// biblioteca estandar, no accede a la base, no genera IDs y no lee el reloj.
package integrante

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"
)

// MaxLargoNombre es el largo maximo del nombre, en caracteres (RN-003-2).
const MaxLargoNombre = 100

// Errores del paquete. Son valores para compararlos con errors.Is, tambien
// cuando vienen unidos con errors.Join (RN-003-8). Los mensajes son los de la
// seccion 7 de la spec.
var (
	ErrNombreVacio   = errors.New("el nombre es obligatorio")
	ErrNombreLargo   = errors.New("el nombre supera el largo maximo (100 caracteres)")
	ErrEmailVacio    = errors.New("el email es obligatorio")
	ErrEmailInvalido = errors.New("el email no tiene un formato valido")
	ErrRolInvalido   = errors.New("el rol tiene que ser product_architect, agile_enabler o product_builder")
)

// Integrante es un integrante de un proyecto ya validado.
type Integrante struct {
	// ID esta vacio hasta que la persistencia lo asigna (RN-003-10).
	ID         string
	ProyectoID string
	Nombre     string
	Email      string
	Rol        Rol
}

// Nuevo valida los datos y arma el Integrante, sin ID.
//
// RN-003-8: valida todo y devuelve todos los errores unidos con errors.Join,
// en orden fijo: nombre, email, rol. Si hay algun error no devuelve un
// Integrante. ErrEmailInvalido y ErrRolInvalido van envueltos con el valor
// recibido; errors.Is los sigue reconociendo.
//
// proyectoID se copia tal cual: que el proyecto exista lo verifica la capa de
// aplicacion (RN-003-6), nunca el dominio. El rol no se normaliza (CL-003-10).
func Nuevo(proyectoID, nombre, email string, rol Rol) (Integrante, error) {
	nombre = strings.TrimSpace(nombre)
	email = strings.TrimSpace(email)

	var errs []error

	// RN-003-1 y RN-003-2: el limite se mide en caracteres, despues de recortar.
	switch {
	case nombre == "":
		errs = append(errs, ErrNombreVacio)
	case utf8.RuneCountInString(nombre) > MaxLargoNombre:
		errs = append(errs, ErrNombreLargo)
	}

	// RN-003-3: un email con solo espacios queda vacio despues de recortar.
	switch {
	case email == "":
		errs = append(errs, ErrEmailVacio)
	case !esDireccionSimple(email):
		errs = append(errs, fmt.Errorf("%w: %q", ErrEmailInvalido, email))
	}

	// RN-003-5: el rol tiene que ser uno de los tres valores exactos.
	if !rol.valido() {
		errs = append(errs, fmt.Errorf("%w: %q", ErrRolInvalido, string(rol)))
	}

	if err := errors.Join(errs...); err != nil {
		return Integrante{}, err
	}

	// RN-003-4: el email se guarda en minuscula, despues de validarlo.
	return Integrante{
		ProyectoID: proyectoID,
		Nombre:     nombre,
		Email:      strings.ToLower(email),
		Rol:        rol,
	}, nil
}

// esDireccionSimple dice si email es una sola direccion `usuario@dominio`, sin
// nombre visible ni corchetes (CL-003-9). net/mail acepta formas mas ricas
// ("Tomas <tomas@ejemplo.test>", "<tomas@ejemplo.test>"), asi que ademas de que
// parsee exige que la direccion leida sea identica al texto recibido.
func esDireccionSimple(email string) bool {
	direccion, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return direccion.Name == "" && direccion.Address == email
}
