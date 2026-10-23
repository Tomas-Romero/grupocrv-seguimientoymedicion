package memoria

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// resultadoPendiente es el resultado de una fila de la relacion mientras el sprint esta
// abierto (RN-009-13).
const resultadoPendiente = "pendiente"

// asignacion es una fila de la relacion entre un item y un sprint.
type asignacion struct {
	sprintID  string
	resultado string
	// numero es el orden de alta de la fila en toda la base: las filas se guardan por
	// item, y con el numero se listan las de un sprint en orden de asignacion (US-011).
	numero int
}

// Base cumple el puerto del caso de uso de US-009 (ObtenerItem es el de US-013).
var _ app.RepositorioAsignaciones = (*Base)(nil)

// ObtenerSprint busca el sprint por su ID en todos los proyectos. Cualquier ID que no
// este, incluido uno mal formado, es un sprint inexistente y no un error.
func (b *Base) ObtenerSprint(_ context.Context, sprintID string) (sprint.Sprint, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s, ok := b.sprintPorID(sprintID)
	return s, ok, nil
}

// ItemEnSprintAbierto dice si el item tiene una fila de la relacion pendiente: su sprint
// actual (RN-009-6). Las filas de sprints cerrados no cuentan.
func (b *Base) ItemEnSprintAbierto(_ context.Context, itemID string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.enSprintAbierto(itemID), nil
}

// AsignarItemASprint replica lo que hace la persistencia (RN-009-11), con mu tomado en
// lugar de los bloqueos, asi dos asignaciones simultaneas no pasan las dos: verifica
// primero el sprint y despues el item, vuelve a verificar las reglas con los datos
// guardados y crea la fila pendiente. Es de todo o nada.
func (b *Base) AsignarItemASprint(_ context.Context, itemID, sprintID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	destino, ok := b.sprintPorID(sprintID)
	if !ok {
		return fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, sprintID)
	}
	item, ok := b.itemPorID(itemID)
	if !ok {
		return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, itemID)
	}

	_, estimado := item.StoryPoints.Valor()
	datosItem := sprint.DatosDelItem{
		ProyectoID:      item.ProyectoID,
		Completado:      item.Estado == backlog.EstadoCompletado,
		Estimado:        estimado,
		Criterios:       len(item.Criterios),
		EnSprintAbierto: b.enSprintAbierto(itemID),
	}
	datosSprint := sprint.DatosDelSprint{ProyectoID: destino.ProyectoID, Nombre: destino.Nombre(), Estado: destino.Estado}
	if err := sprint.VerificarAsignacion(datosItem, datosSprint); err != nil {
		return err
	}

	b.ultimaAsignacion++
	fila := asignacion{sprintID: sprintID, resultado: resultadoPendiente, numero: b.ultimaAsignacion}
	b.asignaciones[itemID] = append(b.asignaciones[itemID], fila)
	return nil
}

// SprintActualDelItem devuelve el sprint abierto en el que esta el item, si esta en uno.
// Lo usan los escenarios para verificar lo que quedo guardado: todavia no hay casos de
// uso de lectura.
func (b *Base) SprintActualDelItem(itemID string) (sprint.Sprint, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, a := range b.asignaciones[itemID] {
		if a.resultado == resultadoPendiente {
			return b.sprintPorID(a.sprintID)
		}
	}
	return sprint.Sprint{}, false
}

// sprintPorID busca el sprint en todos los proyectos. Se llama con mu tomado.
func (b *Base) sprintPorID(sprintID string) (sprint.Sprint, bool) {
	for _, sprints := range b.sprints {
		for _, s := range sprints {
			if s.ID == sprintID {
				return s, true
			}
		}
	}
	return sprint.Sprint{}, false
}

// itemPorID busca el item en todos los proyectos. Se llama con mu tomado.
func (b *Base) itemPorID(itemID string) (backlog.ItemBacklog, bool) {
	for _, items := range b.items {
		for _, item := range items {
			if item.ID == itemID {
				return copiaDe(item), true
			}
		}
	}
	return backlog.ItemBacklog{}, false
}

// enSprintAbierto dice si el item tiene una fila pendiente. Se llama con mu tomado.
func (b *Base) enSprintAbierto(itemID string) bool {
	for _, a := range b.asignaciones[itemID] {
		if a.resultado == resultadoPendiente {
			return true
		}
	}
	return false
}
