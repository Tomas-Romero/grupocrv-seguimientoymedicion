package app

import (
	"context"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
)

// RepositorioRegistrarIntegrante es lo que el caso de uso de US-003 necesita de
// la persistencia.
type RepositorioRegistrarIntegrante interface {
	// ExisteProyecto dice si hay un proyecto con ese ID. Un ID mal formado es un
	// proyecto inexistente, no un error.
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)

	// RegistrarIntegrante guarda al integrante y lo devuelve con el ID que genera
	// la base. Si el proyecto ya no existe devuelve ErrProyectoInexistente; si el
	// email ya esta registrado en el proyecto, ErrEmailDuplicado. Es de todo o
	// nada: si falla, no queda nada guardado.
	RegistrarIntegrante(ctx context.Context, i integrante.Integrante) (integrante.Integrante, error)
}

// RegistrarIntegrante es el caso de uso de US-003: registrar un integrante en un
// proyecto existente.
type RegistrarIntegrante struct {
	repo RepositorioRegistrarIntegrante
}

// NuevoRegistrarIntegrante arma el caso de uso.
func NuevoRegistrarIntegrante(repo RepositorioRegistrarIntegrante) *RegistrarIntegrante {
	return &RegistrarIntegrante{repo: repo}
}

// Ejecutar registra al integrante.
//
// ESQUELETO DEL CICLO RED (US-003): todavia no hace nada. La implementacion real
// llega en el commit GREEN.
func (c *RegistrarIntegrante) Ejecutar(_ context.Context, _, _, _ string, _ integrante.Rol) (integrante.Integrante, error) {
	return integrante.Integrante{}, nil
}
