package httpadapter

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/errores"
)

var (
	formatoEntero  = regexp.MustCompile(`^[+-]?[0-9]+$`)
	formatoDecimal = regexp.MustCompile(`^[+-]?[0-9]+([.,][0-9]+)?$`)
)

// Formulario convierte los valores de un formulario en numeros y fechas, y junta los
// campos mal formados para informarlos todos a la vez (spec T-005).
//
// Solo revisa el formato (RN-T005-10): un campo que no vino o vino vacio da el cero de su
// tipo y sin error, para que el dominio diga "es obligatorio" con su propio mensaje.
type Formulario struct {
	valores url.Values
	errores []errores.ErrorDeCampo
}

// NuevoFormulario arma el formulario sobre los valores de la peticion (r.PostForm o r.Form).
func NuevoFormulario(valores url.Values) *Formulario {
	return &Formulario{valores: valores}
}

// Entero lee un numero entero: digitos con signo opcional, sin espacios en el medio.
func (f *Formulario) Entero(campo string) int {
	texto := strings.TrimSpace(f.valores.Get(campo))
	if texto == "" {
		return 0
	}
	if !formatoEntero.MatchString(texto) {
		f.rechazar(campo, "un numero entero")
		return 0
	}
	n, err := strconv.Atoi(texto)
	if err != nil {
		f.rechazar(campo, "un numero entero")
		return 0
	}
	return n
}

// Decimal lee un numero con parte decimal opcional, con coma o con punto (igual que
// US-019). No acepta notacion cientifica, NaN ni infinitos.
func (f *Formulario) Decimal(campo string) float64 {
	texto := strings.TrimSpace(f.valores.Get(campo))
	if texto == "" {
		return 0
	}
	if !formatoDecimal.MatchString(texto) {
		f.rechazar(campo, "un numero")
		return 0
	}
	n, err := strconv.ParseFloat(strings.Replace(texto, ",", ".", 1), 64)
	if err != nil {
		f.rechazar(campo, "un numero")
		return 0
	}
	return n
}

// Fecha lee una fecha AAAA-MM-DD, el formato de un <input type="date">, a medianoche UTC.
func (f *Formulario) Fecha(campo string) time.Time {
	texto := strings.TrimSpace(f.valores.Get(campo))
	if texto == "" {
		return time.Time{}
	}
	fecha, err := time.Parse(time.DateOnly, texto)
	if err != nil {
		f.rechazar(campo, "una fecha con formato AAAA-MM-DD")
		return time.Time{}
	}
	return fecha
}

// Err devuelve los campos mal formados, en el orden en que se consultaron, o nil si no
// hubo ninguno. Se puede unir con los errores del dominio con errors.Join.
func (f *Formulario) Err() error {
	if len(f.errores) == 0 {
		return nil
	}
	return &errores.ErrorDeValidacion{Errores: append([]errores.ErrorDeCampo(nil), f.errores...)}
}

// rechazar anota el campo como mal formado, una sola vez (RN-T005-12).
func (f *Formulario) rechazar(campo, esperado string) {
	for _, e := range f.errores {
		if e.Campo == campo {
			return
		}
	}
	f.errores = append(f.errores, errores.ErrorDeCampo{
		Campo:   campo,
		Mensaje: "el campo " + strings.ReplaceAll(campo, "_", " ") + " tiene que ser " + esperado,
	})
}
