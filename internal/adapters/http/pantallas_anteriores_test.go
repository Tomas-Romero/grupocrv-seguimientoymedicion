package httpadapter_test

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// cerrarEnLaBase cierra el sprint con la base en memoria, igual que lo hace el cierre de US-011. Asi
// estos tests no dependen del boton de cerrar.
func cerrarEnLaBase(t *testing.T, e entorno, sprintID string) {
	t.Helper()
	if _, err := e.base.GuardarCierre(context.Background(), sprintID); err != nil {
		t.Fatalf("cerrar el sprint %s: %v", sprintID, err)
	}
}

// proyectoConHistorial arma un proyecto con dos sprints cerrados: el 1 planifico una historia de 5
// puntos (completada) y una de 3 (no completada); el 2, una de 8 (completada). Devuelve los IDs.
func proyectoConHistorial(t *testing.T, e entorno) (proyecto, primero, segundo string) {
	t.Helper()
	proyecto = e.crearProyecto(t, "Demo")
	alta := e.historiaLista(t, proyecto, "Alta de proyectos", "5")
	listar := e.historiaLista(t, proyecto, "Listar proyectos", "3")
	editar := e.historiaLista(t, proyecto, "Editar proyectos", "8")

	primero = e.crearSprint(t, proyecto, "MVP navegable")
	for _, id := range []string{alta, listar} {
		e.post(rutaDeSprint(proyecto, primero)+"/historias", url.Values{"item": {id}})
	}
	e.post(rutaDeSprint(proyecto, primero)+"/historias/"+alta+"/completar", nil)
	cerrarEnLaBase(t, e, primero)

	creado := e.post("/proyectos/"+proyecto+"/sprints", datosDeSprintSiguiente("Listados"))
	if creado.Code != http.StatusSeeOther {
		t.Fatalf("crear el segundo sprint: %d %s", creado.Code, creado.Body.String())
	}
	segundo = idDeSprintEnLaRuta.FindStringSubmatch(creado.Header().Get("Location"))[2]
	e.post(rutaDeSprint(proyecto, segundo)+"/historias", url.Values{"item": {editar}})
	e.post(rutaDeSprint(proyecto, segundo)+"/historias/"+editar+"/completar", nil)
	cerrarEnLaBase(t, e, segundo)
	return proyecto, primero, segundo
}

func datosDeSprintSiguiente(goal string) url.Values {
	return url.Values{"sprint_goal": {goal}, "fecha_inicio": {"2026-10-19"}, "fecha_fin": {"2026-11-01"}}
}

func rutaDeAnteriores(proyectoID string) string {
	return "/proyectos/" + proyectoID + "/sprints/anteriores"
}

// US-012 (CA-012-1 y CA-012-2): con el router real y la base en memoria, el historial muestra los sprints
// cerrados del mas reciente al mas antiguo, cada uno con su resultado y con enlace a su detalle, y la
// velocidad del equipo.
func TestPantallasAnteriores_Historial(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto, primero, segundo := proyectoConHistorial(t, e)

	w := e.get(rutaDeAnteriores(proyecto))

	if w.Code != http.StatusOK {
		t.Fatalf("GET: %d %s", w.Code, w.Body.String())
	}
	contiene(t, w, "Sprints anteriores", "MVP navegable", "Listados", "05/10/2026", "18/10/2026", "19/10/2026", "01/11/2026",
		rutaDeSprint(proyecto, primero), rutaDeSprint(proyecto, segundo),
		"SP planificados", "SP completados", "Historias completadas", "Historias no completadas",
		"Velocidad del equipo: 6,5 Story Points por sprint")
	cuerpo := w.Body.String()
	if strings.Index(cuerpo, "Sprint 2") > strings.Index(cuerpo, "Sprint 1") {
		t.Error("el Sprint 2 tiene que ir antes que el Sprint 1 (el mas reciente primero)")
	}
	// Cada fila trae sus cifras en el orden de la tabla: planificados, completados, historias
	// completadas y no completadas.
	filas := map[string]*regexp.Regexp{
		"Sprint 2": regexp.MustCompile(`Sprint 2.*?>8<.*?>8<.*?>1<.*?>0<`),
		"Sprint 1": regexp.MustCompile(`Sprint 1.*?>8<.*?>5<.*?>1<.*?>1<`),
	}
	for nombre, patron := range filas {
		if !patron.MatchString(cuerpo) {
			t.Errorf("la fila del %s no tiene sus cifras en orden:\n%s", nombre, cuerpo)
		}
	}
}

// US-012 (CA-012-3 y CA-012-4), CL-012-1, CL-012-2: el sprint abierto no aparece; sin sprints cerrados no hay
// error, se dice y no se muestra velocidad.
func TestPantallasAnteriores_SinSprintsCerrados(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	contiene(t, e.get(rutaDeAnteriores(proyecto)), "Todavía no hay sprints cerrados")

	e.crearSprint(t, proyecto, "Todavia abierto")
	w := e.get(rutaDeAnteriores(proyecto))
	contiene(t, w, "Todavía no hay sprints cerrados")
	noContiene(t, w, "Todavia abierto", "Velocidad del equipo")
}

// US-012 / CL-012-3: un sprint cerrado sin historias aparece con todo en cero.
func TestPantallasAnteriores_SprintSinHistorias(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	sprintID := e.crearSprint(t, proyecto, "Vacio")
	cerrarEnLaBase(t, e, sprintID)

	w := e.get(rutaDeAnteriores(proyecto))

	contiene(t, w, "Vacio", "Velocidad del equipo: 0 Story Points por sprint")
	if !regexp.MustCompile(`Sprint 1.*?>0<.*?>0<.*?>0<.*?>0<`).MatchString(w.Body.String()) {
		t.Errorf("la fila del Sprint 1 no tiene todo en cero:\n%s", w.Body.String())
	}
}

// US-012 / CA-012-5, RN-012-4: una historia que no se completo en el sprint 1 y se completo despues en
// el 2 sigue contando como no completada en el 1.
func TestPantallasAnteriores_ElResultadoEsElDeLaFila(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	listar := e.historiaLista(t, proyecto, "Listar proyectos", "3")
	primero := e.crearSprint(t, proyecto, "MVP navegable")
	e.post(rutaDeSprint(proyecto, primero)+"/historias", url.Values{"item": {listar}})
	cerrarEnLaBase(t, e, primero)
	creado := e.post("/proyectos/"+proyecto+"/sprints", datosDeSprintSiguiente("Listados"))
	segundo := idDeSprintEnLaRuta.FindStringSubmatch(creado.Header().Get("Location"))[2]
	e.post(rutaDeSprint(proyecto, segundo)+"/historias", url.Values{"item": {listar}})
	e.post(rutaDeSprint(proyecto, segundo)+"/historias/"+listar+"/completar", nil)
	cerrarEnLaBase(t, e, segundo)

	cuerpo := e.get(rutaDeAnteriores(proyecto)).Body.String()

	if !regexp.MustCompile(`Sprint 1.*?>3<.*?>0<.*?>0<.*?>1<`).MatchString(cuerpo) {
		t.Errorf("el Sprint 1 tiene que seguir con 0 completados y 1 historia no completada:\n%s", cuerpo)
	}
	if !regexp.MustCompile(`Sprint 2.*?>3<.*?>3<.*?>1<.*?>0<`).MatchString(cuerpo) {
		t.Errorf("el Sprint 2 tiene que tener la historia completada:\n%s", cuerpo)
	}
}

// RN-T015-10 aplicado a US-012 (CL-012-11, CL-012-12): un proyecto inexistente o con el ID mal
// formado responde 404, y los sprints de otro proyecto no aparecen.
func TestPantallasAnteriores_RutasQueNoCorresponden(t *testing.T) {
	e := nuevoEntorno(t)
	a, _, _ := proyectoConHistorial(t, e)
	b := e.crearProyecto(t, "Proyecto B")

	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "no-es-un-uuid"} {
		if w := e.get(rutaDeAnteriores(id)); w.Code != http.StatusNotFound {
			t.Errorf("GET anteriores del proyecto %s: %d, quiero 404", id, w.Code)
		}
	}
	w := e.get(rutaDeAnteriores(b))
	contiene(t, w, "Todavía no hay sprints cerrados")
	noContiene(t, w, "MVP navegable", "Listados")

	// El historial no le quita su ruta al detalle de un sprint.
	if w := e.get(rutaDeSprint(a, "no-es-un-uuid")); w.Code != http.StatusNotFound {
		t.Errorf("GET de un sprint con ID mal formado: %d, quiero 404", w.Code)
	}
}

// RN-T015-11 y T-011: pagina o fragmento segun la peticion, con el menu del proyecto en Sprints, y el
// Sprint Goal con HTML se muestra escapado (CL-012-10).
func TestPantallasAnteriores_PaginaFragmentoYEscape(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	sprintID := e.crearSprint(t, proyecto, `<script>alert("x")</script> & ñandú`)
	cerrarEnLaBase(t, e, sprintID)

	pagina := e.get(rutaDeAnteriores(proyecto))
	contiene(t, pagina, "<!doctype html>", `aria-current="page">Sprints</a>`, "&lt;script&gt;", "ñandú", "Ver todos los sprints")
	noContiene(t, pagina, "<script>alert")

	fragmento := e.enviar(http.MethodGet, rutaDeAnteriores(proyecto), nil, true)
	contiene(t, fragmento, "Sprints anteriores")
	noContiene(t, fragmento, "<html")
}

// US-012: la lista de sprints (T-015) enlaza al historial.
func TestPantallasAnteriores_LaListaEnlazaAlHistorial(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")

	contiene(t, e.get("/proyectos/"+proyecto+"/sprints"), `href="`+rutaDeAnteriores(proyecto)+`"`, "Ver sprints anteriores")
}
