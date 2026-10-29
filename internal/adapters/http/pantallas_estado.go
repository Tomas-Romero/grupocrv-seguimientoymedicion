package httpadapter

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/metricas"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// Tarjeta de estado general de la pagina de un proyecto (spec US-004). El handler solo pide el estado y lo escribe:
// los calculos viven en el caso de uso y en el dominio.

// tarjetaDeEstado arma la tarjeta. Un proyecto que no existe o una falla de la base se devuelven como error (404 o 500 de la
// pagina); un dato corrupto que rechaza metricas lo informa la tarjeta y el resto de la pagina se muestra (RN-004-9).
func (p *Pantallas) tarjetaDeEstado(r *http.Request, proyectoID string) (plantillas.VistaEstadoProyecto, error) {
	estado, err := p.casos.ConsultarEstado.Ejecutar(r.Context(), proyectoID)
	if err != nil {
		if errors.Is(err, metricas.ErrStoryPointsNegativos) {
			return vistaDelEstadoConError(err), nil
		}
		return plantillas.VistaEstadoProyecto{}, err
	}
	return vistaDelEstado(estado, proyectoID), nil
}

// vistaDelEstadoConError es la tarjeta cuando metricas rechaza un dato corrupto: informa el error en lugar de los numeros.
func vistaDelEstadoConError(err error) plantillas.VistaEstadoProyecto {
	return plantillas.VistaEstadoProyecto{Estado: "No se pudo calcular el avance del proyecto", Error: err.Error()}
}

// vistaDelEstado escribe el estado general como lo lee una persona. Lo usan la pagina del proyecto y el dashboard.
func vistaDelEstado(estado app.EstadoGeneral, proyectoID string) plantillas.VistaEstadoProyecto {
	tarjeta := plantillas.VistaEstadoProyecto{
		Estado: textoDelEstadoDelProyecto(estado.Estado, estado.Dias),
		Historias: "Pendientes: " + strconv.Itoa(estado.Historias.Pendientes) + " · En progreso: " + strconv.Itoa(estado.Historias.EnProgreso) +
			" · Completadas: " + strconv.Itoa(estado.Historias.Completadas) + " (de " + strconv.Itoa(estado.Historias.Total) + ")",
		StoryPoints: "Story Points: " + strconv.Itoa(estado.Completados) + " completados de " + strconv.Itoa(estado.StoryPoints),
		Porcentaje:  textoDePorcentaje(estado.Porcentaje, true) + " de las historias completadas",
		Sprint:      "Sin sprint abierto",
	}
	if estado.SprintAbierto != nil {
		tarjeta.Sprint = estado.SprintAbierto.Nombre() + " en curso"
		tarjeta.SprintEnlace = rutaDeSprint(proyectoID, estado.SprintAbierto.ID)
	}
	return tarjeta
}

// textoDelEstadoDelProyecto es el estado como se lee: «En curso: faltan 40 días», «No iniciado: empieza en 12 días»,
// «Finalizado: terminó hace 3 días» o «En curso: termina hoy» (RN-004-1, RN-004-2).
func textoDelEstadoDelProyecto(estado proyecto.EstadoDelProyecto, dias int) string {
	switch estado {
	case proyecto.NoIniciado:
		return "No iniciado: empieza en " + textoDeDias(dias)
	case proyecto.Finalizado:
		return "Finalizado: terminó hace " + textoDeDias(dias)
	}
	switch dias {
	case 0:
		return "En curso: termina hoy"
	case 1:
		return "En curso: falta 1 día"
	}
	return "En curso: faltan " + textoDeDias(dias)
}

// textoDeDias escribe «1 día» o «N días».
func textoDeDias(n int) string {
	if n == 1 {
		return "1 día"
	}
	return strconv.Itoa(n) + " días"
}
