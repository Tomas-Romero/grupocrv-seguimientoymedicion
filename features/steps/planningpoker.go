package steps

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/cucumber/godog"
)

// Steps del area PlanningPoker (US-014 a US-018). La sesion se nombra por su
// historia: una historia tiene como maximo una sesion abierta (RN-014-7).

func registrarPlanningPoker(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^una sesión de Planning Poker sobre la historia "([^"]*)"$`, m.dadoUnaSesion)
	sc.Step(`^el integrante "([^"]*)" ya votó (-?\d+) en la sesión de la historia "([^"]*)"$`, m.dadoUnVoto)
	sc.Step(`^el integrante "([^"]*)" reveló los votos de la sesión de la historia "([^"]*)"$`, m.dadoRevelo)

	sc.Step(`^se inicia una sesión de Planning Poker sobre la historia "([^"]*)"$`, m.cuandoSeIniciaSesion)
	sc.Step(`^el integrante "([^"]*)" vota (-?\d+) en la sesión de la historia "([^"]*)"$`, m.cuandoVota)
	sc.Step(`^el integrante "([^"]*)" revela los votos de la sesión de la historia "([^"]*)"$`, m.cuandoRevela)
	sc.Step(`^el integrante "([^"]*)" registra la estimación acordada de (-?\d+) en la sesión de la historia "([^"]*)"$`, m.cuandoRegistraAcuerdo)
	sc.Step(`^el integrante "([^"]*)" abre una nueva ronda en la sesión de la historia "([^"]*)"$`, m.cuandoAbreNuevaRonda)

	sc.Step(`^la sesión de la historia "([^"]*)" está "([^"]*)" en la ronda (\d+)$`, m.entoncesEstadoDeLaSesion)
	sc.Step(`^los participantes de la sesión de la historia "([^"]*)" son:$`, m.entoncesParticipantes)
	sc.Step(`^en la sesión de la historia "([^"]*)" votaron:$`, m.entoncesVotaron)
	sc.Step(`^en la sesión de la historia "([^"]*)" faltan votar:$`, m.entoncesFaltanVotar)
	sc.Step(`^los votos de la sesión de la historia "([^"]*)" están ocultos$`, m.entoncesVotosOcultos)
	sc.Step(`^los votos de la sesión de la historia "([^"]*)" son:$`, m.entoncesVotosSon)
	sc.Step(`^los votos de la ronda (\d+) de la sesión de la historia "([^"]*)" son:$`, m.entoncesVotosDeRondaSon)
	sc.Step(`^hay divergencia en la sesión de la historia "([^"]*)"$`, m.entoncesHayDivergencia)
	sc.Step(`^no hay divergencia en la sesión de la historia "([^"]*)"$`, m.entoncesNoHayDivergencia)
	sc.Step(`^la estimación acordada de la sesión de la historia "([^"]*)" es (\d+)$`, m.entoncesEstimacionAcordada)
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

func (m *mundo) dadoRevelo(integrante, titulo string) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	if err := poker.Revelar(integrante, titulo); err != nil {
		return fmt.Errorf("revelar de %q la sesion de %q del contexto: %w", integrante, titulo, err)
	}
	return nil
}

func (m *mundo) cuandoRevela(integrante, titulo string) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	return m.registrarAccion(poker.Revelar(integrante, titulo))
}

// entoncesVotosSon compara los votos revelados con la tabla (integrante | voto), en orden.
// La primera fila es el encabezado.
func (m *mundo) entoncesVotosSon(titulo string, tabla *godog.Table) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	esperados, err := tablaDeVotos(tabla)
	if err != nil {
		return err
	}
	votos, err := poker.VotosRevelados(titulo)
	if err != nil {
		return fmt.Errorf("leer los votos de la sesion de %q: %w", titulo, err)
	}
	if !slices.Equal(votos, esperados) {
		return fmt.Errorf("votos de la sesion de %q: se esperaban %v, son %v", titulo, esperados, votos)
	}
	return nil
}

// entoncesVotosDeRondaSon compara los votos de una ronda de la sesion con la tabla
// (integrante | voto), en orden (US-017).
func (m *mundo) entoncesVotosDeRondaSon(ronda int, titulo string, tabla *godog.Table) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	esperados, err := tablaDeVotos(tabla)
	if err != nil {
		return err
	}
	votos, err := poker.VotosDeRonda(titulo, ronda)
	if err != nil {
		return fmt.Errorf("leer los votos de la ronda %d de la sesion de %q: %w", ronda, titulo, err)
	}
	if !slices.Equal(votos, esperados) {
		return fmt.Errorf("votos de la ronda %d de la sesion de %q: se esperaban %v, son %v", ronda, titulo, esperados, votos)
	}
	return nil
}

// tablaDeVotos lee una tabla integrante | voto. La primera fila es el encabezado.
func tablaDeVotos(tabla *godog.Table) ([]VotoRevelado, error) {
	if len(tabla.Rows) == 0 {
		return nil, fmt.Errorf("la tabla de votos no tiene encabezado")
	}
	votos := make([]VotoRevelado, 0, len(tabla.Rows)-1)
	for i, fila := range tabla.Rows[1:] {
		if len(fila.Cells) != 2 {
			return nil, fmt.Errorf("la fila %d de la tabla tiene %d columnas, se esperaban dos (integrante y voto)", i+2, len(fila.Cells))
		}
		valor, err := strconv.Atoi(fila.Cells[1].Value)
		if err != nil {
			return nil, fmt.Errorf("el voto %q de la fila %d no es un numero: %w", fila.Cells[1].Value, i+2, err)
		}
		votos = append(votos, VotoRevelado{Integrante: fila.Cells[0].Value, Valor: valor})
	}
	return votos, nil
}

func (m *mundo) entoncesHayDivergencia(titulo string) error {
	return m.comprobarDivergencia(titulo, true)
}

func (m *mundo) entoncesNoHayDivergencia(titulo string) error {
	return m.comprobarDivergencia(titulo, false)
}

func (m *mundo) comprobarDivergencia(titulo string, esperada bool) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	hay, err := poker.HayDivergencia(titulo)
	if err != nil {
		return fmt.Errorf("consultar la divergencia de la sesion de %q: %w", titulo, err)
	}
	if hay != esperada {
		return fmt.Errorf("divergencia en la sesion de %q: se esperaba %t, es %t", titulo, esperada, hay)
	}
	return nil
}

func (m *mundo) cuandoRegistraAcuerdo(integrante string, valor int, titulo string) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	return m.registrarAccion(poker.RegistrarEstimacionAcordada(integrante, titulo, valor))
}

func (m *mundo) entoncesEstimacionAcordada(titulo string, esperada int) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	valor, err := poker.EstimacionAcordada(titulo)
	if err != nil {
		return fmt.Errorf("consultar la estimacion acordada de la sesion de %q: %w", titulo, err)
	}
	if valor != esperada {
		return fmt.Errorf("estimacion acordada de la sesion de %q: se esperaba %d, es %d", titulo, esperada, valor)
	}
	return nil
}

func (m *mundo) cuandoAbreNuevaRonda(integrante, titulo string) error {
	poker, err := m.planningPoker()
	if err != nil {
		return err
	}
	return m.registrarAccion(poker.AbrirNuevaRonda(integrante, titulo))
}
