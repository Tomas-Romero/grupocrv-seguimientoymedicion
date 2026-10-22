package memoria_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/planningpoker"
)

// baseParaPoker es una base con un proyecto, una historia pendiente y los integrantes
// pedidos, en ese orden de alta.
type baseParaPoker struct {
	base        *memoria.Base
	proyecto    string
	item        string
	integrantes []string
}

func nuevaBaseParaPoker(t *testing.T, nombres ...string) baseParaPoker {
	t.Helper()
	ctx := context.Background()
	base := memoria.NuevaBase()
	b := baseParaPoker{base: base, proyecto: base.AgregarProyecto()}
	item, err := base.RegistrarItem(ctx, b.proyecto, itemValido(t, "Alta de proyectos"))
	if err != nil {
		t.Fatalf("RegistrarItem: %v", err)
	}
	b.item = item.ID
	for _, nombre := range nombres {
		b.integrantes = append(b.integrantes, b.registrar(t, nombre))
	}
	return b
}

func (b baseParaPoker) registrar(t *testing.T, nombre string) string {
	t.Helper()
	nuevo, err := integrante.Nuevo(b.proyecto, nombre, nombre+"@ejemplo.test", integrante.RolProductBuilder)
	if err != nil {
		t.Fatalf("integrante.Nuevo: %v", err)
	}
	registrado, err := b.base.RegistrarIntegrante(context.Background(), nuevo)
	if err != nil {
		t.Fatalf("RegistrarIntegrante: %v", err)
	}
	return registrado.ID
}

// sesionPara arma con el dominio la sesion que mandaria el caso de uso.
func sesionPara(t *testing.T, b baseParaPoker) planningpoker.SesionPlanningPoker {
	t.Helper()
	sesion, err := planningpoker.NuevaSesion(b.item, planningpoker.DatosDeLaHistoria{}, b.integrantes)
	if err != nil {
		t.Fatalf("NuevaSesion: %v", err)
	}
	return sesion
}

// US-014 / RN-014-4: los integrantes del proyecto, en orden de alta. Un proyecto sin
// integrantes, o que no existe, no tiene ninguno.
func TestBase_IDsDeIntegrantes(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaPoker(t, "ana", "beto", "caro")

	ids, err := b.base.IDsDeIntegrantes(ctx, b.proyecto)
	if err != nil || !slices.Equal(ids, b.integrantes) {
		t.Errorf("IDsDeIntegrantes = (%q, %v), se esperaban %q", ids, err, b.integrantes)
	}
	if ids, err := b.base.IDsDeIntegrantes(ctx, idSinDuenoAsignaciones); err != nil || len(ids) != 0 {
		t.Errorf("de un proyecto que no esta = (%q, %v), se esperaba ninguno", ids, err)
	}
}

// US-014 / CA-014-1, RN-014-4, RN-014-6, RN-014-12: registrar guarda la sesion con un ID
// nuevo, votando, en la ronda 1 y con los integrantes del proyecto como participantes.
// Desde ahi la historia tiene una sesion abierta.
func TestBase_RegistrarSesion(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaPoker(t, "ana", "beto", "caro")
	if abierta, _ := b.base.TieneSesionAbierta(ctx, b.item); abierta {
		t.Fatal("la historia tiene una sesion abierta antes de iniciar ninguna")
	}

	registrada, err := b.base.RegistrarSesion(ctx, b.proyecto, sesionPara(t, b))
	if err != nil {
		t.Fatalf("RegistrarSesion: %v", err)
	}

	if registrada.ID == "" || registrada.ItemID != b.item || registrada.Estado != planningpoker.EstadoVotando || registrada.Ronda != 1 {
		t.Errorf("registrada = %+v, se esperaba con ID, del item, votando y en la ronda 1", registrada)
	}
	if !slices.Equal(registrada.Participantes, b.integrantes) {
		t.Errorf("participantes = %q, se esperaban %q", registrada.Participantes, b.integrantes)
	}
	if abierta, _ := b.base.TieneSesionAbierta(ctx, b.item); !abierta {
		t.Error("despues de registrarla, la historia no tiene una sesion abierta")
	}
	guardada, ok := b.base.SesionAbiertaDeLaHistoria(b.item)
	if !ok || guardada.ID != registrada.ID || !slices.Equal(guardada.Participantes, b.integrantes) {
		t.Errorf("SesionAbiertaDeLaHistoria = (%+v, %t), se esperaba la registrada", guardada, ok)
	}
}

// US-014 / CA-014-3, RN-014-4, CL-014-7: los participantes se fijan al registrar. Un
// integrante que se suma despues no participa, aunque la sesion que llego del caso de
// uso fuera otra: la base guarda los integrantes que tiene en ese momento.
func TestBase_RegistrarSesion_ParticipantesAlIniciar(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaPoker(t, "ana", "beto")
	sesion := sesionPara(t, b)
	caro := b.registrar(t, "caro")

	registrada, err := b.base.RegistrarSesion(ctx, b.proyecto, sesion)
	if err != nil {
		t.Fatalf("RegistrarSesion: %v", err)
	}
	if !slices.Contains(registrada.Participantes, caro) || len(registrada.Participantes) != 3 {
		t.Errorf("participantes = %q, se esperaban los tres integrantes al registrar", registrada.Participantes)
	}

	b.registrar(t, "dani")
	guardada, _ := b.base.SesionAbiertaDeLaHistoria(b.item)
	if len(guardada.Participantes) != 3 {
		t.Errorf("participantes = %q, el que se sumo despues de registrar no participa", guardada.Participantes)
	}
}

// US-014 / RN-014-11, CL-014-5, CL-014-8, CL-014-11: la base vuelve a verificar las reglas
// con lo que tiene guardado. Una segunda sesion abierta, una historia que se completo o
// un proyecto que quedo con un solo integrante se rechazan, y no se guarda nada.
func TestBase_RegistrarSesion_Reglas(t *testing.T) {
	ctx := context.Background()

	t.Run("sesion abierta", func(t *testing.T) {
		b := nuevaBaseParaPoker(t, "ana", "beto")
		if _, err := b.base.RegistrarSesion(ctx, b.proyecto, sesionPara(t, b)); err != nil {
			t.Fatalf("primera sesion: %v", err)
		}
		if _, err := b.base.RegistrarSesion(ctx, b.proyecto, sesionPara(t, b)); !errors.Is(err, planningpoker.ErrSesionAbierta) {
			t.Errorf("segunda sesion: error = %v, se esperaba ErrSesionAbierta", err)
		}
	})

	t.Run("historia completada", func(t *testing.T) {
		b := nuevaBaseParaPoker(t, "ana", "beto")
		completado := itemValido(t, "Listar proyectos")
		completado.Estado = backlog.EstadoCompletado
		item, err := b.base.RegistrarItem(ctx, b.proyecto, completado)
		if err != nil {
			t.Fatalf("RegistrarItem: %v", err)
		}
		sesion := sesionPara(t, b)
		sesion.ItemID = item.ID
		if _, err := b.base.RegistrarSesion(ctx, b.proyecto, sesion); !errors.Is(err, backlog.ErrItemCompletado) {
			t.Errorf("error = %v, se esperaba ErrItemCompletado", err)
		}
		if abierta, _ := b.base.TieneSesionAbierta(ctx, item.ID); abierta {
			t.Error("se guardo la sesion de una historia completada")
		}
	})

	t.Run("un solo integrante", func(t *testing.T) {
		b := nuevaBaseParaPoker(t, "ana")
		sesion := planningpoker.SesionPlanningPoker{ItemID: b.item, Estado: planningpoker.EstadoVotando, Ronda: 1}
		if _, err := b.base.RegistrarSesion(ctx, b.proyecto, sesion); !errors.Is(err, planningpoker.ErrParticipantesInsuficientes) {
			t.Errorf("error = %v, se esperaba ErrParticipantesInsuficientes", err)
		}
	})
}

// US-014 / CA-014-7, RN-014-1: una historia que no esta, o que es de otro proyecto, se
// rechaza con ErrItemInexistente.
func TestBase_RegistrarSesion_HistoriaInexistente(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaPoker(t, "ana", "beto")
	otroProyecto := b.base.AgregarProyecto()

	sesion := sesionPara(t, b)
	if _, err := b.base.RegistrarSesion(ctx, otroProyecto, sesion); !errors.Is(err, app.ErrItemInexistente) {
		t.Errorf("de otro proyecto: error = %v, se esperaba ErrItemInexistente", err)
	}
	sesion.ItemID = idSinDuenoAsignaciones
	if _, err := b.base.RegistrarSesion(ctx, b.proyecto, sesion); !errors.Is(err, app.ErrItemInexistente) {
		t.Errorf("historia que no esta: error = %v, se esperaba ErrItemInexistente", err)
	}
}

// SesionAbiertaDeLaHistoria devuelve una copia: cambiar sus participantes no cambia lo
// guardado.
func TestBase_SesionAbiertaDeLaHistoria_DevuelveUnaCopia(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseParaPoker(t, "ana", "beto")
	if _, err := b.base.RegistrarSesion(ctx, b.proyecto, sesionPara(t, b)); err != nil {
		t.Fatalf("RegistrarSesion: %v", err)
	}

	copia, _ := b.base.SesionAbiertaDeLaHistoria(b.item)
	copia.Participantes[0] = "cambiado"

	if guardada, _ := b.base.SesionAbiertaDeLaHistoria(b.item); guardada.Participantes[0] == "cambiado" {
		t.Error("cambiar la copia modifico lo guardado")
	}
	if _, ok := b.base.SesionAbiertaDeLaHistoria(idSinDuenoAsignaciones); ok {
		t.Error("una historia sin sesiones tiene una sesion abierta")
	}
}
