package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
)

// HistoriaConHoras es una historia del proyecto con las horas de todos sus registros de esfuerzo.
type HistoriaConHoras struct {
	Item  backlog.ItemBacklog
	Horas []float64
}

// RepositorioCompararEsfuerzo es lo que el caso de uso de US-021 necesita de la persistencia.
type RepositorioCompararEsfuerzo interface {
	// ExisteProyecto dice si el proyecto existe. Un ID mal formado es un proyecto inexistente,
	// no un error.
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)

	// LeerHistoriasConHoras devuelve todas las historias del proyecto, cada una con las horas de
	// todos sus registros, sin ordenar. Es una sola lectura consistente.
	LeerHistoriasConHoras(ctx context.Context, proyectoID string) ([]HistoriaConHoras, error)
}

// CompararEsfuerzo es el caso de uso de lectura de US-021.
type CompararEsfuerzo struct {
	repo RepositorioCompararEsfuerzo
}

// NuevoCompararEsfuerzo arma el caso de uso.
func NuevoCompararEsfuerzo(repo RepositorioCompararEsfuerzo) *CompararEsfuerzo {
	return &CompararEsfuerzo{repo: repo}
}

// Ejecutar verifica el proyecto, lee sus historias con las horas y las compara con metricas. La
// estimacion que se usa es la actual de cada historia (RN-021-8) y "completada" es su estado.
func (c *CompararEsfuerzo) Ejecutar(ctx context.Context, proyectoID string) (metricas.ComparacionProyecto, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return metricas.ComparacionProyecto{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return metricas.ComparacionProyecto{}, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}

	leidas, err := c.repo.LeerHistoriasConHoras(ctx, proyectoID)
	if err != nil {
		return metricas.ComparacionProyecto{}, fmt.Errorf("leer las historias con horas del proyecto %s: %w", proyectoID, err)
	}

	historias := make([]metricas.HistoriaParaComparar, len(leidas))
	for i, h := range leidas {
		puntos, _ := h.Item.StoryPoints.Valor()
		historias[i] = metricas.HistoriaParaComparar{
			ItemID:      h.Item.ID,
			Titulo:      h.Item.Titulo,
			StoryPoints: puntos,
			Completada:  h.Item.Estado == backlog.EstadoCompletado,
			Horas:       h.Horas,
		}
	}
	comparacion, err := metricas.CompararHistorias(historias)
	if err != nil {
		return metricas.ComparacionProyecto{}, fmt.Errorf("comparar el esfuerzo del proyecto %s: %w", proyectoID, err)
	}
	return comparacion, nil
}
