// Command trace genera la matriz de trazabilidad (docs/trazabilidad.md): enlaza cada
// historia con su spec, sus criterios de aceptacion, sus escenarios BDD, sus tests y su
// codigo (T-007). Escribe el Markdown en la salida estandar:
//
//	go run ./cmd/trace > docs/trazabilidad.md     (o: make trace)
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/trazabilidad"
)

func main() {
	raiz := flag.String("raiz", ".", "carpeta raiz del repositorio")
	flag.Parse()

	matriz, err := generar(*raiz)
	if err != nil {
		fmt.Fprintln(os.Stderr, "trace:", err)
		os.Exit(1)
	}
	fmt.Print(matriz)
}

// generar lee las specs, los .feature y los archivos de tests del repositorio y devuelve la
// matriz en Markdown. Si una spec no se puede leer (no tiene criterios, por ejemplo) devuelve
// un error que las nombra a todas: una spec incompleta no tiene que pasar sin avisar.
func generar(raiz string) (string, error) {
	var (
		specs      []trazabilidad.Spec
		escenarios []trazabilidad.Escenario
		tests      []trazabilidad.Test
		problemas  []error
	)

	archivos, err := archivosDe(raiz)
	if err != nil {
		return "", err
	}
	for _, rel := range archivos {
		contenido, err := os.ReadFile(filepath.Join(raiz, filepath.FromSlash(rel)))
		if err != nil {
			return "", fmt.Errorf("leer %s: %w", rel, err)
		}
		switch {
		case strings.HasPrefix(rel, "specs/") && strings.HasSuffix(rel, ".md") && strings.Count(rel, "/") == 1:
			spec, err := trazabilidad.LeerSpec(rel, string(contenido))
			if err != nil {
				problemas = append(problemas, err)
				continue
			}
			specs = append(specs, spec)
		case strings.HasPrefix(rel, "features/") && strings.HasSuffix(rel, ".feature"):
			escenarios = append(escenarios, trazabilidad.LeerFeature(rel, string(contenido))...)
		case strings.HasSuffix(rel, "_test.go"):
			leidos, err := trazabilidad.LeerTests(rel, string(contenido))
			if err != nil {
				problemas = append(problemas, err)
				continue
			}
			tests = append(tests, leidos...)
		}
	}
	if len(problemas) > 0 {
		return "", errors.Join(problemas...)
	}
	return trazabilidad.Renderizar(trazabilidad.Armar(specs, escenarios, tests)), nil
}

// archivosDe devuelve, ordenados y con barras, los archivos que le importan a la matriz.
func archivosDe(raiz string) ([]string, error) {
	var archivos []string
	err := filepath.WalkDir(raiz, func(ruta string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "bin", "node_modules", "tmp":
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(raiz, ruta)
		if err != nil {
			return err
		}
		archivos = append(archivos, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("recorrer %s: %w", raiz, err)
	}
	sort.Strings(archivos)
	return archivos, nil
}
