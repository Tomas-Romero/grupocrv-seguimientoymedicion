# US-020 — Consultar el esfuerzo registrado por historia, integrante y sprint

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #23 |
| **Escenarios BDD** | `features/US-020-consultar-esfuerzo.feature` |
| **Código** | `internal/domain/esfuerzo/`, `internal/app/consultar_esfuerzo.go` |
| **Última actualización** | 2026-10-25 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero consultar el esfuerzo registrado filtrando
> por historia, por integrante y por sprint, para saber cuántas horas se
> dedicaron a cada cosa y quién las dedicó.

Es la lectura de los registros que carga US-019. Devuelve los registros de un
proyecto (todos, o los que cumplen un filtro) con los totales de horas por
historia, por integrante y por sprint. Es también la fuente de las horas reales
de un sprint para US-027 (`CalcularHoras`) y de la comparación contra lo
estimado de US-021.

Depende de **US-019** (los registros) y de **US-008** (los sprints y sus
fechas). El sprint de un registro no se guarda: se deduce de su fecha
(RN-019-12), y esta historia es la que hace esa consulta.

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | `string` | Sí | ID de un proyecto existente (viene de la URL, no de un formulario) |
| `filtro.ItemID` | `string` | No | ID de una historia **del proyecto**. Vacío: no se filtra por historia |
| `filtro.IntegranteID` | `string` | No | ID de un integrante **del proyecto**. Vacío: no se filtra por integrante |
| `filtro.SprintID` | `string` | No | ID de un sprint **del proyecto**. Vacío: no se filtra por sprint |

**Orden de verificación.** Primero el proyecto, después los filtros en este
orden: historia, integrante, sprint. El primero que falla se devuelve y no se
une con los demás (es el equivalente a un 404, como en US-019).

La lectura sale de un solo acceso a la persistencia, así todos los datos
(registros, títulos, nombres, sprints) son de la misma foto:

```
LecturaEsfuerzo {
    Registros   []esfuerzo.RegistroEsfuerzo // todos los del proyecto, sin ordenar
    Historias   map[string]string           // ID de historia → título (todas las del proyecto)
    Integrantes map[string]string           // ID de integrante → nombre (todos los del proyecto)
    Sprints     []sprint.Sprint             // todos los del proyecto
}
```

## 3. Salidas esperadas

```
RegistroConsultado {
    Registro     esfuerzo.RegistroEsfuerzo
    Historia     string // título de la historia
    Integrante   string // nombre del integrante
    SprintNumero int    // número del sprint donde cae la fecha; 0 si no cae en ninguno
}

TotalPorHistoria   { ItemID, Titulo string;  Horas float64 }
TotalPorIntegrante { IntegranteID, Nombre string; Horas float64 }
TotalPorSprint     { SprintID string; Numero int; Horas float64 } // SprintID "" y Numero 0: fuera de todo sprint

ConsultaEsfuerzo {
    Registros     []RegistroConsultado
    TotalHoras    float64
    PorHistoria   []TotalPorHistoria
    PorIntegrante []TotalPorIntegrante
    PorSprint     []TotalPorSprint
}

func (c ConsultaEsfuerzo) Horas() []float64
```

Piezas puras del dominio (`internal/domain/esfuerzo`):

```
RangoSprint { ID string; Numero int; Inicio, Fin time.Time }

func SprintDeLaFecha(fecha time.Time, rangos []RangoSprint) (RangoSprint, bool)
func TotalHoras(registros []RegistroEsfuerzo) float64
```

Caso de uso (`internal/app`): `NuevoConsultarEsfuerzo(repo).Ejecutar(ctx, proyectoID, filtro) (ConsultaEsfuerzo, error)`.

- Las horas se redondean a **2 decimales** (sin ruido de coma flotante).
- `Horas()` devuelve las horas de `Registros`, en el mismo orden: es lo que
  recibe `metricas.CalcularHoras` (US-027).

## 4. Reglas de negocio

- **RN-020-1** — Sin filtros se devuelven todos los registros del proyecto.
- **RN-020-2** — Los filtros se combinan con **Y**: un registro se devuelve si
  cumple todos los filtros que vienen completos.
- **RN-020-3** (dominio) — El sprint de un registro es el sprint del proyecto cuyo
  rango contiene su fecha, con el día de inicio y el de fin **incluidos**
  (RN-019-12). Se comparan solo el año, el mes y el día (UTC). Si la fecha no cae
  en ningún sprint, `SprintDeLaFecha` devuelve `false` y el registro queda con
  `SprintNumero = 0`.
- **RN-020-4** — El filtro por sprint trae los registros **cuya fecha cae en el
  sprint**, no los de las historias que se le asignaron. Una historia que pasó
  de un sprint a otro (US-011) aparece en cada uno con las horas que se
  trabajaron en él.
- **RN-020-5** — Los registros salen ordenados por fecha **descendente** (lo más
  reciente primero). Con la misma fecha, por título de la historia, después por
  nombre del integrante y después por actividad, todos ascendentes y sin
  distinguir mayúsculas; si aun así empatan, por ID. El orden es total, así que
  el mismo dato siempre sale igual.
- **RN-020-6** (dominio) — `TotalHoras` es la suma de las horas de los
  registros recibidos, redondeada a 2 decimales. Con una lista vacía es 0.
- **RN-020-7** — `TotalHoras` de la consulta es el total de los registros que
  quedaron **después de filtrar**.
- **RN-020-8** — Los totales por historia, por integrante y por sprint se
  calculan sobre los registros filtrados e incluyen **solo los grupos que tienen
  registros** (no hay filas en cero). Cada registro pertenece a exactamente un
  grupo de cada tipo, así que cada lista suma `TotalHoras`.
- **RN-020-9** — Orden de los totales: por historia y por integrante, de más
  horas a menos (empate: título o nombre ascendente, sin distinguir
  mayúsculas, y después ID). Por sprint, por número ascendente, y el grupo «fuera
  de todo sprint» al final.
- **RN-020-10** — Sin registros (en el proyecto o después de filtrar) **no es un
  error**: `Registros` y las tres listas vienen vacías y `TotalHoras` es 0.
- **RN-020-11** — Los registros de un sprint cerrado se cuentan igual: el esfuerzo
  histórico no se pierde al cerrar.
- **RN-020-12** — La consulta no modifica nada.
- **RN-020-13** (dominio) — Si dos rangos contuvieran la misma fecha (US-008 lo
  impide, pero la función no lo da por hecho), gana el de menor número. Así el
  resultado es determinista y cada registro cuenta en un solo sprint.

## 5. Restricciones

- **Dominio puro.** `internal/domain/esfuerzo` suma `SprintDeLaFecha` y
  `TotalHoras` usando solo la biblioteca estándar. No importa `sprint`: recibe
  los rangos ya armados (`RangoSprint`), igual que recibe los IDs ya verificados.
- **Filtrar y agrupar en la aplicación, no en SQL.** El repositorio lee todo lo
  del proyecto y el caso de uso filtra, ordena y agrupa. El sprint se deduce de
  la fecha, una regla de negocio que no conviene duplicar en una consulta SQL, y
  los volúmenes de un proyecto de la cursada son de cientos de registros.
- **Sin migración.** Usa la tabla `registros_esfuerzo` de US-019. La consulta
  filtra por `proyecto_id`, columna que no tiene índice propio; con estos
  volúmenes no se nota, y si hiciera falta se agrega con la migración de la
  historia que lo mida.
- **Puerto propio** `RepositorioConsultarEsfuerzo` con `ExisteProyecto` y
  `LeerEsfuerzo(ctx, proyectoID) (LecturaEsfuerzo, error)`. Lo implementan la
  base en memoria y Postgres; los dos tienen que dar el mismo resultado, y la
  lectura de Postgres se hace en una sola transacción de solo lectura.
- **Errores existentes.** Se reutilizan `ErrProyectoInexistente`,
  `ErrItemInexistente`, `ErrIntegranteInexistente` y `ErrSprintInexistente`, que
  ya están en el catálogo de errores (T-005). No se agrega ningún error nuevo.
- **Escenarios BDD.** Van en el área Esfuerzo de los steps (se amplía la interfaz
  `Esfuerzo`). Los contextos reutilizan las frases de proyecto, integrante,
  historia y sprint del diccionario; se agregan frases de consulta (ver
  Decisiones).
- **Fuera de alcance.** La pantalla de consulta (se especifica aparte, como T-015
  para los sprints), editar o borrar registros, exportar, el rango de fechas
  libre como filtro y la paginación.

## 6. Casos límite

- **CL-020-1** — Proyecto sin ningún registro: resultado vacío, total 0, sin
  error.
- **CL-020-2** — Dos registros del mismo integrante, la misma historia y el mismo
  día (CL-019-11): aparecen los dos y suman los dos.
- **CL-020-3** — Un registro con fecha exactamente el día de inicio de un sprint,
  y otro el día de fin: los dos pertenecen a ese sprint.
- **CL-020-4** — Un registro con fecha entre dos sprints, o antes del primero
  (CL-019-17): aparece con `SprintNumero = 0` y suma al grupo «fuera de todo
  sprint». Con el filtro de un sprint no aparece.
- **CL-020-5** — Una historia con horas en la fecha del Sprint 1 que no se completó
  y pasó al Sprint 2 (CL-019-16): sus horas del Sprint 1 salen en el Sprint 1 y
  las del Sprint 2, en el Sprint 2.
- **CL-020-6** — Filtro por historia e integrante a la vez: solo los registros de
  ese integrante en esa historia.
- **CL-020-7** — Filtros que no tienen coincidencias (un integrante que nunca
  registró nada en esa historia): resultado vacío, no es un error.
- **CL-020-8** — Una historia de otro proyecto como filtro: `ErrItemInexistente`,
  aunque la historia exista en otro lado. Lo mismo con un integrante
  (`ErrIntegranteInexistente`) y con un sprint (`ErrSprintInexistente`).
- **CL-020-9** — Un ID mal formado como filtro: se trata como inexistente.
- **CL-020-10** — Proyecto inexistente con un filtro que también es inexistente:
  `ErrProyectoInexistente` (el proyecto tiene prioridad).
- **CL-020-11** — Historia y sprint inexistentes a la vez: `ErrItemInexistente`
  (la historia se verifica antes, por el orden de la sección 2).
- **CL-020-12** — Horas que suman con decimales (`0,25 + 0,25 + 0,5`): el total es
  exactamente `1`.
- **CL-020-13** — Dos historias con las mismas horas: salen ordenadas por título.
- **CL-020-14** — Dos rangos de sprint que contienen la misma fecha (dato que
  US-008 impide): gana el de menor número (RN-020-13).
- **CL-020-15** — Sprint cerrado: sus registros se consultan igual (RN-020-11).

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| El proyecto no existe | `app.ErrProyectoInexistente` | "el proyecto no existe" |
| `filtro.ItemID` no es una historia del proyecto | `app.ErrItemInexistente` | "el item no existe" |
| `filtro.IntegranteID` no es un integrante del proyecto | `app.ErrIntegranteInexistente` | "el integrante no existe" |
| `filtro.SprintID` no es un sprint del proyecto | `app.ErrSprintInexistente` | "el sprint no existe" |

Se devuelve el primero que falla, en el orden de la sección 2. Ninguno se une
con otro. Que no haya registros no es un error (RN-020-10).

## 8. Criterios de aceptación

- **CA-020-1** (normal) — Dado un proyecto con registros de dos integrantes en dos
  historias, al consultar sin filtros se devuelven todos, del más reciente al más
  antiguo, con el total de horas.
- **CA-020-2** (normal) — Al filtrar por una historia, solo aparecen sus registros y
  el total es el de esa historia; los totales por integrante cuentan solo lo
  filtrado.
- **CA-020-3** (normal) — Al filtrar por un integrante, solo aparecen sus
  registros, de todas las historias.
- **CA-020-4** (normal) — Al filtrar por un sprint, aparecen los registros cuya
  fecha cae en sus fechas (inicio y fin incluidos), sean de la historia que sean.
- **CA-020-5** (alternativo) — Los totales por historia, por integrante y por sprint
  salen ordenados según RN-020-9 y cada lista suma el total.
- **CA-020-6** (límite) — Un registro cuya fecha no cae en ningún sprint aparece como
  «fuera de todo sprint».
- **CA-020-7** (límite) — Sin registros, la consulta devuelve un resultado vacío con
  total 0 y sin error.
- **CA-020-8** (error) — Un filtro con una historia, un integrante o un sprint de otro
  proyecto se rechaza con el error de inexistencia que corresponde.
- **CA-020-9** (error) — Un proyecto inexistente se rechaza con
  `ErrProyectoInexistente`.

---

## Decisiones tomadas y descartadas

- **El sprint se deduce de la fecha al consultar.** Es lo que fijó RN-019-12; esta
  historia lo implementa con `SprintDeLaFecha` y no agrega una columna `sprint_id`
  a los registros (ver la discusión en la spec de US-019).
- **Un solo filtro por dimensión, combinados con Y.** Se descartó aceptar listas
  (varias historias o varios integrantes a la vez): ninguna pantalla lo pide y
  complica el contrato. Cada uno es un solo ID.
- **Filtrar y agrupar en memoria.** Se descartó una consulta SQL por filtro: la
  regla del sprint por fecha quedaría escrita dos veces (en SQL y en el dominio)
  y los volúmenes no lo justifican. El costo es leer todos los registros del
  proyecto en cada consulta; si crecieran, se pasa el filtro por proyecto y
  fecha a SQL sin cambiar el contrato.
- **Los totales se calculan sobre lo filtrado.** Con el filtro de una historia, el
  total por integrante responde «quién trabajó en esta historia», que es lo que
  se quiere ver. Se descartó mostrar siempre los totales del proyecto entero: el
  que quiera el total completo consulta sin filtros.
- **Solo grupos con registros.** Se descartó listar a todos los integrantes y todos
  los sprints con 0 horas: la pantalla puede completar ceros si los necesita, y
  el dominio no tiene que conocer la lista completa.
- **Fuera de todo sprint como grupo propio.** Se descartó dejar esos registros
  fuera del resultado: las horas se trabajaron, y ocultarlas haría que el total de la
  consulta no coincida con la suma de sus filas.
- **`Horas()` en lugar de un caso de uso aparte para US-027.** Las horas reales de
  un sprint son las de `ConsultaEsfuerzo` filtrada por ese sprint; así US-027 y
  US-020 no pueden dar cifras distintas.
- **Frases nuevas del diccionario** (se aprueban en el PR de implementación), área
  Esfuerzo: `se consulta el esfuerzo del proyecto "P"`, con las variantes
  `... de la historia "H"`, `... del integrante "I"` y `... del sprint "S"` (se
  pueden encadenar con `y`: `se consulta el esfuerzo del proyecto "P" de la
  historia "H" del integrante "I"`); `la consulta devuelve N registros`,
  `el total de horas de la consulta es N`, `la consulta devuelve primero el
  registro de "A"` (por actividad), `las horas de la historia "H" en la consulta
  son N`, `las horas del integrante "I" en la consulta son N`, `las horas del
  sprint "S" en la consulta son N` y `las horas fuera de todo sprint en la
  consulta son N`. Los números aceptan coma o punto decimal.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la
  historia y de las de US-019 y US-027 (que ya dejaban fijada la deducción del
  sprint por fecha); las decisiones sobre filtros, agrupación y orden se proponen
  acá y se confirman en la revisión
- Revisado y entendido por: Tomás Romero
