package features_test

import (
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
)

// escenario es el estado que comparten las areas durante UN escenario: la base
// en memoria y el ID de cada proyecto por nombre. Los steps nombran a los
// proyectos; la aplicacion los identifica por ID.
type escenario struct {
	base      *memoria.Base
	proyectos map[string]string
}

// ahoraEnEscenarios es el reloj de los casos de uso en los escenarios. Es fijo
// para que ningun escenario dependa del dia en que corre.
func ahoraEnEscenarios() time.Time {
	return time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
}

// nuevosServicios arma los servicios que usan los steps. Se llama antes de
// CADA escenario y tiene que devolver un estado limpio.
//
// Cada area se conecta aca cuando la historia que la implementa se mergea:
//
//	Proyectos: US-001, US-002, US-003
//	Backlog:   US-005, US-006, US-013
//	Sprints:   US-008, US-009, US-010, US-011
//	Metricas:  US-025, US-026
//
// Un area sin conectar hace fallar con ErrAreaSinConectar a cualquier escenario
// que la use: preferimos rojo explicito a verde falso.
func nuevosServicios() (steps.Servicios, error) {
	e := &escenario{base: memoria.NuevaBase(), proyectos: map[string]string{}}
	return steps.Servicios{
		// TEMPORAL (T-006): doble en memoria del area Proyectos. Hasta US-001
		// tambien registra los proyectos en la base del escenario, que es donde
		// los busca el caso de uso de US-005. US-001 lo reemplaza por el
		// adaptador real sobre internal/app.
		Proyectos: nuevosProyectosEnMemoria(e),
		// US-005: el caso de uso real sobre la base en memoria.
		Backlog: nuevoBacklogSobreApp(e),
	}, nil
}
