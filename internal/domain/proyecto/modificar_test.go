package proyecto_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// proyectoExistente es el proyecto que la capa de aplicacion cargaria antes de
// pasarselo al dominio para modificarlo.
func proyectoExistente() proyecto.Proyecto {
	return proyecto.Proyecto{
		ID:          "0b6f3c52-4f0e-4a43-9d3e-6a1f2c7d9e10",
		Nombre:      "Demo",
		Descripcion: "Descripcion original",
		FechaInicio: inicioValido,
		FechaFin:    finValido,
	}
}

func mensaje(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// US-002 / CA-002-1 a CA-002-3, RN-002-3, RN-002-4, RN-002-6 y CL-002-1 a 3, 5, 6:
// datos validos reemplazan los cuatro datos del proyecto y conservan el ID.
func TestModificar_Valido(t *testing.T) {
	nuevoInicio := fecha(2026, time.October, 1)
	nuevoFin := fecha(2026, time.December, 15)

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
			nombre:        "US-002 / CA-002-1: cambia el nombre y conserva las fechas",
			entradaNombre: "Demo v2", entradaDescripcion: "Descripcion original",
			inicio: inicioValido, fin: finValido,
			quedaNombre: "Demo v2", quedaDescripcion: "Descripcion original",
		},
		{
			nombre:        "US-002 / CA-002-2: cambia las fechas y conserva el nombre",
			entradaNombre: "Demo", entradaDescripcion: "Descripcion original",
			inicio: nuevoInicio, fin: nuevoFin,
			quedaNombre: "Demo", quedaDescripcion: "Descripcion original",
		},
		{
			nombre:        "US-002 / RN-002-6 / CL-002-1: los mismos datos que ya tenia",
			entradaNombre: "Demo", entradaDescripcion: "Descripcion original",
			inicio: inicioValido, fin: finValido,
			quedaNombre: "Demo", quedaDescripcion: "Descripcion original",
		},
		{
			nombre:        "US-002 / CA-002-3 / CL-002-2: fecha de fin igual a la de inicio",
			entradaNombre: "Demo", entradaDescripcion: "Descripcion original",
			inicio: nuevoInicio, fin: nuevoInicio,
			quedaNombre: "Demo", quedaDescripcion: "Descripcion original",
		},
		{
			nombre:        "US-002 / CL-002-3: espacios en los extremos del nombre se recortan",
			entradaNombre: "  Demo v2  ", entradaDescripcion: "Descripcion original",
			inicio: inicioValido, fin: finValido,
			quedaNombre: "Demo v2", quedaDescripcion: "Descripcion original",
		},
		{
			nombre:        "US-002 / CL-002-5: nombre de exactamente 100 caracteres",
			entradaNombre: strings.Repeat("a", 100), entradaDescripcion: "Descripcion original",
			inicio: inicioValido, fin: finValido,
			quedaNombre: strings.Repeat("a", 100), quedaDescripcion: "Descripcion original",
		},
		{
			nombre:        "US-002 / CL-002-5: cien caracteres no ASCII (200 bytes)",
			entradaNombre: strings.Repeat("ñ", 100), entradaDescripcion: "Descripcion original",
			inicio: inicioValido, fin: finValido,
			quedaNombre: strings.Repeat("ñ", 100), quedaDescripcion: "Descripcion original",
		},
		{
			nombre:        "US-002 / CL-002-6: descripcion vacia borra la que tenia",
			entradaNombre: "Demo", entradaDescripcion: "",
			inicio: inicioValido, fin: finValido,
			quedaNombre: "Demo", quedaDescripcion: "",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			actual := proyectoExistente()

			p, err := proyecto.Modificar(actual, c.entradaNombre, c.entradaDescripcion, c.inicio, c.fin)
			if err != nil {
				t.Fatalf("no se esperaba error, se obtuvo %v", err)
			}
			if p.ID != actual.ID {
				t.Errorf("ID = %q, se esperaba %q: el ID nunca cambia (RN-002-4)", p.ID, actual.ID)
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

// US-002 / CA-002-4, RN-002-2, RN-002-5 y CL-002-5, 7: un dato invalido devuelve
// su error; con varios, se devuelven todos unidos en el orden fijo de RN-001-7
// y errors.Is reconoce cada uno. La operacion es de todo o nada: nunca se
// devuelve un Proyecto, ni siquiera con los datos validos aplicados.
func TestModificar_Invalido(t *testing.T) {
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
			nombre:        "US-002 / CA-002-4: nombre nuevo vacio",
			entradaNombre: "",
			inicio:        inicioValido, fin: finValido,
			errores: []error{proyecto.ErrNombreVacio},
		},
		{
			nombre:        "US-002 / RN-002-2: nombre nuevo con solo espacios",
			entradaNombre: "   \t ",
			inicio:        inicioValido, fin: finValido,
			errores: []error{proyecto.ErrNombreVacio},
		},
		{
			nombre:        "US-002 / CL-002-5: nombre nuevo de 101 caracteres",
			entradaNombre: strings.Repeat("a", 101),
			inicio:        inicioValido, fin: finValido,
			errores: []error{proyecto.ErrNombreLargo},
		},
		{
			nombre:        "US-002 / RN-002-2: falta la fecha de inicio",
			entradaNombre: "Demo",
			inicio:        sinFecha, fin: finValido,
			errores: []error{proyecto.ErrFechaInicioFaltante},
		},
		{
			nombre:        "US-002 / RN-002-2: falta la fecha de fin",
			entradaNombre: "Demo",
			inicio:        inicioValido, fin: sinFecha,
			errores: []error{proyecto.ErrFechaFinFaltante},
		},
		{
			nombre:        "US-002 / RN-002-2: fecha de fin anterior a la de inicio",
			entradaNombre: "Demo",
			inicio:        inicioValido, fin: antes,
			errores: []error{proyecto.ErrFechasIncoherentes},
		},
		{
			nombre:        "US-002 / CL-002-7: nombre nuevo correcto con fechas invertidas, no se aplica el nombre",
			entradaNombre: "Demo v2",
			inicio:        inicioValido, fin: antes,
			errores: []error{proyecto.ErrFechasIncoherentes},
		},
		{
			nombre:        "US-002 / RN-001-7: nombre vacio y fechas invertidas, en orden fijo",
			entradaNombre: "",
			inicio:        inicioValido, fin: antes,
			errores: []error{proyecto.ErrNombreVacio, proyecto.ErrFechasIncoherentes},
		},
		{
			nombre:        "US-002 / RN-001-7: nombre largo y faltan las dos fechas, en orden fijo",
			entradaNombre: strings.Repeat("a", 101),
			inicio:        sinFecha, fin: sinFecha,
			errores: []error{proyecto.ErrNombreLargo, proyecto.ErrFechaInicioFaltante, proyecto.ErrFechaFinFaltante},
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			p, err := proyecto.Modificar(proyectoExistente(), c.entradaNombre, "Descripcion nueva", c.inicio, c.fin)
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
				t.Errorf("no se esperaba un Proyecto con datos invalidos (RN-002-5), se obtuvo %+v", p)
			}
		})
	}
}

// US-002 / RN-002-2: modificar usa las mismas reglas que crear. Para las mismas
// entradas, Modificar y Nuevo dan el mismo resultado (el mismo error, o el mismo
// proyecto con el ID del existente), asi una regla nueva en Nuevo no puede
// quedar sin aplicarse al modificar.
func TestModificar_MismasReglasQueNuevo(t *testing.T) {
	var sinFecha time.Time
	antes := fecha(2026, time.September, 13)

	entradas := []struct {
		caso   string
		nombre string
		inicio time.Time
		fin    time.Time
	}{
		{"datos validos", "Demo", inicioValido, finValido},
		{"nombre con espacios y un solo dia", "  Demo  ", inicioValido, inicioValido},
		{"nombre vacio y fechas invertidas", "", inicioValido, antes},
		{"nombre largo no ASCII y sin fechas", strings.Repeat("ñ", 101), sinFecha, sinFecha},
		{"falta la fecha de inicio", "Demo", sinFecha, finValido},
		{"fechas invertidas", "Demo", inicioValido, antes},
	}

	for _, e := range entradas {
		t.Run(e.caso, func(t *testing.T) {
			actual := proyectoExistente()

			esperado, errNuevo := proyecto.Nuevo(e.nombre, "texto", e.inicio, e.fin)
			obtenido, errModificar := proyecto.Modificar(actual, e.nombre, "texto", e.inicio, e.fin)

			if mensaje(errModificar) != mensaje(errNuevo) {
				t.Fatalf("error de Modificar = %q, el de Nuevo = %q", mensaje(errModificar), mensaje(errNuevo))
			}
			esperado.ID = actual.ID
			if errNuevo == nil && obtenido != esperado {
				t.Errorf("Modificar = %+v, se esperaba %+v", obtenido, esperado)
			}
		})
	}
}
