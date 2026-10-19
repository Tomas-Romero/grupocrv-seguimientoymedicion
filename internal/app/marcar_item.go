package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/sprint"
)

// CambioDeEstado es lo que marcar o desmarcar le hace al item: ItemBacklog.Completar o
// ItemBacklog.Reabrir. El repositorio lo aplica al item que lee con el sprint
// bloqueado, no al que leyo el caso de uso: asi, si otro marcado llego antes, el
// segundo ve el item ya completado y conserva la fecha del primero (CL-010-10).
type CambioDeEstado func(backlog.ItemBacklog) backlog.ItemBacklog

// RepositorioMarcado es lo que los casos de uso de US-010 necesitan de la persistencia.
type RepositorioMarcado interface {
	// ObtenerItem devuelve el item y si existe. Un ID mal formado es un item
	// inexistente, no un error.
	ObtenerItem(ctx context.Context, itemID string) (backlog.ItemBacklog, bool, error)

	// ObtenerSprint devuelve el sprint y si existe. Un ID mal formado es un sprint
	// inexistente, no un error.
	ObtenerSprint(ctx context.Context, sprintID string) (sprint.Sprint, bool, error)

	// ItemEnSprint dice si el item esta en ese sprint: tiene una fila de la relacion
	// pendiente en el (RN-010-2). Haber estado en otro sprint no cuenta.
	ItemEnSprint(ctx context.Context, itemID, sprintID string) (bool, error)

	// GuardarCambioDeEstado sigue los pasos de RN-010-9 en una transaccion: bloquea el
	// sprint, lee el item y si esta en el sprint, vuelve a verificar con
	// sprint.VerificarMarcado, aplica el cambio al item leido y guarda su estado y su
	// fecha de completado. Devuelve el item como quedo. Si el item o el sprint
	// desaparecieron, devuelve ErrItemInexistente o ErrSprintInexistente; si las
	// reglas ya no se cumplen, el error de VerificarMarcado. Es de todo o nada.
	GuardarCambioDeEstado(ctx context.Context, itemID, sprintID string, cambio CambioDeEstado) (backlog.ItemBacklog, error)
}

// MarcarCompletado es el caso de uso de US-010 que marca una historia del sprint
// abierto como completada.
type MarcarCompletado struct {
	repo  RepositorioMarcado
	ahora func() time.Time
}

// NuevoMarcarCompletado arma el caso de uso. ahora es el reloj: se recibe por parametro
// para que el dominio no lea la hora (RN-010-8) y los tests no dependan del dia.
func NuevoMarcarCompletado(repo RepositorioMarcado, ahora func() time.Time) *MarcarCompletado {
	return &MarcarCompletado{repo: repo, ahora: ahora}
}

// Ejecutar marca el item como completado en el sprint, con la fecha del reloj, y
// devuelve el item como quedo. Si ya estaba completado no es un error y conserva su
// fecha (RN-010-6).
func (c *MarcarCompletado) Ejecutar(ctx context.Context, itemID, sprintID string) (backlog.ItemBacklog, error) {
	en := c.ahora()
	return cambiarEstado(ctx, c.repo, itemID, sprintID, func(item backlog.ItemBacklog) backlog.ItemBacklog {
		return item.Completar(en)
	})
}

// DesmarcarCompletado es el caso de uso de US-010 que desmarca una historia completada
// por error: vuelve a pendiente y pierde la fecha (RN-010-7).
type DesmarcarCompletado struct {
	repo RepositorioMarcado
}

// NuevoDesmarcarCompletado arma el caso de uso.
func NuevoDesmarcarCompletado(repo RepositorioMarcado) *DesmarcarCompletado {
	return &DesmarcarCompletado{repo: repo}
}

// Ejecutar desmarca el item en el sprint y devuelve el item como quedo. Desmarcar un
// item que no esta completado no es un error y no cambia nada.
func (c *DesmarcarCompletado) Ejecutar(ctx context.Context, itemID, sprintID string) (backlog.ItemBacklog, error) {
	return cambiarEstado(ctx, c.repo, itemID, sprintID, backlog.ItemBacklog.Reabrir)
}

// cambiarEstado es el recorrido comun de marcar y desmarcar (spec seccion 2): busca el
// item, busca el sprint, valida con el dominio y guarda el cambio. Si el item o el
// sprint no existen responde ese error sin mirar lo demas, y nunca lo une con los de
// validacion (RN-010-1); si faltan los dos, solo ErrItemInexistente. Si algo falla no
// se guarda nada (RN-010-10).
func cambiarEstado(ctx context.Context, repo RepositorioMarcado, itemID, sprintID string, cambio CambioDeEstado) (backlog.ItemBacklog, error) {
	_, existe, err := repo.ObtenerItem(ctx, itemID)
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("consultar el item %s: %w", itemID, err)
	}
	if !existe {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", ErrItemInexistente, itemID)
	}

	destino, existe, err := repo.ObtenerSprint(ctx, sprintID)
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("consultar el sprint %s: %w", sprintID, err)
	}
	if !existe {
		return backlog.ItemBacklog{}, fmt.Errorf("%w (id %s)", ErrSprintInexistente, sprintID)
	}

	enElSprint, err := repo.ItemEnSprint(ctx, itemID, sprintID)
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("consultar si el item %s esta en el sprint %s: %w", itemID, sprintID, err)
	}

	// Los errores del dominio se devuelven tal cual: su mensaje es el que ve la persona
	// (spec US-010, seccion 7).
	if err := sprint.VerificarMarcado(datosDelSprint(destino), enElSprint); err != nil {
		return backlog.ItemBacklog{}, err
	}

	item, err := repo.GuardarCambioDeEstado(ctx, itemID, sprintID, cambio)
	if err != nil {
		return backlog.ItemBacklog{}, fmt.Errorf("guardar el estado del item %s en el sprint %s: %w", itemID, sprintID, err)
	}
	return item, nil
}
