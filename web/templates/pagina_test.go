package plantillas_test

import (
	"slices"
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

// enlacesDeProyecto arma los seis enlaces esperados de un proyecto; base es la
// ruta del proyecto con el ID ya escapado.
func enlacesDeProyecto(base, activo string) []plantillas.Enlace {
	return []plantillas.Enlace{
		{Texto: "Resumen", Href: base, Activo: activo == "Resumen"},
		{Texto: "Backlog", Href: base + "/backlog", Activo: activo == "Backlog"},
		{Texto: "Sprints", Href: base + "/sprints", Activo: activo == "Sprints"},
		{Texto: "Esfuerzo", Href: base + "/esfuerzo", Activo: activo == "Esfuerzo"},
		{Texto: "Métricas", Href: base + "/metricas", Activo: activo == "Métricas"},
		{Texto: "Defectos", Href: base + "/defectos", Activo: activo == "Defectos"},
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
			nombre:    "caso alternativo: Metricas activo",
			id:        "p1",
			activo:    plantillas.SeccionMetricas,
			esperados: enlacesDeProyecto("/proyectos/p1", "Métricas"),
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
