# Retrospectiva — Sprint 0

| | |
|---|---|
| **Fecha** | 2026-09-24 |
| **Periodo** | 2026-09-14 al 2026-09-20 (planificado). Cierre real: 2026-09-24 |
| **Participantes** | Tomas, Conforti, Vergara |
| **Facilita** | Tomas (Agile Enabler) |

## Metricas del sprint

| Metrica | Planificado | Real |
|---|---|---|
| Story Points | 16 | 16 |
| Historias completadas | 3 (T-001, T-002, T-003) | 3 |
| Horas estimadas | 24 (16 SP a 1,5 h/SP) | sin registrar |
| Horas reales | — | sin registrar |
| Desviacion de esfuerzo | — | no calculable |
| Defectos detectados | — | 2 (ver abajo), no registrados como DEF |
| Defectos resueltos | — | 2 |
| Cobertura del dominio | >= 85% | sin codigo de dominio todavia; el CI mide el paquete `config` |

Defectos que aparecieron al primer push y se corrigieron el mismo dia: el linter
`misspell` marcaba como error las palabras en espanol, y `scripts/coverage.sh` se
subio sin permiso de ejecucion desde Windows.

**Sprint Goal:** repositorio, CI, tablero y un walking skeleton que ya toca Go,
Postgres, templ y un test verde.
**Se cumplio:** parcialmente. Quedaron hechos el repositorio con su historial de
commits por tarea, el CI en verde (lint, tests con cobertura, BDD y trazabilidad),
las migraciones con datos de ejemplo, la proteccion de `main` y el tablero con los
43 items del backlog cargados (171 SP). No se cumplio en fecha: el sprint se cerro
el 24/09 en lugar del 20/09, y todavia falta que Conforti y Vergara verifiquen el
setup en una maquina limpia. El spike de templ + HTMX y el layout con Tailwind no
se hicieron en este sprint y pasan al backlog (T-004).

## Acciones de la retro anterior

| Accion | Responsable | Estado |
|---|---|---|
| Ninguna: es el primer sprint | — | — |

## Que funciono

- Un commit por tarea, con el ID del backlog en cada mensaje: el historial ya se
  lee como proceso.
- Automatizar la carga del backlog: 43 issues con etiquetas y campos del tablero
  salieron de un solo CSV, y el total de 171 SP coincide con el plan.
- El CI encontro dos problemas en el primer push, antes de que otra persona
  trabajara encima.
- La proteccion de `main` quedo activa apenas el CI dio verde.

## Que no funciono

- El repositorio se creo antes que el plan y quedo con `Carpeta-Principal` como
  rama por defecto, mientras el CI, los scripts y la convencion asumen `main`.
  Hubo que renombrarla.
- El nombre del modulo de Go no coincidia con los imports: el primer CI habria
  fallado por compilacion. Se detecto antes del push.
- Los scripts de carga de issues fallaron tres veces por causas de Windows y de
  la API de Projects (finales de linea CRLF en el CSV, iteraciones vencidas que
  GitHub mueve a otra lista, y una condicion de carrera con el auto-add del
  tablero).
- No se registraron horas por tarea, asi que no hay base real para calibrar la
  equivalencia de 1 SP = 1,5 h.
- Las fechas del plan no se respetaron y el Sprint 1 arranco sin haber cerrado el
  Sprint 0.

## Que vamos a cambiar

| Accion | Responsable | Para cuando |
|---|---|---|
| Correr `make check` antes de cada push, y anotar en un comentario del issue las horas reales al cerrarlo | Todos; controla Tomas en la revision del PR | 2026-09-27 (Review del Sprint 1) |
| Clonar el repo en una maquina limpia y correr `make up && make check`; reportar cualquier falla como DEF | Conforti y Vergara | 2026-09-26 |

## Riesgos nuevos o que cambiaron

| ID | Riesgo | Cambio |
|---|---|---|
| R1 | Sobrecompromiso: 171 SP en 5 semanas | Sube: el Sprint 0 tardo mas de lo previsto. Se confirma en el Planning del Sprint 1 con la primera velocidad medida |
| R3 | El evaluador no puede levantar el proyecto | Sin cambio: sigue pendiente la prueba en maquina limpia |
| R7 | Parciales en la semana del Sprint 3 | Sin cambio: falta que los tres confirmen el calendario academico |
| R8 | Diferencias entre sistemas (finales de linea, permisos de ejecucion) al trabajar en Windows | Nuevo: causo dos defectos del CI y tres fallas de scripts. Se mitiga con `make check` local antes de pushear |
