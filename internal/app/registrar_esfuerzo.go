package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/esfuerzo"
)

// ErrIntegranteInexistente se devuelve cuando el integrante no existe o no es del
// proyecto de la operacion (RN-019-6). Igual que ErrProyectoInexistente y
// ErrItemInexistente, tiene prioridad sobre los errores de validacion y nunca se
// une con ellos.
var ErrIntegranteInexistente = errors.New("el integrante no existe")

// RepositorioEsfuerzo es lo que el caso de uso de US-019 necesita de la
// persistencia.
type RepositorioEsfuerzo interface {
	// ExisteProyecto dice si el proyecto existe. Un ID mal formado es un proyecto
	// inexistente, no un error.
	ExisteProyecto(ctx context.Context, proyectoID string) (bool, error)

	// ExisteItemEnProyecto dice si la historia existe y es de ese proyecto: una
	// historia de otro proyecto no cuenta. Un ID mal formado es inexistente.
	ExisteItemEnProyecto(ctx context.Context, proyectoID, itemID string) (bool, error)

	// ExisteIntegranteEnProyecto dice si el integrante existe y es de ese
	// proyecto: uno de otro proyecto no cuenta. Un ID mal formado es inexistente.
	ExisteIntegranteEnProyecto(ctx context.Context, proyectoID, integranteID string) (bool, error)

	// RegistrarEsfuerzo guarda el registro y lo devuelve con el ID que genera la
	// base. Si el proyecto, la historia o el integrante desaparecieron despues de
	// verificarlos, devuelve el error de inexistencia que corresponda.
	RegistrarEsfuerzo(ctx context.Context, r esfuerzo.RegistroEsfuerzo) (esfuerzo.RegistroEsfuerzo, error)
}

// RegistrarEsfuerzo es el caso de uso de US-019: registrar las horas que un
// integrante dedico a una historia.
type RegistrarEsfuerzo struct {
	repo  RepositorioEsfuerzo
	ahora func() time.Time
}

// NuevoRegistrarEsfuerzo arma el caso de uso. ahora es el reloj: se recibe por
// parametro para que el dominio no lea la hora y los tests no dependan del dia.
func NuevoRegistrarEsfuerzo(repo RepositorioEsfuerzo, ahora func() time.Time) *RegistrarEsfuerzo {
	return &RegistrarEsfuerzo{repo: repo, ahora: ahora}
}

// Ejecutar sigue el orden de la spec (seccion 2): verifica que el proyecto exista,
// despues que la historia y el integrante existan y sean de ese proyecto, despues
// valida los datos con el dominio y al final registra. El primero que falla se
// devuelve y no se une con los errores de validacion (RN-019-6).
func (c *RegistrarEsfuerzo) Ejecutar(ctx context.Context, proyectoID, itemID, integranteID string, fecha time.Time, actividad string, horas float64) (esfuerzo.RegistroEsfuerzo, error) {
	existe, err := c.repo.ExisteProyecto(ctx, proyectoID)
	if err != nil {
		return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("consultar el proyecto %s: %w", proyectoID, err)
	}
	if !existe {
		return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("%w (id %s)", ErrProyectoInexistente, proyectoID)
	}

	existe, err = c.repo.ExisteItemEnProyecto(ctx, proyectoID, itemID)
	if err != nil {
		return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("consultar la historia %s del proyecto %s: %w", itemID, proyectoID, err)
	}
	if !existe {
		return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("%w (id %s)", ErrItemInexistente, itemID)
	}

	existe, err = c.repo.ExisteIntegranteEnProyecto(ctx, proyectoID, integranteID)
	if err != nil {
		return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("consultar el integrante %s del proyecto %s: %w", integranteID, proyectoID, err)
	}
	if !existe {
		return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("%w (id %s)", ErrIntegranteInexistente, integranteID)
	}

	// RN-019-8: el dominio valida todo y devuelve los errores unidos, tal cual:
	// su mensaje es el que ve la persona (spec US-019, seccion 7).
	nuevo, err := esfuerzo.Nuevo(proyectoID, itemID, integranteID, fecha, actividad, horas, c.ahora())
	if err != nil {
		return esfuerzo.RegistroEsfuerzo{}, err
	}

	registrado, err := c.repo.RegistrarEsfuerzo(ctx, nuevo)
	if err != nil {
		return esfuerzo.RegistroEsfuerzo{}, fmt.Errorf("registrar el esfuerzo en el proyecto %s: %w", proyectoID, err)
	}
	return registrado, nil
}
