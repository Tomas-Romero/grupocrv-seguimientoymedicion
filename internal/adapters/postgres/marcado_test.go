package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Fechas de completado de estos tests.
var (
	marcadoEn  = time.Date(2026, 10, 8, 15, 0, 0, 0, time.UTC)
	segundaVez = time.Date(2026, 10, 9, 9, 0, 0, 0, time.UTC)
)

func completarEn(en time.Time) app.CambioDeEstado {
	return func(item backlog.ItemBacklog) backlog.ItemBacklog { return item.Completar(en) }
}

// estadoGuardado es el estado y la fecha de completado de un item, leidos de la tabla.
type estadoGuardado struct {
	estado       string
	completadoEn *time.Time
}

func leerEstado(ctx context.Context, t *testing.T, e escenarioAsignaciones, itemID string) estadoGuardado {
	t.Helper()
	var g estadoGuardado
	if err := e.pool.QueryRow(ctx, `SELECT estado, completado_en FROM items_backlog WHERE id = $1`, itemID).Scan(&g.estado, &g.completadoEn); err != nil {
		t.Fatalf("leer el estado del item %s: %v", itemID, err)
	}
	return g
}

// asignadoEnPostgres devuelve el escenario con el item ya asignado al sprint abierto.
func asignadoEnPostgres(ctx context.Context, t *testing.T) escenarioAsignaciones {
	t.Helper()
	e := nuevoEscenarioAsignaciones(ctx, t)
	if err := e.repo.AsignarItemASprint(ctx, e.item, e.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	return e
}

// US-010 / RN-010-2, CL-010-5: el item esta en el sprint si tiene una fila pendiente en
// el. Una fila de un sprint cerrado no cuenta, y un ID mal formado no esta en ningun
// sprint (no es un error).
func TestItemEnSprint(t *testing.T) {
	ctx := contexto(t)
	e := asignadoEnPostgres(ctx, t)
	sinAsignar := e.itemParaSprint(ctx, t, e.proyecto, "Listar proyectos", true, 1)

	casos := []struct {
		nombre           string
		itemID, sprintID string
		esperado         bool
	}{
		{"asignado", e.item, e.sprint, true},
		{"sin asignar", sinAsignar, e.sprint, false},
		{"item inexistente", idDeAsignacionSinFila, e.sprint, false},
		{"item mal formado", "no-es-un-uuid", e.sprint, false},
		{"sprint mal formado", e.item, "no-es-un-uuid", false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			en, err := e.repo.ItemEnSprint(ctx, c.itemID, c.sprintID)
			if err != nil || en != c.esperado {
				t.Errorf("ItemEnSprint = (%t, %v), se esperaba (%t, nil)", en, err, c.esperado)
			}
		})
	}

	if _, err := e.pool.Exec(ctx, `UPDATE sprint_items SET resultado = 'no_completado' WHERE item_id = $1`, e.item); err != nil {
		t.Fatalf("cerrar la fila: %v", err)
	}
	if en, _ := e.repo.ItemEnSprint(ctx, e.item, e.sprint); en {
		t.Error("una fila no_completado cuenta como estar en el sprint")
	}
}

// US-010 / CA-010-1, RN-010-5, RN-010-9: marcar guarda el estado completado y la fecha,
// devuelve el item como quedo y deja la fila del sprint pendiente (el resultado lo pone
// US-011). ObtenerItem lee la fecha.
func TestGuardarCambioDeEstado_Marca(t *testing.T) {
	ctx := contexto(t)
	e := asignadoEnPostgres(ctx, t)

	item, err := e.repo.GuardarCambioDeEstado(ctx, e.item, e.sprint, completarEn(marcadoEn))
	if err != nil {
		t.Fatalf("GuardarCambioDeEstado: %v", err)
	}

	if item.ID != e.item || item.Estado != backlog.EstadoCompletado || !item.CompletadoEn.Equal(marcadoEn) {
		t.Errorf("item devuelto = %+v, se esperaba el item completado el %v", item, marcadoEn)
	}
	guardado := leerEstado(ctx, t, e, e.item)
	if guardado.estado != "completado" || guardado.completadoEn == nil || !guardado.completadoEn.Equal(marcadoEn) {
		t.Errorf("guardado = (%q, %v), se esperaba (completado, %v)", guardado.estado, guardado.completadoEn, marcadoEn)
	}
	if filas := filasDeLaRelacion(ctx, t, e.pool, e.item); len(filas) != 1 || filas[0] != "pendiente" {
		t.Errorf("filas de la relacion = %v, se esperaba una pendiente", filas)
	}
	leido, _, err := e.repo.ObtenerItem(ctx, e.item)
	if err != nil || !leido.CompletadoEn.Equal(marcadoEn) {
		t.Errorf("ObtenerItem = (%v, %v), se esperaba la fecha %v", leido.CompletadoEn, err, marcadoEn)
	}
}

// US-010 / CA-010-2, RN-010-7: desmarcar deja el item pendiente y la fecha nula.
// ObtenerItem lee la fecha nula como "sin fecha".
func TestGuardarCambioDeEstado_Desmarca(t *testing.T) {
	ctx := contexto(t)
	e := asignadoEnPostgres(ctx, t)
	if _, err := e.repo.GuardarCambioDeEstado(ctx, e.item, e.sprint, completarEn(marcadoEn)); err != nil {
		t.Fatalf("marcar: %v", err)
	}

	item, err := e.repo.GuardarCambioDeEstado(ctx, e.item, e.sprint, backlog.ItemBacklog.Reabrir)
	if err != nil {
		t.Fatalf("desmarcar: %v", err)
	}

	if item.Estado != backlog.EstadoPendiente || !item.CompletadoEn.IsZero() {
		t.Errorf("item devuelto = (%q, %v), se esperaba (pendiente, sin fecha)", item.Estado, item.CompletadoEn)
	}
	if guardado := leerEstado(ctx, t, e, e.item); guardado.estado != "pendiente" || guardado.completadoEn != nil {
		t.Errorf("guardado = (%q, %v), se esperaba (pendiente, nulo)", guardado.estado, guardado.completadoEn)
	}
	if leido, _, _ := e.repo.ObtenerItem(ctx, e.item); !leido.CompletadoEn.IsZero() {
		t.Errorf("ObtenerItem leyo la fecha %v, se esperaba sin fecha", leido.CompletadoEn)
	}
}

// US-010 / CA-010-3, RN-010-6: marcar otra vez no cambia la fecha guardada.
func TestGuardarCambioDeEstado_MarcarDosVeces(t *testing.T) {
	ctx := contexto(t)
	e := asignadoEnPostgres(ctx, t)
	for _, en := range []time.Time{marcadoEn, segundaVez} {
		if _, err := e.repo.GuardarCambioDeEstado(ctx, e.item, e.sprint, completarEn(en)); err != nil {
			t.Fatalf("marcar el %v: %v", en, err)
		}
	}
	if guardado := leerEstado(ctx, t, e, e.item); guardado.completadoEn == nil || !guardado.completadoEn.Equal(marcadoEn) {
		t.Errorf("completado_en = %v, se esperaba la del primer marcado %v", guardado.completadoEn, marcadoEn)
	}
}

// US-010 / CA-010-4, CA-010-5, RN-010-4, RN-010-9, CL-010-5, CL-010-6, CL-010-9: el
// repositorio vuelve a verificar las reglas con el sprint bloqueado. Un item fuera del
// sprint se rechaza con ErrItemFueraDelSprint. Con el sprint cerrado (aunque se haya
// cerrado despues de la lectura del caso de uso) el unico error es ErrSprintCerrado, y
// el item queda como estaba, tambien uno que quedo completado en ese sprint.
func TestGuardarCambioDeEstado_Reglas(t *testing.T) {
	ctx := contexto(t)

	t.Run("item fuera del sprint", func(t *testing.T) {
		e := nuevoEscenarioAsignaciones(ctx, t)
		_, err := e.repo.GuardarCambioDeEstado(ctx, e.item, e.sprint, completarEn(marcadoEn))
		if !errors.Is(err, sprint.ErrItemFueraDelSprint) || errors.Is(err, sprint.ErrSprintCerrado) {
			t.Fatalf("error = %v, se esperaba solo ErrItemFueraDelSprint", err)
		}
		if guardado := leerEstado(ctx, t, e, e.item); guardado.estado != "pendiente" {
			t.Errorf("se cambio el item: %+v", guardado)
		}
	})

	t.Run("sprint cerrado en el medio", func(t *testing.T) {
		e := asignadoEnPostgres(ctx, t)
		cerrarSprint(ctx, t, e.pool, e.sprint)
		_, err := e.repo.GuardarCambioDeEstado(ctx, e.item, e.sprint, completarEn(marcadoEn))
		if !errors.Is(err, sprint.ErrSprintCerrado) || errors.Is(err, sprint.ErrItemFueraDelSprint) {
			t.Fatalf("error = %v, se esperaba solo ErrSprintCerrado", err)
		}
		if guardado := leerEstado(ctx, t, e, e.item); guardado.estado != "pendiente" {
			t.Errorf("se cambio el item: %+v", guardado)
		}
	})

	t.Run("desmarcar en un sprint cerrado", func(t *testing.T) {
		e := asignadoEnPostgres(ctx, t)
		if _, err := e.repo.GuardarCambioDeEstado(ctx, e.item, e.sprint, completarEn(marcadoEn)); err != nil {
			t.Fatalf("marcar: %v", err)
		}
		// Asi deja US-011 el sprint cerrado: la fila ya no esta pendiente.
		cerrarSprint(ctx, t, e.pool, e.sprint)
		if _, err := e.pool.Exec(ctx, `UPDATE sprint_items SET resultado = 'completado' WHERE item_id = $1`, e.item); err != nil {
			t.Fatalf("cerrar la fila: %v", err)
		}

		_, err := e.repo.GuardarCambioDeEstado(ctx, e.item, e.sprint, backlog.ItemBacklog.Reabrir)

		if !errors.Is(err, sprint.ErrSprintCerrado) || errors.Is(err, sprint.ErrItemFueraDelSprint) {
			t.Fatalf("error = %v, se esperaba solo ErrSprintCerrado", err)
		}
		if guardado := leerEstado(ctx, t, e, e.item); guardado.estado != "completado" || guardado.completadoEn == nil {
			t.Errorf("se cambio lo medido: %+v", guardado)
		}
	})
}

// US-010 / CA-010-6, RN-010-9: un sprint o un item que no existen (o con un ID que no
// es un UUID) se rechazan con ErrSprintInexistente o ErrItemInexistente. El sprint se
// verifica primero, porque es el que se bloquea primero.
func TestGuardarCambioDeEstado_Inexistentes(t *testing.T) {
	ctx := contexto(t)
	e := asignadoEnPostgres(ctx, t)
	casos := []struct {
		nombre           string
		itemID, sprintID string
		errEsperado      error
	}{
		{"sprint inexistente", e.item, idDeAsignacionSinFila, app.ErrSprintInexistente},
		{"sprint mal formado", e.item, "no-es-un-uuid", app.ErrSprintInexistente},
		{"item inexistente", idDeAsignacionSinFila, e.sprint, app.ErrItemInexistente},
		{"item mal formado", "no-es-un-uuid", e.sprint, app.ErrItemInexistente},
		{"los dos", idDeAsignacionSinFila, idDeAsignacionSinFila, app.ErrSprintInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := e.repo.GuardarCambioDeEstado(ctx, c.itemID, c.sprintID, completarEn(marcadoEn))
			if !errors.Is(err, c.errEsperado) {
				t.Errorf("error = %v, se esperaba %v", err, c.errEsperado)
			}
		})
	}
	if guardado := leerEstado(ctx, t, e, e.item); guardado.estado != "pendiente" {
		t.Errorf("se cambio el item: %+v", guardado)
	}
}

// US-010 / spec seccion 5: el CHECK de la migracion ata la fecha al estado. Un item
// completado sin fecha, o con fecha y no completado, es un dato corrupto.
func TestMigracion_CompletadoConFecha(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	for nombre, sentencia := range map[string]string{
		"completado sin fecha":  `UPDATE items_backlog SET estado = 'completado' WHERE id = $1`,
		"pendiente con fecha":   `UPDATE items_backlog SET completado_en = now() WHERE id = $1`,
		"en progreso con fecha": `UPDATE items_backlog SET estado = 'en_progreso', completado_en = now() WHERE id = $1`,
	} {
		_, err := e.pool.Exec(ctx, sentencia, e.item)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
			t.Errorf("%s: error = %v, se esperaba la violacion del CHECK (23514)", nombre, err)
		}
	}
	if _, err := e.pool.Exec(ctx, `UPDATE items_backlog SET estado = 'completado', completado_en = now() WHERE id = $1`, e.item); err != nil {
		t.Errorf("completado con fecha: %v", err)
	}
}

// US-010 / RN-010-9, CL-010-10: dos marcados simultaneos del mismo item esperan el
// candado del sprint. El segundo ve el item ya completado y no cambia nada: los dos
// terminan bien y queda la fecha del primero. Para que compitan de verdad, el test
// retiene el candado con una transaccion propia; el segundo marcado arranca recien
// cuando el primero ya espera, asi el orden en que lo obtienen es fijo.
func TestGuardarCambioDeEstado_Concurrente(t *testing.T) {
	ctx := contexto(t)
	e := asignadoEnPostgres(ctx, t)

	candado, err := e.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir la transaccion del candado: %v", err)
	}
	defer func() { _ = candado.Rollback(ctx) }()
	if _, err := candado.Exec(ctx, `SELECT 1 FROM sprints WHERE id = $1 FOR NO KEY UPDATE`, e.sprint); err != nil {
		t.Fatalf("tomar el candado del sprint: %v", err)
	}

	errores := make(chan error, 2)
	for i, en := range []time.Time{marcadoEn, segundaVez} {
		go func() {
			_, err := e.repo.GuardarCambioDeEstado(ctx, e.item, e.sprint, completarEn(en))
			errores <- err
		}()
		if bloqueadas := esperarBloqueadas(ctx, t, e.pool, i+1); bloqueadas < i+1 {
			t.Fatalf("solo %d de %d marcados esperaron el candado del sprint: el repositorio no lo toma", bloqueadas, i+1)
		}
	}
	if err := candado.Rollback(ctx); err != nil {
		t.Fatalf("soltar el candado: %v", err)
	}

	for range 2 {
		if err := <-errores; err != nil {
			t.Errorf("GuardarCambioDeEstado: %v", err)
		}
	}
	if guardado := leerEstado(ctx, t, e, e.item); guardado.completadoEn == nil || !guardado.completadoEn.Equal(marcadoEn) {
		t.Errorf("completado_en = %v, se esperaba la del primer marcado %v", guardado.completadoEn, marcadoEn)
	}
}
