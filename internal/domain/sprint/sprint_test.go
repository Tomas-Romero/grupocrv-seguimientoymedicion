package sprint_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

func dia(anio int, mes time.Month, d int) time.Time {
	return time.Date(anio, mes, d, 0, 0, 0, 0, time.UTC)
}

// El proyecto de casi todos los tests: del 2026-10-06 al 2026-12-31.
var (
	inicioProyecto = dia(2026, 10, 6)
	finProyecto    = dia(2026, 12, 31)
)

// nuevo crea un sprint en el proyecto de los tests.
func nuevo(goal string, inicio, fin time.Time) (sprint.Sprint, error) {
	return sprint.Nuevo("proyecto-1", goal, inicio, fin, inicioProyecto, finProyecto)
}

// US-008 / CA-008-1, RN-008-1, RN-008-8, RN-008-12: un sprint valido queda abierto,
// con el Sprint Goal recortado y las fechas recibidas. Todavia no tiene ID, numero
// ni nombre: los asigna la persistencia.
func TestNuevo_DatosValidos(t *testing.T) {
	s, err := nuevo("  MVP navegable  ", dia(2026, 10, 6), dia(2026, 10, 19))
	if err != nil {
		t.Fatalf("Nuevo: %v", err)
	}

	if s.SprintGoal != "MVP navegable" {
		t.Errorf("SprintGoal = %q, se esperaba %q (recortado)", s.SprintGoal, "MVP navegable")
	}
	if s.ProyectoID != "proyecto-1" || !s.FechaInicio.Equal(dia(2026, 10, 6)) || !s.FechaFin.Equal(dia(2026, 10, 19)) {
		t.Errorf("sprint = %+v, se esperaban el proyecto y las fechas recibidos", s)
	}
	if s.Estado != sprint.EstadoAbierto {
		t.Errorf("Estado = %q, se esperaba %q", s.Estado, sprint.EstadoAbierto)
	}
	if s.ID != "" || s.Numero != 0 || s.Nombre() != "" {
		t.Errorf("ID, numero y nombre = %q, %d, %q; se esperaban vacios: los asigna la persistencia", s.ID, s.Numero, s.Nombre())
	}
}

// US-008 / RN-008-12: el nombre es "Sprint " seguido del numero, sin ceros a la
// izquierda, y se arma a partir del numero.
func TestSprint_Nombre(t *testing.T) {
	for numero, esperado := range map[int]string{1: "Sprint 1", 2: "Sprint 2", 10: "Sprint 10", 26: "Sprint 26"} {
		if nombre := (sprint.Sprint{Numero: numero}).Nombre(); nombre != esperado {
			t.Errorf("Nombre() con numero %d = %q, se esperaba %q", numero, nombre, esperado)
		}
	}
}

// US-008 / CA-008-3, CL-008-1, CL-008-2, CL-008-3, RN-008-6: se cuenta la duracion
// en el calendario, con los dos extremos: 1 y 28 dias se aceptan, 29 se rechaza
// con ErrDuracionExcedida indicando cuantos dias dura. Tambien en febrero.
func TestNuevo_Duracion(t *testing.T) {
	proyectoLargo := func(inicio, fin time.Time) (sprint.Sprint, error) {
		return sprint.Nuevo("proyecto-1", "Objetivo", inicio, fin, dia(2026, 1, 1), dia(2028, 12, 31))
	}
	casos := []struct {
		nombre string
		inicio time.Time
		fin    time.Time
		dias   int
		excede bool
	}{
		{"un dia", dia(2026, 10, 6), dia(2026, 10, 6), 1, false},
		{"14 dias", dia(2026, 10, 6), dia(2026, 10, 19), 14, false},
		{"28 dias", dia(2026, 10, 6), dia(2026, 11, 2), 28, false},
		{"29 dias", dia(2026, 10, 6), dia(2026, 11, 3), 29, true},
		{"febrero de 2027 (28 dias)", dia(2027, 2, 1), dia(2027, 2, 28), 28, false},
		{"febrero de 2028, bisiesto (29 dias)", dia(2028, 2, 1), dia(2028, 2, 29), 29, true},
		{"cruza el cambio de ano", dia(2026, 12, 20), dia(2027, 1, 2), 14, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := proyectoLargo(c.inicio, c.fin)

			if c.excede != errors.Is(err, sprint.ErrDuracionExcedida) {
				t.Fatalf("error = %v, ErrDuracionExcedida esperado: %t", err, c.excede)
			}
			if !c.excede && err != nil {
				t.Fatalf("error = %v, se esperaba un sprint valido de %d dias", err, c.dias)
			}
			if c.excede && !strings.Contains(err.Error(), "29") {
				t.Errorf("el error %q no indica cuantos dias dura (29)", err)
			}
		})
	}
}

// US-008 / CL-008-6: un sprint con exactamente las fechas del proyecto es valido.
func TestNuevo_ExactamenteLasFechasDelProyecto(t *testing.T) {
	_, err := sprint.Nuevo("proyecto-1", "Objetivo", dia(2026, 10, 6), dia(2026, 11, 2), dia(2026, 10, 6), dia(2026, 11, 2))
	if err != nil {
		t.Errorf("error = %v, se esperaba un sprint valido", err)
	}
}

// US-008 / CA-008-8, CL-008-7, RN-008-7: empezar un dia antes del proyecto o
// terminar un dia despues se rechaza con ErrFueraDelProyecto, un solo error aunque
// las dos fechas esten afuera, e indica las fechas del proyecto.
func TestNuevo_FueraDelProyecto(t *testing.T) {
	casos := []struct {
		nombre string
		inicio time.Time
		fin    time.Time
	}{
		{"empieza un dia antes", dia(2026, 10, 5), dia(2026, 10, 12)},
		{"termina un dia despues", dia(2026, 12, 20), dia(2027, 1, 1)},
		{"las dos fechas afuera, antes", dia(2026, 9, 1), dia(2026, 9, 10)},
		{"las dos fechas afuera, rodeando al proyecto", dia(2026, 10, 5), dia(2027, 1, 1)},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := nuevo("Objetivo", c.inicio, c.fin)

			if !errors.Is(err, sprint.ErrFueraDelProyecto) {
				t.Fatalf("error = %v, se esperaba ErrFueraDelProyecto", err)
			}
			if n := strings.Count(err.Error(), sprint.ErrFueraDelProyecto.Error()); n != 1 {
				t.Errorf("ErrFueraDelProyecto aparece %d veces, se esperaba 1", n)
			}
			if !strings.Contains(err.Error(), "2026-10-06") || !strings.Contains(err.Error(), "2026-12-31") {
				t.Errorf("el error %q no indica las fechas del proyecto", err)
			}
		})
	}
}

// US-008 / CA-008-5, CL-008-7: el primer y el ultimo dia del proyecto estan dentro.
func TestNuevo_ExtremosDelProyectoEstanDentro(t *testing.T) {
	if _, err := nuevo("Objetivo", inicioProyecto, inicioProyecto); err != nil {
		t.Errorf("sprint del primer dia del proyecto: %v", err)
	}
	if _, err := nuevo("Objetivo", finProyecto, finProyecto); err != nil {
		t.Errorf("sprint del ultimo dia del proyecto: %v", err)
	}
}

// US-008 / CA-008-6, CL-008-9, CL-008-10, RN-008-1 a RN-008-3: el Sprint Goal se
// mide en caracteres despues de recortarlo: 200 se aceptan, 201 se rechazan, y
// solo espacios es vacio y no largo.
func TestNuevo_SprintGoal(t *testing.T) {
	casos := []struct {
		nombre string
		goal   string
		error  error
	}{
		{"200 caracteres", strings.Repeat("a", 200), nil},
		{"200 caracteres con espacios en los extremos", "  " + strings.Repeat("a", 200) + "  ", nil},
		{"200 caracteres no ASCII (400 bytes)", strings.Repeat("ñ", 200), nil},
		{"201 caracteres", strings.Repeat("a", 201), sprint.ErrSprintGoalMuyLargo},
		{"201 caracteres no ASCII", strings.Repeat("ñ", 201), sprint.ErrSprintGoalMuyLargo},
		{"vacio", "", sprint.ErrSprintGoalVacio},
		{"solo espacios", "   ", sprint.ErrSprintGoalVacio},
		{"solo espacios y mas de 200", strings.Repeat(" ", 250), sprint.ErrSprintGoalVacio},
		{"solo tabulaciones y saltos de linea", "\t\n\r ", sprint.ErrSprintGoalVacio},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			s, err := nuevo(c.goal, dia(2026, 10, 6), dia(2026, 10, 19))

			if c.error == nil {
				if err != nil {
					t.Fatalf("error = %v, se esperaba un Sprint Goal valido", err)
				}
				if s.SprintGoal != strings.TrimSpace(c.goal) {
					t.Errorf("SprintGoal = %q, se esperaba el recortado", s.SprintGoal)
				}
				return
			}
			if !errors.Is(err, c.error) {
				t.Fatalf("error = %v, se esperaba %v", err, c.error)
			}
		})
	}
}

// US-008 / seccion 7: el error de largo indica cuantos caracteres tiene el goal.
func TestNuevo_SprintGoalLargoIndicaElLargo(t *testing.T) {
	_, err := nuevo(strings.Repeat("a", 201), dia(2026, 10, 6), dia(2026, 10, 19))

	if err == nil || !strings.Contains(err.Error(), "201") {
		t.Errorf("error = %v, se esperaba que indicara los 201 caracteres", err)
	}
}

// US-008 / CA-008-7, CL-008-11, RN-008-4, RN-008-5: falta una fecha, o el fin es
// anterior al inicio. Si falta una, solo se informa la que falta y no se evaluan
// las comparaciones.
func TestNuevo_Fechas(t *testing.T) {
	casos := []struct {
		nombre    string
		inicio    time.Time
		fin       time.Time
		esperados []error
		noEsperan []error
	}{
		{"falta el inicio", time.Time{}, dia(2026, 10, 19), []error{sprint.ErrFechaInicioFaltante},
			[]error{sprint.ErrFechaFinFaltante, sprint.ErrFechasIncoherentes, sprint.ErrDuracionExcedida, sprint.ErrFueraDelProyecto}},
		{"falta el fin", dia(2026, 10, 6), time.Time{}, []error{sprint.ErrFechaFinFaltante},
			[]error{sprint.ErrFechaInicioFaltante, sprint.ErrFechasIncoherentes, sprint.ErrDuracionExcedida, sprint.ErrFueraDelProyecto}},
		{"faltan las dos", time.Time{}, time.Time{}, []error{sprint.ErrFechaInicioFaltante, sprint.ErrFechaFinFaltante},
			[]error{sprint.ErrFechasIncoherentes, sprint.ErrDuracionExcedida, sprint.ErrFueraDelProyecto}},
		{"el fin es anterior al inicio", dia(2026, 10, 19), dia(2026, 10, 6), []error{sprint.ErrFechasIncoherentes},
			[]error{sprint.ErrDuracionExcedida}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := nuevo("Objetivo", c.inicio, c.fin)

			for _, e := range c.esperados {
				if !errors.Is(err, e) {
					t.Errorf("error = %v, se esperaba que incluyera %v", err, e)
				}
			}
			for _, e := range c.noEsperan {
				if errors.Is(err, e) {
					t.Errorf("error = %v, no se esperaba %v", err, e)
				}
			}
		})
	}
}

// US-008 / CL-008-1: un sprint de un dia tiene el fin igual al inicio y es valido.
func TestNuevo_FinIgualAInicio(t *testing.T) {
	if _, err := nuevo("Objetivo", dia(2026, 10, 6), dia(2026, 10, 6)); err != nil {
		t.Errorf("error = %v, se esperaba un sprint valido de un dia", err)
	}
}

// US-008 / RN-008-10, CL-008-12: con varios datos invalidos se devuelven todos
// unidos, en el orden fijo: Sprint Goal, fechas incoherentes, fuera del proyecto.
// No devuelve la duracion: con las fechas invertidas da 0 o menos.
func TestNuevo_VariosErroresEnOrdenFijo(t *testing.T) {
	_, err := sprint.Nuevo("proyecto-1", "", dia(2026, 10, 10), dia(2026, 10, 1), dia(2026, 10, 6), dia(2026, 11, 2))

	esperados := []error{sprint.ErrSprintGoalVacio, sprint.ErrFechasIncoherentes, sprint.ErrFueraDelProyecto}
	for _, e := range esperados {
		if !errors.Is(err, e) {
			t.Errorf("error = %v, se esperaba que incluyera %v", err, e)
		}
	}
	if errors.Is(err, sprint.ErrDuracionExcedida) {
		t.Errorf("error = %v, no se esperaba ErrDuracionExcedida", err)
	}
	mensaje, posicion := err.Error(), -1
	for _, e := range esperados {
		i := strings.Index(mensaje, e.Error())
		if i <= posicion {
			t.Errorf("el mensaje %q no respeta el orden Sprint Goal, fechas incoherentes, fuera del proyecto", mensaje)
			break
		}
		posicion = i
	}
}

// US-008 / RN-008-10: el orden fijo con los siete errores posibles es Sprint Goal,
// falta inicio, falta fin y, solo si estan las dos fechas, incoherentes, duracion y
// fuera del proyecto. Una duracion excedida y fuera del proyecto salen juntas.
func TestNuevo_DuracionYFueraDelProyecto(t *testing.T) {
	_, err := nuevo("Objetivo", dia(2026, 12, 1), dia(2027, 2, 1))

	if !errors.Is(err, sprint.ErrDuracionExcedida) || !errors.Is(err, sprint.ErrFueraDelProyecto) {
		t.Errorf("error = %v, se esperaban ErrDuracionExcedida y ErrFueraDelProyecto", err)
	}
	if strings.Index(err.Error(), sprint.ErrDuracionExcedida.Error()) > strings.Index(err.Error(), sprint.ErrFueraDelProyecto.Error()) {
		t.Errorf("el mensaje %q no respeta el orden duracion, fuera del proyecto", err)
	}
}

// US-008 / RN-008-10: con datos invalidos no se devuelve sprint.
func TestNuevo_ConErroresNoDevuelveSprint(t *testing.T) {
	s, err := nuevo("", time.Time{}, time.Time{})
	if err == nil {
		t.Fatal("se esperaban errores")
	}
	if s != (sprint.Sprint{}) {
		t.Errorf("devolvio el sprint %+v, se esperaba el valor cero", s)
	}
}

// US-008 / RN-008-7, CL-008-18: la verificacion de las fechas contra el proyecto
// es una funcion aparte, que el repositorio repite dentro de la transaccion con las
// fechas bloqueadas. Es la misma que usa Nuevo.
func TestVerificarDentroDelProyecto(t *testing.T) {
	if err := sprint.VerificarDentroDelProyecto(dia(2026, 10, 6), dia(2026, 10, 19), inicioProyecto, finProyecto); err != nil {
		t.Errorf("dentro del proyecto: %v", err)
	}
	err := sprint.VerificarDentroDelProyecto(dia(2026, 10, 5), dia(2026, 10, 19), inicioProyecto, finProyecto)
	if !errors.Is(err, sprint.ErrFueraDelProyecto) {
		t.Errorf("error = %v, se esperaba ErrFueraDelProyecto", err)
	}
}

// ultimo arma el ultimo sprint de un proyecto para las verificaciones de la
// transaccion.
func ultimo(estado sprint.Estado, fin time.Time) *sprint.UltimoSprint {
	return &sprint.UltimoSprint{Numero: 2, Estado: estado, FechaFin: fin}
}

// US-008 / CA-008-4, CL-008-4, RN-008-13: con el ultimo sprint cerrado, el nuevo
// tiene que empezar despues de que termine: el dia siguiente se acepta, el mismo
// dia se rechaza.
func TestVerificarUltimoSprint_Cerrado(t *testing.T) {
	fin := dia(2026, 10, 19)
	if err := sprint.VerificarUltimoSprint(dia(2026, 10, 20), ultimo(sprint.EstadoCerrado, fin)); err != nil {
		t.Errorf("empezar el dia siguiente: %v", err)
	}

	err := sprint.VerificarUltimoSprint(dia(2026, 10, 19), ultimo(sprint.EstadoCerrado, fin))
	if !errors.Is(err, sprint.ErrInicioNoPosteriorAlUltimo) {
		t.Fatalf("empezar el mismo dia: error = %v, se esperaba ErrInicioNoPosteriorAlUltimo", err)
	}
	for _, esperado := range []string{"Sprint 2", "2026-10-19"} {
		if !strings.Contains(err.Error(), esperado) {
			t.Errorf("el error %q no indica %q (el nombre y el fin del ultimo)", err, esperado)
		}
	}
}

// US-008 / CL-008-5, RN-008-13: un sprint completamente anterior al ultimo se
// rechaza: tiene que empezar despues, para que el de numero mas alto sea tambien
// el que termina mas tarde.
func TestVerificarUltimoSprint_Anterior(t *testing.T) {
	err := sprint.VerificarUltimoSprint(dia(2026, 10, 6), ultimo(sprint.EstadoCerrado, dia(2026, 11, 2)))

	if !errors.Is(err, sprint.ErrInicioNoPosteriorAlUltimo) {
		t.Errorf("error = %v, se esperaba ErrInicioNoPosteriorAlUltimo", err)
	}
}

// US-008 / CA-008-11, CL-008-17, RN-008-17: con el ultimo sprint abierto se
// devuelve solo ErrUltimoSprintAbierto, indicando su nombre, sin comparar fechas:
// tanto si el nuevo empieza despues como si empieza antes de que termine.
func TestVerificarUltimoSprint_Abierto(t *testing.T) {
	for _, inicio := range []time.Time{dia(2026, 11, 1), dia(2026, 10, 10)} {
		err := sprint.VerificarUltimoSprint(inicio, ultimo(sprint.EstadoAbierto, dia(2026, 10, 19)))

		if !errors.Is(err, sprint.ErrUltimoSprintAbierto) {
			t.Fatalf("inicio %s: error = %v, se esperaba ErrUltimoSprintAbierto", inicio.Format(time.DateOnly), err)
		}
		if errors.Is(err, sprint.ErrInicioNoPosteriorAlUltimo) {
			t.Errorf("inicio %s: el error %q tambien reporta el inicio: nunca vienen juntos", inicio.Format(time.DateOnly), err)
		}
		if !strings.Contains(err.Error(), "Sprint 2") {
			t.Errorf("el error %q no indica el nombre del ultimo sprint", err)
		}
	}
}

// US-008 / CL-008-8, RN-008-13, RN-008-17: un proyecto sin sprints no tiene nada con
// que comparar.
func TestVerificarUltimoSprint_SinSprints(t *testing.T) {
	if err := sprint.VerificarUltimoSprint(dia(2026, 10, 6), nil); err != nil {
		t.Errorf("error = %v, se esperaba nil para un proyecto sin sprints", err)
	}
}
