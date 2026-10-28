package planningpoker

import (
	"errors"
	"slices"
)

// Errores de abrir una nueva ronda (spec US-017, seccion 7).
var (
	// ErrNuevaRondaSinRevelar: la nueva ronda se abre solo con la sesion revelada, ni votando
	// ni ya finalizada (RN-017-2).
	ErrNuevaRondaSinRevelar = errors.New("la nueva ronda se abre con los votos revelados y la sesion sin finalizar")
	// ErrSinDivergencia: los votos de la ronda coinciden, asi que no hace falta otra ronda
	// (RN-017-4).
	ErrSinDivergencia = errors.New("no hay divergencia: registren la estimacion acordada")
)

// AbrirNuevaRonda valida que se pueda abrir la ronda siguiente y devuelve una copia de la sesion
// votando en esa ronda, con los mismos participantes (RN-017-5). v es la votacion de la ronda
// actual.
//
// La divergencia (RN-016-6) se evalua solo con la sesion revelada: votando, los valores estan
// ocultos (RN-015-7). Devuelve todos los errores unidos con errors.Join, en orden fijo: sesion
// no revelada, no es participante y sin divergencia (RN-017-7). La usa el caso de uso con los
// datos que leyo y la repite el repositorio con la sesion bloqueada (RN-017-8). Los votos no
// se tocan: la ronda nueva empieza sin votos porque los votos se guardan por ronda.
func AbrirNuevaRonda(s SesionPlanningPoker, integranteID string, v Votacion) (SesionPlanningPoker, error) {
	var errs []error
	revelada := s.Estado == EstadoRevelada
	if !revelada {
		errs = append(errs, ErrNuevaRondaSinRevelar)
	}
	if !slices.Contains(s.Participantes, integranteID) {
		errs = append(errs, ErrNoEsParticipante)
	}
	if revelada && !resultado(v).Divergencia {
		errs = append(errs, ErrSinDivergencia)
	}
	if err := errors.Join(errs...); err != nil {
		return SesionPlanningPoker{}, err
	}

	s.Participantes = slices.Clone(s.Participantes)
	s.Estado = EstadoVotando
	s.Ronda++
	return s, nil
}

// ResultadoDeLaRonda arma el resultado de la votacion de una ronda. Solo se puede si sus
// valores se pueden leer (RN-017-6): si no, responde ErrVotosOcultos.
func ResultadoDeLaRonda(v Votacion) (Resultado, error) {
	if !v.legible() {
		return Resultado{}, ErrVotosOcultos
	}
	return resultado(v), nil
}
