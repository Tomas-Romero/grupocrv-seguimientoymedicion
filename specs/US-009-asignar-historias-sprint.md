# US-009 — Asignar historias del backlog a un Sprint

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Angelo Conforti |
| **Issue** | #10 |
| **Escenarios BDD** | `features/US-009-asignar-historias-sprint.feature` |
| **Código** | `internal/domain/sprint/`, caso de uso en `internal/app/`, repositorio en `internal/adapters/postgres/`, migración nueva en `migraciones/` |
| **Última actualización** | 2026-10-13 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero asignar historias del Product Backlog a un
> Sprint, para planificar qué trabajo se hace en ese Sprint.

Relacionar un ítem del backlog con un Sprint. Es el paso que convierte trabajo
pendiente en trabajo planificado, y es donde se hace cumplir la Definition of
Ready: solo entra a un Sprint un ítem estimado y con criterios de aceptación.

Depende de **US-005** (el ítem) y de **US-008** (el Sprint, que se crea
`abierto` y del que un proyecto tiene como máximo uno abierto, el último).
Alimenta a US-010 (marcar completadas), US-011 (cerrar el Sprint, que devuelve
al backlog las no completadas) y US-025 (Story Points planificados). La relación
es histórica: registra cada vez que un ítem estuvo en un Sprint y con qué
resultado, y no se borra al cerrar (RN-009-13).

## 2. Entradas

La operación pasa por tres capas, y cada una valida solo lo que puede saber:

- **Dominio** (`internal/domain/sprint`): función pura que recibe los datos
  mínimos del ítem y del Sprint, valida todo y devuelve éxito o todos los
  errores encontrados. No accede a la base ni genera IDs.
- **Aplicación** (`internal/app`): el caso de uso sigue este orden:
  1. Busca el ítem con una lectura simple. Si no existe, devuelve
     `ErrItemInexistente` sin validar nada más.
  2. Busca el Sprint con una lectura simple. Si no existe, devuelve
     `ErrSprintInexistente` sin validar nada más. Si no existen ninguno de los
     dos, se devuelve solo el del ítem.
  3. Arma los datos mínimos del ítem y del Sprint y valida con el dominio, que
     devuelve todos los errores unidos.
  4. Registra la asignación a través del repositorio.
- **Persistencia** (`internal/adapters/postgres`): dentro de una transacción
  bloquea el Sprint y el ítem, vuelve a verificar las reglas con los datos
  leídos con los bloqueos tomados y guarda la relación (RN-009-11).

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `itemID` | UUID | Sí | Un ítem existente del backlog. Lo busca la capa de aplicación y lo vuelve a verificar el repositorio; nunca el dominio |
| `sprintID` | UUID | Sí | Un Sprint existente. Lo busca la capa de aplicación y lo vuelve a verificar el repositorio; nunca el dominio |
| Datos del ítem para el dominio | estructura mínima | Sí | Proyecto, estado (`pendiente`, `en_progreso` o `completado`), Story Points (presentes o ausentes), cantidad de criterios de aceptación y si ya está en el Sprint abierto del proyecto (su Sprint actual) |
| Datos del Sprint para el dominio | estructura mínima | Sí | Proyecto y estado (`abierto` o `cerrado`) |

Los datos mínimos son tipos propios del paquete `sprint`, igual que
`ItemDelSprint` en US-025: el dominio de `sprint` no usa `ItemBacklog`. La capa
de aplicación adapta `ItemBacklog` y `Sprint` a ellos. Lo único que toma de
`backlog` es el error `ErrItemCompletado` (RN-009-5).

No son entradas: el estado del Sprint ni el del ítem (los lee la persistencia) y
la relación misma, que se identifica por el ítem y el Sprint (RN-009-13).

## 3. Salidas esperadas

Si se cumplen todas las reglas, el ítem queda **asignado al Sprint**: la relación
queda registrada con resultado `pendiente` (RN-009-13) y el ítem conserva su
estado, sus Story Points, sus criterios y su prioridad, sin cambios
(RN-009-9). El caso de uso devuelve la asignación ya registrada:

```
Asignacion {
    ItemID   string // el ítem asignado
    SprintID string // el Sprint al que quedó asignado
}
```

Si alguna regla falla, no se registra nada y se devuelve un único error. Si las
reglas que fallan son las del dominio, ese error une todos los encontrados
(RN-009-8).

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-009-1** (aplicación) — El ítem tiene que existir. Si no existe se rechaza
  con `ErrItemInexistente`, sin validar nada más.
- **RN-009-2** (aplicación) — El Sprint tiene que existir. Si no existe se
  rechaza con `ErrSprintInexistente`, sin validar nada más. Estos dos errores
  tienen prioridad sobre los de validación y nunca se unen con ellos. El ítem se
  verifica primero: si faltan los dos, se devuelve solo `ErrItemInexistente`.
- **RN-009-3** (dominio) — El ítem y el Sprint tienen que ser del mismo
  proyecto. Si no, `ErrProyectoDistinto`.
- **RN-009-4** (dominio) — Solo se asignan ítems a un Sprint `abierto`. Si está
  `cerrado`, `ErrSprintCerrado`.
- **RN-009-5** (dominio) — No se asigna un ítem `completado`. Si no,
  `backlog.ErrItemCompletado`, el de US-013; `sprint` no define uno propio. Es
  la única excepción a que `sprint` no importe nada fuera de la biblioteca
  estándar (sección 5): un ítem completado es el mismo concepto en las dos
  historias, y duplicar el error obligaría a T-005 a traducir dos para la misma
  condición.
- **RN-009-6** (dominio) — Un ítem está como máximo en un Sprint abierto: su
  Sprint actual. Si ya está en el Sprint abierto del proyecto, se rechaza con
  `ErrItemYaAsignado`. Haber estado en Sprints cerrados no impide asignarlo de
  nuevo (CL-009-12). No se "mueve" de un Sprint a otro.
- **RN-009-7** (dominio) — Definition of Ready: el ítem tiene que estar
  **estimado** (Story Points presentes, no ausentes) y tener **al menos un
  criterio de aceptación**. Si no, `ErrItemSinEstimar` y `ErrItemSinCriterios`
  respectivamente. Un ítem con `0` Story Points está estimado: esta regla no
  impone la escala Fibonacci (eso es de US-013), solo la ausencia.
- **RN-009-8** (dominio) — El dominio valida todo y devuelve todos los errores
  juntos, unidos con `errors.Join`, en este orden fijo: proyecto distinto,
  Sprint cerrado, ítem completado, ítem ya asignado, ítem sin estimar, ítem sin
  criterios. Cada error unido sigue siendo reconocible con `errors.Is`.
- **RN-009-9** (aplicación y persistencia) — Asignar no modifica el ítem:
  conserva estado, Story Points, criterios y prioridad. Pertenecer a un Sprint
  es una relación y no un estado (US-005).
- **RN-009-10** (aplicación y persistencia) — La operación es de todo o nada: si
  algo falla no se registra nada.
- **RN-009-11** (persistencia) — El repositorio registra la asignación en una
  transacción, con estos pasos en orden, y se detiene en el primer error:
  1. Bloquea la fila del Sprint y lee su proyecto y su estado:
     `SELECT proyecto_id, estado FROM sprints WHERE id = $1 FOR SHARE`. Si no
     hay fila, el Sprint no existe (`ErrSprintInexistente`).
  2. Bloquea la fila del ítem y lee su proyecto, su estado, sus Story Points y
     sus criterios, con `FOR UPDATE`. Si no hay fila, el ítem no existe
     (`ErrItemInexistente`).
  3. Lee si el ítem ya está en el Sprint abierto: una fila de la relación con
     resultado `pendiente`.
  4. Vuelve a verificar RN-009-3 a RN-009-7 con la misma función del dominio,
     pasándole los datos leídos en los pasos 1 a 3 y no los de la lectura del
     caso de uso. Si falla, devuelve lo que devuelve esa función (los errores
     unidos, en el orden de RN-009-8).
  5. Inserta la relación.

  El orden fijo (primero el Sprint, después el ítem) evita bloqueos cruzados. Una
  segunda asignación simultánea del mismo ítem espera el bloqueo del paso 2:
  cuando lo obtiene, la primera ya está registrada, y se rechaza con
  `ErrItemYaAsignado` (CL-009-10). Un cierre simultáneo del Sprint (US-011) que
  modifica su fila espera al `FOR SHARE`, y una asignación que llega después del
  cierre ve el Sprint `cerrado` y se rechaza con `ErrSprintCerrado`
  (CL-009-14). La clave primaria `(item_id, sprint_id)` y el índice único de
  RN-009-13 quedan como red de seguridad.
- **RN-009-12** (dominio) — No se valida la capacidad del Sprint: se pueden
  asignar ítems aunque los Story Points planificados superen la velocidad del
  equipo. Tampoco se compara con la fecha de hoy ni con las fechas del Sprint.
- **RN-009-13** (persistencia) — La relación es **histórica**: tiene una fila
  por cada vez que un ítem estuvo en un Sprint, identificada por
  `(item_id, sprint_id)` y sin ID propio, así que no hay nada que generar con
  `DEFAULT gen_random_uuid()`. Cada fila guarda el resultado del ítem en ese
  Sprint: `pendiente` mientras el Sprint está abierto (asignar crea la fila
  así), y `completado` o `no_completado` cuando US-011 lo cierra. Ninguna fila
  se borra al cerrar. Un ítem tiene como máximo una fila `pendiente`, la de su
  Sprint actual (RN-009-6).

## 5. Restricciones

- **Dominio puro.** `internal/domain/sprint` no importa nada fuera de la
  biblioteca estándar, salvo una excepción: de `internal/domain/backlog` toma
  solo `ErrItemCompletado` (RN-009-5), no `ItemBacklog` ni ninguna otra parte
  del paquete. `backlog` también es dominio puro, así que `sprint` lo sigue
  siendo. No hace SQL ni HTTP, no genera IDs y no llama a `time.Now()`. Sus
  errores propios se definen solo con la biblioteca estándar, y nunca importa
  `internal/platform`.
- **Lo que el dominio no puede saber** (que el ítem y el Sprint existan, su
  estado actual en la base, si el ítem ya está asignado) lo resuelven la capa de
  aplicación y el repositorio, en el orden de la sección 2.
- **La verificación de la transacción es una función del dominio.** Es la misma
  que usa el caso de uso: `internal/domain/sprint` ofrece una función pura que
  recibe los datos mínimos del ítem y del Sprint y devuelve los errores de
  RN-009-3 a RN-009-7 unidos. El repositorio la llama con los datos leídos con
  los bloqueos tomados. Cualquier repositorio la usa: el de Postgres y el que
  usen los escenarios BDD.
- **Persistencia.** Una migración nueva en `migraciones/` (nunca se edita una ya
  mergeada) crea la tabla de la relación, con clave primaria
  `(item_id, sprint_id)`, las referencias al ítem y al Sprint, y la columna del
  resultado: `CHECK` en `pendiente`, `completado` y `no_completado`, y
  `pendiente` por defecto. Como red de seguridad de RN-009-6, lleva un índice
  único parcial sobre `item_id` donde el resultado es `pendiente`. Como
  referencia las tablas de US-005 y de US-008, **tiene que numerarse después de
  las migraciones de esas dos historias**: goose las aplica por número. El
  número se avisa en el grupo antes de crearla. El nombre de la tabla lo define
  la implementación. La validación de la DoR, del estado del Sprint y del
  proyecto no se imponen en la base: las hacen cumplir el dominio y el
  repositorio.
- **Nivel de aislamiento.** La transacción usa el nivel por defecto de Postgres
  (`READ COMMITTED`), y los datos del ítem y de la asignación se leen en
  sentencias posteriores a los bloqueos. Así, una asignación que esperó ve lo
  que registró la otra y ve el estado del Sprint que dejó un cierre concurrente.
- **Verificación de la concurrencia.** RN-009-11, CL-009-10 y CL-009-14 se
  verifican con un test de integración contra un Postgres real.
- **Relación con US-011.** Cerrar el Sprint no borra la relación: completa el
  resultado de cada fila (`completado` o `no_completado`). Los ítems no
  completados vuelven al backlog porque dejan de estar en un Sprint abierto, y
  su estado vuelve a `pendiente`; no se borra nada. Esa operación es de US-011:
  acá solo se define que un ítem que no está en el Sprint abierto se puede
  asignar, aunque haya estado en Sprints cerrados (CL-009-12).
- **Integración con US-025.** Los Story Points de los ítems asignados son la
  entrada de `ItemDelSprint`; la DoR de esta historia garantiza que ninguno
  llegue sin estimar (RN-025-3). Los planificados de un Sprint cerrado salen de
  todas sus filas, completadas o no: por eso la relación no se borra al
  cerrar.
- **Dependencia con T-005.** El tipo de error común de T-005 vive fuera del
  dominio y traduce los errores de dominio a respuestas para el usuario; el
  dominio no lo conoce. Esa traducción tiene que ser compatible con errores
  unidos con `errors.Join`, porque el dominio de esta historia puede devolver
  varios errores a la vez.
- **Frases BDD y cobertura.** Las frases actuales del diccionario crean ítems sin
  criterios de aceptación (`una historia "X" con prioridad "must" y 5 story
  points`). Con la DoR, cualquier escenario que necesite una asignación exitosa
  precisa un ítem con criterios, y para eso se propone una frase nueva (ver
  "Decisiones"). Con las frases actuales alcanzan CA-009-4, CA-009-5 y CA-009-8;
  CA-009-1, CA-009-2, CA-009-3 y CA-009-7 necesitan la frase nueva; CA-009-6
  necesita un Sprint cerrado y llega con US-011 (mismo criterio que CA-008-4 y
  CA-008-9 en US-008). Los mensajes se comprueban con
  `el mensaje de error indica "..."`, que busca el fragmento dentro del mensaje
  completo, también cuando los errores vienen unidos.
- **Impacto en escenarios existentes.** Con la DoR, un escenario que asigne a un
  Sprint una historia creada con las frases actuales se rechaza por falta de
  criterios. Pasa con el ejemplo de US-026 del diccionario
  (`Y una historia "H1" con prioridad "must" y 20 story points` seguida de
  `Y la historia "H1" asignada al sprint "Sprint 1"`) y con los escenarios de
  US-025 que se escriban igual. Esos escenarios, y los pasos compuestos aprobados
  para US-025 y US-026, tienen que crear sus historias con al menos un criterio
  de aceptación. Se resuelve en el PR de implementación de esta historia, que es
  el que introduce la regla, sin esconderla en los steps.
- **Fuera de alcance.** Sacar un ítem de un Sprint a mano (US-011 devuelve al
  backlog las no completadas al cerrar), marcar completadas (US-010), cerrar el
  Sprint (US-011), mover ítems entre Sprints, validar la capacidad, y la
  pantalla de asignación (T-011).

## 6. Casos límite

- **CL-009-1** — Ítem con exactamente un criterio y un Story Point: cumple la
  DoR y se asigna.
- **CL-009-2** — Ítem con `0` Story Points: está estimado, se asigna
  (RN-009-7).
- **CL-009-3** — Ítem sin Story Points (ausentes): se rechaza con
  `ErrItemSinEstimar`, aunque tenga criterios.
- **CL-009-4** — Ítem con lista de criterios vacía: se rechaza con
  `ErrItemSinCriterios`, aunque esté estimado.
- **CL-009-5** — Ítem sin estimar y sin criterios: se devuelven los dos errores
  unidos, en el orden de RN-009-8.
- **CL-009-6** — Primer ítem asignado a un Sprint recién creado: es válido y el
  Sprint pasa a tener 1.
- **CL-009-7** — Ítem ya asignado al mismo Sprint: se rechaza con
  `ErrItemYaAsignado`; no es una operación sin efecto.
- **CL-009-8** — Ítem que quedó completado en un Sprint anterior, ya cerrado:
  al intentar asignarlo al Sprint abierto se rechaza solo con
  `ErrItemCompletado`. No está en el Sprint abierto, así que no es
  `ErrItemYaAsignado` (RN-009-6).
- **CL-009-9** — Ítem de un proyecto y Sprint de otro: se rechaza con
  `ErrProyectoDistinto`.
- **CL-009-10** — Dos asignaciones simultáneas del mismo ítem al mismo Sprint
  abierto: una termina bien y la otra se rechaza con `ErrItemYaAsignado`.
  Ninguna termina con un error de la clave primaria de la relación
  (RN-009-11). Se verifica con un test concurrente contra un Postgres real.
- **CL-009-11** — Ítem o Sprint inexistente junto con un ítem que además no
  cumple la DoR: se devuelve solo el error de inexistencia, sin mirar el resto
  (RN-009-1 y RN-009-2). Si no existen los dos, solo `ErrItemInexistente`.
- **CL-009-12** — Ítem no completado de un Sprint cerrado, devuelto al backlog
  por US-011: conserva su fila en ese Sprint, con resultado `no_completado`, y
  se puede asignar al Sprint abierto. Queda con dos filas, una por Sprint.
- **CL-009-13** — Sprint cerrado y, además, ítem sin estimar: se devuelven
  `ErrSprintCerrado` y `ErrItemSinEstimar` unidos, en el orden de RN-009-8. El
  dominio no corta en el primer error.
- **CL-009-14** — El Sprint se cierra entre la lectura del caso de uso y la
  transacción de la asignación: el repositorio lo ve `cerrado` con la fila
  bloqueada y rechaza con `ErrSprintCerrado`. Se verifica llamando al
  repositorio con un Sprint ya cerrado en la base.
- **CL-009-15** — Asignar hace que los Story Points planificados superen la
  velocidad del equipo: es válido (RN-009-12).

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) |
|---|---|---|
| El ítem no existe | `app.ErrItemInexistente` (el de US-013 y US-019), envuelto indicando el ID recibido | "el item no existe" |
| El Sprint no existe | `app.ErrSprintInexistente`, envuelto indicando el ID recibido | "el sprint no existe" |
| Ítem y Sprint son de proyectos distintos | `sprint.ErrProyectoDistinto` | "la historia y el sprint son de proyectos distintos" |
| El Sprint está cerrado | `sprint.ErrSprintCerrado`, envuelto indicando el nombre del Sprint | "el sprint esta cerrado" |
| El ítem está completado | `backlog.ErrItemCompletado` (el de US-013) | El mensaje general de US-013 (cambio pedido en el #93) |
| El ítem ya está en un Sprint | `sprint.ErrItemYaAsignado` | "la historia ya esta asignada a un sprint" |
| El ítem no tiene Story Points | `sprint.ErrItemSinEstimar` | "la historia no esta estimada: estimala antes de asignarla" |
| El ítem no tiene criterios de aceptación | `sprint.ErrItemSinCriterios` | "la historia no tiene criterios de aceptacion" |

La tercera columna es el texto de origen de cada error. Es el que ve la persona
mientras T-005 no defina otra traducción.

Los seis últimos los devuelve la validación del dominio
(`internal/domain/sprint`, que reutiliza `ErrItemCompletado` de `backlog`): se
detectan todos en una sola validación, antes de escribir nada en la base, y se
devuelven unidos con `errors.Join` en el orden de RN-009-8. Con varios errores,
el mensaje es la unión de los mensajes, uno por línea. El repositorio los
vuelve a verificar dentro de la transacción con la misma función (RN-009-11),
así que también puede devolverlos, unidos del mismo modo, cuando el estado
cambió entre la lectura del caso de uso y la transacción. `ErrItemYaAsignado` y
`ErrSprintCerrado` son los que ocurren en la práctica por una operación
simultánea (CL-009-10 y CL-009-14).

`ErrItemInexistente` y `ErrSprintInexistente` los detecta la capa de aplicación
(`internal/app`) antes de validar los datos, o el repositorio dentro de la
transacción si el ítem o el Sprint desaparecieron en el medio. **No se unen** con
los errores de validación y **tienen prioridad** sobre ellos.

Todos son valores de paquete comparables con `errors.Is`, también cuando vienen
unidos, y la capa de aplicación los envuelve con contexto (`%w`). En ningún caso
se registra nada.

## 8. Criterios de aceptación

- **CA-009-1** (normal) — Dado un ítem estimado y con criterios, y un Sprint
  abierto del mismo proyecto, al asignarlo el ítem queda en ese Sprint.
- **CA-009-2** (alternativo) — Se pueden asignar varios ítems al mismo Sprint, y
  cada uno queda en él.
- **CA-009-3** (límite) — Un ítem con un solo criterio y un Story Point cumple la
  DoR y se asigna.
- **CA-009-4** (error) — Un ítem sin Story Points → rechazo con "la historia no
  esta estimada" y el ítem sigue en el backlog.
- **CA-009-5** (error) — Un ítem sin criterios de aceptación → rechazo con "la
  historia no tiene criterios de aceptacion" y el ítem sigue en el backlog.
- **CA-009-6** (error) — Un Sprint cerrado → rechazo con "el sprint esta
  cerrado".
- **CA-009-7** (error) — Un ítem ya asignado a un Sprint → rechazo con "la
  historia ya esta asignada a un sprint".
- **CA-009-8** (error) — Un ítem o un Sprint inexistente → rechazo y no se
  registra nada.

---

## Decisiones tomadas y descartadas

- **La asignación es una relación, no un estado del ítem.** Misma decisión que
  US-005: los estados describen el avance del trabajo y pertenecer a un Sprint
  es otra cosa. Se descartó agregar un estado `asignado`.
- **La Definition of Ready se valida acá, no en US-005.** US-005 permite
  registrar trabajo todavía sin refinar; el filtro está en la puerta del Sprint.
  Es la regla que sostiene a RN-025-3.
- **Un ítem está como máximo en el Sprint abierto.** Mover un ítem entre Sprints
  esconde decisiones (¿y sus horas?, ¿y su Sprint de origen?) que ningún
  criterio pide. Se descartó el "mover" y se rechaza con `ErrItemYaAsignado`,
  incluso hacia el mismo Sprint, para no tener una operación silenciosa. Haber
  estado en Sprints cerrados no cuenta: un ítem que volvió al backlog se vuelve
  a planificar (CL-009-12).
- **Solo se asigna a un Sprint abierto.** Por US-008, un proyecto tiene como
  máximo un Sprint abierto y es el último, así que asignar a un Sprint cerrado
  sería planificar trabajo en un período que ya terminó. Se descartó permitirlo
  para "completar" un cierre: los ítems no completados vuelven al backlog (US-011).
- **Se devuelven todos los errores juntos.** Mismo criterio que US-001, US-005 y
  US-008: quien planifica ve de una vez todo lo que falta (por ejemplo, estimar y
  agregar criterios), y alinea con la traducción uniforme de T-005.
- **Inexistencia con prioridad y sin unir.** Mismo criterio que US-005 y
  US-008: si el ítem o el Sprint no existen, no tiene sentido evaluar el resto.
  El ítem se verifica primero porque es el protagonista de la operación. Se
  reutiliza el criterio de `ErrProyectoInexistente`. `ErrItemInexistente` es el
  mismo de US-013 y US-019, con su mensaje ("el item no existe");
  `ErrSprintInexistente` es propio de esta historia.
- **Cada capa valida lo que puede saber, y el repositorio vuelve a verificar.**
  El estado del Sprint y el del ítem pueden cambiar entre la lectura del caso de
  uso y el alta (un cierre, una edición), y dos asignaciones simultáneas
  verificadas fuera de la transacción podrían pasar las dos. Por eso el
  repositorio bloquea y vuelve a verificar con la misma función del dominio, sin
  duplicar la regla (mismo enfoque que RN-008-14).
- **`FOR SHARE` en el Sprint y `FOR UPDATE` en el ítem.** `FOR SHARE` deja que
  varias asignaciones al mismo Sprint avancen a la vez, y hace esperar a un
  cierre (US-011), que modifica la fila. `FOR UPDATE` en el ítem serializa las
  asignaciones del mismo ítem, de modo que la segunda ve a la primera y se
  rechaza con `ErrItemYaAsignado` en vez de chocar con la clave primaria. Se
  descartó bloquear el proyecto como en US-008: serializaría todas las altas de
  sprints, ítems y asignaciones de un proyecto sin necesidad.
- **Sin validar la capacidad del Sprint.** Superar la velocidad es una decisión
  de planificación del equipo, no una regla de negocio. Se descartó bloquear la
  asignación; la comparación contra la velocidad (US-026) se verá en el
  Dashboard.
- **Sin desasignación manual.** Ningún criterio la pide; las no completadas
  vuelven al backlog al cerrar el Sprint (US-011). Se descartó para no sumar una
  operación y sus casos límite.
- **Paquete `internal/domain/sprint`.** Es el mismo que define US-008; los
  paquetes van en español (ADR 0003) y `sprint` se mantiene por ser un término
  de Scrum. Los datos mínimos del ítem y del Sprint son tipos propios del
  paquete, para no depender de `ItemBacklog`. Lo único que toma de `backlog` es
  `ErrItemCompletado`.
- **Se reutiliza `backlog.ErrItemCompletado`.** Es la misma condición que en
  US-013 y US-006 (el ítem está completado), así que hay un solo error con un
  mensaje general. Se descartó un `ErrItemCompletado` propio de `sprint`: dos
  errores con el mismo nombre y distinto mensaje para la misma condición, que
  T-005 tendría que traducir por separado. Es la única excepción a que `sprint`
  no importe nada fuera de la biblioteca estándar (sección 5).
- **`ErrSprintCerrado` con mensaje general.** "el sprint esta cerrado" sirve
  también para US-010, que reutiliza el error al marcar y desmarcar. El
  mensaje anterior nombraba la asignación.
- **La relación es histórica.** Una fila por cada vez que un ítem estuvo en un
  Sprint, con su resultado, y ninguna se borra al cerrar. Sin esa historia se
  pierden los Story Points planificados de los Sprints cerrados (US-025: lo
  planificado incluye lo no completado) y el reporte de historias planificadas
  y completadas de cada Sprint. Se descartó borrar la relación al cerrar, que
  era la versión anterior de esta spec. La clave es `(item_id, sprint_id)`: no
  hay ID propio ni `DEFAULT gen_random_uuid()`. Se descartó una clave sintética
  más una restricción de unicidad: guardaría dos veces la misma garantía.
- **Frase de diccionario con criterios, pendiente de acordar con el equipo.**
  Las frases actuales no dan criterios de aceptación y los escenarios de éxito
  los necesitan. Se propone:

  | Frase | Expresión |
  |---|---|
  | `Y una historia "X" con prioridad "must", 5 story points y 2 criterios de aceptacion` | `^una historia "([^"]*)" con prioridad "([^"]*)", (\d+) story points y (\d+) criterios? de aceptacion$` |

  Como todas las expresiones van con ancla, no choca con
  `una historia "X" con prioridad "must" y 5 story points`. Se aprueba con el
  grupo antes de empezar la implementación y la agrega, en su PR, quien la usa
  (esta historia). Si no se aprueba a tiempo, CA-009-1, CA-009-2, CA-009-3 y
  CA-009-7 se verifican con tests unitarios y de integración y no tienen
  escenario BDD, y el `.feature` queda sin caso normal ni alternativo. No se
  inventa la frase en el `.feature`.
- **Los escenarios de CA-009-6 llegan con US-011.** Necesitan un Sprint
  cerrado, y el step `el sprint "X" cerrado` usa el cierre de US-011. Hasta
  entonces, CA-009-6 se cubre con tests unitarios (la función del dominio con el
  Sprint cerrado como dato) y de integración (el repositorio, con un Sprint
  cerrado armado directamente en la base). Mismo criterio que US-008.
- **Los escenarios existentes se ajustan en el PR de implementación.** La DoR
  hace que los escenarios de métricas que asignan historias necesiten ítems con
  criterios (sección 5). Se descartó darles un criterio por defecto en el
  adaptador de los steps: escondería una regla de negocio en la mecánica del
  test.
- **La migración de la relación es un archivo aparte.** Se numera después de las
  de US-005 y US-008, cuando el grupo confirme cuál sigue libre.

## Uso de IA en esta especificación

- [x] Use IA para: redacción del borrador de los apartados de la especificación
  (reglas, casos límite, errores y criterios de aceptación) a partir de la
  historia y de las specs de US-001, US-005, US-008 y US-025 — revisado por:
  Angelo Conforti.
- [x] Use IA para: la corrección de la relación histórica (RN-009-6, RN-009-13,
  sección 5, CL-009-8, CL-009-12, errores y decisiones) y la excepción de
  RN-009-5, redactadas con Claude Code a partir de las decisiones del equipo —
  revisado y entendido por: