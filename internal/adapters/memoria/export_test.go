package memoria

import "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"

// Este archivo solo se compila en los tests del paquete: da acceso a lo que las historias
// todavia no exponen (revelar es US-016 y abrir una ronda, US-017), sin agregarle al
// codigo de produccion una forma de leer los votos.

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
