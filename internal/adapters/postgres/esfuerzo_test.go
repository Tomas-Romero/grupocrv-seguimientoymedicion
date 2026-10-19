package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// idSinDueno es un UUID bien formado que ninguna fila tiene.
const idSinDueno = "00000000-0000-0000-0000-000000009999"

// escenarioEsfuerzo es un proyecto con una historia y un integrante.
type escenarioEsfuerzo struct {
	pool         *pgxpool.Pool
	repo         *postgres.RepositorioEsfuerzo
	proyecto     string
	item         string
	integrante   string
	otroProyecto string
}

// integranteParaEsfuerzo inserta un integrante y devuelve su ID. El ID se genera
// en el INSERT para no depender del DEFAULT de la tabla.
func integranteParaEsfuerzo(ctx context.Context, t *testing.T, pool *pgxpool.Pool, proyecto, email string) string {
	t.Helper()
	var id string
	const insertar = `
INSERT INTO integrantes (id, proyecto_id, nombre, email, rol)
VALUES (gen_random_uuid(), $1, 'Ana', $2, 'product_builder')
RETURNING id::text`
	if err := pool.QueryRow(ctx, insertar, proyecto, email).Scan(&id); err != nil {
		t.Fatalf("crear el integrante %q: %v", email, err)
	}
	return id
}

// nuevoEscenarioEsfuerzo arma la base migrada con dos proyectos: "Demo", con una
// historia y un integrante, y "Otro", vacio.
func nuevoEscenarioEsfuerzo(ctx context.Context, t *testing.T) escenarioEsfuerzo {
	t.Helper()
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	item, err := postgres.NuevoRepositorioBacklog(pool).RegistrarItem(ctx, proyecto, itemConTitulo(t, "Crear proyecto"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	return escenarioEsfuerzo{
		pool:         pool,
		repo:         postgres.NuevoRepositorioEsfuerzo(pool),
		proyecto:     proyecto,
		item:         item.ID,
		integrante:   integranteParaEsfuerzo(ctx, t, pool, proyecto, "ana@ejemplo.test"),
		otroProyecto: crearProyecto(ctx, t, pool, "Otro"),
	}
}

func (e escenarioEsfuerzo) registro(t *testing.T, fecha time.Time, actividad string, horas float64) esfuerzo.RegistroEsfuerzo {
	t.Helper()
	r, err := esfuerzo.Nuevo(e.proyecto, e.item, e.integrante, fecha, actividad, horas, time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("esfuerzo.Nuevo: %v", err)
	}
	return r
}

var ayerDelEsfuerzo = time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)

// filaDeEsfuerzo es lo que quedo guardado, leido directamente de la tabla.
type filaDeEsfuerzo struct {
	proyectoID, itemID, integranteID string
	fecha                            time.Time
	actividad                        string
	horas                            float64
	creadoEn                         time.Time
}

func leerEsfuerzo(ctx context.Context, t *testing.T, pool *pgxpool.Pool, id string) filaDeEsfuerzo {
	t.Helper()
	var f filaDeEsfuerzo
	const consulta = `
SELECT proyecto_id::text, item_id::text, integrante_id::text, fecha, actividad, horas::float8, creado_en
FROM registros_esfuerzo WHERE id = $1`
	if err := pool.QueryRow(ctx, consulta, id).Scan(&f.proyectoID, &f.itemID, &f.integranteID, &f.fecha, &f.actividad, &f.horas, &f.creadoEn); err != nil {
		t.Fatalf("leer el registro %s: %v", id, err)
	}
	return f
}

// US-019 / CA-019-1, RN-019-9, RN-019-10: el registro se guarda con un ID que
// genera la base y con los datos recibidos; la fecha de carga la pone la base.
func TestRegistrarEsfuerzo_Guarda(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)

	guardado, err := e.repo.RegistrarEsfuerzo(ctx, e.registro(t, ayerDelEsfuerzo, "Desarrollo", 2.5))
	if err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}
	if guardado.ID == "" {
		t.Fatal("el registro volvio sin ID: lo genera la base con RETURNING")
	}

	f := leerEsfuerzo(ctx, t, e.pool, guardado.ID)
	if f.proyectoID != e.proyecto || f.itemID != e.item || f.integranteID != e.integrante {
		t.Errorf("IDs guardados = %s, %s, %s; se esperaban los del escenario", f.proyectoID, f.itemID, f.integranteID)
	}
	if !f.fecha.Equal(ayerDelEsfuerzo) {
		t.Errorf("fecha = %v, se esperaba %v", f.fecha, ayerDelEsfuerzo)
	}
	if f.actividad != "Desarrollo" || f.horas != 2.5 {
		t.Errorf("actividad y horas = %q, %v; se esperaban Desarrollo y 2,5", f.actividad, f.horas)
	}
	if f.creadoEn.IsZero() || time.Since(f.creadoEn) > time.Minute {
		t.Errorf("creado_en = %v, se esperaba la hora de la carga (DEFAULT now())", f.creadoEn)
	}
}

// US-019 / CL-019-1, CL-019-2, CL-019-4: los extremos validos de las horas
// (0,25 y 24) y los multiplos de 0,25 se guardan exactos, sin redondeo.
func TestRegistrarEsfuerzo_HorasExactas(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)

	for _, horas := range []float64{0.25, 0.75, 1.5, 23.75, 24} {
		guardado, err := e.repo.RegistrarEsfuerzo(ctx, e.registro(t, ayerDelEsfuerzo, "Desarrollo", horas))
		if err != nil {
			t.Fatalf("RegistrarEsfuerzo(%v): %v", horas, err)
		}
		if f := leerEsfuerzo(ctx, t, e.pool, guardado.ID); f.horas != horas {
			t.Errorf("horas guardadas = %v, se esperaba %v", f.horas, horas)
		}
	}
}

// US-019 / CL-019-11, CL-019-12, RN-019-9: dos registros del mismo integrante,
// historia y dia son dos filas con IDs distintos, aunque sumen mas de 24 horas.
func TestRegistrarEsfuerzo_RegistrosRepetidos(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)

	primero, err := e.repo.RegistrarEsfuerzo(ctx, e.registro(t, ayerDelEsfuerzo, "Desarrollo", 20))
	if err != nil {
		t.Fatalf("primer registro: %v", err)
	}
	segundo, err := e.repo.RegistrarEsfuerzo(ctx, e.registro(t, ayerDelEsfuerzo, "Desarrollo", 20))
	if err != nil {
		t.Fatalf("segundo registro: %v", err)
	}

	if primero.ID == segundo.ID {
		t.Errorf("los dos registros tienen el mismo ID %s", primero.ID)
	}
}

// US-019 / RN-019-9: dos cargas simultaneas del mismo integrante son dos
// registros distintos, y ninguna falla ni espera a la otra.
func TestRegistrarEsfuerzo_Simultaneos(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)
	const cargas = 6

	var wg sync.WaitGroup
	ids := make([]string, cargas)
	errs := make([]error, cargas)
	for i := 0; i < cargas; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			guardado, err := e.repo.RegistrarEsfuerzo(ctx, e.registro(t, ayerDelEsfuerzo, "Desarrollo", 1))
			ids[i], errs[i] = guardado.ID, err
		}()
	}
	wg.Wait()

	vistos := map[string]bool{}
	for i := 0; i < cargas; i++ {
		if errs[i] != nil {
			t.Fatalf("carga %d: %v", i, errs[i])
		}
		vistos[ids[i]] = true
	}
	if len(vistos) != cargas {
		t.Errorf("hay %d IDs distintos, se esperaban %d", len(vistos), cargas)
	}
}

// US-019 / RN-019-6, CL-019-15: la historia existe dentro de su proyecto, y una
// de otro proyecto o un ID inexistente o mal formado no cuenta.
func TestExisteItemEnProyecto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)

	casos := []struct {
		nombre   string
		proyecto string
		item     string
		esperado bool
	}{
		{"la historia es del proyecto", e.proyecto, e.item, true},
		{"la historia existe pero es de otro proyecto", e.otroProyecto, e.item, false},
		{"la historia no existe", e.proyecto, idSinDueno, false},
		{"el ID de la historia esta mal formado", e.proyecto, "no-es-un-uuid", false},
		{"el ID del proyecto esta mal formado", "no-es-un-uuid", e.item, false},
		{"IDs vacios", "", "", false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			existe, err := e.repo.ExisteItemEnProyecto(ctx, c.proyecto, c.item)
			if err != nil || existe != c.esperado {
				t.Errorf("ExisteItemEnProyecto = (%t, %v), se esperaba (%t, nil)", existe, err, c.esperado)
			}
		})
	}
}

// US-019 / RN-019-6, CL-019-15: lo mismo para el integrante.
func TestExisteIntegranteEnProyecto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)

	casos := []struct {
		nombre     string
		proyecto   string
		integrante string
		esperado   bool
	}{
		{"el integrante es del proyecto", e.proyecto, e.integrante, true},
		{"el integrante existe pero es de otro proyecto", e.otroProyecto, e.integrante, false},
		{"el integrante no existe", e.proyecto, idSinDueno, false},
		{"el ID del integrante esta mal formado", e.proyecto, "no-es-un-uuid", false},
		{"el ID del proyecto esta mal formado", "no-es-un-uuid", e.integrante, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			existe, err := e.repo.ExisteIntegranteEnProyecto(ctx, c.proyecto, c.integrante)
			if err != nil || existe != c.esperado {
				t.Errorf("ExisteIntegranteEnProyecto = (%t, %v), se esperaba (%t, nil)", existe, err, c.esperado)
			}
		})
	}
}

// RN-019-6: el proyecto existe o no, y un ID mal formado es inexistente, no un
// error.
func TestExisteProyectoDelEsfuerzo(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)

	for id, esperado := range map[string]bool{e.proyecto: true, idSinDueno: false, "no-es-un-uuid": false, "": false} {
		if existe, err := e.repo.ExisteProyecto(ctx, id); err != nil || existe != esperado {
			t.Errorf("ExisteProyecto(%q) = (%t, %v), se esperaba (%t, nil)", id, existe, err, esperado)
		}
	}
}

// US-019 / RN-019-6, CL-019-15, CA-019-9, CA-019-10: si la historia, el
// integrante o el proyecto desaparecieron despues de verificarlos, la clave
// foranea rechaza el INSERT y el repositorio devuelve el error de inexistencia
// que corresponde, sin guardar nada.
func TestRegistrarEsfuerzo_LoQueDesaparece(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)
	base := e.registro(t, ayerDelEsfuerzo, "Desarrollo", 1)

	casos := []struct {
		nombre   string
		registro esfuerzo.RegistroEsfuerzo
		esperado error
	}{
		{"la historia no existe", func() esfuerzo.RegistroEsfuerzo { r := base; r.ItemID = idSinDueno; return r }(), app.ErrItemInexistente},
		{"el integrante no existe", func() esfuerzo.RegistroEsfuerzo { r := base; r.IntegranteID = idSinDueno; return r }(), app.ErrIntegranteInexistente},
		{"el proyecto no existe", func() esfuerzo.RegistroEsfuerzo { r := base; r.ProyectoID = idSinDueno; return r }(), app.ErrProyectoInexistente},
		{"el ID de la historia esta mal formado", func() esfuerzo.RegistroEsfuerzo { r := base; r.ItemID = "no-es-un-uuid"; return r }(), app.ErrItemInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			guardado, err := e.repo.RegistrarEsfuerzo(ctx, c.registro)

			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.esperado)
			}
			if guardado.ID != "" {
				t.Errorf("devolvio el registro %+v, se esperaba el valor cero", guardado)
			}
			var filas int
			if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM registros_esfuerzo`).Scan(&filas); err != nil || filas != 0 {
				t.Errorf("la tabla tiene %d filas (error %v), se esperaba ninguna", filas, err)
			}
		})
	}
}

// La tabla repite las reglas de las horas como ultima defensa: un valor que el
// dominio habria rechazado (0 o mas de 24) no se puede guardar, y el error no se
// confunde con uno de inexistencia.
func TestRegistrarEsfuerzo_LaTablaRechazaHorasInvalidas(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)

	for _, horas := range []float64{0, -1, 24.25} {
		r := e.registro(t, ayerDelEsfuerzo, "Desarrollo", 1)
		r.Horas = horas

		_, err := e.repo.RegistrarEsfuerzo(ctx, r)

		if err == nil {
			t.Errorf("horas %v: se guardo, se esperaba que la tabla lo rechazara", horas)
			continue
		}
		if errors.Is(err, app.ErrItemInexistente) || errors.Is(err, app.ErrIntegranteInexistente) || errors.Is(err, app.ErrProyectoInexistente) {
			t.Errorf("horas %v: el error %q se confundio con una inexistencia", horas, err)
		}
	}
}

// Los registros se borran junto con su historia, integrante o proyecto (ON DELETE
// CASCADE): no quedan horas huerfanas.
func TestRegistrarEsfuerzo_SeBorraConLaHistoria(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)
	if _, err := e.repo.RegistrarEsfuerzo(ctx, e.registro(t, ayerDelEsfuerzo, "Desarrollo", 1)); err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}

	if _, err := e.pool.Exec(ctx, `DELETE FROM items_backlog WHERE id = $1`, e.item); err != nil {
		t.Fatalf("borrar la historia: %v", err)
	}

	var filas int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM registros_esfuerzo`).Scan(&filas); err != nil || filas != 0 {
		t.Errorf("quedaron %d registros (error %v), se esperaba ninguno", filas, err)
	}
}
