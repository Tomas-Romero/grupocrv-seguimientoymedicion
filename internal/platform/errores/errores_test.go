package errores_test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/errores"
)

// Errores de prueba: hacen el papel de los del dominio y la aplicacion, que este
// paquete no conoce (spec T-005, seccion 5).
var (
	errTituloVacio  = errors.New("el titulo es obligatorio")
	errTituloLargo  = errors.New("el titulo no puede tener mas de 120 caracteres")
	errPrioridad    = errors.New("la prioridad tiene que ser must, should, could o wont")
	errInexistente  = errors.New("el item no existe")
	errSprintCerr   = errors.New("el sprint esta cerrado")
	errYaAsignado   = errors.New("la historia ya esta asignada a un sprint")
	errSinCampo     = errors.New("la historia y el sprint son de proyectos distintos")
	errDesconocido  = errors.New("connection refused: 10.0.0.5:5432")
	errTambienFuera = errors.New("otro error que nadie registro")
)

func catalogoDePrueba() *errores.Catalogo {
	return errores.NuevoCatalogo(
		errores.Regla{Err: errTituloVacio, Tipo: errores.Validacion, Campo: "titulo"},
		errores.Regla{Err: errTituloLargo, Tipo: errores.Validacion, Campo: "titulo"},
		errores.Regla{Err: errPrioridad, Tipo: errores.Validacion, Campo: "prioridad"},
		errores.Regla{Err: errInexistente, Tipo: errores.NoEncontrado},
		errores.Regla{Err: errSprintCerr, Tipo: errores.Conflicto},
		errores.Regla{Err: errYaAsignado, Tipo: errores.Conflicto},
		errores.Regla{Err: errSinCampo, Tipo: errores.Conflicto},
	)
}

func campo(c, m string) errores.ErrorDeCampo { return errores.ErrorDeCampo{Campo: c, Mensaje: m} }

// RN-T005-3, RN-T005-8: un error del catalogo se reconoce y trae su tipo, su campo y su mensaje.
func TestClasificar_ErrorDelCatalogo(t *testing.T) {
	casos := []struct {
		nombre string
		err    error
		quiero errores.Problema
	}{
		{"validacion con campo", errTituloVacio,
			errores.Problema{Tipo: errores.Validacion, Errores: []errores.ErrorDeCampo{campo("titulo", "el titulo es obligatorio")}}},
		{"no encontrado sin campo", errInexistente,
			errores.Problema{Tipo: errores.NoEncontrado, Errores: []errores.ErrorDeCampo{campo("", "el item no existe")}}},
		{"conflicto sin campo", errSprintCerr,
			errores.Problema{Tipo: errores.Conflicto, Errores: []errores.ErrorDeCampo{campo("", "el sprint esta cerrado")}}},
	}
	cat := catalogoDePrueba()
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := cat.Clasificar(c.err); !reflect.DeepEqual(got, c.quiero) {
				t.Errorf("Clasificar() = %+v, quiero %+v", got, c.quiero)
			}
		})
	}
}

// CL-T005-1: sin error no hay problema.
func TestClasificar_Nil(t *testing.T) {
	got := catalogoDePrueba().Clasificar(nil)
	if got.Tipo != "" || len(got.Errores) != 0 {
		t.Errorf("Clasificar(nil) = %+v, quiero un Problema vacio", got)
	}
}

// RN-T005-5, CL-T005-3: el mensaje es el del catalogo, no el de las capas que envolvieron al error.
func TestClasificar_ErrorEnvuelto(t *testing.T) {
	err := fmt.Errorf("asignar el item 7 al sprint 3: %w", fmt.Errorf("%w (id 7)", errInexistente))
	got := catalogoDePrueba().Clasificar(err)

	quiero := errores.Problema{Tipo: errores.NoEncontrado, Errores: []errores.ErrorDeCampo{campo("", "el item no existe")}}
	if !reflect.DeepEqual(got, quiero) {
		t.Errorf("Clasificar() = %+v, quiero %+v", got, quiero)
	}
}

// RN-T005-4, CA-T005-2: los errores unidos salen todos, en el orden en que vinieron.
func TestClasificar_ErroresUnidos(t *testing.T) {
	err := errors.Join(errPrioridad, errTituloVacio)
	got := catalogoDePrueba().Clasificar(err)

	quiero := errores.Problema{Tipo: errores.Validacion, Errores: []errores.ErrorDeCampo{
		campo("prioridad", "la prioridad tiene que ser must, should, could o wont"),
		campo("titulo", "el titulo es obligatorio"),
	}}
	if !reflect.DeepEqual(got, quiero) {
		t.Errorf("Clasificar() = %+v, quiero %+v", got, quiero)
	}
}

// CL-T005-2: dos errores del mismo campo se conservan los dos.
func TestClasificar_DosErroresDelMismoCampo(t *testing.T) {
	got := catalogoDePrueba().Clasificar(errors.Join(errTituloVacio, errTituloLargo))

	if len(got.Errores) != 2 || got.Errores[0].Campo != "titulo" || got.Errores[1].Campo != "titulo" {
		t.Errorf("Clasificar() = %+v, quiero dos errores del campo titulo", got)
	}
}

// CL-T005-4: una union dentro de otra se aplana y mantiene el orden de lectura.
func TestClasificar_UnionAnidada(t *testing.T) {
	err := errors.Join(errTituloVacio, fmt.Errorf("contexto: %w", errors.Join(errPrioridad, errTituloLargo)))
	got := catalogoDePrueba().Clasificar(err)

	var mensajes []string
	for _, e := range got.Errores {
		mensajes = append(mensajes, e.Mensaje)
	}
	quiero := []string{errTituloVacio.Error(), errPrioridad.Error(), errTituloLargo.Error()}
	if !reflect.DeepEqual(mensajes, quiero) {
		t.Errorf("mensajes = %v, quiero %v", mensajes, quiero)
	}
}

// RN-T005-4: fmt.Errorf con varios %w tambien se separa en sus partes.
func TestClasificar_VariosWrap(t *testing.T) {
	err := fmt.Errorf("%w y %w", errTituloVacio, errPrioridad)
	got := catalogoDePrueba().Clasificar(err)

	if got.Tipo != errores.Validacion || len(got.Errores) != 2 {
		t.Errorf("Clasificar() = %+v, quiero validacion con dos errores", got)
	}
}

// RN-T005-7: lo que no esta en el catalogo es interno, con el mensaje fijo y nada del error original.
func TestClasificar_ErrorDesconocido(t *testing.T) {
	got := catalogoDePrueba().Clasificar(fmt.Errorf("leer el item: %w", errDesconocido))

	quiero := errores.Problema{Tipo: errores.Interno, Errores: []errores.ErrorDeCampo{campo("", errores.MensajeInterno)}}
	if !reflect.DeepEqual(got, quiero) {
		t.Errorf("Clasificar() = %+v, quiero %+v", got, quiero)
	}
}

func TestMensajeInterno(t *testing.T) {
	const quiero = "ocurrio un error inesperado: proba de nuevo en unos minutos"
	if errores.MensajeInterno != quiero {
		t.Errorf("MensajeInterno = %q, quiero %q", errores.MensajeInterno, quiero)
	}
}

// RN-T005-6, CL-T005-5: con una parte interna no se muestra ninguna de las otras.
func TestClasificar_ValidacionConDesconocido(t *testing.T) {
	got := catalogoDePrueba().Clasificar(errors.Join(errTituloVacio, errDesconocido))

	quiero := errores.Problema{Tipo: errores.Interno, Errores: []errores.ErrorDeCampo{campo("", errores.MensajeInterno)}}
	if !reflect.DeepEqual(got, quiero) {
		t.Errorf("Clasificar() = %+v, quiero %+v", got, quiero)
	}
}

// RN-T005-6: el tipo es el mas grave de las partes: interno, no_encontrado, conflicto, validacion.
func TestClasificar_TipoMasGrave(t *testing.T) {
	casos := []struct {
		nombre string
		err    error
		quiero errores.Tipo
	}{
		{"validacion y conflicto", errors.Join(errTituloVacio, errSprintCerr), errores.Conflicto},
		{"conflicto y validacion", errors.Join(errSprintCerr, errTituloVacio), errores.Conflicto},
		{"validacion y no encontrado", errors.Join(errTituloVacio, errInexistente), errores.NoEncontrado},
		{"conflicto y no encontrado", errors.Join(errSprintCerr, errInexistente), errores.NoEncontrado},
		{"todo junto con un desconocido", errors.Join(errTituloVacio, errSprintCerr, errInexistente, errTambienFuera), errores.Interno},
		{"dos conflictos", errors.Join(errSprintCerr, errYaAsignado), errores.Conflicto},
	}
	cat := catalogoDePrueba()
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := cat.Clasificar(c.err).Tipo; got != c.quiero {
				t.Errorf("Tipo = %q, quiero %q", got, c.quiero)
			}
		})
	}
}

// CL-T005-6: validacion y conflicto juntos: tipo conflicto con los mensajes de los dos, en orden.
func TestClasificar_ValidacionYConflictoConservaMensajes(t *testing.T) {
	got := catalogoDePrueba().Clasificar(errors.Join(errTituloVacio, errSprintCerr))

	quiero := []errores.ErrorDeCampo{campo("titulo", "el titulo es obligatorio"), campo("", "el sprint esta cerrado")}
	if !reflect.DeepEqual(got.Errores, quiero) {
		t.Errorf("Errores = %+v, quiero %+v", got.Errores, quiero)
	}
}

// RN-T005-9: un ErrorDeValidacion (lo que arma el formulario) se clasifica como validacion
// y conserva sus campos y mensajes.
func TestClasificar_ErrorDeValidacion(t *testing.T) {
	err := &errores.ErrorDeValidacion{Errores: []errores.ErrorDeCampo{
		campo("horas", "el campo horas tiene que ser un numero"),
		campo("fecha", "el campo fecha tiene que ser una fecha con formato AAAA-MM-DD"),
	}}
	got := catalogoDePrueba().Clasificar(fmt.Errorf("leer el formulario: %w", err))

	quiero := errores.Problema{Tipo: errores.Validacion, Errores: err.Errores}
	if !reflect.DeepEqual(got, quiero) {
		t.Errorf("Clasificar() = %+v, quiero %+v", got, quiero)
	}
}

// CL-T005-7: los campos del formulario y los del dominio salen juntos, en el orden de la union.
func TestClasificar_ErrorDeValidacionUnidoConDominio(t *testing.T) {
	deFormulario := &errores.ErrorDeValidacion{Errores: []errores.ErrorDeCampo{campo("horas", "el campo horas tiene que ser un numero")}}
	got := catalogoDePrueba().Clasificar(errors.Join(deFormulario, errTituloVacio))

	quiero := errores.Problema{Tipo: errores.Validacion, Errores: []errores.ErrorDeCampo{
		campo("horas", "el campo horas tiene que ser un numero"),
		campo("titulo", "el titulo es obligatorio"),
	}}
	if !reflect.DeepEqual(got, quiero) {
		t.Errorf("Clasificar() = %+v, quiero %+v", got, quiero)
	}
}

// El Problema no comparte memoria con el ErrorDeValidacion que lo origino.
func TestClasificar_ErrorDeValidacionNoComparteMemoria(t *testing.T) {
	err := &errores.ErrorDeValidacion{Errores: []errores.ErrorDeCampo{campo("horas", "el campo horas tiene que ser un numero")}}
	got := catalogoDePrueba().Clasificar(err)

	got.Errores[0].Mensaje = "cambiado"
	if err.Errores[0].Mensaje != "el campo horas tiene que ser un numero" {
		t.Error("modificar el Problema modifico el ErrorDeValidacion")
	}
}

func TestErrorDeValidacion_Error(t *testing.T) {
	err := &errores.ErrorDeValidacion{Errores: []errores.ErrorDeCampo{
		campo("horas", "el campo horas tiene que ser un numero"),
		campo("fecha", "el campo fecha tiene que ser una fecha con formato AAAA-MM-DD"),
	}}
	const quiero = "horas: el campo horas tiene que ser un numero; fecha: el campo fecha tiene que ser una fecha con formato AAAA-MM-DD"
	if got := err.Error(); got != quiero {
		t.Errorf("Error() = %q, quiero %q", got, quiero)
	}
}

// Una regla que se repite para el mismo error: gana la primera.
func TestClasificar_GanaLaPrimeraRegla(t *testing.T) {
	cat := errores.NuevoCatalogo(
		errores.Regla{Err: errTituloVacio, Tipo: errores.Validacion, Campo: "titulo"},
		errores.Regla{Err: errTituloVacio, Tipo: errores.Conflicto, Campo: "otro"},
	)
	got := cat.Clasificar(errTituloVacio)
	if got.Tipo != errores.Validacion || got.Errores[0].Campo != "titulo" {
		t.Errorf("Clasificar() = %+v, quiero la primera regla", got)
	}
}
