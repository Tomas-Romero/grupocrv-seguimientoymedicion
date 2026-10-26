package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// RepositorioPlanningPoker cumple tambien el puerto del caso de uso de US-018 (ObtenerSesion
// y ExisteIntegranteEnProyecto son los de US-015, y ObtenerItem el de US-013).
var _ app.RepositorioAcuerdo = (*RepositorioPlanningPoker)(nil)

// GuardarAcuerdo registra la estimacion acordada siguiendo los pasos de RN-018-9, en una
// sola transaccion READ COMMITTED que se detiene en el primer error: si algo falla, la
// sesion sigue revelada y la historia queda como estaba. Devuelve la sesion finalizada.
func (r *RepositorioPlanningPoker) GuardarAcuerdo(ctx context.Context, sesionID, integranteID string, valor int) (planningpoker.SesionPlanningPoker, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("abrir la transaccion: %w", err)
	}
	// Despues de un Commit exitoso, Rollback no hace nada.
	defer func() { _ = tx.Rollback(ctx) }()

	// Paso 1: se bloquea la fila de la sesion (solo esa) y se leen su estado, su historia y
	// sus participantes. FOR NO KEY UPDATE ordena dos registros (el segundo ve la sesion
	// finalizada, CL-018-11) y el registro con la ronda nueva de US-017, con el mismo
	// criterio que RN-015-10 y RN-016-9.
	sesion, existe, err := leerSesion(ctx, tx, sesionID, "FOR NO KEY UPDATE OF s")
	if esIDMalFormado(err) || (err == nil && !existe) {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", app.ErrSesionInexistente, sesionID)
	}
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	// Paso 2: el estado de la historia, en una sentencia posterior al candado.
	var estado string
	err = tx.QueryRow(ctx, `SELECT estado FROM items_backlog WHERE id = $1`, sesion.ItemID).Scan(&estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", app.ErrItemInexistente, sesion.ItemID)
	}
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("leer el item %s: %w", sesion.ItemID, err)
	}

	// Paso 3: la misma funcion del dominio que el caso de uso, con la sesion bloqueada.
	finalizada, err := planningpoker.Acordar(sesion, backlog.Estado(estado) == backlog.EstadoCompletado, integranteID, valor)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	// Paso 4: la historia se estima con la sentencia de US-013, condicionada a que no este
	// completada: si se completo despues del paso 2 (CL-018-10), devuelve ErrItemCompletado y
	// la transaccion se revierte entera.
	if err := guardarEstimacion(ctx, tx, sesion.ItemID, finalizada.ValorAcordado); err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	// Paso 5: la sesion pasa a finalizada con el valor acordado.
	acordado, _ := finalizada.ValorAcordado.Valor()
	const finalizar = `UPDATE sesiones_planning_poker SET estado = 'finalizada', valor_acordado = $2 WHERE id = $1`
	if _, err := tx.Exec(ctx, finalizar, sesionID, acordado); err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("finalizar la sesion %s: %w", sesionID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("confirmar la transaccion: %w", err)
	}
	return finalizada, nil
}
