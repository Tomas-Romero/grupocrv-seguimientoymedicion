package httpadapter

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/a-h/templ"

	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// Responder escribe una pantalla (spec T-004): la pagina completa con el layout, o solo el
// fragmento si la pide HTMX (RN-T004-1). El codigo HTTP lo decide quien llama (RN-T004-3).
//
// Renderiza a un buffer antes de escribir nada: si la plantilla falla, responde un 500
// limpio en vez de una pagina cortada a la mitad, y registra el error (RN-T004-8).
func Responder(w http.ResponseWriter, r *http.Request, estado int, pagina plantillas.Pagina, contenido templ.Component) {
	var componente templ.Component = contenido
	if !esFragmento(r) {
		componente = templ.ComponentFunc(func(ctx context.Context, salida io.Writer) error {
			return plantillas.Layout(pagina).Render(templ.WithChildren(ctx, contenido), salida)
		})
	}

	var cuerpo bytes.Buffer
	if err := componente.Render(r.Context(), &cuerpo); err != nil {
		slog.Error("no se pudo dibujar la pantalla", "method", r.Method, "path", r.URL.Path, "error", err)
		http.Error(w, "error interno", http.StatusInternalServerError)
		return
	}

	// El contenido depende de HX-Request: las caches no pueden mezclar pagina y fragmento.
	w.Header().Add("Vary", "HX-Request")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(estado)
	_, _ = w.Write(cuerpo.Bytes())
}

// esFragmento es verdadero si y solo si la peticion trae HX-Request: true y no trae
// HX-Boosted: true. Una navegacion "boosted" necesita la pagina completa (RN-T004-1).
func esFragmento(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Boosted") != "true"
}
