package memoria

import (
	"context"
	"fmt"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// Base cumple el puerto del caso de uso de US-005.
var _ app.RepositorioBacklog = (*Base)(nil)

// ExisteProyecto dice si el proyecto esta en la base. Cualquier ID que no este,
// incluido uno mal formado, es un proyecto inexistente y no un error.
func (b *Base) ExisteProyecto(_ context.Context, proyectoID string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.proyectos[proyectoID], nil
}

// RegistrarItem guarda el item en el proyecto con el numero siguiente (el
// maximo actual mas 1, RN-005-11) y un ID nuevo. Si el proyecto no esta en la
// base, devuelve app.ErrProyectoInexistente y no guarda nada.
func (b *Base) RegistrarItem(_ context.Context, proyectoID string, item backlog.ItemBacklog) (backlog.ItemBacklog, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.proyectos[proyectoID] {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, proyectoID)
	}

	numero := 1
	for _, existente := range b.items[proyectoID] {
		if existente.Numero >= numero {
			numero = existente.Numero + 1
		}
	}
	item.ID = b.nuevoID()
	item.Numero = numero
	item.ProyectoID = proyectoID
	item.Criterios = slices.Clone(item.Criterios)

	b.items[proyectoID] = append(b.items[proyectoID], item)
	return copiaDe(item), nil
}

// Items devuelve una copia de los items del proyecto, en orden de alta. Lo usan
// los escenarios para verificar lo que quedo guardado: todavia no hay casos de
// uso de lectura (US-007).
func (b *Base) Items(proyectoID string) []backlog.ItemBacklog {
	b.mu.Lock()
	defer b.mu.Unlock()

	items := make([]backlog.ItemBacklog, 0, len(b.items[proyectoID]))
	for _, item := range b.items[proyectoID] {
		items = append(items, copiaDe(item))
	}
	return items
}

// copiaDe devuelve el item con su propia lista de criterios, para que quien lo
// recibe no pueda cambiar lo guardado.
func copiaDe(item backlog.ItemBacklog) backlog.ItemBacklog {
	item.Criterios = slices.Clone(item.Criterios)
	return item
}
