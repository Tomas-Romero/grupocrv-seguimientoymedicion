package httpadapter

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// Handlers de las pantallas de sprints (spec T-015). Igual que los de T-011, solo leen la
// peticion, llaman a un caso de uso y escriben la respuesta.

func rutaDeSprint(proyectoID, sprintID string) string {
	return rutaDeProyecto(proyectoID) + "/sprints/" + sprintID
}

func textoDeEstadoDeSprint(e sprint.Estado) string {
	if e == sprint.EstadoCerrado {
		return "cerrado"
	}
	return "abierto"
}

func (p *Pantallas) listaDeSprints(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detalle, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	sprints, err := p.casos.ListarSprints.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	filas := make([]plantillas.FilaSprint, len(sprints))
	for i, s := range sprints {
		filas[i] = plantillas.FilaSprint{
			Nombre:       s.Sprint.Nombre(),
			Goal:         s.Sprint.SprintGoal,
			Inicio:       fechaVisible(s.Sprint.FechaInicio),
			Fin:          fechaVisible(s.Sprint.FechaFin),
			Estado:       textoDeEstadoDeSprint(s.Sprint.Estado),
			Historias:    s.Historias,
			Planificados: s.Planificados,
			Completados:  s.Completados,
			Detalle:      rutaDeSprint(id, s.Sprint.ID),
		}
	}
	vista := plantillas.VistaSprints{
		Sprints:    filas,
		Alta:       plantillas.FormSprint{Accion: rutaDeProyecto(id) + "/sprints"},
		Anteriores: rutaDeProyecto(id) + "/sprints/anteriores",
	}
	Responder(w, r, http.StatusOK, paginaDeProyecto(detalle.Proyecto, plantillas.SeccionSprints), plantillas.ListaDeSprints(vista))
}

// sprintsAnteriores es el historial de sprints cerrados del proyecto, con su resultado (US-012).
func (p *Pantallas) sprintsAnteriores(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detalle, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	anteriores, err := p.casos.ConsultarSprintsAnteriores.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	filas := make([]plantillas.FilaSprintAnterior, len(anteriores.Sprints))
	for i, s := range anteriores.Sprints {
		filas[i] = plantillas.FilaSprintAnterior{
			Nombre:                 s.Sprint.Nombre(),
			Goal:                   s.Sprint.SprintGoal,
			Inicio:                 fechaVisible(s.Sprint.FechaInicio),
			Fin:                    fechaVisible(s.Sprint.FechaFin),
			Planificados:           s.Planificados,
			Completados:            s.Completados,
			HistoriasCompletadas:   s.HistoriasCompletadas,
			HistoriasNoCompletadas: s.HistoriasNoCompletadas,
			Detalle:                rutaDeSprint(id, s.Sprint.ID),
		}
	}
	vista := plantillas.VistaSprintsAnteriores{
		Sprints:   filas,
		Velocidad: textoDeVelocidad(anteriores.Velocidad),
		Volver:    rutaDeProyecto(id) + "/sprints",
	}
	pagina := paginaDeProyecto(detalle.Proyecto, plantillas.SeccionSprints)
	pagina.Titulo = "Sprints anteriores | " + detalle.Proyecto.Nombre
	Responder(w, r, http.StatusOK, pagina, plantillas.SprintsAnteriores(vista))
}

// textoDeVelocidad escribe la velocidad como la lee una persona: sin ceros de mas y con coma decimal.
func textoDeVelocidad(velocidad float64) string {
	return strings.Replace(strconv.FormatFloat(velocidad, 'f', -1, 64), ".", ",", 1)
}

func (p *Pantallas) crearSprint(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	f, err := leerFormulario(r)
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	inicio, fin := f.Fecha("fecha_inicio"), f.Fecha("fecha_fin")
	if err := f.Err(); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	creado, err := p.casos.CrearSprint.Ejecutar(r.Context(), id, r.PostForm.Get("sprint_goal"), inicio, fin)
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeSprint(id, creado.ID))
}

func (p *Pantallas) verSprint(w http.ResponseWriter, r *http.Request) {
	id, sprintID := r.PathValue("id"), r.PathValue("sprint")
	proyecto, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	detalle, err := p.casos.ConsultarSprint.Ejecutar(r.Context(), id, sprintID)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}

	abierto := detalle.Sprint.Estado != sprint.EstadoCerrado
	historias := make([]plantillas.FilaHistoriaDeSprint, len(detalle.Historias))
	for i, h := range detalle.Historias {
		accion, texto := "completar", "Marcar completada"
		if h.Completada {
			accion, texto = "reabrir", "Desmarcar"
		}
		historias[i] = plantillas.FilaHistoriaDeSprint{
			Numero:      h.Item.Numero,
			Titulo:      h.Item.Titulo,
			Prioridad:   textoDePrioridad(h.Item.Prioridad),
			Puntos:      textoDePuntos(h.Item),
			Estado:      textoDeEstado(h.Item.Estado),
			Completada:  h.Completada,
			Resultado:   textoDeResultado(h.Completada),
			Accion:      rutaDeSprint(id, sprintID) + "/historias/" + h.Item.ID + "/" + accion,
			TextoAccion: texto,
		}
	}
	devueltas := 0
	for _, h := range detalle.Historias {
		if !h.Completada {
			devueltas++
		}
	}
	disponibles := make([]plantillas.Opcion, len(detalle.Disponibles))
	for i, item := range detalle.Disponibles {
		disponibles[i] = plantillas.Opcion{Valor: item.ID, Texto: textoDeHistoriaDisponible(item)}
	}

	vista := plantillas.VistaSprint{
		Nombre:        detalle.Sprint.Nombre(),
		Goal:          detalle.Sprint.SprintGoal,
		Inicio:        fechaVisible(detalle.Sprint.FechaInicio),
		Fin:           fechaVisible(detalle.Sprint.FechaFin),
		Estado:        textoDeEstadoDeSprint(detalle.Sprint.Estado),
		Abierto:       abierto,
		Planificados:  detalle.Planificados,
		Completados:   detalle.Completados,
		Historias:     historias,
		AccionCerrar:  rutaDeSprint(id, sprintID) + "/cerrar",
		AccionAsignar: rutaDeSprint(id, sprintID) + "/historias",
		Disponibles:   disponibles,
		Volver:        rutaDeProyecto(id) + "/sprints",
		Reporte:       rutaDelReporteDelSprint(id, sprintID),
	}
	if abierto {
		vista.ConfirmaCierre = "¿Cerrar el " + vista.Nombre + "? Las historias no completadas vuelven al backlog y el sprint no se puede reabrir."
	} else {
		vista.TextoDevueltas = textoDeCantidadDeHistorias(devueltas)
	}
	pagina := paginaDeProyecto(proyecto.Proyecto, plantillas.SeccionSprints)
	pagina.Titulo = detalle.Sprint.Nombre() + " | " + proyecto.Proyecto.Nombre
	Responder(w, r, http.StatusOK, pagina, plantillas.DetalleDeSprint(vista))
}

// cerrarSprint cierra el sprint (US-011) despues de comprobar que es del proyecto (RN-T015-5): el de otro
// proyecto, o uno que no existe, es un 404 antes de llamar al caso de uso (RN-T016-1, RN-T016-6).
func (p *Pantallas) cerrarSprint(w http.ResponseWriter, r *http.Request) {
	id, sprintID := r.PathValue("id"), r.PathValue("sprint")
	if _, err := p.casos.ConsultarSprint.Ejecutar(r.Context(), id, sprintID); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	if _, err := p.casos.CerrarSprint.Ejecutar(r.Context(), sprintID); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeSprint(id, sprintID))
}

func (p *Pantallas) asignarHistoria(w http.ResponseWriter, r *http.Request) {
	id, sprintID := r.PathValue("id"), r.PathValue("sprint")
	if _, err := leerFormulario(r); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	// RN-T015-5: el sprint se alcanza solo por la ruta de su proyecto.
	if _, err := p.casos.ConsultarSprint.Ejecutar(r.Context(), id, sprintID); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	if _, err := p.casos.AsignarItem.Ejecutar(r.Context(), r.PostForm.Get("item"), sprintID); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeSprint(id, sprintID))
}

func (p *Pantallas) completarHistoria(w http.ResponseWriter, r *http.Request) {
	p.cambiarEstadoDeHistoria(w, r, func(itemID, sprintID string) error {
		_, err := p.casos.MarcarCompletado.Ejecutar(r.Context(), itemID, sprintID)
		return err
	})
}

func (p *Pantallas) reabrirHistoria(w http.ResponseWriter, r *http.Request) {
	p.cambiarEstadoDeHistoria(w, r, func(itemID, sprintID string) error {
		_, err := p.casos.DesmarcarCompletado.Ejecutar(r.Context(), itemID, sprintID)
		return err
	})
}

// cambiarEstadoDeHistoria comprueba que el sprint sea del proyecto y la historia, del sprint
// (RN-T015-5) y despues hace el cambio. Una historia que no esta en ese sprint es un 404.
func (p *Pantallas) cambiarEstadoDeHistoria(w http.ResponseWriter, r *http.Request, cambiar func(itemID, sprintID string) error) {
	id, sprintID, itemID := r.PathValue("id"), r.PathValue("sprint"), r.PathValue("item")
	detalle, err := p.casos.ConsultarSprint.Ejecutar(r.Context(), id, sprintID)
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	enElSprint := false
	for _, h := range detalle.Historias {
		if h.Item.ID == itemID {
			enElSprint = true
			break
		}
	}
	if !enElSprint {
		p.errores.ResponderError(w, r, app.ErrItemInexistente)
		return
	}
	if err := cambiar(itemID, sprintID); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeSprint(id, sprintID))
}

// textoDeResultado es lo que dice un sprint cerrado de cada historia (RN-T016-4, RN-T016-8).
func textoDeResultado(completada bool) string {
	if completada {
		return "Completada"
	}
	return "No completada · volvió al backlog"
}

// textoDeCantidadDeHistorias dice "1 historia" o "N historias"; vacio si no hay ninguna (RN-T016-5).
func textoDeCantidadDeHistorias(n int) string {
	switch n {
	case 0:
		return ""
	case 1:
		return "1 historia"
	}
	return strconv.Itoa(n) + " historias"
}

// textoDePuntos es el texto de los Story Points de una historia: el numero, o un guion si no
// esta estimada.
func textoDePuntos(item backlog.ItemBacklog) string {
	if puntos, estimado := item.StoryPoints.Valor(); estimado {
		return strconv.Itoa(puntos)
	}
	return "—"
}

// textoDeHistoriaDisponible es lo que dice la lista de asignacion: el titulo y, si todavia no
// esta lista para un sprint, que le falta (RN-T015-6). Asignarla igual se rechaza en el dominio.
func textoDeHistoriaDisponible(item backlog.ItemBacklog) string {
	var faltas []string
	if _, estimado := item.StoryPoints.Valor(); !estimado {
		faltas = append(faltas, "sin estimar")
	}
	if len(item.Criterios) == 0 {
		faltas = append(faltas, "sin criterios")
	}
	if len(faltas) == 0 {
		return item.Titulo
	}
	texto := item.Titulo + " ("
	for i, f := range faltas {
		if i > 0 {
			texto += ", "
		}
		texto += f
	}
	return texto + ")"
}
