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
// Lo comparten los casos de uso que trabajan sobre un item existente (US-006,
// US-009, US-010, US-013 y US-019), y lo devuelven tambien los repositorios cuando el item desaparece
// entre la lectura y la escritura.
var ErrItemInexistente = errors.New("el item no existe")

// ErrEmailDuplicado se devuelve cuando el email ya esta registrado en el
// proyecto (US-003, RN-003-7). Lo detecta la persistencia, no el dominio: el
// dominio no sabe que otros integrantes tiene el proyecto. En Postgres es la
// restriccion UNIQUE (proyecto_id, email); la base en memoria la replica.
//
// Los repositorios lo devuelven envuelto indicando el email, y el caso de uso lo
// vuelve a envolver con el proyecto: errors.Is lo sigue reconociendo.
var ErrEmailDuplicado = errors.New("ya hay un integrante con ese email en el proyecto")

// ErrDefectoInexistente se devuelve cuando la operacion apunta a un defecto que no existe o que
// no es del proyecto (US-023, RN-023-7). Como ErrProyectoInexistente, el defecto viene de la URL:
// es el equivalente a un 404, tiene prioridad sobre los errores de validacion y nunca se une con
// ellos. Tambien lo devuelve el repositorio si el defecto desaparece entre la lectura y el UPDATE
// (RN-023-11).
var ErrDefectoInexistente = errors.New("el defecto no existe")

// ErrDefectoModificado se devuelve cuando el defecto cambio de estado entre que el caso de uso lo
// leyo y que el repositorio quiso guardarlo (US-023, RN-023-11). No se reintenta: con el estado
// cambiado, lo pedido quiza ya no es valido.
var ErrDefectoModificado = errors.New("el defecto cambio mientras se lo editaba")
