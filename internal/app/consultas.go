package app

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// Lecturas de las pantallas de proyectos, integrantes y backlog (spec T-011). Los
// repositorios devuelven los datos sin ordenar y el orden lo pone el caso de uso, asi la
// memoria y Postgres se comportan igual (RN-T011-2).

// RepositorioListarProyectos es lo que ListarProyectos necesita de la persistencia.
type RepositorioListarProyectos interface {
	// ListarProyectos devuelve todos los proyectos, en cualquier orden.
	ListarProyectos(ctx context.Context) ([]proyecto.Proyecto, error)
}

// ListarProyectos es la lectura de la lista de proyectos.
type ListarProyectos struct {
	repo RepositorioListarProyectos
}

// NuevoListarProyectos arma el caso de uso.
func NuevoListarProyectos(repo RepositorioListarProyectos) *ListarProyectos {
	return &ListarProyectos{repo: repo}
}

// Ejecutar devuelve los proyectos por fecha de inicio descendente; el empate se
// resuelve por nombre sin distinguir mayusculas y despues por ID (RN-T011-3). Sin
// proyectos devuelve una lista vacia, no un error.
func (c *ListarProyectos) Ejecutar(ctx context.Context) ([]proyecto.Proyecto, error) {
	proyectos, err := c.repo.ListarProyectos(ctx)
	if err != nil {
		return nil, fmt.Errorf("listar los proyectos: %w", err)
	}
	sort.SliceStable(proyectos, func(i, j int) bool {
		a, b := proyectos[i], proyectos[j]
		if !a.FechaInicio.Equal(b.FechaInicio) {
			return a.FechaInicio.After(b.FechaInicio)
		}
		if na, nb := strings.ToLower(a.Nombre), strings.ToLower(b.Nombre); na != nb {
			return na < nb
		}
		return a.ID < b.ID
	})
	return proyectos, nil
}

// RepositorioConsultarProyecto es lo que ConsultarProyecto necesita de la persistencia.
type RepositorioConsultarProyecto interface {
	// BuscarProyecto devuelve el proyecto o ErrProyectoInexistente (es el de US-002).
	BuscarProyecto(ctx context.Context, proyectoID string) (proyecto.Proyecto, error)
	// ListarIntegrantes devuelve los integrantes del proyecto, en cualquier orden.
	ListarIntegrantes(ctx context.Context, proyectoID string) ([]integrante.Integrante, error)
}

// DetalleDeProyecto es un proyecto con sus integrantes.
type DetalleDeProyecto struct {
	Proyecto    proyecto.Proyecto
	Integrantes []integrante.Integrante
}

// ConsultarProyecto es la lectura de un proyecto con sus integrantes.
type ConsultarProyecto struct {
	repo RepositorioConsultarProyecto
}

// NuevoConsultarProyecto arma el caso de uso.
func NuevoConsultarProyecto(repo RepositorioConsultarProyecto) *ConsultarProyecto {
	return &ConsultarProyecto{repo: repo}
}

// Ejecutar devuelve el proyecto y sus integrantes por nombre sin distinguir
// mayusculas y despues por email (RN-T011-5). Un proyecto que no existe es
// ErrProyectoInexistente.
func (c *ConsultarProyecto) Ejecutar(ctx context.Context, proyectoID string) (DetalleDeProyecto, error) {
	p, err := c.repo.BuscarProyecto(ctx, proyectoID)
	if err != nil {
		return DetalleDeProyecto{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	integrantes, err := c.repo.ListarIntegrantes(ctx, proyectoID)
	if err != nil {
		return DetalleDeProyecto{}, fmt.Errorf("listar los integrantes del proyecto %s: %w", proyectoID, err)
	}
	sort.SliceStable(integrantes, func(i, j int) bool {
		a, b := integrantes[i], integrantes[j]
		if na, nb := strings.ToLower(a.Nombre), strings.ToLower(b.Nombre); na != nb {
			return na < nb
		}
		return a.Email < b.Email
	})
	return DetalleDeProyecto{Proyecto: p, Integrantes: integrantes}, nil
}

// RepositorioListarBacklog es lo que ListarBacklog necesita de la persistencia.
type RepositorioListarBacklog interface {
	// ExisteProyecto dice si el proyecto existe (es la lectura de US-005).
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)
	// ListarItems devuelve los items del proyecto, en cualquier orden.
	ListarItems(ctx context.Context, proyectoID string) ([]backlog.ItemBacklog, error)
}

// ListarBacklog es la lectura del Product Backlog de un proyecto.
type ListarBacklog struct {
	repo RepositorioListarBacklog
}

// NuevoListarBacklog arma el caso de uso.
func NuevoListarBacklog(repo RepositorioListarBacklog) *ListarBacklog {
	return &ListarBacklog{repo: repo}
}

// ordenDePrioridad pone must primero y wont ultimo; una prioridad que no conoce va al final.
func ordenDePrioridad(p backlog.Prioridad) int {
	switch p {
	case backlog.PrioridadMust:
		return 0
	case backlog.PrioridadShould:
		return 1
	case backlog.PrioridadCould:
		return 2
	case backlog.PrioridadWont:
		return 3
	default:
		return 4
	}
}

// Ejecutar devuelve el backlog por prioridad (must, should, could, wont) y, dentro de
// cada una, por numero (RN-T011-4). Un backlog vacio no es un error; un proyecto que no
// existe es ErrProyectoInexistente.
func (c *ListarBacklog) Ejecutar(ctx context.Context, proyectoID string) ([]backlog.ItemBacklog, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return nil, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}
	items, err := c.repo.ListarItems(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar el backlog del proyecto %s: %w", proyectoID, err)
	}
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if oa, ob := ordenDePrioridad(a.Prioridad), ordenDePrioridad(b.Prioridad); oa != ob {
			return oa < ob
		}
		return a.Numero < b.Numero
	})
	return items, nil
}

// RepositorioConsultarItem es lo que ConsultarItem necesita de la persistencia.
type RepositorioConsultarItem interface {
	// ObtenerItem devuelve el item y si existe (es la lectura de US-013).
	ObtenerItem(ctx context.Context, itemID string) (backlog.ItemBacklog, bool, error)
}

// ConsultarItem es la lectura de una historia dentro de su proyecto.
type ConsultarItem struct {
	repo RepositorioConsultarItem
}

// NuevoConsultarItem arma el caso de uso.
func NuevoConsultarItem(repo RepositorioConsultarItem) *ConsultarItem {
	return &ConsultarItem{repo: repo}
}

// Ejecutar devuelve la historia solo si es del proyecto (RN-T011-11): una historia de
// otro proyecto, o que no existe, es ErrItemInexistente.
func (c *ConsultarItem) Ejecutar(ctx context.Context, proyectoID, itemID string) (backlog.ItemBacklog, error) {
	item, existe, err := c.repo.ObtenerItem(ctx, itemID)
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("consultar la historia %s: %w", itemID, err)
	}
	if !existe || item.ProyectoID != proyectoID {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", ErrItemInexistente, itemID)
	}
	return item, nil
}
