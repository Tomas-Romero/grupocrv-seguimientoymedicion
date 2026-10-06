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
| Sprints | `Sprints` | US-008, US-009, US-010, US-011 | `features/steps/sprints.go` |
| Métricas | `Metricas` | US-025, US-026 | `features/steps/metricas.go` |

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
| `Y una historia "Crear proyecto" con prioridad "must"` | `^una historia "([^"]*)" con prioridad "([^"]*)"$` |
| `Y una historia "Crear proyecto" con prioridad "must" y 5 story points` | `^una historia "([^"]*)" con prioridad "([^"]*)" y (\d+) story points$` |
| `Y un sprint "Sprint 1" con el objetivo "MVP navegable"` | `^un sprint "([^"]*)" con el objetivo "([^"]*)"$` |
| `Y la historia "Crear proyecto" asignada al sprint "Sprint 1"` | `^la historia "([^"]*)" asignada al sprint "([^"]*)"$` |
| `Y la historia "Crear proyecto" completada` | `^la historia "([^"]*)" completada$` |
| `Y el sprint "Sprint 1" cerrado` | `^el sprint "([^"]*)" cerrado$` |
| `Dado ningún sprint cerrado` | `^ningún sprint cerrado$` |
| `Dado ningún proyecto` | `^ningún proyecto$` |

Un escenario **no repite títulos de historias**: el área Backlog las identifica
por título, y la spec de US-005 permite repetirlos entre sí, así que dos historias
con el mismo título en un mismo escenario serían ambiguas.

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
| `Cuando se registra al integrante "Vergara" con rol "product_builder"` | `^se registra al integrante "([^"]*)" con rol "([^"]*)"$` |
| `Cuando se crea la historia "Crear proyecto" con prioridad "must"` | `^se crea la historia "([^"]*)" con prioridad "([^"]*)"$` |
| `Cuando se crea la historia "Crear proyecto" con prioridad "must", la descripción "Alta" y los criterios:` + tabla de una columna | `^se crea la historia "([^"]*)" con prioridad "([^"]*)", la descripción "([^"]*)" y los criterios:$` |
| `Cuando se crea una historia con un título de 121 caracteres y prioridad "must"` | `^se crea una historia con un título de (\d+) caracteres y prioridad "([^"]*)"$` |
| `Cuando se cambia la prioridad de la historia "Crear proyecto" a "should"` | `^se cambia la prioridad de la historia "([^"]*)" a "([^"]*)"$` |
| `Cuando se estima la historia "Crear proyecto" en 5 story points` | `^se estima la historia "([^"]*)" en (\d+) story points$` |
| `Cuando se crea el sprint "Sprint 1" con el objetivo "MVP navegable"` | `^se crea el sprint "([^"]*)" con el objetivo "([^"]*)"$` |
| `Cuando se asigna la historia "Crear proyecto" al sprint "Sprint 1"` | `^se asigna la historia "([^"]*)" al sprint "([^"]*)"$` |
| `Cuando se marca la historia "Crear proyecto" como completada` | `^se marca la historia "([^"]*)" como completada$` |
| `Cuando se cierra el sprint "Sprint 1"` | `^se cierra el sprint "([^"]*)"$` |
| `Cuando se consultan los story points planificados del sprint "Sprint 1"` | `^se consultan los story points planificados del sprint "([^"]*)"$` |
| `Cuando se consultan los story points completados del sprint "Sprint 1"` | `^se consultan los story points completados del sprint "([^"]*)"$` |
| `Cuando se consulta la velocidad del equipo` | `^se consulta la velocidad del equipo$` |

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
| `Y los story points planificados son 13` | `^los story points planificados son (\d+)$` |
| `Y los story points completados son 8` | `^los story points completados son (\d+)$` |
| `Y la velocidad del equipo es 25` | `^la velocidad del equipo es (\d+(?:[.,]\d+)?)$` |

`la operación es exitosa` y `la operación se rechaza` miran el error que dejó el
último `Cuando`. Son los dos steps que más vas a usar: todo escenario de error
termina con `se rechaza` más `el mensaje de error indica`. Los mensajes de error
del código van sin tildes, así que el escenario los escribe igual: "el titulo es
obligatorio".

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
