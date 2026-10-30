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
//	/proyectos/{id}/esfuerzo       esfuerzo (horas registradas)
//	/proyectos/{id}/dashboard      dashboard de metricas y graficos
//	/proyectos/{id}/defectos       defectos
//	/proyectos/{id}/reporte        reporte en pantalla
package plantillas

import (
	"net/url"
	"strings"
)

const nombreApp = "Software Metrics"

// configuracionHtmx es el contenido de <meta name="htmx-config"> (RN-T004-10). HTMX no
// intercambia las respuestas 4xx ni 5xx por defecto, y sin esto el fragmento de error de
// T-005 no se veria. Los errores uniformes (404, 409, 422 y 500) se intercambian dentro
// de #mensajes; el resto de las reglas son las que trae HTMX 2.x. La primera regla cuyo
// code coincide con el estado es la que vale, por eso la de los errores va antes que la
// de los demas 4xx y 5xx.
const configuracionHtmx = `{"responseHandling":[` +
	`{"code":"204","swap":false},` +
	`{"code":"[23]..","swap":true},` +
	`{"code":"^(404|409|422|500)$","swap":true,"error":true,"target":"#mensajes"},` +
	`{"code":"[45]..","swap":false,"error":true}` +
	`]}`

// Seccion identifica el item del menu que corresponde a la pagina actual.
type Seccion string

const (
	SeccionNinguna   Seccion = ""
	SeccionInicio    Seccion = "inicio"
	SeccionProyectos Seccion = "proyectos"
	SeccionResumen   Seccion = "resumen"
	SeccionBacklog   Seccion = "backlog"
	SeccionSprints   Seccion = "sprints"
	SeccionEsfuerzo  Seccion = "esfuerzo"
	SeccionDashboard Seccion = "dashboard"
	SeccionDefectos  Seccion = "defectos"
	SeccionReporte   Seccion = "reporte"
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
		{Texto: "Esfuerzo", Href: base + "/esfuerzo", Activo: p.Activo == SeccionEsfuerzo},
		{Texto: "Dashboard", Href: base + "/dashboard", Activo: p.Activo == SeccionDashboard},
		{Texto: "Defectos", Href: base + "/defectos", Activo: p.Activo == SeccionDefectos},
		{Texto: "Reporte", Href: base + "/reporte", Activo: p.Activo == SeccionReporte},
	}
}

// Clases son las clases de Tailwind del enlace: el activo se distingue por mas
// que el color (negrita y subrayado), para no depender solo de la vista del color.
func (e Enlace) Clases() string {
	const base = "inline-block rounded-md border-b-2 px-3 py-1.5 text-sm " + claseFoco
	if e.Activo {
		return base + " border-indigo-700 bg-indigo-50 font-semibold text-indigo-800"
	}
	return base + " border-transparent font-medium text-slate-700 hover:bg-slate-100 hover:text-slate-900"
}
