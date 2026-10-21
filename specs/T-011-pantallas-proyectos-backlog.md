# T-011 — Pantallas de proyectos, integrantes y Product Backlog

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás |
| **Issue** | #79 (tarea) · #140 (spec) |
| **Escenarios BDD** | ninguno: es una tarea técnica (ver "Decisiones") |
| **Código** | `internal/app/`, `internal/adapters/memoria/`, `internal/adapters/postgres/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| **Última actualización** | 2026-10-21 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Que una persona use desde el navegador lo que el producto ya sabe hacer con
proyectos, integrantes y Product Backlog: ver la lista de proyectos, crear y
modificar uno, registrar sus integrantes, ver su backlog ordenado, y crear,
editar y estimar historias. Hoy todo eso existe como casos de uso con tests, pero
no hay ninguna pantalla ni ninguna consulta que lea los datos: no se puede
*listar* nada.

Esta tarea cubre los requerimientos 1 (gestión de proyectos) y 2 (Product
Backlog) de la guía. Las pantallas de **sprints** (crear, asignar historias,
marcar completadas) son la tarea T-015, que sale de esta misma tarea para que cada
una se pueda entregar y probar por separado.

Se apoya en T-004 (layout, router y `Responder`) y en T-005 (errores uniformes y
lector de formularios).

## 2. Entradas

Toda ruta cuelga del esquema de URL que fijó T-004. Los formularios se envían con
HTMX (`hx-post`) y sus campos tienen los nombres de la regla RN-T005-8, para que los
errores de T-005 caigan en el campo correcto.

| Ruta | Método | Campos del formulario | Caso de uso |
|---|---|---|---|
| `/proyectos` | GET | — | `ListarProyectos` (nuevo) |
| `/proyectos/nuevo` | GET | — | — |
| `/proyectos` | POST | `nombre`, `descripcion`, `fecha_inicio`, `fecha_fin` | `CrearProyecto` (US-001) |
| `/proyectos/{id}` | GET | — | `ConsultarProyecto` (nuevo) |
| `/proyectos/{id}/editar` | GET | — | `ConsultarProyecto` |
| `/proyectos/{id}` | POST | `nombre`, `descripcion`, `fecha_inicio`, `fecha_fin` | `ModificarProyecto` (US-002) |
| `/proyectos/{id}/integrantes` | POST | `nombre`, `email`, `rol` | `RegistrarIntegrante` (US-003) |
| `/proyectos/{id}/backlog` | GET | — | `ListarBacklog` (nuevo) |
| `/proyectos/{id}/backlog` | POST | `titulo`, `descripcion`, `prioridad`, `criterios` | `CrearItemBacklog` (US-005) |
| `/proyectos/{id}/backlog/{item}/editar` | GET | — | `ConsultarItem` (nuevo) |
| `/proyectos/{id}/backlog/{item}` | POST | `titulo`, `descripcion`, `prioridad`, `criterios` | `EditarItem` (US-006) |
| `/proyectos/{id}/backlog/{item}/estimacion` | POST | `estimacion` | `EstimarItem` (US-013) |

- `{id}` e `{item}` son los IDs que genera la base. Se actualiza con POST y no con
  PUT porque un formulario HTML solo envía GET y POST.
- `criterios` es un texto con **un criterio por línea**.
- `fecha_inicio` y `fecha_fin` llegan como `AAAA-MM-DD` (`<input type="date">`).
- `estimacion` es uno de 1, 2, 3, 5, 8 o 13 (lista desplegable).
- `rol` es `product_architect`, `agile_enabler` o `product_builder` (lista).

## 3. Salidas esperadas

**Casos de uso de lectura nuevos** (`internal/app/`), con sus puertos:

```
ListarProyectos()                          -> []Proyecto
ConsultarProyecto(proyectoID)              -> Proyecto + []Integrante
ListarBacklog(proyectoID)                  -> []ItemBacklog
ConsultarItem(proyectoID, itemID)          -> ItemBacklog
```

Cada puerto es una interfaz propia del caso de uso, como en las historias. La
memoria y Postgres devuelven los datos sin ordenar; **el orden lo pone el caso de
uso**, para que las dos bases se comporten igual (RN-T011-3 a RN-T011-5).

**Pantallas** (HTML con el layout de T-004, o solo el fragmento si la pide HTMX):

| Pantalla | Qué muestra |
|---|---|
| Lista de proyectos | Nombre, fechas de inicio y fin y enlace a cada proyecto; botón "Nuevo proyecto". Si no hay ninguno: "Todavía no hay proyectos" |
| Nuevo y editar proyecto | Formulario con los cuatro campos; en editar, con los valores actuales |
| Resumen de un proyecto | Datos del proyecto, lista de integrantes (nombre, email, rol) y formulario para registrar uno; enlaces a Backlog y Sprints |
| Backlog | Tabla con número, título, prioridad, estado, Story Points (`—` si no está estimada) y cantidad de criterios; lista desplegable para estimar y enlace "Editar" en cada fila; formulario para crear una historia. Sin historias: "El backlog está vacío" |
| Editar historia | Formulario con los valores actuales; criterios uno por línea |

**Después de guardar con éxito**, la respuesta lleva la cabecera `HX-Redirect` con la
página de destino y el cuerpo vacío (HTMX navega solo). Si la petición no es de HTMX,
responde `303 See Other` con `Location`.

| Acción | Destino |
|---|---|
| Crear proyecto | `/proyectos/{id}` del proyecto nuevo |
| Modificar proyecto, registrar integrante | `/proyectos/{id}` |
| Crear, editar o estimar una historia | `/proyectos/{id}/backlog` |

**Si hay un error**, se responde con `ResponderError` (T-005): el estado y el
fragmento de error, que el layout de T-004 muestra dentro de `#mensajes`. El
formulario no se toca, así que conserva lo que la persona escribió.

## 4. Reglas de negocio

No hay reglas de dominio nuevas: las reglas siguen en US-001 a US-013. Estas son
reglas de comportamiento de la interfaz y de las lecturas.

**Capas**

- **RN-T011-1** — Un handler solo lee la petición, llama a un caso de uso y escribe
  la respuesta. No valida reglas de negocio ni habla con un repositorio.
- **RN-T011-2** — Las lecturas son casos de uso de `internal/app` con su puerto. La
  memoria y Postgres los cumplen con el mismo contrato.

**Orden**

- **RN-T011-3** — Los proyectos se listan por fecha de inicio descendente (el más
  reciente primero); empate por nombre sin distinguir mayúsculas y después por ID.
- **RN-T011-4** — El backlog se ordena por prioridad (`must`, `should`, `could`,
  `wont`) y, dentro de cada una, por número ascendente.
- **RN-T011-5** — Los integrantes se ordenan por nombre sin distinguir mayúsculas y
  después por email.

**Formularios**

- **RN-T011-6** — Las fechas, el número de la estimación y el resto de los valores
  con formato se leen con `Formulario` (T-005). Los campos vacíos no se rechazan
  ahí: lo dice el dominio (RN-T005-10).
- **RN-T011-7** — En `criterios` se ignoran las líneas vacías o con solo espacios y
  se recorta cada criterio. Un criterio que el dominio rechaza (por ejemplo,
  demasiado largo) se informa en el campo `criterios`.
- **RN-T011-8** — Un campo de texto con HTML (por ejemplo `<script>`) se muestra
  escapado en cualquier pantalla y se conserva tal cual en el formulario.
- **RN-T011-9** — Cada campo tiene su `<label>`, y los controles de una fila de la
  tabla (estimar, editar) tienen un nombre accesible que dice a qué historia
  corresponden.

**Alcance de las rutas**

- **RN-T011-10** — Un ID que no existe o está mal formado responde 404: página con
  layout en un GET, fragmento de error en un POST.
- **RN-T011-11** — Una historia solo se alcanza por la ruta de **su** proyecto.
  `/proyectos/{A}/backlog/{item de B}/…` responde 404, aunque la historia exista.
  `ConsultarItem` es el que comprueba esa pertenencia, y `EditarItem` y
  `EstimarItem` se llaman recién después de que `ConsultarItem` la confirmó.

**Navegación**

- **RN-T011-12** — `/proyectos` marca "Proyectos" como activo. Las páginas de un
  proyecto muestran el menú del proyecto (`Pagina.Proyecto`) con "Resumen" o
  "Backlog" activo según la página.
- **RN-T011-13** — El enlace "Proyectos" del menú global, que en T-004 llevaba a un
  404, pasa a mostrar la lista.

## 5. Restricciones

- Sin reglas de negocio nuevas: no se toca `internal/domain`. Cualquier cambio de
  regla va en la spec de su historia.
- Las lecturas nuevas no modifican nada y no abren transacciones.
- Los errores nuevos que pudieran aparecer se registran en
  `internal/adapters/http/catalogo.go` (RN-T005-19). Las lecturas devuelven los
  errores que ya existen (`ErrProyectoInexistente`, `ErrItemInexistente`).
- El producto usa HTMX como único JavaScript, servido desde el propio servidor
  (RN-T004-11). **Los formularios requieren JavaScript**: sin él, un error se vería
  como un fragmento sin página. La guía no pide soporte sin JavaScript.
- Sin paginación: se espera un volumen chico (decenas de proyectos y de historias
  por proyecto).
- El diseño visual es funcional (Tailwind, clases utilitarias); se mejora después
  de que todo el producto funcione.
- Los handlers se registran en `NuevoRouter` (T-004); `cmd/server` arma los
  repositorios y los casos de uso y se los pasa. `NuevoRouter` sigue siendo
  testeable sin base de datos.
- Fuera de alcance: sprints (T-015), listar con filtros (US-007), el estado general
  de un proyecto con métricas (US-004), estimar con Planning Poker (US-014 a US-018)
  y el registro de esfuerzo.

## 6. Casos límite

- **CL-T011-1** — Sin proyectos, sin integrantes o sin historias: se muestra el texto
  de lista vacía, no una tabla sin filas.
- **CL-T011-2** — Dos proyectos con la misma fecha de inicio y el mismo nombre
  (con distinta capitalización) salen siempre en el mismo orden.
- **CL-T011-3** — Dos historias con la misma prioridad salen por número; una historia
  sin estimar muestra `—` y no `0`.
- **CL-T011-4** — `criterios` con líneas vacías al principio, en el medio o al final
  se guarda sin esas líneas; con solo líneas vacías, la historia queda sin criterios.
- **CL-T011-5** — Un nombre o título con HTML, comillas, tildes o ñ se guarda y se
  muestra tal cual, sin romper la página.
- **CL-T011-6** — Una fecha que no existe (`2026-02-30`) o con formato incorrecto se
  informa en su campo (RN-T005-11); `fecha_fin` anterior a `fecha_inicio`, también.
- **CL-T011-7** — Envíos repetidos del mismo formulario de alta (doble clic) crean
  dos registros: no hay protección contra duplicados más allá de las que ya tiene
  el dominio (por ejemplo, el email repetido de un integrante).
- **CL-T011-8** — Estimar una historia ya completada responde 409 con el mensaje de
  `ErrItemCompletado`; la fila no cambia.
- **CL-T011-9** — Un `{id}` que no es un UUID (con espacios, texto largo o caracteres
  raros) responde 404, sin llegar a la base como un error interno.
- **CL-T011-10** — Un formulario con un campo de más lo ignora, y con un campo de
  menos lo trata como vacío, de modo que el dominio informa el campo que falta con
  su mensaje.

## 7. Condiciones de error

| Situación | Respuesta |
|---|---|
| Proyecto o historia inexistente, ID mal formado o historia de otro proyecto | 404: página con layout (GET) o fragmento de error (POST) |
| Dato que no cumple una regla de dominio (nombre vacío, fechas incoherentes, prioridad inválida, estimación fuera de escala, email inválido…) | 422 con el campo y el mensaje del catálogo (T-005) |
| Formato inválido (una fecha o un número mal escritos) | 422 con el campo y el mensaje de formato (T-005) |
| Email repetido en el proyecto; estimar una historia completada | 409 con el mensaje del catálogo |
| Falla inesperada (base caída, panic) | 500 con el mensaje fijo de T-005; el detalle solo va al registro |

## 8. Criterios de aceptación

- **CA-T011-1** (normal) — Con el router real y la base en memoria, un recorrido
  completo funciona de punta a punta: crear un proyecto, verlo en la lista,
  registrar un integrante, crear una historia, estimarla, editarla y ver el backlog
  ordenado por prioridad.
- **CA-T011-2** (alternativo) — Cada alta, edición y estimación responde
  `HX-Redirect` a su destino con HTMX y `303` sin HTMX; la misma pantalla se
  responde como página completa o como fragmento según la petición.
- **CA-T011-3** (límite) — Las listas vacías muestran su texto; el orden de proyectos,
  integrantes y backlog es el de las reglas RN-T011-3 a RN-T011-5 con empates.
- **CA-T011-4** (error) — Un dato inválido responde 422 con el campo correcto y el
  formulario conserva lo escrito; una historia pedida por la ruta de otro proyecto
  responde 404; un email repetido responde 409.
- **CA-T011-5** (plataforma) — Las cuatro lecturas nuevas devuelven lo mismo en
  memoria y en Postgres (tests de integración con base real), y con `make dev` y los
  datos de ejemplo se recorre el flujo en el navegador.

---

## Decisiones tomadas y descartadas

- **Las lecturas son casos de uso, no consultas directas desde los handlers.**
  Mantienen la regla de capas de ADR 0001 y dejan testear las pantallas con la base
  en memoria, sin Postgres. Se descartó que los handlers consulten el repositorio
  de Postgres directamente: más corto, pero rompe las capas y obliga a probar
  todo contra la base.
- **El orden lo decide el caso de uso, no el `ORDER BY`.** Si cada base ordenara por
  su cuenta, memoria y Postgres podrían divergir (mayúsculas, empates). Con un
  volumen chico, ordenar en memoria no cuesta nada.
- **Actualizar con `POST` a la misma URL.** Un formulario HTML solo envía GET y POST.
  Se descartaron `PUT` y `PATCH` y las cabeceras de sobreescritura de método.
- **Éxito con `HX-Redirect` y `303`.** Después de guardar se navega a una página que
  se recarga con los datos nuevos, y recargar no repite el alta (patrón
  *post/redirect/get*). Se descartó devolver la fila nueva para insertarla con HTMX:
  habría que mantener la misma fila en dos lugares.
- **Los errores no re-dibujan el formulario.** El fragmento de T-005 cae en
  `#mensajes` y el formulario queda como estaba, con lo que la persona escribió.
  Es lo que permite que HTMX 4xx funcione sin lógica extra por pantalla.
- **Esta tarea se separó de T-015.** La tarea original (8 SP) incluía los sprints, y
  al revisar el código resultó que no había ninguna lectura. Con las lecturas,
  proyectos y backlog son unos 8 SP, y los sprints otros 5: cada una se puede
  entregar y probar por separado.
- **No hay `.feature`.** Es una tarea técnica y las reglas de negocio ya tienen los
  suyos. El recorrido de CA-T011-1 es un test de Go con `httptest`, que hace de
  prueba de punta a punta de las pantallas.
- **El diseño visual queda para después.** Que cada pantalla funcione bien es lo que
  pide la guía y lo que se evalúa como producto funcional; el estilo se mejora
  cuando todo el producto esté completo.

## Uso de IA en esta especificación

- [x] Use IA para: redacción de la especificación a partir de los casos de uso, los
  repositorios y las plantillas que ya existen, y de la guía de la cátedra —
  revisado por: Tomás.
