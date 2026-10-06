package backlog_test

import (
	"errors"
	"reflect"
	"slices"
	"strconv"
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

// US-005 / CA-005-4, RN-005-1, RN-005-3, CL-005-1, CL-005-2, CL-005-3, CL-005-4,
// CL-005-15: el titulo tiene como maximo 120 caracteres, contados en runas y
// despues de recortar los extremos. Los saltos de linea internos cuentan.
func TestNuevoItem_LargoDelTitulo(t *testing.T) {
	a120 := strings.Repeat("a", 120)
	casos := []struct {
		nombre      string
		titulo      string
		guardado    string // titulo esperado si se acepta
		errEsperado error  // error esperado si se rechaza
		mensaje     string // mensaje esperado si se rechaza
	}{
		{nombre: "120 caracteres", titulo: a120, guardado: a120},
		{
			nombre: "121 caracteres", titulo: strings.Repeat("a", 121), errEsperado: backlog.ErrTituloMuyLargo,
			mensaje: "el titulo no puede tener mas de 120 caracteres (tiene 121)",
		},
		{nombre: "120 caracteres con espacios en los extremos", titulo: "  " + a120 + "  ", guardado: a120},
		{nombre: "120 letras ñ, que ocupan 240 bytes", titulo: strings.Repeat("ñ", 120), guardado: strings.Repeat("ñ", 120)},
		{nombre: "121 letras ñ", titulo: strings.Repeat("ñ", 121), errEsperado: backlog.ErrTituloMuyLargo},
		{nombre: "tabulaciones y saltos en los extremos", titulo: "\tCrear proyecto\n", guardado: "Crear proyecto"},
		{
			nombre:   "salto de linea interno dentro del limite",
			titulo:   strings.Repeat("a", 60) + "\n" + strings.Repeat("a", 59),
			guardado: strings.Repeat("a", 60) + "\n" + strings.Repeat("a", 59),
		},
		{
			nombre: "salto de linea interno que pasa el limite",
			titulo: strings.Repeat("a", 60) + "\n" + strings.Repeat("a", 60), errEsperado: backlog.ErrTituloMuyLargo,
		},
		{nombre: "mas de 120 espacios es un titulo vacio, no largo", titulo: strings.Repeat(" ", 130), errEsperado: backlog.ErrTituloVacio},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			datos := datosValidos()
			datos.Titulo = c.titulo

			item, err := backlog.NuevoItem(datos, creadoEn)
			if c.errEsperado == nil {
				if err != nil {
					t.Fatalf("NuevoItem: %v", err)
				}
				if item.Titulo != c.guardado {
					t.Errorf("titulo guardado = %q, se esperaba %q", item.Titulo, c.guardado)
				}
				return
			}
			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if errors.Is(c.errEsperado, backlog.ErrTituloVacio) && errors.Is(err, backlog.ErrTituloMuyLargo) {
				t.Errorf("un titulo de solo espacios tambien se informo como demasiado largo: %v", err)
			}
			if c.mensaje != "" && err.Error() != c.mensaje {
				t.Errorf("mensaje = %q, se esperaba %q", err.Error(), c.mensaje)
			}
		})
	}
}

// US-005 / CA-005-2, RN-005-4, CL-005-5, CL-005-13: la descripcion es opcional,
// se recorta y puede quedar vacia; los saltos de linea internos se conservan.
func TestNuevoItem_Descripcion(t *testing.T) {
	casos := []struct {
		nombre      string
		descripcion string
		guardada    string
	}{
		{"se recortan los extremos", "  Alta de proyectos \n", "Alta de proyectos"},
		{"omitida", "", ""},
		{"solo espacios, tabulaciones y saltos de linea", " \t\n ", ""},
		{"saltos de linea internos", "Linea 1\nLinea 2", "Linea 1\nLinea 2"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			datos := datosValidos()
			datos.Descripcion = c.descripcion

			item, err := backlog.NuevoItem(datos, creadoEn)
			if err != nil {
				t.Fatalf("NuevoItem: %v", err)
			}
			if item.Descripcion != c.guardada {
				t.Errorf("descripcion guardada = %q, se esperaba %q", item.Descripcion, c.guardada)
			}
		})
	}
}

// US-005 / CA-005-6, RN-005-5, CL-005-7, CL-005-12: la prioridad es obligatoria y
// sigue MoSCoW. Se aceptan solo los cuatro valores exactos en minuscula; el
// dominio no la pasa a minuscula ni la recorta.
func TestNuevoItem_Prioridad(t *testing.T) {
	validas := []backlog.Prioridad{
		backlog.PrioridadMust, backlog.PrioridadShould, backlog.PrioridadCould, backlog.PrioridadWont,
	}
	for _, p := range validas {
		t.Run("valida "+string(p), func(t *testing.T) {
			datos := datosValidos()
			datos.Prioridad = p

			item, err := backlog.NuevoItem(datos, creadoEn)
			if err != nil {
				t.Fatalf("NuevoItem: %v", err)
			}
			if item.Prioridad != p {
				t.Errorf("prioridad = %q, se esperaba %q", item.Prioridad, p)
			}
		})
	}

	invalidas := []backlog.Prioridad{"urgente", "", "MUST", " must ", "Must"}
	for _, p := range invalidas {
		t.Run("invalida "+strconv.Quote(string(p)), func(t *testing.T) {
			datos := datosValidos()
			datos.Prioridad = p

			_, err := backlog.NuevoItem(datos, creadoEn)
			if !errors.Is(err, backlog.ErrPrioridadInvalida) {
				t.Fatalf("error = %v, se esperaba ErrPrioridadInvalida", err)
			}
			// El error indica el valor recibido.
			if !strings.Contains(err.Error(), strconv.Quote(string(p))) {
				t.Errorf("el mensaje %q no indica el valor recibido %q", err.Error(), string(p))
			}
		})
	}
}

// errorUnido es lo que devuelve errors.Join: permite ver cada error por separado
// y en su orden, sin una asercion de tipo sobre el error.
type errorUnido interface {
	Unwrap() []error
}

// US-005 / CA-005-7, RN-005-8, RN-005-14, CL-005-6: cada criterio vacio o con
// solo espacios es un error con su posicion, contada desde 1. Se reconoce con
// errors.Is y la posicion se obtiene con errors.As.
func TestNuevoItem_CriteriosVacios(t *testing.T) {
	datos := datosValidos()
	datos.Criterios = []string{"", "válido", "  "}

	_, err := backlog.NuevoItem(datos, creadoEn)
	if !errors.Is(err, backlog.ErrCriterioVacio) {
		t.Fatalf("error = %v, se esperaba ErrCriterioVacio", err)
	}
	var unido errorUnido
	if !errors.As(err, &unido) {
		t.Fatalf("se esperaban los errores unidos con errors.Join, se obtuvo: %v", err)
	}
	errs := unido.Unwrap()
	posiciones := []int{1, 3}
	if len(errs) != len(posiciones) {
		t.Fatalf("se obtuvieron %d errores, se esperaban %d: %v", len(errs), len(posiciones), err)
	}
	for i, posicion := range posiciones {
		var criterio backlog.CriterioVacioError
		if !errors.As(errs[i], &criterio) {
			t.Fatalf("error %d = %v, se esperaba un CriterioVacioError", i+1, errs[i])
		}
		if criterio.Posicion != posicion {
			t.Errorf("error %d: posicion = %d, se esperaba %d", i+1, criterio.Posicion, posicion)
		}
	}
	mensaje := "el criterio de aceptacion de la posicion 1 esta vacio\n" +
		"el criterio de aceptacion de la posicion 3 esta vacio"
	if err.Error() != mensaje {
		t.Errorf("mensaje = %q, se esperaba %q", err.Error(), mensaje)
	}
}

// US-005 / RN-005-8: los criterios se guardan recortados y en el orden en que
// se recibieron.
func TestNuevoItem_CriteriosRecortados(t *testing.T) {
	datos := datosValidos()
	datos.Criterios = []string{"  Se guarda  ", "\tSe lista\n"}

	item, err := backlog.NuevoItem(datos, creadoEn)
	if err != nil {
		t.Fatalf("NuevoItem: %v", err)
	}
	esperados := []string{"Se guarda", "Se lista"}
	if !slices.Equal(item.Criterios, esperados) {
		t.Errorf("criterios = %q, se esperaba %q", item.Criterios, esperados)
	}
}

// US-005 / RN-005-13, CL-005-11: con varios datos invalidos el dominio devuelve
// todos los errores juntos, unidos con errors.Join y en orden fijo: titulo,
// prioridad y criterios por posicion. errors.Is reconoce cada uno.
func TestNuevoItem_VariosErrores(t *testing.T) {
	datos := backlog.DatosItem{
		Titulo:    "   ",
		Prioridad: "urgente",
		Criterios: []string{"Se guarda", ""},
	}

	item, err := backlog.NuevoItem(datos, creadoEn)
	var unido errorUnido
	if !errors.As(err, &unido) {
		t.Fatalf("se esperaban los errores unidos con errors.Join, se obtuvo: %v", err)
	}
	errs := unido.Unwrap()
	esperados := []error{backlog.ErrTituloVacio, backlog.ErrPrioridadInvalida, backlog.ErrCriterioVacio}
	if len(errs) != len(esperados) {
		t.Fatalf("se obtuvieron %d errores, se esperaban %d: %v", len(errs), len(esperados), err)
	}
	for i, esperado := range esperados {
		if !errors.Is(errs[i], esperado) {
			t.Errorf("error %d = %v, se esperaba %v", i+1, errs[i], esperado)
		}
		if !errors.Is(err, esperado) {
			t.Errorf("errors.Is no reconoce %v en el error unido", esperado)
		}
	}
	mensaje := "el titulo es obligatorio\n" +
		"la prioridad tiene que ser must, should, could o wont: se recibio \"urgente\"\n" +
		"el criterio de aceptacion de la posicion 2 esta vacio"
	if err.Error() != mensaje {
		t.Errorf("mensaje = %q, se esperaba %q", err.Error(), mensaje)
	}
	if !reflect.DeepEqual(item, backlog.ItemBacklog{}) {
		t.Errorf("con errores devolvio el item %+v, se esperaba ninguno", item)
	}
}
