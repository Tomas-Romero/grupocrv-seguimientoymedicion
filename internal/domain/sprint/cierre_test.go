package sprint_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// US-011 / CA-011-1, RN-011-5, RN-011-6, CL-011-4: cada historia del sprint queda con
// su resultado, en el mismo orden: completado si estaba completada y no_completado si
// no (pendiente o en progreso, que el dominio recibe igual: no completada).
func TestCerrar(t *testing.T) {
	historias := []sprint.HistoriaDelCierre{
		{ItemID: "h1", Completada: true},
		{ItemID: "h2", Completada: false},
		{ItemID: "h3", Completada: true},
	}

	resultados, err := sprint.Cerrar(sprintAbierto(), historias)
	if err != nil {
		t.Fatalf("Cerrar: %v", err)
	}

	esperados := []sprint.ResultadoDelCierre{
		{ItemID: "h1", Resultado: sprint.ResultadoCompletado},
		{ItemID: "h2", Resultado: sprint.ResultadoNoCompletado},
		{ItemID: "h3", Resultado: sprint.ResultadoCompletado},
	}
	if !slices.Equal(resultados, esperados) {
		t.Errorf("resultados = %+v, se esperaban %+v", resultados, esperados)
	}
}

// US-011 / CL-011-2, CL-011-3: todas completadas o ninguna completada.
func TestCerrar_TodasONinguna(t *testing.T) {
	for _, completadas := range []bool{true, false} {
		historias := []sprint.HistoriaDelCierre{{ItemID: "h1", Completada: completadas}, {ItemID: "h2", Completada: completadas}}
		esperado := sprint.ResultadoNoCompletado
		if completadas {
			esperado = sprint.ResultadoCompletado
		}

		resultados, err := sprint.Cerrar(sprintAbierto(), historias)
		if err != nil {
			t.Fatalf("Cerrar(completadas %t): %v", completadas, err)
		}
		for _, r := range resultados {
			if r.Resultado != esperado {
				t.Errorf("completadas %t: %s quedo %q, se esperaba %q", completadas, r.ItemID, r.Resultado, esperado)
			}
		}
	}
}

// US-011 / CA-011-4, RN-011-4, CL-011-1: un sprint sin historias se cierra y no tiene
// resultados.
func TestCerrar_SinHistorias(t *testing.T) {
	resultados, err := sprint.Cerrar(sprintAbierto(), nil)
	if err != nil {
		t.Fatalf("Cerrar: %v", err)
	}
	if len(resultados) != 0 {
		t.Errorf("resultados = %+v, se esperaba ninguno", resultados)
	}
}

// US-011 / CA-011-5, RN-011-2, CL-011-6: un sprint ya cerrado no se vuelve a cerrar:
// ErrSprintCerrado con el nombre del sprint, y ningun resultado.
func TestCerrar_SprintCerrado(t *testing.T) {
	resultados, err := sprint.Cerrar(sprintCerrado(), []sprint.HistoriaDelCierre{{ItemID: "h1", Completada: true}})

	if !errors.Is(err, sprint.ErrSprintCerrado) {
		t.Fatalf("error = %v, se esperaba ErrSprintCerrado", err)
	}
	if !strings.Contains(err.Error(), "Sprint 1") {
		t.Errorf("el error %q no indica el nombre del sprint", err)
	}
	if resultados != nil {
		t.Errorf("resultados = %+v, se esperaba ninguno", resultados)
	}
}

// Los resultados son los valores que guarda la relacion de US-009 (RN-009-13).
func TestResultados_Valores(t *testing.T) {
	for resultado, esperado := range map[sprint.Resultado]string{
		sprint.ResultadoPendiente:    "pendiente",
		sprint.ResultadoCompletado:   "completado",
		sprint.ResultadoNoCompletado: "no_completado",
	} {
		if string(resultado) != esperado {
			t.Errorf("resultado = %q, se esperaba %q", resultado, esperado)
		}
	}
}
