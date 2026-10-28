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
// Recibe la votacion de la ronda y no los votos en crudo (RN-015-7): ningun puerto entrega los
// valores mientras se vota. Revelar los lee de la votacion al revelar, dentro del dominio, que es
// el unico lugar donde la votacion los muestra antes de que la sesion este revelada.
//
// RN-016-10: no corta en el primer error. Devuelve todos unidos con errors.Join, en orden
// fijo: votacion cerrada, no es participante y sin votos. La usa el caso de uso con los
// datos que leyo y la repite el repositorio con la sesion bloqueada (RN-016-8). El cambio de
// estado a revelada lo guarda quien la llama: esta funcion no modifica la sesion ni la votacion.
func Revelar(s SesionPlanningPoker, integranteID string, v Votacion) (Revelacion, error) {
	var errs []error
	if s.Estado != EstadoVotando {
		errs = append(errs, ErrVotacionCerrada)
	}
	if !slices.Contains(s.Participantes, integranteID) {
		errs = append(errs, ErrNoEsParticipante)
	}
	if len(v.Votaron()) == 0 {
		errs = append(errs, ErrSinVotos)
	}
	if err := errors.Join(errs...); err != nil {
		return Revelacion{}, err
	}

	r := resultado(v)
	return Revelacion{
		SesionID: s.ID, Ronda: s.Ronda,
		Votos: r.Votos, Faltaron: r.Faltaron, Divergencia: r.Divergencia, Minimo: r.Minimo, Maximo: r.Maximo,
	}, nil
}

// Resultado es el resultado de una ronda: sus votos, quienes faltaron, la divergencia y los
// extremos, con el mismo significado que en Revelacion. US-017 lo usa para decidir si se abre
// otra ronda y para leer las rondas anteriores.
type Resultado struct {
	Votos       []Voto
	Faltaron    []string
	Divergencia bool
	Minimo      int
	Maximo      int
}

// resultado arma el resultado de la votacion leyendo sus valores sin preguntar si se pueden
// leer: quien la llama ya lo decidio (Revelar, porque es quien revela).
func resultado(v Votacion) Resultado {
	r := Resultado{Votos: []Voto{}, Faltaron: v.Faltan()}
	for _, p := range v.Votaron() {
		valor := v.votos[p]
		r.Votos = append(r.Votos, Voto{ParticipanteID: p, Valor: valor})
		if len(r.Votos) == 1 || valor < r.Minimo {
			r.Minimo = valor
		}
		if len(r.Votos) == 1 || valor > r.Maximo {
			r.Maximo = valor
		}
	}
	r.Divergencia = r.Minimo != r.Maximo
	return r
}
