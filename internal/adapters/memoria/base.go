// Package memoria guarda los datos de la aplicacion en memoria. Implementa los
// mismos puertos de internal/app que el adaptador de Postgres, y lo usan los
// escenarios BDD, que se prueban contra la capa de aplicacion y no contra la
// base (docs/adr/0002-integracion-bdd-godog.md).
//
// No tiene reglas de negocio propias: replica lo que hace la persistencia (IDs
// generados, numeracion correlativa, proyecto inexistente) y nada mas.
package memoria

import (
	"fmt"
	"sync"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Base es una base de datos en memoria. Se puede usar desde varias goroutines.
type Base struct {
	// mu serializa las altas, como el candado sobre el proyecto en Postgres.
	mu sync.Mutex

	ultimoID  int
	proyectos map[string]bool
	items     map[string][]backlog.ItemBacklog // por proyecto, en orden de alta

	// datosProyectos guarda los datos de los proyectos registrados con
	// RegistrarProyecto (US-001). Todo proyecto de esta tabla esta tambien en
	// proyectos, que es el mapa donde busca ExisteProyecto.
	datosProyectos map[string]proyecto.Proyecto

	// integrantes guarda los integrantes de cada proyecto, en orden de alta
	// (US-003).
	integrantes map[string][]integrante.Integrante

	// sprints guarda los sprints de cada proyecto, en orden de alta: el ultimo de la
	// lista es el de numero mas alto (US-008).
	sprints map[string][]sprint.Sprint

	// esfuerzos guarda los registros de esfuerzo de cada proyecto, en orden de alta
	// (US-019).
	esfuerzos map[string][]esfuerzo.RegistroEsfuerzo

	// defectos guarda los defectos de cada proyecto, en orden de alta (US-022).
	defectos map[string][]defecto.Defecto
	// defectosCreadoEn guarda el instante de alta de cada defecto, por ID, como la columna
	// creado_en (T-021): el defecto no lo lleva, pero listarlos del mas reciente al mas antiguo
	// lo necesita. ultimaAltaDeDefecto es el ultimo que se dio, para que cada alta sea posterior
	// a la anterior.
	defectosCreadoEn    map[string]time.Time
	ultimaAltaDeDefecto time.Time

	// asignaciones guarda las filas de la relacion entre cada item y sus sprints, en
	// orden de alta (US-009). Un item tiene como mucho una pendiente.
	asignaciones map[string][]asignacion
	// ultimaAsignacion es el numero de la ultima fila de la relacion dada de alta.
	ultimaAsignacion int

	// sesiones guarda las sesiones de Planning Poker, en orden de alta (US-014).
	sesiones []planningpoker.SesionPlanningPoker

	// votos guarda el valor de cada voto, por sesion, ronda e integrante (US-015).
	votos map[claveDeVoto]int
}

// NuevaBase devuelve una base vacia.
func NuevaBase() *Base {
	return &Base{
		proyectos:        map[string]bool{},
		items:            map[string][]backlog.ItemBacklog{},
		datosProyectos:   map[string]proyecto.Proyecto{},
		integrantes:      map[string][]integrante.Integrante{},
		sprints:          map[string][]sprint.Sprint{},
		esfuerzos:        map[string][]esfuerzo.RegistroEsfuerzo{},
		defectos:         map[string][]defecto.Defecto{},
		defectosCreadoEn: map[string]time.Time{},
		votos:            map[claveDeVoto]int{},
		asignaciones:     map[string][]asignacion{},
	}
}

// AgregarProyecto registra un proyecto y devuelve su ID. Hasta que US-001 tenga
// su caso de uso, es la forma de tener proyectos en los escenarios.
func (b *Base) AgregarProyecto() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nuevoID()
	b.proyectos[id] = true
	return id
}

// nuevoID imita a gen_random_uuid(): devuelve un ID con formato de UUID que no
// se repite en la base. Es correlativo para que los tests sean deterministas.
// Se llama con mu tomado.
func (b *Base) nuevoID() string {
	b.ultimoID++
	return fmt.Sprintf("00000000-0000-0000-0000-%012d", b.ultimoID)
}
