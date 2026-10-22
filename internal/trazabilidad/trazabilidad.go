// Package trazabilidad arma la matriz Historia → Spec → Criterios → BDD → Tests → Codigo
// (T-007) a partir de los archivos del repositorio: las specs de specs/, los escenarios de
// features/ y los comentarios de los tests de Go.
//
// Solo usa la biblioteca estandar. Lee el contenido que le pasan quien llama: no recorre
// carpetas ni escribe archivos, asi se prueba con textos de ejemplo.
package trazabilidad

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Spec es lo que la matriz necesita de una especificacion.
type Spec struct {
	ID        string
	Titulo    string
	Archivo   string
	Estado    string
	Issue     string
	Codigo    []string
	Criterios []Criterio
}

// Criterio es un criterio de aceptacion de una spec (CA-NNN-k, o CA-TNNN-k en las tareas).
type Criterio struct {
	ID    string
	Tipo  string
	Texto string
}

// Escenario es un escenario de un .feature con los criterios que etiqueta.
type Escenario struct {
	Nombre    string
	Archivo   string
	Criterios []string
}

// Test es un test de Go con los IDs que cita en el comentario de arriba.
type Test struct {
	Nombre  string
	Archivo string
	IDs     []string
}

var (
	// reID reconoce un ID de regla, caso o criterio (RN, CL, CA, con T si es de una tarea) o de
	// historia, defecto o tarea.
	reID          = regexp.MustCompile(`\b(?:(?:CA|RN|CL)-T?\d{3}-\d+|(?:US|DEF|T)-\d{3})\b`)
	reIDDeArchivo = regexp.MustCompile(`^(US|DEF|T)-\d{3}`)
	reFila        = regexp.MustCompile(`^\|\s*\*\*(.+?)\*\*\s*\|\s*(.*?)\s*\|\s*$`)
	reCodigo      = regexp.MustCompile("`([^`]+)`")
	reCriterio    = regexp.MustCompile(`^- \*\*(CA-T?\d{3}-\d+)\*\*\s*\(([^)]*)\)\s*(?:—|-)?\s*(.*)$`)
	reEtiquetaCA  = regexp.MustCompile(`^@(CA-T?\d{3}-\d+)$`)
	reEscenario   = regexp.MustCompile(`^\s*(?:Escenario|Esquema del escenario):\s*(.+?)\s*$`)
)

func idDeArchivo(archivo string) string {
	return reIDDeArchivo.FindString(path.Base(archivo))
}

// LeerSpec lee una spec. Es un error que el archivo no empiece con un ID, que no tenga un
// titulo o que no tenga ningun criterio de aceptacion: una spec asi no se puede enlazar.
func LeerSpec(archivo, contenido string) (Spec, error) {
	id := idDeArchivo(archivo)
	if id == "" {
		return Spec{}, fmt.Errorf("%s: el nombre del archivo tiene que empezar con US-NNN, DEF-NNN o T-NNN", archivo)
	}
	spec := Spec{ID: id, Archivo: archivo}

	for _, linea := range strings.Split(strings.ReplaceAll(contenido, "\r\n", "\n"), "\n") {
		switch {
		case spec.Titulo == "" && strings.HasPrefix(linea, "# "):
			titulo := strings.TrimPrefix(linea, "# ")
			if _, despues, ok := strings.Cut(titulo, " — "); ok {
				titulo = despues
			}
			spec.Titulo = strings.TrimSpace(titulo)
		case reFila.MatchString(linea):
			campos := reFila.FindStringSubmatch(linea)
			switch strings.ToLower(campos[1]) {
			case "estado":
				spec.Estado = campos[2]
			case "issue":
				spec.Issue = campos[2]
			case "código", "codigo":
				for _, c := range reCodigo.FindAllStringSubmatch(campos[2], -1) {
					// Solo rutas: los simbolos (`ItemBacklog.Editar`) no se pueden abrir como archivo.
					if !strings.Contains(c[1], "/") {
						continue
					}
					spec.Codigo = append(spec.Codigo, c[1])
				}
			}
		case reCriterio.MatchString(linea):
			c := reCriterio.FindStringSubmatch(linea)
			spec.Criterios = append(spec.Criterios, Criterio{ID: c[1], Tipo: strings.TrimSpace(c[2]), Texto: strings.TrimSpace(c[3])})
		}
	}

	if spec.Titulo == "" {
		return Spec{}, fmt.Errorf("%s: no tiene un titulo (una linea que empiece con \"# \")", archivo)
	}
	if len(spec.Criterios) == 0 {
		return Spec{}, fmt.Errorf("%s: no tiene criterios de aceptacion (lineas \"- **CA-NNN-k** (tipo) — ...\")", archivo)
	}
	return spec, nil
}

// LeerFeature lee los escenarios de un .feature. Los criterios de un escenario son las
// etiquetas @CA-... que tiene justo arriba; las de la caracteristica (@US-NNN) y las de
// tipo (@limite, @error) no cuentan.
func LeerFeature(archivo, contenido string) []Escenario {
	var (
		escenarios []Escenario
		pendientes []string
	)
	for _, linea := range strings.Split(strings.ReplaceAll(contenido, "\r\n", "\n"), "\n") {
		texto := strings.TrimSpace(linea)
		switch {
		case strings.HasPrefix(texto, "@"):
			pendientes = append(pendientes, strings.Fields(texto)...)
		case strings.HasPrefix(texto, "Característica:") || strings.HasPrefix(texto, "Caracteristica:"):
			pendientes = nil
		case reEscenario.MatchString(linea):
			e := Escenario{Nombre: reEscenario.FindStringSubmatch(linea)[1], Archivo: archivo}
			for _, etiqueta := range pendientes {
				if m := reEtiquetaCA.FindStringSubmatch(etiqueta); m != nil {
					e.Criterios = append(e.Criterios, m[1])
				}
			}
			escenarios = append(escenarios, e)
			pendientes = nil
		}
	}
	return escenarios
}

// idsDe devuelve los IDs que aparecen en el texto, sin repetir y en el orden en que aparecen.
func idsDe(texto string) []string {
	var ids []string
	vistos := map[string]bool{}
	for _, id := range reID.FindAllString(texto, -1) {
		if !vistos[id] {
			vistos[id] = true
			ids = append(ids, id)
		}
	}
	return ids
}

// LeerTests lee un archivo de tests de Go y devuelve los tests (funciones TestXxx) que citan
// algun ID en el comentario de arriba (CONTRIBUTING: "ligá el test a la trazabilidad").
func LeerTests(archivo, contenido string) ([]Test, error) {
	arbol, err := parser.ParseFile(token.NewFileSet(), archivo, contenido, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("leer %s: %w", archivo, err)
	}

	var tests []Test
	for _, decl := range arbol.Decls {
		f, ok := decl.(*ast.FuncDecl)
		if !ok || f.Recv != nil || !strings.HasPrefix(f.Name.Name, "Test") || f.Doc == nil {
			continue
		}
		if ids := idsDe(f.Doc.Text()); len(ids) > 0 {
			tests = append(tests, Test{Nombre: f.Name.Name, Archivo: archivo, IDs: ids})
		}
	}
	return tests, nil
}

// HistoriaDe devuelve el ID de la historia, defecto o tarea de un ID de regla, caso o criterio:
// el numero solo es de una historia (CA-019-1 es de US-019) y con la T, de una tarea (CA-T011-1
// es de T-011). Un ID de historia, defecto o tarea se devuelve tal cual.
func HistoriaDe(id string) string {
	partes := strings.Split(id, "-")
	if len(partes) == 3 && (partes[0] == "CA" || partes[0] == "RN" || partes[0] == "CL") {
		if strings.HasPrefix(partes[1], "T") {
			return "T-" + strings.TrimPrefix(partes[1], "T")
		}
		return "US-" + partes[1]
	}
	return id
}

// FilaDeCriterio es un criterio con la evidencia que lo respalda.
type FilaDeCriterio struct {
	Criterio   Criterio
	Escenarios []Escenario
	Tests      []Test
}

// Historia es una spec con sus criterios, su .feature y los tests que la citan.
type Historia struct {
	Spec      Spec
	Feature   string
	Criterios []FilaDeCriterio
	// Tests son todos los tests que citan algun ID de la historia, con o sin criterio.
	Tests []Test
}

// Matriz es el conjunto de historias, ordenadas.
type Matriz struct {
	Historias []Historia
}

// ordenDe ordena las historias: primero las US, despues los DEF y despues las T, y dentro de
// cada grupo por numero.
func ordenDe(id string) (grupo, numero int) {
	switch {
	case strings.HasPrefix(id, "US-"):
		grupo = 0
	case strings.HasPrefix(id, "DEF-"):
		grupo = 1
	default:
		grupo = 2
	}
	numero, _ = strconv.Atoi(id[strings.LastIndex(id, "-")+1:])
	return grupo, numero
}

// Armar enlaza cada criterio de cada spec con los escenarios que lo etiquetan y con los tests
// que lo citan, y cada historia con su .feature (el que lleva su ID en el nombre del archivo).
func Armar(specs []Spec, escenarios []Escenario, tests []Test) Matriz {
	historias := make([]Historia, 0, len(specs))
	for _, spec := range specs {
		h := Historia{Spec: spec}
		for _, e := range escenarios {
			if idDeArchivo(e.Archivo) == spec.ID && h.Feature == "" {
				h.Feature = e.Archivo
			}
		}
		for _, c := range spec.Criterios {
			fila := FilaDeCriterio{Criterio: c}
			for _, e := range escenarios {
				for _, id := range e.Criterios {
					if id == c.ID {
						fila.Escenarios = append(fila.Escenarios, e)
						break
					}
				}
			}
			for _, t := range tests {
				for _, id := range t.IDs {
					if id == c.ID {
						fila.Tests = append(fila.Tests, t)
						break
					}
				}
			}
			h.Criterios = append(h.Criterios, fila)
		}
		for _, t := range tests {
			for _, id := range t.IDs {
				if HistoriaDe(id) == spec.ID {
					h.Tests = append(h.Tests, t)
					break
				}
			}
		}
		historias = append(historias, h)
	}
	sort.SliceStable(historias, func(i, j int) bool {
		gi, ni := ordenDe(historias[i].Spec.ID)
		gj, nj := ordenDe(historias[j].Spec.ID)
		if gi != gj {
			return gi < gj
		}
		return ni < nj
	})
	return Matriz{Historias: historias}
}

// celda escapa lo que no puede ir tal cual en una celda de una tabla de Markdown.
func celda(texto string) string {
	return strings.ReplaceAll(strings.ReplaceAll(texto, "|", `\|`), "\n", " ")
}

func lista(elementos []string, tope int) string {
	if len(elementos) == 0 {
		return "—"
	}
	vistos := map[string]bool{}
	var unicos []string
	for _, e := range elementos {
		if !vistos[e] {
			vistos[e] = true
			unicos = append(unicos, e)
		}
	}
	if len(unicos) > tope {
		return celda(strings.Join(unicos[:tope], ", ")) + fmt.Sprintf(", +%d más", len(unicos)-tope)
	}
	return celda(strings.Join(unicos, ", "))
}

// Renderizar escribe la matriz en Markdown: un resumen por historia y el detalle de cada una.
// El texto sale igual para los mismos datos, asi se puede comparar con lo que hay en docs/.
func Renderizar(m Matriz) string {
	var b strings.Builder
	b.WriteString("# Matriz de trazabilidad\n\n")
	b.WriteString("> Generada con `make trace` a partir de `specs/`, `features/` y los comentarios de los tests.\n")
	b.WriteString("> No se edita a mano: se regenera. Enlaza Historia → Spec → Criterios de aceptación →\n")
	b.WriteString("> Escenarios BDD → Tests → Código (la guía de la cátedra la pide como evidencia).\n\n")

	var criterios, conEscenario, conTest int
	for _, h := range m.Historias {
		for _, f := range h.Criterios {
			criterios++
			if len(f.Escenarios) > 0 {
				conEscenario++
			}
			if len(f.Tests) > 0 {
				conTest++
			}
		}
	}
	fmt.Fprintf(&b, "Total: %d especificaciones, %d criterios de aceptación, %d con escenario BDD y %d con test.\n\n",
		len(m.Historias), criterios, conEscenario, conTest)
	b.WriteString("Un criterio sin escenario BDD o sin test se muestra con `—`: o falta la evidencia, o el criterio\n")
	b.WriteString("se cubre de otra forma y la spec lo dice (por ejemplo, las tareas técnicas no llevan `.feature`).\n\n")

	b.WriteString("## Resumen por historia\n\n")
	b.WriteString("| Historia | Spec | Estado | Criterios | Con escenario | Con test | Código |\n")
	b.WriteString("|---|---|---|---|---|---|---|\n")
	for _, h := range m.Historias {
		var conE, conT int
		for _, f := range h.Criterios {
			if len(f.Escenarios) > 0 {
				conE++
			}
			if len(f.Tests) > 0 {
				conT++
			}
		}
		codigo := "—"
		if len(h.Spec.Codigo) > 0 {
			partes := make([]string, len(h.Spec.Codigo))
			for i, c := range h.Spec.Codigo {
				partes[i] = "`" + c + "`"
			}
			codigo = celda(strings.Join(partes, ", "))
		}
		fmt.Fprintf(&b, "| %s | [%s](../%s) | %s | %d | %d | %d | %s |\n",
			h.Spec.ID, h.Spec.Archivo, h.Spec.Archivo, celda(h.Spec.Estado), len(h.Criterios), conE, conT, codigo)
	}

	b.WriteString("\n## Detalle\n")
	for _, h := range m.Historias {
		fmt.Fprintf(&b, "\n### %s — %s\n\n", h.Spec.ID, h.Spec.Titulo)
		fmt.Fprintf(&b, "- **Spec:** [%s](../%s)", h.Spec.Archivo, h.Spec.Archivo)
		if h.Spec.Issue != "" {
			fmt.Fprintf(&b, " · **Issue:** %s", h.Spec.Issue)
		}
		b.WriteString("\n")
		if h.Feature != "" {
			fmt.Fprintf(&b, "- **Escenarios BDD:** [%s](../%s)\n", h.Feature, h.Feature)
		} else {
			b.WriteString("- **Escenarios BDD:** ninguno\n")
		}
		if len(h.Spec.Codigo) > 0 {
			partes := make([]string, len(h.Spec.Codigo))
			for i, c := range h.Spec.Codigo {
				partes[i] = "`" + c + "`"
			}
			fmt.Fprintf(&b, "- **Código:** %s\n", strings.Join(partes, ", "))
		}
		fmt.Fprintf(&b, "- **Tests que la citan:** %d\n\n", len(h.Tests))
		b.WriteString("| Criterio | Tipo | Escenarios BDD | Tests |\n")
		b.WriteString("|---|---|---|---|\n")
		for _, f := range h.Criterios {
			nombres := make([]string, len(f.Escenarios))
			for i, e := range f.Escenarios {
				nombres[i] = e.Nombre
			}
			tests := make([]string, len(f.Tests))
			for i, t := range f.Tests {
				tests[i] = t.Nombre
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", f.Criterio.ID, celda(f.Criterio.Tipo), lista(nombres, 3), lista(tests, 4))
		}
	}
	return b.String()
}
