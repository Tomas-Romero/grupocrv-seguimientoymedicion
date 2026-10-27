package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
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

// US-023 / CA-023-1 / CA-023-3 / CL-023-7 / CL-023-12 / RN-023-9 / RN-023-11: el UPDATE guarda el
// estado y el sprint de resolucion y nada mas: la descripcion, la severidad, la historia, el sprint
// de deteccion y la fecha de carga quedan como estaban, y los demas defectos no se tocan. Reabrir un
// resuelto deja el sprint en NULL, y se puede volver a resolver en otro sprint.
func TestActualizarEstadoDefecto_Guarda(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)
	guardado, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla al guardar", defecto.SeveridadAlta, 1))
	if err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}
	otro, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Otra falla", defecto.SeveridadBaja, 2))
	if err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}
	antes := leerDefecto(ctx, t, e.pool, guardado.ID)
	antesOtro := leerDefecto(ctx, t, e.pool, otro.ID)

	cambios := []struct {
		nombre         string
		anterior       defecto.Estado
		nuevo          defecto.Estado
		sprint         string
		sprintEsperado *string
	}{
		{"a en_correccion", defecto.EstadoAbierto, defecto.EstadoEnCorreccion, "", nil},
		{"a resuelto en el Sprint 2", defecto.EstadoEnCorreccion, defecto.EstadoResuelto, e.sprints[1], &e.sprints[1]},
		{"reabierto", defecto.EstadoResuelto, defecto.EstadoAbierto, "", nil},
		{"resuelto de nuevo en el Sprint 3", defecto.EstadoAbierto, defecto.EstadoResuelto, e.sprints[2], &e.sprints[2]},
		{"reabierto otra vez", defecto.EstadoResuelto, defecto.EstadoAbierto, "", nil},
		{"descartado", defecto.EstadoAbierto, defecto.EstadoDescartado, "", nil},
		{"reabierto un descartado", defecto.EstadoDescartado, defecto.EstadoAbierto, "", nil},
	}
	for _, c := range cambios {
		t.Run(c.nombre, func(t *testing.T) {
			cambiado := guardado
			cambiado.Estado, cambiado.SprintResolucionID = c.nuevo, c.sprint

			if err := e.repo.ActualizarEstadoDefecto(ctx, cambiado, c.anterior); err != nil {
				t.Fatalf("ActualizarEstadoDefecto: %v", err)
			}

			f := leerDefecto(ctx, t, e.pool, guardado.ID)
			if f.estado != string(c.nuevo) {
				t.Errorf("estado = %q, quiero %q", f.estado, c.nuevo)
			}
			if (f.sprintResolucionID == nil) != (c.sprintEsperado == nil) || (f.sprintResolucionID != nil && *f.sprintResolucionID != *c.sprintEsperado) {
				t.Errorf("sprint de resolucion = %v, quiero %v", f.sprintResolucionID, c.sprintEsperado)
			}
			if !mismaFilaSalvoEstado(f, antes) {
				t.Errorf("el resto de la fila cambio: %+v, quiero %+v", f, antes)
			}
			if fOtro := leerDefecto(ctx, t, e.pool, otro.ID); !mismaFila(fOtro, antesOtro) {
				t.Errorf("otro defecto cambio: %+v, quiero %+v", fOtro, antesOtro)
			}
		})
	}
}

// mismaFilaSalvoEstado dice si las dos filas son iguales sin mirar el estado ni el sprint de
// resolucion. La fecha de carga se compara con Equal y no con ==.
func mismaFilaSalvoEstado(a, b filaDeDefecto) bool {
	if !a.creadoEn.Equal(b.creadoEn) {
		return false
	}
	a.estado, a.sprintResolucionID, a.creadoEn = b.estado, b.sprintResolucionID, b.creadoEn
	return a == b
}

// mismaFila dice si las dos filas son iguales, incluido el sprint de resolucion (por su valor).
func mismaFila(a, b filaDeDefecto) bool {
	if (a.sprintResolucionID == nil) != (b.sprintResolucionID == nil) ||
		(a.sprintResolucionID != nil && *a.sprintResolucionID != *b.sprintResolucionID) {
		return false
	}
	return mismaFilaSalvoEstado(a, b) && a.estado == b.estado
}

// registrado registra un defecto abierto de "Demo", detectado en el Sprint 1, y devuelve el
// defecto guardado.
func (e escenarioDefectos) registrado(ctx context.Context, t *testing.T) defecto.Defecto {
	t.Helper()
	d, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla al guardar", defecto.SeveridadAlta, 1))
	if err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}
	return d
}

// US-023 / CA-023-14 / CL-023-14 / RN-023-11: TEST PRINCIPAL del cambio simultaneo, determinista. El
// UPDATE esta condicionado al estado que leyo el caso de uso: si el defecto ya no esta en ese estado
// (otro cambio se adelanto) no actualiza ninguna fila, el repositorio relee y devuelve
// ErrDefectoModificado sin guardar nada. El defecto queda como lo dejo el primer cambio, sin mezclar:
// no queda descartado con el sprint de resolucion del segundo.
func TestActualizarEstadoDefecto_RechazaUnEstadoAnteriorQueYaNoEsElGuardado(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)
	guardado := e.registrado(ctx, t)

	// El primer cambio gana: lo descarta.
	descartado := guardado
	descartado.Estado = defecto.EstadoDescartado
	if err := e.repo.ActualizarEstadoDefecto(ctx, descartado, defecto.EstadoAbierto); err != nil {
		t.Fatalf("ActualizarEstadoDefecto del primero: %v", err)
	}
	antes := leerDefecto(ctx, t, e.pool, guardado.ID)

	// El segundo leyo "abierto" antes que el primero y lo quiere resolver en el Sprint 2.
	resuelto := guardado
	resuelto.Estado, resuelto.SprintResolucionID = defecto.EstadoResuelto, e.sprints[1]
	err := e.repo.ActualizarEstadoDefecto(ctx, resuelto, defecto.EstadoAbierto)

	if !errors.Is(err, app.ErrDefectoModificado) {
		t.Fatalf("error = %v, quiero ErrDefectoModificado", err)
	}
	if f := leerDefecto(ctx, t, e.pool, guardado.ID); !mismaFila(f, antes) || f.estado != "descartado" || f.sprintResolucionID != nil {
		t.Errorf("fila = %+v, quiero que siga como la dejo el primero: descartado y sin sprint", f)
	}
}

// US-023 / CL-023-15 / CL-023-10 / RN-023-11: el defecto que no existe en el proyecto (se borro su
// proyecto, es de otro, o el ID esta mal formado o vacio) se distingue de uno modificado:
// ErrDefectoInexistente. No se guarda nada.
func TestActualizarEstadoDefecto_DefectoQueNoEsta(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)
	guardado := e.registrado(ctx, t)
	antes := leerDefecto(ctx, t, e.pool, guardado.ID)

	casos := []struct {
		nombre  string
		cambiar func(d *defecto.Defecto)
	}{
		{"el ID no existe", func(d *defecto.Defecto) { d.ID = idSinDueno }},
		{"el defecto es de otro proyecto", func(d *defecto.Defecto) { d.ProyectoID = e.otro }},
		{"el proyecto no existe", func(d *defecto.Defecto) { d.ProyectoID = idSinDueno }},
		{"el ID esta mal formado", func(d *defecto.Defecto) { d.ID = "no-es-un-uuid" }},
		{"el ID esta vacio", func(d *defecto.Defecto) { d.ID = "" }},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := guardado
			d.Estado, d.SprintResolucionID = defecto.EstadoResuelto, e.sprints[1]
			c.cambiar(&d)

			err := e.repo.ActualizarEstadoDefecto(ctx, d, defecto.EstadoAbierto)

			if !errors.Is(err, app.ErrDefectoInexistente) {
				t.Fatalf("error = %v, quiero ErrDefectoInexistente", err)
			}
			if f := leerDefecto(ctx, t, e.pool, guardado.ID); !mismaFila(f, antes) {
				t.Errorf("fila = %+v, quiero que no cambie: %+v", f, antes)
			}
		})
	}
}

// US-023 / CL-023-15 / RN-023-11: si el sprint de resolucion desaparecio entre la lectura y el
// UPDATE, la clave foranea lo rechaza y se devuelve ErrSprintInexistente, sin guardar nada: ni el
// estado. Un ID de sprint mal formado se traduce igual, sin llegar a la base.
func TestActualizarEstadoDefecto_SprintDeResolucionQueNoEsta(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)
	guardado := e.registrado(ctx, t)
	antes := leerDefecto(ctx, t, e.pool, guardado.ID)

	for _, sprint := range []string{idSinDueno, "no-es-un-uuid"} {
		t.Run(sprint, func(t *testing.T) {
			d := guardado
			d.Estado, d.SprintResolucionID = defecto.EstadoResuelto, sprint

			err := e.repo.ActualizarEstadoDefecto(ctx, d, defecto.EstadoAbierto)

			if !errors.Is(err, app.ErrSprintInexistente) {
				t.Fatalf("error = %v, quiero ErrSprintInexistente", err)
			}
			if f := leerDefecto(ctx, t, e.pool, guardado.ID); !mismaFila(f, antes) {
				t.Errorf("fila = %+v, quiero que no cambie: %+v", f, antes)
			}
		})
	}
}

// US-023 / CA-023-14 / CL-023-14 / RN-023-11: COMPLEMENTO del test determinista, con dos goroutines
// contra Postgres real: dos cambios del mismo defecto abierto (uno a resuelto en el Sprint 2 y otro
// a descartado) que leyeron "abierto". No depende de tiempos ni de quien llega primero: sea cual
// sea el orden, exactamente uno guarda y el otro recibe ErrDefectoModificado, y el defecto queda
// como lo dejo el que gano, sin mezclar (no queda descartado con el sprint 2).
func TestActualizarEstadoDefecto_CambiosSimultaneos(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)
	guardado := e.registrado(ctx, t)

	resuelto := guardado
	resuelto.Estado, resuelto.SprintResolucionID = defecto.EstadoResuelto, e.sprints[1]
	descartado := guardado
	descartado.Estado = defecto.EstadoDescartado
	cambios := []defecto.Defecto{resuelto, descartado}

	// Las dos goroutines arrancan juntas, pero el resultado no depende de eso.
	var wg sync.WaitGroup
	salida := make([]error, len(cambios))
	arranque := make(chan struct{})
	for i := range cambios {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-arranque
			salida[i] = e.repo.ActualizarEstadoDefecto(ctx, cambios[i], defecto.EstadoAbierto)
		}()
	}
	close(arranque)
	wg.Wait()

	ganador := -1
	for i, err := range salida {
		switch {
		case err == nil && ganador == -1:
			ganador = i
		case err == nil:
			t.Fatalf("los dos cambios guardaron, quiero que solo uno gane")
		case !errors.Is(err, app.ErrDefectoModificado):
			t.Fatalf("cambio %d: error = %v, quiero ErrDefectoModificado", i, err)
		}
	}
	if ganador == -1 {
		t.Fatalf("ningun cambio guardo: %v", salida)
	}

	f := leerDefecto(ctx, t, e.pool, guardado.ID)
	esperado := cambios[ganador]
	if f.estado != string(esperado.Estado) || (f.sprintResolucionID == nil) != (esperado.SprintResolucionID == "") ||
		(f.sprintResolucionID != nil && *f.sprintResolucionID != esperado.SprintResolucionID) {
		t.Errorf("fila = estado %q, sprint %v; quiero lo que dejo el cambio ganador: %q y %q",
			f.estado, f.sprintResolucionID, esperado.Estado, esperado.SprintResolucionID)
	}
}
