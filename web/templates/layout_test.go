package plantillas_test

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/a-h/templ"

	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// renderizar dibuja el layout con "contenido" como hijo y devuelve el HTML.
func renderizar(t *testing.T, p plantillas.Pagina, contenido string) string {
	t.Helper()
	var salida bytes.Buffer
	ctx := templ.WithChildren(context.Background(), templ.Raw(contenido))
	if err := plantillas.Layout(p).Render(ctx, &salida); err != nil {
		t.Fatalf("renderizar el layout: %v", err)
	}
	return salida.String()
}

// etiqueta devuelve la primera etiqueta que cumple el patron, o corta el test.
// Los patrones no dependen del orden de los atributos.
func etiqueta(t *testing.T, html, patron string) string {
	t.Helper()
	encontrada := regexp.MustCompile(patron).FindString(html)
	if encontrada == "" {
		t.Fatalf("no hay ninguna etiqueta que cumpla %q en:\n%s", patron, html)
	}
	return encontrada
}

// etiquetaEnlace devuelve la etiqueta <a> con ese href exacto.
func etiquetaEnlace(t *testing.T, html, href string) string {
	t.Helper()
	return etiqueta(t, html, `<a [^>]*href="`+regexp.QuoteMeta(href)+`"[^>]*>`)
}

func proyectoDeEjemplo() *plantillas.ProyectoActual {
	return &plantillas.ProyectoActual{ID: "p1", Nombre: "Demo"}
}

// T-004: toda pagina completa tiene la estructura HTML base accesible.
func TestLayout_EstructuraBase(t *testing.T) {
	html := renderizar(t, plantillas.Pagina{Titulo: "Proyectos"}, "<p>hola</p>")

	if !strings.HasPrefix(strings.ToLower(html), "<!doctype html>") {
		t.Errorf("el documento no empieza con <!doctype html>: %.60q", html)
	}

	patrones := []struct{ descripcion, patron string }{
		{"idioma del documento", `<html[^>]*lang="es"`},
		{"codificacion", `<meta[^>]*charset="utf-8"`},
		{"viewport", `<meta[^>]*name="viewport"`},
		{"titulo de la pestana", `<title>Proyectos \| Software Metrics</title>`},
		{"hoja de estilos", `<link[^>]*href="/static/css/app\.css"`},
		{"script de htmx", `<script[^>]*src="/static/js/htmx\.min\.js"`},
		{"cabecera", `<header[ >]`},
		{"navegacion principal", `<nav[^>]*aria-label="Principal"`},
		{"contenido principal", `<main[ >]`},
	}
	for _, p := range patrones {
		if !regexp.MustCompile(p.patron).MatchString(html) {
			t.Errorf("falta %s: no aparece %q", p.descripcion, p.patron)
		}
	}
}

// T-004: el titulo se escapa, para que un "&" o un "<" no rompan el HTML.
func TestLayout_TituloEscapado(t *testing.T) {
	html := renderizar(t, plantillas.Pagina{Titulo: "A & B"}, "")
	if !strings.Contains(html, "<title>A &amp; B | Software Metrics</title>") {
		t.Errorf("el titulo no se escapo bien en:\n%s", html)
	}
}

// T-004: el contenido de cada pagina va dentro de <main>.
func TestLayout_ContenidoDentroDeMain(t *testing.T) {
	html := renderizar(t, plantillas.Pagina{}, "<p>contenido-de-prueba</p>")

	inicio := strings.Index(html, "<main")
	fin := strings.Index(html, "</main>")
	posicion := strings.Index(html, "contenido-de-prueba")
	if inicio < 0 || fin < inicio || posicion < inicio || posicion > fin {
		t.Errorf("el contenido no esta dentro de <main> (main=%d..%d, contenido=%d)", inicio, fin, posicion)
	}
}

// T-004: hay un enlace para saltar al contenido, antes de la navegacion.
func TestLayout_EnlaceSaltarAlContenido(t *testing.T) {
	html := renderizar(t, plantillas.Pagina{}, "")

	salto := strings.Index(html, `href="#contenido"`)
	navegacion := strings.Index(html, "<nav")
	if salto < 0 || navegacion < 0 || salto > navegacion {
		t.Errorf("el enlace 'saltar al contenido' debe existir y venir antes de la navegacion (salto=%d, nav=%d)", salto, navegacion)
	}
	etiqueta(t, html, `<main[^>]*id="contenido"`)
}

// T-004: hay un lugar reservado para mensajes y errores, antes del contenido.
func TestLayout_EspacioDeMensajes(t *testing.T) {
	html := renderizar(t, plantillas.Pagina{}, "<p>contenido-de-prueba</p>")

	mensajes := etiqueta(t, html, `<div[^>]*id="mensajes"[^>]*>`)
	for _, atributo := range []string{`role="status"`, `aria-live="polite"`} {
		if !strings.Contains(mensajes, atributo) {
			t.Errorf("el espacio de mensajes no tiene %s: %s", atributo, mensajes)
		}
	}

	main := strings.Index(html, "<main")
	espacio := strings.Index(html, `id="mensajes"`)
	contenido := strings.Index(html, "contenido-de-prueba")
	if !(main < espacio && espacio < contenido) {
		t.Errorf("el espacio de mensajes debe estar dentro de <main>, antes del contenido (main=%d, mensajes=%d, contenido=%d)",
			main, espacio, contenido)
	}
}

// T-004: solo el enlace de la pagina actual lleva aria-current.
func TestLayout_ItemActivo(t *testing.T) {
	casos := []struct {
		nombre     string
		pagina     plantillas.Pagina
		hrefActivo string // vacio: ninguno
	}{
		{"caso normal: Proyectos activo", plantillas.Pagina{Activo: plantillas.SeccionProyectos}, "/proyectos"},
		{"caso alternativo: Inicio activo", plantillas.Pagina{Activo: plantillas.SeccionInicio}, "/"},
		{
			"caso alternativo: Backlog activo dentro de un proyecto",
			plantillas.Pagina{Activo: plantillas.SeccionBacklog, Proyecto: proyectoDeEjemplo()},
			"/proyectos/p1/backlog",
		},
		{
			"caso alternativo: Resumen activo dentro de un proyecto",
			plantillas.Pagina{Activo: plantillas.SeccionResumen, Proyecto: proyectoDeEjemplo()},
			"/proyectos/p1",
		},
		{"caso limite: ninguna seccion activa", plantillas.Pagina{}, ""},
		{
			"caso limite: seccion de proyecto sin proyecto no marca nada",
			plantillas.Pagina{Activo: plantillas.SeccionBacklog},
			"",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			html := renderizar(t, c.pagina, "")

			esperados := 0
			if c.hrefActivo != "" {
				esperados = 1
				if !strings.Contains(etiquetaEnlace(t, html, c.hrefActivo), `aria-current="page"`) {
					t.Errorf("el enlace %q deberia tener aria-current=\"page\"", c.hrefActivo)
				}
			}
			if obtenidos := strings.Count(html, `aria-current="page"`); obtenidos != esperados {
				t.Errorf("enlaces con aria-current = %d, se esperaba %d", obtenidos, esperados)
			}
		})
	}
}

// T-004: el menu del proyecto solo aparece dentro de un proyecto.
func TestLayout_MenuDeProyecto(t *testing.T) {
	t.Run("caso normal: con proyecto aparecen su nombre y sus seis destinos", func(t *testing.T) {
		html := renderizar(t, plantillas.Pagina{Proyecto: proyectoDeEjemplo()}, "")

		etiqueta(t, html, `<nav[^>]*aria-label="Proyecto"`)
		if !strings.Contains(html, "Demo") {
			t.Errorf("no aparece el nombre del proyecto en:\n%s", html)
		}
		for _, href := range []string{
			"/proyectos/p1", "/proyectos/p1/backlog", "/proyectos/p1/sprints", "/proyectos/p1/esfuerzo",
			"/proyectos/p1/metricas", "/proyectos/p1/defectos",
		} {
			etiquetaEnlace(t, html, href)
		}
	})

	t.Run("caso limite: sin proyecto no hay menu de proyecto", func(t *testing.T) {
		html := renderizar(t, plantillas.Pagina{}, "")

		if regexp.MustCompile(`<nav[^>]*aria-label="Proyecto"`).MatchString(html) {
			t.Errorf("hay un menu de proyecto en una pagina sin proyecto:\n%s", html)
		}
		if strings.Contains(html, "/backlog") {
			t.Errorf("hay enlaces de proyecto en una pagina sin proyecto:\n%s", html)
		}
	})

	t.Run("error: el nombre del proyecto se escapa", func(t *testing.T) {
		peligroso := &plantillas.ProyectoActual{ID: "p1", Nombre: "<script>alert(1)</script>"}
		html := renderizar(t, plantillas.Pagina{Proyecto: peligroso}, "")

		if strings.Contains(html, "<script>alert(1)</script>") {
			t.Errorf("el nombre del proyecto se escribio sin escapar:\n%s", html)
		}
		if !strings.Contains(html, "&lt;script&gt;") {
			t.Errorf("no aparece el nombre escapado en:\n%s", html)
		}
	})
}
