package memoria_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/adapters/memoria"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/app"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
)

// T-021 / RN-T021-2: la base en memoria cumple el puerto de ListarDefectos.
var _ app.RepositorioListarDefectos = (*memoria.Base)(nil)

// T-021 / RN-T021-3, RN-T021-13 / CA-T021-6: ListarDefectosDelProyecto devuelve los defectos del
// proyecto, de cualquier estado, con el numero y el titulo de la historia y los numeros de los
// sprints de deteccion y de resolucion traducidos desde sus IDs (RN-022-6). Los defectos de otro
// proyecto no aparecen.
func TestBase_ListarDefectosDelProyecto(t *testing.T) {
	ctx := context.Background()
	b := nuevaBaseConDefectos(t)
	abierto := b.registrado(t)
	porResolver := b.registrado(t)
	deOtro, err := defecto.Nuevo(b.otro, b.itemOtro, b.sprintOtro, "De otro proyecto", defecto.SeveridadBaja)
	if err != nil {
		t.Fatalf("defecto.Nuevo: %v", err)
	}
	if _, err := b.base.RegistrarDefecto(ctx, deOtro); err != nil {
		t.Fatalf("RegistrarDefecto: %v", err)
	}

	segundo := b.segundoSprint(t)
	resuelto := porResolver
	resuelto.Estado, resuelto.SprintResolucionID = defecto.EstadoResuelto, segundo
	if err := b.base.ActualizarEstadoDefecto(ctx, resuelto, defecto.EstadoAbierto); err != nil {
		t.Fatalf("ActualizarEstadoDefecto: %v", err)
	}

	got, err := b.base.ListarDefectosDelProyecto(ctx, b.demo)
	if err != nil || len(got) != 2 {
		t.Fatalf("ListarDefectosDelProyecto = %+v, %v; quiero los dos defectos de Demo", got, err)
	}
	porID := map[string]app.DefectoListado{}
	for _, l := range got {
		porID[l.Defecto.ID] = l
	}

	if l := porID[abierto.ID]; l.Defecto != abierto || l.HistoriaNumero != 1 || l.HistoriaTitulo != "Alta de proyectos" ||
		l.SprintDeteccion != 1 || l.SprintResolucion != 0 {
		t.Errorf("defecto abierto = %+v, quiero los datos de %+v, historia #1, Sprint 1 y sin sprint de resolucion", l, abierto)
	}
	if l := porID[resuelto.ID]; l.Defecto != resuelto || l.HistoriaNumero != 1 || l.SprintDeteccion != 1 || l.SprintResolucion != 2 {
		t.Errorf("defecto resuelto = %+v, quiero los datos de %+v, Sprint 1 de deteccion y Sprint 2 de resolucion", l, resuelto)
	}
}

// T-021 / RN-T021-3, CL-T021-1: un proyecto sin defectos, o que no esta, devuelve una lista vacia y sin error.
func TestBase_ListarDefectosDelProyecto_Vacio(t *testing.T) {
	b := nuevaBaseConDefectos(t)

	for _, proyecto := range []string{b.demo, "no-existe", ""} {
		got, err := b.base.ListarDefectosDelProyecto(context.Background(), proyecto)
		if err != nil || got == nil || len(got) != 0 {
			t.Errorf("ListarDefectosDelProyecto(%q) = %v, %v; quiero una lista vacia y sin error", proyecto, got, err)
		}
	}
}

// T-021 / RN-T021-4, CL-T021-8 / CA-T021-6: la base guarda el instante de alta de cada defecto y es
// creciente entre altas, para que el orden "el mas reciente primero" sea el de Postgres y no haya
// dos defectos con la misma fecha de carga.
func TestBase_ListarDefectosDelProyecto_CreadoEnCreciente(t *testing.T) {
	b := nuevaBaseConDefectos(t)
	for i := 0; i < 5; i++ {
		b.registrado(t)
	}

	got, err := b.base.ListarDefectosDelProyecto(context.Background(), b.demo)
	if err != nil || len(got) != 5 {
		t.Fatalf("ListarDefectosDelProyecto = %+v, %v; quiero 5 defectos", got, err)
	}
	// Se ordenan por ID, que en la memoria es correlativo: el de ID mas alto se dio de alta despues.
	slices.SortFunc(got, func(a, c app.DefectoListado) int { return strings.Compare(a.Defecto.ID, c.Defecto.ID) })
	var anterior app.DefectoListado
	for _, l := range got {
		if l.CreadoEn.IsZero() {
			t.Fatalf("el defecto %s no tiene fecha de carga", l.Defecto.ID)
		}
		if anterior.Defecto.ID != "" && !l.CreadoEn.After(anterior.CreadoEn) {
			t.Errorf("CreadoEn de %s (%v) no es posterior al de %s (%v)", l.Defecto.ID, l.CreadoEn, anterior.Defecto.ID, anterior.CreadoEn)
		}
		anterior = l
	}
}
