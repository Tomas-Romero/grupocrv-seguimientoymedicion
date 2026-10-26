package postgres_test

import (
	"reflect"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
)

// US-021 / sección 2: cada historia del proyecto llega con las horas de todos sus registros, en
// orden de fecha, y la que no tiene horas llega sin ellas. Nada de otro proyecto.
func TestLeerHistoriasConHoras(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)
	if _, err := e.repo.RegistrarEsfuerzo(ctx, e.registro(t, ayerDelEsfuerzo, "Desarrollo", 2.5)); err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}
	if _, err := e.repo.RegistrarEsfuerzo(ctx, e.registro(t, ayerDelEsfuerzo.AddDate(0, 0, -1), "Analisis", 1)); err != nil {
		t.Fatalf("RegistrarEsfuerzo: %v", err)
	}
	backlogRepo := postgres.NuevoRepositorioBacklog(e.pool)
	sinHoras, err := backlogRepo.RegistrarItem(ctx, e.proyecto, itemConTitulo(t, "Listar proyectos"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	if _, err := backlogRepo.RegistrarItem(ctx, e.otroProyecto, itemConTitulo(t, "Historia ajena")); err != nil {
		t.Fatalf("RegistrarItem ajeno: %v", err)
	}

	historias, err := e.repo.LeerHistoriasConHoras(ctx, e.proyecto)
	if err != nil {
		t.Fatalf("LeerHistoriasConHoras: %v", err)
	}

	if len(historias) != 2 {
		t.Fatalf("historias = %d, se esperaban las 2 del proyecto", len(historias))
	}
	for _, h := range historias {
		switch h.Item.ID {
		case e.item:
			if h.Item.Titulo != "Crear proyecto" || !reflect.DeepEqual(h.Horas, []float64{1, 2.5}) {
				t.Errorf("%q = %v horas, se esperaba [1 2.5] (por fecha)", h.Item.Titulo, h.Horas)
			}
		case sinHoras.ID:
			if len(h.Horas) != 0 {
				t.Errorf("%q = %v horas, se esperaba ninguna", h.Item.Titulo, h.Horas)
			}
		default:
			t.Errorf("historia inesperada %q: es de otro proyecto", h.Item.Titulo)
		}
	}
}

// US-021 / CL-021-13: un proyecto sin historias, uno que no existe y un ID mal formado dan una lista
// vacia, no un error.
func TestLeerHistoriasConHoras_SinHistorias(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)

	for nombre, proyecto := range map[string]string{"proyecto vacio": e.otroProyecto, "id sin dueno": idSinDueno, "id mal formado": "no-es-un-uuid"} {
		historias, err := e.repo.LeerHistoriasConHoras(ctx, proyecto)
		if err != nil || len(historias) != 0 {
			t.Errorf("%s: historias = %v, error = %v; se esperaba una lista vacia", nombre, historias, err)
		}
	}
}

// US-021: una historia estimada y completada llega con su estimacion y su estado actuales.
func TestLeerHistoriasConHoras_LlevaEstimacionYEstado(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioEsfuerzo(ctx, t)
	if _, err := e.pool.Exec(ctx, `UPDATE items_backlog SET story_points = 5, estado = 'completado', completado_en = now() WHERE id = $1`, e.item); err != nil {
		t.Fatalf("preparar la historia: %v", err)
	}

	historias, err := e.repo.LeerHistoriasConHoras(ctx, e.proyecto)
	if err != nil || len(historias) != 1 {
		t.Fatalf("LeerHistoriasConHoras = %v, %v; se esperaba una historia", historias, err)
	}

	puntos, estimado := historias[0].Item.StoryPoints.Valor()
	if !estimado || puntos != 5 || historias[0].Item.Estado != "completado" {
		t.Errorf("historia = %+v, se esperaba estimada con 5 y completada", historias[0].Item)
	}
}
