package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// RepositorioBacklog cumple tambien el puerto del caso de uso de US-013.
var _ app.RepositorioEstimacion = (*RepositorioBacklog)(nil)

// ObtenerItem lee el item completo. Un ID que ningun item tiene, o que no es un
// UUID, es un item inexistente y no un error.
func (r *RepositorioBacklog) ObtenerItem(ctx context.Context, itemID string) (backlog.ItemBacklog, bool, error) {
	return leerItem(ctx, r.pool, itemID)
}

// consultor es lo que leerItem necesita: lo cumplen el pool y una transaccion.
type consultor interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// leerItem lee el item completo con el pool o dentro de una transaccion (US-010 lo lee
// con el sprint bloqueado). Un ID que ningun item tiene, o que no es un UUID, es un
// item inexistente y no un error.
func leerItem(ctx context.Context, q consultor, itemID string) (backlog.ItemBacklog, bool, error) {
	const consulta = `
SELECT id::text, proyecto_id::text, numero, titulo, descripcion, prioridad, estado, story_points, criterios, creado_en, completado_en
FROM items_backlog WHERE id = $1`
	var (
		item         backlog.ItemBacklog
		prioridad    string
		estado       string
		storyPoints  *int32
		completadoEn *time.Time
	)
	err := q.QueryRow(ctx, consulta, itemID).Scan(&item.ID, &item.ProyectoID, &item.Numero, &item.Titulo,
		&item.Descripcion, &prioridad, &estado, &storyPoints, &item.Criterios, &item.CreadoEn, &completadoEn)
	if errors.Is(err, pgx.ErrNoRows) || esIDMalFormado(err) {
		return backlog.ItemBacklog{}, false, nil
	}
	if err != nil {
		return backlog.ItemBacklog{}, false, fmt.Errorf("leer el item %s: %w", itemID, err)
	}
	item, err = completarItem(item, prioridad, estado, storyPoints, completadoEn)
	if err != nil {
		return backlog.ItemBacklog{}, false, fmt.Errorf("el item %s: %w", itemID, err)
	}
	return item, true, nil
}

// completarItem arma el item con lo que se leyo de la fila: convierte la prioridad, el
// estado, la estimacion y la fecha de completado. Lo usan la lectura de un item y el
// listado del backlog, para que los dos lo armen igual.
func completarItem(item backlog.ItemBacklog, prioridad, estado string, storyPoints *int32, completadoEn *time.Time) (backlog.ItemBacklog, error) {
	item.Prioridad = backlog.Prioridad(prioridad)
	item.Estado = backlog.Estado(estado)
	// El nulo es "sin fecha": solo un item completado la tiene (US-010).
	if completadoEn != nil {
		item.CompletadoEn = *completadoEn
	}

	// El nulo de la columna es "sin estimar" (RN-005-7). Un valor fuera de la
	// escala es un dato corrupto: se informa, no se oculta.
	item.StoryPoints = backlog.SinEstimar()
	if storyPoints != nil {
		puntos, err := backlog.NuevosStoryPoints(int(*storyPoints))
		if err != nil {
			return backlog.ItemBacklog{}, fmt.Errorf("tiene una estimacion guardada invalida: %w", err)
		}
		item.StoryPoints = puntos
	}
	return item, nil
}

// GuardarEstimacion escribe solo la columna story_points con un unico UPDATE
// (RN-013-7). Sin estimar se guarda como nulo, no como 0.
//
// El UPDATE esta condicionado a que el item no este completado (RN-013-4): si
// se completo entre la lectura del caso de uso y este guardado, no actualiza
// ninguna fila. Entonces se relee el item para distinguir los dos motivos: si no
// existe, o el ID no es un UUID, devuelve app.ErrItemInexistente; si esta
// completado, backlog.ErrItemCompletado.
func (r *RepositorioBacklog) GuardarEstimacion(ctx context.Context, itemID string, puntos backlog.StoryPoints) error {
	var valor any // nil se guarda como NULL
	if p, estimado := puntos.Valor(); estimado {
		valor = p
	}

	const actualizar = `UPDATE items_backlog SET story_points = $2 WHERE id = $1 AND estado <> 'completado'`
	resultado, err := r.pool.Exec(ctx, actualizar, itemID, valor)
	if esIDMalFormado(err) {
		return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, itemID)
	}
	if err != nil {
		return fmt.Errorf("guardar la estimacion del item %s: %w", itemID, err)
	}
	if resultado.RowsAffected() > 0 {
		return nil
	}

	// Ninguna fila actualizada: o el item no existe o esta completado.
	var estado string
	err = r.pool.QueryRow(ctx, `SELECT estado FROM items_backlog WHERE id = $1`, itemID).Scan(&estado)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, itemID)
	}
	if err != nil {
		return fmt.Errorf("releer el item %s: %w", itemID, err)
	}
	if backlog.Estado(estado) == backlog.EstadoCompletado {
		return fmt.Errorf("%w (id %s)", backlog.ErrItemCompletado, itemID)
	}
	// El item existe y no esta completado: no hay motivo para que el UPDATE no
	// haya tocado la fila. Se informa en lugar de devolver un exito falso.
	return fmt.Errorf("no se pudo guardar la estimacion del item %s", itemID)
}
