package httpadapter_test

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
)

// Pantallas de esfuerzo (spec T-019). El reloj de los tests es el 2026-10-20: hoy se puede
// registrar, mañana no.

func rutaDeEsfuerzo(proyectoID string) string { return "/proyectos/" + proyectoID + "/esfuerzo" }

// registrarIntegrante da de alta un integrante por la pantalla y devuelve su ID.
func (e entorno) registrarIntegrante(t *testing.T, proyectoID, nombre string) string {
	t.Helper()
	email := strings.ToLower(nombre) + "@ejemplo.test"
	w := e.post("/proyectos/"+proyectoID+"/integrantes", url.Values{"nombre": {nombre}, "email": {email}, "rol": {"product_builder"}})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("registrar al integrante %q: %d %s", nombre, w.Code, w.Body.String())
	}
	for _, in := range e.base.Integrantes(proyectoID) {
		if in.Nombre == nombre {
			return in.ID
		}
	}
	t.Fatalf("no hay un integrante %q en el proyecto %s", nombre, proyectoID)
	return ""
}

// crearSprintConFechas da de alta un sprint por la pantalla, con sus fechas, y devuelve su ID.
func (e entorno) crearSprintConFechas(t *testing.T, proyectoID, goal, inicio, fin string) string {
	t.Helper()
	w := e.post("/proyectos/"+proyectoID+"/sprints", url.Values{"sprint_goal": {goal}, "fecha_inicio": {inicio}, "fecha_fin": {fin}})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("crear el sprint %q: %d %s", goal, w.Code, w.Body.String())
	}
	return idDeSprintEnLaRuta.FindStringSubmatch(w.Header().Get("Location"))[2]
}

// cerrarSprint cierra un sprint por la pantalla: no se crea otro mientras el ultimo sigue abierto.
func (e entorno) cerrarSprint(t *testing.T, proyectoID, sprintID string) {
	t.Helper()
	if w := e.post(rutaDeSprint(proyectoID, sprintID)+"/cerrar", nil); w.Code != http.StatusSeeOther {
		t.Fatalf("cerrar el sprint %s: %d %s", sprintID, w.Code, w.Body.String())
	}
}

// cargarHoras guarda un registro de esfuerzo directo en la base en memoria.
func (e entorno) cargarHoras(t *testing.T, proyectoID, historiaID, integranteID, fecha, actividad string, horas float64) {
	t.Helper()
	dia, err := time.Parse(time.DateOnly, fecha)
	if err != nil {
		t.Fatalf("fecha %q: %v", fecha, err)
	}
	r := esfuerzo.RegistroEsfuerzo{ProyectoID: proyectoID, ItemID: historiaID, IntegranteID: integranteID, Fecha: dia, Actividad: actividad, Horas: horas}
	if _, err := e.base.RegistrarEsfuerzo(context.Background(), r); err != nil {
		t.Fatalf("RegistrarEsfuerzo(%q): %v", actividad, err)
	}
}

// historiaConEstado guarda una historia ya estimada y con el estado dado, directo en la base.
func (e entorno) historiaConEstado(t *testing.T, proyectoID, titulo string, puntos int, estado backlog.Estado) string {
	t.Helper()
	item := backlog.ItemBacklog{Titulo: titulo, Prioridad: backlog.PrioridadMust, Estado: estado}
	if puntos > 0 {
		sp, err := backlog.NuevosStoryPoints(puntos)
		if err != nil {
			t.Fatalf("NuevosStoryPoints(%d): %v", puntos, err)
		}
		item.StoryPoints = sp
	}
	if estado == backlog.EstadoCompletado {
		item.CompletadoEn = time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	}
	registrada, err := e.base.RegistrarItem(context.Background(), proyectoID, item)
	if err != nil {
		t.Fatalf("RegistrarItem(%q): %v", titulo, err)
	}
	return registrada.ID
}

func datosDeHoras(integrante, historia, fecha, actividad, horas string) url.Values {
	return url.Values{"integrante": {integrante}, "historia": {historia}, "fecha": {fecha}, "actividad": {actividad}, "horas": {horas}}
}

func (e entorno) cantidadDeRegistros(proyectoID string) int { return len(e.base.Esfuerzos(proyectoID)) }

// CA-T019-1: el recorrido completo con el router real y la base en memoria: se registran 2,5 horas y la
// pantalla las muestra con su historia, su integrante, su actividad y el total.
func TestPantallasEsfuerzo_Recorrido(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")
	e.crearHistoria(t, proyecto, "Historia clave", "must", "")
	historia := e.idDeHistoria(t, proyecto, "Historia clave")

	vacia := e.get(rutaDeEsfuerzo(proyecto))
	if vacia.Code != http.StatusOK {
		t.Fatalf("GET esfuerzo: %d %s", vacia.Code, vacia.Body.String())
	}
	contiene(t, vacia, "Registrar horas", `name="integrante"`, `name="historia"`, `name="fecha"`, `name="actividad"`, `name="horas"`,
		`value="2026-10-20"`, "Todavía no hay horas registradas", `hx-post="`+rutaDeEsfuerzo(proyecto)+`"`,
		"Ana", "Historia clave", "Comparar con lo estimado", rutaDeEsfuerzo(proyecto)+"/comparacion")

	w := e.post(rutaDeEsfuerzo(proyecto), datosDeHoras(ana, historia, "2026-10-19", "Armar el formulario", "2,5"))
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != rutaDeEsfuerzo(proyecto) {
		t.Fatalf("registrar: %d %q %s", w.Code, w.Header().Get("Location"), w.Body.String())
	}

	pagina := e.get(rutaDeEsfuerzo(proyecto))
	contiene(t, pagina, "19/10/2026", "Historia clave", "Ana", "Armar el formulario", "2,5", "Total: 2,5 h", "Fuera de sprint")
	noContiene(t, pagina, "Todavía no hay horas registradas")
}

// CA-T019-6: con HTMX el alta responde HX-Redirect con el cuerpo vacío y la página se responde como
// fragmento; sin HTMX, un 303 y la página completa con el menú.
func TestPantallasEsfuerzo_ConYSinHTMX(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")
	e.crearHistoria(t, proyecto, "Historia", "must", "")
	historia := e.idDeHistoria(t, proyecto, "Historia")

	w := e.enviar(http.MethodPost, rutaDeEsfuerzo(proyecto), datosDeHoras(ana, historia, "2026-10-20", "Con HTMX", "1"), true)
	if w.Code != http.StatusOK || w.Header().Get("HX-Redirect") != rutaDeEsfuerzo(proyecto) || w.Body.Len() != 0 {
		t.Fatalf("con HTMX: %d, HX-Redirect=%q, cuerpo=%q", w.Code, w.Header().Get("HX-Redirect"), w.Body.String())
	}

	fragmento := e.enviar(http.MethodGet, rutaDeEsfuerzo(proyecto), nil, true)
	contiene(t, fragmento, "Con HTMX")
	noContiene(t, fragmento, "<html")

	completa := e.get(rutaDeEsfuerzo(proyecto))
	contiene(t, completa, "<!doctype html>", `aria-label="Proyecto"`, `href="`+rutaDeEsfuerzo(proyecto)+`"`, `aria-current="page"`)
	contiene(t, e.get(rutaDeEsfuerzo(proyecto)+"/comparacion"), `aria-label="Proyecto"`, `aria-current="page"`)
}

// CL-T019-2: las horas se aceptan con coma y con punto; 2,50 también.
func TestPantallasEsfuerzo_HorasConComaYPunto(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")
	e.crearHistoria(t, proyecto, "Historia", "must", "")
	historia := e.idDeHistoria(t, proyecto, "Historia")

	for _, horas := range []string{"2,5", "2.5", "2,50"} {
		if w := e.post(rutaDeEsfuerzo(proyecto), datosDeHoras(ana, historia, "2026-10-20", "Trabajo "+horas, horas)); w.Code != http.StatusSeeOther {
			t.Errorf("horas %q: %d %s", horas, w.Code, w.Body.String())
		}
	}
	for _, r := range e.base.Esfuerzos(proyecto) {
		if r.Horas != 2.5 {
			t.Errorf("el registro %q quedó con %v horas, se esperaba 2,5", r.Actividad, r.Horas)
		}
	}
	contiene(t, e.get(rutaDeEsfuerzo(proyecto)), "Total: 7,5 h")

	antes := e.cantidadDeRegistros(proyecto)
	w := e.post(rutaDeEsfuerzo(proyecto), datosDeHoras(ana, historia, "2026-10-20", "Con fracción", "1,3"))
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("1,3 horas: %d, se esperaba 422", w.Code)
	}
	contiene(t, w, `data-campo="horas"`)
	w = e.post(rutaDeEsfuerzo(proyecto), datosDeHoras(ana, historia, "2026-10-20", "No es un número", "abc"))
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("abc horas: %d, se esperaba 422", w.Code)
	}
	contiene(t, w, `data-campo="horas"`, "un numero")
	if e.cantidadDeRegistros(proyecto) != antes {
		t.Error("un rechazo registró horas")
	}
}

// CA-T019-5, CL-T019-3, CL-T019-4: los rechazos del dominio y de la pantalla salen en su campo y no
// registran nada.
func TestPantallasEsfuerzo_Rechazos(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")
	e.crearHistoria(t, proyecto, "Historia", "must", "")
	historia := e.idDeHistoria(t, proyecto, "Historia")

	casos := []struct {
		nombre string
		datos  url.Values
		campos []string
		texto  string
	}{
		{"fecha de mañana", datosDeHoras(ana, historia, "2026-10-21", "Trabajo", "1"), []string{"fecha"}, ""},
		{"fecha vacía", datosDeHoras(ana, historia, "", "Trabajo", "1"), []string{"fecha"}, ""},
		{"fecha mal escrita", datosDeHoras(ana, historia, "20/10/2026", "Trabajo", "1"), []string{"fecha"}, "AAAA-MM-DD"},
		{"horas en cero", datosDeHoras(ana, historia, "2026-10-20", "Trabajo", "0"), []string{"horas"}, ""},
		{"horas de más", datosDeHoras(ana, historia, "2026-10-20", "Trabajo", "25"), []string{"horas"}, ""},
		{"horas vacías", datosDeHoras(ana, historia, "2026-10-20", "Trabajo", ""), []string{"horas"}, ""},
		{"actividad vacía", datosDeHoras(ana, historia, "2026-10-20", "   ", "1"), []string{"actividad"}, ""},
		{"actividad de 121", datosDeHoras(ana, historia, "2026-10-20", strings.Repeat("a", 121), "1"), []string{"actividad"}, ""},
		{"integrante sin elegir", datosDeHoras("", historia, "2026-10-20", "Trabajo", "1"), []string{"integrante"}, "Elegí un integrante"},
		{"historia sin elegir", datosDeHoras(ana, "", "2026-10-20", "Trabajo", "1"), []string{"historia"}, "Elegí una historia"},
		{"las dos sin elegir", datosDeHoras("", "", "2026-10-20", "Trabajo", "1"), []string{"integrante", "historia"}, "Elegí un integrante"},
		{"varios a la vez", datosDeHoras(ana, historia, "2026-10-21", "", "1,3"), []string{"fecha", "actividad", "horas"}, ""},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := e.post(rutaDeEsfuerzo(proyecto), c.datos)
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("estado %d, se esperaba 422: %s", w.Code, w.Body.String())
			}
			for _, campo := range c.campos {
				contiene(t, w, `data-campo="`+campo+`"`)
			}
			if c.texto != "" {
				contiene(t, w, c.texto)
			}
			if n := e.cantidadDeRegistros(proyecto); n != 0 {
				t.Errorf("el rechazo registró %d registros", n)
			}
		})
	}

	t.Run("hoy y 120 caracteres se aceptan", func(t *testing.T) {
		w := e.post(rutaDeEsfuerzo(proyecto), datosDeHoras(ana, historia, "2026-10-20", "  "+strings.Repeat("a", 120)+"  ", "0,25"))
		if w.Code != http.StatusSeeOther {
			t.Fatalf("estado %d: %s", w.Code, w.Body.String())
		}
		if r := e.base.Esfuerzos(proyecto); len(r) != 1 || r[0].Actividad != strings.Repeat("a", 120) {
			t.Errorf("registros = %+v, se esperaba uno con la actividad recortada", r)
		}
	})
}

// RN-T019-5, RN-T019-12: lo que es de otro proyecto o no existe es un 404.
func TestPantallasEsfuerzo_AlcanceDelProyecto(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	otro := e.crearProyecto(t, "Otro")
	ana := e.registrarIntegrante(t, proyecto, "Ana")
	beto := e.registrarIntegrante(t, otro, "Beto")
	e.crearHistoria(t, proyecto, "Propia", "must", "")
	e.crearHistoria(t, otro, "Ajena", "must", "")
	propia, ajena := e.idDeHistoria(t, proyecto, "Propia"), e.idDeHistoria(t, otro, "Ajena")

	casos := []struct {
		nombre string
		ruta   string
		datos  url.Values
	}{
		{"integrante de otro proyecto", rutaDeEsfuerzo(proyecto), datosDeHoras(beto, propia, "2026-10-20", "Trabajo", "1")},
		{"historia de otro proyecto", rutaDeEsfuerzo(proyecto), datosDeHoras(ana, ajena, "2026-10-20", "Trabajo", "1")},
		{"integrante inexistente", rutaDeEsfuerzo(proyecto), datosDeHoras("no-existe", propia, "2026-10-20", "Trabajo", "1")},
		{"historia mal formada", rutaDeEsfuerzo(proyecto), datosDeHoras(ana, "no-es-un-id", "2026-10-20", "Trabajo", "1")},
		{"proyecto inexistente", rutaDeEsfuerzo("no-existe"), datosDeHoras(ana, propia, "2026-10-20", "Trabajo", "1")},
		{"proyecto inexistente con campos sin elegir", rutaDeEsfuerzo("no-existe"), datosDeHoras("", "", "", "", "")},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := e.post(c.ruta, c.datos)
			if w.Code != http.StatusNotFound {
				t.Errorf("estado %d, se esperaba 404: %s", w.Code, w.Body.String())
			}
			noContiene(t, w, "<html")
		})
	}
	if e.cantidadDeRegistros(proyecto)+e.cantidadDeRegistros(otro) != 0 {
		t.Error("un rechazo registró horas")
	}

	for _, ruta := range []string{rutaDeEsfuerzo("no-existe"), rutaDeEsfuerzo("no-existe") + "/comparacion"} {
		w := e.get(ruta)
		if w.Code != http.StatusNotFound {
			t.Errorf("GET %s: %d, se esperaba 404", ruta, w.Code)
		}
		contiene(t, w, "<!doctype html>")
	}
	for _, consulta := range []string{"?historia=" + ajena, "?integrante=" + beto, "?sprint=no-existe", "?historia=no-es-un-id"} {
		w := e.get(rutaDeEsfuerzo(proyecto) + consulta)
		if w.Code != http.StatusNotFound {
			t.Errorf("GET esfuerzo%s: %d, se esperaba 404", consulta, w.Code)
		}
		contiene(t, w, "<!doctype html>")
	}
}

// RN-T019-6, CL-T019-5: el HTML se escapa y dos registros iguales son dos filas.
func TestPantallasEsfuerzo_EscapaYAceptaIguales(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")
	e.crearHistoria(t, proyecto, "Con <b>HTML</b>", "must", "")
	historia := e.idDeHistoria(t, proyecto, "Con <b>HTML</b>")

	for i := 0; i < 2; i++ {
		if w := e.post(rutaDeEsfuerzo(proyecto), datosDeHoras(ana, historia, "2026-10-20", "<script>alert(1)</script>", "1")); w.Code != http.StatusSeeOther {
			t.Fatalf("registrar %d: %d %s", i, w.Code, w.Body.String())
		}
	}
	w := e.get(rutaDeEsfuerzo(proyecto))
	contiene(t, w, "&lt;script&gt;alert(1)&lt;/script&gt;", "Con &lt;b&gt;HTML&lt;/b&gt;", "Total: 2 h")
	noContiene(t, w, "<script>alert(1)", "<b>HTML</b>")
	if n := strings.Count(w.Body.String(), "&lt;script&gt;alert(1)&lt;/script&gt;"); n != 2 {
		t.Errorf("la actividad aparece %d veces, se esperaban 2 filas", n)
	}
}

// CL-T019-1, RN-T019-4: sin historias o sin integrantes se avisa y registrar se rechaza en su campo.
func TestPantallasEsfuerzo_SinHistoriasNiIntegrantes(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")

	w := e.get(rutaDeEsfuerzo(proyecto))
	contiene(t, w, "Primero cargá una historia", "Primero registrá un integrante")

	w = e.post(rutaDeEsfuerzo(proyecto), datosDeHoras("", "", "2026-10-20", "Trabajo", "1"))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("estado %d, se esperaba 422", w.Code)
	}
	contiene(t, w, `data-campo="integrante"`, `data-campo="historia"`)
}

// Dos integrantes, dos historias y dos sprints; un registro cae antes del primer sprint.
type escenarioDeEsfuerzo struct {
	e                entorno
	proyecto         string
	ana, beto        string
	login, reporte   string
	sprint1, sprint2 string
}

func conEsfuerzoCargado(t *testing.T) escenarioDeEsfuerzo {
	t.Helper()
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	s := escenarioDeEsfuerzo{e: e, proyecto: proyecto}
	s.ana = e.registrarIntegrante(t, proyecto, "Ana")
	s.beto = e.registrarIntegrante(t, proyecto, "Beto")
	s.login = e.historiaConEstado(t, proyecto, "Login", 3, backlog.EstadoPendiente)
	s.reporte = e.historiaConEstado(t, proyecto, "Reporte", 5, backlog.EstadoPendiente)
	s.sprint1 = e.crearSprintConFechas(t, proyecto, "Primero", "2026-10-05", "2026-10-18")
	e.cerrarSprint(t, proyecto, s.sprint1)
	s.sprint2 = e.crearSprintConFechas(t, proyecto, "Segundo", "2026-10-19", "2026-11-01")
	e.cargarHoras(t, proyecto, s.login, s.ana, "2026-10-06", "Diseño", 2)
	e.cargarHoras(t, proyecto, s.login, s.beto, "2026-10-07", "Código", 3.5)
	e.cargarHoras(t, proyecto, s.reporte, s.ana, "2026-10-19", "Pruebas", 1.5)
	e.cargarHoras(t, proyecto, s.reporte, s.beto, "2026-10-02", "Análisis", 4)
	return s
}

// CA-T019-2: los filtros dejan solo lo que corresponde y los totales son los de lo filtrado.
func TestPantallasEsfuerzo_Filtros(t *testing.T) {
	s := conEsfuerzoCargado(t)
	ruta := rutaDeEsfuerzo(s.proyecto)

	todos := s.e.get(ruta)
	contiene(t, todos, "Total: 11 h", "Diseño", "Código", "Pruebas", "Análisis", "Sprint 1", "Sprint 2", "Fuera de sprint")
	if posicion(t, todos, "Pruebas") > posicion(t, todos, "Diseño") || posicion(t, todos, "Diseño") > posicion(t, todos, "Análisis") {
		t.Error("los registros tienen que salir por fecha descendente")
	}

	casos := []struct {
		nombre     string
		consulta   string
		total      string
		aparecen   []string
		noAparecen []string
		totales    []string
	}{
		{"por historia", "?historia=" + s.login, "Total: 5,5 h", []string{"Diseño", "Código"}, []string{"Pruebas", "Análisis"}, []string{"5,5 h"}},
		{"por integrante", "?integrante=" + s.ana, "Total: 3,5 h", []string{"Diseño", "Pruebas"}, []string{"Código", "Análisis"}, []string{"2 h", "1,5 h"}},
		{"por sprint", "?sprint=" + s.sprint1, "Total: 5,5 h", []string{"Diseño", "Código"}, []string{"Pruebas", "Análisis"}, []string{"Sprint 1"}},
		{"el primer día del sprint cuenta", "?sprint=" + s.sprint2, "Total: 1,5 h", []string{"Pruebas"}, []string{"Diseño", "Código", "Análisis"}, []string{"Sprint 2"}},
		{"historia y integrante", "?historia=" + s.reporte + "&integrante=" + s.beto, "Total: 4 h", []string{"Análisis"}, []string{"Diseño", "Código", "Pruebas"}, []string{"Fuera de sprint"}},
		{"filtros vacíos no filtran", "?historia=&integrante=&sprint=", "Total: 11 h", []string{"Diseño", "Código", "Pruebas", "Análisis"}, nil, nil},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := s.e.get(ruta + c.consulta)
			if w.Code != http.StatusOK {
				t.Fatalf("estado %d: %s", w.Code, w.Body.String())
			}
			contiene(t, w, append([]string{c.total}, append(c.aparecen, c.totales...)...)...)
			noContiene(t, w, c.noAparecen...)
		})
	}

	// Los totales de lo filtrado: por historia, por integrante y por sprint, con «Fuera de sprint» al final.
	w := s.e.get(ruta)
	contiene(t, w, "Por historia", "Por integrante", "Por sprint")
	totalesPorSprint := w.Body.String()[posicion(t, w, "Por sprint"):]
	if strings.LastIndex(totalesPorSprint, "Sprint 2") > strings.LastIndex(totalesPorSprint, "Fuera de sprint") {
		t.Error("«Fuera de sprint» tiene que ir al final de los totales por sprint")
	}
}

// Los filtros elegidos quedan elegidos en el formulario, y las listas salen en el orden de la spec.
func TestPantallasEsfuerzo_ListasYFiltrosElegidos(t *testing.T) {
	s := conEsfuerzoCargado(t)
	w := s.e.get(rutaDeEsfuerzo(s.proyecto) + "?historia=" + s.reporte + "&integrante=" + s.beto + "&sprint=" + s.sprint1)

	for _, id := range []string{s.reporte, s.beto, s.sprint1} {
		if !regexp.MustCompile(`value="` + regexp.QuoteMeta(id) + `"[^>]*selected`).MatchString(w.Body.String()) {
			t.Errorf("la opción %s no quedó elegida:\n%s", id, w.Body.String())
		}
	}
	contiene(t, w, `method="get"`, "Todas", "Todos", "Filtrar", "Ningún registro cumple los filtros")
	if posicion(t, w, "Sprint 2") > posicion(t, w, "Sprint 1") {
		t.Error("los sprints de las listas salen por número descendente")
	}
	if posicion(t, w, ">Ana<") > posicion(t, w, ">Beto<") {
		t.Error("los integrantes de las listas salen por nombre")
	}
}

// CA-T019-4, CL-T019-8, RN-T019-11: sin coincidencias hay un texto de vacío, el total en 0 y ningún error.
func TestPantallasEsfuerzo_SinCoincidencias(t *testing.T) {
	s := conEsfuerzoCargado(t)

	w := s.e.get(rutaDeEsfuerzo(s.proyecto) + "?historia=" + s.login + "&sprint=" + s.sprint2)

	if w.Code != http.StatusOK {
		t.Fatalf("estado %d", w.Code)
	}
	contiene(t, w, "Ningún registro cumple los filtros", "Total: 0 h")
	noContiene(t, w, "Por historia", "Todavía no hay horas registradas")
}

// CL-T019-6: una fecha entre dos sprints se acepta y sale como «Fuera de sprint».
func TestPantallasEsfuerzo_FechaEntreSprints(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")
	e.crearHistoria(t, proyecto, "Historia", "must", "")
	historia := e.idDeHistoria(t, proyecto, "Historia")
	primero := e.crearSprintConFechas(t, proyecto, "Primero", "2026-10-05", "2026-10-09")
	e.cerrarSprint(t, proyecto, primero)
	e.crearSprintConFechas(t, proyecto, "Segundo", "2026-10-14", "2026-10-18")

	if w := e.post(rutaDeEsfuerzo(proyecto), datosDeHoras(ana, historia, "2026-10-11", "Entre sprints", "1")); w.Code != http.StatusSeeOther {
		t.Fatalf("registrar: %d %s", w.Code, w.Body.String())
	}
	contiene(t, e.get(rutaDeEsfuerzo(proyecto)), "Entre sprints", "Fuera de sprint")
}

// CL-T019-9: se puede registrar en una historia completada o que no está en ningún sprint.
func TestPantallasEsfuerzo_HistoriaCompletada(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")
	hecha := e.historiaConEstado(t, proyecto, "Hecha", 2, backlog.EstadoCompletado)

	if w := e.post(rutaDeEsfuerzo(proyecto), datosDeHoras(ana, hecha, "2026-10-20", "Retoques", "1")); w.Code != http.StatusSeeOther {
		t.Fatalf("registrar: %d %s", w.Code, w.Body.String())
	}
}

// CA-T019-3, CL-T019-10, CL-T019-11: la comparación muestra lo que devuelve US-021, la que más se desvió primero.
func TestPantallasEsfuerzo_Comparacion(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")
	a := e.historiaConEstado(t, proyecto, "Alta exacta", 2, backlog.EstadoCompletado)
	b := e.historiaConEstado(t, proyecto, "Baja grande", 5, backlog.EstadoCompletado)
	c := e.historiaConEstado(t, proyecto, "Sigue en curso", 3, backlog.EstadoEnProgreso)
	d := e.historiaConEstado(t, proyecto, "Sin estimar", 0, backlog.EstadoPendiente)
	e.historiaConEstado(t, proyecto, "Hecha sin horas", 1, backlog.EstadoCompletado)
	e.cargarHoras(t, proyecto, a, ana, "2026-10-06", "Trabajo A", 3.5)
	e.cargarHoras(t, proyecto, b, ana, "2026-10-07", "Trabajo B", 4.5)
	e.cargarHoras(t, proyecto, c, ana, "2026-10-08", "Trabajo C", 2)
	e.cargarHoras(t, proyecto, d, ana, "2026-10-09", "Trabajo D", 2)

	w := e.get(rutaDeEsfuerzo(proyecto) + "/comparacion")

	if w.Code != http.StatusOK {
		t.Fatalf("estado %d: %s", w.Code, w.Body.String())
	}
	contiene(t, w, "Completada", "En curso", "Sin estimar",
		"+0,5", "16,7 %", "-3", "-40,0 %", "-2,5", "-55,6 %", "-1,5", "-100,0 %",
		"3 historias comparables: estimadas 12 h", "reales 8 h", "desviación -4 h (-33,3 %)", "Horas sin comparar: 4 h",
		"Volver a la carga de horas", rutaDeEsfuerzo(proyecto))
	// Primero donde más nos equivocamos: 3, 2,5, 2, 1,5 y 0,5 horas de desviación.
	orden := []string{"Baja grande", "Sigue en curso", "Sin estimar", "Hecha sin horas", "Alta exacta"}
	for i := 1; i < len(orden); i++ {
		if posicion(t, w, orden[i-1]) > posicion(t, w, orden[i]) {
			t.Errorf("«%s» tiene que salir antes que «%s»", orden[i-1], orden[i])
		}
	}
}

// La historia sin estimar no muestra una desviación inventada (CL-T019-11).
func TestPantallasEsfuerzo_ComparacionSinEstimar(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")
	d := e.historiaConEstado(t, proyecto, "Sin estimar", 0, backlog.EstadoPendiente)
	e.cargarHoras(t, proyecto, d, ana, "2026-10-09", "Trabajo D", 2)

	w := e.get(rutaDeEsfuerzo(proyecto) + "/comparacion")

	contiene(t, w, "Sin estimar", "0 historias comparables", "Horas sin comparar: 2 h", "—")
	noContiene(t, w, "+2")
}

// Una sola historia comparable se escribe en singular.
func TestPantallasEsfuerzo_ComparacionUnaHistoria(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")
	a := e.historiaConEstado(t, proyecto, "Única", 2, backlog.EstadoCompletado)
	e.cargarHoras(t, proyecto, a, ana, "2026-10-06", "Trabajo", 3)

	w := e.get(rutaDeEsfuerzo(proyecto) + "/comparacion")

	contiene(t, w, "1 historia comparable: estimadas 3 h, reales 3 h, desviación 0 h (0,0 %)", "Horas sin comparar: 0 h")
}

// RN-T019-11: sin historias en la comparación hay un texto de vacío y no un error.
func TestPantallasEsfuerzo_ComparacionVacia(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")

	w := e.get(rutaDeEsfuerzo(proyecto) + "/comparacion")

	if w.Code != http.StatusOK {
		t.Fatalf("estado %d", w.Code)
	}
	contiene(t, w, "Todavía no hay historias para comparar", "Volver a la carga de horas")
	noContiene(t, w, "Horas sin comparar")
}

// RN-T019-13: cada campo tiene su etiqueta y cada fila, un nombre accesible que dice a qué registro corresponde.
func TestPantallasEsfuerzo_Accesibilidad(t *testing.T) {
	s := conEsfuerzoCargado(t)
	w := s.e.get(rutaDeEsfuerzo(s.proyecto))

	for _, campo := range []string{"integrante", "historia", "fecha", "actividad", "horas"} {
		contiene(t, w, `for="`+campo+`"`, `id="`+campo+`"`)
	}
	for _, filtro := range []string{"filtro-historia", "filtro-integrante", "filtro-sprint"} {
		contiene(t, w, `for="`+filtro+`"`, `id="`+filtro+`"`)
	}
	contiene(t, w, "<caption", `scope="col"`)
	if n := strings.Count(w.Body.String(), "<tr"); n < 5 {
		t.Errorf("hay %d filas en la página, se esperaban al menos las 4 de los registros", n)
	}
}
