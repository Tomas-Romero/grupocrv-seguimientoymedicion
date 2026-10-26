package app

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// FiltroEsfuerzo acota la consulta de esfuerzo (US-020). Un campo vacio no filtra; los que vienen
// completos se combinan con Y (RN-020-2).
type FiltroEsfuerzo struct {
	ItemID       string
	IntegranteID string
	SprintID     string
}

// LecturaEsfuerzo es todo lo que la consulta necesita del proyecto, leido de una sola vez para que
// los registros, los nombres y los sprints sean de la misma foto.
type LecturaEsfuerzo struct {
	// Registros son todos los del proyecto, sin ordenar.
	Registros []esfuerzo.RegistroEsfuerzo
	// Historias son todas las historias del proyecto: ID -> titulo.
	Historias map[string]string
	// Integrantes son todos los integrantes del proyecto: ID -> nombre.
	Integrantes map[string]string
	// Sprints son todos los del proyecto.
	Sprints []sprint.Sprint
}

// RepositorioConsultarEsfuerzo es lo que el caso de uso de US-020 necesita de la persistencia.
type RepositorioConsultarEsfuerzo interface {
	// ExisteProyecto dice si el proyecto existe. Un ID mal formado es un proyecto inexistente,
	// no un error.
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)

	// LeerEsfuerzo devuelve los registros, las historias, los integrantes y los sprints del
	// proyecto. Es una sola lectura consistente.
	LeerEsfuerzo(ctx context.Context, proyectoID string) (LecturaEsfuerzo, error)
}

// RegistroConsultado es un registro con los datos que la pantalla muestra.
type RegistroConsultado struct {
	Registro   esfuerzo.RegistroEsfuerzo
	Historia   string
	Integrante string
	// SprintNumero es el sprint donde cae la fecha del registro; 0 si no cae en ninguno.
	SprintNumero int
}

// TotalPorHistoria son las horas de una historia en la consulta.
type TotalPorHistoria struct {
	ItemID string
	Titulo string
	Horas  float64
}

// TotalPorIntegrante son las horas de un integrante en la consulta.
type TotalPorIntegrante struct {
	IntegranteID string
	Nombre       string
	Horas        float64
}

// TotalPorSprint son las horas de un sprint en la consulta. SprintID vacio y Numero 0 son las
// horas cuya fecha cae fuera de todo sprint.
type TotalPorSprint struct {
	SprintID string
	Numero   int
	Horas    float64
}

// ConsultaEsfuerzo es el resultado de US-020.
type ConsultaEsfuerzo struct {
	Registros     []RegistroConsultado
	TotalHoras    float64
	PorHistoria   []TotalPorHistoria
	PorIntegrante []TotalPorIntegrante
	PorSprint     []TotalPorSprint
}

// Horas devuelve las horas de los registros, en el mismo orden. Es lo que recibe
// metricas.CalcularHoras (US-027) cuando la consulta se filtra por sprint.
func (c ConsultaEsfuerzo) Horas() []float64 {
	horas := make([]float64, len(c.Registros))
	for i, r := range c.Registros {
		horas[i] = r.Registro.Horas
	}
	return horas
}

// ConsultarEsfuerzo es el caso de uso de lectura de US-020.
type ConsultarEsfuerzo struct {
	repo RepositorioConsultarEsfuerzo
}

// NuevoConsultarEsfuerzo arma el caso de uso.
func NuevoConsultarEsfuerzo(repo RepositorioConsultarEsfuerzo) *ConsultarEsfuerzo {
	return &ConsultarEsfuerzo{repo: repo}
}

// Ejecutar verifica el proyecto, lee todo lo del proyecto, verifica los filtros (historia,
// integrante y sprint, en ese orden) y devuelve los registros que los cumplen, ordenados, con sus
// totales. El primer error de inexistencia se devuelve solo, sin unirse con otros (seccion 2 de
// la spec).
func (c *ConsultarEsfuerzo) Ejecutar(ctx context.Context, proyectoID string, filtro FiltroEsfuerzo) (ConsultaEsfuerzo, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return ConsultaEsfuerzo{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return ConsultaEsfuerzo{}, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}

	lectura, err := c.repo.LeerEsfuerzo(ctx, proyectoID)
	if err != nil {
		return ConsultaEsfuerzo{}, fmt.Errorf("leer el esfuerzo del proyecto %s: %w", proyectoID, err)
	}

	if err := verificarFiltro(filtro, lectura); err != nil {
		return ConsultaEsfuerzo{}, err
	}

	rangos := make([]esfuerzo.RangoSprint, len(lectura.Sprints))
	for i, s := range lectura.Sprints {
		rangos[i] = esfuerzo.RangoSprint{ID: s.ID, Numero: s.Numero, Inicio: s.FechaInicio, Fin: s.FechaFin}
	}

	var consulta ConsultaEsfuerzo
	var filtrados []esfuerzo.RegistroEsfuerzo
	sprintDe := map[string]esfuerzo.RangoSprint{}
	for _, r := range lectura.Registros {
		rango, _ := esfuerzo.SprintDeLaFecha(r.Fecha, rangos)
		if !cumple(filtro, r, rango) {
			continue
		}
		sprintDe[r.ID] = rango
		filtrados = append(filtrados, r)
		consulta.Registros = append(consulta.Registros, RegistroConsultado{
			Registro:     r,
			Historia:     lectura.Historias[r.ItemID],
			Integrante:   lectura.Integrantes[r.IntegranteID],
			SprintNumero: rango.Numero,
		})
	}

	ordenarRegistros(consulta.Registros)
	consulta.TotalHoras = esfuerzo.TotalHoras(filtrados)
	consulta.PorHistoria = totalesPorHistoria(filtrados, lectura.Historias)
	consulta.PorIntegrante = totalesPorIntegrante(filtrados, lectura.Integrantes)
	consulta.PorSprint = totalesPorSprint(filtrados, sprintDe)
	return consulta, nil
}

// verificarFiltro comprueba que la historia, el integrante y el sprint del filtro sean del
// proyecto, en ese orden. Los que no vienen no se verifican.
func verificarFiltro(filtro FiltroEsfuerzo, lectura LecturaEsfuerzo) error {
	if _, ok := lectura.Historias[filtro.ItemID]; filtro.ItemID != "" && !ok {
		return fmt.Errorf("%w (id %s)", ErrItemInexistente, filtro.ItemID)
	}
	if _, ok := lectura.Integrantes[filtro.IntegranteID]; filtro.IntegranteID != "" && !ok {
		return fmt.Errorf("%w (id %s)", ErrIntegranteInexistente, filtro.IntegranteID)
	}
	if filtro.SprintID != "" && !slices.ContainsFunc(lectura.Sprints, func(s sprint.Sprint) bool { return s.ID == filtro.SprintID }) {
		return fmt.Errorf("%w (id %s)", ErrSprintInexistente, filtro.SprintID)
	}
	return nil
}

// cumple dice si el registro pasa todos los filtros que vienen completos (RN-020-2). rango es el
// sprint de la fecha del registro, vacio si cae fuera de todos.
func cumple(filtro FiltroEsfuerzo, r esfuerzo.RegistroEsfuerzo, rango esfuerzo.RangoSprint) bool {
	return (filtro.ItemID == "" || r.ItemID == filtro.ItemID) &&
		(filtro.IntegranteID == "" || r.IntegranteID == filtro.IntegranteID) &&
		(filtro.SprintID == "" || rango.ID == filtro.SprintID)
}

// compararTexto compara sin distinguir mayusculas.
func compararTexto(a, b string) int {
	return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
}

// ordenarRegistros aplica RN-020-5: fecha descendente y, con la misma fecha, titulo de la
// historia, nombre del integrante, actividad e ID, todos ascendentes.
func ordenarRegistros(registros []RegistroConsultado) {
	slices.SortStableFunc(registros, func(a, b RegistroConsultado) int {
		return cmp.Or(
			b.Registro.Fecha.Compare(a.Registro.Fecha),
			compararTexto(a.Historia, b.Historia),
			compararTexto(a.Integrante, b.Integrante),
			compararTexto(a.Registro.Actividad, b.Registro.Actividad),
			cmp.Compare(a.Registro.ID, b.Registro.ID),
		)
	})
}

// agrupar reparte los registros por la clave que devuelve clave, en el orden en que aparecen.
func agrupar(registros []esfuerzo.RegistroEsfuerzo, clave func(esfuerzo.RegistroEsfuerzo) string) (claves []string, grupos map[string][]esfuerzo.RegistroEsfuerzo) {
	grupos = map[string][]esfuerzo.RegistroEsfuerzo{}
	for _, r := range registros {
		k := clave(r)
		if _, visto := grupos[k]; !visto {
			claves = append(claves, k)
		}
		grupos[k] = append(grupos[k], r)
	}
	return claves, grupos
}

// totalesPorHistoria aplica RN-020-8 y RN-020-9: de mas horas a menos y, con las mismas, por
// titulo y despues por ID.
func totalesPorHistoria(registros []esfuerzo.RegistroEsfuerzo, titulos map[string]string) []TotalPorHistoria {
	ids, grupos := agrupar(registros, func(r esfuerzo.RegistroEsfuerzo) string { return r.ItemID })
	var totales []TotalPorHistoria
	for _, id := range ids {
		totales = append(totales, TotalPorHistoria{ItemID: id, Titulo: titulos[id], Horas: esfuerzo.TotalHoras(grupos[id])})
	}
	slices.SortStableFunc(totales, func(a, b TotalPorHistoria) int {
		return cmp.Or(cmp.Compare(b.Horas, a.Horas), compararTexto(a.Titulo, b.Titulo), cmp.Compare(a.ItemID, b.ItemID))
	})
	return totales
}

// totalesPorIntegrante aplica RN-020-8 y RN-020-9, igual que totalesPorHistoria.
func totalesPorIntegrante(registros []esfuerzo.RegistroEsfuerzo, nombres map[string]string) []TotalPorIntegrante {
	ids, grupos := agrupar(registros, func(r esfuerzo.RegistroEsfuerzo) string { return r.IntegranteID })
	var totales []TotalPorIntegrante
	for _, id := range ids {
		totales = append(totales, TotalPorIntegrante{IntegranteID: id, Nombre: nombres[id], Horas: esfuerzo.TotalHoras(grupos[id])})
	}
	slices.SortStableFunc(totales, func(a, b TotalPorIntegrante) int {
		return cmp.Or(cmp.Compare(b.Horas, a.Horas), compararTexto(a.Nombre, b.Nombre), cmp.Compare(a.IntegranteID, b.IntegranteID))
	})
	return totales
}

// totalesPorSprint aplica RN-020-8 y RN-020-9: por numero de sprint, y las horas que caen fuera
// de todo sprint (numero 0) al final.
func totalesPorSprint(registros []esfuerzo.RegistroEsfuerzo, sprintDe map[string]esfuerzo.RangoSprint) []TotalPorSprint {
	ids, grupos := agrupar(registros, func(r esfuerzo.RegistroEsfuerzo) string { return sprintDe[r.ID].ID })
	var totales []TotalPorSprint
	for _, id := range ids {
		numero := sprintDe[grupos[id][0].ID].Numero
		totales = append(totales, TotalPorSprint{SprintID: id, Numero: numero, Horas: esfuerzo.TotalHoras(grupos[id])})
	}
	slices.SortStableFunc(totales, func(a, b TotalPorSprint) int {
		if (a.SprintID == "") != (b.SprintID == "") {
			if a.SprintID == "" {
				return 1
			}
			return -1
		}
		return cmp.Compare(a.Numero, b.Numero)
	})
	return totales
}
