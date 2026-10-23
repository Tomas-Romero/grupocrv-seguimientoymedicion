package postgres_test

import (
	"context"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// T-015 / CA-T015-5: las lecturas de sprints cumplen los puertos de los casos de uso nuevos.
var (
	_ app.RepositorioListarSprints   = (*postgres.RepositorioAsignaciones)(nil)
	_ app.RepositorioConsultarSprint = (*postgres.RepositorioAsignaciones)(nil)
)

// ListarSprintsConItems devuelve cada sprint con las historias que se le asignaron, tambien los
// cerrados (la relacion es historica, RN-009-13); otro proyecto o un ID mal formado, lista vacia.
func TestListarSprintsConItems(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	if err := e.repo.AsignarItemASprint(ctx, e.item, e.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	// Se cierra el primer sprint para poder abrir el segundo, y se asigna otra historia.
	if _, err := e.pool.Exec(ctx, `UPDATE sprints SET estado = 'cerrado' WHERE id = $1`, e.sprint); err != nil {
		t.Fatalf("cerrar el sprint: %v", err)
	}
	segundo, err := e.sprints.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 20), diaSprint(11, 2)))
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}
	otra := e.itemParaSprint(ctx, t, e.proyecto, "Listar proyectos", true, 1)
	if err := e.repo.AsignarItemASprint(ctx, otra, segundo.ID); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}

	got, err := e.repo.ListarSprintsConItems(ctx, e.proyecto)
	if err != nil || len(got) != 2 {
		t.Fatalf("ListarSprintsConItems = %+v, %v; quiero dos sprints", got, err)
	}
	porID := map[string][]app.HistoriaAsignada{}
	for _, s := range got {
		porID[s.Sprint.ID] = s.Items
	}
	if len(porID[e.sprint]) != 1 || porID[e.sprint][0].Item.ID != e.item || porID[e.sprint][0].Item.Titulo != "Crear proyecto" {
		t.Errorf("historias del sprint cerrado = %+v, quiero Crear proyecto", porID[e.sprint])
	}
	if len(porID[segundo.ID]) != 1 || porID[segundo.ID][0].Item.ID != otra {
		t.Errorf("historias del sprint abierto = %+v, quiero Listar proyectos", porID[segundo.ID])
	}
	for _, s := range got {
		if s.Sprint.ProyectoID != e.proyecto || s.Sprint.Numero == 0 || s.Sprint.SprintGoal == "" {
			t.Errorf("sprint leido incompleto: %+v", s.Sprint)
		}
	}

	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "no-es-un-uuid", ""} {
		if vacio, err := e.repo.ListarSprintsConItems(ctx, id); err != nil || len(vacio) != 0 {
			t.Errorf("ListarSprintsConItems(%q) = %+v, %v; quiero lista vacia y sin error", id, vacio, err)
		}
	}
}

// Un sprint sin historias sale con la lista vacia, no se pierde.
func TestListarSprintsConItems_SinHistorias(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)

	got, err := e.repo.ListarSprintsConItems(ctx, e.proyecto)
	if err != nil || len(got) != 1 || len(got[0].Items) != 0 {
		t.Errorf("ListarSprintsConItems = %+v, %v; quiero un sprint sin historias", got, err)
	}
}

// ItemsDelSprint devuelve las historias con todos sus datos y sin importar el resultado de la fila.
func TestItemsDelSprint(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	if err := e.repo.AsignarItemASprint(ctx, e.item, e.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	if _, err := e.pool.Exec(ctx, `UPDATE sprint_items SET resultado = 'no_completado' WHERE item_id = $1`, e.item); err != nil {
		t.Fatalf("cambiar el resultado: %v", err)
	}

	got, err := e.repo.ItemsDelSprint(ctx, e.sprint)
	if err != nil || len(got) != 1 {
		t.Fatalf("ItemsDelSprint = %+v, %v; quiero una historia", got, err)
	}
	puntos, estimado := got[0].Item.StoryPoints.Valor()
	if got[0].Item.ID != e.item || got[0].Item.Numero != 1 || !estimado || puntos != 5 || len(got[0].Item.Criterios) != 1 || got[0].Item.ProyectoID != e.proyecto {
		t.Errorf("historia leida = %+v", got[0])
	}
	if got[0].Resultado != sprint.ResultadoNoCompletado {
		t.Errorf("resultado = %q, quiero el de la fila (no_completado)", got[0].Resultado)
	}
	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "no-es-un-uuid", ""} {
		if vacio, err := e.repo.ItemsDelSprint(ctx, id); err != nil || len(vacio) != 0 {
			t.Errorf("ItemsDelSprint(%q) = %+v, %v; quiero lista vacia y sin error", id, vacio, err)
		}
	}
}

// ListarItemsDisponibles ofrece las historias que no estan completadas ni en un sprint abierto,
// listas o no (RN-T015-6); una que estuvo en un sprint cerrado y no se completo vuelve a estar disponible.
func TestListarItemsDisponibles(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioAsignaciones(ctx, t)
	asignada := e.item
	if err := e.repo.AsignarItemASprint(ctx, asignada, e.sprint); err != nil {
		t.Fatalf("AsignarItemASprint: %v", err)
	}
	libre := e.itemParaSprint(ctx, t, e.proyecto, "Libre", true, 1)
	sinEstimar := e.itemParaSprint(ctx, t, e.proyecto, "Sin estimar", false, 0)
	completada := e.itemParaSprint(ctx, t, e.proyecto, "Completada", true, 1)
	if _, err := e.pool.Exec(ctx, `UPDATE items_backlog SET estado = 'completado', completado_en = now() WHERE id = $1`, completada); err != nil {
		t.Fatalf("completar la historia: %v", err)
	}
	otroProyecto := proyectoConFechas(ctx, t, e.pool, "Otro", diaSprint(10, 6), diaSprint(12, 31))
	ajena := e.itemParaSprint(ctx, t, otroProyecto, "De otro proyecto", true, 1)

	got, err := e.repo.ListarItemsDisponibles(ctx, e.proyecto)
	if err != nil {
		t.Fatalf("ListarItemsDisponibles: %v", err)
	}
	ids := map[string]bool{}
	for _, i := range got {
		ids[i.ID] = true
	}
	if !ids[libre] || !ids[sinEstimar] || len(ids) != 2 {
		t.Errorf("disponibles = %v, quiero solo la libre y la sin estimar (no %s, %s ni %s)", ids, asignada, completada, ajena)
	}

	// Al cerrar el sprint, la asignada no completada vuelve a estar disponible.
	if _, err := e.pool.Exec(ctx, `UPDATE sprint_items SET resultado = 'no_completado' WHERE item_id = $1`, asignada); err != nil {
		t.Fatalf("cerrar la fila: %v", err)
	}
	got, _ = e.repo.ListarItemsDisponibles(ctx, e.proyecto)
	ids = map[string]bool{}
	for _, i := range got {
		ids[i.ID] = true
	}
	if !ids[asignada] || len(ids) != 3 {
		t.Errorf("despues de cerrar: %v, quiero que vuelva la asignada", ids)
	}

	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "no-es-un-uuid", ""} {
		if vacio, err := e.repo.ListarItemsDisponibles(ctx, id); err != nil || len(vacio) != 0 {
			t.Errorf("ListarItemsDisponibles(%q) = %+v, %v; quiero lista vacia y sin error", id, vacio, err)
		}
	}
	_ = context.Background
}

// DEF-007: las lecturas traen el resultado de la fila de cada sprint. Una historia que no se completo en
// el primer sprint, volvio al backlog y se completo en el segundo sigue siendo "no completada" en el primero,
// aunque su estado actual sea completado.
func TestLecturasDeSprints_TraenElResultadoDeLaFila(t *testing.T) {
	ctx := contexto(t)
	e, otra := sprintParaCerrarEnPostgres(ctx, t)
	if _, err := e.repo.GuardarCierre(ctx, e.sprint); err != nil {
		t.Fatalf("GuardarCierre: %v", err)
	}
	siguiente, err := e.sprints.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, diaSprint(10, 20), diaSprint(11, 2)))
	if err != nil {
		t.Fatalf("RegistrarSprint del siguiente: %v", err)
	}
	if err := e.repo.AsignarItemASprint(ctx, otra, siguiente.ID); err != nil {
		t.Fatalf("asignar la devuelta al siguiente: %v", err)
	}
	if _, err := e.repo.GuardarCambioDeEstado(ctx, otra, siguiente.ID, completarEn(marcadoEn)); err != nil {
		t.Fatalf("completar la devuelta en el siguiente: %v", err)
	}

	resultados := func(historias []app.HistoriaAsignada) map[string]sprint.Resultado {
		m := map[string]sprint.Resultado{}
		for _, h := range historias {
			m[h.Item.ID] = h.Resultado
		}
		return m
	}

	primero, err := e.repo.ItemsDelSprint(ctx, e.sprint)
	if err != nil {
		t.Fatalf("ItemsDelSprint del primero: %v", err)
	}
	if got := resultados(primero); got[e.item] != sprint.ResultadoCompletado || got[otra] != sprint.ResultadoNoCompletado {
		t.Errorf("resultados del primer sprint = %v, quiero completado y no_completado", got)
	}
	segundo, err := e.repo.ItemsDelSprint(ctx, siguiente.ID)
	if err != nil {
		t.Fatalf("ItemsDelSprint del segundo: %v", err)
	}
	if got := resultados(segundo); got[otra] != sprint.ResultadoPendiente {
		t.Errorf("resultado en el sprint abierto = %v, quiero pendiente", got)
	}

	todos, err := e.repo.ListarSprintsConItems(ctx, e.proyecto)
	if err != nil {
		t.Fatalf("ListarSprintsConItems: %v", err)
	}
	for _, s := range todos {
		if s.Sprint.ID == e.sprint {
			if got := resultados(s.Items); got[e.item] != sprint.ResultadoCompletado || got[otra] != sprint.ResultadoNoCompletado {
				t.Errorf("ListarSprintsConItems, primer sprint = %v, quiero completado y no_completado", got)
			}
		}
	}
}
