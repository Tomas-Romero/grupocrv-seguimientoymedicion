# T-004 — Layout base, navegación y sistema visual con Tailwind

| | |
|---|---|
| **Estado** | revision |
| **Autor** | Angelo Conforti |
| **Issue** | #N (spec) · tarea: #14 |
| **Escenarios BDD** | No aplica: tarea técnica sin regla de negocio (ver Restricciones) |
| **Código** | `web/templates/`, `internal/adapters/http/` |
| **Última actualización** | 2026-10-19 |

> Esta especificación se escribe y se mergea **antes** de mergear la
> implementación. Si durante la implementación algo de acá resulta
> equivocado, se corrige la spec en el mismo PR.

## 1. Objetivo

Dar a todas las pantallas del producto un layout común con navegación, y una
convención única para responder páginas completas o fragmentos HTMX. Lo usan
T-011 (pantallas de alta, edición y listado) y todo handler que muestre una
pantalla.

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `Pagina.Titulo` | `string` | Sí | Texto no vacío; se muestra en `<title>` |
| `Pagina.Activo` | `Seccion` | No | Una sección del menú; vacío = ninguna marcada |
| Contenido | `templ.Component` | Sí | El cuerpo de la pantalla |
| Cabecera `HX-Request` | HTTP | No | `true` activa el modo fragmento |
| Cabecera `HX-Boosted` | HTTP | No | `true` fuerza la página completa |
| Ruta y método | HTTP | Sí | Cualquiera |

## 3. Salidas esperadas

- **Página completa:** HTML que empieza con `<!doctype html>`, con
  `lang="es"`, `<title>{Titulo} | Software Metrics</title>`, enlace a
  `/static/css/app.css` y al script de HTMX, navegación y el contenido
  dentro de `<main>`.
- **Fragmento:** solo el HTML del contenido, sin layout.
- **Cabeceras:** `Content-Type: text/html; charset=utf-8` y `Vary: HX-Request`.
- **Estáticos:** el archivo con su tipo de contenido.

**Esquema de URL de la navegación** (lo que T-011 va a implementar):

| Ítem del menú | Ruta | Estado en esta tarea |
|---|---|---|
| Inicio | `/` | implementada |
| Proyectos | `/proyectos` | enlace presente; la ruta responde 404 hasta T-011 |
| Backlog de un proyecto | `/proyectos/{id}/backlog` | solo documentada |
| Sprints de un proyecto | `/proyectos/{id}/sprints` | solo documentada |

## 4. Reglas de negocio

No hay reglas de dominio. Son reglas de comportamiento de la interfaz:

- **RN-004-1** — Una petición es de fragmento si y solo si trae
  `HX-Request: true` y no trae `HX-Boosted: true`.
- **RN-004-2** — Toda respuesta de pantalla lleva `Vary: HX-Request`.
- **RN-004-3** — El código HTTP lo decide quien llama a `Responder`, tanto en
  página completa como en fragmento.
- **RN-004-4** — La navegación marca con `aria-current="page"` solo el ítem
  activo; una página sin sección activa no marca ninguno.
- **RN-004-5** — `GET /` muestra el inicio. Cualquier otra ruta no registrada
  responde 404 con la página de error y código 404 (no 200).
- **RN-004-6** — `/static/` sirve archivos y nunca lista directorios.
- **RN-004-7** — `/health` conserva su comportamiento (JSON, 200 o 503).
- **RN-004-8** — Si el render falla, se responde 500 sin escribir contenido
  parcial, y el error se registra en el log.
- **RN-004-9** — La página completa incluye los landmarks `header`, `nav` y
  `main` y un enlace "saltar al contenido".
- **RN-004-10** — El layout reserva un contenedor para mensajes y errores
  uniformes, sin depender del tipo de error de T-005.
- **RN-004-11** — Las versiones de Tailwind y HTMX quedan fijadas en el
  repositorio y HTMX se sirve como archivo propio, sin CDN.

## 5. Restricciones

- Sin reglas de negocio: no se toca `internal/domain` y no hay `.feature`.
  El comportamiento se verifica con tests Go (`httptest` y render a buffer).
- Stack fijado por el ADR 0001: templ v0.2.793, HTMX 2.x, Tailwind standalone.
- El CSS compilado y los `*_templ.go` no se versionan (`.gitignore`); el
  `Dockerfile` compila el CSS para que el perfil `full` se vea con estilos.
- Los handlers no tienen lógica: renderizan y responden.
- Las pantallas de Proyectos, Backlog y Sprints quedan fuera de alcance (T-011).
- La app tiene que funcionar sin internet (defensa).

## 6. Casos límite

- **CL-004-1** — `HX-Request: false` u otro valor cuenta como petición normal.
- **CL-004-2** — `HX-Request: true` junto con `HX-Boosted: true` devuelve la
  página completa.
- **CL-004-3** — `/static/` y `/static/css/` (directorios) dan 404, sin listado.
- **CL-004-4** — Un estático inexistente da 404 sin pasar por la página 404
  con layout.
- **CL-004-5** — El 404 por HTMX devuelve el fragmento con código 404.
- **CL-004-6** — Una página sin sección activa (el 404) no marca ningún ítem.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| Falla el render de la plantilla | HTTP 500, sin contenido parcial | "error interno" |
| Ruta inexistente | HTTP 404 con la página de error | "Página no encontrada" |
| Estático inexistente o directorio | HTTP 404 | el estándar de `net/http` |
| Base caída en `/health` | HTTP 503 con `{"estado":"error"}` | — |

## 8. Criterios de aceptación

- **CA-004-1** (normal) — `GET /` responde 200 con el inicio dentro del
  layout y con Inicio marcado como activo.
- **CA-004-2** (alternativo) — Con `HX-Request: true`, la misma ruta responde
  solo el fragmento y con `Vary: HX-Request`.
- **CA-004-3** (límite) — Con `HX-Request` y `HX-Boosted`, responde la página
  completa; `/static/` y sus directorios dan 404 sin listado.
- **CA-004-4** (error) — Una ruta inexistente responde 404 con la página de
  error; si el render falla, responde 500 limpio.
- **CA-004-5** (plataforma) — Desde un clon limpio, `make gen`, `make css` y
  `make build` funcionan siguiendo el `README.md`, y el perfil `full` sirve la
  página con estilos y sin internet.

---

## Decisiones tomadas y descartadas

- **`ServeMux` de la biblioteca estándar en vez de `chi`.** Alcanza con los
  patrones de Go 1.22+ y evita una dependencia nueva. Si se cambia, se hace
  con su ADR.
- **Las rutas viven en `NuevoRouter`**, dentro de `httpadapter`, y no en
  `main.go`: así se testean sin levantar el servidor.
- **`Responder` renderiza a un buffer** antes de escribir, para poder devolver
  un 500 limpio si el render falla.
- **HTMX y CSS como archivos propios en `web/static/`**, para funcionar sin
  internet en la defensa.
- **Descartado:** caer a `html/template`. Solo se reconsidera si el spike
  supera el timebox de 4 horas (ADR 0001); el resultado se registra en un ADR.

## Uso de IA en esta especificación

- [x] Use IA para: redacción de la especificación a partir del plan de T-004 —
  revisado por: Angelo Conforti.