package esfuerzo_test

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
)

// hoy es la fecha de hoy de todos los tests: el dominio la recibe por parametro y
// no lee el reloj. Tiene hora a proposito (CL-019-7).
var hoy = time.Date(2026, 10, 17, 23, 59, 0, 0, time.UTC)

func dia(anio int, mes time.Month, d int) time.Time {
	return time.Date(anio, mes, d, 0, 0, 0, 0, time.UTC)
}

// nuevo arma un registro con datos validos, salvo lo que cada test cambia.
func nuevo(fecha time.Time, actividad string, horas float64) (esfuerzo.RegistroEsfuerzo, error) {
	return esfuerzo.Nuevo("proyecto-1", "item-1", "integrante-1", fecha, actividad, horas, hoy)
}

// US-019 / CA-019-1, CA-019-2, RN-019-5: un registro valido guarda los IDs
// recibidos, el dia (sin hora), la actividad recortada y las horas. No tiene ID:
// lo genera la base.
func TestNuevo_DatosValidos(t *testing.T) {
	r, err := esfuerzo.Nuevo("proyecto-1", "item-1", "integrante-1", dia(2026, 10, 16).Add(15*time.Hour), "  Desarrollo  ", 2.5, hoy)
	if err != nil {
		t.Fatalf("Nuevo: %v", err)
	}

	if r.ID != "" {
		t.Errorf("ID = %q, se esperaba vacio: lo genera la base", r.ID)
	}
	if r.ProyectoID != "proyecto-1" || r.ItemID != "item-1" || r.IntegranteID != "integrante-1" {
		t.Errorf("IDs = %q, %q, %q; se esperaban los recibidos", r.ProyectoID, r.ItemID, r.IntegranteID)
	}
	if !r.Fecha.Equal(dia(2026, 10, 16)) {
		t.Errorf("fecha = %v, se esperaba el dia 2026-10-16 sin hora", r.Fecha)
	}
	if r.Actividad != "Desarrollo" {
		t.Errorf("actividad = %q, se esperaba %q (recortada)", r.Actividad, "Desarrollo")
	}
	if r.Horas != 2.5 {
		t.Errorf("horas = %v, se esperaba 2.5", r.Horas)
	}
}

// US-019 / CL-019-1, CL-019-2, CL-019-4: los extremos validos de las horas y los
// multiplos de 0,25.
func TestNuevo_HorasValidas(t *testing.T) {
	for _, horas := range []float64{0.25, 0.5, 1, 1.5, 1.75, 8, 23.75, 24} {
		if _, err := nuevo(dia(2026, 10, 16), "Desarrollo", horas); err != nil {
			t.Errorf("horas %v: %v, se esperaba que fueran validas", horas, err)
		}
	}
}

// US-019 / RN-019-1, CL-019-1 a CL-019-3: cero, negativos, mas de 24, NaN e
// infinito se rechazan con ErrHorasFueraDeRango.
func TestNuevo_HorasFueraDeRango(t *testing.T) {
	for _, horas := range []float64{0, -1, -0.25, 24.25, 25, 100, math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, err := nuevo(dia(2026, 10, 16), "Desarrollo", horas)
		if !errors.Is(err, esfuerzo.ErrHorasFueraDeRango) {
			t.Errorf("horas %v: error = %v, se esperaba ErrHorasFueraDeRango", horas, err)
		}
		if errors.Is(err, esfuerzo.ErrHorasFraccion) {
			t.Errorf("horas %v: el error tambien reporta fraccion: de las horas se devuelve uno solo", horas)
		}
	}
}

// US-019 / RN-019-2, CL-019-4: lo que no es multiplo de 0,25 se rechaza con
// ErrHorasFraccion.
func TestNuevo_HorasConFraccion(t *testing.T) {
	for _, horas := range []float64{0.1, 1.3, 1.26, 2.333, 23.9} {
		_, err := nuevo(dia(2026, 10, 16), "Desarrollo", horas)
		if !errors.Is(err, esfuerzo.ErrHorasFraccion) {
			t.Errorf("horas %v: error = %v, se esperaba ErrHorasFraccion", horas, err)
		}
	}
}

// US-019 / RN-019-8, CL-019-5: si las horas estan fuera de rango y ademas tienen
// fraccion (30,3), se devuelve solo el error de rango.
func TestNuevo_RangoTienePrioridadSobreFraccion(t *testing.T) {
	_, err := nuevo(dia(2026, 10, 16), "Desarrollo", 30.3)

	if !errors.Is(err, esfuerzo.ErrHorasFueraDeRango) {
		t.Fatalf("error = %v, se esperaba ErrHorasFueraDeRango", err)
	}
	if errors.Is(err, esfuerzo.ErrHorasFraccion) {
		t.Errorf("el error %q tambien reporta fraccion", err)
	}
}

// US-019 / RN-019-3, CA-019-5, CL-019-6, CL-019-7: hoy y los dias anteriores se
// aceptan, aunque la hora de la fecha sea posterior a la de hoy; manana se
// rechaza con ErrFechaFutura.
func TestNuevo_Fecha(t *testing.T) {
	casos := []struct {
		nombre string
		fecha  time.Time
		futura bool
	}{
		{"hoy a medianoche", dia(2026, 10, 17), false},
		{"hoy con una hora posterior a la de hoy", dia(2026, 10, 17).Add(23*time.Hour + 59*time.Minute + 59*time.Second), false},
		{"ayer", dia(2026, 10, 16), false},
		{"un dia de hace meses", dia(2026, 1, 5), false},
		{"manana", dia(2026, 10, 18), true},
		{"manana a primera hora", dia(2026, 10, 18).Add(time.Second), true},
		{"un ano despues", dia(2027, 10, 17), true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := nuevo(c.fecha, "Desarrollo", 1)
			if c.futura != errors.Is(err, esfuerzo.ErrFechaFutura) {
				t.Errorf("error = %v, ErrFechaFutura esperado: %t", err, c.futura)
			}
			if !c.futura && err != nil {
				t.Errorf("error = %v, se esperaba una fecha valida", err)
			}
		})
	}
}

// US-019 / RN-019-4, CL-019-8: la fecha sin completar (valor cero) se rechaza con
// ErrFechaVacia y no con ErrFechaFutura ni ninguna otra.
func TestNuevo_FechaVacia(t *testing.T) {
	_, err := nuevo(time.Time{}, "Desarrollo", 1)

	if !errors.Is(err, esfuerzo.ErrFechaVacia) {
		t.Fatalf("error = %v, se esperaba ErrFechaVacia", err)
	}
	if errors.Is(err, esfuerzo.ErrFechaFutura) {
		t.Errorf("el error %q tambien reporta fecha futura", err)
	}
}

// US-019 / RN-019-5, CL-019-9, CL-019-10: la actividad se mide en caracteres
// despues de recortarla: 120 se aceptan, 121 se rechazan, y solo espacios es
// vacia y no larga.
func TestNuevo_Actividad(t *testing.T) {
	casos := []struct {
		nombre    string
		actividad string
		error     error
	}{
		{"120 caracteres", strings.Repeat("a", 120), nil},
		{"120 caracteres con espacios en los extremos", "  " + strings.Repeat("a", 120) + "  ", nil},
		{"120 caracteres no ASCII (240 bytes)", strings.Repeat("ñ", 120), nil},
		{"121 caracteres", strings.Repeat("a", 121), esfuerzo.ErrActividadMuyLarga},
		{"121 caracteres no ASCII", strings.Repeat("ñ", 121), esfuerzo.ErrActividadMuyLarga},
		{"vacia", "", esfuerzo.ErrActividadVacia},
		{"solo espacios", "   ", esfuerzo.ErrActividadVacia},
		{"solo espacios y mas de 120", strings.Repeat(" ", 150), esfuerzo.ErrActividadVacia},
		{"solo tabulaciones y saltos de linea", "\t\n ", esfuerzo.ErrActividadVacia},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := nuevo(dia(2026, 10, 16), c.actividad, 1)

			if c.error == nil && err != nil {
				t.Fatalf("error = %v, se esperaba una actividad valida", err)
			}
			if c.error != nil && !errors.Is(err, c.error) {
				t.Fatalf("error = %v, se esperaba %v", err, c.error)
			}
		})
	}
}

// US-019 / RN-019-5: lo que se guarda es la actividad recortada.
func TestNuevo_ActividadSeGuardaRecortada(t *testing.T) {
	r, err := nuevo(dia(2026, 10, 16), "\t Revision de PR \n", 1)
	if err != nil {
		t.Fatalf("Nuevo: %v", err)
	}
	if r.Actividad != "Revision de PR" {
		t.Errorf("actividad = %q, se esperaba %q", r.Actividad, "Revision de PR")
	}
}

// US-019 / seccion 7: los errores de largo indican cuantos caracteres tiene la
// actividad, y los de fecha y horas, el valor recibido.
func TestNuevo_ErroresIndicanElValorRecibido(t *testing.T) {
	_, err := nuevo(dia(2026, 10, 18), strings.Repeat("a", 121), 1.3)
	if err == nil {
		t.Fatal("se esperaban errores")
	}
	for _, esperado := range []string{"121", "2026-10-18", "1.3"} {
		if !strings.Contains(err.Error(), esperado) {
			t.Errorf("el mensaje %q no indica %q", err.Error(), esperado)
		}
	}
}

// US-019 / RN-019-8, CL-019-13: con varios datos invalidos se devuelven todos
// unidos, en este orden: fecha, actividad, horas.
func TestNuevo_VariosErroresEnOrdenFijo(t *testing.T) {
	_, err := nuevo(dia(2026, 10, 18), "", 0)

	esperados := []error{esfuerzo.ErrFechaFutura, esfuerzo.ErrActividadVacia, esfuerzo.ErrHorasFueraDeRango}
	for _, e := range esperados {
		if !errors.Is(err, e) {
			t.Errorf("error = %v, se esperaba que incluyera %v", err, e)
		}
	}
	mensaje := err.Error()
	posicion := -1
	for _, e := range esperados {
		i := strings.Index(mensaje, e.Error())
		if i <= posicion {
			t.Errorf("el mensaje %q no respeta el orden fecha, actividad, horas", mensaje)
			break
		}
		posicion = i
	}
	if lineas := strings.Count(mensaje, "\n") + 1; lineas != 3 {
		t.Errorf("el mensaje tiene %d lineas, se esperaban 3 (una por error)", lineas)
	}
}

// US-019 / RN-019-8: con datos invalidos no se devuelve registro.
func TestNuevo_ConErroresNoDevuelveRegistro(t *testing.T) {
	r, err := nuevo(time.Time{}, "", -1)
	if err == nil {
		t.Fatal("se esperaban errores")
	}
	if r != (esfuerzo.RegistroEsfuerzo{}) {
		t.Errorf("devolvio el registro %+v, se esperaba el valor cero", r)
	}
}

// US-019 / CL-019-11: dos registros iguales son validos: el dominio no guarda
// nada, asi que no hay nada que unificar ni rechazar por duplicado.
func TestNuevo_RegistrosIgualesSonValidos(t *testing.T) {
	primero, err1 := nuevo(dia(2026, 10, 16), "Desarrollo", 2)
	segundo, err2 := nuevo(dia(2026, 10, 16), "Desarrollo", 2)

	if err1 != nil || err2 != nil {
		t.Fatalf("errores = %v, %v; se esperaba que los dos fueran validos", err1, err2)
	}
	if primero != segundo {
		t.Errorf("los dos registros deberian ser iguales: %+v y %+v", primero, segundo)
	}
}
