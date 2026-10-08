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
	RegistrarIntegrante(proyecto, nombre, rol string) error
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
	CrearSprint(proyecto, nombre, objetivo string) error
	AsignarHistoriaASprint(titulo, sprint string) error
	CompletarHistoria(titulo string) error
	CerrarSprint(nombre string) error
	SprintDeHistoria(titulo string) (string, error)
	SprintEstaCerrado(nombre string) (bool, error)
}

// Metricas es el area de US-025 y US-026.
type Metricas interface {
	// AgregarItemAlSprint arma el contexto: suma una historia al sprint, con sus
	// story points y si esta completada. No valida nada: las reglas se aplican al
	// calcular, igual que en el dominio.
	AgregarItemAlSprint(sprint string, puntos int, completado bool) error
	PuntosPlanificados(sprint string) (int, error)
	PuntosCompletados(sprint string) (int, error)
	VelocidadDelEquipo() (float64, error)
}

// Servicios agrupa las areas conectadas para UN escenario. Un campo nil
// significa "area todavia sin conectar": los steps que la necesiten fallan con
// ErrAreaSinConectar.
type Servicios struct {
	Proyectos Proyectos
	Backlog   Backlog
	Sprints   Sprints
	Metricas  Metricas
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

	// resultados de las consultas de metricas
	puntosPlanificados int
	puntosCompletados  int
	velocidad          float64
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
