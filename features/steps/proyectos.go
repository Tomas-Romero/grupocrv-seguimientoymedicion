package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

// Steps del area Proyectos (US-001, US-002 y US-003).
//
// Los steps de contexto (Dado) SI devuelven el error: si el armado del escenario
// falla, el escenario no tiene sentido y tiene que cortarse ahi. Los de accion
// (Cuando) lo guardan en el mundo para `la operacion se rechaza`.

func registrarProyectos(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^un proyecto "([^"]*)"$`, m.dadoUnProyecto)
	sc.Step(`^un proyecto "([^"]*)" con fechas del "([^"]*)" al "([^"]*)"$`, m.dadoUnProyectoConFechas)
	sc.Step(`^el integrante "([^"]*)" con rol "([^"]*)"$`, m.dadoUnIntegrante)
	sc.Step(`^ningún proyecto$`, m.dadoNingunProyecto)

	sc.Step(`^se crea un proyecto "([^"]*)" con fechas del "([^"]*)" al "([^"]*)"$`, m.cuandoSeCreaProyecto)
	sc.Step(`^se modifica el nombre del proyecto "([^"]*)" a "([^"]*)"$`, m.cuandoSeModificaNombre)
	sc.Step(`^se registra al integrante "([^"]*)" con rol "([^"]*)"$`, m.cuandoSeRegistraIntegrante)

	sc.Step(`^el proyecto "([^"]*)" existe con fechas del "([^"]*)" al "([^"]*)"$`, m.entoncesElProyectoExiste)
	sc.Step(`^el proyecto "([^"]*)" tiene (\d+) integrantes?$`, m.entoncesCantidadIntegrantes)
}

func (m *mundo) dadoUnProyecto(nombre string) error {
	proyectos, err := m.proyectos()
	if err != nil {
		return err
	}
	if err = proyectos.CrearProyecto(nombre, inicioPorDefecto, finPorDefecto); err != nil {
		return fmt.Errorf("crear el proyecto %q del contexto: %w", nombre, err)
	}
	m.proyectoActual = nombre
	return nil
}

func (m *mundo) dadoUnProyectoConFechas(nombre, inicio, fin string) error {
	proyectos, err := m.proyectos()
	if err != nil {
		return err
	}
	desde, err := parsearFecha(inicio)
	if err != nil {
		return err
	}
	hasta, err := parsearFecha(fin)
	if err != nil {
		return err
	}
	if err = proyectos.CrearProyecto(nombre, desde, hasta); err != nil {
		return fmt.Errorf("crear el proyecto %q del contexto: %w", nombre, err)
	}
	m.proyectoActual = nombre
	return nil
}

// proyectoInexistente es el proyecto actual que deja `Dado ningún proyecto`.
// Ningun step crea un proyecto con este nombre, asi que las areas no le
// encuentran ID y llaman a la aplicacion con uno que no existe: el rechazo lo
// produce el caso de uso, no el step (CA-005-8, CA-008-10).
const proyectoInexistente = "(ningún proyecto)"

// dadoNingunProyecto deja al escenario sin un proyecto existente: los steps que
// siguen operan sobre uno que no existe. No necesita ningun area conectada.
func (m *mundo) dadoNingunProyecto() error {
	m.proyectoActual = proyectoInexistente
	return nil
}

func (m *mundo) dadoUnIntegrante(nombre, rol string) error {
	proyectos, err := m.proyectos()
	if err != nil {
		return err
	}
	if err = proyectos.RegistrarIntegrante(m.proyectoActual, nombre, rol); err != nil {
		return fmt.Errorf("registrar al integrante %q del contexto: %w", nombre, err)
	}
	return nil
}

func (m *mundo) cuandoSeCreaProyecto(nombre, inicio, fin string) error {
	proyectos, err := m.proyectos()
	if err != nil {
		return err
	}
	desde, err := parsearFecha(inicio)
	if err != nil {
		return err
	}
	hasta, err := parsearFecha(fin)
	if err != nil {
		return err
	}
	return m.registrarAccion(proyectos.CrearProyecto(nombre, desde, hasta))
}

func (m *mundo) cuandoSeModificaNombre(actual, nuevo string) error {
	proyectos, err := m.proyectos()
	if err != nil {
		return err
	}
	return m.registrarAccion(proyectos.ModificarNombreProyecto(actual, nuevo))
}

func (m *mundo) cuandoSeRegistraIntegrante(nombre, rol string) error {
	proyectos, err := m.proyectos()
	if err != nil {
		return err
	}
	return m.registrarAccion(proyectos.RegistrarIntegrante(m.proyectoActual, nombre, rol))
}

func (m *mundo) entoncesElProyectoExiste(nombre, inicio, fin string) error {
	proyectos, err := m.proyectos()
	if err != nil {
		return err
	}
	desdeEsperado, err := parsearFecha(inicio)
	if err != nil {
		return err
	}
	hastaEsperado, err := parsearFecha(fin)
	if err != nil {
		return err
	}
	desde, hasta, err := proyectos.ProyectoExiste(nombre)
	if err != nil {
		return fmt.Errorf("el proyecto %q no existe: %w", nombre, err)
	}
	if !desde.Equal(desdeEsperado) {
		return fmt.Errorf("fecha de inicio del proyecto %q: se esperaba %s, se obtuvo %s",
			nombre, desdeEsperado.Format("2006-01-02"), desde.Format("2006-01-02"))
	}
	if !hasta.Equal(hastaEsperado) {
		return fmt.Errorf("fecha de fin del proyecto %q: se esperaba %s, se obtuvo %s",
			nombre, hastaEsperado.Format("2006-01-02"), hasta.Format("2006-01-02"))
	}
	return nil
}

func (m *mundo) entoncesCantidadIntegrantes(proyecto string, esperada int) error {
	proyectos, err := m.proyectos()
	if err != nil {
		return err
	}
	cantidad, err := proyectos.CantidadIntegrantes(proyecto)
	if err != nil {
		return fmt.Errorf("contar los integrantes de %q: %w", proyecto, err)
	}
	if cantidad != esperada {
		return fmt.Errorf("integrantes de %q: se esperaban %d, hay %d", proyecto, esperada, cantidad)
	}
	return nil
}
