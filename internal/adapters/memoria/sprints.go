package memoria

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Base cumple el puerto del caso de uso de US-008.
var _ app.RepositorioSprints = (*Base)(nil)

// FechasDelProyecto devuelve las fechas del proyecto y dice si existe. Cualquier
// ID que no este, incluido uno mal formado, es un proyecto inexistente y no un
// error. Solo conoce las fechas de los proyectos registrados con RegistrarProyecto.
func (b *Base) FechasDelProyecto(_ context.Context, proyectoID string) (time.Time, time.Time, bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	p, ok := b.datosProyectos[proyectoID]
	if !ok {
		return time.Time{}, time.Time{}, false, nil
	}
	return p.FechaInicio, p.FechaFin, true, nil
}

// RegistrarSprint guarda el sprint con el numero siguiente del proyecto y un ID
// nuevo. Replica lo que hace la persistencia (RN-008-14), con mu tomado en lugar
// del candado sobre el proyecto, asi dos altas simultaneas no pasan las dos:
//
//   - si el proyecto no esta, devuelve app.ErrProyectoInexistente;
//   - vuelve a verificar las fechas del sprint contra las del proyecto guardadas;
//   - verifica el ultimo sprint del proyecto (abierto o no posterior);
//   - es de todo o nada: si falla, no guarda nada.
func (b *Base) RegistrarSprint(_ context.Context, s sprint.Sprint) (sprint.Sprint, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	p, ok := b.datosProyectos[s.ProyectoID]
	if !ok {
		return sprint.Sprint{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, s.ProyectoID)
	}
	if err := sprint.VerificarDentroDelProyecto(s.FechaInicio, s.FechaFin, p.FechaInicio, p.FechaFin); err != nil {
		return sprint.Sprint{}, err
	}

	// El ultimo sprint es el de numero mas alto, que siempre es el ultimo de la lista.
	var ultimo *sprint.UltimoSprint
	if existentes := b.sprints[s.ProyectoID]; len(existentes) > 0 {
		u := existentes[len(existentes)-1]
		ultimo = &sprint.UltimoSprint{Numero: u.Numero, Estado: u.Estado, FechaFin: u.FechaFin}
	}
	if err := sprint.VerificarUltimoSprint(s.FechaInicio, ultimo); err != nil {
		return sprint.Sprint{}, err
	}

	s.ID = b.nuevoID()
	s.Numero = len(b.sprints[s.ProyectoID]) + 1
	b.sprints[s.ProyectoID] = append(b.sprints[s.ProyectoID], s)
	return s, nil
}

// Sprints devuelve una copia de los sprints del proyecto, en orden de alta. Un
// proyecto sin sprints, o que no existe, devuelve una lista vacia. Lo usan los
// escenarios para verificar lo que quedo guardado: todavia no hay casos de uso de
// lectura (US-012).
func (b *Base) Sprints(proyectoID string) []sprint.Sprint {
	b.mu.Lock()
	defer b.mu.Unlock()

	return slices.Clone(b.sprints[proyectoID])
}
