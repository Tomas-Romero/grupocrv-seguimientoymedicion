package httpadapter

import (
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/errores"
)

// entradaDelCatalogo une un error del dominio o de la aplicacion con su tipo y su campo
// (spec T-005, seccion 7). El nombre ("paquete.ErrAlgo") lo usa el test que comprueba que
// no falte ninguno (RN-T005-19).
type entradaDelCatalogo struct {
	nombre string
	err    error
	tipo   errores.Tipo
	// campo es el nombre del campo del formulario (RN-T005-8), o vacio si es un error general.
	campo string
}

// entradasDelCatalogo es la lista de errores conocidos. Una historia que agrega un error
// lo suma aca, en su propio PR: si lo olvida, el test del catalogo falla y lo nombra.
//
// Tipo, segun RN-T005-2: validacion si depende solo de lo que escribio la persona (o del
// reloj); conflicto si depende del estado guardado; no_encontrado si apunta a algo que
// no existe.
var entradasDelCatalogo = []entradaDelCatalogo{
	// backlog
	{"backlog.ErrTituloVacio", backlog.ErrTituloVacio, errores.Validacion, "titulo"},
	{"backlog.ErrTituloMuyLargo", backlog.ErrTituloMuyLargo, errores.Validacion, "titulo"},
	{"backlog.ErrPrioridadInvalida", backlog.ErrPrioridadInvalida, errores.Validacion, "prioridad"},
	{"backlog.ErrCriterioVacio", backlog.ErrCriterioVacio, errores.Validacion, "criterios"},
	{"backlog.ErrEstimacionFueraDeEscala", backlog.ErrEstimacionFueraDeEscala, errores.Validacion, "estimacion"},
	{"backlog.ErrItemCompletado", backlog.ErrItemCompletado, errores.Conflicto, ""},

	// esfuerzo
	{"esfuerzo.ErrFechaVacia", esfuerzo.ErrFechaVacia, errores.Validacion, "fecha"},
	{"esfuerzo.ErrFechaFutura", esfuerzo.ErrFechaFutura, errores.Validacion, "fecha"},
	{"esfuerzo.ErrActividadVacia", esfuerzo.ErrActividadVacia, errores.Validacion, "actividad"},
	{"esfuerzo.ErrActividadMuyLarga", esfuerzo.ErrActividadMuyLarga, errores.Validacion, "actividad"},
	{"esfuerzo.ErrHorasFueraDeRango", esfuerzo.ErrHorasFueraDeRango, errores.Validacion, "horas"},
	{"esfuerzo.ErrHorasFraccion", esfuerzo.ErrHorasFraccion, errores.Validacion, "horas"},

	// integrante
	{"integrante.ErrNombreVacio", integrante.ErrNombreVacio, errores.Validacion, "nombre"},
	{"integrante.ErrNombreLargo", integrante.ErrNombreLargo, errores.Validacion, "nombre"},
	{"integrante.ErrEmailVacio", integrante.ErrEmailVacio, errores.Validacion, "email"},
	{"integrante.ErrEmailInvalido", integrante.ErrEmailInvalido, errores.Validacion, "email"},
	{"integrante.ErrRolInvalido", integrante.ErrRolInvalido, errores.Validacion, "rol"},

	// planningpoker
	{"planningpoker.ErrSesionAbierta", planningpoker.ErrSesionAbierta, errores.Conflicto, ""},
	{"planningpoker.ErrParticipantesInsuficientes", planningpoker.ErrParticipantesInsuficientes, errores.Conflicto, ""},
	{"planningpoker.ErrVotacionCerrada", planningpoker.ErrVotacionCerrada, errores.Conflicto, ""},
	{"planningpoker.ErrNoEsParticipante", planningpoker.ErrNoEsParticipante, errores.Conflicto, ""},

	// proyecto
	{"proyecto.ErrNombreVacio", proyecto.ErrNombreVacio, errores.Validacion, "nombre"},
	{"proyecto.ErrNombreLargo", proyecto.ErrNombreLargo, errores.Validacion, "nombre"},
	{"proyecto.ErrFechaInicioFaltante", proyecto.ErrFechaInicioFaltante, errores.Validacion, "fecha_inicio"},
	{"proyecto.ErrFechaFinFaltante", proyecto.ErrFechaFinFaltante, errores.Validacion, "fecha_fin"},
	{"proyecto.ErrFechasIncoherentes", proyecto.ErrFechasIncoherentes, errores.Validacion, "fecha_fin"},

	// sprint
	{"sprint.ErrSprintGoalVacio", sprint.ErrSprintGoalVacio, errores.Validacion, "sprint_goal"},
	{"sprint.ErrSprintGoalMuyLargo", sprint.ErrSprintGoalMuyLargo, errores.Validacion, "sprint_goal"},
	{"sprint.ErrFechaInicioFaltante", sprint.ErrFechaInicioFaltante, errores.Validacion, "fecha_inicio"},
	{"sprint.ErrFechaFinFaltante", sprint.ErrFechaFinFaltante, errores.Validacion, "fecha_fin"},
	{"sprint.ErrFechasIncoherentes", sprint.ErrFechasIncoherentes, errores.Validacion, "fecha_fin"},
	{"sprint.ErrDuracionExcedida", sprint.ErrDuracionExcedida, errores.Validacion, "fecha_fin"},
	{"sprint.ErrFueraDelProyecto", sprint.ErrFueraDelProyecto, errores.Conflicto, ""},
	{"sprint.ErrUltimoSprintAbierto", sprint.ErrUltimoSprintAbierto, errores.Conflicto, ""},
	{"sprint.ErrInicioNoPosteriorAlUltimo", sprint.ErrInicioNoPosteriorAlUltimo, errores.Conflicto, "fecha_inicio"},
	{"sprint.ErrProyectoDistinto", sprint.ErrProyectoDistinto, errores.Conflicto, ""},
	{"sprint.ErrSprintCerrado", sprint.ErrSprintCerrado, errores.Conflicto, ""},
	{"sprint.ErrItemYaAsignado", sprint.ErrItemYaAsignado, errores.Conflicto, ""},
	{"sprint.ErrItemSinEstimar", sprint.ErrItemSinEstimar, errores.Conflicto, ""},
	{"sprint.ErrItemSinCriterios", sprint.ErrItemSinCriterios, errores.Conflicto, ""},
	{"sprint.ErrItemFueraDelSprint", sprint.ErrItemFueraDelSprint, errores.Conflicto, ""},

	// app
	{"app.ErrProyectoInexistente", app.ErrProyectoInexistente, errores.NoEncontrado, ""},
	{"app.ErrItemInexistente", app.ErrItemInexistente, errores.NoEncontrado, ""},
	{"app.ErrIntegranteInexistente", app.ErrIntegranteInexistente, errores.NoEncontrado, ""},
	{"app.ErrSprintInexistente", app.ErrSprintInexistente, errores.NoEncontrado, ""},
	{"app.ErrSesionInexistente", app.ErrSesionInexistente, errores.NoEncontrado, ""},
	{"app.ErrEmailDuplicado", app.ErrEmailDuplicado, errores.Conflicto, "email"},
}

// exclusionesDelCatalogo son los errores exportados que no se registran a proposito: salen
// como errores internos. Cada uno lleva el motivo.
var exclusionesDelCatalogo = map[string]string{
	"metricas.ErrStoryPointsNegativos": "describe datos propios del sistema que salieron mal, no algo que la persona escribio",
	"metricas.ErrResumenInconsistente": "describe datos propios del sistema que salieron mal, no algo que la persona escribio",
	"metricas.ErrNombreVacio":          "describe datos propios del sistema que salieron mal, no algo que la persona escribio",
	"metricas.ErrHorasInvalidas":       "describe datos propios del sistema que salieron mal (las horas ya entraron validadas por US-019), no algo que la persona escribio",
	"planningpoker.ErrVotosOcultos":    "solo aparece si un adaptador pide los valores antes de revelar: es un error de programacion (spec US-015, seccion 5)",
}

// CatalogoDeErrores arma el catalogo de errores conocidos de la aplicacion.
func CatalogoDeErrores() *errores.Catalogo {
	reglas := make([]errores.Regla, len(entradasDelCatalogo))
	for i, e := range entradasDelCatalogo {
		reglas[i] = errores.Regla{Err: e.err, Tipo: e.tipo, Campo: e.campo}
	}
	return errores.NuevoCatalogo(reglas...)
}
