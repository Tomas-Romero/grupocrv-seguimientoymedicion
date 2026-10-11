package memoria

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// Base cumple los puertos de los casos de uso de US-001 y US-002.
var (
	_ app.RepositorioProyectos         = (*Base)(nil)
	_ app.RepositorioModificarProyecto = (*Base)(nil)
)

// RegistrarProyecto guarda el proyecto con un ID nuevo, como el RETURNING de
// Postgres (RN-001-6). Lo registra en el mismo mapa que usa ExisteProyecto,
// asi el caso de uso de backlog (US-005) ve los proyectos que se crean aca.
func (b *Base) RegistrarProyecto(_ context.Context, p proyecto.Proyecto) (proyecto.Proyecto, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	p.ID = b.nuevoID()
	b.proyectos[p.ID] = true
	b.datosProyectos[p.ID] = p
	return p, nil
}

// BuscarProyecto devuelve los datos del proyecto con ese ID (US-002). Cualquier
// ID que no este, incluido uno vacio o mal formado, es un proyecto inexistente
// y devuelve app.ErrProyectoInexistente. Solo encuentra los proyectos
// registrados con RegistrarProyecto: los de AgregarProyecto no tienen datos.
func (b *Base) BuscarProyecto(_ context.Context, proyectoID string) (proyecto.Proyecto, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	p, ok := b.datosProyectos[proyectoID]
	if !ok {
		return proyecto.Proyecto{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, proyectoID)
	}
	return p, nil
}

// ActualizarProyecto reemplaza los datos del proyecto con el ID de p (US-002).
// No cambia el ID ni toca lo asociado al proyecto (RN-002-4 y RN-002-7). Si el
// proyecto no esta en la base, devuelve app.ErrProyectoInexistente y no guarda
// nada.
func (b *Base) ActualizarProyecto(_ context.Context, p proyecto.Proyecto) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.datosProyectos[p.ID]; !ok {
		return fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, p.ID)
	}
	b.datosProyectos[p.ID] = p
	return nil
}

// Proyecto devuelve los datos del proyecto registrado con ese ID. Lo usan los
// escenarios para verificar lo que quedo guardado: todavia no hay un caso de
// uso de lectura.
func (b *Base) Proyecto(id string) (proyecto.Proyecto, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	p, ok := b.datosProyectos[id]
	return p, ok
}