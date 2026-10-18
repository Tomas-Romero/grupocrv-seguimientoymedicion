package features_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// sprintsSobreApp conecta el area Sprints con el caso de uso de US-008 sobre la base
// en memoria del escenario (ADR 0002). Los Entonces leen la base directamente,
// porque todavia no hay casos de uso de lectura (US-012). Las demas operaciones del
// area son de US-009, US-010 y US-011.
type sprintsSobreApp struct {
	escenario *escenario
	crear     *app.CrearSprint
}

var _ steps.Sprints = (*sprintsSobreApp)(nil)

func nuevoSprintsSobreApp(e *escenario) *sprintsSobreApp {
	return &sprintsSobreApp{escenario: e, crear: app.NuevoCrearSprint(e.base)}
}

// CrearSprint no valida nada: lo hace el caso de uso. Si el escenario nombra un
// proyecto que no creo (por ejemplo, despues de `Dado ningún proyecto`), llama con un
// ID que no esta en la base, asi el rechazo lo produce la aplicacion (RN-008-9).
// Devuelve el nombre que asigno la aplicacion.
func (a *sprintsSobreApp) CrearSprint(proyecto, objetivo string, inicio, fin time.Time) (string, error) {
	proyectoID, ok := a.escenario.proyectos[proyecto]
	if !ok {
		proyectoID = idDeProyectoInexistente
	}
	creado, err := a.crear.Ejecutar(context.Background(), proyectoID, objetivo, inicio, fin)
	if err != nil {
		return "", err
	}
	return creado.Nombre(), nil
}

// SprintEstaCerrado busca el sprint por nombre dentro del proyecto: "Sprint 1" existe
// en todos los proyectos, porque la numeracion es por proyecto.
func (a *sprintsSobreApp) SprintEstaCerrado(proyecto, nombre string) (bool, error) {
	proyectoID, ok := a.escenario.proyectos[proyecto]
	if !ok {
		return false, fmt.Errorf("el proyecto %q no se creo en este escenario", proyecto)
	}
	for _, s := range a.escenario.base.Sprints(proyectoID) {
		if s.Nombre() == nombre {
			return s.Estado == sprint.EstadoCerrado, nil
		}
	}
	return false, fmt.Errorf("el sprint %q no existe en el proyecto %q", nombre, proyecto)
}

// AsignarHistoriaASprint es de US-009.
func (a *sprintsSobreApp) AsignarHistoriaASprint(_, _ string) error { return errSinConectar }

// CompletarHistoria es de US-010.
func (a *sprintsSobreApp) CompletarHistoria(_ string) error { return errSinConectar }

// CerrarSprint es de US-011.
func (a *sprintsSobreApp) CerrarSprint(_ string) error { return errSinConectar }

// SprintDeHistoria depende de la asignacion (US-009).
func (a *sprintsSobreApp) SprintDeHistoria(_ string) (string, error) { return "", errSinConectar }
