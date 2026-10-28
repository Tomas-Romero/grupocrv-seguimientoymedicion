package memoria

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// Base cumple el puerto del caso de uso de US-023 (ExisteProyecto es el de US-005).
var _ app.RepositorioCambioDeEstadoDefecto = (*Base)(nil)

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

// ActualizarEstadoDefecto guarda el estado y el sprint de resolucion del defecto y nada mas
// (RN-023-9). Replica el UPDATE de Postgres (RN-023-11), condicionado al estado que leyo el caso de
// uso: si el defecto no esta en el proyecto devuelve ErrDefectoInexistente; si esta pero ya no
// tiene ese estado, ErrDefectoModificado; y si el sprint de resolucion no esta, que es lo que
// rechaza su clave foranea, ErrSprintInexistente. En ningun caso guarda nada.
func (b *Base) ActualizarEstadoDefecto(_ context.Context, d defecto.Defecto, estadoAnterior defecto.Estado) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	defectos := b.defectos[d.ProyectoID]
	for i := range defectos {
		if defectos[i].ID != d.ID {
			continue
		}
		if defectos[i].Estado != estadoAnterior {
			return fmt.Errorf("%w (id %s)", app.ErrDefectoModificado, d.ID)
		}
		if d.SprintResolucionID != "" && !b.sprintExiste(d.SprintResolucionID) {
			return fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, d.SprintResolucionID)
		}
		defectos[i].Estado = d.Estado
		defectos[i].SprintResolucionID = d.SprintResolucionID
		return nil
	}
	return fmt.Errorf("%w (id %s)", app.ErrDefectoInexistente, d.ID)
}
