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

// US-023 / RN-023-7 / RN-023-8 / CL-023-10: el defecto se busca DENTRO del proyecto y se devuelve
// tal como esta guardado, de donde sale su estado actual. Uno de otro proyecto, un ID inexistente,
// mal formado o vacio no cuenta, y no es un error.
func TestRepositorioDefectos_DefectoEnProyecto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)
	guardado, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla al guardar", defecto.SeveridadAlta, 1))
	if err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}

	casos := []struct {
		nombre   string
		proyecto string
		id       string
		existe   bool
	}{
		{"de su proyecto", e.proyecto, guardado.ID, true},
		{"de otro proyecto", e.otro, guardado.ID, false},
		{"un ID que no existe", e.proyecto, idSinDueno, false},
		{"un ID mal formado", e.proyecto, "no-es-un-uuid", false},
		{"un ID vacio", e.proyecto, "", false},
		{"un proyecto mal formado", "no-es-un-uuid", guardado.ID, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, existe, err := e.repo.DefectoEnProyecto(ctx, c.proyecto, c.id)
			if err != nil {
				t.Fatalf("DefectoEnProyecto: %v", err)
			}
			if existe != c.existe {
				t.Fatalf("existe = %v, quiero %v", existe, c.existe)
			}
			if c.existe && d != guardado {
				t.Errorf("defecto = %+v, quiero %+v", d, guardado)
			}
			if !c.existe && d != (defecto.Defecto{}) {
				t.Errorf("defecto = %+v, quiero el valor cero", d)
			}
		})
	}
}

// US-023 / CL-023-17 / RN-023-8: el sprint se busca DENTRO del proyecto y se traduce a su numero
// (sprint.Numero), que es lo que compara el dominio: el ID del Sprint 3 da 3. Uno de otro proyecto,
// un ID inexistente, mal formado o vacio no cuenta.
func TestRepositorioDefectos_NumeroDeSprintEnProyecto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	casos := []struct {
		nombre   string
		proyecto string
		id       string
		numero   int
		existe   bool
	}{
		{"el Sprint 1", e.proyecto, e.sprints[0], 1, true},
		{"el Sprint 3", e.proyecto, e.sprints[2], 3, true},
		{"un sprint de otro proyecto", e.proyecto, e.sprintOtro, 0, false},
		{"un ID que no existe", e.proyecto, idSinDueno, 0, false},
		{"un ID mal formado", e.proyecto, "no-es-un-uuid", 0, false},
		{"un ID vacio", e.proyecto, "", 0, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			numero, existe, err := e.repo.NumeroDeSprintEnProyecto(ctx, c.proyecto, c.id)
			if err != nil {
				t.Fatalf("NumeroDeSprintEnProyecto: %v", err)
			}
			if numero != c.numero || existe != c.existe {
				t.Errorf("numero, existe = %d, %v; quiero %d, %v", numero, existe, c.numero, c.existe)
			}
		})
	}
}
