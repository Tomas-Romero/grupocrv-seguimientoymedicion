package memoria

import (
	"context"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// DefectoEnProyecto devuelve el defecto guardado y dice si existe y es de ese proyecto
// (RN-023-7). Uno de otro proyecto, un ID que no esta o uno vacio no cuentan.
func (b *Base) DefectoEnProyecto(_ context.Context, proyectoID, defectoID string) (defecto.Defecto, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, d := range b.defectos[proyectoID] {
		if d.ID == defectoID {
			return d, true, nil
		}
	}
	return defecto.Defecto{}, false, nil
}

// NumeroDeSprintEnProyecto devuelve el numero (sprint.Numero) del sprint y dice si esta en ese
// proyecto (RN-023-7, RN-023-8). Uno de otro proyecto, un ID que no esta o uno vacio no cuentan.
func (b *Base) NumeroDeSprintEnProyecto(_ context.Context, proyectoID, sprintID string) (int, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, s := range b.sprints[proyectoID] {
		if s.ID == sprintID {
			return s.Numero, true, nil
		}
	}
	return 0, false, nil
}
