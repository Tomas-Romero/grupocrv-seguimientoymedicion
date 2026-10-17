package memoria

import (
	"context"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
)

// Base cumple el puerto del caso de uso de US-003.
var _ app.RepositorioRegistrarIntegrante = (*Base)(nil)

// RegistrarIntegrante es de US-003.
//
// ESQUELETO DEL CICLO RED (US-003): todavia no guarda nada. La implementacion
// real llega en el commit GREEN.
func (b *Base) RegistrarIntegrante(_ context.Context, i integrante.Integrante) (integrante.Integrante, error) {
	return i, nil
}

// Integrantes es de US-003.
//
// ESQUELETO DEL CICLO RED (US-003): todavia no devuelve nada.
func (b *Base) Integrantes(_ string) []integrante.Integrante {
	return nil
}
