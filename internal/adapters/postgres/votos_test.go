package postgres_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// conSesionEnPostgres devuelve el escenario con una sesion registrada sobre la historia,
// con tres participantes.
func conSesionEnPostgres(ctx context.Context, t *testing.T) (escenarioPoker, planningpoker.SesionPlanningPoker) {
	t.Helper()
	e := nuevoEscenarioPoker(ctx, t, "ana", "beto", "caro")
	sesion, err := e.repo.RegistrarSesion(ctx, e.proyecto, e.sesion(t))
	if err != nil {
		t.Fatalf("RegistrarSesion: %v", err)
	}
	return e, sesion
}

// votosGuardados devuelve los votos de la sesion leidos de la tabla, por ronda e
// integrante.
func votosGuardados(ctx context.Context, t *testing.T, e escenarioPoker, sesionID string) map[int]map[string]int {
	t.Helper()
	filas, err := e.pool.Query(ctx, `SELECT ronda, integrante_id::text, valor FROM votos_planning_poker WHERE sesion_id = $1`, sesionID)
	if err != nil {
		t.Fatalf("leer los votos: %v", err)
	}
	defer filas.Close()
	votos := map[int]map[string]int{}
	for filas.Next() {
		var (
			ronda, valor int
			integrante   string
		)
		if err := filas.Scan(&ronda, &integrante, &valor); err != nil {
			t.Fatalf("leer un voto: %v", err)
		}
		if votos[ronda] == nil {
			votos[ronda] = map[string]int{}
		}
		votos[ronda][integrante] = valor
	}
	return votos
}

// US-015 / RN-015-1: la sesion se lee con el proyecto de su historia, su estado, su ronda
// y sus participantes en orden de alta. Una que no esta, o con un ID mal formado, es
// inexistente.
func TestObtenerSesion(t *testing.T) {
	ctx := contexto(t)
	e, registrada := conSesionEnPostgres(ctx, t)
	if registrada.ProyectoID != e.proyecto {
		t.Errorf("RegistrarSesion devolvio el proyecto %q, se esperaba %q", registrada.ProyectoID, e.proyecto)
	}

	sesion, existe, err := e.repo.ObtenerSesion(ctx, registrada.ID)
	if err != nil || !existe {
		t.Fatalf("ObtenerSesion = (existe %t, %v), se esperaba encontrarla", existe, err)
	}
	if sesion.ID != registrada.ID || sesion.ItemID != e.item || sesion.ProyectoID != e.proyecto ||
		sesion.Estado != planningpoker.EstadoVotando || sesion.Ronda != 1 || !slices.Equal(sesion.Participantes, e.integrantes) {
		t.Errorf("sesion = %+v, se esperaba la registrada con su proyecto", sesion)
	}
	for _, id := range []string{idQueNoEsDeNingunItem, "no-es-un-uuid"} {
		if _, existe, err := e.repo.ObtenerSesion(ctx, id); err != nil || existe {
			t.Errorf("ObtenerSesion(%q) = (existe %t, %v), se esperaba (false, nil)", id, existe, err)
		}
	}
}

// US-015 / RN-015-1: el integrante tiene que ser del proyecto (el chequeo de US-019).
func TestExisteIntegranteEnProyecto_DesdePlanningPoker(t *testing.T) {
	ctx := contexto(t)
	e, _ := conSesionEnPostgres(ctx, t)
	otro := e.integrante(ctx, t, crearProyecto(ctx, t, e.pool, "Otro"), "dani")

	if existe, err := e.repo.ExisteIntegranteEnProyecto(ctx, e.proyecto, e.integrantes[0]); err != nil || !existe {
		t.Errorf("integrante del proyecto = (%t, %v), se esperaba (true, nil)", existe, err)
	}
	if existe, err := e.repo.ExisteIntegranteEnProyecto(ctx, e.proyecto, otro); err != nil || existe {
		t.Errorf("integrante de otro proyecto = (%t, %v), se esperaba (false, nil)", existe, err)
	}
}

// US-015 / CA-015-1, RN-015-6: el voto se guarda en la ronda actual, y la votacion que
// se devuelve dice quien voto sin dejar leer los valores.
func TestRegistrarVoto(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	ana := e.integrantes[0]

	votacion, err := e.repo.RegistrarVoto(ctx, sesion.ID, ana, 5)
	if err != nil {
		t.Fatalf("RegistrarVoto: %v", err)
	}

	if !slices.Equal(votacion.Votaron(), []string{ana}) || !slices.Equal(votacion.Faltan(), e.integrantes[1:]) {
		t.Errorf("votaron %q y faltan %q, se esperaba ana y los otros dos", votacion.Votaron(), votacion.Faltan())
	}
	if _, err := votacion.Valores(); !errors.Is(err, planningpoker.ErrVotosOcultos) {
		t.Errorf("la votacion deja leer los valores: %v", err)
	}
	if votos := votosGuardados(ctx, t, e, sesion.ID); votos[1][ana] != 5 || len(votos[1]) != 1 {
		t.Errorf("votos guardados = %v, se esperaba el de ana en la ronda 1", votos)
	}
}

// US-015 / CA-015-2, RN-015-5, CL-015-3: votar otra vez reemplaza el voto: queda una sola
// fila, con el ultimo valor.
func TestRegistrarVoto_CambiarElVoto(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	ana := e.integrantes[0]
	for _, valor := range []int{3, 8} {
		if _, err := e.repo.RegistrarVoto(ctx, sesion.ID, ana, valor); err != nil {
			t.Fatalf("RegistrarVoto(%d): %v", valor, err)
		}
	}

	if votos := votosGuardados(ctx, t, e, sesion.ID); votos[1][ana] != 8 || len(votos[1]) != 1 {
		t.Errorf("votos guardados = %v, se esperaba solo el ultimo de ana (8)", votos)
	}
	if n := dbprueba.ContarFilas(ctx, t, e.pool, "votos_planning_poker"); n != 1 {
		t.Errorf("hay %d votos, se esperaba 1", n)
	}
}

// US-015 / RN-015-6: los votos van por ronda. En una ronda nueva (US-017) la votacion
// empieza sin votos y los de la ronda anterior no cambian.
func TestRegistrarVoto_PorRonda(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	ana, beto := e.integrantes[0], e.integrantes[1]
	if _, err := e.repo.RegistrarVoto(ctx, sesion.ID, ana, 3); err != nil {
		t.Fatalf("voto de la ronda 1: %v", err)
	}
	if _, err := e.pool.Exec(ctx, `UPDATE sesiones_planning_poker SET ronda = 2 WHERE id = $1`, sesion.ID); err != nil {
		t.Fatalf("pasar a la ronda 2: %v", err)
	}

	votacion, err := e.repo.RegistrarVoto(ctx, sesion.ID, beto, 5)
	if err != nil {
		t.Fatalf("voto de la ronda 2: %v", err)
	}

	if !slices.Equal(votacion.Votaron(), []string{beto}) {
		t.Errorf("en la ronda 2 votaron %q, se esperaba solo beto", votacion.Votaron())
	}
	votos := votosGuardados(ctx, t, e, sesion.ID)
	if votos[1][ana] != 3 || len(votos[1]) != 1 || votos[2][beto] != 5 || len(votos[2]) != 1 {
		t.Errorf("votos guardados = %v, se esperaba ana en la 1 y beto en la 2", votos)
	}
}

// US-015 / RN-015-9, CL-015-6, CL-015-9: el repositorio vuelve a verificar con la sesion
// bloqueada. Un integrante que no participa, un valor fuera de la escala o una sesion que
// ya no esta votando (por ejemplo, revelada entre la lectura del caso de uso y el voto) se
// rechazan, y no se guarda nada.
func TestRegistrarVoto_Reglas(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	ana := e.integrantes[0]
	dani := e.integrante(ctx, t, e.proyecto, "dani")

	if _, err := e.repo.RegistrarVoto(ctx, sesion.ID, dani, 5); !errors.Is(err, planningpoker.ErrNoEsParticipante) {
		t.Errorf("no participa: error = %v, se esperaba ErrNoEsParticipante", err)
	}
	if _, err := e.repo.RegistrarVoto(ctx, sesion.ID, ana, 4); !errors.Is(err, backlog.ErrEstimacionFueraDeEscala) {
		t.Errorf("fuera de escala: error = %v, se esperaba ErrEstimacionFueraDeEscala", err)
	}
	if _, err := e.pool.Exec(ctx, `UPDATE sesiones_planning_poker SET estado = 'revelada' WHERE id = $1`, sesion.ID); err != nil {
		t.Fatalf("revelar la sesion: %v", err)
	}
	if _, err := e.repo.RegistrarVoto(ctx, sesion.ID, ana, 5); !errors.Is(err, planningpoker.ErrVotacionCerrada) {
		t.Errorf("sesion revelada: error = %v, se esperaba ErrVotacionCerrada", err)
	}
	if n := dbprueba.ContarFilas(ctx, t, e.pool, "votos_planning_poker"); n != 0 {
		t.Errorf("hay %d votos, se esperaba ninguno", n)
	}
}

// US-015 / CA-015-7: una sesion que no existe, o con un ID mal formado, se rechaza con
// ErrSesionInexistente.
func TestRegistrarVoto_SesionInexistente(t *testing.T) {
	ctx := contexto(t)
	e, _ := conSesionEnPostgres(ctx, t)

	for _, id := range []string{idQueNoEsDeNingunItem, "no-es-un-uuid"} {
		if _, err := e.repo.RegistrarVoto(ctx, id, e.integrantes[0], 5); !errors.Is(err, app.ErrSesionInexistente) {
			t.Errorf("RegistrarVoto(%q) = %v, se esperaba ErrSesionInexistente", id, err)
		}
	}
}

// US-015 / RN-015-11: la clave foranea hacia los participantes es la red de seguridad de
// RN-015-2: la base rechaza el voto de quien no participa aunque no pase por el
// repositorio.
func TestMigracion_SoloVotanLosParticipantes(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	dani := e.integrante(ctx, t, e.proyecto, "dani")

	_, err := e.pool.Exec(ctx, `INSERT INTO votos_planning_poker (sesion_id, ronda, integrante_id, valor) VALUES ($1, 1, $2, 5)`, sesion.ID, dani)

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Errorf("error = %v, se esperaba la violacion de la clave foranea (23503)", err)
	}
}

// bloquearSesion abre una transaccion del test que retiene el candado de la sesion.
func bloquearSesion(ctx context.Context, t *testing.T, e escenarioPoker, sesionID string) interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
} {
	t.Helper()
	candado, err := e.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir la transaccion del candado: %v", err)
	}
	t.Cleanup(func() { _ = candado.Rollback(context.Background()) })
	if _, err := candado.Exec(ctx, `SELECT 1 FROM sesiones_planning_poker WHERE id = $1 FOR NO KEY UPDATE`, sesionID); err != nil {
		t.Fatalf("tomar el candado de la sesion: %v", err)
	}
	return candado
}

// US-015 / RN-015-10, CL-015-8: dos votos simultaneos del mismo participante esperan el
// candado de la sesion. Los dos terminan bien y cuenta uno solo, el ultimo que confirma.
// El segundo voto arranca recien cuando el primero ya espera, asi el orden es fijo.
func TestRegistrarVoto_DosVotosSimultaneos(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	ana := e.integrantes[0]
	candado := bloquearSesion(ctx, t, e, sesion.ID)

	errores := make([]chan error, 2)
	for i, valor := range []int{3, 8} {
		errores[i] = make(chan error, 1)
		go func() {
			_, err := e.repo.RegistrarVoto(ctx, sesion.ID, ana, valor)
			errores[i] <- err
		}()
		if bloqueadas := esperarBloqueadas(ctx, t, e.pool, i+1); bloqueadas < i+1 {
			t.Fatalf("solo %d de %d votos esperaron el candado de la sesion", bloqueadas, i+1)
		}
	}
	if err := candado.Rollback(ctx); err != nil {
		t.Fatalf("soltar el candado: %v", err)
	}

	for i := range errores {
		if err := <-errores[i]; err != nil {
			t.Errorf("voto %d: %v", i+1, err)
		}
	}
	if votos := votosGuardados(ctx, t, e, sesion.ID); votos[1][ana] != 8 || len(votos[1]) != 1 {
		t.Errorf("votos guardados = %v, se esperaba uno solo de ana, el ultimo (8)", votos)
	}
}

// US-015 / RN-015-10, CL-015-9: un voto que llega mientras alguien revela (US-016) espera
// el candado de la sesion. Cuando lo obtiene, la sesion ya esta revelada y se rechaza con
// ErrVotacionCerrada. El test hace de quien revela: toma el candado, espera a que el voto
// quede bloqueado, pasa la sesion a revelada y confirma.
func TestRegistrarVoto_MientrasSeRevela(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	revelado := bloquearSesion(ctx, t, e, sesion.ID)

	voto := make(chan error, 1)
	go func() {
		_, err := e.repo.RegistrarVoto(ctx, sesion.ID, e.integrantes[0], 5)
		voto <- err
	}()
	if bloqueadas := esperarBloqueadas(ctx, t, e.pool, 1); bloqueadas < 1 {
		t.Fatal("el voto no espero el candado de la sesion")
	}
	if _, err := revelado.Exec(ctx, `UPDATE sesiones_planning_poker SET estado = 'revelada' WHERE id = $1`, sesion.ID); err != nil {
		t.Fatalf("revelar: %v", err)
	}
	if err := revelado.Commit(ctx); err != nil {
		t.Fatalf("confirmar el revelado: %v", err)
	}

	if err := <-voto; !errors.Is(err, planningpoker.ErrVotacionCerrada) {
		t.Errorf("voto: error = %v, se esperaba ErrVotacionCerrada", err)
	}
	if n := dbprueba.ContarFilas(ctx, t, e.pool, "votos_planning_poker"); n != 0 {
		t.Errorf("hay %d votos, se esperaba ninguno", n)
	}
}
