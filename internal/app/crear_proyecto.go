package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// RepositorioProyectos es lo que el caso de uso necesita de la persistencia.
type RepositorioProyectos interface {
	// RegistrarProyecto guarda el proyecto y lo devuelve con el ID que genera
	// la base (RN-001-6).
	RegistrarProyecto(ctx context.Context, p proyecto.Proyecto) (proyecto.Proyecto, error)
}

// CrearProyecto es el caso de uso de US-001: crear un proyecto.
type CrearProyecto struct {
	repo RepositorioProyectos
}

// NuevoCrearProyecto arma el caso de uso.
func NuevoCrearProyecto(repo RepositorioProyectos) *CrearProyecto {
	return &CrearProyecto{repo: repo}
}

// Ejecutar valida los datos con el dominio y, si son validos, registra el
// proyecto. Los errores se envuelven con el nombre del proyecto (spec US-001,
// seccion 7) sin perder el original: errors.Is sigue reconociendo cada error
// de validacion unido y el mensaje de cada uno queda dentro del resultante.
func (c *CrearProyecto) Ejecutar(ctx context.Context, nombre, descripcion string, fechaInicio, fechaFin time.Time) (proyecto.Proyecto, error) {
	nuevo, err := proyecto.Nuevo(nombre, descripcion, fechaInicio, fechaFin)
	if err != nil {
		return proyecto.Proyecto{}, fmt.Errorf("crear el proyecto %q: %w", nombre, err)
	}

	registrado, err := c.repo.RegistrarProyecto(ctx, nuevo)
	if err != nil {
		return proyecto.Proyecto{}, fmt.Errorf("registrar el proyecto %q: %w", nuevo.Nombre, err)
	}
	return registrado, nil
}