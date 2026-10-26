package memoria

import (
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// Este archivo solo se compila en los tests del paquete: da acceso a lo que las historias
// todavia no exponen (abrir una ronda es US-017, y completar una historia sin pasar por un
// sprint no lo hace ninguna), sin agregarle al codigo de produccion una forma de leer los
// votos.

// VotoGuardado devuelve el valor guardado del voto, si lo hay.
func (b *Base) VotoGuardado(sesionID string, ronda int, integranteID string) (int, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	valor, ok := b.votos[claveDeVoto{sesionID: sesionID, ronda: ronda, integranteID: integranteID}]
	return valor, ok
}

// CambiarSesion cambia el estado y la ronda de la sesion guardada.
func (b *Base) CambiarSesion(sesionID string, estado planningpoker.Estado, ronda int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range b.sesiones {
		if b.sesiones[i].ID == sesionID {
			b.sesiones[i].Estado = estado
			b.sesiones[i].Ronda = ronda
		}
	}
}

// CompletarItem deja la historia completada, como si se hubiera marcado en un sprint
// (US-010), para armar el caso de una historia que se completa con su sesion abierta.
func (b *Base) CompletarItem(itemID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if item := b.itemGuardado(itemID); item != nil {
		item.Estado = backlog.EstadoCompletado
	}
}
