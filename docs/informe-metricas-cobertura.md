# Informe de métricas y cobertura

Este informe reúne los números de calidad del producto (cobertura, tests, escenarios BDD y trazabilidad)
y las métricas del proceso por sprint, sacadas de las actas de retrospectiva. Al final explica cómo
volver a calcular cada número.

**Medición:** 28/10/2026 a las 22:10 (hora de Argentina), sobre `main` en el commit `c71c708`, con
PostgreSQL 16 local. `make check` terminó en verde. Las métricas del proceso llegan hasta el Sprint 4: el
Sprint 5 termina con la presentación del 02/11 y su acta todavía no existe.

## 1. Cobertura

### Los dos números que exige el proyecto

`scripts/coverage.sh` (lo corre `make cover`, `make check` y el CI) mide dos umbrales y falla si no se
alcanzan:

| Qué mide | Resultado | Mínimo | Estado |
|---|---|---|---|
| **Dominio** (`internal/domain/...`, las reglas de negocio) | **99,8 %** | 85 % | Cumple |
| **`internal/`** completo (dominio, aplicación, adaptadores y plataforma) | **93,0 %** | 70 % | Cumple |

Los dos números son el total de sentencias cubiertas sobre el total de sentencias de esos paquetes. Cada
paquete cuenta solo lo que ejecutan **sus propios** tests (el script no usa `-coverpkg`): que la
aplicación ejecute código del dominio no suma cobertura al dominio. `cmd/` queda afuera a propósito:
`cmd/server` es el cableado de dependencias.

### Por paquete

Medida con `go test -cover ./...`, con el mismo criterio: cada paquete cuenta solo lo que ejecutan sus
propios tests. Por eso las plantillas y los pasos de los escenarios salen casi en cero aunque se ejecuten
en cada corrida: los ejercitan los tests de otros paquetes.

| Paquete | Cobertura | Comentario |
|---|---:|---|
| `internal/domain/backlog` | 100,0 % | |
| `internal/domain/defecto` | 100,0 % | |
| `internal/domain/esfuerzo` | 100,0 % | |
| `internal/domain/integrante` | 100,0 % | |
| `internal/domain/metricas` | 99,0 % | |
| `internal/domain/planningpoker` | 100,0 % | |
| `internal/domain/proyecto` | 100,0 % | |
| `internal/domain/sprint` | 100,0 % | |
| `internal/app` | 97,3 % | Casos de uso, con la base en memoria |
| `internal/adapters/memoria` | 99,8 % | |
| `internal/adapters/postgres` | 84,6 % | Contra una base real, una temporal por test |
| `internal/adapters/http` | 90,8 % | Handlers, catálogo de errores y router, con `httptest` |
| `internal/platform/config` | 100,0 % | |
| `internal/platform/errores` | 100,0 % | |
| `internal/platform/db` | 76,2 % | Conexión, migraciones y datos de ejemplo |
| `internal/platform/db/dbprueba` | 81,5 % | Ayuda para los tests: crea y borra bases temporales |
| `internal/trazabilidad` | 98,8 % | El generador de la matriz |
| `cmd/trace` | 70,2 % | |
| `cmd/server` | 0,0 % | Cableado, sin tests unitarios (se prueba con el job *Perfil full* del CI) |
| `web/templates` | 3,6 % | Las plantillas se ejecutan desde los tests de `internal/adapters/http`, que no cuentan acá |
| `features/steps` | 0,9 % | Los pasos se ejecutan desde `features/` (los escenarios), que no cuentan acá |

## 2. Tests y escenarios

| Qué | Cantidad |
|---|---:|
| Funciones de test (`func Test…`) | 867 |
| Tests que se ejecutan, contando subtests (`--- PASS` con `-v`) | 2115 |
| Tests salteados (`--- SKIP`) | 0 |
| Archivos `.feature` | 27 |
| Escenarios BDD definidos (17 son esquemas con ejemplos) | 259 |
| Escenarios BDD ejecutados (cada ejemplo de un esquema cuenta como uno) | 293, todos en verde |
| Pasos ejecutados en los escenarios | 2239, todos en verde |
| Definiciones de pasos (frases del [diccionario](diccionario-steps.md)) | 134 |
| Especificaciones SDD | 39 (31 historias y 8 tareas técnicas) |

Funciones de test por capa:

| Capa | Funciones de test |
|---|---:|
| Dominio (`internal/domain`) | 190 |
| Aplicación (`internal/app`) | 178 |
| Adaptador Postgres | 167 |
| Adaptador HTTP | 148 |
| Adaptador memoria | 117 |
| Plataforma (`internal/platform`) | 29 |
| Plantillas (`web/`) | 16 |
| Trazabilidad (`internal/trazabilidad`) | 16 |
| `cmd/` y `features/` | 6 |

**Ciclo TDD en el historial:** `main` tiene 189 commits `RED` y 194 `GREEN`. En las primeras semanas
tres historias (US-001, US-025 y US-026) entraron con *Squash and merge* y su ciclo solo se ve en la
página del PR (#96, #82 y #92); desde el Sprint 3 todo se mergea con *Rebase and merge*.

## 3. Trazabilidad

La [matriz de trazabilidad](trazabilidad.md) se genera con `make trace` a partir de las specs, los
`.feature` y los comentarios de los tests. Un criterio "tiene escenario" si un escenario lleva su
etiqueta (`@CA-017-1`), y "tiene test" si un test lo cita en el comentario de arriba.

**Total:** 39 especificaciones con 267 criterios de aceptación; **190** tienen escenario BDD y **223**
tienen test.

### Por historia

| Historia | Criterios | Con escenario | Con test | | Historia | Criterios | Con escenario | Con test |
|---|---:|---:|---:|---|---|---:|---:|---:|
| US-001 | 5 | 5 | 5 | | US-018 | 8 | 8 | 8 |
| US-002 | 5 | 5 | 5 | | US-019 | 11 | 9 | 7 |
| US-003 | 9 | 9 | 9 | | US-020 | 9 | 9 | 9 |
| US-004 | 6 | 6 | 6 | | US-021 | 9 | 8 | 9 |
| US-005 | 8 | 8 | 8 | | US-022 | 12 | 12 | 12 |
| US-006 | 7 | 5 | 6 | | US-023 | 14 | 13 | 14 |
| US-008 | 11 | 11 | 11 | | US-024 | 6 | 0 | 0 |
| US-009 | 8 | 8 | 8 | | US-025 | 4 | 4 | 1 |
| US-010 | 6 | 6 | 6 | | US-026 | 4 | 4 | 1 |
| US-011 | 6 | 6 | 5 | | US-027 | 6 | 6 | 6 |
| US-012 | 6 | 6 | 6 | | US-028 | 5 | 5 | 1 |
| US-013 | 6 | 5 | 5 | | US-029 | 6 | 6 | 6 |
| US-014 | 7 | 7 | 7 | | US-030 | 7 | 0 | 0 |
| US-015 | 7 | 7 | 7 | | US-033 | 5 | 0 | 0 |
| US-016 | 8 | 8 | 8 | | US-034 | 4 | 0 | 0 |
| US-017 | 5 | 4 | 5 | | | | | |

Las tareas técnicas (T-004, T-005, T-011, T-015, T-016, T-019, T-020 y T-021) no llevan escenarios BDD:
se prueban con tests de Go sobre el router real. Tienen 47 criterios, 42 de ellos con test.

### Qué falta y por qué

**Historias todavía sin implementar en `main` (22 criterios, sin escenario ni test).** US-024, US-030,
US-033 y US-034 tienen su spec mergeada y el código en curso o pendiente.

**Criterios con test pero sin escenario BDD (5).** Los cubren tests de dominio, aplicación o repositorio;
el `.feature` de cada historia explica por qué no tienen escenario:

| Criterio | Por qué no tiene escenario |
|---|---|
| CA-006-6 | Cuando se escribió US-006 no existía la frase para completar una historia (US-010). El escenario quedó pendiente y no se agregó después |
| CA-013-5 | Igual: necesitaba la frase de historia completada, de US-010 |
| CA-017-5, CA-023-14 | Son pedidos simultáneos: un escenario tiene un solo "Cuando". Los cubren tests de concurrencia contra Postgres |
| CA-021-9 | Datos inválidos ya guardados: ningún flujo de la aplicación puede guardarlos, así que solo se prueban en el dominio |

**Criterios con escenario pero sin test que los cite (14).** CA-011-3, CA-013-3, CA-019-4, CA-019-7,
CA-025-1 a CA-025-3, CA-026-1 a CA-026-3 y CA-028-1 a CA-028-4. Los escenarios existen y pasan, y en
US-025, US-026 y US-028 los tests de dominio existen pero citan las reglas (`RN-…`) y no los criterios en
su comentario. Es un hueco de la cita, no de la prueba.

**Criterios sin escenario ni test que los cite (3).**

- CA-006-3 (editar una historia asignada a un sprint abierto) y CA-019-3 (registrar esfuerzo sobre una
  historia completada): el `.feature` dice que los cubren los tests del dominio y de los repositorios,
  pero ningún test cita el criterio, y el escenario quedó pendiente de las frases de sprints.
- CA-019-11 (las horas cuentan en el sprint de la fecha en que se trabajaron).

No se verificó para este informe qué test cubre cada uno.

**Tareas técnicas (5 criterios sin test que los cite).** CA-T004-5, CA-T015-5 y CA-T016-5 son de
plataforma (clon limpio, igualdad entre memoria y Postgres, comportamiento con Postgres) y CA-T005-5 y
CA-T011-3 son límites. La matriz no encuentra un test que los cite; no se verificó si los cubre un test
sin la cita.

## 4. Métricas del proceso por sprint

Datos de las actas de [`docs/retros/`](retros/) y del [plan de trabajo](plan-de-trabajo.md).

| Sprint | Fechas | SP comprometidos | SP completados | Cumplimiento | Defectos detectados | Defectos resueltos |
|---|---|---:|---:|---:|---|---|
| 0 | 14/09 – 24/09 (cierre previsto el 20/09) | 16 | 16 | 100 % | 2, no registrados como DEF | 2 |
| 1 | 21/09 – 05/10 (dos semanas) | 33 | 5 | 15 % | 5 (DEF-001 a DEF-005) | 3 (DEF-001 a DEF-003) |
| 2 | 06/10 – 12/10 | 16 | 16 | 100 % | 0 | 0 |
| 3 | 13/10 – 19/10 | 21 (techo 36) | 31 | 148 % | 1 (DEF-006) | 2 (DEF-005 y DEF-006); DEF-004 quedó resuelto en código |
| 4 | 20/10 – 26/10 | 35 (techo 46) | 77 (48 del tablero original) | 220 % (137 % sobre el tablero) | 1 (DEF-007) | 1 (DEF-007) |
| 5 | 27/10 – 02/11 | 48 | en curso | — | — | — |

**Velocidad.** El Sprint 0 fue de puesta en marcha (repositorio, CI, tablero), así que la velocidad se
calcula sobre los sprints 1 a 4:

| Cálculo | Velocidad |
|---|---:|
| Promedio de SP completados, sprints 1 a 4 (5, 16, 31 y 77) | 32,25 SP por sprint |
| Igual, pero con los 48 SP del tablero original del Sprint 4 (sin el trabajo adelantado del Sprint 5) | 25 SP por sprint |

El acta del Sprint 4 advierte que 29 de sus 77 SP fueron trabajo adelantado o no previsto, y que la
velocidad del Sprint 3 incluyó trabajo de rescate (riesgos R11 y R5): por eso el Planning del Sprint 5
comprometió 48 SP y no 77.

**Defectos.** Se registraron 7 defectos (DEF-001 a DEF-007), todos resueltos al 26/10. Cinco fueron de
puesta en marcha y de entorno (DEF-001 a DEF-005: Docker, Windows, `make`), cuatro de ellos detectados en
la prueba en máquina limpia del 30/09; uno de documentación (DEF-006) y uno de lógica de negocio (DEF-007: un sprint cerrado contaba como completada una
historia completada después en otro sprint), corregido el mismo día con su test.

**Horas estimadas y reales.** Las horas estimadas salen de los Story Points a 1,5 horas por SP: 24 h en
los sprints 0 y 2, 31,5 h en el 3 y 52,5 h en el 4. **Las horas reales no se registraron en ningún
sprint cerrado**, así que la desviación de esfuerzo no se puede calcular con datos del equipo (las cinco
actas lo anotan como pendiente). La pantalla para cargar horas (T-019) llegó a `main` el 28/10, y la
acción de la retro del Sprint 4 es que desde ese día cada integrante cargue las suyas; las horas del
Sprint 5 se informan en su acta.

## 5. Cómo regenerar cada número

Todos los comandos se corren desde la raíz del repositorio, en Git Bash en Windows, con `DATABASE_URL`
apuntando a un PostgreSQL de prueba.

| Número | Comando |
|---|---|
| Cobertura del dominio y de `internal/` | `make cover` (o `bash scripts/coverage.sh`); deja `coverage.html` |
| Cobertura por paquete | `go test -count=1 -cover ./...` |
| Todo lo del CI, en verde | `make check` |
| Funciones de test | `grep -rhE '^func Test' --include=*_test.go . \| wc -l` |
| Tests ejecutados, salteados y fallidos | `go test -count=1 ./... -v > salida.txt` y después `grep -c -- '--- PASS' salida.txt` (y `--- SKIP`, `--- FAIL`) |
| Escenarios y pasos ejecutados | `go test -count=1 ./features/ -v` (las últimas líneas dicen `293 scenarios` y `2239 steps`) |
| Escenarios definidos | `grep -hE '^\s*(Escenario\|Esquema del escenario):' features/*.feature \| wc -l` |
| Definiciones de pasos | `grep -rhoE '\.Step\(' features/steps/*.go \| wc -l` |
| Matriz de trazabilidad | `make trace` (regenera `docs/trazabilidad.md`); para mirarla sin tocar el archivo, `go run ./cmd/trace > /tmp/matriz.md` |
| Commits RED y GREEN | `git log main --oneline \| grep -c ' RED '` (y `' GREEN '`) |
| Métricas por sprint | Tabla "Métricas del sprint" de cada acta en `docs/retros/sprint-N.md` |
| Defectos | `gh issue list --state all --label tipo:defecto` |
