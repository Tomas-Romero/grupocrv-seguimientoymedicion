package sprint_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// itemListo es un item que cumple todo para asignarse: del proyecto "p1", sin
// completar, estimado, con un criterio y fuera de cualquier sprint abierto.
func itemListo() sprint.DatosDelItem {
	return sprint.DatosDelItem{ProyectoID: "p1", Estimado: true, Criterios: 1}
}

// sprintAbierto es un sprint abierto del proyecto "p1".
func sprintAbierto() sprint.DatosDelSprint {
	return sprint.DatosDelSprint{ProyectoID: "p1", Nombre: "Sprint 2", Estado: sprint.EstadoAbierto}
}

// US-009 / CA-009-1, CA-009-3, CL-009-1, CL-009-6: un item del mismo proyecto,
// estimado y con al menos un criterio entra a un sprint abierto. Con un solo
// criterio y un solo Story Point ya cumple la Definition of Ready.
func TestVerificarAsignacion_Valida(t *testing.T) {
	casos := []struct {
		nombre string
		item   sprint.DatosDelItem
	}{
		{"un criterio, estimado", itemListo()},
		{"varios criterios", sprint.DatosDelItem{ProyectoID: "p1", Estimado: true, Criterios: 5}},
		{"item en progreso (no completado)", itemListo()},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if err := sprint.VerificarAsignacion(c.item, sprintAbierto()); err != nil {
				t.Errorf("error = %v, se esperaba una asignacion valida", err)
			}
		})
	}
}

// US-009 / RN-009-3, CL-009-9: item y sprint de proyectos distintos se rechazan.
func TestVerificarAsignacion_ProyectoDistinto(t *testing.T) {
	item := itemListo()
	item.ProyectoID = "otro"

	err := sprint.VerificarAsignacion(item, sprintAbierto())

	if !errors.Is(err, sprint.ErrProyectoDistinto) {
		t.Errorf("error = %v, se esperaba ErrProyectoDistinto", err)
	}
}

// US-009 / RN-009-4, CA-009-6: solo se asigna a un sprint abierto; el error indica
// el nombre del sprint.
func TestVerificarAsignacion_SprintCerrado(t *testing.T) {
	cerrado := sprintAbierto()
	cerrado.Estado = sprint.EstadoCerrado

	err := sprint.VerificarAsignacion(itemListo(), cerrado)

	if !errors.Is(err, sprint.ErrSprintCerrado) {
		t.Fatalf("error = %v, se esperaba ErrSprintCerrado", err)
	}
	if !strings.Contains(err.Error(), "Sprint 2") {
		t.Errorf("el error %q no indica el nombre del sprint", err)
	}
}

// US-009 / RN-009-5, CL-009-8: un item completado no se asigna, con el error
// general de backlog (el mismo concepto que en US-013 y US-006).
func TestVerificarAsignacion_ItemCompletado(t *testing.T) {
	item := itemListo()
	item.Completado = true

	err := sprint.VerificarAsignacion(item, sprintAbierto())

	if !errors.Is(err, backlog.ErrItemCompletado) {
		t.Errorf("error = %v, se esperaba backlog.ErrItemCompletado", err)
	}
	if errors.Is(err, sprint.ErrItemYaAsignado) {
		t.Errorf("el error %q tambien reporta item ya asignado: no esta en el sprint abierto", err)
	}
}

// US-009 / RN-009-6, CL-009-7, CA-009-7: un item que ya esta en el sprint abierto
// del proyecto se rechaza, incluso si es el mismo sprint: no es una operacion sin
// efecto.
func TestVerificarAsignacion_ItemYaAsignado(t *testing.T) {
	item := itemListo()
	item.EnSprintAbierto = true

	err := sprint.VerificarAsignacion(item, sprintAbierto())

	if !errors.Is(err, sprint.ErrItemYaAsignado) {
		t.Errorf("error = %v, se esperaba ErrItemYaAsignado", err)
	}
}

// US-009 / RN-009-7, CL-009-3, CL-009-4, CA-009-4, CA-009-5: la Definition of Ready
// exige estar estimado y tener al menos un criterio, cada una con su error.
func TestVerificarAsignacion_DefinitionOfReady(t *testing.T) {
	casos := []struct {
		nombre   string
		item     sprint.DatosDelItem
		esperado error
		noEsper  error
	}{
		{"sin estimar, con criterios", sprint.DatosDelItem{ProyectoID: "p1", Estimado: false, Criterios: 2}, sprint.ErrItemSinEstimar, sprint.ErrItemSinCriterios},
		{"sin criterios, estimado", sprint.DatosDelItem{ProyectoID: "p1", Estimado: true, Criterios: 0}, sprint.ErrItemSinCriterios, sprint.ErrItemSinEstimar},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := sprint.VerificarAsignacion(c.item, sprintAbierto())

			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.esperado)
			}
			if errors.Is(err, c.noEsper) {
				t.Errorf("el error %q tambien reporta %v", err, c.noEsper)
			}
		})
	}
}

// US-009 / RN-009-8, CL-009-5, CL-009-13: el dominio no corta en el primer error:
// devuelve todos unidos, en el orden proyecto distinto, sprint cerrado, item
// completado, item ya asignado, sin estimar y sin criterios.
func TestVerificarAsignacion_TodosLosErroresEnOrdenFijo(t *testing.T) {
	item := sprint.DatosDelItem{ProyectoID: "otro", Completado: true, EnSprintAbierto: true, Estimado: false, Criterios: 0}
	cerrado := sprint.DatosDelSprint{ProyectoID: "p1", Nombre: "Sprint 2", Estado: sprint.EstadoCerrado}

	err := sprint.VerificarAsignacion(item, cerrado)

	esperados := []error{
		sprint.ErrProyectoDistinto, sprint.ErrSprintCerrado, backlog.ErrItemCompletado,
		sprint.ErrItemYaAsignado, sprint.ErrItemSinEstimar, sprint.ErrItemSinCriterios,
	}
	mensaje, posicion := err.Error(), -1
	for _, e := range esperados {
		if !errors.Is(err, e) {
			t.Errorf("error = %v, se esperaba que incluyera %v", err, e)
		}
		i := strings.Index(mensaje, e.Error())
		if i <= posicion {
			t.Errorf("el mensaje %q no respeta el orden de RN-009-8", mensaje)
			break
		}
		posicion = i
	}
	if lineas := strings.Count(mensaje, "\n") + 1; lineas != len(esperados) {
		t.Errorf("el mensaje tiene %d lineas, se esperaban %d (una por error)", lineas, len(esperados))
	}
}

// US-009 / CL-009-13: un sprint cerrado y ademas un item sin estimar devuelve los
// dos errores unidos.
func TestVerificarAsignacion_CerradoYSinEstimar(t *testing.T) {
	item := sprint.DatosDelItem{ProyectoID: "p1", Estimado: false, Criterios: 1}
	cerrado := sprintAbierto()
	cerrado.Estado = sprint.EstadoCerrado

	err := sprint.VerificarAsignacion(item, cerrado)

	if !errors.Is(err, sprint.ErrSprintCerrado) || !errors.Is(err, sprint.ErrItemSinEstimar) {
		t.Errorf("error = %v, se esperaban ErrSprintCerrado y ErrItemSinEstimar", err)
	}
}

// US-009 / RN-009-12: no se valida la capacidad del sprint ni las fechas: el
// dominio no recibe nada de eso y asignar nunca falla por superar la velocidad.
func TestVerificarAsignacion_NoValidaCapacidadNiFechas(t *testing.T) {
	if err := sprint.VerificarAsignacion(itemListo(), sprintAbierto()); err != nil {
		t.Errorf("error = %v, se esperaba una asignacion valida", err)
	}
}
