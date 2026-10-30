# T-023 — Datos de demostración: cargar un proyecto completo con un comando

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #236 |
| **Escenarios BDD** | ninguno: es una tarea técnica (ver "Decisiones") |
| **Código** | `internal/demo/`, `cmd/demo/`, `docs/ci-cd.md` |
| **Última actualización** | 2026-10-30 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Que el equipo pueda llenar una base **vacía** con un proyecto de demostración completo, en un solo
comando y siempre con los mismos datos, para mostrar en la presentación **todas** las funcionalidades
y los requerimientos de la guía con información que se vea real: backlog con prioridades, estimaciones,
Planning Poker, sprints cerrados y uno en curso, horas, defectos, dashboard y reportes.

Hoy el único sembrado es `migraciones/datos/ejemplo.sql` (`CARGAR_DATOS_EJEMPLO`): un proyecto con
sus integrantes y nada más, así que el dashboard, los reportes, el esfuerzo y los defectos se verían
vacíos («Sin datos»). Esta tarea **no reemplaza** ese sembrado: agrega un comando aparte, que se corre
una vez a mano contra la base de la demostración (la de Neon, o la local para el plan B).

La carga usa **los casos de uso de la aplicación**, los mismos que usa la pantalla. No inserta filas con
SQL: así los datos cumplen todas las reglas del dominio (Definition of Ready, sprints que no se
solapan, un solo sprint abierto, horas que no pasan de hoy, transiciones de defectos) por construcción.

## 2. Entradas

Un comando, sin parámetros posicionales:

```
go run ./cmd/demo            # carga el proyecto de demostración
go run ./cmd/demo --borrar   # borra el proyecto de demostración y todo lo suyo
```

| Nombre | Tipo | Obligatorio | Rango / formato válido |
|---|---|---|---|
| `DATABASE_URL` (variable de entorno) | cadena de conexión de Postgres | sí | La misma que usa el servidor. No se acepta por argumento, para que no quede en el historial de la terminal |
| `--borrar` | opción | no | Sin valor. Sola, no carga nada |

El comando aplica las migraciones pendientes antes de cargar (igual que el servidor), así que sirve
sobre una base recién creada.

## 3. Salidas esperadas

**Salida estándar** (código de salida `0`): un resumen de lo cargado, sin la cadena de conexión:

```
Proyecto "Tienda online" cargado: 4 integrantes, 16 historias, 4 sprints (3 cerrados y 1 abierto), 3 sesiones de Planning Poker, 27 registros de horas y 10 defectos.
```

**Código de salida `1`** y un mensaje en español en el error estándar cuando algo falla (sección 7).

**Contenido de la carga.** Todos los datos son ficticios y fijos: dos corridas sobre bases vacías dejan
las mismas filas. Fecha de corte de los datos: **2026-10-30** (ninguna fecha cargada es posterior).

*Proyecto:* «Tienda online», del 2026-09-14 al 2026-12-18. Descripción: «Sitio de ventas con catálogo,
carrito y pago. Proyecto de demostración».

*Integrantes (4):* Ana Martínez (`agile_enabler`), Bruno Díaz (`product_architect`), Carla Ruiz y Diego
Sosa (`product_builder`), con correos `…@tienda.test`.

*Historias (16), con su estado final:*

| # | Título | Prioridad | SP | Final |
|---|---|---|---|---|
| 1 | Registrarse e iniciar sesión | must | 5 | completada en el Sprint 1 |
| 2 | Ver el catálogo de productos | must | 5 | completada en el Sprint 1 |
| 3 | Buscar productos por nombre | must | 3 | no completada en el Sprint 1; completada en el 2 |
| 4 | Agregar productos al carrito | must | 5 | completada en el Sprint 2 |
| 5 | Pagar con tarjeta | must | 8 | completada en el Sprint 3 |
| 6 | Filtrar el catálogo por categoría | should | 3 | no completada en el Sprint 2; completada en el 3 |
| 7 | Guardar productos favoritos | should | 2 | en el backlog |
| 8 | Recibir el comprobante por email | should | 3 | no completada en el Sprint 3; completada en el 4 |
| 9 | Ver el historial de compras | should | 5 (por Planning Poker) | en el Sprint 4, sin completar |
| 10 | Calificar un producto | should | 2 | completada en el Sprint 4 |
| 11 | Cupones de descuento | could | sin estimar (Planning Poker revelado con divergencia) | en el backlog |
| 12 | Compartir un producto en redes | could | 2 | en el Sprint 4, sin completar |
| 13 | Modo oscuro | could | 3 | en el backlog |
| 14 | Aplicación móvil nativa | wont | sin estimar | en el backlog |
| 15 | Programa de puntos | wont | sin estimar | en el backlog |
| 16 | Recomendar productos similares | should | sin estimar (Planning Poker votando) | en el backlog |

*Sprints (de 14 días, seguidos):*

| Sprint | Fechas | Estado | Historias | Planificados | Completados |
|---|---|---|---|---|---|
| 1 | 2026-09-14 a 2026-09-27 | cerrado | 1, 2, 3 | 13 | 10 |
| 2 | 2026-09-28 a 2026-10-11 | cerrado | 3, 4, 6 | 11 | 8 |
| 3 | 2026-10-12 a 2026-10-25 | cerrado | 5, 6, 8 | 14 | 11 |
| 4 | 2026-10-26 a 2026-11-08 | **abierto** | 8, 9, 10, 12 | 12 | 5 (historias 8 y 10) |

*Planning Poker (3 sesiones, con los 4 integrantes):* historia 9, con votos 3, 5, 8 y 5 (hay
divergencia), nueva ronda con 5, 5, 5 y 5 y estimación acordada de 5 (la sesión queda finalizada y la
historia estimada); historia 11, con votos 3, 8, 13 y 5, revelados y con divergencia, sin nueva ronda;
historia 16, votando, con solo los votos de Ana y Carla (los demás sin votar).

*Horas (27 registros, múltiplos de 0,25 h, con fecha dentro de cada sprint y las cinco actividades
«Desarrollo», «Pruebas», «Diseño», «Revisión de código» y «Reunión»), con estas desviaciones contra las
horas estimadas (1 Story Point = 1,5 h, US-027):*

| Sprint | Estimadas | Reales | Desviación |
|---|---|---|---|
| 1 | 19,5 h | 24,5 h | +5 h (+25,6 %) |
| 2 | 16,5 h | 14,75 h | −1,75 h (−10,6 %) |
| 3 | 21 h | 21 h | 0 h (0 %) |
| 4 | 18 h | 11,5 h (hasta el 30/10) | en curso |

*Defectos (10), detectados en los Sprints 1 a 4:* críticos 2, altos 2, medios 3 y bajos 3; abiertos 4, en
corrección 1, resueltos 4 (en el sprint donde se resolvieron: 2, 2, 3 y 3) y descartados 1.

## 4. Reglas de negocio

No hay reglas de dominio nuevas. Son reglas de la herramienta.

- **RN-T023-1** — La carga llama a los casos de uso de la aplicación (`CrearProyecto`,
  `RegistrarIntegrante`, `CrearItemBacklog`, `EstimarItem`, `CrearSprint`, `AsignarItemASprint`,
  `MarcarCompletado`, `CerrarSprint`, `IniciarSesionPlanningPoker` y los de votar, revelar, abrir una
  nueva ronda y registrar el acuerdo, `RegistrarEsfuerzo`, `RegistrarDefecto` y `CambiarEstadoDefecto`).
  No ejecuta `INSERT` ni `UPDATE` propios y no usa el dominio directamente.
- **RN-T023-2** — **Reloj controlado.** Los casos de uso que reciben un reloj (`CrearItemBacklog`,
  `MarcarCompletado` y `RegistrarEsfuerzo`) reciben el de la herramienta, que antes de cada paso lo fija
  en una fecha del sprint que se está armando. Así las fechas de completado y de las horas quedan dentro
  de cada sprint aunque se cargue otro día. Ninguna fecha es posterior al 2026-10-30.
- **RN-T023-3** — **Orden de carga.** Proyecto, integrantes, historias, estimaciones y Planning Poker;
  después, sprint por sprint: crear, asignar, completar, registrar horas y defectos, y cerrar, hasta dejar
  el cuarto abierto. Un sprint se crea recién cuando el anterior está cerrado (US-008).
- **RN-T023-4** — **No pisa nada.** Si ya existe un proyecto llamado «Tienda online», el comando se niega
  antes de cargar ningún dato y explica cómo seguir (`--borrar`). Lo único que hace antes es aplicar las
  migraciones pendientes (sección 2). Otros proyectos de la base no se tocan.
- **RN-T023-5** — `--borrar` borra **solo** el proyecto con el nombre exacto «Tienda online», y con él todo
  lo suyo (las claves foráneas borran en cascada). Si no existe, lo informa y termina con `0`.
- **RN-T023-6** — Si un paso de la carga falla, el comando se detiene, informa **qué paso** y el error, y
  termina con `1`. Lo ya cargado queda; para volver a empezar se usa `--borrar`.
- **RN-T023-7** — El comando **nunca imprime ni guarda** `DATABASE_URL`: ni en los mensajes ni en los
  errores. Si la conexión falla, el mensaje no repite la cadena.
- **RN-T023-8** — El contenido es **determinista**: no hay valores al azar ni dependientes del día en que
  se corre.
- **RN-T023-9** — Los datos son ficticios: nombres inventados y correos del dominio reservado `.test`.
  No hay datos de personas reales.

## 5. Restricciones

- **Sin cambios en el dominio ni en los casos de uso existentes**, sin migraciones nuevas y sin
  dependencias nuevas.
- **Dos paquetes:** `internal/demo` tiene la carga (recibe los casos de uso ya armados y un reloj
  modificable, para probarla sin base) y `cmd/demo` solo lee `DATABASE_URL`, conecta, migra y la llama,
  con el mismo armado de repositorios y casos de uso que `cmd/server`.
- **No corre sola.** El servidor no llama a esta carga al arrancar: es una orden manual. El contenedor
  de la demostración sigue con `CARGAR_DATOS_EJEMPLO=false`.
- **Se prueba con memoria y con Postgres,** como el resto de los adaptadores: la misma verificación
  sobre `memoria.Base` y, con `DATABASE_URL`, sobre una base real.
- **Documentación:** `docs/ci-cd.md` explica cómo cargar y borrar los datos en Neon y en local, y el plan
  B de la presentación (la aplicación local «con los mismos datos») apunta a este comando.

## 6. Casos límite

- **CL-T023-1** — Base recién migrada y sin proyectos: carga todo y termina con `0`.
- **CL-T023-2** — La base ya tiene otro proyecto distinto (por ejemplo, uno de prueba): se carga «Tienda
  online» al lado y el otro no cambia.
- **CL-T023-3** — Ya existe «Tienda online»: se niega y no carga ningún dato (RN-T023-4).
- **CL-T023-4** — `--borrar` sin que exista el proyecto: informa que no había nada y termina con `0`.
- **CL-T023-5** — `--borrar` y después cargar: la base queda con los mismos datos que en una carga limpia
  (el borrado en cascada no deja filas huérfanas).
- **CL-T023-6** — Una carga que se corta a la mitad (por ejemplo, se cae la red): lo ya cargado queda, el
  mensaje dice el paso, y `--borrar` más una carga nueva la completan.
- **CL-T023-7** — La historia 3 vuelve al backlog al cerrar el Sprint 1 y se asigna al Sprint 2 (lo mismo
  la 6 y la 8): la historia aparece en dos sprints, con su resultado en cada uno (RN-009-13, US-012).
- **CL-T023-8** — Un defecto se detecta en un sprint cerrado y se resuelve en uno posterior (RN-023-4: el
  sprint de resolución no es anterior al de detección); un defecto descartado queda fuera de los totales
  por sprint (RN-030-6).

## 7. Condiciones de error

| Situación | Resultado | Mensaje al usuario |
|---|---|---|
| Falta `DATABASE_URL` | Sale con `1` sin tocar nada | «Falta la variable DATABASE_URL.» |
| No se puede conectar o migrar | Sale con `1` | «No se pudo conectar a la base: …» (sin la cadena de conexión) |
| Ya existe el proyecto «Tienda online» | Sale con `1` sin cargar ningún dato | «Ya existe el proyecto "Tienda online". Para cargarlo de nuevo, primero: go run ./cmd/demo --borrar» |
| Un caso de uso rechaza un dato (un error de la propia carga) | Sale con `1` y deja lo ya cargado | «Falló el paso "<paso>": <error>. Podés volver a empezar con --borrar.» |
| Argumento desconocido | Sale con `1` | «Uso: go run ./cmd/demo [--borrar]» |

## 8. Criterios de aceptación

- **CA-T023-1** (normal) — Con una base vacía, el comando carga «Tienda online» con 4 integrantes, 16
  historias, 4 sprints (3 cerrados y el Sprint 4 abierto), 3 sesiones de Planning Poker, 27 registros de
  horas y 10 defectos, y termina con `0` y el resumen de la sección 3.
- **CA-T023-2** (alternativo) — Las lecturas de la aplicación sobre esos datos dan lo que dice la sección
  3: los Story Points planificados y completados de cada sprint, las horas estimadas, reales y su
  desviación, los defectos por severidad y por estado, la historia 9 estimada en 5, la 11 con divergencia
  y la 16 votando.
- **CA-T023-3** (alternativo) — Ninguna sección de los reportes del proyecto y del Sprint 4 dice «Sin
  datos», y el dashboard muestra los cuatro sprints, la velocidad, las horas y los defectos.
- **CA-T023-4** (límite) — Con el proyecto ya cargado, una segunda carga se niega sin cargar ningún dato;
  `--borrar` lo borra solo a él (otro proyecto de la base queda igual) y una carga nueva deja los mismos
  datos.
- **CA-T023-5** (error) — Sin `DATABASE_URL`, con una conexión que falla o con un argumento desconocido,
  el comando termina con `1`, con un mensaje claro y sin imprimir la cadena de conexión; si un caso de uso
  rechaza un paso, el mensaje nombra el paso.
- **CA-T023-6** (plataforma) — La carga se prueba sobre `memoria.Base` y, con `DATABASE_URL`, sobre
  Postgres, con las mismas comprobaciones de CA-T023-1 a CA-T023-4; y se corre una vez contra la base de
  la demostración (Neon) y otra en local, comprobando en el navegador el dashboard, los reportes y el
  Planning Poker.

---

## Decisiones tomadas y descartadas

- **Usar los casos de uso y no SQL.** Un `.sql` con `INSERT` es más corto, pero puede dejar datos que la
  aplicación no permitiría (un sprint abierto y otro más, una historia sin estimar en un sprint, horas del
  futuro) y habría que mantenerlo en paralelo con cada regla nueva. Con los casos de uso, si una regla
  cambia, la carga falla en el test y no en la presentación. Se descartó extender `ejemplo.sql`.
- **Un comando aparte y no `CARGAR_DATOS_EJEMPLO`.** Ese sembrado corre en cada arranque y tiene que ser
  idempotente; esta carga es larga, con orden y con fechas, y solo se corre una vez. Se mantiene el
  sembrado mínimo para los entornos que ya lo usan.
- **Reloj controlado.** Los sprints cerrados tienen que tener historias completadas y horas con fechas
  de **dentro** del sprint, y la aplicación no deja registrar horas del futuro. Con el reloj real habría
  que cargar el mismo día en que ocurrieron los hechos.
- **Datos ficticios de una tienda online** y no los del proyecto real del equipo: el dominio es
  conocido por cualquiera que mire la demostración, y no mezcla la información del TP con la del
  ejemplo.
- **Cuatro sprints, tres cerrados y uno abierto.** Son los mínimos para que la velocidad, el
  porcentaje de historias completadas, las historias que vuelven al backlog y el sprint en curso tengan
  algo que mostrar. Las cifras elegidas (+25,6 %, −10,6 % y 0 % de desviación) muestran los tres casos de
  la comparación de horas.
- **Números exactos en la spec.** La spec fija la cantidad de cada cosa para que el test pueda
  comprobarla; los títulos y las descripciones de cada defecto y de cada criterio son de la
  implementación (`internal/demo`), y cualquier cambio de cifras se refleja acá.
- **Borrar solo por nombre.** `--borrar` no es un «vaciar la base»: la cuenta y la base de Neon se
  comparten con la presentación, y borrar de más no tiene vuelta.

## Huecos para decidir con el equipo

Ninguno impide implementar esta tarea:

1. **Nombre y rubro del proyecto.** «Tienda online» es una propuesta; si el equipo prefiere otro
   (por ejemplo, el proyecto real del TP), cambian la sección 3 y los textos de la carga, no la forma.
2. **Cantidad de horas.** Los 27 registros y las desviaciones son un diseño para mostrar los tres casos;
   se pueden ajustar siempre que cada sprint conserve su desviación.
3. **Ejecutar la carga en el contenedor.** Hoy se corre desde una computadora del equipo contra Neon,
   porque el contenedor no tiene el código fuente ni el comando. Si más adelante se quiere correr desde
   el CD, haría falta incluir `cmd/demo` en la imagen.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la guía de la cátedra, las
  specs de US-008 a US-011, US-014 a US-019 y US-022 a US-023 (reglas que ordenan la carga), el armado de
  `cmd/server` y el sembrado actual (`ejemplo.sql`); las cifras de la demostración se proponen acá y se
  confirman en la revisión
- Revisado y entendido por: Tomás Romero
