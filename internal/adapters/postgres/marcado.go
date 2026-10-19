package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// RepositorioAsignaciones cumple tambien el puerto de los casos de uso de US-010:
// ObtenerItem es el de US-013 y ObtenerSprint, el de US-009.
var _ app.RepositorioMarcado = (*RepositorioAsignaciones)(nil)

// ItemEnSprint dice si el item tiene una fila de la relacion pendiente en ese sprint
// (RN-010-2). Las filas de otros sprints, o ya cerradas, no cuentan. Un ID mal formado
// no puede estar en ningun sprint: no es un error.
func (r *RepositorioAsignaciones) ItemEnSprint(ctx context.Context, itemID, sprintID string) (bool, error) {
	en, err := itemEnSprint(ctx, r.pool, itemID, sprintID)
	if esIDMalFormado(err) {
		return false, nil
	}
	return en, err
}

// itemEnSprint consulta la fila pendiente con el pool o dentro de una transaccion.
func itemEnSprint(ctx context.Context, q consultor, itemID, sprintID string) (bool, error) {
	var en bool
	const consulta = `
SELECT EXISTS (SELECT 1 FROM sprint_items WHERE item_id = $1 AND sprint_id = $2 AND resultado = 'pendiente')`
	if err := q.QueryRow(ctx, consulta, itemID, sprintID).Scan(&en); err != nil {
		return false, fmt.Errorf("consultar si el item %s esta en el sprint %s: %w", itemID, sprintID, err)
	}
	return en, nil
}

// GuardarCambioDeEstado marca o desmarca el item siguiendo los pasos de RN-010-9, en una
// transaccion READ COMMITTED y deteniendose en el primer error. Es de todo o nada.
func (r *RepositorioAsignaciones) GuardarCambioDeEstado(ctx context.Context, itemID, sprintID string, cambio app.CambioDeEstado) (backlog.ItemBacklog, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("abrir la transaccion: %w", err)
	}
	// Despues de un Commit exitoso, Rollback no hace nada.
	defer func() { _ = tx.Rollback(ctx) }()

	// Paso 1: se bloquea la fila del sprint y se lee su estado. FOR NO KEY UPDATE hace
	// que dos marcados del mismo sprint, o un marcado y un cierre (US-011), se esperen:
	// el segundo ve lo que dejo el primero (CL-010-9 y CL-010-10). Choca tambien con el
	// FOR SHARE de las asignaciones (US-009), y no bloquea a quien inserta una fila que
	// apunta al sprint. El sprint se bloquea antes que nada, como en RN-009-11.
	var (
		destino sprint.DatosDelSprint
		numero  int
		estado  string
	)
	const bloquearSprint = `SELECT proyecto_id::text, numero, estado FROM sprints WHERE id = $1 FOR NO KEY UPDATE`
	err = tx.QueryRow(ctx, bloquearSprint, sprintID).Scan(&destino.ProyectoID, &numero, &estado)
	if errors.Is(err, pgx.ErrNoRows) || esIDMalFormado(err) {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, sprintID)
	}
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("leer el sprint %s: %w", sprintID, err)
	}
	destino.Estado = sprint.Estado(estado)
	destino.Nombre = sprint.Sprint{Numero: numero}.Nombre()

	// Paso 2: se lee el item y si tiene una fila pendiente en el sprint, en sentencias
	// posteriores al candado: con READ COMMITTED, un marcado que espero ve lo que
	// guardo el otro.
	item, existe, err := leerItem(ctx, tx, itemID)
	if err != nil {
		return backlog.ItemBacklog{}, err
	}
	if !existe {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", app.ErrItemInexistente, itemID)
	}
	enElSprint, err := itemEnSprint(ctx, tx, itemID, sprintID)
	if err != nil {
		return backlog.ItemBacklog{}, err
	}

	// Paso 3: las mismas reglas que el caso de uso, con el sprint bloqueado.
	if err := sprint.VerificarMarcado(destino, enElSprint); err != nil {
		return backlog.ItemBacklog{}, err
	}

	// Paso 4: el cambio se aplica al item recien leido, no al que leyo el caso de uso,
	// y se guardan solo el estado y la fecha. Sin fecha se guarda como nulo.
	nuevo := cambio(item)
	var completadoEn *time.Time
	if !nuevo.CompletadoEn.IsZero() {
		completadoEn = &nuevo.CompletadoEn
	}
	const actualizar = `UPDATE items_backlog SET estado = $2, completado_en = $3 WHERE id = $1`
	if _, err := tx.Exec(ctx, actualizar, itemID, string(nuevo.Estado), completadoEn); err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("guardar el estado del item %s: %w", itemID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("confirmar la transaccion: %w", err)
	}
	return nuevo, nil
}
