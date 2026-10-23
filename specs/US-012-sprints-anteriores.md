# US-012 — Consultar los Sprints anteriores y su resultado

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás |
| **Issue** | #20 |
| **Escenarios BDD** | `features/US-012-sprints-anteriores.feature` |
| **Código** | caso de uso en `internal/app/`, `internal/adapters/http/` y `web/templates/` (la pantalla), `cmd/server/`. Sin cambios en el dominio ni migración |
| **Última actualización** | 2026-10-24 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero consultar los Sprints anteriores y su
> resultado, para aprender de lo que pasó y planificar mejor el próximo.

Mostrar el historial de un proyecto: los sprints que ya se cerraron, del más
reciente al más antiguo, con lo que cada uno planificó, lo que completó y qué
historias quedaron sin completar, más la velocidad del equipo. Cubre la parte de
consulta del requerimiento 3 de la guía (cerrar un sprint y consultar los anteriores).

Se apoya en lo que ya hay: **US-011** deja el resultado de cada historia en su
sprint, **DEF-007** hace que las lecturas lo respeten, **US-025** calcula los Story
Points del sprint y **US-026** la velocidad. T-015 ya lista todos los sprints del
proyecto; esta historia define qué es el *resultado* de un sprint cerrado y agrega la
vista de historial.

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. Viene de la URL; lo verifica el caso de uso |

La pantalla es `GET /proyectos/{id}/sprints/anteriores`. No hay formulario ni
filtros: son todos los sprints cerrados del proyecto.

## 3. Salidas esperadas

El caso de uso `ConsultarSprintsAnteriores` devuelve:

```
SprintsAnteriores {
    Sprints   []SprintAnterior // los sprints cerrados, del más reciente al más antiguo
    Velocidad float64          // promedio de Story Points completados por sprint (US-026)
}

SprintAnterior {
    Sprint               sprint.Sprint
    Planificados         int // Story Points de todas las historias que tuvo
    Completados          int // Story Points de las que se completaron en él
    HistoriasCompletadas int
    HistoriasNoCompletadas int
}
```

**Pantalla** (con el layout de T-004, o solo el fragmento si lo pide HTMX): título
"Sprints anteriores", la velocidad del equipo, y una tabla con una fila por sprint:
nombre (enlace a su detalle), Sprint Goal, fechas, Story Points planificados y
completados, e historias completadas y no completadas. Sin sprints cerrados:
"Todavía no hay sprints cerrados" y sin velocidad. La lista de sprints (T-015) enlaza
a esta pantalla.

## 4. Reglas de negocio

- **RN-012-1** (aplicación) — El proyecto tiene que existir. Si no, `ErrProyectoInexistente`.
  Un ID mal formado es un proyecto inexistente.
- **RN-012-2** (aplicación) — Solo cuentan los sprints **cerrados**. El sprint abierto no
  es "anterior": no aparece ni cuenta para la velocidad.
- **RN-012-3** (aplicación) — Los sprints salen por número descendente (el más reciente
  primero), como en T-015 (RN-T015-3). El orden lo pone el caso de uso, no la base.
- **RN-012-4** (aplicación) — El resultado de cada historia en un sprint cerrado es el
  que quedó en **su fila** de ese sprint (RN-011-5, RN-T015-13): `completado` o
  `no_completado`. No se mira el estado actual de la historia: una que no se completó,
  volvió al backlog y se completó después en otro sprint sigue siendo "no completada"
  en el primero.
- **RN-012-5** (aplicación) — `Planificados` y `Completados` salen de
  `metricas.CalcularResumenSprint` (US-025), y `Velocidad` de `metricas.CalcularVelocidad`
  (US-026), calculada solo con los sprints de esta lista. No se vuelven a calcular acá.
- **RN-012-6** (aplicación) — Las historias completadas y no completadas se cuentan por
  historia, no por Story Points: la suma de las dos es la cantidad de historias del
  sprint. Una historia sin estimar cuenta como historia y como 0 Story Points.
- **RN-012-7** (aplicación) — Sin sprints cerrados, el resultado es una lista vacía y
  velocidad 0, y no es un error (CL-026-1).
- **RN-012-8** (aplicación) — Es una lectura: no modifica nada y no abre transacción.

## 5. Restricciones

- **Sin reglas de dominio nuevas ni puertos nuevos.** El caso de uso usa el puerto
  `RepositorioListarSprints` de T-015 (`ExisteProyecto` y `ListarSprintsConItems`) y
  filtra los sprints cerrados. `internal/domain` no se toca.
- **Depende de DEF-007** (las lecturas traen el resultado de la fila) y de T-015 y
  US-011, que ya están en `main` o en revisión. No se puede implementar antes de que
  DEF-007 esté mergeado.
- **No usa US-028** (porcentaje de historias completadas): se muestran las cantidades.
  Cuando US-028 esté, la pantalla puede sumar el porcentaje en una historia propia.
- **Catálogo de errores (T-005).** No hay errores nuevos: `ErrProyectoInexistente` ya está
  registrado como `no_encontrado`. Un `ErrResumenInconsistente` de `metricas` sería un dato
  corrupto y sale como error interno, como ya está en las exclusiones.
- **Frases BDD.** Se suman al área Sprints de los steps, que implementa la base en memoria
  con el mismo caso de uso. Interfaz: `SprintsAnteriores() (SprintsAnteriores, error)`
  sobre el proyecto actual del escenario. Las frases, que se aprueban con el grupo antes de
  implementar:

  | Frase | Expresión |
  |---|---|
  | `Cuando se consultan los sprints anteriores` | `^se consultan los sprints anteriores$` |
  | `Entonces los sprints anteriores son:` + tabla `sprint \| planificados \| completados \| historias completadas \| historias no completadas` | `^los sprints anteriores son:$` |
  | `Entonces no hay sprints anteriores` | `^no hay sprints anteriores$` |
  | `Y la velocidad de los sprints anteriores es 25` | `^la velocidad de los sprints anteriores es (\d+(?:[.,]\d+)?)$` |

  La tabla se compara en orden y completa: el escenario también comprueba el orden de la
  lista (RN-012-3). El resto del contexto sale de frases que ya existen (`un sprint ... con el
  objetivo ...`, `la historia ... asignada al sprint ...`, `la historia ... completada`,
  `el sprint ... cerrado`).
- **Nombres en el código.** El caso de uso es `ConsultarSprintsAnteriores` y devuelve
  `SprintsAnteriores{Sprints []SprintAnterior, Velocidad}`; usa el puerto `RepositorioListarSprints`
  de T-015 sin agregarle nada. La pantalla está en `Pantallas.sprintsAnteriores`, con las plantillas
  `SprintsAnteriores` y la fila `FilaSprintAnterior`. La velocidad se escribe con coma decimal y sin
  ceros de más ("6,5", "8"). La lista de sprints (T-015) enlaza con "Ver sprints anteriores", y el
  historial con "Ver todos los sprints".
- La pantalla es de solo lectura y no tiene formularios; responde la página o solo el fragmento
  si lo pide HTMX, igual que las de T-011 y T-015.
- El diseño visual es funcional (Tailwind); se mejora después.
- **Fuera de alcance.** Comparar dos sprints, gráficos de velocidad (US-032), exportar el
  historial (US-033/034), el porcentaje de historias completadas (US-028), las horas del
  sprint (US-027) y filtrar o paginar la lista.

## 6. Casos límite

- **CL-012-1** — Proyecto sin sprints: lista vacía, velocidad 0 y el mensaje "Todavía no hay
  sprints cerrados". Sin error.
- **CL-012-2** — Solo hay un sprint abierto: no aparece, la lista está vacía y la velocidad
  es 0.
- **CL-012-3** — Sprint cerrado sin historias: aparece con 0 planificados, 0 completados, 0
  historias completadas y 0 no completadas.
- **CL-012-4** — Sprint cerrado con todas las historias completadas: 0 no completadas, y
  los completados son iguales a los planificados.
- **CL-012-5** — Sprint cerrado sin ninguna historia completada: 0 completados y todas las
  historias como no completadas.
- **CL-012-6** — Una historia no completada en el sprint 1 que se asigna al sprint 2 y se
  completa allí: el sprint 1 la sigue contando como no completada (RN-012-4) y el sprint 2,
  si ya está cerrado, como completada.
- **CL-012-7** — Dos sprints cerrados que completaron 30 y 33 Story Points: la velocidad es
  31,5 (US-026).
- **CL-012-8** — Un sprint cerrado y uno abierto: la velocidad solo cuenta el cerrado.
- **CL-012-9** — Una historia sin estimar en un sprint cerrado: cuenta como historia y como
  0 Story Points (RN-012-6).
- **CL-012-10** — El Sprint Goal con HTML, comillas, tildes o ñ se muestra tal cual,
  escapado.
- **CL-012-11** — Proyecto inexistente o ID mal formado: 404, página con layout en un GET.
- **CL-012-12** — Los sprints de otro proyecto no aparecen.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje (origen) | Tipo (T-005) |
|---|---|---|---|
| El proyecto no existe o el ID está mal formado | `app.ErrProyectoInexistente`, envuelto indicando el ID | "el proyecto no existe" | `no_encontrado` (ya registrado) |
| Un sprint cerrado tiene más completados que planificados (dato corrupto) | `metricas.ErrResumenInconsistente`, envuelto con el nombre del sprint | "tiene mas Story Points completados que planificados: revisa sus datos" | exclusión del catálogo: error interno (ya registrado) |

En la pantalla, el primero responde 404 con layout; el segundo, 500 con el mensaje fijo de T-005.

## 8. Criterios de aceptación

- **CA-012-1** (normal) — Con dos sprints cerrados, la consulta los muestra del más reciente
  al más antiguo, cada uno con sus Story Points planificados y completados y sus historias
  completadas y no completadas.
- **CA-012-2** (alternativo) — Se muestra la velocidad del equipo: el promedio de los Story
  Points completados de los sprints cerrados.
- **CA-012-3** (alternativo) — El sprint abierto no aparece en el historial ni cuenta para la
  velocidad.
- **CA-012-4** (límite) — Sin sprints cerrados no hay error: la lista está vacía y la
  velocidad es 0. Un sprint cerrado sin historias aparece con todo en 0.
- **CA-012-5** (alternativo) — Una historia que no se completó en un sprint, volvió al
  backlog y se completó después en otro sigue figurando como no completada en el primero.
- **CA-012-6** (error) — Un proyecto inexistente se rechaza con `ErrProyectoInexistente` (404
  en la pantalla).

---

## Decisiones tomadas y descartadas

- **Es una vista de historial y no una pantalla nueva por sprint.** El detalle de un sprint
  cerrado ya lo muestra T-015/T-016 (con el resultado de cada historia); esta historia agrega
  la lista de solo los cerrados con su resultado resumido, que es lo que se pide para
  "consultar los anteriores". Se descartó repetir el detalle dentro de la lista.
- **El resultado de un sprint son sus Story Points y la cantidad de historias.** Se descartó
  inventar un "porcentaje de cumplimiento" propio: el porcentaje de historias es US-028 y la
  velocidad ya resume los Story Points. Mostrar cantidades es verificable y no duplica reglas.
- **La velocidad se muestra en esta pantalla.** Es el dato que más se mira junto al historial
  y ya la calcula `metricas`. Se descartó dejarla solo para el dashboard (US-032): sin ella,
  el historial no dice si el equipo está mejorando.
- **Sin puerto nuevo.** El caso de uso filtra lo que ya devuelve `ListarSprintsConItems`. Se
  descartó una consulta dedicada: no hay un volumen de sprints que lo justifique y una sola
  lectura se mantiene igual en memoria y en Postgres.
- **El sprint abierto no es "anterior".** Se descartó mostrarlo marcado: ya está en la lista
  de sprints de T-015, y mezclarlo falsearía la velocidad.
- **El resultado sale de la fila de cada sprint** (US-011, DEF-007). Es la lección del
  defecto: el historial no puede cambiar porque una historia se complete más tarde.
- **La ruta es `/sprints/anteriores`.** Go resuelve el literal antes que `{sprint}`, así que
  no choca con el detalle; un ID de sprint nunca vale "anteriores".

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia:

1. **Mostrar el porcentaje.** Cuando US-028 se mergee, ¿se suma a la tabla o se deja solo en
   el dashboard?
2. **Horas.** Las horas estimadas, reales y la desviación de cada sprint (US-027) podrían
   ir en esta tabla cuando haya consulta de horas por sprint (US-020).

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la historia, de
  US-011, US-025, US-026, T-015 y DEF-007; las decisiones sobre qué es el resultado de un
  sprint y qué se muestra se proponen acá y se confirman en la revisión
- Revisado y entendido por: Tomás Romero
