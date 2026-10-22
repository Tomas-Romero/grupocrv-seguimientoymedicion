package steps

import (
	"fmt"
	"time"

	"github.com/cucumber/godog"
)

// Steps del area Sprints (US-008, US-009, US-010 y US-011).

func registrarSprints(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^un sprint "([^"]*)" con el objetivo "([^"]*)"$`, m.dadoUnSprint)
	sc.Step(`^un sprint "([^"]*)" con el objetivo "([^"]*)" del "([^"]*)" al "([^"]*)"$`, m.dadoUnSprintConFechas)
	sc.Step(`^la historia "([^"]*)" asignada al sprint "([^"]*)"$`, m.dadoHistoriaAsignada)
	sc.Step(`^la historia "([^"]*)" completada$`, m.dadoHistoriaCompletada)
	sc.Step(`^el sprint "([^"]*)" cerrado$`, m.dadoSprintCerrado)
	sc.Step(`^ningún sprint cerrado$`, m.dadoNingunSprintCerrado)

	sc.Step(`^se crea el sprint "([^"]*)" con el objetivo "([^"]*)"$`, m.cuandoSeCreaSprint)
	sc.Step(`^se crea el sprint "([^"]*)" con el objetivo "([^"]*)" del "([^"]*)" al "([^"]*)"$`, m.cuandoSeCreaSprintConFechas)
	sc.Step(`^se asigna la historia "([^"]*)" al sprint "([^"]*)"$`, m.cuandoSeAsigna)
	sc.Step(`^se marca la historia "([^"]*)" como completada$`, m.cuandoSeCompleta)
	sc.Step(`^se desmarca la historia "([^"]*)"$`, m.cuandoSeDesmarca)
	sc.Step(`^se cierra el sprint "([^"]*)"$`, m.cuandoSeCierraSprint)

	sc.Step(`^la historia "([^"]*)" está en el sprint "([^"]*)"$`, m.entoncesHistoriaEnSprint)
	sc.Step(`^el sprint "([^"]*)" está cerrado$`, m.entoncesSprintCerrado)
	sc.Step(`^el sprint "([^"]*)" está abierto$`, m.entoncesSprintAbierto)
}

// fechasPorDefectoDelSprint arma las fechas de un sprint cuando la frase no las
// nombra: el primero empieza el dia de inicio del proyecto y cada siguiente, el dia
// despues de que termina el ultimo sprint del escenario; todos duran 14 dias. Asi
// son siempre validas y consecutivas, y el escenario no tiene que ocuparse de ellas.
func (m *mundo) fechasPorDefectoDelSprint() (inicio, fin time.Time) {
	if ultimoFin, ok := m.finDelUltimoSprint[m.proyectoActual]; ok {
		inicio = ultimoFin.AddDate(0, 0, 1)
	} else {
		inicio = inicioPorDefecto
		if proyectos, err := m.proyectos(); err == nil {
			if desde, _, err := proyectos.ProyectoExiste(m.proyectoActual); err == nil {
				inicio = desde
			}
		}
	}
	return inicio, inicio.AddDate(0, 0, 13)
}

// crearSprint crea el sprint y, si sale bien, anota su fin como el del ultimo sprint
// del proyecto. Una creacion rechazada no cuenta como ultimo sprint. El nombre entre
// comillas de la frase es el nombre ESPERADO: si el sprint se crea con otro, el step
// falla (no es un rechazo de la operacion).
func (m *mundo) crearSprint(esperado, objetivo string, inicio, fin time.Time) (rechazo error, fallo error) {
	sprints, err := m.sprints()
	if err != nil {
		return nil, err
	}
	nombre, err := sprints.CrearSprint(m.proyectoActual, objetivo, inicio, fin)
	if err != nil {
		return err, nil
	}
	if m.finDelUltimoSprint == nil {
		m.finDelUltimoSprint = map[string]time.Time{}
	}
	m.finDelUltimoSprint[m.proyectoActual] = fin
	if nombre != esperado {
		return nil, fmt.Errorf("el sprint se creó con el nombre %q, pero el escenario esperaba %q", nombre, esperado)
	}
	return nil, nil
}

func (m *mundo) dadoUnSprint(nombre, objetivo string) error {
	inicio, fin := m.fechasPorDefectoDelSprint()
	return m.dadoSprintDelContexto(nombre, objetivo, inicio, fin)
}

func (m *mundo) dadoUnSprintConFechas(nombre, objetivo, desde, hasta string) error {
	inicio, fin, err := parsearFechas(desde, hasta)
	if err != nil {
		return err
	}
	return m.dadoSprintDelContexto(nombre, objetivo, inicio, fin)
}

func (m *mundo) dadoSprintDelContexto(nombre, objetivo string, inicio, fin time.Time) error {
	rechazo, fallo := m.crearSprint(nombre, objetivo, inicio, fin)
	if fallo != nil {
		return fallo
	}
	if rechazo != nil {
		return fmt.Errorf("crear el sprint %q del contexto: %w", nombre, rechazo)
	}
	return nil
}

// parsearFechas convierte las dos fechas de una frase.
func parsearFechas(desde, hasta string) (inicio, fin time.Time, err error) {
	if inicio, err = parsearFecha(desde); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if fin, err = parsearFecha(hasta); err != nil {
		return time.Time{}, time.Time{}, err
	}
	return inicio, fin, nil
}

func (m *mundo) dadoHistoriaAsignada(titulo, sprint string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	if err = sprints.AsignarHistoriaASprint(m.proyectoActual, titulo, sprint); err != nil {
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
	if err = sprints.CerrarSprint(m.proyectoActual, nombre); err != nil {
		return fmt.Errorf("cerrar el sprint %q del contexto: %w", nombre, err)
	}
	return nil
}

// dadoNingunSprintCerrado no hace nada: los servicios arrancan vacios antes de
// cada escenario. Existe para que el escenario del caso limite diga
// explicitamente cual es la precondicion, en vez de dejarla implicita. No exige
// ninguna area conectada: lo usan tambien los escenarios de metricas, que no
// pasan por el area Sprints (el paso que sigue es el que exige la suya).
func (m *mundo) dadoNingunSprintCerrado() error {
	return nil
}

func (m *mundo) cuandoSeCreaSprint(nombre, objetivo string) error {
	inicio, fin := m.fechasPorDefectoDelSprint()
	return m.cuandoSeCreaSprintEn(nombre, objetivo, inicio, fin)
}

func (m *mundo) cuandoSeCreaSprintConFechas(nombre, objetivo, desde, hasta string) error {
	inicio, fin, err := parsearFechas(desde, hasta)
	if err != nil {
		return err
	}
	return m.cuandoSeCreaSprintEn(nombre, objetivo, inicio, fin)
}

func (m *mundo) cuandoSeCreaSprintEn(nombre, objetivo string, inicio, fin time.Time) error {
	rechazo, fallo := m.crearSprint(nombre, objetivo, inicio, fin)
	if fallo != nil {
		return fallo
	}
	return m.registrarAccion(rechazo)
}

func (m *mundo) cuandoSeAsigna(titulo, sprint string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	return m.registrarAccion(sprints.AsignarHistoriaASprint(m.proyectoActual, titulo, sprint))
}

func (m *mundo) cuandoSeCompleta(titulo string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	return m.registrarAccion(sprints.CompletarHistoria(titulo))
}

func (m *mundo) cuandoSeDesmarca(titulo string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	return m.registrarAccion(sprints.DesmarcarHistoria(titulo))
}

func (m *mundo) cuandoSeCierraSprint(nombre string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	return m.registrarAccion(sprints.CerrarSprint(m.proyectoActual, nombre))
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
	cerrado, err := sprints.SprintEstaCerrado(m.proyectoActual, nombre)
	if err != nil {
		return fmt.Errorf("consultar el estado del sprint %q: %w", nombre, err)
	}
	if !cerrado {
		return fmt.Errorf("se esperaba que el sprint %q estuviera cerrado, pero sigue abierto", nombre)
	}
	return nil
}

func (m *mundo) entoncesSprintAbierto(nombre string) error {
	sprints, err := m.sprints()
	if err != nil {
		return err
	}
	cerrado, err := sprints.SprintEstaCerrado(m.proyectoActual, nombre)
	if err != nil {
		return fmt.Errorf("consultar el estado del sprint %q: %w", nombre, err)
	}
	if cerrado {
		return fmt.Errorf("se esperaba que el sprint %q estuviera abierto, pero está cerrado", nombre)
	}
	return nil
}
