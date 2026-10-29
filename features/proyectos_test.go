package features_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
)

// proyectosSobreApp conecta el area Proyectos con los casos de uso de US-001,
// US-002 y US-003 sobre la base en memoria del escenario (ADR 0002). Los
// Entonces leen la base directamente, porque todavia no hay un caso de uso de
// lectura.
type proyectosSobreApp struct {
	escenario *escenario
	crear     *app.CrearProyecto
	modificar *app.ModificarProyecto
	registrar *app.RegistrarIntegrante
	estado    *app.ConsultarEstadoDelProyecto
}

var _ steps.Proyectos = (*proyectosSobreApp)(nil)

func nuevosProyectosSobreApp(e *escenario) *proyectosSobreApp {
	return &proyectosSobreApp{
		escenario: e,
		crear:     app.NuevoCrearProyecto(e.base),
		modificar: app.NuevoModificarProyecto(e.base),
		registrar: app.NuevoRegistrarIntegrante(e.base),
		estado:    app.NuevoConsultarEstadoDelProyecto(e.base, ahoraEnEscenarios),
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

// RegistrarIntegrante es de US-003. No valida nada: lo hacen el dominio y el caso
// de uso. El rol se convierte a integrante.Rol sin normalizarlo (RN-003-5), asi
// que un "PRODUCT_BUILDER" llega tal cual y el dominio lo rechaza. Si el escenario
// nombra un proyecto que no creo, se llama con un ID que ningun proyecto tiene y
// el rechazo (ErrProyectoInexistente, RN-003-6) lo produce la aplicacion.
func (a *proyectosSobreApp) RegistrarIntegrante(proyecto, nombre, email, rol string) error {
	_, err := a.registrar.Ejecutar(context.Background(), a.idDe(proyecto), nombre, email, integrante.Rol(rol))
	return err
}

// CantidadIntegrantes es de US-003. Lee la base directamente, igual que
// CantidadHistoriasEnBacklog: todavia no hay un caso de uso de lectura.
func (a *proyectosSobreApp) CantidadIntegrantes(proyecto string) (int, error) {
	id, ok := a.escenario.proyectos[proyecto]
	if !ok {
		return 0, fmt.Errorf("el proyecto %q no se creo en este escenario", proyecto)
	}
	return len(a.escenario.base.Integrantes(id)), nil
}

// ConsultarEstadoProyecto es de US-004: el caso de uso real sobre la base en memoria, con el reloj fijo de los escenarios.
// Si el escenario no creo el proyecto, se manda un ID que ningun proyecto tiene y el rechazo lo produce la aplicacion.
func (a *proyectosSobreApp) ConsultarEstadoProyecto(nombre string) (steps.EstadoDelProyecto, error) {
	e, err := a.estado.Ejecutar(context.Background(), a.idDe(nombre))
	if err != nil {
		return steps.EstadoDelProyecto{}, err
	}
	resultado := steps.EstadoDelProyecto{
		Estado:      string(e.Estado),
		Dias:        e.Dias,
		Pendientes:  e.Historias.Pendientes,
		EnProgreso:  e.Historias.EnProgreso,
		Completadas: e.Historias.Completadas,
		StoryPoints: e.StoryPoints,
		Completados: e.Completados,
		Porcentaje:  e.Porcentaje,
	}
	if e.SprintAbierto != nil {
		resultado.SprintAbierto = e.SprintAbierto.Nombre()
	}
	return resultado, nil
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
