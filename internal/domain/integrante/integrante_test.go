package integrante_test

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
)

// Datos validos que los casos reutilizan: cada caso cambia solo lo que prueba.
const (
	proyectoID = "0b6f3c52-4f0e-4a43-9d3e-6a1f2c7d9e10"
	nombreOK   = "Ana"
	emailOK    = "ana@ejemplo.test"
	rolOK      = integrante.RolProductBuilder
)

// todosLosErrores son los cinco errores de validacion del dominio: sirven para
// comprobar que un caso no devuelve errores de mas.
var todosLosErrores = []error{
	integrante.ErrNombreVacio,
	integrante.ErrNombreLargo,
	integrante.ErrEmailVacio,
	integrante.ErrEmailInvalido,
	integrante.ErrRolInvalido,
}

// verificarErrores comprueba que err une exactamente los errores esperados, en
// ese orden (RN-003-8): errors.Is reconoce los esperados y solo esos, y
// errors.Join deja una linea por error, en el orden en que se unieron.
func verificarErrores(t *testing.T, err error, esperados []error) {
	t.Helper()

	if err == nil {
		t.Fatal("se esperaba un error y no hubo ninguno")
	}
	for _, e := range todosLosErrores {
		esperado := slices.Contains(esperados, e)
		if errors.Is(err, e) != esperado {
			t.Errorf("errors.Is(err, %q) = %v, se esperaba %v; error = %q", e, !esperado, esperado, err)
		}
	}

	lineas := strings.Split(err.Error(), "\n")
	if len(lineas) != len(esperados) {
		t.Fatalf("el error tiene %d lineas, se esperaban %d; error = %q", len(lineas), len(esperados), err)
	}
	for i, e := range esperados {
		if !strings.HasPrefix(lineas[i], e.Error()) {
			t.Errorf("error %d = %q, se esperaba que empezara con %q (orden fijo, RN-003-8)", i+1, lineas[i], e.Error())
		}
	}
}

// US-003 / CA-003-1, CA-003-4 y RN-003-1 a RN-003-5, RN-003-10: datos validos
// arman el integrante, con el nombre recortado, el email recortado y en
// minuscula, el rol tal cual y sin ID.
func TestNuevo_Valido(t *testing.T) {
	casos := []struct {
		caso        string
		nombre      string
		email       string
		rol         integrante.Rol
		quedaNombre string
		quedaEmail  string
	}{
		{caso: "US-003 / CA-003-1: datos validos arman el integrante", nombre: nombreOK, email: emailOK, rol: rolOK, quedaNombre: "Ana", quedaEmail: "ana@ejemplo.test"},
		{caso: "US-003 / RN-003-5: rol product_architect", nombre: "Bruno", email: "bruno@ejemplo.test", rol: integrante.RolProductArchitect, quedaNombre: "Bruno", quedaEmail: "bruno@ejemplo.test"},
		{caso: "US-003 / RN-003-5: rol agile_enabler", nombre: "Carla", email: "carla@ejemplo.test", rol: integrante.RolAgileEnabler, quedaNombre: "Carla", quedaEmail: "carla@ejemplo.test"},
		{caso: "US-003 / CL-003-1: nombre de exactamente 100 caracteres", nombre: strings.Repeat("a", 100), email: emailOK, rol: rolOK, quedaNombre: strings.Repeat("a", 100), quedaEmail: emailOK},
		{caso: "US-003 / CL-003-2: cien caracteres no ASCII (200 bytes)", nombre: strings.Repeat("ñ", 100), email: emailOK, rol: rolOK, quedaNombre: strings.Repeat("ñ", 100), quedaEmail: emailOK},
		{caso: "US-003 / CL-003-4: espacios en los extremos del nombre se recortan", nombre: "  Ana  ", email: emailOK, rol: rolOK, quedaNombre: "Ana", quedaEmail: emailOK},
		{caso: "US-003 / RN-003-2: el limite se mide despues de recortar", nombre: "  " + strings.Repeat("a", 100) + "\t", email: emailOK, rol: rolOK, quedaNombre: strings.Repeat("a", 100), quedaEmail: emailOK},
		{caso: "US-003 / CL-003-5: email con espacios y mayusculas", nombre: "Tomás", email: " Tomas@Ejemplo.Test ", rol: rolOK, quedaNombre: "Tomás", quedaEmail: "tomas@ejemplo.test"},
		{caso: "US-003 / RN-003-4: email todo en mayuscula", nombre: nombreOK, email: "ANA@EJEMPLO.TEST", rol: rolOK, quedaNombre: nombreOK, quedaEmail: "ana@ejemplo.test"},
		{caso: "US-003 / RN-003-3: tabulador y salto de linea en los extremos del email", nombre: nombreOK, email: "\tana@ejemplo.test\n", rol: rolOK, quedaNombre: nombreOK, quedaEmail: "ana@ejemplo.test"},
		{caso: "US-003 / limitacion conocida: net/mail acepta un dominio sin punto", nombre: nombreOK, email: "a@b", rol: rolOK, quedaNombre: nombreOK, quedaEmail: "a@b"},
	}

	for _, c := range casos {
		t.Run(c.caso, func(t *testing.T) {
			i, err := integrante.Nuevo(proyectoID, c.nombre, c.email, c.rol)
			if err != nil {
				t.Fatalf("no se esperaba error, se obtuvo %v", err)
			}
			if i.ID != "" {
				t.Errorf("ID = %q, se esperaba vacio: lo asigna la persistencia (RN-003-10)", i.ID)
			}
			if i.ProyectoID != proyectoID {
				t.Errorf("ProyectoID = %q, se esperaba %q", i.ProyectoID, proyectoID)
			}
			if i.Nombre != c.quedaNombre {
				t.Errorf("Nombre = %q, se esperaba %q", i.Nombre, c.quedaNombre)
			}
			if i.Email != c.quedaEmail {
				t.Errorf("Email = %q, se esperaba %q", i.Email, c.quedaEmail)
			}
			if i.Rol != c.rol {
				t.Errorf("Rol = %q, se esperaba %q: el dominio no normaliza el rol", i.Rol, c.rol)
			}
		})
	}
}

// US-003 / CA-003-5, CA-003-6, CA-003-7, RN-003-8 y CL-003-1 a 3, 9 a 11: cada
// dato invalido devuelve su error; con varios, se devuelven todos unidos en el
// orden fijo (nombre, email, rol) y errors.Is reconoce cada uno. Nunca se
// devuelve un Integrante.
func TestNuevo_Invalido(t *testing.T) {
	casos := []struct {
		caso    string
		nombre  string
		email   string
		rol     integrante.Rol
		errores []error
	}{
		// Nombre.
		{caso: "US-003 / RN-003-1: nombre vacio", nombre: "", email: emailOK, rol: rolOK, errores: []error{integrante.ErrNombreVacio}},
		{caso: "US-003 / CL-003-3: nombre con solo espacios", nombre: "   \t ", email: emailOK, rol: rolOK, errores: []error{integrante.ErrNombreVacio}},
		{caso: "US-003 / CL-003-1: nombre de 101 caracteres", nombre: strings.Repeat("a", 101), email: emailOK, rol: rolOK, errores: []error{integrante.ErrNombreLargo}},
		{caso: "US-003 / CL-003-2: 101 caracteres no ASCII", nombre: strings.Repeat("ñ", 101), email: emailOK, rol: rolOK, errores: []error{integrante.ErrNombreLargo}},
		{caso: "US-003 / RN-003-2: 101 caracteres despues de recortar", nombre: "  " + strings.Repeat("a", 101) + " ", email: emailOK, rol: rolOK, errores: []error{integrante.ErrNombreLargo}},

		// Email.
		{caso: "US-003 / RN-003-3: email vacio", nombre: nombreOK, email: "", rol: rolOK, errores: []error{integrante.ErrEmailVacio}},
		{caso: "US-003 / RN-003-3: email con solo espacios se trata como vacio", nombre: nombreOK, email: "   ", rol: rolOK, errores: []error{integrante.ErrEmailVacio}},
		{caso: "US-003 / RN-003-3: email sin arroba", nombre: nombreOK, email: "sin-arroba", rol: rolOK, errores: []error{integrante.ErrEmailInvalido}},
		{caso: "US-003 / CL-003-9: email con nombre visible", nombre: nombreOK, email: "Tomas <tomas@ejemplo.test>", rol: rolOK, errores: []error{integrante.ErrEmailInvalido}},
		{caso: "US-003 / RN-003-3: email entre corchetes", nombre: nombreOK, email: "<tomas@ejemplo.test>", rol: rolOK, errores: []error{integrante.ErrEmailInvalido}},
		{caso: "US-003 / RN-003-3: email sin dominio", nombre: nombreOK, email: "tomas@", rol: rolOK, errores: []error{integrante.ErrEmailInvalido}},
		{caso: "US-003 / RN-003-3: email sin usuario", nombre: nombreOK, email: "@ejemplo.test", rol: rolOK, errores: []error{integrante.ErrEmailInvalido}},
		{caso: "US-003 / RN-003-3: email con un espacio adentro", nombre: nombreOK, email: "to mas@ejemplo.test", rol: rolOK, errores: []error{integrante.ErrEmailInvalido}},
		{caso: "US-003 / RN-003-3: dos direcciones en el email", nombre: nombreOK, email: "a@x.test, b@x.test", rol: rolOK, errores: []error{integrante.ErrEmailInvalido}},

		// Rol.
		{caso: "US-003 / CL-003-10: rol vacio", nombre: nombreOK, email: emailOK, rol: "", errores: []error{integrante.ErrRolInvalido}},
		{caso: "US-003 / RN-003-5: rol fuera de los tres permitidos", nombre: nombreOK, email: emailOK, rol: "jefe", errores: []error{integrante.ErrRolInvalido}},
		{caso: "US-003 / CL-003-10: rol en mayuscula", nombre: nombreOK, email: emailOK, rol: "PRODUCT_BUILDER", errores: []error{integrante.ErrRolInvalido}},
		{caso: "US-003 / CL-003-10: rol con espacios en los extremos", nombre: nombreOK, email: emailOK, rol: " agile_enabler ", errores: []error{integrante.ErrRolInvalido}},

		// Varios errores a la vez, siempre en el orden nombre, email, rol.
		{caso: "US-003 / CL-003-11: nombre vacio, email sin arroba y rol invalido", nombre: "", email: "sin-arroba", rol: "jefe", errores: []error{integrante.ErrNombreVacio, integrante.ErrEmailInvalido, integrante.ErrRolInvalido}},
		{caso: "US-003 / RN-003-8: nombre vacio, email vacio y rol vacio", nombre: "", email: "", rol: "", errores: []error{integrante.ErrNombreVacio, integrante.ErrEmailVacio, integrante.ErrRolInvalido}},
		{caso: "US-003 / RN-003-8: nombre largo y rol invalido", nombre: strings.Repeat("a", 101), email: emailOK, rol: "jefe", errores: []error{integrante.ErrNombreLargo, integrante.ErrRolInvalido}},
		{caso: "US-003 / RN-003-8: email invalido y rol invalido", nombre: nombreOK, email: "sin-arroba", rol: "jefe", errores: []error{integrante.ErrEmailInvalido, integrante.ErrRolInvalido}},
		{caso: "US-003 / RN-003-8: nombre vacio y email invalido", nombre: "", email: "sin-arroba", rol: rolOK, errores: []error{integrante.ErrNombreVacio, integrante.ErrEmailInvalido}},
	}

	for _, c := range casos {
		t.Run(c.caso, func(t *testing.T) {
			i, err := integrante.Nuevo(proyectoID, c.nombre, c.email, c.rol)

			verificarErrores(t, err, c.errores)
			if i != (integrante.Integrante{}) {
				t.Errorf("no se esperaba un Integrante con datos invalidos, se obtuvo %+v", i)
			}
		})
	}
}

// US-003 / seccion 7: ErrEmailInvalido y ErrRolInvalido van envueltos indicando
// el valor recibido, y errors.Is los sigue reconociendo.
func TestNuevo_ErroresIndicanElValorRecibido(t *testing.T) {
	_, err := integrante.Nuevo(proyectoID, nombreOK, "sin-arroba", rolOK)
	if !errors.Is(err, integrante.ErrEmailInvalido) {
		t.Fatalf("errors.Is no reconoce ErrEmailInvalido dentro de %v", err)
	}
	if !strings.Contains(err.Error(), "sin-arroba") {
		t.Errorf("el error %q no indica el email recibido", err)
	}

	_, err = integrante.Nuevo(proyectoID, nombreOK, emailOK, "jefe")
	if !errors.Is(err, integrante.ErrRolInvalido) {
		t.Fatalf("errors.Is no reconoce ErrRolInvalido dentro de %v", err)
	}
	if !strings.Contains(err.Error(), "jefe") {
		t.Errorf("el error %q no indica el rol recibido", err)
	}
}

// US-003 / RN-003-11: no hay maximo de integrantes y varios pueden tener el
// mismo rol: el dominio valida cada integrante por separado.
func TestNuevo_SinLimiteDeIntegrantesNiExclusividadDeRol(t *testing.T) {
	for _, email := range []string{"uno@ejemplo.test", "dos@ejemplo.test", "tres@ejemplo.test"} {
		i, err := integrante.Nuevo(proyectoID, nombreOK, email, integrante.RolProductBuilder)
		if err != nil {
			t.Fatalf("%s: no se esperaba error, se obtuvo %v", email, err)
		}
		if i.Email != email || i.Rol != integrante.RolProductBuilder {
			t.Errorf("%s: se obtuvo %+v", email, i)
		}
	}
}

// US-003 / seccion 7: el mensaje de cada error es el que ve la persona y el que
// comprueban los escenarios BDD con `el mensaje de error indica "..."`.
func TestErrores_Mensajes(t *testing.T) {
	casos := []struct {
		err      error
		esperado string
	}{
		{integrante.ErrNombreVacio, "el nombre es obligatorio"},
		{integrante.ErrNombreLargo, "el nombre supera el largo maximo (100 caracteres)"},
		{integrante.ErrEmailVacio, "el email es obligatorio"},
		{integrante.ErrEmailInvalido, "el email no tiene un formato valido"},
		{integrante.ErrRolInvalido, "el rol tiene que ser product_architect, agile_enabler o product_builder"},
	}

	for _, c := range casos {
		t.Run(c.esperado, func(t *testing.T) {
			if c.err.Error() != c.esperado {
				t.Errorf("mensaje = %q, se esperaba %q", c.err.Error(), c.esperado)
			}
		})
	}
}

// US-003 / RN-003-5: los tres valores del rol son exactos y en minuscula. Son los
// que usan los escenarios BDD y la columna `rol` de la base.
func TestRol_Valores(t *testing.T) {
	casos := []struct {
		rol      integrante.Rol
		esperado string
	}{
		{integrante.RolProductArchitect, "product_architect"},
		{integrante.RolAgileEnabler, "agile_enabler"},
		{integrante.RolProductBuilder, "product_builder"},
	}

	for _, c := range casos {
		t.Run(c.esperado, func(t *testing.T) {
			if string(c.rol) != c.esperado {
				t.Errorf("rol = %q, se esperaba %q", c.rol, c.esperado)
			}
		})
	}
}
