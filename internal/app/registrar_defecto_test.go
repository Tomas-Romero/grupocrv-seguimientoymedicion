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
