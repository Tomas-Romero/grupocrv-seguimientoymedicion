package steps

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Steps del area Proyectos (US-001, US-002, US-003 y US-004).
//
// Los steps de contexto (Dado) SI devuelven el error: si el armado del escenario
// falla, el escenario no tiene sentido y tiene que cortarse ahi. Los de accion
// (Cuando) lo guardan en el mundo para `la operacion se rechaza`.

func registrarProyectos(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^un proyecto "([^"]*)"$`, m.dadoUnProyecto)
	sc.Step(`^un proyecto "([^"]*)" con fechas del "([^"]*)" al "([^"]*)"$`, m.dadoUnProyectoConFechas)
	sc.Step(`^el integrante "([^"]*)" con rol "([^"]*)"$`, m.dadoUnIntegrante)
	// PROPUESTA US-003: se aprueba en el PR junto con docs/diccionario-steps.md.
	sc.Step(`^el integrante "([^"]*)" con email "([^"]*)" y rol "([^"]*)"$`, m.dadoUnIntegranteConEmail)
	sc.Step(`^ningún proyecto$`, m.dadoNingunProyecto)

	sc.Step(`^se crea un proyecto "([^"]*)" con fechas del "([^"]*)" al "([^"]*)"$`, m.cuandoSeCreaProyecto)
	sc.Step(`^se modifica el nombre del proyecto "([^"]*)" a "([^"]*)"$`, m.cuandoSeModificaNombre)
	sc.Step(`^se modifican las fechas del proyecto "([^"]*)" del "([^"]*)" al "([^"]*)"$`, m.cuandoSeModificanFechas)
	sc.Step(`^se registra al integrante "([^"]*)" con rol "([^"]*)"$`, m.cuandoSeRegistraIntegrante)
	// PROPUESTA US-003: se aprueba en el PR junto con docs/diccionario-steps.md.
	sc.Step(`^se registra al integrante "([^"]*)" con email "([^"]*)" y rol "([^"]*)"$`, m.cuandoSeRegistraIntegranteConEmail)

	sc.Step(`^el proyecto "([^"]*)" existe con fechas del "([^"]*)" al "([^"]*)"$`, m.entoncesElProyectoExiste)
	sc.Step(`^el proyecto "([^"]*)" tiene (\d+) integrantes?$`, m.entoncesCantidadIntegrantes)

	// PROPUESTA US-004: se aprueban en el PR junto con docs/diccionario-steps.md.
	sc.Step(`^se consulta el estado del proyecto "([^"]*)"$`, m.cuandoSeConsultaElEstadoDelProyecto)
	sc.Step(`^el estado del proyecto es "([^"]*)"$`, m.entoncesElEstadoDelProyectoEs)
	sc.Step(`^los días del estado son (\d+)$`, m.entoncesLosDiasDelEstadoSon)
	sc.Step(`^el proyecto tiene (\d+) historias? pendientes?, (\d+) en progreso y (\d+) completadas?$`, m.entoncesLasHistoriasDelProyectoSon)
	sc.Step(`^los story points del proyecto son (\d+) y los completados (\d+)$`, m.entoncesLosStoryPointsDelProyectoSon)
	sc.Step(`^el porcentaje de historias completadas del proyecto es (\d+(?:[.,]\d+)?)$`, m.entoncesElPorcentajeDelProyectoEs)
	sc.Step(`^el sprint abierto del proyecto es "([^"]*)"$`, m.entoncesElSprintAbiertoDelProyectoEs)
	sc.Step(`^el proyecto no tiene sprint abierto$`, m.entoncesElProyectoNoTieneSprintAbierto)
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

// dominioDeEmails es el dominio reservado para pruebas (.test, RFC 2606): los
// emails que arma el step a partir del nombre nunca apuntan a una casilla real.
const dominioDeEmails = "ejemplo.test"

// emailDeNombre arma el email de un integrante cuando la frase no lo nombra
// (`el integrante "X" con rol "Y"`). Cumple dos cosas con CUALQUIER nombre,
// incluso con espacios, tildes o vacio:
//
//   - es una direccion simple valida para net/mail;
//   - es distinta para cada nombre distinto, porque el email es unico por
//     proyecto (RN-003-7) y dos integrantes de un escenario no pueden chocar
//     por accidente.
//
// Conserva solo las letras minusculas ASCII y los digitos, y codifica todo lo
// demas (mayusculas, espacios, tildes, "_") como "_" + codigo Unicode en
// hexadecimal + "_". El "_" final cierra el codigo y nunca aparece suelto, asi
// que el resultado se puede decodificar y no hay dos nombres con el mismo email.
// Ya sale en minuscula, por lo que la normalizacion del dominio no lo cambia.
// Un nombre vacio da "_": el email sigue siendo valido y el unico error del
// escenario es el del nombre.
func emailDeNombre(nombre string) string {
	var local strings.Builder
	for _, r := range nombre {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			local.WriteRune(r)
			continue
		}
		local.WriteString("_" + strconv.FormatInt(int64(r), 16) + "_")
	}
	if local.Len() == 0 {
		local.WriteString("_")
	}
	return local.String() + "@" + dominioDeEmails
}

func (m *mundo) dadoUnIntegrante(nombre, rol string) error {
	return m.dadoUnIntegranteConEmail(nombre, emailDeNombre(nombre), rol)
}

func (m *mundo) dadoUnIntegranteConEmail(nombre, email, rol string) error {
	proyectos, err := m.proyectos()
	if err != nil {
		return err
	}
	if err = proyectos.RegistrarIntegrante(m.proyectoActual, nombre, email, rol); err != nil {
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

func (m *mundo) cuandoSeModificanFechas(nombre, inicio, fin string) error {
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
	return m.registrarAccion(proyectos.ModificarFechasProyecto(nombre, desde, hasta))
}

func (m *mundo) cuandoSeRegistraIntegrante(nombre, rol string) error {
	return m.cuandoSeRegistraIntegranteConEmail(nombre, emailDeNombre(nombre), rol)
}

func (m *mundo) cuandoSeRegistraIntegranteConEmail(nombre, email, rol string) error {
	proyectos, err := m.proyectos()
	if err != nil {
		return err
	}
	return m.registrarAccion(proyectos.RegistrarIntegrante(m.proyectoActual, nombre, email, rol))
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

// cuandoSeConsultaElEstadoDelProyecto guarda el resultado aunque haya error: un escenario de error (CA-004-6) sigue
// hasta el `Entonces la operacion se rechaza`.
func (m *mundo) cuandoSeConsultaElEstadoDelProyecto(nombre string) error {
	proyectos, err := m.proyectos()
	if err != nil {
		return err
	}
	estado, err := proyectos.ConsultarEstadoProyecto(nombre)
	m.estadoProyecto = estado
	return m.registrarAccion(err)
}

func (m *mundo) entoncesElEstadoDelProyectoEs(esperado string) error {
	if m.estadoProyecto.Estado != esperado {
		return fmt.Errorf("el estado del proyecto es %q y se esperaba %q", m.estadoProyecto.Estado, esperado)
	}
	return nil
}

func (m *mundo) entoncesLosDiasDelEstadoSon(esperados int) error {
	if m.estadoProyecto.Dias != esperados {
		return fmt.Errorf("los días del estado son %d y se esperaban %d", m.estadoProyecto.Dias, esperados)
	}
	return nil
}

func (m *mundo) entoncesLasHistoriasDelProyectoSon(pendientes, enProgreso, completadas int) error {
	e := m.estadoProyecto
	if e.Pendientes != pendientes || e.EnProgreso != enProgreso || e.Completadas != completadas {
		return fmt.Errorf("el proyecto tiene %d pendientes, %d en progreso y %d completadas; se esperaban %d, %d y %d",
			e.Pendientes, e.EnProgreso, e.Completadas, pendientes, enProgreso, completadas)
	}
	return nil
}

func (m *mundo) entoncesLosStoryPointsDelProyectoSon(puntos, completados int) error {
	e := m.estadoProyecto
	if e.StoryPoints != puntos || e.Completados != completados {
		return fmt.Errorf("los story points del proyecto son %d y los completados %d; se esperaban %d y %d", e.StoryPoints, e.Completados, puntos, completados)
	}
	return nil
}

func (m *mundo) entoncesElPorcentajeDelProyectoEs(esperado string) error {
	return compararCifra("porcentaje de historias completadas del proyecto", esperado, m.estadoProyecto.Porcentaje)
}

func (m *mundo) entoncesElSprintAbiertoDelProyectoEs(esperado string) error {
	if m.estadoProyecto.SprintAbierto != esperado {
		return fmt.Errorf("el sprint abierto del proyecto es %q y se esperaba %q", m.estadoProyecto.SprintAbierto, esperado)
	}
	return nil
}

func (m *mundo) entoncesElProyectoNoTieneSprintAbierto() error {
	if m.estadoProyecto.SprintAbierto != "" {
		return fmt.Errorf("el proyecto tiene el sprint abierto %q y se esperaba ninguno", m.estadoProyecto.SprintAbierto)
	}
	return nil
}
