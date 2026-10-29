package httpadapter_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// Reportes en pantalla de un proyecto y de un sprint (spec US-033). El reloj de los tests es el 2026-10-20.

func rutaDelReporte(proyectoID string) string { return "/proyectos/" + proyectoID + "/reporte" }

func rutaDelReporteDelSprint(proyectoID, sprintID string) string {
	return rutaDeSprint(proyectoID, sprintID) + "/reporte"
}

// idsDeLosReportes son los identificadores de las secciones, en el orden de la spec (seccion 3).
var (
	seccionesDelProyecto = []string{"proyecto", "integrantes", "backlog", "sprints", "esfuerzo", "defectos", "metricas"}
	seccionesDelSprint   = []string{"sprint", "historias", "metricas", "esfuerzo", "defectos"}
)

// proyectoParaReportes arma un proyecto con dos sprints cerrados y uno abierto: en el Sprint 1 una historia se completa (Login, 8 Story
// Points) y otra vuelve al backlog (Reporte, 2) y despues se completa en el Sprint 2; el Sprint 3 esta abierto con una historia de 5 Story
// Points completada y otra de 3 sin completar. Hay horas dentro de los sprints 1 y 3 y una fuera de todo sprint, y cuatro defectos: uno
// detectado en el Sprint 1 y resuelto en el 2, uno abierto del Sprint 3, uno descartado y otro con un script en la descripcion.
func proyectoParaReportes(t *testing.T) (e entorno, proyecto string, sprints [3]string) {
	t.Helper()
	e = nuevoEntorno(t)
	proyecto = e.crearProyecto(t, "Demo")
	ana := e.registrarIntegrante(t, proyecto, "Ana")

	asignar := func(sprintID string, historias map[string]string) map[string]string {
		ids := map[string]string{}
		for titulo, puntos := range historias {
			id := e.historiaLista(t, proyecto, titulo, puntos)
			ids[titulo] = id
			if w := e.post(rutaDeSprint(proyecto, sprintID)+"/historias", url.Values{"item": {id}}); w.Code != http.StatusSeeOther {
				t.Fatalf("asignar %q: %d %s", titulo, w.Code, w.Body.String())
			}
		}
		return ids
	}
	completar := func(sprintID, itemID string) {
		if w := e.post(rutaDeSprint(proyecto, sprintID)+"/historias/"+itemID+"/completar", nil); w.Code != http.StatusSeeOther {
			t.Fatalf("completar: %d %s", w.Code, w.Body.String())
		}
	}

	sprints[0] = e.crearSprintConFechas(t, proyecto, "Primero", "2026-10-01", "2026-10-07")
	ids1 := asignar(sprints[0], map[string]string{"Login": "8", "Reporte": "2"})
	completar(sprints[0], ids1["Login"])
	e.cerrarSprint(t, proyecto, sprints[0])

	sprints[1] = e.crearSprintConFechas(t, proyecto, "Segundo", "2026-10-08", "2026-10-14")
	// La historia que volvio al backlog se asigna otra vez y se completa en este sprint.
	if w := e.post(rutaDeSprint(proyecto, sprints[1])+"/historias", url.Values{"item": {ids1["Reporte"]}}); w.Code != http.StatusSeeOther {
		t.Fatalf("reasignar: %d %s", w.Code, w.Body.String())
	}
	completar(sprints[1], ids1["Reporte"])
	e.cerrarSprint(t, proyecto, sprints[1])

	sprints[2] = e.crearSprintConFechas(t, proyecto, "Tercero", "2026-10-15", "2026-10-25")
	ids3 := asignar(sprints[2], map[string]string{"Busqueda": "5", "Filtros": "3"})
	completar(sprints[2], ids3["Busqueda"])

	e.cargarHoras(t, proyecto, ids1["Login"], ana, "2026-10-03", "Código del sprint 1", 4)
	e.cargarHoras(t, proyecto, ids3["Busqueda"], ana, "2026-10-16", "Código", 3)
	e.cargarHoras(t, proyecto, ids3["Filtros"], ana, "2026-10-17", "Pruebas", 5)
	e.cargarHoras(t, proyecto, ids3["Busqueda"], ana, "2026-10-02", "Análisis", 1)        // dentro del Sprint 1
	e.cargarHoras(t, proyecto, ids3["Filtros"], ana, "2026-09-30", "Fuera del sprint", 2) // antes del primer sprint

	registrar := func(item, deteccion string, severidad defecto.Severidad, estado defecto.Estado, resolucion, descripcion string) {
		d := defecto.Defecto{ProyectoID: proyecto, ItemID: item, SprintDeteccionID: deteccion, SprintResolucionID: resolucion, Descripcion: descripcion, Severidad: severidad, Estado: estado}
		if _, err := e.base.RegistrarDefecto(context.Background(), d); err != nil {
			t.Fatalf("RegistrarDefecto: %v", err)
		}
	}
	registrar(ids1["Login"], sprints[0], defecto.SeveridadAlta, defecto.EstadoResuelto, sprints[1], "Falla al guardar")
	registrar(ids3["Busqueda"], sprints[2], defecto.SeveridadCritica, defecto.EstadoAbierto, "", "Pierde los datos")
	registrar(ids1["Login"], sprints[0], defecto.SeveridadBaja, defecto.EstadoDescartado, "", "Descartado para siempre")
	registrar(ids3["Filtros"], sprints[2], defecto.SeveridadMedia, defecto.EstadoAbierto, "", "<script>alert(1)</script>")
	return e, proyecto, sprints
}

// ordenDeLasSecciones comprueba que cada seccion esta una sola vez, con su id, y en este orden.
func ordenDeLasSecciones(t *testing.T, cuerpo string, ids []string) {
	t.Helper()
	anterior := -1
	for _, id := range ids {
		marca := `<section id="` + id + `"`
		if strings.Count(cuerpo, marca) != 1 {
			t.Errorf("la seccion %q tiene que estar exactamente una vez", id)
			continue
		}
		if i := strings.Index(cuerpo, marca); i < anterior {
			t.Errorf("la seccion %q esta antes que la anterior: el orden es fijo (%v)", id, ids)
		} else {
			anterior = i
		}
	}
}

// CA-033-1, CA-033-5, CL-033-1, RN-033-2, RN-033-8: el reporte del proyecto es una pagina con sus siete secciones, siempre en el mismo orden,
// cada una dentro de un <section> con su id, con el menu «Reporte» activo.
func TestPantallasReporte_ProyectoConSusSieteSecciones(t *testing.T) {
	e, proyecto, _ := proyectoParaReportes(t)

	w := e.get(rutaDelReporte(proyecto))

	if w.Code != http.StatusOK {
		t.Fatalf("GET reporte: %d %s", w.Code, w.Body.String())
	}
	ordenDeLasSecciones(t, w.Body.String(), seccionesDelProyecto)
	contiene(t, w, "<!doctype html>", "Reporte del proyecto", "Demo", `aria-current="page"`, rutaDelReporte(proyecto))
	noContiene(t, w, "<script>alert")
}

// CA-033-1, RN-033-1: cada seccion trae las cifras que ya muestran las demas pantallas.
func TestPantallasReporte_ContenidoDeCadaSeccion(t *testing.T) {
	e, proyecto, _ := proyectoParaReportes(t)

	w := e.get(rutaDelReporte(proyecto))

	contiene(t, w,
		// 1. Proyecto: el estado y el avance del backlog (US-004)
		"Estado general", "En curso: faltan 56 días", "Pendientes: 1 · En progreso: 0 · Completadas: 3 (de 4)", "Story Points: 15 completados de 18",
		// 2. Integrantes
		"Ana",
		// 3. Product Backlog
		"Login", "Reporte", "Busqueda", "Filtros",
		// 4. Sprints: Story Points, estado y las historias que volvieron al backlog del Sprint 1
		"Sprint 3", "abierto", "Sprint 1", "cerrado",
		// 5. Esfuerzo: el total de todos los registros, tambien los de fuera de todo sprint, y las horas sin comparar
		"Total de horas registradas: 15 h", "Horas sin comparar",
		// 6. Defectos: la lista con todos los estados y los totales por severidad y por estado, sin los descartados en los de cada sprint
		"Falla al guardar", "Pierde los datos", "Descartado para siempre", "Defectos: 4", "Críticos: 1", "Altos: 1", "Medios: 1", "Bajos: 1", "Descartados: 1",
		// 7. Metricas (US-030)
		"Velocidad del equipo", "Sprint de referencia: Sprint 3",
	)
}

// CL-033-10, RN-033-9: un texto con HTML en la descripcion de un defecto se muestra escapado.
func TestPantallasReporte_TextoConHTMLEscapado(t *testing.T) {
	e, proyecto, _ := proyectoParaReportes(t)

	w := e.get(rutaDelReporte(proyecto))

	contiene(t, w, "&lt;script&gt;alert(1)&lt;/script&gt;")
	noContiene(t, w, "<script>alert(1)</script>")
}

// CA-033-3, CL-033-2, RN-033-2: un proyecto recien creado tiene las siete secciones, cada una con «Sin datos», sin error ni NaN.
func TestPantallasReporte_ProyectoNuevoConTodasLasSecciones(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Nuevo")

	w := e.get(rutaDelReporte(proyecto))

	if w.Code != http.StatusOK {
		t.Fatalf("GET reporte: %d", w.Code)
	}
	ordenDeLasSecciones(t, w.Body.String(), seccionesDelProyecto)
	// Seis secciones vacias: el proyecto siempre tiene datos.
	if n := strings.Count(w.Body.String(), "Sin datos"); n != 6 {
		t.Errorf("«Sin datos» aparece %d veces, se esperaban 6 (una por seccion vacia)", n)
	}
	noContiene(t, w, "NaN", `role="alert"`)
}

// CA-033-4, CL-033-8, RN-033-6: un proyecto inexistente o con el ID mal formado responde 404 con layout.
func TestPantallasReporte_ProyectoInexistente(t *testing.T) {
	e := nuevoEntorno(t)
	for _, id := range []string{"no-existe", "00000000-0000-0000-0000-00000000dead"} {
		w := e.get(rutaDelReporte(id))
		if w.Code != http.StatusNotFound {
			t.Errorf("GET reporte de %s: %d, se esperaba 404", id, w.Code)
		}
		contiene(t, w, "<!doctype html>")
	}
}

// RN-033-8: con HTMX se responde como fragmento (sin la pagina), igual que las demas pantallas.
func TestPantallasReporte_Fragmento(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Con HTMX")

	w := e.enviar(http.MethodGet, rutaDelReporte(proyecto), nil, true)

	contiene(t, w, "Reporte del proyecto")
	noContiene(t, w, "<html")
}

// CA-033-2, CA-033-5, RN-033-2: el reporte de un sprint tiene sus cinco secciones, siempre en el mismo orden.
func TestPantallasReporte_SprintConSusCincoSecciones(t *testing.T) {
	e, proyecto, sprints := proyectoParaReportes(t)

	w := e.get(rutaDelReporteDelSprint(proyecto, sprints[2]))

	if w.Code != http.StatusOK {
		t.Fatalf("GET reporte del sprint: %d %s", w.Code, w.Body.String())
	}
	ordenDeLasSecciones(t, w.Body.String(), seccionesDelSprint)
	contiene(t, w, "Reporte del Sprint 3", "Tercero", "abierto", "en curso")
}

// CA-033-2, RN-033-4: las metricas del sprint (Story Points, porcentaje y horas de los registros con fecha dentro de el) y sus historias.
func TestPantallasReporte_ContenidoDelSprint(t *testing.T) {
	e, proyecto, sprints := proyectoParaReportes(t)

	w := e.get(rutaDelReporteDelSprint(proyecto, sprints[2]))

	contiene(t, w,
		"Busqueda", "Completada", "Filtros", "No completada",
		"Story Points: 5 completados de 8", "Historias completadas: 50,0 %",
		"Horas estimadas: 12 h", "Horas reales: 8 h", "Desviación: -4 h (-33,3 %)",
		"Total de horas del sprint: 8 h", "Pierde los datos", "Detectados: 2", "Resueltos: 0",
	)
	// Las horas de fuera de todo sprint (2 h el 26/10) y las del Sprint 1 no estan en este reporte.
	noContiene(t, w, "Fuera del sprint", "Código del sprint 1")
}

// CL-033-3, RN-033-4: una historia que volvio al backlog del Sprint 1 y se completo en el Sprint 2 es «No completada» en el reporte del
// primero y «Completada» en el del segundo.
func TestPantallasReporte_ElResultadoEsElDeLaFilaDeCadaSprint(t *testing.T) {
	e, proyecto, sprints := proyectoParaReportes(t)

	primero := e.get(rutaDelReporteDelSprint(proyecto, sprints[0]))
	segundo := e.get(rutaDelReporteDelSprint(proyecto, sprints[1]))

	contiene(t, primero, "Login", "Completada", "Reporte", "No completada")
	if !strings.Contains(segundo.Body.String(), "Reporte") || strings.Contains(segundo.Body.String(), "No completada") {
		t.Errorf("el reporte del Sprint 2 tiene que mostrar «Reporte» como completada y ninguna «No completada»:\n%s", segundo.Body.String())
	}
}

// CL-033-5, CL-033-6, RN-033-5: el defecto detectado en el Sprint 1 y resuelto en el 2 es «detectado» en el primero y «resuelto» en el
// segundo, y el descartado no cuenta en ninguno.
func TestPantallasReporte_DefectosDetectadosYResueltosPorSprint(t *testing.T) {
	e, proyecto, sprints := proyectoParaReportes(t)

	primero := e.get(rutaDelReporteDelSprint(proyecto, sprints[0]))
	segundo := e.get(rutaDelReporteDelSprint(proyecto, sprints[1]))

	contiene(t, primero, "Detectados: 1", "Resueltos: 0", "Falla al guardar")
	noContiene(t, primero, "Descartado para siempre")
	contiene(t, segundo, "Detectados: 0", "Resueltos: 1", "Falla al guardar")
}

// CL-033-4, RN-033-4: las horas con fecha fuera de todo sprint no estan en el reporte de ningun sprint, pero si en el total del proyecto.
func TestPantallasReporte_HorasFueraDeTodoSprint(t *testing.T) {
	e, proyecto, sprints := proyectoParaReportes(t)

	delProyecto := e.get(rutaDelReporte(proyecto))
	for i, s := range sprints {
		del := e.get(rutaDelReporteDelSprint(proyecto, s))
		noContiene(t, del, "Fuera del sprint")
		_ = i
	}

	contiene(t, delProyecto, "Total de horas registradas: 15 h")
}

// CA-033-3, CL-033-9: un sprint sin historias, horas ni defectos muestra sus secciones con «Sin datos» y el porcentaje en 0.
func TestPantallasReporte_SprintSinDatos(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Vacio")
	s := e.crearSprintConFechas(t, proyecto, "Primero", "2026-10-01", "2026-10-07")

	w := e.get(rutaDelReporteDelSprint(proyecto, s))

	if w.Code != http.StatusOK {
		t.Fatalf("GET reporte del sprint: %d", w.Code)
	}
	ordenDeLasSecciones(t, w.Body.String(), seccionesDelSprint)
	// Historias, esfuerzo y defectos vacios.
	if n := strings.Count(w.Body.String(), "Sin datos"); n != 3 {
		t.Errorf("«Sin datos» aparece %d veces, se esperaban 3", n)
	}
	contiene(t, w, "Historias completadas: 0,0 %", "Horas reales: 0 h")
	noContiene(t, w, "NaN", `role="alert"`)
}

// CA-033-4, CL-033-8, RN-033-6: un sprint que no existe, mal formado o de otro proyecto responde 404 con layout.
func TestPantallasReporte_SprintInexistenteODeOtroProyecto(t *testing.T) {
	e, proyecto, sprints := proyectoParaReportes(t)
	otro := e.crearProyecto(t, "Otro")

	casos := map[string]string{
		"no existe":          rutaDelReporteDelSprint(proyecto, "00000000-0000-0000-0000-00000000dead"),
		"mal formado":        rutaDelReporteDelSprint(proyecto, "no-existe"),
		"de otro proyecto":   rutaDelReporteDelSprint(otro, sprints[0]),
		"proyecto no existe": rutaDelReporteDelSprint("no-existe", sprints[0]),
	}
	for nombre, ruta := range casos {
		w := e.get(ruta)
		if w.Code != http.StatusNotFound {
			t.Errorf("%s: %d, se esperaba 404", nombre, w.Code)
		}
		contiene(t, w, "<!doctype html>")
	}
}

// RN-033-8: el menu del proyecto y el detalle de cada sprint tienen el enlace «Reporte».
func TestPantallasReporte_Enlaces(t *testing.T) {
	e, proyecto, sprints := proyectoParaReportes(t)

	resumen := e.get("/proyectos/" + proyecto)
	detalle := e.get(rutaDeSprint(proyecto, sprints[2]))

	contiene(t, resumen, `href="`+rutaDelReporte(proyecto)+`"`)
	contiene(t, detalle, `href="`+rutaDelReporteDelSprint(proyecto, sprints[2])+`"`)
}
