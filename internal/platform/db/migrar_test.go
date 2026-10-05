package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/migraciones"
)

// DEF-002: migrar una base vacia deja creado el esquema, con la tabla proyectos.
func TestMigrar_BaseVacia(t *testing.T) {
	pool := baseTemporal(t)
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()

	if err := db.Migrar(ctx, pool, migraciones.Archivos); err != nil {
		t.Fatalf("Migrar: %v", err)
	}

	if !existeTabla(ctx, t, pool, "proyectos") {
		t.Fatal("despues de migrar no existe la tabla proyectos")
	}
}
