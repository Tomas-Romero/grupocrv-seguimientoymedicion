package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/migraciones"
)

// DEF-002: migrar una base vacia deja creado el esquema, con la tabla proyectos.
func TestMigrar_BaseVacia(t *testing.T) {
	pool := dbprueba.BaseTemporal(t)
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()

	if err := db.Migrar(ctx, pool, migraciones.Archivos); err != nil {
		t.Fatalf("Migrar: %v", err)
	}

	if !dbprueba.ExisteTabla(ctx, t, pool, "proyectos") {
		t.Fatal("despues de migrar no existe la tabla proyectos")
	}
}

// DEF-002: la app migra en cada arranque, asi que migrar una base ya migrada no
// puede fallar ni volver a aplicar nada.
func TestMigrar_DosVeces(t *testing.T) {
	pool := dbprueba.BaseTemporal(t)
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()

	for vez := 1; vez <= 2; vez++ {
		if err := db.Migrar(ctx, pool, migraciones.Archivos); err != nil {
			t.Fatalf("Migrar (vez %d): %v", vez, err)
		}
	}

	var aplicaciones int
	consulta := "SELECT count(*) FROM goose_db_version WHERE version_id = 1"
	if err := pool.QueryRow(ctx, consulta).Scan(&aplicaciones); err != nil {
		t.Fatalf("leer goose_db_version: %v", err)
	}
	if aplicaciones != 1 {
		t.Fatalf("la migracion 1 figura %d veces en goose_db_version, se esperaba 1", aplicaciones)
	}
}
