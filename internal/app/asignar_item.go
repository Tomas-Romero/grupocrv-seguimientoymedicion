package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// ErrSprintInexistente se devuelve cuando la operacion apunta a un sprint que no
// existe. Igual que ErrProyectoInexistente y ErrItemInexistente, tiene prioridad sobre
// los errores de validacion y nunca se une con ellos (RN-009-2).
var ErrSprintInexistente = errors.New("el sprint no existe")

// RepositorioAsignaciones es lo que el caso de uso de US-009 necesita de la
// persistencia.
type RepositorioAsignaciones interface {
	// ObtenerItem devuelve el item y si existe. Un ID mal formado es un item
	// inexistente, no un error.
	ObtenerItem(ctx context.Context, itemID string) (backlog.ItemBacklog, bool, error)

	// ObtenerSprint devuelve el sprint y si existe. Un ID mal formado es un sprint
	// inexistente, no un error.
	ObtenerSprint(ctx context.Context, sprintID string) (sprint.Sprint, bool, error)

	// ItemEnSprintAbierto dice si el item esta en el sprint abierto de su proyecto: tiene
	// una fila de la relacion con resultado pendiente (RN-009-13). Haber estado en
	// sprints cerrados no cuenta.
	ItemEnSprintAbierto(ctx context.Context, itemID string) (bool, error)

	// AsignarItemASprint registra la asignacion dentro de una transaccion, con los
	// pasos de RN-009-11: bloquea el sprint y el item, vuelve a verificar las reglas con
	// los datos leidos con los bloqueos y crea la fila de la relacion con resultado
	// pendiente. Si el item o el sprint desaparecieron, devuelve ErrItemInexistente o
	// ErrSprintInexistente; si las reglas ya no se cumplen, los errores de
	// sprint.VerificarAsignacion. Es de todo o nada.
	AsignarItemASprint(ctx context.Context, itemID, sprintID string) error
}

// Asignacion es el resultado de asignar un item a un sprint.
type Asignacion struct {
	ItemID   string
	SprintID string
}

// AsignarItemASprint es el caso de uso de US-009: asignar un item del backlog a un sprint.
type AsignarItemASprint struct {
	repo RepositorioAsignaciones
}

// NuevoAsignarItemASprint arma el caso de uso.
func NuevoAsignarItemASprint(repo RepositorioAsignaciones) *AsignarItemASprint {
	return &AsignarItemASprint{repo: repo}
}

// Ejecutar sigue el orden de la spec (seccion 2): busca el item, busca el sprint,
// valida con el dominio y registra la asignacion. Si el item o el sprint no existen
// responde ese error sin mirar lo demas, y nunca lo une con los de validacion
// (RN-009-1 y RN-009-2); si faltan los dos, solo ErrItemInexistente.
func (c *AsignarItemASprint) Ejecutar(ctx context.Context, itemID, sprintID string) (Asignacion, error) {
	item, existe, err := c.repo.ObtenerItem(ctx, itemID)
	if err != nil {
		return Asignacion{}, fmt.Errorf("consultar el item %s: %w", itemID, err)
	}
	if !existe {
		return Asignacion{}, fmt.Errorf("%w (id %s)", ErrItemInexistente, itemID)
	}

	destino, existe, err := c.repo.ObtenerSprint(ctx, sprintID)
	if err != nil {
		return Asignacion{}, fmt.Errorf("consultar el sprint %s: %w", sprintID, err)
	}
	if !existe {
		return Asignacion{}, fmt.Errorf("%w (id %s)", ErrSprintInexistente, sprintID)
	}

	enAbierto, err := c.repo.ItemEnSprintAbierto(ctx, itemID)
	if err != nil {
		return Asignacion{}, fmt.Errorf("consultar si el item %s ya esta en un sprint: %w", itemID, err)
	}

	// RN-009-8: el dominio valida todo y devuelve los errores unidos, tal cual: su
	// mensaje es el que ve la persona (spec US-009, seccion 7).
	if err := sprint.VerificarAsignacion(datosDelItem(item, enAbierto), datosDelSprint(destino)); err != nil {
		return Asignacion{}, err
	}

	if err := c.repo.AsignarItemASprint(ctx, itemID, sprintID); err != nil {
		return Asignacion{}, fmt.Errorf("asignar el item %s al sprint %s: %w", itemID, sprintID, err)
	}
	return Asignacion{ItemID: itemID, SprintID: sprintID}, nil
}

// datosDelItem adapta el item del backlog a los datos minimos que valida el dominio de
// sprint, que no usa ItemBacklog.
func datosDelItem(item backlog.ItemBacklog, enSprintAbierto bool) sprint.DatosDelItem {
	_, estimado := item.StoryPoints.Valor()
	return sprint.DatosDelItem{
		ProyectoID:      item.ProyectoID,
		Completado:      item.Estado == backlog.EstadoCompletado,
		Estimado:        estimado,
		Criterios:       len(item.Criterios),
		EnSprintAbierto: enSprintAbierto,
	}
}

// datosDelSprint adapta el sprint a los datos minimos que valida el dominio.
func datosDelSprint(s sprint.Sprint) sprint.DatosDelSprint {
	return sprint.DatosDelSprint{ProyectoID: s.ProyectoID, Nombre: s.Nombre(), Estado: s.Estado}
}
