package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// conSesionReveladaEnPostgres devuelve el escenario con una sesion revelada: ana voto 3 y
// beto 8, y ana revelo.
func conSesionReveladaEnPostgres(ctx context.Context, t *testing.T) (escenarioPoker, planningpoker.SesionPlanningPoker) {
	t.Helper()
	e, sesion := conSesionEnPostgres(ctx, t)
	votarEn(ctx, t, e, sesion.ID, map[string]int{e.integrantes[0]: 3, e.integrantes[1]: 8})
	if _, err := e.repo.GuardarRevelado(ctx, sesion.ID, e.integrantes[0]); err != nil {
		t.Fatalf("GuardarRevelado: %v", err)
	}
	return e, sesion
}

// acuerdoGuardado es lo que quedo en la base: el estado y el valor acordado de la sesion, y
// los Story Points de la historia.
type acuerdoGuardado struct {
	estado        string
	valorAcordado *int32
	storyPoints   *int32
}

func leerAcuerdo(ctx context.Context, t *testing.T, e escenarioPoker, sesionID string) acuerdoGuardado {
	t.Helper()
	var a acuerdoGuardado
	const consulta = `
SELECT s.estado, s.valor_acordado, i.story_points
FROM sesiones_planning_poker s JOIN items_backlog i ON i.id = s.item_id WHERE s.id = $1`
	if err := e.pool.QueryRow(ctx, consulta, sesionID).Scan(&a.estado, &a.valorAcordado, &a.storyPoints); err != nil {
		t.Fatalf("leer el acuerdo: %v", err)
	}
	return a
}

func valorDe(p *int32) int {
	if p == nil {
		return -1
	}
	return int(*p)
}

// US-018 / CA-018-1, CA-018-2, RN-018-8: en una transaccion, la sesion queda finalizada
// con el valor acordado (aunque nadie lo haya votado) y la historia, estimada con el. Los
// votos no cambian, y ObtenerSesion lee el valor acordado.
func TestGuardarAcuerdo(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionReveladaEnPostgres(ctx, t)

	finalizada, err := e.repo.GuardarAcuerdo(ctx, sesion.ID, e.integrantes[1], 5)
	if err != nil {
		t.Fatalf("GuardarAcuerdo: %v", err)
	}

	if puntos, ok := finalizada.ValorAcordado.Valor(); finalizada.Estado != planningpoker.EstadoFinalizada || !ok || puntos != 5 {
		t.Errorf("devuelta = (%q, %d, %t), se esperaba (finalizada, 5, true)", finalizada.Estado, puntos, ok)
	}
	if a := leerAcuerdo(ctx, t, e, sesion.ID); a.estado != "finalizada" || valorDe(a.valorAcordado) != 5 || valorDe(a.storyPoints) != 5 {
		t.Errorf("guardado = (%q, acordado %d, story points %d), se esperaba (finalizada, 5, 5)", a.estado, valorDe(a.valorAcordado), valorDe(a.storyPoints))
	}
	if votos := votosGuardados(ctx, t, e, sesion.ID); votos[1][e.integrantes[0]] != 3 || votos[1][e.integrantes[1]] != 8 {
		t.Errorf("los votos cambiaron: %v", votos)
	}
	leida, _, err := e.repo.ObtenerSesion(ctx, sesion.ID)
	if puntos, ok := leida.ValorAcordado.Valor(); err != nil || !ok || puntos != 5 {
		t.Errorf("ObtenerSesion leyo el valor acordado (%d, %t, %v), se esperaba 5", puntos, ok, err)
	}
}

// US-018 / CA-018-3, CL-018-5: una historia ya estimada queda con el valor acordado.
func TestGuardarAcuerdo_HistoriaYaEstimada(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionReveladaEnPostgres(ctx, t)
	tres, _ := backlog.NuevosStoryPoints(3)
	if err := e.repo.GuardarEstimacion(ctx, e.item, tres); err != nil {
		t.Fatalf("GuardarEstimacion: %v", err)
	}

	if _, err := e.repo.GuardarAcuerdo(ctx, sesion.ID, e.integrantes[0], 8); err != nil {
		t.Fatalf("GuardarAcuerdo: %v", err)
	}

	if a := leerAcuerdo(ctx, t, e, sesion.ID); valorDe(a.storyPoints) != 8 {
		t.Errorf("story points = %d, se esperaba el valor acordado (8)", valorDe(a.storyPoints))
	}
}

// US-018 / RN-018-9, CL-018-6, CL-018-7, CL-018-8: el repositorio vuelve a verificar con la
// sesion bloqueada. Una sesion votando o ya finalizada, un integrante que no participa, un
// valor fuera de la escala o una historia completada se rechazan, y no cambia nada.
func TestGuardarAcuerdo_Reglas(t *testing.T) {
	ctx := contexto(t)
	casos := []struct {
		nombre      string
		preparar    func(e escenarioPoker, sesionID string)
		integrante  func(e escenarioPoker) string
		valor       int
		errEsperado error
		estado      string
	}{
		{"sesion votando", func(e escenarioPoker, sesionID string) {
			if _, err := e.pool.Exec(ctx, `UPDATE sesiones_planning_poker SET estado = 'votando' WHERE id = $1`, sesionID); err != nil {
				t.Fatalf("volver a votando: %v", err)
			}
		}, func(e escenarioPoker) string { return e.integrantes[0] }, 5, planningpoker.ErrSesionNoRevelada, "votando"},
		{"no participa", func(escenarioPoker, string) {}, func(e escenarioPoker) string {
			return e.integrante(ctx, t, e.proyecto, "dani")
		}, 5, planningpoker.ErrNoEsParticipante, "revelada"},
		{"fuera de la escala", func(escenarioPoker, string) {}, func(e escenarioPoker) string { return e.integrantes[0] }, 4, backlog.ErrEstimacionFueraDeEscala, "revelada"},
		{"historia completada", func(e escenarioPoker, _ string) {
			if _, err := e.pool.Exec(ctx, `UPDATE items_backlog SET estado = 'completado', completado_en = now() WHERE id = $1`, e.item); err != nil {
				t.Fatalf("completar la historia: %v", err)
			}
		}, func(e escenarioPoker) string { return e.integrantes[0] }, 4, backlog.ErrItemCompletado, "revelada"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e, sesion := conSesionReveladaEnPostgres(ctx, t)
			c.preparar(e, sesion.ID)

			_, err := e.repo.GuardarAcuerdo(ctx, sesion.ID, c.integrante(e), c.valor)

			if !errors.Is(err, c.errEsperado) {
				t.Fatalf("error = %v, se esperaba %v", err, c.errEsperado)
			}
			if a := leerAcuerdo(ctx, t, e, sesion.ID); a.estado != c.estado || a.valorAcordado != nil || a.storyPoints != nil {
				t.Errorf("cambio algo: (%q, acordado %d, story points %d)", a.estado, valorDe(a.valorAcordado), valorDe(a.storyPoints))
			}
		})
	}

	t.Run("sesion finalizada", func(t *testing.T) {
		e, sesion := conSesionReveladaEnPostgres(ctx, t)
		if _, err := e.repo.GuardarAcuerdo(ctx, sesion.ID, e.integrantes[0], 5); err != nil {
			t.Fatalf("primer registro: %v", err)
		}
		if _, err := e.repo.GuardarAcuerdo(ctx, sesion.ID, e.integrantes[0], 8); !errors.Is(err, planningpoker.ErrSesionNoRevelada) {
			t.Errorf("segundo registro: error = %v, se esperaba ErrSesionNoRevelada", err)
		}
		if a := leerAcuerdo(ctx, t, e, sesion.ID); valorDe(a.valorAcordado) != 5 || valorDe(a.storyPoints) != 5 {
			t.Errorf("el segundo registro cambio el acuerdo: %+v", a)
		}
	})
}

// US-018 / CA-018-8: una sesion que no existe, o con un ID mal formado, se rechaza con
// ErrSesionInexistente.
func TestGuardarAcuerdo_SesionInexistente(t *testing.T) {
	ctx := contexto(t)
	e, _ := conSesionReveladaEnPostgres(ctx, t)

	for _, id := range []string{idQueNoEsDeNingunItem, "no-es-un-uuid"} {
		if _, err := e.repo.GuardarAcuerdo(ctx, id, e.integrantes[0], 5); !errors.Is(err, app.ErrSesionInexistente) {
			t.Errorf("GuardarAcuerdo(%q) = %v, se esperaba ErrSesionInexistente", id, err)
		}
	}
}

// US-018 / RN-018-11: el CHECK de la migracion ata el valor acordado al estado: una sesion
// finalizada sin valor, o con valor y sin finalizar, es un dato corrupto.
func TestMigracion_ValorAcordadoConFinalizada(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionEnPostgres(ctx, t)
	for nombre, sentencia := range map[string]string{
		"finalizada sin valor": `UPDATE sesiones_planning_poker SET estado = 'finalizada' WHERE id = $1`,
		"votando con valor":    `UPDATE sesiones_planning_poker SET valor_acordado = 5 WHERE id = $1`,
		"revelada con valor":   `UPDATE sesiones_planning_poker SET estado = 'revelada', valor_acordado = 5 WHERE id = $1`,
	} {
		_, err := e.pool.Exec(ctx, sentencia, sesion.ID)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
			t.Errorf("%s: error = %v, se esperaba la violacion del CHECK (23514)", nombre, err)
		}
	}
	if _, err := e.pool.Exec(ctx, `UPDATE sesiones_planning_poker SET estado = 'finalizada', valor_acordado = 5 WHERE id = $1`, sesion.ID); err != nil {
		t.Errorf("finalizada con valor: %v", err)
	}
}

// US-018 / RN-018-9, CL-018-10: si la historia se completa mientras se registra, el UPDATE
// condicionado de US-013 no la estima y se revierte todo: la sesion sigue revelada. Para que
// pase de verdad en el medio, el test completa la historia en una transaccion propia sin
// confirmar: el registro lee la historia pendiente, queda esperando en el UPDATE, y cuando el
// test confirma, el UPDATE ya no encuentra la fila sin completar.
func TestGuardarAcuerdo_HistoriaQueSeCompletaEnElMedio(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionReveladaEnPostgres(ctx, t)

	marcado, err := e.pool.Begin(ctx)
	if err != nil {
		t.Fatalf("abrir la transaccion del marcado: %v", err)
	}
	defer func() { _ = marcado.Rollback(ctx) }()
	if _, err := marcado.Exec(ctx, `UPDATE items_backlog SET estado = 'completado', completado_en = now() WHERE id = $1`, e.item); err != nil {
		t.Fatalf("completar la historia: %v", err)
	}

	resultado := make(chan error, 1)
	go func() {
		_, err := e.repo.GuardarAcuerdo(ctx, sesion.ID, e.integrantes[0], 5)
		resultado <- err
	}()
	if bloqueadas := esperarBloqueadas(ctx, t, e.pool, 1); bloqueadas < 1 {
		t.Fatal("el registro no espero a la historia: no la estima con el UPDATE condicionado")
	}
	if err := marcado.Commit(ctx); err != nil {
		t.Fatalf("confirmar el marcado: %v", err)
	}

	if err := <-resultado; !errors.Is(err, backlog.ErrItemCompletado) {
		t.Fatalf("error = %v, se esperaba ErrItemCompletado", err)
	}
	if a := leerAcuerdo(ctx, t, e, sesion.ID); a.estado != "revelada" || a.valorAcordado != nil || a.storyPoints != nil {
		t.Errorf("cambio algo: (%q, acordado %d, story points %d), se esperaba la sesion revelada y la historia sin estimar", a.estado, valorDe(a.valorAcordado), valorDe(a.storyPoints))
	}
}

// US-018 / RN-018-10, CL-018-11: dos registros simultaneos de la misma sesion esperan el
// candado de la sesion. El primero finaliza y el segundo, que ve la sesion finalizada, se
// rechaza con ErrSesionNoRevelada: queda el valor del primero. El segundo arranca recien
// cuando el primero ya espera, asi el orden es fijo.
func TestGuardarAcuerdo_DosRegistros(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionReveladaEnPostgres(ctx, t)
	candado := bloquearSesion(ctx, t, e, sesion.ID)

	errores := make([]chan error, 2)
	for i, valor := range []int{5, 8} {
		errores[i] = make(chan error, 1)
		go func() {
			_, err := e.repo.GuardarAcuerdo(ctx, sesion.ID, e.integrantes[0], valor)
			errores[i] <- err
		}()
		if bloqueadas := esperarBloqueadas(ctx, t, e.pool, i+1); bloqueadas < i+1 {
			t.Fatalf("solo %d de %d registros esperaron el candado de la sesion", bloqueadas, i+1)
		}
	}
	if err := candado.Rollback(ctx); err != nil {
		t.Fatalf("soltar el candado: %v", err)
	}

	if err := <-errores[0]; err != nil {
		t.Errorf("primer registro: %v", err)
	}
	if err := <-errores[1]; !errors.Is(err, planningpoker.ErrSesionNoRevelada) {
		t.Errorf("segundo registro: error = %v, se esperaba ErrSesionNoRevelada", err)
	}
	if a := leerAcuerdo(ctx, t, e, sesion.ID); valorDe(a.valorAcordado) != 5 || valorDe(a.storyPoints) != 5 {
		t.Errorf("acuerdo = %+v, se esperaba el del primer registro (5)", a)
	}
}

// US-018 / RN-018-9, CL-018-12: si la transaccion falla despues de estimar la historia y
// antes de finalizar la sesion, no cambia nada. Para forzar la falla, un trigger de la base
// de prueba rechaza el ultimo paso (finalizar la sesion).
func TestGuardarAcuerdo_FallaAMitad(t *testing.T) {
	ctx := contexto(t)
	e, sesion := conSesionReveladaEnPostgres(ctx, t)
	const trigger = `
CREATE FUNCTION rechazar_finalizar() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN RAISE EXCEPTION 'falla forzada por el test'; END $$;
CREATE TRIGGER rechazar_finalizar BEFORE UPDATE ON sesiones_planning_poker FOR EACH ROW EXECUTE FUNCTION rechazar_finalizar();`
	if _, err := e.pool.Exec(ctx, trigger); err != nil {
		t.Fatalf("crear el trigger: %v", err)
	}

	if _, err := e.repo.GuardarAcuerdo(ctx, sesion.ID, e.integrantes[0], 5); err == nil {
		t.Fatal("GuardarAcuerdo termino bien, se esperaba la falla forzada")
	}

	if a := leerAcuerdo(ctx, t, e, sesion.ID); a.estado != "revelada" || a.valorAcordado != nil || a.storyPoints != nil {
		t.Errorf("cambio algo: (%q, acordado %d, story points %d), se esperaba nada", a.estado, valorDe(a.valorAcordado), valorDe(a.storyPoints))
	}
}
