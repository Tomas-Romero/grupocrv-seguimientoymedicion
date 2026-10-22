package planningpoker_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// sesionEn arma una sesion de tres participantes en el estado pedido.
func sesionEn(estado planningpoker.Estado) planningpoker.SesionPlanningPoker {
	return planningpoker.SesionPlanningPoker{
		ID: "sesion-1", ItemID: "item-1", Estado: estado, Ronda: 1,
		Participantes: []string{"ana", "beto", "caro"},
	}
}

// US-015 / CA-015-1, CA-015-4, RN-015-2 a RN-015-4: un participante vota con la sesion
// votando y un valor de la escala; 1 y 13 son los extremos.
func TestVerificarVoto_Valido(t *testing.T) {
	for _, valor := range backlog.EscalaFibonacci() {
		if err := planningpoker.VerificarVoto(sesionEn(planningpoker.EstadoVotando), "ana", valor); err != nil {
			t.Errorf("VerificarVoto(%d) = %v, se esperaba un voto valido", valor, err)
		}
	}
}

// US-015 / CA-015-4, CA-015-5, CA-015-6, RN-015-2 a RN-015-4, CL-015-2, CL-015-6: un
// valor fuera de la escala, un integrante que no participa o una sesion que no esta
// votando se rechazan.
func TestVerificarVoto_Rechazos(t *testing.T) {
	casos := []struct {
		nombre      string
		estado      planningpoker.Estado
		integrante  string
		valor       int
		errEsperado error
	}{
		{"cero", planningpoker.EstadoVotando, "ana", 0, backlog.ErrEstimacionFueraDeEscala},
		{"cuatro", planningpoker.EstadoVotando, "ana", 4, backlog.ErrEstimacionFueraDeEscala},
		{"veintiuno", planningpoker.EstadoVotando, "ana", 21, backlog.ErrEstimacionFueraDeEscala},
		{"no participa", planningpoker.EstadoVotando, "dani", 5, planningpoker.ErrNoEsParticipante},
		{"sesion revelada", planningpoker.EstadoRevelada, "ana", 5, planningpoker.ErrVotacionCerrada},
		{"sesion finalizada", planningpoker.EstadoFinalizada, "ana", 5, planningpoker.ErrVotacionCerrada},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			err := planningpoker.VerificarVoto(sesionEn(c.estado), c.integrante, c.valor)
			if !errors.Is(err, c.errEsperado) {
				t.Errorf("error = %v, se esperaba %v", err, c.errEsperado)
			}
		})
	}
}

// US-015 / RN-015-8, CL-015-7: el dominio valida todo y devuelve los errores unidos, en
// orden fijo: votacion cerrada, no es participante, valor fuera de la escala.
func TestVerificarVoto_ErroresUnidosEnOrden(t *testing.T) {
	err := planningpoker.VerificarVoto(sesionEn(planningpoker.EstadoRevelada), "dani", 4)

	esperados := []error{planningpoker.ErrVotacionCerrada, planningpoker.ErrNoEsParticipante, backlog.ErrEstimacionFueraDeEscala}
	posicion := -1
	for _, esperado := range esperados {
		if !errors.Is(err, esperado) {
			t.Fatalf("error = %v, se esperaba que incluya %v", err, esperado)
		}
		i := strings.Index(err.Error(), esperado.Error())
		if i < posicion {
			t.Errorf("el error %q no respeta el orden de RN-015-8", err)
		}
		posicion = i
	}
}

// US-015 / CA-015-1, CA-015-3, RN-015-7, CL-015-4: la votacion dice quien voto y quien
// falta, en el orden de los participantes. Un voto de alguien que no participa no cuenta.
func TestVotacion_VotaronYFaltan(t *testing.T) {
	casos := []struct {
		nombre          string
		votos           map[string]int
		votaron, faltan []string
	}{
		{"nadie voto", nil, []string{}, []string{"ana", "beto", "caro"}},
		{"votaron dos", map[string]int{"caro": 8, "ana": 5}, []string{"ana", "caro"}, []string{"beto"}},
		{"votaron todos", map[string]int{"ana": 5, "beto": 3, "caro": 8}, []string{"ana", "beto", "caro"}, []string{}},
		{"voto alguien que no participa", map[string]int{"dani": 5}, []string{}, []string{"ana", "beto", "caro"}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			v := planningpoker.NuevaVotacion(sesionEn(planningpoker.EstadoVotando), c.votos)
			if !slices.Equal(v.Votaron(), c.votaron) || !slices.Equal(v.Faltan(), c.faltan) {
				t.Errorf("votaron %q y faltan %q, se esperaban %q y %q", v.Votaron(), v.Faltan(), c.votaron, c.faltan)
			}
		})
	}
}

// US-015 / CA-015-3, RN-015-7, CL-015-5, CL-015-11: con la sesion votando no hay forma de
// leer los valores, aunque hayan votado todos. Se prueba cada camino por el que podrian
// salir del dominio: el tipo no exporta campos; sus unicos metodos exportados son
// Votaron, Faltan, Valores, String y GoString; Valores se niega; y ni fmt ni encoding/json
// muestran los votos.
func TestVotacion_NoHayFormaDeLeerLosValoresMientrasSeVota(t *testing.T) {
	// 13 y 8 no aparecen en ningun ID: si salen en algun texto, son los votos.
	v := planningpoker.NuevaVotacion(sesionEn(planningpoker.EstadoVotando), map[string]int{"ana": 13, "beto": 8, "caro": 13})

	tipo := reflect.TypeOf(v)
	for i := range tipo.NumField() {
		if tipo.Field(i).IsExported() {
			t.Errorf("Votacion exporta el campo %s", tipo.Field(i).Name)
		}
	}
	var metodos []string
	for i := range tipo.NumMethod() {
		metodos = append(metodos, tipo.Method(i).Name)
	}
	if !slices.Equal(metodos, []string{"Faltan", "GoString", "String", "Valores", "Votaron"}) {
		t.Errorf("metodos exportados = %q: un metodo nuevo podria exponer los votos", metodos)
	}

	valores, err := v.Valores()
	if !errors.Is(err, planningpoker.ErrVotosOcultos) || valores != nil {
		t.Errorf("Valores() = (%v, %v), se esperaba (nil, ErrVotosOcultos)", valores, err)
	}

	textos := map[string]string{}
	for _, verbo := range []string{"%v", "%+v", "%#v", "%s"} {
		textos[verbo] = fmt.Sprintf(verbo, v)
	}
	if j, err := json.Marshal(v); err == nil {
		textos["json"] = string(j)
	}
	for como, texto := range textos {
		if strings.Contains(texto, "13") || strings.Contains(texto, "8") {
			t.Errorf("con %s la votacion muestra los votos: %q", como, texto)
		}
	}
}

// US-015 / RN-015-7: cuando la sesion deja de estar votando (US-016), los valores se
// pueden leer.
func TestVotacion_ValoresDespuesDeRevelar(t *testing.T) {
	votos := map[string]int{"ana": 5, "beto": 8}
	for _, estado := range []planningpoker.Estado{planningpoker.EstadoRevelada, planningpoker.EstadoFinalizada} {
		t.Run(string(estado), func(t *testing.T) {
			valores, err := planningpoker.NuevaVotacion(sesionEn(estado), votos).Valores()
			if err != nil || !maps.Equal(valores, votos) {
				t.Errorf("Valores() = (%v, %v), se esperaban %v", valores, err, votos)
			}
		})
	}
}

// La votacion tiene su propia copia de los votos y de los participantes: cambiar lo
// recibido, o lo que devuelve, no la cambia.
func TestVotacion_EsUnaCopia(t *testing.T) {
	sesion := sesionEn(planningpoker.EstadoRevelada)
	votos := map[string]int{"ana": 5}
	v := planningpoker.NuevaVotacion(sesion, votos)

	votos["ana"] = 13
	sesion.Participantes[0] = "cambiado"
	devueltos, _ := v.Valores()
	devueltos["ana"] = 1
	v.Votaron()[0] = "otro"

	if valores, _ := v.Valores(); valores["ana"] != 5 {
		t.Errorf("el voto de ana cambio a %d", valores["ana"])
	}
	if v.Votaron()[0] != "ana" {
		t.Errorf("votaron = %q, cambiar lo recibido o lo devuelto modifico la votacion", v.Votaron())
	}
}

// US-015 / spec seccion 7: los mensajes que ve la persona.
func TestErroresDelVoto_Mensajes(t *testing.T) {
	for err, esperado := range map[error]string{
		planningpoker.ErrVotacionCerrada:  "la votacion de esta ronda esta cerrada",
		planningpoker.ErrNoEsParticipante: "el integrante no participa de esta sesion de planning poker",
		planningpoker.ErrVotosOcultos:     "los votos estan ocultos hasta que se revelen",
	} {
		if err.Error() != esperado {
			t.Errorf("mensaje = %q, se esperaba %q", err.Error(), esperado)
		}
	}
}
