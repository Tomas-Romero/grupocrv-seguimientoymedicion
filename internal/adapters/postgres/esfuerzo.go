package postgres

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
)

// sqlstateClaveForaneaEsfuerzo es el SQLSTATE de una clave foranea violada.
const sqlstateClaveForaneaEsfuerzo = "23503"

// Nombres de las claves foraneas de registros_esfuerzo (los que Postgres genera
// por defecto): sirven para saber cual fallo cuando un INSERT choca con una.
const (
	claveForaneaProyecto   = "registros_esfuerzo_proyecto_id_fkey"
	claveForaneaItem       = "registros_esfuerzo_item_id_fkey"
	claveForaneaIntegrante = "registros_esfuerzo_integrante_id_fkey"
)

// formatoUUID reconoce un UUID bien formado (8-4-4-4-12 digitos hexadecimales).
var formatoUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// RepositorioEsfuerzo guarda los registros de esfuerzo (US-019).
type RepositorioEsfuerzo struct {
	pool *pgxpool.Pool
}

// RepositorioEsfuerzo cumple el puerto del caso de uso de US-019.
var _ app.RepositorioEsfuerzo = (*RepositorioEsfuerzo)(nil)

// NuevoRepositorioEsfuerzo arma el repositorio sobre el pool de conexiones.
func NuevoRepositorioEsfuerzo(pool *pgxpool.Pool) *RepositorioEsfuerzo {
	return &RepositorioEsfuerzo{pool: pool}
}

// ExisteProyecto es la lectura simple con la que el caso de uso verifica el
// proyecto antes de validar los datos (RN-019-6). Un ID mal formado es un
// proyecto inexistente, no un error.
func (r *RepositorioEsfuerzo) ExisteProyecto(ctx context.Context, proyectoID string) (bool, error) {
	if !formatoUUID.MatchString(proyectoID) {
		return false, nil
	}
	var existe bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM proyectos WHERE id = $1)`, proyectoID).Scan(&existe)
	if err != nil {
		return false, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	return existe, nil
}

// ExisteItemEnProyecto dice si la historia existe y es de ese proyecto (RN-019-6).
// Una historia de otro proyecto no cuenta, y un ID mal formado es inexistente.
func (r *RepositorioEsfuerzo) ExisteItemEnProyecto(ctx context.Context, proyectoID, itemID string) (bool, error) {
	if !formatoUUID.MatchString(proyectoID) || !formatoUUID.MatchString(itemID) {
		return false, nil
	}
	var existe bool
	const consulta = `SELECT EXISTS (SELECT 1 FROM items_backlog WHERE id = $1 AND proyecto_id = $2)`
	if err := r.pool.QueryRow(ctx, consulta, itemID, proyectoID).Scan(&existe); err != nil {
		return false, fmt.Errorf("consultar la historia %s del proyecto %s: %w", itemID, proyectoID, err)
	}
	return existe, nil
}

// ExisteIntegranteEnProyecto dice si el integrante existe y es de ese proyecto
// (RN-019-6). Uno de otro proyecto no cuenta, y un ID mal formado es inexistente.
func (r *RepositorioEsfuerzo) ExisteIntegranteEnProyecto(ctx context.Context, proyectoID, integranteID string) (bool, error) {
	if !formatoUUID.MatchString(proyectoID) || !formatoUUID.MatchString(integranteID) {
		return false, nil
	}
	var existe bool
	const consulta = `SELECT EXISTS (SELECT 1 FROM integrantes WHERE id = $1 AND proyecto_id = $2)`
	if err := r.pool.QueryRow(ctx, consulta, integranteID, proyectoID).Scan(&existe); err != nil {
		return false, fmt.Errorf("consultar el integrante %s del proyecto %s: %w", integranteID, proyectoID, err)
	}
	return existe, nil
}

// RegistrarEsfuerzo guarda el registro con un unico INSERT y lo devuelve con el ID
// que genera la base (RN-019-9 y RN-019-10). No bloquea filas: los registros son
// independientes entre si.
//
// Las claves foraneas son la defensa final contra un proyecto, una historia o un
// integrante que desaparecieron despues de que el caso de uso los verifico: cada
// una se traduce al error de inexistencia que corresponde. Cualquier otro error,
// como el CHECK de las horas, se devuelve envuelto y sin traducir.
func (r *RepositorioEsfuerzo) RegistrarEsfuerzo(ctx context.Context, registro esfuerzo.RegistroEsfuerzo) (esfuerzo.RegistroEsfuerzo, error) {
	const insertar = `
INSERT INTO registros_esfuerzo (proyecto_id, item_id, integrante_id, fecha, actividad, horas)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id::text`

	// Un ID que no es un UUID no puede ser de ninguna fila: Postgres lo rechazaria
	// con un error de sintaxis, asi que se traduce antes, en el mismo orden en que
	// el caso de uso los verifica (proyecto, historia, integrante).
	if err := idMalFormado(registro); err != nil {
		return esfuerzo.RegistroEsfuerzo{}, err
	}

	var id string
	err := r.pool.QueryRow(ctx, insertar, registro.ProyectoID, registro.ItemID, registro.IntegranteID,
		registro.Fecha, registro.Actividad, registro.Horas).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == sqlstateClaveForaneaEsfuerzo {
		switch pgErr.ConstraintName {
		case claveForaneaProyecto:
			return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, registro.ProyectoID)
		case claveForaneaItem:
			return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("%w (id %s)", app.ErrItemInexistente, registro.ItemID)
		case claveForaneaIntegrante:
			return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("%w (id %s)", app.ErrIntegranteInexistente, registro.IntegranteID)
		}
	}
	if err != nil {
		return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("insertar el registro de esfuerzo: %w", err)
	}

	registro.ID = id
	return registro, nil
}

// idMalFormado devuelve el error de inexistencia del primer ID del registro que no
// es un UUID, o nil si los tres lo son.
func idMalFormado(r esfuerzo.RegistroEsfuerzo) error {
	switch {
	case !formatoUUID.MatchString(r.ProyectoID):
		return fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, r.ProyectoID)
	case !formatoUUID.MatchString(r.ItemID):
		return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, r.ItemID)
	case !formatoUUID.MatchString(r.IntegranteID):
		return fmt.Errorf("%w (id %s)", app.ErrIntegranteInexistente, r.IntegranteID)
	}
	return nil
}
