// Package plantillas contiene las vistas templ de la aplicacion.
//
// No tiene reglas de negocio: recibe datos ya preparados y los dibuja.
//
// Esquema de URL (lo usan las pantallas de T-011):
//
//	/                              inicio
//	/proyectos                     lista de proyectos
//	/proyectos/{id}                resumen de un proyecto
//	/proyectos/{id}/backlog        backlog
//	/proyectos/{id}/sprints        sprints
//	/proyectos/{id}/metricas       metricas
//	/proyectos/{id}/defectos       defectos
package plantillas

import (
	"net/url"
	"strings"
)

const nombreApp = "Software Metrics"

// Seccion identifica el item del menu que corresponde a la pagina actual.
type Seccion string

const (
	SeccionNinguna   Seccion = ""
	SeccionInicio    Seccion = "inicio"
	SeccionProyectos Seccion = "proyectos"
	SeccionResumen   Seccion = "resumen"
	SeccionBacklog   Seccion = "backlog"
	SeccionSprints   Seccion = "sprints"
	SeccionMetricas  Seccion = "metricas"
	SeccionDefectos  Seccion = "defectos"
)

// ProyectoActual es el proyecto dentro del cual esta la pagina, si hay uno.
type ProyectoActual struct {
	ID     string
	Nombre string
}

// Pagina son los datos que el layout necesita de cada pagina.
type Pagina struct {
	Titulo   string
	Activo   Seccion
	Proyecto *ProyectoActual
}

// Enlace es un destino del menu.
type Enlace struct {
	Texto  string
	Href   string
	Activo bool
}

// TituloCompleto es el texto de la pestana del navegador.
func (p Pagina) TituloCompleto() string {
	if titulo := strings.TrimSpace(p.Titulo); titulo != "" {
		return titulo + " | " + nombreApp
	}
	return nombreApp
}

// EnlacesGlobales son los destinos que existen fuera de un proyecto.
func (p Pagina) EnlacesGlobales() []Enlace {
	return []Enlace{
		{Texto: "Inicio", Href: "/", Activo: p.Activo == SeccionInicio},
		{Texto: "Proyectos", Href: "/proyectos", Activo: p.Activo == SeccionProyectos},
	}
}

// EnlacesProyecto son los destinos de un proyecto; vacio si la pagina no tiene proyecto.
func (p Pagina) EnlacesProyecto() []Enlace {
	if p.Proyecto == nil {
		return nil
	}
	// El ID viaja en la ruta: se escapa para que una "/" o un espacio no la rompan.
	base := "/proyectos/" + url.PathEscape(p.Proyecto.ID)
	return []Enlace{
		{Texto: "Resumen", Href: base, Activo: p.Activo == SeccionResumen},
		{Texto: "Backlog", Href: base + "/backlog", Activo: p.Activo == SeccionBacklog},
		{Texto: "Sprints", Href: base + "/sprints", Activo: p.Activo == SeccionSprints},
		{Texto: "Métricas", Href: base + "/metricas", Activo: p.Activo == SeccionMetricas},
		{Texto: "Defectos", Href: base + "/defectos", Activo: p.Activo == SeccionDefectos},
	}
}