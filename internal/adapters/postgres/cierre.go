package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// RepositorioAsignaciones cumple tambien el puerto del caso de uso de US-011
// (ObtenerSprint es el de US-009).
var _ app.RepositorioCierre = (*RepositorioAsignaciones)(nil)

// HistoriasDelSprint devuelve las historias con una fila pendiente en el sprint, en
// orden de asignacion, y si cada una esta completada. Un ID mal formado no es de
// ningun sprint: no tiene historias.
func (r *RepositorioAsignaciones) HistoriasDelSprint(ctx context.Context, sprintID string) ([]sprint.HistoriaDelCierre, error) {
	historias, err := historiasDelSprint(ctx, r.pool, sprintID)
	if esIDMalFormado(err) {
		return nil, nil
	}
	return historias, err
}

// lectorDeFilas es lo que historiasDelSprint necesita: lo cumplen el pool y una
// transaccion.
type lectorDeFilas interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// historiasDelSprint lee las filas pendientes del sprint con el pool o dentro de una
// transaccion.
func historiasDelSprint(ctx context.Context, q lectorDeFilas, sprintID string) ([]sprint.HistoriaDelCierre, error) {
	const consulta = `
SELECT si.item_id::text, i.estado = 'completado'
FROM sprint_items si JOIN items_backlog i ON i.id = si.item_id
WHERE si.sprint_id = $1 AND si.resultado = 'pendiente'
ORDER BY si.asignado_en, si.item_id`
	filas, err := q.Query(ctx, consulta, sprintID)
	if err != nil {
		return nil, fmt.Errorf("leer las historias del sprint %s: %w", sprintID, err)
	}
	historias, err := pgx.CollectRows(filas, func(f pgx.CollectableRow) (sprint.HistoriaDelCierre, error) {
		var h sprint.HistoriaDelCierre
		err := f.Scan(&h.ItemID, &h.Completada)
		return h, err
	})
	if err != nil {
		return nil, fmt.Errorf("leer las historias del sprint %s: %w", sprintID, err)
	}
	return historias, nil
}

// GuardarCierre cierra el sprint siguiendo los pasos de RN-011-8, en una transaccion
// READ COMMITTED que se detiene en el primer error. Es de todo o nada: si algo falla,
// el sprint sigue abierto y las historias quedan como estaban.
func (r *RepositorioAsignaciones) GuardarCierre(ctx context.Context, sprintID string) ([]sprint.ResultadoDelCierre, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("abrir la transaccion: %w", err)
	}
	// Despues de un Commit exitoso, Rollback no hace nada.
	defer func() { _ = tx.Rollback(ctx) }()

	// Paso 1: se bloquea la fila del sprint y se lee su estado. FOR NO KEY UPDATE ordena
	// el cierre con las asignaciones (US-009, FOR SHARE), los marcados (US-010) y otro
	// cierre: lo que llega antes cuenta y lo que llega despues ve el sprint cerrado
	// (RN-011-9). No bloquea a quien inserta una fila que apunta al sprint.
	var (
		destino sprint.DatosDelSprint
		numero  int
		estado  string
	)
	const bloquearSprint = `SELECT proyecto_id::text, numero, estado FROM sprints WHERE id = $1 FOR NO KEY UPDATE`
	err = tx.QueryRow(ctx, bloquearSprint, sprintID).Scan(&destino.ProyectoID, &numero, &estado)
	if errors.Is(err, pgx.ErrNoRows) || esIDMalFormado(err) {
		return nil, fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, sprintID)
	}
	if err != nil {
		return nil, fmt.Errorf("leer el sprint %s: %w", sprintID, err)
	}
	destino.Estado = sprint.Estado(estado)
	destino.Nombre = sprint.Sprint{Numero: numero}.Nombre()

	// Paso 2: las historias, en una sentencia posterior al candado: con READ COMMITTED,
	// un cierre que espero ve lo que asignaron o marcaron antes.
	historias, err := historiasDelSprint(ctx, tx, sprintID)
	if err != nil {
		return nil, err
	}

	// Paso 3: la misma funcion del dominio que el caso de uso, con el sprint bloqueado.
	resultados, err := sprint.Cerrar(destino, historias)
	if err != nil {
		return nil, err
	}

	// Paso 4: el resultado de cada fila (ninguna se borra), las no completadas de vuelta
	// a pendiente (RN-011-6) y el sprint cerrado.
	completadas, noCompletadas := []string{}, []string{}
	for _, res := range resultados {
		if res.Resultado == sprint.ResultadoCompletado {
			completadas = append(completadas, res.ItemID)
		} else {
			noCompletadas = append(noCompletadas, res.ItemID)
		}
	}
	const guardarResultado = `
UPDATE sprint_items SET resultado = $3
WHERE sprint_id = $1 AND item_id = ANY($2::uuid[]) AND resultado = 'pendiente'`
	if _, err := tx.Exec(ctx, guardarResultado, sprintID, completadas, string(sprint.ResultadoCompletado)); err != nil {
		return nil, fmt.Errorf("guardar las historias completadas: %w", err)
	}
	if _, err := tx.Exec(ctx, guardarResultado, sprintID, noCompletadas, string(sprint.ResultadoNoCompletado)); err != nil {
		return nil, fmt.Errorf("guardar las historias no completadas: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE items_backlog SET estado = 'pendiente' WHERE id = ANY($1::uuid[])`, noCompletadas); err != nil {
		return nil, fmt.Errorf("devolver al backlog las historias no completadas: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE sprints SET estado = 'cerrado' WHERE id = $1`, sprintID); err != nil {
		return nil, fmt.Errorf("cerrar el sprint %s: %w", sprintID, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("confirmar la transaccion: %w", err)
	}
	return resultados, nil
}
