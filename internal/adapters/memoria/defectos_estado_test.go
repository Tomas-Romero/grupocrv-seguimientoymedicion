package memoria_test

import (
	"context"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// segundoSprint cierra el Sprint 1 del proyecto "Demo" y suma el Sprint 2 (que empieza despues de
// que termina el 1, RN-008-13). Devuelve su ID.
func (b baseConDefectos) segundoSprint(t *testing.T) string {
	t.Helper()
	if _, err := b.base.GuardarCierre(context.Background(), b.sprint); err != nil {
		t.Fatalf("GuardarCierre: %v", err)
	}
	s, err := b.base.RegistrarSprint(context.Background(), sprintParaProyecto(t, b.demo, diaDeSprint(10, 20), diaDeSprint(11, 2)))
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}
	return s.ID
}

// registrado guarda un defecto abierto de "Demo" y lo devuelve con su ID.
func (b baseConDefectos) registrado(t *testing.T) defecto.Defecto {
	t.Helper()
	d, err := b.base.RegistrarDefecto(context.Background(), b.defecto(t, "Falla al guardar"))
	if err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}
	return d
}

// US-023 / RN-023-7 / CL-023-10: el defecto se busca DENTRO del proyecto: uno de otro proyecto,
// un ID que no esta o uno vacio no cuentan. Devuelve el defecto guardado, de donde sale su estado
// actual (RN-023-8).
func TestBase_DefectoEnProyecto(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseConDefectos(t)
	guardado := b.registrado(t)

	casos := []struct {
		nombre   string
		proyecto string
		id       string
		existe   bool
	}{
		{"de su proyecto", b.demo, guardado.ID, true},
		{"de otro proyecto", b.otro, guardado.ID, false},
		{"un ID que no esta", b.demo, "no-existe", false},
		{"un ID vacio", b.demo, "", false},
		{"un proyecto que no esta", "no-existe", guardado.ID, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, existe, err := b.base.DefectoEnProyecto(ctx, c.proyecto, c.id)
			if err != nil {
				t.Fatalf("DefectoEnProyecto: %v", err)
			}
			if existe != c.existe {
				t.Fatalf("existe = %v, quiero %v", existe, c.existe)
			}
			if c.existe && d != guardado {
				t.Errorf("defecto = %+v, quiero %+v", d, guardado)
			}
			if !c.existe && d != (defecto.Defecto{}) {
				t.Errorf("defecto = %+v, quiero el valor cero", d)
			}
		})
	}
}

// US-023 / RN-023-7 / RN-023-8 / CL-023-17: el sprint se busca DENTRO del proyecto y se traduce a su
// numero (sprint.Numero), que es lo que compara el dominio. Uno de otro proyecto, un ID que no
// esta o uno vacio no cuentan.
func TestBase_NumeroDeSprintEnProyecto(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseConDefectos(t)
	sprint2 := b.segundoSprint(t)

	casos := []struct {
		nombre   string
		proyecto string
		id       string
		numero   int
		existe   bool
	}{
		{"el Sprint 1", b.demo, b.sprint, 1, true},
		{"el Sprint 2", b.demo, sprint2, 2, true},
		{"un sprint de otro proyecto", b.demo, b.sprintOtro, 0, false},
		{"un ID que no esta", b.demo, "no-existe", 0, false},
		{"un ID vacio", b.demo, "", 0, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			numero, existe, err := b.base.NumeroDeSprintEnProyecto(ctx, c.proyecto, c.id)
			if err != nil {
				t.Fatalf("NumeroDeSprintEnProyecto: %v", err)
			}
			if numero != c.numero || existe != c.existe {
				t.Errorf("numero, existe = %d, %v; quiero %d, %v", numero, existe, c.numero, c.existe)
			}
		})
	}
}

// US-023 / CA-023-1 / CL-023-12 / RN-023-9 / RN-023-11: guardar el cambio actualiza el estado y el
// sprint de resolucion y nada mas: la descripcion, la severidad, la historia, el sprint de deteccion
// y el ID quedan como estaban, y los demas defectos no se tocan.
func TestBase_ActualizarEstadoDefecto_Guarda(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseConDefectos(t)
	sprint2 := b.segundoSprint(t)
	guardado := b.registrado(t)
	otro := b.registrado(t)

	resuelto, err := guardado.CambiarEstado(defecto.EstadoResuelto, defecto.SprintDeResolucion{ID: sprint2, Numero: 2}, 1)
	if err != nil {
		t.Fatalf("CambiarEstado: %v", err)
	}
	if err := b.base.ActualizarEstadoDefecto(ctx, resuelto, defecto.EstadoAbierto); err != nil {
		t.Fatalf("ActualizarEstadoDefecto: %v", err)
	}

	defectos := b.base.Defectos(b.demo)
	if len(defectos) != 2 || defectos[0] != resuelto || defectos[1] != otro {
		t.Errorf("defectos = %+v, quiero el primero resuelto en el Sprint 2 y el segundo igual", defectos)
	}

	// Reabrir deja el sprint vacio y no conserva el anterior (RN-023-5, CL-023-7).
	reabierto, err := resuelto.CambiarEstado(defecto.EstadoAbierto, defecto.SprintDeResolucion{}, 1)
	if err != nil {
		t.Fatalf("CambiarEstado (reapertura): %v", err)
	}
	if err := b.base.ActualizarEstadoDefecto(ctx, reabierto, defecto.EstadoResuelto); err != nil {
		t.Fatalf("ActualizarEstadoDefecto (reapertura): %v", err)
	}
	if d, _, _ := b.base.DefectoEnProyecto(ctx, b.demo, guardado.ID); d != reabierto {
		t.Errorf("defecto = %+v, quiero %+v (abierto y sin sprint de resolucion)", d, reabierto)
	}
}
