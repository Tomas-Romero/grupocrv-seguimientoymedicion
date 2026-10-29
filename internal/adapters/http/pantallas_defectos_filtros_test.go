package httpadapter_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// Tests de los filtros de la pantalla de defectos (spec US-024, RN-024-11 y RN-024-12). El filtro viaja en un
// formulario GET con los parametros `sprint` y `severidad` (Hueco 1 de la spec, como en T-019).

// defectosParaFiltrar es un proyecto con tres sprints y tres defectos:
//
//	"Falla uno alta"  Sprint 1, alta     "Falla uno baja"  Sprint 1, baja     "Falla dos alta"  Sprint 2, alta
//
// El Sprint 3 no tiene defectos.
type defectosParaFiltrar struct {
	escenarioDeDefectos
	segundo, tercero string
}

func (e entorno) defectosParaFiltrar(t *testing.T) defectosParaFiltrar {
	t.Helper()
	d := e.proyectoParaDefectos(t, "Demo")
	segundo := e.segundoSprint(t, d)
	e.cerrarSprint(t, d.proyecto, segundo)
	tercero := e.crearSprintConFechas(t, d.proyecto, "Tercera meta", "2026-11-02", "2026-11-15")
	e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Falla uno alta", defecto.SeveridadAlta)
	e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Falla uno baja", defecto.SeveridadBaja)
	e.defectoGuardado(t, d.proyecto, d.historia, segundo, "Falla dos alta", defecto.SeveridadAlta)
	return defectosParaFiltrar{escenarioDeDefectos: d, segundo: segundo, tercero: tercero}
}

// filtrada pide la lista con los parametros del filtro; un valor vacio se manda igual, como lo hace el formulario.
func (e entorno) filtrada(proyecto, sprint, severidad string) *httptest.ResponseRecorder {
	return e.get(rutaDeDefectos(proyecto) + "?" + url.Values{"sprint": {sprint}, "severidad": {severidad}}.Encode())
}

// defectosEnLaTabla son los textos de los tres defectos de defectosParaFiltrar que aparecen en la respuesta.
func defectosEnLaTabla(w *httptest.ResponseRecorder) []string {
	var presentes []string
	for _, descripcion := range []string{"Falla uno alta", "Falla uno baja", "Falla dos alta"} {
		if strings.Contains(w.Body.String(), descripcion) {
			presentes = append(presentes, descripcion)
		}
	}
	return presentes
}

func quieroDefectos(t *testing.T, w *httptest.ResponseRecorder, quiero ...string) {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("estado %d, quiero 200:\n%s", w.Code, w.Body.String())
	}
	if got := defectosEnLaTabla(w); strings.Join(got, "|") != strings.Join(quiero, "|") {
		t.Errorf("defectos en la lista = %v, quiero %v", got, quiero)
	}
}

// US-024 / CA-024-1 / RN-024-2 / RN-024-11: filtrar por un sprint muestra solo los defectos detectados en el.
func TestPantallasDefectosFiltros_PorSprint(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)

	quieroDefectos(t, e.filtrada(d.proyecto, d.sprint, ""), "Falla uno alta", "Falla uno baja")
	quieroDefectos(t, e.filtrada(d.proyecto, d.segundo, ""), "Falla dos alta")
}

// US-024 / CA-024-1 / RN-024-3 / RN-024-11: filtrar por una severidad muestra solo los defectos de esa severidad,
// de cualquier sprint.
func TestPantallasDefectosFiltros_PorSeveridad(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)

	quieroDefectos(t, e.filtrada(d.proyecto, "", "alta"), "Falla uno alta", "Falla dos alta")
	quieroDefectos(t, e.filtrada(d.proyecto, "", "baja"), "Falla uno baja")
}

// US-024 / CA-024-2 / RN-024-4: los dos filtros se combinan con Y.
func TestPantallasDefectosFiltros_Combinados(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)

	quieroDefectos(t, e.filtrada(d.proyecto, d.sprint, "alta"), "Falla uno alta")
	quieroDefectos(t, e.filtrada(d.proyecto, d.segundo, "alta"), "Falla dos alta")
}

// US-024 / CA-024-5 / RN-024-1 / CL-024-1: sin parametros, o con los dos vacios, la pantalla es la de T-021:
// todos los defectos. Un solo valor vacio no filtra por esa dimension.
func TestPantallasDefectosFiltros_SinFiltros(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)
	todos := []string{"Falla uno alta", "Falla uno baja", "Falla dos alta"}

	quieroDefectos(t, e.get(rutaDeDefectos(d.proyecto)), todos...)
	quieroDefectos(t, e.filtrada(d.proyecto, "", ""), todos...)
	noContiene(t, e.get(rutaDeDefectos(d.proyecto)), "Ningún defecto cumple los filtros", "Todavía no hay defectos")
}

// US-024 / CA-024-3 / RN-024-5 / RN-024-12 / CL-024-6 / CL-024-7 / CL-024-8: un filtro sin coincidencias responde
// 200 con la lista vacia, sin tabla y sin mensaje de error. El texto es el de los filtros y no el de "todavia no
// hay defectos", que seria falso: el proyecto si tiene.
func TestPantallasDefectosFiltros_SinCoincidencias(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)

	casos := []struct{ nombre, sprint, severidad string }{
		{"sprint del proyecto sin defectos", d.tercero, ""},
		{"severidad sin defectos", "", "critica"},
		{"combinacion sin interseccion", d.segundo, "baja"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := e.filtrada(d.proyecto, c.sprint, c.severidad)
			quieroDefectos(t, w)
			contiene(t, w, "Ningún defecto cumple los filtros")
			noContiene(t, w, "<table", `role="alert"`, "data-campo=", "Todavía no hay defectos")
		})
	}
}

// US-024 / RN-024-5 / CL-024-9: un proyecto sin ningun defecto, con filtros validos, responde 200 con la lista
// vacia. Sin defectos el proyecto no tiene nada que "cumplir", pero el filtro esta puesto: se informa el filtro.
func TestPantallasDefectosFiltros_ProyectoSinDefectos(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.proyectoParaDefectos(t, "Sin defectos")

	w := e.filtrada(d.proyecto, d.sprint, "alta")
	if w.Code != http.StatusOK {
		t.Fatalf("estado %d, quiero 200:\n%s", w.Code, w.Body.String())
	}
	contiene(t, w, "Ningún defecto cumple los filtros")
	noContiene(t, w, "<table", `role="alert"`)
}

// selectDe devuelve el <select> con ese id, para mirar sus opciones sin confundirlo con las del formulario de alta,
// que tiene sus propias listas de severidades y sprints.
func selectDe(t *testing.T, w *httptest.ResponseRecorder, id string) string {
	t.Helper()
	cuerpo := w.Body.String()
	inicio := strings.Index(cuerpo, `<select id="`+id+`"`)
	if inicio < 0 {
		t.Fatalf("la pantalla no tiene el <select id=%q>:\n%s", id, cuerpo)
	}
	return cuerpo[inicio : inicio+strings.Index(cuerpo[inicio:], "</select>")]
}

// textosDeOpciones son los textos de las <option> de un <select>, en el orden en que salen.
func textosDeOpciones(desplegable string) []string {
	var textos []string
	for _, m := range regexp.MustCompile(`<option[^>]*>([^<]*)</option>`).FindAllStringSubmatch(desplegable, -1) {
		textos = append(textos, m[1])
	}
	return textos
}

// US-024 / RN-024-11 / CA-024-3: la pantalla ofrece los dos filtros en un formulario GET. El de sprint ofrece TODOS
// los sprints del proyecto, tengan defectos o no, por numero descendente; el de severidad, las cuatro con sus
// etiquetas. Cada uno tiene una opcion para no filtrar, que es la elegida cuando no hay filtro.
func TestPantallasDefectosFiltros_Controles(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)

	w := e.get(rutaDeDefectos(d.proyecto))
	contiene(t, w, `<form method="get" action="`+rutaDeDefectos(d.proyecto)+`"`,
		`<label for="filtro-sprint"`, `<label for="filtro-severidad"`, `name="sprint"`, `name="severidad"`, "Filtrar")

	sprints := selectDe(t, w, "filtro-sprint")
	if got := strings.Join(textosDeOpciones(sprints), "|"); got != "Todos|Sprint 3|Sprint 2|Sprint 1" {
		t.Errorf("opciones del filtro de sprint = %q, quiero todos los sprints, del mas nuevo al mas viejo", got)
	}
	for _, id := range []string{d.sprint, d.segundo, d.tercero} {
		contiene(t, w, `<option value="`+id+`"`)
	}
	if !regexp.MustCompile(`<option value=""[^>]*\sselected[^>]*>Todos<`).MatchString(sprints) {
		t.Errorf("sin filtro tiene que estar elegida la opcion Todos:\n%s", sprints)
	}

	severidades := selectDe(t, w, "filtro-severidad")
	if got := strings.Join(textosDeOpciones(severidades), "|"); got != "Todas|Crítica|Alta|Media|Baja" {
		t.Errorf("opciones del filtro de severidad = %q, quiero Todas y las cuatro", got)
	}
	for _, valor := range []string{"critica", "alta", "media", "baja"} {
		contiene(t, w, `<option value="`+valor+`"`)
	}
	if !regexp.MustCompile(`<option value=""[^>]*\sselected[^>]*>Todas<`).MatchString(severidades) {
		t.Errorf("sin filtro tiene que estar elegida la opcion Todas:\n%s", severidades)
	}
}

// US-024 / RN-024-11: despues de filtrar, la pantalla deja elegido lo que se pidio, para que se vea que filtro
// esta activo y se pueda cambiar.
func TestPantallasDefectosFiltros_ConservaLoElegido(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)

	w := e.filtrada(d.proyecto, d.segundo, "alta")
	if !regexp.MustCompile(`<option value="` + d.segundo + `"[^>]*\sselected[^>]*>Sprint 2<`).MatchString(selectDe(t, w, "filtro-sprint")) {
		t.Errorf("tiene que quedar elegido el Sprint 2:\n%s", selectDe(t, w, "filtro-sprint"))
	}
	if !regexp.MustCompile(`<option value="alta"[^>]*\sselected[^>]*>Alta<`).MatchString(selectDe(t, w, "filtro-severidad")) {
		t.Errorf("tiene que quedar elegida la severidad Alta:\n%s", selectDe(t, w, "filtro-severidad"))
	}
}

// US-024 / RN-024-11 / CL-024-12: una fila se ve igual con o sin filtro: las mismas columnas, etiquetas y botones
// (RN-T021-13 y RN-T021-14), con el sprint de deteccion y de resolucion que corresponden.
func TestPantallasDefectosFiltros_LaFilaSeVeIgual(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)
	resuelto := e.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Falla resuelta", defecto.SeveridadCritica)
	e.forzar(t, resuelto, defecto.EstadoResuelto, d.segundo)

	completa := e.get(rutaDeDefectos(d.proyecto))
	noContiene(t, e.filtrada(d.proyecto, d.sprint, ""), "Falla dos alta")
	for _, descripcion := range []string{"Falla uno alta", "Falla resuelta"} {
		filtrada := e.filtrada(d.proyecto, d.sprint, "")
		if fila(t, filtrada, descripcion) != fila(t, completa, descripcion) {
			t.Errorf("la fila de %q cambia al filtrar:\nfiltrada: %s\ncompleta: %s", descripcion, fila(t, filtrada, descripcion), fila(t, completa, descripcion))
		}
	}
	contiene(t, e.filtrada(d.proyecto, d.sprint, "critica"), "Resuelto", "Sprint 2")
}

// US-024 / CA-024-4 / RN-024-7 / RN-024-12 / CL-024-13 / CL-024-14: un sprint de otro proyecto, inexistente o mal
// formado es un 404 con la pagina del layout, igual que el sprint de RN-T021-12; no una lista vacia ni un 200.
func TestPantallasDefectosFiltros_SprintInexistente(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)
	otro := e.proyectoParaDefectos(t, "Otro proyecto")

	casos := []struct{ nombre, sprint string }{
		{"sprint de otro proyecto", otro.sprint},
		{"sprint inexistente", "00000000-0000-0000-0000-000000000000"},
		{"sprint mal formado", "no-es-un-uuid"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			w := e.filtrada(d.proyecto, c.sprint, "")
			if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), "<!doctype html>") {
				t.Fatalf("estado %d, quiero 404 con layout:\n%s", w.Code, w.Body.String())
			}
			noContiene(t, w, "Falla uno alta", "Ningún defecto cumple los filtros")
		})
	}
}

// US-024 / CA-024-4 / RN-024-8 / RN-024-12 / CL-024-16: una severidad que no existe es un 422 con el mensaje del
// catalogo en el campo `severidad`, sin lista. No se normaliza: "Alta" y " alta " tambien se rechazan.
func TestPantallasDefectosFiltros_SeveridadInvalida(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)

	for _, severidad := range []string{"urgente", "Alta", "CRITICA", " alta ", " "} {
		t.Run(severidad, func(t *testing.T) {
			w := e.filtrada(d.proyecto, "", severidad)
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("estado %d, quiero 422:\n%s", w.Code, w.Body.String())
			}
			contiene(t, w, `data-campo="severidad"`, "la severidad tiene que ser critica, alta, media o baja")
			if n := strings.Count(w.Body.String(), "data-campo="); n != 1 {
				t.Errorf("hay %d errores de campo, quiero uno:\n%s", n, w.Body.String())
			}
			noContiene(t, w, "Falla uno alta")
		})
	}
}

// US-024 / CA-024-6 / RN-024-9 / CL-024-17 / CL-024-18: con el sprint inexistente y la severidad invalida a la vez
// se responde solo por el sprint (404); con el proyecto inexistente, solo por el proyecto (404). Corregido el
// sprint, aparece el error de la severidad.
func TestPantallasDefectosFiltros_OrdenDeLosRechazos(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)
	inexistente := "00000000-0000-0000-0000-000000000000"

	w := e.filtrada(d.proyecto, inexistente, "urgente")
	if w.Code != http.StatusNotFound {
		t.Errorf("sprint inexistente y severidad invalida: estado %d, quiero 404", w.Code)
	}
	noContiene(t, w, "data-campo=", "la severidad tiene que ser")

	if w := e.filtrada(inexistente, inexistente, "urgente"); w.Code != http.StatusNotFound {
		t.Errorf("proyecto inexistente con filtros invalidos: estado %d, quiero 404", w.Code)
	}
	if w := e.filtrada("no-es-un-uuid", d.sprint, "urgente"); w.Code != http.StatusNotFound {
		t.Errorf("proyecto mal formado con filtros invalidos: estado %d, quiero 404", w.Code)
	}
	if w := e.filtrada(d.proyecto, d.sprint, "urgente"); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("sprint corregido y severidad invalida: estado %d, quiero 422", w.Code)
	}
}

// US-024 / RN-024-10: filtrar solo lee, no cambia ningun defecto.
func TestPantallasDefectosFiltros_NoModificaNada(t *testing.T) {
	e := nuevoEntorno(t)
	d := e.defectosParaFiltrar(t)
	antes := e.base.Defectos(d.proyecto)

	e.filtrada(d.proyecto, d.sprint, "alta")
	e.filtrada(d.proyecto, d.tercero, "")
	e.filtrada(d.proyecto, "", "urgente")

	despues := e.base.Defectos(d.proyecto)
	if len(antes) != len(despues) {
		t.Fatalf("hay %d defectos, habia %d", len(despues), len(antes))
	}
	for i := range antes {
		if antes[i] != despues[i] {
			t.Errorf("el defecto %d cambio: %+v, antes %+v", i, despues[i], antes[i])
		}
	}
}

// baseCaidaAlVerificarElSprint es la base en memoria con la consulta de pertenencia del sprint caida: falla con un
// error que no es de ninguna regla.
type baseCaidaAlVerificarElSprint struct {
	*memoria.Base
	err error
}

func (b baseCaidaAlVerificarElSprint) ExisteSprintEnProyecto(context.Context, string, string) (bool, error) {
	return false, b.err
}

// US-024 / CL-024-20 / RN-024-12: una falla inesperada de la base al verificar el sprint o al listar, con filtro,
// responde 500 con el mensaje fijo de T-005, sin el detalle y sin lista parcial.
func TestPantallasDefectosFiltros_FallaInesperadaDeLaBase(t *testing.T) {
	detalle := "conexion rechazada por 10.0.0.7"
	base := memoria.NuevaBase()
	casos := []struct {
		nombre   string
		entorno  entorno
		pedirCon string
	}{
		{"al verificar el sprint", entornoConBase(base, baseCaidaAlVerificarElSprint{Base: base, err: errors.New(detalle)}), "sprint"},
		{"al listar con filtro de severidad", entornoConBase(base, baseCaida{Base: base, err: errors.New(detalle)}), "severidad"},
		{"al listar con filtro de sprint", entornoConBase(base, baseCaida{Base: base, err: errors.New(detalle)}), "sprint"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := c.entorno.proyectoParaDefectos(t, "Demo "+c.nombre)
			c.entorno.defectoGuardado(t, d.proyecto, d.historia, d.sprint, "Falla uno alta", defecto.SeveridadAlta)

			sprint, severidad := "", "alta"
			if c.pedirCon == "sprint" {
				sprint, severidad = d.sprint, ""
			}
			w := c.entorno.filtrada(d.proyecto, sprint, severidad)
			if w.Code != http.StatusInternalServerError {
				t.Fatalf("estado %d, quiero 500:\n%s", w.Code, w.Body.String())
			}
			contiene(t, w, "ocurrio un error inesperado")
			noContiene(t, w, detalle, "Falla uno alta", "<table")
		})
	}
}
