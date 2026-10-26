# US-018 — Registrar la estimación acordada como valor final de la historia

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #29 |
| **Escenarios BDD** | `features/US-018-registrar-estimacion-acordada.feature` |
| **Código** | `internal/domain/planningpoker/` (`Acordar`, `SesionPlanningPoker.ValorAcordado`), caso de uso `RegistrarEstimacionAcordada` en `internal/app/`, `GuardarAcuerdo` en `internal/adapters/postgres/` y `internal/adapters/memoria/`, migración `migraciones/00011_sesiones_valor_acordado.sql` |
| **Última actualización** | 2026-10-26 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero registrar la estimación que acordamos en la
> sesión de Planning Poker, para que quede como el valor final de la historia.

Cerrar la sesión con el valor que el equipo acordó después de ver los votos. En
una sola operación, la sesión pasa a `finalizada` guardando ese valor, y la
historia queda estimada con él, igual que si se la estimara con US-013. Es la
última historia de Planning Poker: usa la sesión de US-014, los votos de US-015
y la revelación de US-016.

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `sesionID` | UUID | Sí | Una sesión existente (US-014). Viene de la URL; la verifica el caso de uso |
| `integranteID` | UUID | Sí | Un integrante del proyecto de la sesión. Es quien registra. Viene de la selección del formulario; lo verifica el caso de uso |
| `valor` | entero | Sí | Un valor de la escala de US-013: 1, 2, 3, 5, 8 o 13 (RN-013-1) |

Registrar pasa por las capas en este orden:

1. **Aplicación:** verifica con lecturas simples que la sesión exista
   (`ErrSesionInexistente`) y que el integrante exista y sea del proyecto de la
   sesión (`ErrIntegranteInexistente`), en ese orden y sin validar nada más
   (RN-018-1), con las mismas lecturas que votar y revelar (RN-015-1 y
   RN-016-1). Después lee la historia de la sesión, para saber si está
   completada.
2. **Dominio** (`internal/domain/planningpoker`): una función pura recibe la
   sesión, si la historia está completada, quién registra y el valor; valida
   todo y devuelve los Story Points acordados o los errores.
3. **Persistencia:** en una sola transacción, con la sesión bloqueada, finaliza
   la sesión con el valor y estima la historia (RN-018-9).

## 3. Salidas esperadas

| Qué | Después de registrar |
|---|---|
| Estado de la sesión | `finalizada` |
| Valor acordado de la sesión | El valor registrado (`ValorAcordado`, nuevo en `SesionPlanningPoker`) |
| Story Points de la historia | El valor registrado: reemplaza una estimación anterior, como re-estimar (RN-013-3) |
| Votos de la sesión, de todas las rondas | Sin cambios (RN-015-6) |
| Estado de la historia | Sin cambios |

El caso de uso devuelve la sesión finalizada, con su valor acordado. Si hay un
error, la sesión sigue `revelada`, la historia queda como estaba y no se
devuelve sesión.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-018-1** (aplicación) — La sesión y el integrante tienen que existir, y
  el integrante tiene que ser del proyecto de la sesión. Se verifican en el
  orden de la sección 2, con prioridad y sin unirse con los errores de
  validación (mismo criterio que RN-015-1 y RN-016-1).
- **RN-018-2** (dominio) — Solo se registra con la sesión `revelada`: después
  de ver los votos (US-016) y antes de finalizarla. Con la sesión `votando` o
  ya `finalizada`, `ErrSesionNoRevelada`.
- **RN-018-3** (dominio) — Solo registra un participante de la sesión
  (RN-014-4). Un integrante del proyecto que no participa se rechaza con
  `ErrNoEsParticipante`, el de US-015 (mismo criterio que RN-016-3).
- **RN-018-4** (dominio) — El valor es de la escala de US-013: se valida con
  `backlog.NuevosStoryPoints` (RN-014-9), sin copiar la escala. Si no
  pertenece, `backlog.ErrEstimacionFueraDeEscala`, con su mensaje.
- **RN-018-5** (dominio) — El valor puede ser **cualquiera de la escala**, no
  solo uno de los votados: el equipo puede converger en un punto intermedio.
  Tampoco hace falta que no haya divergencia (RN-016-6): para eso se discute,
  y si no hay acuerdo se abre otra ronda (US-017).
- **RN-018-6** (dominio) — Una historia `completado` no se estima (RN-013-4).
  Se valida primero: si está completada, el único error es
  `backlog.ErrItemCompletado`, aunque haya otros problemas (mismo criterio que
  RN-013-5). La sesión queda `revelada`, sin finalizar.
- **RN-018-7** (dominio) — Fuera del caso de RN-018-6, el dominio valida todo
  y devuelve todos los errores unidos con `errors.Join` (RN-014-10), en este
  orden fijo: sesión no revelada, no es participante, valor fuera de la escala.
- **RN-018-8** (dominio y persistencia) — Registrar hace dos cambios a la vez:
  la sesión pasa a `finalizada` con el valor acordado, y la historia queda
  estimada con ese valor. No hay forma de hacer uno sin el otro.
- **RN-018-9** (persistencia) — El repositorio registra en una transacción
  `READ COMMITTED`, con estos pasos en orden, y se detiene en el primer error.
  Si algo falla, no cambia nada.
  1. Bloquea la fila de la sesión con `FOR NO KEY UPDATE` y lee su estado, su
     historia y sus participantes, como RN-015-9. Si no hay fila,
     `ErrSesionInexistente`.
  2. Lee el estado de la historia, en una sentencia posterior al candado.
  3. Vuelve a verificar RN-018-2 a RN-018-6 con la misma función del dominio y
     esos datos.
  4. Estima la historia con el `UPDATE` condicionado de US-013
     (`... WHERE id = $1 AND estado <> 'completado'`, RN-013-7 y #110). Si no
     actualiza ninguna fila, la historia se completó en el medio:
     `ErrItemCompletado`, y la transacción se revierte.
  5. Pasa la sesión a `finalizada` con el valor acordado.
- **RN-018-10** (persistencia) — El candado del paso 1 ordena lo que pasa a la
  vez sobre la sesión, con el criterio de RN-015-10 y RN-016-9. Dos registros
  simultáneos: el segundo espera, ve la sesión `finalizada` y se rechaza con
  `ErrSesionNoRevelada`. Un registro y una ronda nueva (US-017), que vuelve la
  sesión a `votando` con el mismo candado: el que llega segundo ve el estado
  que dejó el primero. El marcado de US-010 no toma ese candado: lo cubre el
  `UPDATE` condicionado del paso 4 (CL-018-10).
- **RN-018-11** (persistencia) — La migración `00011` agrega a la tabla de
  sesiones la columna del valor acordado (entero, nula), con
  `CHECK ((estado = 'finalizada') = (valor_acordado IS NOT NULL))`: solo una
  sesión finalizada tiene valor, y toda sesión finalizada lo tiene. No repite
  la escala en la base (mismo criterio que RN-015-11).

## 5. Restricciones

- **Dominio puro, errores unidos, 404 con prioridad, READ COMMITTED:** igual
  que US-014 y US-015 (sección 5).
- **Se reutiliza la regla de US-013, no su caso de uso.** El valor se valida
  con `backlog.NuevosStoryPoints` y la historia se estima con el mismo `UPDATE`
  condicionado de `GuardarEstimacion` (#110). El caso de uso `EstimarItem` no se
  llama desde este: no podría compartir la transacción, y la sesión podría
  quedar finalizada con la historia sin estimar. En el PR de implementación, la
  sentencia de `GuardarEstimacion` se extrae a una función que acepta el pool o
  una transacción (como `leerItem` en US-010), y la usan los dos: es
  `guardarEstimacion`. `ValorAcordado` es un `backlog.StoryPoints`, con el
  mismo cero ("sin acordar") que los Story Points de la historia.
- **Depende de US-016 (#162).** Usa el estado `revelada`, que pone el
  revelado. La implementación se abre con US-016 mergeada; esta spec solo cita
  sus nombres.
- **Migración `00011`.** El plan de trabajo dice que US-018 no llevaba
  migración y anota "00011 en adelante" para defectos (US-022 y US-023), con el
  número a reservar al escribir su spec. Guardar el valor acordado necesita una
  columna, así que esta spec toma la `00011`, que ninguna rama usa, y lo avisa
  en el PR: defectos empezaría en la `00012`.
- **Catálogo de errores (T-005).** En el PR de implementación se registra
  `planningpoker.ErrSesionNoRevelada` en `internal/adapters/http/catalogo.go`
  como `conflicto` (RN-T005-2). Los demás ya están registrados.
- **Frases BDD.** Suma a la interfaz `PlanningPoker`
  `RegistrarEstimacionAcordada(integrante, titulo string, valor int) error` y
  `EstimacionAcordada(titulo string) (int, error)`. Las frases nuevas, que se
  aprueban con el grupo antes de implementar:

  | Frase | Expresión |
  |---|---|
  | `Cuando el integrante "X" registra la estimación acordada de 5 en la sesión de la historia "H"` | `^el integrante "([^"]*)" registra la estimación acordada de (-?\d+) en la sesión de la historia "([^"]*)"$` |
  | `Y la estimación acordada de la sesión de la historia "H" es 5` | `^la estimación acordada de la sesión de la historia "([^"]*)" es (\d+)$` |

  Una sesión finalizada ya no está abierta (RN-014-7): para estas frases y para
  `la sesión de la historia "X" está "finalizada" en la ronda N`, el área lee la
  **última** sesión de la historia, no solo la abierta. Los Story Points se
  comprueban con `la historia "H" tiene 5 story points` (US-013), y la sesión
  revelada se arma con `el integrante "X" reveló los votos ...` (US-016). La
  parte de CA-018-5 de una sesión ya finalizada no tiene escenario: no hay frase
  de contexto para una sesión finalizada y el escenario necesitaría dos
  registros. La cubren los tests del dominio, de la aplicación, de memoria y de
  Postgres.
- **Fuera de alcance.** Abrir una ronda nueva (US-017), el reporte de
  estimaciones (lo alimenta el valor guardado en la sesión), cancelar una
  sesión y la pantalla (tarea propia del Sprint 5).

## 6. Casos límite

- **CL-018-1** — Valor igual a todos los votos: se registra.
- **CL-018-2** — Valor que nadie votó (por ejemplo, 5 con votos de 3 y 8): se
  registra (RN-018-5).
- **CL-018-3** — Con divergencia: se registra (RN-018-5).
- **CL-018-4** — Valores 1 y 13: se aceptan. Valores 0, 4 y 21: se rechazan
  con `ErrEstimacionFueraDeEscala` (RN-013-2).
- **CL-018-5** — Historia ya estimada antes de la sesión: queda con el valor
  acordado (RN-013-3).
- **CL-018-6** — Sesión `votando` (todavía no se reveló) o `finalizada` (ya se
  registró): `ErrSesionNoRevelada`.
- **CL-018-7** — Integrante del proyecto que no participa de la sesión:
  `ErrNoEsParticipante`.
- **CL-018-8** — Historia completada con la sesión revelada (hueco 2 de
  US-014): solo `ErrItemCompletado`, aunque además el valor sea inválido
  (RN-018-6). La sesión sigue `revelada`.
- **CL-018-9** — Sesión `votando`, integrante que no participa y valor fuera de
  la escala: los tres errores unidos, en el orden de RN-018-7.
- **CL-018-10** — La historia se completa (US-010) entre la lectura del caso de
  uso y la transacción: el `UPDATE` condicionado no toca filas, se devuelve
  `ErrItemCompletado` y la sesión sigue `revelada`. Se verifica llamando al
  repositorio con la historia ya completada en la base.
- **CL-018-11** — Dos registros simultáneos de la misma sesión: uno termina
  bien y el otro se rechaza con `ErrSesionNoRevelada`; queda el valor del
  primero. Se verifica con un test concurrente determinístico contra un
  Postgres real, como el de US-015.
- **CL-018-12** — Falla la transacción después de estimar la historia y antes
  de finalizar la sesión: no cambia nada (RN-018-9). Se verifica forzando la
  falla, como CL-011-11.
- **CL-018-13** — Sesión o integrante inexistentes, integrante de otro
  proyecto o IDs mal formados: solo el error de inexistencia (RN-018-1).

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) | Tipo (T-005) |
|---|---|---|---|
| La sesión no existe | `app.ErrSesionInexistente`, envuelto indicando el ID | "la sesion de planning poker no existe" | `no_encontrado` (ya registrado) |
| El integrante no existe o es de otro proyecto | `app.ErrIntegranteInexistente`, envuelto indicando el ID | "el integrante no existe" | `no_encontrado` (ya registrado) |
| La historia está completada | `backlog.ErrItemCompletado` | "el item esta completado" | `conflicto` (ya registrado) |
| La sesión no está revelada | `planningpoker.ErrSesionNoRevelada` | "la estimacion acordada se registra con los votos revelados y la sesion sin finalizar" | `conflicto` (**nuevo**) |
| El integrante no participa de la sesión | `planningpoker.ErrNoEsParticipante` | "el integrante no participa de esta sesion de planning poker" | `conflicto` (ya registrado) |
| El valor no es de la escala | `backlog.ErrEstimacionFueraDeEscala`, envuelto indicando el valor | "la estimacion tiene que ser 1, 2, 3, 5, 8 o 13" | `validacion`, campo `estimacion` (ya registrado) |

Los dos primeros tienen prioridad y no se unen con los demás (RN-018-1).
`ErrItemCompletado` va solo (RN-018-6). Los tres últimos se unen en el orden de
RN-018-7. El repositorio puede devolver los del dominio con la misma función,
y `ErrItemCompletado` también por el `UPDATE` condicionado. Todos se comparan
con `errors.Is`, y en ningún caso cambia nada.

## 8. Criterios de aceptación

- **CA-018-1** (normal) — Con la sesión revelada, un participante registra la
  estimación acordada: la sesión queda finalizada con ese valor y la historia
  queda estimada con él.
- **CA-018-2** (alternativo) — Se puede acordar un valor que nadie votó, y
  también con divergencia.
- **CA-018-3** (alternativo) — Una historia ya estimada queda con el valor
  acordado.
- **CA-018-4** (límite) — Los valores 1 y 13 se aceptan; los de 0, 4 y 21 se
  rechazan, y la sesión sigue revelada.
- **CA-018-5** (error) — Con la sesión votando o ya finalizada no se registra.
- **CA-018-6** (error) — Un integrante que no participa de la sesión no
  registra.
- **CA-018-7** (error) — Si la historia está completada, no se registra y la
  sesión sigue revelada.
- **CA-018-8** (error) — Una sesión o un integrante inexistentes se rechazan y
  no se cambia nada.

---

## Decisiones tomadas y descartadas

- **Solo con la sesión revelada.** El acuerdo sale de discutir los votos ya
  vistos. Se descartó registrar con la sesión votando (el equipo no vio los
  votos) o volver a registrar sobre una finalizada (cambiaría hacia atrás el
  reporte de estimaciones).
- **Cualquier valor de la escala, con o sin divergencia.** El equipo puede
  converger en un punto intermedio. Se descartó exigir uno de los votados o que
  no haya divergencia: si no hay acuerdo, el camino es otra ronda (US-017).
- **Una sola transacción para la sesión y la historia.** Si se hicieran por
  separado, una falla en el medio dejaría la sesión finalizada con la historia
  sin estimar, o al revés. El `UPDATE` condicionado de US-013 dentro de la
  transacción hace que una historia completada en el medio revierta todo.
- **Se reutiliza la regla de US-013, no su caso de uso.** Un caso de uso que
  llama a otro no comparte su transacción. Se reutilizan
  `NuevosStoryPoints` y la sentencia de `GuardarEstimacion`.
- **Historia completada: se rechaza y la sesión queda revelada.** Es el hueco 2
  de US-014. No molesta porque US-014 no deja abrir sesiones sobre historias
  completadas (RN-014-2). Se descartó impedir completar historias con una
  sesión abierta: complica US-010 por un caso raro.
- **El valor acordado queda guardado en la sesión.** Lo pide el reporte de
  estimaciones de la cátedra, y la historia sola no alcanza: sus Story Points
  pueden cambiar después con US-013. Por eso la migración.
- **Un solo error para "no revelada".** `ErrSesionNoRevelada` cubre la sesión
  votando y la finalizada, con un mensaje que nombra las dos condiciones. Se
  descartó un error por estado: suma una entrada al catálogo y la persona
  resuelve las dos igual, mirando el estado de la sesión.
- **Registra un participante.** Mismo criterio que RN-015-2 y RN-016-3: una
  sola regla de quién interviene en la sesión.

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia; se listan para no agregar reglas por
cuenta propia:

1. **Quién registra por quién.** Como al votar y revelar, no hay usuarios:
   cualquiera podría registrar en nombre de un participante.
2. **Historia desmarcada después del rechazo.** Si una historia completada se
   desmarca (US-010), vuelve a `pendiente` con su sesión todavía `revelada`, y
   esa sesión se puede finalizar. No se pierde nada, pero el equipo podría
   preferir otra ronda.
3. **Estimar con US-013 durante la sesión.** US-013 no mira las sesiones: se
   puede estimar a mano una historia con una sesión abierta, y al registrar la
   estimación acordada se pisa ese valor (RN-013-3).

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia y de las decisiones del equipo (solo con la sesión revelada,
  cualquier valor de la escala, una sola transacción con la sesión bloqueada,
  la regla de US-013 sin su caso de uso, historia completada rechazada, el
  valor guardado en la sesión), usando las specs de US-013, US-014, US-015,
  US-016 y T-005 como modelo
- [x] Use IA para: el ajuste tras la implementación (nombres del código, el tipo
  de `ValorAcordado` y la cobertura de CA-018-5), redactado con Claude Code
- Revisado y entendido por: Juan Ignacio Vergara
