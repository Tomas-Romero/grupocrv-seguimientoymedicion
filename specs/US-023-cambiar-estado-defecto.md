# US-023 — Cambiar el estado de un defecto y registrar su sprint de resolución

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Angelo Conforti |
| **Issue** | #31 |
| **Depende de** | US-022 (tabla `defectos`, paquete `defecto`, estado inicial y errores base): se implementa después de que esa historia esté en `main`; esta spec se puede mergear antes, como la de US-019 |
| **Escenarios BDD** | `features/US-023-cambiar-estado-defecto.feature` |
| **Código** | `internal/domain/defecto/` (los estados nuevos y `CambiarEstado`), caso de uso en `internal/app/`, repositorios en `internal/adapters/postgres/` y `internal/adapters/memoria/`, migración nueva `migraciones/00013_defectos_estado.sql` |
| **Última actualización** | 2026-10-26 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero cambiar el estado de un defecto y, cuando
> se resuelve, registrar el sprint en el que se resolvió, para seguir su ciclo
> de vida y medir cuántos defectos resolvemos en cada sprint.

Mover un defecto entre sus estados (`abierto`, `en_correccion`, `resuelto` y
`descartado`) respetando las transiciones permitidas, y guardar el sprint de
resolución cuando pasa a `resuelto`. Completa los campos de la plantilla
`.github/ISSUE_TEMPLATE/02-defect.yml` que US-022 dejó vacíos. Es la materia
prima de los defectos resueltos por sprint (US-029) y de la lista de defectos
(US-024).

## 2. Entradas

La carga pasa por dos capas, y cada una valida solo lo que puede saber:

- **Dominio** (`internal/domain/defecto`): método puro que valida todo y
  devuelve el defecto con el estado nuevo o todos los errores. No lee la base,
  no genera IDs, no lee el reloj y no sabe si los sprints existen: recibe sus
  números ya traducidos.
- **Aplicación y persistencia** (`internal/app` y `internal/adapters/postgres`):
  el caso de uso sigue este orden:
  1. Verifica que el proyecto exista (`ErrProyectoInexistente`) sin verificar
     nada más.
  2. Verifica que el defecto exista **dentro de ese proyecto**
     (`ErrDefectoInexistente`) y lo lee: de ahí sale su estado actual.
  3. Si se informó sprint de resolución, verifica que exista **dentro de ese
     proyecto** (`ErrSprintInexistente`). Un ID vacío es "sin sprint de
     resolución": no se verifica.
  4. Traduce los sprints a su número (`sprint.Numero`): el de detección del
     defecto y el de resolución.
  5. Valida con el dominio, que devuelve todos los errores unidos.
  6. Guarda con un único `UPDATE` condicionado al estado que leyó.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. Viene de la URL; lo verifica el caso de uso, nunca el dominio |
| `defectoID` | UUID | Sí | Un defecto existente de ese proyecto, en cualquier estado. Viene de la URL; lo verifica el caso de uso |
| `estado` | `string` | Sí | El estado pedido: exactamente `abierto`, `en_correccion`, `resuelto` o `descartado`, en minúscula, sin tilde y sin espacios. Viene del formulario; lo valida el dominio |
| `sprintResolucionID` | UUID | Solo con `resuelto` | Un sprint existente del proyecto, abierto o cerrado. Vacío si el estado no es `resuelto`. Viene del formulario |

Un ID de defecto vacío o mal formado es un defecto inexistente. Un
`sprintResolucionID` vacío no: es "sin resolver"; con `resuelto` lo rechaza el
dominio. Uno no vacío que no existe sí es un sprint inexistente.

No son entradas: el estado actual (lo lee el caso de uso, nunca llega del
formulario), los números de sprint ni la fecha y el autor del cambio (no se
guardan). Convertir las etiquetas del formulario en estos valores es tarea del
adaptador HTTP.

## 3. Salidas esperadas

El mismo `Defecto` de US-022 con dos atributos cambiados: el **estado** (el
pedido) y el **sprint de resolución** (el recibido, como ID, si el estado es
`resuelto`; vacío —`NULL`— en cualquier otro estado). El resto no cambia.

```
type SprintDeResolucion struct {
    ID     string // vacío si no se informó
    Numero int    // sprint.Numero, traducido por la capa de aplicación
}

func (d Defecto) CambiarEstado(nuevo Estado, resolucion SprintDeResolucion, numeroDeteccion int) (Defecto, error)
```

El dominio devuelve una copia del defecto; el caso de uso, el defecto ya
guardado. Si hay errores, el dominio devuelve un único error que une todos los
encontrados (RN-023-6) y ningún defecto.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-023-1** (dominio) — Los estados son exactamente cuatro: `abierto` (el
  inicial, RN-022-3), `en_correccion`, `resuelto` y `descartado`. Cualquier otro
  valor, incluido el vacío y los que difieren en mayúsculas, tildes o espacios,
  se rechaza con `ErrEstadoInvalido`, envuelto indicando el valor. El dominio no
  recorta ni pasa a minúscula.
- **RN-023-2** (dominio) — Solo se permiten estas transiciones; cualquier otra,
  incluida la de un estado a sí mismo, se rechaza con `ErrTransicionInvalida`,
  envuelto indicando el estado actual y el pedido (`de resuelto a
  en_correccion`):

  | Actual \ Pedido | `abierto` | `en_correccion` | `resuelto` | `descartado` |
  |---|---|---|---|---|
  | `abierto` | no (mismo) | sí | sí | sí |
  | `en_correccion` | sí | no (mismo) | sí | sí |
  | `resuelto` | sí (reapertura) | no | no (mismo) | no |
  | `descartado` | sí (reapertura) | no | no | no (mismo) |

- **RN-023-3** (dominio) — Solo `resuelto` cuenta como resuelto y solo `resuelto`
  lleva sprint de resolución, que es obligatorio ahí y prohibido en los demás
  estados. `resuelto` sin sprint se rechaza con `ErrSprintResolucionFaltante`; otro
  estado con sprint, con `ErrSprintResolucionNoCorresponde` (no se descarta en
  silencio).
- **RN-023-4** (dominio) — Con `resuelto`, el número del sprint de resolución
  tiene que ser mayor o igual que el del sprint de detección. Uno anterior se
  rechaza con `ErrSprintResolucionAnterior`, envuelto con los dos números; igual
  se acepta. Se compara por **número** (`sprint.Numero`), no por ID. Solo se
  evalúa con `resuelto` y un sprint informado.
- **RN-023-5** (dominio) — Salvo en `resuelto`, el defecto queda sin sprint de
  resolución. Reabrir un `resuelto` (pasarlo a `abierto`) deja el sprint vacío y
  no conserva el anterior; si se vuelve a resolver, queda el sprint que se
  informe entonces.
- **RN-023-6** (dominio) — El dominio devuelve todos los errores juntos, unidos
  con `errors.Join`, en este orden fijo: `ErrEstadoInvalido`; con el estado
  válido, `ErrTransicionInvalida`, después `ErrSprintResolucionFaltante` o
  `ErrSprintResolucionNoCorresponde` (excluyentes) y al final
  `ErrSprintResolucionAnterior`. Con un estado inválido se devuelve solo
  `ErrEstadoInvalido`: sin saber a qué estado se pasa, las demás reglas no se
  pueden evaluar. Cada error se reconoce con `errors.Is`.
- **RN-023-7** (aplicación) — El proyecto, el defecto y el sprint de resolución
  (si se informó) tienen que existir y los dos últimos ser **de ese proyecto**.
  Se verifican en ese orden (`ErrProyectoInexistente`, `ErrDefectoInexistente`,
  `ErrSprintInexistente`) antes de validar los datos; el primero que falla se
  devuelve y no se une con los de validación.
- **RN-023-8** (aplicación) — El estado actual lo lee el caso de uso del defecto
  guardado y los números de sprint los traduce a partir de los IDs antes de
  llamar al dominio, que nunca recibe un ID para compararlo.
- **RN-023-9** (aplicación) — Se puede cambiar el estado sea cual sea el estado
  de la historia relacionada, y el sprint de resolución puede ser cualquier
  sprint del proyecto: abierto o cerrado, haya empezado o no. El cambio solo
  modifica el estado y el sprint de resolución del defecto.
- **RN-023-10** (aplicación) — Es de todo o nada: estado y sprint se guardan
  juntos o no se guarda nada. Si algo falla, el defecto queda como estaba.
- **RN-023-11** (persistencia) — Se guarda con un único `UPDATE ... SET estado =
  $3, sprint_resolucion_id = $4 WHERE id = $1 AND proyecto_id = $2 AND estado =
  $5`, donde `$5` es el estado que leyó el caso de uso. Si no actualiza ninguna
  fila, el repositorio relee: si el defecto no existe en el proyecto,
  `ErrDefectoInexistente`; si existe, otro cambio se adelantó y devuelve
  `ErrDefectoModificado` sin guardar nada. No hay transacción ni bloqueo. Si el
  sprint de resolución desaparece, la clave foránea lo rechaza y se devuelve
  `ErrSprintInexistente`.
- **RN-023-12** (persistencia) — La base repite dos reglas con `CHECK` como
  última defensa: el estado es uno de los cuatro, y hay sprint de resolución si
  y solo si el estado es `resuelto`. La comparación de números (RN-023-4) y la
  pertenencia al proyecto no caben en un `CHECK` de una fila: las hacen cumplir
  el dominio y el caso de uso.
- **RN-023-13** (consulta) — Compatibilidad con US-029. La capa de aplicación
  arma su lista con los defectos del proyecto que **no** están `descartado`:
  `SprintDeteccion` es el número del sprint de detección y `SprintResolucion` el
  del de resolución, o `0` si es `NULL`. Por RN-023-3, RN-023-4 y RN-023-12,
  `SprintResolucion > 0` equivale a `resuelto` y nunca es anterior a la
  detección, así que US-029 no recibe datos corruptos y su spec no cambia. Los
  `descartado` se dejan afuera (RN-029-9) y no se les pone `0`: seguirían
  contando como detectados. La consulta la escribe quien integre US-029 (US-030).

## 5. Restricciones

- **Dominio puro.** `internal/domain/defecto` no importa nada fuera de la
  biblioteca estándar, no hace SQL ni HTTP, no genera IDs y no llama a
  `time.Now()`. Agrega al paquete de US-022 `EstadoEnCorreccion`,
  `EstadoResuelto`, `EstadoDescartado`, `SprintDeResolucion` y `CambiarEstado`.
  No importa `sprint` ni `metricas`.
- **Pertenencia al proyecto.** Las claves foráneas impiden un sprint inexistente
  pero no uno de otro proyecto: lo verifica el caso de uso, como en US-022. No se
  desactualiza entre la lectura y el `UPDATE` porque ninguna historia mueve
  sprints o defectos de proyecto ni cambia su sprint de detección o su número.
- **Persistencia.** La migración `00013_defectos_estado.sql` (nunca se edita una
  mergeada) no crea columnas, porque `estado` y `sprint_resolucion_id` ya las crea
  la `00012`: agrega `CONSTRAINT estado_de_defecto_valido CHECK (estado IN
  ('abierto', 'en_correccion', 'resuelto', 'descartado'))` y `CONSTRAINT
  resolucion_segun_estado CHECK ((estado = 'resuelto') = (sprint_resolucion_id IS
  NOT NULL))`; su `Down` las borra. Al aplicarse no hay defectos que las violen.
- **Número de la migración.** La `00011` es de US-018 (PR #170) y la `00012`, de
  US-022 (PR #179). Al 2026-10-26, con US-022 ya en `main`, la `00012` está en
  `main` y ninguna rama ni PR abierto usa la `00013`: la implementación lo
  confirmó antes de crearla. Esta migración modifica la tabla de la `00012`, y
  `goose -allow-missing` (T-012) no la aplica antes de que exista.
  Con los `CHECK` nuevos, el test de US-022 que afirmaba que `estado` no tenía
  `CHECK` de valores (`TestDefectos_EstadoSinCheckDeValores`) deja de tener sentido y
  se reemplaza por los de CL-023-16; el de la clave foránea del sprint de resolución
  (`TestDefectos_SprintDeResolucionConClaveForanea`) pasa a escribir el estado
  `resuelto` junto con el sprint, porque la base ya no admite uno sin el otro. Se
  ajustan en este PR; ningún nombre público de US-022 cambia.
- **Dependencias.** US-022 (tabla, paquete, `Estado`, área `Defectos` de los
  steps) y US-008 (`sprints`, con su número). No depende de US-024 ni US-029.
- **Puertos y caso de uso.** El caso de uso es `app.CambiarEstadoDefecto`
  (`NuevoCambiarEstadoDefecto`, `Ejecutar(ctx, proyectoID, defectoID, estado,
  sprintResolucionID)`) y su puerto, `app.RepositorioCambioDeEstadoDefecto`, con
  `ExisteProyecto`, `DefectoEnProyecto` (devuelve el defecto guardado y si existe en
  ese proyecto), `NumeroDeSprintEnProyecto` (verifica que el sprint sea del proyecto
  y devuelve su `sprint.Numero`) y `ActualizarEstadoDefecto(ctx, defecto,
  estadoAnterior)`. Lo implementan `postgres.RepositorioDefectos` y `memoria.Base`.
- **Catálogo de errores (T-005).** Los siete errores nuevos se registran en
  `internal/adapters/http/catalogo.go` en el PR de implementación, sin
  exclusiones nuevas. En el bloque `// defecto` (el segundo PR que llegue a
  `main` agrega sus líneas al que ya exista): `ErrEstadoInvalido` como
  `validacion` con campo `estado`; `ErrTransicionInvalida` como `conflicto` con
  campo `estado`; `ErrSprintResolucionFaltante` y
  `ErrSprintResolucionNoCorresponde` como `validacion` con campo
  `sprint_resolucion`; y `ErrSprintResolucionAnterior` como `conflicto` con campo
  `sprint_resolucion`, porque compara contra el sprint guardado (RN-T005-2). No
  reutiliza `metricas.ErrResolucionAnteriorADeteccion`, que se excluye del
  catálogo por ser un dato corrupto del sistema. En el bloque `// app`:
  `ErrDefectoInexistente` (`no_encontrado`) y `ErrDefectoModificado`
  (`conflicto`), sin campo. Los nombres de campo del catálogo son una lista
  cerrada (RN-T005-8): `estado` y `sprint_resolucion` no estaban, y el PR de
  implementación los suma a esa regla (spec de T-005) y a la lista de
  `TestCatalogo_EntradasValidas`, sin tocar ninguna otra entrada, como hizo US-022
  con `descripcion` y `severidad`.
- **Steps.** Esta historia no crea un área: suma a la interfaz `Defectos` de
  US-022 los métodos `CambiarEstadoDeDefecto(historia, estado, sprintResolucion
  string) error` y `SprintDeResolucionDeDefecto(historia string) (sprint string,
  resuelto bool, err error)`, y estas frases (ADR 0002), que se aprueban en el PR
  de implementación y se suman a `docs/diccionario-steps.md`:
  - `Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "en_correccion"`
  - `Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 2"`
  - `Y el defecto de la historia "Crear proyecto" en estado "en_correccion"` (contexto)
  - `Y el defecto de la historia "Crear proyecto" resuelto en el sprint "Sprint 2"` (contexto)
  - `Y el defecto de la historia "Crear proyecto" fue resuelto en el sprint "Sprint 2"`

  Las de contexto existen porque el escenario tiene un solo `Cuando`. Todas operan sobre el último defecto de la historia y buscan los sprints por nombre en
  el proyecto de esa historia (los métodos no reciben el proyecto); lo que no existe
  (la historia, su defecto o el sprint) se manda con un ID inexistente, así el
  rechazo lo produce la aplicación. Las dos de contexto usan el mismo caso de uso: la
  de `en estado` sirve para `en_correccion`, `descartado` y la reapertura (no lleva
  sprint, así que no pasa a `resuelto`) y pedir el estado en el que el defecto ya
  está se rechaza, como en el dominio. "De otro proyecto", el cambio simultáneo y los
  `CHECK` los prueban el caso de uso y el repositorio.
- **Fuera de alcance.** Registrar (US-022) y listar (US-024) defectos, la
  pantalla (T-004), editar o borrar defectos, el historial de cambios y las
  métricas (US-029, US-030, US-032).

## 6. Casos límite

- **CL-023-1** (RN-023-2) — Transiciones permitidas: `abierto` a `en_correccion`,
  a `descartado` y a `resuelto` con el Sprint 2; `en_correccion` a `abierto`, a
  `descartado` y a `resuelto` con el Sprint 2; `resuelto` a `abierto`;
  `descartado` a `abierto`. El defecto queda con el estado pedido.
- **CL-023-2** (RN-023-2) — Transiciones prohibidas (`resuelto` a
  `en_correccion` o `descartado`; `descartado` a `en_correccion` o `resuelto`) y
  pedir el estado actual (`abierto` a `abierto`, `resuelto` a `resuelto` aun con
  otro sprint): `ErrTransicionInvalida` con los dos estados, y el defecto no
  cambia. El mismo estado no es un no-op.
- **CL-023-3** (RN-023-1) — Estado `cerrado`, `Resuelto`, `" resuelto "`,
  `en correccion`, `en_corrección` o vacío: `ErrEstadoInvalido`.
- **CL-023-4** (RN-023-3) — `resuelto` sin sprint (ID vacío):
  `ErrSprintResolucionFaltante`. `en_correccion`, `descartado` o `abierto` con el
  Sprint 2: `ErrSprintResolucionNoCorresponde`.
- **CL-023-5** (RN-023-4) — Detectado en el Sprint 3 y resuelto en el Sprint 2:
  `ErrSprintResolucionAnterior` con los números 2 y 3. Resuelto en el Sprint 3
  (el mismo) o en el 4: se acepta. Detectado y resuelto en el Sprint 1: se acepta
  (borde inferior).
- **CL-023-6** (RN-023-4, RN-023-9) — Sprint de resolución cerrado, abierto o
  que todavía no empezó (fechas posteriores a hoy): se acepta (ver Huecos).
- **CL-023-7** (RN-023-5) — Defecto resuelto en el Sprint 2 y reabierto: queda
  `abierto` y sin sprint (`NULL`); si se resuelve de nuevo en el Sprint 3, queda
  con el 3, no con el 2. Un `descartado` reabierto queda `abierto`, sin sprint.
- **CL-023-8** (RN-023-13) — El defecto reabierto de CL-023-7 deja de contar
  como resuelto en el Sprint 2 (su `SprintResolucion` es `0`). Uno `descartado`
  no aparece en la lista de US-029.
- **CL-023-9** (RN-023-6) — Varios errores a la vez: `descartado` que se pasa a
  `resuelto` sin sprint devuelve `ErrTransicionInvalida` y
  `ErrSprintResolucionFaltante`; `resuelto` detectado en el Sprint 3 que se
  resuelve de nuevo en el 2 devuelve `ErrTransicionInvalida` y
  `ErrSprintResolucionAnterior`. Nunca hay más de dos. Con estado `urgente`, con
  o sin sprint, solo `ErrEstadoInvalido`.
- **CL-023-10** (RN-023-7, RN-023-10) — Proyecto inexistente, aunque el defecto
  y el sprint no existan y los datos sean inválidos: solo
  `ErrProyectoInexistente`. Defecto inexistente, con ID vacío o mal formado, o de
  otro proyecto: solo `ErrDefectoInexistente`. Sprint de resolución inexistente,
  mal formado o de otro proyecto (por ejemplo con `en_correccion`): solo
  `ErrSprintInexistente`. Un sprint **vacío** con `resuelto` no es inexistente:
  es CL-023-4. En todos los casos no se modifica nada.
- **CL-023-11** (RN-023-8) — La pantalla muestra el defecto `abierto` pero otra
  persona ya lo resolvió: pedir `descartado` se evalúa contra `resuelto` y se
  rechaza (`de resuelto a descartado`); pedir `abierto` es una reapertura válida.
- **CL-023-12** (RN-023-9) — Historia relacionada `pendiente`, `en_progreso`,
  `completado` o sin sprint: se acepta y la historia no cambia. Después de
  cualquier cambio, la descripción, la severidad, la historia, el sprint de
  detección y la fecha de carga son los mismos.
- **CL-023-13** (RN-023-10) — Cualquier rechazo deja el defecto con el mismo
  estado y el mismo sprint de resolución.
- **CL-023-14** (RN-023-11) — Dos cambios simultáneos del mismo defecto `abierto`
  (uno a `resuelto` con el Sprint 2 y otro a `descartado`): el segundo `UPDATE` no
  actualiza ninguna fila y se rechaza con `ErrDefectoModificado`. El defecto
  queda como lo dejó el primero, sin mezclar (no queda `descartado` con el Sprint
  2). La prueba principal es determinista: llama al repositorio con un
  estado anterior que ya no es el guardado y comprueba `ErrDefectoModificado` sin
  guardar nada. Un test con dos goroutines contra Postgres real la complementa y se
  escribe sin depender de tiempos ni de quién llega primero: exactamente un cambio
  guarda, el otro recibe `ErrDefectoModificado` y el defecto queda como lo dejó el que
  ganó.
- **CL-023-15** (RN-023-11) — El defecto (se borró su proyecto) o el sprint de
  resolución desaparecen entre la lectura y el `UPDATE`: `ErrDefectoInexistente`
  o `ErrSprintInexistente`, sin guardar nada. Se prueba en el repositorio. Si lo que
  falta entre la lectura y la comparación es el sprint de detección del defecto
  guardado, no hay número contra el que comparar: el caso de uso devuelve
  `ErrSprintInexistente` en vez de comparar contra un `0`.
- **CL-023-16** (RN-023-12) — Un `UPDATE` directo en la base que deja un
  `resuelto` sin sprint, un `en_correccion` con sprint o un estado fuera de los
  cuatro: la base lo rechaza con el `CHECK`. Se prueba en el repositorio.
- **CL-023-17** (RN-023-13) — Defecto detectado en el "Sprint 1" y resuelto en el
  "Sprint 3": se guarda el ID del Sprint 3 y, al consultar `sprints` con ese ID,
  su número es 3. Se prueba en el repositorio.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario | Origen |
|---|---|---|---|
| Estado pedido fuera de los cuatro | `defecto.ErrEstadoInvalido`, envuelto con el valor | "el estado tiene que ser abierto, en_correccion, resuelto o descartado" | Dominio |
| Transición no permitida, o estado pedido igual al actual | `defecto.ErrTransicionInvalida`, envuelto con el estado actual y el pedido | "el defecto no puede pasar a ese estado desde el actual" | Dominio |
| `resuelto` sin sprint de resolución | `defecto.ErrSprintResolucionFaltante` | "un defecto resuelto necesita un sprint de resolucion" | Dominio |
| Sprint de resolución con un estado distinto de `resuelto` | `defecto.ErrSprintResolucionNoCorresponde`, envuelto con el estado pedido | "solo un defecto resuelto tiene sprint de resolucion" | Dominio |
| Sprint de resolución anterior al de detección | `defecto.ErrSprintResolucionAnterior`, envuelto con los dos números | "el sprint de resolucion no puede ser anterior al de deteccion" | Dominio |
| El proyecto no existe | `app.ErrProyectoInexistente`, envuelto con el ID | "el proyecto no existe" | Aplicación |
| El defecto no existe o no es de ese proyecto | `app.ErrDefectoInexistente`, envuelto con el ID | "el defecto no existe" | Aplicación |
| El sprint de resolución no existe o no es de ese proyecto | `app.ErrSprintInexistente`, envuelto con el ID | "el sprint no existe" | Aplicación |
| El defecto cambió entre la lectura y el guardado | `app.ErrDefectoModificado`, envuelto con el ID | "el defecto cambio mientras se lo editaba" | Persistencia |

Los cinco primeros son del dominio y se devuelven unidos en el orden de
RN-023-6 (el mensaje es la unión, uno por línea). Los tres siguientes los
detecta la aplicación **antes** de validar, en el orden de la tabla, y se
devuelven de a uno; el último, el repositorio (RN-023-11), también solo.
`ErrDefectoInexistente` y `ErrDefectoModificado` son nuevos y van en
`internal/app/errores.go`; los otros dos ya existen. Todos son valores de
paquete comparables con `errors.Is`.

## 8. Criterios de aceptación

- **CA-023-1** (normal) — Un defecto `abierto` detectado en el Sprint 1, cambiado
  a `resuelto` con el Sprint 2, queda `resuelto` con ese sprint de resolución y
  el mismo sprint de detección.
- **CA-023-2** (alternativo) — Un defecto pasa de `abierto` a `en_correccion` y de
  ahí a `resuelto` con un sprint; otro pasa de `abierto` a `descartado`. En
  `en_correccion` y `descartado` queda sin sprint de resolución.
- **CA-023-3** (alternativo) — Un defecto `resuelto` o `descartado` se reabre:
  queda `abierto`, sin sprint de resolución, y se puede resolver de nuevo en otro
  sprint.
- **CA-023-4** (alternativo) — Un defecto de una historia completada se resuelve
  en un sprint cerrado posterior al de su detección, y la historia no cambia.
- **CA-023-5** (límite) — Un sprint de resolución igual al de detección se
  acepta; uno anterior se rechaza.
- **CA-023-6** (error) — Una transición no permitida o el mismo estado → rechazo
  indicando el estado actual y el pedido; el defecto no cambia.
- **CA-023-7** (error) — Resolver sin sprint de resolución, o informar uno con un
  estado distinto de `resuelto` → rechazo.
- **CA-023-8** (error) — Un sprint de resolución anterior al de detección →
  rechazo indicando los dos números.
- **CA-023-9** (error) — Un estado fuera de los cuatro (por ejemplo `cerrado` o
  vacío) → rechazo.
- **CA-023-10** (error) — Un defecto que no existe o es de otro proyecto →
  rechazo.
- **CA-023-11** (error) — Un sprint de resolución que no existe o es de otro
  proyecto → rechazo.
- **CA-023-12** (error) — Proyecto inexistente → rechazo, no se modifica nada.
- **CA-023-13** (error) — Dos errores a la vez (un `descartado` que se pasa a
  `resuelto` sin sprint) → se rechazan juntos, en el orden de RN-023-6.
- **CA-023-14** (error) — Si el defecto cambia de estado entre que se lee y se
  guarda, el segundo cambio se rechaza y no pisa al primero. No tiene frase BDD
  (un solo `Cuando`): se cubre con tests del repositorio y del caso de uso.

---

## Decisiones tomadas y descartadas

- **Los estados los propone esta spec: la plantilla no los lista.**
  `02-defect.yml` nombra el "estado" en su introducción pero no tiene ese campo,
  y US-022 los dejó a esta historia. Se usan nombres en minúscula, sin tilde y
  con guion bajo como `en_progreso`. Se descartó un `cerrado` (resuelto y
  verificado): ninguna historia pide verificar, y sería un segundo estado
  "resuelto" con las mismas reglas de sprint.
- **`descartado` no lleva sprint de resolución ni cuenta como resuelto.** Un
  defecto que no se va a corregir no se "resolvió" en un sprint, y contarlo
  inflaría la métrica. Con un solo estado que lleva sprint, "resuelto" en US-029
  (tener sprint) y en esta historia coinciden y se protegen con un `CHECK`.
- **Transiciones: lo no listado está prohibido.** Desde `abierto` y
  `en_correccion` se puede ir a cualquier otro estado, incluido el salto directo
  a `resuelto` (en un equipo chico se registra el arreglo sin avisar que se
  empezó). Desde `resuelto` y `descartado` solo se reabre. Se descartó la máquina
  sin restricciones y que `resuelto` pase directo a `en_correccion`: reabrir y
  empezar a corregir son dos pasos.
- **Pedir el estado actual es un error, no un no-op.** US-010 hizo un no-op al
  marcar dos veces porque el doble clic no traía otro dato; acá un `resuelto` a
  `resuelto` con otro sprint sería una corrección silenciosa o un descarte
  silencioso del sprint pedido. Se descartó el no-op solo si el sprint coincide:
  dos reglas para el mismo caso.
- **La falta de sprint de resolución tiene error propio.** En US-022 el sprint de
  detección es siempre obligatorio y un ID vacío es inexistente. Acá solo es
  obligatorio con `resuelto`, así que el caso de uso verifica el sprint solo si
  se informó y la falta la detecta el dominio, que sabe a qué estado se pasa.
- **El estado actual no viaja en el formulario.** Se descartó un campo oculto con
  lo que la persona vio: evaluar contra lo guardado ya rechaza lo que no
  corresponde (CL-023-11) sin un dato que se pueda falsear.
- **La comparación de sprints es del dominio, por número, con datos traducidos.**
  Con IDs no se puede saber si una resolución es anterior a una detección. Se
  descartó copiar el número al defecto (duplica `sprints`) y comparar en la base
  con un trigger (lógica en SQL para algo que el dominio ya prueba).
- **Cualquier sprint del proyecto sirve para resolver, incluido uno que no
  empezó.** Como RN-022-6: un defecto se suele resolver y registrar después, en
  la Review o con el sprint ya cerrado. Se descartó exigir sprint abierto o ya
  iniciado: necesita el reloj y una regla con fechas que ninguna historia pide.
- **Un `UPDATE` condicionado al estado leído, sin transacción.** Es el criterio de
  las historias que cambian una fila (RN-013-7). Con `READ COMMITTED` el segundo
  `UPDATE` espera la fila, reevalúa el `WHERE` y no la encuentra. Se descartó
  `FOR NO KEY UPDATE` como US-010: allá había que serializar con el cierre de un
  sprint, que toca otra tabla. Se descartó "gana la última escritura"
  (RN-006-9): dejaría combinaciones prohibidas, como `descartado` con el sprint
  del otro cambio. `ErrDefectoModificado` no reintenta: con el estado cambiado,
  lo pedido quizá ya no es válido.
- **Los `CHECK` van en una migración nueva.** US-022 dejó `estado` sin lista
  porque los valores los define esta historia, y la `00012` no se edita. Siguen el
  criterio de US-008 y US-010 (repetir el dominio como última defensa). Se
  descartó confiar solo en el dominio: US-029 depende de que estado y sprint no
  se desacoplen por un `UPDATE` a mano.
- **Sin historial de cambios.** Se guarda solo el estado y el sprint de
  resolución actuales: ninguna historia pide quién o cuándo (no hay usuarios, como
  en US-022) y US-029 solo necesita el sprint vigente. Costo aceptado: reabrir
  pierde el sprint de la resolución anterior.

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia:

1. **Estados.** Se proponen `abierto`, `en_correccion`, `resuelto` y
   `descartado`. Si el equipo o la cátedra quieren otros (un `cerrado`, un
   `rechazado`), cambian RN-023-1 a RN-023-3, la tabla de transiciones, el
   `CHECK` y las frases.
2. **`descartado` y las métricas.** Esta spec los deja fuera de los detectados de
   US-029; la spec de Tomás sugiere ponerles `SprintResolucion = 0`, con lo que
   seguirían contando. Hay que confirmar el criterio con Tomás: solo cambia
   RN-023-13.
3. **Reapertura y corrección.** Un `resuelto` solo vuelve a `abierto` (no a
   `en_correccion`), y corregir el sprint de resolución exige reabrir y resolver
   de nuevo. Confirmar que no molesta, o permitir `resuelto` a `resuelto` cuando
   cambia el sprint.
4. **Sprint de resolución que todavía no empezó.** Se acepta (CL-023-6) porque
   rechazarlo necesita la fecha de hoy. Confirmar si se quiere pedir que ya haya
   empezado.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia, la plantilla de defectos (`02-defect.yml`) y las specs de US-019,
  US-022 y US-029 como modelo, incluida la frase "Como / quiero / para" de la
  sección 1; los estados, las transiciones, la regla del sprint de resolución y
  el control de cambios simultáneos se proponen acá y se confirman en la
  revisión
- Revisado y entendido por: Angelo Conforti
