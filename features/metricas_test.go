package features_test

import (
	"context"
	"fmt"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
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
	defectos  []metricas.DefectoRegistrado
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

// PorcentajeHistoriasCompletadas calcula el porcentaje de US-028 sobre los mismos items
// que arma el contexto para US-025. Un sprint sin items es una lista vacia, que el dominio
// resuelve con 0 sin error (RN-028-4 / CL-028-1); los Story Points negativos los rechaza el
// dominio con la posicion del item (RN-028-5), asi que aca no se valida nada.
func (m *metricasEnMemoria) PorcentajeHistoriasCompletadas(sprint string) (float64, error) {
	porcentaje, err := metricas.CalcularPorcentajeCompletadas(m.items[sprint])
	if err != nil {
		return 0, fmt.Errorf("porcentaje de historias del sprint %q: %w", sprint, err)
	}
	return porcentaje, nil
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

// horasDelResumen convierte el resumen de horas del dominio en lo que comprueban los steps.
func horasDelResumen(r metricas.ResumenHoras, porcentaje float64, calculable bool) steps.HorasDelSprint {
	return steps.HorasDelSprint{
		Estimadas:            r.Estimadas,
		Reales:               r.Reales,
		Desviacion:           r.Desviacion,
		Porcentaje:           porcentaje,
		PorcentajeCalculable: calculable,
	}
}

// CompararEsfuerzo usa el caso de uso real de US-021 sobre la base del escenario. El proyecto que el
// escenario no creo se manda con un ID que no existe, asi el rechazo lo produce la aplicacion.
func (m *metricasEnMemoria) CompararEsfuerzo(proyecto string) (steps.ComparacionDeEsfuerzo, error) {
	proyectoID, ok := m.escenario.proyectos[proyecto]
	if !ok {
		proyectoID = idDeProyectoInexistente
	}
	c, err := app.NuevoCompararEsfuerzo(m.escenario.base).Ejecutar(context.Background(), proyectoID)
	if err != nil {
		return steps.ComparacionDeEsfuerzo{}, err
	}

	totalPorcentaje, totalCalculable := c.Total.DesviacionPorcentual()
	resultado := steps.ComparacionDeEsfuerzo{
		Total:            horasDelResumen(c.Total, totalPorcentaje, totalCalculable),
		Comparables:      c.Comparables,
		HorasSinComparar: c.HorasSinComparar,
	}
	for _, h := range c.Historias {
		porcentaje, calculable := h.DesviacionPorcentual()
		resultado.Historias = append(resultado.Historias, steps.HistoriaComparada{
			Titulo: h.Titulo,
			Horas:  horasDelResumen(h.Horas, porcentaje, calculable),
		})
	}
	return resultado, nil
}

// RegistrarDefecto suma un defecto al contexto, en el orden en que llega. No valida nada: lo hace
// el dominio al calcular.
func (m *metricasEnMemoria) RegistrarDefecto(detectadoEn, resueltoEn int) error {
	m.defectos = append(m.defectos, metricas.DefectoRegistrado{SprintDeteccion: detectadoEn, SprintResolucion: resueltoEn})
	return nil
}

// DefectosDelSprint calcula los defectos detectados y resueltos del sprint con los defectos que
// armo el contexto (US-029).
func (m *metricasEnMemoria) DefectosDelSprint(sprint int) (steps.ResumenDeDefectos, error) {
	resumen, err := metricas.CalcularDefectosDelSprint(sprint, m.defectos)
	if err != nil {
		return steps.ResumenDeDefectos{}, fmt.Errorf("defectos del sprint %d: %w", sprint, err)
	}
	return steps.ResumenDeDefectos{Detectados: resumen.Detectados, Resueltos: resumen.Resueltos}, nil
}
