package steps

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Steps del area Defectos (US-022).

func registrarDefectos(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^un defecto de severidad "([^"]*)" con la descripción "([^"]*)" en la historia "([^"]*)" detectado en el sprint "([^"]*)"$`, m.dadoDefectoRegistrado)

	sc.Step(`^se registra un defecto de severidad "([^"]*)" con la descripción "([^"]*)" en la historia "([^"]*)" detectado en el sprint "([^"]*)"$`, m.cuandoSeRegistraDefecto)
	sc.Step(`^se registra un defecto de severidad "([^"]*)" con una descripción de (\d+) caracteres en la historia "([^"]*)" detectado en el sprint "([^"]*)"$`, m.cuandoSeRegistraDefectoConDescripcionLarga)

	sc.Step(`^el proyecto "([^"]*)" tiene (\d+) defectos?$`, m.entoncesCantidadDeDefectos)
	sc.Step(`^el defecto de la historia "([^"]*)" tiene severidad "([^"]*)"$`, m.entoncesSeveridadDelDefecto)
	sc.Step(`^el defecto de la historia "([^"]*)" tiene estado "([^"]*)"$`, m.entoncesEstadoDelDefecto)
	sc.Step(`^el defecto de la historia "([^"]*)" fue detectado en el sprint "([^"]*)"$`, m.entoncesSprintDeDeteccionDelDefecto)
	sc.Step(`^el defecto de la historia "([^"]*)" está sin resolver$`, m.entoncesDefectoSinResolver)
}

// descripcionDeLargo arma una descripcion de n letras ñ. Cada una ocupa dos bytes, asi que el
// escenario tambien prueba que el limite cuenta caracteres y no bytes (CL-022-4).
func descripcionDeLargo(n string) (string, error) {
	largo, err := strconv.Atoi(n)
	if err != nil {
		return "", fmt.Errorf("el largo %q no es un número: %w", n, err)
	}
	return strings.Repeat("ñ", largo), nil
}

func (m *mundo) dadoDefectoRegistrado(severidad, descripcion, historia, sprint string) error {
	defectos, err := m.areaDefectos()
	if err != nil {
		return err
	}
	if err = defectos.RegistrarDefecto(m.proyectoActual, historia, sprint, severidad, descripcion); err != nil {
		return fmt.Errorf("registrar el defecto de %q del contexto: %w", historia, err)
	}
	return nil
}

func (m *mundo) cuandoSeRegistraDefecto(severidad, descripcion, historia, sprint string) error {
	defectos, err := m.areaDefectos()
	if err != nil {
		return err
	}
	return m.registrarAccion(defectos.RegistrarDefecto(m.proyectoActual, historia, sprint, severidad, descripcion))
}

func (m *mundo) cuandoSeRegistraDefectoConDescripcionLarga(severidad, largo, historia, sprint string) error {
	defectos, err := m.areaDefectos()
	if err != nil {
		return err
	}
	descripcion, err := descripcionDeLargo(largo)
	if err != nil {
		return err
	}
	return m.registrarAccion(defectos.RegistrarDefecto(m.proyectoActual, historia, sprint, severidad, descripcion))
}

func (m *mundo) entoncesCantidadDeDefectos(proyecto string, esperada int) error {
	defectos, err := m.areaDefectos()
	if err != nil {
		return err
	}
	cantidad, err := defectos.CantidadDefectos(proyecto)
	if err != nil {
		return fmt.Errorf("contar los defectos de %q: %w", proyecto, err)
	}
	if cantidad != esperada {
		return fmt.Errorf("defectos de %q: se esperaban %d, hay %d", proyecto, esperada, cantidad)
	}
	return nil
}

func (m *mundo) entoncesSeveridadDelDefecto(historia, esperada string) error {
	defectos, err := m.areaDefectos()
	if err != nil {
		return err
	}
	severidad, err := defectos.SeveridadDeDefecto(historia)
	if err != nil {
		return fmt.Errorf("consultar la severidad del defecto de %q: %w", historia, err)
	}
	if severidad != esperada {
		return fmt.Errorf("severidad del defecto de %q: se esperaba %q, es %q", historia, esperada, severidad)
	}
	return nil
}

func (m *mundo) entoncesEstadoDelDefecto(historia, esperado string) error {
	defectos, err := m.areaDefectos()
	if err != nil {
		return err
	}
	estado, err := defectos.EstadoDeDefecto(historia)
	if err != nil {
		return fmt.Errorf("consultar el estado del defecto de %q: %w", historia, err)
	}
	if estado != esperado {
		return fmt.Errorf("estado del defecto de %q: se esperaba %q, es %q", historia, esperado, estado)
	}
	return nil
}

func (m *mundo) entoncesSprintDeDeteccionDelDefecto(historia, esperado string) error {
	defectos, err := m.areaDefectos()
	if err != nil {
		return err
	}
	sprint, err := defectos.SprintDeDeteccionDeDefecto(historia)
	if err != nil {
		return fmt.Errorf("consultar el sprint de detección del defecto de %q: %w", historia, err)
	}
	if sprint != esperado {
		return fmt.Errorf("sprint de detección del defecto de %q: se esperaba %q, es %q", historia, esperado, sprint)
	}
	return nil
}

func (m *mundo) entoncesDefectoSinResolver(historia string) error {
	defectos, err := m.areaDefectos()
	if err != nil {
		return err
	}
	sinResolver, err := defectos.DefectoSinResolver(historia)
	if err != nil {
		return fmt.Errorf("consultar el sprint de resolución del defecto de %q: %w", historia, err)
	}
	if !sinResolver {
		return fmt.Errorf("el defecto de %q tiene sprint de resolución: se esperaba que estuviera sin resolver", historia)
	}
	return nil
}
