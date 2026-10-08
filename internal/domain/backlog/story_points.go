package backlog

// StoryPoints es la estimacion de un item. Su valor cero significa "sin
// estimar": sin estimar y estimado en 0 no son lo mismo (RN-005-7), y asi un
// item nunca queda con 0 puntos por accidente. Estimar es US-013.
type StoryPoints struct {
	valor    int
	estimado bool
}

// SinEstimar devuelve unos Story Points sin estimar.
func SinEstimar() StoryPoints {
	return StoryPoints{}
}

// Valor devuelve los puntos y si el item esta estimado. Sin estimar, devuelve
// (0, false).
func (s StoryPoints) Valor() (puntos int, estimado bool) {
	return s.valor, s.estimado
}
