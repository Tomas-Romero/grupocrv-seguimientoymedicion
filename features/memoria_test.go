package features_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
)

// errNoImplementado lo devuelve el doble en las operaciones que el escenario de
// cableado no usa: si algun escenario las necesita, falla en voz alta.
var errNoImplementado = errors.New("el doble en memoria no implementa esta operacion")

// proyectosEnMemoria es un doble del area Proyectos. Nacio para que el escenario
// de cableado (T-006-cableado-bdd.feature) pruebe el runner, los steps y los
// servicios antes de que exista la primera historia. Hasta que llegue US-001,
// tambien les da proyectos a los escenarios de US-005: cada proyecto que crea
// queda registrado en la base en memoria del escenario.
//
// No tiene reglas de negocio a proposito: acepta todo, para no inventar un
// dominio distinto del real. Se borra, junto con el escenario de cableado,
// cuando US-001 conecte el area con internal/app.
type proyectosEnMemoria struct {
	escenario *escenario
	proyectos map[string]fechas
}

type fechas struct{ inicio, fin time.Time }

var _ steps.Proyectos = (*proyectosEnMemoria)(nil)

func nuevosProyectosEnMemoria(e *escenario) *proyectosEnMemoria {
	return &proyectosEnMemoria{escenario: e, proyectos: map[string]fechas{}}
}

func (p *proyectosEnMemoria) CrearProyecto(nombre string, inicio, fin time.Time) error {
	p.proyectos[nombre] = fechas{inicio: inicio, fin: fin}
	p.escenario.proyectos[nombre] = p.escenario.base.AgregarProyecto()
	return nil
}

func (p *proyectosEnMemoria) ProyectoExiste(nombre string) (time.Time, time.Time, error) {
	f, ok := p.proyectos[nombre]
	if !ok {
		return time.Time{}, time.Time{}, fmt.Errorf("proyecto %q: no existe", nombre)
	}
	return f.inicio, f.fin, nil
}

func (p *proyectosEnMemoria) ModificarNombreProyecto(_, _ string) error {
	return errNoImplementado
}

func (p *proyectosEnMemoria) RegistrarIntegrante(_, _, _ string) error {
	return errNoImplementado
}

func (p *proyectosEnMemoria) CantidadIntegrantes(_ string) (int, error) {
	return 0, errNoImplementado
}
