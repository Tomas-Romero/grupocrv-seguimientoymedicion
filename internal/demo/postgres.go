package demo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
)

// CasosPostgres arma los casos de uso de la carga sobre una base real, con los mismos repositorios que cmd/server.
func CasosPostgres(pool *pgxpool.Pool, reloj *Reloj) (Casos, Lecturas) {
	proyectos := postgres.NuevoRepositorioProyectos(pool)
	integrantes := postgres.NuevoRepositorioIntegrantes(pool)
	items := postgres.NuevoRepositorioBacklog(pool)
	sprints := postgres.NuevoRepositorioSprints(pool)
	asignaciones := postgres.NuevoRepositorioAsignaciones(pool)
	esfuerzo := postgres.NuevoRepositorioEsfuerzo(pool)
	poker := postgres.NuevoRepositorioPlanningPoker(pool)
	defectos := postgres.NuevoRepositorioDefectos(pool)

	casos := Casos{
		ListarProyectos: app.NuevoListarProyectos(proyectos),
		CrearProyecto:   app.NuevoCrearProyecto(proyectos),
		RegistrarInteg:  app.NuevoRegistrarIntegrante(integrantes),
		CrearItem:       app.NuevoCrearItemBacklog(items, reloj.Ahora),
		EstimarItem:     app.NuevoEstimarItem(items),
		CrearSprint:     app.NuevoCrearSprint(sprints),
		AsignarItem:     app.NuevoAsignarItemASprint(asignaciones),
		MarcarCompleta:  app.NuevoMarcarCompletado(asignaciones, reloj.Ahora),
		CerrarSprint:    app.NuevoCerrarSprint(asignaciones),
		IniciarSesion:   app.NuevoIniciarSesionPlanningPoker(poker),
		Votar:           app.NuevoVotarEnSesion(poker),
		Revelar:         app.NuevoRevelarVotos(poker),
		NuevaRonda:      app.NuevoAbrirNuevaRonda(poker),
		Acordar:         app.NuevoRegistrarEstimacionAcordada(poker),
		RegistrarHoras:  app.NuevoRegistrarEsfuerzo(esfuerzo, reloj.Ahora),
		RegistrarDef:    app.NuevoRegistrarDefecto(defectos),
		CambiarEstado:   app.NuevoCambiarEstadoDefecto(defectos),
	}
	lecturas := Lecturas{
		ListarBacklog:     app.NuevoListarBacklog(items),
		ListarSprints:     app.NuevoListarSprints(asignaciones),
		ConsultarEsfuerzo: app.NuevoConsultarEsfuerzo(esfuerzo),
		ListarDefectos:    app.NuevoListarDefectos(defectos),
		ConsultarSesion:   app.NuevoConsultarSesionPlanningPoker(poker),
	}
	return casos, lecturas
}

// BorrarProyecto borra el proyecto de demostracion y todo lo suyo, y dice cuantos proyectos borro. Busca por el nombre exacto y no toca
// ningun otro proyecto (RN-T023-5). Las claves foraneas hacia el proyecto borran en cascada: integrantes, historias, sprints, horas,
// sesiones de Planning Poker y defectos.
func BorrarProyecto(ctx context.Context, pool *pgxpool.Pool) (int64, error) {
	etiqueta, err := pool.Exec(ctx, `DELETE FROM proyectos WHERE nombre = $1`, NombreProyecto)
	if err != nil {
		return 0, fmt.Errorf("borrar el proyecto %q: %w", NombreProyecto, err)
	}
	return etiqueta.RowsAffected(), nil
}
