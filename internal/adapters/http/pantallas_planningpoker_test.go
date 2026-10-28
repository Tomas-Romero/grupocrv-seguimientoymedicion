package httpadapter_test

import (
	"net/http"
	"net/url"
	"regexp"
	"testing"
)

var idDeSesionEnLaRuta = regexp.MustCompile(`^/proyectos/([^/]+)/planning-poker/([^/]+)$`)

// mesaDePoker crea un proyecto con Ana y Beto y una historia sin estimar. Devuelve los IDs del
// proyecto y de la historia, y los de los integrantes por nombre.
func mesaDePoker(t *testing.T, e entorno) (proyecto, historia string, integrantes map[string]string) {
	t.Helper()
	proyecto = e.crearProyecto(t, "Demo")
	for _, datos := range []url.Values{
		{"nombre": {"Ana"}, "email": {"ana@ejemplo.test"}, "rol": {"product_architect"}},
		{"nombre": {"Beto"}, "email": {"beto@ejemplo.test"}, "rol": {"agile_enabler"}},
	} {
		if w := e.post("/proyectos/"+proyecto+"/integrantes", datos); w.Code != http.StatusSeeOther {
			t.Fatalf("registrar %s: %d %s", datos.Get("nombre"), w.Code, w.Body.String())
		}
	}
	e.crearHistoria(t, proyecto, "Alta de proyectos", "must", "Un criterio")
	historia = e.idDeHistoria(t, proyecto, "Alta de proyectos")
	return proyecto, historia, e.integrantesPorNombre(proyecto)
}

func (e entorno) integrantesPorNombre(proyecto string) map[string]string {
	ids := map[string]string{}
	for _, i := range e.base.Integrantes(proyecto) {
		ids[i.Nombre] = i.ID
	}
	return ids
}

// iniciarSesion inicia una sesion sobre la historia desde el backlog y devuelve la ruta de la
// pantalla de la sesion.
func (e entorno) iniciarSesion(t *testing.T, proyecto, historia string) string {
	t.Helper()
	w := e.post("/proyectos/"+proyecto+"/backlog/"+historia+"/planning-poker", nil)
	if w.Code != http.StatusSeeOther || idDeSesionEnLaRuta.FindStringSubmatch(w.Header().Get("Location")) == nil {
		t.Fatalf("iniciar la sesion: %d, Location=%q, cuerpo %s; quiero 303 a la sesion", w.Code, w.Header().Get("Location"), w.Body.String())
	}
	return w.Header().Get("Location")
}

// accion hace un POST de la sesion y exige que vuelva a la pantalla de la sesion.
func (e entorno) accion(t *testing.T, sesion, accion string, f url.Values) {
	t.Helper()
	if w := e.post(sesion+"/"+accion, f); w.Code != http.StatusSeeOther || w.Header().Get("Location") != sesion {
		t.Fatalf("%s: %d, Location=%q, cuerpo %s; quiero 303 a la sesion", accion, w.Code, w.Header().Get("Location"), w.Body.String())
	}
}

func voto(integrante, valor string) url.Values {
	return url.Values{"integrante": {integrante}, "estimacion": {valor}}
}

// CA-T020-1 a CA-T020-4, RN-T020-2, RN-T020-5: el recorrido completo con el router real y la base
// en memoria. Desde el backlog se inicia la sesion; se vota sin ver valores; al revelar se ven los
// votos y la divergencia; al registrar el acuerdo la sesion queda finalizada y la historia estimada.
func TestPantallasPoker_Recorrido(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto, historia, ids := mesaDePoker(t, e)
	iniciar := "/proyectos/" + proyecto + "/backlog/" + historia + "/planning-poker"

	contiene(t, e.get("/proyectos/"+proyecto+"/backlog"), `hx-post="`+iniciar+`"`, "Planning Poker")

	sesion := e.iniciarSesion(t, proyecto, historia)
	votando := e.get(sesion)
	contiene(t, votando, "Alta de proyectos", "Ronda 1", "votando", "Ana", "Beto",
		`name="integrante"`, `name="estimacion"`, `aria-label="Votar 5"`, `hx-post="`+sesion+`/votos"`,
		`aria-label="Revelar los votos de la ronda 1"`)
	noContiene(t, votando, "Abrir nueva ronda", `hx-post="`+sesion+`/acuerdo"`)

	// RN-T020-2: votar muestra quien voto y quien falta, nunca el valor.
	e.accion(t, sesion, "votos", voto(ids["Ana"], "3"))
	conUnVoto := e.get(sesion)
	contiene(t, conUnVoto, "Votaron", "Faltan")
	noContiene(t, conUnVoto, "<td>3</td>", "divergencia")

	e.accion(t, sesion, "votos", voto(ids["Beto"], "8"))
	e.accion(t, sesion, "revelar", url.Values{"integrante": {ids["Ana"]}})
	revelada := e.get(sesion)
	contiene(t, revelada, "revelada", "<td>3</td>", "<td>8</td>", "Hay divergencia: de 3 a 8",
		`hx-post="`+sesion+`/nueva-ronda"`, `hx-post="`+sesion+`/acuerdo"`)
	noContiene(t, revelada, `hx-post="`+sesion+`/votos"`, `/revelar"`)

	e.accion(t, sesion, "acuerdo", voto(ids["Beto"], "5"))
	finalizada := e.get(sesion)
	contiene(t, finalizada, "finalizada", "Estimación acordada: 5")
	noContiene(t, finalizada, `name="estimacion"`, `/votos"`, `/revelar"`, `/nueva-ronda"`, `/acuerdo"`)

	// CA-T020-4: la historia queda estimada con el valor acordado.
	puntos, _ := e.base.Items(proyecto)[0].StoryPoints.Valor()
	if puntos != 5 {
		t.Errorf("la historia tiene %d story points, quiero 5", puntos)
	}
}

// RN-T020-1: con HTMX, cada accion responde HX-Redirect a la sesion con el cuerpo vacio.
func TestPantallasPoker_ConHTMX(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto, historia, ids := mesaDePoker(t, e)
	sesion := e.iniciarSesion(t, proyecto, historia)

	w := e.enviar(http.MethodPost, sesion+"/votos", voto(ids["Ana"], "5"), true)

	if w.Code != http.StatusOK || w.Header().Get("HX-Redirect") != sesion || w.Body.Len() != 0 {
		t.Errorf("con HTMX: %d, HX-Redirect=%q, cuerpo=%q; quiero 200, la sesion y cuerpo vacio", w.Code, w.Header().Get("HX-Redirect"), w.Body.String())
	}
}

// CA-T020-5, CL-T020-6, CL-T020-7: con divergencia se abre la ronda 2, que vuelve a ocultar los
// votos; la ronda 1 se sigue viendo. Sin divergencia no se ofrece la nueva ronda.
func TestPantallasPoker_NuevaRonda(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto, historia, ids := mesaDePoker(t, e)
	sesion := e.iniciarSesion(t, proyecto, historia)
	e.accion(t, sesion, "votos", voto(ids["Ana"], "3"))
	e.accion(t, sesion, "votos", voto(ids["Beto"], "8"))
	e.accion(t, sesion, "revelar", url.Values{"integrante": {ids["Ana"]}})

	e.accion(t, sesion, "nueva-ronda", url.Values{"integrante": {ids["Beto"]}})
	e.accion(t, sesion, "votos", voto(ids["Ana"], "13"))

	ronda2 := e.get(sesion)
	contiene(t, ronda2, "Ronda 2", "votando", "Rondas anteriores", "<td>3</td>", "<td>8</td>",
		`aria-label="Revelar los votos de la ronda 2"`)
	noContiene(t, ronda2, "<td>13</td>")

	// Sin divergencia: solo el acuerdo.
	e.accion(t, sesion, "votos", voto(ids["Beto"], "13"))
	e.accion(t, sesion, "revelar", url.Values{"integrante": {ids["Beto"]}})
	sinDivergencia := e.get(sesion)
	contiene(t, sinDivergencia, "<td>13</td>", "Sin divergencia", `hx-post="`+sesion+`/acuerdo"`)
	noContiene(t, sinDivergencia, "Abrir nueva ronda")
}

// CA-T020-6, CL-T020-1, CL-T020-2: con una sesion abierta el backlog lleva a ella y no se inicia
// otra; una historia completada no ofrece Planning Poker.
func TestPantallasPoker_SesionAbiertaYCompletada(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto, historia, _ := mesaDePoker(t, e)
	sesion := e.iniciarSesion(t, proyecto, historia)

	backlog := e.get("/proyectos/" + proyecto + "/backlog")
	contiene(t, backlog, `href="`+sesion+`"`)
	noContiene(t, backlog, `hx-post="/proyectos/`+proyecto+`/backlog/`+historia+`/planning-poker"`)

	w := e.post("/proyectos/"+proyecto+"/backlog/"+historia+"/planning-poker", nil)
	if w.Code != http.StatusConflict {
		t.Errorf("segunda sesion: %d, quiero 409", w.Code)
	}

	completada := e.historiaLista(t, proyecto, "Historia completada", "3")
	sprint := e.crearSprint(t, proyecto, "Primer incremento")
	e.post(rutaDeSprint(proyecto, sprint)+"/historias", url.Values{"item": {completada}})
	if w := e.post(rutaDeSprint(proyecto, sprint)+"/historias/"+completada+"/completar", nil); w.Code != http.StatusSeeOther {
		t.Fatalf("completar: %d %s", w.Code, w.Body.String())
	}
	noContiene(t, e.get("/proyectos/"+proyecto+"/backlog"), `/backlog/`+completada+`/planning-poker"`)
}

// CA-T020-7, CL-T020-3, CL-T020-4, CL-T020-10: los errores salen con el formato uniforme de T-005
// y la sesion no cambia.
func TestPantallasPoker_Errores(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto, historia, ids := mesaDePoker(t, e)
	sesion := e.iniciarSesion(t, proyecto, historia)
	antes := e.get(sesion).Body.String()

	casos := []struct {
		nombre, accion string
		datos          url.Values
		estado         int
		textos         []string
	}{
		{"fuera de la escala", "votos", voto(ids["Ana"], "4"), http.StatusUnprocessableEntity, []string{`data-campo="estimacion"`}},
		{"revelar sin votos", "revelar", url.Values{"integrante": {ids["Ana"]}}, http.StatusConflict, []string{"no hay votos para revelar en esta ronda"}},
		{"acordar sin revelar", "acuerdo", voto(ids["Ana"], "5"), http.StatusConflict, []string{"la estimacion acordada se registra con los votos revelados"}},
		{"nueva ronda sin revelar", "nueva-ronda", url.Values{"integrante": {ids["Ana"]}}, http.StatusConflict, []string{"la nueva ronda se abre con los votos revelados"}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := e.post(sesion+"/"+c.accion, c.datos)
			if w.Code != c.estado {
				t.Fatalf("%s: %d %s, quiero %d", c.accion, w.Code, w.Body.String(), c.estado)
			}
			contiene(t, w, c.textos...)
			if despues := e.get(sesion).Body.String(); despues != antes {
				t.Errorf("la sesion cambio despues del rechazo")
			}
		})
	}

	// Un integrante que no participa: se registra despues de iniciar la sesion.
	e.post("/proyectos/"+proyecto+"/integrantes", url.Values{"nombre": {"Caro"}, "email": {"caro@ejemplo.test"}, "rol": {"product_builder"}})
	w := e.post(sesion+"/votos", voto(e.integrantesPorNombre(proyecto)["Caro"], "5"))
	if w.Code != http.StatusConflict {
		t.Fatalf("voto de quien no participa: %d, quiero 409", w.Code)
	}
	contiene(t, w, "el integrante no participa de esta sesion de planning poker")

	// CL-T020-3: con un solo integrante no se inicia.
	solo := e.crearProyecto(t, "Solo")
	e.post("/proyectos/"+solo+"/integrantes", url.Values{"nombre": {"Dani"}, "email": {"dani@ejemplo.test"}, "rol": {"product_builder"}})
	e.crearHistoria(t, solo, "Historia sola", "must", "")
	w = e.post("/proyectos/"+solo+"/backlog/"+e.idDeHistoria(t, solo, "Historia sola")+"/planning-poker", nil)
	if w.Code != http.StatusConflict {
		t.Errorf("iniciar con un integrante: %d, quiero 409", w.Code)
	}
}

// RN-T020-4, CL-T020-9: una sesion solo se alcanza por la ruta de su proyecto; una inexistente o
// con el ID mal formado responde 404.
func TestPantallasPoker_SesionDeOtroProyecto(t *testing.T) {
	e := nuevoEntorno(t)
	proyecto, historia, ids := mesaDePoker(t, e)
	sesion := e.iniciarSesion(t, proyecto, historia)
	sesionID := idDeSesionEnLaRuta.FindStringSubmatch(sesion)[2]
	otro := e.crearProyecto(t, "Otro")

	for _, ruta := range []string{
		"/proyectos/" + otro + "/planning-poker/" + sesionID,
		"/proyectos/" + proyecto + "/planning-poker/00000000-0000-0000-0000-000000000000",
		"/proyectos/" + proyecto + "/planning-poker/no-es-un-id",
	} {
		if w := e.get(ruta); w.Code != http.StatusNotFound {
			t.Errorf("GET %s: %d, quiero 404", ruta, w.Code)
		}
	}
	w := e.post("/proyectos/"+otro+"/planning-poker/"+sesionID+"/votos", voto(ids["Ana"], "5"))
	if w.Code != http.StatusNotFound {
		t.Errorf("votar por la ruta de otro proyecto: %d, quiero 404", w.Code)
	}
}
