package postgres_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// idDeAsignacionSinFila es un UUID bien formado que ninguna fila tiene.
const idDeAsignacionSinFila = "00000000-0000-0000-0000-000000009999"

// escenarioAsignaciones es la base migrada con un proyecto "Demo" (del 2026-10-06 al
// 2026-12-31) con un sprint abierto y un item listo para asignar: estimado y con un
// criterio.
type escenarioAsignaciones struct {
	pool     *pgxpool.Pool
	repo     *postgres.RepositorioAsignaciones
	backlog  *postgres.RepositorioBacklog
	sprints  *postgres.RepositorioSprints
	proyecto string
	sprint   string
	item     string
}

// itemParaSprint registra un item con estas condiciones: estimado con 5 puntos (o sin
// estimar) y con cierta cantidad de criterios.
func (e escenarioAsignaciones) itemParaSprint(ctx context.Context, t *testing.T, proyecto, titulo string, estimado bool, criterios int) string {
	t.Helper()
	cs := make([]string, criterios)
	for i := range cs {
		cs[i] = "Criterio"
	}
	item, err := backlog.NuevoItem(backlog.DatosItem{Titulo: titulo, Prioridad: backlog.PrioridadMust, Criterios: cs}, creadoEn)
	if err != nil {
		t.Fatalf("NuevoItem: %v", err)
	}
	registrado, err := e.backlog.RegistrarItem(ctx, proyecto, item)
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	if estimado {
		puntos, _ := backlog.NuevosStoryPoints(5)
		if err := e.backlog.GuardarEstimacion(ctx, registrado.ID, puntos); err != nil {
			t.Fatalf("GuardarEstimacion: %v", err)
		}
	}
	return registrado.ID
}

func nuevoEscenarioAsignaciones(ctx context.Context, t *testing.T) escenarioAsignaciones {
	t.Helper()
	pool := dbprueba.BaseMigrada(ctx, t)
	e := escenarioAsignaciones{
		pool:     pool,
		repo:     postgres.NuevoRepositorioAsignaciones(pool),
		backlog:  postgres.NuevoRepositorioBacklog(pool),
		sprints:  postgres.NuevoRepositorioSprints(pool),
		proyecto: proyectoConFechas(ctx, t, pool, "Demo", diaSprint(10, 6), diaSprint(12, 31)),
	}
	abierto, err := e.sprints.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 6), diaSprint(10, 19)))
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}
	e.sprint = abierto.ID
	e.item = e.itemParaSprint(ctx, t, e.proyecto, "Crear proyecto", true, 1)
	return e
}

// filasDeLaRelacion devuelve los resultados de las filas de un item, de la mas vieja a
// la mas nueva.
func filasDeLaRelacion(ctx context.Context, t *testing.T, pool *pgxpool.Pool, itemID string) []string {
	t.Helper()
	filas, err := pool.Query(ctx, `SELECT resultado FROM sprint_items WHERE item_id = $1 ORDER BY asignado_en, sprint_id`, itemID)
	if err != nil {
		t.Fatalf("leer la relacion: %v", err)
	}
	defer filas.Close()
	var resultados []string
	for filas.Next() {
		var r string
		if err := filas.Scan(&r); err != nil {
			t.Fatalf("leer la fila: %v", err)
		}
		resultados = append(resultados, r)
	}
	return resultados
}

// US-009 / CA-009-1, RN-009-9, RN-009-13: asignar crea una fila de la relacion con
// resultado pendiente y no modifica el item: conserva su estado, sus Story Points, sus
// criterios y su prioridad.
func TestAsignarItemASprint_Asigna(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	antes := leerFila(ctx, t, e.pool, e.item)

	if err := e.repo.AsignarItemASprint(ctx, e.item, e.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}

	if filas := filasDeLaRelacion(ctx, t, e.pool, e.item); len(filas) != 1 || filas[0] != "pendiente" {
		t.Errorf("filas de la relacion = %v, se esperaba una pendiente", filas)
	}
	despues := leerFila(ctx, t, e.pool, e.item)
	if despues.estado != antes.estado || despues.prioridad != antes.prioridad || despues.titulo != antes.titulo ||
		*despues.storyPoints != *antes.storyPoints || len(despues.criterios) != len(antes.criterios) {
		t.Errorf("el item cambio al asignarlo: antes %+v, despues %+v", antes, despues)
	}
}

// US-009 / CA-009-2, CL-009-6: se pueden asignar varios items al mismo sprint.
func TestAsignarItemASprint_VariosItemsAlMismoSprint(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	otro := e.itemParaSprint(ctx, t, e.proyecto, "Listar proyectos", true, 2)

	for _, id := range []string{e.item, otro} {
		if err := e.repo.AsignarItemASprint(ctx, id, e.sprint); err != nil {
			t.Fatalf("AsignarItemASprint(%s): %v", id, err)
		}
	}

	var n int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM sprint_items WHERE sprint_id = $1`, e.sprint).Scan(&n); err != nil || n != 2 {
		t.Errorf("el sprint tiene %d items (error %v), se esperaban 2", n, err)
	}
}

// US-009 / RN-009-1, RN-009-2: ObtenerSprint devuelve el sprint con su nombre, proyecto
// y estado; un ID que ningun sprint tiene o mal formado es un sprint inexistente.
func TestObtenerSprint(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)

	s, existe, err := e.repo.ObtenerSprint(ctx, e.sprint)
	if err != nil || !existe {
		t.Fatalf("ObtenerSprint = (existe %t, err %v), se esperaba encontrarlo", existe, err)
	}
	if s.ID != e.sprint || s.ProyectoID != e.proyecto || s.Numero != 1 || s.Nombre() != "Sprint 1" || s.Estado != sprint.EstadoAbierto {
		t.Errorf("sprint = %+v, se esperaba el Sprint 1 abierto del proyecto", s)
	}

	for _, id := range []string{idDeAsignacionSinFila, "no-es-un-uuid", ""} {
		if _, existe, err := e.repo.ObtenerSprint(ctx, id); err != nil || existe {
			t.Errorf("ObtenerSprint(%q) = (existe %t, err %v), se esperaba (false, nil)", id, existe, err)
		}
	}
}

// El repositorio tambien cumple la lectura de items de US-013 y US-006: es la misma
// consulta, asi que se reutiliza.
func TestAsignaciones_ObtenerItem(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)

	item, existe, err := e.repo.ObtenerItem(ctx, e.item)

	if err != nil || !existe || item.ID != e.item || item.ProyectoID != e.proyecto || len(item.Criterios) != 1 {
		t.Errorf("ObtenerItem = (%+v, %t, %v), se esperaba el item del escenario", item, existe, err)
	}
}

// US-009 / RN-009-6, RN-009-13, CL-009-12: ItemEnSprintAbierto es verdadero solo con una
// fila pendiente: un item que estuvo en un sprint cerrado (fila no_completado) no
// cuenta, y un ID mal formado es falso.
func TestItemEnSprintAbierto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)

	if en, err := e.repo.ItemEnSprintAbierto(ctx, e.item); err != nil || en {
		t.Fatalf("antes de asignar: (%t, %v), se esperaba (false, nil)", en, err)
	}
	if err := e.repo.AsignarItemASprint(ctx, e.item, e.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	if en, err := e.repo.ItemEnSprintAbierto(ctx, e.item); err != nil || !en {
		t.Errorf("despues de asignar: (%t, %v), se esperaba (true, nil)", en, err)
	}

	// El sprint se cierra y la fila queda como no completada (lo hace US-011).
	if _, err := e.pool.Exec(ctx, `UPDATE sprint_items SET resultado = 'no_completado' WHERE item_id = $1`, e.item); err != nil {
		t.Fatalf("pasar la fila a no_completado: %v", err)
	}
	if en, err := e.repo.ItemEnSprintAbierto(ctx, e.item); err != nil || en {
		t.Errorf("con la fila no_completado: (%t, %v), se esperaba (false, nil)", en, err)
	}
	for _, id := range []string{idDeAsignacionSinFila, "no-es-un-uuid"} {
		if en, err := e.repo.ItemEnSprintAbierto(ctx, id); err != nil || en {
			t.Errorf("ItemEnSprintAbierto(%q) = (%t, %v), se esperaba (false, nil)", id, en, err)
		}
	}
}

// US-009 / CL-009-12: un item no completado de un sprint cerrado, devuelto al backlog,
// conserva su fila en ese sprint (no_completado) y se puede asignar al sprint abierto:
// queda con dos filas, una por sprint.
func TestAsignarItemASprint_DespuesDeUnSprintCerrado(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	if err := e.repo.AsignarItemASprint(ctx, e.item, e.sprint); err != nil {
		t.Fatalf("primera asignacion: %v", err)
	}
	cerrarSprint(ctx, t, e.pool, e.sprint)
	if _, err := e.pool.Exec(ctx, `UPDATE sprint_items SET resultado = 'no_completado' WHERE item_id = $1`, e.item); err != nil {
		t.Fatalf("pasar la fila a no_completado: %v", err)
	}
	siguiente, err := e.sprints.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 20), diaSprint(11, 2)))
	if err != nil {
		t.Fatalf("segundo sprint: %v", err)
	}

	if err := e.repo.AsignarItemASprint(ctx, e.item, siguiente.ID); err != nil {
		t.Fatalf("segunda asignacion: %v", err)
	}

	filas := filasDeLaRelacion(ctx, t, e.pool, e.item)
	if len(filas) != 2 || filas[0] != "no_completado" || filas[1] != "pendiente" {
		t.Errorf("filas de la relacion = %v, se esperaban no_completado y pendiente", filas)
	}
}

// US-009 / RN-009-6, CL-009-7, CA-009-7: un item ya asignado al sprint abierto se
// rechaza con ErrItemYaAsignado, tambien si es el mismo sprint, y no con un error de la
// clave primaria.
func TestAsignarItemASprint_YaAsignado(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	if err := e.repo.AsignarItemASprint(ctx, e.item, e.sprint); err != nil {
		t.Fatalf("primera asignacion: %v", err)
	}

	err := e.repo.AsignarItemASprint(ctx, e.item, e.sprint)

	if !errors.Is(err, sprint.ErrItemYaAsignado) {
		t.Fatalf("error = %v, se esperaba ErrItemYaAsignado", err)
	}
	if filas := filasDeLaRelacion(ctx, t, e.pool, e.item); len(filas) != 1 {
		t.Errorf("filas de la relacion = %v, se esperaba una", filas)
	}
}

// US-009 / RN-009-3, CL-009-9: un item y un sprint de proyectos distintos se rechazan.
func TestAsignarItemASprint_ProyectoDistinto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	otro := proyectoConFechas(ctx, t, e.pool, "Otro", diaSprint(10, 6), diaSprint(12, 31))
	delOtro := e.itemParaSprint(ctx, t, otro, "Historia del otro proyecto", true, 1)

	err := e.repo.AsignarItemASprint(ctx, delOtro, e.sprint)

	if !errors.Is(err, sprint.ErrProyectoDistinto) {
		t.Errorf("error = %v, se esperaba ErrProyectoDistinto", err)
	}
}

// US-009 / RN-009-4, CL-009-14, CA-009-6: el repositorio vuelve a verificar el estado con
// la fila del sprint bloqueada: si se cerro entre la lectura del caso de uso y la
// transaccion, se rechaza con ErrSprintCerrado y no se guarda nada.
func TestAsignarItemASprint_SprintCerrado(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	cerrarSprint(ctx, t, e.pool, e.sprint)

	err := e.repo.AsignarItemASprint(ctx, e.item, e.sprint)

	if !errors.Is(err, sprint.ErrSprintCerrado) {
		t.Fatalf("error = %v, se esperaba ErrSprintCerrado", err)
	}
	if filas := filasDeLaRelacion(ctx, t, e.pool, e.item); len(filas) != 0 {
		t.Errorf("filas de la relacion = %v, se esperaba ninguna", filas)
	}
}

// US-009 / RN-009-5, CL-009-8: un item completado se rechaza solo con
// backlog.ErrItemCompletado, tambien si completo en un sprint anterior: no esta en el
// sprint abierto, asi que no es ErrItemYaAsignado.
func TestAsignarItemASprint_ItemCompletado(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	if _, err := e.pool.Exec(ctx, `UPDATE items_backlog SET estado = 'completado' WHERE id = $1`, e.item); err != nil {
		t.Fatalf("completar el item: %v", err)
	}
	if _, err := e.pool.Exec(ctx, `INSERT INTO sprint_items (item_id, sprint_id, resultado) VALUES ($1, $2, 'completado')`, e.item, e.sprint); err != nil {
		t.Fatalf("registrar que completo en un sprint anterior: %v", err)
	}
	cerrarSprint(ctx, t, e.pool, e.sprint)
	nuevo, err := e.sprints.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 20), diaSprint(11, 2)))
	if err != nil {
		t.Fatalf("segundo sprint: %v", err)
	}

	err = e.repo.AsignarItemASprint(ctx, e.item, nuevo.ID)

	if !errors.Is(err, backlog.ErrItemCompletado) {
		t.Fatalf("error = %v, se esperaba ErrItemCompletado", err)
	}
	if errors.Is(err, sprint.ErrItemYaAsignado) {
		t.Errorf("el error %q tambien reporta item ya asignado", err)
	}
}

// US-009 / RN-009-7, CL-009-3, CL-009-4, CL-009-5, CA-009-4, CA-009-5: la Definition of
// Ready se verifica con los datos leidos con los bloqueos: sin estimar, sin criterios, o
// las dos cosas unidas.
func TestAsignarItemASprint_DefinitionOfReady(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)

	casos := []struct {
		nombre    string
		estimado  bool
		criterios int
		esperados []error
		noEsperan []error
	}{
		{"sin estimar, con criterios", false, 2, []error{sprint.ErrItemSinEstimar}, []error{sprint.ErrItemSinCriterios}},
		{"estimado, sin criterios", true, 0, []error{sprint.ErrItemSinCriterios}, []error{sprint.ErrItemSinEstimar}},
		{"sin estimar y sin criterios", false, 0, []error{sprint.ErrItemSinEstimar, sprint.ErrItemSinCriterios}, nil},
	}
	for i, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			item := e.itemParaSprint(ctx, t, e.proyecto, "Historia "+string(rune('A'+i)), c.estimado, c.criterios)

			err := e.repo.AsignarItemASprint(ctx, item, e.sprint)

			for _, esperado := range c.esperados {
				if !errors.Is(err, esperado) {
					t.Errorf("error = %v, se esperaba que incluyera %v", err, esperado)
				}
			}
			for _, noEsperado := range c.noEsperan {
				if errors.Is(err, noEsperado) {
					t.Errorf("error = %v, no se esperaba %v", err, noEsperado)
				}
			}
			if filas := filasDeLaRelacion(ctx, t, e.pool, item); len(filas) != 0 {
				t.Errorf("filas de la relacion = %v, se esperaba ninguna", filas)
			}
		})
	}
}

// US-009 / RN-009-1, RN-009-2, RN-009-11, CA-009-8: si el sprint o el item no existen (o el
// ID esta mal formado) se devuelve el error de inexistencia que corresponde y no se
// guarda nada. El sprint se verifica primero, como en los pasos de RN-009-11.
func TestAsignarItemASprint_LoQueNoExiste(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)

	casos := []struct {
		nombre   string
		item     string
		sprint   string
		esperado error
	}{
		{"el sprint no existe", e.item, idDeAsignacionSinFila, app.ErrSprintInexistente},
		{"el ID del sprint esta mal formado", e.item, "no-es-un-uuid", app.ErrSprintInexistente},
		{"el item no existe", idDeAsignacionSinFila, e.sprint, app.ErrItemInexistente},
		{"el ID del item esta mal formado", "no-es-un-uuid", e.sprint, app.ErrItemInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := e.repo.AsignarItemASprint(ctx, c.item, c.sprint)

			if !errors.Is(err, c.esperado) {
				t.Errorf("error = %v, se esperaba %v", err, c.esperado)
			}
		})
	}
	var n int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM sprint_items`).Scan(&n); err != nil || n != 0 {
		t.Errorf("la relacion tiene %d filas (error %v), se esperaba ninguna", n, err)
	}
}

// US-009 / CL-009-10, RN-009-11: dos asignaciones simultaneas del mismo item al mismo
// sprint: la segunda espera el bloqueo del item, ve la fila de la primera y se rechaza
// siempre con ErrItemYaAsignado. Ninguna termina con un error de la clave primaria ni
// del indice de la relacion.
func TestAsignarItemASprint_Simultaneas(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	const intentos = 5

	var wg sync.WaitGroup
	errs := make([]error, intentos)
	for i := 0; i < intentos; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = e.repo.AsignarItemASprint(ctx, e.item, e.sprint)
		}()
	}
	wg.Wait()

	exitosas := 0
	for i, err := range errs {
		switch {
		case err == nil:
			exitosas++
		case errors.Is(err, sprint.ErrItemYaAsignado):
		default:
			t.Errorf("intento %d: error = %v, se esperaba nil o ErrItemYaAsignado", i, err)
		}
	}
	if exitosas != 1 {
		t.Errorf("terminaron bien %d asignaciones, se esperaba exactamente 1", exitosas)
	}
	if filas := filasDeLaRelacion(ctx, t, e.pool, e.item); len(filas) != 1 {
		t.Errorf("filas de la relacion = %v, se esperaba una", filas)
	}
}

// RN-009-6, RN-009-13: las restricciones de la tabla son la red de seguridad: un item con
// dos filas pendientes, una fila repetida para el mismo sprint y un resultado
// desconocido se rechazan aunque el repositorio no los verificara.
func TestSprintItems_RedDeSeguridadDeLaTabla(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	siguiente := func() string {
		cerrarSprint(ctx, t, e.pool, e.sprint)
		s, err := e.sprints.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 20), diaSprint(11, 2)))
		if err != nil {
			t.Fatalf("segundo sprint: %v", err)
		}
		return s.ID
	}()
	const insertar = `INSERT INTO sprint_items (item_id, sprint_id, resultado) VALUES ($1, $2, $3)`
	if _, err := e.pool.Exec(ctx, insertar, e.item, e.sprint, "pendiente"); err != nil {
		t.Fatalf("primera fila: %v", err)
	}

	casos := []struct {
		nombre    string
		sprint    string
		resultado string
		parte     string
	}{
		{"dos filas pendientes del mismo item", siguiente, "pendiente", "un_sprint_pendiente_por_item"},
		{"fila repetida para el mismo sprint", e.sprint, "no_completado", "sprint_items_pkey"},
		{"resultado desconocido", siguiente, "perdido", "resultado"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := e.pool.Exec(ctx, insertar, e.item, c.sprint, c.resultado)

			if err == nil || !strings.Contains(err.Error(), c.parte) {
				t.Errorf("error = %v, se esperaba que lo rechazara %q", err, c.parte)
			}
		})
	}
}

// La fila de la relacion se borra con el item o con el sprint (ON DELETE CASCADE): no
// quedan asignaciones huerfanas.
func TestSprintItems_SeBorraConElItem(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	if err := e.repo.AsignarItemASprint(ctx, e.item, e.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}

	if _, err := e.pool.Exec(ctx, `DELETE FROM items_backlog WHERE id = $1`, e.item); err != nil {
		t.Fatalf("borrar el item: %v", err)
	}

	var n int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM sprint_items`).Scan(&n); err != nil || n != 0 {
		t.Errorf("quedaron %d filas (error %v), se esperaba ninguna", n, err)
	}
}
