package memoria_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
)

// nuevoIntegranteDePrueba arma un integrante como lo entrega el dominio: ya
// validado, con el email en minuscula y sin ID.
func nuevoIntegranteDePrueba(proyectoID, email string) integrante.Integrante {
	return integrante.Integrante{
		ProyectoID: proyectoID,
		Nombre:     "Ana",
		Email:      email,
		Rol:        integrante.RolProductBuilder,
	}
}

// US-003 / CA-003-1, CL-003-12, RN-003-10: registrar guarda al integrante y lo
// devuelve con un ID generado, como el RETURNING de Postgres.
func TestBase_RegistrarIntegrante_Guarda(t *testing.T) {
	base := memoria.NuevaBase()
	proyectoID := base.AgregarProyecto()

	registrado, err := base.RegistrarIntegrante(context.Background(), nuevoIntegranteDePrueba(proyectoID, "ana@ejemplo.test"))
	if err != nil {
		t.Fatalf("RegistrarIntegrante: %v", err)
	}

	if registrado.ID == "" {
		t.Error("el integrante registrado no tiene ID: lo genera la base (RN-003-10)")
	}
	if registrado.ProyectoID != proyectoID || registrado.Nombre != "Ana" ||
		registrado.Email != "ana@ejemplo.test" || registrado.Rol != integrante.RolProductBuilder {
		t.Errorf("integrante devuelto = %+v, se esperaban los datos recibidos", registrado)
	}

	guardados := base.Integrantes(proyectoID)
	if len(guardados) != 1 || guardados[0] != registrado {
		t.Errorf("integrantes del proyecto = %+v, se esperaba solo %+v", guardados, registrado)
	}
}

// US-003 / CA-003-2, CL-003-8, RN-003-11: cada alta recibe un ID distinto, y se
// pueden registrar varios integrantes, con el mismo nombre y con el mismo rol.
func TestBase_RegistrarIntegrante_VariosIntegrantes(t *testing.T) {
	base := memoria.NuevaBase()
	proyectoID := base.AgregarProyecto()

	ids := map[string]bool{}
	for _, email := range []string{"uno@ejemplo.test", "dos@ejemplo.test", "tres@ejemplo.test"} {
		registrado, err := base.RegistrarIntegrante(context.Background(), nuevoIntegranteDePrueba(proyectoID, email))
		if err != nil {
			t.Fatalf("RegistrarIntegrante(%s): %v", email, err)
		}
		if ids[registrado.ID] {
			t.Errorf("el ID %q se repite", registrado.ID)
		}
		ids[registrado.ID] = true
	}

	if cantidad := len(base.Integrantes(proyectoID)); cantidad != 3 {
		t.Errorf("integrantes del proyecto = %d, se esperaban 3", cantidad)
	}
}

// US-003 / CA-003-8, CL-003-6, RN-003-7: el email repetido en el mismo proyecto se
// rechaza con ErrEmailDuplicado sin importar las mayusculas, y no se guarda nada.
func TestBase_RegistrarIntegrante_EmailDuplicado(t *testing.T) {
	casos := []struct {
		nombre string
		email  string
	}{
		{"el mismo email", "tomas@ejemplo.test"},
		{"otra capitalizacion", "TOMAS@ejemplo.test"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			base := memoria.NuevaBase()
			proyectoID := base.AgregarProyecto()
			if _, err := base.RegistrarIntegrante(context.Background(), nuevoIntegranteDePrueba(proyectoID, "tomas@ejemplo.test")); err != nil {
				t.Fatalf("primer alta: %v", err)
			}

			_, err := base.RegistrarIntegrante(context.Background(), nuevoIntegranteDePrueba(proyectoID, c.email))
			if !errors.Is(err, app.ErrEmailDuplicado) {
				t.Fatalf("error = %v, se esperaba ErrEmailDuplicado", err)
			}
			// Envuelto indicando el email recibido.
			if !strings.Contains(err.Error(), c.email) {
				t.Errorf("el mensaje %q no indica el email", err.Error())
			}
			if cantidad := len(base.Integrantes(proyectoID)); cantidad != 1 {
				t.Errorf("integrantes del proyecto = %d, se esperaba 1: el rechazo no guarda nada (RN-003-9)", cantidad)
			}
		})
	}
}

// US-003 / CA-003-3, CL-003-7: el mismo email es valido en otro proyecto.
func TestBase_RegistrarIntegrante_MismoEmailEnOtroProyecto(t *testing.T) {
	base := memoria.NuevaBase()
	demo := base.AgregarProyecto()
	otro := base.AgregarProyecto()

	for _, proyectoID := range []string{demo, otro} {
		if _, err := base.RegistrarIntegrante(context.Background(), nuevoIntegranteDePrueba(proyectoID, "ana@ejemplo.test")); err != nil {
			t.Fatalf("RegistrarIntegrante en %s: %v", proyectoID, err)
		}
	}

	for _, proyectoID := range []string{demo, otro} {
		if cantidad := len(base.Integrantes(proyectoID)); cantidad != 1 {
			t.Errorf("integrantes de %s = %d, se esperaba 1", proyectoID, cantidad)
		}
	}
}

// US-003 / CA-003-9, RN-003-9: si el proyecto no existe devuelve
// ErrProyectoInexistente, incluso con un ID vacio o mal formado, y no guarda nada.
func TestBase_RegistrarIntegrante_ProyectoInexistente(t *testing.T) {
	for _, proyectoID := range []string{"00000000-0000-0000-0000-000000000000", "", "no-es-un-uuid"} {
		t.Run("id "+proyectoID, func(t *testing.T) {
			base := memoria.NuevaBase()

			_, err := base.RegistrarIntegrante(context.Background(), nuevoIntegranteDePrueba(proyectoID, "ana@ejemplo.test"))
			if !errors.Is(err, app.ErrProyectoInexistente) {
				t.Fatalf("error = %v, se esperaba ErrProyectoInexistente", err)
			}
			if cantidad := len(base.Integrantes(proyectoID)); cantidad != 0 {
				t.Errorf("integrantes = %d, se esperaba ninguno", cantidad)
			}
		})
	}
}

// US-003 / CL-003-12: un proyecto sin integrantes, o que no existe, no tiene
// ninguno. Lo que devuelve Integrantes es una copia: modificarla no cambia la base.
func TestBase_Integrantes_SinIntegrantesYCopia(t *testing.T) {
	base := memoria.NuevaBase()
	proyectoID := base.AgregarProyecto()

	if cantidad := len(base.Integrantes(proyectoID)); cantidad != 0 {
		t.Errorf("integrantes de un proyecto nuevo = %d, se esperaba 0", cantidad)
	}
	if cantidad := len(base.Integrantes("00000000-0000-0000-0000-000000000000")); cantidad != 0 {
		t.Errorf("integrantes de un proyecto inexistente = %d, se esperaba 0", cantidad)
	}

	if _, err := base.RegistrarIntegrante(context.Background(), nuevoIntegranteDePrueba(proyectoID, "ana@ejemplo.test")); err != nil {
		t.Fatalf("RegistrarIntegrante: %v", err)
	}
	copia := base.Integrantes(proyectoID)
	if len(copia) != 1 {
		t.Fatalf("integrantes del proyecto = %d, se esperaba 1", len(copia))
	}
	copia[0].Nombre = "Modificado"

	if nombre := base.Integrantes(proyectoID)[0].Nombre; nombre != "Ana" {
		t.Errorf("nombre guardado = %q, se esperaba %q: Integrantes tiene que devolver una copia", nombre, "Ana")
	}
}

// US-003 / CL-003-14, RN-003-9: con varias altas simultaneas del mismo email en
// el mismo proyecto, una termina bien y las demas devuelven ErrEmailDuplicado.
func TestBase_RegistrarIntegrante_AltasSimultaneas(t *testing.T) {
	const altas = 20
	base := memoria.NuevaBase()
	proyectoID := base.AgregarProyecto()

	errs := make([]error, altas)
	var espera sync.WaitGroup
	for n := 0; n < altas; n++ {
		espera.Add(1)
		go func(n int) {
			defer espera.Done()
			_, errs[n] = base.RegistrarIntegrante(context.Background(), nuevoIntegranteDePrueba(proyectoID, "ana@ejemplo.test"))
		}(n)
	}
	espera.Wait()

	exitosas := 0
	for _, err := range errs {
		switch {
		case err == nil:
			exitosas++
		case !errors.Is(err, app.ErrEmailDuplicado):
			t.Errorf("error inesperado: %v", err)
		}
	}
	if exitosas != 1 {
		t.Errorf("altas exitosas = %d, se esperaba 1", exitosas)
	}
	if cantidad := len(base.Integrantes(proyectoID)); cantidad != 1 {
		t.Errorf("integrantes del proyecto = %d, se esperaba 1", cantidad)
	}
}
