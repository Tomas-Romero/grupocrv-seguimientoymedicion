package features_test

import (
	"context"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// defectosSobreApp conecta el area Defectos con el caso de uso de US-022 sobre la base en
// memoria del escenario (ADR 0002). Los Entonces leen la base directamente: listar defectos es
// US-024.
type defectosSobreApp struct {
	escenario *escenario
	registrar *app.RegistrarDefecto
}

var _ steps.Defectos = (*defectosSobreApp)(nil)

func nuevoDefectosSobreApp(e *escenario) *defectosSobreApp {
	return &defectosSobreApp{escenario: e, registrar: app.NuevoRegistrarDefecto(e.base)}
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
