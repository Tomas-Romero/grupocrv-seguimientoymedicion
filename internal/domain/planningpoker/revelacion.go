package planningpoker

import (
	"errors"
	"slices"
)

// ErrSinVotos: no hay ningun voto en la ronda, asi que no hay nada que revelar ni que
// comparar (RN-016-4).
var ErrSinVotos = errors.New("no hay votos para revelar en esta ronda")

// Voto es el voto de un participante, ya revelado.
type Voto struct {
	ParticipanteID string
	Valor          int
}

// Revelacion es lo que sale de revelar la ronda actual de una sesion (spec US-016,
// seccion 3).
type Revelacion struct {
	SesionID string
	Ronda    int
	// Votos son los participantes que votaron, en el orden de la sesion.
	Votos []Voto
	// Faltaron son los participantes que no votaron, en ese mismo orden (RN-016-5).
	Faltaron []string
	// Divergencia dice si los valores de Votos no son todos iguales (RN-016-6).
	Divergencia bool
	// Minimo y Maximo son el menor y el mayor de los valores votados.
	Minimo int
	Maximo int
}

// Revelar valida que se pueda revelar la ronda actual de la sesion y arma la revelacion
// (RN-016-2 a RN-016-6). Quien revela es un participante, la sesion tiene que estar votando
// y tiene que haber al menos un voto; pueden faltar votos de otros participantes.
//
// RN-016-10: no corta en el primer error. Devuelve todos unidos con errors.Join, en orden
// fijo: votacion cerrada, no es participante y sin votos. La usa el caso de uso con los
// datos que leyo y la repite el repositorio con la sesion bloqueada (RN-016-8). El cambio de
// estado a revelada lo guarda quien la llama: esta funcion no modifica la sesion ni los votos.
//
// Los valores salen del dominio solo por Votacion.Valores (RN-015-7): la revelacion los lee
// de una votacion de la sesion ya revelada.
func Revelar(s SesionPlanningPoker, integranteID string, votos map[string]int) (Revelacion, error) {
	var errs []error
	if s.Estado != EstadoVotando {
		errs = append(errs, ErrVotacionCerrada)
	}
	if !slices.Contains(s.Participantes, integranteID) {
		errs = append(errs, ErrNoEsParticipante)
	}
	if len(NuevaVotacion(s, votos).Votaron()) == 0 {
		errs = append(errs, ErrSinVotos)
	}
	if err := errors.Join(errs...); err != nil {
		return Revelacion{}, err
	}

	revelada := s
	revelada.Estado = EstadoRevelada
	votacion := NuevaVotacion(revelada, votos)
	// La sesion ya figura revelada: Valores no se niega.
	valores, _ := votacion.Valores()

	r := Revelacion{SesionID: s.ID, Ronda: s.Ronda, Votos: []Voto{}, Faltaron: votacion.Faltan()}
	for _, p := range votacion.Votaron() {
		v := valores[p]
		r.Votos = append(r.Votos, Voto{ParticipanteID: p, Valor: v})
		if len(r.Votos) == 1 || v < r.Minimo {
			r.Minimo = v
		}
		if len(r.Votos) == 1 || v > r.Maximo {
			r.Maximo = v
		}
	}
	r.Divergencia = r.Minimo != r.Maximo
	return r, nil
}
