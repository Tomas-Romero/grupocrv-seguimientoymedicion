package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// RepositorioProyectos guarda los proyectos (US-001).
type RepositorioProyectos struct {
	pool *pgxpool.Pool
}

// RepositorioProyectos cumple el puerto del caso de uso de US-001.
var _ app.RepositorioProyectos = (*RepositorioProyectos)(nil)

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