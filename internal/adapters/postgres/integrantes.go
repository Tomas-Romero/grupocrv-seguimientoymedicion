package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
)

// SQLSTATE que devuelve Postgres cuando el INSERT de un integrante choca con una
// restriccion de la tabla.
const (
	// codigoUnicaViolada: UNIQUE (proyecto_id, email), el email ya esta en el proyecto.
	codigoUnicaViolada = "23505"
	// codigoClaveForaneaViolada: integrantes.proyecto_id apunta a un proyecto que
	// no existe.
	codigoClaveForaneaViolada = "23503"
)

// RepositorioIntegrantes guarda los integrantes de un proyecto (US-003).
type RepositorioIntegrantes struct {
	pool *pgxpool.Pool
}

// RepositorioIntegrantes cumple el puerto del caso de uso de US-003.
var _ app.RepositorioRegistrarIntegrante = (*RepositorioIntegrantes)(nil)

// NuevoRepositorioIntegrantes arma el repositorio sobre el pool de conexiones.
func NuevoRepositorioIntegrantes(pool *pgxpool.Pool) *RepositorioIntegrantes {
	return &RepositorioIntegrantes{pool: pool}
}

// ExisteProyecto es la lectura simple con la que el caso de uso verifica el
// proyecto antes de validar los datos (RN-003-6). Un ID mal formado no es un
// error: no puede ser de ningun proyecto, asi que no existe.
func (r *RepositorioIntegrantes) ExisteProyecto(ctx context.Context, proyectoID string) (bool, error) {
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

// RegistrarIntegrante guarda al integrante y lo devuelve con el ID que genera la
// base (DEFAULT gen_random_uuid() de la migracion 00005, RN-003-10).
//
// Es un unico INSERT, asi que es de todo o nada (RN-003-9). Las restricciones de
// la tabla son la defensa final y se traducen asi: UNIQUE (proyecto_id, email)
// a ErrEmailDuplicado, y una clave foranea rota o un ID mal formado a
// ErrProyectoInexistente, por si el proyecto desaparece despues de que el caso de
// uso lo verifico. Cualquier otro error, como el CHECK del rol, se devuelve
// envuelto y sin traducir.
func (r *RepositorioIntegrantes) RegistrarIntegrante(ctx context.Context, i integrante.Integrante) (integrante.Integrante, error) {
	const insertar = `
INSERT INTO integrantes (proyecto_id, nombre, email, rol)
VALUES ($1, $2, $3, $4)
RETURNING id::text`
	var id string
	err := r.pool.QueryRow(ctx, insertar, i.ProyectoID, i.Nombre, i.Email, string(i.Rol)).Scan(&id)
	if esIDMalFormado(err) {
		return integrante.Integrante{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, i.ProyectoID)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case codigoUnicaViolada:
			return integrante.Integrante{}, fmt.Errorf("%w (email %s)", app.ErrEmailDuplicado, i.Email)
		case codigoClaveForaneaViolada:
			return integrante.Integrante{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, i.ProyectoID)
		}
	}
	if err != nil {
		return integrante.Integrante{}, fmt.Errorf("insertar el integrante: %w", err)
	}

	i.ID = id
	return i, nil
}
