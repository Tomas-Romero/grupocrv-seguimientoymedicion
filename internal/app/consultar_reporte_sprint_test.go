package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Reporte de un sprint (spec US-033). Igual que el del proyecto, compone lecturas que ya existen.

type fuentesDelReporteDelSprint struct {
	app.FuentesDelReporteDelSprint
	llamadas map[string]int
}

func nuevasFuentesDelReporteDelSprint(detalle app.DetalleDeSprint) *fuentesDelReporteDelSprint {
	f := &fuentesDelReporteDelSprint{llamadas: map[string]int{}}
	f.Proyecto = func(context.Context, string) (app.DetalleDeProyecto, error) {
		f.llamadas["proyecto"]++
		return app.DetalleDeProyecto{Proyecto: proyecto.Proyecto{ID: "p1"}}, nil
	}
	f.Sprint = func(context.Context, string, string) (app.DetalleDeSprint, error) {
		f.llamadas["sprint"]++
		return detalle, nil
	}
	f.Esfuerzo = func(context.Context, string, app.FiltroEsfuerzo) (app.ConsultaEsfuerzo, error) {
		f.llamadas["esfuerzo"]++
		return app.ConsultaEsfuerzo{}, nil
	}
	f.Defectos = func(context.Context, string, app.FiltroDefectos) ([]app.DefectoListado, error) {
		f.llamadas["defectos"]++
		return nil, nil
	}
	return f
}

func (f *fuentesDelReporteDelSprint) reporte() (app.ReporteDelSprint, error) {
	return app.NuevoConsultarReporteDelSprint(f.FuentesDelReporteDelSprint).Ejecutar(context.Background(), "p1", "s1")
}

// detalleDeUnSprint arma un sprint cerrado con tres historias: una completada de 5 Story Points, otra que volvio al backlog (3) y una sin
// estimar. Los Story Points planificados son 8.
func detalleDeUnSprint() app.DetalleDeSprint {
	return app.DetalleDeSprint{
		Sprint: sprint.Sprint{ID: "s1", Numero: 1, ProyectoID: "p1", SprintGoal: "MVP", Estado: sprint.EstadoCerrado, FechaInicio: dia(time.October, 1), FechaFin: dia(time.October, 14)},
		Historias: []app.HistoriaDelSprint{
			{Item: historiaDe("h1", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado), Completada: true, Resultado: sprint.ResultadoCompletado},
			{Item: historiaDe("h2", 2, backlog.PrioridadShould, 3, backlog.EstadoPendiente), Completada: false, Resultado: sprint.ResultadoNoCompletado},
		},
		Planificados: 8,
		Completados:  5,
	}
}

func registroConsultado(fecha time.Time, horas float64, historia string) app.RegistroConsultado {
	return app.RegistroConsultado{Registro: esfuerzo.RegistroEsfuerzo{Fecha: fecha, Horas: horas, Actividad: "Trabajo"}, Historia: historia, Integrante: "Ana", SprintNumero: 1}
}

// CA-033-2, RN-033-4: cada historia se muestra con el resultado de su fila en ese sprint, y las metricas salen de los totales del sprint, del
// porcentaje de historias completadas y de las horas de los registros con fecha dentro del sprint (1,5 horas por Story Point).
func TestConsultarReporteDelSprint_HistoriasYMetricas(t *testing.T) {
	f := nuevasFuentesDelReporteDelSprint(detalleDeUnSprint())
	f.Esfuerzo = func(_ context.Context, _ string, filtro app.FiltroEsfuerzo) (app.ConsultaEsfuerzo, error) {
		if filtro != (app.FiltroEsfuerzo{SprintID: "s1"}) {
			t.Errorf("el esfuerzo del sprint se pide filtrando por su ID, no por %+v", filtro)
		}
		return app.ConsultaEsfuerzo{
			Registros:  []app.RegistroConsultado{registroConsultado(dia(time.October, 3), 3, "Login"), registroConsultado(dia(time.October, 5), 5, "Login")},
			TotalHoras: 8,
		}, nil
	}

	r, err := f.reporte()

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if r.Sprint.Numero != 1 || r.Sprint.SprintGoal != "MVP" || r.Sprint.Estado != sprint.EstadoCerrado {
		t.Errorf("sprint = %+v", r.Sprint)
	}
	if len(r.Historias.Items) != 2 || !r.Historias.Items[0].Completada || r.Historias.Items[1].Completada {
		t.Errorf("historias = %+v, se esperaba la primera completada y la segunda no", r.Historias)
	}
	m := r.Metricas
	if m.Err != nil || m.Planificados != 8 || m.Completados != 5 || m.PorcentajeHistorias != 50 {
		t.Errorf("metricas = %+v, se esperaban 8 planificados, 5 completados y 50 %%", m)
	}
	if m.Horas.Estimadas != 12 || m.Horas.Reales != 8 || m.Horas.Desviacion != -4 {
		t.Errorf("horas = %+v, se esperaban 12 estimadas, 8 reales y -4 de desviacion", m.Horas)
	}
	if len(r.Esfuerzo.Registros) != 2 || r.Esfuerzo.TotalHoras != 8 {
		t.Errorf("esfuerzo = %+v", r.Esfuerzo)
	}
}

// CL-033-3, RN-033-4: una historia que no se completo en el Sprint 1 y se completo en el Sprint 2 es «no completada» en el primero y
// «completada» en el segundo: manda la fila de cada sprint y no el estado actual de la historia.
func TestConsultarReporteDelSprint_ElResultadoEsElDeLaFila(t *testing.T) {
	actual := historiaDe("h1", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado) // completada hoy, en el Sprint 2
	for _, c := range []struct {
		sprint    string
		resultado sprint.Resultado
		completa  bool
	}{{"Sprint 1", sprint.ResultadoNoCompletado, false}, {"Sprint 2", sprint.ResultadoCompletado, true}} {
		detalle := app.DetalleDeSprint{
			Sprint:       sprint.Sprint{ID: "s1", Numero: 1, ProyectoID: "p1", Estado: sprint.EstadoCerrado},
			Historias:    []app.HistoriaDelSprint{{Item: actual, Completada: c.completa, Resultado: c.resultado}},
			Planificados: 5,
		}
		r, err := nuevasFuentesDelReporteDelSprint(detalle).reporte()
		if err != nil {
			t.Fatalf("Ejecutar (%s): %v", c.sprint, err)
		}
		if r.Historias.Items[0].Completada != c.completa {
			t.Errorf("%s: Completada = %v, se esperaba %v", c.sprint, r.Historias.Items[0].Completada, c.completa)
		}
	}
}

// CA-033-3, CL-033-9: un sprint sin historias ni horas muestra todas las secciones vacias, con el porcentaje y las horas en 0 y sin error.
func TestConsultarReporteDelSprint_SprintSinDatos(t *testing.T) {
	vacio := app.DetalleDeSprint{Sprint: sprint.Sprint{ID: "s1", Numero: 1, ProyectoID: "p1", Estado: sprint.EstadoAbierto}}

	r, err := nuevasFuentesDelReporteDelSprint(vacio).reporte()

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if !r.Historias.Vacia() || !r.Esfuerzo.Vacia() || !r.Defectos.Vacia() {
		t.Errorf("las secciones tienen que estar vacias: %+v", r)
	}
	if r.Metricas.Err != nil || r.Metricas.PorcentajeHistorias != 0 || r.Metricas.Horas.Estimadas != 0 || r.Metricas.Horas.Reales != 0 {
		t.Errorf("metricas = %+v, se esperaba todo en 0 y sin error", r.Metricas)
	}
}

// CL-033-5, CL-033-6, RN-033-5: un defecto detectado en el sprint es «detectado» y uno resuelto en el sprint es «resuelto»; uno
// detectado en el Sprint 2 y resuelto en el 4 aparece como detectado en el del 2 y como resuelto en el del 4, y en ninguno mas. Los
// descartados no cuentan.
func TestConsultarReporteDelSprint_DefectosDetectadosYResueltos(t *testing.T) {
	d := func(id string, estado defecto.Estado, deteccion, resolucion string) app.DefectoListado {
		return app.DefectoListado{Defecto: defecto.Defecto{ID: id, Estado: estado, SprintDeteccionID: deteccion, SprintResolucionID: resolucion}}
	}
	defectos := []app.DefectoListado{
		d("a", defecto.EstadoAbierto, "s1", ""),      // detectado en este sprint
		d("b", defecto.EstadoResuelto, "s1", "s1"),   // detectado y resuelto en este sprint: en las dos listas
		d("c", defecto.EstadoResuelto, "s2", "s1"),   // detectado antes, resuelto en este
		d("d", defecto.EstadoResuelto, "s1", "s4"),   // detectado en este, resuelto despues
		d("e", defecto.EstadoDescartado, "s1", ""),   // descartado: no cuenta
		d("f", defecto.EstadoAbierto, "s2", ""),      // de otro sprint
		d("g", defecto.EstadoEnCorreccion, "s1", ""), // detectado en este sprint
		d("h", defecto.EstadoResuelto, "s2", "s4"),   // de otros sprints
	}
	f := nuevasFuentesDelReporteDelSprint(detalleDeUnSprint())
	f.Defectos = func(context.Context, string, app.FiltroDefectos) ([]app.DefectoListado, error) { return defectos, nil }

	r, err := f.reporte()

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	ids := func(l []app.DefectoListado) string {
		s := ""
		for _, x := range l {
			s += x.Defecto.ID
		}
		return s
	}
	if ids(r.Defectos.Detectados) != "abdg" {
		t.Errorf("detectados = %q, se esperaban a, b, d y g", ids(r.Defectos.Detectados))
	}
	if ids(r.Defectos.Resueltos) != "bc" {
		t.Errorf("resueltos = %q, se esperaban b y c", ids(r.Defectos.Resueltos))
	}
	if r.Defectos.Vacia() {
		t.Error("la seccion de defectos no esta vacia")
	}
}

// CA-033-4, RN-033-6: el proyecto que no existe se rechaza antes de leer el sprint, y un sprint que no existe o es de otro proyecto,
// antes de leer el esfuerzo y los defectos.
func TestConsultarReporteDelSprint_Inexistentes(t *testing.T) {
	t.Run("proyecto inexistente", func(t *testing.T) {
		f := nuevasFuentesDelReporteDelSprint(detalleDeUnSprint())
		f.Proyecto = func(context.Context, string) (app.DetalleDeProyecto, error) {
			return app.DetalleDeProyecto{}, errors.Join(app.ErrProyectoInexistente)
		}

		_, err := f.reporte()

		if !errors.Is(err, app.ErrProyectoInexistente) || len(f.llamadas) != 0 {
			t.Errorf("err = %v, llamadas = %v: se esperaba ErrProyectoInexistente sin leer nada mas", err, f.llamadas)
		}
	})
	t.Run("sprint inexistente o de otro proyecto", func(t *testing.T) {
		f := nuevasFuentesDelReporteDelSprint(detalleDeUnSprint())
		f.Sprint = func(context.Context, string, string) (app.DetalleDeSprint, error) {
			return app.DetalleDeSprint{}, errors.Join(app.ErrSprintInexistente)
		}

		_, err := f.reporte()

		if !errors.Is(err, app.ErrSprintInexistente) {
			t.Fatalf("err = %v, se esperaba ErrSprintInexistente", err)
		}
		if f.llamadas["esfuerzo"] != 0 || f.llamadas["defectos"] != 0 {
			t.Errorf("no se tenia que leer el esfuerzo ni los defectos: %v", f.llamadas)
		}
	})
	t.Run("falla la lectura del sprint", func(t *testing.T) {
		f := nuevasFuentesDelReporteDelSprint(detalleDeUnSprint())
		f.Sprint = func(context.Context, string, string) (app.DetalleDeSprint, error) {
			return app.DetalleDeSprint{}, errDeLectura
		}

		if _, err := f.reporte(); !errors.Is(err, errDeLectura) {
			t.Errorf("err = %v, se esperaba el error de la lectura envuelto", err)
		}
	})
}

// CL-033-7, RN-033-3: si falla el esfuerzo, las metricas conservan los Story Points y el porcentaje pero informan que las horas no se
// pudieron leer, y la seccion de esfuerzo informa el error; las historias y los defectos se muestran. Si fallan los defectos, solo ellos.
func TestConsultarReporteDelSprint_CadaSeccionFallaPorSeparado(t *testing.T) {
	t.Run("esfuerzo", func(t *testing.T) {
		f := nuevasFuentesDelReporteDelSprint(detalleDeUnSprint())
		f.Esfuerzo = func(context.Context, string, app.FiltroEsfuerzo) (app.ConsultaEsfuerzo, error) {
			return app.ConsultaEsfuerzo{}, errDeLectura
		}

		r, err := f.reporte()

		if err != nil {
			t.Fatalf("Ejecutar: %v", err)
		}
		if !errors.Is(r.Esfuerzo.Err, errDeLectura) || !errors.Is(r.Metricas.Err, errDeLectura) {
			t.Errorf("errores: esfuerzo = %v, metricas = %v", r.Esfuerzo.Err, r.Metricas.Err)
		}
		if r.Metricas.Planificados != 8 || r.Metricas.PorcentajeHistorias != 50 || len(r.Historias.Items) != 2 || r.Defectos.Err != nil {
			t.Errorf("el resto del reporte tiene que mostrarse: %+v", r)
		}
	})
	t.Run("defectos", func(t *testing.T) {
		f := nuevasFuentesDelReporteDelSprint(detalleDeUnSprint())
		f.Defectos = func(context.Context, string, app.FiltroDefectos) ([]app.DefectoListado, error) {
			return nil, errDeLectura
		}

		r, err := f.reporte()

		if err != nil || !errors.Is(r.Defectos.Err, errDeLectura) || r.Esfuerzo.Err != nil || r.Metricas.Err != nil {
			t.Errorf("err = %v, defectos = %v, esfuerzo = %v, metricas = %v", err, r.Defectos.Err, r.Esfuerzo.Err, r.Metricas.Err)
		}
	})
}

// RN-033-1: una sola lectura de cada tipo.
func TestConsultarReporteDelSprint_UnaLecturaDeCadaTipo(t *testing.T) {
	f := nuevasFuentesDelReporteDelSprint(detalleDeUnSprint())

	if _, err := f.reporte(); err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if len(f.llamadas) != 4 {
		t.Errorf("se llamaron %d lecturas distintas, se esperaban 4: %v", len(f.llamadas), f.llamadas)
	}
	for nombre, n := range f.llamadas {
		if n != 1 {
			t.Errorf("la lectura %q se llamo %d veces, se esperaba 1", nombre, n)
		}
	}
}
