package plantillas_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"

	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

func dibujar(t *testing.T, c templ.Component) string {
	t.Helper()
	var salida bytes.Buffer
	if err := c.Render(context.Background(), &salida); err != nil {
		t.Fatalf("dibujar el componente: %v", err)
	}
	return salida.String()
}

// CA-T004-1: el inicio tiene un titulo con el nombre de la aplicacion.
func TestInicio(t *testing.T) {
	html := dibujar(t, plantillas.Inicio())

	etiqueta(t, html, `<h1[^>]*>Software Metrics &amp; Estimation</h1>`)
	if strings.Contains(strings.ToLower(html), "<html") {
		t.Errorf("el inicio es un contenido, no una pagina completa: %s", html)
	}
}

// CA-T004-4: la pagina de error 404 dice que no se encontro y ofrece volver al inicio.
func TestNoEncontrado(t *testing.T) {
	html := dibujar(t, plantillas.NoEncontrado())

	etiqueta(t, html, `<h1[^>]*>Página no encontrada</h1>`)
	etiquetaEnlace(t, html, "/")
}
