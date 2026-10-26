package esfuerzo_test

import (
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
)

// Dos sprints seguidos con un hueco de un dia entre ellos: el 1 va del 6 al 19 de octubre y el 2,
// del 21 de octubre al 3 de noviembre.
func rangos() []esfuerzo.RangoSprint {
	return []esfuerzo.RangoSprint{
		{ID: "s1", Numero: 1, Inicio: dia(2026, 10, 6), Fin: dia(2026, 10, 19)},
		{ID: "s2", Numero: 2, Inicio: dia(2026, 10, 21), Fin: dia(2026, 11, 3)},
	}
}

// US-020 / RN-020-3, CA-020-4: la fecha cae en el sprint cuyo rango la contiene.
func TestSprintDeLaFecha_DentroDelRango(t *testing.T) {
	r, ok := esfuerzo.SprintDeLaFecha(dia(2026, 10, 10), rangos())
	if !ok {
		t.Fatal("ok = false, se esperaba el Sprint 1")
	}
	if r.ID != "s1" || r.Numero != 1 {
		t.Errorf("sprint = %+v, se esperaba el Sprint 1", r)
	}
}

// US-020 / RN-020-3, CL-020-3: el dia de inicio y el de fin son del sprint.
func TestSprintDeLaFecha_InicioYFinIncluidos(t *testing.T) {
	for nombre, fecha := range map[string]time.Time{"inicio": dia(2026, 10, 6), "fin": dia(2026, 10, 19)} {
		r, ok := esfuerzo.SprintDeLaFecha(fecha, rangos())
		if !ok || r.Numero != 1 {
			t.Errorf("%s: sprint = %+v, ok = %v; se esperaba el Sprint 1", nombre, r, ok)
		}
	}
}

// US-020 / RN-020-3: se comparan solo el anio, el mes y el dia, no la hora.
func TestSprintDeLaFecha_IgnoraLaHora(t *testing.T) {
	r, ok := esfuerzo.SprintDeLaFecha(dia(2026, 10, 19).Add(23*time.Hour+59*time.Minute), rangos())
	if !ok || r.Numero != 1 {
		t.Errorf("sprint = %+v, ok = %v; el ultimo dia a las 23:59 es del Sprint 1", r, ok)
	}
}

// US-020 / RN-020-3, CL-020-4: una fecha entre dos sprints, antes del primero o despues del
// ultimo no es de ninguno.
func TestSprintDeLaFecha_FueraDeTodoSprint(t *testing.T) {
	for nombre, fecha := range map[string]time.Time{
		"entre dos sprints":  dia(2026, 10, 20),
		"antes del primero":  dia(2026, 10, 5),
		"despues del ultimo": dia(2026, 11, 4),
	} {
		if r, ok := esfuerzo.SprintDeLaFecha(fecha, rangos()); ok {
			t.Errorf("%s: sprint = %+v, se esperaba que no cayera en ninguno", nombre, r)
		}
	}
}

// US-020 / RN-020-3: sin sprints, ninguna fecha es de un sprint, y una lista nil es lo mismo.
func TestSprintDeLaFecha_SinSprints(t *testing.T) {
	if _, ok := esfuerzo.SprintDeLaFecha(dia(2026, 10, 10), nil); ok {
		t.Error("ok = true, se esperaba false sin sprints")
	}
}

// US-020 / RN-020-13, CL-020-14: si dos rangos contuvieran la misma fecha gana el de menor
// numero, sin importar el orden en que llegan.
func TestSprintDeLaFecha_RangosSuperpuestos(t *testing.T) {
	superpuestos := []esfuerzo.RangoSprint{
		{ID: "s3", Numero: 3, Inicio: dia(2026, 10, 10), Fin: dia(2026, 10, 20)},
		{ID: "s2", Numero: 2, Inicio: dia(2026, 10, 15), Fin: dia(2026, 10, 25)},
	}
	r, ok := esfuerzo.SprintDeLaFecha(dia(2026, 10, 17), superpuestos)
	if !ok || r.Numero != 2 {
		t.Errorf("sprint = %+v, ok = %v; se esperaba el Sprint 2 (el de menor numero)", r, ok)
	}
}

func registrosDe(horas ...float64) []esfuerzo.RegistroEsfuerzo {
	registros := make([]esfuerzo.RegistroEsfuerzo, len(horas))
	for i, h := range horas {
		registros[i] = esfuerzo.RegistroEsfuerzo{Horas: h}
	}
	return registros
}

// US-020 / RN-020-6: el total es la suma de las horas; sin registros es 0.
func TestTotalHoras(t *testing.T) {
	casos := []struct {
		nombre    string
		registros []esfuerzo.RegistroEsfuerzo
		esperado  float64
	}{
		{"sin registros", nil, 0},
		{"uno solo", registrosDe(2.5), 2.5},
		{"varios", registrosDe(2, 1.5, 4), 7.5},
		// CL-020-12: la suma es exacta, sin ruido de coma flotante.
		{"cuartos de hora", registrosDe(0.25, 0.25, 0.5), 1},
		{"redondea el ruido", registrosDe(0.1, 0.2), 0.3},
	}
	for _, c := range casos {
		if obtenido := esfuerzo.TotalHoras(c.registros); obtenido != c.esperado {
			t.Errorf("%s: total = %v, se esperaba %v", c.nombre, obtenido, c.esperado)
		}
	}
}
