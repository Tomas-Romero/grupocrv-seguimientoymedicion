package postgres_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// codigoCheckViolado es el SQLSTATE de un CHECK que no se cumple.
const codigoCheckViolado = "23514"

// US-023 / CA-023-14 / CL-023-16 / RN-023-12: la tabla repite dos reglas del dominio como ultima
// defensa. Un UPDATE directo que deja un estado fuera de los cuatro, un resuelto sin sprint de
// resolucion o un estado distinto de resuelto con sprint lo rechaza un CHECK (y se distingue
// cual), y las combinaciones validas se aceptan.
func TestDefectos_LosCheckDelEstado(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)
	guardado, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla", defecto.SeveridadAlta, 1))
	if err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}

	const actualizar = `UPDATE defectos SET estado = $2, sprint_resolucion_id = $3 WHERE id = $1`
	rechazados := []struct {
		nombre      string
		estado      string
		sprint      *string
		restriccion string
	}{
		{"un estado fuera de los cuatro", "cerrado", nil, "estado_de_defecto_valido"},
		{"un estado con mayuscula", "Resuelto", &e.sprints[1], "estado_de_defecto_valido"},
		{"resuelto sin sprint de resolucion", "resuelto", nil, "resolucion_segun_estado"},
		{"en_correccion con sprint de resolucion", "en_correccion", &e.sprints[1], "resolucion_segun_estado"},
		{"abierto con sprint de resolucion", "abierto", &e.sprints[1], "resolucion_segun_estado"},
		{"descartado con sprint de resolucion", "descartado", &e.sprints[1], "resolucion_segun_estado"},
	}
	for _, c := range rechazados {
		t.Run("rechaza "+c.nombre, func(t *testing.T) {
			_, err := e.pool.Exec(ctx, actualizar, guardado.ID, c.estado, c.sprint)

			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != codigoCheckViolado || pgErr.ConstraintName != c.restriccion {
				t.Fatalf("error = %v, quiero el CHECK %s", err, c.restriccion)
			}
			if f := leerDefecto(ctx, t, e.pool, guardado.ID); f.estado != "abierto" || f.sprintResolucionID != nil {
				t.Errorf("el defecto quedo en %q con sprint %v, quiero que no cambie", f.estado, f.sprintResolucionID)
			}
		})
	}

	aceptados := []struct {
		nombre string
		estado string
		sprint *string
	}{
		{"resuelto con sprint", "resuelto", &e.sprints[1]},
		{"abierto sin sprint (reapertura)", "abierto", nil},
		{"en_correccion sin sprint", "en_correccion", nil},
		{"descartado sin sprint", "descartado", nil},
	}
	for _, c := range aceptados {
		t.Run("acepta "+c.nombre, func(t *testing.T) {
			if _, err := e.pool.Exec(ctx, actualizar, guardado.ID, c.estado, c.sprint); err != nil {
				t.Errorf("la tabla rechazo una combinacion valida: %v", err)
			}
		})
	}
}
