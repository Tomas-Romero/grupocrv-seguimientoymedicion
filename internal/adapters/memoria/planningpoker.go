package memoria

import (
	"context"
	"fmt"
	"slices"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// Base cumple el puerto del caso de uso de US-014 (ExisteProyecto es el de US-005 y
// ObtenerItem, el de US-013).
var _ app.RepositorioPlanningPoker = (*Base)(nil)

// TieneSesionAbierta dice si la historia tiene una sesion que no esta finalizada
// (RN-014-7).
func (b *Base) TieneSesionAbierta(_ context.Context, itemID string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	_, abierta := b.sesionAbierta(itemID)
	return abierta, nil
}

// IDsDeIntegrantes devuelve los IDs de los integrantes del proyecto, en orden de alta.
// Un proyecto sin integrantes, o que no existe, no tiene ninguno.
func (b *Base) IDsDeIntegrantes(_ context.Context, proyectoID string) ([]string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.idsDeIntegrantes(proyectoID), nil
}

// RegistrarSesion replica lo que hace la persistencia (RN-014-11), con mu tomado en
// lugar del candado sobre la historia: verifica que la historia este y sea del proyecto,
// vuelve a armar la sesion con planningpoker.NuevaSesion con lo guardado (su estado, si
// tiene sesion abierta y los integrantes que tiene el proyecto ahora) y la guarda con
// un ID nuevo. Los participantes son los de ese momento, no los de la sesion recibida.
// Es de todo o nada.
func (b *Base) RegistrarSesion(_ context.Context, proyectoID string, sesion planningpoker.SesionPlanningPoker) (planningpoker.SesionPlanningPoker, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	item, ok := b.itemPorID(sesion.ItemID)
	if !ok || item.ProyectoID != proyectoID {
		return planningpoker.SesionPlanningPoker{}, fmt.Errorf("%w (id %s)", app.ErrItemInexistente, sesion.ItemID)
	}
	_, abierta := b.sesionAbierta(sesion.ItemID)
	historia := planningpoker.DatosDeLaHistoria{Completada: item.Estado == backlog.EstadoCompletado, ConSesionAbierta: abierta}
	nueva, err := planningpoker.NuevaSesion(sesion.ItemID, historia, b.idsDeIntegrantes(proyectoID))
	if err != nil {
		return planningpoker.SesionPlanningPoker{}, err
	}

	nueva.ID = b.nuevoID()
	b.sesiones = append(b.sesiones, nueva)
	return copiaDeSesion(nueva), nil
}

// SesionAbiertaDeLaHistoria devuelve la sesion abierta de la historia, si tiene una. Lo
// usan los escenarios para verificar lo que quedo guardado: todavia no hay casos de uso
// de lectura.
func (b *Base) SesionAbiertaDeLaHistoria(itemID string) (planningpoker.SesionPlanningPoker, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.sesionAbierta(itemID)
}

// sesionAbierta busca la sesion de la historia que no esta finalizada. Se llama con mu
// tomado.
func (b *Base) sesionAbierta(itemID string) (planningpoker.SesionPlanningPoker, bool) {
	for _, s := range b.sesiones {
		if s.ItemID == itemID && s.Estado.Abierta() {
			return copiaDeSesion(s), true
		}
	}
	return planningpoker.SesionPlanningPoker{}, false
}

// idsDeIntegrantes devuelve los IDs de los integrantes del proyecto, en orden de alta.
// Se llama con mu tomado.
func (b *Base) idsDeIntegrantes(proyectoID string) []string {
	ids := []string{}
	for _, i := range b.integrantes[proyectoID] {
		ids = append(ids, i.ID)
	}
	return ids
}

// copiaDeSesion devuelve la sesion con su propia lista de participantes.
func copiaDeSesion(s planningpoker.SesionPlanningPoker) planningpoker.SesionPlanningPoker {
	s.Participantes = slices.Clone(s.Participantes)
	return s
}
