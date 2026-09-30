# US-005 — Crear un ítem de Product Backlog

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #7 |
| **Escenarios BDD** | `features/US-005-crear-item-backlog.feature` |
| **Código** | `internal/domain/backlog/`, caso de uso en `internal/app/`, repositorio en `internal/adapters/postgres/`, migración nueva en `migrations/` |
| **Última actualización** | 2026-09-29 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero cargar un ítem en el Product Backlog del
> proyecto con título, descripción, prioridad y criterios de aceptación, para
> tener registrado el trabajo pendiente antes de planificarlo.

Registrar trabajo pendiente como ítems del Product Backlog de un proyecto, con
prioridad MoSCoW y, si ya se conocen, sus criterios de aceptación. Es la base de
las historias que trabajan sobre el backlog: editar y repriorizar (US-006),
listar y filtrar (US-007), asignar a un sprint (US-009) y estimar (US-013).

## 2. Entradas

La creación pasa por dos capas, y cada una valida solo lo que puede saber:

- **Dominio** (`internal/domain/backlog`): función pura que recibe los datos
  del ítem y la fecha de creación, y devuelve un `ItemBacklog` válido o un
  error. No sabe si el proyecto existe ni qué número le toca.
- **Aplicación y persistencia** (`internal/app` y `internal/adapters/postgres`):
  el caso de uso verifica que el proyecto exista y registra el ítem; el
  repositorio le asigna el número correlativo dentro de la transacción.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. Lo valida la aplicación, no el dominio |
| `titulo` | `string` | Sí | Después de recortar los espacios del principio y del final: entre 1 y 120 caracteres, contados en runas |
| `descripcion` | `string` | No | Texto libre |
| `prioridad` | `Prioridad` | Sí | `must`, `should`, `could` o `wont` |
| `criterios` | `[]string` | No (la lista puede estar vacía) | Si se cargan, cada uno no vacío después de recortar sus espacios |
| `creadoEn` | `time.Time` | Sí | Fecha de creación; la pasa quien llama al dominio |

No son entradas: el estado (siempre `pendiente`), los Story Points (siempre sin
estimar) ni el número correlativo (lo asigna la persistencia).

## 3. Salidas esperadas

Un `ItemBacklog` con estos atributos:

| Atributo | Valor al crear |
|---|---|
| ID | UUID interno |
| Número | Correlativo dentro del proyecto (#1, #2, …). Lo asigna la persistencia |
| Proyecto | El proyecto recibido |
| Título | El recibido, recortado |
| Descripción | La recibida, o vacía si se omitió |
| Prioridad | La recibida |
| Estado | `pendiente` |
| Story Points | Sin estimar: valor ausente, no `0` |
| Criterios de aceptación | Los recibidos (la lista puede estar vacía) |
| Fecha de creación | La recibida |

El dominio devuelve el ítem todavía sin número; el caso de uso devuelve el ítem
ya registrado, con su número asignado.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-005-1** (dominio) — El título se recorta (espacios al principio y al
  final) antes de validarlo, y lo que se guarda es el título recortado.
- **RN-005-2** (dominio) — El título recortado no puede quedar vacío.
- **RN-005-3** (dominio) — El título recortado tiene como máximo 120
  caracteres, contados en runas (`utf8.RuneCountInString`), no en bytes: un
  título con tildes o ñ no se corta antes de tiempo.
- **RN-005-4** (dominio) — La descripción es opcional.
- **RN-005-5** (dominio) — La prioridad es obligatoria y sigue MoSCoW: `must`,
  `should`, `could` o `wont`. Cualquier otro valor se rechaza.
- **RN-005-6** (dominio) — Todo ítem se crea en estado `pendiente`. Los
  estados posibles de un ítem son `pendiente`, `en_progreso` y `completado`;
  esta historia solo usa el inicial.
- **RN-005-7** (dominio) — Todo ítem se crea sin Story Points: el valor queda
  ausente (sin estimar), no en `0`. Al crear no se pueden cargar Story Points.
- **RN-005-8** (dominio) — Los criterios de aceptación son opcionales. Si se
  cargan, cada uno tiene que quedar no vacío después de recortar sus espacios;
  si alguno está vacío, se rechaza la creación entera indicando la posición de
  ese criterio.
- **RN-005-9** (aplicación) — El ítem pertenece a un proyecto existente. Si el
  proyecto no existe, se rechaza y no se registra nada.
- **RN-005-10** (persistencia) — Cada ítem recibe un número correlativo dentro
  de su proyecto: el primero es el #1 y cada ítem nuevo toma el siguiente. La
  numeración de un proyecto es independiente de la de los demás.
- **RN-005-11** (persistencia) — El número se asigna dentro de la misma
  transacción que registra el ítem, y la base garantiza con
  `UNIQUE (proyecto_id, numero)` que dos ítems de un mismo proyecto no
  compartan número.
- **RN-005-12** (aplicación) — Si cualquier validación falla, no se registra
  nada: ni el ítem ni sus criterios.

## 5. Restricciones

- **Dominio puro.** `internal/domain/backlog` no importa nada fuera de la
  biblioteca estándar, no hace SQL ni HTTP y no llama a `time.Now()`: la fecha
  de creación entra como parámetro.
- **Lo que el dominio no puede saber** (que el proyecto exista, qué número le
  toca al ítem) se valida en la capa de aplicación y en el repositorio de
  `internal/adapters/postgres`, nunca en el dominio.
- **Definition of Ready.** Crear un ítem no exige criterios de aceptación ni
  Story Points. La exigencia de tener ambos para entrar a un sprint (Definition
  of Ready) se valida en **US-009**, al asignar el ítem a un sprint, no acá.
- **Sprint.** Pertenecer a un sprint no es un estado del ítem: es una relación
  entre el ítem y el sprint, que modela US-009. Esta historia no agrega ningún
  estado ni dato de sprint.
- **Story Points.** Esta historia no valida la escala Fibonacci: estimar es
  US-013 y Planning Poker.
- **Persistencia.** Una migración nueva en `migrations/` (nunca se edita
  `00001_init.sql`), con la restricción `UNIQUE (proyecto_id, numero)` y los
  Story Points sin estimar guardados como valor nulo, no como `0`.
- **Fuera de alcance.** La pantalla de alta (depende de T-004), editar y
  repriorizar (US-006), listar y filtrar (US-007), asignar a un sprint (US-009)
  y estimar (US-013).

## 6. Casos límite

- **CL-005-1** — Título de exactamente 120 caracteres: se acepta. Con 121, se
  rechaza con `ErrTituloMuyLargo`.
- **CL-005-2** — Título de 120 caracteres con espacios al principio y al final:
  se acepta, porque el límite se mide después del recorte, y se guarda sin esos
  espacios.
- **CL-005-3** — Título de 120 caracteres con tildes o ñ (más de 120 bytes): se
  acepta, porque el límite cuenta caracteres y no bytes. Con 121 caracteres de
  ese tipo, se rechaza.
- **CL-005-4** — Título con solo espacios, aunque sean más de 120: es un título
  vacío (`ErrTituloVacio`), no uno demasiado largo, porque se recorta antes de
  validar.
- **CL-005-5** — Sin descripción y con la lista de criterios vacía: se acepta
  (CA-005-2).
- **CL-005-6** — Un criterio vacío o con solo espacios entre otros válidos: se
  rechaza la creación entera indicando la posición de ese criterio, y no se
  registra ninguno de los válidos.
- **CL-005-7** — Prioridad vacía: está fuera de `must`, `should`, `could` y
  `wont`, así que se rechaza con `ErrPrioridadInvalida`.
- **CL-005-8** — Primer ítem de un proyecto: es el #1, aunque otros proyectos
  ya tengan ítems (CA-005-3).
- **CL-005-9** — Dos creaciones simultáneas en el mismo proyecto: nunca quedan
  dos ítems con el mismo número (RN-005-11).
- **CL-005-10** — Story Points al crear: un ítem recién creado está siempre sin
  estimar; nunca tiene `0` Story Points.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| Título vacío o con solo espacios | `backlog.ErrTituloVacio` | "el titulo es obligatorio" |
| Título de más de 120 caracteres después de recortarlo | `backlog.ErrTituloMuyLargo`, envuelto indicando cuántos caracteres tiene | "el titulo no puede tener mas de 120 caracteres" |
| Prioridad fuera de `must`, `should`, `could`, `wont` | `backlog.ErrPrioridadInvalida`, envuelto indicando el valor recibido | "la prioridad tiene que ser must, should, could o wont" |
| Un criterio de aceptación vacío o con solo espacios | `backlog.ErrCriterioVacio`, envuelto indicando la posición del criterio | "el criterio de aceptacion de la posicion N esta vacio" |
| El proyecto no existe | `app.ErrProyectoInexistente`, envuelto indicando el ID recibido | "el proyecto no existe" |

Los cuatro primeros los devuelve el dominio (`internal/domain/backlog`), antes
de tocar la base; el último lo devuelve la capa de aplicación (`internal/app`).
Todos son valores de paquete comparables con `errors.Is`, y en ningún caso se
registra nada.

## 8. Criterios de aceptación

- **CA-005-1** (normal) — Con un proyecto existente, al crear un ítem con
  título, descripción, prioridad y criterios válidos, queda registrado en
  estado pendiente, sin Story Points y con el siguiente número correlativo del
  proyecto.
- **CA-005-2** (alternativo) — Se puede crear sin descripción y sin criterios.
- **CA-005-3** (alternativo) — El correlativo es independiente por proyecto; el
  primer ítem de cada proyecto es el #1.
- **CA-005-4** (límite) — Título de 120 caracteres se acepta y de 121 se
  rechaza; espacios al principio y al final se recortan antes de validar y
  guardar.
- **CA-005-5** (error) — Título vacío o solo espacios → rechazo, no se registra
  nada.
- **CA-005-6** (error) — Prioridad fuera de must/should/could/wont → rechazo.
- **CA-005-7** (error) — Un criterio vacío o solo espacios → rechazo indicando
  su posición.
- **CA-005-8** (error) — Proyecto inexistente → rechazo, no se registra nada.

---

## Decisiones tomadas y descartadas

- **Identificador doble: UUID interno y número correlativo por proyecto.** El
  número (#1, #2, …) es la referencia legible dentro del proyecto. Lo asigna la
  persistencia y no el dominio: depende de los ítems que ya existen en el
  proyecto, que una función pura no conoce, y solo dentro de la transacción,
  con `UNIQUE (proyecto_id, numero)`, se puede garantizar que dos altas
  simultáneas no lo repitan.
- **Estar en un sprint no es un estado.** Los estados (`pendiente`,
  `en_progreso`, `completado`) describen el avance del trabajo; la pertenencia a
  un sprint es una relación, y la resuelve US-009.
- **Story Points ausentes al crear, no `0`.** "Sin estimar" y "estimado en 0"
  no significan lo mismo. Estimar es US-013.
- **La Definition of Ready no se exige al crear.** La historia es registrar
  trabajo pendiente antes de planificarlo: exigir criterios y Story Points al
  cargarlo impediría anotar trabajo que todavía no se refinó. Se valida en
  US-009, al asignar a un sprint.
- **El título se cuenta en runas, no en bytes.** `len()` cuenta bytes, y en
  UTF-8 cada tilde o ñ ocupa dos: un título en español se rechazaría antes de
  llegar a los 120 caracteres.
- **La pantalla queda fuera de alcance.** Depende de T-004 (layout base y
  navegación). Esta historia entrega dominio, caso de uso, repositorio y
  migración.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code (los 8
  apartados), a partir de la historia, las decisiones y los criterios de
  aceptación que definió el equipo — revisado por: ...
