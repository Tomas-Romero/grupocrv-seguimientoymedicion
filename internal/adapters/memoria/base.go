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

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
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
}

// NuevaBase devuelve una base vacia.
func NuevaBase() *Base {
	return &Base{
		proyectos:      map[string]bool{},
		items:          map[string][]backlog.ItemBacklog{},
		datosProyectos: map[string]proyecto.Proyecto{},
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
