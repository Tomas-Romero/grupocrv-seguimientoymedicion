package demo

import (
	"math"
	"testing"
	"time"
)

// Estas pruebas miran las tablas de datos.go contra lo que dice la spec T-023, sin base: son lo que hace que un cambio de cifras en la
// carga no pase sin cambiar tambien la spec.

var corte = dia(time.October, 30)

func puntosDeHistoria(n int) int {
	if n == historiaHistorial {
		return estimacionAcordadaHistorial
	}
	return historias[n-1].puntos
}

// RN-T023-2: ninguna fecha cargada es posterior al 30/10/2026, y las horas caen dentro del sprint al que pertenecen.
func TestDatos_FechasDentroDeCadaSprintYHastaElCorte(t *testing.T) {
	for _, r := range horas {
		s := sprints[r.sprint-1]
		if r.fecha.Before(s.inicio) || r.fecha.After(s.fin) {
			t.Errorf("horas del %s en la historia %d: fuera del Sprint %d (%s a %s)", r.fecha.Format("2006-01-02"), r.historia, r.sprint, s.inicio.Format("2006-01-02"), s.fin.Format("2006-01-02"))
		}
		if r.fecha.After(corte) {
			t.Errorf("horas del %s: posteriores al corte", r.fecha.Format("2006-01-02"))
		}
	}
	for n, s := range sprints {
		for _, c := range s.completadas {
			if c.fecha.Before(s.inicio) || c.fecha.After(s.fin) || c.fecha.After(corte) {
				t.Errorf("Sprint %d: la historia %d se completa el %s, fuera del sprint o despues del corte", n+1, c.historia, c.fecha.Format("2006-01-02"))
			}
		}
	}
}

// Los sprints duran 14 dias, son seguidos y caen dentro del proyecto (RN-008-6, RN-008-7 y RN-008-13).
func TestDatos_SprintsDe14DiasSeguidosDentroDelProyecto(t *testing.T) {
	for n, s := range sprints {
		if dias := int(s.fin.Sub(s.inicio).Hours()/24) + 1; dias != 14 {
			t.Errorf("Sprint %d dura %d dias", n+1, dias)
		}
		if s.inicio.Before(inicioProyecto) || s.fin.After(finProyecto) {
			t.Errorf("Sprint %d cae fuera del proyecto", n+1)
		}
		if n > 0 && !s.inicio.Equal(sprints[n-1].fin.AddDate(0, 0, 1)) {
			t.Errorf("Sprint %d no empieza el dia despues del anterior", n+1)
		}
		if s.cerrado != (n < 3) {
			t.Errorf("Sprint %d: cerrado = %t, solo los tres primeros se cierran", n+1, s.cerrado)
		}
	}
}

// Los Story Points planificados y completados de cada sprint son los de la spec, y todo lo que se asigna esta estimado (RN-009-7).
func TestDatos_StoryPointsPorSprint(t *testing.T) {
	esperados := [][2]int{{13, 10}, {11, 8}, {14, 11}, {12, 5}}
	for n, s := range sprints {
		var planificados, completados int
		for _, h := range s.historias {
			p := puntosDeHistoria(h)
			if p == 0 {
				t.Errorf("Sprint %d: la historia %d se asigna sin estimar", n+1, h)
			}
			planificados += p
		}
		for _, c := range s.completadas {
			completados += puntosDeHistoria(c.historia)
			if !contiene(s.historias, c.historia) {
				t.Errorf("Sprint %d: se completa la historia %d, que no esta en el sprint", n+1, c.historia)
			}
		}
		if planificados != esperados[n][0] || completados != esperados[n][1] {
			t.Errorf("Sprint %d: %d planificados y %d completados, se esperaban %d y %d", n+1, planificados, completados, esperados[n][0], esperados[n][1])
		}
	}
}

// Las horas son multiplos de 0,25, hay 27 registros con las cinco actividades y cada sprint tiene su desviacion contra 1,5 h por Story
// Point: +25,6 %, -10,6 %, 0 % y un sprint en curso.
func TestDatos_HorasYDesviaciones(t *testing.T) {
	totales := map[int]float64{}
	actividades := map[string]bool{}
	for _, r := range horas {
		if math.Mod(r.horas*4, 1) != 0 || r.horas <= 0 || r.horas > 24 {
			t.Errorf("horas invalidas: %v", r.horas)
		}
		totales[r.sprint] += r.horas
		actividades[r.actividad] = true
	}
	if len(horas) != 27 {
		t.Errorf("hay %d registros de horas y se esperaban 27", len(horas))
	}
	if len(actividades) != 5 {
		t.Errorf("hay %d actividades y se esperaban 5", len(actividades))
	}
	esperadas := map[int]float64{1: 24.5, 2: 14.75, 3: 21, 4: 11.5}
	for n, e := range esperadas {
		if totales[n] != e {
			t.Errorf("Sprint %d: %v h, se esperaban %v h", n, totales[n], e)
		}
	}
	for n, e := range map[int]float64{1: 25.6, 2: -10.6, 3: 0} {
		var planificados int
		for _, h := range sprints[n-1].historias {
			planificados += puntosDeHistoria(h)
		}
		estimadas := float64(planificados) * 1.5
		desviacion := (totales[n] - estimadas) / estimadas * 100
		if math.Abs(desviacion-e) > 0.05 {
			t.Errorf("Sprint %d: desviacion %.2f %%, se esperaba %.1f %%", n, desviacion, e)
		}
	}
}

// Los defectos son los de la spec: severidades, estados y sprints coherentes (RN-023-4: se resuelve en el sprint de deteccion o uno
// posterior), y el descartado no se resuelve.
func TestDatos_Defectos(t *testing.T) {
	if len(defectos) != 10 {
		t.Fatalf("hay %d defectos y se esperaban 10", len(defectos))
	}
	for _, d := range defectos {
		if d.historia < 1 || d.historia > len(historias) || d.sprintDeteccion < 1 || d.sprintDeteccion > len(sprints) {
			t.Errorf("defecto %q con historia o sprint inexistente", d.descripcion)
		}
		if d.estado == "resuelto" && d.sprintResolucion < d.sprintDeteccion {
			t.Errorf("defecto %q se resuelve antes de detectarse", d.descripcion)
		}
		if d.estado != "resuelto" && d.sprintResolucion != 0 {
			t.Errorf("defecto %q tiene sprint de resolucion sin estar resuelto", d.descripcion)
		}
	}
}

// Las historias son 16, con los numeros que usan los sprints, las sesiones y los defectos, y todas con al menos un criterio (RN-009-7).
func TestDatos_Historias(t *testing.T) {
	if len(historias) != 16 {
		t.Fatalf("hay %d historias y se esperaban 16", len(historias))
	}
	for i, h := range historias {
		if len(h.criterios) == 0 {
			t.Errorf("la historia %d no tiene criterios", i+1)
		}
	}
	for _, n := range []int{historiaHistorial, historiaCupones, historiaRecomendada} {
		if historias[n-1].puntos != 0 {
			t.Errorf("la historia %d la estima Planning Poker, no EstimarItem", n)
		}
	}
	if len(integrantes) != 4 {
		t.Errorf("hay %d integrantes y se esperaban 4", len(integrantes))
	}
}

func contiene(lista []int, n int) bool {
	for _, x := range lista {
		if x == n {
			return true
		}
	}
	return false
}
