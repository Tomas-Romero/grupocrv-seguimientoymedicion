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

// proyectosSobreApp conecta el area Proyectos con los casos de uso de US-001 y
// US-002 sobre la base en memoria del escenario (ADR 0002). Los Entonces leen la
// base directamente, porque todavia no hay un caso de uso de lectura.
type proyectosSobreApp struct {
	escenario *escenario
	crear     *app.CrearProyecto
	modificar *app.ModificarProyecto
}

var _ steps.Proyectos = (*proyectosSobreApp)(nil)

func nuevosProyectosSobreApp(e *escenario) *proyectosSobreApp {
	return &proyectosSobreApp{
		escenario: e,
		crear:     app.NuevoCrearProyecto(e.base),
		modificar: app.NuevoModificarProyecto(e.base),
	}
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

// ModificarNombreProyecto es de US-002. El caso de uso recibe siempre los cuatro
// datos (RN-002-3), asi que el adaptador completa los que el escenario no
// nombra con los que ya tiene guardados el proyecto. Si el cambio sale bien, el
// escenario pasa a conocer al proyecto por su nombre nuevo, ya recortado.
func (a *proyectosSobreApp) ModificarNombreProyecto(actual, nuevo string) error {
	id := a.idDe(actual)
	// Si el proyecto no existe, guardado queda vacio: no importa, porque el caso
	// de uso responde ErrProyectoInexistente antes de mirar los datos (RN-002-1).
	guardado, _ := a.escenario.base.Proyecto(id)

	p, err := a.modificar.Ejecutar(context.Background(), id, nuevo, guardado.Descripcion, guardado.FechaInicio, guardado.FechaFin)
	if err != nil {
		return err
	}
	delete(a.escenario.proyectos, actual)
	a.escenario.proyectos[p.Nombre] = p.ID
	return nil
}

// ModificarFechasProyecto es de US-002. Reenvia el nombre y la descripcion que
// ya tiene el proyecto, igual que ModificarNombreProyecto con las fechas.
func (a *proyectosSobreApp) ModificarFechasProyecto(nombre string, inicio, fin time.Time) error {
	id := a.idDe(nombre)
	guardado, _ := a.escenario.base.Proyecto(id)

	_, err := a.modificar.Ejecutar(context.Background(), id, guardado.Nombre, guardado.Descripcion, inicio, fin)
	return err
}

// RegistrarIntegrante es de US-003.
func (a *proyectosSobreApp) RegistrarIntegrante(_, _, _ string) error {
	return errNoImplementado
}

// CantidadIntegrantes es de US-003.
func (a *proyectosSobreApp) CantidadIntegrantes(_ string) (int, error) {
	return 0, errNoImplementado
}

// idDe devuelve el ID del proyecto que el escenario creo con ese nombre. Si no lo
// creo, devuelve un ID que ningun proyecto tiene, para que el rechazo lo
// produzca la aplicacion y no este adaptador (como en backlog_test.go).
func (a *proyectosSobreApp) idDe(nombre string) string {
	if id, ok := a.escenario.proyectos[nombre]; ok {
		return id
	}
	return idDeProyectoInexistente
}
