package features_test

import (
	"flag"
	"testing"
)

// T-006: los flags de godog tienen que estar registrados en el paquete flag
// estandar, que es el que parsea `go test`. Si quedan en otro lado, usar
// `go test ./features/... -godog.tags=@US-026` falla con "flag provided but not
// defined" y no se puede correr el escenario de una sola historia.
func TestFlagsDeGodogRegistrados(t *testing.T) {
	for _, nombre := range []string{"godog.tags", "godog.format"} {
		if flag.Lookup(nombre) == nil {
			t.Errorf("el flag -%s no esta registrado en el paquete flag", nombre)
		}
	}
}
