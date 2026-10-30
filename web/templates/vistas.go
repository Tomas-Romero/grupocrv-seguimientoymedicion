package plantillas

import (
	"strings"
	"unicode"
)

// Datos que dibujan las pantallas de proyectos, integrantes y backlog (spec T-011). Son
// datos ya preparados: las fechas vienen formateadas, las opciones armadas y los textos
// listos. Las plantillas no conocen el dominio ni deciden nada.

// Opcion es una opcion de una lista desplegable.
type Opcion struct {
	Valor   string
	Texto   string
	Elegida bool
}

// FilaProyecto es un proyecto en la lista.
type FilaProyecto struct {
	ID     string
	Nombre string
	Inicio string
	Fin    string
}

// VistaProyectos es la lista de proyectos.
type VistaProyectos struct {
	Proyectos []FilaProyecto
}

// FormProyecto es el formulario de alta o de edicion de un proyecto. Las fechas van como
// AAAA-MM-DD, el formato de un <input type="date">.
type FormProyecto struct {
	Encabezado  string
	Accion      string
	Volver      string
	Nombre      string
	Descripcion string
	FechaInicio string
	FechaFin    string
}

// FilaIntegrante es un integrante en el resumen de un proyecto.
type FilaIntegrante struct {
	Nombre string
	Email  string
	Rol    string
}

// VistaEstadoProyecto es la tarjeta de estado general de la pagina de un proyecto (spec US-004). Son textos ya armados:
// la plantilla solo los dibuja. Si Error no esta vacio, la tarjeta informa el error y no los demas datos.
type VistaEstadoProyecto struct {
	Estado      string
	Historias   string
	StoryPoints string
	Porcentaje  string
	// Sprint es «Sprint 2 en curso» (con su enlace en SprintEnlace) o «Sin sprint abierto» (sin enlace).
	Sprint       string
	SprintEnlace string
	Error        string
}

// VistaProyecto es el resumen de un proyecto.
type VistaProyecto struct {
	// Estado es la tarjeta de estado general (US-004).
	Estado      VistaEstadoProyecto
	ID          string
	Nombre      string
	Descripcion string
	Inicio      string
	Fin         string
	Editar      string
	Backlog     string
	Sprints     string
	Integrantes []FilaIntegrante
	// AltaIntegrante es la URL del formulario de integrantes; Roles, sus opciones.
	AltaIntegrante string
	Roles          []Opcion
}

// FilaItem es una historia en la tabla del backlog.
type FilaItem struct {
	Numero    int
	Titulo    string
	Prioridad string
	Estado    string
	// Puntos es el texto de los Story Points: el numero, o un guion si no esta estimada.
	Puntos    string
	Criterios int
	Editar    string
	Estimar   string
	Escala    []Opcion
	// IniciarPoker es la ruta para iniciar una sesion de Planning Poker, y SesionAbierta la de
	// la sesion abierta de la historia si la tiene (T-020). Una historia completada no tiene
	// ninguna de las dos.
	IniciarPoker  string
	SesionAbierta string
}

// FormItem es el formulario de alta o de edicion de una historia.
type FormItem struct {
	Encabezado  string
	Accion      string
	Volver      string
	Titulo      string
	Descripcion string
	// Criterios son los criterios de aceptacion, uno por linea.
	Criterios   string
	Prioridades []Opcion
}

// VistaBacklog es el backlog de un proyecto con el formulario para crear una historia.
type VistaBacklog struct {
	Items []FilaItem
	Alta  FormItem
}

// Estilos compartidos de Tailwind (spec T-024, RN-T024-2): cada pieza se define una sola vez y las pantallas la usan. Los colores
// de texto y de fondo de cada par se eligieron con un contraste de al menos 4,5:1 (RN-T024-4): texto -800 o -900 sobre fondo -50 o
// -100, y blanco sobre el indigo -700.
const (
	claseFoco = "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"

	claseCampo    = "mt-1 block w-full rounded-lg border border-slate-500 bg-white px-3 py-2 text-sm shadow-sm focus-visible:border-indigo-600 focus-visible:outline-2 focus-visible:outline-indigo-600"
	claseEtiqueta = "block text-sm font-medium text-slate-700"

	claseBoton                = "inline-flex items-center rounded-lg bg-indigo-700 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-indigo-800 " + claseFoco
	claseBotonChico           = "inline-flex items-center rounded-md bg-indigo-700 px-3 py-1 text-xs font-semibold text-white hover:bg-indigo-800 " + claseFoco
	claseBotonSecundario      = "inline-flex items-center rounded-lg border border-slate-500 bg-white px-4 py-2 text-sm font-semibold text-slate-800 shadow-sm hover:bg-slate-50 " + claseFoco
	claseBotonSecundarioChico = "inline-flex items-center rounded-md border border-slate-500 bg-white px-3 py-1 text-xs font-semibold text-slate-800 hover:bg-slate-50 " + claseFoco

	claseEnlace = "font-medium text-indigo-700 underline decoration-indigo-300 underline-offset-2 hover:text-indigo-900 hover:decoration-indigo-700 " + claseFoco

	claseTitulo    = "text-2xl font-semibold tracking-tight text-slate-900"
	claseSubtitulo = "text-lg font-semibold text-slate-900"

	// claseTarjeta es el bloque de informacion: borde suave, relleno y una sombra liviana.
	claseTarjeta = "rounded-xl border border-slate-200 bg-white p-5 shadow-sm"
	// claseDato es una cifra o un dato suelto dentro de una tarjeta.
	claseDato = "rounded-lg bg-slate-50 px-3 py-2 text-sm text-slate-800"

	// Las tablas van dentro de un contenedor que se desplaza en horizontal si no entran, para que la pagina nunca lo haga (RN-T024-6).
	claseTablaContenedor = "mt-4 overflow-x-auto rounded-xl border border-slate-200 bg-white shadow-sm"
	claseTabla           = "w-full border-collapse text-left text-sm"
	claseEncabezado      = "bg-slate-50 px-4 py-2.5 text-xs font-semibold uppercase tracking-wide text-slate-600"
	claseFila            = "border-t border-slate-200 align-top hover:bg-slate-50"
	claseCelda           = "px-4 py-3"
)

// Tonos de las insignias. El color nunca es el unico medio: cada insignia lleva su texto y su tono en data-tono (RN-T024-3).
const (
	tonoNeutro   = "neutro"
	tonoPrimario = "primario"
	tonoInfo     = "info"
	tonoOk       = "ok"
	tonoAtencion = "atencion"
	tonoApagado  = "apagado"
)

const claseInsigniaBase = "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold ring-1 ring-inset"

// claseDeInsignia devuelve las clases de una insignia segun su tono. Un tono que no conoce se dibuja como neutro.
func claseDeInsignia(tono string) string {
	switch tono {
	case tonoPrimario:
		return claseInsigniaBase + " bg-indigo-100 text-indigo-900 ring-indigo-300"
	case tonoInfo:
		return claseInsigniaBase + " bg-sky-100 text-sky-900 ring-sky-300"
	case tonoOk:
		return claseInsigniaBase + " bg-emerald-100 text-emerald-900 ring-emerald-300"
	case tonoAtencion:
		return claseInsigniaBase + " bg-amber-100 text-amber-900 ring-amber-300"
	case tonoApagado:
		return claseInsigniaBase + " bg-white text-slate-600 ring-slate-300"
	default:
		return claseInsigniaBase + " bg-slate-100 text-slate-800 ring-slate-300"
	}
}

// tonoDePrioridad es el tono de la prioridad MoSCoW (el texto llega con su mayuscula inicial: «Must», «Should»...).
func tonoDePrioridad(prioridad string) string {
	switch strings.ToLower(prioridad) {
	case "must":
		return tonoPrimario
	case "should":
		return tonoInfo
	case "could":
		return tonoNeutro
	case "wont":
		return tonoApagado
	}
	return tonoNeutro
}

// tonoDeEstado es el tono del estado de una historia: verde si esta completada, ambar si esta pendiente.
func tonoDeEstado(estado string) string {
	switch strings.ToLower(estado) {
	case "completado":
		return tonoOk
	case "pendiente":
		return tonoAtencion
	case "en progreso":
		return tonoInfo
	}
	return tonoNeutro
}

// iniciales son las dos primeras letras de las dos primeras palabras de un nombre («Ana Martínez» es «AM»), para el circulo de cada
// integrante. Sin nombre, un signo de pregunta.
func iniciales(nombre string) string {
	var salida []rune
	for _, palabra := range strings.Fields(nombre) {
		if len(salida) == 2 {
			break
		}
		salida = append(salida, unicode.ToUpper([]rune(palabra)[0]))
	}
	if len(salida) == 0 {
		return "?"
	}
	return string(salida)
}
