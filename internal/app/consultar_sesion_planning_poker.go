package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// RepositorioConsultaDeSesion es lo que la pantalla de la sesion (T-020) necesita leer. Ningun
// metodo entrega los votos en crudo (RN-016-12): las votaciones llegan como Votacion.
type RepositorioConsultaDeSesion interface {
	// ObtenerSesion es el de US-015.
	ObtenerSesion(ctx context.Context, sesionID string) (planningpoker.SesionPlanningPoker, bool, error)
	// ObtenerItem es el de US-014.
	ObtenerItem(ctx context.Context, itemID string) (backlog.ItemBacklog, bool, error)
	// VotacionDeLaRonda es el de US-016.
	VotacionDeLaRonda(ctx context.Context, sesionID string) (planningpoker.Votacion, error)
	LectorDeRondasAnteriores
	// SesionesAbiertas devuelve, para cada historia del proyecto con una sesion abierta, el ID
	// de esa sesion (RN-014-7). Un ID mal formado no tiene sesiones.
	SesionesAbiertas(ctx context.Context, proyectoID string) (map[string]string, error)
}

// DetalleDeSesion es lo que muestra la pantalla de una sesion (spec T-020, seccion 3).
type DetalleDeSesion struct {
	Sesion   planningpoker.SesionPlanningPoker
	Historia backlog.ItemBacklog
	// Votaron y Faltan son los participantes de la ronda actual, en el orden de la sesion.
	Votaron []string
	Faltan  []string
	// Revelada dice si la ronda actual se puede leer (la sesion no esta votando); solo
	// entonces Resultado tiene sus votos (RN-T020-2).
	Revelada  bool
	Resultado planningpoker.Resultado
	// Anteriores son las rondas anteriores a la actual, de la primera a la ultima (RN-017-6).
	Anteriores []ResultadoDeRonda
}

// ResultadoDeRonda es el resultado de una ronda anterior.
type ResultadoDeRonda struct {
	Numero    int
	Resultado planningpoker.Resultado
}

// ConsultarSesionPlanningPoker es la lectura de la pantalla de la sesion (T-020, RN-T020-3).
// Compone lo que ya existe y arma los resultados con la funcion del dominio de US-017: la
// pantalla no calcula la divergencia ni los extremos.
type ConsultarSesionPlanningPoker struct {
	repo RepositorioConsultaDeSesion
}

// NuevoConsultarSesionPlanningPoker arma la lectura.
func NuevoConsultarSesionPlanningPoker(repo RepositorioConsultaDeSesion) *ConsultarSesionPlanningPoker {
	return &ConsultarSesionPlanningPoker{repo: repo}
}

// Ejecutar devuelve el detalle de la sesion. Una sesion que no existe, o que no es del proyecto,
// es ErrSesionInexistente (RN-T020-4).
func (c *ConsultarSesionPlanningPoker) Ejecutar(ctx context.Context, proyectoID, sesionID string) (DetalleDeSesion, error) {
	sesion, existe, err := c.repo.ObtenerSesion(ctx, sesionID)
	if err != nil {
		return DetalleDeSesion{}, fmt.Errorf("consultar la sesion %s: %w", sesionID, err)
	}
	if !existe || sesion.ProyectoID != proyectoID {
		return DetalleDeSesion{}, fmt.Errorf("%w (id %s)", ErrSesionInexistente, sesionID)
	}
	item, _, err := c.repo.ObtenerItem(ctx, sesion.ItemID)
	if err != nil {
		return DetalleDeSesion{}, fmt.Errorf("consultar la historia %s: %w", sesion.ItemID, err)
	}
	votacion, err := c.repo.VotacionDeLaRonda(ctx, sesionID)
	if err != nil {
		return DetalleDeSesion{}, fmt.Errorf("leer los votos de la sesion %s: %w", sesionID, err)
	}
	rondas, err := c.repo.RondasAnteriores(ctx, sesionID)
	if err != nil {
		return DetalleDeSesion{}, fmt.Errorf("leer las rondas anteriores de la sesion %s: %w", sesionID, err)
	}

	d := DetalleDeSesion{Sesion: sesion, Historia: item, Votaron: votacion.Votaron(), Faltan: votacion.Faltan(), Anteriores: []ResultadoDeRonda{}}
	// Votando, ResultadoDeLaRonda responde ErrVotosOcultos: la ronda actual queda sin resultado.
	if r, err := planningpoker.ResultadoDeLaRonda(votacion); err == nil {
		d.Revelada, d.Resultado = true, r
	}
	for _, ronda := range rondas {
		r, err := planningpoker.ResultadoDeLaRonda(ronda.Votacion)
		if err != nil {
			return DetalleDeSesion{}, fmt.Errorf("leer la ronda %d de la sesion %s: %w", ronda.Numero, sesionID, err)
		}
		d.Anteriores = append(d.Anteriores, ResultadoDeRonda{Numero: ronda.Numero, Resultado: r})
	}
	return d, nil
}

// AbiertasDelProyecto devuelve, para cada historia del proyecto con una sesion abierta, el ID de
// esa sesion: el backlog lleva a ella en lugar de ofrecer iniciar otra (CL-T020-2).
func (c *ConsultarSesionPlanningPoker) AbiertasDelProyecto(ctx context.Context, proyectoID string) (map[string]string, error) {
	abiertas, err := c.repo.SesionesAbiertas(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("consultar las sesiones abiertas del proyecto %s: %w", proyectoID, err)
	}
	return abiertas, nil
}
