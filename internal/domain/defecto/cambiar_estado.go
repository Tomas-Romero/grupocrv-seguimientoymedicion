package defecto

// SprintDeResolucion es el sprint en el que se resolvio un defecto, ya traducido por la capa
// de aplicacion (RN-023-8).
type SprintDeResolucion struct {
	// ID esta vacio si no se informo sprint.
	ID string
	// Numero es sprint.Numero. El dominio compara numeros y nunca IDs (RN-023-4).
	Numero int
}

// CambiarEstado devuelve una copia del defecto con el estado pedido. Solo resuelto lleva sprint
// de resolucion; en cualquier otro estado el defecto queda sin sprint, sin conservar el anterior
// (RN-023-5). No lee la base, no genera IDs y no sabe si los sprints existen: recibe sus
// numeros ya traducidos (RN-023-8).
func (d Defecto) CambiarEstado(nuevo Estado, resolucion SprintDeResolucion, _ int) (Defecto, error) {
	d.Estado = nuevo
	d.SprintResolucionID = resolucion.ID
	return d, nil
}
