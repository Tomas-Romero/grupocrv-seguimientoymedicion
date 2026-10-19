# T-005 — Validaciones de formulario y manejo uniforme de errores

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás |
| **Issue** | #15 |
| **Escenarios BDD** | ninguno: es una tarea técnica (ver "Decisiones") |
| **Código** | `internal/platform/errores/` y `internal/adapters/http/` |
| **Última actualización** | 2026-10-18 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Que toda pantalla y todo endpoint respondan los errores de la misma forma: una
persona que se equivoca en un formulario ve qué campo corregir y por qué; una
referencia a algo que no existe se responde como tal; y un fallo del sistema
nunca muestra detalles internos.

Hoy cada caso de uso devuelve sus errores como valores (`errors.Is`, a veces
unidos con `errors.Join`) y el servidor no sabe qué hacer con ellos. Esta tarea
pone el puente entre esos errores y la respuesta HTTP, antes de que las
pantallas de T-011 y las historias siguientes empiecen a necesitarlo.

## 2. Entradas

Cuatro piezas. Ninguna accede a la base. El `ManejadorDeErrores` se arma una vez
con `NuevoManejadorDeErrores(catalogo, logger)`; sin logger usa el de por defecto.

| Pieza | Entrada | Obligatoria | Detalle |
|---|---|---|---|
| `Catalogo.Clasificar` | `err error` | No (puede ser `nil`) | Cualquier error que devuelva un caso de uso o un parseo de formulario |
| `ManejadorDeErrores.ResponderError` | `w`, `r`, `err` | `w` y `r` sí | Usa el `Accept` de la petición para elegir el formato |
| `Formulario` | `url.Values` (el `r.PostForm` o `r.Form`) | Sí | Se consulta campo por campo: `Entero`, `Decimal`, `Fecha` |
| `ManejadorDeErrores.Recuperar` | `http.Handler` | Sí | Middleware: envuelve al handler que sigue |

## 3. Salidas esperadas

**`errores.Problema`** — el resultado de clasificar un error:

```
Tipo    Tipo            // validacion | no_encontrado | conflicto | interno
Errores []ErrorDeCampo  // en el orden en que vinieron; nunca vacío si Tipo no es vacío
ErrorDeCampo { Campo string; Mensaje string }  // Campo vacío = error general
```

**Estado HTTP** de cada tipo: `validacion` 422, `no_encontrado` 404,
`conflicto` 409, `interno` 500.

**Respuesta en JSON** (cuando el `Accept` pide JSON):

```json
{"tipo": "validacion", "errores": [{"campo": "titulo", "mensaje": "el titulo es obligatorio"}]}
```

`campo` se omite cuando el error es general.

**Respuesta en HTML** (todo lo demás, incluido HTMX): un fragmento sin estilos,
escapado con `html/template`:

```html
<div role="alert" data-tipo="validacion"><ul><li data-campo="titulo">el titulo es obligatorio</li></ul></div>
```

`data-campo` se omite cuando el error es general. El estilo lo pone T-004; acá
solo se fija la estructura.

**Formulario**: cada consulta devuelve el valor ya convertido (`int`, `float64`,
`time.Time`) y `Err()` devuelve un `error` que es un `*errores.ErrorDeValidacion` con los
campos mal formados, o `nil` si no hubo ninguno (un `nil` de verdad, no un
puntero nulo dentro de la interfaz).

## 4. Reglas de negocio

**Tipos de error**

- **RN-005-1** — Hay cuatro tipos y cada uno tiene su estado HTTP: `validacion`
  (422), `no_encontrado` (404), `conflicto` (409) e `interno` (500).
- **RN-005-2** — Qué tipo es cada error se decide por esta regla, no caso por
  caso: es `validacion` si depende solo de los valores recibidos (y del reloj,
  como una fecha futura); es `conflicto` si depende del estado que ya hay
  guardado (un sprint cerrado, una historia ya asignada, un email repetido);
  es `no_encontrado` si apunta a algo que no existe; y es `interno` todo lo
  demás.
- **RN-005-3** — Un error se reconoce con `errors.Is` contra un **catálogo**
  explícito. Nunca se mira el texto del mensaje.

**Clasificar**

- **RN-005-4** — Un error unido con `errors.Join` (o con varios `%w`) se separa
  en sus partes, también si hay uniones anidadas. Cada parte se clasifica por
  separado y los mensajes salen en el mismo orden en que vinieron.
- **RN-005-5** — El mensaje de cada parte es el mensaje del error del catálogo,
  no el texto de las capas que lo envolvieron. Así el usuario no ve ids ni
  frases de contexto como "asignar el item X al sprint Y".
- **RN-005-6** — El `Tipo` del `Problema` es el más grave de sus partes, en este
  orden: `interno`, `no_encontrado`, `conflicto`, `validacion`. Si el tipo es
  `interno`, el problema lleva un único error general con un mensaje fijo y
  ninguno de los otros. En cualquier otro caso se conservan los mensajes de
  todas las partes.
- **RN-005-7** — Una parte que no está en el catálogo es `interno`. El mensaje
  fijo es `ocurrio un error inesperado: proba de nuevo en unos minutos`.
- **RN-005-8** — Cada entrada del catálogo trae el nombre del campo del
  formulario al que pertenece, o ninguno si es un error general. Los nombres
  van en `snake_case` y en español, y son los que usan los formularios de
  T-011: `nombre`, `email`, `rol`, `titulo`, `prioridad`, `criterios`,
  `estimacion`, `sprint_goal`, `fecha_inicio`, `fecha_fin`, `fecha`,
  `actividad`, `horas`.
- **RN-005-9** — Un `*ErrorDeValidacion` (lo que arma el formulario) se
  clasifica como `validacion` y conserva sus campos y mensajes tal cual.
  Puede venir solo o unido con errores del dominio.

**Formulario**

- **RN-005-10** — El formulario solo informa errores de **formato** (un texto
  donde iba un número). La obligatoriedad no se revisa acá: un valor vacío
  devuelve el cero del tipo y sin error, para que el dominio diga "es
  obligatorio" con su propio mensaje y no haya dos redacciones para lo mismo.
- **RN-005-11** — Formatos aceptados, ignorando espacios al principio y al final:
  `Entero` acepta dígitos con signo opcional; `Decimal` acepta dígitos con
  signo opcional y, si hay parte decimal, coma **o** punto (igual que
  US-019); `Fecha` acepta `AAAA-MM-DD`, que es lo que envía un
  `<input type="date">`, y devuelve la fecha a medianoche UTC. Todo lo demás
  es un error de formato, incluidos `NaN`, `Inf` y la notación científica.
- **RN-005-12** — Los errores del formulario se informan en el orden en que se
  consultaron los campos, y un campo se informa una sola vez aunque se
  consulte más de una.
- **RN-005-13** — El mensaje de formato nombra el campo con espacios en lugar
  de guiones bajos, por ejemplo: `el campo story points tiene que ser un
  numero entero`.

**Respuesta**

- **RN-005-14** — `ResponderError` elige JSON si el `Accept` menciona
  `application/json` y no menciona `text/html`; en cualquier otro caso (incluido
  un `Accept` vacío o `*/*`) responde el fragmento HTML.
- **RN-005-15** — Siempre escribe el estado de su tipo, un `Content-Type` con
  `charset=utf-8` y el cuerpo de la sección 3. Escapa todo el texto que
  muestra.
- **RN-005-16** — Solo los errores `interno` se registran, con `slog.Error`, e
  incluyen el método, la ruta y el error completo con todas sus capas. El
  detalle nunca llega al cliente.
- **RN-005-17** — `ResponderError(nil)` es un error de programación: responde
  `interno` y lo registra.

**Recuperar**

- **RN-005-18** — Si el handler que sigue hace `panic`, `Recuperar` lo registra
  con el valor del `panic` y la pila, responde un `interno` uniforme y el
  servidor sigue atendiendo. Si la respuesta ya había empezado, solo registra.
  `http.ErrAbortHandler` se vuelve a lanzar, porque es la forma legítima de
  cortar una respuesta.

**Catálogo**

- **RN-005-19** — Todo error exportado `Err…` de `internal/domain` y de
  `internal/app` está en el catálogo o en la lista de exclusiones, que lleva
  el motivo de cada una. Un test lo comprueba leyendo el código fuente, así
  que una historia nueva que agregue un error y no lo registre rompe el CI.

## 5. Restricciones

- `internal/platform/errores` usa solo la biblioteca estándar y no importa el
  dominio ni la aplicación: conoce los tipos de error, no los errores
  concretos. El catálogo con los errores concretos vive en el adaptador HTTP.
- El dominio no cambia: no conoce HTTP, campos de formulario ni tipos de
  respuesta.
- El catálogo se arma una vez al iniciar y no se modifica después: se puede
  usar desde varias goroutines sin bloqueos.
- `Recuperar` es un `func(http.Handler) http.Handler`, así que sirve igual con
  el `ServeMux` de hoy y con `chi` (ADR 0001).
- `cmd/server` arma el `ManejadorDeErrores` con el catálogo y envuelve el mux con
  `Recuperar`. No agrega rutas.
- Fuera de alcance: la pantalla de error con estilos (T-004), la respuesta
  para rutas que no existen o métodos no permitidos (hoy las contesta el
  `ServeMux` en texto plano) y los formularios concretos (T-011).

## 6. Casos límite

- **CL-005-1** — `Clasificar(nil)` devuelve un `Problema` vacío (sin tipo y sin
  errores). No es un error.
- **CL-005-2** — Dos errores de validación del mismo campo, unidos: se
  conservan los dos, en orden.
- **CL-005-3** — Un error del catálogo envuelto en varias capas con `%w`: se
  reconoce, y el mensaje es el del catálogo (RN-005-5), no el de las capas.
- **CL-005-4** — Una unión dentro de otra unión: se aplana y el orden de los
  mensajes es el de lectura.
- **CL-005-5** — Validación unida con un error que no está en el catálogo: el
  tipo es `interno` y solo sale el mensaje fijo. No se muestran los de
  validación.
- **CL-005-6** — Validación unida con un conflicto: el tipo es `conflicto`
  (409) y salen los mensajes de los dos, en orden.
- **CL-005-7** — Un `*ErrorDeValidacion` unido con errores del dominio: salen
  los campos del formulario y los del dominio, en el orden de la unión.
- **CL-005-8** — Entero: `" 5 "` es 5; `"12.5"`, `"1e3"`, `"abc"` y un número
  que no entra en un `int` son error de formato; `""` es 0 sin error.
- **CL-005-9** — Decimal: `"1,5"` y `"1.5"` son 1,5; `"-2,25"` es −2,25 (el
  dominio decide que no es válido); `"1,5,2"`, `"1.2.3"`, `"NaN"`, `"Inf"` y
  `"1e3"` son error de formato; `""` es 0 sin error.
- **CL-005-10** — Fecha: `"2026-10-06"` es válida; `"2026-02-30"` (día que no
  existe), `"06/10/2026"` y `"2026-10-6"` son error de formato; `""` es la
  fecha cero sin error.
- **CL-005-11** — Varios campos mal formados: se informan todos juntos. Si el
  mismo campo se consulta dos veces, aparece una sola vez.
- **CL-005-12** — `Accept: text/html,application/json;q=0.9` responde HTML
  (menciona `text/html`). `Accept: application/json` responde JSON. Sin
  `Accept` responde HTML.
- **CL-005-13** — Un `panic` con un valor que no es un `error` (por ejemplo un
  texto) también se responde como `interno`.
- **CL-005-14** — `panic` después de que el handler escribió parte de la
  respuesta: no se intenta escribir otra, solo se registra.
- **CL-005-15** — Un mensaje con caracteres de HTML (por ejemplo `<script>`)
  sale escapado en el fragmento y como texto en el JSON.

## 7. Condiciones de error

Este apartado define el comportamiento de los errores, así que el cuadro es el
propio catálogo.

| Situación | Tipo | Estado | Mensaje al usuario |
|---|---|---|---|
| Un error del catálogo de tipo `validacion` | `validacion` | 422 | El del catálogo |
| Un error del catálogo de tipo `conflicto` | `conflicto` | 409 | El del catálogo |
| Un error del catálogo de tipo `no_encontrado` | `no_encontrado` | 404 | El del catálogo |
| Un campo con formato inválido | `validacion` | 422 | `el campo <campo> tiene que ser un numero entero` (o `un numero`, o `una fecha con formato AAAA-MM-DD`) |
| Un error que no está en el catálogo, un `panic`, un `nil` | `interno` | 500 | `ocurrio un error inesperado: proba de nuevo en unos minutos` |

### Catálogo inicial

`validacion` (422):

| Paquete | Error | Campo |
|---|---|---|
| `backlog` | `ErrTituloVacio`, `ErrTituloMuyLargo` | `titulo` |
| `backlog` | `ErrPrioridadInvalida` | `prioridad` |
| `backlog` | `ErrCriterioVacio` | `criterios` |
| `backlog` | `ErrEstimacionFueraDeEscala` | `estimacion` |
| `esfuerzo` | `ErrFechaVacia`, `ErrFechaFutura` | `fecha` |
| `esfuerzo` | `ErrActividadVacia`, `ErrActividadMuyLarga` | `actividad` |
| `esfuerzo` | `ErrHorasFueraDeRango`, `ErrHorasFraccion` | `horas` |
| `integrante` | `ErrNombreVacio`, `ErrNombreLargo` | `nombre` |
| `integrante` | `ErrEmailVacio`, `ErrEmailInvalido` | `email` |
| `integrante` | `ErrRolInvalido` | `rol` |
| `proyecto` | `ErrNombreVacio`, `ErrNombreLargo` | `nombre` |
| `proyecto` | `ErrFechaInicioFaltante` | `fecha_inicio` |
| `proyecto` | `ErrFechaFinFaltante`, `ErrFechasIncoherentes` | `fecha_fin` |
| `sprint` | `ErrSprintGoalVacio`, `ErrSprintGoalMuyLargo` | `sprint_goal` |
| `sprint` | `ErrFechaInicioFaltante` | `fecha_inicio` |
| `sprint` | `ErrFechaFinFaltante`, `ErrFechasIncoherentes`, `ErrDuracionExcedida` | `fecha_fin` |

`conflicto` (409):

| Paquete | Error | Campo |
|---|---|---|
| `app` | `ErrEmailDuplicado` | `email` |
| `backlog` | `ErrItemCompletado` | — |
| `sprint` | `ErrInicioNoPosteriorAlUltimo` | `fecha_inicio` |
| `sprint` | `ErrFueraDelProyecto`, `ErrUltimoSprintAbierto` | — |
| `sprint` | `ErrProyectoDistinto`, `ErrSprintCerrado`, `ErrItemYaAsignado`, `ErrItemSinEstimar`, `ErrItemSinCriterios` | — |

`no_encontrado` (404), sin campo: `app.ErrProyectoInexistente`,
`app.ErrItemInexistente`, `app.ErrIntegranteInexistente` y
`app.ErrSprintInexistente`.

**Exclusiones** (no se registran, así que salen como `interno`):

| Error | Motivo |
|---|---|
| `metricas.ErrStoryPointsNegativos`, `ErrResumenInconsistente`, `ErrNombreVacio` | Describen datos propios del sistema que salieron mal, no algo que la persona escribió |
| `config.*`, `db.ErrURLVacia` | Errores de arranque: el servidor no llega a atender peticiones |

Los errores de las historias que se sumen después (US-010, US-011, …) los
registra cada una en su propio PR.

## 8. Criterios de aceptación

- **CA-005-1** (normal) — Dado un error de validación del dominio (por
  ejemplo `ErrTituloVacio`), la respuesta tiene estado 422, el tipo
  `validacion`, el campo `titulo` y el mensaje del catálogo.
- **CA-005-2** (normal) — Dados varios errores unidos con `errors.Join`, la
  respuesta los informa todos, en el mismo orden y cada uno con su campo.
- **CA-005-3** (alternativo) — Un error de inexistencia responde 404, uno
  de estado (por ejemplo un sprint cerrado) responde 409, y la misma
  petición con `Accept: application/json` recibe JSON mientras que una de
  HTMX recibe el fragmento HTML.
- **CA-005-4** (alternativo) — Dado un formulario con un campo mal formado (un
  texto en las horas), la respuesta es 422 con el campo `horas`, en el mismo
  formato que un error del dominio, y se puede combinar con los errores del
  dominio de otros campos.
- **CA-005-5** (límite) — Un error del catálogo envuelto en varias capas, o
  dentro de una unión anidada, se reconoce y muestra solo el mensaje del
  catálogo.
- **CA-005-6** (error) — Dado un error que no está en el catálogo, la respuesta
  es 500 con el mensaje fijo, no incluye ninguna parte del error original y
  el error completo queda en el registro.
- **CA-005-7** (error) — Dado un handler que hace `panic`, la respuesta es el 500
  uniforme y el servidor atiende la petición siguiente.
- **CA-005-8** (error) — Si un error exportado de `internal/domain` o
  `internal/app` no está en el catálogo ni en las exclusiones, el test del
  catálogo falla y nombra el error.

---

## Decisiones tomadas y descartadas

- **El tipo de error vive en `internal/platform`, el catálogo en el adaptador
  HTTP.** La plataforma no puede importar el dominio, y el dominio no puede
  saber de HTTP. El adaptador HTTP ya conoce a los dos, así que es el único
  lugar donde puede estar la lista que los une.
- **Catálogo explícito con `errors.Is`.** Se descartó que cada error del dominio
  implemente una interfaz (`Tipo()`, `Campo()`): obligaba a convertir unos 40
  errores de 7 paquetes en tipos propios y metía conceptos de pantalla en el
  dominio, que hoy es puro. Tampoco se parsea el texto del mensaje: cualquier
  corrección de redacción rompería la clasificación sin avisar. El precio es
  mantener una tabla, y por eso existe RN-005-19, que hace fallar el CI si se
  olvida.
- **422 para validación y 409 para estado.** 400 queda para una petición que no
  se puede ni leer, y acá el servidor entiende la petición pero rechaza los
  valores. Separar lo que depende de lo escrito (se corrige en el formulario)
  de lo que depende del estado (puede que haya que ir a otra pantalla) lo
  hace la regla RN-005-2, que se puede aplicar a los errores que lleguen.
- **Se muestra el mensaje del catálogo, no el del error envuelto.** Los errores
  reales llevan ids y contexto (`el item no existe (id …)`), útil para el log
  y molesto para quien lo lee. Se descartó recortar el texto a mano.
- **Con un error interno no se muestra nada de los demás errores.** Si se
  mostraran los de validación junto a un fallo del sistema, la persona
  corregiría lo que escribió y volvería a fallar sin motivo aparente.
- **El formulario no revisa que los campos estén completos.** Hacerlo
  duplicaría las reglas del dominio con otro texto. Solo traduce texto a
  números y fechas, que es lo único que el dominio no puede ver.
- **Dependencia con T-004: HTMX no intercambia respuestas 4xx ni 5xx por
  defecto.** El fragmento de error no se vería. El layout tiene que configurar
  la regla de intercambio para 404, 409, 422 y 500 (`htmx.config.responseHandling`
  en HTMX 2.x). Esa configuración va en T-004 porque es parte del layout, y esta
  tarea la deja anotada en su aceptación.
- **No hay `.feature`.** Es una tarea técnica sin comportamiento de negocio y
  los pasos de BDD trabajan sobre los casos de uso, no sobre HTTP. Los criterios
  se verifican con tests de tabla de Go y `httptest`. El chequeo de
  trazabilidad del CI solo exige escenario a las specs `US-*`. Las pantallas de
  T-011 vuelven a probar esto de punta a punta.
- **No se registran los errores de `metricas`.** Si una velocidad no se puede
  calcular porque un resumen está corrupto, no hay nada que la persona pueda
  corregir en un formulario: es un fallo del sistema y se trata como tal.

## Uso de IA en esta especificación

- [x] Use IA para: redacción completa de la especificación (los 8 apartados) y
  armado del catálogo a partir de los errores que hay en el código — revisado
  por: Tomás.
