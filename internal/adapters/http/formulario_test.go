package httpadapter_test

import (
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	httpadapter "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/http"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/errores"
)

func formulario(campo, valor string) *httpadapter.Formulario {
	return httpadapter.NuevoFormulario(url.Values{campo: {valor}})
}

// erroresDe devuelve los errores de campo del formulario, o nil si no hubo ninguno.
func erroresDe(t *testing.T, f *httpadapter.Formulario) []errores.ErrorDeCampo {
	t.Helper()
	err := f.Err()
	if err == nil {
		return nil
	}
	var deValidacion *errores.ErrorDeValidacion
	if !errors.As(err, &deValidacion) {
		t.Fatalf("Err() = %T, quiero *errores.ErrorDeValidacion", err)
	}
	return deValidacion.Errores
}

// RN-T005-11 y CL-T005-8: Entero acepta digitos con signo, ignora espacios y rechaza el resto.
func TestFormulario_Entero(t *testing.T) {
	casos := []struct {
		valor  string
		quiero int
		valido bool
	}{
		{"5", 5, true},
		{" 5 ", 5, true},
		{"+5", 5, true},
		{"-3", -3, true},
		{"007", 7, true},
		{"", 0, true},
		{"   ", 0, true},
		{"12.5", 0, false},
		{"12,5", 0, false},
		{"1e3", 0, false},
		{"abc", 0, false},
		{"5 5", 0, false},
		{"0x10", 0, false},
		{"99999999999999999999", 0, false},
	}
	for _, c := range casos {
		t.Run(c.valor, func(t *testing.T) {
			f := formulario("story_points", c.valor)
			got := f.Entero("story_points")

			if got != c.quiero {
				t.Errorf("Entero(%q) = %d, quiero %d", c.valor, got, c.quiero)
			}
			errs := erroresDe(t, f)
			if c.valido && errs != nil {
				t.Errorf("Entero(%q) informo %+v, quiero ningun error", c.valor, errs)
			}
			if !c.valido {
				quiero := []errores.ErrorDeCampo{{Campo: "story_points", Mensaje: "el campo story points tiene que ser un numero entero"}}
				if !reflect.DeepEqual(errs, quiero) {
					t.Errorf("Entero(%q) informo %+v, quiero %+v", c.valor, errs, quiero)
				}
			}
		})
	}
}

// RN-T005-11 y CL-T005-9: Decimal acepta coma o punto y rechaza notacion cientifica, NaN e Inf.
func TestFormulario_Decimal(t *testing.T) {
	casos := []struct {
		valor  string
		quiero float64
		valido bool
	}{
		{"1,5", 1.5, true},
		{"1.5", 1.5, true},
		{"2", 2, true},
		{" 0,25 ", 0.25, true},
		{"-2,25", -2.25, true},
		{"+3", 3, true},
		{"", 0, true},
		{"1,5,2", 0, false},
		{"1.2.3", 0, false},
		{"1,2.3", 0, false},
		{"NaN", 0, false},
		{"Inf", 0, false},
		{"-Inf", 0, false},
		{"1e3", 0, false},
		{".5", 0, false},
		{"5.", 0, false},
		{"abc", 0, false},
		{strings.Repeat("9", 400), 0, false}, // no entra en un float64
	}
	for _, c := range casos {
		t.Run(c.valor[:min(len(c.valor), 12)], func(t *testing.T) {
			f := formulario("horas", c.valor)
			got := f.Decimal("horas")

			if got != c.quiero {
				t.Errorf("Decimal(%q) = %v, quiero %v", c.valor, got, c.quiero)
			}
			errs := erroresDe(t, f)
			if c.valido && errs != nil {
				t.Errorf("Decimal(%q) informo %+v, quiero ningun error", c.valor, errs)
			}
			if !c.valido {
				quiero := []errores.ErrorDeCampo{{Campo: "horas", Mensaje: "el campo horas tiene que ser un numero"}}
				if !reflect.DeepEqual(errs, quiero) {
					t.Errorf("Decimal(%q) informo %+v, quiero %+v", c.valor, errs, quiero)
				}
			}
		})
	}
}

// RN-T005-11 y CL-T005-10: Fecha acepta AAAA-MM-DD, a medianoche UTC.
func TestFormulario_Fecha(t *testing.T) {
	casos := []struct {
		valor  string
		quiero time.Time
		valido bool
	}{
		{"2026-10-06", time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC), true},
		{" 2026-10-06 ", time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC), true},
		{"2028-02-29", time.Date(2028, 2, 29, 0, 0, 0, 0, time.UTC), true},
		{"", time.Time{}, true},
		{"2026-02-30", time.Time{}, false},
		{"2026-02-29", time.Time{}, false},
		{"06/10/2026", time.Time{}, false},
		{"2026-10-6", time.Time{}, false},
		{"2026-10-06T10:00:00Z", time.Time{}, false},
		{"mañana", time.Time{}, false},
	}
	for _, c := range casos {
		t.Run(c.valor, func(t *testing.T) {
			f := formulario("fecha", c.valor)
			got := f.Fecha("fecha")

			if !got.Equal(c.quiero) || got.Location() != c.quiero.Location() {
				t.Errorf("Fecha(%q) = %v, quiero %v", c.valor, got, c.quiero)
			}
			errs := erroresDe(t, f)
			if c.valido && errs != nil {
				t.Errorf("Fecha(%q) informo %+v, quiero ningun error", c.valor, errs)
			}
			if !c.valido {
				quiero := []errores.ErrorDeCampo{{Campo: "fecha", Mensaje: "el campo fecha tiene que ser una fecha con formato AAAA-MM-DD"}}
				if !reflect.DeepEqual(errs, quiero) {
					t.Errorf("Fecha(%q) informo %+v, quiero %+v", c.valor, errs, quiero)
				}
			}
		})
	}
}

// RN-T005-10: un campo que no vino o vino vacio no es un error de formato.
func TestFormulario_CampoAusenteNoEsError(t *testing.T) {
	f := httpadapter.NuevoFormulario(url.Values{})

	if f.Entero("a") != 0 || f.Decimal("b") != 0 || !f.Fecha("c").IsZero() {
		t.Error("un campo ausente tiene que dar el cero de su tipo")
	}
	if err := f.Err(); err != nil {
		t.Errorf("Err() = %v, quiero nil", err)
	}
}

// Err() devuelve un nil de verdad (no un puntero nulo dentro de una interfaz).
func TestFormulario_ErrSinProblemasEsNil(t *testing.T) {
	f := formulario("horas", "2")
	f.Decimal("horas")
	if f.Err() != nil {
		t.Errorf("Err() = %#v, quiero nil", f.Err())
	}
}

// RN-T005-12 y CL-T005-11: todos los campos mal formados salen juntos, en el orden en que se
// consultaron, y un campo se informa una sola vez.
func TestFormulario_VariosCamposMalFormados(t *testing.T) {
	f := httpadapter.NuevoFormulario(url.Values{
		"horas":         {"mucho"},
		"fecha":         {"ayer"},
		"story_points":  {"cinco"},
		"actividad":     {"revision"},
		"fecha_fin":     {"2026-10-20"},
		"otro_decimal":  {"1,2,3"},
		"entero_bueno":  {"4"},
		"decimal_bueno": {"4,5"},
	})
	f.Decimal("horas")
	f.Fecha("fecha")
	f.Fecha("fecha_fin")
	f.Entero("story_points")
	f.Decimal("horas") // el mismo campo otra vez
	f.Decimal("otro_decimal")
	f.Entero("entero_bueno")

	var campos []string
	for _, e := range erroresDe(t, f) {
		campos = append(campos, e.Campo)
	}
	quiero := []string{"horas", "fecha", "story_points", "otro_decimal"}
	if !reflect.DeepEqual(campos, quiero) {
		t.Errorf("campos = %v, quiero %v", campos, quiero)
	}
}

// El valor valido de un campo se lee igual aunque otro campo este mal.
func TestFormulario_UnCampoMalNoAfectaAOtro(t *testing.T) {
	f := httpadapter.NuevoFormulario(url.Values{"horas": {"x"}, "story_points": {"8"}})

	f.Decimal("horas")
	if got := f.Entero("story_points"); got != 8 {
		t.Errorf("Entero(story_points) = %d, quiero 8", got)
	}
}

// Si un campo viene repetido se usa el primer valor, como url.Values.Get.
func TestFormulario_CampoRepetidoUsaElPrimero(t *testing.T) {
	f := httpadapter.NuevoFormulario(url.Values{"horas": {"2", "9"}})
	if got := f.Decimal("horas"); got != 2 {
		t.Errorf("Decimal(horas) = %v, quiero 2", got)
	}
}
