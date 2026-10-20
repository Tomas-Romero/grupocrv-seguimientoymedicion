package httpadapter_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/a-h/templ"

	httpadapter "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/http"
	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

func contenidoDePrueba(texto string) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, texto)
		return err
	})
}

func responderPantalla(estado int, cabeceras map[string]string, contenido templ.Component) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodGet, "/proyectos", nil)
	for clave, valor := range cabeceras {
		r.Header.Set(clave, valor)
	}
	w := httptest.NewRecorder()
	httpadapter.Responder(w, r, estado, plantillas.Pagina{Titulo: "Proyectos", Activo: plantillas.SeccionProyectos}, contenido)
	return w
}

// CA-T004-1: sin cabeceras de HTMX responde la pagina completa, dentro del layout.
func TestResponder_PaginaCompleta(t *testing.T) {
	w := responderPantalla(http.StatusOK, nil, contenidoDePrueba("<p>hola</p>"))

	if w.Code != http.StatusOK {
		t.Errorf("estado = %d, quiero 200", w.Code)
	}
	cuerpo := w.Body.String()
	for _, quiero := range []string{"<!doctype html>", `<html lang="es">`, "<title>Proyectos | Software Metrics</title>", "<main", "<p>hola</p>"} {
		if !strings.Contains(strings.ToLower(cuerpo), strings.ToLower(quiero)) {
			t.Errorf("la pagina no tiene %q:\n%s", quiero, cuerpo)
		}
	}
	if !strings.Contains(cuerpo, `aria-current="page"`) {
		t.Errorf("la pagina no marca el item activo")
	}
}

// RN-T004-1, RN-T004-2, CA-T004-2: con HX-Request responde solo el fragmento.
func TestResponder_Fragmento(t *testing.T) {
	w := responderPantalla(http.StatusOK, map[string]string{"HX-Request": "true"}, contenidoDePrueba("<p>hola</p>"))

	if got := w.Body.String(); got != "<p>hola</p>" {
		t.Errorf("cuerpo = %q, quiero solo el fragmento", got)
	}
}

// RN-T004-2: toda respuesta de pantalla lleva Vary: HX-Request y su tipo de contenido.
func TestResponder_Cabeceras(t *testing.T) {
	for _, cabeceras := range []map[string]string{nil, {"HX-Request": "true"}} {
		w := responderPantalla(http.StatusOK, cabeceras, contenidoDePrueba("<p>hola</p>"))

		if got := w.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
			t.Errorf("%v: Content-Type = %q", cabeceras, got)
		}
		if got := w.Header().Values("Vary"); len(got) != 1 || got[0] != "HX-Request" {
			t.Errorf("%v: Vary = %v, quiero [HX-Request]", cabeceras, got)
		}
	}
}

// RN-T004-2: el Vary que ya habia no se pisa.
func TestResponder_ConservaElVaryQueYaHabia(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	w.Header().Set("Vary", "Accept-Encoding")

	httpadapter.Responder(w, r, http.StatusOK, plantillas.Pagina{}, contenidoDePrueba("x"))

	if got := strings.Join(w.Header().Values("Vary"), ","); got != "Accept-Encoding,HX-Request" {
		t.Errorf("Vary = %q, quiero Accept-Encoding,HX-Request", got)
	}
}

// RN-T004-1, CL-T004-1 y CL-T004-2: solo es fragmento con HX-Request: true y sin HX-Boosted: true.
func TestResponder_CuandoEsFragmento(t *testing.T) {
	casos := []struct {
		nombre    string
		cabeceras map[string]string
		fragmento bool
	}{
		{"sin cabeceras", nil, false},
		{"HX-Request true", map[string]string{"HX-Request": "true"}, true},
		{"HX-Request false", map[string]string{"HX-Request": "false"}, false},
		{"HX-Request con otro valor", map[string]string{"HX-Request": "si"}, false},
		{"HX-Request vacio", map[string]string{"HX-Request": ""}, false},
		{"HX-Boosted solo", map[string]string{"HX-Boosted": "true"}, false},
		{"HX-Request y HX-Boosted", map[string]string{"HX-Request": "true", "HX-Boosted": "true"}, false},
		{"HX-Request y HX-Boosted false", map[string]string{"HX-Request": "true", "HX-Boosted": "false"}, true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := responderPantalla(http.StatusOK, c.cabeceras, contenidoDePrueba("<p>hola</p>"))

			esFragmento := w.Body.String() == "<p>hola</p>"
			if esFragmento != c.fragmento {
				t.Errorf("fragmento = %v, quiero %v", esFragmento, c.fragmento)
			}
		})
	}
}

// RN-T004-3, CL-T004-5: el codigo lo decide quien llama, en pagina completa y en fragmento.
func TestResponder_ElCodigoLoDecideQuienLlama(t *testing.T) {
	for _, estado := range []int{http.StatusOK, http.StatusNotFound, http.StatusUnprocessableEntity} {
		for _, cabeceras := range []map[string]string{nil, {"HX-Request": "true"}} {
			if w := responderPantalla(estado, cabeceras, contenidoDePrueba("x")); w.Code != estado {
				t.Errorf("estado %d con %v: responde %d", estado, cabeceras, w.Code)
			}
		}
	}
}

// RN-T004-8, CA-T004-4: si el render falla responde 500 sin contenido parcial y lo registra.
func TestResponder_FallaElRender(t *testing.T) {
	var registro bytes.Buffer
	anterior := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&registro, nil)))
	t.Cleanup(func() { slog.SetDefault(anterior) })

	roto := templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, _ = io.WriteString(w, "<p>contenido parcial")
		return errors.New("la plantilla fallo")
	})
	for _, cabeceras := range []map[string]string{nil, {"HX-Request": "true"}} {
		registro.Reset()
		w := responderPantalla(http.StatusOK, cabeceras, roto)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("%v: estado = %d, quiero 500", cabeceras, w.Code)
		}
		if strings.Contains(w.Body.String(), "parcial") || strings.Contains(w.Body.String(), "<html") {
			t.Errorf("%v: el cuerpo tiene contenido parcial: %q", cabeceras, w.Body.String())
		}
		if got := strings.TrimSpace(w.Body.String()); got != "error interno" {
			t.Errorf("%v: cuerpo = %q, quiero el mensaje generico", cabeceras, got)
		}
		for _, quiero := range []string{"level=ERROR", "la plantilla fallo", "path=/proyectos"} {
			if !strings.Contains(registro.String(), quiero) {
				t.Errorf("%v: el registro no tiene %q: %s", cabeceras, quiero, registro.String())
			}
		}
	}
}
