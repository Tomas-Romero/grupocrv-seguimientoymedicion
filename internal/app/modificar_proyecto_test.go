package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// repoModificarFalso simula el repositorio sin base de datos. Responde lo que se
// le configura y anota lo que le piden.
type repoModificarFalso struct {
	guardado      proyecto.Proyecto
	existe        bool
	errBuscar     error
	errActualizar error

	busquedas []string
	cambios   []proyecto.Proyecto
}

func (r *repoModificarFalso) BuscarProyecto(_ context.Context, proyectoID string) (proyecto.Proyecto, error) {
	r.busquedas = append(r.busquedas, proyectoID)
	if r.errBuscar != nil {
		return proyecto.Proyecto{}, r.errBuscar
	}
	if !r.existe {
		return proyecto.Proyecto{}, app.ErrProyectoInexistente
	}
	return r.guardado, nil
}

func (r *repoModificarFalso) ActualizarProyecto(_ context.Context, p proyecto.Proyecto) error {
	r.cambios = append(r.cambios, p)
	return r.errActualizar
}

// proyectoGuardado es el proyecto que el repositorio falso tiene cargado.
func proyectoGuardado() proyecto.Proyecto {
	return proyecto.Proyecto{
		ID:          "proyecto-1",
		Nombre:      "Demo",
		Descripcion: "Descripcion original",
		FechaInicio: fechaDe(2026, time.September, 14),
		FechaFin:    fechaDe(2026, time.November, 16),
	}
}

// US-002 / CA-002-1, CA-002-2, RN-002-3, RN-002-4: con un proyecto existente y
// datos validos, el caso de uso busca el proyecto, lo modifica con el dominio
// (nombre recortado, mismo ID) y manda a guardar los cuatro datos nuevos.
func TestModificarProyecto_Actualiza(t *testing.T) {
	repo := &repoModificarFalso{guardado: proyectoGuardado(), existe: true}
	modificar := app.NuevoModificarProyecto(repo)
	inicio, fin := fechaDe(2026, time.October, 1), fechaDe(2026, time.December, 15)

	p, err := modificar.Ejecutar(context.Background(), "proyecto-1", "  Demo v2  ", "", inicio, fin)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if len(repo.busquedas) != 1 || repo.busquedas[0] != "proyecto-1" {
		t.Errorf("busquedas = %q, se esperaba una por %q", repo.busquedas, "proyecto-1")
	}
	if len(repo.cambios) != 1 {
		t.Fatalf("se actualizaron %d proyectos, se esperaba 1", len(repo.cambios))
	}
	esperado := proyecto.Proyecto{
		ID:          "proyecto-1",
		Nombre:      "Demo v2",
		Descripcion: "",
		FechaInicio: inicio,
		FechaFin:    fin,
	}
	if repo.cambios[0] != esperado {
		t.Errorf("proyecto enviado a guardar = %+v, se esperaba %+v", repo.cambios[0], esperado)
	}
	if p != esperado {
		t.Errorf("proyecto devuelto = %+v, se esperaba %+v", p, esperado)
	}
}

// US-002 / CA-002-5, RN-002-1: si el proyecto no existe, el caso de uso responde
// ErrProyectoInexistente antes de validar los datos: no lo une con los errores
// de validacion (aunque los datos tambien sean invalidos) y no guarda nada.
func TestModificarProyecto_ProyectoInexistente(t *testing.T) {
	repo := &repoModificarFalso{existe: false}
	modificar := app.NuevoModificarProyecto(repo)

	_, err := modificar.Ejecutar(context.Background(), "proyecto-inexistente", "",
		"", fechaDe(2026, time.November, 18), fechaDe(2026, time.September, 14))
	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Fatalf("error = %v, se esperaba ErrProyectoInexistente", err)
	}
	if errors.Is(err, proyecto.ErrNombreVacio) || errors.Is(err, proyecto.ErrFechasIncoherentes) {
		t.Errorf("el proyecto inexistente vino unido con errores de validacion: %v", err)
	}
	for _, fragmento := range []string{"el proyecto no existe", "proyecto-inexistente"} {
		if !strings.Contains(err.Error(), fragmento) {
			t.Errorf("el mensaje %q no contiene %q", err.Error(), fragmento)
		}
	}
	if len(repo.cambios) != 0 {
		t.Errorf("se actualizaron %d proyectos, se esperaba ninguno", len(repo.cambios))
	}
}

// US-002 / CA-002-4, RN-002-2, RN-002-5, CL-002-7: con datos invalidos no se
// guarda nada; los errores del dominio vienen unidos, siguen reconocibles con
// errors.Is y el mensaje indica el proyecto.
func TestModificarProyecto_DatosInvalidos(t *testing.T) {
	repo := &repoModificarFalso{guardado: proyectoGuardado(), existe: true}
	modificar := app.NuevoModificarProyecto(repo)

	p, err := modificar.Ejecutar(context.Background(), "proyecto-1", "", "",
		fechaDe(2026, time.November, 18), fechaDe(2026, time.September, 14))
	if err == nil {
		t.Fatal("se esperaba un error y no hubo ninguno")
	}
	for _, esperado := range []error{proyecto.ErrNombreVacio, proyecto.ErrFechasIncoherentes} {
		if !errors.Is(err, esperado) {
			t.Errorf("errors.Is no reconoce %v dentro de %q", esperado, err)
		}
	}
	for _, fragmento := range []string{"el nombre es obligatorio", "fecha de fin anterior a la de inicio", "proyecto-1"} {
		if !strings.Contains(err.Error(), fragmento) {
			t.Errorf("el mensaje %q no contiene %q", err.Error(), fragmento)
		}
	}
	if len(repo.cambios) != 0 {
		t.Errorf("se actualizaron %d proyectos, se esperaba ninguno (todo o nada)", len(repo.cambios))
	}
	if p != (proyecto.Proyecto{}) {
		t.Errorf("no se esperaba un proyecto con datos invalidos, se obtuvo %+v", p)
	}
}

// US-002 / seccion 7: los errores del repositorio se devuelven envueltos con el
// proyecto, sin perder el original. Si el proyecto desaparece entre la busqueda
// y la actualizacion, el repositorio devuelve ErrProyectoInexistente y el caso
// de uso lo deja reconocible.
func TestModificarProyecto_ErroresDelRepositorio(t *testing.T) {
	errBase := errors.New("la base no responde")
	casos := []struct {
		nombre   string
		repo     *repoModificarFalso
		esperado error
		cambios  int
	}{
		{"falla la busqueda del proyecto", &repoModificarFalso{errBuscar: errBase}, errBase, 0},
		{"el proyecto desaparece antes de actualizar", &repoModificarFalso{guardado: proyectoGuardado(), existe: true, errActualizar: app.ErrProyectoInexistente}, app.ErrProyectoInexistente, 1},
		{"falla la actualizacion", &repoModificarFalso{guardado: proyectoGuardado(), existe: true, errActualizar: errBase}, errBase, 1},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			modificar := app.NuevoModificarProyecto(c.repo)

			_, err := modificar.Ejecutar(context.Background(), "proyecto-1", "Demo v2", "",
				fechaDe(2026, time.September, 14), fechaDe(2026, time.November, 16))
			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.esperado)
			}
			if !strings.Contains(err.Error(), "proyecto-1") {
				t.Errorf("el mensaje %q no indica el proyecto", err.Error())
			}
			if len(c.repo.cambios) != c.cambios {
				t.Errorf("llamadas a ActualizarProyecto = %d, se esperaban %d", len(c.repo.cambios), c.cambios)
			}
		})
	}
}