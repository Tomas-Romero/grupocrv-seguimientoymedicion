package defecto_test

import (
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

// US-023 / CL-023-1 / RN-023-2 / RN-023-5: las transiciones permitidas dejan el defecto con el
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
