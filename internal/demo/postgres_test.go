package demo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/demo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// CA-T023-6: la misma carga sobre una base Postgres real (con DATABASE_URL; sin ella la prueba se omite, como las demas de Postgres).
// Cubre la carga, la comprobacion, la negativa a cargar dos veces, --borrar sin tocar otro proyecto y cargar de nuevo con los mismos datos.
func TestPostgres_CargarBorrarYCargarDeNuevo(t *testing.T) {
	ctx := context.Background()
	pool := dbprueba.BaseMigrada(ctx, t)
	reloj := demo.NuevoReloj()
	casos, lecturas := demo.CasosPostgres(pool, reloj)

	// Otro proyecto en la base: la carga se hace al lado y el borrado no lo toca (CL-T023-2 y RN-T023-5).
	otro, err := casos.CrearProyecto.Ejecutar(ctx, "Proyecto de prueba", "otro", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("CrearProyecto: %v", err)
	}

	r, err := demo.Cargar(ctx, casos, reloj)
	if err != nil {
		t.Fatalf("Cargar: %v", err)
	}
	if got := r.Resumen.String(); got != resumenEsperado {
		t.Errorf("resumen:\n got  %s\n want %s", got, resumenEsperado)
	}
	if err := demo.Verificar(ctx, lecturas, r); err != nil {
		t.Fatalf("Verificar: %v", err)
	}

	if _, err := demo.Cargar(ctx, casos, reloj); !errors.Is(err, demo.ErrYaExiste) {
		t.Fatalf("la segunda carga dio %v, se esperaba ErrYaExiste", err)
	}
	if n := dbprueba.ContarFilas(ctx, t, pool, "proyectos"); n != 2 {
		t.Fatalf("hay %d proyectos despues de la segunda carga, se esperaban 2", n)
	}

	borrados, err := demo.BorrarProyecto(ctx, pool)
	if err != nil || borrados != 1 {
		t.Fatalf("BorrarProyecto = %d, %v; se esperaba 1", borrados, err)
	}
	// El borrado en cascada no deja filas huerfanas (CL-T023-5): el otro proyecto esta vacio, asi que todas las tablas quedan en cero.
	for _, tabla := range []string{"integrantes", "items_backlog", "sprints", "sprint_items", "registros_esfuerzo", "sesiones_planning_poker", "defectos"} {
		if n := dbprueba.ContarFilas(ctx, t, pool, tabla); n != 0 {
			t.Errorf("la tabla %s quedo con %d filas despues de borrar", tabla, n)
		}
	}
	if n := dbprueba.ContarFilas(ctx, t, pool, "proyectos"); n != 1 {
		t.Fatalf("hay %d proyectos despues de borrar, se esperaba 1 (el otro)", n)
	}
	if _, err := lecturas.ListarBacklog.Ejecutar(ctx, otro.ID, app.FiltroBacklog{}); err != nil {
		t.Errorf("el otro proyecto dejo de existir: %v", err)
	}

	// Cargar de nuevo deja los mismos datos que una carga limpia.
	r2, err := demo.Cargar(ctx, casos, reloj)
	if err != nil {
		t.Fatalf("Cargar de nuevo: %v", err)
	}
	if err := demo.Verificar(ctx, lecturas, r2); err != nil {
		t.Errorf("Verificar despues de cargar de nuevo: %v", err)
	}
	if borrados, err := demo.BorrarProyecto(ctx, pool); err != nil || borrados != 1 {
		t.Errorf("segundo BorrarProyecto = %d, %v", borrados, err)
	}
	if borrados, err := demo.BorrarProyecto(ctx, pool); err != nil || borrados != 0 {
		t.Errorf("borrar sin que exista dio %d, %v; se esperaba 0 sin error (CL-T023-4)", borrados, err)
	}
}
