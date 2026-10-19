package httpadapter_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpadapter "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/http"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/errores"
)

var (
	errTitulo      = errors.New("el titulo es obligatorio")
	errPrioridad   = errors.New("la prioridad tiene que ser must, should, could o wont")
	errNoExiste    = errors.New("el item no existe")
	errCerrado     = errors.New("el sprint esta cerrado")
	errConEtiqueta = errors.New(`el titulo "<script>alert(1)</script>" & compania`)
	errDeLaBase    = errors.New("pq: password authentication failed for user metrics")
)

func catalogoParaErrores() *errores.Catalogo {
	return errores.NuevoCatalogo(
		errores.Regla{Err: errTitulo, Tipo: errores.Validacion, Campo: "titulo"},
		errores.Regla{Err: errPrioridad, Tipo: errores.Validacion, Campo: "prioridad"},
		errores.Regla{Err: errNoExiste, Tipo: errores.NoEncontrado},
		errores.Regla{Err: errCerrado, Tipo: errores.Conflicto},
		errores.Regla{Err: errConEtiqueta, Tipo: errores.Validacion, Campo: "titulo"},
	)
}

// manejadorConRegistro devuelve el manejador y el buffer donde escribe su registro.
func manejadorConRegistro() (*httpadapter.ManejadorDeErrores, *bytes.Buffer) {
	var registro bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&registro, nil))
	return httpadapter.NuevoManejadorDeErrores(catalogoParaErrores(), logger), &registro
}

func responder(t *testing.T, err error, accept string) (*httptest.ResponseRecorder, *bytes.Buffer) {
	t.Helper()
	m, registro := manejadorConRegistro()
	r := httptest.NewRequest(http.MethodPost, "/proyectos/1/items", nil)
	if accept != "" {
		r.Header.Set("Accept", accept)
	}
	w := httptest.NewRecorder()
	m.ResponderError(w, r, err)
	return w, registro
}

// RN-T005-1: cada tipo tiene su estado HTTP.
func TestEstadoHTTP(t *testing.T) {
	casos := map[errores.Tipo]int{
		errores.Validacion:   http.StatusUnprocessableEntity,
		errores.NoEncontrado: http.StatusNotFound,
		errores.Conflicto:    http.StatusConflict,
		errores.Interno:      http.StatusInternalServerError,
	}
	for tipo, quiero := range casos {
		if got := httpadapter.EstadoHTTP(tipo); got != quiero {
			t.Errorf("EstadoHTTP(%q) = %d, quiero %d", tipo, got, quiero)
		}
	}
}

// CA-T005-1, RN-T005-15: un error de validacion responde 422 con su campo y su mensaje.
func TestResponderError_ValidacionEnJSON(t *testing.T) {
	w, _ := responder(t, errTitulo, "application/json")

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("estado = %d, quiero 422", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	const quiero = `{"tipo":"validacion","errores":[{"campo":"titulo","mensaje":"el titulo es obligatorio"}]}`
	if got := strings.TrimSpace(w.Body.String()); got != quiero {
		t.Errorf("cuerpo = %s, quiero %s", got, quiero)
	}
}

// Spec, seccion 3: el campo se omite cuando el error es general.
func TestResponderError_JSONSinCampo(t *testing.T) {
	w, _ := responder(t, errCerrado, "application/json")

	const quiero = `{"tipo":"conflicto","errores":[{"mensaje":"el sprint esta cerrado"}]}`
	if got := strings.TrimSpace(w.Body.String()); got != quiero {
		t.Errorf("cuerpo = %s, quiero %s", got, quiero)
	}
}

// CA-T005-3: inexistencia 404, estado 409.
func TestResponderError_EstadoPorTipo(t *testing.T) {
	casos := []struct {
		err    error
		quiero int
	}{
		{errNoExiste, http.StatusNotFound},
		{errCerrado, http.StatusConflict},
		{errTitulo, http.StatusUnprocessableEntity},
		{errDeLaBase, http.StatusInternalServerError},
	}
	for _, c := range casos {
		w, _ := responder(t, c.err, "")
		if w.Code != c.quiero {
			t.Errorf("responder(%v) estado = %d, quiero %d", c.err, w.Code, c.quiero)
		}
	}
}

// CA-T005-2: todos los errores unidos salen, en orden y cada uno con su campo.
func TestResponderError_ErroresUnidos(t *testing.T) {
	w, _ := responder(t, errors.Join(errPrioridad, errTitulo), "application/json")

	var cuerpo struct {
		Tipo    string `json:"tipo"`
		Errores []struct {
			Campo   string `json:"campo"`
			Mensaje string `json:"mensaje"`
		} `json:"errores"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("el cuerpo no es JSON: %v", err)
	}
	if len(cuerpo.Errores) != 2 || cuerpo.Errores[0].Campo != "prioridad" || cuerpo.Errores[1].Campo != "titulo" {
		t.Errorf("errores = %+v, quiero prioridad y despues titulo", cuerpo.Errores)
	}
}

// RN-T005-15: el fragmento HTML tiene una estructura fija y sin estilos.
func TestResponderError_FragmentoHTML(t *testing.T) {
	w, _ := responder(t, errors.Join(errTitulo, errCerrado), "text/html")

	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	const quiero = `<div role="alert" data-tipo="conflicto"><ul>` +
		`<li data-campo="titulo">el titulo es obligatorio</li>` +
		`<li>el sprint esta cerrado</li>` +
		`</ul></div>`
	if got := w.Body.String(); got != quiero {
		t.Errorf("cuerpo = %s, quiero %s", got, quiero)
	}
}

// CL-T005-12: se elige JSON solo si el Accept lo menciona y no menciona text/html.
func TestResponderError_NegociacionDelFormato(t *testing.T) {
	casos := []struct {
		accept   string
		esperado string
	}{
		{"", "text/html; charset=utf-8"},
		{"*/*", "text/html; charset=utf-8"},
		{"text/html", "text/html; charset=utf-8"},
		{"text/html,application/json;q=0.9", "text/html; charset=utf-8"},
		{"application/json", "application/json; charset=utf-8"},
		{"Application/JSON", "application/json; charset=utf-8"},
		{"application/json, text/plain, */*", "application/json; charset=utf-8"},
	}
	for _, c := range casos {
		w, _ := responder(t, errTitulo, c.accept)
		if got := w.Header().Get("Content-Type"); got != c.esperado {
			t.Errorf("Accept %q: Content-Type = %q, quiero %q", c.accept, got, c.esperado)
		}
	}
}

// CL-T005-15: el texto se escapa en el fragmento y viaja como texto en el JSON.
func TestResponderError_EscapaElTexto(t *testing.T) {
	w, _ := responder(t, errConEtiqueta, "text/html")
	if strings.Contains(w.Body.String(), "<script>") {
		t.Errorf("el fragmento tiene HTML sin escapar: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "&lt;script&gt;") {
		t.Errorf("el fragmento no tiene el texto escapado: %s", w.Body.String())
	}

	w, _ = responder(t, errConEtiqueta, "application/json")
	var cuerpo struct {
		Errores []struct {
			Mensaje string `json:"mensaje"`
		} `json:"errores"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("el cuerpo no es JSON: %v", err)
	}
	if got := cuerpo.Errores[0].Mensaje; got != errConEtiqueta.Error() {
		t.Errorf("mensaje = %q, quiero %q", got, errConEtiqueta.Error())
	}
}

// CA-T005-6, RN-T005-7 y RN-T005-16: un error desconocido responde 500 con el mensaje fijo, nada
// del original llega al cliente y el detalle completo queda en el registro.
func TestResponderError_InternoNoFiltraNiOculta(t *testing.T) {
	err := fmt.Errorf("leer el item 7: %w", errDeLaBase)
	for _, accept := range []string{"application/json", "text/html"} {
		w, registro := responder(t, err, accept)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("estado = %d, quiero 500", w.Code)
		}
		if strings.Contains(w.Body.String(), "password") || strings.Contains(w.Body.String(), "metrics") {
			t.Errorf("la respuesta filtra el detalle: %s", w.Body.String())
		}
		if !strings.Contains(w.Body.String(), errores.MensajeInterno) {
			t.Errorf("la respuesta no tiene el mensaje fijo: %s", w.Body.String())
		}
		log := registro.String()
		for _, quiero := range []string{"level=ERROR", "method=POST", "path=/proyectos/1/items", "leer el item 7: pq: password authentication failed"} {
			if !strings.Contains(log, quiero) {
				t.Errorf("el registro no tiene %q: %s", quiero, log)
			}
		}
	}
}

// RN-T005-16: solo los errores internos se registran.
func TestResponderError_LosDemasTiposNoSeRegistran(t *testing.T) {
	for _, err := range []error{errTitulo, errNoExiste, errCerrado} {
		_, registro := responder(t, err, "")
		if registro.Len() != 0 {
			t.Errorf("%v se registro: %s", err, registro.String())
		}
	}
}

// RN-T005-17, CL-T005-1: pasar nil es un error de programacion: responde interno y lo registra.
func TestResponderError_NilEsInterno(t *testing.T) {
	w, registro := responder(t, nil, "application/json")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("estado = %d, quiero 500", w.Code)
	}
	if !strings.Contains(registro.String(), "level=ERROR") {
		t.Errorf("no se registro: %q", registro.String())
	}
}

// El ErrorDeValidacion de un formulario sale con el mismo formato que un error del dominio (CA-T005-4).
func TestResponderError_ErrorDeFormulario(t *testing.T) {
	err := errors.Join(
		&errores.ErrorDeValidacion{Errores: []errores.ErrorDeCampo{{Campo: "horas", Mensaje: "el campo horas tiene que ser un numero"}}},
		errTitulo,
	)
	w, _ := responder(t, err, "application/json")

	const quiero = `{"tipo":"validacion","errores":[{"campo":"horas","mensaje":"el campo horas tiene que ser un numero"},{"campo":"titulo","mensaje":"el titulo es obligatorio"}]}`
	if w.Code != http.StatusUnprocessableEntity || strings.TrimSpace(w.Body.String()) != quiero {
		t.Errorf("estado %d cuerpo %s, quiero 422 y %s", w.Code, w.Body.String(), quiero)
	}
}

// Un logger nulo usa el de por defecto en vez de romper.
func TestNuevoManejadorDeErrores_SinLogger(t *testing.T) {
	m := httpadapter.NuevoManejadorDeErrores(catalogoParaErrores(), nil)
	w := httptest.NewRecorder()
	m.ResponderError(w, httptest.NewRequest(http.MethodGet, "/", nil), errTitulo)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("estado = %d, quiero 422", w.Code)
	}
}
