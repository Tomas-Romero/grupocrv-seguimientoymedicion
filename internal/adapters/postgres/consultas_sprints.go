package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// RepositorioAsignaciones cumple tambien los puertos de lectura de las pantallas de sprints
// (T-015): ObtenerSprint es el de US-009 y ExisteProyecto, el del backlog que incluye.
var (
	_ app.RepositorioListarSprints   = (*RepositorioAsignaciones)(nil)
	_ app.RepositorioConsultarSprint = (*RepositorioAsignaciones)(nil)
)

// columnasDeItem son las columnas de items_backlog que arma filaDeItem, con el alias i.
const columnasDeItem = `i.id::text, i.proyecto_id::text, i.numero, i.titulo, i.descripcion, i.prioridad, i.estado, i.story_points, i.criterios, i.creado_en, i.completado_en`

// filaDeItem es lo que se lee de una fila de items_backlog antes de armar el item.
type filaDeItem struct {
	item         backlog.ItemBacklog
	prioridad    string
	estado       string
	storyPoints  *int32
	completadoEn *time.Time
}

// destinos son los destinos de Scan, en el orden de columnasDeItem.
func (f *filaDeItem) destinos() []any {
	return []any{&f.item.ID, &f.item.ProyectoID, &f.item.Numero, &f.item.Titulo, &f.item.Descripcion,
		&f.prioridad, &f.estado, &f.storyPoints, &f.item.Criterios, &f.item.CreadoEn, &f.completadoEn}
}

// armar convierte lo leido en el item.
func (f *filaDeItem) armar() (backlog.ItemBacklog, error) {
	item, err := completarItem(f.item, f.prioridad, f.estado, f.storyPoints, f.completadoEn)
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("el item de numero %d: %w", f.item.Numero, err)
	}
	return item, nil
}

// leerItems lee todas las filas de una consulta que devuelve columnasDeItem. Un ID mal formado
// es una lista vacia, no un error: no puede ser de ningun proyecto ni sprint.
func (r *RepositorioAsignaciones) leerItems(ctx context.Context, consulta, id, que string) ([]backlog.ItemBacklog, error) {
	filas, err := r.pool.Query(ctx, consulta, id)
	if esIDMalFormado(err) {
		return []backlog.ItemBacklog{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", que, id, err)
	}
	defer filas.Close()

	items := []backlog.ItemBacklog{}
	for filas.Next() {
		var f filaDeItem
		if err := filas.Scan(f.destinos()...); err != nil {
			return nil, fmt.Errorf("leer un item: %w", err)
		}
		item, err := f.armar()
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := filas.Err(); err != nil {
		if esIDMalFormado(err) {
			return []backlog.ItemBacklog{}, nil
		}
		return nil, fmt.Errorf("%s %s: %w", que, id, err)
	}
	return items, nil
}

// ListarSprintsConItems devuelve los sprints del proyecto con las historias que se les
// asignaron, sin orden: el orden lo pone el caso de uso (T-015). Incluye los sprints
// cerrados y las historias de cualquier resultado, porque la relacion es historica
// (RN-009-13). Un ID mal formado, o un proyecto sin sprints, devuelve una lista vacia.
func (r *RepositorioAsignaciones) ListarSprintsConItems(ctx context.Context, proyectoID string) ([]app.SprintConItems, error) {
	const listarSprints = `
SELECT id::text, proyecto_id::text, numero, sprint_goal, fecha_inicio, fecha_fin, estado
FROM sprints WHERE proyecto_id = $1`
	filas, err := r.pool.Query(ctx, listarSprints, proyectoID)
	if esIDMalFormado(err) {
		return []app.SprintConItems{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("listar los sprints del proyecto %s: %w", proyectoID, err)
	}

	sprints := []app.SprintConItems{}
	posicion := map[string]int{}
	for filas.Next() {
		var (
			s      sprint.Sprint
			estado string
		)
		if err := filas.Scan(&s.ID, &s.ProyectoID, &s.Numero, &s.SprintGoal, &s.FechaInicio, &s.FechaFin, &estado); err != nil {
			filas.Close()
			return nil, fmt.Errorf("leer un sprint: %w", err)
		}
		s.Estado = sprint.Estado(estado)
		posicion[s.ID] = len(sprints)
		sprints = append(sprints, app.SprintConItems{Sprint: s, Items: []backlog.ItemBacklog{}})
	}
	filas.Close()
	if err := filas.Err(); err != nil {
		if esIDMalFormado(err) {
			return []app.SprintConItems{}, nil
		}
		return nil, fmt.Errorf("listar los sprints del proyecto %s: %w", proyectoID, err)
	}
	if len(sprints) == 0 {
		return sprints, nil
	}

	const listarItems = `
SELECT si.sprint_id::text, ` + columnasDeItem + `
FROM sprint_items si
JOIN items_backlog i ON i.id = si.item_id
JOIN sprints s ON s.id = si.sprint_id
WHERE s.proyecto_id = $1`
	asignadas, err := r.pool.Query(ctx, listarItems, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar las historias de los sprints del proyecto %s: %w", proyectoID, err)
	}
	defer asignadas.Close()
	for asignadas.Next() {
		var (
			sprintID string
			f        filaDeItem
		)
		if err := asignadas.Scan(append([]any{&sprintID}, f.destinos()...)...); err != nil {
			return nil, fmt.Errorf("leer una historia de un sprint: %w", err)
		}
		item, err := f.armar()
		if err != nil {
			return nil, err
		}
		i := posicion[sprintID]
		sprints[i].Items = append(sprints[i].Items, item)
	}
	if err := asignadas.Err(); err != nil {
		return nil, fmt.Errorf("listar las historias de los sprints del proyecto %s: %w", proyectoID, err)
	}
	return sprints, nil
}

// ItemsDelSprint devuelve las historias que se asignaron al sprint, sin orden y sin importar
// el resultado de la fila. Un ID mal formado, o un sprint sin historias, devuelve una lista vacia.
func (r *RepositorioAsignaciones) ItemsDelSprint(ctx context.Context, sprintID string) ([]backlog.ItemBacklog, error) {
	const consulta = `
SELECT ` + columnasDeItem + `
FROM sprint_items si
JOIN items_backlog i ON i.id = si.item_id
WHERE si.sprint_id = $1`
	return r.leerItems(ctx, consulta, sprintID, "listar las historias del sprint")
}

// ListarItemsDisponibles devuelve las historias del proyecto que no estan completadas y no
// tienen una fila pendiente en ningun sprint, sin orden. Un ID mal formado devuelve una lista vacia.
func (r *RepositorioAsignaciones) ListarItemsDisponibles(ctx context.Context, proyectoID string) ([]backlog.ItemBacklog, error) {
	const consulta = `
SELECT ` + columnasDeItem + `
FROM items_backlog i
WHERE i.proyecto_id = $1
  AND i.estado <> 'completado'
  AND NOT EXISTS (SELECT 1 FROM sprint_items si WHERE si.item_id = i.id AND si.resultado = 'pendiente')`
	return r.leerItems(ctx, consulta, proyectoID, "listar las historias disponibles del proyecto")
}
