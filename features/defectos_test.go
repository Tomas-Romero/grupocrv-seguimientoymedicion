package features_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// idDeDefectoInexistente es el ID con el que se llama al caso de uso de US-023 cuando el
// escenario nombra un defecto que no existe.
const idDeDefectoInexistente = "00000000-0000-0000-0000-000000000000"

// defectosSobreApp conecta el area Defectos con los casos de uso de US-022 y US-023 sobre la base
// en memoria del escenario (ADR 0002), y el de US-024 que los lista con filtros. Los Entonces de un
// defecto leen la base directamente; los de la lista leen lo que devolvio el caso de uso.
type defectosSobreApp struct {
	escenario *escenario
	registrar *app.RegistrarDefecto
	cambiar   *app.CambiarEstadoDefecto
	listar    *app.ListarDefectos

	// listados es la ultima lista que devolvio el caso de uso, nula si todavia no se listo o si el
	// ultimo listado se rechazo.
	listados []app.DefectoListado
	listado  bool
}

var _ steps.Defectos = (*defectosSobreApp)(nil)

func nuevoDefectosSobreApp(e *escenario) *defectosSobreApp {
	return &defectosSobreApp{
		escenario: e,
		registrar: app.NuevoRegistrarDefecto(e.base),
		cambiar:   app.NuevoCambiarEstadoDefecto(e.base),
		listar:    app.NuevoListarDefectos(e.base),
	}
}

// RegistrarDefecto no valida nada: lo hace el caso de uso. Busca la historia por titulo y el
// sprint por nombre DENTRO del proyecto; si el escenario nombra uno que no esta ahi (por ejemplo,
// de otro proyecto), llama con un ID que no existe, asi el rechazo lo produce la aplicacion.
func (a *defectosSobreApp) RegistrarDefecto(proyecto, historia, nombreDelSprint, severidad, descripcion string) error {
	proyectoID, ok := a.escenario.proyectos[proyecto]
	if !ok {
		proyectoID = idDeProyectoInexistente
	}
	itemID, sprintID := idDeItemInexistente, idDeSprintInexistente
	for _, item := range a.escenario.base.Items(proyectoID) {
		if item.Titulo == historia {
			itemID = item.ID
		}
	}
	for _, s := range a.escenario.base.Sprints(proyectoID) {
		if s.Nombre() == nombreDelSprint {
			sprintID = s.ID
		}
	}
	_, err := a.registrar.Ejecutar(context.Background(), proyectoID, itemID, sprintID, descripcion, defecto.Severidad(severidad))
	return err
}

func (a *defectosSobreApp) CantidadDefectos(proyecto string) (int, error) {
	proyectoID, ok := a.escenario.proyectos[proyecto]
	if !ok {
		return 0, fmt.Errorf("el proyecto %q no se creo en este escenario", proyecto)
	}
	return len(a.escenario.base.Defectos(proyectoID)), nil
}

func (a *defectosSobreApp) SeveridadDeDefecto(historia string) (string, error) {
	d, err := a.ultimoDefecto(historia)
	return string(d.Severidad), err
}

func (a *defectosSobreApp) EstadoDeDefecto(historia string) (string, error) {
	d, err := a.ultimoDefecto(historia)
	return string(d.Estado), err
}

func (a *defectosSobreApp) SprintDeDeteccionDeDefecto(historia string) (string, error) {
	d, err := a.ultimoDefecto(historia)
	if err != nil {
		return "", err
	}
	for _, proyectoID := range a.escenario.proyectos {
		for _, s := range a.escenario.base.Sprints(proyectoID) {
			if s.ID == d.SprintDeteccionID {
				return s.Nombre(), nil
			}
		}
	}
	return "", fmt.Errorf("el sprint de detección %s del defecto de %q no existe", d.SprintDeteccionID, historia)
}

func (a *defectosSobreApp) DefectoSinResolver(historia string) (bool, error) {
	d, err := a.ultimoDefecto(historia)
	return d.SprintResolucionID == "", err
}

// ultimoDefecto devuelve el ultimo defecto registrado de la historia, que se busca por titulo en
// todos los proyectos del escenario: los escenarios no repiten titulos.
func (a *defectosSobreApp) ultimoDefecto(historia string) (defecto.Defecto, error) {
	for _, proyectoID := range a.escenario.proyectos {
		for _, item := range a.escenario.base.Items(proyectoID) {
			if item.Titulo != historia {
				continue
			}
			defectos := a.escenario.base.Defectos(proyectoID)
			for i := len(defectos) - 1; i >= 0; i-- {
				if defectos[i].ItemID == item.ID {
					return defectos[i], nil
				}
			}
			return defecto.Defecto{}, fmt.Errorf("la historia %q no tiene defectos", historia)
		}
	}
	return defecto.Defecto{}, fmt.Errorf("la historia %q no existe en el escenario", historia)
}

// CambiarEstadoDeDefecto cambia el estado del ultimo defecto de la historia con el caso de uso real
// (US-023). El proyecto es el de la historia y el sprint de resolucion se busca por nombre DENTRO de
// ese proyecto. Lo que el escenario nombra y no existe (la historia, su defecto o el sprint) se manda
// con un ID que no existe, asi el rechazo lo produce la aplicacion y no los steps. Un defecto "de
// otro proyecto" no se puede armar con estas frases: lo prueban el caso de uso y el repositorio.
func (a *defectosSobreApp) CambiarEstadoDeDefecto(historia, estado, sprintResolucion string) error {
	proyectoID, defectoID := idDeProyectoInexistente, idDeDefectoInexistente
	if d, proyecto, ok := a.buscarDefecto(historia); ok {
		proyectoID, defectoID = proyecto, d.ID
	} else if proyecto != "" {
		proyectoID = proyecto
	}

	sprintID := ""
	if sprintResolucion != "" {
		sprintID = idDeSprintInexistente
		for _, s := range a.escenario.base.Sprints(proyectoID) {
			if s.Nombre() == sprintResolucion {
				sprintID = s.ID
			}
		}
	}
	_, err := a.cambiar.Ejecutar(context.Background(), proyectoID, defectoID, defecto.Estado(estado), sprintID)
	return err
}

func (a *defectosSobreApp) SprintDeResolucionDeDefecto(historia string) (string, bool, error) {
	d, err := a.ultimoDefecto(historia)
	if err != nil || d.SprintResolucionID == "" {
		return "", false, err
	}
	for _, s := range a.escenario.base.Sprints(d.ProyectoID) {
		if s.ID == d.SprintResolucionID {
			return s.Nombre(), true, nil
		}
	}
	return "", false, fmt.Errorf("el sprint de resolución %s del defecto de %q no existe", d.SprintResolucionID, historia)
}

// buscarDefecto busca el ultimo defecto de la historia y devuelve el proyecto donde esta la historia.
// Si la historia existe pero no tiene defectos devuelve false y su proyecto; si la historia no
// existe en ningun proyecto, false y un proyecto vacio.
func (a *defectosSobreApp) buscarDefecto(historia string) (defecto.Defecto, string, bool) {
	for _, proyectoID := range a.escenario.proyectos {
		for _, item := range a.escenario.base.Items(proyectoID) {
			if item.Titulo != historia {
				continue
			}
			defectos := a.escenario.base.Defectos(proyectoID)
			for i := len(defectos) - 1; i >= 0; i-- {
				if defectos[i].ItemID == item.ID {
					return defectos[i], proyectoID, true
				}
			}
			return defecto.Defecto{}, proyectoID, false
		}
	}
	return defecto.Defecto{}, "", false
}

// ListarDefectos lista los defectos del proyecto con el caso de uso real (US-024). El sprint se busca por
// nombre DENTRO del proyecto; si el escenario nombra uno que no esta ahi, llama con un ID que no existe, asi el
// rechazo lo produce la aplicacion y no los steps. Un sprint "de otro proyecto" no se puede armar con estas
// frases: lo prueban el caso de uso y la pantalla. La severidad va tal cual.
func (a *defectosSobreApp) ListarDefectos(proyecto, nombreDelSprint, severidad string) error {
	proyectoID, ok := a.escenario.proyectos[proyecto]
	if !ok {
		proyectoID = idDeProyectoInexistente
	}
	filtro := app.FiltroDefectos{Severidad: defecto.Severidad(severidad)}
	if nombreDelSprint != "" {
		filtro.SprintDeteccionID = idDeSprintInexistente
		for _, s := range a.escenario.base.Sprints(proyectoID) {
			if s.Nombre() == nombreDelSprint {
				filtro.SprintDeteccionID = s.ID
			}
		}
	}

	a.listados, a.listado = nil, false
	listados, err := a.listar.Ejecutar(context.Background(), proyectoID, filtro)
	if err != nil {
		return err
	}
	a.listados, a.listado = listados, true
	return nil
}

func (a *defectosSobreApp) HistoriasDeLosDefectosListados() ([]string, error) {
	if !a.listado {
		return nil, errors.New("no hay una lista de defectos: ningun listado se hizo o el ultimo se rechazo")
	}
	historias := make([]string, len(a.listados))
	for i, l := range a.listados {
		historias[i] = l.HistoriaTitulo
	}
	return historias, nil
}
