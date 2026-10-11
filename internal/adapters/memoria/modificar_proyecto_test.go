package memoria_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
)

// US-002 / RN-002-1: BuscarProyecto devuelve los datos de un proyecto registrado.
// Un ID desconocido, o vacio, es un proyecto inexistente.
func TestBase_BuscarProyecto(t *testing.T) {
	ctx := context.Background()
	base := memoria.NuevaBase()

	alfa, err := base.RegistrarProyecto(ctx, proyectoValido("Alfa"))
	if err != nil {
		t.Fatalf("RegistrarProyecto: %v", err)
	}

	encontrado, err := base.BuscarProyecto(ctx, alfa.ID)
	if err != nil {
		t.Fatalf("BuscarProyecto(%q): %v", alfa.ID, err)
	}
	if encontrado != alfa {
		t.Errorf("BuscarProyecto(%q) = %+v, se esperaba %+v", alfa.ID, encontrado, alfa)
	}

	for _, id := range []string{"no-existe", ""} {
		if _, err := base.BuscarProyecto(ctx, id); !errors.Is(err, app.ErrProyectoInexistente) {
			t.Errorf("BuscarProyecto(%q): error = %v, se esperaba ErrProyectoInexistente", id, err)
		}
	}
}

// US-002 / RN-002-3, RN-002-4, RN-002-7: ActualizarProyecto reemplaza los datos
// del proyecto y conserva su ID; no toca a los demas proyectos ni lo que cuelga
// de este (sus items de backlog).
func TestBase_ActualizarProyecto(t *testing.T) {
	ctx := context.Background()
	base := memoria.NuevaBase()

	alfa, err := base.RegistrarProyecto(ctx, proyectoValido("Alfa"))
	if err != nil {
		t.Fatalf("RegistrarProyecto(Alfa): %v", err)
	}
	beta, err := base.RegistrarProyecto(ctx, proyectoValido("Beta"))
	if err != nil {
		t.Fatalf("RegistrarProyecto(Beta): %v", err)
	}
	if _, err := base.RegistrarItem(ctx, alfa.ID, itemValido(t, "Alta de proyectos")); err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}

	cambiado := alfa
	cambiado.Nombre = "Alfa v2"
	cambiado.Descripcion = ""
	cambiado.FechaInicio = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	cambiado.FechaFin = time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC)

	if err := base.ActualizarProyecto(ctx, cambiado); err != nil {
		t.Fatalf("ActualizarProyecto: %v", err)
	}

	guardado, ok := base.Proyecto(alfa.ID)
	if !ok || guardado != cambiado {
		t.Errorf("Proyecto(%q) = %+v, %t; se esperaba %+v", alfa.ID, guardado, ok, cambiado)
	}
	if intacto, _ := base.Proyecto(beta.ID); intacto != beta {
		t.Errorf("Beta cambio al modificar Alfa: %+v, se esperaba %+v", intacto, beta)
	}
	if existe, err := base.ExisteProyecto(ctx, alfa.ID); err != nil || !existe {
		t.Errorf("ExisteProyecto(%q) = %t, %v; se esperaba true", alfa.ID, existe, err)
	}
	if n := len(base.Items(alfa.ID)); n != 1 {
		t.Errorf("el proyecto tiene %d items, se esperaba 1: modificar no toca lo asociado", n)
	}
}

// US-002 / RN-002-1: actualizar un proyecto que no esta en la base devuelve
// ErrProyectoInexistente y no guarda nada.
func TestBase_ActualizarProyecto_Inexistente(t *testing.T) {
	base := memoria.NuevaBase()
	fantasma := proyectoValido("Fantasma")
	fantasma.ID = "no-existe"

	err := base.ActualizarProyecto(context.Background(), fantasma)
	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Fatalf("error = %v, se esperaba ErrProyectoInexistente", err)
	}
	if _, ok := base.Proyecto("no-existe"); ok {
		t.Error("quedo guardado un proyecto que no existia")
	}
}