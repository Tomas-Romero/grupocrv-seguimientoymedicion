package features_test

import "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"

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
	return steps.Servicios{}, nil
}
