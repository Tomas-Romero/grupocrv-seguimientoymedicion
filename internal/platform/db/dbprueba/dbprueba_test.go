package dbprueba_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/db/dbprueba"
)

// La base temporal existe mientras dura el test y se borra al terminar: si
// quedara, los tests irian llenando el servidor de bases sueltas.
func TestBaseTemporal_SeBorraAlTerminar(t *testing.T) {
	var nombre string
	t.Run("usa una base temporal", func(t *testing.T) {
		pool := dbprueba.BaseTemporal(t)
		if err := pool.QueryRow(context.Background(), "SELECT current_database()").Scan(&nombre); err != nil {
			t.Fatalf("leer el nombre de la base: %v", err)
		}
	})
	if nombre == "" {
		t.Skip("el subtest se omitio: no hay DATABASE_URL")
	}

	ctx, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	admin, err := db.Conectar(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("conectar a la base de DATABASE_URL: %v", err)
	}
	defer admin.Close()

	var existe bool
	if err := admin.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", nombre).Scan(&existe); err != nil {
		t.Fatalf("buscar la base %s: %v", nombre, err)
	}
	if existe {
		t.Errorf("la base temporal %s sigue existiendo despues del test", nombre)
	}
}

// BaseMigrada deja el esquema aplicado y las tablas vacias.
func TestBaseMigrada(t *testing.T) {
	ctx, cancelar := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelar()
	pool := dbprueba.BaseMigrada(ctx, t)

	if !dbprueba.ExisteTabla(ctx, t, pool, "proyectos") {
		t.Fatal("despues de migrar no existe la tabla proyectos")
	}
	if dbprueba.ExisteTabla(ctx, t, pool, "tabla_que_no_existe") {
		t.Error("ExisteTabla encontro una tabla que no existe")
	}
	if n := dbprueba.ContarFilas(ctx, t, pool, "proyectos"); n != 0 {
		t.Errorf("proyectos = %d, se esperaba una base vacia", n)
	}
}
