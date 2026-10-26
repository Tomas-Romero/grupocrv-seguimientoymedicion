package memoria

import (
	"context"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
)

// Base cumple el puerto del caso de uso de US-021.
var _ app.RepositorioCompararEsfuerzo = (*Base)(nil)

// LeerHistoriasConHoras devuelve todas las historias del proyecto, en orden de alta, cada una con
// las horas de todos sus registros de esfuerzo, en orden de alta. Todo es una copia: quien llama
// puede modificarlo sin tocar la base. Un proyecto sin historias, o que no existe, devuelve una
// lista vacia.
func (b *Base) LeerHistoriasConHoras(_ context.Context, proyectoID string) ([]app.HistoriaConHoras, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	horas := map[string][]float64{}
	for _, r := range b.esfuerzos[proyectoID] {
		horas[r.ItemID] = append(horas[r.ItemID], r.Horas)
	}

	historias := make([]app.HistoriaConHoras, 0, len(b.items[proyectoID]))
	for _, item := range b.items[proyectoID] {
		historias = append(historias, app.HistoriaConHoras{Item: item, Horas: horas[item.ID]})
	}
	return historias, nil
}
