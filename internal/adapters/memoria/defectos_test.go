package memoria_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// baseConDefectos es una base con dos proyectos: "demo", con una historia y un sprint, y
// "otro", con su propia historia y su propio sprint.
type baseConDefectos struct {
	base               *memoria.Base
	demo, otro         string
	item, itemOtro     string
	sprint, sprintOtro string
}

func nuevaBaseConDefectos(t *testing.T) baseConDefectos {
	t.Helper()
	ctx := context.Background()
	base := memoria.NuevaBase()
	b := baseConDefectos{base: base}
	b.demo = proyectoParaSprints(t, base, "Demo")
	b.otro = proyectoParaSprints(t, base, "Otro")

	item, err := base.RegistrarItem(ctx, b.demo, itemValido(t, "Alta de proyectos"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	itemOtro, err := base.RegistrarItem(ctx, b.otro, itemValido(t, "Listar proyectos"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	b.item, b.itemOtro = item.ID, itemOtro.ID

	s, err := base.RegistrarSprint(ctx, sprintParaProyecto(t, b.demo, diaDeSprint(10, 6), diaDeSprint(10, 19)))
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}
	sOtro, err := base.RegistrarSprint(ctx, sprintParaProyecto(t, b.otro, diaDeSprint(10, 6), diaDeSprint(10, 19)))
	if err != nil {
		t.Fatalf("RegistrarSprint: %v", err)
	}
	b.sprint, b.sprintOtro = s.ID, sOtro.ID
	return b
}

func (b baseConDefectos) defecto(t *testing.T, descripcion string) defecto.Defecto {
	t.Helper()
	d, err := defecto.Nuevo(b.demo, b.item, b.sprint, descripcion, defecto.SeveridadAlta)
	if err != nil {
		t.Fatalf("defecto.Nuevo: %v", err)
	}
	return d
}

// US-022 / RN-022-5: la historia y el sprint se buscan dentro del proyecto. Los de otro
// proyecto, un ID que no esta, uno mal formado o uno vacio no cuentan.
func TestBase_ExisteItemYSprintEnProyecto(t *testing.T) {
	b := nuevaBaseConDefectos(t)
	ctx := context.Background()

	casos := []struct {
		nombre   string
		proyecto string
		id       string
		item     bool
		quiere   bool
	}{
		{"historia del proyecto", b.demo, b.item, true, true},
		{"historia de otro proyecto", b.demo, b.itemOtro, true, false},
		{"historia que no esta", b.demo, idSinDuenoEnMemoria, true, false},
		{"historia mal formada", b.demo, "no-es-un-uuid", true, false},
		{"historia vacia", b.demo, "", true, false},
		{"sprint del proyecto", b.demo, b.sprint, false, true},
		{"sprint de otro proyecto", b.demo, b.sprintOtro, false, false},
		{"sprint que no esta", b.demo, idSinDuenoEnMemoria, false, false},
		{"sprint mal formado", b.demo, "no-es-un-uuid", false, false},
		{"sprint vacio", b.demo, "", false, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			var existe bool
			var err error
			if c.item {
				existe, err = b.base.ExisteItemEnProyecto(ctx, c.proyecto, c.id)
			} else {
				existe, err = b.base.ExisteSprintEnProyecto(ctx, c.proyecto, c.id)
			}
			if err != nil || existe != c.quiere {
				t.Errorf("existe = (%t, %v), quiero (%t, nil)", existe, err, c.quiere)
			}
		})
	}
}

// US-022 / CA-022-1 / RN-022-10 / CL-022-14: cada defecto recibe un ID propio, como el
// RETURNING de Postgres, y queda guardado abierto y sin sprint de resolucion.
func TestBase_RegistrarDefecto(t *testing.T) {
	b := nuevaBaseConDefectos(t)

	guardado, err := b.base.RegistrarDefecto(context.Background(), b.defecto(t, "Falla al guardar"))
	if err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}

	if guardado.ID == "" {
		t.Fatal("el defecto volvio sin ID")
	}
	defectos := b.base.Defectos(b.demo)
	if len(defectos) != 1 || defectos[0] != guardado {
		t.Fatalf("defectos guardados = %+v, quiero el devuelto %+v", defectos, guardado)
	}
	if guardado.Estado != defecto.EstadoAbierto || guardado.SprintResolucionID != "" || guardado.SprintDeteccionID != b.sprint {
		t.Errorf("defecto = %+v, quiero abierto, sin sprint de resolucion y detectado en %s", guardado, b.sprint)
	}
	if otros := b.base.Defectos(b.otro); len(otros) != 0 {
		t.Errorf("el otro proyecto tiene %d defectos, quiero ninguno", len(otros))
	}
}

// US-022 / RN-022-8 / CL-022-13: dos defectos iguales son dos defectos con IDs distintos.
func TestBase_RegistrarDefecto_Repetidos(t *testing.T) {
	b := nuevaBaseConDefectos(t)
	ctx := context.Background()

	primero, err1 := b.base.RegistrarDefecto(ctx, b.defecto(t, "Falla"))
	segundo, err2 := b.base.RegistrarDefecto(ctx, b.defecto(t, "Falla"))

	if err1 != nil || err2 != nil {
		t.Fatalf("errores = %v, %v; quiero que los dos se guarden", err1, err2)
	}
	if primero.ID == segundo.ID {
		t.Errorf("los dos defectos recibieron el ID %q", primero.ID)
	}
	if n := len(b.base.Defectos(b.demo)); n != 2 {
		t.Errorf("hay %d defectos guardados, quiero 2", n)
	}
}

// US-022 / RN-022-10 / CL-022-15: replica las claves foraneas de la tabla. Si el proyecto, la
// historia o el sprint no estan devuelve el error de inexistencia que corresponde, en ese
// orden, y no guarda nada.
func TestBase_RegistrarDefecto_LoQueNoExiste(t *testing.T) {
	b := nuevaBaseConDefectos(t)

	casos := []struct {
		nombre  string
		cambiar func(d *defecto.Defecto)
		quiere  error
	}{
		{"el proyecto no esta", func(d *defecto.Defecto) { d.ProyectoID = idSinDuenoEnMemoria }, app.ErrProyectoInexistente},
		{"la historia no esta", func(d *defecto.Defecto) { d.ItemID = idSinDuenoEnMemoria }, app.ErrItemInexistente},
		{"el sprint no esta", func(d *defecto.Defecto) { d.SprintDeteccionID = idSinDuenoEnMemoria }, app.ErrSprintInexistente},
		{"el sprint esta vacio", func(d *defecto.Defecto) { d.SprintDeteccionID = "" }, app.ErrSprintInexistente},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := b.defecto(t, "Falla")
			c.cambiar(&d)

			_, err := b.base.RegistrarDefecto(context.Background(), d)

			if !errors.Is(err, c.quiere) {
				t.Errorf("error = %v, quiero %v", err, c.quiere)
			}
			if n := len(b.base.Defectos(b.demo)); n != 0 {
				t.Errorf("hay %d defectos guardados, quiero ninguno", n)
			}
		})
	}
}
