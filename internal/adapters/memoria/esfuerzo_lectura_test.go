package memoria_test

import (
	"context"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// US-020 / sección 2: la lectura trae los registros, los titulos, los nombres y los sprints del
// proyecto, y nada de otro proyecto.
func TestBase_LeerEsfuerzo(t *testing.T) {
	ctx := context.Background()
	base := memoria.NuevaBase()
	p, err := base.RegistrarProyecto(ctx, proyecto.Proyecto{Nombre: "Demo",
		FechaInicio: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), FechaFin: time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("RegistrarProyecto: %v", err)
	}
	item, err := base.RegistrarItem(ctx, p.ID, itemValido(t, "Alta de proyectos"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	ana, err := integrante.Nuevo(p.ID, "Ana", "ana@ejemplo.test", integrante.RolProductBuilder)
	if err != nil {
		t.Fatalf("integrante.Nuevo: %v", err)
	}
	ana, err = base.RegistrarIntegrante(ctx, ana)
	if err != nil {
		t.Fatalf("RegistrarIntegrante: %v", err)
	}
	hoy := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	nuevo, err := esfuerzo.Nuevo(p.ID, item.ID, ana.ID, time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), "Desarrollo", 2.5, hoy)
	if err != nil {
		t.Fatalf("esfuerzo.Nuevo: %v", err)
	}
	registrado, err := base.RegistrarEsfuerzo(ctx, nuevo)
	if err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}
	s, err := sprint.Nuevo(p.ID, "Objetivo", time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC), time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC),
		p.FechaInicio, p.FechaFin)
	if err != nil {
		t.Fatalf("sprint.Nuevo: %v", err)
	}
	guardado, err := base.RegistrarSprint(ctx, s)
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}

	l, err := base.LeerEsfuerzo(ctx, p.ID)
	if err != nil {
		t.Fatalf("LeerEsfuerzo: %v", err)
	}

	if len(l.Registros) != 1 || l.Registros[0].ID != registrado.ID || l.Registros[0].Horas != 2.5 {
		t.Errorf("registros = %+v, se esperaba el registro de 2,5 horas", l.Registros)
	}
	if l.Historias[item.ID] != "Alta de proyectos" || len(l.Historias) != 1 {
		t.Errorf("historias = %v, se esperaba solo %q", l.Historias, "Alta de proyectos")
	}
	if l.Integrantes[ana.ID] != "Ana" || len(l.Integrantes) != 1 {
		t.Errorf("integrantes = %v, se esperaba solo Ana", l.Integrantes)
	}
	if len(l.Sprints) != 1 || l.Sprints[0].ID != guardado.ID || l.Sprints[0].Numero != 1 {
		t.Errorf("sprints = %+v, se esperaba el Sprint 1", l.Sprints)
	}
}

// US-020 / CL-020-1: un proyecto de otro (o sin nada) no ve los datos del primero.
func TestBase_LeerEsfuerzo_NoMezclaProyectos(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseConEsfuerzo(t)
	if _, err := b.base.RegistrarEsfuerzo(ctx, b.registro(t, 1)); err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}

	l, err := b.base.LeerEsfuerzo(ctx, b.otro)
	if err != nil {
		t.Fatalf("LeerEsfuerzo: %v", err)
	}

	if len(l.Registros) != 0 || len(l.Historias) != 0 || len(l.Integrantes) != 0 || len(l.Sprints) != 0 {
		t.Errorf("lectura = %+v, se esperaba vacia para el otro proyecto", l)
	}
}

// US-020 / RN-020-12: lo que devuelve la lectura es una copia; modificarla no cambia la base.
func TestBase_LeerEsfuerzo_DevuelveCopias(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseConEsfuerzo(t)
	if _, err := b.base.RegistrarEsfuerzo(ctx, b.registro(t, 1)); err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}

	l, _ := b.base.LeerEsfuerzo(ctx, b.demo)
	l.Registros[0].Horas = 99
	delete(l.Historias, b.item)

	otra, _ := b.base.LeerEsfuerzo(ctx, b.demo)
	if otra.Registros[0].Horas != 1 || otra.Historias[b.item] == "" {
		t.Errorf("lectura = %+v, se esperaba intacta", otra)
	}
}
