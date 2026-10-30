# T-024 — Pulido visual: layout y pantallas principales

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Tomás Romero |
| **Issue** | #238 |
| **Escenarios BDD** | ninguno: es una tarea técnica (ver "Decisiones") |
| **Código** | `web/templates/` (`layout.templ`, `navegacion.templ`, `vistas.go`, `proyectos.templ`, `backlog.templ`, `dashboard.templ`), `web/static/css/input.css` |
| **Última actualización** | 2026-10-30 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

Que la aplicación se vea más cómoda y más cuidada en la presentación, sin cambiar lo que hace. La guía de
la cátedra evalúa una **interfaz usable** (el plan de trabajo le asigna T-011), y T-011 dejó dicho en su
sección 5 que el diseño era «funcional» y que se mejoraría cuando todo el producto funcionara. Hoy
funciona, y mirándolo en las capturas del manual se ven cuatro problemas concretos:

1. **Todo pesa lo mismo.** Las cifras importantes (estado del proyecto, Story Points, horas, defectos) son
   líneas de texto dentro de una caja, sin jerarquía.
2. **El backlog no se lee de un vistazo.** La prioridad y el estado son texto plano en una tabla, y las
   acciones de cada fila (elegir los puntos, «Estimar», «Editar», «Planning Poker») están apiñadas.
3. **Los gráficos del dashboard se desbordan:** el SVG ocupa todo el ancho de la tarjeta y las barras y las
   letras salen enormes, con mucho espacio vacío al lado.
4. **El encabezado no tiene identidad** y no está pensado para una pantalla chica.

Esta tarea mejora **el layout y cuatro pantallas** (Proyectos, Resumen del proyecto, Backlog y Dashboard) y
los estilos compartidos. El resto de las pantallas hereda el layout y los estilos compartidos, y no se
rediseña una por una. **No cambia ninguna funcionalidad.**

## 2. Entradas

No hay entradas nuevas: no hay rutas, formularios ni casos de uso nuevos. Las pantallas reciben los mismos
datos que hoy; cambia cómo se dibujan.

| Pantalla | Ruta | Plantilla |
|---|---|---|
| Proyectos (lista) | `/proyectos` | `proyectos.templ` |
| Resumen del proyecto | `/proyectos/{id}` | `proyectos.templ` |
| Backlog | `/proyectos/{id}/backlog` | `backlog.templ` |
| Dashboard | `/proyectos/{id}/dashboard` | `dashboard.templ` |
| Layout de todas las páginas | — | `layout.templ`, `navegacion.templ` |

## 3. Salidas esperadas

Lo que se ve, por parte:

| Parte | Cómo queda |
|---|---|
| Encabezado | Marca con identidad (nombre y un ícono simple dibujado en el propio HTML/SVG), menú global y menú del proyecto con la sección activa marcada; en pantalla chica el menú se acomoda en filas o se desplaza horizontalmente, sin romper la página |
| Contenido | Ancho máximo cómodo para leer, aire entre bloques, título de página con jerarquía clara |
| Tarjetas | Un componente común (borde suave, relleno y sombra liviana) para los bloques de información |
| Cifras destacadas | En el Resumen y el Dashboard, cada cifra va en su propio recuadro dentro de una tarjeta, y las que llegan como una línea sola (el estado del proyecto, la velocidad y el total de defectos) se dibujan con **tipografía grande** |
| Insignias | Prioridad (must, should, could, wont) y estado (pendiente, en progreso, completado) de las historias: un componente con **color y texto** a la vez. La severidad de los defectos se muestra en la pantalla Defectos, que queda fuera del alcance |
| Botones | Primario y secundario con el mismo estilo en todas las pantallas; las acciones de una fila del backlog agrupadas |
| Tablas | Encabezado diferenciado, filas con separación y resaltado al pasar el mouse, y desplazamiento horizontal **dentro de su contenedor** si no entran |
| Gráficos del dashboard | Tamaño acotado (alto máximo) y proporción estable; los valores siguen visibles como texto, no solo como barra |
| Alertas | Los mensajes de error y de éxito (`#mensajes`) con un estilo de alerta legible |

Las **fuentes** son las del sistema (sin descargar nada) y los **íconos** son SVG en línea: la aplicación
tiene que verse igual sin internet (`web/static/js/LEEME.md`, defensa y evaluación).

## 4. Reglas de negocio

No hay reglas de dominio. Son reglas de la vista.

- **RN-T024-1** — **Nada de lo que hace la aplicación cambia.** No se tocan rutas, controladores, casos de
  uso, dominio ni persistencia; ni los textos visibles que existen hoy, ni los `id`, `name`, `data-campo`,
  `role` y `aria-*` de las plantillas, ni los landmarks `header`, `nav` y `main`, el `aria-current="page"`
  del ítem activo del menú y el contenedor `#mensajes` (RN-T004-4, RN-T004-9 y RN-T004-10). Se puede
  **agregar** texto, ícono o estructura, no quitar ni renombrar. El texto de una cifra («Horas reales: 11,5 h»)
  tampoco se parte en etiqueta y número: los tests lo buscan corrido.
- **RN-T024-2** — **Estilos compartidos, definidos una vez.** Tarjeta, insignia, botón primario y
  secundario, tabla y alerta son constantes (`clase…` en `vistas.go`) o plantillas reutilizables. Las
  pantallas las usan; no repiten cadenas largas de clases.
- **RN-T024-3** — **El color nunca es el único medio.** Cada insignia lleva su texto («Must», «Completado»,
  «Alta»…), y las barras de los gráficos tienen su valor escrito (continúa RN-034-10).
- **RN-T024-4** — **Contraste.** El texto tiene una relación de contraste de al menos **4,5:1** con su
  fondo, y los bordes de campos, botones y partes de los gráficos de al menos **3:1** con el fondo que los
  rodea (WCAG 2.1 AA). Vale también para las insignias, con su color de texto y de fondo.
- **RN-T024-5** — **Foco visible y teclado.** Todo lo interactivo se alcanza con el teclado, en un orden
  lógico, y muestra un foco visible. El enlace «Saltar al contenido» sigue siendo lo primero que se
  alcanza (T-004).
- **RN-T024-6** — **Pantalla chica.** A **375 px** de ancho la página no tiene desplazamiento horizontal;
  una tabla que no entre se desplaza **dentro de su contenedor**, no la página. A **1280 px** el contenido
  queda centrado, con un ancho de línea cómodo.
- **RN-T024-7** — **La impresión no cambia.** Las clases y las reglas de US-034 (`no-imprimir`,
  `solo-impresion`, `reporte`, `@media print`) se conservan tal cual: el reporte imprimible sale igual que
  en los PDF de evidencia.
- **RN-T024-8** — **Sin JavaScript ni dependencias nuevas.** Todo es HTML y Tailwind (v4, ya instalado). Se
  puede definir la paleta y la tipografía en `input.css` (`@theme`). No hay fuentes ni íconos externos.
- **RN-T024-9** — **Se mantiene la identidad.** Sigue el azul/índigo y los grises actuales, con una paleta
  completa para los estados: verde para completado, ámbar para pendiente o atención, rojo para crítico o
  alto, gris para neutro. Las combinaciones elegidas se documentan en `docs/` con su contraste.

## 5. Restricciones

- **Plazo duro.** Esta tarea tiene que estar **mergeada el sábado 31/10 a las 18:00** como máximo, para
  que el código se congele con tiempo de ensayar el domingo y de rehacer las capturas. Si no llega, **no se
  mergea** y la presentación sale con el diseño actual. Por eso el alcance es cerrado.
- **Alcance cerrado:** layout, estilos compartidos y las cuatro pantallas de la sección 2. Sprints,
  Esfuerzo, Defectos, Planning Poker y los reportes **no se rediseñan**; se ven mejor por heredar el layout
  y los estilos compartidos y no se editan salvo para cambiar el nombre de una constante.
- **Los tests existentes pasan sin cambiar lo que comprueban** (textos, roles, `id`, redirecciones). Solo
  pueden ajustarse si dependían de una cadena de clases que cambió; ninguna aserción sobre texto o
  estructura se debilita ni se borra.
- **El CSS se compila donde ya se compila:** `make css` en local y el `Dockerfile` en la imagen. `app.css`
  sigue sin versionarse (`.gitignore`).
- **Sin cambios en el manual por cuenta propia:** las capturas de `docs/manual/` y de `docs/evidencias/`
  se rehacen como último paso de la tarea, con los datos de T-023.

## 6. Casos límite

- **CL-T024-1** — Un nombre de proyecto o de historia muy largo, o una palabra sin espacios: se ajusta o
  se corta con puntos suspensivos dentro de su caja y no rompe el encabezado ni la tabla.
- **CL-T024-2** — Un proyecto sin datos (sin historias, sin sprints, sin defectos): los estados vacíos
  tienen el mismo cuidado que las pantallas con datos y no dejan tarjetas vacías ni gráficos rotos.
- **CL-T024-3** — Un backlog de muchas filas (más de 30): la tabla se desplaza o se alarga sin romper el
  layout, y el encabezado de la tabla se distingue.
- **CL-T024-4** — Un dashboard con un solo sprint cerrado, o con cuatro: los gráficos mantienen su tamaño y
  sus proporciones.
- **CL-T024-5** — Zoom del navegador al 200 % y fuente grande: el contenido se reacomoda sin superponerse
  ni salirse de la pantalla.
- **CL-T024-6** — Navegador sin JavaScript (RN-033-8): las pantallas se ven igual de bien; el menú no
  depende de un script.
- **CL-T024-7** — Modo oscuro del sistema operativo: no se promete; la página se ve con el tema claro como
  hoy.

## 7. Condiciones de error

No hay errores nuevos. Los mensajes de error de validación y de conflicto (T-005) se siguen mostrando en
`#mensajes` con `role="alert"` y el campo en `data-campo`; cambia su apariencia, no su contenido ni su
lugar.

| Situación | Resultado |
|---|---|
| Un formulario con un dato inválido (por ejemplo, una historia sin título) | El mismo error, con estilo de alerta y el contraste de RN-T024-4, arriba y anunciado; el campo señalado conserva lo escrito |
| Una página que da 404 o 500 | Con el mismo layout y estilo que el resto |

## 8. Criterios de aceptación

- **CA-T024-1** (normal) — Las cuatro pantallas (Proyectos, Resumen del proyecto, Backlog y Dashboard) se
  ven con el diseño nuevo, con los mismos datos y los mismos textos que antes, y todos los tests
  existentes pasan sin debilitar ninguna aserción (`make check` en verde).
- **CA-T024-2** (alternativo) — Tarjeta, insignia (prioridad y estado), botón primario y secundario,
  tabla y alerta están definidos una sola vez y las pantallas los usan; el Resumen y el Dashboard muestran
  cada cifra en su recuadro (con tipografía grande la que llega como línea sola), y los gráficos tienen un
  tamaño acotado con sus valores como texto.
- **CA-T024-3** (alternativo) — En el Backlog la prioridad y el estado se leen como insignias con texto y
  las acciones de cada fila están agrupadas; estimar, editar y abrir Planning Poker funcionan igual que
  antes.
- **CA-T024-4** (límite) — A 375 px de ancho ninguna de las cuatro pantallas tiene desplazamiento
  horizontal de la página (las tablas se desplazan dentro de su contenedor), y a 1280 px el contenido queda
  centrado y legible. Se comprueba con capturas a los dos anchos, con los datos de T-023.
- **CA-T024-5** (error) — Un error de validación en un formulario (por ejemplo, una historia sin título) se
  muestra en `#mensajes` con estilo de alerta legible, con `role="alert"` y el campo en `data-campo`, sin
  perder lo escrito.
- **CA-T024-6** (plataforma) — Accesibilidad: el contraste de las combinaciones de color usadas cumple
  RN-T024-4, todo lo interactivo tiene foco visible y se alcanza con el teclado, y Lighthouse (categoría
  Accesibilidad) no informa errores de contraste ni de nombres en ninguna de las cuatro pantallas. La
  impresión de US-034 sigue igual (sus tests pasan y los PDF de evidencia se vuelven a generar), el CSS se
  compila en el `Dockerfile`, y quedan en `docs/evidencias/T-024/` las capturas de antes y después de las
  cuatro pantallas a 1280 px y a 375 px.

---

## Decisiones tomadas y descartadas

- **Alcance cerrado en vez de un rediseño total.** Cambiar las once plantillas a la vez, con un fin de
  semana por delante, invalida las 18 capturas del manual y los PDF de evidencia sin tiempo de rehacerlos.
  El layout y los estilos compartidos mejoran **todas** las pantallas de una vez, y las cuatro que se
  muestran más en la demostración (la lista, el resumen, el backlog y el dashboard) reciben el cuidado
  extra.
- **Plazo duro y «si no llega, no se mergea».** El diseño es lo que menos urgencia tiene y lo que más puede
  romper la presentación. Un límite explícito evita que un cambio visual a medias llegue al congelamiento.
- **Tailwind y no un framework de componentes ni una librería de íconos.** Ya está instalado, no necesita
  JavaScript y la aplicación se defiende sin internet. Se descartó traer Bootstrap, una fuente de Google o
  una librería de íconos.
- **Fuentes del sistema.** Una fuente web sería una descarga más y un parpadeo al cargar; las del sistema
  se ven bien y funcionan sin conexión.
- **Insignias con color y texto.** El color solo no alcanza para quien no distingue colores ni para un
  PDF en blanco y negro (RN-034-10).
- **Constantes de clases y no CSS propio.** El repositorio ya concentra los estilos en `vistas.go`
  (`claseCampo`, `claseBoton`…); se extiende ese patrón en lugar de escribir hojas de estilo aparte.
- **Sin modo oscuro.** Es una tarea en sí misma, con su propia paleta y sus propios contrastes, y no
  aporta a lo que se evalúa.
- **Las pruebas siguen siendo las de siempre.** Lo visual no se puede comprobar con tests de HTML más allá de
  que estén las clases y los textos; por eso los criterios de aspecto se verifican con capturas a dos
  anchos y con Lighthouse, y se guardan como evidencia.

## Descubierto al implementar

- **Las cifras no se parten.** Cada línea de una tarjeta («Story Points: 5 completados de 12») llega como un
  solo texto y los tests de las pantallas lo buscan corrido, así que no se puede poner el número en grande
  y la etiqueta chica sin debilitar esos tests, lo que la spec prohíbe (RN-T024-1). La jerarquía sale de los
  recuadros, de la tipografía de las líneas que ya llegan solas y de las insignias.
- **El ícono de la marca son tres `<span>` y no un SVG.** Un test de US-030 comprueba que un proyecto sin
  datos no dibuja ningún gráfico buscando `<svg`, y un ícono en la cabecera lo rompería en todas las páginas.
- **Los bordes de los campos y de los botones secundarios pasan de `slate-300` a `slate-500`:** el primero
  tiene cerca de 1,5:1 con el blanco y RN-T024-4 pide 3:1. El resto de las pantallas lo hereda porque usan las
  mismas constantes.
- **La paleta usada** es índigo (principal y «must»), celeste («should» y «en progreso»), verde (completado), ámbar
  (pendiente), gris (neutro y «wont») y rojo (alertas). El contraste de cada par se calcula con
  `scripts/contraste.py` sobre el CSS compilado y está en `docs/diseno.md`.
- **Lighthouse no se pudo correr desde la herramienta** (descargaba código externo). Se comprobó por script, en
  las cuatro pantallas, lo que suele marcar primero: idioma, un solo `h1`, encabezados sin saltos, landmarks,
  `aria-current`, campos con etiqueta y enlaces o botones con nombre. La medición de Lighthouse queda como paso
  manual en Chrome (`docs/diseno.md`).
- **La impresión no cambió:** el reporte del proyecto da 5 hojas A4 con el texto idéntico, hoja por hoja, antes
  y después del diseño.

## Huecos para decidir con el equipo

Ninguno impide implementar esta tarea:

1. **Paleta de estados.** Verde, ámbar, rojo y gris son una propuesta; si el equipo prefiere otra, se
   cambia la tabla de `docs/` y las constantes, no la estructura.
2. **Ícono de la marca.** Un ícono simple dibujado en SVG (por ejemplo, un gráfico de barras) es una
   propuesta; sin él, queda solo el nombre.
3. **Qué pasa con las capturas si la tarea no llega.** Si no se mergea, el manual conserva las capturas
   actuales; si se mergea, se rehacen todas (CA-T024-6) y es trabajo de Vergara.

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de las capturas del manual, el
  código de las plantillas y de los estilos compartidos, y las specs de T-004, T-011 y US-034; el alcance
  y el plazo los decidió el equipo y las decisiones de diseño se proponen acá y se confirman en la revisión
- Revisado y entendido por: Tomás Romero
