# US-010 — Marcar historias como completadas dentro de un Sprint

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #18 |
| **Escenarios BDD** | `features/US-010-marcar-historias-completadas.feature` |
| **Código** | `internal/domain/sprint/` (`VerificarMarcado`), `internal/domain/backlog/` (`ItemBacklog.Completar` y `Reabrir`), casos de uso `MarcarCompletado` y `DesmarcarCompletado` en `internal/app/`, `GuardarCambioDeEstado` en `internal/adapters/postgres/` y `internal/adapters/memoria/`, migración `migraciones/00008_items_completado_en.sql` |
| **Última actualización** | 2026-10-19 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero marcar como completadas las historias de un
> Sprint, para registrar el avance del trabajo mientras el Sprint está en curso.

Marcar (y, si fue un error de carga, desmarcar) una historia asignada al Sprint
abierto. Es el dato que usan los Story Points completados (US-025), la
velocidad (US-026) y el burndown (US-031), y el que decide qué historias
vuelven al backlog al cerrar el Sprint (US-011).

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `itemID` | UUID | Sí | Un ítem existente. Viene de la URL; lo verifica el caso de uso |
| `sprintID` | UUID | Sí | Un Sprint existente. Viene de la URL (el tablero del Sprint); lo verifica el caso de uso |
| `completadoEn` | `time.Time` | Solo al marcar | La pone el caso de uso con el reloj inyectado, como `creadoEn` en US-005; el dominio la recibe |

Son dos operaciones, **marcar** y **desmarcar**, con el mismo recorrido:

1. **Aplicación:** carga el ítem y después el Sprint, con lecturas simples. Si
   alguno no existe, devuelve `ErrItemInexistente` o `ErrSprintInexistente` sin
   validar nada más, con el orden y la prioridad de RN-009-1 y RN-009-2.
2. **Dominio de `sprint`:** valida, con datos mínimos del ítem y del Sprint
   (tipos propios del paquete, como en US-009), que el ítem esté en ese Sprint
   (que tenga una fila `pendiente` en él) y que el Sprint esté abierto.
3. **Dominio de `backlog`:** `ItemBacklog.Completar(en)` o `Reabrir()` devuelve
   una copia con el estado nuevo, igual que `Estimar` en US-013.
4. **Persistencia:** guarda el cambio en una transacción con el Sprint bloqueado
   (RN-010-9).

## 3. Salidas esperadas

| Atributo | Después de marcar | Después de desmarcar |
|---|---|---|
| Estado | `completado` | `pendiente` |
| Fecha de completado | `completadoEn` (o la original, si ya estaba completado) | Sin fecha |
| Resto del ítem y su fila en el Sprint | Sin cambios: la fila sigue `pendiente` (el resultado lo pone US-011) | Sin cambios |

`ItemBacklog` suma el campo `CompletadoEn time.Time`: su valor cero significa
"sin fecha", y solo un ítem `completado` la tiene. Si hay errores, no se
devuelve ítem y no se modifica nada.

## 4. Reglas de negocio

Entre paréntesis, la capa que la hace cumplir.

- **RN-010-1** (aplicación) — El ítem y el Sprint tienen que existir. Se
  verifican en el orden de RN-009-1 y RN-009-2, con la misma prioridad y sin
  unirse con los errores de validación. Un ID mal formado es un inexistente.
- **RN-010-2** (dominio) — Solo se marca o desmarca un ítem que está en ese
  Sprint. Con la relación histórica de US-009 (RN-009-13), "estar en el Sprint"
  es tener una fila `pendiente` en él; haber estado en otro Sprint no cuenta.
  Si no, `ErrItemFueraDelSprint`.
- **RN-010-3** (dominio) — Solo con el Sprint `abierto`. Con el Sprint
  `cerrado` no se marca ni se desmarca: `ErrSprintCerrado` (el de US-009).
- **RN-010-4** (dominio) — Con el Sprint `cerrado` se devuelve solo
  `ErrSprintCerrado`, sin mirar si el ítem está en él. Al cerrar, ninguna fila
  queda `pendiente` (US-011), así que unirlo con `ErrItemFueraDelSprint` diría
  que no está asignada una historia que sí estuvo. Con el Sprint abierto, el
  único error posible es `ErrItemFueraDelSprint`. Es una excepción a propósito
  al criterio de RN-009-8.
- **RN-010-5** (dominio) — Marcar deja el ítem `completado` con
  `completadoEn`. Vale desde `pendiente` o `en_progreso`.
- **RN-010-6** (dominio) — Marcar un ítem que ya está `completado` no es un
  error y no cambia nada: conserva la fecha original.
- **RN-010-7** (dominio) — Desmarcar deja el ítem `pendiente` y sin fecha,
  aunque antes de completarse estuviera `en_progreso`. Desmarcar un ítem que no
  está completado no es un error y no cambia nada.
- **RN-010-8** (dominio) — El dominio no lee el reloj: la fecha entra como
  parámetro.
- **RN-010-9** (persistencia) — El repositorio guarda el cambio en una
  transacción `READ COMMITTED`, con estos pasos en orden, y se detiene en el
  primer error:
  1. Bloquea la fila del Sprint y lee su estado:
     `SELECT ... FROM sprints WHERE id = $1 FOR NO KEY UPDATE`. Si no hay fila,
     `ErrSprintInexistente`.
  2. Lee el estado del ítem y si tiene una fila `pendiente` en ese Sprint. Si
     el ítem no existe, `ErrItemInexistente`.
  3. Vuelve a verificar RN-010-2 y RN-010-3 con la misma función del dominio y
     los datos leídos con el candado tomado.
  4. Actualiza el estado y la fecha del ítem.

  El orden de bloqueo (primero el Sprint, después el ítem) es el de RN-009-11.
  Un cierre (US-011) o una asignación (US-009) del mismo Sprint esperan, o hacen
  esperar, a este candado; un marcado que llega después del cierre ve el Sprint
  `cerrado` y se rechaza (CL-010-9).
- **RN-010-10** (aplicación) — La operación es de todo o nada: si algo falla,
  el ítem queda como estaba.

## 5. Restricciones

- **Dominio puro:** igual que US-005 y US-008 (sección 5).
- **Migración nueva.** Agrega a `items_backlog` la columna `completado_en`
  (`TIMESTAMPTZ`, nula) y la red de seguridad
  `CHECK ((estado = 'completado') = (completado_en IS NOT NULL))`. Es la
  `00008`, después de la `00007` de la relación de US-009. Un ítem que ya
  estuviera completado al migrar recibe la fecha de la migración, para cumplir el
  `CHECK`.
- **El repositorio aplica el cambio con el Sprint bloqueado.** El caso de uso le
  pasa a `GuardarCambioDeEstado` el cambio del dominio (`Completar(en)` o
  `Reabrir()`), y el repositorio lo aplica al ítem que lee después de tomar el
  candado, no al que leyó el caso de uso. Así un segundo marcado simultáneo ve
  el ítem ya completado y conserva la fecha del primero (CL-010-10).
- **Dependencia con US-009.** Necesita la relación histórica ítem-Sprint
  (RN-009-13, corregida en el #104): sin US-009 no hay ítems en un Sprint que
  marcar. Marcar y desmarcar cambian el estado del ítem, no el resultado de su
  fila, que completa US-011 al cerrar.
- **Dependencia con US-013 (#100).** US-013 guardaba la estimación con un
  `UPDATE` sin condición (RN-013-7): si un ítem se completaba entre la lectura
  y ese `UPDATE`, se re-estimaba un ítem completado. El #110 lo condicionó a
  `estado <> 'completado'`, igual que RN-006-8, y ya está en `main`.
- **Dependencia con US-011.** El escenario de CA-010-5 necesita un Sprint
  cerrado, y el cierre es US-011. Hasta entonces, CA-010-5 se cubre con tests
  unitarios (el dominio con el Sprint cerrado como dato) y de integración (el
  repositorio con un Sprint cerrado armado en la base). Mismo criterio que
  US-008 y US-009.
- **Frases BDD.** `Dado la historia "X" completada` y
  `Cuando se marca la historia "X" como completada` ya están en el diccionario.
  Operan sobre el último Sprint del proyecto actual, que es el único que puede
  estar abierto (RN-008-17); así CA-010-4 se prueba con una historia que no está
  en él. El estado se verifica con `la historia "X" tiene estado "completado"`
  (US-005). Se agrega:

  | Frase | Expresión |
  |---|---|
  | `Cuando se desmarca la historia "X"` | `^se desmarca la historia "([^"]*)"$` |

  La interfaz `Sprints` de `features/steps/mundo.go` suma
  `DesmarcarHistoria(titulo string) error`. La fecha de completado no se
  verifica en los escenarios (el reloj de los escenarios es fijo): CA-010-1 y
  CA-010-3 la cubren con tests unitarios.
- **Fuera de alcance.** Pasar un ítem a `en_progreso` (ninguna historia lo pide
  todavía), el tablero del Sprint (T-011), el burndown (US-031) y cerrar el
  Sprint (US-011).

## 6. Casos límite

- **CL-010-1** — Ítem `pendiente` o `en_progreso` asignado al Sprint abierto:
  se marca, con la fecha recibida.
- **CL-010-2** — Ítem ya completado, marcado otra vez con otra fecha: no cambia
  nada y conserva la fecha original (RN-010-6).
- **CL-010-3** — Ítem completado, desmarcado: queda `pendiente` y sin fecha. Si
  antes de completarse estaba `en_progreso`, igual queda `pendiente`.
- **CL-010-4** — Ítem `pendiente`, desmarcado: no cambia nada (RN-010-7).
- **CL-010-5** — Ítem sin fila `pendiente` en ese Sprint: está en el backlog
  (aunque haya estado en Sprints cerrados) o en el Sprint de otro proyecto.
  Se rechaza con `ErrItemFueraDelSprint`.
- **CL-010-6** — Sprint cerrado e ítem que quedó completado en él, desmarcado:
  se rechaza solo con `ErrSprintCerrado` (RN-010-4), aunque al cerrar su fila
  dejó de estar `pendiente`. Lo medido no se toca.
- **CL-010-7** — Sprint cerrado e ítem que no está en él: solo
  `ErrSprintCerrado` (RN-010-4).
- **CL-010-8** — Ítem o Sprint inexistente, o con ID mal formado, aunque además
  haya otros problemas: solo el error de inexistencia (RN-010-1).
- **CL-010-9** — El Sprint se cierra entre la lectura del caso de uso y la
  transacción: el repositorio lo ve `cerrado` con el candado tomado y rechaza
  con `ErrSprintCerrado`. Se verifica llamando al repositorio con un Sprint ya
  cerrado en la base.
- **CL-010-10** — Dos marcados simultáneos del mismo ítem: el segundo espera el
  candado del Sprint, ve el ítem completado y no cambia nada. Queda la fecha
  del primero y ninguno da error.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje del error (origen) |
|---|---|---|
| El ítem no existe | `app.ErrItemInexistente` (el de US-013 y US-019), envuelto indicando el ID | "el item no existe" |
| El Sprint no existe | `app.ErrSprintInexistente` (el de US-009), envuelto indicando el ID | "el sprint no existe" |
| El Sprint está cerrado | `sprint.ErrSprintCerrado` (el de US-009), envuelto indicando el nombre del Sprint | "el sprint esta cerrado": el mensaje general (hecho en la corrección de US-009, #104) |
| El ítem no está en ese Sprint (no tiene una fila `pendiente` en él) | `sprint.ErrItemFueraDelSprint` | "la historia no esta asignada a este sprint" |

Los dos primeros tienen prioridad y no se unen con los demás (RN-010-1). Los dos
últimos son del dominio y nunca se devuelven juntos: con el Sprint cerrado, solo
`ErrSprintCerrado` (RN-010-4). El repositorio puede devolverlos también, con la
misma función, si el Sprint se cerró en el medio. Todos se comparan con `errors.Is`, y en ningún caso se
modifica el ítem.

## 8. Criterios de aceptación

- **CA-010-1** (normal) — Una historia asignada al Sprint abierto se marca como
  completada: queda completada y con la fecha de completado.
- **CA-010-2** (alternativo) — Una historia completada se desmarca con el Sprint
  abierto: vuelve a pendiente y pierde la fecha.
- **CA-010-3** (límite) — Marcar una historia que ya está completada no cambia
  nada: conserva su fecha original.
- **CA-010-4** (error) — Una historia que no está asignada a ese Sprint no se
  marca.
- **CA-010-5** (error) — Con el Sprint cerrado no se marca ni se desmarca.
- **CA-010-6** (error) — Una historia o un Sprint inexistentes se rechazan y no
  se modifica nada.

---

## Decisiones tomadas y descartadas

- **Solo historias del Sprint, y con el Sprint abierto.** Marcar es registrar el
  avance del Sprint en curso; lo de un Sprint cerrado ya está medido (US-025 y
  US-026). Se descartó marcar desde el backlog, fuera de un Sprint.
- **Se guarda la fecha de completado.** El burndown (US-031) la necesita para
  el avance diario. Entra como parámetro porque el dominio no lee el reloj. Se
  descartó no guardarla y usar la fecha de cierre: el burndown quedaría plano
  hasta el último día. Es un instante (`TIMESTAMPTZ`), como `creado_en`, y el
  burndown la agrupa por día.
- **Marcar dos veces no es un error.** Un doble clic o un reintento no tienen
  que fallar ni mover la fecha (y con ella el burndown). Se descartó rechazarlo.
  Por simetría, desmarcar un ítem que no está completado tampoco es un error.
- **Se puede desmarcar mientras el Sprint esté abierto.** Los errores de carga
  pasan, y con el Sprint abierto todavía no hay nada medido. Con el Sprint
  cerrado, no: cambiaría hacia atrás la velocidad. Desmarcar vuelve a
  `pendiente` porque no se guarda el estado anterior. Se descartó no permitir
  desmarcar.
- **Candado `FOR NO KEY UPDATE` sobre el Sprint.** Así el marcado no compite
  con el cierre de US-011, que modifica esa fila: uno espera al otro y el
  segundo ve el resultado del primero. También serializa con las asignaciones
  (`FOR SHARE` en US-009) y no bloquea las claves foráneas (mismo criterio que
  US-005 y US-008). Se descartó bloquear solo el ítem: no protege contra un
  cierre simultáneo.
- **La validación es del dominio de `sprint` y el cambio de estado, del de
  `backlog`.** La regla (ítem en el Sprint, Sprint abierto) es del Sprint y usa
  datos mínimos, como US-009. El estado y la fecha son del ítem, como `Estimar`
  en US-013.
- **Con el Sprint cerrado, solo `ErrSprintCerrado`.** Con la relación
  histórica, en un Sprint cerrado ninguna historia tiene fila `pendiente`, así
  que unir los dos errores, como hace US-009 (RN-009-8), agregaría siempre "la
  historia no esta asignada a este sprint", también para las que sí estuvieron.
  Se descartó unirlos, que era la versión anterior de RN-010-4.
- **Se reutiliza `ErrSprintCerrado` de US-009.** La condición es la misma: el
  Sprint está cerrado. Usa el mensaje general (hecho en la corrección de
  US-009, #104).
- **US-013 condiciona su estimación.** Ver la dependencia con US-013 en la
  sección 5 (#100, resuelto en el #110).
- **Las frases existentes operan sobre el Sprint abierto.** Evita una frase que
  nombre el Sprint en cada marcado, y alcanza porque un proyecto tiene como
  máximo un Sprint abierto (RN-008-17).

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de
  la historia y de las decisiones del equipo, usando las specs de US-005,
  US-008, US-009 y US-013 como modelo
- [x] Use IA para: la corrección de RN-010-4 (solo `ErrSprintCerrado` con el
  Sprint cerrado) y el ajuste tras la implementación (migración, cambio aplicado
  con el Sprint bloqueado, dependencia con US-013), redactados con Claude Code
- Revisado y entendido por:
