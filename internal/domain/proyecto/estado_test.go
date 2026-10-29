package proyecto_test

import (
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// El proyecto de estos tests va del 1 al 31 de diciembre de 2026.
func proyectoDeDiciembre() proyecto.Proyecto {
	return proyecto.Proyecto{Nombre: "Demo", FechaInicio: fecha(2026, time.December, 1), FechaFin: fecha(2026, time.December, 31)}
}

// US-004 / RN-004-1, RN-004-2, CA-004-1, CA-004-2, CA-004-4, CL-004-1 a CL-004-3: el estado se deriva de las fechas y
// los dias son de calendario. El primer y el ultimo dia son «en curso».
func TestEstadoEn(t *testing.T) {
	casos := []struct {
		nombre string
		hoy    time.Time
		estado proyecto.EstadoDelProyecto
		dias   int
	}{
		{"CA-004-1: en el medio, faltan los dias hasta el fin", fecha(2026, time.December, 6), proyecto.EnCurso, 25},
		{"CL-004-1: el dia de inicio ya esta en curso", fecha(2026, time.December, 1), proyecto.EnCurso, 30},
		{"CL-004-2: el dia de fin esta en curso con 0 dias", fecha(2026, time.December, 31), proyecto.EnCurso, 0},
		{"CL-004-2: al dia siguiente esta finalizado hace 1 dia", fecha(2027, time.January, 1), proyecto.Finalizado, 1},
		{"CA-004-2: terminado hace 3 dias", fecha(2027, time.January, 3), proyecto.Finalizado, 3},
		{"CL-004-3: el dia anterior al inicio empieza en 1 dia", fecha(2026, time.November, 30), proyecto.NoIniciado, 1},
		{"CA-004-2: empieza dentro de 12 dias", fecha(2026, time.November, 19), proyecto.NoIniciado, 12},
		{"cruza el cambio de mes y de anio", fecha(2026, time.January, 1), proyecto.NoIniciado, 334},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			estado, dias := proyectoDeDiciembre().EstadoEn(c.hoy)
			if estado != c.estado || dias != c.dias {
				t.Errorf("EstadoEn(%s) = (%q, %d), se esperaba (%q, %d)", c.hoy.Format(time.DateOnly), estado, dias, c.estado, c.dias)
			}
		})
	}
}

// RN-004-2 / CL-004-4: un proyecto de un solo dia esta en curso ese dia, con 0 dias; antes no inicio y despues termino.
func TestEstadoEn_UnSoloDia(t *testing.T) {
	p := proyecto.Proyecto{Nombre: "Un dia", FechaInicio: fecha(2026, time.October, 6), FechaFin: fecha(2026, time.October, 6)}

	if estado, dias := p.EstadoEn(fecha(2026, time.October, 6)); estado != proyecto.EnCurso || dias != 0 {
		t.Errorf("ese dia = (%q, %d), se esperaba (en_curso, 0)", estado, dias)
	}
	if estado, dias := p.EstadoEn(fecha(2026, time.October, 5)); estado != proyecto.NoIniciado || dias != 1 {
		t.Errorf("el dia anterior = (%q, %d), se esperaba (no_iniciado, 1)", estado, dias)
	}
	if estado, dias := p.EstadoEn(fecha(2026, time.October, 7)); estado != proyecto.Finalizado || dias != 1 {
		t.Errorf("el dia siguiente = (%q, %d), se esperaba (finalizado, 1)", estado, dias)
	}
}

// RN-004-2 / CL-004-5: un «hoy» con hora es el mismo dia; no cambia el estado ni los dias, tampoco con otra zona horaria
// (se comparan solo anio, mes y dia en UTC).
func TestEstadoEn_HoyConHora(t *testing.T) {
	p := proyectoDeDiciembre()
	tarde := time.Date(2026, time.December, 31, 23, 59, 59, 0, time.UTC)
	if estado, dias := p.EstadoEn(tarde); estado != proyecto.EnCurso || dias != 0 {
		t.Errorf("31/12 a las 23:59 = (%q, %d), se esperaba (en_curso, 0)", estado, dias)
	}
	manana := time.Date(2027, time.January, 1, 0, 0, 1, 0, time.UTC)
	if estado, dias := p.EstadoEn(manana); estado != proyecto.Finalizado || dias != 1 {
		t.Errorf("1/1 a las 00:00:01 = (%q, %d), se esperaba (finalizado, 1)", estado, dias)
	}
	// 31/12 a las 22:00 en Argentina (UTC-3) es el 1/1 a la 1:00 en UTC: el dominio compara en UTC.
	argentina := time.FixedZone("ART", -3*60*60)
	enArgentina := time.Date(2026, time.December, 31, 22, 0, 0, 0, argentina)
	if estado, _ := p.EstadoEn(enArgentina); estado != proyecto.Finalizado {
		t.Errorf("31/12 22:00 ART = %q, se esperaba finalizado: es el 1/1 en UTC", estado)
	}
}

// RN-004-3 / CL-004-6: sin fechas (valor cero) no falla ni entra en panico: es «no iniciado».
func TestEstadoEn_SinFechas(t *testing.T) {
	for _, p := range []proyecto.Proyecto{
		{},
		{FechaInicio: fecha(2026, time.December, 1)},
		{FechaFin: fecha(2026, time.December, 31)},
	} {
		if estado, dias := p.EstadoEn(fecha(2026, time.October, 6)); estado != proyecto.NoIniciado || dias != 0 {
			t.Errorf("EstadoEn con %+v = (%q, %d), se esperaba (no_iniciado, 0)", p, estado, dias)
		}
	}
}

// RN-004-1 / CL-004-12: el estado depende solo de las fechas; los valores son los que se muestran y se guardan en los
// escenarios (minuscula, sin tilde, con guion bajo).
func TestEstadoDelProyecto_Valores(t *testing.T) {
	for estado, esperado := range map[proyecto.EstadoDelProyecto]string{
		proyecto.NoIniciado: "no_iniciado", proyecto.EnCurso: "en_curso", proyecto.Finalizado: "finalizado",
	} {
		if string(estado) != esperado {
			t.Errorf("el estado %q deberia valer %q", estado, esperado)
		}
	}
}
