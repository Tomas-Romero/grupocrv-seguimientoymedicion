package memoria_test

import (
	"context"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

func proyectoValido(nombre string) proyecto.Proyecto {
	return proyecto.Proyecto{
		Nombre:      nombre,
		Descripcion: "descripcion de " + nombre,
		FechaInicio: time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC),
	}
}

// US-001 / RN-001-6, CA-001-2: la base asigna un ID nuevo a cada proyecto
// (como el RETURNING de Postgres), guarda sus datos y un segundo proyecto no
// modifica el primero.
func TestBase_RegistrarProyecto(t *testing.T) {
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

	if alfa.ID == "" || beta.ID == "" || alfa.ID == beta.ID {
		t.Fatalf("IDs = %q y %q, se esperaban dos distintos y no vacios", alfa.ID, beta.ID)
	}
	guardado, ok := base.Proyecto(alfa.ID)
	if !ok || guardado != alfa {
		t.Errorf("Proyecto(%q) = %+v, %t; se esperaba %+v", alfa.ID, guardado, ok, alfa)
	}
	if _, ok := base.Proyecto("no-existe"); ok {
		t.Error("Proyecto(\"no-existe\") encontro algo, se esperaba que no")
	}
}

// US-001: un proyecto registrado es visible para el resto de los casos de uso.
// ExisteProyecto lo encuentra y el backlog (US-005) puede registrarle items.
func TestBase_RegistrarProyecto_VisibleParaBacklog(t *testing.T) {
	ctx := context.Background()
	base := memoria.NuevaBase()

	p, err := base.RegistrarProyecto(ctx, proyectoValido("Demo"))
	if err != nil {
		t.Fatalf("RegistrarProyecto: %v", err)
	}

	existe, err := base.ExisteProyecto(ctx, p.ID)
	if err != nil || !existe {
		t.Fatalf("ExisteProyecto(%q) = %t, %v; se esperaba true", p.ID, existe, err)
	}
	item, err := base.RegistrarItem(ctx, p.ID, itemValido(t, "Alta de proyectos"))
	if err != nil {
		t.Fatalf("RegistrarItem en el proyecto registrado: %v", err)
	}
	if item.Numero != 1 {
		t.Errorf("numero = %d, se esperaba 1", item.Numero)
	}
}