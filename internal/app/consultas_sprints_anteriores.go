package app

import (
	"context"
	"fmt"
	"sort"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// SprintAnterior es un sprint cerrado con su resultado (spec US-012, seccion 3).
type SprintAnterior struct {
	Sprint sprint.Sprint
	// Planificados y Completados son los Story Points del sprint (US-025).
	Planificados int
	Completados  int
	// HistoriasCompletadas y HistoriasNoCompletadas cuentan historias, no puntos (RN-012-6).
	HistoriasCompletadas   int
	HistoriasNoCompletadas int
}

// SprintsAnteriores son los sprints cerrados de un proyecto, del mas reciente al mas antiguo, y
// la velocidad del equipo calculada con ellos (US-026).
type SprintsAnteriores struct {
	Sprints   []SprintAnterior
	Velocidad float64
}

// ConsultarSprintsAnteriores es el caso de uso de US-012: el historial de sprints cerrados de un
// proyecto con su resultado. Usa el mismo puerto de lectura que la lista de sprints (T-015), asi
// que no necesita nada nuevo de la persistencia.
type ConsultarSprintsAnteriores struct {
	repo RepositorioListarSprints
}

// NuevoConsultarSprintsAnteriores arma el caso de uso.
func NuevoConsultarSprintsAnteriores(repo RepositorioListarSprints) *ConsultarSprintsAnteriores {
	return &ConsultarSprintsAnteriores{repo: repo}
}

// Ejecutar devuelve los sprints cerrados del proyecto (RN-012-2) por numero descendente
// (RN-012-3). El resultado de cada historia es el de su fila en ese sprint, no su estado actual
// (RN-012-4). Los Story Points salen de metricas.CalcularResumenSprint (US-025) y la velocidad de
// metricas.CalcularVelocidad (US-026), solo con estos sprints (RN-012-5). Sin sprints cerrados
// devuelve una lista vacia y velocidad 0, y no es un error (RN-012-7). Un proyecto que no existe es
// ErrProyectoInexistente (RN-012-1). Es una lectura: no modifica nada (RN-012-8).
func (c *ConsultarSprintsAnteriores) Ejecutar(ctx context.Context, proyectoID string) (SprintsAnteriores, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return SprintsAnteriores{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return SprintsAnteriores{}, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}
	sprints, err := c.repo.ListarSprintsConItems(ctx, proyectoID)
	if err != nil {
		return SprintsAnteriores{}, fmt.Errorf("listar los sprints del proyecto %s: %w", proyectoID, err)
	}

	anteriores := []SprintAnterior{}
	cerrados := []metricas.SprintCerrado{}
	for _, s := range sprints {
		if s.Sprint.Estado != sprint.EstadoCerrado {
			continue
		}
		totales, err := resumenDeItems(s.Items)
		if err != nil {
			return SprintsAnteriores{}, fmt.Errorf("calcular los Story Points del %s: %w", s.Sprint.Nombre(), err)
		}
		anterior := SprintAnterior{Sprint: s.Sprint, Planificados: totales.Planificados, Completados: totales.Completados}
		for _, h := range s.Items {
			if h.Completada() {
				anterior.HistoriasCompletadas++
			} else {
				anterior.HistoriasNoCompletadas++
			}
		}
		anteriores = append(anteriores, anterior)
		cerrados = append(cerrados, metricas.SprintCerrado{Nombre: s.Sprint.Nombre(), Resumen: totales})
	}
	sort.SliceStable(anteriores, func(i, j int) bool { return anteriores[i].Sprint.Numero > anteriores[j].Sprint.Numero })

	velocidad, err := metricas.CalcularVelocidad(cerrados)
	if err != nil {
		return SprintsAnteriores{}, fmt.Errorf("calcular la velocidad del equipo: %w", err)
	}
	return SprintsAnteriores{Sprints: anteriores, Velocidad: velocidad}, nil
}
