package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// RepositorioProyectos guarda los proyectos (US-001 y US-002).
type RepositorioProyectos struct {
	pool *pgxpool.Pool
}

// RepositorioProyectos cumple los puertos de los casos de uso de US-001 y US-002.
var (
	_ app.RepositorioProyectos         = (*RepositorioProyectos)(nil)
	_ app.RepositorioModificarProyecto = (*RepositorioProyectos)(nil)
	_ app.RepositorioListarProyectos   = (*RepositorioProyectos)(nil)
)

// NuevoRepositorioProyectos arma el repositorio sobre el pool de conexiones.
func NuevoRepositorioProyectos(pool *pgxpool.Pool) *RepositorioProyectos {
	return &RepositorioProyectos{pool: pool}
}

// RegistrarProyecto guarda el proyecto y lo devuelve con el ID que genera la
// base (DEFAULT gen_random_uuid() de la migracion 00002, RN-001-6). Las marcas
// creado_en y actualizado_en las completa la base con su DEFAULT.
func (r *RepositorioProyectos) RegistrarProyecto(ctx context.Context, p proyecto.Proyecto) (proyecto.Proyecto, error) {
	const insertar = `
INSERT INTO proyectos (nombre, descripcion, fecha_inicio, fecha_fin)
VALUES ($1, $2, $3, $4)
RETURNING id::text`
	var id string
	err := r.pool.QueryRow(ctx, insertar, p.Nombre, p.Descripcion, p.FechaInicio, p.FechaFin).Scan(&id)
	if err != nil {
		return proyecto.Proyecto{}, fmt.Errorf("insertar el proyecto: %w", err)
	}

	p.ID = id
	return p, nil
}

// BuscarProyecto devuelve el proyecto con ese ID (US-002). Compara el ID como
// texto: un ID vacio o mal formado no es un error de sintaxis de uuid sino un
// proyecto inexistente, igual que en el resto de los repositorios.
func (r *RepositorioProyectos) BuscarProyecto(ctx context.Context, proyectoID string) (proyecto.Proyecto, error) {
	const buscar = `
SELECT id::text, nombre, descripcion, fecha_inicio, fecha_fin
FROM proyectos
WHERE id::text = $1`
	var p proyecto.Proyecto
	err := r.pool.QueryRow(ctx, buscar, proyectoID).
		Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.FechaInicio, &p.FechaFin)
	if errors.Is(err, pgx.ErrNoRows) {
		return proyecto.Proyecto{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, proyectoID)
	}
	if err != nil {
		return proyecto.Proyecto{}, fmt.Errorf("buscar el proyecto: %w", err)
	}
	return p, nil
}

// ActualizarProyecto reemplaza el nombre, la descripcion y las fechas del
// proyecto con el ID de p (US-002). No toca el ID, creado_en ni los datos
// asociados (RN-002-4 y RN-002-7). actualizado_en lo mueve esta funcion porque
// la tabla no tiene trigger: su DEFAULT solo corre en el INSERT. Si ninguna fila
// coincide con el ID, devuelve app.ErrProyectoInexistente.
func (r *RepositorioProyectos) ActualizarProyecto(ctx context.Context, p proyecto.Proyecto) error {
	const actualizar = `
UPDATE proyectos
SET nombre = $2, descripcion = $3, fecha_inicio = $4, fecha_fin = $5, actualizado_en = now()
WHERE id::text = $1`
	etiqueta, err := r.pool.Exec(ctx, actualizar, p.ID, p.Nombre, p.Descripcion, p.FechaInicio, p.FechaFin)
	if err != nil {
		return fmt.Errorf("actualizar el proyecto: %w", err)
	}
	if etiqueta.RowsAffected() == 0 {
		return fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, p.ID)
	}
	return nil
}

// ListarProyectos devuelve todos los proyectos, sin orden: el orden lo pone el caso de
// uso (T-011), asi la memoria y Postgres se comportan igual.
func (r *RepositorioProyectos) ListarProyectos(ctx context.Context) ([]proyecto.Proyecto, error) {
	const listar = `SELECT id::text, nombre, descripcion, fecha_inicio, fecha_fin FROM proyectos`
	filas, err := r.pool.Query(ctx, listar)
	if err != nil {
		return nil, fmt.Errorf("listar los proyectos: %w", err)
	}
	defer filas.Close()

	proyectos := []proyecto.Proyecto{}
	for filas.Next() {
		var p proyecto.Proyecto
		if err := filas.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.FechaInicio, &p.FechaFin); err != nil {
			return nil, fmt.Errorf("leer un proyecto: %w", err)
		}
		proyectos = append(proyectos, p)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar los proyectos: %w", err)
	}
	return proyectos, nil
}
