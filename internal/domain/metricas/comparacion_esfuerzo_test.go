package metricas_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
)

func historia(id, titulo string, puntos int, completada bool, horas ...float64) metricas.HistoriaParaComparar {
	return metricas.HistoriaParaComparar{ItemID: id, Titulo: titulo, StoryPoints: puntos, Completada: completada, Horas: horas}
}

func idsDeComparacion(c metricas.ComparacionProyecto) []string {
	ids := make([]string, len(c.Historias))
	for i, h := range c.Historias {
		ids[i] = h.ItemID
	}
	return ids
}

// US-021 (CA-021-1, CA-021-2, CA-021-4 y CA-021-5) / RN-021-1, RN-021-2 y RN-021-6: cada historia
// se compara con el mismo calculo de US-027.
func TestCompararHistorias_UnaHistoria(t *testing.T) {
	casos := []struct {
		nombre     string
		historia   metricas.HistoriaParaComparar
		estimadas  float64
		reales     float64
		desviacion float64
		porcentaje float64
		calculable bool
		estimada   bool
	}{
		// CA-021-1 / CL-021-1: 6,67 se redondea a 6,7.
		{"5 story points y 8 horas", historia("h", "H", 5, true, 3, 5), 7.5, 8, 0.5, 6.7, true, true},
		// CA-021-2
		{"8 story points y 9 horas en tres registros", historia("h", "H", 8, true, 2, 3, 4), 12, 9, -3, -25, true, true},
		// CL-021-2
		{"horas iguales a las estimadas", historia("h", "H", 3, true, 4.5), 4.5, 4.5, 0, 0, true, true},
		// CA-021-4 / CL-021-3: completada y estimada, nadie cargo horas.
		{"completada y estimada sin horas", historia("h", "H", 8, true), 12, 0, -12, -100, true, true},
		// CA-021-5 / CL-021-4 / RN-021-6: sin estimar, el porcentaje no se puede calcular.
		{"sin estimar con horas", historia("h", "H", 0, false, 4), 0, 4, 4, 0, false, false},
		// CL-021-17: la suma es exacta, sin ruido de coma flotante.
		{"cuartos de hora", historia("h", "H", 1, true, 0.25, 0.25, 0.5), 1.5, 1, -0.5, -33.3, true, true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			comparacion, err := metricas.CompararHistorias([]metricas.HistoriaParaComparar{c.historia})
			if err != nil {
				t.Fatalf("CompararHistorias: %v", err)
			}
			if len(comparacion.Historias) != 1 {
				t.Fatalf("historias = %d, se esperaba 1", len(comparacion.Historias))
			}
			h := comparacion.Historias[0]
			if h.Horas.Estimadas != c.estimadas || h.Horas.Reales != c.reales || h.Horas.Desviacion != c.desviacion {
				t.Errorf("horas = %+v, se esperaba estimadas %v, reales %v y desviacion %v", h.Horas, c.estimadas, c.reales, c.desviacion)
			}
			porcentaje, calculable := h.DesviacionPorcentual()
			if calculable != c.calculable || (calculable && porcentaje != c.porcentaje) {
				t.Errorf("porcentaje = %v (calculable %v), se esperaba %v (%v)", porcentaje, calculable, c.porcentaje, c.calculable)
			}
			if h.Estimada != c.estimada || h.StoryPoints != c.historia.StoryPoints || h.Completada != c.historia.Completada {
				t.Errorf("historia = %+v, se esperaba estimada=%v, %d story points y completada=%v", h, c.estimada, c.historia.StoryPoints, c.historia.Completada)
			}
			if h.ItemID != "h" || h.Titulo != "H" || h.Registros != len(c.historia.Horas) {
				t.Errorf("identidad = %+v, se esperaba ID h, titulo H y %d registros", h, len(c.historia.Horas))
			}
		})
	}
}

// US-021 (CA-021-6) / RN-021-3: aparece la que tiene horas, o la que esta completada y
// estimada. Las demas no tienen nada para comparar.
func TestCompararHistorias_QueAparece(t *testing.T) {
	historias := []metricas.HistoriaParaComparar{
		historia("pendiente-estimada", "A", 5, false),                 // CL-021-5: no aparece
		historia("completada-sin-estimar", "B", 0, true),              // CL-021-6: no aparece
		historia("pendiente-sin-estimar", "C", 0, false),              // no aparece
		historia("en-curso-con-horas", "D", 5, false, 2),              // CL-021-7: aparece
		historia("completada-estimada-sin-horas", "E", 2, true),       // CL-021-3: aparece
		historia("completada-sin-estimar-con-horas", "F", 0, true, 1), // aparece
	}

	c, err := metricas.CompararHistorias(historias)
	if err != nil {
		t.Fatalf("CompararHistorias: %v", err)
	}

	aparecen := map[string]bool{}
	for _, id := range idsDeComparacion(c) {
		aparecen[id] = true
	}
	for _, id := range []string{"pendiente-estimada", "completada-sin-estimar", "pendiente-sin-estimar"} {
		if aparecen[id] {
			t.Errorf("%s aparece en la comparacion, se esperaba que no", id)
		}
	}
	for _, id := range []string{"en-curso-con-horas", "completada-estimada-sin-horas", "completada-sin-estimar-con-horas"} {
		if !aparecen[id] {
			t.Errorf("%s no aparece en la comparacion, se esperaba que si", id)
		}
	}
}

// US-021 (CA-021-5 y CA-021-6) / RN-021-4 y RN-021-5, CL-021-4, CL-021-7 y CL-021-14: solo suman al
// total las historias completadas y estimadas; las horas del resto quedan en HorasSinComparar.
func TestCompararHistorias_SoloLasCompletadasYEstimadasSumanAlTotal(t *testing.T) {
	c, err := metricas.CompararHistorias([]metricas.HistoriaParaComparar{
		historia("a", "A", 8, true, 9),   // comparable
		historia("b", "B", 5, true, 8),   // comparable
		historia("c", "C", 0, false, 4),  // sin estimar: no comparable
		historia("d", "D", 13, false, 3), // en curso: no comparable
		historia("e", "E", 0, true, 1.5), // completada sin estimar: no comparable
	})
	if err != nil {
		t.Fatalf("CompararHistorias: %v", err)
	}

	// CL-021-11: estimadas 19,5, reales 17 y desviacion -2,5 (-12,8 %).
	if c.Total.Estimadas != 19.5 || c.Total.Reales != 17 || c.Total.Desviacion != -2.5 {
		t.Errorf("total = %+v, se esperaba estimadas 19.5, reales 17 y desviacion -2.5", c.Total)
	}
	if porcentaje, calculable := c.Total.DesviacionPorcentual(); !calculable || porcentaje != -12.8 {
		t.Errorf("porcentaje del total = %v (%v), se esperaba -12.8", porcentaje, calculable)
	}
	if c.Comparables != 2 {
		t.Errorf("comparables = %d, se esperaba 2", c.Comparables)
	}
	if c.HorasSinComparar != 8.5 {
		t.Errorf("horas sin comparar = %v, se esperaba 8.5 (4 + 3 + 1,5)", c.HorasSinComparar)
	}
}

// US-021 / CL-021-14: hay historias, pero ninguna es comparable.
func TestCompararHistorias_NingunaComparable(t *testing.T) {
	c, err := metricas.CompararHistorias([]metricas.HistoriaParaComparar{
		historia("a", "A", 0, false, 4),
		historia("b", "B", 5, false, 2),
	})
	if err != nil {
		t.Fatalf("CompararHistorias: %v", err)
	}

	if c.Total != (metricas.ResumenHoras{}) || c.Comparables != 0 || c.HorasSinComparar != 6 {
		t.Errorf("comparacion = %+v, se esperaba total en 0, ninguna comparable y 6 horas sin comparar", c)
	}
	if _, calculable := c.Total.DesviacionPorcentual(); calculable {
		t.Error("el porcentaje del total es calculable, se esperaba que no")
	}
}

// US-021 (CA-021-7) / RN-021-10, CL-021-13 y CL-021-16: sin historias el resultado es vacio y no es
// un error; una lista nil es lo mismo.
func TestCompararHistorias_SinHistorias(t *testing.T) {
	for nombre, entrada := range map[string][]metricas.HistoriaParaComparar{"vacia": {}, "nil": nil} {
		c, err := metricas.CompararHistorias(entrada)
		if err != nil {
			t.Errorf("%s: se esperaba sin error, llego %v", nombre, err)
			continue
		}
		if len(c.Historias) != 0 || c.Total != (metricas.ResumenHoras{}) || c.Comparables != 0 || c.HorasSinComparar != 0 {
			t.Errorf("%s: comparacion = %+v, se esperaba vacia", nombre, c)
		}
		if _, calculable := c.Total.DesviacionPorcentual(); calculable {
			t.Errorf("%s: el porcentaje del total es calculable, se esperaba que no", nombre)
		}
	}
}

// US-021 (CA-021-3) / RN-021-7: primero la historia donde mas nos equivocamos, para arriba o para
// abajo, por el valor absoluto de la desviacion en horas y no por el porcentaje.
func TestCompararHistorias_Orden(t *testing.T) {
	casos := []struct {
		nombre    string
		historias []metricas.HistoriaParaComparar
		esperado  []string
	}{
		{
			// CL-021-10: -3 sale antes que +0,5.
			nombre:    "por desviacion absoluta",
			historias: []metricas.HistoriaParaComparar{historia("chica", "Chica", 5, true, 8), historia("grande", "Grande", 8, true, 9)},
			esperado:  []string{"grande", "chica"},
		},
		{
			// Una de 1 punto con 1 hora de mas (+66,7 %) no tapa a una de 13 con 8 horas de mas (+41 %).
			nombre:    "no por porcentaje",
			historias: []metricas.HistoriaParaComparar{historia("uno", "Uno", 1, true, 2.5), historia("trece", "Trece", 13, true, 27.5)},
			esperado:  []string{"trece", "uno"},
		},
		{
			// CL-021-12: la misma desviacion en valor absoluto se ordena por titulo, sin distinguir mayusculas.
			nombre:    "empate por titulo",
			historias: []metricas.HistoriaParaComparar{historia("b", "beta", 2, true, 4), historia("a", "Alfa", 2, true, 2)},
			esperado:  []string{"a", "b"},
		},
		{
			// Si ademas empatan en el titulo, por ID: el orden es total.
			nombre:    "empate por titulo e ID",
			historias: []metricas.HistoriaParaComparar{historia("z", "Igual", 2, true, 4), historia("m", "Igual", 2, true, 4)},
			esperado:  []string{"m", "z"},
		},
		{
			// Las que no se pueden comparar tambien se ordenan por su desviacion (igual a sus horas).
			nombre:    "sin estimar entre las demas",
			historias: []metricas.HistoriaParaComparar{historia("a", "A", 2, true, 3.5), historia("b", "B", 0, false, 6)},
			esperado:  []string{"b", "a"},
		},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			comparacion, err := metricas.CompararHistorias(c.historias)
			if err != nil {
				t.Fatalf("CompararHistorias: %v", err)
			}
			if obtenido := idsDeComparacion(comparacion); !reflect.DeepEqual(obtenido, c.esperado) {
				t.Errorf("orden = %v, se esperaba %v", obtenido, c.esperado)
			}
		})
	}
}

// US-021 / RN-021-1: no modifica la lista que recibe.
func TestCompararHistorias_NoModificaLaEntrada(t *testing.T) {
	entrada := []metricas.HistoriaParaComparar{historia("chica", "Chica", 5, true, 8), historia("grande", "Grande", 8, true, 9)}

	if _, err := metricas.CompararHistorias(entrada); err != nil {
		t.Fatalf("CompararHistorias: %v", err)
	}

	if entrada[0].ItemID != "chica" || entrada[1].ItemID != "grande" {
		t.Errorf("la entrada quedo reordenada: %v, %v", entrada[0].ItemID, entrada[1].ItemID)
	}
}

// US-021 (CA-021-9) / RN-021-9 y CL-021-15: los datos invalidos se rechazan, todos juntos y con la
// posicion de la historia, y no se devuelve comparacion.
func TestCompararHistorias_DatosInvalidos(t *testing.T) {
	c, err := metricas.CompararHistorias([]metricas.HistoriaParaComparar{
		historia("a", "A", -1, true, 2),
		historia("b", "B", 3, true, 1, 0),
		historia("c", "C", 2, true, -4),
		historia("d", "D", 2, true, 3), // valida
	})

	if err == nil {
		t.Fatal("se esperaba un error")
	}
	if !errors.Is(err, metricas.ErrStoryPointsNegativos) || !errors.Is(err, metricas.ErrHorasInvalidas) {
		t.Errorf("error = %v, se esperaba que reconociera ErrStoryPointsNegativos y ErrHorasInvalidas", err)
	}
	for _, posicion := range []string{"historia 1", "historia 2", "historia 3"} {
		if !strings.Contains(err.Error(), posicion) {
			t.Errorf("error = %q, se esperaba que indicara %q", err, posicion)
		}
	}
	if strings.Contains(err.Error(), "historia 4") {
		t.Errorf("error = %q, la historia 4 es valida", err)
	}
	if len(c.Historias) != 0 || c.Total != (metricas.ResumenHoras{}) {
		t.Errorf("comparacion = %+v, con error no se devuelve comparacion", c)
	}
}

// US-021 / RN-021-9: los datos de una historia que no aparecería también se validan.
func TestCompararHistorias_ValidaTambienLasQueNoAparecen(t *testing.T) {
	_, err := metricas.CompararHistorias([]metricas.HistoriaParaComparar{historia("a", "A", -5, false)})

	if !errors.Is(err, metricas.ErrStoryPointsNegativos) {
		t.Errorf("error = %v, se esperaba ErrStoryPointsNegativos", err)
	}
}
