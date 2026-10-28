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
}

// VistaDefectos es la lista de defectos de un proyecto.
type VistaDefectos struct {
	Defectos []FilaDefecto
}
