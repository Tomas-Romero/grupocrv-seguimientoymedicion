package planningpoker

import (
	"errors"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// ErrSesionNoRevelada: la estimacion acordada se registra solo con la sesion revelada, ni
// votando ni ya finalizada (RN-018-2).
var ErrSesionNoRevelada = errors.New("la estimacion acordada se registra con los votos revelados y la sesion sin finalizar")

// Acordar valida que se pueda registrar la estimacion acordada y devuelve una copia de la
// sesion finalizada con ese valor (RN-018-8). historiaCompletada dice si la historia de la
// sesion esta completada.
//
// El valor puede ser cualquiera de la escala de US-013, votado o no, y aunque haya
// divergencia (RN-018-5); se valida con backlog.NuevosStoryPoints (RN-018-4). Una historia
// completada se valida primero y su error va solo (RN-018-6, mismo criterio que RN-013-5).
// Si no, devuelve todos los errores unidos con errors.Join, en orden fijo: sesion no
// revelada, no es participante y valor fuera de la escala (RN-018-7).
//
// La usa el caso de uso con los datos que leyo y la repite el repositorio con la sesion
// bloqueada (RN-018-9). Estimar la historia con el valor lo hace quien guarda: esta funcion
// no la recibe.
func Acordar(s SesionPlanningPoker, historiaCompletada bool, integranteID string, valor int) (SesionPlanningPoker, error) {
	if historiaCompletada {
		return SesionPlanningPoker{}, backlog.ErrItemCompletado
	}
	var errs []error
	if s.Estado != EstadoRevelada {
		errs = append(errs, ErrSesionNoRevelada)
	}
	if !slices.Contains(s.Participantes, integranteID) {
		errs = append(errs, ErrNoEsParticipante)
	}
	puntos, err := backlog.NuevosStoryPoints(valor)
	if err != nil {
		errs = append(errs, err)
	}
	if err := errors.Join(errs...); err != nil {
		return SesionPlanningPoker{}, err
	}

	s.Participantes = slices.Clone(s.Participantes)
	s.Estado = EstadoFinalizada
	s.ValorAcordado = puntos
	return s, nil
}
