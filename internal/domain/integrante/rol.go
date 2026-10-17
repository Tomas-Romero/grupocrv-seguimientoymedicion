package integrante

// Rol es el rol de un integrante dentro de un proyecto (RN-003-5).
type Rol string

// Los tres valores que acepta el dominio, exactos y en minuscula.
const (
	RolProductArchitect Rol = "product_architect"
	RolAgileEnabler     Rol = "agile_enabler"
	RolProductBuilder   Rol = "product_builder"
)
