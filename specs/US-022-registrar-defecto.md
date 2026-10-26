# US-022 — Registrar un defecto con severidad, historia relacionada y sprint de detección

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Angelo Conforti |
| **Issue** | #30 |
| **Escenarios BDD** | `features/US-022-registrar-defecto.feature` |
| **Código** | `internal/domain/defecto/`, caso de uso en `internal/app/`, repositorios en `internal/adapters/postgres/` y `internal/adapters/memoria/`, migración nueva `migraciones/00012_defectos.sql` |
| **Última actualización** | 2026-10-25 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero registrar un defecto con su severidad, la
> historia donde se introdujo y el sprint en el que lo detectamos, para llevar
> la cuenta de los defectos y medir cómo evoluciona la calidad del producto de
> un sprint a otro.

Cargar un defecto de un proyecto con los campos del requerimiento 6 de la guía,
tal como los define la plantilla `.github/ISSUE_TEMPLATE/02-defect.yml`: su
descripción, su severidad, la historia relacionada y el sprint de detección.
Todo defecto nace `abierto` y sin sprint de resolución.

Es la primera historia de la épica de defectos. Es la materia prima de cambiar
el estado y registrar el sprint de resolución (US-023), de listar defectos por
sprint y severidad (US-024) y, a través de ellas, de los defectos detectados y
resueltos por sprint (US-029).

## 2. Entradas

La carga pasa por dos capas, y cada una valida solo lo que puede saber:

- **Dominio** (`internal/domain/defecto`): función pura que recibe los datos del
  defecto, los valida todos y devuelve un `Defecto` válido o todos los errores
  encontrados. No genera IDs, no lee el reloj y no sabe si el proyecto, la
  historia o el sprint existen.
- **Aplicación y persistencia** (`internal/app` y `internal/adapters/postgres`):
  el caso de uso sigue este orden:
  1. Verifica con una lectura simple que el proyecto exista. Si no existe,
     devuelve `ErrProyectoInexistente` sin verificar nada más.
  2. Verifica que la historia exista **dentro de ese proyecto**
     (`ErrItemInexistente`).
  3. Verifica que el sprint de detección exista **dentro de ese proyecto**
     (`ErrSprintInexistente`).
  4. Valida los datos con el dominio, que devuelve todos los errores unidos.
  5. Registra el defecto a través del repositorio, con un único `INSERT`. El ID
     lo genera la base.

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `proyectoID` | UUID | Sí | Un proyecto existente. Viene del contexto de navegación (la URL); lo verifica el caso de uso, nunca el dominio |
| `itemID` | UUID | Sí | La historia relacionada: una historia existente del proyecto, en cualquier estado. Viene de la selección del formulario; la verifica el caso de uso |
| `sprintDeteccionID` | UUID | Sí | El sprint donde se detectó el defecto: un sprint existente del proyecto, abierto o cerrado, sea o no el de la historia. Viene de la selección del formulario; lo verifica el caso de uso |
| `descripcion` | `string` | Sí | Después de recortar los extremos con `strings.TrimSpace`: entre 1 y 2000 caracteres, contados en runas. Admite saltos de línea |
| `severidad` | `string` | Sí | Exactamente uno de `critica`, `alta`, `media` o `baja`: en minúscula, sin tilde y sin espacios alrededor |

Un ID vacío o mal formado es lo mismo que uno inexistente: el repositorio
responde que no existe, no que falló. Por eso, dejar sin elegir la historia o el
sprint de detección se rechaza con el mismo error que elegir uno que no existe.

No son entradas: el ID (lo genera la base), el estado (siempre `abierto`), el
sprint de resolución (siempre vacío, lo completa US-023) ni la fecha de carga
(`creado_en`, `DEFAULT now()`).

La plantilla de defectos muestra la severidad como "Crítica", "Alta", "Media" y
"Baja". Esas son las etiquetas que ve la persona; el valor que viaja al dominio
y se guarda es el de la tabla de arriba. Convertir una en otra es tarea del
adaptador HTTP, no del dominio.

## 3. Salidas esperadas

Un `Defecto` con estos atributos:

| Atributo | Valor al registrar |
|---|---|
| ID | UUID generado por la base (`DEFAULT gen_random_uuid()`) y devuelto con `RETURNING` |
| Proyecto | El recibido |
| Historia relacionada | La recibida (su ID) |
| Sprint de detección | El recibido (su ID, no su número) |
| Descripción | La recibida, recortada |
| Severidad | La recibida |
| Estado | `abierto` |
| Sprint de resolución | Vacío (`NULL` en la base). Solo lo completa US-023 |
| Fecha de carga | La que asigna la base |

```
func Nuevo(proyectoID, itemID, sprintDeteccionID, descripcion string, severidad Severidad) (Defecto, error)
```

El dominio devuelve el defecto todavía sin ID; el caso de uso devuelve el
defecto ya guardado, con ID. Si hay errores de validación, el dominio no
devuelve defecto: devuelve un único error que une todos los encontrados
(RN-022-4).

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-022-1** (dominio) — La descripción se recorta con `strings.TrimSpace`
  antes de validarla, y lo que se guarda es la descripción recortada. No puede
  quedar vacía (`ErrDescripcionVacia`) y tiene como máximo 2000 caracteres
  contados en runas (`ErrDescripcionMuyLarga`). Los saltos de línea del medio se
  conservan: la plantilla de defectos pide "esperado", "observado" y "pasos para
  reproducir" en líneas separadas.
- **RN-022-2** (dominio) — La severidad es una de `critica`, `alta`, `media` o
  `baja`, exactas. Cualquier otro valor, incluido el vacío, se rechaza con
  `ErrSeveridadInvalida`. El dominio no pasa a minúscula ni recorta.
- **RN-022-3** (dominio) — Todo defecto nace con estado `abierto` y sin sprint
  de resolución. Ni el dominio ni el caso de uso reciben esos dos datos: no se
  puede registrar un defecto ya resuelto. Cambiar el estado y completar el
  sprint de resolución es de US-023.
- **RN-022-4** (dominio) — El dominio valida todos los datos y devuelve todos
  los errores juntos, unidos con `errors.Join`, en este orden fijo: primero el
  de la descripción y después el de la severidad. Cada error unido sigue siendo
  reconocible con `errors.Is`. De la descripción se devuelve un solo error: el
  de vacía y el de muy larga no pueden darse a la vez.
- **RN-022-5** (aplicación) — El proyecto, la historia y el sprint de detección
  tienen que existir, y la historia y el sprint tienen que ser **de ese
  proyecto**. El caso de uso los verifica en el orden de la sección 2, antes de
  validar los datos; el primero que falla se devuelve y no se une con los
  errores de validación (sección 7).
- **RN-022-6** (aplicación) — El sprint de detección es obligatorio y puede ser
  cualquier sprint del proyecto: abierto o cerrado, anterior, posterior o
  distinto del sprint de la historia. Se guarda como ID con clave foránea; el
  número del sprint (`sprint.Numero`) se obtiene de la tabla `sprints` cuando
  alguien lo necesita (US-029), no se copia al defecto.
- **RN-022-7** (aplicación) — Se puede registrar un defecto sobre una historia
  en cualquier estado (`pendiente`, `en_progreso` o `completado`) y esté o no
  asignada a un sprint. Registrar un defecto no cambia la historia.
- **RN-022-8** (aplicación) — No hay unicidad: dos defectos con la misma
  descripción, historia, sprint y severidad son dos defectos distintos, con ID
  distinto. Nadie sabe si es un duplicado o dos fallas parecidas.
- **RN-022-9** (aplicación) — Si cualquier verificación o validación falla, no
  se registra nada.
- **RN-022-10** (persistencia) — El defecto se guarda con un único `INSERT`. El
  ID es un UUID que genera la base (`DEFAULT gen_random_uuid()`) y el
  repositorio lo devuelve con `RETURNING`; el dominio no genera IDs. El
  `INSERT` no escribe el sprint de resolución, que queda `NULL`. Si el
  proyecto, la historia o el sprint desaparecen entre la verificación y el
  `INSERT`, el repositorio devuelve el error de inexistencia que corresponda
  (`ErrProyectoInexistente`, `ErrItemInexistente` o `ErrSprintInexistente`) y no
  guarda nada.

## 5. Restricciones

- **Dominio puro.** `internal/domain/defecto` no importa nada fuera de la
  biblioteca estándar, no hace SQL ni HTTP, no genera IDs y no llama a
  `time.Now()`: no necesita el reloj, porque la fecha de carga la pone la base.
  Sus errores se definen solo con la biblioteca estándar. Define los tipos
  `Defecto`, `Severidad` (con `SeveridadCritica`, `SeveridadAlta`,
  `SeveridadMedia` y `SeveridadBaja`) y `Estado` (con `EstadoAbierto`, el único
  valor que existe hasta que US-023 defina los demás).
- **Paquete nuevo.** `defecto` es un paquete propio, en singular como pide
  `CONTRIBUTING.md` (§1) y en español (ADR 0003). No existe otro paquete con ese
  nombre en `internal/domain`. El dominio no importa `backlog` ni `sprint`:
  recibe los IDs ya verificados.
- **Lo que el dominio no puede saber** (que el proyecto, la historia o el sprint
  existan y se correspondan) lo resuelven la capa de aplicación y el
  repositorio, en el orden de la sección 2. Las claves foráneas impiden un ID
  inexistente, pero no que la historia o el sprint sean de otro proyecto: eso
  lo verifica el caso de uso con una lectura por proyecto, como en US-019.
- **Persistencia.** La migración `00012_defectos.sql` (nunca se edita una
  mergeada) crea `defectos` con: `id UUID PRIMARY KEY DEFAULT
  gen_random_uuid()`; `proyecto_id`, `item_id` y `sprint_deteccion_id` `NOT
  NULL` con clave foránea a `proyectos`, `items_backlog` y `sprints`;
  `sprint_resolucion_id` con clave foránea a `sprints` y **sin** `NOT NULL`;
  `descripcion TEXT NOT NULL CHECK (char_length(btrim(descripcion)) BETWEEN 1
  AND 2000)`; `severidad TEXT NOT NULL CHECK (severidad IN ('critica', 'alta',
  'media', 'baja'))`; `estado TEXT NOT NULL DEFAULT 'abierto'`; `creado_en
  TIMESTAMPTZ NOT NULL DEFAULT now()`; y tres índices: por `proyecto_id`, por
  `item_id` y por `sprint_deteccion_id` (`idx_defectos_proyecto`,
  `idx_defectos_item` e `idx_defectos_sprint_deteccion`). Las cuatro claves
  foráneas llevan `ON DELETE CASCADE`, como en `registros_esfuerzo`.
  El `CHECK` de la severidad repite la regla del dominio como última defensa. El
  de `estado` no lista valores: los define US-023 (ver Decisiones).
- **Número de la migración y `goose`.** El plan de trabajo reserva "`00011` en
  adelante" para defectos; el PR #170 (abierto) le da `00011` a US-018 y deja
  "`00012` en adelante" para US-022 y US-023. Esta historia toma `00012`. Antes
  de abrir la rama de implementación hay que confirmar que `00012` sigue libre
  en `main`. `Migrar` (`internal/platform/db/migrar.go`) se configura con
  `goose.WithAllowOutofOrder(true)` y `make migrate` corre `goose
  -allow-missing`, así que una migración que llegue fuera de orden se aplica
  igual (T-012). Con las columnas `estado` y `sprint_resolucion_id` ya creadas,
  US-023 no necesita otra migración para guardar el cambio de estado.
- **Dependencias.** La implementación parte de US-001 (`proyectos`), US-005
  (`items_backlog`) y US-008 (`sprints`), las tres ya en `main`. No depende de
  US-023, US-024 ni US-029: son ellas las que dependen de esta. La spec se puede
  mergear antes.
- **Catálogo de errores (T-005).** Los tres errores nuevos se registran en
  `internal/adapters/http/catalogo.go` en el PR de implementación (el test del
  catálogo falla si no están), en un bloque `// defecto` antes del de
  `esfuerzo`, todos de tipo `validacion`: `defecto.ErrDescripcionVacia` y
  `defecto.ErrDescripcionMuyLarga` con el campo `descripcion`, y
  `defecto.ErrSeveridadInvalida` con el campo `severidad`.
  `app.ErrProyectoInexistente`, `app.ErrItemInexistente` y
  `app.ErrSprintInexistente` ya están registrados como `no_encontrado`. No hay
  exclusiones nuevas.
- **Área nueva en los steps.** Los escenarios de defectos necesitan un área
  propia, `Defectos`, en `features/steps/mundo.go` (se agrega a `Servicios`, con
  sus steps en `features/steps/defectos.go`, y se conecta en
  `features/servicios_test.go`), con la política del ADR 0002: cada historia
  conecta solo lo que necesita. El escenario se arma con las frases que ya
  existen (proyecto, historia, sprint) y se agregan estas, que se aprueban en el
  PR de implementación y se suman a `docs/diccionario-steps.md`:

  | Frase | Expresión |
  |---|---|
  | `Cuando se registra un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"` | `^se registra un defecto de severidad "([^"]*)" con la descripción "([^"]*)" en la historia "([^"]*)" detectado en el sprint "([^"]*)"$` |
  | `Cuando se registra un defecto de severidad "alta" con una descripción de 2001 caracteres en la historia "Crear proyecto" detectado en el sprint "Sprint 1"` | `^se registra un defecto de severidad "([^"]*)" con una descripción de (\d+) caracteres en la historia "([^"]*)" detectado en el sprint "([^"]*)"$` |
  | `Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"` (contexto) | `^un defecto de severidad "([^"]*)" con la descripción "([^"]*)" en la historia "([^"]*)" detectado en el sprint "([^"]*)"$` |
  | `Y el proyecto "Demo" tiene 2 defectos` | `^el proyecto "([^"]*)" tiene (\d+) defectos?$` |
  | `Y el defecto de la historia "Crear proyecto" tiene severidad "alta"` | `^el defecto de la historia "([^"]*)" tiene severidad "([^"]*)"$` |
  | `Y el defecto de la historia "Crear proyecto" tiene estado "abierto"` | `^el defecto de la historia "([^"]*)" tiene estado "([^"]*)"$` |
  | `Y el defecto de la historia "Crear proyecto" fue detectado en el sprint "Sprint 1"` | `^el defecto de la historia "([^"]*)" fue detectado en el sprint "([^"]*)"$` |
  | `Y el defecto de la historia "Crear proyecto" está sin resolver` | `^el defecto de la historia "([^"]*)" está sin resolver$` |

  La historia y el sprint se buscan por nombre dentro del proyecto actual (el
  sprint por su nombre `Sprint N`, que existe en todos los proyectos): uno que
  no esté ahí se manda con un ID que no existe, así el rechazo lo produce la
  aplicación y no los steps. Eso cubre "inexistente" y "de otro proyecto"
  (CL-022-8 y CL-022-9); el caso con un ID que sí existe en otro proyecto lo
  prueban el caso de uso y el repositorio. Las frases de consulta leen el
  último defecto registrado de esa historia. La descripción de N caracteres se
  arma con N letras `ñ`, para que el escenario también pruebe que el límite
  cuenta caracteres y no bytes. Una descripción de solo espacios no tiene frase:
  se prueba con tests unitarios.
- **Compatibilidad con US-029.** US-029 recibe una lista de
  `DefectoRegistrado{SprintDeteccion, SprintResolucion}` con los sprints por su
  número. Este modelo permite armarla: todo defecto registrado tiene un sprint
  de detección obligatorio y del proyecto (su número es `>= 1`, porque
  `sprints.numero` tiene `CHECK (numero > 0)`), y un sprint de resolución vacío
  que solo cambia US-023. La capa de aplicación traduce cada ID de sprint a su
  número con la tabla `sprints`. Esta historia no escribe esa consulta.
- **Fuera de alcance.** Cambiar el estado o resolver un defecto (US-023),
  listarlos (US-024), la pantalla de registro (depende de T-004), editar o
  borrar un defecto (no hay historia que lo pida), guardar quién lo registró y
  las métricas (US-029, US-030 y US-032).

## 6. Casos límite

- **CL-022-1** (RN-022-1) — Descripción de exactamente 2000 caracteres: se acepta. Con
  2001, se rechaza con `ErrDescripcionMuyLarga`.
- **CL-022-2** (RN-022-1) — Descripción con espacios, tabulaciones o saltos de línea al
  principio y al final: se recortan y se guarda sin ellos. Los saltos de línea
  del medio (`"Esperado: guardar\nObservado: error 500"`) se conservan.
- **CL-022-3** (RN-022-1) — Descripción vacía o solo con espacios: `ErrDescripcionVacia`,
  no `ErrDescripcionMuyLarga`, aunque tenga 3000 espacios.
- **CL-022-4** (RN-022-1) — Descripción de 2000 letras `ñ` (4000 bytes): se acepta, porque
  el límite cuenta runas y no bytes.
- **CL-022-5** (RN-022-2) — Severidad `critica`, `alta`, `media` o `baja`: se aceptan las
  cuatro. `Critica`, `CRITICA`, `crítica`, `" alta "`, `urgente` o la severidad
  vacía: se rechazan con `ErrSeveridadInvalida`, envuelto indicando el valor
  recibido.
- **CL-022-6** (RN-022-4) — Dos errores a la vez: descripción vacía y severidad `urgente`.
  El dominio devuelve los dos unidos, en este orden: `ErrDescripcionVacia`,
  `ErrSeveridadInvalida`. `errors.Is` reconoce cada uno.
- **CL-022-7** (RN-022-5, RN-022-9) — Proyecto inexistente, con la historia y el sprint también
  inexistentes y los datos inválidos: se devuelve solo
  `ErrProyectoInexistente` y no se registra nada.
- **CL-022-8** (RN-022-5) — Proyecto existente, historia inexistente o de otro proyecto,
  sprint inexistente y datos inválidos: se devuelve solo `ErrItemInexistente`,
  aunque la historia exista en otro proyecto.
- **CL-022-9** (RN-022-5) — Proyecto e historia correctos, sprint de detección inexistente
  o de otro proyecto, y datos inválidos: se devuelve solo
  `ErrSprintInexistente`, aunque el sprint exista en otro proyecto. Una
  historia del proyecto A con un sprint del proyecto B se rechaza igual.
- **CL-022-10** (RN-022-6) — Sprint de detección sin elegir (ID vacío o mal formado): se
  rechaza con `ErrSprintInexistente`. El sprint de detección es obligatorio y
  no hay forma de registrar un defecto sin sprint de detección.
- **CL-022-11** (RN-022-7) — Registrar sobre una historia `pendiente`, `en_progreso` o
  `completado`, y sobre una que no está en ningún sprint: se acepta en todos los
  casos y la historia no cambia (RN-022-7).
- **CL-022-12** (RN-022-6) — Historia asignada al Sprint 1 y defecto detectado en el
  Sprint 3: se acepta y se guarda el Sprint 3. Detectado en un sprint cerrado o
  en uno abierto: se acepta. Detectado en un sprint que todavía no empezó
  (fechas posteriores a hoy): se acepta (ver Huecos).
- **CL-022-13** (RN-022-8) — Dos defectos con la misma descripción, severidad, historia y
  sprint: se registran los dos, con ID distinto. Varios defectos sobre la misma
  historia: se registran todos.
- **CL-022-14** (RN-022-3) — Un defecto recién registrado, de cualquiera de las cuatro
  severidades (incluida `critica`), queda con estado `abierto` y sin sprint de
  resolución (`NULL`).
- **CL-022-15** (RN-022-10) — El proyecto, la historia o el sprint desaparecen entre la
  verificación y el `INSERT`: el repositorio devuelve el error de inexistencia
  correspondiente y no guarda nada. Se prueba en el test del repositorio.
- **CL-022-16** (RN-022-6) — Defecto detectado en el sprint "Sprint 3" del proyecto: se
  guarda el ID de ese sprint y, al consultar la tabla `sprints` con ese ID, su
  número es 3. Es lo que necesita US-029; se prueba en el test del repositorio.
- **CL-022-17** (RN-022-10) — Borrar un proyecto que tiene defectos: la base borra también
  sus defectos, sin error. Se prueba en el test del repositorio.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario | Origen |
|---|---|---|---|
| Descripción vacía o con solo espacios | `defecto.ErrDescripcionVacia` | "la descripcion es obligatoria" | Dominio |
| Descripción de más de 2000 caracteres después de recortarla | `defecto.ErrDescripcionMuyLarga`, envuelto indicando cuántos caracteres tiene | "la descripcion no puede tener mas de 2000 caracteres" | Dominio |
| Severidad distinta de `critica`, `alta`, `media` o `baja`, o vacía | `defecto.ErrSeveridadInvalida`, envuelto indicando el valor recibido | "la severidad tiene que ser critica, alta, media o baja" | Dominio |
| El proyecto no existe | `app.ErrProyectoInexistente`, envuelto indicando el ID recibido | "el proyecto no existe" | Aplicación |
| La historia no existe, no es de ese proyecto o no se eligió | `app.ErrItemInexistente`, envuelto indicando el ID recibido | "el item no existe" | Aplicación |
| El sprint de detección no existe, no es de ese proyecto o no se eligió | `app.ErrSprintInexistente`, envuelto indicando el ID recibido | "el sprint no existe" | Aplicación |

Los tres primeros son del dominio (`internal/domain/defecto`): se detectan
todos en una sola validación, antes de escribir en la base, y se devuelven
unidos con `errors.Join` en el orden de RN-022-4. Con varios errores, el
mensaje es la unión de los mensajes, uno por línea. Los tres últimos los
detecta la capa de aplicación **antes** de validar los datos, en el orden de la
tabla, y se devuelven de a uno: no se unen con los de validación.
`ErrProyectoInexistente` y `ErrItemInexistente` ya existen en `internal/app`
(`ErrItemInexistente` lo comparten US-006, US-009, US-010, US-013 y US-019);
`ErrSprintInexistente` también existe, en `internal/app/asignar_item.go`
(US-009), así que esta historia no define ningún error nuevo en `app`. Todos
son valores de paquete comparables con `errors.Is`, y en ningún caso se
registra nada.

## 8. Criterios de aceptación

- **CA-022-1** (normal) — Con un proyecto, una historia y un sprint
  existentes, registrar un defecto de severidad `alta` con la descripción
  "Falla al guardar" en esa historia, detectado en ese sprint, deja un defecto
  guardado con esos datos, con ID, estado `abierto` y sin sprint de resolución.
- **CA-022-2** (alternativo) — Se pueden registrar defectos de las cuatro
  severidades: `critica`, `alta`, `media` y `baja`.
- **CA-022-3** (alternativo) — El sprint de detección puede ser distinto del
  sprint de la historia, y puede estar cerrado: una historia del Sprint 1 con
  un defecto detectado en el Sprint 2 guarda el Sprint 2.
- **CA-022-4** (alternativo) — Se puede registrar un defecto sobre una
  historia completada o que no está en ningún sprint, y la historia no cambia.
- **CA-022-5** (alternativo) — Se pueden registrar varios defectos sobre la
  misma historia, incluso dos idénticos: todos quedan registrados.
- **CA-022-6** (límite) — Una descripción de 2000 caracteres se acepta y una
  de 2001 se rechaza.
- **CA-022-7** (error) — Descripción vacía o solo espacios → rechazo, no se
  registra nada.
- **CA-022-8** (error) — Severidad fuera de `critica`, `alta`, `media` y
  `baja` (por ejemplo `urgente`, `Critica` o vacía) → rechazo.
- **CA-022-9** (error) — Historia que no existe o que es de otro proyecto →
  rechazo.
- **CA-022-10** (error) — Sprint de detección que no existe, que es de otro
  proyecto o que no se eligió → rechazo.
- **CA-022-11** (error) — Proyecto inexistente → rechazo, no se registra
  nada.
- **CA-022-12** (error) — Descripción vacía y severidad inválida a la vez →
  se rechazan los dos errores juntos, en ese orden.

---

## Decisiones tomadas y descartadas

- **El modelo es el de la plantilla de defectos.** `02-defect.yml` dice de sí
  misma que sus campos son los del requerimiento 6 de la guía y "el modelo de
  datos de la app". Se respetan sus campos y su obligatoriedad: descripción,
  severidad, historia relacionada y sprint de detección son obligatorios; el
  sprint de resolución se completa "cuando el defecto se resuelve", así que
  queda vacío al registrar. No se agregan campos que la plantilla no tiene
  (título, quién lo reportó, pasos separados).
- **El campo "Identificador" (`DEF-NNN`) de la plantilla no entra.** Es el
  nombre del issue de GitHub, no uno de los seis campos del requerimiento 6, y
  el defecto de la aplicación se identifica por su UUID. Se descartó un número
  correlativo por proyecto como el de US-005 (RN-005-10): obligaría a asignarlo
  en el repositorio, con bloqueo y una restricción `UNIQUE`, para una historia
  que no lo pide (ver Huecos).
- **Severidad: cuatro valores cerrados, en minúscula y sin tilde.** Se guardan
  `critica`, `alta`, `media` y `baja`, como la prioridad (`must`, `should`, ...) y
  el rol de un integrante, que también son cadenas exactas validadas en el
  dominio y con un `CHECK` en la base. Se descartó guardar "Crítica" con tilde:
  los identificadores y los valores del dominio van sin acentos (ADR 0003) y el
  escenario usaría el mismo valor que el dominio, sin conversiones. También se
  descartó un número (1 a 4): `media` se entiende sin tabla de equivalencias.
  Un valor inválido devuelve `ErrSeveridadInvalida`, como `ErrPrioridadInvalida`.
- **Descripción de 1 a 2000 caracteres, recortada.** La plantilla pide qué se
  esperaba, qué pasó y cómo reproducirlo, que es más largo que un título (120)
  pero no ilimitado. Los 2000 son una propuesta (ver Huecos), contados en
  runas y recortados con `TrimSpace`, como la actividad de US-019. Se descartó
  "sin largo máximo" como en la descripción de US-005: allí es opcional, acá es
  el único texto del defecto y tiene que entrar en una fila de lista.
- **Estado inicial `abierto`, y la tabla ya incluye `estado` y
  `sprint_resolucion_id`.** La plantilla nombra el estado pero no lista sus
  valores, y los sprints ya usan `abierto` para lo que está en curso. Se crean
  las dos columnas ahora para que US-023 cambie el estado y complete el sprint
  de resolución con un `UPDATE`, sin otra migración (la alternativa, agregarlas
  después con `ALTER TABLE`, obliga a US-023 a tocar la tabla y a numerar otra
  migración). `estado` lleva `DEFAULT 'abierto'` y **no** un `CHECK` con lista
  de valores: los define US-023, y un `CHECK` con solo `abierto` habría que
  reemplazarlo con otra migración, justo lo que se quería evitar. A cambio, hasta
  entonces la base no protege los valores del estado: lo hace el dominio.
- **El sprint de detección es obligatorio, es cualquier sprint del proyecto y
  se guarda como ID.** Obligatorio porque la plantilla lo exige y porque US-029
  no tiene cómo contar un defecto sin sprint de detección. Cualquier sprint del
  proyecto, y no solo el de la historia, porque un defecto se suele detectar
  después, en una Review o en un sprint posterior, y esa diferencia es lo que
  mide la épica. Se descartó exigir que el sprint esté abierto (se registrarían
  tarde los defectos que aparecen al cerrar) y que ya haya empezado (haría falta
  el reloj y una regla que ninguna historia pide). Se guarda el ID con clave
  foránea y no el número: el número ya está en `sprints`, no se duplica, y
  US-029 lo obtiene con la tabla.
- **La pertenencia al proyecto la verifica el caso de uso, no la base.** Se
  descartó una clave foránea compuesta (`proyecto_id` con `item_id` y con
  `sprint_deteccion_id`): exigiría agregar una restricción `UNIQUE` a
  `items_backlog` y a `sprints` en esta migración. Como en US-019, el caso de
  uso verifica la pertenencia con una lectura por proyecto, y como ninguna
  historia mueve historias o sprints de un proyecto a otro, la verificación no
  se desactualiza entre la lectura y el `INSERT`.
- **Claves foráneas con `ON DELETE CASCADE`.** Igual que `registros_esfuerzo`.
  Se descartó `RESTRICT` en el sprint y la historia para proteger los defectos:
  ninguna historia borra un sprint o una historia sueltos, así que solo actuaría
  al borrar un proyecto, donde obligaría a borrar los defectos antes que sus
  sprints.
- **Se puede registrar sobre una historia en cualquier estado.** Como esfuerzo
  (RN-019-7): un defecto aparece, típicamente, sobre una historia ya completada.
  Se descartó limitarlo a historias completadas o asignadas a un sprint.
- **Orden de las verificaciones: proyecto, historia, sprint.** Igual que
  RN-019-6: el primero que falla se devuelve y no se une con los de
  validación. Se descartó unir todos los errores de existencia y de validación:
  no tiene sentido validar los datos de un defecto que no se puede ubicar en el
  proyecto, y es lo que ya hace esfuerzo.
- **Sin error nuevo para el sprint.** `app.ErrSprintInexistente` ya existe
  (US-009); además se descartó un `ErrSprintDeteccionFaltante` aparte: como la
  historia, un sprint sin elegir llega como un ID vacío, que es inexistente. La
  pantalla puede hacer obligatoria la lista; la regla vive en el caso de uso.
- **Sin unicidad.** No hay una clave natural de un defecto. Se descartó rechazar
  repetidos: pueden ser dos fallas parecidas, y el costo de uno duplicado es
  bajo y se resuelve con US-023.
- **Paquete `defecto`, no `defectos`.** `CONTRIBUTING.md` (§1) pide paquetes en
  singular y nombra `defecto` como ejemplo; los de dominio existentes (`sprint`,
  `integrante`, `proyecto`) siguen esa regla. La tabla sí va en plural
  (`defectos`), como `sprints` o `proyectos`. El nombre `defecto` no choca con
  ningún paquete, tipo o error existente.
- **Compatibilidad con US-029.** US-029 cuenta con una función pura que recibe
  `DefectoRegistrado{SprintDeteccion, SprintResolucion}` con los sprints por
  número, sin importar este paquete. El modelo de acá deja que la capa de
  aplicación arme esa lista: el sprint de detección siempre existe y es del
  proyecto (número `>= 1`), y el de resolución está vacío al registrar (se
  traduce a `0`, "sin resolver"). Qué estados cuentan como "resueltos" lo define
  US-023, no esta historia.
- **La pantalla queda fuera de alcance.** Esta historia entrega dominio, caso de
  uso, repositorios y migración, como US-019. El formulario depende del layout
  base (T-004).
- **Frases nuevas del diccionario** (aprobadas en el PR de implementación).
  Las ocho de la sección 5 pertenecen al área `Defectos`. La frase de contexto
  `un defecto de severidad ...` hace falta para los escenarios de varios
  defectos: el escenario tiene un solo `Cuando`.

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia:

1. **Estados del defecto y estado inicial.** La plantilla de defectos nombra el
   estado pero no lista sus valores. Se propone `abierto` como inicial. Si
   US-023 elige otro nombre, cambian el `DEFAULT` de la migración, `EstadoAbierto`
   y RN-022-3.
2. **Largo máximo de la descripción.** Los 2000 caracteres son una propuesta:
   ningún documento del repo fija un límite.
3. **Identificador legible (`DEF-NNN`).** La plantilla lo pide como texto del
   issue. Esta spec lo deja afuera; si la pantalla de defectos tiene que mostrar
   uno, hace falta un número correlativo por proyecto (columna, restricción
   `UNIQUE` y asignación en el repositorio, como en US-005).
4. **"Sprint 0".** La plantilla ofrece `Sprint 0` como sprint de detección,
   pero la aplicación solo tiene sprints con número `>= 1` (`CHECK (numero >
   0)`, y US-029 exige `>= 1`). Un defecto detectado antes del primer sprint
   del proyecto no se puede registrar con este modelo.
5. **Sprint de detección que todavía no empezó.** Se acepta (CL-022-12) porque
   nada lo impide hoy y rechazarlo necesita una regla con la fecha de hoy.
   Confirmar si el equipo quiere pedir que el sprint ya haya empezado.
6. **Quién registra el defecto.** La plantilla no tiene un campo "reportado por"
   y la aplicación no tiene usuarios (ver el mismo hueco de US-018), así que no
   se guarda.
7. **La frase "Como / quiero / para" y los criterios del issue.** El issue #30
   no trae la frase de la sección 1 (está redactada para esta spec) y sus
   criterios `CA-022-1` y `CA-022-2` están sin completar. Los criterios de la
   sección 8 son los propuestos para el issue.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia, la plantilla de defectos (`02-defect.yml`) y las specs de US-019
  y US-029 como modelo, incluida la frase "Como / quiero / para" de la sección 1;
  las decisiones sobre severidad, estado inicial, sprint de detección y
  persistencia se proponen acá y se confirman en la revisión
- Revisado y entendido por: Angelo Conforti
