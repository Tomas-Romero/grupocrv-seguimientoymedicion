package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// Registrar conecta todas las frases del diccionario con su implementacion.
//
// Cada area registra sus propias frases en su archivo (proyectos.go,
// backlog.go, sprints.go, metricas.go). Si agregas un step, agregalo tambien a
// docs/diccionario-steps.md en el mismo commit — un step sin documentar es un
// step que nadie va a usar y que nadie va a borrar.
func Registrar(sc *godog.ScenarioContext, nuevo Fabrica) {
	m := &mundo{}

	// Cada escenario arranca con servicios limpios.
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		servicios, err := nuevo()
		if err != nil {
			return ctx, fmt.Errorf("crear los servicios del escenario: %w", err)
		}
		*m = mundo{servicios: servicios}
		return ctx, nil
	})

	registrarProyectos(sc, m)
	registrarBacklog(sc, m)
	registrarSprints(sc, m)
	registrarMetricas(sc, m)
	registrarResultados(sc, m)
}

// registrarResultados registra los steps de resultado que no dependen de
// ninguna area: miran el error que dejo el ultimo `Cuando`.
func registrarResultados(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^la operación es exitosa$`, m.entoncesEsExitosa)
	sc.Step(`^la operación se rechaza$`, m.entoncesSeRechaza)
	sc.Step(`^el mensaje de error indica "([^"]*)"$`, m.entoncesElMensajeIndica)
}

func (m *mundo) entoncesEsExitosa() error {
	if m.ultimoError != nil {
		return fmt.Errorf("se esperaba que la operación fuera exitosa, pero devolvió: %w", m.ultimoError)
	}
	return nil
}

func (m *mundo) entoncesSeRechaza() error {
	if m.ultimoError == nil {
		return fmt.Errorf("se esperaba que la operación se rechazara, pero fue exitosa")
	}
	return nil
}

func (m *mundo) entoncesElMensajeIndica(fragmento string) error {
	if m.ultimoError == nil {
		return fmt.Errorf("se esperaba un error que contuviera %q, pero la operación fue exitosa", fragmento)
	}
	if !strings.Contains(strings.ToLower(m.ultimoError.Error()), strings.ToLower(fragmento)) {
		return fmt.Errorf("el mensaje de error no contiene %q; mensaje recibido: %q", fragmento, m.ultimoError.Error())
	}
	return nil
}
