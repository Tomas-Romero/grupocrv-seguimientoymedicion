package steps

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cucumber/godog"
)

// Steps del area Backlog (US-005, US-006 y US-013).

func registrarBacklog(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^una historia "([^"]*)" con prioridad "([^"]*)"$`, m.dadoUnaHistoria)
	sc.Step(`^una historia "([^"]*)" con prioridad "([^"]*)" y (\d+) story points$`, m.dadoUnaHistoriaEstimada)

	sc.Step(`^se crea la historia "([^"]*)" con prioridad "([^"]*)"$`, m.cuandoSeCreaHistoria)
	sc.Step(`^se crea la historia "([^"]*)" con prioridad "([^"]*)", la descripción "([^"]*)" y los criterios:$`, m.cuandoSeCreaHistoriaCompleta)
	sc.Step(`^se crea una historia con un título de (\d+) caracteres y prioridad "([^"]*)"$`, m.cuandoSeCreaHistoriaConTituloDeLargo)
	sc.Step(`^se cambia la prioridad de la historia "([^"]*)" a "([^"]*)"$`, m.cuandoSeCambiaPrioridad)
	sc.Step(`^se estima la historia "([^"]*)" en (\d+) story points$`, m.cuandoSeEstima)

	sc.Step(`^el backlog del proyecto "([^"]*)" tiene (\d+) historias?$`, m.entoncesCantidadHistorias)
	sc.Step(`^la historia "([^"]*)" tiene prioridad "([^"]*)"$`, m.entoncesPrioridadDeHistoria)
	sc.Step(`^la historia "([^"]*)" tiene (\d+) story points$`, m.entoncesPuntosDeHistoria)
	sc.Step(`^la historia "([^"]*)" está sin estimar$`, m.entoncesHistoriaSinEstimar)
	sc.Step(`^la historia "([^"]*)" tiene el número (\d+)$`, m.entoncesNumeroDeHistoria)
	sc.Step(`^la historia "([^"]*)" tiene estado "([^"]*)"$`, m.entoncesEstadoDeHistoria)
	sc.Step(`^la historia "([^"]*)" tiene los criterios:$`, m.entoncesCriteriosDeHistoria)
	sc.Step(`^la historia "([^"]*)" está en el backlog$`, m.entoncesHistoriaEnBacklog)
}

func (m *mundo) dadoUnaHistoria(titulo, prioridad string) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	if err = backlog.CrearHistoria(m.proyectoActual, DatosHistoria{Titulo: titulo, Prioridad: prioridad}); err != nil {
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
	return m.registrarAccion(backlog.CrearHistoria(m.proyectoActual, DatosHistoria{Titulo: titulo, Prioridad: prioridad}))
}

func (m *mundo) cuandoSeCreaHistoriaCompleta(titulo, prioridad, descripcion string, tabla *godog.Table) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	criterios, err := columnaUnica(tabla)
	if err != nil {
		return err
	}
	historia := DatosHistoria{Titulo: titulo, Descripcion: descripcion, Prioridad: prioridad, Criterios: criterios}
	return m.registrarAccion(backlog.CrearHistoria(m.proyectoActual, historia))
}

// cuandoSeCreaHistoriaConTituloDeLargo arma el titulo con letras ñ: cada una
// ocupa dos bytes, asi el escenario tambien prueba que el limite cuenta
// caracteres y no bytes (CL-005-3).
func (m *mundo) cuandoSeCreaHistoriaConTituloDeLargo(largo int, prioridad string) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	historia := DatosHistoria{Titulo: strings.Repeat("ñ", largo), Prioridad: prioridad}
	return m.registrarAccion(backlog.CrearHistoria(m.proyectoActual, historia))
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
	puntos, estimada, err := backlog.PuntosDeHistoria(titulo)
	if err != nil {
		return fmt.Errorf("consultar los story points de %q: %w", titulo, err)
	}
	if !estimada {
		return fmt.Errorf("story points de %q: se esperaban %d, pero la historia está sin estimar", titulo, esperados)
	}
	if puntos != esperados {
		return fmt.Errorf("story points de %q: se esperaban %d, se obtuvieron %d", titulo, esperados, puntos)
	}
	return nil
}

func (m *mundo) entoncesHistoriaSinEstimar(titulo string) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	puntos, estimada, err := backlog.PuntosDeHistoria(titulo)
	if err != nil {
		return fmt.Errorf("consultar los story points de %q: %w", titulo, err)
	}
	if estimada {
		return fmt.Errorf("se esperaba que la historia %q estuviera sin estimar, pero tiene %d story points", titulo, puntos)
	}
	return nil
}

func (m *mundo) entoncesNumeroDeHistoria(titulo string, esperado int) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	numero, err := backlog.NumeroDeHistoria(titulo)
	if err != nil {
		return fmt.Errorf("consultar el número de %q: %w", titulo, err)
	}
	if numero != esperado {
		return fmt.Errorf("número de %q: se esperaba %d, se obtuvo %d", titulo, esperado, numero)
	}
	return nil
}

func (m *mundo) entoncesEstadoDeHistoria(titulo, esperado string) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	estado, err := backlog.EstadoDeHistoria(titulo)
	if err != nil {
		return fmt.Errorf("consultar el estado de %q: %w", titulo, err)
	}
	if estado != esperado {
		return fmt.Errorf("estado de %q: se esperaba %q, se obtuvo %q", titulo, esperado, estado)
	}
	return nil
}

func (m *mundo) entoncesCriteriosDeHistoria(titulo string, tabla *godog.Table) error {
	backlog, err := m.backlog()
	if err != nil {
		return err
	}
	esperados, err := columnaUnica(tabla)
	if err != nil {
		return err
	}
	criterios, err := backlog.CriteriosDeHistoria(titulo)
	if err != nil {
		return fmt.Errorf("consultar los criterios de %q: %w", titulo, err)
	}
	if !slices.Equal(criterios, esperados) {
		return fmt.Errorf("criterios de %q: se esperaban %q, se obtuvieron %q", titulo, esperados, criterios)
	}
	return nil
}

// columnaUnica devuelve los valores de una tabla de una sola columna, en orden.
// Gherkin recorta los espacios de cada celda, asi que una celda con solo
// espacios llega vacia.
func columnaUnica(tabla *godog.Table) ([]string, error) {
	valores := make([]string, 0, len(tabla.Rows))
	for i, fila := range tabla.Rows {
		if len(fila.Cells) != 1 {
			return nil, fmt.Errorf("la fila %d de la tabla tiene %d columnas, se esperaba una", i+1, len(fila.Cells))
		}
		valores = append(valores, fila.Cells[0].Value)
	}
	return valores, nil
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
