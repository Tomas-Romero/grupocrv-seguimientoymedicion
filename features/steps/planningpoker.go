package steps

import (
	"fmt"
	"slices"

	"github.com/cucumber/godog"
)

// Steps del area PlanningPoker (US-014 a US-018). La sesion se nombra por su
// historia: una historia tiene como maximo una sesion abierta (RN-014-7).

func registrarPlanningPoker(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^una sesión de Planning Poker sobre la historia "([^"]*)"$`, m.dadoUnaSesion)

	sc.Step(`^se inicia una sesión de Planning Poker sobre la historia "([^"]*)"$`, m.cuandoSeIniciaSesion)

	sc.Step(`^la sesión de la historia "([^"]*)" está "([^"]*)" en la ronda (\d+)$`, m.entoncesEstadoDeLaSesion)
	sc.Step(`^los participantes de la sesión de la historia "([^"]*)" son:$`, m.entoncesParticipantes)
}

func (m *mundo) dadoUnaSesion(titulo string) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	if err := poker.IniciarSesion(m.proyectoActual, titulo); err != nil {
		return fmt.Errorf("iniciar la sesion de %q del contexto: %w", titulo, err)
	}
	return nil
}

func (m *mundo) cuandoSeIniciaSesion(titulo string) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	return m.registrarAccion(poker.IniciarSesion(m.proyectoActual, titulo))
}

func (m *mundo) entoncesEstadoDeLaSesion(titulo, estadoEsperado string, rondaEsperada int) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	estado, ronda, err := poker.EstadoDeLaSesion(titulo)
	if err != nil {
		return fmt.Errorf("consultar la sesion de %q: %w", titulo, err)
	}
	if estado != estadoEsperado || ronda != rondaEsperada {
		return fmt.Errorf("sesion de %q: se esperaba %q en la ronda %d, esta %q en la ronda %d", titulo, estadoEsperado, rondaEsperada, estado, ronda)
	}
	return nil
}

func (m *mundo) entoncesParticipantes(titulo string, tabla *godog.Table) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	esperados, err := columnaUnica(tabla)
	if err != nil {
		return err
	}
	participantes, err := poker.ParticipantesDeLaSesion(titulo)
	if err != nil {
		return fmt.Errorf("consultar los participantes de la sesion de %q: %w", titulo, err)
	}
	if !slices.Equal(participantes, esperados) {
		return fmt.Errorf("participantes de la sesion de %q: se esperaban %q, son %q", titulo, esperados, participantes)
	}
	return nil
}
