// Package demo carga en una base vacia un proyecto de demostracion completo (spec T-023): backlog, estimaciones, Planning Poker, sprints
// cerrados y uno abierto, horas y defectos, para mostrar en la presentacion todas las funcionalidades con datos que se vean reales.
//
// La carga usa los casos de uso de la aplicacion, los mismos que usa la pantalla, y no inserta filas por su cuenta: asi los datos
// cumplen todas las reglas del dominio por construccion.
package demo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// ErrYaExiste se devuelve, sin haber escrito nada, si la base ya tiene el proyecto de demostracion (RN-T023-4).
var ErrYaExiste = errors.New("ya existe el proyecto " + `"` + NombreProyecto + `"`)

// PasoError es el error de un paso de la carga: dice cual fue y envuelve lo que rechazo el caso de uso (RN-T023-6).
type PasoError struct {
	Paso string
	Err  error
}

func (e *PasoError) Error() string { return fmt.Sprintf("falló el paso %q: %v", e.Paso, e.Err) }
func (e *PasoError) Unwrap() error { return e.Err }

// Reloj es el reloj de la carga: los casos de uso que reciben uno (CrearItemBacklog, MarcarCompletado y RegistrarEsfuerzo) reciben este,
// y la carga lo fija antes de cada paso en una fecha del sprint que esta armando (RN-T023-2).
type Reloj struct{ ahora time.Time }

// NuevoReloj arma un reloj en el primer dia del proyecto.
func NuevoReloj() *Reloj { return &Reloj{ahora: mediodia(inicioProyecto)} }

// Ahora es lo que se le pasa a los casos de uso como su reloj.
func (r *Reloj) Ahora() time.Time { return r.ahora }

// Fijar mueve el reloj al mediodia del dia pedido: asi una fecha de ese dia (a medianoche) nunca es posterior a "hoy".
func (r *Reloj) Fijar(d time.Time) { r.ahora = mediodia(d) }

func mediodia(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, time.UTC)
}

// Casos son los casos de uso que usa la carga. Los arma quien la llama: cmd/demo con Postgres y la prueba con la base en memoria. Los
// tres que reciben reloj tienen que haberse armado con Reloj.Ahora.
type Casos struct {
	ListarProyectos *app.ListarProyectos
	CrearProyecto   *app.CrearProyecto
	RegistrarInteg  *app.RegistrarIntegrante
	CrearItem       *app.CrearItemBacklog
	EstimarItem     *app.EstimarItem
	CrearSprint     *app.CrearSprint
	AsignarItem     *app.AsignarItemASprint
	MarcarCompleta  *app.MarcarCompletado
	CerrarSprint    *app.CerrarSprint
	IniciarSesion   *app.IniciarSesionPlanningPoker
	Votar           *app.VotarEnSesion
	Revelar         *app.RevelarVotos
	NuevaRonda      *app.AbrirNuevaRonda
	Acordar         *app.RegistrarEstimacionAcordada
	RegistrarHoras  *app.RegistrarEsfuerzo
	RegistrarDef    *app.RegistrarDefecto
	CambiarEstado   *app.CambiarEstadoDefecto
}

// Resumen cuenta lo que se cargo. Es lo que imprime el comando.
type Resumen struct {
	Integrantes, Historias, Sprints, Cerrados, Sesiones, Horas, Defectos int
}

func (r Resumen) String() string {
	return fmt.Sprintf("Proyecto %q cargado: %d integrantes, %d historias, %d sprints (%d cerrados y %d abierto), %d sesiones de Planning Poker, %d registros de horas y %d defectos.",
		NombreProyecto, r.Integrantes, r.Historias, r.Sprints, r.Cerrados, r.Sprints-r.Cerrados, r.Sesiones, r.Horas, r.Defectos)
}

// Resultado es lo que devuelve la carga: el proyecto, la sesion de Planning Poker de cada historia que tiene una y el resumen.
type Resultado struct {
	ProyectoID string
	Sesiones   map[int]string
	Resumen    Resumen
}

// Existe dice si la base ya tiene el proyecto de demostracion, mirando solo su nombre exacto (RN-T023-4).
func Existe(ctx context.Context, c Casos) (bool, error) {
	proyectos, err := c.ListarProyectos.Ejecutar(ctx)
	if err != nil {
		return false, fmt.Errorf("listar los proyectos: %w", err)
	}
	for _, p := range proyectos {
		if p.Nombre == NombreProyecto {
			return true, nil
		}
	}
	return false, nil
}

// Cargar arma el proyecto de demostracion siguiendo el orden de RN-T023-3. Si ya existe, devuelve ErrYaExiste sin escribir nada. Si un
// paso falla, se detiene y lo ya cargado queda (RN-T023-6).
func Cargar(ctx context.Context, c Casos, reloj *Reloj) (Resultado, error) {
	existe, err := Existe(ctx, c)
	if err != nil {
		return Resultado{}, &PasoError{Paso: "buscar el proyecto", Err: err}
	}
	if existe {
		return Resultado{}, ErrYaExiste
	}

	g := &carga{
		ctx: ctx, c: c, reloj: reloj,
		items:      map[int]backlog.ItemBacklog{},
		sprints:    map[int]sprint.Sprint{},
		defectoIDs: map[int]string{},
		sesiones:   map[int]string{},
	}
	pasos := []struct {
		nombre string
		hacer  func() error
	}{
		{"crear el proyecto", g.proyecto},
		{"registrar los integrantes", g.integrantes},
		{"crear las historias", g.historias},
		{"estimar las historias", g.estimaciones},
		{"Planning Poker de «Ver el historial de compras»", g.pokerHistorial},
		{"Planning Poker de «Cupones de descuento»", g.pokerCupones},
		{"Planning Poker de «Recomendar productos similares»", g.pokerRecomendada},
	}
	for _, p := range pasos {
		if err := p.hacer(); err != nil {
			return g.resultado(), &PasoError{Paso: p.nombre, Err: err}
		}
	}
	for n := range sprints {
		if err := g.sprint(n + 1); err != nil {
			return g.resultado(), err
		}
	}
	return g.resultado(), nil
}

// carga guarda lo que van necesitando los pasos: los ID que devolvio cada caso de uso.
type carga struct {
	ctx   context.Context
	c     Casos
	reloj *Reloj

	proyectoID string
	equipo     []integrante.Integrante
	items      map[int]backlog.ItemBacklog // por numero de historia (1 a 16)
	sprints    map[int]sprint.Sprint       // por numero de sprint (1 a 4), se completa al crearlos
	defectoIDs map[int]string              // por posicion en la lista de defectos
	sesiones   map[int]string              // sesion de Planning Poker por numero de historia

	resumen Resumen
}

func (g *carga) resultado() Resultado {
	return Resultado{ProyectoID: g.proyectoID, Sesiones: g.sesiones, Resumen: g.resumen}
}

func (g *carga) proyecto() error {
	p, err := g.c.CrearProyecto.Ejecutar(g.ctx, NombreProyecto, descripcionProyecto, inicioProyecto, finProyecto)
	if err != nil {
		return err
	}
	g.proyectoID = p.ID
	return nil
}

func (g *carga) integrantes() error {
	for _, d := range integrantes {
		i, err := g.c.RegistrarInteg.Ejecutar(g.ctx, g.proyectoID, d.nombre, d.email, d.rol)
		if err != nil {
			return fmt.Errorf("%s: %w", d.nombre, err)
		}
		g.equipo = append(g.equipo, i)
		g.resumen.Integrantes++
	}
	return nil
}

func (g *carga) historias() error {
	g.reloj.Fijar(inicioProyecto)
	for i, h := range historias {
		item, err := g.c.CrearItem.Ejecutar(g.ctx, g.proyectoID, backlog.DatosItem{
			Titulo: h.titulo, Descripcion: h.descripcion, Prioridad: h.prioridad, Criterios: h.criterios,
		})
		if err != nil {
			return fmt.Errorf("historia %d (%s): %w", i+1, h.titulo, err)
		}
		g.items[i+1] = item
		g.resumen.Historias++
	}
	return nil
}

func (g *carga) estimaciones() error {
	for i, h := range historias {
		if h.puntos == 0 {
			continue
		}
		if _, err := g.c.EstimarItem.Ejecutar(g.ctx, g.items[i+1].ID, h.puntos); err != nil {
			return fmt.Errorf("historia %d (%s): %w", i+1, h.titulo, err)
		}
	}
	return nil
}

// iniciar abre la sesion de Planning Poker de una historia con los cuatro integrantes.
func (g *carga) iniciar(historia int) (string, error) {
	s, err := g.c.IniciarSesion.Ejecutar(g.ctx, g.proyectoID, g.items[historia].ID)
	if err != nil {
		return "", err
	}
	g.sesiones[historia] = s.ID
	g.resumen.Sesiones++
	return s.ID, nil
}

// votar registra el voto de cada integrante, en el orden de la lista.
func (g *carga) votar(sesionID string, valores []int) error {
	for i, valor := range valores {
		if _, err := g.c.Votar.Ejecutar(g.ctx, sesionID, g.equipo[i].ID, valor); err != nil {
			return fmt.Errorf("voto de %s: %w", g.equipo[i].Nombre, err)
		}
	}
	return nil
}

// pokerCupones deja la sesion revelada y con divergencia, sin abrir una nueva ronda.
func (g *carga) pokerCupones() error {
	id, err := g.iniciar(historiaCupones)
	if err != nil {
		return err
	}
	if err := g.votar(id, votosCupones); err != nil {
		return err
	}
	rev, err := g.c.Revelar.Ejecutar(g.ctx, id, g.equipo[ana].ID)
	if err != nil {
		return fmt.Errorf("revelar: %w", err)
	}
	if !rev.Divergencia {
		return errors.New("se esperaba que los votos tuvieran divergencia")
	}
	return nil
}

// pokerRecomendada deja la sesion votando, con solo los votos de Ana y Carla: los de Bruno y Diego quedan sin votar.
func (g *carga) pokerRecomendada() error {
	id, err := g.iniciar(historiaRecomendada)
	if err != nil {
		return err
	}
	for _, v := range votosRecomendada {
		if _, err := g.c.Votar.Ejecutar(g.ctx, id, g.equipo[v.quien].ID, v.valor); err != nil {
			return fmt.Errorf("voto de %s: %w", g.equipo[v.quien].Nombre, err)
		}
	}
	return nil
}

// pokerHistorial es el ciclo completo de la historia 9: votos con divergencia, nueva ronda, acuerdo y estimacion. Tiene que estar hecho
// antes de asignarla al Sprint 4, porque asignar exige que este estimada (RN-009-7).
func (g *carga) pokerHistorial() error {
	id, err := g.iniciar(historiaHistorial)
	if err != nil {
		return err
	}
	if err := g.votar(id, votosHistorialRonda1); err != nil {
		return err
	}
	rev, err := g.c.Revelar.Ejecutar(g.ctx, id, g.equipo[ana].ID)
	if err != nil {
		return fmt.Errorf("revelar la primera ronda: %w", err)
	}
	if !rev.Divergencia {
		return errors.New("se esperaba divergencia en la primera ronda")
	}
	if _, err := g.c.NuevaRonda.Ejecutar(g.ctx, id, g.equipo[ana].ID); err != nil {
		return fmt.Errorf("abrir la segunda ronda: %w", err)
	}
	if err := g.votar(id, votosHistorialRonda2); err != nil {
		return err
	}
	if _, err := g.c.Revelar.Ejecutar(g.ctx, id, g.equipo[ana].ID); err != nil {
		return fmt.Errorf("revelar la segunda ronda: %w", err)
	}
	if _, err := g.c.Acordar.Ejecutar(g.ctx, id, g.equipo[ana].ID, estimacionAcordadaHistorial); err != nil {
		return fmt.Errorf("registrar la estimacion acordada: %w", err)
	}
	return nil
}

// sprint arma el sprint n: lo crea, le asigna sus historias, registra sus horas, completa las que se completan, registra sus defectos
// (y pone el estado final a los que corresponde) y, si tiene que quedar cerrado, lo cierra. Un sprint se crea recien cuando el anterior
// esta cerrado (RN-T023-3).
func (g *carga) sprint(n int) error {
	paso := func(nombre string, err error) error {
		if err == nil {
			return nil
		}
		return &PasoError{Paso: fmt.Sprintf("Sprint %d: %s", n, nombre), Err: err}
	}
	d := sprints[n-1]

	g.reloj.Fijar(d.inicio)
	s, err := g.c.CrearSprint.Ejecutar(g.ctx, g.proyectoID, d.objetivo, d.inicio, d.fin)
	if err := paso("crear el sprint", err); err != nil {
		return err
	}
	g.sprints[n] = s
	g.resumen.Sprints++

	for _, h := range d.historias {
		if _, err := g.c.AsignarItem.Ejecutar(g.ctx, g.items[h].ID, s.ID); err != nil {
			return paso(fmt.Sprintf("asignar la historia %d", h), err)
		}
	}
	for _, r := range horas {
		if r.sprint != n {
			continue
		}
		g.reloj.Fijar(r.fecha)
		if _, err := g.c.RegistrarHoras.Ejecutar(g.ctx, g.proyectoID, g.items[r.historia].ID, g.equipo[r.quien].ID, r.fecha, r.actividad, r.horas); err != nil {
			return paso(fmt.Sprintf("registrar %.2f h de %s en la historia %d", r.horas, r.actividad, r.historia), err)
		}
		g.resumen.Horas++
	}
	for _, c := range d.completadas {
		g.reloj.Fijar(c.fecha)
		if _, err := g.c.MarcarCompleta.Ejecutar(g.ctx, g.items[c.historia].ID, s.ID); err != nil {
			return paso(fmt.Sprintf("completar la historia %d", c.historia), err)
		}
	}
	if err := paso("defectos", g.defectos(n)); err != nil {
		return err
	}

	if d.cerrado {
		if _, err := g.c.CerrarSprint.Ejecutar(g.ctx, s.ID); err != nil {
			return paso("cerrar el sprint", err)
		}
		g.resumen.Cerrados++
	}
	return nil
}

// defectos registra los defectos que se detectan en el sprint n y aplica el estado final de los que quedan listos en este sprint: los
// resueltos recien cuando existe el sprint donde se resuelven.
func (g *carga) defectos(n int) error {
	for i, d := range defectos {
		if d.sprintDeteccion == n {
			creado, err := g.c.RegistrarDef.Ejecutar(g.ctx, g.proyectoID, g.items[d.historia].ID, g.sprints[n].ID, d.descripcion, d.severidad)
			if err != nil {
				return fmt.Errorf("registrar %q: %w", d.descripcion, err)
			}
			g.defectoIDs[i] = creado.ID
			g.resumen.Defectos++
		}
	}
	for i, d := range defectos {
		if d.fase() != n || d.estado == defecto.EstadoAbierto {
			continue
		}
		resolucion := ""
		if d.estado == defecto.EstadoResuelto {
			resolucion = g.sprints[d.sprintResolucion].ID
		}
		if _, err := g.c.CambiarEstado.Ejecutar(g.ctx, g.proyectoID, g.defectoIDs[i], d.estado, resolucion); err != nil {
			return fmt.Errorf("pasar %q a %s: %w", d.descripcion, d.estado, err)
		}
	}
	return nil
}
