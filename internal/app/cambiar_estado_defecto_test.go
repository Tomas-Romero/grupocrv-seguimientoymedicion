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
