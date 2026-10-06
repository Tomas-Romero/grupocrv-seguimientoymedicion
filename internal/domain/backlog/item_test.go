package backlog_test

import (
	"slices"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// creadoEn es la fecha de creacion de todos los items de estos tests: el dominio
// la recibe por parametro y no lee el reloj.
var creadoEn = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

// datosValidos devuelve los datos de un item valido; cada test cambia solo lo
// que prueba.
func datosValidos() backlog.DatosItem {
	return backlog.DatosItem{
		Titulo:      "Crear proyecto",
		Descripcion: "Alta de proyectos con sus fechas",
		Prioridad:   backlog.PrioridadMust,
		Criterios:   []string{"Se guarda con sus fechas", "Aparece en el listado"},
	}
}

// US-005 / CA-005-1, RN-005-6, RN-005-7: un item valido queda pendiente, sin
// estimar y con los datos recibidos. No tiene ID, numero ni proyecto: los
// completa la persistencia.
func TestNuevoItem_DatosValidos(t *testing.T) {
	item, err := backlog.NuevoItem(datosValidos(), creadoEn)
	if err != nil {
		t.Fatalf("NuevoItem: %v", err)
	}

	if item.Titulo != "Crear proyecto" {
		t.Errorf("titulo = %q, se esperaba %q", item.Titulo, "Crear proyecto")
	}
	if item.Descripcion != "Alta de proyectos con sus fechas" {
		t.Errorf("descripcion = %q, se esperaba %q", item.Descripcion, "Alta de proyectos con sus fechas")
	}
	if item.Prioridad != backlog.PrioridadMust {
		t.Errorf("prioridad = %q, se esperaba %q", item.Prioridad, backlog.PrioridadMust)
	}
	if item.Estado != backlog.EstadoPendiente {
		t.Errorf("estado = %q, se esperaba %q", item.Estado, backlog.EstadoPendiente)
	}
	if puntos, estimado := item.StoryPoints.Valor(); estimado {
		t.Errorf("story points = %d, se esperaba un item sin estimar", puntos)
	}
	criterios := []string{"Se guarda con sus fechas", "Aparece en el listado"}
	if !slices.Equal(item.Criterios, criterios) {
		t.Errorf("criterios = %q, se esperaba %q", item.Criterios, criterios)
	}
	if !item.CreadoEn.Equal(creadoEn) {
		t.Errorf("creado en = %v, se esperaba %v", item.CreadoEn, creadoEn)
	}
	if item.ID != "" || item.Numero != 0 || item.ProyectoID != "" {
		t.Errorf("ID, numero y proyecto = %q, %d, %q; se esperaban vacios hasta registrar el item",
			item.ID, item.Numero, item.ProyectoID)
	}
}
