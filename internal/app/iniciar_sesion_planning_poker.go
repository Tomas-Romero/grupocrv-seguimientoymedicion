package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// RepositorioPlanningPoker es lo que el caso de uso de US-014 necesita de la
// persistencia.
type RepositorioPlanningPoker interface {
	// ExisteProyecto dice si el proyecto existe. Un ID mal formado es un proyecto
	// inexistente, no un error.
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)

	// ObtenerItem devuelve el item y si existe. Un ID mal formado es un item
	// inexistente, no un error.
	ObtenerItem(ctx context.Context, itemID string) (backlog.ItemBacklog, bool, error)

	// TieneSesionAbierta dice si la historia tiene una sesion que no esta finalizada
	// (RN-014-7).
	TieneSesionAbierta(ctx context.Context, itemID string) (bool, error)

	// IDsDeIntegrantes devuelve los IDs de los integrantes del proyecto, en orden de
	// alta (RN-014-4).
	IDsDeIntegrantes(ctx context.Context, proyectoID string) ([]string, error)

	// RegistrarSesion registra la sesion y sus participantes en una transaccion, con los
	// pasos de RN-014-11: bloquea la historia, vuelve a leer su estado, si tiene sesion
	// abierta y los integrantes del proyecto, vuelve a armar la sesion con
	// planningpoker.NuevaSesion y la guarda. Devuelve la sesion con su ID y los
	// participantes que guardo. Si la historia desaparecio o no es del proyecto,
	// devuelve ErrItemInexistente; si las reglas ya no se cumplen, los errores de
	// NuevaSesion. Es de todo o nada.
	RegistrarSesion(ctx context.Context, proyectoID string, sesion planningpoker.SesionPlanningPoker) (planningpoker.SesionPlanningPoker, error)
}

// IniciarSesionPlanningPoker es el caso de uso de US-014: iniciar una sesion de
// Planning Poker sobre una historia del backlog.
type IniciarSesionPlanningPoker struct {
	repo RepositorioPlanningPoker
}

// NuevoIniciarSesionPlanningPoker arma el caso de uso.
func NuevoIniciarSesionPlanningPoker(repo RepositorioPlanningPoker) *IniciarSesionPlanningPoker {
	return &IniciarSesionPlanningPoker{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): verifica el proyecto y la historia,
// valida con el dominio y registra la sesion. Si el proyecto o la historia no existen,
// o la historia es de otro proyecto, responde ese error sin mirar lo demas (RN-014-1).
// Devuelve la sesion registrada; si algo falla no registra nada (RN-014-14).
func (c *IniciarSesionPlanningPoker) Ejecutar(ctx context.Context, proyectoID, itemID string) (planningpoker.SesionPlanningPoker, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}

	item, existe, err := c.repo.ObtenerItem(ctx, itemID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("consultar el item %s: %w", itemID, err)
	}
	// Una historia de otro proyecto es una historia inexistente (CL-019-15).
	if !existe || item.ProyectoID != proyectoID {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", ErrItemInexistente, itemID)
	}

	abierta, err := c.repo.TieneSesionAbierta(ctx, itemID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("consultar las sesiones del item %s: %w", itemID, err)
	}
	participantes, err := c.repo.IDsDeIntegrantes(ctx, proyectoID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("leer los integrantes del proyecto %s: %w", proyectoID, err)
	}

	// Los errores del dominio se devuelven tal cual: su mensaje es el que ve la persona
	// (spec US-014, seccion 7).
	historia := planningpoker.DatosDeLaHistoria{Completada: item.Estado == backlog.EstadoCompletado, ConSesionAbierta: abierta}
	sesion, err := planningpoker.NuevaSesion(itemID, historia, participantes)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	registrada, err := c.repo.RegistrarSesion(ctx, proyectoID, sesion)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("registrar la sesion del item %s: %w", itemID, err)
	}
	return registrada, nil
}
