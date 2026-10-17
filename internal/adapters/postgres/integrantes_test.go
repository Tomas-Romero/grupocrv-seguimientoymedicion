package postgres_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// integranteDePrueba arma un integrante como lo entrega el dominio: ya validado,
// con el email en minuscula y sin ID.
func integranteDePrueba(proyectoID, email string) integrante.Integrante {
	return integrante.Integrante{
		ProyectoID: proyectoID,
		Nombre:     "Ana",
		Email:      email,
		Rol:        integrante.RolProductBuilder,
	}
}

// filaIntegrante es lo que quedo guardado de un integrante, leido directamente
// de la tabla.
type filaIntegrante struct {
	proyectoID string
	nombre     string
	email      string
	rol        string
	creadoEn   time.Time
}

func leerIntegrante(ctx context.Context, t *testing.T, pool *pgxpool.Pool, id string) filaIntegrante {
	t.Helper()
	var f filaIntegrante
	const consulta = `
SELECT proyecto_id::text, nombre, email, rol, creado_en
FROM integrantes WHERE id = $1`
	if err := pool.QueryRow(ctx, consulta, id).Scan(&f.proyectoID, &f.nombre, &f.email, &f.rol, &f.creadoEn); err != nil {
		t.Fatalf("leer el integrante %s: %v", id, err)
	}
	return f
}

// US-003 / RN-003-10, CL-003-12: integrantes.id tiene DEFAULT gen_random_uuid()
// (migracion 00005), asi que un INSERT sin ID es valido y la base lo genera.
func TestTablaIntegrantes_IDPorDefecto(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")

	var id string
	const insertar = `
INSERT INTO integrantes (proyecto_id, nombre, email, rol)
VALUES ($1, 'Ana', 'ana@ejemplo.test', 'product_builder')
RETURNING id::text`
	if err := pool.QueryRow(ctx, insertar, proyecto).Scan(&id); err != nil {
		t.Fatalf("INSERT sin ID: %v; falta el DEFAULT de integrantes.id", err)
	}
	if id == "" {
		t.Error("la base no genero el ID del integrante")
	}
}

// US-003 / CA-003-1, RN-003-10: el integrante queda guardado con los datos
// recibidos y un ID que genera la base.
func TestRegistrarIntegrante_Guarda(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	repo := postgres.NuevoRepositorioIntegrantes(pool)

	registrado, err := repo.RegistrarIntegrante(ctx, integranteDePrueba(proyecto, "ana@ejemplo.test"))
	if err != nil {
		t.Fatalf("RegistrarIntegrante: %v", err)
	}
	if registrado.ID == "" {
		t.Fatal("el integrante registrado no tiene ID: lo genera la base")
	}
	if registrado.ProyectoID != proyecto || registrado.Nombre != "Ana" ||
		registrado.Email != "ana@ejemplo.test" || registrado.Rol != integrante.RolProductBuilder {
		t.Errorf("integrante devuelto = %+v, se esperaban los datos recibidos", registrado)
	}

	guardado := leerIntegrante(ctx, t, pool, registrado.ID)
	if guardado.proyectoID != proyecto || guardado.nombre != "Ana" ||
		guardado.email != "ana@ejemplo.test" || guardado.rol != "product_builder" {
		t.Errorf("fila guardada = %+v, se esperaban los datos recibidos", guardado)
	}
	if guardado.creadoEn.IsZero() {
		t.Error("creado_en quedo vacio: lo completa el DEFAULT de la base")
	}
}

// US-003 / CA-003-2, CL-003-8, RN-003-11: se pueden registrar varios integrantes
// en el mismo proyecto, con el mismo nombre y el mismo rol, y cada uno recibe un
// ID distinto.
func TestRegistrarIntegrante_Varios(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	repo := postgres.NuevoRepositorioIntegrantes(pool)

	ids := map[string]bool{}
	for _, email := range []string{"uno@ejemplo.test", "dos@ejemplo.test", "tres@ejemplo.test"} {
		registrado, err := repo.RegistrarIntegrante(ctx, integranteDePrueba(proyecto, email))
		if err != nil {
			t.Fatalf("RegistrarIntegrante(%s): %v", email, err)
		}
		if ids[registrado.ID] {
			t.Errorf("el ID %q se repite", registrado.ID)
		}
		ids[registrado.ID] = true
	}

	if n := dbprueba.ContarFilas(ctx, t, pool, "integrantes"); n != 3 {
		t.Errorf("integrantes = %d, se esperaban 3", n)
	}
}

// US-003 / CA-003-8, CL-003-6, RN-003-7, RN-003-9: el email repetido en el mismo
// proyecto lo rechaza la restriccion UNIQUE (SQLSTATE 23505); el repositorio lo
// traduce a ErrEmailDuplicado indicando el email y no guarda nada.
func TestRegistrarIntegrante_EmailDuplicado(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	repo := postgres.NuevoRepositorioIntegrantes(pool)
	if _, err := repo.RegistrarIntegrante(ctx, integranteDePrueba(proyecto, "tomas@ejemplo.test")); err != nil {
		t.Fatalf("primer alta: %v", err)
	}

	_, err := repo.RegistrarIntegrante(ctx, integranteDePrueba(proyecto, "tomas@ejemplo.test"))
	if !errors.Is(err, app.ErrEmailDuplicado) {
		t.Fatalf("error = %v, se esperaba ErrEmailDuplicado", err)
	}
	if !strings.Contains(err.Error(), "tomas@ejemplo.test") {
		t.Errorf("el mensaje %q no indica el email", err.Error())
	}
	if n := dbprueba.ContarFilas(ctx, t, pool, "integrantes"); n != 1 {
		t.Errorf("integrantes = %d, se esperaba 1: el rechazo no guarda nada", n)
	}
}

// US-003 / CA-003-3, CL-003-7: el mismo email es valido en otro proyecto.
func TestRegistrarIntegrante_MismoEmailEnOtroProyecto(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	demo := crearProyecto(ctx, t, pool, "Demo")
	otro := crearProyecto(ctx, t, pool, "Otro")
	repo := postgres.NuevoRepositorioIntegrantes(pool)

	for _, proyecto := range []string{demo, otro} {
		if _, err := repo.RegistrarIntegrante(ctx, integranteDePrueba(proyecto, "ana@ejemplo.test")); err != nil {
			t.Fatalf("RegistrarIntegrante en %s: %v", proyecto, err)
		}
	}

	if n := dbprueba.ContarFilas(ctx, t, pool, "integrantes"); n != 2 {
		t.Errorf("integrantes = %d, se esperaban 2", n)
	}
}

// US-003 / CA-003-9, RN-003-9: si el proyecto no existe, RegistrarIntegrante
// devuelve ErrProyectoInexistente y no guarda nada. Un UUID que no esta en la
// base lo rechaza la clave foranea (SQLSTATE 23503); un ID mal formado, la
// conversion a UUID (22P02).
func TestRegistrarIntegrante_ProyectoInexistente(t *testing.T) {
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
			repo := postgres.NuevoRepositorioIntegrantes(pool)

			_, err := repo.RegistrarIntegrante(ctx, integranteDePrueba(c.id, "ana@ejemplo.test"))
			if !errors.Is(err, app.ErrProyectoInexistente) {
				t.Fatalf("error = %v, se esperaba ErrProyectoInexistente", err)
			}
			if n := dbprueba.ContarFilas(ctx, t, pool, "integrantes"); n != 0 {
				t.Errorf("integrantes = %d, se esperaba ninguno", n)
			}
		})
	}
}

// US-003 / RN-003-9: ExisteProyecto es la lectura simple del caso de uso. Un ID
// que no esta en la base o que no es un UUID valido es un proyecto inexistente,
// no un error.
func TestRepositorioIntegrantes_ExisteProyecto(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	repo := postgres.NuevoRepositorioIntegrantes(pool)

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

// US-003 / CL-003-14, RN-003-9: con varias altas simultaneas del mismo email en
// el mismo proyecto, la restriccion UNIQUE deja pasar una sola: una termina bien
// y las demas devuelven ErrEmailDuplicado. Las altas esperan una senal comun para
// salir juntas.
func TestRegistrarIntegrante_AltasSimultaneas(t *testing.T) {
	const altas = 10
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	repo := postgres.NuevoRepositorioIntegrantes(pool)

	errs := make([]error, altas)
	salida := make(chan struct{})
	var espera sync.WaitGroup
	for n := 0; n < altas; n++ {
		espera.Add(1)
		go func(n int) {
			defer espera.Done()
			<-salida
			_, errs[n] = repo.RegistrarIntegrante(ctx, integranteDePrueba(proyecto, "ana@ejemplo.test"))
		}(n)
	}
	close(salida)
	espera.Wait()

	exitosas := 0
	for _, err := range errs {
		switch {
		case err == nil:
			exitosas++
		case !errors.Is(err, app.ErrEmailDuplicado):
			t.Errorf("error inesperado: %v", err)
		}
	}
	if exitosas != 1 {
		t.Errorf("altas exitosas = %d, se esperaba 1", exitosas)
	}
	if n := dbprueba.ContarFilas(ctx, t, pool, "integrantes"); n != 1 {
		t.Errorf("integrantes = %d, se esperaba 1", n)
	}
}

// US-003 / RN-003-3: el CHECK de la tabla es la ultima defensa del rol. Un rol
// que no es uno de los tres (sin pasar por el dominio) hace fallar el INSERT con
// SQLSTATE 23514 y no deja nada guardado.
func TestRegistrarIntegrante_RolInvalidoLoRechazaLaTabla(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	proyecto := crearProyecto(ctx, t, pool, "Demo")
	repo := postgres.NuevoRepositorioIntegrantes(pool)

	invalido := integranteDePrueba(proyecto, "ana@ejemplo.test")
	invalido.Rol = "administrador"
	_, err := repo.RegistrarIntegrante(ctx, invalido)
	if err == nil {
		t.Fatal("se esperaba que el INSERT fallara por el CHECK del rol")
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
		t.Errorf("error = %v, se esperaba una violacion de CHECK (23514)", err)
	}
	if n := dbprueba.ContarFilas(ctx, t, pool, "integrantes"); n != 0 {
		t.Errorf("integrantes = %d, se esperaba ninguno", n)
	}
}
