package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

func diaDelSprint(mes time.Month, d int) time.Time {
	return time.Date(2026, mes, d, 0, 0, 0, 0, time.UTC)
}

// repositorioSprintsFalso simula la persistencia de los sprints sin base de datos.
// Responde lo que se le configura y anota lo que le piden.
type repositorioSprintsFalso struct {
	existe       bool
	inicio, fin  time.Time
	errFechas    error
	errRegistrar error

	consultas []string
	altas     []sprint.Sprint
}

func (r *repositorioSprintsFalso) FechasDelProyecto(_ context.Context, proyectoID string) (time.Time, time.Time, bool, error) {
	r.consultas = append(r.consultas, proyectoID)
	return r.inicio, r.fin, r.existe, r.errFechas
}

func (r *repositorioSprintsFalso) RegistrarSprint(_ context.Context, s sprint.Sprint) (sprint.Sprint, error) {
	r.altas = append(r.altas, s)
	if r.errRegistrar != nil {
		return sprint.Sprint{}, r.errRegistrar
	}
	s.ID = "id-del-repositorio"
	s.Numero = len(r.altas)
	return s, nil
}

// proyectoDelSprint es un repositorio con un proyecto del 2026-10-06 al 2026-12-31.
func proyectoDelSprint() *repositorioSprintsFalso {
	return &repositorioSprintsFalso{existe: true, inicio: diaDelSprint(10, 6), fin: diaDelSprint(12, 31)}
}

// US-008 / CA-008-1: el caso de uso lee las fechas del proyecto, valida con el
// dominio y registra una vez. Devuelve el sprint con el ID, el numero y el nombre
// que asigna el repositorio.
func TestCrearSprint_Registra(t *testing.T) {
	repo := proyectoDelSprint()

	s, err := app.NuevoCrearSprint(repo).Ejecutar(context.Background(), "proyecto-1", "  MVP navegable ", diaDelSprint(10, 6), diaDelSprint(10, 19))
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if s.ID != "id-del-repositorio" || s.Numero != 1 || s.Nombre() != "Sprint 1" {
		t.Errorf("sprint = %+v, se esperaban el ID y el numero del repositorio (Sprint 1)", s)
	}
	if s.ProyectoID != "proyecto-1" || s.SprintGoal != "MVP navegable" || s.Estado != sprint.EstadoAbierto {
		t.Errorf("sprint = %+v, se esperaba abierto, del proyecto recibido y con el Sprint Goal recortado", s)
	}
	if len(repo.consultas) != 1 || repo.consultas[0] != "proyecto-1" {
		t.Errorf("consultas = %v, se esperaba leer el proyecto una vez", repo.consultas)
	}
	if len(repo.altas) != 1 {
		t.Errorf("altas = %d, se esperaba 1", len(repo.altas))
	}
}

// US-008 / CA-008-10, CL-008-14, RN-008-9: si el proyecto no existe responde
// ErrProyectoInexistente sin mirar los datos (aca todos invalidos), sin unirlo
// con los errores de validacion y sin registrar nada.
func TestCrearSprint_ProyectoInexistente(t *testing.T) {
	repo := &repositorioSprintsFalso{existe: false}

	_, err := app.NuevoCrearSprint(repo).Ejecutar(context.Background(), "no-existe", "", time.Time{}, time.Time{})

	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Fatalf("error = %v, se esperaba ErrProyectoInexistente", err)
	}
	if errors.Is(err, sprint.ErrSprintGoalVacio) || errors.Is(err, sprint.ErrFechaInicioFaltante) {
		t.Errorf("el error %q se unio con errores de validacion", err)
	}
	if !strings.Contains(err.Error(), "no-existe") {
		t.Errorf("el error %q no indica el ID recibido", err)
	}
	if len(repo.altas) != 0 {
		t.Errorf("se registraron %d sprints, se esperaba ninguno", len(repo.altas))
	}
}

// US-008 / CA-008-8, RN-008-7: el dominio recibe las fechas del proyecto que leyo
// el caso de uso: un sprint fuera de ellas se rechaza sin registrar nada.
func TestCrearSprint_UsaLasFechasDelProyecto(t *testing.T) {
	repo := proyectoDelSprint()

	_, err := app.NuevoCrearSprint(repo).Ejecutar(context.Background(), "proyecto-1", "Objetivo", diaDelSprint(10, 5), diaDelSprint(10, 12))

	if !errors.Is(err, sprint.ErrFueraDelProyecto) {
		t.Fatalf("error = %v, se esperaba ErrFueraDelProyecto", err)
	}
	if len(repo.altas) != 0 {
		t.Errorf("se registraron %d sprints, se esperaba ninguno", len(repo.altas))
	}
}

// US-008 / RN-008-10, RN-008-15, CL-008-13: con datos invalidos se devuelven los
// errores del dominio tal cual, unidos, y el repositorio no recibe el alta.
func TestCrearSprint_DatosInvalidos(t *testing.T) {
	repo := proyectoDelSprint()

	_, err := app.NuevoCrearSprint(repo).Ejecutar(context.Background(), "proyecto-1", "   ", diaDelSprint(10, 19), diaDelSprint(10, 6))

	for _, e := range []error{sprint.ErrSprintGoalVacio, sprint.ErrFechasIncoherentes} {
		if !errors.Is(err, e) {
			t.Errorf("error = %v, se esperaba que incluyera %v", err, e)
		}
	}
	if len(repo.altas) != 0 {
		t.Errorf("se registraron %d sprints, se esperaba ninguno", len(repo.altas))
	}
}

// Los errores que devuelve el repositorio dentro de la transaccion (el ultimo
// sprint abierto, el inicio no posterior, las fechas del proyecto que cambiaron o
// el proyecto que desaparecio) llegan al que llama, reconocibles con errors.Is y
// sin unirse con otros.
func TestCrearSprint_ErroresDelRepositorio(t *testing.T) {
	for _, e := range []error{
		sprint.ErrUltimoSprintAbierto, sprint.ErrInicioNoPosteriorAlUltimo, sprint.ErrFueraDelProyecto, app.ErrProyectoInexistente,
	} {
		repo := proyectoDelSprint()
		repo.errRegistrar = e

		_, err := app.NuevoCrearSprint(repo).Ejecutar(context.Background(), "proyecto-1", "Objetivo", diaDelSprint(10, 6), diaDelSprint(10, 19))

		if !errors.Is(err, e) {
			t.Errorf("error = %v, se esperaba %v", err, e)
		}
	}
}

// Un fallo de lectura o de escritura del repositorio se devuelve envuelto, con el
// proyecto, y sigue reconociendose con errors.Is. No devuelve sprint.
func TestCrearSprint_FallosDeInfraestructura(t *testing.T) {
	falla := errors.New("se corto la conexion")
	casos := []struct {
		nombre string
		repo   *repositorioSprintsFalso
	}{
		{"al leer el proyecto", &repositorioSprintsFalso{errFechas: falla}},
		{"al registrar", func() *repositorioSprintsFalso { r := proyectoDelSprint(); r.errRegistrar = falla; return r }()},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			s, err := app.NuevoCrearSprint(c.repo).Ejecutar(context.Background(), "proyecto-1", "Objetivo", diaDelSprint(10, 6), diaDelSprint(10, 19))

			if !errors.Is(err, falla) {
				t.Fatalf("error = %v, se esperaba el del repositorio", err)
			}
			if !strings.Contains(err.Error(), "proyecto-1") {
				t.Errorf("el error %q no indica el proyecto", err)
			}
			if s.ID != "" {
				t.Errorf("devolvio el sprint %+v, se esperaba el valor cero", s)
			}
		})
	}
}
