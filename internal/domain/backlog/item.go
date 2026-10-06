// Package backlog contiene las reglas del Product Backlog: los items, su
// prioridad MoSCoW, su estado y sus criterios de aceptacion (US-005).
//
// Es dominio puro: no importa nada fuera de la biblioteca estandar, no genera
// IDs y no lee el reloj. Que el proyecto exista, el ID y el numero del item los
// resuelven la aplicacion y la persistencia.
package backlog

import "time"

// DatosItem son los datos que carga la persona al crear un item.
type DatosItem struct {
	Titulo      string
	Descripcion string
	Prioridad   Prioridad
	Criterios   []string
}

// ItemBacklog es un item del Product Backlog de un proyecto.
type ItemBacklog struct {
	// ID lo genera la persistencia: esta vacio hasta que el item se registra.
	ID string
	// Numero es el correlativo dentro del proyecto (#1, #2, ...); lo asigna la
	// persistencia.
	Numero int
	// ProyectoID lo completa la persistencia al registrar el item.
	ProyectoID  string
	Titulo      string
	Descripcion string
	Prioridad   Prioridad
	Estado      Estado
	StoryPoints StoryPoints
	Criterios   []string
	CreadoEn    time.Time
}

// NuevoItem arma un item pendiente y sin estimar con los datos recibidos.
func NuevoItem(datos DatosItem, creadoEn time.Time) (ItemBacklog, error) {
	return ItemBacklog{
		Titulo:      datos.Titulo,
		Descripcion: datos.Descripcion,
		Prioridad:   datos.Prioridad,
		Estado:      EstadoPendiente,
		StoryPoints: SinEstimar(),
		Criterios:   datos.Criterios,
		CreadoEn:    creadoEn,
	}, nil
}
