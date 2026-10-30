package demo

import (
	"time"

	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/backlog"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/defecto"
	"github.com/Tomas-Romero/grupocrv-seguimientoymedicion/internal/domain/integrante"
)

// Los datos de la demostracion son fijos y ficticios (spec T-023, seccion 3). Las cantidades y las cifras de esta
// pagina son las de la spec: si se cambia una, se cambia tambien la spec y la prueba que la comprueba.

// NombreProyecto es el nombre del proyecto de demostracion. --borrar lo busca por este nombre exacto.
const NombreProyecto = "Tienda online"

const descripcionProyecto = "Sitio de ventas con catálogo, carrito y pago. Proyecto de demostración."

// dia arma una fecha de 2026 a medianoche UTC. Las fechas de los sprints y de las horas son dias, no instantes.
func dia(mes time.Month, d int) time.Time { return time.Date(2026, mes, d, 0, 0, 0, 0, time.UTC) }

var (
	inicioProyecto = dia(time.September, 14)
	finProyecto    = dia(time.December, 18)
)

// datosDeIntegrante es una persona del equipo ficticio. El orden es el de alta: Planning Poker los toma en ese orden.
type datosDeIntegrante struct {
	nombre, email string
	rol           integrante.Rol
}

var integrantes = []datosDeIntegrante{
	{"Ana Martínez", "ana@tienda.test", integrante.RolAgileEnabler},
	{"Bruno Díaz", "bruno@tienda.test", integrante.RolProductArchitect},
	{"Carla Ruiz", "carla@tienda.test", integrante.RolProductBuilder},
	{"Diego Sosa", "diego@tienda.test", integrante.RolProductBuilder},
}

// Posiciones de cada integrante en la lista de arriba.
const (
	ana = iota
	bruno
	carla
	diego
)

// historiaDemo es una historia del backlog. Puntos en cero: no se estima con EstimarItem (las de prioridad wont, y las que
// estiman o dejan abiertas las sesiones de Planning Poker).
type historiaDemo struct {
	titulo, descripcion string
	prioridad           backlog.Prioridad
	criterios           []string
	puntos              int
}

// historias son las 16 del backlog, en el orden en que se crean: el numero de cada una (#1, #2...) es su posicion mas uno.
var historias = []historiaDemo{
	{"Registrarse e iniciar sesión", "Una persona crea su cuenta con correo y contraseña, y entra con ellas.", backlog.PrioridadMust,
		[]string{"Se puede crear una cuenta con un correo nuevo", "Con las credenciales correctas se inicia la sesión", "Con una contraseña incorrecta se muestra un error"}, 5},
	{"Ver el catálogo de productos", "El catálogo lista los productos con su foto, nombre y precio.", backlog.PrioridadMust,
		[]string{"Se ven todos los productos activos", "Cada producto muestra su precio"}, 5},
	{"Buscar productos por nombre", "Una barra de búsqueda filtra el catálogo por el nombre del producto.", backlog.PrioridadMust,
		[]string{"Se encuentran productos con parte del nombre", "Sin coincidencias se muestra un aviso"}, 3},
	{"Agregar productos al carrito", "Se agrega un producto al carrito, se cambia la cantidad y se quita.", backlog.PrioridadMust,
		[]string{"Se agrega un producto al carrito", "Se cambia la cantidad de un producto", "El total se actualiza al cambiar el carrito"}, 5},
	{"Pagar con tarjeta", "El pago del carrito con tarjeta de crédito o de débito.", backlog.PrioridadMust,
		[]string{"Se cobra el total del carrito", "Un pago rechazado se informa sin perder el carrito", "Un pedido pagado queda registrado"}, 8},
	{"Filtrar el catálogo por categoría", "Se filtra el catálogo eligiendo una categoría.", backlog.PrioridadShould,
		[]string{"Se ven solo los productos de la categoría elegida", "Se puede volver a ver todo el catálogo"}, 3},
	{"Guardar productos favoritos", "Se marcan productos como favoritos para encontrarlos después.", backlog.PrioridadShould,
		[]string{"Se marca y se desmarca un favorito", "La lista de favoritos queda guardada"}, 2},
	{"Recibir el comprobante por email", "Después de pagar, llega un correo con el comprobante de la compra.", backlog.PrioridadShould,
		[]string{"Llega un correo al pagar", "El correo incluye el detalle de la compra"}, 3},
	{"Ver el historial de compras", "Una persona ve sus compras anteriores con su estado.", backlog.PrioridadShould,
		[]string{"Se listan las compras de la persona", "Se abre el detalle de una compra"}, 0},
	{"Calificar un producto", "Se califica un producto comprado con estrellas.", backlog.PrioridadShould,
		[]string{"Se califica de una a cinco estrellas", "El promedio se muestra en el producto"}, 2},
	{"Cupones de descuento", "Se aplica un cupón con un código para bajar el total.", backlog.PrioridadCould,
		[]string{"Un cupón válido baja el total", "Un cupón vencido se rechaza"}, 0},
	{"Compartir un producto en redes", "Un botón para compartir el enlace de un producto.", backlog.PrioridadCould,
		[]string{"Se copia el enlace del producto"}, 2},
	{"Modo oscuro", "Una opción para ver la tienda con colores oscuros.", backlog.PrioridadCould,
		[]string{"Se cambia entre el modo claro y el oscuro"}, 3},
	{"Aplicación móvil nativa", "Una aplicación propia para teléfonos.", backlog.PrioridadWont,
		[]string{"Se instala desde las tiendas de aplicaciones"}, 0},
	{"Programa de puntos", "Los clientes acumulan puntos por sus compras.", backlog.PrioridadWont,
		[]string{"Cada compra suma puntos"}, 0},
	{"Recomendar productos similares", "Debajo de un producto se sugieren otros parecidos.", backlog.PrioridadShould,
		[]string{"Se muestran hasta cuatro productos parecidos", "No se sugiere el mismo producto"}, 0},
}

// Numeros de las historias que usan las sesiones de Planning Poker.
const (
	historiaHistorial   = 9
	historiaCupones     = 11
	historiaRecomendada = 16
)

// sprintDemo es un sprint de la demostracion: sus fechas, sus historias, cuales se completan y cuando.
type sprintDemo struct {
	objetivo    string
	inicio, fin time.Time
	historias   []int
	completadas []completado
	cerrado     bool
}

type completado struct {
	historia int
	fecha    time.Time
}

// sprints son los cuatro, de 14 dias y seguidos. Los tres primeros se cierran y el cuarto queda abierto.
var sprints = []sprintDemo{
	{"Poder entrar a la tienda y ver los productos", dia(time.September, 14), dia(time.September, 27),
		[]int{1, 2, 3}, []completado{{1, dia(time.September, 23)}, {2, dia(time.September, 25)}}, true},
	{"Armar el carrito de compras", dia(time.September, 28), dia(time.October, 11),
		[]int{3, 4, 6}, []completado{{4, dia(time.October, 6)}, {3, dia(time.October, 9)}}, true},
	{"Poder pagar", dia(time.October, 12), dia(time.October, 25),
		[]int{5, 6, 8}, []completado{{5, dia(time.October, 21)}, {6, dia(time.October, 23)}}, true},
	{"Comprobante e historial de compras", dia(time.October, 26), dia(time.November, 8),
		[]int{8, 9, 10, 12}, []completado{{8, dia(time.October, 29)}, {10, dia(time.October, 30)}}, false},
}

// Actividades en las que se registran horas.
const (
	desarrollo = "Desarrollo"
	pruebas    = "Pruebas"
	diseno     = "Diseño"
	revision   = "Revisión de código"
	reunion    = "Reunión"
)

// registroDeHoras es un registro de esfuerzo. Las horas son multiplos de 0,25 y la fecha cae dentro de su sprint.
type registroDeHoras struct {
	sprint    int
	fecha     time.Time
	historia  int
	quien     int
	actividad string
	horas     float64
}

// horas son los 27 registros. Totales por sprint: 24,5 h, 14,75 h, 21 h y 11,5 h (hasta el 30/10), contra 19,5 h, 16,5 h, 21 h y
// 18 h estimadas (1 Story Point = 1,5 h): +25,6 %, -10,6 %, 0 % y un sprint en curso.
var horas = []registroDeHoras{
	{1, dia(time.September, 14), 1, ana, reunion, 1.5},
	{1, dia(time.September, 15), 2, carla, diseno, 3},
	{1, dia(time.September, 16), 1, diego, desarrollo, 4},
	{1, dia(time.September, 17), 2, carla, desarrollo, 4.5},
	{1, dia(time.September, 21), 3, diego, desarrollo, 3.5},
	{1, dia(time.September, 22), 1, bruno, pruebas, 2.5},
	{1, dia(time.September, 23), 2, ana, revision, 2},
	{1, dia(time.September, 24), 3, bruno, pruebas, 3.5},

	{2, dia(time.September, 28), 4, ana, reunion, 1},
	{2, dia(time.September, 29), 4, carla, diseno, 2.5},
	{2, dia(time.September, 30), 3, diego, desarrollo, 2.25},
	{2, dia(time.October, 1), 4, carla, desarrollo, 3},
	{2, dia(time.October, 5), 4, bruno, pruebas, 1.75},
	{2, dia(time.October, 6), 6, diego, desarrollo, 2.5},
	{2, dia(time.October, 8), 3, ana, revision, 1.75},

	{3, dia(time.October, 12), 5, ana, reunion, 1},
	{3, dia(time.October, 13), 5, carla, diseno, 3},
	{3, dia(time.October, 14), 5, diego, desarrollo, 5},
	{3, dia(time.October, 15), 5, carla, desarrollo, 3.5},
	{3, dia(time.October, 19), 6, bruno, pruebas, 2.5},
	{3, dia(time.October, 20), 5, ana, revision, 2},
	{3, dia(time.October, 22), 8, diego, desarrollo, 4},

	{4, dia(time.October, 26), 9, ana, reunion, 1},
	{4, dia(time.October, 27), 8, diego, desarrollo, 3},
	{4, dia(time.October, 28), 10, carla, desarrollo, 2.5},
	{4, dia(time.October, 29), 8, bruno, pruebas, 2},
	{4, dia(time.October, 30), 9, carla, diseno, 3},
}

// defectoDemo es un defecto: donde se detecto, su estado final y, si esta resuelto, en que sprint se resolvio. El estado se aplica
// cuando existe el sprint mas avanzado de los dos (el de resolucion, o el de deteccion si no esta resuelto).
type defectoDemo struct {
	historia         int
	sprintDeteccion  int
	descripcion      string
	severidad        defecto.Severidad
	estado           defecto.Estado
	sprintResolucion int
}

// defectos son los 10: criticos 2, altos 2, medios 3 y bajos 3; abiertos 4, en correccion 1, resueltos 4 y descartado 1.
var defectos = []defectoDemo{
	{3, 1, "El buscador distingue mayúsculas de minúsculas", defecto.SeveridadMedia, defecto.EstadoResuelto, 2},
	{2, 1, "El logo se corta en pantallas chicas", defecto.SeveridadBaja, defecto.EstadoDescartado, 0},
	{4, 2, "Se pierde el carrito al cerrar la sesión", defecto.SeveridadAlta, defecto.EstadoResuelto, 3},
	{1, 2, "El inicio de sesión falla con contraseñas que tienen ñ", defecto.SeveridadAlta, defecto.EstadoResuelto, 2},
	{5, 3, "El pago se cobra dos veces si la persona vuelve atrás", defecto.SeveridadCritica, defecto.EstadoResuelto, 3},
	{6, 3, "El filtro por categoría no se limpia al cambiar de página", defecto.SeveridadMedia, defecto.EstadoEnCorreccion, 0},
	{8, 4, "El asunto del correo tiene un error de ortografía", defecto.SeveridadBaja, defecto.EstadoAbierto, 0},
	{8, 4, "El comprobante llega sin el costo de envío", defecto.SeveridadMedia, defecto.EstadoAbierto, 0},
	{9, 4, "El historial no carga con más de 100 compras", defecto.SeveridadCritica, defecto.EstadoAbierto, 0},
	{10, 4, "Las calificaciones aceptan seis estrellas", defecto.SeveridadBaja, defecto.EstadoAbierto, 0},
}

// fase es el sprint en el que se aplica el estado final del defecto.
func (d defectoDemo) fase() int {
	if d.sprintResolucion > d.sprintDeteccion {
		return d.sprintResolucion
	}
	return d.sprintDeteccion
}

// votos de las sesiones de Planning Poker, por integrante en el orden de la lista de arriba.
var (
	votosHistorialRonda1 = []int{3, 5, 8, 5}
	votosHistorialRonda2 = []int{5, 5, 5, 5}
	votosCupones         = []int{3, 8, 13, 5}
	votosRecomendada     = []struct{ quien, valor int }{{ana, 5}, {carla, 8}} // Bruno y Diego todavia no votaron
)

// estimacionAcordadaHistorial es lo que el equipo acuerda para la historia 9 despues de la segunda ronda.
const estimacionAcordadaHistorial = 5
