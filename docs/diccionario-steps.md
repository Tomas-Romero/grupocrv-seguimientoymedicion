# Diccionario de steps BDD

Catálogo de las frases que se pueden usar en los archivos `.feature`. **Si una
frase no está acá, no existe** — no la inventes en tu escenario: proponela, se
agrega a este archivo y se implementa el step en el mismo PR.

Este documento es el contrato entre quien escribe los escenarios y quien
implementa los steps. Un `.feature` que use una frase que no está implementada
hace fallar el job de BDD en CI, y con razón.

---

## Las seis reglas

### 1. Nunca escribas "que" después de "Dado"

En el dialecto español de Gherkin las palabras clave de contexto son `Dado`,
`Dada`, `Dados` y `Dadas`. **`que` no es parte de la palabra clave**, así que
`Dado que existe un proyecto "Demo"` le pasa al matcher el texto
`que existe un proyecto "Demo"`, con el `que` adentro. Si el step está definido
como `^existe un proyecto "([^"]*)"$` con ancla, no matchea y el escenario queda
indefinido.

```gherkin
# mal
Dado que existe un proyecto "Demo"

# bien
Dado un proyecto "Demo"
```

> Verificalo vos mismo la primera vez: escribí un escenario con `Dado que ...`,
> corré `make bdd` y mirá qué frase te reporta godog como *undefined*. Es el tipo
> de detalle que conviene confirmar en treinta segundos en vez de confiar en la
> documentación.

### 2. Todos los steps van con ancla

`^...$` en cada expresión. Sin ancla, `existe un proyecto` matchearía también
`no existe un proyecto`, que es exactamente el caso de error que queríamos
distinguir.

### 3. Tercera persona impersonal

`se crea un proyecto`, no `yo creo un proyecto` ni `crear proyecto`. Consistente
en todo el catálogo.

### 4. Los steps describen negocio, no mecánica

```gherkin
# mal — acopla el escenario a la implementación
Cuando se hace POST a "/proyectos" con el body {...}
Cuando se hace click en el boton "Guardar"

# bien — sobrevive a un cambio de UI o de API
Cuando se crea un proyecto "Demo" con fechas del "2026-09-14" al "2026-11-18"
```

### 5. Los valores variables van entre comillas dobles; los números, sin comillas

```gherkin
Dado una historia "Crear proyecto" con prioridad "must" y 5 story points
```

Las fechas siempre en formato `AAAA-MM-DD` y entre comillas.

Los valores de prioridad van en **minúscula** y son los que acepta el dominio:
`must`, `should`, `could` y `wont`. Así el escenario usa el mismo valor que el
dominio y los steps no convierten nada.

### 6. El `.feature`, los steps y el dominio entran en el mismo PR

El job de BDD corre en modo estricto: un step pendiente o indefinido es una
falla, no un aviso. Entonces no se mergea un escenario cuyo step todavía no está
implementado. Si necesitás una frase nueva, el PR trae las tres cosas: el
escenario, el step y el dominio que lo hace pasar.

---

## Cómo se conectan con la aplicación

Los steps no conocen la base de datos ni HTTP: hablan con **una interfaz por área**,
definidas en `features/steps/mundo.go`. Cada área se conecta en
`features/servicios_test.go` cuando se mergea la historia que la implementa.

| Área | Interfaz | Historias | Steps en |
|---|---|---|---|
| Proyectos | `Proyectos` | US-001, US-002, US-003 | `features/steps/proyectos.go` |
| Backlog | `Backlog` | US-005, US-006, US-013 | `features/steps/backlog.go` |
| Sprints | `Sprints` | US-008, US-009, US-010, US-011, US-012 | `features/steps/sprints.go` |
| Métricas | `Metricas` | US-025, US-026, US-027, US-028, US-029 | `features/steps/metricas.go` |
| Planning Poker | `PlanningPoker` | US-014 a US-018 | `features/steps/planningpoker.go` |
| Defectos | `Defectos` | US-022 | `features/steps/defectos.go` |

Un escenario solo necesita las áreas que usa. Un `.feature` de proyectos no
exige que existan backlog, sprints ni métricas, así que el CI no se pone rojo por
trabajo de otra persona. Si un escenario usa un área que todavía no está
conectada, falla con `area sin conectar a los steps: <área>`: es a propósito.

---

## Contexto — `Dado`

| Frase | Expresión |
|---|---|
| `Dado un proyecto "Demo"` | `^un proyecto "([^"]*)"$` |
| `Dado un proyecto "Demo" con fechas del "2026-09-14" al "2026-11-18"` | `^un proyecto "([^"]*)" con fechas del "([^"]*)" al "([^"]*)"$` |
| `Y el integrante "Conforti" con rol "product_builder"` | `^el integrante "([^"]*)" con rol "([^"]*)"$` |
| `Y el integrante "Conforti" con email "conforti@ejemplo.test" y rol "product_builder"` | `^el integrante "([^"]*)" con email "([^"]*)" y rol "([^"]*)"$` |
| `Y una historia "Crear proyecto" con prioridad "must"` | `^una historia "([^"]*)" con prioridad "([^"]*)"$` |
| `Y una historia "Crear proyecto" con prioridad "must" y 5 story points` | `^una historia "([^"]*)" con prioridad "([^"]*)" y (\d+) story points$` |
| `Y una historia "Crear proyecto" con prioridad "must", 5 story points y 2 criterios de aceptación` | `^una historia "([^"]*)" con prioridad "([^"]*)", (\d+) story points y (\d+) criterios? de aceptación$` |
| `Y un sprint "Sprint 1" con el objetivo "MVP navegable"` | `^un sprint "([^"]*)" con el objetivo "([^"]*)"$` |
| `Y un sprint "Sprint 1" con el objetivo "MVP navegable" del "2026-10-06" al "2026-10-19"` | `^un sprint "([^"]*)" con el objetivo "([^"]*)" del "([^"]*)" al "([^"]*)"$` |
| `Y la historia "Crear proyecto" asignada al sprint "Sprint 1"` | `^la historia "([^"]*)" asignada al sprint "([^"]*)"$` |
| `Y la historia "Crear proyecto" completada` | `^la historia "([^"]*)" completada$` |
| `Y el sprint "Sprint 1" cerrado` | `^el sprint "([^"]*)" cerrado$` |
| `Dado ningún sprint cerrado` | `^ningún sprint cerrado$` |
| `Dado ningún proyecto` | `^ningún proyecto$` |
| `Y el sprint "Sprint 1" con una historia de 5 story points completada` | `^el sprint "([^"]*)" con una historia de (-?\d+) story points completada$` |
| `Y el sprint "Sprint 1" con una historia de 3 story points sin completar` | `^el sprint "([^"]*)" con una historia de (-?\d+) story points sin completar$` |
| `Y el sprint cerrado "Sprint 1" con 20 story points completados de 25 planificados` | `^el sprint cerrado "([^"]*)" con (\d+) story points completados de (\d+) planificados$` |
| `Y una sesión de Planning Poker sobre la historia "Crear proyecto"` | `^una sesión de Planning Poker sobre la historia "([^"]*)"$` |
| `Y el integrante "Ana" ya votó 5 en la sesión de la historia "Crear proyecto"` | `^el integrante "([^"]*)" ya votó (-?\d+) en la sesión de la historia "([^"]*)"$` |
| `Y el sprint "Sprint 1" con 8 horas registradas` | `^el sprint "([^"]*)" con (-?\d+(?:[.,]\d+)?) horas registradas$` |
| `Y el integrante "Ana" reveló los votos de la sesión de la historia "Crear proyecto"` | `^el integrante "([^"]*)" reveló los votos de la sesión de la historia "([^"]*)"$` |
| `Y un defecto detectado en el sprint 2 y resuelto en el sprint 3` | `^un defecto detectado en el sprint (-?\d+) y resuelto en el sprint (-?\d+)$` |
| `Y un defecto detectado en el sprint 2 sin resolver` | `^un defecto detectado en el sprint (-?\d+) sin resolver$` |
| `Y el integrante "Ana" ya registró 3 horas de "Desarrollo" en la historia "Crear proyecto" el 2026-10-05` | `^el integrante "([^"]*)" ya registró (-?\d+(?:[.,]\d+)?) horas de "([^"]*)" en la historia "([^"]*)" el (\S+)$` |
| `Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"` | `^un defecto de severidad "([^"]*)" con la descripción "([^"]*)" en la historia "([^"]*)" detectado en el sprint "([^"]*)"$` |

Un escenario **no repite títulos de historias**: el área Backlog las identifica
por título, y la spec de US-005 permite repetirlos entre sí, así que dos historias
con el mismo título en un mismo escenario serían ambiguas.

**Propuesta US-003** (se aprueba en el PR): las frases con email son la variante
de las dos frases sin email. Se usan cuando el escenario necesita controlar el
email: uno repetido, el mismo email en dos proyectos o uno con formato inválido.
Las frases sin email arman uno a partir del nombre, con dominio `ejemplo.test`:
siempre es válido y es distinto para cada nombre distinto, así que el escenario
no tiene que ocuparse de él.

Las dos frases `el sprint "S" con una historia de N story points ...` arman el sprint de
los escenarios de métricas (US-025 y US-026) sin pasar por el backlog ni por los sprints
reales: cada línea suma una historia al sprint. Aceptan un número negativo a propósito,
porque es la única forma de escribir el escenario de error de US-025 (CA-025-4).

`el sprint "S" con N horas registradas` suma un registro de esfuerzo al sprint de un
escenario de métricas (US-027), sin pasar por US-019. Los story points del sprint salen de las
frases `el sprint "S" con una historia de N story points ...`. Acepta 0 y un número negativo a
propósito, porque es la única forma de escribir el escenario de error de CA-027-5. Los números
aceptan coma o punto decimal.

**Porcentaje de historias completadas (US-028):** `se consulta el porcentaje de historias
completadas del sprint "S"` y `el porcentaje de historias completadas es N` usan el contexto
de las frases `el sprint "S" con una historia de N story points ...` (las mismas de US-025):
cada línea es una historia del sprint, y se cuentan historias, no Story Points. N acepta coma
o punto decimal y se compara con una tolerancia de 0,001. Un sprint sin historias da 0 sin
error; un story point negativo se rechaza con el mensaje de US-025.

**Defectos por sprint (US-029):** `un defecto detectado en el sprint N y resuelto en el sprint M` y
`un defecto detectado en el sprint N sin resolver` arman los defectos del escenario sin pasar por el
registro de defectos (US-022): cada línea suma un defecto. Los sprints se escriben por su número, sin
comillas, y aceptan 0 y números negativos a propósito: es la única forma de escribir los escenarios
de error. `se consultan los defectos del sprint N` calcula con todos los defectos del contexto, y
`los defectos detectados son N` y `los defectos resueltos son N` comparan las dos cifras del sprint.
Un sprint sin defectos da 0 y 0, sin error.

`el sprint cerrado "S" con N story points completados de M planificados` arma un sprint
ya cerrado para los escenarios de la velocidad (US-026). No valida nada: "8 completados de
5 planificados" es justo el dato corrupto que tiene que rechazar el dominio. Un nombre
vacío (`el sprint cerrado "" ...`) sirve para probar el rechazo del sprint sin nombre.

El primer `Dado un proyecto "X"` usa fechas por defecto (del 2026-01-01 al
2026-12-31) y queda como **proyecto actual**: los steps que siguen y no nombran
proyecto operan sobre ese. Eso evita repetir el nombre del proyecto en cada
línea. Cada `Dado un proyecto "X"` nuevo cambia el proyecto actual.

`Dado ningún proyecto` deja como proyecto actual uno que no existe: las áreas no
le encuentran ID y llaman a la aplicación con uno inexistente, así que el rechazo
lo produce el caso de uso y no el step. Es para los escenarios de proyecto
inexistente (CA-005-8, CA-008-10); con esa frase el escenario no lleva
`Antecedentes` que creen un proyecto.

---

## Acción — `Cuando`

| Frase | Expresión |
|---|---|
| `Cuando se crea un proyecto "Demo" con fechas del "2026-09-14" al "2026-11-18"` | `^se crea un proyecto "([^"]*)" con fechas del "([^"]*)" al "([^"]*)"$` |
| `Cuando se modifica el nombre del proyecto "Demo" a "Demo v2"` | `^se modifica el nombre del proyecto "([^"]*)" a "([^"]*)"$` |
| `Cuando se modifican las fechas del proyecto "Demo" del "2026-10-01" al "2026-12-15"` | `^se modifican las fechas del proyecto "([^"]*)" del "([^"]*)" al "([^"]*)"$` |
| `Cuando se registra al integrante "Vergara" con rol "product_builder"` | `^se registra al integrante "([^"]*)" con rol "([^"]*)"$` |
| `Cuando se registra al integrante "Vergara" con email "vergara@ejemplo.test" y rol "product_builder"` | `^se registra al integrante "([^"]*)" con email "([^"]*)" y rol "([^"]*)"$` |
| `Cuando se crea la historia "Crear proyecto" con prioridad "must"` | `^se crea la historia "([^"]*)" con prioridad "([^"]*)"$` |
| `Cuando se crea la historia "Crear proyecto" con prioridad "must", la descripción "Alta" y los criterios:` + tabla de una columna | `^se crea la historia "([^"]*)" con prioridad "([^"]*)", la descripción "([^"]*)" y los criterios:$` |
| `Cuando se crea una historia con un título de 121 caracteres y prioridad "must"` | `^se crea una historia con un título de (\d+) caracteres y prioridad "([^"]*)"$` |
| `Cuando se edita la historia "Crear proyecto" con el título "Alta de proyecto", la prioridad "must", la descripción "Alta" y los criterios:` + tabla de una columna | `^se edita la historia "([^"]*)" con el título "([^"]*)", la prioridad "([^"]*)", la descripción "([^"]*)" y los criterios:$` |
| `Cuando se cambia la prioridad de la historia "Crear proyecto" a "should"` | `^se cambia la prioridad de la historia "([^"]*)" a "([^"]*)"$` |
| `Cuando se estima la historia "Crear proyecto" en 5 story points` | `^se estima la historia "([^"]*)" en (\d+) story points$` |
| `Cuando se crea el sprint "Sprint 1" con el objetivo "MVP navegable"` | `^se crea el sprint "([^"]*)" con el objetivo "([^"]*)"$` |
| `Cuando se crea el sprint "Sprint 1" con el objetivo "MVP navegable" del "2026-10-06" al "2026-10-19"` | `^se crea el sprint "([^"]*)" con el objetivo "([^"]*)" del "([^"]*)" al "([^"]*)"$` |
| `Cuando se asigna la historia "Crear proyecto" al sprint "Sprint 1"` | `^se asigna la historia "([^"]*)" al sprint "([^"]*)"$` |
| `Cuando se marca la historia "Crear proyecto" como completada` | `^se marca la historia "([^"]*)" como completada$` |
| `Cuando se desmarca la historia "Crear proyecto"` | `^se desmarca la historia "([^"]*)"$` |
| `Cuando se cierra el sprint "Sprint 1"` | `^se cierra el sprint "([^"]*)"$` |
| `Cuando se consultan los sprints anteriores` | `^se consultan los sprints anteriores$` |
| `Cuando se consultan los story points planificados del sprint "Sprint 1"` | `^se consultan los story points planificados del sprint "([^"]*)"$` |
| `Cuando se consultan los story points completados del sprint "Sprint 1"` | `^se consultan los story points completados del sprint "([^"]*)"$` |
| `Cuando se consulta la velocidad del equipo` | `^se consulta la velocidad del equipo$` |
| `Cuando se consulta el porcentaje de historias completadas del sprint "Sprint 1"` | `^se consulta el porcentaje de historias completadas del sprint "([^"]*)"$` |
| `Cuando se compara el esfuerzo del proyecto "Demo"` | `^se compara el esfuerzo del proyecto "([^"]*)"$` |
| `Cuando se consultan los defectos del sprint 3` | `^se consultan los defectos del sprint (-?\d+)$` |
| `Cuando se inicia una sesión de Planning Poker sobre la historia "Crear proyecto"` | `^se inicia una sesión de Planning Poker sobre la historia "([^"]*)"$` |
| `Cuando el integrante "Ana" vota 5 en la sesión de la historia "Crear proyecto"` | `^el integrante "([^"]*)" vota (-?\d+) en la sesión de la historia "([^"]*)"$` |
| `Cuando se consultan las horas del sprint "Sprint 1"` | `^se consultan las horas del sprint "([^"]*)"$` |
| `Cuando se registra un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"` | `^se registra un defecto de severidad "([^"]*)" con la descripción "([^"]*)" en la historia "([^"]*)" detectado en el sprint "([^"]*)"$` |
| `Cuando se registra un defecto de severidad "alta" con una descripción de 2001 caracteres en la historia "Crear proyecto" detectado en el sprint "Sprint 1"` | `^se registra un defecto de severidad "([^"]*)" con una descripción de (\d+) caracteres en la historia "([^"]*)" detectado en el sprint "([^"]*)"$` |
| `Cuando el integrante "Ana" revela los votos de la sesión de la historia "Crear proyecto"` | `^el integrante "([^"]*)" revela los votos de la sesión de la historia "([^"]*)"$` |
| `Cuando el integrante "Ana" registra la estimación acordada de 5 en la sesión de la historia "Crear proyecto"` | `^el integrante "([^"]*)" registra la estimación acordada de (-?\d+) en la sesión de la historia "([^"]*)"$` |
| `Cuando el integrante "Ana" registra 2,5 horas de "Desarrollo" en la historia "Crear proyecto" el 2026-10-05` | `^el integrante "([^"]*)" registra (-?\d+(?:[.,]\d+)?) horas de "([^"]*)" en la historia "([^"]*)" el (\S+)$` |
| `Cuando se consulta el esfuerzo del proyecto "Demo"` | `^se consulta el esfuerzo del proyecto "([^"]*)"(?: de la historia "([^"]*)")?(?: del integrante "([^"]*)")?(?: del sprint "([^"]*)")?$` |
| `Cuando se consulta el esfuerzo del proyecto "Demo" de la historia "Crear proyecto" del integrante "Ana" del sprint "Sprint 1"` | (la misma frase, con los tres filtros) |

**Propuesta US-003:** `se registra al integrante "X" con email "..." y rol "Y"`
es la variante con email de la frase anterior (ver la nota de Contexto). Un
nombre de 100 o 101 caracteres no tiene frase: se prueba con tests unitarios.

**Comparar el esfuerzo (US-021):** `se compara el esfuerzo del proyecto "P"` usa el caso de uso real sobre
la base del escenario, así que el contexto se arma con las frases de siempre: la historia con su
estimación (`una historia "H" con prioridad "must", 5 story points y 1 criterio de aceptación`),
asignada al sprint y `completada`, y las horas con `el integrante "X" ya registró N horas ...`. Los
Entonces por historia (`... en la comparación`) buscan la historia por título. Después de comparar,
las frases de US-027 (`las horas estimadas son N`, `las horas reales son N`, `la desviación de horas
es N`, `la desviación porcentual es N` y `la desviación porcentual no se puede calcular`) se refieren
al **total de las historias comparables**; `las horas sin comparar son N` es el esfuerzo de las
historias que aparecen y no entran en ese total. Una historia pendiente sin horas no aparece.

**Sprints (US-008):** el nombre entre comillas (`"Sprint 1"`) es el nombre
**esperado**: la persona no lo escribe, lo asigna la aplicación (`Sprint N`, con N el
número de sprint del proyecto). El step crea el sprint sin nombre y, si la creación
sale bien, verifica que el nombre generado coincida; si no coincide, el step falla (no
es un rechazo de la operación). Las frases sin fechas arman fechas válidas y
consecutivas: el primer sprint de un proyecto empieza el día de inicio del proyecto,
cada siguiente el día después de que termina el último sprint del escenario, y todos
duran 14 días. Una creación rechazada no cuenta como último sprint. Como el último
sprint tiene que estar cerrado para crear el siguiente (RN-008-17), un escenario que
necesite dos sprints cierra el primero antes (`Y el sprint "Sprint 1" cerrado`, que
llega con US-011). Un sprint que no entre en las fechas del proyecto tiene que usar la
frase con fechas.

**Esfuerzo (US-019):** `el integrante "X" registra N horas de "A" en la historia "H"
el AAAA-MM-DD` y su variante de contexto `ya registró` registran las horas del
integrante en el proyecto actual. Aceptan coma o punto decimal y un número negativo
(es la única forma de escribir CA-019-6). En los escenarios "hoy" es el
2026-10-06 (el reloj fijo): una fecha posterior se rechaza. El integrante y la
historia se buscan por nombre dentro del proyecto actual: uno que no esté ahí se
rechaza por la aplicación, no por los steps.

**Registrar un defecto (US-022):** `se registra un defecto de severidad "S" con la descripción "D" en la
historia "H" detectado en el sprint "Sprint N"` y su variante de contexto `un defecto de severidad ...`
(un escenario tiene un solo `Cuando`, así que los defectos previos se arman en el contexto) registran el
defecto en el proyecto actual con el caso de uso real. La historia y el sprint se buscan por nombre
dentro del proyecto actual (el sprint por su nombre `Sprint N`, que existe en todos los proyectos): uno
que no esté ahí, incluido uno de otro proyecto, se manda con un ID que no existe y lo rechaza la
aplicación, no los steps. La severidad se escribe con el valor del dominio (`critica`, `alta`, `media` o
`baja`, sin tilde y en minúscula) y se manda tal cual, así que `urgente`, `Critica` o vacía son la forma de
escribir los escenarios de error. `con una descripción de N caracteres` arma la descripción con N letras
`ñ`, para que el escenario también pruebe que el límite cuenta caracteres y no bytes; una descripción de
solo espacios no tiene frase y la cubren los tests del dominio. Los Entonces de consulta
(`el defecto de la historia "H" tiene severidad ...`, `tiene estado ...`, `fue detectado en el sprint ...`
y `está sin resolver`) leen el último defecto registrado de esa historia.

**Consultar el esfuerzo (US-020):** `se consulta el esfuerzo del proyecto "P"` acepta tres
filtros opcionales que van en este orden: `de la historia "H"`, `del integrante "I"` y
`del sprint "S"`. Los nombres se buscan dentro del proyecto consultado: uno que no esté ahí
(por ejemplo, de otro proyecto) se manda con un ID que no existe y lo rechaza la aplicación. El
sprint se nombra como lo conoce el escenario (`Sprint N`). Las frases `las horas ... en la
consulta son N` comprueban los totales de lo que quedó después de filtrar; un nombre que no
aparece en la consulta vale 0 horas. Como "hoy" es el 2026-10-06, el esfuerzo de los
escenarios se registra antes de esa fecha y los sprints se crean con la frase con fechas.

**Marcar y desmarcar (US-010):** `la historia "X" completada`, `se marca la historia "X"
como completada` y `se desmarca la historia "X"` operan sobre el último sprint del
proyecto de la historia, que es el único que puede estar abierto (RN-008-17). Una
historia que no está asignada a ese sprint se rechaza. El estado se verifica con
`la historia "X" tiene estado "..."`; la fecha de completado no, porque el reloj de los
escenarios es fijo.

**Historia lista para un sprint (US-009):** `una historia "X" con prioridad "must", 5
story points y 2 criterios de aceptación` crea una historia estimada y con esa cantidad
de criterios, que son genéricos ("Criterio 1", "Criterio 2", ...). Es la que cumple la
Definition of Ready, y la que necesita cualquier escenario que asigne una historia a un
sprint con éxito. Con `una historia "X" con prioridad "must" y 5 story points` la
historia está estimada pero no tiene criterios, así que la asignación se rechaza.

**Un solo `Cuando` por escenario.** Si necesitás dos acciones, la primera es
contexto y va en un `Dado`. Un escenario con dos `Cuando` casi siempre está
probando dos cosas y conviene partirlo.

Los criterios de aceptación van en una tabla de una sola columna, un criterio por
fila y en orden:

```gherkin
Cuando se crea la historia "Alta" con prioridad "must", la descripción "Alta de proyectos" y los criterios:
  | Se guarda el proyecto |
  |                       |
```

Gherkin recorta los espacios de cada celda, así que una celda con solo espacios
llega vacía: un criterio "solo espacios" se prueba con tests unitarios, no acá.

`se edita la historia "X" con el título "Y", ...` manda los cuatro datos a la
vez, como el formulario de edición (US-006): los que no cambian se repiten. Desde
ahí la historia se nombra por su título nuevo. `se cambia la prioridad ...` es la
misma edición con los demás datos sin cambios.

`se crea una historia con un título de N caracteres` arma el título con N letras
`ñ`. Cada una ocupa dos bytes, así que el escenario también prueba que el límite
cuenta caracteres y no bytes.

---

## Resultado — `Entonces`

| Frase | Expresión |
|---|---|
| `Entonces la operación es exitosa` | `^la operación es exitosa$` |
| `Entonces la operación se rechaza` | `^la operación se rechaza$` |
| `Y el mensaje de error indica "fecha de fin anterior a la de inicio"` | `^el mensaje de error indica "([^"]*)"$` |
| `Y el proyecto "Demo" existe con fechas del "2026-09-14" al "2026-11-18"` | `^el proyecto "([^"]*)" existe con fechas del "([^"]*)" al "([^"]*)"$` |
| `Y el proyecto "Demo" tiene 3 integrantes` | `^el proyecto "([^"]*)" tiene (\d+) integrantes?$` |
| `Y el backlog del proyecto "Demo" tiene 2 historias` | `^el backlog del proyecto "([^"]*)" tiene (\d+) historias?$` |
| `Y la historia "Crear proyecto" tiene prioridad "should"` | `^la historia "([^"]*)" tiene prioridad "([^"]*)"$` |
| `Y la historia "Crear proyecto" tiene 5 story points` | `^la historia "([^"]*)" tiene (\d+) story points$` |
| `Y la historia "Crear proyecto" está sin estimar` | `^la historia "([^"]*)" está sin estimar$` |
| `Y la historia "Crear proyecto" tiene el número 2` | `^la historia "([^"]*)" tiene el número (\d+)$` |
| `Y la historia "Crear proyecto" tiene estado "pendiente"` | `^la historia "([^"]*)" tiene estado "([^"]*)"$` |
| `Y la historia "Crear proyecto" tiene los criterios:` + tabla de una columna | `^la historia "([^"]*)" tiene los criterios:$` |
| `Y la historia "Crear proyecto" está en el sprint "Sprint 1"` | `^la historia "([^"]*)" está en el sprint "([^"]*)"$` |
| `Y la historia "Crear proyecto" está en el backlog` | `^la historia "([^"]*)" está en el backlog$` |
| `Y el sprint "Sprint 1" está cerrado` | `^el sprint "([^"]*)" está cerrado$` |
| `Y el sprint "Sprint 1" está abierto` | `^el sprint "([^"]*)" está abierto$` |
| `Y los sprints anteriores son:` + tabla `sprint \| planificados \| completados \| historias completadas \| historias no completadas` | `^los sprints anteriores son:$` |
| `Y no hay sprints anteriores` | `^no hay sprints anteriores$` |
| `Y la velocidad de los sprints anteriores es 6,5` | `^la velocidad de los sprints anteriores es (\d+(?:[.,]\d+)?)$` |
| `Y los story points planificados son 13` | `^los story points planificados son (\d+)$` |
| `Y los story points completados son 8` | `^los story points completados son (\d+)$` |
| `Y la velocidad del equipo es 25` | `^la velocidad del equipo es (\d+(?:[.,]\d+)?)$` |
| `Y el porcentaje de historias completadas es 66,7` | `^el porcentaje de historias completadas es (\d+(?:[.,]\d+)?)$` |
| `Y la comparación tiene 2 historias` | `^la comparación tiene (\d+) historias?$` |
| `Y la primera historia de la comparación es "Crear proyecto"` | `^la primera historia de la comparación es "([^"]*)"$` |
| `Y la historia "Crear proyecto" tiene 7,5 horas estimadas y 8 reales en la comparación` | `^la historia "([^"]*)" tiene (-?\d+(?:[.,]\d+)?) horas estimadas y (-?\d+(?:[.,]\d+)?) reales en la comparación$` |
| `Y la desviación de la historia "Crear proyecto" en la comparación es -3` | `^la desviación de la historia "([^"]*)" en la comparación es (-?\d+(?:[.,]\d+)?)$` |
| `Y la desviación porcentual de la historia "Crear proyecto" en la comparación es 6,7` | `^la desviación porcentual de la historia "([^"]*)" en la comparación es (-?\d+(?:[.,]\d+)?)$` |
| `Y la desviación porcentual de la historia "Crear proyecto" en la comparación no se puede calcular` | `^la desviación porcentual de la historia "([^"]*)" en la comparación no se puede calcular$` |
| `Y el total comparable tiene 2 historias` | `^el total comparable tiene (\d+) historias?$` |
| `Y las horas sin comparar son 4` | `^las horas sin comparar son (\d+(?:[.,]\d+)?)$` |
| `Y los defectos detectados son 3` | `^los defectos detectados son (\d+)$` |
| `Y los defectos resueltos son 2` | `^los defectos resueltos son (\d+)$` |
| `Y las horas estimadas son 7,5` | `^las horas estimadas son (-?\d+(?:[.,]\d+)?)$` |
| `Y las horas reales son 8` | `^las horas reales son (-?\d+(?:[.,]\d+)?)$` |
| `Y la desviación de horas es -3` | `^la desviación de horas es (-?\d+(?:[.,]\d+)?)$` |
| `Y la desviación porcentual es 6,7` | `^la desviación porcentual es (-?\d+(?:[.,]\d+)?)$` |
| `Y la desviación porcentual no se puede calcular` | `^la desviación porcentual no se puede calcular$` |
| `Y el proyecto "Demo" tiene 1 registro de esfuerzo` | `^el proyecto "([^"]*)" tiene (\d+) registros? de esfuerzo$` |
| `Y la historia "Crear proyecto" tiene 2,5 horas registradas` | `^la historia "([^"]*)" tiene (\d+(?:[.,]\d+)?) horas registradas$` |
| `Y la consulta devuelve 4 registros` | `^la consulta devuelve (\d+) registros?$` |
| `Y la consulta devuelve primero el registro de "Pruebas"` | `^la consulta devuelve primero el registro de "([^"]*)"$` |
| `Y el total de horas de la consulta es 7,5` | `^el total de horas de la consulta es (\d+(?:[.,]\d+)?)$` |
| `Y las horas de la historia "Crear proyecto" en la consulta son 2` | `^las horas de la historia "([^"]*)" en la consulta son (\d+(?:[.,]\d+)?)$` |
| `Y las horas del integrante "Ana" en la consulta son 3,5` | `^las horas del integrante "([^"]*)" en la consulta son (\d+(?:[.,]\d+)?)$` |
| `Y las horas del sprint "Sprint 1" en la consulta son 5` | `^las horas del sprint "([^"]*)" en la consulta son (\d+(?:[.,]\d+)?)$` |
| `Y las horas fuera de todo sprint en la consulta son 1` | `^las horas fuera de todo sprint en la consulta son (\d+(?:[.,]\d+)?)$` |
| `Y el proyecto "Demo" tiene 2 defectos` | `^el proyecto "([^"]*)" tiene (\d+) defectos?$` |
| `Y el defecto de la historia "Crear proyecto" tiene severidad "alta"` | `^el defecto de la historia "([^"]*)" tiene severidad "([^"]*)"$` |
| `Y el defecto de la historia "Crear proyecto" tiene estado "abierto"` | `^el defecto de la historia "([^"]*)" tiene estado "([^"]*)"$` |
| `Y el defecto de la historia "Crear proyecto" fue detectado en el sprint "Sprint 1"` | `^el defecto de la historia "([^"]*)" fue detectado en el sprint "([^"]*)"$` |
| `Y el defecto de la historia "Crear proyecto" está sin resolver` | `^el defecto de la historia "([^"]*)" está sin resolver$` |
| `Y la sesión de la historia "Crear proyecto" está "votando" en la ronda 1` | `^la sesión de la historia "([^"]*)" está "([^"]*)" en la ronda (\d+)$` |
| `Y los participantes de la sesión de la historia "Crear proyecto" son:` + tabla de nombres | `^los participantes de la sesión de la historia "([^"]*)" son:$` |
| `Y en la sesión de la historia "Crear proyecto" votaron:` + tabla de nombres | `^en la sesión de la historia "([^"]*)" votaron:$` |
| `Y en la sesión de la historia "Crear proyecto" faltan votar:` + tabla de nombres | `^en la sesión de la historia "([^"]*)" faltan votar:$` |
| `Y los votos de la sesión de la historia "Crear proyecto" están ocultos` | `^los votos de la sesión de la historia "([^"]*)" están ocultos$` |
| `Y los votos de la sesión de la historia "Crear proyecto" son:` + tabla `integrante \| voto` | `^los votos de la sesión de la historia "([^"]*)" son:$` |
| `Y hay divergencia en la sesión de la historia "Crear proyecto"` | `^hay divergencia en la sesión de la historia "([^"]*)"$` |
| `Y no hay divergencia en la sesión de la historia "Crear proyecto"` | `^no hay divergencia en la sesión de la historia "([^"]*)"$` |
| `Y la estimación acordada de la sesión de la historia "Crear proyecto" es 5` | `^la estimación acordada de la sesión de la historia "([^"]*)" es (\d+)$` |

`la operación es exitosa` y `la operación se rechaza` miran el error que dejó el
último `Cuando`. Son los dos steps que más vas a usar: todo escenario de error
termina con `se rechaza` más `el mensaje de error indica`. Los mensajes de error
del código van sin tildes, así que el escenario los escribe igual: "el titulo es
obligatorio".

**Sprints anteriores (US-012):** `se consultan los sprints anteriores` pide el historial del
proyecto actual: solo los sprints cerrados, del más reciente al más antiguo. `los sprints
anteriores son:` compara la lista completa y en orden con una tabla de cinco columnas (la primera
fila es el encabezado), con las historias contadas por cantidad y los Story Points por suma; un
sprint abierto no aparece. La velocidad es la de esos mismos sprints y, igual que
`la velocidad del equipo es`, acepta coma o punto decimal. El contexto sale de las frases que ya
existen: crear el sprint, asignar y completar historias, y `el sprint "X" cerrado`.

**Planning Poker (US-014):** la sesión se nombra por su historia, que tiene como
máximo una sesión abierta. `una sesión de Planning Poker sobre la historia "X"` y
`se inicia una sesión ...` la inician en el proyecto actual, con todos sus integrantes
como participantes. Los participantes se comparan por nombre y en orden de alta en el
proyecto. Los estados se escriben como los define el dominio: `votando`, `revelada`,
`finalizada`.

**Votos (US-015):** `el integrante "X" vota N` y su variante de contexto `ya votó` votan en
la sesión abierta de la historia. El integrante se busca por nombre en el proyecto de la
historia; uno que no esté ahí se rechaza por la aplicación. El número acepta negativos
para poder escribir el rechazo de un valor fuera de la escala. Con la sesión votando, los
escenarios ven quién votó y quién falta, pero no los valores: `los votos ... están
ocultos` comprueba justamente eso. Una tabla de Gherkin no puede estar vacía, así que
"no falta nadie" no tiene frase.

**Revelar (US-016):** `el integrante "X" revela los votos ...` cierra la votación de la ronda
de la sesión abierta de la historia, y su variante de contexto `reveló` sirve para armar una
sesión ya revelada. Quien revela se busca por nombre, igual que quien vota. Después de
revelar, `los votos ... son:` compara los votos con una tabla de dos columnas
(`integrante | voto`, la primera fila es el encabezado) en el orden de los participantes, y
solo incluye a quienes votaron: los que faltan se comprueban con `faltan votar`. Con la
sesión todavía votando (por ejemplo, después de un revelado rechazado), pedir los votos
falla: es la misma garantía de `están ocultos`. `hay divergencia` y `no hay divergencia`
leen lo que devolvió la última revelación de esa historia.

**Estimación acordada (US-018):** `el integrante "X" registra la estimación acordada de N`
finaliza la sesión de la historia con ese valor y la deja estimada. Una sesión finalizada
ya no está abierta, así que las frases de la sesión (estado, participantes, votos y
estimación acordada) leen la **última** sesión de la historia. El número acepta negativos
para poder escribir el rechazo de un valor fuera de la escala.

`está sin estimar` y `tiene N story points` no son lo mismo: una historia recién
creada está sin estimar, no tiene 0 story points (RN-005-7). Los estados se
escriben como los define el dominio, en minúscula: `pendiente`, `en_progreso`,
`completado`.

---

## Un escenario completo, de ejemplo

Así se ve `features/US-026-velocidad-equipo.feature` usando solo frases del
catálogo:

```gherkin
# language: es
@US-026
Característica: Velocidad del equipo
  Como Agile Enabler
  quiero conocer la velocidad del equipo
  para estimar cuánto trabajo cabe en el próximo sprint

  Especificación: specs/US-026-velocidad-equipo.md

  Antecedentes:
    Dado un proyecto "Demo"

  @CA-026-1
  Escenario: La velocidad es el promedio de los sprints cerrados
    Dado un sprint "Sprint 1" con el objetivo "MVP"
    Y una historia "H1" con prioridad "must" y 20 story points
    Y la historia "H1" asignada al sprint "Sprint 1"
    Y la historia "H1" completada
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Interfaz"
    Y una historia "H2" con prioridad "must" y 30 story points
    Y la historia "H2" asignada al sprint "Sprint 2"
    Y la historia "H2" completada
    Y el sprint "Sprint 2" cerrado
    Cuando se consulta la velocidad del equipo
    Entonces la velocidad del equipo es 25

  @CA-026-2 @limite
  Escenario: Sin sprints cerrados la velocidad es cero, no un error
    Dado ningún sprint cerrado
    Cuando se consulta la velocidad del equipo
    Entonces la operación es exitosa
    Y la velocidad del equipo es 0

  @CA-026-3 @limite
  Escenario: Los sprints abiertos no cuentan
    Dado un sprint "Sprint 1" con el objetivo "MVP"
    Y una historia "H1" con prioridad "must" y 20 story points
    Y la historia "H1" asignada al sprint "Sprint 1"
    Y la historia "H1" completada
    Cuando se consulta la velocidad del equipo
    Entonces la velocidad del equipo es 0
```

Este ejemplo usa las cuatro áreas: no corre hasta que estén conectadas.

Fijate que los `Antecedentes` sacan afuera lo que comparten todos los
escenarios, y que cada escenario lleva la etiqueta del criterio de aceptación
que verifica. Esa etiqueta es el eslabón de la trazabilidad entre la spec y el
test.

---

## Cómo agregar una frase nueva

1. Proponela en el grupo antes de escribirla en tu `.feature`. Dos frases
   distintas para la misma idea es el problema que este documento evita.
2. Revisá si el catálogo ya cubre el caso con otras palabras. La mayoría de las
   veces sí: lo que parece una frase nueva es un `Dado` existente más un
   `Entonces` existente.
3. Si hace falta de verdad: agregala a la tabla de arriba, implementá el step en
   el archivo de su área dentro de `features/steps/` (`proyectos.go`, `backlog.go`,
   `sprints.go` o `metricas.go`) y, si necesita algo nuevo de la aplicación,
   agregá el método a la interfaz de esa área en `features/steps/mundo.go`.
4. Todo eso en el mismo PR que el escenario que la usa.
