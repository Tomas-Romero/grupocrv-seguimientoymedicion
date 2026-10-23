# T-016 — Cerrar un sprint desde el navegador

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás |
| **Issue** | #157 (tarea) · #158 (spec) |
| **Escenarios BDD** | ninguno: es una tarea técnica (ver "Decisiones") |
| **Código** | `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| **Última actualización** | 2026-10-24 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Que una persona cierre un sprint desde el navegador y vea qué pasó con cada historia:
cuáles se completaron y cuáles volvieron al backlog. Cubre la parte de interfaz del
requerimiento 3 de la guía (cerrar el sprint y devolver al backlog lo que no se completó).

Completa lo que T-015 dejó para esta tarea (su decisión "no hay botón de cerrar sprint"):
el lugar es el detalle del sprint. El caso de uso ya existe (US-011, `CerrarSprint`), así que
**no hay reglas de dominio ni lecturas nuevas**: es una ruta, un botón y la forma de mostrar
un sprint cerrado.

Depende de **DEF-007**: el detalle de un sprint cerrado se apoya en que cada historia traiga
el resultado de su fila (`HistoriaDelSprint.Resultado`) y en que los Story Points completados
no cambien después del cierre (RN-T015-13).

## 2. Entradas

| Ruta | Método | Campos del formulario | Caso de uso |
|---|---|---|---|
| `/proyectos/{id}/sprints/{sprint}/cerrar` | POST | — | `CerrarSprint` (US-011) |
| `/proyectos/{id}/sprints/{sprint}` | GET | — | `ConsultarSprint` (T-015, sin cambios) |

- `{id}` y `{sprint}` son los IDs que genera la base. La acción es POST (un formulario solo
  envía GET y POST) y el nombre dice lo que hace.
- El formulario se envía con HTMX, igual que los de T-011 y T-015.
- Cerrar no lleva campos: el resultado de cada historia lo decide el dominio con su estado
  (`sprint.Cerrar`, RN-011-5).

## 3. Salidas esperadas

**Después de cerrar con éxito**, igual que en T-015: `HX-Redirect` con HTMX o `303` sin él,
hacia `/proyectos/{id}/sprints/{sprint}`. Los errores se responden con `ResponderError` (T-005)
y la página no se toca.

**Detalle de un sprint abierto** (cambia una sola cosa): debajo del resumen de Story Points,
un botón **"Cerrar sprint"** dentro de un formulario. Pide confirmación antes de enviar, porque
un sprint cerrado no se vuelve a abrir: no existe esa operación.

**Detalle de un sprint cerrado:**

| Parte | Qué muestra |
|---|---|
| Resumen | Estado "cerrado", los Story Points planificados y completados, y "Volvieron al backlog: 1 historia" o "N historias" (se omite si N es 0) |
| Tabla de historias | Las columnas "Estado" y "Acciones" se reemplazan por **"Resultado"**: "Completada" o "No completada · volvió al backlog". Sin botones de marcar. Se quita "Estado" porque es el estado *actual* de la historia y contradice al resultado cuando se completa en otro sprint |
| Asignación | No se muestra (RN-T015-7) |

La lista de sprints no cambia: ya muestra el estado y los Story Points de cada sprint.

## 4. Reglas de negocio

No hay reglas de dominio nuevas. Son reglas de comportamiento de la interfaz.

- **RN-T016-1** — El handler solo lee la petición, comprueba que el sprint sea del proyecto
  (`ConsultarSprint`, RN-T015-5), llama a `CerrarSprint` y escribe la respuesta. No valida
  reglas de negocio ni habla con un repositorio.
- **RN-T016-2** — El botón "Cerrar sprint" solo aparece con el sprint abierto. Tiene un nombre
  accesible que dice a qué sprint corresponde ("Cerrar Sprint 2") y pide confirmación con un
  texto que dice qué pasa: las historias no completadas vuelven al backlog y el sprint no se
  puede reabrir.
- **RN-T016-3** — Un sprint sin historias se puede cerrar (RN-011-4).
- **RN-T016-4** — En un sprint cerrado, cada historia muestra su resultado (columna "Resultado",
  sin el estado actual) según el de **su fila** (`HistoriaDelSprint.Resultado`), no según su estado actual (RN-T015-13).
- **RN-T016-5** — "Volvieron al backlog" cuenta las historias del sprint con resultado
  `no_completado`. Las completadas no aparecen ahí.
- **RN-T016-6** — Cerrar un sprint que ya está cerrado se rechaza con `ErrSprintCerrado`
  (409, RN-011-2); el sprint de otro proyecto o inexistente responde 404 antes de llamar a
  `CerrarSprint`. La pantalla no es la que protege la regla: las rutas existen igual.
- **RN-T016-7** — Después de cerrar, las historias no completadas están de nuevo en el
  backlog como pendientes y se pueden asignar al sprint siguiente (RN-011-6). La pantalla no
  hace nada especial: lo muestra la lectura de T-015 (`Disponibles`).
- **RN-T016-8** — El texto del resultado ("Completada" / "No completada · volvió al backlog")
  se arma en el handler junto con el resto de la vista, como `textoDeEstado`. La plantilla no
  decide nada.

## 5. Restricciones

- Sin reglas de negocio nuevas: no se toca `internal/domain` ni `internal/app`.
- `CerrarSprint` se agrega a `CasosDeUso` y se conecta en `cmd/server`; `RepositorioAsignaciones`
  ya cumple su puerto (`RepositorioCierre`).
- Los errores que aparezcan ya están en el catálogo (`ErrSprintCerrado`, `ErrSprintInexistente`).
- Los formularios requieren JavaScript (HTMX), como en T-011 y T-015. La confirmación usa
  `hx-confirm`.
- El diseño visual es funcional (Tailwind); se mejora después.
- Fuera de alcance: reabrir un sprint, cerrar varios a la vez, el detalle de las métricas del
  cierre (velocidad, US-026) en pantalla, y la historia US-012 (consultar sprints anteriores),
  que tiene su propia spec: esta tarea solo agrega lo que necesita un sprint cerrado para
  mostrarse bien.

## 6. Casos límite

- **CL-T016-1** — Cerrar un sprint sin historias: queda cerrado, la tabla muestra el texto de
  vacío y no aparece "Volvieron al backlog".
- **CL-T016-2** — Todas las historias completadas: el resumen no muestra "Volvieron al
  backlog" y todas dicen "Completada".
- **CL-T016-3** — Ninguna completada: todas dicen "No completada · volvió al backlog" y los
  Story Points completados son 0.
- **CL-T016-4** — Un segundo cierre (por ejemplo, un doble clic): el primero cierra y el
  segundo responde 409 con el mensaje del catálogo, sin cambiar nada.
- **CL-T016-5** — Un sprint de otro proyecto, o un ID que no existe o está mal formado,
  responde 404: fragmento de error en un POST.
- **CL-T016-6** — Una historia en progreso cuando se cierra vuelve al backlog como pendiente
  (RN-011-6) y figura como no completada.
- **CL-T016-7** — Una historia no completada que se asigna al sprint siguiente y se completa
  allí: el sprint cerrado sigue mostrándola "No completada" y sin sumarla a sus Story Points
  completados (RN-T015-13).
- **CL-T016-8** — Con el sprint cerrado ya se puede crear el siguiente: el mensaje "el último
  sprint sigue abierto" desaparece (US-008).
- **CL-T016-9** — Falla inesperada de la base al cerrar: 500 con el mensaje fijo de T-005, y
  el sprint sigue abierto (el cierre es de todo o nada, RN-011-8).

## 7. Condiciones de error

| Situación | Respuesta |
|---|---|
| Sprint inexistente, ID mal formado o sprint de otro proyecto | 404: fragmento de error (POST) |
| Sprint ya cerrado | 409 con el mensaje de `ErrSprintCerrado` |
| Falla inesperada (base caída, panic) | 500 con el mensaje fijo de T-005; el detalle solo va al registro |

## 8. Criterios de aceptación

- **CA-T016-1** (normal) — Con el router real y la base en memoria: crear un sprint, asignarle
  dos historias, marcar una como completada y cerrarlo. El detalle queda "cerrado", la
  completada dice "Completada", la otra "No completada · volvió al backlog", y se ve
  "Volvieron al backlog: 1 historia" con los Story Points planificados y completados.
- **CA-T016-2** (alternativo) — Cerrar responde `HX-Redirect` con HTMX y `303` sin él, hacia el
  detalle del sprint. La historia devuelta aparece de nuevo como pendiente en el backlog y se
  puede asignar al sprint siguiente.
- **CA-T016-3** (límite) — Un sprint sin historias se cierra, y un sprint cerrado no muestra el
  botón de cerrar, los botones de marcar ni la asignación.
- **CA-T016-4** (error) — Cerrar un sprint ya cerrado responde 409, y un sprint de otro
  proyecto o inexistente responde 404.
- **CA-T016-5** (plataforma) — Con Postgres, la historia que se completa en el sprint siguiente
  sigue figurando "No completada" en el sprint cerrado, y con `make dev` se recorre el flujo en
  el navegador.

---

## Decisiones tomadas y descartadas

- **Es una tarea y no una historia.** El caso de uso y sus reglas ya están (US-011); esto es la
  interfaz, como T-011 y T-015. Se estima en 3 SP.
- **Confirmación antes de cerrar.** Un sprint cerrado no se reabre y las historias
  vuelven al backlog; un clic de más tiene costo. Se descartó una pantalla intermedia de
  confirmación: `hx-confirm` alcanza y no suma una ruta.
- **El resultado de cada historia sale de la fila del sprint**, no de su estado. Es la lección
  de DEF-007: un sprint cerrado no puede cambiar de resultado porque la historia se complete
  después en otro.
- **No se muestran las métricas del cierre** (velocidad) en esta pantalla. La velocidad es del
  proyecto y de varios sprints (US-026); su lugar es el dashboard.
- **No hay `.feature`.** Es una tarea técnica y las reglas ya tienen los suyos (US-011). El
  recorrido de CA-T016-1 es un test de Go con `httptest`.
- **No se agrega un botón "reabrir sprint".** Ninguna historia lo pide y rompería el
  historial de resultados.

## Uso de IA en esta especificación

- [x] Use IA para: redacción de la especificación a partir de US-011, T-015 y DEF-007, con
  Claude Code — revisado por: Tomás.
