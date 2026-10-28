package httpadapter_test

import (
	"context"
	"net/http"
	"net/url"
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
	noContiene(t, w, "Todavía no hay defectos")
	// La tabla muestra las etiquetas de la pantalla y no los valores del dominio. Se mira solo la tabla:
	// el formulario ofrece esos valores como value de sus opciones.
	tabla := w.Body.String()[strings.Index(w.Body.String(), "<table"):strings.Index(w.Body.String(), "</table>")]
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
	if n := strings.Count(lista.Body.String(), "Falla repetida"); n != 2 {
		t.Errorf("la lista tiene %d filas con la descripcion repetida, quiero 2", n)
	}
	contiene(t, lista, "Sprint 1")
}
