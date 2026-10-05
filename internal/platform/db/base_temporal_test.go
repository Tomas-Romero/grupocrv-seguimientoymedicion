package db_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db"
)

// baseTemporal crea una base de datos vacia para un solo test y la borra al
// terminar. Asi los tests de integracion no tocan la base de desarrollo ni la
// del CI, y no dependen del orden en que corren.
//
// Se omite si DATABASE_URL no esta definida, igual que TestConectar_BaseReal.
// La base de DATABASE_URL solo se usa para crear y borrar la temporal: el usuario
// necesita permiso para crear bases (en el compose y en el CI lo tiene).
func baseTemporal(t *testing.T) *pgxpool.Pool {
	t.Helper()

	direccion := os.Getenv("DATABASE_URL")
	if direccion == "" {
		t.Skip("DATABASE_URL no definida: se omite el test de integracion")
	}

	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()

	admin, err := db.Conectar(ctx, direccion)
	if err != nil {
		t.Fatalf("conectar a la base de DATABASE_URL: %v", err)
	}
	t.Cleanup(admin.Close)

	nombre := "prueba_" + sufijoAleatorio(t)
	identificador := pgx.Identifier{nombre}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+identificador); err != nil {
		t.Fatalf("crear la base temporal %s: %v", nombre, err)
	}
	t.Cleanup(func() {
		ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancelar()
		// WITH (FORCE) corta las conexiones que hayan quedado abiertas (Postgres 13+).
		if _, err := admin.Exec(ctx, "DROP DATABASE IF EXISTS "+identificador+" WITH (FORCE)"); err != nil {
			t.Errorf("borrar la base temporal %s: %v", nombre, err)
		}
	})

	pool, err := db.Conectar(ctx, conOtraBase(t, direccion, nombre))
	if err != nil {
		t.Fatalf("conectar a la base temporal %s: %v", nombre, err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// sufijoAleatorio evita que dos tests, o dos corridas en paralelo, usen la misma base.
func sufijoAleatorio(t *testing.T) string {
	t.Helper()
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("generar el nombre de la base temporal: %v", err)
	}
	return hex.EncodeToString(b)
}

// conOtraBase devuelve la misma URL de conexion apuntando a otra base.
func conOtraBase(t *testing.T, direccion, base string) string {
	t.Helper()
	u, err := url.Parse(direccion)
	if err != nil {
		t.Fatalf("leer DATABASE_URL: %v", err)
	}
	u.Path = "/" + base
	return u.String()
}

// existeTabla dice si la tabla existe en el esquema public de la base del pool.
func existeTabla(ctx context.Context, t *testing.T, pool *pgxpool.Pool, tabla string) bool {
	t.Helper()
	var existe bool
	if err := pool.QueryRow(ctx, "SELECT to_regclass('public.' || $1) IS NOT NULL", tabla).Scan(&existe); err != nil {
		t.Fatalf("buscar la tabla %s: %v", tabla, err)
	}
	return existe
}
