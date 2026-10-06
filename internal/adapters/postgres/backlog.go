// Package postgres implementa los repositorios de internal/app sobre PostgreSQL.
// El SQL de la aplicacion vive aca y en ningun otro lado.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// codigoTextoInvalido es el SQLSTATE que devuelve Postgres cuando un texto no
// se puede convertir al tipo de la columna; con un UUID, es un ID mal formado.
const codigoTextoInvalido = "22P02"

// esIDMalFormado dice si err es el error de un ID que no es un UUID valido. Un
// ID asi no puede ser de ningun proyecto: se trata como proyecto inexistente.
func esIDMalFormado(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == codigoTextoInvalido
}

// RepositorioBacklog guarda los items del Product Backlog (US-005).
type RepositorioBacklog struct {
	pool *pgxpool.Pool
}

// RepositorioBacklog cumple el puerto del caso de uso de US-005.
var _ app.RepositorioBacklog = (*RepositorioBacklog)(nil)

// NuevoRepositorioBacklog arma el repositorio sobre el pool de conexiones.
func NuevoRepositorioBacklog(pool *pgxpool.Pool) *RepositorioBacklog {
	return &RepositorioBacklog{pool: pool}
}

// ExisteProyecto es la lectura simple con la que el caso de uso verifica el
// proyecto antes de validar los datos (RN-005-9). Un ID mal formado es un
// proyecto inexistente, no un error.
func (r *RepositorioBacklog) ExisteProyecto(ctx context.Context, proyectoID string) (bool, error) {
	var existe bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM proyectos WHERE id = $1)`, proyectoID).Scan(&existe)
	if esIDMalFormado(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	return existe, nil
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

	// RN-005-9: el proyecto se vuelve a verificar dentro de la transaccion, por
	// si desaparecio despues de la lectura del caso de uso.
	var existe int
	err = tx.QueryRow(ctx, `SELECT 1 FROM proyectos WHERE id = $1`, proyectoID).Scan(&existe)
	if errors.Is(err, pgx.ErrNoRows) || esIDMalFormado(err) {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, proyectoID)
	}
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("leer el proyecto %s: %w", proyectoID, err)
	}

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
