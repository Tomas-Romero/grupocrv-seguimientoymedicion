# Software Metrics & Estimation

Aplicación web para administrar un proyecto de software y obtener información de su **estimación, planificación, seguimiento y calidad**:
backlog, sprints, Planning Poker, esfuerzo, defectos, métricas, dashboard y reportes.

Trabajo Práctico Integrador — **Ingeniería y Calidad de Software**, Ingeniería en Sistemas, UTN Facultad Regional San Rafael, 2026.

| Rol | Integrante | Legajo |
|---|---|:---:|
| **Product Architect** | Cátedra | — |
| **Agile Enabler** | Tomás Romero | `10289` |
| **Product Builder** | Angelo Conforti | `8861` |
| **Product Builder** | Juan Ignacio Vergara | `9896` |

**Stack:** Go 1.23 · PostgreSQL 16 · `net/http` + [templ](https://templ.guide) + HTMX + Tailwind · godog (BDD) · Docker · GitHub Actions · Azure Container Instances.

---

## Índice

1. [Probarlo en 2 minutos (camino del evaluador)](#1-probarlo-en-2-minutos-camino-del-evaluador)
2. [Datos de demostración](#2-datos-de-demostración)
3. [Cumplimiento de la guía del TP](#3-cumplimiento-de-la-guía-del-tp)
4. [Recorrido para probar cada funcionalidad](#4-recorrido-para-probar-cada-funcionalidad)
5. [Todos los comandos](#5-todos-los-comandos)
6. [Sistemas operativos](#6-sistemas-operativos)
7. [Metodología: Scrum, SDD, BDD, TDD e IA](#7-metodología-scrum-sdd-bdd-tdd-e-ia)
8. [Trazabilidad](#8-trazabilidad)
9. [Arquitectura y estructura del repositorio](#9-arquitectura-y-estructura-del-repositorio)
10. [Automatización: CI/CD y Azure](#10-automatización-cicd-y-azure)
11. [Mapa de la documentación](#11-mapa-de-la-documentación)
12. [Si algo falla](#12-si-algo-falla)

---

## 1. Probarlo en 2 minutos (camino del evaluador)

**Requisitos: solo Docker y Docker Compose.** No hace falta tener Go ni nada más; funciona igual en Linux, macOS y Windows, porque la
aplicación corre dentro de un contenedor Linux.

```bash
git clone https://github.com/Tomas-Romero/grupocrv-seguimientoymedicion.git
cd grupocrv-seguimientoymedicion
docker compose --profile full up --build
```

La aplicación queda en **http://localhost:8080** (la primera vez compila la imagen, lo que tarda unos minutos). Arranca con un proyecto
de ejemplo mínimo cargado. Para detenerla: `Ctrl+C`, y `docker compose --profile full down` (agregá `-v` para borrar también la base).

Para ver **todas** las funcionalidades con datos completos (sprints cerrados, Planning Poker, horas, defectos, dashboard y reportes), cargá
los [datos de demostración](#2-datos-de-demostración) con un comando más.

> El CI corre este mismo camino en cada PR, sobre un volumen vacío, en Ubuntu (job «Perfil full», `scripts/verificar-perfil-full.sh`).

---

## 2. Datos de demostración

El comando `cmd/demo` llena la base con un proyecto completo, **«Tienda online»**, siempre con los mismos datos ficticios: 4 integrantes, 16
historias, 4 sprints (3 cerrados y 1 abierto), 3 sesiones de Planning Poker, 27 registros de horas y 10 defectos. Usa los casos de uso de la
aplicación (no SQL) y, al terminar, comprueba lo cargado contra las cifras de su especificación
([`specs/T-023-datos-de-demostracion.md`](specs/T-023-datos-de-demostracion.md)).

Con la aplicación levantada (la base queda en el puerto 5432) y **Go 1.23+** instalado:

```bash
DATABASE_URL="postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable" go run ./cmd/demo
```

Para volver a empezar: `go run ./cmd/demo --borrar` (borra **solo** ese proyecto, por su nombre exacto) y se carga de nuevo. Detalle, cómo
cargarlo en una base remota sin exponer la cadena de conexión y qué mostrar en cada pantalla:
[`docs/datos-de-demostracion.md`](docs/datos-de-demostracion.md).

---

## 3. Cumplimiento de la guía del TP

Cada requerimiento de la guía, dónde se ve en la aplicación y dónde está su especificación (`specs/US-NNN-….md`) y sus escenarios
(`features/US-NNN-….feature`). Las pantallas cuelgan de `/proyectos/{id}/…`.

| # | Requerimiento de la guía | Dónde se ve | Historias |
|---|---|---|---|
| 1 | Crear y modificar proyectos, integrantes, fechas, estado | `/proyectos`, Resumen del proyecto (tarjeta «Estado general») | US-001, US-002, US-003, US-004 |
| 2 | Product Backlog (id, título, descripción, prioridad, estado, Story Points, criterios) | Backlog | US-005, US-006, US-007 (el filtro está en la aplicación, no en la pantalla) |
| 3 | Sprints: crear, Sprint Goal, asignar, completar, cerrar, anteriores | Sprints, detalle del sprint, Sprints anteriores | US-008 a US-012 |
| 4 | Estimación con Story Points y **Planning Poker** (votos ocultos, mostrar, detectar diferencias, nuevas rondas, estimación acordada) | Backlog → Planning Poker | US-013 a US-018 |
| 5 | Registro de esfuerzo y comparación estimado contra real | Esfuerzo | US-019, US-020, US-021 |
| 6 | Defectos (descripción, severidad, estado, historia, sprint de detección y de resolución) | Defectos | US-022, US-023, US-024 |
| 7 | Métricas: Story Points planificados y completados, velocidad, horas estimadas y reales, desviación, % completadas, defectos detectados y resueltos | Dashboard, detalle del sprint, Reporte | US-025 a US-029 |
| 8 | Dashboard con gráficos | Dashboard | US-030 |
| 9 | Reportes de un proyecto o sprint y **exportación a PDF** | Reporte (proyecto y cada sprint) → «Imprimir o guardar como PDF» | US-033, US-034 |

Notas honestas sobre lo que **no** es literal:

- **PDF:** el reporte se exporta con el diálogo de impresión del navegador («Guardar como PDF»), con una hoja de estilos de impresión propia
  (A4, sin menú, encabezado con proyecto y fecha, tablas que no se parten). No se genera un archivo en el servidor; la decisión está en
  [`specs/US-034-vista-imprimible.md`](specs/US-034-vista-imprimible.md) y hay PDF de ejemplo en [`docs/evidencias/US-034/`](docs/evidencias/US-034/).
- **Fuera de alcance, a propósito:** el burndown (US-031) y los gráficos de velocidad histórica (US-032) no se implementaron; la guía no los
  exige y el dashboard ya trae dos gráficos de métricas.
- **Sin autenticación:** la aplicación no tiene login (la guía no lo pide). Quien conozca la URL puede ver y cambiar datos; por eso el despliegue
  en Azure se enciende solo durante la demostración.

---

## 4. Recorrido para probar cada funcionalidad

Con los datos de demostración cargados, abrí **«Tienda online»** en `/proyectos`:

1. **Resumen:** estado del proyecto según sus fechas, historias por estado, Story Points, porcentaje e integrantes con su rol.
2. **Backlog:** historias con prioridad MoSCoW y estado en insignias. Elegí los puntos de una historia sin estimar (la 11, la 14, la 15 o la 16) y tocá
   «Estimar»: la escala es Fibonacci (1, 2, 3, 5, 8, 13).
3. **Planning Poker en vivo:** en la historia 16 («votando», con 2 votos ocultos) votá como Bruno y como Diego, tocá «Revelar los votos» y, si hay
   diferencia, «Abrir nueva ronda» y registrá la estimación acordada. La 11 ya está revelada con divergencia.
4. **Sprints:** el Sprint 4 está abierto; marcá una historia como completada. **Sprints anteriores** muestra el resultado de cada sprint cerrado y las
   historias que volvieron al backlog. (Cerrar un sprint no se puede deshacer: hacelo al final.)
5. **Esfuerzo:** horas por integrante, actividad y fecha, y la comparación de horas estimadas (1 Story Point = 1,5 h) contra las reales.
6. **Defectos:** los 10 defectos con severidad, estado y sprints de detección y resolución; filtrá por sprint y severidad.
7. **Dashboard:** estado general, sprint de referencia, velocidad, defectos y dos gráficos.
8. **Reporte:** el reporte completo del proyecto y el de cada sprint. «Imprimir o guardar como PDF» abre el diálogo del navegador.

Todo lo que se hace en vivo queda guardado en la base. Para volver al punto de partida: `go run ./cmd/demo --borrar` y cargar de nuevo. El manual
con capturas de cada pantalla está en [`docs/manual-de-usuario.md`](docs/manual-de-usuario.md).

---

## 5. Todos los comandos

### Con Docker (sin instalar nada más)

| Comando | Qué hace |
|---|---|
| `docker compose --profile full up --build` | Base + aplicación en http://localhost:8080, con datos de ejemplo |
| `docker compose up -d` | Solo la base (para desarrollar) |
| `docker compose --profile full down` | Baja todo; los datos quedan en el volumen (`-v` los borra) |
| `bash scripts/verificar-perfil-full.sh` | Comprueba el camino del evaluador sobre un volumen vacío (lo corre el CI) |

### Para desarrollar (Go + make)

Requisitos: **Go 1.23+**, Docker, `make` y `psql` (el cliente de PostgreSQL, que usa `make seed`). En Linux y macOS `make` viene con las herramientas
de compilación; en Windows ver [Sistemas operativos](#6-sistemas-operativos).

```bash
cp .env.example .env                    # variables de entorno (puerto de la base, dirección HTTP, nivel de log)
export DATABASE_URL="postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable"

make tools     # instala templ, goose, air, golangci-lint y el binario de Tailwind (una sola vez)
make up        # levanta Postgres, aplica las migraciones y carga los datos de ejemplo
make dev       # servidor con recarga automática en http://localhost:8080
```

| Comando | Qué hace |
|---|---|
| `make help` | Lista todos los targets |
| `make up` / `make down` | Levanta (con migraciones y datos de ejemplo) o baja la base |
| `make gen` | Genera el código Go de las plantillas `templ` |
| `make css` | Compila Tailwind (baja el binario la primera vez) |
| `make dev` | Servidor con recarga automática |
| `make run` / `make build` | Compila y ejecuta / solo compila a `./bin/server` |
| `make migrate` / `make seed` | Aplica las migraciones pendientes / carga los datos de ejemplo |
| `go run ./cmd/demo` | Carga el proyecto de demostración completo (ver [sección 2](#2-datos-de-demostración)) |
| `make test` | Todos los tests unitarios, con `-race` |
| `make bdd` | Los escenarios BDD (godog) |
| `go test ./features/ -godog.tags=@US-007` | Los escenarios de **una** historia (o de un criterio: `@CA-017-1`) |
| `make cover` | Tests con cobertura y verificación de umbrales (`scripts/coverage.sh`) |
| `make lint` / `make fmt` | `go vet` y `golangci-lint` / formato |
| `make check` | **Todo lo que corre el CI** (lint + cobertura + BDD). Correrlo antes de abrir un PR |
| `make trace` | Regenera la matriz de trazabilidad (`docs/trazabilidad.md`) |
| `make clean` | Borra los artefactos generados |

Tres cosas que conviene saber:

- **Después de clonar hay que generar el código:** los archivos de `templ` (`*_templ.go`) y el CSS compilado (`app.css`) **no se versionan**. Por
  eso un `go test ./...` o `go build ./...` directo falla hasta correr `make gen` (y `make css` para ver la interfaz). `make test`, `make bdd`,
  `make run` y `make dev` ya lo hacen solos, igual que el `Dockerfile`.
- **Sin `DATABASE_URL`, los tests de Postgres se omiten** (y pasan en verde). Para correr toda la suite contra una base real, exportala antes
  (`make up` levanta una).
- **Cobertura exigida:** el dominio (`internal/domain/…`) al menos **85 %** y `internal/` al menos **70 %**; `make cover` falla si no se cumplen.
  Los valores actuales están en [`docs/informe-metricas-cobertura.md`](docs/informe-metricas-cobertura.md).

---

## 6. Sistemas operativos

| Sistema | Estado | Cómo se sabe |
|---|---|---|
| **Linux (Ubuntu)** | ✅ **Verificado en cada PR** | El CI corre en `ubuntu-latest`: lint, tests con `-race` contra un Postgres 16, escenarios BDD, cobertura y el camino del evaluador (`docker compose --profile full`). La imagen se construye en Ubuntu y **corre en Azure Container Instances (Linux)** |
| **macOS** | ➖ **No verificado en CI** | El código y el `Makefile` no tienen nada específico de un sistema: el `Makefile` baja el binario correcto de Tailwind para macOS Intel y Apple Silicon y el `Dockerfile` soporta `arm64`. El camino con Docker es idéntico. Una salvedad: el `Makefile` usa `/bin/bash`, que en macOS es la versión 3.2; los scripts que invocan los `make` no usan construcciones de bash 4+ (revisado por búsqueda, **no ejecutado en macOS**), pero `scripts/crear-issues.sh` y `scripts/mover-sprints.sh` (gestión del tablero, que no hacen falta para correr ni probar nada) usan `declare -A` y requieren bash 4+ |
| **Windows** | ✅ Es donde se desarrolló | Correr los `make` desde **Git Bash**, no desde PowerShell |

**¿Influye el sistema operativo?** Para correr la aplicación, **no**: es Go, PostgreSQL y HTML, sin código que dependa del sistema. Solo influye en
las **herramientas de desarrollo** (`make`, `bash`, `psql` y el binario de Tailwind). Por eso el camino recomendado para probar es Docker.

Notas por sistema:

- **Linux y macOS:** `git clone` y los comandos de la [sección 5](#5-todos-los-comandos) tal cual. Todos los scripts de `scripts/` empiezan con
  `#!/usr/bin/env bash` y el repositorio fuerza finales de línea LF (`.gitattributes`).
- **Windows:**
  - `make` no viene con Git for Windows: `winget install ezwinports.make` y abrir Git Bash de nuevo.
  - `git config core.autocrlf input` antes de clonar, para que los scripts no queden con finales de línea de Windows.
  - Después de `make tools`, `export PATH="$PATH:$HOME/go/bin"`.
  - Si `make check` falla en lint con un error de «export data», `export GOTOOLCHAIN=go1.23.12`.
  - `make css` y `make tools` bajan `tailwindcss.exe` a `bin/` la primera vez (con internet; `bin/` no se versiona).
- **Go más viejo que 1.23:** con `GOTOOLCHAIN=auto` (el valor por defecto desde Go 1.21) el comando `go` baja solo la versión que pide `go.mod`.
- **Puerto 5432 ocupado** (por ejemplo, por un PostgreSQL ya instalado): `DB_PORT=5433` en `.env` y el mismo puerto en `DATABASE_URL`.

---

## 7. Metodología: Scrum, SDD, BDD, TDD e IA

### Scrum

- **Roles:** Product Architect (la cátedra), Agile Enabler (Tomás) y Product Builders (Angelo y Juan).
- **Sprints semanales** que cierran los lunes: Sprint 0 (preparación) y cinco sprints hasta la presentación del 02/11. Las **ceremonias**: Planning
  los martes, Refinement los jueves (con las specs del sprint siguiente ya abiertas como PR), Review y Retrospectiva los lunes a la noche, y
  Daily escrita en el grupo todos los días.
- **Artefactos:** el Product Backlog y los Sprint Backlogs están en el tablero de GitHub Projects (con *Definition of Ready* y *Definition of Done* en
  [`CONTRIBUTING.md`](CONTRIBUTING.md)), el plan de cada sprint en [`docs/plan-de-trabajo.md`](docs/plan-de-trabajo.md) y las actas de retrospectiva en
  [`docs/retros/`](docs/retros/).

### SDD — la especificación va antes del código

Cada funcionalidad principal tiene su especificación en [`specs/`](specs/), con los ocho apartados que pide la guía: objetivo, entradas, salidas
esperadas, reglas de negocio, restricciones, casos límite, condiciones de error y criterios de aceptación (plantilla en
[`docs/plantillas/spec-sdd.md`](docs/plantillas/spec-sdd.md)). **La spec se mergea antes de abrir la rama de implementación** (*Definition of Ready*) y, si la
implementación descubre algo, se corrige en el mismo PR.

### BDD — escenarios Given–When–Then en español

Los escenarios están en [`features/`](features/) (Gherkin en español, un `.feature` por historia) y se **automatizan con godog**. Cubren casos normales,
alternativos, límite (`@limite`) y errores (`@error`). Las frases permitidas están en [`docs/diccionario-steps.md`](docs/diccionario-steps.md).

### TDD — RED → GREEN → REFACTOR en el historial

Las reglas de negocio y los cálculos se desarrollan con el ciclo **RED → GREEN → REFACTOR**, y la evidencia es el historial de Git: un commit `test(…)` que
falla **antes** de cada `feat(…)` que lo hace pasar. Para verlo:

```bash
git log --oneline --grep='RED' --grep='GREEN' --grep='refactor(' | head -30
git log --oneline --grep='\[US-017\]'          # el ciclo completo de una historia
```

Las pruebas unitarias cubren cálculo de métricas, cálculos de estimación, reglas de negocio y validaciones. Convención de commits y ramas en
[`CONTRIBUTING.md`](CONTRIBUTING.md).

### Inteligencia Artificial

Se usó IA como soporte (análisis de requisitos, especificaciones, código, pruebas, refactorización, revisión y documentación), con cuatro reglas:
**declarar** el uso en cada PR, **entender** todo antes de mergear, que **los tests salgan de la especificación** y que **sin spec no se genera código
del dominio**. Detalle en [`docs/uso-de-ia.md`](docs/uso-de-ia.md); cada PR lleva su sección «Uso de IA».

---

## 8. Trazabilidad

La guía pide demostrar **Historia de Usuario → Especificación SDD → Criterios de Aceptación → Escenarios BDD → Tests → Código Go**. La matriz completa se
genera sola con `make trace` y está en [`docs/trazabilidad.md`](docs/trazabilidad.md): por cada criterio de aceptación, su escenario y los tests que lo citan.

Ejemplo, **US-017 (nueva ronda de Planning Poker)**:

| Eslabón | Dónde |
|---|---|
| Historia | issue de la historia en el tablero |
| Especificación | [`specs/US-017-nueva-ronda.md`](specs/US-017-nueva-ronda.md) |
| Criterios de aceptación | `CA-017-1` … en la sección 8 de la spec |
| Escenarios BDD | [`features/US-017-nueva-ronda.feature`](features/US-017-nueva-ronda.feature) (`go test ./features/ -godog.tags=@CA-017-1`) |
| Tests | los que citan `CA-017-…`: `internal/domain/planningpoker/nueva_ronda_test.go`, `internal/app/abrir_nueva_ronda_test.go` y los de `internal/adapters/memoria` y `internal/adapters/postgres` |
| Código | `internal/domain/planningpoker/nueva_ronda.go` y el caso de uso `internal/app/abrir_nueva_ronda.go` |

---

## 9. Arquitectura y estructura del repositorio

Go en el núcleo, arquitectura en capas con el **dominio puro** (sin SQL, HTTP ni reloj) aislado de la infraestructura; los errores se devuelven envueltos
con contexto y se traducen a un formato uniforme en el borde. Las decisiones están en [`docs/adr/`](docs/adr/).

```
cmd/server/            punto de entrada de la aplicación web
cmd/demo/              carga del proyecto de demostración
cmd/trace/             genera la matriz de trazabilidad
internal/domain/       reglas de negocio puras: backlog, sprint, planningpoker, esfuerzo, defecto, metricas, proyecto, integrante
internal/app/          casos de uso (orquestan el dominio y los repositorios)
internal/adapters/     http (handlers) · postgres (persistencia) · memoria (para los tests y los escenarios BDD)
internal/platform/     configuración, conexión a la base y migraciones
web/templates/         plantillas templ        web/static/  CSS y JS estáticos (HTMX incluido, sin CDN)
migraciones/           migraciones de la base (goose)
specs/                 especificaciones SDD     features/    escenarios BDD y sus steps
docs/                  documentación, actas, ADR y evidencias
scripts/               scripts de automatización, verificación y gestión del tablero
.github/workflows/     CI y CD
```

---

## 10. Automatización: CI/CD y Azure

Todo corre en **GitHub Actions**, en Ubuntu. El detalle, con cómo se protege `main` y cómo se configuró Azure, está en [`docs/ci-cd.md`](docs/ci-cd.md).

| Workflow | Qué automatiza |
|---|---|
| `ci.yml` | En cada PR y push: **Lint** (`go vet`, `golangci-lint`, formato y `actionlint` sobre los workflows), **Tests y cobertura** (con umbrales), **BDD**, **Trazabilidad** y **Perfil full** (el camino del evaluador) |
| `pr-title.yml` | Convenciones: título del PR con su ID, nombre de rama, `Closes #N` y el ID en cada commit |
| `cd-azure.yml` | Construye la imagen, la publica en el registro (`ingcalidad2026.azurecr.io/grupo03-grupocrv`) y la despliega en **Azure Container Instances**; comprueba `/health` y la apaga para no gastar créditos |
| `aci-energia.yml` | Enciende, apaga o informa el estado del contenedor (y lo apaga todas las noches) |
| `verificar-azure.yml` | Comprueba la configuración real de Azure sin cambiar nada (solo lectura) |

- **`main` está protegida:** checks obligatorios, rama al día, una aprobación de alguien que no sea el autor, dueños del código (`.github/CODEOWNERS`) e
  historia lineal (*Rebase and merge*, para conservar los commits RED y GREEN).
- **Credenciales fuera del repositorio:** `AZURE_CREDENTIALS` (el JSON del *service principal*), `ACR_PASSWORD` y `DATABASE_URL` son secretos de un
  *Environment* limitado a `main`; el nombre del registro, de la imagen y del contenedor son variables. Las acciones de terceros que reciben
  secretos (`azure/login`, `docker/*`) están fijadas por el **hash exacto** de su versión.
- **Base de datos en la nube:** PostgreSQL externo y gratuito (Neon). Cada integrante desarrolla con su Postgres local.
- **Costos:** el contenedor de ACI cobra mientras está encendido; por eso el despliegue automático está apagado por defecto (`CD_ACTIVO`) y se
  enciende unos minutos antes de la demostración. `bash scripts/probar-despliegue-azure.sh <región>` hace la prueba completa y lo vuelve a apagar.

---

## 11. Mapa de la documentación

| Documento | Para qué |
|---|---|
| [`docs/manual-de-usuario.md`](docs/manual-de-usuario.md) | Recorrido pantalla por pantalla, con capturas |
| [`docs/documentacion-tecnica.md`](docs/documentacion-tecnica.md) | Arquitectura en capas, modelo de datos, decisiones clave, cómo correrlo y cómo se trabajó |
| [`docs/informe-metricas-cobertura.md`](docs/informe-metricas-cobertura.md) | Cobertura, tests, trazabilidad y métricas del proceso por sprint |
| [`docs/trazabilidad.md`](docs/trazabilidad.md) | Historia → Spec → Criterios → BDD → Tests |
| [`docs/datos-de-demostracion.md`](docs/datos-de-demostracion.md) | Los datos de la demo, cómo cargarlos y qué mostrar |
| [`docs/diseno.md`](docs/diseno.md) | Guía de estilos de la interfaz y contrastes medidos |
| [`docs/ci-cd.md`](docs/ci-cd.md) | CI, protección de `main`, despliegue en Azure, costos y qué se comprobó con la cuenta real |
| [`docs/plan-de-trabajo.md`](docs/plan-de-trabajo.md) | Sprint Goals, responsables por historia y calendario hasta la entrega |
| [`docs/retros/`](docs/retros/) | Actas de retrospectiva |
| [`docs/uso-de-ia.md`](docs/uso-de-ia.md) | Política de uso de inteligencia artificial |
| [`docs/adr/`](docs/adr/) | Decisiones de arquitectura y por qué se tomaron |
| [`docs/diccionario-steps.md`](docs/diccionario-steps.md) | Frases permitidas en los escenarios BDD |
| [`docs/plantillas/`](docs/plantillas/) | Plantillas de spec SDD, escenario BDD y acta de retro |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Cómo trabajamos: carpetas, nombres, ramas, commits, PR, Definition of Ready y de Done |
| [`specs/`](specs/) · [`features/`](features/) | Las especificaciones SDD y los escenarios BDD |

---

## 12. Si algo falla

| Síntoma | Causa y solución |
|---|---|
| `go test ./...` o `go build ./...` falla con «undefined» en `web/templates` | Faltan los archivos generados: `make gen` (y `make css` para los estilos) |
| La interfaz se ve sin estilos | Falta el CSS compilado: `make css` |
| Los tests de Postgres «pasan» muy rápido | Sin `DATABASE_URL` se omiten: exportala (`make up` levanta la base) |
| `port is already allocated` / el puerto 5432 está ocupado | `DB_PORT=5433` en `.env` y el mismo puerto en `DATABASE_URL` |
| `docker compose` dice que ya existen `metrics-db` o `metrics-app` | `make down` (los datos quedan en el volumen) o `docker compose --profile full down` |
| `make check` falla en lint con «export data» | `export GOTOOLCHAIN=go1.23.12` |
| La demo en Azure «no carga» después de encenderla | ACI cambia la IP y el DNS local puede guardar la vieja unos minutos: ventana privada o `ipconfig /flushdns` (ver `docs/ci-cd.md`) |
| `go run ./cmd/demo` dice que ya existe el proyecto | `go run ./cmd/demo --borrar` y volver a cargar |
