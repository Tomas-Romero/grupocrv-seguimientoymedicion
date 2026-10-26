package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// RepositorioAcuerdo es lo que el caso de uso de US-018 necesita de la persistencia. Como
// los de US-015 y US-016, ningun metodo entrega los votos en crudo (RN-016-12).
type RepositorioAcuerdo interface {
	// ObtenerSesion devuelve la sesion, con su proyecto y sus participantes, y si existe.
	// Un ID mal formado es una sesion inexistente, no un error.
	ObtenerSesion(ctx context.Context, sesionID string) (planningpoker.SesionPlanningPoker, bool, error)

	// ExisteIntegranteEnProyecto dice si el integrante existe y es de ese proyecto (el
	// de US-019).
	ExisteIntegranteEnProyecto(ctx context.Context, proyectoID, integranteID string) (bool, error)

	// ObtenerItem devuelve la historia y si existe (el de US-013).
	ObtenerItem(ctx context.Context, itemID string) (backlog.ItemBacklog, bool, error)

	// GuardarAcuerdo registra la estimacion acordada en una sola transaccion, con los pasos
	// de RN-018-9: bloquea la sesion, vuelve a verificar con planningpoker.Acordar, estima la
	// historia con el UPDATE condicionado de US-013 y finaliza la sesion con el valor.
	// Devuelve la sesion finalizada. Si la sesion desaparecio, devuelve ErrSesionInexistente;
	// si las reglas ya no se cumplen (o la historia se completo en el medio), los errores de
	// Acordar. Es de todo o nada.
	GuardarAcuerdo(ctx context.Context, sesionID, integranteID string, valor int) (planningpoker.SesionPlanningPoker, error)
}

// RegistrarEstimacionAcordada es el caso de uso de US-018: registrar la estimacion
// acordada en una sesion revelada, que la finaliza y deja estimada la historia.
type RegistrarEstimacionAcordada struct {
	repo RepositorioAcuerdo
}

// NuevoRegistrarEstimacionAcordada arma el caso de uso.
func NuevoRegistrarEstimacionAcordada(repo RepositorioAcuerdo) *RegistrarEstimacionAcordada {
	return &RegistrarEstimacionAcordada{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): verifica la sesion y el integrante, lee la
// historia de la sesion, valida con el dominio y guarda el acuerdo. Si la sesion o el
// integrante no existen, o el integrante es de otro proyecto, responde ese error sin mirar
// lo demas (RN-018-1). No llama al caso de uso de US-013: estimar la historia va en la misma
// transaccion que finalizar la sesion (RN-018-8). Si algo falla no cambia nada.
func (c *RegistrarEstimacionAcordada) Ejecutar(ctx context.Context, sesionID, integranteID string, valor int) (planningpoker.SesionPlanningPoker, error) {
	sesion, existe, err := c.repo.ObtenerSesion(ctx, sesionID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("consultar la sesion %s: %w", sesionID, err)
	}
	if !existe {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", ErrSesionInexistente, sesionID)
	}

	existe, err = c.repo.ExisteIntegranteEnProyecto(ctx, sesion.ProyectoID, integranteID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("consultar el integrante %s: %w", integranteID, err)
	}
	if !existe {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", ErrIntegranteInexistente, integranteID)
	}

	item, existe, err := c.repo.ObtenerItem(ctx, sesion.ItemID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("consultar el item %s: %w", sesion.ItemID, err)
	}
	if !existe {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", ErrItemInexistente, sesion.ItemID)
	}

	// Los errores del dominio se devuelven tal cual: su mensaje es el que ve la persona
	// (spec US-018, seccion 7).
	if _, err := planningpoker.Acordar(sesion, item.Estado == backlog.EstadoCompletado, integranteID, valor); err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	finalizada, err := c.repo.GuardarAcuerdo(ctx, sesionID, integranteID, valor)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("registrar el acuerdo de la sesion %s: %w", sesionID, err)
	}
	return finalizada, nil
}
