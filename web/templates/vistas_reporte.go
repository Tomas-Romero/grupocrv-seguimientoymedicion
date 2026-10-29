package plantillas

// Datos que dibujan los reportes en pantalla de un proyecto y de un sprint (spec US-033). Son datos ya preparados: los textos y las
// tablas vienen armados por el adaptador HTTP a partir de lo que devuelven los casos de uso, y la plantilla solo los dibuja.

// VistaReporte es la pagina de un reporte: su titulo y sus secciones, siempre las mismas y en el mismo orden (RN-033-2).
type VistaReporte struct {
	// Titulo es «Reporte del proyecto» o «Reporte del Sprint 3».
	Titulo    string
	Secciones []SeccionDeReporte
	// Volver es el enlace a la pantalla de origen y TextoVolver, lo que dice.
	Volver      string
	TextoVolver string
}

// SeccionDeReporte es una seccion con su identificador (#proyecto, #backlog...) para poder enlazarla. Si Error no esta vacio, la
// seccion informa el error y no las cifras que no se pudieron armar; si Vacio no esta vacio («Sin datos»), no hay grupos que mostrar.
type SeccionDeReporte struct {
	ID     string
	Titulo string
	Error  string
	Vacio  string
	Grupos []GrupoDeReporte
}

// GrupoDeReporte es un bloque de una seccion: un subtitulo opcional, lineas de texto y una tabla opcional.
type GrupoDeReporte struct {
	Titulo      string
	Lineas      []string
	Encabezados []string
	Filas       [][]string
}
