package plantillas

// Datos que dibujan las pantallas de defectos (spec T-021). Igual que los de T-011 y T-015, son
// datos ya preparados: las etiquetas de severidad y de estado vienen traducidas y las opciones
// armadas.

// FilaDefecto es un defecto en la lista. La fecha de carga no se muestra (CA-T021-2).
type FilaDefecto struct {
	Descripcion string
	// Severidad y Estado son las etiquetas de la pantalla ("Crítica", "En corrección"), no los
	// valores del dominio.
	Severidad string
	Estado    string
	// Historia es "#N título", Sprint de detección y de resolución, "Sprint N"; la resolución es
	// "—" si el defecto no tiene sprint de resolución (RN-T021-13).
	Historia         string
	SprintDeteccion  string
	SprintResolucion string
	// Nombre es el comienzo de la descripcion (hasta 40 caracteres): el defecto no tiene numero ni
	// titulo, y es lo que dice a que defecto corresponde cada control (RN-T021-17).
	Nombre string
	// Accion es la URL del cambio de estado; Acciones, los botones que ofrece el estado de la fila
	// (RN-T021-14); Sprints, las opciones del sprint de resolucion de "Resolver".
	Accion   string
	Acciones []AccionDeDefecto
	Sprints  []Opcion
}

// AccionDeDefecto es un boton de la fila: un formulario que pide un estado. Estado es el valor del
// dominio que viaja en el campo oculto; ConSprint dice si el formulario lleva la lista del sprint de
// resolucion, que solo tiene "Resolver" (RN-T021-14).
type AccionDeDefecto struct {
	Texto     string
	Estado    string
	ConSprint bool
}

// FormDefecto es el formulario de alta de un defecto. Cada lista de opciones trae como primera una
// opcion sin valor, para que la persona tenga que elegir (RN-T021-15). SinHistorias y SinSprints
// muestran el aviso de que todavia no hay nada para elegir (CL-T021-2).
type FormDefecto struct {
	Accion       string
	Severidades  []Opcion
	Historias    []Opcion
	Sprints      []Opcion
	SinHistorias bool
	SinSprints   bool
}

// FiltrosDefectos es el formulario GET con el que se filtra la lista (US-024, RN-024-11): cada lista
// empieza con la opcion para no filtrar («Todos» o «Todas») y deja elegido lo que se pidio. Activo dice
// que se pidio algun filtro, para que una lista vacia diga que ningun defecto lo cumple y no que el
// proyecto no tiene defectos.
type FiltrosDefectos struct {
	Accion      string
	Sprints     []Opcion
	Severidades []Opcion
	Activo      bool
}

// VistaDefectos es la lista de defectos de un proyecto con los filtros y el formulario para registrar uno.
type VistaDefectos struct {
	Defectos []FilaDefecto
	Filtros  FiltrosDefectos
	Alta     FormDefecto
}
