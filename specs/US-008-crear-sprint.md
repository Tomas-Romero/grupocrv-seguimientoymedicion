# US-008 — Crear un Sprint con Sprint Goal y fechas

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #9 |
| **Escenarios BDD** | `features/US-008-crear-sprint.feature` |
| **Código** | `internal/domain/sprint/`, caso de uso en `internal/app/`, repositorio en `internal/adapters/postgres/`, migración nueva en `migraciones/` |
| **Última actualización** | 2026-10-06 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero crear un Sprint con su Sprint Goal y sus
> fechas, para planificar el trabajo de un período.

Registrar los sprints de un proyecto, cada uno con su Sprint Goal y sus fechas
de inicio y fin, numerados en el mismo orden en que transcurren. Es la base de
las historias que trabajan sobre sprints: asignar historias (US-009), marcarlas
como completadas (US-010), cerrar el sprint (US-011) y consultar los sprints
anteriores (US-012). Las métricas (US-025 y US-026) y sus escenarios nombran a
cada sprint con el nombre que asigna esta historia: "Sprint 1", "Sprint 2", …

## 2. Entradas

La creación pasa por dos capas, y cada una valida solo lo que puede saber:

- **Dominio** (`internal/domain/sprint`): función pura que recibe el Sprint
  Goal, las fechas del sprint y las fechas del proyecto, valida todo y devuelve
  un `Sprint` válido o todos los errores encontrados. No genera IDs, no sabe si
  el proyecto existe, qué sprints tiene ni qué número le toca al nuevo.
- **Aplicación y persistencia** (`internal/app` y `internal/adapters/postgres`):
  el caso de uso sigue este orden:
  1. Lee el proyecto con una lectura simple. Si no existe, devuelve
     `ErrProyectoInexistente` (el equivalente a un 404) sin validar nada más.
     Si existe, toma de esa misma lectura sus fechas de inicio y fin.
  2. Valida los datos con el dominio, pasándole las fechas del proyecto. El
     dominio devuelve todos los errores unidos.
  3. Registra el sprint a través del repositorio, que dentro de una
     transacción bloquea el proyecto con `FOR NO KEY UPDATE`, vuelve a
     verificar que exista (por si desapareció entre el paso 1 y este) y que las
     fechas del sprint estén dentro de las suyas, lee el último sprint del
     proyecto, verifica que esté cerrado y que el nuevo empiece después de que
     ese termine, le asigna el número correlativo e inserta el sprint. El ID lo
     genera la base.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. No es un campo del formulario: viene del contexto de navegación (la URL). Lo verifica el caso de uso antes de validar los datos, y otra vez el repositorio dentro de la transacción; nunca el dominio |
| `sprintGoal` | `string` | Sí | Después de recortar los extremos con `strings.TrimSpace`: entre 1 y 200 caracteres, contados en runas |
| `fechaInicio` | `time.Time` | Sí | Un día, sin hora. Distinta del valor cero y dentro de las fechas del proyecto |
| `fechaFin` | `time.Time` | Sí | Un día, sin hora. Distinta del valor cero, no anterior a `fechaInicio`, dentro de las fechas del proyecto y con una duración de 28 días como máximo, contando los dos extremos |
| `inicioProyecto`, `finProyecto` | `time.Time` | Sí | Las fechas del proyecto. No las carga la persona: las obtiene el caso de uso en el paso 1 y se las pasa al dominio |

Las fechas son solo día: llegan normalizadas a medianoche UTC (formato
`AAAA-MM-DD` en la interfaz), igual que en US-001, y el dominio solo las compara
y cuenta días. Convertir el texto del formulario a `time.Time` es tarea del
adaptador.

No son entradas: el ID (lo genera la base), el número y el nombre (los asigna la
persistencia: la persona no escribe el nombre) ni el estado (siempre `abierto`).

## 3. Salidas esperadas

Un `Sprint` con estos atributos:

| Atributo | Valor al crear |
|---|---|
| ID | UUID generado por la base (`DEFAULT gen_random_uuid()`) y devuelto con `RETURNING` |
| Número | Correlativo dentro del proyecto (1, 2, …). Lo asigna la persistencia |
| Nombre | `"Sprint N"`, donde N es el número, sin ceros a la izquierda. Se arma a partir del número (RN-008-12) |
| Proyecto | El proyecto recibido |
| Sprint Goal | El recibido, recortado |
| Fecha de inicio | La recibida |
| Fecha de fin | La recibida |
| Estado | `abierto` |

El dominio devuelve el sprint todavía sin ID ni número, y por lo tanto sin
nombre; el caso de uso devuelve el sprint ya registrado, con los tres. Si hay
errores de validación, el dominio no devuelve sprint: devuelve un único error
que une todos los encontrados (RN-008-10).

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-008-1** (dominio) — El Sprint Goal se recorta con `strings.TrimSpace`
  (espacios, tabulaciones, saltos de línea y demás espacios Unicode de los
  extremos) antes de validarlo, y lo que se guarda es el Sprint Goal recortado.
- **RN-008-2** (dominio) — El Sprint Goal recortado no puede quedar vacío.
- **RN-008-3** (dominio) — El Sprint Goal recortado tiene como máximo 200
  caracteres, contados en runas (`utf8.RuneCountInString`) tal como llegan, sin
  normalizar, y no en bytes.
- **RN-008-4** (dominio) — La fecha de inicio y la de fin son obligatorias. Una
  fecha con el valor cero de `time.Time` cuenta como faltante.
- **RN-008-5** (dominio) — La fecha de fin no puede ser anterior a la de inicio.
  Que sean iguales es válido: es un sprint de un día.
- **RN-008-6** (dominio) — El sprint dura como máximo 28 días, contando el día
  de inicio y el de fin:
  `duración = fechaFin − fechaInicio + 1` (en días), y `duración <= 28`.
- **RN-008-7** (dominio, y otra vez en la persistencia) — Las fechas del sprint
  están dentro de las del proyecto, con los extremos incluidos:
  `inicioProyecto <= fechaInicio` y `fechaFin <= finProyecto`. El dominio
  recibe las fechas del proyecto como parámetros (sección 2). El repositorio la
  vuelve a verificar dentro de la transacción, con las fechas del proyecto
  leídas con la fila bloqueada (RN-008-14).
- **RN-008-8** (dominio) — Todo sprint se crea en estado `abierto`. Los estados
  posibles son `abierto` y `cerrado`; esta historia solo usa el inicial, y
  cerrar un sprint es US-011.
- **RN-008-9** (aplicación y persistencia) — El sprint pertenece a un proyecto
  existente. El caso de uso lo verifica primero, con una lectura simple y antes
  de validar los datos: si el proyecto no existe, devuelve
  `ErrProyectoInexistente` (el de US-005) sin validar nada más y no se registra
  nada. El repositorio lo vuelve a verificar dentro de la transacción
  (RN-008-14), por si el proyecto desapareció entre esa lectura y el alta. Ese
  error nunca se une con los errores de validación del dominio (sección 7).
- **RN-008-10** (dominio) — El dominio valida todos los datos y devuelve todos
  los errores juntos, unidos con `errors.Join`, en este orden fijo: Sprint Goal
  (vacío o demasiado largo), falta la fecha de inicio, falta la fecha de fin,
  fechas incoherentes (RN-008-5), duración (RN-008-6) y fuera del proyecto
  (RN-008-7). Esas tres últimas comparan fechas y solo se evalúan si están las
  dos. Cada error unido sigue siendo reconocible con `errors.Is`.
- **RN-008-11** (persistencia) — Cada sprint recibe un número correlativo dentro
  de su proyecto: el primero es el 1 y cada sprint nuevo toma el siguiente. La
  numeración de un proyecto es independiente de la de los demás.
- **RN-008-12** (dominio) — El nombre del sprint es `"Sprint "` seguido de su
  número: `"Sprint 1"`, `"Sprint 2"`, … La persona no lo escribe, y no se guarda
  aparte: se arma a partir del número, así que no puede quedar distinto de él.
- **RN-008-13** (persistencia, con una función del dominio) — Si el último
  sprint del proyecto está cerrado, el sprint nuevo empieza después de que ese
  termina: `fechaInicio > fechaFin del último`. El último es el de número más
  alto; por esta misma regla, es también el que termina más tarde. Si el último
  está abierto, esta comparación no se evalúa: se aplica RN-008-17. Si el
  proyecto no tiene sprints, no hay nada con qué comparar. Los sprints de otros
  proyectos no cuentan.
- **RN-008-14** (persistencia) — El repositorio registra el sprint en una
  transacción, con estos pasos en orden, y se detiene en el primer error:
  1. Bloquea la fila del proyecto y lee sus fechas:
     `SELECT fecha_inicio, fecha_fin FROM proyectos`
     `WHERE id = $1 FOR NO KEY UPDATE`. Si no hay fila, el proyecto no existe
     (RN-008-9).
  2. Vuelve a verificar RN-008-7 con la función del dominio, pasándole las
     fechas del proyecto leídas en el paso 1 y no las de la lectura del caso
     de uso: si una modificación del proyecto (US-002) las cambió en el medio,
     vale la versión bloqueada.
  3. Lee el último sprint del proyecto y verifica RN-008-17 y RN-008-13 con la
     función del dominio.
  4. Asigna como número el máximo actual del proyecto más 1, o 1 si todavía no
     tiene sprints, e inserta el sprint.

  Una segunda alta simultánea en el mismo proyecto espera el bloqueo del paso 1.
  Si la primera se registró, cuando la segunda obtiene el bloqueo el sprint de
  la primera ya existe y está abierto, así que se rechaza por RN-008-17
  (CL-008-15). Como el número se calcula dentro de la transacción, una
  transacción revertida no deja huecos. `UNIQUE (proyecto_id, numero)` queda
  como red de seguridad.
- **RN-008-15** (aplicación) — Si cualquier validación falla, no se registra
  nada.
- **RN-008-16** (persistencia) — El ID del sprint es un UUID que genera la base
  (`DEFAULT gen_random_uuid()`) y el repositorio lo devuelve con `RETURNING`.
  El dominio no genera IDs.
- **RN-008-17** (persistencia, con una función del dominio) — No se puede crear
  un sprint mientras el último sprint del proyecto siga abierto: primero hay
  que cerrarlo (US-011). Si el último está abierto, se devuelve solo
  `ErrUltimoSprintAbierto`, sin comparar fechas (RN-008-13). Si el proyecto no
  tiene sprints, la regla no aplica. Así, un proyecto tiene como máximo un
  sprint abierto, y es siempre el último.

## 5. Restricciones

- **Dominio puro.** `internal/domain/sprint` no importa nada fuera de la
  biblioteca estándar, no hace SQL ni HTTP, no genera IDs y no llama a
  `time.Now()`. Sus errores se definen solo con la biblioteca estándar, y nunca
  importa `internal/platform`.
- **No se compara con la fecha de hoy.** Como el dominio no lee el reloj,
  ninguna regla mira la fecha actual: un sprint puede tener fechas pasadas o
  futuras.
- **Lo que el dominio no puede saber** (que el proyecto exista y cuáles son sus
  fechas, qué sprints tiene, qué número le toca al nuevo, qué ID tiene) lo
  resuelven la capa de aplicación y el repositorio de
  `internal/adapters/postgres`, en el orden de la sección 2; nunca el dominio.
- **Las verificaciones de la transacción son funciones del dominio.**
  `internal/domain/sprint` ofrece una función pura que recibe el inicio del
  sprint nuevo y el último sprint del proyecto (número, estado y fecha de fin).
  Si el último está abierto, devuelve `ErrUltimoSprintAbierto` (RN-008-17); si
  está cerrado y el nuevo no empieza después de su fin, devuelve
  `ErrInicioNoPosteriorAlUltimo` (RN-008-13). El repositorio la llama dentro de
  la transacción, con el último sprint leído con el proyecto bloqueado, y
  verifica RN-008-7 con la función del dominio y las fechas del proyecto leídas
  en la transacción. Cualquier repositorio usa esas mismas funciones: el de
  Postgres y el que usen los escenarios BDD.
- **Nivel de aislamiento.** La transacción usa el nivel por defecto de Postgres
  (`READ COMMITTED`), y el último sprint se lee en una sentencia posterior al
  `FOR NO KEY UPDATE`. Así, una alta que esperó el bloqueo ve el sprint que
  registró la otra, y el `FOR NO KEY UPDATE` devuelve las fechas del proyecto
  que haya guardado una modificación concurrente. Con `REPEATABLE READ`, la
  foto de los datos sería la de antes de esperar el bloqueo, y la segunda alta
  no vería el sprint de la primera.
- **Persistencia.** Una migración nueva en `migraciones/` (nunca se edita
  `00001_init.sql`) crea la tabla `sprints`, con el ID generado en la base
  (`DEFAULT gen_random_uuid()`), la referencia al proyecto, el número, el
  Sprint Goal (`sprint_goal`), las fechas como `DATE` y el estado. Igual que
  `proyectos`, lleva restricciones como red de seguridad:
  `UNIQUE (proyecto_id, numero)`, `CHECK (fecha_fin >= fecha_inicio)` y
  `CHECK (estado IN ('abierto', 'cerrado'))`. Como red de seguridad de
  RN-008-17 lleva además un índice único parcial, que impide dos sprints
  abiertos en el mismo proyecto:
  `CREATE UNIQUE INDEX ... ON sprints (proyecto_id) WHERE estado = 'abierto'`.
  El nombre no se guarda (RN-008-12). La duración, las fechas del proyecto y
  RN-008-13 no se imponen en la base, y RN-008-17 no depende del índice: las
  hacen cumplir el dominio y el repositorio.
- **Verificación de la concurrencia.** RN-008-14 y CL-008-15 se verifican con un
  test concurrente contra un Postgres real.
- **Dependencia con US-002.** RN-008-7 se verifica al crear el sprint dos
  veces: en el caso de uso y otra vez en la transacción, con la fila del
  proyecto bloqueada (RN-008-14). Así, una modificación de las fechas del
  proyecto que llegue entre las dos lecturas no deja pasar un sprint fuera de
  ellas, y mientras dura la transacción el proyecto no se puede modificar. Lo
  que esta historia no cubre es una modificación posterior al alta: US-002 no
  valida las fechas nuevas del proyecto contra sus sprints (lo dejó como
  alcance diferido), así que un proyecto puede quedar con sprints fuera de sus
  fechas.
- **Dependencia con T-005.** El tipo de error común de T-005 vive fuera del
  dominio (en `internal/platform` o en el adaptador HTTP) y traduce los errores
  de dominio a respuestas para el usuario; el dominio no lo conoce. Esa
  traducción tiene que ser compatible con errores unidos con `errors.Join`,
  porque el dominio de esta historia puede devolver varios errores a la vez.
- **Frases BDD.** Se mantienen las frases del diccionario que nombran el sprint
  (`Dado un sprint "Sprint 1" con el objetivo "MVP navegable"` y
  `Cuando se crea el sprint "Sprint 1" con el objetivo "MVP navegable"`). El
  nombre entre comillas es el nombre **esperado**: el step crea el sprint sin
  nombre y, si la creación es exitosa, verifica que el nombre generado
  coincida. Si no coincide, el step falla; no es un rechazo de la operación.
  Se agregan al diccionario cuatro frases: la frase con fechas, en sus dos
  formas, y dos frases más:

  | Frase | Expresión |
  |---|---|
  | `Dado un sprint "Sprint 1" con el objetivo "MVP navegable" del "2026-10-06" al "2026-10-19"` | `^un sprint "([^"]*)" con el objetivo "([^"]*)" del "([^"]*)" al "([^"]*)"$` |
  | `Cuando se crea el sprint "Sprint 1" con el objetivo "MVP navegable" del "2026-10-06" al "2026-10-19"` | `^se crea el sprint "([^"]*)" con el objetivo "([^"]*)" del "([^"]*)" al "([^"]*)"$` |
  | `Dado ningún proyecto` | `^ningún proyecto$` |
  | `Y el sprint "Sprint 1" está abierto` | `^el sprint "([^"]*)" está abierto$` |

  `Dado ningún proyecto` deja el proyecto actual del escenario apuntando a uno
  que no existe. La usa CA-008-10 y sirve también para CA-005-8 de US-005. Para
  que el rechazo lo produzca la aplicación (RN-008-9) y no el step, el
  adaptador de los steps tiene que llegar al caso de uso con un proyecto que no
  existe, en lugar de fallar él mismo al buscarlo. `el sprint "X" está abierto`
  es el espejo de `el sprint "X" está cerrado`, y la usa CA-008-1.

  Como todas las expresiones van con ancla, las frases sin fechas no matchean
  las que tienen fechas. El diccionario, los steps y la interfaz `Sprints` de
  `features/steps/mundo.go` se actualizan en el PR de implementación, junto con
  el `.feature` (regla 6 del diccionario). La interfaz cambia: crear un sprint
  recibe el Sprint Goal y las fechas, ya no el nombre, y devuelve el nombre
  generado.
- **Fechas de las frases sin fechas.** Es una restricción para quien implemente
  los steps. Las dos frases sin fechas (las que usan los ejemplos de métricas)
  generan fechas válidas y consecutivas, así:
  - El primer sprint de un proyecto empieza el día de inicio del proyecto. Los
    steps lo pueden obtener con `ProyectoExiste` del área Proyectos.
  - Cada sprint siguiente empieza el día después de que termina el último
    sprint del proyecto en el escenario, se haya creado con fechas o sin ellas.
  - Cada sprint dura 14 días: termina 13 días después de empezar.
  - Una creación rechazada no cuenta como último sprint.
  - Como el último sprint tiene que estar cerrado para crear el siguiente
    (RN-008-17), el escenario cierra el anterior antes
    (`Y el sprint "Sprint 1" cerrado`), como ya hacen los ejemplos de métricas
    del diccionario.

  Con el proyecto de `Dado un proyecto "Demo"` (del 2026-01-01 al 2026-12-31),
  el Sprint 1 va del 2026-01-01 al 2026-01-14, el Sprint 2 del 2026-01-15 al
  2026-01-28, y así; entran 26 sprints. Si el sprint generado no entra en las
  fechas del proyecto (por ejemplo, en un proyecto más corto creado con fechas),
  la creación se rechaza como cualquier otra y, en un `Dado`, el escenario se
  corta. Ese escenario tiene que usar la frase con fechas.
- **Criterios que dependen de US-011.** CA-008-4 y CA-008-9 necesitan un sprint
  anterior cerrado, y cerrar un sprint es US-011. En el PR de US-008, esos dos
  criterios se cubren con tests unitarios (la función del dominio de RN-008-13,
  con el último sprint cerrado como parámetro) y de integración (el
  repositorio, con un sprint cerrado armado directamente en la base). Sus
  escenarios BDD se agregan en el PR de US-011. CA-008-1 se escribe con el
  Sprint 1, así que no necesita cerrar ningún sprint.
- **Fuera de alcance.** La pantalla de alta (depende de T-004), asignar
  historias (US-009), marcarlas como completadas (US-010), cerrar el sprint
  (US-011), consultar los sprints anteriores (US-012) y editar o cancelar un
  sprint.

## 6. Casos límite

- **CL-008-1** — Sprint de un día (inicio igual a fin): se acepta; dura 1 día
  (CA-008-3).
- **CL-008-2** — Sprint de 28 días (fin = inicio + 27 días): se acepta. Con 29
  (fin = inicio + 28 días), se rechaza con `ErrDuracionExcedida` (CA-008-3).
- **CL-008-3** — Los días se cuentan en el calendario, también en febrero. En un
  proyecto que contenga esas fechas, del 2027-02-01 al 2027-02-28 son 28 días y
  se acepta; del 2028-02-01 al 2028-02-29 (2028 es bisiesto) son 29 y se
  rechaza con `ErrDuracionExcedida`.
- **CL-008-4** — El último sprint del proyecto está cerrado y termina el
  2026-10-19: un sprint que empieza el 2026-10-20 se acepta; uno que empieza el
  2026-10-19 se rechaza con `ErrInicioNoPosteriorAlUltimo` (CA-008-4).
- **CL-008-5** — Sprint completamente anterior al último, sin superponerse (el
  último está cerrado y va del 2026-10-20 al 2026-11-02, y el nuevo va del
  2026-10-06 al 2026-10-19): se rechaza con `ErrInicioNoPosteriorAlUltimo`,
  porque la regla es empezar después del último (RN-008-13). Si se aceptara, el
  sprint de número más alto sería anterior en el tiempo.
- **CL-008-6** — Sprint con exactamente las fechas del proyecto: se acepta
  (CA-008-5). Como el sprint dura 28 días como máximo, el caso necesita un
  proyecto de 28 días o menos.
- **CL-008-7** — Sprint que empieza el día anterior al inicio del proyecto, o
  que termina el día siguiente al fin del proyecto: se rechaza con
  `ErrFueraDelProyecto`. Es un solo error, aunque las dos fechas estén afuera.
- **CL-008-8** — Primer sprint de un proyecto: es el "Sprint 1" y no se compara
  con ningún otro, aunque otro proyecto tenga sprints en las mismas fechas o un
  sprint abierto (CA-008-2).
- **CL-008-9** — Sprint Goal de exactamente 200 caracteres: se acepta. Con 201,
  se rechaza con `ErrSprintGoalMuyLargo`. Con 200 caracteres con tildes o ñ
  precompuestas (más de 200 bytes), se acepta. Con espacios al principio y al
  final, el límite se mide después del recorte.
- **CL-008-10** — Sprint Goal con solo espacios, tabulaciones o saltos de línea,
  aunque sean más de 200: es un Sprint Goal vacío (`ErrSprintGoalVacio`), no
  uno demasiado largo, porque se recorta antes de validar (CA-008-6).
- **CL-008-11** — Falta una de las dos fechas: se devuelve el error de la que
  falta, y las validaciones que comparan fechas (RN-008-5 a RN-008-7) no se
  evalúan, porque no hay con qué comparar.
- **CL-008-12** — Varios errores a la vez: en un proyecto del 2026-10-06 al
  2026-11-02, un sprint con el Sprint Goal vacío, del 2026-10-10 al 2026-10-01.
  El dominio devuelve `ErrSprintGoalVacio`, `ErrFechasIncoherentes` y
  `ErrFueraDelProyecto`, en ese orden, y `errors.Is` reconoce cada uno. No
  devuelve `ErrDuracionExcedida`: con las fechas invertidas, la duración da 0 o
  menos.
- **CL-008-13** — Datos inválidos con el último sprint abierto, o en un sprint
  que además empieza antes del fin del último: solo se devuelven los errores del
  dominio. RN-008-17 y RN-008-13 no llegan a verificarse, porque el repositorio
  solo se llama si el dominio no encontró errores; al corregir los datos, el
  siguiente intento puede rechazarse por esas reglas.
- **CL-008-14** — Proyecto inexistente y datos inválidos a la vez: solo se
  devuelve `ErrProyectoInexistente` (CA-008-10).
- **CL-008-15** — Dos altas simultáneas en el mismo proyecto: la segunda espera
  el bloqueo de la primera sobre la fila del proyecto. Si la primera se
  registró, cuando la segunda obtiene el bloqueo el sprint de la primera ya
  existe y está abierto: la segunda se rechaza siempre con
  `ErrUltimoSprintAbierto`, aunque sus fechas empiecen después. Ninguna termina
  con un error de la restricción `UNIQUE` ni del índice de sprints abiertos
  (RN-008-14). Se verifica con un test concurrente contra un Postgres real.
- **CL-008-16** — Una transacción que se revierte después de calcular el número
  no deja hueco: la siguiente alta del proyecto toma ese mismo número.
- **CL-008-17** — El último sprint del proyecto está abierto y el nuevo, además,
  empieza antes de que ese termine: se devuelve solo `ErrUltimoSprintAbierto`;
  la comparación de fechas de RN-008-13 no se evalúa (CA-008-11).
- **CL-008-18** — Las fechas del proyecto cambian entre la lectura del caso de
  uso y el alta, y el sprint queda fuera de las nuevas: el repositorio lo
  detecta con la fila bloqueada y lo rechaza con `ErrFueraDelProyecto`, sin
  ningún otro error (RN-008-14). Se verifica llamando al repositorio con un
  sprint fuera de las fechas guardadas del proyecto.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) |
|---|---|---|
| Sprint Goal vacío o con solo espacios | `sprint.ErrSprintGoalVacio` | "el Sprint Goal es obligatorio" |
| Sprint Goal de más de 200 caracteres después de recortarlo | `sprint.ErrSprintGoalMuyLargo`, envuelto indicando cuántos caracteres tiene | "el Sprint Goal no puede tener mas de 200 caracteres" |
| Falta la fecha de inicio | `sprint.ErrFechaInicioFaltante` | "falta la fecha de inicio" |
| Falta la fecha de fin | `sprint.ErrFechaFinFaltante` | "falta la fecha de fin" |
| La fecha de fin es anterior a la de inicio | `sprint.ErrFechasIncoherentes` | "fecha de fin anterior a la de inicio" |
| El sprint dura más de 28 días, contando el inicio y el fin | `sprint.ErrDuracionExcedida`, envuelto indicando cuántos días dura | "el sprint no puede durar mas de 28 dias" |
| Alguna de las fechas queda fuera de las del proyecto | `sprint.ErrFueraDelProyecto`, envuelto indicando las fechas del proyecto | "las fechas del sprint tienen que estar dentro de las del proyecto" |
| El último sprint del proyecto sigue abierto | `sprint.ErrUltimoSprintAbierto`, envuelto indicando el nombre del último sprint | "no se puede crear un sprint mientras el ultimo sigue abierto: primero hay que cerrarlo" |
| El último sprint del proyecto está cerrado y el nuevo empieza el mismo día en que ese termina, o antes | `sprint.ErrInicioNoPosteriorAlUltimo`, envuelto indicando el nombre y la fecha de fin del último sprint | "el sprint tiene que empezar despues de que termine el ultimo sprint del proyecto" |
| El proyecto no existe | `app.ErrProyectoInexistente` (el de US-005), envuelto indicando el ID recibido | "el proyecto no existe" |

La tercera columna es el texto de origen de cada error. Es el que ve la
persona mientras T-005 no defina otra traducción.

Los siete primeros son del dominio (`internal/domain/sprint`): se detectan
todos en una sola validación, antes de escribir nada en la base, y se
devuelven unidos con `errors.Join` en el orden de RN-008-10. Con varios
errores, el mensaje es la unión de los mensajes, uno por línea.
`ErrFueraDelProyecto` también puede devolverlo el repositorio, solo, cuando la
segunda verificación de RN-008-7 encuentra que las fechas del proyecto
cambiaron (RN-008-14 y CL-008-18).

`ErrUltimoSprintAbierto` y `ErrInicioNoPosteriorAlUltimo` también son valores
del dominio, pero los devuelve el repositorio dentro de la transacción
(RN-008-13, RN-008-14 y RN-008-17), porque solo ahí se ve el último sprint con
el proyecto bloqueado. Nunca vienen juntos: si el último está abierto, se
devuelve solo `ErrUltimoSprintAbierto`. Tampoco se unen con los errores de
validación: el repositorio solo se llama si el dominio no encontró ninguno
(CL-008-13).

`ErrProyectoInexistente` lo detecta la capa de aplicación (`internal/app`)
antes de validar los datos, o el repositorio dentro de la transacción si el
proyecto desapareció en el medio (RN-008-9 y RN-008-14). **No se une** con los
errores de validación y **tiene prioridad** sobre ellos: el caso de uso
verifica el proyecto antes de validar, y si no existe responde con ese error
sin mirar los datos. El proyecto no es un campo del formulario: viene del
contexto de navegación (la URL). Si no existe, la solicitud es inválida (el
equivalente a un 404), no es un error de carga del sprint que la persona pueda
corregir en el formulario.

Todos son valores de paquete comparables con `errors.Is`, también cuando vienen
unidos, y en ningún caso se registra nada.

## 8. Criterios de aceptación

- **CA-008-1** (normal) — Con un proyecto existente, crear un sprint con
  objetivo y fechas válidas lo deja abierto, nombrado "Sprint N" con el número
  siguiente del proyecto.
- **CA-008-2** (alternativo) — El primer sprint de cada proyecto es "Sprint 1";
  la numeración es independiente por proyecto.
- **CA-008-3** (límite) — Un sprint de 1 día es válido; uno de 28 días se acepta
  y uno de 29 se rechaza.
- **CA-008-4** (límite) — Empezar el día siguiente al fin del último sprint es
  válido; empezar el mismo día en que termina se rechaza.
- **CA-008-5** (límite) — Un sprint que ocupa exactamente las fechas del
  proyecto es válido.
- **CA-008-6** (error) — Objetivo vacío o solo espacios → rechazo.
- **CA-008-7** (error) — Fecha de fin anterior a la de inicio → rechazo.
- **CA-008-8** (error) — Fechas fuera del rango del proyecto → rechazo.
- **CA-008-9** (error) — Empezar antes de que termine el último sprint →
  rechazo.
- **CA-008-10** (error) — Proyecto inexistente → rechazo, con prioridad sobre
  los demás errores.
- **CA-008-11** (error) — Con el último sprint del proyecto abierto, crear otro
  → rechazo, indicando que primero hay que cerrarlo.

---

## Decisiones tomadas y descartadas

- **Paquete `internal/domain/sprint`.** Los paquetes van en español (ADR 0003),
  y `sprint` se mantiene porque Sprint es un término de Scrum de la lista del
  ADR. Por la misma razón, el campo y los errores dicen `SprintGoal`, aunque las
  frases BDD digan "objetivo".
- **Nombre automático "Sprint N".** La persona no escribe el nombre. El número
  lo asigna la persistencia, igual que el número de los ítems en US-005, porque
  depende de los sprints que ya tiene el proyecto, que una función pura no
  conoce. Así no hay nombres repetidos, vacíos ni mal escritos, y el nombre
  dice el orden del sprint. Se descartó el nombre libre: necesitaría sus propias
  reglas (obligatorio, largo máximo, único en el proyecto) y no diría nada del
  orden. Consecuencia aceptada: al cargar en la app los sprints del propio
  equipo (T-009), el Sprint 0 del plan de trabajo queda como "Sprint 1", y los
  nombres quedan corridos uno respecto de las actas.
- **El nombre se arma a partir del número.** Guardarlo aparte sería guardar dos
  veces el mismo dato, con el riesgo de que no coincidan.
- **El Sprint Goal es obligatorio.** En Scrum, el Sprint Goal es el compromiso
  del Sprint Backlog: un sprint sin objetivo no tiene contra qué evaluarse en la
  Review.
  Se recorta con `strings.TrimSpace` y se cuenta en runas por las mismas razones
  que el título de US-005: el recorte cubre las tabulaciones y los saltos de
  línea de los extremos, y una tilde o una ñ cuentan como un carácter. El
  máximo de 200 caracteres alcanza para una o dos oraciones: un objetivo más
  largo deja de ser un objetivo.
- **Fechas obligatorias y solo día.** Los sprints se planifican por días (en
  este equipo, de martes a lunes): la hora no aporta y complicaría comparar y
  contar días. Coincide con las columnas `DATE` de `proyectos` y con cómo llegan
  las fechas en US-001.
- **Fin igual a inicio es válido.** Un sprint de un día es raro, pero no es un
  error, igual que un proyecto de un día en US-001.
- **Máximo de 28 días, contando los dos extremos.** La Guía de Scrum fija
  sprints de un mes o menos. Se toma 28 porque es el mes más corto: un límite
  fijo en días no depende del mes en que empieza el sprint y se prueba igual en
  cualquier fecha. Se cuentan los dos extremos (`fin − inicio + 1`) porque un
  sprint de un día dura un día, no cero. Se descartó "un mes calendario": el
  límite variaría entre 28 y 31 días según el mes.
- **Dentro de las fechas del proyecto, con los extremos incluidos.** Un sprint
  es parte del proyecto, así que no puede empezar antes ni terminar después. Los
  extremos se incluyen para que un sprint pueda empezar el día en que empieza el
  proyecto y terminar el día en que termina (CA-008-5).
- **Empezar después de que termine el último sprint.** Con la numeración por
  orden de creación, esta regla garantiza que el número coincide con el orden
  en el tiempo: el Sprint N+1 siempre es posterior al Sprint N. Además, la regla
  de no superposición solo tiene que comparar con el último sprint, no con
  todos. Por eso también se rechaza un sprint completamente anterior al último,
  aunque no se superponga (CL-008-5), y uno que empieza el mismo día en que
  termina el último, porque ese día todavía es del último sprint. Se descartó
  comparar contra todos los sprints del proyecto: permitiría cargar un sprint
  anterior a los existentes, con un número que no respeta el orden en el
  tiempo.
- **Un sprint a la vez.** No se puede crear un sprint mientras el último siga
  abierto: primero hay que cerrarlo (US-011). En Scrum hay un sprint a la vez;
  así, "el sprint en curso" no es ambiguo (el burndown de US-031 lo necesita),
  y los ejemplos de métricas del diccionario ya cierran un sprint antes de crear
  el siguiente. Si el último está abierto, se devuelve solo ese error y no se
  comparan fechas: mientras no se cierre, ningún cambio en las fechas permite
  crear el sprint. Se descartó permitir varios sprints abiertos: obligaría a
  definir cuál está "en curso" con la fecha de hoy, que el dominio no puede
  leer.
- **Estados abierto y cerrado.** Un sprint se crea abierto y se cierra en
  US-011, que además devuelve al backlog las historias no completadas. La
  velocidad del equipo (US-026) se calcula solo sobre sprints cerrados.
- **Todos los errores de validación juntos.** Como en US-005: quien crea el
  sprint ve de una vez todo lo que tiene que corregir, en lugar de un error por
  intento. El orden es fijo, así el mensaje y los tests son deterministas, y
  `errors.Is` sigue reconociendo cada error.
- **Cada capa valida lo que puede saber.** El dominio valida todo lo que se
  decide con los datos del sprint y las fechas del proyecto, sin acceder a la
  base. El caso de uso verifica primero que el proyecto exista y, en la misma
  lectura, obtiene sus fechas. El repositorio verifica RN-008-17 y RN-008-13 y
  asigna el número dentro de la transacción, con el proyecto bloqueado: el
  último sprint puede cambiar entre la lectura del caso de uso y el alta, y dos
  altas simultáneas verificadas fuera de la transacción podrían pasar las dos.
  Por esa misma carrera se descartó pasarle al dominio el último sprint leído
  antes de la transacción.
- **RN-008-7 se verifica otra vez en la transacción.** Las fechas del proyecto
  también pueden cambiar entre la lectura del caso de uso y el alta, si alguien
  modifica el proyecto (US-002) en el medio. Con la fila bloqueada, las fechas
  leídas en la transacción no cambian hasta que termine, así que verificarlas
  ahí cierra esa ventana. Cuesta poco, porque el `FOR NO KEY UPDATE` ya lee esa
  fila, y se usa la misma función del dominio para no duplicar la regla.
- **El proyecto se bloquea con `FOR NO KEY UPDATE`.** Dos `FOR NO KEY UPDATE`
  sobre la misma fila se excluyen, así que las altas de sprints de un proyecto
  se serializan igual, y una modificación del proyecto (US-002) también espera,
  porque su `UPDATE` toma ese mismo candado. A diferencia de `FOR UPDATE`, no
  choca con el `FOR KEY SHARE` que toma Postgres al verificar una clave foránea:
  mientras se registra un sprint, se pueden seguir insertando filas que solo
  apuntan al proyecto, como un integrante nuevo. Las altas de ítems de US-005
  usan el mismo candado, así que esas sí esperan. Se descartó `FOR UPDATE`:
  además de serializar las altas, bloqueaba esas claves foráneas sin
  necesidad.
- **En la transacción, primero las fechas del proyecto.** El paso 2 de
  RN-008-14 va antes que RN-008-17 y RN-008-13 por dos razones. Usa solo la
  fila que se acaba de bloquear: si falla, la transacción termina sin leer los
  sprints. Y mantiene el orden del caso de uso, donde RN-008-7 se verifica en el
  dominio antes de llamar al repositorio: un sprint fuera de las fechas del
  proyecto se rechaza con `ErrFueraDelProyecto` aunque el último sprint esté
  abierto, tanto si lo detecta el caso de uso como si lo detecta la
  transacción porque las fechas cambiaron en el medio.
- **Primero el proyecto, después los datos.** Igual que en US-005: si la URL
  apunta a un proyecto que no existe, es una solicitud inválida (el equivalente
  a un 404) y no un error del formulario, así que tiene prioridad sobre los
  demás. Además, sin proyecto no hay fechas contra las cuales validar el sprint.
- **Se reutiliza `ErrProyectoInexistente`.** Es el mismo caso que en US-005 y
  US-002. Un segundo error para decir lo mismo obligaría a T-005 a traducir los
  dos.
- **RN-008-17 y RN-008-13 son una función del dominio.** El repositorio aporta
  el dato (el último sprint, leído con el proyecto bloqueado) y el dominio
  decide. Así las reglas se prueban con tests unitarios, incluidos los bordes
  de CA-008-4, y no se duplican en el repositorio que usen los escenarios BDD,
  que se prueban contra la capa de aplicación y no contra Postgres (ADR 0002).
  Se descartó definir los errores en `internal/app`, como
  `ErrProyectoInexistente`: aquel no es una regla del sprint, estos sí.
- **Se permiten fechas pasadas.** Ninguna regla compara con la fecha de hoy:
  hace falta para cargar en la app los sprints del propio equipo, que ya
  transcurrieron (T-009). Además, el dominio no lee el reloj.
- **Mensajes de fechas iguales a los de US-001.** "falta la fecha de inicio",
  "falta la fecha de fin" y "fecha de fin anterior a la de inicio" dicen lo
  mismo en proyectos y en sprints. Los errores son valores propios del paquete
  `sprint`, para que el dominio de sprint no dependa del de proyecto.
- **Las frases BDD con nombre se mantienen.** El diccionario y su ejemplo de
  US-026 ya usan `un sprint "Sprint 1" con el objetivo "..."`; mantenerlas evita
  reescribir esos escenarios. Como el nombre lo genera la aplicación, el de la
  frase pasa a ser el esperado, y cada creación en un escenario verifica también
  la numeración.
- **Frases nuevas `ningún proyecto` y `está abierto`.** Los steps de sprints
  operan sobre el proyecto actual del escenario, así que CA-008-10 necesita una
  forma de que ese proyecto no exista; la misma frase sirve para CA-005-8 de
  US-005. `el sprint "X" está abierto` verifica el estado de CA-008-1 y es el
  espejo de `el sprint "X" está cerrado`, que ya está en el diccionario.
- **Las frases sin fechas generan fechas válidas.** Los escenarios de métricas
  no dependen de las fechas, y escribirlas en cada sprint los haría más largos
  sin probar nada nuevo. Las fechas generadas empiezan con el proyecto, para
  cumplir RN-008-7, y son consecutivas, para cumplir RN-008-13. Se eligieron 14
  días porque es una duración habitual de sprint y entran 26 sprints en el
  proyecto por defecto de los steps.
- **Los escenarios de CA-008-4 y CA-008-9 llegan con US-011.** Necesitan un
  sprint anterior cerrado, y el step `el sprint "X" cerrado` usa el cierre de
  US-011. Hasta entonces, esos criterios se cubren con tests unitarios y de
  integración, que pueden armar un sprint cerrado sin pasar por el cierre
  (sección 5). Se descartó un cierre mínimo en el adaptador de los steps:
  cerraría el sprint sin pasar por la aplicación, y el escenario no probaría
  nada real.
- **Fuera de alcance.** Asignar historias (US-009), marcarlas como completadas
  (US-010), cerrar el sprint (US-011) y consultar los sprints anteriores
  (US-012) son historias propias. Editar o cancelar un sprint no está en el
  backlog. La pantalla depende de T-004 (layout base y navegación).

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia, decisiones y criterios definidos por el equipo
- Revisado y entendido por: Juan Ignacio Vergara
