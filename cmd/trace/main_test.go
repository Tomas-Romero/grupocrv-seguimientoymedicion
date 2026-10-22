package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// T-007: la matriz del repositorio se arma sin errores: toda spec tiene titulo y criterios, y
// todo archivo de tests se puede leer. Si una spec nueva no sigue la plantilla, este test la nombra.
func TestGenerar_ElRepositorioSeLee(t *testing.T) {
	salida, err := generar(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("generar: %v", err)
	}

	for _, quiero := range []string{"# Matriz de trazabilidad", "### US-001 —", "### T-005 —", "| CA-001-1 |"} {
		if !strings.Contains(salida, quiero) {
			t.Errorf("la matriz del repositorio no tiene %q", quiero)
		}
	}
}

// Una spec sin criterios hace fallar a generar y se nombra en el error.
func TestGenerar_UnaSpecIncompleta(t *testing.T) {
	raiz := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raiz, "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raiz, "specs", "US-001-x.md"), []byte("# US-001 — Algo\n\nSin criterios.\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := generar(raiz)

	if err == nil || !strings.Contains(err.Error(), "specs/US-001-x.md") {
		t.Errorf("generar = %v, quiero un error que nombre specs/US-001-x.md", err)
	}
}

// generar arma la matriz de una carpeta con una spec, un feature y un test.
func TestGenerar_EnlazaLosTresTiposDeArchivo(t *testing.T) {
	raiz := t.TempDir()
	escribir := func(rel, contenido string) {
		t.Helper()
		ruta := filepath.Join(raiz, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta, []byte(contenido), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	escribir("specs/US-001-crear.md", "# US-001 — Crear\n\n- **CA-001-1** (normal) — Se crea.\n")
	escribir("features/US-001-crear.feature", "@US-001\nCaracterística: Crear\n  @CA-001-1\n  Escenario: Se crea un proyecto\n    Dado algo\n")
	escribir("internal/app/crear_test.go", "package app\n\nimport \"testing\"\n\n// US-001 / CA-001-1: se crea.\nfunc TestCrear(t *testing.T) {}\n")
	escribir("bin/ignorado_test.go", "esto no es go")

	salida, err := generar(raiz)
	if err != nil {
		t.Fatalf("generar: %v", err)
	}
	if !strings.Contains(salida, "| CA-001-1 | normal | Se crea un proyecto | TestCrear |") {
		t.Errorf("la matriz no enlaza el criterio con su escenario y su test:\n%s", salida)
	}
}
