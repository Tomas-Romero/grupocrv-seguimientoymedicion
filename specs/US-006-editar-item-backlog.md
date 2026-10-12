# US-006 — Editar y repriorizar ítems del backlog

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #8 |
| **Escenarios BDD** | `features/US-006-editar-item-backlog.feature` |
| **Código** | `internal/domain/backlog/` (`ItemBacklog.Editar`), caso de uso en `internal/app/`, repositorio en `internal/adapters/postgres/`. Sin migración nueva |
| **Última actualización** | 2026-10-12 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero editar los datos de un ítem del backlog y
> cambiar su prioridad, para mantener el backlog al día a medida que se refina
> el trabajo.

Corregir el título, la descripción, la prioridad y los criterios de aceptación
de un ítem creado con US-005, sin tener que crearlo de nuevo. "Repriorizar" es
cambiar la prioridad MoSCoW con esta misma operación.

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `itemID` | UUID | Sí | Un ítem existente. Viene del contexto de navegación (la URL); lo verifica el caso de uso, nunca el dominio |
| `datos` | `backlog.DatosItem` (US-005) | Sí | Título, descripción, prioridad y criterios, con los rangos de la sección 2 de US-005 |

Se mandan siempre los cuatro datos, no solo los que cambian, como en US-002
(RN-002-3): el formulario los muestra precargados.

La edición pasa por las mismas capas que US-005, en este orden:

1. **Aplicación:** carga el ítem con una lectura simple. Si no existe, devuelve
   `ErrItemInexistente` sin validar nada más (RN-006-7).
2. **Dominio:** `ItemBacklog.Editar(datos)` es una función pura que valida el
   estado del ítem y los datos nuevos, y devuelve una copia editada o los
   errores. No lee la base ni el reloj.
3. **Persistencia:** guarda los cuatro datos con un único `UPDATE` condicionado
   (RN-006-8).

## 3. Salidas esperadas

El `ItemBacklog` con los cuatro datos reemplazados, recortados igual que al
crear (RN-005-1, RN-005-4 y RN-005-8):

| Atributo | Después de editar |
|---|---|
| Título, descripción, prioridad, criterios | Los recibidos, recortados |
| ID, número, proyecto, estado, Story Points, fecha de creación | Sin cambios |
| Asignación a un sprint (US-009) | Sin cambios |

Si hay errores, no se devuelve ítem y la base queda como estaba.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-006-1** (dominio) — Los datos nuevos cumplen exactamente las reglas de
  creación: RN-005-1 a RN-005-5 y RN-005-8, con los errores unidos en el orden
  de RN-005-13 y las posiciones de los criterios contadas desde 1 (RN-005-14).
  `Editar` usa la misma validación que `NuevoItem`; no se escribe una segunda.
- **RN-006-2** (dominio) — La edición es completa: reemplaza los cuatro datos a
  la vez. La lista de criterios se reemplaza entera, y puede quedar vacía
  (RN-005-8). No hay edición parcial.
- **RN-006-3** (dominio) — Repriorizar es cambiar la prioridad MoSCoW con esta
  operación. No hay orden fino dentro de una misma prioridad.
- **RN-006-4** (dominio) — No se editan el número, el estado ni los Story
  Points. Los Story Points son de US-013 y el estado, de US-010.
- **RN-006-5** (dominio) — Un ítem `completado` no se edita, por la misma razón
  que RN-013-4: ya forma parte de un sprint medido. Se valida primero el estado:
  si está completado, es el único error aunque los datos también sean
  inválidos (mismo criterio que RN-013-5).
- **RN-006-6** (dominio) — Un ítem `pendiente` o `en_progreso` se edita esté o
  no asignado a un sprint abierto. La Definition of Ready (US-009) no se vuelve
  a verificar al editar.
- **RN-006-7** (aplicación) — Si el ítem no existe, el caso de uso devuelve
  `ErrItemInexistente` antes de llamar al dominio, con prioridad y sin unirlo
  con los errores de validación (mismo criterio que RN-005-9 y RN-013-6). Un ID
  mal formado es un ítem inexistente, como en US-005.
- **RN-006-8** (persistencia) — Se guarda con un único `UPDATE` condicionado:
  `UPDATE items_backlog SET titulo, descripcion, prioridad, criterios ...`
  `WHERE id = $1 AND estado <> 'completado'`. Si no actualiza ninguna fila, el
  repositorio relee el ítem: si no existe, devuelve `ErrItemInexistente`; si
  está completado, `ErrItemCompletado`. Así, un ítem que pasa a completado
  (US-010) entre la lectura del caso de uso y el `UPDATE` no se edita.
- **RN-006-9** (persistencia) — Dos ediciones simultáneas del mismo ítem
  terminan las dos bien y gana la última, igual que RN-013-7. No hay bloqueo de
  filas ni versionado.
- **RN-006-10** (aplicación) — La operación es de todo o nada: si algo falla,
  el ítem queda exactamente como estaba.

## 5. Restricciones

- **Dominio puro y errores unidos:** igual que US-005 (sección 5).
- **Validación compartida.** La validación de `NuevoItem` se extrae a una
  función del paquete que usan `NuevoItem` y `Editar`. Es un refactor dentro
  del PR de implementación, con los tests de US-005 en verde.
- **Sin migración.** Se escriben columnas que ya crea la migración de US-005.
- **Dependencia con US-010.** RN-006-8 es la protección de esta historia contra
  el marcado de US-010. El escenario de CA-006-6 necesita
  `Dado la historia "X" completada`, que conecta US-010: hasta entonces,
  CA-006-6 se cubre con tests unitarios (el dominio con un ítem completado) y de
  integración (el repositorio con un ítem completado armado en la base). Mismo
  criterio que US-008 y US-013.
- **Dependencia con US-008 y US-009.** El escenario de CA-006-3 necesita crear
  un sprint y asignarle la historia. Si esas historias no están en `main` cuando
  se implemente esta, CA-006-3 se cubre con un test de integración y su
  escenario llega con ellas.
- **Frases BDD.** Se reutilizan
  `se cambia la prioridad de la historia "X" a "should"` (CA-006-2) y las de
  resultado de US-005. Se agrega una frase de edición completa:

  | Frase | Expresión |
  |---|---|
  | `Cuando se edita la historia "X" con el título "Y", la prioridad "must", la descripción "D" y los criterios:` + tabla de una columna | `^se edita la historia "([^"]*)" con el título "([^"]*)", la prioridad "([^"]*)", la descripción "([^"]*)" y los criterios:$` |

  La interfaz `Backlog` de `features/steps/mundo.go` suma
  `EditarHistoria(titulo string, historia DatosHistoria) error`.
  `CambiarPrioridadHistoria` se implementa con el mismo caso de uso: el
  adaptador lee los datos actuales y manda los cuatro con la prioridad nueva.
  Una historia que el escenario no creó se edita con un ID que no existe, así
  que el rechazo de CA-006-7 lo produce la aplicación (mismo enfoque que
  `Dado ningún proyecto`).
- **Fuera de alcance.** El orden fino dentro de una prioridad, editar Story
  Points (US-013), cambiar el estado (US-010), el historial de cambios y la
  pantalla (T-011).

## 6. Casos límite

- **CL-006-1** — Editar con los mismos datos que ya tenía, o repriorizar a la
  misma prioridad: es válido y el ítem queda igual (como RN-002-6).
- **CL-006-2** — Lista de criterios vacía: reemplaza a la anterior y es válida.
  Con un criterio vacío, se rechaza con `ErrCriterioVacio` y su posición
  (RN-005-8).
- **CL-006-3** — Título de 120 caracteres, de 121, con espacios en los extremos
  o con solo espacios: igual que CL-005-1 a CL-005-4.
- **CL-006-4** — Ítem asignado a un sprint abierto, editado sin criterios: se
  acepta y sigue en el sprint (RN-006-6).
- **CL-006-5** — Ítem `en_progreso`: se edita.
- **CL-006-6** — Ítem completado y, además, datos inválidos: solo
  `ErrItemCompletado` (RN-006-5).
- **CL-006-7** — El ítem pasa a completado entre la lectura y el `UPDATE`: no se
  actualiza ninguna fila y se rechaza con `ErrItemCompletado` (RN-006-8). Se
  verifica llamando al repositorio con un ítem ya completado en la base.
- **CL-006-8** — Dos ediciones simultáneas del mismo ítem: las dos terminan bien
  y queda la última (RN-006-9).
- **CL-006-9** — Ítem inexistente o ID mal formado, con datos también
  inválidos: solo `ErrItemInexistente` (RN-006-7).

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) |
|---|---|---|
| Datos inválidos | Los de US-005 (`ErrTituloVacio`, `ErrTituloMuyLargo`, `ErrPrioridadInvalida`, `ErrCriterioVacio`), unidos en el orden de RN-005-13 | Los de la sección 7 de US-005 |
| El ítem está completado | `backlog.ErrItemCompletado` (el de US-013), envuelto indicando el ID | El mensaje general (cambio pedido en el #93) |
| El ítem no existe | `app.ErrItemInexistente` (el de US-013 y US-019), envuelto indicando el ID | "el item no existe" |

Los errores de datos se unen entre sí como en US-005. `ErrItemCompletado` y
`ErrItemInexistente` se devuelven solos, sin unirse con los de datos. Todos son
valores de paquete comparables con `errors.Is`, y en ningún caso se modifica el
ítem.

## 8. Criterios de aceptación

- **CA-006-1** (normal) — Al editar el título, la descripción, la prioridad y
  los criterios de un ítem pendiente, queda con los datos nuevos y conserva su
  número, su estado y sus Story Points.
- **CA-006-2** (alternativo) — Al repriorizar un ítem, cambia su prioridad y el
  resto queda igual.
- **CA-006-3** (alternativo) — Un ítem asignado a un sprint abierto se edita y
  sigue en ese sprint.
- **CA-006-4** (límite) — Un título de 120 caracteres se acepta y uno de 121 se
  rechaza; una lista de criterios vacía reemplaza a la anterior.
- **CA-006-5** (error) — Con datos inválidos (título vacío, prioridad fuera de
  MoSCoW o un criterio vacío) se rechaza y el ítem queda como estaba.
- **CA-006-6** (error) — Un ítem completado no se edita.
- **CA-006-7** (error) — Un ítem inexistente se rechaza y no se modifica nada.

---

## Decisiones tomadas y descartadas

- **Edición completa, no parcial.** Mismo criterio que US-002: con la
  descripción opcional, un `""` en una edición parcial sería ambiguo (¿borrar o
  no tocar?). Se descartó la edición campo por campo.
- **Se reutiliza la validación de US-005.** Las reglas de un ítem viven en un
  solo lugar: si cambia el largo máximo del título, cambia para crear y para
  editar a la vez. Se descartó una validación propia de la edición.
- **Repriorizar es cambiar la prioridad MoSCoW.** No hay orden fino dentro de
  cada prioridad. Se descartó una columna de posición con sus reordenamientos:
  es mucho para una historia de 3 SP. Si hace falta, es una historia nueva.
- **Número, estado y Story Points no se editan acá.** El número es la referencia
  fija del ítem (US-005); el estado lo cambia US-010 y los Story Points, US-013.
- **Un ítem completado no se edita.** Misma razón que RN-013-4: ya forma parte
  de un sprint medido (US-025 y US-026) y de sus reportes. Se descartó permitir
  editar solo el texto: dos reglas distintas para un mismo estado no aportan.
- **Se reutiliza `ErrItemCompletado` de US-013.** La condición es la misma: el
  ítem está completado. Usa el mensaje general (cambio pedido en el #93).
- **`UPDATE` condicionado en lugar de bloqueo.** Una sola sentencia atómica
  alcanza para no editar un ítem que se completó en el medio, sin bloquear
  filas ni abrir una transacción. Es la misma protección que necesita US-013
  (dependencia anotada en US-010). Se descartó `SELECT ... FOR UPDATE` sobre el
  ítem: más código para el mismo resultado.
- **Gana la última edición.** Dos personas editando el mismo ítem a la vez es un
  caso raro, y la última versión es la esperable. Se descartó el bloqueo
  optimista (columna de versión y conflicto al guardar) por alcance.
- **Un ítem en un sprint abierto se edita.** El refinamiento sigue durante el
  sprint. Consecuencia: la DoR no se vuelve a verificar, y un ítem asignado
  puede quedar sin criterios.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia y de las decisiones del equipo, usando las specs de US-005,
  US-002 y US-013 como modelo
- Revisado y entendido por:
