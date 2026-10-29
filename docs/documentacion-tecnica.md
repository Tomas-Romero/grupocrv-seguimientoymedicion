# Documentación técnica

Este documento explica cómo está hecho **Software Metrics & Estimation**: qué resuelve, cómo se organiza
el código, cómo se guardan los datos, qué decisiones se tomaron y por qué, cómo se ejecuta y cómo se
trabajó. Está pensado para alguien que no conoce el proyecto. Para usar la aplicación, ver el
[manual de usuario](manual-de-usuario.md); para los números de calidad, el
[informe de métricas y cobertura](informe-metricas-cobertura.md).

## 1. Qué es el sistema

Es una aplicación web para administrar proyectos de software con Scrum y obtener información de
estimación, planificación, seguimiento y calidad. Un equipo crea su proyecto, registra a sus
integrantes, arma el Product Backlog, planifica Sprints, estima las historias con Planning Poker,
registra las horas que trabajó y los defectos que encontró, y la aplicación calcula las métricas del
proceso (Story Points, velocidad, horas estimadas contra reales, porcentaje de completitud y defectos
por sprint).

Es el Trabajo Práctico Integrador de **Ingeniería y Calidad de Software** (UTN FRSR, 2026). La cátedra
exige que el núcleo y las reglas de negocio estén en Go y que el desarrollo siga SDD, BDD y TDD.

### Requerimientos de la cátedra que cubre

La guía de la cátedra tiene nueve requerimientos funcionales. Estado en `main` al 28/10/2026:

| # | Requerimiento | Historias | Estado |
|---|---|---|---|
| 1 | Gestión de proyectos e integrantes, y estado general del proyecto | US-001 a US-004, T-011 | Hecho |
| 2 | Product Backlog: ítems con prioridad, estado, Story Points y criterios de aceptación | US-005, US-006, US-013, T-011 | Hecho |
| 3 | Sprints: Sprint Goal, fechas, asignación de historias, cierre y sprints anteriores | US-008 a US-012, T-015, T-016 | Hecho |
| 4 | Estimación con Planning Poker: votación oculta, revelado, divergencias, nueva ronda y estimación acordada | US-014 a US-018, T-020 | Hecho |
| 5 | Registro de esfuerzo por integrante, fecha y actividad, y comparación estimado contra real | US-019 a US-021, T-019 | Hecho |
| 6 | Defectos con severidad, historia relacionada, sprint de detección y de resolución | US-022 a US-024, T-021 | Hecho; el filtro por sprint y severidad (US-024) tiene spec y falta implementarlo |
| 7 | Métricas: Story Points, velocidad, horas, desviación, completitud y defectos por sprint | US-025 a US-029 | Cálculos hechos en el dominio; se ven en el dashboard (requerimiento 8) |
| 8 | Dashboard con representaciones gráficas | US-030 | En desarrollo: tiene spec, falta la pantalla |
| 9 | Reportes de proyecto o sprint, exportables a PDF | US-033, US-034 | En desarrollo: tiene spec, falta la pantalla |

El detalle de cada historia, con sus criterios, escenarios y tests, está en la
[matriz de trazabilidad](trazabilidad.md).

## 2. Arquitectura

### Capas

El código sigue una **arquitectura en capas con el dominio aislado** (también llamada hexagonal o de
puertos y adaptadores). Las dependencias apuntan siempre hacia adentro: el dominio no conoce a nadie,
la aplicación conoce al dominio, y los adaptadores conocen a la aplicación.

```mermaid
flowchart TB
    nav[Navegador<br/>HTML + HTMX]
    subgraph adaptadores [Adaptadores]
        http[http<br/>internal/adapters/http<br/>router, handlers, errores T-005]
        web[web<br/>web/templates<br/>plantillas templ]
        pg[postgres<br/>internal/adapters/postgres<br/>pgx, transacciones]
        mem[memoria<br/>internal/adapters/memoria<br/>base en memoria]
    end
    app[Aplicación<br/>internal/app<br/>casos de uso y puertos]
    dom[Dominio<br/>internal/domain<br/>reglas de negocio puras]
    db[(PostgreSQL 16)]
    bdd[Escenarios BDD<br/>features/]
    main[cmd/server<br/>cableado]

    nav -->|HTTP| http
    http --> web
    http --> app
    app --> dom
    pg -.implementa los puertos.-> app
    mem -.implementa los puertos.-> app
    pg --> db
    bdd --> app
    bdd --> mem
    main --> http
    main --> pg
```

| Capa | Dónde | Qué hace | De qué depende |
|---|---|---|---|
| **Dominio** | `internal/domain/` (paquetes `proyecto`, `integrante`, `backlog`, `sprint`, `planningpoker`, `esfuerzo`, `defecto`, `metricas`) | Las reglas de negocio: qué es un ítem válido, cuándo se puede cerrar un sprint, cómo se calcula la velocidad, que un voto no se vea antes de revelarse. Funciones puras que reciben datos y devuelven resultados o errores | Solo la biblioteca estándar de Go (`errors`, `fmt`, `time`, `slices`…). Ningún paquete importa base de datos, HTTP ni bibliotecas externas |
| **Aplicación** | `internal/app/` | Un caso de uso por acción del usuario (`CrearItemBacklog`, `CerrarSprint`, `VotarEnSesion`…). Valida que existan las cosas referenciadas, llama al dominio y guarda. Declara los **puertos**: las interfaces de repositorio que necesita | Dominio |
| **Adaptador memoria** | `internal/adapters/memoria/` | Implementa los puertos con mapas en memoria. Lo usan los escenarios BDD y los tests de las pantallas | Aplicación y dominio |
| **Adaptador postgres** | `internal/adapters/postgres/` | Implementa los mismos puertos contra PostgreSQL con `pgx`: consultas, transacciones y candados | Aplicación, dominio y `pgx` |
| **Adaptador http** | `internal/adapters/http/` | El router (`net/http` de la biblioteca estándar), los handlers de cada pantalla, el catálogo de errores y `Responder`, que decide si devolver la página completa o solo un fragmento para HTMX | Aplicación, dominio y `web/templates` |
| **web** | `web/templates/` y `web/static/` | Las plantillas `templ` (HTML tipado que se compila a Go), los estilos de Tailwind y HTMX | Solo tipos de vista propios |
| **Plataforma** | `internal/platform/` | Configuración por variables de entorno, conexión a la base, migraciones y el tipo de error uniforme de T-005 | Biblioteca estándar, `pgx` y `goose` |
| **Cableado** | `cmd/server/main.go` | El único lugar donde se eligen los adaptadores: arma los repositorios de Postgres, los casos de uso y el router | Todo lo anterior |

### Por qué así

- **El dominio se prueba sin base de datos.** Es la condición para sostener TDD durante todo el
  proyecto: un test de dominio corre en milisegundos y no necesita Docker. Hoy el dominio tiene 190
  funciones de test y una cobertura del 99,8 % (ver el [informe](informe-metricas-cobertura.md)).
- **La misma regla se aplica en todos lados.** Cuando dos personas piden lo mismo a la vez, el
  repositorio de Postgres bloquea la fila, vuelve a leer y llama **a la misma función del dominio**
  para verificar; no hay una copia de la regla en SQL.
- **Dos implementaciones de cada puerto.** La base en memoria permite que los escenarios BDD y los
  tests de las pantallas corran rápido y sin infraestructura, y la de Postgres se prueba aparte contra
  una base real, con los mismos casos que la de memoria.
- **Todo el código evaluable es Go.** La interfaz se renderiza en el servidor con `templ`; HTMX agrega
  interactividad sin escribir JavaScript propio. Un solo lenguaje, un solo build, un solo despliegue
  ([ADR 0001](adr/0001-stack-tecnologico.md)).

### Cómo viaja un pedido

Ejemplo: alguien vota 5 en una sesión de Planning Poker.

```mermaid
sequenceDiagram
    actor U as Usuario
    participant H as http (handler)
    participant A as app.VotarEnSesion
    participant R as postgres (repositorio)
    participant D as dominio (planningpoker)
    participant B as PostgreSQL

    U->>H: POST /proyectos/{id}/planning-poker/{sesion}/votos
    H->>A: Ejecutar(sesión, integrante, 5)
    A->>R: ObtenerSesion y ExisteIntegranteEnProyecto
    A->>D: VerificarVoto(sesión, integrante, 5)
    D-->>A: ok o errores unidos (fuera de la escala, no participa, votación cerrada)
    A->>R: RegistrarVoto(sesión, integrante, 5)
    R->>B: BEGIN, SELECT ... FOR NO KEY UPDATE OF s (bloquea la sesión)
    R->>D: VerificarVoto con lo que leyó bajo el candado
    R->>B: INSERT ... ON CONFLICT DO UPDATE (un segundo voto reemplaza al primero), COMMIT
    R-->>A: ok
    A-->>H: ok
    H-->>U: HX-Redirect a la sesión (o 422/409 con el mensaje)
```

Si algo falla, el handler no arma el mensaje: lo busca en el **catálogo de errores**
(`internal/adapters/http/catalogo.go`), que asigna a cada error del dominio un tipo (validación 422,
conflicto 409, no encontrado 404) y el campo del formulario al que corresponde (spec
[T-005](../specs/T-005-validaciones-errores-uniformes.md)). Un test lee el código fuente y falla si
un error exportado del dominio o de la aplicación no está en el catálogo.

## 3. Modelo de datos

Las tablas se crean con las migraciones de `migraciones/` (goose, en SQL, numeradas de `00001` a
`00013`). La aplicación las aplica sola al arrancar. Los datos de ejemplo están en
`migraciones/datos/ejemplo.sql`.

```mermaid
erDiagram
    proyectos ||--o{ integrantes : tiene
    proyectos ||--o{ items_backlog : tiene
    proyectos ||--o{ sprints : tiene
    proyectos ||--o{ registros_esfuerzo : tiene
    proyectos ||--o{ defectos : tiene
    items_backlog ||--o{ sprint_items : "estuvo en"
    sprints ||--o{ sprint_items : "contuvo"
    items_backlog ||--o{ registros_esfuerzo : "se trabajó en"
    integrantes ||--o{ registros_esfuerzo : "registró"
    items_backlog ||--o{ sesiones_planning_poker : "se estima en"
    sesiones_planning_poker ||--o{ participantes_planning_poker : tiene
    integrantes ||--o{ participantes_planning_poker : participa
    participantes_planning_poker ||--o{ votos_planning_poker : vota
    items_backlog ||--o{ defectos : "tiene"
    sprints ||--o{ defectos : "detectado en"
    sprints |o--o{ defectos : "resuelto en"

    proyectos {
        uuid id PK
        text nombre
        text descripcion
        date fecha_inicio
        date fecha_fin
    }
    integrantes {
        uuid id PK
        uuid proyecto_id FK
        text nombre
        text email "único por proyecto"
        text rol "product_architect, agile_enabler, product_builder"
    }
    items_backlog {
        uuid id PK
        uuid proyecto_id FK
        int numero "único por proyecto"
        text titulo
        text prioridad "must, should, could, wont"
        text estado "pendiente, en_progreso, completado"
        int story_points "nulo si no está estimado"
        text_array criterios
        timestamptz completado_en
    }
    sprints {
        uuid id PK
        uuid proyecto_id FK
        int numero
        text sprint_goal
        date fecha_inicio
        date fecha_fin
        text estado "abierto, cerrado"
    }
    sprint_items {
        uuid item_id PK,FK
        uuid sprint_id PK,FK
        text resultado "pendiente, completado, no_completado"
    }
    registros_esfuerzo {
        uuid id PK
        uuid item_id FK
        uuid integrante_id FK
        date fecha
        text actividad
        numeric horas "más de 0 y hasta 24"
    }
    sesiones_planning_poker {
        uuid id PK
        uuid item_id FK
        text estado "votando, revelada, finalizada"
        int ronda
        int valor_acordado "solo si está finalizada"
    }
    participantes_planning_poker {
        uuid sesion_id PK,FK
        uuid integrante_id PK,FK
    }
    votos_planning_poker {
        uuid sesion_id PK,FK
        int ronda PK
        uuid integrante_id PK,FK
        int valor
    }
    defectos {
        uuid id PK
        uuid item_id FK
        uuid sprint_deteccion_id FK
        uuid sprint_resolucion_id FK "solo si está resuelto"
        text descripcion
        text severidad "critica, alta, media, baja"
        text estado "abierto, en_correccion, resuelto, descartado"
    }
```

Lo que la base garantiza por sí misma (además de lo que valida el dominio):

- **Un solo sprint abierto por proyecto** (índice único parcial `un_sprint_abierto_por_proyecto`).
- **Una sola sesión de Planning Poker abierta por historia** (`una_sesion_abierta_por_item`).
- **Solo vota quien participa de la sesión:** la clave foránea de `votos_planning_poker` apunta a
  `participantes_planning_poker`, no a `integrantes`.
- **Coherencia de estados:** una historia completada tiene fecha de completado; una sesión finalizada
  tiene valor acordado; un defecto resuelto tiene sprint de resolución, y solo él (`CHECK`).

### La relación histórica entre sprint e ítem

`sprint_items` no dice "en qué sprint está la historia" sino **cada vez que una historia estuvo en un
sprint y cómo terminó** (`pendiente`, `completado` o `no_completado`). Al cerrar un sprint, lo no
completado vuelve al backlog, pero la fila **no se borra**: queda con `no_completado`, y si la historia
se asigna a otro sprint se agrega una fila nueva. Un índice único parcial asegura que una historia
tenga a lo sumo una asignación `pendiente` a la vez.

**Por qué:** sin esa historia se perdían los Story Points planificados de los sprints cerrados. La
velocidad y el resumen de un sprint (US-025 y US-026) cuentan como planificado también lo que no se
completó, y el reporte de un sprint necesita saber qué tenía. La primera versión de la spec de US-009
borraba la relación al cerrar; la revisión cruzada lo detectó antes de implementarlo y se corrigió
(T-013). Más tarde, DEF-007 mostró que "completado" también tiene que ser por sprint: una historia no
completada en el Sprint 3 y completada en el 4 cuenta como completada solo en el 4. Ver
[US-009, sección "Decisiones"](../specs/US-009-asignar-historias-sprint.md) y
[US-011](../specs/US-011-cerrar-sprint.md).

## 4. Decisiones clave

| Decisión | Qué se decidió | Por qué | Dónde está |
|---|---|---|---|
| Stack | Aplicación web renderizada en el servidor, toda en Go: `templ`, HTMX, Tailwind, PostgreSQL 16 con `pgx`, migraciones con `goose` | La cátedra pide Go en el núcleo; un solo lenguaje concentra la calidad evaluada en un lugar y el dominio queda aislado | [ADR 0001](adr/0001-stack-tecnologico.md) |
| BDD con godog | Los `.feature` se ejecutan como un test de Go (`go test ./features/`), en modo estricto, contra la capa de aplicación con la base en memoria | Un paso sin implementar no puede pasar en verde; los escenarios prueban reglas, no la base ni el HTML | [ADR 0002](adr/0002-integracion-bdd-godog.md), [diccionario de frases](diccionario-steps.md) |
| Idioma del dominio | Dominio, tablas, specs, paquetes y archivos en español sin acentos en los identificadores; los términos de Scrum en inglés (`SesionPlanningPoker`, `SprintGoal`) | La trazabilidad spec → código queda literal: el término de la spec es el identificador que se busca | [ADR 0003](adr/0003-idioma-del-dominio.md) |
| Errores unidos | El dominio valida **todo** y devuelve todos los errores juntos con `errors.Join`, en un orden fijo; cada uno se reconoce con `errors.Is`. La aplicación los envuelve con contexto (`%w`) | Quien completa un formulario ve todo lo que tiene que corregir de una vez, y los tests pueden comprobar cada error por separado | [US-001, RN-001-7](../specs/US-001-crear-proyecto.md), [T-005](../specs/T-005-validaciones-errores-uniformes.md) |
| Errores uniformes | Un catálogo une cada error con su tipo (422, 409 o 404) y su campo; un fallo inesperado nunca muestra detalles internos | Todas las pantallas responden igual; un test falla si alguien agrega un error y no lo registra | [T-005](../specs/T-005-validaciones-errores-uniformes.md) |
| Candados `FOR NO KEY UPDATE` | Las operaciones que leen, verifican y guardan (asignar, marcar, cerrar, votar, revelar, acordar, nueva ronda) corren en una transacción que bloquea la fila con `SELECT ... FOR NO KEY UPDATE` y vuelve a verificar con la misma función del dominio | Dos pedidos simultáneos no pueden dejar datos inconsistentes (por ejemplo, abrir dos rondas). Es `NO KEY UPDATE` y no `UPDATE` porque no bloquea a quien inserta una fila que apunta a la bloqueada, como un registro de esfuerzo | [US-009, "Decisiones"](../specs/US-009-asignar-historias-sprint.md), [US-008](../specs/US-008-crear-sprint.md), [US-018](../specs/US-018-registrar-estimacion-acordada.md) |
| Voto oculto en el dominio | Los votos de una ronda son un tipo (`planningpoker.Votacion`) que no exporta sus valores: expone quién votó y quién falta, y `Valores()` devuelve `ErrVotosOcultos` mientras se vota. Ningún puerto entrega los votos en crudo | Un error en una pantalla o en un adaptador no puede mostrar un voto antes de tiempo: la regla la impone el compilador, no la disciplina | [US-015, RN-015-7](../specs/US-015-voto-individual-oculto.md), [US-016, RN-016-12](../specs/US-016-revelar-estimaciones.md) |
| ID generado por la base | Los registros nuevos reciben su UUID de `gen_random_uuid()`; el dominio no lo genera | Un solo lugar decide los IDs y el dominio queda sin dependencias | [US-001](../specs/US-001-crear-proyecto.md), migraciones `00002` y `00005` |
| Migraciones fuera de orden | `goose` acepta una migración con número menor que la última aplicada | Varias ramas agregan migraciones en paralelo y pueden llegar a `main` en otro orden | T-012 (#91), `internal/platform/db/migrar.go` |
| Matriz de trazabilidad generada | `docs/trazabilidad.md` se genera con `make trace` a partir de las specs, los `.feature` y los comentarios de los tests | Se mantiene sola y no se desactualiza; los conflictos se resuelven regenerándola | T-007, `cmd/trace/` |

**Diferencias con el ADR 0001.** El ADR eligió `chi` como router, Chart.js para los gráficos y `maroto`
para el PDF. En la práctica se usa el `ServeMux` de la biblioteca estándar (desde Go 1.22 acepta
métodos y parámetros en las rutas, y no hizo falta otra dependencia); el PDF se resolvió con una vista
imprimible del navegador, que la cátedra aceptó el 26/10 (US-034); y los gráficos son del dashboard
(US-030), todavía en desarrollo.

## 5. Cómo correrlo

### Para evaluarlo (perfil `full`)

Requisitos: Docker y Docker Compose.

```bash
git clone https://github.com/Tomas-Romero/grupocrv-seguimientoymedicion.git
cd grupocrv-seguimientoymedicion
docker compose --profile full up
```

La aplicación queda en <http://localhost:8080> con un proyecto de ejemplo cargado. El perfil `full`
construye la imagen, levanta PostgreSQL, aplica las migraciones y carga los datos de ejemplo
(`CARGAR_DATOS_EJEMPLO=true`).

### Para desarrollar

Requisitos: Go 1.23 o superior, Docker, GNU Make y `psql`. En Windows, todos los `make` se corren
desde Git Bash. El detalle (y los problemas conocidos de Windows) está en el [README](../README.md#para-desarrollar).

```bash
cp .env.example .env
export DATABASE_URL="postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable"
make tools   # templ, goose, air, golangci-lint y Tailwind, con versión fija (una sola vez)
make up      # PostgreSQL, migraciones y datos de ejemplo
make dev     # servidor con recarga automática en http://localhost:8080
```

| Variable | Para qué | Valor por defecto |
|---|---|---|
| `DATABASE_URL` | Conexión a PostgreSQL. Obligatoria | — |
| `HTTP_ADDR` | Dirección donde escucha el servidor | `:8080` |
| `LOG_LEVEL` | Se lee de la configuración, pero hoy el servidor no la usa: los logs salen con el nivel por defecto de `slog` | `info` |
| `CARGAR_DATOS_EJEMPLO` | Carga el proyecto de ejemplo al arrancar (es idempotente) | `false` |
| `DB_PORT` | Puerto de PostgreSQL en tu máquina, si el 5432 está ocupado | `5432` |

Los archivos `*_templ.go` se generan y no se versionan: si faltan, `make gen`.

### Pruebas

```bash
make test    # todos los tests de Go
make bdd     # los escenarios Gherkin (go test ./features/)
make cover   # tests con cobertura y umbrales: 85 % en el dominio y 70 % en internal/
make check   # todo lo que corre el CI: lint, tests, cobertura y BDD
```

Los tests de los repositorios de Postgres necesitan `DATABASE_URL`: cada uno crea una base temporal
propia, la migra y la borra al terminar (`internal/platform/db/dbprueba`), así que no tocan los datos
de desarrollo. Para correr un solo escenario:
`go test ./features/ -godog.tags=@CA-017-1`.

### CI/CD

Cada pull request pasa por GitHub Actions: **Lint** (`go vet`, formato, `golangci-lint` y `actionlint`),
**Tests y cobertura** (contra un PostgreSQL real, con los umbrales), **BDD**, **Trazabilidad** (cada spec
de historia tiene su `.feature`), **Perfil full** (levanta el `docker compose` completo sobre un volumen
vacío) y **Convenciones del PR** (título, rama, `Closes #N` y el ID en cada commit). `main` está
protegida: se necesitan los chequeos en verde, una aprobación de otra persona y se mergea con *Rebase
and merge* para conservar los commits RED y GREEN.

Cada merge a `main` que cambia la imagen la construye y, con las credenciales de la cátedra cargadas,
la publica en el registro `ingcalidad2026` y la despliega en Azure Container Instances; la base de datos
es externa (Neon). Para cuidar los créditos de la cuenta compartida, el contenedor queda apagado
después de comprobar `/health` y se enciende a mano. El detalle está en [`docs/ci-cd.md`](ci-cd.md).

## 6. Cómo se trabajó

### SDD → BDD → TDD

Cada historia con reglas de negocio siguió el mismo camino, descrito en
[`CONTRIBUTING.md`](../CONTRIBUTING.md#2-el-ciclo-de-trabajo-de-una-historia):

1. **SDD (especificación).** Antes del código se escribe y se mergea la spec en `specs/`, con la
   plantilla de `docs/plantillas/`: objetivo, entradas, salidas, reglas de negocio (`RN-…`),
   restricciones, casos límite (`CL-…`), errores y criterios de aceptación (`CA-…`). Es la Definition of
   Ready: sin spec mergeada, la historia no entra al sprint.
2. **BDD (escenarios).** Cada criterio de aceptación se escribe como escenario Gherkin en `features/`,
   con los cuatro tipos que pide la guía (normal, alternativo, límite y error) y etiquetado con su
   criterio (`@CA-017-1`). Los escenarios usan frases de un [diccionario compartido](diccionario-steps.md)
   y se commitean primero, en rojo.
3. **TDD (código).** Por capa (dominio, aplicación, memoria, Postgres, HTTP): un commit `test(...): RED`
   con el test que falla por la razón correcta, y después `feat(...): GREEN` con lo mínimo para pasarlo;
   los `refactor` van con los tests en verde. Con *Rebase and merge*, el ciclo queda en el historial de
   `main`: hoy hay 189 commits RED y 194 GREEN (`git log --oneline | grep -c ' RED '`).

### Trazabilidad

Cada test cita en su comentario los criterios y reglas que cubre (`// US-017 / CA-017-1, RN-017-5`).
`make trace` lee las specs, los `.feature` y esos comentarios, y genera
[`docs/trazabilidad.md`](trazabilidad.md): Historia → Spec → Criterios → Escenarios BDD → Tests → Código.
Cada commit lleva el ID de su historia (`[US-017]`), y un chequeo del CI lo exige (T-014). El resumen de
la matriz está en el [informe de métricas](informe-metricas-cobertura.md#3-trazabilidad).

### Gestión

Scrum con sprints semanales que cierran los lunes (Sprint 0 más cinco sprints, hasta la presentación
del 02/11). Cada sprint tiene su acta de retrospectiva en [`docs/retros/`](retros/), y el plan, los
responsables y la lista de recorte están en el [plan de trabajo](plan-de-trabajo.md). Las revisiones
siguen una rotación (a Tomás lo revisa Vergara, a Vergara lo revisa Conforti y a Conforti lo revisa
Tomás) y los cambios de dominio necesitan además a un dueño del código (`.github/CODEOWNERS`).

### Uso de IA

La cátedra permite usar IA y nos hace responsables de todo el código incorporado. La política está en
[`docs/uso-de-ia.md`](uso-de-ia.md): se declara en cada PR qué parte se generó con asistencia y quién
la revisó; quien revisa puede pedir que se explique cualquier función; los casos límite y de error
salen de la spec, no del modelo; y sin spec no se genera código de dominio. Se usaron Claude / Claude
Code (specs, código, tests, revisión y documentación) y GitHub Copilot (autocompletado). La declaración
de cada cambio está en la sección "Uso de IA" de su pull request.

## 7. Dónde está cada cosa

| Qué | Dónde |
|---|---|
| Especificaciones SDD | [`specs/`](../specs/) |
| Escenarios BDD y sus pasos | [`features/`](../features/), [`features/steps/`](../features/steps/) |
| Dominio, aplicación y adaptadores | [`internal/`](../internal/) |
| Servidor y generador de la matriz | [`cmd/server/`](../cmd/server/), [`cmd/trace/`](../cmd/trace/) |
| Migraciones y datos de ejemplo | [`migraciones/`](../migraciones/) |
| Plantillas y estáticos | [`web/`](../web/) |
| Workflows de CI/CD | [`.github/workflows/`](../.github/workflows/) |
| ADRs | [`docs/adr/`](adr/) |
| Actas de retrospectiva | [`docs/retros/`](retros/) |
| Convenciones de trabajo | [`CONTRIBUTING.md`](../CONTRIBUTING.md) |
