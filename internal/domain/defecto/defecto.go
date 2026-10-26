// Package defecto contiene las reglas de un defecto: lo que se registra de una falla
// del producto, con su severidad, la historia donde se introdujo y el sprint en el que
// se detecto (US-022).
//
// Es dominio puro: no importa nada fuera de la biblioteca estandar, no accede a la
// base, no genera IDs y no lee el reloj (la fecha de carga la pone la base). Que el
// proyecto, la historia y el sprint existan y se correspondan lo verifica la capa de
// aplicacion.
package defecto

// Severidad es la gravedad de un defecto (RN-022-2).
type Severidad string

// Valores de Severidad. Son los que viajan al dominio y se guardan: las etiquetas con
// tilde que ve la persona ("Critica") las convierte el adaptador HTTP.
const (
	SeveridadCritica Severidad = "critica"
	SeveridadAlta    Severidad = "alta"
	SeveridadMedia   Severidad = "media"
	SeveridadBaja    Severidad = "baja"
)

// Estado es el estado de un defecto en su ciclo de vida.
type Estado string

// EstadoAbierto es el unico estado que existe hasta que US-023 define los demas:
// todo defecto nace abierto (RN-022-3).
const EstadoAbierto Estado = "abierto"

// Defecto es una falla registrada en un proyecto.
type Defecto struct {
	// ID esta vacio hasta que la persistencia lo asigna (RN-022-10).
	ID         string
	ProyectoID string
	// ItemID es la historia relacionada: donde se introdujo el defecto.
	ItemID string
	// SprintDeteccionID es el sprint donde se detecto, por su ID y no por su numero
	// (RN-022-6).
	SprintDeteccionID string
	Descripcion       string
	Severidad         Severidad
	Estado            Estado
	// SprintResolucionID esta vacio mientras el defecto no se resuelve; solo lo
	// completa US-023 (RN-022-3).
	SprintResolucionID string
}

// Nuevo arma el defecto, sin ID, siempre abierto y sin sprint de resolucion (RN-022-3).
//
// Los IDs se copian tal cual: que existan y sean del proyecto lo verifica la capa de
// aplicacion (RN-022-5), nunca el dominio.
func Nuevo(proyectoID, itemID, sprintDeteccionID, descripcion string, severidad Severidad) (Defecto, error) {
	return Defecto{
		ProyectoID:        proyectoID,
		ItemID:            itemID,
		SprintDeteccionID: sprintDeteccionID,
		Descripcion:       descripcion,
		Severidad:         severidad,
		Estado:            EstadoAbierto,
	}, nil
}
