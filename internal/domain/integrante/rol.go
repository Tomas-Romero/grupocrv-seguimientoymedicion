package integrante

// Rol es el rol de un integrante dentro de un proyecto (RN-003-5).
type Rol string

// Los tres valores que acepta el dominio, exactos y en minuscula.
const (
	RolProductArchitect Rol = "product_architect"
	RolAgileEnabler     Rol = "agile_enabler"
	RolProductBuilder   Rol = "product_builder"
)

// valido dice si r es uno de los tres valores exactos. No pasa a minuscula ni
// recorta: convertir el texto de un formulario a Rol es tarea del adaptador, y
// "PRODUCT_BUILDER" o " agile_enabler " se rechazan (CL-003-10).
func (r Rol) valido() bool {
	switch r {
	case RolProductArchitect, RolAgileEnabler, RolProductBuilder:
		return true
	}
	return false
}
