# US-002 — Modificar los datos de un proyecto existente

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Angelo Conforti |
| **Issue** | #5 |
| **Escenarios BDD** | `features/US-002-modificar-proyecto.feature` |
| **Código** | `internal/domain/project/` |
| **Última actualización** | 2026-10-05 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Permitir corregir los datos de un proyecto ya creado (nombre, descripción y
fechas de inicio y fin) sin tener que crearlo de nuevo. Los proyectos cambian:
se corrige un nombre mal escrito o se mueven las fechas de la cursada, y eso no
debe obligar a perder el backlog, los sprints ni los integrantes asociados.

Depende de **US-001**: reutiliza el tipo `Proyecto` y sus reglas de validación,
sin duplicarlas.

## 2. Entradas

La capa de aplicación busca el proyecto existente y se lo pasa a la función de
dominio junto con los datos nuevos. El dominio no accede a la base de datos.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyecto` | `Proyecto` (de US-001) | Sí | Un proyecto ya existente, cargado por la capa de aplicación |
| `nombre` | `string` | Sí | Entre 1 y 100 caracteres, contados después de quitar los espacios de los extremos (RN-001-1 y RN-001-2) |
| `descripcion` | `string` | No | Cualquier texto, incluida la cadena vacía (RN-001-5) |
| `fechaInicio` | `time.Time` | Sí | Fecha distinta del valor cero (RN-001-3) |
| `fechaFin` | `time.Time` | Sí | Fecha distinta del valor cero y no anterior a `fechaInicio` (RN-001-3 y RN-001-4) |

Se envían siempre los cuatro datos, no solo los que cambian: el formulario los
muestra precargados con los valores actuales.

## 3. Salidas esperadas

```
Proyecto {
    ID          string    // el mismo que tenía; nunca cambia
    Nombre      string    // el nuevo, sin espacios en los extremos
    Descripcion string    // la nueva
    FechaInicio time.Time // la nueva
    FechaFin    time.Time // la nueva
}
```

Si algún dato es inválido, o el proyecto no existe, no se modifica nada y se
devuelve solo el error correspondiente (sección 7). La función devuelve un
`Proyecto` nuevo; no altera el recibido.

## 4. Reglas de negocio

- **RN-002-1** — Solo se puede modificar un proyecto que existe. Si no existe, la
  operación se rechaza. Lo detecta la capa de aplicación al buscarlo, antes de
  llamar al dominio.
- **RN-002-2** — Los datos nuevos cumplen exactamente las mismas reglas que al
  crear un proyecto (RN-001-1 a RN-001-5). Se reutiliza la misma validación de
  US-001 y no se escribe una segunda.
- **RN-002-3** — La modificación reemplaza los cuatro datos a la vez. No existe
  la modificación parcial.
- **RN-002-4** — El identificador del proyecto no cambia.
- **RN-002-5** — La operación es de todo o nada: si algún dato es inválido, el
  proyecto queda exactamente como estaba.
- **RN-002-6** — Modificar un proyecto sin cambiar ningún dato es válido y no
  produce un error.
- **RN-002-7** — Los integrantes y demás datos asociados al proyecto no se tocan.

## 5. Restricciones

- No usa `time.Now()`, números aleatorios ni variables de entorno, y no importa
  nada fuera de la biblioteca estándar. La marca `actualizado_en` de la tabla
  `proyectos` la asigna la persistencia al guardar; el dominio no la maneja.
- No se exige que el nuevo nombre sea único: puede coincidir con el de otro
  proyecto (igual que en US-001) y también con el del propio proyecto.
- No se valida el rango de fechas contra los sprints del proyecto (ver
  "Decisiones").
- Si hay más de un dato inválido, se devuelve el primer error, en el mismo orden
  que en US-001: nombre vacío, nombre largo, falta fecha de inicio, falta fecha
  de fin, fechas incoherentes.

## 6. Casos límite

- **CL-002-1** — Se envían exactamente los mismos datos que ya tenía el
  proyecto: es válido y el proyecto queda igual (RN-002-6).
- **CL-002-2** — Fecha de fin igual a la de inicio: es válido, igual que al
  crear.
- **CL-002-3** — Nuevo nombre con espacios al principio o al final: es válido y
  queda recortado.
- **CL-002-4** — Nuevo nombre igual al de otro proyecto existente: es válido,
  porque el nombre no es único.
- **CL-002-5** — Nuevo nombre de exactamente 100 caracteres: es válido. Con 101:
  se rechaza con `ErrNombreLargo`.
- **CL-002-6** — Descripción vacía sobre un proyecto que tenía descripción: es
  válido y sirve para borrarla.
- **CL-002-7** — Un dato inválido junto con datos válidos (nombre nuevo correcto
  pero fechas invertidas): no se aplica ni siquiera el nombre (RN-002-5).

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| El proyecto a modificar no existe | `ErrProyectoNoEncontrado` (capa de aplicación) | "el proyecto no existe" |
| El nombre nuevo está vacío o tiene solo espacios | `ErrNombreVacio` (US-001) | "el nombre es obligatorio" |
| El nombre nuevo supera los 100 caracteres | `ErrNombreLargo` (US-001) | "el nombre supera el largo maximo (100 caracteres)" |
| Falta la fecha de inicio | `ErrFechaInicioFaltante` (US-001) | "falta la fecha de inicio" |
| Falta la fecha de fin | `ErrFechaFinFaltante` (US-001) | "falta la fecha de fin" |
| La fecha de fin es anterior a la de inicio | `ErrFechasIncoherentes` (US-001) | "fecha de fin anterior a la de inicio" |

Los errores de dominio son los mismos valores de US-001: no se definen otros
para decir lo mismo. La capa de aplicación los envuelve con contexto (`%w`)
indicando el proyecto afectado.

## 8. Criterios de aceptación

- **CA-002-1** (normal) — Dado un proyecto existente, al modificar su nombre el
  proyecto queda con el nombre nuevo y conserva sus fechas.
- **CA-002-2** (alternativo) — Dado un proyecto existente, al modificar sus
  fechas el proyecto queda con las fechas nuevas y conserva su nombre.
- **CA-002-3** (límite) — Dado un proyecto existente, al modificar sus fechas
  para que inicio y fin sean iguales, la modificación es válida.
- **CA-002-4** (error) — Dado un nombre nuevo vacío, la operación se rechaza con
  el mensaje "el nombre es obligatorio" y el proyecto conserva su nombre
  anterior.
- **CA-002-5** (error) — Dado un proyecto que no existe, la operación se
  rechaza con el mensaje "el proyecto no existe".

---

## Decisiones tomadas y descartadas

- **Modificación completa y no parcial.** Se envían los cuatro datos y se
  reemplazan todos (RN-002-3). Se descartó la modificación parcial (solo los
  campos informados) porque obliga a distinguir "no informado" de "vacío" y, con
  la descripción opcional, un `""` sería ambiguo: ¿borrar o no tocar?
- **Se reutiliza la validación de US-001.** Las reglas de datos viven en un solo
  lugar; si cambia el largo máximo del nombre, cambia para crear y para
  modificar a la vez. Se descartó copiar las validaciones en una función
  separada.
- **No se valida el rango de fechas contra los sprints.** Todavía no existen
  sprints (US-008 es posterior). Cuando existan, mover las fechas del proyecto
  podría dejar un sprint fuera del rango; eso se resolverá en una historia
  propia con su spec, no con un cambio silencioso acá. Se anota como alcance
  diferido.
- **No se guarda historial de cambios.** Se descartó registrar los valores
  anteriores porque ningún criterio lo pide y agrega una tabla nueva.
- **La marca `actualizado_en` la pone la persistencia.** Así el dominio sigue
  sin depender del reloj (no usa `time.Now()`). Se descartó pasarla como
  parámetro porque nada de la lógica de negocio la necesita.
- **Frase de fechas pendiente de acordar con el equipo.** El diccionario de
  steps solo tiene `se modifica el nombre del proyecto "X" a "Y"`, que cubre
  CA-002-1, CA-002-4 y CA-002-5. Para CA-002-2 y CA-002-3 hace falta una frase de
  fechas; se propone `se modifican las fechas del proyecto "X" del "..." al
  "..."`. Hasta que se apruebe e incorpore al diccionario, esos dos criterios se
  verifican con tests unitarios y no tienen escenario BDD. No se inventa la
  frase en el `.feature`.

## Uso de IA en esta especificación

- [x] Use IA para: redacción del borrador de los apartados de la especificación
  (reglas, casos límite, errores y criterios de aceptación) a partir de la
  historia y de la spec de US-001 — revisado por: Angelo Conforti.