// Runner de los escenarios BDD.
//
// Este archivo no sabe nada del dominio: solo arranca godog y le pasa los
// servicios que arma servicios_test.go. Los step definitions estan en
// features/steps y las frases disponibles en docs/diccionario-steps.md.
//
// Correr con:  make bdd
// Un escenario solo:  go test ./features/... -godog.tags=@US-026
package features_test

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
)

var opciones = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty",
	Paths:  []string{"."},

	// Estricto: un step pendiente o indefinido es una falla, no un aviso.
	// Es lo que impide que un escenario quede "en verde" sin haber probado nada.
	Strict: true,

	// Los escenarios no pueden depender del orden en que corren.
	Randomize: -1,
}

func init() {
	godog.BindCommandLineFlags("godog.", &opciones)
}

func TestEscenarios(t *testing.T) {
	o := opciones
	o.TestingT = t

	suite := godog.TestSuite{
		Name:                "seguimientoymedicion",
		ScenarioInitializer: func(sc *godog.ScenarioContext) { steps.Registrar(sc, nuevosServicios) },
		Options:             &o,
	}

	if suite.Run() != 0 {
		t.Fatal("hay escenarios BDD que no pasan")
	}
}
