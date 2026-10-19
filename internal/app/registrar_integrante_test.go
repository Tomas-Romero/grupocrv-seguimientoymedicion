package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
)

// repositorioIntegrantesFalso simula el repositorio de integrantes sin base de
// datos. Responde lo que se le configura y anota lo que le piden.
type repositorioIntegrantesFalso struct {
	existe       bool
	errExiste    error
	errRegistrar error

	consultas []string
	altas     []integrante.Integrante
}

func (r *repositorioIntegrantesFalso) ExisteProyecto(_ context.Context, proyectoID string) (bool, error) {
	r.consultas = append(r.consultas, proyectoID)
	return r.existe, r.errExiste
}

func (r *repositorioIntegrantesFalso) RegistrarIntegrante(_ context.Context, i integrante.Integrante) (integrante.Integrante, error) {
	r.altas = append(r.altas, i)
	if r.errRegistrar != nil {
		return integrante.Integrante{}, r.errRegistrar
	}
	i.ID = "id-del-repositorio"
	return i, nil
}

// US-003 / CA-003-1, CA-003-4: con un proyecto existente y datos validos, el caso
// de uso le pasa al repositorio el integrante que armo el dominio (nombre y email
// recortados, email en minuscula, sin ID) y devuelve el que registro, con el ID
// que asigna la persistencia.
func TestRegistrarIntegrante_Registra(t *testing.T) {
	repo := &repositorioIntegrantesFalso{existe: true}
	registrar := app.NuevoRegistrarIntegrante(repo)

	registrado, err := registrar.Ejecutar(context.Background(), "proyecto-1", "  Tomás ", " Tomas@Ejemplo.Test ", integrante.RolProductBuilder)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if len(repo.altas) != 1 {
		t.Fatalf("se registraron %d integrantes, se esperaba 1", len(repo.altas))
	}
	esperado := integrante.Integrante{
		ProyectoID: "proyecto-1",
		Nombre:     "Tomás",
		Email:      "tomas@ejemplo.test",
		Rol:        integrante.RolProductBuilder,
	}
	if repo.altas[0] != esperado {
		t.Errorf("alta en el repositorio = %+v, se esperaba %+v", repo.altas[0], esperado)
	}
	if registrado.ID != "id-del-repositorio" {
		t.Errorf("ID = %q, se esperaba el del repositorio", registrado.ID)
	}
	if registrado.ProyectoID != "proyecto-1" || registrado.Email != "tomas@ejemplo.test" {
		t.Errorf("integrante devuelto = %+v, se esperaban los datos normalizados", registrado)
	}
}

// US-003 / CA-003-9, CL-003-13, RN-003-6: si el proyecto no existe, el caso de
// uso responde ErrProyectoInexistente antes de validar los datos: no lo une con
// los errores de validacion (aunque los datos tambien sean invalidos) y no
// registra nada.
func TestRegistrarIntegrante_ProyectoInexistente(t *testing.T) {
	repo := &repositorioIntegrantesFalso{existe: false}
	registrar := app.NuevoRegistrarIntegrante(repo)

	_, err := registrar.Ejecutar(context.Background(), "proyecto-inexistente", "   ", "sin-arroba", "jefe")
	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Fatalf("error = %v, se esperaba ErrProyectoInexistente", err)
	}
	for _, validacion := range []error{integrante.ErrNombreVacio, integrante.ErrEmailInvalido, integrante.ErrRolInvalido} {
		if errors.Is(err, validacion) {
			t.Errorf("el proyecto inexistente vino unido con un error de validacion (%v): %v", validacion, err)
		}
	}
	// Envuelto indicando el ID recibido.
	if !strings.Contains(err.Error(), "proyecto-inexistente") {
		t.Errorf("el mensaje %q no indica el ID del proyecto", err.Error())
	}
	if len(repo.consultas) != 1 || repo.consultas[0] != "proyecto-inexistente" {
		t.Errorf("consultas al repositorio = %q, se esperaba una por %q", repo.consultas, "proyecto-inexistente")
	}
	if len(repo.altas) != 0 {
		t.Errorf("se registraron %d integrantes, se esperaba ninguno", len(repo.altas))
	}
}

// US-003 / CA-003-5 a CA-003-7, CL-003-11, RN-003-9: con datos invalidos no se
// registra nada, y los errores del dominio se devuelven unidos y tal cual, sin
// prefijos: su mensaje es el que ve la persona (spec, seccion 7).
func TestRegistrarIntegrante_DatosInvalidos(t *testing.T) {
	repo := &repositorioIntegrantesFalso{existe: true}
	registrar := app.NuevoRegistrarIntegrante(repo)

	_, err := registrar.Ejecutar(context.Background(), "proyecto-1", "   ", "sin-arroba", "jefe")
	if err == nil {
		t.Fatal("se esperaba un error y no hubo ninguno")
	}
	for _, esperado := range []error{integrante.ErrNombreVacio, integrante.ErrEmailInvalido, integrante.ErrRolInvalido} {
		if !errors.Is(err, esperado) {
			t.Errorf("errors.Is no reconoce %v dentro de %q", esperado, err)
		}
	}
	if errors.Is(err, app.ErrProyectoInexistente) {
		t.Errorf("un dato invalido no es un proyecto inexistente: %v", err)
	}
	if !strings.HasPrefix(err.Error(), integrante.ErrNombreVacio.Error()) {
		t.Errorf("mensaje = %q, se esperaba el del dominio sin prefijos", err.Error())
	}
	if len(repo.altas) != 0 {
		t.Errorf("se registraron %d integrantes, se esperaba ninguno", len(repo.altas))
	}
}

// US-003 / CA-003-8, RN-003-7, RN-003-9: los errores del repositorio se devuelven
// envueltos con el proyecto, sin perder el original. Si el proyecto desaparece
// antes del alta, o el email ya esta registrado, el caso de uso deja el error
// reconocible.
func TestRegistrarIntegrante_ErroresDelRepositorio(t *testing.T) {
	errBase := errors.New("la base no responde")
	casos := []struct {
		nombre    string
		repo      *repositorioIntegrantesFalso
		esperado  error
		registros int
	}{
		{"falla la consulta del proyecto", &repositorioIntegrantesFalso{errExiste: errBase}, errBase, 0},
		{"el proyecto desaparece antes del alta", &repositorioIntegrantesFalso{existe: true, errRegistrar: app.ErrProyectoInexistente}, app.ErrProyectoInexistente, 1},
		{"el email ya esta registrado en el proyecto", &repositorioIntegrantesFalso{existe: true, errRegistrar: app.ErrEmailDuplicado}, app.ErrEmailDuplicado, 1},
		{"falla el alta", &repositorioIntegrantesFalso{existe: true, errRegistrar: errBase}, errBase, 1},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			registrar := app.NuevoRegistrarIntegrante(c.repo)

			_, err := registrar.Ejecutar(context.Background(), "proyecto-1", "Ana", "ana@ejemplo.test", integrante.RolProductBuilder)
			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.esperado)
			}
			if !strings.Contains(err.Error(), "proyecto-1") {
				t.Errorf("el mensaje %q no indica el proyecto", err.Error())
			}
			if len(c.repo.altas) != c.registros {
				t.Errorf("llamadas a RegistrarIntegrante = %d, se esperaban %d", len(c.repo.altas), c.registros)
			}
		})
	}
}
