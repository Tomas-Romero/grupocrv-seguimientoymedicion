package postgres_test

import (
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// T-011 / CA-T011-5: ListarProyectos devuelve los proyectos con todos sus datos, igual que la
// memoria; sin proyectos, una lista vacia.
func TestListarProyectos(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioProyectos(pool)

	vacio, err := repo.ListarProyectos(ctx)
	if err != nil || len(vacio) != 0 {
		t.Fatalf("sin proyectos: %v, %v; quiero lista vacia y sin error", vacio, err)
	}

	enviado := proyectoDePrueba("Demo")
	registrado, err := repo.RegistrarProyecto(ctx, enviado)
	if err != nil {
		t.Fatalf("RegistrarProyecto: %v", err)
	}
	if _, err := repo.RegistrarProyecto(ctx, proyectoDePrueba("Otro")); err != nil {
		t.Fatalf("RegistrarProyecto: %v", err)
	}

	got, err := repo.ListarProyectos(ctx)
	if err != nil {
		t.Fatalf("ListarProyectos: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("hay %d proyectos, quiero 2: %+v", len(got), got)
	}
	for _, p := range got {
		if p.ID != registrado.ID {
			continue
		}
		if p.Nombre != "Demo" || p.Descripcion != enviado.Descripcion ||
			!p.FechaInicio.Equal(enviado.FechaInicio) || !p.FechaFin.Equal(enviado.FechaFin) {
			t.Errorf("proyecto leido = %+v, quiero los datos enviados %+v", p, enviado)
		}
		return
	}
	t.Errorf("el proyecto %s no esta en la lista: %+v", registrado.ID, got)
}

// ListarIntegrantes devuelve los del proyecto con todos sus datos; otro proyecto, un
// proyecto sin integrantes y un ID mal formado devuelven una lista vacia, sin error.
func TestListarIntegrantes(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioIntegrantes(pool)
	uno := crearProyecto(ctx, t, pool, "Uno")
	otro := crearProyecto(ctx, t, pool, "Otro")
	registrado, err := repo.RegistrarIntegrante(ctx, integranteDePrueba(uno, "ana@ejemplo.test"))
	if err != nil {
		t.Fatalf("RegistrarIntegrante: %v", err)
	}

	got, err := repo.ListarIntegrantes(ctx, uno)
	if err != nil || len(got) != 1 {
		t.Fatalf("ListarIntegrantes(uno) = %+v, %v; quiero un integrante", got, err)
	}
	if got[0] != registrado || got[0].Rol != integrante.RolProductBuilder {
		t.Errorf("integrante leido = %+v, quiero %+v", got[0], registrado)
	}
	for _, id := range []string{otro, "00000000-0000-0000-0000-000000000000", "no-es-un-uuid", ""} {
		got, err := repo.ListarIntegrantes(ctx, id)
		if err != nil || len(got) != 0 {
			t.Errorf("ListarIntegrantes(%q) = %+v, %v; quiero lista vacia y sin error", id, got, err)
		}
	}
}

// ListarItems devuelve los items del proyecto con numero, estado, estimacion, criterios y fecha
// de creacion; otro proyecto o un ID mal formado, una lista vacia.
func TestListarItems(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioBacklog(pool)
	uno := crearProyecto(ctx, t, pool, "Uno")
	otro := crearProyecto(ctx, t, pool, "Otro")

	sinCriterios, err := repo.RegistrarItem(ctx, uno, itemConTitulo(t, "Sin criterios"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	conCriterios, err := repo.RegistrarItem(ctx, uno, nuevoItem(t, backlog.DatosItem{
		Titulo: "Con criterios", Descripcion: "Detalle", Prioridad: backlog.PrioridadShould, Criterios: []string{"uno", "dos"},
	}))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE items_backlog SET story_points = 5 WHERE id = $1`, conCriterios.ID); err != nil {
		t.Fatalf("estimar la historia: %v", err)
	}

	got, err := repo.ListarItems(ctx, uno)
	if err != nil || len(got) != 2 {
		t.Fatalf("ListarItems(uno) = %+v, %v; quiero dos historias", got, err)
	}
	porID := map[string]backlog.ItemBacklog{}
	for _, i := range got {
		porID[i.ID] = i
	}
	a := porID[sinCriterios.ID]
	if a.Numero != 1 || a.ProyectoID != uno || a.Titulo != "Sin criterios" || a.Estado != backlog.EstadoPendiente || len(a.Criterios) != 0 {
		t.Errorf("historia sin criterios = %+v", a)
	}
	if _, estimado := a.StoryPoints.Valor(); estimado {
		t.Errorf("la historia sin estimar figura estimada: %+v", a.StoryPoints)
	}
	b := porID[conCriterios.ID]
	puntos, estimado := b.StoryPoints.Valor()
	if b.Numero != 2 || b.Prioridad != backlog.PrioridadShould || b.Descripcion != "Detalle" ||
		len(b.Criterios) != 2 || b.Criterios[0] != "uno" || b.Criterios[1] != "dos" || !estimado || puntos != 5 || b.CreadoEn.IsZero() {
		t.Errorf("historia con criterios = %+v", b)
	}

	for _, id := range []string{otro, "00000000-0000-0000-0000-000000000000", "no-es-un-uuid", ""} {
		got, err := repo.ListarItems(ctx, id)
		if err != nil || len(got) != 0 {
			t.Errorf("ListarItems(%q) = %+v, %v; quiero lista vacia y sin error", id, got, err)
		}
	}
}

// Las lecturas cumplen los puertos de los casos de uso nuevos.
var (
	_ app.RepositorioListarProyectos = (*postgres.RepositorioProyectos)(nil)
	_ app.RepositorioListarBacklog   = (*postgres.RepositorioBacklog)(nil)
	_ app.RepositorioConsultarItem   = (*postgres.RepositorioBacklog)(nil)
)
