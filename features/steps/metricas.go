package steps

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Steps del area Metricas (US-025 y US-026).

func registrarMetricas(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^se consultan los story points planificados del sprint "([^"]*)"$`, m.cuandoSeConsultanPlanificados)
	sc.Step(`^se consultan los story points completados del sprint "([^"]*)"$`, m.cuandoSeConsultanCompletados)
	sc.Step(`^se consulta la velocidad del equipo$`, m.cuandoSeConsultaVelocidad)

	sc.Step(`^los story points planificados son (\d+)$`, m.entoncesPlanificadosSon)
	sc.Step(`^los story points completados son (\d+)$`, m.entoncesCompletadosSon)
	sc.Step(`^la velocidad del equipo es (\d+(?:[.,]\d+)?)$`, m.entoncesVelocidadEs)
}

func (m *mundo) cuandoSeConsultanPlanificados(sprint string) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	puntos, err := metricas.PuntosPlanificados(sprint)
	m.puntosPlanificados = puntos
	return m.registrarAccion(err)
}

func (m *mundo) cuandoSeConsultanCompletados(sprint string) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	puntos, err := metricas.PuntosCompletados(sprint)
	m.puntosCompletados = puntos
	return m.registrarAccion(err)
}

func (m *mundo) cuandoSeConsultaVelocidad() error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	velocidad, err := metricas.VelocidadDelEquipo()
	m.velocidad = velocidad
	return m.registrarAccion(err)
}

func (m *mundo) entoncesPlanificadosSon(esperados int) error {
	if m.puntosPlanificados != esperados {
		return fmt.Errorf("story points planificados: se esperaban %d, se obtuvieron %d", esperados, m.puntosPlanificados)
	}
	return nil
}

func (m *mundo) entoncesCompletadosSon(esperados int) error {
	if m.puntosCompletados != esperados {
		return fmt.Errorf("story points completados: se esperaban %d, se obtuvieron %d", esperados, m.puntosCompletados)
	}
	return nil
}

// entoncesVelocidadEs acepta tanto "25" como "25.5" y "25,5": el escenario lo
// escribe una persona y la coma decimal es lo natural en espanol.
func (m *mundo) entoncesVelocidadEs(valor string) error {
	esperada, err := strconv.ParseFloat(strings.Replace(valor, ",", ".", 1), 64)
	if err != nil {
		return fmt.Errorf("la velocidad esperada %q no es un número: %w", valor, err)
	}
	const tolerancia = 0.001
	if math.Abs(m.velocidad-esperada) > tolerancia {
		return fmt.Errorf("velocidad del equipo: se esperaba %g, se obtuvo %g", esperada, m.velocidad)
	}
	return nil
}
