package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// RepositorioPlanningPoker guarda las sesiones de Planning Poker (US-014). Lee el
// proyecto y la historia con el mismo codigo que el repositorio del backlog
// (ExisteProyecto es el de US-005 y ObtenerItem, el de US-013), asi que lo incluye.
type RepositorioPlanningPoker struct {
	*RepositorioBacklog
	pool *pgxpool.Pool
}

// RepositorioPlanningPoker cumple el puerto del caso de uso de US-014.
var _ app.RepositorioPlanningPoker = (*RepositorioPlanningPoker)(nil)

// NuevoRepositorioPlanningPoker arma el repositorio sobre el pool de conexiones.
func NuevoRepositorioPlanningPoker(pool *pgxpool.Pool) *RepositorioPlanningPoker {
	return &RepositorioPlanningPoker{RepositorioBacklog: NuevoRepositorioBacklog(pool), pool: pool}
}

// TieneSesionAbierta dice si la historia tiene una sesion que no esta finalizada
// (RN-014-7). Un ID mal formado no es de ninguna historia: no tiene sesiones.
func (r *RepositorioPlanningPoker) TieneSesionAbierta(ctx context.Context, itemID string) (bool, error) {
	abierta, err := tieneSesionAbierta(ctx, r.pool, itemID)
	if esIDMalFormado(err) {
		return false, nil
	}
	return abierta, err
}

func tieneSesionAbierta(ctx context.Context, q consultor, itemID string) (bool, error) {
	var abierta bool
	const consulta = `SELECT EXISTS (SELECT 1 FROM sesiones_planning_poker WHERE item_id = $1 AND estado <> 'finalizada')`
	if err := q.QueryRow(ctx, consulta, itemID).Scan(&abierta); err != nil {
		return false, fmt.Errorf("consultar las sesiones del item %s: %w", itemID, err)
	}
	return abierta, nil
}

// IDsDeIntegrantes devuelve los IDs de los integrantes del proyecto, en orden de alta
// (RN-014-4). Un ID mal formado no es de ningun proyecto: no tiene integrantes.
func (r *RepositorioPlanningPoker) IDsDeIntegrantes(ctx context.Context, proyectoID string) ([]string, error) {
	ids, err := idsDeIntegrantes(ctx, r.pool, proyectoID)
	if esIDMalFormado(err) {
		return []string{}, nil
	}
	return ids, err
}

func idsDeIntegrantes(ctx context.Context, q lectorDeFilas, proyectoID string) ([]string, error) {
	const consulta = `SELECT id::text FROM integrantes WHERE proyecto_id = $1 ORDER BY creado_en, id`
	filas, err := q.Query(ctx, consulta, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("leer los integrantes del proyecto %s: %w", proyectoID, err)
	}
	ids, err := pgx.CollectRows(filas, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("leer los integrantes del proyecto %s: %w", proyectoID, err)
	}
	return ids, nil
}

// RegistrarSesion registra la sesion siguiendo los pasos de RN-014-11, en una
// transaccion READ COMMITTED que se detiene en el primer error. Los participantes son
// los integrantes que tiene el proyecto dentro de la transaccion, no los de la sesion
// recibida. Es de todo o nada.
func (r *RepositorioPlanningPoker) RegistrarSesion(ctx context.Context, proyectoID string, sesion planningpoker.SesionPlanningPoker) (planningpoker.SesionPlanningPoker, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("abrir la transaccion: %w", err)
	}
	// Despues de un Commit exitoso, Rollback no hace nada.
	defer func() { _ = tx.Rollback(ctx) }()

	// Paso 1: se bloquea la fila de la historia y se lee su proyecto y su estado. FOR NO
	// KEY UPDATE serializa dos inicios sobre la misma historia (el segundo ve la sesion del
	// primero, CL-014-8) y ordena el inicio con el marcado de US-010, que actualiza esa
	// fila. No bloquea a quien inserta filas que apuntan a la historia.
	var deProyecto, estado string
	const bloquear = `SELECT proyecto_id::text, estado FROM items_backlog WHERE id = $1 FOR NO KEY UPDATE`
	err = tx.QueryRow(ctx, bloquear, sesion.ItemID).Scan(&deProyecto, &estado)
	if errors.Is(err, pgx.ErrNoRows) || esIDMalFormado(err) || (err == nil && deProyecto != proyectoID) {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", app.ErrItemInexistente, sesion.ItemID)
	}
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("leer el item %s: %w", sesion.ItemID, err)
	}

	// Paso 2: en sentencias posteriores al candado, si tiene sesion abierta y los
	// integrantes del proyecto.
	abierta, err := tieneSesionAbierta(ctx, tx, sesion.ItemID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}
	participantes, err := idsDeIntegrantes(ctx, tx, proyectoID)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	// Paso 3: la misma funcion del dominio que el caso de uso, con esos datos.
	historia := planningpoker.DatosDeLaHistoria{Completada: backlog.Estado(estado) == backlog.EstadoCompletado, ConSesionAbierta: abierta}
	nueva, err := planningpoker.NuevaSesion(sesion.ItemID, historia, participantes)
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	// Paso 4: la sesion (el ID, el estado y la ronda los pone la base) y sus participantes.
	const insertar = `INSERT INTO sesiones_planning_poker (item_id) VALUES ($1) RETURNING id::text, estado, ronda`
	var estadoGuardado string
	if err := tx.QueryRow(ctx, insertar, nueva.ItemID).Scan(&nueva.ID, &estadoGuardado, &nueva.Ronda); err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("insertar la sesion: %w", err)
	}
	nueva.Estado = planningpoker.Estado(estadoGuardado)
	nueva.ProyectoID = proyectoID
	const insertarParticipantes = `
INSERT INTO participantes_planning_poker (sesion_id, integrante_id)
SELECT $1, unnest($2::uuid[])`
	if _, err := tx.Exec(ctx, insertarParticipantes, nueva.ID, nueva.Participantes); err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("insertar los participantes: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("confirmar la transaccion: %w", err)
	}
	return nueva, nil
}
