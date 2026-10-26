package defecto_test

import (
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// US-022 / CA-022-1 / RN-022-3: un defecto nace abierto y sin sprint de resolucion, y
// conserva los IDs y los datos recibidos. Todavia no tiene ID: lo genera la base.
func TestNuevo_NaceAbiertoYSinResolver(t *testing.T) {
	d, err := defecto.Nuevo("proyecto-1", "item-1", "sprint-1", "Falla al guardar", defecto.SeveridadAlta)
	if err != nil {
		t.Fatalf("Nuevo() devolvio %v, quiero nil", err)
	}

	esperado := defecto.Defecto{
		ProyectoID:        "proyecto-1",
		ItemID:            "item-1",
		SprintDeteccionID: "sprint-1",
		Descripcion:       "Falla al guardar",
		Severidad:         defecto.SeveridadAlta,
		Estado:            defecto.EstadoAbierto,
	}
	if d != esperado {
		t.Errorf("Nuevo() = %+v, quiero %+v", d, esperado)
	}
}

// US-022 / CA-022-2 / CL-022-14: las cuatro severidades se aceptan y todas nacen
// abiertas y sin sprint de resolucion, incluida la critica.
func TestNuevo_CuatroSeveridades(t *testing.T) {
	casos := []struct {
		nombre    string
		severidad defecto.Severidad
	}{
		{"critica", defecto.SeveridadCritica},
		{"alta", defecto.SeveridadAlta},
		{"media", defecto.SeveridadMedia},
		{"baja", defecto.SeveridadBaja},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := defecto.Nuevo("p", "i", "s", "Falla", c.severidad)
			if err != nil {
				t.Fatalf("Nuevo() devolvio %v, quiero nil", err)
			}
			if d.Severidad != c.severidad || d.Estado != defecto.EstadoAbierto || d.SprintResolucionID != "" {
				t.Errorf("Nuevo() = %+v, quiero severidad %q, abierto y sin sprint de resolucion", d, c.severidad)
			}
		})
	}
}
