package memoria_test

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// itemValido arma un item valido con el dominio, como lo haria el caso de uso.
func itemValido(t *testing.T, titulo string) backlog.ItemBacklog {
	t.Helper()
	datos := backlog.DatosItem{Titulo: titulo, Prioridad: backlog.PrioridadMust, Criterios: []string{"Se guarda"}}
	item, err := backlog.NuevoItem(datos, time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NuevoItem: %v", err)
	}
	return item
}

// US-005 / CA-005-3, RN-005-10, CL-005-8: cada item recibe un numero
// correlativo dentro de su proyecto, empezando en 1, y un ID propio. La
// numeracion de un proyecto no depende de la de los demas.
func TestBase_RegistrarItem(t *testing.T) {
	ctx := context.Background()
	base := memoria.NuevaBase()
	demo, otro := base.AgregarProyecto(), base.AgregarProyecto()

	altas := []struct {
		proyecto string
		titulo   string
		numero   int
	}{
		{demo, "Alta de proyectos", 1},
		{demo, "Listar proyectos", 2},
		{otro, "Alta de sprints", 1},
	}
	ids := map[string]bool{}
	for _, a := range altas {
		item, err := base.RegistrarItem(ctx, a.proyecto, itemValido(t, a.titulo))
		if err != nil {
			t.Fatalf("RegistrarItem(%q): %v", a.titulo, err)
		}
		if item.Numero != a.numero {
			t.Errorf("%q: numero = %d, se esperaba %d", a.titulo, item.Numero, a.numero)
		}
		if item.ProyectoID != a.proyecto {
			t.Errorf("%q: proyecto = %q, se esperaba %q", a.titulo, item.ProyectoID, a.proyecto)
		}
		if item.ID == "" || ids[item.ID] {
			t.Errorf("%q: ID = %q, se esperaba uno nuevo y no vacio", a.titulo, item.ID)
		}
		ids[item.ID] = true
	}

	var titulos []string
	for _, item := range base.Items(demo) {
		titulos = append(titulos, item.Titulo)
	}
	if esperados := []string{"Alta de proyectos", "Listar proyectos"}; !slices.Equal(titulos, esperados) {
		t.Errorf("items del proyecto = %q, se esperaba %q", titulos, esperados)
	}
}

// US-005 / RN-005-9: registrar en un proyecto que no esta en la base devuelve
// ErrProyectoInexistente y no guarda nada.
func TestBase_RegistrarItem_ProyectoInexistente(t *testing.T) {
	base := memoria.NuevaBase()

	_, err := base.RegistrarItem(context.Background(), "no-existe", itemValido(t, "Alta de proyectos"))
	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Fatalf("error = %v, se esperaba ErrProyectoInexistente", err)
	}
	if n := len(base.Items("no-existe")); n != 0 {
		t.Errorf("quedaron %d items guardados, se esperaba ninguno", n)
	}
}

// US-005 / RN-005-11: altas simultaneas en el mismo proyecto terminan todas bien,
// con numeros distintos y consecutivos. Con -race tambien prueba que la base se
// puede usar desde varias goroutines.
func TestBase_RegistrarItem_Simultaneas(t *testing.T) {
	const altas = 20
	base := memoria.NuevaBase()
	proyecto := base.AgregarProyecto()

	numeros := make([]int, altas)
	errs := make([]error, altas)
	var grupo sync.WaitGroup
	for i := range altas {
		grupo.Add(1)
		go func() {
			defer grupo.Done()
			item, err := base.RegistrarItem(context.Background(), proyecto, itemValido(t, "Item"))
			numeros[i], errs[i] = item.Numero, err
		}()
	}
	grupo.Wait()

	if err := errors.Join(errs...); err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	slices.Sort(numeros)
	for i, numero := range numeros {
		if numero != i+1 {
			t.Fatalf("numeros = %v, se esperaban del 1 al %d sin repetir", numeros, altas)
		}
	}
}
