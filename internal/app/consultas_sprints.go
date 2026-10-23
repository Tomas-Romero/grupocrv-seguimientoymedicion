package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Lecturas de las pantallas de sprints (spec T-015). Igual que las de T-011, los
// repositorios devuelven los datos sin ordenar y el orden lo pone el caso de uso.

// HistoriaAsignada es una historia que se asigno a un sprint, con el resultado que quedo en la
// fila de la relacion entre ambos (RN-009-13).
type HistoriaAsignada struct {
	Item      backlog.ItemBacklog
	Resultado sprint.Resultado
}

// Completada dice si la historia quedo completada en ese sprint. En un sprint cerrado manda el
// resultado de la fila: una historia que no se completo y volvio al backlog sigue siendo "no
// completada" en ese sprint aunque despues se complete en otro. Mientras el sprint esta abierto
// la fila sigue pendiente y manda el estado de la historia, porque marcarla (US-010) no cambia la
// fila hasta el cierre.
func (h HistoriaAsignada) Completada() bool {
	switch h.Resultado {
	case sprint.ResultadoCompletado:
		return true
	case sprint.ResultadoNoCompletado:
		return false
	default:
		return h.Item.Estado == backlog.EstadoCompletado
	}
}

// SprintConItems es un sprint con las historias que alguna vez se le asignaron.
type SprintConItems struct {
	Sprint sprint.Sprint
	Items  []HistoriaAsignada
}

// resumenDeItems calcula los Story Points planificados y completados con metricas (US-025).
// Una historia sin estimar cuenta como 0 (RN-T015-8).
func resumenDeItems(items []HistoriaAsignada) (metricas.ResumenSprint, error) {
	datos := make([]metricas.ItemDelSprint, len(items))
	for i, h := range items {
		puntos, _ := h.Item.StoryPoints.Valor()
		datos[i] = metricas.ItemDelSprint{StoryPoints: puntos, Completado: h.Completada()}
	}
	return metricas.CalcularResumenSprint(datos)
}

// RepositorioListarSprints es lo que ListarSprints necesita de la persistencia.
type RepositorioListarSprints interface {
	// ExisteProyecto dice si el proyecto existe (es la lectura de US-005).
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)
	// ListarSprintsConItems devuelve los sprints del proyecto, cada uno con sus historias,
	// en cualquier orden.
	ListarSprintsConItems(ctx context.Context, proyectoID string) ([]SprintConItems, error)
}

// ResumenDeSprint es un sprint en la lista, con sus totales.
type ResumenDeSprint struct {
	Sprint       sprint.Sprint
	Historias    int
	Planificados int
	Completados  int
}

// ListarSprints es la lectura de la lista de sprints de un proyecto.
type ListarSprints struct {
	repo RepositorioListarSprints
}

// NuevoListarSprints arma el caso de uso.
func NuevoListarSprints(repo RepositorioListarSprints) *ListarSprints {
	return &ListarSprints{repo: repo}
}

// Ejecutar devuelve los sprints por numero descendente (RN-T015-3), con la cantidad de
// historias y los Story Points planificados y completados. Sin sprints devuelve una lista
// vacia; un proyecto que no existe es ErrProyectoInexistente.
func (c *ListarSprints) Ejecutar(ctx context.Context, proyectoID string) ([]ResumenDeSprint, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return nil, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}
	sprints, err := c.repo.ListarSprintsConItems(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar los sprints del proyecto %s: %w", proyectoID, err)
	}

	resumenes := make([]ResumenDeSprint, 0, len(sprints))
	for _, s := range sprints {
		totales, err := resumenDeItems(s.Items)
		if err != nil {
			return nil, fmt.Errorf("calcular los Story Points del %s: %w", s.Sprint.Nombre(), err)
		}
		resumenes = append(resumenes, ResumenDeSprint{
			Sprint:       s.Sprint,
			Historias:    len(s.Items),
			Planificados: totales.Planificados,
			Completados:  totales.Completados,
		})
	}
	sort.SliceStable(resumenes, func(i, j int) bool { return resumenes[i].Sprint.Numero > resumenes[j].Sprint.Numero })
	return resumenes, nil
}

// RepositorioConsultarSprint es lo que ConsultarSprint necesita de la persistencia.
type RepositorioConsultarSprint interface {
	// ObtenerSprint devuelve el sprint y si existe (es la lectura de US-009).
	ObtenerSprint(ctx context.Context, sprintID string) (sprint.Sprint, bool, error)
	// ItemsDelSprint devuelve las historias que se asignaron al sprint, cada una con el resultado
	// de su fila, en cualquier orden.
	ItemsDelSprint(ctx context.Context, sprintID string) ([]HistoriaAsignada, error)
	// ListarItemsDisponibles devuelve las historias del proyecto que se pueden asignar: no
	// estan completadas y no estan en ningun sprint abierto. Incluye las que todavia no
	// estan listas (sin estimar o sin criterios): lo decide el dominio al asignar.
	ListarItemsDisponibles(ctx context.Context, proyectoID string) ([]backlog.ItemBacklog, error)
}

// HistoriaDelSprint es una historia asignada a un sprint.
type HistoriaDelSprint struct {
	Item       backlog.ItemBacklog
	Completada bool
	// Resultado es el de la fila de la historia en este sprint: pendiente mientras esta abierto.
	Resultado sprint.Resultado
}

// DetalleDeSprint es un sprint con sus historias, sus totales y las historias que se le
// pueden asignar.
type DetalleDeSprint struct {
	Sprint       sprint.Sprint
	Historias    []HistoriaDelSprint
	Planificados int
	Completados  int
	Disponibles  []backlog.ItemBacklog
}

// ConsultarSprint es la lectura del detalle de un sprint.
type ConsultarSprint struct {
	repo RepositorioConsultarSprint
}

// NuevoConsultarSprint arma el caso de uso.
func NuevoConsultarSprint(repo RepositorioConsultarSprint) *ConsultarSprint {
	return &ConsultarSprint{repo: repo}
}

// Ejecutar devuelve el detalle del sprint solo si es del proyecto (RN-T015-5): el sprint de
// otro proyecto, o que no existe, es ErrSprintInexistente. Las historias y las disponibles
// salen por prioridad y numero (RN-T015-4).
func (c *ConsultarSprint) Ejecutar(ctx context.Context, proyectoID, sprintID string) (DetalleDeSprint, error) {
	s, existe, err := c.repo.ObtenerSprint(ctx, sprintID)
	if err != nil {
		return DetalleDeSprint{}, fmt.Errorf("consultar el sprint %s: %w", sprintID, err)
	}
	if !existe || s.ProyectoID != proyectoID {
		return DetalleDeSprint{}, fmt.Errorf("%w (id %s)", ErrSprintInexistente, sprintID)
	}

	items, err := c.repo.ItemsDelSprint(ctx, sprintID)
	if err != nil {
		return DetalleDeSprint{}, fmt.Errorf("listar las historias del sprint %s: %w", sprintID, err)
	}
	sort.SliceStable(items, func(i, j int) bool { return antesPorPrioridad(items[i].Item, items[j].Item) })
	totales, err := resumenDeItems(items)
	if err != nil {
		return DetalleDeSprint{}, fmt.Errorf("calcular los Story Points del %s: %w", s.Nombre(), err)
	}
	historias := make([]HistoriaDelSprint, len(items))
	for i, h := range items {
		historias[i] = HistoriaDelSprint{Item: h.Item, Completada: h.Completada(), Resultado: h.Resultado}
	}

	disponibles, err := c.repo.ListarItemsDisponibles(ctx, proyectoID)
	if err != nil {
		return DetalleDeSprint{}, fmt.Errorf("listar las historias disponibles del proyecto %s: %w", proyectoID, err)
	}
	ordenarPorPrioridad(disponibles)

	return DetalleDeSprint{
		Sprint:       s,
		Historias:    historias,
		Planificados: totales.Planificados,
		Completados:  totales.Completados,
		Disponibles:  disponibles,
	}, nil
}
