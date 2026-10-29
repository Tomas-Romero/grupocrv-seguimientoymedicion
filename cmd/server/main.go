// Command server levanta la aplicacion web de Software Metrics & Estimation.
//
// Este archivo es el unico lugar donde se arman las dependencias: lee la
// configuracion, construye los adaptadores y se los pasa a los casos de uso.
// No contiene reglas de negocio.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	httpadapter "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/http"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/config"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/migraciones"
)

func main() {
	if err := run(); err != nil {
		slog.Error("el servidor termino con error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Cargar(os.Getenv)
	if err != nil {
		return fmt.Errorf("cargar configuracion: %w", err)
	}

	// Apagado ordenado: al recibir Ctrl+C se dejan terminar las peticiones en curso.
	ctx, detener := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer detener()

	// Si la base no responde al arrancar, el servidor no arranca.
	conexion, cancelarConexion := context.WithTimeout(ctx, 10*time.Second)
	pool, err := db.Conectar(conexion, cfg.DatabaseURL)
	cancelarConexion()
	if err != nil {
		return fmt.Errorf("conectar a la base de datos: %w", err)
	}
	defer pool.Close()

	// DEF-002: la base queda migrada, y con los datos de ejemplo si se pidieron,
	// antes de servir. Si algo falla el servidor no arranca: es mejor un error
	// visible que servir sobre un esquema viejo o vacio.
	preparacion, cancelarPreparacion := context.WithTimeout(ctx, time.Minute)
	err = prepararBase(preparacion, pool, cfg.CargarDatosEjemplo)
	cancelarPreparacion()
	if err != nil {
		return err
	}

	// T-005: si un handler hace panic, la persona recibe el error interno uniforme y el
	// servidor sigue atendiendo. Los handlers responden sus errores con
	// manejadorDeErrores.ResponderError.
	manejadorDeErrores := httpadapter.NuevoManejadorDeErrores(httpadapter.CatalogoDeErrores(), slog.Default())

	// T-004: las rutas viven en NuevoRouter. Los archivos estaticos se sirven desde
	// web/static (en la imagen de Docker esta junto al binario). T-011: las pantallas de
	// proyectos, integrantes y backlog.
	pantallas := httpadapter.NuevasPantallas(casosDeUsoDePantallas(pool), manejadorDeErrores)
	router := httpadapter.NuevoRouter(pool, 2*time.Second, http.Dir("web/static"), httpadapter.ConPantallas(pantallas))

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           manejadorDeErrores.Recuperar(router),
		ReadHeaderTimeout: 10 * time.Second,
	}

	errc := make(chan error, 1)
	go func() {
		slog.Info("servidor escuchando", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- err
		}
	}()

	select {
	case err := <-errc:
		return fmt.Errorf("escuchar en %s: %w", cfg.HTTPAddr, err)
	case <-ctx.Done():
		slog.Info("apagando el servidor")
		cierre, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelar()
		if err := srv.Shutdown(cierre); err != nil {
			return fmt.Errorf("apagar el servidor: %w", err)
		}
		return nil
	}
}

// prepararBase aplica las migraciones pendientes y, si cargarDatos es verdadero,
// carga los datos de ejemplo. Los dos pasos son idempotentes: corren en cada
// arranque sin repetir nada.
func prepararBase(ctx context.Context, pool *pgxpool.Pool, cargarDatos bool) error {
	if err := db.Migrar(ctx, pool, migraciones.Archivos); err != nil {
		return fmt.Errorf("aplicar migraciones: %w", err)
	}
	slog.Info("migraciones al dia")

	if !cargarDatos {
		return nil
	}
	if err := db.CargarDatosEjemplo(ctx, pool, migraciones.DatosEjemplo); err != nil {
		return fmt.Errorf("cargar datos de ejemplo: %w", err)
	}
	// Mensaje neutro a proposito: con ON CONFLICT DO NOTHING puede no haber
	// insertado nada, y el Exec de varias sentencias solo informa las filas de la
	// ultima, asi que no hay un numero confiable para mostrar.
	slog.Info("datos de ejemplo aplicados")
	return nil
}

// lecturaDeProyecto junta las dos lecturas que necesita ConsultarProyecto: el proyecto y
// sus integrantes. Cada una es de un repositorio distinto.
type lecturaDeProyecto struct {
	*postgres.RepositorioProyectos
	*postgres.RepositorioIntegrantes
}

var _ app.RepositorioConsultarProyecto = lecturaDeProyecto{}

// lecturaDeEstado junta las tres lecturas de ConsultarEstadoDelProyecto (US-004): el proyecto, sus historias y sus sprints
// con historias. Cada una es de un repositorio distinto y ya existe: no hay consultas nuevas.
type lecturaDeEstado struct {
	*postgres.RepositorioProyectos
	*postgres.RepositorioBacklog
	*postgres.RepositorioAsignaciones
}

var _ app.RepositorioEstadoGeneral = lecturaDeEstado{}

// lecturaDeDashboard junta las cinco lecturas de ConsultarDashboard (US-030): las tres del estado general, los registros de
// esfuerzo y los defectos con su sprint. Cada una es de un repositorio distinto y ya existe: no hay consultas nuevas.
type lecturaDeDashboard struct {
	*postgres.RepositorioProyectos
	*postgres.RepositorioBacklog
	*postgres.RepositorioAsignaciones
	*postgres.RepositorioEsfuerzo
	*postgres.RepositorioDefectos
}

var _ app.RepositorioDashboard = lecturaDeDashboard{}

// casosDeUsoDePantallas arma los casos de uso de las pantallas sobre los repositorios de
// Postgres (T-011, T-015, T-019, T-020, T-021, US-004, US-030 y US-033). Es el unico lugar donde se eligen los adaptadores.
func casosDeUsoDePantallas(pool *pgxpool.Pool) httpadapter.CasosDeUso {
	proyectos := postgres.NuevoRepositorioProyectos(pool)
	integrantes := postgres.NuevoRepositorioIntegrantes(pool)
	items := postgres.NuevoRepositorioBacklog(pool)
	sprints := postgres.NuevoRepositorioSprints(pool)
	asignaciones := postgres.NuevoRepositorioAsignaciones(pool)
	esfuerzo := postgres.NuevoRepositorioEsfuerzo(pool)
	poker := postgres.NuevoRepositorioPlanningPoker(pool)
	defectos := postgres.NuevoRepositorioDefectos(pool)
	return httpadapter.CasosDeUso{
		ListarProyectos:     app.NuevoListarProyectos(proyectos),
		ConsultarProyecto:   app.NuevoConsultarProyecto(lecturaDeProyecto{proyectos, integrantes}),
		ConsultarEstado:     app.NuevoConsultarEstadoDelProyecto(lecturaDeEstado{proyectos, items, asignaciones}, time.Now),
		ConsultarDashboard:  app.NuevoConsultarDashboard(lecturaDeDashboard{proyectos, items, asignaciones, esfuerzo, defectos}, time.Now),
		CrearProyecto:       app.NuevoCrearProyecto(proyectos),
		ModificarProyecto:   app.NuevoModificarProyecto(proyectos),
		RegistrarIntegrante: app.NuevoRegistrarIntegrante(integrantes),
		ListarBacklog:       app.NuevoListarBacklog(items),
		CrearItem:           app.NuevoCrearItemBacklog(items, time.Now),
		ConsultarItem:       app.NuevoConsultarItem(items),
		EditarItem:          app.NuevoEditarItem(items),
		EstimarItem:         app.NuevoEstimarItem(items),
		ListarSprints:       app.NuevoListarSprints(asignaciones),
		ConsultarSprint:     app.NuevoConsultarSprint(asignaciones),
		CrearSprint:         app.NuevoCrearSprint(sprints),
		AsignarItem:         app.NuevoAsignarItemASprint(asignaciones),
		MarcarCompletado:    app.NuevoMarcarCompletado(asignaciones, time.Now),
		DesmarcarCompletado: app.NuevoDesmarcarCompletado(asignaciones),
		CerrarSprint:        app.NuevoCerrarSprint(asignaciones),

		ConsultarSprintsAnteriores: app.NuevoConsultarSprintsAnteriores(asignaciones),

		RegistrarEsfuerzo:          app.NuevoRegistrarEsfuerzo(esfuerzo, time.Now),
		ConsultarEsfuerzo:          app.NuevoConsultarEsfuerzo(esfuerzo),
		CompararEsfuerzo:           app.NuevoCompararEsfuerzo(esfuerzo),
		Ahora:                      time.Now,
		IniciarSesionPlanningPoker: app.NuevoIniciarSesionPlanningPoker(poker),
		ConsultarSesion:            app.NuevoConsultarSesionPlanningPoker(poker),
		VotarEnSesion:              app.NuevoVotarEnSesion(poker),
		RevelarVotos:               app.NuevoRevelarVotos(poker),
		AbrirNuevaRonda:            app.NuevoAbrirNuevaRonda(poker),
		RegistrarAcuerdo:           app.NuevoRegistrarEstimacionAcordada(poker),
		ListarDefectos:             app.NuevoListarDefectos(defectos),
		RegistrarDefecto:           app.NuevoRegistrarDefecto(defectos),
		CambiarEstadoDefecto:       app.NuevoCambiarEstadoDefecto(defectos),
	}.ConReportes()
}
