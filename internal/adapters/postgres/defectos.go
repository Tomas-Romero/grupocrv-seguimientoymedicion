package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// Nombres de las claves foraneas de defectos que se traducen a un error de inexistencia (los que
// Postgres genera por defecto). La de sprint_resolucion_id no se escribe en esta historia.
const (
	claveForaneaDefectoProyecto = "defectos_proyecto_id_fkey"
	claveForaneaDefectoItem     = "defectos_item_id_fkey"
	claveForaneaDefectoSprint   = "defectos_sprint_deteccion_id_fkey"
)

// RepositorioDefectos guarda los defectos (US-022).
type RepositorioDefectos struct {
	pool *pgxpool.Pool
}

// RepositorioDefectos cumple el puerto del caso de uso de US-022.
var _ app.RepositorioDefectos = (*RepositorioDefectos)(nil)

// NuevoRepositorioDefectos arma el repositorio sobre el pool de conexiones.
func NuevoRepositorioDefectos(pool *pgxpool.Pool) *RepositorioDefectos {
	return &RepositorioDefectos{pool: pool}
}

// ExisteProyecto es la lectura simple con la que el caso de uso verifica el proyecto antes de
// validar los datos (RN-022-5). Un ID mal formado es un proyecto inexistente, no un error.
func (r *RepositorioDefectos) ExisteProyecto(ctx context.Context, proyectoID string) (bool, error) {
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

// ExisteItemEnProyecto dice si la historia existe y es de ese proyecto (RN-022-5). Una historia de
// otro proyecto no cuenta, y un ID vacio o mal formado es inexistente.
func (r *RepositorioDefectos) ExisteItemEnProyecto(ctx context.Context, proyectoID, itemID string) (bool, error) {
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

// ExisteSprintEnProyecto dice si el sprint existe y es de ese proyecto (RN-022-5). Uno de otro
// proyecto no cuenta, y un ID vacio o mal formado es inexistente: un sprint de deteccion sin elegir
// se rechaza igual que uno que no esta (CL-022-10).
func (r *RepositorioDefectos) ExisteSprintEnProyecto(ctx context.Context, proyectoID, sprintID string) (bool, error) {
	if !formatoUUID.MatchString(proyectoID) || !formatoUUID.MatchString(sprintID) {
		return false, nil
	}
	var existe bool
	const consulta = `SELECT EXISTS (SELECT 1 FROM sprints WHERE id = $1 AND proyecto_id = $2)`
	if err := r.pool.QueryRow(ctx, consulta, sprintID, proyectoID).Scan(&existe); err != nil {
		return false, fmt.Errorf("consultar el sprint %s del proyecto %s: %w", sprintID, proyectoID, err)
	}
	return existe, nil
}

// RegistrarDefecto guarda el defecto con un unico INSERT y lo devuelve con el ID que genera la base
// (RN-022-10). El INSERT no escribe el estado ni el sprint de resolucion: quedan 'abierto' y NULL
// (RN-022-3). No bloquea filas: los defectos son independientes entre si (RN-022-8).
//
// Las claves foraneas son la defensa final contra un proyecto, una historia o un sprint que
// desaparecieron despues de que el caso de uso los verifico: cada una se traduce al error de
// inexistencia que corresponde. Cualquier otro error, como un CHECK, se devuelve envuelto y sin
// traducir.
func (r *RepositorioDefectos) RegistrarDefecto(ctx context.Context, d defecto.Defecto) (defecto.Defecto, error) {
	const insertar = `
INSERT INTO defectos (proyecto_id, item_id, sprint_deteccion_id, descripcion, severidad)
VALUES ($1, $2, $3, $4, $5)
RETURNING id::text`

	// Un ID que no es un UUID no puede ser de ninguna fila: Postgres lo rechazaria con un error de
	// sintaxis, asi que se traduce antes, en el mismo orden en que el caso de uso los verifica
	// (proyecto, historia, sprint).
	if err := defectoConIDMalFormado(d); err != nil {
		return defecto.Defecto{}, err
	}

	var id string
	err := r.pool.QueryRow(ctx, insertar, d.ProyectoID, d.ItemID, d.SprintDeteccionID, d.Descripcion, string(d.Severidad)).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codigoClaveForaneaViolada {
		switch pgErr.ConstraintName {
		case claveForaneaDefectoProyecto:
			return defecto.Defecto{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, d.ProyectoID)
		case claveForaneaDefectoItem:
			return defecto.Defecto{}, fmt.Errorf("%w (id %s)", app.ErrItemInexistente, d.ItemID)
		case claveForaneaDefectoSprint:
			return defecto.Defecto{}, fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, d.SprintDeteccionID)
		}
	}
	if err != nil {
		return defecto.Defecto{}, fmt.Errorf("insertar el defecto: %w", err)
	}

	d.ID = id
	return d, nil
}

// defectoConIDMalFormado devuelve el error de inexistencia del primer ID del defecto que no es un
// UUID, o nil si los tres lo son.
func defectoConIDMalFormado(d defecto.Defecto) error {
	switch {
	case !formatoUUID.MatchString(d.ProyectoID):
		return fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, d.ProyectoID)
	case !formatoUUID.MatchString(d.ItemID):
		return fmt.Errorf("%w (id %s)", app.ErrItemInexistente, d.ItemID)
	case !formatoUUID.MatchString(d.SprintDeteccionID):
		return fmt.Errorf("%w (id %s)", app.ErrSprintInexistente, d.SprintDeteccionID)
	}
	return nil
}
