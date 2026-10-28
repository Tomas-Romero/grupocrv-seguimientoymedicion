package httpadapter

import (
	"fmt"
	"net/http"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

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

// listaDeDefectos es la pantalla con la lista de defectos del proyecto (RN-T021-2).
func (p *Pantallas) listaDeDefectos(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	detalle, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	defectos, err := p.casos.ListarDefectos.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}

	filas := make([]plantillas.FilaDefecto, len(defectos))
	for i, d := range defectos {
		filas[i] = plantillas.FilaDefecto{
			Descripcion:      d.Defecto.Descripcion,
			Severidad:        textoDeSeveridad(d.Defecto.Severidad),
			Estado:           textoDeEstadoDeDefecto(d.Defecto.Estado),
			Historia:         fmt.Sprintf("#%d %s", d.HistoriaNumero, d.HistoriaTitulo),
			SprintDeteccion:  textoDeSprintDeDefecto(d.SprintDeteccion),
			SprintResolucion: textoDeSprintDeDefecto(d.SprintResolucion),
		}
	}
	Responder(w, r, http.StatusOK, paginaDeProyecto(detalle.Proyecto, plantillas.SeccionDefectos),
		plantillas.ListaDeDefectos(plantillas.VistaDefectos{Defectos: filas}))
}
