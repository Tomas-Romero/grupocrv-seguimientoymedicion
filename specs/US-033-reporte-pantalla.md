# US-033 — Generar en pantalla el reporte completo de un proyecto o Sprint

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #39 |
| **Escenarios BDD** | `features/US-033-reporte-pantalla.feature` |
| **Código** | `internal/app/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| **Última actualización** | 2026-10-27 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero generar en pantalla el reporte completo de un proyecto o de un
> sprint, para revisar todo lo que pasó en un solo lugar y poder imprimirlo.

Dos páginas de solo lectura que reúnen, en un orden fijo, lo que ya muestran las demás pantallas:
el **reporte del proyecto** (`/proyectos/{id}/reporte`) y el **reporte de un sprint**
(`/proyectos/{id}/sprints/{sprint}/reporte`). Cubre el requerimiento 9 de la guía (reportes). La
exportación a PDF (US-034) no genera un archivo en el servidor: la cátedra confirmó el 26/10 que alcanza con que
el navegador imprima la página como PDF, así que esta historia deja el reporte **pensado para imprimir**: una
sola página con secciones ordenadas y títulos claros.

No calcula nada nuevo: **compone casos de uso que ya existen** (el dashboard de US-030, el esfuerzo de
T-019, los sprints de T-015, los defectos de T-021) y los muestra en un solo documento.

Depende de **US-030** (las métricas), **US-004** (el estado), **T-019** (el esfuerzo) y **T-021** (la lectura de
defectos).

## 2. Entradas

| Ruta | Método | Parámetros | Caso de uso |
|---|---|---|---|
| `/proyectos/{id}/reporte` | GET | — | `ConsultarReporteDelProyecto` |
| `/proyectos/{id}/sprints/{sprint}/reporte` | GET | — | `ConsultarReporteDelSprint` |

`{id}` y `{sprint}` son los IDs que genera la base. Un sprint solo se alcanza por la ruta de **su** proyecto
(RN-T015-5): `/proyectos/{A}/sprints/{sprint de B}/reporte` responde 404 aunque el sprint exista. En el detalle
de cada sprint y en el menú del proyecto hay un enlace «Reporte».

## 3. Salidas esperadas

**Reporte del proyecto**, con estas secciones **siempre en este orden**:

| # | Sección | Contenido |
|---|---|---|
| 1 | Proyecto | Nombre, descripción, fechas, estado y días (US-004) y el avance del backlog |
| 2 | Integrantes | Nombre, rol y email, ordenados como en T-011 |
| 3 | Product Backlog | Cada historia con su prioridad, estado y Story Points, en el orden de T-011 |
| 4 | Sprints | Cada sprint con su objetivo, fechas, estado, Story Points planificados y completados y cuántas historias volvieron al backlog |
| 5 | Esfuerzo | El total de horas registradas y la comparación de estimado contra real por historia (T-019), con las horas sin comparar |
| 6 | Defectos | La lista de defectos (severidad, estado, historia, sprint de detección y de resolución) y los totales por severidad y por estado |
| 7 | Métricas | La velocidad del equipo, el sprint de referencia y los sprints cerrados con sus defectos detectados y resueltos (US-030) |

**Reporte de un sprint**, también en orden fijo:

| # | Sección | Contenido |
|---|---|---|
| 1 | Sprint | Nombre, objetivo, fechas y estado |
| 2 | Historias | Cada historia del sprint con sus Story Points y si se completó en él (el resultado de su fila) |
| 3 | Métricas | Story Points planificados y completados, porcentaje de historias completadas y horas estimadas, reales y desviación |
| 4 | Esfuerzo registrado | Los registros de esfuerzo cuya fecha cae en el sprint, con su total |
| 5 | Defectos | Los detectados y los resueltos en el sprint, con sus totales |

Cada sección tiene un título y un identificador (`#proyecto`, `#integrantes`, `#backlog`, `#sprints`,
`#esfuerzo`, `#defectos` y `#metricas`; en el de sprint, `#sprint`, `#historias`, `#metricas`, `#esfuerzo` y
`#defectos`) para poder enlazarlas. Una sección sin datos se muestra igual, con el texto «Sin datos» (por
ejemplo «Sin defectos»): el reporte siempre tiene todas sus secciones. Fechas como `dd/mm/aaaa` y números con el
formato de T-019.

## 4. Reglas de negocio

- **RN-033-1** — Los reportes son **de solo lectura** y componen casos de uso existentes (`ConsultarDashboard`,
  `ConsultarProyecto`, `ListarBacklog`, `ListarSprints`, `ConsultarSprint`, `CompararEsfuerzo`,
  `ConsultarEsfuerzo` y `ListarDefectos`). El caso de uso del reporte **no accede a repositorios** y **no
  recalcula** nada: usa las cifras que devuelve cada uno.
- **RN-033-2** — El orden de las secciones es fijo (sección 3). Dentro de cada sección rige el orden de la
  pantalla de origen: integrantes por nombre (RN-T011-5), backlog por prioridad y número (RN-T011-4), sprints
  por número descendente (RN-T015-3), defectos del más reciente al más antiguo.
- **RN-033-3** — Cada sección falla por separado (como las tarjetas del dashboard, RN-030-9): un error al armar una
  sección se informa en ella y el resto del reporte se muestra.
- **RN-033-4** — El reporte de un sprint muestra el resultado de cada historia según **su fila** en ese sprint, no
  su estado actual (RN-T015-13, DEF-007); y las horas son las de los registros con fecha dentro del sprint
  (RN-019-12).
- **RN-033-5** — Los defectos del reporte de un sprint son los que tienen ese sprint como **sprint de detección**
  (los detectados) y los que tienen ese sprint como **sprint de resolución** (los resueltos). Los `descartado`
  no se cuentan en los totales de detectados y resueltos (RN-030-6), pero aparecen en la lista del proyecto.
- **RN-033-6** — Un proyecto o un sprint que no existe, o un sprint de otro proyecto, se rechaza con su error de
  inexistencia **antes** de armar nada: 404 con layout.
- **RN-033-7** — Los reportes no modifican nada y no hay una operación «generar» que guarde un documento: se
  calculan en cada visita con los datos de ese momento.
- **RN-033-8** (impresión) — Cada sección va dentro de un `<section>` con su título; la página funciona sin
  JavaScript. La hoja de estilos de impresión y el botón «Imprimir o guardar como PDF» son de US-034, que no
  cambia las secciones ni su contenido.
- **RN-033-9** — El texto con HTML (descripciones de proyecto, historias y defectos) se muestra escapado
  (RN-T011-8). El handler solo lee la petición, llama al caso de uso y escribe la respuesta (RN-T011-1).

## 5. Restricciones

- **Sin cálculos ni lecturas nuevas.** No se toca el dominio ni los repositorios.
- **Casos de uso nuevos:** `ConsultarReporteDelProyecto` y `ConsultarReporteDelSprint`, en `internal/app`, que
  reciben los casos de uso existentes por el constructor. `cmd/server` los arma con los que ya tiene.
- **Sin migración ni errores nuevos.** Reutiliza los de inexistencia del catálogo.
- **Una página con todas las secciones,** no una página por sección ni paginación: es lo que se imprime.
- **Escenarios BDD.** Van en el área Métricas de los steps y comprueban lo que cada sección contiene y el orden;
  se agregan frases de consulta (ver Decisiones).
- **Fuera de alcance.** Generar un PDF en el servidor, enviar el reporte por correo, guardar versiones del
  reporte, filtros por fecha, gráficos (US-032) y comparar sprints.

## 6. Casos límite

- **CL-033-1** (RN-033-2) — Un proyecto con todo cargado: las siete secciones aparecen, en el orden de la sección
  3, con los datos de cada una.
- **CL-033-2** (RN-033-2) — Un proyecto recién creado (sin integrantes, historias, sprints, horas ni defectos): las
  siete secciones aparecen igual con «Sin datos», sin error.
- **CL-033-3** (RN-033-4) — Una historia no completada en el Sprint 1 y completada en el Sprint 2: el reporte del
  Sprint 1 la muestra «No completada» y el del Sprint 2, «Completada».
- **CL-033-4** (RN-033-4) — Horas con fecha entre dos sprints, o antes del primero: no están en el reporte de
  ningún sprint, pero sí en el total del proyecto.
- **CL-033-5** (RN-033-5) — Un defecto detectado en el Sprint 2 y resuelto en el 4: aparece como detectado en el
  reporte del Sprint 2 y como resuelto en el del 4, y en ninguno más.
- **CL-033-6** (RN-033-5) — Un defecto descartado: está en la lista de defectos del proyecto con su estado y no
  cuenta en los totales de detectados ni resueltos.
- **CL-033-7** (RN-033-3) — Un dato inválido guardado en una sección (por ejemplo, Story Points inconsistentes): esa
  sección informa el error y las demás se muestran.
- **CL-033-8** (RN-033-6) — Un sprint de otro proyecto, o un ID inexistente o mal formado: 404 con layout.
- **CL-033-9** (RN-033-2) — Un sprint sin historias: el reporte muestra la sección de historias vacía y el
  porcentaje en 0.
- **CL-033-10** (RN-033-9) — Un texto con `<script>` en la descripción de un defecto: se muestra escapado.
- **CL-033-11** — Un sprint abierto: el reporte lo marca «en curso» y las cifras son las de hoy.

## 7. Condiciones de error

| Situación | Error devuelto | Respuesta |
|---|---|---|
| El proyecto no existe, o el ID está mal formado | `app.ErrProyectoInexistente` | 404 con layout |
| El sprint no existe, o no es de ese proyecto | `app.ErrSprintInexistente` | 404 con layout |
| Una sección no se puede armar | El error de esa sección, envuelto | La sección informa el error y el resto del reporte se muestra |
| Falla inesperada de la persistencia | Error envuelto con contexto | 500 con el mensaje fijo de T-005 |

Que no haya datos **no es un error** (RN-033-2 y CL-033-2).

## 8. Criterios de aceptación

- **CA-033-1** (normal) — El reporte del proyecto muestra sus siete secciones en el orden de la sección 3: el
  proyecto, los integrantes, el backlog con sus estimaciones, los sprints con su resultado, el esfuerzo, los
  defectos y las métricas.
- **CA-033-2** (alternativo) — El reporte de un sprint muestra sus historias con Story Points y si se
  completaron, el porcentaje y las horas, el esfuerzo registrado y los defectos detectados y resueltos en él.
- **CA-033-3** (límite) — Un proyecto o un sprint sin datos muestra todas las secciones con «Sin datos» y sin error.
- **CA-033-4** (error) — Un proyecto o un sprint inexistente, o un sprint de otro proyecto, responde 404 con layout.
- **CA-033-5** (plataforma) — Es una sola página con las secciones siempre en el mismo orden y sin JavaScript;
  no duplica cálculos: las cifras son las que devuelven el dashboard y los demás casos de uso.

---

## Decisiones tomadas y descartadas

- **Se compone y no se recalcula.** El reporte es la suma de lo que ya muestran las pantallas. Se descartó un
  reporte con consultas y cálculos propios: cada cifra tendría dos lugares donde mantenerse y podrían no coincidir
  con el dashboard.
- **Siempre las mismas secciones, aunque estén vacías.** Un reporte que cambia de forma según los datos es
  difícil de comparar entre sprints y de imprimir. Se descartó ocultar las secciones vacías.
- **Sin generación en el servidor.** La cátedra aceptó imprimir desde el navegador; US-034 solo suma la hoja de
  estilos y un botón. Se descartó una librería de PDF: una dependencia nueva (y su ADR) para algo que el navegador
  ya hace.
- **Dos rutas y no una con un filtro.** El reporte de un sprint tiene otras secciones y otro título, y su ruta
  cuelga del sprint, como el resto de las páginas de sprints.
- **Cada sección falla por separado,** como el dashboard: un dato corrupto no deja al equipo sin reporte para
  imprimir.
- **Frases nuevas del diccionario** (se aprueban en el PR de implementación), área Métricas:
  `se consulta el reporte del proyecto "P"`, `se consulta el reporte del sprint "Sprint N" del proyecto "P"`,
  `el reporte tiene las secciones: ...` (con las secciones en una tabla, en orden), `la sección "S" del reporte está
  vacía`, `la sección "S" del reporte indica "texto"` y `el reporte del sprint muestra la historia "H" como
  "Completada"` (o `"No completada"`).

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la historia y de las specs de US-004,
  US-030, T-015 y T-019; el criterio de componer casos de uso y de las secciones fijas se propone acá y se
  confirma en la revisión
- Revisado y entendido por: Tomás Romero
