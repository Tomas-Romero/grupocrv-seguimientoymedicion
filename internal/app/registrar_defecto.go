package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// RepositorioDefectos es lo que el caso de uso de US-022 necesita de la persistencia.
type RepositorioDefectos interface {
	// RegistrarDefecto guarda el defecto y lo devuelve con el ID que genera la base.
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

// Ejecutar valida los datos con el dominio y registra el defecto. No recibe el estado ni
// el sprint de resolucion: todo defecto nace abierto y sin resolver (RN-022-3).
func (c *RegistrarDefecto) Ejecutar(ctx context.Context, proyectoID, itemID, sprintDeteccionID, descripcion string, severidad defecto.Severidad) (defecto.Defecto, error) {
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
