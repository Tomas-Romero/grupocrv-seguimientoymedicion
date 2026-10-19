package backlog_test

import (
	"slices"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// Fechas de completado de estos tests: el dominio las recibe por parametro y no lee
// el reloj (RN-010-8).
var (
	completadoEn = time.Date(2026, 10, 20, 15, 30, 0, 0, time.UTC)
	otraFecha    = time.Date(2026, 10, 22, 9, 0, 0, 0, time.UTC)
)

// itemCompletado arma un item completado en la fecha pedida.
func itemCompletado(t *testing.T, en time.Time) backlog.ItemBacklog {
	t.Helper()
	return itemEnEstado(t, backlog.EstadoPendiente, 5).Completar(en)
}

// US-010 / CA-010-1, RN-010-5, CL-010-1: marcar deja el item completado con la fecha
// recibida, desde pendiente o en progreso.
func TestCompletar(t *testing.T) {
	for _, estado := range []backlog.Estado{backlog.EstadoPendiente, backlog.EstadoEnProgreso} {
		t.Run(string(estado), func(t *testing.T) {
			item := itemEnEstado(t, estado, 5)

			completado := item.Completar(completadoEn)

			if completado.Estado != backlog.EstadoCompletado {
				t.Errorf("estado = %q, se esperaba %q", completado.Estado, backlog.EstadoCompletado)
			}
			if !completado.CompletadoEn.Equal(completadoEn) {
				t.Errorf("completado en = %v, se esperaba %v", completado.CompletadoEn, completadoEn)
			}
		})
	}
}

// US-010 / RN-010-5: marcar cambia solo el estado y la fecha; el resto del item queda
// igual.
func TestCompletar_NoCambiaElResto(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoEnProgreso, 8)

	completado := item.Completar(completadoEn)

	if completado.ID != item.ID || completado.Numero != item.Numero || completado.ProyectoID != item.ProyectoID ||
		completado.Titulo != item.Titulo || completado.Descripcion != item.Descripcion ||
		completado.Prioridad != item.Prioridad || completado.StoryPoints != item.StoryPoints ||
		!completado.CreadoEn.Equal(item.CreadoEn) || !slices.Equal(completado.Criterios, item.Criterios) {
		t.Errorf("Completar cambio algo mas que el estado y la fecha:\n recibido: %+v\n original: %+v", completado, item)
	}
}

// US-010 / CA-010-3, RN-010-6, CL-010-2: marcar un item que ya esta completado no
// cambia nada: conserva la fecha original.
func TestCompletar_YaCompletado(t *testing.T) {
	item := itemCompletado(t, completadoEn)

	otraVez := item.Completar(otraFecha)

	if otraVez.Estado != backlog.EstadoCompletado || !otraVez.CompletadoEn.Equal(completadoEn) {
		t.Errorf("marcar otra vez dejo (%q, %v), se esperaba (completado, %v)", otraVez.Estado, otraVez.CompletadoEn, completadoEn)
	}
}

// US-010 / CA-010-2, RN-010-7, CL-010-3: desmarcar deja el item pendiente y sin
// fecha, aunque antes de completarse estuviera en progreso.
func TestReabrir(t *testing.T) {
	for _, antes := range []backlog.Estado{backlog.EstadoPendiente, backlog.EstadoEnProgreso} {
		t.Run(string(antes), func(t *testing.T) {
			item := itemEnEstado(t, antes, 5).Completar(completadoEn)

			reabierto := item.Reabrir()

			if reabierto.Estado != backlog.EstadoPendiente {
				t.Errorf("estado = %q, se esperaba %q", reabierto.Estado, backlog.EstadoPendiente)
			}
			if !reabierto.CompletadoEn.IsZero() {
				t.Errorf("completado en = %v, se esperaba sin fecha", reabierto.CompletadoEn)
			}
			if reabierto.Titulo != item.Titulo || reabierto.StoryPoints != item.StoryPoints {
				t.Errorf("Reabrir cambio algo mas que el estado y la fecha: %+v", reabierto)
			}
		})
	}
}

// US-010 / RN-010-7, CL-010-4: desmarcar un item que no esta completado no cambia
// nada: un item en progreso sigue en progreso.
func TestReabrir_NoCompletado(t *testing.T) {
	for _, estado := range []backlog.Estado{backlog.EstadoPendiente, backlog.EstadoEnProgreso} {
		t.Run(string(estado), func(t *testing.T) {
			item := itemEnEstado(t, estado, 5)

			reabierto := item.Reabrir()

			if reabierto.Estado != estado || !reabierto.CompletadoEn.IsZero() {
				t.Errorf("desmarcar dejo (%q, %v), se esperaba (%q, sin fecha)", reabierto.Estado, reabierto.CompletadoEn, estado)
			}
		})
	}
}

// US-010 / spec seccion 3: un item nuevo no tiene fecha de completado.
func TestNuevoItem_SinFechaDeCompletado(t *testing.T) {
	if item := itemEnEstado(t, backlog.EstadoPendiente, 0); !item.CompletadoEn.IsZero() {
		t.Errorf("completado en = %v, se esperaba sin fecha", item.CompletadoEn)
	}
}

// Completar y Reabrir devuelven una copia: el item recibido no cambia y no comparte
// los criterios con la copia.
func TestCompletarYReabrir_NoModificanElOriginal(t *testing.T) {
	item := itemEnEstado(t, backlog.EstadoPendiente, 5)

	completado := item.Completar(completadoEn)
	if item.Estado != backlog.EstadoPendiente || !item.CompletadoEn.IsZero() {
		t.Errorf("Completar modifico el original: %+v", item)
	}
	completado.Criterios[0] = "cambiado"
	if item.Criterios[0] == "cambiado" {
		t.Error("Completar: el original y la copia comparten la lista de criterios")
	}

	reabierto := completado.Reabrir()
	if completado.Estado != backlog.EstadoCompletado {
		t.Errorf("Reabrir modifico el original: %+v", completado)
	}
	reabierto.Criterios[0] = "otra vez"
	if completado.Criterios[0] == "otra vez" {
		t.Error("Reabrir: el original y la copia comparten la lista de criterios")
	}
}
