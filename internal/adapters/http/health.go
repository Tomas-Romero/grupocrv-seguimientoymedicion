// Package httpadapter contiene los handlers HTTP. No tiene reglas de negocio:
// traduce peticiones a llamadas y resultados a respuestas.
//
// El paquete se llama httpadapter y no http para no chocar con net/http.
package httpadapter

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Verificador es todo lo que el health check necesita saber de la base de datos.
// El pool de pgx lo cumple tal cual, y en los tests se reemplaza por uno falso.
type Verificador interface {
	Ping(ctx context.Context) error
}

type respuestaHealth struct {
	Estado string `json:"estado"`
}

// Health devuelve 200 si la base responde dentro de espera y 503 en cualquier
// otro caso. El limite de tiempo evita que un health check quede colgado si la
// base no contesta.
func Health(v Verificador, espera time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelar := context.WithTimeout(r.Context(), espera)
		defer cancelar()

		codigo, estado := http.StatusOK, "ok"
		if err := v.Ping(ctx); err != nil {
			codigo, estado = http.StatusServiceUnavailable, "error"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(codigo)
		_ = json.NewEncoder(w).Encode(respuestaHealth{Estado: estado})
	}
}
