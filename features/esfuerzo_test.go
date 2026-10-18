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
// base en memoria del escenario (ADR 0002). Los Entonces leen la base directamente,
// porque todavia no hay casos de uso de lectura (US-020).
type esfuerzoSobreApp struct {
	escenario *escenario
	registrar *app.RegistrarEsfuerzo
}

var _ steps.Esfuerzo = (*esfuerzoSobreApp)(nil)

func nuevoEsfuerzoSobreApp(e *escenario) *esfuerzoSobreApp {
	return &esfuerzoSobreApp{escenario: e, registrar: app.NuevoRegistrarEsfuerzo(e.base, ahoraEnEscenarios)}
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
