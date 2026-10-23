package features_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// idDeSesionInexistente es el ID con el que se llama a VotarEnSesion cuando la historia
// no tiene una sesion abierta. El rechazo (ErrSesionInexistente) lo produce la aplicacion.
const idDeSesionInexistente = "00000000-0000-0000-0000-000000000000"

// pokerSobreApp conecta el area PlanningPoker con los casos de uso de US-014, US-015 y US-016
// sobre la base en memoria del escenario (ADR 0002). Los Entonces leen la base
// directamente, porque todavia no hay casos de uso de lectura de las sesiones.
type pokerSobreApp struct {
	escenario *escenario
	iniciar   *app.IniciarSesionPlanningPoker
	votar     *app.VotarEnSesion
	revelar   *app.RevelarVotos
	// reveladas guarda la ultima revelacion de cada historia, para preguntar por la divergencia.
	reveladas map[string]planningpoker.Revelacion
}

var _ steps.PlanningPoker = (*pokerSobreApp)(nil)

func nuevoPokerSobreApp(e *escenario) *pokerSobreApp {
	return &pokerSobreApp{
		escenario: e,
		iniciar:   app.NuevoIniciarSesionPlanningPoker(e.base),
		votar:     app.NuevoVotarEnSesion(e.base),
		revelar:   app.NuevoRevelarVotos(e.base),
		reveladas: map[string]planningpoker.Revelacion{},
	}
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

// ParticipantesDeLaSesion devuelve los nombres de los participantes, en orden de alta.
func (p *pokerSobreApp) ParticipantesDeLaSesion(titulo string) ([]string, error) {
	sesion, proyectoID, err := p.sesion(titulo)
	if err != nil {
		return nil, err
	}
	return p.nombres(proyectoID, sesion.Participantes), nil
}

// nombres traduce IDs de integrantes del proyecto a los nombres con que el escenario los
// conoce, en el mismo orden.
func (p *pokerSobreApp) nombres(proyectoID string, ids []string) []string {
	porID := map[string]string{}
	for _, i := range p.escenario.base.Integrantes(proyectoID) {
		porID[i.ID] = i.Nombre
	}
	nombres := make([]string, 0, len(ids))
	for _, id := range ids {
		nombres = append(nombres, porID[id])
	}
	return nombres
}

// Votar vota con el caso de uso de US-015 en la sesion abierta de la historia. No valida
// nada: lo hace el caso de uso. Una historia sin sesion o un integrante que no esta en el
// proyecto de la historia se mandan con un ID que no existe (RN-015-1).
func (p *pokerSobreApp) Votar(integrante, titulo string, valor int) error {
	itemID, proyectoID := p.historia(titulo)
	sesionID := idDeSesionInexistente
	if sesion, ok := p.escenario.base.SesionAbiertaDeLaHistoria(itemID); ok {
		sesionID = sesion.ID
	}
	integranteID := idDeIntegranteInexistente
	for _, i := range p.escenario.base.Integrantes(proyectoID) {
		if i.Nombre == integrante {
			integranteID = i.ID
		}
	}
	_, err := p.votar.Ejecutar(context.Background(), sesionID, integranteID, valor)
	return err
}

// Revelar revela con el caso de uso de US-016 la sesion abierta de la historia. No valida
// nada: lo hace el caso de uso. Una historia sin sesion o un integrante que no esta en el
// proyecto de la historia se mandan con un ID que no existe (RN-016-1).
func (p *pokerSobreApp) Revelar(integrante, titulo string) error {
	itemID, proyectoID := p.historia(titulo)
	sesionID := idDeSesionInexistente
	if sesion, ok := p.escenario.base.SesionAbiertaDeLaHistoria(itemID); ok {
		sesionID = sesion.ID
	}
	integranteID := idDeIntegranteInexistente
	for _, i := range p.escenario.base.Integrantes(proyectoID) {
		if i.Nombre == integrante {
			integranteID = i.ID
		}
	}
	revelacion, err := p.revelar.Ejecutar(context.Background(), sesionID, integranteID)
	if err == nil {
		p.reveladas[titulo] = revelacion
	}
	return err
}

// VotosRevelados lee los votos de la ronda actual de la votacion guardada: con la sesion
// votando se niegan (RN-015-7).
func (p *pokerSobreApp) VotosRevelados(titulo string) ([]steps.VotoRevelado, error) {
	votacion, proyectoID, err := p.votacion(titulo)
	if err != nil {
		return nil, err
	}
	valores, err := votacion.Valores()
	if err != nil {
		return nil, err
	}
	votos := []steps.VotoRevelado{}
	for _, id := range votacion.Votaron() {
		nombre := p.nombres(proyectoID, []string{id})[0]
		votos = append(votos, steps.VotoRevelado{Integrante: nombre, Valor: valores[id]})
	}
	return votos, nil
}

// HayDivergencia dice si la ultima revelacion de la historia detecto divergencia.
func (p *pokerSobreApp) HayDivergencia(titulo string) (bool, error) {
	revelacion, ok := p.reveladas[titulo]
	if !ok {
		return false, fmt.Errorf("la sesion de la historia %q no se revelo en este escenario", titulo)
	}
	return revelacion.Divergencia, nil
}

// QuienesVotaron devuelve los nombres de los que votaron en la ronda actual.
func (p *pokerSobreApp) QuienesVotaron(titulo string) ([]string, error) {
	votacion, proyectoID, err := p.votacion(titulo)
	if err != nil {
		return nil, err
	}
	return p.nombres(proyectoID, votacion.Votaron()), nil
}

// QuienesFaltan devuelve los nombres de los que todavia no votaron en la ronda actual.
func (p *pokerSobreApp) QuienesFaltan(titulo string) ([]string, error) {
	votacion, proyectoID, err := p.votacion(titulo)
	if err != nil {
		return nil, err
	}
	return p.nombres(proyectoID, votacion.Faltan()), nil
}

// VotosOcultos pide los valores a la votacion: estan ocultos si los niega con
// ErrVotosOcultos (RN-015-7).
func (p *pokerSobreApp) VotosOcultos(titulo string) (bool, error) {
	votacion, _, err := p.votacion(titulo)
	if err != nil {
		return false, err
	}
	_, err = votacion.Valores()
	if errors.Is(err, planningpoker.ErrVotosOcultos) {
		return true, nil
	}
	return false, err
}

// votacion devuelve la votacion de la ronda actual de la sesion abierta de la historia, y
// el proyecto de la historia.
func (p *pokerSobreApp) votacion(titulo string) (planningpoker.Votacion, string, error) {
	sesion, proyectoID, err := p.sesion(titulo)
	if err != nil {
		return planningpoker.Votacion{}, "", err
	}
	votacion, _ := p.escenario.base.VotacionDeLaSesion(sesion.ID)
	return votacion, proyectoID, nil
}

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
