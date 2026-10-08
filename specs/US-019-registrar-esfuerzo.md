# US-019 — Registrar esfuerzo con integrante, fecha, actividad y horas

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #22 |
| **Escenarios BDD** | `features/US-019-registrar-esfuerzo.feature` |
| **Código** | `internal/domain/esfuerzo/`, caso de uso en `internal/app/`, repositorio en `internal/adapters/postgres/`, migración nueva en `migraciones/` |
| **Última actualización** | 2026-10-07 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero registrar las horas que dediqué a una
> historia, con la fecha y la actividad, para medir el esfuerzo real y
> compararlo contra lo estimado.

Cargar un registro de esfuerzo: quién trabajó (un integrante del proyecto), en
qué historia, qué día, en qué actividad y cuántas horas. Es la materia prima de
la consulta de esfuerzo (US-020), de la comparación estimado contra real
(US-021) y de las horas reales y la desviación (US-027).

## 2. Entradas

La carga pasa por dos capas, y cada una valida solo lo que puede saber:

- **Dominio** (`internal/domain/esfuerzo`): función pura que recibe los datos del
  registro y la fecha de hoy, valida todo y devuelve un `RegistroEsfuerzo`
  válido o todos los errores encontrados. No genera IDs ni sabe si el
  integrante o la historia existen.
- **Aplicación y persistencia** (`internal/app` y `internal/adapters/postgres`):
  el caso de uso sigue este orden:
  1. Verifica con una lectura simple que el proyecto exista. Si no existe,
     devuelve `ErrProyectoInexistente` sin validar nada más.
  2. Verifica que la historia exista **dentro de ese proyecto**
     (`ErrItemInexistente`) y que el integrante pertenezca a ese proyecto
     (`ErrIntegranteInexistente`).
  3. Valida los datos con el dominio, que devuelve todos los errores unidos.
  4. Registra el esfuerzo a través del repositorio, con un único `INSERT`. El ID
     lo genera la base.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. Viene del contexto de navegación (la URL); lo verifica el caso de uso, nunca el dominio |
| `itemID` | UUID | Sí | Una historia existente del proyecto. Viene del contexto de navegación o de la selección del formulario; la verifica el caso de uso |
| `integranteID` | UUID | Sí | Un integrante del proyecto. La verifica el caso de uso |
| `fecha` | `time.Time` | Sí | Un día calendario, no futuro. Se comparan solo año, mes y día (UTC); la hora se ignora |
| `actividad` | `string` | Sí | Después de recortar los extremos con `strings.TrimSpace`: entre 1 y 120 caracteres, contados en runas |
| `horas` | `float64` | Sí | Mayor que 0 y como máximo 24, en múltiplos de 0,25 |
| `hoy` | `time.Time` | Sí | La fecha actual; la pasa quien llama al dominio (el dominio no lee el reloj) |

No son entradas: el ID (lo genera la base) ni la fecha de carga (`creado_en`,
`DEFAULT now()`). El sprint no se guarda: se deduce de la fecha del registro
(RN-019-12).

El valor de `horas` llega de un formulario como texto (`2,5` o `2.5`).
Convertirlo a número es tarea del adaptador HTTP, no del dominio.

## 3. Salidas esperadas

Un `RegistroEsfuerzo` con estos atributos:

| Atributo | Valor al registrar |
|---|---|
| ID | UUID generado por la base (`DEFAULT gen_random_uuid()`) y devuelto con `RETURNING` |
| Proyecto, historia, integrante | Los recibidos |
| Fecha | El día recibido, sin hora |
| Actividad | La recibida, recortada |
| Horas | Las recibidas |
| Fecha de carga | La que asigna la base |

El dominio devuelve el registro todavía sin ID; el caso de uso devuelve el
registro ya guardado, con ID. Si hay errores de validación, el dominio no
devuelve registro: devuelve un único error que une todos los encontrados
(RN-019-8).

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-019-1** (dominio) — Las horas deben ser mayores que 0 y no superar 24.
  Un registro de `0`, un negativo o más de 24 se rechaza con
  `ErrHorasFueraDeRango`. Un valor que no es un número (`NaN` o infinito) se
  rechaza igual.
- **RN-019-2** (dominio) — Las horas se registran en múltiplos de 0,25 (15
  minutos). Cualquier otra fracción (por ejemplo `1,3`) se rechaza con
  `ErrHorasFraccion`. Los múltiplos de 0,25 se representan exactamente en
  `float64`, así que la comprobación no tiene errores de redondeo.
- **RN-019-3** (dominio) — La fecha no puede ser posterior a hoy
  (`ErrFechaFutura`). Se comparan solo año, mes y día. Hoy y cualquier día
  anterior se aceptan.
- **RN-019-4** (dominio) — La fecha es obligatoria: el valor cero de
  `time.Time` se rechaza con `ErrFechaVacia`.
- **RN-019-5** (dominio) — La actividad se recorta con `strings.TrimSpace`
  antes de validarla, y lo que se guarda es la actividad recortada. No puede
  quedar vacía (`ErrActividadVacia`) y tiene como máximo 120 caracteres
  contados en runas (`ErrActividadMuyLarga`).
- **RN-019-6** (aplicación) — El proyecto, la historia y el integrante tienen
  que existir, y la historia y el integrante tienen que ser **de ese
  proyecto**. El caso de uso los verifica en el orden de la sección 2, antes de
  validar los datos; el primero que falla se devuelve y no se unen con los
  errores de validación (sección 7).
- **RN-019-7** (aplicación) — Se puede registrar esfuerzo sobre una historia en
  cualquier estado (`pendiente`, `en_progreso` o `completado`) y esté o no
  asignada a un sprint. Registrar horas no cambia el estado de la historia.
- **RN-019-12** (consulta) — El sprint de un registro no se guarda: se deduce de
  su `fecha`. Es el sprint del proyecto cuyo rango (inicio y fin, ambos días
  incluidos) contiene esa fecha, o ninguno si cae fuera de todo sprint. US-008
  garantiza que los sprints de un proyecto no se superponen, así que hay a lo
  sumo uno. Esta historia no hace esa consulta: la hacen US-020, US-021 y
  US-027. Lo que fija acá es que el sprint depende de la fecha del trabajo y
  **no de la historia**: si una historia no completada vuelve al backlog y pasa
  al sprint siguiente (US-011), las horas ya cargadas siguen en el sprint donde
  se trabajaron.
- **RN-019-8** (dominio) — El dominio valida todos los datos y devuelve todos
  los errores juntos, unidos con `errors.Join`, en este orden fijo: primero el
  de la fecha, después el de la actividad y al final el de las horas. Cada
  error unido sigue siendo reconocible con `errors.Is`. De las horas se
  devuelve un solo error: el de rango (RN-019-1) tiene prioridad sobre el de
  fracción (RN-019-2).
- **RN-019-9** (persistencia) — El registro se guarda con un único `INSERT`.
  No hace falta bloquear filas: los registros son independientes entre sí, y
  dos cargas simultáneas del mismo integrante son dos registros distintos.
  Los datos de integridad (historia, integrante y proyecto) los protegen las
  claves foráneas.
- **RN-019-10** (persistencia) — El ID es un UUID que genera la base
  (`DEFAULT gen_random_uuid()`) y el repositorio lo devuelve con `RETURNING`.
  El dominio no genera IDs.
- **RN-019-11** (aplicación) — Si cualquier validación falla, no se registra
  nada.

## 5. Restricciones

- **Dominio puro.** `internal/domain/esfuerzo` no importa nada fuera de la
  biblioteca estándar, no hace SQL ni HTTP, no genera IDs y no llama a
  `time.Now()`: la fecha de hoy entra como parámetro. Sus errores se definen
  solo con la biblioteca estándar.
- **Paquete nuevo.** `esfuerzo` es un paquete propio y no parte de `backlog` ni
  de `metricas`: un registro de esfuerzo tiene su propio ciclo de vida. El
  dominio no importa `backlog` ni `proyectos`: recibe los IDs ya verificados.
- **Lo que el dominio no puede saber** (que el proyecto, la historia o el
  integrante existan y se correspondan) lo resuelven la capa de aplicación y el
  repositorio, en el orden de la sección 2.
- **Persistencia.** Una migración nueva en `migraciones/` (nunca se edita una
  mergeada) que crea `registros_esfuerzo` con: `id UUID PRIMARY KEY DEFAULT
  gen_random_uuid()`, `proyecto_id`, `item_id` e `integrante_id` con clave
  foránea, `fecha DATE NOT NULL`, `actividad TEXT NOT NULL`,
  `horas NUMERIC(4,2) NOT NULL CHECK (horas > 0 AND horas <= 24)`, `creado_en
  TIMESTAMPTZ NOT NULL DEFAULT now()` y un índice por `item_id` y otro por
  `integrante_id`. El número de la migración es el siguiente libre al abrir la
  rama; antes hay que confirmar que `goose` se configuró con
  `WithAllowOutofOrder` (riesgo conocido de las migraciones con numeración
  intercalada).
- **Dependencias.** La implementación parte de US-005 (la tabla de historias) y
  de US-001/US-002 (la tabla de integrantes ya existe en `00001_init.sql`). La
  spec se puede mergear antes.
- **Área nueva en los steps.** Los escenarios de esfuerzo necesitan un área
  propia, `Esfuerzo`, en `features/steps/mundo.go` (se agrega a `Servicios` y se
  conecta en `features/servicios_test.go`), con la política del ADR 0002: cada
  historia conecta solo lo que necesita.
- **Fuera de alcance.** La pantalla de carga (T-004), consultar el esfuerzo
  (US-020), comparar contra lo estimado (US-021), las horas estimadas y la
  desviación (US-027), editar o borrar un registro (no hay historia que lo
  pida) y el total diario por integrante (ver Decisiones).

## 6. Casos límite

- **CL-019-1** — `24` horas: se acepta. `24,25`: se rechaza con
  `ErrHorasFueraDeRango`.
- **CL-019-2** — `0,25` horas (el mínimo): se acepta. `0`: se rechaza con
  `ErrHorasFueraDeRango`.
- **CL-019-3** — `-1` hora: se rechaza con `ErrHorasFueraDeRango`.
- **CL-019-4** — `1,3` horas: se rechaza con `ErrHorasFraccion`. `1,5` y `2`
  se aceptan.
- **CL-019-5** — `30,3` horas: se rechaza solo con `ErrHorasFueraDeRango`,
  porque el error de rango tiene prioridad sobre el de fracción (RN-019-8).
- **CL-019-6** — Fecha de hoy: se acepta. Fecha de mañana: se rechaza con
  `ErrFechaFutura`.
- **CL-019-7** — Fecha de hoy con una hora posterior a la actual (por ejemplo
  23:59): se acepta, porque solo se comparan año, mes y día.
- **CL-019-8** — Fecha sin completar (valor cero): se rechaza con
  `ErrFechaVacia`.
- **CL-019-9** — Actividad de exactamente 120 caracteres: se acepta. Con 121,
  se rechaza con `ErrActividadMuyLarga`.
- **CL-019-10** — Actividad con espacios al principio y al final: se recortan y
  se guarda sin ellos. Actividad solo con espacios: `ErrActividadVacia`, no
  `ErrActividadMuyLarga`.
- **CL-019-11** — Dos registros del mismo integrante, la misma historia y el
  mismo día: se aceptan los dos. No se unifican ni se rechazan por duplicados.
- **CL-019-12** — Dos registros del mismo integrante el mismo día que suman más
  de 24 horas: se aceptan (ver Decisiones).
- **CL-019-13** — Tres errores a la vez: fecha de mañana, actividad vacía y
  `0` horas. El dominio devuelve los tres unidos, en este orden:
  `ErrFechaFutura`, `ErrActividadVacia`, `ErrHorasFueraDeRango`.
- **CL-019-14** — Registrar sobre una historia completada o que no está en un
  sprint: se acepta (RN-019-7).
- **CL-019-16** — Una historia con horas registradas con fecha dentro del
  Sprint 1 no se completa, vuelve al backlog y pasa al Sprint 2: esas horas
  siguen en el Sprint 1. Las que se registren con fecha dentro del Sprint 2
  cuentan en el Sprint 2.
- **CL-019-17** — Registrar con una fecha que cae entre dos sprints o antes del
  primero: se acepta y el registro no pertenece a ningún sprint.
- **CL-019-15** — Un integrante de otro proyecto, o una historia de otro
  proyecto: se rechaza con `ErrIntegranteInexistente` o `ErrItemInexistente`
  respectivamente, aunque el integrante o la historia existan en otro lado.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) |
|---|---|---|
| Fecha sin completar | `esfuerzo.ErrFechaVacia` | "la fecha es obligatoria" |
| Fecha posterior a hoy | `esfuerzo.ErrFechaFutura`, envuelto indicando la fecha recibida | "la fecha no puede ser posterior a hoy" |
| Actividad vacía o con solo espacios | `esfuerzo.ErrActividadVacia` | "la actividad es obligatoria" |
| Actividad de más de 120 caracteres después de recortarla | `esfuerzo.ErrActividadMuyLarga`, envuelto indicando cuántos caracteres tiene | "la actividad no puede tener mas de 120 caracteres" |
| Horas menores o iguales a 0, mayores a 24, `NaN` o infinito | `esfuerzo.ErrHorasFueraDeRango`, envuelto indicando el valor recibido | "las horas tienen que ser mayores a 0 y no pasar de 24" |
| Horas que no son múltiplo de 0,25 | `esfuerzo.ErrHorasFraccion`, envuelto indicando el valor recibido | "las horas se registran en multiplos de 0,25" |
| El proyecto no existe | `app.ErrProyectoInexistente`, envuelto indicando el ID recibido | "el proyecto no existe" |
| La historia no existe o no es de ese proyecto | `app.ErrItemInexistente`, envuelto indicando el ID recibido | "el item no existe" |
| El integrante no existe o no es de ese proyecto | `app.ErrIntegranteInexistente`, envuelto indicando el ID recibido | "el integrante no existe" |

Los seis primeros son del dominio (`internal/domain/esfuerzo`): se detectan todos
en una sola validación, antes de escribir en la base, y se devuelven unidos con
`errors.Join` en el orden de RN-019-8. Con varios errores, el mensaje es la
unión de los mensajes, uno por línea. Los tres últimos los detecta la capa de
aplicación **antes** de validar los datos, en el orden de la tabla, y se
devuelven de a uno: no se unen con los de validación. `ErrProyectoInexistente`
ya existe en `internal/app`; `ErrItemInexistente` lo comparte con US-013 (lo
define quien llegue primero a `main`); `ErrIntegranteInexistente` es nuevo. Todos
son valores de paquete comparables con `errors.Is`, y en ningún caso se registra
nada.

## 8. Criterios de aceptación

- **CA-019-1** (normal) — Con un proyecto, una historia y un integrante
  existentes, registrar 2,5 horas de "Desarrollo" de ayer deja un registro
  guardado con esos datos.
- **CA-019-2** (alternativo) — Se pueden registrar varias veces horas del mismo
  integrante sobre la misma historia el mismo día.
- **CA-019-3** (alternativo) — Se puede registrar esfuerzo sobre una historia
  completada.
- **CA-019-4** (límite) — 24 horas se aceptan y 24,25 se rechazan; 0,25 se
  acepta y 0 se rechaza.
- **CA-019-5** (límite) — La fecha de hoy se acepta y la de mañana se rechaza.
- **CA-019-6** (error) — Horas negativas o en cero → rechazo, no se registra
  nada.
- **CA-019-7** (error) — Horas que no son múltiplo de 0,25 → rechazo.
- **CA-019-8** (error) — Actividad vacía o solo espacios → rechazo.
- **CA-019-11** (alternativo) — Las horas de una historia que cambia de sprint
  siguen contando en el sprint de la fecha en que se trabajaron.
- **CA-019-9** (error) — Integrante o historia que no son del proyecto →
  rechazo.
- **CA-019-10** (error) — Proyecto inexistente → rechazo, no se registra nada.

---

## Decisiones tomadas y descartadas

- **Horas en múltiplos de 0,25, máximo 24 por registro.** El máximo viene de
  el ejemplo de commit de `CONTRIBUTING.md` (DEF-007: "rechazar horas
  negativas y mayores a 24"). Los cuartos de hora evitan cargas como `1,333` y
  se representan exactos en `float64`; se descartó guardar minutos enteros
  porque obligaría al formulario a convertir. En la base se guarda
  `NUMERIC(4,2)`, exacto, con un `CHECK` como red de seguridad.
- **El tope de 24 es por registro, no por día.** Validar que las horas de un
  integrante en un día no pasen de 24 exigiría sumar los registros existentes
  dentro de una transacción con bloqueo, para un caso que en la práctica no
  ocurre (nadie carga más de 24 horas ni duplica el mismo día a propósito). Se
  deja para una historia aparte si alguien lo pide.
- **Fecha: no futura, sin límite hacia atrás.** Se puede cargar esfuerzo
  atrasado (es lo habitual: se completa al final del día o de la semana). Se
  descartó validar la fecha contra el rango del proyecto porque ese rango se
  puede editar (US-002) y dejaría registros ya cargados fuera de rango.
- **Actividad como texto libre.** Se descartó una lista cerrada (desarrollo,
  testing, reunión, etc.): nadie definió las categorías y cada una sería una
  regla más. Se limita a 120 caracteres, como el título de un ítem, para que
  entre en una fila de tabla.
- **Esfuerzo atado a una historia.** La historia dice "registrar esfuerzo"
  pero US-020 consulta "por historia, integrante y sprint" y US-021 lo compara
  contra la estimación de la historia: sin historia no hay con qué comparar.
- **El sprint se deduce de la fecha del registro, no de la historia ni se
  guarda.** Primera versión: deducirlo de la historia. Problema que marcó
  Vergara: cuando una historia no completada vuelve al backlog y pasa al sprint
  siguiente (US-011), sus horas ya cargadas se mudarían y cambiarían las
  métricas de un sprint cerrado. Se evaluó guardar `sprint_id` al registrar y
  se descartó: duplica un dato que ya se puede derivar, depende de US-009 (qué
  sprint tiene la historia) y de la tabla `sprints` para la clave foránea, y
  las horas cargadas antes de que la historia entre a un sprint quedarían sin
  sprint para siempre. Por fecha, como los sprints de un proyecto no se
  superponen (US-008), cada fecha cae en un solo sprint o en ninguno. Límite
  conocido: si las fechas de un sprint cambiaran, cambiaría el sprint de sus
  registros; hoy no se pueden editar (US-008 deja fuera editar o cancelar).
- **0 horas no se acepta** (RN-019-1): un registro sin horas no aporta nada.
- **Historia e integrante tienen que ser del proyecto.** Las claves foráneas
  solas no lo garantizan (podría ser un integrante de otro proyecto). Lo
  verifica el caso de uso con una lectura por proyecto.
- **Paquete `esfuerzo`.** Los nombres de paquetes y archivos van en español
  (ADR 0003); `esfuerzo` es el término del dominio.
- **Duplicados y registros múltiples.** Se aceptan: dos sesiones de trabajo del
  mismo día son dos registros, no uno repetido.
- **La pantalla queda fuera de alcance.** Depende de T-004 (layout base y
  navegación). Esta historia entrega dominio, caso de uso, repositorio y
  migración.
- **Frases nuevas del diccionario** (a aprobar en el PR de implementación):
  `el integrante "X" registra N horas de "A" en la historia "H" el AAAA-MM-DD`,
  `el proyecto "P" tiene N registros de esfuerzo` y `la historia "H" tiene N
  horas registradas`. Las tres pertenecen al área nueva `Esfuerzo`.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia, usando la spec de US-005 como modelo; las decisiones sobre
  horas, fecha y actividad se proponen acá y se confirman en la revisión
- Revisado y entendido por: Tomás Romero
