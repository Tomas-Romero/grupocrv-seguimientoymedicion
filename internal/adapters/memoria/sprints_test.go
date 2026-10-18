package memoria_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

func diaDeSprint(mes time.Month, d int) time.Time {
	return time.Date(2026, mes, d, 0, 0, 0, 0, time.UTC)
}

// proyectoParaSprints registra en la base un proyecto del 2026-10-06 al 2026-12-31 con
// RegistrarProyecto, que es el que guarda las fechas.
func proyectoParaSprints(t *testing.T, base *memoria.Base, nombre string) string {
	t.Helper()
	nuevo, err := proyecto.Nuevo(nombre, "", diaDeSprint(10, 6), diaDeSprint(12, 31))
	if err != nil {
		t.Fatalf("proyecto.Nuevo: %v", err)
	}
	registrado, err := base.RegistrarProyecto(context.Background(), nuevo)
	if err != nil {
		t.Fatalf("RegistrarProyecto: %v", err)
	}
	return registrado.ID
}

func sprintParaProyecto(t *testing.T, proyectoID string, inicio, fin time.Time) sprint.Sprint {
	t.Helper()
	s, err := sprint.Nuevo(proyectoID, "Objetivo", inicio, fin, diaDeSprint(10, 6), diaDeSprint(12, 31))
	if err != nil {
		t.Fatalf("sprint.Nuevo: %v", err)
	}
	return s
}

// US-008 / RN-008-9: se leen las fechas de un proyecto registrado, y cualquier ID
// que no este (o este mal formado) es un proyecto inexistente y no un error.
func TestBase_FechasDelProyecto(t *testing.T) {
	ctx := context.Background()
	base := memoria.NuevaBase()
	id := proyectoParaSprints(t, base, "Demo")

	inicio, fin, existe, err := base.FechasDelProyecto(ctx, id)
	if err != nil || !existe || !inicio.Equal(diaDeSprint(10, 6)) || !fin.Equal(diaDeSprint(12, 31)) {
		t.Errorf("FechasDelProyecto = (%v, %v, %t, %v), se esperaban 2026-10-06 a 2026-12-31", inicio, fin, existe, err)
	}
	for _, otro := range []string{"00000000-0000-0000-0000-000000009999", "no-es-un-uuid", ""} {
		if _, _, existe, err := base.FechasDelProyecto(ctx, otro); err != nil || existe {
			t.Errorf("FechasDelProyecto(%q) = (existe %t, err %v), se esperaba (false, nil)", otro, existe, err)
		}
	}
}

// US-008 / CA-008-1, RN-008-8, RN-008-11, RN-008-16: el primer sprint de un proyecto
// es el 1, queda abierto y recibe un ID propio, como el RETURNING de Postgres.
func TestBase_RegistrarSprint(t *testing.T) {
	base := memoria.NuevaBase()
	id := proyectoParaSprints(t, base, "Demo")

	registrado, err := base.RegistrarSprint(context.Background(), sprintParaProyecto(t, id, diaDeSprint(10, 6), diaDeSprint(10, 19)))
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}

	if registrado.ID == "" || registrado.Numero != 1 || registrado.Nombre() != "Sprint 1" || registrado.Estado != sprint.EstadoAbierto {
		t.Errorf("sprint = %+v, se esperaba el Sprint 1 abierto con ID", registrado)
	}
	if sprints := base.Sprints(id); len(sprints) != 1 || sprints[0].ID != registrado.ID {
		t.Errorf("sprints guardados = %+v, se esperaba el registrado", sprints)
	}
}

// US-008 / CA-008-2, CL-008-8, RN-008-11: la numeracion es por proyecto: el primer
// sprint de otro proyecto es el 1 aunque el primero tenga un sprint abierto en las
// mismas fechas.
func TestBase_RegistrarSprint_NumeracionPorProyecto(t *testing.T) {
	ctx := context.Background()
	base := memoria.NuevaBase()
	demo, otro := proyectoParaSprints(t, base, "Demo"), proyectoParaSprints(t, base, "Otro")

	if _, err := base.RegistrarSprint(ctx, sprintParaProyecto(t, demo, diaDeSprint(10, 6), diaDeSprint(10, 19))); err != nil {
		t.Fatalf("primer sprint de Demo: %v", err)
	}
	delOtro, err := base.RegistrarSprint(ctx, sprintParaProyecto(t, otro, diaDeSprint(10, 6), diaDeSprint(10, 19)))
	if err != nil {
		t.Fatalf("primer sprint de Otro: %v", err)
	}

	if delOtro.Numero != 1 {
		t.Errorf("primer sprint de Otro = %d, se esperaba 1", delOtro.Numero)
	}
}

// US-008 / CA-008-11, CL-008-17, RN-008-17: con el ultimo sprint abierto se rechaza
// solo con ErrUltimoSprintAbierto, empiece el nuevo despues o antes de que termine,
// y no se guarda nada.
func TestBase_RegistrarSprint_UltimoAbierto(t *testing.T) {
	ctx := context.Background()
	base := memoria.NuevaBase()
	id := proyectoParaSprints(t, base, "Demo")
	if _, err := base.RegistrarSprint(ctx, sprintParaProyecto(t, id, diaDeSprint(10, 6), diaDeSprint(10, 19))); err != nil {
		t.Fatalf("primer sprint: %v", err)
	}

	for _, inicio := range []time.Time{diaDeSprint(10, 20), diaDeSprint(10, 10)} {
		_, err := base.RegistrarSprint(ctx, sprintParaProyecto(t, id, inicio, diaDeSprint(11, 2)))
		if !errors.Is(err, sprint.ErrUltimoSprintAbierto) {
			t.Errorf("inicio %s: error = %v, se esperaba ErrUltimoSprintAbierto", inicio.Format(time.DateOnly), err)
		}
	}
	if n := len(base.Sprints(id)); n != 1 {
		t.Errorf("hay %d sprints, se esperaba 1", n)
	}
}

// US-008 / CL-008-18, RN-008-7: igual que en Postgres, el repositorio vuelve a
// verificar las fechas contra las del proyecto guardadas: un sprint fuera de ellas
// se rechaza solo con ErrFueraDelProyecto.
func TestBase_RegistrarSprint_FueraDelProyecto(t *testing.T) {
	base := memoria.NuevaBase()
	id := proyectoParaSprints(t, base, "Demo")
	// El sprint es valido para un proyecto que empieza antes, pero no para este.
	afuera, err := sprint.Nuevo(id, "Objetivo", diaDeSprint(10, 1), diaDeSprint(10, 10), diaDeSprint(9, 1), diaDeSprint(12, 31))
	if err != nil {
		t.Fatalf("sprint.Nuevo: %v", err)
	}

	_, err = base.RegistrarSprint(context.Background(), afuera)

	if !errors.Is(err, sprint.ErrFueraDelProyecto) {
		t.Errorf("error = %v, se esperaba ErrFueraDelProyecto", err)
	}
	if n := len(base.Sprints(id)); n != 0 {
		t.Errorf("hay %d sprints, se esperaba ninguno", n)
	}
}

// US-008 / CA-008-10, RN-008-9: un proyecto que no esta en la base se rechaza con
// ErrProyectoInexistente y no se guarda nada.
func TestBase_RegistrarSprint_ProyectoInexistente(t *testing.T) {
	base := memoria.NuevaBase()

	_, err := base.RegistrarSprint(context.Background(), sprintParaProyecto(t, "00000000-0000-0000-0000-000000009999", diaDeSprint(10, 6), diaDeSprint(10, 19)))

	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Errorf("error = %v, se esperaba ErrProyectoInexistente", err)
	}
}

// US-008 / CL-008-15, RN-008-14: altas simultaneas en el mismo proyecto: termina bien
// exactamente una y las demas se rechazan con ErrUltimoSprintAbierto.
func TestBase_RegistrarSprint_Simultaneos(t *testing.T) {
	base := memoria.NuevaBase()
	id := proyectoParaSprints(t, base, "Demo")
	const altas = 8

	var wg sync.WaitGroup
	errs := make([]error, altas)
	for i := 0; i < altas; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = base.RegistrarSprint(context.Background(), sprintParaProyecto(t, id, diaDeSprint(10, 6), diaDeSprint(10, 19)))
		}()
	}
	wg.Wait()

	exitosas := 0
	for i, err := range errs {
		switch {
		case err == nil:
			exitosas++
		case errors.Is(err, sprint.ErrUltimoSprintAbierto):
		default:
			t.Errorf("alta %d: error = %v, se esperaba nil o ErrUltimoSprintAbierto", i, err)
		}
	}
	if exitosas != 1 {
		t.Errorf("terminaron bien %d altas, se esperaba exactamente 1", exitosas)
	}
}

// Sprints devuelve una copia en orden de alta: quien la recibe no puede cambiar lo
// guardado, y un proyecto sin sprints devuelve una lista vacia.
func TestBase_Sprints_DevuelveUnaCopia(t *testing.T) {
	base := memoria.NuevaBase()
	id := proyectoParaSprints(t, base, "Demo")
	if _, err := base.RegistrarSprint(context.Background(), sprintParaProyecto(t, id, diaDeSprint(10, 6), diaDeSprint(10, 19))); err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}

	copia := base.Sprints(id)
	copia[0].SprintGoal = "cambiado"

	if original := base.Sprints(id); original[0].SprintGoal != "Objetivo" {
		t.Errorf("cambiar la copia modifico lo guardado: %q", original[0].SprintGoal)
	}
	if vacio := base.Sprints("00000000-0000-0000-0000-000000009999"); len(vacio) != 0 {
		t.Errorf("un proyecto sin sprints devolvio %v", vacio)
	}
}
