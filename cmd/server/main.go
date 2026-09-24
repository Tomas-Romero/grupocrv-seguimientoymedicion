// Command server levanta la aplicacion web de Software Metrics & Estimation.
//
// Este archivo es el unico lugar donde se arman las dependencias: lee la
// configuracion, construye los adaptadores y se los pasa a los casos de uso.
// No contiene reglas de negocio.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpadapter "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/http"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/config"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db"
)

func main() {
	if err := run(); err != nil {
		slog.Error("el servidor termino con error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Cargar(os.Getenv)
	if err != nil {
		return fmt.Errorf("cargar configuracion: %w", err)
	}

	// Apagado ordenado: al recibir Ctrl+C se dejan terminar las peticiones en curso.
	ctx, detener := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer detener()

	// Si la base no responde al arrancar, el servidor no arranca.
	conexion, cancelarConexion := context.WithTimeout(ctx, 10*time.Second)
	pool, err := db.Conectar(conexion, cfg.DatabaseURL)
	cancelarConexion()
	if err != nil {
		return fmt.Errorf("conectar a la base de datos: %w", err)
	}
	defer pool.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", httpadapter.Health(pool, 2*time.Second))
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<h1>Software Metrics &amp; Estimation</h1><p>Sprint 0 — walking skeleton.</p>"))
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errc := make(chan error, 1)
	go func() {
		slog.Info("servidor escuchando", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- err
		}
	}()

	select {
	case err := <-errc:
		return fmt.Errorf("escuchar en %s: %w", cfg.HTTPAddr, err)
	case <-ctx.Done():
		slog.Info("apagando el servidor")
		cierre, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelar()
		if err := srv.Shutdown(cierre); err != nil {
			return fmt.Errorf("apagar el servidor: %w", err)
		}
		return nil
	}
}
