package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
)

// estimacionGuardada es una llamada a GuardarEstimacion que anota el repositorio
// falso.
type estimacionGuardada struct {
	itemID string
	puntos backlog.StoryPoints
}

// repositorioEstimacionFalso simula la persistencia de la estimacion sin base de
// datos. Responde lo que se le configura y anota lo que le piden.
type repositorioEstimacionFalso struct {
	item        backlog.ItemBacklog
	existe      bool
	errObtener  error
	errGuardar  error
	consultados []string
	guardadas   []estimacionGuardada
}

func (r *repositorioEstimacionFalso) ObtenerItem(_ context.Context, itemID string) (backlog.ItemBacklog, bool, error) {
	r.consultados = append(r.consultados, itemID)
	return r.item, r.existe, r.errObtener
}

func (r *repositorioEstimacionFalso) GuardarEstimacion(_ context.Context, itemID string, puntos backlog.StoryPoints) error {
	r.guardadas = append(r.guardadas, estimacionGuardada{itemID: itemID, puntos: puntos})
	return r.errGuardar
}

// itemDelRepositorio arma el item que devuelve el repositorio falso: valido, en
// el estado pedido y con el ID que usan estos tests.
func itemDelRepositorio(t *testing.T, estado backlog.Estado) backlog.ItemBacklog {
	t.Helper()
	item, err := backlog.NuevoItem(datosValidos(), ahora)
	if err != nil {
		t.Fatalf("NuevoItem: %v", err)
	}
	item.ID = "item-1"
	item.Numero = 4
	item.ProyectoID = "proyecto-1"
	item.Estado = estado
	return item
}

// US-013 / CA-013-1: el caso de uso carga el item, lo estima con el dominio y
// guarda solo la estimacion. Devuelve el item ya estimado.
func TestEstimarItem_Estima(t *testing.T) {
	repo := &repositorioEstimacionFalso{item: itemDelRepositorio(t, backlog.EstadoPendiente), existe: true}
	estimar := app.NuevoEstimarItem(repo)

	item, err := estimar.Ejecutar(context.Background(), "item-1", 5)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if puntos, estimado := item.StoryPoints.Valor(); puntos != 5 || !estimado {
		t.Errorf("story points del item = (%d, %t), se esperaba (5, true)", puntos, estimado)
	}
	if item.ID != "item-1" || item.Numero != 4 {
		t.Errorf("el item devuelto es %+v, se esperaba el item 1 numero 4", item)
	}
	if len(repo.guardadas) != 1 {
		t.Fatalf("se guardaron %d estimaciones, se esperaba 1", len(repo.guardadas))
	}
	guardada := repo.guardadas[0]
	if puntos, estimado := guardada.puntos.Valor(); guardada.itemID != "item-1" || puntos != 5 || !estimado {
		t.Errorf("se guardo %+v, se esperaba el item 1 con 5 story points", guardada)
	}
}

// US-013 / CA-013-2, RN-013-3: un item ya estimado se vuelve a estimar.
func TestEstimarItem_ReEstima(t *testing.T) {
	item := itemDelRepositorio(t, backlog.EstadoEnProgreso)
	item, err := item.Estimar(3)
	if err != nil {
		t.Fatalf("Estimar: %v", err)
	}
	repo := &repositorioEstimacionFalso{item: item, existe: true}

	estimado, err := app.NuevoEstimarItem(repo).Ejecutar(context.Background(), "item-1", 8)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}
	if puntos, _ := estimado.StoryPoints.Valor(); puntos != 8 {
		t.Errorf("story points = %d, se esperaba 8", puntos)
	}
}

// US-013 / CA-013-6, RN-013-6: si el item no existe, el caso de uso devuelve
// ErrItemInexistente sin mirar el valor y no guarda nada. El error no se une con
// los del dominio.
func TestEstimarItem_ItemInexistente(t *testing.T) {
	repo := &repositorioEstimacionFalso{existe: false}

	// 4 esta fuera de la escala: si se validara primero, el error seria otro.
	_, err := app.NuevoEstimarItem(repo).Ejecutar(context.Background(), "no-existe", 4)

	if !errors.Is(err, app.ErrItemInexistente) {
		t.Fatalf("error = %v, se esperaba ErrItemInexistente", err)
	}
	if errors.Is(err, backlog.ErrEstimacionFueraDeEscala) {
		t.Errorf("el error %q tambien reporta el valor: ErrItemInexistente no se une con los de validacion", err)
	}
	if !strings.Contains(err.Error(), "no-existe") {
		t.Errorf("el error %q no indica el ID recibido", err)
	}
	if len(repo.guardadas) != 0 {
		t.Errorf("se guardaron %d estimaciones, se esperaba ninguna", len(repo.guardadas))
	}
}

// US-013 / CA-013-4, RN-013-8: un valor fuera de la escala o un item completado
// se rechazan con el error del dominio, sin envolver, y no se guarda nada.
func TestEstimarItem_Rechazos(t *testing.T) {
	casos := []struct {
		nombre string
		estado backlog.Estado
		puntos int
		error  error
	}{
		{"fuera de la escala", backlog.EstadoPendiente, 4, backlog.ErrEstimacionFueraDeEscala},
		{"cero", backlog.EstadoPendiente, 0, backlog.ErrEstimacionFueraDeEscala},
		{"item completado", backlog.EstadoCompletado, 5, backlog.ErrItemCompletado},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			repo := &repositorioEstimacionFalso{item: itemDelRepositorio(t, c.estado), existe: true}

			_, err := app.NuevoEstimarItem(repo).Ejecutar(context.Background(), "item-1", c.puntos)

			if !errors.Is(err, c.error) {
				t.Fatalf("error = %v, se esperaba %v", err, c.error)
			}
			if len(repo.guardadas) != 0 {
				t.Errorf("se guardaron %d estimaciones, se esperaba ninguna", len(repo.guardadas))
			}
		})
	}
}

// Un fallo del repositorio al cargar el item se devuelve envuelto, con el ID, y
// sigue reconociendose con errors.Is.
func TestEstimarItem_ErrorAlObtener(t *testing.T) {
	falla := errors.New("se corto la conexion")
	repo := &repositorioEstimacionFalso{errObtener: falla}

	_, err := app.NuevoEstimarItem(repo).Ejecutar(context.Background(), "item-1", 5)

	if !errors.Is(err, falla) {
		t.Fatalf("error = %v, se esperaba el del repositorio", err)
	}
	if !strings.Contains(err.Error(), "item-1") {
		t.Errorf("el error %q no indica el item", err)
	}
	if errors.Is(err, app.ErrItemInexistente) {
		t.Errorf("un fallo del repositorio no es un item inexistente: %v", err)
	}
}

// Un fallo del repositorio al guardar se devuelve envuelto y no se devuelve item.
func TestEstimarItem_ErrorAlGuardar(t *testing.T) {
	falla := errors.New("se corto la conexion")
	repo := &repositorioEstimacionFalso{item: itemDelRepositorio(t, backlog.EstadoPendiente), existe: true, errGuardar: falla}

	item, err := app.NuevoEstimarItem(repo).Ejecutar(context.Background(), "item-1", 5)

	if !errors.Is(err, falla) {
		t.Fatalf("error = %v, se esperaba el del repositorio", err)
	}
	if item.ID != "" {
		t.Errorf("devolvio el item %+v, se esperaba el valor cero", item)
	}
}

// Si el item desaparece entre la lectura y el guardado, el repositorio devuelve
// ErrItemInexistente y el caso de uso lo deja pasar reconocible.
func TestEstimarItem_ItemQueDesaparece(t *testing.T) {
	repo := &repositorioEstimacionFalso{
		item:       itemDelRepositorio(t, backlog.EstadoPendiente),
		existe:     true,
		errGuardar: app.ErrItemInexistente,
	}

	_, err := app.NuevoEstimarItem(repo).Ejecutar(context.Background(), "item-1", 5)

	if !errors.Is(err, app.ErrItemInexistente) {
		t.Fatalf("error = %v, se esperaba ErrItemInexistente", err)
	}
}
