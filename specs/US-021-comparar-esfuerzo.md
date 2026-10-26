# US-021 — Comparar el esfuerzo estimado contra el esfuerzo real

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #24 |
| **Escenarios BDD** | `features/US-021-comparar-esfuerzo.feature` |
| **Código** | `internal/domain/metricas/`, `internal/app/comparar_esfuerzo.go` |
| **Última actualización** | 2026-10-25 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como Agile Enabler, quiero comparar el esfuerzo estimado de cada historia contra
> las horas que realmente se trabajaron, para ver dónde nos equivocamos al estimar
> y calibrar el próximo Planning.

US-027 hace esta comparación para un **Sprint entero**. Esta historia la hace
**historia por historia**: para cada una, las horas que implica su estimación
(Story Points × 1,5 horas), las horas que se registraron y la desviación entre
ambas. Muestra primero las historias donde más nos equivocamos, y un total de las
que se pueden comparar.

Depende de **US-013** y **US-018** (la estimación de la historia, venga de donde
venga), de **US-019** (las horas) y de **US-027** (la equivalencia y el cálculo de
la desviación, que esta historia reutiliza sin reescribir).

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | `string` | Sí | ID de un proyecto existente (viene de la URL, no de un formulario) |

No hay filtros: la comparación es del proyecto entero. El corte por Sprint es de
US-027 (ver Decisiones).

La capa de aplicación verifica que el proyecto exista (`ErrProyectoInexistente`) y
hace **una sola lectura** a la persistencia, así que las historias, sus
estimaciones y sus horas son de la misma foto:

```
HistoriaConHoras {
    Item  backlog.ItemBacklog // la historia, con su estimación y su estado actuales
    Horas []float64           // las horas de todos sus registros de esfuerzo (US-019)
}

LeerHistoriasConHoras(ctx, proyectoID) ([]HistoriaConHoras, error) // todas las del proyecto, sin ordenar
```

El dominio no recibe la historia entera, sino lo que necesita para calcular:

```
HistoriaParaComparar {
    ItemID      string
    Titulo      string
    StoryPoints int       // 0 si la historia no está estimada
    Completada  bool
    Horas       []float64 // cada valor mayor que 0 (como en US-027)
}
```

## 3. Salidas esperadas

```
ComparacionHistoria {
    ItemID, Titulo   string
    StoryPoints      int
    Estimada         bool        // false si la historia no tiene estimación
    Completada       bool
    Registros        int         // cantidad de registros de esfuerzo
    Horas            ResumenHoras // Estimadas, Reales y Desviacion, el tipo de US-027
}

ComparacionProyecto {
    Historias         []ComparacionHistoria
    Total             ResumenHoras // solo las historias comparables (RN-021-4)
    Comparables       int
    HorasSinComparar  float64      // horas reales de las historias que no son comparables
}

func CompararHistorias(historias []HistoriaParaComparar) (ComparacionProyecto, error)
func (c ComparacionHistoria) DesviacionPorcentual() (porcentaje float64, calculable bool)
```

- Las horas se redondean a **2 decimales** y el porcentaje a **1 decimal**, con las
  mismas reglas que US-027.
- `DesviacionPorcentual` de una historia es la de `ResumenHoras`: no es calculable
  si la historia no tiene estimación.
- Caso de uso: `NuevoCompararEsfuerzo(repo).Ejecutar(ctx, proyectoID)
  (ComparacionProyecto, error)`.

## 4. Reglas de negocio

- **RN-021-1** (dominio) — Para cada historia, las horas estimadas son sus Story
  Points por `HorasPorStoryPoint` (1,5), las reales son la suma de sus horas y la
  desviación es `Reales − Estimadas`. Es el mismo cálculo de US-027 (se llama a
  `CalcularHoras`), así que las dos historias no pueden dar cifras distintas.
- **RN-021-2** (dominio) — Las horas reales de una historia son **todas** las de
  sus registros, sin importar la fecha, el integrante ni el Sprint donde se
  trabajaron: la comparación es contra la historia entera, no contra un período.
- **RN-021-3** (dominio) — Una historia **aparece** en la comparación si tiene al
  menos un registro de esfuerzo, o si está completada **y** estimada. Las demás no
  aparecen: no hay nada para comparar (una historia pendiente sin horas, o
  completada sin estimar y sin horas).
- **RN-021-4** (dominio) — Una historia es **comparable** si está completada y
  estimada: solo entonces la estimación y el esfuerzo son del mismo trabajo
  terminado. `Total` suma solo las comparables: sus Story Points por 1,5 como
  estimadas, todas sus horas como reales (se calcula con `CalcularHoras`).
  `Comparables` es cuántas son.
- **RN-021-5** (dominio) — `HorasSinComparar` es la suma de las horas de las
  historias que aparecen pero no son comparables (en curso, o sin estimar). Así se
  ve cuánto esfuerzo quedó fuera del total sin esconderlo.
- **RN-021-6** (dominio) — Una historia sin estimación aparece con `Estimada =
  false`, estimadas 0 y desviación igual a sus horas reales; su porcentaje no es
  calculable. No es un error.
- **RN-021-7** (dominio) — Orden de `Historias`: por el **valor absoluto de la
  desviación**, de mayor a menor (primero donde más nos equivocamos, para arriba o
  para abajo). Con el mismo valor absoluto, por título ascendente sin distinguir
  mayúsculas y, si aun así empatan, por ID. El orden es total: el mismo dato sale
  siempre igual.
- **RN-021-8** (dominio) — La estimación que se usa es la **actual** de la
  historia: si se re-estimó (RN-013-3) después de cargar horas, vale la nueva.
- **RN-021-9** (dominio) — Los Story Points negativos o una hora que no sea un
  número mayor que 0 son datos del sistema que salieron mal (ya entraron
  validados por US-013 y US-019). La función devuelve esos errores de
  `CalcularHoras` (`ErrStoryPointsNegativos`, `ErrHorasInvalidas`) unidos con
  `errors.Join`, cada uno envuelto indicando la posición de la historia, contada
  desde 1. Con algún error no se devuelve comparación.
- **RN-021-10** (aplicación) — Sin historias que aparezcan (o sin historias en el
  proyecto) la comparación está vacía, con todo en 0 y `Comparables = 0`. No es un
  error.
- **RN-021-11** (aplicación) — La consulta no modifica nada.

## 5. Restricciones

- **Dominio puro.** `internal/domain/metricas` no importa nada fuera de la
  biblioteca estándar ni de sus propios archivos. Recibe `HistoriaParaComparar` y
  no el `ItemBacklog`: así `metricas` no depende de `backlog`, igual que con
  `ItemDelSprint` de US-025.
- **La capa de aplicación traduce.** Arma cada `HistoriaParaComparar` desde el
  `ItemBacklog` (Story Points con `StoryPoints.Valor()`, completada con el estado)
  y las horas desde los registros.
- **No usa el código de US-020.** Es otra agregación (por historia, de toda la
  vida de la historia, sin deducir el Sprint por fecha). Las dos historias solo
  comparten la tabla `registros_esfuerzo` de US-019, así que se pueden
  implementar en cualquier orden.
- **Puerto propio** `RepositorioCompararEsfuerzo` con `ExisteProyecto` y
  `LeerHistoriasConHoras`. Lo implementan la base en memoria y Postgres con el
  mismo resultado; la lectura de Postgres va en una sola transacción de solo
  lectura (con `LEFT JOIN` de historias a registros, no una consulta por
  historia).
- **Sin migración.**
- **Catálogo de errores (T-005).** No se agrega ningún error: se reutilizan
  `ErrProyectoInexistente` y los de `metricas`, que ya están en las exclusiones del
  catálogo por ser datos del sistema.
- **Escenarios BDD.** Van en el área Métricas de los steps. Los contextos reutilizan
  las frases de proyecto, historia, estimación y completar; se agregan frases de
  horas por historia y de consulta (ver Decisiones).
- **Fuera de alcance.** La pantalla (se especifica aparte), la comparación por
  Sprint (US-027), el desglose por integrante (US-020), las proyecciones y
  exportar.

## 6. Casos límite

- **CL-021-1** — Historia completada de 5 Story Points con 8 horas registradas:
  estimadas 7,5, reales 8, desviación 0,5 y porcentaje 6,7.
- **CL-021-2** — Historia completada de 3 Story Points con 4,5 horas: desviación 0 y
  porcentaje 0, calculable.
- **CL-021-3** — Historia completada y estimada, sin ninguna hora (8 Story Points):
  aparece con estimadas 12, reales 0, desviación −12 y porcentaje −100. Es la
  señal de que nadie cargó las horas.
- **CL-021-4** — Historia sin estimar con horas (4 horas): aparece con
  `Estimada = false`, estimadas 0, reales 4, desviación 4 y porcentaje no
  calculable. No suma al total; sus 4 horas van a `HorasSinComparar`.
- **CL-021-5** — Historia pendiente, estimada y sin horas: no aparece (RN-021-3).
- **CL-021-6** — Historia completada y sin estimar, sin horas: no aparece.
- **CL-021-7** — Historia en curso (no completada) con horas: aparece con
  `Completada = false`; no suma al total y sus horas van a `HorasSinComparar`.
- **CL-021-8** — Historia a la que se le cargó esfuerzo con 3 Story Points y después
  se re-estimó a 5: se compara contra 5 (RN-021-8), es decir 7,5 horas estimadas.
- **CL-021-9** — Horas de una historia de varias fechas, integrantes y Sprints: se
  suman todas (RN-021-2).
- **CL-021-10** — Dos historias completadas: una de 8 Story Points con 9 horas
  (desviación −3) y otra de 5 con 8 horas (desviación 0,5). La primera sale antes
  (3 > 0,5).
- **CL-021-11** — El total de CL-021-10: estimadas 19,5, reales 17, desviación −2,5
  y porcentaje −12,8, con `Comparables = 2`.
- **CL-021-12** — Dos historias con la misma desviación en valor absoluto (una +1 y
  otra −1): salen ordenadas por título.
- **CL-021-13** — Proyecto sin historias, o sin ninguna que aparezca: comparación
  vacía, total en 0 y porcentaje del total no calculable. Sin error.
- **CL-021-14** — Hay historias que aparecen pero ninguna es comparable: `Total` en
  0, `Comparables = 0` y todo su esfuerzo en `HorasSinComparar`.
- **CL-021-15** — Un Story Point negativo y dos horas inválidas en historias
  distintas a la vez: se devuelven los tres errores unidos, con la posición de cada
  historia, y no se devuelve comparación.
- **CL-021-16** — Lista `nil` de historias: es lo mismo que una lista vacía
  (CL-021-13).
- **CL-021-17** — Horas con decimales (`0,25 + 0,25 + 0,5`): el total de la historia
  es exactamente `1`, sin ruido de coma flotante.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| El proyecto no existe | `app.ErrProyectoInexistente` | "el proyecto no existe" |
| Una historia tiene Story Points negativos | `metricas.ErrStoryPointsNegativos`, envuelto indicando la posición de la historia | El de US-025 |
| Una historia tiene una hora que no es un número mayor que 0 | `metricas.ErrHorasInvalidas`, envuelto indicando la posición de la historia | El de US-027 |

Con varios datos inválidos, los errores se devuelven unidos con `errors.Join`
(RN-021-9) y se comparan con `errors.Is`. Los dos últimos no los provoca quien usa
la aplicación: son datos del sistema que salieron mal. Que no haya historias, horas
o estimaciones no es un error (CL-021-3 a CL-021-6, CL-021-13 y CL-021-14).

## 8. Criterios de aceptación

- **CA-021-1** (normal) — Dada una historia completada de 5 Story Points con 8 horas
  registradas, la comparación muestra 7,5 horas estimadas, 8 reales y una
  desviación de 0,5 (6,7 % por encima).
- **CA-021-2** (normal) — Dada una historia completada de 8 Story Points con 9
  horas, las estimadas son 12, las reales 9 y la desviación −3 (25 % por debajo).
- **CA-021-3** (alternativo) — Con dos historias completadas, la que más se desvió
  (en horas, para arriba o para abajo) sale primero, y el total suma las dos.
- **CA-021-4** (límite) — Una historia completada y estimada sin horas aparece con
  desviación igual a sus estimadas en negativo.
- **CA-021-5** (límite) — Una historia sin estimar con horas aparece sin porcentaje, no
  suma al total y sus horas quedan en las horas sin comparar.
- **CA-021-6** (límite) — Una historia pendiente, sin horas, no aparece; una en curso
  con horas aparece pero no suma al total.
- **CA-021-7** (límite) — Un proyecto sin historias da una comparación vacía, con
  todo en 0 y sin error.
- **CA-021-8** (error) — Un proyecto inexistente se rechaza con
  `ErrProyectoInexistente`.
- **CA-021-9** (error) — Story Points negativos o una hora inválida en una historia
  se rechazan indicando su posición, y no se devuelve ninguna comparación.

---

## Decisiones tomadas y descartadas

- **Comparar historia por historia y no por Sprint.** El Sprint ya lo hace US-027.
  Esta historia responde otra pregunta: *qué* historias estimamos mal, que es lo
  que sirve para calibrar el Planning Poker.
- **Se reutiliza `CalcularHoras` de US-027 para cada historia.** Se descartó un
  segundo cálculo con las mismas fórmulas: dejaría la equivalencia de 1,5 horas
  por Story Point y el redondeo escritos en dos lugares.
- **Solo las historias completadas y estimadas suman al total.** Una historia en
  curso tiene medio esfuerzo contra la estimación entera, y compararlas mostraría
  una desviación negativa falsa. Se descartó sumar todo: el total diría que
  «trabajamos menos de lo estimado» cuando en realidad algo todavía no terminó. El
  esfuerzo que queda afuera no se esconde: sale en `HorasSinComparar`.
- **Las historias en curso y sin estimar aparecen igual** (si tienen horas). Se
  descartó ocultarlas: es justo el esfuerzo que se escapa de la comparación y
  conviene verlo.
- **No hay filtro por Sprint.** Se evaluó (historias asignadas al Sprint, horas de
  las fechas del Sprint) y se descartó: las horas de una historia se reparten en
  varios Sprints cuando no se completa (US-011), y comparar la estimación entera
  contra un pedazo del esfuerzo daría una desviación falsa. El corte por Sprint
  correcto es el de US-027, que compara Story Points y horas del mismo Sprint.
- **El orden por desviación absoluta** pone arriba las historias donde más nos
  equivocamos en cualquier dirección. Se descartó ordenar por porcentaje: una
  historia de 1 Story Point con 1 hora de más tendría 67 % y taparía a una de 13
  con 8 horas de más.
- **Una sola lectura con las historias y sus horas.** Se descartó que el caso de uso
  pida las horas de cada historia por separado (N+1 consultas) y que reuse la
  consulta de US-020: son agregaciones distintas y así las dos historias quedan
  independientes.
- **Frases nuevas del diccionario** (se aprueban en el PR de implementación), área
  Métricas: `se compara el esfuerzo del proyecto "P"`,
  `la comparación tiene N historias`,
  `la primera historia de la comparación es "H"`,
  `la historia "H" tiene N horas estimadas y M reales en la comparación`,
  `la desviación de la historia "H" en la comparación es N`,
  `la desviación porcentual de la historia "H" en la comparación es N`,
  `la desviación porcentual de la historia "H" en la comparación no se puede
  calcular`, `el total comparable tiene N historias` y
  `las horas sin comparar son N`. Los números aceptan coma o punto decimal y un
  signo negativo. Después de comparar, las frases de US-027 (`las horas estimadas
  son N`, `las horas reales son N`, `la desviación de horas es N`, `la desviación
  porcentual es N` y `la desviación porcentual no se puede calcular`) se refieren
  al **total de las historias comparables**.
- **El contexto usa las frases que ya existen.** La primera versión de esta spec
  proponía una frase de contexto `la historia "H" con N horas registradas` que sumaba
  horas sin pasar por US-019. No hizo falta: las historias, su estimación, su
  completado y las horas se arman con las frases de proyecto, historia, sprint y
  `el integrante "X" ya registró N horas ...`, que recorren los mismos casos de uso
  que la aplicación. El caso de datos inválidos (CA-021-9) no tiene escenario porque
  ningún flujo puede guardar horas o Story Points negativos: lo cubren los tests del
  dominio.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la
  historia y de las de US-013, US-018, US-019, US-020 y US-027; las decisiones
  sobre qué es comparable, el orden y el corte por Sprint se proponen acá y se
  confirman en la revisión
- Revisado y entendido por: Tomás Romero
