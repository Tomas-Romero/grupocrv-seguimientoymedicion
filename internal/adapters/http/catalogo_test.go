package httpadapter

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/errores"
)

// directoriosConErrores devuelve internal/app y cada paquete de internal/domain. Los de
// internal/domain se leen del disco y no de una lista fija: asi un paquete nuevo (por
// ejemplo, planningpoker en US-014) no queda afuera del control de RN-T005-19.
func directoriosConErrores(t *testing.T) []string {
	t.Helper()
	dirs := []string{"../../app"}
	entradas, err := os.ReadDir("../../domain")
	if err != nil {
		t.Fatalf("leer internal/domain: %v", err)
	}
	for _, e := range entradas {
		if e.IsDir() {
			dirs = append(dirs, "../../domain/"+e.Name())
		}
	}
	return dirs
}

// erroresDeCodigoFuente devuelve el nombre ("paquete.ErrAlgo") de cada error exportado de
// internal/domain y de internal/app, leyendo el codigo fuente (RN-T005-19).
func erroresDeCodigoFuente(t *testing.T) []string {
	t.Helper()
	var nombres []string
	for _, dir := range directoriosConErrores(t) {
		paquetes, err := parser.ParseDir(token.NewFileSet(), dir, func(f fs.FileInfo) bool {
			return !strings.HasSuffix(f.Name(), "_test.go")
		}, 0)
		if err != nil {
			t.Fatalf("leer %s: %v", dir, err)
		}
		for nombre, paquete := range paquetes {
			for _, archivo := range paquete.Files {
				for _, decl := range archivo.Decls {
					general, ok := decl.(*ast.GenDecl)
					if !ok || general.Tok != token.VAR {
						continue
					}
					for _, spec := range general.Specs {
						for _, id := range spec.(*ast.ValueSpec).Names {
							if strings.HasPrefix(id.Name, "Err") && id.IsExported() {
								nombres = append(nombres, nombre+"."+id.Name)
							}
						}
					}
				}
			}
		}
	}
	sort.Strings(nombres)
	return nombres
}

// RN-T005-19, CA-T005-8: todo error exportado esta en el catalogo o en las exclusiones, y todo lo
// que el catalogo nombra existe en el codigo.
func TestCatalogo_CubreTodosLosErrores(t *testing.T) {
	existentes := map[string]bool{}
	for _, n := range erroresDeCodigoFuente(t) {
		existentes[n] = true
	}
	if len(existentes) < 30 {
		t.Fatalf("se leyeron %d errores del codigo, esperaba mas de 30: el test no esta mirando donde debe", len(existentes))
	}

	registrados := map[string]bool{}
	for _, e := range entradasDelCatalogo {
		if registrados[e.nombre] {
			t.Errorf("%s esta dos veces en el catalogo", e.nombre)
		}
		registrados[e.nombre] = true
		if !existentes[e.nombre] {
			t.Errorf("el catalogo nombra %s, que no existe en el codigo", e.nombre)
		}
	}
	for nombre, motivo := range exclusionesDelCatalogo {
		if registrados[nombre] {
			t.Errorf("%s esta en el catalogo y en las exclusiones", nombre)
		}
		if !existentes[nombre] {
			t.Errorf("las exclusiones nombran %s, que no existe en el codigo", nombre)
		}
		if strings.TrimSpace(motivo) == "" {
			t.Errorf("la exclusion de %s no tiene motivo", nombre)
		}
	}

	var faltan []string
	for nombre := range existentes {
		if !registrados[nombre] {
			if _, excluido := exclusionesDelCatalogo[nombre]; !excluido {
				faltan = append(faltan, nombre)
			}
		}
	}
	sort.Strings(faltan)
	if len(faltan) > 0 {
		t.Errorf("estos errores no estan en el catalogo ni en las exclusiones (agregalos en catalogo.go): %s", strings.Join(faltan, ", "))
	}
}

// Cada entrada tiene un error real, un tipo valido y, si tiene campo, uno de los de RN-T005-8.
func TestCatalogo_EntradasValidas(t *testing.T) {
	camposValidos := map[string]bool{"": true, "nombre": true, "email": true, "rol": true, "titulo": true, "prioridad": true,
		"criterios": true, "estimacion": true, "sprint_goal": true, "fecha_inicio": true, "fecha_fin": true, "fecha": true,
		"actividad": true, "horas": true}
	tipos := map[errores.Tipo]bool{errores.Validacion: true, errores.NoEncontrado: true, errores.Conflicto: true}

	for _, e := range entradasDelCatalogo {
		if e.err == nil {
			t.Errorf("%s: no tiene error", e.nombre)
		}
		if !tipos[e.tipo] {
			t.Errorf("%s: tipo %q no valido para el catalogo", e.nombre, e.tipo)
		}
		if !camposValidos[e.campo] {
			t.Errorf("%s: campo %q fuera de los nombres de RN-T005-8", e.nombre, e.campo)
		}
		if e.tipo == errores.NoEncontrado && e.campo != "" {
			t.Errorf("%s: un error de inexistencia no lleva campo", e.nombre)
		}
	}
}

// Los mismos errores no se repiten por valor: dos nombres distintos no pueden apuntar al mismo error.
func TestCatalogo_NoRepiteErrores(t *testing.T) {
	vistos := map[error]string{}
	for _, e := range entradasDelCatalogo {
		if otro, ok := vistos[e.err]; ok {
			t.Errorf("%s y %s son el mismo error", e.nombre, otro)
		}
		vistos[e.err] = e.nombre
	}
}

// El catalogo clasifica con el tipo, el campo y el mensaje de la spec (seccion 7).
func TestCatalogoDeErrores_Clasifica(t *testing.T) {
	casos := []struct {
		nombre string
		err    error
		tipo   errores.Tipo
		campo  string
		estado int
	}{
		{"validacion de un campo", backlog.ErrTituloVacio, errores.Validacion, "titulo", http.StatusUnprocessableEntity},
		{"horas", esfuerzo.ErrHorasFraccion, errores.Validacion, "horas", http.StatusUnprocessableEntity},
		{"fecha de fin del sprint", sprint.ErrDuracionExcedida, errores.Validacion, "fecha_fin", http.StatusUnprocessableEntity},
		{"mismo nombre en otro paquete", proyecto.ErrNombreLargo, errores.Validacion, "nombre", http.StatusUnprocessableEntity},
		{"email del integrante", integrante.ErrEmailInvalido, errores.Validacion, "email", http.StatusUnprocessableEntity},
		{"inexistente", app.ErrItemInexistente, errores.NoEncontrado, "", http.StatusNotFound},
		{"inexistente envuelto", fmt.Errorf("asignar: %w (id 9)", app.ErrProyectoInexistente), errores.NoEncontrado, "", http.StatusNotFound},
		{"email duplicado", app.ErrEmailDuplicado, errores.Conflicto, "email", http.StatusConflict},
		{"item completado", backlog.ErrItemCompletado, errores.Conflicto, "", http.StatusConflict},
		{"ultimo sprint abierto", sprint.ErrUltimoSprintAbierto, errores.Conflicto, "", http.StatusConflict},
		{"inicio no posterior al ultimo", sprint.ErrInicioNoPosteriorAlUltimo, errores.Conflicto, "fecha_inicio", http.StatusConflict},
	}
	catalogo := CatalogoDeErrores()
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			problema := catalogo.Clasificar(c.err)
			if problema.Tipo != c.tipo || len(problema.Errores) != 1 || problema.Errores[0].Campo != c.campo {
				t.Fatalf("Clasificar() = %+v, quiero tipo %s con un error del campo %q", problema, c.tipo, c.campo)
			}
			if got := EstadoHTTP(problema.Tipo); got != c.estado {
				t.Errorf("EstadoHTTP = %d, quiero %d", got, c.estado)
			}
		})
	}
}

// Un error de las exclusiones y uno desconocido salen como internos (spec, seccion 7).
func TestCatalogoDeErrores_LoExcluidoEsInterno(t *testing.T) {
	for _, err := range []error{errors.New("sin registrar"), metricas.ErrResumenInconsistente} {
		if tipo := CatalogoDeErrores().Clasificar(err).Tipo; tipo != errores.Interno {
			t.Errorf("Clasificar(%v) = %s, quiero interno", err, tipo)
		}
	}
}

// Un caso de uso real devuelve un error que el manejador responde con el estado de la spec.
func TestCatalogoDeErrores_ConElManejador(t *testing.T) {
	m := NuevoManejadorDeErrores(CatalogoDeErrores(), nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.Header.Set("Accept", "application/json")

	m.ResponderError(w, r, errors.Join(backlog.ErrTituloVacio, backlog.ErrPrioridadInvalida))

	const quiero = `{"tipo":"validacion","errores":[{"campo":"titulo","mensaje":"el titulo es obligatorio"},` +
		`{"campo":"prioridad","mensaje":"la prioridad tiene que ser must, should, could o wont"}]}`
	if w.Code != http.StatusUnprocessableEntity || strings.TrimSpace(w.Body.String()) != quiero {
		t.Errorf("respuesta = %d %s, quiero 422 y %s", w.Code, w.Body.String(), quiero)
	}
}
