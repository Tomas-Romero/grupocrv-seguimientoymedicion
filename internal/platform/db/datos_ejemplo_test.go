package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/migraciones"
)

// baseMigrada es una base temporal con el esquema ya aplicado.
func baseMigrada(ctx context.Context, t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := baseTemporal(t)
	if err := db.Migrar(ctx, pool, migraciones.Archivos); err != nil {
		t.Fatalf("Migrar: %v", err)
	}
	return pool
}

// contarFilas devuelve cuantas filas tiene la tabla.
func contarFilas(ctx context.Context, t *testing.T, pool *pgxpool.Pool, tabla string) int {
	t.Helper()
	var cantidad int
	consulta := "SELECT count(*) FROM " + pgx.Identifier{tabla}.Sanitize()
	if err := pool.QueryRow(ctx, consulta).Scan(&cantidad); err != nil {
		t.Fatalf("contar las filas de %s: %v", tabla, err)
	}
	return cantidad
}

// DEF-002: los datos de ejemplo dejan cargado el proyecto del equipo con sus tres
// integrantes, que es lo que el evaluador espera ver con el perfil full.
func TestCargarDatosEjemplo(t *testing.T) {
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()
	pool := baseMigrada(ctx, t)

	if err := db.CargarDatosEjemplo(ctx, pool, migraciones.DatosEjemplo); err != nil {
		t.Fatalf("CargarDatosEjemplo: %v", err)
	}

	if n := contarFilas(ctx, t, pool, "proyectos"); n != 1 {
		t.Errorf("proyectos = %d, se esperaba 1", n)
	}
	if n := contarFilas(ctx, t, pool, "integrantes"); n != 3 {
		t.Errorf("integrantes = %d, se esperaba 3", n)
	}
}

// DEF-002: el perfil full carga los datos en cada arranque, asi que cargarlos dos
// veces no puede fallar ni duplicar filas.
func TestCargarDatosEjemplo_DosVeces(t *testing.T) {
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()
	pool := baseMigrada(ctx, t)

	for vez := 1; vez <= 2; vez++ {
		if err := db.CargarDatosEjemplo(ctx, pool, migraciones.DatosEjemplo); err != nil {
			t.Fatalf("CargarDatosEjemplo (vez %d): %v", vez, err)
		}
	}

	if n := contarFilas(ctx, t, pool, "proyectos"); n != 1 {
		t.Errorf("proyectos = %d, se esperaba 1", n)
	}
	if n := contarFilas(ctx, t, pool, "integrantes"); n != 3 {
		t.Errorf("integrantes = %d, se esperaba 3", n)
	}
}

// DEF-002: si ya existe un integrante del proyecto de ejemplo con el mismo email
// pero otro id, cargar los datos no falla por la restriccion UNIQUE
// (proyecto_id, email) de integrantes: el ON CONFLICT de los datos va sin columna.
func TestCargarDatosEjemplo_EmailYaUsado(t *testing.T) {
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()
	pool := baseMigrada(ctx, t)

	// Mismo proyecto y mismo email que "Tomas Romero" en los datos de ejemplo,
	// pero con otro id.
	const previo = `
INSERT INTO proyectos (id, nombre, fecha_inicio, fecha_fin)
VALUES ('00000000-0000-0000-0000-000000000001', 'Proyecto cargado antes', '2026-09-14', '2026-11-16');
INSERT INTO integrantes (id, proyecto_id, nombre, email, rol)
VALUES ('00000000-0000-0000-0000-0000000000ff', '00000000-0000-0000-0000-000000000001',
        'Tomas (cargado a mano)', 'tomas@ejemplo.test', 'agile_enabler');`
	if _, err := pool.Exec(ctx, previo); err != nil {
		t.Fatalf("cargar el integrante previo: %v", err)
	}

	if err := db.CargarDatosEjemplo(ctx, pool, migraciones.DatosEjemplo); err != nil {
		t.Fatalf("CargarDatosEjemplo con un email ya usado: %v", err)
	}
}
