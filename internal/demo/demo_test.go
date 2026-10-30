package demo_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/demo"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// entorno es la carga armada sobre la base en memoria: todos los puertos los resuelve la misma Base, y los casos que reciben reloj
// reciben el mismo que la carga fija en cada paso.
type entorno struct {
	base     *memoria.Base
	reloj    *demo.Reloj
	casos    demo.Casos
	lecturas demo.Lecturas
}

func nuevoEntorno() entorno {
	b, reloj := memoria.NuevaBase(), demo.NuevoReloj()
	return entorno{
		base:  b,
		reloj: reloj,
		casos: demo.Casos{
			ListarProyectos: app.NuevoListarProyectos(b),
			CrearProyecto:   app.NuevoCrearProyecto(b),
			RegistrarInteg:  app.NuevoRegistrarIntegrante(b),
			CrearItem:       app.NuevoCrearItemBacklog(b, reloj.Ahora),
			EstimarItem:     app.NuevoEstimarItem(b),
			CrearSprint:     app.NuevoCrearSprint(b),
			AsignarItem:     app.NuevoAsignarItemASprint(b),
			MarcarCompleta:  app.NuevoMarcarCompletado(b, reloj.Ahora),
			CerrarSprint:    app.NuevoCerrarSprint(b),
			IniciarSesion:   app.NuevoIniciarSesionPlanningPoker(b),
			Votar:           app.NuevoVotarEnSesion(b),
			Revelar:         app.NuevoRevelarVotos(b),
			NuevaRonda:      app.NuevoAbrirNuevaRonda(b),
			Acordar:         app.NuevoRegistrarEstimacionAcordada(b),
			RegistrarHoras:  app.NuevoRegistrarEsfuerzo(b, reloj.Ahora),
			RegistrarDef:    app.NuevoRegistrarDefecto(b),
			CambiarEstado:   app.NuevoCambiarEstadoDefecto(b),
		},
		lecturas: demo.Lecturas{
			ListarBacklog:     app.NuevoListarBacklog(b),
			ListarSprints:     app.NuevoListarSprints(b),
			ConsultarEsfuerzo: app.NuevoConsultarEsfuerzo(b),
			ListarDefectos:    app.NuevoListarDefectos(b),
			ConsultarSesion:   app.NuevoConsultarSesionPlanningPoker(b),
		},
	}
}

func (e entorno) cargar(t *testing.T) demo.Resultado {
	t.Helper()
	r, err := demo.Cargar(context.Background(), e.casos, e.reloj)
	if err != nil {
		t.Fatalf("Cargar: %v", err)
	}
	return r
}

const resumenEsperado = `Proyecto "Tienda online" cargado: 4 integrantes, 16 historias, 4 sprints (3 cerrados y 1 abierto), 3 sesiones de Planning Poker, 27 registros de horas y 10 defectos.`

// CA-T023-1: con una base vacia se carga el proyecto completo, con las cantidades de la spec y el resumen que imprime el comando.
func TestCargar_ProyectoCompleto(t *testing.T) {
	e := nuevoEntorno()
	r := e.cargar(t)

	if got := r.Resumen.String(); got != resumenEsperado {
		t.Errorf("resumen:\n got  %s\n want %s", got, resumenEsperado)
	}
	if r.ProyectoID == "" || len(r.Sesiones) != 3 {
		t.Errorf("resultado = %+v, se esperaba el proyecto y las 3 sesiones", r)
	}
}

// CA-T023-2: las lecturas de la aplicacion sobre lo cargado dan las cifras de la spec (Verificar las compara una por una).
func TestCargar_LasLecturasDanLasCifrasDeLaSpec(t *testing.T) {
	e := nuevoEntorno()
	r := e.cargar(t)

	if err := demo.Verificar(context.Background(), e.lecturas, r); err != nil {
		t.Fatalf("Verificar: %v", err)
	}
}

// La comprobacion no da por buena una carga que no coincide: si falta una historia, lo dice.
func TestVerificar_DetectaLoQueNoCoincide(t *testing.T) {
	e := nuevoEntorno()
	r := e.cargar(t)
	r.Sesiones = map[int]string{} // como si faltaran las tres sesiones de Planning Poker

	err := demo.Verificar(context.Background(), e.lecturas, r)
	if err == nil {
		t.Fatal("Verificar no detecto que faltaban las sesiones")
	}
}

// CA-T023-3: ninguna seccion de los reportes del proyecto y del Sprint 4 esta vacia, y el reporte del proyecto no tiene errores.
func TestCargar_LosReportesNoTienenSeccionesVacias(t *testing.T) {
	e := nuevoEntorno()
	r := e.cargar(t)
	ctx, b := context.Background(), e.base

	rp, err := app.NuevoConsultarReporteDelProyecto(app.FuentesDelReporteDelProyecto{
		Proyecto: app.NuevoConsultarProyecto(b).Ejecutar,
		Backlog: func(ctx context.Context, proyectoID string) ([]backlog.ItemBacklog, error) {
			return app.NuevoListarBacklog(b).Ejecutar(ctx, proyectoID, app.FiltroBacklog{})
		},
		Sprints:     app.NuevoListarSprints(b).Ejecutar,
		Anteriores:  app.NuevoConsultarSprintsAnteriores(b).Ejecutar,
		Esfuerzo:    app.NuevoConsultarEsfuerzo(b).Ejecutar,
		Comparacion: app.NuevoCompararEsfuerzo(b).Ejecutar,
		Defectos:    app.NuevoListarDefectos(b).Ejecutar,
		Dashboard:   app.NuevoConsultarDashboard(b, func() time.Time { return ahoraDeLosReportes }).Ejecutar,
	}).Ejecutar(ctx, r.ProyectoID)
	if err != nil {
		t.Fatalf("reporte del proyecto: %v", err)
	}
	secciones := map[string]bool{
		"Integrantes":     rp.IntegrantesVacios(),
		"Product Backlog": rp.Backlog.Vacia(),
		"Sprints":         rp.Sprints.Vacia(),
		"Esfuerzo":        rp.Esfuerzo.Vacia(),
		"Defectos":        rp.Defectos.Vacia(),
		"Métricas":        rp.Metricas.Vacia(),
	}
	for titulo, vacia := range secciones {
		if vacia {
			t.Errorf("la seccion %q del reporte del proyecto esta vacia", titulo)
		}
	}
	for titulo, err := range map[string]error{
		"Proyecto": rp.Proyecto.Err, "Product Backlog": rp.Backlog.Err, "Sprints": rp.Sprints.Err,
		"Esfuerzo": rp.Esfuerzo.Err, "Defectos": rp.Defectos.Err, "Métricas": rp.Metricas.Err,
	} {
		if err != nil {
			t.Errorf("la seccion %q del reporte del proyecto informa un error: %v", titulo, err)
		}
	}

	sprints, err := app.NuevoListarSprints(b).Ejecutar(ctx, r.ProyectoID)
	if err != nil {
		t.Fatalf("sprints: %v", err)
	}
	var cuarto string
	for _, s := range sprints {
		if s.Sprint.Numero == 4 {
			cuarto = s.Sprint.ID
		}
	}
	rs, err := app.NuevoConsultarReporteDelSprint(app.FuentesDelReporteDelSprint{
		Proyecto: app.NuevoConsultarProyecto(b).Ejecutar,
		Sprint:   app.NuevoConsultarSprint(b).Ejecutar,
		Esfuerzo: app.NuevoConsultarEsfuerzo(b).Ejecutar,
		Defectos: app.NuevoListarDefectos(b).Ejecutar,
	}).Ejecutar(ctx, r.ProyectoID, cuarto)
	if err != nil {
		t.Fatalf("reporte del Sprint 4: %v", err)
	}
	if rs.Historias.Vacia() || rs.Esfuerzo.Vacia() || rs.Defectos.Vacia() {
		t.Errorf("el reporte del Sprint 4 tiene secciones vacias: historias=%t esfuerzo=%t defectos=%t",
			rs.Historias.Vacia(), rs.Esfuerzo.Vacia(), rs.Defectos.Vacia())
	}
}

var ahoraDeLosReportes = time.Date(2026, time.October, 30, 12, 0, 0, 0, time.UTC)

// CA-T023-4 / CL-T023-3: con el proyecto ya cargado, una segunda carga se niega sin escribir nada.
func TestCargar_SegundaVezSeNiegaSinEscribir(t *testing.T) {
	e := nuevoEntorno()
	e.cargar(t)
	antes := contar(t, e)

	_, err := demo.Cargar(context.Background(), e.casos, e.reloj)

	if !errors.Is(err, demo.ErrYaExiste) {
		t.Fatalf("err = %v, se esperaba ErrYaExiste", err)
	}
	if despues := contar(t, e); despues != antes {
		t.Errorf("la segunda carga escribio: antes %+v, despues %+v", antes, despues)
	}
}

// CL-T023-2: si la base ya tiene otro proyecto, "Tienda online" se carga al lado y el otro no cambia.
func TestCargar_ConOtroProyectoEnLaBase(t *testing.T) {
	e := nuevoEntorno()
	ctx := context.Background()
	otro, err := e.casos.CrearProyecto.Ejecutar(ctx, "Proyecto de prueba", "otro", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("CrearProyecto: %v", err)
	}

	r := e.cargar(t)

	if r.ProyectoID == otro.ID {
		t.Fatal("la carga uso el proyecto que ya existia")
	}
	proyectos, err := e.casos.ListarProyectos.Ejecutar(ctx)
	if err != nil || len(proyectos) != 2 {
		t.Fatalf("proyectos = %d (err %v), se esperaban 2", len(proyectos), err)
	}
	items, err := e.lecturas.ListarBacklog.Ejecutar(ctx, otro.ID, app.FiltroBacklog{})
	if err != nil || len(items) != 0 {
		t.Errorf("el otro proyecto tiene %d historias (err %v), se esperaban 0", len(items), err)
	}
	if err := demo.Verificar(ctx, e.lecturas, r); err != nil {
		t.Errorf("Verificar: %v", err)
	}
}

// RN-T023-6: si un paso falla, la carga se detiene, el error nombra el paso y lo ya cargado queda.
func TestCargar_UnPasoQueFallaNombraElPaso(t *testing.T) {
	e := nuevoEntorno()
	// Las estimaciones se piden sobre una base distinta, donde las historias no existen: el paso "estimar las historias" falla.
	e.casos.EstimarItem = app.NuevoEstimarItem(memoria.NuevaBase())

	r, err := demo.Cargar(context.Background(), e.casos, e.reloj)

	var paso *demo.PasoError
	if !errors.As(err, &paso) {
		t.Fatalf("err = %v, se esperaba un *PasoError", err)
	}
	if paso.Paso != "estimar las historias" {
		t.Errorf("paso = %q, se esperaba %q", paso.Paso, "estimar las historias")
	}
	if r.Resumen.Historias != 16 || r.Resumen.Sprints != 0 {
		t.Errorf("resumen = %+v, se esperaban las 16 historias ya cargadas y ningun sprint", r.Resumen)
	}
}

type conteo struct{ proyectos, historias, sprints, defectos, horas int }

func contar(t *testing.T, e entorno) conteo {
	t.Helper()
	ctx := context.Background()
	proyectos, err := e.casos.ListarProyectos.Ejecutar(ctx)
	if err != nil {
		t.Fatalf("ListarProyectos: %v", err)
	}
	c := conteo{proyectos: len(proyectos)}
	for _, p := range proyectos {
		items, _ := e.lecturas.ListarBacklog.Ejecutar(ctx, p.ID, app.FiltroBacklog{})
		sprints, _ := e.lecturas.ListarSprints.Ejecutar(ctx, p.ID)
		defectos, _ := e.lecturas.ListarDefectos.Ejecutar(ctx, p.ID, app.FiltroDefectos{})
		horas, _ := e.lecturas.ConsultarEsfuerzo.Ejecutar(ctx, p.ID, app.FiltroEsfuerzo{})
		c.historias += len(items)
		c.sprints += len(sprints)
		c.defectos += len(defectos)
		c.horas += len(horas.Registros)
	}
	return c
}
