package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/proyecto"
)

// repoProyectosFalso simula el repositorio de proyectos sin base de datos.
// Responde lo que se le configura y anota lo que le piden.
type repoProyectosFalso struct {
	errRegistrar error
	altas        []proyecto.Proyecto
}

func (r *repoProyectosFalso) RegistrarProyecto(_ context.Context, p proyecto.Proyecto) (proyecto.Proyecto, error) {
	r.altas = append(r.altas, p)
	if r.errRegistrar != nil {
		return proyecto.Proyecto{}, r.errRegistrar
	}
	p.ID = "id-del-repositorio"
	return p, nil
}

func fechaDe(anio int, mes time.Month, dia int) time.Time {
	return time.Date(anio, mes, dia, 0, 0, 0, 0, time.UTC)
}

// US-001 / CA-001-1, RN-001-6: con datos validos el caso de uso registra el
// proyecto ya validado (nombre recortado, sin ID) y devuelve el que arma la
// persistencia, con su ID.
func TestCrearProyecto_Registra(t *testing.T) {
	repo := &repoProyectosFalso{}
	crear := app.NuevoCrearProyecto(repo)
	inicio, fin := fechaDe(2026, time.September, 14), fechaDe(2026, time.November, 18)

	p, err := crear.Ejecutar(context.Background(), "  Demo  ", "Proyecto de prueba", inicio, fin)
	if err != nil {
		t.Fatalf("Ejecutar: %v", err)
	}

	if len(repo.altas) != 1 {
		t.Fatalf("se registraron %d proyectos, se esperaba 1", len(repo.altas))
	}
	enviado := repo.altas[0]
	if enviado.Nombre != "Demo" || enviado.Descripcion != "Proyecto de prueba" {
		t.Errorf("proyecto enviado = %+v, se esperaba el validado por el dominio", enviado)
	}
	if enviado.ID != "" {
		t.Errorf("ID enviado = %q, se esperaba vacio: lo asigna la persistencia", enviado.ID)
	}
	if p.ID != "id-del-repositorio" || p.Nombre != "Demo" {
		t.Errorf("proyecto devuelto = %+v, se esperaba el del repositorio", p)
	}
	if !p.FechaInicio.Equal(inicio) || !p.FechaFin.Equal(fin) {
		t.Errorf("fechas devueltas = %v y %v, se esperaban %v y %v", p.FechaInicio, p.FechaFin, inicio, fin)
	}
}

// US-001 / CA-001-4, CA-001-5, RN-001-7: con datos invalidos no se registra
// nada; los errores del dominio vienen unidos, siguen reconocibles con
// errors.Is y el mensaje indica el proyecto.
func TestCrearProyecto_DatosInvalidos(t *testing.T) {
	repo := &repoProyectosFalso{}
	crear := app.NuevoCrearProyecto(repo)

	p, err := crear.Ejecutar(context.Background(), "", "", fechaDe(2026, time.November, 18), fechaDe(2026, time.September, 14))
	if err == nil {
		t.Fatal("se esperaba un error y no hubo ninguno")
	}
	for _, esperado := range []error{proyecto.ErrNombreVacio, proyecto.ErrFechasIncoherentes} {
		if !errors.Is(err, esperado) {
			t.Errorf("errors.Is no reconoce %v dentro de %q", esperado, err)
		}
	}
	for _, fragmento := range []string{"el nombre es obligatorio", "fecha de fin anterior a la de inicio"} {
		if !strings.Contains(err.Error(), fragmento) {
			t.Errorf("el mensaje %q no contiene %q", err.Error(), fragmento)
		}
	}
	if len(repo.altas) != 0 {
		t.Errorf("se registraron %d proyectos, se esperaba ninguno", len(repo.altas))
	}
	if p != (proyecto.Proyecto{}) {
		t.Errorf("no se esperaba un proyecto con datos invalidos, se obtuvo %+v", p)
	}
}

// US-001 / seccion 7: un error del repositorio se devuelve envuelto con el
// nombre del proyecto, sin perder el original.
func TestCrearProyecto_ErrorDelRepositorio(t *testing.T) {
	errBase := errors.New("la base no responde")
	repo := &repoProyectosFalso{errRegistrar: errBase}
	crear := app.NuevoCrearProyecto(repo)

	_, err := crear.Ejecutar(context.Background(), "Demo", "", fechaDe(2026, time.September, 14), fechaDe(2026, time.November, 18))
	if !errors.Is(err, errBase) {
		t.Fatalf("error = %v, se esperaba %v", err, errBase)
	}
	if !strings.Contains(err.Error(), "Demo") {
		t.Errorf("el mensaje %q no indica el proyecto", err.Error())
	}
	if len(repo.altas) != 1 {
		t.Errorf("llamadas a RegistrarProyecto = %d, se esperaba 1", len(repo.altas))
	}
}
