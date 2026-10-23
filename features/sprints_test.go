package features_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// idDeSprintInexistente es el ID con el que se llama a AsignarItemASprint cuando el
// escenario nombra un sprint que no existe en el proyecto. Igual que con el proyecto y
// la historia, el rechazo (ErrSprintInexistente) lo produce la aplicacion.
const idDeSprintInexistente = "00000000-0000-0000-0000-000000000000"

// sprintsSobreApp conecta el area Sprints con los casos de uso de US-008, US-009 y
// US-010 sobre la base en memoria del escenario (ADR 0002). Los Entonces leen la base
// directamente, porque todavia no hay casos de uso de lectura (US-012). Cerrar un
// sprint es de US-011.
type sprintsSobreApp struct {
	escenario  *escenario
	crear      *app.CrearSprint
	asignar    *app.AsignarItemASprint
	marcar     *app.MarcarCompletado
	desmarcar  *app.DesmarcarCompletado
	cerrar     *app.CerrarSprint
	anteriores *app.ConsultarSprintsAnteriores
}

var _ steps.Sprints = (*sprintsSobreApp)(nil)

func nuevoSprintsSobreApp(e *escenario) *sprintsSobreApp {
	return &sprintsSobreApp{
		escenario:  e,
		crear:      app.NuevoCrearSprint(e.base),
		asignar:    app.NuevoAsignarItemASprint(e.base),
		marcar:     app.NuevoMarcarCompletado(e.base, ahoraEnEscenarios),
		desmarcar:  app.NuevoDesmarcarCompletado(e.base),
		cerrar:     app.NuevoCerrarSprint(e.base),
		anteriores: app.NuevoConsultarSprintsAnteriores(e.base),
	}
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

// AsignarHistoriaASprint no valida nada: lo hace el caso de uso. Busca la historia por
// titulo en todos los proyectos del escenario (una historia de otro proyecto tiene que
// llegar a la aplicacion para que ella la rechace) y el sprint por nombre dentro del
// proyecto actual. Lo que no encuentra lo manda con un ID que no existe.
func (a *sprintsSobreApp) AsignarHistoriaASprint(proyecto, titulo, nombreDelSprint string) error {
	sprintID := idDeSprintInexistente
	if proyectoID, ok := a.escenario.proyectos[proyecto]; ok {
		for _, s := range a.escenario.base.Sprints(proyectoID) {
			if s.Nombre() == nombreDelSprint {
				sprintID = s.ID
			}
		}
	}
	_, err := a.asignar.Ejecutar(context.Background(), a.idDeHistoria(titulo), sprintID)
	return err
}

// SprintDeHistoria devuelve el nombre del sprint abierto en el que esta la historia.
func (a *sprintsSobreApp) SprintDeHistoria(titulo string) (string, error) {
	itemID := a.idDeHistoria(titulo)
	actual, ok := a.escenario.base.SprintActualDelItem(itemID)
	if !ok {
		return "", fmt.Errorf("la historia %q no esta en ningun sprint", titulo)
	}
	return actual.Nombre(), nil
}

// idDeHistoria busca la historia por titulo en todos los proyectos del escenario: los
// escenarios no repiten titulos (docs/diccionario-steps.md). Si no esta, devuelve un ID
// que no existe.
func (a *sprintsSobreApp) idDeHistoria(titulo string) string {
	for _, proyectoID := range a.escenario.proyectos {
		for _, item := range a.escenario.base.Items(proyectoID) {
			if item.Titulo == titulo {
				return item.ID
			}
		}
	}
	return idDeItemInexistente
}

// CompletarHistoria marca la historia como completada con el caso de uso de US-010, en
// el ultimo sprint de su proyecto. No valida nada: lo hace el caso de uso.
func (a *sprintsSobreApp) CompletarHistoria(titulo string) error {
	itemID, sprintID := a.historiaYUltimoSprint(titulo)
	_, err := a.marcar.Ejecutar(context.Background(), itemID, sprintID)
	return err
}

// DesmarcarHistoria desmarca la historia con el caso de uso de US-010, en el ultimo
// sprint de su proyecto.
func (a *sprintsSobreApp) DesmarcarHistoria(titulo string) error {
	itemID, sprintID := a.historiaYUltimoSprint(titulo)
	_, err := a.desmarcar.Ejecutar(context.Background(), itemID, sprintID)
	return err
}

// historiaYUltimoSprint busca la historia por titulo y el ultimo sprint de su proyecto,
// que es el unico que puede estar abierto (RN-008-17). Una historia que el escenario no
// creo, o un proyecto sin sprints, se mandan con un ID que no existe: el rechazo lo
// produce la aplicacion (RN-010-1).
func (a *sprintsSobreApp) historiaYUltimoSprint(titulo string) (itemID, sprintID string) {
	itemID, sprintID = idDeItemInexistente, idDeSprintInexistente
	for _, proyectoID := range a.escenario.proyectos {
		for _, item := range a.escenario.base.Items(proyectoID) {
			if item.Titulo != titulo {
				continue
			}
			itemID = item.ID
			if sprints := a.escenario.base.Sprints(proyectoID); len(sprints) > 0 {
				sprintID = sprints[len(sprints)-1].ID
			}
			return itemID, sprintID
		}
	}
	return itemID, sprintID
}

// CerrarSprint cierra el sprint del proyecto con el caso de uso de US-011. No valida
// nada: lo hace el caso de uso. Un sprint que no esta en el proyecto se manda con un ID
// que no existe, asi el rechazo lo produce la aplicacion (RN-011-1).
func (a *sprintsSobreApp) CerrarSprint(proyecto, nombre string) error {
	sprintID := idDeSprintInexistente
	if proyectoID, ok := a.escenario.proyectos[proyecto]; ok {
		for _, s := range a.escenario.base.Sprints(proyectoID) {
			if s.Nombre() == nombre {
				sprintID = s.ID
			}
		}
	}
	_, err := a.cerrar.Ejecutar(context.Background(), sprintID)
	return err
}

// SprintsAnteriores consulta el historial con el caso de uso de US-012. Un proyecto que el
// escenario no creo (despues de `Dado ningún proyecto`) se manda con un ID que no esta en la base,
// asi el rechazo lo produce la aplicacion.
func (a *sprintsSobreApp) SprintsAnteriores(proyecto string) (steps.SprintsAnteriores, error) {
	proyectoID, ok := a.escenario.proyectos[proyecto]
	if !ok {
		proyectoID = idDeProyectoInexistente
	}
	resultado, err := a.anteriores.Ejecutar(context.Background(), proyectoID)
	if err != nil {
		return steps.SprintsAnteriores{}, err
	}
	sprints := make([]steps.SprintAnterior, 0, len(resultado.Sprints))
	for _, s := range resultado.Sprints {
		sprints = append(sprints, steps.SprintAnterior{
			Nombre: s.Sprint.Nombre(), Planificados: s.Planificados, Completados: s.Completados,
			HistoriasCompletadas: s.HistoriasCompletadas, HistoriasNoCompletadas: s.HistoriasNoCompletadas,
		})
	}
	return steps.SprintsAnteriores{Sprints: sprints, Velocidad: resultado.Velocidad}, nil
}
