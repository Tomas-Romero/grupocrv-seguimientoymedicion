package features_test

import (
	"fmt"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// metricasEnMemoria es el area Metricas de los escenarios: guarda los items de
// cada sprint que arma el contexto y los pasa al dominio al consultar. No hay
// caso de uso intermedio porque US-025 y US-026 son calculos puros del dominio.
//
// Ademas calcula sobre los sprints que se cerraron de verdad en la base del escenario
// (US-011): cada historia que estuvo en el sprint, con sus Story Points y si quedo
// completada en el (spec US-011, seccion 5).
type metricasEnMemoria struct {
	escenario *escenario
	items     map[string][]metricas.ItemDelSprint
	cerrados  []metricas.SprintCerrado
	horas     map[string][]float64
}

var _ steps.Metricas = (*metricasEnMemoria)(nil)

func nuevasMetricasEnMemoria(e *escenario) *metricasEnMemoria {
	return &metricasEnMemoria{escenario: e, items: map[string][]metricas.ItemDelSprint{}, horas: map[string][]float64{}}
}

// sprintsCerradosDeLaBase devuelve los sprints cerrados de la base del escenario, con
// sus items, por proyecto (en orden de ID) y en orden de numero.
func (m *metricasEnMemoria) sprintsCerradosDeLaBase() (nombres []string, items [][]metricas.ItemDelSprint) {
	proyectos := make([]string, 0, len(m.escenario.proyectos))
	for _, id := range m.escenario.proyectos {
		proyectos = append(proyectos, id)
	}
	slices.Sort(proyectos)
	for _, proyectoID := range proyectos {
		for _, s := range m.escenario.base.Sprints(proyectoID) {
			if s.Estado != sprint.EstadoCerrado {
				continue
			}
			var delSprint []metricas.ItemDelSprint
			for _, fila := range m.escenario.base.FilasDelSprint(s.ID) {
				puntos, _ := fila.StoryPoints.Valor()
				delSprint = append(delSprint, metricas.ItemDelSprint{StoryPoints: puntos, Completado: fila.Resultado == sprint.ResultadoCompletado})
			}
			nombres = append(nombres, s.Nombre())
			items = append(items, delSprint)
		}
	}
	return nombres, items
}

func (m *metricasEnMemoria) AgregarItemAlSprint(sprint string, puntos int, completado bool) error {
	m.items[sprint] = append(m.items[sprint], metricas.ItemDelSprint{StoryPoints: puntos, Completado: completado})
	return nil
}

// resumen calcula el resumen del sprint. Un sprint sin items es una lista vacia,
// que el dominio resuelve con 0 y 0 (CA-025-3).
func (m *metricasEnMemoria) resumen(nombre string) (metricas.ResumenSprint, error) {
	items, armado := m.items[nombre]
	if !armado {
		// No lo armo el contexto: puede ser un sprint cerrado de verdad (US-011).
		nombres, deLaBase := m.sprintsCerradosDeLaBase()
		if i := slices.Index(nombres, nombre); i >= 0 {
			items = deLaBase[i]
		}
	}
	resumen, err := metricas.CalcularResumenSprint(items)
	if err != nil {
		return metricas.ResumenSprint{}, fmt.Errorf("resumen del sprint %q: %w", nombre, err)
	}
	return resumen, nil
}

func (m *metricasEnMemoria) PuntosPlanificados(sprint string) (int, error) {
	resumen, err := m.resumen(sprint)
	return resumen.Planificados, err
}

func (m *metricasEnMemoria) PuntosCompletados(sprint string) (int, error) {
	resumen, err := m.resumen(sprint)
	return resumen.Completados, err
}

// RegistrarSprintCerrado suma un sprint cerrado, en el orden en que llega.
func (m *metricasEnMemoria) RegistrarSprintCerrado(nombre string, planificados, completados int) error {
	m.cerrados = append(m.cerrados, metricas.SprintCerrado{
		Nombre:  nombre,
		Resumen: metricas.ResumenSprint{Planificados: planificados, Completados: completados},
	})
	return nil
}

// VelocidadDelEquipo calcula con los sprints cerrados que armo el contexto y con los
// que se cerraron de verdad en la base del escenario.
func (m *metricasEnMemoria) VelocidadDelEquipo() (float64, error) {
	cerrados := slices.Clone(m.cerrados)
	nombres, items := m.sprintsCerradosDeLaBase()
	for i, nombre := range nombres {
		resumen, err := metricas.CalcularResumenSprint(items[i])
		if err != nil {
			return 0, fmt.Errorf("resumen del sprint %q: %w", nombre, err)
		}
		cerrados = append(cerrados, metricas.SprintCerrado{Nombre: nombre, Resumen: resumen})
	}
	velocidad, err := metricas.CalcularVelocidad(cerrados)
	if err != nil {
		return 0, fmt.Errorf("velocidad del equipo: %w", err)
	}
	return velocidad, nil
}

// RegistrarHorasDelSprint suma al sprint un registro de esfuerzo, en el orden en que llega.
func (m *metricasEnMemoria) RegistrarHorasDelSprint(sprint string, horas float64) error {
	m.horas[sprint] = append(m.horas[sprint], horas)
	return nil
}

// HorasDelSprint calcula las horas con los story points planificados del sprint (los
// mismos de US-025) y con las horas que armo el contexto.
func (m *metricasEnMemoria) HorasDelSprint(sprint string) (steps.HorasDelSprint, error) {
	resumen, err := m.resumen(sprint)
	if err != nil {
		return steps.HorasDelSprint{}, err
	}
	horas, err := metricas.CalcularHoras(resumen.Planificados, m.horas[sprint])
	if err != nil {
		return steps.HorasDelSprint{}, fmt.Errorf("horas del sprint %q: %w", sprint, err)
	}
	porcentaje, calculable := horas.DesviacionPorcentual()
	return steps.HorasDelSprint{
		Estimadas:            horas.Estimadas,
		Reales:               horas.Reales,
		Desviacion:           horas.Desviacion,
		Porcentaje:           porcentaje,
		PorcentajeCalculable: calculable,
	}, nil
}
