# Retrospectiva — Sprint 5

| | |
|---|---|
| **Fecha** | 2026-10-31 |
| **Periodo** | 2026-10-27 al 2026-11-02 |
| **Participantes** | Tomas, Conforti, Vergara |
| **Facilita** | Tomas (Agile Enabler) |

> **El Sprint se cierra hoy, antes de tiempo, por decision del equipo:** la presentacion final es el lunes 2/11 y todo lo demas esta
> hecho. El plan decia que los sprints cierran los lunes a la noche; este cierra el sabado 31/10. Lo que pase en la exposicion se anota
> en una addenda al final de esta acta.

## Metricas del sprint

| Metrica | Planificado | Real |
|---|---|---|
| Story Points | 48 comprometidos (38 de funcionalidad y 10 de documentacion y presentacion) | 70 |
| Historias completadas | 9 historias y tareas de funcionalidad, mas T-008 y T-009 | US-004, US-017, US-023, US-024 (estirado), US-030, US-033 y US-034; T-019, T-020, T-021, T-008 y T-009; no previstas: T-022, T-023, T-024, T-025, T-026 y T-027 |
| Horas estimadas | 72 (48 SP a 1,5 h/SP) | — |
| Horas reales | — | sin registrar en el repositorio |
| Desviacion de esfuerzo | — | no calculable |
| Defectos detectados | — | 1 (DEF-008) |
| Defectos resueltos | — | 1 (DEF-008); no queda ninguno abierto |
| Cobertura del dominio | >= 85% | 99,8% (`internal`: 91,9%), medida el 31/10 con `scripts/coverage.sh` sobre el codigo de `main` |

**Velocidad del sprint: 70 SP.** Se compone de:

| Concepto | SP |
|---|---|
| Compromiso (48 SP): T-019 (5), US-004 (2), US-030 (5), US-033 (5), US-017 (3), T-020 (8), US-034 (2), US-023 (3), T-021 (5), T-008 (5) y T-009 (5) | 48 |
| Objetivo estirado de Conforti: US-024 | 3 |
| Trabajo no previsto: T-022 (2), T-023 (5), T-024 (8), T-025 (2), T-026 (1) y T-027 (1) | 19 |

Se cumplio el 100% del compromiso y el objetivo estirado. T-018 (despliegue en Azure, 5 SP) lleva la etiqueta `sprint:5`, pero el acta
del Sprint 4 ya la conto como trabajo adelantado, asi que no se suma de nuevo. T-009 (la presentacion, 5 SP) se da por entregada con el
PR #248; la exposicion en si es el 2/11.

El trabajo no previsto (19 SP) es el 27% de la velocidad: el equipo agrego datos de demostracion (T-023), un pulido visual del layout
y de las cuatro pantallas principales (T-024), un README completo (T-025), el cierre de huecos de trazabilidad (T-022) y la actualizacion del
manual y del informe de metricas (T-026 y T-027).

### Quien hizo que

Story Points cerrados con etiqueta `sprint:5`:

| Integrante | SP | % |
|---|---:|---:|
| Tomas | 38 | 54 % |
| Vergara | 21 | 30 % |
| Conforti | 11 | 16 % |

Tomas: T-019 (5), US-004 (2), US-030 (5), US-033 (5), T-009 (5), T-023 (5), T-024 (8), T-025 (2) y T-027 (1). Vergara: US-017 (3), T-020 (8),
US-034 (2), T-008 (5), T-022 (2) y T-026 (1). Conforti: US-023 (3), T-021 (5) y US-024 (3). En el Sprint 4 Tomas habia hecho el 68 %.

**Sprint Goal:** el proyecto se puede mostrar completo, de punta a punta, desde el navegador y desplegado en Azure: Planning Poker,
esfuerzo, defectos, dashboard y reporte imprimible, con la documentacion terminada, el recorrido de trazabilidad ensayado y datos
propios cargados.

**Se cumplio:** parcialmente. Las nueve funcionalidades de la guia se ven desde el navegador y corren desplegadas en Azure; la demo
completa se probo de punta a punta el 30/10 (Planning Poker con divergencia, nueva ronda y acuerdo; completar una historia; dashboard;
reporte y PDF). La documentacion esta terminada y el recorrido de trazabilidad de US-017 se repaso contra el repositorio (`docs/presentacion/guion.md`).
Queda afuera lo de **datos propios**: la demostracion usa un proyecto de ejemplo («Tienda online»), porque los sprints cerrados del equipo no tienen
horas reales.

## Sprint Review

Lo que se reviso del incremento, con evidencia:

- **Funcionalidades:** las pantallas de esfuerzo (T-019), Planning Poker (T-020) y defectos (T-021), el dashboard (US-030), el reporte en pantalla y su
  vista imprimible (US-033 y US-034) y la nueva ronda de Planning Poker (US-017) estan en `main`. Con ellas, los 9 requerimientos de la guia
  se ven en el navegador (mapa en el README, seccion 3).
- **Despliegue:** la aplicacion corre en Azure Container Instances (Chile Central) con la base en Neon; el CD se probo de punta a punta con la cuenta
  real, y el contenedor queda apagado por defecto. La demo completa se recorrio sobre esa instancia el 30/10 y las pantallas se midieron el 31/10.
- **Calidad:** 995 funciones de test y 355 escenarios BDD en verde, cobertura de dominio de 99,8 %; 42 especificaciones y 285 criterios en la matriz de trazabilidad.
- **Medidas del sistema** (`docs/rendimiento.md`): imagen de 8,3 MB comprimida; 106 s y 139 s de CD completo; arranque en 0,5 s con 14,8 MiB de memoria;
  entre 5 y 64 ms por pantalla en local y entre 135 ms y 2 s en Azure.
- **Quien reviso:** como en todo el proyecto, cada PR lo aprobo alguien que no es su autor, con los seis chequeos del CI en verde.

La revision formal del incremento con la catedra es la presentacion del 2/11.

## Acciones de la retro anterior

| Accion | Responsable | Estado |
|---|---|---|
| La pantalla de horas (T-019) esta en `main` el 28/10 y desde ese dia cada integrante carga sus horas a diario; el 29/10 se revisa que haya horas de los tres | Tomas implementa; todos cargan | parcialmente: T-019 llego a `main` el 28/10 (#189), pero **no hay constancia en el repositorio de horas reales cargadas por el equipo ni de la revision del 29/10**. No se puede dar por cumplida |
| Cada integrante tiene su tarea de punta a punta y el 28/10 a las 22:00 cada una tiene su PR abierto con el RED a la vista; quien no lo tiene la pasa | Todos; decide Tomas | cumplida: T-020 (#190) y T-021 (#191) estaban mergeadas el 28/10 a las 21:31 y 21:36, y US-023 (#31) el 27/10 |

> Registrar las horas reales es una accion que se arrastra desde el Sprint 1 (`docs/retros/sprint-1.md`). Segun la regla de las actas se
> descarta o se reformula: no se reformula, porque no hay otro sprint; queda anotada aca como lo que no se logro.

## Que funciono

- **Las pantallas que faltaban entraron en dos dias** (T-019, T-020 y T-021 el 28/10), estimadas como tareas propias desde el Planning, como se habia decidido.
- **El despliegue en Azure quedo verificado de punta a punta** con la cuenta real, con el contenedor apagado por defecto, la region elegida por variable y los secretos en un entorno limitado a `main`. Se probo con `scripts/probar-despliegue-azure.sh`, que ademas lo vuelve a apagar.
- **Los chequeos del PR atajaron errores reales:** el chequeo de titulo y rama rechazo PRs de specs sin `Closes #N` (#237 y #239), y se corrigieron antes de mergear.
- **La revision cruzada siguio siendo real:** Vergara senalo que las capturas de 375 px estaban recortadas (#243) y que `make up` se cortaba al arrancar sobre un volumen vacio (#250); las dos cosas se arreglaron (#246 y DEF-008).
- **El reparto mejoro:** Tomas hizo el 54 % de los SP, contra 68 % del Sprint 4; Vergara paso de 18 a 21 SP y Conforti de 7 a 11.
- **Los datos de demostracion se cargan con un comando** (`go run ./cmd/demo`) y se verifican contra su spec, lo que dejo la demo repetible.
- **Se midio el sistema** (tiempos de CI/CD, peso de la imagen y latencias) y se escribio como se mide, en vez de afirmar que es rapido.

## Que no funciono

- **Las horas reales siguieron sin registrarse**: la pantalla llego el 28/10, tarde para tener datos de los sprints cerrados. Por eso la demostracion usa un proyecto de ejemplo y no el nuestro, y la desviacion de esfuerzo del equipo no se puede calcular.
- **Hubo trabajo no previsto y tardio:** 19 SP, de los cuales T-023, T-024 y T-025 (15 SP) salieron despues del 28/10, con el congelamiento a tres dias. El rediseno (T-024) cubrio cuatro pantallas y no las demas: las tablas de sprints, esfuerzo, defectos y reporte quedaron con el estilo anterior. Es un limite de alcance, no un defecto.
- **El despliegue topo con restricciones de la cuenta compartida:** dos regiones rechazadas por politica de la suscripcion (South Central US y Brazil South) antes de encontrar una que anduviera (Chile Central). Azure no informa cuales estan permitidas.
- **Encender el contenedor tarda entre 1 y 5 minutos** (tres mediciones), y el reporte tarda 2 s en Azure contra 64 ms en local; la causa mas probable es la distancia entre el contenedor y la base, pero no se midio consulta por consulta. Tampoco hay compresion HTTP ni cabeceras de cache (`docs/rendimiento.md`).
- **`make up` se cortaba sobre un volumen vacio** si Postgres todavia arrancaba (DEF-008, #251). Se arreglo el 31/10, antes del congelamiento; no se pudo reproducir el fallo de punta a punta, solo medir la ventana (0,22 s).
- **Las Dailies y las Reviews no quedaron documentadas** mas alla de las actas, en ningun sprint.

## Que vamos a cambiar

Este es el ultimo sprint de la materia, asi que las acciones son las del cierre y no las de un sprint que no hay.

| Accion | Responsable | Para cuando |
|---|---|---|
| Encender el contenedor 15 minutos antes de la presentacion, navegar dos o tres pantallas para despertar la base, exponer y apagarlo al terminar | Tomas | lunes 02/11 |
| Anotar en una addenda de esta acta como salio la exposicion y las preguntas que no supimos contestar; borrar el contenedor y solo nuestra imagen `grupo03-grupocrv` del registro compartido | Tomas | martes 03/11 |

## Riesgos nuevos o que cambiaron

| ID | Riesgo | Cambio |
|---|---|---|
| R1 | Sobrecompromiso | Cerrado: el compromiso de 48 SP se cumplio al 100 % y la velocidad fue de 70 SP, con 19 SP no previstos |
| R5 | Distribucion despareja del trabajo entre los tres | Baja de «alto» a «medio»: Tomas 54 % de los SP del sprint (68 % en el Sprint 4); Vergara 30 % y Conforti 16 % |
| R12 | El despliegue en Azure depende de datos de la catedra y comparte cuenta con otros grupos | Cerrado: se desplego y se verifico de punta a punta; el contenedor queda apagado por defecto para cuidar los creditos |
| R13 | El requerimiento 4 (Planning Poker) no se demuestra en el navegador | Cerrado: T-020 y US-017 estan en `main` y se probaron en Azure el 30/10 |
| R14 | US-023 de Conforti no se llego a implementar | Cerrado: US-023 se mergeo el 27/10 |
| R15 | Nuevo: la demostracion depende de Azure y de una base que se suspende por inactividad | Mitigado: plan B con la aplicacion local y las capturas del anexo A; el contenedor se enciende con 15 minutos de margen |

## Addenda: la exposicion del 2/11

*A completar el 2/11 y el 3/11, despues de la presentacion.*
