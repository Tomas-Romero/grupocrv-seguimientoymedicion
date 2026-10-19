package sprint

import (
	"errors"
	"fmt"
)

// ErrItemFueraDelSprint: el item no esta en el sprint, es decir, no tiene una fila
// pendiente en el (RN-010-2). Haber estado en otro sprint no cuenta (spec US-010,
// seccion 7).
var ErrItemFueraDelSprint = errors.New("la historia no esta asignada a este sprint")

// VerificarMarcado comprueba las reglas de marcar o desmarcar una historia (RN-010-2 y
// RN-010-3): el sprint esta abierto y el item esta en el. enElSprint dice si el item
// tiene una fila pendiente en ese sprint (RN-009-13).
//
// Con el sprint cerrado devuelve solo ErrSprintCerrado, envuelto con el nombre del
// sprint, sin mirar si el item esta en el (RN-010-4): al cerrar ninguna fila queda
// pendiente, asi que el otro error saldria tambien para historias que si estuvieron.
// Con el sprint abierto, el unico error posible es ErrItemFueraDelSprint.
//
// La usa el caso de uso con los datos que leyo y la repite el repositorio con el
// sprint bloqueado (RN-010-9). Que el item y el sprint existan no se verifica aca.
func VerificarMarcado(s DatosDelSprint, enElSprint bool) error {
	if s.Estado == EstadoCerrado {
		return fmt.Errorf("%w (%s)", ErrSprintCerrado, s.Nombre)
	}
	if !enElSprint {
		return ErrItemFueraDelSprint
	}
	return nil
}
