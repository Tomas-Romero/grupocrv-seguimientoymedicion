# US-016 — Revelar las estimaciones y detectar divergencias entre integrantes

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #27 |
| **Escenarios BDD** | `features/US-016-revelar-estimaciones.feature` |
| **Código** | `internal/domain/planningpoker/` (revelar y divergencia), caso de uso en `internal/app/`, repositorio en `internal/adapters/postgres/` y `internal/adapters/memoria/`. Sin migración nueva |
| **Última actualización** | 2026-10-22 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero revelar las estimaciones de todos a la vez
> y ver si hay divergencias, para decidir si hace falta discutir y votar de
> nuevo.

Revelar cierra la votación de la ronda: la sesión pasa de `votando` a
`revelada` y, desde ese momento, los valores de los votos se ven. Junto con los
votos, el sistema dice si hay **divergencia** (los participantes votaron cosas
distintas), que es lo que decide si el equipo abre una ronda nueva (US-017) o
registra la estimación acordada (US-018).

Es la tercera historia de Planning Poker. Usa lo que definen US-014 (la sesión,
sus estados y su ronda) y US-015 (la votación que oculta los valores).

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `sesionID` | UUID | Sí | Una sesión existente (US-014). Viene de la URL; la verifica el caso de uso |
| `integranteID` | UUID | Sí | Un integrante del proyecto de la sesión. Es quien revela. Viene de la selección del formulario; lo verifica el caso de uso |

La ronda no es una entrada: se revela la ronda actual de la sesión (RN-016-7).

Revelar pasa por las capas en este orden:

1. **Aplicación:** verifica con lecturas simples que la sesión exista
   (`ErrSesionInexistente`) y que el integrante exista y sea del proyecto de la
   sesión (`ErrIntegranteInexistente`), en ese orden y sin validar nada más
   (RN-016-1). Son las mismas lecturas y los mismos errores que votar (US-015).
2. **Dominio** (`internal/domain/planningpoker`): una función pura recibe el
   estado de la sesión, la lista de participantes, los votos de la ronda y quién
   revela; valida todo y devuelve la revelación o todos los errores.
3. **Persistencia:** pasa la sesión a `revelada` en una transacción con la
   sesión bloqueada (RN-016-8).

## 3. Salidas esperadas

La **revelación** de la ronda:

```
Revelacion {
    SesionID    string // la sesión revelada
    Ronda       int    // la ronda que se reveló
    Votos       []Voto // los participantes que votaron, en el orden de la lista de la sesión
    Faltaron    []string // IDs de los participantes que no votaron, en ese mismo orden
    Divergencia bool   // true si los valores de Votos no son todos iguales
    Minimo      int    // el valor más bajo de Votos
    Maximo      int    // el valor más alto de Votos
}

Voto { ParticipanteID string; Valor int }
```

Si hay errores, la sesión sigue `votando`, no se guarda nada y no se devuelve
revelación.

Los valores salen del dominio solo por `Valores()` de la votación de US-015
(RN-015-7), que con la sesión ya `revelada` o `finalizada` los devuelve. Esta
historia no agrega otra forma de leerlos.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-016-1** (aplicación) — La sesión y el integrante tienen que existir, y
  el integrante tiene que ser del proyecto de la sesión. Se verifican en el
  orden de la sección 2, con prioridad y sin unirse con los errores de
  validación (mismo criterio que RN-015-1). Un integrante de otro proyecto o un
  ID mal formado son inexistentes (CL-019-15).
- **RN-016-2** (dominio) — Solo se revela con la sesión en `votando`
  (RN-014-8). Con la sesión `revelada` o `finalizada`, `ErrVotacionCerrada` (el
  de US-015, por la misma razón: la votación de esta ronda ya está cerrada).
  Revelar dos veces no es idempotente: la segunda se rechaza.
- **RN-016-3** (dominio) — Solo revela un participante de la sesión (RN-014-4).
  Un integrante del proyecto que no participa se rechaza con
  `ErrNoEsParticipante`, el de US-015.
- **RN-016-4** (dominio) — Hace falta al menos un voto en la ronda. Con la ronda
  vacía, `ErrSinVotos`: no hay nada que revelar ni que comparar.
- **RN-016-5** (dominio) — **Se puede revelar con votos faltantes.** No hace
  falta que hayan votado todos: el cierre es manual (decisión de US-015), y
  exigirlo trabaría la sesión si alguien no vota. Los que no votaron figuran en
  `Faltaron`.
- **RN-016-6** (dominio) — **Divergencia.** Hay divergencia si entre los votos
  emitidos hay más de un valor distinto. Cualquier diferencia cuenta, incluso
  entre cartas contiguas de la escala (5 y 8). Los participantes que no votaron
  no cuentan. Con un solo voto, o con todos iguales, no hay divergencia. `Minimo`
  y `Maximo` son el menor y el mayor de los valores votados (iguales si no hay
  divergencia).
- **RN-016-7** (dominio y persistencia) — Se revela la **ronda actual** de la
  sesión. Los votos de las rondas anteriores (RN-015-6) no entran en la
  revelación.
- **RN-016-8** (persistencia) — El repositorio revela en una transacción
  `READ COMMITTED`, con estos pasos en orden, y se detiene en el primer error:
  1. Bloquea la fila de la sesión con `FOR NO KEY UPDATE` y lee su estado y su
     ronda. Si no hay fila, `ErrSesionInexistente`.
  2. Lee los participantes y los votos de la ronda actual, en sentencias
     posteriores al candado.
  3. Vuelve a verificar RN-016-2 a RN-016-4 con la misma función del dominio y
     esos datos.
  4. Pasa la sesión a `revelada` (`UPDATE ... SET estado = 'revelada'`).
- **RN-016-9** (persistencia) — El candado del paso 1 ordena lo que pasa a la
  vez sobre la sesión, con el mismo criterio que RN-015-10. Un voto y el
  revelado: si el voto llega antes, entra en la revelación; si llega después,
  ve la sesión `revelada` y se rechaza con `ErrVotacionCerrada`. Dos revelados
  simultáneos: el primero revela y el segundo, que espera el candado, ve la
  sesión `revelada` y se rechaza con `ErrVotacionCerrada` (CL-016-9).
- **RN-016-10** (dominio) — El dominio valida todo y devuelve todos los errores
  unidos con `errors.Join` (RN-014-10), en este orden fijo: votación cerrada, no
  es participante, sin votos.
- **RN-016-11** (aplicación) — La operación es de todo o nada: revelar no
  modifica ningún voto.

## 5. Restricciones

- **Dominio puro, errores unidos, 404 con prioridad:** igual que US-014 y
  US-015 (sección 5).
- **Depende de US-014 y de US-015, implementadas.** Usa la sesión, sus estados
  y su ronda (US-014), y la votación de la ronda con `Votaron()`, `Faltan()` y
  `Valores()`, `ErrVotacionCerrada` y `ErrNoEsParticipante` (US-015). La
  implementación de esta historia se abre cuando esas dos estén mergeadas. Esta
  spec, en cambio, ya se puede revisar: solo cita sus nombres.
- **Sin migración.** El estado `revelada` ya está permitido por el `CHECK` de la
  tabla de sesiones (RN-014-13) y los votos ya se guardan por ronda (RN-015-11).
  Revelar solo cambia el estado de la sesión.
- **Catálogo de errores (T-005).** En el PR de implementación se registra
  `planningpoker.ErrSinVotos` en `internal/adapters/http/catalogo.go` como
  `conflicto`: depende de lo que hay guardado (RN-T005-2). `ErrVotacionCerrada`,
  `ErrNoEsParticipante` y `app.ErrSesionInexistente` los registra el PR de
  US-015.
- **Frases BDD.** Suma a la interfaz `PlanningPoker` (US-014 y US-015)
  `Revelar(integrante, titulo string) error`,
  `VotosRevelados(titulo string) ([]VotoRevelado, error)` y
  `HayDivergencia(titulo string) (bool, error)`, con
  `VotoRevelado{Integrante string; Valor int}`. Las frases nuevas, que se
  aprueban con el grupo antes de implementar:

  | Frase | Expresión |
  |---|---|
  | `Cuando el integrante "X" revela los votos de la sesión de la historia "H"` | `^el integrante "([^"]*)" revela los votos de la sesión de la historia "([^"]*)"$` |
  | `Entonces los votos de la sesión de la historia "H" son:` + tabla `integrante \| voto` | `^los votos de la sesión de la historia "([^"]*)" son:$` |
  | `Y hay divergencia en la sesión de la historia "H"` | `^hay divergencia en la sesión de la historia "([^"]*)"$` |
  | `Y no hay divergencia en la sesión de la historia "H"` | `^no hay divergencia en la sesión de la historia "([^"]*)"$` |

  Los escenarios reutilizan las frases de US-014 (`Dado una sesión de Planning
  Poker sobre la historia "X"`, `Entonces la sesión de la historia "X" está
  "revelada" en la ronda 1`) y de US-015 (`Dado el integrante "X" ya votó 5 en
  la sesión de la historia "H"`, `Y en la sesión de la historia "H" faltan
  votar:`, `Y los votos de la sesión de la historia "H" están ocultos`). Después
  de un rechazo, los escenarios comprueban que la sesión sigue `votando` y que
  los votos siguen ocultos.
- **Fuera de alcance.** Votar (US-015), abrir una ronda nueva con los mismos
  participantes (US-017), registrar la estimación acordada y finalizar la sesión
  (US-018), consultar una votación ya revelada desde la pantalla, la pantalla de
  revelado y la forma de mostrar los extremos (tarea propia del Sprint 5).

## 6. Casos límite

- **CL-016-1** — Votaron todos con el mismo valor: la sesión pasa a `revelada`,
  `Divergencia` es falso, `Faltaron` está vacío y `Minimo` es igual a `Maximo`.
- **CL-016-2** — Votaron todos y hay al menos dos valores distintos:
  `Divergencia` es verdadero.
- **CL-016-3** — Valores contiguos en la escala (5 y 8): hay divergencia
  (RN-016-6).
- **CL-016-4** — Un solo voto de varios participantes: se revela, no hay
  divergencia y los demás figuran en `Faltaron` (RN-016-5).
- **CL-016-5** — Votaron algunos y los que votaron coinciden: no hay
  divergencia, aunque falten otros. `Faltaron` los nombra.
- **CL-016-6** — Nadie votó: `ErrSinVotos` y la sesión sigue `votando`
  (RN-016-4).
- **CL-016-7** — Un participante cambió su voto antes de revelar: la revelación
  trae el último (RN-015-5), una sola vez.
- **CL-016-8** — La sesión ya está `revelada` o `finalizada`, y además nadie
  votó: `ErrVotacionCerrada` y `ErrSinVotos` unidos, en el orden de RN-016-10.
  Con datos reales no se da (una sesión revelada ya tuvo votos), pero la función
  del dominio es pura y se prueba con ese caso.
- **CL-016-9** — Dos revelados simultáneos: los dos terminan sin error de la
  base, uno revela y el otro se rechaza con `ErrVotacionCerrada`. Se verifica
  con un test concurrente determinístico contra un Postgres real, como el de
  US-014.
- **CL-016-10** — Un voto y el revelado a la vez: si el voto confirma antes,
  figura en la revelación; si llega después, se rechaza con
  `ErrVotacionCerrada` y no cambia los votos revelados. Se verifica con el mismo
  tipo de test concurrente.
- **CL-016-11** — Después de revelar, un voto de la misma ronda se rechaza con
  `ErrVotacionCerrada` (RN-015-3) y no modifica los votos revelados.
- **CL-016-12** — Revela un integrante que no participa de la sesión (se
  registró después de iniciarla): `ErrNoEsParticipante`.
- **CL-016-13** — Sesión o integrante inexistentes, integrante de otro proyecto
  o IDs mal formados: solo el error de inexistencia (RN-016-1).
- **CL-016-14** — La sesión se revela entre la lectura del caso de uso y la
  transacción: el repositorio la ve `revelada` con el candado tomado y rechaza
  con `ErrVotacionCerrada`. Se verifica llamando al repositorio con una sesión
  ya revelada en la base.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) | Tipo (T-005) |
|---|---|---|---|
| La sesión no existe | `app.ErrSesionInexistente`, envuelto indicando el ID | "la sesion de planning poker no existe" | `no_encontrado` (lo registra US-015) |
| El integrante no existe o es de otro proyecto | `app.ErrIntegranteInexistente`, envuelto indicando el ID | "el integrante no existe" | `no_encontrado` (ya registrado) |
| La sesión no está votando | `planningpoker.ErrVotacionCerrada` | "la votacion de esta ronda esta cerrada" | `conflicto` (lo registra US-015) |
| El integrante no participa de la sesión | `planningpoker.ErrNoEsParticipante` | "el integrante no participa de esta sesion de planning poker" | `conflicto` (lo registra US-015) |
| No hay ningún voto en la ronda | `planningpoker.ErrSinVotos` | "no hay votos para revelar en esta ronda" | `conflicto` (**nuevo**) |

Los dos primeros tienen prioridad y no se unen con los demás (RN-016-1). Los
tres últimos son del dominio y se unen en el orden de RN-016-10; el repositorio
puede devolverlos también, con la misma función. Todos se comparan con
`errors.Is`, y en ningún caso se cambia el estado de la sesión.

## 8. Criterios de aceptación

- **CA-016-1** (normal) — Cuando todos los participantes votaron lo mismo y uno
  revela, la sesión queda revelada, se ven los votos de todos y no hay
  divergencia.
- **CA-016-2** (alternativo) — Cuando los participantes votaron valores
  distintos y uno revela, se ven los votos y hay divergencia.
- **CA-016-3** (alternativo) — Se puede revelar aunque falten votos: la sesión
  queda revelada, se ven los votos emitidos y se sabe quiénes no votaron.
- **CA-016-4** (límite) — Con valores contiguos de la escala (5 y 8) hay
  divergencia; con un único voto emitido, no la hay.
- **CA-016-5** (error) — Si nadie votó, no se revela y la sesión sigue votando.
- **CA-016-6** (error) — Un integrante que no participa de la sesión no revela.
- **CA-016-7** (error) — Una sesión ya revelada o finalizada no se revela de
  nuevo, y una vez revelada ya no se vota.
- **CA-016-8** (error) — Una sesión o un integrante inexistentes se rechazan y
  no se cambia nada.

---

## Decisiones tomadas y descartadas

- **Revelar es manual y cierra la ronda.** Lo decidió el equipo en US-014/015:
  se descartó cerrar solo cuando votaron todos, porque una sesión con alguien
  que no vota quedaría trabada.
- **Se puede revelar con votos faltantes.** Es la consecuencia del cierre
  manual. La revelación nombra a quienes no votaron para que el equipo decida si
  discute o vuelve a votar. Se descartó exigir un mínimo de votos (por ejemplo,
  la mitad): es una regla inventada que ningún criterio pide.
- **Divergencia es "más de un valor distinto", sin tolerancia.** Se descartó
  tratar las cartas contiguas como consenso (5 y 8): es una regla de equipo, no
  del producto, y con ella dos personas podrían creer que coinciden cuando no.
  La decisión de aceptar una diferencia chica es de US-018, que registra el valor
  acordado.
- **Revela un participante.** Sin usuarios en el sistema, el integrante llega
  desde el formulario, igual que al votar. Se descartó que revele cualquier
  integrante del proyecto: el mismo criterio de RN-015-2 deja una sola regla de
  "quién interviene en la sesión".
- **Revelar dos veces es un error.** `ErrVotacionCerrada` es el mismo error de
  votar con la sesión cerrada, así que el catálogo no suma nada. Se descartó
  hacerlo idempotente: ocultaría un doble clic o a dos personas revelando a la
  vez, y la pantalla puede mostrar la votación ya revelada por su cuenta.
- **No se guarda el resultado de la revelación.** La divergencia, el mínimo y el
  máximo se calculan con los votos de la ronda, que ya están guardados: no hay
  migración ni datos que puedan quedar desactualizados.
- **Los valores salen del dominio solo por `Valores()`.** Revelar no suma otra
  forma de leer los votos, así que la garantía de US-015 (el voto oculto lo
  asegura el dominio) tiene un único punto de control.

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia; se listan para no agregar reglas por
cuenta propia:

1. **Quién revela por quién.** Igual que al votar, cualquiera podría mandar el
   revelado en nombre de otro participante, porque no hay usuarios.
2. **Revelar con votos faltantes en una sesión grande.** Si un integrante no
   vota nunca, el equipo igual puede revelar. La pantalla (Sprint 5) puede
   advertir "faltan N votos" antes de confirmar.
3. **Qué pasa con una revelación sin divergencia pero con faltantes.** El
   sistema dice que no hay divergencia entre los que votaron; US-018 es quien
   decide si alcanza para registrar la estimación.
4. **Consultar la votación revelada.** Esta historia devuelve la revelación al
   que revela. Mostrarla después a quien entra a la sesión es una lectura que
   pertenece a la pantalla del Sprint 5.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la
  historia y de las decisiones del equipo (cierre manual, revelar con votos
  faltantes, divergencia por valores distintos, mismo candado que la votación),
  usando las specs de US-014 y US-015 como modelo
- Revisado y entendido por: Tomás Romero
