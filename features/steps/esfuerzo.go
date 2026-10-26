package steps

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Steps del area Esfuerzo (US-019).

func registrarEsfuerzo(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^el integrante "([^"]*)" ya registró (-?\d+(?:[.,]\d+)?) horas de "([^"]*)" en la historia "([^"]*)" el (\S+)$`, m.dadoEsfuerzoRegistrado)

	sc.Step(`^el integrante "([^"]*)" registra (-?\d+(?:[.,]\d+)?) horas de "([^"]*)" en la historia "([^"]*)" el (\S+)$`, m.cuandoRegistraEsfuerzo)

	sc.Step(`^el proyecto "([^"]*)" tiene (\d+) registros? de esfuerzo$`, m.entoncesCantidadDeRegistros)
	sc.Step(`^la historia "([^"]*)" tiene (\d+(?:[.,]\d+)?) horas registradas$`, m.entoncesHorasDeHistoria)

	// US-020: consultar el esfuerzo. Los tres filtros son opcionales y van en este orden.
	sc.Step(`^se consulta el esfuerzo del proyecto "([^"]*)"(?: de la historia "([^"]*)")?(?: del integrante "([^"]*)")?(?: del sprint "([^"]*)")?$`, m.cuandoSeConsultaElEsfuerzo)

	sc.Step(`^la consulta devuelve (\d+) registros?$`, m.entoncesLaConsultaDevuelveRegistros)
	sc.Step(`^la consulta devuelve primero el registro de "([^"]*)"$`, m.entoncesLaConsultaDevuelvePrimero)
	sc.Step(`^el total de horas de la consulta es (\d+(?:[.,]\d+)?)$`, m.entoncesElTotalDeLaConsultaEs)
	sc.Step(`^las horas de la historia "([^"]*)" en la consulta son (\d+(?:[.,]\d+)?)$`, m.entoncesHorasDeLaHistoriaEnLaConsulta)
	sc.Step(`^las horas del integrante "([^"]*)" en la consulta son (\d+(?:[.,]\d+)?)$`, m.entoncesHorasDelIntegranteEnLaConsulta)
	sc.Step(`^las horas del sprint "([^"]*)" en la consulta son (\d+(?:[.,]\d+)?)$`, m.entoncesHorasDelSprintEnLaConsulta)
	sc.Step(`^las horas fuera de todo sprint en la consulta son (\d+(?:[.,]\d+)?)$`, m.entoncesHorasFueraDeSprintEnLaConsulta)
}

// horasDe convierte las horas de un escenario. Acepta coma o punto decimal: la
// coma es lo natural al escribir en español.
func horasDe(texto string) (float64, error) {
	horas, err := strconv.ParseFloat(strings.Replace(texto, ",", ".", 1), 64)
	if err != nil {
		return 0, fmt.Errorf("las horas %q no son un número: %w", texto, err)
	}
	return horas, nil
}

func (m *mundo) dadoEsfuerzoRegistrado(integrante, horasTexto, actividad, historia, fechaTexto string) error {
	esfuerzo, err := m.esfuerzo()
	if err != nil {
		return err
	}
	horas, err := horasDe(horasTexto)
	if err != nil {
		return err
	}
	fecha, err := parsearFecha(fechaTexto)
	if err != nil {
		return err
	}
	if err = esfuerzo.RegistrarEsfuerzo(m.proyectoActual, integrante, historia, fecha, actividad, horas); err != nil {
		return fmt.Errorf("registrar el esfuerzo de %q del contexto: %w", integrante, err)
	}
	return nil
}

func (m *mundo) cuandoRegistraEsfuerzo(integrante, horasTexto, actividad, historia, fechaTexto string) error {
	esfuerzo, err := m.esfuerzo()
	if err != nil {
		return err
	}
	horas, err := horasDe(horasTexto)
	if err != nil {
		return err
	}
	fecha, err := parsearFecha(fechaTexto)
	if err != nil {
		return err
	}
	return m.registrarAccion(esfuerzo.RegistrarEsfuerzo(m.proyectoActual, integrante, historia, fecha, actividad, horas))
}

func (m *mundo) entoncesCantidadDeRegistros(proyecto string, esperada int) error {
	esfuerzo, err := m.esfuerzo()
	if err != nil {
		return err
	}
	cantidad, err := esfuerzo.CantidadRegistros(proyecto)
	if err != nil {
		return fmt.Errorf("contar los registros de esfuerzo de %q: %w", proyecto, err)
	}
	if cantidad != esperada {
		return fmt.Errorf("registros de esfuerzo de %q: se esperaban %d, hay %d", proyecto, esperada, cantidad)
	}
	return nil
}

func (m *mundo) entoncesHorasDeHistoria(historia, esperadasTexto string) error {
	esfuerzo, err := m.esfuerzo()
	if err != nil {
		return err
	}
	esperadas, err := horasDe(esperadasTexto)
	if err != nil {
		return err
	}
	horas, err := esfuerzo.HorasDeHistoria(m.proyectoActual, historia)
	if err != nil {
		return fmt.Errorf("consultar las horas de %q: %w", historia, err)
	}
	const tolerancia = 0.001
	if math.Abs(horas-esperadas) > tolerancia {
		return fmt.Errorf("horas registradas de %q: se esperaban %g, hay %g", historia, esperadas, horas)
	}
	return nil
}

// cuandoSeConsultaElEsfuerzo guarda el resultado de la consulta, o el error si la rechazan.
func (m *mundo) cuandoSeConsultaElEsfuerzo(proyecto, historia, integrante, sprint string) error {
	esfuerzo, err := m.esfuerzo()
	if err != nil {
		return err
	}
	consulta, err := esfuerzo.ConsultarEsfuerzo(proyecto, historia, integrante, sprint)
	m.consultaEsfuerzo = consulta
	return m.registrarAccion(err)
}

func (m *mundo) entoncesLaConsultaDevuelveRegistros(esperados int) error {
	if obtenidos := len(m.consultaEsfuerzo.Actividades); obtenidos != esperados {
		return fmt.Errorf("registros de la consulta: se esperaban %d, hay %d (%v)", esperados, obtenidos, m.consultaEsfuerzo.Actividades)
	}
	return nil
}

func (m *mundo) entoncesLaConsultaDevuelvePrimero(actividad string) error {
	if len(m.consultaEsfuerzo.Actividades) == 0 {
		return fmt.Errorf("la consulta no devolvio registros: se esperaba %q primero", actividad)
	}
	if primera := m.consultaEsfuerzo.Actividades[0]; primera != actividad {
		return fmt.Errorf("primer registro de la consulta: se esperaba %q, es %q", actividad, primera)
	}
	return nil
}

// comparaHoras compara las horas de la consulta con las esperadas por el escenario.
func comparaHoras(que string, obtenidas float64, esperadasTexto string) error {
	esperadas, err := horasDe(esperadasTexto)
	if err != nil {
		return err
	}
	const tolerancia = 0.001
	if math.Abs(obtenidas-esperadas) > tolerancia {
		return fmt.Errorf("%s: se esperaban %g horas, hay %g", que, esperadas, obtenidas)
	}
	return nil
}

func (m *mundo) entoncesElTotalDeLaConsultaEs(esperadas string) error {
	return comparaHoras("total de la consulta", m.consultaEsfuerzo.TotalHoras, esperadas)
}

func (m *mundo) entoncesHorasDeLaHistoriaEnLaConsulta(historia, esperadas string) error {
	return comparaHoras(fmt.Sprintf("historia %q en la consulta", historia), m.consultaEsfuerzo.PorHistoria[historia], esperadas)
}

func (m *mundo) entoncesHorasDelIntegranteEnLaConsulta(integrante, esperadas string) error {
	return comparaHoras(fmt.Sprintf("integrante %q en la consulta", integrante), m.consultaEsfuerzo.PorIntegrante[integrante], esperadas)
}

func (m *mundo) entoncesHorasDelSprintEnLaConsulta(sprint, esperadas string) error {
	return comparaHoras(fmt.Sprintf("sprint %q en la consulta", sprint), m.consultaEsfuerzo.PorSprint[sprint], esperadas)
}

func (m *mundo) entoncesHorasFueraDeSprintEnLaConsulta(esperadas string) error {
	return comparaHoras("fuera de todo sprint en la consulta", m.consultaEsfuerzo.FueraDeSprint, esperadas)
}
