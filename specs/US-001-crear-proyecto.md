# US-001 — Crear un proyecto con nombre, descripción y fechas de inicio y fin

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Angelo Conforti |
| **Issue** | #4 |
| **Escenarios BDD** | `features/US-001-crear-proyecto.feature` |
| **Código** | `internal/domain/proyecto/`, caso de uso en `internal/app/`, migración nueva en `migraciones/` |
| **Última actualización** | 2026-10-05 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Permitir crear un proyecto de software con su nombre, una descripción opcional y
sus fechas de inicio y fin. Es el punto de partida de todo lo demás: el backlog,
los sprints, el esfuerzo y las métricas cuelgan de un proyecto, así que ninguna
otra historia se puede usar hasta que esta exista.

Esta historia define el tipo `Proyecto` y sus validaciones. Las historias US-002
(modificar) y US-003 (integrantes) las reutilizan. También alinea con US-005 la
forma de asignar el identificador y de informar los errores de validación.

## 2. Entradas

Función de dominio que valida los datos y arma el `Proyecto`. No accede a la base
de datos.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `nombre` | `string` | Sí | Entre 1 y 100 caracteres, contados después de quitar los espacios de los extremos |
| `descripcion` | `string` | No | Cualquier texto, incluida la cadena vacía; sin límite de largo |
| `fechaInicio` | `time.Time` | Sí | Fecha distinta del valor cero |
| `fechaFin` | `time.Time` | Sí | Fecha distinta del valor cero y no anterior a `fechaInicio` |

El identificador no es una entrada: lo genera la base de datos
(`DEFAULT gen_random_uuid()`) y la persistencia lo devuelve con `RETURNING`,
igual que en US-005.

Las fechas llegan normalizadas a medianoche UTC (formato `AAAA-MM-DD` en la
interfaz); el dominio solo las compara.

## 3. Salidas esperadas

```
Proyecto {
    ID          string    // vacío hasta que la persistencia lo asigna
    Nombre      string    // sin espacios en los extremos
    Descripcion string    // tal como se recibió; "" si no se informó
    FechaInicio time.Time
    FechaFin    time.Time
}
```

El dominio devuelve el `Proyecto` sin ID; el caso de uso devuelve el proyecto ya
registrado, con ID.

Si algún dato es inválido no se devuelve ningún `Proyecto`, solo los errores
correspondientes (sección 7). Las marcas de creación y actualización
(`creado_en`, `actualizado_en`) no forman parte del dominio: las completa la
persistencia.

## 4. Reglas de negocio

- **RN-001-1** — El nombre es obligatorio. Se descartan los espacios de los
  extremos y el resultado no puede quedar vacío.
- **RN-001-2** — El nombre tiene como máximo 100 caracteres, contados después de
  quitar los espacios de los extremos.
- **RN-001-3** — La fecha de inicio y la de fin son obligatorias.
- **RN-001-4** — La fecha de fin no puede ser anterior a la de inicio. Que sean
  iguales es válido (proyecto de un solo día).
- **RN-001-5** — La descripción es opcional: si no se informa, queda vacía. Se
  guarda tal cual se recibe, sin recortar.
- **RN-001-6** — El identificador lo genera la base de datos al guardar. El
  dominio no lo recibe, no lo genera ni lo modifica.
- **RN-001-7** — El dominio valida todos los datos y devuelve todos los errores
  unidos con `errors.Join`, en este orden fijo: nombre (vacío o largo), falta
  fecha de inicio, falta fecha de fin, fechas incoherentes.
  `ErrFechasIncoherentes` solo se evalúa si las dos fechas están. Cada error
  unido se reconoce con `errors.Is`.

## 5. Restricciones

- No usa `time.Now()`, números aleatorios ni variables de entorno, y no importa
  nada fuera de la biblioteca estándar.
- Persistencia: una migración nueva en `migraciones/` agrega
  `DEFAULT gen_random_uuid()` a `proyectos.id`. Nunca se edita `00001_init.sql`.
- El largo del nombre se mide en **caracteres**, no en bytes: `ñ` o `é` cuentan
  como uno.
- No se exige que el nombre sea único entre proyectos: la base de datos no
  impone esa restricción (`00001_init.sql`) y no hay una regla de negocio que la
  pida.
- Las reglas son coherentes con las restricciones de la tabla `proyectos`
  (`fechas_coherentes` y `nombre_no_vacio`). El dominio es la primera defensa y
  la base de datos la última; el máximo de 100 caracteres solo lo impone el
  dominio.
- Si hay más de un dato inválido se devuelven todos, unidos, en el orden de
  RN-001-7 (igual que US-005).

## 6. Casos límite

- **CL-001-1** — Fecha de fin igual a la de inicio: es válido (RN-001-4). Un
  proyecto de un solo día no es un error.
- **CL-001-2** — Nombre formado solo por espacios: se trata como vacío y se
  rechaza con `ErrNombreVacio`.
- **CL-001-3** — Nombre con espacios al principio o al final (`"  Demo  "`): es
  válido y el proyecto queda con el nombre `"Demo"`.
- **CL-001-4** — Nombre de exactamente 100 caracteres: es válido. Con 101: se
  rechaza con `ErrNombreLargo`.
- **CL-001-5** — Nombre de 100 caracteres no ASCII (por ejemplo cien `ñ`, que
  ocupan 200 bytes): es válido, porque el límite cuenta caracteres.
- **CL-001-6** — Descripción vacía u omitida: es válido y queda como cadena
  vacía.
- **CL-001-7** — Varios datos inválidos a la vez (nombre vacío y fechas
  invertidas): se devuelven los dos errores unidos, en el orden de RN-001-7.
  `errors.Is` reconoce cada uno.
- **CL-001-8** — Falta una de las dos fechas y la otra está informada: se
  devuelve solo el error de la fecha faltante; `ErrFechasIncoherentes` no se
  evalúa porque no hay con qué comparar.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| El nombre está vacío o tiene solo espacios | `ErrNombreVacio` | "el nombre es obligatorio" |
| El nombre supera los 100 caracteres | `ErrNombreLargo` | "el nombre supera el largo maximo (100 caracteres)" |
| Falta la fecha de inicio | `ErrFechaInicioFaltante` | "falta la fecha de inicio" |
| Falta la fecha de fin | `ErrFechaFinFaltante` | "falta la fecha de fin" |
| La fecha de fin es anterior a la de inicio | `ErrFechasIncoherentes` | "fecha de fin anterior a la de inicio" |

Los errores son valores del paquete (`errors.New`) y se devuelven unidos con
`errors.Join` (RN-001-7), así que se comparan con `errors.Is`. La capa de
aplicación los envuelve con contexto (`%w`) indicando el nombre del proyecto; el
mensaje de cada error queda dentro del error resultante, por lo que el escenario
BDD puede comprobarlo con `el mensaje de error indica "..."`.

## 8. Criterios de aceptación

- **CA-001-1** (normal) — Dado un nombre válido y fechas coherentes, el proyecto
  se crea y queda disponible con ese nombre y esas fechas.
- **CA-001-2** (alternativo) — Dado un proyecto ya existente, crear otro proyecto
  con un nombre distinto es válido y el primero no cambia.
- **CA-001-3** (límite) — Dadas una fecha de inicio y una de fin iguales, el
  proyecto se crea.
- **CA-001-4** (error) — Dada una fecha de fin anterior a la de inicio, la
  operación se rechaza con el mensaje "fecha de fin anterior a la de inicio" y no
  se crea ningún proyecto.
- **CA-001-5** (error) — Dado un nombre vacío, la operación se rechaza con el
  mensaje "el nombre es obligatorio" y no se crea ningún proyecto.

---

## Decisiones tomadas y descartadas

- **La descripción es opcional.** La tabla `proyectos` define
  `descripcion TEXT NOT NULL DEFAULT ''`, es decir, ya prevé que no se informe.
  Obligarla agregaría fricción sin una regla de negocio detrás. Se descartó
  exigir un mínimo de caracteres por la misma razón.
- **Máximo de 100 caracteres para el nombre.** Es suficiente para un título y
  evita que un nombre desmesurado rompa el Dashboard y el reporte PDF. Se
  descartó no poner límite porque la interfaz tendría que lidiar con cualquier
  largo.
- **No se exige nombre único.** Dos proyectos pueden llamarse igual. Se descartó
  la unicidad porque agrega un error y un caso de borde que ningún criterio de
  la guía pide.
- **El identificador lo genera la base de datos** (`DEFAULT gen_random_uuid()`).
  Alinea con US-005, evita una dependencia de UUID en la app y mantiene el
  dominio puro.
- **Se devuelven todos los errores juntos:** así quien carga el formulario ve
  todo lo que tiene que corregir, y alinea con US-005 y con la traducción
  uniforme de errores de T-005.
- **La descripción no se prueba en el escenario BDD.** La frase del diccionario
  de steps `se crea un proyecto "X" con fechas del "..." al "..."` no recibe
  descripción. Su regla (RN-001-5) y su caso límite (CL-001-6) se cubren con
  tests unitarios de dominio. Si más adelante se acuerda una frase con
  descripción, se agrega un escenario.
- **Las reglas RN-001-2 y CL-001-3 a CL-001-5 se cubren con tests unitarios.**
  Se pueden expresar en escenarios, pero obligarían a escribir un nombre de 100
  caracteres en el `.feature`, lo que lo vuelve ilegible.

## Uso de IA en esta especificación

- [x] Use IA para: redacción del borrador de los apartados de la especificación
  (reglas, casos límite, errores y criterios de aceptación) a partir de la
  historia y de la tabla `proyectos` — revisado por: Angelo Conforti.