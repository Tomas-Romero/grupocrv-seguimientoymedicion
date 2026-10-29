package httpadapter_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	httpadapter "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/http"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// entorno es el router real de la aplicacion sobre la base en memoria: las pantallas se
// prueban de punta a punta sin Postgres (spec T-011, CA-T011-1).
type entorno struct {
	router http.Handler
	base   *memoria.Base
}

func nuevoEntorno(t *testing.T) entorno {
	t.Helper()
	base := memoria.NuevaBase()
	ahora := func() time.Time { return time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC) }
	casos := httpadapter.CasosDeUso{
		ListarProyectos:            app.NuevoListarProyectos(base),
		ConsultarProyecto:          app.NuevoConsultarProyecto(base),
		CrearProyecto:              app.NuevoCrearProyecto(base),
		ModificarProyecto:          app.NuevoModificarProyecto(base),
		RegistrarIntegrante:        app.NuevoRegistrarIntegrante(base),
		ListarBacklog:              app.NuevoListarBacklog(base),
		CrearItem:                  app.NuevoCrearItemBacklog(base, ahora),
		ConsultarItem:              app.NuevoConsultarItem(base),
		EditarItem:                 app.NuevoEditarItem(base),
		EstimarItem:                app.NuevoEstimarItem(base),
		ListarSprints:              app.NuevoListarSprints(base),
		ConsultarSprint:            app.NuevoConsultarSprint(base),
		CrearSprint:                app.NuevoCrearSprint(base),
		AsignarItem:                app.NuevoAsignarItemASprint(base),
		MarcarCompletado:           app.NuevoMarcarCompletado(base, ahora),
		DesmarcarCompletado:        app.NuevoDesmarcarCompletado(base),
		CerrarSprint:               app.NuevoCerrarSprint(base),
		ConsultarSprintsAnteriores: app.NuevoConsultarSprintsAnteriores(base),
		ConsultarEstado:            app.NuevoConsultarEstadoDelProyecto(base, ahora),
		ConsultarDashboard:         app.NuevoConsultarDashboard(base, ahora),
		RegistrarEsfuerzo:          app.NuevoRegistrarEsfuerzo(base, ahora),
		ConsultarEsfuerzo:          app.NuevoConsultarEsfuerzo(base),
		CompararEsfuerzo:           app.NuevoCompararEsfuerzo(base),
		Ahora:                      ahora,
		IniciarSesionPlanningPoker: app.NuevoIniciarSesionPlanningPoker(base),
		ConsultarSesion:            app.NuevoConsultarSesionPlanningPoker(base),
		VotarEnSesion:              app.NuevoVotarEnSesion(base),
		RevelarVotos:               app.NuevoRevelarVotos(base),
		AbrirNuevaRonda:            app.NuevoAbrirNuevaRonda(base),
		RegistrarAcuerdo:           app.NuevoRegistrarEstimacionAcordada(base),
		ListarDefectos:             app.NuevoListarDefectos(base),
		RegistrarDefecto:           app.NuevoRegistrarDefecto(base),
		CambiarEstadoDefecto:       app.NuevoCambiarEstadoDefecto(base),
	}
	casos = casos.ConReportes()
	manejador := httpadapter.NuevoManejadorDeErrores(httpadapter.CatalogoDeErrores(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	router := httpadapter.NuevoRouter(verificadorFalso{}, time.Second, http.FS(fstest.MapFS{}),
		httpadapter.ConPantallas(httpadapter.NuevasPantallas(casos, manejador)))
	return entorno{router: router, base: base}
}

// enviar hace una peticion; con formulario la manda como POST de formulario. htmx agrega las
// cabeceras que manda HTMX.
func (e entorno) enviar(metodo, ruta string, formulario url.Values, htmx bool) *httptest.ResponseRecorder {
	var cuerpo io.Reader
	if formulario != nil {
		cuerpo = strings.NewReader(formulario.Encode())
	}
	r := httptest.NewRequest(metodo, ruta, cuerpo)
	if formulario != nil {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if htmx {
		r.Header.Set("HX-Request", "true")
	}
	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, r)
	return w
}

func (e entorno) get(ruta string) *httptest.ResponseRecorder {
	return e.enviar(http.MethodGet, ruta, nil, false)
}
func (e entorno) post(ruta string, f url.Values) *httptest.ResponseRecorder {
	return e.enviar(http.MethodPost, ruta, f, false)
}

func datosDeProyecto(nombre string) url.Values {
	return url.Values{"nombre": {nombre}, "descripcion": {"Descripcion de " + nombre}, "fecha_inicio": {"2026-10-01"}, "fecha_fin": {"2026-12-15"}}
}

var idEnLaRuta = regexp.MustCompile(`^/proyectos/([^/]+)$`)

// crearProyecto da de alta un proyecto por la pantalla y devuelve su ID.
func (e entorno) crearProyecto(t *testing.T, nombre string) string {
	t.Helper()
	w := e.post("/proyectos", datosDeProyecto(nombre))
	if w.Code != http.StatusSeeOther {
		t.Fatalf("crear el proyecto %q: estado %d, cuerpo %s", nombre, w.Code, w.Body.String())
	}
	coincide := idEnLaRuta.FindStringSubmatch(w.Header().Get("Location"))
	if coincide == nil {
		t.Fatalf("Location = %q, quiero /proyectos/{id}", w.Header().Get("Location"))
	}
	return coincide[1]
}

func (e entorno) crearHistoria(t *testing.T, proyectoID, titulo, prioridad, criterios string) {
	t.Helper()
	w := e.post("/proyectos/"+proyectoID+"/backlog", url.Values{"titulo": {titulo}, "descripcion": {"Detalle"}, "prioridad": {prioridad}, "criterios": {criterios}})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("crear la historia %q: estado %d, cuerpo %s", titulo, w.Code, w.Body.String())
	}
}

// idDeHistoria busca el ID de una historia por su titulo en la base en memoria.
func (e entorno) idDeHistoria(t *testing.T, proyectoID, titulo string) string {
	t.Helper()
	for _, item := range e.base.Items(proyectoID) {
		if item.Titulo == titulo {
			return item.ID
		}
	}
	t.Fatalf("no hay una historia %q en el proyecto %s", titulo, proyectoID)
	return ""
}

func contiene(t *testing.T, w *httptest.ResponseRecorder, textos ...string) {
	t.Helper()
	for _, quiero := range textos {
		if !strings.Contains(w.Body.String(), quiero) {
			t.Errorf("la respuesta no tiene %q:\n%s", quiero, w.Body.String())
		}
	}
}

func noContiene(t *testing.T, w *httptest.ResponseRecorder, textos ...string) {
	t.Helper()
	for _, no := range textos {
		if strings.Contains(w.Body.String(), no) {
			t.Errorf("la respuesta no deberia tener %q:\n%s", no, w.Body.String())
		}
	}
}

func posicion(t *testing.T, w *httptest.ResponseRecorder, texto string) int {
	t.Helper()
	i := strings.Index(w.Body.String(), texto)
	if i < 0 {
		t.Fatalf("la respuesta no tiene %q:\n%s", texto, w.Body.String())
	}
	return i
}

// CA-T011-1: el recorrido completo funciona de punta a punta con el router real.
func TestPantallas_Recorrido(t *testing.T) {
	e := nuevoEntorno(t)

	if w := e.get("/proyectos"); w.Code != http.StatusOK {
		t.Fatalf("la lista de proyectos responde %d", w.Code)
	}

	id := e.crearProyecto(t, "Demo")
	w := e.get("/proyectos")
	contiene(t, w, "Demo", "/proyectos/"+id, "01/10/2026", "15/12/2026")
	noContiene(t, w, "Todavía no hay proyectos")

	// El resumen muestra los datos y el formulario de integrantes.
	w = e.get("/proyectos/" + id)
	contiene(t, w, "Demo", "Descripcion de Demo", "No hay integrantes", `name="email"`, `name="rol"`)

	// Registrar un integrante vuelve al resumen y lo muestra.
	w = e.post("/proyectos/"+id+"/integrantes", url.Values{"nombre": {"Ana"}, "email": {"Ana@Ejemplo.test"}, "rol": {"product_builder"}})
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/proyectos/"+id {
		t.Fatalf("registrar un integrante: %d %q", w.Code, w.Header().Get("Location"))
	}
	w = e.get("/proyectos/" + id)
	contiene(t, w, "Ana", "ana@ejemplo.test", "Product Builder")
	noContiene(t, w, "No hay integrantes")

	// El backlog vacio, y despues con dos historias ordenadas por prioridad.
	contiene(t, e.get("/proyectos/"+id+"/backlog"), "El backlog está vacío")
	e.crearHistoria(t, id, "Historia opcional", "could", "")
	e.crearHistoria(t, id, "Historia clave", "must", "Se guarda\n\n  Se lista  \n")
	w = e.get("/proyectos/" + id + "/backlog")
	if posicion(t, w, "Historia clave") > posicion(t, w, "Historia opcional") {
		t.Error("el backlog tiene que mostrar primero la historia must")
	}
	contiene(t, w, "—")

	// Estimar y editar.
	clave := e.idDeHistoria(t, id, "Historia clave")
	w = e.post("/proyectos/"+id+"/backlog/"+clave+"/estimacion", url.Values{"estimacion": {"5"}})
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/proyectos/"+id+"/backlog" {
		t.Fatalf("estimar: %d %q", w.Code, w.Header().Get("Location"))
	}
	w = e.post("/proyectos/"+id+"/backlog/"+clave, url.Values{"titulo": {"Historia clave editada"}, "descripcion": {""}, "prioridad": {"should"}, "criterios": {"Uno\nDos"}})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("editar: %d %s", w.Code, w.Body.String())
	}
	w = e.get("/proyectos/" + id + "/backlog")
	contiene(t, w, "Historia clave editada", ">5<")
	items := e.base.Items(id)
	for _, item := range items {
		if item.ID == clave {
			if len(item.Criterios) != 2 || item.Criterios[0] != "Uno" || item.Prioridad != backlog.PrioridadShould {
				t.Errorf("historia editada = %+v", item)
			}
		}
	}
}

// RN-T011-3, CL-T011-2: los proyectos salen por fecha de inicio descendente.
func TestPantallas_ProyectosPorFechaDeInicio(t *testing.T) {
	e := nuevoEntorno(t)
	viejo := datosDeProyecto("Proyecto viejo")
	viejo.Set("fecha_inicio", "2026-01-10")
	nuevo := datosDeProyecto("Proyecto nuevo")
	nuevo.Set("fecha_inicio", "2026-08-10")
	e.post("/proyectos", viejo)
	e.post("/proyectos", nuevo)

	w := e.get("/proyectos")
	if posicion(t, w, "Proyecto nuevo") > posicion(t, w, "Proyecto viejo") {
		t.Error("el proyecto con la fecha de inicio mas reciente tiene que salir primero")
	}
}

// CL-T011-1: sin proyectos ni historias se muestra el texto de lista vacia y no una tabla sin filas.
func TestPantallas_ListasVacias(t *testing.T) {
	e := nuevoEntorno(t)
	w := e.get("/proyectos")
	contiene(t, w, "Todavía no hay proyectos")
	noContiene(t, w, "<table")

	id := e.crearProyecto(t, "Demo")
	w = e.get("/proyectos/" + id + "/backlog")
	contiene(t, w, "El backlog está vacío")
	noContiene(t, w, "<table")
}

// CA-T011-2: con HTMX, guardar responde HX-Redirect con el cuerpo vacio; sin HTMX, 303.
func TestPantallas_RedireccionesConYSinHTMX(t *testing.T) {
	e := nuevoEntorno(t)

	w := e.enviar(http.MethodPost, "/proyectos", datosDeProyecto("Con HTMX"), true)
	redireccion := w.Header().Get("HX-Redirect")
	if w.Code != http.StatusOK || !idEnLaRuta.MatchString(redireccion) || w.Body.Len() != 0 {
		t.Errorf("con HTMX: %d, HX-Redirect=%q, cuerpo=%q; quiero 200, /proyectos/{id} y cuerpo vacio", w.Code, redireccion, w.Body.String())
	}
	if w.Header().Get("Location") != "" {
		t.Errorf("con HTMX no tiene que haber Location: %q", w.Header().Get("Location"))
	}

	w = e.enviar(http.MethodPost, "/proyectos", datosDeProyecto("Sin HTMX"), false)
	if w.Code != http.StatusSeeOther || !idEnLaRuta.MatchString(w.Header().Get("Location")) || w.Header().Get("HX-Redirect") != "" {
		t.Errorf("sin HTMX: %d, Location=%q; quiero 303 y /proyectos/{id}", w.Code, w.Header().Get("Location"))
	}
}

// CA-T011-2: la misma pantalla se responde como pagina completa o como fragmento.
func TestPantallas_PaginaOFragmento(t *testing.T) {
	e := nuevoEntorno(t)
	e.crearProyecto(t, "Demo")

	completa := e.get("/proyectos")
	contiene(t, completa, "<!doctype html>", "Demo")

	fragmento := e.enviar(http.MethodGet, "/proyectos", nil, true)
	contiene(t, fragmento, "Demo")
	noContiene(t, fragmento, "<html")
}

// RN-T011-12: la lista marca Proyectos como activo; las paginas de un proyecto, su menu.
func TestPantallas_Navegacion(t *testing.T) {
	e := nuevoEntorno(t)
	id := e.crearProyecto(t, "Demo")

	lista := e.get("/proyectos")
	if n := strings.Count(lista.Body.String(), `aria-current="page"`); n != 1 {
		t.Errorf("la lista marca %d items, quiero 1", n)
	}

	resumen := e.get("/proyectos/" + id)
	contiene(t, resumen, `aria-label="Proyecto"`, "Proyecto: <strong", "/proyectos/"+id+"/backlog")
	backlog := e.get("/proyectos/" + id + "/backlog")
	contiene(t, backlog, `aria-label="Proyecto"`)
	if !regexp.MustCompile(`href="/proyectos/` + id + `/backlog"[^>]*aria-current="page"|aria-current="page"[^>]*href="/proyectos/` + id + `/backlog"`).MatchString(backlog.Body.String()) {
		t.Errorf("el backlog tiene que marcar Backlog como activo:\n%s", backlog.Body.String())
	}
}

// CA-T011-4, CL-T011-6: un dato invalido responde 422 con el campo correcto, sin tocar la base.
func TestPantallas_ErroresDeValidacion(t *testing.T) {
	e := nuevoEntorno(t)

	sinNombre := datosDeProyecto("")
	w := e.post("/proyectos", sinNombre)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("proyecto sin nombre: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="nombre"`, "el nombre es obligatorio")

	malFecha := datosDeProyecto("Demo")
	malFecha.Set("fecha_inicio", "2026-02-30")
	w = e.post("/proyectos", malFecha)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("fecha inexistente: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="fecha_inicio"`)
	noContiene(t, w, "falta la fecha de inicio")

	incoherente := datosDeProyecto("Demo")
	incoherente.Set("fecha_fin", "2026-09-01")
	w = e.post("/proyectos", incoherente)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("fechas incoherentes: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="fecha_fin"`)

	if lista := e.get("/proyectos"); !strings.Contains(lista.Body.String(), "Todavía no hay proyectos") {
		t.Errorf("un alta rechazada no puede guardar nada:\n%s", lista.Body.String())
	}
}

func TestPantallas_ErroresDelBacklog(t *testing.T) {
	e := nuevoEntorno(t)
	id := e.crearProyecto(t, "Demo")

	w := e.post("/proyectos/"+id+"/backlog", url.Values{"titulo": {""}, "prioridad": {"urgente"}})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("historia invalida: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="titulo"`, `data-campo="prioridad"`)

	e.crearHistoria(t, id, "Historia", "must", "")
	historia := e.idDeHistoria(t, id, "Historia")
	w = e.post("/proyectos/"+id+"/backlog/"+historia+"/estimacion", url.Values{"estimacion": {"4"}})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("estimacion fuera de escala: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="estimacion"`)

	w = e.post("/proyectos/"+id+"/backlog/"+historia+"/estimacion", url.Values{"estimacion": {"cinco"}})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("estimacion mal escrita: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="estimacion"`, "numero entero")
}

// CA-T011-4: un email repetido en el proyecto responde 409.
func TestPantallas_EmailRepetido(t *testing.T) {
	e := nuevoEntorno(t)
	id := e.crearProyecto(t, "Demo")
	datos := url.Values{"nombre": {"Ana"}, "email": {"ana@ejemplo.test"}, "rol": {"product_builder"}}
	e.post("/proyectos/"+id+"/integrantes", datos)

	w := e.post("/proyectos/"+id+"/integrantes", datos)

	if w.Code != http.StatusConflict {
		t.Errorf("email repetido: %d, quiero 409", w.Code)
	}
	contiene(t, w, "ya hay un integrante con ese email")
}

// CL-T011-8: estimar una historia completada responde 409 y no cambia la fila.
func TestPantallas_EstimarHistoriaCompletada(t *testing.T) {
	e := nuevoEntorno(t)
	id := e.crearProyecto(t, "Demo")
	completada := backlog.ItemBacklog{Titulo: "Hecha", Prioridad: backlog.PrioridadMust, Estado: backlog.EstadoCompletado, CompletadoEn: time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)}
	registrada, err := e.base.RegistrarItem(context.Background(), id, completada)
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}

	w := e.post("/proyectos/"+id+"/backlog/"+registrada.ID+"/estimacion", url.Values{"estimacion": {"5"}})

	if w.Code != http.StatusConflict {
		t.Errorf("estimar una historia completada: %d, quiero 409", w.Code)
	}
	if puntos, estimado := e.base.Items(id)[0].StoryPoints.Valor(); estimado {
		t.Errorf("la historia completada quedo estimada en %d", puntos)
	}
}

// RN-T011-10, CL-T011-9: un proyecto que no existe o un ID mal formado responde 404.
func TestPantallas_ProyectoInexistente(t *testing.T) {
	e := nuevoEntorno(t)
	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "no-es-un-uuid", "con%20espacios"} {
		for _, ruta := range []string{"/proyectos/" + id, "/proyectos/" + id + "/editar", "/proyectos/" + id + "/backlog"} {
			w := e.get(ruta)
			if w.Code != http.StatusNotFound {
				t.Errorf("GET %s: %d, quiero 404", ruta, w.Code)
			}
			contiene(t, w, "<!doctype html>", "Página no encontrada")
		}
		datos := url.Values{"nombre": {"X"}, "email": {"x@x.test"}, "rol": {"product_builder"}, "titulo": {"T"}, "prioridad": {"must"}, "fecha_inicio": {"2026-10-01"}, "fecha_fin": {"2026-10-02"}}
		for _, ruta := range []string{"/proyectos/" + id, "/proyectos/" + id + "/integrantes", "/proyectos/" + id + "/backlog"} {
			w := e.post(ruta, datos)
			if w.Code != http.StatusNotFound {
				t.Errorf("POST %s: %d, quiero 404", ruta, w.Code)
			}
			noContiene(t, w, "<html")
		}
	}
}

// RN-T011-11, CL-T011-8: una historia solo se alcanza por la ruta de su proyecto.
func TestPantallas_HistoriaDeOtroProyecto(t *testing.T) {
	e := nuevoEntorno(t)
	a := e.crearProyecto(t, "Proyecto A")
	b := e.crearProyecto(t, "Proyecto B")
	e.crearHistoria(t, b, "Historia de B", "must", "")
	historia := e.idDeHistoria(t, b, "Historia de B")

	w := e.get("/proyectos/" + a + "/backlog/" + historia + "/editar")
	if w.Code != http.StatusNotFound {
		t.Errorf("editar por el proyecto equivocado: %d, quiero 404", w.Code)
	}
	w = e.post("/proyectos/"+a+"/backlog/"+historia+"/estimacion", url.Values{"estimacion": {"8"}})
	if w.Code != http.StatusNotFound {
		t.Errorf("estimar por el proyecto equivocado: %d, quiero 404", w.Code)
	}
	w = e.post("/proyectos/"+a+"/backlog/"+historia, url.Values{"titulo": {"Robada"}, "prioridad": {"must"}})
	if w.Code != http.StatusNotFound {
		t.Errorf("editar por el proyecto equivocado: %d, quiero 404", w.Code)
	}
	item := e.base.Items(b)[0]
	if _, estimado := item.StoryPoints.Valor(); estimado || item.Titulo != "Historia de B" {
		t.Errorf("la historia de B cambio: %+v", item)
	}
}

// RN-T011-8, CL-T011-5: el texto con HTML se guarda tal cual y se muestra escapado.
func TestPantallas_EscapaElTexto(t *testing.T) {
	e := nuevoEntorno(t)
	id := e.crearProyecto(t, `<script>alert("x")</script> & ñandú`)

	w := e.get("/proyectos")
	noContiene(t, w, "<script>alert")
	contiene(t, w, "&lt;script&gt;", "ñandú")

	e.crearHistoria(t, id, `Historia <b>negrita</b>`, "must", "")
	w = e.get("/proyectos/" + id + "/backlog")
	noContiene(t, w, "<b>negrita</b>")
	historia := e.idDeHistoria(t, id, `Historia <b>negrita</b>`)

	// El formulario de edicion conserva el texto tal cual, escapado.
	w = e.get("/proyectos/" + id + "/backlog/" + historia + "/editar")
	contiene(t, w, "&lt;b&gt;negrita&lt;/b&gt;")
	noContiene(t, w, "<b>negrita</b>")
}

// RN-T011-7, CL-T011-4: los criterios se guardan uno por linea, sin las lineas vacias, y el formulario
// de edicion los muestra de la misma forma.
func TestPantallas_Criterios(t *testing.T) {
	e := nuevoEntorno(t)
	id := e.crearProyecto(t, "Demo")
	e.crearHistoria(t, id, "Con criterios", "must", "\n  Primero  \n\n   \nSegundo\n")
	e.crearHistoria(t, id, "Sin criterios", "must", " \n\n")

	con := e.base.Items(id)[0]
	if len(con.Criterios) != 2 || con.Criterios[0] != "Primero" || con.Criterios[1] != "Segundo" {
		t.Errorf("criterios guardados = %q, quiero Primero y Segundo", con.Criterios)
	}
	if sin := e.base.Items(id)[1]; len(sin.Criterios) != 0 {
		t.Errorf("la historia sin criterios guardo %q", sin.Criterios)
	}

	w := e.get("/proyectos/" + id + "/backlog/" + con.ID + "/editar")
	contiene(t, w, "Primero\nSegundo", `name="criterios"`)
}

// Editar un proyecto: el formulario muestra los valores actuales y guardar los cambia.
func TestPantallas_EditarProyecto(t *testing.T) {
	e := nuevoEntorno(t)
	id := e.crearProyecto(t, "Demo")

	w := e.get("/proyectos/" + id + "/editar")
	contiene(t, w, `value="Demo"`, `value="2026-10-01"`, `value="2026-12-15"`, "Descripcion de Demo")

	datos := datosDeProyecto("Demo renombrado")
	w = e.post("/proyectos/"+id, datos)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != "/proyectos/"+id {
		t.Fatalf("modificar: %d %q", w.Code, w.Header().Get("Location"))
	}
	if p, _ := e.base.Proyecto(id); p.Nombre != "Demo renombrado" {
		t.Errorf("nombre guardado = %q", p.Nombre)
	}

	datos.Set("nombre", "")
	if w = e.post("/proyectos/"+id, datos); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("modificar con nombre vacio: %d, quiero 422", w.Code)
	}
	if p, _ := e.base.Proyecto(id); p.Nombre != "Demo renombrado" {
		t.Errorf("un cambio rechazado modifico el proyecto: %q", p.Nombre)
	}
}

// El formulario de nuevo proyecto y la lista tienen sus controles etiquetados (RN-T011-9).
func TestPantallas_FormulariosAccesibles(t *testing.T) {
	e := nuevoEntorno(t)
	w := e.get("/proyectos/nuevo")
	if w.Code != http.StatusOK {
		t.Fatalf("GET /proyectos/nuevo: %d", w.Code)
	}
	for _, campo := range []string{"nombre", "descripcion", "fecha_inicio", "fecha_fin"} {
		if !regexp.MustCompile(`<label[^>]*for="` + campo + `"`).MatchString(w.Body.String()) {
			t.Errorf("el campo %s no tiene <label>", campo)
		}
		contiene(t, w, `id="`+campo+`"`)
	}
	contiene(t, w, `hx-post="/proyectos"`)

	id := e.crearProyecto(t, "Demo")
	e.crearHistoria(t, id, "Historia uno", "must", "")
	w = e.get("/proyectos/" + id + "/backlog")
	contiene(t, w, `aria-label="Estimación de Historia uno"`, `aria-label="Editar Historia uno"`)
}

// Las rutas de pantallas no existen si no se registran: el router de T-004 sigue funcionando solo.
func TestRouter_SinPantallasSigueIgual(t *testing.T) {
	router := httpadapter.NuevoRouter(verificadorFalso{}, time.Second, http.FS(fstest.MapFS{}))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/proyectos", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("GET /proyectos sin pantallas: %d, quiero 404", w.Code)
	}
}
