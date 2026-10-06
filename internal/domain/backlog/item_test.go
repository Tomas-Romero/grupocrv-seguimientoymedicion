package backlog_test

import (
	"errors"
	"reflect"
	"slices"
	"strings"
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

// US-005 / CA-005-5, RN-005-1, RN-005-2, CL-005-4: un titulo vacio o con solo
// espacios se rechaza, aunque los espacios sean mas de 120, y no se devuelve
// ningun item.
func TestNuevoItem_TituloVacio(t *testing.T) {
	casos := []struct {
		nombre string
		titulo string
	}{
		{"vacio", ""},
		{"solo espacios", "   "},
		{"tabulaciones y saltos de linea", "\t\n \r\n"},
		{"mas de 120 espacios", strings.Repeat(" ", 130)},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			datos := datosValidos()
			datos.Titulo = c.titulo

			item, err := backlog.NuevoItem(datos, creadoEn)
			if !errors.Is(err, backlog.ErrTituloVacio) {
				t.Fatalf("error = %v, se esperaba ErrTituloVacio", err)
			}
			if !reflect.DeepEqual(item, backlog.ItemBacklog{}) {
				t.Errorf("con un error devolvio el item %+v, se esperaba ninguno", item)
			}
		})
	}
}
