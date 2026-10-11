package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// RepositorioModificarProyecto es lo que el caso de uso de US-002 necesita de la
// persistencia.
type RepositorioModificarProyecto interface {
	// BuscarProyecto devuelve el proyecto con ese ID. Si no existe, un ID mal
	// formado incluido, devuelve ErrProyectoInexistente.
	BuscarProyecto(ctx context.Context, proyectoID string) (proyecto.Proyecto, error)

	// ActualizarProyecto reemplaza el nombre, la descripcion y las fechas del
	// proyecto con el ID de p. Si el proyecto ya no existe, devuelve
	// ErrProyectoInexistente.
	ActualizarProyecto(ctx context.Context, p proyecto.Proyecto) error
}

// ModificarProyecto es el caso de uso de US-002: modificar los datos de un
// proyecto existente.
type ModificarProyecto struct {
	repo RepositorioModificarProyecto
}

// NuevoModificarProyecto arma el caso de uso.
func NuevoModificarProyecto(repo RepositorioModificarProyecto) *ModificarProyecto {
	return &ModificarProyecto{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): primero busca el proyecto,
// despues valida los datos con el dominio y al final lo actualiza. Si el
// proyecto no existe responde ErrProyectoInexistente sin mirar los datos
// (RN-002-1). Los errores se envuelven con el proyecto (spec US-002, seccion 7)
// sin perder el original: errors.Is sigue reconociendo cada error unido.
func (c *ModificarProyecto) Ejecutar(ctx context.Context, proyectoID, nombre, descripcion string, fechaInicio, fechaFin time.Time) (proyecto.Proyecto, error) {
	actual, err := c.repo.BuscarProyecto(ctx, proyectoID)
	if err != nil {
		return proyecto.Proyecto{}, fmt.Errorf("modificar el proyecto %s: %w", proyectoID, err)
	}

	// RN-002-2: la misma validacion que al crear; RN-002-5: todo o nada.
	modificado, err := proyecto.Modificar(actual, nombre, descripcion, fechaInicio, fechaFin)
	if err != nil {
		return proyecto.Proyecto{}, fmt.Errorf("modificar el proyecto %s: %w", proyectoID, err)
	}

	if err := c.repo.ActualizarProyecto(ctx, modificado); err != nil {
		return proyecto.Proyecto{}, fmt.Errorf("actualizar el proyecto %s: %w", proyectoID, err)
	}
	return modificado, nil
}