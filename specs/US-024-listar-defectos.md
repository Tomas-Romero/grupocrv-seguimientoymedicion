# US-024 — Listar los defectos de un proyecto filtrando por sprint de detección y severidad

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Angelo Conforti |
| **Issue** | #32 (historia) · issue de la spec: pendiente |
| **Depende de** | US-022 (tabla `defectos`, severidades), US-023 (estados) y T-021 (la lista y `ListarDefectos` que esta historia extiende), las tres ya en `main`; y US-008 (`sprints`) |
| **Escenarios BDD** | `features/US-024-listar-defectos.feature` |
| **Código** | `internal/app/` (`ListarDefectos` y su filtro), `internal/domain/defecto/` (`ValidarSeveridad`), `internal/adapters/http/` y `web/templates/` (los controles del filtro en la pantalla de T-021) |
| **Última actualización** | 2026-10-28 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero listar los defectos de un proyecto
> filtrando por el sprint en que se detectaron y por su severidad, para
> encontrar enseguida los que me interesan (por ejemplo, los críticos de un
> sprint) sin recorrer la lista entera.

Agregar dos filtros opcionales a la lista de defectos de T-021: el **sprint de
detección** y la **severidad**. Con uno, con los dos a la vez (solo salen los
defectos que cumplen ambos) o con ninguno (la lista completa de T-021). Un filtro
sin coincidencias da una lista vacía, no un error; un sprint que no es del
proyecto o una severidad que no existe se rechazan con el error uniforme de
T-005.

Es una mejora de lectura: no crea ni cambia defectos. Extiende `ListarDefectos`
de T-021 sin cambiarle el nombre ni el resultado (T-021, sección 5), y el orden,
las columnas y los botones de cada fila siguen siendo los de esa tarea. Es el
objetivo estirado del Sprint 5: se toma solo si US-023 y T-021 están en `main`,
y ya lo están.

## 2. Entradas

El caso de uso es `app.ListarDefectos`, el de T-021, con un parámetro más:

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. Viene de la URL; lo verifica el caso de uso, nunca el dominio |
| `filtro.SprintDeteccionID` | UUID | No | Un sprint existente **del proyecto**, abierto o cerrado, tenga o no defectos. Vacío: no se filtra por sprint |
| `filtro.Severidad` | `defecto.Severidad` | No | Exactamente `critica`, `alta`, `media` o `baja`: en minúscula, sin tilde y sin espacios alrededor (RN-022-2). Vacía: no se filtra por severidad |

**Orden de verificación** (RN-024-9). El caso de uso sigue este orden y devuelve
el primer error que encuentra, sin unirlo con los demás:

1. Verifica que el proyecto exista (`ErrProyectoInexistente`).
2. Si se informó sprint, verifica que exista **dentro de ese proyecto**
   (`ErrSprintInexistente`).
3. Si se informó severidad, la valida con el dominio (`ErrSeveridadInvalida`).
4. Lee los defectos del proyecto, aplica los filtros y los ordena.

Un `filtro.SprintDeteccionID` mal formado es lo mismo que uno inexistente: el
repositorio responde que no existe, no que falló. Uno **vacío** no lo es: es
"sin filtro de sprint".

No son entradas: el estado, la historia, el sprint de resolución ni ningún otro
dato del defecto. La URL de la pantalla y la forma en que la persona elige el
filtro (parámetros, rótulos) quedan en el Hueco 1; convertir lo que llega de la
pantalla en este `filtro` es tarea del adaptador HTTP.

## 3. Salidas esperadas

```
type FiltroDefectos struct {
    SprintDeteccionID string
    Severidad         defecto.Severidad
}

func (c *ListarDefectos) Ejecutar(ctx context.Context, proyectoID string, filtro FiltroDefectos) ([]DefectoListado, error)

func ValidarSeveridad(severidad Severidad) error // internal/domain/defecto
```

El resultado es el `[]DefectoListado` de T-021 (sección 3), sin campos nuevos: los
mismos defectos con los mismos datos (el defecto, la historia como `#N título`,
los números de los dos sprints y la fecha de carga), pero solo los que cumplen el
filtro y en el mismo orden (RN-024-6). Sin filtros, o sin coincidencias, el
resultado es la lista completa o una lista vacía, respectivamente; en ningún
caso un error.

`ValidarSeveridad` es la validación de RN-022-2 que hoy está escrita dentro de
`defecto.Nuevo`, extraída para que el filtro y el alta usen **la misma regla**.
Devuelve `nil` si la severidad es una de las cuatro, y `ErrSeveridadInvalida`
envuelto con el valor recibido (`se recibio "urgente"`) en cualquier otro caso,
incluida la vacía. Es la regla del alta: que en el filtro la severidad vacía
signifique "sin filtro" lo decide el caso de uso antes de llamarla (RN-024-8).

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-024-1** (aplicación) — Sin filtros (`SprintDeteccionID` y `Severidad`
  vacíos) el resultado es **exactamente** el de T-021: todos los defectos del
  proyecto, en cualquier estado y en el mismo orden. Agregar los filtros no
  cambia lo que ya hacía la lista.
- **RN-024-2** (aplicación) — El filtro por sprint deja solo los defectos cuyo
  **sprint de detección** es el pedido. El de resolución no cuenta: un defecto
  detectado en el Sprint 1 y resuelto en el Sprint 2 sale al filtrar por el
  Sprint 1 y no al filtrar por el Sprint 2.
- **RN-024-3** (aplicación) — El filtro por severidad deja solo los defectos de
  esa severidad, de cualquier sprint y en cualquier estado.
- **RN-024-4** (aplicación) — Los dos filtros se combinan con **Y**, como los de
  US-020 (RN-020-2): un defecto sale si cumple todos los filtros que vienen
  completos. Con uno solo informado, se aplica ese solo.
- **RN-024-5** (aplicación) — Que ningún defecto cumpla el filtro **no es un
  error**: el resultado es una lista vacía. Vale para un sprint del proyecto sin
  defectos, para una severidad sin defectos, para dos filtros cuya combinación no
  tiene coincidencias aunque cada uno tenga por separado, y para un proyecto sin
  defectos.
- **RN-024-6** (aplicación) — El filtro solo **quita** defectos: no cambia el
  contenido de los que quedan ni su orden. Cada defecto es el mismo
  `DefectoListado` que sin filtrar, y la lista sigue de más reciente a más
  antiguo por fecha de carga, con la misma fecha por ID descendente (RN-T021-4).
  Un defecto `descartado` sale si cumple el filtro: la lista no oculta ningún
  estado (RN-T021-3).
- **RN-024-7** (aplicación) — Si se informó un sprint, tiene que existir y ser
  **de ese proyecto**; si no, `ErrSprintInexistente`, envuelto con el ID. Un
  sprint de otro proyecto, uno que no existe y uno mal formado son lo mismo. Se
  acepta cualquier sprint del proyecto como filtro: abierto, cerrado, que todavía
  no empezó o sin defectos (RN-022-6, RN-T021-15). Si no se informó, no se
  verifica nada.
- **RN-024-8** (dominio) — Si se informó una severidad, tiene que ser una de las
  cuatro y exacta: no se pasa a minúscula ni se recorta. Cualquier otro valor
  (`Alta`, `CRITICA`, `crítica`, `" alta "`, `urgente`) se rechaza con
  `ErrSeveridadInvalida`, envuelto con el valor. La regla es la de RN-022-2 y
  está escrita **una sola vez** (`ValidarSeveridad`). La diferencia con el alta es
  que acá la severidad **vacía** no es inválida: es "sin filtro", y el caso de uso
  no llama al dominio. Una severidad de solo espacios no está vacía: es inválida.
- **RN-024-9** (aplicación) — Los chequeos se hacen en este orden: proyecto,
  sprint y severidad. El primero que falla se devuelve **solo**, sin unirse con
  los demás, y no se lee la lista. Con el sprint inexistente y la severidad
  inválida a la vez se devuelve solo `ErrSprintInexistente`; con el proyecto
  inexistente, solo `ErrProyectoInexistente`, aunque los filtros también sean
  inválidos. Es el criterio de RN-022-5, RN-023-7 y US-020: la existencia se
  verifica antes que los datos, y no se unen a sus errores.
- **RN-024-10** (aplicación) — La consulta no modifica nada: ni los defectos ni
  la historia ni los sprints.
- **RN-024-11** (pantalla) — La pantalla de defectos de T-021 ofrece los dos
  filtros: el de sprint con **todos** los sprints del proyecto, por número
  descendente (RN-T021-15, RN-T021-16), y el de severidad con las cuatro
  (etiquetas «Crítica», «Alta», «Media» y «Baja», como en el alta), cada uno con
  una opción para no filtrar. Muestra la lista filtrada con las mismas columnas,
  etiquetas, orden y botones por fila que T-021 (RN-T021-13 y RN-T021-14): una
  fila se ve igual con o sin filtro. Ofrecer todos los sprints, y no solo los que
  tienen defectos, es lo que permite llegar al caso sin coincidencias (RN-024-5)
  desde la pantalla.
- **RN-024-12** (pantalla) — Los errores del filtro se responden con el
  mecanismo uniforme de T-005, sin entradas nuevas en el catálogo:
  `ErrSprintInexistente` como `no_encontrado` (404, igual que RN-T021-12 y
  RN-T019-7) y `ErrSeveridadInvalida` como `validacion` (422, campo `severidad`).
  Un proyecto inexistente o con ID mal formado sigue siendo el 404 de RN-T021-10.
  Una lista filtrada sin coincidencias responde 200 con la lista vacía y sin
  mensaje de error (RN-024-5); el texto que la acompaña está en el Hueco 1. Una
  falla inesperada de la base es un 500 con el mensaje fijo de T-005.

## 5. Restricciones

- **Dominio puro.** `internal/domain/defecto` suma `ValidarSeveridad` usando solo
  la biblioteca estándar, sin SQL, HTTP, IDs ni `time.Now()`. `Nuevo` pasa a
  llamarla: es un refactor sin cambio de comportamiento, con el mismo error, el
  mismo mensaje y el mismo orden de errores unidos (RN-022-4). No se agrega
  ningún error ni tipo nuevo al dominio.
- **Caso de uso.** `app.ListarDefectos` conserva su nombre, su constructor
  (`NuevoListarDefectos`) y su resultado, como pide T-021 (sección 5); solo suma el
  parámetro `filtro` a `Ejecutar`. Hay que ajustar sus llamadores (el handler de la
  lista y los tests de T-021) en el PR de implementación, y esa línea de la sección
  3 de la spec de T-021 (`ListarDefectos(proyectoID)`), sin cambiar nada más de ella.
- **Puerto.** `app.RepositorioListarDefectos` suma `ExisteSprintEnProyecto(ctx,
  proyectoID, sprintID) (bool, error)`, el mismo método (y con el mismo
  significado) que ya cumplen `postgres.RepositorioDefectos` y `memoria.Base` por
  US-022: un sprint de otro proyecto, inexistente, vacío o mal formado no cuenta.
  `ListarDefectosDelProyecto` no cambia, así que esta historia no toca ningún
  adaptador de persistencia.
- **El filtro se aplica en la aplicación, no en SQL.** El repositorio sigue
  devolviendo todos los defectos del proyecto sin ordenar y el caso de uso filtra y
  ordena, como ya hace con el orden (T-021) y como hace US-020 con sus filtros. Así
  la memoria y Postgres no pueden filtrar distinto.
- **Sin migración.** No se agregan columnas ni índices: la tabla `defectos` ya tiene
  `severidad` y `sprint_deteccion_id` (`00012`).
- **Catálogo de errores (T-005).** Esta historia no suma entradas ni nombres de
  campo. `defecto.ErrSeveridadInvalida` ya está como `validacion` con el campo
  `severidad` (US-022), que ya figura en la lista cerrada de RN-T005-8;
  `app.ErrSprintInexistente` y `app.ErrProyectoInexistente` ya están como
  `no_encontrado`, que por regla no lleva campo. Por eso el sprint del filtro no
  necesita un campo nuevo: `sprint_deteccion`, que T-021 agregó para el formulario
  de alta, queda para ese formulario. El nombre del parámetro de la URL del filtro
  no es un campo del catálogo.
- **Steps.** No crea un área: suma a la interfaz `Defectos` de US-022 un método que
  lista con los filtros y otro que informa qué historias tienen los defectos de la
  última lista, y estas frases (ADR 0002), que se aprueban en el PR de implementación
  y se suman a `docs/diccionario-steps.md`, donde el área `Defectos` pasa a figurar
  también con US-024:
  - `Cuando se listan los defectos`
  - `Cuando se listan los defectos del sprint "Sprint 1"`
  - `Cuando se listan los defectos de severidad "alta"`
  - `Cuando se listan los defectos del sprint "Sprint 1" y de severidad "alta"`
  - `Y la lista de defectos tiene 2 defectos`
  - `Y la lista de defectos incluye el defecto de la historia "Crear proyecto"`
  - `Y la lista de defectos no incluye el defecto de la historia "Crear proyecto"`

  Los defectos del escenario se arman con las frases de US-022 (un defecto por
  historia, así cada uno se nombra por su historia) y los cambios de estado, con las
  de US-023. Como en esas historias, el sprint se busca por nombre dentro del proyecto
  y uno que no esté ahí se manda con un ID que no existe, así el rechazo lo produce la
  aplicación: eso cubre "inexistente" (CL-024-14). La severidad se pasa tal cual,
  para poder escribir `urgente`. "Sprint de otro proyecto" y "proyecto inexistente" no
  se pueden escribir con estas frases: los prueban el caso de uso y el repositorio.
  Los rechazos se afirman con las frases `la operación se rechaza` y `el mensaje de
  error indica "..."` que ya existen.
- **Fuera de alcance.** Filtrar por sprint de resolución, por estado, por historia o
  por texto; elegir varias severidades o varios sprints a la vez; elegir el orden;
  paginar; mostrar totales; recordar el filtro entre visitas; registrar o cambiar el
  estado de un defecto (US-022, US-023, T-021); y el desglose por severidad de
  US-029.

## 6. Casos límite

- **CL-024-1** (RN-024-1) — Sin filtros, o con los dos valores vacíos, la lista es
  igual a la de T-021: todos los defectos del proyecto, abiertos, en corrección,
  resueltos y descartados. Un solo valor vacío no filtra por esa dimensión y el otro
  se aplica solo (RN-024-4).
- **CL-024-2** (RN-024-2) — Defectos en el Sprint 1 y en el Sprint 2, filtro por el
  Sprint 1: salen solo los del Sprint 1, de cualquier severidad y en cualquier
  estado.
- **CL-024-3** (RN-024-2) — Defecto detectado en el Sprint 1 y resuelto en el
  Sprint 2: sale al filtrar por el Sprint 1 y no al filtrar por el Sprint 2.
- **CL-024-4** (RN-024-3) — Filtro por `alta`: salen solo los defectos de severidad
  `alta`, de todos los sprints. Se prueba con cada una de las cuatro severidades.
- **CL-024-5** (RN-024-4) — Filtro por el Sprint 1 y por `alta`: sale solo el
  defecto del Sprint 1 con severidad `alta`. No salen ni el del Sprint 1 con
  severidad `baja` ni el del Sprint 2 con severidad `alta`.
- **CL-024-6** (RN-024-5) — Sprint del proyecto que no tiene defectos (por ejemplo el
  Sprint 3, recién creado): lista vacía, sin error.
- **CL-024-7** (RN-024-5) — Severidad válida sin defectos (por ejemplo `critica` en un
  proyecto que solo tiene `alta` y `baja`): lista vacía, sin error.
- **CL-024-8** (RN-024-5) — El Sprint 1 tiene defectos y hay defectos `critica`, pero
  ninguno del Sprint 1 es `critica`: la combinación da una lista vacía, sin error.
- **CL-024-9** (RN-024-5) — Proyecto sin ningún defecto, con filtros válidos o sin
  filtros: lista vacía, sin error.
- **CL-024-10** (RN-024-6) — Un defecto `descartado` que cumple el filtro sale, con
  estado «Descartado» y «—» de sprint de resolución. Cambiar el estado de un defecto
  (US-023) no lo saca de una lista filtrada, porque el filtro mira su sprint de
  detección y su severidad, que el cambio no toca (RN-023-9).
- **CL-024-11** (RN-024-6) — Dos defectos que cumplen el filtro y tienen la misma fecha
  de carga: salen siempre en el mismo orden, por ID descendente. Con distinta fecha,
  el más reciente primero. El filtro no reordena.
- **CL-024-12** (RN-024-6) — Un defecto filtrado tiene los mismos datos que en la
  lista completa: la misma historia (`#N título`), el mismo número de sprint de
  detección y el mismo de resolución (o `0`, sin resolver).
- **CL-024-13** (RN-024-7) — Sprint que existe pero es de **otro proyecto**:
  `ErrSprintInexistente`, aunque el ID exista en otro lado. No se devuelve una
  lista vacía.
- **CL-024-14** (RN-024-7) — Sprint que no existe, o con un ID mal formado (`abc`):
  `ErrSprintInexistente`. Un sprint **vacío** no es inexistente: es "sin filtro de
  sprint" (CL-024-1).
- **CL-024-15** (RN-024-7) — Sprint cerrado, abierto o que todavía no empezó (fechas
  posteriores a hoy): se acepta como filtro.
- **CL-024-16** (RN-024-8) — Severidad `Alta`, `CRITICA`, `crítica`, `" alta "`, un
  solo espacio o `urgente`: `ErrSeveridadInvalida`, envuelto con el valor recibido. La
  severidad vacía no está en esta lista: es "sin filtro" (CL-024-1).
- **CL-024-17** (RN-024-9) — Sprint inexistente y severidad `urgente` a la vez: se
  devuelve solo `ErrSprintInexistente`, y `errors.Is` no reconoce
  `ErrSeveridadInvalida`. Con el sprint corregido y la misma severidad, se devuelve
  `ErrSeveridadInvalida`.
- **CL-024-18** (RN-024-9) — Proyecto inexistente, con un sprint también inexistente y
  una severidad inválida: se devuelve solo `ErrProyectoInexistente`.
- **CL-024-19** (RN-024-9, RN-024-10) — Cualquier rechazo ocurre antes de leer la
  lista y no modifica nada.
- **CL-024-20** (RN-024-12) — Falla inesperada de la base al verificar el sprint o al
  listar, con o sin filtro: 500 con el mensaje fijo de T-005, el detalle solo al
  registro (CL-T021-18), y ninguna lista parcial.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario | Origen |
|---|---|---|---|
| El proyecto no existe | `app.ErrProyectoInexistente`, envuelto con el ID | "el proyecto no existe" | Aplicación |
| El sprint del filtro no existe, no es de ese proyecto o está mal formado | `app.ErrSprintInexistente`, envuelto con el ID | "el sprint no existe" | Aplicación |
| La severidad del filtro no es `critica`, `alta`, `media` ni `baja` | `defecto.ErrSeveridadInvalida`, envuelto con el valor recibido | "la severidad tiene que ser critica, alta, media o baja" | Dominio |

Los tres se detectan **antes** de leer la lista, en el orden de la tabla (RN-024-9),
y se devuelven de a uno: nunca unidos. Ninguno es nuevo: los tres ya existen y
están en el catálogo de T-005, y son valores de paquete comparables con
`errors.Is`. Que el filtro no tenga coincidencias no es un error (RN-024-5).

En la pantalla, `ErrProyectoInexistente` y `ErrSprintInexistente` responden 404 y
`ErrSeveridadInvalida`, 422 con el campo `severidad` (RN-024-12). Una falla
inesperada de la base es un 500 con el mensaje fijo de T-005.

## 8. Criterios de aceptación

- **CA-024-1** (normal) — Con defectos de distintas severidades en el Sprint 1 y en
  el Sprint 2, filtrar por el Sprint 1 devuelve solo los defectos detectados en él, y
  filtrar por una severidad devuelve solo los defectos de esa severidad, de
  cualquier sprint.
- **CA-024-2** (alternativo) — Los dos filtros se combinan: filtrar por el Sprint 1 y
  por `alta` devuelve solo los defectos que cumplen ambos.
- **CA-024-3** (límite) — Un filtro sin coincidencias (un sprint del proyecto sin
  defectos, una severidad sin defectos o una combinación sin intersección) devuelve
  una lista vacía, no un error.
- **CA-024-4** (error) — Un sprint de otro proyecto o inexistente se rechaza con
  `ErrSprintInexistente`, y una severidad inválida (por ejemplo `urgente` o `Alta`),
  con `ErrSeveridadInvalida`; en ambos casos con el error uniforme de T-005.
- **CA-024-5** (alternativo) — Sin filtros se devuelve la lista completa de defectos
  del proyecto, igual que antes de esta historia.
- **CA-024-6** (error) — Con un sprint inexistente y una severidad inválida a la vez
  se rechaza solo por el sprint (`ErrSprintInexistente`), sin unir los dos errores.

---

## Decisiones tomadas y descartadas

- **Se extiende `ListarDefectos`; no se crea otro caso de uso.** T-021 lo dejó
  previsto: US-024 le suma los filtros "sin cambiar su nombre ni su resultado".
  Se descartó un `ListarDefectosFiltrados` aparte: serían dos lecturas de la misma
  lista que pueden divergir, y la pantalla tendría que elegir entre ellas.
- **El filtro es un struct con el valor vacío como "sin filtro".** Es el criterio de
  `FiltroEsfuerzo` (US-020) y de T-019 ("los filtros vacíos no filtran"). Se
  descartaron los punteros o un tipo opcional: para dos cadenas, el cero de Go ya dice
  "no informado" y mantiene `Ejecutar` fácil de llamar y de probar.
- **El sprint se identifica por su ID.** CA-024-4 habla de "un sprint de otro
  proyecto", y eso solo se puede verificar con un ID (el nombre `Sprint N` y el número
  se repiten en todos los proyectos). Es lo que hacen US-020, US-022 y US-023. Se
  descartó filtrar por nombre o por número (ver Hueco 2).
- **Se filtra por el sprint de detección, no por el de resolución.** Es lo que dice el
  criterio del issue. El sprint de resolución se sigue viendo en su columna; filtrar
  por él es otra pregunta ("qué se resolvió en el sprint") y queda fuera de alcance.
- **Los filtros se aplican en la aplicación, no en SQL.** Mismo motivo que el orden en
  T-021: un único código para la memoria y Postgres, así no pueden filtrar distinto.
  Se descartó un `WHERE` en la consulta: sumaría parámetros al puerto y dos
  implementaciones del mismo filtro. Costo aceptado: cada pedido lee todos los defectos
  del proyecto, como ya hace la lista de T-021 y la consulta de US-020; con los
  volúmenes de la cursada no se nota.
- **Una sola regla de severidad: se extrae `ValidarSeveridad`.** Repetir el `switch` de
  `Nuevo` en el caso de uso dejaría la regla de RN-022-2 escrita dos veces. Se descartó
  tratar la severidad vacía como inválida, como en el alta: no habría forma de pedir
  "todas las severidades". Esa diferencia vive en el caso de uso, no en el dominio.
- **Los errores de existencia van primero y solos.** Es la convención de RN-022-5 y
  RN-023-7 (el proyecto, después el sprint, y recién después los datos), y la de US-020
  para sus filtros. Se descartó unir con `errors.Join` el sprint inexistente y la
  severidad inválida: la unión del repo es para errores de **validación del dominio** de
  una misma entrada (RN-022-4, RN-023-6), y ninguna historia mezcla un error de
  existencia con uno de validación. Costo aceptado: con los dos mal, la persona ve
  primero el del sprint y después el de la severidad (ver Hueco 4).
- **Un solo valor por filtro, combinados con Y.** Es el criterio de US-020 y lo que
  dicen los criterios ("un sprint", "una severidad"). Se descartó elegir varias
  severidades o varios sprints a la vez: ningún criterio lo pide y complica el
  contrato.
- **Sin coincidencias es una lista vacía, y no hay una regla por filtro.** Es el mismo
  criterio que RN-020-10 y CL-026-1: no encontrar nada es el resultado normal de una
  búsqueda, no un dato corrupto. Se descartó un error o un 404 para "ese sprint no
  tiene defectos": se confundiría con el sprint que no existe.
- **La pantalla ofrece todos los sprints, no solo los que tienen defectos.** Es lo que
  hace RN-T021-15 para el alta y la resolución, y es lo que permite llegar a CA-024-3
  desde el navegador. Se descartó listar solo los sprints con defectos: el vacío
  quedaría imposible de producir y la lista cambiaría con cada alta.
- **Sin cambios de catálogo, de migración ni de adaptadores.** Los tres errores y los
  nombres de campo ya existen, y `ExisteSprintEnProyecto` ya está en los dos
  repositorios. Todo lo nuevo es una firma, una función extraída y los controles de la
  pantalla.

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia:

1. **Cómo viaja el filtro y cómo se ve en la pantalla.** Ningún criterio ni spec
   existente lo fija para defectos. Se propone lo de T-019 (la única pantalla con
   filtros): un formulario `GET` común sobre `/proyectos/{id}/defectos`, con los
   parámetros `sprint` y `severidad` (valores vacíos no filtran), de modo que la URL
   quede compartible y funcione sin JavaScript. Quedan por fijar los rótulos de la opción
   para no filtrar («Todos» / «Todas»), el texto de la lista vacía por filtro (T-019 usa
   «Ningún registro cumple los filtros», distinto del «Todavía no hay defectos» de
   T-021), si la pantalla conserva lo elegido al mostrar el resultado y si hay un
   enlace para limpiar. Un parámetro repetido en la URL (`?severidad=alta&severidad=baja`)
   se resolvería leyendo solo el primero.
2. **El sprint se filtra por ID.** Se propone por CA-024-4 y por US-020, US-022 y
   US-023. Si el equipo prefiere filtrar por número de sprint (`Sprint 2`), cambian
   RN-024-7 y la entrada, y la verificación de pertenencia se hace igual dentro del
   proyecto.
3. **Valor vacío es "sin filtro", y la severidad no se normaliza.** Un parámetro vacío
   no filtra (como en T-019), pero una severidad de solo espacios o con otra
   capitalización es inválida (`ErrSeveridadInvalida`), igual que en el alta. Confirmar
   que no se quiere aceptar `Alta` o `" alta "` por ser un filtro.
4. **Qué se informa con dos filtros inválidos.** Se aplica la convención de existencia
   primero y solos (RN-024-9): se rechaza el sprint y la severidad se ve recién al
   corregirlo. Si se prefiere mostrar los dos a la vez, cambian RN-024-9, CL-024-17,
   CA-024-6 y la forma en que T-005 clasifica un error unido de dos tipos.
5. **El filtro y las acciones de la fila.** Registrar un defecto o cambiar su estado
   redirige a `/proyectos/{id}/defectos` (T-021), así que con un filtro activo la
   persona vuelve a la lista **sin** filtrar. Confirmar que alcanza, o ajustar el
   destino de esas redirecciones (toca T-021, no esta historia).
6. **Alcance del filtro.** Se filtra solo por sprint de detección y por severidad, de a
   un valor. Filtrar por estado o por sprint de resolución (para ver "lo que se
   resolvió en el Sprint 2") sería otra historia. Confirmar que no hace falta ahora.
7. **Criterios del issue #32 y la frase "Como / quiero / para".** El issue no trae esa
   frase (está redactada para esta spec). Los criterios `CA-024-1` a `CA-024-4` se
   afinan sin cambiar su sentido; `CA-024-5` (sin filtros) y `CA-024-6` (dos filtros
   inválidos a la vez) son nuevos. Los de la sección 8 son los propuestos para el
   issue, que hay que corregir a mano.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir del issue
  #32, las specs de US-022, US-023 y T-021 y las de US-020 y T-019 como modelo de
  filtros, y el código de `ListarDefectos` y del catálogo de errores; las decisiones
  sobre el orden de los errores, el valor vacío, el identificador del sprint y el
  alcance del filtro se proponen acá y se confirman en la revisión
- Revisado y entendido por: Angelo Conforti
