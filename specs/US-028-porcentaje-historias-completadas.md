# US-028 — Calcular el porcentaje de historias completadas

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #34 |
| **Escenarios BDD** | `features/US-028-porcentaje-historias-completadas.feature` |
| **Código** | `internal/domain/metricas/` |
| **Última actualización** | 2026-10-18 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como Agile Enabler, quiero conocer el porcentaje de historias completadas, para
> ver de un vistazo qué parte del trabajo planificado se terminó.

Calcular qué porcentaje de las historias de una lista están completadas. Es la
métrica de "completitud" del Dashboard (US-030) y de la Review: complementa a
los Story Points completados de US-025, que miden el valor entregado, con una
cifra que cuenta historias y es más fácil de explicar en voz alta ("terminamos 3
de 4").

Depende de **US-025**: usa su mismo tipo de entrada, `ItemDelSprint`, así que
la misma lista de ítems alimenta las dos métricas.

## 2. Entradas

Función pura sobre la lista de historias que arma quien llama: las de un Sprint
(para el avance del Sprint) o las de todo el proyecto (para el avance general).
No accede a la base ni sabe de qué Sprint son.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `items` | `[]ItemDelSprint` (de US-025) | Sí (puede ser una lista vacía) | — |
| `ItemDelSprint.Completado` | `bool` | Sí | — |
| `ItemDelSprint.StoryPoints` | `int` | Sí | `>= 0`, como en US-025. Esta función no usa los puntos, pero rechaza los inválidos: la misma lista alimenta las dos métricas y un dato corrupto no se calcula en silencio |

## 3. Salidas esperadas

```
Porcentaje float64 // completadas ÷ total × 100
```

Redondeado a **1 decimal** (por ejemplo `66.7`), como la velocidad y la
desviación porcentual, porque se muestra a personas.

## 4. Reglas de negocio

- **RN-028-1** — El porcentaje es la cantidad de historias con
  `Completado = true` dividida por la cantidad total de historias, por 100.
- **RN-028-2** — Se cuentan **historias**, no Story Points: una historia de 13
  puntos completada pesa lo mismo que una de 1 punto. Para ponderar por tamaño
  está US-025.
- **RN-028-3** — El resultado se redondea a 1 decimal (`math.Round(v*10)/10`).
- **RN-028-4** — Una lista vacía da `0`, sin error (CL-028-1).
- **RN-028-5** — Se valida cada ítem con la regla de US-025: `StoryPoints`
  negativos se rechazan con `ErrStoryPointsNegativos`, indicando la posición del
  ítem (contada desde 1). Igual que en US-025, se devuelve el error del primer
  ítem inválido y ningún porcentaje parcial.

## 5. Restricciones

- **Dominio puro.** `internal/domain/metricas` no importa nada fuera de la
  biblioteca estándar y no lee el reloj ni la base.
- **Qué ítems entran lo decide quien llama.** Si el porcentaje es del Sprint o de
  todo el proyecto, y si una historia que volvió al backlog (US-011) cuenta o
  no, lo resuelve la capa de aplicación al armar la lista. Esta historia entrega
  la función de cálculo, igual que US-025 y US-026.
- **Escenarios BDD.** Se agregan al área Métricas de los steps, con las mismas
  frases de contexto que US-025 (`el sprint "S" con una historia de N story
  points completada` y `... sin completar`) y dos frases nuevas (ver
  Decisiones).
- **Fuera de alcance.** La pantalla, el desglose por prioridad o por integrante y
  el porcentaje de Story Points completados (es `Completados ÷ Planificados` de
  US-025, y se calcula con esos mismos números si hace falta).

## 6. Casos límite

- **CL-028-1** — Lista vacía (un Sprint recién creado): `0`, sin error. No es lo
  mismo que "ninguna completada" (CL-028-3): acá no hay historias.
- **CL-028-2** — Todas las historias completadas: `100`.
- **CL-028-3** — Ninguna historia completada (hay historias): `0`.
- **CL-028-4** — Una sola historia, completada: `100`; sin completar: `0`.
- **CL-028-5** — El resultado no es entero: 1 de 3 da `33.3` y 2 de 3 da `66.7`.
  Se redondea a 1 decimal, no se trunca ni se redondea a entero.
- **CL-028-6** — Historias con `0` Story Points: cuentan como cualquier otra
  (`StoryPoints = 0` es válido, RN-025-3).
- **CL-028-7** — Una historia de 13 puntos completada y tres de 1 punto sin
  completar: `25`, porque se cuentan historias y no puntos (RN-028-2). Con
  Story Points daría `81.3`.
- **CL-028-8** — Dos ítems con `StoryPoints` negativos en la lista (el 2 y el 4):
  se devuelve el error del primero, con la posición 2, y ningún porcentaje.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| Algún ítem tiene `StoryPoints` negativos | `metricas.ErrStoryPointsNegativos` (el de US-025), envuelto indicando la posición del ítem, contada desde 1 | El de US-025: "item en la posicion N: los story points de un item del sprint no pueden ser negativos" |

Se devuelve el error del primer ítem inválido, comparable con `errors.Is`. Una
lista vacía y una lista sin completadas no son errores: son casos límite con
resultado.

## 8. Criterios de aceptación

- **CA-028-1** (normal) — Dadas 4 historias en un Sprint, 3 completadas, el
  porcentaje de historias completadas es 75.
- **CA-028-2** (alternativo) — Dadas una historia grande completada y tres
  pequeñas sin completar, el porcentaje es 25: se cuentan historias, no Story
  Points.
- **CA-028-3** (límite) — Dado un Sprint sin historias, el porcentaje es 0, sin
  error.
- **CA-028-4** (límite) — Dadas 3 historias con 2 completadas, el porcentaje es
  66,7: se redondea a 1 decimal.
- **CA-028-5** (error) — Dada una historia con Story Points negativos, la
  función devuelve `ErrStoryPointsNegativos` indicando su posición y ningún
  porcentaje.

---

## Decisiones tomadas y descartadas

- **Se cuentan historias, no Story Points.** Es lo que dice el título de la
  historia y lo que se puede explicar en una frase. Los Story Points completados
  ya los da US-025; una segunda métrica con la misma fórmula no aportaría nada.
- **Se reutiliza `ItemDelSprint` de US-025.** La misma lista de ítems alimenta
  las dos métricas y se valida con la misma regla (RN-025-3 y su error), en lugar
  de definir un tipo paralelo con un solo campo.
- **Se rechazan los Story Points negativos aunque esta función no los use.**
  Aceptar en silencio un dato que US-025 rechazaría haría que el Dashboard
  muestre un porcentaje bueno junto a unos puntos que fallan. Se descartó
  ignorarlos: para eso bastaba una lista de booleanos.
- **Lista vacía es `0` y no un error.** Un Sprint recién creado no tiene
  historias: es el estado normal, igual que cero sprints cerrados en US-026
  (CL-026-1). Que la pantalla muestre "sin historias" en lugar de "0 %" es una
  decisión de interfaz (T-011), no de este cálculo.
- **Redondeo a 1 decimal**, como la velocidad (RN-026-2) y la desviación
  porcentual (US-027), para que todas las cifras del Dashboard tengan la misma
  precisión.
- **Frases nuevas del diccionario** (se aprueban en el PR de implementación):
  `se consulta el porcentaje de historias completadas del sprint "S"` y
  `el porcentaje de historias completadas es N` (acepta coma o punto decimal).
  Las dos pertenecen al área Métricas.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la
  historia, usando las de US-025 y US-026 como modelo; las decisiones se
  proponen acá y se confirman en la revisión
- Revisado y entendido por: Tomás Romero
