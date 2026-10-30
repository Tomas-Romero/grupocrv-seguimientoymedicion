package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// salida corre el comando con esas entradas y devuelve el codigo, lo que imprimio y lo que mando al error estandar.
func salida(t *testing.T, args []string, env map[string]string) (int, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	codigo := ejecutar(context.Background(), args, func(k string) string { return env[k] }, &stdout, &stderr)
	return codigo, stdout.String(), stderr.String()
}

// CA-T023-5: sin DATABASE_URL, con un argumento desconocido o con una conexion que falla, el comando termina con 1, con un mensaje
// claro, y sin repetir la cadena de conexion.
func TestEjecutar_Errores(t *testing.T) {
	const clave = "contrasena-secreta-123"
	// Un puerto cerrado en localhost: la conexion falla de inmediato.
	direccion := "postgres://usuario:" + clave + "@127.0.0.1:1/base?sslmode=disable&connect_timeout=2"

	casos := []struct {
		nombre string
		args   []string
		env    map[string]string
		quiere string
	}{
		{"sin DATABASE_URL", nil, nil, "Falta la variable DATABASE_URL."},
		{"DATABASE_URL en blanco", nil, map[string]string{"DATABASE_URL": "   "}, "Falta la variable DATABASE_URL."},
		{"argumento desconocido", []string{"--otra-cosa"}, map[string]string{"DATABASE_URL": direccion}, "Uso: go run ./cmd/demo [--borrar]"},
		{"dos argumentos", []string{"--borrar", "--borrar"}, map[string]string{"DATABASE_URL": direccion}, "Uso: go run ./cmd/demo [--borrar]"},
		{"conexion que falla", nil, map[string]string{"DATABASE_URL": direccion}, "No se pudo conectar a la base:"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			codigo, stdout, stderr := salida(t, c.args, c.env)
			if codigo != 1 {
				t.Errorf("codigo = %d, se esperaba 1", codigo)
			}
			if stdout != "" {
				t.Errorf("imprimio en la salida estandar: %q", stdout)
			}
			if !strings.Contains(stderr, c.quiere) {
				t.Errorf("error estandar = %q, se esperaba que contuviera %q", stderr, c.quiere)
			}
			if strings.Contains(stderr, clave) || strings.Contains(stderr, direccion) {
				t.Errorf("el mensaje repite la cadena de conexion o su contrasena: %q", stderr)
			}
		})
	}
}

// RN-T023-7: aunque el driver repita la cadena de conexion en su mensaje, no sale.
func TestSinSecretos(t *testing.T) {
	const direccion = "postgres://usuario:clave-larga@host:5432/base"
	err := errors.New("fallo con " + direccion + " y con la clave clave-larga")

	got := sinSecretos(err, direccion)

	if strings.Contains(got, "clave-larga") || strings.Contains(got, direccion) {
		t.Errorf("sinSecretos = %q todavia tiene secretos", got)
	}
}

// El comando entero contra una base Postgres real y vacia: carga, se niega a cargar dos veces, borra y avisa si no habia nada.
func TestEjecutar_CargarNegarseBorrar(t *testing.T) {
	ctx := context.Background()
	direccion := dbprueba.BaseMigrada(ctx, t).Config().ConnString()
	env := map[string]string{"DATABASE_URL": direccion}

	codigo, stdout, stderr := salida(t, nil, env)
	if codigo != 0 || !strings.HasPrefix(stdout, `Proyecto "Tienda online" cargado: 4 integrantes, 16 historias, 4 sprints`) {
		t.Fatalf("carga: codigo %d, salida %q, error %q", codigo, stdout, stderr)
	}
	if strings.Contains(stdout+stderr, direccion) {
		t.Error("el comando imprimio la cadena de conexion")
	}

	codigo, stdout, stderr = salida(t, nil, env)
	if codigo != 1 || stdout != "" || !strings.Contains(stderr, `Ya existe el proyecto "Tienda online". Para cargarlo de nuevo, primero: go run ./cmd/demo --borrar`) {
		t.Errorf("segunda carga: codigo %d, salida %q, error %q", codigo, stdout, stderr)
	}

	codigo, stdout, _ = salida(t, []string{"--borrar"}, env)
	if codigo != 0 || !strings.Contains(stdout, "borrado") {
		t.Errorf("borrar: codigo %d, salida %q", codigo, stdout)
	}
	codigo, stdout, _ = salida(t, []string{"--borrar"}, env)
	if codigo != 0 || !strings.Contains(stdout, "No había nada que borrar") {
		t.Errorf("borrar sin proyecto: codigo %d, salida %q", codigo, stdout)
	}

	codigo, stdout, stderr = salida(t, nil, env)
	if codigo != 0 {
		t.Errorf("cargar despues de borrar: codigo %d, salida %q, error %q", codigo, stdout, stderr)
	}
}
