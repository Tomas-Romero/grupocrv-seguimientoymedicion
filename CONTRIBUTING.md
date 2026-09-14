# Cómo trabajamos en este repositorio

Esta es la única fuente de verdad sobre dónde va cada archivo, cómo se llama cada
cosa y qué pasos se siguen para que un cambio llegue a `main`. Si algo de acá te
parece mal, se discute en la retro y se cambia este archivo — no se ignora en
silencio.

**Leelo entero una vez antes de tu primer commit.** Después consultalo salteado.

---

## 0. Puesta en marcha (una sola vez)

```bash
git clone git@github.com:Tomas-Romero/metrics-estimation.git
cd metrics-estimation

cp .env.example .env         # y exportá DATABASE_URL en tu shell
make tools                   # instala templ, goose y air
make up                      # levanta Postgres, migra y carga datos de ejemplo
make check                   # tiene que dar todo verde antes de tocar nada
make dev                     # http://localhost:8080
```

Si `make check` no pasa en un repo recién clonado, **eso es un defecto** y se
reporta como `DEF-NNN` antes de seguir. No lo arregles en tu rama de feature.

---

## 1. Dónde va cada cosa

La regla que ordena todo: **`internal/domain` no importa nada de afuera.** Ni
base de datos, ni HTTP, ni templates, ni librerías de terceros (salvo la
biblioteca estándar). Si estás por agregar un `import` de `pgx` dentro de
`domain`, algo está mal ubicado.

```
cmd/server/main.go          Arranque. Lee config, arma dependencias, levanta el server.
                            Es el único lugar donde se "conecta todo".

internal/domain/<area>/     Reglas de negocio puras. Structs, validaciones, cálculos.
                            Acá vive lo que se evalúa con TDD.
                            Sin SQL. Sin HTTP. Sin logs. Sin efectos.

internal/app/               Casos de uso. Orquestan dominio + repositorios,
                            manejan transacciones. Un archivo por caso de uso.

internal/adapters/postgres/ Implementaciones de los repositorios. SQL acá y en
                            ningún otro lado.
internal/adapters/http/     Handlers, router, middlewares. Traducen HTTP ↔ casos de uso.
                            Un handler NO tiene lógica de negocio: valida, llama, responde.
internal/adapters/report/   Generación del PDF.

internal/platform/          Config, logger, conexión a la base, tipos de error comunes.

web/templates/              Archivos .templ. Un archivo por página o parcial.
web/static/                 CSS compilado y JS (Chart.js).

specs/                      Especificaciones SDD. Un archivo por historia.
features/                   Escenarios BDD en Gherkin + sus step definitions.
migrations/                 SQL versionado con goose. Nunca se edita una ya mergeada.
docs/adr/                   Decisiones de arquitectura.
docs/retros/                Actas de retrospectiva.
docs/plantillas/            Plantillas de spec, escenario y acta. Se copian, no se editan.
scripts/                    Automatización del repo.
```

### Cómo se llaman los archivos

| Qué | Convención | Ejemplo |
|---|---|---|
| Paquete Go | una palabra, minúscula, singular | `metrics`, `sprint`, `defect` |
| Archivo Go | `snake_case.go`, nombrado por el concepto | `velocity.go`, `planning_poker.go` |
| Test | mismo nombre + `_test.go` | `velocity_test.go` |
| Plantilla templ | `snake_case.templ` | `dashboard_page.templ`, `backlog_row.templ` |
| Spec SDD | `US-NNN-slug-corto.md` | `US-026-velocidad-equipo.md` |
| Escenario BDD | `US-NNN-slug-corto.feature` | `US-026-velocidad-equipo.feature` |
| Migración | `NNNNN_descripcion.sql` | `00004_tabla_defectos.sql` |
| ADR | `NNNN-titulo.md` | `0002-por-que-htmx.md` |

El `slug` es el mismo en la spec, el feature y la rama. Eso es lo que hace que
la trazabilidad se pueda seguir sin una planilla aparte.

---

## 2. El ciclo de trabajo de una historia

### Paso 1 — Tomá la historia del tablero

En el Sprint Board, mové el issue a **In Progress** y asignátelo. Si ya hay dos
issues tuyos en In Progress, terminá uno antes de empezar otro. El trabajo en
paralelo se ve lindo y no entrega nada.

### Paso 2 — Escribí la especificación (si toca reglas de negocio)

Copiá `docs/plantillas/spec-sdd.md` a `specs/US-NNN-slug.md`, completá los ocho apartados
y abrí un PR **solo con la spec**. Se mergea antes de escribir una línea de
código. Sí, es un PR aparte. Sí, vale la pena: es el 25% de la nota y la fecha de
los commits es la prueba de que lo hicimos en el orden correcto.

### Paso 3 — Creá la rama

```bash
git checkout main
git pull --rebase origin main
git checkout -b feat/US-026-velocidad-equipo
```

**Formato obligatorio:** `tipo/ID-slug-corto`

```
feat/US-026-velocidad-equipo        nueva funcionalidad
fix/DEF-007-horas-negativas         corrección de un defecto
refactor/US-026-extraer-summary     mejora sin cambio de comportamiento
test/US-026-casos-limite            solo tests
docs/US-026-spec                    solo documentación
chore/T-002-actualizar-golangci     mantenimiento
ci/T-002-umbral-cobertura           pipeline
```

Todo en minúsculas, palabras separadas por guiones, sin acentos ni ñ. El
workflow `pr-title.yml` rechaza el PR si la rama no cumple.

Una rama = una historia. Si mientras trabajás encontrás otra cosa para arreglar,
anotala como issue nuevo y seguí con lo tuyo.

### Paso 4 — Escribí los escenarios BDD

Copiá `docs/plantillas/escenario-bdd.feature` a `features/US-NNN-slug.feature`. Un escenario
por criterio de aceptación, etiquetado con `@CA-NNN-k`. Tienen que estar los
cuatro tipos que exige la guía: **normal, alternativo, límite y error.**

### Paso 5 — TDD, de verdad

Este es el paso que más nota vale y el único que no se puede simular después.

```bash
# 1. RED — escribí el test que falla
#    Corré los tests y confirmá que falla POR LA RAZÓN CORRECTA.
go test ./internal/domain/metrics/ -run TestVelocidad -v
git add internal/domain/metrics/velocity_test.go
git commit -m "test(metrics): RED velocidad del equipo sobre sprints cerrados [US-026]"

# 2. GREEN — la implementación más simple que hace pasar el test
#    Sin generalizar de más. Sin "ya que estoy".
git add internal/domain/metrics/velocity.go
git commit -m "feat(metrics): GREEN calculo de velocidad del equipo [US-026]"

# 3. REFACTOR — limpiá con los tests en verde
git commit -m "refactor(metrics): extraer SprintSummary y eliminar duplicacion [US-026]"
```

Repetí el ciclo por cada regla de negocio y cada caso límite de la spec. Un PR
sano tiene varios ciclos, no uno solo.

> **La regla del Agile Enabler:** no se mergea un PR de reglas de negocio si no
> hay un commit `test(...)` **anterior** al `feat`/`fix` correspondiente. En la
> defensa vamos a abrir el historial y mostrarlo. Si escribiste el código
> primero, no lo disfraces reordenando commits: avisá en la retro y lo hacemos
> distinto la próxima.

### Paso 6 — Antes de pushear

```bash
make check     # lint + cobertura con umbrales + BDD. Lo mismo que corre CI.
```

Si `make check` falla en tu máquina, va a fallar en CI. Arreglalo antes de
gastarle el tiempo a quien te revisa.

### Paso 7 — Push

```bash
# La primera vez en la rama
git push -u origin feat/US-026-velocidad-equipo

# Las siguientes
git push
```

**Nunca** `git push origin main`. **Nunca** `git push --force` sobre una rama que
otro esté mirando (si necesitás reescribir tu propia rama, usá
`git push --force-with-lease`).

Si `main` avanzó mientras trabajabas:

```bash
git fetch origin
git rebase origin/main        # rebase, no merge: el historial queda lineal
# resolvés conflictos si los hay
git push --force-with-lease
```

### Paso 8 — Abrí el PR

**Título con el mismo formato que los commits, terminando en el ID:**

```
feat(metrics): calculo de velocidad del equipo [US-026]
```

Completá la plantilla entera. Los tres puntos que más se olvidan y que hacen que
el PR vuelva:

1. `Closes #26` en el cuerpo — sin eso el issue no se cierra y el tablero miente.
2. Los commits del ciclo TDD pegados en la sección de evidencia.
3. La declaración de uso de IA.

Movés el issue a **In Review** y pedís revisión a quien no escribió el código.

### Paso 9 — Revisión

Quien revisa mira, en este orden:

1. ¿Hay `RED` antes de `GREEN` en el historial?
2. ¿Los tests cubren los casos límite y los errores que dice la spec?
3. ¿El dominio quedó libre de dependencias externas?
4. ¿Los errores se devuelven envueltos con contexto, sin `panic` ni errores ignorados?
5. ¿Puede el autor explicar cualquier línea? (preguntá una al azar — es literalmente lo que va a pasar en la defensa)

Comentarios directos y sobre el código, no sobre la persona. Quien revisa aprueba
o pide cambios; no mergea por el otro.

### Paso 10 — Merge

**Squash merge** siempre, con el título del PR como mensaje del commit. GitHub
borra la rama sola. El issue se cierra solo. El item del tablero pasa a **Done**
solo.

```bash
git checkout main
git pull --rebase origin main
```

---

## 3. Convención de commits

```
tipo(alcance): descripcion en imperativo [ID]
```

| Parte | Regla |
|---|---|
| `tipo` | `feat` `fix` `test` `refactor` `docs` `chore` `ci` `perf` |
| `alcance` | el paquete o área tocada: `metrics`, `sprint`, `ui`, `db`, `workflows` |
| `descripcion` | minúscula, imperativo, sin punto final, sin acentos, máximo ~70 caracteres |
| `[ID]` | `[US-026]`, `[DEF-007]`, `[T-002]` — obligatorio |

En los commits del ciclo TDD, agregá `RED` o `GREEN` al principio de la
descripción. Es lo que convierte el historial en evidencia.

**Ejemplos buenos**

```
test(metrics): RED velocidad con cero sprints cerrados [US-026]
feat(metrics): GREEN calculo de velocidad del equipo [US-026]
refactor(metrics): extraer SprintSummary [US-026]
fix(effort): rechazar horas negativas y mayores a 24 [DEF-007]
docs(specs): especificacion de velocidad del equipo [US-026]
ci(workflows): fallar si la cobertura del dominio baja de 85 [T-002]
chore(deps): actualizar pgx a v5.7.1 [T-001]
```

**Ejemplos malos y por qué**

```
arreglos varios                    sin tipo, sin ID, no dice nada
feat: cambios                      "cambios" no es una descripción
Feat(Metrics): Velocidad.          mayúsculas y punto final
WIP                                no se commitea trabajo a medias en main
feat(metrics): agregue la velocidad y de paso arregle el dashboard y toque el css
                                   son tres commits, no uno
```

**Un commit = un cambio con sentido.** Si al escribir el mensaje necesitás la
palabra "y", probablemente sean dos commits.

---

## 4. Cómo se escribe el código Go

No inventamos estilo: seguimos [Effective Go](https://go.dev/doc/effective_go) y
lo que diga `golangci-lint`. Lo que sí es nuestro:

**Errores.** Se devuelven, no se ignoran ni se hace `panic`. Se envuelven con
contexto usando `%w`:

```go
if err != nil {
    return fmt.Errorf("calcular velocidad del proyecto %s: %w", id, err)
}
```

Los errores de dominio se definen como valores del paquete para poder compararlos
con `errors.Is`:

```go
var ErrSprintSinCerrar = errors.New("el sprint no esta cerrado")
```

**Dominio sin efectos.** Nada de `time.Now()` adentro de una función de dominio
— la fecha se pasa como parámetro. Si no, el test depende del día en que corra.

**Tests con casos en tabla.** Es lo idiomático en Go y hace obvio qué casos
límite faltan:

```go
func TestVelocidad(t *testing.T) {
    casos := []struct {
        nombre   string
        sprints  []Sprint
        esperado float64
    }{
        {"sin sprints cerrados", nil, 0},
        {"un solo sprint", []Sprint{{Completados: 20}}, 20},
        {"promedio de tres", []Sprint{{Completados: 20}, {Completados: 30}, {Completados: 25}}, 25},
    }
    for _, c := range casos {
        t.Run(c.nombre, func(t *testing.T) {
            require.InDelta(t, c.esperado, Velocidad(c.sprints), 0.001)
        })
    }
}
```

**Comentá el porqué, no el qué.** `// suma los puntos` sobrando; `// RN-026-2: los
sprints cancelados no cuentan para la velocidad` vale oro en la defensa.

**Ligá el test a la trazabilidad.** Un comentario arriba de cada test con el
identificador:

```go
// US-026 / CA-026-3: con cero sprints cerrados la velocidad es 0, no un error.
```

---

## 5. Definition of Ready y Definition of Done

Copiadas del plan. Si una historia no cumple DoR, no entra al sprint. Si no
cumple DoD, no se cuenta como completada aunque "funcione".

**Ready**

- Redactada como Como / quiero / para
- Criterios de aceptación numerados `CA-NNN-k`
- Spec SDD mergeada si toca reglas de negocio
- Estimada en Story Points por el equipo
- Dependencias identificadas y resueltas
- Entra en un sprint (si es mayor a 8 SP, se parte)

**Done**

- Código en `main` vía PR aprobado por otro integrante
- Ciclo RED → GREEN → REFACTOR visible en el historial
- Escenarios BDD automatizados y en verde
- Cobertura del dominio ≥ 85%, CI completo en verde
- Sin hallazgos de `golangci-lint`
- Errores manejados y validaciones de entrada cubiertas
- Spec SDD actualizada si la implementación la contradijo
- Issue cerrado desde el PR y movido a Done
- Demostrable en la Review sin pasos manuales ocultos

---

## 6. La Daily

Todos los días, asíncrono, antes de las 22:00. Formato fijo, tres líneas:

```
Ayer:     US-026 — terminé el cálculo de velocidad, falta el caso de sprints cancelados
Hoy:      US-026 — casos límite y abro el PR
Bloqueos: ninguno   (o: espero que se mergee #24 para poder seguir)
```

Un bloqueo escrito es responsabilidad del Agile Enabler destrabarlo. Un bloqueo
callado es tuyo.

---

## 7. Qué hacer cuando

**Rompiste `main`.** Avisá en el grupo primero, arreglá después. Un `revert` es
más rápido y más seguro que un fix apurado.

**Tu rama tiene conflictos con `main`.** `git fetch origin && git rebase
origin/main`. Nunca `git merge main` adentro de tu rama: ensucia el historial y
complica la lectura en la defensa.

**Commiteaste en `main` por accidente.**

```bash
git branch feat/US-0NN-lo-que-sea    # guardás el trabajo en una rama
git reset --hard origin/main         # limpiás main
git checkout feat/US-0NN-lo-que-sea
```

**Necesitás guardar algo a medias.** `git stash` o commiteá en tu rama con un
mensaje honesto. No hay problema: el squash merge convierte tu rama entera en un
solo commit limpio.

**Encontraste un bug de otro.** Issue `DEF-NNN` con el template de defecto. No lo
arregles de paso en tu PR — se pierde la trazabilidad y además necesitamos el
defecto registrado para las métricas (el requerimiento 6 de la guía es
literalmente eso).

**No sabés cómo hacer algo.** Preguntá en el grupo antes de quemar dos horas. El
timebox informal del equipo es 45 minutos: si después de eso seguís trabado, es
un bloqueo y va a la Daily.
