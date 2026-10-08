// Package app contiene los casos de uso: orquestan el dominio y los
// repositorios. Hay un archivo por caso de uso, y cada uno declara la interfaz
// del repositorio que consume.
package app

import "errors"

// ErrProyectoInexistente se devuelve cuando la operacion apunta a un proyecto
// que no existe. El proyecto viene del contexto de navegacion (la URL), no de un
// formulario: es el equivalente a un 404, tiene prioridad sobre los errores de
// validacion y nunca se une con ellos.
//
// Lo comparten los casos de uso que trabajan sobre un proyecto existente
// (US-005, US-008, US-002), y lo devuelven tambien los repositorios cuando el
// proyecto desaparece dentro de la transaccion.
var ErrProyectoInexistente = errors.New("el proyecto no existe")

// ErrItemInexistente se devuelve cuando la operacion apunta a un item del
// backlog que no existe. Igual que ErrProyectoInexistente, el item viene del
// contexto de navegacion (la URL): es el equivalente a un 404, tiene prioridad
// sobre los errores de validacion y nunca se une con ellos.
//
// Lo comparten los casos de uso que trabajan sobre un item existente (US-013 y
// US-019), y lo devuelven tambien los repositorios cuando el item desaparece
// entre la lectura y la escritura.
var ErrItemInexistente = errors.New("el item no existe")
