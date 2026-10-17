package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
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

// ExisteProyecto todavia no consulta la base (esqueleto del paso RED).
func (r *RepositorioIntegrantes) ExisteProyecto(_ context.Context, _ string) (bool, error) {
	return false, nil
}

// RegistrarIntegrante todavia no guarda nada (esqueleto del paso RED).
func (r *RepositorioIntegrantes) RegistrarIntegrante(_ context.Context, i integrante.Integrante) (integrante.Integrante, error) {
	return i, nil
}
