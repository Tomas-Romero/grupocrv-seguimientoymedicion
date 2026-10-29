package httpadapter_test

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// Dashboard del proyecto (spec US-030). El reloj de los tests es el 2026-10-20.

func rutaDelDashboard(proyectoID string) string { return "/proyectos/" + proyectoID + "/dashboard" }

// proyectoConDashboard arma un proyecto con dos sprints cerrados (10 Story Points con 8 completados, y 5 con 5), un sprint abierto
// con 8 Story Points y 5 completados, 8 horas dentro del sprint abierto y tres defectos.
func proyectoConDashboard(t *testing.T) (e entorno, proyecto string) {
	t.Helper()
	e = nuevoEntorno(t)
	proyecto = e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")

	asignarYCompletar := func(sprintID string, historias map[string]string, completar ...string) map[string]string {
		ids := map[string]string{}
		for titulo, puntos := range historias {
			id := e.historiaLista(t, proyecto, titulo, puntos)
			ids[titulo] = id
			if w := e.post(rutaDeSprint(proyecto, sprintID)+"/historias", url.Values{"item": {id}}); w.Code != http.StatusSeeOther {
				t.Fatalf("asignar %q: %d %s", titulo, w.Code, w.Body.String())
			}
		}
		for _, titulo := range completar {
			if w := e.post(rutaDeSprint(proyecto, sprintID)+"/historias/"+ids[titulo]+"/completar", nil); w.Code != http.StatusSeeOther {
				t.Fatalf("completar %q: %d %s", titulo, w.Code, w.Body.String())
			}
		}
		return ids
	}

	s1 := e.crearSprintConFechas(t, proyecto, "Primero", "2026-10-01", "2026-10-07")
	ids1 := asignarYCompletar(s1, map[string]string{"Login": "8", "Reporte": "2"}, "Login")
	e.cerrarSprint(t, proyecto, s1)
	s2 := e.crearSprintConFechas(t, proyecto, "Segundo", "2026-10-08", "2026-10-14")
	asignarYCompletar(s2, map[string]string{"Perfil": "5"}, "Perfil")
	e.cerrarSprint(t, proyecto, s2)
	s3 := e.crearSprintConFechas(t, proyecto, "Tercero", "2026-10-15", "2026-10-25")
	ids3 := asignarYCompletar(s3, map[string]string{"Busqueda": "5", "Filtros": "3"}, "Busqueda")

	e.cargarHoras(t, proyecto, ids3["Busqueda"], ana, "2026-10-16", "Código", 3)
	e.cargarHoras(t, proyecto, ids3["Filtros"], ana, "2026-10-17", "Pruebas", 5)
	e.cargarHoras(t, proyecto, ids1["Login"], ana, "2026-10-03", "Código del sprint 1", 4) // no es del sprint de referencia

	registrar := func(item, sprintID string, severidad defecto.Severidad, estado defecto.Estado) {
		d := defecto.Defecto{ProyectoID: proyecto, ItemID: item, SprintDeteccionID: sprintID, Descripcion: "Falla", Severidad: severidad, Estado: estado}
		if _, err := e.base.RegistrarDefecto(context.Background(), d); err != nil {
			t.Fatalf("RegistrarDefecto: %v", err)
		}
	}
	registrar(ids1["Login"], s1, defecto.SeveridadAlta, defecto.EstadoAbierto)
	registrar(ids1["Login"], s2, defecto.SeveridadCritica, defecto.EstadoEnCorreccion)
	registrar(ids1["Reporte"], s1, defecto.SeveridadBaja, defecto.EstadoDescartado)
	return e, proyecto
}

// CA-030-1, CA-030-2, CA-030-3, RN-030-12: la página del dashboard junta el estado del proyecto, la tarjeta del sprint abierto, la
// velocidad con la tabla de sprints cerrados (el más reciente primero) y los defectos, con el menú «Dashboard» activo.
func TestPantallasDashboard_Tarjetas(t *testing.T) {
	e, proyecto := proyectoConDashboard(t)

	w := e.get(rutaDelDashboard(proyecto))

	if w.Code != http.StatusOK {
		t.Fatalf("GET dashboard: %d %s", w.Code, w.Body.String())
	}
	contiene(t, w, "<!doctype html>", "Dashboard", "Estado general", "En curso: faltan 56 días",
		"Sprint 3", "abierto", "Story Points: 5 completados de 8", "Historias completadas: 50,0 %",
		"Horas estimadas: 12 h", "Horas reales: 8 h", "Desviación: -4 h (-33,3 %)",
		"Velocidad del equipo: 6,5", "Defectos: 3",
		"Críticos: 1", "Altos: 1", "Medios: 0", "Bajos: 1", "Abiertos: 1", "En corrección: 1", "Resueltos: 0", "Descartados: 1",
		`aria-current="page"`, rutaDelDashboard(proyecto))
	if posicion(t, w, ">Sprint 2<") > posicion(t, w, ">Sprint 1<") {
		t.Error("la tabla de sprints cerrados tiene que ir del más reciente al más antiguo")
	}
}

// RN-030-8, RN-030-6: una fila por sprint cerrado con sus Story Points y sus defectos detectados y resueltos; los defectos
// descartados no cuentan en las columnas por sprint.
func TestPantallasDashboard_TablaDeSprintsCerrados(t *testing.T) {
	e, proyecto := proyectoConDashboard(t)

	w := e.get(rutaDelDashboard(proyecto))

	fila1 := regexp.MustCompile(`(?s)<tr[^>]*>\s*<th scope="row"[^>]*>Sprint 1</th>\s*<td[^>]*>10</td>\s*<td[^>]*>8</td>\s*<td[^>]*>1</td>\s*<td[^>]*>0</td>`)
	fila2 := regexp.MustCompile(`(?s)<tr[^>]*>\s*<th scope="row"[^>]*>Sprint 2</th>\s*<td[^>]*>5</td>\s*<td[^>]*>5</td>\s*<td[^>]*>1</td>\s*<td[^>]*>0</td>`)
	if !fila1.MatchString(w.Body.String()) {
		t.Errorf("no está la fila del Sprint 1 (10 planificados, 8 completados, 1 detectado y 0 resueltos):\n%s", w.Body.String())
	}
	if !fila2.MatchString(w.Body.String()) {
		t.Errorf("no está la fila del Sprint 2 (5, 5, 1 detectado y 0 resueltos)")
	}
}

// CA-030-6, RN-030-14, RN-030-16: el gráfico de Story Points es un SVG con rol, título y descripción con los mismos números,
// los grupos en orden cronológico, el valor escrito en cada barra y las alturas de una escala común (100, 80, 50 y 50).
func TestPantallasDashboard_GraficoDeStoryPoints(t *testing.T) {
	e, proyecto := proyectoConDashboard(t)

	w := e.get(rutaDelDashboard(proyecto))

	contiene(t, w, `role="img"`, "Story Points planificados y completados por sprint cerrado",
		"Sprint 1: 10 planificados, 8 completados", "Sprint 2: 5 planificados, 5 completados", "Planificados", "Completados")
	for _, barra := range []struct{ sprint, serie, alto string }{
		{"1", "planificados", "100"}, {"1", "completados", "80"}, {"2", "planificados", "50"}, {"2", "completados", "50"},
	} {
		patron := regexp.MustCompile(`data-sprint="` + barra.sprint + `" data-serie="` + barra.serie + `"[^>]*height="` + barra.alto + `"`)
		if !patron.MatchString(w.Body.String()) {
			t.Errorf("no está la barra del sprint %s (%s) con alto %s", barra.sprint, barra.serie, barra.alto)
		}
	}
	if svg := regexp.MustCompile(`(?s)<svg.*?</svg>`).FindString(w.Body.String()); regexp.MustCompile(`\d\.\d{3,}`).MatchString(svg) {
		t.Errorf("las medidas del SVG no tienen que arrastrar decimales de coma flotante:\n%s", svg)
	}
	if posicion(t, w, "Sprint 1: 10 planificados") > posicion(t, w, "Sprint 2: 5 planificados") {
		t.Error("los grupos del gráfico van del sprint más antiguo al más reciente")
	}
}

// RN-030-15, RN-030-16: el gráfico de historias completadas dibuja el porcentaje del sprint de referencia, con el valor escrito.
func TestPantallasDashboard_GraficoDePorcentaje(t *testing.T) {
	e, proyecto := proyectoConDashboard(t)

	w := e.get(rutaDelDashboard(proyecto))

	contiene(t, w, "Historias completadas del Sprint 3", "50,0 % de las historias completadas")
	if !regexp.MustCompile(`data-serie="porcentaje"[^>]*width="50"`).MatchString(w.Body.String()) {
		t.Errorf("la barra del porcentaje tiene que medir 50 sobre 100:\n%s", w.Body.String())
	}
}

// CA-030-4, CA-030-7, CL-030-3, CL-030-9, RN-030-17: un proyecto nuevo, sin sprints, horas ni defectos, muestra los textos de vacío, no dibuja
// ningún gráfico y no tiene error.
func TestPantallasDashboard_ProyectoNuevo(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Nuevo")

	w := e.get(rutaDelDashboard(proyecto))

	if w.Code != http.StatusOK {
		t.Fatalf("GET dashboard: %d", w.Code)
	}
	contiene(t, w, "Sin sprints cerrados", "Sin sprints", "Sin defectos", "Estado general")
	noContiene(t, w, "<svg", "NaN", `role="alert"`)
}

// CL-030-1, CL-030-2: con todos los sprints cerrados, la tarjeta del sprint es la del de número más alto.
func TestPantallasDashboard_ReferenciaConTodosCerrados(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Cerrados")
	s1 := e.crearSprintConFechas(t, proyecto, "Primero", "2026-10-01", "2026-10-07")
	e.cerrarSprint(t, proyecto, s1)
	s2 := e.crearSprintConFechas(t, proyecto, "Segundo", "2026-10-08", "2026-10-14")
	e.cerrarSprint(t, proyecto, s2)

	w := e.get(rutaDelDashboard(proyecto))

	contiene(t, w, "Sprint 2", "cerrado", "Historias completadas del Sprint 2")
	noContiene(t, w, "Sin sprints</")
}

// CA-030-5, CL-030-11, RN-030-10: un proyecto inexistente o con ID mal formado responde 404 con layout.
func TestPantallasDashboard_ProyectoInexistente(t *testing.T) {
	e := nuevoEntorno(t)
	for _, id := range []string{"no-existe", "00000000-0000-0000-0000-00000000dead"} {
		w := e.get(rutaDelDashboard(id))
		if w.Code != http.StatusNotFound {
			t.Errorf("GET dashboard de %s: %d, se esperaba 404", id, w.Code)
		}
		contiene(t, w, "<!doctype html>")
	}
}

// RN-030-12: con HTMX se responde como fragmento (sin la página), igual que las demás pantallas.
func TestPantallasDashboard_Fragmento(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Con HTMX")

	w := e.enviar(http.MethodGet, rutaDelDashboard(proyecto), nil, true)

	contiene(t, w, "Dashboard")
	noContiene(t, w, "<html")
}
