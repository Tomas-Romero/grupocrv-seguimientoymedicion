package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Reporte de un proyecto (spec US-033). El caso de uso compone las lecturas que ya existen: los tests le pasan lecturas simuladas
// y cuentan cuantas veces se llama a cada una.

var errDeLectura = errors.New("la lectura fallo")

// fuentesDelReporte arma las ocho lecturas con datos vacios y cuenta las llamadas de cada una. Cada test cambia la que le interesa.
type fuentesDelReporte struct {
	app.FuentesDelReporteDelProyecto
	llamadas map[string]int
}

func nuevasFuentesDelReporte() *fuentesDelReporte {
	f := &fuentesDelReporte{llamadas: map[string]int{}}
	p := proyecto.Proyecto{ID: "p1", Nombre: "Demo", Descripcion: "Un proyecto", FechaInicio: dia(time.October, 1), FechaFin: dia(time.December, 15)}
	f.Proyecto = func(context.Context, string) (app.DetalleDeProyecto, error) {
		f.llamadas["proyecto"]++
		return app.DetalleDeProyecto{Proyecto: p}, nil
	}
	f.Backlog = func(context.Context, string) ([]backlog.ItemBacklog, error) { f.llamadas["backlog"]++; return nil, nil }
	f.Sprints = func(context.Context, string) ([]app.ResumenDeSprint, error) { f.llamadas["sprints"]++; return nil, nil }
	f.Anteriores = func(context.Context, string) (app.SprintsAnteriores, error) {
		f.llamadas["anteriores"]++
		return app.SprintsAnteriores{}, nil
	}
	f.Esfuerzo = func(context.Context, string, app.FiltroEsfuerzo) (app.ConsultaEsfuerzo, error) {
		f.llamadas["esfuerzo"]++
		return app.ConsultaEsfuerzo{}, nil
	}
	f.Comparacion = func(context.Context, string) (metricas.ComparacionProyecto, error) {
		f.llamadas["comparacion"]++
		return metricas.ComparacionProyecto{}, nil
	}
	f.Defectos = func(context.Context, string, app.FiltroDefectos) ([]app.DefectoListado, error) {
		f.llamadas["defectos"]++
		return nil, nil
	}
	f.Dashboard = func(context.Context, string) (app.Dashboard, error) {
		f.llamadas["dashboard"]++
		return app.Dashboard{Estado: app.EstadoGeneral{Proyecto: p}}, nil
	}
	return f
}

func (f *fuentesDelReporte) reporte() (app.ReporteDelProyecto, error) {
	return app.NuevoConsultarReporteDelProyecto(f.FuentesDelReporteDelProyecto).Ejecutar(context.Background(), "p1")
}

// CA-033-1, CL-033-1, RN-033-1, RN-033-2: con todo cargado, cada seccion trae lo que devuelve su lectura, sin recalcular nada.
func TestConsultarReporteDelProyecto_TodoCargado(t *testing.T) {
	f := nuevasFuentesDelReporte()
	f.Proyecto = func(context.Context, string) (app.DetalleDeProyecto, error) {
		return app.DetalleDeProyecto{
			Proyecto:    proyecto.Proyecto{ID: "p1", Nombre: "Demo"},
			Integrantes: []integrante.Integrante{{Nombre: "Ana", Email: "ana@x.test", Rol: integrante.RolAgileEnabler}},
		}, nil
	}
	f.Backlog = func(context.Context, string) ([]backlog.ItemBacklog, error) {
		return []backlog.ItemBacklog{historiaDe("h1", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado)}, nil
	}
	f.Sprints = func(context.Context, string) ([]app.ResumenDeSprint, error) {
		return []app.ResumenDeSprint{
			{Sprint: sprint.Sprint{ID: "s2", Numero: 2, Estado: sprint.EstadoAbierto}, Historias: 1, Planificados: 3},
			{Sprint: sprint.Sprint{ID: "s1", Numero: 1, Estado: sprint.EstadoCerrado}, Historias: 3, Planificados: 10, Completados: 8},
		}, nil
	}
	f.Anteriores = func(context.Context, string) (app.SprintsAnteriores, error) {
		return app.SprintsAnteriores{Sprints: []app.SprintAnterior{{Sprint: sprint.Sprint{ID: "s1", Numero: 1}, HistoriasNoCompletadas: 1}}}, nil
	}
	f.Esfuerzo = func(context.Context, string, app.FiltroEsfuerzo) (app.ConsultaEsfuerzo, error) {
		return app.ConsultaEsfuerzo{TotalHoras: 12.5}, nil
	}
	f.Comparacion = func(context.Context, string) (metricas.ComparacionProyecto, error) {
		return metricas.ComparacionProyecto{Comparables: 1, HorasSinComparar: 2, Total: metricas.ResumenHoras{Estimadas: 7.5, Reales: 8, Desviacion: 0.5}}, nil
	}
	f.Defectos = func(context.Context, string, app.FiltroDefectos) ([]app.DefectoListado, error) {
		return []app.DefectoListado{defectoDelDashboard(defecto.SeveridadAlta, defecto.EstadoAbierto, 1, 0)}, nil
	}
	f.Dashboard = func(context.Context, string) (app.Dashboard, error) {
		return app.Dashboard{
			Estado:     app.EstadoGeneral{Proyecto: proyecto.Proyecto{ID: "p1", Nombre: "Demo"}, StoryPoints: 18, Completados: 8, Porcentaje: 50},
			Defectos:   app.TarjetaDeDefectos{Total: 1, PorSeveridad: app.ConteoPorSeveridad{Alta: 1}, PorEstado: app.ConteoPorEstado{Abierto: 1}},
			Velocidad:  app.TarjetaDeVelocidad{Velocidad: 8, Cerrados: []app.SprintCerradoDelDashboard{{Sprint: sprint.Sprint{ID: "s1", Numero: 1}, Planificados: 10, Completados: 8}}},
			Referencia: &app.TarjetaDeSprint{Sprint: sprint.Sprint{ID: "s2", Numero: 2}, Planificados: 3},
		}, nil
	}

	r, err := f.reporte()

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if r.Proyecto.Proyecto.Nombre != "Demo" || r.Proyecto.Estado.StoryPoints != 18 || r.Proyecto.Estado.Porcentaje != 50 || r.Proyecto.Err != nil {
		t.Errorf("proyecto = %+v", r.Proyecto)
	}
	if len(r.Integrantes) != 1 || r.Integrantes[0].Nombre != "Ana" {
		t.Errorf("integrantes = %+v", r.Integrantes)
	}
	if len(r.Backlog.Items) != 1 || r.Backlog.Items[0].ID != "h1" {
		t.Errorf("backlog = %+v", r.Backlog)
	}
	if len(r.Sprints.Sprints) != 2 || r.Sprints.Sprints[0].Sprint.Numero != 2 {
		t.Fatalf("sprints = %+v, se esperaban 2 con el Sprint 2 primero", r.Sprints)
	}
	if r.Sprints.Sprints[1].VolvieronAlBacklog != 1 || r.Sprints.Sprints[0].VolvieronAlBacklog != 0 {
		t.Errorf("volvieron al backlog: Sprint 2 = %d y Sprint 1 = %d, se esperaban 0 y 1",
			r.Sprints.Sprints[0].VolvieronAlBacklog, r.Sprints.Sprints[1].VolvieronAlBacklog)
	}
	if r.Esfuerzo.TotalHoras != 12.5 || r.Esfuerzo.Comparacion.Comparables != 1 || r.Esfuerzo.Comparacion.HorasSinComparar != 2 {
		t.Errorf("esfuerzo = %+v", r.Esfuerzo)
	}
	if len(r.Defectos.Lista) != 1 || r.Defectos.Totales.Total != 1 || r.Defectos.Totales.PorSeveridad.Alta != 1 {
		t.Errorf("defectos = %+v", r.Defectos)
	}
	if r.Metricas.Velocidad.Velocidad != 8 || r.Metricas.Referencia == nil || r.Metricas.Referencia.Sprint.Numero != 2 {
		t.Errorf("metricas = %+v", r.Metricas)
	}
	for nombre, vacia := range map[string]bool{
		"integrantes": r.IntegrantesVacios(), "backlog": r.Backlog.Vacia(), "sprints": r.Sprints.Vacia(),
		"esfuerzo": r.Esfuerzo.Vacia(), "defectos": r.Defectos.Vacia(), "metricas": r.Metricas.Vacia(),
	} {
		if vacia {
			t.Errorf("la seccion %s no tiene que estar vacia", nombre)
		}
	}
}

// CA-033-3, CL-033-2, RN-033-2: un proyecto recien creado tiene todas las secciones, vacias y sin error.
func TestConsultarReporteDelProyecto_ProyectoNuevo(t *testing.T) {
	r, err := nuevasFuentesDelReporte().reporte()

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	for nombre, vacia := range map[string]bool{
		"integrantes": r.IntegrantesVacios(), "backlog": r.Backlog.Vacia(), "sprints": r.Sprints.Vacia(),
		"esfuerzo": r.Esfuerzo.Vacia(), "defectos": r.Defectos.Vacia(), "metricas": r.Metricas.Vacia(),
	} {
		if !vacia {
			t.Errorf("la seccion %s tiene que estar vacia", nombre)
		}
	}
	for nombre, e := range map[string]error{
		"proyecto": r.Proyecto.Err, "backlog": r.Backlog.Err, "sprints": r.Sprints.Err, "esfuerzo": r.Esfuerzo.Err,
		"defectos": r.Defectos.Err, "metricas": r.Metricas.Err,
	} {
		if e != nil {
			t.Errorf("la seccion %s informa un error: %v", nombre, e)
		}
	}
}

// CA-033-5, RN-033-1: una sola lectura de cada tipo, y la lectura de esfuerzo y la de defectos sin filtro.
func TestConsultarReporteDelProyecto_UnaLecturaDeCadaTipo(t *testing.T) {
	f := nuevasFuentesDelReporte()
	var filtroEsfuerzo app.FiltroEsfuerzo
	var filtroDefectos app.FiltroDefectos
	llamarEsfuerzo, llamarDefectos := f.Esfuerzo, f.Defectos
	f.Esfuerzo = func(ctx context.Context, id string, filtro app.FiltroEsfuerzo) (app.ConsultaEsfuerzo, error) {
		filtroEsfuerzo = filtro
		return llamarEsfuerzo(ctx, id, filtro)
	}
	f.Defectos = func(ctx context.Context, id string, filtro app.FiltroDefectos) ([]app.DefectoListado, error) {
		filtroDefectos = filtro
		return llamarDefectos(ctx, id, filtro)
	}

	if _, err := f.reporte(); err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	for nombre, n := range f.llamadas {
		if n != 1 {
			t.Errorf("la lectura %q se llamo %d veces, se esperaba 1", nombre, n)
		}
	}
	if len(f.llamadas) != 8 {
		t.Errorf("se llamaron %d lecturas distintas, se esperaban 8", len(f.llamadas))
	}
	if filtroEsfuerzo != (app.FiltroEsfuerzo{}) || filtroDefectos != (app.FiltroDefectos{}) {
		t.Errorf("las lecturas del reporte del proyecto van sin filtros: esfuerzo %+v, defectos %+v", filtroEsfuerzo, filtroDefectos)
	}
}

// CA-033-4, RN-033-6: un proyecto que no existe se rechaza antes de armar nada.
func TestConsultarReporteDelProyecto_ProyectoInexistente(t *testing.T) {
	f := nuevasFuentesDelReporte()
	f.Proyecto = func(context.Context, string) (app.DetalleDeProyecto, error) {
		return app.DetalleDeProyecto{}, errors.Join(app.ErrProyectoInexistente)
	}

	_, err := f.reporte()

	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Fatalf("err = %v, se esperaba ErrProyectoInexistente", err)
	}
	if len(f.llamadas) != 0 {
		t.Errorf("no se tenia que leer nada mas, se llamaron %v", f.llamadas)
	}
}

// Condiciones de error: si falla la lectura del proyecto no hay con que armar la pagina, y es un error envuelto.
func TestConsultarReporteDelProyecto_FallaLaLecturaDelProyecto(t *testing.T) {
	f := nuevasFuentesDelReporte()
	f.Proyecto = func(context.Context, string) (app.DetalleDeProyecto, error) {
		return app.DetalleDeProyecto{}, errDeLectura
	}

	_, err := f.reporte()

	if !errors.Is(err, errDeLectura) {
		t.Fatalf("err = %v, se esperaba el error de la lectura envuelto", err)
	}
}

// CL-033-7, RN-033-3: una seccion que no se puede armar informa su error y las demas se muestran.
func TestConsultarReporteDelProyecto_CadaSeccionFallaPorSeparado(t *testing.T) {
	casos := []struct {
		nombre string
		romper func(*fuentesDelReporte)
		error  func(app.ReporteDelProyecto) error
	}{
		{"backlog", func(f *fuentesDelReporte) {
			f.Backlog = func(context.Context, string) ([]backlog.ItemBacklog, error) { return nil, errDeLectura }
		}, func(r app.ReporteDelProyecto) error { return r.Backlog.Err }},
		{"sprints", func(f *fuentesDelReporte) {
			f.Sprints = func(context.Context, string) ([]app.ResumenDeSprint, error) { return nil, errDeLectura }
		}, func(r app.ReporteDelProyecto) error { return r.Sprints.Err }},
		{"sprints (los que volvieron al backlog)", func(f *fuentesDelReporte) {
			f.Anteriores = func(context.Context, string) (app.SprintsAnteriores, error) {
				return app.SprintsAnteriores{}, errDeLectura
			}
		}, func(r app.ReporteDelProyecto) error { return r.Sprints.Err }},
		{"esfuerzo (el total)", func(f *fuentesDelReporte) {
			f.Esfuerzo = func(context.Context, string, app.FiltroEsfuerzo) (app.ConsultaEsfuerzo, error) {
				return app.ConsultaEsfuerzo{}, errDeLectura
			}
		}, func(r app.ReporteDelProyecto) error { return r.Esfuerzo.Err }},
		{"esfuerzo (la comparacion)", func(f *fuentesDelReporte) {
			f.Comparacion = func(context.Context, string) (metricas.ComparacionProyecto, error) {
				return metricas.ComparacionProyecto{}, errDeLectura
			}
		}, func(r app.ReporteDelProyecto) error { return r.Esfuerzo.Err }},
		{"defectos (la lista)", func(f *fuentesDelReporte) {
			f.Defectos = func(context.Context, string, app.FiltroDefectos) ([]app.DefectoListado, error) {
				return nil, errDeLectura
			}
		}, func(r app.ReporteDelProyecto) error { return r.Defectos.Err }},
		{"el dashboard rompe el proyecto, los defectos y las metricas", func(f *fuentesDelReporte) {
			f.Dashboard = func(context.Context, string) (app.Dashboard, error) { return app.Dashboard{}, errDeLectura }
		}, func(r app.ReporteDelProyecto) error {
			if r.Proyecto.Err == nil || r.Defectos.Err == nil {
				return nil
			}
			return r.Metricas.Err
		}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			f := nuevasFuentesDelReporte()
			c.romper(f)

			r, err := f.reporte()

			if err != nil {
				t.Fatalf("Ejecutar: %v: el reporte se tiene que mostrar con la seccion en error", err)
			}
			if !errors.Is(c.error(r), errDeLectura) {
				t.Errorf("la seccion no informa el error de su lectura: %v", c.error(r))
			}
			// El resto de las secciones no se ve afectado: el backlog siempre es independiente de las demas, salvo cuando es el roto.
			if c.nombre != "backlog" && r.Backlog.Err != nil {
				t.Errorf("el backlog informa un error que no es suyo: %v", r.Backlog.Err)
			}
			if c.nombre == "backlog" && (r.Sprints.Err != nil || r.Esfuerzo.Err != nil || r.Defectos.Err != nil || r.Metricas.Err != nil) {
				t.Errorf("el error del backlog se filtro a otras secciones: %+v", r)
			}
		})
	}
}

// CL-033-7, RN-033-3: el error de la tarjeta de metricas del dashboard (ErrEstado) lo informa la seccion del proyecto.
func TestConsultarReporteDelProyecto_ErrorDelEstadoDelDashboard(t *testing.T) {
	f := nuevasFuentesDelReporte()
	f.Dashboard = func(context.Context, string) (app.Dashboard, error) {
		return app.Dashboard{ErrEstado: errDeLectura}, nil
	}

	r, err := f.reporte()

	if err != nil || !errors.Is(r.Proyecto.Err, errDeLectura) {
		t.Fatalf("Ejecutar = %v, %v: la seccion del proyecto tiene que informar el error del estado", r.Proyecto.Err, err)
	}
}

// RN-033-2: los sprints van por numero descendente, y un sprint cerrado dice cuantas historias volvieron al backlog (los abiertos, 0).
func TestConsultarReporteDelProyecto_VolvieronAlBacklogSoloEnLosCerrados(t *testing.T) {
	f := nuevasFuentesDelReporte()
	f.Sprints = func(context.Context, string) ([]app.ResumenDeSprint, error) {
		return []app.ResumenDeSprint{
			{Sprint: sprint.Sprint{ID: "s3", Numero: 3, Estado: sprint.EstadoAbierto}},
			{Sprint: sprint.Sprint{ID: "s2", Numero: 2, Estado: sprint.EstadoCerrado}},
			{Sprint: sprint.Sprint{ID: "s1", Numero: 1, Estado: sprint.EstadoCerrado}},
		}, nil
	}
	f.Anteriores = func(context.Context, string) (app.SprintsAnteriores, error) {
		return app.SprintsAnteriores{Sprints: []app.SprintAnterior{
			{Sprint: sprint.Sprint{ID: "s2"}, HistoriasNoCompletadas: 2},
			{Sprint: sprint.Sprint{ID: "s1"}, HistoriasNoCompletadas: 0},
		}}, nil
	}

	r, err := f.reporte()

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	obtenidos := []int{}
	for _, s := range r.Sprints.Sprints {
		obtenidos = append(obtenidos, s.Sprint.Numero*10+s.VolvieronAlBacklog)
	}
	if len(obtenidos) != 3 || obtenidos[0] != 30 || obtenidos[1] != 22 || obtenidos[2] != 10 {
		t.Errorf("sprint*10+volvieron = %v, se esperaba [30 22 10]", obtenidos)
	}
}
