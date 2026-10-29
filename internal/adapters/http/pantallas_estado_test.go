package httpadapter_test

import (
	"net/http"
	"net/url"
	"testing"
)

// Tarjeta de estado general del proyecto (spec US-004). El reloj de los tests es el 2026-10-20.

func proyectoConFechas(e entorno, t *testing.T, nombre, inicio, fin string) string {
	t.Helper()
	datos := datosDeProyecto(nombre)
	datos.Set("fecha_inicio", inicio)
	datos.Set("fecha_fin", fin)
	w := e.post("/proyectos", datos)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("crear el proyecto %q: %d %s", nombre, w.Code, w.Body.String())
	}
	return idEnLaRuta.FindStringSubmatch(w.Header().Get("Location"))[1]
}

// CA-004-1, CA-004-3, RN-004-9: la pagina del proyecto muestra la tarjeta arriba del resumen, con el estado y los dias, las
// historias por estado, los Story Points, el porcentaje y el sprint abierto con su enlace. No agrega una ruta.
func TestPantallasEstado_Tarjeta(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo") // del 01/10/2026 al 15/12/2026, hoy es el 20/10
	completada := e.historiaLista(t, proyecto, "Hecha", "5")
	e.historiaLista(t, proyecto, "En curso", "3")
	e.historiaLista(t, proyecto, "Pendiente", "2")
	sprintID := e.crearSprint(t, proyecto, "Primer incremento")
	for _, titulo := range []string{"Hecha", "En curso"} {
		if w := e.post(rutaDeSprint(proyecto, sprintID)+"/historias", url.Values{"item": {e.idDeHistoria(t, proyecto, titulo)}}); w.Code != http.StatusSeeOther {
			t.Fatalf("asignar %q: %d %s", titulo, w.Code, w.Body.String())
		}
	}
	if w := e.post(rutaDeSprint(proyecto, sprintID)+"/historias/"+completada+"/completar", nil); w.Code != http.StatusSeeOther {
		t.Fatalf("completar: %d %s", w.Code, w.Body.String())
	}

	w := e.get("/proyectos/" + proyecto)

	if w.Code != http.StatusOK {
		t.Fatalf("GET proyecto: %d", w.Code)
	}
	contiene(t, w, "Estado general", "En curso: faltan 56 días", "Pendientes: 2", "En progreso: 0", "Completadas: 1", "(de 3)",
		"Story Points: 5 completados de 10", "33,3 % de las historias completadas", "Sprint 1 en curso", rutaDeSprint(proyecto, sprintID))
	if posicion(t, w, "Estado general") > posicion(t, w, "Descripción") {
		t.Error("la tarjeta tiene que ir arriba del resumen del proyecto")
	}
	noContiene(t, w, "Sin sprint abierto")
}

// CA-004-2, CA-004-4, CL-004-1 a CL-004-3: el texto del estado para cada situacion, con los dias en singular o plural.
func TestPantallasEstado_Textos(t *testing.T) {
	e := nuevoEntorno(t) // hoy es el 20/10/2026
	casos := []struct {
		nombre, inicio, fin, texto string
	}{
		{"no iniciado: empieza en 12 dias", "2026-11-01", "2026-12-15", "No iniciado: empieza en 12 días"},
		{"no iniciado: empieza manana", "2026-10-21", "2026-12-15", "No iniciado: empieza en 1 día"},
		{"en curso: el primer dia", "2026-10-20", "2026-10-30", "En curso: faltan 10 días"},
		{"en curso: falta 1 dia", "2026-10-01", "2026-10-21", "En curso: falta 1 día"},
		{"en curso: termina hoy", "2026-10-01", "2026-10-20", "En curso: termina hoy"},
		{"un solo dia: termina hoy", "2026-10-20", "2026-10-20", "En curso: termina hoy"},
		{"finalizado: termino hace 3 dias", "2026-09-01", "2026-10-17", "Finalizado: terminó hace 3 días"},
		{"finalizado: termino ayer", "2026-09-01", "2026-10-19", "Finalizado: terminó hace 1 día"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			id := proyectoConFechas(e, t, c.nombre, c.inicio, c.fin)
			contiene(t, e.get("/proyectos/"+id), c.texto)
		})
	}
}

// CA-004-5, CL-004-7: un proyecto sin historias ni sprints muestra todo en 0, 0 % y «Sin sprint abierto», sin error.
func TestPantallasEstado_ProyectoVacio(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Vacio")

	w := e.get("/proyectos/" + proyecto)

	if w.Code != http.StatusOK {
		t.Fatalf("GET proyecto: %d", w.Code)
	}
	contiene(t, w, "Pendientes: 0", "En progreso: 0", "Completadas: 0", "(de 0)", "Story Points: 0 completados de 0", "0,0 % de las historias completadas", "Sin sprint abierto")
	noContiene(t, w, "NaN")
}

// CL-004-10: con todos los sprints cerrados no hay sprint abierto.
func TestPantallasEstado_SprintCerrado(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Cerrado")
	sprintID := e.crearSprint(t, proyecto, "Primero")
	if w := e.post(rutaDeSprint(proyecto, sprintID)+"/cerrar", nil); w.Code != http.StatusSeeOther {
		t.Fatalf("cerrar el sprint: %d %s", w.Code, w.Body.String())
	}

	contiene(t, e.get("/proyectos/"+proyecto), "Sin sprint abierto")
}

// CA-004-6, CL-004-11, RN-T011-10: un proyecto inexistente o con ID mal formado sigue siendo un 404 con layout.
func TestPantallasEstado_ProyectoInexistente(t *testing.T) {
	e := nuevoEntorno(t)
	for _, id := range []string{"no-existe", "00000000-0000-0000-0000-00000000dead"} {
		w := e.get("/proyectos/" + id)
		if w.Code != http.StatusNotFound {
			t.Errorf("GET /proyectos/%s: %d, se esperaba 404", id, w.Code)
		}
		contiene(t, w, "<!doctype html>")
		noContiene(t, w, "Estado general")
	}
}

// Con HTMX la pagina del proyecto se responde como fragmento y lleva la tarjeta; el HTML se escapa en el nombre del sprint.
func TestPantallasEstado_Fragmento(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Con HTMX")

	w := e.enviar(http.MethodGet, "/proyectos/"+proyecto, nil, true)

	contiene(t, w, "Estado general", "En curso: faltan 56 días")
	noContiene(t, w, "<html")
}
