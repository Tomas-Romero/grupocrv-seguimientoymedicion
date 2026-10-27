# US-017 — Abrir una nueva ronda de votación cuando hay divergencia

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #28 |
| **Escenarios BDD** | `features/US-017-nueva-ronda.feature` |
| **Código** | `internal/domain/planningpoker/`, caso de uso en `internal/app/`, repositorios en `internal/adapters/postgres/` y `internal/adapters/memoria/`. Sin migración nueva |
| **Última actualización** | 2026-10-27 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero abrir una nueva ronda de votación cuando
> los votos divergen, para que el equipo discuta y vuelva a votar.

Es la parte de "realizar nuevas rondas" del requerimiento 4 de la guía. Después
de revelar (US-016), si hay divergencia, la sesión vuelve a `votando` en la
ronda siguiente, con los mismos participantes y sin votos. Los votos de las
rondas anteriores se conservan y se pueden ver, porque ya se revelaron.

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `sesionID` | UUID | Sí | Una sesión existente (US-014). Viene de la URL |
| `integranteID` | UUID | Sí | Un integrante del proyecto de la sesión: quien abre la ronda. Viene del formulario |

Abrir una ronda pasa por las capas como revelar (sección 2 de US-016):

1. **Aplicación:** la sesión y el integrante existen y el integrante es del
   proyecto de la sesión (RN-017-1).
2. **Dominio** (`internal/domain/planningpoker`): una función pura recibe la
   sesión, quién abre la ronda y la votación de la ronda actual, valida todo y
   devuelve la sesión en la ronda siguiente o los errores.
3. **Persistencia:** cambia la ronda y el estado en una transacción con la
   sesión bloqueada (RN-017-8).

## 3. Salidas esperadas

| Qué | Después de abrir la ronda |
|---|---|
| Estado de la sesión | `votando` |
| Ronda | La siguiente (N + 1) |
| Participantes | Los mismos (RN-014-4) |
| Votos de la ronda nueva | Ninguno; ocultos mientras se vote (RN-015-7) |
| Votos de las rondas anteriores | Sin cambios, y se pueden leer (RN-017-6) |

El caso de uso devuelve la sesión en la ronda nueva. Si hay errores, la sesión
sigue `revelada` en la misma ronda y no cambia nada.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-017-1** (aplicación) — La sesión y el integrante tienen que existir, y el
  integrante tiene que ser del proyecto de la sesión, con prioridad y sin unirse
  con los errores de validación (mismo criterio que RN-016-1 y RN-018-1).
- **RN-017-2** (dominio) — Solo se abre con la sesión `revelada`. Con la sesión
  `votando` o `finalizada`, `ErrNuevaRondaSinRevelar`.
- **RN-017-3** (dominio) — Solo abre la ronda un participante (RN-014-4). Si
  no, `ErrNoEsParticipante`, el de US-015.
- **RN-017-4** (dominio) — Solo se abre si hay divergencia en la ronda actual,
  con la definición de RN-016-6 (más de un valor distinto entre los votos
  emitidos). Sin divergencia, `ErrSinDivergencia`: el equipo coincidió y
  registra la estimación acordada (US-018). La divergencia se evalúa solo con
  la sesión `revelada`: votando, los valores no se pueden leer (RN-015-7).
- **RN-017-5** (dominio) — Abrir la ronda deja la sesión `votando`, con la
  ronda sumada en 1 y los mismos participantes. No hay límite de rondas.
- **RN-017-6** (dominio) — **Votos de rondas anteriores.** Se conservan
  (RN-015-6) y se pueden leer: ya se revelaron. RN-015-7 oculta los valores de
  la **ronda actual** mientras la sesión vota; una ronda anterior a la actual
  ya está cerrada. La votación de una ronda deja leer sus valores si la ronda es
  anterior a la actual, o si la sesión no está `votando`. Así queda resuelto el
  hueco 4 de US-015.
- **RN-017-7** (dominio) — Los errores se unen con `errors.Join` (RN-014-10),
  en este orden fijo: nueva ronda sin revelar, no es participante, sin
  divergencia.
- **RN-017-8** (persistencia) — Transacción `READ COMMITTED` con los pasos de
  RN-016-8 y RN-018-9:
  1. Bloquea la sesión con `FOR NO KEY UPDATE` y lee su estado, su ronda y sus
     participantes. Si no hay fila, `ErrSesionInexistente`.
  2. Lee la votación de la ronda actual, después del candado.
  3. Vuelve a verificar RN-017-2 a RN-017-4 con la misma función del dominio.
  4. Pasa la sesión a `votando` con la ronda siguiente. No toca ningún voto.
- **RN-017-9** (persistencia) — El candado ordena lo que pasa a la vez sobre la
  sesión (RN-015-10, RN-016-9 y RN-018-10). Dos pedidos simultáneos de nueva
  ronda abren una sola: el segundo ve la sesión `votando` y se rechaza con
  `ErrNuevaRondaSinRevelar`. Una nueva ronda y un registro de la estimación
  acordada (US-018): el que llega segundo ve el estado del primero y se
  rechaza.
- **RN-017-10** (dominio y puertos) — **Ningún puerto entrega valores en
  crudo** (RN-016-12). El repositorio entrega la votación de una ronda como
  `Votacion`, y los valores salen del dominio solo por `Valores()`, con la regla
  de RN-017-6. `Votacion` no gana métodos exportados.

## 5. Restricciones

- **Dominio puro, errores unidos, 404 con prioridad, READ COMMITTED:** igual que
  US-014, US-015 y US-016.
- **Sin migración.** Los votos ya se guardan por ronda (RN-015-11) y la ronda y
  el estado son columnas de la sesión (RN-014-13).
- **El resultado de una ronda revelada es una función del dominio.** US-016
  arma la divergencia, el mínimo y el máximo dentro de `Revelar`. Esta historia
  necesita lo mismo sin revelar (para RN-017-4 y para leer rondas anteriores),
  así que el cálculo se extrae a una función que recibe una `Votacion` con los
  valores legibles y devuelve el resultado de la ronda (los votos, quiénes
  faltaron, la divergencia, el mínimo y el máximo). `Revelar` la usa, en un
  refactor con los tests de US-016 en verde. T-020 la usa para mostrar una
  sesión revelada y sus rondas anteriores.
- **Lectura de rondas anteriores.** El puerto suma una lectura que devuelve,
  para cada ronda anterior a la actual, su número y su `Votacion` (RN-017-10).
  La usan los escenarios de CA-017-2 y la pantalla de T-020.
- **Catálogo de errores (T-005).** En el PR de implementación se registran
  `planningpoker.ErrNuevaRondaSinRevelar` y `planningpoker.ErrSinDivergencia`
  como `conflicto` (RN-T005-2). Los demás ya están registrados.
- **Frases BDD.** Suma a la interfaz `PlanningPoker`
  `AbrirNuevaRonda(integrante, titulo string) error` y
  `VotosDeRonda(titulo string, ronda int) ([]VotoRevelado, error)`. Las frases
  nuevas, que se aprueban con el grupo antes de implementar:

  | Frase | Expresión |
  |---|---|
  | `Cuando el integrante "X" abre una nueva ronda en la sesión de la historia "H"` | `^el integrante "([^"]*)" abre una nueva ronda en la sesión de la historia "([^"]*)"$` |
  | `Y los votos de la ronda 1 de la sesión de la historia "H" son:` + tabla `integrante \| voto` | `^los votos de la ronda (\d+) de la sesión de la historia "([^"]*)" son:$` |

  Reutilizan las de US-014 a US-016: `está "votando" en la ronda 2`,
  `faltan votar:`, `están ocultos` y `reveló los votos`.
- **Fuera de alcance.** Cambiar los participantes entre rondas, un límite de
  rondas, cancelar la sesión y la pantalla (T-020).

## 6. Casos límite

- **CL-017-1** — Votos de 3 y 8 revelados: se abre la ronda 2, votando, sin
  votos, y todos los participantes figuran en `Faltan()`.
- **CL-017-2** — Votos contiguos (5 y 8): hay divergencia (RN-016-6) y se abre.
- **CL-017-3** — Todos votaron lo mismo, o hubo un solo voto: sin divergencia,
  `ErrSinDivergencia`.
- **CL-017-4** — Tercera ronda después de dos con divergencia: se abre la
  ronda 3, y las rondas 1 y 2 conservan sus votos.
- **CL-017-5** — En la ronda nueva, un participante vota: el voto va a la ronda
  nueva (RN-015-6) y los de la anterior no cambian.
- **CL-017-6** — Votos de la ronda anterior con la sesión votando la ronda
  nueva: se leen; los de la ronda actual siguen ocultos (RN-017-6).
- **CL-017-7** — Sesión `votando` (no se reveló) o `finalizada`:
  `ErrNuevaRondaSinRevelar`, sin evaluar la divergencia.
- **CL-017-8** — Integrante que no participa: `ErrNoEsParticipante`.
- **CL-017-9** — Sesión votando, integrante que no participa: los dos errores
  unidos, en el orden de RN-017-7.
- **CL-017-10** — Dos pedidos simultáneos: uno abre la ronda y el otro se
  rechaza con `ErrNuevaRondaSinRevelar`. Test concurrente determinístico contra
  un Postgres real, como el de US-005.
- **CL-017-11** — Una nueva ronda y un registro de la estimación acordada a la
  vez: si gana la ronda, el registro se rechaza con `ErrSesionNoRevelada`; si
  gana el registro, la ronda se rechaza con `ErrNuevaRondaSinRevelar`.
- **CL-017-12** — Sesión o integrante inexistentes, de otro proyecto o con IDs
  mal formados: solo el error de inexistencia (RN-017-1).

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) | Tipo (T-005) |
|---|---|---|---|
| La sesión no existe | `app.ErrSesionInexistente` | "la sesion de planning poker no existe" | `no_encontrado` (ya registrado) |
| El integrante no existe o es de otro proyecto | `app.ErrIntegranteInexistente` | "el integrante no existe" | `no_encontrado` (ya registrado) |
| La sesión no está revelada | `planningpoker.ErrNuevaRondaSinRevelar` | "la nueva ronda se abre con los votos revelados y la sesion sin finalizar" | `conflicto` (**nuevo**) |
| El integrante no participa | `planningpoker.ErrNoEsParticipante` | "el integrante no participa de esta sesion de planning poker" | `conflicto` (ya registrado) |
| No hay divergencia | `planningpoker.ErrSinDivergencia` | "no hay divergencia: registren la estimacion acordada" | `conflicto` (**nuevo**) |

Los dos primeros tienen prioridad y no se unen con los demás. Los tres últimos
se unen en el orden de RN-017-7, y el repositorio puede devolverlos con la
misma función. En ningún caso cambia nada.

## 8. Criterios de aceptación

- **CA-017-1** (normal) — Con la sesión revelada y con divergencia, abrir una
  nueva ronda deja la sesión votando en la ronda siguiente, con los mismos
  participantes y sin votos.
- **CA-017-2** (alternativo) — Los votos de las rondas anteriores se conservan,
  se pueden ver y no se confunden con los de la ronda nueva, que quedan ocultos
  mientras se vota.
- **CA-017-3** (límite) — Sin divergencia (todos votaron lo mismo) no se abre
  una nueva ronda.
- **CA-017-4** (error) — Con la sesión votando o finalizada, una sesión
  inexistente o un integrante que no participa, se rechaza con el error
  uniforme y no cambia nada.
- **CA-017-5** (error) — Dos pedidos simultáneos de nueva ronda abren una sola.

---

## Decisiones tomadas y descartadas

- **Solo con divergencia.** Sin divergencia el equipo ya coincidió: el paso
  siguiente es registrar la estimación (US-018). Se descartó permitir una ronda
  sin divergencia: es votar lo mismo otra vez.
- **Los votos anteriores se ven.** Ya se revelaron, así que ocultarlos no
  protege nada y le saca al equipo la referencia de qué cambió. RN-015-7 sigue
  protegiendo la ronda actual.
- **Mismos participantes.** La sesión fija sus participantes al iniciar
  (RN-014-4); cambiarlos entre rondas mezclaría votos de personas distintas.
- **Un error propio para "sin revelar".** `ErrSesionNoRevelada` (US-018) nombra
  la estimación acordada en su mensaje; se descartó generalizarlo para no
  cambiar una spec mergeada ni confundir a quien lee el error.
- **Mismo candado que votar, revelar y acordar.** Una sola regla de
  concurrencia para toda la sesión.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia y de las decisiones del equipo (solo con divergencia y la sesión
  revelada, mismos participantes, votos anteriores visibles, mismo candado),
  usando las specs de US-014, US-015, US-016 y US-018 como modelo
- Revisado y entendido por:
