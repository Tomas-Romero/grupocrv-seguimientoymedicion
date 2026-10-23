package steps

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Steps del area Metricas (US-025, US-026 y US-027).

func registrarMetricas(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^el sprint "([^"]*)" con una historia de (-?\d+) story points completada$`, m.dadoHistoriaDeSprintCompletada)
	sc.Step(`^el sprint "([^"]*)" con una historia de (-?\d+) story points sin completar$`, m.dadoHistoriaDeSprintSinCompletar)
	sc.Step(`^el sprint cerrado "([^"]*)" con (\d+) story points completados de (\d+) planificados$`, m.dadoSprintCerradoConPuntos)

	sc.Step(`^se consultan los story points planificados del sprint "([^"]*)"$`, m.cuandoSeConsultanPlanificados)
	sc.Step(`^se consultan los story points completados del sprint "([^"]*)"$`, m.cuandoSeConsultanCompletados)
	sc.Step(`^se consulta la velocidad del equipo$`, m.cuandoSeConsultaVelocidad)

	sc.Step(`^los story points planificados son (\d+)$`, m.entoncesPlanificadosSon)
	sc.Step(`^los story points completados son (\d+)$`, m.entoncesCompletadosSon)
	sc.Step(`^la velocidad del equipo es (\d+(?:[.,]\d+)?)$`, m.entoncesVelocidadEs)

	sc.Step(`^el sprint "([^"]*)" con (-?\d+(?:[.,]\d+)?) horas registradas$`, m.dadoHorasRegistradas)
	sc.Step(`^se consultan las horas del sprint "([^"]*)"$`, m.cuandoSeConsultanLasHoras)
	sc.Step(`^las horas estimadas son (-?\d+(?:[.,]\d+)?)$`, m.entoncesHorasEstimadasSon)
	sc.Step(`^las horas reales son (-?\d+(?:[.,]\d+)?)$`, m.entoncesHorasRealesSon)
	sc.Step(`^la desviación de horas es (-?\d+(?:[.,]\d+)?)$`, m.entoncesDesviacionDeHorasEs)
	sc.Step(`^la desviación porcentual es (-?\d+(?:[.,]\d+)?)$`, m.entoncesDesviacionPorcentualEs)
	sc.Step(`^la desviación porcentual no se puede calcular$`, m.entoncesDesviacionPorcentualNoSePuedeCalcular)
}

// dadoHistoriaDeSprintCompletada y dadoHistoriaDeSprintSinCompletar arman el
// sprint de un escenario de metricas. Aceptan numeros negativos a proposito: es
// la unica forma de escribir el escenario de error de US-025 (CA-025-4).
func (m *mundo) dadoHistoriaDeSprintCompletada(sprint string, puntos int) error {
	return m.agregarHistoriaDeSprint(sprint, puntos, true)
}

func (m *mundo) dadoHistoriaDeSprintSinCompletar(sprint string, puntos int) error {
	return m.agregarHistoriaDeSprint(sprint, puntos, false)
}

func (m *mundo) agregarHistoriaDeSprint(sprint string, puntos int, completada bool) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	if err = metricas.AgregarItemAlSprint(sprint, puntos, completada); err != nil {
		return fmt.Errorf("agregar una historia al sprint %q del contexto: %w", sprint, err)
	}
	return nil
}

// dadoSprintCerradoConPuntos arma un sprint cerrado para el escenario de la
// velocidad. No valida nada: "8 completados de 5 planificados" es justo el dato
// corrupto que tiene que rechazar el dominio (CA-026-4).
func (m *mundo) dadoSprintCerradoConPuntos(nombre string, completados, planificados int) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	if err = metricas.RegistrarSprintCerrado(nombre, planificados, completados); err != nil {
		return fmt.Errorf("registrar el sprint cerrado %q del contexto: %w", nombre, err)
	}
	return nil
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

// dadoHorasRegistradas suma un registro de esfuerzo al sprint del contexto. Acepta
// 0 y numeros negativos a proposito: es la unica forma de escribir los escenarios de
// error de US-027 (CA-027-5).
func (m *mundo) dadoHorasRegistradas(sprint, horas string) error {
	valor, err := numeroDelEscenario(horas)
	if err != nil {
		return err
	}
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	if err = metricas.RegistrarHorasDelSprint(sprint, valor); err != nil {
		return fmt.Errorf("registrar horas en el sprint %q del contexto: %w", sprint, err)
	}
	return nil
}

func (m *mundo) cuandoSeConsultanLasHoras(sprint string) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	horas, err := metricas.HorasDelSprint(sprint)
	m.horas = horas
	return m.registrarAccion(err)
}

func (m *mundo) entoncesHorasEstimadasSon(esperadas string) error {
	return compararCifra("horas estimadas", esperadas, m.horas.Estimadas)
}

func (m *mundo) entoncesHorasRealesSon(esperadas string) error {
	return compararCifra("horas reales", esperadas, m.horas.Reales)
}

func (m *mundo) entoncesDesviacionDeHorasEs(esperada string) error {
	return compararCifra("desviación de horas", esperada, m.horas.Desviacion)
}

func (m *mundo) entoncesDesviacionPorcentualEs(esperada string) error {
	if !m.horas.PorcentajeCalculable {
		return fmt.Errorf("la desviación porcentual no se puede calcular y se esperaba %s", esperada)
	}
	return compararCifra("desviación porcentual", esperada, m.horas.Porcentaje)
}

func (m *mundo) entoncesDesviacionPorcentualNoSePuedeCalcular() error {
	if m.horas.PorcentajeCalculable {
		return fmt.Errorf("la desviación porcentual se pudo calcular (%g) y se esperaba que no", m.horas.Porcentaje)
	}
	return nil
}

// numeroDelEscenario lee un numero como lo escribe una persona: acepta coma o punto
// decimal, y uno negativo.
func numeroDelEscenario(texto string) (float64, error) {
	valor, err := strconv.ParseFloat(strings.Replace(texto, ",", ".", 1), 64)
	if err != nil {
		return 0, fmt.Errorf("%q no es un número: %w", texto, err)
	}
	return valor, nil
}

// compararCifra compara lo que dice el escenario con lo que calculo el dominio, con
// la tolerancia que ya usa la velocidad.
func compararCifra(nombre, esperada string, obtenida float64) error {
	quiero, err := numeroDelEscenario(esperada)
	if err != nil {
		return err
	}
	const tolerancia = 0.001
	if math.Abs(obtenida-quiero) > tolerancia {
		return fmt.Errorf("%s: se esperaba %g, se obtuvo %g", nombre, quiero, obtenida)
	}
	return nil
}
