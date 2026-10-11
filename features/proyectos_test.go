package features_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
)

// errNoImplementado lo devuelven las operaciones que todavia no tienen caso de
// uso (o calculo) conectado: si algun escenario las necesita, falla en voz alta.
var errNoImplementado = errors.New("el doble en memoria no implementa esta operacion")

// proyectosSobreApp conecta el area Proyectos con el caso de uso de US-001 sobre
// la base en memoria del escenario (ADR 0002). Los Entonces leen la base
// directamente, porque todavia no hay un caso de uso de lectura.
type proyectosSobreApp struct {
	escenario *escenario
	crear     *app.CrearProyecto
}

var _ steps.Proyectos = (*proyectosSobreApp)(nil)

func nuevosProyectosSobreApp(e *escenario) *proyectosSobreApp {
	return &proyectosSobreApp{escenario: e, crear: app.NuevoCrearProyecto(e.base)}
}

// CrearProyecto no valida nada: lo hace el dominio dentro del caso de uso. Si el
// alta sale bien, anota el ID por nombre en el escenario, que es donde lo busca
// el area Backlog (US-005). Se anota con el nombre ya recortado.
func (a *proyectosSobreApp) CrearProyecto(nombre string, inicio, fin time.Time) error {
	p, err := a.crear.Ejecutar(context.Background(), nombre, "", inicio, fin)
	if err != nil {
		return err
	}
	a.escenario.proyectos[p.Nombre] = p.ID
	return nil
}

func (a *proyectosSobreApp) ProyectoExiste(nombre string) (time.Time, time.Time, error) {
	id, ok := a.escenario.proyectos[nombre]
	if !ok {
		return time.Time{}, time.Time{}, fmt.Errorf("el proyecto %q no se creo en este escenario", nombre)
	}
	p, ok := a.escenario.base.Proyecto(id)
	if !ok {
		return time.Time{}, time.Time{}, fmt.Errorf("el proyecto %q no esta en la base", nombre)
	}
	return p.FechaInicio, p.FechaFin, nil
}

// ModificarNombreProyecto es de US-002.
func (a *proyectosSobreApp) ModificarNombreProyecto(_, _ string) error {
	return errNoImplementado
}

// RegistrarIntegrante es de US-003.
func (a *proyectosSobreApp) RegistrarIntegrante(_, _, _ string) error {
	return errNoImplementado
}

// CantidadIntegrantes es de US-003.
func (a *proyectosSobreApp) CantidadIntegrantes(_ string) (int, error) {
	return 0, errNoImplementado
}
