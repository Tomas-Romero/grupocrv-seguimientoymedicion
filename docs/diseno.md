# Diseño de la interfaz

Guía corta de cómo se ve la aplicación y por qué. Está especificada en [`specs/T-024-pulido-visual.md`](../specs/T-024-pulido-visual.md). Los
estilos son clases de Tailwind definidas **una sola vez** en [`web/templates/vistas.go`](../web/templates/vistas.go); las pantallas las usan y no
repiten cadenas de clases. No hay JavaScript propio, ni fuentes web, ni íconos externos: la aplicación se ve igual sin internet.

## Piezas

| Pieza | Para qué | Dónde |
|---|---|---|
| Tarjeta | Un bloque de información (borde suave, relleno y sombra liviana) | `claseTarjeta` |
| Cifra (recuadro) | Cada dato suelto de una tarjeta; su texto no se parte | `claseDato` |
| Insignia | Prioridad y estado de una historia, con **color y texto** a la vez | `insignia`, `claseDeInsignia` |
| Botones | Primario, secundario y sus versiones chicas | `claseBoton…` |
| Tabla | Dentro de un contenedor que se desplaza si no entra, para que la página nunca lo haga | `claseTablaContenedor`, `claseTabla` |
| Menú | La sección activa se distingue por **más que el color**: fondo, negrita y subrayado | `Enlace.Clases` |
| Alerta | Los errores de `#mensajes`, con borde rojo a la izquierda | `web/static/css/input.css` |

## Colores y contraste

El color nunca es el único medio: cada insignia lleva su texto y su tono en `data-tono`. La prioridad usa índigo (`must`), celeste
(`should`), gris (`could`) y gris apagado (`wont`); el estado usa verde (completado), ámbar (pendiente) y celeste (en progreso).

El contraste de cada par se calcula con `scripts/contraste.py` sobre el CSS **compilado** (`make css`): el texto tiene que llegar a 4,5:1 y
los bordes de campos, botones y gráficos a 3:1 (WCAG 2.1 AA). El script sale con error si algún par queda debajo.

```bash
make css
python scripts/contraste.py
```

| Qué | Colores | Contraste | Mínimo |
|---|---|---|---|
| insignia primaria | `indigo-900` sobre `indigo-100` | 9,31:1 | 4,5:1 |
| insignia info | `sky-900` sobre `sky-100` | 8,24:1 | 4,5:1 |
| insignia ok (completado) | `emerald-900` sobre `emerald-100` | 8,53:1 | 4,5:1 |
| insignia atencion (pendiente) | `amber-900` sobre `amber-100` | 8,17:1 | 4,5:1 |
| insignia apagada (wont) | `slate-600` sobre `white` | 7,56:1 | 4,5:1 |
| insignia neutra (could) | `slate-800` sobre `slate-100` | 13,37:1 | 4,5:1 |
| boton primario | `white` sobre `indigo-700` | 8,07:1 | 4,5:1 |
| boton primario (hover) | `white` sobre `indigo-800` | 10,09:1 | 4,5:1 |
| boton secundario | `slate-800` sobre `white` | 14,66:1 | 4,5:1 |
| enlace sobre blanco | `indigo-700` sobre `white` | 8,07:1 | 4,5:1 |
| enlace sobre gris claro | `indigo-700` sobre `slate-50` | 7,71:1 | 4,5:1 |
| enlace (hover) | `indigo-900` sobre `white` | 11,47:1 | 4,5:1 |
| menu activo | `indigo-800` sobre `indigo-50` | 9,02:1 | 4,5:1 |
| menu inactivo sobre blanco | `slate-700` sobre `white` | 10,34:1 | 4,5:1 |
| menu inactivo sobre la barra del proyecto | `slate-700` sobre `slate-50` | 9,88:1 | 4,5:1 |
| menu (hover) | `slate-900` sobre `slate-100` | 16,28:1 | 4,5:1 |
| chip del proyecto | `slate-600` sobre `white` | 7,56:1 | 4,5:1 |
| encabezado de tabla | `slate-600` sobre `slate-50` | 7,23:1 | 4,5:1 |
| numero de fila | `slate-600` sobre `white` | 7,56:1 | 4,5:1 |
| cifra de una tarjeta | `slate-800` sobre `slate-50` | 14,00:1 | 4,5:1 |
| titulo de la velocidad | `indigo-800` sobre `white` | 10,09:1 | 4,5:1 |
| alerta | `red-900` sobre `red-50` | 9,21:1 | 4,5:1 |
| campo: borde | `slate-500` sobre `white` | 4,77:1 | 3:1 |
| boton secundario: borde | `slate-500` sobre `white` | 4,77:1 | 3:1 |
| barra de Story Points planificados | `slate-500` sobre `white` | 4,77:1 | 3:1 |
| barra de Story Points completados | `indigo-700` sobre `white` | 8,07:1 | 3:1 |

## Cómo medir la accesibilidad con Lighthouse

En Chrome: abrir la pantalla, `F12` → pestaña **Lighthouse** → categoría **Accesibilidad** → *Analyze page load*. Repetir en el Resumen, el
Backlog, el Dashboard y la lista de Proyectos, con los datos de la demostración (`go run ./cmd/demo`), y anotar el resultado en el PR.

## Evidencia

En [`docs/evidencias/T-024/`](evidencias/T-024/) están las capturas de las cuatro pantallas antes y después del diseño, a 1280 px y a 375 px de
ancho, con los datos de la demostración: `antes-<pantalla>-<ancho>.png` y `despues-<pantalla>-<ancho>.png`.
