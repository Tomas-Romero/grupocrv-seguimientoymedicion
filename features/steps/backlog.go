package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

// Steps del area Backlog (US-005, US-006 y US-013).

func registrarBacklog(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^una historia "([^"]*)" con prioridad "([^"]*)"$`, m.dadoUnaHistoria)
	sc.Step(`^una historia "([^"]*)" con prioridad "([^"]*)" y (\d+) story points$`, m.dadoUnaHistoriaEstimada)

	sc.Step(`^se crea la historia "([^"]*)" con prioridad "([^"]*)"$`, m.cuandoSeCreaHistoria)
	sc.Step(`^se cambia la prioridad de la historia "([^"]*)" a "([^"]*)"$`, m.cuandoSeCambiaPrioridad)
	sc.Step(`^se estima la historia "([^"]*)" en (\d+) story points$`, m.cuandoSeEstima)

	sc.Step(`^el backlog del proyecto "([^"]*)" tiene (\d+) historias?$`, m.entoncesCantidadHistorias)
	sc.Step(`^la historia "([^"]*)" tiene prioridad "([^"]*)"$`, m.entoncesPrioridadDeHistoria)
	sc.Step(`^la historia "([^"]*)" tiene (\d+) story points$`, m.entoncesPuntosDeHistoria)
	sc.Step(`^la historia "([^"]*)" está en el backlog$`, m.entoncesHistoriaEnBacklog)
}

func (m *mundo) dadoUnaHistoria(titulo, prioridad string) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	if err = backlog.CrearHistoria(m.proyectoActual, titulo, prioridad); err != nil {
		return fmt.Errorf("crear la historia %q del contexto: %w", titulo, err)
	}
	return nil
}

func (m *mundo) dadoUnaHistoriaEstimada(titulo, prioridad string, puntos int) error {
	if err := m.dadoUnaHistoria(titulo, prioridad); err != nil {
		return err
	}
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	if err = backlog.EstimarHistoria(titulo, puntos); err != nil {
		return fmt.Errorf("estimar la historia %q del contexto: %w", titulo, err)
	}
	return nil
}

func (m *mundo) cuandoSeCreaHistoria(titulo, prioridad string) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	return m.registrarAccion(backlog.CrearHistoria(m.proyectoActual, titulo, prioridad))
}

func (m *mundo) cuandoSeCambiaPrioridad(titulo, prioridad string) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	return m.registrarAccion(backlog.CambiarPrioridadHistoria(titulo, prioridad))
}

func (m *mundo) cuandoSeEstima(titulo string, puntos int) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	return m.registrarAccion(backlog.EstimarHistoria(titulo, puntos))
}

func (m *mundo) entoncesCantidadHistorias(proyecto string, esperada int) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	cantidad, err := backlog.CantidadHistoriasEnBacklog(proyecto)
	if err != nil {
		return fmt.Errorf("contar las historias del backlog de %q: %w", proyecto, err)
	}
	if cantidad != esperada {
		return fmt.Errorf("historias en el backlog de %q: se esperaban %d, hay %d", proyecto, esperada, cantidad)
	}
	return nil
}

func (m *mundo) entoncesPrioridadDeHistoria(titulo, esperada string) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	prioridad, err := backlog.PrioridadDeHistoria(titulo)
	if err != nil {
		return fmt.Errorf("consultar la prioridad de %q: %w", titulo, err)
	}
	if prioridad != esperada {
		return fmt.Errorf("prioridad de %q: se esperaba %q, se obtuvo %q", titulo, esperada, prioridad)
	}
	return nil
}

func (m *mundo) entoncesPuntosDeHistoria(titulo string, esperados int) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	puntos, err := backlog.PuntosDeHistoria(titulo)
	if err != nil {
		return fmt.Errorf("consultar los story points de %q: %w", titulo, err)
	}
	if puntos != esperados {
		return fmt.Errorf("story points de %q: se esperaban %d, se obtuvieron %d", titulo, esperados, puntos)
	}
	return nil
}

func (m *mundo) entoncesHistoriaEnBacklog(titulo string) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	enBacklog, err := backlog.HistoriaEstaEnBacklog(titulo)
	if err != nil {
		return fmt.Errorf("consultar si %q está en el backlog: %w", titulo, err)
	}
	if !enBacklog {
		return fmt.Errorf("se esperaba que la historia %q estuviera en el backlog, pero sigue asignada a un sprint", titulo)
	}
	return nil
}
