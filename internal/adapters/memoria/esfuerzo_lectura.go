package memoria

import (
	"context"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
)

// Base cumple el puerto del caso de uso de US-020.
var _ app.RepositorioConsultarEsfuerzo = (*Base)(nil)

// LeerEsfuerzo devuelve los registros, los titulos de las historias, los nombres de los
// integrantes y los sprints del proyecto, todo copiado: quien llama puede modificarlo sin tocar
// la base. Un proyecto que no existe o no tiene nada devuelve una lectura vacia.
func (b *Base) LeerEsfuerzo(_ context.Context, proyectoID string) (app.LecturaEsfuerzo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	lectura := app.LecturaEsfuerzo{
		Registros:   slices.Clone(b.esfuerzos[proyectoID]),
		Historias:   map[string]string{},
		Integrantes: map[string]string{},
		Sprints:     slices.Clone(b.sprints[proyectoID]),
	}
	for _, item := range b.items[proyectoID] {
		lectura.Historias[item.ID] = item.Titulo
	}
	for _, i := range b.integrantes[proyectoID] {
		lectura.Integrantes[i.ID] = i.Nombre
	}
	return lectura, nil
}
