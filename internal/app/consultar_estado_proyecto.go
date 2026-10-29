package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// RepositorioEstadoGeneral es lo que ConsultarEstadoDelProyecto necesita de la persistencia (spec US-004, seccion 5). Son
// tres lecturas que ya existen en Postgres y en la base en memoria: no hay consultas nuevas.
type RepositorioEstadoGeneral interface {
	// BuscarProyecto devuelve el proyecto o ErrProyectoInexistente (es el de US-002).
	BuscarProyecto(ctx context.Context, proyectoID string) (proyecto.Proyecto, error)
	// ListarItems devuelve las historias del proyecto, en cualquier orden.
	ListarItems(ctx context.Context, proyectoID string) ([]backlog.ItemBacklog, error)
	// ListarSprintsConItems devuelve los sprints del proyecto con las historias que se les asignaron.
	ListarSprintsConItems(ctx context.Context, proyectoID string) ([]SprintConItems, error)
}

// HistoriasPorEstado cuenta las historias del proyecto por su estado actual.
type HistoriasPorEstado struct {
	Pendientes  int
	EnProgreso  int
	Completadas int
	Total       int
}

// EstadoGeneral es el estado general de un proyecto (spec US-004, seccion 3).
type EstadoGeneral struct {
	Proyecto proyecto.Proyecto
	// Estado y Dias son los de proyecto.EstadoEn con el reloj del caso de uso.
	Estado proyecto.EstadoDelProyecto
	Dias   int
	// Historias son todas las del proyecto, esten o no en un sprint (CL-004-9).
	Historias HistoriasPorEstado
	// StoryPoints son los de las historias estimadas (las sin estimar suman 0) y Completados, los de las completadas.
	StoryPoints int
	Completados int
	// Porcentaje es el de historias completadas, con 1 decimal (metricas, US-028).
	Porcentaje float64
	// SprintAbierto es nil si el proyecto no tiene un sprint abierto.
	SprintAbierto *sprint.Sprint
}

// ConsultarEstadoDelProyecto es el caso de uso de lectura de US-004.
type ConsultarEstadoDelProyecto struct {
	repo  RepositorioEstadoGeneral
	ahora func() time.Time
}

// NuevoConsultarEstadoDelProyecto arma el caso de uso. ahora es el reloj: time.Now en cmd/server y uno fijo en los
// tests y en los escenarios.
func NuevoConsultarEstadoDelProyecto(repo RepositorioEstadoGeneral, ahora func() time.Time) *ConsultarEstadoDelProyecto {
	return &ConsultarEstadoDelProyecto{repo: repo, ahora: ahora}
}

// Ejecutar verifica el proyecto (un proyecto que no existe es ErrProyectoInexistente y no se lee nada mas, RN-004-4), lee
// sus historias y sus sprints (una sola lectura de cada tipo, RN-004-8) y arma el estado general. No modifica nada.
func (c *ConsultarEstadoDelProyecto) Ejecutar(ctx context.Context, proyectoID string) (EstadoGeneral, error) {
	p, err := c.repo.BuscarProyecto(ctx, proyectoID)
	if err != nil {
		return EstadoGeneral{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	items, err := c.repo.ListarItems(ctx, proyectoID)
	if err != nil {
		return EstadoGeneral{}, fmt.Errorf("listar las historias del proyecto %s: %w", proyectoID, err)
	}
	sprints, err := c.repo.ListarSprintsConItems(ctx, proyectoID)
	if err != nil {
		return EstadoGeneral{}, fmt.Errorf("listar los sprints del proyecto %s: %w", proyectoID, err)
	}

	e, err := armarEstadoGeneral(p, items, sprints, c.ahora())
	if err != nil {
		return EstadoGeneral{}, fmt.Errorf("calcular el avance del proyecto %s: %w", proyectoID, err)
	}
	return e, nil
}

// armarEstadoGeneral calcula el estado general con lo que ya se leyo, sin tocar la persistencia: lo usan el caso de uso de US-004
// y el dashboard (US-030), que asi no vuelve a leer el proyecto, sus historias ni sus sprints.
func armarEstadoGeneral(p proyecto.Proyecto, items []backlog.ItemBacklog, sprints []SprintConItems, hoy time.Time) (EstadoGeneral, error) {
	e := EstadoGeneral{Proyecto: p}
	e.Estado, e.Dias = p.EstadoEn(hoy)

	// RN-004-5: se cuentan por su estado actual y los Story Points de las sin estimar suman 0.
	datos := make([]metricas.ItemDelSprint, len(items))
	for i, item := range items {
		puntos, _ := item.StoryPoints.Valor()
		completada := item.Estado == backlog.EstadoCompletado
		datos[i] = metricas.ItemDelSprint{StoryPoints: puntos, Completado: completada}
		e.Historias.Total++
		e.StoryPoints += puntos
		switch item.Estado {
		case backlog.EstadoCompletado:
			e.Historias.Completadas++
			e.Completados += puntos
		case backlog.EstadoEnProgreso:
			e.Historias.EnProgreso++
		default:
			e.Historias.Pendientes++
		}
	}

	// RN-004-6: el porcentaje es el de metricas (US-028) sobre todas las historias del proyecto.
	porcentaje, err := metricas.CalcularPorcentajeCompletadas(datos)
	if err != nil {
		// El proyecto se conserva: quien muestra la pagina lo necesita aunque la tarjeta falle (RN-030-9).
		return EstadoGeneral{Proyecto: p}, err
	}
	e.Porcentaje = porcentaje

	// RN-004-7: el sprint abierto es el unico con estado abierto (US-008 garantiza que hay como mucho uno).
	for _, s := range sprints {
		if s.Sprint.Estado == sprint.EstadoAbierto {
			abierto := s.Sprint
			e.SprintAbierto = &abierto
			break
		}
	}
	return e, nil
}
