# US-011 — Cerrar un Sprint y devolver las historias no completadas al backlog

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #19 |
| **Escenarios BDD** | `features/US-011-cerrar-sprint.feature` |
| **Código** | `internal/domain/sprint/` (`Cerrar`, `Resultado`), caso de uso `CerrarSprint` en `internal/app/`, `HistoriasDelSprint` y `GuardarCierre` en `internal/adapters/postgres/` y `internal/adapters/memoria/`. Sin migración nueva |
| **Última actualización** | 2026-10-22 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero cerrar un Sprint, para dar por terminado
> el período, dejar registrado qué se completó y devolver al backlog lo que no.

Cerrar el Sprint abierto de un proyecto. Cada historia queda registrada en ese
Sprint como completada o no completada (la relación histórica de US-009), y
las no completadas vuelven al backlog para planificarse de nuevo. Con el Sprint
cerrado, sus números quedan fijos para la velocidad (US-026), y el proyecto
puede crear el Sprint siguiente (RN-008-17).

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `sprintID` | UUID | Sí | Un Sprint existente. Viene de la URL; lo verifica el caso de uso, nunca el dominio |

No son entradas la fecha (cerrar no mira el reloj ni las fechas del Sprint) ni
las historias: son las que la relación de US-009 tiene en ese Sprint.

El cierre pasa por las capas en este orden:

1. **Aplicación:** carga el Sprint con una lectura simple. Si no existe,
   devuelve `ErrSprintInexistente` sin validar nada más (RN-011-1).
2. **Dominio** (`internal/domain/sprint`): una función pura recibe los datos
   mínimos del Sprint y de sus historias (estado de cada una), valida que el
   Sprint esté abierto y devuelve el resultado de cada historia. Usa tipos
   propios del paquete, como US-009.
3. **Persistencia:** aplica el cierre en una sola transacción, con el Sprint
   bloqueado (RN-011-8).

## 3. Salidas esperadas

| Qué | Después de cerrar |
|---|---|
| Estado del Sprint | `cerrado` |
| Fila de cada historia en ese Sprint (US-009) | Resultado `completado` si la historia estaba completada; `no_completado` si no |
| Historias completadas | Sin cambios: siguen `completado`, con su fecha (US-010) |
| Historias no completadas | Estado `pendiente` (también si estaban `en_progreso`); vuelven al backlog |

Ninguna fila de la relación se borra. Si hay un error, no cambia nada.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-011-1** (aplicación) — El Sprint tiene que existir. Si no, se rechaza
  con `ErrSprintInexistente` (el de US-009), sin validar nada más. Un ID mal
  formado es un Sprint inexistente.
- **RN-011-2** (dominio) — Solo se cierra un Sprint `abierto`. Si ya está
  `cerrado`, `ErrSprintCerrado` (el de US-009, con su mensaje general) y no
  cambia nada.
- **RN-011-3** (dominio) — Se cierra en cualquier fecha: el dominio no lee el
  reloj ni compara con las fechas del Sprint.
- **RN-011-4** (dominio) — Un Sprint sin historias también se cierra.
- **RN-011-5** (dominio) — Cada historia del Sprint queda registrada con su
  resultado en la relación de US-009 (RN-009-13): `completado` si su estado es
  `completado`, y `no_completado` si es `pendiente` o `en_progreso`. La
  relación no se borra.
- **RN-011-6** (dominio) — Las historias no completadas vuelven al backlog:
  dejan de tener Sprint actual, porque ese Sprint ya no está abierto, y su
  estado pasa a `pendiente`, también si estaban `en_progreso`. Conservan su
  número, sus Story Points, sus criterios y su prioridad, y se pueden asignar a
  otro Sprint (CL-009-12).
- **RN-011-7** (dominio) — Las historias completadas quedan como están, y desde
  el cierre ya no se desmarcan (RN-010-3), no se editan (RN-006-5) ni se
  estiman (RN-013-4).
- **RN-011-8** (persistencia) — El cierre es atómico: una transacción
  `READ COMMITTED` con estos pasos en orden, que se detiene en el primer error.
  Si algo falla, no cambia nada.
  1. Bloquea la fila del Sprint y lee su estado:
     `SELECT ... FROM sprints WHERE id = $1 FOR NO KEY UPDATE`. Si no hay fila,
     `ErrSprintInexistente`.
  2. Lee las historias del Sprint (sus filas `pendiente` en la relación) y el
     estado de cada una.
  3. Vuelve a verificar RN-011-2 y calcula los resultados con la misma función
     del dominio, con los datos leídos con el candado tomado.
  4. Guarda el resultado de cada fila, pasa a `pendiente` las historias no
     completadas y pone el Sprint en `cerrado`.
- **RN-011-9** (persistencia) — El candado del paso 1 serializa el cierre con
  las operaciones que dependen del estado del Sprint: las asignaciones (US-009,
  `FOR SHARE`) y el marcado (US-010, `FOR NO KEY UPDATE`). Lo que llega antes
  del cierre cuenta; lo que llega después ve el Sprint `cerrado` y se rechaza
  (CL-009-14 y CL-010-9). Dos cierres simultáneos del mismo Sprint: el segundo
  espera y se rechaza con `ErrSprintCerrado`.
- **RN-011-10** (dominio) — Después de cerrar, el proyecto puede crear el
  Sprint siguiente (RN-008-17), que empieza después del fin del cerrado
  (RN-008-13).

## 5. Restricciones

- **Dominio puro y datos mínimos:** igual que US-008 y US-009 (sección 5).
- **Sin migración nueva.** El resultado de cada fila lo crea la migración de
  la relación de US-009, y el estado del Sprint, la de US-008.
- **Depende de la corrección de US-009 (#104).** Esta spec usa la relación
  histórica (RN-009-13 corregida). Con la versión anterior, que borraba la
  relación al cerrar, RN-011-5 no se podría cumplir.
- **Integración con US-025 y US-026.** Los Story Points planificados de un
  Sprint cerrado son los de todas sus filas, completadas o no (RN-025-1); los
  completados, los de las filas `completado` (RN-025-2). Con eso la capa de
  aplicación arma el `SprintCerrado` de US-026, como anticipa la sección 2 de
  US-025. Para el escenario de CA-011-3, el área Métricas de los escenarios
  tiene que calcular sobre los Sprints cerrados de la base del escenario, además
  de los que arma `el sprint cerrado "S" con N story points completados de M`.
- **Escenarios que otras historias dejaron para esta.** Se agregan en el PR de
  implementación, cada uno en el `.feature` de su historia y con su etiqueta:
  CA-008-4 y CA-008-9 (US-008), CA-009-6 (US-009) y CA-010-5 (US-010). Todos
  necesitan un Sprint cerrado.
- **Frases BDD.** Alcanzan las del diccionario:
  - `Cuando se cierra el sprint "X"` y `Dado el sprint "X" cerrado` (el cierre);
  - `el sprint "X" está cerrado`, `la historia "X" está en el backlog` (no
    está en el Sprint abierto) y `la historia "X" tiene estado "pendiente"`
    (los resultados);
  - `se consulta la velocidad del equipo` y `la velocidad del equipo es N`
    (CA-011-3).

  Las historias que se asignan necesitan criterios, por la DoR de US-009 (la
  frase con criterios que propone esa spec). La interfaz `Sprints` ya tiene
  `CerrarSprint` y `SprintEstaCerrado`; `CerrarSprint` pasa a recibir el
  proyecto, como `AsignarHistoriaASprint` en US-009, porque "Sprint 1" existe
  en todos los proyectos.
- **Métricas de los escenarios sobre Sprints cerrados.** Para CA-011-3, la base
  en memoria suma `FilasDelSprint`, una lectura de las historias que estuvieron
  en un Sprint con su resultado, que usan solo los escenarios (como
  `SprintActualDelItem`). El área Métricas calcula con ella los planificados y
  completados de los Sprints cerrados de la base, además de los que arma
  `el sprint cerrado "S" con N story points completados de M`.
- **Sin errores nuevos.** El cierre usa `ErrSprintInexistente` y
  `ErrSprintCerrado`, que ya están en el catálogo de T-005.
- **Fuera de alcance.** Reabrir o cancelar un Sprint, guardar la fecha de
  cierre, consultar los Sprints anteriores (US-012), el burndown (US-031) y la
  pantalla (T-011).

## 6. Casos límite

- **CL-011-1** — Sprint sin historias: se cierra; sus planificados y
  completados son 0 (CL-025-1).
- **CL-011-2** — Todas las historias completadas: todas quedan `completado` y
  ninguna vuelve al backlog.
- **CL-011-3** — Ninguna completada: todas quedan `no_completado` y vuelven al
  backlog como `pendiente`.
- **CL-011-4** — Historia `en_progreso`: queda `no_completado` y vuelve como
  `pendiente`.
- **CL-011-5** — Cerrar antes o después de la fecha de fin del Sprint: es
  válido (RN-011-3).
- **CL-011-6** — Sprint ya cerrado: `ErrSprintCerrado` y no cambia nada.
- **CL-011-7** — Sprint inexistente o ID mal formado: `ErrSprintInexistente`.
- **CL-011-8** — Historia devuelta al backlog y asignada al Sprint siguiente:
  queda con dos filas, `no_completado` en el cerrado y `pendiente` en el nuevo
  (CL-009-12).
- **CL-011-9** — Un marcado (US-010) simultáneo con el cierre: si obtiene el
  candado antes, la historia cuenta como completada; si llega después, se
  rechaza con `ErrSprintCerrado` (RN-011-9). Se verifica con un test de
  integración contra un Postgres real.
- **CL-011-10** — Dos cierres simultáneos del mismo Sprint: uno termina bien y
  el otro se rechaza con `ErrSprintCerrado`.
- **CL-011-11** — Falla la transacción a mitad del cierre: no cambia nada, el
  Sprint sigue abierto y las historias quedan como estaban (RN-011-8).

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) |
|---|---|---|
| El Sprint no existe | `app.ErrSprintInexistente` (el de US-009), envuelto indicando el ID | "el sprint no existe" |
| El Sprint ya está cerrado | `sprint.ErrSprintCerrado` (el de US-009), envuelto indicando el nombre del Sprint | "el sprint esta cerrado" |

`ErrSprintInexistente` tiene prioridad y lo detecta el caso de uso, o el
repositorio si el Sprint desaparece en el medio. `ErrSprintCerrado` lo
devuelven el dominio y, con la misma función, el repositorio cuando otro cierre
llegó antes. Los dos se comparan con `errors.Is`, y en ningún caso cambia nada.

## 8. Criterios de aceptación

- **CA-011-1** (normal) — Al cerrar un Sprint abierto, queda cerrado: sus
  historias completadas quedan registradas como completadas en ese Sprint y las
  no completadas vuelven al backlog como pendientes.
- **CA-011-2** (alternativo) — Después de cerrar un Sprint se puede crear el
  siguiente, y una historia que volvió al backlog se puede asignar a él.
- **CA-011-3** (alternativo) — Los Sprints cerrados conservan lo planificado y
  lo completado: los Story Points planificados incluyen las historias no
  completadas, y la velocidad del equipo se calcula con lo completado en cada
  Sprint cerrado.
- **CA-011-4** (límite) — Un Sprint sin historias se cierra.
- **CA-011-5** (error) — Un Sprint ya cerrado no se vuelve a cerrar.
- **CA-011-6** (error) — Un Sprint inexistente se rechaza y no se modifica
  nada.

---

## Decisiones tomadas y descartadas

- **La relación es histórica y no se borra al cerrar.** Cada historia queda
  registrada en ese Sprint como completada o no completada. Sin esa historia se
  pierden los Story Points planificados de los Sprints cerrados (US-025) y el
  reporte de historias planificadas y completadas. Se descartó guardar solo los
  totales al cerrar: no permite listar qué historias se planificaron. Requiere
  la corrección de US-009 (#104).
- **"Volver al backlog" no borra nada.** Una historia está en el backlog cuando
  no está en el Sprint abierto. Al cerrar, las no completadas dejan de estarlo
  porque el Sprint ya no está abierto, y su estado vuelve a `pendiente` para
  planificarse de nuevo. Se descartó conservar `en_progreso`: arrancarían el
  Sprint siguiente como si alguien ya las estuviera trabajando.
- **Se cierra en cualquier fecha.** El dominio no lee el reloj, y el equipo
  puede cerrar antes (se terminó todo) o después (la Review se corrió un día).
  Se descartó exigir que la fecha de fin haya pasado.
- **Un Sprint sin historias se cierra.** Es un Sprint que no planificó nada, y
  sin cerrarlo el proyecto no podría crear el siguiente (RN-008-17).
- **Cierre atómico con el Sprint bloqueado.** Cerrar toca muchas filas (la
  relación, las historias y el Sprint): a medias dejaría historias fuera del
  backlog o un Sprint abierto con resultados. `FOR NO KEY UPDATE` sobre el
  Sprint serializa el cierre con las asignaciones y los marcados sin bloquear
  las claves foráneas (mismo criterio que US-008 y US-010). Se descartó
  bloquear cada historia: el candado del Sprint ya ordena todo lo que depende
  de su estado.
- **Los escenarios pendientes de otras historias van en sus `.feature`.**
  CA-008-4, CA-008-9, CA-009-6 y CA-010-5 son criterios de esas historias, y la
  trazabilidad spec → escenario va por archivo y etiqueta. Llegan en este PR
  porque recién acá existe el cierre.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia y de las decisiones del equipo (relación histórica, cierre en
  cualquier fecha, cierre atómico con el Sprint bloqueado), usando las specs de
  US-008, US-009 corregida, US-010, US-025 y US-026 como modelo
- [x] Use IA para: el ajuste tras la implementación (nombres del código,
  `CerrarSprint` con el proyecto, métricas de los escenarios sobre Sprints
  cerrados, sin errores nuevos), redactado con Claude Code
- Revisado y entendido por: Juan Ignacio Vergara
