# T-020 — Pantallas de Planning Poker: iniciar, votar, revelar y acordar

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #190 |
| **Escenarios BDD** | ninguno: es una tarea técnica (ver "Decisiones") |
| **Código** | `internal/adapters/http/`, `web/templates/`, `internal/app/` (la lectura de la sesión), `cmd/server/` |
| **Última actualización** | 2026-10-27 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Que el equipo haga una sesión de Planning Poker completa desde el navegador:
iniciarla sobre una historia, votar sin ver los votos de los demás, revelar,
abrir otra ronda si hay divergencia y registrar la estimación acordada. Es la
interfaz del requerimiento 4 de la guía. Los casos de uso ya existen (US-014,
US-015, US-016 y US-018; US-017 para la nueva ronda), así que **no hay reglas de
negocio nuevas**: son rutas, una lectura de la sesión y la forma de mostrarla,
con el layout, el router y `Responder` de T-004 y el molde de T-011, T-015 y
T-016.

## 2. Entradas

| Ruta | Método | Campos del formulario | Caso de uso |
|---|---|---|---|
| `/proyectos/{id}/backlog/{item}/planning-poker` | POST | — | `IniciarSesionPlanningPoker` (US-014) |
| `/proyectos/{id}/planning-poker/{sesion}` | GET | — | `ConsultarSesionPlanningPoker` (nuevo) |
| `/proyectos/{id}/planning-poker/{sesion}/votos` | POST | `integrante`, `estimacion` | `VotarEnSesion` (US-015) |
| `/proyectos/{id}/planning-poker/{sesion}/revelar` | POST | `integrante` | `RevelarVotos` (US-016) |
| `/proyectos/{id}/planning-poker/{sesion}/nueva-ronda` | POST | `integrante` | `AbrirNuevaRonda` (US-017) |
| `/proyectos/{id}/planning-poker/{sesion}/acuerdo` | POST | `integrante`, `estimacion` | `RegistrarEstimacionAcordada` (US-018) |

- `{id}`, `{item}` y `{sesion}` son los IDs que genera la base. Las acciones son
  POST con HTMX, como en T-011, T-015 y T-016.
- `integrante` es el ID de un participante, elegido de una lista desplegable con
  sus nombres: no hay usuarios en el sistema, así que cada persona elige quién es
  (hueco 1 de US-015).
- `estimacion` es uno de los valores de la escala (`backlog.EscalaFibonacci()`),
  con el nombre de campo del catálogo de T-005 (US-015, sección 5). Se elige con
  una carta por valor.

## 3. Salidas esperadas

**Después de una acción con éxito**, igual que en T-015 y T-016: `HX-Redirect`
con HTMX o `303` sin él, hacia la pantalla de la sesión. Los errores se
responden con `ResponderError` (T-005) y la pantalla no se toca.

**Backlog** (T-011, cambia una cosa): cada historia que no está completada tiene
un botón **"Planning Poker"**. Si la historia tiene una sesión abierta, el botón
es un enlace a esa sesión; si no, inicia una.

**Pantalla de la sesión**, según su estado:

| Parte | `votando` | `revelada` | `finalizada` |
|---|---|---|---|
| Encabezado | La historia (número y título), la ronda y el estado | igual | igual, con la estimación acordada |
| Participantes | Quiénes votaron y quiénes faltan, **sin valores** | — | — |
| Votar | Lista de participantes y una carta por valor de la escala | — | — |
| Revelar | Botón con la lista de participantes | — | — |
| Resultado de la ronda | — | Cada voto con su nombre, quiénes faltaron, el mínimo, el máximo y la divergencia **marcada** | — |
| Nueva ronda | — | Botón, solo si hay divergencia | — |
| Estimación acordada | — | Formulario con la lista de participantes y la escala | — |
| Rondas anteriores | Su resultado, ronda por ronda | igual | igual |

## 4. Reglas de negocio

No hay reglas de dominio nuevas. Son reglas de comportamiento de la interfaz.

- **RN-T020-1** — El handler solo lee la petición, comprueba que la sesión sea
  del proyecto (con la lectura de RN-T020-3), llama a un caso de uso y escribe la
  respuesta, como RN-T015-1 y RN-T016-1.
- **RN-T020-2** — **La pantalla nunca muestra el valor de un voto de la ronda
  actual mientras se vota.** La vista se arma con la `Votacion` (RN-016-12):
  `Votaron()` y `Faltan()` para la ronda actual, y los valores solo con la sesión
  revelada o finalizada, o para rondas anteriores (RN-017-6). Ni el handler ni la
  plantilla reciben valores en crudo.
- **RN-T020-3** — La lectura `ConsultarSesionPlanningPoker` es un caso de uso de
  `internal/app` con su puerto, como las de T-015 (RN-T015-2). Compone lo que ya
  existe: la sesión (`ObtenerSesion`), la historia (`ObtenerItem`), los nombres
  de los participantes (`ListarIntegrantes`), la votación de la ronda actual
  (`VotacionDeLaRonda`) y las de las rondas anteriores (US-017). El resultado de
  una ronda revelada se arma con la función de dominio que extrae US-017 de
  `Revelar`: la pantalla no calcula la divergencia ni los extremos.
- **RN-T020-4** — Una sesión solo se alcanza por la ruta de **su** proyecto (el
  de su historia): otra combinación responde 404, como RN-T015-5.
- **RN-T020-5** — Los botones dependen del estado (tabla de la sección 3): votar
  y revelar con la sesión `votando`; nueva ronda (solo con divergencia) y
  estimación acordada con la sesión `revelada`; ninguno con la sesión
  `finalizada`. La pantalla no es la que protege las reglas: las rutas responden
  el error del caso de uso igual (RN-T016-6).
- **RN-T020-6** — La divergencia se marca con texto, no solo con color
  ("Hay divergencia: de 3 a 8"), para que se entienda sin distinguir colores.
- **RN-T020-7** — Cada campo tiene su `<label>`, y cada carta y cada botón tienen
  un nombre accesible ("Votar 5", "Revelar los votos de la ronda 2"), como
  RN-T015-12.
- **RN-T020-8** — Los textos del estado y de la divergencia se arman en el
  handler junto con el resto de la vista (RN-T016-8). La plantilla no decide.

## 5. Restricciones

- **Sin reglas de dominio nuevas.** Usa la función del resultado de una ronda que
  agrega US-017. En `internal/app` solo suma la lectura de RN-T020-3.
- **Depende de US-014, US-015, US-016 y US-018**, en `main`, y de **US-017** para
  el botón de nueva ronda y las rondas anteriores. Si US-017 no llega a tiempo, la
  pantalla sale sin esas dos partes y se suman en un PR aparte.
- Los casos de uso se agregan a `CasosDeUso` y se conectan en `cmd/server`;
  `RepositorioPlanningPoker` ya cumple sus puertos.
- Los errores que aparezcan ya están en el catálogo (US-014 a US-018).
- Los formularios requieren JavaScript (HTMX), como en T-011, T-015 y T-016. El
  diseño visual es funcional (Tailwind) y se mejora después.
- **Fuera de alcance:** actualización en vivo (cada persona recarga para ver los
  votos de las demás), usuarios o login, cancelar una sesión y el historial de
  sesiones finalizadas de una historia.

## 6. Casos límite

- **CL-T020-1** — Historia completada: el backlog no ofrece "Planning Poker".
- **CL-T020-2** — Historia con una sesión abierta: el botón lleva a esa sesión;
  iniciar otra por la ruta responde 409 con `ErrSesionAbierta`.
- **CL-T020-3** — Proyecto con un solo integrante: iniciar responde 409 con
  `ErrParticipantesInsuficientes`.
- **CL-T020-4** — Nadie votó todavía: todos figuran en "Faltan"; revelar responde
  409 con `ErrSinVotos`.
- **CL-T020-5** — Un participante vota dos veces: figura una sola vez en
  "Votaron" y el valor no aparece en la pantalla.
- **CL-T020-6** — Revelado sin divergencia: no aparece "Abrir nueva ronda", sí el
  formulario de la estimación acordada.
- **CL-T020-7** — Ronda 2 votando: la ronda 1 se ve con sus valores en "Rondas
  anteriores" y la ronda 2 no muestra ninguno.
- **CL-T020-8** — Sesión finalizada: muestra la estimación acordada y no tiene
  formularios; la historia aparece estimada con ese valor en el backlog.
- **CL-T020-9** — Una sesión de otro proyecto, o un ID inexistente o mal
  formado: 404.
- **CL-T020-10** — Doble clic en revelar o en registrar: el primero cambia la
  sesión y el segundo responde 409 con el mensaje del catálogo, sin cambiar nada.

## 7. Condiciones de error

| Situación | Respuesta |
|---|---|
| Sesión, historia o integrante inexistentes, ID mal formado o sesión de otro proyecto | 404: página de error (GET) o fragmento (POST) |
| Valor fuera de la escala | 422 con el campo `estimacion` y el mensaje del catálogo |
| Regla de la sesión (no participa, votación cerrada, sin votos, sin revelar, sin divergencia, sesión abierta, participantes insuficientes, historia completada) | 409 con el mensaje del catálogo |
| Falla inesperada | 500 con el mensaje fijo de T-005; el detalle solo va al registro |

## 8. Criterios de aceptación

- **CA-T020-1** (normal) — Desde una historia del backlog se puede iniciar una
  sesión de Planning Poker. La pantalla de la sesión muestra la historia, la
  ronda, el estado y los participantes.
- **CA-T020-2** (normal) — Un participante elige quién es y vota con un valor de
  la escala. La pantalla muestra quiénes ya votaron y quiénes faltan, y **nunca
  el valor de otro participante** hasta revelar (RN-016-12).
- **CA-T020-3** (alternativo) — Revelar muestra todos los votos, el mínimo, el
  máximo y si hay divergencia (US-016).
- **CA-T020-4** (alternativo) — Registrar la estimación acordada (US-018)
  finaliza la sesión y deja la historia estimada con ese valor, que se ve en el
  backlog.
- **CA-T020-5** (alternativo) — Con divergencia aparece "Abrir nueva ronda"
  (US-017) y la pantalla vuelve a ocultar los votos de la ronda nueva.
- **CA-T020-6** (límite) — Una sesión finalizada o una historia completada
  muestran su estado y no ofrecen votar; con una sesión abierta no se puede
  iniciar otra para la misma historia.
- **CA-T020-7** (error) — Votar fuera de la escala, que vote quien no participa,
  revelar sin votos o registrar la estimación con la sesión sin revelar muestran
  el error uniforme (T-005) sin romper la pantalla.

---

## Decisiones tomadas y descartadas

- **Es una tarea y no una historia.** Los casos de uso y sus reglas ya están; esto
  es la interfaz, como T-011, T-015 y T-016. No hay `.feature`: el recorrido se
  prueba con tests de Go con `httptest` sobre el router real y la base en
  memoria, como CA-T016-1.
- **Inicio desde la fila del backlog.** T-011 no tiene una página de detalle de
  la historia; la fila del backlog es donde ya están sus acciones (estimar,
  editar).
- **Cada persona elige quién es.** No hay login; se descartó sumar usuarios para
  esta tarea. La pantalla lo dice ("Votás como…") para que nadie vote por otro
  sin darse cuenta.
- **Sin actualización en vivo.** Se recarga la pantalla para ver quién votó. Se
  descartó sondear al servidor o usar websockets: es alcance de más para una
  demo con el equipo en la misma sala.
- **La pantalla no ve valores en crudo.** Usa la `Votacion` y el resultado del
  dominio. Así la garantía del voto oculto (US-015) llega hasta el navegador.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la
  tarea y de las decisiones del equipo (pantalla por sesión, botones según el
  estado, votar eligiendo integrante y carta, errores por el catálogo), usando
  las specs de T-011, T-015, T-016 y US-014 a US-018 como modelo
- Revisado y entendido por:
