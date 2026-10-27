package memoria_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
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

// US-023 / CA-023-14 / CL-023-14 / RN-023-11: el guardado esta condicionado al estado que leyo el
// caso de uso. Si el defecto ya no esta en ese estado (otro cambio se adelanto) se devuelve
// ErrDefectoModificado sin guardar nada: no queda "descartado" con el sprint del otro cambio.
func TestBase_ActualizarEstadoDefecto_RechazaElEstadoQueYaNoEsElGuardado(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseConDefectos(t)
	sprint2 := b.segundoSprint(t)
	guardado := b.registrado(t)

	// El primer cambio gana: lo descarta.
	descartado, err := guardado.CambiarEstado(defecto.EstadoDescartado, defecto.SprintDeResolucion{}, 1)
	if err != nil {
		t.Fatalf("CambiarEstado: %v", err)
	}
	if err := b.base.ActualizarEstadoDefecto(ctx, descartado, defecto.EstadoAbierto); err != nil {
		t.Fatalf("ActualizarEstadoDefecto del primero: %v", err)
	}

	// El segundo leyo "abierto" antes del primero y lo quiere resolver en el Sprint 2.
	resuelto, err := guardado.CambiarEstado(defecto.EstadoResuelto, defecto.SprintDeResolucion{ID: sprint2, Numero: 2}, 1)
	if err != nil {
		t.Fatalf("CambiarEstado: %v", err)
	}
	err = b.base.ActualizarEstadoDefecto(ctx, resuelto, defecto.EstadoAbierto)

	if !errors.Is(err, app.ErrDefectoModificado) {
		t.Fatalf("error = %v, quiero ErrDefectoModificado", err)
	}
	if d, _, _ := b.base.DefectoEnProyecto(ctx, b.demo, guardado.ID); d != descartado {
		t.Errorf("defecto = %+v, quiero el que dejo el primero: %+v", d, descartado)
	}
}

// US-023 / CL-023-15 / RN-023-11: el defecto que no esta en el proyecto (se borro, o es de otro) se
// rechaza con ErrDefectoInexistente, y un sprint de resolucion que desaparecio, con
// ErrSprintInexistente (la clave foranea de Postgres). En los dos casos no se guarda nada.
func TestBase_ActualizarEstadoDefecto_RechazaLoQueNoEsta(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseConDefectos(t)
	guardado := b.registrado(t)

	resuelto := guardado
	resuelto.Estado = defecto.EstadoResuelto
	resuelto.SprintResolucionID = b.sprint

	casos := []struct {
		nombre   string
		cambio   func(d defecto.Defecto) defecto.Defecto
		esperado error
	}{
		{"el defecto no esta", func(d defecto.Defecto) defecto.Defecto { d.ID = "no-existe"; return d }, app.ErrDefectoInexistente},
		{"el defecto es de otro proyecto", func(d defecto.Defecto) defecto.Defecto { d.ProyectoID = b.otro; return d }, app.ErrDefectoInexistente},
		{"el sprint de resolucion no esta", func(d defecto.Defecto) defecto.Defecto { d.SprintResolucionID = "no-existe"; return d }, app.ErrSprintInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := b.base.ActualizarEstadoDefecto(ctx, c.cambio(resuelto), defecto.EstadoAbierto)

			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, quiero %v", err, c.esperado)
			}
			if d, _, _ := b.base.DefectoEnProyecto(ctx, b.demo, guardado.ID); d != guardado {
				t.Errorf("defecto = %+v, quiero que no cambie: %+v", d, guardado)
			}
		})
	}
}
