package app

import (
	"context"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// RepositorioDashboard es lo que ConsultarDashboard necesita de la persistencia (spec US-030, seccion 2): las tres lecturas del
// estado general (proyecto, historias y sprints con historias), los registros de esfuerzo (US-020) y los defectos con sus
// numeros de sprint (T-021). Todas ya existen en Postgres y en la base en memoria: no hay consultas nuevas.
type RepositorioDashboard interface {
	RepositorioEstadoGeneral
	// LeerEsfuerzo devuelve los registros de esfuerzo del proyecto (y sus nombres y sprints), sin ordenar.
	LeerEsfuerzo(ctx context.Context, proyectoID string) (LecturaEsfuerzo, error)
	// ListarDefectosDelProyecto devuelve los defectos del proyecto, de cualquier estado y en cualquier orden.
	ListarDefectosDelProyecto(ctx context.Context, proyectoID string) ([]DefectoListado, error)
}

// Dashboard son las tarjetas y los graficos del dashboard de un proyecto. No calcula nada nuevo: junta los resultados de las
// funciones de metricas (US-025 a US-029) y el estado general (US-004). Cada tarjeta lleva su propio error: si un calculo falla,
// solo falla esa tarjeta (RN-030-9).
type Dashboard struct {
	Estado    EstadoGeneral
	ErrEstado error
	// Referencia es la tarjeta del sprint abierto o, si no hay, del ultimo cerrado; nil si el proyecto no tiene sprints.
	Referencia *TarjetaDeSprint
	Velocidad  TarjetaDeVelocidad
	Defectos   TarjetaDeDefectos
	Graficos   GraficosDelDashboard
}

// TarjetaDeSprint es el sprint de referencia con sus numeros.
type TarjetaDeSprint struct {
	Sprint                    sprint.Sprint
	Planificados, Completados int
	PorcentajeHistorias       float64
	Horas                     metricas.ResumenHoras
	Err                       error
}

// TarjetaDeVelocidad es la velocidad del equipo y la tabla de los sprints cerrados.
type TarjetaDeVelocidad struct {
	Velocidad float64
	// Cerrados van del mas reciente al mas antiguo, una fila por sprint cerrado (RN-030-8).
	Cerrados []SprintCerradoDelDashboard
	Err      error
}

// SprintCerradoDelDashboard es una fila de la tabla de sprints cerrados.
type SprintCerradoDelDashboard struct {
	Sprint                    sprint.Sprint
	Planificados, Completados int
	Defectos                  metricas.ResumenDefectos
}

// ConteoPorSeveridad y ConteoPorEstado cuentan los defectos del proyecto (RN-030-7).
type ConteoPorSeveridad struct{ Critica, Alta, Media, Baja int }

type ConteoPorEstado struct{ Abierto, EnCorreccion, Resuelto, Descartado int }

// TarjetaDeDefectos cuenta todos los defectos del proyecto, tambien los descartados, por severidad y por estado.
type TarjetaDeDefectos struct {
	Total        int
	PorSeveridad ConteoPorSeveridad
	PorEstado    ConteoPorEstado
	Err          error
}

// GraficosDelDashboard se arman con los mismos datos de las tarjetas (RN-030-14 a RN-030-17).
type GraficosDelDashboard struct {
	// StoryPoints tiene un grupo por sprint cerrado, del mas antiguo al mas reciente. Sigue la suerte de la tarjeta de velocidad.
	StoryPoints    []BarrasDeSprint
	ErrStoryPoints error
	// Completadas es el porcentaje del sprint de referencia; nil si no hay o si su tarjeta fallo.
	Completadas *BarraDePorcentaje
}

// BarrasDeSprint son las dos barras de un sprint cerrado. Las alturas son porcentajes de 0 a 100 de la barra mas alta del grafico.
type BarrasDeSprint struct {
	Sprint                                int
	Planificados, Completados             float64
	AlturaPlanificados, AlturaCompletados float64
}

// BarraDePorcentaje es el porcentaje de historias completadas del sprint de referencia.
type BarraDePorcentaje struct {
	Sprint     int
	Porcentaje float64
}

// ConsultarDashboard es el caso de uso de lectura de US-030.
type ConsultarDashboard struct {
	repo  RepositorioDashboard
	ahora func() time.Time
}

// NuevoConsultarDashboard arma el caso de uso. ahora es el reloj del estado general (US-004).
func NuevoConsultarDashboard(repo RepositorioDashboard, ahora func() time.Time) *ConsultarDashboard {
	return &ConsultarDashboard{repo: repo, ahora: ahora}
}

// Ejecutar verifica el proyecto (si no existe es ErrProyectoInexistente y no se lee nada mas, RN-030-10), hace UNA lectura de
// cada tipo (RN-030-11) y calcula cada tarjeta por separado con las funciones de metricas (RN-030-9). No modifica nada.
func (c *ConsultarDashboard) Ejecutar(ctx context.Context, proyectoID string) (Dashboard, error) {
	p, err := c.repo.BuscarProyecto(ctx, proyectoID)
	if err != nil {
		return Dashboard{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	items, err := c.repo.ListarItems(ctx, proyectoID)
	if err != nil {
		return Dashboard{}, fmt.Errorf("listar las historias del proyecto %s: %w", proyectoID, err)
	}
	sprints, err := c.repo.ListarSprintsConItems(ctx, proyectoID)
	if err != nil {
		return Dashboard{}, fmt.Errorf("listar los sprints del proyecto %s: %w", proyectoID, err)
	}
	lectura, err := c.repo.LeerEsfuerzo(ctx, proyectoID)
	if err != nil {
		return Dashboard{}, fmt.Errorf("leer el esfuerzo del proyecto %s: %w", proyectoID, err)
	}
	defectos, err := c.repo.ListarDefectosDelProyecto(ctx, proyectoID)
	if err != nil {
		return Dashboard{}, fmt.Errorf("listar los defectos del proyecto %s: %w", proyectoID, err)
	}

	var d Dashboard
	d.Estado, d.ErrEstado = armarEstadoGeneral(p, items, sprints, c.ahora())
	if d.ErrEstado != nil {
		d.ErrEstado = fmt.Errorf("calcular el estado del proyecto %s: %w", proyectoID, d.ErrEstado)
	}
	d.Referencia = tarjetaDelSprint(sprints, lectura.Registros)
	d.Velocidad = tarjetaDeVelocidad(sprints, defectos)
	d.Defectos = tarjetaDeDefectos(defectos)
	d.Graficos = graficos(d.Velocidad, d.Referencia)
	return d, nil
}

// historiasDelSprint son las historias de un sprint como las pide metricas, con la regla de HistoriaAsignada.Completada: en un
// sprint cerrado manda el resultado de su fila, y en uno abierto, el estado actual (RN-030-2).
func historiasDelSprint(s SprintConItems) []metricas.ItemDelSprint {
	datos := make([]metricas.ItemDelSprint, len(s.Items))
	for i, h := range s.Items {
		puntos, _ := h.Item.StoryPoints.Valor()
		datos[i] = metricas.ItemDelSprint{StoryPoints: puntos, Completado: h.Completada()}
	}
	return datos
}

// tarjetaDelSprint arma la tarjeta del sprint de referencia (RN-030-1): el abierto o, si no hay, el cerrado de numero mas alto.
func tarjetaDelSprint(sprints []SprintConItems, registros []esfuerzo.RegistroEsfuerzo) *TarjetaDeSprint {
	if len(sprints) == 0 {
		return nil
	}
	referencia := sprints[0]
	if abierto := slices.IndexFunc(sprints, func(s SprintConItems) bool { return s.Sprint.Estado == sprint.EstadoAbierto }); abierto >= 0 {
		referencia = sprints[abierto]
	} else {
		for _, s := range sprints {
			if s.Sprint.Numero > referencia.Sprint.Numero {
				referencia = s
			}
		}
	}

	t := &TarjetaDeSprint{Sprint: referencia.Sprint}
	datos := historiasDelSprint(referencia)
	resumen, err := metricas.CalcularResumenSprint(datos)
	if err != nil {
		t.Err = fmt.Errorf("calcular los Story Points del %s: %w", referencia.Sprint.Nombre(), err)
		return t
	}
	t.Planificados, t.Completados = resumen.Planificados, resumen.Completados
	if t.PorcentajeHistorias, err = metricas.CalcularPorcentajeCompletadas(datos); err != nil {
		t.Err = fmt.Errorf("calcular el porcentaje de historias completadas del %s: %w", referencia.Sprint.Nombre(), err)
		return t
	}

	// RN-030-4: las horas reales son las de los registros cuya fecha cae en el rango del sprint, de cualquier historia e integrante.
	rangos := make([]esfuerzo.RangoSprint, len(sprints))
	for i, s := range sprints {
		rangos[i] = esfuerzo.RangoSprint{ID: s.Sprint.ID, Numero: s.Sprint.Numero, Inicio: s.Sprint.FechaInicio, Fin: s.Sprint.FechaFin}
	}
	var horas []float64
	for _, r := range registros {
		if rango, ok := esfuerzo.SprintDeLaFecha(r.Fecha, rangos); ok && rango.ID == referencia.Sprint.ID {
			horas = append(horas, r.Horas)
		}
	}
	if t.Horas, err = metricas.CalcularHoras(resumen.Planificados, horas); err != nil {
		t.Err = fmt.Errorf("calcular las horas del %s: %w", referencia.Sprint.Nombre(), err)
	}
	return t
}

// tarjetaDeVelocidad arma la velocidad (RN-030-5) y una fila por sprint cerrado, del mas reciente al mas antiguo (RN-030-8), con
// sus defectos detectados y resueltos (RN-030-6).
func tarjetaDeVelocidad(sprints []SprintConItems, defectos []DefectoListado) TarjetaDeVelocidad {
	var cerrados []SprintConItems
	for _, s := range sprints {
		if s.Sprint.Estado == sprint.EstadoCerrado {
			cerrados = append(cerrados, s)
		}
	}
	slices.SortFunc(cerrados, func(a, b SprintConItems) int { return b.Sprint.Numero - a.Sprint.Numero })

	// RN-030-6: los descartados no entran en la lista que se le pasa a metricas (RN-023-13): no son defectos que haya que contar.
	var registrados []metricas.DefectoRegistrado
	for _, d := range defectos {
		if d.Defecto.Estado != defecto.EstadoDescartado {
			registrados = append(registrados, metricas.DefectoRegistrado{SprintDeteccion: d.SprintDeteccion, SprintResolucion: d.SprintResolucion})
		}
	}

	var t TarjetaDeVelocidad
	var paraLaVelocidad []metricas.SprintCerrado
	for _, s := range cerrados {
		resumen, err := metricas.CalcularResumenSprint(historiasDelSprint(s))
		if err != nil {
			t.Err = fmt.Errorf("calcular los Story Points del %s: %w", s.Sprint.Nombre(), err)
			return t
		}
		porDefectos, err := metricas.CalcularDefectosDelSprint(s.Sprint.Numero, registrados)
		if err != nil {
			t.Err = fmt.Errorf("calcular los defectos del %s: %w", s.Sprint.Nombre(), err)
			return t
		}
		t.Cerrados = append(t.Cerrados, SprintCerradoDelDashboard{Sprint: s.Sprint, Planificados: resumen.Planificados, Completados: resumen.Completados, Defectos: porDefectos})
		paraLaVelocidad = append(paraLaVelocidad, metricas.SprintCerrado{Nombre: s.Sprint.Nombre(), Resumen: resumen})
	}
	velocidad, err := metricas.CalcularVelocidad(paraLaVelocidad)
	if err != nil {
		t.Err = fmt.Errorf("calcular la velocidad del equipo: %w", err)
		return TarjetaDeVelocidad{Err: t.Err}
	}
	t.Velocidad = velocidad
	return t
}

// tarjetaDeDefectos cuenta todos los defectos del proyecto, tambien los descartados (RN-030-7).
func tarjetaDeDefectos(defectos []DefectoListado) TarjetaDeDefectos {
	var t TarjetaDeDefectos
	for _, d := range defectos {
		t.Total++
		switch d.Defecto.Severidad {
		case defecto.SeveridadCritica:
			t.PorSeveridad.Critica++
		case defecto.SeveridadAlta:
			t.PorSeveridad.Alta++
		case defecto.SeveridadMedia:
			t.PorSeveridad.Media++
		case defecto.SeveridadBaja:
			t.PorSeveridad.Baja++
		}
		switch d.Defecto.Estado {
		case defecto.EstadoAbierto:
			t.PorEstado.Abierto++
		case defecto.EstadoEnCorreccion:
			t.PorEstado.EnCorreccion++
		case defecto.EstadoResuelto:
			t.PorEstado.Resuelto++
		case defecto.EstadoDescartado:
			t.PorEstado.Descartado++
		}
	}
	return t
}

// graficos arma los dos graficos con los datos de las tarjetas, sin recalcular nada (RN-030-14 a RN-030-17). El de Story Points
// va en orden cronologico y con una escala comun: la barra mas alta mide 100 y las demas, valor / maximo * 100; si el maximo es
// 0, todas miden 0.
func graficos(velocidad TarjetaDeVelocidad, referencia *TarjetaDeSprint) GraficosDelDashboard {
	var g GraficosDelDashboard
	if velocidad.Err != nil {
		g.ErrStoryPoints = velocidad.Err
	} else {
		maximo := 0
		for _, f := range velocidad.Cerrados {
			maximo = max(maximo, f.Planificados, f.Completados)
		}
		altura := func(valor int) float64 {
			if maximo == 0 {
				return 0
			}
			return math.Round(float64(valor)/float64(maximo)*1000) / 10
		}
		for i := len(velocidad.Cerrados) - 1; i >= 0; i-- {
			f := velocidad.Cerrados[i]
			g.StoryPoints = append(g.StoryPoints, BarrasDeSprint{
				Sprint: f.Sprint.Numero, Planificados: float64(f.Planificados), Completados: float64(f.Completados),
				AlturaPlanificados: altura(f.Planificados), AlturaCompletados: altura(f.Completados),
			})
		}
	}
	if referencia != nil && referencia.Err == nil {
		g.Completadas = &BarraDePorcentaje{Sprint: referencia.Sprint.Numero, Porcentaje: referencia.PorcentajeHistorias}
	}
	return g
}
