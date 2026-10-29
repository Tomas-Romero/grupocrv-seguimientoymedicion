# JavaScript de terceros

Se sirven desde el propio servidor, sin CDN, para que la aplicacion funcione sin
internet (defensa y evaluacion).

| Archivo | Version | Origen | SHA-256 |
|---|---|---|---|
| htmx.min.js | 2.0.4 | https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js | e209dda5c8235479f3166defc7750e1dbcd5a5c1808b7792fc2e6733768fb447 |

Para actualizar: bajar la version nueva, reemplazar el archivo y corregir esta tabla.

## JavaScript propio

| Archivo | Para qué |
|---|---|
| imprimir.js | Muestra el botón "Imprimir o guardar como PDF" del reporte y llama a `window.print()` (US-034). Sin JavaScript, el botón queda oculto |
