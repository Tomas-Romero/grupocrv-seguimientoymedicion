# Plan de trabajo — Sprint 1 al Sprint 5

Este documento dice **quién hace qué y cuándo** hasta la entrega. Complementa a
`CONTRIBUTING.md`, que dice **cómo** se trabaja. Si los dos se contradicen,
manda `CONTRIBUTING.md` y se corrige este archivo.

Las fechas son estimativas; lo que no es negociable es que **haya avance
verificable todas las semanas**: un incremento demostrable, un acta de
retrospectiva y el tablero al día. El reparto y los objetivos se revisan en cada
Sprint Planning y cualquier cambio se anota en el acta del sprint.

| Sprint | Fechas | Sprint Goal (resumen) | SP | Estado |
|---|---|---|---|---|
| 0 | 14/09 – 20/09 | Repo, CI, tablero y walking skeleton | 16 | cerrado |
| 1 | 21/09 – 05/10 | El MVP: proyectos, backlog, sprints y primeras métricas | 5 de 33 | cerrado, ver `docs/retros/sprint-1.md` |
| 2 | 06/10 – 12/10 | Las cinco historias con spec mergeada, implementadas con TDD y BDD | 16 de 16 | cerrado, ver `docs/retros/sprint-2.md` |
| 3 | 13/10 – 19/10 | El ciclo proyecto → backlog → sprint funciona en dominio y aplicación; layout base de la UI | 31 de 21 comprometidos (techo 36) | cerrado, ver `docs/retros/sprint-3.md` |
| 4 | 20/10 – 26/10 | El ciclo completo se usa desde el navegador, base de Planning Poker y métricas de esfuerzo y completitud | 35 comprometidos, techo 46 | en curso, ver "Compromiso del Sprint 4" |
| 5 | 27/10 – 02/11 | Cierre: dashboard, reporte PDF, documentación y ensayo de la presentación | 43 | |
| — | 02/11 | **Presentación final** | — | |

Los sprints cierran los **lunes a la noche** con Review y Retrospective, y el
siguiente arranca el martes. El Sprint 1 duró dos semanas (21/09 – 05/10).

## Estado al 05/10 y plan decidido en el Planning del 06/10

El Sprint 1 cerró con **5 de 33 SP**: ninguna historia de negocio llegó a `main`,
pero quedaron mergeadas las especificaciones de US-001, US-002, US-005, US-025 y
US-026, y la infraestructura (godog, perfil `full`, `CODEOWNERS`, ADR 0002 y 0003).
Quedaban **150 SP** para cuatro sprints, y el equipo midió una velocidad de 5.
El plan de abajo es lo que se decidió con esos datos.

### Reglas desde ahora

1. **Definition of Ready estricta.** Solo entra a un sprint una historia con su
   spec **mergeada** en `main`. El Sprint 2 se arma con eso y nada más.
2. **Las specs del sprint siguiente se abren como PR antes del jueves** (Refinement)
   y se revisan en menos de 24 horas.
3. **El plan por sprint es el techo, no el compromiso.** Cada Planning fija el
   compromiso con la velocidad medida: si el sprint anterior completó menos del
   75 % de lo comprometido, se recorta el siguiente antes de empezar.
4. **Una historia entrega dominio, caso de uso, repositorio y BDD.** Las pantallas
   quedan fuera de las historias (las specs lo dicen) y se construyen en T-011,
   sobre el layout de T-004.

### Recorte aplicado

Salen del alcance comprometido (sin sprint; vuelven solo si sobra capacidad):
US-017 (3), US-024 (3), US-032 (5), T-007 (5, la matriz de trazabilidad se arma a
mano en `docs/trazabilidad.md`) y US-007 (3, el listado entra en T-011). Son
**19 SP**.

### Item nuevo

**T-011 — Pantallas de alta, edición y listado de proyecto, backlog y sprint** (8 SP,
Conforti, Sprint 4). Ninguna historia cubría las pantallas, y sin ellas la
aplicación no se puede usar.

### Segundo recorte, preparado de antemano

Si el Sprint 2 completa **menos de 12 SP**, el Planning del 13/10 saca, en este
orden: US-031 (5), US-021 (3), US-020 (3), US-012 (2) y US-004 (2). Son **15 SP**.

### Plan por sprint (139 SP)

**Sprint 2 — 06/10 a 12/10 — 16 SP.** Las cinco historias con spec mergeada.

| ID | Historia | SP | Responsable |
|---|---|---|---|
| US-001 | Crear un proyecto | 3 | Conforti |
| US-002 | Modificar un proyecto | 2 | Conforti |
| US-005 | Crear un ítem de backlog | 5 | Vergara |
| US-025 | Story Points planificados y completados | 3 | Tomás |
| US-026 | Velocidad del equipo | 3 | Tomás |

Además, durante el sprint se escriben las specs del Sprint 3. Primera tanda, en PR
antes del **jueves 08/10**: US-003 (Conforti), US-008 (Vergara), US-013 y US-019
(Tomás). Segunda tanda, antes del **lunes 12/10**: US-006, US-010 y US-011
(Vergara) y US-009 (Conforti).

**Sprint 3 — 13/10 a 19/10 — 36 SP.** Carga: Tomás 12, Vergara 13, Conforti 11.

| ID | Historia | SP | Responsable |
|---|---|---|---|
| US-003 | Registrar integrantes | 3 | Conforti |
| US-009 | Asignar historias a un sprint | 3 | Conforti |
| T-004 | Layout base y navegación | 5 | Conforti |
| US-006 | Editar y repriorizar ítems | 3 | Vergara |
| US-008 | Crear un sprint | 3 | Vergara |
| US-010 | Marcar historias como completadas | 2 | Vergara |
| US-011 | Cerrar un sprint | 5 | Vergara |
| US-013 | Estimar en Fibonacci | 2 | Tomás |
| US-019 | Registrar esfuerzo | 5 | Tomás |
| T-005 | Validaciones y errores uniformes | 5 | Tomás |

### Compromiso del Sprint 3 (Planning del 13/10)

El techo es de 36 SP, pero la velocidad medida es de 16 (Sprint 2). El
compromiso se fija en **21 SP**: lo que tiene su spec mergeada, y la cadena que
sostiene el Sprint Goal. Lo demás es objetivo estirado y se toma solo si hay
capacidad.

| Persona | Compromiso | SP | Objetivo estirado | SP |
|---|---|---|---|---|
| Conforti | US-003, US-009 | 6 | T-004 | 5 |
| Vergara | US-008, US-006, US-010 | 8 | US-011 | 5 |
| Tomás | US-013 (ya implementada), US-019 | 7 | T-005 | 5 |

**Orden de la cadena.** US-008 (Vergara) → US-009 (Conforti) → US-010 (Vergara) →
US-011 (Vergara). Cada una necesita a la anterior mergeada en `main`, así que la
primera tiene que entrar a mitad de semana. US-003, US-006, US-013 y US-019 no
dependen de la cadena.

**Antes de empezar cada historia.**

- Su spec está mergeada en `main` (Definition of Ready). Las de US-006, US-010 y
  US-011 y la corrección de US-009 estaban en revisión el 12/10.
- US-010 espera además a #100 (la estimación no puede cambiar un ítem
  completado), que hace Tomás.

**Migraciones.** Se numeran por adelantado para que dos ramas no usen el mismo
número (`Migrar` acepta migraciones fuera de orden, T-012):

| Número | Historia | Qué crea |
|---|---|---|
| 00004 | US-008 | Tabla `sprints` |
| 00005 | US-003 | `DEFAULT gen_random_uuid()` en `integrantes.id` |
| 00006 | US-019 | Tabla `registros_esfuerzo` |
| 00007 | US-009 | Tabla de la relación Sprint-ítem (histórica) |
| 00008 | US-010 | Columna `completado_en` en `items_backlog` |

US-006, US-011 y US-013 no llevan migración.

**Fechas del sprint.**

| Día | Qué pasa |
|---|---|
| Martes 13/10 | Planning. Specs de la segunda tanda mergeadas |
| Miércoles 14/10 | US-008 mergeada (de ella depende US-009) |
| Jueves 15/10 | Refinement. Cada historia tiene su PR en borrador con el RED a la vista |
| Viernes 16/10 | Revisión cruzada de todo lo abierto |
| Fin de semana | Colchón: todo mergeado o replanificado |
| Lunes 19/10 | Review con demo y Retro. El acta se commitea el mismo lunes |

**Reglas que se refuerzan** (acciones de la retro del Sprint 2):

1. Todo se mergea con *Rebase and merge*; no se usa *Squash*.
2. Todo PR se revisa en menos de 24 horas.
3. Antes de abrir un PR: `make check` con la base levantada, y el cuerpo del PR
   sin texto de terminal pegado.
4. "Revisado y entendido por" completo en la sección de IA, en el PR y en la spec.

**Sprint 4 — 20/10 a 26/10 — 44 SP.** Carga: Tomás 15, Vergara 13, Conforti 16.

| ID | Historia | SP | Responsable |
|---|---|---|---|
| US-014 | Iniciar sesión de Planning Poker | 5 | Vergara |
| US-015 | Voto individual oculto | 5 | Vergara |
| US-018 | Registrar la estimación acordada | 3 | Vergara |
| US-016 | Revelar votos y detectar divergencias | 5 | Tomás |
| US-027 | Horas estimadas, reales y desviación | 5 | Tomás |
| US-028 | Porcentaje de historias completadas | 2 | Tomás |
| US-029 | Defectos detectados y resueltos | 3 | Tomás |
| US-022 | Registrar un defecto | 5 | Conforti |
| US-023 | Cambiar el estado de un defecto | 3 | Conforti |
| T-011 | Pantallas de alta, edición y listado | 8 | Conforti |

**Sprint 5 — 27/10 a 02/11 — 43 SP.** Carga: Tomás 15, Vergara 15, Conforti 13.
**Congelamiento de funcionalidades: sábado 31/10** (corrido desde el jueves 29/10 en el
Planning del 20/10). De ahí a la presentación solo
hay documentación, defectos y ensayo.

| ID | Historia | SP | Responsable |
|---|---|---|---|
| US-030 | Dashboard | 5 | Conforti |
| US-034 | Exportar el reporte a PDF | 8 | Conforti |
| US-031 | Burndown del sprint | 5 | Vergara |
| US-033 | Reporte en pantalla | 5 | Vergara |
| T-008 | Documentación, manual e informe | 5 | Vergara |
| US-004 | Estado general de un proyecto | 2 | Tomás |
| US-012 | Sprints anteriores | 2 | Tomás |
| US-020 | Consultar esfuerzo | 3 | Tomás |
| US-021 | Comparar estimado contra real | 3 | Tomás |
| T-009 | Presentación y ensayo | 5 | Tomás |

**Riesgo R1 sigue crítico:** el plan exige unos 35 SP por sprint con una velocidad
medida de 5. Solo es posible si las specs ya están mergeadas, el trabajo se
paraleliza y no se repiten los errores del Sprint 1. Por eso el punto de control
del **12/10** es obligatorio.

### Compromiso del Sprint 4 (Planning del 20/10)

**Qué pide la guía y qué falta.** Se releyó la guía de la cátedra contra el backlog.
Todo lo que quedó en el plan, salvo cinco ítems, responde a un requerimiento mínimo
(Planning Poker incluido, sección 4, con "realizar nuevas rondas", que es US-017):

| Requerimiento de la guía | Historias que faltan | SP |
|---|---|---|
| 3. Sprints: cerrar, consultar anteriores | US-011, US-012 | 7 |
| 4. Planning Poker (votos ocultos, revelar, diferencias, rondas, estimación acordada) | US-014, US-015, US-016, US-017, US-018 | 21 |
| 5. Esfuerzo: comparar estimado y real | US-020, US-021 | 6 |
| 6. Defectos | US-022, US-023 | 8 |
| 7. Métricas | US-027, US-028, US-029 | 10 |
| 1. Estado de un proyecto | US-004 | 2 |
| 8. Dashboard, 9. Reportes y PDF | US-030, US-033, US-034 | 18 |
| Interfaz usable (criterio de evaluación) | T-011 | 8 |
| Entregables: documentación, manual y presentación | T-008, T-009 | 10 |
| **Total exigido por la guía** | | **90** |

**No lo exige la guía** (extras, 21 SP): US-031 burndown (5), US-007 (3), US-024 (3),
US-032 (5) y T-007 (5). US-031 sale de Sprint 5 y queda sin sprint como los otros
cuatro. **US-017 vuelve al plan:** había quedado en el recorte del Sprint 2 y es el
requerimiento "realizar nuevas rondas".

**La cuenta no cierra con la velocidad medida.** Quedan 90 SP exigidos, 80 de
funcionalidad (los 10 de documentación y presentación se hacen después del
congelamiento). Con el congelamiento corrido al sábado 31/10 hay 12 días: a 31 SP
por semana (el Sprint 3, con trabajo de rescate) son unos 53 SP, y a 16 SP por
semana (el Sprint 2), unos 27. Falta capacidad por entre 27 y 53 SP, y las
pantallas funcionales de cada requerimiento todavía no tienen SP propios. No se
resuelve recortando requerimientos de la guía. En el Planning del 20/10 el equipo
decidió:

1. **Planning Poker se entrega con una sola pantalla de votación** como mínimo
   defendible. Los otros requerimientos se definen igual en el Planning del Sprint 5.
2. **El congelamiento se corre al sábado 31/10**, lo que suma dos días de
   funcionalidad. La fecha de entrega podría extenderse más; el plan sigue hasta el
   02/11 y se ajusta si la cátedra lo confirma.
3. **Las pantallas no se simplifican:** cada requerimiento tiene que quedar
   funcional de punta a punta en el navegador, y el diseño visual se mejora al
   final. Las pantallas de Planning Poker, defectos, esfuerzo, métricas y dashboard
   se estiman como tareas propias (T-015 en adelante) en el Planning del Sprint 5.

**Compromiso del Sprint 4: 35 SP**: la velocidad del Sprint 3 (31) más T-015, que
sale de partir T-011 en dos al descubrir que no existe ninguna lectura (listar
proyectos, backlog o sprints). T-011 queda con proyectos, integrantes y backlog (8 SP) y
T-015 con los sprints (5 SP). Se toma el camino que desbloquea al resto: cierre de
sprint, pantallas y la base de Planning Poker.

**Sprint Goal:** el ciclo proyecto → backlog → sprint → cierre se usa desde el
navegador, se pueden iniciar sesiones de Planning Poker con votos ocultos, y el
producto calcula horas estimadas, reales, desviación y porcentaje de completitud.

| Persona | Compromiso | SP | Objetivo estirado | SP |
|---|---|---|---|---|
| Vergara | US-011, US-014, US-015 | 15 | US-018 | 3 |
| Tomás | T-011, T-015, US-027 | 18 | US-016, US-029 | 8 |
| Conforti | US-028 | 2 | — | — |

US-016 espera a que Vergara mergee US-014 y US-015, así que es el primero del
objetivo estirado de Tomás. Pasan al Sprint 5 con requerimiento pendiente: US-017,
US-022 y US-023 (11 SP).

**Cambios de responsable** (el campo *Assignees* del tablero y de cada issue se
actualiza con el Planning):

| ID | Antes | Ahora | Por qué |
|---|---|---|---|
| T-011 y T-015 | Conforti (T-011) | Tomás | T-004 y T-005 las dejó Tomás, y las pantallas usan el layout, el router, `Responder` y el manejo de errores que construyó |
| US-028 | Tomás | Conforti | Función pura con la spec ya escrita: trabajo acotado y fácil de verificar |

**Migraciones del Sprint 4.** Siguen la numeración del Sprint 3 y se reservan por
adelantado, como entonces:

| Número | Historia | Qué crea |
|---|---|---|
| 00009 | US-014 | Tablas de sesiones de Planning Poker y de sus participantes |
| 00010 | US-015 | Tabla de votos por ronda |
| 00011 en adelante | US-022 y US-023 | Defectos; el número se reserva al escribir su spec (Sprint 5) |

Según las specs actuales, US-011, US-016, US-017, US-018, US-027, US-028, US-029,
T-011 y T-015 no llevan migración. Si alguna la necesita, lo avisa en el PR de su spec.

Cada persona escribe las specs de lo que implementa, salvo US-027 y US-028, cuya
spec ya está escrita y se publica como PR por Tomás. Las de US-014 y US-015 (Vergara)
y la de T-011 (Tomás) tienen que estar mergeadas el **miércoles 21/10**; las de
US-016 y US-018, antes de empezarlas.

Los identificadores de las reglas, casos y criterios de las tareas llevan la `T`
(`RN-T011-k`, `CL-T011-k`, `CA-T011-k`), porque `RN-011-k` es de US-011 y las
etiquetas `@CA-011-k` de los escenarios se mezclarían.

**Acciones de la retro del Sprint 3** (tarea nueva T-014 y una regla de proceso):

- **T-014 — chequeo de `[ID]` en todos los commits de un PR** (Tomás, revisa
  Vergara, antes del 22/10). No cuenta en el compromiso.
- **Borrador del jueves:** el jueves 22/10 a las 22:00 cada historia tiene su PR
  en borrador con el RED a la vista. Quien no lo tiene pasa su historia a quien
  tenga capacidad; decide Tomás en el momento. Lo mismo vale para una revisión sin
  responder en 24 horas.

| Día | Qué pasa |
|---|---|
| Martes 20/10 | Planning. Cambios de responsable. Se publican las specs de US-027 y US-028 |
| Miércoles 21/10 | Specs de US-014, US-015, US-027, US-028 y T-011 mergeadas. Empieza el RED de US-011 |
| Jueves 22/10 | Refinement. **Borradores con RED a las 22:00.** T-014 mergeada |
| Viernes 23/10 | Revisión cruzada de todo lo abierto |
| Fin de semana | Colchón: todo mergeado o replanificado |
| Lunes 26/10 | Review con demo y Retro. El acta se commitea el mismo lunes. Planning del Sprint 5: estimación de las pantallas (T-015 en adelante) |

## Cómo se reparte y por qué

- **Tomás** (Agile Enabler): dominio de métricas y estimación, plataforma y
  trazabilidad. Lleva menos carga de historias porque además facilita las
  ceremonias, mantiene el tablero y revisa PRs.
- **Vergara**: dominio y persistencia (backlog, sprints, planning poker, esfuerzo).
- **Conforti**: proyectos, defectos, interfaz templ/HTMX, dashboard y reportes.
- **Todos escriben dominio con TDD en cada sprint**, aunque su área sea la
  interfaz. El criterio de evaluación mide participación individual y se
  revisa con `git shortlog -sn` y los PRs revisados por persona.
- **Quien implementa una historia escribe su spec y su `.feature`.** Otra
  persona los revisa (ver rotación de revisiones más abajo).

Cada historia se implementa de punta a punta: dominio con TDD y, si la
historia lo necesita, su repositorio en `internal/adapters/postgres/` con
migración nueva (nunca se edita una migración ya mergeada) y su handler y vista.

## Semana tipo

Los sprints van de martes a lunes.

| Día | Qué pasa |
|---|---|
| Martes | Sprint Planning (45 min). Sprint Goal escrito, issues del sprint a *Sprint Backlog* con responsable. **Se escriben las specs** de las historias del sprint |
| Miércoles | Specs mergeadas (PR aparte por historia). Empieza el ciclo RED → GREEN → REFACTOR |
| Jueves | Refinement (30 min): las historias del **próximo** sprint quedan con criterios `CA-NNN-k` y Story Points. Los PRs se abren en cuanto haya algo revisable, no el último día |
| Viernes | Revisión cruzada de todo lo abierto |
| Sábado y domingo | Colchón: todo mergeado o replanificado |
| Lunes | Review (30 min, demo con `docker compose up`) y Retro (30 min) a la noche. El acta `docs/retros/sprint-N.md` se commitea el mismo lunes |

Daily **escrita** en el grupo todos los días antes de las 22:00 con el formato
*hecho / hoy / bloqueos*, aunque se hable en persona: lo que no queda escrito no
sirve como evidencia.

> Las secciones siguientes son el **plan original**, que se conserva como referencia
> (dependencias y notas por historia). El plan vigente es el de arriba.

## Sprint 1 — El MVP (plan original, 33 SP)

**Sprint Goal:** se puede crear un proyecto con sus integrantes, cargar ítems de
backlog con criterios de aceptación, crear un sprint y asignarle historias, y el
dominio calcula los Story Points del sprint y la velocidad del equipo. El primer
escenario BDD corre automatizado de punta a punta.

| ID | Historia | SP | Responsable | Notas |
|---|---|---|---|---|
| US-001 | Crear un proyecto | 3 | Conforti | Primer escenario BDD de punta a punta (con T-006) |
| US-002 | Modificar un proyecto | 2 | Conforti | Depende de US-001 |
| US-003 | Registrar integrantes | 3 | Conforti | Depende de US-001 |
| US-005 | Crear un ítem de backlog | 5 | Vergara | Define `ItemBacklog`, prioridad y criterios de aceptación |
| US-006 | Editar y repriorizar ítems | 3 | Vergara | Depende de US-005 |
| US-008 | Crear un Sprint | 3 | Vergara | Define `Sprint` y Sprint Goal |
| US-009 | Asignar historias a un Sprint | 3 | Conforti | Depende de US-005 y US-008: se hace al final del sprint |
| US-025 | SP planificados y completados de un Sprint | 3 | Tomás | Función pura sobre un resumen del sprint, sin depender de la base |
| US-026 | Velocidad del equipo | 3 | Tomás | Depende de US-025 (`ResumenSprint`) |
| T-006 | Integrar godog y automatizar el primer escenario | 5 | Tomás | Se apoya en la spec y el `.feature` de US-001 |

Carga: Tomás 11 · Vergara 11 · Conforti 11.

## Sprint 2 — La interfaz (plan original, 37 SP)

**Sprint Goal:** un usuario recorre desde el navegador el ciclo completo:
proyecto → backlog filtrable → sprint → marcar historias completadas → cerrar el
sprint. Puede estimar en Fibonacci, registrar esfuerzo y comparar estimado
contra real, con validaciones y errores manejados de forma uniforme.

| ID | Historia | SP | Responsable | Notas |
|---|---|---|---|---|
| T-004 | Layout base, navegación y Tailwind | 5 | Conforti | Incluye el spike de templ + HTMX que quedó pendiente del Sprint 0. Va primero: el resto usa el layout |
| US-004 | Consultar el estado general de un proyecto | 2 | Conforti | |
| US-007 | Listar y filtrar el backlog | 3 | Conforti | Should |
| US-012 | Consultar sprints anteriores | 2 | Conforti | |
| US-011 | Cerrar un Sprint y devolver historias no completadas | 5 | Vergara | Regla de negocio central: mucho caso límite |
| US-010 | Marcar historias como completadas | 2 | Vergara | |
| US-019 | Registrar esfuerzo | 5 | Vergara | Va primero: US-020 y US-021 dependen de este |
| T-005 | Validaciones de formulario y errores uniformes | 5 | Tomás | El tipo de error común en `internal/platform` va antes que el resto |
| US-013 | Estimar con Story Points Fibonacci | 2 | Tomás | |
| US-020 | Consultar esfuerzo por historia, integrante y sprint | 3 | Tomás | Depende de US-019 |
| US-021 | Comparar esfuerzo estimado contra real | 3 | Tomás | Depende de US-019 y US-013 |

Carga: Tomás 13 · Vergara 12 · Conforti 12.

## Sprint 3 — Funcionalidad y calidad (plan original, 47 SP)

Es el sprint más cargado. **Antes de empezar, cada uno confirma su calendario
académico** (riesgo R7): si hay parciales esta semana, se mueve alcance al
Sprint 4 en el Planning, no a mitad de semana.

**Sprint Goal:** el equipo estima con Planning Poker completo (votación oculta,
revelado, divergencia, nuevas rondas y valor final), se registran defectos con
su ciclo de vida y el Dashboard muestra las métricas principales del proyecto.

| ID | Historia | SP | Responsable | Notas |
|---|---|---|---|---|
| US-014 | Iniciar sesión de Planning Poker | 5 | Vergara | Define `SesionPlanningPoker`. Va primero: el resto del poker depende de esto |
| US-015 | Voto individual oculto | 5 | Vergara | |
| US-017 | Nueva ronda ante divergencia | 3 | Vergara | Should. Primer candidato a recorte |
| US-018 | Registrar la estimación acordada | 3 | Vergara | |
| US-016 | Revelar votos y detectar divergencias | 5 | Tomás | Depende de US-014 y US-015 |
| US-027 | Horas estimadas, reales y desviación | 5 | Tomás | |
| US-028 | Porcentaje de historias completadas | 2 | Tomás | |
| US-029 | Defectos detectados y resueltos por sprint | 3 | Tomás | Función pura sobre conteos; se integra con US-022 y US-023 |
| US-022 | Registrar un defecto | 5 | Conforti | Defecto con severidad, historia relacionada y sprint de detección |
| US-023 | Cambiar el estado de un defecto | 3 | Conforti | Depende de US-022 |
| US-024 | Listar defectos por sprint y severidad | 3 | Conforti | Should. Candidato a recorte |
| US-030 | Dashboard con métricas principales | 5 | Conforti | Va al final: consume las métricas de Tomás |

Carga: Tomás 15 · Vergara 16 · Conforti 16.

## Sprint 4 — Cierre y entrega (plan original, 38 SP)

**Sprint Goal:** el proyecto se puede mostrar completo: gráficos, reporte del
proyecto o sprint en pantalla y exportable a PDF, documentación y manual
terminados, matriz de trazabilidad generada y una demo ensayada con nuestros
propios datos cargados en la app.

**Congelamiento de funcionalidades: sábado 31/10** (corrido desde el jueves 29/10 en el
Planning del 20/10). De ahí al lunes 02/11 es
solo documentación, corrección de defectos y ensayo de la presentación.

| ID | Historia | SP | Responsable | Notas |
|---|---|---|---|---|
| US-033 | Reporte completo de proyecto o Sprint en pantalla | 5 | Conforti | |
| US-034 | Exportar el reporte a PDF | 8 | Conforti | Depende de US-033. Es la historia más grande: arrancar el lunes |
| US-031 | Burndown de Story Points del sprint en curso | 5 | Vergara | |
| US-032 | Gráficos de velocidad histórica y defectos | 5 | Vergara | Should. Candidato a recorte |
| T-008 | Documentación técnica, manual de usuario e informe de métricas y cobertura | 5 | Vergara | Arranca el lunes y se va completando; no se deja para el final |
| T-007 | Generar la matriz de trazabilidad | 5 | Tomás | Could. Si no entra, la matriz se arma a mano en `docs/trazabilidad.md` |
| T-009 | Presentación final y ensayo del recorrido de trazabilidad | 5 | Tomás | Con datos reales: cargamos en la app nuestros propios sprints |

Carga: Tomás 10 · Vergara 15 · Conforti 13.

## Rotación de revisiones

Nadie aprueba su propio PR, y todos revisan a alguien:

| Autor | Lo revisa |
|---|---|
| Tomás | Vergara |
| Vergara | Conforti |
| Conforti | Tomás |

Si el revisor asignado no puede, lo revisa el tercero. Quien revisa puede
pedirle al autor que explique cualquier función (`docs/uso-de-ia.md`, regla 2).

## Lista de recorte

Si a mitad de sprint la velocidad no alcanza, se recorta en este orden y se
negocia con el Product Architect en la Review. **Nunca se recorta calidad, tests
ni specs.**

1. US-017, US-024, US-032 (Should)
2. T-007 (Could; la matriz se hace a mano)
3. Lo que se decida en el Planning, escrito en el acta

## Evidencia que buscan los profesores, y dónde queda

| Criterio (peso) | Qué tiene que existir | Dónde |
|---|---|---|
| Producto funcional (25%) | App corriendo con `docker compose --profile full up` y datos de ejemplo | Repo, README |
| SDD, BDD y TDD (25%) | Spec mergeada **antes** del código, `.feature` con los 4 tipos de caso, commits `test RED` → `feat GREEN` → `refactor` | `specs/`, `features/`, historial |
| Calidad del software (20%) | Dominio sin imports externos, errores envueltos, cobertura del dominio ≥ 85%, lint sin hallazgos, ADR por cada dependencia nueva | CI, `docs/adr/` |
| Gestión del proyecto (20%) | Tablero con las 4 vistas, un acta por sprint, PRs con revisión cruzada, commits convencionales | Project, `docs/retros/`, PRs |
| Equipo y presentación (10%) | Participación pareja, demo con datos propios, recorrido de trazabilidad ensayado | `git shortlog -sn`, Review final |

### Evidencia mínima al cierre de cada sprint

- [ ] Todas las historias del sprint con spec y `.feature` en `main`
- [ ] Ciclo RED → GREEN visible en el historial de cada historia con reglas de negocio
- [ ] `make check` y CI en verde en `main`
- [ ] Incremento demostrable en una máquina que no es la del autor
- [ ] Acta `docs/retros/sprint-N.md` con métricas reales y dos acciones con responsable
- [ ] Tablero al día: issues cerrados en *Done*, lo no terminado devuelto a *Backlog*
- [ ] Reparto revisado con `git shortlog -sn`

## Cosas que se configuran a mano

Están resueltas o asignadas en este orden, y quedan tildadas en el acta del
Sprint 1:

- [ ] Sumar a Vergara como co-dueño de `internal/domain/` en `.github/CODEOWNERS`, para que los PRs de dominio de Tomás también tengan un revisor posible
- [ ] Acceso al Project: profesores como lectura, compañeros con escritura
- [ ] Automatizaciones del Project: *Item closed* y *Pull request merged* a *Done*, *Item added* a *Backlog*
- [ ] Asignar los responsables de cada issue según las tablas de arriba
- [ ] Marcar con `sdd:pendiente` y `bdd:pendiente` las historias que todavía no tienen spec ni escenarios
- [ ] Cada integrante: `git config core.autocrlf input`, mismo email en Git y en GitHub, `make check` verde en su máquina
- [ ] Toda dependencia nueva fuera del stack de `docs/adr/0001-stack-tecnologico.md` lleva su ADR antes de usarse

## Cuando algo no sale como se planeó

- **Un integrante está trabado más de 45 minutos:** lo avisa en la Daily, no espera al día siguiente.
- **Una historia se agranda:** se divide o se baja alcance en el mismo sprint, y se anota en el acta.
- **Un defecto aparece:** issue nuevo con la plantilla de defecto y etiqueta `tipo:defecto`; se prioriza en el Planning siguiente.
- **CI rojo en `main`:** frena todo lo demás hasta arreglarlo, con un PR `fix`.
- **El sprint se atrasa:** se dice en la Review y se replanifica en la Retro. No se maquilla el tablero.
