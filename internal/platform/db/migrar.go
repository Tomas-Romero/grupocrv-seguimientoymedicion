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
//
// Acepta migraciones con numero menor al ultimo aplicado (T-012): cuando dos
// ramas agregan migraciones en paralelo, la de numero menor puede llegar despues
// a una base que ya aplico la otra. Sin esto, goose la ve como faltante y no
// migra nada.
func Migrar(ctx context.Context, pool *pgxpool.Pool, archivos fs.FS) error {
	sqlDB := stdlib.OpenDBFromPool(pool)
	// Cerrar este *sql.DB no cierra el pool: solo suelta lo que goose uso.
	defer func() { _ = sqlDB.Close() }()

	proveedor, err := goose.NewProvider(goose.DialectPostgres, sqlDB, archivos,
		goose.WithAllowOutofOrder(true))
	if err != nil {
		return fmt.Errorf("preparar las migraciones: %w", err)
	}
	if _, err := proveedor.Up(ctx); err != nil {
		return fmt.Errorf("aplicar las migraciones: %w", err)
	}
	return nil
}
