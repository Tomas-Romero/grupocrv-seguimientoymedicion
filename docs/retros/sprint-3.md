# Retrospectiva — Sprint 3

| | |
|---|---|
| **Fecha** | 2026-10-19 |
| **Periodo** | 2026-10-13 al 2026-10-19 |
| **Participantes** | Tomas, Conforti, Vergara |
| **Facilita** | Tomas (Agile Enabler) |

## Metricas del sprint

| Metrica | Planificado | Real |
|---|---|---|
| Story Points | 21 comprometidos (techo 36) | 31 |
| Historias completadas | 5 | 7 historias y 2 tareas tecnicas: US-003, US-006, US-008, US-009, US-010, US-013 y US-019; T-004 y T-005 |
| Horas estimadas | 31,5 (21 SP a 1,5 h/SP) | — |
| Horas reales | — | sin registrar |
| Desviacion de esfuerzo | — | no calculable |
| Defectos detectados | — | 1 nuevo (DEF-006); sigue abierto DEF-004 |
| Defectos resueltos | — | 2 (DEF-005 y DEF-006); DEF-004 queda resuelto en codigo, ver abajo |
| Cobertura del dominio | >= 85% | 100% (`internal`: 95,1%) |

**Velocidad del sprint: 31 SP**, contra 16 SP del Sprint 2: el 148% de lo comprometido
y el 86% del techo de 36. No se activa el recorte: se completo mas del 75%.

**Sprint Goal:** el ciclo proyecto → backlog → sprint funciona en dominio y
aplicacion; layout base de la UI.

**Se cumplio:** si. Se puede crear un proyecto, registrar integrantes y esfuerzo,
cargar, editar y estimar items, crear un sprint, asignarle historias y marcarlas como
completadas, todo con TDD y BDD. T-004 deja el layout, la navegacion y el router, y T-005
el manejo uniforme de errores. Queda fuera el cierre del sprint (US-011), que era
objetivo estirado.

### Quien hizo que

| Item | SP | Responsable previsto | Lo implemento |
|---|---|---|---|
| US-003 Registrar integrantes | 3 | Conforti | Conforti |
| US-006 Editar y repriorizar items | 3 | Vergara | Vergara |
| US-008 Crear un sprint | 3 | Vergara | Tomas |
| US-009 Asignar historias a un sprint | 3 | Conforti | Tomas |
| US-010 Marcar historias como completadas | 2 | Vergara | Vergara |
| US-013 Estimar en Fibonacci | 2 | Tomas | Tomas |
| US-019 Registrar esfuerzo | 5 | Tomas | Tomas |
| T-004 Layout base y navegacion | 5 | Conforti | Conforti (herramientas, layout y navegacion) y Tomas (errores de HTMX, router y paginas) |
| T-005 Validaciones y errores uniformes | 5 | Tomas | Tomas |
| US-011 Cerrar un sprint | 5 | Vergara | pasa al Sprint 4 |

US-008 y US-009 las tomo Tomas el 17/10 y el 18/10 para destrabar la cadena
US-008 → US-009 → US-010: sin ellas, US-010 no podia empezar. T-004 se completo el
19/10 sobre la rama de Conforti, que habia dejado el layout y las herramientas pero no
llegaba al router ni a la spec.

### Trabajo que no estaba en el plan

- **Chequeo del catalogo de errores (T-005):** un test lee el codigo fuente y falla si
  un error exportado del dominio o de la aplicacion no esta registrado. En su primera
  semana encontro un error sin registrar en US-010 (`ErrItemFueraDelSprint`) antes de
  que llegara a `main`.
- **DEF-006 (#118, #119):** el comando documentado para correr un escenario con
  `go test ./features/... -godog.tags=...` falla; el correcto es `./features/`.
- **DEF-005 (#111):** requisitos de puesta en marcha incompletos.
- **`.mailmap` (#116):** une las identidades de Git de cada integrante.
  `git shortlog -sn` pasa a ser confiable.
- **DEF-004 (#55):** la causa se confirmo con un experimento (sin ningun `.templ`, templ
  nunca lanza el servidor) y quedo resuelta con el primer `.templ` de T-004. Su cierre
  formal es el PR de `scripts/verificar-dev.sh`, que falla contra el codigo anterior y
  pasa con T-004.
- 20 PRs mergeados desde el 13/10 (Tomas 12, Vergara 6, Conforti 2); el #127 de T-004
  cierra el sprint. En `main`, 36 commits `RED` y 33 `GREEN` conservan el ciclo.

## Acciones de la retro anterior

| Accion | Responsable | Estado |
|---|---|---|
| Desactivar *Squash and merge* y mergear siempre con *Rebase and merge*; revisar con `git log --grep RED` que cada historia conserve su ciclo en `main` | Tomas | cumplida: ningun PR del sprint se mergeo con squash y `git log --grep RED` muestra el ciclo de cada historia |
| A mitad de sprint (jueves) cada historia tiene su PR en borrador con el RED a la vista, y todo PR se revisa en menos de 24 horas; el que no pueda lo avisa en el grupo el mismo dia | Todos; controla Tomas | parcialmente: las revisiones fueron rapidas (#117 en 1 h 30, #126, #125 y #129 en cerca de una hora), pero el primer PR de una historia de implementacion llego el sabado 17/10 (#113); #114, #115, #117 y #122 el domingo 18/10; y #125, #126 y #127 el lunes 19/10. Ningun borrador el jueves |

> Si una accion aparece dos retros seguidas sin cumplirse, se descarta o se
> reformula. La del jueves en borrador se cumplio a medias dos veces seguidas
> (entregas pegadas al fin de semana): se reformula, ver abajo.

## Que funciono

- **Rebase and merge:** el ciclo RED → GREEN esta en el historial de `main` para las
  nueve piezas del sprint. Era lo que la catedra iba a mirar.
- **Specs primero:** las especificaciones de US-003, US-006, US-008, US-009, US-010,
  US-013 y US-019 estaban mergeadas antes de empezar el codigo, y T-005 y T-004 las
  tuvieron antes de mergear su implementacion. Hubo retrabajo menor de revision
  (por ejemplo, la correccion de RN-010-4), no por reglas ambiguas.
- **La revision cruzada encontro problemas reales:** identificadores de spec que
  chocaban entre tareas e historias (`RN-005-k` de T-005 contra US-005, y lo mismo con
  T-004 y US-004), un error sin registrar que habria roto `main`, y la regla de HTMX
  para que se vean los errores 4xx y 5xx.
- **Reasignar a tiempo:** pasar US-008 y US-009 a Tomas destrabo la cadena y permitio
  que US-010 llegara el mismo dia del cierre.
- **Convenciones compartidas:** errores unidos con `errors.Join`, posiciones desde 1,
  ID generado por la base, `FOR NO KEY UPDATE` y dominio en espanol se repitieron en las
  historias nuevas sin renegociarlas.

## Que no funciono

- **La entrega se concentro otra vez en los ultimos dias.** De los 31 SP, 29 llegaron a
  `main` entre el sabado 17/10 y el lunes 19/10, y T-005, US-010 y T-004 el ultimo dia.
  Eso dejo sin margen para corregir: el cambio de US-010 por el catalogo de errores y los
  identificadores de la spec de T-004 se arreglaron en las ultimas horas.
- **Un sprint con una sola persona sosteniendo la cadena.** Tomas implemento 18 de los
  31 SP (mas la mitad de T-004); Conforti 3 (mas la otra mitad de T-004) y Vergara 5
  (mas cuatro specs: #99, #102, #104 y #106). El reparto no es sostenible: R5 sigue sin resolverse.
- **Mas de un dia de espera en una revision.** #122 espero casi 30 horas para su primera
  revision, mientras otras se resolvieron en menos de una hora. Cuando la revision de una
  persona se cuelga, nadie la reasigna.
- **Mensajes de commit sin el ID de la historia.** Cuatro commits de `main`
  (`docs(bdd)` de US-008, US-019 y US-009, y el de `.mailmap`) no llevan `[ID]`. Con
  *Rebase and merge* entran asi al historial. Vergara lo detecto en la revision de #115.
- **Las horas reales siguen sin registrarse**, por tercera retro seguida. US-019 ya
  esta en el producto, pero todavia no hay pantalla para cargar horas (T-011), asi que
  no hay donde registrarlas.

## Que vamos a cambiar

Maximo dos acciones. Concretas, con responsable y con fecha. Una accion sin
responsable no es una accion.

| Accion | Responsable | Para cuando |
|---|---|---|
| Agregar a `pr-title.yml` un chequeo que falle si algun commit del PR (sin contar los merge) no lleva `[US-NNN]`, `[DEF-NNN]` o `[T-NNN]` en el asunto; corregir con *reword* los PRs abiertos antes | Tomas, revisa Vergara | 2026-10-22 |
| Cada historia del sprint tiene su PR en borrador el jueves a las 22:00, con el RED a la vista. Quien no tiene borrador ese dia pasa su historia a quien tenga capacidad, y Tomas decide la reasignacion en el momento; el mismo criterio vale para una revision sin responder en 24 horas | Todos; decide Tomas | 2026-10-22 (primer jueves del Sprint 4) |

## Riesgos nuevos o que cambiaron

| ID | Riesgo | Cambio |
|---|---|---|
| R1 | Sobrecompromiso: quedan 111 SP en el backlog (Sprint 4: 44, Sprint 5: 43, sin sprint: 19, mas US-011) y la velocidad medida es de 31 SP | Sigue critico. Aun con 31 SP por sprint, dos sprints no alcanzan: el recorte de alcance se decide en el Planning del Sprint 4 |
| R5 | Distribucion despareja del trabajo entre los tres | Sigue alto, ahora con numeros: Tomas 18 SP, Vergara 5, Conforti 3 (y T-004 compartida). La velocidad de 31 depende de que una persona haga trabajo de rescate |
| R9 | Inconsistencias entre specs escritas en paralelo | Baja: los choques de identificadores se encontraron en revision, antes de implementar |
| R10 | Dependencias entre US-008, US-009, US-010 y US-011 | Cerrado en parte: US-008, US-009 y US-010 estan en `main`; queda US-011 |
| R11 | Nuevo: la velocidad del Sprint 3 incluye trabajo de rescate y no es repetible | Para el Planning, comprometer sobre una velocidad entre 16 y 31 SP y no sobre 31 |

## Decisiones para el Planning del Sprint 4 (20/10)

- **US-011** (cerrar un sprint, 5 SP) pasa al Sprint 4. Con US-010 en `main`, Vergara
  puede arrancarla; tiene que registrar sus errores en el catalogo de
  `internal/adapters/http/catalogo.go`.
- **T-011** (pantallas de alta, edicion y listado, 8 SP) queda desbloqueada por T-004 y
  por T-005: el layout, el router, `Responder` y el manejo de errores ya existen.
- Los 111 SP restantes no entran en dos sprints: el Planning decide que sale del
  alcance, empezando por lo que no sostiene el flujo principal.
- Las specs del Sprint 4 se abren como PR a mas tardar el jueves 22/10.
