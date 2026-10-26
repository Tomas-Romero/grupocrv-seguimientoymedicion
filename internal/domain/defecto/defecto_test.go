package defecto_test

import (
	"errors"
	"strings"
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

// US-022 / CA-022-6 / CL-022-1 / CL-022-4 / RN-022-1: el limite es de 2000 caracteres
// contados en runas. 2000 se acepta, 2001 se rechaza, y 2000 letras ñ (4000 bytes) se
// aceptan porque no se cuentan bytes.
func TestNuevo_LargoDeLaDescripcion(t *testing.T) {
	casos := []struct {
		nombre      string
		descripcion string
		quiereErr   error
	}{
		{"2000 caracteres se aceptan", strings.Repeat("a", 2000), nil},
		{"2001 caracteres se rechazan", strings.Repeat("a", 2001), defecto.ErrDescripcionMuyLarga},
		{"2000 ñ son 4000 bytes y se aceptan", strings.Repeat("ñ", 2000), nil},
		{"2001 ñ se rechazan", strings.Repeat("ñ", 2001), defecto.ErrDescripcionMuyLarga},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := defecto.Nuevo("p", "i", "s", c.descripcion, defecto.SeveridadAlta)
			if !errors.Is(err, c.quiereErr) {
				t.Errorf("Nuevo() = %v, quiero %v", err, c.quiereErr)
			}
		})
	}
}

// US-022 / CA-022-7 / CL-022-3 / RN-022-1: una descripcion vacia o de solo espacios es
// ErrDescripcionVacia, y nunca ErrDescripcionMuyLarga aunque tenga 3000 espacios.
func TestNuevo_DescripcionVacia(t *testing.T) {
	casos := []struct {
		nombre      string
		descripcion string
	}{
		{"vacia", ""},
		{"solo espacios", "   "},
		{"espacios, tabulaciones y saltos de linea", " \t\n "},
		{"3000 espacios", strings.Repeat(" ", 3000)},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := defecto.Nuevo("p", "i", "s", c.descripcion, defecto.SeveridadAlta)
			if !errors.Is(err, defecto.ErrDescripcionVacia) {
				t.Errorf("Nuevo() = %v, quiero ErrDescripcionVacia", err)
			}
			if errors.Is(err, defecto.ErrDescripcionMuyLarga) {
				t.Errorf("Nuevo() = %v, no quiero ErrDescripcionMuyLarga a la vez", err)
			}
			if d != (defecto.Defecto{}) {
				t.Errorf("Nuevo() devolvio %+v con error, quiero el defecto vacio", d)
			}
		})
	}
}

// US-022 / CL-022-2 / RN-022-1: se recortan los extremos y se guarda recortada, pero los
// saltos de linea del medio se conservan.
func TestNuevo_RecortaLaDescripcion(t *testing.T) {
	casos := []struct {
		nombre      string
		descripcion string
		esperada    string
	}{
		{"espacios a los costados", "  Falla al guardar  ", "Falla al guardar"},
		{"tabulaciones y saltos en los extremos", "\t\nFalla\n ", "Falla"},
		{"saltos del medio se conservan", "Esperado: guardar\nObservado: error 500", "Esperado: guardar\nObservado: error 500"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := defecto.Nuevo("p", "i", "s", c.descripcion, defecto.SeveridadAlta)
			if err != nil {
				t.Fatalf("Nuevo() devolvio %v, quiero nil", err)
			}
			if d.Descripcion != c.esperada {
				t.Errorf("Descripcion = %q, quiero %q", d.Descripcion, c.esperada)
			}
		})
	}
}

// US-022 / CA-022-6: el largo se mide despues de recortar, asi que 2000 caracteres
// rodeados de espacios se aceptan.
func TestNuevo_ElLargoSeMideRecortado(t *testing.T) {
	if _, err := defecto.Nuevo("p", "i", "s", "  "+strings.Repeat("a", 2000)+"  ", defecto.SeveridadAlta); err != nil {
		t.Errorf("Nuevo() = %v, quiero nil", err)
	}
}
