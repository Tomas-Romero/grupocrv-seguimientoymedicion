package memoria

import (
	"context"
	"fmt"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Base cumple el puerto del caso de uso de US-011 (ObtenerSprint es el de US-009).
var _ app.RepositorioCierre = (*Base)(nil)

// HistoriasDelSprint devuelve las historias con una fila pendiente en el sprint, en
// orden de asignacion, y si cada una esta completada.
func (b *Base) HistoriasDelSprint(_ context.Context, sprintID string) ([]sprint.HistoriaDelCierre, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.historiasDelSprint(sprintID), nil
}

// GuardarCierre replica lo que hace la persistencia (RN-011-8), con mu tomado en lugar
// del candado sobre el sprint: verifica que el sprint este, calcula los resultados con
// sprint.Cerrar y, si puede cerrar, guarda el resultado de cada fila (sin borrar
// ninguna), pasa a pendiente las historias no completadas y deja el sprint cerrado. Es
// de todo o nada: si Cerrar rechaza, no cambia nada.
func (b *Base) GuardarCierre(_ context.Context, sprintID string) ([]sprint.ResultadoDelCierre, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	guardado := b.sprintGuardado(sprintID)
	if guardado == nil {
		return nil, fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, sprintID)
	}
	datos := sprint.DatosDelSprint{ProyectoID: guardado.ProyectoID, Nombre: guardado.Nombre(), Estado: guardado.Estado}
	resultados, err := sprint.Cerrar(datos, b.historiasDelSprint(sprintID))
	if err != nil {
		return nil, err
	}

	for _, r := range resultados {
		filas := b.asignaciones[r.ItemID]
		for i := range filas {
			if filas[i].sprintID == sprintID && filas[i].resultado == resultadoPendiente {
				filas[i].resultado = string(r.Resultado)
			}
		}
		// RN-011-6: la no completada vuelve al backlog como pendiente, tambien si estaba
		// en progreso.
		if r.Resultado == sprint.ResultadoNoCompletado {
			b.itemGuardado(r.ItemID).Estado = backlog.EstadoPendiente
		}
	}
	guardado.Estado = sprint.EstadoCerrado
	return resultados, nil
}

// FilaDelSprint es una historia que estuvo en un sprint, con su resultado en el y sus
// Story Points actuales.
type FilaDelSprint struct {
	ItemID      string
	StoryPoints backlog.StoryPoints
	Resultado   sprint.Resultado
}

// FilasDelSprint devuelve las historias que estuvieron en el sprint, abierto o cerrado,
// con su resultado, en orden de asignacion. Lo usan los escenarios para calcular las
// metricas de los sprints cerrados (US-025 y US-026): todavia no hay casos de uso de
// lectura.
func (b *Base) FilasDelSprint(sprintID string) []FilaDelSprint {
	b.mu.Lock()
	defer b.mu.Unlock()

	var filas []FilaDelSprint
	for _, a := range b.filasDelSprint(sprintID) {
		item := b.itemGuardado(a.itemID)
		filas = append(filas, FilaDelSprint{ItemID: a.itemID, StoryPoints: item.StoryPoints, Resultado: sprint.Resultado(a.resultado)})
	}
	return filas
}

// filaConItem es una fila de la relacion con el item al que pertenece.
type filaConItem struct {
	itemID    string
	resultado string
	numero    int
}

// filasDelSprint devuelve las filas del sprint en orden de asignacion. Se llama con mu
// tomado.
func (b *Base) filasDelSprint(sprintID string) []filaConItem {
	var filas []filaConItem
	for itemID, asignaciones := range b.asignaciones {
		for _, a := range asignaciones {
			if a.sprintID == sprintID {
				filas = append(filas, filaConItem{itemID: itemID, resultado: a.resultado, numero: a.numero})
			}
		}
	}
	slices.SortFunc(filas, func(x, y filaConItem) int { return x.numero - y.numero })
	return filas
}

// historiasDelSprint son las filas pendientes del sprint con si su item esta
// completado. Se llama con mu tomado.
func (b *Base) historiasDelSprint(sprintID string) []sprint.HistoriaDelCierre {
	var historias []sprint.HistoriaDelCierre
	for _, f := range b.filasDelSprint(sprintID) {
		if f.resultado != resultadoPendiente {
			continue
		}
		item := b.itemGuardado(f.itemID)
		historias = append(historias, sprint.HistoriaDelCierre{ItemID: f.itemID, Completada: item.Estado == backlog.EstadoCompletado})
	}
	return historias
}

// sprintGuardado devuelve el sprint guardado, para cambiarlo en el lugar, o nil si no
// esta. Se llama con mu tomado.
func (b *Base) sprintGuardado(sprintID string) *sprint.Sprint {
	for proyectoID := range b.sprints {
		sprints := b.sprints[proyectoID]
		for i := range sprints {
			if sprints[i].ID == sprintID {
				return &sprints[i]
			}
		}
	}
	return nil
}
