# US-007 — Listar y filtrar el backlog de un proyecto por estado y prioridad

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Angelo Conforti |
| **Issue** | #17 (historia) · issue de la spec: pendiente |
| **Depende de** | US-005 (ítems, estados y prioridades), US-010 (estado `completado`) y T-011 (el listado y `ListarBacklog` que esta historia extiende), las tres ya en `main` |
| **Escenarios BDD** | `features/US-007-listar-backlog.feature` |
| **Código** | `internal/app/` (`ListarBacklog` y su filtro), `internal/domain/backlog/` (`ValidarPrioridad`, `ValidarEstado`, `ErrEstadoInvalido`) y `internal/adapters/http/catalogo.go` |
| **Sprint** | ninguno: la historia salió del alcance comprometido (plan de trabajo, «Recorte aplicado»). Entra solo si sobra capacidad |
| **Última actualización** | 2026-10-29 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero listar el Product Backlog de un proyecto
> filtrando por el estado y por la prioridad de las historias, para encontrar
> enseguida las que me interesan (por ejemplo, las pendientes de prioridad
> `must`) sin recorrer el backlog entero.

Agregar dos filtros opcionales al listado del backlog que T-011 dejó hecho: el
**estado** de la historia y su **prioridad** MoSCoW. Con uno, con los dos a la vez
(solo salen las historias que cumplen ambos) o con ninguno (el backlog completo de
T-011). Un filtro sin coincidencias da una lista vacía, no un error; un estado o
una prioridad que no existen se rechazan con el error uniforme de T-005.

Es una mejora de lectura: no crea ni cambia historias. Extiende `ListarBacklog` de
T-011 sin cambiarle el nombre ni el resultado, y el orden sigue siendo el de esa
tarea (RN-T011-4). T-011 dejó este filtro fuera de su alcance a propósito
(sección 5 de esa spec).

Esta historia entrega dominio, caso de uso y BDD. No agrega repositorio nuevo (la
lectura ya existe) ni controles en pantalla: las pantallas quedan fuera de las
historias (plan de trabajo, regla 4).

## 2. Entradas

El caso de uso es `app.ListarBacklog`, el de T-011, con un parámetro más:

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. Viene de la URL; lo verifica el caso de uso, nunca el dominio |
| `filtro.Estado` | `backlog.Estado` | No | Exactamente `pendiente`, `en_progreso` o `completado`: en minúscula y sin espacios alrededor (RN-005-6). Vacío: no se filtra por estado |
| `filtro.Prioridad` | `backlog.Prioridad` | No | Exactamente `must`, `should`, `could` o `wont`: en minúscula y sin espacios alrededor (RN-005-5). Vacía: no se filtra por prioridad |

**Orden de verificación** (RN-007-9). El caso de uso sigue este orden:

1. Verifica que el proyecto exista (`ErrProyectoInexistente`). Si no existe, devuelve
   solo ese error, sin mirar los filtros.
2. Valida los filtros que se informaron, primero el estado y después la prioridad.
   Si alguno es inválido devuelve los errores encontrados **unidos**, en ese orden.
3. Lee las historias del proyecto, aplica los filtros y las ordena.

No son entradas: el sprint al que está asignada la historia, sus Story Points, su
título ni ningún otro dato. La forma en que la persona elige el filtro en la
pantalla queda en el Hueco 1; convertir lo que llega de la pantalla en este
`filtro` será tarea del adaptador HTTP.

## 3. Salidas esperadas

```
type FiltroBacklog struct {
    Estado    backlog.Estado
    Prioridad backlog.Prioridad
}

func (c *ListarBacklog) Ejecutar(ctx context.Context, proyectoID string, filtro FiltroBacklog) ([]backlog.ItemBacklog, error)

func ValidarPrioridad(prioridad Prioridad) error // internal/domain/backlog
func ValidarEstado(estado Estado) error          // internal/domain/backlog

var ErrEstadoInvalido = errors.New("el estado tiene que ser pendiente, en_progreso o completado") // internal/domain/backlog
```

El resultado es el `[]backlog.ItemBacklog` de T-011, sin campos nuevos: las mismas
historias con los mismos datos, pero solo las que cumplen el filtro y en el mismo
orden (RN-007-6). Sin filtros, o sin coincidencias, el resultado es el backlog
completo o una lista vacía, respectivamente; en ningún caso un error.

`ValidarPrioridad` es la validación de RN-005-5 que hoy está escrita dentro de
`validarDatos` (la usan `NuevoItem` y `Editar`), expuesta para que el filtro y el
alta usen **la misma regla**. Devuelve `nil` si la prioridad es una de las cuatro, y
`ErrPrioridadInvalida` envuelto con el valor recibido (`se recibio "urgente"`) en
cualquier otro caso, incluida la vacía.

`ValidarEstado` hace lo mismo con los tres estados y devuelve `ErrEstadoInvalido`,
envuelto con el valor recibido. En ambas, que el valor vacío signifique "sin filtro"
lo decide el caso de uso antes de llamarlas (RN-007-7 y RN-007-8).

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-007-1** (aplicación) — Sin filtros (`Estado` y `Prioridad` vacíos) el
  resultado es **exactamente** el de T-011: todas las historias del proyecto, en
  cualquier estado y en el mismo orden. Agregar los filtros no cambia lo que el
  listado ya hacía.
- **RN-007-2** (aplicación) — El filtro por estado deja solo las historias cuyo
  **estado actual** es el pedido, de cualquier prioridad. Una historia `completada`
  sale si se pide `completado`: el listado no oculta ningún estado por su cuenta.
- **RN-007-3** (aplicación) — El filtro por prioridad deja solo las historias de esa
  prioridad, en cualquier estado.
- **RN-007-4** (aplicación) — Los dos filtros se combinan con **Y**, como los de
  US-020 (RN-020-2) y US-024 (RN-024-4): una historia sale si cumple todos los
  filtros que vienen completos. Con uno solo informado, se aplica ese solo.
- **RN-007-5** (aplicación) — Que ninguna historia cumpla el filtro **no es un
  error**: el resultado es una lista vacía. Vale para un estado sin historias, para
  una prioridad sin historias, para dos filtros cuya combinación no tiene
  coincidencias aunque cada uno tenga por separado, y para un proyecto sin
  historias.
- **RN-007-6** (aplicación) — El filtro solo **quita** historias: no cambia el
  contenido de las que quedan ni su orden. Cada historia es la misma
  `ItemBacklog` que sin filtrar, y la lista sigue por prioridad (`must`, `should`,
  `could`, `wont`) y, dentro de cada una, por número ascendente (RN-T011-4). El
  orden se calcula sobre lo que queda: filtrar por un estado no pone primero a
  ninguna prioridad distinta de la de siempre.
- **RN-007-7** (dominio) — Si se informó un estado, tiene que ser uno de los tres y
  exacto: no se pasa a minúscula ni se recorta. Cualquier otro valor (`Pendiente`,
  `COMPLETADO`, `en progreso`, `" pendiente "`, `terminado`) se rechaza con
  `ErrEstadoInvalido`, envuelto con el valor. La regla está escrita **una sola vez**
  (`ValidarEstado`). El estado **vacío** no es inválido: es "sin filtro", y el caso
  de uso no llama al dominio. Un estado de solo espacios no está vacío: es inválido.
- **RN-007-8** (dominio) — Lo mismo para la prioridad: si se informó, tiene que ser
  una de las cuatro y exacta (`Must`, `MUST`, `" must "` y `urgente` se rechazan con
  `ErrPrioridadInvalida`, envuelto con el valor). La regla es la de RN-005-5 y está
  escrita **una sola vez** (`ValidarPrioridad`, que `validarDatos` pasa a usar). La
  diferencia con el alta es que acá la prioridad **vacía** es "sin filtro". Una
  prioridad de solo espacios es inválida.
- **RN-007-9** (aplicación) — Los chequeos se hacen en este orden: proyecto, estado y
  prioridad. Un proyecto inexistente se devuelve **solo**, aunque los filtros
  también sean inválidos. Con el proyecto existente, el estado y la prioridad se
  validan **los dos** y, si ambos son inválidos, se devuelven unidos con
  `errors.Join` en el orden estado, prioridad; `errors.Is` reconoce cada uno. Con
  algún error no se lee la lista. Es el criterio de US-005 (RN-005-9 y RN-005-13): la
  existencia se verifica antes que los datos y no se une a sus errores, y los
  errores de validación de una misma entrada se devuelven todos juntos.
- **RN-007-10** (aplicación) — La consulta no modifica nada: ni las historias ni los
  sprints a los que están asignadas.
- **RN-007-11** (aplicación) — El filtro mira el estado de la historia y nada más.
  Estar asignada a un sprint, abierto o cerrado, o no estarlo, no es un estado
  (RN-005-6) y no cambia si la historia sale o no.

## 5. Restricciones

- **Dominio puro.** `internal/domain/backlog` suma `ValidarPrioridad`, `ValidarEstado`
  y `ErrEstadoInvalido` usando solo la biblioteca estándar, sin SQL, HTTP, IDs ni
  `time.Now()`. `validarDatos` pasa a llamar a `ValidarPrioridad`: es un refactor sin
  cambio de comportamiento, con el mismo error, el mismo mensaje y el mismo orden de
  errores unidos (RN-005-13), y los tests de US-005 y US-006 siguen en verde.
- **Caso de uso.** `app.ListarBacklog` conserva su nombre, su constructor
  (`NuevoListarBacklog`) y su resultado, como pide T-011; solo suma el parámetro
  `filtro` a `Ejecutar`. Hay que ajustar sus llamadores en el PR de implementación
  para que pasen un `FiltroBacklog{}` vacío: la pantalla del backlog
  (`pantallas.go`), la lista de historias de la pantalla de defectos y la de la
  pantalla de esfuerzo, con sus tests. También esa línea de la sección 3 de la spec
  de T-011 (`ListarBacklog(proyectoID)`), sin cambiar nada más de ella. Las specs de
  T-019, T-021 y US-033 solo nombran a `ListarBacklog`: no cambian.
- **Puerto.** `app.RepositorioListarBacklog` no cambia (`ExisteProyecto` y
  `ListarItems`). Esta historia no toca ningún adaptador de persistencia.
- **El filtro se aplica en la aplicación, no en SQL.** El repositorio sigue
  devolviendo todas las historias del proyecto sin ordenar y el caso de uso filtra y
  ordena, como ya hace con el orden (T-011) y como hacen US-020 y US-024. Así la
  memoria y Postgres no pueden filtrar distinto.
- **Sin migración.** No se agregan columnas ni índices: `items_backlog` ya tiene
  `estado` y `prioridad`.
- **Catálogo de errores (T-005).** `backlog.ErrEstadoInvalido` es el único error
  nuevo y se registra en `internal/adapters/http/catalogo.go` como `validacion` con el
  campo `estado` (ya figura en la lista cerrada de RN-T005-8), y se suma a la tabla
  de la sección 7 de la spec de T-005, en el mismo PR. `backlog.ErrPrioridadInvalida`
  (campo `prioridad`) y `app.ErrProyectoInexistente` (`no_encontrado`) ya están. Si se
  olvida el registro, el test del catálogo falla y nombra el error (RN-T005-19).
- **`en_progreso` no se alcanza hoy.** Ninguna operación de la aplicación pasa una
  historia a ese estado (US-004, Hueco 3; US-010, «Fuera de alcance»). El filtro lo
  acepta, porque es un estado del modelo (RN-005-6), pero hoy siempre da una lista
  vacía. Ningún escenario BDD puede armar una historia `en_progreso`: ese caso lo
  cubren los tests del caso de uso y de la base en memoria, con una historia armada
  directamente en el repositorio.
- **Steps.** No crea un área: suma a la interfaz `Backlog` un método que lista con los
  filtros y otro que informa qué historias tiene la última lista, y estas frases
  (ADR 0002), que se aprueban en el PR de implementación y se suman a
  `docs/diccionario-steps.md`, donde el área `Backlog` pasa a figurar también con
  US-007:
  - `Cuando se lista el backlog`
  - `Cuando se lista el backlog con estado "pendiente"`
  - `Cuando se lista el backlog con prioridad "must"`
  - `Cuando se lista el backlog con estado "pendiente" y prioridad "must"`
  - `Y la lista del backlog tiene 2 historias`
  - `Y la lista del backlog incluye la historia "Crear proyecto"`
  - `Y la lista del backlog no incluye la historia "Crear proyecto"`
  - `Y la lista del backlog muestra las historias en el orden:` + tabla de una columna

  Las historias del escenario se arman con las frases de US-005; una historia
  `completada` se arma con las de US-009 y US-010 (`una historia "X" con prioridad
  "must", 5 story points y 2 criterios de aceptación`, un sprint,
  `la historia "X" asignada al sprint "Sprint 1"` y `la historia "X" completada`).
  El estado y la prioridad se mandan tal cual, para poder escribir `urgente` o
  `Pendiente`. Un proyecto inexistente se escribe con `Dado ningún proyecto`: sin
  proyecto actual, el step manda un ID que no existe y el rechazo lo produce la
  aplicación, igual que `CrearHistoria` en US-005. Los rechazos se afirman con
  `la operación se rechaza` y `el mensaje de error indica "..."`, que ya existen. Que
  se devuelva **solo** el error del proyecto (CA-007-6) no se puede afirmar con esas
  frases (el mensaje se compara por contenido, no por exclusión): lo prueba el caso
  de uso con `errors.Is`.
- **Fuera de alcance.** Los controles del filtro en la pantalla del backlog (Hueco 1);
  filtrar por sprint, por Story Points, por título o por texto; elegir varios
  estados o varias prioridades a la vez; elegir el orden; paginar; mostrar totales;
  recordar el filtro entre visitas; y crear, editar o estimar historias (US-005,
  US-006, US-013).

## 6. Casos límite

- **CL-007-1** (RN-007-1) — Sin filtros, o con los dos valores vacíos, la lista es
  igual a la de T-011: todas las historias, pendientes, completadas y de las cuatro
  prioridades, en el orden de siempre. Un solo valor vacío no filtra por esa
  dimensión y el otro se aplica solo (RN-007-4).
- **CL-007-2** (RN-007-2) — Historias pendientes y completadas, filtro por
  `pendiente`: salen solo las pendientes, de cualquier prioridad. Se prueba con los
  tres estados.
- **CL-007-3** (RN-007-2) — Una historia completada (US-010) sale al filtrar por
  `completado` y no al filtrar por `pendiente`. Si se desmarca (US-010), pasa a
  salir en el otro filtro: el filtro mira el estado actual.
- **CL-007-4** (RN-007-2, RN-007-5) — Filtro por `en_progreso`: es un estado válido,
  pero ninguna historia lo tiene hoy, así que la lista es vacía y sin error. Con una
  historia `en_progreso` armada en el repositorio, sale.
- **CL-007-5** (RN-007-3) — Filtro por `must`: salen solo las historias `must`, de
  todos los estados. Se prueba con cada una de las cuatro prioridades.
- **CL-007-6** (RN-007-4) — Filtro por `pendiente` y por `must`: sale solo la
  historia pendiente de prioridad `must`. No salen ni la pendiente `could` ni la
  completada `must`.
- **CL-007-7** (RN-007-5) — Estado válido sin historias (por ejemplo `completado` en
  un backlog donde todo está pendiente): lista vacía, sin error.
- **CL-007-8** (RN-007-5) — Prioridad válida sin historias (por ejemplo `wont` en un
  backlog que solo tiene `must` y `could`): lista vacía, sin error.
- **CL-007-9** (RN-007-5) — Hay historias pendientes y hay historias `wont`, pero
  ninguna pendiente es `wont`: la combinación da una lista vacía, sin error.
- **CL-007-10** (RN-007-5) — Proyecto sin ninguna historia, con filtros válidos o sin
  filtros: lista vacía, sin error.
- **CL-007-11** (RN-007-6) — Historias `should` con los números 4, 2 y 7, filtro por
  `should`: salen 2, 4 y 7. Con el filtro por `pendiente` salen primero las `must`,
  después las `should`, etc., y dentro de cada prioridad por número. El filtro no
  reordena.
- **CL-007-12** (RN-007-6) — Una historia filtrada tiene los mismos datos que en el
  backlog completo: el mismo número, título, descripción, prioridad, estado, Story
  Points (o sin estimar, que no es `0`) y los mismos criterios, en el mismo orden.
- **CL-007-13** (RN-007-11) — Una historia asignada a un sprint abierto o cerrado sale
  igual que una que no lo está, según su estado. Asignar o desasignar no la saca de
  la lista filtrada.
- **CL-007-14** (RN-007-7) — Estado `Pendiente`, `COMPLETADO`, `en progreso`,
  `" pendiente "`, un solo espacio o `terminado`: `ErrEstadoInvalido`, envuelto con
  el valor recibido. El estado vacío no está en esta lista: es "sin filtro"
  (CL-007-1).
- **CL-007-15** (RN-007-8) — Prioridad `Must`, `MUST`, `" must "`, un solo espacio o
  `urgente`: `ErrPrioridadInvalida`, envuelto con el valor recibido. La prioridad
  vacía no está en esta lista: es "sin filtro" (CL-007-1).
- **CL-007-16** (RN-007-9) — Estado `terminado` y prioridad `urgente` a la vez: se
  devuelven los dos errores unidos, `ErrEstadoInvalido` primero y
  `ErrPrioridadInvalida` después. `errors.Is` reconoce cada uno, y el mensaje es la
  unión de los dos, uno por línea.
- **CL-007-17** (RN-007-9) — Estado inválido con prioridad válida: solo
  `ErrEstadoInvalido`. Estado válido con prioridad inválida: solo
  `ErrPrioridadInvalida`.
- **CL-007-18** (RN-007-9) — Proyecto inexistente (o con un ID mal formado), con un
  estado y una prioridad también inválidos: se devuelve solo
  `ErrProyectoInexistente`, y `errors.Is` no reconoce ninguno de los otros dos.
- **CL-007-19** (RN-007-9, RN-007-10) — Cualquier rechazo ocurre antes de leer la
  lista y no modifica nada.
- **CL-007-20** (RN-007-9) — Falla inesperada de la base al verificar el proyecto o al
  listar, con o sin filtro: se devuelve el error envuelto con su contexto y ninguna
  lista parcial.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario | Origen |
|---|---|---|---|
| El proyecto no existe | `app.ErrProyectoInexistente`, envuelto con el ID | "el proyecto no existe" | Aplicación |
| El estado del filtro no es `pendiente`, `en_progreso` ni `completado` | `backlog.ErrEstadoInvalido` (**nuevo**), envuelto con el valor recibido | "el estado tiene que ser pendiente, en_progreso o completado" | Dominio |
| La prioridad del filtro no es `must`, `should`, `could` ni `wont` | `backlog.ErrPrioridadInvalida`, envuelto con el valor recibido | "la prioridad tiene que ser must, should, could o wont" | Dominio |

El proyecto inexistente se detecta primero y se devuelve solo. Los otros dos se
detectan **antes** de leer la lista y, si los dos fallan, se devuelven unidos con
`errors.Join`, estado primero (RN-007-9). Son valores de paquete comparables con
`errors.Is`, también cuando vienen unidos. Que el filtro no tenga coincidencias no es
un error (RN-007-5).

Los dos de validación se clasifican como `validacion` (T-005): `ErrEstadoInvalido`
con el campo `estado` y `ErrPrioridadInvalida` con el campo `prioridad`;
`ErrProyectoInexistente`, como `no_encontrado`. Cómo se muestran en la pantalla es
tarea del adaptador, fuera de esta historia.

## 8. Criterios de aceptación

Los criterios CA-007-1 y CA-007-2 del issue #17 estaban sin texto (el primero decía
«completar en Refinement»); los de esta sección son los propuestos para el issue, que
hay que corregir a mano.

- **CA-007-1** (normal) — Con historias de distintos estados y prioridades, filtrar
  por un estado devuelve solo las historias en ese estado, y filtrar por una
  prioridad devuelve solo las de esa prioridad, en el orden del backlog.
- **CA-007-2** (alternativo) — Los dos filtros se combinan: filtrar por un estado y
  por una prioridad devuelve solo las historias que cumplen ambos.
- **CA-007-3** (alternativo) — Sin filtros se devuelve el backlog completo del
  proyecto, ordenado por prioridad y, dentro de cada una, por número.
- **CA-007-4** (límite) — Un filtro sin coincidencias (un estado sin historias, una
  prioridad sin historias o una combinación sin intersección) devuelve una lista
  vacía, no un error.
- **CA-007-5** (error) — Un estado que no es `pendiente`, `en_progreso` ni
  `completado` se rechaza con `ErrEstadoInvalido`, y una prioridad que no es
  `must`, `should`, `could` ni `wont` (por ejemplo `urgente` o `Must`), con
  `ErrPrioridadInvalida`. Con los dos inválidos a la vez se informan los dos.
- **CA-007-6** (error) — Un proyecto inexistente se rechaza con
  `ErrProyectoInexistente`, y solo con ese error, aunque los filtros también sean
  inválidos.

---

## Decisiones tomadas y descartadas

- **Se extiende `ListarBacklog`; no se crea otro caso de uso.** T-011 lo dejó
  previsto al marcar este filtro como fuera de su alcance. Se descartó un
  `ListarBacklogFiltrado` aparte: serían dos lecturas del mismo backlog que pueden
  divergir, y las tres pantallas que ya llaman a `ListarBacklog` tendrían que
  elegir entre ellas. Costo aceptado: las tres llaman ahora con un filtro vacío.
- **El filtro es un struct con el valor vacío como "sin filtro".** Es el criterio de
  `FiltroEsfuerzo` (US-020) y de `FiltroDefectos` (US-024). Se descartaron los
  punteros o un tipo opcional: para dos cadenas, el cero de Go ya dice "no
  informado".
- **Los filtros se aplican en la aplicación, no en SQL.** Mismo motivo que el orden en
  T-011 y que US-024: un único código para la memoria y Postgres, así no pueden
  filtrar distinto. Se descartó un `WHERE` en la consulta: sumaría parámetros al
  puerto y dos implementaciones del mismo filtro. Costo aceptado: cada pedido lee
  todas las historias del proyecto, como ya hace el listado de T-011; con los
  volúmenes de la cursada (decenas de historias) no se nota.
- **Dos reglas extraídas al dominio: `ValidarPrioridad` y `ValidarEstado`.** Repetir
  el `switch` de la prioridad en el caso de uso dejaría la regla de RN-005-5 escrita
  dos veces. Se descartó tratar el valor vacío como inválido, como en el alta: no
  habría forma de pedir "todas las prioridades". Esa diferencia vive en el caso de
  uso, no en el dominio.
- **Error nuevo `backlog.ErrEstadoInvalido`; la prioridad reutiliza el que ya
  existe.** No había ningún error de estado de historia inválido. Se descartó
  reutilizar `defecto.ErrEstadoInvalido`: es de otro agregado, nombra otros cuatro
  estados y su mensaje sería falso para una historia. Se descartó un
  `ErrFiltroInvalido` genérico: T-005 clasifica por campo, y con `estado` y
  `prioridad` el mensaje cae en el campo correcto.
- **Con los dos filtros inválidos se informan los dos, unidos.** Es la convención de
  los errores de validación de una misma entrada (RN-005-13, RN-022-4): quien
  filtra ve de una vez todo lo que tiene que corregir. Es distinto de US-024, que
  devuelve solo el error del sprint: allí uno de los errores era de **existencia**
  y el otro de validación, y la existencia va primero y sola (RN-005-9). Acá los dos
  son de validación, así que se unen. El proyecto inexistente sigue yendo solo.
  Se descartó cortar en el primer filtro inválido, que obligaría a corregir de a
  uno.
- **Estado primero, prioridad después.** Es el orden del título de la historia
  («por estado y prioridad»); hace que el mensaje y los tests sean deterministas.
- **El valor vacío es "sin filtro" y no se normaliza.** `Must` o `" pendiente "` son
  inválidos, igual que en el alta (RN-005-5) y que en US-024. Se descartó aceptarlos
  por ser un filtro: una regla de prioridad que se escribe una sola vez no admite
  excepciones por lugar de uso.
- **Un solo valor por filtro, combinados con Y.** Es el criterio de US-020 y de
  US-024. Se descartó elegir varios estados o varias prioridades a la vez: ningún
  criterio lo pide y complica el contrato.
- **El listado no oculta las historias completadas.** Sin filtros salen todas, como
  en T-011. Se descartó esconder las completadas por defecto: cambiaría lo que ya
  hace la pantalla del backlog y las de defectos y esfuerzo, que arman sus listas
  desplegables con este mismo caso de uso.
- **El orden no cambia con el filtro.** Prioridad y número, siempre (RN-T011-4). Se
  descartó ordenar por el campo filtrado: dos criterios de orden que dependen del
  filtro harían que una misma historia aparezca en lugares distintos según cómo se
  llegó a ella.
- **`en_progreso` es un filtro válido aunque hoy dé siempre vacío.** Es parte del
  modelo (RN-005-6) y de la validación; rechazarlo habría que cambiarlo el día que
  una historia pueda estar en progreso. Se descartó quitarlo de la lista de valores
  válidos para no inventar una regla que nadie pidió.
- **Sin cambios de migración ni de adaptadores de persistencia.** La lectura ya
  existe. Todo lo nuevo es una firma, dos funciones extraídas, un error y su entrada
  de catálogo.

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia:

1. **Los controles del filtro en la pantalla.** Las pantallas quedan fuera de las
   historias, así que esta no agrega un formulario a `/proyectos/{id}/backlog`. Para
   usar el filtro desde el navegador hace falta una tarea técnica aparte (o ampliar
   T-011). Se propone lo de T-019 y US-024: un formulario `GET` común con los
   parámetros `estado` y `prioridad` (valores vacíos no filtran), con la URL
   compartible y sin JavaScript. Quedarían por fijar los rótulos de la opción para
   no filtrar, el texto de la lista vacía por filtro (T-011 usa «El backlog está
   vacío»), cómo se muestran los errores y si se conserva lo elegido. Mientras tanto,
   US-007 se demuestra por el caso de uso y los escenarios BDD. El issue #17 lleva la
   etiqueta `area:ui`, que no coincide con este alcance: conviene quitarla o crear la
   tarea de pantalla.
2. **`en_progreso` no se alcanza.** El filtro lo acepta y siempre da lista vacía
   hasta que alguna historia pueda pasar a ese estado. Confirmar que se deja como
   valor válido (propuesto) y no se rechaza.
3. **Valor vacío es "sin filtro" y no se normaliza.** Confirmar que no se quiere
   aceptar `Must` ni `" pendiente "` por ser un filtro.
4. **Alcance del filtro.** Se filtra solo por estado y por prioridad, de a un valor.
   Filtrar por sprint (por ejemplo «las que no están en ningún sprint»), por Story
   Points o por texto sería otra historia. Confirmar que no hace falta ahora.
5. **Criterios del issue #17 y la frase «Como / quiero / para».** El issue no trae
   esa frase (está redactada para esta spec) y sus dos criterios están sin texto.
   Los de la sección 8 son los propuestos y hay que cargarlos a mano en el issue.
   También sigue diciendo «Sprint tentativo: 2», y la historia hoy no tiene sprint.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code, leyendo el issue
  #17, las specs de T-011, US-005, US-006, US-010 y US-004, las de US-024 y US-020
  como modelo de filtros, y el código de `ListarBacklog`, del dominio `backlog` y
  del catálogo de errores; las decisiones sobre la unión de errores, el valor
  vacío, el error nuevo y el alcance de la pantalla se confirman en la revisión
- Revisado y entendido por:
