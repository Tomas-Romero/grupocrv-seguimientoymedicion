# T-015 — Pantallas de sprints

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás |
| **Issue** | #141 (tarea) · #142 (spec) |
| **Escenarios BDD** | ninguno: es una tarea técnica (ver "Decisiones") |
| **Código** | `internal/app/`, `internal/adapters/memoria/`, `internal/adapters/postgres/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| **Última actualización** | 2026-10-21 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Que una persona gestione los sprints de un proyecto desde el navegador: ver la lista
de sprints, crear uno, ver su detalle, asignarle historias del backlog y marcarlas
como completadas o desmarcarlas. Cubre la parte de interfaz del requerimiento 3 de
la guía (crear sprints, definir el Sprint Goal, asignar historias y registrar las
completadas).

Sale de la tarea T-011, que quedó solo con proyectos, integrantes y backlog. Se
apoya en T-011 (las pantallas de proyecto y la lectura del backlog), en T-004 y en
T-005, y en los casos de uso de US-008, US-009 y US-010. **Cerrar un sprint
(US-011) y consultar los anteriores con su resultado (US-012) tienen su propia
historia**; esta tarea deja el lugar donde esos botones se agregan.

## 2. Entradas

| Ruta | Método | Campos del formulario | Caso de uso |
|---|---|---|---|
| `/proyectos/{id}/sprints` | GET | — | `ListarSprints` (nuevo) |
| `/proyectos/{id}/sprints` | POST | `sprint_goal`, `fecha_inicio`, `fecha_fin` | `CrearSprint` (US-008) |
| `/proyectos/{id}/sprints/{sprint}` | GET | — | `ConsultarSprint` (nuevo) |
| `/proyectos/{id}/sprints/{sprint}/historias` | POST | `item` (ID de la historia) | `AsignarItemASprint` (US-009) |
| `/proyectos/{id}/sprints/{sprint}/historias/{item}/completar` | POST | — | `MarcarCompletado` (US-010) |
| `/proyectos/{id}/sprints/{sprint}/historias/{item}/reabrir` | POST | — | `DesmarcarCompletado` (US-010) |

- `{id}`, `{sprint}` e `{item}` son los IDs que genera la base. Las acciones son POST
  (un formulario solo envía GET y POST) y los nombres dicen lo que hacen.
- Los formularios se envían con HTMX, igual que en T-011, y los campos tienen los
  nombres de RN-T005-8 (`sprint_goal`, `fecha_inicio`, `fecha_fin`).
- Las fechas llegan como `AAAA-MM-DD`.

## 3. Salidas esperadas

**Casos de uso de lectura nuevos** (`internal/app/`), cada uno con su puerto:

```
ListarSprints(proyectoID)                -> []ResumenDeSprint
ConsultarSprint(proyectoID, sprintID)    -> DetalleDeSprint
```

```
ResumenDeSprint { Sprint, Historias int, Planificados int, Completados int }
DetalleDeSprint { Sprint, Historias []HistoriaDelSprint, Planificados int, Completados int,
                  Disponibles []ItemBacklog }
HistoriaDelSprint { Item ItemBacklog, Completada bool }
```

- `Planificados` y `Completados` son los Story Points del sprint y salen de
  `metricas.CalcularResumenSprint` (US-025): no se vuelven a calcular acá.
- `Disponibles` son las historias del proyecto que se pueden asignar: no están
  completadas y no están en ningún sprint abierto. Incluye las que todavía no están
  estimadas o no tienen criterios, marcadas como tales, para que la persona vea por
  qué asignarlas se rechaza.
- Como en T-011, los repositorios devuelven los datos sin ordenar y **el orden lo
  pone el caso de uso** (RN-T015-3 y RN-T015-4).

**Pantallas** (con el layout de T-004, o solo el fragmento si lo pide HTMX):

| Pantalla | Qué muestra |
|---|---|
| Sprints de un proyecto | Tabla con nombre (`Sprint N`), Sprint Goal, fechas, estado, cantidad de historias y Story Points planificados y completados; formulario para crear un sprint. Sin sprints: "Todavía no hay sprints" |
| Detalle de un sprint | Sprint Goal, fechas y estado; Story Points planificados y completados; tabla de historias asignadas con su estado y el botón "Marcar completada" o "Desmarcar"; formulario para asignar una historia (lista desplegable de las disponibles) |

Con un sprint **cerrado**, el detalle muestra las historias pero no los botones ni
el formulario de asignación (el dominio igual los rechaza, RN-T015-7).

**Después de guardar con éxito**, igual que en T-011: `HX-Redirect` con HTMX o `303`
sin él.

| Acción | Destino |
|---|---|
| Crear un sprint | `/proyectos/{id}/sprints/{sprint}` del sprint nuevo |
| Asignar, completar o desmarcar | `/proyectos/{id}/sprints/{sprint}` |

Los errores se responden con `ResponderError` (T-005), y el formulario no se toca.

## 4. Reglas de negocio

No hay reglas de dominio nuevas. Son reglas de comportamiento de la interfaz y de las
lecturas.

- **RN-T015-1** — Un handler solo lee la petición, llama a un caso de uso y escribe la
  respuesta. No valida reglas de negocio ni habla con un repositorio.
- **RN-T015-2** — Las lecturas son casos de uso de `internal/app` con su puerto, que
  cumplen igual la memoria y Postgres.
- **RN-T015-3** — Los sprints se listan por número descendente (el más reciente
  primero).
- **RN-T015-4** — Las historias de un sprint se ordenan por prioridad (`must`, `should`,
  `could`, `wont`) y después por número, igual que el backlog; las disponibles, lo
  mismo.
- **RN-T015-5** — Un sprint solo se alcanza por la ruta de **su** proyecto, y una
  historia solo por la de su sprint. Una combinación que no corresponde (el sprint de
  otro proyecto, o una historia que no está en ese sprint) responde 404.
  `ConsultarSprint` comprueba la pertenencia antes de que se llame a un caso de uso de
  escritura.
- **RN-T015-6** — Las historias se asignan a un sprint por su ID, desde la lista de
  disponibles. Que una historia esté lista para el sprint (estimada y con criterios)
  lo decide `AsignarItemASprint`; la pantalla no lo repite.
- **RN-T015-7** — Con el sprint cerrado no se muestran los botones de marcar ni la
  asignación, pero las rutas igual existen: si se piden, el dominio responde
  `ErrSprintCerrado` (409). La pantalla no es la que protege la regla.
- **RN-T015-8** — Los Story Points planificados y completados vienen de
  `metricas.CalcularResumenSprint`; una historia sin estimar cuenta como 0 y se
  muestra como `—`.
- **RN-T015-9** — Las fechas se muestran como `dd/mm/aaaa` y se escriben como
  `AAAA-MM-DD`. El texto del Sprint Goal se muestra escapado.
- **RN-T015-10** — Un ID que no existe o está mal formado responde 404: página con
  layout en un GET, fragmento de error en un POST.
- **RN-T015-11** — Las páginas de sprints muestran el menú del proyecto con "Sprints"
  activo. El enlace "Sprints" del menú, que en T-004 llevaba a un 404, pasa a mostrar
  la lista.
- **RN-T015-12** — Cada campo tiene su `<label>`, y los botones de cada fila de la
  tabla tienen un nombre accesible que dice a qué historia corresponden.

## 5. Restricciones

- Sin reglas de negocio nuevas: no se toca `internal/domain`.
- Las lecturas no modifican nada. `ConsultarSprint` hace sus consultas sin transacción;
  las escrituras siguen siendo atómicas (US-009 y US-010 bloquean el sprint).
- Los errores que aparezcan se registran en `internal/adapters/http/catalogo.go`
  (RN-T005-19); se reutilizan `ErrSprintInexistente`, `ErrItemInexistente`,
  `ErrSprintCerrado` y los de US-008, US-009 y US-010.
- Los formularios requieren JavaScript (HTMX), como en T-011. Sin paginación.
- El diseño visual es funcional (Tailwind); se mejora después.
- Fuera de alcance: cerrar un sprint (US-011), consultar sprints anteriores con su
  resultado (US-012), capacidad del sprint, el burndown (US-031), el registro de
  esfuerzo por historia, y sugerir fechas para el sprint nuevo.

## 6. Casos límite

- **CL-T015-1** — Un proyecto sin sprints, o un sprint sin historias, muestra el texto de
  lista vacía y no una tabla sin filas.
- **CL-T015-2** — Un sprint sin historias estimadas muestra 0 Story Points planificados
  y 0 completados.
- **CL-T015-3** — Sin historias disponibles, la lista desplegable de asignación no se
  muestra y en su lugar dice "No hay historias para asignar".
- **CL-T015-4** — Una historia sin estimar o sin criterios aparece como disponible,
  marcada, y asignarla responde 409 con el mensaje del dominio (RN-T015-6).
- **CL-T015-5** — Asignar una historia que ya está en un sprint abierto, o ya
  completada, se rechaza con el mensaje del dominio; la lista de disponibles no
  la ofrece.
- **CL-T015-6** — Marcar dos veces una historia completada no es un error ni cambia
  nada (US-010); desmarcar una pendiente tampoco.
- **CL-T015-7** — Crear un sprint con un Sprint Goal vacío, fechas incoherentes o fuera
  del proyecto, o mientras el último sigue abierto, responde con el mensaje del catálogo
  en su campo, y el formulario conserva lo escrito.
- **CL-T015-8** — Un sprint de otro proyecto, o una historia que no está en ese sprint,
  responde 404 aunque existan por separado (RN-T015-5).
- **CL-T015-9** — Las historias de un sprint con la misma prioridad salen por número,
  siempre en el mismo orden.
- **CL-T015-10** — Un Sprint Goal con HTML, comillas, tildes o ñ se guarda y se muestra
  tal cual, escapado.

## 7. Condiciones de error

| Situación | Respuesta |
|---|---|
| Proyecto, sprint o historia inexistente; ID mal formado; sprint o historia que no corresponde a la ruta | 404: página con layout (GET) o fragmento de error (POST) |
| Dato que no cumple una regla de dominio (Sprint Goal vacío, fechas incoherentes, duración excedida…) | 422 con el campo y el mensaje del catálogo (T-005) |
| Formato inválido (una fecha mal escrita) | 422 con el campo y el mensaje de formato (T-005) |
| Sprint cerrado, historia ya asignada o completada, historia sin estimar o sin criterios, último sprint todavía abierto | 409 con el mensaje del catálogo |
| Falla inesperada (base caída, panic) | 500 con el mensaje fijo de T-005; el detalle solo va al registro |

## 8. Criterios de aceptación

- **CA-T015-1** (normal) — Con el router real y la base en memoria, un recorrido
  completo funciona: crear un proyecto con historias estimadas y con criterios, crear
  un sprint, asignarle dos historias, marcar una como completada y ver los Story Points
  planificados y completados del detalle.
- **CA-T015-2** (alternativo) — Cada alta o cambio responde `HX-Redirect` con HTMX y
  `303` sin él; la lista y el detalle se responden como página o como fragmento según
  la petición.
- **CA-T015-3** (límite) — Las listas vacías muestran su texto, el orden respeta
  RN-T015-3 y RN-T015-4 con empates, y un sprint cerrado no muestra botones.
- **CA-T015-4** (error) — Una historia sin estimar se rechaza al asignarla con 409 y su
  mensaje, un sprint de otro proyecto responde 404 y asignar a un sprint cerrado
  responde 409.
- **CA-T015-5** (plataforma) — Las lecturas nuevas devuelven lo mismo en memoria y en
  Postgres (tests de integración con base real), y con `make dev` se recorre el flujo
  en el navegador.

---

## Decisiones tomadas y descartadas

- **Esta tarea sale de T-011.** La tarea original (8 SP) reunía proyectos, backlog y
  sprints, pero no había ninguna lectura. Con las lecturas, los sprints son una tarea
  aparte de unos 5 SP que se puede entregar y probar sola.
- **Las lecturas son casos de uso**, por la misma razón que en T-011: respetan las
  capas y dejan probar las pantallas con la base en memoria. El orden lo pone el caso
  de uso, no el `ORDER BY`.
- **Los totales no se recalculan.** `metricas.CalcularResumenSprint` es la fuente de la
  verdad de los Story Points del sprint (US-025); la pantalla lo llama. Se descartó
  sumar en la plantilla.
- **La pantalla no repite las reglas de asignación.** Se ofrecen también las historias
  que todavía no están listas, marcadas, y es el dominio el que las rechaza. Se descartó
  filtrarlas: la persona se quedaba sin saber por qué una historia no aparece.
- **Acciones como rutas con verbo** (`completar`, `reabrir`) y no con métodos HTTP
  distintos, porque un formulario HTML solo envía GET y POST.
- **No hay botón de cerrar sprint** hasta que exista US-011: se agrega ahí, en el
  detalle del sprint, que es el lugar que deja esta tarea.
- **No hay `.feature`.** Es una tarea técnica y las reglas ya tienen los suyos. El
  recorrido de CA-T015-1 es un test de Go con `httptest`.
- **Sin sugerencia de fechas.** Calcularlas (por ejemplo, el día siguiente al fin del
  último sprint) es una mejora de uso; no es necesaria para cumplir el requerimiento.

## Uso de IA en esta especificación

- [x] Use IA para: redacción de la especificación a partir de los casos de uso de
  US-008, US-009 y US-010, `metricas` (US-025) y la guía de la cátedra — revisado
  por: Tomás.
