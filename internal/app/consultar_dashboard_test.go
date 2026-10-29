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

// dashboardFalso simula las cinco lecturas del dashboard (spec US-030) sin base de datos y cuenta cuantas veces se llama a cada
// una (RN-030-11).
type dashboardFalso struct {
	proyecto proyecto.Proyecto
	items    []backlog.ItemBacklog
	sprints  []app.SprintConItems
	horas    []esfuerzo.RegistroEsfuerzo
	defectos []app.DefectoListado
	err      error

	lecturas map[string]int
}

func (f *dashboardFalso) contar(nombre string) {
	if f.lecturas == nil {
		f.lecturas = map[string]int{}
	}
	f.lecturas[nombre]++
}

func (f *dashboardFalso) BuscarProyecto(_ context.Context, id string) (proyecto.Proyecto, error) {
	f.contar("proyecto")
	if f.proyecto.ID != id {
		return proyecto.Proyecto{}, app.ErrProyectoInexistente
	}
	return f.proyecto, nil
}

func (f *dashboardFalso) ListarItems(context.Context, string) ([]backlog.ItemBacklog, error) {
	f.contar("items")
	return append([]backlog.ItemBacklog(nil), f.items...), f.err
}

func (f *dashboardFalso) ListarSprintsConItems(context.Context, string) ([]app.SprintConItems, error) {
	f.contar("sprints")
	return append([]app.SprintConItems(nil), f.sprints...), f.err
}

func (f *dashboardFalso) LeerEsfuerzo(context.Context, string) (app.LecturaEsfuerzo, error) {
	f.contar("esfuerzo")
	return app.LecturaEsfuerzo{Registros: append([]esfuerzo.RegistroEsfuerzo(nil), f.horas...)}, f.err
}

func (f *dashboardFalso) ListarDefectosDelProyecto(context.Context, string) ([]app.DefectoListado, error) {
	f.contar("defectos")
	return append([]app.DefectoListado(nil), f.defectos...), f.err
}

func dashboardDe(repo *dashboardFalso) (app.Dashboard, error) {
	return app.NuevoConsultarDashboard(repo, func() time.Time { return seisDeOctubre }).Ejecutar(context.Background(), "p1")
}

// sprintCerradoCon arma un sprint cerrado de las fechas dadas, con sus historias ya con el resultado de su fila.
func sprintCerradoCon(numero int, inicio, fin time.Time, historias ...app.HistoriaAsignada) app.SprintConItems {
	s := sprint.Sprint{ID: "s" + string(rune('0'+numero)), Numero: numero, ProyectoID: "p1", SprintGoal: "Objetivo", Estado: sprint.EstadoCerrado, FechaInicio: inicio, FechaFin: fin}
	return app.SprintConItems{Sprint: s, Items: historias}
}

func dia(mes time.Month, d int) time.Time { return time.Date(2026, mes, d, 0, 0, 0, 0, time.UTC) }

func registroDeHoras(fecha time.Time, horas float64) esfuerzo.RegistroEsfuerzo {
	return esfuerzo.RegistroEsfuerzo{ProyectoID: "p1", Fecha: fecha, Horas: horas, Actividad: "Trabajo"}
}

func defectoDelDashboard(severidad defecto.Severidad, estado defecto.Estado, deteccion, resolucion int) app.DefectoListado {
	return app.DefectoListado{Defecto: defecto.Defecto{Severidad: severidad, Estado: estado}, SprintDeteccion: deteccion, SprintResolucion: resolucion}
}

// US-030 / CA-030-1, RN-030-1 a RN-030-4: la tarjeta del sprint abierto trae sus Story Points, el porcentaje de historias
// completadas y las horas (estimadas por 1,5 h por Story Point, reales las de las fechas del sprint).
func TestConsultarDashboard_TarjetaDelSprintAbierto(t *testing.T) {
	abierto := sprintDe("s1", 1, "p1") // del 1 al 14 de octubre
	repo := &dashboardFalso{
		proyecto: proyectoDeEstado(),
		sprints: []app.SprintConItems{{Sprint: abierto, Items: []app.HistoriaAsignada{
			asignada(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado)),
			asignada(historiaDe("b", 2, backlog.PrioridadMust, 3, backlog.EstadoPendiente)),
			asignada(historiaDe("c", 3, backlog.PrioridadShould, 2, backlog.EstadoPendiente)),
		}}},
		horas: []esfuerzo.RegistroEsfuerzo{registroDeHoras(dia(time.October, 2), 3), registroDeHoras(dia(time.October, 14), 5)},
	}

	d, err := dashboardDe(repo)

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	r := d.Referencia
	if r == nil || r.Err != nil || r.Sprint.ID != "s1" {
		t.Fatalf("Referencia = %+v, se esperaba la tarjeta del sprint abierto s1 sin error", r)
	}
	if r.Planificados != 10 || r.Completados != 5 || r.PorcentajeHistorias != 33.3 {
		t.Errorf("sprint = %d planificados, %d completados, %v %%, se esperaba 10, 5 y 33,3", r.Planificados, r.Completados, r.PorcentajeHistorias)
	}
	if r.Horas.Estimadas != 15 || r.Horas.Reales != 8 || r.Horas.Desviacion != -7 {
		t.Errorf("horas = %+v, se esperaban 15 estimadas, 8 reales y -7 de desviacion", r.Horas)
	}
}

// US-030 / RN-030-1, CL-030-1, CL-030-2, CL-030-3: la referencia es el sprint abierto; si no hay, el cerrado de numero mas
// alto; sin sprints no hay tarjeta de sprint ni gráfico de porcentaje.
func TestConsultarDashboard_SprintDeReferencia(t *testing.T) {
	cerrado1 := sprintCerradoCon(1, dia(time.September, 1), dia(time.September, 14))
	cerrado2 := sprintCerradoCon(2, dia(time.September, 15), dia(time.September, 28))
	abierto := sprintDe("s3", 3, "p1")

	casos := []struct {
		nombre  string
		sprints []app.SprintConItems
		quiero  int
	}{
		{"con un sprint abierto y otros cerrados: el abierto", []app.SprintConItems{cerrado1, {Sprint: abierto}, cerrado2}, 3},
		{"con todos cerrados: el de numero mas alto", []app.SprintConItems{cerrado2, cerrado1}, 2},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := dashboardDe(&dashboardFalso{proyecto: proyectoDeEstado(), sprints: c.sprints})
			if err != nil || d.Referencia == nil || d.Referencia.Sprint.Numero != c.quiero {
				t.Fatalf("Referencia = %+v (%v), se esperaba el sprint %d", d.Referencia, err, c.quiero)
			}
			if d.Graficos.Completadas == nil || d.Graficos.Completadas.Sprint != c.quiero {
				t.Errorf("el gráfico de historias completadas = %+v, se esperaba el del sprint %d", d.Graficos.Completadas, c.quiero)
			}
		})
	}

	d, err := dashboardDe(&dashboardFalso{proyecto: proyectoDeEstado()})
	if err != nil || d.Referencia != nil || d.Graficos.Completadas != nil {
		t.Errorf("sin sprints: Referencia = %+v, gráfico = %+v (%v), se esperaba nada de los dos", d.Referencia, d.Graficos.Completadas, err)
	}
}

// US-030 / CL-030-4: sin horas registradas, las reales son 0 y la desviación es la de las estimadas en negativo (-100 %); sin
// Story Points planificados y con horas, el porcentaje de desviación no se puede calcular.
func TestConsultarDashboard_HorasLimite(t *testing.T) {
	conPuntos := []app.SprintConItems{{Sprint: sprintDe("s1", 1, "p1"), Items: []app.HistoriaAsignada{asignada(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoPendiente))}}}
	d, err := dashboardDe(&dashboardFalso{proyecto: proyectoDeEstado(), sprints: conPuntos})
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	h := d.Referencia.Horas
	if p, calculable := h.DesviacionPorcentual(); h.Estimadas != 7.5 || h.Reales != 0 || h.Desviacion != -7.5 || !calculable || p != -100 {
		t.Errorf("sin horas: %+v (%v %%, calculable %v), se esperaban 7,5 estimadas, 0 reales, -7,5 y -100 %%", h, p, calculable)
	}

	sinPuntos := []app.SprintConItems{{Sprint: sprintDe("s1", 1, "p1")}}
	d, err = dashboardDe(&dashboardFalso{proyecto: proyectoDeEstado(), sprints: sinPuntos, horas: []esfuerzo.RegistroEsfuerzo{registroDeHoras(dia(time.October, 3), 2)}})
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if _, calculable := d.Referencia.Horas.DesviacionPorcentual(); calculable || d.Referencia.Horas.Reales != 2 {
		t.Errorf("sin Story Points: %+v, se esperaban 2 horas reales y porcentaje no calculable", d.Referencia.Horas)
	}
}

// US-030 / RN-030-4, CL-030-5: una fecha entre dos sprints, o antes del primero, no entra en las horas de ningun sprint.
func TestConsultarDashboard_HorasFueraDeTodoSprint(t *testing.T) {
	repo := &dashboardFalso{
		proyecto: proyectoDeEstado(),
		sprints:  []app.SprintConItems{{Sprint: sprintDe("s1", 1, "p1")}}, // del 1 al 14 de octubre
		horas:    []esfuerzo.RegistroEsfuerzo{registroDeHoras(dia(time.September, 30), 4), registroDeHoras(dia(time.October, 15), 6), registroDeHoras(dia(time.October, 1), 1)},
	}

	d, err := dashboardDe(repo)

	if err != nil || d.Referencia.Horas.Reales != 1 {
		t.Errorf("horas reales = %v (%v), se esperaba solo la hora del primer dia del sprint", d.Referencia.Horas.Reales, err)
	}
}

// US-030 / CA-030-2, RN-030-5, RN-030-8, CL-030-6: con dos sprints cerrados de 8 y 5 Story Points completados la velocidad es
// 6,5 y la tabla trae una fila por sprint cerrado, del más reciente al más antiguo.
func TestConsultarDashboard_VelocidadYSprintsCerrados(t *testing.T) {
	s1 := sprintCerradoCon(1, dia(time.September, 1), dia(time.September, 14),
		cerrada(historiaDe("a", 1, backlog.PrioridadMust, 8, backlog.EstadoCompletado), sprint.ResultadoCompletado),
		cerrada(historiaDe("b", 2, backlog.PrioridadMust, 2, backlog.EstadoPendiente), sprint.ResultadoNoCompletado))
	s2 := sprintCerradoCon(2, dia(time.September, 15), dia(time.September, 28),
		cerrada(historiaDe("c", 3, backlog.PrioridadMust, 5, backlog.EstadoCompletado), sprint.ResultadoCompletado))

	d, err := dashboardDe(&dashboardFalso{proyecto: proyectoDeEstado(), sprints: []app.SprintConItems{s1, s2}})

	if err != nil || d.Velocidad.Err != nil {
		t.Fatalf("Ejecutar: %v / %v", err, d.Velocidad.Err)
	}
	if d.Velocidad.Velocidad != 6.5 {
		t.Errorf("velocidad = %v, se esperaba 6,5", d.Velocidad.Velocidad)
	}
	filas := d.Velocidad.Cerrados
	if len(filas) != 2 || filas[0].Sprint.Numero != 2 || filas[1].Sprint.Numero != 1 {
		t.Fatalf("filas = %+v, se esperaban el sprint 2 y despues el 1", filas)
	}
	if filas[1].Planificados != 10 || filas[1].Completados != 8 || filas[0].Planificados != 5 || filas[0].Completados != 5 {
		t.Errorf("Story Points de las filas = %+v, se esperaban 10/8 en el sprint 1 y 5/5 en el 2", filas)
	}

	solo, _ := dashboardDe(&dashboardFalso{proyecto: proyectoDeEstado(), sprints: []app.SprintConItems{s1}})
	if solo.Velocidad.Velocidad != 8 {
		t.Errorf("con un solo sprint cerrado la velocidad = %v, se esperaba la de ese sprint (8)", solo.Velocidad.Velocidad)
	}
}

// US-030 / CL-030-12 (DEF-007): una historia que no se completó en el sprint 1 y se completó después sigue sin contar en el
// sprint 1, porque manda el resultado de su fila.
func TestConsultarDashboard_ElResultadoDeLaFilaManda(t *testing.T) {
	s1 := sprintCerradoCon(1, dia(time.September, 1), dia(time.September, 14),
		cerrada(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado), sprint.ResultadoNoCompletado)) // hoy completada, pero no en el sprint 1

	d, err := dashboardDe(&dashboardFalso{proyecto: proyectoDeEstado(), sprints: []app.SprintConItems{s1}})

	if err != nil || d.Velocidad.Cerrados[0].Completados != 0 || d.Velocidad.Velocidad != 0 {
		t.Errorf("sprint 1 = %+v, velocidad %v (%v): no tiene que contar la historia que no se completó en él", d.Velocidad.Cerrados, d.Velocidad.Velocidad, err)
	}
}

// US-030 / CA-030-3, RN-030-6, RN-030-7, CL-030-7, CL-030-8: la tarjeta de defectos cuenta todos (también los descartados) por
// severidad y por estado; en las columnas por sprint un defecto cuenta como detectado en su sprint y como resuelto en el de su
// resolución, y los descartados no cuentan en ninguna.
func TestConsultarDashboard_Defectos(t *testing.T) {
	cerrados := []app.SprintConItems{
		sprintCerradoCon(1, dia(time.August, 1), dia(time.August, 14)),
		sprintCerradoCon(2, dia(time.August, 15), dia(time.August, 28)),
		sprintCerradoCon(3, dia(time.August, 29), dia(time.September, 11)),
		sprintCerradoCon(4, dia(time.September, 12), dia(time.September, 25)),
	}
	repo := &dashboardFalso{proyecto: proyectoDeEstado(), sprints: cerrados, defectos: []app.DefectoListado{
		defectoDelDashboard(defecto.SeveridadCritica, defecto.EstadoResuelto, 2, 4),  // detectado en el 2, resuelto en el 4
		defectoDelDashboard(defecto.SeveridadAlta, defecto.EstadoResuelto, 3, 3),     // detectado y resuelto en el mismo sprint
		defectoDelDashboard(defecto.SeveridadAlta, defecto.EstadoAbierto, 3, 0),      // sin resolver
		defectoDelDashboard(defecto.SeveridadBaja, defecto.EstadoEnCorreccion, 4, 0), //
		defectoDelDashboard(defecto.SeveridadMedia, defecto.EstadoDescartado, 2, 0),  // descartado: no cuenta por sprint
		defectoDelDashboard(defecto.SeveridadBaja, defecto.EstadoDescartado, 1, 0),   //
	}}

	d, err := dashboardDe(repo)

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	td := d.Defectos
	if td.Total != 6 || td.PorSeveridad != (app.ConteoPorSeveridad{Critica: 1, Alta: 2, Media: 1, Baja: 2}) ||
		td.PorEstado != (app.ConteoPorEstado{Abierto: 1, EnCorreccion: 1, Resuelto: 2, Descartado: 2}) {
		t.Errorf("tarjeta de defectos = %+v, se esperaban 6 (1 critica, 2 altas, 1 media, 2 bajas; 1 abierto, 1 en correccion, 2 resueltos, 2 descartados)", td)
	}
	porSprint := map[int][2]int{}
	for _, fila := range d.Velocidad.Cerrados {
		porSprint[fila.Sprint.Numero] = [2]int{fila.Defectos.Detectados, fila.Defectos.Resueltos}
	}
	esperado := map[int][2]int{1: {0, 0}, 2: {1, 0}, 3: {2, 1}, 4: {1, 1}}
	for n, quiero := range esperado {
		if porSprint[n] != quiero {
			t.Errorf("sprint %d: detectados y resueltos = %v, se esperaba %v", n, porSprint[n], quiero)
		}
	}
}

// US-030 / CA-030-4, CL-030-3, CL-030-9: un proyecto nuevo, sin sprints, horas ni defectos, lo muestra todo en 0 y sin error.
func TestConsultarDashboard_ProyectoNuevo(t *testing.T) {
	d, err := dashboardDe(&dashboardFalso{proyecto: proyectoDeEstado()})

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if d.Referencia != nil || d.Velocidad.Velocidad != 0 || len(d.Velocidad.Cerrados) != 0 || d.Velocidad.Err != nil {
		t.Errorf("velocidad = %+v y referencia = %+v, se esperaba todo vacío", d.Velocidad, d.Referencia)
	}
	if d.Defectos.Total != 0 || d.Defectos.PorSeveridad != (app.ConteoPorSeveridad{}) || d.Defectos.PorEstado != (app.ConteoPorEstado{}) {
		t.Errorf("defectos = %+v, se esperaba todo en 0", d.Defectos)
	}
	if len(d.Graficos.StoryPoints) != 0 || d.Graficos.Completadas != nil || d.Graficos.ErrStoryPoints != nil {
		t.Errorf("gráficos = %+v, se esperaban vacíos y sin error", d.Graficos)
	}
	if d.Estado.Estado != proyecto.EnCurso || d.ErrEstado != nil {
		t.Errorf("estado = %q (%v), se esperaba el de US-004: en curso", d.Estado.Estado, d.ErrEstado)
	}
}

// US-030 / RN-030-14, CL-030-15: el gráfico de Story Points tiene un grupo por sprint cerrado en orden cronológico, y la escala es
// una para todo el gráfico: la barra mas alta mide 100 y las demás, valor / máximo * 100.
func TestConsultarDashboard_GraficoDeStoryPoints(t *testing.T) {
	s1 := sprintCerradoCon(1, dia(time.September, 1), dia(time.September, 14),
		cerrada(historiaDe("a", 1, backlog.PrioridadMust, 8, backlog.EstadoCompletado), sprint.ResultadoCompletado),
		cerrada(historiaDe("b", 2, backlog.PrioridadMust, 2, backlog.EstadoPendiente), sprint.ResultadoNoCompletado))
	s2 := sprintCerradoCon(2, dia(time.September, 15), dia(time.September, 28),
		cerrada(historiaDe("c", 3, backlog.PrioridadMust, 5, backlog.EstadoCompletado), sprint.ResultadoCompletado))

	d, err := dashboardDe(&dashboardFalso{proyecto: proyectoDeEstado(), sprints: []app.SprintConItems{s2, s1}})

	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	g := d.Graficos.StoryPoints
	if len(g) != 2 || g[0].Sprint != 1 || g[1].Sprint != 2 {
		t.Fatalf("grupos = %+v, se esperaban el sprint 1 y despues el 2 (orden cronológico)", g)
	}
	if g[0] != (app.BarrasDeSprint{Sprint: 1, Planificados: 10, Completados: 8, AlturaPlanificados: 100, AlturaCompletados: 80}) {
		t.Errorf("sprint 1 = %+v, se esperaban alturas 100 y 80", g[0])
	}
	if g[1] != (app.BarrasDeSprint{Sprint: 2, Planificados: 5, Completados: 5, AlturaPlanificados: 50, AlturaCompletados: 50}) {
		t.Errorf("sprint 2 = %+v, se esperaban alturas 50 y 50", g[1])
	}
}

// US-030 / CL-030-14: sprints cerrados sin Story Points: las barras miden 0 y no hay división por cero.
func TestConsultarDashboard_GraficoSinStoryPoints(t *testing.T) {
	vacio := sprintCerradoCon(1, dia(time.September, 1), dia(time.September, 14))

	d, err := dashboardDe(&dashboardFalso{proyecto: proyectoDeEstado(), sprints: []app.SprintConItems{vacio}})

	if err != nil || len(d.Graficos.StoryPoints) != 1 {
		t.Fatalf("gráfico = %+v (%v), se esperaba un grupo", d.Graficos.StoryPoints, err)
	}
	if b := d.Graficos.StoryPoints[0]; b.AlturaPlanificados != 0 || b.AlturaCompletados != 0 {
		t.Errorf("alturas = %v y %v, se esperaban 0 y 0", b.AlturaPlanificados, b.AlturaCompletados)
	}
}

// US-030 / CL-030-16, RN-030-15: el gráfico de porcentaje es el de la tarjeta del sprint de referencia.
func TestConsultarDashboard_GraficoDePorcentaje(t *testing.T) {
	abierto := sprintDe("s1", 1, "p1")
	repo := &dashboardFalso{proyecto: proyectoDeEstado(), sprints: []app.SprintConItems{{Sprint: abierto, Items: []app.HistoriaAsignada{
		asignada(historiaDe("a", 1, backlog.PrioridadMust, 5, backlog.EstadoCompletado)),
		asignada(historiaDe("b", 2, backlog.PrioridadMust, 5, backlog.EstadoCompletado)),
		asignada(historiaDe("c", 3, backlog.PrioridadMust, 5, backlog.EstadoPendiente)),
		asignada(historiaDe("d", 4, backlog.PrioridadMust, 5, backlog.EstadoPendiente)),
	}}}}

	d, err := dashboardDe(repo)

	if err != nil || d.Graficos.Completadas == nil || d.Graficos.Completadas.Porcentaje != 50 || d.Graficos.Completadas.Sprint != 1 {
		t.Errorf("gráfico = %+v (%v), se esperaba el 50 %% del sprint 1", d.Graficos.Completadas, err)
	}
}

// US-030 / RN-030-10, CL-030-11: un proyecto que no existe se rechaza con ErrProyectoInexistente, envuelto con el ID, sin leer
// nada más.
func TestConsultarDashboard_ProyectoInexistente(t *testing.T) {
	repo := &dashboardFalso{proyecto: proyectoDeEstado()}

	_, err := app.NuevoConsultarDashboard(repo, func() time.Time { return seisDeOctubre }).Ejecutar(context.Background(), "otro")

	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Fatalf("error = %v, se esperaba ErrProyectoInexistente", err)
	}
	if n := repo.lecturas["items"] + repo.lecturas["sprints"] + repo.lecturas["esfuerzo"] + repo.lecturas["defectos"]; n != 0 {
		t.Errorf("lecturas despues de un proyecto inexistente = %v, no se esperaba ninguna", repo.lecturas)
	}
}

// US-030 / RN-030-11: una sola lectura de cada tipo, no una por sprint ni por historia.
func TestConsultarDashboard_UnaLecturaDeCadaTipo(t *testing.T) {
	repo := &dashboardFalso{proyecto: proyectoDeEstado(), sprints: []app.SprintConItems{
		sprintCerradoCon(1, dia(time.September, 1), dia(time.September, 14)), sprintCerradoCon(2, dia(time.September, 15), dia(time.September, 28)), {Sprint: sprintDe("s3", 3, "p1")},
	}}

	if _, err := dashboardDe(repo); err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	for _, tipo := range []string{"proyecto", "items", "sprints", "esfuerzo", "defectos"} {
		if repo.lecturas[tipo] != 1 {
			t.Errorf("lecturas de %s = %d, se esperaba 1 (todas: %v)", tipo, repo.lecturas[tipo], repo.lecturas)
		}
	}
}

// US-030 / sección 7: una falla de la persistencia se devuelve envuelta, sin ocultarla (500 en la pantalla).
func TestConsultarDashboard_FallaDeLaPersistencia(t *testing.T) {
	caido := errors.New("la base no responde")

	_, err := dashboardDe(&dashboardFalso{proyecto: proyectoDeEstado(), err: caido})

	if !errors.Is(err, caido) {
		t.Errorf("error = %v, se esperaba el de la persistencia envuelto", err)
	}
}
