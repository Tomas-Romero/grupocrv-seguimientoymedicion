// Package postgres implementa los repositorios de internal/app sobre PostgreSQL.
// El SQL de la aplicacion vive aca y en ningun otro lado.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// RepositorioBacklog guarda los items del Product Backlog (US-005).
type RepositorioBacklog struct {
	pool *pgxpool.Pool
}

// NuevoRepositorioBacklog arma el repositorio sobre el pool de conexiones.
func NuevoRepositorioBacklog(pool *pgxpool.Pool) *RepositorioBacklog {
	return &RepositorioBacklog{pool: pool}
}

// RegistrarItem guarda el item en el proyecto con el numero siguiente y lo
// devuelve con el ID que genera la base (RN-005-15).
//
// El numero se calcula dentro de la transaccion, asi que una transaccion
// revertida no deja huecos (CL-005-16).
func (r *RepositorioBacklog) RegistrarItem(ctx context.Context, proyectoID string, item backlog.ItemBacklog) (backlog.ItemBacklog, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("abrir la transaccion: %w", err)
	}
	// Despues de un Commit exitoso, Rollback no hace nada.
	defer func() { _ = tx.Rollback(ctx) }()

	var numero int
	const siguiente = `SELECT COALESCE(MAX(numero), 0) + 1 FROM items_backlog WHERE proyecto_id = $1`
	if err := tx.QueryRow(ctx, siguiente, proyectoID).Scan(&numero); err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("calcular el numero del item: %w", err)
	}

	// story_points no se inserta: un item nuevo siempre esta sin estimar y la
	// columna queda nula (RN-005-7). Estimar es US-013.
	const insertar = `
INSERT INTO items_backlog (proyecto_id, numero, titulo, descripcion, prioridad, estado, criterios, creado_en)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id::text`
	var id string
	err = tx.QueryRow(ctx, insertar, proyectoID, numero, item.Titulo, item.Descripcion,
		string(item.Prioridad), string(item.Estado), item.Criterios, item.CreadoEn).Scan(&id)
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("insertar el item: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("confirmar la transaccion: %w", err)
	}

	item.ID = id
	item.Numero = numero
	item.ProyectoID = proyectoID
	return item, nil
}
