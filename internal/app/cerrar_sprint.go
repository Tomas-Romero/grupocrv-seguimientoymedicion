package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// RepositorioCierre es lo que el caso de uso de US-011 necesita de la persistencia.
type RepositorioCierre interface {
	// ObtenerSprint devuelve el sprint y si existe. Un ID mal formado es un sprint
	// inexistente, no un error.
	ObtenerSprint(ctx context.Context, sprintID string) (sprint.Sprint, bool, error)

	// HistoriasDelSprint devuelve las historias del sprint (sus filas pendientes en la
	// relacion de US-009) y si cada una esta completada.
	HistoriasDelSprint(ctx context.Context, sprintID string) ([]sprint.HistoriaDelCierre, error)

	// GuardarCierre cierra el sprint en una transaccion, con los pasos de RN-011-8:
	// bloquea el sprint, lee sus historias, vuelve a calcular con sprint.Cerrar, guarda
	// el resultado de cada fila, pasa a pendiente las no completadas y deja el sprint
	// cerrado. Devuelve los resultados. Si el sprint desaparecio, devuelve
	// ErrSprintInexistente; si otro cierre llego antes, sprint.ErrSprintCerrado. Es de
	// todo o nada.
	GuardarCierre(ctx context.Context, sprintID string) ([]sprint.ResultadoDelCierre, error)
}

// CerrarSprint es el caso de uso de US-011: cerrar un sprint y devolver al backlog las
// historias no completadas.
type CerrarSprint struct {
	repo RepositorioCierre
}

// NuevoCerrarSprint arma el caso de uso.
func NuevoCerrarSprint(repo RepositorioCierre) *CerrarSprint {
	return &CerrarSprint{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): busca el sprint, valida con el
// dominio y guarda el cierre. Si el sprint no existe responde ErrSprintInexistente sin
// mirar lo demas (RN-011-1). Devuelve el resultado de cada historia; si algo falla no
// cambia nada (RN-011-8).
func (c *CerrarSprint) Ejecutar(ctx context.Context, sprintID string) ([]sprint.ResultadoDelCierre, error) {
	s, existe, err := c.repo.ObtenerSprint(ctx, sprintID)
	if err != nil {
		return nil, fmt.Errorf("consultar el sprint %s: %w", sprintID, err)
	}
	if !existe {
		return nil, fmt.Errorf("%w (id %s)", ErrSprintInexistente, sprintID)
	}

	historias, err := c.repo.HistoriasDelSprint(ctx, sprintID)
	if err != nil {
		return nil, fmt.Errorf("consultar las historias del sprint %s: %w", sprintID, err)
	}

	// El error del dominio se devuelve tal cual: su mensaje es el que ve la persona
	// (spec US-011, seccion 7).
	if _, err := sprint.Cerrar(datosDelSprint(s), historias); err != nil {
		return nil, err
	}

	resultados, err := c.repo.GuardarCierre(ctx, sprintID)
	if err != nil {
		return nil, fmt.Errorf("cerrar el sprint %s: %w", sprintID, err)
	}
	return resultados, nil
}
