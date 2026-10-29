package httpadapter

import (
	"context"
	"net/http"
	"strconv"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// Reportes en pantalla de un proyecto y de un sprint (spec US-033). Igual que las demas pantallas, el handler lee la peticion, llama al
// caso de uso y escribe la respuesta (RN-033-9): los numeros vienen armados de los casos de uso que ya existen, y aca solo se escriben.

const textoSinDatos = "Sin datos"

func rutaDelReporteDelSprint(proyectoID, sprintID string) string {
	return rutaDeSprint(proyectoID, sprintID) + "/reporte"
}

// ConReportes arma los casos de uso de los dos reportes con los que la aplicacion ya tiene: no leen nada por su cuenta (RN-033-1).
func (c CasosDeUso) ConReportes() CasosDeUso {
	c.ReporteDelProyecto = app.NuevoConsultarReporteDelProyecto(app.FuentesDelReporteDelProyecto{
		Proyecto: c.ConsultarProyecto.Ejecutar,
		Backlog: func(ctx context.Context, proyectoID string) ([]backlog.ItemBacklog, error) {
			return c.ListarBacklog.Ejecutar(ctx, proyectoID, app.FiltroBacklog{}) // el reporte lista el backlog completo (US-007)
		},
		Sprints:     c.ListarSprints.Ejecutar,
		Anteriores:  c.ConsultarSprintsAnteriores.Ejecutar,
		Esfuerzo:    c.ConsultarEsfuerzo.Ejecutar,
		Comparacion: c.CompararEsfuerzo.Ejecutar,
		Defectos:    c.ListarDefectos.Ejecutar,
		Dashboard:   c.ConsultarDashboard.Ejecutar,
	})
	c.ReporteDelSprint = app.NuevoConsultarReporteDelSprint(app.FuentesDelReporteDelSprint{
		Proyecto: c.ConsultarProyecto.Ejecutar,
		Sprint:   c.ConsultarSprint.Ejecutar,
		Esfuerzo: c.ConsultarEsfuerzo.Ejecutar,
		Defectos: c.ListarDefectos.Ejecutar,
	})
	return c
}

// verReporte muestra /proyectos/{id}/reporte. Un proyecto que no existe es la pagina 404 (RN-033-6).
func (p *Pantallas) verReporte(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	reporte, err := p.casos.ReporteDelProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}

	vista := plantillas.VistaReporte{
		Titulo:      "Reporte del proyecto",
		Volver:      rutaDeProyecto(id),
		TextoVolver: "Volver al proyecto",
		Secciones: []plantillas.SeccionDeReporte{
			seccionDelProyecto(reporte.Proyecto, id),
			seccionDeIntegrantes(reporte),
			seccionDelBacklog(reporte.Backlog),
			seccionDeSprints(reporte.Sprints),
			seccionDeEsfuerzo(reporte.Esfuerzo),
			seccionDeDefectos(reporte.Defectos),
			seccionDeMetricas(reporte.Metricas),
		},
	}
	pagina := paginaDeProyecto(reporte.Proyecto.Proyecto, plantillas.SeccionReporte)
	pagina.Titulo = "Reporte | " + reporte.Proyecto.Proyecto.Nombre
	Responder(w, r, http.StatusOK, pagina, plantillas.ReporteDeProyecto(vista))
}

// verReporteDelSprint muestra /proyectos/{id}/sprints/{sprint}/reporte. Un sprint que no existe o es de otro proyecto es la pagina 404.
func (p *Pantallas) verReporteDelSprint(w http.ResponseWriter, r *http.Request) {
	id, sprintID := r.PathValue("id"), r.PathValue("sprint")
	reporte, err := p.casos.ReporteDelSprint.Ejecutar(r.Context(), id, sprintID)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	detalle, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}

	nombre := reporte.Sprint.Nombre()
	vista := plantillas.VistaReporte{
		Titulo:      "Reporte del " + nombre,
		Volver:      rutaDeSprint(id, sprintID),
		TextoVolver: "Volver al " + nombre,
		Secciones: []plantillas.SeccionDeReporte{
			seccionDelSprint(reporte),
			seccionDeHistoriasDelSprint(reporte.Historias),
			seccionDeMetricasDelSprint(reporte.Metricas),
			seccionDeEsfuerzoDelSprint(reporte.Esfuerzo),
			seccionDeDefectosDelSprint(reporte.Defectos),
		},
	}
	pagina := paginaDeProyecto(detalle.Proyecto, plantillas.SeccionSprints)
	pagina.Titulo = "Reporte del " + nombre + " | " + detalle.Proyecto.Nombre
	Responder(w, r, http.StatusOK, pagina, plantillas.ReporteDeSprint(vista))
}

// vacia es una seccion sin datos: el titulo y el texto «Sin datos», siempre presente para que el reporte no cambie de forma (RN-033-2).
func vacia(id, titulo string) plantillas.SeccionDeReporte {
	return plantillas.SeccionDeReporte{ID: id, Titulo: titulo, Vacio: textoSinDatos}
}

// conError es una seccion que no se pudo armar: informa su error y el resto del reporte se muestra (RN-033-3).
func conError(id, titulo, mensaje string, err error) plantillas.SeccionDeReporte {
	return plantillas.SeccionDeReporte{ID: id, Titulo: titulo, Error: mensaje + ": " + err.Error()}
}

// ------------------------------------------------------------------ reporte del proyecto

func seccionDelProyecto(s app.SeccionDelProyecto, proyectoID string) plantillas.SeccionDeReporte {
	if s.Err != nil {
		return conError("proyecto", "Proyecto", "No se pudo calcular el estado del proyecto", s.Err)
	}
	datos := []string{"Nombre: " + s.Proyecto.Nombre}
	if s.Proyecto.Descripcion != "" {
		datos = append(datos, "Descripción: "+s.Proyecto.Descripcion)
	}
	datos = append(datos, "Inicio: "+fechaVisible(s.Proyecto.FechaInicio), "Fin: "+fechaVisible(s.Proyecto.FechaFin))
	estado := vistaDelEstado(s.Estado, proyectoID)
	return plantillas.SeccionDeReporte{ID: "proyecto", Titulo: "Proyecto", Grupos: []plantillas.GrupoDeReporte{
		{Lineas: datos},
		{Titulo: "Estado general", Lineas: []string{estado.Estado, estado.Historias, estado.StoryPoints, estado.Porcentaje, estado.Sprint}},
	}}
}

func seccionDeIntegrantes(r app.ReporteDelProyecto) plantillas.SeccionDeReporte {
	if r.IntegrantesVacios() {
		return vacia("integrantes", "Integrantes")
	}
	filas := make([][]string, len(r.Integrantes))
	for i, in := range r.Integrantes {
		filas[i] = []string{in.Nombre, textoDeRol(in.Rol), in.Email}
	}
	return plantillas.SeccionDeReporte{ID: "integrantes", Titulo: "Integrantes", Grupos: []plantillas.GrupoDeReporte{
		{Encabezados: []string{"Nombre", "Rol", "Email"}, Filas: filas},
	}}
}

func seccionDelBacklog(s app.SeccionDelBacklog) plantillas.SeccionDeReporte {
	if s.Err != nil {
		return conError("backlog", "Product Backlog", "No se pudo leer el backlog", s.Err)
	}
	if s.Vacia() {
		return vacia("backlog", "Product Backlog")
	}
	filas := make([][]string, len(s.Items))
	for i, item := range s.Items {
		filas[i] = []string{"#" + strconv.Itoa(item.Numero), item.Titulo, textoDePrioridad(item.Prioridad), textoDeEstado(item.Estado), textoDePuntos(item)}
	}
	return plantillas.SeccionDeReporte{ID: "backlog", Titulo: "Product Backlog", Grupos: []plantillas.GrupoDeReporte{
		{Encabezados: []string{"N.º", "Historia", "Prioridad", "Estado", "Story Points"}, Filas: filas},
	}}
}

func seccionDeSprints(s app.SeccionDeSprints) plantillas.SeccionDeReporte {
	if s.Err != nil {
		return conError("sprints", "Sprints", "No se pudieron leer los sprints", s.Err)
	}
	if s.Vacia() {
		return vacia("sprints", "Sprints")
	}
	filas := make([][]string, len(s.Sprints))
	for i, sp := range s.Sprints {
		filas[i] = []string{
			sp.Sprint.Nombre(), sp.Sprint.SprintGoal, fechaVisible(sp.Sprint.FechaInicio) + " – " + fechaVisible(sp.Sprint.FechaFin),
			textoDeEstadoDeSprint(sp.Sprint.Estado), strconv.Itoa(sp.Planificados), strconv.Itoa(sp.Completados), strconv.Itoa(sp.VolvieronAlBacklog),
		}
	}
	return plantillas.SeccionDeReporte{ID: "sprints", Titulo: "Sprints", Grupos: []plantillas.GrupoDeReporte{
		{Encabezados: []string{"Sprint", "Objetivo", "Fechas", "Estado", "Story Points planificados", "Story Points completados", "Historias que volvieron al backlog"}, Filas: filas},
	}}
}

func seccionDeEsfuerzo(s app.SeccionDeEsfuerzo) plantillas.SeccionDeReporte {
	if s.Err != nil {
		return conError("esfuerzo", "Esfuerzo", "No se pudo leer el esfuerzo", s.Err)
	}
	if s.Vacia() {
		return vacia("esfuerzo", "Esfuerzo")
	}
	c := s.Comparacion
	porcentaje, calculable := c.Total.DesviacionPorcentual()
	filas := make([][]string, len(c.Historias))
	for i, h := range c.Historias {
		estimadas, desviacion, pct := "—", "—", "—"
		if h.Estimada {
			p, ok := h.DesviacionPorcentual()
			estimadas, desviacion, pct = textoDeHoras(h.Horas.Estimadas)+" h", textoDeDesviacion(h.Horas.Desviacion)+" h", textoDePorcentaje(p, ok)
		}
		filas[i] = []string{h.Titulo, estimadas, textoDeHoras(h.Horas.Reales) + " h", desviacion, pct}
	}
	return plantillas.SeccionDeReporte{ID: "esfuerzo", Titulo: "Esfuerzo", Grupos: []plantillas.GrupoDeReporte{
		{Lineas: []string{"Total de horas registradas: " + textoDeHoras(s.TotalHoras) + " h"}},
		{Titulo: "Estimado contra real", Lineas: []string{
			"Historias comparables: " + strconv.Itoa(c.Comparables),
			"Horas estimadas: " + textoDeHoras(c.Total.Estimadas) + " h",
			"Horas reales: " + textoDeHoras(c.Total.Reales) + " h",
			"Desviación: " + textoDeDesviacion(c.Total.Desviacion) + " h (" + textoDePorcentaje(porcentaje, calculable) + ")",
			"Horas sin comparar: " + textoDeHoras(c.HorasSinComparar) + " h",
		}, Encabezados: []string{"Historia", "Horas estimadas", "Horas reales", "Desviación", "Desviación %"}, Filas: filas},
	}}
}

// filasDeDefectos escribe la lista de defectos; con sprints dice tambien en cuales se detecto y se resolvio cada uno.
func filasDeDefectos(lista []app.DefectoListado, conSprints bool) [][]string {
	filas := make([][]string, len(lista))
	for i, d := range lista {
		fila := []string{d.Defecto.Descripcion, textoDeSeveridad(d.Defecto.Severidad), textoDeEstadoDeDefecto(d.Defecto.Estado), "#" + strconv.Itoa(d.HistoriaNumero) + " " + d.HistoriaTitulo}
		if conSprints {
			fila = append(fila, textoDeSprintDeDefecto(d.SprintDeteccion), textoDeSprintDeDefecto(d.SprintResolucion))
		}
		filas[i] = fila
	}
	return filas
}

func seccionDeDefectos(s app.SeccionDeDefectos) plantillas.SeccionDeReporte {
	if s.Err != nil {
		return conError("defectos", "Defectos", "No se pudieron leer los defectos", s.Err)
	}
	if s.Vacia() {
		return vacia("defectos", "Defectos")
	}
	totales := vistaDeLosDefectos(s.Totales)
	return plantillas.SeccionDeReporte{ID: "defectos", Titulo: "Defectos", Grupos: []plantillas.GrupoDeReporte{
		{Lineas: []string{totales.Total, totales.PorSeveridad, totales.PorEstado}},
		{Titulo: "Lista", Encabezados: []string{"Descripción", "Severidad", "Estado", "Historia", "Sprint de detección", "Sprint de resolución"}, Filas: filasDeDefectos(s.Lista, true)},
	}}
}

func seccionDeMetricas(s app.SeccionDeMetricas) plantillas.SeccionDeReporte {
	if s.Err != nil {
		return conError("metricas", "Métricas", "No se pudieron calcular las métricas", s.Err)
	}
	if s.Vacia() {
		return vacia("metricas", "Métricas")
	}
	velocidad := vistaDeLaVelocidad(s.Velocidad)
	secciones := []plantillas.GrupoDeReporte{{Lineas: []string{velocidad.Velocidad}}}
	if r := vistaDelSprintDeReferencia(s.Referencia); s.Referencia != nil {
		lineas := []string{"Sprint de referencia: " + r.Titulo}
		if r.Error != "" {
			lineas = append(lineas, r.Error)
		}
		secciones = append(secciones, plantillas.GrupoDeReporte{Lineas: append(lineas, r.Lineas...)})
	}
	if len(velocidad.Filas) > 0 {
		filas := make([][]string, len(velocidad.Filas))
		for i, f := range velocidad.Filas {
			filas[i] = []string{f.Nombre, f.Planificados, f.Completados, f.Detectados, f.Resueltos}
		}
		secciones = append(secciones, plantillas.GrupoDeReporte{
			Titulo:      "Sprints cerrados",
			Encabezados: []string{"Sprint", "Story Points planificados", "Story Points completados", "Defectos detectados", "Defectos resueltos"},
			Filas:       filas,
		})
	}
	return plantillas.SeccionDeReporte{ID: "metricas", Titulo: "Métricas", Grupos: secciones}
}

// ------------------------------------------------------------------ reporte de un sprint

func seccionDelSprint(r app.ReporteDelSprint) plantillas.SeccionDeReporte {
	estado := "Estado: " + textoDeEstadoDeSprint(r.Sprint.Estado)
	if r.Sprint.Estado == sprint.EstadoAbierto {
		estado += " (en curso)"
	}
	return plantillas.SeccionDeReporte{ID: "sprint", Titulo: "Sprint", Grupos: []plantillas.GrupoDeReporte{
		{Lineas: []string{"Nombre: " + r.Sprint.Nombre(), "Objetivo: " + r.Sprint.SprintGoal,
			"Fechas: " + fechaVisible(r.Sprint.FechaInicio) + " – " + fechaVisible(r.Sprint.FechaFin), estado}},
	}}
}

func seccionDeHistoriasDelSprint(s app.SeccionDeHistoriasDelSprint) plantillas.SeccionDeReporte {
	if s.Vacia() {
		return vacia("historias", "Historias")
	}
	filas := make([][]string, len(s.Items))
	for i, h := range s.Items {
		resultado := "No completada"
		if h.Completada {
			resultado = "Completada"
		}
		filas[i] = []string{"#" + strconv.Itoa(h.Item.Numero), h.Item.Titulo, textoDePrioridad(h.Item.Prioridad), textoDePuntos(h.Item), resultado}
	}
	return plantillas.SeccionDeReporte{ID: "historias", Titulo: "Historias", Grupos: []plantillas.GrupoDeReporte{
		{Encabezados: []string{"N.º", "Historia", "Prioridad", "Story Points", "Resultado en este sprint"}, Filas: filas},
	}}
}

// seccionDeMetricasDelSprint escribe siempre los Story Points y el porcentaje; si las horas no se pudieron leer, lo dice en lugar de las cifras.
func seccionDeMetricasDelSprint(m app.SeccionDeMetricasDelSprint) plantillas.SeccionDeReporte {
	lineas := []string{
		"Story Points: " + strconv.Itoa(m.Completados) + " completados de " + strconv.Itoa(m.Planificados),
		"Historias completadas: " + textoDePorcentaje(m.PorcentajeHistorias, true),
	}
	s := plantillas.SeccionDeReporte{ID: "metricas", Titulo: "Métricas"}
	if m.Err != nil {
		s.Error = "No se pudieron calcular todas las métricas: " + m.Err.Error()
	} else {
		porcentaje, calculable := m.Horas.DesviacionPorcentual()
		lineas = append(lineas,
			"Horas estimadas: "+textoDeHoras(m.Horas.Estimadas)+" h",
			"Horas reales: "+textoDeHoras(m.Horas.Reales)+" h",
			"Desviación: "+textoDeDesviacion(m.Horas.Desviacion)+" h ("+textoDePorcentaje(porcentaje, calculable)+")",
		)
	}
	s.Grupos = []plantillas.GrupoDeReporte{{Lineas: lineas}}
	return s
}

func seccionDeEsfuerzoDelSprint(s app.SeccionDeEsfuerzoDelSprint) plantillas.SeccionDeReporte {
	if s.Err != nil {
		return conError("esfuerzo", "Esfuerzo registrado", "No se pudo leer el esfuerzo", s.Err)
	}
	if s.Vacia() {
		return vacia("esfuerzo", "Esfuerzo registrado")
	}
	filas := make([][]string, len(s.Registros))
	for i, r := range s.Registros {
		filas[i] = []string{fechaVisible(r.Registro.Fecha), r.Historia, r.Integrante, r.Registro.Actividad, textoDeHoras(r.Registro.Horas) + " h"}
	}
	return plantillas.SeccionDeReporte{ID: "esfuerzo", Titulo: "Esfuerzo registrado", Grupos: []plantillas.GrupoDeReporte{
		{Lineas: []string{"Total de horas del sprint: " + textoDeHoras(s.TotalHoras) + " h"}, Encabezados: []string{"Fecha", "Historia", "Integrante", "Actividad", "Horas"}, Filas: filas},
	}}
}

func seccionDeDefectosDelSprint(s app.SeccionDeDefectosDelSprint) plantillas.SeccionDeReporte {
	if s.Err != nil {
		return conError("defectos", "Defectos", "No se pudieron leer los defectos", s.Err)
	}
	if s.Vacia() {
		return vacia("defectos", "Defectos")
	}
	encabezados := []string{"Descripción", "Severidad", "Estado", "Historia"}
	return plantillas.SeccionDeReporte{ID: "defectos", Titulo: "Defectos", Grupos: []plantillas.GrupoDeReporte{
		{Titulo: "Detectados en el sprint", Lineas: []string{"Detectados: " + strconv.Itoa(len(s.Detectados))}, Encabezados: encabezados, Filas: filasDeDefectos(s.Detectados, false)},
		{Titulo: "Resueltos en el sprint", Lineas: []string{"Resueltos: " + strconv.Itoa(len(s.Resueltos))}, Encabezados: encabezados, Filas: filasDeDefectos(s.Resueltos, false)},
	}}
}
