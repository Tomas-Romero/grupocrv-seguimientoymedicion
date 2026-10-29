# US-004 — Consultar el estado general de un proyecto

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #16 |
| **Escenarios BDD** | `features/US-004-estado-proyecto.feature` |
| **Código** | `internal/domain/proyecto/`, `internal/app/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| **Última actualización** | 2026-10-28 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero consultar el estado general de un proyecto, para saber de un
> vistazo si está por empezar, en marcha o terminado, y cuánto del trabajo está hecho.

Mostrar en la página del proyecto (`/proyectos/{id}`) una tarjeta de **estado general**: el estado
del proyecto según sus fechas (no iniciado, en curso o finalizado), los días que faltan o que
pasaron, el avance del backlog (historias por estado y Story Points) y el sprint abierto, si lo hay.
Cubre el requerimiento 1 de la guía (estado de un proyecto). El dashboard (US-030) reutiliza este
resumen.

La guía pide el «estado de un proyecto» sin listar los estados. Esta spec propone que **se derive de
las fechas del proyecto**, porque es lo único que el proyecto ya tiene (US-001 y US-002) y no
obliga a una operación nueva para cambiarlo (ver Decisiones y Huecos).

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | `string` | Sí | ID de un proyecto existente. Viene de la URL (`GET /proyectos/{id}`); lo verifica el caso de uso |
| `hoy` | `time.Time` | Sí | La fecha de hoy, sin hora. Entra por parámetro: el dominio no lee el reloj |

La consulta pasa por dos capas, y cada una valida solo lo que puede saber:

- **Dominio** (`internal/domain/proyecto`): una función pura recibe el proyecto y la fecha de hoy y
  devuelve el estado y los días. No lee la base ni el reloj.
- **Aplicación** (`internal/app`): `ConsultarEstadoDelProyecto` verifica que el proyecto exista, lee
  sus historias y sus sprints, calcula el avance con `metricas` (US-025 y US-028) y arma la respuesta.

## 3. Salidas esperadas

```
EstadoDelProyecto { // en internal/domain/proyecto
    NoIniciado | EnCurso | Finalizado
}

func (p Proyecto) EstadoEn(hoy time.Time) (estado EstadoDelProyecto, dias int)
```

| Estado | Cuándo | `dias` |
|---|---|---|
| `no_iniciado` | `hoy` es anterior a la fecha de inicio | Días hasta que empieza (`inicio − hoy`, al menos 1) |
| `en_curso` | `hoy` está entre el inicio y el fin, ambos incluidos | Días que faltan para el fin (`fin − hoy`; 0 el último día) |
| `finalizado` | `hoy` es posterior a la fecha de fin | Días desde el fin (`hoy − fin`, al menos 1) |

Los días son de calendario y se comparan solo año, mes y día (UTC).

El caso de uso devuelve:

```
type EstadoGeneral struct {
    Proyecto       proyecto.Proyecto
    Estado         proyecto.EstadoDelProyecto
    Dias           int
    Historias      HistoriasPorEstado // Pendientes, EnProgreso, Completadas y Total
    StoryPoints    int                // de las historias estimadas
    Completados    int                // Story Points de las historias completadas
    Porcentaje     float64            // % de historias completadas, a 1 decimal
    SprintAbierto  *sprint.Sprint     // nil si no hay
}
```

**En pantalla**, arriba del resumen del proyecto: el estado con su texto («No iniciado: empieza en 12
días», «En curso: faltan 40 días», «Finalizado: terminó hace 3 días»; «En curso: termina hoy» si `dias` es 0),
las fechas, las tres cantidades de historias, «Story Points: 13 completados de 21», el porcentaje y el
sprint abierto («Sprint 2 en curso» con un enlace a su detalle, o «Sin sprint abierto»).

## 4. Reglas de negocio

- **RN-004-1** (dominio) — El estado se deriva de las fechas del proyecto y de `hoy`, según la tabla de
  la sección 3. El día de inicio y el día de fin son «en curso».
- **RN-004-2** (dominio) — Los días se calculan en días de calendario. Un proyecto de un solo día
  (inicio igual a fin) está en curso ese día, con 0 días.
- **RN-004-3** (dominio) — La función no valida las fechas (US-001 ya garantiza que el fin no es anterior
  al inicio): recibe un proyecto armado. Con un proyecto sin fechas (valor cero) devuelve `no_iniciado`
  sin fallar; no es un caso que pueda darse con datos guardados.
- **RN-004-4** (aplicación) — Un proyecto que no existe se rechaza con `ErrProyectoInexistente`,
  sin leer nada más.
- **RN-004-5** (aplicación) — Las historias se cuentan por su estado actual: `pendiente`, `en_progreso` y
  `completado`. Los Story Points son la suma de los de las historias estimadas (las sin estimar suman 0) y
  los completados, la de las historias `completado`.
- **RN-004-6** (aplicación) — El porcentaje es el de `metricas.CalcularPorcentajeCompletadas` (US-028) sobre
  todas las historias del proyecto: sin historias es 0 y no es un error.
- **RN-004-7** (aplicación) — El sprint abierto es el sprint del proyecto con estado `abierto` (US-008
  garantiza que hay como mucho uno). Sin sprints, o con todos cerrados, es `nil`.
- **RN-004-8** (aplicación) — La consulta no modifica nada y usa una sola lectura de cada tipo (proyecto,
  historias y sprints con sus historias), sin una consulta por historia.
- **RN-004-9** (interfaz) — La tarjeta se muestra en la página del proyecto y no agrega una ruta. Si falla
  el cálculo (datos corruptos), el resto de la página se muestra igual y la tarjeta informa el error.

## 5. Restricciones

- **Dominio puro.** `internal/domain/proyecto` sigue sin importar nada fuera de la biblioteca estándar. El
  estado es un tipo del paquete; no se guarda en la base, siempre se deriva (ver Decisiones).
- **Puerto propio** `RepositorioEstadoGeneral` con `BuscarProyecto`, `ListarItems` y `ListarSprintsConItems`,
  que son métodos que ya existen en `postgres` y en `memoria`: no hay consultas nuevas. `cmd/server` los
  junta como hizo con `lecturaDeProyecto`.
- **Reloj.** `NuevoConsultarEstadoDelProyecto(repo, ahora)` recibe el reloj (`time.Now` en `cmd/server`, uno
  fijo en los tests y en los escenarios, que usan el 2026-10-06).
- **Sin migración ni errores nuevos.** Reutiliza `ErrProyectoInexistente`.
- **Escenarios BDD.** Van en el área Proyectos de los steps. Se agregan frases (ver Decisiones). Los
  escenarios de «no iniciado» y «finalizado» usan `Dado un proyecto "X" con fechas del ... al ...`, que ya existe.
- **Fuera de alcance.** Cambiar a mano el estado del proyecto, estados que no se derivan de las fechas
  (por ejemplo «pausado»), el historial de estados, gráficos y la lista de proyectos (T-011).

## 6. Casos límite

- **CL-004-1** (RN-004-1) — Hoy es el día de inicio: `en_curso`, con los días del fin.
- **CL-004-2** (RN-004-1) — Hoy es el día de fin: `en_curso` con 0 días («termina hoy»). Al día siguiente:
  `finalizado` con 1 día.
- **CL-004-3** (RN-004-1) — Hoy es el día anterior al inicio: `no_iniciado` con 1 día.
- **CL-004-4** (RN-004-2) — Proyecto de un solo día: en curso ese día con 0 días; antes, no iniciado; después,
  finalizado.
- **CL-004-5** (RN-004-2) — Un `hoy` con hora (23:59) se trata como el mismo día: no cambia el estado.
- **CL-004-6** (RN-004-3) — Proyecto sin fechas (valor cero): `no_iniciado` sin fallar.
- **CL-004-7** (RN-004-5, RN-004-6) — Proyecto sin historias ni sprints: 0 historias, 0 Story Points, porcentaje 0
  y sin sprint abierto, sin error ni división por cero.
- **CL-004-8** (RN-004-5) — Una historia sin estimar y completada: cuenta como historia completada y suma 0 Story
  Points.
- **CL-004-9** (RN-004-5) — Una historia que no está en ningún sprint: se cuenta igual (el avance es del
  backlog completo, no solo del sprint).
- **CL-004-10** (RN-004-7) — Con todos los sprints cerrados: sin sprint abierto.
- **CL-004-11** (RN-004-4) — Proyecto inexistente o con ID mal formado: `ErrProyectoInexistente`, y la página
  responde 404.
- **CL-004-12** (RN-004-9) — Un proyecto finalizado con historias sin completar: el estado es `finalizado` y el
  avance dice cuánto faltó; el estado no depende del avance.

## 7. Condiciones de error

| Situación | Error devuelto | Respuesta de la página |
|---|---|---|
| El proyecto no existe, o el ID está mal formado | `app.ErrProyectoInexistente`, envuelto con el ID | 404 con layout (RN-T011-10) |
| Falla inesperada de la persistencia | Error envuelto con contexto | 500 con el mensaje fijo de T-005 |
| Un dato corrupto que `metricas` rechaza (Story Points negativos) | El error de `metricas`, envuelto | La tarjeta informa el error y el resto de la página se muestra (RN-004-9) |

Que el proyecto no tenga historias, sprints o sprint abierto **no es un error**: son casos límite con resultado.

## 8. Criterios de aceptación

- **CA-004-1** (normal) — Un proyecto del 1/1 al 31/12, con hoy el 6/10, está **en curso** y la tarjeta dice
  cuántos días faltan para el fin.
- **CA-004-2** (alternativo) — Un proyecto que empieza dentro de 12 días está **no iniciado** («empieza en 12
  días») y uno que terminó hace 3 días está **finalizado** («terminó hace 3 días»).
- **CA-004-3** (alternativo) — Con tres historias (una pendiente, una en progreso y una completada de 5
  Story Points) y un sprint abierto, la tarjeta muestra 1, 1 y 1 historias, «5 completados de N» Story Points,
  33,3 % y el sprint abierto con su enlace.
- **CA-004-4** (límite) — El día de inicio, el día de fin y un proyecto de un solo día están en curso, y el día
  siguiente al fin está finalizado.
- **CA-004-5** (límite) — Un proyecto sin historias ni sprints muestra todo en 0, 0 % y «Sin sprint abierto»,
  sin error.
- **CA-004-6** (error) — Un proyecto inexistente se rechaza con `ErrProyectoInexistente` y la página
  responde 404.

---

## Decisiones tomadas y descartadas

- **El estado se deriva de las fechas y no se guarda.** La guía no lista los estados ni pide cambiarlos a
  mano. Derivarlo evita un campo nuevo, una migración y una operación («cambiar estado») que ninguna historia
  pide, y no puede quedar desactualizado. Se descartó guardarlo: habría que sincronizarlo con las fechas
  cuando US-002 las modifica.
- **Tres estados** (`no_iniciado`, `en_curso`, `finalizado`) con nombres como los de las historias y los
  sprints (minúscula, sin tilde y con guion bajo). Se descartó un estado «atrasado» (en curso pero con poco
  avance): necesita una regla de qué es poco, que nadie definió.
- **El avance es del backlog completo.** Una historia sin sprint también cuenta: el avance del proyecto es lo
  que se hizo sobre todo lo que se planeó, no solo sobre lo del sprint (que ya muestra la pantalla de sprints).
- **El sprint abierto se muestra, pero no sus números.** Los Story Points y las horas del sprint van al
  dashboard (US-030).
- **Sin ruta nueva.** La tarjeta vive en la página del proyecto, que es donde una persona mira el «estado» del
  proyecto. Se descartó una página aparte: sería la misma información un clic más lejos.
- **Frases nuevas del diccionario** (se aprueban en el PR de implementación), área Proyectos:
  `se consulta el estado del proyecto "P"`, `el estado del proyecto es "en_curso"`,
  `los días del estado son N`, `el proyecto tiene N historias pendientes, M en progreso y K completadas`,
  `los story points del proyecto son Y y los completados X`,
  `el porcentaje de historias completadas del proyecto es N` y `el sprint abierto del proyecto es "Sprint N"` o
  `el proyecto no tiene sprint abierto`.

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia:

1. **Qué entiende la guía por «estado».** Se propone el derivado de las fechas. Si la cátedra espera otros estados
   (planificado, en ejecución, cerrado…), cambian los nombres y la tabla de la sección 3, no el resto.
2. **Días de calendario o hábiles.** Se usan de calendario, que es lo que se ve en un calendario.
3. **`en_progreso` no se alcanza hoy** (se descubrió al implementar). La spec cuenta las historias `en_progreso`, pero ninguna
   operación de la aplicación pasa una historia a ese estado (ni asignarla a un sprint, ni US-010): siempre vale 0 en
   la práctica. Se deja la cantidad porque es parte del modelo del backlog y no cuesta nada. Los escenarios la verifican en 0
   y el caso de uso la prueba con un repositorio simulado (`TestConsultarEstado_Avance`). El «una en progreso» de CA-004-3 se
   cubre así y no con un escenario.
4. **Un dato corrupto no se puede armar desde afuera.** RN-004-9 (la tarjeta informa el error de `metricas`) tiene su manejo en
   la pantalla, pero `backlog.StoryPoints` no deja crear Story Points negativos, así que no hay forma de probarlo con datos reales:
   queda como defensa y sin test de pantalla.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la historia, de US-001, US-025,
  US-028 y de la forma de las specs de metricas; el criterio de derivar el estado de las fechas se propone acá y se
  confirma en la revisión
- Revisado y entendido por: Tomás Romero
