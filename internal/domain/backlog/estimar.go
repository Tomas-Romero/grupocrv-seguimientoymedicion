package backlog

import (
	"errors"
	"slices"
)

// ErrItemCompletado: el item esta completado, asi que no se puede estimar ni
// re-estimar, porque cambiaria hacia atras los Story Points completados de un
// sprint ya medido (RN-013-4). El mensaje es general a proposito: lo reusan las
// otras operaciones sobre un item (editar en US-006, asignar en US-009).
var ErrItemCompletado = errors.New("el item esta completado")

// Estimar devuelve una copia del item con los Story Points puestos. Un item
// pendiente o en progreso se puede estimar y volver a estimar: la ultima
// estimacion reemplaza a la anterior (RN-013-3).
//
// Se valida primero el estado y despues el valor (RN-013-5): un item completado
// devuelve solo ErrItemCompletado, aunque el valor tambien sea invalido. Si hay
// error no se devuelve item, y el item recibido no cambia nunca (CL-013-11).
func (i ItemBacklog) Estimar(puntos int) (ItemBacklog, error) {
	if i.Estado == EstadoCompletado {
		return ItemBacklog{}, ErrItemCompletado
	}
	estimacion, err := NuevosStoryPoints(puntos)
	if err != nil {
		return ItemBacklog{}, err
	}
	i.StoryPoints = estimacion
	i.Criterios = slices.Clone(i.Criterios)
	return i, nil
}
