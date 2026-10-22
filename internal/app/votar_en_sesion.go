package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// ErrSesionInexistente se devuelve cuando la operacion apunta a una sesion de Planning
// Poker que no existe. Igual que ErrItemInexistente, tiene prioridad sobre los errores de
// validacion y nunca se une con ellos (RN-015-1).
var ErrSesionInexistente = errors.New("la sesion de planning poker no existe")

// RepositorioVotos es lo que el caso de uso de US-015 necesita de la persistencia.
type RepositorioVotos interface {
	// ObtenerSesion devuelve la sesion, con su proyecto y sus participantes, y si existe.
	// Un ID mal formado es una sesion inexistente, no un error.
	ObtenerSesion(ctx context.Context, sesionID string) (planningpoker.SesionPlanningPoker, bool, error)

	// ExisteIntegranteEnProyecto dice si el integrante existe y es de ese proyecto (el
	// de US-019).
	ExisteIntegranteEnProyecto(ctx context.Context, proyectoID, integranteID string) (bool, error)

	// RegistrarVoto guarda el voto en la ronda actual en una transaccion, con los pasos de
	// RN-015-9: bloquea la sesion, vuelve a verificar con planningpoker.VerificarVoto e
	// inserta el voto o reemplaza el anterior del mismo participante. Devuelve la votacion
	// de la ronda. Si la sesion desaparecio, devuelve ErrSesionInexistente; si las reglas
	// ya no se cumplen, los errores de VerificarVoto. Es de todo o nada.
	RegistrarVoto(ctx context.Context, sesionID, integranteID string, valor int) (planningpoker.Votacion, error)
}

// VotarEnSesion es el caso de uso de US-015: registrar el voto de un participante,
// oculto hasta que se revele.
type VotarEnSesion struct {
	repo RepositorioVotos
}

// NuevoVotarEnSesion arma el caso de uso.
func NuevoVotarEnSesion(repo RepositorioVotos) *VotarEnSesion {
	return &VotarEnSesion{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): verifica la sesion y el integrante,
// valida con el dominio y guarda el voto. Si la sesion o el integrante no existen, o el
// integrante es de otro proyecto, responde ese error sin mirar lo demas (RN-015-1).
// Devuelve la votacion de la ronda, que mientras se vota no deja leer los valores
// (RN-015-7); si algo falla no guarda nada (RN-015-12).
func (c *VotarEnSesion) Ejecutar(ctx context.Context, sesionID, integranteID string, valor int) (planningpoker.Votacion, error) {
	sesion, existe, err := c.repo.ObtenerSesion(ctx, sesionID)
	if err != nil {
		return planningpoker.Votacion{}, fmt.Errorf("consultar la sesion %s: %w", sesionID, err)
	}
	if !existe {
		return planningpoker.Votacion{}, fmt.Errorf("%w (id %s)", ErrSesionInexistente, sesionID)
	}

	existe, err = c.repo.ExisteIntegranteEnProyecto(ctx, sesion.ProyectoID, integranteID)
	if err != nil {
		return planningpoker.Votacion{}, fmt.Errorf("consultar el integrante %s: %w", integranteID, err)
	}
	if !existe {
		return planningpoker.Votacion{}, fmt.Errorf("%w (id %s)", ErrIntegranteInexistente, integranteID)
	}

	// Los errores del dominio se devuelven tal cual: su mensaje es el que ve la persona
	// (spec US-015, seccion 7).
	if err := planningpoker.VerificarVoto(sesion, integranteID, valor); err != nil {
		return planningpoker.Votacion{}, err
	}

	votacion, err := c.repo.RegistrarVoto(ctx, sesionID, integranteID, valor)
	if err != nil {
		return planningpoker.Votacion{}, fmt.Errorf("guardar el voto en la sesion %s: %w", sesionID, err)
	}
	return votacion, nil
}
