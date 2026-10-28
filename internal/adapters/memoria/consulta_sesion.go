package memoria

import (
	"context"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
)

// Base cumple el puerto de la lectura de la pantalla de la sesion (T-020).
var _ app.RepositorioConsultaDeSesion = (*Base)(nil)

// SesionesAbiertas devuelve, para cada historia del proyecto con una sesion abierta, el ID de esa
// sesion. Un proyecto que no esta no tiene sesiones.
func (b *Base) SesionesAbiertas(_ context.Context, proyectoID string) (map[string]string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	abiertas := map[string]string{}
	for _, s := range b.sesiones {
		if s.ProyectoID == proyectoID && s.Estado.Abierta() {
			abiertas[s.ItemID] = s.ID
		}
	}
	return abiertas, nil
}
