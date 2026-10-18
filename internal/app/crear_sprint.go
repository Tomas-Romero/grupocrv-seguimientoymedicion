package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// RepositorioSprints es lo que el caso de uso de US-008 necesita de la
// persistencia.
type RepositorioSprints interface {
	// FechasDelProyecto lee las fechas de inicio y fin del proyecto y dice si
	// existe. Un ID mal formado es un proyecto inexistente, no un error.
	FechasDelProyecto(ctx context.Context, proyectoID string) (inicio, fin time.Time, existe bool, err error)

	// RegistrarSprint guarda el sprint y lo devuelve con su ID y su numero. Lo hace
	// dentro de una transaccion con el proyecto bloqueado (RN-008-14): vuelve a
	// verificar que el proyecto exista (ErrProyectoInexistente) y las fechas del
	// sprint contra las suyas (ErrFueraDelProyecto), y verifica el ultimo sprint
	// del proyecto (ErrUltimoSprintAbierto o ErrInicioNoPosteriorAlUltimo). Es de
	// todo o nada: si falla, no queda nada guardado.
	RegistrarSprint(ctx context.Context, s sprint.Sprint) (sprint.Sprint, error)
}

// CrearSprint es el caso de uso de US-008: crear un sprint con su Sprint Goal y sus
// fechas en un proyecto.
type CrearSprint struct {
	repo RepositorioSprints
}

// NuevoCrearSprint arma el caso de uso.
func NuevoCrearSprint(repo RepositorioSprints) *CrearSprint {
	return &CrearSprint{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): lee el proyecto con una lectura
// simple, valida los datos con el dominio pasandole las fechas del proyecto y
// registra el sprint. Si el proyecto no existe responde ErrProyectoInexistente sin
// mirar los datos (RN-008-9), y nunca lo une con los errores de validacion.
func (c *CrearSprint) Ejecutar(ctx context.Context, proyectoID, sprintGoal string, inicio, fin time.Time) (sprint.Sprint, error) {
	inicioProyecto, finProyecto, existe, err := c.repo.FechasDelProyecto(ctx, proyectoID)
	if err != nil {
		return sprint.Sprint{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return sprint.Sprint{}, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}

	nuevo, err := sprint.Nuevo(proyectoID, sprintGoal, inicio, fin, inicioProyecto, finProyecto)
	if err != nil {
		// Los errores de validacion se devuelven tal cual: su mensaje es el que ve
		// la persona (spec US-008, seccion 7). El repositorio solo se llama si el
		// dominio no encontro ninguno (CL-008-13).
		return sprint.Sprint{}, err
	}

	registrado, err := c.repo.RegistrarSprint(ctx, nuevo)
	if err != nil {
		return sprint.Sprint{}, fmt.Errorf("registrar el sprint en el proyecto %s: %w", proyectoID, err)
	}
	return registrado, nil
}
