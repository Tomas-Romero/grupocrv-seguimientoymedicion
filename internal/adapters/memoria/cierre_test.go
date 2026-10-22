package memoria_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// sprintConDosHistorias devuelve la base con dos historias asignadas al sprint abierto:
// la primera completada y la segunda no.
func sprintConDosHistorias(t *testing.T) (baseParaAsignar, string) {
	t.Helper()
	ctx := context.Background()
	b := asignado(t)
	otra := b.itemEn(t, b.proyecto, "Listar proyectos", true, 1)
	if err := b.base.AsignarItemASprint(ctx, otra, b.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	if _, err := b.base.GuardarCambioDeEstado(ctx, b.item, b.sprint, completar(marcadoEn)); err != nil {
		t.Fatalf("marcar: %v", err)
	}
	return b, otra
}

// historiaEnProgresoAsignada registra una historia en progreso (ninguna historia del
// backlog pone ese estado todavia), la estima y la asigna al sprint abierto.
func historiaEnProgresoAsignada(t *testing.T, b baseParaAsignar) string {
	t.Helper()
	ctx := context.Background()
	item := itemValido(t, "Editar proyectos")
	item.Estado = backlog.EstadoEnProgreso
	registrado, err := b.base.RegistrarItem(ctx, b.proyecto, item)
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	puntos, _ := backlog.NuevosStoryPoints(3)
	if err := b.base.GuardarEstimacion(ctx, registrado.ID, puntos); err != nil {
		t.Fatalf("GuardarEstimacion: %v", err)
	}
	if err := b.base.AsignarItemASprint(ctx, registrado.ID, b.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	return registrado.ID
}

// US-011 / RN-011-8: las historias del sprint son sus filas pendientes, con si cada una
// esta completada, en orden de asignacion.
func TestBase_HistoriasDelSprint(t *testing.T) {
	b, otra := sprintConDosHistorias(t)

	historias, err := b.base.HistoriasDelSprint(context.Background(), b.sprint)
	if err != nil {
		t.Fatalf("HistoriasDelSprint: %v", err)
	}

	esperadas := []sprint.HistoriaDelCierre{{ItemID: b.item, Completada: true}, {ItemID: otra, Completada: false}}
	if !slices.Equal(historias, esperadas) {
		t.Errorf("historias = %+v, se esperaban %+v", historias, esperadas)
	}
}

// US-011 / CA-011-1, RN-011-5, RN-011-6: cerrar deja el sprint cerrado, guarda el
// resultado de cada fila sin borrar ninguna, deja la completada como esta y pasa la no
// completada a pendiente y fuera del sprint (en el backlog).
func TestBase_GuardarCierre(t *testing.T) {
	ctx := context.Background()
	b, otra := sprintConDosHistorias(t)
	enProgreso := historiaEnProgresoAsignada(t, b)

	resultados, err := b.base.GuardarCierre(ctx, b.sprint)
	if err != nil {
		t.Fatalf("GuardarCierre: %v", err)
	}

	esperados := []sprint.ResultadoDelCierre{
		{ItemID: b.item, Resultado: sprint.ResultadoCompletado},
		{ItemID: otra, Resultado: sprint.ResultadoNoCompletado},
		{ItemID: enProgreso, Resultado: sprint.ResultadoNoCompletado},
	}
	if !slices.Equal(resultados, esperados) {
		t.Errorf("resultados = %+v, se esperaban %+v", resultados, esperados)
	}
	if s, _, _ := b.base.ObtenerSprint(ctx, b.sprint); s.Estado != sprint.EstadoCerrado {
		t.Errorf("estado del sprint = %q, se esperaba cerrado", s.Estado)
	}
	if filas := b.base.FilasDelSprint(b.sprint); len(filas) != 3 ||
		filas[0].Resultado != sprint.ResultadoCompletado || filas[1].Resultado != sprint.ResultadoNoCompletado {
		t.Errorf("filas del sprint = %+v, se esperaban las tres con su resultado", filas)
	}
	completada, _, _ := b.base.ObtenerItem(ctx, b.item)
	if completada.Estado != backlog.EstadoCompletado || !completada.CompletadoEn.Equal(marcadoEn) {
		t.Errorf("la completada cambio: %+v", completada)
	}
	for _, id := range []string{otra, enProgreso} {
		devuelta, _, _ := b.base.ObtenerItem(ctx, id)
		if devuelta.Estado != backlog.EstadoPendiente {
			t.Errorf("estado de la no completada %s = %q, se esperaba pendiente (CL-011-4)", id, devuelta.Estado)
		}
		if en, _ := b.base.ItemEnSprintAbierto(ctx, id); en {
			t.Errorf("la no completada %s sigue en un sprint abierto", id)
		}
	}
}

// US-011 / CA-011-4, RN-011-4: un sprint sin historias se cierra.
func TestBase_GuardarCierre_SinHistorias(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaAsignar(t)

	resultados, err := b.base.GuardarCierre(ctx, b.sprint)
	if err != nil || len(resultados) != 0 {
		t.Fatalf("GuardarCierre = (%+v, %v), se esperaba (ninguno, nil)", resultados, err)
	}
	if s, _, _ := b.base.ObtenerSprint(ctx, b.sprint); s.Estado != sprint.EstadoCerrado {
		t.Errorf("estado del sprint = %q, se esperaba cerrado", s.Estado)
	}
}

// US-011 / CA-011-5, CA-011-6, RN-011-2, RN-011-8, CL-011-6, CL-011-7: un sprint ya
// cerrado se rechaza con ErrSprintCerrado y uno que no esta, con ErrSprintInexistente.
// En los dos casos no cambia nada.
func TestBase_GuardarCierre_Rechazos(t *testing.T) {
	ctx := context.Background()
	b, otra := sprintConDosHistorias(t)
	if _, err := b.base.GuardarCierre(ctx, b.sprint); err != nil {
		t.Fatalf("primer cierre: %v", err)
	}

	_, err := b.base.GuardarCierre(ctx, b.sprint)
	if !errors.Is(err, sprint.ErrSprintCerrado) {
		t.Errorf("segundo cierre: error = %v, se esperaba ErrSprintCerrado", err)
	}
	_, err = b.base.GuardarCierre(ctx, idSinDuenoAsignaciones)
	if !errors.Is(err, app.ErrSprintInexistente) {
		t.Errorf("sprint inexistente: error = %v, se esperaba ErrSprintInexistente", err)
	}
	if filas := b.base.FilasDelSprint(b.sprint); len(filas) != 2 || filas[1].ItemID != otra ||
		filas[1].Resultado != sprint.ResultadoNoCompletado {
		t.Errorf("filas del sprint = %+v, el segundo cierre las cambio", filas)
	}
}

// US-011 / CA-011-2, RN-011-10, CL-011-8: despues de cerrar se crea el sprint siguiente
// y la historia devuelta se asigna a el: queda con dos filas, no_completado en el
// cerrado y pendiente en el nuevo.
func TestBase_GuardarCierre_HistoriaDevueltaSeAsignaAlSiguiente(t *testing.T) {
	ctx := context.Background()
	b, otra := sprintConDosHistorias(t)
	if _, err := b.base.GuardarCierre(ctx, b.sprint); err != nil {
		t.Fatalf("GuardarCierre: %v", err)
	}
	siguiente, err := b.base.RegistrarSprint(ctx, sprintParaProyecto(t, b.proyecto, diaDeSprint(10, 20), diaDeSprint(11, 2)))
	if err != nil {
		t.Fatalf("RegistrarSprint del siguiente: %v", err)
	}

	if err := b.base.AsignarItemASprint(ctx, otra, siguiente.ID); err != nil {
		t.Fatalf("asignar la devuelta al siguiente: %v", err)
	}

	if filas := b.base.FilasDelSprint(siguiente.ID); len(filas) != 1 || filas[0].Resultado != sprint.ResultadoPendiente {
		t.Errorf("filas del siguiente = %+v, se esperaba una pendiente", filas)
	}
	if filas := b.base.FilasDelSprint(b.sprint); len(filas) != 2 {
		t.Errorf("el cerrado tiene %d filas, se esperaban 2", len(filas))
	}
}

// FilasDelSprint devuelve cada historia que estuvo en el sprint con su resultado, en
// orden de asignacion, y con sus datos actuales (los Story Points, por ejemplo). Lo
// usan los escenarios de metricas sobre sprints cerrados.
func TestBase_FilasDelSprint(t *testing.T) {
	b, otra := sprintConDosHistorias(t)

	filas := b.base.FilasDelSprint(b.sprint)

	if len(filas) != 2 || filas[0].ItemID != b.item || filas[1].ItemID != otra {
		t.Fatalf("filas = %+v, se esperaban las dos historias en orden", filas)
	}
	if puntos, _ := filas[0].StoryPoints.Valor(); puntos != 5 || filas[0].Resultado != sprint.ResultadoPendiente {
		t.Errorf("fila = %+v, se esperaban 5 story points y pendiente", filas[0])
	}
	if filas := b.base.FilasDelSprint(idSinDuenoAsignaciones); len(filas) != 0 {
		t.Errorf("un sprint que no esta tiene %d filas", len(filas))
	}
}
