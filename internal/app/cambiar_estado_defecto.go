package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// RepositorioCambioDeEstadoDefecto es lo que el caso de uso de US-023 necesita de la persistencia.
type RepositorioCambioDeEstadoDefecto interface {
	// ExisteProyecto dice si el proyecto existe. Un ID mal formado es un proyecto
	// inexistente, no un error.
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)

	// DefectoEnProyecto devuelve el defecto guardado y dice si existe y es de ese proyecto: uno
	// de otro proyecto no cuenta. Un ID vacio o mal formado es inexistente.
	DefectoEnProyecto(ctx context.Context, proyectoID, defectoID string) (defecto.Defecto, bool, error)

	// NumeroDeSprintEnProyecto devuelve el numero (sprint.Numero) del sprint y dice si existe y es
	// de ese proyecto: uno de otro proyecto no cuenta. Un ID vacio o mal formado es inexistente.
	NumeroDeSprintEnProyecto(ctx context.Context, proyectoID, sprintID string) (int, bool, error)

	// ActualizarEstadoDefecto guarda el estado y el sprint de resolucion de d con un unico UPDATE
	// condicionado a que el defecto siga en estadoAnterior, el que leyo el caso de uso. Si el
	// defecto ya no esta en ese estado devuelve ErrDefectoModificado; si no existe,
	// ErrDefectoInexistente; si el sprint de resolucion desaparecio, ErrSprintInexistente.
	ActualizarEstadoDefecto(ctx context.Context, d defecto.Defecto, estadoAnterior defecto.Estado) error
}

// CambiarEstadoDefecto es el caso de uso de US-023: cambiar el estado de un defecto y, cuando se
// resuelve, registrar el sprint de resolucion.
type CambiarEstadoDefecto struct {
	repo RepositorioCambioDeEstadoDefecto
}

// NuevoCambiarEstadoDefecto arma el caso de uso. No recibe reloj: no se guarda cuando se cambio.
func NuevoCambiarEstadoDefecto(repo RepositorioCambioDeEstadoDefecto) *CambiarEstadoDefecto {
	return &CambiarEstadoDefecto{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): verifica el proyecto, lee el defecto del
// proyecto, verifica el sprint de resolucion si se informo, traduce los sprints a su numero,
// valida con el dominio y guarda con un UPDATE condicionado al estado leido.
func (c *CambiarEstadoDefecto) Ejecutar(ctx context.Context, proyectoID, defectoID string, estado defecto.Estado, sprintResolucionID string) (defecto.Defecto, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return defecto.Defecto{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return defecto.Defecto{}, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}

	// RN-023-8: el estado actual sale del defecto guardado, nunca del formulario.
	actual, existe, err := c.repo.DefectoEnProyecto(ctx, proyectoID, defectoID)
	if err != nil {
		return defecto.Defecto{}, fmt.Errorf("consultar el defecto %s del proyecto %s: %w", defectoID, proyectoID, err)
	}
	if !existe {
		return defecto.Defecto{}, fmt.Errorf("%w (id %s, proyecto %s)", ErrDefectoInexistente, defectoID, proyectoID)
	}

	var resolucion defecto.SprintDeResolucion
	if sprintResolucionID != "" {
		// Un ID vacio es "sin sprint de resolucion" y no se verifica: con resuelto lo rechaza el
		// dominio (RN-023-3). Uno informado tiene que ser un sprint de este proyecto.
		numero, existe, err := c.repo.NumeroDeSprintEnProyecto(ctx, proyectoID, sprintResolucionID)
		if err != nil {
			return defecto.Defecto{}, fmt.Errorf("consultar el sprint %s del proyecto %s: %w", sprintResolucionID, proyectoID, err)
		}
		if !existe {
			return defecto.Defecto{}, fmt.Errorf("%w (id %s)", ErrSprintInexistente, sprintResolucionID)
		}
		resolucion = defecto.SprintDeResolucion{ID: sprintResolucionID, Numero: numero}
	}
	numeroDeteccion, _, err := c.repo.NumeroDeSprintEnProyecto(ctx, proyectoID, actual.SprintDeteccionID)
	if err != nil {
		return defecto.Defecto{}, fmt.Errorf("consultar el sprint %s del proyecto %s: %w", actual.SprintDeteccionID, proyectoID, err)
	}

	nuevo, err := actual.CambiarEstado(estado, resolucion, numeroDeteccion)
	if err != nil {
		return defecto.Defecto{}, err
	}

	if err := c.repo.ActualizarEstadoDefecto(ctx, nuevo, actual.Estado); err != nil {
		return defecto.Defecto{}, fmt.Errorf("cambiar el estado del defecto %s: %w", defectoID, err)
	}
	return nuevo, nil
}
