package features_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
)

// idDeIntegranteInexistente es el ID con el que se llama al caso de uso cuando el
// escenario nombra a un integrante que no esta en el proyecto. Igual que con el
// proyecto y la historia, el rechazo (ErrIntegranteInexistente) lo produce la
// aplicacion y no este adaptador.
const idDeIntegranteInexistente = "00000000-0000-0000-0000-000000000000"

// esfuerzoSobreApp conecta el area Esfuerzo con el caso de uso de US-019 sobre la
// base en memoria del escenario (ADR 0002). Los Entonces de US-019 leen la base
// directamente; la consulta de US-020 usa el caso de uso real.
type esfuerzoSobreApp struct {
	escenario *escenario
	registrar *app.RegistrarEsfuerzo
	consultar *app.ConsultarEsfuerzo
}

var _ steps.Esfuerzo = (*esfuerzoSobreApp)(nil)

func nuevoEsfuerzoSobreApp(e *escenario) *esfuerzoSobreApp {
	return &esfuerzoSobreApp{
		escenario: e,
		registrar: app.NuevoRegistrarEsfuerzo(e.base, ahoraEnEscenarios),
		consultar: app.NuevoConsultarEsfuerzo(e.base),
	}
}

// RegistrarEsfuerzo no valida nada: lo hace el caso de uso. Busca al integrante y
// la historia por nombre DENTRO del proyecto; si el escenario nombra uno que no
// esta ahi (por ejemplo, de otro proyecto), llama con un ID que no existe.
func (a *esfuerzoSobreApp) RegistrarEsfuerzo(proyecto, integrante, historia string, fecha time.Time, actividad string, horas float64) error {
	proyectoID, ok := a.escenario.proyectos[proyecto]
	if !ok {
		proyectoID = idDeProyectoInexistente
	}
	itemID, integranteID := idDeItemInexistente, idDeIntegranteInexistente
	for _, item := range a.escenario.base.Items(proyectoID) {
		if item.Titulo == historia {
			itemID = item.ID
		}
	}
	for _, i := range a.escenario.base.Integrantes(proyectoID) {
		if i.Nombre == integrante {
			integranteID = i.ID
		}
	}
	_, err := a.registrar.Ejecutar(context.Background(), proyectoID, itemID, integranteID, fecha, actividad, horas)
	return err
}

func (a *esfuerzoSobreApp) CantidadRegistros(proyecto string) (int, error) {
	proyectoID, ok := a.escenario.proyectos[proyecto]
	if !ok {
		return 0, fmt.Errorf("el proyecto %q no se creo en este escenario", proyecto)
	}
	return len(a.escenario.base.Esfuerzos(proyectoID)), nil
}

func (a *esfuerzoSobreApp) HorasDeHistoria(proyecto, historia string) (float64, error) {
	proyectoID, ok := a.escenario.proyectos[proyecto]
	if !ok {
		return 0, fmt.Errorf("el proyecto %q no se creo en este escenario", proyecto)
	}
	itemID := ""
	for _, item := range a.escenario.base.Items(proyectoID) {
		if item.Titulo == historia {
			itemID = item.ID
		}
	}
	if itemID == "" {
		return 0, fmt.Errorf("la historia %q no esta en el proyecto %q", historia, proyecto)
	}
	var total float64
	for _, r := range a.escenario.base.Esfuerzos(proyectoID) {
		if r.ItemID == itemID {
			total += r.Horas
		}
	}
	return total, nil
}

// ConsultarEsfuerzo no valida nada: lo hace el caso de uso. Busca la historia, el integrante y el
// sprint por nombre DENTRO del proyecto; lo que el escenario nombra y no esta ahi se manda con
// un ID que no existe. Devuelve los totales indexados por nombre.
func (a *esfuerzoSobreApp) ConsultarEsfuerzo(proyecto, historia, integrante, nombreDelSprint string) (steps.ConsultaDeEsfuerzo, error) {
	proyectoID, ok := a.escenario.proyectos[proyecto]
	if !ok {
		proyectoID = idDeProyectoInexistente
	}
	var filtro app.FiltroEsfuerzo
	if historia != "" {
		filtro.ItemID = idDeItemInexistente
		for _, item := range a.escenario.base.Items(proyectoID) {
			if item.Titulo == historia {
				filtro.ItemID = item.ID
			}
		}
	}
	if integrante != "" {
		filtro.IntegranteID = idDeIntegranteInexistente
		for _, i := range a.escenario.base.Integrantes(proyectoID) {
			if i.Nombre == integrante {
				filtro.IntegranteID = i.ID
			}
		}
	}
	if nombreDelSprint != "" {
		filtro.SprintID = idDeSprintInexistente
		for _, s := range a.escenario.base.Sprints(proyectoID) {
			if s.Nombre() == nombreDelSprint {
				filtro.SprintID = s.ID
			}
		}
	}

	c, err := a.consultar.Ejecutar(context.Background(), proyectoID, filtro)
	if err != nil {
		return steps.ConsultaDeEsfuerzo{}, err
	}

	resultado := steps.ConsultaDeEsfuerzo{
		TotalHoras:    c.TotalHoras,
		PorHistoria:   map[string]float64{},
		PorIntegrante: map[string]float64{},
		PorSprint:     map[string]float64{},
	}
	for _, r := range c.Registros {
		resultado.Actividades = append(resultado.Actividades, r.Registro.Actividad)
	}
	for _, t := range c.PorHistoria {
		resultado.PorHistoria[t.Titulo] = t.Horas
	}
	for _, t := range c.PorIntegrante {
		resultado.PorIntegrante[t.Nombre] = t.Horas
	}
	for _, t := range c.PorSprint {
		if t.Numero == 0 {
			resultado.FueraDeSprint = t.Horas
			continue
		}
		resultado.PorSprint[fmt.Sprintf("Sprint %d", t.Numero)] = t.Horas
	}
	return resultado, nil
}
