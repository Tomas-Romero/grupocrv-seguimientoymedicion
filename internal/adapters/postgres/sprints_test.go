package postgres_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// idDeProyectoSinFila es un UUID bien formado que ningun proyecto tiene.
const idDeProyectoSinFila = "00000000-0000-0000-0000-000000009999"

func diaSprint(mes time.Month, d int) time.Time {
	return time.Date(2026, mes, d, 0, 0, 0, 0, time.UTC)
}

// proyectoConFechas inserta un proyecto con esas fechas y devuelve su ID.
func proyectoConFechas(ctx context.Context, t *testing.T, pool *pgxpool.Pool, nombre string, inicio, fin time.Time) string {
	t.Helper()
	var id string
	const insertar = `
INSERT INTO proyectos (id, nombre, fecha_inicio, fecha_fin)
VALUES (gen_random_uuid(), $1, $2, $3)
RETURNING id::text`
	if err := pool.QueryRow(ctx, insertar, nombre, inicio, fin).Scan(&id); err != nil {
		t.Fatalf("crear el proyecto %q: %v", nombre, err)
	}
	return id
}

// escenarioSprints es la base migrada con un proyecto "Demo" del 2026-10-06 al
// 2026-12-31.
type escenarioSprints struct {
	pool     *pgxpool.Pool
	repo     *postgres.RepositorioSprints
	proyecto string
}

func nuevoEscenarioSprints(ctx context.Context, t *testing.T) escenarioSprints {
	t.Helper()
	pool := dbprueba.BaseMigrada(ctx, t)
	return escenarioSprints{
		pool:     pool,
		repo:     postgres.NuevoRepositorioSprints(pool),
		proyecto: proyectoConFechas(ctx, t, pool, "Demo", diaSprint(10, 6), diaSprint(12, 31)),
	}
}

// sprintNuevo arma un sprint valido con el dominio, como lo haria el caso de uso.
func sprintNuevo(t *testing.T, proyecto string, inicio, fin time.Time) sprint.Sprint {
	t.Helper()
	s, err := sprint.Nuevo(proyecto, "Objetivo", inicio, fin, diaSprint(1, 1), diaSprint(12, 31))
	if err != nil {
		t.Fatalf("sprint.Nuevo: %v", err)
	}
	return s
}

// cerrarSprint cierra un sprint directamente en la tabla: cerrar es US-011.
func cerrarSprint(ctx context.Context, t *testing.T, pool *pgxpool.Pool, id string) {
	t.Helper()
	if _, err := pool.Exec(ctx, `UPDATE sprints SET estado = 'cerrado' WHERE id = $1`, id); err != nil {
		t.Fatalf("cerrar el sprint %s: %v", id, err)
	}
}

// filaDeSprint es lo que quedo guardado, leido directamente de la tabla.
type filaDeSprint struct {
	proyectoID string
	numero     int
	goal       string
	inicio     time.Time
	fin        time.Time
	estado     string
	creadoEn   time.Time
}

func leerSprint(ctx context.Context, t *testing.T, pool *pgxpool.Pool, id string) filaDeSprint {
	t.Helper()
	var f filaDeSprint
	const consulta = `
SELECT proyecto_id::text, numero, sprint_goal, fecha_inicio, fecha_fin, estado, creado_en
FROM sprints WHERE id = $1`
	if err := pool.QueryRow(ctx, consulta, id).Scan(&f.proyectoID, &f.numero, &f.goal, &f.inicio, &f.fin, &f.estado, &f.creadoEn); err != nil {
		t.Fatalf("leer el sprint %s: %v", id, err)
	}
	return f
}

func cantidadDeSprints(ctx context.Context, t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM sprints`).Scan(&n); err != nil {
		t.Fatalf("contar los sprints: %v", err)
	}
	return n
}

// US-008 / RN-008-9: FechasDelProyecto devuelve las fechas del proyecto y dice si
// existe. Un ID mal formado es un proyecto inexistente, no un error.
func TestFechasDelProyecto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioSprints(ctx, t)

	inicio, fin, existe, err := e.repo.FechasDelProyecto(ctx, e.proyecto)
	if err != nil || !existe {
		t.Fatalf("FechasDelProyecto = (existe %t, err %v), se esperaba encontrarlo", existe, err)
	}
	if !inicio.Equal(diaSprint(10, 6)) || !fin.Equal(diaSprint(12, 31)) {
		t.Errorf("fechas = %v a %v, se esperaban 2026-10-06 a 2026-12-31", inicio, fin)
	}

	for _, id := range []string{idDeProyectoSinFila, "no-es-un-uuid", ""} {
		if _, _, existe, err := e.repo.FechasDelProyecto(ctx, id); err != nil || existe {
			t.Errorf("FechasDelProyecto(%q) = (existe %t, err %v), se esperaba (false, nil)", id, existe, err)
		}
	}
}

// US-008 / CA-008-1, CA-008-2, RN-008-8, RN-008-11, RN-008-16: el primer sprint de un
// proyecto es el 1, queda abierto y con un ID que genera la base; las fechas y el
// Sprint Goal se guardan tal cual.
func TestRegistrarSprint_Primero(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioSprints(ctx, t)

	registrado, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 6), diaSprint(10, 19)))
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}

	if registrado.ID == "" || registrado.Numero != 1 || registrado.Nombre() != "Sprint 1" {
		t.Fatalf("ID, numero y nombre = %q, %d, %q; se esperaba un ID de la base y Sprint 1", registrado.ID, registrado.Numero, registrado.Nombre())
	}
	f := leerSprint(ctx, t, e.pool, registrado.ID)
	if f.proyectoID != e.proyecto || f.numero != 1 || f.goal != "Objetivo" || f.estado != "abierto" {
		t.Errorf("fila guardada = %+v, se esperaba el Sprint 1 abierto del proyecto", f)
	}
	if !f.inicio.Equal(diaSprint(10, 6)) || !f.fin.Equal(diaSprint(10, 19)) {
		t.Errorf("fechas guardadas = %v a %v, se esperaban 2026-10-06 a 2026-10-19", f.inicio, f.fin)
	}
	if f.creadoEn.IsZero() || time.Since(f.creadoEn) > time.Minute {
		t.Errorf("creado_en = %v, se esperaba la hora del alta (DEFAULT now())", f.creadoEn)
	}
}

// US-008 / CA-008-2, CL-008-8, RN-008-11: la numeracion es por proyecto; el primer
// sprint de otro proyecto es el 1 aunque el primero tenga un sprint abierto en las
// mismas fechas, y el siguiente de un proyecto toma el numero que sigue.
func TestRegistrarSprint_NumeracionPorProyecto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioSprints(ctx, t)
	otro := proyectoConFechas(ctx, t, e.pool, "Otro", diaSprint(10, 6), diaSprint(12, 31))

	primero, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 6), diaSprint(10, 19)))
	if err != nil {
		t.Fatalf("primer sprint de Demo: %v", err)
	}
	delOtro, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, otro, diaSprint(10, 6), diaSprint(10, 19)))
	if err != nil {
		t.Fatalf("primer sprint de Otro: %v", err)
	}
	cerrarSprint(ctx, t, e.pool, primero.ID)
	segundo, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 20), diaSprint(11, 2)))
	if err != nil {
		t.Fatalf("segundo sprint de Demo: %v", err)
	}

	if delOtro.Numero != 1 {
		t.Errorf("primer sprint de Otro = %d, se esperaba 1", delOtro.Numero)
	}
	if segundo.Numero != 2 || segundo.Nombre() != "Sprint 2" {
		t.Errorf("segundo sprint de Demo = %d (%s), se esperaba el Sprint 2", segundo.Numero, segundo.Nombre())
	}
}

// US-008 / CA-008-10, RN-008-9, RN-008-14: un proyecto que no existe (o un ID mal
// formado) se rechaza con ErrProyectoInexistente y no se guarda nada.
func TestRegistrarSprint_ProyectoInexistente(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioSprints(ctx, t)

	for _, id := range []string{idDeProyectoSinFila, "no-es-un-uuid"} {
		_, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, id, diaSprint(10, 6), diaSprint(10, 19)))
		if !errors.Is(err, app.ErrProyectoInexistente) {
			t.Errorf("proyecto %q: error = %v, se esperaba ErrProyectoInexistente", id, err)
		}
	}
	if n := cantidadDeSprints(ctx, t, e.pool); n != 0 {
		t.Errorf("hay %d sprints, se esperaba ninguno", n)
	}
}

// US-008 / CL-008-18, RN-008-7, RN-008-14: el repositorio vuelve a verificar las
// fechas con las del proyecto leidas con la fila bloqueada: si el proyecto cambio
// y el sprint quedo afuera, se rechaza solo con ErrFueraDelProyecto.
func TestRegistrarSprint_FueraDeLasFechasGuardadasDelProyecto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioSprints(ctx, t)
	// El sprint estaba dentro del proyecto cuando el caso de uso lo valido, pero
	// las fechas del proyecto cambiaron antes del alta.
	if _, err := e.pool.Exec(ctx, `UPDATE proyectos SET fecha_fin = '2026-10-10' WHERE id = $1`, e.proyecto); err != nil {
		t.Fatalf("acortar el proyecto: %v", err)
	}

	_, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 6), diaSprint(10, 19)))

	if !errors.Is(err, sprint.ErrFueraDelProyecto) {
		t.Fatalf("error = %v, se esperaba ErrFueraDelProyecto", err)
	}
	if errors.Is(err, sprint.ErrUltimoSprintAbierto) || errors.Is(err, sprint.ErrInicioNoPosteriorAlUltimo) {
		t.Errorf("el error %q trae algo mas que ErrFueraDelProyecto", err)
	}
	if n := cantidadDeSprints(ctx, t, e.pool); n != 0 {
		t.Errorf("hay %d sprints, se esperaba ninguno", n)
	}
}

// US-008 / CA-008-11, CL-008-17, RN-008-17: con el ultimo sprint abierto se rechaza
// solo con ErrUltimoSprintAbierto, empiece el nuevo despues o antes de que termine
// el abierto, sin comparar fechas.
func TestRegistrarSprint_UltimoAbierto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioSprints(ctx, t)
	if _, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 6), diaSprint(10, 19))); err != nil {
		t.Fatalf("primer sprint: %v", err)
	}

	for _, inicio := range []time.Time{diaSprint(10, 20), diaSprint(10, 10)} {
		_, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, inicio, diaSprint(11, 2)))

		if !errors.Is(err, sprint.ErrUltimoSprintAbierto) {
			t.Fatalf("inicio %s: error = %v, se esperaba ErrUltimoSprintAbierto", inicio.Format(time.DateOnly), err)
		}
		if errors.Is(err, sprint.ErrInicioNoPosteriorAlUltimo) {
			t.Errorf("inicio %s: el error %q tambien reporta el inicio", inicio.Format(time.DateOnly), err)
		}
	}
	if n := cantidadDeSprints(ctx, t, e.pool); n != 1 {
		t.Errorf("hay %d sprints, se esperaba 1", n)
	}
}

// US-008 / CA-008-4, CA-008-9, CL-008-4, CL-008-5, RN-008-13: con el ultimo sprint
// cerrado (armado directamente en la tabla, porque cerrar es US-011), el nuevo
// tiene que empezar despues de que termine: el dia siguiente se acepta; el mismo
// dia, o un sprint completamente anterior, se rechazan.
func TestRegistrarSprint_UltimoCerrado(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioSprints(ctx, t)
	primero, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 20), diaSprint(11, 2)))
	if err != nil {
		t.Fatalf("primer sprint: %v", err)
	}
	cerrarSprint(ctx, t, e.pool, primero.ID)

	rechazados := []struct{ inicio, fin time.Time }{
		{diaSprint(11, 2), diaSprint(11, 3)},  // empieza el mismo dia en que termina el ultimo
		{diaSprint(10, 6), diaSprint(10, 19)}, // completamente anterior al ultimo
	}
	for _, c := range rechazados {
		_, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, c.inicio, c.fin))
		if !errors.Is(err, sprint.ErrInicioNoPosteriorAlUltimo) {
			t.Errorf("inicio %s: error = %v, se esperaba ErrInicioNoPosteriorAlUltimo", c.inicio.Format(time.DateOnly), err)
		}
	}

	siguiente, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(11, 3), diaSprint(11, 16)))
	if err != nil {
		t.Fatalf("empezar el dia siguiente: %v", err)
	}
	if siguiente.Numero != 2 {
		t.Errorf("numero = %d, se esperaba 2", siguiente.Numero)
	}
}

// US-008 / CL-008-15, RN-008-14: dos altas simultaneas en el mismo proyecto: la
// segunda espera el bloqueo del proyecto, ve el sprint abierto de la primera y se
// rechaza siempre con ErrUltimoSprintAbierto. Ninguna termina con un error de la
// restriccion UNIQUE ni del indice de sprints abiertos.
func TestRegistrarSprint_Simultaneos(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioSprints(ctx, t)
	const altas = 5

	var wg sync.WaitGroup
	errs := make([]error, altas)
	for i := 0; i < altas; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = e.repo.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 6), diaSprint(10, 19)))
		}()
	}
	wg.Wait()

	exitosas := 0
	for i, err := range errs {
		switch {
		case err == nil:
			exitosas++
		case errors.Is(err, sprint.ErrUltimoSprintAbierto):
		default:
			t.Errorf("alta %d: error = %v, se esperaba nil o ErrUltimoSprintAbierto", i, err)
		}
	}
	if exitosas != 1 {
		t.Errorf("terminaron bien %d altas, se esperaba exactamente 1", exitosas)
	}
	if n := cantidadDeSprints(ctx, t, e.pool); n != 1 {
		t.Errorf("hay %d sprints, se esperaba 1", n)
	}
}

// US-008 / CL-008-16, RN-008-14: una transaccion que se revierte despues de calcular
// el numero no deja hueco: aca el INSERT falla por la restriccion de fechas de la
// tabla (se arma un sprint con el fin antes del inicio, salteando el dominio) y el
// siguiente alta toma el mismo numero.
func TestRegistrarSprint_RollbackSinHueco(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioSprints(ctx, t)
	invalido := sprintNuevo(t, e.proyecto, diaSprint(10, 6), diaSprint(10, 19))
	invalido.FechaInicio, invalido.FechaFin = diaSprint(10, 19), diaSprint(10, 10)

	if _, err := e.repo.RegistrarSprint(ctx, invalido); err == nil {
		t.Fatal("se esperaba que la tabla rechazara el sprint con el fin antes del inicio")
	}
	siguiente, err := e.repo.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 6), diaSprint(10, 19)))
	if err != nil {
		t.Fatalf("alta despues del rollback: %v", err)
	}

	if siguiente.Numero != 1 {
		t.Errorf("numero = %d, se esperaba 1: el rollback no tiene que dejar hueco", siguiente.Numero)
	}
}

// RN-008-14, RN-008-17: las restricciones de la tabla son la red de seguridad: dos
// sprints abiertos en el mismo proyecto, un numero repetido y un estado desconocido
// se rechazan aunque el repositorio no los verificara.
func TestSprints_RedDeSeguridadDeLaTabla(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioSprints(ctx, t)
	const insertar = `
INSERT INTO sprints (proyecto_id, numero, sprint_goal, fecha_inicio, fecha_fin, estado)
VALUES ($1, $2, 'Objetivo', '2026-10-06', '2026-10-19', $3)`
	if _, err := e.pool.Exec(ctx, insertar, e.proyecto, 1, "abierto"); err != nil {
		t.Fatalf("primer sprint: %v", err)
	}

	casos := []struct {
		nombre string
		numero int
		estado string
		parte  string
	}{
		{"dos sprints abiertos en el mismo proyecto", 2, "abierto", "un_sprint_abierto_por_proyecto"},
		{"numero repetido", 1, "cerrado", "numero_de_sprint_unico_por_proyecto"},
		{"estado desconocido", 3, "suspendido", "estado"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := e.pool.Exec(ctx, insertar, e.proyecto, c.numero, c.estado)

			if err == nil || !strings.Contains(err.Error(), c.parte) {
				t.Errorf("error = %v, se esperaba que lo rechazara %q", err, c.parte)
			}
		})
	}
}
