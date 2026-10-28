# T-021 — Pantallas de defectos: registrar, listar y cambiar el estado

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Angelo Conforti |
| **Issue** | #191 (tarea) · issue de la spec: pendiente |
| **Escenarios BDD** | ninguno: es una tarea técnica (ver "Decisiones") |
| **Código** | `internal/app/`, `internal/adapters/memoria/`, `internal/adapters/postgres/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| **Última actualización** | 2026-10-27 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Que una persona registre defectos desde el navegador, los vea en una lista del proyecto y
cambie su estado (pasarlo a en corrección, resolverlo con su sprint de resolución, descartarlo
o reabrirlo). Cubre la parte de interfaz del requerimiento 6 de la guía (defectos).

Los dos casos de uso de escritura ya están en `main` y con sus tests: **US-022** registra el
defecto y **US-023** cambia su estado. Lo único que no existe es **la lectura para listarlos**,
así que esta tarea define un caso de uso de lectura nuevo (`ListarDefectos`, con su puerto y sus
dos implementaciones, como hizo T-015 con `ListarSprints`). No agrega reglas de dominio ni
migraciones: son rutas, un formulario, una lista con botones por fila y esa lectura.

Se apoya en T-004 (layout, router y `Responder`), T-005 (errores uniformes y lector de
formularios) y en el patrón de T-011, T-015 y T-019. Filtrar la lista por sprint y severidad es
US-024 (objetivo estirado del Sprint 5): extiende la lectura de esta tarea, no la reemplaza.

## 2. Entradas

| Ruta | Método | Campos | Caso de uso |
|---|---|---|---|
| `/proyectos/{id}/defectos` | GET | — | `ListarDefectos` (nuevo) |
| `/proyectos/{id}/defectos` | POST | `descripcion`, `severidad`, `historia`, `sprint_deteccion` | `RegistrarDefecto` (US-022) |
| `/proyectos/{id}/defectos/{defecto}/estado` | POST | `estado`, `sprint_resolucion` | `CambiarEstadoDefecto` (US-023) |

- `{id}` y `{defecto}` son los IDs que genera la base. `historia`, `sprint_deteccion` y
  `sprint_resolucion` son IDs: la persona los elige de listas que muestran la historia (`#N título`)
  y el sprint (`Sprint N`).
- `descripcion` es un texto de varias líneas (`<textarea>`). `severidad` es una lista con las
  cuatro severidades y una opción inicial sin valor («Elegí una severidad»); el valor de cada
  opción es el del dominio (`critica`, `alta`, `media`, `baja`) y la etiqueta, la de la
  persona («Crítica», «Alta», «Media», «Baja»). Esa conversión es del adaptador HTTP
  (US-022, sección 2).
- `estado` viaja en un campo oculto de cada formulario de la fila, con el valor del dominio
  (`abierto`, `en_correccion`, `resuelto` o `descartado`). `sprint_resolucion` solo lo manda el
  formulario de «Resolver»; en los demás no viaja y es «sin sprint de resolución» (US-023, sección 2).
- Los formularios se envían con HTMX (`hx-post`), igual que los de T-011, T-015 y T-019.
- Las listas de la pantalla salen de lecturas que ya existen: el proyecto de `ConsultarProyecto`,
  las historias de `ListarBacklog` y los sprints de `ListarSprints` (T-011 y T-015).

## 3. Salidas esperadas

**Caso de uso de lectura nuevo** (`internal/app/`), con su puerto:

```
ListarDefectos(proyectoID) -> []DefectoListado
```

```
DefectoListado { Defecto defecto.Defecto, HistoriaNumero int, HistoriaTitulo string,
                 SprintDeteccion int, SprintResolucion int, CreadoEn time.Time }
```

- `SprintDeteccion` y `SprintResolucion` son el número del sprint (`sprint.Numero`) traducido
  desde el ID guardado, como pide RN-022-6; `SprintResolucion` es `0` si el defecto no tiene
  sprint de resolución. `HistoriaNumero` y `HistoriaTitulo` son los de la historia relacionada.
- `CreadoEn` es la fecha de carga (`creado_en`). Viaja en el modelo de lectura para ordenar y la
  lista **no** la muestra. No se agrega al tipo `Defecto` del dominio.
- Como en T-011 y T-015, el repositorio devuelve los datos sin ordenar y **el orden lo pone el
  caso de uso** (RN-T021-4).

**Después de guardar con éxito:** `HX-Redirect` con HTMX o `303` sin él, hacia
`/proyectos/{id}/defectos`, tanto al registrar como al cambiar el estado. Los errores se
responden con `ResponderError` (T-005): el mensaje cae en `#mensajes`, el formulario no se toca y
no se registra ni se cambia nada.

**`/proyectos/{id}/defectos`** tiene, en este orden:

| Parte | Qué muestra |
|---|---|
| Lista de defectos | Tabla con descripción, severidad, estado, historia, sprint de detección y sprint de resolución, del más reciente al más antiguo; en cada fila, los botones de cambio de estado. Sin defectos: «Todavía no hay defectos» |
| Formulario «Registrar defecto» | Descripción, severidad, historia (lista con las del backlog) y sprint de detección (lista con todos los del proyecto) |

**Cada fila de la lista:**

| Columna | Qué muestra |
|---|---|
| Descripción | El texto completo, respetando los saltos de línea |
| Severidad | «Crítica», «Alta», «Media» o «Baja» |
| Estado | «Abierto», «En corrección», «Resuelto» o «Descartado» |
| Historia | `#N título` |
| Sprint de detección | `Sprint N` |
| Sprint de resolución | `Sprint N`, o «—» si no tiene |

**Botones de cada fila, según el estado que muestra la fila** (solo las transiciones válidas de
US-023, RN-023-2):

| Estado de la fila | Botones |
|---|---|
| Abierto | «En corrección» · «Resolver» (con una lista de sprint de resolución) · «Descartar» |
| En corrección | «Volver a abierto» · «Resolver» (con una lista de sprint de resolución) · «Descartar» |
| Resuelto | «Reabrir» |
| Descartado | «Reabrir» |

**Menú del proyecto:** el item «Defectos» ya existe y apunta a `/proyectos/{id}/defectos`, que hoy
responde 404; pasa a mostrar la lista y queda activo.

## 4. Reglas de negocio

No hay reglas de dominio nuevas. Son reglas de comportamiento de la interfaz y de la lectura.

**Capas**

- **RN-T021-1** — Un handler solo lee la petición, llama a un caso de uso y escribe la respuesta.
  No valida reglas de negocio ni habla con un repositorio (RN-T011-1).
- **RN-T021-2** — Los casos de uso de escritura no cambian: se usan `RegistrarDefecto` y
  `CambiarEstadoDefecto`, más `ConsultarProyecto`, `ListarBacklog` y `ListarSprints` para llenar las
  listas. Lo único nuevo es `ListarDefectos` (`internal/app`) con su puerto
  `RepositorioListarDefectos`, que cumplen igual la memoria y Postgres (RN-T011-2). Los cuatro
  casos de uso se agregan a `CasosDeUso` y se conectan en `cmd/server`.
- **RN-T021-3** — `ListarDefectos` devuelve los defectos del proyecto con los datos de
  `DefectoListado` (sección 3), en cualquier estado, incluidos los descartados. Un proyecto que
  no existe es `ErrProyectoInexistente`; un proyecto sin defectos, una lista vacía. No modifica nada.
- **RN-T021-4** — Los defectos se listan por `CreadoEn` descendente (el más reciente primero) y,
  con la misma fecha de carga, por ID descendente, para que el orden sea siempre el mismo.

**Formularios**

- **RN-T021-5** — `historia` y `sprint_deteccion` son listas obligatorias. Si llegan vacías, se
  informa en su campo «Elegí la historia relacionada» o «Elegí el sprint de detección» y no se llama
  al caso de uso. Los nombres de campo `historia` y `sprint_deteccion` se suman a la lista cerrada
  de RN-T005-8 y a la de `TestCatalogo_EntradasValidas`, sin tocar ninguna otra entrada, como
  hicieron US-022, US-023 y T-019. Estos dos textos son los únicos que no salen del catálogo.
- **RN-T021-6** — `descripcion` y `severidad` se pasan a `RegistrarDefecto` tal como llegan: el
  recorte, el largo de 1 a 2000 caracteres y las cuatro severidades los decide el dominio de US-022
  (RN-022-1, RN-022-2). Sus errores salen del catálogo, con el campo `descripcion` o `severidad`, y
  si hay dos se informan juntos y en ese orden (RN-022-4). Un valor de `severidad` que no es una de
  las cuatro (la opción inicial, o uno mandado a mano) es `ErrSeveridadInvalida`.
- **RN-T021-7** — `estado` y `sprint_resolucion` se pasan a `CambiarEstadoDefecto` tal como llegan: un
  estado fuera de los cuatro, una transición no permitida, un sprint de resolución que falta o que
  sobra, o uno anterior al de detección los decide US-023 (RN-023-1 a RN-023-4) y se informan con el
  mensaje del catálogo, en el campo `estado` o `sprint_resolucion`. Un `sprint_resolucion` vacío es
  «sin sprint de resolución» (US-023, sección 2).
- **RN-T021-8** — Los errores siguen el mecanismo de T-011 y T-019: el fragmento de T-005, con el
  campo en `data-campo`, cae en `#mensajes`. Los mensajes del catálogo y los dos de RN-T021-5 nombran
  el campo al que corresponden. El formulario no se redibuja y conserva lo escrito.
- **RN-T021-9** — Un texto con HTML (la descripción, el título de una historia) se muestra escapado
  en cualquier pantalla y se conserva tal cual en el formulario (RN-T011-8). La descripción se
  muestra completa, con los saltos de línea que tenga.

**Alcance de las rutas**

- **RN-T021-10** — Un proyecto que no existe o con ID mal formado responde 404: página con layout
  en un GET, fragmento de error en un POST (RN-T011-10).
- **RN-T021-11** — Un defecto solo se alcanza por la ruta de **su** proyecto. Un defecto de otro
  proyecto, inexistente o con ID mal formado responde `ErrDefectoInexistente` (404 en un POST:
  fragmento de error), aunque exista en otro lado (RN-023-7).
- **RN-T021-12** — Una historia, un sprint de detección o un sprint de resolución **informados** que
  son de otro proyecto o no existen se rechazan con el error de inexistencia del catálogo (404 en un
  POST: fragmento de error), sin sumarse a los errores de validación (RN-022-5, RN-023-7).

**Lista y acciones**

- **RN-T021-13** — Las columnas, su orden y su formato son los de la sección 3. Las etiquetas de
  severidad y de estado las pone el adaptador HTTP; el dominio y la base guardan los valores sin
  tilde. El sprint de resolución vacío se muestra como «—».
- **RN-T021-14** — Los botones de una fila son los de la tabla de la sección 3 para el estado que
  la fila muestra. La pantalla no protege la regla: si el defecto ya cambió (otra pestaña, otra
  persona), el caso de uso evalúa el pedido contra el estado guardado y responde su error
  (CL-023-11), igual que RN-T015-7.
- **RN-T021-15** — La lista de sprint de resolución de «Resolver» y la de sprint de detección del
  formulario muestran **todos** los sprints del proyecto, abiertos y cerrados, sin filtrar por el
  sprint de detección del defecto (RN-022-6, RN-023-9). Así el error de un sprint de resolución
  anterior al de detección se puede producir desde la pantalla. Cada una empieza con una opción sin
  valor («Elegí un sprint»).
- **RN-T021-16** — Las listas se ordenan así: historias en el orden del backlog (RN-T011-4) y sprints
  por número descendente (RN-T015-3).
- **RN-T021-17** — Cada campo tiene su `<label>`, y cada control de la tabla tiene un nombre
  accesible que dice a qué defecto corresponde: la acción y el comienzo de su descripción (hasta 40
  caracteres), porque el defecto no tiene número ni título (RN-T011-9).
- **RN-T021-18** — El menú del proyecto muestra «Defectos» activo en esta página.

## 5. Restricciones

- **Sin cambios de dominio.** No se toca `internal/domain`. `Defecto` no recibe `CreadoEn`; viaja
  solo en el modelo de lectura.
- **Sin migración.** La tabla `defectos` ya tiene `creado_en` (`00012`). Las lecturas no
  modifican nada y se hacen sin transacción.
- **La lectura nueva.** `RepositorioListarDefectos` tiene `ExisteProyecto` y una consulta que
  devuelve los `DefectoListado` del proyecto sin ordenar. En Postgres, una sola consulta que trae la
  historia y los números de sprint con `JOIN`. La memoria guarda el instante de alta de cada defecto
  (con el reloj del sistema, que en un adaptador sí se puede usar) y lo devuelve creciente entre
  altas, para que su orden sea el de Postgres. El nombre `ListarDefectos` y su resultado no cambian
  cuando US-024 le agregue los filtros.
- **Catálogo de errores (T-005).** Todos los errores que aparecen ya están registrados (US-022,
  US-023 y `app`). Lo único que se suma son los dos nombres de campo de RN-T021-5. Si T-019 llega
  antes a `main` con `historia` en la lista cerrada, el PR de implementación conserva las dos
  partes y no la repite.
- **Los formularios requieren JavaScript (HTMX),** como en T-011, T-015 y T-019.
- **Sin paginación,** como T-015. El diseño visual es funcional (Tailwind); se mejora después.
- **Fuera de alcance:** filtrar la lista por sprint o severidad (US-024), una página de detalle de
  un defecto, editar o borrar un defecto (ninguna historia lo pide), el historial de cambios,
  guardar quién registró o cambió un defecto (la aplicación no tiene usuarios), un identificador
  legible `DEF-NNN` y los defectos por sprint del dashboard (US-030).

## 6. Casos límite

- **CL-T021-1** — Proyecto sin defectos: en lugar de la tabla, el texto «Todavía no hay defectos»; el
  formulario se muestra igual y no hay error.
- **CL-T021-2** — Proyecto sin historias o sin sprints: el formulario muestra la lista vacía (solo
  la opción inicial) y un aviso («Primero cargá una historia» / «Primero creá un sprint»); registrar se
  rechaza con el error del campo `historia` o `sprint_deteccion` (RN-T021-5).
- **CL-T021-3** — Descripción con varias líneas: se guarda y se muestra con sus saltos de línea. De
  2000 caracteres se acepta y de 2001 se rechaza con `ErrDescripcionMuyLarga`; vacía o con solo
  espacios, con `ErrDescripcionVacia` (CL-022-1 a CL-022-4).
- **CL-T021-4** — Severidad sin elegir (la opción inicial) o un valor mandado a mano distinto de las
  cuatro: `ErrSeveridadInvalida` en el campo `severidad`.
- **CL-T021-5** — Descripción vacía y severidad sin elegir a la vez: los dos errores juntos, en ese
  orden (CL-022-6). Historia y sprint de detección sin elegir a la vez: los dos mensajes de RN-T021-5
  juntos, sin llamar al caso de uso; los errores de descripción y severidad no se informan hasta que
  los dos estén elegidos.
- **CL-T021-6** — Sprint de detección cerrado, anterior o distinto del sprint de la historia: se
  acepta (CL-022-12). Historia completada o que no está en ningún sprint: se acepta y la historia no
  cambia (CL-022-11).
- **CL-T021-7** — Dos defectos con los mismos datos: se registran los dos y aparecen en dos filas
  (CL-022-13).
- **CL-T021-8** — Dos defectos con la misma fecha de carga: salen siempre en el mismo orden, por ID
  descendente (RN-T021-4). La fecha de carga no se muestra.
- **CL-T021-9** — Resolver un defecto abierto con el Sprint 2: queda «Resuelto», la columna de
  resolución muestra «Sprint 2» y la fila solo ofrece «Reabrir». Resolver sin elegir sprint:
  `ErrSprintResolucionFaltante` en `sprint_resolucion`. Resolver en un sprint cerrado o en uno posterior
  al de detección: se acepta (CL-023-5, CL-023-6).
- **CL-T021-10** — Resolver con un sprint anterior al de detección (la lista lo ofrece, RN-T021-15):
  `ErrSprintResolucionAnterior` en `sprint_resolucion`, con 409, y el defecto no cambia (CL-023-5).
- **CL-T021-11** — Reabrir un defecto resuelto o descartado: vuelve a «Abierto» y a «—» en el sprint
  de resolución. Resolverlo de nuevo con otro sprint deja el nuevo (CL-023-7). «Volver a abierto»
  sobre uno en corrección: queda «Abierto».
- **CL-T021-12** — Página desactualizada: la fila muestra «Abierto» pero otra persona ya lo resolvió y
  se pide «Descartar»; o se pide «Volver a abierto» sobre uno que ya está abierto. Se rechaza con
  `ErrTransicionInvalida` en el campo `estado` (409) y el defecto no cambia; pedir «Reabrir» sobre el
  resuelto sí es una reapertura válida (CL-023-2, CL-023-11).
- **CL-T021-13** — `estado` mandado a mano fuera de los cuatro (`cerrado`, vacío): `ErrEstadoInvalido`
  (422). `sprint_resolucion` informado con un estado distinto de `resuelto`:
  `ErrSprintResolucionNoCorresponde` (422) y el defecto no cambia (CL-023-3, CL-023-4).
- **CL-T021-14** — Dos cambios simultáneos del mismo defecto: el segundo responde `ErrDefectoModificado`
  (409, sin campo) y no cambia nada; la página queda como estaba y se actualiza al recargar (CL-023-14).
- **CL-T021-15** — Historia, sprint de detección o sprint de resolución de otro proyecto o inexistentes, y
  un defecto de otro proyecto, inexistente o con ID mal formado: 404, aunque existan por separado.
- **CL-T021-16** — Descripción con HTML, comillas, tildes o ñ (por ejemplo `<script>`): se guarda y se
  muestra escapada, igual que el título de la historia (RN-T021-9).
- **CL-T021-17** — Un defecto descartado sigue en la lista, con estado «Descartado» y «—» en el sprint
  de resolución: la lista no oculta ningún estado.
- **CL-T021-18** — Falla inesperada de la base al registrar, cambiar el estado o listar: 500 con el
  mensaje fijo de T-005, sin registrar ni cambiar nada.

## 7. Condiciones de error

| Situación | Respuesta |
|---|---|
| Proyecto inexistente o con ID mal formado | 404: página con layout (GET) o fragmento de error (POST) |
| `historia` o `sprint_deteccion` sin elegir | 422: error en el campo, sin llamar al caso de uso |
| Descripción vacía o de más de 2000 caracteres, o severidad inválida | 422: el mensaje de US-022 en el campo correspondiente, juntos |
| `estado` fuera de los cuatro, `resuelto` sin sprint de resolución o sprint con otro estado | 422: el mensaje de US-023 en el campo `estado` o `sprint_resolucion` |
| Transición no permitida o sprint de resolución anterior al de detección | 409: el mensaje de US-023 en el campo `estado` o `sprint_resolucion` |
| El defecto cambió entre la lectura y el guardado | 409: `ErrDefectoModificado`, sin campo |
| Historia, sprint o defecto de otro proyecto o inexistente | 404 |
| Falla inesperada (base caída, panic) | 500 con el mensaje fijo de T-005; el detalle solo va al registro |

## 8. Criterios de aceptación

- **CA-T021-1** (normal) — Con el router real y la base en memoria: un proyecto con una historia y un
  sprint; se registra un defecto con una descripción de varias líneas, severidad Alta, esa historia y ese
  sprint de detección. La respuesta es `HX-Redirect` con HTMX y `303` sin él; la lista muestra el
  defecto con su descripción, «Alta», «Abierto», su historia, su sprint de detección y «—» de resolución.
- **CA-T021-2** (alternativo) — La lista de defectos del proyecto muestra descripción, severidad, estado,
  historia, sprint de detección y sprint de resolución, del más reciente al más antiguo, con las
  etiquetas de la sección 3 y sin mostrar la fecha de carga.
- **CA-T021-3** (alternativo) — Cada fila ofrece solo las transiciones válidas desde su estado, y un defecto
  se puede pasar a en corrección, resolver (con su sprint de resolución), descartar y reabrir; cada cambio
  responde `HX-Redirect` con HTMX y `303` sin él, y la lista se actualiza con el estado y el sprint de
  resolución nuevos.
- **CA-T021-4** (límite) — Sin defectos hay un estado vacío; la lista de sprints del formulario y la de
  resolución ofrecen todos los del proyecto, y un sprint de detección cerrado se acepta; un proyecto sin
  historias o sin sprints muestra el aviso.
- **CA-T021-5** (error) — Una descripción vacía o de más de 2000 caracteres, una severidad inválida, una
  historia o un sprint de detección sin elegir, una transición no permitida o un sprint de resolución
  anterior al de detección muestran un error que nombra el campo, con el campo en `data-campo`, en
  `#mensajes`, y no registran ni cambian nada; una historia, un sprint o un defecto de otro proyecto
  responde 404.
- **CA-T021-6** (plataforma) — `ListarDefectos` devuelve lo mismo en memoria y en Postgres (test de
  integración con base real): los datos del defecto, los números de la historia y de los sprints, y el orden
  del más reciente al más antiguo. Las rutas usan el layout, el router y `Responder`, funcionan con y sin
  HTMX, el menú muestra «Defectos» activo, y con `make dev` en el navegador se recorre registrar, listar,
  resolver y reabrir.

---

## Decisiones tomadas y descartadas

- **Es una tarea y no una historia.** Los casos de uso de escritura y sus reglas ya están (US-022 y
  US-023); esto es la interfaz, como T-011, T-015, T-016 y T-019. Se estima en 5 SP. El issue proponía
  empezar por registrar y listar y sumar el cambio de estado cuando US-023 estuviera mergeada; como ya
  lo está (PR #192), la spec cubre las tres cosas juntas.
- **La lectura para listar es de esta tarea, y no de US-024.** US-022 dejó fuera listar (dice que es
  US-024), pero US-024 es el objetivo estirado y trata de filtrar. Sin una lectura no hay lista, y CA-T021-2
  la pide. Se define acá, con su puerto y sus dos implementaciones como `ListarSprints` (T-015), y US-024
  le suma los filtros sin cambiar su nombre ni su resultado. Se descartó esperar a US-024: puede no
  hacerse.
- **`creado_en` viaja en el modelo de lectura, no en `Defecto`.** El orden «más reciente primero» lo necesita
  y la lista no lo muestra (CA-T021-2 no tiene una columna de fecha). Se descartó agregarlo a `Defecto`: el
  dominio no lo usa y lo pone la base. Se descartó ordenar en la consulta: el orden lo pone el caso de uso
  (T-011, T-015).
- **Los errores siguen cayendo en `#mensajes`.** El criterio del issue decía «junto a su campo», pero T-011
  decidió que los errores no redibujan el formulario y T-019 usa el mismo mecanismo. Mostrar el mensaje pegado
  al campo exige una pieza nueva en el layout compartido, que cambia T-011, T-015 y T-019. Se descartó por
  alcance. Para ser fiel a lo que se hace, CA-T021-5 se reformuló: el error **nombra el campo** y lo marca con
  `data-campo`. Los mensajes del catálogo ya lo nombran («la descripcion es obligatoria», «la severidad tiene
  que ser…»).
- **Historia y sprint de detección sin elegir: chequeo de pantalla, no 404.** Para el dominio un ID vacío es
  inexistente (US-022), y el catálogo clasifica esos errores como `no_encontrado`, que por regla no lleva
  campo y responde 404. Mostrar un 404 a quien se olvidó de elegir una opción es mala respuesta. Se sigue
  el patrón de T-019 (RN-T019-4): la pantalla los informa en su campo y no llama al caso de uso. Cuesta dos
  textos fuera del catálogo y dos nombres nuevos en la lista cerrada. Se descartó un error nuevo en el
  dominio: la regla ya existe y no cambia.
- **Los botones ofrecen solo las transiciones válidas.** Es lo que ve una persona y evita probar a ciegas.
  El dominio igual decide contra el estado guardado, así que una página desactualizada produce el error de
  transición (CL-T021-12), que es la única forma de verlo desde la pantalla. Se descartó ofrecer los cuatro
  estados y dejar que el dominio rechace.
- **El sprint de resolución lista todos los del proyecto.** No se filtra por el sprint de detección: sería
  una regla de la pantalla que repite la del dominio (RN-023-4), y el error de CA-T021-5 dejaría de poder
  producirse. El sprint de detección también los lista todos, incluidos los cerrados (CA-T021-4, RN-022-6).
- **Acciones en la fila, sin página de detalle.** Se descartó una ruta de detalle: el defecto no tiene más
  datos que mostrar que la fila, y suma una ruta y un ir y volver.
- **Una sola ruta de cambio de estado.** `POST …/defectos/{defecto}/estado` con el campo `estado`, y no una
  ruta por acción como las de T-015 (`completar`, `reabrir`): allá eran dos casos de uso distintos y acá es
  uno solo, que recibe el estado pedido. Cada botón es un formulario con el estado en un campo oculto.
- **El botón hacia «abierto» se llama «Reabrir» desde resuelto y descartado,** como llama US-023 a esa
  reapertura. Desde en corrección no es una reapertura, así que se llama «Volver a abierto» (ver Huecos).
- **La severidad empieza con una opción vacía.** Se descartó una severidad elegida por defecto: la persona
  tiene que decidirla y el error de `severidad` sigue pudiendo producirse. La etiqueta con tilde es de la
  pantalla; el valor, el del dominio (US-022).
- **El nombre accesible usa el comienzo de la descripción.** Un defecto no tiene número ni título, y el
  identificador legible `DEF-NNN` quedó afuera en US-022 (hueco 3). Se descartó agregarlo acá: necesita
  una columna y un correlativo por proyecto.
- **Historia como `#N título`.** Es lo que identifica a una historia en el backlog (T-011). La descripción
  se muestra completa, respetando los saltos de línea, porque la plantilla de defectos pide esperado,
  observado y pasos en líneas separadas.
- **No hay `.feature`.** Es una tarea técnica y las reglas ya tienen los suyos (US-022 y US-023). Los
  recorridos de CA-T021-1 a CA-T021-5 son tests de Go con `httptest`, como en T-011, T-015 y T-019, y
  CA-T021-6 se cubre con el test de integración de la lectura.

## Huecos para decidir con el equipo

Ninguno impide implementar esta tarea:

1. **Texto de «Volver a abierto».** Desde en corrección, la transición a abierto no tiene nombre en US-023
   (solo llama «reapertura» a la de resuelto y descartado). «Volver a abierto» es una propuesta; si el equipo
   prefiere «Reabrir» en los tres estados, cambia esa tabla y RN-T021-14.
2. **Lista larga.** Sin paginación ni filtros, un proyecto con muchos defectos muestra todos. Los filtros son
   de US-024 y no hay una historia de paginación.
3. **Orden de llegada con T-019.** Las dos tareas suman `historia` a la lista cerrada del catálogo. El que
   llegue segundo conserva las dos partes (RN-T005-8); confirmar el orden de merge con Tomás.
4. **Página desactualizada.** Después de un `ErrDefectoModificado` la fila queda con los botones viejos hasta
   recargar, porque los errores no redibujan la página (T-011). Confirmar que alcanza.
5. **Reloj de la memoria.** La memoria necesita un instante de alta creciente para ordenar como Postgres. Cómo
   lo obtiene (reloj propio o un contador) es un detalle de la implementación mientras el orden sea el de alta.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir del issue #191, las specs de US-022
  y US-023 y las de T-011, T-015 y T-019 como modelo, y el código de las pantallas y del catálogo de errores;
  las decisiones se proponen acá y se confirman en la revisión
- Revisado y entendido por: Angelo Conforti
