package features_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/features/steps"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// errSinConectar lo devuelven las operaciones del area Backlog que todavia no
// tienen caso de uso: si un escenario las necesita, falla en voz alta.
var errSinConectar = errors.New("la operacion todavia no esta conectada a la aplicacion")

// idDeProyectoInexistente es el ID con el que se llama al caso de uso cuando el
// escenario nombra un proyecto que no creo (por ejemplo, despues de
// `Dado ningún proyecto`). Ningun proyecto de la base en memoria lo tiene, asi
// que el rechazo lo produce la aplicacion (RN-005-9) y no este adaptador.
const idDeProyectoInexistente = "00000000-0000-0000-0000-000000000000"

// backlogSobreApp conecta el area Backlog con el caso de uso de US-005 sobre la
// base en memoria del escenario (ADR 0002). Las consultas de los Entonces leen
// la base directamente, porque todavia no hay casos de uso de lectura (US-007).
type backlogSobreApp struct {
	escenario *escenario
	crear     *app.CrearItemBacklog

	// proyectoDe guarda en que proyecto quedo cada historia, por titulo. Los
	// escenarios no repiten titulos (docs/diccionario-steps.md).
	proyectoDe map[string]string
}

var _ steps.Backlog = (*backlogSobreApp)(nil)

func nuevoBacklogSobreApp(e *escenario) *backlogSobreApp {
	return &backlogSobreApp{
		escenario:  e,
		crear:      app.NuevoCrearItemBacklog(e.base, ahoraEnEscenarios),
		proyectoDe: map[string]string{},
	}
}

func (b *backlogSobreApp) CrearHistoria(proyecto string, historia steps.DatosHistoria) error {
	proyectoID, ok := b.escenario.proyectos[proyecto]
	if !ok {
		proyectoID = idDeProyectoInexistente
	}
	item, err := b.crear.Ejecutar(context.Background(), proyectoID, backlog.DatosItem{
		Titulo:      historia.Titulo,
		Descripcion: historia.Descripcion,
		Prioridad:   backlog.Prioridad(historia.Prioridad),
		Criterios:   historia.Criterios,
	})
	if err != nil {
		return err
	}
	// Se guarda con el titulo que quedo registrado, ya recortado.
	b.proyectoDe[item.Titulo] = proyectoID
	return nil
}

func (b *backlogSobreApp) CantidadHistoriasEnBacklog(proyecto string) (int, error) {
	proyectoID, ok := b.escenario.proyectos[proyecto]
	if !ok {
		return 0, fmt.Errorf("el proyecto %q no se creo en este escenario", proyecto)
	}
	return len(b.escenario.base.Items(proyectoID)), nil
}

func (b *backlogSobreApp) PrioridadDeHistoria(titulo string) (string, error) {
	item, err := b.historia(titulo)
	return string(item.Prioridad), err
}

func (b *backlogSobreApp) PuntosDeHistoria(titulo string) (int, bool, error) {
	item, err := b.historia(titulo)
	if err != nil {
		return 0, false, err
	}
	puntos, estimada := item.StoryPoints.Valor()
	return puntos, estimada, nil
}

func (b *backlogSobreApp) NumeroDeHistoria(titulo string) (int, error) {
	item, err := b.historia(titulo)
	return item.Numero, err
}

func (b *backlogSobreApp) EstadoDeHistoria(titulo string) (string, error) {
	item, err := b.historia(titulo)
	return string(item.Estado), err
}

func (b *backlogSobreApp) CriteriosDeHistoria(titulo string) ([]string, error) {
	item, err := b.historia(titulo)
	return item.Criterios, err
}

// CambiarPrioridadHistoria es de US-006.
func (b *backlogSobreApp) CambiarPrioridadHistoria(_, _ string) error {
	return errSinConectar
}

// EstimarHistoria es de US-013.
func (b *backlogSobreApp) EstimarHistoria(_ string, _ int) error {
	return errSinConectar
}

// HistoriaEstaEnBacklog depende de los sprints (US-009 y US-011).
func (b *backlogSobreApp) HistoriaEstaEnBacklog(_ string) (bool, error) {
	return false, errSinConectar
}

// historia busca en la base el item que el escenario creo con ese titulo.
func (b *backlogSobreApp) historia(titulo string) (backlog.ItemBacklog, error) {
	proyectoID, ok := b.proyectoDe[titulo]
	if !ok {
		return backlog.ItemBacklog{}, fmt.Errorf("la historia %q no se creo en este escenario", titulo)
	}
	for _, item := range b.escenario.base.Items(proyectoID) {
		if item.Titulo == titulo {
			return item, nil
		}
	}
	return backlog.ItemBacklog{}, fmt.Errorf("la historia %q no esta en la base", titulo)
}
