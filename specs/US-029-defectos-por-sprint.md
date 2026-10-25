# US-029 — Calcular defectos detectados y resueltos por sprint

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás |
| **Issue** | #35 |
| **Escenarios BDD** | `features/US-029-defectos-por-sprint.feature` |
| **Código** | `internal/domain/metricas/` |
| **Última actualización** | 2026-10-25 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como Agile Enabler, quiero conocer cuántos defectos se detectaron y cuántos se
> resolvieron en un Sprint, para ver si la calidad del producto mejora o se
> deteriora de un Sprint a otro.

Calcular, para un Sprint, la cantidad de defectos **detectados** en él y la
cantidad de defectos **resueltos** en él. Son las dos cifras de calidad del
requerimiento 7 de la guía, la entrada del Dashboard (US-030) y del gráfico de
evolución de defectos (US-032).

Es una función pura, como US-025, US-026 y US-027: no accede a la base ni
conoce la forma en que se guardan los defectos. **No depende de US-022 ni de
US-023** (que definen el defecto y su ciclo de vida): recibe solo los dos datos
que necesita de cada defecto, el Sprint donde se detectó y el Sprint donde se
resolvió, y quien la llama adapta lo que tenga guardado.

## 2. Entradas

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `sprint` | `int` | Sí | El número del Sprint a consultar, `>= 1` (el `Sprint N` del proyecto) |
| `defectos` | `[]DefectoRegistrado` | Sí (puede ser una lista vacía) | Todos los defectos del proyecto, de cualquier Sprint |
| `DefectoRegistrado.SprintDeteccion` | `int` | Sí | `>= 1`: el número del Sprint donde se detectó |
| `DefectoRegistrado.SprintResolucion` | `int` | Sí | `0` si el defecto sigue sin resolver; si no, el número del Sprint donde se resolvió, `>= SprintDeteccion` |

Los Sprints se identifican por su **número** y no por su ID: así la función
puede verificar que un defecto no se haya resuelto antes de detectarse. Quien
la llama traduce el ID del Sprint de cada defecto al número (`sprint.Numero`).
La lista que recibe son todos los defectos del proyecto y la función elige los
que corresponden al Sprint consultado.

## 3. Salidas esperadas

```
ResumenDefectos {
    Detectados int // defectos cuyo Sprint de detección es el consultado
    Resueltos  int // defectos cuyo Sprint de resolución es el consultado
}

func CalcularDefectosDelSprint(sprint int, defectos []DefectoRegistrado) (ResumenDefectos, error)
```

Son conteos, sin redondeo. `Resueltos` cuenta los defectos que se resolvieron
**durante** ese Sprint, sin importar en cuál se detectaron.

## 4. Reglas de negocio

- **RN-029-1** — `Detectados` es la cantidad de defectos con
  `SprintDeteccion == sprint`.
- **RN-029-2** — `Resueltos` es la cantidad de defectos con
  `SprintResolucion == sprint`. Un defecto se cuenta como resuelto en el Sprint
  donde se resolvió, no en el que se detectó.
- **RN-029-3** — Un defecto detectado y resuelto en el mismo Sprint cuenta en
  las dos cifras de ese Sprint.
- **RN-029-4** — Un defecto con `SprintResolucion = 0` está sin resolver: no
  cuenta como resuelto en ningún Sprint.
- **RN-029-5** — Una lista vacía, o sin defectos de ese Sprint, da `0` y `0` y
  no es un error.
- **RN-029-6** — El Sprint consultado tiene que ser `>= 1`. Si no,
  `ErrSprintInvalido`, envuelto indicando el valor.
- **RN-029-7** — Cada defecto tiene que ser coherente: `SprintDeteccion >= 1`
  (si no, `ErrSprintInvalido`) y, si está resuelto, `SprintResolucion >=
  SprintDeteccion` (si no, `ErrResolucionAnteriorADeteccion`). Cada error indica
  la posición del defecto en la lista, contada desde 1.
- **RN-029-8** — La función valida todo y devuelve todos los errores juntos,
  unidos con `errors.Join`, en este orden fijo: primero el del Sprint consultado
  y después uno por cada defecto inválido, por posición ascendente. Con algún
  error no se devuelve resumen: una cifra calculada sobre datos corruptos sería
  peor que un error. Cada error unido se reconoce con `errors.Is`.
- **RN-029-9** — La severidad, el estado y la historia relacionada del defecto
  no intervienen: los filtra quien llama antes de pasarle la lista.

## 5. Restricciones

- **Dominio puro.** `internal/domain/metricas` no importa nada fuera de la
  biblioteca estándar, no lee el reloj ni la base.
- **Tipos propios.** `DefectoRegistrado` es un tipo del paquete, como
  `ItemDelSprint` (US-025); no usa el tipo de defecto de US-022. Cuando US-022 y
  US-023 existan, la capa de aplicación arma esta lista con el número del Sprint
  de detección y de resolución de cada defecto.
- **Catálogo de errores (T-005).** `ErrSprintInvalido` y
  `ErrResolucionAnteriorADeteccion` van en las exclusiones de
  `internal/adapters/http/catalogo.go` en el PR de implementación (el test del
  catálogo falla si no están), con el motivo de los otros errores de métricas: son
  datos del sistema que salieron mal, no algo que la persona escribió.
- **Escenarios BDD.** Se agregan al área Métricas de los steps. Los defectos se
  arman en el contexto, sin pasar por US-022, y los Sprints se escriben por su
  número. Frases nuevas, que se aprueban en el PR de implementación:

  | Frase | Expresión |
  |---|---|
  | `Dado un defecto detectado en el sprint 2 y resuelto en el sprint 3` | `^un defecto detectado en el sprint (-?\d+) y resuelto en el sprint (-?\d+)$` |
  | `Y un defecto detectado en el sprint 2 sin resolver` | `^un defecto detectado en el sprint (-?\d+) sin resolver$` |
  | `Cuando se consultan los defectos del sprint 2` | `^se consultan los defectos del sprint (-?\d+)$` |
  | `Entonces los defectos detectados son 3` | `^los defectos detectados son (\d+)$` |
  | `Y los defectos resueltos son 2` | `^los defectos resueltos son (\d+)$` |

  Los números aceptan 0 y negativos a propósito: es la única forma de escribir
  los escenarios de error. Los mensajes se comprueban con `el mensaje de error
  indica "..."`.
- **Nombres en el código.** La función es `CalcularDefectosDelSprint(sprint int, defectos
  []DefectoRegistrado) (ResumenDefectos, error)`, en `internal/domain/metricas/defectos.go`. Los
  mensajes de los errores son "el numero de sprint tiene que ser 1 o mayor" y "el defecto se
  resolvio antes de detectarse", y cada uno se envuelve con el sprint consultado o con la posición
  y los sprints del defecto. Para los escenarios, la interfaz `Metricas` de los steps suma
  `RegistrarDefecto(detectadoEn, resueltoEn int)` y `DefectosDelSprint(sprint int)`, que el
  adaptador en memoria resuelve con esta misma función.
- **Fuera de alcance.** Registrar y cambiar defectos (US-022 y US-023), listarlos
  (US-024), los defectos abiertos acumulados al cierre de cada Sprint, el gráfico
  de evolución (US-032), la pantalla y el desglose por severidad.

## 6. Casos límite

- **CL-029-1** — Lista vacía: `Detectados = 0` y `Resueltos = 0`, sin error.
- **CL-029-2** — Defecto detectado y resuelto en el Sprint consultado: cuenta en
  `Detectados` y en `Resueltos` (RN-029-3).
- **CL-029-3** — Defecto detectado en el Sprint 2 y resuelto en el 4: al
  consultar el 2, `Detectados = 1` y `Resueltos = 0`; al consultar el 4,
  `Detectados = 0` y `Resueltos = 1`; al consultar el 3, `0` y `0`.
- **CL-029-4** — Defecto sin resolver (`SprintResolucion = 0`): cuenta como
  detectado en su Sprint y nunca como resuelto.
- **CL-029-5** — Defectos de otros Sprints: no cuentan. Un defecto detectado en
  un Sprint posterior al consultado tampoco.
- **CL-029-6** — Varios defectos en el mismo Sprint: se suman uno por uno, sin
  deduplicar.
- **CL-029-7** — Más resueltos que detectados en un Sprint (por ejemplo, 1
  detectado y 3 resueltos de Sprints anteriores): es un resultado válido, no un
  error.
- **CL-029-8** — Un defecto resuelto en un Sprint anterior al de su detección:
  `ErrResolucionAnteriorADeteccion` con su posición, y ningún resumen.
- **CL-029-9** — Un defecto con `SprintDeteccion` igual a 0 o negativo:
  `ErrSprintInvalido` con su posición.
- **CL-029-10** — El Sprint consultado igual a 0 o negativo: `ErrSprintInvalido`,
  aunque la lista esté vacía.
- **CL-029-11** — Sprint consultado inválido y dos defectos inválidos a la vez:
  se devuelven los tres errores unidos, en el orden de RN-029-8. `errors.Is`
  reconoce cada uno.
- **CL-029-12** — Una lista `nil` es lo mismo que una lista vacía.

## 7. Condiciones de error

| Situación | Error devuelto | Mensaje al usuario |
|---|---|---|
| El Sprint consultado, o el de detección de un defecto, es menor que 1 | `metricas.ErrSprintInvalido`, envuelto indicando el valor (y la posición del defecto, si es de uno) | "el numero de sprint tiene que ser 1 o mayor" |
| Un defecto resuelto tiene un Sprint de resolución anterior al de detección | `metricas.ErrResolucionAnteriorADeteccion`, envuelto indicando la posición del defecto | "el defecto se resolvio antes de detectarse" |

Con varios datos inválidos, los errores se devuelven unidos con `errors.Join`
(RN-029-8) y se comparan con `errors.Is`. Que no haya defectos no es un error
(CL-029-1).

## 8. Criterios de aceptación

- **CA-029-1** (normal) — Dado un Sprint con tres defectos detectados en él y
  dos resueltos en él, se obtienen 3 detectados y 2 resueltos.
- **CA-029-2** (alternativo) — Un defecto detectado en el Sprint 2 y resuelto
  en el Sprint 4 cuenta como detectado en el 2 y como resuelto en el 4, y en
  ningún otro.
- **CA-029-3** (alternativo) — Un defecto detectado y resuelto en el mismo
  Sprint cuenta en las dos cifras de ese Sprint.
- **CA-029-4** (límite) — Un Sprint sin defectos da 0 detectados y 0 resueltos,
  sin error; un defecto sin resolver cuenta como detectado y nunca como
  resuelto.
- **CA-029-5** (error) — Un defecto resuelto antes de su detección se rechaza
  indicando su posición, y no se devuelve ningún resumen.
- **CA-029-6** (error) — Un número de Sprint igual a 0 o negativo se rechaza.

---

## Decisiones tomadas y descartadas

- **Se cuenta por el Sprint donde ocurrió cada hecho.** Los detectados, por el
  Sprint de detección; los resueltos, por el de resolución. Se descartó contar
  "los detectados en el Sprint que ya están resueltos": reescribiría el pasado de
  un Sprint cuando se resuelve un defecto en uno posterior, y el Sprint de
  resolución es justo uno de los campos que exige el requerimiento 6 de la guía.
- **Sprints por número y no por ID.** Con IDs no se puede saber si una
  resolución es anterior a una detección. El número ya existe en el dominio
  (`Sprint N`) y quien llama lo tiene a mano.
- **La función no depende de US-022 ni de US-023.** Recibe un tipo mínimo propio
  (`DefectoRegistrado`), como US-025 con `ItemDelSprint`. Se descartó importar el
  defecto de US-022: se podría empezar a implementar solo cuando esa historia
  esté mergeada, y esta es una cuenta que no necesita el resto del defecto.
- **"Resuelto" es tener Sprint de resolución.** El estado del defecto (US-023)
  no interviene: la coherencia entre el estado y el Sprint de resolución es de
  esa historia. Si un estado como "descartado" no tiene que contar como
  resuelto, quien arma la lista le pone `SprintResolucion = 0`.
- **Sin "defectos abiertos".** El acumulado (detectados menos resueltos hasta
  cada Sprint) lo necesita el gráfico de evolución (US-032), que puede calcularlo
  con esta función Sprint por Sprint. Se descartó sumarlo acá para no definir una
  regla que ninguna historia pide todavía.
- **Más resueltos que detectados no es un error.** Pasa cuando un Sprint
  corrige defectos de Sprints anteriores, y es una buena noticia.

## Huecos para decidir con el equipo

Ninguno impide implementar esta historia:

1. **Qué estados cuentan como resueltos.** Lo define US-023 (por ejemplo, si
   "cerrado", "rechazado" o "duplicado" cuentan). Esta función cuenta lo que le
   pasan como resuelto.
2. **Defectos sin Sprint de detección.** Un defecto siempre se detecta en algún
   Sprint (campo obligatorio en la plantilla de defectos). Si US-022 permitiera
   que no lo tenga, habría que decidir cómo contarlo.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la
  historia, el requerimiento 6 de la guía, la plantilla de defectos del
  repositorio y las specs de US-025, US-026 y US-027 como modelo; las decisiones
  sobre cómo contar y la validación se proponen acá y se confirman en la revisión
- Revisado y entendido por: Tomás Romero
