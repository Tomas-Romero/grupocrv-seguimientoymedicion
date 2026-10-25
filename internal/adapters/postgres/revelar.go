package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// RepositorioPlanningPoker cumple tambien el puerto del caso de uso de US-016 (ObtenerSesion
// y ExisteIntegranteEnProyecto son los de US-015).
var _ app.RepositorioRevelar = (*RepositorioPlanningPoker)(nil)

// VotacionDeLaRonda devuelve la votacion de la ronda actual de la sesion. Mientras se vota no deja
// leer los valores (RN-015-7): el puerto nunca los entrega en crudo. Un ID mal formado, o una sesion
// que no esta, no tiene votos: devuelve una votacion vacia.
func (r *RepositorioPlanningPoker) VotacionDeLaRonda(ctx context.Context, sesionID string) (planningpoker.Votacion, error) {
	sesion, existe, err := leerSesion(ctx, r.pool, sesionID, "")
	if esIDMalFormado(err) || (err == nil && !existe) {
		return planningpoker.Votacion{}, nil
	}
	if err != nil {
		return planningpoker.Votacion{}, err
	}
	votos, err := votosDeLaRonda(ctx, r.pool, sesionID)
	if err != nil {
		return planningpoker.Votacion{}, err
	}
	return planningpoker.NuevaVotacion(sesion, votos), nil
}

// votosDeLaRonda lee los votos de la ronda actual de la sesion con el pool o dentro de una
// transaccion.
func votosDeLaRonda(ctx context.Context, q lectorDeFilas, sesionID string) (map[string]int, error) {
	const consulta = `
SELECT v.integrante_id::text, v.valor
FROM votos_planning_poker v JOIN sesiones_planning_poker s ON s.id = v.sesion_id AND s.ronda = v.ronda
WHERE v.sesion_id = $1`
	filas, err := q.Query(ctx, consulta, sesionID)
	if err != nil {
		return nil, fmt.Errorf("leer los votos de la sesion %s: %w", sesionID, err)
	}
	votos := map[string]int{}
	var (
		integrante string
		valor      int
	)
	if _, err := pgx.ForEachRow(filas, []any{&integrante, &valor}, func() error {
		votos[integrante] = valor
		return nil
	}); err != nil {
		return nil, fmt.Errorf("leer los votos de la sesion %s: %w", sesionID, err)
	}
	return votos, nil
}

// GuardarRevelado revela la sesion siguiendo los pasos de RN-016-8, en una transaccion READ
// COMMITTED que se detiene en el primer error. Devuelve la revelacion. Es de todo o nada.
func (r *RepositorioPlanningPoker) GuardarRevelado(ctx context.Context, sesionID, integranteID string) (planningpoker.Revelacion, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return planningpoker.Revelacion{}, fmt.Errorf("abrir la transaccion: %w", err)
	}
	// Despues de un Commit exitoso, Rollback no hace nada.
	defer func() { _ = tx.Rollback(ctx) }()

	// Paso 1: se bloquea la fila de la sesion (solo esa) y se leen su estado, su ronda y sus
	// participantes. FOR NO KEY UPDATE ordena el revelado con los votos de US-015 (un voto
	// que llega antes entra, uno que llega despues ve la sesion revelada, CL-016-10) y dos
	// revelados entre si (el segundo ve la sesion revelada, CL-016-9). No bloquea a quien
	// inserta filas que apuntan a la sesion.
	sesion, existe, err := leerSesion(ctx, tx, sesionID, "FOR NO KEY UPDATE OF s")
	if esIDMalFormado(err) || (err == nil && !existe) {
		return planningpoker.Revelacion{}, fmt.Errorf("%w (id %s)", app.ErrSesionInexistente, sesionID)
	}
	if err != nil {
		return planningpoker.Revelacion{}, err
	}

	// Paso 2: los votos de la ronda actual, en una sentencia posterior al candado.
	votos, err := votosDeLaRonda(ctx, tx, sesionID)
	if err != nil {
		return planningpoker.Revelacion{}, err
	}

	// Paso 3: la misma funcion del dominio que el caso de uso, con la sesion bloqueada.
	revelacion, err := planningpoker.Revelar(sesion, integranteID, planningpoker.NuevaVotacion(sesion, votos))
	if err != nil {
		return planningpoker.Revelacion{}, err
	}

	// Paso 4: la sesion pasa a revelada. Los votos no se tocan.
	if _, err := tx.Exec(ctx, `UPDATE sesiones_planning_poker SET estado = 'revelada' WHERE id = $1`, sesionID); err != nil {
		return planningpoker.Revelacion{}, fmt.Errorf("revelar la sesion %s: %w", sesionID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return planningpoker.Revelacion{}, fmt.Errorf("confirmar la transaccion: %w", err)
	}
	return revelacion, nil
}
