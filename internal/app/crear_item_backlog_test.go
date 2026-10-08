package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// ahora es la hora del reloj fijo que reciben los casos de uso en estos tests.
var ahora = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

func relojFijo() time.Time { return ahora }

// alta es una llamada a RegistrarItem que guarda el repositorio falso.
type alta struct {
	proyectoID string
	item       backlog.ItemBacklog
}

// repositorioFalso simula el repositorio de backlog sin base de datos. Responde
// lo que se le configura y anota lo que le piden.
type repositorioFalso struct {
	existe       bool
	errExiste    error
	errRegistrar error

	consultas []string
	altas     []alta
}

func (r *repositorioFalso) ExisteProyecto(_ context.Context, proyectoID string) (bool, error) {
	r.consultas = append(r.consultas, proyectoID)
	return r.existe, r.errExiste
}

func (r *repositorioFalso) RegistrarItem(_ context.Context, proyectoID string, item backlog.ItemBacklog) (backlog.ItemBacklog, error) {
	r.altas = append(r.altas, alta{proyectoID: proyectoID, item: item})
	if r.errRegistrar != nil {
		return backlog.ItemBacklog{}, r.errRegistrar
	}
	item.ID = "id-del-repositorio"
	item.Numero = len(r.altas)
	item.ProyectoID = proyectoID
	return item, nil
}

func datosValidos() backlog.DatosItem {
	return backlog.DatosItem{
		Titulo:    "Crear proyecto",
		Prioridad: backlog.PrioridadMust,
		Criterios: []string{"Se guarda con sus fechas"},
	}
}

// US-005 / CA-005-1: con un proyecto existente y datos validos, el caso de uso
// registra el item con la fecha del reloj y devuelve el ID y el numero que
// asigna la persistencia.
func TestCrearItemBacklog_Registra(t *testing.T) {
	repo := &repositorioFalso{existe: true}
	crear := app.NuevoCrearItemBacklog(repo, relojFijo)

	item, err := crear.Ejecutar(context.Background(), "proyecto-1", datosValidos())
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if len(repo.altas) != 1 {
		t.Fatalf("se registraron %d items, se esperaba 1", len(repo.altas))
	}
	registrado := repo.altas[0]
	if registrado.proyectoID != "proyecto-1" {
		t.Errorf("proyecto del alta = %q, se esperaba %q", registrado.proyectoID, "proyecto-1")
	}
	if !registrado.item.CreadoEn.Equal(ahora) {
		t.Errorf("creado en = %v, se esperaba la hora del reloj %v", registrado.item.CreadoEn, ahora)
	}
	if registrado.item.Estado != backlog.EstadoPendiente {
		t.Errorf("estado = %q, se esperaba %q", registrado.item.Estado, backlog.EstadoPendiente)
	}
	if item.ID != "id-del-repositorio" || item.Numero != 1 || item.ProyectoID != "proyecto-1" {
		t.Errorf("ID, numero y proyecto = %q, %d, %q; se esperaban los del repositorio",
			item.ID, item.Numero, item.ProyectoID)
	}
}

// US-005 / CA-005-8, RN-005-9: si el proyecto no existe, el caso de uso responde
// ErrProyectoInexistente antes de validar los datos: no lo une con los errores
// de validacion (aunque los datos tambien sean invalidos) y no registra nada.
func TestCrearItemBacklog_ProyectoInexistente(t *testing.T) {
	repo := &repositorioFalso{existe: false}
	crear := app.NuevoCrearItemBacklog(repo, relojFijo)
	datos := backlog.DatosItem{Titulo: "   ", Prioridad: "urgente"}

	_, err := crear.Ejecutar(context.Background(), "proyecto-inexistente", datos)
	if !errors.Is(err, app.ErrProyectoInexistente) {
		t.Fatalf("error = %v, se esperaba ErrProyectoInexistente", err)
	}
	if errors.Is(err, backlog.ErrTituloVacio) || errors.Is(err, backlog.ErrPrioridadInvalida) {
		t.Errorf("el proyecto inexistente vino unido con errores de validacion: %v", err)
	}
	// Envuelto indicando el ID recibido.
	if !strings.Contains(err.Error(), "proyecto-inexistente") {
		t.Errorf("el mensaje %q no indica el ID del proyecto", err.Error())
	}
	if len(repo.consultas) != 1 || repo.consultas[0] != "proyecto-inexistente" {
		t.Errorf("consultas al repositorio = %q, se esperaba una por %q", repo.consultas, "proyecto-inexistente")
	}
	if len(repo.altas) != 0 {
		t.Errorf("se registraron %d items, se esperaba ninguno", len(repo.altas))
	}
}

// US-005 / CA-005-5, RN-005-12: con datos invalidos no se registra nada, y los
// errores del dominio se devuelven tal cual, sin prefijos: su mensaje es el que
// ve la persona (spec, seccion 7).
func TestCrearItemBacklog_DatosInvalidos(t *testing.T) {
	repo := &repositorioFalso{existe: true}
	crear := app.NuevoCrearItemBacklog(repo, relojFijo)
	datos := datosValidos()
	datos.Titulo = "   "

	_, err := crear.Ejecutar(context.Background(), "proyecto-1", datos)
	if !errors.Is(err, backlog.ErrTituloVacio) {
		t.Fatalf("error = %v, se esperaba ErrTituloVacio", err)
	}
	if err.Error() != "el titulo es obligatorio" {
		t.Errorf("mensaje = %q, se esperaba el del dominio sin cambios", err.Error())
	}
	if len(repo.altas) != 0 {
		t.Errorf("se registraron %d items, se esperaba ninguno", len(repo.altas))
	}
}

// US-005 / RN-005-9, RN-005-11: los errores del repositorio se devuelven
// envueltos con el proyecto, sin perder el original. Si el proyecto desaparece
// dentro de la transaccion, el repositorio devuelve ErrProyectoInexistente y
// el caso de uso lo deja reconocible.
func TestCrearItemBacklog_ErroresDelRepositorio(t *testing.T) {
	errBase := errors.New("la base no responde")
	casos := []struct {
		nombre    string
		repo      *repositorioFalso
		esperado  error
		registros int
	}{
		{"falla la consulta del proyecto", &repositorioFalso{errExiste: errBase}, errBase, 0},
		{"el proyecto desaparece antes del alta", &repositorioFalso{existe: true, errRegistrar: app.ErrProyectoInexistente}, app.ErrProyectoInexistente, 1},
		{"falla el alta", &repositorioFalso{existe: true, errRegistrar: errBase}, errBase, 1},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			crear := app.NuevoCrearItemBacklog(c.repo, relojFijo)

			_, err := crear.Ejecutar(context.Background(), "proyecto-1", datosValidos())
			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.esperado)
			}
			if !strings.Contains(err.Error(), "proyecto-1") {
				t.Errorf("el mensaje %q no indica el proyecto", err.Error())
			}
			if len(c.repo.altas) != c.registros {
				t.Errorf("llamadas a RegistrarItem = %d, se esperaban %d", len(c.repo.altas), c.registros)
			}
		})
	}
}
