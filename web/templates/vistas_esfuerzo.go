package plantillas

// Datos que dibujan las pantallas de esfuerzo (spec T-019). Como las demas vistas, son datos ya
// preparados: las horas, las fechas y los porcentajes vienen escritos como los lee una persona.

// FormEsfuerzo es el formulario «Registrar horas». Fecha va como AAAA-MM-DD, el formato de un
// <input type="date">; los avisos no estan vacios cuando no hay integrantes o historias para elegir.
type FormEsfuerzo struct {
	Accion           string
	Integrantes      []Opcion
	Historias        []Opcion
	Fecha            string
	AvisoIntegrantes string
	AvisoHistorias   string
}

// FiltrosEsfuerzo es el formulario GET de filtros: cada lista empieza con «Todas» o «Todos».
type FiltrosEsfuerzo struct {
	Accion      string
	Historias   []Opcion
	Integrantes []Opcion
	Sprints     []Opcion
}

// FilaEsfuerzo es un registro de la tabla.
type FilaEsfuerzo struct {
	Fecha      string
	Historia   string
	Integrante string
	Actividad  string
	Horas      string
	Sprint     string
}

// FilaTotalEsfuerzo es una fila de los totales por historia, integrante o sprint.
type FilaTotalEsfuerzo struct {
	Nombre string
	Horas  string
}

// VistaEsfuerzo es la pagina /proyectos/{id}/esfuerzo.
type VistaEsfuerzo struct {
	Alta    FormEsfuerzo
	Filtros FiltrosEsfuerzo
	// Total es «Total: 7,5 h», de los registros que quedaron despues de filtrar.
	Total string
	// Vacio es el texto que reemplaza a la tabla y a los totales; vacio si hay registros.
	Vacio         string
	Registros     []FilaEsfuerzo
	PorHistoria   []FilaTotalEsfuerzo
	PorIntegrante []FilaTotalEsfuerzo
	PorSprint     []FilaTotalEsfuerzo
	Comparar      string
}

// FilaComparacion es una historia de la comparacion de lo estimado contra lo real.
type FilaComparacion struct {
	Titulo     string
	Puntos     string
	Estimadas  string
	Reales     string
	Desviacion string
	Porcentaje string
	Situacion  string
}

// VistaComparacion es la pagina /proyectos/{id}/esfuerzo/comparacion.
type VistaComparacion struct {
	Filas []FilaComparacion
	// Resumen y SinComparar estan vacios cuando no hay historias; Vacio, cuando las hay.
	Resumen     string
	SinComparar string
	Vacio       string
	Volver      string
}
