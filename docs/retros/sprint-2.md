# Retrospectiva — Sprint 2

| | |
|---|---|
| **Fecha** | 2026-10-12 |
| **Periodo** | 2026-10-06 al 2026-10-12 |
| **Participantes** | Tomas, Conforti, Vergara |
| **Facilita** | Tomas (Agile Enabler) |

## Metricas del sprint

| Metrica | Planificado | Real |
|---|---|---|
| Story Points | 16 | 16 |
| Historias completadas | 5 | 5 (US-001, US-002, US-005, US-025 y US-026) |
| Horas estimadas | 24 (16 SP a 1,5 h/SP) | sin registrar |
| Horas reales | — | sin registrar |
| Desviacion de esfuerzo | — | no calculable |
| Defectos detectados | — | 0 nuevos; siguen abiertos DEF-004 y DEF-005 |
| Defectos resueltos | — | 0 |
| Cobertura del dominio | >= 85% | 100% en `backlog`, `metricas` y `proyecto` |

**Velocidad del sprint: 16 SP de 16 (100%)**, contra 5 SP del Sprint 1.

**Sprint Goal:** las cinco historias con spec mergeada quedan implementadas con
TDD y BDD.

**Se cumplio:** si. Las cinco historias llegaron a `main` con su spec, sus
escenarios `.feature` y los commits RED antes de cada GREEN, por capa (dominio,
aplicacion, memoria y Postgres). El CI de `main` esta en verde. Tres de las cinco
(US-001, US-025 y US-026) entraron con *Squash and merge*, asi que en `main` cada
una es un solo commit y el ciclo RED → GREEN solo se ve en la pagina de su PR
(#96, #82 y #92). Ver "Que no funciono".

### Trabajo que no estaba en el plan

- **T-012 (#91):** `Migrar` y `make migrate` aceptan migraciones con numero menor
  al ultimo aplicado. Salio de la revision de #83: la 00002 de US-001 iba a
  llegar despues de la 00003 de US-005 y goose no habria migrado.
- **Specs del Sprint 3, primera tanda, mergeadas antes del jueves 08/10:**
  US-003 (#89), US-008 (#75), US-013 (#87) y US-019 (#88). US-009 (#95) entro la
  noche del 11/10.
- **Specs del Sprint 3, segunda tanda, abiertas el 12/10 y en revision:** US-006
  (#99), US-010 (#102) y US-011 (#106), mas una correccion de US-009 (#104) para
  que la relacion entre Sprint e item sea historica.
- **US-013 (#93)** quedo implementada y aprobada; se mergea con el mensaje
  general de `ErrItemCompletado` que pidio la revision. Cuenta en el Sprint 3.
- 15 PRs mergeados desde el 06/10 (Tomas 7, Vergara 4, Conforti 4).

## Acciones de la retro anterior

| Accion | Responsable | Estado |
|---|---|---|
| Cada historia empieza con su spec mergeada: specs del sprint siguiente como PR a mas tardar el jueves 08/10, revisadas en menos de 24 horas | Cada duenio de historia; quien tiene PRs pendientes los atiende primero | parcialmente: la primera tanda entro el 08/10; la segunda tanda y US-009 entraron el ultimo dia. #93 espero cuatro dias su revision |
| Daily asincrono antes de las 22:00 y horas reales en un comentario del issue al cerrarlo | Todos; controla Tomas en cada Refinement | no cumplida en las horas: ningun issue cerrado del sprint tiene horas anotadas. Dos retros seguidas sin cumplirse: se reformula (ver abajo) |

> Si una accion aparece dos retros seguidas sin cumplirse, se descarta o se
> reformula. La de las horas se reformula: el registro de esfuerzo real es
> US-019, que se implementa en el Sprint 3, y desde entonces las horas se
> cargan en el propio producto y no en un comentario.

## Que funciono

- **Spec primero:** las cinco historias tenian su spec mergeada antes de empezar
  el codigo y no hubo retrabajo por reglas ambiguas. La velocidad paso de 5 a 16
  SP con el mismo equipo.
- **La revision encontro problemas antes de que existieran en el codigo:** la
  relacion historica entre Sprint e item (sin ella se perdian los Story Points
  planificados de los sprints cerrados), el candado `FOR UPDATE` frente a
  `FOR NO KEY UPDATE`, la ventana entre leer y guardar una estimacion, y el
  riesgo de migraciones fuera de orden (T-012).
- **Convenciones compartidas:** errores unidos con `errors.Join`, posiciones
  desde 1, ID generado por la base y dominio en espanol sin acentos
  (ADR 0003) se repitieron en las cinco historias sin renegociarlas.
- **Cobertura:** 100% en los tres paquetes de dominio, con tests de integracion
  contra Postgres real para los repositorios.

## Que no funciono

- **Se perdio el ciclo RED → GREEN en `main` para tres historias.** El boton de
  *Squash and merge* estaba habilitado y se uso en #82, #92 y #96, aunque cada
  PR decia "Rebase and merge". La evidencia sigue en los PRs, pero no en el
  historial de `main`, que es donde la busca la catedra.
- **La entrega se concentro en el fin de semana.** US-001 y US-026 se
  mergearon la noche del 10/10, y US-002 y la spec de US-009 la del 11/10. Esa
  spec de US-009 estuvo abierta tres dias con el archivo vacio (0 bytes): el CI
  solo avisa si falta el `.feature`, no si la spec esta vacia.
- **La revision no cumplio las 24 horas.** #93 espero cuatro dias y las tres
  specs de la segunda tanda se abrieron el ultimo dia, sin margen para
  corregirlas antes del cierre.
- **Las horas reales siguen sin registrarse**, asi que no hay base para
  calibrar 1 SP = 1,5 h ni para calcular la desviacion de esfuerzo.
- **Reparto en el historial:** desde el 06/10, `git shortlog -sn` en `main`
  muestra Vergara 42, Conforti 19 (en dos identidades de Git, `ConfortiAngelo`
  y `Angelo`) y Tomas 7. El squash reduce los commits de Tomas y la doble
  identidad parte los de Conforti: la cifra no representa el trabajo real.

## Que vamos a cambiar

Maximo dos acciones. Concretas, con responsable y con fecha.

| Accion | Responsable | Para cuando |
|---|---|---|
| Desactivar *Squash and merge* en la configuracion del repo y mergear siempre con *Rebase and merge*; revisar con `git log --grep RED` que cada historia conserve su ciclo en `main` | Tomas | 2026-10-13 |
| A mitad de sprint (jueves) cada historia tiene su PR en borrador con el RED a la vista, y todo PR se revisa en menos de 24 horas; el que no pueda lo avisa en el grupo el mismo dia. Tomas controla la columna *In Review* del tablero en cada Refinement | Todos; controla Tomas | 2026-10-15 |

## Riesgos nuevos o que cambiaron

| ID | Riesgo | Cambio |
|---|---|---|
| R1 | Sobrecompromiso: el Sprint 3 tiene un techo de 36 SP y la velocidad medida es de 16 | Baja de critico a alto: el Sprint 2 cumplio el 100% y no se activa el segundo recorte. El Planning del 13/10 compromete segun lo medido, no segun el techo |
| R5 | Distribucion despareja del trabajo entre los tres | Sigue alto: Conforti entrego US-001 y US-002 con ciclo completo, pero recien el 10 y 11/10. Se unifica la identidad de Git de Conforti para que `git shortlog` sea confiable |
| R9 | Inconsistencias entre specs escritas en paralelo | Baja: la revision cruzada encontro las contradicciones (mensaje de `ErrItemCompletado`, relacion historica) antes de implementar |
| R10 | Dependencias entre las specs del Sprint 3: US-009, US-010 y US-011 comparten la relacion Sprint-item y `ErrItemCompletado` | Nuevo. Se implementan en orden US-008 → US-009 → US-010 → US-011, y #100 (condicionar la estimacion al estado) entra antes que US-010 |

## Decisiones para el Planning del Sprint 3 (13/10)

- El compromiso se fija con la velocidad medida (16 SP) y no con el techo de 36.
  Lo que sobre queda en *Sprint Backlog* ordenado, y se toma si hay capacidad.
- Orden sugerido por dependencias: US-008, US-013 (ya implementada), T-005 y
  US-019 primero; US-009, US-006 y US-010 despues; US-011 al final, porque cierra
  el ciclo y necesita a las demas.
- Las specs de US-006, US-010, US-011 y la correccion de US-009 (#99, #102, #106
  y #104) tienen que estar mergeadas antes de empezar su implementacion.
