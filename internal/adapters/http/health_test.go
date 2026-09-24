package httpadapter_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpadapter "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/http"
)

// verificadorFalso simula la base de datos sin levantar Postgres.
type verificadorFalso struct {
	err     error
	bloquea bool // si es true, espera hasta que el contexto se cancele
}

func (v verificadorFalso) Ping(ctx context.Context) error {
	if v.bloquea {
		<-ctx.Done()
		return ctx.Err()
	}
	return v.err
}

// T-001: /health informa 200 si la base responde y 503 si falla o tarda demasiado.
func TestHealth(t *testing.T) {
	casos := []struct {
		nombre         string
		verificador    verificadorFalso
		codigoEsperado int
		estadoEsperado string
	}{
		{
			nombre:         "la base responde",
			verificador:    verificadorFalso{},
			codigoEsperado: http.StatusOK,
			estadoEsperado: "ok",
		},
		{
			nombre:         "la base devuelve un error",
			verificador:    verificadorFalso{err: errors.New("conexion rechazada")},
			codigoEsperado: http.StatusServiceUnavailable,
			estadoEsperado: "error",
		},
		{
			nombre:         "la base no responde a tiempo",
			verificador:    verificadorFalso{bloquea: true},
			codigoEsperado: http.StatusServiceUnavailable,
			estadoEsperado: "error",
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			handler := httpadapter.Health(c.verificador, 50*time.Millisecond)
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != c.codigoEsperado {
				t.Fatalf("codigo = %d, se esperaba %d", rec.Code, c.codigoEsperado)
			}
			if tipo := rec.Header().Get("Content-Type"); tipo != "application/json" {
				t.Fatalf("Content-Type = %q, se esperaba application/json", tipo)
			}
			var cuerpo struct {
				Estado string `json:"estado"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&cuerpo); err != nil {
				t.Fatalf("el cuerpo no es JSON valido: %v", err)
			}
			if cuerpo.Estado != c.estadoEsperado {
				t.Fatalf("estado = %q, se esperaba %q", cuerpo.Estado, c.estadoEsperado)
			}
		})
	}
}
