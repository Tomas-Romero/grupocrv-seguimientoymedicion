package postgres_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// sprintParaCerrarEnPostgres es el escenario de asignaciones con dos historias en el
// sprint abierto: la primera completada y la segunda no.
func sprintParaCerrarEnPostgres(ctx context.Context, t *testing.T) (escenarioAsignaciones, string) {
	t.Helper()
	e := asignadoEnPostgres(ctx, t)
	otra := e.itemParaSprint(ctx, t, e.proyecto, "Listar proyectos", true, 1)
	if err := e.repo.AsignarItemASprint(ctx, otra, e.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	if _, err := e.repo.GuardarCambioDeEstado(ctx, e.item, e.sprint, completarEn(marcadoEn)); err != nil {
		t.Fatalf("marcar: %v", err)
	}
	return e, otra
}

func estadoDelSprint(ctx context.Context, t *testing.T, e escenarioAsignaciones) string {
	t.Helper()
	var estado string
	if err := e.pool.QueryRow(ctx, `SELECT estado FROM sprints WHERE id = $1`, e.sprint).Scan(&estado); err != nil {
		t.Fatalf("leer el estado del sprint: %v", err)
	}
	return estado
}

// US-011 / RN-011-8: las historias del sprint son sus filas pendientes, en orden de
// asignacion, con si cada una esta completada. Las filas de otros sprints no cuentan.
func TestHistoriasDelSprint(t *testing.T) {
	ctx := contexto(t)
	e, otra := sprintParaCerrarEnPostgres(ctx, t)

	historias, err := e.repo.HistoriasDelSprint(ctx, e.sprint)
	if err != nil {
		t.Fatalf("HistoriasDelSprint: %v", err)
	}

	esperadas := []sprint.HistoriaDelCierre{{ItemID: e.item, Completada: true}, {ItemID: otra, Completada: false}}
	if !slices.Equal(historias, esperadas) {
		t.Errorf("historias = %+v, se esperaban %+v", historias, esperadas)
	}
	if historias, err := e.repo.HistoriasDelSprint(ctx, "no-es-un-uuid"); err != nil || len(historias) != 0 {
		t.Errorf("con un ID mal formado = (%+v, %v), se esperaba (ninguna, nil)", historias, err)
	}
}

// US-011 / CA-011-1, RN-011-5, RN-011-6, RN-011-7, CL-011-4: cerrar deja el sprint
// cerrado y cada fila con su resultado, sin borrar ninguna. La completada queda como
// esta, con su fecha; las no completadas (tambien una en progreso) pasan a pendiente y
// quedan en el backlog.
func TestGuardarCierre(t *testing.T) {
	ctx := contexto(t)
	e, otra := sprintParaCerrarEnPostgres(ctx, t)
	enProgreso := e.itemParaSprint(ctx, t, e.proyecto, "Editar proyectos", true, 1)
	if err := e.repo.AsignarItemASprint(ctx, enProgreso, e.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	if _, err := e.pool.Exec(ctx, `UPDATE items_backlog SET estado = 'en_progreso' WHERE id = $1`, enProgreso); err != nil {
		t.Fatalf("poner la historia en progreso: %v", err)
	}

	resultados, err := e.repo.GuardarCierre(ctx, e.sprint)
	if err != nil {
		t.Fatalf("GuardarCierre: %v", err)
	}

	esperados := []sprint.ResultadoDelCierre{
		{ItemID: e.item, Resultado: sprint.ResultadoCompletado},
		{ItemID: otra, Resultado: sprint.ResultadoNoCompletado},
		{ItemID: enProgreso, Resultado: sprint.ResultadoNoCompletado},
	}
	if !slices.Equal(resultados, esperados) {
		t.Errorf("resultados = %+v, se esperaban %+v", resultados, esperados)
	}
	if estado := estadoDelSprint(ctx, t, e); estado != "cerrado" {
		t.Errorf("estado del sprint = %q, se esperaba cerrado", estado)
	}
	if filas := filasDeLaRelacion(ctx, t, e.pool, e.item); !slices.Equal(filas, []string{"completado"}) {
		t.Errorf("filas de la completada = %v, se esperaba [completado]", filas)
	}
	if completada := leerEstado(ctx, t, e, e.item); completada.estado != "completado" ||
		completada.completadoEn == nil || !completada.completadoEn.Equal(marcadoEn) {
		t.Errorf("la completada cambio: %+v", completada)
	}
	for _, id := range []string{otra, enProgreso} {
		if filas := filasDeLaRelacion(ctx, t, e.pool, id); !slices.Equal(filas, []string{"no_completado"}) {
			t.Errorf("filas de %s = %v, se esperaba [no_completado]", id, filas)
		}
		if devuelta := leerEstado(ctx, t, e, id); devuelta.estado != "pendiente" || devuelta.completadoEn != nil {
			t.Errorf("la no completada %s quedo %+v, se esperaba pendiente y sin fecha", id, devuelta)
		}
		if en, _ := e.repo.ItemEnSprintAbierto(ctx, id); en {
			t.Errorf("la no completada %s sigue en un sprint abierto", id)
		}
	}
}

// US-011 / CA-011-3: el sprint cerrado conserva lo planificado y lo completado. Despues del
// cierre, los Story Points planificados incluyen la historia no completada (5 + 5), los
// completados son solo los de la completada (5) y la velocidad sale de lo completado.
func TestGuardarCierre_ConservaLoPlanificado(t *testing.T) {
	ctx := contexto(t)
	e, _ := sprintParaCerrarEnPostgres(ctx, t)
	if _, err := e.repo.GuardarCierre(ctx, e.sprint); err != nil {
		t.Fatalf("GuardarCierre: %v", err)
	}

	anteriores, err := app.NuevoConsultarSprintsAnteriores(e.repo).Ejecutar(ctx, e.proyecto)
	if err != nil {
		t.Fatalf("ConsultarSprintsAnteriores: %v", err)
	}

	if len(anteriores.Sprints) != 1 {
		t.Fatalf("sprints anteriores = %+v, se esperaba solo el cerrado", anteriores.Sprints)
	}
	s := anteriores.Sprints[0]
	if s.Planificados != 10 || s.Completados != 5 || s.HistoriasCompletadas != 1 || s.HistoriasNoCompletadas != 1 {
		t.Errorf("sprint cerrado = %+v, se esperaban 10 planificados, 5 completados y una historia de cada", s)
	}
	if anteriores.Velocidad != 5 {
		t.Errorf("velocidad = %v, se esperaba 5", anteriores.Velocidad)
	}
}

// US-011 / CA-011-4, RN-011-4, CL-011-1: un sprint sin historias se cierra.
func TestGuardarCierre_SinHistorias(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)

	resultados, err := e.repo.GuardarCierre(ctx, e.sprint)
	if err != nil || len(resultados) != 0 {
		t.Fatalf("GuardarCierre = (%+v, %v), se esperaba (ninguno, nil)", resultados, err)
	}
	if estado := estadoDelSprint(ctx, t, e); estado != "cerrado" {
		t.Errorf("estado del sprint = %q, se esperaba cerrado", estado)
	}
}

// US-011 / CA-011-5, CA-011-6, RN-011-2, CL-011-6, CL-011-7: un sprint ya cerrado se
// rechaza con ErrSprintCerrado, y uno que no existe (o con un ID mal formado), con
// ErrSprintInexistente. No cambia nada.
func TestGuardarCierre_Rechazos(t *testing.T) {
	ctx := contexto(t)
	e, otra := sprintParaCerrarEnPostgres(ctx, t)
	if _, err := e.repo.GuardarCierre(ctx, e.sprint); err != nil {
		t.Fatalf("primer cierre: %v", err)
	}

	if _, err := e.repo.GuardarCierre(ctx, e.sprint); !errors.Is(err, sprint.ErrSprintCerrado) {
		t.Errorf("segundo cierre: error = %v, se esperaba ErrSprintCerrado", err)
	}
	for _, id := range []string{idDeAsignacionSinFila, "no-es-un-uuid"} {
		if _, err := e.repo.GuardarCierre(ctx, id); !errors.Is(err, app.ErrSprintInexistente) {
			t.Errorf("GuardarCierre(%q) = %v, se esperaba ErrSprintInexistente", id, err)
		}
	}
	if filas := filasDeLaRelacion(ctx, t, e.pool, otra); !slices.Equal(filas, []string{"no_completado"}) {
		t.Errorf("filas = %v, el segundo cierre las cambio", filas)
	}
}

// US-011 / CA-011-2, RN-011-10, CL-011-8: despues de cerrar se crea el sprint siguiente
// y la historia devuelta se asigna a el: queda con dos filas, no_completado en el
// cerrado y pendiente en el nuevo.
func TestGuardarCierre_HistoriaDevueltaSeAsignaAlSiguiente(t *testing.T) {
	ctx := contexto(t)
	e, otra := sprintParaCerrarEnPostgres(ctx, t)
	if _, err := e.repo.GuardarCierre(ctx, e.sprint); err != nil {
		t.Fatalf("GuardarCierre: %v", err)
	}
	siguiente, err := e.sprints.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 20), diaSprint(11, 2)))
	if err != nil {
		t.Fatalf("RegistrarSprint del siguiente: %v", err)
	}

	if err := e.repo.AsignarItemASprint(ctx, otra, siguiente.ID); err != nil {
		t.Fatalf("asignar la devuelta al siguiente: %v", err)
	}

	if filas := filasDeLaRelacion(ctx, t, e.pool, otra); !slices.Equal(filas, []string{"no_completado", "pendiente"}) {
		t.Errorf("filas de la devuelta = %v, se esperaba [no_completado pendiente]", filas)
	}
}

// US-011 / RN-011-8, CL-011-11: si la transaccion falla a mitad del cierre, no cambia
// nada: el sprint sigue abierto, las filas pendientes y las historias como estaban. Para
// forzar la falla, un trigger de la base de prueba rechaza el ultimo paso (cerrar el
// sprint), despues de que ya se guardaron los resultados.
func TestGuardarCierre_FallaAMitad(t *testing.T) {
	ctx := contexto(t)
	e, otra := sprintParaCerrarEnPostgres(ctx, t)
	const trigger = `
CREATE FUNCTION rechazar_cierre() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'falla forzada por el test'; END $$;
CREATE TRIGGER rechazar_cierre BEFORE UPDATE ON sprints FOR EACH ROW EXECUTE FUNCTION rechazar_cierre();`
	if _, err := e.pool.Exec(ctx, trigger); err != nil {
		t.Fatalf("crear el trigger: %v", err)
	}

	if _, err := e.repo.GuardarCierre(ctx, e.sprint); err == nil {
		t.Fatal("GuardarCierre termino bien, se esperaba la falla forzada")
	}

	if estado := estadoDelSprint(ctx, t, e); estado != "abierto" {
		t.Errorf("estado del sprint = %q, se esperaba abierto", estado)
	}
	for _, id := range []string{e.item, otra} {
		if filas := filasDeLaRelacion(ctx, t, e.pool, id); !slices.Equal(filas, []string{"pendiente"}) {
			t.Errorf("filas de %s = %v, se esperaba [pendiente]", id, filas)
		}
	}
	if completada := leerEstado(ctx, t, e, e.item); completada.estado != "completado" {
		t.Errorf("la completada quedo %+v", completada)
	}
}

// operacionSobreElSprint es algo que el test lanza contra el sprint bloqueado, y su
// resultado.
type operacionSobreElSprint func() error

// enOrdenConElSprintBloqueado lanza las operaciones con la fila del sprint bloqueada por
// una transaccion del test. Cada una arranca recien cuando la anterior ya espera el
// candado (se ve en pg_stat_activity), asi el orden en que lo obtienen es fijo. Despues
// suelta el candado y devuelve el error de cada una, en el mismo orden.
func enOrdenConElSprintBloqueado(ctx context.Context, t *testing.T, e escenarioAsignaciones, operaciones ...operacionSobreElSprint) []error {
	t.Helper()
	candado, err := e.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir la transaccion del candado: %v", err)
	}
	defer func() { _ = candado.Rollback(ctx) }()
	if _, err := candado.Exec(ctx, `SELECT 1 FROM sprints WHERE id = $1 FOR NO KEY UPDATE`, e.sprint); err != nil {
		t.Fatalf("tomar el candado del sprint: %v", err)
	}

	canales := make([]chan error, len(operaciones))
	for i, op := range operaciones {
		canales[i] = make(chan error, 1)
		go func() { canales[i] <- op() }()
		if bloqueadas := esperarBloqueadas(ctx, t, e.pool, i+1); bloqueadas < i+1 {
			t.Fatalf("solo %d de %d operaciones esperaron el candado del sprint", bloqueadas, i+1)
		}
	}
	if err := candado.Rollback(ctx); err != nil {
		t.Fatalf("soltar el candado: %v", err)
	}

	errs := make([]error, len(operaciones))
	for i, c := range canales {
		errs[i] = <-c
	}
	return errs
}

// US-011 / RN-011-9, CL-011-9: un marcado (US-010) que obtiene el candado antes que el
// cierre cuenta: la historia queda completada en el sprint cerrado.
func TestGuardarCierre_MarcadoAntesDelCierre(t *testing.T) {
	ctx := contexto(t)
	e, otra := sprintParaCerrarEnPostgres(ctx, t)

	errs := enOrdenConElSprintBloqueado(ctx, t, e,
		func() error {
			_, err := e.repo.GuardarCambioDeEstado(ctx, otra, e.sprint, completarEn(segundaVez))
			return err
		},
		func() error { _, err := e.repo.GuardarCierre(ctx, e.sprint); return err },
	)

	if errs[0] != nil || errs[1] != nil {
		t.Fatalf("errores = %v, se esperaba que terminaran bien el marcado y el cierre", errs)
	}
	if filas := filasDeLaRelacion(ctx, t, e.pool, otra); !slices.Equal(filas, []string{"completado"}) {
		t.Errorf("filas = %v, se esperaba [completado]", filas)
	}
}

// US-011 / RN-011-9, CL-011-9, CL-010-9: un marcado que llega despues del cierre ve el
// sprint cerrado y se rechaza con ErrSprintCerrado; la historia queda no completada y
// en el backlog.
func TestGuardarCierre_MarcadoDespuesDelCierre(t *testing.T) {
	ctx := contexto(t)
	e, otra := sprintParaCerrarEnPostgres(ctx, t)

	errs := enOrdenConElSprintBloqueado(ctx, t, e,
		func() error { _, err := e.repo.GuardarCierre(ctx, e.sprint); return err },
		func() error {
			_, err := e.repo.GuardarCambioDeEstado(ctx, otra, e.sprint, completarEn(segundaVez))
			return err
		},
	)

	if errs[0] != nil {
		t.Fatalf("GuardarCierre: %v", errs[0])
	}
	if !errors.Is(errs[1], sprint.ErrSprintCerrado) {
		t.Errorf("marcado: error = %v, se esperaba ErrSprintCerrado", errs[1])
	}
	if devuelta := leerEstado(ctx, t, e, otra); devuelta.estado != "pendiente" {
		t.Errorf("la historia quedo %+v, se esperaba pendiente", devuelta)
	}
}

// US-011 / RN-011-9, CL-011-10: dos cierres simultaneos del mismo sprint: el primero
// termina bien y el segundo se rechaza con ErrSprintCerrado.
func TestGuardarCierre_DosCierres(t *testing.T) {
	ctx := contexto(t)
	e, _ := sprintParaCerrarEnPostgres(ctx, t)
	cerrar := func() error { _, err := e.repo.GuardarCierre(ctx, e.sprint); return err }

	errs := enOrdenConElSprintBloqueado(ctx, t, e, cerrar, cerrar)

	if errs[0] != nil {
		t.Errorf("primer cierre: %v", errs[0])
	}
	if !errors.Is(errs[1], sprint.ErrSprintCerrado) {
		t.Errorf("segundo cierre: error = %v, se esperaba ErrSprintCerrado", errs[1])
	}
}

// US-011 / RN-011-9: una asignacion (US-009, FOR SHARE) que obtiene el candado antes que
// el cierre cuenta: la historia entra al sprint y el cierre la registra como no
// completada.
func TestGuardarCierre_AsignacionAntesDelCierre(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)

	errs := enOrdenConElSprintBloqueado(ctx, t, e,
		func() error { return e.repo.AsignarItemASprint(ctx, e.item, e.sprint) },
		func() error { _, err := e.repo.GuardarCierre(ctx, e.sprint); return err },
	)

	if errs[0] != nil || errs[1] != nil {
		t.Fatalf("errores = %v, se esperaba que terminaran bien la asignacion y el cierre", errs)
	}
	if filas := filasDeLaRelacion(ctx, t, e.pool, e.item); !slices.Equal(filas, []string{"no_completado"}) {
		t.Errorf("filas = %v, se esperaba [no_completado]", filas)
	}
}

// El cierre no toca a las historias del backlog que no estaban en el sprint.
func TestGuardarCierre_NoTocaElBacklog(t *testing.T) {
	ctx := contexto(t)
	e, _ := sprintParaCerrarEnPostgres(ctx, t)
	fuera := e.itemParaSprint(ctx, t, e.proyecto, "Borrar proyectos", true, 1)
	if _, err := e.pool.Exec(ctx, `UPDATE items_backlog SET estado = 'en_progreso' WHERE id = $1`, fuera); err != nil {
		t.Fatalf("poner la historia en progreso: %v", err)
	}

	if _, err := e.repo.GuardarCierre(ctx, e.sprint); err != nil {
		t.Fatalf("GuardarCierre: %v", err)
	}

	if g := leerEstado(ctx, t, e, fuera); g.estado != string(backlog.EstadoEnProgreso) {
		t.Errorf("la historia fuera del sprint quedo %q, se esperaba en_progreso", g.estado)
	}
	if n := dbprueba.ContarFilas(ctx, t, e.pool, "sprint_items"); n != 2 {
		t.Errorf("sprint_items tiene %d filas, se esperaban 2: el cierre no borra ni agrega", n)
	}
}
