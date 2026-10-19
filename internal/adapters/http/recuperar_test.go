package httpadapter_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func peticion(accept string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/sprints/3", nil)
	if accept != "" {
		r.Header.Set("Accept", accept)
	}
	return r
}

// CA-T005-7, RN-T005-18: un panic se responde como error interno uniforme, sin filtrar su valor.
func TestRecuperar_PanicResponde500Uniforme(t *testing.T) {
	m, registro := manejadorConRegistro()
	handler := m.Recuperar(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("indice fuera de rango en la clave secreta")
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, peticion("application/json"))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("estado = %d, quiero 500", w.Code)
	}
	const quiero = `{"tipo":"interno","errores":[{"mensaje":"ocurrio un error inesperado: proba de nuevo en unos minutos"}]}`
	if got := strings.TrimSpace(w.Body.String()); got != quiero {
		t.Errorf("cuerpo = %s, quiero %s", got, quiero)
	}
	if strings.Contains(w.Body.String(), "secreta") {
		t.Errorf("la respuesta filtra el valor del panic: %s", w.Body.String())
	}

	log := registro.String()
	for _, quiero := range []string{"level=ERROR", "method=GET", "path=/sprints/3", "indice fuera de rango en la clave secreta", "stack="} {
		if !strings.Contains(log, quiero) {
			t.Errorf("el registro no tiene %q: %s", quiero, log)
		}
	}
	if n := strings.Count(log, "level=ERROR"); n != 1 {
		t.Errorf("el panic se registro %d veces, quiero 1", n)
	}
}

// CL-T005-13: un panic con un error tambien se responde como interno.
func TestRecuperar_PanicConUnError(t *testing.T) {
	m, registro := manejadorConRegistro()
	handler := m.Recuperar(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(errors.New("puntero nulo"))
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, peticion(""))

	if w.Code != http.StatusInternalServerError || !strings.Contains(w.Body.String(), `data-tipo="interno"`) {
		t.Errorf("estado %d cuerpo %s, quiero 500 con el fragmento interno", w.Code, w.Body.String())
	}
	if !strings.Contains(registro.String(), "puntero nulo") {
		t.Errorf("el registro no tiene el error: %s", registro.String())
	}
}

// CA-T005-7: despues de un panic el servidor sigue atendiendo.
func TestRecuperar_ElServidorSigueAtendiendo(t *testing.T) {
	m, _ := manejadorConRegistro()
	handler := m.Recuperar(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/falla" {
			panic("boom")
		}
		_, _ = w.Write([]byte("bien"))
	}))
	servidor := httptest.NewServer(handler)
	defer servidor.Close()

	respuesta, err := http.Get(servidor.URL + "/falla")
	if err != nil {
		t.Fatalf("la peticion que falla no recibio respuesta: %v", err)
	}
	_ = respuesta.Body.Close()
	if respuesta.StatusCode != http.StatusInternalServerError {
		t.Errorf("estado = %d, quiero 500", respuesta.StatusCode)
	}

	respuesta, err = http.Get(servidor.URL + "/ok")
	if err != nil {
		t.Fatalf("la peticion siguiente fallo: %v", err)
	}
	_ = respuesta.Body.Close()
	if respuesta.StatusCode != http.StatusOK {
		t.Errorf("estado = %d, quiero 200", respuesta.StatusCode)
	}
}

// Sin panic el middleware no toca la respuesta.
func TestRecuperar_SinPanicNoCambiaNada(t *testing.T) {
	m, registro := manejadorConRegistro()
	handler := m.Recuperar(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Prueba", "si")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("creado"))
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, peticion(""))

	if w.Code != http.StatusCreated || w.Body.String() != "creado" || w.Header().Get("X-Prueba") != "si" {
		t.Errorf("respuesta alterada: %d %q %v", w.Code, w.Body.String(), w.Header())
	}
	if registro.Len() != 0 {
		t.Errorf("se registro algo sin panic: %s", registro.String())
	}
}

// CL-T005-14: si la respuesta ya empezo no se escribe otra encima: solo se registra.
func TestRecuperar_PanicConLaRespuestaYaIniciada(t *testing.T) {
	m, registro := manejadorConRegistro()
	handler := m.Recuperar(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("parcial"))
		panic("a mitad de la respuesta")
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, peticion(""))

	if w.Code != http.StatusAccepted || w.Body.String() != "parcial" {
		t.Errorf("respuesta = %d %q, quiero 202 y solo lo que ya se escribio", w.Code, w.Body.String())
	}
	if !strings.Contains(registro.String(), "a mitad de la respuesta") {
		t.Errorf("el panic no se registro: %s", registro.String())
	}
}

// Escribir el cuerpo sin llamar a WriteHeader tambien cuenta como respuesta iniciada.
func TestRecuperar_PanicDespuesDeEscribirElCuerpo(t *testing.T) {
	m, _ := manejadorConRegistro()
	handler := m.Recuperar(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("parcial"))
		panic("despues del cuerpo")
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, peticion(""))

	if w.Code != http.StatusOK || w.Body.String() != "parcial" {
		t.Errorf("respuesta = %d %q, quiero 200 y solo lo que ya se escribio", w.Code, w.Body.String())
	}
}

// RN-T005-18: http.ErrAbortHandler es la forma legitima de cortar una respuesta: se vuelve a lanzar.
func TestRecuperar_ErrAbortHandlerSeRelanza(t *testing.T) {
	m, registro := manejadorConRegistro()
	handler := m.Recuperar(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	defer func() {
		got := recover()
		if err, ok := got.(error); !ok || !errors.Is(err, http.ErrAbortHandler) {
			t.Errorf("recover() = %v, quiero http.ErrAbortHandler", got)
		}
		if registro.Len() != 0 {
			t.Errorf("un aborto deliberado no se registra: %s", registro.String())
		}
	}()
	handler.ServeHTTP(httptest.NewRecorder(), peticion(""))
}

// El ResponseController del handler sigue llegando al escritor original (por ejemplo, Flush).
func TestRecuperar_PermiteHacerFlush(t *testing.T) {
	m, _ := manejadorConRegistro()
	var flushError error
	handler := m.Recuperar(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		flushError = http.NewResponseController(w).Flush()
	}))

	handler.ServeHTTP(httptest.NewRecorder(), peticion(""))

	if flushError != nil {
		t.Errorf("Flush() = %v, quiero nil", flushError)
	}
}
