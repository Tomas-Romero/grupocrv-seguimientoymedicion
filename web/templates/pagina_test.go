package plantillas_test

import (
	"slices"
	"strings"
	"testing"

	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// T-004: el titulo de la pestana siempre lleva el nombre de la aplicacion.
func TestTituloCompleto(t *testing.T) {
	casos := []struct {
		nombre   string
		titulo   string
		esperado string
	}{
		{"caso normal: titulo de la pagina mas la aplicacion", "Proyectos", "Proyectos | Software Metrics"},
		{"caso alternativo: se recortan los espacios de los extremos", "  Backlog  ", "Backlog | Software Metrics"},
		{"caso limite: sin titulo queda solo la aplicacion", "", "Software Metrics"},
		{"caso limite: un titulo de solo espacios cuenta como vacio", "   ", "Software Metrics"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			pagina := plantillas.Pagina{Titulo: c.titulo}
			if obtenido := pagina.TituloCompleto(); obtenido != c.esperado {
				t.Errorf("TituloCompleto() = %q, se esperaba %q", obtenido, c.esperado)
			}
		})
	}
}

// T-004: el menu global marca como activo solo el destino de la seccion actual.
func TestEnlacesGlobales(t *testing.T) {
	casos := []struct {
		nombre    string
		activo    plantillas.Seccion
		esperados []plantillas.Enlace
	}{
		{
			nombre: "caso normal: sin seccion activa",
			activo: plantillas.SeccionNinguna,
			esperados: []plantillas.Enlace{
				{Texto: "Inicio", Href: "/", Activo: false},
				{Texto: "Proyectos", Href: "/proyectos", Activo: false},
			},
		},
		{
			nombre: "caso alternativo: Proyectos activa",
			activo: plantillas.SeccionProyectos,
			esperados: []plantillas.Enlace{
				{Texto: "Inicio", Href: "/", Activo: false},
				{Texto: "Proyectos", Href: "/proyectos", Activo: true},
			},
		},
		{
			nombre: "caso alternativo: Inicio activo",
			activo: plantillas.SeccionInicio,
			esperados: []plantillas.Enlace{
				{Texto: "Inicio", Href: "/", Activo: true},
				{Texto: "Proyectos", Href: "/proyectos", Activo: false},
			},
		},
		{
			nombre: "caso limite: una seccion de proyecto no activa ningun enlace global",
			activo: plantillas.SeccionBacklog,
			esperados: []plantillas.Enlace{
				{Texto: "Inicio", Href: "/", Activo: false},
				{Texto: "Proyectos", Href: "/proyectos", Activo: false},
			},
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			pagina := plantillas.Pagina{Activo: c.activo}
			if obtenidos := pagina.EnlacesGlobales(); !slices.Equal(obtenidos, c.esperados) {
				t.Errorf("EnlacesGlobales() = %+v, se esperaba %+v", obtenidos, c.esperados)
			}
		})
	}
}

// enlacesDeProyecto arma los siete enlaces esperados de un proyecto; base es la
// ruta del proyecto con el ID ya escapado.
func enlacesDeProyecto(base, activo string) []plantillas.Enlace {
	return []plantillas.Enlace{
		{Texto: "Resumen", Href: base, Activo: activo == "Resumen"},
		{Texto: "Backlog", Href: base + "/backlog", Activo: activo == "Backlog"},
		{Texto: "Sprints", Href: base + "/sprints", Activo: activo == "Sprints"},
		{Texto: "Esfuerzo", Href: base + "/esfuerzo", Activo: activo == "Esfuerzo"},
		{Texto: "Dashboard", Href: base + "/dashboard", Activo: activo == "Dashboard"},
		{Texto: "Defectos", Href: base + "/defectos", Activo: activo == "Defectos"},
		{Texto: "Reporte", Href: base + "/reporte", Activo: activo == "Reporte"},
	}
}

// T-004: el menu del proyecto cuelga de la ruta /proyectos/{id}.
func TestEnlacesProyecto(t *testing.T) {
	casos := []struct {
		nombre    string
		id        string
		activo    plantillas.Seccion
		esperados []plantillas.Enlace
	}{
		{
			nombre:    "caso normal: Backlog activo",
			id:        "p1",
			activo:    plantillas.SeccionBacklog,
			esperados: enlacesDeProyecto("/proyectos/p1", "Backlog"),
		},
		{
			nombre:    "caso alternativo: Resumen activo",
			id:        "p1",
			activo:    plantillas.SeccionResumen,
			esperados: enlacesDeProyecto("/proyectos/p1", "Resumen"),
		},
		{
			nombre:    "caso alternativo: Esfuerzo activo (T-019)",
			id:        "p1",
			activo:    plantillas.SeccionEsfuerzo,
			esperados: enlacesDeProyecto("/proyectos/p1", "Esfuerzo"),
		},
		{
			nombre:    "caso alternativo: Dashboard activo (US-030)",
			id:        "p1",
			activo:    plantillas.SeccionDashboard,
			esperados: enlacesDeProyecto("/proyectos/p1", "Dashboard"),
		},
		{
			nombre:    "caso alternativo: Reporte activo (US-033)",
			id:        "p1",
			activo:    plantillas.SeccionReporte,
			esperados: enlacesDeProyecto("/proyectos/p1", "Reporte"),
		},
		{
			nombre:    "caso limite: una seccion global no activa ningun enlace del proyecto",
			id:        "p1",
			activo:    plantillas.SeccionProyectos,
			esperados: enlacesDeProyecto("/proyectos/p1", ""),
		},
		{
			nombre:    "caso limite: el ID se escapa para no romper la ruta",
			id:        "a/b c",
			activo:    plantillas.SeccionNinguna,
			esperados: enlacesDeProyecto("/proyectos/a%2Fb%20c", ""),
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			pagina := plantillas.Pagina{Activo: c.activo, Proyecto: &plantillas.ProyectoActual{ID: c.id, Nombre: "Demo"}}
			if obtenidos := pagina.EnlacesProyecto(); !slices.Equal(obtenidos, c.esperados) {
				t.Errorf("EnlacesProyecto() = %+v, se esperaba %+v", obtenidos, c.esperados)
			}
		})
	}
}

// T-004: fuera de un proyecto no hay menu de proyecto.
func TestEnlacesProyecto_SinProyecto(t *testing.T) {
	pagina := plantillas.Pagina{Activo: plantillas.SeccionBacklog}
	if obtenidos := pagina.EnlacesProyecto(); len(obtenidos) != 0 {
		t.Errorf("EnlacesProyecto() sin proyecto = %+v, se esperaba vacio", obtenidos)
	}
}

// US-030 / RN-030-12: el menu del proyecto no tiene un enlace a /metricas. Esa pantalla nunca existio y el enlace respondia
// «Pagina no encontrada»: el dashboard es la pantalla de las metricas.
func TestEnlacesProyecto_SinEnlaceAMetricas(t *testing.T) {
	pagina := plantillas.Pagina{Proyecto: &plantillas.ProyectoActual{ID: "p1", Nombre: "Demo"}}

	for _, e := range pagina.EnlacesProyecto() {
		if strings.Contains(e.Href, "metricas") || e.Texto == "Métricas" {
			t.Errorf("el menu tiene el enlace %q (%s): esa pantalla no existe", e.Texto, e.Href)
		}
	}
}
