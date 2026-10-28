package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// RepositorioNuevaRonda es lo que el caso de uso de US-017 necesita de la persistencia.
type RepositorioNuevaRonda interface {
	// ObtenerSesion devuelve la sesion, con su proyecto y sus participantes, y si existe.
	// Un ID mal formado es una sesion inexistente, no un error.
	ObtenerSesion(ctx context.Context, sesionID string) (planningpoker.SesionPlanningPoker, bool, error)

	// ExisteIntegranteEnProyecto dice si el integrante existe y es de ese proyecto (el
	// de US-019).
	ExisteIntegranteEnProyecto(ctx context.Context, proyectoID, integranteID string) (bool, error)

	// VotacionDeLaRonda devuelve la votacion de la ronda actual de la sesion (la de US-016).
	VotacionDeLaRonda(ctx context.Context, sesionID string) (planningpoker.Votacion, error)

	// GuardarNuevaRonda abre la ronda siguiente en una transaccion, con los pasos de
	// RN-017-8: bloquea la sesion, lee sus participantes y los votos de la ronda, vuelve a
	// verificar con planningpoker.AbrirNuevaRonda y la pasa a votando en la ronda siguiente.
	// Devuelve la sesion en la ronda nueva. Si la sesion desaparecio, devuelve
	// ErrSesionInexistente; si las reglas ya no se cumplen, los errores de AbrirNuevaRonda.
	// No toca ningun voto.
	GuardarNuevaRonda(ctx context.Context, sesionID, integranteID string) (planningpoker.SesionPlanningPoker, error)
}

// RondaAnterior es una ronda anterior a la actual de una sesion, con su votacion, que deja
// leer los valores porque ya se revelo (RN-017-6). Ningun puerto entrega los votos en crudo
// (RN-017-10).
type RondaAnterior struct {
	Numero   int
	Votacion planningpoker.Votacion
}

// LectorDeRondasAnteriores es la lectura de las rondas anteriores de una sesion (spec US-017,
// seccion 5). La usan la pantalla de la sesion (T-020) y los escenarios.
type LectorDeRondasAnteriores interface {
	// RondasAnteriores devuelve las rondas anteriores a la actual, de la primera a la ultima.
	// Una sesion que no esta, o que esta en la ronda 1, no tiene rondas anteriores.
	RondasAnteriores(ctx context.Context, sesionID string) ([]RondaAnterior, error)
}

// AbrirNuevaRonda es el caso de uso de US-017: abrir la ronda siguiente de una sesion revelada
// con divergencia, para volver a votar.
type AbrirNuevaRonda struct {
	repo RepositorioNuevaRonda
}

// NuevoAbrirNuevaRonda arma el caso de uso.
func NuevoAbrirNuevaRonda(repo RepositorioNuevaRonda) *AbrirNuevaRonda {
	return &AbrirNuevaRonda{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): verifica la sesion y el integrante, valida
// con el dominio y abre la ronda. Si la sesion o el integrante no existen, o el integrante es
// de otro proyecto, responde ese error sin mirar lo demas (RN-017-1). Si algo falla no cambia
// nada.
func (c *AbrirNuevaRonda) Ejecutar(ctx context.Context, sesionID, integranteID string) (planningpoker.SesionPlanningPoker, error) {
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

	votacion, err := c.repo.VotacionDeLaRonda(ctx, sesionID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("leer los votos de la sesion %s: %w", sesionID, err)
	}

	// Los errores del dominio se devuelven tal cual: su mensaje es el que ve la persona
	// (spec US-017, seccion 7).
	if _, err := planningpoker.AbrirNuevaRonda(sesion, integranteID, votacion); err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	nueva, err := c.repo.GuardarNuevaRonda(ctx, sesionID, integranteID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("abrir una nueva ronda en la sesion %s: %w", sesionID, err)
	}
	return nueva, nil
}
