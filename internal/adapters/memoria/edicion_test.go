package memoria_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// editado devuelve el item con los cuatro datos cambiados, como lo deja
// ItemBacklog.Editar.
func editado(t *testing.T, item backlog.ItemBacklog) backlog.ItemBacklog {
	t.Helper()
	nuevo, err := item.Editar(backlog.DatosItem{
		Titulo:      "Alta de proyectos con integrantes",
		Descripcion: "Con su equipo",
		Prioridad:   backlog.PrioridadCould,
		Criterios:   []string{"Se guardan los integrantes", "Se listan"},
	})
	if err != nil {
		t.Fatalf("Editar: %v", err)
	}
	return nuevo
}

// US-006 / CA-006-1, RN-006-2, RN-006-4, RN-006-8: guardar la edicion cambia
// solo el titulo, la descripcion, la prioridad y los criterios del item pedido.
// Aunque el item recibido traiga otro estado o estimacion, esos datos no se
// escriben; el otro item queda igual.
func TestBase_GuardarEdicion(t *testing.T) {
	ctx := context.Background()
	base, primero, segundo := baseConDosItems(t)
	puntos, _ := backlog.NuevosStoryPoints(5)
	if err := base.GuardarEstimacion(ctx, primero.ID, puntos); err != nil {
		t.Fatalf("GuardarEstimacion: %v", err)
	}
	nuevo := editado(t, primero)
	nuevo.Estado = backlog.EstadoEnProgreso
	nuevo.StoryPoints = backlog.SinEstimar()
	nuevo.Numero = 99

	if err := base.GuardarEdicion(ctx, nuevo); err != nil {
		t.Fatalf("GuardarEdicion: %v", err)
	}

	guardado, _, _ := base.ObtenerItem(ctx, primero.ID)
	if guardado.Titulo != nuevo.Titulo || guardado.Descripcion != nuevo.Descripcion ||
		guardado.Prioridad != nuevo.Prioridad || !slices.Equal(guardado.Criterios, nuevo.Criterios) {
		t.Errorf("guardado = %+v, se esperaban los datos editados", guardado)
	}
	if guardado.Estado != backlog.EstadoPendiente || guardado.StoryPoints != puntos || guardado.Numero != primero.Numero {
		t.Errorf("se escribio el estado, la estimacion o el numero: %+v", guardado)
	}
	if otro, _, _ := base.ObtenerItem(ctx, segundo.ID); otro.Titulo != segundo.Titulo {
		t.Errorf("la edicion se guardo tambien en otro item: %+v", otro)
	}
	if items := base.Items(primero.ProyectoID); len(items) != 1 {
		t.Errorf("el proyecto tiene %d items, se esperaba 1", len(items))
	}
}

// US-006 / CA-006-4, CL-006-2: una lista de criterios vacia reemplaza a la
// anterior.
func TestBase_GuardarEdicion_CriteriosVacios(t *testing.T) {
	ctx := context.Background()
	base, primero, _ := baseConDosItems(t)
	nuevo, err := primero.Editar(backlog.DatosItem{Titulo: primero.Titulo, Prioridad: primero.Prioridad})
	if err != nil {
		t.Fatalf("Editar: %v", err)
	}

	if err := base.GuardarEdicion(ctx, nuevo); err != nil {
		t.Fatalf("GuardarEdicion: %v", err)
	}

	if guardado, _, _ := base.ObtenerItem(ctx, primero.ID); len(guardado.Criterios) != 0 {
		t.Errorf("criterios = %q, se esperaba la lista vacia", guardado.Criterios)
	}
}

// GuardarEdicion guarda una copia: cambiar despues el item recibido no modifica
// lo guardado.
func TestBase_GuardarEdicion_GuardaUnaCopia(t *testing.T) {
	ctx := context.Background()
	base, primero, _ := baseConDosItems(t)
	nuevo := editado(t, primero)

	if err := base.GuardarEdicion(ctx, nuevo); err != nil {
		t.Fatalf("GuardarEdicion: %v", err)
	}
	nuevo.Criterios[0] = "cambiado"

	if guardado, _, _ := base.ObtenerItem(ctx, primero.ID); guardado.Criterios[0] == "cambiado" {
		t.Error("cambiar el item recibido modifico lo guardado")
	}
}

// US-006 / CA-006-7: guardar la edicion de un item que no esta devuelve
// ErrItemInexistente y no cambia nada.
func TestBase_GuardarEdicion_ItemInexistente(t *testing.T) {
	ctx := context.Background()
	base, primero, _ := baseConDosItems(t)
	nuevo := editado(t, primero)
	nuevo.ID = "00000000-0000-0000-0000-000000009999"

	err := base.GuardarEdicion(ctx, nuevo)

	if !errors.Is(err, app.ErrItemInexistente) {
		t.Fatalf("error = %v, se esperaba ErrItemInexistente", err)
	}
	if guardado, _, _ := base.ObtenerItem(ctx, primero.ID); guardado.Titulo != primero.Titulo {
		t.Errorf("se cambio un item existente: %+v", guardado)
	}
}

// US-006 / RN-006-8, CL-006-7: guardar la edicion de un item que ya esta
// completado devuelve backlog.ErrItemCompletado y no lo cambia. Cubre la ventana
// entre leer el item y guardarlo: si se completo en el medio, no se edita.
func TestBase_GuardarEdicion_ItemCompletado(t *testing.T) {
	ctx := context.Background()
	base := memoria.NuevaBase()
	completado := itemValido(t, "Alta de proyectos")
	completado.Estado = backlog.EstadoCompletado
	registrado, err := base.RegistrarItem(ctx, base.AgregarProyecto(), completado)
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	// La edicion se arma sobre el item como lo leyo el caso de uso, cuando
	// todavia estaba pendiente; en la base ya esta completado.
	leido := registrado
	leido.Estado = backlog.EstadoPendiente
	nuevo := editado(t, leido)

	err = base.GuardarEdicion(ctx, nuevo)

	if !errors.Is(err, backlog.ErrItemCompletado) {
		t.Fatalf("error = %v, se esperaba ErrItemCompletado", err)
	}
	if guardado, _, _ := base.ObtenerItem(ctx, registrado.ID); guardado.Titulo != registrado.Titulo {
		t.Errorf("se edito un item completado: %+v", guardado)
	}
}
