package postgres_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// US-016: el repositorio cumple el puerto del caso de uso de revelar.
var _ app.RepositorioRevelar = (*postgres.RepositorioPlanningPoker)(nil)

// votarEn guarda el voto de cada integrante con el repositorio, en la ronda actual.
func votarEn(ctx context.Context, t *testing.T, e escenarioPoker, sesionID string, valores map[string]int) {
	t.Helper()
	for integrante, valor := range valores {
		if _, err := e.repo.RegistrarVoto(ctx, sesionID, integrante, valor); err != nil {
			t.Fatalf("RegistrarVoto de %s: %v", integrante, err)
		}
	}
}

func estadoYRondaGuardados(ctx context.Context, t *testing.T, e escenarioPoker, sesionID string) (string, int) {
	t.Helper()
	f := leerSesion(ctx, t, e.pool, sesionID)
	return f.estado, f.ronda
}

// US-016 / RN-016-7, RN-015-7: la votacion que entrega el repositorio es la de la ronda actual (los votos de
// una ronda anterior no entran) y, mientras se vota, no deja leer los valores: el puerto nunca los entrega en
// crudo. Una sesion que no esta, o con un ID mal formado, no tiene votos.
func TestVotacionDeLaRonda(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	ana, beto, caro := e.integrantes[0], e.integrantes[1], e.integrantes[2]
	votarEn(ctx, t, e, sesion.ID, map[string]int{ana: 5, beto: 8})

	v, err := e.repo.VotacionDeLaRonda(ctx, sesion.ID)
	if err != nil || !reflect.DeepEqual(v.Votaron(), []string{ana, beto}) || !reflect.DeepEqual(v.Faltan(), []string{caro}) {
		t.Fatalf("votaron %q y faltan %q (%v); se esperaba ana y beto, y falta caro", v.Votaron(), v.Faltan(), err)
	}
	if _, err := v.Valores(); !errors.Is(err, planningpoker.ErrVotosOcultos) {
		t.Errorf("la votacion deja leer los valores mientras se vota: %v", err)
	}

	// Revelada, los valores se pueden leer.
	if _, err := e.pool.Exec(ctx, `UPDATE sesiones_planning_poker SET estado = 'revelada' WHERE id = $1`, sesion.ID); err != nil {
		t.Fatalf("revelar la sesion: %v", err)
	}
	revelada, _ := e.repo.VotacionDeLaRonda(ctx, sesion.ID)
	if valores, err := revelada.Valores(); err != nil || !reflect.DeepEqual(valores, map[string]int{ana: 5, beto: 8}) {
		t.Errorf("Valores con la sesion revelada = %v, %v; se esperaba el voto de ana y de beto", valores, err)
	}

	if _, err := e.pool.Exec(ctx, `UPDATE sesiones_planning_poker SET estado = 'votando', ronda = 2 WHERE id = $1`, sesion.ID); err != nil {
		t.Fatalf("pasar a la ronda 2: %v", err)
	}
	v, err = e.repo.VotacionDeLaRonda(ctx, sesion.ID)
	if err != nil || len(v.Votaron()) != 0 || len(v.Faltan()) != 3 {
		t.Errorf("en la ronda 2 votaron %q y faltan %q (%v); se esperaba que no hubiera votos (los de la ronda 1 no entran)", v.Votaron(), v.Faltan(), err)
	}

	for _, id := range []string{idQueNoEsDeNingunItem, "no-es-un-uuid"} {
		if v, err := e.repo.VotacionDeLaRonda(ctx, id); err != nil || len(v.Votaron()) != 0 {
			t.Errorf("VotacionDeLaRonda(%q) votaron %q (%v); se esperaba una votacion vacia y sin error", id, v.Votaron(), err)
		}
	}
}

// US-016 / CA-016-1 a CA-016-3: revelar pasa la sesion a revelada, no cambia los votos y
// devuelve la revelacion con quienes votaron, quienes faltan y la divergencia.
func TestGuardarRevelado(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	ana, beto, caro := e.integrantes[0], e.integrantes[1], e.integrantes[2]
	votarEn(ctx, t, e, sesion.ID, map[string]int{ana: 5, beto: 8})

	r, err := e.repo.GuardarRevelado(ctx, sesion.ID, caro)
	if err != nil {
		t.Fatalf("GuardarRevelado: %v", err)
	}

	quiero := planningpoker.Revelacion{
		SesionID: sesion.ID, Ronda: 1,
		Votos:    []planningpoker.Voto{{ParticipanteID: ana, Valor: 5}, {ParticipanteID: beto, Valor: 8}},
		Faltaron: []string{caro}, Divergencia: true, Minimo: 5, Maximo: 8,
	}
	if !reflect.DeepEqual(r, quiero) {
		t.Errorf("revelacion = %+v, se esperaba %+v", r, quiero)
	}
	if estado, ronda := estadoYRondaGuardados(ctx, t, e, sesion.ID); estado != "revelada" || ronda != 1 {
		t.Errorf("la sesion quedo %q en la ronda %d, se esperaba revelada en la 1", estado, ronda)
	}
	if votos := votosGuardados(ctx, t, e, sesion.ID); !reflect.DeepEqual(votos, map[int]map[string]int{1: {ana: 5, beto: 8}}) {
		t.Errorf("votos guardados = %v: revelar no tiene que cambiarlos", votos)
	}
}

// US-016 / CA-016-5, CA-016-6, CA-016-7, RN-016-8, CL-016-14: el repositorio vuelve a verificar
// con la sesion bloqueada. Una sesion ya revelada (por ejemplo, entre la lectura del caso de
// uso y la transaccion), sin votos en la ronda o con un integrante que no participa se
// rechaza, y la sesion queda como estaba.
func TestGuardarRevelado_Reglas(t *testing.T) {
	ctx := contexto(t)
	casos := []struct {
		nombre string
		armar  func(t *testing.T, e escenarioPoker, sesionID string) (quien string)
		quiero error
	}{
		{"nadie voto", func(_ *testing.T, e escenarioPoker, _ string) string { return e.integrantes[0] }, planningpoker.ErrSinVotos},
		{"quien revela no participa", func(t *testing.T, e escenarioPoker, id string) string {
			votarEn(ctx, t, e, id, map[string]int{e.integrantes[0]: 5})
			return e.integrante(ctx, t, e.proyecto, "dani")
		}, planningpoker.ErrNoEsParticipante},
		{"la sesion ya esta revelada", func(t *testing.T, e escenarioPoker, id string) string {
			votarEn(ctx, t, e, id, map[string]int{e.integrantes[0]: 5})
			if _, err := e.pool.Exec(ctx, `UPDATE sesiones_planning_poker SET estado = 'revelada' WHERE id = $1`, id); err != nil {
				t.Fatalf("revelar la sesion: %v", err)
			}
			return e.integrantes[0]
		}, planningpoker.ErrVotacionCerrada},
		{"los votos son de una ronda anterior", func(t *testing.T, e escenarioPoker, id string) string {
			votarEn(ctx, t, e, id, map[string]int{e.integrantes[0]: 5})
			if _, err := e.pool.Exec(ctx, `UPDATE sesiones_planning_poker SET ronda = 2 WHERE id = $1`, id); err != nil {
				t.Fatalf("pasar a la ronda 2: %v", err)
			}
			return e.integrantes[0]
		}, planningpoker.ErrSinVotos},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e, sesion := conSesionEnPostgres(ctx, t)
			quien := c.armar(t, e, sesion.ID)
			estadoAntes, rondaAntes := estadoYRondaGuardados(ctx, t, e, sesion.ID)

			r, err := e.repo.GuardarRevelado(ctx, sesion.ID, quien)

			if !errors.Is(err, c.quiero) {
				t.Fatalf("error = %v, se esperaba %v", err, c.quiero)
			}
			if !reflect.DeepEqual(r, planningpoker.Revelacion{}) {
				t.Errorf("con error la revelacion tiene que estar vacia: %+v", r)
			}
			if estado, ronda := estadoYRondaGuardados(ctx, t, e, sesion.ID); estado != estadoAntes || ronda != rondaAntes {
				t.Errorf("la sesion paso de %q/%d a %q/%d: un rechazo no cambia nada", estadoAntes, rondaAntes, estado, ronda)
			}
		})
	}
}

// US-016 / RN-016-8: una sesion que no existe, o con un ID mal formado, se rechaza con
// ErrSesionInexistente.
func TestGuardarRevelado_SesionInexistente(t *testing.T) {
	ctx := contexto(t)
	e, _ := conSesionEnPostgres(ctx, t)

	for _, id := range []string{idQueNoEsDeNingunItem, "no-es-un-uuid"} {
		if _, err := e.repo.GuardarRevelado(ctx, id, e.integrantes[0]); !errors.Is(err, app.ErrSesionInexistente) {
			t.Errorf("GuardarRevelado(%q) = %v, se esperaba ErrSesionInexistente", id, err)
		}
	}
}

// US-016 / CA-016-7, CL-016-11: despues de revelar ya no se vota, y los votos revelados no
// cambian.
func TestGuardarRevelado_DespuesNoSeVota(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	ana, beto := e.integrantes[0], e.integrantes[1]
	votarEn(ctx, t, e, sesion.ID, map[string]int{ana: 5})
	if _, err := e.repo.GuardarRevelado(ctx, sesion.ID, ana); err != nil {
		t.Fatalf("GuardarRevelado: %v", err)
	}

	_, err := e.repo.RegistrarVoto(ctx, sesion.ID, beto, 8)

	if !errors.Is(err, planningpoker.ErrVotacionCerrada) {
		t.Fatalf("votar con la sesion revelada = %v, se esperaba ErrVotacionCerrada", err)
	}
	if votos := votosGuardados(ctx, t, e, sesion.ID); !reflect.DeepEqual(votos, map[int]map[string]int{1: {ana: 5}}) {
		t.Errorf("votos guardados = %v, se esperaba solo el de ana", votos)
	}
}

// US-016 / RN-016-9, CL-016-9: dos revelados simultaneos esperan el candado de la sesion. El
// primero revela; el segundo, cuando lo obtiene, ve la sesion revelada y se rechaza con
// ErrVotacionCerrada, nunca con un error de la base. El segundo arranca recien cuando el
// primero ya espera, asi el orden es fijo.
func TestGuardarRevelado_DosRevelados(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	ana, beto := e.integrantes[0], e.integrantes[1]
	votarEn(ctx, t, e, sesion.ID, map[string]int{ana: 5, beto: 8})
	candado := bloquearSesion(ctx, t, e, sesion.ID)

	errores := make([]chan error, 2)
	for i, quien := range []string{ana, beto} {
		errores[i] = make(chan error, 1)
		go func() {
			_, err := e.repo.GuardarRevelado(ctx, sesion.ID, quien)
			errores[i] <- err
		}()
		if bloqueadas := esperarBloqueadas(ctx, t, e.pool, i+1); bloqueadas < i+1 {
			t.Fatalf("solo %d de %d revelados esperaron el candado de la sesion", bloqueadas, i+1)
		}
	}
	if err := candado.Rollback(ctx); err != nil {
		t.Fatalf("soltar el candado: %v", err)
	}

	if err := <-errores[0]; err != nil {
		t.Errorf("primer revelado: %v", err)
	}
	if err := <-errores[1]; !errors.Is(err, planningpoker.ErrVotacionCerrada) {
		t.Errorf("segundo revelado: error = %v, se esperaba ErrVotacionCerrada", err)
	}
	if estado, _ := estadoYRondaGuardados(ctx, t, e, sesion.ID); estado != "revelada" {
		t.Errorf("la sesion quedo %q, se esperaba revelada", estado)
	}
}

// US-016 / RN-016-9, CL-016-10: un voto que confirma antes del revelado entra en la
// revelacion. El test hace de quien vota: toma el candado, espera a que el revelado quede
// bloqueado, guarda el voto de beto y confirma. Como el revelado lee los votos despues del
// candado, ve el voto de beto.
func TestGuardarRevelado_ConUnVotoQueLlegaAntes(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	ana, beto := e.integrantes[0], e.integrantes[1]
	votarEn(ctx, t, e, sesion.ID, map[string]int{ana: 5})
	votando := bloquearSesion(ctx, t, e, sesion.ID)

	revelado := make(chan planningpoker.Revelacion, 1)
	fallo := make(chan error, 1)
	go func() {
		r, err := e.repo.GuardarRevelado(ctx, sesion.ID, ana)
		revelado <- r
		fallo <- err
	}()
	if bloqueadas := esperarBloqueadas(ctx, t, e.pool, 1); bloqueadas < 1 {
		t.Fatal("el revelado no espero el candado de la sesion")
	}
	if _, err := votando.Exec(ctx, `INSERT INTO votos_planning_poker (sesion_id, ronda, integrante_id, valor) VALUES ($1, 1, $2, 8)`, sesion.ID, beto); err != nil {
		t.Fatalf("guardar el voto de beto: %v", err)
	}
	if err := votando.Commit(ctx); err != nil {
		t.Fatalf("confirmar el voto: %v", err)
	}

	r := <-revelado
	if err := <-fallo; err != nil {
		t.Fatalf("GuardarRevelado: %v", err)
	}
	quiero := []planningpoker.Voto{{ParticipanteID: ana, Valor: 5}, {ParticipanteID: beto, Valor: 8}}
	if !reflect.DeepEqual(r.Votos, quiero) || !r.Divergencia {
		t.Errorf("revelacion = %+v, se esperaba con el voto de beto y divergencia", r)
	}
}
