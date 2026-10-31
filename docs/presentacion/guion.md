# Guion de la presentación final

Presentación del lunes 2/11/2026: **14:45** con diapositivas y demo en vivo, tres oradores. Las diapositivas están en
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
| 10 | CI/CD hasta Azure | Tomás | 0:45 | 12:25 |
| 11 | **Rendimiento: tiempos, peso y velocidad** | Tomás | 0:50 | 13:15 |
| 12 | Decisiones honestas e IA | Angelo | 0:50 | 14:05 |
| 13 | Cierre y preguntas | Tomás | 0:40 | 14:45 |

Anexos (no se presentan; se muestran si hace falta): **A** capturas del plan B, **B** preguntas probables, **C** mapa de requerimientos, **D** tiempos del CI/CD, del arranque y de los tests.

### Si hay que llegar a 10 minutos

Sacar, en este orden, hasta llegar:

1. Diapositiva 6 (trazabilidad en cadena): la 7 ya la recorre en vivo. **−0:40**
2. Diapositivas 10, 11 y 12 a 0:30 cada una. **−0:55**
3. Diapositiva 3 a 0:40. **−0:20**
4. Demo a 4:00 (Juan salta la nueva ronda y solo muestra la divergencia). **−1:00**
5. Diapositiva 7 sin terminal en vivo (leer la salida de la pantalla). **−1:00**

Total: **≈ 10:50**. Nunca se saca la demo.

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
5. Terminal: `go test ./features/ -godog.tags=@CA-017-1 -v` (el test tarda ~0,5 s; la primera vez compilar lleva ~20 s: **correrlo una vez antes**).
6. Pull request **#209** en GitHub: los chequeos en verde.

La diapositiva ya trae la salida real de ese comando: si algo falla, se lee de ahí y se sigue.

### Resultado del repaso de la cadena (30/10, sobre `main`)

Recorrí la cadena de la diapositiva 6 contra el repositorio y los números coinciden. Es un repaso de datos; falta decirlo en voz alta (ensayo del domingo).

| Eslabón | Comprobado |
|---|---|
| Spec | `specs/US-017-nueva-ronda.md`: 5 criterios (CA-017-1 a CA-017-5), mergeada en el #202 |
| Escenarios | `features/US-017-nueva-ronda.feature`: 10 escenarios |
| Commits | `git log --grep='\[US-017\]'`: 5 RED, 6 GREEN y 1 refactor; el primer RED (`665ef62`) va antes de todo GREEN |
| Pull request | #209 de Juan, aprobado por Tomás, mergeado el 28/10; chequeos en verde |
| Test en vivo | `go test ./features/ -godog.tags=@CA-017-1 -v`: PASS, 0,46 s (21 s la primera vez, compilando) |
| CI | 6 chequeos en verde: título y rama, Lint, Tests y cobertura, BDD, Trazabilidad, Perfil full |

Dos correcciones que salieron de este repaso: la salida de `go test -v` real lleva el prefijo `TestEscenarios/` (la diapositiva 7 ya lo muestra
así) y los chequeos son seis, no cinco.

## Plan B

| Si falla… | Hacer |
|---|---|
| Azure no responde | Abrir la app local (`docker compose --profile full up` + `go run ./cmd/demo`) en `http://localhost:8080` |
| La app local tampoco | Anexo A (capturas reales) y decir lo que se vería |
| Se cae el proyector o el archivo | `presentacion.pdf` (mismo contenido, 17 páginas) |
| Sin internet | La presentación es autocontenida; la demo corre en local |
| Falla GitHub (PR #209, tablero) | Mostrar el historial con `git log` en la terminal |

## El lunes, en cuatro pasos

1. **15 a 20 minutos antes:** encender el contenedor (`gh workflow run aci-energia.yml -f accion=encender`, o desde GitHub: *Actions → «ACI: encender o apagar»*). Tarda entre 1 y 5 minutos.
2. **Cuando contesta:** abrir la aplicación y navegar dos o tres pantallas (Proyectos, Backlog, Dashboard) para despertar la base. Los datos ya están cargados: no hay que cargar nada.
3. **Exponer.**
4. **Al terminar:** apagar el contenedor (`gh workflow run aci-energia.yml -f accion=apagar`).

Todo lo demás de esta sección es el detalle de cada paso y qué hacer si algo falla.

## Comandos

Todos se corren desde la raíz del repositorio, en Git Bash (Windows) o en una terminal (Linux/macOS). El prefijo `gh` necesita la sesión
iniciada con `gh auth status`. Encender y apagar el contenedor también se puede hacer **desde GitHub**: *Actions → «ACI: encender o apagar» →
Run workflow* y elegir `encender`, `apagar` o `estado`.

### Antes (lunes, 15 minutos antes)

```bash
git switch main && git pull                          # el repositorio al día
gh workflow run aci-energia.yml -f accion=encender   # enciende el contenedor (entre 1 y 3 minutos)
gh workflow run aci-energia.yml -f accion=estado     # opcional: pregunta si está encendido
curl -s -o /dev/null -w '%{http_code}' http://grupo03-grupocrv-seguimientoymedicion.chilecentral.azurecontainer.io:8080/health; echo   # tiene que decir 200
make gen && go test ./features/ -godog.tags=@CA-017-1   # deja compilado el test de la trazabilidad en vivo
```

Si la demo se corrió antes (ensayo) hay que volver a dejar los datos como estaban: `bash cargar-neon.sh --borrar` y después `bash cargar-neon.sh`
(pide la cadena de conexión de Neon sin mostrarla). Está en la carpeta de trabajo de Tomás, no en el repositorio.

### Durante

No hace falta ningún comando para la demo: se hace desde el navegador. Para la trazabilidad en vivo (diapositiva 7):

```bash
git log --oneline --reverse --grep='\[US-017\]'
go test ./features/ -godog.tags=@CA-017-1 -v
```

### Después (el mismo lunes)

```bash
gh workflow run aci-energia.yml -f accion=apagar     # apaga el contenedor (la cuenta es compartida)
```

Cuando la cátedra confirme que ya no hace falta la demostración: borrar el contenedor y **solo** nuestra imagen `grupo03-grupocrv` del registro.
Además, el contenedor se apaga solo de noche.

## Si el profesor pide probar algo en vivo

Se responde con el comando o la pantalla exacta. Ninguno de estos modifica el repositorio.

| Piden | Se muestra |
|---|---|
| Que corran **todos los tests** | `make test` (unitarios, con `-race`) |
| Los **escenarios BDD** | `make bdd`; los de una historia: `go test ./features/ -godog.tags=@US-007` |
| La **cobertura** | `make cover` (verifica los umbrales) y `docs/informe-metricas-cobertura.md` |
| **Todo lo que corre el CI** | `make check` (lint + cobertura + BDD) |
| La **matriz de trazabilidad** | `make trace` y abrir `docs/trazabilidad.md` |
| Que **levanten el proyecto desde cero** | `docker compose --profile full up --build` y abrir http://localhost:8080 (la primera vez compila: unos minutos) |
| Cargar los **datos de demostración** | `go run ./cmd/demo` (y `go run ./cmd/demo --borrar` para sacarlos) |
| El **CI en GitHub** | pestaña *Actions* o los chequeos de un PR (por ejemplo el #209) |
| El **despliegue** | *Actions → «CD»* y, en Azure, el contenedor; la app en la URL de la demo |
| Que **creen algo nuevo** en la app | *Proyectos → Nuevo proyecto*: crear, agregar integrantes, cargar una historia, crear un sprint |
| El **tablero** | GitHub Projects (el #3 del repositorio, público) |
| Un **PR con sus revisiones** | cualquiera de los últimos: título, `Closes #N`, sección «Uso de IA», aprobación de otro integrante |

Si el comando falla por una herramienta que falta (`make`, `go`, `psql`), decirlo y pasar al plan B: es un problema de la computadora, no del
proyecto, y el CI corre lo mismo en Ubuntu en cada PR.

## Prueba completa en Azure (30/10, 21:03)

Recorrida completa de la demo contra la aplicación desplegada, con los mismos pasos que el guion:

| Paso | Resultado |
|---|---|
| Encender el contenedor (`aci-energia.yml`) | `/health` en 200 a **~1 minuto** de lanzar el flujo; otra vez, el 31/10, tardó **~3 minutos** |
| Las 7 pantallas del proyecto | 200 en todas; entre 0,3 y 1,9 s (el Reporte es el más lento, ~2 s; Esfuerzo ~1,2 s) |
| Planning Poker, historia 16 | votos de Bruno (3) y Diego (13) → revelar → «Hay divergencia: de 3 a 13» → nueva ronda → 5 y 5 → «Sin divergencia», «No votaron: Carla Ruiz, Diego Sosa» → acuerdo 5 → sesión «finalizada» |
| Sprint 4, completar una historia | «Ver el historial de compras»: 5 → 10 Story Points completados |
| Dashboard | 34 → 39 de 46 → 51 (la 16 ya está estimada); sprint 4: 10 de 12 |
| Reporte a PDF (Chrome, A4) | 5 páginas, con el mismo contenido |

**La base de Neon se suspende cuando está inactiva.** El 31/10, justo después de recargar los datos, `/health` contestó 503 una vez y a los pocos segundos
200. Lo más probable (no está comprobado) es que Neon estaba despertando. Por eso, antes de presentar, **navegar dos o tres pantallas** (Proyectos, Backlog,
Dashboard) y no solo abrir el enlace.

Lo que hay que tener en cuenta al presentar: **cada acción tarda 1–2 segundos** en refrescar la pantalla (la base está en la nube); hay que esperar
y hablar mientras tanto, no volver a hacer clic. La demo **modifica los datos**: después de cada ensayo hay que recargarlos (ver «Antes»).

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
| ¿Cuánto tarda en desplegarse un cambio? | La corrida completa de CD (imagen, despliegue en Azure y apagado) tardó 106 s y 139 s. Si el código Go no cambió, publicar la imagen tarda 7 s por el caché de capas. El CI de un PR tarda unos 3 min 26 s (los chequeos corren en paralelo). Detalle en `docs/rendimiento.md` y en el anexo D. |
| ¿Cuánto pesa la imagen? | 8,3 MB comprimida (21 MB descomprimida): un binario de Go de 13,8 MB sobre Alpine, sin compiladores y sin privilegios. |
| ¿Es rápida la aplicación? | Local, con la base en la misma máquina, cada pantalla contesta en 5 a 64 ms (mediana) y aguanta unos 120 pedidos por segundo en las pantallas comunes. En Azure, 135 ms a 2 s. Arranca en 0,5 s y usa unos 15 MiB de memoria. |
| ¿Por qué en Azure es mucho más lento que en local? | Llegar al contenedor ya cuesta 116 ms. El resto es del servidor, y lo más probable es que sea la distancia entre el contenedor (Chile Central) y la base (Neon, en São Paulo). No lo medimos consulta por consulta, así que es una hipótesis. |
| ¿Qué optimizaron y qué no? | Optimizado: imagen en dos etapas, binario sin símbolos, usuario sin privilegios, caché de capas, CSS minificado, casi nada de JavaScript propio, plantillas compiladas. No optimizado: no hay compresión HTTP (el Backlog pasaría de 42,6 KB a 3,8 KB) ni cabeceras de caché en los estáticos. |
| ¿Cuánto tarda en encenderse el contenedor? | Entre 1 y 5 minutos según el día (tres mediciones: 1, 3 y 5). Por eso se enciende 15 minutos antes y se navega un par de pantallas para despertar la base. |
| ¿Qué quedó fuera? | US-031 y US-032 (burndown e históricos): la guía no los exige y el dashboard ya trae gráficos de métricas. |

## Ensayos

El equipo ensayó la exposición y confirma que entra en el tiempo; **no se anotaron tiempos por diapositiva ni del total**, así que no hay una tabla de
ensayos. El plan suma **14:45**: dentro de los 10 a 15 minutos que pide la cátedra, y con la lista de recortes de arriba si hace falta acortar.
