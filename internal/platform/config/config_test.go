package config_test

import (
	"errors"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/config"
)

// entorno simula las variables de entorno sin tocar el proceso real.
func entorno(vars map[string]string) config.LectorEntorno {
	return func(clave string) string { return vars[clave] }
}

// T-001: la configuracion se carga del entorno, aplica defaults y valida.
// Este test es el ejemplo del formato de tabla que usamos en todo el proyecto.
func TestCargar(t *testing.T) {
	casos := []struct {
		nombre      string
		vars        map[string]string
		esperaError error
		esperaAddr  string
		esperaNivel string
	}{
		{
			nombre:      "caso normal: toma todos los valores del entorno",
			vars:        map[string]string{"DATABASE_URL": "postgres://x", "HTTP_ADDR": ":9090", "LOG_LEVEL": "debug"},
			esperaAddr:  ":9090",
			esperaNivel: "debug",
		},
		{
			nombre:      "caso alternativo: aplica los valores por defecto",
			vars:        map[string]string{"DATABASE_URL": "postgres://x"},
			esperaAddr:  ":8080",
			esperaNivel: "info",
		},
		{
			nombre:      "caso limite: los espacios en blanco no cuentan como valor",
			vars:        map[string]string{"DATABASE_URL": "postgres://x", "HTTP_ADDR": "   "},
			esperaAddr:  ":8080",
			esperaNivel: "info",
		},
		{
			nombre:      "error: falta DATABASE_URL",
			vars:        map[string]string{},
			esperaError: config.ErrFaltaDatabaseURL,
		},
		{
			nombre:      "error: HTTP_ADDR sin dos puntos",
			vars:        map[string]string{"DATABASE_URL": "postgres://x", "HTTP_ADDR": "8080"},
			esperaError: config.ErrAddrInvalida,
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			cfg, err := config.Cargar(entorno(c.vars))

			if c.esperaError != nil {
				if !errors.Is(err, c.esperaError) {
					t.Fatalf("se esperaba el error %v, se obtuvo %v", c.esperaError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("no se esperaba error, se obtuvo %v", err)
			}
			if cfg.HTTPAddr != c.esperaAddr {
				t.Errorf("HTTPAddr: se esperaba %q, se obtuvo %q", c.esperaAddr, cfg.HTTPAddr)
			}
			if cfg.LogLevel != c.esperaNivel {
				t.Errorf("LogLevel: se esperaba %q, se obtuvo %q", c.esperaNivel, cfg.LogLevel)
			}
		})
	}
}

// DEF-002: CARGAR_DATOS_EJEMPLO decide si la app carga los datos de ejemplo al
// arrancar. Es falsa salvo que se active (solo en el perfil full del compose).
func TestCargar_DatosEjemplo(t *testing.T) {
	casos := []struct {
		nombre      string
		valor       string
		esperaError error
		esperaDatos bool
	}{
		{nombre: "caso normal: sin definir es falsa", valor: "", esperaDatos: false},
		{nombre: "caso alternativo: true la activa", valor: "true", esperaDatos: true},
		{nombre: "caso alternativo: false la deja apagada", valor: "false", esperaDatos: false},
		{nombre: "caso limite: solo espacios cuenta como sin definir", valor: "   ", esperaDatos: false},
		{nombre: "error: un valor que no es booleano", valor: "quizas", esperaError: config.ErrCargarDatosEjemploInvalido},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			cfg, err := config.Cargar(entorno(map[string]string{
				"DATABASE_URL":         "postgres://x",
				"CARGAR_DATOS_EJEMPLO": c.valor,
			}))

			if c.esperaError != nil {
				if !errors.Is(err, c.esperaError) {
					t.Fatalf("se esperaba el error %v, se obtuvo %v", c.esperaError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("no se esperaba error, se obtuvo %v", err)
			}
			if cfg.CargarDatosEjemplo != c.esperaDatos {
				t.Errorf("CargarDatosEjemplo: se esperaba %v, se obtuvo %v", c.esperaDatos, cfg.CargarDatosEjemplo)
			}
		})
	}
}
