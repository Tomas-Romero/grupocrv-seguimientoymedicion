package app_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
)

// lecturaDeComparacionFalsa simula la lectura de US-021 sin base de datos.
type lecturaDeComparacionFalsa struct {
	proyectos map[string]bool
	historias []app.HistoriaConHoras
	err       error
	lecturas  int
}

func (l *lecturaDeComparacionFalsa) ExisteProyecto(_ context.Context, id string) (bool, error) {
	return l.proyectos[id], l.err
}

func (l *lecturaDeComparacionFalsa) LeerHistoriasConHoras(context.Context, string) ([]app.HistoriaConHoras, error) {
	l.lecturas++
	return append([]app.HistoriaConHoras(nil), l.historias...), l.err
}

func conHoras(item backlog.ItemBacklog, horas ...float64) app.HistoriaConHoras {
	return app.HistoriaConHoras{Item: item, Horas: horas}
}

func compararEsf(t *testing.T, historias ...app.HistoriaConHoras) (metricas.ComparacionProyecto, error) {
	t.Helper()
	repo := &lecturaDeComparacionFalsa{proyectos: map[string]bool{"p1": true}, historias: historias}
	return app.NuevoCompararEsfuerzo(repo).Ejecutar(context.Background(), "p1")
}

func titulosDe(c metricas.ComparacionProyecto) []string {
	titulos := make([]string, len(c.Historias))
	for i, h := range c.Historias {
		titulos[i] = h.Titulo
	}
	return titulos
}

// US-021 / CA-021-3, CA-021-6, RN-021-3, RN-021-4 y RN-021-7: el caso de uso arma la comparacion con
// la estimacion actual y el estado de cada historia. Una historia completada es comparable; una en
// curso aparece si tiene horas pero no suma al total; una pendiente sin horas no aparece.
func TestCompararEsfuerzo(t *testing.T) {
	c, err := compararEsf(t,
		conHoras(historiaDe("a", 1, backlog.PrioridadMust, 8, backlog.EstadoCompletado), 4, 5),
		conHoras(historiaDe("b", 2, backlog.PrioridadMust, 5, backlog.EstadoCompletado), 8),
		conHoras(historiaDe("c", 3, backlog.PrioridadShould, 3, backlog.EstadoPendiente)),
		conHoras(historiaDe("d", 4, backlog.PrioridadShould, 5, backlog.EstadoEnProgreso), 2),
		conHoras(historiaDe("e", 5, backlog.PrioridadCould, 0, backlog.EstadoPendiente)),
	)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	// d: 7,5 estimadas y 2 reales (-5,5); a: 12 y 9 (-3); b: 7,5 y 8 (+0,5).
	if esperado := []string{"Historia d", "Historia a", "Historia b"}; !reflect.DeepEqual(titulosDe(c), esperado) {
		t.Errorf("historias = %v, se esperaba %v", titulosDe(c), esperado)
	}
	if c.Comparables != 2 || c.Total.Estimadas != 19.5 || c.Total.Reales != 17 || c.Total.Desviacion != -2.5 {
		t.Errorf("total = %+v con %d comparables, se esperaba 19.5, 17 y -2.5 con 2", c.Total, c.Comparables)
	}
	if c.HorasSinComparar != 2 {
		t.Errorf("horas sin comparar = %v, se esperaba 2 (las de la historia en curso)", c.HorasSinComparar)
	}
	d := c.Historias[0]
	if d.Completada || !d.Estimada || d.StoryPoints != 5 || d.Registros != 1 || d.ItemID != "d" {
		t.Errorf("historia en curso = %+v, se esperaba estimada, no completada, 5 story points y 1 registro", d)
	}
}

// US-021 / RN-021-6: una historia sin estimar llega con 0 story points y no es estimada.
func TestCompararEsfuerzo_SinEstimar(t *testing.T) {
	c, err := compararEsf(t, conHoras(historiaDe("a", 1, backlog.PrioridadMust, 0, backlog.EstadoEnProgreso), 4))
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if len(c.Historias) != 1 || c.Historias[0].Estimada || c.Historias[0].Horas.Desviacion != 4 || c.HorasSinComparar != 4 {
		t.Errorf("comparacion = %+v, se esperaba una historia sin estimar con 4 horas sin comparar", c)
	}
}

// US-021 (CA-021-7) / RN-021-10: un proyecto sin historias da una comparacion vacia.
func TestCompararEsfuerzo_SinHistorias(t *testing.T) {
	c, err := compararEsf(t)

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if len(c.Historias) != 0 || c.Comparables != 0 || c.Total != (metricas.ResumenHoras{}) {
		t.Errorf("comparacion = %+v, se esperaba vacia", c)
	}
}

// US-021 (CA-021-8): un proyecto inexistente se rechaza sin leer nada mas.
func TestCompararEsfuerzo_ProyectoInexistente(t *testing.T) {
	repo := &lecturaDeComparacionFalsa{proyectos: map[string]bool{}}

	_, err := app.NuevoCompararEsfuerzo(repo).Ejecutar(context.Background(), "p1")

	if !errors.Is(err, app.ErrProyectoInexistente) || repo.lecturas != 0 {
		t.Errorf("error = %v, lecturas = %d; se esperaba ErrProyectoInexistente y ninguna lectura", err, repo.lecturas)
	}
}

// US-021 (CA-021-9) / RN-021-9: horas guardadas que no son validas (datos del sistema que salieron
// mal) se devuelven con el error de metricas y sin comparacion.
func TestCompararEsfuerzo_DatosInvalidos(t *testing.T) {
	c, err := compararEsf(t, conHoras(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado), 0))

	if !errors.Is(err, metricas.ErrHorasInvalidas) {
		t.Errorf("error = %v, se esperaba ErrHorasInvalidas", err)
	}
	if len(c.Historias) != 0 {
		t.Errorf("con error no se devuelve comparacion, llegaron %d historias", len(c.Historias))
	}
}

// US-021: los fallos de la persistencia se devuelven envueltos, sin traducir.
func TestCompararEsfuerzo_ErrorDeLaPersistencia(t *testing.T) {
	falla := errors.New("conexion perdida")
	repo := &lecturaDeComparacionFalsa{proyectos: map[string]bool{"p1": true}, err: falla}

	_, err := app.NuevoCompararEsfuerzo(repo).Ejecutar(context.Background(), "p1")

	if !errors.Is(err, falla) {
		t.Errorf("error = %v, se esperaba que envolviera %v", err, falla)
	}
}

// US-021 / RN-021-11: la comparacion hace una sola lectura.
func TestCompararEsfuerzo_UnaSolaLectura(t *testing.T) {
	repo := &lecturaDeComparacionFalsa{proyectos: map[string]bool{"p1": true},
		historias: []app.HistoriaConHoras{conHoras(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado), 8)}}

	if _, err := app.NuevoCompararEsfuerzo(repo).Ejecutar(context.Background(), "p1"); err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if repo.lecturas != 1 {
		t.Errorf("lecturas = %d, se esperaba 1", repo.lecturas)
	}
}
