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
git clone https://github.com/Tomas-Romero/metrics-estimation.git
cd metrics-estimation
docker compose --profile full up
```

La aplicación queda en **http://localhost:8080** con datos de ejemplo cargados.

### Para desarrollar

Requisitos adicionales: Go 1.23 o superior.

```bash
cp .env.example .env
export DATABASE_URL="postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable"

make tools     # instala templ, goose y air (una sola vez)
make up        # levanta Postgres, aplica migraciones y carga datos de ejemplo
make dev       # servidor con recarga automática en http://localhost:8080
make check     # lint + tests + cobertura + BDD (lo mismo que corre CI)
make help      # todos los targets disponibles
```

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
| [`CONTRIBUTING.md`](CONTRIBUTING.md) | Cómo trabajamos: carpetas, nombres, ramas, commits, PRs |
| [`specs/`](specs/) | Especificaciones SDD, una por funcionalidad |
| [`features/`](features/) | Escenarios BDD en Gherkin |
| [`docs/adr/`](docs/adr/) | Decisiones de arquitectura y por qué se tomaron |
| [`docs/retros/`](docs/retros/) | Actas de retrospectiva |
| [`docs/plantillas/`](docs/plantillas/) | Plantillas de spec SDD, escenario BDD y acta de retro |
| [`docs/uso-de-ia.md`](docs/uso-de-ia.md) | Política de uso de inteligencia artificial |
| [`docs/trazabilidad.md`](docs/trazabilidad.md) | Historia → Spec → Criterios → BDD → Tests → Código |

## Metodología

Scrum con sprints de una semana (Sprint 0 más cuatro sprints de desarrollo), Specification-Driven
Development para las funcionalidades principales, Behavior-Driven Development
para los escenarios de aceptación y Test-Driven Development para las reglas de
negocio y los cálculos.

El ciclo RED → GREEN → REFACTOR queda registrado commit a commit en el historial
de Git; ver `CONTRIBUTING.md` para la convención.
