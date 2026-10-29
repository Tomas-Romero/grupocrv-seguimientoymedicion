# Software Metrics & Estimation

Aplicación web para administrar proyectos de software y obtener información de
estimación, planificación, seguimiento y calidad.

Trabajo Práctico Integrador — **Ingeniería y Calidad de Software**, Ingeniería en
Sistemas, UTN Facultad Regional San Rafael, 2026.

| Rol | Integrante | Legajo |
|---|---|:---:|
| **Product Architect** | Cátedra | — |
| **Agile Enabler** | Tomás Romero | `10289` |
| **Product Builder** | Angelo Conforti | `8861` |
| **Product Builder** | Juan Ignacio Vergara | `9896` |

---

## Levantarlo en tres comandos

Requisitos: Docker y Docker Compose.

```bash
git clone https://github.com/Tomas-Romero/grupocrv-seguimientoymedicion.git
cd grupocrv-seguimientoymedicion
docker compose --profile full up
```

La aplicación queda en **http://localhost:8080** con datos de ejemplo cargados.

### Para desarrollar

Requisitos adicionales:

- Go 1.23 o superior.
- `make` (GNU Make). En Linux y macOS viene con las herramientas de compilación. En
  Windows, Git for Windows no lo trae: instalalo con `winget install ezwinports.make`
  y abrí Git Bash de nuevo.
- `psql` (el cliente de PostgreSQL), que usa `make seed` en tu máquina. En Windows,
  el instalador oficial de PostgreSQL permite elegir solo "Command Line Tools".

`make tools` instala las herramientas con la versión fija que usa el proyecto
(templ, goose, air y golangci-lint), así que todos tienen las mismas.

```bash
cp .env.example .env
export DATABASE_URL="postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable"

make tools     # instala templ, goose, air, golangci-lint y el binario de Tailwind (una sola vez)
make up        # levanta Postgres, aplica migraciones y carga datos de ejemplo
make css       # compila Tailwind (baja el binario la primera vez; make dev ya lo corre)
make dev       # servidor con recarga automática en http://localhost:8080
make check     # lint + tests + cobertura + BDD (lo mismo que corre CI)
make help      # todos los targets disponibles
```

### En Windows

- Corré **todos los `make` desde Git Bash**, no desde PowerShell: el `Makefile` usa `bash`.
- `git config core.autocrlf input` antes de clonar o trabajar, para que los scripts no queden con finales de línea de Windows.
- Después de `make tools`, agregá `~/go/bin` al `PATH` de Git Bash: `export PATH="$PATH:$HOME/go/bin"`.
- El linter no soporta versiones de Go mucho más nuevas que la del CI. Si `make check` falla en lint con un error de "export data", corré `export GOTOOLCHAIN=go1.23.12`.
- Si ya tenés un PostgreSQL local usando el puerto 5432, o lo frenás o ponés `DB_PORT=5433` en `.env` y cambiás el puerto en `DATABASE_URL`.
- `make css` (y `make tools`) bajan `tailwindcss.exe` a `bin/` la primera vez, desde Git Bash y con internet. `bin/` no se versiona.

---

## Qué hace

- Gestión de proyectos e integrantes
- Product Backlog con prioridad, estado, Story Points y criterios de aceptación
- Sprints con Sprint Goal, asignación de historias y cierre
- Estimación con Story Points y **Planning Poker** con votación oculta y rondas
- Registro de esfuerzo real por integrante, fecha y actividad
- Gestión de defectos con severidad y trazabilidad al sprint
- Métricas: velocidad, desviación de esfuerzo, porcentaje de completitud, defectos
- Dashboard con gráficos
- Reporte de proyecto o sprint exportable a PDF

## Cómo está hecho

Go en el núcleo, arquitectura en capas con el dominio aislado de infraestructura.
Interfaz web renderizada en el servidor con `templ` y HTMX. PostgreSQL.

| Documento | Para qué |
|---|---|
| [`docs/documentacion-tecnica.md`](docs/documentacion-tecnica.md) | Qué es el sistema, arquitectura en capas, modelo de datos, decisiones clave, cómo correrlo y cómo se trabajó |
| [`docs/manual-de-usuario.md`](docs/manual-de-usuario.md) | Recorrido pantalla por pantalla, con capturas |
| [`docs/informe-metricas-cobertura.md`](docs/informe-metricas-cobertura.md) | Cobertura, tests, trazabilidad y métricas del proceso por sprint |
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Cómo trabajamos: carpetas, nombres, ramas, commits, PRs |
| [`docs/plan-de-trabajo.md`](docs/plan-de-trabajo.md) | Sprint Goals, responsables por historia y calendario hasta la entrega |
| [`specs/`](specs/) | Especificaciones SDD, una por funcionalidad |
| [`features/`](features/) | Escenarios BDD en Gherkin |
| [`docs/adr/`](docs/adr/) | Decisiones de arquitectura y por qué se tomaron |
| [`docs/retros/`](docs/retros/) | Actas de retrospectiva |
| [`docs/plantillas/`](docs/plantillas/) | Plantillas de spec SDD, escenario BDD y acta de retro |
| [`docs/uso-de-ia.md`](docs/uso-de-ia.md) | Política de uso de inteligencia artificial |
| [`docs/ci-cd.md`](docs/ci-cd.md) | Integración y entrega continua: qué verifica el CI, cómo se protege `main` y cómo se publica la imagen en Azure |
| [`docs/trazabilidad.md`](docs/trazabilidad.md) | Historia → Spec → Criterios → BDD → Tests → Código |

## Metodología

Scrum con sprints semanales que cierran los lunes (Sprint 0 más cinco sprints hasta la presentación del 02/11), Specification-Driven
Development para las funcionalidades principales, Behavior-Driven Development
para los escenarios de aceptación y Test-Driven Development para las reglas de
negocio y los cálculos.

El ciclo RED → GREEN → REFACTOR queda registrado commit a commit en el historial
de Git; ver `CONTRIBUTING.md` para la convención.
