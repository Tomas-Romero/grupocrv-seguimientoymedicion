package proyecto

import "time"

// EstadoDelProyecto es el estado general de un proyecto (US-004). Se deriva siempre de sus fechas y no se guarda:
// no puede quedar desactualizado cuando US-002 las modifica.
type EstadoDelProyecto string

// Los tres estados. Sus valores van en minuscula, sin tilde y con guion bajo, como los de las historias y los sprints.
const (
	NoIniciado EstadoDelProyecto = "no_iniciado"
	EnCurso    EstadoDelProyecto = "en_curso"
	Finalizado EstadoDelProyecto = "finalizado"
)

// EstadoEn devuelve el estado del proyecto el dia hoy y los dias que importan en ese estado (RN-004-1 y RN-004-2):
//
//   - NoIniciado: hoy es anterior al inicio; dias son los que faltan para que empiece (al menos 1).
//   - EnCurso: hoy esta entre el inicio y el fin, ambos incluidos; dias son los que faltan para el fin (0 el ultimo dia).
//   - Finalizado: hoy es posterior al fin; dias son los que pasaron desde el fin (al menos 1).
//
// Los dias son de calendario y se comparan solo anio, mes y dia en UTC, asi que una hora distinta no cambia el
// resultado. La funcion no valida las fechas (US-001 ya garantiza que el fin no es anterior al inicio) y no lee el
// reloj: hoy llega por parametro. Un proyecto sin fechas es NoIniciado con 0 dias, sin fallar (RN-004-3).
func (p Proyecto) EstadoEn(hoy time.Time) (EstadoDelProyecto, int) {
	if p.FechaInicio.IsZero() || p.FechaFin.IsZero() {
		return NoIniciado, 0
	}
	dia, inicio, fin := soloElDia(hoy), soloElDia(p.FechaInicio), soloElDia(p.FechaFin)
	switch {
	case dia.Before(inicio):
		return NoIniciado, diasEntre(dia, inicio)
	case dia.After(fin):
		return Finalizado, diasEntre(fin, dia)
	}
	return EnCurso, diasEntre(dia, fin)
}

// soloElDia deja el anio, el mes y el dia en UTC, a medianoche.
func soloElDia(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

// diasEntre cuenta los dias de calendario de desde a hasta; las dos fechas ya estan a medianoche UTC, que no tiene
// cambios de hora, asi que cada dia dura 24 horas.
func diasEntre(desde, hasta time.Time) int {
	return int(hasta.Sub(desde).Hours() / 24)
}
