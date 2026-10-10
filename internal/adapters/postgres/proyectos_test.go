package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

func proyectoDePrueba(nombre string) proyecto.Proyecto {
	return proyecto.Proyecto{
		Nombre:      nombre,
		Descripcion: "Proyecto de prueba",
		FechaInicio: time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(2026, 11, 18, 0, 0, 0, 0, time.UTC),
	}
}

// filaProyecto es lo que quedo guardado de un proyecto, leido de la tabla.
type filaProyecto struct {
	nombre      string
	descripcion string
	inicio, fin time.Time
}

func leerProyecto(ctx context.Context, t *testing.T, pool *pgxpool.Pool, id string) filaProyecto {
	t.Helper()
	var f filaProyecto
	const consulta = `SELECT nombre, descripcion, fecha_inicio, fecha_fin FROM proyectos WHERE id = $1`
	if err := pool.QueryRow(ctx, consulta, id).Scan(&f.nombre, &f.descripcion, &f.inicio, &f.fin); err != nil {
		t.Fatalf("leer el proyecto %s: %v", id, err)
	}
	return f
}

// US-001 / CA-001-1, RN-001-6: el proyecto se guarda con un ID que genera la
// base (DEFAULT de la migracion 00002) y se devuelve con RETURNING.
func TestRegistrarProyecto(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioProyectos(pool)
	enviado := proyectoDePrueba("Demo")

	registrado, err := repo.RegistrarProyecto(ctx, enviado)
	if err != nil {
		t.Fatalf("RegistrarProyecto: %v", err)
	}
	if registrado.ID == "" {
		t.Fatal("ID vacio, se esperaba el que genera la base")
	}
	if registrado.Nombre != enviado.Nombre || registrado.Descripcion != enviado.Descripcion {
		t.Errorf("proyecto devuelto = %+v, se esperaban los datos enviados", registrado)
	}

	guardado := leerProyecto(ctx, t, pool, registrado.ID)
	if guardado.nombre != "Demo" || guardado.descripcion != "Proyecto de prueba" {
		t.Errorf("fila guardada = %+v, se esperaban los datos enviados", guardado)
	}
	if !guardado.inicio.Equal(enviado.FechaInicio) || !guardado.fin.Equal(enviado.FechaFin) {
		t.Errorf("fechas guardadas = %v y %v, se esperaban %v y %v",
			guardado.inicio, guardado.fin, enviado.FechaInicio, enviado.FechaFin)
	}
}

// US-001 / CA-001-2, seccion 5: el nombre no es unico. Dos proyectos con el
// mismo nombre se guardan, cada uno con su ID.
func TestRegistrarProyecto_NombreRepetido(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioProyectos(pool)

	primero, err := repo.RegistrarProyecto(ctx, proyectoDePrueba("Demo"))
	if err != nil {
		t.Fatalf("RegistrarProyecto (primero): %v", err)
	}
	segundo, err := repo.RegistrarProyecto(ctx, proyectoDePrueba("Demo"))
	if err != nil {
		t.Fatalf("RegistrarProyecto (segundo): %v", err)
	}
	if primero.ID == segundo.ID {
		t.Errorf("los dos proyectos tienen el mismo ID %q", primero.ID)
	}
	if n := dbprueba.ContarFilas(ctx, t, pool, "proyectos"); n != 2 {
		t.Errorf("proyectos = %d, se esperaban 2", n)
	}
}

// US-001 / seccion 5: los CHECK de la tabla son la ultima defensa. Un proyecto
// que llegara sin pasar por el dominio se rechaza y no se guarda nada.
func TestRegistrarProyecto_ChecksDeLaTabla(t *testing.T) {
	ctx := contexto(t)
	pool := dbprueba.BaseMigrada(ctx, t)
	repo := postgres.NuevoRepositorioProyectos(pool)

	casos := []struct {
		nombre string
		editar func(*proyecto.Proyecto)
	}{
		{"nombre_no_vacio: nombre con solo espacios", func(p *proyecto.Proyecto) { p.Nombre = "   " }},
		{"fechas_coherentes: fin anterior al inicio", func(p *proyecto.Proyecto) { p.FechaFin = p.FechaInicio.AddDate(0, 0, -1) }},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			p := proyectoDePrueba("Demo")
			c.editar(&p)

			if _, err := repo.RegistrarProyecto(ctx, p); err == nil {
				t.Fatal("se esperaba que el INSERT fallara por un CHECK de la tabla")
			}
			if n := dbprueba.ContarFilas(ctx, t, pool, "proyectos"); n != 0 {
				t.Errorf("proyectos = %d, se esperaba ninguno", n)
			}
		})
	}
}