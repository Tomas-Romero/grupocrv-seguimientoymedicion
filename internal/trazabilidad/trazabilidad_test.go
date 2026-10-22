package trazabilidad_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/trazabilidad"
)

const specDeEjemplo = `# US-019 — Registrar esfuerzo con integrante, fecha, actividad y horas

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #22 |
| **Escenarios BDD** | ` + "`features/US-019-registrar-esfuerzo.feature`" + ` |
| **Código** | ` + "`internal/domain/esfuerzo/`, caso de uso en `internal/app/`, `ItemBacklog.Editar`, migración en `migraciones/`" + ` |

## 8. Criterios de aceptación

- **CA-019-1** (normal) — Un integrante registra sus horas.
- **CA-019-2** (alternativo) — Se pueden registrar varias veces el mismo día,
  y la descripción sigue en la línea de abajo.
- **CA-019-9** (error) — Una historia inexistente se rechaza.

Esta línea no es un criterio: CA-019-3 aparece solo como mención.
`

// T-007: de una spec salen el ID, el titulo, el estado, el issue, el codigo y los criterios.
func TestLeerSpec(t *testing.T) {
	got, err := trazabilidad.LeerSpec("specs/US-019-registrar-esfuerzo.md", specDeEjemplo)
	if err != nil {
		t.Fatalf("LeerSpec: %v", err)
	}

	if got.ID != "US-019" || got.Titulo != "Registrar esfuerzo con integrante, fecha, actividad y horas" {
		t.Errorf("ID y titulo = %q, %q", got.ID, got.Titulo)
	}
	if got.Archivo != "specs/US-019-registrar-esfuerzo.md" || got.Estado != "borrador" || got.Issue != "#22" {
		t.Errorf("archivo, estado e issue = %q, %q, %q", got.Archivo, got.Estado, got.Issue)
	}
	// Solo las rutas: los simbolos de Go que la spec nombra en esa fila (ItemBacklog.Editar) no son codigo que se pueda abrir.
	quiero := []string{"internal/domain/esfuerzo/", "internal/app/", "migraciones/"}
	if !reflect.DeepEqual(got.Codigo, quiero) {
		t.Errorf("Codigo = %v, quiero %v", got.Codigo, quiero)
	}
	if len(got.Criterios) != 3 || got.Criterios[0].ID != "CA-019-1" || got.Criterios[0].Tipo != "normal" ||
		got.Criterios[1].ID != "CA-019-2" || got.Criterios[2].ID != "CA-019-9" || got.Criterios[2].Tipo != "error" {
		t.Errorf("Criterios = %+v, quiero CA-019-1, CA-019-2 y CA-019-9", got.Criterios)
	}
	if got.Criterios[1].Texto != "Se pueden registrar varias veces el mismo día," {
		t.Errorf("Criterios[1].Texto = %q, quiero la primera linea del criterio", got.Criterios[1].Texto)
	}
}

// Las tareas llevan la T en sus identificadores (RN-T005-k, CA-T005-k) y en el nombre del archivo.
func TestLeerSpec_DeUnaTarea(t *testing.T) {
	contenido := "# T-005 — Validaciones y errores uniformes\n\n| | |\n|---|---|\n| **Estado** | borrador |\n\n- **CA-T005-1** (normal) — Un error de validación responde 422.\n"

	got, err := trazabilidad.LeerSpec("specs/T-005-validaciones-errores-uniformes.md", contenido)
	if err != nil {
		t.Fatalf("LeerSpec: %v", err)
	}
	if got.ID != "T-005" || len(got.Criterios) != 1 || got.Criterios[0].ID != "CA-T005-1" {
		t.Errorf("LeerSpec = %+v, quiero T-005 con CA-T005-1", got)
	}
}

// Una spec sin criterios es un error: la matriz no puede dejar pasar una spec incompleta sin avisar.
func TestLeerSpec_Errores(t *testing.T) {
	casos := []struct {
		nombre, archivo, contenido string
	}{
		{"el nombre del archivo no empieza con un ID", "specs/leeme.md", "# Algo\n- **CA-001-1** (normal) — x\n"},
		{"no tiene titulo", "specs/US-001-algo.md", "- **CA-001-1** (normal) — x\n"},
		{"no tiene criterios", "specs/US-001-algo.md", "# US-001 — Algo\n\nSin criterios.\n"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if _, err := trazabilidad.LeerSpec(c.archivo, c.contenido); err == nil {
				t.Error("LeerSpec no devolvio un error")
			}
		})
	}
}

const featureDeEjemplo = `# language: es
@US-019
Característica: Registrar esfuerzo

  # ---------------------------------------------------------------- caso normal
  @CA-019-1
  Escenario: Un integrante registra sus horas
    Dado un proyecto "Demo"

  @CA-019-2 @limite
  Escenario: Dos registros suman más de 24 horas
    Dado un proyecto "Demo"

  @CA-019-9 @CA-019-10 @error
  Escenario: Una historia inexistente se rechaza
    Dado un proyecto "Demo"

  Escenario: Sin etiqueta de criterio
    Dado un proyecto "Demo"
`

// T-007: de un .feature salen los escenarios con los criterios que etiquetan.
func TestLeerFeature(t *testing.T) {
	got := trazabilidad.LeerFeature("features/US-019-registrar-esfuerzo.feature", featureDeEjemplo)

	if len(got) != 4 {
		t.Fatalf("hay %d escenarios, quiero 4: %+v", len(got), got)
	}
	if got[0].Nombre != "Un integrante registra sus horas" || !reflect.DeepEqual(got[0].Criterios, []string{"CA-019-1"}) {
		t.Errorf("escenario 1 = %+v", got[0])
	}
	if !reflect.DeepEqual(got[1].Criterios, []string{"CA-019-2"}) {
		t.Errorf("escenario 2 = %+v, quiero solo CA-019-2 (@limite no es un criterio)", got[1])
	}
	if !reflect.DeepEqual(got[2].Criterios, []string{"CA-019-9", "CA-019-10"}) {
		t.Errorf("escenario 3 = %+v, quiero CA-019-9 y CA-019-10", got[2])
	}
	if len(got[3].Criterios) != 0 {
		t.Errorf("escenario 4 = %+v, quiero sin criterios", got[3])
	}
	if got[0].Archivo != "features/US-019-registrar-esfuerzo.feature" {
		t.Errorf("Archivo = %q", got[0].Archivo)
	}
}

// La etiqueta de un escenario no se pasa al siguiente, ni la de la caracteristica a ningun escenario.
func TestLeerFeature_LasEtiquetasNoSeFiltran(t *testing.T) {
	contenido := "@US-001\nCaracterística: X\n  @CA-001-1\n  Escenario: Uno\n    Dado algo\n  Escenario: Dos\n    Dado algo\n"

	got := trazabilidad.LeerFeature("features/US-001-x.feature", contenido)

	if len(got) != 2 || len(got[0].Criterios) != 1 || len(got[1].Criterios) != 0 {
		t.Errorf("escenarios = %+v, quiero Uno con CA-001-1 y Dos sin criterios", got)
	}
}

const testDeEjemplo = `package esfuerzo_test

import "testing"

// US-019 / CA-019-1, RN-019-9: el registro se guarda con un ID.
func TestRegistrar(t *testing.T) {}

// T-011 / CA-T011-1: el recorrido completo.
// Segunda linea del comentario con RN-T011-3.
func TestRecorrido(t *testing.T) {}

// Un test sin ID en el comentario.
func TestSinID(t *testing.T) {}

func TestSinComentario(t *testing.T) {}

// US-019 / CA-019-2: un helper no es un test.
func ayuda() {}
`

// T-007: de un archivo de tests salen los tests con los IDs que citan en el comentario de arriba.
func TestLeerTests(t *testing.T) {
	got, err := trazabilidad.LeerTests("internal/adapters/postgres/esfuerzo_test.go", testDeEjemplo)
	if err != nil {
		t.Fatalf("LeerTests: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("hay %d tests con IDs, quiero 2: %+v", len(got), got)
	}
	if got[0].Nombre != "TestRegistrar" || got[0].Archivo != "internal/adapters/postgres/esfuerzo_test.go" {
		t.Errorf("test 1 = %+v", got[0])
	}
	if !reflect.DeepEqual(got[0].IDs, []string{"US-019", "CA-019-1", "RN-019-9"}) {
		t.Errorf("IDs del test 1 = %v", got[0].IDs)
	}
	if !reflect.DeepEqual(got[1].IDs, []string{"T-011", "CA-T011-1", "RN-T011-3"}) {
		t.Errorf("IDs del test 2 = %v, quiero los dos renglones del comentario", got[1].IDs)
	}
}

func TestLeerTests_ArchivoInvalido(t *testing.T) {
	if _, err := trazabilidad.LeerTests("x_test.go", "esto no es go"); err == nil {
		t.Error("LeerTests no devolvio un error con un archivo que no es Go")
	}
}

// El ID de la historia de un criterio: el numero solo es de una historia y con la T, de una tarea.
func TestHistoriaDe(t *testing.T) {
	casos := map[string]string{
		"CA-019-1":  "US-019",
		"RN-009-6":  "US-009",
		"CL-014-10": "US-014",
		"CA-T005-3": "T-005",
		"RN-T011-3": "T-011",
		"US-019":    "US-019",
		"T-011":     "T-011",
		"DEF-004":   "DEF-004",
	}
	for id, quiero := range casos {
		if got := trazabilidad.HistoriaDe(id); got != quiero {
			t.Errorf("HistoriaDe(%q) = %q, quiero %q", id, got, quiero)
		}
	}
}

func armarMatriz(t *testing.T) trazabilidad.Matriz {
	t.Helper()
	spec, err := trazabilidad.LeerSpec("specs/US-019-registrar-esfuerzo.md", specDeEjemplo)
	if err != nil {
		t.Fatalf("LeerSpec: %v", err)
	}
	otra, err := trazabilidad.LeerSpec("specs/T-005-errores.md", "# T-005 — Errores\n\n- **CA-T005-1** (normal) — Un error responde 422.\n")
	if err != nil {
		t.Fatalf("LeerSpec: %v", err)
	}
	escenarios := trazabilidad.LeerFeature("features/US-019-registrar-esfuerzo.feature", featureDeEjemplo)
	tests, err := trazabilidad.LeerTests("internal/app/registrar_esfuerzo_test.go", testDeEjemplo)
	if err != nil {
		t.Fatalf("LeerTests: %v", err)
	}
	return trazabilidad.Armar([]trazabilidad.Spec{otra, spec}, escenarios, tests)
}

// T-007: cada criterio queda enlazado con los escenarios que lo etiquetan y con los tests que lo citan.
func TestArmar_EnlazaCriteriosEscenariosYTests(t *testing.T) {
	m := armarMatriz(t)

	if len(m.Historias) != 2 || m.Historias[0].Spec.ID != "T-005" && m.Historias[0].Spec.ID != "US-019" {
		t.Fatalf("historias = %+v", m.Historias)
	}
	var h trazabilidad.Historia
	for _, x := range m.Historias {
		if x.Spec.ID == "US-019" {
			h = x
		}
	}
	porCriterio := map[string]trazabilidad.FilaDeCriterio{}
	for _, f := range h.Criterios {
		porCriterio[f.Criterio.ID] = f
	}
	if len(porCriterio["CA-019-1"].Escenarios) != 1 || porCriterio["CA-019-1"].Escenarios[0].Nombre != "Un integrante registra sus horas" {
		t.Errorf("CA-019-1 deberia tener su escenario: %+v", porCriterio["CA-019-1"])
	}
	if len(porCriterio["CA-019-1"].Tests) != 1 || porCriterio["CA-019-1"].Tests[0].Nombre != "TestRegistrar" {
		t.Errorf("CA-019-1 deberia tener TestRegistrar: %+v", porCriterio["CA-019-1"])
	}
	if len(porCriterio["CA-019-9"].Escenarios) != 1 || len(porCriterio["CA-019-9"].Tests) != 0 {
		t.Errorf("CA-019-9 tiene escenario y ningun test: %+v", porCriterio["CA-019-9"])
	}
	if len(porCriterio["CA-019-2"].Escenarios) != 1 {
		t.Errorf("CA-019-2 deberia tener su escenario")
	}
	if h.Feature != "features/US-019-registrar-esfuerzo.feature" {
		t.Errorf("Feature = %q", h.Feature)
	}
	if len(h.Tests) != 1 {
		t.Errorf("los tests que citan la historia = %d, quiero 1 (TestRegistrar)", len(h.Tests))
	}
}

// Las historias salen ordenadas: US primero, despues DEF y despues T, cada grupo por numero.
func TestArmar_Orden(t *testing.T) {
	especs := []string{"specs/T-004-a.md", "specs/US-010-a.md", "specs/US-002-a.md", "specs/T-011-a.md", "specs/US-100-a.md"}
	var specs []trazabilidad.Spec
	for _, e := range especs {
		id := strings.TrimSuffix(e[len("specs/"):], "-a.md")
		s, err := trazabilidad.LeerSpec(e, "# "+id+" — Titulo\n\n- **CA-"+strings.TrimPrefix(id, "US-")+"-1** (normal) — x\n")
		if err != nil {
			// las tareas llevan la T en el criterio
			s, err = trazabilidad.LeerSpec(e, "# "+id+" — Titulo\n\n- **CA-"+strings.ReplaceAll(id, "-", "")+"-1** (normal) — x\n")
			if err != nil {
				t.Fatalf("LeerSpec %s: %v", e, err)
			}
		}
		specs = append(specs, s)
	}

	m := trazabilidad.Armar(specs, nil, nil)

	var ids []string
	for _, h := range m.Historias {
		ids = append(ids, h.Spec.ID)
	}
	quiero := []string{"US-002", "US-010", "US-100", "T-004", "T-011"}
	if !reflect.DeepEqual(ids, quiero) {
		t.Errorf("orden = %v, quiero %v", ids, quiero)
	}
}

// T-007: la matriz en Markdown tiene el resumen y el detalle de cada historia.
func TestRenderizar(t *testing.T) {
	salida := trazabilidad.Renderizar(armarMatriz(t))

	for _, quiero := range []string{
		"# Matriz de trazabilidad",
		"make trace",
		"| US-019 |",
		"[specs/US-019-registrar-esfuerzo.md](../specs/US-019-registrar-esfuerzo.md)",
		"### US-019 — Registrar esfuerzo con integrante, fecha, actividad y horas",
		"| CA-019-1 | normal | Un integrante registra sus horas | TestRegistrar |",
		"| CA-019-9 | error | Una historia inexistente se rechaza | — |",
		"`internal/domain/esfuerzo/`",
		"### T-005 — Errores",
	} {
		if !strings.Contains(salida, quiero) {
			t.Errorf("la matriz no tiene %q:\n%s", quiero, salida)
		}
	}
	if salida != trazabilidad.Renderizar(armarMatriz(t)) {
		t.Error("renderizar dos veces lo mismo da salidas distintas: la matriz tiene que ser determinista")
	}
}

// El resumen cuenta los criterios con escenario y con test, para ver donde falta evidencia.
func TestRenderizar_Resumen(t *testing.T) {
	salida := trazabilidad.Renderizar(armarMatriz(t))

	// US-019: 3 criterios, 3 con escenario, 1 con test.
	if !strings.Contains(salida, "| US-019 | [specs/US-019-registrar-esfuerzo.md](../specs/US-019-registrar-esfuerzo.md) | borrador | 3 | 3 | 1 |") {
		t.Errorf("el resumen de US-019 no cuenta bien:\n%s", salida)
	}
	if !strings.Contains(salida, "Total: 2 especificaciones, 4 criterios de aceptación, 3 con escenario BDD y 1 con test") {
		t.Errorf("el total no cuenta bien:\n%s", salida)
	}
}
