package app

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Reporte de un sprint (spec US-033). Igual que el del proyecto, no accede a repositorios: compone lecturas que ya existen.

// FuentesDelReporteDelSprint son las lecturas que el reporte de un sprint compone. Cada una se llama una sola vez.
type FuentesDelReporteDelSprint struct {
	Proyecto func(ctx context.Context, proyectoID string) (DetalleDeProyecto, error)
	Sprint   func(ctx context.Context, proyectoID, sprintID string) (DetalleDeSprint, error)
	Esfuerzo func(ctx context.Context, proyectoID string, filtro FiltroEsfuerzo) (ConsultaEsfuerzo, error)
	Defectos func(ctx context.Context, proyectoID string, filtro FiltroDefectos) ([]DefectoListado, error)
}

// ReporteDelSprint tiene las cinco secciones del reporte, siempre las mismas y en este orden (RN-033-2): el sprint, sus historias, las
// metricas, el esfuerzo registrado y los defectos. Las historias vienen de la misma lectura que el sprint y no tienen error propio.
type ReporteDelSprint struct {
	Sprint    sprint.Sprint
	Historias SeccionDeHistoriasDelSprint
	Metricas  SeccionDeMetricasDelSprint
	Esfuerzo  SeccionDeEsfuerzoDelSprint
	Defectos  SeccionDeDefectosDelSprint
}

// SeccionDeHistoriasDelSprint son las historias asignadas al sprint por prioridad y numero (RN-T015-4), cada una con el resultado de su
// fila en este sprint y no con su estado actual (RN-033-4).
type SeccionDeHistoriasDelSprint struct {
	Items []HistoriaDelSprint
}

// Vacia dice si el sprint no tiene historias.
func (s SeccionDeHistoriasDelSprint) Vacia() bool { return len(s.Items) == 0 }

// SeccionDeMetricasDelSprint son los Story Points, el porcentaje de historias completadas y las horas del sprint. Los Story Points y el
// porcentaje salen del propio sprint; las horas, del esfuerzo cuyas fechas caen dentro (RN-019-12). Si las horas no se pudieron leer, Err
// lo informa y los Story Points y el porcentaje siguen siendo validos.
type SeccionDeMetricasDelSprint struct {
	Planificados        int
	Completados         int
	PorcentajeHistorias float64
	Horas               metricas.ResumenHoras
	Err                 error
}

// SeccionDeEsfuerzoDelSprint son los registros de esfuerzo con fecha dentro del sprint y su total.
type SeccionDeEsfuerzoDelSprint struct {
	Registros  []RegistroConsultado
	TotalHoras float64
	Err        error
}

// Vacia dice si no hay horas registradas dentro del sprint.
func (s SeccionDeEsfuerzoDelSprint) Vacia() bool { return len(s.Registros) == 0 }

// SeccionDeDefectosDelSprint son los defectos detectados en el sprint y los resueltos en el (RN-033-5). Un defecto detectado y resuelto
// en el mismo sprint esta en las dos listas; los descartados no estan en ninguna (RN-030-6).
type SeccionDeDefectosDelSprint struct {
	Detectados []DefectoListado
	Resueltos  []DefectoListado
	Err        error
}

// Vacia dice si el sprint no tiene defectos detectados ni resueltos.
func (s SeccionDeDefectosDelSprint) Vacia() bool {
	return len(s.Detectados) == 0 && len(s.Resueltos) == 0
}

// ConsultarReporteDelSprint es el caso de uso de lectura del reporte de un sprint (US-033).
type ConsultarReporteDelSprint struct {
	fuentes FuentesDelReporteDelSprint
}

// NuevoConsultarReporteDelSprint arma el caso de uso con las lecturas que compone.
func NuevoConsultarReporteDelSprint(fuentes FuentesDelReporteDelSprint) *ConsultarReporteDelSprint {
	return &ConsultarReporteDelSprint{fuentes: fuentes}
}

// Ejecutar verifica el proyecto y despues el sprint, que tiene que ser de ese proyecto: si no existen es ErrProyectoInexistente o
// ErrSprintInexistente y no se lee nada mas (RN-033-6). Despues arma cada seccion por separado (RN-033-3).
func (c *ConsultarReporteDelSprint) Ejecutar(ctx context.Context, proyectoID, sprintID string) (ReporteDelSprint, error) {
	if _, err := c.fuentes.Proyecto(ctx, proyectoID); err != nil {
		return ReporteDelSprint{}, fmt.Errorf("armar el reporte del sprint %s: %w", sprintID, err)
	}
	detalle, err := c.fuentes.Sprint(ctx, proyectoID, sprintID)
	if err != nil {
		return ReporteDelSprint{}, fmt.Errorf("armar el reporte del sprint %s: %w", sprintID, err)
	}

	r := ReporteDelSprint{Sprint: detalle.Sprint, Historias: SeccionDeHistoriasDelSprint{Items: detalle.Historias}}
	r.Metricas.Planificados, r.Metricas.Completados = detalle.Planificados, detalle.Completados

	datos := make([]metricas.ItemDelSprint, len(detalle.Historias))
	for i, h := range detalle.Historias {
		puntos, _ := h.Item.StoryPoints.Valor()
		datos[i] = metricas.ItemDelSprint{StoryPoints: puntos, Completado: h.Completada}
	}
	porcentaje, err := metricas.CalcularPorcentajeCompletadas(datos)
	if err != nil {
		r.Metricas.Err = fmt.Errorf("calcular el porcentaje de historias completadas del %s: %w", detalle.Sprint.Nombre(), err)
	}
	r.Metricas.PorcentajeHistorias = porcentaje

	// Las horas del sprint son las de los registros con fecha dentro de su rango: la lectura de esfuerzo filtrada por sprint.
	consulta, err := c.fuentes.Esfuerzo(ctx, proyectoID, FiltroEsfuerzo{SprintID: sprintID})
	if err != nil {
		err = fmt.Errorf("consultar el esfuerzo del %s: %w", detalle.Sprint.Nombre(), err)
		r.Esfuerzo.Err = err
		r.Metricas.Err = primerError(r.Metricas.Err, err)
	} else {
		r.Esfuerzo.Registros, r.Esfuerzo.TotalHoras = consulta.Registros, consulta.TotalHoras
		horas, err := metricas.CalcularHoras(detalle.Planificados, consulta.Horas())
		if err != nil {
			r.Metricas.Err = primerError(r.Metricas.Err, fmt.Errorf("calcular las horas del %s: %w", detalle.Sprint.Nombre(), err))
		}
		r.Metricas.Horas = horas
	}

	lista, err := c.fuentes.Defectos(ctx, proyectoID, FiltroDefectos{})
	if err != nil {
		r.Defectos.Err = fmt.Errorf("listar los defectos del proyecto %s: %w", proyectoID, err)
		return r, nil
	}
	for _, d := range lista {
		if d.Defecto.Estado == defecto.EstadoDescartado {
			continue
		}
		if d.Defecto.SprintDeteccionID == sprintID {
			r.Defectos.Detectados = append(r.Defectos.Detectados, d)
		}
		if d.Defecto.Estado == defecto.EstadoResuelto && d.Defecto.SprintResolucionID == sprintID {
			r.Defectos.Resueltos = append(r.Defectos.Resueltos, d)
		}
	}
	return r, nil
}

// primerError conserva el primer error de una seccion: es el que explica por que no se pudo armar.
func primerError(primero, otro error) error {
	if primero != nil {
		return primero
	}
	return otro
}
