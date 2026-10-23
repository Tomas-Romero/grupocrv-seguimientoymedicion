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
	sc.Step(`^el integrante "([^"]*)" ya votó (-?\d+) en la sesión de la historia "([^"]*)"$`, m.dadoUnVoto)

	sc.Step(`^se inicia una sesión de Planning Poker sobre la historia "([^"]*)"$`, m.cuandoSeIniciaSesion)
	sc.Step(`^el integrante "([^"]*)" vota (-?\d+) en la sesión de la historia "([^"]*)"$`, m.cuandoVota)

	sc.Step(`^la sesión de la historia "([^"]*)" está "([^"]*)" en la ronda (\d+)$`, m.entoncesEstadoDeLaSesion)
	sc.Step(`^los participantes de la sesión de la historia "([^"]*)" son:$`, m.entoncesParticipantes)
	sc.Step(`^en la sesión de la historia "([^"]*)" votaron:$`, m.entoncesVotaron)
	sc.Step(`^en la sesión de la historia "([^"]*)" faltan votar:$`, m.entoncesFaltanVotar)
	sc.Step(`^los votos de la sesión de la historia "([^"]*)" están ocultos$`, m.entoncesVotosOcultos)
}

func (m *mundo) dadoUnVoto(integrante string, valor int, titulo string) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	if err := poker.Votar(integrante, titulo, valor); err != nil {
		return fmt.Errorf("votar de %q en la sesion de %q del contexto: %w", integrante, titulo, err)
	}
	return nil
}

func (m *mundo) cuandoVota(integrante string, valor int, titulo string) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	return m.registrarAccion(poker.Votar(integrante, titulo, valor))
}

func (m *mundo) entoncesVotaron(titulo string, tabla *godog.Table) error {
	return m.compararIntegrantes(titulo, tabla, "votaron", PlanningPoker.QuienesVotaron)
}

func (m *mundo) entoncesFaltanVotar(titulo string, tabla *godog.Table) error {
	return m.compararIntegrantes(titulo, tabla, "faltan votar", PlanningPoker.QuienesFaltan)
}

// compararIntegrantes compara los nombres que devuelve la consulta con los de la tabla,
// en orden.
func (m *mundo) compararIntegrantes(titulo string, tabla *godog.Table, que string, consulta func(PlanningPoker, string) ([]string, error)) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	esperados, err := columnaUnica(tabla)
	if err != nil {
		return err
	}
	nombres, err := consulta(poker, titulo)
	if err != nil {
		return fmt.Errorf("consultar quienes %s en la sesion de %q: %w", que, titulo, err)
	}
	if !slices.Equal(nombres, esperados) {
		return fmt.Errorf("en la sesion de %q %s: se esperaban %q, son %q", titulo, que, esperados, nombres)
	}
	return nil
}

func (m *mundo) entoncesVotosOcultos(titulo string) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	ocultos, err := poker.VotosOcultos(titulo)
	if err != nil {
		return fmt.Errorf("consultar los votos de la sesion de %q: %w", titulo, err)
	}
	if !ocultos {
		return fmt.Errorf("se esperaba que los votos de la sesion de %q estuvieran ocultos, pero se pueden leer", titulo)
	}
	return nil
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
