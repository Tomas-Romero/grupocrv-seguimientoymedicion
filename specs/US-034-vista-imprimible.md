# US-034 — Imprimir el reporte o guardarlo como PDF desde el navegador

| | |
|---|---|
| **Estado** | borrador |
| **Autor** | Juan Ignacio Vergara |
| **Issue** | #40 |
| **Escenarios BDD** | ninguno (ver "Decisiones") |
| **Código** | `web/static/css/input.css`, `web/static/js/`, `web/templates/`, `internal/adapters/http/` |
| **Última actualización** | 2026-10-27 |

> Esta especificación se escribe y se mergea **antes** de abrir la rama de
> implementación. Si durante la implementación descubrís que algo de acá está
> mal, corregís la spec en el mismo PR — no la dejás desactualizada.

## 1. Objetivo

> Como integrante del equipo, quiero imprimir el reporte o guardarlo como PDF,
> para entregarlo o archivarlo fuera del sistema.

La cátedra confirmó el 26/10 que alcanza con que el navegador imprima la página
como PDF (US-033, sección 1). Esta historia suma, sobre los dos reportes de
US-033, **una hoja de estilos de impresión y un botón**: no genera archivos en el
servidor, no agrega rutas ni casos de uso y no cambia las secciones ni su
contenido (RN-033-8).

Depende de **US-033** (los reportes, sus rutas y sus secciones).

## 2. Entradas

No hay rutas nuevas. Aplica a las dos páginas de US-033:

| Ruta | Método | Caso de uso |
|---|---|---|
| `/proyectos/{id}/reporte` | GET | `ConsultarReporteDelProyecto` (US-033) |
| `/proyectos/{id}/sprints/{sprint}/reporte` | GET | `ConsultarReporteDelSprint` (US-033) |

La entrada es la orden de imprimir del navegador: el botón de la página o el
menú del navegador (`Ctrl+P`). Las dos dan el mismo resultado.

## 3. Salidas esperadas

**En pantalla:** el reporte de US-033 sin cambios, más un botón **"Imprimir o
guardar como PDF"** arriba, junto al título.

**Al imprimir** (vista de impresión, hoja A4 vertical):

| Parte | Qué se ve |
|---|---|
| Encabezado de la impresión | El nombre del proyecto (y del sprint, en su reporte), el título del reporte y "Generado el dd/mm/aaaa" |
| Navegación del sitio, enlace "Saltar al contenido", botones y enlaces de acciones | No se imprimen |
| Cada sección de US-033 | Empieza en una página nueva, con su título |
| Tablas | Sin filas partidas; si una tabla no entra en una página, sigue en la siguiente repitiendo el encabezado |
| Colores | Texto oscuro sobre fondo blanco; nada depende de un fondo de color |

El usuario elige en el diálogo del navegador entre la impresora y "Guardar como
PDF". El PDF lo genera el navegador.

## 4. Reglas de negocio

No hay reglas de dominio. Son reglas de la vista.

- **RN-034-1** — El botón "Imprimir o guardar como PDF" llama a
  `window.print()` y nada más: no hace peticiones ni cambia la página.
- **RN-034-2** — El comportamiento del botón vive en un archivo estático
  (`web/static/js/imprimir.js`), no en un atributo `onclick` ni en un `<script>`
  dentro de la plantilla, igual que `htmx.min.js`.
- **RN-034-3** — El botón se dibuja oculto (`hidden`) y el script lo muestra. Sin
  JavaScript no aparece un botón que no hace nada y la página sigue funcionando
  como pide RN-033-8: se imprime con el menú del navegador y sale igual.
- **RN-034-4** — Los estilos de impresión están en la hoja de estilos del sitio
  (`input.css`, con `@media print` o la variante `print:` de Tailwind), no en un
  archivo aparte ni en una ruta `/imprimir`. Ocultar la cabecera del layout vale
  para todo el sitio (no molesta en ninguna página); los saltos de página, el
  encabezado de impresión y las reglas de las tablas solo se usan en los reportes.
- **RN-034-5** — Al imprimir se ocultan la cabecera del layout (marca y
  navegación), el enlace "Saltar al contenido", el área `#mensajes`, el botón de
  imprimir y los enlaces de navegación que tenga el reporte. Lo que se oculta se marca con una clase que dice qué es
  ("no se imprime"), no con selectores de la estructura del layout.
- **RN-034-6** — Cada `<section>` de US-033 empieza en una página nueva
  (`break-before: page`), salvo la primera. El título de una sección no queda
  solo al pie de una página (`break-after: avoid` en los títulos).
- **RN-034-7** — Una fila de tabla no se parte entre dos páginas
  (`break-inside: avoid` en `tr`). Una tabla que entra en una página no se parte;
  una que no entra se reparte entre filas y repite su encabezado en cada página
  (`thead` con `display: table-header-group`).
- **RN-034-8** — El tamaño de página es A4 vertical con márgenes fijos
  (`@page { size: A4; margin: ... }`). El contenido usa todo el ancho de la hoja:
  sin el ancho máximo ni el relleno de pantalla.
- **RN-034-9** — El encabezado de la impresión (sección 3) es un elemento del
  reporte que se ve solo al imprimir. La fecha es la del día de la visita, con el
  formato `dd/mm/aaaa` de US-033, y la arma el handler con el mismo reloj que ya
  recibe el servidor (`func() time.Time`). La plantilla no decide.
- **RN-034-10** — Ningún dato se pierde al imprimir: lo que en pantalla depende
  del color (por ejemplo un estado o una divergencia) también tiene texto, y los
  fondos de color no son necesarios para leer.

## 5. Restricciones

- **Sin cambios en el dominio, en `internal/app` ni en los repositorios.** Sin
  migración ni errores nuevos: los errores de los reportes son los de US-033.
- **Sin dependencias nuevas.** Se descartó una librería de PDF (US-033,
  "Decisiones").
- La página de US-033 tiene que estar en `main` antes de abrir la rama de
  implementación; el botón y el encabezado se agregan a sus plantillas.
- `app.css` se regenera con `make css` y se commitea, como en el resto del sitio.
- **Evidencia (CA-034-4):** una captura del PDF de cada reporte, guardada en
  `docs/evidencias/US-034/`, generada desde Chrome o Edge con "Guardar como PDF".
- **Fuera de alcance:** generar el PDF en el servidor, numerar las páginas desde
  la aplicación (lo hace el navegador si el usuario lo elige), un índice,
  orientación horizontal y diseñar la impresión de otras páginas.

## 6. Casos límite

- **CL-034-1** — Un reporte corto (proyecto recién creado, todo «Sin datos»): se
  imprime con una sección por página y sin páginas en blanco.
- **CL-034-2** — Un backlog o una lista de defectos de más de una página: la tabla
  sigue en la página siguiente con su encabezado y sin filas cortadas.
- **CL-034-3** — Una fila con un texto largo (descripción de un defecto): la fila
  entera pasa a la página siguiente; el texto se ajusta al ancho y no se corta a
  la derecha.
- **CL-034-4** — Una sección que falló (RN-033-3): se imprime con su aviso de
  error, que es parte del contenido y no del área `#mensajes`.
- **CL-034-5** — El navegador sin JavaScript: no se ve el botón y `Ctrl+P`
  imprime igual.
- **CL-034-6** — Imprimir con `Ctrl+P` sin usar el botón: mismo resultado que con
  el botón.
- **CL-034-7** — Un reporte que da 404 (US-033): la página de error no tiene
  botón ni estilos de impresión propios.

## 7. Condiciones de error

No hay errores nuevos. Los del reporte son los de US-033 (sección 7).

| Situación | Respuesta |
|---|---|
| Proyecto o sprint inexistente, o sprint de otro proyecto | 404 con layout (US-033), sin botón de imprimir |
| El usuario cancela el diálogo de impresión | Nada: la página queda como estaba |

## 8. Criterios de aceptación

- **CA-034-1** (normal) — El reporte tiene un botón "Imprimir o guardar como PDF"
  que abre el diálogo de impresión del navegador.
- **CA-034-2** (alternativo) — La vista de impresión oculta la navegación y los
  botones, usa saltos de página entre secciones y no parte las tablas por la
  mitad, en A4.
- **CA-034-3** (límite) — Un reporte largo se reparte en varias páginas
  repitiendo los encabezados de las tablas.
- **CA-034-4** (plataforma) — Queda como evidencia en la documentación una
  captura del PDF generado desde el navegador, y se comprueba que el archivo es
  legible.

---

## Decisiones tomadas y descartadas

- **El navegador genera el PDF.** Es lo que aceptó la cátedra y lo que dejó
  preparado US-033. Se descartó una librería de PDF en el servidor.
- **Sin escenarios BDD.** No hay comportamiento del dominio: la historia es una
  hoja de estilos y un botón. Lo que se puede automatizar va en tests de Go con
  `httptest` sobre las páginas de reporte: el botón está, con su texto, oculto y
  enlazado a `imprimir.js`; el encabezado de impresión tiene el proyecto y la
  fecha; lo que no se imprime lleva su clase. El resultado impreso se verifica a
  mano con la captura de CA-034-4.
- **"No partir las tablas por la mitad" (CA-034-2) y CA-034-3.** Una tabla más
  larga que una hoja tiene que partirse: se lee como "no cortar filas y no
  partir una tabla que entra en una página" (RN-034-7), y la que no entra repite
  el encabezado, que es lo que pide CA-034-3.
- **Una sección por página.** Es lo que pide CA-034-2 ("saltos de página entre
  secciones"). Se descartó dejar que el navegador corte donde caiga: una tabla
  quedaría con el título en una hoja y las filas en otra.
- **El botón solo con JavaScript.** RN-033-8 pide que la página funcione sin
  JavaScript; un botón oculto hasta que el script lo muestra cumple eso y evita
  un botón muerto. Se descartó un enlace a una página aparte para imprimir.
- **Encabezado con proyecto y fecha.** Una hoja suelta del PDF tiene que decir de
  qué proyecto es y de qué día son las cifras, porque el reporte se calcula en
  cada visita (RN-033-7).

## Uso de IA en esta especificación

- [ ] No use IA
- [x] Use IA para: redacción de la especificación con Claude Code a partir de la
  historia y de las decisiones del equipo (imprimir desde el navegador, hoja de
  estilos `@media print` sobre el reporte de US-033, ocultar navegación y
  botones, no partir tablas, encabezado con proyecto y fecha, botón con
  `window.print`), usando las specs de US-033 y T-016 como modelo
- Revisado y entendido por:
