// Package db abre la conexion con PostgreSQL.
//
// Devuelve un pool de pgx: un conjunto de conexiones reutilizables que es seguro
// usar desde varias goroutines, que es lo que necesita un servidor HTTP.
package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrURLVacia se devuelve cuando no se le pasa ninguna URL de conexion.
var ErrURLVacia = errors.New("la URL de la base de datos esta vacia")

// Conectar crea el pool y comprueba con un Ping que la base responde.
//
// pgxpool.New no abre conexiones hasta que se usa el pool. Sin el Ping, una base
// caida se descubriria en la primera consulta de un usuario y no al arrancar.
func Conectar(ctx context.Context, url string) (*pgxpool.Pool, error) {
	if strings.TrimSpace(url) == "" {
		return nil, ErrURLVacia
	}

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("crear el pool de conexiones: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("conectar a la base: %w", err)
	}
	return pool, nil
}
