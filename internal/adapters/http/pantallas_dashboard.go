package httpadapter

import (
	"math"
	"net/http"
	"strconv"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// Dashboard de un proyecto (spec US-030). El handler pide el dashboard al caso de uso y lo escribe: los numeros vienen de metricas,
// y aca solo se arman los textos y las medidas de los dos graficos SVG (RN-030-13).

// verDashboard muestra /proyectos/{id}/dashboard. Un proyecto que no existe es la pagina 404 (RN-030-10); una tarjeta que no se pudo
// calcular informa su error y las demas se muestran (RN-030-9).
func (p *Pantallas) verDashboard(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	d, err := p.casos.ConsultarDashboard.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}

	vista := plantillas.VistaDashboard{
		Sprint:      vistaDelSprintDeReferencia(d.Referencia),
		Velocidad:   vistaDeLaVelocidad(d.Velocidad),
		Defectos:    vistaDeLosDefectos(d.Defectos),
		StoryPoints: graficoDeStoryPoints(d.Graficos, d.Velocidad.Err),
		Porcentaje:  graficoDePorcentaje(d.Graficos.Completadas, d.Referencia),
	}
	if d.ErrEstado != nil {
		vista.Estado = vistaDelEstadoConError(d.ErrEstado)
	} else {
		vista.Estado = vistaDelEstado(d.Estado, id)
	}

	pagina := paginaDeProyecto(d.Estado.Proyecto, plantillas.SeccionDashboard)
	pagina.Titulo = "Dashboard | " + d.Estado.Proyecto.Nombre
	Responder(w, r, http.StatusOK, pagina, plantillas.DashboardDeProyecto(vista))
}

func vistaDelSprintDeReferencia(t *app.TarjetaDeSprint) plantillas.VistaTarjetaSprint {
	if t == nil {
		return plantillas.VistaTarjetaSprint{SinSprints: "Sin sprints"}
	}
	vista := plantillas.VistaTarjetaSprint{Titulo: t.Sprint.Nombre() + " (" + textoDeEstadoDeSprint(t.Sprint.Estado) + ")"}
	if t.Err != nil {
		vista.Error = "No se pudieron calcular los números del sprint: " + t.Err.Error()
		return vista
	}
	porcentaje, calculable := t.Horas.DesviacionPorcentual()
	vista.Lineas = []string{
		"Story Points: " + strconv.Itoa(t.Completados) + " completados de " + strconv.Itoa(t.Planificados),
		"Historias completadas: " + textoDePorcentaje(t.PorcentajeHistorias, true),
		"Horas estimadas: " + textoDeHoras(t.Horas.Estimadas) + " h",
		"Horas reales: " + textoDeHoras(t.Horas.Reales) + " h",
		"Desviación: " + textoDeDesviacion(t.Horas.Desviacion) + " h (" + textoDePorcentaje(porcentaje, calculable) + ")",
	}
	return vista
}

func vistaDeLaVelocidad(t app.TarjetaDeVelocidad) plantillas.VistaTarjetaVelocidad {
	if t.Err != nil {
		return plantillas.VistaTarjetaVelocidad{Error: "No se pudo calcular la velocidad del equipo: " + t.Err.Error()}
	}
	if len(t.Cerrados) == 0 {
		return plantillas.VistaTarjetaVelocidad{Velocidad: "Sin sprints cerrados"}
	}
	vista := plantillas.VistaTarjetaVelocidad{Velocidad: "Velocidad del equipo: " + textoDeVelocidad(t.Velocidad)}
	for _, c := range t.Cerrados {
		vista.Filas = append(vista.Filas, plantillas.FilaSprintCerrado{
			Nombre:       c.Sprint.Nombre(),
			Planificados: strconv.Itoa(c.Planificados),
			Completados:  strconv.Itoa(c.Completados),
			Detectados:   strconv.Itoa(c.Defectos.Detectados),
			Resueltos:    strconv.Itoa(c.Defectos.Resueltos),
		})
	}
	return vista
}

func vistaDeLosDefectos(t app.TarjetaDeDefectos) plantillas.VistaTarjetaDefectos {
	if t.Err != nil {
		return plantillas.VistaTarjetaDefectos{Error: "No se pudieron contar los defectos: " + t.Err.Error()}
	}
	if t.Total == 0 {
		return plantillas.VistaTarjetaDefectos{Total: "Sin defectos"}
	}
	s, e := t.PorSeveridad, t.PorEstado
	return plantillas.VistaTarjetaDefectos{
		Total: "Defectos: " + strconv.Itoa(t.Total),
		PorSeveridad: "Críticos: " + strconv.Itoa(s.Critica) + " · Altos: " + strconv.Itoa(s.Alta) +
			" · Medios: " + strconv.Itoa(s.Media) + " · Bajos: " + strconv.Itoa(s.Baja),
		PorEstado: "Abiertos: " + strconv.Itoa(e.Abierto) + " · En corrección: " + strconv.Itoa(e.EnCorreccion) +
			" · Resueltos: " + strconv.Itoa(e.Resuelto) + " · Descartados: " + strconv.Itoa(e.Descartado),
	}
}

// Medidas del grafico de Story Points, en unidades del viewBox. Las barras crecen desde la linea base hacia arriba y miden lo que dice
// el caso de uso (0 a 100, sobre la barra mas alta); arriba queda lugar para el valor y abajo, para el nombre del sprint y la leyenda.
const (
	margenIzquierdo   = 10
	pasoDelGrupo      = 70
	anchoDeBarra      = 25
	separacionBarras  = 4
	lineaBase         = 120
	anchoMinimo       = 200
	altoDelGrafico    = 165
	alturaDeLaLeyenda = 148
)

// graficoDeStoryPoints arma el SVG de Story Points planificados y completados por sprint cerrado, en orden cronologico. Si la tarjeta de
// velocidad fallo, el grafico sigue su suerte (RN-030-9).
func graficoDeStoryPoints(g app.GraficosDelDashboard, errVelocidad error) plantillas.VistaGraficoDeBarras {
	vista := plantillas.VistaGraficoDeBarras{Vacio: "Sin sprints cerrados"}
	if errVelocidad != nil || g.ErrStoryPoints != nil {
		vista.Error = "No se pudo dibujar el gráfico de Story Points"
		return vista
	}
	if len(g.StoryPoints) == 0 {
		return vista
	}

	vista.Titulo = "Story Points planificados y completados por sprint cerrado"
	ancho := max(margenIzquierdo+len(g.StoryPoints)*pasoDelGrupo, anchoMinimo)
	vista.Ancho, vista.Alto = strconv.Itoa(ancho), strconv.Itoa(altoDelGrafico)
	descripcion := ""
	for i, b := range g.StoryPoints {
		numero := strconv.Itoa(b.Sprint)
		x := margenIzquierdo + i*pasoDelGrupo
		for _, barra := range []struct {
			serie  string
			valor  float64
			altura float64
			dx     int
			clase  string
		}{
			{"planificados", b.Planificados, b.AlturaPlanificados, 0, "fill-slate-500"},
			{"completados", b.Completados, b.AlturaCompletados, anchoDeBarra + separacionBarras, "fill-indigo-700"},
		} {
			izquierda := x + barra.dx
			arriba := lineaBase - barra.altura
			vista.Barras = append(vista.Barras, plantillas.BarraSVG{
				Sprint: numero, Serie: barra.serie, X: strconv.Itoa(izquierda), Y: numeroSVG(arriba),
				Ancho: strconv.Itoa(anchoDeBarra), Alto: numeroSVG(barra.altura), Clase: barra.clase,
			})
			vista.Etiquetas = append(vista.Etiquetas, plantillas.TextoSVG{
				X: numeroSVG(float64(izquierda) + anchoDeBarra/2.0), Y: numeroSVG(arriba - 3), Texto: numeroSVG(barra.valor), Clase: "fill-slate-900 text-[9px]",
			})
		}
		vista.Etiquetas = append(vista.Etiquetas, plantillas.TextoSVG{
			X: numeroSVG(float64(x) + (2*anchoDeBarra+separacionBarras)/2.0), Y: strconv.Itoa(lineaBase + 14), Texto: "Sprint " + numero, Clase: "fill-slate-800 text-[10px]",
		})
		if i > 0 {
			descripcion += ". "
		}
		descripcion += "Sprint " + numero + ": " + numeroSVG(b.Planificados) + " planificados, " + numeroSVG(b.Completados) + " completados"
	}
	vista.Descripcion = descripcion + "."
	vista.Leyenda = []plantillas.LeyendaSVG{
		{X: strconv.Itoa(margenIzquierdo), TextoX: strconv.Itoa(margenIzquierdo + 14), Y: strconv.Itoa(alturaDeLaLeyenda), Clase: "fill-slate-500", Nombre: "Planificados"},
		{X: strconv.Itoa(margenIzquierdo + 90), TextoX: strconv.Itoa(margenIzquierdo + 104), Y: strconv.Itoa(alturaDeLaLeyenda), Clase: "fill-indigo-700", Nombre: "Completados"},
	}
	return vista
}

// graficoDePorcentaje arma la barra del porcentaje de historias completadas del sprint de referencia, sobre una escala de 0 a 100
// (RN-030-15). Si no hay sprint de referencia no se dibuja nada; si su tarjeta fallo, el grafico informa el error.
func graficoDePorcentaje(b *app.BarraDePorcentaje, referencia *app.TarjetaDeSprint) plantillas.VistaGraficoDePorcentaje {
	if referencia == nil {
		return plantillas.VistaGraficoDePorcentaje{Vacio: "Sin sprints"}
	}
	if referencia.Err != nil || b == nil {
		return plantillas.VistaGraficoDePorcentaje{Error: "No se pudo dibujar el gráfico de historias completadas"}
	}
	nombre := "Sprint " + strconv.Itoa(b.Sprint)
	return plantillas.VistaGraficoDePorcentaje{
		Dibujar:     true,
		Titulo:      "Historias completadas del " + nombre,
		Descripcion: nombre + ": " + textoDePorcentaje(b.Porcentaje, true) + " de las historias completadas",
		AnchoBarra:  numeroSVG(b.Porcentaje),
		Valor:       textoDePorcentaje(b.Porcentaje, true),
	}
}

// numeroSVG escribe un numero para el SVG: punto decimal, sin ceros de mas y redondeado a 2 decimales para que no queden restos de coma flotante.
func numeroSVG(n float64) string { return strconv.FormatFloat(math.Round(n*100)/100, 'f', -1, 64) }
