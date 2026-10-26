# US-023 — Cambiar el estado de un defecto y registrar su sprint de resolución

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Angelo Conforti |
| **Issue** | #31 |
| **Depende de** | US-022 (tabla `defectos`, paquete `defecto`, estado inicial y errores base): se implementa después de que esa historia esté en `main`; esta spec se puede mergear antes, como la de US-019 |
| **Escenarios BDD** | `features/US-023-cambiar-estado-defecto.feature` |
| **Código** | `internal/domain/defecto/` (los estados nuevos y `CambiarEstado`), caso de uso en `internal/app/`, repositorios en `internal/adapters/postgres/` y `internal/adapters/memoria/`, migración nueva `migraciones/00013_defectos_estado.sql` |
| **Última actualización** | 2026-10-25 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero cambiar el estado de un defecto y, cuando
> se resuelve, registrar el sprint en el que se resolvió, para seguir su ciclo
> de vida y medir cuántos defectos resolvemos en cada sprint.

Mover un defecto de un proyecto entre sus estados (`abierto`, `en_correccion`,
`resuelto` y `descartado`) respetando las transiciones permitidas, y guardar el
sprint de resolución cuando el defecto pasa a `resuelto`. Es el segundo de los
seis campos del requerimiento 6 de la guía que US-022 dejó vacíos: el estado y
el sprint de resolución, tal como los nombra la plantilla
`.github/ISSUE_TEMPLATE/02-defect.yml`.

Es la materia prima de los defectos resueltos por sprint (US-029), de la lista
de defectos por sprint y severidad (US-024) y, a través de ellas, del Dashboard
(US-030) y del gráfico de evolución (US-032).

## 2. Entradas

El cambio pasa por dos capas, y cada una valida solo lo que puede saber:

- **Dominio** (`internal/domain/defecto`): método puro que recibe el defecto, el
  estado pedido, el sprint de resolución y el número del sprint de detección;
  valida todo y devuelve el defecto con el estado nuevo o todos los errores
  encontrados. No lee la base, no genera IDs, no lee el reloj y no sabe si los
  sprints existen: los números ya vienen traducidos.
- **Aplicación y persistencia** (`internal/app` y `internal/adapters/postgres`):
  el caso de uso sigue este orden:
  1. Verifica con una lectura simple que el proyecto exista. Si no existe,
     devuelve `ErrProyectoInexistente` sin verificar nada más.
  2. Verifica que el defecto exista **dentro de ese proyecto**
     (`ErrDefectoInexistente`) y lo lee: de ahí sale su estado actual.
  3. Si se informó un sprint de resolución, verifica que exista **dentro de ese
     proyecto** (`ErrSprintInexistente`). Un ID vacío es "sin sprint de
     resolución": no se verifica nada.
  4. Traduce los sprints a su número (`sprint.Numero`): el de detección del
     defecto y, si se informó, el de resolución.
  5. Valida con el dominio, que devuelve todos los errores unidos.
  6. Guarda el cambio a través del repositorio, con un único `UPDATE`
     condicionado al estado que leyó.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. Viene del contexto de navegación (la URL); lo verifica el caso de uso, nunca el dominio |
| `defectoID` | UUID | Sí | Un defecto existente de ese proyecto, en cualquier estado. Viene del contexto de navegación (la ficha o la fila del defecto); lo verifica el caso de uso |
| `estado` | `string` | Sí | El estado al que se quiere pasar: exactamente uno de `abierto`, `en_correccion`, `resuelto` o `descartado`, en minúscula, sin tilde y sin espacios alrededor. Viene de la selección del formulario; lo valida el dominio |
| `sprintResolucionID` | UUID | Solo con `resuelto` | El sprint donde se resolvió: un sprint existente del proyecto, abierto o cerrado, cuyo número sea mayor o igual que el del sprint de detección. Vacío si el estado no es `resuelto`. Viene de la selección del formulario; el caso de uso verifica que exista y el dominio, el resto |

Un ID de defecto vacío o mal formado es lo mismo que uno inexistente: el
repositorio responde que no existe, no que falló. En cambio, un
`sprintResolucionID` vacío no es un sprint inexistente: significa "sin sprint de
resolución" (la opción "sin resolver" de la plantilla de defectos). Si el estado
pedido es `resuelto`, eso lo rechaza el dominio (`ErrSprintResolucionFaltante`);
si no, es lo correcto. Un ID no vacío que no existe o está mal formado sí es un
sprint inexistente.

No son entradas:

- El **estado actual** del defecto: lo lee el caso de uso de lo guardado, nunca
  llega del formulario. Así, una pantalla desactualizada se evalúa contra el
  estado real (CL-023-18).
- Los **números de sprint**: los traduce el caso de uso a partir de los IDs.
- La fecha del cambio y quién lo hizo: no se guardan (ver Decisiones).

La plantilla de defectos muestra los estados y el sprint como etiquetas para la
persona. Convertir el texto del formulario en los valores de la tabla de arriba
es tarea del adaptador HTTP, no del dominio.

## 3. Salidas esperadas

El mismo `Defecto` de US-022, con dos atributos cambiados:

| Atributo | Valor después del cambio |
|---|---|
| Estado | El pedido |
| Sprint de resolución | El recibido (su ID, no su número) si el estado es `resuelto`; vacío (`NULL` en la base) en cualquier otro estado |
| Resto (ID, proyecto, historia, sprint de detección, descripción, severidad, fecha de carga) | Sin cambios |

```
type SprintDeResolucion struct {
    ID     string // el ID del sprint; vacío si no se informó
    Numero int    // sprint.Numero, traducido por la capa de aplicación
}

func (d Defecto) CambiarEstado(nuevo Estado, resolucion SprintDeResolucion, numeroDeteccion int) (Defecto, error)
```

El dominio devuelve una copia del defecto con el estado nuevo; el caso de uso
devuelve el defecto ya guardado. Si hay errores de validación, el dominio no
devuelve defecto: devuelve un único error que une todos los encontrados
(RN-023-6) y no se modifica nada.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-023-1** (dominio) — Los estados de un defecto son exactamente cuatro:
  `abierto` (el inicial, RN-022-3), `en_correccion`, `resuelto` y `descartado`.
  Cualquier otro valor, incluido el vacío y los que difieren en mayúsculas,
  tildes o espacios, se rechaza con `ErrEstadoInvalido`, envuelto indicando el
  valor recibido. El dominio no pasa a minúscula ni recorta.
- **RN-023-2** (dominio) — Solo se permiten estas transiciones; cualquier otra,
  incluida la de un estado a sí mismo, se rechaza con `ErrTransicionInvalida`,
  envuelto indicando el estado actual y el pedido (`de resuelto a
  en_correccion`):

  | Actual \ Pedido | `abierto` | `en_correccion` | `resuelto` | `descartado` |
  |---|---|---|---|---|
  | `abierto` | no (mismo estado) | sí | sí | sí |
  | `en_correccion` | sí | no (mismo estado) | sí | sí |
  | `resuelto` | sí (reapertura) | no | no (mismo estado) | no |
  | `descartado` | sí (reapertura) | no | no | no (mismo estado) |

  Un defecto `resuelto` o `descartado` solo se reabre: vuelve a `abierto` y, si
  corresponde, sigue desde ahí. Pedir el estado que ya tiene no es un no-op: es
  un error (ver Decisiones).
- **RN-023-3** (dominio) — El único estado que cuenta como **resuelto** es
  `resuelto`, y es el único que lleva sprint de resolución:

  | Estado | Sprint de resolución | ¿Cuenta como resuelto? |
  |---|---|---|
  | `abierto` | Prohibido | No |
  | `en_correccion` | Prohibido | No |
  | `resuelto` | Obligatorio | Sí, en su sprint de resolución |
  | `descartado` | Prohibido | No (ver RN-023-13) |

  Pasar a `resuelto` sin sprint de resolución se rechaza con
  `ErrSprintResolucionFaltante`. Pasar a cualquier otro estado informando un
  sprint de resolución se rechaza con `ErrSprintResolucionNoCorresponde`: no se
  descarta en silencio.
- **RN-023-4** (dominio) — Con el estado `resuelto`, el número del sprint de
  resolución tiene que ser mayor o igual que el del sprint de detección del
  defecto. Uno anterior se rechaza con `ErrSprintResolucionAnterior`, envuelto
  indicando los dos números. Igual se acepta: un defecto se puede detectar y
  resolver en el mismo sprint. La comparación es por **número** (`sprint.Numero`),
  que es lo que recibe US-029, y no por ID. Esta regla solo se evalúa con el
  estado `resuelto` y un sprint informado; con otro estado, el sprint informado
  ya es el error de RN-023-3.
- **RN-023-5** (dominio) — Salvo en `resuelto`, el defecto queda sin sprint de
  resolución. Por eso **reabrir** un defecto `resuelto` (pasarlo a `abierto`) deja
  su sprint de resolución vacío: el sprint anterior no se conserva. Si después
  se vuelve a resolver, queda el sprint que se informe en ese momento.
- **RN-023-6** (dominio) — El dominio valida todo y devuelve todos los errores
  juntos, unidos con `errors.Join`, en este orden fijo: primero el del estado
  (`ErrEstadoInvalido`); con el estado válido, después el de la transición
  (`ErrTransicionInvalida`), después el del sprint de resolución
  (`ErrSprintResolucionFaltante` o `ErrSprintResolucionNoCorresponde`, que no
  pueden darse a la vez) y al final el de la comparación
  (`ErrSprintResolucionAnterior`). Con un estado inválido se devuelve solo
  `ErrEstadoInvalido`: sin saber a qué estado se quiere pasar, las demás reglas
  no se pueden evaluar. Cada error unido sigue siendo reconocible con
  `errors.Is`.
- **RN-023-7** (aplicación) — El proyecto, el defecto y el sprint de resolución
  (si se informó) tienen que existir, y el defecto y el sprint tienen que ser
  **de ese proyecto**. El caso de uso los verifica en ese orden
  (`ErrProyectoInexistente`, `ErrDefectoInexistente`, `ErrSprintInexistente`),
  antes de validar los datos; el primero que falla se devuelve y no se une con
  los errores de validación (sección 7).
- **RN-023-8** (aplicación) — El estado actual lo lee el caso de uso del defecto
  guardado, y los dos números de sprint los traduce a partir de los IDs
  (`sprint.Numero`) antes de llamar al dominio. El dominio nunca recibe un ID
  para compararlo y nunca consulta la base.
- **RN-023-9** (aplicación) — Se puede cambiar el estado de un defecto sea cual
  sea el estado de su historia relacionada (`pendiente`, `en_progreso` o
  `completado`, asignada o no a un sprint), y el sprint de resolución puede ser
  cualquier sprint del proyecto: abierto o cerrado, anterior o posterior al de la
  historia, haya empezado o no. El cambio solo modifica el defecto: ni la
  historia ni los sprints cambian, ni tampoco la descripción, la severidad, la
  historia relacionada o el sprint de detección del defecto.
- **RN-023-10** (aplicación) — La operación es de todo o nada: el estado y el
  sprint de resolución se guardan juntos o no se guarda ninguno. Si cualquier
  verificación o validación falla, no se modifica nada.
- **RN-023-11** (persistencia) — El cambio se guarda con un único `UPDATE`:
  `UPDATE defectos SET estado = $3, sprint_resolucion_id = $4 WHERE id = $1 AND
  proyecto_id = $2 AND estado = $5`, donde `$5` es el estado que leyó el caso de
  uso y `$4` es `NULL` salvo con `resuelto`. Si no actualiza ninguna fila, el
  repositorio relee el defecto: si no existe en el proyecto, devuelve
  `ErrDefectoInexistente`; si existe, es que otro cambio se le adelantó entre la
  lectura y el `UPDATE`, y devuelve `ErrDefectoModificado` sin guardar nada. No
  hace falta una transacción ni bloquear filas. Si el sprint de resolución
  desaparece entre la verificación y el `UPDATE`, la clave foránea lo rechaza y
  el repositorio devuelve `ErrSprintInexistente`.
- **RN-023-12** (persistencia) — La base repite dos reglas como última defensa,
  con `CHECK` en la migración nueva: el estado es uno de los cuatro de RN-023-1
  y el defecto tiene sprint de resolución si y solo si está `resuelto`. La
  comparación de números de RN-023-4 y la pertenencia al proyecto no se pueden
  expresar en un `CHECK` de una fila: las hacen cumplir el dominio y el caso de
  uso.
- **RN-023-13** (consulta) — Lo que esta historia guarda alcanza para armar la
  lista que recibe US-029 sin agregarle reglas. La capa de aplicación toma los
  defectos del proyecto que **no** están `descartado` y los traduce a
  `DefectoRegistrado`: `SprintDeteccion` es el número del sprint de
  `sprint_deteccion_id`; `SprintResolucion` es el número del sprint de
  `sprint_resolucion_id`, o `0` si es `NULL`. Por RN-023-3, RN-023-4 y
  RN-023-12, un defecto con `SprintResolucion > 0` está `resuelto` y nunca se
  resolvió antes de detectarse, así que US-029 no puede recibir un dato
  corrupto. Esta historia no escribe esa consulta: la escribe quien integre
  US-029 (el Dashboard, US-030). Lo que fija acá es lo que se guarda y cómo se
  lee.

## 5. Restricciones

- **Dominio puro.** `internal/domain/defecto` no importa nada fuera de la
  biblioteca estándar, no hace SQL ni HTTP, no genera IDs y no llama a
  `time.Now()`: no necesita el reloj. Sus errores se definen solo con la
  biblioteca estándar. Esta historia agrega al paquete que crea US-022 los
  valores `EstadoEnCorreccion`, `EstadoResuelto` y `EstadoDescartado` (US-022 ya
  define `Estado` y `EstadoAbierto`), el tipo `SprintDeResolucion` y el método
  `CambiarEstado`. No importa `sprint` ni `metricas`: recibe los números ya
  traducidos y no comparte tipos con US-029.
- **Lo que el dominio no puede saber** (que el proyecto, el defecto y el sprint
  existan y se correspondan, y el número de cada sprint) lo resuelven la capa de
  aplicación y el repositorio, en el orden de la sección 2. Las claves foráneas
  impiden un sprint inexistente, pero no que sea de otro proyecto: eso lo verifica
  el caso de uso con una lectura por proyecto, como en US-022. Como ninguna
  historia mueve sprints ni defectos de un proyecto a otro, ni cambia el sprint de
  detección ni el número de un sprint, esa verificación no se desactualiza entre
  la lectura y el `UPDATE`.
- **Persistencia.** La migración `00013_defectos_estado.sql` (nunca se edita una
  mergeada) agrega a `defectos` dos restricciones y no crea columnas, porque
  `estado` y `sprint_resolucion_id` ya las crea la `00012`: `CONSTRAINT
  estado_de_defecto_valido CHECK (estado IN ('abierto', 'en_correccion',
  'resuelto', 'descartado'))` y `CONSTRAINT resolucion_segun_estado CHECK
  ((estado = 'resuelto') = (sprint_resolucion_id IS NOT NULL))`. Su `Down` las
  borra. Cuando se aplica no hay defectos en estado distinto de `abierto` ni con
  sprint de resolución, porque nada puede escribirlos antes de esta historia.
- **Número de la migración y `goose`.** El plan de trabajo reserva "`00011` en
  adelante" para defectos; el PR #170 (abierto) le da la `00011` a US-018 y
  deja "`00012` en adelante" para US-022 y US-023, y la spec de US-022 (PR #179,
  abierto) toma la `00012`. A la fecha de esta spec, `main` llega a la `00010` y
  ninguna rama ni PR abierto usa la `00013`. Antes de abrir la rama de
  implementación hay que confirmar que `00013` sigue libre y que la `00012` de
  US-022 ya está en `main`: esta migración la necesita, porque agrega
  restricciones a la tabla que esa crea. `Migrar` se configura con
  `goose.WithAllowOutofOrder(true)` y `make migrate` corre `goose
  -allow-missing`, así que una migración que llegue fuera de orden se aplica
  igual (T-012), pero no antes de la tabla que modifica.
- **Dependencias.** La implementación parte de US-022 (tabla `defectos`,
  paquete `defecto`, `Estado`, `EstadoAbierto`, el área `Defectos` de los steps y
  los errores de existencia que comparte) y de US-008 (`sprints`, con su número).
  No depende de US-024 ni de US-029: son ellas las que se apoyan en esta. La
  spec se puede mergear antes.
- **Catálogo de errores (T-005).** Los siete errores nuevos se registran en
  `internal/adapters/http/catalogo.go` en el PR de implementación (el test del
  catálogo falla si no están), sin exclusiones nuevas. Los cinco del dominio
  van en el bloque `// defecto` (el que crea US-022; el segundo PR que llegue a
  `main` agrega sus líneas al bloque que ya exista): `defecto.ErrEstadoInvalido`
  como `validacion` con el campo `estado`; `defecto.ErrTransicionInvalida` como
  `conflicto` con el campo `estado`, porque depende del estado guardado;
  `defecto.ErrSprintResolucionFaltante` y `defecto.ErrSprintResolucionNoCorresponde`
  como `validacion` con el campo `sprint_resolucion`; y
  `defecto.ErrSprintResolucionAnterior` como `conflicto` con el campo
  `sprint_resolucion`, porque compara contra el sprint de detección guardado
  (RN-T005-2). Los dos de `app` van en el bloque `// app`, de tipo
  `no_encontrado` el primero y `conflicto` el segundo, sin campo:
  `app.ErrDefectoInexistente` y `app.ErrDefectoModificado`.
  `app.ErrProyectoInexistente` y `app.ErrSprintInexistente` ya están
  registrados.
- **Área `Defectos` en los steps.** La define US-022 (en
  `features/steps/mundo.go`, con sus steps en `features/steps/defectos.go` y la
  conexión en `features/servicios_test.go`, con la política del ADR 0002: cada
  historia conecta solo lo que necesita). Esta historia no crea un área: suma a
  la interfaz `Defectos` dos métodos, `CambiarEstadoDeDefecto(historia, estado,
  sprintResolucion string) error` y `SprintDeResolucionDeDefecto(historia string)
  (sprint string, resuelto bool, err error)`, y estas frases, que se aprueban en
  el PR de implementación y se suman a `docs/diccionario-steps.md`. Para armar el
  contexto reutiliza las que define US-022 (un proyecto, una historia, un sprint
  y `un defecto de severidad ... detectado en el sprint ...`) y para verificar
  el resultado, `el defecto de la historia "X" tiene estado "Y"`, `está sin
  resolver` y `fue detectado en el sprint`:

  | Frase | Expresión |
  |---|---|
  | `Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "en_correccion"` | `^se cambia el estado del defecto de la historia "([^"]*)" a "([^"]*)"$` |
  | `Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 2"` | `^se cambia el estado del defecto de la historia "([^"]*)" a "([^"]*)" con el sprint de resolución "([^"]*)"$` |
  | `Y el defecto de la historia "Crear proyecto" en estado "en_correccion"` (contexto) | `^el defecto de la historia "([^"]*)" en estado "([^"]*)"$` |
  | `Y el defecto de la historia "Crear proyecto" resuelto en el sprint "Sprint 2"` (contexto) | `^el defecto de la historia "([^"]*)" resuelto en el sprint "([^"]*)"$` |
  | `Y el defecto de la historia "Crear proyecto" fue resuelto en el sprint "Sprint 2"` | `^el defecto de la historia "([^"]*)" fue resuelto en el sprint "([^"]*)"$` |

  Las dos frases de contexto hacen falta porque el escenario tiene un solo
  `Cuando`: llevan al defecto, desde `abierto`, al estado que el escenario
  necesita, pasando por el caso de uso real. Si ese cambio se rechaza, el
  escenario falla (es una precondición mal armada, no el rechazo que se prueba).
  Con `descartado` y `abierto` alcanza la primera. La historia y el sprint se
  buscan por nombre dentro del proyecto actual, y las frases operan sobre el
  último defecto registrado de esa historia, igual que las de US-022: una
  historia sin defecto o un sprint que no existe se mandan con un ID que no
  existe, así el rechazo lo produce la aplicación y no los steps. Eso cubre el
  defecto y el sprint "inexistentes" (CL-023-16 y CL-023-17); "de otro
  proyecto", el cambio simultáneo y los `CHECK` los prueban el caso de uso y el
  repositorio.
- **Compatibilidad con US-029.** US-029 recibe una lista de
  `DefectoRegistrado{SprintDeteccion, SprintResolucion}` con los sprints por su
  número, y rechaza `ErrResolucionAnteriorADeteccion`. Este modelo evita que
  llegue ese dato: RN-023-4 lo impide al escribir, RN-023-12 asegura que
  "tiene sprint de resolución" y "está `resuelto`" sean lo mismo, y RN-023-13
  dice cómo se arma la lista. No cambia la spec de US-029. Ver también
  Decisiones.
- **Fuera de alcance.** Registrar defectos (US-022), listarlos (US-024), la
  pantalla del cambio de estado (depende de T-004), editar la descripción, la
  severidad, la historia o el sprint de detección de un defecto, borrar
  defectos, el historial de cambios de estado (ver Decisiones) y las métricas
  (US-029, US-030 y US-032).

## 6. Casos límite

- **CL-023-1** (RN-023-2) — Transiciones permitidas: `abierto` a `en_correccion`;
  `abierto` a `descartado`; `abierto` a `resuelto` con el Sprint 2 (sin pasar por
  `en_correccion`); `en_correccion` a `abierto`; `en_correccion` a `resuelto` con
  el Sprint 2; `en_correccion` a `descartado`; `resuelto` a `abierto`;
  `descartado` a `abierto`. En los ocho casos el defecto queda con el estado
  pedido.
- **CL-023-2** (RN-023-2) — Transiciones prohibidas: `resuelto` a `en_correccion`,
  `resuelto` a `descartado`, `descartado` a `en_correccion` y `descartado` a
  `resuelto`. Se rechazan con `ErrTransicionInvalida`, envuelto con los dos
  estados (`de resuelto a en_correccion`), y el defecto no cambia.
- **CL-023-3** (RN-023-2) — Pedir el estado que ya tiene: `abierto` a `abierto`,
  `en_correccion` a `en_correccion`, `descartado` a `descartado` y `resuelto` a
  `resuelto` (incluso con un sprint de resolución distinto del guardado) se
  rechazan con `ErrTransicionInvalida`. No son un no-op.
- **CL-023-4** (RN-023-1) — Estado `cerrado`, `Resuelto`, `RESUELTO`,
  `" resuelto "`, `en correccion`, `en_corrección` o vacío: se rechaza con
  `ErrEstadoInvalido`, envuelto indicando el valor recibido. Solo `abierto`,
  `en_correccion`, `resuelto` y `descartado` son estados.
- **CL-023-5** (RN-023-3) — Resolver sin sprint de resolución (ID vacío): se
  rechaza con `ErrSprintResolucionFaltante`, aunque la transición sea válida.
- **CL-023-6** (RN-023-3) — Un sprint de resolución con un estado que no es
  `resuelto` (`en_correccion` con el Sprint 2, `descartado` con el Sprint 2,
  `abierto` con el Sprint 2): se rechaza con `ErrSprintResolucionNoCorresponde`.
  No se ignora el sprint ni se guarda.
- **CL-023-7** (RN-023-4) — Defecto detectado en el Sprint 3 y resuelto en el
  Sprint 2: se rechaza con `ErrSprintResolucionAnterior`, envuelto con los
  números 2 y 3, y el defecto no cambia.
- **CL-023-8** (RN-023-4) — Defecto detectado en el Sprint 3: resolverlo en el
  Sprint 3 (el mismo) o en el Sprint 4 se acepta. Detectado en el Sprint 1 y
  resuelto en el Sprint 1: se acepta (es el borde inferior).
- **CL-023-9** (RN-023-4, RN-023-9) — El sprint de resolución es uno cerrado, uno
  abierto o uno que todavía no empezó (con fechas posteriores a hoy): se acepta
  en los tres casos (ver Huecos).
- **CL-023-10** (RN-023-5) — Defecto resuelto en el Sprint 2 y reabierto: queda
  `abierto` y sin sprint de resolución (`NULL`). Si después se resuelve en el
  Sprint 3, queda con el Sprint 3, no con el 2. Un defecto `descartado` y
  reabierto queda `abierto`, también sin sprint.
- **CL-023-11** (RN-023-13) — El defecto reabierto de CL-023-10 deja de contar
  como resuelto en el Sprint 2: al armar la lista de US-029 su `SprintResolucion`
  es `0`. Un defecto `descartado` no aparece en la lista, ni como detectado ni
  como resuelto.
- **CL-023-12** (RN-023-6) — Dos errores a la vez: defecto `descartado` que se
  quiere pasar a `resuelto` sin sprint. El dominio devuelve los dos unidos, en
  este orden: `ErrTransicionInvalida`, `ErrSprintResolucionFaltante`.
  `errors.Is` reconoce cada uno.
- **CL-023-13** (RN-023-6) — Dos errores a la vez: defecto ya `resuelto`,
  detectado en el Sprint 3, que se vuelve a resolver en el Sprint 2. Se devuelven
  `ErrTransicionInvalida` y `ErrSprintResolucionAnterior`, en ese orden. Nunca
  hay más de dos a la vez: la comparación solo corre con `resuelto` y con un
  sprint informado, así que no se da junto con `ErrSprintResolucionFaltante` (sin
  sprint no hay qué comparar) ni con `ErrSprintResolucionNoCorresponde`.
- **CL-023-14** (RN-023-6) — Estado `urgente`, con o sin sprint de resolución, y
  aunque el defecto esté `resuelto`: se devuelve solo `ErrEstadoInvalido`.
- **CL-023-15** (RN-023-7, RN-023-10) — Proyecto inexistente, con el defecto y el
  sprint también inexistentes y los datos inválidos: se devuelve solo
  `ErrProyectoInexistente` y no se modifica nada.
- **CL-023-16** (RN-023-7) — Proyecto existente, defecto inexistente (o con ID
  vacío o mal formado) o de otro proyecto, sprint inexistente y datos inválidos:
  se devuelve solo `ErrDefectoInexistente`, aunque el defecto exista en otro
  proyecto.
- **CL-023-17** (RN-023-7) — Proyecto y defecto correctos, sprint de resolución
  inexistente, con ID mal formado o de otro proyecto, y datos inválidos (por
  ejemplo, con el estado `en_correccion`): se devuelve solo `ErrSprintInexistente`,
  aunque el sprint exista en otro proyecto. En cambio, un sprint de resolución
  **vacío** con `resuelto` no es un sprint inexistente: es
  `ErrSprintResolucionFaltante` (CL-023-5).
- **CL-023-18** (RN-023-8) — La pantalla muestra el defecto `abierto`, pero otra
  persona ya lo resolvió: pedir `descartado` se evalúa contra el estado real,
  `resuelto`, y se rechaza con `ErrTransicionInvalida` (`de resuelto a
  descartado`). Pedir `abierto` es una reapertura válida.
- **CL-023-19** (RN-023-9) — Cambiar el estado de un defecto cuya historia está
  `pendiente`, `en_progreso` o `completado`, o no está en ningún sprint: se
  acepta en todos los casos y la historia no cambia.
- **CL-023-20** (RN-023-9) — Después de cualquier cambio, la descripción, la
  severidad, la historia relacionada, el sprint de detección y la fecha de carga
  del defecto son los mismos que antes.
- **CL-023-21** (RN-023-10) — Cualquier rechazo (de existencia, de validación o
  del repositorio): el defecto queda exactamente como estaba, con el mismo estado
  y el mismo sprint de resolución.
- **CL-023-22** (RN-023-11) — Dos cambios simultáneos del mismo defecto `abierto`,
  uno a `resuelto` con el Sprint 2 y otro a `descartado`: el segundo `UPDATE` no
  actualiza ninguna fila, porque el estado ya no es `abierto`, y se rechaza con
  `ErrDefectoModificado`. El defecto queda como lo dejó el primero y no se mezclan
  los dos cambios (no queda `descartado` con el Sprint 2). Se verifica con un test
  concurrente contra un Postgres real y llamando al repositorio con un estado
  anterior que ya no es el guardado.
- **CL-023-23** (RN-023-11) — El defecto desaparece entre la lectura y el
  `UPDATE` (se borró su proyecto): el repositorio devuelve
  `ErrDefectoInexistente`. El sprint de resolución desaparece en el mismo
  intervalo: devuelve `ErrSprintInexistente`. En los dos casos no guarda nada. Se
  prueba en el test del repositorio.
- **CL-023-24** (RN-023-12) — Un `UPDATE` hecho directo en la base que deja un
  defecto `resuelto` sin sprint de resolución, uno `en_correccion` con sprint o
  un estado fuera de los cuatro: la base lo rechaza con el `CHECK`. No es un
  camino normal (el dominio ya lo impide); se prueba en el test del repositorio.
- **CL-023-25** (RN-023-13) — Defecto detectado en el "Sprint 1" y resuelto en el
  "Sprint 3" del proyecto: se guarda el ID del Sprint 3 y, al consultar la tabla
  `sprints` con ese ID, su número es 3. Es lo que necesita US-029; se prueba en
  el test del repositorio.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario | Origen |
|---|---|---|---|
| El estado pedido no es `abierto`, `en_correccion`, `resuelto` ni `descartado` | `defecto.ErrEstadoInvalido`, envuelto indicando el valor recibido | "el estado tiene que ser abierto, en_correccion, resuelto o descartado" | Dominio |
| La transición no está permitida, o el estado pedido es el actual | `defecto.ErrTransicionInvalida`, envuelto indicando el estado actual y el pedido | "el defecto no puede pasar a ese estado desde el actual" | Dominio |
| El estado pedido es `resuelto` y no hay sprint de resolución | `defecto.ErrSprintResolucionFaltante` | "un defecto resuelto necesita un sprint de resolucion" | Dominio |
| Hay sprint de resolución y el estado pedido no es `resuelto` | `defecto.ErrSprintResolucionNoCorresponde`, envuelto indicando el estado pedido | "solo un defecto resuelto tiene sprint de resolucion" | Dominio |
| El sprint de resolución es anterior al de detección | `defecto.ErrSprintResolucionAnterior`, envuelto indicando el número de cada uno | "el sprint de resolucion no puede ser anterior al de deteccion" | Dominio |
| El proyecto no existe | `app.ErrProyectoInexistente`, envuelto indicando el ID recibido | "el proyecto no existe" | Aplicación |
| El defecto no existe o no es de ese proyecto | `app.ErrDefectoInexistente`, envuelto indicando el ID recibido | "el defecto no existe" | Aplicación |
| El sprint de resolución no existe o no es de ese proyecto | `app.ErrSprintInexistente`, envuelto indicando el ID recibido | "el sprint no existe" | Aplicación |
| El defecto cambió entre la lectura y el guardado | `app.ErrDefectoModificado`, envuelto indicando el ID | "el defecto cambio mientras se lo editaba" | Persistencia |

Los cinco primeros son del dominio (`internal/domain/defecto`): se detectan
todos en una sola validación, antes de escribir en la base, y se devuelven
unidos con `errors.Join` en el orden de RN-023-6. Con varios errores, el mensaje
es la unión de los mensajes, uno por línea. Los tres siguientes los detecta la
capa de aplicación **antes** de validar los datos, en el orden de la tabla, y se
devuelven de a uno: no se unen con los de validación. El último lo detecta el
repositorio al guardar (RN-023-11), cuando el estado del defecto ya no es el que
leyó el caso de uso, y también se devuelve solo.

`ErrProyectoInexistente` y `ErrSprintInexistente` ya existen en `internal/app`
(el segundo en `asignar_item.go`, de US-009); `ErrDefectoInexistente` y
`ErrDefectoModificado` son nuevos y se definen en `internal/app/errores.go` (el
criterio de `ErrItemInexistente`: lo define quien llegue primero a `main`, y la
otra historia lo reutiliza; US-022 no lo necesita, porque solo busca proyecto,
historia y sprint). Todos son valores de paquete comparables con `errors.Is`, y
en ningún caso se modifica el defecto.

## 8. Criterios de aceptación

- **CA-023-1** (normal) — Con un proyecto, un defecto `abierto` detectado en el
  Sprint 1 y un Sprint 2 existentes, cambiar el defecto a `resuelto` con el
  Sprint 2 lo deja `resuelto`, con el Sprint 2 como sprint de resolución y el
  mismo sprint de detección.
- **CA-023-2** (alternativo) — Un defecto `abierto` pasa a `en_correccion` y de
  ahí a `resuelto` con un sprint; otro pasa de `abierto` a `descartado`. En
  `en_correccion` y en `descartado` el defecto queda sin sprint de resolución.
- **CA-023-3** (alternativo) — Un defecto `resuelto` o `descartado` se reabre:
  queda `abierto` y sin sprint de resolución, y se puede volver a resolver en
  otro sprint.
- **CA-023-4** (alternativo) — Un defecto de una historia completada se puede
  resolver en un sprint cerrado posterior al de su detección, y la historia no
  cambia.
- **CA-023-5** (límite) — El sprint de resolución igual al de detección se
  acepta; uno anterior se rechaza.
- **CA-023-6** (error) — Una transición no permitida (por ejemplo `resuelto` a
  `en_correccion` o `descartado` a `resuelto`) o pedir el mismo estado → rechazo
  indicando el estado actual y el pedido; el defecto no cambia.
- **CA-023-7** (error) — Resolver sin sprint de resolución, o informar uno con
  un estado distinto de `resuelto` → rechazo, no se modifica nada.
- **CA-023-8** (error) — Un sprint de resolución anterior al de detección →
  rechazo indicando los dos números.
- **CA-023-9** (error) — Un estado fuera de `abierto`, `en_correccion`,
  `resuelto` y `descartado` (por ejemplo `cerrado` o vacío) → rechazo.
- **CA-023-10** (error) — Un defecto que no existe o que es de otro proyecto →
  rechazo.
- **CA-023-11** (error) — Un sprint de resolución que no existe o que es de otro
  proyecto → rechazo.
- **CA-023-12** (error) — Proyecto inexistente → rechazo, no se modifica nada.
- **CA-023-13** (error) — Dos errores a la vez (por ejemplo, un defecto
  `descartado` que se quiere pasar a `resuelto` sin sprint) → se rechazan los
  dos juntos, en el orden de RN-023-6.
- **CA-023-14** (error) — Si el defecto cambia de estado entre que se lee y se
  guarda, el segundo cambio se rechaza y no pisa al primero. No tiene frase BDD
  (el escenario tiene un solo `Cuando`): se cubre con tests del repositorio y
  del caso de uso.

---

## Decisiones tomadas y descartadas

- **Los estados los propone esta spec: la plantilla no los lista.** `02-defect.yml`
  nombra el "estado" en su texto introductorio, pero no tiene un campo de estado:
  solo trae la severidad, la historia, el sprint de detección y el de resolución
  (con la opción "sin resolver"). Tampoco los fija ninguna spec del repo, y US-022
  los dejó a cargo de esta historia. Se proponen
  cuatro, con los nombres en minúscula, sin tilde y con guion bajo como
  `en_progreso` de las historias: `abierto` (el que ya usa US-022), `en_correccion`,
  `resuelto` y `descartado`. Se descartó un `cerrado` (resuelto y además
  verificado): ninguna historia pide verificar, y sería un segundo estado
  "resuelto" con las mismas reglas de sprint (ver Huecos).
- **`resuelto` es el único estado que lleva sprint de resolución.** La plantilla
  dice que el sprint de resolución "se completa cuando el defecto se resuelve", y
  US-029 define "resuelto" como "tener sprint de resolución". Con un solo estado
  que lo lleva, las dos definiciones coinciden y se pueden proteger con un
  `CHECK` (`resuelto` si y solo si hay sprint). Se descartó que `descartado`
  también lo lleve: un defecto que no se va a corregir no se "resolvió" en un
  sprint, y contarlo como resuelto inflaría la métrica. Se descartó también que
  el sprint de resolución sea opcional con `resuelto`: US-029 no podría contar
  ese defecto como resuelto en ningún sprint.
- **Sprint de resolución prohibido fuera de `resuelto`, y rechazado si llega.**
  Se descartó ignorarlo en silencio: la persona creería que quedó guardado. La
  pantalla puede ocultar la lista salvo con `resuelto`; la regla vive en el
  dominio.
- **Reabrir deja el sprint de resolución vacío.** Si se conservara, un defecto
  `abierto` tendría sprint de resolución y US-029 lo contaría como resuelto, que
  es lo contrario de reabrirlo. Se descartó guardarlo "por las dudas": sin
  historial (ver más abajo) no hay quién lo lea, y rompería el `CHECK`. Costo
  aceptado: el sprint de la primera resolución se pierde.
- **Transiciones: lo que no está listado está prohibido.** Desde `abierto` y
  `en_correccion` se puede ir a cualquier otro estado (incluido el salto directo
  a `resuelto`, porque en un equipo chico se registra el arreglo sin avisar que
  se empezó), y `en_correccion` puede volver a `abierto` si se deja de trabajar.
  Desde `resuelto` y `descartado` solo se reabre a `abierto`. Se descartó la
  máquina sin restricciones (cualquier estado a cualquier estado): un
  `descartado` que pasa directo a `resuelto` no tendría sentido, y la tabla es lo
  que hace testeable la regla. Se descartó también permitir `resuelto` a
  `en_correccion`: reabrir y empezar a corregir son dos pasos y dejan el
  rastro en el estado.
- **Pedir el estado actual es un error, no un no-op.** US-010 hizo un no-op al
  marcar dos veces porque el doble clic no traía ningún otro dato. Acá sí hay
  otro dato: un `resuelto` a `resuelto` con otro sprint de resolución sería una
  corrección silenciosa del sprint, o un descarte silencioso del sprint pedido.
  Con un error, la persona sabe que no pasó nada y por qué (indica el estado
  actual). Corregir un sprint de resolución equivocado se hace reabriendo y
  volviendo a resolver (ver Huecos). Se descartó el no-op solo cuando el sprint
  coincide: dos reglas para el mismo caso.
- **Un error propio para la transición.** `ErrTransicionInvalida` indica el estado
  actual y el pedido (`de resuelto a en_correccion`), como pidió la consigna. Se
  descartó un error por transición: crecería con cada estado nuevo. Es
  `conflicto` porque depende del estado guardado (RN-T005-2).
- **Falta de sprint de resolución: error propio, no `ErrSprintInexistente`.** En
  US-022 el sprint de detección es siempre obligatorio, y un ID vacío se trata
  como inexistente. Acá el sprint solo es obligatorio con `resuelto`: un ID
  vacío es válido con los demás estados. Por eso el caso de uso solo verifica el
  sprint si se informó, y la falta la detecta el dominio, que es quien sabe a
  qué estado se pasa.
- **El estado actual lo lee el caso de uso, no lo manda el formulario.** Se
  descartó un campo oculto con el estado que la persona vio, para detectar
  pantallas desactualizadas: evaluar la transición contra lo guardado ya rechaza
  lo que no corresponde (CL-023-18) sin sumar un dato que se pueda falsear.
- **La comparación de sprints la hace el dominio, por número, con datos
  traducidos.** El número de sprint (`sprint.Numero`) es lo que usa US-029 y lo
  que permite saber si una resolución es anterior a una detección (con IDs no se
  puede). El defecto guarda IDs (RN-022-6) y el número se lee de `sprints`; el
  caso de uso lo traduce y se lo pasa al dominio, que sigue sin SQL, sin IDs
  generados y sin reloj. Se descartó copiar el número al defecto: duplica un dato
  que ya está en `sprints`. Se descartó comparar en la base con un trigger:
  agrega lógica en SQL para algo que el dominio ya prueba con tests unitarios.
- **El sprint de resolución puede ser de cualquier estado, también uno que no
  empezó.** Igual que RN-022-6 con el de detección: un defecto se suele resolver
  y registrar después, en la Review o cuando el sprint ya cerró, y rechazar un
  sprint cerrado haría imposible registrarlo. Se descartó exigir que ya haya
  empezado: necesita el reloj y una regla con fechas que ninguna historia pide
  (ver Huecos). Se descartó exigirlo abierto, por lo mismo. Lo que sí se exige es
  que sea del mismo proyecto y no anterior al de detección.
- **El defecto se puede cambiar con la historia en cualquier estado.** Como
  RN-022-7 y RN-019-7: un defecto aparece típicamente sobre una historia ya
  completada, y se resuelve después. Se descartó limitarlo, y se descartó que
  cambiar el defecto reabra o toque la historia.
- **Un único `UPDATE` condicionado al estado leído, sin transacción.** Es el
  criterio de las historias que cambian una sola fila (RN-013-7, `WHERE ... AND
  estado <> 'completado'`). El estado y el sprint de resolución se guardan en la
  misma sentencia, así que van juntos o no van. Si dos cambios compiten, con
  `READ COMMITTED` el segundo espera la fila, vuelve a evaluar el `WHERE` y no la
  encuentra: se rechaza con `ErrDefectoModificado` y no se pisan. Se descartó
  bloquear la fila con `FOR NO KEY UPDATE` en una transacción, como US-010: allá
  había que serializar el marcado con el cierre de un sprint, que toca otra
  tabla; acá hay una sola fila y la condición alcanza. Se descartó "gana la
  última escritura" (RN-006-9): con un estado que decide qué transiciones valen,
  dejar que el segundo pise al primero produciría combinaciones que el dominio
  prohíbe (por ejemplo, `descartado` con el sprint del otro cambio).
- **`ErrDefectoModificado` en vez de reintentar o releer.** Si el estado cambió en
  el medio, lo que la persona pidió quizá ya no es válido; reintentar con el
  estado nuevo podría aplicar una transición que no vio. Se le avisa y vuelve a
  cargar.
- **Los `CHECK` van en una migración nueva, y la numeración se confirma al
  abrir la rama.** US-022 dejó `estado` sin lista de valores porque los define
  esta historia, y un `CHECK` con solo `abierto` habría que reemplazarlo. Con los
  cuatro estados definidos, los dos `CHECK` (estado válido, y `resuelto` si y solo
  si hay sprint) siguen el criterio de US-008 y US-010: repetir las reglas del
  dominio como última defensa. La migración no crea columnas, y no podía ir en la
  `00012` (ya es de US-022 y no se edita una mergeada). Se descartó no agregar
  restricciones y confiar solo en el dominio: US-029 depende de que la
  coincidencia entre estado y sprint no se rompa por un `UPDATE` a mano.
- **No hay historial de cambios de estado.** Se guarda solo el estado actual y el
  sprint de resolución actual. Ninguna historia pide el historial ni quién o
  cuándo se cambió (la aplicación no tiene usuarios, como se anotó en US-022);
  US-029 solo necesita el sprint de resolución vigente. Se descartó una tabla de
  cambios: más migración, más consultas y un modelo que nadie va a leer en este
  proyecto. Costo aceptado: reabrir pierde el sprint de la resolución anterior.
- **Compatibilidad con US-029.** La función de US-029 cuenta como resuelto lo que
  tenga `SprintResolucion > 0` y exige `SprintResolucion >= SprintDeteccion`.
  Esta spec cierra los dos huecos de US-029: (1) el estado `resuelto` es el único
  que cuenta, y es el único con sprint de resolución, de modo que "resuelto" en
  US-029 y en esta historia son lo mismo; (2) la regla de orden se impone al
  escribir (RN-023-4), así que la lista que arme la capa de aplicación (ID de
  sprint a número, `NULL` a `0`, RN-023-13) nunca trae un defecto resuelto
  antes de detectarse. No se modifica la spec de US-029.
- **Los `descartado` no se pasan a US-029, en vez de ponerles `0`.** US-029 sugiere
  que quien arma la lista le ponga `SprintResolucion = 0` a un estado que no deba
  contar como resuelto. Pero con `0` el defecto seguiría contando como
  **detectado**, y uno descartado (no era un defecto, o se decidió no corregirlo)
  inflaría los detectados del sprint. RN-029-9 permite filtrar por estado antes de
  pasar la lista, y eso es lo que se hace (RN-023-13). Los `abierto` y
  `en_correccion` sí cuentan como detectados. Es un punto para avisarle a Tomás
  (ver Huecos).
- **Un error propio para "resolución anterior a detección", distinto del de
  US-029.** El de esta historia es
  `defecto.ErrSprintResolucionAnterior`, no `ErrResolucionAnteriorADeteccion`:
  aquel es de `metricas`, se excluye del catálogo porque describe datos del
  sistema que salieron mal, y este se registra como `conflicto` porque lo
  provoca lo que escribió la persona. Dos paquetes con el mismo nombre de error
  serían válidos para Go pero confusos al leer el catálogo.
- **Frases nuevas del diccionario** (aprobadas en el PR de implementación). Las
  cinco de la sección 5 pertenecen al área `Defectos` que define US-022. Las dos
  de contexto hacen falta porque el escenario tiene un solo `Cuando`. No se
  agregan frases para consultar el estado: se reutilizan las de US-022.
- **La pantalla queda fuera de alcance.** Esta historia entrega dominio, caso de
  uso, repositorios y migración, como US-019 y US-022. El formulario depende del
  layout base (T-004).

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia:

1. **Estados del defecto.** La plantilla de defectos no los lista (ver
   Decisiones). Se proponen `abierto`, `en_correccion`, `resuelto` y
   `descartado`. Si el equipo o la cátedra quieren otros nombres, o un estado
   `cerrado` (verificado) o `rechazado`, cambian RN-023-1 a RN-023-3, la tabla de
   transiciones, el `CHECK` de la migración y las frases. Si sumara un segundo
   estado con sprint de resolución, también cambia el `CHECK`.
2. **Los `descartado` y las métricas.** Esta spec los deja fuera de los
   detectados de US-029; Tomás sugiere en US-029 ponerles `SprintResolucion = 0`,
   con lo que seguirían contando como detectados. Confirmar con Tomás cuál es el
   criterio del equipo. Alcanza con ajustar RN-023-13; no cambia el resto.
3. **Corregir el sprint de resolución de un defecto ya resuelto.** Hoy hay que
   reabrirlo y volver a resolverlo (dos cambios), porque `resuelto` a `resuelto`
   es un error. Si el equipo quiere corregirlo en un paso, habría que permitir
   esa transición cuando cambia el sprint, y decidir qué pasa si no cambia.
4. **Sprint de resolución que todavía no empezó.** Se acepta (CL-023-9), como el de
   detección en US-022, porque rechazarlo necesita la fecha de hoy y una regla
   que ninguna historia pide. Confirmar si el equipo quiere pedir que el sprint
   ya haya empezado.
5. **Historial y autoría.** No se guarda cuándo ni quién cambió el estado, ni el
   sprint de una resolución anterior (la aplicación no tiene usuarios). Confirmar
   que ninguna historia posterior (por ejemplo, el reporte en PDF) lo necesita.
6. **Reapertura solo a `abierto`.** Un `resuelto` no pasa directo a
   `en_correccion`. Confirmar que no molesta en el uso real.
7. **Dependencia con la spec de US-022 (PR #179, todavía abierta).** Esta spec usa
   su tabla `defectos` (con `estado` y `sprint_resolucion_id` sin `CHECK`), el
   paquete `defecto` y el área `Defectos` de los steps. Si esa spec cambia el
   nombre de una columna, del paquete o de una frase, hay que ajustar esta antes
   de implementarla. También hay que repartirse el bloque `// defecto` del
   catálogo (T-005): el segundo en llegar a `main` agrega sus líneas.
8. **Los criterios del issue.** El issue #31 no trae la frase "Como / quiero /
   para" (está redactada para esta spec) y sus criterios `CA-023-1` y `CA-023-2`
   siguen vacíos. Los criterios de la sección 8 son los propuestos para el issue.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia, la plantilla de defectos (`02-defect.yml`) y las specs de US-019,
  US-022 y US-029 como modelo, incluida la frase "Como / quiero / para" de la
  sección 1; los estados, las transiciones, la regla del sprint de resolución y
  el control de cambios simultáneos se proponen acá y se confirman en la
  revisión
- Revisado y entendido por: Angelo Conforti
