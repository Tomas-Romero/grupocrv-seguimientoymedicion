package app_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// repositorioEdicionFalso simula la persistencia de la edicion sin base de datos.
// Responde lo que se le configura y anota lo que le piden.
type repositorioEdicionFalso struct {
	item       backlog.ItemBacklog
	existe     bool
	errObtener error
	errGuardar error
	guardados  []backlog.ItemBacklog
}

func (r *repositorioEdicionFalso) ObtenerItem(_ context.Context, _ string) (backlog.ItemBacklog, bool, error) {
	return r.item, r.existe, r.errObtener
}

func (r *repositorioEdicionFalso) GuardarEdicion(_ context.Context, item backlog.ItemBacklog) error {
	r.guardados = append(r.guardados, item)
	return r.errGuardar
}

// datosEditados son los datos con los que se edita el item en estos tests.
func datosEditados() backlog.DatosItem {
	return backlog.DatosItem{
		Titulo:      "  Crear proyecto con integrantes ",
		Descripcion: "Con su equipo",
		Prioridad:   backlog.PrioridadShould,
		Criterios:   []string{"Se guardan los integrantes"},
	}
}

// US-006 / CA-006-1, RN-006-2: el caso de uso carga el item, lo edita con el
// dominio y guarda el item editado. Devuelve el item con los datos nuevos,
// recortados, y el resto sin cambios.
func TestEditarItem_Edita(t *testing.T) {
	original := itemDelRepositorio(t, backlog.EstadoPendiente)
	repo := &repositorioEdicionFalso{item: original, existe: true}

	item, err := app.NuevoEditarItem(repo).Ejecutar(context.Background(), "item-1", datosEditados())
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if item.Titulo != "Crear proyecto con integrantes" || item.Prioridad != backlog.PrioridadShould ||
		item.Descripcion != "Con su equipo" || !slices.Equal(item.Criterios, []string{"Se guardan los integrantes"}) {
		t.Errorf("item devuelto = %+v, se esperaban los datos nuevos recortados", item)
	}
	if item.ID != original.ID || item.Numero != original.Numero || item.Estado != original.Estado {
		t.Errorf("item devuelto = %+v, se esperaba el mismo ID, numero y estado que %+v", item, original)
	}
	if len(repo.guardados) != 1 {
		t.Fatalf("se guardaron %d ediciones, se esperaba 1", len(repo.guardados))
	}
	if guardado := repo.guardados[0]; guardado.ID != "item-1" || guardado.Titulo != item.Titulo {
		t.Errorf("se guardo %+v, se esperaba el item editado", guardado)
	}
}

// US-006 / CA-006-7, RN-006-7, CL-006-9: si el item no existe, el caso de uso
// devuelve ErrItemInexistente sin validar los datos y no guarda nada. El error
// no se une con los de validacion.
func TestEditarItem_ItemInexistente(t *testing.T) {
	repo := &repositorioEdicionFalso{existe: false}
	invalidos := datosEditados()
	invalidos.Titulo = ""

	_, err := app.NuevoEditarItem(repo).Ejecutar(context.Background(), "no-existe", invalidos)

	if !errors.Is(err, app.ErrItemInexistente) {
		t.Fatalf("error = %v, se esperaba ErrItemInexistente", err)
	}
	if errors.Is(err, backlog.ErrTituloVacio) {
		t.Errorf("el error %q tambien reporta los datos: ErrItemInexistente no se une con los de validacion", err)
	}
	if !strings.Contains(err.Error(), "no-existe") {
		t.Errorf("el error %q no indica el ID recibido", err)
	}
	if len(repo.guardados) != 0 {
		t.Errorf("se guardaron %d ediciones, se esperaba ninguna", len(repo.guardados))
	}
}

// US-006 / CA-006-5, CA-006-6, RN-006-10: datos invalidos o un item completado
// se rechazan con el error del dominio y no se guarda nada.
func TestEditarItem_Rechazos(t *testing.T) {
	invalidos := datosEditados()
	invalidos.Prioridad = "urgente"
	casos := []struct {
		nombre      string
		estado      backlog.Estado
		datos       backlog.DatosItem
		errEsperado error
	}{
		{"datos invalidos", backlog.EstadoPendiente, invalidos, backlog.ErrPrioridadInvalida},
		{"item completado", backlog.EstadoCompletado, datosEditados(), backlog.ErrItemCompletado},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := &repositorioEdicionFalso{item: itemDelRepositorio(t, c.estado), existe: true}

			item, err := app.NuevoEditarItem(repo).Ejecutar(context.Background(), "item-1", c.datos)

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if item.ID != "" {
				t.Errorf("devolvio el item %+v, se esperaba el valor cero", item)
			}
			if len(repo.guardados) != 0 {
				t.Errorf("se guardaron %d ediciones, se esperaba ninguna", len(repo.guardados))
			}
		})
	}
}

// US-006 / RN-006-8, CL-006-7: si el item desaparece o se completa entre la
// lectura y el guardado, el repositorio devuelve ErrItemInexistente o
// ErrItemCompletado, y el caso de uso los deja pasar reconocibles, sin
// devolver item.
func TestEditarItem_CambiaEntreLecturaYGuardado(t *testing.T) {
	for _, errGuardar := range []error{app.ErrItemInexistente, backlog.ErrItemCompletado} {
		t.Run(errGuardar.Error(), func(t *testing.T) {
			repo := &repositorioEdicionFalso{
				item:       itemDelRepositorio(t, backlog.EstadoPendiente),
				existe:     true,
				errGuardar: errGuardar,
			}

			item, err := app.NuevoEditarItem(repo).Ejecutar(context.Background(), "item-1", datosEditados())

			if !errors.Is(err, errGuardar) {
				t.Fatalf("error = %v, se esperaba %v", err, errGuardar)
			}
			if item.ID != "" {
				t.Errorf("devolvio el item %+v, se esperaba el valor cero", item)
			}
		})
	}
}

// Un fallo del repositorio al cargar el item se devuelve envuelto, con el ID, y
// no es un item inexistente.
func TestEditarItem_ErrorAlObtener(t *testing.T) {
	falla := errors.New("se corto la conexion")
	repo := &repositorioEdicionFalso{errObtener: falla}

	_, err := app.NuevoEditarItem(repo).Ejecutar(context.Background(), "item-1", datosEditados())

	if !errors.Is(err, falla) {
		t.Fatalf("error = %v, se esperaba el del repositorio", err)
	}
	if !strings.Contains(err.Error(), "item-1") {
		t.Errorf("el error %q no indica el item", err)
	}
	if errors.Is(err, app.ErrItemInexistente) {
		t.Errorf("un fallo del repositorio no es un item inexistente: %v", err)
	}
	if len(repo.guardados) != 0 {
		t.Errorf("se guardaron %d ediciones, se esperaba ninguna", len(repo.guardados))
	}
}

// Un fallo del repositorio al guardar se devuelve envuelto y no se devuelve item.
func TestEditarItem_ErrorAlGuardar(t *testing.T) {
	falla := errors.New("se corto la conexion")
	repo := &repositorioEdicionFalso{item: itemDelRepositorio(t, backlog.EstadoPendiente), existe: true, errGuardar: falla}

	item, err := app.NuevoEditarItem(repo).Ejecutar(context.Background(), "item-1", datosEditados())

	if !errors.Is(err, falla) {
		t.Fatalf("error = %v, se esperaba el del repositorio", err)
	}
	if !strings.Contains(err.Error(), "item-1") {
		t.Errorf("el error %q no indica el item", err)
	}
	if item.ID != "" {
		t.Errorf("devolvio el item %+v, se esperaba el valor cero", item)
	}
}
