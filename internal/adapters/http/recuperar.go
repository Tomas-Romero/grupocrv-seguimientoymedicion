package httpadapter

import (
	"errors"
	"net/http"
	"runtime/debug"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/platform/errores"
)

// seguimiento recuerda si el handler ya empezo a escribir la respuesta: despues de eso
// ya no se puede cambiar el estado ni el cuerpo.
type seguimiento struct {
	http.ResponseWriter
	empezo bool
}

func (s *seguimiento) WriteHeader(codigo int) {
	s.empezo = true
	s.ResponseWriter.WriteHeader(codigo)
}

func (s *seguimiento) Write(datos []byte) (int, error) {
	s.empezo = true
	return s.ResponseWriter.Write(datos)
}

// Unwrap deja que http.ResponseController llegue al escritor original (Flush, etc.).
func (s *seguimiento) Unwrap() http.ResponseWriter { return s.ResponseWriter }

// Recuperar es un middleware: si el handler que sigue hace panic, lo registra con su
// valor y la pila, responde un error interno uniforme y el servidor sigue atendiendo
// (RN-T005-18). Si la respuesta ya habia empezado, solo registra. http.ErrAbortHandler
// se vuelve a lanzar, porque es la forma legitima de cortar una respuesta.
func (m *ManejadorDeErrores) Recuperar(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		escritor := &seguimiento{ResponseWriter: w}
		defer func() {
			valor := recover()
			if valor == nil {
				return
			}
			if err, ok := valor.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				panic(valor)
			}
			m.log.Error("panic en el handler",
				"method", r.Method, "path", r.URL.Path, "panic", valor, "stack", string(debug.Stack()))
			if escritor.empezo {
				return
			}
			escribirProblema(w, r, errores.Problema{
				Tipo:    errores.Interno,
				Errores: []errores.ErrorDeCampo{{Mensaje: errores.MensajeInterno}},
			})
		}()
		siguiente.ServeHTTP(escritor, r)
	})
}
