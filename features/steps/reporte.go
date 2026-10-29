package steps

import (
	"fmt"
	"slices"

	"github.com/cucumber/godog"
)

// Steps de los reportes en pantalla (US-033), en el area Metricas. Despues de consultar el reporte de un sprint, las frases de Story Points,
// porcentaje y horas de US-025, US-027 y US-028 se refieren a las metricas de ese sprint.

// SeccionDelReporte es una seccion del reporte tal como la comprueban los escenarios: su titulo, cuantos elementos tiene, si esta vacia y si
// informa un error.
type SeccionDelReporte struct {
	Titulo    string
	Elementos int
	Vacia     bool
	ConError  bool
}

// Reporte es lo que los Entonces de US-033 comprueban. Las secciones salen en el orden del reporte.
type Reporte struct {
	Secciones []SeccionDelReporte
	// Solo en el reporte de un sprint: el resultado de cada historia en ese sprint, por titulo, y los defectos que lista.
	Historias  map[string]bool
	Detectados int
	Resueltos  int
	// Metricas del sprint (solo en el reporte de un sprint).
	PuntosPlanificados  int
	PuntosCompletados   int
	PorcentajeHistorias float64
	Horas               HorasDelSprint
}

func registrarReporte(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^se consulta el reporte del proyecto "([^"]*)"$`, m.cuandoSeConsultaElReporteDelProyecto)
	sc.Step(`^se consulta el reporte del sprint "([^"]*)" del proyecto "([^"]*)"$`, m.cuandoSeConsultaElReporteDelSprint)
	sc.Step(`^el reporte tiene las secciones:$`, m.entoncesElReporteTieneLasSecciones)
	sc.Step(`^la sección "([^"]*)" del reporte está vacía$`, m.entoncesLaSeccionEstaVacia)
	sc.Step(`^la sección "([^"]*)" del reporte tiene (\d+) elementos?$`, m.entoncesLaSeccionTieneElementos)
	sc.Step(`^ninguna sección del reporte informa un error$`, m.entoncesNingunaSeccionInformaError)
	sc.Step(`^el reporte del sprint muestra la historia "([^"]*)" como "(Completada|No completada)"$`, m.entoncesElReporteMuestraLaHistoria)
	sc.Step(`^el reporte del sprint lista (\d+) defectos? detectados? y (\d+) resueltos?$`, m.entoncesElReporteListaDefectos)
}

// cuandoSeConsultaElReporteDelProyecto guarda el resultado aunque haya error: un escenario de error (CA-033-4) sigue hasta el
// `Entonces la operacion se rechaza`.
func (m *mundo) cuandoSeConsultaElReporteDelProyecto(proyecto string) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	reporte, err := metricas.ConsultarReporteDelProyecto(proyecto)
	m.reporte = reporte
	return m.registrarAccion(err)
}

func (m *mundo) cuandoSeConsultaElReporteDelSprint(sprint, proyecto string) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	reporte, err := metricas.ConsultarReporteDelSprint(proyecto, sprint)
	m.reporte = reporte
	m.puntosPlanificados = reporte.PuntosPlanificados
	m.puntosCompletados = reporte.PuntosCompletados
	m.porcentajeHistorias = reporte.PorcentajeHistorias
	m.horas = reporte.Horas
	return m.registrarAccion(err)
}

// entoncesElReporteTieneLasSecciones compara los titulos de la tabla, en orden, con los del reporte (RN-033-2).
func (m *mundo) entoncesElReporteTieneLasSecciones(tabla *godog.Table) error {
	esperadas := make([]string, 0, len(tabla.Rows))
	for _, fila := range tabla.Rows {
		esperadas = append(esperadas, fila.Cells[0].Value)
	}
	obtenidas := make([]string, len(m.reporte.Secciones))
	for i, s := range m.reporte.Secciones {
		obtenidas[i] = s.Titulo
	}
	if !slices.Equal(esperadas, obtenidas) {
		return fmt.Errorf("las secciones del reporte son %q y se esperaban %q, en ese orden", obtenidas, esperadas)
	}
	return nil
}

func (m *mundo) seccion(titulo string) (SeccionDelReporte, error) {
	for _, s := range m.reporte.Secciones {
		if s.Titulo == titulo {
			return s, nil
		}
	}
	return SeccionDelReporte{}, fmt.Errorf("el reporte no tiene la sección %q", titulo)
}

func (m *mundo) entoncesLaSeccionEstaVacia(titulo string) error {
	s, err := m.seccion(titulo)
	if err != nil {
		return err
	}
	if !s.Vacia {
		return fmt.Errorf("la sección %q tiene %d elementos y se esperaba que estuviera vacía", titulo, s.Elementos)
	}
	return nil
}

func (m *mundo) entoncesLaSeccionTieneElementos(titulo string, esperados int) error {
	s, err := m.seccion(titulo)
	if err != nil {
		return err
	}
	if s.Elementos != esperados {
		return fmt.Errorf("la sección %q tiene %d elementos y se esperaban %d", titulo, s.Elementos, esperados)
	}
	return nil
}

func (m *mundo) entoncesNingunaSeccionInformaError() error {
	for _, s := range m.reporte.Secciones {
		if s.ConError {
			return fmt.Errorf("la sección %q informa un error", s.Titulo)
		}
	}
	return nil
}

func (m *mundo) entoncesElReporteMuestraLaHistoria(historia, resultado string) error {
	completada, ok := m.reporte.Historias[historia]
	if !ok {
		return fmt.Errorf("el reporte del sprint no tiene la historia %q", historia)
	}
	if esperada := resultado == "Completada"; completada != esperada {
		return fmt.Errorf("el reporte muestra la historia %q como completada=%v y se esperaba %v", historia, completada, esperada)
	}
	return nil
}

func (m *mundo) entoncesElReporteListaDefectos(detectados, resueltos int) error {
	if m.reporte.Detectados != detectados || m.reporte.Resueltos != resueltos {
		return fmt.Errorf("el reporte lista %d detectados y %d resueltos y se esperaban %d y %d", m.reporte.Detectados, m.reporte.Resueltos, detectados, resueltos)
	}
	return nil
}
