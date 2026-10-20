package httpadapter_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	httpadapter "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/http"
)

func routerDePrueba() http.Handler {
	estaticos := fstest.MapFS{
		"css/app.css":    {Data: []byte("body{color:red}")},
		"js/htmx.min.js": {Data: []byte("var htmx={};")},
	}
	return httpadapter.NuevoRouter(verificadorFalso{}, time.Second, http.FS(estaticos))
}

func pedir(h http.Handler, metodo, ruta string, cabeceras map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, nil)
	for clave, valor := range cabeceras {
		r.Header.Set(clave, valor)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// CA-T004-1, RN-T004-5: GET / responde el inicio dentro del layout, con Inicio como item activo.
func TestRouter_Inicio(t *testing.T) {
	w := pedir(routerDePrueba(), http.MethodGet, "/", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("estado = %d, quiero 200", w.Code)
	}
	cuerpo := w.Body.String()
	for _, quiero := range []string{"<!doctype html>", "Software Metrics &amp; Estimation", "<main"} {
		if !strings.Contains(strings.ToLower(cuerpo), strings.ToLower(quiero)) {
			t.Errorf("el inicio no tiene %q:\n%s", quiero, cuerpo)
		}
	}
	if n := strings.Count(cuerpo, `aria-current="page"`); n != 1 || !strings.Contains(cuerpo, `<a href="/" class="`) {
		t.Errorf("hay %d items marcados, quiero solo Inicio", n)
	}
}

// CA-T004-2: con HX-Request la misma ruta responde solo el fragmento y con Vary.
func TestRouter_InicioComoFragmento(t *testing.T) {
	w := pedir(routerDePrueba(), http.MethodGet, "/", map[string]string{"HX-Request": "true"})

	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), "<html") || !strings.Contains(w.Body.String(), "<h1") {
		t.Errorf("estado %d cuerpo %q, quiero 200 y solo el fragmento", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Vary"); got != "HX-Request" {
		t.Errorf("Vary = %q, quiero HX-Request", got)
	}
}

// CA-T004-3, CL-T004-2: con HX-Request y HX-Boosted responde la pagina completa.
func TestRouter_InicioBoosted(t *testing.T) {
	w := pedir(routerDePrueba(), http.MethodGet, "/", map[string]string{"HX-Request": "true", "HX-Boosted": "true"})

	if !strings.Contains(strings.ToLower(w.Body.String()), "<!doctype html>") {
		t.Errorf("una navegacion boosted tiene que recibir la pagina completa: %q", w.Body.String())
	}
}

// RN-T004-7: /health conserva su comportamiento.
func TestRouter_Health(t *testing.T) {
	sano := httpadapter.NuevoRouter(verificadorFalso{}, time.Second, http.FS(fstest.MapFS{}))
	if w := pedir(sano, http.MethodGet, "/health", nil); w.Code != http.StatusOK || strings.TrimSpace(w.Body.String()) != `{"estado":"ok"}` {
		t.Errorf("sano: %d %q, quiero 200 y {\"estado\":\"ok\"}", w.Code, w.Body.String())
	}

	caido := httpadapter.NuevoRouter(verificadorFalso{err: errors.New("base caida")}, time.Second, http.FS(fstest.MapFS{}))
	w := pedir(caido, http.MethodGet, "/health", nil)
	if w.Code != http.StatusServiceUnavailable || strings.TrimSpace(w.Body.String()) != `{"estado":"error"}` {
		t.Errorf("caido: %d %q, quiero 503 y {\"estado\":\"error\"}", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, quiero application/json", ct)
	}
}

// RN-T004-5, CA-T004-4, CL-T004-6: una ruta que no existe responde 404 con la pagina de error, sin
// item marcado. /proyectos es una de ellas hasta que T-011 la implemente.
func TestRouter_RutaInexistente(t *testing.T) {
	for _, ruta := range []string{"/nada", "/proyectos", "/proyectos/1/backlog", "/index.html"} {
		w := pedir(routerDePrueba(), http.MethodGet, ruta, nil)

		if w.Code != http.StatusNotFound {
			t.Errorf("%s: estado = %d, quiero 404", ruta, w.Code)
		}
		cuerpo := w.Body.String()
		if !strings.Contains(cuerpo, "Página no encontrada") || !strings.Contains(strings.ToLower(cuerpo), "<!doctype html>") {
			t.Errorf("%s: no es la pagina de error con layout:\n%s", ruta, cuerpo)
		}
		if strings.Contains(cuerpo, `aria-current="page"`) {
			t.Errorf("%s: la pagina de error no marca ningun item", ruta)
		}
	}
}

// El metodo no cambia el resultado de una ruta que no existe, y / solo acepta GET.
func TestRouter_OtrosMetodos(t *testing.T) {
	for _, c := range []struct{ metodo, ruta string }{{http.MethodPost, "/nada"}, {http.MethodPost, "/"}, {http.MethodDelete, "/"}} {
		w := pedir(routerDePrueba(), c.metodo, c.ruta, nil)
		if w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s: estado = %d, quiero 404 o 405", c.metodo, c.ruta, w.Code)
		}
	}
}

// CL-T004-5: el 404 por HTMX devuelve el fragmento con codigo 404.
func TestRouter_RutaInexistenteComoFragmento(t *testing.T) {
	w := pedir(routerDePrueba(), http.MethodGet, "/nada", map[string]string{"HX-Request": "true"})

	if w.Code != http.StatusNotFound {
		t.Errorf("estado = %d, quiero 404", w.Code)
	}
	if strings.Contains(w.Body.String(), "<html") || !strings.Contains(w.Body.String(), "Página no encontrada") {
		t.Errorf("cuerpo = %q, quiero el fragmento de la pagina de error", w.Body.String())
	}
}

// Los estaticos se sirven con su tipo de contenido.
func TestRouter_Estaticos(t *testing.T) {
	casos := []struct{ ruta, tipo, cuerpo string }{
		{"/static/css/app.css", "text/css", "body{color:red}"},
		{"/static/js/htmx.min.js", "text/javascript", "var htmx={};"},
	}
	for _, c := range casos {
		w := pedir(routerDePrueba(), http.MethodGet, c.ruta, nil)

		if w.Code != http.StatusOK || w.Body.String() != c.cuerpo {
			t.Errorf("%s: %d %q, quiero 200 y %q", c.ruta, w.Code, w.Body.String(), c.cuerpo)
		}
		if got := w.Header().Get("Content-Type"); !strings.HasPrefix(got, c.tipo) {
			t.Errorf("%s: Content-Type = %q, quiero %s", c.ruta, got, c.tipo)
		}
	}
}

// RN-T004-6, CL-T004-3: /static/ nunca lista directorios.
func TestRouter_EstaticosNoListanDirectorios(t *testing.T) {
	for _, ruta := range []string{"/static/", "/static/css/", "/static/css", "/static/js/"} {
		w := pedir(routerDePrueba(), http.MethodGet, ruta, nil)

		if w.Code != http.StatusNotFound {
			t.Errorf("%s: estado = %d, quiero 404", ruta, w.Code)
		}
		if strings.Contains(w.Body.String(), "app.css") || strings.Contains(w.Body.String(), "htmx.min.js") {
			t.Errorf("%s: lista el contenido del directorio: %q", ruta, w.Body.String())
		}
	}
}

// CL-T004-4: un estatico que no existe da 404 sin pasar por la pagina 404 con layout.
func TestRouter_EstaticoInexistente(t *testing.T) {
	w := pedir(routerDePrueba(), http.MethodGet, "/static/css/otro.css", nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("estado = %d, quiero 404", w.Code)
	}
	if strings.Contains(w.Body.String(), "<html") || strings.Contains(w.Body.String(), "Página no encontrada") {
		t.Errorf("un estatico inexistente no pasa por la pagina de error: %q", w.Body.String())
	}
}
