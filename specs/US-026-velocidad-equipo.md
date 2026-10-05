# US-026 — Calcular la velocidad del equipo sobre los sprints cerrados

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás |
| **Issue** | #12 |
| **Escenarios BDD** | `features/US-026-velocidad-equipo.feature` |
| **Código** | `internal/domain/metrics/` |
| **Última actualización** | 2026-09-29 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Calcular la velocidad del equipo: el promedio de Story Points completados por
sprint, a partir de los sprints ya cerrados. Es el número que se usa para
replanificar en cada Sprint Planning (riesgo R1 del plan) y el que va a
graficar el Dashboard (US-032) como velocidad histórica.

Depende de **US-025**: cada sprint cerrado aporta su `ResumenSprint`, y esta
historia promedia el campo `Completados` de esos resúmenes.

## 2. Entradas

Función pura sobre una lista de sprints ya cerrados. No accede a la base ni
sabe qué sprint está "en curso": eso lo filtra quien llama, antes de pasarle
la lista.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `sprints` | `[]SprintCerrado` | Sí (puede ser una lista vacía) | — |
| `SprintCerrado.Nombre` | `string` | Sí | No vacío (identifica el sprint, ej. `"Sprint 1"`) |
| `SprintCerrado.Resumen` | `ResumenSprint` (de US-025) | Sí | `Resumen.Completados <= Resumen.Planificados` |

## 3. Salidas esperadas

```
Velocidad float64 // Story Points completados, promedio por sprint
```

Redondeada a **1 decimal** (por ejemplo `31.5`), porque se muestra a personas
en el Dashboard y en la Review; más precisión que esa no aporta nada y
complica la lectura.

## 4. Reglas de negocio

- **RN-026-1** — La velocidad es el promedio aritmético del campo
  `Completados` de todos los `ResumenSprint` recibidos.
- **RN-026-2** — El resultado se redondea a 1 decimal (`math.Round(v*10)/10`).
- **RN-026-3** — Antes de promediar, se valida cada resumen contra el
  invariante de US-025 (RN-025-4: `Completados <= Planificados`). Si algún
  sprint lo viola, la función no promedia nada: es un dato corrupto, no un
  caso a calcular.

## 5. Restricciones

- No usa `time.Now()` ni ningún dato del reloj: qué sprints están "cerrados"
  lo decide quien llama (la capa de aplicación, contra el estado real del
  Sprint), no esta función.
- No pondera sprints recientes por sobre los antiguos: es un promedio simple.
  Si el equipo necesita una media móvil más adelante, es una historia nueva,
  no un cambio silencioso acá.

## 6. Casos límite

- **CL-026-1** — Ningún sprint cerrado todavía (lista vacía): `Velocidad = 0`,
  sin error. Es la situación normal antes de cerrar el primer sprint; no
  hay velocidad medida y 0 es más honesto que inventar un valor.
- **CL-026-2** — Un solo sprint cerrado: la velocidad es exactamente los
  Story Points completados de ese sprint (promedio de un elemento).
- **CL-026-3** — Todos los sprints cerrados con 0 completados: `Velocidad = 0`
  (no es el mismo caso que CL-026-1: acá sí hubo sprints, simplemente no se
  completó nada).
- **CL-026-4** — El promedio da un número no entero (por ejemplo, 31.5): se
  redondea a 1 decimal, no se trunca ni se redondea a entero.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| Algún `SprintCerrado` tiene `Resumen.Completados > Resumen.Planificados` | `ErrResumenInconsistente`, envuelto indicando el nombre del sprint | "el sprint «Nombre» tiene mas Story Points completados que planificados: revisa sus datos" |
| Algún `SprintCerrado.Nombre` está vacío | `ErrNombreVacio`, envuelto indicando la posición en la lista, contada desde 1 (igual que en US-005, RN-005-14) | "todo sprint cerrado necesita un nombre para poder mostrarlo" |

## 8. Criterios de aceptación

- **CA-026-1** (normal) — Dados varios sprints cerrados con distintos Story
  Points completados, la velocidad es el promedio de esos valores, redondeado
  a 1 decimal.
- **CA-026-2** (alternativo) — Dado un único sprint cerrado, la velocidad es
  igual a los Story Points completados de ese sprint.
- **CA-026-3** (límite) — Dada una lista vacía de sprints cerrados, la
  velocidad es `0`, sin error.
- **CA-026-4** (error) — Dado un sprint cerrado con `Completados` mayor a
  `Planificados`, la función devuelve `ErrResumenInconsistente` y ninguna
  velocidad.

---

## Decisiones tomadas y descartadas

- **Se reutiliza `ResumenSprint` de US-025 en vez de recibir listas sueltas de
  enteros.** Mantiene una sola fuente de verdad para "qué es un resumen de
  sprint válido" (el invariante `Completados <= Planificados` vive en un solo
  lugar) y evita que esta historia reimplemente esa regla.
- **Se agregó `Nombre` a `SprintCerrado`**, aunque el cálculo no lo necesita
  para promediar. Se incluye porque US-032 (gráficos de velocidad histórica)
  va a necesitar mostrar la velocidad por sprint, y es más simple llevar el
  nombre desde acá que inventar una estructura paralela después.
- **Se descartó una media móvil o un promedio ponderado** por sprints
  recientes. Es más difícil de explicar en la defensa y la guía no lo pide;
  si hace falta suavizar el número, se discute como una historia nueva con su
  propia spec.
- **Se descartó devolver un error cuando la lista está vacía.** No cerrar
  ningún sprint todavía es el estado normal del proyecto al principio, no un
  dato corrupto: por eso es un caso límite (CL-026-1) y no un error.

## Uso de IA en esta especificación

- [x] Use IA para: redacción completa de la especificación (los 8 apartados)
  a partir de la historia, la épica y su dependencia con US-025 — revisado
  por: Tomás.
