package memoria_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// baseConDosItems devuelve una base con dos items de proyectos distintos y los
// items ya registrados, con su ID.
func baseConDosItems(t *testing.T) (*memoria.Base, backlog.ItemBacklog, backlog.ItemBacklog) {
	t.Helper()
	ctx := context.Background()
	base := memoria.NuevaBase()
	primero, err := base.RegistrarItem(ctx, base.AgregarProyecto(), itemValido(t, "Alta de proyectos"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	segundo, err := base.RegistrarItem(ctx, base.AgregarProyecto(), itemValido(t, "Alta de sprints"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	return base, primero, segundo
}

// US-013 / RN-013-6: se puede cargar un item por su ID aunque se desconozca el
// proyecto, y un ID que no esta en la base (incluido uno mal formado) es un
// item inexistente y no un error.
func TestBase_ObtenerItem(t *testing.T) {
	ctx := context.Background()
	base, primero, segundo := baseConDosItems(t)

	for _, esperado := range []backlog.ItemBacklog{primero, segundo} {
		item, existe, err := base.ObtenerItem(ctx, esperado.ID)
		if err != nil || !existe {
			t.Fatalf("ObtenerItem(%s) = (existe %t, err %v), se esperaba encontrarlo", esperado.ID, existe, err)
		}
		if item.ID != esperado.ID || item.Titulo != esperado.Titulo || item.ProyectoID != esperado.ProyectoID {
			t.Errorf("ObtenerItem(%s) devolvio %+v, se esperaba %+v", esperado.ID, item, esperado)
		}
	}

	for _, id := range []string{"00000000-0000-0000-0000-000000009999", "no-es-un-uuid", ""} {
		if _, existe, err := base.ObtenerItem(ctx, id); err != nil || existe {
			t.Errorf("ObtenerItem(%q) = (existe %t, err %v), se esperaba (false, nil)", id, existe, err)
		}
	}
}

// ObtenerItem devuelve una copia: quien la recibe no puede cambiar lo guardado.
func TestBase_ObtenerItem_DevuelveUnaCopia(t *testing.T) {
	ctx := context.Background()
	base, primero, _ := baseConDosItems(t)

	copia, _, _ := base.ObtenerItem(ctx, primero.ID)
	copia.Criterios[0] = "cambiado"

	original, _, _ := base.ObtenerItem(ctx, primero.ID)
	if original.Criterios[0] == "cambiado" {
		t.Error("cambiar la copia modifico lo guardado")
	}
}

// US-013 / RN-013-7: guardar la estimacion cambia solo los Story Points del
// item pedido; el otro item y el resto de los datos quedan igual.
func TestBase_GuardarEstimacion(t *testing.T) {
	ctx := context.Background()
	base, primero, segundo := baseConDosItems(t)
	puntos, err := backlog.NuevosStoryPoints(8)
	if err != nil {
		t.Fatalf("NuevosStoryPoints: %v", err)
	}

	if err := base.GuardarEstimacion(ctx, primero.ID, puntos); err != nil {
		t.Fatalf("GuardarEstimacion: %v", err)
	}

	guardado, _, _ := base.ObtenerItem(ctx, primero.ID)
	if valor, estimado := guardado.StoryPoints.Valor(); valor != 8 || !estimado {
		t.Errorf("story points guardados = (%d, %t), se esperaba (8, true)", valor, estimado)
	}
	if guardado.Titulo != primero.Titulo || guardado.Numero != primero.Numero || guardado.Estado != primero.Estado {
		t.Errorf("se cambio algo mas que los story points: %+v", guardado)
	}
	otro, _, _ := base.ObtenerItem(ctx, segundo.ID)
	if _, estimado := otro.StoryPoints.Valor(); estimado {
		t.Error("la estimacion se guardo tambien en otro item")
	}
	if items := base.Items(primero.ProyectoID); len(items) != 1 {
		t.Errorf("el proyecto tiene %d items, se esperaba 1", len(items))
	}
}

// US-013 / CA-013-6: guardar la estimacion de un item que no esta devuelve
// ErrItemInexistente y no cambia nada.
func TestBase_GuardarEstimacion_ItemInexistente(t *testing.T) {
	ctx := context.Background()
	base, primero, _ := baseConDosItems(t)
	puntos, _ := backlog.NuevosStoryPoints(5)

	err := base.GuardarEstimacion(ctx, "00000000-0000-0000-0000-000000009999", puntos)

	if !errors.Is(err, app.ErrItemInexistente) {
		t.Fatalf("error = %v, se esperaba ErrItemInexistente", err)
	}
	if guardado, _, _ := base.ObtenerItem(ctx, primero.ID); guardado.StoryPoints != backlog.SinEstimar() {
		t.Errorf("se cambio un item existente: %+v", guardado.StoryPoints)
	}
}
