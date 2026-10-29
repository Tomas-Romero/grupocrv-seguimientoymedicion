package plantillas

// Datos que dibuja el dashboard de un proyecto (spec US-030). Son datos ya preparados: los textos, los numeros y hasta las medidas de
// los graficos vienen armados por el adaptador HTTP a partir de lo que calcula el caso de uso. La plantilla solo dibuja (RN-030-13).

// VistaDashboard es la pagina /proyectos/{id}/dashboard.
type VistaDashboard struct {
	// Estado es la tarjeta de estado general de US-004.
	Estado      VistaEstadoProyecto
	Sprint      VistaTarjetaSprint
	Velocidad   VistaTarjetaVelocidad
	Defectos    VistaTarjetaDefectos
	StoryPoints VistaGraficoDeBarras
	Porcentaje  VistaGraficoDePorcentaje
}

// VistaTarjetaSprint es la tarjeta del sprint de referencia. Con Error no esta vacio, informa el error y no las lineas.
type VistaTarjetaSprint struct {
	Titulo string
	// SinSprints es el texto cuando el proyecto no tiene sprints; vacio si hay.
	SinSprints string
	Lineas     []string
	Error      string
}

// VistaTarjetaVelocidad es la velocidad del equipo y la tabla de sprints cerrados.
type VistaTarjetaVelocidad struct {
	// Velocidad es «Velocidad del equipo: 6,5» o «Sin sprints cerrados».
	Velocidad string
	Filas     []FilaSprintCerrado
	Error     string
}

// FilaSprintCerrado es un sprint cerrado de la tabla.
type FilaSprintCerrado struct {
	Nombre       string
	Planificados string
	Completados  string
	Detectados   string
	Resueltos    string
}

// VistaTarjetaDefectos cuenta los defectos del proyecto; con Total vacio, se muestra «Sin defectos».
type VistaTarjetaDefectos struct {
	Total        string
	PorSeveridad string
	PorEstado    string
	Error        string
}

// VistaGraficoDeBarras es el grafico de Story Points por sprint cerrado: un SVG con sus barras ya medidas. Sin Barras, se muestra Vacio
// (o Error) y no se dibuja nada.
type VistaGraficoDeBarras struct {
	Titulo      string
	Descripcion string
	Ancho       string
	Alto        string
	Barras      []BarraSVG
	Etiquetas   []TextoSVG
	Leyenda     []LeyendaSVG
	Vacio       string
	Error       string
}

// BarraSVG es un rectangulo del grafico: Sprint y Serie lo identifican («planificados» o «completados»); sus medidas son numeros ya
// formateados.
type BarraSVG struct {
	Sprint string
	Serie  string
	X      string
	Y      string
	Ancho  string
	Alto   string
	Clase  string
}

// TextoSVG es un texto del grafico (el valor de una barra o el nombre de un sprint).
type TextoSVG struct {
	X     string
	Y     string
	Texto string
	Clase string
}

// LeyendaSVG es un cuadrado de color con su nombre: la leyenda dice que serie es cada una, para que nada dependa del color.
type LeyendaSVG struct {
	X      string
	TextoX string
	Y      string
	Clase  string
	Nombre string
}

// VistaGraficoDePorcentaje es el grafico del porcentaje de historias completadas del sprint de referencia: una barra de 0 a 100.
type VistaGraficoDePorcentaje struct {
	Titulo      string
	Descripcion string
	// Ancho de la barra: es el porcentaje (la escala del grafico es 0 a 100).
	AnchoBarra string
	Valor      string
	Vacio      string
	Error      string
	Dibujar    bool
}
