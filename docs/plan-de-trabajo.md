# Plan de trabajo — Sprint 1 al Sprint 4

Este documento dice **quién hace qué y cuándo** hasta la entrega. Complementa a
`CONTRIBUTING.md`, que dice **cómo** se trabaja. Si los dos se contradicen,
manda `CONTRIBUTING.md` y se corrige este archivo.

Las fechas son estimativas; lo que no es negociable es que **haya avance
verificable todas las semanas**: un incremento demostrable, un acta de
retrospectiva y el tablero al día. El reparto y los objetivos se revisan en cada
Sprint Planning y cualquier cambio se anota en el acta del sprint.

| Sprint | Semana | Sprint Goal (resumen) | SP |
|---|---|---|---|
| 0 | 14/09 – 20/09 | Repo, CI, tablero y walking skeleton | 16 (cerrado) |
| 1 | 21/09 – 27/09 | El MVP: proyectos, backlog, sprints y primeras métricas | 33 |
| 2 | 28/09 – 04/10 | La interfaz: recorrido completo desde el navegador | 37 |
| 3 | 05/10 – 11/10 | Planning Poker, defectos y métricas completas | 47 |
| 4 | 12/10 – 18/10 | Gráficos, reporte PDF, documentación y demo | 38 |
| — | 19/10 | Review final y entrega | — |

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

| Día | Qué pasa |
|---|---|
| Lunes | Sprint Planning (45 min). Sprint Goal escrito, issues del sprint a *Sprint Backlog* con responsable. **Se escriben las specs** de las historias del sprint |
| Martes | Specs mergeadas (PR aparte por historia). Empieza el ciclo RED → GREEN → REFACTOR |
| Miércoles | Refinement (30 min): las historias del **próximo** sprint quedan con criterios `CA-NNN-k` y Story Points |
| Jueves | Implementación. Los PRs se abren en cuanto haya algo revisable, no el último día |
| Viernes | Todo lo del sprint mergeado o replanificado. Sábado queda de colchón |
| Domingo | Review (30 min, demo con `docker compose up`) y Retro (30 min). El acta `docs/retros/sprint-N.md` se commitea el mismo domingo |

Daily asíncrono todos los días antes de las 22:00 con el formato *hecho / hoy /
bloqueos*.

## Sprint 1 — El MVP (21/09 – 27/09, 33 SP)

**Sprint Goal:** se puede crear un proyecto con sus integrantes, cargar ítems de
backlog con criterios de aceptación, crear un sprint y asignarle historias, y el
dominio calcula los Story Points del sprint y la velocidad del equipo. El primer
escenario BDD corre automatizado de punta a punta.

| ID | Historia | SP | Responsable | Notas |
|---|---|---|---|---|
| US-001 | Crear un proyecto | 3 | Conforti | Primer escenario BDD de punta a punta (con T-006) |
| US-002 | Modificar un proyecto | 2 | Conforti | Depende de US-001 |
| US-003 | Registrar integrantes | 3 | Conforti | Depende de US-001 |
| US-005 | Crear un ítem de backlog | 5 | Vergara | Define `BacklogItem`, prioridad y criterios de aceptación |
| US-006 | Editar y repriorizar ítems | 3 | Vergara | Depende de US-005 |
| US-008 | Crear un Sprint | 3 | Vergara | Define `Sprint` y Sprint Goal |
| US-009 | Asignar historias a un Sprint | 3 | Conforti | Depende de US-005 y US-008: se hace al final del sprint |
| US-025 | SP planificados y completados de un Sprint | 3 | Tomás | Función pura sobre un resumen del sprint, sin depender de la base |
| US-026 | Velocidad del equipo | 3 | Tomás | Depende de US-025 (`SprintSummary`) |
| T-006 | Integrar godog y automatizar el primer escenario | 5 | Tomás | Se apoya en la spec y el `.feature` de US-001 |

Carga: Tomás 11 · Vergara 11 · Conforti 11.

## Sprint 2 — La interfaz (28/09 – 04/10, 37 SP)

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

## Sprint 3 — Funcionalidad y calidad (05/10 – 11/10, 47 SP)

Es el sprint más cargado. **Antes de empezar, cada uno confirma su calendario
académico** (riesgo R7): si hay parciales esta semana, se mueve alcance al
Sprint 4 en el Planning, no a mitad de semana.

**Sprint Goal:** el equipo estima con Planning Poker completo (votación oculta,
revelado, divergencia, nuevas rondas y valor final), se registran defectos con
su ciclo de vida y el Dashboard muestra las métricas principales del proyecto.

| ID | Historia | SP | Responsable | Notas |
|---|---|---|---|---|
| US-014 | Iniciar sesión de Planning Poker | 5 | Vergara | Define `PokerSession`. Va primero: el resto del poker depende de esto |
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

## Sprint 4 — Cierre y entrega (12/10 – 18/10, 38 SP)

**Sprint Goal:** el proyecto se puede mostrar completo: gráficos, reporte del
proyecto o sprint en pantalla y exportable a PDF, documentación y manual
terminados, matriz de trazabilidad generada y una demo ensayada con nuestros
propios datos cargados en la app.

**Congelamiento de funcionalidades: jueves 15/10.** Viernes a domingo es solo
documentación, corrección de defectos y ensayo.

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
