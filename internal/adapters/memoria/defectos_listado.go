package memoria

import (
	"context"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
)

// Base cumple el puerto de la lectura de defectos de T-021 (ExisteProyecto es el de US-005).
var _ app.RepositorioListarDefectos = (*Base)(nil)

// ListarDefectosDelProyecto devuelve los defectos del proyecto, de cualquier estado y sin ordenar:
// el orden lo pone el caso de uso (RN-T021-4). Cada uno trae el numero y el titulo de su historia
// y los numeros de sus sprints, traducidos desde los IDs guardados (RN-022-6), como el JOIN de
// Postgres. Un proyecto sin defectos, o que no esta, devuelve una lista vacia.
func (b *Base) ListarDefectosDelProyecto(_ context.Context, proyectoID string) ([]app.DefectoListado, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	listados := []app.DefectoListado{}
	for _, d := range b.defectos[proyectoID] {
		l := app.DefectoListado{Defecto: d, CreadoEn: b.defectosCreadoEn[d.ID]}
		for _, item := range b.items[proyectoID] {
			if item.ID == d.ItemID {
				l.HistoriaNumero, l.HistoriaTitulo = item.Numero, item.Titulo
			}
		}
		for _, s := range b.sprints[proyectoID] {
			if s.ID == d.SprintDeteccionID {
				l.SprintDeteccion = s.Numero
			}
			if d.SprintResolucionID != "" && s.ID == d.SprintResolucionID {
				l.SprintResolucion = s.Numero
			}
		}
		listados = append(listados, l)
	}
	return listados, nil
}
