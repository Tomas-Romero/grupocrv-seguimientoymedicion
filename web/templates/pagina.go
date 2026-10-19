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
func (p Pagina) TituloCompleto() string { return "" }

// EnlacesGlobales son los destinos que existen fuera de un proyecto.
func (p Pagina) EnlacesGlobales() []Enlace { return nil }

// EnlacesProyecto son los destinos de un proyecto; vacio si la pagina no tiene proyecto.
func (p Pagina) EnlacesProyecto() []Enlace { return nil }