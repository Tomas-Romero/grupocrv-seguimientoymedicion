package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

// Steps del area Sprints (US-008, US-009, US-010 y US-011).

func registrarSprints(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^un sprint "([^"]*)" con el objetivo "([^"]*)"$`, m.dadoUnSprint)
	sc.Step(`^la historia "([^"]*)" asignada al sprint "([^"]*)"$`, m.dadoHistoriaAsignada)
	sc.Step(`^la historia "([^"]*)" completada$`, m.dadoHistoriaCompletada)
	sc.Step(`^el sprint "([^"]*)" cerrado$`, m.dadoSprintCerrado)
	sc.Step(`^ningún sprint cerrado$`, m.dadoNingunSprintCerrado)

	sc.Step(`^se crea el sprint "([^"]*)" con el objetivo "([^"]*)"$`, m.cuandoSeCreaSprint)
	sc.Step(`^se asigna la historia "([^"]*)" al sprint "([^"]*)"$`, m.cuandoSeAsigna)
	sc.Step(`^se marca la historia "([^"]*)" como completada$`, m.cuandoSeCompleta)
	sc.Step(`^se cierra el sprint "([^"]*)"$`, m.cuandoSeCierraSprint)

	sc.Step(`^la historia "([^"]*)" está en el sprint "([^"]*)"$`, m.entoncesHistoriaEnSprint)
	sc.Step(`^el sprint "([^"]*)" está cerrado$`, m.entoncesSprintCerrado)
}

func (m *mundo) dadoUnSprint(nombre, objetivo string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	if err = sprints.CrearSprint(m.proyectoActual, nombre, objetivo); err != nil {
		return fmt.Errorf("crear el sprint %q del contexto: %w", nombre, err)
	}
	return nil
}

func (m *mundo) dadoHistoriaAsignada(titulo, sprint string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	if err = sprints.AsignarHistoriaASprint(titulo, sprint); err != nil {
		return fmt.Errorf("asignar la historia %q al sprint %q en el contexto: %w", titulo, sprint, err)
	}
	return nil
}

func (m *mundo) dadoHistoriaCompletada(titulo string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	if err = sprints.CompletarHistoria(titulo); err != nil {
		return fmt.Errorf("completar la historia %q del contexto: %w", titulo, err)
	}
	return nil
}

func (m *mundo) dadoSprintCerrado(nombre string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	if err = sprints.CerrarSprint(nombre); err != nil {
		return fmt.Errorf("cerrar el sprint %q del contexto: %w", nombre, err)
	}
	return nil
}

// dadoNingunSprintCerrado no hace nada: los servicios arrancan vacios antes de
// cada escenario. Existe para que el escenario del caso limite diga
// explicitamente cual es la precondicion, en vez de dejarla implicita.
func (m *mundo) dadoNingunSprintCerrado() error {
	_, err := m.sprints()
	return err
}

func (m *mundo) cuandoSeCreaSprint(nombre, objetivo string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	return m.registrarAccion(sprints.CrearSprint(m.proyectoActual, nombre, objetivo))
}

func (m *mundo) cuandoSeAsigna(titulo, sprint string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	return m.registrarAccion(sprints.AsignarHistoriaASprint(titulo, sprint))
}

func (m *mundo) cuandoSeCompleta(titulo string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	return m.registrarAccion(sprints.CompletarHistoria(titulo))
}

func (m *mundo) cuandoSeCierraSprint(nombre string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	return m.registrarAccion(sprints.CerrarSprint(nombre))
}

func (m *mundo) entoncesHistoriaEnSprint(titulo, esperado string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	sprint, err := sprints.SprintDeHistoria(titulo)
	if err != nil {
		return fmt.Errorf("consultar el sprint de %q: %w", titulo, err)
	}
	if sprint != esperado {
		return fmt.Errorf("sprint de %q: se esperaba %q, se obtuvo %q", titulo, esperado, sprint)
	}
	return nil
}

func (m *mundo) entoncesSprintCerrado(nombre string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	cerrado, err := sprints.SprintEstaCerrado(nombre)
	if err != nil {
		return fmt.Errorf("consultar el estado del sprint %q: %w", nombre, err)
	}
	if !cerrado {
		return fmt.Errorf("se esperaba que el sprint %q estuviera cerrado, pero sigue abierto", nombre)
	}
	return nil
}
