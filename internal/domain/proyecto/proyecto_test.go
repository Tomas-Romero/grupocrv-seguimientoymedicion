package proyecto_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// fecha arma una fecha a medianoche UTC, que es como llegan al dominio.
func fecha(anio int, mes time.Month, dia int) time.Time {
	return time.Date(anio, mes, dia, 0, 0, 0, 0, time.UTC)
}

var (
	inicioValido = fecha(2026, time.September, 14)
	finValido    = fecha(2026, time.November, 18)
)

// US-001 / CA-001-1, CA-001-3 y RN-001-1 a RN-001-6: datos validos arman el
// proyecto, con el nombre recortado, la descripcion tal cual y sin ID.
func TestNuevo_Valido(t *testing.T) {
	casos := []struct {
		nombre             string
		entradaNombre      string
		entradaDescripcion string
		inicio             time.Time
		fin                time.Time
		quedaNombre        string
		quedaDescripcion   string
	}{
		{
			nombre:        "US-001 / CA-001-1: nombre valido y fechas coherentes",
			entradaNombre: "Demo", entradaDescripcion: "Proyecto de prueba",
			inicio: inicioValido, fin: finValido,
			quedaNombre: "Demo", quedaDescripcion: "Proyecto de prueba",
		},
		{
			nombre:        "US-001 / CA-001-3 / CL-001-1: fecha de fin igual a la de inicio",
			entradaNombre: "Relampago",
			inicio:        inicioValido, fin: inicioValido,
			quedaNombre: "Relampago",
		},
		{
			nombre:        "US-001 / CL-001-3: espacios en los extremos del nombre se recortan",
			entradaNombre: "  Demo  ",
			inicio:        inicioValido, fin: finValido,
			quedaNombre: "Demo",
		},
		{
			nombre:        "US-001 / CL-001-4: nombre de exactamente 100 caracteres",
			entradaNombre: strings.Repeat("a", 100),
			inicio:        inicioValido, fin: finValido,
			quedaNombre: strings.Repeat("a", 100),
		},
		{
			nombre:        "US-001 / RN-001-2: el limite se mide despues de recortar",
			entradaNombre: "  " + strings.Repeat("a", 100) + "\t",
			inicio:        inicioValido, fin: finValido,
			quedaNombre: strings.Repeat("a", 100),
		},
		{
			nombre:        "US-001 / CL-001-5: cien caracteres no ASCII (200 bytes)",
			entradaNombre: strings.Repeat("ñ", 100),
			inicio:        inicioValido, fin: finValido,
			quedaNombre: strings.Repeat("ñ", 100),
		},
		{
			nombre:        "US-001 / CL-001-6: descripcion vacia",
			entradaNombre: "Demo", entradaDescripcion: "",
			inicio: inicioValido, fin: finValido,
			quedaNombre: "Demo", quedaDescripcion: "",
		},
		{
			nombre:        "US-001 / RN-001-5: la descripcion se guarda sin recortar",
			entradaNombre: "Demo", entradaDescripcion: "  texto con espacios  ",
			inicio: inicioValido, fin: finValido,
			quedaNombre: "Demo", quedaDescripcion: "  texto con espacios  ",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			p, err := proyecto.Nuevo(c.entradaNombre, c.entradaDescripcion, c.inicio, c.fin)
			if err != nil {
				t.Fatalf("no se esperaba error, se obtuvo %v", err)
			}
			if p.ID != "" {
				t.Errorf("ID = %q, se esperaba vacio: lo asigna la persistencia (RN-001-6)", p.ID)
			}
			if p.Nombre != c.quedaNombre {
				t.Errorf("Nombre = %q, se esperaba %q", p.Nombre, c.quedaNombre)
			}
			if p.Descripcion != c.quedaDescripcion {
				t.Errorf("Descripcion = %q, se esperaba %q", p.Descripcion, c.quedaDescripcion)
			}
			if !p.FechaInicio.Equal(c.inicio) {
				t.Errorf("FechaInicio = %v, se esperaba %v", p.FechaInicio, c.inicio)
			}
			if !p.FechaFin.Equal(c.fin) {
				t.Errorf("FechaFin = %v, se esperaba %v", p.FechaFin, c.fin)
			}
		})
	}
}

// US-001 / CA-001-4, CA-001-5, RN-001-7 y CL-001-2, 4, 7, 8: cada dato invalido
// devuelve su error; con varios, se devuelven todos unidos en el orden fijo
// (nombre, falta inicio, falta fin, fechas incoherentes) y errors.Is reconoce
// cada uno. Nunca se devuelve un Proyecto.
func TestNuevo_Invalido(t *testing.T) {
	var sinFecha time.Time
	antes := fecha(2026, time.September, 13)

	casos := []struct {
		nombre        string
		entradaNombre string
		inicio        time.Time
		fin           time.Time
		errores       []error // en el orden en que tienen que venir unidos
	}{
		{
			nombre:        "US-001 / CA-001-5: nombre vacio",
			entradaNombre: "",
			inicio:        inicioValido, fin: finValido,
			errores: []error{proyecto.ErrNombreVacio},
		},
		{
			nombre:        "US-001 / CL-001-2: nombre con solo espacios",
			entradaNombre: "   \t ",
			inicio:        inicioValido, fin: finValido,
			errores: []error{proyecto.ErrNombreVacio},
		},
		{
			nombre:        "US-001 / CL-001-4: nombre de 101 caracteres",
			entradaNombre: strings.Repeat("a", 101),
			inicio:        inicioValido, fin: finValido,
			errores: []error{proyecto.ErrNombreLargo},
		},
		{
			nombre:        "US-001 / CL-001-4: 101 caracteres no ASCII",
			entradaNombre: strings.Repeat("ñ", 101),
			inicio:        inicioValido, fin: finValido,
			errores: []error{proyecto.ErrNombreLargo},
		},
		{
			nombre:        "US-001 / RN-001-3: falta la fecha de inicio",
			entradaNombre: "Demo",
			inicio:        sinFecha, fin: finValido,
			errores: []error{proyecto.ErrFechaInicioFaltante},
		},
		{
			nombre:        "US-001 / RN-001-3: falta la fecha de fin",
			entradaNombre: "Demo",
			inicio:        inicioValido, fin: sinFecha,
			errores: []error{proyecto.ErrFechaFinFaltante},
		},
		{
			nombre:        "US-001 / CA-001-4 / RN-001-4: fecha de fin anterior a la de inicio",
			entradaNombre: "Demo",
			inicio:        inicioValido, fin: antes,
			errores: []error{proyecto.ErrFechasIncoherentes},
		},
		{
			nombre:        "US-001 / CL-001-8: falta una fecha, no se evalua la coherencia",
			entradaNombre: "Demo",
			inicio:        sinFecha, fin: antes,
			errores: []error{proyecto.ErrFechaInicioFaltante},
		},
		{
			nombre:        "US-001 / CL-001-8: faltan las dos fechas, no se evalua la coherencia",
			entradaNombre: "Demo",
			inicio:        sinFecha, fin: sinFecha,
			errores: []error{proyecto.ErrFechaInicioFaltante, proyecto.ErrFechaFinFaltante},
		},
		{
			nombre:        "US-001 / CL-001-7: nombre vacio y fechas invertidas",
			entradaNombre: "",
			inicio:        inicioValido, fin: antes,
			errores: []error{proyecto.ErrNombreVacio, proyecto.ErrFechasIncoherentes},
		},
		{
			nombre:        "US-001 / RN-001-7: nombre largo y faltan las dos fechas, en orden fijo",
			entradaNombre: strings.Repeat("a", 101),
			inicio:        sinFecha, fin: sinFecha,
			errores: []error{proyecto.ErrNombreLargo, proyecto.ErrFechaInicioFaltante, proyecto.ErrFechaFinFaltante},
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			p, err := proyecto.Nuevo(c.entradaNombre, "", c.inicio, c.fin)
			if err == nil {
				t.Fatal("se esperaba un error y no hubo ninguno")
			}
			for _, esperado := range c.errores {
				if !errors.Is(err, esperado) {
					t.Errorf("errors.Is no reconoce %v dentro de %q", esperado, err)
				}
			}
			// Mismo mensaje que la union esperada: comprueba el orden fijo y
			// que no haya errores de mas.
			if esperado := errors.Join(c.errores...); err.Error() != esperado.Error() {
				t.Errorf("error = %q, se esperaba %q", err.Error(), esperado.Error())
			}
			if p != (proyecto.Proyecto{}) {
				t.Errorf("no se esperaba un Proyecto con datos invalidos, se obtuvo %+v", p)
			}
		})
	}
}

// US-001 / CA-001-2: crear un segundo proyecto no modifica el primero.
func TestNuevo_ProyectosIndependientes(t *testing.T) {
	alfa, err := proyecto.Nuevo("Alfa", "primero", fecha(2026, time.January, 5), fecha(2026, time.March, 6))
	if err != nil {
		t.Fatalf("crear Alfa: %v", err)
	}
	if _, err = proyecto.Nuevo("Beta", "segundo", fecha(2026, time.April, 1), fecha(2026, time.June, 30)); err != nil {
		t.Fatalf("crear Beta: %v", err)
	}
	if alfa.Nombre != "Alfa" || alfa.Descripcion != "primero" {
		t.Errorf("Alfa cambio despues de crear Beta: %+v", alfa)
	}
}

// US-001 / seccion 7: el mensaje de cada error es el que ve la persona y el que
// comprueban los escenarios BDD con `el mensaje de error indica "..."`.
func TestErrores_Mensajes(t *testing.T) {
	casos := []struct {
		err      error
		esperado string
	}{
		{proyecto.ErrNombreVacio, "el nombre es obligatorio"},
		{proyecto.ErrNombreLargo, "el nombre supera el largo maximo (100 caracteres)"},
		{proyecto.ErrFechaInicioFaltante, "falta la fecha de inicio"},
		{proyecto.ErrFechaFinFaltante, "falta la fecha de fin"},
		{proyecto.ErrFechasIncoherentes, "fecha de fin anterior a la de inicio"},
	}

	for _, c := range casos {
		t.Run(c.esperado, func(t *testing.T) {
			if c.err.Error() != c.esperado {
				t.Errorf("mensaje = %q, se esperaba %q", c.err.Error(), c.esperado)
			}
		})
	}
}