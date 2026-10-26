package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// escenarioPoker es la base migrada con un proyecto, una historia pendiente y los
// integrantes pedidos, en ese orden de alta.
type escenarioPoker struct {
	pool        *pgxpool.Pool
	repo        *postgres.RepositorioPlanningPoker
	proyecto    string
	item        string
	integrantes []string
}

func nuevoEscenarioPoker(ctx context.Context, t *testing.T, nombres ...string) escenarioPoker {
	t.Helper()
	pool := dbprueba.BaseMigrada(ctx, t)
	e := escenarioPoker{pool: pool, repo: postgres.NuevoRepositorioPlanningPoker(pool), proyecto: crearProyecto(ctx, t, pool, "Demo")}
	item, err := e.repo.RegistrarItem(ctx, e.proyecto, itemConTitulo(t, "Crear proyecto"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	e.item = item.ID
	for _, nombre := range nombres {
		e.integrantes = append(e.integrantes, e.integrante(ctx, t, e.proyecto, nombre))
	}
	return e
}

// integrante inserta un integrante con una fecha de alta posterior a la de los que ya
// tiene la base, para que el orden de alta sea fijo.
func (e escenarioPoker) integrante(ctx context.Context, t *testing.T, proyecto, nombre string) string {
	t.Helper()
	var id string
	const insertar = `
INSERT INTO integrantes (proyecto_id, nombre, email, rol, creado_en)
VALUES ($1, $2, $3, 'product_builder',
        '2026-01-01'::timestamptz + (SELECT count(*) FROM integrantes) * interval '1 minute')
RETURNING id::text`
	if err := e.pool.QueryRow(ctx, insertar, proyecto, nombre, nombre+"@ejemplo.test").Scan(&id); err != nil {
		t.Fatalf("insertar el integrante %q: %v", nombre, err)
	}
	return id
}

func (e escenarioPoker) sesion(t *testing.T) planningpoker.SesionPlanningPoker {
	t.Helper()
	sesion, err := planningpoker.NuevaSesion(e.item, planningpoker.DatosDeLaHistoria{}, e.integrantes)
	if err != nil {
		t.Fatalf("NuevaSesion: %v", err)
	}
	return sesion
}

// filaDeSesion es lo que quedo guardado de una sesion, leido de las tablas.
type filaDeSesion struct {
	itemID        string
	estado        string
	ronda         int
	participantes []string
}

func leerSesion(ctx context.Context, t *testing.T, pool *pgxpool.Pool, id string) filaDeSesion {
	t.Helper()
	var f filaDeSesion
	if err := pool.QueryRow(ctx, `SELECT item_id::text, estado, ronda FROM sesiones_planning_poker WHERE id = $1`, id).Scan(&f.itemID, &f.estado, &f.ronda); err != nil {
		t.Fatalf("leer la sesion %s: %v", id, err)
	}
	filas, err := pool.Query(ctx, `
SELECT p.integrante_id::text FROM participantes_planning_poker p JOIN integrantes i ON i.id = p.integrante_id
WHERE p.sesion_id = $1 ORDER BY i.creado_en, i.id`, id)
	if err != nil {
		t.Fatalf("leer los participantes: %v", err)
	}
	defer filas.Close()
	for filas.Next() {
		var p string
		if err := filas.Scan(&p); err != nil {
			t.Fatalf("leer un participante: %v", err)
		}
		f.participantes = append(f.participantes, p)
	}
	return f
}

// US-014 / RN-014-4: los integrantes del proyecto, en orden de alta. Los de otro
// proyecto no cuentan, y un ID mal formado no es de ningun proyecto.
func TestIDsDeIntegrantes(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioPoker(ctx, t, "ana", "beto", "caro")
	e.integrante(ctx, t, crearProyecto(ctx, t, e.pool, "Otro"), "dani")

	ids, err := e.repo.IDsDeIntegrantes(ctx, e.proyecto)
	if err != nil || !slices.Equal(ids, e.integrantes) {
		t.Errorf("IDsDeIntegrantes = (%q, %v), se esperaban %q", ids, err, e.integrantes)
	}
	if ids, err := e.repo.IDsDeIntegrantes(ctx, "no-es-un-uuid"); err != nil || len(ids) != 0 {
		t.Errorf("con un ID mal formado = (%q, %v), se esperaba ninguno", ids, err)
	}
}

// US-014 / CA-014-1, RN-014-4, RN-014-6, RN-014-12: la sesion queda guardada con un ID que
// genera la base, votando, en la ronda 1 y con una fila por participante. Desde ahi la
// historia tiene una sesion abierta.
func TestRegistrarSesion(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioPoker(ctx, t, "ana", "beto", "caro")
	if abierta, err := e.repo.TieneSesionAbierta(ctx, e.item); err != nil || abierta {
		t.Fatalf("TieneSesionAbierta antes de iniciar = (%t, %v), se esperaba (false, nil)", abierta, err)
	}

	registrada, err := e.repo.RegistrarSesion(ctx, e.proyecto, e.sesion(t))
	if err != nil {
		t.Fatalf("RegistrarSesion: %v", err)
	}

	if registrada.ID == "" || registrada.Estado != planningpoker.EstadoVotando || registrada.Ronda != 1 ||
		!slices.Equal(registrada.Participantes, e.integrantes) {
		t.Errorf("registrada = %+v, se esperaba con ID, votando, ronda 1 y los tres integrantes", registrada)
	}
	guardada := leerSesion(ctx, t, e.pool, registrada.ID)
	if guardada.itemID != e.item || guardada.estado != "votando" || guardada.ronda != 1 || !slices.Equal(guardada.participantes, e.integrantes) {
		t.Errorf("guardada = %+v, se esperaba la sesion registrada", guardada)
	}
	if abierta, _ := e.repo.TieneSesionAbierta(ctx, e.item); !abierta {
		t.Error("despues de registrarla, la historia no tiene una sesion abierta")
	}
}

// US-014 / CA-014-3, RN-014-4, CL-014-7: los participantes se leen dentro de la
// transaccion: son los integrantes del proyecto al registrar, y uno que se suma despues
// no participa.
func TestRegistrarSesion_ParticipantesAlIniciar(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioPoker(ctx, t, "ana", "beto")
	sesion := e.sesion(t)
	caro := e.integrante(ctx, t, e.proyecto, "caro")

	registrada, err := e.repo.RegistrarSesion(ctx, e.proyecto, sesion)
	if err != nil {
		t.Fatalf("RegistrarSesion: %v", err)
	}
	e.integrante(ctx, t, e.proyecto, "dani")

	esperados := append(slices.Clone(e.integrantes), caro)
	if got := leerSesion(ctx, t, e.pool, registrada.ID).participantes; !slices.Equal(got, esperados) {
		t.Errorf("participantes = %q, se esperaban %q", got, esperados)
	}
}

// US-014 / RN-014-7, RN-014-11, CL-014-5, CL-014-6, CL-014-11: el repositorio vuelve a
// verificar las reglas con la historia bloqueada. Una sesion votando o revelada esta
// abierta; una finalizada no impide iniciar otra. Una historia completada o un proyecto
// con un solo integrante se rechazan. En los rechazos no se guarda nada.
func TestRegistrarSesion_Reglas(t *testing.T) {
	ctx := contexto(t)

	for _, estado := range []string{"votando", "revelada"} {
		t.Run("sesion "+estado, func(t *testing.T) {
			e := nuevoEscenarioPoker(ctx, t, "ana", "beto")
			primera, err := e.repo.RegistrarSesion(ctx, e.proyecto, e.sesion(t))
			if err != nil {
				t.Fatalf("primera sesion: %v", err)
			}
			if _, err := e.pool.Exec(ctx, `UPDATE sesiones_planning_poker SET estado = $2 WHERE id = $1`, primera.ID, estado); err != nil {
				t.Fatalf("poner la sesion %s: %v", estado, err)
			}
			if _, err := e.repo.RegistrarSesion(ctx, e.proyecto, e.sesion(t)); !errors.Is(err, planningpoker.ErrSesionAbierta) {
				t.Errorf("segunda sesion: error = %v, se esperaba ErrSesionAbierta", err)
			}
			if n := dbprueba.ContarFilas(ctx, t, e.pool, "sesiones_planning_poker"); n != 1 {
				t.Errorf("hay %d sesiones, se esperaba 1", n)
			}
		})
	}

	t.Run("sesion finalizada", func(t *testing.T) {
		e := nuevoEscenarioPoker(ctx, t, "ana", "beto")
		primera, err := e.repo.RegistrarSesion(ctx, e.proyecto, e.sesion(t))
		if err != nil {
			t.Fatalf("primera sesion: %v", err)
		}
		if _, err := e.pool.Exec(ctx, `UPDATE sesiones_planning_poker SET estado = 'finalizada', valor_acordado = 5 WHERE id = $1`, primera.ID); err != nil {
			t.Fatalf("finalizar la sesion: %v", err)
		}
		if abierta, _ := e.repo.TieneSesionAbierta(ctx, e.item); abierta {
			t.Error("una sesion finalizada cuenta como abierta")
		}
		if _, err := e.repo.RegistrarSesion(ctx, e.proyecto, e.sesion(t)); err != nil {
			t.Errorf("sesion nueva despues de una finalizada: %v", err)
		}
	})

	t.Run("historia completada", func(t *testing.T) {
		e := nuevoEscenarioPoker(ctx, t, "ana", "beto")
		if _, err := e.pool.Exec(ctx, `UPDATE items_backlog SET estado = 'completado', completado_en = now() WHERE id = $1`, e.item); err != nil {
			t.Fatalf("completar la historia: %v", err)
		}
		if _, err := e.repo.RegistrarSesion(ctx, e.proyecto, e.sesion(t)); !errors.Is(err, backlog.ErrItemCompletado) {
			t.Errorf("error = %v, se esperaba ErrItemCompletado", err)
		}
		if n := dbprueba.ContarFilas(ctx, t, e.pool, "sesiones_planning_poker"); n != 0 {
			t.Errorf("hay %d sesiones, se esperaba ninguna", n)
		}
	})

	t.Run("un solo integrante", func(t *testing.T) {
		e := nuevoEscenarioPoker(ctx, t, "ana")
		sesion := planningpoker.SesionPlanningPoker{ItemID: e.item, Estado: planningpoker.EstadoVotando, Ronda: 1}
		if _, err := e.repo.RegistrarSesion(ctx, e.proyecto, sesion); !errors.Is(err, planningpoker.ErrParticipantesInsuficientes) {
			t.Errorf("error = %v, se esperaba ErrParticipantesInsuficientes", err)
		}
	})
}

// US-014 / CA-014-7, RN-014-1: una historia que no existe, que es de otro proyecto o con
// un ID mal formado se rechaza con ErrItemInexistente.
func TestRegistrarSesion_HistoriaInexistente(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioPoker(ctx, t, "ana", "beto")
	otro := crearProyecto(ctx, t, e.pool, "Otro")

	casos := []struct {
		nombre, proyecto, item string
	}{
		{"de otro proyecto", otro, e.item},
		{"inexistente", e.proyecto, idQueNoEsDeNingunItem},
		{"mal formado", e.proyecto, "no-es-un-uuid"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			sesion := e.sesion(t)
			sesion.ItemID = c.item
			if _, err := e.repo.RegistrarSesion(ctx, c.proyecto, sesion); !errors.Is(err, app.ErrItemInexistente) {
				t.Errorf("error = %v, se esperaba ErrItemInexistente", err)
			}
		})
	}
	if abierta, err := e.repo.TieneSesionAbierta(ctx, "no-es-un-uuid"); err != nil || abierta {
		t.Errorf("TieneSesionAbierta con un ID mal formado = (%t, %v), se esperaba (false, nil)", abierta, err)
	}
}

// US-014 / RN-014-13: el indice unico parcial es la red de seguridad de RN-014-7: la base
// rechaza una segunda sesion abierta de la misma historia aunque no pase por el
// repositorio.
func TestMigracion_UnaSesionAbiertaPorHistoria(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioPoker(ctx, t, "ana", "beto")
	insertar := `INSERT INTO sesiones_planning_poker (item_id, estado) VALUES ($1, $2)`
	// Desde US-018 una sesion finalizada tiene su valor acordado.
	if _, err := e.pool.Exec(ctx, `INSERT INTO sesiones_planning_poker (item_id, estado, valor_acordado) VALUES ($1, 'finalizada', 5)`, e.item); err != nil {
		t.Fatalf("insertar una finalizada: %v", err)
	}
	if _, err := e.pool.Exec(ctx, insertar, e.item, "votando"); err != nil {
		t.Fatalf("insertar la abierta: %v", err)
	}

	_, err := e.pool.Exec(ctx, insertar, e.item, "revelada")

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Errorf("error = %v, se esperaba la violacion del indice unico (23505)", err)
	}
}

// US-014 / RN-014-11, CL-014-8: dos inicios simultaneos sobre la misma historia esperan el
// candado de la historia. El segundo ve la sesion del primero y se rechaza con
// ErrSesionAbierta, nunca con un error del indice unico. Para que compitan de verdad, el
// test retiene el candado; el segundo inicio arranca recien cuando el primero ya espera,
// asi el orden es fijo.
func TestRegistrarSesion_Concurrente(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioPoker(ctx, t, "ana", "beto")

	candado, err := e.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir la transaccion del candado: %v", err)
	}
	defer func() { _ = candado.Rollback(ctx) }()
	if _, err := candado.Exec(ctx, `SELECT 1 FROM items_backlog WHERE id = $1 FOR NO KEY UPDATE`, e.item); err != nil {
		t.Fatalf("tomar el candado de la historia: %v", err)
	}

	errores := make([]chan error, 2)
	for i := range errores {
		errores[i] = make(chan error, 1)
		go func() {
			_, err := e.repo.RegistrarSesion(ctx, e.proyecto, e.sesion(t))
			errores[i] <- err
		}()
		if bloqueadas := esperarBloqueadas(ctx, t, e.pool, i+1); bloqueadas < i+1 {
			t.Fatalf("solo %d de %d inicios esperaron el candado de la historia", bloqueadas, i+1)
		}
	}
	if err := candado.Rollback(ctx); err != nil {
		t.Fatalf("soltar el candado: %v", err)
	}

	if err := <-errores[0]; err != nil {
		t.Errorf("primer inicio: %v", err)
	}
	segundo := <-errores[1]
	if !errors.Is(segundo, planningpoker.ErrSesionAbierta) {
		t.Errorf("segundo inicio: error = %v, se esperaba ErrSesionAbierta", segundo)
	}
	var pgErr *pgconn.PgError
	if errors.As(segundo, &pgErr) {
		t.Errorf("segundo inicio choco con la base (%s): %v", pgErr.Code, fmt.Sprint(segundo))
	}
}
