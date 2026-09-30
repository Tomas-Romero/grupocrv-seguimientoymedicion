# ADR 0002 — Idioma del dominio

- **Estado:** aceptada
- **Fecha:** 2026-09-29
- **Decide:** el equipo

## Contexto

US-005 necesita nombrar su tipo principal, y el repositorio ya tiene nombres en
los dos idiomas:

- La migración `00001_init.sql` crea `proyectos` e `integrantes`, con columnas
  como `fecha_inicio` y `creado_en`.
- La spec de US-026 (PR #52, en revisión) usa `SprintCerrado`, `ResumenSprint`
  y `ErrResumenInconsistente`.
- El código Go que ya existe usa `config.Cargar`, `LectorEntorno`,
  `db.Conectar` y `ErrFaltaDatabaseURL`.
- En cambio, el plan de trabajo y `CONTRIBUTING.md` tienen ejemplos en inglés:
  tipos como `BacklogItem`, `SprintSummary` y `PokerSession`, y paquetes y
  archivos como `metrics`, `defect` y `velocity.go`.

Sin una regla, cada historia elige por su cuenta y el mismo concepto termina
con dos nombres.

## Decisión

**El dominio, las tablas, las especificaciones y los nombres de paquetes y
archivos Go se escriben en español.**

| Qué | Idioma | Ejemplo |
|---|---|---|
| Tipos, campos, funciones y errores del dominio | Español | `ItemBacklog`, `ResumenSprint` |
| Paquetes y archivos Go | Español | `metricas`, `velocidad.go` |
| Tablas y columnas de las migraciones | Español | `proyectos`, `fecha_inicio` |
| Especificaciones SDD | Español | `specs/US-026-velocidad-equipo.md` |
| Términos de Scrum | Inglés | Sprint, Sprint Goal, Backlog, Story Points, Planning Poker, Daily, Review, Retrospective, Definition of Ready, Definition of Done |

- Los términos de Scrum de la lista quedan en inglés dentro de cualquier
  nombre, y el resto del nombre va en español: `SesionPlanningPoker`,
  `SprintGoal`, `ItemBacklog`.
- Solo son términos de Scrum los de la lista. *Velocidad*, por ejemplo, va en
  español: `velocidad.go`.
- Los identificadores Go van **sin acentos ni ñ**: `Descripcion`, no
  `Descripción`.

## Alternativas consideradas

**Todo en inglés.** Coincidiría con los ejemplos que ya tienen el plan y
`CONTRIBUTING.md`. Descartada: la migración `00001_init.sql` y la spec de US-026
ya están en español, y con el dominio en español la trazabilidad
spec → código queda literal: el término que usa la spec es el identificador
que se busca en el código.

## Consecuencias

**A favor**

- La trazabilidad spec → código es literal: lo que nombra la spec (por
  ejemplo, `ItemBacklog`) es lo que aparece en el código.
- No hay que renombrar la migración `00001_init.sql` ni los tipos de la spec de
  US-026; de esa spec solo cambia el paquete (ver abajo).

**En contra**

- Los nombres mezclan idiomas cuando incluyen un término de Scrum
  (`ItemBacklog`, `SesionPlanningPoker`).
- El plan de trabajo, `CONTRIBUTING.md` y la spec de US-026 tienen nombres en
  inglés que no siguen esta decisión. **No se editan en este ADR**; quedan
  listados para corregirlos aparte:

  | Archivo | Línea | Dónde | Nombres que no siguen la decisión |
  |---|---|---|---|
  | `docs/plan-de-trabajo.md` | 64 | Fila de US-005 | `BacklogItem` |
  | `docs/plan-de-trabajo.md` | 69 | Fila de US-026 | `SprintSummary` |
  | `docs/plan-de-trabajo.md` | 109 | Fila de US-014 | `PokerSession` |
  | `CONTRIBUTING.md` | 80 | §1, tabla de nombres: paquete Go | `metrics`, `defect` |
  | `CONTRIBUTING.md` | 81 | §1, tabla de nombres: archivo Go | `velocity.go` |
  | `CONTRIBUTING.md` | 148-154 | §2, Paso 5 (ejemplo del ciclo TDD) | `internal/domain/metrics/`, `velocity_test.go`, `velocity.go`, alcance `metrics` |
  | `CONTRIBUTING.md` | 158 | §2, Paso 5 (ejemplo del ciclo TDD) | `SprintSummary` |
  | `CONTRIBUTING.md` | 254 | §3, convención de commits: alcance | `metrics` |
  | `CONTRIBUTING.md` | 266 | §3, ejemplos de commits buenos | `SprintSummary` |
  | `specs/US-026-velocidad-equipo.md` (PR #52) | 9 | Encabezado, campo Código | `internal/domain/metrics/` |

  El PR #52 todavía no está mergeado: el ajuste de
  `specs/US-026-velocidad-equipo.md` lo hace su autor en ese PR.
