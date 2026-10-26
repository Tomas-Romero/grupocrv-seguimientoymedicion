package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// repositorioDefectosFalso simula la persistencia de los defectos sin base de datos.
// Responde lo que se le configura y anota el orden en que le preguntan.
type repositorioDefectosFalso struct {
	existeProyecto bool
	existeItem     bool
	existeSprint   bool
	errProyecto    error
	errItem        error
	errSprint      error
	errRegistrar   error
	llamadas       []string
	registrados    []defecto.Defecto
	itemConsultado [2]string // proyecto e item
	sprintConsult  [2]string // proyecto y sprint
}

func (r *repositorioDefectosFalso) ExisteProyecto(_ context.Context, _ string) (bool, error) {
	r.llamadas = append(r.llamadas, "proyecto")
	return r.existeProyecto, r.errProyecto
}

func (r *repositorioDefectosFalso) ExisteItemEnProyecto(_ context.Context, proyectoID, itemID string) (bool, error) {
	r.llamadas = append(r.llamadas, "item")
	r.itemConsultado = [2]string{proyectoID, itemID}
	return r.existeItem, r.errItem
}

func (r *repositorioDefectosFalso) ExisteSprintEnProyecto(_ context.Context, proyectoID, sprintID string) (bool, error) {
	r.llamadas = append(r.llamadas, "sprint")
	r.sprintConsult = [2]string{proyectoID, sprintID}
	return r.existeSprint, r.errSprint
}

func (r *repositorioDefectosFalso) RegistrarDefecto(_ context.Context, d defecto.Defecto) (defecto.Defecto, error) {
	r.llamadas = append(r.llamadas, "registrar")
	r.registrados = append(r.registrados, d)
	if r.errRegistrar != nil {
		return defecto.Defecto{}, r.errRegistrar
	}
	d.ID = "id-del-repositorio"
	return d, nil
}

// todoExisteParaDefectos es un repositorio donde el proyecto, la historia y el sprint existen.
func todoExisteParaDefectos() *repositorioDefectosFalso {
	return &repositorioDefectosFalso{existeProyecto: true, existeItem: true, existeSprint: true}
}

// US-022 / CA-022-1 / RN-022-3 / RN-022-10: con todo existente y datos validos se registra
// una sola vez un defecto abierto y sin sprint de resolucion, que todavia no tiene ID, y se
// devuelve el defecto con el ID que le dio el repositorio.
func TestRegistrarDefecto_Registra(t *testing.T) {
	repo := todoExisteParaDefectos()
	registrar := app.NuevoRegistrarDefecto(repo)

	d, err := registrar.Ejecutar(context.Background(), "proyecto-1", "item-1", "sprint-1", "  Falla al guardar  ", defecto.SeveridadAlta)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	esperado := defecto.Defecto{
		ID:                "id-del-repositorio",
		ProyectoID:        "proyecto-1",
		ItemID:            "item-1",
		SprintDeteccionID: "sprint-1",
		Descripcion:       "Falla al guardar",
		Severidad:         defecto.SeveridadAlta,
		Estado:            defecto.EstadoAbierto,
	}
	if d != esperado {
		t.Errorf("defecto = %+v, quiero %+v", d, esperado)
	}
	if len(repo.registrados) != 1 || repo.registrados[0].ID != "" {
		t.Errorf("registrados = %+v, quiero uno solo y sin ID: lo genera la base", repo.registrados)
	}
}

// US-022 / RN-022-5 / RN-022-9 / CL-022-7 / CL-022-8 / CL-022-9 / CA-022-9 / CA-022-10 / CA-022-11: se verifica en el orden
// proyecto, historia, sprint de deteccion. Lo primero que no existe se devuelve solo,
// reconocible con errors.Is y con el ID recibido, sin validar los datos (que aca son
// invalidos a proposito) y sin registrar nada (RN-022-9).
func TestRegistrarDefecto_LoQueNoExiste(t *testing.T) {
	casos := []struct {
		nombre   string
		repo     *repositorioDefectosFalso
		esperado error
		idEnMsg  string
		llamadas string
	}{
		{"el proyecto no existe", &repositorioDefectosFalso{}, app.ErrProyectoInexistente, "proyecto-1", "proyecto"},
		{"la historia no existe o no es del proyecto", &repositorioDefectosFalso{existeProyecto: true}, app.ErrItemInexistente, "item-1", "proyecto,item"},
		{"el sprint no existe o no es del proyecto", &repositorioDefectosFalso{existeProyecto: true, existeItem: true}, app.ErrSprintInexistente, "sprint-1", "proyecto,item,sprint"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			registrar := app.NuevoRegistrarDefecto(c.repo)

			// Datos invalidos (descripcion y severidad) para comprobar que no se miran.
			_, err := registrar.Ejecutar(context.Background(), "proyecto-1", "item-1", "sprint-1", "", "urgente")

			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, quiero %v", err, c.esperado)
			}
			if errors.Is(err, defecto.ErrDescripcionVacia) || errors.Is(err, defecto.ErrSeveridadInvalida) {
				t.Errorf("el error %q se unio con errores de validacion", err)
			}
			if !strings.Contains(err.Error(), c.idEnMsg) {
				t.Errorf("el error %q no indica el ID %s", err, c.idEnMsg)
			}
			if strings.Join(c.repo.llamadas, ",") != c.llamadas {
				t.Errorf("llamadas = %v, quiero %s", c.repo.llamadas, c.llamadas)
			}
			if len(c.repo.registrados) != 0 {
				t.Errorf("se registraron %d defectos, quiero ninguno", len(c.repo.registrados))
			}
		})
	}
}

// US-022 / RN-022-5: la historia y el sprint se buscan DENTRO del proyecto, y el orden de
// las llamadas es el de la spec, con el registro al final.
func TestRegistrarDefecto_VerificaDentroDelProyecto(t *testing.T) {
	repo := todoExisteParaDefectos()

	_, err := app.NuevoRegistrarDefecto(repo).Ejecutar(context.Background(), "proyecto-1", "item-1", "sprint-1", "Falla", defecto.SeveridadBaja)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if esperado := "proyecto,item,sprint,registrar"; strings.Join(repo.llamadas, ",") != esperado {
		t.Errorf("orden de llamadas = %v, quiero %s", repo.llamadas, esperado)
	}
	if repo.itemConsultado != [2]string{"proyecto-1", "item-1"} {
		t.Errorf("la historia se busco con %v: tiene que buscarse dentro del proyecto", repo.itemConsultado)
	}
	if repo.sprintConsult != [2]string{"proyecto-1", "sprint-1"} {
		t.Errorf("el sprint se busco con %v: tiene que buscarse dentro del proyecto", repo.sprintConsult)
	}
}

// US-022 / RN-022-4 / RN-022-9 / CA-022-7, CA-022-8, CA-022-12 / CL-022-6: con datos
// invalidos se devuelven los errores del dominio tal cual, unidos y en orden, y no se
// registra nada.
func TestRegistrarDefecto_DatosInvalidos(t *testing.T) {
	repo := todoExisteParaDefectos()
	registrar := app.NuevoRegistrarDefecto(repo)

	_, err := registrar.Ejecutar(context.Background(), "proyecto-1", "item-1", "sprint-1", "   ", "urgente")

	for _, e := range []error{defecto.ErrDescripcionVacia, defecto.ErrSeveridadInvalida} {
		if !errors.Is(err, e) {
			t.Errorf("error = %v, quiero que incluya %v", err, e)
		}
	}
	if len(repo.registrados) != 0 {
		t.Errorf("se registraron %d defectos, quiero ninguno", len(repo.registrados))
	}
}

// Los fallos del repositorio se devuelven envueltos con contexto y siguen reconociendose
// con errors.Is. Cada consulta tiene el suyo.
func TestRegistrarDefecto_ErroresDelRepositorio(t *testing.T) {
	falla := errors.New("se corto la conexion")
	casos := []struct {
		nombre string
		repo   *repositorioDefectosFalso
	}{
		{"al buscar el proyecto", &repositorioDefectosFalso{errProyecto: falla}},
		{"al buscar la historia", &repositorioDefectosFalso{existeProyecto: true, errItem: falla}},
		{"al buscar el sprint", &repositorioDefectosFalso{existeProyecto: true, existeItem: true, errSprint: falla}},
		{"al registrar", &repositorioDefectosFalso{existeProyecto: true, existeItem: true, existeSprint: true, errRegistrar: falla}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := app.NuevoRegistrarDefecto(c.repo).Ejecutar(context.Background(), "proyecto-1", "item-1", "sprint-1", "Falla", defecto.SeveridadMedia)

			if !errors.Is(err, falla) {
				t.Fatalf("error = %v, quiero el del repositorio", err)
			}
			if !strings.Contains(err.Error(), "proyecto-1") {
				t.Errorf("el error %q no indica el proyecto", err)
			}
			if d != (defecto.Defecto{}) {
				t.Errorf("devolvio el defecto %+v, quiero el valor cero", d)
			}
		})
	}
}

// US-022 / RN-022-10 / CL-022-15: si el proyecto, la historia o el sprint desaparecen entre
// la lectura y el INSERT, el repositorio devuelve el error de inexistencia y el caso de uso
// lo deja pasar reconocible.
func TestRegistrarDefecto_LoQueDesapareceEnElMedio(t *testing.T) {
	for _, e := range []error{app.ErrProyectoInexistente, app.ErrItemInexistente, app.ErrSprintInexistente} {
		repo := todoExisteParaDefectos()
		repo.errRegistrar = e

		_, err := app.NuevoRegistrarDefecto(repo).Ejecutar(context.Background(), "proyecto-1", "item-1", "sprint-1", "Falla", defecto.SeveridadMedia)

		if !errors.Is(err, e) {
			t.Errorf("error = %v, quiero %v", err, e)
		}
	}
}

// US-022 / RN-022-8 / CL-022-13 / CA-022-5: dos defectos identicos son dos defectos: el caso
// de uso no rechaza repetidos.
func TestRegistrarDefecto_RepetidosSonValidos(t *testing.T) {
	repo := todoExisteParaDefectos()
	registrar := app.NuevoRegistrarDefecto(repo)

	for i := 0; i < 2; i++ {
		if _, err := registrar.Ejecutar(context.Background(), "proyecto-1", "item-1", "sprint-1", "Falla", defecto.SeveridadAlta); err != nil {
			t.Fatalf("defecto %d: %v", i+1, err)
		}
	}
	if len(repo.registrados) != 2 {
		t.Errorf("se registraron %d defectos, quiero 2", len(repo.registrados))
	}
}
