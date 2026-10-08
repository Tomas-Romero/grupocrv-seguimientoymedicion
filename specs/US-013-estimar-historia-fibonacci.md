# US-013 — Estimar una historia con Story Points en escala Fibonacci

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #21 |
| **Escenarios BDD** | `features/US-013-estimar-historia-fibonacci.feature` |
| **Código** | `internal/domain/backlog/` (`StoryPoints`, `ItemBacklog.Estimar`), caso de uso en `internal/app/`, repositorio en `internal/adapters/postgres/` |
| **Última actualización** | 2026-10-07 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero asignarle Story Points a una historia del
> backlog usando la escala Fibonacci, para dimensionar el trabajo antes de
> planificar el sprint.

Poner (o cambiar) la estimación de un ítem del Product Backlog con un valor de
la escala Fibonacci del proyecto: 1, 2, 3, 5, 8 o 13. Es lo que vuelve a un
ítem apto para entrar a un sprint (la Definition of Ready, que valida US-009) y
la base de las sumas de US-025, de la velocidad de US-026 y de la comparación
estimado contra real de US-021.

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `itemID` | UUID | Sí | Un ítem existente. No es un campo del formulario: viene del contexto de navegación (la URL). Lo verifica el caso de uso; nunca el dominio |
| `puntos` | `int` | Sí | Exactamente uno de `1`, `2`, `3`, `5`, `8`, `13` |

La estimación pasa por dos capas, y cada una valida solo lo que puede saber:

- **Dominio** (`internal/domain/backlog`): `ItemBacklog.Estimar(puntos)` es una
  función pura que valida el valor y el estado del ítem y devuelve una copia del
  ítem estimada, o el error. No lee la base ni el reloj.
- **Aplicación y persistencia** (`internal/app` y `internal/adapters/postgres`):
  el caso de uso sigue este orden:
  1. Carga el ítem. Si no existe, devuelve `ErrItemInexistente` sin validar nada
     más.
  2. Llama a `Estimar` del dominio con el valor recibido.
  3. Guarda el ítem estimado a través del repositorio, con un único `UPDATE`
     sobre la columna de Story Points.

El valor viene de un formulario como texto. Convertirlo a `int` es tarea del
adaptador HTTP, no del dominio; si no es un número entero, el adaptador responde
con su propio error de formato y no llama al caso de uso.

## 3. Salidas esperadas

El `ItemBacklog` con los mismos datos que tenía y los Story Points puestos:

| Atributo | Valor después de estimar |
|---|---|
| Story Points | `puntos` (estimado), distinto de "sin estimar" |
| Estado, título, prioridad, criterios, número | Sin cambios |

Si hay error, no se devuelve ítem y no se modifica nada en la base.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-013-1** (dominio) — La escala es 1, 2, 3, 5, 8 y 13. Cualquier otro
  valor se rechaza, incluidos `0`, los negativos y los números Fibonacci que no
  están en la escala (por ejemplo `4`, `21` o `34`).
- **RN-013-2** (dominio) — `0` no es una estimación: un ítem "sin estimar" y un
  ítem "estimado en 0" no significan lo mismo (RN-005-7). Sin estimar se
  representa con el valor ausente de `StoryPoints`, y esta historia no permite
  volver a ese estado.
- **RN-013-3** (dominio) — Un ítem `pendiente` o `en_progreso` se puede
  estimar y volver a estimar las veces que haga falta: la última estimación
  reemplaza a la anterior.
- **RN-013-4** (dominio) — Un ítem `completado` no se puede estimar ni
  re-estimar, porque cambiaría hacia atrás los Story Points completados de un
  sprint ya medido (US-025 y US-026).
- **RN-013-5** (dominio) — Se valida primero el estado del ítem y después el
  valor. Si el ítem está completado, ese es el único error, aunque el valor
  también sea inválido.
- **RN-013-6** (aplicación) — Si el ítem no existe, el caso de uso devuelve
  `ErrItemInexistente` antes de llamar al dominio. Ese error no se une con los
  del dominio.
- **RN-013-7** (persistencia) — La estimación se guarda con un único `UPDATE`
  sobre el ítem. No hace falta bloquear filas: si dos personas estiman a la vez,
  gana la última escritura.
- **RN-013-8** (aplicación) — Si cualquier validación falla, no se modifica
  nada.

## 5. Restricciones

- **Dominio puro.** `internal/domain/backlog` no importa nada fuera de la
  biblioteca estándar, no hace SQL ni HTTP y no lee el reloj.
- **Tipo `StoryPoints`.** Sigue siendo el struct de US-005 (valor cero =
  "sin estimar"). Se agrega el constructor `NuevosStoryPoints(puntos int)
  (StoryPoints, error)`, que valida contra la escala, y la función
  `EscalaFibonacci() []int`, que devuelve una copia de la escala (para que la
  pantalla de Planning Poker y esta validación usen la misma fuente).
- **Sin migración nueva.** US-005 ya guarda los Story Points como valor nulo
  cuando no hay estimación; esta historia solo escribe en esa columna. Si la
  columna tuviera una restricción incompatible con la escala, se corrige en
  una migración nueva, nunca editando una ya mergeada.
- **Dependencia con US-005.** La implementación parte de US-005 mergeada: usa
  `ItemBacklog`, `Estado` y `StoryPoints`. La spec se puede mergear antes.
- **No cambia la escala de US-025.** `CalcularResumenSprint` sigue aceptando
  cualquier entero no negativo (no valida Fibonacci); la escala se impone acá,
  al estimar.
- **Fuera de alcance.** La pantalla de estimación y el Planning Poker (depende
  de T-004 y de la historia de sesiones de poker), la Definition of Ready (se
  valida en US-009), listar ítems sin estimar (US-007) y el cambio de estado a
  completado (US-010).

## 6. Casos límite

- **CL-013-1** — `1` y `13`, los extremos de la escala: se aceptan.
- **CL-013-2** — `0`: se rechaza con `ErrEstimacionFueraDeEscala`. No es una
  forma de "borrar" la estimación.
- **CL-013-3** — `-3`: se rechaza con `ErrEstimacionFueraDeEscala`.
- **CL-013-4** — `4`, `6`, `7`, `9`, `10`, `12`: números que no son Fibonacci,
  se rechazan.
- **CL-013-5** — `21`: es Fibonacci pero está fuera de la escala del proyecto,
  se rechaza. Un ítem de más de 13 puntos es demasiado grande y se divide.
- **CL-013-6** — Re-estimar un ítem ya estimado (de `3` a `5`): se acepta y
  queda `5`.
- **CL-013-7** — Estimar un ítem `en_progreso`: se acepta.
- **CL-013-8** — Estimar un ítem `completado`: se rechaza con
  `ErrItemCompletado`, aunque el valor sea válido.
- **CL-013-9** — Ítem `completado` y valor `4`: se rechaza solo con
  `ErrItemCompletado` (RN-013-5).
- **CL-013-10** — Estimar con el mismo valor que ya tenía: se acepta, sin
  cambios.
- **CL-013-11** — `Estimar` no modifica el ítem que recibe: devuelve una copia.
  Si la estimación falla, el ítem original queda intacto.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) |
|---|---|---|
| El valor no está en {1, 2, 3, 5, 8, 13} | `backlog.ErrEstimacionFueraDeEscala`, envuelto indicando el valor recibido | "la estimacion tiene que ser 1, 2, 3, 5, 8 o 13" |
| El ítem está completado | `backlog.ErrItemCompletado` | "un item completado no se puede estimar" |
| El ítem no existe | `app.ErrItemInexistente`, envuelto indicando el ID recibido | "el item no existe" |

Los dos primeros son del dominio, sin unirse entre sí (RN-013-5: solo se
devuelve uno). El último lo detecta la capa de aplicación antes de llamar al
dominio. Todos son valores de paquete comparables con `errors.Is`, y en ningún
caso se modifica el ítem. `app.ErrItemInexistente` lo define quien llegue
primero a `main` (esta historia o US-019, que también lo usa); la otra lo
reutiliza.

## 8. Criterios de aceptación

- **CA-013-1** (normal) — Un ítem pendiente y sin estimar, estimado con `5`,
  queda con 5 Story Points.
- **CA-013-2** (alternativo) — Un ítem ya estimado en `3` y re-estimado con `8`
  queda con 8 Story Points.
- **CA-013-3** (límite) — `1` y `13` se aceptan; `0` y `21` se rechazan.
- **CA-013-4** (error) — Un valor fuera de la escala (`4`) se rechaza y el
  ítem conserva su estimación anterior.
- **CA-013-5** (error) — Un ítem completado no se puede estimar.
- **CA-013-6** (error) — Un ítem que no existe se rechaza y no se modifica
  nada.

---

## Decisiones tomadas y descartadas

- **La escala es 1, 2, 3, 5, 8, 13.** Es la que ya asume la spec de US-025.
  Se descartó incluir `21`: un ítem que no entra en 13 puntos es demasiado
  grande para un sprint de dos semanas y la práctica es dividirlo, no
  estimarlo más alto.
- **`0` no es una estimación válida.** "Sin estimar" y "estimado en 0" son
  cosas distintas desde US-005; permitir `0` dejaría un ítem que parece
  estimado pero no aporta nada al sprint, y rompería la Definition of Ready
  sin que nadie lo note.
- **No se puede volver a "sin estimar".** Una vez estimado, el ítem se
  re-estima pero no se "desestima". Si el equipo necesita lo contrario, es una
  historia nueva.
- **Se puede re-estimar mientras el ítem no esté completado.** Planning Poker
  es iterativo y el equipo aprende durante el sprint. Se descartó bloquear la
  re-estimación al entrar a un sprint porque nadie lo pidió y complicaría la
  historia con una dependencia de US-009. Un ítem completado sí se bloquea: su
  estimación ya forma parte de una velocidad medida.
- **La escala vive en el dominio de `backlog`, no en `metricas`.** Es una regla
  de estimación, no de cálculo. `metricas` sigue aceptando cualquier entero no
  negativo.
- **Última escritura gana.** No hay bloqueos ni versionado: dos estimaciones
  simultáneas del mismo ítem son un caso raro y el resultado (la última)
  es el esperable.
- **El escenario de ítem completado depende de US-010.** El paso `Dado la
  historia "X" completada` pertenece al área Sprints y no está conectado hasta
  que US-010 llegue a `main`. Hasta entonces CA-013-5 y CL-013-8/9 se cubren con
  tests unitarios del dominio y el escenario BDD se agrega cuando ese paso
  exista.
- **Frases nuevas del diccionario** (a aprobar en el PR de implementación):
  `la historia "X" está sin estimar`. El resto usa frases que ya existen
  (`se estima la historia "X" en N story points`,
  `la historia "X" tiene N story points`).

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia, usando la spec de US-005 como modelo; las decisiones sobre la
  escala y la re-estimación se proponen acá y se confirman en la revisión
- Revisado y entendido por: Tomás Romero
