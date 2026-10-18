package postgres_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// registrarParaEditar registra un item con descripcion y criterios y lo
// devuelve con su ID.
func registrarParaEditar(ctx context.Context, t *testing.T, pool *pgxpool.Pool, repo *postgres.RepositorioBacklog, proyecto, titulo string) backlog.ItemBacklog {
	t.Helper()
	registrado, err := repo.RegistrarItem(ctx, proyecto, nuevoItem(t, backlog.DatosItem{
		Titulo:      titulo,
		Descripcion: "Alta de proyectos",
		Prioridad:   backlog.PrioridadMust,
		Criterios:   []string{"Se guarda", "Se lista"},
	}))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	return registrado
}

// editarCon devuelve el item editado por el dominio con el titulo pedido.
func editarCon(t *testing.T, item backlog.ItemBacklog, titulo string) backlog.ItemBacklog {
	t.Helper()
	editado, err := item.Editar(backlog.DatosItem{
		Titulo:      titulo,
		Descripcion: "Con su equipo",
		Prioridad:   backlog.PrioridadCould,
		Criterios:   []string{"Se guardan los integrantes"},
	})
	if err != nil {
		t.Fatalf("Editar: %v", err)
	}
	return editado
}

// US-006 / CA-006-1, RN-006-2, RN-006-4, RN-006-8: GuardarEdicion escribe solo el
// titulo, la descripcion, la prioridad y los criterios del item pedido. El
// estado, los Story Points, el numero y la fecha de alta quedan como estaban
// aunque el item recibido traiga otros; el otro item no cambia.
func TestGuardarEdicion(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	repo := postgres.NuevoRepositorioBacklog(pool)
	primero := registrarParaEditar(ctx, t, pool, repo, proyecto, "Crear proyecto")
	segundo := registrarParaEditar(ctx, t, pool, repo, proyecto, "Listar proyectos")
	cinco, _ := backlog.NuevosStoryPoints(5)
	if err := repo.GuardarEstimacion(ctx, primero.ID, cinco); err != nil {
		t.Fatalf("GuardarEstimacion: %v", err)
	}
	editado := editarCon(t, primero, "Crear proyecto con integrantes")
	editado.Estado = backlog.EstadoEnProgreso
	editado.StoryPoints = backlog.SinEstimar()
	editado.Numero = 99

	if err := repo.GuardarEdicion(ctx, editado); err != nil {
		t.Fatalf("GuardarEdicion: %v", err)
	}

	guardado := leerFila(ctx, t, pool, primero.ID)
	if guardado.titulo != "Crear proyecto con integrantes" || guardado.descripcion != "Con su equipo" ||
		guardado.prioridad != "could" || !slices.Equal(guardado.criterios, []string{"Se guardan los integrantes"}) {
		t.Errorf("guardado = %+v, se esperaban los datos editados", guardado)
	}
	if guardado.estado != "pendiente" || guardado.storyPoints == nil || *guardado.storyPoints != 5 ||
		guardado.numero != 1 || !guardado.creadoEn.Equal(creadoEn) || guardado.proyectoID != proyecto {
		t.Errorf("se escribio algo mas que los cuatro datos: %+v", guardado)
	}
	if otro := leerFila(ctx, t, pool, segundo.ID); otro.titulo != "Listar proyectos" || otro.prioridad != "must" {
		t.Errorf("la edicion se guardo tambien en otro item: %+v", otro)
	}
}

// US-006 / CA-006-4, CL-006-2: una lista de criterios vacia reemplaza a la
// anterior y queda como arreglo vacio, no nulo (la columna es NOT NULL). Vale
// tanto para la lista vacia que arma el dominio como para una nula.
func TestGuardarEdicion_CriteriosVacios(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioBacklog(pool)
	registrado := registrarParaEditar(ctx, t, pool, repo, crearProyecto(ctx, t, pool, "Demo"), "Crear proyecto")

	for _, criterios := range [][]string{{}, nil} {
		editado := editarCon(t, registrado, "Crear proyecto")
		editado.Criterios = criterios

		if err := repo.GuardarEdicion(ctx, editado); err != nil {
			t.Fatalf("GuardarEdicion(%q): %v", criterios, err)
		}
		if guardado := leerFila(ctx, t, pool, registrado.ID); guardado.criterios == nil || len(guardado.criterios) != 0 {
			t.Errorf("criterios = %#v, se esperaba el arreglo vacio", guardado.criterios)
		}
	}
}

// US-006 / CA-006-7, RN-006-7: guardar la edicion de un item que no esta, o con
// un ID que no es un UUID, devuelve ErrItemInexistente.
func TestGuardarEdicion_ItemInexistente(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioBacklog(pool)
	registrado := registrarParaEditar(ctx, t, pool, repo, crearProyecto(ctx, t, pool, "Demo"), "Crear proyecto")

	for _, id := range []string{idQueNoEsDeNingunItem, "no-es-un-uuid"} {
		editado := editarCon(t, registrado, "Otro titulo")
		editado.ID = id

		err := repo.GuardarEdicion(ctx, editado)
		if !errors.Is(err, app.ErrItemInexistente) {
			t.Errorf("GuardarEdicion(%q) = %v, se esperaba ErrItemInexistente", id, err)
		}
	}
	if guardado := leerFila(ctx, t, pool, registrado.ID); guardado.titulo != "Crear proyecto" {
		t.Errorf("se cambio un item existente: %+v", guardado)
	}
}

// US-006 / CA-006-6, RN-006-8, CL-006-7: GuardarEdicion no cambia un item que ya
// esta completado: el UPDATE esta condicionado al estado, no actualiza ninguna
// fila y el repositorio devuelve ErrItemCompletado. Cubre la ventana entre leer
// el item y guardar la edicion, igual que la estimacion de US-013.
func TestGuardarEdicion_ItemCompletado(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioBacklog(pool)
	registrado := registrarParaEditar(ctx, t, pool, repo, crearProyecto(ctx, t, pool, "Demo"), "Crear proyecto")
	editado := editarCon(t, registrado, "Crear proyecto con integrantes")
	if _, err := pool.Exec(ctx, `UPDATE items_backlog SET estado = 'completado' WHERE id = $1`, registrado.ID); err != nil {
		t.Fatalf("completar el item: %v", err)
	}

	err := repo.GuardarEdicion(ctx, editado)

	if !errors.Is(err, backlog.ErrItemCompletado) {
		t.Fatalf("error = %v, se esperaba ErrItemCompletado", err)
	}
	if errors.Is(err, app.ErrItemInexistente) {
		t.Errorf("un item completado no es un item inexistente: %v", err)
	}
	guardado := leerFila(ctx, t, pool, registrado.ID)
	if guardado.titulo != "Crear proyecto" || guardado.prioridad != "must" ||
		!slices.Equal(guardado.criterios, []string{"Se guarda", "Se lista"}) {
		t.Errorf("se edito un item completado: %+v", guardado)
	}
}

// US-006 / RN-006-9, CL-006-8: dos ediciones simultaneas del mismo item terminan
// las dos bien y queda la ultima. Para que compitan de verdad, el test retiene el
// candado de la fila con una transaccion propia. La segunda edicion arranca
// recien cuando la primera ya espera, asi el orden en que obtienen el candado es
// fijo; despues el test lo suelta.
func TestGuardarEdicion_Concurrente(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioBacklog(pool)
	registrado := registrarParaEditar(ctx, t, pool, repo, crearProyecto(ctx, t, pool, "Demo"), "Crear proyecto")

	candado, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir la transaccion del candado: %v", err)
	}
	defer func() { _ = candado.Rollback(ctx) }()
	if _, err := candado.Exec(ctx, `SELECT 1 FROM items_backlog WHERE id = $1 FOR UPDATE`, registrado.ID); err != nil {
		t.Fatalf("tomar el candado del item: %v", err)
	}

	errores := make(chan error, 2)
	for i, titulo := range []string{"Primera edicion", "Segunda edicion"} {
		editado := editarCon(t, registrado, titulo)
		go func() { errores <- repo.GuardarEdicion(ctx, editado) }()
		if bloqueadas := esperarBloqueadas(ctx, t, pool, i+1); bloqueadas < i+1 {
			t.Fatalf("solo %d de %d ediciones esperaron el candado de la fila", bloqueadas, i+1)
		}
	}
	if err := candado.Rollback(ctx); err != nil {
		t.Fatalf("soltar el candado: %v", err)
	}

	for range 2 {
		if err := <-errores; err != nil {
			t.Errorf("GuardarEdicion: %v", err)
		}
	}
	if guardado := leerFila(ctx, t, pool, registrado.ID); guardado.titulo != "Segunda edicion" {
		t.Errorf("titulo = %q, se esperaba el de la ultima edicion", guardado.titulo)
	}
}
