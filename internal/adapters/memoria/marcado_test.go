package memoria_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// Fechas de completado de estos tests.
var (
	marcadoEn  = time.Date(2026, 10, 8, 15, 0, 0, 0, time.UTC)
	segundaVez = time.Date(2026, 10, 9, 9, 0, 0, 0, time.UTC)
)

func completar(en time.Time) app.CambioDeEstado {
	return func(item backlog.ItemBacklog) backlog.ItemBacklog { return item.Completar(en) }
}

// asignado devuelve la base con el item listo ya asignado al sprint abierto.
func asignado(t *testing.T) baseParaAsignar {
	t.Helper()
	b := nuevaBaseParaAsignar(t)
	if err := b.base.AsignarItemASprint(context.Background(), b.item, b.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	return b
}

// sprintCerradoEn registra en el proyecto un sprint ya cerrado (cerrar es US-011) y
// devuelve su ID.
func sprintCerradoEn(t *testing.T, b baseParaAsignar, proyecto string) string {
	t.Helper()
	s := sprintParaProyecto(t, proyecto, diaDeSprint(10, 6), diaDeSprint(10, 19))
	s.Estado = sprint.EstadoCerrado
	creado, err := b.base.RegistrarSprint(context.Background(), s)
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}
	return creado.ID
}

// US-010 / RN-010-2, CL-010-5: el item esta en el sprint si tiene una fila pendiente en
// el. No lo esta si no se asigno, si se pregunta por otro sprint o si los IDs no estan.
func TestBase_ItemEnSprint(t *testing.T) {
	ctx := context.Background()
	b := asignado(t)
	sinAsignar := b.itemEn(t, b.proyecto, "Listar proyectos", true, 1)
	otroSprint := sprintCerradoEn(t, b, b.otro)

	casos := []struct {
		nombre           string
		itemID, sprintID string
		esperado         bool
	}{
		{"asignado", b.item, b.sprint, true},
		{"sin asignar", sinAsignar, b.sprint, false},
		{"otro sprint", b.item, otroSprint, false},
		{"item inexistente", idSinDuenoAsignaciones, b.sprint, false},
		{"sprint inexistente", b.item, idSinDuenoAsignaciones, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			en, err := b.base.ItemEnSprint(ctx, c.itemID, c.sprintID)
			if err != nil || en != c.esperado {
				t.Errorf("ItemEnSprint = (%t, %v), se esperaba (%t, nil)", en, err, c.esperado)
			}
		})
	}
}

// US-010 / CA-010-1, RN-010-5: guardar el marcado deja el item completado con la fecha,
// lo devuelve como quedo y lo deja en el sprint: la fila sigue pendiente (el resultado
// lo pone US-011).
func TestBase_GuardarCambioDeEstado_Marca(t *testing.T) {
	ctx := context.Background()
	b := asignado(t)

	item, err := b.base.GuardarCambioDeEstado(ctx, b.item, b.sprint, completar(marcadoEn))
	if err != nil {
		t.Fatalf("GuardarCambioDeEstado: %v", err)
	}

	guardado, _, _ := b.base.ObtenerItem(ctx, b.item)
	for nombre, i := range map[string]backlog.ItemBacklog{"devuelto": item, "guardado": guardado} {
		if i.Estado != backlog.EstadoCompletado || !i.CompletadoEn.Equal(marcadoEn) {
			t.Errorf("%s = (%q, %v), se esperaba (completado, %v)", nombre, i.Estado, i.CompletadoEn, marcadoEn)
		}
	}
	if en, _ := b.base.ItemEnSprint(ctx, b.item, b.sprint); !en {
		t.Error("marcar saco al item del sprint: la fila tiene que seguir pendiente")
	}
}

// US-010 / CA-010-3, RN-010-6, CL-010-10: el cambio se aplica al item guardado, no al
// que leyo el caso de uso. Un segundo marcado ve el item ya completado y conserva la
// fecha del primero.
func TestBase_GuardarCambioDeEstado_SegundoMarcadoConservaLaFecha(t *testing.T) {
	ctx := context.Background()
	b := asignado(t)
	if _, err := b.base.GuardarCambioDeEstado(ctx, b.item, b.sprint, completar(marcadoEn)); err != nil {
		t.Fatalf("primer marcado: %v", err)
	}

	item, err := b.base.GuardarCambioDeEstado(ctx, b.item, b.sprint, completar(segundaVez))
	if err != nil {
		t.Fatalf("segundo marcado: %v", err)
	}
	if !item.CompletadoEn.Equal(marcadoEn) {
		t.Errorf("completado en = %v, se esperaba la del primer marcado %v", item.CompletadoEn, marcadoEn)
	}
}

// US-010 / CA-010-2, RN-010-7: desmarcar deja el item pendiente y sin fecha.
func TestBase_GuardarCambioDeEstado_Desmarca(t *testing.T) {
	ctx := context.Background()
	b := asignado(t)
	if _, err := b.base.GuardarCambioDeEstado(ctx, b.item, b.sprint, completar(marcadoEn)); err != nil {
		t.Fatalf("marcar: %v", err)
	}

	if _, err := b.base.GuardarCambioDeEstado(ctx, b.item, b.sprint, backlog.ItemBacklog.Reabrir); err != nil {
		t.Fatalf("desmarcar: %v", err)
	}

	guardado, _, _ := b.base.ObtenerItem(ctx, b.item)
	if guardado.Estado != backlog.EstadoPendiente || !guardado.CompletadoEn.IsZero() {
		t.Errorf("guardado = (%q, %v), se esperaba (pendiente, sin fecha)", guardado.Estado, guardado.CompletadoEn)
	}
}

// US-010 / RN-010-9, CL-010-5, CL-010-7, CL-010-9: el repositorio vuelve a verificar
// las reglas con lo guardado. Un item fuera del sprint se rechaza con
// ErrItemFueraDelSprint; un sprint cerrado, solo con ErrSprintCerrado. En los dos casos
// no cambia nada.
func TestBase_GuardarCambioDeEstado_Reglas(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaAsignar(t)
	cerrado := sprintCerradoEn(t, b, b.otro)
	casos := []struct {
		nombre      string
		sprintID    string
		errEsperado error
		noEsperado  error
	}{
		{"item fuera del sprint", b.sprint, sprint.ErrItemFueraDelSprint, sprint.ErrSprintCerrado},
		{"sprint cerrado", cerrado, sprint.ErrSprintCerrado, sprint.ErrItemFueraDelSprint},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := b.base.GuardarCambioDeEstado(ctx, b.item, c.sprintID, completar(marcadoEn))

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if errors.Is(err, c.noEsperado) {
				t.Errorf("el error %q no deberia incluir %v", err, c.noEsperado)
			}
			if guardado, _, _ := b.base.ObtenerItem(ctx, b.item); guardado.Estado != backlog.EstadoPendiente {
				t.Errorf("se cambio el item: %+v", guardado)
			}
		})
	}
}

// US-010 / CA-010-6, RN-010-9: si el sprint o el item no estan, se devuelve
// ErrSprintInexistente o ErrItemInexistente. El sprint se verifica primero, como en la
// transaccion de Postgres.
func TestBase_GuardarCambioDeEstado_Inexistentes(t *testing.T) {
	ctx := context.Background()
	b := asignado(t)
	casos := []struct {
		nombre           string
		itemID, sprintID string
		errEsperado      error
	}{
		{"sprint inexistente", b.item, idSinDuenoAsignaciones, app.ErrSprintInexistente},
		{"item inexistente", idSinDuenoAsignaciones, b.sprint, app.ErrItemInexistente},
		{"los dos", idSinDuenoAsignaciones, idSinDuenoAsignaciones, app.ErrSprintInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := b.base.GuardarCambioDeEstado(ctx, c.itemID, c.sprintID, completar(marcadoEn))
			if !errors.Is(err, c.errEsperado) {
				t.Errorf("error = %v, se esperaba %v", err, c.errEsperado)
			}
		})
	}
	if guardado, _, _ := b.base.ObtenerItem(ctx, b.item); guardado.Estado != backlog.EstadoPendiente {
		t.Errorf("se cambio el item: %+v", guardado)
	}
}
