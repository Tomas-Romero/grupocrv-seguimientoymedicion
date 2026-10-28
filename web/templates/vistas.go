package plantillas

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

// VistaProyecto es el resumen de un proyecto.
type VistaProyecto struct {
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

// Clases de Tailwind compartidas por los formularios y las tablas. El diseno es funcional
// (spec T-011, seccion 5): se mejora despues de que todo el producto funcione.
const (
	claseCampo      = "mt-1 block w-full rounded border border-slate-300 bg-white px-3 py-2 text-sm focus-visible:outline-2 focus-visible:outline-indigo-600"
	claseEtiqueta   = "block text-sm font-medium text-slate-700"
	claseBoton      = "rounded bg-indigo-700 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-800 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
	claseBotonChico = "rounded bg-indigo-700 px-3 py-1 text-xs font-semibold text-white hover:bg-indigo-800 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
	claseEnlace     = "text-indigo-700 underline hover:text-indigo-900 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
	claseTitulo     = "text-2xl font-semibold text-slate-900"
)
