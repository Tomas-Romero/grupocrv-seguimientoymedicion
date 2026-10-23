package planningpoker

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// Errores de votar (spec US-015, seccion 7). Son valores para compararlos con errors.Is,
// tambien cuando vienen unidos con errors.Join (RN-015-8). El de un valor fuera de la
// escala es backlog.ErrEstimacionFueraDeEscala (RN-015-4).
var (
	// ErrVotacionCerrada: la sesion no esta votando (RN-015-3).
	ErrVotacionCerrada = errors.New("la votacion de esta ronda esta cerrada")
	// ErrNoEsParticipante: el integrante no participa de la sesion (RN-015-2).
	ErrNoEsParticipante = errors.New("el integrante no participa de esta sesion de planning poker")
	// ErrVotosOcultos: se pidieron los valores con la sesion votando (RN-015-7). Solo
	// lo provoca un adaptador que los pide antes de tiempo: es un error interno.
	ErrVotosOcultos = errors.New("los votos estan ocultos hasta que se revelen")
)

// VerificarVoto comprueba las reglas de un voto (RN-015-2 a RN-015-4): la sesion esta
// votando, el integrante participa y el valor es de la escala de US-013, que se valida
// con backlog.NuevosStoryPoints sin copiarla.
//
// RN-015-8: no corta en el primer error. Devuelve todos unidos con errors.Join, en orden
// fijo: votacion cerrada, no es participante y valor fuera de la escala. La usa el caso
// de uso con los datos que leyo y la repite el repositorio con la sesion bloqueada
// (RN-015-9).
func VerificarVoto(s SesionPlanningPoker, integranteID string, valor int) error {
	var errs []error
	if s.Estado != EstadoVotando {
		errs = append(errs, ErrVotacionCerrada)
	}
	if !slices.Contains(s.Participantes, integranteID) {
		errs = append(errs, ErrNoEsParticipante)
	}
	if _, err := backlog.NuevosStoryPoints(valor); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// Votacion es la votacion de la ronda actual de una sesion. Garantiza el voto oculto
// (RN-015-7): no exporta sus campos, y mientras la sesion esta votando solo dice quien
// voto y quien falta. Los valores recien se pueden leer cuando la sesion deja de estar
// votando (US-016). Tampoco los muestra al imprimirla.
type Votacion struct {
	estado        Estado
	participantes []string
	votos         map[string]int
}

// NuevaVotacion arma la votacion de la sesion con los votos de su ronda actual, por
// integrante. Se queda con su propia copia de los participantes y de los votos, y no
// cuenta el voto de quien no participa.
func NuevaVotacion(s SesionPlanningPoker, votos map[string]int) Votacion {
	v := Votacion{estado: s.Estado, participantes: slices.Clone(s.Participantes), votos: map[string]int{}}
	for integrante, valor := range votos {
		if slices.Contains(v.participantes, integrante) {
			v.votos[integrante] = valor
		}
	}
	return v
}

// Votaron devuelve los participantes que votaron, en el orden de la sesion.
func (v Votacion) Votaron() []string {
	votaron := []string{}
	for _, p := range v.participantes {
		if _, ok := v.votos[p]; ok {
			votaron = append(votaron, p)
		}
	}
	return votaron
}

// Faltan devuelve los participantes que todavia no votaron, en el orden de la sesion.
func (v Votacion) Faltan() []string {
	faltan := []string{}
	for _, p := range v.participantes {
		if _, ok := v.votos[p]; !ok {
			faltan = append(faltan, p)
		}
	}
	return faltan
}

// Valores devuelve el voto de cada participante que voto. Con la sesion votando no
// devuelve ninguno y responde ErrVotosOcultos (RN-015-7).
func (v Votacion) Valores() (map[string]int, error) {
	if v.estado == EstadoVotando {
		return nil, ErrVotosOcultos
	}
	return maps.Clone(v.votos), nil
}

// String resume la votacion sin los valores, para que imprimirla (por ejemplo, en un log)
// no muestre los votos.
func (v Votacion) String() string {
	return fmt.Sprintf("votacion %s: votaron %d de %d", v.estado, len(v.Votaron()), len(v.participantes))
}

// GoString es como String: con %#v tampoco se ven los votos.
func (v Votacion) GoString() string {
	return v.String()
}
