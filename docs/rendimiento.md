# Tiempos, peso y rendimiento

Todo lo que se puede **medir** de cómo se construye, se despliega y se ejecuta la aplicación, con cómo se midió y qué no se midió. No es un
benchmark riguroso: es una medición honesta, hecha el 31/10/2026 sobre `main`, para poder contestar «¿cuánto tarda?» y «¿cuánto pesa?» con números y no con
adjetivos.

**Dónde se midió.** *Local:* Docker Desktop en Windows, contenedor de la aplicación y Postgres 16 en la misma máquina, con los datos de demostración
(`go run ./cmd/demo`), pedidos desde un cliente Python de un solo proceso. *CI/CD:* GitHub Actions, en `ubuntu-latest`. *Azure:* Container Instances en Chile
Central (1 vCPU, 1 GB), base en Neon, pedidos desde una máquina en Argentina.

## 1. Integración y despliegue continuos

Tiempos reales de las corridas de GitHub Actions (`gh run view`), sin la cola de espera.

### CI en cada pull request (se corren en paralelo)

| Chequeo | Duración |
|---|---:|
| Trazabilidad | 3–5 s |
| Lint | 32–43 s |
| Perfil full (levanta base y aplicación desde cero) | 65–67 s |
| BDD (355 escenarios) | 72–80 s |
| Tests y cobertura | 203 s |
| **Total del PR** (el más lento marca el tiempo) | **≈ 3 min 26 s** |

### CD: del merge al contenedor funcionando

| Paso | Duración |
|---|---:|
| Construir la imagen, sin desplegar (corrida de un `push` a `main`): el paso, y el job completo | 36–43 s; 50–61 s |
| Construir y publicar la imagen, con las capas ya en el registro | **7 s** |
| Desplegar el contenedor en ACI y verificar `/health` | 46–72 s |
| Apagar el contenedor al terminar la prueba | 3 s |
| **Corrida completa: imagen + despliegue + apagado** | **106 s y 139 s** (dos corridas) |

La diferencia entre 36–50 s y 7 s es el caché de capas del registro: si el código Go no cambió respecto de la última imagen, no se vuelve a compilar.

### Encender y apagar para la demostración

| Qué | Tiempo |
|---|---:|
| Flujo `aci-energia.yml` (encender) | 58–60 s |
| Flujo `aci-energia.yml` (apagar) | 20–33 s |
| Desde lanzar «encender» hasta que `/health` contesta 200 | **1, 3 y 5 minutos** (tres veces; **varía mucho**) |

Por eso el contenedor se enciende **con 15 minutos de margen**. Además, la primera pantalla después de un rato sin uso puede tardar más, porque Neon
suspende la base cuando está inactiva.

## 2. Peso

### Imagen Docker

| Qué | Tamaño |
|---|---:|
| Imagen comprimida (`docker save`, lo que viaja al registro) | **8,3 MB** |
| Sistema de archivos descomprimido de la imagen | 21,0 MB |
| Binario del servidor (`/app/server`) | 13,8 MB |
| Carpeta de estáticos copiada a la imagen (`/app/web/static`, incluye el CSS fuente y un LEEME) | 115 KB |

### Archivos que baja el navegador

| Archivo | Tamaño |
|---|---:|
| CSS compilado y minificado (`app.css`) | 21,5 KB |
| HTMX (`htmx.min.js`) | 50,9 KB |
| JavaScript propio (`imprimir.js`, el botón de imprimir) | 0,5 KB |
| Fuentes e imágenes | ninguna |

### Páginas (HTML, con los datos de demostración)

| Pantalla | Sin comprimir | Con gzip (estimado) |
|---|---:|---:|
| Proyectos | 3,8 KB | 1,3 KB |
| Resumen | 11,1 KB | 2,3 KB |
| Sprints | 10,4 KB | 2,1 KB |
| Dashboard | 13,0 KB | 2,7 KB |
| Reporte | 23,9 KB | 3,7 KB |
| Esfuerzo | 24,4 KB | 3,6 KB |
| Defectos | 28,3 KB | 3,9 KB |
| Backlog | 42,6 KB | 3,8 KB |

La columna «con gzip» es lo que daría comprimir el mismo HTML con `gzip -6`: **la aplicación hoy no comprime las respuestas** (no hay compresión HTTP
en el código). Es una mejora posible, no una hecha.

## 3. Ejecución

### Arranque y memoria (local, contenedor de producción)

| Qué | Resultado |
|---|---:|
| Del `docker start` hasta que `/health` contesta 200 (5 veces) | **0,5–0,8 s** |
| Memoria después de ≈ 1.800 pedidos | **14,8 MiB** |
| CPU en reposo | 0 % |

El arranque incluye aplicar las migraciones y conectarse a la base.

### Latencia por pantalla (local, 60 pedidos seguidos cada una)

| Pantalla | p50 | p95 |
|---|---:|---:|
| Proyectos | 5 ms | 6 ms |
| Sprints | 14 ms | 17 ms |
| Resumen | 15 ms | 17 ms |
| Backlog | 16 ms | 24 ms |
| Dashboard | 24 ms | 32 ms |
| Defectos | 26 ms | 41 ms |
| Esfuerzo | 30 ms | 34 ms |
| Reporte | 64 ms | 110 ms |
| `/health` | 5 ms | 8 ms |

### Carga (local, pedidos en paralelo)

| Prueba | Pedidos por segundo | p50 | p95 |
|---|---:|---:|---:|
| Dashboard, 300 pedidos, 10 en paralelo | 117 | 75 ms | 151 ms |
| Backlog, 300 pedidos, 10 en paralelo | 125 | 62 ms | 229 ms |
| Reporte, 200 pedidos, 10 en paralelo | 52 | 181 ms | 255 ms |
| `/health`, 500 pedidos, 20 en paralelo | 232 | 72 ms | 188 ms |

Es un **piso**, no un techo: el cliente es un solo proceso de Python y el contenedor corre dentro de Docker Desktop. Sin errores en ninguna prueba.

### En Azure (Chile Central, 20 pedidos seguidos cada una, desde Argentina)

| Pantalla | p50 | p95 |
|---|---:|---:|
| Archivo estático (`app.css`, sin base de datos) | 116 ms | 139 ms |
| `/health` | 132 ms | 151 ms |
| Proyectos | 135 ms | 152 ms |
| Sprints | 379 ms | 419 ms |
| Resumen | 441 ms | 455 ms |
| Backlog | 447 ms | 483 ms |
| Defectos | 652 ms | 674 ms |
| Dashboard | 736 ms | 757 ms |
| Esfuerzo | 956 ms | 985 ms |
| Reporte | 2.033 ms | 2.057 ms |

El archivo estático, que no toca la base, tarda 116 ms: ese es el costo de llegar al contenedor. El resto del tiempo es del lado del servidor. **Lo más
probable** es que lo explique la distancia entre el contenedor (Chile Central) y la base (Neon, en São Paulo): cada consulta cruza de región, y el reporte
es la pantalla que más datos junta. **No lo medí consulta por consulta**, así que es una hipótesis. Localmente, con base y aplicación juntas, el reporte
tarda 64 ms.

### Los tests

| Qué | Duración (local, con Postgres) |
|---|---:|
| `go test ./...` completo, sin caché | ≈ 104 s |
| Escenarios BDD (355 escenarios, 2.857 pasos) | 6–8 s |
| Dominio (`internal/domain/...`) | 4–5 s |

## 4. Qué se hizo para que sea liviano y rápido

Cosas verificables en el código y en el `Dockerfile`:

- **Imagen en dos etapas:** se compila en `golang:alpine` y se copia solo el binario y los estáticos a `alpine`. Nada de compiladores en producción.
- **Binario chico y autocontenido:** `CGO_ENABLED=0`, `-trimpath` y `-ldflags="-s -w"` (sin símbolos de depuración): 13,8 MB.
- **Corre sin privilegios** (usuario `app`, no `root`).
- **Caché de capas:** `go mod download` va antes de copiar el código, y el registro guarda las capas: reconstruir sin cambios de Go tarda 7 s.
- **CSS mínimo:** Tailwind compila solo las clases que se usan, y se minifica (21,5 KB).
- **Casi sin JavaScript propio:** 0,5 KB (el botón de imprimir); lo demás es HTMX, de 50,9 KB, que se carga una vez.
- **Plantillas compiladas:** `templ` las convierte en código Go al construir; no se interpretan plantillas en cada pedido.
- **Sin fuentes web ni imágenes:** la aplicación se ve igual sin internet.

## 5. Qué no está optimizado

- **Sin compresión HTTP.** El Backlog mide 42,6 KB y con gzip serían unos 3,8 KB.
- **Sin cabeceras de caché** en los estáticos (no hay `Cache-Control` en el código).
- **El reporte es la pantalla más lenta** (64 ms local y 2 s en Azure) y es la que más trabajo hace.
- **La distancia contenedor–base en Azure** (hipótesis arriba): la cuenta compartida solo dejó desplegar en Chile Central.
- **No se probó carga sostenida ni muchos usuarios a la vez** en Azure: la cuenta es compartida y tiene créditos limitados.

## 6. Cómo volver a medir

| Qué | Cómo |
|---|---|
| Duración de cada chequeo y paso del CI/CD | `gh run view <id> --json jobs` (campos `startedAt` y `completedAt` de cada job y paso) |
| Peso comprimido de la imagen | `docker build -t app . && docker save app \| wc -c` |
| Peso descomprimido | `docker create --name x app && docker export x \| wc -c` |
| Arranque | `docker start <contenedor>` y consultar `/health` hasta que dé 200 |
| Memoria | `docker stats --no-stream <contenedor>` |
| Latencia de una pantalla | 60 pedidos seguidos a la misma URL y tomar la mediana y el percentil 95 |
| Tests | `go test -count=1 ./...` con `DATABASE_URL` apuntando a un Postgres local |
