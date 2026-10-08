# US-003 — Registrar integrantes en un proyecto

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Angelo Conforti |
| **Issue** | #6 |
| **Escenarios BDD** | `features/US-003-registrar-integrantes.feature` |
| **Código** | `internal/domain/integrante/`, caso de uso en `internal/app/`, repositorio en `internal/adapters/postgres/`, migración nueva en `migraciones/` |
| **Última actualización** | 2026-10-07 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero registrar a las personas que trabajan en un
> proyecto, con su nombre, email y rol, para saber quién participa y poder
> asignarle trabajo y esfuerzo más adelante.

Registrar integrantes en un proyecto existente. Es la base de las historias que
necesitan saber quién hizo qué, como el registro de esfuerzo (US-019). Depende de
**US-001**: el integrante cuelga de un `Proyecto`.

## 2. Entradas

La operación pasa por tres capas, y cada una valida solo lo que puede saber:

- **Dominio** (`internal/domain/integrante`): función pura que valida los datos y
  devuelve un `Integrante` o todos los errores encontrados. No genera IDs, no sabe
  si el proyecto existe ni si el email ya está usado.
- **Aplicación** (`internal/app`): verifica primero que el proyecto exista; si no,
  devuelve `ErrProyectoInexistente` sin validar nada más. Después valida con el
  dominio y registra a través del repositorio.
- **Persistencia** (`internal/adapters/postgres`): inserta y devuelve el ID con
  `RETURNING`. La restricción `UNIQUE (proyecto_id, email)` es la que detecta el
  email repetido.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. Viene del contexto de navegación (la URL), no del formulario. Lo verifica la capa de aplicación, nunca el dominio |
| `nombre` | `string` | Sí | Entre 1 y 100 caracteres, contados en runas después de quitar los espacios de los extremos |
| `email` | `string` | Sí | Después de quitar los espacios de los extremos, una dirección simple válida (`usuario@dominio`), sin nombre visible ni corchetes |
| `rol` | `Rol` | Sí | Exactamente `product_architect`, `agile_enabler` o `product_builder`, en minúscula |

No son entradas: el ID (lo genera la base) ni `creado_en` (lo completa la
persistencia).

## 3. Salidas esperadas

```
Integrante {
    ID         string // vacío hasta que la persistencia lo asigna
    ProyectoID string // el proyecto recibido
    Nombre     string // sin espacios en los extremos
    Email      string // sin espacios en los extremos y en minúscula
    Rol        Rol    // el recibido
}
```

El dominio devuelve el `Integrante` sin ID; el caso de uso devuelve el integrante
ya registrado, con ID. Si hay errores de validación no se devuelve ningún
integrante: se devuelve un único error que une todos los encontrados (RN-003-8).

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-003-1** (dominio) — El nombre es obligatorio. Se recortan los espacios de
  los extremos y el resultado no puede quedar vacío.
- **RN-003-2** (dominio) — El nombre recortado tiene como máximo 100 caracteres,
  contados en runas y no en bytes.
- **RN-003-3** (dominio) — El email es obligatorio. Se recorta y se valida con
  `net/mail` de la biblioteca estándar: tiene que ser una dirección simple, sin
  nombre visible. Un email vacío o con formato inválido se rechaza.
- **RN-003-4** (dominio) — El email se guarda en minúscula, para que la unicidad
  no dependa de cómo se escribió.
- **RN-003-5** (dominio) — El rol es obligatorio y solo puede ser
  `product_architect`, `agile_enabler` o `product_builder`, exactos y en
  minúscula. Cualquier otro valor se rechaza, incluidos `"PRODUCT_BUILDER"` y
  `" agile_enabler "`.
- **RN-003-6** (aplicación) — El integrante pertenece a un proyecto existente. El
  caso de uso lo verifica antes de validar los datos; si no existe devuelve
  `ErrProyectoInexistente` y no se registra nada. Ese error no se une con los de
  validación.
- **RN-003-7** (persistencia) — El email es único dentro de un proyecto. Registrar
  un email ya usado en el mismo proyecto se rechaza con `ErrEmailDuplicado`. El
  mismo email en otro proyecto es válido.
- **RN-003-8** (dominio) — El dominio valida todos los datos y devuelve todos los
  errores unidos con `errors.Join`, en este orden fijo: nombre, email, rol. Cada
  error unido se reconoce con `errors.Is`.
- **RN-003-9** (aplicación y persistencia) — La operación es de todo o nada: si
  algo falla no se registra nada.
- **RN-003-10** (persistencia) — El ID lo genera la base
  (`DEFAULT gen_random_uuid()`) y se devuelve con `RETURNING`.
- **RN-003-11** (dominio) — No hay máximo de integrantes por proyecto, y varios
  integrantes pueden tener el mismo rol.

## 5. Restricciones

- **Dominio puro.** `internal/domain/integrante` no importa nada fuera de la
  biblioteca estándar, no hace SQL ni HTTP, no genera IDs y no usa `time.Now()`.
- **Lo que el dominio no puede saber** (que el proyecto exista, que el email ya
  esté usado, qué ID tiene) lo resuelven la capa de aplicación y el repositorio.
- **Conversión del rol.** Convertir el texto de un formulario a `Rol` es tarea del
  adaptador: el dominio no pasa a minúscula ni recorta el rol.
- **Persistencia.** Una migración nueva en `migraciones/` agrega
  `DEFAULT gen_random_uuid()` a `integrantes.id`. Nunca se edita
  `00001_init.sql`. El número de la migración se avisa en el grupo antes de crearla.
- **Limitación conocida del email.** `net/mail` acepta direcciones sin punto en el
  dominio (por ejemplo `a@b`). No se agrega una validación más estricta para no
  sumar dependencias ni reglas que nadie pidió.
- **Concurrencia.** Dos altas simultáneas con el mismo email en el mismo proyecto:
  una termina bien y la otra devuelve `ErrEmailDuplicado`. Se verifica con un test
  de integración contra un Postgres real.
- **Fuera de alcance.** Modificar o dar de baja integrantes, la pantalla de alta
  (T-011), asignar trabajo a un integrante (US-019) y validar que el proyecto tenga
  exactamente los tres roles de la cátedra.

## 6. Casos límite

- **CL-003-1** — Nombre de exactamente 100 caracteres: se acepta. Con 101: se
  rechaza con `ErrNombreLargo`.
- **CL-003-2** — Nombre de 100 caracteres con `ñ` o tildes: se acepta, porque el
  límite cuenta caracteres y no bytes.
- **CL-003-3** — Nombre con solo espacios: se trata como vacío (`ErrNombreVacio`),
  no como largo.
- **CL-003-4** — Nombre con espacios al principio o al final: se acepta y se guarda
  recortado.
- **CL-003-5** — Email con espacios en los extremos o con mayúsculas
  (`" Tomas@Ejemplo.Test "`): se acepta y se guarda como `tomas@ejemplo.test`.
- **CL-003-6** — Mismo email con otra capitalización en el mismo proyecto
  (`TOMAS@ejemplo.test` cuando ya existe `tomas@ejemplo.test`): se rechaza con
  `ErrEmailDuplicado`.
- **CL-003-7** — Mismo email en dos proyectos distintos: es válido en los dos.
- **CL-003-8** — Mismo nombre con distinto email en el mismo proyecto: es válido.
- **CL-003-9** — Email en formato `"Tomas <tomas@ejemplo.test>"`: se rechaza con
  `ErrEmailInvalido`, porque no es una dirección simple.
- **CL-003-10** — Rol `"PRODUCT_BUILDER"`, `" agile_enabler "` o vacío: se rechaza
  con `ErrRolInvalido`.
- **CL-003-11** — Tres errores a la vez (nombre vacío, email `"sin-arroba"` y rol
  `"jefe"`): se devuelven los tres unidos, en el orden `ErrNombreVacio`,
  `ErrEmailInvalido`, `ErrRolInvalido`. `errors.Is` reconoce cada uno.
- **CL-003-12** — Primer integrante de un proyecto: es válido y el proyecto pasa a
  tener 1.
- **CL-003-13** — Proyecto inexistente con datos inválidos: se devuelve solo
  `ErrProyectoInexistente`, sin mirar los datos (RN-003-6).
- **CL-003-14** — Dos altas simultáneas con el mismo email en el mismo proyecto:
  una termina bien y la otra devuelve `ErrEmailDuplicado`.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| Nombre vacío o con solo espacios | `integrante.ErrNombreVacio` | "el nombre es obligatorio" |
| Nombre de más de 100 caracteres | `integrante.ErrNombreLargo` | "el nombre supera el largo maximo (100 caracteres)" |
| Email vacío | `integrante.ErrEmailVacio` | "el email es obligatorio" |
| Email con formato inválido | `integrante.ErrEmailInvalido`, envuelto indicando el valor recibido | "el email no tiene un formato valido" |
| Rol distinto de los tres permitidos | `integrante.ErrRolInvalido`, envuelto indicando el valor recibido | "el rol tiene que ser product_architect, agile_enabler o product_builder" |
| El proyecto no existe | `app.ErrProyectoInexistente` (el mismo de US-005) | "el proyecto no existe" |
| El email ya está registrado en el proyecto | `app.ErrEmailDuplicado`, envuelto indicando el email | "ya hay un integrante con ese email en el proyecto" |

Los cinco primeros son del dominio, se detectan juntos en una sola validación y se
devuelven unidos con `errors.Join` (RN-003-8). `ErrProyectoInexistente` lo detecta
la capa de aplicación antes de validar y no se une con los demás.
`ErrEmailDuplicado` lo detecta el repositorio al chocar con
`UNIQUE (proyecto_id, email)`. Todos son valores de paquete comparables con
`errors.Is`, y la capa de aplicación los envuelve con contexto (`%w`).

## 8. Criterios de aceptación

- **CA-003-1** (normal) — Con un proyecto existente, al registrar un integrante con
  nombre, email y rol válidos, el proyecto pasa a tener ese integrante.
- **CA-003-2** (alternativo) — Se pueden registrar varios integrantes con roles
  distintos en el mismo proyecto, y el proyecto los cuenta todos.
- **CA-003-3** (alternativo) — El mismo email puede registrarse en dos proyectos
  distintos.
- **CA-003-4** (límite) — Un nombre de 100 caracteres se acepta y uno de 101 se
  rechaza; el email se guarda recortado y en minúscula.
- **CA-003-5** (error) — Nombre vacío o solo espacios → rechazo con "el nombre es
  obligatorio" y no se registra nada.
- **CA-003-6** (error) — Email vacío o con formato inválido → rechazo.
- **CA-003-7** (error) — Rol fuera de los tres permitidos → rechazo.
- **CA-003-8** (error) — Email ya registrado en el proyecto (sin importar
  mayúsculas) → rechazo con "ya hay un integrante con ese email en el proyecto".
- **CA-003-9** (error) — Proyecto inexistente → rechazo con "el proyecto no
  existe" y no se registra nada.

---

## Decisiones tomadas y descartadas

- **Paquete `internal/domain/integrante`.** Singular y en español, según el
  ADR 0003.
- **El email se normaliza a minúscula en el dominio.** La restricción `UNIQUE` de
  la base distingue mayúsculas; si no se normalizara, `A@x.com` y `a@x.com`
  serían dos integrantes distintos. Se descartó una restricción
  `UNIQUE (proyecto_id, lower(email))` para no tocar el esquema más de lo necesario.
- **Unicidad del email por proyecto y no global.** Es lo que ya impone
  `00001_init.sql`, y una misma persona puede participar de varios proyectos.
- **Validación del email con `net/mail`.** Es biblioteca estándar y mantiene el
  dominio puro. Se descartó una expresión regular propia porque es fácil de
  equivocar y no la pide ningún criterio. Se acepta la limitación documentada en
  Restricciones.
- **Sin máximo de integrantes ni exclusividad de rol.** Ningún criterio de la guía
  lo pide. Se descartó exigir un único `product_architect` porque ese rol es de la
  cátedra y no siempre está cargado.
- **El proyecto inexistente no se une con los errores de validación y tiene
  prioridad.** Misma decisión que US-005: el proyecto viene de la URL, no del
  formulario.
- **`ErrEmailDuplicado` vive en `internal/app`.** Lo detecta la persistencia, no el
  dominio, igual que `ErrProyectoInexistente`.
- **Frase de diccionario con email pendiente de acordar con el equipo.** Las frases
  actuales (`el integrante "X" con rol "Y"` y `se registra al integrante "X" con
  rol "Y"`) no reciben email. Se propone, por ejemplo,
  `se registra al integrante "X" con email "..." y rol "Y"`. Hasta que se apruebe e
  incorpore al diccionario, CA-003-4, CA-003-6 y CA-003-8 se verifican con tests
  unitarios y de integración, y no tienen escenario BDD. No se inventa la frase en
  el `.feature`.
- **La migración del `DEFAULT` del ID es un archivo aparte.** Se numera cuando el
  grupo confirme cuál sigue libre.

## Uso de IA en esta especificación

- [x] Use IA para: redacción del borrador de los apartados de la especificación
  (reglas, casos límite, errores y criterios de aceptación) a partir de la
  historia, la tabla `integrantes` y las specs de US-001, US-002 y US-005 —
  revisado por: Angelo Conforti.