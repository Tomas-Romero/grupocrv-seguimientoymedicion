# ADR 0002 — Integración de BDD con godog

- **Estado:** aceptada
- **Fecha:** 2026-10-05
- **Decide:** Tomás (T-006), a partir del esqueleto inicial de Angelo (runner,
  steps y diccionario de frases)

## Contexto

La cátedra pide escenarios BDD automatizados, en Gherkin, con los cuatro tipos
de caso (normal, alternativo, límite y error). El stack ya eligió `godog`
(ADR 0001), pero faltaba decidir **cómo se ejecuta** dentro del proyecto: qué
comando lo corre, contra qué se prueban los escenarios y cómo se reparten los
steps entre tres personas que trabajan en historias distintas a la vez.

`godog` es una dependencia fuera del stack original de `go.mod`; este ADR es el
que exige `CONTRIBUTING.md` antes de usarla.

## Decisión

**1. godog se ejecuta como test de Go.** `features/main_test.go` define
`TestEscenarios`, que corre todos los `.feature` de `features/`. Es lo que ya
invocan `make bdd` y el job `BDD` del CI (`go test ./features/...`); no hace
falta un binario aparte.

**2. Modo estricto.** Un paso sin implementar cuenta como falla, no como
pendiente. Sin esto, un escenario nuevo sin sus steps pasaría en verde sin
haber probado nada.

**3. Los steps hablan con la aplicación a través de una interfaz por área**
(`Proyectos`, `Backlog`, `Sprints`, `Metricas`), agrupadas en `Servicios`. Cada
área se conecta en `features/servicios_test.go` cuando se mergea la historia que
la implementa. Un área sin conectar hace fallar con un mensaje explícito
(`area sin conectar a los steps: proyectos`) a los escenarios que la usan, y a
ningún otro.

**4. Los escenarios se prueban contra la capa de aplicación**, no contra HTTP ni
contra Postgres. Los `Servicios` los arma el runner con implementaciones sobre
`internal/app`. La persistencia se prueba aparte, en los tests del adaptador.

**5. Las frases están en `docs/diccionario-steps.md`.** Es el contrato entre
quien escribe un `.feature` y quien implementa el step. No se inventan frases
sueltas: se proponen y se agregan al diccionario. Después de `Dado` no va
`que`: en el dialecto español esa palabra clave es solo `Dado`, y el `que`
quedaría dentro del texto del step.

**6. El `.feature`, sus steps y el dominio entran en el mismo PR.** Por el modo
estricto, mergear un `.feature` sin su implementación deja el CI en rojo.

**7. Hay un escenario de cableado** (`features/T-006-cableado-bdd.feature`) que
atraviesa runner, steps y servicios con un doble en memoria del área de
proyectos. Prueba que la integración funciona antes de que exista la primera
historia. Se borra, junto con el doble, cuando entre el primer escenario real
(US-001).

## Alternativas consideradas

**Una sola interfaz `Sistema` con todos los métodos** (la versión inicial, 22
métodos de proyectos, backlog, sprints y métricas). Descartada: para que
compile, una sola struct tiene que implementar los 22, y mientras falte uno
cualquier `.feature` pone el CI en rojo. Además mezcla historias de tres
personas en un único punto de conflicto, y cada historia nueva lo engorda.

**Escenarios contra HTTP.** Descartada: acopla la especificación al transporte y
a la interfaz de usuario, y es más lenta. Los escenarios describen reglas de
negocio, no pantallas.

**Escenarios contra Postgres real.** Descartada para BDD: cada escenario
tendría que limpiar la base, y el CI de BDD tardaría más sin probar nada de
negocio que los tests de dominio no prueben. La persistencia tiene sus propios
tests.

**Binario `godog` aparte.** Descartada: separa la corrida de BDD del resto de
`go test`, no se integra con `-race` ni con la cobertura y suma un paso al CI.

## Consecuencias

**A favor**

- Cada historia agrega sus steps y conecta su área sin tocar las de otra persona.
- `make bdd`, `make check` y el CI funcionan sin cambios.
- Los escenarios corren en milisegundos y no dependen de Docker.

**En contra**

- Los escenarios no prueban el cableado HTTP ni la base real: eso queda a cargo
  de los tests de adaptador.
- Las frases del diccionario son un contrato compartido; agregar una requiere
  acordarla antes. Es un costo de coordinación aceptado a cambio de que no haya
  dos frases para la misma idea.
