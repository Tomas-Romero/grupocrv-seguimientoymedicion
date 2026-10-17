// Package integrante contiene las reglas de negocio de un integrante de un
// proyecto: su nombre, su email y su rol.
//
// Es dominio puro (CONTRIBUTING.md, seccion 1): no importa nada fuera de la
// biblioteca estandar, no accede a la base, no genera IDs y no lee el reloj.
package integrante

import "errors"

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
// ESQUELETO DEL CICLO RED (US-003): todavia no valida nada. La implementacion
// real llega en el commit GREEN.
func Nuevo(_, _, _ string, _ Rol) (Integrante, error) {
	return Integrante{}, nil
}
