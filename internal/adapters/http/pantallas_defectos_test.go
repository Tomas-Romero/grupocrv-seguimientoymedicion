package httpadapter_test

import (
	"context"
	"errors"
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
	noContiene(t, w, "Todavía no hay defectos")
	// La tabla muestra las etiquetas de la pantalla y no los valores del dominio. Se miran solo las celdas: los
	// formularios llevan esos valores en sus campos ocultos y en las opciones.
	tabla := celdas(w.Body.String()[strings.Index(w.Body.String(), "<table"):strings.Index(w.Body.String(), "</table>")])
	for _, valor := range []string{"critica", "en_correccion", "descartado"} {
		if strings.Contains(tabla, valor) {
			t.Errorf("la tabla no tiene que mostrar el valor del dominio %q:\n%s", valor, tabla)
		}
	}

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

// escenarioDeDefectos es un proyecto con una historia y un sprint, lo minimo para registrar un defecto.
type escenarioDeDefectos struct {
	proyecto, historia, sprint string
}

func (e entorno) proyectoParaDefectos(t *testing.T, nombre string) escenarioDeDefectos {
	t.Helper()
	proyecto := e.crearProyecto(t, nombre)
	e.crearHistoria(t, proyecto, "Historia uno", "must", "Un criterio")
	return escenarioDeDefectos{proyecto: proyecto, historia: e.idDeHistoria(t, proyecto, "Historia uno"), sprint: e.crearSprint(t, proyecto, "Meta")}
}

// datos arma el formulario de registrar un defecto con la historia y el sprint del escenario.
func (d escenarioDeDefectos) datos(descripcion, severidad string) url.Values {
	return url.Values{"descripcion": {descripcion}, "severidad": {severidad}, "historia": {d.historia}, "sprint_deteccion": {d.sprint}}
}

var idDeDefectoEnLaRuta = regexp.MustCompile(`^/proyectos/([^/]+)/defectos$`)

// T-021 / CA-T021-1: el recorrido funciona de punta a punta con el router real y la base en memoria:
// un formulario registra un defecto con descripcion de varias lineas, severidad, historia y sprint de
// deteccion, vuelve a la lista y el defecto aparece abierto y sin sprint de resolucion.
func TestPantallasDefectos_Registrar(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")

	formulario := e.get(rutaDeDefectos(d.proyecto))
	contiene(t, formulario, `hx-post="`+rutaDeDefectos(d.proyecto)+`"`, `<textarea`, `name="descripcion"`, `name="severidad"`,
		`name="historia"`, `name="sprint_deteccion"`, "Registrar defecto")

	descripcion := "Esperado: guarda\nObservado: falla\nPasos: abrir y guardar"
	w := e.post(rutaDeDefectos(d.proyecto), d.datos(descripcion, "alta"))
	if w.Code != http.StatusSeeOther || w.Header().Get("Location") != rutaDeDefectos(d.proyecto) {
		t.Fatalf("registrar: %d %q %s, quiero 303 a la lista", w.Code, w.Header().Get("Location"), w.Body.String())
	}

	lista := e.get(rutaDeDefectos(d.proyecto))
	contiene(t, lista, descripcion, "Alta", "Abierto", "#1 Historia uno", "Sprint 1", "—")
	noContiene(t, lista, "Todavía no hay defectos")
	guardados := e.base.Defectos(d.proyecto)
	if len(guardados) != 1 || guardados[0].Descripcion != descripcion || guardados[0].Severidad != defecto.SeveridadAlta ||
		guardados[0].ItemID != d.historia || guardados[0].SprintDeteccionID != d.sprint {
		t.Errorf("defectos guardados = %+v, quiero uno con los datos del formulario", guardados)
	}
}

// T-021 / CA-T021-1: con HTMX el alta responde HX-Redirect con el cuerpo vacio; sin HTMX, 303.
func TestPantallasDefectos_RegistrarConYSinHTMX(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")

	w := e.enviar(http.MethodPost, rutaDeDefectos(d.proyecto), d.datos("Con HTMX", "baja"), true)
	if w.Code != http.StatusOK || w.Header().Get("HX-Redirect") != rutaDeDefectos(d.proyecto) || w.Body.Len() != 0 {
		t.Errorf("con HTMX: %d, HX-Redirect=%q, cuerpo=%q; quiero 200, la lista y cuerpo vacio", w.Code, w.Header().Get("HX-Redirect"), w.Body.String())
	}
	w = e.enviar(http.MethodPost, rutaDeDefectos(d.proyecto), d.datos("Sin HTMX", "baja"), false)
	if w.Code != http.StatusSeeOther || !idDeDefectoEnLaRuta.MatchString(w.Header().Get("Location")) {
		t.Errorf("sin HTMX: %d, Location=%q; quiero 303 a la lista", w.Code, w.Header().Get("Location"))
	}
}

// T-021 / RN-T021-15, RN-T021-16, CA-T021-4: el formulario ofrece las cuatro severidades con su etiqueta, las
// historias en el orden del backlog y TODOS los sprints del proyecto, abiertos y cerrados, por numero
// descendente. Cada lista empieza con una opcion sin valor.
func TestPantallasDefectos_ListasDelFormulario(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")
	e.crearHistoria(t, d.proyecto, "Historia opcional", "could", "")
	e.crearHistoria(t, d.proyecto, "Historia importante", "must", "")

	// Se cierra el Sprint 1 y se abre el 2: la lista tiene que ofrecer los dos.
	if w := e.post(rutaDeSprint(d.proyecto, d.sprint)+"/cerrar", nil); w.Code != http.StatusSeeOther {
		t.Fatalf("cerrar el sprint: %d %s", w.Code, w.Body.String())
	}
	segundo := url.Values{"sprint_goal": {"Otra meta"}, "fecha_inicio": {"2026-10-19"}, "fecha_fin": {"2026-11-01"}}
	if w := e.post("/proyectos/"+d.proyecto+"/sprints", segundo); w.Code != http.StatusSeeOther {
		t.Fatalf("crear el segundo sprint: %d %s", w.Code, w.Body.String())
	}

	w := e.get(rutaDeDefectos(d.proyecto))
	contiene(t, w, `<option value="">Elegí una severidad</option>`, `<option value="critica">Crítica</option>`, `<option value="alta">Alta</option>`,
		`<option value="media">Media</option>`, `<option value="baja">Baja</option>`,
		`<option value="">Elegí una historia</option>`, `<option value="">Elegí un sprint</option>`,
		"#1 Historia uno", "Sprint 1", "Sprint 2")
	// Historias por prioridad (must antes que could) y sprints del mas nuevo al mas viejo.
	if !(posicion(t, w, "#1 Historia uno") < posicion(t, w, "#3 Historia importante")) || !(posicion(t, w, "#3 Historia importante") < posicion(t, w, "#2 Historia opcional")) {
		t.Errorf("las historias tienen que salir en el orden del backlog:\n%s", w.Body.String())
	}
	if posicion(t, w, `>Sprint 2<`) > posicion(t, w, `>Sprint 1<`) {
		t.Errorf("los sprints tienen que salir por numero descendente:\n%s", w.Body.String())
	}
}

// T-021 / CL-T021-2, CA-T021-4: sin historias o sin sprints el formulario muestra la lista vacia (solo la
// opcion inicial) y un aviso.
func TestPantallasDefectos_SinHistoriasOSinSprints(t *testing.T) {
	e := nuevoEntorno(t)
	vacio := e.crearProyecto(t, "Vacio")

	w := e.get(rutaDeDefectos(vacio))
	contiene(t, w, "Primero cargá una historia", "Primero creá un sprint", `<option value="">Elegí una historia</option>`, `<option value="">Elegí un sprint</option>`)

	completo := e.proyectoParaDefectos(t, "Completo")
	noContiene(t, e.get(rutaDeDefectos(completo.proyecto)), "Primero cargá una historia", "Primero creá un sprint")
}

// T-021 / RN-T021-17: cada campo del formulario tiene su <label>.
func TestPantallasDefectos_FormularioAccesible(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")

	w := e.get(rutaDeDefectos(d.proyecto))
	for _, campo := range []string{"descripcion", "severidad", "historia", "sprint_deteccion"} {
		if !regexp.MustCompile(`<label[^>]*for="` + campo + `"`).MatchString(w.Body.String()) {
			t.Errorf("el campo %s no tiene <label>", campo)
		}
	}
}

// T-021 / CA-T021-5, RN-T021-6, RN-T021-8, CL-T021-3, CL-T021-4, CL-T021-5: los errores de descripcion y severidad
// los decide el dominio de US-022; salen con el campo en data-campo (juntos y en ese orden si hay dos), en
// 422, y no registran nada. De 2000 caracteres se acepta y de 2001 se rechaza.
func TestPantallasDefectos_ErroresAlRegistrar(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")

	casos := []struct {
		nombre      string
		descripcion string
		severidad   string
		campos      []string // en el orden en que tienen que aparecer
	}{
		{"descripcion vacia", "", "alta", []string{"descripcion"}},
		{"descripcion con solo espacios", " \n\t ", "alta", []string{"descripcion"}},
		{"descripcion de 2001 caracteres", strings.Repeat("a", 2001), "alta", []string{"descripcion"}},
		{"severidad sin elegir", "Falla", "", []string{"severidad"}},
		{"severidad mandada a mano", "Falla", "grave", []string{"severidad"}},
		{"descripcion vacia y severidad sin elegir", "", "", []string{"descripcion", "severidad"}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := e.post(rutaDeDefectos(d.proyecto), d.datos(c.descripcion, c.severidad))
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("estado %d, quiero 422:\n%s", w.Code, w.Body.String())
			}
			anterior := -1
			for _, campo := range c.campos {
				i := strings.Index(w.Body.String(), `data-campo="`+campo+`"`)
				if i < 0 || i < anterior {
					t.Fatalf("el error del campo %s falta o esta fuera de orden:\n%s", campo, w.Body.String())
				}
				anterior = i
			}
			if n := strings.Count(w.Body.String(), "data-campo="); n != len(c.campos) {
				t.Errorf("hay %d errores, quiero %d:\n%s", n, len(c.campos), w.Body.String())
			}
			if guardados := e.base.Defectos(d.proyecto); len(guardados) != 0 {
				t.Errorf("un alta rechazada no puede guardar nada, hay %+v", guardados)
			}
		})
	}

	contiene(t, e.post(rutaDeDefectos(d.proyecto), d.datos("", "alta")), "la descripcion es obligatoria")
	contiene(t, e.post(rutaDeDefectos(d.proyecto), d.datos(strings.Repeat("a", 2001), "alta")), "no puede tener mas de 2000 caracteres")

	// El limite: 2000 caracteres se acepta.
	if w := e.post(rutaDeDefectos(d.proyecto), d.datos(strings.Repeat("a", 2000), "alta")); w.Code != http.StatusSeeOther {
		t.Errorf("descripcion de 2000 caracteres: %d, quiero 303", w.Code)
	}
}

// T-021 / RN-T021-5, CL-T021-2, CL-T021-5: historia y sprint de deteccion son listas obligatorias: sin elegir se
// informan en su campo, juntos si faltan los dos, y no se llama al caso de uso: los errores de
// descripcion y severidad no se informan hasta que los dos estan elegidos.
func TestPantallasDefectos_HistoriaYSprintSinElegir(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")

	sinHistoria := d.datos("", "")
	sinHistoria.Set("historia", "")
	w := e.post(rutaDeDefectos(d.proyecto), sinHistoria)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("sin historia: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="historia"`, "Elegí la historia relacionada")
	noContiene(t, w, `data-campo="sprint_deteccion"`, `data-campo="descripcion"`, `data-campo="severidad"`)

	sinSprint := d.datos("Falla", "alta")
	sinSprint.Set("sprint_deteccion", "")
	w = e.post(rutaDeDefectos(d.proyecto), sinSprint)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("sin sprint: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="sprint_deteccion"`, "Elegí el sprint de detección")
	noContiene(t, w, `data-campo="historia"`)

	ninguno := url.Values{"descripcion": {""}, "severidad": {""}}
	w = e.post(rutaDeDefectos(d.proyecto), ninguno)
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("sin historia ni sprint: %d, quiero 422", w.Code)
	}
	if !(posicion(t, w, `data-campo="historia"`) < posicion(t, w, `data-campo="sprint_deteccion"`)) {
		t.Errorf("los dos errores tienen que salir juntos y en ese orden:\n%s", w.Body.String())
	}
	noContiene(t, w, `data-campo="descripcion"`, `data-campo="severidad"`)

	if guardados := e.base.Defectos(d.proyecto); len(guardados) != 0 {
		t.Errorf("un alta rechazada no puede guardar nada, hay %+v", guardados)
	}
}

// T-021 / RN-T021-10, RN-T021-12, CL-T021-15: una historia o un sprint de otro proyecto, o que no existen, o un
// proyecto que no existe, responden 404 aunque existan por separado, y no registran nada. El 404 del
// proyecto gana sobre el aviso de una lista sin elegir.
func TestPantallasDefectos_RutasQueNoCorresponden(t *testing.T) {
	e := nuevoEntorno(t)
	a := e.proyectoParaDefectos(t, "Proyecto A")
	b := e.proyectoParaDefectos(t, "Proyecto B")

	deOtraHistoria := a.datos("Falla", "alta")
	deOtraHistoria.Set("historia", b.historia)
	deOtroSprint := a.datos("Falla", "alta")
	deOtroSprint.Set("sprint_deteccion", b.sprint)
	inexistente := a.datos("Falla", "alta")
	inexistente.Set("historia", "00000000-0000-0000-0000-000000000000")
	malFormado := a.datos("Falla", "alta")
	malFormado.Set("sprint_deteccion", "no-es-un-uuid")
	// Con datos que de otro modo se rechazarian por validacion, el 404 tiene prioridad (RN-022-5).
	invalidoYDeOtro := b.datos("", "")
	invalidoYDeOtro.Set("historia", a.historia)

	casos := []struct {
		nombre   string
		proyecto string
		datos    url.Values
	}{
		{"historia de otro proyecto", a.proyecto, deOtraHistoria},
		{"sprint de otro proyecto", a.proyecto, deOtroSprint},
		{"historia inexistente", a.proyecto, inexistente},
		{"sprint mal formado", a.proyecto, malFormado},
		{"404 antes que los errores de validacion", b.proyecto, invalidoYDeOtro},
		{"proyecto inexistente", "00000000-0000-0000-0000-000000000000", a.datos("Falla", "alta")},
		{"proyecto mal formado", "no-es-un-uuid", a.datos("Falla", "alta")},
		{"proyecto inexistente con historia sin elegir", "00000000-0000-0000-0000-000000000000", url.Values{"descripcion": {"Falla"}, "severidad": {"alta"}}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if w := e.post(rutaDeDefectos(c.proyecto), c.datos); w.Code != http.StatusNotFound {
				t.Errorf("estado %d, quiero 404:\n%s", w.Code, w.Body.String())
			}
		})
	}
	if len(e.base.Defectos(a.proyecto))+len(e.base.Defectos(b.proyecto)) != 0 {
		t.Error("un alta rechazada no puede guardar nada")
	}
}

// T-021 / CL-T021-6, CL-T021-7, CA-T021-4: el sprint de deteccion puede ser uno cerrado, y dos defectos con los mismos
// datos se registran los dos y aparecen en dos filas.
func TestPantallasDefectos_SprintCerradoYRepetidos(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")
	if w := e.post(rutaDeSprint(d.proyecto, d.sprint)+"/cerrar", nil); w.Code != http.StatusSeeOther {
		t.Fatalf("cerrar el sprint: %d %s", w.Code, w.Body.String())
	}

	for i := 0; i < 2; i++ {
		if w := e.post(rutaDeDefectos(d.proyecto), d.datos("Falla repetida", "media")); w.Code != http.StatusSeeOther {
			t.Fatalf("registrar en un sprint cerrado: %d %s", w.Code, w.Body.String())
		}
	}
	lista := e.get(rutaDeDefectos(d.proyecto))
	if n := strings.Count(lista.Body.String(), ">Falla repetida</td>"); n != 2 {
		t.Errorf("la lista tiene %d filas con la descripcion repetida, quiero 2", n)
	}
	contiene(t, lista, "Sprint 1")
}

func rutaDeEstado(proyectoID, defectoID string) string {
	return rutaDeDefectos(proyectoID) + "/" + defectoID + "/estado"
}

// cambiar manda el formulario de un boton de la fila: el estado pedido y, si lo lleva, el sprint de resolucion.
func (e entorno) cambiar(proyectoID, defectoID, estado, sprintResolucion string) *httptest.ResponseRecorder {
	return e.post(rutaDeEstado(proyectoID, defectoID), url.Values{"estado": {estado}, "sprint_resolucion": {sprintResolucion}})
}

// segundoSprint cierra el Sprint 1 del escenario y abre el Sprint 2 (que empieza despues de que termina el 1).
func (e entorno) segundoSprint(t *testing.T, d escenarioDeDefectos) string {
	t.Helper()
	if w := e.post(rutaDeSprint(d.proyecto, d.sprint)+"/cerrar", nil); w.Code != http.StatusSeeOther {
		t.Fatalf("cerrar el sprint: %d %s", w.Code, w.Body.String())
	}
	segundo := url.Values{"sprint_goal": {"Otra meta"}, "fecha_inicio": {"2026-10-19"}, "fecha_fin": {"2026-11-01"}}
	w := e.post("/proyectos/"+d.proyecto+"/sprints", segundo)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("crear el segundo sprint: %d %s", w.Code, w.Body.String())
	}
	return idDeSprintEnLaRuta.FindStringSubmatch(w.Header().Get("Location"))[2]
}

// forzar deja el defecto en ese estado directamente en la base, como si otra persona lo hubiera cambiado
// desde otra pestana: la pagina que se abrio antes queda desactualizada.
func (e entorno) forzar(t *testing.T, d defecto.Defecto, estado defecto.Estado, sprintResolucion string) defecto.Defecto {
	t.Helper()
	anterior := d.Estado
	d.Estado, d.SprintResolucionID = estado, sprintResolucion
	if err := e.base.ActualizarEstadoDefecto(context.Background(), d, anterior); err != nil {
		t.Fatalf("ActualizarEstadoDefecto: %v", err)
	}
	return d
}

// fila devuelve el <tr> de la lista que contiene el texto, que tiene que ser uno solo.
func fila(t *testing.T, w *httptest.ResponseRecorder, texto string) string {
	t.Helper()
	var encontradas []string
	for _, tr := range strings.Split(w.Body.String(), "<tr")[1:] {
		if fin := strings.Index(tr, "</tr>"); fin >= 0 {
			tr = tr[:fin]
		}
		if strings.Contains(tr, texto) {
			encontradas = append(encontradas, tr)
		}
	}
	if len(encontradas) != 1 {
		t.Fatalf("hay %d filas con %q, quiero una:\n%s", len(encontradas), texto, w.Body.String())
	}
	return encontradas[0]
}

var formularios = regexp.MustCompile(`(?s)<form.*?</form>`)

// celdas es la fila sin los formularios de sus botones: lo que la fila muestra como texto. Las etiquetas de los
// botones y las opciones de las listas repiten palabras como "Abierto" o "Sprint 2" que no son el estado ni el
// sprint de resolucion del defecto.
func celdas(fila string) string { return formularios.ReplaceAllString(fila, "") }

// defectoEnLaLista es el estado y el sprint de resolucion que muestra la fila del defecto.
func defectoEnLaLista(t *testing.T, e entorno, proyecto, descripcion string) string {
	t.Helper()
	return fila(t, e.get(rutaDeDefectos(proyecto)), descripcion)
}

// T-021 / RN-T021-14, CA-T021-3: cada fila ofrece solo las transiciones validas desde el estado que muestra, y
// cada boton lleva el estado pedido en un campo oculto. Solo "Resolver" lleva la lista del sprint de resolucion.
func TestPantallasDefectos_BotonesSegunElEstado(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")
	e.segundoSprint(t, d)
	abierto := e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Fila abierta", defecto.SeveridadAlta)
	enCorreccion := e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Fila en correccion", defecto.SeveridadAlta)
	resuelto := e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Fila resuelta", defecto.SeveridadAlta)
	descartado := e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Fila descartada", defecto.SeveridadAlta)
	e.forzar(t, enCorreccion, defecto.EstadoEnCorreccion, "")
	e.forzar(t, resuelto, defecto.EstadoResuelto, e.base.Sprints(d.proyecto)[1].ID)
	e.forzar(t, descartado, defecto.EstadoDescartado, "")

	w := e.get(rutaDeDefectos(d.proyecto))
	casos := []struct {
		descripcion string
		defecto     defecto.Defecto
		botones     []string
		estados     []string // valor del campo oculto de cada boton
		conSprint   bool
	}{
		{"Fila abierta", abierto, []string{"En corrección", "Resolver", "Descartar"}, []string{"en_correccion", "resuelto", "descartado"}, true},
		{"Fila en correccion", enCorreccion, []string{"Volver a abierto", "Resolver", "Descartar"}, []string{"abierto", "resuelto", "descartado"}, true},
		{"Fila resuelta", resuelto, []string{"Reabrir"}, []string{"abierto"}, false},
		{"Fila descartada", descartado, []string{"Reabrir"}, []string{"abierto"}, false},
	}
	todos := []string{"En corrección", "Volver a abierto", "Resolver", "Descartar", "Reabrir"}
	for _, c := range casos {
		t.Run(c.descripcion, func(t *testing.T) {
			f := fila(t, w, c.descripcion)
			for _, boton := range todos {
				tiene := strings.Contains(f, `aria-label="`+boton+`: `+c.descripcion)
				quiere := false
				for _, b := range c.botones {
					quiere = quiere || b == boton
				}
				if tiene != quiere {
					t.Errorf("boton %q: presente=%v, quiero %v:\n%s", boton, tiene, quiere, f)
				}
			}
			for _, estado := range c.estados {
				if !strings.Contains(f, `name="estado" value="`+estado+`"`) {
					t.Errorf("falta el campo oculto estado=%s:\n%s", estado, f)
				}
			}
			if n := strings.Count(f, `hx-post="`+rutaDeEstado(d.proyecto, c.defecto.ID)+`"`); n != len(c.botones) {
				t.Errorf("hay %d formularios hacia la ruta del cambio de estado, quiero %d:\n%s", n, len(c.botones), f)
			}
			if tiene := strings.Contains(f, `name="sprint_resolucion"`); tiene != c.conSprint {
				t.Errorf("lista de sprint de resolucion presente=%v, quiero %v", tiene, c.conSprint)
			}
		})
	}
	// La lista de resolucion ofrece todos los sprints del proyecto, sin filtrar (RN-T021-15).
	contiene(t, w, `<option value="">Elegí un sprint</option>`)
	if n := strings.Count(fila(t, w, "Fila abierta"), "<option"); n != 3 {
		t.Errorf("la lista de resolucion tiene %d opciones, quiero 3 (la inicial y los dos sprints)", n)
	}
}

// T-021 / CA-T021-3, CL-T021-9, CL-T021-11: un defecto se pasa a en correccion, se resuelve con su sprint de
// resolucion, se descarta y se reabre; cada cambio responde 303 sin HTMX y HX-Redirect con HTMX hacia la lista, y
// la lista muestra el estado y el sprint de resolucion nuevos.
func TestPantallasDefectos_CambiarElEstado(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")
	segundo := e.segundoSprint(t, d)
	defecto1 := e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Falla al guardar", defecto.SeveridadAlta)

	pasos := []struct {
		nombre, estado, sprint string
		quiero                 []string
		noQuiero               []string
	}{
		{"pasar a en correccion", "en_correccion", "", []string{">En corrección</td>", ">—</td>", `aria-label="Volver a abierto: Falla al guardar"`}, []string{">Sprint 2</td>"}},
		{"resolver con el Sprint 2", "resuelto", segundo, []string{">Resuelto</td>", ">Sprint 2</td>", `aria-label="Reabrir: Falla al guardar"`}, []string{`aria-label="Resolver: Falla al guardar"`, `aria-label="Descartar: Falla al guardar"`}},
		{"reabrir", "abierto", "", []string{">Abierto</td>", ">—</td>", `aria-label="Resolver: Falla al guardar"`}, []string{">Resuelto</td>", `aria-label="Reabrir: Falla al guardar"`}},
		{"descartar", "descartado", "", []string{">Descartado</td>", ">—</td>", `aria-label="Reabrir: Falla al guardar"`}, []string{`aria-label="Descartar: Falla al guardar"`}},
		{"reabrir el descartado", "abierto", "", []string{">Abierto</td>", ">—</td>"}, []string{">Descartado</td>"}},
		{"volver a abierto desde en correccion", "en_correccion", "", []string{">En corrección</td>"}, nil},
		{"y volver a abierto", "abierto", "", []string{">Abierto</td>"}, []string{">En corrección</td>"}},
	}
	for _, p := range pasos {
		t.Run(p.nombre, func(t *testing.T) {
			w := e.cambiar(d.proyecto, defecto1.ID, p.estado, p.sprint)
			if w.Code != http.StatusSeeOther || w.Header().Get("Location") != rutaDeDefectos(d.proyecto) {
				t.Fatalf("estado %d, Location=%q, cuerpo=%s; quiero 303 a la lista", w.Code, w.Header().Get("Location"), w.Body.String())
			}
			f := defectoEnLaLista(t, e, d.proyecto, "Falla al guardar")
			for _, q := range p.quiero {
				if !strings.Contains(f, q) {
					t.Errorf("la fila no tiene %q:\n%s", q, f)
				}
			}
			for _, q := range p.noQuiero {
				if strings.Contains(f, q) {
					t.Errorf("la fila no deberia tener %q:\n%s", q, f)
				}
			}
		})
	}

	// Con HTMX responde HX-Redirect y el cuerpo vacio.
	w := e.enviar(http.MethodPost, rutaDeEstado(d.proyecto, defecto1.ID), url.Values{"estado": {"en_correccion"}}, true)
	if w.Code != http.StatusOK || w.Header().Get("HX-Redirect") != rutaDeDefectos(d.proyecto) || w.Body.Len() != 0 {
		t.Errorf("con HTMX: %d, HX-Redirect=%q, cuerpo=%q; quiero 200, la lista y cuerpo vacio", w.Code, w.Header().Get("HX-Redirect"), w.Body.String())
	}
}

// T-021 / CL-T021-9, CL-T021-11: resolver sin elegir sprint es ErrSprintResolucionFaltante en su campo; resolver en
// un sprint cerrado o posterior al de deteccion se acepta; y resolver de nuevo con otro sprint deja el nuevo.
func TestPantallasDefectos_Resolver(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")
	segundo := e.segundoSprint(t, d)
	falla := e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Falla", defecto.SeveridadAlta)

	w := e.cambiar(d.proyecto, falla.ID, "resuelto", "")
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("resolver sin sprint: %d, quiero 422", w.Code)
	}
	contiene(t, w, `data-campo="sprint_resolucion"`)
	contiene(t, e.get(rutaDeDefectos(d.proyecto)), ">Abierto</td>")

	// El Sprint 1 esta cerrado y es el de deteccion: se acepta (CL-023-5, CL-023-6).
	if w := e.cambiar(d.proyecto, falla.ID, "resuelto", d.sprint); w.Code != http.StatusSeeOther {
		t.Fatalf("resolver en el sprint cerrado: %d %s", w.Code, w.Body.String())
	}
	// "Sprint 1" es tambien el de deteccion: la fila lo muestra en las dos columnas.
	if f := celdas(defectoEnLaLista(t, e, d.proyecto, "Falla")); !strings.Contains(f, ">Resuelto</td>") || strings.Count(f, ">Sprint 1</td>") != 2 {
		t.Errorf("la fila tiene que mostrar Resuelto en el Sprint 1:\n%s", f)
	}

	// Reabrir lo deja abierto y sin sprint de resolucion; resolverlo de nuevo con otro sprint deja el nuevo.
	e.cambiar(d.proyecto, falla.ID, "abierto", "")
	if f := celdas(defectoEnLaLista(t, e, d.proyecto, "Falla")); !strings.Contains(f, ">Abierto</td>") || strings.Contains(f, ">Resuelto</td>") || !strings.Contains(f, ">—</td>") {
		t.Errorf("reabrir tiene que dejar el defecto abierto y sin sprint de resolucion:\n%s", f)
	}
	if w := e.cambiar(d.proyecto, falla.ID, "resuelto", segundo); w.Code != http.StatusSeeOther {
		t.Fatalf("resolver con otro sprint: %d %s", w.Code, w.Body.String())
	}
	if f := celdas(defectoEnLaLista(t, e, d.proyecto, "Falla")); !strings.Contains(f, ">Sprint 2</td>") {
		t.Errorf("la fila tiene que mostrar el nuevo sprint de resolucion:\n%s", f)
	}
}

// T-021 / CL-T021-10, CA-T021-5: resolver con un sprint anterior al de deteccion (la lista lo ofrece) es
// ErrSprintResolucionAnterior en el campo sprint_resolucion, con 409, y el defecto no cambia.
func TestPantallasDefectos_ResolverEnUnSprintAnterior(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")
	segundo := e.segundoSprint(t, d)
	deteccionEnElDos := d.datos("Detectado en el Sprint 2", "alta")
	deteccionEnElDos.Set("sprint_deteccion", segundo)
	if w := e.post(rutaDeDefectos(d.proyecto), deteccionEnElDos); w.Code != http.StatusSeeOther {
		t.Fatalf("registrar: %d %s", w.Code, w.Body.String())
	}
	guardado := e.base.Defectos(d.proyecto)[0]

	w := e.cambiar(d.proyecto, guardado.ID, "resuelto", d.sprint)
	if w.Code != http.StatusConflict {
		t.Fatalf("resolver en un sprint anterior: %d, quiero 409", w.Code)
	}
	contiene(t, w, `data-campo="sprint_resolucion"`)
	if f := celdas(defectoEnLaLista(t, e, d.proyecto, "Detectado en el Sprint 2")); !strings.Contains(f, ">Abierto</td>") {
		t.Errorf("el defecto no tiene que cambiar:\n%s", f)
	}
}

// T-021 / CL-T021-12, CA-T021-5: con la pagina desactualizada se evalua el pedido contra el estado guardado:
// descartar uno que otra persona ya resolvio, o volver a abierto uno que ya esta abierto, es
// ErrTransicionInvalida en el campo estado, con 409 y sin cambiar nada; reabrir el resuelto es valido.
func TestPantallasDefectos_PaginaDesactualizada(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")
	falla := e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Falla", defecto.SeveridadAlta)

	// La fila mostraba "Abierto", pero otra persona ya lo resolvio.
	e.forzar(t, falla, defecto.EstadoResuelto, d.sprint)
	w := e.cambiar(d.proyecto, falla.ID, "descartado", "")
	if w.Code != http.StatusConflict {
		t.Fatalf("descartar uno resuelto: %d, quiero 409", w.Code)
	}
	contiene(t, w, `data-campo="estado"`)
	if f := celdas(defectoEnLaLista(t, e, d.proyecto, "Falla")); !strings.Contains(f, ">Resuelto</td>") {
		t.Errorf("el defecto no tiene que cambiar:\n%s", f)
	}

	// Reabrir el resuelto si es una reapertura valida.
	if w := e.cambiar(d.proyecto, falla.ID, "abierto", ""); w.Code != http.StatusSeeOther {
		t.Errorf("reabrir el resuelto: %d, quiero 303", w.Code)
	}
	// Y "Volver a abierto" sobre uno que ya esta abierto no es una transicion.
	w = e.cambiar(d.proyecto, falla.ID, "abierto", "")
	if w.Code != http.StatusConflict {
		t.Errorf("volver a abierto uno abierto: %d, quiero 409", w.Code)
	}
	contiene(t, w, `data-campo="estado"`)
}

// T-021 / CL-T021-13, CA-T021-5: un estado a mano fuera de los cuatro es ErrEstadoInvalido (422) en el campo estado, y
// un sprint de resolucion con un estado distinto de resuelto es ErrSprintResolucionNoCorresponde (422) en su
// campo; el defecto no cambia.
func TestPantallasDefectos_EstadoMandadoAMano(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")
	falla := e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Falla", defecto.SeveridadAlta)

	casos := []struct {
		nombre, estado, sprint, campo string
	}{
		{"estado fuera de los cuatro", "cerrado", "", "estado"},
		{"estado vacio", "", "", "estado"},
		{"sprint con un estado que no es resuelto", "descartado", d.sprint, "sprint_resolucion"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := e.cambiar(d.proyecto, falla.ID, c.estado, c.sprint)
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("estado %d, quiero 422:\n%s", w.Code, w.Body.String())
			}
			contiene(t, w, `data-campo="`+c.campo+`"`)
			if f := celdas(defectoEnLaLista(t, e, d.proyecto, "Falla")); !strings.Contains(f, ">Abierto</td>") {
				t.Errorf("el defecto no tiene que cambiar:\n%s", f)
			}
		})
	}
}

// T-021 / RN-T021-11, RN-T021-12, RN-T021-10, CL-T021-15: un defecto solo se alcanza por la ruta de su proyecto, y un sprint
// de resolucion tiene que ser de ese proyecto: lo de otro proyecto, inexistente o mal formado responde 404 y no
// cambia nada.
func TestPantallasDefectos_CambioQueNoCorresponde(t *testing.T) {
	e := nuevoEntorno(t)
	a := e.proyectoParaDefectos(t, "Proyecto A")
	b := e.proyectoParaDefectos(t, "Proyecto B")
	deA := e.defectoGuardado(t, a.proyecto, a.historia, a.sprint, "Defecto de A", defecto.SeveridadAlta)

	// Control: por la ruta de su proyecto el defecto si se cambia. Sin esto, un 404 de cualquier ruta que no
	// existe pasaria por un 404 de "no corresponde".
	if w := e.cambiar(a.proyecto, deA.ID, "en_correccion", ""); w.Code != http.StatusSeeOther {
		t.Fatalf("cambiar por la ruta de su proyecto: %d %s, quiero 303", w.Code, w.Body.String())
	}

	casos := []struct {
		nombre, proyecto, defecto, sprint string
	}{
		{"defecto de otro proyecto", b.proyecto, deA.ID, ""},
		{"defecto inexistente", a.proyecto, "00000000-0000-0000-0000-000000000000", ""},
		{"defecto mal formado", a.proyecto, "no-es-un-uuid", ""},
		{"proyecto inexistente", "00000000-0000-0000-0000-000000000000", deA.ID, ""},
		{"proyecto mal formado", "no-es-un-uuid", deA.ID, ""},
		{"sprint de resolucion de otro proyecto", a.proyecto, deA.ID, b.sprint},
		{"sprint de resolucion inexistente", a.proyecto, deA.ID, "00000000-0000-0000-0000-000000000000"},
		{"sprint de resolucion mal formado", a.proyecto, deA.ID, "no-es-un-uuid"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			estado := "resuelto"
			if c.sprint == "" {
				estado = "descartado"
			}
			if w := e.cambiar(c.proyecto, c.defecto, estado, c.sprint); w.Code != http.StatusNotFound {
				t.Errorf("estado %d, quiero 404:\n%s", w.Code, w.Body.String())
			}
		})
	}
	if f := celdas(defectoEnLaLista(t, e, a.proyecto, "Defecto de A")); !strings.Contains(f, ">En corrección</td>") {
		t.Errorf("el defecto no tiene que cambiar despues del control:\n%s", f)
	}
}

// T-021 / RN-T021-17, RN-T021-9: cada control de la fila dice a que defecto corresponde con la accion y el comienzo de su
// descripcion (hasta 40 caracteres, sin saltos de linea), y el HTML de la descripcion se escapa tambien ahi.
func TestPantallasDefectos_NombresAccesibles(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Demo")
	e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Esperado: guarda\nObservado: falla al guardar el proyecto en la base", defecto.SeveridadAlta)
	e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, `<b>Corto</b>`, defecto.SeveridadBaja)

	w := e.get(rutaDeDefectos(d.proyecto))
	largo := "Esperado: guarda Observado: falla al gua"
	contiene(t, w, `aria-label="Resolver: `+largo+`"`, `aria-label="Descartar: `+largo+`"`, `aria-label="En corrección: `+largo+`"`,
		`aria-label="Sprint de resolución: `+largo+`"`)
	noContiene(t, w, largo+"r")
	contiene(t, w, `aria-label="Descartar: &lt;b&gt;Corto&lt;/b&gt;"`)
}

// baseQueCambiaAntesDeGuardar es la base en memoria con otra persona que cambia el defecto justo despues de que el
// caso de uso lo leyo y antes de que lo guarde: la condicion de carrera de CL-023-14.
type baseQueCambiaAntesDeGuardar struct {
	*memoria.Base
	antes func()
}

func (b baseQueCambiaAntesDeGuardar) ActualizarEstadoDefecto(ctx context.Context, d defecto.Defecto, anterior defecto.Estado) error {
	b.antes()
	return b.Base.ActualizarEstadoDefecto(ctx, d, anterior)
}

// repositorioDeDefectos es lo que necesitan los tres casos de uso de defectos de las pantallas.
type repositorioDeDefectos interface {
	app.RepositorioListarDefectos
	app.RepositorioDefectos
	app.RepositorioCambioDeEstadoDefecto
}

// entornoConBase arma el router real con los casos de uso de las pantallas de defectos sobre el repositorio
// indicado, para probar fallas que la base en memoria no produce sola. El proyecto, las historias y los
// sprints salen siempre de la base.
func entornoConBase(base *memoria.Base, defectos repositorioDeDefectos) entorno {
	casos := httpadapter.CasosDeUso{
		ConsultarProyecto:    app.NuevoConsultarProyecto(base),
		ListarBacklog:        app.NuevoListarBacklog(base),
		ListarSprints:        app.NuevoListarSprints(base),
		CrearItem:            app.NuevoCrearItemBacklog(base, time.Now),
		CrearSprint:          app.NuevoCrearSprint(base),
		CrearProyecto:        app.NuevoCrearProyecto(base),
		ListarDefectos:       app.NuevoListarDefectos(defectos),
		RegistrarDefecto:     app.NuevoRegistrarDefecto(defectos),
		CambiarEstadoDefecto: app.NuevoCambiarEstadoDefecto(defectos),
	}
	manejador := httpadapter.NuevoManejadorDeErrores(httpadapter.CatalogoDeErrores(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	router := httpadapter.NuevoRouter(verificadorFalso{}, time.Second, http.FS(fstest.MapFS{}),
		httpadapter.ConPantallas(httpadapter.NuevasPantallas(casos, manejador)))
	return entorno{router: router, base: base}
}

// T-021 / CL-T021-14: dos cambios simultaneos del mismo defecto: el segundo responde ErrDefectoModificado (409, sin
// campo) y no cambia nada; la pagina queda como estaba.
func TestPantallasDefectos_CambiosSimultaneos(t *testing.T) {
	base := memoria.NuevaBase()
	var otraPersona func()
	e := entornoConBase(base, baseQueCambiaAntesDeGuardar{Base: base, antes: func() { otraPersona() }})
	d := e.proyectoParaDefectos(t, "Demo")
	falla := e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Falla", defecto.SeveridadAlta)
	otraPersona = func() {
		otraPersona = func() {}
		e.forzar(t, falla, defecto.EstadoEnCorreccion, "")
	}

	w := e.cambiar(d.proyecto, falla.ID, "descartado", "")
	if w.Code != http.StatusConflict {
		t.Fatalf("estado %d, quiero 409:\n%s", w.Code, w.Body.String())
	}
	noContiene(t, w, "data-campo=")
	if guardados := base.Defectos(d.proyecto); len(guardados) != 1 || guardados[0].Estado != defecto.EstadoEnCorreccion {
		t.Errorf("el defecto tiene que quedar como lo dejo la otra persona (en correccion): %+v", guardados)
	}
}

// baseCaida es la base en memoria con los defectos caidos: leer, registrar y guardar un cambio de estado fallan con
// un error que no es de ninguna regla.
type baseCaida struct {
	*memoria.Base
	err error
}

func (b baseCaida) ListarDefectosDelProyecto(context.Context, string) ([]app.DefectoListado, error) {
	return nil, b.err
}

func (b baseCaida) RegistrarDefecto(context.Context, defecto.Defecto) (defecto.Defecto, error) {
	return defecto.Defecto{}, b.err
}

func (b baseCaida) ActualizarEstadoDefecto(context.Context, defecto.Defecto, defecto.Estado) error {
	return b.err
}

// T-021 / CL-T021-18: una falla inesperada de la base al listar, registrar o cambiar el estado responde 500 con el
// mensaje fijo de T-005, sin mostrar el detalle, y no registra ni cambia nada.
func TestPantallasDefectos_FallaInesperadaDeLaBase(t *testing.T) {
	base := memoria.NuevaBase()
	detalle := "conexion rechazada por 10.0.0.7"
	e := entornoConBase(base, baseCaida{Base: base, err: errors.New(detalle)})
	d := e.proyectoParaDefectos(t, "Demo")
	falla := e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Falla", defecto.SeveridadAlta)

	casos := []struct {
		nombre string
		pedir  func() *httptest.ResponseRecorder
	}{
		{"listar", func() *httptest.ResponseRecorder { return e.get(rutaDeDefectos(d.proyecto)) }},
		{"registrar", func() *httptest.ResponseRecorder {
			return e.post(rutaDeDefectos(d.proyecto), d.datos("Otra falla", "alta"))
		}},
		{"cambiar el estado", func() *httptest.ResponseRecorder { return e.cambiar(d.proyecto, falla.ID, "descartado", "") }},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := c.pedir()
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("estado %d, quiero 500:\n%s", w.Code, w.Body.String())
			}
			contiene(t, w, "ocurrio un error inesperado")
			noContiene(t, w, detalle)
		})
	}
	guardados := base.Defectos(d.proyecto)
	if len(guardados) != 1 || guardados[0].Estado != defecto.EstadoAbierto {
		t.Errorf("las fallas no pueden registrar ni cambiar nada: %+v", guardados)
	}
}
