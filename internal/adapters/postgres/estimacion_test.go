package postgres_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// idQueNoEsDeNingunItem es un UUID bien formado que ningun item tiene.
const idQueNoEsDeNingunItem = "00000000-0000-0000-0000-000000009999"

// US-013 / RN-013-6: ObtenerItem devuelve el item completo tal como se
// registro (con su ID, numero, proyecto, estado, criterios y fecha de alta) y
// sin estimar: el nulo de la columna vuelve como SinEstimar, no como 0.
func TestObtenerItem(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	repo := postgres.NuevoRepositorioBacklog(pool)
	registrado, err := repo.RegistrarItem(ctx, proyecto, nuevoItem(t, backlog.DatosItem{
		Titulo:      "Crear proyecto",
		Descripcion: "Alta de proyectos",
		Prioridad:   backlog.PrioridadShould,
		Criterios:   []string{"Se guarda", "Se lista"},
	}))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}

	item, existe, err := repo.ObtenerItem(ctx, registrado.ID)
	if err != nil || !existe {
		t.Fatalf("ObtenerItem = (existe %t, err %v), se esperaba encontrarlo", existe, err)
	}

	if item.ID != registrado.ID || item.Numero != 1 || item.ProyectoID != proyecto ||
		item.Titulo != "Crear proyecto" || item.Descripcion != "Alta de proyectos" ||
		item.Prioridad != backlog.PrioridadShould || item.Estado != backlog.EstadoPendiente {
		t.Errorf("item = %+v, se esperaba el que se registro", item)
	}
	if !slices.Equal(item.Criterios, []string{"Se guarda", "Se lista"}) {
		t.Errorf("criterios = %q, se esperaban los dos registrados", item.Criterios)
	}
	if !item.CreadoEn.Equal(creadoEn) {
		t.Errorf("creado_en = %v, se esperaba %v", item.CreadoEn, creadoEn)
	}
	if item.StoryPoints != backlog.SinEstimar() {
		t.Errorf("story points = %+v, se esperaba SinEstimar() (nulo en la base)", item.StoryPoints)
	}
}

// US-013 / RN-013-4: el estado que guarda la base (por ejemplo, completado, que
// cambia US-010) es el que vuelve, para que el dominio pueda rechazar la
// estimacion.
func TestObtenerItem_LeeElEstado(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioBacklog(pool)
	registrado, err := repo.RegistrarItem(ctx, crearProyecto(ctx, t, pool, "Demo"), itemConTitulo(t, "Crear proyecto"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE items_backlog SET estado = 'completado' WHERE id = $1`, registrado.ID); err != nil {
		t.Fatalf("completar el item: %v", err)
	}

	item, _, err := repo.ObtenerItem(ctx, registrado.ID)
	if err != nil {
		t.Fatalf("ObtenerItem: %v", err)
	}
	if item.Estado != backlog.EstadoCompletado {
		t.Errorf("estado = %q, se esperaba %q", item.Estado, backlog.EstadoCompletado)
	}
}

// US-013 / CA-013-6: un ID que ningun item tiene, o que no es un UUID, es un
// item inexistente y no un error.
func TestObtenerItem_Inexistente(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioBacklog(pool)

	for _, id := range []string{idQueNoEsDeNingunItem, "no-es-un-uuid", ""} {
		if _, existe, err := repo.ObtenerItem(ctx, id); err != nil || existe {
			t.Errorf("ObtenerItem(%q) = (existe %t, err %v), se esperaba (false, nil)", id, existe, err)
		}
	}
}

// US-013 / RN-013-7, CL-013-6: GuardarEstimacion escribe solo la columna
// story_points del item pedido, y una segunda estimacion reemplaza a la primera.
func TestGuardarEstimacion(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	repo := postgres.NuevoRepositorioBacklog(pool)
	primero, err := repo.RegistrarItem(ctx, proyecto, itemConTitulo(t, "Crear proyecto"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	segundo, err := repo.RegistrarItem(ctx, proyecto, itemConTitulo(t, "Listar proyectos"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}

	for _, valor := range []int{3, 13} {
		puntos, err := backlog.NuevosStoryPoints(valor)
		if err != nil {
			t.Fatalf("NuevosStoryPoints(%d): %v", valor, err)
		}
		if err := repo.GuardarEstimacion(ctx, primero.ID, puntos); err != nil {
			t.Fatalf("GuardarEstimacion(%d): %v", valor, err)
		}

		guardado := leerFila(ctx, t, pool, primero.ID)
		if guardado.storyPoints == nil || int(*guardado.storyPoints) != valor {
			t.Errorf("story_points = %v, se esperaba %d", guardado.storyPoints, valor)
		}
	}

	guardado := leerFila(ctx, t, pool, primero.ID)
	if guardado.titulo != "Crear proyecto" || guardado.numero != 1 || guardado.estado != "pendiente" {
		t.Errorf("se cambio algo mas que los story points: %+v", guardado)
	}
	if otro := leerFila(ctx, t, pool, segundo.ID); otro.storyPoints != nil {
		t.Errorf("la estimacion se guardo tambien en otro item: story_points = %d", *otro.storyPoints)
	}
}

// US-013 / CA-013-6: guardar la estimacion de un item que no esta, o con un ID
// que no es un UUID, devuelve ErrItemInexistente.
func TestGuardarEstimacion_ItemInexistente(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioBacklog(pool)
	puntos, err := backlog.NuevosStoryPoints(5)
	if err != nil {
		t.Fatalf("NuevosStoryPoints: %v", err)
	}

	for _, id := range []string{idQueNoEsDeNingunItem, "no-es-un-uuid"} {
		err := repo.GuardarEstimacion(ctx, id, puntos)
		if !errors.Is(err, app.ErrItemInexistente) {
			t.Errorf("GuardarEstimacion(%q) = %v, se esperaba ErrItemInexistente", id, err)
		}
	}
}

// Guardar SinEstimar deja la columna nula: el nulo es "sin estimar", no 0. El
// caso de uso de US-013 no lo hace (RN-013-2), pero el repositorio escribe
// fielmente lo que recibe.
func TestGuardarEstimacion_SinEstimarDejaNulo(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioBacklog(pool)
	registrado, err := repo.RegistrarItem(ctx, crearProyecto(ctx, t, pool, "Demo"), itemConTitulo(t, "Crear proyecto"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	cinco, _ := backlog.NuevosStoryPoints(5)
	if err := repo.GuardarEstimacion(ctx, registrado.ID, cinco); err != nil {
		t.Fatalf("GuardarEstimacion(5): %v", err)
	}

	if err := repo.GuardarEstimacion(ctx, registrado.ID, backlog.SinEstimar()); err != nil {
		t.Fatalf("GuardarEstimacion(SinEstimar): %v", err)
	}

	if guardado := leerFila(ctx, t, pool, registrado.ID); guardado.storyPoints != nil {
		t.Errorf("story_points = %d, se esperaba nulo", *guardado.storyPoints)
	}
}

// Un valor que la columna acepta pero la escala no (la tabla solo exige >= 0)
// es un dato corrupto: ObtenerItem lo informa con contexto y no lo oculta ni lo
// confunde con un item inexistente.
func TestObtenerItem_EstimacionFueraDeEscalaEnLaBase(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioBacklog(pool)
	registrado, err := repo.RegistrarItem(ctx, crearProyecto(ctx, t, pool, "Demo"), itemConTitulo(t, "Crear proyecto"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE items_backlog SET story_points = 4 WHERE id = $1`, registrado.ID); err != nil {
		t.Fatalf("poner una estimacion fuera de la escala: %v", err)
	}

	_, existe, err := repo.ObtenerItem(ctx, registrado.ID)

	if !errors.Is(err, backlog.ErrEstimacionFueraDeEscala) {
		t.Fatalf("error = %v, se esperaba ErrEstimacionFueraDeEscala", err)
	}
	if existe {
		t.Error("ObtenerItem informo que el item existe junto con un error")
	}
}
