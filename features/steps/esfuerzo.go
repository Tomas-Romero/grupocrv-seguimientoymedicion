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
