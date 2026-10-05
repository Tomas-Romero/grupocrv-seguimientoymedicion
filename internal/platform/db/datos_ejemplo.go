package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CargarDatosEjemplo ejecuta el SQL de los datos de ejemplo sobre la base del
// pool. El SQL tiene que ser idempotente, porque la app lo corre en cada
// arranque del perfil full.
//
// Exec sin argumentos usa el protocolo simple de Postgres, que acepta varias
// sentencias en una sola llamada y las corre como una unica transaccion: o se
// cargan todos los datos o ninguno.
func CargarDatosEjemplo(ctx context.Context, pool *pgxpool.Pool, sentencias string) error {
	if _, err := pool.Exec(ctx, sentencias); err != nil {
		return fmt.Errorf("cargar los datos de ejemplo: %w", err)
	}
	return nil
}
