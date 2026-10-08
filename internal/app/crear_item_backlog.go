package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// RepositorioBacklog es lo que el caso de uso necesita de la persistencia.
type RepositorioBacklog interface {
	// ExisteProyecto dice si el proyecto existe. Un ID mal formado es un
	// proyecto inexistente, no un error.
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)

	// RegistrarItem guarda el item en el proyecto y lo devuelve con su ID y su
	// numero correlativo. Si el proyecto ya no existe, devuelve
	// ErrProyectoInexistente.
	RegistrarItem(ctx context.Context, proyectoID string, item backlog.ItemBacklog) (backlog.ItemBacklog, error)
}

// CrearItemBacklog es el caso de uso de US-005: crear un item en el Product
// Backlog de un proyecto.
type CrearItemBacklog struct {
	repo  RepositorioBacklog
	ahora func() time.Time
}

// NuevoCrearItemBacklog arma el caso de uso. ahora es el reloj: se recibe por
// parametro para que el dominio no lea la hora y los tests no dependan del dia.
func NuevoCrearItemBacklog(repo RepositorioBacklog, ahora func() time.Time) *CrearItemBacklog {
	return &CrearItemBacklog{repo: repo, ahora: ahora}
}

// Ejecutar sigue el orden de la spec (seccion 2): primero verifica que el
// proyecto exista, despues valida los datos con el dominio y al final registra
// el item. Si el proyecto no existe responde ErrProyectoInexistente sin mirar
// los datos (RN-005-9).
func (c *CrearItemBacklog) Ejecutar(ctx context.Context, proyectoID string, datos backlog.DatosItem) (backlog.ItemBacklog, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}

	item, err := backlog.NuevoItem(datos, c.ahora())
	if err != nil {
		// Los errores de validacion se devuelven tal cual: su mensaje es el que
		// ve la persona (spec US-005, seccion 7).
		return backlog.ItemBacklog{}, err
	}

	registrado, err := c.repo.RegistrarItem(ctx, proyectoID, item)
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("registrar el item en el proyecto %s: %w", proyectoID, err)
	}
	return registrado, nil
}
