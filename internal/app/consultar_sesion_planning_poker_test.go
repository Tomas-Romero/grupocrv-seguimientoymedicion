package app_test

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// repositorioConsultaSesionFalso simula las lecturas de la pantalla de la sesion (T-020).
type repositorioConsultaSesionFalso struct {
	sesion       planningpoker.SesionPlanningPoker
	existeSesion bool
	item         backlog.ItemBacklog
	votos        map[string]int
	anteriores   map[int]map[string]int
	abiertas     map[string]string

	errSesion, errItem, errVotos, errRondas, errAbiertas error
}

func (r *repositorioConsultaSesionFalso) ObtenerSesion(_ context.Context, _ string) (planningpoker.SesionPlanningPoker, bool, error) {
	return r.sesion, r.existeSesion, r.errSesion
}

func (r *repositorioConsultaSesionFalso) ObtenerItem(_ context.Context, _ string) (backlog.ItemBacklog, bool, error) {
	return r.item, true, r.errItem
}

func (r *repositorioConsultaSesionFalso) VotacionDeLaRonda(_ context.Context, _ string) (planningpoker.Votacion, error) {
	return planningpoker.NuevaVotacion(r.sesion, r.votos), r.errVotos
}

func (r *repositorioConsultaSesionFalso) RondasAnteriores(_ context.Context, _ string) ([]app.RondaAnterior, error) {
	rondas := []app.RondaAnterior{}
	for n := 1; n < r.sesion.Ronda; n++ {
		rondas = append(rondas, app.RondaAnterior{Numero: n, Votacion: planningpoker.NuevaVotacionDeRonda(r.sesion, n, r.anteriores[n])})
	}
	return rondas, r.errRondas
}

func (r *repositorioConsultaSesionFalso) SesionesAbiertas(_ context.Context, _ string) (map[string]string, error) {
	return r.abiertas, r.errAbiertas
}

// sesionDelProyecto es una sesion del proyecto 1 en la ronda 2, votando: ana voto en la ronda
// 2, y en la ronda 1 votaron 3 y 8.
func sesionDelProyecto() *repositorioConsultaSesionFalso {
	return &repositorioConsultaSesionFalso{
		sesion: planningpoker.SesionPlanningPoker{
			ID: "sesion-1", ItemID: "item-1", ProyectoID: "proyecto-1",
			Estado: planningpoker.EstadoVotando, Ronda: 2, Participantes: []string{"ana", "beto"},
		},
		existeSesion: true,
		item:         backlog.ItemBacklog{ID: "item-1", Titulo: "Alta de proyectos"},
		votos:        map[string]int{"ana": 13},
		anteriores:   map[int]map[string]int{1: {"ana": 3, "beto": 8}},
	}
}

func consultarSesion(repo *repositorioConsultaSesionFalso, proyectoID string) (app.DetalleDeSesion, error) {
	return app.NuevoConsultarSesionPlanningPoker(repo).Ejecutar(context.Background(), proyectoID, "sesion-1")
}

// T-020 / CA-T020-2, CA-T020-5, RN-T020-2, RN-T020-3: votando, la lectura dice quien voto y quien
// falta, sin el resultado de la ronda actual; las rondas anteriores vienen con su resultado.
func TestConsultarSesionPlanningPoker_Votando(t *testing.T) {
	repo := sesionDelProyecto()

	d, err := consultarSesion(repo, "proyecto-1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if d.Sesion.ID != "sesion-1" || d.Historia.Titulo != "Alta de proyectos" {
		t.Errorf("detalle = %+v, se esperaba la sesion y su historia", d)
	}
	if !slices.Equal(d.Votaron, []string{"ana"}) || !slices.Equal(d.Faltan, []string{"beto"}) {
		t.Errorf("votaron %q y faltan %q, se esperaba ana y beto", d.Votaron, d.Faltan)
	}
	if d.Revelada || !reflect.DeepEqual(d.Resultado, planningpoker.Resultado{}) {
		t.Errorf("votando la lectura trae el resultado de la ronda actual: %+v", d.Resultado)
	}
	if len(d.Anteriores) != 1 || d.Anteriores[0].Numero != 1 || !d.Anteriores[0].Resultado.Divergencia ||
		d.Anteriores[0].Resultado.Minimo != 3 || d.Anteriores[0].Resultado.Maximo != 8 {
		t.Errorf("rondas anteriores = %+v, se esperaba la ronda 1 de 3 a 8", d.Anteriores)
	}
}

// T-020 / CA-T020-3: con la sesion revelada la lectura trae el resultado de la ronda actual.
func TestConsultarSesionPlanningPoker_Revelada(t *testing.T) {
	repo := sesionDelProyecto()
	repo.sesion.Estado = planningpoker.EstadoRevelada
	repo.votos = map[string]int{"ana": 5, "beto": 5}

	d, err := consultarSesion(repo, "proyecto-1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	esperado := planningpoker.Resultado{
		Votos:    []planningpoker.Voto{{ParticipanteID: "ana", Valor: 5}, {ParticipanteID: "beto", Valor: 5}},
		Faltaron: []string{}, Minimo: 5, Maximo: 5,
	}
	if !d.Revelada || !reflect.DeepEqual(d.Resultado, esperado) {
		t.Errorf("resultado = (%t, %+v), se esperaba %+v", d.Revelada, d.Resultado, esperado)
	}
}

// T-020 / RN-T020-4, CL-T020-9: una sesion que no existe, o de otro proyecto, es
// ErrSesionInexistente.
func TestConsultarSesionPlanningPoker_LoQueNoExiste(t *testing.T) {
	repo := sesionDelProyecto()
	if _, err := consultarSesion(repo, "proyecto-2"); !errors.Is(err, app.ErrSesionInexistente) {
		t.Errorf("sesion de otro proyecto: %v, se esperaba ErrSesionInexistente", err)
	}
	repo.existeSesion = false
	if _, err := consultarSesion(repo, "proyecto-1"); !errors.Is(err, app.ErrSesionInexistente) {
		t.Errorf("sesion inexistente: %v, se esperaba ErrSesionInexistente", err)
	}
}

// T-020 / CL-T020-2: el backlog sabe que sesion abierta tiene cada historia.
func TestConsultarSesionPlanningPoker_Abiertas(t *testing.T) {
	repo := sesionDelProyecto()
	repo.abiertas = map[string]string{"item-1": "sesion-1"}

	abiertas, err := app.NuevoConsultarSesionPlanningPoker(repo).AbiertasDelProyecto(context.Background(), "proyecto-1")

	if err != nil || !reflect.DeepEqual(abiertas, repo.abiertas) {
		t.Errorf("AbiertasDelProyecto = %v (%v), se esperaba %v", abiertas, err, repo.abiertas)
	}
}

// Los fallos de infraestructura llegan envueltos con contexto.
func TestConsultarSesionPlanningPoker_FallosDeInfraestructura(t *testing.T) {
	errBase := errors.New("base caida")
	for nombre, romper := range map[string]func(*repositorioConsultaSesionFalso){
		"sesion": func(r *repositorioConsultaSesionFalso) { r.errSesion = errBase },
		"item":   func(r *repositorioConsultaSesionFalso) { r.errItem = errBase },
		"votos":  func(r *repositorioConsultaSesionFalso) { r.errVotos = errBase },
		"rondas": func(r *repositorioConsultaSesionFalso) { r.errRondas = errBase },
	} {
		t.Run(nombre, func(t *testing.T) {
			repo := sesionDelProyecto()
			romper(repo)
			if _, err := consultarSesion(repo, "proyecto-1"); !errors.Is(err, errBase) {
				t.Errorf("error = %v, se esperaba que envolviera %v", err, errBase)
			}
		})
	}
	repo := sesionDelProyecto()
	repo.errAbiertas = errBase
	if _, err := app.NuevoConsultarSesionPlanningPoker(repo).AbiertasDelProyecto(context.Background(), "proyecto-1"); !errors.Is(err, errBase) {
		t.Errorf("AbiertasDelProyecto: %v, se esperaba que envolviera %v", err, errBase)
	}
}
