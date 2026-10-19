package sprint_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// sprintCerrado es un sprint cerrado del proyecto "p1".
func sprintCerrado() sprint.DatosDelSprint {
	return sprint.DatosDelSprint{ProyectoID: "p1", Nombre: "Sprint 1", Estado: sprint.EstadoCerrado}
}

// US-010 / CA-010-1, CA-010-2, RN-010-2, RN-010-3: con el sprint abierto y el item en
// el (una fila pendiente), se puede marcar y desmarcar.
func TestVerificarMarcado_Valido(t *testing.T) {
	if err := sprint.VerificarMarcado(sprintAbierto(), true); err != nil {
		t.Errorf("error = %v, se esperaba que se pudiera marcar", err)
	}
}

// US-010 / CA-010-4, RN-010-2, CL-010-5: un item que no esta en el sprint (no tiene
// una fila pendiente en el) se rechaza con ErrItemFueraDelSprint.
func TestVerificarMarcado_ItemFueraDelSprint(t *testing.T) {
	err := sprint.VerificarMarcado(sprintAbierto(), false)

	if !errors.Is(err, sprint.ErrItemFueraDelSprint) {
		t.Fatalf("error = %v, se esperaba ErrItemFueraDelSprint", err)
	}
	if errors.Is(err, sprint.ErrSprintCerrado) {
		t.Errorf("el error %q reporta el sprint cerrado, pero esta abierto", err)
	}
}

// US-010 / CA-010-5, RN-010-3, RN-010-4, CL-010-6, CL-010-7: con el sprint cerrado no
// se marca ni se desmarca, y el unico error es ErrSprintCerrado, envuelto con el
// nombre del sprint: este o no el item en el sprint.
func TestVerificarMarcado_SprintCerrado(t *testing.T) {
	for _, enElSprint := range []bool{true, false} {
		err := sprint.VerificarMarcado(sprintCerrado(), enElSprint)

		if !errors.Is(err, sprint.ErrSprintCerrado) {
			t.Fatalf("en el sprint %t: error = %v, se esperaba ErrSprintCerrado", enElSprint, err)
		}
		if errors.Is(err, sprint.ErrItemFueraDelSprint) {
			t.Errorf("en el sprint %t: el error %q tambien reporta el item fuera del sprint: con el sprint cerrado va solo ErrSprintCerrado", enElSprint, err)
		}
		if !strings.Contains(err.Error(), "Sprint 1") {
			t.Errorf("en el sprint %t: el error %q no indica el nombre del sprint", enElSprint, err)
		}
	}
}

// US-010 / spec seccion 7: el mensaje que ve la persona cuando la historia no esta en
// el sprint.
func TestErrItemFueraDelSprint_Mensaje(t *testing.T) {
	const esperado = "la historia no esta asignada a este sprint"

	if sprint.ErrItemFueraDelSprint.Error() != esperado {
		t.Errorf("mensaje = %q, se esperaba %q", sprint.ErrItemFueraDelSprint.Error(), esperado)
	}
}
