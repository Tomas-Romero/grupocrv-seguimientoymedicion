package defecto_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// defectoEn arma un defecto guardado en el estado y con el sprint de resolucion indicados,
// detectado en el sprint "sprint-deteccion". Los IDs son fijos para comparar el defecto entero.
func defectoEn(estado defecto.Estado, sprintResolucionID string) defecto.Defecto {
	return defecto.Defecto{
		ID:                 "defecto-1",
		ProyectoID:         "proyecto-1",
		ItemID:             "item-1",
		SprintDeteccionID:  "sprint-deteccion",
		Descripcion:        "Falla al guardar",
		Severidad:          defecto.SeveridadAlta,
		Estado:             estado,
		SprintResolucionID: sprintResolucionID,
	}
}

// US-023 / CA-023-2 / CA-023-3 / CL-023-1 / RN-023-2 / RN-023-5: las transiciones permitidas dejan el defecto con el
// estado pedido, con el sprint de resolucion solo si es resuelto, y sin tocar nada mas. Reabrir un
// resuelto deja el sprint vacio y no conserva el anterior.
func TestCambiarEstado_TransicionesPermitidas(t *testing.T) {
	sprint2 := defecto.SprintDeResolucion{ID: "sprint-2", Numero: 2}
	casos := []struct {
		nombre         string
		actual         defecto.Defecto
		nuevo          defecto.Estado
		resolucion     defecto.SprintDeResolucion
		sprintEsperado string
	}{
		{"abierto a en_correccion", defectoEn(defecto.EstadoAbierto, ""), defecto.EstadoEnCorreccion, defecto.SprintDeResolucion{}, ""},
		{"abierto a descartado", defectoEn(defecto.EstadoAbierto, ""), defecto.EstadoDescartado, defecto.SprintDeResolucion{}, ""},
		{"abierto a resuelto", defectoEn(defecto.EstadoAbierto, ""), defecto.EstadoResuelto, sprint2, "sprint-2"},
		{"en_correccion a abierto", defectoEn(defecto.EstadoEnCorreccion, ""), defecto.EstadoAbierto, defecto.SprintDeResolucion{}, ""},
		{"en_correccion a descartado", defectoEn(defecto.EstadoEnCorreccion, ""), defecto.EstadoDescartado, defecto.SprintDeResolucion{}, ""},
		{"en_correccion a resuelto", defectoEn(defecto.EstadoEnCorreccion, ""), defecto.EstadoResuelto, sprint2, "sprint-2"},
		{"resuelto a abierto (reapertura)", defectoEn(defecto.EstadoResuelto, "sprint-2"), defecto.EstadoAbierto, defecto.SprintDeResolucion{}, ""},
		{"descartado a abierto (reapertura)", defectoEn(defecto.EstadoDescartado, ""), defecto.EstadoAbierto, defecto.SprintDeResolucion{}, ""},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := c.actual.CambiarEstado(c.nuevo, c.resolucion, 1)
			if err != nil {
				t.Fatalf("CambiarEstado() devolvio %v, quiero nil", err)
			}

			esperado := c.actual
			esperado.Estado = c.nuevo
			esperado.SprintResolucionID = c.sprintEsperado
			if d != esperado {
				t.Errorf("CambiarEstado() = %+v, quiero %+v", d, esperado)
			}
		})
	}
}

// US-023 / CA-023-6 / CL-023-2 / RN-023-2: las transiciones que la tabla no lista, y pedir el estado
// actual (que no es un no-op, ni siquiera con otro sprint), se rechazan con ErrTransicionInvalida
// indicando el estado actual y el pedido. No se devuelve ningun defecto.
func TestCambiarEstado_TransicionesProhibidas(t *testing.T) {
	sprint2 := defecto.SprintDeResolucion{ID: "sprint-2", Numero: 2}
	casos := []struct {
		nombre     string
		actual     defecto.Defecto
		nuevo      defecto.Estado
		resolucion defecto.SprintDeResolucion
		mensaje    string
	}{
		{"resuelto a en_correccion", defectoEn(defecto.EstadoResuelto, "sprint-2"), defecto.EstadoEnCorreccion, defecto.SprintDeResolucion{}, "de resuelto a en_correccion"},
		{"resuelto a descartado", defectoEn(defecto.EstadoResuelto, "sprint-2"), defecto.EstadoDescartado, defecto.SprintDeResolucion{}, "de resuelto a descartado"},
		{"descartado a en_correccion", defectoEn(defecto.EstadoDescartado, ""), defecto.EstadoEnCorreccion, defecto.SprintDeResolucion{}, "de descartado a en_correccion"},
		{"descartado a resuelto", defectoEn(defecto.EstadoDescartado, ""), defecto.EstadoResuelto, sprint2, "de descartado a resuelto"},
		{"abierto a abierto", defectoEn(defecto.EstadoAbierto, ""), defecto.EstadoAbierto, defecto.SprintDeResolucion{}, "de abierto a abierto"},
		{"en_correccion a en_correccion", defectoEn(defecto.EstadoEnCorreccion, ""), defecto.EstadoEnCorreccion, defecto.SprintDeResolucion{}, "de en_correccion a en_correccion"},
		{"resuelto a resuelto con otro sprint", defectoEn(defecto.EstadoResuelto, "sprint-1"), defecto.EstadoResuelto, sprint2, "de resuelto a resuelto"},
		{"descartado a descartado", defectoEn(defecto.EstadoDescartado, ""), defecto.EstadoDescartado, defecto.SprintDeResolucion{}, "de descartado a descartado"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := c.actual.CambiarEstado(c.nuevo, c.resolucion, 1)
			if !errors.Is(err, defecto.ErrTransicionInvalida) {
				t.Fatalf("CambiarEstado() devolvio %v, quiero ErrTransicionInvalida", err)
			}
			if !strings.Contains(err.Error(), c.mensaje) {
				t.Errorf("el error es %q, quiero que indique %q", err, c.mensaje)
			}
			if d != (defecto.Defecto{}) {
				t.Errorf("CambiarEstado() devolvio el defecto %+v, quiero ninguno", d)
			}
		})
	}
}

// US-023 / CA-023-9 / CL-023-3 / RN-023-1: un estado fuera de los cuatro se rechaza con
// ErrEstadoInvalido indicando el valor recibido. El dominio no recorta ni pasa a minuscula, y con
// un estado invalido no se evalua ninguna otra regla: es el unico error, con o sin sprint.
func TestCambiarEstado_EstadoInvalido(t *testing.T) {
	invalidos := []defecto.Estado{"cerrado", "Resuelto", " resuelto ", "en correccion", "en_corrección", ""}
	resoluciones := map[string]defecto.SprintDeResolucion{
		"sin sprint": {},
		"con sprint": {ID: "sprint-2", Numero: 2},
	}
	for _, nuevo := range invalidos {
		for nombre, resolucion := range resoluciones {
			t.Run(fmt.Sprintf("%q %s", nuevo, nombre), func(t *testing.T) {
				d, err := defectoEn(defecto.EstadoAbierto, "").CambiarEstado(nuevo, resolucion, 1)
				if !errors.Is(err, defecto.ErrEstadoInvalido) {
					t.Fatalf("CambiarEstado() devolvio %v, quiero ErrEstadoInvalido", err)
				}
				if errors.Is(err, defecto.ErrTransicionInvalida) || errors.Is(err, defecto.ErrSprintResolucionFaltante) ||
					errors.Is(err, defecto.ErrSprintResolucionNoCorresponde) || errors.Is(err, defecto.ErrSprintResolucionAnterior) {
					t.Errorf("el error es %q, quiero solo ErrEstadoInvalido", err)
				}
				if !strings.Contains(err.Error(), fmt.Sprintf("%q", string(nuevo))) {
					t.Errorf("el error es %q, quiero que indique el valor %q", err, string(nuevo))
				}
				if d != (defecto.Defecto{}) {
					t.Errorf("CambiarEstado() devolvio el defecto %+v, quiero ninguno", d)
				}
			})
		}
	}
}

// US-023 / CA-023-7 / CL-023-4 / RN-023-3: resuelto exige sprint de resolucion
// (ErrSprintResolucionFaltante) y cualquier otro estado lo rechaza
// (ErrSprintResolucionNoCorresponde, indicando el estado pedido): no se descarta en silencio.
func TestCambiarEstado_SprintDeResolucionSegunEstado(t *testing.T) {
	sprint2 := defecto.SprintDeResolucion{ID: "sprint-2", Numero: 2}
	casos := []struct {
		nombre     string
		actual     defecto.Defecto
		nuevo      defecto.Estado
		resolucion defecto.SprintDeResolucion
		esperado   error
		mensaje    string
	}{
		{"abierto a resuelto sin sprint", defectoEn(defecto.EstadoAbierto, ""), defecto.EstadoResuelto, defecto.SprintDeResolucion{}, defecto.ErrSprintResolucionFaltante, ""},
		{"en_correccion a resuelto sin sprint", defectoEn(defecto.EstadoEnCorreccion, ""), defecto.EstadoResuelto, defecto.SprintDeResolucion{}, defecto.ErrSprintResolucionFaltante, ""},
		{"abierto a en_correccion con sprint", defectoEn(defecto.EstadoAbierto, ""), defecto.EstadoEnCorreccion, sprint2, defecto.ErrSprintResolucionNoCorresponde, "en_correccion"},
		{"abierto a descartado con sprint", defectoEn(defecto.EstadoAbierto, ""), defecto.EstadoDescartado, sprint2, defecto.ErrSprintResolucionNoCorresponde, "descartado"},
		{"en_correccion a abierto con sprint", defectoEn(defecto.EstadoEnCorreccion, ""), defecto.EstadoAbierto, sprint2, defecto.ErrSprintResolucionNoCorresponde, "abierto"},
		{"resuelto a abierto con sprint", defectoEn(defecto.EstadoResuelto, "sprint-2"), defecto.EstadoAbierto, sprint2, defecto.ErrSprintResolucionNoCorresponde, "abierto"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := c.actual.CambiarEstado(c.nuevo, c.resolucion, 1)
			if !errors.Is(err, c.esperado) {
				t.Fatalf("CambiarEstado() devolvio %v, quiero %v", err, c.esperado)
			}
			if !strings.Contains(err.Error(), c.mensaje) {
				t.Errorf("el error es %q, quiero que indique %q", err, c.mensaje)
			}
			if d != (defecto.Defecto{}) {
				t.Errorf("CambiarEstado() devolvio el defecto %+v, quiero ninguno", d)
			}
		})
	}
}

// US-023 / CA-023-5 / CA-023-8 / CL-023-5 / RN-023-4: resolver en un sprint anterior al de deteccion
// se rechaza con ErrSprintResolucionAnterior indicando los dos numeros; el mismo sprint o uno
// posterior se acepta, y el borde inferior (detectado y resuelto en el Sprint 1) tambien. Se
// compara por numero: los IDs de los sprints no dicen cual es anterior.
func TestCambiarEstado_SprintDeResolucionNoAnteriorAlDeDeteccion(t *testing.T) {
	casos := []struct {
		nombre     string
		deteccion  int
		resolucion int
		anterior   bool
	}{
		{"resuelto en el Sprint 2, detectado en el 3", 3, 2, true},
		{"resuelto en el Sprint 1, detectado en el 3", 3, 1, true},
		{"resuelto en el mismo sprint (3)", 3, 3, false},
		{"resuelto en el Sprint 4, detectado en el 3", 3, 4, false},
		{"detectado y resuelto en el Sprint 1 (borde inferior)", 1, 1, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			resolucion := defecto.SprintDeResolucion{ID: "sprint-resolucion", Numero: c.resolucion}
			d, err := defectoEn(defecto.EstadoAbierto, "").CambiarEstado(defecto.EstadoResuelto, resolucion, c.deteccion)

			if !c.anterior {
				if err != nil {
					t.Fatalf("CambiarEstado() devolvio %v, quiero nil", err)
				}
				if d.Estado != defecto.EstadoResuelto || d.SprintResolucionID != "sprint-resolucion" {
					t.Errorf("CambiarEstado() = %+v, quiero resuelto con el sprint de resolucion", d)
				}
				return
			}
			if !errors.Is(err, defecto.ErrSprintResolucionAnterior) {
				t.Fatalf("CambiarEstado() devolvio %v, quiero ErrSprintResolucionAnterior", err)
			}
			esperado := fmt.Sprintf("sprint %d", c.resolucion)
			if !strings.Contains(err.Error(), esperado) || !strings.Contains(err.Error(), fmt.Sprintf("sprint %d", c.deteccion)) {
				t.Errorf("el error es %q, quiero que indique los sprints %d y %d", err, c.resolucion, c.deteccion)
			}
			if d != (defecto.Defecto{}) {
				t.Errorf("CambiarEstado() devolvio el defecto %+v, quiero ninguno", d)
			}
		})
	}
}

// US-023 / CA-023-13 / CL-023-9 / RN-023-6: el dominio devuelve todos los errores juntos, unidos con
// errors.Join, en el orden fijo transicion, sprint faltante o que no corresponde, y sprint
// anterior. Nunca hay mas de dos: faltante, no corresponde y anterior se excluyen entre si.
func TestCambiarEstado_UneLosErroresEnOrden(t *testing.T) {
	sprint2 := defecto.SprintDeResolucion{ID: "sprint-2", Numero: 2}
	casos := []struct {
		nombre     string
		actual     defecto.Defecto
		nuevo      defecto.Estado
		resolucion defecto.SprintDeResolucion
		deteccion  int
		esperados  []error
	}{
		{"descartado a resuelto sin sprint", defectoEn(defecto.EstadoDescartado, ""), defecto.EstadoResuelto, defecto.SprintDeResolucion{}, 1,
			[]error{defecto.ErrTransicionInvalida, defecto.ErrSprintResolucionFaltante}},
		{"resuelto a resuelto en un sprint anterior", defectoEn(defecto.EstadoResuelto, "sprint-3"), defecto.EstadoResuelto, sprint2, 3,
			[]error{defecto.ErrTransicionInvalida, defecto.ErrSprintResolucionAnterior}},
		{"resuelto a en_correccion con sprint", defectoEn(defecto.EstadoResuelto, "sprint-3"), defecto.EstadoEnCorreccion, sprint2, 1,
			[]error{defecto.ErrTransicionInvalida, defecto.ErrSprintResolucionNoCorresponde}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := c.actual.CambiarEstado(c.nuevo, c.resolucion, c.deteccion)

			unidos, ok := err.(interface{ Unwrap() []error })
			if !ok {
				t.Fatalf("CambiarEstado() devolvio %v, quiero los errores unidos con errors.Join", err)
			}
			encontrados := unidos.Unwrap()
			if len(encontrados) != len(c.esperados) {
				t.Fatalf("CambiarEstado() unio %d errores (%v), quiero %d", len(encontrados), err, len(c.esperados))
			}
			for i, esperado := range c.esperados {
				if !errors.Is(encontrados[i], esperado) {
					t.Errorf("error %d = %v, quiero %v", i, encontrados[i], esperado)
				}
				if !errors.Is(err, esperado) {
					t.Errorf("errors.Is(err, %v) = false, quiero true", esperado)
				}
			}
			if d != (defecto.Defecto{}) {
				t.Errorf("CambiarEstado() devolvio el defecto %+v, quiero ninguno", d)
			}
		})
	}
}
