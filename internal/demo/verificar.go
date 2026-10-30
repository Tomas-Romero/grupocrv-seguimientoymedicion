package demo

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Lecturas son los casos de uso de lectura con los que se comprueba lo que quedo cargado.
type Lecturas struct {
	ListarBacklog     *app.ListarBacklog
	ListarSprints     *app.ListarSprints
	ConsultarEsfuerzo *app.ConsultarEsfuerzo
	ListarDefectos    *app.ListarDefectos
	ConsultarSesion   *app.ConsultarSesionPlanningPoker
}

// Cifras de la spec T-023 (seccion 3), escritas aca a proposito y no calculadas desde las tablas de datos.go: asi la comprobacion no
// puede darse por buena con unos datos que cambiaron sin cambiar la spec.
var (
	prioridadesEsperadas = map[backlog.Prioridad]int{
		backlog.PrioridadMust: 5, backlog.PrioridadShould: 6, backlog.PrioridadCould: 3, backlog.PrioridadWont: 2,
	}
	severidadesEsperadas = map[defecto.Severidad]int{
		defecto.SeveridadCritica: 2, defecto.SeveridadAlta: 2, defecto.SeveridadMedia: 3, defecto.SeveridadBaja: 3,
	}
	estadosDeDefectoEsperados = map[defecto.Estado]int{
		defecto.EstadoAbierto: 4, defecto.EstadoEnCorreccion: 1, defecto.EstadoResuelto: 4, defecto.EstadoDescartado: 1,
	}
	// Por numero de sprint: estado, Story Points planificados y completados, y horas reales.
	sprintsEsperados = []struct {
		numero                    int
		estado                    sprint.Estado
		planificados, completados int
		horas                     float64
	}{
		{1, sprint.EstadoCerrado, 13, 10, 24.5},
		{2, sprint.EstadoCerrado, 11, 8, 14.75},
		{3, sprint.EstadoCerrado, 14, 11, 21},
		{4, sprint.EstadoAbierto, 12, 5, 11.5},
	}
)

const (
	historiasEsperadas   = 16
	completadasEsperadas = 8
	registrosEsperados   = 27
	defectosEsperados    = 10
)

// Verificar lee lo cargado con los casos de uso de lectura y lo compara con las cifras de la spec. Devuelve todas las diferencias juntas.
func Verificar(ctx context.Context, l Lecturas, r Resultado) error {
	var fallas []error
	falla := func(formato string, args ...any) { fallas = append(fallas, fmt.Errorf(formato, args...)) }

	items, err := l.ListarBacklog.Ejecutar(ctx, r.ProyectoID, app.FiltroBacklog{})
	if err != nil {
		return fmt.Errorf("leer el backlog: %w", err)
	}
	prioridades, completadas := map[backlog.Prioridad]int{}, 0
	for _, it := range items {
		prioridades[it.Prioridad]++
		if it.Estado == backlog.EstadoCompletado {
			completadas++
		}
	}
	if len(items) != historiasEsperadas {
		falla("el backlog tiene %d historias y se esperaban %d", len(items), historiasEsperadas)
	}
	for p, n := range prioridadesEsperadas {
		if prioridades[p] != n {
			falla("hay %d historias %s y se esperaban %d", prioridades[p], p, n)
		}
	}
	if completadas != completadasEsperadas {
		falla("hay %d historias completadas y se esperaban %d", completadas, completadasEsperadas)
	}

	lista, err := l.ListarSprints.Ejecutar(ctx, r.ProyectoID)
	if err != nil {
		return fmt.Errorf("leer los sprints: %w", err)
	}
	porNumero := map[int]app.ResumenDeSprint{}
	for _, s := range lista {
		porNumero[s.Sprint.Numero] = s
	}
	if len(lista) != len(sprintsEsperados) {
		falla("hay %d sprints y se esperaban %d", len(lista), len(sprintsEsperados))
	}
	for _, e := range sprintsEsperados {
		s, ok := porNumero[e.numero]
		switch {
		case !ok:
			falla("falta el Sprint %d", e.numero)
		case s.Sprint.Estado != e.estado || s.Planificados != e.planificados || s.Completados != e.completados:
			falla("el Sprint %d esta %s con %d planificados y %d completados, y se esperaba %s con %d y %d",
				e.numero, s.Sprint.Estado, s.Planificados, s.Completados, e.estado, e.planificados, e.completados)
		}
		if !ok {
			continue
		}
		c, err := l.ConsultarEsfuerzo.Ejecutar(ctx, r.ProyectoID, app.FiltroEsfuerzo{SprintID: s.Sprint.ID})
		if err != nil {
			return fmt.Errorf("leer las horas del Sprint %d: %w", e.numero, err)
		}
		if math.Abs(c.TotalHoras-e.horas) > 1e-9 {
			falla("el Sprint %d tiene %.2f h y se esperaban %.2f h", e.numero, c.TotalHoras, e.horas)
		}
	}
	todas, err := l.ConsultarEsfuerzo.Ejecutar(ctx, r.ProyectoID, app.FiltroEsfuerzo{})
	if err != nil {
		return fmt.Errorf("leer las horas: %w", err)
	}
	if len(todas.Registros) != registrosEsperados {
		falla("hay %d registros de horas y se esperaban %d", len(todas.Registros), registrosEsperados)
	}

	listados, err := l.ListarDefectos.Ejecutar(ctx, r.ProyectoID, app.FiltroDefectos{})
	if err != nil {
		return fmt.Errorf("leer los defectos: %w", err)
	}
	severidades, estados := map[defecto.Severidad]int{}, map[defecto.Estado]int{}
	for _, d := range listados {
		severidades[d.Defecto.Severidad]++
		estados[d.Defecto.Estado]++
	}
	if len(listados) != defectosEsperados {
		falla("hay %d defectos y se esperaban %d", len(listados), defectosEsperados)
	}
	for s, n := range severidadesEsperadas {
		if severidades[s] != n {
			falla("hay %d defectos %s y se esperaban %d", severidades[s], s, n)
		}
	}
	for e, n := range estadosDeDefectoEsperados {
		if estados[e] != n {
			falla("hay %d defectos %s y se esperaban %d", estados[e], e, n)
		}
	}

	sesion := func(historia int) (app.DetalleDeSesion, bool) {
		id, ok := r.Sesiones[historia]
		if !ok {
			falla("falta la sesion de Planning Poker de la historia %d", historia)
			return app.DetalleDeSesion{}, false
		}
		d, err := l.ConsultarSesion.Ejecutar(ctx, r.ProyectoID, id)
		if err != nil {
			falla("leer la sesion de la historia %d: %v", historia, err)
			return app.DetalleDeSesion{}, false
		}
		return d, true
	}
	if d, ok := sesion(historiaHistorial); ok {
		puntos, estimada := d.Historia.StoryPoints.Valor()
		if d.Sesion.Estado != planningpoker.EstadoFinalizada || !estimada || puntos != estimacionAcordadaHistorial {
			falla("la historia %d tiene la sesion %s y %d puntos (estimada: %t), y se esperaba finalizada con %d",
				historiaHistorial, d.Sesion.Estado, puntos, estimada, estimacionAcordadaHistorial)
		}
		if len(d.Anteriores) != 1 {
			falla("la historia %d tiene %d rondas anteriores y se esperaba 1", historiaHistorial, len(d.Anteriores))
		}
	}
	if d, ok := sesion(historiaCupones); ok {
		if d.Sesion.Estado != planningpoker.EstadoRevelada || !d.Revelada || !d.Resultado.Divergencia {
			falla("la sesion de la historia %d tendria que estar revelada y con divergencia", historiaCupones)
		}
	}
	if d, ok := sesion(historiaRecomendada); ok {
		if d.Sesion.Estado != planningpoker.EstadoVotando || len(d.Votaron) != len(votosRecomendada) || len(d.Faltan) != len(integrantes)-len(votosRecomendada) {
			falla("la sesion de la historia %d tendria que estar votando con %d votos y %d sin votar (hay %d y %d)",
				historiaRecomendada, len(votosRecomendada), len(integrantes)-len(votosRecomendada), len(d.Votaron), len(d.Faltan))
		}
	}

	if len(fallas) > 0 {
		return fmt.Errorf("lo cargado no coincide con la spec T-023: %w", errors.Join(fallas...))
	}
	return nil
}
