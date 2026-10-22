package httpadapter_test

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

func datosDeSprint(goal string) url.Values {
	return url.Values{"sprint_goal": {goal}, "fecha_inicio": {"2026-10-05"}, "fecha_fin": {"2026-10-18"}}
}

var idDeSprintEnLaRuta = regexp.MustCompile(`^/proyectos/([^/]+)/sprints/([^/]+)$`)

// crearSprint da de alta un sprint por la pantalla y devuelve su ID.
func (e entorno) crearSprint(t *testing.T, proyectoID, goal string) string {
	t.Helper()
	w := e.post("/proyectos/"+proyectoID+"/sprints", datosDeSprint(goal))
	if w.Code != http.StatusSeeOther {
		t.Fatalf("crear el sprint %q: estado %d, cuerpo %s", goal, w.Code, w.Body.String())
	}
	coincide := idDeSprintEnLaRuta.FindStringSubmatch(w.Header().Get("Location"))
	if coincide == nil || coincide[1] != proyectoID {
		t.Fatalf("Location = %q, quiero /proyectos/%s/sprints/{id}", w.Header().Get("Location"), proyectoID)
	}
	return coincide[2]
}

// historiaLista crea una historia estimada y con un criterio, o sea lista para un sprint.
func (e entorno) historiaLista(t *testing.T, proyectoID, titulo string, puntos string) string {
	t.Helper()
	e.crearHistoria(t, proyectoID, titulo, "must", "Un criterio")
	id := e.idDeHistoria(t, proyectoID, titulo)
	w := e.post("/proyectos/"+proyectoID+"/backlog/"+id+"/estimacion", url.Values{"estimacion": {puntos}})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("estimar %q: %d %s", titulo, w.Code, w.Body.String())
	}
	return id
}

func rutaDeSprint(proyectoID, sprintID string) string {
	return "/proyectos/" + proyectoID + "/sprints/" + sprintID
}

// CA-T015-1: el recorrido completo funciona de punta a punta con el router real: crear un sprint,
// asignarle dos historias, marcar una como completada y ver los Story Points del detalle.
func TestPantallasSprints_Recorrido(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	clave := e.historiaLista(t, proyecto, "Historia clave", "5")
	e.historiaLista(t, proyecto, "Historia opcional", "3")

	contiene(t, e.get("/proyectos/"+proyecto+"/sprints"), "Todavía no hay sprints", `name="sprint_goal"`)

	sprintID := e.crearSprint(t, proyecto, "Primer incremento")
	lista := e.get("/proyectos/" + proyecto + "/sprints")
	contiene(t, lista, "Sprint 1", "Primer incremento", "05/10/2026", "18/10/2026", "abierto", rutaDeSprint(proyecto, sprintID))

	// El detalle ofrece las dos historias disponibles y todavia ninguna asignada.
	detalle := e.get(rutaDeSprint(proyecto, sprintID))
	contiene(t, detalle, "Primer incremento", "Historia clave", "Historia opcional", `name="item"`, "El sprint todavía no tiene historias")

	for _, titulo := range []string{"Historia clave", "Historia opcional"} {
		w := e.post(rutaDeSprint(proyecto, sprintID)+"/historias", url.Values{"item": {e.idDeHistoria(t, proyecto, titulo)}})
		if w.Code != http.StatusSeeOther || w.Header().Get("Location") != rutaDeSprint(proyecto, sprintID) {
			t.Fatalf("asignar %q: %d %q %s", titulo, w.Code, w.Header().Get("Location"), w.Body.String())
		}
	}
	detalle = e.get(rutaDeSprint(proyecto, sprintID))
	contiene(t, detalle, "Planificados: 8", "Completados: 0", "No hay historias para asignar", "Marcar completada")

	// Marcar completada una historia actualiza los Story Points y el boton pasa a desmarcar.
	w := e.post(rutaDeSprint(proyecto, sprintID)+"/historias/"+clave+"/completar", nil)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != rutaDeSprint(proyecto, sprintID) {
		t.Fatalf("completar: %d %q %s", w.Code, w.Header().Get("Location"), w.Body.String())
	}
	detalle = e.get(rutaDeSprint(proyecto, sprintID))
	contiene(t, detalle, "Planificados: 8", "Completados: 5", "Desmarcar")

	lista = e.get("/proyectos/" + proyecto + "/sprints")
	contiene(t, lista, ">2<", ">8<", ">5<")

	// Desmarcar la devuelve a pendiente.
	if w := e.post(rutaDeSprint(proyecto, sprintID)+"/historias/"+clave+"/reabrir", nil); w.Code != http.StatusSeeOther {
		t.Fatalf("reabrir: %d %s", w.Code, w.Body.String())
	}
	contiene(t, e.get(rutaDeSprint(proyecto, sprintID)), "Completados: 0")
}

// CA-T015-2: con HTMX, cada alta o cambio responde HX-Redirect con el cuerpo vacio; sin HTMX, 303.
func TestPantallasSprints_RedireccionesConYSinHTMX(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	historia := e.historiaLista(t, proyecto, "Historia", "5")

	w := e.enviar(http.MethodPost, "/proyectos/"+proyecto+"/sprints", datosDeSprint("Con HTMX"), true)
	redireccion := w.Header().Get("HX-Redirect")
	if w.Code != http.StatusOK || !idDeSprintEnLaRuta.MatchString(redireccion) || w.Body.Len() != 0 {
		t.Fatalf("con HTMX: %d, HX-Redirect=%q, cuerpo=%q; quiero 200, /proyectos/{id}/sprints/{id} y cuerpo vacio", w.Code, redireccion, w.Body.String())
	}
	sprintID := idDeSprintEnLaRuta.FindStringSubmatch(redireccion)[2]

	w = e.enviar(http.MethodPost, rutaDeSprint(proyecto, sprintID)+"/historias", url.Values{"item": {historia}}, true)
	if w.Code != http.StatusOK || w.Header().Get("HX-Redirect") != rutaDeSprint(proyecto, sprintID) {
		t.Errorf("asignar con HTMX: %d, HX-Redirect=%q", w.Code, w.Header().Get("HX-Redirect"))
	}

	// El detalle y la lista se responden como pagina o como fragmento segun la peticion.
	fragmento := e.enviar(http.MethodGet, rutaDeSprint(proyecto, sprintID), nil, true)
	contiene(t, fragmento, "Con HTMX")
	noContiene(t, fragmento, "<html")
	contiene(t, e.get(rutaDeSprint(proyecto, sprintID)), "<!doctype html>")
}

// RN-T015-11: las paginas de sprints muestran el menu del proyecto con Sprints activo.
func TestPantallasSprints_Navegacion(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	sprintID := e.crearSprint(t, proyecto, "Meta")

	for _, ruta := range []string{"/proyectos/" + proyecto + "/sprints", rutaDeSprint(proyecto, sprintID)} {
		w := e.get(ruta)
		contiene(t, w, `aria-label="Proyecto"`)
		if !regexp.MustCompile(`href="/proyectos/` + proyecto + `/sprints"[^>]*aria-current="page"`).MatchString(w.Body.String()) {
			t.Errorf("%s: tiene que marcar Sprints como activo:\n%s", ruta, w.Body.String())
		}
	}
}

// CL-T015-7, CA-T015-4: los errores de dominio al crear un sprint caen en su campo, sin guardar nada.
func TestPantallasSprints_ErroresAlCrear(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	ruta := "/proyectos/" + proyecto + "/sprints"

	w := e.post(ruta, datosDeSprint(""))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("sin Sprint Goal: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="sprint_goal"`)

	incoherente := datosDeSprint("Meta")
	incoherente.Set("fecha_fin", "2026-10-01")
	if w = e.post(ruta, incoherente); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("fechas incoherentes: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="fecha_fin"`)

	malFecha := datosDeSprint("Meta")
	malFecha.Set("fecha_inicio", "2026-02-30")
	if w = e.post(ruta, malFecha); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("fecha inexistente: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="fecha_inicio"`)

	if lista := e.get(ruta); !strings.Contains(lista.Body.String(), "Todavía no hay sprints") {
		t.Errorf("un alta rechazada no puede guardar nada:\n%s", lista.Body.String())
	}

	// Fuera de las fechas del proyecto, y con el ultimo sprint todavia abierto: 409.
	fuera := datosDeSprint("Meta")
	fuera.Set("fecha_fin", "2027-03-01")
	if w = e.post(ruta, fuera); w.Code != http.StatusConflict {
		t.Errorf("fuera del proyecto: %d, quiero 409", w.Code)
	}
	e.crearSprint(t, proyecto, "Primero")
	segundo := datosDeSprint("Segundo")
	segundo.Set("fecha_inicio", "2026-10-20")
	segundo.Set("fecha_fin", "2026-11-02")
	if w = e.post(ruta, segundo); w.Code != http.StatusConflict {
		t.Errorf("con el ultimo sprint abierto: %d, quiero 409", w.Code)
	}
	contiene(t, w, "el ultimo sigue abierto")
}

// CL-T015-4, CL-T015-5: una historia sin estimar o sin criterios se ofrece marcada y asignarla se
// rechaza con el mensaje del dominio; una ya asignada no se ofrece mas y no se asigna otra vez.
func TestPantallasSprints_ReglasDeAsignacion(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	sprintID := e.crearSprint(t, proyecto, "Meta")
	e.crearHistoria(t, proyecto, "Sin estimar", "must", "Un criterio")
	sinEstimar := e.idDeHistoria(t, proyecto, "Sin estimar")
	e.crearHistoria(t, proyecto, "Sin criterios", "should", "")
	sinCriterios := e.idDeHistoria(t, proyecto, "Sin criterios")
	e.post("/proyectos/"+proyecto+"/backlog/"+sinCriterios+"/estimacion", url.Values{"estimacion": {"3"}})
	lista := e.historiaLista(t, proyecto, "Lista", "5")

	detalle := e.get(rutaDeSprint(proyecto, sprintID))
	contiene(t, detalle, "Sin estimar (sin estimar)", "Sin criterios (sin criterios)")

	ruta := rutaDeSprint(proyecto, sprintID) + "/historias"
	w := e.post(ruta, url.Values{"item": {sinEstimar}})
	if w.Code != http.StatusConflict {
		t.Errorf("asignar una historia sin estimar: %d, quiero 409", w.Code)
	}
	contiene(t, w, "no esta estimada")
	if w = e.post(ruta, url.Values{"item": {sinCriterios}}); w.Code != http.StatusConflict {
		t.Errorf("asignar una historia sin criterios: %d, quiero 409", w.Code)
	}
	contiene(t, w, "no tiene criterios")

	if w = e.post(ruta, url.Values{"item": {lista}}); w.Code != http.StatusSeeOther {
		t.Fatalf("asignar una historia lista: %d %s", w.Code, w.Body.String())
	}
	if w = e.post(ruta, url.Values{"item": {lista}}); w.Code != http.StatusConflict {
		t.Errorf("asignar dos veces: %d, quiero 409", w.Code)
	}
	contiene(t, w, "ya esta asignada")
	noContiene(t, e.get(rutaDeSprint(proyecto, sprintID)), "Lista (")

	if w = e.post(ruta, url.Values{"item": {"no-es-un-uuid"}}); w.Code != http.StatusNotFound {
		t.Errorf("asignar una historia que no existe: %d, quiero 404", w.Code)
	}
}

// CL-T015-6: marcar dos veces una historia, o desmarcar una pendiente, no es un error.
func TestPantallasSprints_MarcarYDesmarcarSonIdempotentes(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	sprintID := e.crearSprint(t, proyecto, "Meta")
	historia := e.historiaLista(t, proyecto, "Historia", "5")
	e.post(rutaDeSprint(proyecto, sprintID)+"/historias", url.Values{"item": {historia}})
	base := rutaDeSprint(proyecto, sprintID) + "/historias/" + historia

	for _, ruta := range []string{base + "/reabrir", base + "/completar", base + "/completar", base + "/reabrir", base + "/reabrir"} {
		if w := e.post(ruta, nil); w.Code != http.StatusSeeOther {
			t.Errorf("POST %s: %d %s, quiero 303", ruta, w.Code, w.Body.String())
		}
	}
}

// RN-T015-7, CA-T015-3: con el sprint cerrado no hay botones ni formulario de asignacion, pero las
// rutas existen y el dominio responde 409.
func TestPantallasSprints_SprintCerrado(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	historia := e.historiaLista(t, proyecto, "Historia", "5")
	inicio := time.Date(2026, time.October, 5, 0, 0, 0, 0, time.UTC)
	cerrado, err := e.base.RegistrarSprint(context.Background(), sprint.Sprint{
		ProyectoID: proyecto, SprintGoal: "Terminado", FechaInicio: inicio, FechaFin: inicio.AddDate(0, 0, 13), Estado: sprint.EstadoCerrado,
	})
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}

	w := e.get(rutaDeSprint(proyecto, cerrado.ID))
	contiene(t, w, "Terminado", "cerrado")
	noContiene(t, w, `name="item"`, "Marcar completada")

	if w = e.post(rutaDeSprint(proyecto, cerrado.ID)+"/historias", url.Values{"item": {historia}}); w.Code != http.StatusConflict {
		t.Errorf("asignar a un sprint cerrado: %d, quiero 409", w.Code)
	}
	contiene(t, w, "el sprint esta cerrado")
}

// RN-T015-5, CL-T015-8, RN-T015-10: un sprint solo se alcanza por la ruta de su proyecto y una historia
// solo por la de su sprint; lo que no existe o esta mal formado responde 404.
func TestPantallasSprints_RutasQueNoCorresponden(t *testing.T) {
	e := nuevoEntorno(t)
	a := e.crearProyecto(t, "Proyecto A")
	b := e.crearProyecto(t, "Proyecto B")
	sprintDeA := e.crearSprint(t, a, "Meta de A")
	historiaDeA := e.historiaLista(t, a, "Historia de A", "5")
	e.post(rutaDeSprint(a, sprintDeA)+"/historias", url.Values{"item": {historiaDeA}})
	otraDeA := e.historiaLista(t, a, "Otra de A", "3")

	// El sprint de A por la ruta de B.
	if w := e.get(rutaDeSprint(b, sprintDeA)); w.Code != http.StatusNotFound {
		t.Errorf("GET sprint por el proyecto equivocado: %d, quiero 404", w.Code)
	}
	for _, ruta := range []string{rutaDeSprint(b, sprintDeA) + "/historias", rutaDeSprint(b, sprintDeA) + "/historias/" + historiaDeA + "/completar"} {
		if w := e.post(ruta, url.Values{"item": {otraDeA}}); w.Code != http.StatusNotFound {
			t.Errorf("POST %s: %d, quiero 404", ruta, w.Code)
		}
	}
	// Una historia que no esta en ese sprint no se marca ni se desmarca.
	for _, accion := range []string{"completar", "reabrir"} {
		if w := e.post(rutaDeSprint(a, sprintDeA)+"/historias/"+otraDeA+"/"+accion, nil); w.Code != http.StatusNotFound {
			t.Errorf("%s una historia que no esta en el sprint: %d, quiero 404", accion, w.Code)
		}
	}
	// IDs que no existen o estan mal formados.
	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "no-es-un-uuid"} {
		if w := e.get(rutaDeSprint(a, id)); w.Code != http.StatusNotFound {
			t.Errorf("GET sprint %s: %d, quiero 404", id, w.Code)
		}
		if w := e.get("/proyectos/" + id + "/sprints"); w.Code != http.StatusNotFound {
			t.Errorf("GET sprints del proyecto %s: %d, quiero 404", id, w.Code)
		}
		if w := e.post("/proyectos/"+id+"/sprints", datosDeSprint("Meta")); w.Code != http.StatusNotFound {
			t.Errorf("POST sprint en el proyecto %s: %d, quiero 404", id, w.Code)
		}
	}
	// Nada cambio.
	contiene(t, e.get(rutaDeSprint(a, sprintDeA)), "Completados: 0")
}

// RN-T015-9, CL-T015-10: el Sprint Goal con HTML se guarda tal cual y se muestra escapado.
func TestPantallasSprints_EscapaElTexto(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	sprintID := e.crearSprint(t, proyecto, `<script>alert("x")</script> & ñandú`)

	for _, ruta := range []string{"/proyectos/" + proyecto + "/sprints", rutaDeSprint(proyecto, sprintID)} {
		w := e.get(ruta)
		noContiene(t, w, "<script>alert")
		contiene(t, w, "&lt;script&gt;", "ñandú")
	}
}

// RN-T015-12: los campos tienen su <label> y los botones de cada fila dicen a que historia corresponden.
func TestPantallasSprints_FormulariosAccesibles(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	w := e.get("/proyectos/" + proyecto + "/sprints")
	for _, campo := range []string{"sprint_goal", "fecha_inicio", "fecha_fin"} {
		if !regexp.MustCompile(`<label[^>]*for="` + campo + `"`).MatchString(w.Body.String()) {
			t.Errorf("el campo %s no tiene <label>", campo)
		}
	}
	contiene(t, w, `hx-post="/proyectos/`+proyecto+`/sprints"`)

	sprintID := e.crearSprint(t, proyecto, "Meta")
	historia := e.historiaLista(t, proyecto, "Historia uno", "5")
	e.historiaLista(t, proyecto, "Historia sin asignar", "3")
	e.post(rutaDeSprint(proyecto, sprintID)+"/historias", url.Values{"item": {historia}})
	w = e.get(rutaDeSprint(proyecto, sprintID))
	contiene(t, w, `aria-label="Marcar completada: Historia uno"`, `<label for="item"`)
}
