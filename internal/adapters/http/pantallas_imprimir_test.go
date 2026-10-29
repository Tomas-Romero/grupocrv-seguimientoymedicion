package httpadapter_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	httpadapter "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/http"
)

// Vista imprimible del reporte (spec US-034). El reloj de los tests es el 2026-10-20.

// paginasDelReporte devuelve el HTML de los dos reportes de US-033 sobre el proyecto de proyectoParaReportes.
func paginasDelReporte(t *testing.T) map[string]string {
	t.Helper()
	e, proyecto, sprints := proyectoParaReportes(t)
	paginas := map[string]string{}
	for nombre, ruta := range map[string]string{
		"proyecto": rutaDelReporte(proyecto),
		"sprint":   rutaDelReporteDelSprint(proyecto, sprints[0]),
	} {
		w := e.get(ruta)
		if w.Code != http.StatusOK {
			t.Fatalf("GET %s = %d %s", ruta, w.Code, w.Body.String())
		}
		paginas[nombre] = w.Body.String()
	}
	return paginas
}

var botonDeImprimir = regexp.MustCompile(`<button[^>]*id="imprimir"[^>]*>([^<]*)</button>`)

// US-034 / CA-034-1, RN-034-1, RN-034-2, RN-034-3: los dos reportes tienen el boton "Imprimir o guardar como PDF", oculto hasta que lo
// muestra imprimir.js, sin onclick y sin script en linea; el comportamiento vive en el archivo estatico.
func TestReporte_BotonDeImprimir(t *testing.T) {
	for nombre, html := range paginasDelReporte(t) {
		t.Run(nombre, func(t *testing.T) {
			boton := botonDeImprimir.FindStringSubmatch(html)
			if boton == nil {
				t.Fatalf("no esta el boton de imprimir:\n%s", html)
			}
			if strings.TrimSpace(boton[1]) != "Imprimir o guardar como PDF" {
				t.Errorf("texto del boton = %q", boton[1])
			}
			if !regexp.MustCompile(`\shidden[\s>]`).MatchString(boton[0]) {
				t.Errorf("el boton no arranca oculto: %s", boton[0])
			}
			if strings.Contains(boton[0], "onclick") || !strings.Contains(boton[0], `type="button"`) {
				t.Errorf("el boton tiene onclick o no es type=button: %s", boton[0])
			}
			if !strings.Contains(html, `<script src="/static/js/imprimir.js" defer></script>`) {
				t.Error("la pagina no carga /static/js/imprimir.js")
			}
			if regexp.MustCompile(`<script(?:\s[^>]*)?>[^<]+</script>`).MatchString(html) {
				t.Error("la pagina tiene un script en linea")
			}
		})
	}
}

var encabezadoDeImpresion = regexp.MustCompile(`(?s)<div[^>]*class="[^"]*solo-impresion[^"]*"[^>]*>(.*?)</div>`)

// US-034 / RN-034-9: el encabezado que se ve solo al imprimir tiene el proyecto, el titulo del reporte (con el sprint, en su reporte) y la
// fecha del reloj del servidor en dd/mm/aaaa.
func TestReporte_EncabezadoDeImpresion(t *testing.T) {
	paginas := paginasDelReporte(t)
	esperados := map[string][]string{
		"proyecto": {"Demo", "Reporte del proyecto", "Generado el 20/10/2026"},
		"sprint":   {"Demo", "Reporte del Sprint 1", "Generado el 20/10/2026"},
	}
	for nombre, textos := range esperados {
		encabezado := encabezadoDeImpresion.FindStringSubmatch(paginas[nombre])
		if encabezado == nil {
			t.Errorf("%s: no esta el encabezado de impresion", nombre)
			continue
		}
		for _, texto := range textos {
			if !strings.Contains(encabezado[1], texto) {
				t.Errorf("%s: el encabezado de impresion no dice %q: %s", nombre, texto, encabezado[1])
			}
		}
	}
}

// US-034 / CA-034-2, RN-034-5: lo que no se imprime se marca con la clase no-imprimir: la cabecera del layout, el enlace "Saltar al
// contenido", el area #mensajes, el boton de imprimir y el enlace para volver.
func TestReporte_LoQueNoSeImprime(t *testing.T) {
	marcados := map[string]*regexp.Regexp{
		"cabecera":            regexp.MustCompile(`<header[^>]*class="[^"]*no-imprimir`),
		"saltar al contenido": regexp.MustCompile(`<a[^>]*href="#contenido"[^>]*class="[^"]*no-imprimir`),
		"mensajes":            regexp.MustCompile(`<div[^>]*id="mensajes"[^>]*class="[^"]*no-imprimir`),
		"boton":               regexp.MustCompile(`<button[^>]*id="imprimir"[^>]*class="[^"]*no-imprimir`),
		"volver":              regexp.MustCompile(`<a[^>]*class="[^"]*no-imprimir[^"]*"[^>]*>Volver al `),
	}
	for nombre, html := range paginasDelReporte(t) {
		for que, patron := range marcados {
			if !patron.MatchString(html) {
				t.Errorf("%s: %s no lleva la clase no-imprimir", nombre, que)
			}
		}
	}
}

// US-034 / CL-034-7: un reporte que da 404 no tiene boton de imprimir ni encabezado de impresion.
func TestReporte_NoEncontradoSinBoton(t *testing.T) {
	e := nuevoEntorno(t)
	w := e.get(rutaDelReporte("00000000-0000-0000-0000-00000000dead"))
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET = %d, se esperaba 404", w.Code)
	}
	if html := w.Body.String(); strings.Contains(html, `id="imprimir"`) || strings.Contains(html, "solo-impresion") {
		t.Errorf("la pagina 404 tiene el boton o el encabezado de impresion:\n%s", html)
	}
}

// US-034 / RN-034-1, RN-034-2: el servidor sirve imprimir.js desde web/static. Muestra el boton y al tocarlo llama a window.print(); no hace
// peticiones ni cambia la pagina.
func TestImprimirJS(t *testing.T) {
	router := httpadapter.NuevoRouter(verificadorFalso{}, time.Second, http.Dir("../../../web/static"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/static/js/imprimir.js", nil))

	if w.Code != http.StatusOK || !strings.Contains(w.Header().Get("Content-Type"), "javascript") {
		t.Fatalf("GET /static/js/imprimir.js = %d %q", w.Code, w.Header().Get("Content-Type"))
	}
	js := w.Body.String()
	for _, necesario := range []string{"window.print()", "addEventListener", "hidden"} {
		if !strings.Contains(js, necesario) {
			t.Errorf("imprimir.js no tiene %q", necesario)
		}
	}
	for _, prohibido := range []string{"fetch(", "XMLHttpRequest", "innerHTML", "location"} {
		if strings.Contains(js, prohibido) {
			t.Errorf("imprimir.js usa %q: solo tiene que imprimir", prohibido)
		}
	}
}

// reglasDeImpresion devuelve el bloque @media print de input.css, que es de donde Tailwind arma el app.css que sirve el sitio
// (app.css se genera con make css y no se versiona).
func reglasDeImpresion(t *testing.T) string {
	t.Helper()
	css, err := os.ReadFile("../../../web/static/css/input.css")
	if err != nil {
		t.Fatalf("leer input.css: %v", err)
	}
	inicio := strings.Index(string(css), "@media print")
	if inicio < 0 {
		t.Fatal("input.css no tiene un bloque @media print")
	}
	return strings.Join(strings.Fields(string(css[inicio:])), " ")
}

// US-034 / CA-034-2, CA-034-3, RN-034-4, RN-034-6 a RN-034-8: la hoja de estilos del sitio tiene las reglas de impresion. A4 vertical con
// margenes; lo marcado no se imprime y el encabezado de impresion si; una seccion que entra en una pagina no se corta, un titulo no queda
// solo al pie; una fila no se parte y el encabezado de una tabla se repite en cada pagina; el contenido usa todo el ancho.
func TestEstilosDeImpresion(t *testing.T) {
	reglas := reglasDeImpresion(t)
	for _, regla := range []string{
		"@page { size: A4 portrait; margin: 15mm; }",
		".no-imprimir { display: none !important; }",
		".solo-impresion { display: block !important; }",
		".reporte section { break-inside: avoid; }",
		".reporte h2, .reporte h3 { break-after: avoid; }",
		".reporte tr { break-inside: avoid; }",
		".reporte thead { display: table-header-group; }",
		"main { max-width: none; padding: 0; }",
	} {
		if !strings.Contains(reglas, regla) {
			t.Errorf("las reglas de impresion no tienen %q", regla)
		}
	}
	css, _ := os.ReadFile("../../../web/static/css/input.css")
	if !strings.Contains(strings.Join(strings.Fields(string(css)), " "), ".solo-impresion { display: none; }") {
		t.Error("el encabezado de impresion no esta oculto en pantalla")
	}
}
