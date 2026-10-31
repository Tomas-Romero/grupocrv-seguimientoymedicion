# Informe de métricas y cobertura

Este informe reúne los números de calidad del producto (cobertura, tests, escenarios BDD y trazabilidad)
y las métricas del proceso por sprint, sacadas de las actas de retrospectiva. Al final explica cómo
volver a calcular cada número.

**Medición:** 31/10/2026 (hora de Argentina), sobre `main` en el commit `b7094af`, con PostgreSQL 16
local. `scripts/coverage.sh`, `go test ./...` y los escenarios BDD terminaron en verde; el análisis
estático (`make lint`) lo corre el CI en cada PR y está en verde en `main`, y no se repitió a mano para esta
medición. El código de funcionalidades se congela el 31/10 a las 18:00: lo que sigue es documentación y
ensayo. Las métricas del proceso llegan hasta el Sprint 5, que se cerró el 31/10 (su acta es
[`retros/sprint-5.md`](retros/sprint-5.md)); la exposición del 02/11 se anota en una addenda de esa acta.

## 1. Cobertura

### Los dos números que exige el proyecto

`scripts/coverage.sh` (lo corre `make cover`, `make check` y el CI) mide dos umbrales y falla si no se
alcanzan:

| Qué mide | Resultado | Mínimo | Estado |
|---|---|---|---|
| **Dominio** (`internal/domain/...`, las reglas de negocio) | **99,8 %** | 85 % | Cumple |
| **`internal/`** completo (dominio, aplicación, adaptadores y plataforma) | **91,9 %** | 70 % | Cumple |

Los dos números son el total de sentencias cubiertas sobre el total de sentencias de esos paquetes. Cada
paquete cuenta solo lo que ejecutan **sus propios** tests (el script no usa `-coverpkg`): que la
aplicación ejecute código del dominio no suma cobertura al dominio. `cmd/` queda afuera a propósito:
`cmd/server` es el cableado de dependencias. `internal/demo` (el cargador de datos de demostración) cuenta dentro de `internal/`.

### Por paquete

Medida con `go test -cover ./...`, con el mismo criterio: cada paquete cuenta solo lo que ejecutan sus
propios tests. Por eso las plantillas y los pasos de los escenarios salen bajos aunque se ejecuten en cada
corrida: los ejercitan los tests de otros paquetes.

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
| `internal/app` | 95,8 % | Casos de uso, con la base en memoria |
| `internal/adapters/memoria` | 99,8 % | |
| `internal/adapters/postgres` | 84,6 % | Contra una base real, una temporal por test |
| `internal/adapters/http` | 90,6 % | Handlers, catálogo de errores y router, con `httptest` |
| `internal/platform/config` | 100,0 % | |
| `internal/platform/errores` | 100,0 % | |
| `internal/platform/db` | 76,2 % | Conexión, migraciones y datos de ejemplo |
| `internal/platform/db/dbprueba` | 81,5 % | Ayuda para los tests: crea y borra bases temporales |
| `internal/demo` | 77,4 % | Cargador del proyecto de demostración (T-023); verifica lo cargado contra su spec |
| `internal/trazabilidad` | 98,8 % | El generador de la matriz |
| `cmd/demo` | 82,0 % | |
| `cmd/trace` | 70,2 % | |
| `cmd/server` | 0,0 % | Cableado, sin tests unitarios (se prueba con el job *Perfil full* del CI) |
| `web/templates` | 16,9 % | Las plantillas se ejecutan sobre todo desde los tests de `internal/adapters/http`, que no cuentan acá |
| `features/steps` | 0,7 % | Los pasos se ejecutan desde `features/` (los escenarios), que no cuentan acá |

## 2. Tests y escenarios

| Qué | Cantidad |
|---|---:|
| Funciones de test (`func Test…`) | 995 |
| Tests que se ejecutan, contando subtests (`--- PASS` con `-v`) | 2435 |
| Tests salteados (`--- SKIP`) | 0 |
| Archivos `.feature` | 31 |
| Escenarios BDD definidos (22 son esquemas con ejemplos) | 309 |
| Escenarios BDD ejecutados (cada ejemplo de un esquema cuenta como uno) | 355, todos en verde |
| Pasos ejecutados en los escenarios | 2857, todos en verde |
| Definiciones de pasos (frases del [diccionario](diccionario-steps.md)) | 166 |
| Especificaciones SDD | 42 (32 historias y 10 tareas técnicas) |

Funciones de test por capa:

| Capa | Funciones de test |
|---|---:|
| Dominio (`internal/domain`) | 194 |
| Aplicación (`internal/app`) | 234 |
| Adaptador Postgres | 169 |
| Adaptador HTTP | 190 |
| Adaptador memoria | 119 |
| Plataforma (`internal/platform`) | 29 |
| Plantillas (`web/`) | 21 |
| Trazabilidad (`internal/trazabilidad`) | 16 |
| Datos de demostración (`internal/demo`) | 14 |
| `cmd/` y `features/` | 9 |

**Ciclo TDD en el historial:** `main` tiene 210 commits `RED` y 217 `GREEN`. En las primeras semanas
tres historias (US-001, US-025 y US-026) entraron con *Squash and merge* y su ciclo solo se ve en la
página del PR (#96, #82 y #92); desde el Sprint 3 todo se mergea con *Rebase and merge*.

## 3. Trazabilidad

La [matriz de trazabilidad](trazabilidad.md) se genera con `make trace` a partir de las specs, los
`.feature` y los comentarios de los tests. Un criterio "tiene escenario" si un escenario lleva su
etiqueta (`@CA-017-1`), y "tiene test" si un test lo cita en el comentario de arriba.

**Total:** 42 especificaciones con 285 criterios de aceptación; **216** tienen escenario BDD y **261**
tienen test.

### Por historia

| Historia | Criterios | Con escenario | Con test | | Historia | Criterios | Con escenario | Con test |
|---|---:|---:|---:|---|---|---:|---:|---:|
| US-001 | 5 | 5 | 5 | | US-017 | 5 | 4 | 5 |
| US-002 | 5 | 5 | 5 | | US-018 | 8 | 8 | 8 |
| US-003 | 9 | 9 | 9 | | US-019 | 11 | 9 | 7 |
| US-004 | 6 | 6 | 6 | | US-020 | 9 | 9 | 9 |
| US-005 | 8 | 8 | 8 | | US-021 | 9 | 8 | 9 |
| US-006 | 7 | 7 | 7 | | US-022 | 12 | 12 | 12 |
| US-007 | 6 | 6 | 6 | | US-023 | 14 | 13 | 14 |
| US-008 | 11 | 11 | 11 | | US-024 | 6 | 6 | 6 |
| US-009 | 8 | 8 | 8 | | US-025 | 4 | 4 | 1 |
| US-010 | 6 | 6 | 6 | | US-026 | 4 | 4 | 1 |
| US-011 | 6 | 6 | 6 | | US-027 | 6 | 6 | 6 |
| US-012 | 6 | 6 | 6 | | US-028 | 5 | 5 | 1 |
| US-013 | 6 | 6 | 5 | | US-029 | 6 | 6 | 6 |
| US-014 | 7 | 7 | 7 | | US-030 | 7 | 7 | 7 |
| US-015 | 7 | 7 | 7 | | US-033 | 5 | 4 | 5 |
| US-016 | 8 | 8 | 8 | | US-034 | 4 | 0 | 3 |

Las tareas técnicas (T-004, T-005, T-011, T-015, T-016, T-019, T-020, T-021, T-023 y T-024) no llevan escenarios
BDD: se prueban con tests de Go sobre el router real, o sobre el cargador y las plantillas. Tienen 59
criterios, 51 de ellos con test.

### Qué falta y por qué

**Todas las historias de la guía están implementadas en `main`**; ya no queda ninguna historia con spec y
sin código (en la medición del 28/10 faltaban US-024, US-030, US-033 y US-034).

**Criterios con test pero sin escenario BDD (7 de historias).** Los cubren tests de dominio, aplicación,
repositorio o de la pantalla; el `.feature` o la spec de cada historia explica por qué no tienen escenario:

| Criterio | Por qué no tiene escenario |
|---|---|
| CA-017-5, CA-023-14 | Son pedidos simultáneos: un escenario tiene un solo "Cuando". Los cubren tests de concurrencia contra Postgres |
| CA-021-9 | Datos inválidos ya guardados: ningún flujo de la aplicación puede guardarlos, así que solo se prueban en el dominio |
| CA-033-5 | Es un criterio de plataforma del reporte (sus secciones y el número de lecturas); lo cubren tests de la pantalla y del caso de uso |
| CA-034-1, CA-034-2, CA-034-3 | Son del botón de imprimir y de la hoja de estilos de impresión de US-034, que no tiene `.feature`: los cubren tests de la plantilla y del CSS (`TestReporte_BotonDeImprimir`, `TestEstilosDeImpresion`) |

CA-006-6 y CA-013-5, que figuraban sin escenario en la medición del 28/10, ya lo tienen.

**Criterios con escenario pero sin test que los cite (13).** CA-013-3, CA-019-4, CA-019-7, CA-025-1 a
CA-025-3, CA-026-1 a CA-026-3 y CA-028-1 a CA-028-4. Los escenarios existen y pasan, y en US-025, US-026 y
US-028 los tests de dominio existen pero citan las reglas (`RN-…`) y no los criterios en su comentario. Es un
hueco de la cita, no de la prueba.

**Criterios de historias sin escenario ni test que los cite (3).** CA-019-3 (registrar esfuerzo sobre una
historia completada) y CA-019-11 (las horas cuentan en el sprint de la fecha en que se trabajaron): no se
verificó para este informe qué test cubre cada uno. El tercero, CA-034-4, es una **evidencia documental** y no
un test: los PDF de ejemplo están en [`docs/evidencias/US-034/`](evidencias/US-034/).

**Tareas técnicas (8 criterios sin test que los cite).** CA-T004-5, CA-T015-5 y CA-T016-5 son de plataforma
(clon limpio, igualdad entre memoria y Postgres, comportamiento con Postgres); CA-T005-5 y CA-T011-3 son
límites; y CA-T024-1 (las cuatro pantallas con el diseño nuevo y los tests existentes sin debilitar), CA-T024-5
(un error de validación con estilo de alerta) y CA-T024-6 (accesibilidad) son de diseño. De este último hay
medición con `scripts/contraste.py` (26 combinaciones de color) y con Lighthouse, que dio 100 en Accesibilidad
en las cuatro pantallas (ver [`docs/diseno.md`](diseno.md)). La matriz no encuentra un test que los cite; no
se verificó si los cubre un test sin la cita.

## 4. Métricas del proceso por sprint

Datos de las actas de [`docs/retros/`](retros/) y del [plan de trabajo](plan-de-trabajo.md).

| Sprint | Fechas | SP comprometidos | SP completados | Cumplimiento | Defectos detectados | Defectos resueltos |
|---|---|---:|---:|---:|---|---|
| 0 | 14/09 – 24/09 (cierre previsto el 20/09) | 16 | 16 | 100 % | 2, no registrados como DEF | 2 |
| 1 | 21/09 – 05/10 (dos semanas) | 33 | 5 | 15 % | 5 (DEF-001 a DEF-005) | 3 (DEF-001 a DEF-003) |
| 2 | 06/10 – 12/10 | 16 | 16 | 100 % | 0 | 0 |
| 3 | 13/10 – 19/10 | 21 (techo 36) | 31 | 148 % | 1 (DEF-006) | 2 (DEF-005 y DEF-006); DEF-004 quedó resuelto en código |
| 4 | 20/10 – 26/10 | 35 (techo 46) | 77 (48 del tablero original) | 220 % (137 % sobre el tablero) | 1 (DEF-007) | 1 (DEF-007) |
| 5 | 27/10 – 02/11 (cerrado el 31/10) | 48 | 70 | 146 % (100 % del compromiso) | 1 (DEF-008) | 1 (DEF-008) |

**Sprint 5.** Cerró el 31/10, un día antes de lo previsto por decisión del equipo; el detalle está en su
[acta](retros/sprint-5.md). Los 70 SP son:

| Concepto | SP |
|---|---:|
| Compromiso entregado (48 SP): T-019, US-004, US-030, US-033, US-017, T-020, US-034, US-023, T-021, T-008 y T-009 | 48 |
| Objetivo estirado de Conforti: US-024 | 3 |
| Trabajo no previsto: T-022 (2), T-023 (5), T-024 (8), T-025 (2), T-026 (1) y T-027 (1) | 19 |
| **Total entregado** | **70** |

T-018 (despliegue en Azure, 5 SP) lleva la etiqueta `sprint:5` pero ya se contó dentro de los 77 SP del Sprint 4 (era
trabajo adelantado), así que no se vuelve a sumar. T-009 (la presentación) se da por entregada con su PR; la exposición es el 02/11.

**Velocidad.** El Sprint 0 fue de puesta en marcha (repositorio, CI, tablero), así que la velocidad se
calcula sobre los sprints 1 a 4:

| Cálculo | Velocidad |
|---|---:|
| Promedio de SP completados, sprints 1 a 4 (5, 16, 31 y 77) | 32,25 SP por sprint |
| Con el Sprint 5, sprints 1 a 5 (5, 16, 31, 77 y 70) | 39,8 SP por sprint |
| Igual, pero con los 48 SP del tablero original del Sprint 4 (sin el trabajo adelantado del Sprint 5) | 25 SP por sprint |
| Sprints 1 a 5 con los 48 SP del tablero original del Sprint 4 (5, 16, 31, 48 y 70) | 34 SP por sprint |

El acta del Sprint 4 advierte que 29 de sus 77 SP fueron trabajo adelantado o no previsto, y que la
velocidad del Sprint 3 incluyó trabajo de rescate (riesgos R11 y R5): por eso el Planning del Sprint 5
comprometió 48 SP y no 77.

**Defectos.** Se registraron 8 defectos (DEF-001 a DEF-008), todos resueltos. Cinco fueron de
puesta en marcha y de entorno (DEF-001 a DEF-005: Docker, Windows, `make`), cuatro de ellos detectados en
la prueba en máquina limpia del 30/09; uno de documentación (DEF-006) y uno de lógica de negocio (DEF-007: un sprint cerrado contaba como completada una
historia completada después en otro sprint), corregido el mismo día con su test. DEF-008 (del Sprint 5) fue de puesta en marcha: `make up` se cortaba si Postgres todavía arrancaba; se corrigió el 31/10 (#251).

**Horas estimadas y reales.** Las horas estimadas salen de los Story Points a 1,5 horas por SP: 24 h en
los sprints 0 y 2, 31,5 h en el 3 y 52,5 h en el 4. **Las horas reales no se registraron en ningún
sprint cerrado**, así que la desviación de esfuerzo no se puede calcular con datos del equipo (las cinco
actas lo anotan como pendiente). La pantalla para cargar horas (T-019) llegó a `main` el 28/10, y la
acción de la retro del Sprint 4 es que desde ese día cada integrante cargue las suyas; las horas del
Sprint 5 no se registraron en el repositorio (ver su acta).

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
| Escenarios y pasos ejecutados | `go test -count=1 ./features/ -v` (las últimas líneas dicen `355 scenarios` y `2857 steps`) |
| Escenarios definidos | `grep -hE '^\s*(Escenario\|Esquema del escenario):' features/*.feature \| wc -l` |
| Definiciones de pasos | `grep -rhoE '\.Step\(' features/steps/*.go \| wc -l` |
| Matriz de trazabilidad | `make trace` (regenera `docs/trazabilidad.md`); para mirarla sin tocar el archivo, `go run ./cmd/trace > /tmp/matriz.md` |
| Commits RED y GREEN | `git log main --oneline \| grep -c ' RED '` (y `' GREEN '`) |
| Métricas por sprint | Tabla "Métricas del sprint" de cada acta en `docs/retros/sprint-N.md` |
| Defectos | `gh issue list --state all --label tipo:defecto` |
