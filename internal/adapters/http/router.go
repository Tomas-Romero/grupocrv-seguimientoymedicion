package httpadapter

import (
	"io/fs"
	"net/http"
	"time"

	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// NuevoRouter arma las rutas de la aplicacion (spec T-004). Vive aca y no en main.go para
// poder probarlas sin levantar el servidor. Usa el ServeMux de la biblioteca estandar.
//
//	GET /health       estado de la base (JSON)
//	GET /static/...   archivos estaticos, sin listar directorios
//	GET /             inicio
//	/proyectos...     las pantallas, si se pasa ConPantallas
//	cualquier otra    404 con la pagina de error
func NuevoRouter(salud Verificador, espera time.Duration, estaticos http.FileSystem, opciones ...OpcionDeRouter) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", Health(salud, espera))
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(sinDirectorios{estaticos})))
	mux.HandleFunc("GET /{$}", inicio)
	for _, opcion := range opciones {
		opcion(mux)
	}
	mux.HandleFunc("/", noEncontrado)
	return mux
}

// OpcionDeRouter agrega rutas al router antes de la ruta que responde 404.
type OpcionDeRouter func(mux *http.ServeMux)

// ConPantallas agrega las pantallas de proyectos, integrantes y backlog (spec T-011). Sin
// esta opcion el router tiene solo el inicio, /health y los estaticos.
func ConPantallas(p *Pantallas) OpcionDeRouter {
	return p.registrar
}

func inicio(w http.ResponseWriter, r *http.Request) {
	Responder(w, r, http.StatusOK, plantillas.Pagina{Titulo: "Inicio", Activo: plantillas.SeccionInicio}, plantillas.Inicio())
}

// noEncontrado responde cualquier ruta que no esta registrada. La pagina de error no marca
// ningun item del menu (RN-T004-4) y el codigo es 404, no 200 (RN-T004-5).
func noEncontrado(w http.ResponseWriter, r *http.Request) {
	Responder(w, r, http.StatusNotFound, plantillas.Pagina{Titulo: "Página no encontrada"}, plantillas.NoEncontrado())
}

// sinDirectorios hace que un directorio no exista: FileServer lista el contenido de los
// directorios que encuentra, y /static/ no tiene que mostrar que archivos hay (RN-T004-6).
type sinDirectorios struct {
	fs http.FileSystem
}

func (s sinDirectorios) Open(nombre string) (http.File, error) {
	archivo, err := s.fs.Open(nombre)
	if err != nil {
		return nil, err
	}
	info, err := archivo.Stat()
	if err != nil {
		_ = archivo.Close()
		return nil, err
	}
	if info.IsDir() {
		_ = archivo.Close()
		return nil, fs.ErrNotExist
	}
	return archivo, nil
}
