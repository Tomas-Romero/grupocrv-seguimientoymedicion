# Retrospectiva — Sprint 1

| | |
|---|---|
| **Fecha** | 2026-10-05 |
| **Periodo** | 2026-09-21 al 2026-10-05 |
| **Participantes** | Tomas, Conforti, Vergara |
| **Facilita** | Tomas (Agile Enabler) |

## Metricas del sprint

| Metrica | Planificado | Real |
|---|---|---|
| Story Points | 33 | 5 |
| Historias completadas | 10 items | 1 (T-006, una tarea tecnica; ninguna historia de usuario) |
| Horas estimadas | — | sin registrar |
| Horas reales | — | sin registrar |
| Desviacion de esfuerzo | — | no calculable |
| Defectos detectados | — | 5 (DEF-001 a DEF-005) |
| Defectos resueltos | — | 3 (DEF-001, DEF-002 y DEF-003); quedan abiertos DEF-004 y DEF-005 |
| Cobertura del dominio | >= 85% | sin paquetes de dominio en `main` todavia; `internal/` cerca del 89% en CI |

**Velocidad del sprint: 5 SP de 33 (15%).**

**Sprint Goal:** se puede crear un proyecto con sus integrantes, cargar items de
backlog con criterios de aceptacion, crear un sprint y asignarle historias, y el
dominio calcula los Story Points del sprint y la velocidad del equipo. El primer
escenario BDD corre automatizado de punta a punta.

**Se cumplio:** no. De los 10 items planificados solo se completo T-006, que deja
el primer escenario BDD corriendo de punta a punta (con un doble en memoria, sin
dominio real). Ninguna historia de negocio llego a `main`. El trabajo del sprint
se fue en lo que habilita a las historias: las especificaciones y la
infraestructura.

### Trabajo que no estaba en el plan

- Cuatro defectos de puesta en marcha detectados en la prueba en maquina limpia
  de Vergara (DEF-002 a DEF-005), y DEF-001 (el perfil `full` no construia la
  imagen), detectado el 27/09. DEF-001, DEF-002 y DEF-003 se corrigieron: DEF-002
  con el RED reproducido en CI, y el perfil `full` de Docker ahora arranca con la
  base migrada y con datos, verificado por un job nuevo ("Perfil full").
- ADR 0002 (integracion de BDD con godog) y ADR 0003 (idioma del dominio).
- Revision cruzada con `CODEOWNERS`: los tres son duenos de `specs/` y
  `features/`, y de `internal/domain/`, `.github/` y `scripts/` lo son dos.
- Cinco specs mergeadas: US-001, US-002, US-005, US-025 y US-026. Ademas, el
  dominio de US-025 (RED y GREEN) esta terminado en una rama local, pendiente de
  su escenario BDD.
- 18 PRs mergeados desde el 21/09 (Tomas 12, Vergara 4, Conforti 2).

## Acciones de la retro anterior

| Accion | Responsable | Estado |
|---|---|---|
| Correr `make check` antes de cada push y anotar las horas reales al cerrar cada issue | Todos | parcialmente: `make check` se uso y se cita en los PRs de codigo; las horas no se anotaron |
| Clonar el repo en una maquina limpia, correr `make up && make check` y reportar las fallas como DEF | Conforti y Vergara | Vergara: cumplida el 30/09, con 4 defectos reportados (cuatro dias despues de la fecha). Conforti: cumplida, informada en la reunion |

## Que funciono

- La revision cruzada fue real: Vergara reviso 9 PRs ajenos y varios cambios que
  pidio mejoraron el producto (por ejemplo, el flag `-godog.tags` de T-006 no
  funcionaba y salio a la luz en la revision, y RN-025-3 de la spec de US-025 se
  corrigio por una inconsistencia con US-005).
- La prueba en maquina limpia encontro defectos reales antes de que los viera un
  evaluador, y el mas grave (DEF-002) se arreglo con un test que falla antes del
  fix, ejecutado en CI.
- Las decisiones de convencion quedaron escritas antes de que existiera el codigo
  de dominio (ADR 0003, diccionario de frases de BDD, `CODEOWNERS`).
- Las specs que entraron son completas: los ocho apartados, los cuatro tipos de
  caso y decisiones justificadas.

## Que no funciono

- No se completo ninguna historia de negocio. Las specs, que tienen que estar
  mergeadas antes del codigo, recien entraron en los ultimos dias del sprint, asi
  que el codigo no pudo empezar a tiempo.
- Las Dailys se hicieron a veces, hablando en persona o por chat, sin regularidad
  y sin registro escrito: no quedo evidencia del avance y el atraso se hizo evidente
  recien en la ultima semana.
- Hubo informacion sobre el repositorio que circulo sin verificarse contra el
  codigo (por ejemplo, la existencia de archivos y de una plantilla corregida
  que todavia no estaban en `main`). Eso duplico trabajo (la integracion de
  godog estaba en dos lugares a la vez) y llevo a un PR de spec con la plantilla
  en blanco.
- La convencion de como cerrar issues desde un PR de spec no estaba definida. El
  check de CI exige `Closes #N`, y dos PRs de spec quedaron bloqueados hasta
  definir que una spec cierra su propio issue `[SPEC ...]` y no el de la
  historia.
- Specs escritas en paralelo salieron con decisiones distintas para lo mismo
  (generacion de IDs, un error o todos los errores, nombre del paquete). Se
  alinearon al revisarlas, pero fue retrabajo.
- Las horas reales no se registraron, asi que sigue sin haber base para calibrar
  la equivalencia de 1 SP = 1,5 h.
- La participacion fue despareja: al cierre del sprint, el historial de `main`
  tiene 42 commits de Tomas, 22 de Vergara y 2 de Conforti, porque buena parte
  del trabajo de Conforti estuvo en PRs abiertos hasta el ultimo dia.
- Los tres defectos de entorno en Windows (finales de linea, permisos de
  ejecucion, orden de argumentos de `psql`) siguen costando tiempo.

## Que vamos a cambiar

Maximo dos acciones. Concretas, con responsable y con fecha.

| Accion | Responsable | Para cuando |
|---|---|---|
| Cada historia del Sprint 2 empieza con su spec mergeada: las specs de las historias del sprint se abren como PR a mas tardar el jueves 08/10, y se revisan en menos de 24 horas | Cada duenio de historia escribe; quien tiene PRs pendientes de revision los atiende antes de empezar trabajo nuevo | 2026-10-08 |
| Daily asincrono todos los dias antes de las 22:00 (hecho / hoy / bloqueos) en el grupo, y las horas reales se anotan en un comentario del issue al cerrarlo | Todos; controla Tomas en cada Refinement | 2026-10-12 (Review del Sprint 2) |

## Riesgos nuevos o que cambiaron

| ID | Riesgo | Cambio |
|---|---|---|
| R1 | Sobrecompromiso: 155 SP en cuatro sprints con una velocidad medida de 5 | Sube a critico. Se aplica la lista de recorte en el Planning del 06/10, no a mitad de sprint |
| R3 | El evaluador no puede levantar el proyecto | Baja: el perfil `full` esta verificado en CI y la prueba en maquina limpia ya se hizo. Quedan abiertos DEF-004 y DEF-005 |
| R5 | Distribucion despareja del trabajo entre los tres | Sube: ver las cifras de commits de arriba. Se revisa `git shortlog -sn` en cada Review |
| R8 | Diferencias de entorno al trabajar en Windows | Baja: DEF-003 corregido (#73). Siguen abiertos DEF-004 y DEF-005 |
| R9 | Inconsistencias entre specs escritas en paralelo | Nuevo. Se mitiga con el ADR 0003, el diccionario de frases y revisar cada spec contra las ya mergeadas |

## Decision de calendario

Se mantiene el calendario ajustado el 05/10: el Sprint 1 dura dos semanas
(21/09 al 05/10) y los sprints siguientes cierran los lunes, con la presentacion
final el 02/11. Vergara propuso conservar el Sprint 1 y el Sprint 2 con sus
fechas originales, con velocidad cero y un acta cada uno, y numerar desde el
Sprint 3 a partir del 06/10. Se decidio no hacerlo.

**Motivo:** por motivos de tiempo, el equipo no llego a terminar las tareas
asignadas a cada persona dentro del sprint. Se prefirio cerrarlo el 05/10, con su
acta y su velocidad real, y mantener sprints semanales que cierran los lunes, en
lugar de renumerar los sprints ya transcurridos. Ademas, el lunes 12/10 es
feriado y los sprints siguientes quedan alineados con la presentacion del 02/11.

## Decisiones del Planning del Sprint 2 (06/10)

El detalle esta en `docs/plan-de-trabajo.md`. Lo decidido, con los datos de este
sprint:

- **Solo entra al sprint lo que tiene su spec mergeada** (Definition of Ready). El
  Sprint 2 queda en 16 SP: US-001, US-002, US-005, US-025 y US-026.
- **Recorte aplicado desde el principio**: US-017, US-024, US-032, T-007 y US-007
  salen del alcance comprometido (19 SP) y vuelven solo si sobra capacidad.
- **Item nuevo T-011** (8 SP): pantallas de alta, edicion y listado. Ninguna
  historia las cubria y sin ellas la aplicacion no se puede usar.
- **Reparto rebalanceado**: US-019, US-004 y US-012 pasan a Tomas, y US-033 a
  Vergara, para que la carga por sprint quede pareja.
- **Segundo recorte preparado**: si el Sprint 2 completa menos de 12 SP, el Planning
  del 13/10 saca US-031, US-021, US-020, US-012 y US-004.
- **Ceremonias**: Planning los martes, Refinement los jueves (con las specs del
  sprint siguiente ya abiertas como PR), Review y Retro los lunes a la noche, y
  Daily escrita en el grupo todos los dias.

