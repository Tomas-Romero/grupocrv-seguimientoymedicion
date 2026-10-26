package postgres_test

import (
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
)

// US-020 / sección 2: la lectura trae los registros, los titulos, los nombres y los sprints del
// proyecto, y nada de otro proyecto.
func TestLeerEsfuerzo(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)
	registrado, err := e.repo.RegistrarEsfuerzo(ctx, e.registro(t, ayerDelEsfuerzo, "Desarrollo", 2.5))
	if err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}
	// Un registro de otro proyecto no tiene que aparecer.
	itemAjeno, err := postgres.NuevoRepositorioBacklog(e.pool).RegistrarItem(ctx, e.otroProyecto, itemConTitulo(t, "Historia ajena"))
	if err != nil {
		t.Fatalf("RegistrarItem ajeno: %v", err)
	}
	integranteAjeno := integranteParaEsfuerzo(ctx, t, e.pool, e.otroProyecto, "beto@ejemplo.test")
	ajeno, err := esfuerzo.Nuevo(e.otroProyecto, itemAjeno.ID, integranteAjeno, ayerDelEsfuerzo, "Ajeno", 3, ayerDelEsfuerzo.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("esfuerzo.Nuevo: %v", err)
	}
	if _, err := e.repo.RegistrarEsfuerzo(ctx, ajeno); err != nil {
		t.Fatalf("RegistrarEsfuerzo ajeno: %v", err)
	}
	s := sprintNuevo(t, e.proyecto, diaSprint(time.October, 6), diaSprint(time.October, 19))
	guardado, err := postgres.NuevoRepositorioSprints(e.pool).RegistrarSprint(ctx, s)
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}

	l, err := e.repo.LeerEsfuerzo(ctx, e.proyecto)
	if err != nil {
		t.Fatalf("LeerEsfuerzo: %v", err)
	}

	if len(l.Registros) != 1 || l.Registros[0].ID != registrado.ID || l.Registros[0].Horas != 2.5 ||
		!l.Registros[0].Fecha.Equal(ayerDelEsfuerzo) || l.Registros[0].Actividad != "Desarrollo" ||
		l.Registros[0].ItemID != e.item || l.Registros[0].IntegranteID != e.integrante || l.Registros[0].ProyectoID != e.proyecto {
		t.Errorf("registros = %+v, se esperaba el registro de 2,5 horas del proyecto", l.Registros)
	}
	if l.Historias[e.item] != "Crear proyecto" || len(l.Historias) != 1 {
		t.Errorf("historias = %v, se esperaba solo %q", l.Historias, "Crear proyecto")
	}
	if l.Integrantes[e.integrante] != "Ana" || len(l.Integrantes) != 1 {
		t.Errorf("integrantes = %v, se esperaba solo Ana", l.Integrantes)
	}
	if len(l.Sprints) != 1 || l.Sprints[0].ID != guardado.ID || l.Sprints[0].Numero != 1 ||
		!l.Sprints[0].FechaInicio.Equal(diaSprint(time.October, 6)) || !l.Sprints[0].FechaFin.Equal(diaSprint(time.October, 19)) {
		t.Errorf("sprints = %+v, se esperaba el Sprint 1 del 6 al 19 de octubre", l.Sprints)
	}
}

// US-020 / CL-020-1: un proyecto sin nada, y uno con un ID mal formado, dan una lectura vacia.
func TestLeerEsfuerzo_SinNada(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)

	for nombre, proyecto := range map[string]string{"proyecto vacio": e.otroProyecto, "id mal formado": "no-es-un-uuid", "id sin dueno": idSinDueno} {
		l, err := e.repo.LeerEsfuerzo(ctx, proyecto)
		if err != nil {
			t.Errorf("%s: LeerEsfuerzo: %v", nombre, err)
			continue
		}
		if len(l.Registros) != 0 || len(l.Historias) != 0 || len(l.Integrantes) != 0 || len(l.Sprints) != 0 {
			t.Errorf("%s: lectura = %+v, se esperaba vacia", nombre, l)
		}
	}
}
