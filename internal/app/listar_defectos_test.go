package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// lecturaDeDefectosFalsa simula la lectura de defectos sin base de datos. Devuelve los defectos
// tal como los tiene, sin ordenar: el orden lo pone el caso de uso (RN-T021-4).
type lecturaDeDefectosFalsa struct {
	proyectos map[string]bool
	defectos  []app.DefectoListado
	errExiste error
	errListar error
}

func (l *lecturaDeDefectosFalsa) ExisteProyecto(_ context.Context, id string) (bool, error) {
	return l.proyectos[id], l.errExiste
}

func (l *lecturaDeDefectosFalsa) ListarDefectosDelProyecto(context.Context, string) ([]app.DefectoListado, error) {
	return append([]app.DefectoListado(nil), l.defectos...), l.errListar
}

func defectoListado(id string, estado defecto.Estado, creadoEn time.Time) app.DefectoListado {
	return app.DefectoListado{
		Defecto: defecto.Defecto{ID: id, ProyectoID: "p1", Descripcion: "Descripcion " + id,
			Severidad: defecto.SeveridadAlta, Estado: estado},
		HistoriaNumero: 1, HistoriaTitulo: "Historia", SprintDeteccion: 1, CreadoEn: creadoEn,
	}
}

func idsDeDefectos(listados []app.DefectoListado) []string {
	salida := make([]string, len(listados))
	for i, l := range listados {
		salida[i] = l.Defecto.ID
	}
	return salida
}

// T-021 / RN-T021-4: los defectos salen del mas reciente al mas antiguo.
func TestListarDefectos_MasRecientePrimero(t *testing.T) {
	base := time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC)
	repo := &lecturaDeDefectosFalsa{
		proyectos: map[string]bool{"p1": true},
		defectos: []app.DefectoListado{
			defectoListado("d2", defecto.EstadoAbierto, base.Add(time.Minute)),
			defectoListado("d1", defecto.EstadoAbierto, base),
			defectoListado("d3", defecto.EstadoAbierto, base.Add(2*time.Minute)),
		},
	}

	got, err := app.NuevoListarDefectos(repo).Ejecutar(context.Background(), "p1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if orden := idsDeDefectos(got); len(orden) != 3 || orden[0] != "d3" || orden[1] != "d2" || orden[2] != "d1" {
		t.Errorf("orden = %v, quiero d3, d2, d1", orden)
	}
}

// T-021 / CL-T021-8 / RN-T021-4: con la misma fecha de carga el orden es siempre el mismo, por ID
// descendente, sin importar como los entrega el repositorio.
func TestListarDefectos_MismaFechaDeCargaPorIDDescendente(t *testing.T) {
	fecha := time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC)
	entregas := [][]string{{"a", "b", "c"}, {"c", "b", "a"}, {"b", "c", "a"}}

	for _, entrega := range entregas {
		repo := &lecturaDeDefectosFalsa{proyectos: map[string]bool{"p1": true}}
		for _, id := range entrega {
			repo.defectos = append(repo.defectos, defectoListado(id, defecto.EstadoAbierto, fecha))
		}

		got, err := app.NuevoListarDefectos(repo).Ejecutar(context.Background(), "p1")
		if err != nil {
			t.Fatalf("Ejecutar: %v", err)
		}
		if orden := idsDeDefectos(got); len(orden) != 3 || orden[0] != "c" || orden[1] != "b" || orden[2] != "a" {
			t.Errorf("entregados %v: orden = %v, quiero c, b, a", entrega, orden)
		}
	}
}

// T-021 / RN-T021-3 / CL-T021-17: se listan los defectos de cualquier estado, tambien los
// descartados, con los datos de la lectura tal cual.
func TestListarDefectos_TodosLosEstados(t *testing.T) {
	base := time.Date(2026, time.October, 20, 12, 0, 0, 0, time.UTC)
	resuelto := defectoListado("d3", defecto.EstadoResuelto, base.Add(2*time.Minute))
	resuelto.Defecto.SprintResolucionID = "s2"
	resuelto.SprintResolucion = 2
	repo := &lecturaDeDefectosFalsa{
		proyectos: map[string]bool{"p1": true},
		defectos: []app.DefectoListado{
			defectoListado("d1", defecto.EstadoDescartado, base),
			defectoListado("d2", defecto.EstadoEnCorreccion, base.Add(time.Minute)),
			resuelto,
		},
	}

	got, err := app.NuevoListarDefectos(repo).Ejecutar(context.Background(), "p1")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("devolvio %d defectos, quiero los 3 (tambien el descartado)", len(got))
	}
	if got[0].Defecto.Estado != defecto.EstadoResuelto || got[0].SprintResolucion != 2 || got[0].HistoriaTitulo != "Historia" || got[0].SprintDeteccion != 1 {
		t.Errorf("el mas reciente = %+v, quiero el resuelto en el Sprint 2 con sus datos de lectura", got[0])
	}
	if got[2].Defecto.Estado != defecto.EstadoDescartado {
		t.Errorf("el mas antiguo = %+v, quiero el descartado", got[2].Defecto)
	}
}

// T-021 / RN-T021-3 / CL-T021-1: un proyecto sin defectos es una lista vacia, no un error.
func TestListarDefectos_SinDefectos(t *testing.T) {
	repo := &lecturaDeDefectosFalsa{proyectos: map[string]bool{"p1": true}}

	got, err := app.NuevoListarDefectos(repo).Ejecutar(context.Background(), "p1")
	if err != nil || got == nil || len(got) != 0 {
		t.Errorf("Ejecutar = %v, %v; quiero una lista vacia y sin error", got, err)
	}
}

// T-021 / RN-T021-3: un proyecto que no existe es ErrProyectoInexistente, y no se lee nada mas.
func TestListarDefectos_ProyectoInexistente(t *testing.T) {
	repo := &lecturaDeDefectosFalsa{proyectos: map[string]bool{"p1": true}, errListar: errors.New("no se tendria que listar")}

	_, err := app.NuevoListarDefectos(repo).Ejecutar(context.Background(), "otro")
	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Errorf("Ejecutar = %v, quiero ErrProyectoInexistente", err)
	}
}

// T-021 / CL-T021-18: una falla de la base se devuelve envuelta, sin tocarla.
func TestListarDefectos_FallaDeLaBase(t *testing.T) {
	caida := errors.New("base caida")
	casos := []struct {
		nombre string
		repo   *lecturaDeDefectosFalsa
	}{
		{"al consultar el proyecto", &lecturaDeDefectosFalsa{proyectos: map[string]bool{"p1": true}, errExiste: caida}},
		{"al listar", &lecturaDeDefectosFalsa{proyectos: map[string]bool{"p1": true}, errListar: caida}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := app.NuevoListarDefectos(c.repo).Ejecutar(context.Background(), "p1")
			if !errors.Is(err, caida) || errors.Is(err, app.ErrProyectoInexistente) {
				t.Errorf("Ejecutar = %v, quiero la falla de la base envuelta", err)
			}
		})
	}
}
