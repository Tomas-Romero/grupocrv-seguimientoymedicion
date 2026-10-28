package httpadapter_test

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

func rutaDeDefectos(proyectoID string) string { return "/proyectos/" + proyectoID + "/defectos" }

// defectoGuardado da de alta un defecto directamente en la base en memoria, sin pasar por la
// pantalla: sirve para probar la lista sin depender del formulario.
func (e entorno) defectoGuardado(t *testing.T, proyectoID, historiaID, sprintID, descripcion string, severidad defecto.Severidad) defecto.Defecto {
	t.Helper()
	d, err := defecto.Nuevo(proyectoID, historiaID, sprintID, descripcion, severidad)
	if err != nil {
		t.Fatalf("defecto.Nuevo(%q): %v", descripcion, err)
	}
	guardado, err := e.base.RegistrarDefecto(context.Background(), d)
	if err != nil {
		t.Fatalf("RegistrarDefecto(%q): %v", descripcion, err)
	}
	return guardado
}

// T-021 / CL-T021-1: sin defectos la pagina muestra el estado vacio, sin tabla y sin error.
func TestPantallasDefectos_ListaVacia(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")

	w := e.get(rutaDeDefectos(proyecto))
	if w.Code != http.StatusOK {
		t.Fatalf("GET defectos: %d, quiero 200", w.Code)
	}
	contiene(t, w, "Todavía no hay defectos")
	noContiene(t, w, "<table", `role="alert"`)
}

// T-021 / CA-T021-2, RN-T021-13, RN-T021-4: la lista muestra descripcion, severidad, estado, historia,
// sprint de deteccion y sprint de resolucion con las etiquetas de la pantalla, del mas reciente al
// mas antiguo, y no muestra la fecha de carga.
func TestPantallasDefectos_Lista(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	e.crearHistoria(t, proyecto, "Alta de proyectos", "must", "Un criterio")
	historia := e.idDeHistoria(t, proyecto, "Alta de proyectos")
	primero := e.crearSprint(t, proyecto, "Meta uno")

	viejo := e.defectoGuardado(t, proyecto, historia, primero, "Defecto viejo", defecto.SeveridadCritica)
	medio := e.defectoGuardado(t, proyecto, historia, primero, "Defecto medio", defecto.SeveridadMedia)
	e.defectoGuardado(t, proyecto, historia, primero, "Defecto nuevo", defecto.SeveridadBaja)

	// Se cambia el estado directo en la base: cambiarlo desde la pantalla es otro ciclo.
	viejo.Estado = defecto.EstadoDescartado
	medio.Estado = defecto.EstadoEnCorreccion
	for _, d := range []defecto.Defecto{viejo, medio} {
		if err := e.base.ActualizarEstadoDefecto(context.Background(), d, defecto.EstadoAbierto); err != nil {
			t.Fatalf("ActualizarEstadoDefecto: %v", err)
		}
	}

	w := e.get(rutaDeDefectos(proyecto))
	contiene(t, w, "Defecto viejo", "Defecto medio", "Defecto nuevo",
		"Crítica", "Media", "Baja", "Abierto", "En corrección", "Descartado",
		"#1 Alta de proyectos", "Sprint 1", "—")
	noContiene(t, w, "Todavía no hay defectos", "critica", "en_correccion")

	if !(posicion(t, w, "Defecto nuevo") < posicion(t, w, "Defecto medio") && posicion(t, w, "Defecto medio") < posicion(t, w, "Defecto viejo")) {
		t.Error("la lista tiene que mostrar primero el defecto mas reciente")
	}
	// La fecha de carga viaja en el modelo de lectura pero no se muestra (CA-T021-2).
	if regexp.MustCompile(`\d{2}/\d{2}/\d{4}|\d{4}-\d{2}-\d{2}`).MatchString(w.Body.String()) {
		t.Errorf("la lista no tiene que mostrar ninguna fecha:\n%s", w.Body.String())
	}
}

// T-021 / RN-T021-9, CL-T021-3: la descripcion se muestra completa, con sus saltos de linea.
func TestPantallasDefectos_DescripcionConSaltosDeLinea(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	e.crearHistoria(t, proyecto, "Historia", "must", "Un criterio")
	sprintID := e.crearSprint(t, proyecto, "Meta")
	e.defectoGuardado(t, proyecto, e.idDeHistoria(t, proyecto, "Historia"), sprintID,
		"Esperado: guarda\nObservado: falla\nPasos: abrir y guardar", defecto.SeveridadAlta)

	w := e.get(rutaDeDefectos(proyecto))
	contiene(t, w, "Esperado: guarda\nObservado: falla\nPasos: abrir y guardar", "whitespace-pre-wrap")
}

// T-021 / RN-T021-9, CL-T021-16: el HTML de la descripcion y del titulo de la historia se muestra
// escapado.
func TestPantallasDefectos_EscapaElTexto(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")
	e.crearHistoria(t, proyecto, `<b>Historia</b> & co`, "must", "Un criterio")
	sprintID := e.crearSprint(t, proyecto, "Meta")
	e.defectoGuardado(t, proyecto, e.idDeHistoria(t, proyecto, `<b>Historia</b> & co`), sprintID,
		`<script>alert("x")</script> ñandú "comillas"`, defecto.SeveridadAlta)

	w := e.get(rutaDeDefectos(proyecto))
	noContiene(t, w, "<script>alert", "<b>Historia")
	contiene(t, w, "&lt;script&gt;", "ñandú", "&lt;b&gt;Historia&lt;/b&gt; &amp; co")
}

// T-021 / RN-T021-18, CA-T021-6: la pagina usa el layout con el menu del proyecto y Defectos activo; con
// HTMX responde solo el fragmento.
func TestPantallasDefectos_NavegacionYFragmento(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto := e.crearProyecto(t, "Demo")

	w := e.get(rutaDeDefectos(proyecto))
	contiene(t, w, "<!doctype html>", `aria-label="Proyecto"`)
	if !regexp.MustCompile(`href="/proyectos/` + proyecto + `/defectos"[^>]*aria-current="page"`).MatchString(w.Body.String()) {
		t.Errorf("tiene que marcar Defectos como activo:\n%s", w.Body.String())
	}

	fragmento := e.enviar(http.MethodGet, rutaDeDefectos(proyecto), nil, true)
	contiene(t, fragmento, "Todavía no hay defectos")
	noContiene(t, fragmento, "<html")
}

// T-021 / RN-T021-10, CL-T021-15: un proyecto que no existe o con un ID mal formado responde 404 con
// la pagina del layout.
func TestPantallasDefectos_ProyectoInexistente(t *testing.T) {
	e := nuevoEntorno(t)

	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "no-es-un-uuid"} {
		w := e.get(rutaDeDefectos(id))
		if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), "<!doctype html>") {
			t.Errorf("GET defectos del proyecto %s: %d, quiero 404 con layout", id, w.Code)
		}
	}
}
