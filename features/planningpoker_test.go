package features_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// pokerSobreApp conecta el area PlanningPoker con el caso de uso de US-014 sobre la base
// en memoria del escenario (ADR 0002). Los Entonces leen la base directamente, porque
// todavia no hay casos de uso de lectura de las sesiones.
type pokerSobreApp struct {
	escenario *escenario
	iniciar   *app.IniciarSesionPlanningPoker
}

var _ steps.PlanningPoker = (*pokerSobreApp)(nil)

func nuevoPokerSobreApp(e *escenario) *pokerSobreApp {
	return &pokerSobreApp{escenario: e, iniciar: app.NuevoIniciarSesionPlanningPoker(e.base)}
}

// IniciarSesion no valida nada: lo hace el caso de uso. Un proyecto o una historia que el
// escenario no creo se mandan con un ID que no esta en la base, asi el rechazo lo produce
// la aplicacion (RN-014-1).
func (p *pokerSobreApp) IniciarSesion(proyecto, titulo string) error {
	proyectoID, ok := p.escenario.proyectos[proyecto]
	if !ok {
		proyectoID = idDeProyectoInexistente
	}
	itemID, _ := p.historia(titulo)
	_, err := p.iniciar.Ejecutar(context.Background(), proyectoID, itemID)
	return err
}

func (p *pokerSobreApp) EstadoDeLaSesion(titulo string) (string, int, error) {
	sesion, _, err := p.sesion(titulo)
	return string(sesion.Estado), sesion.Ronda, err
}

// ParticipantesDeLaSesion traduce los IDs de los participantes a los nombres con que el
// escenario conoce a los integrantes.
func (p *pokerSobreApp) ParticipantesDeLaSesion(titulo string) ([]string, error) {
	sesion, proyectoID, err := p.sesion(titulo)
	if err != nil {
		return nil, err
	}
	nombres := map[string]string{}
	for _, i := range p.escenario.base.Integrantes(proyectoID) {
		nombres[i.ID] = i.Nombre
	}
	participantes := make([]string, 0, len(sesion.Participantes))
	for _, id := range sesion.Participantes {
		participantes = append(participantes, nombres[id])
	}
	return participantes, nil
}

// Votar es de US-015.
func (p *pokerSobreApp) Votar(_, _ string, _ int) error { return errVotoSinConectar }

// QuienesVotaron es de US-015.
func (p *pokerSobreApp) QuienesVotaron(_ string) ([]string, error) { return nil, errVotoSinConectar }

// QuienesFaltan es de US-015.
func (p *pokerSobreApp) QuienesFaltan(_ string) ([]string, error) { return nil, errVotoSinConectar }

// VotosOcultos es de US-015.
func (p *pokerSobreApp) VotosOcultos(_ string) (bool, error) { return false, errVotoSinConectar }

// errVotoSinConectar lo devuelven las operaciones de votar mientras no estan conectadas.
var errVotoSinConectar = errors.New("votar todavia no esta conectado a la aplicacion")

// sesion devuelve la sesion abierta de la historia y el proyecto de la historia.
func (p *pokerSobreApp) sesion(titulo string) (planningpoker.SesionPlanningPoker, string, error) {
	itemID, proyectoID := p.historia(titulo)
	sesion, ok := p.escenario.base.SesionAbiertaDeLaHistoria(itemID)
	if !ok {
		return planningpoker.SesionPlanningPoker{}, "", fmt.Errorf("la historia %q no tiene una sesion abierta", titulo)
	}
	return sesion, proyectoID, nil
}

// historia busca la historia por titulo en todos los proyectos del escenario: los
// escenarios no repiten titulos (docs/diccionario-steps.md). Si no esta, devuelve un ID
// que no existe.
func (p *pokerSobreApp) historia(titulo string) (itemID, proyectoID string) {
	for _, id := range p.escenario.proyectos {
		for _, item := range p.escenario.base.Items(id) {
			if item.Titulo == titulo {
				return item.ID, id
			}
		}
	}
	return idDeItemInexistente, ""
}
