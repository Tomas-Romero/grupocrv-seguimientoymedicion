package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// RepositorioDefectos es lo que el caso de uso de US-022 necesita de la persistencia.
type RepositorioDefectos interface {
	// ExisteProyecto dice si el proyecto existe. Un ID mal formado es un proyecto
	// inexistente, no un error.
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)

	// ExisteItemEnProyecto dice si la historia existe y es de ese proyecto: una
	// historia de otro proyecto no cuenta. Un ID vacio o mal formado es inexistente.
	ExisteItemEnProyecto(ctx context.Context, proyectoID, itemID string) (bool, error)

	// ExisteSprintEnProyecto dice si el sprint existe y es de ese proyecto: uno de otro
	// proyecto no cuenta. Un ID vacio o mal formado es inexistente.
	ExisteSprintEnProyecto(ctx context.Context, proyectoID, sprintID string) (bool, error)

	// RegistrarDefecto guarda el defecto y lo devuelve con el ID que genera la base. Si
	// el proyecto, la historia o el sprint desaparecieron despues de verificarlos,
	// devuelve el error de inexistencia que corresponda.
	RegistrarDefecto(ctx context.Context, d defecto.Defecto) (defecto.Defecto, error)
}

// RegistrarDefecto es el caso de uso de US-022: registrar un defecto con su severidad,
// la historia relacionada y el sprint de deteccion.
type RegistrarDefecto struct {
	repo RepositorioDefectos
}

// NuevoRegistrarDefecto arma el caso de uso. No recibe reloj: la fecha de carga la pone
// la base.
func NuevoRegistrarDefecto(repo RepositorioDefectos) *RegistrarDefecto {
	return &RegistrarDefecto{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): verifica que el proyecto exista,
// despues que la historia y el sprint de deteccion existan y sean de ese proyecto,
// despues valida los datos con el dominio y al final registra. El primero que falla se
// devuelve y no se une con los errores de validacion (RN-022-5). No recibe el estado ni
// el sprint de resolucion: todo defecto nace abierto y sin resolver (RN-022-3).
func (c *RegistrarDefecto) Ejecutar(ctx context.Context, proyectoID, itemID, sprintDeteccionID, descripcion string, severidad defecto.Severidad) (defecto.Defecto, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return defecto.Defecto{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return defecto.Defecto{}, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}

	existe, err = c.repo.ExisteItemEnProyecto(ctx, proyectoID, itemID)
	if err != nil {
		return defecto.Defecto{}, fmt.Errorf("consultar la historia %s del proyecto %s: %w", itemID, proyectoID, err)
	}
	if !existe {
		return defecto.Defecto{}, fmt.Errorf("%w (id %s)", ErrItemInexistente, itemID)
	}

	// RN-022-6: el sprint de deteccion es obligatorio y puede ser cualquiera del proyecto;
	// uno sin elegir llega como un ID vacio, que el repositorio responde como inexistente.
	existe, err = c.repo.ExisteSprintEnProyecto(ctx, proyectoID, sprintDeteccionID)
	if err != nil {
		return defecto.Defecto{}, fmt.Errorf("consultar el sprint %s del proyecto %s: %w", sprintDeteccionID, proyectoID, err)
	}
	if !existe {
		return defecto.Defecto{}, fmt.Errorf("%w (id %s)", ErrSprintInexistente, sprintDeteccionID)
	}

	// RN-022-4: el dominio valida todo y devuelve los errores unidos, tal cual: su mensaje
	// es el que ve la persona (spec US-022, seccion 7).
	nuevo, err := defecto.Nuevo(proyectoID, itemID, sprintDeteccionID, descripcion, severidad)
	if err != nil {
		return defecto.Defecto{}, err
	}

	registrado, err := c.repo.RegistrarDefecto(ctx, nuevo)
	if err != nil {
		return defecto.Defecto{}, fmt.Errorf("registrar el defecto en el proyecto %s: %w", proyectoID, err)
	}
	return registrado, nil
}
