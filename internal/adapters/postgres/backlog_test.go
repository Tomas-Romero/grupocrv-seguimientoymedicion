package postgres_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// creadoEn es la fecha de creacion de los items de estos tests.
var creadoEn = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

// contexto devuelve un contexto con un limite holgado para un test de
// integracion.
func contexto(t *testing.T) context.Context {
	t.Helper()
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancelar)
	return ctx
}

// crearProyecto inserta un proyecto y devuelve su ID. El ID se genera en el
// INSERT porque proyectos.id todavia no tiene DEFAULT (lo agrega US-001).
func crearProyecto(ctx context.Context, t *testing.T, pool *pgxpool.Pool, nombre string) string {
	t.Helper()
	var id string
	const insertar = `
INSERT INTO proyectos (id, nombre, fecha_inicio, fecha_fin)
VALUES (gen_random_uuid(), $1, '2026-01-01', '2026-12-31')
RETURNING id::text`
	if err := pool.QueryRow(ctx, insertar, nombre).Scan(&id); err != nil {
		t.Fatalf("crear el proyecto %q: %v", nombre, err)
	}
	return id
}

// nuevoItem arma un item valido con el dominio, como lo haria el caso de uso.
func nuevoItem(t *testing.T, datos backlog.DatosItem) backlog.ItemBacklog {
	t.Helper()
	item, err := backlog.NuevoItem(datos, creadoEn)
	if err != nil {
		t.Fatalf("NuevoItem: %v", err)
	}
	return item
}

// itemConTitulo es un item valido, con prioridad must y sin criterios.
func itemConTitulo(t *testing.T, titulo string) backlog.ItemBacklog {
	t.Helper()
	return nuevoItem(t, backlog.DatosItem{Titulo: titulo, Prioridad: backlog.PrioridadMust})
}

// fila es lo que quedo guardado de un item, leido directamente de la tabla.
type fila struct {
	proyectoID  string
	numero      int
	titulo      string
	descripcion string
	prioridad   string
	estado      string
	storyPoints *int32
	criterios   []string
	creadoEn    time.Time
}

func leerFila(ctx context.Context, t *testing.T, pool *pgxpool.Pool, id string) fila {
	t.Helper()
	var f fila
	const consulta = `
SELECT proyecto_id::text, numero, titulo, descripcion, prioridad, estado, story_points, criterios, creado_en
FROM items_backlog WHERE id = $1`
	err := pool.QueryRow(ctx, consulta, id).Scan(&f.proyectoID, &f.numero, &f.titulo, &f.descripcion,
		&f.prioridad, &f.estado, &f.storyPoints, &f.criterios, &f.creadoEn)
	if err != nil {
		t.Fatalf("leer el item %s: %v", id, err)
	}
	return f
}

// US-005 / CA-005-1, RN-005-6, RN-005-7, RN-005-8, RN-005-10, RN-005-15,
// CL-005-10, CL-005-14: el primer item de un proyecto queda guardado con el
// numero 1, un ID que genera la base, en estado pendiente, sin Story Points
// (nulo, no 0) y con sus criterios en orden, repetidos incluidos.
func TestRegistrarItem_PrimerItem(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	repo := postgres.NuevoRepositorioBacklog(pool)
	item := nuevoItem(t, backlog.DatosItem{
		Titulo:      "Crear proyecto",
		Descripcion: "Alta de proyectos",
		Prioridad:   backlog.PrioridadShould,
		Criterios:   []string{"Se guarda", "Se guarda", "Se lista"},
	})

	registrado, err := repo.RegistrarItem(ctx, proyecto, item)
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	if registrado.ID == "" || registrado.Numero != 1 || registrado.ProyectoID != proyecto {
		t.Fatalf("ID, numero y proyecto = %q, %d, %q; se esperaba un ID de la base, 1 y %q",
			registrado.ID, registrado.Numero, registrado.ProyectoID, proyecto)
	}

	guardado := leerFila(ctx, t, pool, registrado.ID)
	esperado := fila{
		proyectoID:  proyecto,
		numero:      1,
		titulo:      "Crear proyecto",
		descripcion: "Alta de proyectos",
		prioridad:   "should",
		estado:      "pendiente",
		criterios:   []string{"Se guarda", "Se guarda", "Se lista"},
	}
	if guardado.proyectoID != esperado.proyectoID || guardado.numero != esperado.numero ||
		guardado.titulo != esperado.titulo || guardado.descripcion != esperado.descripcion ||
		guardado.prioridad != esperado.prioridad || guardado.estado != esperado.estado {
		t.Errorf("fila guardada = %+v, se esperaba %+v", guardado, esperado)
	}
	if guardado.storyPoints != nil {
		t.Errorf("story_points = %d, se esperaba nulo (sin estimar)", *guardado.storyPoints)
	}
	if !slices.Equal(guardado.criterios, esperado.criterios) {
		t.Errorf("criterios = %q, se esperaba %q", guardado.criterios, esperado.criterios)
	}
	if !guardado.creadoEn.Equal(creadoEn) {
		t.Errorf("creado_en = %v, se esperaba %v", guardado.creadoEn, creadoEn)
	}
}

// US-005 / RN-005-9, RN-005-11, CA-005-8: si el proyecto no existe, RegistrarItem
// devuelve ErrProyectoInexistente y no guarda nada. Un ID mal formado (SQLSTATE
// 22P02) tambien es un proyecto inexistente.
func TestRegistrarItem_ProyectoInexistente(t *testing.T) {
	casos := []struct {
		nombre string
		id     string
	}{
		{"UUID que no esta en la base", "00000000-0000-0000-0000-000000000000"},
		{"ID mal formado", "no-es-un-uuid"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			ctx := contexto(t)
			pool := dbprueba.BaseMigrada(ctx, t)
			repo := postgres.NuevoRepositorioBacklog(pool)

			_, err := repo.RegistrarItem(ctx, c.id, itemConTitulo(t, "Alta de proyectos"))
			if !errors.Is(err, app.ErrProyectoInexistente) {
				t.Fatalf("error = %v, se esperaba ErrProyectoInexistente", err)
			}
			if n := dbprueba.ContarFilas(ctx, t, pool, "items_backlog"); n != 0 {
				t.Errorf("items_backlog = %d, se esperaba ninguno", n)
			}
		})
	}
}

// US-005 / RN-005-9: ExisteProyecto es la lectura simple del caso de uso. Un ID
// que no esta en la base o que no es un UUID valido es un proyecto inexistente,
// no un error.
func TestExisteProyecto(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	repo := postgres.NuevoRepositorioBacklog(pool)

	casos := []struct {
		nombre string
		id     string
		existe bool
	}{
		{"proyecto de la base", proyecto, true},
		{"UUID que no esta en la base", "00000000-0000-0000-0000-000000000000", false},
		{"ID mal formado", "no-es-un-uuid", false},
		{"ID vacio", "", false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			existe, err := repo.ExisteProyecto(ctx, c.id)
			if err != nil {
				t.Fatalf("ExisteProyecto: %v", err)
			}
			if existe != c.existe {
				t.Errorf("ExisteProyecto(%q) = %t, se esperaba %t", c.id, existe, c.existe)
			}
		})
	}
}

// US-005 / CA-005-3, RN-005-10, CL-005-8, CA-005-2: los items de un proyecto se
// numeran 1, 2, ...; el primero de otro proyecto es el 1 aunque el primero ya
// tenga items. Un item sin criterios se guarda con una lista vacia, no nula.
func TestRegistrarItem_Correlativo(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	demo := crearProyecto(ctx, t, pool, "Demo")
	otro := crearProyecto(ctx, t, pool, "Otro")
	repo := postgres.NuevoRepositorioBacklog(pool)

	altas := []struct {
		proyecto string
		titulo   string
		numero   int
	}{
		{demo, "Alta de proyectos", 1},
		{demo, "Listar proyectos", 2},
		{otro, "Alta de sprints", 1},
	}
	for _, a := range altas {
		registrado, err := repo.RegistrarItem(ctx, a.proyecto, itemConTitulo(t, a.titulo))
		if err != nil {
			t.Fatalf("RegistrarItem(%q): %v", a.titulo, err)
		}
		if registrado.Numero != a.numero {
			t.Errorf("%q: numero = %d, se esperaba %d", a.titulo, registrado.Numero, a.numero)
		}
		guardado := leerFila(ctx, t, pool, registrado.ID)
		if guardado.criterios == nil || len(guardado.criterios) != 0 {
			t.Errorf("%q: criterios = %#v, se esperaba una lista vacia", a.titulo, guardado.criterios)
		}
	}
}
