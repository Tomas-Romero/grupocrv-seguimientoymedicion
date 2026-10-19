package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// RepositorioAsignaciones guarda la relacion entre un sprint y sus items (US-009). Lee los
// items con el mismo codigo que el repositorio del backlog (ObtenerItem es el de US-013), asi
// que lo incluye en lugar de repetir la consulta.
type RepositorioAsignaciones struct {
	*RepositorioBacklog
	pool *pgxpool.Pool
}

// RepositorioAsignaciones cumple el puerto del caso de uso de US-009.
var _ app.RepositorioAsignaciones = (*RepositorioAsignaciones)(nil)

// NuevoRepositorioAsignaciones arma el repositorio sobre el pool de conexiones.
func NuevoRepositorioAsignaciones(pool *pgxpool.Pool) *RepositorioAsignaciones {
	return &RepositorioAsignaciones{RepositorioBacklog: NuevoRepositorioBacklog(pool), pool: pool}
}

// ObtenerSprint lee el sprint con el que el caso de uso arma los datos que valida el
// dominio (RN-009-2). Un ID que ningun sprint tiene, o que no es un UUID, es un sprint
// inexistente y no un error.
func (r *RepositorioAsignaciones) ObtenerSprint(ctx context.Context, sprintID string) (sprint.Sprint, bool, error) {
	const consulta = `
SELECT id::text, proyecto_id::text, numero, sprint_goal, fecha_inicio, fecha_fin, estado
FROM sprints WHERE id = $1`
	var (
		s      sprint.Sprint
		estado string
	)
	err := r.pool.QueryRow(ctx, consulta, sprintID).Scan(&s.ID, &s.ProyectoID, &s.Numero, &s.SprintGoal, &s.FechaInicio, &s.FechaFin, &estado)
	if errors.Is(err, pgx.ErrNoRows) || esIDMalFormado(err) {
		return sprint.Sprint{}, false, nil
	}
	if err != nil {
		return sprint.Sprint{}, false, fmt.Errorf("leer el sprint %s: %w", sprintID, err)
	}
	s.Estado = sprint.Estado(estado)
	return s, true, nil
}

// ItemEnSprintAbierto dice si el item tiene una fila de la relacion con resultado
// pendiente: su sprint actual (RN-009-6 y RN-009-13). Las filas de sprints cerrados no
// cuentan. Un ID mal formado no puede ser de ningun item: no esta en ningun sprint.
func (r *RepositorioAsignaciones) ItemEnSprintAbierto(ctx context.Context, itemID string) (bool, error) {
	var en bool
	const consulta = `SELECT EXISTS (SELECT 1 FROM sprint_items WHERE item_id = $1 AND resultado = 'pendiente')`
	err := r.pool.QueryRow(ctx, consulta, itemID).Scan(&en)
	if esIDMalFormado(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("consultar si el item %s esta en un sprint: %w", itemID, err)
	}
	return en, nil
}

// AsignarItemASprint registra la asignacion siguiendo los pasos de RN-009-11, en una
// transaccion y deteniendose en el primer error. Es de todo o nada.
func (r *RepositorioAsignaciones) AsignarItemASprint(ctx context.Context, itemID, sprintID string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("abrir la transaccion: %w", err)
	}
	// Despues de un Commit exitoso, Rollback no hace nada.
	defer func() { _ = tx.Rollback(ctx) }()

	// Paso 1: se bloquea la fila del sprint y se lee su proyecto y su estado. FOR SHARE
	// deja que varias asignaciones al mismo sprint avancen a la vez y hace esperar a un
	// cierre (US-011), que modifica esa fila: una asignacion que llega despues del cierre
	// ve el sprint cerrado (CL-009-14). Se bloquea el sprint antes que el item en todas
	// las operaciones, para no tener bloqueos cruzados.
	var (
		destino sprint.DatosDelSprint
		numero  int
		estado  string
	)
	const bloquearSprint = `SELECT proyecto_id::text, numero, estado FROM sprints WHERE id = $1 FOR SHARE`
	err = tx.QueryRow(ctx, bloquearSprint, sprintID).Scan(&destino.ProyectoID, &numero, &estado)
	if errors.Is(err, pgx.ErrNoRows) || esIDMalFormado(err) {
		return fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, sprintID)
	}
	if err != nil {
		return fmt.Errorf("leer el sprint %s: %w", sprintID, err)
	}
	destino.Estado = sprint.Estado(estado)
	destino.Nombre = sprint.Sprint{Numero: numero}.Nombre()

	// Paso 2: se bloquea la fila del item y se leen los datos que valida la Definition of
	// Ready. Es FOR NO KEY UPDATE y no FOR UPDATE: serializa las asignaciones del mismo
	// item (la segunda ve a la primera y se rechaza con ErrItemYaAsignado en vez de chocar
	// con la clave primaria) sin bloquear a quien inserta una fila que apunta al item, como
	// un registro de esfuerzo (US-019).
	var (
		item         sprint.DatosDelItem
		estadoItem   string
		estimado     bool
		cantCriterio int
	)
	const bloquearItem = `
SELECT proyecto_id::text, estado, story_points IS NOT NULL, COALESCE(cardinality(criterios), 0)
FROM items_backlog WHERE id = $1 FOR NO KEY UPDATE`
	err = tx.QueryRow(ctx, bloquearItem, itemID).Scan(&item.ProyectoID, &estadoItem, &estimado, &cantCriterio)
	if errors.Is(err, pgx.ErrNoRows) || esIDMalFormado(err) {
		return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, itemID)
	}
	if err != nil {
		return fmt.Errorf("leer el item %s: %w", itemID, err)
	}
	item.Completado = backlog.Estado(estadoItem) == backlog.EstadoCompletado
	item.Estimado = estimado
	item.Criterios = cantCriterio

	// Paso 3: si el item ya esta en el sprint abierto. Se lee despues del bloqueo del
	// item, asi una asignacion que espero ve la fila que registro la otra.
	const yaAsignado = `SELECT EXISTS (SELECT 1 FROM sprint_items WHERE item_id = $1 AND resultado = 'pendiente')`
	if err := tx.QueryRow(ctx, yaAsignado, itemID).Scan(&item.EnSprintAbierto); err != nil {
		return fmt.Errorf("consultar si el item %s ya esta en un sprint: %w", itemID, err)
	}

	// Paso 4: las mismas reglas que el caso de uso, con los datos bloqueados.
	if err := sprint.VerificarAsignacion(item, destino); err != nil {
		return err
	}

	// Paso 5: la fila nace con resultado pendiente (el DEFAULT de la tabla).
	if _, err := tx.Exec(ctx, `INSERT INTO sprint_items (item_id, sprint_id) VALUES ($1, $2)`, itemID, sprintID); err != nil {
		return fmt.Errorf("insertar la asignacion: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("confirmar la transaccion: %w", err)
	}
	return nil
}
