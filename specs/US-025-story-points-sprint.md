# US-025 — Calcular Story Points planificados y completados de un Sprint

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás |
| **Issue** | #11 |
| **Escenarios BDD** | `features/US-025-story-points-sprint.feature` |
| **Código** | `internal/domain/metricas/` |
| **Última actualización** | 2026-10-05 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Calcular, para un Sprint, cuántos Story Points se planificaron y cuántos se
completaron. Es la métrica base del Sprint: el Sprint Board y la Review la
necesitan para mostrar avance, y **US-026 (velocidad del equipo) la usa como
entrada**, promediando esta misma cifra sobre los sprints cerrados.

## 2. Entradas

El cálculo es una función pura: recibe la lista de ítems del backlog asignados
a un sprint, no el `Sprint` completo ni accede a la base. Los dominios de
`backlog` y `sprint` (US-005, US-008) todavía no existen, así que se define acá
el tipo mínimo que la función necesita; cuando esos dominios existan, la capa
de aplicación (`internal/app/`) va a adaptar sus tipos a este.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `items` | `[]ItemDelSprint` | Sí (puede ser una lista vacía) | — |
| `ItemDelSprint.StoryPoints` | `int` | Sí | >= 0 |
| `ItemDelSprint.Completado` | `bool` | Sí | — |

## 3. Salidas esperadas

Un `ResumenSprint` con dos enteros:

```
ResumenSprint{
    Planificados int // suma de StoryPoints de TODOS los items de la lista
    Completados  int // suma de StoryPoints solo de los items con Completado = true
}
```

Story Points se suman como enteros; no hay unidades fraccionarias en la escala
Fibonacci que usa el proyecto (US-013).

## 4. Reglas de negocio

- **RN-025-1** — `Planificados` es la suma de los Story Points de todos los
  ítems recibidos, estén completados o no.
- **RN-025-2** — `Completados` es la suma de los Story Points únicamente de
  los ítems con `Completado = true`.
- **RN-025-3** — Un ítem con `StoryPoints = 0` participa del cálculo sin
  alterar ninguna de las dos sumas y no es un error: esta función no impone
  la escala de estimación (eso es de US-013). Un ítem *sin estimar* no se
  representa como `0`: en US-005 es un valor ausente, y la capa de aplicación
  lo trata como error al adaptarlo a `ItemDelSprint`. La Definition of Ready,
  que valida US-009, ya impide que un ítem sin estimar entre a un sprint.
- **RN-025-4** — Por construcción, `Completados` nunca puede superar a
  `Planificados`, porque todo ítem completado forma parte del mismo conjunto
  que se suma en `Planificados`. No hace falta una validación aparte para esto.

## 5. Restricciones

- La función no lee el reloj, no genera números aleatorios ni accede a
  variables de entorno ni a la base de datos: es determinística y solo
  depende de sus parámetros.
- No valida que `StoryPoints` pertenezca a la escala Fibonacci (1, 2, 3, 5, 8,
  13): esa regla es responsabilidad de la estimación (US-013), no de este
  cálculo. Acá solo importa que no sea negativo.

## 6. Casos límite

- **CL-025-1** — Lista de ítems vacía (sprint recién creado, sin historias
  asignadas todavía): `Planificados = 0` y `Completados = 0`.
- **CL-025-2** — Un solo ítem, completado: `Planificados = Completados` =
  los Story Points de ese ítem.
- **CL-025-3** — Todos los ítems completados: `Planificados = Completados`.
- **CL-025-4** — Ningún ítem completado: `Completados = 0` y `Planificados`
  es la suma total.
- **CL-025-5** — Un ítem con `StoryPoints = 0`, completado o no: no altera
  ninguna de las dos sumas (RN-025-3).

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| Algún ítem tiene `StoryPoints < 0` | `ErrStoryPointsNegativos`, envuelto con `fmt.Errorf` indicando la posición del ítem en la lista, contada desde 1 (igual que en US-005, RN-005-14) | "los story points de un item del sprint no pueden ser negativos" |

## 8. Criterios de aceptación

- **CA-025-1** (normal) — Dado un sprint con ítems completados y sin
  completar, el resumen calcula `Planificados` como la suma de todos y
  `Completados` como la suma de solo los completados.
- **CA-025-2** (alternativo) — Dado un sprint donde todos los ítems están
  completados, `Planificados` y `Completados` son iguales.
- **CA-025-3** (límite) — Dado un sprint sin ítems, el resumen devuelve
  `Planificados = 0` y `Completados = 0`, sin error.
- **CA-025-4** (error) — Dado un sprint con un ítem de Story Points
  negativos, la función devuelve `ErrStoryPointsNegativos` y ningún
  `ResumenSprint` parcial.

---

## Decisiones tomadas y descartadas

- **Se definió `ItemDelSprint` como tipo propio de `metricas`, en vez de
  esperar a los dominios de `backlog` y `sprint` (US-005, US-008).** Esas
  historias todavía no están implementadas y esta métrica no necesita nada
  de ellas salvo dos datos (Story Points y si está completado). Definir un
  tipo mínimo acá evita bloquear el Sprint Goal de "primeras métricas" por
  una dependencia que no le hace falta, y es fácil de adaptar después desde
  `internal/app/` cuando exista `backlog.ItemBacklog`.
- **Se descartó devolver un `ResumenSprint` parcial junto con el error** de
  Story Points negativos (por ejemplo, sumando lo válido e ignorando lo
  inválido). Un dato negativo indica una inconsistencia previa (un bug en
  quien llama, o datos corruptos), y devolver una cifra parcial sin avisar
  sería más engañoso que devolver un error y nada.
- **Se descartó validar la escala Fibonacci acá.** Mezclaría dos
  responsabilidades: esta función mide lo que ya está estimado, no valida
  cómo se estimó. Esa validación es de US-013.

## Uso de IA en esta especificación

- [x] Use IA para: redacción completa de la spec a partir de la historia, la
  épica y el plan de trabajo — revisado por: Tomás.
