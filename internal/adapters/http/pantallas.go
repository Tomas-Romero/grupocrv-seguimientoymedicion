package httpadapter

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/errores"
	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// CasosDeUso son los casos de uso que usan las pantallas de proyectos, integrantes y
// backlog (spec T-011). Un handler solo lee la peticion, llama a uno de estos y escribe la
// respuesta: no valida reglas de negocio ni habla con un repositorio (RN-T011-1).
type CasosDeUso struct {
	ListarProyectos     *app.ListarProyectos
	ConsultarProyecto   *app.ConsultarProyecto
	CrearProyecto       *app.CrearProyecto
	ModificarProyecto   *app.ModificarProyecto
	RegistrarIntegrante *app.RegistrarIntegrante
	ListarBacklog       *app.ListarBacklog
	CrearItem           *app.CrearItemBacklog
	ConsultarItem       *app.ConsultarItem
	EditarItem          *app.EditarItem
	EstimarItem         *app.EstimarItem

	// Pantallas de sprints (T-015).
	ListarSprints       *app.ListarSprints
	ConsultarSprint     *app.ConsultarSprint
	CrearSprint         *app.CrearSprint
	AsignarItem         *app.AsignarItemASprint
	MarcarCompletado    *app.MarcarCompletado
	DesmarcarCompletado *app.DesmarcarCompletado

	// Cierre de un sprint (T-016).
	CerrarSprint *app.CerrarSprint

	// Historial de sprints cerrados (US-012).
	ConsultarSprintsAnteriores *app.ConsultarSprintsAnteriores

	// Pantallas de esfuerzo (T-019).
	RegistrarEsfuerzo *app.RegistrarEsfuerzo
	ConsultarEsfuerzo *app.ConsultarEsfuerzo
	CompararEsfuerzo  *app.CompararEsfuerzo
	// Ahora es el reloj de las pantallas: el valor inicial de la fecha de las horas. Si falta, time.Now.
	Ahora func() time.Time
	// Planning Poker (T-020).
	IniciarSesionPlanningPoker *app.IniciarSesionPlanningPoker
	ConsultarSesion            *app.ConsultarSesionPlanningPoker
	VotarEnSesion              *app.VotarEnSesion
	RevelarVotos               *app.RevelarVotos
	AbrirNuevaRonda            *app.AbrirNuevaRonda
	RegistrarAcuerdo           *app.RegistrarEstimacionAcordada
	// Pantallas de defectos (T-021).
	ListarDefectos   *app.ListarDefectos
	RegistrarDefecto *app.RegistrarDefecto
}

// Pantallas son los handlers de las pantallas de proyectos, integrantes y backlog.
type Pantallas struct {
	casos   CasosDeUso
	errores *ManejadorDeErrores
}

// NuevasPantallas arma los handlers. Los errores se responden con el manejador de T-005.
func NuevasPantallas(casos CasosDeUso, manejador *ManejadorDeErrores) *Pantallas {
	return &Pantallas{casos: casos, errores: manejador}
}

// registrar agrega las rutas de las pantallas al mux, con el esquema de URL de T-004.
func (p *Pantallas) registrar(mux *http.ServeMux) {
	mux.HandleFunc("GET /proyectos", p.listaDeProyectos)
	mux.HandleFunc("GET /proyectos/nuevo", p.nuevoProyecto)
	mux.HandleFunc("POST /proyectos", p.crearProyecto)
	mux.HandleFunc("GET /proyectos/{id}", p.verProyecto)
	mux.HandleFunc("GET /proyectos/{id}/editar", p.editarProyecto)
	mux.HandleFunc("POST /proyectos/{id}", p.actualizarProyecto)
	mux.HandleFunc("POST /proyectos/{id}/integrantes", p.registrarIntegrante)
	mux.HandleFunc("GET /proyectos/{id}/backlog", p.verBacklog)
	mux.HandleFunc("POST /proyectos/{id}/backlog", p.crearHistoria)
	mux.HandleFunc("GET /proyectos/{id}/backlog/{item}/editar", p.editarHistoria)
	mux.HandleFunc("POST /proyectos/{id}/backlog/{item}", p.actualizarHistoria)
	mux.HandleFunc("POST /proyectos/{id}/backlog/{item}/estimacion", p.estimarHistoria)
	mux.HandleFunc("POST /proyectos/{id}/backlog/{item}/planning-poker", p.iniciarSesion)
	mux.HandleFunc("GET /proyectos/{id}/planning-poker/{sesion}", p.verSesion)
	mux.HandleFunc("POST /proyectos/{id}/planning-poker/{sesion}/votos", p.votarEnSesion)
	mux.HandleFunc("POST /proyectos/{id}/planning-poker/{sesion}/revelar", p.revelarSesion)
	mux.HandleFunc("POST /proyectos/{id}/planning-poker/{sesion}/nueva-ronda", p.abrirNuevaRonda)
	mux.HandleFunc("POST /proyectos/{id}/planning-poker/{sesion}/acuerdo", p.registrarAcuerdo)

	// Sprints (T-015).
	mux.HandleFunc("GET /proyectos/{id}/sprints", p.listaDeSprints)
	mux.HandleFunc("POST /proyectos/{id}/sprints", p.crearSprint)
	mux.HandleFunc("GET /proyectos/{id}/sprints/anteriores", p.sprintsAnteriores)
	mux.HandleFunc("GET /proyectos/{id}/sprints/{sprint}", p.verSprint)
	mux.HandleFunc("POST /proyectos/{id}/sprints/{sprint}/historias", p.asignarHistoria)
	mux.HandleFunc("POST /proyectos/{id}/sprints/{sprint}/historias/{item}/completar", p.completarHistoria)
	mux.HandleFunc("POST /proyectos/{id}/sprints/{sprint}/historias/{item}/reabrir", p.reabrirHistoria)
	mux.HandleFunc("POST /proyectos/{id}/sprints/{sprint}/cerrar", p.cerrarSprint)

	// Esfuerzo (T-019).
	mux.HandleFunc("GET /proyectos/{id}/esfuerzo", p.verEsfuerzo)
	mux.HandleFunc("POST /proyectos/{id}/esfuerzo", p.registrarEsfuerzo)
	mux.HandleFunc("GET /proyectos/{id}/esfuerzo/comparacion", p.compararEsfuerzo)
	// Defectos (T-021).
	mux.HandleFunc("GET /proyectos/{id}/defectos", p.listaDeDefectos)
	mux.HandleFunc("POST /proyectos/{id}/defectos", p.registrarDefecto)
}

// ---------------------------------------------------------------- proyectos

func (p *Pantallas) listaDeProyectos(w http.ResponseWriter, r *http.Request) {
	proyectos, err := p.casos.ListarProyectos.Ejecutar(r.Context())
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	filas := make([]plantillas.FilaProyecto, len(proyectos))
	for i, pr := range proyectos {
		filas[i] = plantillas.FilaProyecto{ID: pr.ID, Nombre: pr.Nombre, Inicio: fechaVisible(pr.FechaInicio), Fin: fechaVisible(pr.FechaFin)}
	}
	Responder(w, r, http.StatusOK, plantillas.Pagina{Titulo: "Proyectos", Activo: plantillas.SeccionProyectos},
		plantillas.ListaDeProyectos(plantillas.VistaProyectos{Proyectos: filas}))
}

func (p *Pantallas) nuevoProyecto(w http.ResponseWriter, r *http.Request) {
	Responder(w, r, http.StatusOK, plantillas.Pagina{Titulo: "Nuevo proyecto", Activo: plantillas.SeccionProyectos},
		plantillas.FormularioDeProyecto(plantillas.FormProyecto{Encabezado: "Nuevo proyecto", Accion: "/proyectos", Volver: "/proyectos"}))
}

func (p *Pantallas) crearProyecto(w http.ResponseWriter, r *http.Request) {
	f, err := leerFormulario(r)
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	inicio, fin := f.Fecha("fecha_inicio"), f.Fecha("fecha_fin")
	if err := f.Err(); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	creado, err := p.casos.CrearProyecto.Ejecutar(r.Context(), r.PostForm.Get("nombre"), r.PostForm.Get("descripcion"), inicio, fin)
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeProyecto(creado.ID))
}

func (p *Pantallas) verProyecto(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detalle, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	integrantes := make([]plantillas.FilaIntegrante, len(detalle.Integrantes))
	for i, in := range detalle.Integrantes {
		integrantes[i] = plantillas.FilaIntegrante{Nombre: in.Nombre, Email: in.Email, Rol: textoDeRol(in.Rol)}
	}
	vista := plantillas.VistaProyecto{
		ID:             id,
		Nombre:         detalle.Proyecto.Nombre,
		Descripcion:    detalle.Proyecto.Descripcion,
		Inicio:         fechaVisible(detalle.Proyecto.FechaInicio),
		Fin:            fechaVisible(detalle.Proyecto.FechaFin),
		Editar:         rutaDeProyecto(id) + "/editar",
		Backlog:        rutaDeProyecto(id) + "/backlog",
		Sprints:        rutaDeProyecto(id) + "/sprints",
		Integrantes:    integrantes,
		AltaIntegrante: rutaDeProyecto(id) + "/integrantes",
		Roles:          opcionesDeRol(integrante.RolProductBuilder),
	}
	Responder(w, r, http.StatusOK, paginaDeProyecto(detalle.Proyecto, plantillas.SeccionResumen), plantillas.ResumenDeProyecto(vista))
}

func (p *Pantallas) editarProyecto(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detalle, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	pr := detalle.Proyecto
	formulario := plantillas.FormProyecto{
		Encabezado:  "Editar proyecto",
		Accion:      rutaDeProyecto(id),
		Volver:      rutaDeProyecto(id),
		Nombre:      pr.Nombre,
		Descripcion: pr.Descripcion,
		FechaInicio: fechaDeFormulario(pr.FechaInicio),
		FechaFin:    fechaDeFormulario(pr.FechaFin),
	}
	Responder(w, r, http.StatusOK, paginaDeProyecto(pr, plantillas.SeccionResumen), plantillas.FormularioDeProyecto(formulario))
}

func (p *Pantallas) actualizarProyecto(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	f, err := leerFormulario(r)
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	inicio, fin := f.Fecha("fecha_inicio"), f.Fecha("fecha_fin")
	if err := f.Err(); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	if _, err := p.casos.ModificarProyecto.Ejecutar(r.Context(), id, r.PostForm.Get("nombre"), r.PostForm.Get("descripcion"), inicio, fin); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeProyecto(id))
}

func (p *Pantallas) registrarIntegrante(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := leerFormulario(r); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	_, err := p.casos.RegistrarIntegrante.Ejecutar(r.Context(), id, r.PostForm.Get("nombre"), r.PostForm.Get("email"), integrante.Rol(r.PostForm.Get("rol")))
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeProyecto(id))
}

// ------------------------------------------------------------------ backlog

func (p *Pantallas) verBacklog(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detalle, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	items, err := p.casos.ListarBacklog.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	sesiones, err := p.sesionesDelBacklog(r, id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	filas := make([]plantillas.FilaItem, len(items))
	for i, item := range items {
		puntos, estimado := item.StoryPoints.Valor()
		filas[i] = plantillas.FilaItem{
			Numero:    item.Numero,
			Titulo:    item.Titulo,
			Prioridad: textoDePrioridad(item.Prioridad),
			Estado:    textoDeEstado(item.Estado),
			Puntos:    textoDePuntos(item),
			Criterios: len(item.Criterios),
			Editar:    rutaDeHistoria(id, item.ID) + "/editar",
			Estimar:   rutaDeHistoria(id, item.ID) + "/estimacion",
			Escala:    opcionesDeEscala(puntos, estimado),
		}
		// T-020: una historia completada no ofrece Planning Poker (CL-T020-1); con una sesion
		// abierta, lleva a ella (CL-T020-2).
		if item.Estado != backlog.EstadoCompletado {
			filas[i].SesionAbierta = sesiones[item.ID]
			if filas[i].SesionAbierta == "" {
				filas[i].IniciarPoker = rutaDeHistoria(id, item.ID) + "/planning-poker"
			}
		}
	}
	vista := plantillas.VistaBacklog{
		Items: filas,
		Alta: plantillas.FormItem{
			Encabezado:  "Nueva historia",
			Accion:      rutaDeProyecto(id) + "/backlog",
			Prioridades: opcionesDePrioridad(backlog.PrioridadMust),
		},
	}
	Responder(w, r, http.StatusOK, paginaDeProyecto(detalle.Proyecto, plantillas.SeccionBacklog), plantillas.BacklogDeProyecto(vista))
}

func (p *Pantallas) crearHistoria(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := leerFormulario(r); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	if _, err := p.casos.CrearItem.Ejecutar(r.Context(), id, datosDeHistoria(r)); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeProyecto(id)+"/backlog")
}

func (p *Pantallas) editarHistoria(w http.ResponseWriter, r *http.Request) {
	id, itemID := r.PathValue("id"), r.PathValue("item")
	item, err := p.casos.ConsultarItem.Ejecutar(r.Context(), id, itemID)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	detalle, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	formulario := plantillas.FormItem{
		Encabezado:  "Editar historia",
		Accion:      rutaDeHistoria(id, itemID),
		Volver:      rutaDeProyecto(id) + "/backlog",
		Titulo:      item.Titulo,
		Descripcion: item.Descripcion,
		Criterios:   strings.Join(item.Criterios, "\n"),
		Prioridades: opcionesDePrioridad(item.Prioridad),
	}
	Responder(w, r, http.StatusOK, paginaDeProyecto(detalle.Proyecto, plantillas.SeccionBacklog), plantillas.FormularioDeItem(formulario))
}

func (p *Pantallas) actualizarHistoria(w http.ResponseWriter, r *http.Request) {
	id, itemID := r.PathValue("id"), r.PathValue("item")
	if _, err := leerFormulario(r); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	// RN-T011-11: la historia se alcanza solo por la ruta de su proyecto.
	if _, err := p.casos.ConsultarItem.Ejecutar(r.Context(), id, itemID); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	if _, err := p.casos.EditarItem.Ejecutar(r.Context(), itemID, datosDeHistoria(r)); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeProyecto(id)+"/backlog")
}

func (p *Pantallas) estimarHistoria(w http.ResponseWriter, r *http.Request) {
	id, itemID := r.PathValue("id"), r.PathValue("item")
	f, err := leerFormulario(r)
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	puntos := f.Entero("estimacion")
	if _, err := p.casos.ConsultarItem.Ejecutar(r.Context(), id, itemID); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	if err := f.Err(); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	if _, err := p.casos.EstimarItem.Ejecutar(r.Context(), itemID, puntos); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeProyecto(id)+"/backlog")
}

// ------------------------------------------------------------------ ayudas

// fallaDeLectura responde el error de un GET: lo que no existe (el proyecto, una historia, un sprint, un integrante
// o una sesion) es la pagina 404 con layout; lo demas, el error uniforme de T-005 (RN-T011-10).
func (p *Pantallas) fallaDeLectura(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, app.ErrProyectoInexistente) || errors.Is(err, app.ErrItemInexistente) || errors.Is(err, app.ErrSprintInexistente) ||
		errors.Is(err, app.ErrIntegranteInexistente) || errors.Is(err, app.ErrSesionInexistente) {
		noEncontrado(w, r)
		return
	}
	p.errores.ResponderError(w, r, err)
}

// leerFormulario lee el cuerpo del POST. Si no se puede leer, es un error de validacion
// general y no una falla del servidor.
func leerFormulario(r *http.Request) (*Formulario, error) {
	if err := r.ParseForm(); err != nil {
		return nil, &errores.ErrorDeValidacion{Errores: []errores.ErrorDeCampo{{Mensaje: "no se pudo leer el formulario"}}}
	}
	return NuevoFormulario(r.PostForm), nil
}

// redirigir lleva a la persona a la pagina de destino despues de guardar: con HTMX,
// con la cabecera HX-Redirect y el cuerpo vacio; sin HTMX, con un 303 (RN-T011-6).
func redirigir(w http.ResponseWriter, r *http.Request, destino string) {
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", destino)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, destino, http.StatusSeeOther)
}

func rutaDeProyecto(id string) string { return "/proyectos/" + id }

func rutaDeHistoria(proyectoID, itemID string) string {
	return rutaDeProyecto(proyectoID) + "/backlog/" + itemID
}

// paginaDeProyecto arma los datos del layout de una pagina dentro de un proyecto, con su menu.
func paginaDeProyecto(pr proyecto.Proyecto, activo plantillas.Seccion) plantillas.Pagina {
	return plantillas.Pagina{
		Titulo:   pr.Nombre,
		Activo:   activo,
		Proyecto: &plantillas.ProyectoActual{ID: pr.ID, Nombre: pr.Nombre},
	}
}

// datosDeHistoria lee los campos de una historia del formulario. Los criterios llegan uno
// por linea; las lineas vacias se ignoran (RN-T011-7).
func datosDeHistoria(r *http.Request) backlog.DatosItem {
	return backlog.DatosItem{
		Titulo:      r.PostForm.Get("titulo"),
		Descripcion: r.PostForm.Get("descripcion"),
		Prioridad:   backlog.Prioridad(r.PostForm.Get("prioridad")),
		Criterios:   criteriosDeTexto(r.PostForm.Get("criterios")),
	}
}

func criteriosDeTexto(texto string) []string {
	var criterios []string
	for _, linea := range strings.Split(strings.ReplaceAll(texto, "\r\n", "\n"), "\n") {
		if linea = strings.TrimSpace(linea); linea != "" {
			criterios = append(criterios, linea)
		}
	}
	return criterios
}

// fechaVisible es la fecha como se lee (dd/mm/aaaa); fechaDeFormulario, como la escribe un
// <input type="date"> (aaaa-mm-dd).
func fechaVisible(t time.Time) string { return t.Format("02/01/2006") }

func fechaDeFormulario(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.DateOnly)
}

func textoDePrioridad(p backlog.Prioridad) string {
	switch p {
	case backlog.PrioridadMust:
		return "Must"
	case backlog.PrioridadShould:
		return "Should"
	case backlog.PrioridadCould:
		return "Could"
	case backlog.PrioridadWont:
		return "Won't"
	}
	return string(p)
}

func textoDeEstado(e backlog.Estado) string {
	switch e {
	case backlog.EstadoPendiente:
		return "Pendiente"
	case backlog.EstadoEnProgreso:
		return "En progreso"
	case backlog.EstadoCompletado:
		return "Completado"
	}
	return string(e)
}

func textoDeRol(r integrante.Rol) string {
	switch r {
	case integrante.RolProductArchitect:
		return "Product Architect"
	case integrante.RolAgileEnabler:
		return "Agile Enabler"
	case integrante.RolProductBuilder:
		return "Product Builder"
	}
	return string(r)
}

func opcionesDeRol(elegido integrante.Rol) []plantillas.Opcion {
	roles := []integrante.Rol{integrante.RolProductBuilder, integrante.RolAgileEnabler, integrante.RolProductArchitect}
	opciones := make([]plantillas.Opcion, len(roles))
	for i, r := range roles {
		opciones[i] = plantillas.Opcion{Valor: string(r), Texto: textoDeRol(r), Elegida: r == elegido}
	}
	return opciones
}

func opcionesDePrioridad(elegida backlog.Prioridad) []plantillas.Opcion {
	prioridades := []backlog.Prioridad{backlog.PrioridadMust, backlog.PrioridadShould, backlog.PrioridadCould, backlog.PrioridadWont}
	opciones := make([]plantillas.Opcion, len(prioridades))
	for i, pr := range prioridades {
		opciones[i] = plantillas.Opcion{Valor: string(pr), Texto: textoDePrioridad(pr), Elegida: pr == elegida}
	}
	return opciones
}

// opcionesDeEscala son las cartas de la escala de Story Points, con la actual elegida. Si la
// historia no esta estimada, la primera opcion es un "Elegir" vacio.
func opcionesDeEscala(actual int, estimada bool) []plantillas.Opcion {
	var opciones []plantillas.Opcion
	if !estimada {
		opciones = append(opciones, plantillas.Opcion{Valor: "", Texto: "Elegir", Elegida: true})
	}
	for _, valor := range backlog.EscalaFibonacci() {
		opciones = append(opciones, plantillas.Opcion{Valor: strconv.Itoa(valor), Texto: strconv.Itoa(valor), Elegida: estimada && valor == actual})
	}
	return opciones
}
