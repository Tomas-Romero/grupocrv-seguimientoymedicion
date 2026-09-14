# ADR 0001 — Stack tecnológico

- **Estado:** aceptada
- **Fecha:** 2026-09-14
- **Decide:** el equipo, en la reunión de arranque del Sprint 0

## Contexto

La cátedra exige que el núcleo de la solución y sus reglas de negocio estén
desarrollados en Go, y deja libre la forma de la interfaz (web, escritorio o
consola). El equipo son tres personas con cinco semanas de desarrollo efectivo y
experiencia previa en desarrollo web, no en Go.

Los criterios de evaluación premian la usabilidad, la calidad del código Go, la
arquitectura y la cobertura de pruebas.

## Decisión

**Aplicación web renderizada en el servidor, íntegramente en Go.**

| Componente | Elección |
|---|---|
| Lenguaje | Go 1.23 |
| Router | `go-chi/chi/v5` |
| Vistas | `a-h/templ` (plantillas tipadas compiladas a Go) |
| Interactividad | HTMX 2.x |
| Estilos | Tailwind CSS (binario standalone, sin Node) |
| Persistencia | PostgreSQL 16 con `jackc/pgx/v5` |
| Migraciones | `pressly/goose/v3` |
| Gráficos | Chart.js sobre JSON servido por handlers Go |
| PDF | `johnfercher/maroto/v2` |

## Alternativas consideradas

**API Go + frontend React o Next.** Más cómodo por experiencia previa del equipo,
pero divide el esfuerzo en dos stacks y hace que parte del trabajo evaluado en
"calidad del código Go" ocurra en código que no es Go. Descartada.

**Aplicación de consola o TUI.** La más rápida de construir y 100% Go, pero
"usabilidad" es un criterio explícito y el requerimiento 8 pide un Dashboard con
representaciones gráficas. Descartada.

**SQLite en vez de PostgreSQL.** Menos fricción para el evaluador (archivo único,
sin Docker), pero menos defendible en el criterio de arquitectura. Se mitiga la
fricción de Postgres con `docker compose --profile full up` y datos de ejemplo
precargados. Descartada, con el riesgo R3 registrado.

## Consecuencias

**A favor**

- Todo el código evaluable es Go: el criterio de calidad se juega en un solo lugar.
- Un solo repositorio, un solo build, un solo despliegue.
- El dominio queda aislado y se puede testear sin base de datos, que es la
  condición para sostener TDD durante todo el proyecto.

**En contra**

- `templ` y HTMX son nuevos para el equipo. Se mitiga con un spike en el Sprint 0
  (riesgo R2) y con el plan de contingencia de caer a `html/template` de la
  biblioteca estándar si el spike supera las 4 horas.
- Postgres exige Docker para evaluar. Se mitiga con el perfil `full` del compose
  y con una prueba en máquina limpia al cierre de cada sprint (riesgo R3).
