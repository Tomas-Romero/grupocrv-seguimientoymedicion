package memoria

import (
	"context"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// Base cumple el puerto del caso de uso de US-001.
var _ app.RepositorioProyectos = (*Base)(nil)

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

// Proyecto devuelve los datos del proyecto registrado con ese ID. Lo usan los
// escenarios para verificar lo que quedo guardado: todavia no hay un caso de
// uso de lectura.
func (b *Base) Proyecto(id string) (proyecto.Proyecto, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	p, ok := b.datosProyectos[id]
	return p, ok
}
