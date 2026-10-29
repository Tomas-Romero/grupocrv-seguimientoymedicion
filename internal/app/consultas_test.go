package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// lecturasFalsas simula las lecturas de proyectos, integrantes e items sin base de datos.
type lecturasFalsas struct {
	proyectos   []proyecto.Proyecto
	integrantes map[string][]integrante.Integrante
	items       map[string][]backlog.ItemBacklog
	err         error
}

func (l *lecturasFalsas) ListarProyectos(context.Context) ([]proyecto.Proyecto, error) {
	return append([]proyecto.Proyecto(nil), l.proyectos...), l.err
}

func (l *lecturasFalsas) BuscarProyecto(_ context.Context, id string) (proyecto.Proyecto, error) {
	if l.err != nil {
		return proyecto.Proyecto{}, l.err
	}
	for _, p := range l.proyectos {
		if p.ID == id {
			return p, nil
		}
	}
	return proyecto.Proyecto{}, app.ErrProyectoInexistente
}

func (l *lecturasFalsas) ListarIntegrantes(_ context.Context, proyectoID string) ([]integrante.Integrante, error) {
	return append([]integrante.Integrante(nil), l.integrantes[proyectoID]...), l.err
}

func (l *lecturasFalsas) ExisteProyecto(_ context.Context, id string) (bool, error) {
	for _, p := range l.proyectos {
		if p.ID == id {
			return true, l.err
		}
	}
	return false, l.err
}

func (l *lecturasFalsas) ListarItems(_ context.Context, proyectoID string) ([]backlog.ItemBacklog, error) {
	return append([]backlog.ItemBacklog(nil), l.items[proyectoID]...), l.err
}

func (l *lecturasFalsas) ObtenerItem(_ context.Context, itemID string) (backlog.ItemBacklog, bool, error) {
	if l.err != nil {
		return backlog.ItemBacklog{}, false, l.err
	}
	for _, items := range l.items {
		for _, i := range items {
			if i.ID == itemID {
				return i, true, nil
			}
		}
	}
	return backlog.ItemBacklog{}, false, nil
}

func proyectoDe(id, nombre string, inicio time.Time) proyecto.Proyecto {
	return proyecto.Proyecto{ID: id, Nombre: nombre, FechaInicio: inicio, FechaFin: inicio.AddDate(0, 2, 0)}
}

func itemDe(id string, numero int, prioridad backlog.Prioridad) backlog.ItemBacklog {
	return backlog.ItemBacklog{ID: id, Numero: numero, ProyectoID: "p1", Titulo: "Historia " + id, Prioridad: prioridad}
}

func ids[T any](elementos []T, id func(T) string) []string {
	salida := make([]string, len(elementos))
	for i, e := range elementos {
		salida[i] = id(e)
	}
	return salida
}

func igual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// T-011 / RN-T011-3, CL-T011-2: los proyectos salen por fecha de inicio descendente; el empate,
// por nombre sin distinguir mayusculas y despues por ID.
func TestListarProyectos_Orden(t *testing.T) {
	repo := &lecturasFalsas{proyectos: []proyecto.Proyecto{
		proyectoDe("a", "Zeta", fechaDe(2026, time.September, 1)),
		proyectoDe("b", "beta", fechaDe(2026, time.October, 1)),
		proyectoDe("c", "Alfa", fechaDe(2026, time.October, 1)),
		proyectoDe("e", "Alfa", fechaDe(2026, time.October, 1)),
		proyectoDe("d", "alfa", fechaDe(2026, time.October, 1)),
	}}

	got, err := app.NuevoListarProyectos(repo).Ejecutar(context.Background())
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	quiero := []string{"c", "d", "e", "b", "a"}
	if obtenido := ids(got, func(p proyecto.Proyecto) string { return p.ID }); !igual(obtenido, quiero) {
		t.Errorf("orden = %v, quiero %v", obtenido, quiero)
	}
}

// CL-T011-1: sin proyectos no es un error: la lista queda vacia.
func TestListarProyectos_Vacio(t *testing.T) {
	got, err := app.NuevoListarProyectos(&lecturasFalsas{}).Ejecutar(context.Background())
	if err != nil || len(got) != 0 {
		t.Errorf("Ejecutar() = %v, %v; quiero lista vacia y sin error", got, err)
	}
}

func TestListarProyectos_ErrorDeLaBase(t *testing.T) {
	errBase := errors.New("base caida")
	_, err := app.NuevoListarProyectos(&lecturasFalsas{err: errBase}).Ejecutar(context.Background())
	if !errors.Is(err, errBase) {
		t.Errorf("Ejecutar() = %v, quiero un error que envuelva %v", err, errBase)
	}
}

// RN-T011-5, CA-T011-1: el proyecto sale con sus integrantes, por nombre sin distinguir mayusculas
// y despues por email.
func TestConsultarProyecto_ConIntegrantesOrdenados(t *testing.T) {
	repo := &lecturasFalsas{
		proyectos: []proyecto.Proyecto{proyectoDe("p1", "Demo", fechaDe(2026, time.October, 1))},
		integrantes: map[string][]integrante.Integrante{"p1": {
			{ID: "1", Nombre: "vergara", Email: "v@x.test"},
			{ID: "2", Nombre: "Conforti", Email: "z@x.test"},
			{ID: "3", Nombre: "Conforti", Email: "a@x.test"},
			{ID: "4", Nombre: "Angelo", Email: "m@x.test"},
		}},
	}

	got, err := app.NuevoConsultarProyecto(repo).Ejecutar(context.Background(), "p1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if got.Proyecto.Nombre != "Demo" {
		t.Errorf("Proyecto.Nombre = %q, quiero Demo", got.Proyecto.Nombre)
	}
	quiero := []string{"4", "3", "2", "1"}
	if obtenido := ids(got.Integrantes, func(i integrante.Integrante) string { return i.ID }); !igual(obtenido, quiero) {
		t.Errorf("integrantes = %v, quiero %v", obtenido, quiero)
	}
}

// CL-T011-1: un proyecto sin integrantes no es un error.
func TestConsultarProyecto_SinIntegrantes(t *testing.T) {
	repo := &lecturasFalsas{proyectos: []proyecto.Proyecto{proyectoDe("p1", "Demo", fechaDe(2026, time.October, 1))}}

	got, err := app.NuevoConsultarProyecto(repo).Ejecutar(context.Background(), "p1")
	if err != nil || len(got.Integrantes) != 0 {
		t.Errorf("Ejecutar() = %+v, %v; quiero un proyecto sin integrantes y sin error", got, err)
	}
}

// RN-T011-10: un proyecto que no existe es app.ErrProyectoInexistente, con el ID.
func TestConsultarProyecto_Inexistente(t *testing.T) {
	_, err := app.NuevoConsultarProyecto(&lecturasFalsas{}).Ejecutar(context.Background(), "no-esta")
	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Errorf("Ejecutar() = %v, quiero ErrProyectoInexistente", err)
	}
}

func TestConsultarProyecto_ErrorDeLaBase(t *testing.T) {
	errBase := errors.New("base caida")
	repo := &lecturasFalsas{proyectos: []proyecto.Proyecto{proyectoDe("p1", "Demo", fechaDe(2026, time.October, 1))}, err: errBase}
	_, err := app.NuevoConsultarProyecto(repo).Ejecutar(context.Background(), "p1")
	if !errors.Is(err, errBase) {
		t.Errorf("Ejecutar() = %v, quiero un error que envuelva %v", err, errBase)
	}
}

// RN-T011-4, CL-T011-3: el backlog sale por prioridad (must, should, could, wont) y, dentro de
// cada una, por numero.
func TestListarBacklog_Orden(t *testing.T) {
	repo := &lecturasFalsas{
		proyectos: []proyecto.Proyecto{proyectoDe("p1", "Demo", fechaDe(2026, time.October, 1))},
		items: map[string][]backlog.ItemBacklog{"p1": {
			itemDe("a", 1, backlog.PrioridadWont),
			itemDe("b", 2, backlog.PrioridadCould),
			itemDe("c", 3, backlog.PrioridadMust),
			itemDe("d", 5, backlog.PrioridadShould),
			itemDe("e", 4, backlog.PrioridadMust),
			itemDe("f", 6, backlog.PrioridadShould),
		}},
	}

	got, err := app.NuevoListarBacklog(repo).Ejecutar(context.Background(), "p1", app.FiltroBacklog{})
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	quiero := []string{"c", "e", "d", "f", "b", "a"}
	if obtenido := ids(got, func(i backlog.ItemBacklog) string { return i.ID }); !igual(obtenido, quiero) {
		t.Errorf("orden = %v, quiero %v", obtenido, quiero)
	}
}

// CL-T011-1: un backlog vacio no es un error; un proyecto que no existe, si.
func TestListarBacklog_VacioEInexistente(t *testing.T) {
	repo := &lecturasFalsas{proyectos: []proyecto.Proyecto{proyectoDe("p1", "Demo", fechaDe(2026, time.October, 1))}}

	got, err := app.NuevoListarBacklog(repo).Ejecutar(context.Background(), "p1", app.FiltroBacklog{})
	if err != nil || len(got) != 0 {
		t.Errorf("proyecto sin items: %v, %v; quiero lista vacia y sin error", got, err)
	}
	if _, err := app.NuevoListarBacklog(repo).Ejecutar(context.Background(), "otro", app.FiltroBacklog{}); !errors.Is(err, app.ErrProyectoInexistente) {
		t.Errorf("proyecto inexistente: %v, quiero ErrProyectoInexistente", err)
	}
}

func TestListarBacklog_ErrorDeLaBase(t *testing.T) {
	errBase := errors.New("base caida")
	_, err := app.NuevoListarBacklog(&lecturasFalsas{err: errBase}).Ejecutar(context.Background(), "p1", app.FiltroBacklog{})
	if !errors.Is(err, errBase) {
		t.Errorf("Ejecutar() = %v, quiero un error que envuelva %v", err, errBase)
	}
}

// RN-T011-11, CL-T011-8: una historia solo se alcanza por la ruta de su proyecto.
func TestConsultarItem(t *testing.T) {
	repo := &lecturasFalsas{
		proyectos: []proyecto.Proyecto{proyectoDe("p1", "Uno", fechaDe(2026, time.October, 1)), proyectoDe("p2", "Dos", fechaDe(2026, time.October, 1))},
		items:     map[string][]backlog.ItemBacklog{"p1": {itemDe("a", 1, backlog.PrioridadMust)}},
	}
	consultar := app.NuevoConsultarItem(repo)

	casos := []struct {
		nombre             string
		proyectoID, itemID string
		quiereError        bool
	}{
		{"la historia es del proyecto", "p1", "a", false},
		{"la historia es de otro proyecto", "p2", "a", true},
		{"la historia no existe", "p1", "no-esta", true},
		{"el proyecto no existe", "no-esta", "a", true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			item, err := consultar.Ejecutar(context.Background(), c.proyectoID, c.itemID)

			if c.quiereError {
				if !errors.Is(err, app.ErrItemInexistente) {
					t.Errorf("Ejecutar() = %v, quiero ErrItemInexistente", err)
				}
				return
			}
			if err != nil || item.ID != "a" {
				t.Errorf("Ejecutar() = %+v, %v; quiero la historia a", item, err)
			}
		})
	}
}

func TestConsultarItem_ErrorDeLaBase(t *testing.T) {
	errBase := errors.New("base caida")
	_, err := app.NuevoConsultarItem(&lecturasFalsas{err: errBase}).Ejecutar(context.Background(), "p1", "a")
	if !errors.Is(err, errBase) {
		t.Errorf("Ejecutar() = %v, quiero un error que envuelva %v", err, errBase)
	}
}
