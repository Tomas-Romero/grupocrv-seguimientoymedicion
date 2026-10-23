package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// RepositorioRevelar es lo que el caso de uso de US-016 necesita de la persistencia.
type RepositorioRevelar interface {
	// ObtenerSesion devuelve la sesion, con su proyecto y sus participantes, y si existe.
	// Un ID mal formado es una sesion inexistente, no un error.
	ObtenerSesion(ctx context.Context, sesionID string) (planningpoker.SesionPlanningPoker, bool, error)

	// ExisteIntegranteEnProyecto dice si el integrante existe y es de ese proyecto (el
	// de US-019).
	ExisteIntegranteEnProyecto(ctx context.Context, proyectoID, integranteID string) (bool, error)

	// VotosDeLaRonda devuelve el voto de cada participante en la ronda actual de la
	// sesion, por integrante. Es la entrada del dominio para validar: los valores no
	// salen del caso de uso (RN-015-7).
	VotosDeLaRonda(ctx context.Context, sesionID string) (map[string]int, error)

	// GuardarRevelado revela la sesion en una transaccion, con los pasos de RN-016-8:
	// bloquea la sesion, lee sus participantes y los votos de la ronda, vuelve a
	// verificar con planningpoker.Revelar y la pasa a revelada. Devuelve la revelacion.
	// Si la sesion desaparecio, devuelve ErrSesionInexistente; si las reglas ya no se
	// cumplen, los errores de Revelar. Es de todo o nada.
	GuardarRevelado(ctx context.Context, sesionID, integranteID string) (planningpoker.Revelacion, error)
}

// RevelarVotos es el caso de uso de US-016: cerrar la votacion de la ronda actual y
// mostrar los votos y si hay divergencia.
type RevelarVotos struct {
	repo RepositorioRevelar
}

// NuevoRevelarVotos arma el caso de uso.
func NuevoRevelarVotos(repo RepositorioRevelar) *RevelarVotos {
	return &RevelarVotos{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): verifica la sesion y el integrante, valida
// con el dominio y revela. Si la sesion o el integrante no existen, o el integrante es de
// otro proyecto, responde ese error sin mirar lo demas (RN-016-1). Devuelve la revelacion; si
// algo falla no cambia nada (RN-016-11).
func (c *RevelarVotos) Ejecutar(ctx context.Context, sesionID, integranteID string) (planningpoker.Revelacion, error) {
	sesion, existe, err := c.repo.ObtenerSesion(ctx, sesionID)
	if err != nil {
		return planningpoker.Revelacion{}, fmt.Errorf("consultar la sesion %s: %w", sesionID, err)
	}
	if !existe {
		return planningpoker.Revelacion{}, fmt.Errorf("%w (id %s)", ErrSesionInexistente, sesionID)
	}

	existe, err = c.repo.ExisteIntegranteEnProyecto(ctx, sesion.ProyectoID, integranteID)
	if err != nil {
		return planningpoker.Revelacion{}, fmt.Errorf("consultar el integrante %s: %w", integranteID, err)
	}
	if !existe {
		return planningpoker.Revelacion{}, fmt.Errorf("%w (id %s)", ErrIntegranteInexistente, integranteID)
	}

	votos, err := c.repo.VotosDeLaRonda(ctx, sesionID)
	if err != nil {
		return planningpoker.Revelacion{}, fmt.Errorf("leer los votos de la sesion %s: %w", sesionID, err)
	}

	// Los errores del dominio se devuelven tal cual: su mensaje es el que ve la persona
	// (spec US-016, seccion 7).
	if _, err := planningpoker.Revelar(sesion, integranteID, votos); err != nil {
		return planningpoker.Revelacion{}, err
	}

	revelacion, err := c.repo.GuardarRevelado(ctx, sesionID, integranteID)
	if err != nil {
		return planningpoker.Revelacion{}, fmt.Errorf("revelar la sesion %s: %w", sesionID, err)
	}
	return revelacion, nil
}
