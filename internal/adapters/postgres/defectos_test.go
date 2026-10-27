package postgres_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// escenarioDefectos son dos proyectos con fechas que alcanzan para tres sprints: "Demo", con
// una historia y tres sprints (los dos primeros cerrados), y "Otro", con su historia y su
// sprint.
type escenarioDefectos struct {
	pool       *pgxpool.Pool
	repo       *postgres.RepositorioDefectos
	proyecto   string
	otro       string
	item       string
	itemOtro   string
	sprints    [3]string // Sprint 1, 2 y 3 del proyecto
	sprintOtro string
}

func nuevoEscenarioDefectos(ctx context.Context, t *testing.T) escenarioDefectos {
	t.Helper()
	pool := dbprueba.BaseMigrada(ctx, t)
	e := escenarioDefectos{pool: pool, repo: postgres.NuevoRepositorioDefectos(pool)}
	e.proyecto = proyectoConFechas(ctx, t, pool, "Demo", diaSprint(10, 6), diaSprint(12, 31))
	e.otro = proyectoConFechas(ctx, t, pool, "Otro", diaSprint(10, 6), diaSprint(12, 31))

	backlog := postgres.NuevoRepositorioBacklog(pool)
	item, err := backlog.RegistrarItem(ctx, e.proyecto, itemConTitulo(t, "Crear proyecto"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	itemOtro, err := backlog.RegistrarItem(ctx, e.otro, itemConTitulo(t, "Listar proyectos"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	e.item, e.itemOtro = item.ID, itemOtro.ID

	sprints := postgres.NuevoRepositorioSprints(pool)
	fechas := [3][2]time.Time{
		{diaSprint(10, 6), diaSprint(10, 19)},
		{diaSprint(10, 20), diaSprint(11, 2)},
		{diaSprint(11, 3), diaSprint(11, 16)},
	}
	for i, f := range fechas {
		s, err := sprints.RegistrarSprint(ctx, sprintNuevo(t, e.proyecto, f[0], f[1]))
		if err != nil {
			t.Fatalf("RegistrarSprint %d: %v", i+1, err)
		}
		e.sprints[i] = s.ID
		if i < 2 {
			cerrarSprint(ctx, t, pool, s.ID)
		}
	}
	sOtro, err := sprints.RegistrarSprint(ctx, sprintNuevo(t, e.otro, diaSprint(10, 6), diaSprint(10, 19)))
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}
	e.sprintOtro = sOtro.ID
	return e
}

// defecto arma un defecto valido con el dominio, como lo haria el caso de uso: de "Demo", sobre
// su historia y detectado en el Sprint indicado (1 a 3).
func (e escenarioDefectos) defecto(t *testing.T, descripcion string, severidad defecto.Severidad, sprintDeteccion int) defecto.Defecto {
	t.Helper()
	d, err := defecto.Nuevo(e.proyecto, e.item, e.sprints[sprintDeteccion-1], descripcion, severidad)
	if err != nil {
		t.Fatalf("defecto.Nuevo: %v", err)
	}
	return d
}

func (e escenarioDefectos) cantidad(ctx context.Context, t *testing.T) int {
	t.Helper()
	var n int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM defectos`).Scan(&n); err != nil {
		t.Fatalf("contar los defectos: %v", err)
	}
	return n
}

// filaDeDefecto es lo que quedo guardado, leido directamente de la tabla.
type filaDeDefecto struct {
	proyectoID, itemID, sprintDeteccionID string
	sprintResolucionID                    *string
	descripcion, severidad, estado        string
	creadoEn                              time.Time
}

func leerDefecto(ctx context.Context, t *testing.T, pool *pgxpool.Pool, id string) filaDeDefecto {
	t.Helper()
	var f filaDeDefecto
	const consulta = `
SELECT proyecto_id::text, item_id::text, sprint_deteccion_id::text, sprint_resolucion_id::text,
       descripcion, severidad, estado, creado_en
FROM defectos WHERE id = $1`
	err := pool.QueryRow(ctx, consulta, id).Scan(&f.proyectoID, &f.itemID, &f.sprintDeteccionID, &f.sprintResolucionID,
		&f.descripcion, &f.severidad, &f.estado, &f.creadoEn)
	if err != nil {
		t.Fatalf("leer el defecto %s: %v", id, err)
	}
	return f
}

// US-022 / CA-022-1 / RN-022-3 / RN-022-10 / CL-022-14: el defecto se guarda con un ID que genera
// la base y con los datos recibidos, abierto y con el sprint de resolucion en NULL; la fecha de
// carga la pone la base.
func TestRegistrarDefecto_Guarda(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	guardado, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla al guardar", defecto.SeveridadAlta, 1))
	if err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}
	if guardado.ID == "" {
		t.Fatal("el defecto volvio sin ID: lo genera la base con RETURNING")
	}

	f := leerDefecto(ctx, t, e.pool, guardado.ID)
	if f.proyectoID != e.proyecto || f.itemID != e.item || f.sprintDeteccionID != e.sprints[0] {
		t.Errorf("IDs guardados = %s, %s, %s; quiero los del escenario", f.proyectoID, f.itemID, f.sprintDeteccionID)
	}
	if f.descripcion != "Falla al guardar" || f.severidad != "alta" || f.estado != "abierto" {
		t.Errorf("descripcion, severidad y estado = %q, %q, %q; quiero Falla al guardar, alta y abierto", f.descripcion, f.severidad, f.estado)
	}
	if f.sprintResolucionID != nil {
		t.Errorf("sprint_resolucion_id = %q, quiero NULL", *f.sprintResolucionID)
	}
	if f.creadoEn.IsZero() || time.Since(f.creadoEn) > time.Minute {
		t.Errorf("creado_en = %v, quiero la hora de la carga (DEFAULT now())", f.creadoEn)
	}
	if guardado.Estado != defecto.EstadoAbierto || guardado.SprintResolucionID != "" {
		t.Errorf("defecto devuelto = %+v, quiero abierto y sin sprint de resolucion", guardado)
	}
}

// US-022 / CA-022-2 / CL-022-14: las cuatro severidades se guardan y nacen abiertas.
func TestRegistrarDefecto_CuatroSeveridades(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	for _, s := range []defecto.Severidad{defecto.SeveridadCritica, defecto.SeveridadAlta, defecto.SeveridadMedia, defecto.SeveridadBaja} {
		guardado, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla", s, 1))
		if err != nil {
			t.Fatalf("RegistrarDefecto(%s): %v", s, err)
		}
		if f := leerDefecto(ctx, t, e.pool, guardado.ID); f.severidad != string(s) || f.estado != "abierto" || f.sprintResolucionID != nil {
			t.Errorf("severidad %s: fila = %+v, quiero la severidad, abierto y sin resolver", s, f)
		}
	}
}

// US-022 / CA-022-3 / CL-022-12 / CL-022-16 / RN-022-6: se guarda el ID del sprint de deteccion,
// que puede estar cerrado o abierto, y su numero se obtiene de la tabla sprints (lo que necesita
// US-029).
func TestRegistrarDefecto_GuardaElIDDelSprintDeDeteccion(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	for numero := 1; numero <= 3; numero++ {
		guardado, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla", defecto.SeveridadMedia, numero))
		if err != nil {
			t.Fatalf("RegistrarDefecto en el Sprint %d: %v", numero, err)
		}

		var obtenido int
		const consulta = `SELECT s.numero FROM defectos d JOIN sprints s ON s.id = d.sprint_deteccion_id WHERE d.id = $1`
		if err := e.pool.QueryRow(ctx, consulta, guardado.ID).Scan(&obtenido); err != nil || obtenido != numero {
			t.Errorf("numero del sprint de deteccion = %d (error %v), quiero %d", obtenido, err, numero)
		}
	}
}

// US-022 / RN-022-8 / CL-022-13 / CA-022-5: dos defectos iguales son dos filas con IDs distintos.
func TestRegistrarDefecto_Repetidos(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	primero, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla", defecto.SeveridadAlta, 1))
	if err != nil {
		t.Fatalf("primer defecto: %v", err)
	}
	segundo, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla", defecto.SeveridadAlta, 1))
	if err != nil {
		t.Fatalf("segundo defecto: %v", err)
	}

	if primero.ID == segundo.ID {
		t.Errorf("los dos defectos tienen el mismo ID %s", primero.ID)
	}
	if n := e.cantidad(ctx, t); n != 2 {
		t.Errorf("la tabla tiene %d filas, quiero 2", n)
	}
}

// US-022 / RN-022-10: dos cargas simultaneas son dos defectos distintos, y ninguna falla ni espera
// a la otra.
func TestRegistrarDefecto_Simultaneos(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)
	const cargas = 6

	var wg sync.WaitGroup
	ids := make([]string, cargas)
	errs := make([]error, cargas)
	for i := 0; i < cargas; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			guardado, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla", defecto.SeveridadBaja, 1))
			ids[i], errs[i] = guardado.ID, err
		}()
	}
	wg.Wait()

	vistos := map[string]bool{}
	for i := 0; i < cargas; i++ {
		if errs[i] != nil {
			t.Fatalf("carga %d: %v", i, errs[i])
		}
		vistos[ids[i]] = true
	}
	if len(vistos) != cargas {
		t.Errorf("hay %d IDs distintos, quiero %d", len(vistos), cargas)
	}
}

// US-022 / RN-022-5 / CL-022-8 / CL-022-9: la historia y el sprint existen dentro de su proyecto;
// uno de otro proyecto, un ID inexistente, mal formado o vacio no cuenta.
func TestExisteItemYSprintEnProyecto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	casos := []struct {
		nombre   string
		proyecto string
		id       string
		item     bool
		esperado bool
	}{
		{"la historia es del proyecto", e.proyecto, e.item, true, true},
		{"la historia es de otro proyecto", e.proyecto, e.itemOtro, true, false},
		{"la historia no existe", e.proyecto, idSinDueno, true, false},
		{"el ID de la historia esta mal formado", e.proyecto, "no-es-un-uuid", true, false},
		{"el ID de la historia esta vacio", e.proyecto, "", true, false},
		{"el ID del proyecto esta mal formado", "no-es-un-uuid", e.item, true, false},
		{"el sprint es del proyecto", e.proyecto, e.sprints[2], false, true},
		{"el sprint es de otro proyecto", e.proyecto, e.sprintOtro, false, false},
		{"el sprint no existe", e.proyecto, idSinDueno, false, false},
		{"el ID del sprint esta mal formado", e.proyecto, "no-es-un-uuid", false, false},
		{"el ID del sprint esta vacio", e.proyecto, "", false, false},
		{"el ID del proyecto esta mal formado (sprint)", "no-es-un-uuid", e.sprints[0], false, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			var existe bool
			var err error
			if c.item {
				existe, err = e.repo.ExisteItemEnProyecto(ctx, c.proyecto, c.id)
			} else {
				existe, err = e.repo.ExisteSprintEnProyecto(ctx, c.proyecto, c.id)
			}
			if err != nil || existe != c.esperado {
				t.Errorf("existe = (%t, %v), quiero (%t, nil)", existe, err, c.esperado)
			}
		})
	}
}

// US-022 / RN-022-5: el proyecto existe o no, y un ID mal formado es inexistente, no un error.
func TestExisteProyectoDelDefecto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	for id, esperado := range map[string]bool{e.proyecto: true, idSinDueno: false, "no-es-un-uuid": false, "": false} {
		if existe, err := e.repo.ExisteProyecto(ctx, id); err != nil || existe != esperado {
			t.Errorf("ExisteProyecto(%q) = (%t, %v), quiero (%t, nil)", id, existe, err, esperado)
		}
	}
}

// US-022 / RN-022-10 / CL-022-10 / CL-022-15 / CA-022-9 a CA-022-11: si el proyecto, la historia o
// el sprint desaparecieron despues de verificarlos, la clave foranea rechaza el INSERT y el
// repositorio devuelve el error de inexistencia que corresponde, sin guardar nada. Un sprint sin
// elegir (ID vacio) es un sprint inexistente.
func TestRegistrarDefecto_LoQueDesaparece(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	casos := []struct {
		nombre   string
		cambiar  func(d *defecto.Defecto)
		esperado error
	}{
		{"el proyecto no existe", func(d *defecto.Defecto) { d.ProyectoID = idSinDueno }, app.ErrProyectoInexistente},
		{"la historia no existe", func(d *defecto.Defecto) { d.ItemID = idSinDueno }, app.ErrItemInexistente},
		{"el sprint no existe", func(d *defecto.Defecto) { d.SprintDeteccionID = idSinDueno }, app.ErrSprintInexistente},
		{"el ID del proyecto esta mal formado", func(d *defecto.Defecto) { d.ProyectoID = "no-es-un-uuid" }, app.ErrProyectoInexistente},
		{"el ID de la historia esta mal formado", func(d *defecto.Defecto) { d.ItemID = "no-es-un-uuid" }, app.ErrItemInexistente},
		{"el ID del sprint esta mal formado", func(d *defecto.Defecto) { d.SprintDeteccionID = "no-es-un-uuid" }, app.ErrSprintInexistente},
		{"el sprint de deteccion no se eligio", func(d *defecto.Defecto) { d.SprintDeteccionID = "" }, app.ErrSprintInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := e.defecto(t, "Falla", defecto.SeveridadAlta, 1)
			c.cambiar(&d)

			guardado, err := e.repo.RegistrarDefecto(ctx, d)

			if !errors.Is(err, c.esperado) {
				t.Fatalf("error = %v, quiero %v", err, c.esperado)
			}
			if guardado.ID != "" {
				t.Errorf("devolvio el defecto %+v, quiero el valor cero", guardado)
			}
			if n := e.cantidad(ctx, t); n != 0 {
				t.Errorf("la tabla tiene %d filas, quiero ninguna", n)
			}
		})
	}
}

// US-022 / RN-022-1 / RN-022-2: la tabla repite las reglas del dominio como ultima defensa: una
// descripcion vacia o de mas de 2000 caracteres y una severidad fuera de la lista no se pueden
// guardar, y el error no se confunde con uno de inexistencia.
func TestRegistrarDefecto_LaTablaRechazaDatosInvalidos(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	casos := []struct {
		nombre  string
		cambiar func(d *defecto.Defecto)
	}{
		{"descripcion vacia", func(d *defecto.Defecto) { d.Descripcion = "" }},
		{"descripcion de solo espacios", func(d *defecto.Defecto) { d.Descripcion = "   " }},
		{"descripcion de 2001 caracteres", func(d *defecto.Defecto) { d.Descripcion = strings.Repeat("ñ", 2001) }},
		{"severidad fuera de la lista", func(d *defecto.Defecto) { d.Severidad = "urgente" }},
		{"severidad con mayuscula", func(d *defecto.Defecto) { d.Severidad = "Critica" }},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := e.defecto(t, "Falla", defecto.SeveridadAlta, 1)
			c.cambiar(&d)

			_, err := e.repo.RegistrarDefecto(ctx, d)

			if err == nil {
				t.Fatal("se guardo, quiero que la tabla lo rechace")
			}
			if errors.Is(err, app.ErrProyectoInexistente) || errors.Is(err, app.ErrItemInexistente) || errors.Is(err, app.ErrSprintInexistente) {
				t.Errorf("el error %q se confundio con una inexistencia", err)
			}
		})
	}
}

// US-022 / RN-022-1: la tabla acepta el limite exacto, 2000 caracteres (4000 bytes de ñ), porque el
// CHECK cuenta caracteres y no bytes.
func TestRegistrarDefecto_LaTablaAceptaElLimite(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	guardado, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, strings.Repeat("ñ", 2000), defecto.SeveridadAlta, 1))
	if err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}
	if f := leerDefecto(ctx, t, e.pool, guardado.ID); len([]rune(f.descripcion)) != 2000 {
		t.Errorf("se guardaron %d caracteres, quiero 2000", len([]rune(f.descripcion)))
	}
}

// US-022 / Decisiones: el estado no tiene CHECK con lista de valores, porque los define US-023. La
// base acepta cualquier texto: lo protege el dominio hasta entonces.
func TestDefectos_EstadoSinCheckDeValores(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	const insertar = `
INSERT INTO defectos (proyecto_id, item_id, sprint_deteccion_id, descripcion, severidad, estado)
VALUES ($1, $2, $3, 'Falla', 'alta', $4)`
	if _, err := e.pool.Exec(ctx, insertar, e.proyecto, e.item, e.sprints[0], "en_correccion"); err != nil {
		t.Errorf("la tabla rechazo un estado distinto de abierto: %v (lo define US-023)", err)
	}
}

// US-022 / RN-022-10 / CL-022-17: los defectos se borran junto con su proyecto (ON DELETE CASCADE)
// sin error, y los de otro proyecto no se tocan.
func TestRegistrarDefecto_SeBorraConElProyecto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)
	if _, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla", defecto.SeveridadAlta, 3)); err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}
	otro, err := defecto.Nuevo(e.otro, e.itemOtro, e.sprintOtro, "Otra falla", defecto.SeveridadBaja)
	if err != nil {
		t.Fatalf("defecto.Nuevo: %v", err)
	}
	if _, err := e.repo.RegistrarDefecto(ctx, otro); err != nil {
		t.Fatalf("RegistrarDefecto del otro proyecto: %v", err)
	}

	if _, err := e.pool.Exec(ctx, `DELETE FROM proyectos WHERE id = $1`, e.proyecto); err != nil {
		t.Fatalf("borrar el proyecto con defectos: %v", err)
	}

	if n := e.cantidad(ctx, t); n != 1 {
		t.Errorf("la tabla tiene %d defectos, quiero 1 (el del otro proyecto)", n)
	}
}

// US-022 / Persistencia: el sprint de resolucion tiene clave foranea a sprints y puede ser NULL. Un
// ID que no existe se rechaza.
func TestDefectos_SprintDeResolucionConClaveForanea(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)
	guardado, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla", defecto.SeveridadAlta, 1))
	if err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}

	if _, err := e.pool.Exec(ctx, `UPDATE defectos SET sprint_resolucion_id = $2 WHERE id = $1`, guardado.ID, e.sprints[1]); err != nil {
		t.Errorf("completar el sprint de resolucion con un sprint real: %v", err)
	}
	if _, err := e.pool.Exec(ctx, `UPDATE defectos SET sprint_resolucion_id = $2 WHERE id = $1`, guardado.ID, idSinDueno); err == nil {
		t.Error("se acepto un sprint de resolucion que no existe, quiero que la clave foranea lo rechace")
	}
}

// US-022 / Persistencia: los tres indices que pide la spec existen.
func TestDefectos_Indices(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	for _, indice := range []string{"idx_defectos_proyecto", "idx_defectos_item", "idx_defectos_sprint_deteccion"} {
		var existe bool
		const consulta = `SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE tablename = 'defectos' AND indexname = $1)`
		if err := e.pool.QueryRow(ctx, consulta, indice).Scan(&existe); err != nil || !existe {
			t.Errorf("indice %s: existe = %t (error %v), quiero true", indice, existe, err)
		}
	}
}

// US-022 / CA-022-4 / CL-022-11 / RN-022-7: se puede registrar un defecto sobre una historia
// pendiente, en progreso o completada, este o no en un sprint, y la historia no cambia.
func TestRegistrarDefecto_SobreHistoriaEnCualquierEstado(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	casos := []struct {
		nombre string
		estado string
	}{
		{"pendiente", "pendiente"},
		{"en progreso", "en_progreso"},
		{"completada", "completado"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			// completado_con_fecha exige la fecha de completado si, y solo si, esta completado.
			const preparar = `UPDATE items_backlog SET estado = $2, completado_en = CASE WHEN $2 = 'completado' THEN now() END WHERE id = $1`
			if _, err := e.pool.Exec(ctx, preparar, e.item, c.estado); err != nil {
				t.Fatalf("dejar la historia %s: %v", c.estado, err)
			}

			if _, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, "Falla", defecto.SeveridadAlta, 1)); err != nil {
				t.Fatalf("RegistrarDefecto: %v", err)
			}

			var estado string
			if err := e.pool.QueryRow(ctx, `SELECT estado FROM items_backlog WHERE id = $1`, e.item).Scan(&estado); err != nil || estado != c.estado {
				t.Errorf("estado de la historia = %q (error %v), quiero %q: registrar un defecto no la cambia", estado, err, c.estado)
			}
		})
	}
}
