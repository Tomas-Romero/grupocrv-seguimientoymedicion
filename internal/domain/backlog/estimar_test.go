package backlog_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// itemEnEstado arma un item valido en el estado y con la estimacion pedidos. Se
// parte de NuevoItem y se le cambian los datos que la persistencia o US-010
// cambiarian despues.
func itemEnEstado(t *testing.T, estado backlog.Estado, estimado int) backlog.ItemBacklog {
	t.Helper()
	item, err := backlog.NuevoItem(datosValidos(), creadoEn)
	if err != nil {
		t.Fatalf("NuevoItem: %v", err)
	}
	item.ID = "id-del-item"
	item.Numero = 3
	item.ProyectoID = "id-del-proyecto"
	item.Estado = estado
	if estimado != 0 {
		sp, err := backlog.NuevosStoryPoints(estimado)
		if err != nil {
			t.Fatalf("NuevosStoryPoints(%d): %v", estimado, err)
		}
		item.StoryPoints = sp
	}
	return item
}

func puntosDe(t *testing.T, item backlog.ItemBacklog) (int, bool) {
	t.Helper()
	return item.StoryPoints.Valor()
}

// US-013 / CA-013-1, RN-013-3, CL-013-6, CL-013-7, CL-013-10: un item pendiente o
// en progreso se estima, y se puede volver a estimar (incluso con el mismo
// valor): la ultima estimacion reemplaza a la anterior.
func TestEstimar(t *testing.T) {
	casos := []struct {
		nombre string
		estado backlog.Estado
		antes  int // 0 = sin estimar
		puntos int
	}{
		{"pendiente sin estimar", backlog.EstadoPendiente, 0, 5},
		{"pendiente ya estimado", backlog.EstadoPendiente, 3, 8},
		{"en progreso", backlog.EstadoEnProgreso, 0, 2},
		{"el mismo valor", backlog.EstadoPendiente, 5, 5},
		{"extremo inferior", backlog.EstadoPendiente, 0, 1},
		{"extremo superior", backlog.EstadoPendiente, 0, 13},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			item := itemEnEstado(t, c.estado, c.antes)

			estimado, err := item.Estimar(c.puntos)
			if err != nil {
				t.Fatalf("Estimar(%d): %v", c.puntos, err)
			}
			if valor, esta := puntosDe(t, estimado); valor != c.puntos || !esta {
				t.Errorf("story points = (%d, %t), se esperaba (%d, true)", valor, esta, c.puntos)
			}
		})
	}
}

// US-013 / RN-013-3: estimar cambia solo los Story Points; el resto del item
// queda igual.
func TestEstimar_NoCambiaElResto(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoEnProgreso, 3)

	estimado, err := item.Estimar(8)
	if err != nil {
		t.Fatalf("Estimar: %v", err)
	}

	esperado := item
	esperado.StoryPoints = estimado.StoryPoints
	if estimado.ID != esperado.ID || estimado.Numero != esperado.Numero ||
		estimado.ProyectoID != esperado.ProyectoID || estimado.Titulo != esperado.Titulo ||
		estimado.Descripcion != esperado.Descripcion || estimado.Prioridad != esperado.Prioridad ||
		estimado.Estado != esperado.Estado || estimado.CreadoEn != esperado.CreadoEn ||
		!slices.Equal(estimado.Criterios, esperado.Criterios) {
		t.Errorf("Estimar cambio algo mas que los story points:\n recibido: %+v\n esperado: %+v", estimado, esperado)
	}
}

// US-013 / CL-013-11: Estimar devuelve una copia: el item original conserva su
// estimacion, y los criterios no se comparten entre el original y la copia.
func TestEstimar_NoModificaElOriginal(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoPendiente, 3)

	estimado, err := item.Estimar(8)
	if err != nil {
		t.Fatalf("Estimar: %v", err)
	}

	if valor, _ := puntosDe(t, item); valor != 3 {
		t.Errorf("el item original quedo con %d story points, se esperaba 3", valor)
	}
	estimado.Criterios[0] = "cambiado"
	if item.Criterios[0] == "cambiado" {
		t.Error("el original y la copia comparten la lista de criterios")
	}
}

// US-013 / CA-013-4, CL-013-11: un valor fuera de la escala se rechaza, no
// devuelve item y el original conserva su estimacion anterior.
func TestEstimar_FueraDeEscala(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoPendiente, 3)

	estimado, err := item.Estimar(4)
	if !errors.Is(err, backlog.ErrEstimacionFueraDeEscala) {
		t.Fatalf("error = %v, se esperaba ErrEstimacionFueraDeEscala", err)
	}
	if estimado.Titulo != "" {
		t.Errorf("devolvio un item %+v, se esperaba el valor cero", estimado)
	}
	if valor, _ := puntosDe(t, item); valor != 3 {
		t.Errorf("el item original quedo con %d story points, se esperaba 3", valor)
	}
}

// US-013 / RN-013-4, CL-013-8, CA-013-5: un item completado no se estima ni se
// re-estima, porque cambiaria hacia atras los Story Points completados de un
// sprint ya medido.
func TestEstimar_ItemCompletado(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoCompletado, 5)

	estimado, err := item.Estimar(8)
	if !errors.Is(err, backlog.ErrItemCompletado) {
		t.Fatalf("error = %v, se esperaba ErrItemCompletado", err)
	}
	if estimado.Titulo != "" {
		t.Errorf("devolvio un item %+v, se esperaba el valor cero", estimado)
	}
	if valor, _ := puntosDe(t, item); valor != 5 {
		t.Errorf("el item original quedo con %d story points, se esperaba 5", valor)
	}
}

// US-013 / RN-013-5, CL-013-9: se valida primero el estado y despues el valor.
// Un item completado con un valor invalido devuelve solo ErrItemCompletado.
func TestEstimar_ItemCompletadoConValorInvalido(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoCompletado, 5)

	_, err := item.Estimar(4)
	if !errors.Is(err, backlog.ErrItemCompletado) {
		t.Fatalf("error = %v, se esperaba ErrItemCompletado", err)
	}
	if errors.Is(err, backlog.ErrEstimacionFueraDeEscala) {
		t.Errorf("el error %q tambien reporta el valor fuera de escala: debe ser solo ErrItemCompletado", err)
	}
}
