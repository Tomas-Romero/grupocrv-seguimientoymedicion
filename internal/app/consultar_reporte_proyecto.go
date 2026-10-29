package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Reporte de un proyecto (spec US-033). No accede a repositorios ni recalcula nada (RN-033-1): compone las lecturas que ya existen,
// que le llegan por el constructor como funciones (cada caso de uso se pasa con su metodo Ejecutar).

// FuentesDelReporteDelProyecto son las lecturas que el reporte del proyecto compone. Cada una se llama una sola vez.
type FuentesDelReporteDelProyecto struct {
	Proyecto    func(ctx context.Context, proyectoID string) (DetalleDeProyecto, error)
	Backlog     func(ctx context.Context, proyectoID string) ([]backlog.ItemBacklog, error)
	Sprints     func(ctx context.Context, proyectoID string) ([]ResumenDeSprint, error)
	Anteriores  func(ctx context.Context, proyectoID string) (SprintsAnteriores, error)
	Esfuerzo    func(ctx context.Context, proyectoID string, filtro FiltroEsfuerzo) (ConsultaEsfuerzo, error)
	Comparacion func(ctx context.Context, proyectoID string) (metricas.ComparacionProyecto, error)
	Defectos    func(ctx context.Context, proyectoID string, filtro FiltroDefectos) ([]DefectoListado, error)
	Dashboard   func(ctx context.Context, proyectoID string) (Dashboard, error)
}

// ReporteDelProyecto tiene las siete secciones del reporte, siempre las mismas y en este orden (RN-033-2): el proyecto, los
// integrantes, el backlog, los sprints, el esfuerzo, los defectos y las metricas. Cada seccion lleva su propio error (RN-033-3).
type ReporteDelProyecto struct {
	Proyecto    SeccionDelProyecto
	Integrantes []integrante.Integrante
	Backlog     SeccionDelBacklog
	Sprints     SeccionDeSprints
	Esfuerzo    SeccionDeEsfuerzo
	Defectos    SeccionDeDefectos
	Metricas    SeccionDeMetricas
}

// IntegrantesVacios dice si el proyecto no tiene integrantes. Los trae la misma lectura que el proyecto, asi que no tienen error propio.
func (r ReporteDelProyecto) IntegrantesVacios() bool { return len(r.Integrantes) == 0 }

// SeccionDelProyecto son los datos del proyecto, su estado y el avance del backlog (US-004).
type SeccionDelProyecto struct {
	Proyecto proyecto.Proyecto
	Estado   EstadoGeneral
	Err      error
}

// SeccionDelBacklog son las historias en el orden de la pantalla del backlog (RN-T011-4).
type SeccionDelBacklog struct {
	Items []backlog.ItemBacklog
	Err   error
}

// Vacia dice si el proyecto no tiene historias.
func (s SeccionDelBacklog) Vacia() bool { return len(s.Items) == 0 }

// SprintDelReporte es un sprint con sus totales y, si esta cerrado, cuantas historias volvieron al backlog (RN-009-13).
type SprintDelReporte struct {
	ResumenDeSprint
	VolvieronAlBacklog int
}

// SeccionDeSprints son los sprints por numero descendente (RN-T015-3).
type SeccionDeSprints struct {
	Sprints []SprintDelReporte
	Err     error
}

// Vacia dice si el proyecto no tiene sprints.
func (s SeccionDeSprints) Vacia() bool { return len(s.Sprints) == 0 }

// SeccionDeEsfuerzo es el total de horas registradas del proyecto y la comparacion de lo estimado contra lo trabajado (T-019).
type SeccionDeEsfuerzo struct {
	TotalHoras  float64
	Comparacion metricas.ComparacionProyecto
	Err         error
}

// Vacia dice si no hay horas registradas ni historias que comparar.
func (s SeccionDeEsfuerzo) Vacia() bool {
	return s.TotalHoras == 0 && len(s.Comparacion.Historias) == 0
}

// SeccionDeDefectos es la lista de defectos del proyecto (del mas reciente al mas antiguo) y sus totales por severidad y por estado.
type SeccionDeDefectos struct {
	Lista   []DefectoListado
	Totales TarjetaDeDefectos
	Err     error
}

// Vacia dice si el proyecto no tiene defectos.
func (s SeccionDeDefectos) Vacia() bool { return len(s.Lista) == 0 }

// SeccionDeMetricas son la velocidad, el sprint de referencia y los sprints cerrados con sus defectos (US-030).
type SeccionDeMetricas struct {
	Velocidad  TarjetaDeVelocidad
	Referencia *TarjetaDeSprint
	Err        error
}

// Vacia dice si no hay sprint de referencia ni sprints cerrados.
func (s SeccionDeMetricas) Vacia() bool { return s.Referencia == nil && len(s.Velocidad.Cerrados) == 0 }

// ConsultarReporteDelProyecto es el caso de uso de lectura del reporte del proyecto (US-033).
type ConsultarReporteDelProyecto struct {
	fuentes FuentesDelReporteDelProyecto
}

// NuevoConsultarReporteDelProyecto arma el caso de uso con las lecturas que compone.
func NuevoConsultarReporteDelProyecto(fuentes FuentesDelReporteDelProyecto) *ConsultarReporteDelProyecto {
	return &ConsultarReporteDelProyecto{fuentes: fuentes}
}

// Ejecutar lee el proyecto primero: si no existe es ErrProyectoInexistente y no se lee nada mas (RN-033-6), y si falla la lectura no hay con
// que armar la pagina. Despues arma cada seccion por separado (RN-033-3): una que falla informa su error y las demas se muestran.
func (c *ConsultarReporteDelProyecto) Ejecutar(ctx context.Context, proyectoID string) (ReporteDelProyecto, error) {
	detalle, err := c.fuentes.Proyecto(ctx, proyectoID)
	if err != nil {
		return ReporteDelProyecto{}, fmt.Errorf("armar el reporte del proyecto %s: %w", proyectoID, err)
	}
	r := ReporteDelProyecto{Integrantes: detalle.Integrantes}
	r.Proyecto.Proyecto = detalle.Proyecto

	// El dashboard aporta el estado, los totales de defectos y las metricas: si falla, las tres secciones lo informan.
	dashboard, errDashboard := c.fuentes.Dashboard(ctx, proyectoID)
	if errDashboard != nil {
		errDashboard = fmt.Errorf("leer el dashboard del proyecto %s: %w", proyectoID, errDashboard)
		r.Proyecto.Err, r.Metricas.Err = errDashboard, errDashboard
	} else {
		r.Proyecto.Estado, r.Proyecto.Err = dashboard.Estado, dashboard.ErrEstado
		r.Metricas = SeccionDeMetricas{Velocidad: dashboard.Velocidad, Referencia: dashboard.Referencia, Err: dashboard.Velocidad.Err}
	}

	r.Backlog = c.backlog(ctx, proyectoID)
	r.Sprints = c.sprints(ctx, proyectoID)
	r.Esfuerzo = c.esfuerzo(ctx, proyectoID)
	r.Defectos = c.defectos(ctx, proyectoID, dashboard.Defectos, errDashboard)
	return r, nil
}

func (c *ConsultarReporteDelProyecto) backlog(ctx context.Context, proyectoID string) SeccionDelBacklog {
	items, err := c.fuentes.Backlog(ctx, proyectoID)
	if err != nil {
		return SeccionDelBacklog{Err: fmt.Errorf("listar el backlog del proyecto %s: %w", proyectoID, err)}
	}
	return SeccionDelBacklog{Items: items}
}

// sprints junta la lista de sprints (por numero descendente) con las historias que volvieron al backlog de cada sprint cerrado.
func (c *ConsultarReporteDelProyecto) sprints(ctx context.Context, proyectoID string) SeccionDeSprints {
	resumenes, err := c.fuentes.Sprints(ctx, proyectoID)
	if err != nil {
		return SeccionDeSprints{Err: fmt.Errorf("listar los sprints del proyecto %s: %w", proyectoID, err)}
	}
	anteriores, err := c.fuentes.Anteriores(ctx, proyectoID)
	if err != nil {
		return SeccionDeSprints{Err: fmt.Errorf("leer los sprints cerrados del proyecto %s: %w", proyectoID, err)}
	}
	volvieron := make(map[string]int, len(anteriores.Sprints))
	for _, a := range anteriores.Sprints {
		volvieron[a.Sprint.ID] = a.HistoriasNoCompletadas
	}
	sprints := make([]SprintDelReporte, len(resumenes))
	for i, s := range resumenes {
		sprints[i] = SprintDelReporte{ResumenDeSprint: s}
		if s.Sprint.Estado == sprint.EstadoCerrado {
			sprints[i].VolvieronAlBacklog = volvieron[s.Sprint.ID]
		}
	}
	return SeccionDeSprints{Sprints: sprints}
}

func (c *ConsultarReporteDelProyecto) esfuerzo(ctx context.Context, proyectoID string) SeccionDeEsfuerzo {
	consulta, err := c.fuentes.Esfuerzo(ctx, proyectoID, FiltroEsfuerzo{})
	if err != nil {
		return SeccionDeEsfuerzo{Err: fmt.Errorf("consultar el esfuerzo del proyecto %s: %w", proyectoID, err)}
	}
	comparacion, err := c.fuentes.Comparacion(ctx, proyectoID)
	if err != nil {
		return SeccionDeEsfuerzo{Err: fmt.Errorf("comparar el esfuerzo del proyecto %s: %w", proyectoID, err)}
	}
	return SeccionDeEsfuerzo{TotalHoras: consulta.TotalHoras, Comparacion: comparacion}
}

// defectos trae la lista (sin filtro: todos los estados) y los totales que calculo el dashboard (RN-030-7).
func (c *ConsultarReporteDelProyecto) defectos(ctx context.Context, proyectoID string, totales TarjetaDeDefectos, errDashboard error) SeccionDeDefectos {
	lista, err := c.fuentes.Defectos(ctx, proyectoID, FiltroDefectos{})
	if err != nil {
		return SeccionDeDefectos{Err: fmt.Errorf("listar los defectos del proyecto %s: %w", proyectoID, err)}
	}
	if errDashboard != nil {
		return SeccionDeDefectos{Lista: lista, Err: errDashboard}
	}
	return SeccionDeDefectos{Lista: lista, Totales: totales, Err: totales.Err}
}
