package app

import (
	"context"
	"fmt"

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

// Ejecutar sigue el orden de la spec (seccion 2): primero verifica que el
// proyecto exista, despues valida los datos con el dominio y al final registra al
// integrante.
//
// RN-003-6: si el proyecto no existe responde ErrProyectoInexistente sin mirar
// los datos y sin unirlo con los errores de validacion. Los errores del dominio
// se devuelven tal cual, sin prefijos: su mensaje es el que ve la persona. Los
// del repositorio se envuelven con el proyecto sin perder el original, asi
// errors.Is sigue reconociendo ErrProyectoInexistente y ErrEmailDuplicado.
func (c *RegistrarIntegrante) Ejecutar(ctx context.Context, proyectoID, nombre, email string, rol integrante.Rol) (integrante.Integrante, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return integrante.Integrante{}, fmt.Errorf("registrar un integrante en el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return integrante.Integrante{}, fmt.Errorf("registrar un integrante en el proyecto %s: %w", proyectoID, ErrProyectoInexistente)
	}

	// RN-003-8: el dominio valida todo y devuelve los errores unidos.
	nuevo, err := integrante.Nuevo(proyectoID, nombre, email, rol)
	if err != nil {
		return integrante.Integrante{}, err
	}

	// RN-003-9: todo o nada, lo garantiza el repositorio.
	registrado, err := c.repo.RegistrarIntegrante(ctx, nuevo)
	if err != nil {
		return integrante.Integrante{}, fmt.Errorf("registrar un integrante en el proyecto %s: %w", proyectoID, err)
	}
	return registrado, nil
}
