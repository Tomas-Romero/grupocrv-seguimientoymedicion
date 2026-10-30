# Guion de la presentación final

Presentación del lunes 2/11/2026: **14:20** con diapositivas y demo en vivo, tres oradores. Las diapositivas están en
[`index.html`](index.html) (autocontenida, sin internet) y como respaldo en [`presentacion.pdf`](presentacion.pdf). Cada diapositiva lleva
sus **notas del orador** (tecla `N`).

## Cómo se usa la presentación

| Tecla | Hace |
|---|---|
| `→` `Espacio` `Enter` | Siguiente |
| `←` | Anterior |
| `Inicio` / `Fin` | Primera / última |
| `F` | Pantalla completa |
| `N` | Panel de notas del orador y cronómetro (**no se comparte en el proyector**: abrirlo en otra ventana) |
| `T` | Reinicia el cronómetro |
| `?` | Ayuda |

Se puede saltar a una diapositiva con `index.html#7`. Para imprimir (o guardar como PDF) basta `Ctrl+P`: una diapositiva por página.

## Reparto y tiempos

El reparto es una propuesta: se puede cambiar editando `data-orador` en cada `<section>` de `index.html`.

| # | Diapositiva | Quién | Tiempo | Acumulado |
|---|---|---|---|---|
| 1 | Portada | Tomás | 0:20 | 0:20 |
| 2 | Los 9 requerimientos | Tomás | 0:50 | 1:10 |
| 3 | Scrum con evidencia | Tomás | 1:00 | 2:10 |
| 4 | El proceso en números | Tomás | 0:50 | 3:00 |
| 5 | SDD, BDD y TDD | Juan | 1:10 | 4:10 |
| 6 | Trazabilidad US-017 | Juan | 0:40 | 4:50 |
| 7 | En vivo: de la spec al test en verde | Juan | 1:00 | 5:50 |
| 8 | Arquitectura y calidad | Angelo | 0:50 | 6:40 |
| 9 | **Demo en vivo** | Angelo → Juan → Tomás | 5:00 | 11:40 |
| 10 | CI/CD hasta Azure | Tomás | 1:00 | 12:40 |
| 11 | Decisiones honestas e IA | Angelo | 1:00 | 13:40 |
| 12 | Cierre y preguntas | Tomás | 0:40 | 14:20 |

Anexos (no se presentan; se muestran si hace falta): **A** capturas del plan B, **B** preguntas probables, **C** mapa de requerimientos.

### Si hay que llegar a 10 minutos

Sacar, en este orden, hasta llegar:

1. Diapositiva 6 (trazabilidad en cadena): la 7 ya la recorre en vivo. **−0:40**
2. Diapositivas 10 y 11 a 0:30 cada una. **−1:00**
3. Diapositiva 3 a 0:40. **−0:20**
4. Demo a 4:00 (Juan salta la nueva ronda y solo muestra la divergencia). **−1:00**
5. Diapositiva 7 sin terminal en vivo (leer la salida de la pantalla). **−1:00**

Total: **≈ 10:20**. Nunca se saca la demo.

## Demo en vivo (5:00)

Datos: proyecto **Tienda online**, cargado con `go run ./cmd/demo` (ver [`../datos-de-demostracion.md`](../datos-de-demostracion.md)). Es un
proyecto de demostración, no el nuestro: lo decimos al empezar.

### Angelo · 1:30 · Backlog y defectos (solo lectura)

1. **Resumen** del proyecto: estado general, integrantes.
2. **Backlog**: prioridades (MoSCoW), estados, Story Points. Señalar una historia completada y una pendiente.
3. **Defectos**: filtrar por sprint y por severidad.

No modificar nada: así Juan y Tomás parten de los mismos datos.

### Juan · 2:00 · Planning Poker

1. Backlog → historia **16** → *Planning Poker*.
2. Votar como **Bruno** y como **Diego** (los votos quedan ocultos). Mostrar quién votó y quién falta.
3. **Revelar** → aparece la divergencia («de 3 a 13»).
4. **Abrir nueva ronda**.
5. Votar **5** y **5** (no hace falta que voten todos) → **Revelar**.
6. **Registrar acuerdo** con 5 → vuelve al backlog con la estimación guardada.

### Tomás · 1:30 · Sprint, dashboard y reporte

1. **Sprints** → Sprint 4 → completar una historia: el avance se mueve solo. (Se puede deshacer con *Reabrir*.)
2. **Esfuerzo**: horas estimadas contra reales.
3. **Dashboard**: velocidad, defectos, gráficos.
4. **Reporte** → **Imprimir o guardar como PDF** (se abre el diálogo del navegador; cancelarlo).

### Trazabilidad en vivo (Juan, diapositiva 7 · 1:00)

Preparar tres pestañas/ventanas antes de empezar.

1. `specs/US-017-nueva-ronda.md`, sección 8, criterio **CA-017-1**.
2. `features/US-017-nueva-ronda.feature`, escenario con `@CA-017-1`.
3. Terminal: `git log --oneline --reverse --grep='\[US-017\]'` (se ve RED antes de GREEN).
4. `internal/domain/planningpoker/nueva_ronda.go`.
5. Terminal: `go test ./features/ -godog.tags=@CA-017-1 -v` (≈ 3 s; **correrlo una vez antes** para que esté compilado).
6. Pull request **#209** en GitHub: los chequeos en verde.

La diapositiva ya trae la salida real de ese comando: si algo falla, se lee de ahí y se sigue.

## Plan B

| Si falla… | Hacer |
|---|---|
| Azure no responde | Abrir la app local (`docker compose --profile full up` + `go run ./cmd/demo`) en `http://localhost:8080` |
| La app local tampoco | Anexo A (capturas reales) y decir lo que se vería |
| Se cae el proyector o el archivo | `presentacion.pdf` (mismo contenido, 15 páginas) |
| Sin internet | La presentación es autocontenida; la demo corre en local |
| Falla GitHub (PR #209, tablero) | Mostrar el historial con `git log` en la terminal |

## Antes de presentar (checklist)

**Viernes/sábado**
- [ ] Congelar el código el sábado 31/10 a las 18:00. Después, solo documentación.
- [ ] Volver a medir las cifras de la diapositiva si se mergeó algo (commits, escenarios, specs, cobertura) y ajustar `index.html`.
- [ ] Tablero público para que se vea sin iniciar sesión: `gh project edit 3 --owner Tomas-Romero --visibility PUBLIC` y confirmarlo en una ventana privada.
- [ ] Neon con la contraseña rotada y el secreto `DATABASE_URL` del entorno `azure` actualizado.
- [ ] Variables del entorno `azure`: `CD_ACTIVO=false`, `CARGAR_DATOS_EJEMPLO=false`, `ACI_REGION=chilecentral`.

**Domingo (ensayo)**
- [ ] Ensayo completo con cronómetro (`N`): anotar el tiempo de cada uno y recortar lo que sobre.
- [ ] Ensayo de la demo en la computadora con la que se presenta (resolución, zoom del navegador al 125 %, sin notificaciones).
- [ ] Probar `Ctrl+P` desde el reporte.

**Lunes, 15 minutos antes**
- [ ] Encender el contenedor: `gh workflow run aci-energia.yml -f accion=encender`.
- [ ] Abrir el enlace **una vez** desde la computadora de la presentación (la IP cambia en cada reinicio y el DNS local puede tardar; si no carga, ventana privada o `ipconfig /flushdns`).
- [ ] Probar la demo completa **y volver a cargar los datos** si se cambió algo (`go run ./cmd/demo --borrar` y de nuevo `go run ./cmd/demo`, con la base de Neon).
- [ ] **No mostrar la URL en el proyector**: la aplicación no tiene login. Ocultar la barra de direcciones (`F11`) o tapar la pestaña.
- [ ] Preparar las pestañas de la trazabilidad en vivo y la app local de respaldo.

**Después**
- [ ] Apagar: `gh workflow run aci-energia.yml -f accion=apagar`.
- [ ] Borrar el contenedor y **solo** nuestra imagen (`grupo03-grupocrv`) del registro compartido, no las de otros grupos.

## Preguntas probables

Las respuestas cortas también están en el anexo B de la presentación. Todas dicen lo que hay, sin inflarlo.

| Pregunta | Respuesta |
|---|---|
| ¿Por qué un proyecto ficticio en la demo? | Los sprints cerrados del equipo no tienen horas reales (la pantalla de esfuerzo llegó el 28/10). «Tienda online» muestra todas las funciones con datos completos; nuestro proceso real está en el tablero y en las actas. |
| ¿Y las horas reales del equipo? | Las anotamos en cuatro retros seguidas. La pantalla llegó el 28/10 y la acción de la retro es que desde ahí cada integrante cargue las suyas. |
| ¿Por qué el PDF es del navegador? | Decisión documentada en US-034: mismo contenido que el reporte, con hoja de estilos de impresión. Hay PDFs de ejemplo en `docs/evidencias/US-034`. Tener a mano la confirmación de la cátedra si la hubo. |
| ¿Casi no hay commits de refactor? | Hay 10 explícitos; el RED→GREEN está en cada historia. Ejemplo: el refactor de US-017. |
| ¿Funciona en Linux y Mac? | Linux: el CI corre en Ubuntu en cada PR y la imagen corre en Azure. Mac: no está en el CI; el camino con Docker es idéntico. Dos scripts de gestión necesitan bash 4. |
| ¿Cómo aseguran que entienden el código de IA? | Se declara en cada PR, lo revisa otro integrante y cada uno puede explicar cualquier parte (`docs/uso-de-ia.md`). |
| ¿Por qué el Sprint 1 completó solo el 15 %? | Comprometimos 33 SP y cerramos 5. El sprint se fue en lo que habilita a las historias (especificaciones, infraestructura, defectos de puesta en marcha) y ninguna historia de negocio llegó a `main`. Está escrito en `docs/retros/sprint-1.md` y de ahí salió el ajuste de alcance de los sprints siguientes. |
| ¿Por qué no hay login? | La guía no lo pide. Por eso el despliegue se enciende solo para la demostración. |
| ¿Por qué el reparto de trabajo es desparejo? | Es el riesgo R5 de las retros y lo medimos con `git shortlog`. En el Sprint 4, Tomás hizo 52 de los 77 SP (68 %); mejoró respecto del Sprint 3 (Vergara pasó de 5 a 18 SP y Angelo de 3 a 7) pero sigue desparejo. Para el Sprint 5 se pasaron dos tareas de Angelo a los otros dos (`docs/retros/sprint-4.md`). |
| ¿Qué quedó fuera? | US-031 y US-032 (burndown e históricos): la guía no los exige y el dashboard ya trae gráficos de métricas. |

## Ensayo: hoja de tiempos

| Intento | Fecha | Tiempo total | Demo | Qué se recorta |
|---|---|---|---|---|
| 1 | Sáb 31/10 | | | |
| 2 | Dom 1/11 | | | |
| 3 | Dom 1/11 | | | |

Objetivo: terminar entre **13:00 y 14:30**. Si pasa de 15:00, aplicar la lista de recortes.
