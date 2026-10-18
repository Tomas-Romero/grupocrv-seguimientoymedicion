package backlog

import "fmt"

// Editar devuelve una copia del item con los cuatro datos reemplazados:
// titulo, descripcion, prioridad y criterios (RN-006-2). Repriorizar es esta
// misma operacion cambiando solo la prioridad (RN-006-3). El numero, el estado,
// los Story Points y la fecha de creacion no cambian (RN-006-4).
//
// Se valida primero el estado (RN-006-5): un item completado devuelve solo
// ErrItemCompletado, envuelto con su ID, aunque los datos tambien sean
// invalidos. Despues se validan los datos con la misma validacion que al crear
// (RN-006-1). Si hay error no se devuelve item, y el item recibido no cambia
// nunca (RN-006-10).
//
// No se vuelve a verificar la Definition of Ready (RN-006-6): un item asignado
// a un sprint puede quedar sin criterios.
func (i ItemBacklog) Editar(datos DatosItem) (ItemBacklog, error) {
	if i.Estado == EstadoCompletado {
		return ItemBacklog{}, fmt.Errorf("%w (id %s)", ErrItemCompletado, i.ID)
	}
	validos, err := validarDatos(datos)
	if err != nil {
		return ItemBacklog{}, err
	}
	i.Titulo = validos.Titulo
	i.Descripcion = validos.Descripcion
	i.Prioridad = validos.Prioridad
	// recortarCriterios ya devuelve una lista nueva: no se comparte con datos.
	i.Criterios = validos.Criterios
	return i, nil
}
