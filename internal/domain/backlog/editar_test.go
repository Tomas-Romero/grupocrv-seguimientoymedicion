package backlog_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// datosNuevos son los datos con los que se edita el item de estos tests: todos
// distintos de los de datosValidos.
func datosNuevos() backlog.DatosItem {
	return backlog.DatosItem{
		Titulo:      "Crear proyecto con integrantes",
		Descripcion: "Alta de proyectos con su equipo",
		Prioridad:   backlog.PrioridadShould,
		Criterios:   []string{"Se guardan los integrantes"},
	}
}

// US-006 / CA-006-1, RN-006-2, RN-006-4: editar reemplaza los cuatro datos a la
// vez y conserva el ID, el numero, el proyecto, el estado, los Story Points y la
// fecha de creacion.
func TestEditar(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoPendiente, 5)

	editado, err := item.Editar(datosNuevos())
	if err != nil {
		t.Fatalf("Editar: %v", err)
	}

	nuevos := datosNuevos()
	if editado.Titulo != nuevos.Titulo || editado.Descripcion != nuevos.Descripcion ||
		editado.Prioridad != nuevos.Prioridad || !slices.Equal(editado.Criterios, nuevos.Criterios) {
		t.Errorf("datos editados = %+v, se esperaban %+v", editado, nuevos)
	}
	if editado.ID != item.ID || editado.Numero != item.Numero || editado.ProyectoID != item.ProyectoID ||
		editado.Estado != item.Estado || editado.StoryPoints != item.StoryPoints || !editado.CreadoEn.Equal(item.CreadoEn) {
		t.Errorf("Editar cambio algo mas que los cuatro datos:\n recibido: %+v\n original: %+v", editado, item)
	}
}

// US-006 / RN-006-1, CL-006-3: los datos se recortan igual que al crear
// (RN-005-1, RN-005-4 y RN-005-8).
func TestEditar_Recorta(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoPendiente, 0)

	editado, err := item.Editar(backlog.DatosItem{
		Titulo:      "  Crear proyecto  ",
		Descripcion: "\tAlta\n",
		Prioridad:   backlog.PrioridadCould,
		Criterios:   []string{" Se guarda ", "Se lista "},
	})
	if err != nil {
		t.Fatalf("Editar: %v", err)
	}
	if editado.Titulo != "Crear proyecto" || editado.Descripcion != "Alta" ||
		!slices.Equal(editado.Criterios, []string{"Se guarda", "Se lista"}) {
		t.Errorf("no se recorto como al crear: %+v", editado)
	}
}

// US-006 / CA-006-2, RN-006-3, CL-006-1: repriorizar es editar cambiando solo la
// prioridad, y editar con los mismos datos (o la misma prioridad) es valido y
// deja el item igual.
func TestEditar_Repriorizar(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoPendiente, 3)
	actuales := backlog.DatosItem{
		Titulo:      item.Titulo,
		Descripcion: item.Descripcion,
		Prioridad:   item.Prioridad,
		Criterios:   item.Criterios,
	}

	for _, prioridad := range []backlog.Prioridad{backlog.PrioridadWont, item.Prioridad} {
		datos := actuales
		datos.Prioridad = prioridad

		editado, err := item.Editar(datos)
		if err != nil {
			t.Fatalf("Editar(prioridad %q): %v", prioridad, err)
		}
		esperado := item
		esperado.Prioridad = prioridad
		if editado.Prioridad != prioridad || editado.Titulo != esperado.Titulo ||
			editado.Descripcion != esperado.Descripcion || !slices.Equal(editado.Criterios, esperado.Criterios) ||
			editado.StoryPoints != esperado.StoryPoints || editado.Numero != esperado.Numero {
			t.Errorf("Editar(prioridad %q) = %+v, se esperaba %+v", prioridad, editado, esperado)
		}
	}
}

// US-006 / CL-006-5, RN-006-6: un item pendiente o en progreso se edita. La
// Definition of Ready no se vuelve a verificar: un item estimado puede quedar
// sin criterios (CL-006-4).
func TestEditar_PendienteOEnProgreso(t *testing.T) {
	for _, estado := range []backlog.Estado{backlog.EstadoPendiente, backlog.EstadoEnProgreso} {
		t.Run(string(estado), func(t *testing.T) {
			item := itemEnEstado(t, estado, 8)
			datos := datosNuevos()
			datos.Criterios = nil

			editado, err := item.Editar(datos)
			if err != nil {
				t.Fatalf("Editar: %v", err)
			}
			if editado.Estado != estado {
				t.Errorf("estado = %q, se esperaba %q", editado.Estado, estado)
			}
		})
	}
}

// US-006 / CA-006-4, RN-006-2, CL-006-2: la lista de criterios se reemplaza
// entera, y una lista vacia reemplaza a la anterior.
func TestEditar_CriteriosVacios(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoPendiente, 0)
	if len(item.Criterios) == 0 {
		t.Fatal("el item de partida no tiene criterios: el test no prueba nada")
	}

	for _, criterios := range [][]string{nil, {}} {
		datos := datosNuevos()
		datos.Criterios = criterios

		editado, err := item.Editar(datos)
		if err != nil {
			t.Fatalf("Editar(%q): %v", criterios, err)
		}
		if len(editado.Criterios) != 0 {
			t.Errorf("criterios = %q, se esperaba la lista vacia", editado.Criterios)
		}
	}
}

// US-006 / CA-006-4, CA-006-5, RN-006-1, CL-006-2, CL-006-3: los datos invalidos
// se rechazan con exactamente los mismos errores que al crear, unidos en el orden
// de RN-005-13. Se compara con lo que devuelve NuevoItem: hay una sola
// validacion.
func TestEditar_DatosInvalidos(t *testing.T) {
	casos := []struct {
		nombre  string
		cambiar func(*backlog.DatosItem)
		errs    []error
	}{
		{"titulo vacio", func(d *backlog.DatosItem) { d.Titulo = "" }, []error{backlog.ErrTituloVacio}},
		{"titulo con solo espacios", func(d *backlog.DatosItem) { d.Titulo = "   " }, []error{backlog.ErrTituloVacio}},
		{"titulo de 121", func(d *backlog.DatosItem) { d.Titulo = strings.Repeat("a", 121) }, []error{backlog.ErrTituloMuyLargo}},
		{"prioridad fuera de MoSCoW", func(d *backlog.DatosItem) { d.Prioridad = "urgente" }, []error{backlog.ErrPrioridadInvalida}},
		{"criterio vacio", func(d *backlog.DatosItem) { d.Criterios = []string{"Se guarda", " "} }, []error{backlog.ErrCriterioVacio}},
		{"todo invalido", func(d *backlog.DatosItem) {
			d.Titulo = ""
			d.Prioridad = "MUST"
			d.Criterios = []string{""}
		}, []error{backlog.ErrTituloVacio, backlog.ErrPrioridadInvalida, backlog.ErrCriterioVacio}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			item := itemEnEstado(t, backlog.EstadoPendiente, 5)
			datos := datosNuevos()
			c.cambiar(&datos)

			editado, err := item.Editar(datos)

			for _, esperado := range c.errs {
				if !errors.Is(err, esperado) {
					t.Errorf("error = %v, se esperaba que incluya %v", err, esperado)
				}
			}
			_, errAlCrear := backlog.NuevoItem(datos, creadoEn)
			if err == nil || err.Error() != errAlCrear.Error() {
				t.Errorf("error = %v, se esperaba el mismo que al crear: %v", err, errAlCrear)
			}
			if editado.Titulo != "" {
				t.Errorf("devolvio un item %+v, se esperaba el valor cero", editado)
			}
		})
	}
}

// US-006 / CA-006-4, CL-006-3: un titulo de 120 caracteres se acepta.
func TestEditar_TituloDe120(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoPendiente, 0)
	datos := datosNuevos()
	datos.Titulo = strings.Repeat("ñ", 120)

	if _, err := item.Editar(datos); err != nil {
		t.Errorf("Editar con un titulo de 120 caracteres: %v", err)
	}
}

// US-006 / RN-006-10: Editar devuelve una copia. Si hay error, o si sale bien,
// el item recibido no cambia, y la copia no comparte los criterios con los
// datos recibidos.
func TestEditar_NoModificaElOriginal(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoPendiente, 3)
	original := item
	original.Criterios = slices.Clone(item.Criterios)
	datos := datosNuevos()

	editado, err := item.Editar(datos)
	if err != nil {
		t.Fatalf("Editar: %v", err)
	}
	invalidos := datosNuevos()
	invalidos.Titulo = ""
	_, _ = item.Editar(invalidos)

	if item.Titulo != original.Titulo || item.Prioridad != original.Prioridad ||
		!slices.Equal(item.Criterios, original.Criterios) {
		t.Errorf("el item original cambio: %+v, era %+v", item, original)
	}
	datos.Criterios[0] = "cambiado"
	if editado.Criterios[0] == "cambiado" {
		t.Error("el item editado comparte la lista de criterios con los datos recibidos")
	}
}

// US-006 / CA-006-6, RN-006-5, CL-006-6: un item completado no se edita. Se
// valida primero el estado: aunque los datos tambien sean invalidos, el unico
// error es ErrItemCompletado, envuelto con el ID del item.
func TestEditar_ItemCompletado(t *testing.T) {
	invalidos := datosNuevos()
	invalidos.Titulo = ""
	invalidos.Prioridad = "urgente"

	for nombre, datos := range map[string]backlog.DatosItem{"datos validos": datosNuevos(), "datos invalidos": invalidos} {
		t.Run(nombre, func(t *testing.T) {
			item := itemEnEstado(t, backlog.EstadoCompletado, 5)

			editado, err := item.Editar(datos)

			if !errors.Is(err, backlog.ErrItemCompletado) {
				t.Fatalf("error = %v, se esperaba ErrItemCompletado", err)
			}
			if errors.Is(err, backlog.ErrTituloVacio) || errors.Is(err, backlog.ErrPrioridadInvalida) {
				t.Errorf("el error %q tambien reporta los datos: debe ser solo ErrItemCompletado", err)
			}
			if !strings.Contains(err.Error(), item.ID) {
				t.Errorf("el error %q no indica el ID del item", err)
			}
			if editado.Titulo != "" {
				t.Errorf("devolvio un item %+v, se esperaba el valor cero", editado)
			}
		})
	}
}
