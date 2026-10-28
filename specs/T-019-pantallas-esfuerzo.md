# T-019 — Pantallas de esfuerzo: cargar horas, consultar y comparar

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás |
| **Issue** | #189 (tarea) · #193 (spec) |
| **Escenarios BDD** | ninguno: es una tarea técnica (ver "Decisiones") |
| **Código** | `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| **Última actualización** | 2026-10-28 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Que una persona cargue desde el navegador las horas que dedicó a una historia, vea lo que se
registró (filtrando por historia, integrante y sprint) y compare lo estimado contra lo que se
trabajó de verdad. Cubre la parte de interfaz del requerimiento 5 de la guía (esfuerzo).

Los tres casos de uso ya están en `main` y con sus tests: **US-019** registra las horas,
**US-020** las consulta y **US-021** las compara con lo estimado. Esta tarea no agrega reglas de
dominio, casos de uso ni lecturas nuevas: son rutas, formularios y la forma de mostrar los
resultados.

Es la **primera tarea del Sprint 5** y tiene fecha: tiene que estar en `main` el miércoles
28/10 a las 22:00. Desde ese día cada integrante carga sus horas todos los días, porque las
horas reales llevan cuatro retros sin registrarse por falta de un lugar donde cargarlas
(acción de la retro del Sprint 4). El dashboard (US-030) y el reporte (US-033) muestran estos
datos.

Se apoya en T-004 (layout, router y `Responder`), T-005 (errores uniformes y lector de
formularios) y en el patrón de T-011 y T-015.

## 2. Entradas

| Ruta | Método | Campos | Caso de uso |
|---|---|---|---|
| `/proyectos/{id}/esfuerzo` | GET | query opcional: `historia`, `integrante`, `sprint` | `ConsultarEsfuerzo` (US-020) |
| `/proyectos/{id}/esfuerzo` | POST | `integrante`, `historia`, `fecha`, `actividad`, `horas` | `RegistrarEsfuerzo` (US-019) |
| `/proyectos/{id}/esfuerzo/comparacion` | GET | — | `CompararEsfuerzo` (US-021) |

- `{id}` es el ID del proyecto. `integrante`, `historia` y `sprint` son los IDs que genera la base:
  la persona los elige de listas que muestran los nombres.
- `fecha` llega como `AAAA-MM-DD` (`<input type="date">`), con la fecha de hoy como valor inicial.
- `horas` se escribe con coma o con punto (`2,5` o `2.5`), como en US-019.
- `actividad` es un texto de una línea de hasta 120 caracteres.
- La carga se envía con HTMX (`hx-post`), igual que los formularios de T-011 y T-015. Los filtros son
  un formulario `GET` común: la URL queda compartible.
- Los filtros vacíos no filtran. Las listas de la pantalla salen de lecturas que ya existen:
  los integrantes de `ConsultarProyecto`, las historias de `ListarBacklog` y los sprints de
  `ListarSprints` (T-011 y T-015).

## 3. Salidas esperadas

**Después de registrar con éxito:** `HX-Redirect` con HTMX o `303` sin él, hacia
`/proyectos/{id}/esfuerzo`. Los errores se responden con `ResponderError` (T-005): el mensaje cae en
`#mensajes` con el campo en `data-campo`, como en T-011 y T-015 (el formulario no se redibuja), y no se registra
nada.

**`/proyectos/{id}/esfuerzo`** tiene, en este orden:

| Parte | Qué muestra |
|---|---|
| Formulario «Registrar horas» | Integrante (lista con los del proyecto), historia (lista con las del backlog), fecha, actividad y horas |
| Filtros | Historia, integrante y sprint, cada uno con la opción «Todas» o «Todos», y el botón «Filtrar» |
| Total | «Total: 7,5 h» de los registros que quedaron después de filtrar |
| Tabla de registros | Fecha (`dd/mm/aaaa`), historia, integrante, actividad, horas y sprint (o «Fuera de sprint»), en el orden de US-020 |
| Totales | Por historia, por integrante y por sprint (con «Fuera de sprint» al final), solo de lo filtrado |
| Enlace | «Comparar con lo estimado», hacia la comparación |

**`/proyectos/{id}/esfuerzo/comparacion`:** una tabla con una fila por historia que aparece en la
comparación de US-021 (la que más se desvió primero) con sus Story Points, las horas estimadas, las
reales, la desviación en horas y en porcentaje, y su situación: «Completada», «En curso» o «Sin
estimar». Una historia sin estimar no tiene base de comparación: sus Story Points, sus horas estimadas y
su desviación (en horas y en porcentaje) se muestran como «—» y no como un 0 que parezca un dato; sus
horas reales sí se ven. Debajo, el total de las historias comparables («N historias comparables:
estimadas X h, reales Y h, desviación Z h (P %)», y «1 historia comparable» en singular) y «Horas sin
comparar: W h». Un enlace («Volver a la carga de horas») vuelve a la carga.

**Menú del proyecto:** aparece el item «Esfuerzo», activo en estas páginas.

**Formato de los números:** las horas con coma decimal y sin ceros de más (`2,5`, `8`, `12`);
la desviación con signo (`+0,5`, `-3`); el porcentaje con un decimal y el símbolo (`6,7 %`); el
porcentaje que no se puede calcular (historia sin estimar) se muestra como «—».

## 4. Reglas de negocio

No hay reglas de dominio nuevas. Son reglas de comportamiento de la interfaz.

**Capas**

- **RN-T019-1** — Un handler solo lee la petición, llama a un caso de uso y escribe la respuesta.
  No valida reglas de negocio ni habla con un repositorio (RN-T011-1).
- **RN-T019-2** — No hay casos de uso ni puertos nuevos: se usan `RegistrarEsfuerzo`,
  `ConsultarEsfuerzo` y `CompararEsfuerzo`, más `ConsultarProyecto`, `ListarBacklog` y `ListarSprints`
  para llenar las listas. Se agregan a `CasosDeUso` y se conectan en `cmd/server`.

**Formularios**

- **RN-T019-3** — `fecha` y `horas` se leen con `Formulario` (T-005): una fecha que no es
  `AAAA-MM-DD` o unas horas que no son un número se informan en su campo. Lo demás
  (fecha futura, horas fuera de rango o que no son múltiplo de 0,25, actividad vacía o muy larga) lo
  dice el dominio de US-019 y se informa en el campo del error (RN-T005-10).
- **RN-T019-4** — `integrante` e `historia` son listas obligatorias: si llegan vacías, se informa en
  su campo «Elegí un integrante» o «Elegí una historia» y no se llama al caso de uso. Los nombres de
  campo `integrante` e `historia` se suman a la lista cerrada de RN-T005-8 y a la de
  `TestCatalogo_EntradasValidas`, sin tocar ninguna otra entrada, como hicieron US-022 y US-023.
- **RN-T019-5** — Un integrante o una historia de otro proyecto, o inexistente, se rechaza con el
  error de inexistencia del catálogo (404 en un POST: fragmento de error), aunque exista en otro
  lado (RN-019-6).
- **RN-T019-6** — Un texto con HTML (por ejemplo `<script>` en la actividad) se muestra escapado en
  cualquier pantalla (RN-T011-8).

**Consulta**

- **RN-T019-7** — Los filtros se combinan con «y» (US-020, RN-020-2). Una historia, un integrante o
  un sprint de otro proyecto, o que no existe, en la URL responde 404: página con layout.
- **RN-T019-8** — Los totales, el orden y la pertenencia a un sprint (por fecha) son los de
  `ConsultaEsfuerzo`: la pantalla no recalcula nada.
- **RN-T019-9** — La comparación muestra exactamente lo que devuelve `ComparacionProyecto` (US-021): el
  mismo orden, las mismas historias y los mismos totales. Lo que no es comparable (en curso o sin
  estimar) se identifica en la fila y su esfuerzo entra solo en «Horas sin comparar».
- **RN-T019-10** — Las listas de la pantalla se ordenan así: integrantes por nombre sin distinguir
  mayúsculas (RN-T011-5), historias en el orden del backlog (RN-T011-4) y sprints por número
  descendente (RN-T015-3).
- **RN-T019-11** — Sin registros, o con filtros sin coincidencias, la tabla y los totales se
  reemplazan por un texto de vacío («Todavía no hay horas registradas», «Ningún registro cumple los
  filtros») y no hay error. Lo mismo vale para la comparación sin historias.

**Alcance de las rutas y navegación**

- **RN-T019-12** — Un proyecto que no existe o con ID mal formado responde 404: página con layout en un
  GET, fragmento de error en un POST (RN-T011-10).
- **RN-T019-13** — Cada campo tiene su `<label>`, y cada control de la tabla tiene un nombre accesible
  que dice a qué registro corresponde (RN-T011-9).
- **RN-T019-14** — El menú del proyecto suma «Esfuerzo» (sección nueva), activa en las dos páginas. Las
  fechas se muestran como `dd/mm/aaaa` (RN-T015-9).

## 5. Restricciones

- **Sin cambios de dominio ni de aplicación.** Los tres casos de uso están en `main` y los repositorios
  de memoria y de Postgres ya cumplen sus puertos (`RepositorioEsfuerzo`, `RepositorioConsultarEsfuerzo` y
  `RepositorioCompararEsfuerzo`): `cmd/server` solo los conecta.
- **Reloj.** El valor inicial de `fecha` (hoy) sale del mismo reloj que usa `RegistrarEsfuerzo`: `time.Now`
  en `cmd/server` y uno fijo en los tests. Llega a las pantallas como el campo `Ahora` de `CasosDeUso` (si falta,
  `time.Now`). Es la hora del servidor, que en Azure es UTC: pasadas las 21:00 de Argentina el valor inicial ya
  es la fecha de mañana, y el dominio lo acepta porque compara contra el mismo reloj (una fecha futura recién se
  rechaza si es posterior a «hoy» en UTC). Si molesta, se cambia el día a mano en el formulario.
- **Catálogo de errores.** Los errores que aparecen ya están registrados, salvo los dos campos nuevos de
  RN-T019-4.
- **Los formularios requieren JavaScript (HTMX),** como en T-011 y T-015. El de filtros no.
- **El diseño visual es funcional** (Tailwind); se mejora después.
- **Fuera de alcance:** editar o borrar un registro (ninguna historia lo pide), un total por día o por
  semana, gráficos, exportar, registrar en nombre de otra persona con un control de acceso (la
  aplicación no tiene usuarios) y las horas por sprint de US-027 en esta pantalla, que van al dashboard
  (US-030).

## 6. Casos límite

- **CL-T019-1** — Proyecto sin historias o sin integrantes: el formulario muestra las listas vacías y
  un aviso («Primero cargá una historia» / «Primero registrá un integrante»); registrar se rechaza con
  el error del campo.
- **CL-T019-2** — Horas `2,5` y `2.5`: las dos se aceptan y quedan como 2,5. `2,50` se acepta. `1,3`
  se rechaza con el error de US-019 y `abc` con el de lectura, los dos en el campo `horas`.
- **CL-T019-3** — Fecha de hoy: se acepta; de mañana, se rechaza con el error de US-019 en el campo
  `fecha`.
- **CL-T019-4** — Actividad de 120 caracteres: se acepta; de 121, se rechaza. Con espacios al principio y
  al final se recorta.
- **CL-T019-5** — Dos registros iguales (misma historia, integrante y día): se aceptan los dos y los dos
  aparecen (CL-019-11).
- **CL-T019-6** — Registro con una fecha que cae entre dos sprints: se acepta y aparece como «Fuera de
  sprint».
- **CL-T019-7** — Filtrar por un sprint: salen los registros cuya fecha cae en él, con el primer y el
  último día incluidos (RN-020-3), aunque la historia sea de otro sprint.
- **CL-T019-8** — Filtros que no coinciden con nada (una historia sin horas del integrante elegido): texto de
  vacío, total en 0 y sin error.
- **CL-T019-9** — Un registro en una historia completada o que no está en un sprint: se acepta
  (RN-019-7).
- **CL-T019-10** — Comparación con una historia completada y estimada sin horas: aparece con las
  estimadas, 0 reales y la desviación negativa de lo estimado (CL-021-3).
- **CL-T019-11** — Comparación con una historia sin estimar y con horas: situación «Sin estimar», porcentaje
  «—» y su esfuerzo en «Horas sin comparar».
- **CL-T019-12** — Una falla inesperada de la base al registrar: 500 con el mensaje fijo de T-005, sin
  registrar nada.

## 7. Condiciones de error

| Situación | Respuesta |
|---|---|
| Proyecto inexistente o con ID mal formado | 404: página con layout (GET) o fragmento de error (POST) |
| `integrante` o `historia` sin elegir | 422: error en el campo, sin llamar al caso de uso |
| `fecha` o `horas` con formato inválido | 422: error en el campo |
| Fecha futura, fecha vacía, horas fuera de rango o con fracción, actividad vacía o muy larga | 422: el mensaje de US-019 en el campo correspondiente, todos juntos |
| Integrante, historia o sprint de otro proyecto o inexistente | 404 |
| Falla inesperada (base caída, panic) | 500 con el mensaje fijo de T-005; el detalle solo va al registro |

## 8. Criterios de aceptación

- **CA-T019-1** (normal) — Con el router real y la base en memoria: un proyecto con un integrante y una
  historia; se registran 2,5 horas con una actividad y una fecha válidas. La respuesta es `HX-Redirect`
  con HTMX y `303` sin él; la página de esfuerzo muestra el registro con su historia, su integrante, su
  actividad y sus horas, y el total pasa a 2,5 h.
- **CA-T019-2** (alternativo) — Con registros de dos integrantes en dos historias y dos sprints, filtrar por
  historia, por integrante y por sprint deja solo los registros que corresponden; el total y los totales por
  historia, por integrante y por sprint son los de lo filtrado, y las horas fuera de todo sprint salen
  aparte.
- **CA-T019-3** (alternativo) — La comparación muestra cada historia con sus horas estimadas, reales,
  desviación y porcentaje, la que más se desvió primero, el total de las comparables y las horas sin
  comparar; una historia en curso o sin estimar se identifica y no suma al total.
- **CA-T019-4** (límite) — Sin registros, o con filtros sin coincidencias, las pantallas muestran el
  texto de vacío y no un error; las horas se aceptan con coma y con punto.
- **CA-T019-5** (error) — Una fecha futura, horas en 0, con fracción o fuera de rango, una actividad vacía o
  de más de 120 caracteres, y un integrante o una historia sin elegir se informan nombrando su campo (con
  `data-campo`, en `#mensajes`) y no registran nada; un integrante o una historia de otro proyecto responde 404.
- **CA-T019-6** (plataforma) — Las rutas usan el layout, el router y `Responder`, funcionan con y sin
  HTMX, el menú muestra «Esfuerzo», y con Postgres (y con `make dev` en el navegador) se recorre el flujo
  completo: registrar, filtrar y comparar.

---

## Decisiones tomadas y descartadas

- **Es una tarea y no una historia.** Los casos de uso y sus reglas ya están (US-019, US-020 y US-021);
  esto es la interfaz, como T-011, T-015 y T-016. Se estima en 5 SP.
- **Una sola página para cargar y consultar.** La carga y la tabla conviven: quien registra ve enseguida
  lo que quedó. Se descartó una página aparte para cargar: suma una ruta y obliga a ir y volver.
- **La comparación es una página aparte** y no una pestaña de la tabla: tiene otra pregunta («dónde nos
  equivocamos al estimar») y otra forma (una fila por historia, no por registro).
- **Los filtros son un `GET`.** Se descartó un filtro con HTMX: la URL queda compartible y la página
  funciona sin JavaScript para consultar.
- **No hay control de acceso por integrante.** La aplicación no tiene usuarios (el mismo hueco de US-018 y
  US-022): cualquier integrante del proyecto se puede elegir. Se descartó guardar «quién cargó» aparte.
- **El valor inicial de la fecha es hoy.** Es lo habitual (se carga al terminar el día) y evita el error más
  común, que es la fecha futura.
- **No hay `.feature`.** Es una tarea técnica y las reglas ya tienen los suyos (US-019, US-020 y US-021).
  Los recorridos de CA-T019-1 a CA-T019-5 son tests de Go con `httptest`, como en T-011 y T-015.
- **Qué pasa si el dashboard ya está:** las horas por sprint de US-027 se muestran ahí (US-030), no en
  esta pantalla.

## Uso de IA en esta especificación

- [x] Use IA para: redacción de la especificación a partir de US-019, US-020, US-021, T-011 y T-015, con
  Claude Code — revisado por: Tomás.
