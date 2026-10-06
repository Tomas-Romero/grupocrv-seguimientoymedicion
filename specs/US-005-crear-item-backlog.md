# US-005 — Crear un ítem de Product Backlog

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #7 |
| **Escenarios BDD** | `features/US-005-crear-item-backlog.feature` |
| **Código** | `internal/domain/backlog/`, caso de uso en `internal/app/`, repositorio en `internal/adapters/postgres/`, migración nueva en `migraciones/` |
| **Última actualización** | 2026-10-06 |

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
  del ítem y la fecha de creación, valida todo y devuelve un `ItemBacklog`
  válido o todos los errores encontrados. No genera IDs, no sabe si el proyecto
  existe ni qué número le toca.
- **Aplicación y persistencia** (`internal/app` y `internal/adapters/postgres`):
  el caso de uso sigue este orden:
  1. Verifica con una lectura simple que el proyecto exista. Si no existe,
     devuelve `ErrProyectoInexistente` (el equivalente a un 404) sin validar
     nada más.
  2. Valida los datos con el dominio, que devuelve todos los errores unidos.
  3. Registra el ítem a través del repositorio, que dentro de una transacción
     bloquea el proyecto con `FOR NO KEY UPDATE`, vuelve a verificar que exista
     (por si desapareció entre el paso 1 y este), le asigna el número
     correlativo e inserta el ítem. El ID lo genera la base.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. No es un campo del formulario: viene del contexto de navegación (la URL). Lo verifica el caso de uso antes de validar los datos, y otra vez el repositorio dentro de la transacción; nunca el dominio |
| `titulo` | `string` | Sí | Después de recortar los extremos con `strings.TrimSpace`: entre 1 y 120 caracteres, contados en runas. Se aceptan saltos de línea internos |
| `descripcion` | `string` | No | Texto libre, sin largo máximo. Se recorta con `strings.TrimSpace`; si queda vacía, es válida |
| `prioridad` | `Prioridad` | Sí | Exactamente `must`, `should`, `could` o `wont`, en minúscula |
| `criterios` | `[]string` | No (la lista puede estar vacía) | Cada uno no vacío después de recortarlo con `strings.TrimSpace`. Sin largo ni cantidad máxima; se permiten duplicados |
| `creadoEn` | `time.Time` | Sí | Fecha de creación; la pasa quien llama al dominio |

No son entradas: el ID (lo genera la base), el estado (siempre `pendiente`), los
Story Points (siempre sin estimar) ni el número correlativo (lo asigna la
persistencia).

## 3. Salidas esperadas

Un `ItemBacklog` con estos atributos:

| Atributo | Valor al crear |
|---|---|
| ID | UUID generado por la base (`DEFAULT gen_random_uuid()`) y devuelto con `RETURNING` |
| Número | Correlativo dentro del proyecto (#1, #2, …). Lo asigna la persistencia |
| Proyecto | El proyecto recibido |
| Título | El recibido, recortado |
| Descripción | La recibida, recortada; vacía si se omitió o tenía solo espacios |
| Prioridad | La recibida |
| Estado | `pendiente` |
| Story Points | Sin estimar: valor ausente, no `0` |
| Criterios de aceptación | Los recibidos, recortados y en el mismo orden (la lista puede estar vacía) |
| Fecha de creación | La recibida |

El dominio devuelve el ítem todavía sin ID ni número; el caso de uso devuelve el
ítem ya registrado, con los dos. Si hay errores de validación, el dominio no
devuelve ítem: devuelve un único error que une todos los encontrados
(RN-005-13).

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-005-1** (dominio) — El título se recorta con `strings.TrimSpace`
  (espacios, tabulaciones, saltos de línea y demás espacios Unicode de los
  extremos) antes de validarlo, y lo que se guarda es el título recortado. Los
  saltos de línea internos se aceptan tal cual.
- **RN-005-2** (dominio) — El título recortado no puede quedar vacío.
- **RN-005-3** (dominio) — El título recortado tiene como máximo 120
  caracteres, contados en runas (`utf8.RuneCountInString`) tal como llegan, sin
  normalizar, y no en bytes: un título con tildes o ñ no se corta antes de
  tiempo.
- **RN-005-4** (dominio) — La descripción es opcional y no tiene largo máximo.
  Se recorta igual que el título; si queda vacía (por ejemplo, porque tenía
  solo espacios), es válida y se guarda vacía.
- **RN-005-5** (dominio) — La prioridad es obligatoria y sigue MoSCoW. El
  dominio acepta solo los cuatro valores exactos, en minúscula: `must`,
  `should`, `could` o `wont`. Cualquier otro valor se rechaza, incluidos
  `"MUST"` y `" must "`.
- **RN-005-6** (dominio) — Todo ítem se crea en estado `pendiente`. Los
  estados posibles de un ítem son `pendiente`, `en_progreso` y `completado`;
  esta historia solo usa el inicial.
- **RN-005-7** (dominio) — Todo ítem se crea sin Story Points: el valor queda
  ausente (sin estimar), no en `0`. Al crear no se pueden cargar Story Points.
- **RN-005-8** (dominio) — Los criterios de aceptación son opcionales. Cada uno
  se recorta igual que el título y tiene que quedar no vacío; cada criterio
  vacío es un error con su posición (RN-005-13 y RN-005-14). Se guardan
  recortados y en el orden en que se recibieron. Se permiten duplicados, y no
  hay largo máximo ni cantidad máxima.
- **RN-005-9** (aplicación y persistencia) — El ítem pertenece a un proyecto
  existente. El caso de uso lo verifica primero, con una lectura simple y antes
  de validar los datos: si el proyecto no existe, devuelve
  `ErrProyectoInexistente` sin validar nada más y no se registra nada. El
  repositorio lo vuelve a verificar dentro de la transacción (RN-005-11), por
  si el proyecto desapareció entre esa lectura y el alta. Ese error nunca se
  une con los errores de validación del dominio (sección 7).
- **RN-005-10** (persistencia) — Cada ítem recibe un número correlativo dentro
  de su proyecto: el primero es el #1 y cada ítem nuevo toma el siguiente. La
  numeración de un proyecto es independiente de la de los demás.
- **RN-005-11** (persistencia) — Dentro de la transacción que registra el
  ítem, el repositorio bloquea la fila del proyecto
  (`SELECT ... FROM proyectos WHERE id = $1 FOR NO KEY UPDATE`). Si no hay
  fila, el proyecto no existe (RN-005-9). Si la hay, el número es el máximo
  actual del proyecto más 1. Una segunda alta simultánea en el mismo proyecto
  espera ese bloqueo y termina bien, con el número siguiente y sin error. Como
  el número se calcula dentro de la transacción, una transacción revertida no
  deja huecos. `UNIQUE (proyecto_id, numero)` queda como red de seguridad. El
  candado es `FOR NO KEY UPDATE` y no `FOR UPDATE`: serializa las altas de
  ítems del mismo proyecto sin bloquear las claves foráneas que apuntan al
  proyecto (ver "Decisiones tomadas y descartadas").
- **RN-005-12** (aplicación) — Si cualquier validación falla, no se registra
  nada: ni el ítem ni sus criterios.
- **RN-005-13** (dominio) — El dominio valida todos los datos y devuelve todos
  los errores juntos, unidos con `errors.Join`, en este orden fijo: primero el
  del título, después el de la prioridad y al final uno por cada criterio
  vacío, por posición ascendente. Cada error unido sigue siendo reconocible con
  `errors.Is`.
- **RN-005-14** (dominio) — Las posiciones de los criterios se cuentan desde 1.
- **RN-005-15** (persistencia) — El ID del ítem es un UUID que genera la base
  (`DEFAULT gen_random_uuid()`) y el repositorio lo devuelve con `RETURNING`.
  El dominio no genera IDs.

## 5. Restricciones

- **Dominio puro.** `internal/domain/backlog` no importa nada fuera de la
  biblioteca estándar, no hace SQL ni HTTP, no genera IDs y no llama a
  `time.Now()`: la fecha de creación entra como parámetro. Sus errores se
  definen solo con la biblioteca estándar, y nunca importa `internal/platform`.
- **Lo que el dominio no puede saber** (que el proyecto exista, qué número le
  toca al ítem, qué ID tiene) lo resuelven la capa de aplicación y el
  repositorio de `internal/adapters/postgres`, en el orden de la sección 2;
  nunca el dominio.
- **Conversión de la prioridad.** Convertir el texto que llega de un
  formulario a `Prioridad` es tarea del adaptador, no del dominio: el dominio no
  pasa a minúscula ni recorta la prioridad.
- **Limitación conocida: no se normaliza Unicode.** El título se cuenta en
  runas tal como llega. Una letra con tilde puede llegar precompuesta (una
  runa, NFC) o descompuesta en letra más tilde combinante (dos runas, NFD); en
  el segundo caso cuenta doble, y un título puede rechazarse antes de mostrar
  120 caracteres. Normalizar requeriría `golang.org/x/text`, y el dominio solo
  usa la biblioteca estándar.
- **Definition of Ready.** Crear un ítem no exige criterios de aceptación ni
  Story Points. La exigencia de tener ambos para entrar a un sprint (Definition
  of Ready) se valida en **US-009**, al asignar el ítem a un sprint, no acá.
- **Sprint.** Pertenecer a un sprint no es un estado del ítem: es una relación
  entre el ítem y el sprint, que modela US-009. Esta historia no agrega ningún
  estado ni dato de sprint.
- **Story Points.** Esta historia no valida la escala Fibonacci: estimar es
  US-013 y Planning Poker.
- **Dependencia con T-005.** El tipo de error común de T-005 vive fuera del
  dominio (en `internal/platform` o en el adaptador HTTP) y traduce los errores
  de dominio a respuestas para el usuario; el dominio no lo conoce. Esa
  traducción tiene que ser compatible con errores unidos con `errors.Join`,
  porque el dominio de esta historia puede devolver varios errores a la vez.
- **Persistencia.** Una migración nueva en `migraciones/` (nunca se edita
  `00001_init.sql`), con la restricción `UNIQUE (proyecto_id, numero)` y los
  Story Points sin estimar guardados como valor nulo, no como `0`. Las tablas
  nuevas generan sus IDs en la base con `DEFAULT gen_random_uuid()`. La tabla
  `proyectos` no tiene ese `DEFAULT`; alinearla queda fuera de alcance y
  corresponde a US-001.
- **Verificación de la concurrencia.** RN-005-11 y CL-005-9 se verifican con un
  test concurrente contra un Postgres real.
- **Fuera de alcance.** La pantalla de alta (depende de T-004), editar y
  repriorizar (US-006), listar y filtrar (US-007), asignar a un sprint (US-009),
  estimar (US-013) y alinear la generación de IDs de `proyectos` (US-001).

## 6. Casos límite

- **CL-005-1** — Título de exactamente 120 caracteres: se acepta. Con 121, se
  rechaza con `ErrTituloMuyLargo`.
- **CL-005-2** — Título de 120 caracteres con espacios al principio y al final:
  se acepta, porque el límite se mide después del recorte, y se guarda sin esos
  espacios.
- **CL-005-3** — Título de 120 caracteres con tildes o ñ precompuestas (más de
  120 bytes): se acepta, porque el límite cuenta caracteres y no bytes. Con 121
  caracteres de ese tipo, se rechaza.
- **CL-005-4** — Título con solo espacios, aunque sean más de 120: es un título
  vacío (`ErrTituloVacio`), no uno demasiado largo, porque se recorta antes de
  validar.
- **CL-005-5** — Sin descripción y con la lista de criterios vacía: se acepta
  (CA-005-2).
- **CL-005-6** — Criterios `["", "válido", "  "]`: dos errores
  `ErrCriterioVacio`, uno por la posición 1 y otro por la 3, en ese orden. No
  se registra ninguno de los criterios.
- **CL-005-7** — Prioridad vacía: está fuera de `must`, `should`, `could` y
  `wont`, así que se rechaza con `ErrPrioridadInvalida`.
- **CL-005-8** — Primer ítem de un proyecto: es el #1, aunque otros proyectos
  ya tengan ítems (CA-005-3).
- **CL-005-9** — Dos altas simultáneas en el mismo proyecto: la segunda espera
  el bloqueo de la primera sobre la fila del proyecto, y las dos terminan bien,
  con números consecutivos y sin error (RN-005-11). Se verifica con un test
  concurrente contra un Postgres real.
- **CL-005-10** — Story Points al crear: un ítem recién creado está siempre sin
  estimar; nunca tiene `0` Story Points.
- **CL-005-11** — Tres errores a la vez: título con solo espacios, prioridad
  `"urgente"` y el criterio 2 vacío. El dominio devuelve los tres unidos, en
  este orden: `ErrTituloVacio`, `ErrPrioridadInvalida` y `ErrCriterioVacio`
  (posición 2). `errors.Is` reconoce cada uno.
- **CL-005-12** — Prioridad `"MUST"` o `" must "`: se rechaza con
  `ErrPrioridadInvalida`.
- **CL-005-13** — Descripción con solo espacios, tabulaciones o saltos de
  línea: es válida y se guarda vacía.
- **CL-005-14** — Criterios repetidos (`["probado", "probado"]`): se aceptan y
  se guardan los dos, en ese orden.
- **CL-005-15** — Título con tabulaciones o saltos de línea al principio o al
  final: se recortan. Con un salto de línea en el medio: se acepta tal cual, y
  el salto cuenta como un carácter para el límite de 120.
- **CL-005-16** — Una transacción que se revierte después de calcular el
  número no deja hueco: la siguiente alta del proyecto toma ese mismo número.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) |
|---|---|---|
| Título vacío o con solo espacios | `backlog.ErrTituloVacio` | "el titulo es obligatorio" |
| Título de más de 120 caracteres después de recortarlo | `backlog.ErrTituloMuyLargo`, envuelto indicando cuántos caracteres tiene | "el titulo no puede tener mas de 120 caracteres" |
| Prioridad distinta de `must`, `should`, `could`, `wont` (exactos, en minúscula) | `backlog.ErrPrioridadInvalida`, envuelto indicando el valor recibido | "la prioridad tiene que ser must, should, could o wont" |
| Un criterio de aceptación vacío o con solo espacios (un error por cada uno) | `backlog.ErrCriterioVacio`, envuelto indicando la posición del criterio, contada desde 1 | "el criterio de aceptacion de la posicion N esta vacio" |
| El proyecto no existe | `app.ErrProyectoInexistente`, envuelto indicando el ID recibido | "el proyecto no existe" |

La tercera columna es el texto de origen de cada error. Es el que ve la
persona mientras T-005 no defina otra traducción.

Los cuatro primeros son del dominio (`internal/domain/backlog`): se detectan
todos en una sola validación, antes de escribir nada en la base, y se
devuelven unidos con `errors.Join` en el orden de RN-005-13. Con varios
errores, el mensaje es la unión de los mensajes, uno por línea. El último lo
detecta la capa de aplicación (`internal/app`) antes de validar los datos, o
el repositorio dentro de la transacción si el proyecto desapareció en el medio
(RN-005-9 y RN-005-11). Todos son valores de paquete comparables con
`errors.Is`, también cuando vienen unidos, y en ningún caso se registra nada.

`ErrProyectoInexistente` **no se une** con los errores de validación y **tiene
prioridad** sobre ellos: el caso de uso verifica el proyecto antes de validar,
y si no existe responde con ese error sin mirar los datos. El proyecto no es un
campo del formulario: viene del contexto de navegación (la URL). Si no existe,
la solicitud es inválida (el equivalente a un 404), no es un error de carga del
ítem que la persona pueda corregir en el formulario.

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
  persistencia y no el dominio, porque depende de los ítems que ya existen en
  el proyecto, que una función pura no conoce. El UUID lo genera la base
  (`DEFAULT gen_random_uuid()`) y se devuelve con `RETURNING`: así el dominio no
  genera IDs, sigue siendo puro y no hace falta ninguna dependencia nueva.
- **El número se asigna con el proyecto bloqueado.**
  `SELECT ... FOR NO KEY UPDATE` sobre la fila del proyecto serializa las altas
  de un mismo proyecto: la segunda espera y termina bien con el número
  siguiente, en vez de fallar. La misma consulta vuelve a verificar que el
  proyecto exista (sin fila, `ErrProyectoInexistente`), y como el número se
  calcula dentro de la transacción, un rollback no deja huecos.
  `UNIQUE (proyecto_id, numero)` queda como red de seguridad. Se descartó
  calcular el máximo más 1 sin bloquear: la segunda alta simultánea podía
  chocar con el `UNIQUE` y fallar.
- **El candado es `FOR NO KEY UPDATE`, no `FOR UPDATE`.** Dos
  `FOR NO KEY UPDATE` sobre la misma fila se excluyen, así que las altas de
  ítems de un proyecto se serializan igual. A diferencia de `FOR UPDATE`, no
  choca con el `FOR KEY SHARE` que toma Postgres al verificar una clave
  foránea: mientras se registra un ítem, se pueden seguir insertando filas que
  solo apuntan al proyecto, como un integrante nuevo. Se descartó `FOR UPDATE`:
  además de serializar las altas, bloqueaba esas claves foráneas sin
  necesidad. Las altas de sprints de US-008 usan el mismo candado, así que esas
  sí esperan.
- **Todos los errores de validación juntos.** El dominio valida todo y devuelve
  los errores unidos con `errors.Join`, para que quien carga el ítem vea de una
  vez todo lo que tiene que corregir, en lugar de un error por intento. El
  orden es fijo (título, prioridad, criterios por posición), así el mensaje y
  los tests son deterministas, y `errors.Is` sigue reconociendo cada error.
- **El proyecto inexistente no se une con los errores de validación.** El
  proyecto no es un campo del formulario: viene del contexto de navegación (la
  URL). Si no existe, es una solicitud inválida, el equivalente a un 404, y no
  un error de carga del ítem.
- **Primero el proyecto, después los datos.** El caso de uso verifica que el
  proyecto exista con una lectura simple antes de validar con el dominio, para
  que el 404 tenga prioridad: si la URL apunta a un proyecto que no existe, no
  tiene sentido devolver errores de un formulario que no se puede cargar ahí.
  El repositorio lo vuelve a verificar con `FOR NO KEY UPDATE` dentro de la
  transacción, porque el proyecto puede desaparecer entre esa lectura y el
  alta.
- **Las posiciones de los criterios se cuentan desde 1.** El mensaje lo lee una
  persona.
- **La prioridad no se normaliza en el dominio.** El dominio acepta solo los
  cuatro valores exactos en minúscula; convertir el texto de un formulario a
  `Prioridad` es tarea del adaptador, no del dominio.
- **Recorte con `strings.TrimSpace`.** Es la función de la biblioteca estándar
  y cubre también las tabulaciones y los saltos de línea de los extremos, que
  puede mandar un formulario. Los saltos de línea internos del título se
  aceptan tal cual.
- **Descripción y criterios sin límites.** Son `TEXT`, y cada límite (largo
  máximo, cantidad máxima de criterios, prohibir duplicados) sería una regla
  más que nadie pidió. Se recortan como el título. Una descripción vacía es
  válida porque la descripción es opcional, y los criterios se guardan en el
  orden en que se cargaron.
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
- **No se normaliza Unicode.** Normalizar requeriría `golang.org/x/text`, y el
  dominio solo usa la biblioteca estándar. Las runas se cuentan tal como
  llegan, y queda como limitación conocida (ver Restricciones).
- **Paquete `internal/domain/backlog`.** Los nombres de paquetes y archivos van
  en español; `backlog` se mantiene porque Backlog es un término de Scrum.
- **La pantalla queda fuera de alcance.** Depende de T-004 (layout base y
  navegación). Esta historia entrega dominio, caso de uso, repositorio y
  migración.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia, decisiones y criterios definidos por el equipo
- Revisado y entendido por: Juan Ignacio Vergara
