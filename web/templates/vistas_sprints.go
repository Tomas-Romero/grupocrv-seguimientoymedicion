package plantillas

// Datos que dibujan las pantallas de sprints (spec T-015). Igual que los de T-011, son datos
// ya preparados: las fechas vienen formateadas y las opciones armadas.

// FilaSprint es un sprint en la lista, con sus totales.
type FilaSprint struct {
	Nombre       string
	Goal         string
	Inicio       string
	Fin          string
	Estado       string
	Historias    int
	Planificados int
	Completados  int
	Detalle      string
}

// FormSprint es el formulario de alta de un sprint. Las fechas van como AAAA-MM-DD.
type FormSprint struct {
	Accion      string
	SprintGoal  string
	FechaInicio string
	FechaFin    string
}

// VistaSprints es la lista de sprints de un proyecto con el formulario para crear uno.
type VistaSprints struct {
	Sprints []FilaSprint
	Alta    FormSprint
}

// FilaHistoriaDeSprint es una historia asignada a un sprint.
type FilaHistoriaDeSprint struct {
	Numero     int
	Titulo     string
	Prioridad  string
	Puntos     string
	Estado     string
	Completada bool
	// Accion es la URL de marcar o desmarcar; TextoAccion, lo que dice el boton.
	Accion      string
	TextoAccion string
}

// VistaSprint es el detalle de un sprint.
type VistaSprint struct {
	Nombre       string
	Goal         string
	Inicio       string
	Fin          string
	Estado       string
	Abierto      bool
	Planificados int
	Completados  int
	Historias    []FilaHistoriaDeSprint
	// AccionAsignar es la URL del formulario de asignacion y Disponibles, sus opciones. Con el
	// sprint cerrado no se muestran.
	AccionAsignar string
	Disponibles   []Opcion
	Volver        string
}
