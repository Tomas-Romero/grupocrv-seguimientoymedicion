# US-015 — Registrar el voto individual manteniéndolo oculto hasta el cierre

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #26 |
| **Escenarios BDD** | `features/US-015-voto-individual-oculto.feature` |
| **Código** | `internal/domain/planningpoker/` (votación de una ronda), caso de uso en `internal/app/`, repositorio en `internal/adapters/postgres/`, migración `migraciones/00010_votos_planning_poker.sql` |
| **Última actualización** | 2026-10-21 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero registrar mi voto en una sesión de
> Planning Poker sin que los demás lo vean, para estimar sin dejarme influir
> por los votos de los otros.

Cada participante elige una carta de la escala. Mientras la sesión está
votando, el sistema muestra quién votó y quién falta, pero nunca los valores:
eso lo garantiza el dominio, no la pantalla. Los valores se ven recién al
revelar (US-016).

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `sesionID` | UUID | Sí | Una sesión existente (US-014). Viene de la URL; la verifica el caso de uso |
| `integranteID` | UUID | Sí | Un integrante del proyecto de la sesión. Viene de la selección del formulario; lo verifica el caso de uso |
| `valor` | entero | Sí | Un valor de la escala de US-013: 1, 2, 3, 5, 8 o 13 (RN-013-1) |

La ronda no es una entrada: el voto va a la ronda actual de la sesión
(RN-015-6).

Votar pasa por las capas en este orden:

1. **Aplicación:** verifica con lecturas simples que la sesión exista
   (`ErrSesionInexistente`) y que el integrante exista y sea del proyecto de la
   sesión (`ErrIntegranteInexistente`), en ese orden y sin validar nada más
   (RN-015-1).
2. **Dominio** (`internal/domain/planningpoker`): valida, con el estado y los
   participantes de la sesión, que el integrante participe, que la sesión esté
   `votando` y que el valor sea de la escala, y devuelve todos los errores.
3. **Persistencia:** guarda el voto en una transacción con la sesión bloqueada
   (RN-015-9).

## 3. Salidas esperadas

El caso de uso devuelve la **votación de la ronda** como la ve cualquiera
mientras la sesión está `votando`: quién votó y quién falta, sin valores.

| Lo que expone la votación | Con la sesión `votando` | Con la sesión `revelada` o `finalizada` |
|---|---|---|
| Quién votó (IDs de participantes) | Sí | Sí |
| Quién falta (IDs de participantes) | Sí | Sí |
| El valor de cada voto | **No**: `ErrVotosOcultos` | Sí (lo usa US-016) |

Si hay errores, no se guarda nada y no se devuelve votación.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-015-1** (aplicación) — La sesión y el integrante tienen que existir, y
  el integrante tiene que ser del proyecto de la sesión. Se verifican en el
  orden de la sección 2, con prioridad y sin unirse con los errores de
  validación (mismo criterio que RN-014-1 y RN-019-6). Un integrante de otro
  proyecto o un ID mal formado son inexistentes (CL-019-15).
- **RN-015-2** (dominio) — Solo votan los participantes de la sesión
  (RN-014-4). Un integrante del proyecto que no participa (por ejemplo, porque
  se registró después de iniciar) se rechaza con `ErrNoEsParticipante`.
- **RN-015-3** (dominio) — Solo se vota con la sesión en `votando`
  (RN-014-8). Con la sesión `revelada` o `finalizada`,
  `ErrVotacionCerrada`.
- **RN-015-4** (dominio) — El valor es de la escala de US-013: se valida con
  `backlog.NuevosStoryPoints` (RN-014-9), sin copiar la escala. Si no
  pertenece, `backlog.ErrEstimacionFueraDeEscala`, con su mensaje.
- **RN-015-5** (dominio) — Mientras la sesión esté `votando`, un participante
  puede cambiar su voto: cuenta el último. Votar otra vez no es un error.
- **RN-015-6** (dominio y persistencia) — Los votos se guardan por ronda: hay
  como máximo un voto por `(sesión, ronda, integrante)`, y el voto va siempre a
  la ronda actual. Una ronda nueva (US-017) empieza sin votos y no cambia los
  de las rondas anteriores.
- **RN-015-7** (dominio) — **Voto oculto.** La votación de una ronda es un
  tipo del dominio que no exporta sus valores. Expone `Votaron()` y `Faltan()`
  (IDs de participantes, en el orden de la lista de la sesión), y `Valores()`,
  que con la sesión `votando` devuelve `ErrVotosOcultos` y ningún valor. Así,
  un error en la pantalla o en un adaptador no puede mostrar votos antes de
  tiempo: el dato no sale del dominio.
- **RN-015-8** (dominio) — El dominio valida todo y devuelve todos los errores
  unidos con `errors.Join` (RN-014-10), en este orden fijo: votación cerrada,
  no es participante, valor fuera de la escala.
- **RN-015-9** (persistencia) — El repositorio guarda el voto en una
  transacción `READ COMMITTED`, con estos pasos en orden, y se detiene en el
  primer error:
  1. Bloquea la fila de la sesión con `FOR NO KEY UPDATE` y lee su estado y su
     ronda. Si no hay fila, `ErrSesionInexistente`.
  2. Lee los participantes, en una sentencia posterior al candado.
  3. Vuelve a verificar RN-015-2 a RN-015-4 con la misma función del dominio y
     esos datos.
  4. Inserta el voto de la ronda actual, o reemplaza el valor si el
     participante ya había votado en ella
     (`INSERT ... ON CONFLICT (sesion_id, ronda, integrante_id) DO UPDATE`).
- **RN-015-10** (persistencia) — El candado del paso 1 ordena lo que pasa a la
  vez sobre la sesión. Dos votos simultáneos del mismo participante: el
  segundo espera, y cuenta uno solo, el último que confirma (CL-015-8). Un voto
  y el revelado de US-016, que cambia el estado de esa fila: si el voto llega
  antes, cuenta; si llega después, ve la sesión `revelada` y se rechaza con
  `ErrVotacionCerrada` (CL-015-9).
- **RN-015-11** (persistencia) — La migración `00010` crea la tabla de votos:
  sesión, ronda (`CHECK (ronda >= 1)`), integrante, valor (`NOT NULL`, sin
  repetir la escala en la base) y la fecha del voto con `DEFAULT now()`. La
  clave es `(sesion_id, ronda, integrante_id)`, y `(sesion_id, integrante_id)`
  referencia a la tabla de participantes de US-014: es la red de seguridad de
  RN-015-2.
- **RN-015-12** (aplicación) — La operación es de todo o nada.

## 5. Restricciones

- **Dominio puro, errores unidos, 404 con prioridad:** igual que US-005
  (sección 5) y US-014 (sección 5).
- **Depende de US-014.** Usa `SesionPlanningPoker`, sus estados, su ronda y la
  lista de participantes (RN-014-4, RN-014-8 y RN-014-13). Esta spec cita la
  de US-014, que se revisa en paralelo; conviene mergear la de US-014 primero.
- **Dependencia con US-016: el cierre es manual.** La votación de una ronda se
  cierra solo cuando alguien revela (US-016), que pasa la sesión a `revelada`
  con el mismo candado `FOR NO KEY UPDATE` sobre la sesión (RN-015-10). Desde
  ahí `Valores()` devuelve los votos de la ronda. Esta historia no cierra la
  votación aunque hayan votado todos.
- **Migración `00010`.** El plan de trabajo no reserva números para Planning
  Poker; `00010` es la siguiente a la `00009` de US-014. Se avisa en el PR
  para reservarla.
- **Catálogo de errores (T-005).** En el PR de implementación se registran en
  `internal/adapters/http/catalogo.go` (RN-T005-19):
  `app.ErrSesionInexistente` como `no_encontrado`, y
  `planningpoker.ErrVotacionCerrada` y `planningpoker.ErrNoEsParticipante` como
  `conflicto` (RN-T005-2). `planningpoker.ErrVotosOcultos` va en las
  exclusiones, con el motivo: solo aparece si un adaptador pide los valores
  antes de tiempo, que es un error de programación y tiene que salir como
  error interno. `backlog.ErrEstimacionFueraDeEscala` ya está registrado como
  `validacion` del campo `estimacion`, así que el campo del voto en el
  formulario se llama `estimacion`.
- **Frases BDD.** Suma a la interfaz `PlanningPoker` de US-014
  `Votar(integrante, titulo string, valor int) error`,
  `QuienesVotaron(titulo string) ([]string, error)`,
  `QuienesFaltan(titulo string) ([]string, error)` y
  `VotosOcultos(titulo string) (bool, error)`. Las frases, que se aprueban con
  el grupo antes de implementar:

  | Frase | Expresión |
  |---|---|
  | `Dado el integrante "X" ya votó 5 en la sesión de la historia "H"` | `^el integrante "([^"]*)" ya votó (-?\d+) en la sesión de la historia "([^"]*)"$` |
  | `Cuando el integrante "X" vota 5 en la sesión de la historia "H"` | `^el integrante "([^"]*)" vota (-?\d+) en la sesión de la historia "([^"]*)"$` |
  | `Entonces en la sesión de la historia "H" votaron:` + tabla de nombres | `^en la sesión de la historia "([^"]*)" votaron:$` |
  | `Y en la sesión de la historia "H" faltan votar:` + tabla de nombres | `^en la sesión de la historia "([^"]*)" faltan votar:$` |
  | `Y los votos de la sesión de la historia "H" están ocultos` | `^los votos de la sesión de la historia "([^"]*)" están ocultos$` |

  El número acepta negativos para poder escribir el rechazo de un valor fuera
  de la escala. Los escenarios no comprueban qué valor quedó guardado: con la
  sesión votando no se puede ver (RN-015-7). Eso lo cubren los tests del
  repositorio, y los escenarios de US-016.
- **Fuera de alcance.** Revelar y detectar divergencias (US-016), abrir una
  ronda nueva (US-017), la estimación final (US-018) y la pantalla de votación
  (tarea propia del Sprint 5).

## 6. Casos límite

- **CL-015-1** — Votos de 1 y de 13: se aceptan (RN-013-1).
- **CL-015-2** — Votos de 0, 4 y 21: se rechazan con
  `ErrEstimacionFueraDeEscala` (RN-013-2).
- **CL-015-3** — El participante vota dos veces en la misma ronda: queda un
  solo voto, el último (RN-015-5), y figura una sola vez en `Votaron()`.
- **CL-015-4** — Nadie votó todavía: `Votaron()` vacío y `Faltan()` con todos
  los participantes.
- **CL-015-5** — Votaron todos: la sesión sigue `votando` (el cierre es manual)
  y los valores siguen ocultos.
- **CL-015-6** — Un integrante del proyecto registrado después de iniciar la
  sesión: `ErrNoEsParticipante`.
- **CL-015-7** — Sesión `revelada` y, además, un valor fuera de la escala:
  `ErrVotacionCerrada` y `ErrEstimacionFueraDeEscala` unidos, en el orden de
  RN-015-8.
- **CL-015-8** — Dos votos simultáneos del mismo participante: los dos
  terminan bien y queda el valor del último que confirma. Se verifica con un
  test concurrente determinístico contra un Postgres real, como el de US-005.
- **CL-015-9** — La sesión se revela entre la lectura del caso de uso y la
  transacción del voto: el repositorio la ve `revelada` con el candado tomado y
  rechaza con `ErrVotacionCerrada`. Se verifica llamando al repositorio con una
  sesión ya revelada en la base.
- **CL-015-10** — Sesión o integrante inexistentes, integrante de otro proyecto
  o IDs mal formados, aunque además el valor sea inválido: solo el error de
  inexistencia (RN-015-1).
- **CL-015-11** — Pedir los valores con la sesión `votando`:
  `ErrVotosOcultos` y ningún valor, aunque hayan votado todos.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) | Tipo (T-005) |
|---|---|---|---|
| La sesión no existe | `app.ErrSesionInexistente`, envuelto indicando el ID | "la sesion de planning poker no existe" | `no_encontrado` (**nuevo**) |
| El integrante no existe o es de otro proyecto | `app.ErrIntegranteInexistente` (el de US-019), envuelto indicando el ID | "el integrante no existe" | `no_encontrado` (ya registrado) |
| La sesión no está votando | `planningpoker.ErrVotacionCerrada` | "la votacion de esta ronda esta cerrada" | `conflicto` (**nuevo**) |
| El integrante no participa de la sesión | `planningpoker.ErrNoEsParticipante` | "el integrante no participa de esta sesion de planning poker" | `conflicto` (**nuevo**) |
| El valor no es de la escala | `backlog.ErrEstimacionFueraDeEscala`, envuelto indicando el valor | "la estimacion tiene que ser 1, 2, 3, 5, 8 o 13" | `validacion`, campo `estimacion` (ya registrado) |
| Se piden los valores con la sesión votando | `planningpoker.ErrVotosOcultos` | "los votos estan ocultos hasta que se revelen" | exclusión del catálogo: error interno (**nuevo**) |

Los dos primeros tienen prioridad y no se unen con los demás (RN-015-1). Los
tres siguientes son del dominio y se unen en el orden de RN-015-8; el
repositorio puede devolverlos también, con la misma función. Todos se comparan
con `errors.Is`, y en ningún caso se guarda nada.

## 8. Criterios de aceptación

- **CA-015-1** (normal) — Un participante vota con la sesión votando: queda
  registrado que votó, sin que se vea su valor.
- **CA-015-2** (alternativo) — Un participante cambia su voto mientras la
  sesión está votando: cuenta el último y figura una sola vez entre los que
  votaron.
- **CA-015-3** (alternativo) — Mientras la sesión está votando se ve quién votó
  y quién falta, y nunca los valores, aunque hayan votado todos.
- **CA-015-4** (límite) — Los votos de 1 y de 13 se aceptan; los de 0, 4 y 21
  se rechazan.
- **CA-015-5** (error) — Un integrante que no participa de la sesión no vota.
- **CA-015-6** (error) — Con la sesión revelada o finalizada no se vota.
- **CA-015-7** (error) — Una sesión o un integrante inexistentes se rechazan y
  no se guarda nada.

---

## Decisiones tomadas y descartadas

- **El voto oculto lo garantiza el dominio.** La votación no exporta sus
  valores y `Valores()` se niega mientras la sesión está votando. Se descartó
  ocultarlos solo en la pantalla: un error en la interfaz o en el adaptador
  mostraría votos antes de tiempo.
- **Se puede cambiar el voto mientras se vota.** Los errores al elegir una
  carta pasan, y hasta que se revela nadie vio nada. Cuenta el último.
- **Los votos se guardan por ronda.** La clave `(sesión, ronda, integrante)`
  deja lista la ronda nueva de US-017 sin cambiar el modelo ni borrar votos.
- **El cierre es manual, al revelar (US-016).** Se descartó cerrar solo cuando
  votaron todos: si alguien no vota, la sesión se traba.
- **Candado `FOR NO KEY UPDATE` sobre la sesión.** Ordena los votos del mismo
  participante y el voto con el revelado, sin bloquear a quien inserta filas
  que apuntan a la sesión (mismo criterio que RN-010-9 sobre el Sprint).
- **Se reutilizan la escala y `ErrEstimacionFueraDeEscala` de US-013.** El
  valor que se acuerde termina en los Story Points de la historia (US-018).
- **`ErrVotosOcultos` es un error interno.** No lo provoca nada que la persona
  escriba: solo un adaptador que pide los valores antes de tiempo. Por eso va
  en las exclusiones del catálogo y no como `conflicto`.

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia; se listan para no agregar reglas por
cuenta propia:

1. **Quién vota por quién.** No hay usuarios en el sistema: el integrante llega
   desde el formulario, así que cualquiera podría mandar el voto de otro
   participante.
2. **La persona no ve su propia carta.** Con la sesión votando el dominio no
   expone ningún valor, tampoco el propio. La pantalla puede mostrar "ya
   votaste", no "votaste 5".
3. **Sin carta de abstención.** No hay "?" ni "pasar": un participante que no
   quiere votar no vota, y figura en `Faltan()`.
4. **Votos de rondas anteriores.** Quedan guardados (RN-015-6); si se muestran
   o no lo decide US-017.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia y de las decisiones del equipo (solo votan los participantes,
  cuenta el último voto, votos por ronda, voto oculto garantizado por el
  dominio, cierre manual al revelar, concurrencia), usando las specs de US-013,
  US-014, US-010, US-019 y T-005 como modelo
- Revisado y entendido por:
