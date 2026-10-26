package app_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// lecturaDeEsfuerzoFalsa simula la lectura de US-020 sin base de datos.
type lecturaDeEsfuerzoFalsa struct {
	proyectos map[string]bool
	lectura   app.LecturaEsfuerzo
	err       error
	lecturas  int
}

func (l *lecturaDeEsfuerzoFalsa) ExisteProyecto(_ context.Context, id string) (bool, error) {
	return l.proyectos[id], l.err
}

func (l *lecturaDeEsfuerzoFalsa) LeerEsfuerzo(context.Context, string) (app.LecturaEsfuerzo, error) {
	l.lecturas++
	return l.lectura, l.err
}

func diaEsf(mes time.Month, d int) time.Time { return time.Date(2026, mes, d, 0, 0, 0, 0, time.UTC) }

func registroEsf(id, item, integrante string, fecha time.Time, horas float64) esfuerzo.RegistroEsfuerzo {
	return esfuerzo.RegistroEsfuerzo{ID: id, ProyectoID: "p1", ItemID: item, IntegranteID: integrante,
		Fecha: fecha, Actividad: "Trabajo " + id, Horas: horas}
}

// lecturaBase tiene dos sprints (del 6 al 19 de octubre y del 21 de octubre al 3 de noviembre),
// dos historias, dos integrantes y cinco registros. Uno cae entre los dos sprints.
func lecturaBase() app.LecturaEsfuerzo {
	return app.LecturaEsfuerzo{
		Registros: []esfuerzo.RegistroEsfuerzo{
			registroEsf("r1", "h1", "i1", diaEsf(time.October, 7), 2),
			registroEsf("r2", "h1", "i2", diaEsf(time.October, 8), 3),
			registroEsf("r3", "h2", "i1", diaEsf(time.October, 8), 1.5),
			registroEsf("r4", "h2", "i1", diaEsf(time.October, 22), 4),
			registroEsf("r5", "h1", "i2", diaEsf(time.October, 20), 1),
		},
		Historias:   map[string]string{"h1": "Alta de proyectos", "h2": "Backlog"},
		Integrantes: map[string]string{"i1": "Ana", "i2": "Beto"},
		Sprints: []sprint.Sprint{
			{ID: "s2", Numero: 2, ProyectoID: "p1", FechaInicio: diaEsf(time.October, 21), FechaFin: diaEsf(time.November, 3)},
			{ID: "s1", Numero: 1, ProyectoID: "p1", FechaInicio: diaEsf(time.October, 6), FechaFin: diaEsf(time.October, 19)},
		},
	}
}

func consultarEsf(t *testing.T, lectura app.LecturaEsfuerzo, filtro app.FiltroEsfuerzo) (app.ConsultaEsfuerzo, error) {
	t.Helper()
	repo := &lecturaDeEsfuerzoFalsa{proyectos: map[string]bool{"p1": true}, lectura: lectura}
	return app.NuevoConsultarEsfuerzo(repo).Ejecutar(context.Background(), "p1", filtro)
}

func idsDe(c app.ConsultaEsfuerzo) []string {
	ids := make([]string, len(c.Registros))
	for i, r := range c.Registros {
		ids[i] = r.Registro.ID
	}
	return ids
}

// US-020 / CA-020-1, RN-020-1, RN-020-5, RN-020-6: sin filtros salen todos los registros, del
// mas reciente al mas antiguo, y con la misma fecha por titulo de historia.
func TestConsultarEsfuerzo_SinFiltros(t *testing.T) {
	c, err := consultarEsf(t, lecturaBase(), app.FiltroEsfuerzo{})
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if esperado := []string{"r4", "r5", "r2", "r3", "r1"}; !reflect.DeepEqual(idsDe(c), esperado) {
		t.Errorf("orden = %v, se esperaba %v", idsDe(c), esperado)
	}
	if c.TotalHoras != 11.5 {
		t.Errorf("total = %v, se esperaba 11.5", c.TotalHoras)
	}
}

// US-020 / CA-020-6, RN-020-3: cada registro lleva el titulo de su historia, el nombre del integrante y el
// numero del sprint donde cae su fecha.
func TestConsultarEsfuerzo_CadaRegistroLlevaSuDetalle(t *testing.T) {
	c, _ := consultarEsf(t, lecturaBase(), app.FiltroEsfuerzo{})

	porID := map[string]app.RegistroConsultado{}
	for _, r := range c.Registros {
		porID[r.Registro.ID] = r
	}
	casos := []struct {
		id, historia, integrante string
		sprint                   int
	}{
		{"r1", "Alta de proyectos", "Ana", 1},
		{"r3", "Backlog", "Ana", 1},
		{"r4", "Backlog", "Ana", 2},
		{"r5", "Alta de proyectos", "Beto", 0}, // el 20 de octubre no es de ningun sprint
	}
	for _, caso := range casos {
		r := porID[caso.id]
		if r.Historia != caso.historia || r.Integrante != caso.integrante || r.SprintNumero != caso.sprint {
			t.Errorf("%s = {%q, %q, sprint %d}, se esperaba {%q, %q, sprint %d}",
				caso.id, r.Historia, r.Integrante, r.SprintNumero, caso.historia, caso.integrante, caso.sprint)
		}
	}
}

// US-020 / CA-020-5, RN-020-8, RN-020-9: los totales salen ordenados y cada lista suma el total.
func TestConsultarEsfuerzo_TotalesPorGrupo(t *testing.T) {
	c, _ := consultarEsf(t, lecturaBase(), app.FiltroEsfuerzo{})

	historias := []app.TotalPorHistoria{{ItemID: "h1", Titulo: "Alta de proyectos", Horas: 6}, {ItemID: "h2", Titulo: "Backlog", Horas: 5.5}}
	if !reflect.DeepEqual(c.PorHistoria, historias) {
		t.Errorf("por historia = %+v, se esperaba %+v", c.PorHistoria, historias)
	}
	integrantes := []app.TotalPorIntegrante{{IntegranteID: "i1", Nombre: "Ana", Horas: 7.5}, {IntegranteID: "i2", Nombre: "Beto", Horas: 4}}
	if !reflect.DeepEqual(c.PorIntegrante, integrantes) {
		t.Errorf("por integrante = %+v, se esperaba %+v", c.PorIntegrante, integrantes)
	}
	// RN-020-9: por numero de sprint, y lo que cae fuera de todo sprint al final.
	sprints := []app.TotalPorSprint{{SprintID: "s1", Numero: 1, Horas: 6.5}, {SprintID: "s2", Numero: 2, Horas: 4}, {Horas: 1}}
	if !reflect.DeepEqual(c.PorSprint, sprints) {
		t.Errorf("por sprint = %+v, se esperaba %+v", c.PorSprint, sprints)
	}
}

// US-020 / CA-020-2, RN-020-2, RN-020-7: filtrar por historia deja sus registros, y el total y los
// totales por integrante son de lo filtrado.
func TestConsultarEsfuerzo_FiltroPorHistoria(t *testing.T) {
	c, err := consultarEsf(t, lecturaBase(), app.FiltroEsfuerzo{ItemID: "h1"})
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if esperado := []string{"r5", "r2", "r1"}; !reflect.DeepEqual(idsDe(c), esperado) {
		t.Errorf("registros = %v, se esperaba %v", idsDe(c), esperado)
	}
	if c.TotalHoras != 6 {
		t.Errorf("total = %v, se esperaba 6", c.TotalHoras)
	}
	if esperado := []app.TotalPorIntegrante{{IntegranteID: "i2", Nombre: "Beto", Horas: 4}, {IntegranteID: "i1", Nombre: "Ana", Horas: 2}}; !reflect.DeepEqual(c.PorIntegrante, esperado) {
		t.Errorf("por integrante = %+v, se esperaba %+v", c.PorIntegrante, esperado)
	}
}

// US-020 / CA-020-3: filtrar por integrante deja sus registros de todas las historias.
func TestConsultarEsfuerzo_FiltroPorIntegrante(t *testing.T) {
	c, err := consultarEsf(t, lecturaBase(), app.FiltroEsfuerzo{IntegranteID: "i1"})
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if esperado := []string{"r4", "r3", "r1"}; !reflect.DeepEqual(idsDe(c), esperado) {
		t.Errorf("registros = %v, se esperaba %v", idsDe(c), esperado)
	}
	if c.TotalHoras != 7.5 {
		t.Errorf("total = %v, se esperaba 7.5", c.TotalHoras)
	}
}

// US-020 / CA-020-4, RN-020-4, CL-020-3: filtrar por sprint deja los registros cuya fecha cae en
// el sprint, incluidos el primer y el ultimo dia. El del 20 de octubre no es de ninguno.
func TestConsultarEsfuerzo_FiltroPorSprint(t *testing.T) {
	lectura := lecturaBase()
	lectura.Registros = append(lectura.Registros,
		registroEsf("r6", "h2", "i2", diaEsf(time.October, 6), 1),  // primer dia del Sprint 1
		registroEsf("r7", "h2", "i2", diaEsf(time.October, 19), 1), // ultimo dia del Sprint 1
	)

	c, err := consultarEsf(t, lectura, app.FiltroEsfuerzo{SprintID: "s1"})
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if esperado := []string{"r7", "r2", "r3", "r1", "r6"}; !reflect.DeepEqual(idsDe(c), esperado) {
		t.Errorf("registros = %v, se esperaba %v", idsDe(c), esperado)
	}
	if c.TotalHoras != 8.5 {
		t.Errorf("total = %v, se esperaba 8.5", c.TotalHoras)
	}
	// RN-020-10 (US-027): las horas salen en el mismo orden que los registros.
	if esperado := []float64{1, 3, 1.5, 2, 1}; !reflect.DeepEqual(c.Horas(), esperado) {
		t.Errorf("Horas() = %v, se esperaba %v", c.Horas(), esperado)
	}
}

// US-020 / RN-020-4, CL-020-5: una historia con horas en dos sprints aparece en cada uno con las
// horas que se trabajaron en el.
func TestConsultarEsfuerzo_LasHorasSeQuedanEnElSprintDondeSeTrabajaron(t *testing.T) {
	lectura := lecturaBase()
	lectura.Registros = []esfuerzo.RegistroEsfuerzo{
		registroEsf("r1", "h1", "i1", diaEsf(time.October, 10), 3),
		registroEsf("r2", "h1", "i1", diaEsf(time.October, 23), 5),
	}

	uno, _ := consultarEsf(t, lectura, app.FiltroEsfuerzo{SprintID: "s1"})
	dos, _ := consultarEsf(t, lectura, app.FiltroEsfuerzo{SprintID: "s2"})

	if uno.TotalHoras != 3 || dos.TotalHoras != 5 {
		t.Errorf("Sprint 1 = %v h y Sprint 2 = %v h, se esperaban 3 y 5", uno.TotalHoras, dos.TotalHoras)
	}
}

// US-020 / RN-020-2, CL-020-6: los filtros se combinan con Y.
func TestConsultarEsfuerzo_FiltrosCombinados(t *testing.T) {
	c, err := consultarEsf(t, lecturaBase(), app.FiltroEsfuerzo{ItemID: "h2", IntegranteID: "i1", SprintID: "s2"})
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if esperado := []string{"r4"}; !reflect.DeepEqual(idsDe(c), esperado) {
		t.Errorf("registros = %v, se esperaba %v", idsDe(c), esperado)
	}
}

// US-020 / CA-020-7, RN-020-10, CL-020-1, CL-020-7: sin registros, o sin coincidencias, el
// resultado es vacio y no es un error.
func TestConsultarEsfuerzo_SinRegistros(t *testing.T) {
	casos := map[string]struct {
		lectura app.LecturaEsfuerzo
		filtro  app.FiltroEsfuerzo
	}{
		"proyecto sin registros":   {app.LecturaEsfuerzo{}, app.FiltroEsfuerzo{}},
		"filtro sin coincidencias": {lecturaBase(), app.FiltroEsfuerzo{ItemID: "h2", IntegranteID: "i2"}},
	}
	for nombre, caso := range casos {
		c, err := consultarEsf(t, caso.lectura, caso.filtro)
		if err != nil {
			t.Errorf("%s: se esperaba sin error, llego %v", nombre, err)
			continue
		}
		if len(c.Registros) != 0 || c.TotalHoras != 0 || len(c.PorHistoria) != 0 || len(c.PorIntegrante) != 0 || len(c.PorSprint) != 0 {
			t.Errorf("%s: se esperaba un resultado vacio, llego %+v", nombre, c)
		}
	}
}

// US-020 / CL-020-2, CL-020-12: dos registros iguales son dos registros, y los cuartos de hora
// suman exacto.
func TestConsultarEsfuerzo_RegistrosIgualesSumanLosDos(t *testing.T) {
	lectura := lecturaBase()
	lectura.Registros = []esfuerzo.RegistroEsfuerzo{
		registroEsf("r1", "h1", "i1", diaEsf(time.October, 7), 0.25),
		registroEsf("r2", "h1", "i1", diaEsf(time.October, 7), 0.25),
		registroEsf("r3", "h1", "i1", diaEsf(time.October, 7), 0.5),
	}

	c, _ := consultarEsf(t, lectura, app.FiltroEsfuerzo{})

	if len(c.Registros) != 3 || c.TotalHoras != 1 {
		t.Errorf("registros = %d, total = %v; se esperaban 3 y 1", len(c.Registros), c.TotalHoras)
	}
}

// US-020 / RN-020-5: con la misma fecha y la misma historia desempata el integrante, despues la
// actividad y al final el ID, sin distinguir mayusculas.
func TestConsultarEsfuerzo_OrdenTotal(t *testing.T) {
	fecha := diaEsf(time.October, 8)
	lectura := lecturaBase()
	lectura.Historias = map[string]string{"h1": "alta", "h2": "Alta"}
	lectura.Registros = []esfuerzo.RegistroEsfuerzo{
		{ID: "d", ItemID: "h1", IntegranteID: "i2", Fecha: fecha, Actividad: "x", Horas: 1},
		{ID: "c", ItemID: "h1", IntegranteID: "i1", Fecha: fecha, Actividad: "B", Horas: 1},
		{ID: "b", ItemID: "h1", IntegranteID: "i1", Fecha: fecha, Actividad: "a", Horas: 1},
		{ID: "a", ItemID: "h1", IntegranteID: "i1", Fecha: fecha, Actividad: "a", Horas: 1},
	}

	c, _ := consultarEsf(t, lectura, app.FiltroEsfuerzo{})

	if esperado := []string{"a", "b", "c", "d"}; !reflect.DeepEqual(idsDe(c), esperado) {
		t.Errorf("orden = %v, se esperaba %v", idsDe(c), esperado)
	}
}

// US-020 / RN-020-9, CL-020-13: dos historias con las mismas horas salen por titulo.
func TestConsultarEsfuerzo_TotalesEmpatadosSalenPorNombre(t *testing.T) {
	lectura := lecturaBase()
	lectura.Registros = []esfuerzo.RegistroEsfuerzo{
		registroEsf("r1", "h2", "i2", diaEsf(time.October, 7), 2),
		registroEsf("r2", "h1", "i1", diaEsf(time.October, 7), 2),
	}

	c, _ := consultarEsf(t, lectura, app.FiltroEsfuerzo{})

	if c.PorHistoria[0].ItemID != "h1" || c.PorIntegrante[0].IntegranteID != "i1" {
		t.Errorf("por historia = %+v, por integrante = %+v; con empate sale primero el nombre menor",
			c.PorHistoria, c.PorIntegrante)
	}
}

// US-020 / RN-020-11, CL-020-15: los registros de un sprint cerrado se cuentan igual.
func TestConsultarEsfuerzo_SprintCerrado(t *testing.T) {
	lectura := lecturaBase()
	lectura.Sprints[1].Estado = sprint.EstadoCerrado

	c, err := consultarEsf(t, lectura, app.FiltroEsfuerzo{SprintID: "s1"})
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if c.TotalHoras != 6.5 {
		t.Errorf("total = %v, se esperaba 6.5", c.TotalHoras)
	}
}

// US-020 / CA-020-8, CA-020-9, CL-020-8 a CL-020-11: orden de las verificaciones y errores de
// inexistencia. Una historia, un integrante o un sprint que no son del proyecto (aunque existan en
// otro) no estan en la lectura.
func TestConsultarEsfuerzo_Inexistentes(t *testing.T) {
	casos := []struct {
		nombre   string
		proyecto string
		filtro   app.FiltroEsfuerzo
		esperado error
	}{
		{"proyecto", "otro", app.FiltroEsfuerzo{}, app.ErrProyectoInexistente},
		{"historia", "p1", app.FiltroEsfuerzo{ItemID: "ajena"}, app.ErrItemInexistente},
		{"integrante", "p1", app.FiltroEsfuerzo{IntegranteID: "ajeno"}, app.ErrIntegranteInexistente},
		{"sprint", "p1", app.FiltroEsfuerzo{SprintID: "ajeno"}, app.ErrSprintInexistente},
		{"id mal formado", "p1", app.FiltroEsfuerzo{ItemID: "no-es-un-uuid"}, app.ErrItemInexistente},
		// CL-020-10: el proyecto tiene prioridad sobre los filtros.
		{"proyecto y filtro", "otro", app.FiltroEsfuerzo{ItemID: "ajena"}, app.ErrProyectoInexistente},
		// CL-020-11: historia, integrante y sprint, en ese orden.
		{"historia y sprint", "p1", app.FiltroEsfuerzo{ItemID: "ajena", SprintID: "ajeno"}, app.ErrItemInexistente},
		{"integrante y sprint", "p1", app.FiltroEsfuerzo{IntegranteID: "ajeno", SprintID: "ajeno"}, app.ErrIntegranteInexistente},
	}
	for _, c := range casos {
		repo := &lecturaDeEsfuerzoFalsa{proyectos: map[string]bool{"p1": true}, lectura: lecturaBase()}
		resultado, err := app.NuevoConsultarEsfuerzo(repo).Ejecutar(context.Background(), c.proyecto, c.filtro)

		if !errors.Is(err, c.esperado) {
			t.Errorf("%s: error = %v, se esperaba %v", c.nombre, err, c.esperado)
		}
		if len(resultado.Registros) != 0 {
			t.Errorf("%s: con error no se devuelve consulta, llegaron %d registros", c.nombre, len(resultado.Registros))
		}
	}
}

// US-020: un proyecto inexistente no llega a leer los registros.
func TestConsultarEsfuerzo_ProyectoInexistenteNoLee(t *testing.T) {
	repo := &lecturaDeEsfuerzoFalsa{proyectos: map[string]bool{}}

	_, err := app.NuevoConsultarEsfuerzo(repo).Ejecutar(context.Background(), "p1", app.FiltroEsfuerzo{})

	if !errors.Is(err, app.ErrProyectoInexistente) || repo.lecturas != 0 {
		t.Errorf("error = %v, lecturas = %d; se esperaba ErrProyectoInexistente y ninguna lectura", err, repo.lecturas)
	}
}

// US-020: los fallos de la persistencia se devuelven envueltos, sin traducir.
func TestConsultarEsfuerzo_ErrorDeLaPersistencia(t *testing.T) {
	falla := errors.New("conexion perdida")

	repo := &lecturaDeEsfuerzoFalsa{proyectos: map[string]bool{"p1": true}, err: falla}
	_, err := app.NuevoConsultarEsfuerzo(repo).Ejecutar(context.Background(), "p1", app.FiltroEsfuerzo{})

	if !errors.Is(err, falla) {
		t.Errorf("error = %v, se esperaba que envolviera %v", err, falla)
	}
}

// US-020 / RN-020-12: consultar no modifica la lectura que recibe.
func TestConsultarEsfuerzo_NoModificaLaLectura(t *testing.T) {
	lectura := lecturaBase()
	antes := append([]esfuerzo.RegistroEsfuerzo(nil), lectura.Registros...)

	consultarEsf(t, lectura, app.FiltroEsfuerzo{})

	if !reflect.DeepEqual(lectura.Registros, antes) {
		t.Error("la consulta reordeno los registros de la lectura")
	}
}
