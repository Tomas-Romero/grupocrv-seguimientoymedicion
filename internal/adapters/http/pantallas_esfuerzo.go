package httpadapter

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/errores"
	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// Handlers de las pantallas de esfuerzo (spec T-019). Igual que los de T-011 y T-015, solo leen la
// peticion, llaman a un caso de uso y escriben la respuesta.

func rutaDeEsfuerzo(proyectoID string) string { return rutaDeProyecto(proyectoID) + "/esfuerzo" }

// ahora es el reloj de la pantalla: el valor inicial de la fecha (RN-T019, seccion 5).
func (p *Pantallas) ahora() time.Time {
	if p.casos.Ahora != nil {
		return p.casos.Ahora()
	}
	return time.Now()
}

func (p *Pantallas) verEsfuerzo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detalle, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	consulta := r.URL.Query()
	filtro := app.FiltroEsfuerzo{ItemID: consulta.Get("historia"), IntegranteID: consulta.Get("integrante"), SprintID: consulta.Get("sprint")}
	registros, err := p.casos.ConsultarEsfuerzo.Ejecutar(r.Context(), id, filtro)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	items, err := p.casos.ListarBacklog.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	sprints, err := p.casos.ListarSprints.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}

	historias := []plantillas.Opcion{{Valor: "", Texto: "Todas", Elegida: filtro.ItemID == ""}}
	elegibles := make([]plantillas.Opcion, len(items))
	for i, item := range items {
		historias = append(historias, plantillas.Opcion{Valor: item.ID, Texto: item.Titulo, Elegida: item.ID == filtro.ItemID})
		elegibles[i] = plantillas.Opcion{Valor: item.ID, Texto: item.Titulo}
	}
	integrantes := []plantillas.Opcion{{Valor: "", Texto: "Todos", Elegida: filtro.IntegranteID == ""}}
	quienes := make([]plantillas.Opcion, len(detalle.Integrantes))
	for i, in := range detalle.Integrantes {
		integrantes = append(integrantes, plantillas.Opcion{Valor: in.ID, Texto: in.Nombre, Elegida: in.ID == filtro.IntegranteID})
		quienes[i] = plantillas.Opcion{Valor: in.ID, Texto: in.Nombre}
	}
	sprintsFiltro := []plantillas.Opcion{{Valor: "", Texto: "Todos", Elegida: filtro.SprintID == ""}}
	for _, s := range sprints {
		sprintsFiltro = append(sprintsFiltro, plantillas.Opcion{Valor: s.Sprint.ID, Texto: s.Sprint.Nombre(), Elegida: s.Sprint.ID == filtro.SprintID})
	}

	alta := plantillas.FormEsfuerzo{Accion: rutaDeEsfuerzo(id), Integrantes: quienes, Historias: elegibles, Fecha: fechaDeFormulario(p.ahora())}
	if len(quienes) == 0 {
		alta.AvisoIntegrantes = "Primero registrá un integrante"
	}
	if len(elegibles) == 0 {
		alta.AvisoHistorias = "Primero cargá una historia"
	}

	vista := plantillas.VistaEsfuerzo{
		Alta:     alta,
		Filtros:  plantillas.FiltrosEsfuerzo{Accion: rutaDeEsfuerzo(id), Historias: historias, Integrantes: integrantes, Sprints: sprintsFiltro},
		Total:    "Total: " + textoDeHoras(registros.TotalHoras) + " h",
		Comparar: rutaDeEsfuerzo(id) + "/comparacion",
	}
	if len(registros.Registros) == 0 {
		vista.Vacio = "Todavía no hay horas registradas"
		if filtro != (app.FiltroEsfuerzo{}) {
			vista.Vacio = "Ningún registro cumple los filtros"
		}
	}
	for _, reg := range registros.Registros {
		vista.Registros = append(vista.Registros, plantillas.FilaEsfuerzo{
			Fecha:      fechaVisible(reg.Registro.Fecha),
			Historia:   reg.Historia,
			Integrante: reg.Integrante,
			Actividad:  reg.Registro.Actividad,
			Horas:      textoDeHoras(reg.Registro.Horas),
			Sprint:     textoDeSprintDelRegistro(reg.SprintNumero),
		})
	}
	for _, t := range registros.PorHistoria {
		vista.PorHistoria = append(vista.PorHistoria, plantillas.FilaTotalEsfuerzo{Nombre: t.Titulo, Horas: textoDeHoras(t.Horas) + " h"})
	}
	for _, t := range registros.PorIntegrante {
		vista.PorIntegrante = append(vista.PorIntegrante, plantillas.FilaTotalEsfuerzo{Nombre: t.Nombre, Horas: textoDeHoras(t.Horas) + " h"})
	}
	for _, t := range registros.PorSprint {
		vista.PorSprint = append(vista.PorSprint, plantillas.FilaTotalEsfuerzo{Nombre: textoDeSprintDelRegistro(t.Numero), Horas: textoDeHoras(t.Horas) + " h"})
	}

	pagina := paginaDeProyecto(detalle.Proyecto, plantillas.SeccionEsfuerzo)
	pagina.Titulo = "Esfuerzo | " + detalle.Proyecto.Nombre
	Responder(w, r, http.StatusOK, pagina, plantillas.EsfuerzoDeProyecto(vista))
}

// registrarEsfuerzo registra las horas (US-019). El proyecto se verifica primero (RN-T019-12); despues,
// los campos que la pantalla revisa por su cuenta (RN-T019-3 y RN-T019-4) se informan todos juntos y sin
// llamar al caso de uso. Lo demas lo decide el dominio.
func (p *Pantallas) registrarEsfuerzo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	f, err := leerFormulario(r)
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	if _, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	integranteID, itemID := r.PostForm.Get("integrante"), r.PostForm.Get("historia")
	fecha, horas := f.Fecha("fecha"), f.Decimal("horas")

	var sinElegir []errores.ErrorDeCampo
	if integranteID == "" {
		sinElegir = append(sinElegir, errores.ErrorDeCampo{Campo: "integrante", Mensaje: "Elegí un integrante"})
	}
	if itemID == "" {
		sinElegir = append(sinElegir, errores.ErrorDeCampo{Campo: "historia", Mensaje: "Elegí una historia"})
	}
	if err := juntarErroresDeCampo(sinElegir, f.Err()); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}

	if _, err := p.casos.RegistrarEsfuerzo.Ejecutar(r.Context(), id, itemID, integranteID, fecha, r.PostForm.Get("actividad"), horas); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeEsfuerzo(id))
}

// juntarErroresDeCampo une los campos sin elegir con los mal formados del formulario, en ese orden.
func juntarErroresDeCampo(sinElegir []errores.ErrorDeCampo, deFormato error) error {
	var formato *errores.ErrorDeValidacion
	if errors.As(deFormato, &formato) {
		sinElegir = append(sinElegir, formato.Errores...)
	}
	if len(sinElegir) == 0 {
		return nil
	}
	return &errores.ErrorDeValidacion{Errores: sinElegir}
}

func (p *Pantallas) compararEsfuerzo(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detalle, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	comparacion, err := p.casos.CompararEsfuerzo.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}

	vista := plantillas.VistaComparacion{Volver: rutaDeEsfuerzo(id)}
	if len(comparacion.Historias) == 0 {
		vista.Vacio = "Todavía no hay historias para comparar"
	} else {
		for _, h := range comparacion.Historias {
			vista.Filas = append(vista.Filas, filaDeComparacion(h))
		}
		porcentaje, calculable := comparacion.Total.DesviacionPorcentual()
		vista.Resumen = textoDeComparables(comparacion.Comparables) + ": estimadas " + textoDeHoras(comparacion.Total.Estimadas) +
			" h, reales " + textoDeHoras(comparacion.Total.Reales) + " h, desviación " + textoDeDesviacion(comparacion.Total.Desviacion) +
			" h (" + textoDePorcentaje(porcentaje, calculable) + ")"
		vista.SinComparar = "Horas sin comparar: " + textoDeHoras(comparacion.HorasSinComparar) + " h"
	}
	pagina := paginaDeProyecto(detalle.Proyecto, plantillas.SeccionEsfuerzo)
	pagina.Titulo = "Estimado contra real | " + detalle.Proyecto.Nombre
	Responder(w, r, http.StatusOK, pagina, plantillas.ComparacionDeEsfuerzo(vista))
}

// filaDeComparacion escribe una historia de la comparacion. Una historia sin estimar no tiene base de
// comparacion: sus Story Points, sus horas estimadas y su desviacion se muestran como un guion y no como un cero
// que parezca un dato.
func filaDeComparacion(h metricas.ComparacionHistoria) plantillas.FilaComparacion {
	fila := plantillas.FilaComparacion{
		Titulo: h.Titulo, Puntos: "—", Estimadas: "—", Reales: textoDeHoras(h.Horas.Reales), Desviacion: "—", Porcentaje: "—",
		Situacion: "En curso",
	}
	switch {
	case !h.Estimada:
		fila.Situacion = "Sin estimar"
	case h.Completada:
		fila.Situacion = "Completada"
	}
	if h.Estimada {
		porcentaje, calculable := h.DesviacionPorcentual()
		fila.Puntos = strconv.Itoa(h.StoryPoints)
		fila.Estimadas = textoDeHoras(h.Horas.Estimadas)
		fila.Desviacion = textoDeDesviacion(h.Horas.Desviacion)
		fila.Porcentaje = textoDePorcentaje(porcentaje, calculable)
	}
	return fila
}

// textoDeSprintDelRegistro es el sprint donde cae un registro; el numero 0 son las horas fuera de todo sprint.
func textoDeSprintDelRegistro(numero int) string {
	if numero == 0 {
		return "Fuera de sprint"
	}
	return "Sprint " + strconv.Itoa(numero)
}

// textoDeHoras escribe las horas como las lee una persona: coma decimal y sin ceros de mas (2,5; 8; 12).
func textoDeHoras(horas float64) string { return textoDeVelocidad(horas) }

// textoDeDesviacion escribe la desviacion en horas con signo (+0,5; -3); cero es 0.
func textoDeDesviacion(horas float64) string {
	texto := textoDeHoras(horas)
	if horas > 0 {
		return "+" + texto
	}
	return texto
}

// textoDePorcentaje escribe un porcentaje con un decimal y el simbolo (6,7 %); si no se puede calcular, un guion.
func textoDePorcentaje(porcentaje float64, calculable bool) string {
	if !calculable {
		return "—"
	}
	if porcentaje == 0 || math.IsNaN(porcentaje) {
		porcentaje = 0
	}
	return strings.Replace(strconv.FormatFloat(porcentaje, 'f', 1, 64), ".", ",", 1) + " %"
}

func textoDeComparables(n int) string {
	if n == 1 {
		return "1 historia comparable"
	}
	return strconv.Itoa(n) + " historias comparables"
}
