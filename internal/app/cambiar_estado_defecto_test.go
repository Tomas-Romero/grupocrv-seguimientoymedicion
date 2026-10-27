package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// cambioDeEstadoGuardado es lo que el caso de uso le pidio guardar al repositorio.
type cambioDeEstadoGuardado struct {
	defecto        defecto.Defecto
	estadoAnterior defecto.Estado
}

// repositorioCambioDeEstadoFalso simula la persistencia del cambio de estado sin base de datos.
// Responde lo que se le configura y anota el orden en que le preguntan.
type repositorioCambioDeEstadoFalso struct {
	existeProyecto bool
	defectos       map[string]defecto.Defecto // por ID, todos del proyecto "proyecto-1"
	sprints        map[string]int             // ID -> numero, todos del proyecto "proyecto-1"

	errProyecto   error
	errDefecto    error
	errSprint     error
	errActualizar error

	llamadas []string
	guardado []cambioDeEstadoGuardado
}

func (r *repositorioCambioDeEstadoFalso) ExisteProyecto(_ context.Context, _ string) (bool, error) {
	r.llamadas = append(r.llamadas, "proyecto")
	return r.existeProyecto, r.errProyecto
}

func (r *repositorioCambioDeEstadoFalso) DefectoEnProyecto(_ context.Context, _, defectoID string) (defecto.Defecto, bool, error) {
	r.llamadas = append(r.llamadas, "defecto")
	d, ok := r.defectos[defectoID]
	return d, ok, r.errDefecto
}

func (r *repositorioCambioDeEstadoFalso) NumeroDeSprintEnProyecto(_ context.Context, _, sprintID string) (int, bool, error) {
	r.llamadas = append(r.llamadas, "sprint:"+sprintID)
	numero, ok := r.sprints[sprintID]
	return numero, ok, r.errSprint
}

func (r *repositorioCambioDeEstadoFalso) ActualizarEstadoDefecto(_ context.Context, d defecto.Defecto, estadoAnterior defecto.Estado) error {
	r.llamadas = append(r.llamadas, "actualizar")
	r.guardado = append(r.guardado, cambioDeEstadoGuardado{defecto: d, estadoAnterior: estadoAnterior})
	return r.errActualizar
}

// defectoGuardadoEn es un defecto del "proyecto-1", detectado en el "sprint-1" (numero 1).
func defectoGuardadoEn(estado defecto.Estado, sprintResolucionID string) defecto.Defecto {
	return defecto.Defecto{
		ID:                 "defecto-1",
		ProyectoID:         "proyecto-1",
		ItemID:             "item-1",
		SprintDeteccionID:  "sprint-1",
		Descripcion:        "Falla al guardar",
		Severidad:          defecto.SeveridadAlta,
		Estado:             estado,
		SprintResolucionID: sprintResolucionID,
	}
}

// repositorioConDefecto es un repositorio donde el proyecto existe, el defecto esta guardado en el
// estado indicado y hay tres sprints (sprint-1, sprint-2 y sprint-3, con esos numeros).
func repositorioConDefecto(estado defecto.Estado, sprintResolucionID string) *repositorioCambioDeEstadoFalso {
	return &repositorioCambioDeEstadoFalso{
		existeProyecto: true,
		defectos:       map[string]defecto.Defecto{"defecto-1": defectoGuardadoEn(estado, sprintResolucionID)},
		sprints:        map[string]int{"sprint-1": 1, "sprint-2": 2, "sprint-3": 3},
	}
}

// US-023 / CA-023-1 / RN-023-7 / RN-023-8 / RN-023-11: con todo existente se verifica en el orden
// proyecto, defecto, sprint de resolucion; se traduce el sprint de deteccion a su numero; y se
// guarda una sola vez el defecto con el estado nuevo, condicionado al estado que se leyo. Devuelve
// el defecto ya guardado.
func TestCambiarEstadoDefecto_Resuelve(t *testing.T) {
	repo := repositorioConDefecto(defecto.EstadoAbierto, "")
	cambiar := app.NuevoCambiarEstadoDefecto(repo)

	d, err := cambiar.Ejecutar(context.Background(), "proyecto-1", "defecto-1", defecto.EstadoResuelto, "sprint-2")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	esperado := defectoGuardadoEn(defecto.EstadoResuelto, "sprint-2")
	if d != esperado {
		t.Errorf("defecto = %+v, quiero %+v", d, esperado)
	}
	if len(repo.guardado) != 1 || repo.guardado[0].defecto != esperado || repo.guardado[0].estadoAnterior != defecto.EstadoAbierto {
		t.Errorf("guardado = %+v, quiero un solo cambio del defecto resuelto, condicionado a abierto", repo.guardado)
	}
	if llamadas := strings.Join(repo.llamadas, ","); llamadas != "proyecto,defecto,sprint:sprint-2,sprint:sprint-1,actualizar" {
		t.Errorf("llamadas = %s, quiero proyecto, defecto, sprint de resolucion, sprint de deteccion, actualizar", llamadas)
	}
}

// US-023 / CA-023-10 / CA-023-11 / CA-023-12 / CL-023-10 / RN-023-7 / RN-023-10: se verifica en el
// orden proyecto, defecto, sprint de resolucion. Lo primero que no existe se devuelve solo,
// reconocible con errors.Is y con el ID recibido, sin validar los datos (el estado es invalido a
// proposito) y sin guardar nada.
func TestCambiarEstadoDefecto_LoQueNoExiste(t *testing.T) {
	casos := []struct {
		nombre   string
		repo     func() *repositorioCambioDeEstadoFalso
		defecto  string
		sprint   string
		esperado error
		idEnMsg  string
		llamadas string
	}{
		{"el proyecto no existe", func() *repositorioCambioDeEstadoFalso {
			r := repositorioConDefecto(defecto.EstadoAbierto, "")
			r.existeProyecto = false
			return r
		}, "defecto-1", "sprint-9", app.ErrProyectoInexistente, "proyecto-1", "proyecto"},
		{"el defecto no existe", func() *repositorioCambioDeEstadoFalso {
			return repositorioConDefecto(defecto.EstadoAbierto, "")
		}, "defecto-9", "sprint-9", app.ErrDefectoInexistente, "defecto-9", "proyecto,defecto"},
		{"el defecto tiene el ID vacio", func() *repositorioCambioDeEstadoFalso {
			return repositorioConDefecto(defecto.EstadoAbierto, "")
		}, "", "sprint-9", app.ErrDefectoInexistente, "proyecto-1", "proyecto,defecto"},
		{"el sprint de resolucion no existe o no es del proyecto", func() *repositorioCambioDeEstadoFalso {
			return repositorioConDefecto(defecto.EstadoAbierto, "")
		}, "defecto-1", "sprint-9", app.ErrSprintInexistente, "sprint-9", "proyecto,defecto,sprint:sprint-9"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := c.repo()

			_, err := app.NuevoCambiarEstadoDefecto(repo).Ejecutar(context.Background(), "proyecto-1", c.defecto, "urgente", c.sprint)

			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, quiero %v", err, c.esperado)
			}
			if errors.Is(err, defecto.ErrEstadoInvalido) {
				t.Errorf("el error %q se unio con errores de validacion", err)
			}
			if !strings.Contains(err.Error(), c.idEnMsg) {
				t.Errorf("el error %q no indica el ID %s", err, c.idEnMsg)
			}
			if strings.Join(repo.llamadas, ",") != c.llamadas {
				t.Errorf("llamadas = %v, quiero %s", repo.llamadas, c.llamadas)
			}
			if len(repo.guardado) != 0 {
				t.Errorf("se guardaron %d cambios, quiero ninguno", len(repo.guardado))
			}
		})
	}
}

// US-023 / CL-023-15 / RN-023-8: si el sprint de deteccion del defecto no esta (el defecto
// guardado apunta a uno que desaparecio), no hay numero para comparar: se devuelve
// ErrSprintInexistente y no se guarda nada, en vez de comparar contra un 0 inventado.
func TestCambiarEstadoDefecto_SprintDeDeteccionQueNoEsta(t *testing.T) {
	repo := repositorioConDefecto(defecto.EstadoAbierto, "")
	delete(repo.sprints, "sprint-1")
	repo.sprints["sprint-2"] = 2

	_, err := app.NuevoCambiarEstadoDefecto(repo).Ejecutar(context.Background(), "proyecto-1", "defecto-1", defecto.EstadoResuelto, "sprint-2")

	if !errors.Is(err, app.ErrSprintInexistente) {
		t.Fatalf("error = %v, quiero ErrSprintInexistente", err)
	}
	if !strings.Contains(err.Error(), "sprint-1") {
		t.Errorf("el error %q no indica el sprint de deteccion", err)
	}
	if len(repo.guardado) != 0 {
		t.Errorf("se guardaron %d cambios, quiero ninguno", len(repo.guardado))
	}
}

// US-023 / CA-023-13 / CL-023-9 / RN-023-6 / RN-023-8: los errores de validacion los devuelve el
// dominio tal cual, unidos y en orden, y los numeros de sprint se comparan ya traducidos: un
// defecto detectado en el Sprint 3 y resuelto de nuevo en el 2 devuelve transicion y sprint
// anterior. Un resuelto sin sprint no consulta el sprint de resolucion (ID vacio). No se guarda nada.
func TestCambiarEstadoDefecto_DatosInvalidos(t *testing.T) {
	casos := []struct {
		nombre    string
		repo      func() *repositorioCambioDeEstadoFalso
		estado    defecto.Estado
		sprint    string
		esperados []error
	}{
		{"descartado a resuelto sin sprint", func() *repositorioCambioDeEstadoFalso {
			return repositorioConDefecto(defecto.EstadoDescartado, "")
		}, defecto.EstadoResuelto, "", []error{defecto.ErrTransicionInvalida, defecto.ErrSprintResolucionFaltante}},
		{"resuelto a resuelto en un sprint anterior al de deteccion", func() *repositorioCambioDeEstadoFalso {
			r := repositorioConDefecto(defecto.EstadoResuelto, "sprint-3")
			d := r.defectos["defecto-1"]
			d.SprintDeteccionID = "sprint-3"
			r.defectos["defecto-1"] = d
			return r
		}, defecto.EstadoResuelto, "sprint-2", []error{defecto.ErrTransicionInvalida, defecto.ErrSprintResolucionAnterior}},
		{"sprint de resolucion con en_correccion", func() *repositorioCambioDeEstadoFalso {
			return repositorioConDefecto(defecto.EstadoAbierto, "")
		}, defecto.EstadoEnCorreccion, "sprint-2", []error{defecto.ErrSprintResolucionNoCorresponde}},
		{"estado invalido", func() *repositorioCambioDeEstadoFalso {
			return repositorioConDefecto(defecto.EstadoAbierto, "")
		}, "cerrado", "sprint-2", []error{defecto.ErrEstadoInvalido}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := c.repo()

			d, err := app.NuevoCambiarEstadoDefecto(repo).Ejecutar(context.Background(), "proyecto-1", "defecto-1", c.estado, c.sprint)

			for _, esperado := range c.esperados {
				if !errors.Is(err, esperado) {
					t.Errorf("error = %v, quiero que incluya %v", err, esperado)
				}
			}
			if d != (defecto.Defecto{}) {
				t.Errorf("devolvio el defecto %+v, quiero el valor cero", d)
			}
			if len(repo.guardado) != 0 {
				t.Errorf("se guardaron %d cambios, quiero ninguno", len(repo.guardado))
			}
		})
	}
}

// US-023 / CL-023-11 / RN-023-8: el estado actual sale del defecto guardado, no de lo que vio la
// persona. La pantalla mostraba el defecto abierto pero otra persona ya lo resolvio: pedir
// descartado se evalua contra resuelto y se rechaza ("de resuelto a descartado"); pedir abierto es
// una reapertura valida, que guarda el sprint vacio condicionada a resuelto (CL-023-7).
func TestCambiarEstadoDefecto_EvaluaContraLoGuardado(t *testing.T) {
	repo := repositorioConDefecto(defecto.EstadoResuelto, "sprint-2")
	cambiar := app.NuevoCambiarEstadoDefecto(repo)

	_, err := cambiar.Ejecutar(context.Background(), "proyecto-1", "defecto-1", defecto.EstadoDescartado, "")
	if !errors.Is(err, defecto.ErrTransicionInvalida) || !strings.Contains(err.Error(), "de resuelto a descartado") {
		t.Fatalf("error = %v, quiero ErrTransicionInvalida indicando de resuelto a descartado", err)
	}
	if len(repo.guardado) != 0 {
		t.Fatalf("se guardaron %d cambios, quiero ninguno", len(repo.guardado))
	}

	d, err := cambiar.Ejecutar(context.Background(), "proyecto-1", "defecto-1", defecto.EstadoAbierto, "")
	if err != nil {
		t.Fatalf("Ejecutar (reapertura): %v", err)
	}
	if d.Estado != defecto.EstadoAbierto || d.SprintResolucionID != "" {
		t.Errorf("defecto = %+v, quiero abierto y sin sprint de resolucion", d)
	}
	if len(repo.guardado) != 1 || repo.guardado[0].estadoAnterior != defecto.EstadoResuelto {
		t.Errorf("guardado = %+v, quiero un cambio condicionado a resuelto", repo.guardado)
	}
}

// US-023 / CA-023-14 / CL-023-14 / CL-023-15 / RN-023-11: lo que devuelve el repositorio al guardar
// (otro cambio se adelanto, el defecto o el sprint desaparecieron) se deja pasar reconocible con
// errors.Is. El caso de uso no reintenta.
func TestCambiarEstadoDefecto_ErroresAlGuardar(t *testing.T) {
	for _, e := range []error{app.ErrDefectoModificado, app.ErrDefectoInexistente, app.ErrSprintInexistente} {
		repo := repositorioConDefecto(defecto.EstadoAbierto, "")
		repo.errActualizar = e

		d, err := app.NuevoCambiarEstadoDefecto(repo).Ejecutar(context.Background(), "proyecto-1", "defecto-1", defecto.EstadoResuelto, "sprint-2")

		if !errors.Is(err, e) {
			t.Errorf("error = %v, quiero %v", err, e)
		}
		if d != (defecto.Defecto{}) {
			t.Errorf("devolvio el defecto %+v, quiero el valor cero", d)
		}
		if len(repo.guardado) != 1 {
			t.Errorf("intento guardar %d veces, quiero 1 (sin reintento)", len(repo.guardado))
		}
	}
}

// Los fallos del repositorio se devuelven envueltos con contexto y siguen reconociendose con
// errors.Is. Cada consulta tiene el suyo.
func TestCambiarEstadoDefecto_ErroresDelRepositorio(t *testing.T) {
	falla := errors.New("se corto la conexion")
	casos := []struct {
		nombre string
		repo   func() *repositorioCambioDeEstadoFalso
	}{
		{"al buscar el proyecto", func() *repositorioCambioDeEstadoFalso {
			r := repositorioConDefecto(defecto.EstadoAbierto, "")
			r.errProyecto = falla
			return r
		}},
		{"al buscar el defecto", func() *repositorioCambioDeEstadoFalso {
			r := repositorioConDefecto(defecto.EstadoAbierto, "")
			r.errDefecto = falla
			return r
		}},
		{"al buscar los sprints", func() *repositorioCambioDeEstadoFalso {
			r := repositorioConDefecto(defecto.EstadoAbierto, "")
			r.errSprint = falla
			return r
		}},
		{"al guardar", func() *repositorioCambioDeEstadoFalso {
			r := repositorioConDefecto(defecto.EstadoAbierto, "")
			r.errActualizar = falla
			return r
		}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := app.NuevoCambiarEstadoDefecto(c.repo()).Ejecutar(context.Background(), "proyecto-1", "defecto-1", defecto.EstadoResuelto, "sprint-2")

			if !errors.Is(err, falla) {
				t.Fatalf("error = %v, quiero el del repositorio", err)
			}
			if d != (defecto.Defecto{}) {
				t.Errorf("devolvio el defecto %+v, quiero el valor cero", d)
			}
		})
	}
}

// US-023 / CL-023-1 / CL-023-12 / RN-023-9: el cambio solo modifica el estado y el sprint de
// resolucion; la descripcion, la severidad, la historia y el sprint de deteccion quedan como
// estaban. En_correccion no consulta el sprint de resolucion: no se informo.
func TestCambiarEstadoDefecto_SoloCambiaEstadoYSprintDeResolucion(t *testing.T) {
	repo := repositorioConDefecto(defecto.EstadoAbierto, "")

	d, err := app.NuevoCambiarEstadoDefecto(repo).Ejecutar(context.Background(), "proyecto-1", "defecto-1", defecto.EstadoEnCorreccion, "")
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	esperado := defectoGuardadoEn(defecto.EstadoEnCorreccion, "")
	if d != esperado {
		t.Errorf("defecto = %+v, quiero %+v", d, esperado)
	}
	if llamadas := strings.Join(repo.llamadas, ","); llamadas != "proyecto,defecto,sprint:sprint-1,actualizar" {
		t.Errorf("llamadas = %s, quiero proyecto, defecto, sprint de deteccion, actualizar", llamadas)
	}
}
