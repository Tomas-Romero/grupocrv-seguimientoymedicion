package metricas

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
)

// HistoriaParaComparar es lo que CompararHistorias necesita de una historia: no recibe el item del
// backlog, asi metricas no depende de backlog (igual que ItemDelSprint en US-025).
type HistoriaParaComparar struct {
	ItemID string
	Titulo string
	// StoryPoints es la estimacion actual de la historia; 0 si no esta estimada.
	StoryPoints int
	Completada  bool
	// Horas son las de todos sus registros de esfuerzo, de cualquier fecha, integrante y sprint
	// (RN-021-2).
	Horas []float64
}

// ComparacionHistoria es una historia con lo estimado contra lo trabajado.
type ComparacionHistoria struct {
	ItemID      string
	Titulo      string
	StoryPoints int
	// Estimada es false si la historia no tiene estimacion (RN-021-6).
	Estimada   bool
	Completada bool
	// Registros es la cantidad de registros de esfuerzo de la historia.
	Registros int
	Horas     ResumenHoras
}

// DesviacionPorcentual es la de sus horas (RN-027-4): no se puede calcular si la historia no esta
// estimada.
func (c ComparacionHistoria) DesviacionPorcentual() (porcentaje float64, calculable bool) {
	return c.Horas.DesviacionPorcentual()
}

// comparable dice si la estimacion y el esfuerzo son del mismo trabajo terminado (RN-021-4).
func (c ComparacionHistoria) comparable() bool {
	return c.Estimada && c.Completada
}

// ComparacionProyecto es la comparacion de todas las historias que aparecen.
type ComparacionProyecto struct {
	// Historias salen primero donde mas nos equivocamos (RN-021-7).
	Historias []ComparacionHistoria
	// Total suma solo las historias comparables (RN-021-4).
	Total       ResumenHoras
	Comparables int
	// HorasSinComparar son las horas reales de las historias que aparecen y no son comparables
	// (RN-021-5).
	HorasSinComparar float64
}

// CompararHistorias compara, historia por historia, las horas que implica su estimacion contra las
// que se registraron (US-021). Cada historia se calcula con CalcularHoras (RN-021-1).
//
// Aparece toda historia con al menos un registro de esfuerzo, o completada y estimada (RN-021-3).
// Solo las completadas y estimadas suman al total (RN-021-4); las horas de las otras van a
// HorasSinComparar (RN-021-5).
//
// Valida todas las historias, aparezcan o no (RN-021-9): devuelve los errores de CalcularHoras
// unidos con errors.Join, cada uno envuelto con la posicion de la historia contada desde 1, y
// ninguna comparacion. Una lista vacia o nil es una comparacion vacia, no un error (RN-021-10).
func CompararHistorias(historias []HistoriaParaComparar) (ComparacionProyecto, error) {
	// conHoras lleva las horas de cada historia que aparece, para sumar las comparables.
	type conHoras struct {
		comparacion ComparacionHistoria
		horas       []float64
	}
	var (
		errs     []error
		aparecen []conHoras
	)
	for i, h := range historias {
		horas, err := CalcularHoras(h.StoryPoints, h.Horas)
		if err != nil {
			errs = append(errs, fmt.Errorf("historia %d (%s): %w", i+1, h.ItemID, err))
			continue
		}
		c := ComparacionHistoria{
			ItemID: h.ItemID, Titulo: h.Titulo, StoryPoints: h.StoryPoints,
			Estimada: h.StoryPoints > 0, Completada: h.Completada,
			Registros: len(h.Horas), Horas: horas,
		}
		if c.Registros > 0 || (c.Completada && c.Estimada) {
			aparecen = append(aparecen, conHoras{comparacion: c, horas: h.Horas})
		}
	}
	if err := errors.Join(errs...); err != nil {
		return ComparacionProyecto{}, err
	}

	var (
		comparacion   ComparacionProyecto
		puntosTotales int
		horasTotales  []float64
		sinComparar   float64
	)
	for _, a := range aparecen {
		if !a.comparacion.comparable() {
			sinComparar += a.comparacion.Horas.Reales
			continue
		}
		comparacion.Comparables++
		puntosTotales += a.comparacion.StoryPoints
		horasTotales = append(horasTotales, a.horas...)
	}
	// Los datos ya se validaron arriba, asi que este calculo no puede fallar.
	total, err := CalcularHoras(puntosTotales, horasTotales)
	if err != nil {
		return ComparacionProyecto{}, fmt.Errorf("sumar las historias comparables: %w", err)
	}
	comparacion.Total = total
	comparacion.HorasSinComparar = redondear(sinComparar, 100)

	slices.SortStableFunc(aparecen, func(a, b conHoras) int {
		return cmp.Or(
			cmp.Compare(math.Abs(b.comparacion.Horas.Desviacion), math.Abs(a.comparacion.Horas.Desviacion)),
			cmp.Compare(strings.ToLower(a.comparacion.Titulo), strings.ToLower(b.comparacion.Titulo)),
			cmp.Compare(a.comparacion.ItemID, b.comparacion.ItemID),
		)
	})
	for _, a := range aparecen {
		comparacion.Historias = append(comparacion.Historias, a.comparacion)
	}
	return comparacion, nil
}
