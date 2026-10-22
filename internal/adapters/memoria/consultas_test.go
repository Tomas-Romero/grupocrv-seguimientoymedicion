package memoria_test

import (
	"context"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

func registrarProyectoDePrueba(t *testing.T, base *memoria.Base, nombre string) proyecto.Proyecto {
	t.Helper()
	inicio := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	p, err := base.RegistrarProyecto(context.Background(), proyecto.Proyecto{Nombre: nombre, FechaInicio: inicio, FechaFin: inicio.AddDate(0, 1, 0)})
	if err != nil {
		t.Fatalf("RegistrarProyecto: %v", err)
	}
	return p
}

// T-011 / RN-T011-2: ListarProyectos devuelve los proyectos registrados, con sus datos.
func TestBase_ListarProyectos(t *testing.T) {
	base := memoria.NuevaBase()
	ctx := context.Background()

	vacio, err := base.ListarProyectos(ctx)
	if err != nil || len(vacio) != 0 {
		t.Fatalf("base vacia: %v, %v; quiero lista vacia y sin error", vacio, err)
	}

	uno := registrarProyectoDePrueba(t, base, "Uno")
	dos := registrarProyectoDePrueba(t, base, "Dos")
	base.AgregarProyecto() // un proyecto sin datos no se lista

	got, err := base.ListarProyectos(ctx)
	if err != nil {
		t.Fatalf("ListarProyectos: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("hay %d proyectos, quiero 2: %+v", len(got), got)
	}
	porID := map[string]string{}
	for _, p := range got {
		porID[p.ID] = p.Nombre
	}
	if porID[uno.ID] != "Uno" || porID[dos.ID] != "Dos" {
		t.Errorf("proyectos = %v, quiero Uno y Dos con sus IDs", porID)
	}
}

// Lo que se devuelve es una copia: cambiarla no cambia lo guardado.
func TestBase_ListarProyectos_DevuelveUnaCopia(t *testing.T) {
	base := memoria.NuevaBase()
	p := registrarProyectoDePrueba(t, base, "Demo")

	got, _ := base.ListarProyectos(context.Background())
	got[0].Nombre = "cambiado"

	guardado, ok := base.Proyecto(p.ID)
	if !ok || guardado.Nombre != "Demo" {
		t.Errorf("el proyecto guardado cambio: %+v", guardado)
	}
}

// ListarIntegrantes devuelve los del proyecto; en otro proyecto o inexistente, una lista vacia.
func TestBase_ListarIntegrantes(t *testing.T) {
	base := memoria.NuevaBase()
	ctx := context.Background()
	uno := base.AgregarProyecto()
	otro := base.AgregarProyecto()
	if _, err := base.RegistrarIntegrante(ctx, integrante.Integrante{ProyectoID: uno, Nombre: "Ana", Email: "ana@x.test", Rol: integrante.RolProductBuilder}); err != nil {
		t.Fatalf("RegistrarIntegrante: %v", err)
	}

	got, err := base.ListarIntegrantes(ctx, uno)
	if err != nil || len(got) != 1 || got[0].Nombre != "Ana" {
		t.Errorf("ListarIntegrantes(uno) = %+v, %v; quiero a Ana", got, err)
	}
	for _, id := range []string{otro, "no-existe", ""} {
		got, err := base.ListarIntegrantes(ctx, id)
		if err != nil || len(got) != 0 {
			t.Errorf("ListarIntegrantes(%q) = %+v, %v; quiero lista vacia y sin error", id, got, err)
		}
	}

	got, _ = base.ListarIntegrantes(ctx, uno)
	got[0].Nombre = "cambiado"
	if guardados := base.Integrantes(uno); guardados[0].Nombre != "Ana" {
		t.Errorf("cambiar la lista cambio lo guardado: %+v", guardados)
	}
}

// ListarItems devuelve los items del proyecto con su propia lista de criterios.
func TestBase_ListarItems(t *testing.T) {
	base := memoria.NuevaBase()
	ctx := context.Background()
	uno := base.AgregarProyecto()
	otro := base.AgregarProyecto()
	item := backlog.ItemBacklog{Titulo: "Historia", Prioridad: backlog.PrioridadMust, Estado: backlog.EstadoPendiente, Criterios: []string{"c1", "c2"}}
	if _, err := base.RegistrarItem(ctx, uno, item); err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}

	got, err := base.ListarItems(ctx, uno)
	if err != nil || len(got) != 1 || got[0].Titulo != "Historia" || got[0].Numero != 1 {
		t.Fatalf("ListarItems(uno) = %+v, %v; quiero la historia numero 1", got, err)
	}
	for _, id := range []string{otro, "no-existe", ""} {
		got, err := base.ListarItems(ctx, id)
		if err != nil || len(got) != 0 {
			t.Errorf("ListarItems(%q) = %+v, %v; quiero lista vacia y sin error", id, got, err)
		}
	}

	got[0].Criterios[0] = "cambiado"
	if guardados := base.Items(uno); guardados[0].Criterios[0] != "c1" {
		t.Errorf("cambiar los criterios devueltos cambio lo guardado: %v", guardados[0].Criterios)
	}
}
