package memoria_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// idSinDuenoAsignaciones es un ID con formato de UUID que ninguna fila tiene.
const idSinDuenoAsignaciones = "00000000-0000-0000-0000-000000009999"

// baseParaAsignar es una base con un proyecto "Demo" con un sprint abierto y dos items:
// uno listo para asignar (estimado y con un criterio) y uno de otro proyecto.
type baseParaAsignar struct {
	base     *memoria.Base
	proyecto string
	sprint   string
	item     string
	otro     string // proyecto distinto
}

func nuevaBaseParaAsignar(t *testing.T) baseParaAsignar {
	t.Helper()
	ctx := context.Background()
	base := memoria.NuevaBase()
	proyecto := proyectoParaSprints(t, base, "Demo")
	creado, err := base.RegistrarSprint(ctx, sprintParaProyecto(t, proyecto, diaDeSprint(10, 6), diaDeSprint(10, 19)))
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}
	b := baseParaAsignar{base: base, proyecto: proyecto, sprint: creado.ID, otro: proyectoParaSprints(t, base, "Otro")}
	b.item = b.itemEn(t, proyecto, "Crear proyecto", true, 1)
	return b
}

// itemEn registra un item estimado o no y con cierta cantidad de criterios.
func (b baseParaAsignar) itemEn(t *testing.T, proyecto, titulo string, estimado bool, criterios int) string {
	t.Helper()
	cs := make([]string, criterios)
	for i := range cs {
		cs[i] = "Criterio"
	}
	item, err := backlog.NuevoItem(backlog.DatosItem{Titulo: titulo, Prioridad: backlog.PrioridadMust, Criterios: cs}, diaDeSprint(10, 6))
	if err != nil {
		t.Fatalf("NuevoItem: %v", err)
	}
	registrado, err := b.base.RegistrarItem(context.Background(), proyecto, item)
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	if estimado {
		puntos, _ := backlog.NuevosStoryPoints(5)
		if err := b.base.GuardarEstimacion(context.Background(), registrado.ID, puntos); err != nil {
			t.Fatalf("GuardarEstimacion: %v", err)
		}
	}
	return registrado.ID
}

// US-009 / CA-009-1, RN-009-9: asignar deja al item en el sprint abierto y no lo modifica:
// conserva su estado, sus Story Points, sus criterios y su prioridad.
func TestBase_AsignarItemASprint(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaAsignar(t)
	antes, _, _ := b.base.ObtenerItem(ctx, b.item)

	if err := b.base.AsignarItemASprint(ctx, b.item, b.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}

	actual, ok := b.base.SprintActualDelItem(b.item)
	if !ok || actual.ID != b.sprint {
		t.Errorf("sprint actual = (%+v, %t), se esperaba el sprint asignado", actual, ok)
	}
	despues, _, _ := b.base.ObtenerItem(ctx, b.item)
	puntosAntes, _ := antes.StoryPoints.Valor()
	puntosDespues, _ := despues.StoryPoints.Valor()
	if despues.Estado != antes.Estado || despues.Prioridad != antes.Prioridad || puntosDespues != puntosAntes || len(despues.Criterios) != len(antes.Criterios) {
		t.Errorf("el item cambio al asignarlo: antes %+v, despues %+v", antes, despues)
	}
}

// US-009 / RN-009-6, RN-009-13: un item sin asignar no tiene sprint actual, y ItemEnSprintAbierto
// refleja la asignacion; un ID mal formado o inexistente no esta en ningun sprint.
func TestBase_ItemEnSprintAbierto(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaAsignar(t)

	if en, err := b.base.ItemEnSprintAbierto(ctx, b.item); err != nil || en {
		t.Fatalf("antes de asignar: (%t, %v), se esperaba (false, nil)", en, err)
	}
	if _, ok := b.base.SprintActualDelItem(b.item); ok {
		t.Error("un item sin asignar tiene sprint actual")
	}
	if err := b.base.AsignarItemASprint(ctx, b.item, b.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	if en, err := b.base.ItemEnSprintAbierto(ctx, b.item); err != nil || !en {
		t.Errorf("despues de asignar: (%t, %v), se esperaba (true, nil)", en, err)
	}
	for _, id := range []string{idSinDuenoAsignaciones, "no-es-un-uuid"} {
		if en, err := b.base.ItemEnSprintAbierto(ctx, id); err != nil || en {
			t.Errorf("ItemEnSprintAbierto(%q) = (%t, %v), se esperaba (false, nil)", id, en, err)
		}
	}
}

// US-009 / CA-009-2: se pueden asignar varios items al mismo sprint.
func TestBase_AsignarItemASprint_Varios(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaAsignar(t)
	otro := b.itemEn(t, b.proyecto, "Listar proyectos", true, 2)

	for _, id := range []string{b.item, otro} {
		if err := b.base.AsignarItemASprint(ctx, id, b.sprint); err != nil {
			t.Fatalf("AsignarItemASprint(%s): %v", id, err)
		}
	}

	for _, id := range []string{b.item, otro} {
		if actual, ok := b.base.SprintActualDelItem(id); !ok || actual.ID != b.sprint {
			t.Errorf("sprint actual de %s = (%+v, %t), se esperaba el sprint asignado", id, actual, ok)
		}
	}
}

// US-009 / RN-009-1, RN-009-2: ObtenerSprint devuelve el sprint registrado, y un ID que no
// esta o esta mal formado es un sprint inexistente y no un error.
func TestBase_ObtenerSprint(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaAsignar(t)

	s, existe, err := b.base.ObtenerSprint(ctx, b.sprint)
	if err != nil || !existe || s.ID != b.sprint || s.Nombre() != "Sprint 1" {
		t.Errorf("ObtenerSprint = (%+v, %t, %v), se esperaba el Sprint 1", s, existe, err)
	}
	for _, id := range []string{idSinDuenoAsignaciones, "no-es-un-uuid", ""} {
		if _, existe, err := b.base.ObtenerSprint(ctx, id); err != nil || existe {
			t.Errorf("ObtenerSprint(%q) = (existe %t, err %v), se esperaba (false, nil)", id, existe, err)
		}
	}
}

// US-009 / RN-009-6, CL-009-7, CA-009-7: un item ya asignado se rechaza con
// ErrItemYaAsignado, tambien al mismo sprint, y sigue en uno solo.
func TestBase_AsignarItemASprint_YaAsignado(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaAsignar(t)
	if err := b.base.AsignarItemASprint(ctx, b.item, b.sprint); err != nil {
		t.Fatalf("primera asignacion: %v", err)
	}

	err := b.base.AsignarItemASprint(ctx, b.item, b.sprint)

	if !errors.Is(err, sprint.ErrItemYaAsignado) {
		t.Errorf("error = %v, se esperaba ErrItemYaAsignado", err)
	}
}

// US-009 / RN-009-3, RN-009-5, RN-009-7, CL-009-3 a CL-009-5, CL-009-9: igual que en Postgres,
// el repositorio vuelve a verificar las reglas con los datos guardados: proyecto
// distinto, item completado y Definition of Ready, y no asigna nada.
func TestBase_AsignarItemASprint_ReglasDelDominio(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaAsignar(t)
	completado := itemValido(t, "Historia terminada")
	completado.Estado = backlog.EstadoCompletado
	registrado, err := b.base.RegistrarItem(ctx, b.proyecto, completado)
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}

	casos := []struct {
		nombre    string
		item      string
		esperados []error
	}{
		{"de otro proyecto", b.itemEn(t, b.otro, "Del otro", true, 1), []error{sprint.ErrProyectoDistinto}},
		{"sin estimar", b.itemEn(t, b.proyecto, "Sin estimar", false, 1), []error{sprint.ErrItemSinEstimar}},
		{"sin criterios", b.itemEn(t, b.proyecto, "Sin criterios", true, 0), []error{sprint.ErrItemSinCriterios}},
		{"sin estimar y sin criterios", b.itemEn(t, b.proyecto, "Nada", false, 0), []error{sprint.ErrItemSinEstimar, sprint.ErrItemSinCriterios}},
		{"completado", registrado.ID, []error{backlog.ErrItemCompletado}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := b.base.AsignarItemASprint(ctx, c.item, b.sprint)

			for _, e := range c.esperados {
				if !errors.Is(err, e) {
					t.Errorf("error = %v, se esperaba que incluyera %v", err, e)
				}
			}
			if _, ok := b.base.SprintActualDelItem(c.item); ok {
				t.Error("se asigno un item que no cumplia las reglas")
			}
		})
	}
}

// US-009 / RN-009-1, RN-009-2: igual que en Postgres, si el sprint o el item no estan se
// devuelve el error de inexistencia que corresponde (el sprint primero) y no se
// guarda nada.
func TestBase_AsignarItemASprint_LoQueNoExiste(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaAsignar(t)

	casos := []struct {
		nombre   string
		item     string
		sprint   string
		esperado error
	}{
		{"el sprint no existe", b.item, idSinDuenoAsignaciones, app.ErrSprintInexistente},
		{"el item no existe", idSinDuenoAsignaciones, b.sprint, app.ErrItemInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if err := b.base.AsignarItemASprint(ctx, c.item, c.sprint); !errors.Is(err, c.esperado) {
				t.Errorf("error = %v, se esperaba %v", err, c.esperado)
			}
		})
	}
}

// US-009 / CL-009-10, RN-009-11: asignaciones simultaneas del mismo item al mismo sprint:
// termina bien exactamente una y las demas se rechazan con ErrItemYaAsignado.
func TestBase_AsignarItemASprint_Simultaneas(t *testing.T) {
	b := nuevaBaseParaAsignar(t)
	const intentos = 8

	var wg sync.WaitGroup
	errs := make([]error, intentos)
	for i := 0; i < intentos; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = b.base.AsignarItemASprint(context.Background(), b.item, b.sprint)
		}()
	}
	wg.Wait()

	exitosas := 0
	for i, err := range errs {
		switch {
		case err == nil:
			exitosas++
		case errors.Is(err, sprint.ErrItemYaAsignado):
		default:
			t.Errorf("intento %d: error = %v, se esperaba nil o ErrItemYaAsignado", i, err)
		}
	}
	if exitosas != 1 {
		t.Errorf("terminaron bien %d asignaciones, se esperaba exactamente 1", exitosas)
	}
}
