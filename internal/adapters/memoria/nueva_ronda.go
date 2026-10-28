package memoria

import (
	"context"
	"fmt"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// Base cumple el puerto del caso de uso de US-017 y la lectura de las rondas anteriores.
var (
	_ app.RepositorioNuevaRonda    = (*Base)(nil)
	_ app.LectorDeRondasAnteriores = (*Base)(nil)
)

// GuardarNuevaRonda replica lo que hace la persistencia (RN-017-8), con mu tomado en lugar del
// candado sobre la sesion: verifica que la sesion este, vuelve a verificar con
// planningpoker.AbrirNuevaRonda con lo guardado y, si puede, la pasa a votando en la ronda
// siguiente. No toca ningun voto. Es de todo o nada.
func (b *Base) GuardarNuevaRonda(_ context.Context, sesionID, integranteID string) (planningpoker.SesionPlanningPoker, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s := b.sesionGuardada(sesionID)
	if s == nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", app.ErrSesionInexistente, sesionID)
	}
	nueva, err := planningpoker.AbrirNuevaRonda(*s, integranteID, b.votacion(*s))
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}
	s.Estado, s.Ronda = nueva.Estado, nueva.Ronda
	return nueva, nil
}

// RondasAnteriores devuelve las rondas anteriores a la actual de la sesion, de la primera a la
// ultima, con su votacion legible (RN-017-6). Una sesion que no esta no tiene rondas.
func (b *Base) RondasAnteriores(_ context.Context, sesionID string) ([]app.RondaAnterior, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	s := b.sesionGuardada(sesionID)
	if s == nil {
		return []app.RondaAnterior{}, nil
	}
	rondas := []app.RondaAnterior{}
	for ronda := 1; ronda < s.Ronda; ronda++ {
		rondas = append(rondas, app.RondaAnterior{Numero: ronda, Votacion: b.votacionDeRonda(*s, ronda)})
	}
	return rondas, nil
}

// votacionDeRonda arma la votacion de una ronda de la sesion con los votos guardados. Se llama
// con mu tomado.
func (b *Base) votacionDeRonda(s planningpoker.SesionPlanningPoker, ronda int) planningpoker.Votacion {
	votos := map[string]int{}
	for clave, valor := range b.votos {
		if clave.sesionID == s.ID && clave.ronda == ronda {
			votos[clave.integranteID] = valor
		}
	}
	s.Participantes = slices.Clone(s.Participantes)
	return planningpoker.NuevaVotacionDeRonda(s, ronda, votos)
}
