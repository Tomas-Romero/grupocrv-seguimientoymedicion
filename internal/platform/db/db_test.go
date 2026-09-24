package db_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db"
)

// T-001: Conectar rechaza URLs invalidas y servidores inalcanzables con un error
// envuelto. No necesita una base levantada: todos los casos tienen que fallar.
func TestConectar_Errores(t *testing.T) {
	casos := []struct {
		nombre   string
		url      string
		esperado error // nil: alcanza con que devuelva algun error
	}{
		{
			nombre:   "URL vacia",
			url:      "",
			esperado: db.ErrURLVacia,
		},
		{
			nombre: "URL con formato invalido",
			url:    "postgres://metrics:metrics@localhost:puerto_no_numerico/metrics",
		},
		{
			nombre: "servidor inalcanzable",
			url:    "postgres://metrics:metrics@127.0.0.1:1/metrics?sslmode=disable&connect_timeout=1",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelar()

			pool, err := db.Conectar(ctx, c.url)
			if err == nil {
				pool.Close()
				t.Fatal("se esperaba un error y no hubo ninguno")
			}
			if c.esperado != nil && !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.esperado)
			}
		})
	}
}

// T-001: contra una base real, Conectar devuelve un pool que responde al Ping.
// Se omite si DATABASE_URL no esta definida, asi `go test` funciona sin Docker.
func TestConectar_BaseReal(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL no definida: se omite el test de integracion")
	}

	ctx, cancelar := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelar()

	pool, err := db.Conectar(ctx, url)
	if err != nil {
		t.Fatalf("Conectar: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}
