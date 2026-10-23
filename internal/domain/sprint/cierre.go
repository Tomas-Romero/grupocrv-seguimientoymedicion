package sprint

import "fmt"

// Resultado es lo que guarda la relacion de US-009 para cada historia en un sprint
// (RN-009-13): pendiente mientras el sprint esta abierto, y completado o no_completado
// cuando se cierra (US-011).
type Resultado string

// Los resultados posibles de una historia en un sprint.
const (
	ResultadoPendiente    Resultado = "pendiente"
	ResultadoCompletado   Resultado = "completado"
	ResultadoNoCompletado Resultado = "no_completado"
)

// HistoriaDelCierre son los datos minimos de una historia del sprint que necesita el
// cierre: si esta completada. Pendiente y en progreso son, para el cierre, lo mismo.
type HistoriaDelCierre struct {
	ItemID     string
	Completada bool
}

// ResultadoDelCierre es el resultado con que queda una historia en el sprint cerrado.
type ResultadoDelCierre struct {
	ItemID    string
	Resultado Resultado
}

// Cerrar valida que el sprint se pueda cerrar y devuelve el resultado de cada una de
// sus historias, en el mismo orden (RN-011-5): completado si esta completada y
// no_completado si no. Las no completadas vuelven al backlog: quien guarda el cierre
// las pasa a pendiente (RN-011-6). Un sprint sin historias tambien se cierra
// (RN-011-4), y no se mira la fecha (RN-011-3).
//
// Un sprint ya cerrado devuelve ErrSprintCerrado, envuelto con su nombre, y ningun
// resultado (RN-011-2). La usa el caso de uso con los datos que leyo y la repite el
// repositorio con el sprint bloqueado (RN-011-8).
func Cerrar(s DatosDelSprint, historias []HistoriaDelCierre) ([]ResultadoDelCierre, error) {
	if s.Estado == EstadoCerrado {
		return nil, fmt.Errorf("%w (%s)", ErrSprintCerrado, s.Nombre)
	}
	resultados := make([]ResultadoDelCierre, len(historias))
	for i, h := range historias {
		resultados[i] = ResultadoDelCierre{ItemID: h.ItemID, Resultado: ResultadoNoCompletado}
		if h.Completada {
			resultados[i].Resultado = ResultadoCompletado
		}
	}
	return resultados, nil
}
