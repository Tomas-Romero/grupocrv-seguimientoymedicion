package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// ahora es la hora del reloj fijo que reciben los casos de uso en estos tests.
var ahora = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

func relojFijo() time.Time { return ahora }

// alta es una llamada a RegistrarItem que guarda el repositorio falso.
type alta struct {
	proyectoID string
	item       backlog.ItemBacklog
}

// repositorioFalso simula el repositorio de backlog sin base de datos. Responde
// lo que se le configura y anota lo que le piden.
type repositorioFalso struct {
	existe       bool
	errExiste    error
	errRegistrar error

	consultas []string
	altas     []alta
}

func (r *repositorioFalso) ExisteProyecto(_ context.Context, proyectoID string) (bool, error) {
	r.consultas = append(r.consultas, proyectoID)
	return r.existe, r.errExiste
}

func (r *repositorioFalso) RegistrarItem(_ context.Context, proyectoID string, item backlog.ItemBacklog) (backlog.ItemBacklog, error) {
	r.altas = append(r.altas, alta{proyectoID: proyectoID, item: item})
	if r.errRegistrar != nil {
		return backlog.ItemBacklog{}, r.errRegistrar
	}
	item.ID = "id-del-repositorio"
	item.Numero = len(r.altas)
	item.ProyectoID = proyectoID
	return item, nil
}

func datosValidos() backlog.DatosItem {
	return backlog.DatosItem{
		Titulo:    "Crear proyecto",
		Prioridad: backlog.PrioridadMust,
		Criterios: []string{"Se guarda con sus fechas"},
	}
}

// US-005 / CA-005-1: con un proyecto existente y datos validos, el caso de uso
// registra el item con la fecha del reloj y devuelve el ID y el numero que
// asigna la persistencia.
func TestCrearItemBacklog_Registra(t *testing.T) {
	repo := &repositorioFalso{existe: true}
	crear := app.NuevoCrearItemBacklog(repo, relojFijo)

	item, err := crear.Ejecutar(context.Background(), "proyecto-1", datosValidos())
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if len(repo.altas) != 1 {
		t.Fatalf("se registraron %d items, se esperaba 1", len(repo.altas))
	}
	registrado := repo.altas[0]
	if registrado.proyectoID != "proyecto-1" {
		t.Errorf("proyecto del alta = %q, se esperaba %q", registrado.proyectoID, "proyecto-1")
	}
	if !registrado.item.CreadoEn.Equal(ahora) {
		t.Errorf("creado en = %v, se esperaba la hora del reloj %v", registrado.item.CreadoEn, ahora)
	}
	if registrado.item.Estado != backlog.EstadoPendiente {
		t.Errorf("estado = %q, se esperaba %q", registrado.item.Estado, backlog.EstadoPendiente)
	}
	if item.ID != "id-del-repositorio" || item.Numero != 1 || item.ProyectoID != "proyecto-1" {
		t.Errorf("ID, numero y proyecto = %q, %d, %q; se esperaban los del repositorio",
			item.ID, item.Numero, item.ProyectoID)
	}
}
