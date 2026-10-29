package steps

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
)

// Steps del area Metricas (US-025, US-026, US-027, US-028, US-029 y US-030).

func registrarMetricas(sc *godog.ScenarioContext, m *mundo) {
	sc.Step(`^el sprint "([^"]*)" con una historia de (-?\d+) story points completada$`, m.dadoHistoriaDeSprintCompletada)
	sc.Step(`^el sprint "([^"]*)" con una historia de (-?\d+) story points sin completar$`, m.dadoHistoriaDeSprintSinCompletar)
	sc.Step(`^el sprint cerrado "([^"]*)" con (\d+) story points completados de (\d+) planificados$`, m.dadoSprintCerradoConPuntos)

	sc.Step(`^se consultan los story points planificados del sprint "([^"]*)"$`, m.cuandoSeConsultanPlanificados)
	sc.Step(`^se consultan los story points completados del sprint "([^"]*)"$`, m.cuandoSeConsultanCompletados)
	sc.Step(`^se consulta la velocidad del equipo$`, m.cuandoSeConsultaVelocidad)

	sc.Step(`^los story points planificados son (\d+)$`, m.entoncesPlanificadosSon)
	sc.Step(`^los story points completados son (\d+)$`, m.entoncesCompletadosSon)
	sc.Step(`^la velocidad del equipo es (\d+(?:[.,]\d+)?)$`, m.entoncesVelocidadEs)

	sc.Step(`^el sprint "([^"]*)" con (-?\d+(?:[.,]\d+)?) horas registradas$`, m.dadoHorasRegistradas)
	sc.Step(`^se consultan las horas del sprint "([^"]*)"$`, m.cuandoSeConsultanLasHoras)
	sc.Step(`^las horas estimadas son (-?\d+(?:[.,]\d+)?)$`, m.entoncesHorasEstimadasSon)
	sc.Step(`^las horas reales son (-?\d+(?:[.,]\d+)?)$`, m.entoncesHorasRealesSon)
	sc.Step(`^la desviación de horas es (-?\d+(?:[.,]\d+)?)$`, m.entoncesDesviacionDeHorasEs)
	sc.Step(`^la desviación porcentual es (-?\d+(?:[.,]\d+)?)$`, m.entoncesDesviacionPorcentualEs)
	sc.Step(`^la desviación porcentual no se puede calcular$`, m.entoncesDesviacionPorcentualNoSePuedeCalcular)

	sc.Step(`^se consulta el porcentaje de historias completadas del sprint "([^"]*)"$`, m.cuandoSeConsultaPorcentajeDeHistorias)
	sc.Step(`^el porcentaje de historias completadas es (\d+(?:[.,]\d+)?)$`, m.entoncesPorcentajeDeHistoriasEs)

	// US-021: comparar el esfuerzo estimado contra el real, historia por historia.
	sc.Step(`^se compara el esfuerzo del proyecto "([^"]*)"$`, m.cuandoSeComparaElEsfuerzo)
	sc.Step(`^la comparación tiene (\d+) historias?$`, m.entoncesLaComparacionTieneHistorias)
	sc.Step(`^la primera historia de la comparación es "([^"]*)"$`, m.entoncesLaPrimeraHistoriaDeLaComparacionEs)
	sc.Step(`^la historia "([^"]*)" tiene (-?\d+(?:[.,]\d+)?) horas estimadas y (-?\d+(?:[.,]\d+)?) reales en la comparación$`, m.entoncesHorasDeLaHistoriaComparada)
	sc.Step(`^la desviación de la historia "([^"]*)" en la comparación es (-?\d+(?:[.,]\d+)?)$`, m.entoncesDesviacionDeLaHistoriaComparada)
	sc.Step(`^la desviación porcentual de la historia "([^"]*)" en la comparación es (-?\d+(?:[.,]\d+)?)$`, m.entoncesPorcentajeDeLaHistoriaComparada)
	sc.Step(`^la desviación porcentual de la historia "([^"]*)" en la comparación no se puede calcular$`, m.entoncesPorcentajeDeLaHistoriaComparadaNoSePuedeCalcular)
	sc.Step(`^el total comparable tiene (\d+) historias?$`, m.entoncesElTotalComparableTieneHistorias)
	sc.Step(`^las horas sin comparar son (\d+(?:[.,]\d+)?)$`, m.entoncesLasHorasSinCompararSon)

	sc.Step(`^un defecto detectado en el sprint (-?\d+) y resuelto en el sprint (-?\d+)$`, m.dadoDefectoResuelto)
	sc.Step(`^un defecto detectado en el sprint (-?\d+) sin resolver$`, m.dadoDefectoSinResolver)
	sc.Step(`^se consultan los defectos del sprint (-?\d+)$`, m.cuandoSeConsultanLosDefectos)
	sc.Step(`^los defectos detectados son (\d+)$`, m.entoncesDefectosDetectadosSon)
	sc.Step(`^los defectos resueltos son (\d+)$`, m.entoncesDefectosResueltosSon)

	registrarDashboard(sc, m)
}

// dadoHistoriaDeSprintCompletada y dadoHistoriaDeSprintSinCompletar arman el
// sprint de un escenario de metricas. Aceptan numeros negativos a proposito: es
// la unica forma de escribir el escenario de error de US-025 (CA-025-4).
func (m *mundo) dadoHistoriaDeSprintCompletada(sprint string, puntos int) error {
	return m.agregarHistoriaDeSprint(sprint, puntos, true)
}

func (m *mundo) dadoHistoriaDeSprintSinCompletar(sprint string, puntos int) error {
	return m.agregarHistoriaDeSprint(sprint, puntos, false)
}

func (m *mundo) agregarHistoriaDeSprint(sprint string, puntos int, completada bool) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	if err = metricas.AgregarItemAlSprint(sprint, puntos, completada); err != nil {
		return fmt.Errorf("agregar una historia al sprint %q del contexto: %w", sprint, err)
	}
	return nil
}

// dadoSprintCerradoConPuntos arma un sprint cerrado para el escenario de la
// velocidad. No valida nada: "8 completados de 5 planificados" es justo el dato
// corrupto que tiene que rechazar el dominio (CA-026-4).
func (m *mundo) dadoSprintCerradoConPuntos(nombre string, completados, planificados int) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	if err = metricas.RegistrarSprintCerrado(nombre, planificados, completados); err != nil {
		return fmt.Errorf("registrar el sprint cerrado %q del contexto: %w", nombre, err)
	}
	return nil
}

func (m *mundo) cuandoSeConsultanPlanificados(sprint string) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	puntos, err := metricas.PuntosPlanificados(sprint)
	m.puntosPlanificados = puntos
	return m.registrarAccion(err)
}

func (m *mundo) cuandoSeConsultanCompletados(sprint string) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	puntos, err := metricas.PuntosCompletados(sprint)
	m.puntosCompletados = puntos
	return m.registrarAccion(err)
}

func (m *mundo) cuandoSeConsultaVelocidad() error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	velocidad, err := metricas.VelocidadDelEquipo()
	m.velocidad = velocidad
	return m.registrarAccion(err)
}

func (m *mundo) entoncesPlanificadosSon(esperados int) error {
	if m.puntosPlanificados != esperados {
		return fmt.Errorf("story points planificados: se esperaban %d, se obtuvieron %d", esperados, m.puntosPlanificados)
	}
	return nil
}

func (m *mundo) entoncesCompletadosSon(esperados int) error {
	if m.puntosCompletados != esperados {
		return fmt.Errorf("story points completados: se esperaban %d, se obtuvieron %d", esperados, m.puntosCompletados)
	}
	return nil
}

// entoncesVelocidadEs acepta tanto "25" como "25.5" y "25,5": el escenario lo
// escribe una persona y la coma decimal es lo natural en espanol.
func (m *mundo) entoncesVelocidadEs(valor string) error {
	esperada, err := strconv.ParseFloat(strings.Replace(valor, ",", ".", 1), 64)
	if err != nil {
		return fmt.Errorf("la velocidad esperada %q no es un número: %w", valor, err)
	}
	const tolerancia = 0.001
	if math.Abs(m.velocidad-esperada) > tolerancia {
		return fmt.Errorf("velocidad del equipo: se esperaba %g, se obtuvo %g", esperada, m.velocidad)
	}
	return nil
}

// dadoHorasRegistradas suma un registro de esfuerzo al sprint del contexto. Acepta
// 0 y numeros negativos a proposito: es la unica forma de escribir los escenarios de
// error de US-027 (CA-027-5).
func (m *mundo) dadoHorasRegistradas(sprint, horas string) error {
	valor, err := numeroDelEscenario(horas)
	if err != nil {
		return err
	}
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	if err = metricas.RegistrarHorasDelSprint(sprint, valor); err != nil {
		return fmt.Errorf("registrar horas en el sprint %q del contexto: %w", sprint, err)
	}
	return nil
}

func (m *mundo) cuandoSeConsultanLasHoras(sprint string) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	horas, err := metricas.HorasDelSprint(sprint)
	m.horas = horas
	return m.registrarAccion(err)
}

func (m *mundo) entoncesHorasEstimadasSon(esperadas string) error {
	return compararCifra("horas estimadas", esperadas, m.horas.Estimadas)
}

func (m *mundo) entoncesHorasRealesSon(esperadas string) error {
	return compararCifra("horas reales", esperadas, m.horas.Reales)
}

func (m *mundo) entoncesDesviacionDeHorasEs(esperada string) error {
	return compararCifra("desviación de horas", esperada, m.horas.Desviacion)
}

func (m *mundo) entoncesDesviacionPorcentualEs(esperada string) error {
	if !m.horas.PorcentajeCalculable {
		return fmt.Errorf("la desviación porcentual no se puede calcular y se esperaba %s", esperada)
	}
	return compararCifra("desviación porcentual", esperada, m.horas.Porcentaje)
}

func (m *mundo) entoncesDesviacionPorcentualNoSePuedeCalcular() error {
	if m.horas.PorcentajeCalculable {
		return fmt.Errorf("la desviación porcentual se pudo calcular (%g) y se esperaba que no", m.horas.Porcentaje)
	}
	return nil
}

// cuandoSeConsultaPorcentajeDeHistorias guarda el resultado aunque haya error: un escenario
// de error (CA-028-5) sigue hasta el `Entonces la operacion se rechaza`.
func (m *mundo) cuandoSeConsultaPorcentajeDeHistorias(sprint string) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	porcentaje, err := metricas.PorcentajeHistoriasCompletadas(sprint)
	m.porcentajeHistorias = porcentaje
	return m.registrarAccion(err)
}

func (m *mundo) entoncesPorcentajeDeHistoriasEs(esperado string) error {
	return compararCifra("porcentaje de historias completadas", esperado, m.porcentajeHistorias)
}

// numeroDelEscenario lee un numero como lo escribe una persona: acepta coma o punto
// decimal, y uno negativo.
func numeroDelEscenario(texto string) (float64, error) {
	valor, err := strconv.ParseFloat(strings.Replace(texto, ",", ".", 1), 64)
	if err != nil {
		return 0, fmt.Errorf("%q no es un número: %w", texto, err)
	}
	return valor, nil
}

// compararCifra compara lo que dice el escenario con lo que calculo el dominio, con
// la tolerancia que ya usa la velocidad.
func compararCifra(nombre, esperada string, obtenida float64) error {
	quiero, err := numeroDelEscenario(esperada)
	if err != nil {
		return err
	}
	const tolerancia = 0.001
	if math.Abs(obtenida-quiero) > tolerancia {
		return fmt.Errorf("%s: se esperaba %g, se obtuvo %g", nombre, quiero, obtenida)
	}
	return nil
}

// cuandoSeComparaElEsfuerzo guarda la comparacion aunque haya error. Deja ademas el total de las
// historias comparables en el resultado de horas, para reutilizar las frases de US-027
// (`las horas estimadas son N`, `la desviación porcentual es N`...).
func (m *mundo) cuandoSeComparaElEsfuerzo(proyecto string) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	comparacion, err := metricas.CompararEsfuerzo(proyecto)
	m.comparacion = comparacion
	m.horas = comparacion.Total
	return m.registrarAccion(err)
}

func (m *mundo) entoncesLaComparacionTieneHistorias(esperadas int) error {
	if obtenidas := len(m.comparacion.Historias); obtenidas != esperadas {
		return fmt.Errorf("historias de la comparación: se esperaban %d, hay %d", esperadas, obtenidas)
	}
	return nil
}

func (m *mundo) entoncesLaPrimeraHistoriaDeLaComparacionEs(titulo string) error {
	if len(m.comparacion.Historias) == 0 {
		return fmt.Errorf("la comparación no tiene historias y se esperaba %q primero", titulo)
	}
	if primera := m.comparacion.Historias[0].Titulo; primera != titulo {
		return fmt.Errorf("primera historia de la comparación: se esperaba %q, es %q", titulo, primera)
	}
	return nil
}

// historiaComparada busca la historia por titulo en la comparacion.
func (m *mundo) historiaComparada(titulo string) (HistoriaComparada, error) {
	for _, h := range m.comparacion.Historias {
		if h.Titulo == titulo {
			return h, nil
		}
	}
	return HistoriaComparada{}, fmt.Errorf("la historia %q no aparece en la comparación", titulo)
}

func (m *mundo) entoncesHorasDeLaHistoriaComparada(titulo, estimadas, reales string) error {
	h, err := m.historiaComparada(titulo)
	if err != nil {
		return err
	}
	if err := compararCifra("horas estimadas de "+titulo, estimadas, h.Horas.Estimadas); err != nil {
		return err
	}
	return compararCifra("horas reales de "+titulo, reales, h.Horas.Reales)
}

func (m *mundo) entoncesDesviacionDeLaHistoriaComparada(titulo, esperada string) error {
	h, err := m.historiaComparada(titulo)
	if err != nil {
		return err
	}
	return compararCifra("desviación de "+titulo, esperada, h.Horas.Desviacion)
}

func (m *mundo) entoncesPorcentajeDeLaHistoriaComparada(titulo, esperado string) error {
	h, err := m.historiaComparada(titulo)
	if err != nil {
		return err
	}
	if !h.Horas.PorcentajeCalculable {
		return fmt.Errorf("la desviación porcentual de %q no se puede calcular y se esperaba %s", titulo, esperado)
	}
	return compararCifra("desviación porcentual de "+titulo, esperado, h.Horas.Porcentaje)
}

func (m *mundo) entoncesPorcentajeDeLaHistoriaComparadaNoSePuedeCalcular(titulo string) error {
	h, err := m.historiaComparada(titulo)
	if err != nil {
		return err
	}
	if h.Horas.PorcentajeCalculable {
		return fmt.Errorf("la desviación porcentual de %q se pudo calcular (%g) y se esperaba que no", titulo, h.Horas.Porcentaje)
	}
	return nil
}

func (m *mundo) entoncesElTotalComparableTieneHistorias(esperadas int) error {
	if obtenidas := m.comparacion.Comparables; obtenidas != esperadas {
		return fmt.Errorf("historias comparables: se esperaban %d, hay %d", esperadas, obtenidas)
	}
	return nil
}

func (m *mundo) entoncesLasHorasSinCompararSon(esperadas string) error {
	return compararCifra("horas sin comparar", esperadas, m.comparacion.HorasSinComparar)
}

// dadoDefectoResuelto y dadoDefectoSinResolver arman los defectos del contexto de US-029. Aceptan
// numeros negativos y resoluciones anteriores a la deteccion a proposito: es la unica forma de
// escribir los escenarios de error (CA-029-5).
func (m *mundo) dadoDefectoResuelto(detectadoEn, resueltoEn int) error {
	return m.agregarDefecto(detectadoEn, resueltoEn)
}

func (m *mundo) dadoDefectoSinResolver(detectadoEn int) error {
	return m.agregarDefecto(detectadoEn, 0)
}

func (m *mundo) agregarDefecto(detectadoEn, resueltoEn int) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	if err = metricas.RegistrarDefecto(detectadoEn, resueltoEn); err != nil {
		return fmt.Errorf("registrar un defecto del contexto: %w", err)
	}
	return nil
}

func (m *mundo) cuandoSeConsultanLosDefectos(sprint int) error {
	metricas, err := m.metricas()
	if err != nil {
		return err
	}
	defectos, err := metricas.DefectosDelSprint(sprint)
	m.defectos = defectos
	return m.registrarAccion(err)
}

func (m *mundo) entoncesDefectosDetectadosSon(esperados int) error {
	if m.defectos.Detectados != esperados {
		return fmt.Errorf("defectos detectados: se esperaban %d, se obtuvieron %d", esperados, m.defectos.Detectados)
	}
	return nil
}

func (m *mundo) entoncesDefectosResueltosSon(esperados int) error {
	if m.defectos.Resueltos != esperados {
		return fmt.Errorf("defectos resueltos: se esperaban %d, se obtuvieron %d", esperados, m.defectos.Resueltos)
	}
	return nil
}
