package db

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Migrar aplica sobre la base del pool las migraciones de archivos que todavia
// no se aplicaron. Si ya estan todas, no hace nada.
//
// Usa goose como biblioteca con un Provider, que recibe todo por parametro en
// lugar de guardarlo en variables globales. El *sql.DB que necesita goose se arma
// sobre el mismo pool, asi que no abre una segunda conexion.
func Migrar(ctx context.Context, pool *pgxpool.Pool, archivos fs.FS) error {
	sqlDB := stdlib.OpenDBFromPool(pool)
	// Cerrar este *sql.DB no cierra el pool: solo suelta lo que goose uso.
	defer func() { _ = sqlDB.Close() }()

	proveedor, err := goose.NewProvider(goose.DialectPostgres, sqlDB, archivos)
	if err != nil {
		return fmt.Errorf("preparar las migraciones: %w", err)
	}
	if _, err := proveedor.Up(ctx); err != nil {
		return fmt.Errorf("aplicar las migraciones: %w", err)
	}
	return nil
}
