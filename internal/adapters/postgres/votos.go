package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// RepositorioPlanningPoker cumple tambien el puerto del caso de uso de US-015.
var _ app.RepositorioVotos = (*RepositorioPlanningPoker)(nil)

// ObtenerSesion lee la sesion con el proyecto de su historia y sus participantes en orden
// de alta. Un ID que ninguna sesion tiene, o que no es un UUID, es una sesion inexistente
// y no un error.
func (r *RepositorioPlanningPoker) ObtenerSesion(ctx context.Context, sesionID string) (planningpoker.SesionPlanningPoker, bool, error) {
	sesion, existe, err := leerSesion(ctx, r.pool, sesionID, "")
	if esIDMalFormado(err) {
		return planningpoker.SesionPlanningPoker{}, false, nil
	}
	return sesion, existe, err
}

// ExisteIntegranteEnProyecto dice si el integrante existe y es de ese proyecto. Es el
// chequeo de US-019: el repositorio lo delega en el de esfuerzo en lugar de repetirlo.
func (r *RepositorioPlanningPoker) ExisteIntegranteEnProyecto(ctx context.Context, proyectoID, integranteID string) (bool, error) {
	return NuevoRepositorioEsfuerzo(r.pool).ExisteIntegranteEnProyecto(ctx, proyectoID, integranteID)
}

// leerSesion lee la sesion y sus participantes con el pool o dentro de una transaccion.
// bloqueo, si no esta vacio, se agrega a la lectura de la fila de la sesion (por ejemplo,
// "FOR NO KEY UPDATE OF s").
func leerSesion(ctx context.Context, q interface {
	consultor
	lectorDeFilas
}, sesionID, bloqueo string) (planningpoker.SesionPlanningPoker, bool, error) {
	consulta := `
SELECT s.id::text, s.item_id::text, i.proyecto_id::text, s.estado, s.ronda
FROM sesiones_planning_poker s JOIN items_backlog i ON i.id = s.item_id
WHERE s.id = $1 ` + bloqueo
	var (
		s      planningpoker.SesionPlanningPoker
		estado string
	)
	err := q.QueryRow(ctx, consulta, sesionID).Scan(&s.ID, &s.ItemID, &s.ProyectoID, &estado, &s.Ronda)
	if errors.Is(err, pgx.ErrNoRows) {
		return planningpoker.SesionPlanningPoker{}, false, nil
	}
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, false, fmt.Errorf("leer la sesion %s: %w", sesionID, err)
	}
	s.Estado = planningpoker.Estado(estado)

	const participantes = `
SELECT p.integrante_id::text
FROM participantes_planning_poker p JOIN integrantes i ON i.id = p.integrante_id
WHERE p.sesion_id = $1 ORDER BY i.creado_en, i.id`
	filas, err := q.Query(ctx, participantes, sesionID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, false, fmt.Errorf("leer los participantes de la sesion %s: %w", sesionID, err)
	}
	if s.Participantes, err = pgx.CollectRows(filas, pgx.RowTo[string]); err != nil {
		return planningpoker.SesionPlanningPoker{}, false, fmt.Errorf("leer los participantes de la sesion %s: %w", sesionID, err)
	}
	return s, true, nil
}

// RegistrarVoto guarda el voto siguiendo los pasos de RN-015-9, en una transaccion READ
// COMMITTED que se detiene en el primer error. Devuelve la votacion de la ronda actual,
// que mientras la sesion esta votando no deja leer los valores (RN-015-7). Es de todo o
// nada.
func (r *RepositorioPlanningPoker) RegistrarVoto(ctx context.Context, sesionID, integranteID string, valor int) (planningpoker.Votacion, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return planningpoker.Votacion{}, fmt.Errorf("abrir la transaccion: %w", err)
	}
	// Despues de un Commit exitoso, Rollback no hace nada.
	defer func() { _ = tx.Rollback(ctx) }()

	// Pasos 1 y 2: se bloquea la fila de la sesion (solo esa, no la de la historia) y se
	// leen su estado, su ronda y sus participantes. FOR NO KEY UPDATE ordena dos votos del
	// mismo participante (cuenta el ultimo que confirma, CL-015-8) y el voto con el
	// revelado de US-016, que cambia esa fila: un voto que llega despues ve la sesion
	// revelada (CL-015-9). No bloquea a quien inserta filas que apuntan a la sesion.
	sesion, existe, err := leerSesion(ctx, tx, sesionID, "FOR NO KEY UPDATE OF s")
	if esIDMalFormado(err) || (err == nil && !existe) {
		return planningpoker.Votacion{}, fmt.Errorf("%w (id %s)", app.ErrSesionInexistente, sesionID)
	}
	if err != nil {
		return planningpoker.Votacion{}, err
	}

	// Paso 3: las mismas reglas que el caso de uso, con la sesion bloqueada.
	if err := planningpoker.VerificarVoto(sesion, integranteID, valor); err != nil {
		return planningpoker.Votacion{}, err
	}

	// Paso 4: el voto de la ronda actual; si el participante ya habia votado en ella, se
	// reemplaza el valor (RN-015-5).
	const votar = `
INSERT INTO votos_planning_poker (sesion_id, ronda, integrante_id, valor) VALUES ($1, $2, $3, $4)
ON CONFLICT (sesion_id, ronda, integrante_id) DO UPDATE SET valor = EXCLUDED.valor, votado_en = now()`
	if _, err := tx.Exec(ctx, votar, sesionID, sesion.Ronda, integranteID, valor); err != nil {
		return planningpoker.Votacion{}, fmt.Errorf("guardar el voto: %w", err)
	}

	// La votacion de la ronda, con lo que quedo guardado. Los valores entran al dominio y
	// la votacion no los expone mientras se vota.
	const votosDeLaRonda = `SELECT integrante_id::text, valor FROM votos_planning_poker WHERE sesion_id = $1 AND ronda = $2`
	filas, err := tx.Query(ctx, votosDeLaRonda, sesionID, sesion.Ronda)
	if err != nil {
		return planningpoker.Votacion{}, fmt.Errorf("leer los votos de la ronda: %w", err)
	}
	votos := map[string]int{}
	var (
		integrante string
		v          int
	)
	if _, err := pgx.ForEachRow(filas, []any{&integrante, &v}, func() error {
		votos[integrante] = v
		return nil
	}); err != nil {
		return planningpoker.Votacion{}, fmt.Errorf("leer los votos de la ronda: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return planningpoker.Votacion{}, fmt.Errorf("confirmar la transaccion: %w", err)
	}
	return planningpoker.NuevaVotacion(sesion, votos), nil
}
