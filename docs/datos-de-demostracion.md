# Datos de la demostración

El comando `cmd/demo` llena una base **vacía** con un proyecto de demostración completo, siempre con los mismos datos, para mostrar en
la presentación todas las funcionalidades con información que se vea real. Está especificado en
[`specs/T-023-datos-de-demostracion.md`](../specs/T-023-datos-de-demostracion.md).

No reemplaza al sembrado mínimo de `CARGAR_DATOS_EJEMPLO` (un proyecto con sus integrantes, que corre en cada arranque): se corre **una
vez, a mano**, contra la base de la demostración. El contenedor de Azure sigue con `CARGAR_DATOS_EJEMPLO=false`.

## Qué carga

Un proyecto, **«Tienda online»**, con datos ficticios (correos `…@tienda.test`) y fecha de corte el **30/10/2026**:

| Qué | Cuánto |
|---|---|
| Integrantes | 4, con los tres roles |
| Historias | 16 (5 must, 6 should, 3 could y 2 wont), con criterios de aceptación; 8 completadas y 8 pendientes |
| Sprints | 4 de 14 días: los tres primeros **cerrados** y el cuarto **abierto** (26/10 al 08/11) |
| Planning Poker | 3 sesiones: una **votando** (historia 16, con 2 votos ocultos y 2 sin votar), una **revelada con divergencia** (historia 11) y una **finalizada** con dos rondas (historia 9, estimada en 5) |
| Horas | 27 registros en las cinco actividades; desviación de +25,6 %, −10,6 % y 0 % en los sprints cerrados |
| Defectos | 10 (2 críticos, 2 altos, 3 medios y 3 bajos; 4 abiertos, 1 en corrección, 4 resueltos y 1 descartado) |

Las cifras exactas de cada sprint están en la sección 3 de la spec.

## Cargar los datos en Neon (la base de la presentación)

La cadena de conexión de Neon es un secreto: **no se escribe en el chat, ni en un archivo, ni en la línea de comandos** (quedaría en el
historial de la terminal). Se pide sin mostrarla, se usa y se borra de la sesión:

```bash
read -rsp "DATABASE_URL de Neon: " DATABASE_URL; echo; export DATABASE_URL
go run ./cmd/demo
unset DATABASE_URL
```

La cadena es la misma que usa el servidor: la **directa** (sin `-pooler`), con `sslmode=require` y **sin** `channel_binding`
(ver «La base de datos» en [`ci-cd.md`](ci-cd.md)). El comando aplica las migraciones pendientes, carga todo, **lee lo que quedó
cargado y lo compara con las cifras de la spec**, y recién entonces imprime:

```
Proyecto "Tienda online" cargado: 4 integrantes, 16 historias, 4 sprints (3 cerrados y 1 abierto), 3 sesiones de Planning Poker, 27 registros de horas y 10 defectos.
```

La primera conexión puede tardar unos segundos si Neon estaba dormida. Nunca imprime la cadena de conexión, ni siquiera en un error.

## Volver a empezar

- Si el proyecto **ya existe**, el comando se niega y no escribe nada. Para cargarlo de nuevo, primero se borra:

  ```bash
  go run ./cmd/demo --borrar
  ```

  Borra **solo** el proyecto «Tienda online» (por su nombre exacto) y todo lo suyo; no toca ningún otro proyecto de la base. Si no
  existía, lo avisa y termina bien.
- Si la carga **se corta a la mitad** (por ejemplo, se cayó la red), lo ya cargado queda y el mensaje dice en qué paso falló. Se
  borra con `--borrar` y se carga de nuevo.
- La carga no depende del día en que se corre: usa un reloj propio para que las horas y las historias completadas queden con fechas
  dentro de cada sprint.

## Plan B: la aplicación en local con los mismos datos

Con la base local (`make up`) y la misma carga:

```bash
DATABASE_URL="postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable" go run ./cmd/demo
make run
```

(La cadena de la base local no es un secreto; la de Neon sí, y va siempre por `read -rs`.)

## Qué mostrar y dónde

| Para mostrar | Dónde |
|---|---|
| Estado general del proyecto y sus integrantes | **Resumen** |
| Backlog con prioridades MoSCoW, estimaciones y criterios | **Backlog** |
| Estimar con la escala Fibonacci | **Backlog**: elegir los puntos de una historia sin estimar y «Estimar» (las historias 11, 14, 15 y 16) |
| Planning Poker en vivo | **Backlog → Planning Poker** de la historia 16 (**votando**): votar como Bruno y como Diego, revelar y, si hay divergencia, abrir una nueva ronda y registrar el acuerdo. La historia 11 ya está **revelada con divergencia**: se puede abrir una nueva ronda y acordar un valor |
| Sprint en curso, historias completadas y cierre | **Sprints → Sprint 4** (marcar una historia completada; cerrar el sprint recién al final, porque no se puede deshacer) |
| Historias que volvieron al backlog | **Sprints anteriores** (las historias 3, 6 y 8) |
| Horas, y estimado contra real | **Esfuerzo** |
| Defectos por severidad y estado | **Defectos** (y su filtro por sprint y severidad) |
| Velocidad, porcentaje de completadas y gráficos | **Dashboard** |
| Reporte completo y PDF | **Reporte** (proyecto y de cada sprint) → «Imprimir o guardar como PDF» |

**Ojo con lo que se modifica en vivo:** todo lo que se hace en la demostración se guarda en la base. Si se cierra el Sprint 4 o se
cambian estados, la base deja de coincidir con la spec. Para volver al punto de partida: `go run ./cmd/demo --borrar` y cargar de nuevo.
