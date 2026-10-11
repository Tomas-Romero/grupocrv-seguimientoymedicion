package proyecto

import "time"

// Modificar valida los datos nuevos y devuelve el proyecto modificado.
//
// RN-002-2: las reglas son exactamente las de Nuevo (RN-001-1 a RN-001-5 y
// RN-001-7); se reutiliza esa validacion y no se escribe una segunda, asi el
// orden fijo de los errores y los limites no pueden divergir.
//
// RN-002-3: reemplaza los cuatro datos a la vez. RN-002-4: el ID es el de
// actual y nunca cambia. RN-002-5: es de todo o nada; si algun dato es
// invalido no devuelve un Proyecto. No altera actual: devuelve uno nuevo.
//
// RN-002-1 (que el proyecto exista) no se chequea aca: lo detecta la capa de
// aplicacion al buscarlo, antes de llamar a esta funcion.
func Modificar(actual Proyecto, nombre, descripcion string, fechaInicio, fechaFin time.Time) (Proyecto, error) {
	modificado, err := Nuevo(nombre, descripcion, fechaInicio, fechaFin)
	if err != nil {
		return Proyecto{}, err
	}

	modificado.ID = actual.ID
	return modificado, nil
}
