package memoria

import (
	"context"
	"fmt"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// Base cumple el puerto del caso de uso de US-022 (ExisteProyecto y ExisteItemEnProyecto
// son los mismos que usan los de US-005 y US-019).
var _ app.RepositorioDefectos = (*Base)(nil)

// ExisteSprintEnProyecto dice si el sprint esta en ese proyecto (RN-022-5). Uno de otro
// proyecto, un ID que no esta, uno mal formado o uno vacio no cuentan.
func (b *Base) ExisteSprintEnProyecto(_ context.Context, proyectoID, sprintID string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.sprintEnProyecto(proyectoID, sprintID), nil
}

// RegistrarDefecto guarda el defecto con un ID nuevo, como el RETURNING de Postgres
// (RN-022-10). Replica las claves foraneas de la tabla: si el proyecto, la historia o el
// sprint de deteccion no estan, devuelve el error de inexistencia que corresponde y no
// guarda nada. Defectos iguales son defectos distintos (RN-022-8), y no mira el estado de
// la historia (RN-022-7).
func (b *Base) RegistrarDefecto(_ context.Context, d defecto.Defecto) (defecto.Defecto, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch {
	case !b.proyectos[d.ProyectoID]:
		return defecto.Defecto{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, d.ProyectoID)
	case !b.itemExiste(d.ItemID):
		return defecto.Defecto{}, fmt.Errorf("%w (id %s)", app.ErrItemInexistente, d.ItemID)
	case !b.sprintExiste(d.SprintDeteccionID):
		return defecto.Defecto{}, fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, d.SprintDeteccionID)
	}

	d.ID = b.nuevoID()
	b.defectos[d.ProyectoID] = append(b.defectos[d.ProyectoID], d)
	return d, nil
}

// Defectos devuelve una copia de los defectos del proyecto, en orden de alta. Un proyecto
// sin defectos, o que no existe, devuelve una lista vacia. Lo usan los escenarios de US-022
// para verificar lo que quedo guardado; listar defectos es US-024.
func (b *Base) Defectos(proyectoID string) []defecto.Defecto {
	b.mu.Lock()
	defer b.mu.Unlock()

	return slices.Clone(b.defectos[proyectoID])
}

// sprintEnProyecto se llama con mu tomado.
func (b *Base) sprintEnProyecto(proyectoID, sprintID string) bool {
	for _, s := range b.sprints[proyectoID] {
		if s.ID == sprintID {
			return true
		}
	}
	return false
}

// sprintExiste busca el sprint en todos los proyectos. Se llama con mu tomado.
func (b *Base) sprintExiste(sprintID string) bool {
	for proyectoID := range b.sprints {
		if b.sprintEnProyecto(proyectoID, sprintID) {
			return true
		}
	}
	return false
}
