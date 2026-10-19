package memoria

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Base cumple el puerto de los casos de uso de US-010 (ObtenerItem es el de US-013 y
// ObtenerSprint, el de US-009).
var _ app.RepositorioMarcado = (*Base)(nil)

// ItemEnSprint dice si el item tiene una fila pendiente en ese sprint (RN-010-2). Las
// filas de otros sprints no cuentan, y un ID que no esta no esta en ningun sprint.
func (b *Base) ItemEnSprint(_ context.Context, itemID, sprintID string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.enSprint(itemID, sprintID), nil
}

// GuardarCambioDeEstado replica lo que hace la persistencia (RN-010-9), con mu tomado
// en lugar del candado sobre el sprint: verifica primero el sprint y despues el item,
// vuelve a verificar las reglas con lo guardado, aplica el cambio al item guardado y lo
// devuelve como quedo. Es de todo o nada.
func (b *Base) GuardarCambioDeEstado(_ context.Context, itemID, sprintID string, cambio app.CambioDeEstado) (backlog.ItemBacklog, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s, ok := b.sprintPorID(sprintID)
	if !ok {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, sprintID)
	}
	guardado := b.itemGuardado(itemID)
	if guardado == nil {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", app.ErrItemInexistente, itemID)
	}

	datos := sprint.DatosDelSprint{ProyectoID: s.ProyectoID, Nombre: s.Nombre(), Estado: s.Estado}
	if err := sprint.VerificarMarcado(datos, b.enSprint(itemID, sprintID)); err != nil {
		return backlog.ItemBacklog{}, err
	}

	// Solo se guardan el estado y la fecha: el cambio no toca lo demas.
	nuevo := cambio(copiaDe(*guardado))
	guardado.Estado = nuevo.Estado
	guardado.CompletadoEn = nuevo.CompletadoEn
	return copiaDe(*guardado), nil
}

// itemGuardado devuelve el item guardado, para cambiarlo en el lugar, o nil si no
// esta. Se llama con mu tomado.
func (b *Base) itemGuardado(itemID string) *backlog.ItemBacklog {
	for _, items := range b.items {
		for i := range items {
			if items[i].ID == itemID {
				return &items[i]
			}
		}
	}
	return nil
}

// enSprint dice si el item tiene una fila pendiente en ese sprint. Se llama con mu
// tomado.
func (b *Base) enSprint(itemID, sprintID string) bool {
	for _, a := range b.asignaciones[itemID] {
		if a.sprintID == sprintID && a.resultado == resultadoPendiente {
			return true
		}
	}
	return false
}
