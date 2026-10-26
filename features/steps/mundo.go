// Package steps implementa los step definitions de los escenarios BDD.
//
// El catalogo de frases disponibles esta en docs/diccionario-steps.md y es el
// contrato entre quien escribe los .feature y quien implementa los steps. La
// decision de como se conectan al resto de la aplicacion esta en
// docs/adr/0002-integracion-bdd-godog.md.
package steps

import (
	"errors"
	"fmt"
	"time"
)

// ErrAreaSinConectar se devuelve cuando un escenario usa un area (proyectos,
// backlog, sprints o metricas) que todavia no esta conectada en
// features/servicios_test.go. Pasa si alguien escribe un escenario antes de que
// exista el dominio que lo soporta: preferimos un error explicito a un
// escenario en verde que no probo nada.
var ErrAreaSinConectar = errors.New("area sin conectar a los steps")

// areaSinConectar arma el error con el nombre del area, para que quien lo lea
// sepa que historia le falta conectar.
func areaSinConectar(area string) error {
	return fmt.Errorf("%w: %s (conectarla en features/servicios_test.go)", ErrAreaSinConectar, area)
}

// Cada area es la parte de la aplicacion que usa una familia de steps. Las
// implementa un adaptador sobre internal/app; los steps no conocen la base de
// datos ni HTTP.
//
// Estan separadas a proposito: cada historia conecta SOLO la que necesita. Un
// .feature de proyectos no exige que existan backlog, sprints ni metricas, asi
// que el CI no se pone rojo por trabajo de otra persona.
//
// Si una historia necesita un metodo que no esta aca, se agrega en el mismo PR
// que el escenario que lo usa.

// Proyectos es el area de US-001, US-002 y US-003.
type Proyectos interface {
	CrearProyecto(nombre string, inicio, fin time.Time) error
	ModificarNombreProyecto(actual, nuevo string) error
	ModificarFechasProyecto(nombre string, inicio, fin time.Time) error
	RegistrarIntegrante(proyecto, nombre, email, rol string) error
	ProyectoExiste(nombre string) (inicio, fin time.Time, err error)
	CantidadIntegrantes(proyecto string) (int, error)
}

// DatosHistoria son los datos con los que un escenario crea una historia. Las
// frases que no nombran descripcion ni criterios los dejan vacios.
type DatosHistoria struct {
	Titulo      string
	Descripcion string
	Prioridad   string
	Criterios   []string
}

// Backlog es el area de US-005, US-006 y US-013.
type Backlog interface {
	CrearHistoria(proyecto string, historia DatosHistoria) error
	// EditarHistoria reemplaza los cuatro datos de la historia (US-006). Despues
	// de editarla, los steps la nombran por su titulo nuevo.
	EditarHistoria(titulo string, historia DatosHistoria) error
	CambiarPrioridadHistoria(titulo, prioridad string) error
	EstimarHistoria(titulo string, puntos int) error
	CantidadHistoriasEnBacklog(proyecto string) (int, error)
	PrioridadDeHistoria(titulo string) (string, error)
	// PuntosDeHistoria devuelve estimada = false si la historia esta sin
	// estimar: sin estimar no es lo mismo que 0 puntos (RN-005-7).
	PuntosDeHistoria(titulo string) (puntos int, estimada bool, err error)
	NumeroDeHistoria(titulo string) (int, error)
	EstadoDeHistoria(titulo string) (string, error)
	CriteriosDeHistoria(titulo string) ([]string, error)
	HistoriaEstaEnBacklog(titulo string) (bool, error)
}

// Sprints es el area de US-008, US-009, US-010 y US-011.
type Sprints interface {
	// CrearSprint crea un sprint en el proyecto con su objetivo (el Sprint Goal) y
	// sus fechas, y devuelve el nombre que le asigno la aplicacion ("Sprint N"): la
	// persona no lo escribe (RN-008-12).
	CrearSprint(proyecto, objetivo string, inicio, fin time.Time) (nombre string, err error)
	// AsignarHistoriaASprint asigna la historia (que se busca por titulo) al sprint del
	// proyecto. Lo que no existe se manda con un ID que no esta en la base, asi el
	// rechazo lo produce la aplicacion.
	AsignarHistoriaASprint(proyecto, titulo, sprint string) error
	// CompletarHistoria y DesmarcarHistoria marcan y desmarcan la historia en el
	// ultimo sprint de su proyecto, el unico que puede estar abierto (US-010).
	CompletarHistoria(titulo string) error
	DesmarcarHistoria(titulo string) error
	// CerrarSprint cierra el sprint del proyecto (US-011): "Sprint 1" existe en todos
	// los proyectos, porque la numeracion es por proyecto.
	CerrarSprint(proyecto, nombre string) error
	SprintDeHistoria(titulo string) (string, error)
	SprintEstaCerrado(proyecto, nombre string) (bool, error)
	// SprintsAnteriores devuelve el historial de sprints cerrados del proyecto, del mas
	// reciente al mas antiguo, con su resultado (US-012).
	SprintsAnteriores(proyecto string) (SprintsAnteriores, error)
}

// SprintAnterior es un sprint cerrado con su resultado, como lo muestra el historial (US-012).
type SprintAnterior struct {
	Nombre                 string
	Planificados           int
	Completados            int
	HistoriasCompletadas   int
	HistoriasNoCompletadas int
}

// SprintsAnteriores es el historial de sprints cerrados y la velocidad del equipo.
type SprintsAnteriores struct {
	Sprints   []SprintAnterior
	Velocidad float64
}

// HorasDelSprint son las horas de un sprint tal como las muestra la consulta de
// US-027: estimadas, reales, la desviacion en horas y, si se puede calcular, en
// porcentaje.
type HorasDelSprint struct {
	Estimadas            float64
	Reales               float64
	Desviacion           float64
	Porcentaje           float64
	PorcentajeCalculable bool
}

// Metricas es el area de US-025, US-026, US-027 y US-028.
type Metricas interface {
	// AgregarItemAlSprint arma el contexto: suma una historia al sprint, con sus
	// story points y si esta completada. No valida nada: las reglas se aplican al
	// calcular, igual que en el dominio.
	AgregarItemAlSprint(sprint string, puntos int, completado bool) error

	// RegistrarSprintCerrado arma el contexto de la velocidad: suma un sprint ya
	// cerrado con sus story points planificados y completados.
	RegistrarSprintCerrado(nombre string, planificados, completados int) error
	PuntosPlanificados(sprint string) (int, error)
	PuntosCompletados(sprint string) (int, error)
	VelocidadDelEquipo() (float64, error)

	// RegistrarHorasDelSprint arma el contexto de US-027: suma al sprint un registro
	// de esfuerzo con esas horas. No valida nada: 0 o un numero negativo es justo el
	// dato que tiene que rechazar el dominio. Los story points del sprint salen de
	// AgregarItemAlSprint.
	RegistrarHorasDelSprint(sprint string, horas float64) error
	HorasDelSprint(sprint string) (HorasDelSprint, error)

	// PorcentajeHistoriasCompletadas es el de US-028: el porcentaje de historias
	// completadas del sprint, sobre los mismos items que arma AgregarItemAlSprint.
	PorcentajeHistoriasCompletadas(sprint string) (float64, error)
}

// Esfuerzo es el area de US-019 (US-020 y US-021 la amplian).
type Esfuerzo interface {
	// RegistrarEsfuerzo registra las horas que un integrante dedico a una
	// historia, todos nombrados como los conoce el escenario. Lo que no existe
	// (un integrante o una historia que el escenario no creo en el proyecto) se
	// manda con un ID que no esta en la base, asi el rechazo lo produce la
	// aplicacion.
	RegistrarEsfuerzo(proyecto, integrante, historia string, fecha time.Time, actividad string, horas float64) error
	CantidadRegistros(proyecto string) (int, error)
	HorasDeHistoria(proyecto, historia string) (float64, error)
}

// PlanningPoker es el area de US-014 a US-018. La sesion se nombra por su historia,
// que tiene como maximo una sesion abierta (RN-014-7).
type PlanningPoker interface {
	// IniciarSesion inicia una sesion sobre la historia del proyecto. Lo que no existe
	// (el proyecto o la historia) se manda con un ID que no esta en la base, asi el
	// rechazo lo produce la aplicacion.
	IniciarSesion(proyecto, titulo string) error
	EstadoDeLaSesion(titulo string) (estado string, ronda int, err error)
	// ParticipantesDeLaSesion devuelve los nombres de los participantes, en orden de
	// alta en el proyecto.
	ParticipantesDeLaSesion(titulo string) ([]string, error)
	// Votar registra el voto del integrante, nombrado como lo conoce el escenario, en
	// la sesion abierta de la historia (US-015). Lo que no existe se manda con un ID que
	// no esta en la base, asi el rechazo lo produce la aplicacion.
	Votar(integrante, titulo string, valor int) error
	// QuienesVotaron y QuienesFaltan devuelven nombres, en el orden de los participantes.
	QuienesVotaron(titulo string) ([]string, error)
	QuienesFaltan(titulo string) ([]string, error)
	// VotosOcultos dice si los valores de la ronda actual no se pueden leer (RN-015-7).
	VotosOcultos(titulo string) (bool, error)
	// Revelar revela los votos de la sesion abierta de la historia (US-016). Lo que no
	// existe se manda con un ID que no esta en la base, asi el rechazo lo produce la
	// aplicacion.
	Revelar(integrante, titulo string) error
	// VotosRevelados devuelve los votos de la ronda actual con los nombres de quienes
	// votaron, en el orden de los participantes. Solo se pueden leer con la sesion revelada.
	VotosRevelados(titulo string) ([]VotoRevelado, error)
	// HayDivergencia dice si la ultima revelacion de la sesion detecto divergencia (RN-016-6).
	HayDivergencia(titulo string) (bool, error)
	// RegistrarEstimacionAcordada registra el valor acordado en la ultima sesion de la
	// historia (US-018). Lo que no existe se manda con un ID que no esta en la base.
	RegistrarEstimacionAcordada(integrante, titulo string, valor int) error
	// EstimacionAcordada devuelve el valor acordado de la ultima sesion de la historia.
	EstimacionAcordada(titulo string) (int, error)
}

// VotoRevelado es el voto de un integrante, ya revelado.
type VotoRevelado struct {
	Integrante string
	Valor      int
}

// Servicios agrupa las areas conectadas para UN escenario. Un campo nil
// significa "area todavia sin conectar": los steps que la necesiten fallan con
// ErrAreaSinConectar.
type Servicios struct {
	Proyectos Proyectos
	Backlog   Backlog
	Sprints   Sprints
	Metricas  Metricas
	Esfuerzo  Esfuerzo
	// PlanningPoker es de US-014 a US-018.
	PlanningPoker PlanningPoker
}

// Fabrica construye los Servicios de un escenario. Se llama antes de CADA
// escenario y tiene que devolver un estado limpio: asi ningun escenario depende
// del anterior. La define features/servicios_test.go, que es el unico lugar que
// conoce las implementaciones concretas.
type Fabrica func() (Servicios, error)

// fechasPorDefecto es el rango que usa `Dado un proyecto "X"` cuando el
// escenario no le interesan las fechas.
var (
	inicioPorDefecto = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	finPorDefecto    = time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
)

// mundo es el estado compartido entre los steps de UN escenario.
// Se crea de cero antes de cada escenario, asi que no hay estado que se filtre
// de uno a otro.
type mundo struct {
	servicios Servicios

	// proyectoActual es el ultimo proyecto nombrado en un Dado. Los steps que
	// no nombran proyecto operan sobre este, para no repetirlo en cada linea.
	proyectoActual string

	// ultimoError es el error que dejo el ultimo step de accion (Cuando).
	// Lo leen `la operacion es exitosa` y `la operacion se rechaza`.
	ultimoError error

	// finDelUltimoSprint es el fin del ultimo sprint creado en cada proyecto del
	// escenario. Las frases de sprint sin fechas lo usan para armar las del siguiente.
	finDelUltimoSprint map[string]time.Time

	// resultados de las consultas de metricas
	puntosPlanificados  int
	puntosCompletados   int
	velocidad           float64
	horas               HorasDelSprint
	porcentajeHistorias float64
	anteriores          SprintsAnteriores
}

// registrarAccion guarda el resultado de un step de accion.
//
// Los steps de accion NO devuelven el error del dominio: lo guardan. Asi un
// escenario de error puede seguir corriendo hasta el `Entonces la operacion se
// rechaza`, en vez de cortarse en el `Cuando`.
func (m *mundo) registrarAccion(err error) error {
	m.ultimoError = err
	return nil
}

// proyectos, backlog, sprints y metricas devuelven el area o ErrAreaSinConectar.
// Protegen a los steps de correr contra un dominio que todavia no existe.
func (m *mundo) proyectos() (Proyectos, error) {
	if m.servicios.Proyectos == nil {
		return nil, areaSinConectar("proyectos")
	}
	return m.servicios.Proyectos, nil
}

func (m *mundo) backlog() (Backlog, error) {
	if m.servicios.Backlog == nil {
		return nil, areaSinConectar("backlog")
	}
	return m.servicios.Backlog, nil
}

func (m *mundo) sprints() (Sprints, error) {
	if m.servicios.Sprints == nil {
		return nil, areaSinConectar("sprints")
	}
	return m.servicios.Sprints, nil
}

func (m *mundo) esfuerzo() (Esfuerzo, error) {
	if m.servicios.Esfuerzo == nil {
		return nil, areaSinConectar("esfuerzo")
	}
	return m.servicios.Esfuerzo, nil
}

func (m *mundo) planningPoker() (PlanningPoker, error) {
	if m.servicios.PlanningPoker == nil {
		return nil, areaSinConectar("planning poker")
	}
	return m.servicios.PlanningPoker, nil
}

func (m *mundo) metricas() (Metricas, error) {
	if m.servicios.Metricas == nil {
		return nil, areaSinConectar("metricas")
	}
	return m.servicios.Metricas, nil
}

// parsearFecha convierte una fecha de un escenario en time.Time.
// El formato es siempre AAAA-MM-DD, por el diccionario de steps.
func parsearFecha(valor string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", valor)
	if err != nil {
		return time.Time{}, fmt.Errorf("la fecha %q no tiene el formato AAAA-MM-DD: %w", valor, err)
	}
	return t, nil
}
