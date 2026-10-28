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

// US-017: el repositorio cumple el puerto del caso de uso y la lectura de rondas anteriores.
var (
	_ app.RepositorioNuevaRonda    = (*postgres.RepositorioPlanningPoker)(nil)
	_ app.LectorDeRondasAnteriores = (*postgres.RepositorioPlanningPoker)(nil)
)

// idQueNoParticipa es un integrante que no esta en ninguna sesion.
const idQueNoParticipa = "00000000-0000-0000-0000-00000000dead"

func contarVotos(ctx context.Context, t *testing.T, e escenarioPoker, sesionID string, ronda int) int {
	t.Helper()
	var n int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM votos_planning_poker WHERE sesion_id = $1 AND ronda = $2`, sesionID, ronda).Scan(&n); err != nil {
		t.Fatalf("contar los votos: %v", err)
	}
	return n
}

// US-017 / CA-017-1, CA-017-2, RN-017-5, RN-017-6, RN-017-8, CL-017-5: la sesion queda votando
// en la ronda 2, sin votos; los de la ronda 1 se conservan y se leen como ronda anterior, y un
// voto nuevo va a la ronda 2.
func TestGuardarNuevaRonda(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionReveladaEnPostgres(ctx, t)
	ana, beto := e.integrantes[0], e.integrantes[1]

	nueva, err := e.repo.GuardarNuevaRonda(ctx, sesion.ID, beto)
	if err != nil {
		t.Fatalf("GuardarNuevaRonda: %v", err)
	}
	if nueva.Estado != planningpoker.EstadoVotando || nueva.Ronda != 2 {
		t.Fatalf("la sesion quedo %q en la ronda %d, se esperaba votando en la 2", nueva.Estado, nueva.Ronda)
	}
	if guardada, _, err := e.repo.ObtenerSesion(ctx, sesion.ID); err != nil || !reflect.DeepEqual(guardada, nueva) {
		t.Errorf("la sesion guardada %+v (%v) no es la devuelta %+v", guardada, err, nueva)
	}
	if v, err := e.repo.VotacionDeLaRonda(ctx, sesion.ID); err != nil || len(v.Votaron()) != 0 {
		t.Errorf("en la ronda nueva votaron %q (%v), se esperaba nadie", v.Votaron(), err)
	}

	votarEn(ctx, t, e, sesion.ID, map[string]int{ana: 13})

	rondas, err := e.repo.RondasAnteriores(ctx, sesion.ID)
	if err != nil || len(rondas) != 1 || rondas[0].Numero != 1 {
		t.Fatalf("RondasAnteriores = %+v (%v), se esperaba solo la ronda 1", rondas, err)
	}
	if valores, err := rondas[0].Votacion.Valores(); err != nil || !reflect.DeepEqual(valores, map[string]int{ana: 3, beto: 8}) {
		t.Errorf("valores de la ronda 1 = %v (%v), se esperaban los de ana y beto sin cambios", valores, err)
	}
	if n := contarVotos(ctx, t, e, sesion.ID, 1); n != 2 {
		t.Errorf("la ronda 1 tiene %d votos, se esperaban 2", n)
	}
	actual, err := e.repo.VotacionDeLaRonda(ctx, sesion.ID)
	if err != nil || len(actual.Votaron()) != 1 {
		t.Fatalf("en la ronda 2 votaron %q (%v), se esperaba solo ana", actual.Votaron(), err)
	}
	if _, err := actual.Valores(); !errors.Is(err, planningpoker.ErrVotosOcultos) {
		t.Errorf("los votos de la ronda 2 se leen mientras se vota: %v", err)
	}
}

// US-017 / CA-017-3, CA-017-4, RN-017-2 a RN-017-4, RN-017-8: con la sesion bloqueada se vuelve a
// verificar con el dominio. Votando, sin divergencia o con quien no participa se rechaza y no
// cambia nada.
func TestGuardarNuevaRonda_Reglas(t *testing.T) {
	casos := []struct {
		nombre string
		armar  func(context.Context, *testing.T) (escenarioPoker, planningpoker.SesionPlanningPoker)
		quien  func(escenarioPoker) string
		quiero error
		estado string
	}{
		{"sesion votando", func(ctx context.Context, t *testing.T) (escenarioPoker, planningpoker.SesionPlanningPoker) {
			e, s := conSesionEnPostgres(ctx, t)
			votarEn(ctx, t, e, s.ID, map[string]int{e.integrantes[0]: 3, e.integrantes[1]: 8})
			return e, s
		}, func(e escenarioPoker) string { return e.integrantes[0] }, planningpoker.ErrNuevaRondaSinRevelar, "votando"},
		{"sin divergencia", func(ctx context.Context, t *testing.T) (escenarioPoker, planningpoker.SesionPlanningPoker) {
			e, s := conSesionEnPostgres(ctx, t)
			votarEn(ctx, t, e, s.ID, map[string]int{e.integrantes[0]: 5, e.integrantes[1]: 5})
			if _, err := e.repo.GuardarRevelado(ctx, s.ID, e.integrantes[0]); err != nil {
				t.Fatalf("GuardarRevelado: %v", err)
			}
			return e, s
		}, func(e escenarioPoker) string { return e.integrantes[0] }, planningpoker.ErrSinDivergencia, "revelada"},
		{"no participa", conSesionReveladaEnPostgres, func(escenarioPoker) string { return idQueNoParticipa },
			planningpoker.ErrNoEsParticipante, "revelada"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			ctx := contexto(t)
			e, sesion := c.armar(ctx, t)

			_, err := e.repo.GuardarNuevaRonda(ctx, sesion.ID, c.quien(e))

			if !errors.Is(err, c.quiero) {
				t.Fatalf("GuardarNuevaRonda = %v, se esperaba %v", err, c.quiero)
			}
			if s, _, _ := e.repo.ObtenerSesion(ctx, sesion.ID); string(s.Estado) != c.estado || s.Ronda != 1 {
				t.Errorf("la sesion quedo %q en la ronda %d, se esperaba %s en la 1", s.Estado, s.Ronda, c.estado)
			}
		})
	}
}

// RN-017-8: una sesion que no esta, o un ID mal formado, responde ErrSesionInexistente.
func TestGuardarNuevaRonda_SesionInexistente(t *testing.T) {
	ctx := contexto(t)
	e, _ := conSesionReveladaEnPostgres(ctx, t)
	for _, id := range []string{idQueNoParticipa, "no-es-un-uuid"} {
		if _, err := e.repo.GuardarNuevaRonda(ctx, id, e.integrantes[0]); !errors.Is(err, app.ErrSesionInexistente) {
			t.Errorf("GuardarNuevaRonda(%q) = %v, se esperaba ErrSesionInexistente", id, err)
		}
	}
}

// RN-017-6: una sesion en la ronda 1, o que no esta, no tiene rondas anteriores.
func TestRondasAnteriores_Ninguna(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionReveladaEnPostgres(ctx, t)
	for _, id := range []string{sesion.ID, idQueNoParticipa, "no-es-un-uuid"} {
		if rondas, err := e.repo.RondasAnteriores(ctx, id); err != nil || len(rondas) != 0 {
			t.Errorf("RondasAnteriores(%q) = %+v (%v), se esperaba ninguna", id, rondas, err)
		}
	}
}

// US-017 / CA-017-5, RN-017-9, CL-017-10: dos pedidos simultaneos esperan el candado de la sesion.
// El primero abre la ronda 2 y el segundo, que ve la sesion votando, se rechaza con
// ErrNuevaRondaSinRevelar: se abre una sola. El segundo arranca recien cuando el primero ya
// espera, asi el orden es fijo.
func TestGuardarNuevaRonda_DosPedidos(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionReveladaEnPostgres(ctx, t)
	candado := bloquearSesion(ctx, t, e, sesion.ID)

	errores := make([]chan error, 2)
	for i := range errores {
		errores[i] = make(chan error, 1)
		go func() {
			_, err := e.repo.GuardarNuevaRonda(ctx, sesion.ID, e.integrantes[i])
			errores[i] <- err
		}()
		if bloqueadas := esperarBloqueadas(ctx, t, e.pool, i+1); bloqueadas < i+1 {
			t.Fatalf("solo %d de %d pedidos esperaron el candado de la sesion", bloqueadas, i+1)
		}
	}
	if err := candado.Rollback(ctx); err != nil {
		t.Fatalf("soltar el candado: %v", err)
	}

	if err := <-errores[0]; err != nil {
		t.Errorf("primer pedido: %v", err)
	}
	if err := <-errores[1]; !errors.Is(err, planningpoker.ErrNuevaRondaSinRevelar) {
		t.Errorf("segundo pedido: error = %v, se esperaba ErrNuevaRondaSinRevelar", err)
	}
	if s, _, _ := e.repo.ObtenerSesion(ctx, sesion.ID); s.Ronda != 2 || s.Estado != planningpoker.EstadoVotando {
		t.Errorf("la sesion quedo %q en la ronda %d, se esperaba votando en la 2", s.Estado, s.Ronda)
	}
}

// US-017 / RN-017-9, CL-017-11: una nueva ronda y un registro de la estimacion acordada a la vez.
// El que llega primero al candado gana y el otro ve su estado y se rechaza.
func TestGuardarNuevaRonda_ConAcuerdo(t *testing.T) {
	ronda := func(ctx context.Context, e escenarioPoker, sesionID string) error {
		_, err := e.repo.GuardarNuevaRonda(ctx, sesionID, e.integrantes[0])
		return err
	}
	acuerdo := func(ctx context.Context, e escenarioPoker, sesionID string) error {
		_, err := e.repo.GuardarAcuerdo(ctx, sesionID, e.integrantes[0], 5)
		return err
	}
	casos := []struct {
		nombre           string
		primero, segundo func(context.Context, escenarioPoker, string) error
		errSegundo       error
		estado           planningpoker.Estado
	}{
		{"gana la ronda", ronda, acuerdo, planningpoker.ErrSesionNoRevelada, planningpoker.EstadoVotando},
		{"gana el acuerdo", acuerdo, ronda, planningpoker.ErrNuevaRondaSinRevelar, planningpoker.EstadoFinalizada},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			ctx := contexto(t)
			e, sesion := conSesionReveladaEnPostgres(ctx, t)
			candado := bloquearSesion(ctx, t, e, sesion.ID)

			primero, segundo := make(chan error, 1), make(chan error, 1)
			go func() { primero <- c.primero(ctx, e, sesion.ID) }()
			if bloqueadas := esperarBloqueadas(ctx, t, e.pool, 1); bloqueadas < 1 {
				t.Fatal("el primero no espero el candado de la sesion")
			}
			go func() { segundo <- c.segundo(ctx, e, sesion.ID) }()
			if bloqueadas := esperarBloqueadas(ctx, t, e.pool, 2); bloqueadas < 2 {
				t.Fatal("el segundo no espero el candado de la sesion")
			}
			if err := candado.Rollback(ctx); err != nil {
				t.Fatalf("soltar el candado: %v", err)
			}

			if err := <-primero; err != nil {
				t.Errorf("primero: %v", err)
			}
			if err := <-segundo; !errors.Is(err, c.errSegundo) {
				t.Errorf("segundo: error = %v, se esperaba %v", err, c.errSegundo)
			}
			if s, _, _ := e.repo.ObtenerSesion(ctx, sesion.ID); s.Estado != c.estado {
				t.Errorf("la sesion quedo %q, se esperaba %q", s.Estado, c.estado)
			}
		})
	}
}
