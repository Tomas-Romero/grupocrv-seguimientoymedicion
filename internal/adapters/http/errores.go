package httpadapter

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/errores"
)

// ManejadorDeErrores responde los errores de la aplicacion siempre igual (spec T-005):
// clasifica el error con el catalogo, elige el estado HTTP de su tipo y lo escribe en
// JSON o en un fragmento HTML segun el Accept de la peticion.
type ManejadorDeErrores struct {
	catalogo *errores.Catalogo
	log      *slog.Logger
}

// NuevoManejadorDeErrores arma el manejador. Sin logger usa el de por defecto.
func NuevoManejadorDeErrores(catalogo *errores.Catalogo, log *slog.Logger) *ManejadorDeErrores {
	if log == nil {
		log = slog.Default()
	}
	return &ManejadorDeErrores{catalogo: catalogo, log: log}
}

// EstadoHTTP devuelve el estado de cada tipo de error (RN-T005-1). Un tipo que no
// conoce es un error interno.
func EstadoHTTP(tipo errores.Tipo) int {
	switch tipo {
	case errores.Validacion:
		return http.StatusUnprocessableEntity
	case errores.NoEncontrado:
		return http.StatusNotFound
	case errores.Conflicto:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

type errorEnJSON struct {
	Campo   string `json:"campo,omitempty"`
	Mensaje string `json:"mensaje"`
}

type problemaEnJSON struct {
	Tipo    errores.Tipo  `json:"tipo"`
	Errores []errorEnJSON `json:"errores"`
}

// fragmentoDeError es la estructura fija del error en HTML, sin estilos: el estilo lo
// pone el layout (T-004). html/template escapa todo el texto que muestra (RN-T005-15).
var fragmentoDeError = template.Must(template.New("error").Parse(
	`<div role="alert" data-tipo="{{.Tipo}}"><ul>` +
		`{{range .Errores}}<li{{if .Campo}} data-campo="{{.Campo}}"{{end}}>{{.Mensaje}}</li>{{end}}` +
		`</ul></div>`))

// ResponderError clasifica el error y lo escribe como respuesta. Solo registra los
// errores internos, con el detalle completo (RN-T005-16): el cliente nunca lo ve.
// Con err nil responde un error interno, porque es un error de programacion (RN-T005-17).
func (m *ManejadorDeErrores) ResponderError(w http.ResponseWriter, r *http.Request, err error) {
	problema := m.catalogo.Clasificar(err)
	if err == nil {
		problema = errores.Problema{Tipo: errores.Interno, Errores: []errores.ErrorDeCampo{{Mensaje: errores.MensajeInterno}}}
		m.log.Error("se pidio responder un error nulo", "method", r.Method, "path", r.URL.Path)
	} else if problema.Tipo == errores.Interno {
		m.log.Error("error interno", "method", r.Method, "path", r.URL.Path, "error", err)
	}

	escribirProblema(w, r, problema)
}

// escribirProblema escribe el estado de su tipo y el cuerpo en JSON o en HTML.
func escribirProblema(w http.ResponseWriter, r *http.Request, problema errores.Problema) {
	if pideJSON(r) {
		cuerpo := problemaEnJSON{Tipo: problema.Tipo, Errores: make([]errorEnJSON, len(problema.Errores))}
		for i, e := range problema.Errores {
			cuerpo.Errores[i] = errorEnJSON{Campo: e.Campo, Mensaje: e.Mensaje}
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(EstadoHTTP(problema.Tipo))
		_ = json.NewEncoder(w).Encode(cuerpo)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(EstadoHTTP(problema.Tipo))
	_ = fragmentoDeError.Execute(w, problema)
}

// pideJSON es verdadero si el Accept menciona application/json y no menciona text/html
// (RN-T005-14). Un navegador o HTMX piden HTML; una llamada de codigo pide JSON.
func pideJSON(r *http.Request) bool {
	accept := strings.ToLower(r.Header.Get("Accept"))
	return strings.Contains(accept, "application/json") && !strings.Contains(accept, "text/html")
}
