package app

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// DefectoListado es un defecto de la lista, con los datos que la pantalla necesita y el defecto
// no guarda (spec T-021, seccion 3).
//
// SprintDeteccion y SprintResolucion son el numero del sprint (sprint.Numero) traducido desde el
// ID guardado (RN-022-6); SprintResolucion es 0 si el defecto no tiene sprint de resolucion.
// CreadoEn viaja en el modelo de lectura y no en Defecto: el orden lo necesita, el dominio no.
type DefectoListado struct {
	Defecto          defecto.Defecto
	HistoriaNumero   int
	HistoriaTitulo   string
	SprintDeteccion  int
	SprintResolucion int
	CreadoEn         time.Time
}

// RepositorioListarDefectos es lo que ListarDefectos necesita de la persistencia.
type RepositorioListarDefectos interface {
	// ExisteProyecto dice si el proyecto existe (es la lectura de US-005).
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)
	// ListarDefectosDelProyecto devuelve los defectos del proyecto, de cualquier estado, en
	// cualquier orden. Un proyecto sin defectos devuelve una lista vacia.
	ListarDefectosDelProyecto(ctx context.Context, proyectoID string) ([]DefectoListado, error)
}

// ListarDefectos es la lectura de la lista de defectos de un proyecto.
type ListarDefectos struct {
	repo RepositorioListarDefectos
}

// NuevoListarDefectos arma el caso de uso.
func NuevoListarDefectos(repo RepositorioListarDefectos) *ListarDefectos {
	return &ListarDefectos{repo: repo}
}

// Ejecutar devuelve los defectos del proyecto del mas reciente al mas antiguo (RN-T021-4), sin
// ocultar ninguno por su estado (RN-T021-3). Un proyecto que no existe es ErrProyectoInexistente;
// uno sin defectos, una lista vacia. No modifica nada.
//
// Los repositorios devuelven los datos sin ordenar y el orden lo pone el caso de uso, como en
// T-011 y T-015, para que la memoria y Postgres no puedan ordenar distinto.
func (c *ListarDefectos) Ejecutar(ctx context.Context, proyectoID string) ([]DefectoListado, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return nil, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}
	defectos, err := c.repo.ListarDefectosDelProyecto(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar los defectos del proyecto %s: %w", proyectoID, err)
	}

	// CL-T021-8: dos defectos con la misma fecha de carga salen siempre en el mismo orden, por
	// ID descendente.
	listados := append([]DefectoListado{}, defectos...)
	sort.SliceStable(listados, func(i, j int) bool {
		if !listados[i].CreadoEn.Equal(listados[j].CreadoEn) {
			return listados[i].CreadoEn.After(listados[j].CreadoEn)
		}
		return listados[i].Defecto.ID > listados[j].Defecto.ID
	})
	return listados, nil
}
