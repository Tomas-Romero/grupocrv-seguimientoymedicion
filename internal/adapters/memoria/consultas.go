package memoria

import (
	"context"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// Base cumple los puertos de lectura de las pantallas de proyectos y backlog (T-011).
// BuscarProyecto, ExisteProyecto y ObtenerItem ya existen, de US-002, US-005 y US-013.
var (
	_ app.RepositorioListarProyectos   = (*Base)(nil)
	_ app.RepositorioConsultarProyecto = (*Base)(nil)
	_ app.RepositorioListarBacklog     = (*Base)(nil)
	_ app.RepositorioConsultarItem     = (*Base)(nil)
)

// ListarProyectos devuelve los proyectos registrados con RegistrarProyecto, sin orden
// (el orden lo pone el caso de uso). Un proyecto que solo se agrego con AgregarProyecto
// no tiene datos y no se lista.
func (b *Base) ListarProyectos(_ context.Context) ([]proyecto.Proyecto, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	proyectos := make([]proyecto.Proyecto, 0, len(b.datosProyectos))
	for _, p := range b.datosProyectos {
		proyectos = append(proyectos, p)
	}
	return proyectos, nil
}

// ListarIntegrantes devuelve los integrantes del proyecto, sin orden. Un proyecto que no
// existe, o sin integrantes, devuelve una lista vacia: quien lo llama ya verifico que el
// proyecto exista.
func (b *Base) ListarIntegrantes(_ context.Context, proyectoID string) ([]integrante.Integrante, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return slices.Clone(b.integrantes[proyectoID]), nil
}

// ListarItems devuelve los items del proyecto, sin orden, cada uno con su propia lista de
// criterios. Un proyecto que no existe, o sin items, devuelve una lista vacia.
func (b *Base) ListarItems(_ context.Context, proyectoID string) ([]backlog.ItemBacklog, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	items := make([]backlog.ItemBacklog, 0, len(b.items[proyectoID]))
	for _, item := range b.items[proyectoID] {
		items = append(items, copiaDe(item))
	}
	return items, nil
}
