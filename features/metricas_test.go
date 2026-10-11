package features_test

import (
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
)

// metricasEnMemoria es el area Metricas de los escenarios: guarda los items de
// cada sprint que arma el contexto y los pasa al dominio al consultar. No hay
// caso de uso intermedio porque US-025 y US-026 son calculos puros del dominio;
// cuando existan los sprints y el backlog reales (US-008 y US-005), la capa de
// aplicacion va a adaptar sus items a metricas.ItemDelSprint.
type metricasEnMemoria struct {
	items    map[string][]metricas.ItemDelSprint
	cerrados []metricas.SprintCerrado
}

var _ steps.Metricas = (*metricasEnMemoria)(nil)

func nuevasMetricasEnMemoria() *metricasEnMemoria {
	return &metricasEnMemoria{items: map[string][]metricas.ItemDelSprint{}}
}

func (m *metricasEnMemoria) AgregarItemAlSprint(sprint string, puntos int, completado bool) error {
	m.items[sprint] = append(m.items[sprint], metricas.ItemDelSprint{StoryPoints: puntos, Completado: completado})
	return nil
}

// resumen calcula el resumen del sprint. Un sprint sin items es una lista vacia,
// que el dominio resuelve con 0 y 0 (CA-025-3).
func (m *metricasEnMemoria) resumen(sprint string) (metricas.ResumenSprint, error) {
	resumen, err := metricas.CalcularResumenSprint(m.items[sprint])
	if err != nil {
		return metricas.ResumenSprint{}, fmt.Errorf("resumen del sprint %q: %w", sprint, err)
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

func (m *metricasEnMemoria) VelocidadDelEquipo() (float64, error) {
	velocidad, err := metricas.CalcularVelocidad(m.cerrados)
	if err != nil {
		return 0, fmt.Errorf("velocidad del equipo: %w", err)
	}
	return velocidad, nil
}
