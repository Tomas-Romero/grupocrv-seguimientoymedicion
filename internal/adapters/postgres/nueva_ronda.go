package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// RepositorioPlanningPoker cumple tambien el puerto del caso de uso de US-017 y la lectura de
// las rondas anteriores.
var (
	_ app.RepositorioNuevaRonda    = (*RepositorioPlanningPoker)(nil)
	_ app.LectorDeRondasAnteriores = (*RepositorioPlanningPoker)(nil)
)

// GuardarNuevaRonda abre la ronda siguiente siguiendo los pasos de RN-017-8, en una transaccion
// READ COMMITTED que se detiene en el primer error. Devuelve la sesion en la ronda nueva. Es de
// todo o nada.
func (r *RepositorioPlanningPoker) GuardarNuevaRonda(ctx context.Context, sesionID, integranteID string) (planningpoker.SesionPlanningPoker, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("abrir la transaccion: %w", err)
	}
	// Despues de un Commit exitoso, Rollback no hace nada.
	defer func() { _ = tx.Rollback(ctx) }()

	// Paso 1: el mismo candado que votar, revelar y acordar (RN-017-9): dos pedidos de nueva
	// ronda, o una nueva ronda y un acuerdo, se ordenan y el segundo ve el estado del primero.
	sesion, existe, err := leerSesion(ctx, tx, sesionID, "FOR NO KEY UPDATE OF s")
	if esIDMalFormado(err) || (err == nil && !existe) {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", app.ErrSesionInexistente, sesionID)
	}
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	// Paso 2: los votos de la ronda actual, en una sentencia posterior al candado.
	votos, err := votosDeLaRonda(ctx, tx, sesionID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	// Paso 3: la misma funcion del dominio que el caso de uso, con la sesion bloqueada.
	nueva, err := planningpoker.AbrirNuevaRonda(sesion, integranteID, planningpoker.NuevaVotacion(sesion, votos))
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	// Paso 4: la sesion vuelve a votando en la ronda siguiente. Los votos no se tocan: se
	// guardan por ronda (RN-015-11), asi que la ronda nueva empieza sin votos.
	if _, err := tx.Exec(ctx, `UPDATE sesiones_planning_poker SET estado = 'votando', ronda = $2 WHERE id = $1`, sesionID, nueva.Ronda); err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("abrir la ronda %d de la sesion %s: %w", nueva.Ronda, sesionID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("confirmar la transaccion: %w", err)
	}
	return nueva, nil
}

// RondasAnteriores devuelve las rondas anteriores a la actual de la sesion, de la primera a la
// ultima, con su votacion legible (RN-017-6). Un ID mal formado, o una sesion que no esta, no
// tiene rondas anteriores.
func (r *RepositorioPlanningPoker) RondasAnteriores(ctx context.Context, sesionID string) ([]app.RondaAnterior, error) {
	sesion, existe, err := leerSesion(ctx, r.pool, sesionID, "")
	if esIDMalFormado(err) || (err == nil && !existe) {
		return []app.RondaAnterior{}, nil
	}
	if err != nil {
		return nil, err
	}

	const consulta = `
SELECT ronda, integrante_id::text, valor FROM votos_planning_poker
WHERE sesion_id = $1 AND ronda < $2`
	filas, err := r.pool.Query(ctx, consulta, sesionID, sesion.Ronda)
	if err != nil {
		return nil, fmt.Errorf("leer las rondas anteriores de la sesion %s: %w", sesionID, err)
	}
	porRonda := map[int]map[string]int{}
	var (
		ronda, valor int
		integrante   string
	)
	if _, err := pgx.ForEachRow(filas, []any{&ronda, &integrante, &valor}, func() error {
		if porRonda[ronda] == nil {
			porRonda[ronda] = map[string]int{}
		}
		porRonda[ronda][integrante] = valor
		return nil
	}); err != nil {
		return nil, fmt.Errorf("leer las rondas anteriores de la sesion %s: %w", sesionID, err)
	}

	rondas := make([]app.RondaAnterior, 0, sesion.Ronda-1)
	for n := 1; n < sesion.Ronda; n++ {
		rondas = append(rondas, app.RondaAnterior{Numero: n, Votacion: planningpoker.NuevaVotacionDeRonda(sesion, n, porRonda[n])})
	}
	return rondas, nil
}
