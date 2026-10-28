package httpadapter

import (
	"net/http"
	"strconv"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// Pantallas de Planning Poker (spec T-020): iniciar una sesion desde el backlog, la pantalla de
// la sesion y sus acciones. Cada handler solo lee la peticion, comprueba que la sesion sea del
// proyecto con ConsultarSesion, llama a un caso de uso y escribe la respuesta (RN-T020-1).

func rutaDeSesion(proyectoID, sesionID string) string {
	return rutaDeProyecto(proyectoID) + "/planning-poker/" + sesionID
}

// iniciarSesion inicia una sesion sobre la historia (US-014) y lleva a su pantalla.
func (p *Pantallas) iniciarSesion(w http.ResponseWriter, r *http.Request) {
	id, itemID := r.PathValue("id"), r.PathValue("item")
	sesion, err := p.casos.IniciarSesionPlanningPoker.Ejecutar(r.Context(), id, itemID)
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeSesion(id, sesion.ID))
}

func (p *Pantallas) verSesion(w http.ResponseWriter, r *http.Request) {
	id, sesionID := r.PathValue("id"), r.PathValue("sesion")
	proyecto, err := p.casos.ConsultarProyecto.Ejecutar(r.Context(), id)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}
	detalle, err := p.casos.ConsultarSesion.Ejecutar(r.Context(), id, sesionID)
	if err != nil {
		p.fallaDeLectura(w, r, err)
		return
	}

	nombres := map[string]string{}
	for _, i := range proyecto.Integrantes {
		nombres[i.ID] = i.Nombre
	}
	deNombres := func(ids []string) []string {
		out := make([]string, len(ids))
		for i, id := range ids {
			out[i] = nombres[id]
		}
		return out
	}
	s := detalle.Sesion
	integrantes := make([]plantillas.Opcion, len(s.Participantes))
	for i, pid := range s.Participantes {
		integrantes[i] = plantillas.Opcion{Valor: pid, Texto: nombres[pid], Elegida: i == 0}
	}
	ruta := rutaDeSesion(id, sesionID)
	vista := plantillas.VistaSesion{
		Historia:      "#" + strconv.Itoa(detalle.Historia.Numero) + " " + detalle.Historia.Titulo,
		Ronda:         s.Ronda,
		Estado:        string(s.Estado),
		Participantes: deNombres(s.Participantes),
		Integrantes:   integrantes,
		Volver:        rutaDeProyecto(id) + "/backlog",
		Anteriores:    make([]plantillas.ResultadoVisible, len(detalle.Anteriores)),
	}
	for i, a := range detalle.Anteriores {
		vista.Anteriores[i] = resultadoVisible(a.Numero, a.Resultado, nombres)
	}

	// RN-T020-5: lo que se muestra y se ofrece depende del estado.
	switch s.Estado {
	case planningpoker.EstadoVotando:
		vista.Votando = true
		vista.Votaron, vista.Faltan = deNombres(detalle.Votaron), deNombres(detalle.Faltan)
		vista.AccionVotar, vista.AccionRevelar = ruta+"/votos", ruta+"/revelar"
		vista.NombreRevelar = "Revelar los votos de la ronda " + strconv.Itoa(s.Ronda)
		for _, valor := range backlog.EscalaFibonacci() {
			texto := strconv.Itoa(valor)
			vista.Cartas = append(vista.Cartas, plantillas.Carta{Valor: texto, Nombre: "Votar " + texto})
		}
	case planningpoker.EstadoRevelada:
		vista.Revelada = true
		vista.Resultado = resultadoVisible(s.Ronda, detalle.Resultado, nombres)
		if detalle.Resultado.Divergencia {
			vista.AccionNuevaRonda = ruta + "/nueva-ronda"
		}
		vista.AccionAcuerdo = ruta + "/acuerdo"
		vista.Escala = opcionesDeEscala(0, false)
	case planningpoker.EstadoFinalizada:
		vista.Finalizada = true
		if valor, ok := s.ValorAcordado.Valor(); ok {
			vista.Acordada = strconv.Itoa(valor)
		}
	}

	pagina := paginaDeProyecto(proyecto.Proyecto, plantillas.SeccionBacklog)
	pagina.Titulo = "Planning Poker | " + proyecto.Proyecto.Nombre
	Responder(w, r, http.StatusOK, pagina, plantillas.SesionDePlanningPoker(vista))
}

// resultadoVisible arma el resultado de una ronda con los nombres y el texto de la divergencia
// (RN-T020-6, RN-T020-8).
func resultadoVisible(ronda int, r planningpoker.Resultado, nombres map[string]string) plantillas.ResultadoVisible {
	v := plantillas.ResultadoVisible{Ronda: ronda, Divergencia: "Sin divergencia"}
	for _, voto := range r.Votos {
		v.Votos = append(v.Votos, plantillas.VotoVisible{Integrante: nombres[voto.ParticipanteID], Valor: voto.Valor})
	}
	for _, id := range r.Faltaron {
		v.Faltaron = append(v.Faltaron, nombres[id])
	}
	if r.Divergencia {
		v.Divergencia = "Hay divergencia: de " + strconv.Itoa(r.Minimo) + " a " + strconv.Itoa(r.Maximo)
	}
	return v
}

func (p *Pantallas) votarEnSesion(w http.ResponseWriter, r *http.Request) {
	p.accionDeSesion(w, r, func(sesionID, integranteID string, f *Formulario) error {
		valor := f.Entero("estimacion")
		if err := f.Err(); err != nil {
			return err
		}
		_, err := p.casos.VotarEnSesion.Ejecutar(r.Context(), sesionID, integranteID, valor)
		return err
	})
}

func (p *Pantallas) revelarSesion(w http.ResponseWriter, r *http.Request) {
	p.accionDeSesion(w, r, func(sesionID, integranteID string, _ *Formulario) error {
		_, err := p.casos.RevelarVotos.Ejecutar(r.Context(), sesionID, integranteID)
		return err
	})
}

func (p *Pantallas) abrirNuevaRonda(w http.ResponseWriter, r *http.Request) {
	p.accionDeSesion(w, r, func(sesionID, integranteID string, _ *Formulario) error {
		_, err := p.casos.AbrirNuevaRonda.Ejecutar(r.Context(), sesionID, integranteID)
		return err
	})
}

func (p *Pantallas) registrarAcuerdo(w http.ResponseWriter, r *http.Request) {
	p.accionDeSesion(w, r, func(sesionID, integranteID string, f *Formulario) error {
		valor := f.Entero("estimacion")
		if err := f.Err(); err != nil {
			return err
		}
		_, err := p.casos.RegistrarAcuerdo.Ejecutar(r.Context(), sesionID, integranteID, valor)
		return err
	})
}

// accionDeSesion lee el formulario, comprueba que la sesion sea del proyecto (RN-T020-4: la de
// otro proyecto, o una que no existe, es un 404 antes de llamar al caso de uso), hace la accion
// y vuelve a la pantalla de la sesion.
func (p *Pantallas) accionDeSesion(w http.ResponseWriter, r *http.Request, accion func(sesionID, integranteID string, f *Formulario) error) {
	id, sesionID := r.PathValue("id"), r.PathValue("sesion")
	f, err := leerFormulario(r)
	if err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	if _, err := p.casos.ConsultarSesion.Ejecutar(r.Context(), id, sesionID); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	if err := accion(sesionID, r.PostForm.Get("integrante"), f); err != nil {
		p.errores.ResponderError(w, r, err)
		return
	}
	redirigir(w, r, rutaDeSesion(id, sesionID))
}

// sesionesDelBacklog devuelve, para cada historia, la ruta de su sesion abierta.
func (p *Pantallas) sesionesDelBacklog(r *http.Request, proyectoID string) (map[string]string, error) {
	abiertas, err := p.casos.ConsultarSesion.AbiertasDelProyecto(r.Context(), proyectoID)
	if err != nil {
		return nil, err
	}
	rutas := make(map[string]string, len(abiertas))
	for item, sesion := range abiertas {
		rutas[item] = rutaDeSesion(proyectoID, sesion)
	}
	return rutas, nil
}
