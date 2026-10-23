# US-027 — Calcular horas estimadas, horas reales y la desviación entre ambas

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #33 |
| **Escenarios BDD** | `features/US-027-horas-estimadas-reales.feature` |
| **Código** | `internal/domain/metricas/` |
| **Última actualización** | 2026-10-22 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como Agile Enabler, quiero conocer las horas estimadas, las horas reales y la
> desviación entre ambas de un Sprint, para saber cuánto nos equivocamos al
> estimar y calibrar el próximo Planning.

Calcular, para un Sprint, cuántas horas se estimaron, cuántas se trabajaron de
verdad y cuánto se desvió una cifra de la otra. Son las tres filas de esfuerzo
del acta de cada Sprint ("Horas estimadas", "Horas reales" y "Desviación de
esfuerzo"), que hasta ahora quedaron sin registrar, y la entrada del Dashboard
(US-030) y del reporte (US-033).

Depende de tres historias: **US-025** da los Story Points planificados del
Sprint, **US-013** garantiza que esos puntos sean de la escala Fibonacci y
**US-019** da las horas reales que registra cada integrante.

## 2. Entradas

Función pura sobre datos que ya armó quien llama. No accede a la base, no sabe
a qué Sprint pertenecen las horas ni cómo se calcularon los Story Points:
recibe los números.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `storyPointsPlanificados` | `int` | Sí | `>= 0`. Es `ResumenSprint.Planificados` de US-025: la suma de los Story Points de **todas** las historias del Sprint, completadas o no (RN-025-1) |
| `horasRegistradas` | `[]float64` | Sí (puede ser una lista vacía) | Cada valor es un número mayor que 0. Son las horas de cada registro de esfuerzo (US-019) del Sprint, de cualquier historia y de cualquier integrante |

Quién arma la lista de horas (la capa de aplicación, con una consulta de
US-020) toma los registros cuya **fecha** cae dentro de las fechas del Sprint,
como fija RN-019-12: el esfuerzo pertenece al Sprint donde se trabajó, no al
Sprint de la historia. Esta función no lo decide ni lo verifica.

## 3. Salidas esperadas

```
ResumenHoras {
    Estimadas  float64 // storyPointsPlanificados × HorasPorStoryPoint
    Reales     float64 // suma de horasRegistradas
    Desviacion float64 // Reales − Estimadas
}

func CalcularHoras(storyPointsPlanificados int, horasRegistradas []float64) (ResumenHoras, error)
func (r ResumenHoras) DesviacionPorcentual() (porcentaje float64, calculable bool)
```

- Las tres cifras se redondean a **2 decimales**, para no mostrar ruido de
  coma flotante (por ejemplo `7.500000000000001`).
- `DesviacionPorcentual` es `Desviacion / Estimadas × 100`, redondeada a **1
  decimal**, y solo se puede calcular si `Estimadas > 0` (RN-027-4).
- Una desviación **positiva** significa que se trabajó más de lo estimado; una
  **negativa**, que se trabajó menos.

`HorasPorStoryPoint` es la constante `1.5` (RN-027-1). Se exporta para que la
pantalla pueda mostrar la equivalencia que se usó.

## 4. Reglas de negocio

- **RN-027-1** — Las horas estimadas son los Story Points planificados por
  `HorasPorStoryPoint`, la equivalencia de **1 Story Point = 1,5 horas** que el
  equipo usa desde el Sprint 0 (acta del Sprint 0: "16 SP a 1,5 h/SP"). Es una
  constante del dominio y no un dato del proyecto (ver Decisiones).
- **RN-027-2** — Las horas reales son la suma de las horas recibidas. No se
  filtran por historia, por integrante ni por estado de la historia: lo que se
  trabajó, se cuenta.
- **RN-027-3** — La desviación es `Reales − Estimadas`, en horas.
- **RN-027-4** — La desviación porcentual es `Desviacion / Estimadas × 100`. Si
  `Estimadas` es 0 no hay base de comparación: `calculable` es `false`, y no es
  un error.
- **RN-027-5** — Los Story Points negativos se rechazan con
  `ErrStoryPointsNegativos` (el mismo error de US-025: es la misma condición).
- **RN-027-6** — Cada valor de `horasRegistradas` tiene que ser mayor que 0 y
  un número (no `NaN` ni infinito). Uno que no cumple se rechaza con
  `ErrHorasInvalidas`, indicando su posición en la lista, contada desde 1
  (igual que RN-005-14).
- **RN-027-7** — La función valida todo y devuelve todos los errores juntos,
  unidos con `errors.Join`, en este orden fijo: primero el de los Story Points y
  después uno por cada hora inválida, por posición ascendente. Con algún error
  no se devuelve resumen. Cada error unido se reconoce con `errors.Is`.
- **RN-027-8** — No se revalidan los límites de US-019 (hasta 24 horas por
  registro y múltiplos de 0,25): esos datos ya entraron validados, y repetir la
  regla acá la dejaría en dos lugares.

## 5. Restricciones

- **Dominio puro.** `internal/domain/metricas` no importa nada fuera de la
  biblioteca estándar, no lee el reloj ni la base, y reutiliza
  `ErrStoryPointsNegativos` de US-025.
- **No consulta nada.** Qué registros son de un Sprint (por fecha, RN-019-12) y
  cuál es el total de Story Points lo resuelve la capa de aplicación; esta
  historia entrega la función de cálculo, igual que US-025 y US-026. La consulta
  de las horas por Sprint es de US-020.
- **La equivalencia se calibra con datos reales.** Hasta ahora no hubo horas
  registradas, así que 1,5 h por Story Point es una convención sin medición
  (retros de los Sprints 0 a 2). Cuando US-019 esté en uso, esta historia es la
  que muestra cuánto se desvía.
- **Escenarios BDD.** Se agregan al área Métricas de los steps. Alcanzan las
  frases del diccionario para el contexto de Story Points
  (`el sprint "S" con una historia de N story points ...`); se agregan frases
  para las horas (ver Decisiones). Los mensajes se comprueban con
  `el mensaje de error indica "..."`.
- **Catálogo de errores (T-005).** `metricas.ErrHorasInvalidas` va en las exclusiones
  de `internal/adapters/http/catalogo.go`, con el motivo de los otros errores de
  métricas: las horas ya entraron validadas por US-019, así que una inválida es un
  dato del sistema que salió mal y no algo que la persona escribió. El test del
  catálogo falla si no está.
- **Fuera de alcance.** La pantalla, las horas por integrante o por historia
  (eso es US-020 y US-021), el desglose de la desviación y cualquier
  proyección hacia sprints futuros.

## 6. Casos límite

- **CL-027-1** — Sprint sin horas registradas, con Story Points planificados
  (por ejemplo 8): `Estimadas = 12`, `Reales = 0`, `Desviacion = −12`,
  desviación porcentual `−100`. No es un error: es lo que pasa cuando nadie
  cargó horas.
- **CL-027-2** — Sprint sin Story Points planificados (0) pero con horas
  registradas (por ejemplo 4): `Estimadas = 0`, `Reales = 4`, `Desviacion = 4`,
  y la desviación porcentual no es calculable.
- **CL-027-3** — Sprint sin nada (0 Story Points y lista vacía): las tres cifras
  son 0 y la desviación porcentual no es calculable. Sin error.
- **CL-027-4** — Horas reales iguales a las estimadas (5 Story Points y 7,5
  horas): `Desviacion = 0` y desviación porcentual `0`, calculable.
- **CL-027-5** — La desviación porcentual no es un número redondo (5 Story
  Points = 7,5 horas estimadas y 8 reales): `0.5` de desviación y `6.7` de
  porcentaje, redondeado a 1 decimal y no truncado.
- **CL-027-6** — Un solo registro de 0,25 horas (el mínimo de US-019): se suma
  tal cual.
- **CL-027-7** — Muchos registros que suman una cifra con decimales
  (`0.25 + 0.25 + 0.5`): la suma es exacta (`1`), sin ruido de coma flotante.
- **CL-027-8** — Story Points negativos y dos horas inválidas a la vez: se
  devuelven los tres errores unidos, en el orden de RN-027-7. `errors.Is`
  reconoce cada uno.
- **CL-027-9** — Una hora igual a 0 o negativa: se rechaza con
  `ErrHorasInvalidas` indicando su posición. Una lista `nil` es lo mismo que una
  lista vacía (CL-027-1).

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| `storyPointsPlanificados` es negativo | `metricas.ErrStoryPointsNegativos` (el de US-025) | El de US-025 |
| Alguna hora registrada es 0, negativa, `NaN` o infinita | `metricas.ErrHorasInvalidas`, envuelto indicando la posición de la hora, contada desde 1 | "la hora de la posicion N no es un numero mayor a 0" |

Con varios datos inválidos, los errores se devuelven unidos con `errors.Join`
(RN-027-7) y se comparan con `errors.Is`. Ninguno de los dos casos de "no hay
horas" o "no hay Story Points" es un error: son casos límite con resultado
(CL-027-1 a CL-027-3).

## 8. Criterios de aceptación

- **CA-027-1** (normal) — Dado un Sprint con 5 Story Points planificados y 8
  horas registradas, las horas estimadas son 7,5, las reales 8 y la desviación
  0,5, es decir un 6,7 % por encima de lo estimado.
- **CA-027-2** (alternativo) — Dado un Sprint con 8 Story Points planificados y
  9 horas registradas en tres registros, las estimadas son 12, las reales 9 y la
  desviación −3 (un 25 % por debajo de lo estimado).
- **CA-027-3** (límite) — Dado un Sprint con Story Points planificados y ninguna
  hora registrada, las horas reales son 0 y la desviación es la de las estimadas
  en negativo, sin error.
- **CA-027-4** (límite) — Dado un Sprint sin Story Points planificados y con horas
  registradas, la desviación es igual a las horas reales y el porcentaje no se
  puede calcular.
- **CA-027-5** (error) — Una hora igual a 0 o negativa se rechaza indicando su
  posición, y no se devuelve ningún resumen.
- **CA-027-6** (error) — Story Points planificados negativos se rechazan y no se
  devuelve ningún resumen.

---

## Decisiones tomadas y descartadas

- **La equivalencia de 1,5 horas por Story Point es una constante del dominio.**
  Es la convención que el equipo escribió en las actas desde el Sprint 0. Se
  descartó guardarla como configuración del proyecto: ninguna historia pide
  cambiarla por proyecto, y una constante se prueba y se explica sin base de
  datos. Si la medición real la contradice, se cambia la constante en una
  historia propia, con su spec.
- **Las horas estimadas salen de los Story Points del Sprint, no de las horas
  que alguien estima por historia.** Ninguna historia carga horas estimadas: lo
  que se estima es en Story Points (US-013). Se descartó agregar un segundo
  campo de estimación, que duplicaría el trabajo de Planning Poker.
- **Se comparan las horas del Sprint contra los Story Points planificados de
  todo el Sprint**, incluidas las historias no completadas, porque el esfuerzo
  de una historia que no se terminó también se trabajó. Se descartó comparar
  solo lo completado: dejaría afuera el esfuerzo perdido, que es justo el que
  conviene ver.
- **Las horas se asignan al Sprint por la fecha del trabajo** (RN-019-12), no por
  la historia. Es la consecuencia de que el esfuerzo de una historia que vuelve
  al backlog (US-011) se quede en el Sprint donde se trabajó.
- **La desviación se informa en horas y en porcentaje.** Las horas dicen cuánto
  se desvió, y el porcentaje permite comparar un Sprint con otro. Se descartó
  informar solo una de las dos.
- **Sin base de comparación no es un error.** Un Sprint sin Story Points
  planificados existe (recién creado); por eso el porcentaje no calculable es un
  resultado y no un fallo, igual que la velocidad con cero sprints cerrados
  (CL-026-1).
- **Redondeo.** Las cifras en horas a 2 decimales (suficiente para cuartos de
  hora y para ocultar el ruido de coma flotante) y el porcentaje a 1 decimal,
  como la velocidad (RN-026-2).
- **Frases nuevas del diccionario** (se aprueban en el PR de implementación):
  `el sprint "S" con N horas registradas` (contexto: suma un registro de esfuerzo
  al Sprint), `se consultan las horas del sprint "S"`,
  `las horas estimadas son N`, `las horas reales son N`,
  `la desviación de horas es N` y `la desviación porcentual es N`. Las cuatro
  últimas aceptan coma o punto decimal y un número negativo. Para el caso en que
  el porcentaje no se puede calcular, el escenario comprueba
  `la desviación porcentual no se puede calcular`. Todas pertenecen al área
  Métricas.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la
  historia, usando las de US-025 y US-026 como modelo; las decisiones sobre la
  equivalencia y el redondeo se proponen acá y se confirman en la revisión
- Revisado y entendido por: Tomás Romero
