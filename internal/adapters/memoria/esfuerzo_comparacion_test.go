package memoria_test

import (
	"context"
	"reflect"
	"testing"
)

// US-021 / sección 2: cada historia del proyecto llega con las horas de todos sus registros, y las
// que no tienen horas llegan sin ellas. Nada de otro proyecto.
func TestBase_LeerHistoriasConHoras(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseConEsfuerzo(t)
	if _, err := b.base.RegistrarEsfuerzo(ctx, b.registro(t, 2.5)); err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}
	if _, err := b.base.RegistrarEsfuerzo(ctx, b.registro(t, 1)); err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}
	sinHoras, err := b.base.RegistrarItem(ctx, b.demo, itemValido(t, "Listar proyectos"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	if _, err := b.base.RegistrarItem(ctx, b.otro, itemValido(t, "Historia ajena")); err != nil {
		t.Fatalf("RegistrarItem ajeno: %v", err)
	}

	historias, err := b.base.LeerHistoriasConHoras(ctx, b.demo)
	if err != nil {
		t.Fatalf("LeerHistoriasConHoras: %v", err)
	}

	if len(historias) != 2 {
		t.Fatalf("historias = %d, se esperaban las 2 del proyecto", len(historias))
	}
	for _, h := range historias {
		switch h.Item.ID {
		case b.item:
			if !reflect.DeepEqual(h.Horas, []float64{2.5, 1}) {
				t.Errorf("horas de %q = %v, se esperaba [2.5 1] en orden de alta", h.Item.Titulo, h.Horas)
			}
		case sinHoras.ID:
			if len(h.Horas) != 0 {
				t.Errorf("horas de %q = %v, se esperaba ninguna", h.Item.Titulo, h.Horas)
			}
		default:
			t.Errorf("historia inesperada %q: es de otro proyecto", h.Item.Titulo)
		}
	}
}

// US-021 / CL-021-13: un proyecto sin historias, o que no existe, da una lista vacia.
func TestBase_LeerHistoriasConHoras_SinHistorias(t *testing.T) {
	b := nuevaBaseConEsfuerzo(t)

	for nombre, proyecto := range map[string]string{"proyecto vacio": b.otro, "proyecto inexistente": idSinDuenoEnMemoria} {
		historias, err := b.base.LeerHistoriasConHoras(context.Background(), proyecto)
		if err != nil || len(historias) != 0 {
			t.Errorf("%s: historias = %v, error = %v; se esperaba una lista vacia", nombre, historias, err)
		}
	}
}

// US-021 / RN-021-11: lo que devuelve es una copia; modificarla no cambia la base.
func TestBase_LeerHistoriasConHoras_DevuelveCopias(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseConEsfuerzo(t)
	if _, err := b.base.RegistrarEsfuerzo(ctx, b.registro(t, 2)); err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}

	historias, _ := b.base.LeerHistoriasConHoras(ctx, b.demo)
	historias[0].Horas[0] = 99

	otra, _ := b.base.LeerHistoriasConHoras(ctx, b.demo)
	if otra[0].Horas[0] != 2 {
		t.Errorf("horas = %v, se esperaba 2: la lectura no devuelve una copia", otra[0].Horas)
	}
}
