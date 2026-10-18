package memoria

import (
	"context"
	"fmt"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
)

// Base cumple el puerto del caso de uso de US-019 (ExisteProyecto es el mismo que
// usa el de US-005).
var _ app.RepositorioEsfuerzo = (*Base)(nil)

// ExisteItemEnProyecto dice si la historia esta en ese proyecto (RN-019-6). Una
// historia de otro proyecto, un ID que no esta o uno mal formado no cuentan.
func (b *Base) ExisteItemEnProyecto(_ context.Context, proyectoID, itemID string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.itemEnProyecto(proyectoID, itemID), nil
}

// ExisteIntegranteEnProyecto dice si el integrante esta en ese proyecto
// (RN-019-6). Uno de otro proyecto, un ID que no esta o uno mal formado no cuentan.
func (b *Base) ExisteIntegranteEnProyecto(_ context.Context, proyectoID, integranteID string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.integranteEnProyecto(proyectoID, integranteID), nil
}

// RegistrarEsfuerzo guarda el registro con un ID nuevo, como el RETURNING de
// Postgres (RN-019-10). Replica las claves foraneas de la tabla: si el proyecto,
// la historia o el integrante no estan, devuelve el error de inexistencia que
// corresponde y no guarda nada. Registros iguales son registros distintos, y no
// mira el estado de la historia (RN-019-7) ni las horas del dia.
func (b *Base) RegistrarEsfuerzo(_ context.Context, r esfuerzo.RegistroEsfuerzo) (esfuerzo.RegistroEsfuerzo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch {
	case !b.proyectos[r.ProyectoID]:
		return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("%w (id %s)", app.ErrProyectoInexistente, r.ProyectoID)
	case !b.itemExiste(r.ItemID):
		return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("%w (id %s)", app.ErrItemInexistente, r.ItemID)
	case !b.integranteExiste(r.IntegranteID):
		return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("%w (id %s)", app.ErrIntegranteInexistente, r.IntegranteID)
	}

	r.ID = b.nuevoID()
	b.esfuerzos[r.ProyectoID] = append(b.esfuerzos[r.ProyectoID], r)
	return r, nil
}

// Esfuerzos devuelve una copia de los registros de esfuerzo del proyecto, en orden
// de alta. Un proyecto sin registros, o que no existe, devuelve una lista vacia. Lo
// usan los escenarios para verificar lo que quedo guardado: todavia no hay casos de
// uso de lectura (US-020).
func (b *Base) Esfuerzos(proyectoID string) []esfuerzo.RegistroEsfuerzo {
	b.mu.Lock()
	defer b.mu.Unlock()

	return slices.Clone(b.esfuerzos[proyectoID])
}

// itemEnProyecto se llama con mu tomado.
func (b *Base) itemEnProyecto(proyectoID, itemID string) bool {
	for _, item := range b.items[proyectoID] {
		if item.ID == itemID {
			return true
		}
	}
	return false
}

// itemExiste busca la historia en todos los proyectos. Se llama con mu tomado.
func (b *Base) itemExiste(itemID string) bool {
	for proyectoID := range b.items {
		if b.itemEnProyecto(proyectoID, itemID) {
			return true
		}
	}
	return false
}

// integranteEnProyecto se llama con mu tomado.
func (b *Base) integranteEnProyecto(proyectoID, integranteID string) bool {
	for _, i := range b.integrantes[proyectoID] {
		if i.ID == integranteID {
			return true
		}
	}
	return false
}

// integranteExiste busca al integrante en todos los proyectos. Se llama con mu
// tomado.
func (b *Base) integranteExiste(integranteID string) bool {
	for proyectoID := range b.integrantes {
		if b.integranteEnProyecto(proyectoID, integranteID) {
			return true
		}
	}
	return false
}
