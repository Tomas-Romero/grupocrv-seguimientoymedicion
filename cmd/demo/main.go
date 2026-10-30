// Comando demo: carga en la base un proyecto de demostracion completo, o lo borra (spec T-023).
//
//	go run ./cmd/demo            carga el proyecto "Tienda online"
//	go run ./cmd/demo --borrar   borra ese proyecto y todo lo suyo
//
// Lee la cadena de conexion de la variable DATABASE_URL, igual que el servidor, y nunca la imprime.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/demo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/migraciones"
)

func main() {
	os.Exit(ejecutar(context.Background(), os.Args[1:], os.Getenv, os.Stdout, os.Stderr))
}

const uso = "Uso: go run ./cmd/demo [--borrar]"

// ejecutar es el comando entero, con sus entradas y salidas a la vista para probarlo: devuelve el codigo de salida (0 o 1).
func ejecutar(ctx context.Context, args []string, getenv func(string) string, stdout, stderr io.Writer) int {
	borrar := false
	switch {
	case len(args) == 0:
	case len(args) == 1 && args[0] == "--borrar":
		borrar = true
	default:
		fmt.Fprintln(stderr, uso)
		return 1
	}

	direccion := getenv("DATABASE_URL")
	if strings.TrimSpace(direccion) == "" {
		fmt.Fprintln(stderr, "Falta la variable DATABASE_URL.")
		return 1
	}

	pool, err := db.Conectar(ctx, direccion)
	if err != nil {
		fmt.Fprintf(stderr, "No se pudo conectar a la base: %s\n", sinSecretos(err, direccion))
		return 1
	}
	defer pool.Close()
	if err := db.Migrar(ctx, pool, migraciones.Archivos); err != nil {
		fmt.Fprintf(stderr, "No se pudo conectar a la base: no se pudieron aplicar las migraciones: %s\n", sinSecretos(err, direccion))
		return 1
	}

	if borrar {
		n, err := demo.BorrarProyecto(ctx, pool)
		if err != nil {
			fmt.Fprintf(stderr, "%s\n", sinSecretos(err, direccion))
			return 1
		}
		if n == 0 {
			fmt.Fprintf(stdout, "No había nada que borrar: la base no tiene el proyecto %q.\n", demo.NombreProyecto)
			return 0
		}
		fmt.Fprintf(stdout, "Proyecto %q borrado, con todo lo suyo.\n", demo.NombreProyecto)
		return 0
	}

	reloj := demo.NuevoReloj()
	casos, lecturas := demo.CasosPostgres(pool, reloj)
	resultado, err := demo.Cargar(ctx, casos, reloj)
	if err != nil {
		var paso *demo.PasoError
		switch {
		case errors.Is(err, demo.ErrYaExiste):
			fmt.Fprintf(stderr, "Ya existe el proyecto %q. Para cargarlo de nuevo, primero: go run ./cmd/demo --borrar\n", demo.NombreProyecto)
		case errors.As(err, &paso):
			fmt.Fprintf(stderr, "Falló el paso %q: %s. Podés volver a empezar con --borrar.\n", paso.Paso, sinSecretos(paso.Err, direccion))
		default:
			fmt.Fprintf(stderr, "%s\n", sinSecretos(err, direccion))
		}
		return 1
	}
	if err := demo.Verificar(ctx, lecturas, resultado); err != nil {
		fmt.Fprintf(stderr, "La carga terminó, pero %s\nPodés volver a empezar con --borrar.\n", sinSecretos(err, direccion))
		return 1
	}
	fmt.Fprintln(stdout, resultado.Resumen)
	return 0
}

// sinSecretos devuelve el texto del error sin la cadena de conexion ni su contrasena, por si el driver los repite en su mensaje
// (RN-T023-7).
func sinSecretos(err error, direccion string) string {
	texto := err.Error()
	if direccion != "" {
		texto = strings.ReplaceAll(texto, direccion, "[DATABASE_URL]")
	}
	if u, e := url.Parse(direccion); e == nil && u.User != nil {
		if clave, ok := u.User.Password(); ok && clave != "" {
			texto = strings.ReplaceAll(texto, clave, "***")
		}
	}
	return texto
}
