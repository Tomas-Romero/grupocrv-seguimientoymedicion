package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// RepositorioSprints guarda los sprints de un proyecto (US-008).
type RepositorioSprints struct {
	pool *pgxpool.Pool
}

// RepositorioSprints cumple el puerto del caso de uso de US-008.
var _ app.RepositorioSprints = (*RepositorioSprints)(nil)

// NuevoRepositorioSprints arma el repositorio sobre el pool de conexiones.
func NuevoRepositorioSprints(pool *pgxpool.Pool) *RepositorioSprints {
	return &RepositorioSprints{pool: pool}
}

// FechasDelProyecto es la lectura simple con la que el caso de uso verifica el
// proyecto y toma sus fechas antes de validar los datos (RN-008-9). Un ID mal
// formado es un proyecto inexistente, no un error.
func (r *RepositorioSprints) FechasDelProyecto(ctx context.Context, proyectoID string) (time.Time, time.Time, bool, error) {
	var inicio, fin time.Time
	err := r.pool.QueryRow(ctx, `SELECT fecha_inicio, fecha_fin FROM proyectos WHERE id = $1`, proyectoID).Scan(&inicio, &fin)
	if errors.Is(err, pgx.ErrNoRows) || esIDMalFormado(err) {
		return time.Time{}, time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	return inicio, fin, true, nil
}

// RegistrarSprint guarda el sprint con el numero siguiente del proyecto y lo
// devuelve con el ID que genera la base (RN-008-16).
//
// Sigue los pasos de RN-008-14, en una transaccion y deteniendose en el primer
// error. El numero se calcula dentro de la transaccion, asi que una transaccion
// revertida no deja huecos (CL-008-16).
func (r *RepositorioSprints) RegistrarSprint(ctx context.Context, s sprint.Sprint) (sprint.Sprint, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return sprint.Sprint{}, fmt.Errorf("abrir la transaccion: %w", err)
	}
	// Despues de un Commit exitoso, Rollback no hace nada.
	defer func() { _ = tx.Rollback(ctx) }()

	// Paso 1: el candado sobre la fila del proyecto serializa las altas de sprints
	// del mismo proyecto: una segunda alta espera aca y despues ve el sprint que
	// registro la primera. Es FOR NO KEY UPDATE y no FOR UPDATE para no bloquear las
	// claves foraneas que apuntan al proyecto (por ejemplo, un integrante nuevo).
	// La misma lectura vuelve a verificar que el proyecto exista (RN-008-9) y trae
	// sus fechas vigentes, por si una modificacion (US-002) las cambio despues de la
	// lectura del caso de uso.
	var inicioProyecto, finProyecto time.Time
	const bloquear = `SELECT fecha_inicio, fecha_fin FROM proyectos WHERE id = $1 FOR NO KEY UPDATE`
	err = tx.QueryRow(ctx, bloquear, s.ProyectoID).Scan(&inicioProyecto, &finProyecto)
	if errors.Is(err, pgx.ErrNoRows) || esIDMalFormado(err) {
		return sprint.Sprint{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, s.ProyectoID)
	}
	if err != nil {
		return sprint.Sprint{}, fmt.Errorf("leer el proyecto %s: %w", s.ProyectoID, err)
	}

	// Paso 2: RN-008-7 con la misma funcion del dominio y las fechas bloqueadas.
	if err := sprint.VerificarDentroDelProyecto(s.FechaInicio, s.FechaFin, inicioProyecto, finProyecto); err != nil {
		return sprint.Sprint{}, err
	}

	// Paso 3: el ultimo sprint se lee en una sentencia aparte, despues del candado:
	// con READ COMMITTED cada sentencia ve lo confirmado hasta ese momento, asi una
	// alta que espero ve el sprint de la otra (RN-008-17 y RN-008-13).
	ultimo, err := leerUltimoSprint(ctx, tx, s.ProyectoID)
	if err != nil {
		return sprint.Sprint{}, err
	}
	if err := sprint.VerificarUltimoSprint(s.FechaInicio, ultimo); err != nil {
		return sprint.Sprint{}, err
	}

	// Paso 4: el numero es el maximo actual mas 1, o 1 si todavia no hay sprints.
	var numero int
	const siguiente = `SELECT COALESCE(MAX(numero), 0) + 1 FROM sprints WHERE proyecto_id = $1`
	if err := tx.QueryRow(ctx, siguiente, s.ProyectoID).Scan(&numero); err != nil {
		return sprint.Sprint{}, fmt.Errorf("calcular el numero del sprint: %w", err)
	}

	const insertar = `
INSERT INTO sprints (proyecto_id, numero, sprint_goal, fecha_inicio, fecha_fin, estado)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id::text`
	var id string
	err = tx.QueryRow(ctx, insertar, s.ProyectoID, numero, s.SprintGoal, s.FechaInicio, s.FechaFin, string(s.Estado)).Scan(&id)
	if err != nil {
		return sprint.Sprint{}, fmt.Errorf("insertar el sprint: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return sprint.Sprint{}, fmt.Errorf("confirmar la transaccion: %w", err)
	}

	s.ID = id
	s.Numero = numero
	return s, nil
}

// leerUltimoSprint devuelve el sprint de numero mas alto del proyecto, o nil si el
// proyecto todavia no tiene sprints.
func leerUltimoSprint(ctx context.Context, tx pgx.Tx, proyectoID string) (*sprint.UltimoSprint, error) {
	var ultimo sprint.UltimoSprint
	var estado string
	const consulta = `SELECT numero, estado, fecha_fin FROM sprints WHERE proyecto_id = $1 ORDER BY numero DESC LIMIT 1`
	err := tx.QueryRow(ctx, consulta, proyectoID).Scan(&ultimo.Numero, &estado, &ultimo.FechaFin)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("leer el ultimo sprint del proyecto %s: %w", proyectoID, err)
	}
	ultimo.Estado = sprint.Estado(estado)
	return &ultimo, nil
}
