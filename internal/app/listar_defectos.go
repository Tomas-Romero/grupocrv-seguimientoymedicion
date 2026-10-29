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

// FiltroDefectos son los filtros opcionales de la lista (spec US-024, seccion 3). El valor vacio de
// cada campo significa "sin filtro", como en FiltroEsfuerzo (US-020).
type FiltroDefectos struct {
	SprintDeteccionID string
	Severidad         defecto.Severidad
}

// RepositorioListarDefectos es lo que ListarDefectos necesita de la persistencia.
type RepositorioListarDefectos interface {
	// ExisteProyecto dice si el proyecto existe (es la lectura de US-005).
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)
	// ExisteSprintEnProyecto dice si el sprint existe y es de ese proyecto (RN-024-7): el mismo
	// metodo, con el mismo significado, que ya cumplen los repositorios de US-022.
	ExisteSprintEnProyecto(ctx context.Context, proyectoID, sprintID string) (bool, error)
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
// ocultar ninguno por su estado (RN-T021-3), y solo los que cumplen el filtro (US-024). Un proyecto que
// no existe es ErrProyectoInexistente; uno sin defectos, o un filtro sin coincidencias, una lista vacia
// (RN-024-5). No modifica nada.
//
// RN-024-9: los rechazos se hacen en este orden (proyecto, sprint y severidad), se devuelve el primero
// solo, sin unirlo con los demas, y antes de leer la lista. Es el criterio de RN-022-5: la existencia se
// verifica antes que los datos.
//
// Los repositorios devuelven los datos sin ordenar y el orden y el filtro los pone el caso de uso, como
// en T-011 y T-015, para que la memoria y Postgres no puedan ordenar ni filtrar distinto.
func (c *ListarDefectos) Ejecutar(ctx context.Context, proyectoID string, filtro FiltroDefectos) ([]DefectoListado, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return nil, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}

	// RN-024-7: el sprint vacio es "sin filtro" y no se verifica. Uno informado tiene que ser del proyecto;
	// el repositorio responde que no existe (no que fallo) si es de otro, inexistente o esta mal formado.
	if filtro.SprintDeteccionID != "" {
		existe, err = c.repo.ExisteSprintEnProyecto(ctx, proyectoID, filtro.SprintDeteccionID)
		if err != nil {
			return nil, fmt.Errorf("consultar el sprint %s del proyecto %s: %w", filtro.SprintDeteccionID, proyectoID, err)
		}
		if !existe {
			return nil, fmt.Errorf("%w (id %s)", ErrSprintInexistente, filtro.SprintDeteccionID)
		}
	}
	// RN-024-8: la severidad vacia tambien es "sin filtro" (el dominio la rechazaria, como en el alta), asi
	// que el caso de uso decide eso antes de llamar a la regla compartida con el alta.
	if filtro.Severidad != "" {
		if err = defecto.ValidarSeveridad(filtro.Severidad); err != nil {
			return nil, err
		}
	}

	defectos, err := c.repo.ListarDefectosDelProyecto(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar los defectos del proyecto %s: %w", proyectoID, err)
	}

	// RN-024-2 a RN-024-4: se filtra por el sprint de DETECCION (el de resolucion no cuenta) y por la
	// severidad, combinados con Y. El filtro solo quita defectos: no cambia los que quedan (RN-024-6).
	listados := make([]DefectoListado, 0, len(defectos))
	for _, d := range defectos {
		if filtro.SprintDeteccionID != "" && d.Defecto.SprintDeteccionID != filtro.SprintDeteccionID {
			continue
		}
		if filtro.Severidad != "" && d.Defecto.Severidad != filtro.Severidad {
			continue
		}
		listados = append(listados, d)
	}

	// CL-T021-8: dos defectos con la misma fecha de carga salen siempre en el mismo orden, por ID
	// descendente.
	sort.SliceStable(listados, func(i, j int) bool {
		if !listados[i].CreadoEn.Equal(listados[j].CreadoEn) {
			return listados[i].CreadoEn.After(listados[j].CreadoEn)
		}
		return listados[i].Defecto.ID > listados[j].Defecto.ID
	})
	return listados, nil
}
