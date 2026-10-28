package plantillas

import "strings"

// Datos que dibujan la pantalla de una sesion de Planning Poker (spec T-020). Igual que los de
// T-011 y T-015, son datos ya preparados: los textos del estado y de la divergencia los arma el
// handler (RN-T020-8). Ninguno trae el valor de un voto de la ronda actual mientras se vota
// (RN-T020-2).

// VotoVisible es un voto ya revelado, con el nombre de quien voto.
type VotoVisible struct {
	Integrante string
	Valor      int
}

// ResultadoVisible es el resultado de una ronda revelada.
type ResultadoVisible struct {
	Ronda    int
	Votos    []VotoVisible
	Faltaron []string
	// Divergencia es el texto de la divergencia: "Hay divergencia: de 3 a 8" o "Sin divergencia".
	Divergencia string
}

// Carta es un valor de la escala para votar.
type Carta struct {
	Valor string
	// Nombre es el nombre accesible del boton ("Votar 5").
	Nombre string
}

// VistaSesion es la pantalla /proyectos/{id}/planning-poker/{sesion}.
type VistaSesion struct {
	Historia      string
	Ronda         int
	Estado        string
	Participantes []string
	// Integrantes son los participantes para elegir quien vota, revela o registra.
	Integrantes []Opcion

	// Votando: quienes votaron y quienes faltan, y las acciones de votar y revelar.
	Votando       bool
	Votaron       []string
	Faltan        []string
	AccionVotar   string
	Cartas        []Carta
	AccionRevelar string
	NombreRevelar string

	// Revelada: el resultado de la ronda actual, la nueva ronda (solo con divergencia) y el
	// acuerdo.
	Revelada         bool
	Resultado        ResultadoVisible
	AccionNuevaRonda string
	AccionAcuerdo    string
	Escala           []Opcion

	// Finalizada: la estimacion acordada.
	Finalizada bool
	Acordada   string

	Anteriores []ResultadoVisible
	Volver     string
}

// joinNombres une los nombres con comas, para las listas cortas de la sesion.
func joinNombres(nombres []string) string {
	return strings.Join(nombres, ", ")
}
