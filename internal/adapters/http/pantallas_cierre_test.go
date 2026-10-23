package httpadapter_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// datosDelSegundoSprint son las fechas del sprint siguiente al que crea datosDeSprint.
func datosDelSegundoSprint(goal string) url.Values {
	return url.Values{"sprint_goal": {goal}, "fecha_inicio": {"2026-10-19"}, "fecha_fin": {"2026-11-01"}}
}

// sprintConDosHistorias crea un proyecto con un sprint que tiene dos historias asignadas, de 5 y 3
// puntos, y marca la primera como completada. Devuelve los IDs.
func sprintConDosHistorias(t *testing.T, e entorno) (proyecto, sprintID, clave, opcional string) {
	t.Helper()
	proyecto = e.crearProyecto(t, "Demo")
	clave = e.historiaLista(t, proyecto, "Historia clave", "5")
	opcional = e.historiaLista(t, proyecto, "Historia opcional", "3")
	sprintID = e.crearSprint(t, proyecto, "Primer incremento")
	for _, id := range []string{clave, opcional} {
		if w := e.post(rutaDeSprint(proyecto, sprintID)+"/historias", url.Values{"item": {id}}); w.Code != http.StatusSeeOther {
			t.Fatalf("asignar: %d %s", w.Code, w.Body.String())
		}
	}
	if w := e.post(rutaDeSprint(proyecto, sprintID)+"/historias/"+clave+"/completar", nil); w.Code != http.StatusSeeOther {
		t.Fatalf("completar: %d %s", w.Code, w.Body.String())
	}
	return proyecto, sprintID, clave, opcional
}

// CA-T016-1, CA-T016-2, CL-T016-8: el recorrido completo con el router real y la base en memoria. El
// detalle de un sprint abierto ofrece cerrarlo, con confirmacion; al cerrar, el detalle muestra el
// resultado de cada historia y cuantas volvieron al backlog, y la devuelta se puede asignar al
// sprint siguiente.
func TestPantallasCierre_Recorrido(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto, sprintID, _, _ := sprintConDosHistorias(t, e)
	cerrar := rutaDeSprint(proyecto, sprintID) + "/cerrar"

	// RN-T016-2: el boton tiene un nombre accesible y pide confirmacion diciendo que pasa.
	abierto := e.get(rutaDeSprint(proyecto, sprintID))
	contiene(t, abierto, `hx-post="`+cerrar+`"`, `aria-label="Cerrar Sprint 1"`, "hx-confirm=", "vuelven al backlog", "no se puede reabrir")

	w := e.post(cerrar, nil)
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != rutaDeSprint(proyecto, sprintID) {
		t.Fatalf("cerrar: %d, Location=%q, cuerpo %s; quiero 303 al detalle", w.Code, w.Header().Get("Location"), w.Body.String())
	}

	// RN-T016-4, RN-T016-5: cada historia con su resultado, y las que volvieron al backlog.
	cerrado := e.get(rutaDeSprint(proyecto, sprintID))
	contiene(t, cerrado, "cerrado", "Planificados: 8", "Completados: 5", "Completada", "No completada · volvió al backlog", "Volvieron al backlog: 1 historia")
	noContiene(t, cerrado, "Cerrar sprint", "Marcar completada", "Desmarcar", `name="item"`, "hx-confirm")

	// RN-T016-7, CL-T016-8: ya se puede crear el siguiente y la devuelta se le asigna.
	creado := e.post("/proyectos/"+proyecto+"/sprints", datosDelSegundoSprint("Segundo incremento"))
	if creado.Code != http.StatusSeeOther {
		t.Fatalf("crear el sprint siguiente: %d %s", creado.Code, creado.Body.String())
	}
	segundo := idDeSprintEnLaRuta.FindStringSubmatch(creado.Header().Get("Location"))[2]
	disponibles := e.get(rutaDeSprint(proyecto, segundo))
	contiene(t, disponibles, "<option", "Historia opcional")
	noContiene(t, disponibles, "Historia clave")
}

// CA-T016-2: con HTMX, cerrar responde HX-Redirect al detalle con el cuerpo vacio.
func TestPantallasCierre_ConHTMX(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto, sprintID, _, _ := sprintConDosHistorias(t, e)

	w := e.enviar(http.MethodPost, rutaDeSprint(proyecto, sprintID)+"/cerrar", nil, true)

	if w.Code != http.StatusOK || w.Header().Get("HX-Redirect") != rutaDeSprint(proyecto, sprintID) || w.Body.Len() != 0 {
		t.Errorf("con HTMX: %d, HX-Redirect=%q, cuerpo=%q; quiero 200, el detalle y cuerpo vacio", w.Code, w.Header().Get("HX-Redirect"), w.Body.String())
	}
	contiene(t, e.get(rutaDeSprint(proyecto, sprintID)), "cerrado")
}

// CA-T016-3, CL-T016-1: un sprint sin historias se cierra, y no dice que algo volvio al backlog.
func TestPantallasCierre_SprintSinHistorias(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	sprintID := e.crearSprint(t, proyecto, "Vacio")

	if w := e.post(rutaDeSprint(proyecto, sprintID)+"/cerrar", nil); w.Code != http.StatusSeeOther {
		t.Fatalf("cerrar un sprint vacio: %d %s", w.Code, w.Body.String())
	}

	cerrado := e.get(rutaDeSprint(proyecto, sprintID))
	contiene(t, cerrado, "cerrado", "El sprint todavía no tiene historias", "Planificados: 0")
	noContiene(t, cerrado, "Volvieron al backlog", "Cerrar sprint")
}

// CL-T016-2, CL-T016-3: si se completaron todas, no vuelve ninguna; si no se completo ninguna, vuelven
// todas y se cuentan en plural.
func TestPantallasCierre_TodasONinguna(t *testing.T) {
	e := nuevoEntorno(t)

	// Todas completadas.
	proyecto, sprintID, _, opcional := sprintConDosHistorias(t, e)
	e.post(rutaDeSprint(proyecto, sprintID)+"/historias/"+opcional+"/completar", nil)
	e.post(rutaDeSprint(proyecto, sprintID)+"/cerrar", nil)
	todas := e.get(rutaDeSprint(proyecto, sprintID))
	contiene(t, todas, "Completados: 8")
	noContiene(t, todas, "Volvieron al backlog", "No completada")

	// Ninguna completada.
	otro := e.crearProyecto(t, "Otro")
	a := e.historiaLista(t, otro, "Historia A", "2")
	b := e.historiaLista(t, otro, "Historia B", "3")
	sprint := e.crearSprint(t, otro, "Sin suerte")
	for _, id := range []string{a, b} {
		e.post(rutaDeSprint(otro, sprint)+"/historias", url.Values{"item": {id}})
	}
	e.post(rutaDeSprint(otro, sprint)+"/cerrar", nil)
	ninguna := e.get(rutaDeSprint(otro, sprint))
	contiene(t, ninguna, "Completados: 0", "Volvieron al backlog: 2 historias")
	if n := strings.Count(ninguna.Body.String(), "No completada · volvió al backlog"); n != 2 {
		t.Errorf("hay %d historias no completadas, quiero 2", n)
	}
}

// CA-T016-4, RN-T016-6, CL-T016-4, CL-T016-5: un segundo cierre responde 409 y no cambia nada; el sprint
// de otro proyecto, uno que no existe o con un ID mal formado responden 404.
func TestPantallasCierre_Rechazos(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto, sprintID, _, _ := sprintConDosHistorias(t, e)
	otroProyecto := e.crearProyecto(t, "Otro")
	cerrar := rutaDeSprint(proyecto, sprintID) + "/cerrar"
	if w := e.post(cerrar, nil); w.Code != http.StatusSeeOther {
		t.Fatalf("primer cierre: %d %s", w.Code, w.Body.String())
	}
	antes := e.get(rutaDeSprint(proyecto, sprintID)).Body.String()

	w := e.post(cerrar, nil)
	if w.Code != http.StatusConflict {
		t.Errorf("segundo cierre: %d, quiero 409", w.Code)
	}
	contiene(t, w, "el sprint esta cerrado")

	casos := []struct{ nombre, ruta string }{
		{"el sprint es de otro proyecto", rutaDeSprint(otroProyecto, sprintID) + "/cerrar"},
		{"el sprint no existe", rutaDeSprint(proyecto, "00000000-0000-0000-0000-000000000000") + "/cerrar"},
		{"el ID del sprint esta mal formado", rutaDeSprint(proyecto, "no-es-un-uuid") + "/cerrar"},
		{"el proyecto no existe", rutaDeSprint("00000000-0000-0000-0000-000000000000", sprintID) + "/cerrar"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if w := e.post(c.ruta, nil); w.Code != http.StatusNotFound {
				t.Errorf("POST %s: %d, quiero 404", c.ruta, w.Code)
			}
		})
	}

	if despues := e.get(rutaDeSprint(proyecto, sprintID)).Body.String(); despues != antes {
		t.Error("los rechazos cambiaron el detalle del sprint")
	}
}

// CL-T016-7, RN-T015-13 (DEF-007): una historia que no se completo en el sprint 1, volvio al backlog y se
// completo en el sprint 2 sigue "no completada" en el 1, con sus Story Points completados intactos.
func TestPantallasCierre_ElResultadoNoCambiaDespues(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto, primero, _, opcional := sprintConDosHistorias(t, e)
	e.post(rutaDeSprint(proyecto, primero)+"/cerrar", nil)
	w := e.post("/proyectos/"+proyecto+"/sprints", datosDelSegundoSprint("Segundo incremento"))
	segundo := idDeSprintEnLaRuta.FindStringSubmatch(w.Header().Get("Location"))[2]
	e.post(rutaDeSprint(proyecto, segundo)+"/historias", url.Values{"item": {opcional}})
	e.post(rutaDeSprint(proyecto, segundo)+"/historias/"+opcional+"/completar", nil)

	enElPrimero := e.get(rutaDeSprint(proyecto, primero))
	contiene(t, enElPrimero, "Completados: 5", "No completada · volvió al backlog", "Volvieron al backlog: 1 historia")

	enElSegundo := e.get(rutaDeSprint(proyecto, segundo))
	contiene(t, enElSegundo, "Completados: 3", "Desmarcar")
}
