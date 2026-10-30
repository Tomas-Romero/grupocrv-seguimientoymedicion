package plantillas_test

import (
	"regexp"
	"testing"

	plantillas "github.com/Tomas-Romero/grupocrv-seguimientoymedicion/web/templates"
)

// T-024 / CA-T024-3, RN-T024-3: la prioridad y el estado de una historia son insignias con color y texto a la vez. El tono va en
// data-tono para que la prueba no dependa de las clases de Tailwind: el color nunca es el unico medio.
func TestBacklog_PrioridadYEstadoSonInsigniasConTexto(t *testing.T) {
	escala := []plantillas.Opcion{{Valor: "5", Texto: "5"}}
	item := func(n int, titulo, prioridad, estado string) plantillas.FilaItem {
		return plantillas.FilaItem{Numero: n, Titulo: titulo, Prioridad: prioridad, Estado: estado, Puntos: "5", Criterios: 2, Editar: "/editar", Estimar: "/estimar", Escala: escala}
	}
	html := dibujar(t, plantillas.BacklogDeProyecto(plantillas.VistaBacklog{Items: []plantillas.FilaItem{
		item(1, "Login", "Must", "Completado"),
		item(2, "Perfil", "Should", "Pendiente"),
		item(3, "Tema oscuro", "Could", "En progreso"),
		item(4, "App movil", "Wont", "Pendiente"),
	}}))

	casos := []struct{ texto, tono string }{
		{"Must", "primario"}, {"Should", "info"}, {"Could", "neutro"}, {"Wont", "apagado"},
		{"Completado", "ok"}, {"Pendiente", "atencion"}, {"En progreso", "info"},
	}
	for _, c := range casos {
		patron := regexp.MustCompile(`<span[^>]*data-tono="` + c.tono + `"[^>]*>\s*` + regexp.QuoteMeta(c.texto) + `\s*</span>`)
		if !patron.MatchString(html) {
			t.Errorf("no hay una insignia %q con el tono %q", c.texto, c.tono)
		}
	}
}

// T-024 / CA-T024-4, RN-T024-6: a 375 px la pagina no se desplaza en horizontal; una tabla que no entra se desplaza dentro de su contenedor.
func TestBacklog_LaTablaSeDesplazaDentroDeSuContenedor(t *testing.T) {
	html := dibujar(t, plantillas.BacklogDeProyecto(plantillas.VistaBacklog{Items: []plantillas.FilaItem{
		{Numero: 1, Titulo: "Login", Prioridad: "Must", Estado: "Pendiente", Puntos: "5", Criterios: 1, Editar: "/e", Estimar: "/s"},
	}}))

	if !regexp.MustCompile(`class="[^"]*overflow-x-auto[^"]*"[^>]*>\s*<table`).MatchString(html) {
		t.Error("la tabla del backlog no esta dentro de un contenedor con overflow-x-auto")
	}
}

// T-024 / CA-T024-2: una insignia de un valor que no conoce se dibuja igual, en tono neutro y con su texto: nada se pierde.
func TestBacklog_UnValorDesconocidoSeDibujaNeutro(t *testing.T) {
	html := dibujar(t, plantillas.BacklogDeProyecto(plantillas.VistaBacklog{Items: []plantillas.FilaItem{
		{Numero: 1, Titulo: "Raro", Prioridad: "Otra", Estado: "Rarisimo", Puntos: "—", Criterios: 1, Editar: "/e", Estimar: "/s"},
	}}))

	for _, texto := range []string{"Otra", "Rarisimo"} {
		patron := regexp.MustCompile(`<span[^>]*data-tono="neutro"[^>]*>\s*` + texto + `\s*</span>`)
		if !patron.MatchString(html) {
			t.Errorf("la insignia %q desconocida no se dibuja en tono neutro", texto)
		}
	}
}

// T-024 / CA-T024-2: cada integrante del resumen lleva un circulo con sus iniciales. Es decoracion (aria-hidden): el nombre completo sigue escrito al lado.
func TestResumen_IntegrantesConSusIniciales(t *testing.T) {
	html := dibujar(t, plantillas.ResumenDeProyecto(plantillas.VistaProyecto{
		Nombre: "Tienda online",
		Integrantes: []plantillas.FilaIntegrante{
			{Nombre: "Ana Martínez", Email: "ana@tienda.test", Rol: "Agile Enabler"},
			{Nombre: "bruno", Email: "bruno@tienda.test", Rol: "Product Builder"},
		},
	}))

	for _, c := range []struct{ iniciales, nombre string }{{"AM", "Ana Martínez"}, {"B", "bruno"}} {
		circulo := regexp.MustCompile(`<span[^>]*aria-hidden="true"[^>]*>` + c.iniciales + `</span>`)
		if !circulo.MatchString(html) {
			t.Errorf("no hay un circulo aria-hidden con las iniciales %q", c.iniciales)
		}
		if !regexp.MustCompile(`>\s*` + c.nombre + `\s*<`).MatchString(html) {
			t.Errorf("el nombre %q ya no esta escrito", c.nombre)
		}
	}
}
