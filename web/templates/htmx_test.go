package plantillas_test

import (
	"encoding/json"
	"html"
	"regexp"
	"testing"

	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// manejoDeRespuesta es una entrada de htmx.config.responseHandling.
type manejoDeRespuesta struct {
	Code   string `json:"code"`
	Swap   bool   `json:"swap"`
	Error  bool   `json:"error"`
	Target string `json:"target"`
}

// configuracionDeHtmx lee el JSON de <meta name="htmx-config"> de la pagina.
func configuracionDeHtmx(t *testing.T) []manejoDeRespuesta {
	t.Helper()
	pagina := renderizar(t, plantillas.Pagina{Titulo: "Proyectos"}, "<p>hola</p>")
	meta := etiqueta(t, pagina, `<meta[^>]*name="htmx-config"[^>]*>`)
	contenido := regexp.MustCompile(`content="([^"]*)"`).FindStringSubmatch(meta)
	if contenido == nil {
		t.Fatalf("la meta htmx-config no tiene content: %s", meta)
	}
	var configuracion struct {
		ResponseHandling []manejoDeRespuesta `json:"responseHandling"`
	}
	if err := json.Unmarshal([]byte(html.UnescapeString(contenido[1])), &configuracion); err != nil {
		t.Fatalf("el content de htmx-config no es JSON valido: %v", err)
	}
	return configuracion.ResponseHandling
}

// manejoPara hace lo que htmx 2.x: devuelve la primera entrada cuyo code (una expresion
// regular) coincide con el estado. Sin ninguna, htmx no intercambia nada.
func manejoPara(t *testing.T, entradas []manejoDeRespuesta, estado string) manejoDeRespuesta {
	t.Helper()
	for _, e := range entradas {
		if regexp.MustCompile(e.Code).MatchString(estado) {
			return e
		}
	}
	return manejoDeRespuesta{}
}

// RN-T004-10: los errores uniformes de T-005 (404, 409, 422 y 500) se intercambian dentro de
// #mensajes; htmx no intercambia 4xx ni 5xx por defecto.
func TestLayout_HtmxMuestraLosErroresUniformes(t *testing.T) {
	entradas := configuracionDeHtmx(t)

	for _, estado := range []string{"404", "409", "422", "500"} {
		m := manejoPara(t, entradas, estado)
		if !m.Swap || m.Target != "#mensajes" {
			t.Errorf("estado %s: %+v, quiero swap en #mensajes", estado, m)
		}
	}
}

// RN-T004-10: las demas respuestas conservan el comportamiento por defecto de htmx.
func TestLayout_HtmxConservaLosValoresPorDefecto(t *testing.T) {
	entradas := configuracionDeHtmx(t)

	casos := []struct {
		estado string
		swap   bool
		error  bool
	}{
		{"200", true, false},
		{"201", true, false},
		{"302", true, false},
		{"204", false, false},
		{"400", false, true},
		{"401", false, true},
		{"403", false, true},
		{"503", false, true},
	}
	for _, c := range casos {
		m := manejoPara(t, entradas, c.estado)
		if m.Swap != c.swap || m.Error != c.error || m.Target != "" {
			t.Errorf("estado %s: %+v, quiero swap=%v error=%v sin target", c.estado, m, c.swap, c.error)
		}
	}
}

// El code de la regla de errores no puede coincidir con un estado ajeno por ser una subcadena.
func TestLayout_HtmxLaReglaDeErroresEsExacta(t *testing.T) {
	entradas := configuracionDeHtmx(t)

	for _, estado := range []string{"1404", "4045", "422x", "0422"} {
		if m := manejoPara(t, entradas, estado); m.Target == "#mensajes" {
			t.Errorf("estado %s: la regla de errores coincidio con algo que no es 404, 409, 422 ni 500", estado)
		}
	}
}
