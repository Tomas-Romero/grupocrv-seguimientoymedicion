package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
)

// hoyDelEsfuerzo es la fecha del reloj fijo de estos tests.
var hoyDelEsfuerzo = time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)

func relojDelEsfuerzo() time.Time { return hoyDelEsfuerzo }

// repositorioEsfuerzoFalso simula la persistencia del esfuerzo sin base de datos.
// Responde lo que se le configura y anota el orden en que le preguntan.
type repositorioEsfuerzoFalso struct {
	existeProyecto    bool
	existeItem        bool
	existeIntegrante  bool
	errProyecto       error
	errItem           error
	errIntegrante     error
	errRegistrar      error
	llamadas          []string
	registrados       []esfuerzo.RegistroEsfuerzo
	itemConsultado    [2]string // proyecto e item
	integranteConsult [2]string // proyecto e integrante
}

func (r *repositorioEsfuerzoFalso) ExisteProyecto(_ context.Context, _ string) (bool, error) {
	r.llamadas = append(r.llamadas, "proyecto")
	return r.existeProyecto, r.errProyecto
}

func (r *repositorioEsfuerzoFalso) ExisteItemEnProyecto(_ context.Context, proyectoID, itemID string) (bool, error) {
	r.llamadas = append(r.llamadas, "item")
	r.itemConsultado = [2]string{proyectoID, itemID}
	return r.existeItem, r.errItem
}

func (r *repositorioEsfuerzoFalso) ExisteIntegranteEnProyecto(_ context.Context, proyectoID, integranteID string) (bool, error) {
	r.llamadas = append(r.llamadas, "integrante")
	r.integranteConsult = [2]string{proyectoID, integranteID}
	return r.existeIntegrante, r.errIntegrante
}

func (r *repositorioEsfuerzoFalso) RegistrarEsfuerzo(_ context.Context, registro esfuerzo.RegistroEsfuerzo) (esfuerzo.RegistroEsfuerzo, error) {
	r.llamadas = append(r.llamadas, "registrar")
	r.registrados = append(r.registrados, registro)
	if r.errRegistrar != nil {
		return esfuerzo.RegistroEsfuerzo{}, r.errRegistrar
	}
	registro.ID = "id-del-repositorio"
	return registro, nil
}

// todoExiste es un repositorio donde el proyecto, el item y el integrante existen.
func todoExiste() *repositorioEsfuerzoFalso {
	return &repositorioEsfuerzoFalso{existeProyecto: true, existeItem: true, existeIntegrante: true}
}

func ayer() time.Time { return time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC) }

// US-019 / CA-019-1, RN-019-6, RN-019-9: con todo existente y datos validos, el
// caso de uso verifica en orden (proyecto, historia, integrante), valida con el
// dominio y registra una sola vez. Devuelve el registro con el ID del repositorio.
func TestRegistrarEsfuerzo_Registra(t *testing.T) {
	repo := todoExiste()
	registrar := app.NuevoRegistrarEsfuerzo(repo, relojDelEsfuerzo)

	r, err := registrar.Ejecutar(context.Background(), "proyecto-1", "item-1", "integrante-1", ayer(), "Desarrollo", 2.5)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if r.ID != "id-del-repositorio" || r.ProyectoID != "proyecto-1" || r.ItemID != "item-1" || r.IntegranteID != "integrante-1" {
		t.Errorf("registro = %+v, se esperaba el del repositorio con los IDs recibidos", r)
	}
	if r.Horas != 2.5 || r.Actividad != "Desarrollo" || !r.Fecha.Equal(ayer()) {
		t.Errorf("registro = %+v, se esperaban 2,5 horas de Desarrollo de ayer", r)
	}
	if esperado := "proyecto,item,integrante,registrar"; strings.Join(repo.llamadas, ",") != esperado {
		t.Errorf("orden de llamadas = %v, se esperaba %s", repo.llamadas, esperado)
	}
	if repo.itemConsultado != [2]string{"proyecto-1", "item-1"} {
		t.Errorf("la historia se busco con %v: tiene que buscarse dentro del proyecto", repo.itemConsultado)
	}
	if repo.integranteConsult != [2]string{"proyecto-1", "integrante-1"} {
		t.Errorf("el integrante se busco con %v: tiene que buscarse dentro del proyecto", repo.integranteConsult)
	}
}

// US-019 / RN-019-3: la fecha de hoy la da el reloj inyectado, no el del sistema.
func TestRegistrarEsfuerzo_UsaElRelojInyectado(t *testing.T) {
	repo := todoExiste()
	registrar := app.NuevoRegistrarEsfuerzo(repo, relojDelEsfuerzo)

	manana := hoyDelEsfuerzo.AddDate(0, 0, 1)
	_, err := registrar.Ejecutar(context.Background(), "proyecto-1", "item-1", "integrante-1", manana, "Desarrollo", 1)

	if !errors.Is(err, esfuerzo.ErrFechaFutura) {
		t.Fatalf("error = %v, se esperaba ErrFechaFutura", err)
	}
	if len(repo.registrados) != 0 {
		t.Errorf("se registraron %d esfuerzos, se esperaba ninguno", len(repo.registrados))
	}
}

// US-019 / RN-019-6, CL-019-15, CA-019-9, CA-019-10: si algo no existe se devuelve
// ese error, reconocible con errors.Is y con el ID recibido, sin validar los datos
// (que aca son invalidos a proposito) y sin registrar nada. Se verifica en el
// orden proyecto, historia, integrante: el primero que falla gana.
func TestRegistrarEsfuerzo_LoQueNoExiste(t *testing.T) {
	casos := []struct {
		nombre   string
		repo     *repositorioEsfuerzoFalso
		esperado error
		idEnMsg  string
		llamadas string
	}{
		{"el proyecto no existe", &repositorioEsfuerzoFalso{}, app.ErrProyectoInexistente, "proyecto-1", "proyecto"},
		{"la historia no existe o no es del proyecto", &repositorioEsfuerzoFalso{existeProyecto: true}, app.ErrItemInexistente, "item-1", "proyecto,item"},
		{"el integrante no existe o no es del proyecto", &repositorioEsfuerzoFalso{existeProyecto: true, existeItem: true}, app.ErrIntegranteInexistente, "integrante-1", "proyecto,item,integrante"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			registrar := app.NuevoRegistrarEsfuerzo(c.repo, relojDelEsfuerzo)

			// Datos invalidos (todo mal) para comprobar que no se miran.
			_, err := registrar.Ejecutar(context.Background(), "proyecto-1", "item-1", "integrante-1", time.Time{}, "", 0)

			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.esperado)
			}
			if errors.Is(err, esfuerzo.ErrFechaVacia) || errors.Is(err, esfuerzo.ErrActividadVacia) || errors.Is(err, esfuerzo.ErrHorasFueraDeRango) {
				t.Errorf("el error %q se unio con errores de validacion", err)
			}
			if !strings.Contains(err.Error(), c.idEnMsg) {
				t.Errorf("el error %q no indica el ID %s", err, c.idEnMsg)
			}
			if strings.Join(c.repo.llamadas, ",") != c.llamadas {
				t.Errorf("llamadas = %v, se esperaba %s", c.repo.llamadas, c.llamadas)
			}
		})
	}
}

// US-019 / RN-019-8, RN-019-11, CA-019-6 a CA-019-8: con datos invalidos se
// devuelven los errores del dominio tal cual, unidos, y no se registra nada.
func TestRegistrarEsfuerzo_DatosInvalidos(t *testing.T) {
	repo := todoExiste()
	registrar := app.NuevoRegistrarEsfuerzo(repo, relojDelEsfuerzo)

	_, err := registrar.Ejecutar(context.Background(), "proyecto-1", "item-1", "integrante-1", ayer(), "  ", -1)

	for _, e := range []error{esfuerzo.ErrActividadVacia, esfuerzo.ErrHorasFueraDeRango} {
		if !errors.Is(err, e) {
			t.Errorf("error = %v, se esperaba que incluyera %v", err, e)
		}
	}
	if len(repo.registrados) != 0 {
		t.Errorf("se registraron %d esfuerzos, se esperaba ninguno", len(repo.registrados))
	}
}

// Los fallos del repositorio se devuelven envueltos con contexto y siguen
// reconociendose con errors.Is. Cada consulta tiene el suyo.
func TestRegistrarEsfuerzo_ErroresDelRepositorio(t *testing.T) {
	falla := errors.New("se corto la conexion")
	casos := []struct {
		nombre string
		repo   *repositorioEsfuerzoFalso
	}{
		{"al buscar el proyecto", &repositorioEsfuerzoFalso{errProyecto: falla}},
		{"al buscar la historia", &repositorioEsfuerzoFalso{existeProyecto: true, errItem: falla}},
		{"al buscar el integrante", &repositorioEsfuerzoFalso{existeProyecto: true, existeItem: true, errIntegrante: falla}},
		{"al registrar", &repositorioEsfuerzoFalso{existeProyecto: true, existeItem: true, existeIntegrante: true, errRegistrar: falla}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			registrar := app.NuevoRegistrarEsfuerzo(c.repo, relojDelEsfuerzo)

			r, err := registrar.Ejecutar(context.Background(), "proyecto-1", "item-1", "integrante-1", ayer(), "Desarrollo", 1)

			if !errors.Is(err, falla) {
				t.Fatalf("error = %v, se esperaba el del repositorio", err)
			}
			if !strings.Contains(err.Error(), "proyecto-1") {
				t.Errorf("el error %q no indica el proyecto", err)
			}
			if r.ID != "" {
				t.Errorf("devolvio el registro %+v, se esperaba el valor cero", r)
			}
		})
	}
}

// Si el item, el integrante o el proyecto desaparecen entre la lectura y el
// INSERT, el repositorio devuelve el error de inexistencia y el caso de uso lo
// deja pasar reconocible.
func TestRegistrarEsfuerzo_LoQueDesapareceEnElMedio(t *testing.T) {
	for _, e := range []error{app.ErrProyectoInexistente, app.ErrItemInexistente, app.ErrIntegranteInexistente} {
		repo := todoExiste()
		repo.errRegistrar = e

		_, err := app.NuevoRegistrarEsfuerzo(repo, relojDelEsfuerzo).Ejecutar(context.Background(), "proyecto-1", "item-1", "integrante-1", ayer(), "Desarrollo", 1)

		if !errors.Is(err, e) {
			t.Errorf("error = %v, se esperaba %v", err, e)
		}
	}
}

// US-019 / CL-019-11, CL-019-12: dos registros del mismo integrante, historia y
// dia son dos registros: no se rechazan por duplicados ni por sumar mas de 24.
func TestRegistrarEsfuerzo_RegistrosRepetidosSonValidos(t *testing.T) {
	repo := todoExiste()
	registrar := app.NuevoRegistrarEsfuerzo(repo, relojDelEsfuerzo)

	for i := 0; i < 2; i++ {
		if _, err := registrar.Ejecutar(context.Background(), "proyecto-1", "item-1", "integrante-1", ayer(), "Desarrollo", 20); err != nil {
			t.Fatalf("registro %d: %v", i+1, err)
		}
	}
	if len(repo.registrados) != 2 {
		t.Errorf("se registraron %d esfuerzos, se esperaban 2", len(repo.registrados))
	}
}
