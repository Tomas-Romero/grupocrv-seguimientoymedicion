package postgres_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// US-002 / RN-002-1: BuscarProyecto devuelve los datos guardados. Un ID que no
// esta, vacio o mal formado es un proyecto inexistente, no un error de la base.
func TestBuscarProyecto(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioProyectos(pool)

	registrado, err := repo.RegistrarProyecto(ctx, proyectoDePrueba("Demo"))
	if err != nil {
		t.Fatalf("RegistrarProyecto: %v", err)
	}

	encontrado, err := repo.BuscarProyecto(ctx, registrado.ID)
	if err != nil {
		t.Fatalf("BuscarProyecto(%q): %v", registrado.ID, err)
	}
	if encontrado.ID != registrado.ID ||
		encontrado.Nombre != registrado.Nombre ||
		encontrado.Descripcion != registrado.Descripcion {
		t.Errorf("BuscarProyecto = %+v, se esperaba %+v", encontrado, registrado)
	}
	if !encontrado.FechaInicio.Equal(registrado.FechaInicio) || !encontrado.FechaFin.Equal(registrado.FechaFin) {
		t.Errorf("fechas = %v y %v, se esperaban %v y %v",
			encontrado.FechaInicio, encontrado.FechaFin, registrado.FechaInicio, registrado.FechaFin)
	}

	inexistentes := map[string]string{
		"UUID que no esta en la base": "00000000-0000-0000-0000-000000000000",
		"ID mal formado":              "no-es-un-uuid",
		"ID vacio":                    "",
	}
	for nombre, id := range inexistentes {
		t.Run(nombre, func(t *testing.T) {
			if _, err := repo.BuscarProyecto(ctx, id); !errors.Is(err, app.ErrProyectoInexistente) {
				t.Errorf("BuscarProyecto(%q): error = %v, se esperaba ErrProyectoInexistente", id, err)
			}
		})
	}
}

// US-002 / CA-002-1, CA-002-2, RN-002-3, RN-002-4, RN-002-7: ActualizarProyecto
// reemplaza los cuatro datos, conserva el ID y creado_en, mueve actualizado_en,
// y no toca a los demas proyectos ni a los integrantes del proyecto.
func TestActualizarProyecto(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioProyectos(pool)

	alfa, err := repo.RegistrarProyecto(ctx, proyectoDePrueba("Alfa"))
	if err != nil {
		t.Fatalf("RegistrarProyecto(Alfa): %v", err)
	}
	beta, err := repo.RegistrarProyecto(ctx, proyectoDePrueba("Beta"))
	if err != nil {
		t.Fatalf("RegistrarProyecto(Beta): %v", err)
	}

	const altaIntegrante = `
INSERT INTO integrantes (id, proyecto_id, nombre, email, rol)
VALUES (gen_random_uuid(), $1, 'Ana', 'ana@example.com', 'product_builder')`
	if _, err := pool.Exec(ctx, altaIntegrante, alfa.ID); err != nil {
		t.Fatalf("alta del integrante: %v", err)
	}

	// Se retrocede la marca un dia para comprobar que la actualizacion la
	// mueve, sin depender de la velocidad del reloj.
	const retroceder = `UPDATE proyectos SET actualizado_en = now() - interval '1 day' WHERE id = $1`
	if _, err := pool.Exec(ctx, retroceder, alfa.ID); err != nil {
		t.Fatalf("retroceder actualizado_en: %v", err)
	}
	const leerMarcas = `SELECT creado_en, actualizado_en FROM proyectos WHERE id = $1`
	var creadoAntes, actualizadoAntes time.Time
	if err := pool.QueryRow(ctx, leerMarcas, alfa.ID).Scan(&creadoAntes, &actualizadoAntes); err != nil {
		t.Fatalf("leer las marcas: %v", err)
	}

	cambiado := alfa
	cambiado.Nombre = "Alfa v2"
	cambiado.Descripcion = ""
	cambiado.FechaInicio = time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	cambiado.FechaFin = time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC)

	if err := repo.ActualizarProyecto(ctx, cambiado); err != nil {
		t.Fatalf("ActualizarProyecto: %v", err)
	}

	guardado := leerProyecto(ctx, t, pool, alfa.ID)
	if guardado.nombre != "Alfa v2" || guardado.descripcion != "" {
		t.Errorf("fila guardada = %+v, se esperaban los datos nuevos", guardado)
	}
	if !guardado.inicio.Equal(cambiado.FechaInicio) || !guardado.fin.Equal(cambiado.FechaFin) {
		t.Errorf("fechas guardadas = %v y %v, se esperaban %v y %v",
			guardado.inicio, guardado.fin, cambiado.FechaInicio, cambiado.FechaFin)
	}

	var creado, actualizado time.Time
	if err := pool.QueryRow(ctx, leerMarcas, alfa.ID).Scan(&creado, &actualizado); err != nil {
		t.Fatalf("leer las marcas: %v", err)
	}
	if !creado.Equal(creadoAntes) {
		t.Errorf("creado_en cambio: %v, se esperaba %v", creado, creadoAntes)
	}
	if !actualizado.After(actualizadoAntes) {
		t.Errorf("actualizado_en = %v, se esperaba posterior a %v", actualizado, actualizadoAntes)
	}

	intacto := leerProyecto(ctx, t, pool, beta.ID)
	if intacto.nombre != "Beta" || intacto.descripcion != "Proyecto de prueba" {
		t.Errorf("Beta cambio al modificar Alfa: %+v", intacto)
	}
	if n := dbprueba.ContarFilas(ctx, t, pool, "integrantes"); n != 1 {
		t.Errorf("integrantes = %d, se esperaba 1: modificar no toca lo asociado", n)
	}
}

// US-002 / RN-002-1: actualizar un proyecto que no esta devuelve
// ErrProyectoInexistente, tambien con un ID mal formado, y no guarda nada.
func TestActualizarProyecto_Inexistente(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioProyectos(pool)

	for _, id := range []string{"00000000-0000-0000-0000-000000000000", "no-es-un-uuid", ""} {
		t.Run(id, func(t *testing.T) {
			p := proyectoDePrueba("Fantasma")
			p.ID = id

			err := repo.ActualizarProyecto(ctx, p)
			if !errors.Is(err, app.ErrProyectoInexistente) {
				t.Errorf("ActualizarProyecto(%q): error = %v, se esperaba ErrProyectoInexistente", id, err)
			}
			if n := dbprueba.ContarFilas(ctx, t, pool, "proyectos"); n != 0 {
				t.Errorf("proyectos = %d, se esperaba ninguno", n)
			}
		})
	}
}

// US-002 / seccion 5: los CHECK de la tabla son la ultima defensa. Un cambio
// que llegara sin pasar por el dominio se rechaza y el proyecto queda como
// estaba (RN-002-5).
func TestActualizarProyecto_ChecksDeLaTabla(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioProyectos(pool)

	registrado, err := repo.RegistrarProyecto(ctx, proyectoDePrueba("Demo"))
	if err != nil {
		t.Fatalf("RegistrarProyecto: %v", err)
	}

	casos := []struct {
		nombre string
		editar func(*proyecto.Proyecto)
	}{
		{"nombre_no_vacio: nombre con solo espacios", func(p *proyecto.Proyecto) { p.Nombre = "   " }},
		{"fechas_coherentes: fin anterior al inicio", func(p *proyecto.Proyecto) { p.FechaFin = p.FechaInicio.AddDate(0, 0, -1) }},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			p := registrado
			c.editar(&p)

			err := repo.ActualizarProyecto(ctx, p)
			if err == nil {
				t.Fatal("se esperaba que el UPDATE fallara por un CHECK de la tabla")
			}
			if errors.Is(err, app.ErrProyectoInexistente) {
				t.Errorf("un CHECK violado no es un proyecto inexistente: %v", err)
			}
			guardado := leerProyecto(ctx, t, pool, registrado.ID)
			if guardado.nombre != "Demo" || guardado.descripcion != "Proyecto de prueba" {
				t.Errorf("fila guardada = %+v, se esperaba que no cambiara", guardado)
			}
		})
	}
}