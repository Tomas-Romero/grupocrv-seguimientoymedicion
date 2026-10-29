package httpadapter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/errores"
	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// Textos de los dos campos que la pantalla revisa antes de llamar al caso de uso (RN-T021-5). Son los
// unicos mensajes de esta pantalla que no salen del catalogo de errores.
const (
	mensajeHistoriaSinElegir = "Elegí la historia relacionada"
	mensajeSprintSinElegir   = "Elegí el sprint de detección"
)

func rutaDeDefectos(proyectoID string) string { return rutaDeProyecto(proyectoID) + "/defectos" }

// Handlers de las pantallas de defectos (spec T-021). Igual que los de T-011 y T-015, solo leen la
// peticion, llaman a un caso de uso y escriben la respuesta (RN-T021-1).

func textoDeSeveridad(s defecto.Severidad) string {
	switch s {
	case defecto.SeveridadCritica:
		return "Crítica"
	case defecto.SeveridadAlta:
		return "Alta"
	case defecto.SeveridadMedia:
		return "Media"
	case defecto.SeveridadBaja:
		return "Baja"
	}
	return string(s)
}

func textoDeEstadoDeDefecto(e defecto.Estado) string {
	switch e {
	case defecto.EstadoAbierto:
		return "Abierto"
	case defecto.EstadoEnCorreccion:
		return "En corrección"
	case defecto.EstadoResuelto:
		return "Resuelto"
	case defecto.EstadoDescartado:
		return "Descartado"
	}
	return string(e)
}

// textoDeSprintDeDefecto es "Sprint N", o "—" si el defecto no tiene sprint de resolucion (el
// numero 0 de DefectoListado, RN-T021-13).
func textoDeSprintDeDefecto(numero int) string {
	if numero == 0 {
		return "—"
	}
	return sprint.Sprint{Numero: numero}.Nombre()
}

// opcionesDeSeveridad son las cuatro severidades con su etiqueta, tras una opcion inicial sin valor
// (RN-T021-15): la persona tiene que elegir una, no hay una por defecto.
func opcionesDeSeveridad() []plantillas.Opcion {
	severidades := []defecto.Severidad{defecto.SeveridadCritica, defecto.SeveridadAlta, defecto.SeveridadMedia, defecto.SeveridadBaja}
	opciones := []plantillas.Opcion{{Valor: "", Texto: "Elegí una severidad"}}
	for _, s := range severidades {
		opciones = append(opciones, plantillas.Opcion{Valor: string(s), Texto: textoDeSeveridad(s)})
	}
	return opciones
}

// opcionesDeHistoriasDeDefecto son las historias del backlog como "#N titulo", en el orden en que las
// devuelve ListarBacklog (RN-T021-16), tras una opcion inicial sin valor.
func opcionesDeHistoriasDeDefecto(items []backlog.ItemBacklog) []plantillas.Opcion {
	opciones := []plantillas.Opcion{{Valor: "", Texto: "Elegí una historia"}}
	for _, item := range items {
		opciones = append(opciones, plantillas.Opcion{Valor: item.ID, Texto: fmt.Sprintf("#%d %s", item.Numero, item.Titulo)})
	}
	return opciones
}

// opcionesDeSprintsDeDefecto son todos los sprints del proyecto, abiertos y cerrados, por numero
// descendente como los devuelve ListarSprints (RN-T021-15, RN-T021-16), tras una opcion inicial sin
// valor. No se filtran por el sprint de deteccion: esa regla es del dominio (RN-023-4).
func opcionesDeSprintsDeDefecto(sprints []app.ResumenDeSprint) []plantillas.Opcion {
	opciones := []plantillas.Opcion{{Valor: "", Texto: "Elegí un sprint"}}
	for _, s := range sprints {
		opciones = append(opciones, plantillas.Opcion{Valor: s.Sprint.ID, Texto: s.Sprint.Nombre()})
	}
	return opciones
}

// largoDeNombreDeDefecto es cuantos caracteres de la descripcion se usan para nombrar los controles de
// una fila (RN-T021-17).
const largoDeNombreDeDefecto = 40

// nombreDeDefecto es el comienzo de la descripcion, con los espacios y saltos de linea reducidos a un
// espacio, para el nombre accesible de los controles de la fila. El defecto no tiene numero ni titulo.
func nombreDeDefecto(descripcion string) string {
	texto := strings.Join(strings.Fields(descripcion), " ")
	if letras := []rune(texto); len(letras) > largoDeNombreDeDefecto {
		texto = strings.TrimSpace(string(letras[:largoDeNombreDeDefecto]))
	}
	return texto
}

// accionesDeDefecto son los botones que ofrece una fila segun el estado que muestra: solo las
// transiciones validas de US-023 (RN-023-2, RN-T021-14). Desde en correccion, ir a abierto no es una
// reapertura y se llama "Volver a abierto"; desde resuelto o descartado si, y se llama "Reabrir". La
// pantalla no protege la regla: el caso de uso evalua el pedido contra el estado guardado.
func accionesDeDefecto(estado defecto.Estado) []plantillas.AccionDeDefecto {
	resolver := plantillas.AccionDeDefecto{Texto: "Resolver", Estado: string(defecto.EstadoResuelto), ConSprint: true}
	descartar := plantillas.AccionDeDefecto{Texto: "Descartar", Estado: string(defecto.EstadoDescartado)}
	switch estado {
	case defecto.EstadoAbierto:
		return []plantillas.AccionDeDefecto{{Texto: "En corrección", Estado: string(defecto.EstadoEnCorreccion)}, resolver, descartar}
	case defecto.EstadoEnCorreccion:
		return []plantillas.AccionDeDefecto{{Texto: "Volver a abierto", Estado: string(defecto.EstadoAbierto)}, resolver, descartar}
	case defecto.EstadoResuelto, defecto.EstadoDescartado:
		return []plantillas.AccionDeDefecto{{Texto: "Reabrir", Estado: string(defecto.EstadoAbierto)}}
	}
	return nil
}

// listaDeDefectos es la pantalla con la lista de defectos del proyecto y el formulario de alta
// (RN-T021-2). Las listas del formulario salen de lecturas que ya existen.
func (p *Pantallas) listaDeDefectos(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detalle, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	defectos, err := p.casos.ListarDefectos.Ejecutar(r.Context(), id, app.FiltroDefectos{})
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	historias, err := p.casos.ListarBacklog.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	sprints, err := p.casos.ListarSprints.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}

	sprintsDeResolucion := opcionesDeSprintsDeDefecto(sprints)
	filas := make([]plantillas.FilaDefecto, len(defectos))
	for i, d := range defectos {
		filas[i] = plantillas.FilaDefecto{
			Descripcion:      d.Defecto.Descripcion,
			Severidad:        textoDeSeveridad(d.Defecto.Severidad),
			Estado:           textoDeEstadoDeDefecto(d.Defecto.Estado),
			Historia:         fmt.Sprintf("#%d %s", d.HistoriaNumero, d.HistoriaTitulo),
			SprintDeteccion:  textoDeSprintDeDefecto(d.SprintDeteccion),
			SprintResolucion: textoDeSprintDeDefecto(d.SprintResolucion),
			Nombre:           nombreDeDefecto(d.Defecto.Descripcion),
			Accion:           rutaDeDefectos(id) + "/" + d.Defecto.ID + "/estado",
			Acciones:         accionesDeDefecto(d.Defecto.Estado),
			Sprints:          sprintsDeResolucion,
		}
	}
	vista := plantillas.VistaDefectos{
		Defectos: filas,
		Alta: plantillas.FormDefecto{
			Accion:       rutaDeDefectos(id),
			Severidades:  opcionesDeSeveridad(),
			Historias:    opcionesDeHistoriasDeDefecto(historias),
			Sprints:      sprintsDeResolucion,
			SinHistorias: len(historias) == 0,
			SinSprints:   len(sprints) == 0,
		},
	}
	Responder(w, r, http.StatusOK, paginaDeProyecto(detalle.Proyecto, plantillas.SeccionDefectos), plantillas.ListaDeDefectos(vista))
}

// registrarDefecto registra el defecto (US-022). La descripcion y la severidad se pasan tal como llegan:
// el recorte, el largo y las cuatro severidades los decide el dominio (RN-T021-6). Lo unico que revisa la
// pantalla es que se hayan elegido la historia y el sprint de deteccion (RN-T021-5): un ID vacio es
// "inexistente" para el dominio, y mostrarle un 404 a quien se olvido de elegir seria mala respuesta.
func (p *Pantallas) registrarDefecto(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := leerFormulario(r); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	historia, sprintDeteccion := r.PostForm.Get("historia"), r.PostForm.Get("sprint_deteccion")

	var sinElegir []errores.ErrorDeCampo
	if historia == "" {
		sinElegir = append(sinElegir, errores.ErrorDeCampo{Campo: "historia", Mensaje: mensajeHistoriaSinElegir})
	}
	if sprintDeteccion == "" {
		sinElegir = append(sinElegir, errores.ErrorDeCampo{Campo: "sprint_deteccion", Mensaje: mensajeSprintSinElegir})
	}
	if len(sinElegir) > 0 {
		// RN-T021-10: un proyecto que no existe es un 404 antes que el aviso de una lista sin elegir.
		if _, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id); err != nil {
			p.errores.ResponderError(w, r, err)
			return
		}
		p.errores.ResponderError(w, r, &errores.ErrorDeValidacion{Errores: sinElegir})
		return
	}

	_, err := p.casos.RegistrarDefecto.Ejecutar(r.Context(), id, historia, sprintDeteccion,
		r.PostForm.Get("descripcion"), defecto.Severidad(r.PostForm.Get("severidad")))
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeDefectos(id))
}

// cambiarEstadoDeDefecto cambia el estado del defecto (US-023). El estado y el sprint de resolucion se pasan
// tal como llegan: un estado fuera de los cuatro, una transicion no permitida, un sprint que falta o sobra o
// uno anterior al de deteccion los decide el dominio (RN-T021-7). Un defecto o un sprint de otro proyecto,
// inexistente o mal formado es un 404 que resuelve el caso de uso (RN-T021-11, RN-T021-12).
func (p *Pantallas) cambiarEstadoDeDefecto(w http.ResponseWriter, r *http.Request) {
	id, defectoID := r.PathValue("id"), r.PathValue("defecto")
	if _, err := leerFormulario(r); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	_, err := p.casos.CambiarEstadoDefecto.Ejecutar(r.Context(), id, defectoID,
		defecto.Estado(r.PostForm.Get("estado")), r.PostForm.Get("sprint_resolucion"))
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeDefectos(id))
}
