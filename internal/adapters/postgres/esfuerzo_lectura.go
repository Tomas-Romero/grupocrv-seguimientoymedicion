package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// RepositorioEsfuerzo tambien cumple el puerto del caso de uso de US-020.
var _ app.RepositorioConsultarEsfuerzo = (*RepositorioEsfuerzo)(nil)

// LeerEsfuerzo devuelve los registros, los titulos de las historias, los nombres de los
// integrantes y los sprints del proyecto. Las cuatro lecturas van en una transaccion de solo
// lectura con un mismo snapshot, asi todo es de la misma foto aunque alguien registre horas
// mientras tanto. Un ID mal formado, o un proyecto sin nada, devuelve una lectura vacia.
func (r *RepositorioEsfuerzo) LeerEsfuerzo(ctx context.Context, proyectoID string) (app.LecturaEsfuerzo, error) {
	lectura := app.LecturaEsfuerzo{Historias: map[string]string{}, Integrantes: map[string]string{}}
	if !formatoUUID.MatchString(proyectoID) {
		return lectura, nil
	}

	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return app.LecturaEsfuerzo{}, fmt.Errorf("abrir la lectura del esfuerzo: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const registros = `
SELECT id::text, proyecto_id::text, item_id::text, integrante_id::text, fecha, actividad, horas::float8
FROM registros_esfuerzo WHERE proyecto_id = $1`
	filas, err := tx.Query(ctx, registros, proyectoID)
	if err != nil {
		return app.LecturaEsfuerzo{}, fmt.Errorf("leer los registros de esfuerzo del proyecto %s: %w", proyectoID, err)
	}
	for filas.Next() {
		var rg esfuerzo.RegistroEsfuerzo
		if err := filas.Scan(&rg.ID, &rg.ProyectoID, &rg.ItemID, &rg.IntegranteID, &rg.Fecha, &rg.Actividad, &rg.Horas); err != nil {
			filas.Close()
			return app.LecturaEsfuerzo{}, fmt.Errorf("leer un registro de esfuerzo: %w", err)
		}
		lectura.Registros = append(lectura.Registros, rg)
	}
	filas.Close()
	if err := filas.Err(); err != nil {
		return app.LecturaEsfuerzo{}, fmt.Errorf("leer los registros de esfuerzo del proyecto %s: %w", proyectoID, err)
	}

	if lectura.Historias, err = leerNombres(ctx, tx, `SELECT id::text, titulo FROM items_backlog WHERE proyecto_id = $1`, proyectoID); err != nil {
		return app.LecturaEsfuerzo{}, fmt.Errorf("leer las historias del proyecto %s: %w", proyectoID, err)
	}
	if lectura.Integrantes, err = leerNombres(ctx, tx, `SELECT id::text, nombre FROM integrantes WHERE proyecto_id = $1`, proyectoID); err != nil {
		return app.LecturaEsfuerzo{}, fmt.Errorf("leer los integrantes del proyecto %s: %w", proyectoID, err)
	}

	const sprints = `
SELECT id::text, proyecto_id::text, numero, sprint_goal, fecha_inicio, fecha_fin, estado
FROM sprints WHERE proyecto_id = $1`
	filasSprint, err := tx.Query(ctx, sprints, proyectoID)
	if err != nil {
		return app.LecturaEsfuerzo{}, fmt.Errorf("leer los sprints del proyecto %s: %w", proyectoID, err)
	}
	for filasSprint.Next() {
		var (
			s      sprint.Sprint
			estado string
		)
		if err := filasSprint.Scan(&s.ID, &s.ProyectoID, &s.Numero, &s.SprintGoal, &s.FechaInicio, &s.FechaFin, &estado); err != nil {
			filasSprint.Close()
			return app.LecturaEsfuerzo{}, fmt.Errorf("leer un sprint: %w", err)
		}
		s.Estado = sprint.Estado(estado)
		lectura.Sprints = append(lectura.Sprints, s)
	}
	filasSprint.Close()
	if err := filasSprint.Err(); err != nil {
		return app.LecturaEsfuerzo{}, fmt.Errorf("leer los sprints del proyecto %s: %w", proyectoID, err)
	}

	return lectura, nil
}

// leerNombres ejecuta una consulta de dos columnas (ID y nombre) y las devuelve como mapa.
func leerNombres(ctx context.Context, tx pgx.Tx, consulta, proyectoID string) (map[string]string, error) {
	filas, err := tx.Query(ctx, consulta, proyectoID)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	nombres := map[string]string{}
	for filas.Next() {
		var id, nombre string
		if err := filas.Scan(&id, &nombre); err != nil {
			return nil, err
		}
		nombres[id] = nombre
	}
	return nombres, filas.Err()
}
