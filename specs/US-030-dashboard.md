# US-030 — Ver un Dashboard con el estado del proyecto y sus métricas principales

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #36 |
| **Escenarios BDD** | `features/US-030-dashboard.feature` |
| **Código** | `internal/app/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| **Última actualización** | 2026-10-27 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero ver un dashboard con el estado del proyecto y sus métricas
> principales, para saber de un vistazo cómo vamos y dónde nos estamos equivocando.

Una página por proyecto (`/proyectos/{id}/dashboard`) que junta, con tarjetas, tablas y dos gráficos, los
cálculos que ya existen en el dominio: los Story Points del sprint (US-025), la velocidad del equipo (US-026), las horas
estimadas, reales y su desviación (US-027), el porcentaje de historias completadas (US-028) y los defectos
detectados y resueltos por sprint (US-029), más el estado general del proyecto (US-004). Cubre los
requerimientos 7 y 8 de la guía (métricas y dashboard).

Esta historia **no calcula nada nuevo**: arma, con un caso de uso de lectura, los resultados de funciones que
ya están en `main`. Los dos gráficos (los Story Points planificados contra los completados de cada sprint
cerrado y el porcentaje de historias completadas) son la «representación gráfica» que pide el requerimiento 8:
se dibujan en el servidor, en SVG, sin librerías ni JavaScript. Los gráficos más elaborados de velocidad
histórica y de evolución de defectos siguen siendo US-032, fuera del alcance del Sprint 5.

Depende de **US-004** (el estado general), de **T-019** (las horas registradas, que son la entrada de las
horas reales) y de **T-021** (la lectura de defectos).

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | `string` | Sí | ID de un proyecto existente. Viene de la URL (`GET /proyectos/{id}/dashboard`); lo verifica el caso de uso |

El caso de uso `ConsultarDashboard` hace **una lectura de cada tipo** y calcula todo con el dominio:

1. el proyecto (y, con US-004, su estado y su avance);
2. los sprints del proyecto con sus historias (la lectura de US-012: `ListarSprintsConItems`);
3. los registros de esfuerzo (la lectura de US-020: `LeerEsfuerzo`);
4. los defectos del proyecto (`ListarDefectosDelProyecto`, de T-021), con el número de sprint de detección y de
   resolución de cada uno.

## 3. Salidas esperadas

```
Dashboard {
    Estado        EstadoGeneral            // el de US-004
    Referencia    *SprintDelDashboard      // nil si el proyecto no tiene sprints
    Velocidad     float64                  // promedio de los sprints cerrados (US-026)
    Cerrados      []SprintCerradoDelDashboard // del más reciente al más antiguo
    Defectos      DefectosDelProyecto
    Graficos      GraficosDelDashboard     // se arman con los mismos datos de las tarjetas
}

GraficosDelDashboard {
    StoryPoints   []BarrasDeSprint         // un grupo por sprint cerrado, del más antiguo al más reciente
    Completadas   *BarraDePorcentaje       // nil si no hay sprint de referencia
}
BarrasDeSprint     { Sprint int; Planificados, Completados float64; AlturaPlanificados, AlturaCompletados float64 }
BarraDePorcentaje  { Sprint int; Porcentaje float64 }
```

Las alturas son porcentajes de 0 a 100 de la barra más alta del gráfico: las calcula el caso de uso, la
plantilla solo las dibuja (RN-030-14).

Cada tarjeta lleva su propio error (`Err`): si un cálculo falla, solo falla esa tarjeta (RN-030-9).

| Tarjeta | Qué muestra |
|---|---|
| **Estado del proyecto** | La tarjeta de US-004: estado, días, avance del backlog y sprint abierto |
| **Sprint** (el abierto o, si no hay, el último cerrado) | Su nombre y estado; los Story Points planificados y completados; el porcentaje de historias completadas; las horas estimadas, reales, la desviación en horas y en porcentaje |
| **Velocidad del equipo** | El promedio de Story Points completados de los sprints cerrados, y la tabla de los sprints cerrados con su nombre, Story Points planificados y completados, defectos detectados y defectos resueltos |
| **Defectos** | Cuántos hay, por severidad (crítica, alta, media, baja) y por estado (abierto, en corrección, resuelto, descartado) |
| **Gráfico de Story Points** | Barras de Story Points planificados y completados por sprint cerrado, del más antiguo al más reciente |
| **Gráfico de historias completadas** | Una barra horizontal con el porcentaje de historias completadas del sprint de referencia |

Un cálculo que no se puede hacer se escribe como texto, no como un 0 engañoso: «Sin sprints cerrados»
(velocidad y gráfico de Story Points), «—» (porcentaje de desviación sin horas estimadas), «Sin sprints» (la tarjeta del sprint) y «Sin
defectos» (la de defectos). Los números se muestran como en T-019: horas con coma decimal y sin ceros de más,
desviación con signo y porcentaje con un decimal.

## 4. Reglas de negocio

- **RN-030-1** — El **sprint de referencia** es el sprint abierto del proyecto; si no hay, el cerrado de
  número más alto; si no hay sprints, no hay tarjeta de sprint.
- **RN-030-2** — Los Story Points planificados y completados de un sprint son los de
  `metricas.CalcularResumenSprint` (US-025), con la regla de `HistoriaAsignada.Completada()`: en un sprint
  cerrado manda el resultado de la fila de la historia y en uno abierto, su estado actual.
- **RN-030-3** — El porcentaje de historias completadas es el de `metricas.CalcularPorcentajeCompletadas`
  (US-028) sobre las historias del sprint, con el mismo criterio de «completada».
- **RN-030-4** — Las **horas reales** de un sprint son las de los registros de esfuerzo cuya fecha cae en el
  rango del sprint (RN-019-12 y `esfuerzo.SprintDeLaFecha`), de cualquier historia e integrante (RN-027-2). Las
  horas estimadas y la desviación son las de `metricas.CalcularHoras` (US-027) con los Story Points
  planificados del sprint.
- **RN-030-5** — La **velocidad** es la de `metricas.CalcularVelocidad` (US-026) con los sprints **cerrados**
  y solo con ellos. Sin sprints cerrados es 0, y se muestra «Sin sprints cerrados».
- **RN-030-6** — Los **defectos por sprint** son los de `metricas.CalcularDefectosDelSprint` (US-029): cada
  defecto se traduce a los **números** de sus sprints de detección y de resolución (con `0` si no se resolvió).
  **Los defectos `descartado` no entran** en la lista que se le pasa a la función (RN-023-13): no son defectos
  que haya que contar.
- **RN-030-7** — La tarjeta de **defectos** cuenta todos los del proyecto por severidad y por estado,
  **incluidos** los descartados (en su estado), así se ve cuántos hubo y qué pasó con cada uno.
- **RN-030-8** — Los sprints cerrados se listan del más reciente al más antiguo (por número descendente) y
  la tabla tiene una fila por sprint cerrado, aunque no tenga defectos ni horas.
- **RN-030-9** — Cada tarjeta se calcula por separado. Un error en una (por ejemplo, Story Points negativos
  guardados) se informa en esa tarjeta, y el resto de la página se muestra con normalidad. Los datos
  corruptos no se esconden ni se convierten en un 0.
- **RN-030-10** — Un proyecto que no existe se rechaza con `ErrProyectoInexistente` sin calcular nada.
- **RN-030-11** — El dashboard es una lectura: no modifica nada. Hace una sola lectura de cada tipo (RN-004-8),
  no una por sprint ni una por historia.
- **RN-030-12** — La página usa el layout, el router y `Responder` existentes. El menú del proyecto suma
  «Dashboard» (sección nueva, activa en la página) y **saca el enlace «Métricas»**, que apuntaba a una pantalla que nunca
  existió (respondía «Página no encontrada»): el dashboard es la pantalla de las métricas. Un texto con HTML se muestra escapado (RN-T011-8).
- **RN-030-13** — El handler solo lee la petición, llama a `ConsultarDashboard` y escribe la respuesta
  (RN-T011-1). Todo el armado y los cálculos viven en el caso de uso y el dominio, no en la plantilla.
- **RN-030-14** — El **gráfico de Story Points** tiene un grupo de dos barras (planificados y completados) por
  sprint cerrado, en orden cronológico (número ascendente; la tabla va al revés, RN-030-8). La escala es común:
  la barra más alta del gráfico mide 100 y las demás son proporcionales (`valor / máximo × 100`). Si el máximo es
  0, todas las barras miden 0 (sin dividir por cero). Los valores salen de los mismos resultados de
  `CalcularResumenSprint` que la tarjeta de velocidad (RN-030-2); no se recalculan.
- **RN-030-15** — El **gráfico de historias completadas** dibuja el porcentaje de RN-030-3 del sprint de
  referencia como una barra de 0 a 100, con el valor escrito al lado. Sin sprint de referencia no se dibuja.
- **RN-030-16** — Cada gráfico es un SVG con `role="img"`, un título y una descripción con los mismos números
  que dibuja (por ejemplo, «Sprint 1: 10 planificados, 8 completados»), y cada barra lleva su valor escrito.
  Los números nunca se leen solo por la altura. Un gráfico no depende del color: planificados y completados se
  distinguen también por una leyenda y por su posición.
- **RN-030-17** — Un gráfico sigue la suerte de su tarjeta (RN-030-9): si el cálculo de la que lo alimenta
  falla, en su lugar se informa el error y no se dibuja nada. Con datos vacíos se escribe el texto («Sin
  sprints cerrados», «Sin sprints») y no se dibuja un gráfico vacío.

## 5. Restricciones

- **Sin cálculos nuevos.** No se toca `internal/domain/metricas`: todas las cifras salen de sus funciones.
- **Casos de uso y puertos.** `ConsultarDashboard` recibe un puerto con las lecturas del punto 2. Postgres y
  memoria ya implementan las tres primeras, y también `ListarDefectosDelProyecto(ctx, proyectoID)
  ([]DefectoListado, error)` de T-021 y `LeerEsfuerzo` de US-020. `cmd/server` junta los
  repositorios como hizo con `lecturaDeProyecto`.
- **Reloj.** El estado general usa el de US-004 (`time.Now` en `cmd/server`; uno fijo en los tests y escenarios).
- **Sin migración ni errores nuevos.**
- **Sin JavaScript ni librerías de gráficos:** HTML con Tailwind y SVG escrito en la plantilla templ, como el
  resto. Sin nuevas dependencias de Go ni de npm.
- **Escenarios BDD.** Van en el área Métricas de los steps y reutilizan las frases de contexto de sprints,
  historias, esfuerzo y defectos; se agregan frases de consulta (ver Decisiones).
- **Fuera de alcance.** Los gráficos de velocidad histórica y de evolución de defectos (US-032), el burndown (US-031), comparar proyectos, filtrar por fechas, exportar
  (es el reporte, US-033) y el desglose de horas por integrante (T-019).

## 6. Casos límite

- **CL-030-1** (RN-030-1) — Con un sprint abierto y otros cerrados: la tarjeta del sprint es la del abierto.
- **CL-030-2** (RN-030-1) — Con todos los sprints cerrados: la tarjeta es la del de número más alto.
- **CL-030-3** (RN-030-1) — Sin sprints: no hay tarjeta de sprint («Sin sprints») y la velocidad dice
  «Sin sprints cerrados»; el resto del dashboard se muestra.
- **CL-030-4** (RN-030-4) — Sprint sin horas registradas: horas reales 0, desviación igual a las estimadas en
  negativo y porcentaje -100 (CL-027-1). Sin Story Points planificados y con horas: el porcentaje de
  desviación es «—» (CL-027-2).
- **CL-030-5** (RN-030-4) — Un registro de esfuerzo con una fecha entre dos sprints, o antes del primero: no
  entra en las horas de ningún sprint.
- **CL-030-6** (RN-030-5) — Con un solo sprint cerrado, la velocidad es la de ese sprint.
- **CL-030-7** (RN-030-6) — Un defecto detectado en el Sprint 2 y resuelto en el 4: cuenta como detectado en el
  2 y resuelto en el 4, y en ningún otro. Uno detectado y resuelto en el mismo sprint cuenta en las dos cifras.
- **CL-030-8** (RN-030-6) — Un defecto `descartado` no cuenta como detectado ni como resuelto en ningún sprint,
  pero sí aparece en la tarjeta de defectos, en su estado.
- **CL-030-9** (RN-030-7) — Sin defectos: «Sin defectos» y los sprints cerrados con 0 y 0.
- **CL-030-10** (RN-030-9) — Un sprint con Story Points inconsistentes (completados mayores que planificados):
  la tarjeta de velocidad informa el error y las demás se muestran. *Ninguna operación de la aplicación lo
  produce:* los completados salen de las mismas historias que los planificados, y los Story Points no pueden ser
  negativos. La regla queda en el código (cada tarjeta guarda su error) y la prueba el dominio de `metricas`
  (US-025 y US-026); no tiene escenario ni test de pantalla.
- **CL-030-11** (RN-030-10) — Proyecto inexistente o con ID mal formado: `ErrProyectoInexistente`, y la página
  responde 404.
- **CL-030-12** (RN-030-2) — Una historia no completada en el Sprint 1 y completada después en el Sprint 2: el
  Sprint 1 sigue sin contarla como completada (DEF-007).
- **CL-030-13** (RN-030-14) — Sin sprints cerrados: no hay gráfico de Story Points, se escribe «Sin sprints
  cerrados».
- **CL-030-14** (RN-030-14) — Sprints cerrados sin Story Points planificados ni completados (todo 0): las
  barras miden 0 y no hay división por cero.
- **CL-030-15** (RN-030-14) — Dos sprints con 10 y 5 Story Points planificados y 8 y 5 completados: la barra
  de 10 mide 100, la de 8 mide 80 y las de 5 miden 50. La escala es una para todo el gráfico, no una por sprint.
- **CL-030-16** (RN-030-15) — Un sprint de referencia con 0 % de historias completadas, o con 100 %: la barra
  mide 0 o 100 y el valor está escrito. Un sprint sin historias da 0 % (CL-028-1): la barra mide 0.
- **CL-030-17** (RN-030-17) — Story Points inconsistentes en un sprint cerrado (CL-030-10, que la aplicación no
  puede producir): ni la tarjeta de velocidad ni el gráfico de Story Points se dibujan, ambos informan el error;
  el gráfico de historias completadas y el resto de las tarjetas se muestran.

## 7. Condiciones de error

| Situación | Error devuelto | Respuesta de la página |
|---|---|---|
| El proyecto no existe, o el ID está mal formado | `app.ErrProyectoInexistente`, envuelto con el ID | 404 con layout |
| Un cálculo de `metricas` rechaza sus datos | El error de `metricas`, envuelto con el sprint | La tarjeta informa el error; el resto se muestra |
| Falla inesperada de la persistencia | Error envuelto con contexto | 500 con el mensaje fijo de T-005 |

Que no haya sprints, horas, historias o defectos **no es un error**: son casos límite con resultado.

## 8. Criterios de aceptación

- **CA-030-1** (normal) — Con un sprint abierto con tres historias (una completada de 5 Story Points) y 8 horas
  registradas con fecha dentro de él, el dashboard muestra 5 Story Points completados, el porcentaje de
  historias completadas, las horas estimadas, las reales (8) y la desviación.
- **CA-030-2** (normal) — Con dos sprints cerrados de 8 y 5 Story Points completados, muestra la velocidad
  (6,5) y una fila por sprint cerrado, el más reciente primero, con sus Story Points y sus defectos
  detectados y resueltos.
- **CA-030-3** (alternativo) — Muestra los defectos del proyecto por severidad y por estado, y la tarjeta del
  estado general del proyecto.
- **CA-030-4** (límite) — Un proyecto nuevo, sin sprints, horas ni defectos, muestra todo en 0 y los textos
  «Sin sprints cerrados», «Sin sprints» y «Sin defectos», sin error.
- **CA-030-5** (error) — Un proyecto inexistente se rechaza con `ErrProyectoInexistente` y la página responde
  404; si un cálculo falla por datos inválidos, solo su tarjeta informa el error.
- **CA-030-6** (normal) — Con dos sprints cerrados de 10 y 5 Story Points planificados y 8 y 5 completados, el
  gráfico de Story Points tiene dos grupos de barras en orden cronológico, con el valor escrito en cada una y las
  alturas 100, 80, 50 y 50; el gráfico de historias completadas dibuja el porcentaje del sprint de referencia.
  Ambos gráficos tienen su descripción con los mismos números.
- **CA-030-7** (límite) — Un proyecto sin sprints no dibuja ningún gráfico y escribe «Sin sprints cerrados» y
  «Sin sprints» en su lugar, sin error.

---

## Decisiones tomadas y descartadas

- **Un caso de uso de lectura y no cálculos en la pantalla.** Todo se arma en `internal/app` con las funciones de
  `metricas`: la plantilla solo pinta. Se descartó repetir los cálculos en el handler: quedarían dos lugares
  para cada regla.
- **Una sola página con tarjetas, tablas y dos gráficos simples.** La primera versión de esta spec no tenía
  gráficos porque decía que la guía no los exigía; la revisión de Vergara lo corrigió: el requerimiento 8 pide un
  dashboard «con representaciones gráficas» (ya lo dice el ADR 0001). Se eligieron los dos que se pueden hacer
  con datos que ya existen: Story Points planificados contra completados por sprint cerrado y porcentaje de
  historias completadas. Los gráficos de velocidad histórica y de defectos con más detalle quedan en US-032.
- **Gráficos en SVG escritos en la plantilla, con las medidas calculadas en el caso de uso.** No hay librería ni
  JavaScript: se prueban con el mismo test de HTML que las tarjetas (las alturas son datos, no píxeles que se
  miden en un navegador). Se descartó una librería de gráficos: es una dependencia más para dos gráficos de barras
  y rompe la regla de no usar JavaScript. Se descartaron las barras con `div` y `style`: el SVG permite el título,
  la descripción y el rol para lectores de pantalla.
- **Se saca el enlace «Métricas» del menú en vez de dejarlo junto a «Dashboard».** Venía de T-011 como un lugar reservado y
  nunca tuvo ruta: en la demo habría quedado un enlace que da 404 al lado del que sí funciona. Se descartó que «Métricas» fuera
  el nombre de la pestaña del dashboard: la spec y el layout ya lo llaman «Dashboard», como el requerimiento 8 de la guía.
- **Cada tarjeta falla por separado.** Un dato corrupto en un sprint no tiene que dejar a todo el equipo sin
  dashboard; tampoco se esconde. Se descartó un error único para toda la página.
- **El sprint de referencia es el abierto, o el último cerrado.** Es el que importa en la Daily y en la Review; el
  historial está en la tabla de sprints cerrados.
- **Los descartados no cuentan** en los defectos por sprint, como acordó la revisión de US-023: un defecto que no
  se va a corregir no es un defecto detectado. Se descartó contarlos con `SprintResolucion = 0`: seguirían
  inflando la cifra.
- **Las horas de un sprint salen de la fecha del trabajo**, no de la historia (RN-019-12), para que el esfuerzo de
  una historia que pasa de sprint se quede en el sprint donde se trabajó.
- **Frases nuevas del diccionario** (se aprueban en el PR de implementación), área Métricas:
  `se consulta el dashboard del proyecto "P"`, `la tarjeta del sprint es el "Sprint N"`,
  `la velocidad del dashboard es N`, `el dashboard no tiene sprint de referencia`,
  `el gráfico de Story Points tiene N grupos de barras`, `la barra de "completados" del "Sprint N" mide N`,
  `el dashboard lista N sprints cerrados`, `el sprint cerrado "Sprint N" figura con N defectos detectados y M resueltos`
  y `el dashboard cuenta N defectos`. Las de horas y Story Points reutilizan las de US-027 y US-025 aplicadas a la
  tarjeta del sprint.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la historia y de las specs de
  US-025 a US-029, US-004 y US-012; el criterio de qué sprint muestra la tarjeta y de las tarjetas independientes se
  propone acá y se confirma en la revisión; los gráficos se agregaron después del pedido de cambios de la
  revisión de Vergara (requerimiento 8)
- Revisado y entendido por: Tomás Romero
