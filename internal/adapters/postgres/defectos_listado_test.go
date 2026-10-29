package postgres_test

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/postgres"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// T-021 / RN-T021-2: el repositorio de defectos cumple el puerto de ListarDefectos.
var _ app.RepositorioListarDefectos = (*postgres.RepositorioDefectos)(nil)

// T-021 / RN-T021-3, RN-T021-13 / CA-T021-6: ListarDefectosDelProyecto devuelve los defectos del
// proyecto, de cualquier estado, con el numero y el titulo de la historia y los numeros de los
// sprints de deteccion y de resolucion que trae el JOIN (RN-022-6). Los defectos de otro proyecto
// no aparecen.
func TestRepositorioDefectos_ListarDefectosDelProyecto(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)
	abierto := e.registrado(ctx, t)
	porResolver := e.registrado(ctx, t)
	otro, err := defecto.Nuevo(e.otro, e.itemOtro, e.sprintOtro, "De otro proyecto", defecto.SeveridadBaja)
	if err != nil {
		t.Fatalf("defecto.Nuevo: %v", err)
	}
	if _, err := e.repo.RegistrarDefecto(ctx, otro); err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}
	resuelto := porResolver
	resuelto.Estado, resuelto.SprintResolucionID = defecto.EstadoResuelto, e.sprints[2]
	if err := e.repo.ActualizarEstadoDefecto(ctx, resuelto, defecto.EstadoAbierto); err != nil {
		t.Fatalf("ActualizarEstadoDefecto: %v", err)
	}

	got, err := e.repo.ListarDefectosDelProyecto(ctx, e.proyecto)
	if err != nil || len(got) != 2 {
		t.Fatalf("ListarDefectosDelProyecto = %+v, %v; quiero los dos defectos de Demo", got, err)
	}
	porID := map[string]app.DefectoListado{}
	for _, l := range got {
		porID[l.Defecto.ID] = l
	}

	if l := porID[abierto.ID]; l.Defecto != abierto || l.HistoriaNumero != 1 || l.HistoriaTitulo != "Crear proyecto" ||
		l.SprintDeteccion != 1 || l.SprintResolucion != 0 {
		t.Errorf("defecto abierto = %+v, quiero los datos de %+v, historia #1, Sprint 1 y sin sprint de resolucion", l, abierto)
	}
	if l := porID[resuelto.ID]; l.Defecto != resuelto || l.SprintDeteccion != 1 || l.SprintResolucion != 3 {
		t.Errorf("defecto resuelto = %+v, quiero los datos de %+v, Sprint 1 de deteccion y Sprint 3 de resolucion", l, resuelto)
	}
	for _, l := range got {
		if l.CreadoEn.IsZero() {
			t.Errorf("el defecto %s no trae su fecha de carga (creado_en)", l.Defecto.ID)
		}
	}
}

// T-021 / RN-T021-3, CL-T021-1: un proyecto sin defectos devuelve una lista vacia y sin error; un
// ID que no es de ningun proyecto o esta mal formado, tambien.
func TestRepositorioDefectos_ListarDefectosDelProyecto_Vacio(t *testing.T) {
	ctx := contexto(t)
	e := nuevoEscenarioDefectos(ctx, t)

	for _, proyecto := range []string{e.proyecto, "00000000-0000-0000-0000-000000000000", "no-es-un-uuid", ""} {
		got, err := e.repo.ListarDefectosDelProyecto(ctx, proyecto)
		if err != nil || got == nil || len(got) != 0 {
			t.Errorf("ListarDefectosDelProyecto(%q) = %v, %v; quiero una lista vacia y sin error", proyecto, got, err)
		}
	}
}

// proyeccionDeListado son los datos de una fila de la lista que no dependen de como cada
// adaptador genera los IDs ni la fecha de carga.
type proyeccionDeListado struct {
	descripcion      string
	severidad        defecto.Severidad
	estado           defecto.Estado
	historiaNumero   int
	historiaTitulo   string
	sprintDeteccion  int
	sprintResolucion int
}

func proyectar(listados []app.DefectoListado) []proyeccionDeListado {
	salida := make([]proyeccionDeListado, len(listados))
	for i, l := range listados {
		salida[i] = proyeccionDeListado{l.Defecto.Descripcion, l.Defecto.Severidad, l.Defecto.Estado,
			l.HistoriaNumero, l.HistoriaTitulo, l.SprintDeteccion, l.SprintResolucion}
	}
	return salida
}

// T-021 / RN-T021-2, RN-T021-4 / CA-T021-6: con los mismos datos, ListarDefectos devuelve lo mismo
// sobre la base en memoria y sobre Postgres: los datos del defecto, los numeros de la historia y de
// los sprints, y el orden del mas reciente al mas antiguo.
//
// US-024 / CA-024-1, CA-024-2, CA-024-3, CA-024-4: tambien filtrando. El filtro se aplica en la aplicacion
// sobre lo que entrega cada repositorio (spec US-024, seccion 5), asi que la memoria y Postgres no pueden
// filtrar distinto; y el sprint de otro proyecto o mal formado se rechaza igual en las dos.
func TestListarDefectos_MemoriaYPostgresDevuelvenLoMismo(t *testing.T) {
	ctx := contexto(t)

	// Postgres: Demo con tres sprints y una historia (el escenario de los demas tests de defectos).
	e := nuevoEscenarioDefectos(ctx, t)
	registrarEnPostgres := func(descripcion string, severidad defecto.Severidad, sprintDeteccion int) defecto.Defecto {
		d, err := e.repo.RegistrarDefecto(ctx, e.defecto(t, descripcion, severidad, sprintDeteccion))
		if err != nil {
			t.Fatalf("RegistrarDefecto %q: %v", descripcion, err)
		}
		return d
	}
	a := registrarEnPostgres("Primero", defecto.SeveridadAlta, 1)
	b := registrarEnPostgres("Segundo", defecto.SeveridadCritica, 3)
	registrarEnPostgres("Tercero", defecto.SeveridadBaja, 2)
	a.Estado, a.SprintResolucionID = defecto.EstadoResuelto, e.sprints[2]
	b.Estado = defecto.EstadoDescartado
	if err := e.repo.ActualizarEstadoDefecto(ctx, a, defecto.EstadoAbierto); err != nil {
		t.Fatalf("ActualizarEstadoDefecto: %v", err)
	}
	if err := e.repo.ActualizarEstadoDefecto(ctx, b, defecto.EstadoAbierto); err != nil {
		t.Fatalf("ActualizarEstadoDefecto: %v", err)
	}

	// Memoria: lo mismo, con sus propios IDs.
	base := memoria.NuevaBase()
	p, err := base.RegistrarProyecto(ctx, proyecto.Proyecto{Nombre: "Demo", FechaInicio: diaSprint(10, 6), FechaFin: diaSprint(12, 31)})
	if err != nil {
		t.Fatalf("RegistrarProyecto: %v", err)
	}
	item, err := base.RegistrarItem(ctx, p.ID, itemConTitulo(t, "Crear proyecto"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	var sprints [3]string
	fechas := [3][2]time.Time{
		{diaSprint(10, 6), diaSprint(10, 19)},
		{diaSprint(10, 20), diaSprint(11, 2)},
		{diaSprint(11, 3), diaSprint(11, 16)},
	}
	for i, f := range fechas {
		s, err := base.RegistrarSprint(ctx, sprintNuevo(t, p.ID, f[0], f[1]))
		if err != nil {
			t.Fatalf("RegistrarSprint %d: %v", i+1, err)
		}
		sprints[i] = s.ID
		// Un sprint nuevo exige que el ultimo este cerrado (RN-008-13); el tercero queda abierto.
		if i < 2 {
			if _, err := base.GuardarCierre(ctx, s.ID); err != nil {
				t.Fatalf("GuardarCierre: %v", err)
			}
		}
	}
	registrarEnMemoria := func(descripcion string, severidad defecto.Severidad, sprintDeteccion int) defecto.Defecto {
		d, err := defecto.Nuevo(p.ID, item.ID, sprints[sprintDeteccion-1], descripcion, severidad)
		if err != nil {
			t.Fatalf("defecto.Nuevo: %v", err)
		}
		guardado, err := base.RegistrarDefecto(ctx, d)
		if err != nil {
			t.Fatalf("RegistrarDefecto %q: %v", descripcion, err)
		}
		return guardado
	}
	ma := registrarEnMemoria("Primero", defecto.SeveridadAlta, 1)
	mb := registrarEnMemoria("Segundo", defecto.SeveridadCritica, 3)
	registrarEnMemoria("Tercero", defecto.SeveridadBaja, 2)
	ma.Estado, ma.SprintResolucionID = defecto.EstadoResuelto, sprints[2]
	mb.Estado = defecto.EstadoDescartado
	if err := base.ActualizarEstadoDefecto(ctx, ma, defecto.EstadoAbierto); err != nil {
		t.Fatalf("ActualizarEstadoDefecto: %v", err)
	}
	if err := base.ActualizarEstadoDefecto(ctx, mb, defecto.EstadoAbierto); err != nil {
		t.Fatalf("ActualizarEstadoDefecto: %v", err)
	}

	enPostgres, err := app.NuevoListarDefectos(e.repo).Ejecutar(ctx, e.proyecto, app.FiltroDefectos{})
	if err != nil {
		t.Fatalf("ListarDefectos sobre Postgres: %v", err)
	}
	enMemoria, err := app.NuevoListarDefectos(base).Ejecutar(ctx, p.ID, app.FiltroDefectos{})
	if err != nil {
		t.Fatalf("ListarDefectos sobre la memoria: %v", err)
	}

	quiero := []proyeccionDeListado{
		{"Tercero", defecto.SeveridadBaja, defecto.EstadoAbierto, 1, "Crear proyecto", 2, 0},
		{"Segundo", defecto.SeveridadCritica, defecto.EstadoDescartado, 1, "Crear proyecto", 3, 0},
		{"Primero", defecto.SeveridadAlta, defecto.EstadoResuelto, 1, "Crear proyecto", 1, 3},
	}
	if got := proyectar(enPostgres); !slices.Equal(got, quiero) {
		t.Errorf("Postgres = %+v, quiero %+v", got, quiero)
	}
	if got := proyectar(enMemoria); !slices.Equal(got, quiero) {
		t.Errorf("memoria = %+v, quiero %+v", got, quiero)
	}

	// US-024: Primero es del Sprint 1 y alta, Segundo del Sprint 3 y critica, Tercero del Sprint 2 y baja.
	descripciones := func(listados []app.DefectoListado) []string {
		salida := []string{}
		for _, l := range listados {
			salida = append(salida, l.Defecto.Descripcion)
		}
		return salida
	}
	filtros := []struct {
		nombre    string
		sprint    int // 0: sin filtro de sprint
		severidad defecto.Severidad
		quiero    []string
	}{
		{"por sprint", 1, "", []string{"Primero"}},
		{"por sprint en el que se resolvio pero no se detecto", 3, "alta", []string{}},
		{"por severidad", 0, defecto.SeveridadCritica, []string{"Segundo"}},
		{"por sprint y severidad", 3, defecto.SeveridadCritica, []string{"Segundo"}},
		{"sin coincidencias", 2, defecto.SeveridadAlta, []string{}},
	}
	for _, c := range filtros {
		t.Run(c.nombre, func(t *testing.T) {
			filtroPostgres, filtroMemoria := app.FiltroDefectos{Severidad: c.severidad}, app.FiltroDefectos{Severidad: c.severidad}
			if c.sprint != 0 {
				filtroPostgres.SprintDeteccionID, filtroMemoria.SprintDeteccionID = e.sprints[c.sprint-1], sprints[c.sprint-1]
			}
			enPostgres, err := app.NuevoListarDefectos(e.repo).Ejecutar(ctx, e.proyecto, filtroPostgres)
			if err != nil {
				t.Fatalf("ListarDefectos filtrado sobre Postgres: %v", err)
			}
			enMemoria, err := app.NuevoListarDefectos(base).Ejecutar(ctx, p.ID, filtroMemoria)
			if err != nil {
				t.Fatalf("ListarDefectos filtrado sobre la memoria: %v", err)
			}
			if got := descripciones(enPostgres); !slices.Equal(got, c.quiero) {
				t.Errorf("Postgres = %v, quiero %v", got, c.quiero)
			}
			if got := descripciones(enMemoria); !slices.Equal(got, c.quiero) {
				t.Errorf("memoria = %v, quiero %v", got, c.quiero)
			}
		})
	}

	// US-024 / RN-024-7: un sprint de otro proyecto o mal formado es ErrSprintInexistente, no una lista vacia.
	for _, sprintID := range []string{e.sprintOtro, "no-es-un-uuid"} {
		if _, err := app.NuevoListarDefectos(e.repo).Ejecutar(ctx, e.proyecto, app.FiltroDefectos{SprintDeteccionID: sprintID}); !errors.Is(err, app.ErrSprintInexistente) {
			t.Errorf("Postgres con el sprint %q: error = %v, quiero ErrSprintInexistente", sprintID, err)
		}
	}
}
