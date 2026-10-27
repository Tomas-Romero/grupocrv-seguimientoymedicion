# Retrospectiva — Sprint 4

| | |
|---|---|
| **Fecha** | 2026-10-26 |
| **Periodo** | 2026-10-20 al 2026-10-26 |
| **Participantes** | Tomas, Conforti, Vergara |
| **Facilita** | Tomas (Agile Enabler) |

## Metricas del sprint

| Metrica | Planificado | Real |
|---|---|---|
| Story Points | 35 comprometidos (techo 46) | 77 |
| Historias completadas | 8 comprometidas y estiradas | 12 historias y 7 tareas tecnicas: US-011, US-012, US-014, US-015, US-016, US-018, US-020, US-021, US-022, US-027, US-028 y US-029; T-007, T-011, T-014, T-015, T-016, T-017 y T-018 |
| Horas estimadas | 52,5 (35 SP a 1,5 h/SP) | — |
| Horas reales | — | sin registrar |
| Desviacion de esfuerzo | — | no calculable |
| Defectos detectados | — | 1 (DEF-007) |
| Defectos resueltos | — | 1 (DEF-007); no queda ninguno abierto |
| Cobertura del dominio | >= 85% | 99,8% (`internal`: 93,9%), medida con `scripts/coverage.sh` sobre el codigo de `main` |

**Velocidad del sprint: 77 SP.** Se compone de:

| Concepto | SP |
|---|---|
| Compromiso (35 SP): US-011, US-014, US-015, T-011, T-015, US-027 y US-028 | 35 |
| Objetivo estirado (11 SP): US-016, US-018 y US-029 | 11 |
| Tarea no prevista del tablero: T-014 | 2 |
| Trabajo adelantado del Sprint 5 o fuera del tablero: US-012 (2), US-020 (3), US-021 (3), US-022 (5), T-016 (3), T-017 (3), T-018 (5) y T-007 (5) | 29 |

Se cumplio el 100% del compromiso y el 100% del objetivo estirado (46 de 46 SP). Los 48 SP
del tablero original son el numero comparable con el Sprint 3 (31); los 29 SP restantes son
trabajo que habia que hacer igual y que baja la carga del Sprint 5.

**Sprint Goal:** el ciclo proyecto → backlog → sprint → cierre se usa desde el navegador,
se pueden iniciar sesiones de Planning Poker con votos ocultos, y el producto calcula horas
estimadas, reales, desviacion y porcentaje de completitud.

**Se cumplio:** parcialmente.

- **Si, el ciclo en el navegador:** se puede crear un proyecto, registrar integrantes, cargar y
  estimar el backlog, crear un sprint, asignarle historias, marcarlas, cerrarlo y consultar los
  sprints anteriores (T-011, T-015, T-016 y US-012).
- **Si, los calculos:** horas estimadas, horas reales, desviacion, porcentaje de completitud y
  defectos por sprint (US-027, US-028 y US-029), y las consultas de esfuerzo por historia,
  integrante y sprint, y de estimado contra real (US-020 y US-021).
- **Solo en dominio y aplicacion, no en el navegador:** Planning Poker y los defectos. Se puede
  iniciar una sesion, votar con votos ocultos, revelar, detectar divergencias y registrar la
  estimacion acordada (US-014, US-015, US-016 y US-018), y registrar un defecto (US-022), todo
  con TDD y BDD, pero no hay ninguna pantalla ni ruta HTTP. La pantalla de votacion minima que se
  decidio en el Planning del 20/10 no llego a estimarse ni a asignarse. Lo mismo pasa con las
  metricas y el esfuerzo: no hay donde verlos ni cargarlos.
- **Entrega continua:** el CI/CD esta completo hasta la imagen publicada y el despliegue en la
  Container App (T-017 y T-018); falta probarlo contra Azure real cuando la catedra entregue las
  credenciales.

### Quien hizo que

| Item | SP | Responsable previsto | Lo implemento |
|---|---|---|---|
| US-011 Cerrar un sprint | 5 | Vergara | Vergara |
| US-014 Iniciar una sesion de Planning Poker | 5 | Vergara | Vergara |
| US-015 Voto individual oculto | 5 | Vergara | Vergara |
| US-018 Registrar la estimacion acordada (estirado) | 3 | Vergara | Vergara |
| T-011 Pantallas de proyectos, integrantes y backlog | 8 | Tomas | Tomas |
| T-015 Pantallas de sprints | 5 | Tomas | Tomas |
| US-027 Horas estimadas, reales y desviacion | 5 | Tomas | Tomas |
| US-016 Revelar las estimaciones (estirado) | 5 | Tomas | Tomas |
| US-029 Defectos por sprint (estirado) | 3 | Tomas | Tomas |
| US-028 Porcentaje de historias completadas | 2 | Conforti | Conforti |
| T-014 Chequeo del ID en cada commit | 2 | — | Tomas |
| US-022 Registrar un defecto (adelantado del Sprint 5) | 5 | Conforti | Conforti |
| T-018 Despliegue continuo en Azure Container Apps (fuera del tablero) | 5 | — | Tomas |
| T-007, US-012, T-016, US-020, US-021 y T-017 (adelantado o fuera del tablero) | 19 | Tomas (US-012, US-020 y US-021 estaban en el Sprint 5) | Tomas |

**SP por persona:** con todo lo entregado, Tomas 52 de 77 (68%), Vergara 18 (23%) y Conforti 7
(9%). En el tablero original (48 SP), Tomas 28 (58%), Vergara 18 (38%) y Conforti 2 (4%).
Vergara paso de 5 SP en el Sprint 3 a 18, y Conforti de 3 a 7: la implementacion de US-022 entro
el lunes por la noche, y escribio ademas las specs de US-022 y US-023.

### Trabajo que no estaba en el plan

- **DEF-007 (#155):** un sprint cerrado contaba como completada una historia que se habia
  completado despues en otro sprint. Se detecto el 23/10 y se corrigio el mismo dia con un
  resultado por historia y sprint (`HistoriaAsignada.Completada()`), con test.
- **T-014 (#146):** chequeo de CI que falla si algun commit de un PR no lleva su `[ID]`. Desde
  entonces ningun commit de `main` carece de ID.
- **T-007 (#148):** la matriz de trazabilidad (`docs/trazabilidad.md`) se genera sola con
  `go run ./cmd/trace` y une specs, criterios, tests y escenarios.
- **T-016 (#163) y US-012 (#164):** cerrar un sprint y consultar los sprints anteriores desde el
  navegador, que habian quedado sin pantalla.
- **T-017 (#173) y T-018 (#187):** entrega continua. El workflow `cd-azure.yml` construye la
  imagen Docker en cada push a `main`, la publica en el registro compartido `ingcalidad2026` y
  actualiza la Container App `grupo03-grupocrv-seguimientoymedicion` con 0,25 vCPU, 0,5 GiB y minimo de
  replicas en 0, para cuidar los creditos de la cuenta compartida; despues comprueba `/health`. Sin
  los secretos construye, avisa y no despliega. La base de datos es externa (Neon) porque la
  catedra pidio no usar un contenedor de Azure. Falta recibir el JSON de credenciales. El CI
  valida ademas todos los workflows con `actionlint`.
- **Specs del Sprint 5 por adelantado:** US-020, US-021, US-022 y US-023 estaban mergeadas antes
  de empezar su codigo.
- 37 PRs mergeados desde el 20/10 (Tomas 26, Vergara 7, Conforti 4). En `main`, 70 commits `RED` y
  76 `GREEN` desde el 20/10 conservan el ciclo, ningun commit carece de `[ID]` y el CI de `main` esta
  en verde.

## Acciones de la retro anterior

| Accion | Responsable | Estado |
|---|---|---|
| Agregar a `pr-title.yml` un chequeo que falle si algun commit del PR no lleva `[US-NNN]`, `[DEF-NNN]` o `[T-NNN]`; corregir con *reword* los PRs abiertos antes | Tomas, revisa Vergara | cumplida: #146 se mergeo el 22/10 y ningun commit de `main` desde el 20/10 carece de ID |
| Cada historia tiene su PR en borrador el jueves 22/10 a las 22:00, con el RED a la vista; quien no lo tiene pasa su historia a quien tenga capacidad. Lo mismo para una revision sin responder en 24 horas | Todos; decide Tomas | parcialmente: Vergara abrio como borrador US-011 (#149), US-014 (#152) y US-015 (#153) el jueves; Tomas tenia T-011 y T-015 ya en revision o mergeados y US-027 (#154) abierto el jueves a las 21:16, pero no como borrador, y US-016 (#162) llego el viernes 23/10 a las 18:38; Conforti abrio US-028 (#165) el sabado 24/10 y no se reasigno. Casi todos los PR se mergearon en menos de 24 horas desde que se abrieron; #162 tardo dos dias por cambios pedidos y por la aprobacion de un code owner |

> Si una accion aparece dos retros seguidas sin cumplirse, se descarta o se
> reformula. La del borrador del jueves ya se reformulo una vez y se cumplio a medias de
> nuevo (la entrega mejoro, el borrador del jueves no): se reemplaza, ver abajo.

## Que funciono

- **La entrega se reparte mejor en la semana.** De los 77 SP, 25 estaban en `main` el jueves por la
  noche y 45 el viernes (58%); el sabado, el domingo y el lunes entraron 32 SP, y 25 el lunes. En
  el Sprint 3, 29 de 31 SP (94%) llegaron en los ultimos tres dias. Todavia queda trabajo pegado al
  ultimo dia: ver abajo.
- **Vergara sostuvo su cadena de Planning Poker:** US-011, US-014, US-015 y US-018 con sus
  specs, abriendo los PR como borrador el jueves y con un test de concurrencia determinista
  (espera el candado consultando la base, sin `sleep`) en US-018.
- **Conforti entrego una historia completa de dominio:** US-022 con 12 criterios, una migracion y
  23 escenarios, y descubrio al implementarla que la lista de campos del catalogo de errores
  (RN-T005-8) es cerrada y no incluia `descripcion` ni `severidad`; la corrigio dejandolo anotado
  en las dos specs.
- **Specs primero, otra vez:** las doce historias implementadas del sprint tenian su spec
  mergeada antes del codigo, y las de US-020 a US-023 se escribieron antes de que se las necesite.
- **La revision pidio cambios reales:** en #162 (US-016) se pidio que ningun puerto entregue los
  votos en crudo (RN-016-12), y las specs de US-022 y US-023 recibieron observaciones antes de
  mergearse.
- **Los chequeos automaticos hicieron su trabajo:** T-014 (ID en cada commit), T-005 (catalogo de
  errores) y T-007 (matriz de trazabilidad) funcionan sin intervencion; el dominio tiene
  99,8% de cobertura.
- **Las pantallas existen:** T-011 y T-015 dejaron el ciclo proyecto → backlog → sprint
  utilizable desde el navegador, y T-016 y US-012 lo completaron.

## Que no funciono

- **El reparto sigue desparejo.** Tomas implemento 52 de los 77 SP del sprint (68%) y 28 de los 48
  del tablero original; Vergara 18 y Conforti 7. R5 sigue sin resolverse, aunque los otros dos ya
  sostienen una cadena propia.
- **Las horas reales siguen sin registrarse, por cuarta retro seguida.** US-019, US-020 y US-021
  ya estan en el producto, pero no hay pantalla para cargar horas, asi que las horas estimadas,
  reales y la desviacion siguen sin datos reales.
- **Planning Poker, esfuerzo, defectos y metricas no se ven en el navegador.** El requerimiento 4
  de la guia (Planning Poker) no se puede demostrar desde la interfaz, y la pantalla minima que se
  decidio en el Planning del 20/10 no se estimo ni se asigno.
- **El lunes concentro 25 SP.** US-018, US-029, T-017, T-018, US-020, US-021 y US-022 se mergearon
  el ultimo dia, algunas de ellas pasada la tarde, lo que dejo poco margen para la revision.
- **Cuello de botella en las aprobaciones.** El dominio y la configuracion tienen como code owners a
  Tomas y Vergara, y Tomas escribio la mayoria de los PR: las aprobaciones de los PR de codigo de dominio
  recayeron en Vergara. Con `main` exigiendo estar al dia y descartando las aprobaciones al actualizar, el lunes
  hubo que volver a aprobar uno por uno los PR del lote final.
- **Los mismos dos archivos chocan en casi cada PR de implementacion:** `docs/trazabilidad.md` y
  `docs/diccionario-steps.md` (por ejemplo en US-016, US-021 y US-029), y cada choque se resuelve a
  mano o regenerando la matriz.
- **T-017 llego con el Lint en rojo:** el chequeo local de los workflows (`actionlint`) no tenia
  `shellcheck` instalado, asi que no vio un hallazgo que el CI si marco (SC2016). Se corrigio
  pasando los datos de otros pasos por variables de entorno.
- **Un PR llego con una afirmacion sin comprobar:** el cuerpo de #185 decia que tres tests de Planning
  Poker fallaban en `main`; no se reproduce en el CI ni en otra maquina y era el entorno local.
- **El tablero quedo desactualizado.** US-012, US-020, US-021 y US-022 figuraban en el Sprint 5 y T-016, T-017,
  T-018 y T-007 no tenian sprint, aunque se entregaron en este; varias etiquetas `sprint:N` estaban viejas.
- **El entorno de desarrollo es inestable.** Docker Desktop se cayo varias veces y los tests de
  Postgres de una maquina cargada fallaron o no pudieron correr; el CI fue la unica verificacion
  independiente en esos momentos.

## Que vamos a cambiar

Maximo dos acciones. Concretas, con responsable y con fecha. Una accion sin
responsable no es una accion.

| Accion | Responsable | Para cuando |
|---|---|---|
| La pantalla para cargar horas (T-019, la primera del Sprint 5) esta en `main` el miercoles 28/10, y desde ese dia cada integrante carga sus horas a diario. El jueves 29/10 se revisa que haya horas de los tres | Tomas implementa la pantalla; todos cargan; Tomas revisa | 2026-10-28 (pantalla) y 2026-10-29 (revision) |
| Cada integrante tiene su tarea del Sprint 5 de punta a punta (Tomas: esfuerzo y dashboard; Vergara: Planning Poker; Conforti: defectos) y el miercoles 28/10 a las 22:00 cada una tiene su PR abierto con el RED a la vista. Quien no lo tiene pasa su tarea a quien tenga capacidad; decide Tomas en el momento. Reemplaza al borrador del jueves | Todos; decide Tomas | 2026-10-28 |

## Riesgos nuevos o que cambiaron

| ID | Riesgo | Cambio |
|---|---|---|
| R1 | Sobrecompromiso | Baja de critico a alto. El Sprint 5 compromete 48 SP (38 de funcionalidad y 10 de documentacion y presentacion) con las pantallas ya estimadas, contra 77 SP entregados en el Sprint 4 (48 del tablero original); el congelamiento sigue siendo el sabado 31/10, y las pantallas dependen de dominio que ya esta hecho |
| R5 | Distribucion despareja del trabajo entre los tres | Sigue alto: Tomas 68% de los SP del sprint (58% en el tablero original). Mejora: Vergara paso de 5 a 18 SP y Conforti de 3 a 7. Para el Sprint 5, dos tareas de Conforti pasan a los otros dos |
| R11 | La velocidad incluye trabajo de rescate y no es repetible | Sigue: 29 de los 77 SP son trabajo adelantado o no previsto; para el Planning, comprometer sobre 48 SP y no sobre 77 |
| R12 | Nuevo: el despliegue en Azure depende de datos de la catedra y comparte cuenta con otros grupos | Ya se conocen el registro, el grupo de recursos y el prefijo de la Container App; falta el JSON de credenciales, que se pasa mas tarde. La cuenta es compartida y con creditos limitados: el pipeline solo puede tocar recursos `grupo03-*` y nunca borrar el registro |
| R13 | Nuevo: el requerimiento 4 (Planning Poker) no se demuestra en el navegador | Se estima y se asigna en el Planning del 27/10 (T-020 y US-017) |
| R14 | Nuevo: US-023 de Conforti no se llego a implementar | Pasa al Sprint 5 con una regla: si el miercoles 28/10 a las 22:00 no tiene PR abierto, la toma Vergara |

## Decisiones para el Planning del Sprint 5 (27/10)

El detalle (alcance, responsables, dependencias, calendario y lista de recorte) esta en el plan de
trabajo, seccion «Compromiso del Sprint 5». Lo que se decide a partir de este sprint:

- **Las pantallas que faltan se estiman como tareas propias:** T-019 (carga y consulta de esfuerzo, 5 SP),
  T-020 (Planning Poker, 8 SP) y T-021 (defectos, 5 SP).
- **US-034 (exportar a PDF, 8 SP) se reduce a una vista imprimible del reporte en pantalla (US-033), 2 SP:**
  la catedra confirmo el 26/10 que alcanza con que el navegador genere la vista y se imprima como PDF.
- **Conforti pasa dos historias a los otros dos:** US-030 (dashboard) a Tomas y US-034 (vista imprimible) a
  Vergara, para dejarle US-023 y las pantallas de defectos.
- **Tablero:** las historias y tareas entregadas en este sprint se mueven al Sprint 4 y las etiquetas
  `sprint:N` viejas se corrigen.
- **Despliegue en Azure:** la catedra confirmo que hay que demostrar despliegue continuo con cada merge, en una
  cuenta compartida y cuidando los creditos (registro `ingcalidad2026`, grupo de recursos `frsrutn` y
  Container App con prefijo `grupo03-`). La base de datos no puede ser un contenedor de Azure: se usa Neon, y la cadena
  de conexion y el JSON de credenciales se guardan como secretos de un Environment limitado a `main`, visibles solo
  para quien los carga. La presentacion es de formato libre, con capturas de apoyo pero mostrando el proyecto andando.
