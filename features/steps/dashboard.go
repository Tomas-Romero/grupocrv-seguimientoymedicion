package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

// Steps del dashboard (US-030), en el area Metricas. Despues de consultar, las frases de Story Points, porcentaje y horas de
// US-025, US-027 y US-028 se refieren a la tarjeta del sprint de referencia.

// DashboardDelProyecto es lo que los Entonces de US-030 comprueban.
type DashboardDelProyecto struct {
	// Referencia es el nombre ("Sprint 1") del sprint de la tarjeta, o vacio si el proyecto no tiene sprints.
	Referencia          string
	PuntosPlanificados  int
	PuntosCompletados   int
	PorcentajeHistorias float64
	Horas               HorasDelSprint
	Velocidad           float64
	// Cerrados van del sprint mas reciente al mas antiguo.
	Cerrados []SprintCerradoDelDashboard
	// Defectos son todos los del proyecto, tambien los descartados.
	Defectos int
	// Grupos son las barras del grafico de Story Points, del sprint mas antiguo al mas reciente.
	Grupos []GrupoDeBarras
}

// SprintCerradoDelDashboard es una fila de la tabla de sprints cerrados.
type SprintCerradoDelDashboard struct {
	Nombre     string
	Detectados int
	Resueltos  int
}

// GrupoDeBarras son las dos barras de un sprint cerrado; las alturas van de 0 a 100.
type GrupoDeBarras struct {
	Sprint                                string
	AlturaPlanificados, AlturaCompletados float64
}

func registrarDashboard(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^se consulta el dashboard del proyecto "([^"]*)"$`, m.cuandoSeConsultaElDashboard)
	sc.Step(`^la tarjeta del sprint es el "([^"]*)"$`, m.entoncesLaTarjetaDelSprintEs)
	sc.Step(`^la velocidad del dashboard es (\d+(?:[.,]\d+)?)$`, m.entoncesLaVelocidadDelDashboardEs)
	sc.Step(`^el dashboard no tiene sprint de referencia$`, m.entoncesElDashboardNoTieneSprintDeReferencia)
	sc.Step(`^el gráfico de Story Points tiene (\d+) grupos? de barras$`, m.entoncesElGraficoTieneGrupos)
	sc.Step(`^la barra de "(planificados|completados)" del "([^"]*)" mide (\d+(?:[.,]\d+)?)$`, m.entoncesLaBarraMide)
	sc.Step(`^el dashboard lista (\d+) sprints? cerrados?$`, m.entoncesElDashboardListaSprintsCerrados)
	sc.Step(`^el sprint cerrado "([^"]*)" figura con (\d+) defectos? detectados? y (\d+) resueltos?$`, m.entoncesElSprintCerradoFiguraConDefectos)
	sc.Step(`^el dashboard cuenta (\d+) defectos?$`, m.entoncesElDashboardCuentaDefectos)
}

// cuandoSeConsultaElDashboard guarda el resultado aunque haya error: un escenario de error (CA-030-5) sigue hasta el
// `Entonces la operacion se rechaza`. Deja ademas los numeros de la tarjeta del sprint donde los leen las frases de US-025,
// US-027 y US-028.
func (m *mundo) cuandoSeConsultaElDashboard(proyecto string) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	dashboard, err := metricas.ConsultarDashboard(proyecto)
	m.dashboard = dashboard
	m.puntosPlanificados = dashboard.PuntosPlanificados
	m.puntosCompletados = dashboard.PuntosCompletados
	m.porcentajeHistorias = dashboard.PorcentajeHistorias
	m.horas = dashboard.Horas
	return m.registrarAccion(err)
}

func (m *mundo) entoncesLaTarjetaDelSprintEs(esperado string) error {
	if m.dashboard.Referencia != esperado {
		return fmt.Errorf("la tarjeta del sprint es %q y se esperaba %q", m.dashboard.Referencia, esperado)
	}
	return nil
}

func (m *mundo) entoncesLaVelocidadDelDashboardEs(esperada string) error {
	return compararCifra("velocidad del dashboard", esperada, m.dashboard.Velocidad)
}

func (m *mundo) entoncesElDashboardNoTieneSprintDeReferencia() error {
	if m.dashboard.Referencia != "" {
		return fmt.Errorf("el dashboard tiene sprint de referencia (%q) y se esperaba que no", m.dashboard.Referencia)
	}
	return nil
}

func (m *mundo) entoncesElGraficoTieneGrupos(esperados int) error {
	if obtenidos := len(m.dashboard.Grupos); obtenidos != esperados {
		return fmt.Errorf("el gráfico de Story Points tiene %d grupos de barras y se esperaban %d", obtenidos, esperados)
	}
	return nil
}

func (m *mundo) entoncesLaBarraMide(serie, sprint, esperada string) error {
	for _, g := range m.dashboard.Grupos {
		if g.Sprint != sprint {
			continue
		}
		altura := g.AlturaPlanificados
		if serie == "completados" {
			altura = g.AlturaCompletados
		}
		return compararCifra("barra de "+serie+" del "+sprint, esperada, altura)
	}
	return fmt.Errorf("el gráfico no tiene un grupo de barras para el %q", sprint)
}

func (m *mundo) entoncesElDashboardListaSprintsCerrados(esperados int) error {
	if obtenidos := len(m.dashboard.Cerrados); obtenidos != esperados {
		return fmt.Errorf("el dashboard lista %d sprints cerrados y se esperaban %d", obtenidos, esperados)
	}
	return nil
}

func (m *mundo) entoncesElSprintCerradoFiguraConDefectos(sprint string, detectados, resueltos int) error {
	for _, c := range m.dashboard.Cerrados {
		if c.Nombre != sprint {
			continue
		}
		if c.Detectados != detectados || c.Resueltos != resueltos {
			return fmt.Errorf("el %s figura con %d detectados y %d resueltos y se esperaban %d y %d", sprint, c.Detectados, c.Resueltos, detectados, resueltos)
		}
		return nil
	}
	return fmt.Errorf("el dashboard no lista el %q entre los sprints cerrados", sprint)
}

func (m *mundo) entoncesElDashboardCuentaDefectos(esperados int) error {
	if m.dashboard.Defectos != esperados {
		return fmt.Errorf("el dashboard cuenta %d defectos y se esperaban %d", m.dashboard.Defectos, esperados)
	}
	return nil
}
