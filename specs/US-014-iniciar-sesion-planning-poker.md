# US-014 — Iniciar una sesión de Planning Poker sobre una historia

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #25 |
| **Escenarios BDD** | `features/US-014-iniciar-sesion-planning-poker.feature` |
| **Código** | `internal/domain/planningpoker/` (`SesionPlanningPoker`, `NuevaSesion`), caso de uso `IniciarSesionPlanningPoker` en `internal/app/`, `RepositorioPlanningPoker` en `internal/adapters/postgres/` y la base en memoria, migración `migraciones/00009_sesiones_planning_poker.sql` |
| **Última actualización** | 2026-10-23 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero iniciar una sesión de Planning Poker sobre
> una historia del backlog, para que el equipo la estime votando.

Abrir la votación de una historia con todos los integrantes del proyecto como
participantes. Es la primera de las historias de Planning Poker: define
`SesionPlanningPoker`, sus estados y sus rondas, que usan US-015 (votar),
US-016 (revelar), US-017 (nueva ronda) y US-018 (estimación acordada).

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. Viene de la URL; lo verifica el caso de uso, nunca el dominio |
| `itemID` | UUID | Sí | Una historia existente **de ese proyecto**. Viene de la URL; la verifica el caso de uso |

No son entradas los participantes (son los integrantes del proyecto al iniciar,
RN-014-4), el estado ni la ronda (RN-014-6).

Iniciar pasa por las capas en este orden:

1. **Aplicación:** verifica con lecturas simples que el proyecto exista
   (`ErrProyectoInexistente`) y que la historia exista dentro de ese proyecto
   (`ErrItemInexistente`), en ese orden y sin validar nada más (RN-014-1).
2. **Dominio** (`internal/domain/planningpoker`): una función pura recibe datos
   mínimos de la historia (si está completada y si ya tiene una sesión abierta)
   y la lista de participantes, valida todo y devuelve la sesión o todos los
   errores. Usa tipos propios del paquete, como `sprint` en US-009.
3. **Persistencia:** registra la sesión y sus participantes en una transacción,
   con la historia bloqueada (RN-014-11).

## 3. Salidas esperadas

La sesión recién iniciada:

```
SesionPlanningPoker {
    ID            string   // lo genera la base (RN-014-12)
    ItemID        string   // la historia que se estima
    Estado        Estado   // votando
    Ronda         int      // 1
    Participantes []string // IDs de los integrantes, en orden de alta en el proyecto
}
```

Si hay errores, no se devuelve sesión y no se registra nada.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-014-1** (aplicación) — El proyecto y la historia tienen que existir, y
  la historia tiene que ser de ese proyecto. Se verifican en el orden de la
  sección 2, con prioridad y sin unirse con los errores de validación (mismo
  criterio que RN-005-9 y RN-019-6). Una historia de otro proyecto o un ID mal
  formado son una historia inexistente (CL-019-15).
- **RN-014-2** (dominio) — La historia no puede estar `completado`: si lo
  está, `backlog.ErrItemCompletado`, el de US-013, por la misma razón que
  RN-013-4 y RN-006-5.
- **RN-014-3** (dominio) — La historia puede estar ya estimada: re-estimar está
  permitido (RN-013-3). Puede estar `pendiente` o `en_progreso`, y en el
  backlog o en un Sprint.
- **RN-014-4** (aplicación y persistencia) — Participan todos los integrantes
  del proyecto **en el momento de iniciar**, y esa lista se guarda con la
  sesión. Quien se registre en el proyecto después (US-003) no participa de esa
  sesión.
- **RN-014-5** (dominio) — Hacen falta al menos 2 participantes: con uno no hay
  nada que comparar. Si no, `ErrParticipantesInsuficientes`.
- **RN-014-6** (dominio) — La sesión se crea en estado `votando` y en la
  ronda 1.
- **RN-014-7** (dominio) — Una historia tiene como máximo una sesión abierta a
  la vez. Una sesión está **abierta** mientras no está `finalizada` (está
  `votando` o `revelada`). Si la historia ya tiene una, `ErrSesionAbierta`.
  Haber tenido sesiones finalizadas no impide iniciar otra.
- **RN-014-8** (dominio) — Los estados de una sesión son `votando`, `revelada`
  y `finalizada`, y la ronda es un entero que empieza en 1. Esta historia solo
  crea; las transiciones son de otras historias y se nombran acá para que no
  renombren nada:

  | Estado | Qué significa | Lo pone |
  |---|---|---|
  | `votando` | Los participantes votan; los votos están ocultos | US-014 al iniciar; US-017 al abrir una ronda nueva |
  | `revelada` | Se cerró la votación de la ronda y los votos se ven | US-016 |
  | `finalizada` | Se registró la estimación acordada; la sesión no cambia más | US-018 |

  US-017 suma 1 a la ronda al volver a `votando`.
- **RN-014-9** (dominio) — La escala de la sesión es la de US-013
  (`backlog.EscalaFibonacci()`, RN-013-1): no se define otra ni se copia.
  Esta historia no la usa; la usan US-015 al votar y US-018 al registrar el
  valor final.
- **RN-014-10** (dominio) — El dominio valida todo y devuelve todos los errores
  unidos con `errors.Join` (mismo criterio que RN-009-8), en este orden fijo:
  historia completada, sesión abierta, participantes insuficientes.
- **RN-014-11** (persistencia) — El repositorio registra la sesión en una
  transacción `READ COMMITTED`, con estos pasos en orden, y se detiene en el
  primer error:
  1. Bloquea la fila de la historia con `FOR NO KEY UPDATE` y lee su proyecto
     y su estado. Si no hay fila o es de otro proyecto, `ErrItemInexistente`.
  2. Lee si la historia tiene una sesión abierta y la lista de integrantes del
     proyecto, en sentencias posteriores al candado.
  3. Vuelve a verificar RN-014-2, RN-014-5 y RN-014-7 con la misma función del
     dominio y esos datos.
  4. Inserta la sesión y una fila por participante.

  El candado serializa dos inicios simultáneos sobre la misma historia: el
  segundo espera, ve la sesión del primero y se rechaza con `ErrSesionAbierta`
  (CL-014-8). También ordena el inicio con el marcado de US-010, que actualiza
  esa fila: si la historia se completa antes, el inicio la ve completada. El
  índice único parcial de RN-014-13 queda como red de seguridad.
- **RN-014-12** (persistencia) — El ID de la sesión lo genera la base, como
  RN-005-15.
- **RN-014-13** (persistencia) — La migración `00009` crea la tabla de
  sesiones (historia, estado con `CHECK` en los tres valores y `votando` por
  defecto, ronda con `CHECK (ronda >= 1)` y 1 por defecto, y la fecha de inicio
  con `DEFAULT now()`), con un índice único parcial sobre la historia donde el
  estado no es `finalizada`, y la tabla de participantes con clave
  `(sesion_id, integrante_id)`.
- **RN-014-14** (aplicación) — La operación es de todo o nada.

## 5. Restricciones

- **Dominio puro, errores unidos, 404 con prioridad, IDs por la base:** igual
  que US-005 (sección 5) y US-009 (datos mínimos del dominio).
- **Paquete `internal/domain/planningpoker`.** Planning Poker es un término de
  Scrum y queda en inglés; el resto del nombre, en español (ADR 0003): el tipo
  es `SesionPlanningPoker`. Fuera de la biblioteca estándar, el paquete toma
  solo `backlog.ErrItemCompletado` y `backlog.EscalaFibonacci`, como `sprint`
  toma `ErrItemCompletado` (RN-009-5).
- **Migración `00009`.** El plan de trabajo no reserva números para Planning
  Poker; `00009` es el siguiente libre después de la `00008` de US-010. Se
  avisa en el PR para reservarlo.
- **Catálogo de errores (T-005).** Los dos errores nuevos se registran en
  `internal/adapters/http/catalogo.go` en el PR de implementación
  (RN-T005-19), como `conflicto`: dependen de lo que hay guardado (RN-T005-2).
  El test de RN-T005-19 leía una lista fija de paquetes de dominio y no
  incluía `planningpoker`; en el PR de implementación pasa a leer todos los
  paquetes de `internal/domain`, para que un paquete nuevo no quede afuera.
- **Frases BDD y área nueva.** No hay área de Planning Poker en
  `features/steps/mundo.go`. Se propone la interfaz `PlanningPoker`, con
  `IniciarSesion(proyecto, titulo string) error` (con el proyecto, para poder
  escribir el rechazo de un proyecto inexistente, como `CerrarSprint` en
  US-011),
  `EstadoDeLaSesion(titulo string) (estado string, ronda int, err error)` y
  `ParticipantesDeLaSesion(titulo string) ([]string, error)`. La sesión se
  nombra por la historia: tiene como máximo una abierta (RN-014-7). Las
  frases, que se aprueban con el grupo antes de implementar:

  | Frase | Expresión |
  |---|---|
  | `Dado una sesión de Planning Poker sobre la historia "X"` | `^una sesión de Planning Poker sobre la historia "([^"]*)"$` |
  | `Cuando se inicia una sesión de Planning Poker sobre la historia "X"` | `^se inicia una sesión de Planning Poker sobre la historia "([^"]*)"$` |
  | `Entonces la sesión de la historia "X" está "votando" en la ronda 1` | `^la sesión de la historia "([^"]*)" está "([^"]*)" en la ronda (\d+)$` |
  | `Y los participantes de la sesión de la historia "X" son:` + tabla de nombres | `^los participantes de la sesión de la historia "([^"]*)" son:$` |

  Los integrantes se arman con `el integrante "X" con rol "Y"` (US-003), y la
  historia completada, con `la historia "X" completada` (US-010).
- **Fuera de alcance.** Votar (US-015), revelar (US-016), abrir rondas
  (US-017), finalizar (US-018), elegir participantes, cancelar una sesión y la
  pantalla de votación (tarea propia del Sprint 5).

## 6. Casos límite

- **CL-014-1** — Proyecto con exactamente 2 integrantes: se inicia
  (RN-014-5).
- **CL-014-2** — Proyecto con 1 integrante o sin integrantes:
  `ErrParticipantesInsuficientes`.
- **CL-014-3** — Historia ya estimada: se inicia (RN-014-3).
- **CL-014-4** — Historia `en_progreso` o asignada a un Sprint abierto: se
  inicia.
- **CL-014-5** — Historia con una sesión `revelada`: está abierta, se rechaza
  con `ErrSesionAbierta`.
- **CL-014-6** — Historia con sesiones `finalizada` anteriores: se inicia una
  nueva.
- **CL-014-7** — Integrante registrado después de iniciar: no está en la lista
  de participantes (RN-014-4).
- **CL-014-8** — Dos inicios simultáneos sobre la misma historia: uno termina
  bien y el otro se rechaza con `ErrSesionAbierta`, nunca con un error del
  índice único. Se verifica con un test concurrente determinístico contra un
  Postgres real, como el de US-005.
- **CL-014-9** — Historia completada y, además, proyecto con un integrante:
  `ErrItemCompletado` y `ErrParticipantesInsuficientes` unidos, en el orden de
  RN-014-10.
- **CL-014-10** — Proyecto o historia inexistentes, historia de otro proyecto o
  ID mal formado, aunque además la historia esté completada: solo el error de
  inexistencia (RN-014-1).
- **CL-014-11** — La historia se completa (US-010) entre la lectura del caso de
  uso y la transacción: el repositorio la ve completada con el candado tomado y
  rechaza con `ErrItemCompletado`.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) | Tipo (T-005) |
|---|---|---|---|
| El proyecto no existe | `app.ErrProyectoInexistente`, envuelto indicando el ID | "el proyecto no existe" | `no_encontrado` (ya registrado) |
| La historia no existe o es de otro proyecto | `app.ErrItemInexistente`, envuelto indicando el ID | "el item no existe" | `no_encontrado` (ya registrado) |
| La historia está completada | `backlog.ErrItemCompletado` | "el item esta completado" | `conflicto` (ya registrado) |
| La historia ya tiene una sesión abierta | `planningpoker.ErrSesionAbierta` | "la historia ya tiene una sesion de planning poker abierta" | `conflicto` (**nuevo**) |
| El proyecto tiene menos de 2 integrantes | `planningpoker.ErrParticipantesInsuficientes` | "se necesitan al menos 2 integrantes en el proyecto para una sesion de planning poker" | `conflicto` (**nuevo**) |

Los dos primeros tienen prioridad y no se unen con los demás (RN-014-1). Los
tres últimos son del dominio y se unen en el orden de RN-014-10; el repositorio
puede devolverlos también, con la misma función. Todos se comparan con
`errors.Is`, y en ningún caso se registra nada.

## 8. Criterios de aceptación

- **CA-014-1** (normal) — Al iniciar una sesión sobre una historia pendiente de
  un proyecto con tres integrantes, la sesión queda votando, en la ronda 1, con
  los tres como participantes.
- **CA-014-2** (alternativo) — Una historia ya estimada se puede volver a
  estimar en una sesión.
- **CA-014-3** (alternativo) — Un integrante que se registra después de iniciar
  la sesión no participa de ella.
- **CA-014-4** (límite) — Con dos integrantes la sesión se inicia; con uno se
  rechaza.
- **CA-014-5** (error) — Una historia completada no se pone en sesión.
- **CA-014-6** (error) — Una historia que ya tiene una sesión abierta no admite
  otra.
- **CA-014-7** (error) — Un proyecto o una historia inexistentes se rechazan y
  no se registra nada.

---

## Decisiones tomadas y descartadas

- **Participan todos los integrantes del proyecto al iniciar.** La lista se
  guarda con la sesión, así una sesión no cambia de participantes a mitad de la
  votación. Se descartó elegir participantes: suma una pantalla y reglas
  (mínimos, quién elige) que ningún criterio pide.
- **Mínimo dos participantes.** Con uno no hay votos que comparar (US-016).
- **Una sola sesión abierta por historia.** Dos sesiones abiertas sobre la
  misma historia darían dos estimaciones en conflicto (US-018). El candado
  sobre la historia lo garantiza; el índice único parcial es la red de
  seguridad.
- **Los estados y la ronda se definen acá.** US-016, US-017 y US-018 cambian el
  estado y la ronda, pero el modelo es uno solo: definirlo en la primera
  historia evita renombres y migraciones de corrección.
- **Se reutiliza la escala de US-013.** La estimación que sale de una sesión
  termina en los Story Points de la historia (US-018), que solo aceptan esa
  escala. Se descartó una escala propia del Planning Poker (con `?` o `∞`).
- **Se reutiliza `ErrItemCompletado`.** Misma condición que US-013, US-006 y
  US-009: el ítem está completado.
- **Una historia de otro proyecto es inexistente.** Mismo criterio que
  CL-019-15: la URL apunta a una historia que no está en ese proyecto.
- **Candado `FOR NO KEY UPDATE` sobre la historia.** Ordena los inicios sobre
  la misma historia y el inicio con el marcado de US-010, sin bloquear a quien
  inserta filas que apuntan a la historia, como un registro de esfuerzo
  (mismo criterio que RN-009-11).

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia; se listan para no agregar reglas por
cuenta propia:

1. **Cancelar una sesión.** Ninguna historia lo pide. Una sesión abandonada
   queda abierta y bloquea nuevas sesiones de esa historia (RN-014-7) hasta que
   US-018 la finalice.
2. **La historia se completa con una sesión abierta.** US-010 no mira las
   sesiones. Al finalizar, US-018 no podrá registrar la estimación
   (RN-013-4), y la sesión seguirá abierta.
3. **Quién inicia la sesión.** No hay usuarios en el sistema: no se registra
   quién la inicia.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia y de las decisiones del equipo (participantes al iniciar, mínimo
  de dos, una sesión abierta por historia, estados y rondas, escala de US-013),
  usando las specs de US-005, US-009, US-010, US-011, US-013, US-019 y T-005
  como modelo
- [x] Use IA para: el ajuste tras la implementación (nombres del código,
  `IniciarSesion` con el proyecto, el test del catálogo sobre todo
  `internal/domain`), redactado con Claude Code
- Revisado y entendido por: Juan Ignacio Vergara
