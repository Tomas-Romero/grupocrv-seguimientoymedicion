package memoria_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
)

// idSinDuenoEnMemoria es un ID con formato de UUID que ninguna fila tiene.
const idSinDuenoEnMemoria = "00000000-0000-0000-0000-000000009999"

// baseConEsfuerzo es una base con dos proyectos: "demo", con una historia y un
// integrante, y "otro", vacio.
type baseConEsfuerzo struct {
	base       *memoria.Base
	demo       string
	otro       string
	item       string
	integrante string
}

func nuevaBaseConEsfuerzo(t *testing.T) baseConEsfuerzo {
	t.Helper()
	ctx := context.Background()
	base := memoria.NuevaBase()
	demo, otro := base.AgregarProyecto(), base.AgregarProyecto()
	item, err := base.RegistrarItem(ctx, demo, itemValido(t, "Alta de proyectos"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	nuevo, err := integrante.Nuevo(demo, "Ana", "ana@ejemplo.test", integrante.RolProductBuilder)
	if err != nil {
		t.Fatalf("integrante.Nuevo: %v", err)
	}
	registrado, err := base.RegistrarIntegrante(ctx, nuevo)
	if err != nil {
		t.Fatalf("RegistrarIntegrante: %v", err)
	}
	return baseConEsfuerzo{base: base, demo: demo, otro: otro, item: item.ID, integrante: registrado.ID}
}

func (b baseConEsfuerzo) registro(t *testing.T, horas float64) esfuerzo.RegistroEsfuerzo {
	t.Helper()
	hoy := time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)
	r, err := esfuerzo.Nuevo(b.demo, b.item, b.integrante, time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC), "Desarrollo", horas, hoy)
	if err != nil {
		t.Fatalf("esfuerzo.Nuevo: %v", err)
	}
	return r
}

// US-019 / CA-019-1, RN-019-10: cada registro recibe un ID propio, como el
// RETURNING de Postgres, y queda guardado con los datos recibidos.
func TestBase_RegistrarEsfuerzo(t *testing.T) {
	b := nuevaBaseConEsfuerzo(t)

	guardado, err := b.base.RegistrarEsfuerzo(context.Background(), b.registro(t, 2.5))
	if err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}

	if guardado.ID == "" {
		t.Fatal("el registro volvio sin ID")
	}
	registros := b.base.Esfuerzos(b.demo)
	if len(registros) != 1 || registros[0].ID != guardado.ID || registros[0].Horas != 2.5 {
		t.Errorf("registros guardados = %+v, se esperaba el registro de 2,5 horas", registros)
	}
}

// US-019 / CL-019-11, CL-019-12, RN-019-9: dos registros del mismo integrante,
// historia y dia son dos registros con IDs distintos, aunque sumen mas de 24 horas.
func TestBase_RegistrarEsfuerzo_Repetidos(t *testing.T) {
	b := nuevaBaseConEsfuerzo(t)
	ctx := context.Background()

	primero, err1 := b.base.RegistrarEsfuerzo(ctx, b.registro(t, 20))
	segundo, err2 := b.base.RegistrarEsfuerzo(ctx, b.registro(t, 20))

	if err1 != nil || err2 != nil {
		t.Fatalf("errores = %v, %v; se esperaba que los dos se guardaran", err1, err2)
	}
	if primero.ID == segundo.ID {
		t.Errorf("los dos registros tienen el mismo ID %s", primero.ID)
	}
	if n := len(b.base.Esfuerzos(b.demo)); n != 2 {
		t.Errorf("hay %d registros, se esperaban 2", n)
	}
}

// US-019 / RN-019-7: se puede registrar esfuerzo sobre una historia completada:
// el repositorio no mira el estado.
func TestBase_RegistrarEsfuerzo_ItemCompletado(t *testing.T) {
	b := nuevaBaseConEsfuerzo(t)
	completado := itemValido(t, "Historia terminada")
	completado.Estado = backlog.EstadoCompletado
	item, err := b.base.RegistrarItem(context.Background(), b.demo, completado)
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	r := b.registro(t, 1)
	r.ItemID = item.ID

	if _, err := b.base.RegistrarEsfuerzo(context.Background(), r); err != nil {
		t.Fatalf("RegistrarEsfuerzo sobre una historia completada: %v", err)
	}
}

// US-019 / RN-019-9: varias cargas simultaneas son registros distintos y ninguna
// falla.
func TestBase_RegistrarEsfuerzo_Simultaneos(t *testing.T) {
	b := nuevaBaseConEsfuerzo(t)
	const cargas = 8

	var wg sync.WaitGroup
	errs := make([]error, cargas)
	for i := 0; i < cargas; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = b.base.RegistrarEsfuerzo(context.Background(), b.registro(t, 1))
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("carga %d: %v", i, err)
		}
	}
	if n := len(b.base.Esfuerzos(b.demo)); n != cargas {
		t.Errorf("hay %d registros, se esperaban %d", n, cargas)
	}
}

// US-019 / RN-019-6, CL-019-15: la historia existe dentro de su proyecto; una de
// otro proyecto, un ID inexistente o uno mal formado no cuentan.
func TestBase_ExisteItemEnProyecto(t *testing.T) {
	b := nuevaBaseConEsfuerzo(t)
	ctx := context.Background()

	casos := []struct {
		nombre   string
		proyecto string
		item     string
		esperado bool
	}{
		{"la historia es del proyecto", b.demo, b.item, true},
		{"la historia es de otro proyecto", b.otro, b.item, false},
		{"la historia no existe", b.demo, idSinDuenoEnMemoria, false},
		{"ID mal formado", b.demo, "no-es-un-uuid", false},
		{"ID vacio", "", "", false},
	}
	for _, c := range casos {
		existe, err := b.base.ExisteItemEnProyecto(ctx, c.proyecto, c.item)
		if err != nil || existe != c.esperado {
			t.Errorf("%s: ExisteItemEnProyecto = (%t, %v), se esperaba (%t, nil)", c.nombre, existe, err, c.esperado)
		}
	}
}

// US-019 / RN-019-6, CL-019-15: lo mismo para el integrante.
func TestBase_ExisteIntegranteEnProyecto(t *testing.T) {
	b := nuevaBaseConEsfuerzo(t)
	ctx := context.Background()

	casos := []struct {
		nombre     string
		proyecto   string
		integrante string
		esperado   bool
	}{
		{"el integrante es del proyecto", b.demo, b.integrante, true},
		{"el integrante es de otro proyecto", b.otro, b.integrante, false},
		{"el integrante no existe", b.demo, idSinDuenoEnMemoria, false},
		{"ID mal formado", b.demo, "no-es-un-uuid", false},
	}
	for _, c := range casos {
		existe, err := b.base.ExisteIntegranteEnProyecto(ctx, c.proyecto, c.integrante)
		if err != nil || existe != c.esperado {
			t.Errorf("%s: ExisteIntegranteEnProyecto = (%t, %v), se esperaba (%t, nil)", c.nombre, existe, err, c.esperado)
		}
	}
}

// US-019 / RN-019-6: igual que la clave foranea de Postgres, registrar con un
// proyecto, historia o integrante que no estan devuelve el error de inexistencia
// que corresponde y no guarda nada.
func TestBase_RegistrarEsfuerzo_LoQueNoExiste(t *testing.T) {
	b := nuevaBaseConEsfuerzo(t)
	base := b.registro(t, 1)

	casos := []struct {
		nombre   string
		cambio   func(r *esfuerzo.RegistroEsfuerzo)
		esperado error
	}{
		{"el proyecto no existe", func(r *esfuerzo.RegistroEsfuerzo) { r.ProyectoID = idSinDuenoEnMemoria }, app.ErrProyectoInexistente},
		{"la historia no existe", func(r *esfuerzo.RegistroEsfuerzo) { r.ItemID = idSinDuenoEnMemoria }, app.ErrItemInexistente},
		{"el integrante no existe", func(r *esfuerzo.RegistroEsfuerzo) { r.IntegranteID = idSinDuenoEnMemoria }, app.ErrIntegranteInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r := base
			c.cambio(&r)

			guardado, err := b.base.RegistrarEsfuerzo(context.Background(), r)

			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.esperado)
			}
			if guardado.ID != "" {
				t.Errorf("devolvio el registro %+v, se esperaba el valor cero", guardado)
			}
			if n := len(b.base.Esfuerzos(b.demo)); n != 0 {
				t.Errorf("hay %d registros, se esperaba ninguno", n)
			}
		})
	}
}

// Esfuerzos devuelve una copia: quien la recibe no puede cambiar lo guardado.
func TestBase_Esfuerzos_DevuelveUnaCopia(t *testing.T) {
	b := nuevaBaseConEsfuerzo(t)
	if _, err := b.base.RegistrarEsfuerzo(context.Background(), b.registro(t, 2)); err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}

	copia := b.base.Esfuerzos(b.demo)
	copia[0].Horas = 99

	if original := b.base.Esfuerzos(b.demo); original[0].Horas != 2 {
		t.Errorf("cambiar la copia modifico lo guardado: horas = %v", original[0].Horas)
	}
	if vacio := b.base.Esfuerzos(b.otro); len(vacio) != 0 {
		t.Errorf("un proyecto sin registros devolvio %v", vacio)
	}
}
