# Matriz de trazabilidad

> Generada con `make trace` a partir de `specs/`, `features/` y los comentarios de los tests.
> No se edita a mano: se regenera. Enlaza Historia → Spec → Criterios de aceptación →
> Escenarios BDD → Tests → Código (la guía de la cátedra la pide como evidencia).

Total: 19 especificaciones, 124 criterios de aceptación, 75 con escenario BDD y 87 con test.

Un criterio sin escenario BDD o sin test se muestra con `—`: o falta la evidencia, o el criterio
se cubre de otra forma y la spec lo dice (por ejemplo, las tareas técnicas no llevan `.feature`).

## Resumen por historia

| Historia | Spec | Estado | Criterios | Con escenario | Con test | Código |
|---|---|---|---|---|---|---|
| US-001 | [specs/US-001-crear-proyecto.md](../specs/US-001-crear-proyecto.md) | borrador | 5 | 5 | 5 | `internal/domain/proyecto/`, `internal/app/`, `migraciones/` |
| US-002 | [specs/US-002-modificar-proyecto.md](../specs/US-002-modificar-proyecto.md) | borrador | 5 | 5 | 5 | `internal/domain/proyecto/`, `internal/app/` |
| US-003 | [specs/US-003-registrar-integrantes.md](../specs/US-003-registrar-integrantes.md) | borrador | 9 | 9 | 9 | `internal/domain/integrante/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00005_integrantes_id_por_defecto.sql` |
| US-005 | [specs/US-005-crear-item-backlog.md](../specs/US-005-crear-item-backlog.md) | borrador | 8 | 8 | 8 | `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/` |
| US-006 | [specs/US-006-editar-item-backlog.md](../specs/US-006-editar-item-backlog.md) | borrador | 7 | 5 | 6 | `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/` |
| US-008 | [specs/US-008-crear-sprint.md](../specs/US-008-crear-sprint.md) | borrador | 11 | 9 | 11 | `internal/domain/sprint/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/` |
| US-009 | [specs/US-009-asignar-historias-sprint.md](../specs/US-009-asignar-historias-sprint.md) | borrador | 8 | 7 | 8 | `internal/domain/sprint/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/` |
| US-010 | [specs/US-010-marcar-historias-completadas.md](../specs/US-010-marcar-historias-completadas.md) | borrador | 6 | 5 | 6 | `internal/domain/sprint/`, `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00008_items_completado_en.sql` |
| US-011 | [specs/US-011-cerrar-sprint.md](../specs/US-011-cerrar-sprint.md) | borrador | 6 | 0 | 0 | `internal/domain/sprint/`, `internal/app/`, `internal/adapters/postgres/` |
| US-013 | [specs/US-013-estimar-historia-fibonacci.md](../specs/US-013-estimar-historia-fibonacci.md) | borrador | 6 | 5 | 5 | `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/` |
| US-019 | [specs/US-019-registrar-esfuerzo.md](../specs/US-019-registrar-esfuerzo.md) | borrador | 11 | 9 | 7 | `internal/domain/esfuerzo/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/` |
| US-025 | [specs/US-025-story-points-sprint.md](../specs/US-025-story-points-sprint.md) | borrador | 4 | 4 | 1 | `internal/domain/metricas/` |
| US-026 | [specs/US-026-velocidad-equipo.md](../specs/US-026-velocidad-equipo.md) | borrador | 4 | 4 | 1 | `internal/domain/metricas/` |
| US-027 | [specs/US-027-horas-estimadas-reales.md](../specs/US-027-horas-estimadas-reales.md) | borrador | 6 | 0 | 0 | `internal/domain/metricas/` |
| US-028 | [specs/US-028-porcentaje-historias-completadas.md](../specs/US-028-porcentaje-historias-completadas.md) | borrador | 5 | 0 | 0 | `internal/domain/metricas/` |
| T-004 | [specs/T-004-layout-base.md](../specs/T-004-layout-base.md) | revision | 5 | 0 | 4 | `web/templates/`, `internal/adapters/http/`, `cmd/server/` |
| T-005 | [specs/T-005-validaciones-errores-uniformes.md](../specs/T-005-validaciones-errores-uniformes.md) | borrador | 8 | 0 | 7 | `internal/platform/errores/`, `internal/adapters/http/` |
| T-011 | [specs/T-011-pantallas-proyectos-backlog.md](../specs/T-011-pantallas-proyectos-backlog.md) | borrador | 5 | 0 | 4 | `internal/app/`, `internal/adapters/memoria/`, `internal/adapters/postgres/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| T-015 | [specs/T-015-pantallas-sprints.md](../specs/T-015-pantallas-sprints.md) | borrador | 5 | 0 | 0 | `internal/app/`, `internal/adapters/memoria/`, `internal/adapters/postgres/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |

## Detalle

### US-001 — Crear un proyecto con nombre, descripción y fechas de inicio y fin

- **Spec:** [specs/US-001-crear-proyecto.md](../specs/US-001-crear-proyecto.md) · **Issue:** #4
- **Escenarios BDD:** [features/US-001-crear-proyecto.feature](../features/US-001-crear-proyecto.feature)
- **Código:** `internal/domain/proyecto/`, `internal/app/`, `migraciones/`
- **Tests que la citan:** 14

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-001-1 | normal | Un proyecto con nombre válido y fechas coherentes se crea | TestRegistrarProyecto, TestCrearProyecto_Registra, TestNuevo_Valido |
| CA-001-2 | alternativo | Crear un segundo proyecto no modifica el primero | TestBase_RegistrarProyecto, TestRegistrarProyecto_NombreRepetido, TestNuevo_ProyectosIndependientes |
| CA-001-3 | límite | Un proyecto de un solo día es válido | TestNuevo_Valido |
| CA-001-4 | error | La fecha de fin anterior a la de inicio se rechaza | TestCrearProyecto_DatosInvalidos, TestNuevo_Invalido |
| CA-001-5 | error | Un nombre vacío se rechaza | TestCrearProyecto_DatosInvalidos, TestNuevo_Invalido |

### US-002 — Modificar los datos de un proyecto existente

- **Spec:** [specs/US-002-modificar-proyecto.md](../specs/US-002-modificar-proyecto.md) · **Issue:** #5
- **Escenarios BDD:** [features/US-002-modificar-proyecto.feature](../features/US-002-modificar-proyecto.feature)
- **Código:** `internal/domain/proyecto/`, `internal/app/`
- **Tests que la citan:** 14

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-002-1 | normal | Al modificar el nombre, el proyecto conserva sus fechas | TestActualizarProyecto, TestModificarProyecto_Actualiza, TestModificar_Valido |
| CA-002-2 | alternativo | Al modificar las fechas, el proyecto conserva su nombre | TestActualizarProyecto, TestModificarProyecto_Actualiza |
| CA-002-3 | límite | Las fechas de inicio y fin pueden ser iguales | TestModificar_Valido |
| CA-002-4 | error | Un nombre nuevo vacío se rechaza y el proyecto no cambia | TestModificarProyecto_DatosInvalidos, TestModificar_Invalido |
| CA-002-5 | error | Modificar un proyecto que no existe se rechaza | TestModificarProyecto_ProyectoInexistente |

### US-003 — Registrar integrantes en un proyecto

- **Spec:** [specs/US-003-registrar-integrantes.md](../specs/US-003-registrar-integrantes.md) · **Issue:** #6
- **Escenarios BDD:** [features/US-003-registrar-integrantes.feature](../features/US-003-registrar-integrantes.feature)
- **Código:** `internal/domain/integrante/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00005_integrantes_id_por_defecto.sql`
- **Tests que la citan:** 26

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-003-1 | normal | Registrar un integrante con datos válidos | TestBase_RegistrarIntegrante_Guarda, TestRegistrarIntegrante_Guarda, TestRegistrarIntegrante_Registra, TestNuevo_Valido |
| CA-003-2 | alternativo | Varios integrantes con roles distintos quedan en el mismo proyecto, Dos integrantes con el mismo nombre y distinto email son válidos | TestBase_RegistrarIntegrante_VariosIntegrantes, TestRegistrarIntegrante_Varios |
| CA-003-3 | alternativo | El mismo email puede registrarse en dos proyectos distintos | TestBase_RegistrarIntegrante_MismoEmailEnOtroProyecto, TestRegistrarIntegrante_MismoEmailEnOtroProyecto |
| CA-003-4 | límite | Un email con espacios y mayúsculas se acepta | TestRegistrarIntegrante_Registra, TestNuevo_Valido |
| CA-003-5 | error | Un nombre vacío se rechaza y no se registra nada, Un nombre con solo espacios se trata como vacío | TestRegistrarIntegrante_DatosInvalidos, TestNuevo_Invalido |
| CA-003-6 | error | Un email vacío se rechaza, Un email sin formato válido se rechaza, Un email con nombre visible se rechaza | TestNuevo_Invalido |
| CA-003-7 | error | Un rol fuera de los tres permitidos se rechaza, El rol no se normaliza: en mayúsculas se rechaza | TestRegistrarIntegrante_DatosInvalidos, TestNuevo_Invalido |
| CA-003-8 | error | Un email ya registrado en el proyecto se rechaza, El email repetido se rechaza sin importar mayúsculas ni espacios | TestBase_RegistrarIntegrante_EmailDuplicado, TestRegistrarIntegrante_EmailDuplicado, TestRegistrarIntegrante_ErroresDelRepositorio |
| CA-003-9 | error | Registrar un integrante en un proyecto que no existe se rechaza | TestBase_RegistrarIntegrante_ProyectoInexistente, TestRegistrarIntegrante_ProyectoInexistente |

### US-005 — Crear un ítem de Product Backlog

- **Spec:** [specs/US-005-crear-item-backlog.md](../specs/US-005-crear-item-backlog.md) · **Issue:** #7
- **Escenarios BDD:** [features/US-005-crear-item-backlog.feature](../features/US-005-crear-item-backlog.feature)
- **Código:** `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/`
- **Tests que la citan:** 30

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-005-1 | normal | Un ítem completo queda pendiente, sin estimar y con el número siguiente | TestRegistrarItem_PrimerItem, TestCrearItemBacklog_Registra, TestNuevoItem_DatosValidos |
| CA-005-2 | alternativo | Se puede crear sin descripción y sin criterios | TestRegistrarItem_Correlativo, TestNuevoItem_Descripcion, TestNuevoItem_SinDescripcionNiCriterios |
| CA-005-3 | alternativo | La numeración es independiente por proyecto | TestBase_RegistrarItem, TestRegistrarItem_Correlativo |
| CA-005-4 | límite | El título tiene como máximo 120 caracteres, Los espacios de los extremos del título se recortan | TestNuevoItem_LargoDelTitulo |
| CA-005-5 | error | Un título vacío se rechaza y no se registra nada | TestCrearItemBacklog_DatosInvalidos, TestNuevoItem_TituloVacio |
| CA-005-6 | error | Una prioridad fuera de MoSCoW se rechaza | TestNuevoItem_Prioridad |
| CA-005-7 | error | Un criterio vacío se rechaza indicando su posición | TestNuevoItem_CriteriosVacios |
| CA-005-8 | error | Un proyecto inexistente se rechaza antes que los demás errores | TestRegistrarItem_ProyectoInexistente, TestCrearItemBacklog_ProyectoInexistente |

### US-006 — Editar y repriorizar ítems del backlog

- **Spec:** [specs/US-006-editar-item-backlog.md](../specs/US-006-editar-item-backlog.md) · **Issue:** #8
- **Escenarios BDD:** [features/US-006-editar-item-backlog.feature](../features/US-006-editar-item-backlog.feature)
- **Código:** `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`
- **Tests que la citan:** 25

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-006-1 | normal | Una historia editada queda con los datos nuevos y conserva número, estado y estimación | TestBase_GuardarEdicion, TestGuardarEdicion, TestEditarItem_Edita, TestEditar |
| CA-006-2 | alternativo | Repriorizar cambia solo la prioridad | TestEditar_Repriorizar |
| CA-006-3 | alternativo | — | — |
| CA-006-4 | límite | El título editado tiene como máximo 120 caracteres | TestBase_GuardarEdicion_CriteriosVacios, TestGuardarEdicion_CriteriosVacios, TestEditar_CriteriosVacios, TestEditar_DatosInvalidos, +1 más |
| CA-006-5 | error | Con datos inválidos se rechaza y la historia queda como estaba | TestEditarItem_Rechazos, TestEditar_DatosInvalidos |
| CA-006-6 | error | — | TestGuardarEdicion_ItemCompletado, TestEditarItem_Rechazos, TestEditar_ItemCompletado |
| CA-006-7 | error | Editar una historia que no existe se rechaza y no cambia nada | TestBase_GuardarEdicion_ItemInexistente, TestGuardarEdicion_ItemInexistente, TestEditarItem_ItemInexistente |

### US-008 — Crear un Sprint con Sprint Goal y fechas

- **Spec:** [specs/US-008-crear-sprint.md](../specs/US-008-crear-sprint.md) · **Issue:** #9
- **Escenarios BDD:** [features/US-008-crear-sprint.feature](../features/US-008-crear-sprint.feature)
- **Código:** `internal/domain/sprint/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/`
- **Tests que la citan:** 39

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-008-1 | normal | Un sprint válido queda abierto y con el número siguiente del proyecto | TestBase_RegistrarSprint, TestRegistrarSprint_Primero, TestCrearSprint_Registra, TestNuevo_DatosValidos |
| CA-008-2 | alternativo | El primer sprint de cada proyecto es el Sprint 1 | TestBase_RegistrarSprint_NumeracionPorProyecto, TestRegistrarSprint_Primero, TestRegistrarSprint_NumeracionPorProyecto |
| CA-008-3 | límite | Un sprint dura entre 1 y 28 días | TestNuevo_Duracion |
| CA-008-4 | límite | — | TestRegistrarSprint_UltimoCerrado, TestVerificarUltimoSprint_Cerrado |
| CA-008-5 | límite | Un sprint que ocupa exactamente las fechas del proyecto es válido | TestNuevo_ExtremosDelProyectoEstanDentro |
| CA-008-6 | error | Un Sprint Goal vacío se rechaza | TestNuevo_SprintGoal |
| CA-008-7 | error | La fecha de fin anterior a la de inicio se rechaza | TestNuevo_Fechas |
| CA-008-8 | error | Un sprint fuera de las fechas del proyecto se rechaza, Con varios datos inválidos se informan todos juntos | TestCrearSprint_UsaLasFechasDelProyecto, TestNuevo_FueraDelProyecto |
| CA-008-9 | error | — | TestRegistrarSprint_UltimoCerrado |
| CA-008-10 | error | Crear un sprint en un proyecto que no existe se rechaza, sin mirar los demás datos | TestBase_RegistrarSprint_ProyectoInexistente, TestRegistrarSprint_ProyectoInexistente, TestCrearSprint_ProyectoInexistente |
| CA-008-11 | error | Con el último sprint abierto no se puede crear otro | TestBase_RegistrarSprint_UltimoAbierto, TestRegistrarSprint_UltimoAbierto, TestVerificarUltimoSprint_Abierto |

### US-009 — Asignar historias del backlog a un Sprint

- **Spec:** [specs/US-009-asignar-historias-sprint.md](../specs/US-009-asignar-historias-sprint.md) · **Issue:** #10
- **Escenarios BDD:** [features/US-009-asignar-historias-sprint.feature](../features/US-009-asignar-historias-sprint.feature)
- **Código:** `internal/domain/sprint/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/`
- **Tests que la citan:** 35

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-009-1 | normal | Una historia estimada y con criterios queda en el sprint | TestBase_AsignarItemASprint, TestAsignarItemASprint_Asigna, TestVerificarAsignacion_Valida |
| CA-009-2 | alternativo | Se pueden asignar varias historias al mismo sprint | TestBase_AsignarItemASprint_Varios, TestAsignarItemASprint_VariosItemsAlMismoSprint |
| CA-009-3 | límite | Una historia con un solo criterio y un solo story point cumple la Definition of Ready | TestVerificarAsignacion_Valida |
| CA-009-4 | error | Una historia sin estimar se rechaza y sigue en el backlog | TestAsignarItemASprint_DefinitionOfReady, TestVerificarAsignacion_DefinitionOfReady |
| CA-009-5 | error | Una historia sin criterios de aceptación se rechaza y sigue en el backlog, Una historia sin estimar y sin criterios informa las dos faltas juntas | TestAsignarItemASprint_DefinitionOfReady, TestVerificarAsignacion_DefinitionOfReady |
| CA-009-6 | error | — | TestAsignarItemASprint_SprintCerrado, TestVerificarAsignacion_SprintCerrado |
| CA-009-7 | error | Una historia que ya está en un sprint no se asigna otra vez | TestBase_AsignarItemASprint_YaAsignado, TestAsignarItemASprint_YaAsignado, TestVerificarAsignacion_ItemYaAsignado |
| CA-009-8 | error | Asignar a un sprint que no existe se rechaza, Asignar una historia que no existe se rechaza, Una historia de otro proyecto no se asigna a un sprint de este | TestAsignarItemASprint_LoQueNoExiste |

### US-010 — Marcar historias como completadas dentro de un Sprint

- **Spec:** [specs/US-010-marcar-historias-completadas.md](../specs/US-010-marcar-historias-completadas.md) · **Issue:** #18
- **Escenarios BDD:** [features/US-010-marcar-historias-completadas.feature](../features/US-010-marcar-historias-completadas.feature)
- **Código:** `internal/domain/sprint/`, `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00008_items_completado_en.sql`
- **Tests que la citan:** 33

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-010-1 | normal | Una historia del sprint abierto queda completada | TestBase_GuardarCambioDeEstado_Marca, TestGuardarCambioDeEstado_Marca, TestMarcarCompletado_Marca, TestCompletar, +1 más |
| CA-010-2 | alternativo | Una historia completada se desmarca y vuelve a pendiente | TestBase_GuardarCambioDeEstado_Desmarca, TestGuardarCambioDeEstado_Desmarca, TestDesmarcarCompletado_Desmarca, TestReabrir, +1 más |
| CA-010-3 | límite | Marcar una historia que ya está completada no es un error | TestBase_GuardarCambioDeEstado_SegundoMarcadoConservaLaFecha, TestGuardarCambioDeEstado_MarcarDosVeces, TestMarcarCompletado_YaCompletado, TestCompletar_YaCompletado |
| CA-010-4 | error | Una historia que no está en el sprint no se marca | TestGuardarCambioDeEstado_Reglas, TestMarcarYDesmarcar_ReglasDelDominio, TestVerificarMarcado_ItemFueraDelSprint |
| CA-010-5 | error | — | TestGuardarCambioDeEstado_Reglas, TestMarcarYDesmarcar_ReglasDelDominio, TestVerificarMarcado_SprintCerrado |
| CA-010-6 | error | Marcar una historia que no existe se rechaza, Marcar una historia de un proyecto sin sprints se rechaza y no cambia nada | TestBase_GuardarCambioDeEstado_Inexistentes, TestGuardarCambioDeEstado_Inexistentes, TestMarcarYDesmarcar_LoQueNoExiste |

### US-011 — Cerrar un Sprint y devolver las historias no completadas al backlog

- **Spec:** [specs/US-011-cerrar-sprint.md](../specs/US-011-cerrar-sprint.md) · **Issue:** #19
- **Escenarios BDD:** ninguno
- **Código:** `internal/domain/sprint/`, `internal/app/`, `internal/adapters/postgres/`
- **Tests que la citan:** 3

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-011-1 | normal | — | — |
| CA-011-2 | alternativo | — | — |
| CA-011-3 | alternativo | — | — |
| CA-011-4 | límite | — | — |
| CA-011-5 | error | — | — |
| CA-011-6 | error | — | — |

### US-013 — Estimar una historia con Story Points en escala Fibonacci

- **Spec:** [specs/US-013-estimar-historia-fibonacci.md](../specs/US-013-estimar-historia-fibonacci.md) · **Issue:** #21
- **Escenarios BDD:** [features/US-013-estimar-historia-fibonacci.feature](../features/US-013-estimar-historia-fibonacci.feature)
- **Código:** `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`
- **Tests que la citan:** 29

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-013-1 | normal | Una historia sin estimar queda con los story points elegidos | TestEstimarItem_Estima, TestNuevosStoryPoints_ValoresDeLaEscala, TestEstimar |
| CA-013-2 | alternativo | Se puede volver a estimar y la última estimación reemplaza a la anterior | TestEstimarItem_ReEstima |
| CA-013-3 | límite | La escala va de 1 a 13 y no incluye el 0 ni el 21 | — |
| CA-013-4 | error | Un valor fuera de la escala se rechaza y la historia conserva su estimación, Un valor fuera de la escala no deja la historia estimada | TestEstimarItem_Rechazos, TestNuevosStoryPoints_FueraDeLaEscala, TestEstimar_FueraDeEscala |
| CA-013-5 | error | — | TestEstimar_ItemCompletado |
| CA-013-6 | error | Estimar una historia que no existe se rechaza | TestBase_GuardarEstimacion_ItemInexistente, TestObtenerItem_Inexistente, TestGuardarEstimacion_ItemInexistente, TestEstimarItem_ItemInexistente |

### US-019 — Registrar esfuerzo con integrante, fecha, actividad y horas

- **Spec:** [specs/US-019-registrar-esfuerzo.md](../specs/US-019-registrar-esfuerzo.md) · **Issue:** #22
- **Escenarios BDD:** [features/US-019-registrar-esfuerzo.feature](../features/US-019-registrar-esfuerzo.feature)
- **Código:** `internal/domain/esfuerzo/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/`
- **Tests que la citan:** 33

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-019-1 | normal | Un integrante registra sus horas de una historia | TestBase_RegistrarEsfuerzo, TestRegistrarEsfuerzo_Guarda, TestRegistrarEsfuerzo_Registra, TestNuevo_DatosValidos |
| CA-019-2 | alternativo | Se pueden registrar varias veces horas del mismo integrante el mismo día, Dos registros del mismo día que suman más de 24 horas se aceptan | TestNuevo_DatosValidos |
| CA-019-3 | alternativo | — | — |
| CA-019-4 | límite | Las horas van de 0,25 a 24 en múltiplos de 0,25 | — |
| CA-019-5 | límite | La fecha no puede ser posterior a hoy | TestNuevo_Fecha |
| CA-019-6 | error | Horas negativas se rechazan y no se registra nada | TestRegistrarEsfuerzo_DatosInvalidos |
| CA-019-7 | error | Horas que no son múltiplo de 0,25 se rechazan | — |
| CA-019-8 | error | Una actividad vacía se rechaza, Con varios datos inválidos se informan todos juntos | TestRegistrarEsfuerzo_DatosInvalidos |
| CA-019-11 | alternativo | — | — |
| CA-019-9 | error | Un integrante de otro proyecto no puede registrar esfuerzo, No se registra esfuerzo sobre una historia de otro proyecto | TestRegistrarEsfuerzo_LoQueDesaparece, TestRegistrarEsfuerzo_LoQueNoExiste |
| CA-019-10 | error | Registrar esfuerzo en un proyecto que no existe se rechaza | TestRegistrarEsfuerzo_LoQueDesaparece, TestRegistrarEsfuerzo_LoQueNoExiste |

### US-025 — Calcular Story Points planificados y completados de un Sprint

- **Spec:** [specs/US-025-story-points-sprint.md](../specs/US-025-story-points-sprint.md) · **Issue:** #11
- **Escenarios BDD:** [features/US-025-story-points-sprint.feature](../features/US-025-story-points-sprint.feature)
- **Código:** `internal/domain/metricas/`
- **Tests que la citan:** 2

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-025-1 | normal | Los planificados suman todas las historias del sprint, Los completados suman solo las historias completadas | — |
| CA-025-2 | alternativo | Con todas las historias completadas, los completados igualan a los planificados | — |
| CA-025-3 | límite | Un sprint sin historias no suma nada y no es un error | — |
| CA-025-4 | error | Una historia con story points negativos se rechaza | TestCalcularResumenSprint_StoryPointsNegativos |

### US-026 — Calcular la velocidad del equipo sobre los sprints cerrados

- **Spec:** [specs/US-026-velocidad-equipo.md](../specs/US-026-velocidad-equipo.md) · **Issue:** #12
- **Escenarios BDD:** [features/US-026-velocidad-equipo.feature](../features/US-026-velocidad-equipo.feature)
- **Código:** `internal/domain/metricas/`
- **Tests que la citan:** 5

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-026-1 | normal | La velocidad es el promedio de los story points completados | — |
| CA-026-2 | alternativo | Con un solo sprint cerrado la velocidad es lo que completó ese sprint | — |
| CA-026-3 | límite | Sin sprints cerrados la velocidad es cero, no un error | — |
| CA-026-4 | error | Un sprint con más completados que planificados es un dato corrupto | TestCalcularVelocidad_ResumenInconsistente |

### US-027 — Calcular horas estimadas, horas reales y la desviación entre ambas

- **Spec:** [specs/US-027-horas-estimadas-reales.md](../specs/US-027-horas-estimadas-reales.md) · **Issue:** #33
- **Escenarios BDD:** ninguno
- **Código:** `internal/domain/metricas/`
- **Tests que la citan:** 0

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-027-1 | normal | — | — |
| CA-027-2 | alternativo | — | — |
| CA-027-3 | límite | — | — |
| CA-027-4 | límite | — | — |
| CA-027-5 | error | — | — |
| CA-027-6 | error | — | — |

### US-028 — Calcular el porcentaje de historias completadas

- **Spec:** [specs/US-028-porcentaje-historias-completadas.md](../specs/US-028-porcentaje-historias-completadas.md) · **Issue:** #34
- **Escenarios BDD:** ninguno
- **Código:** `internal/domain/metricas/`
- **Tests que la citan:** 0

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-028-1 | normal | — | — |
| CA-028-2 | alternativo | — | — |
| CA-028-3 | límite | — | — |
| CA-028-4 | límite | — | — |
| CA-028-5 | error | — | — |

### T-004 — Layout base, navegación y sistema visual con Tailwind

- **Spec:** [specs/T-004-layout-base.md](../specs/T-004-layout-base.md) · **Issue:** #128 (spec) · tarea: #14
- **Escenarios BDD:** ninguno
- **Código:** `web/templates/`, `internal/adapters/http/`, `cmd/server/`
- **Tests que la citan:** 31

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-T004-1 | normal | — | TestResponder_PaginaCompleta, TestRouter_Inicio, TestInicio |
| CA-T004-2 | alternativo | — | TestResponder_Fragmento, TestRouter_InicioComoFragmento |
| CA-T004-3 | límite | — | TestRouter_InicioBoosted |
| CA-T004-4 | error | — | TestResponder_FallaElRender, TestRouter_RutaInexistente, TestNoEncontrado |
| CA-T004-5 | plataforma | — | — |

### T-005 — Validaciones de formulario y manejo uniforme de errores

- **Spec:** [specs/T-005-validaciones-errores-uniformes.md](../specs/T-005-validaciones-errores-uniformes.md) · **Issue:** #15
- **Escenarios BDD:** ninguno
- **Código:** `internal/platform/errores/`, `internal/adapters/http/`
- **Tests que la citan:** 37

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-T005-1 | normal | — | TestResponderError_ValidacionEnJSON |
| CA-T005-2 | normal | — | TestResponderError_ErroresUnidos, TestClasificar_ErroresUnidos |
| CA-T005-3 | alternativo | — | TestResponderError_EstadoPorTipo |
| CA-T005-4 | alternativo | — | TestResponderError_ErrorDeFormulario |
| CA-T005-5 | límite | — | — |
| CA-T005-6 | error | — | TestResponderError_InternoNoFiltraNiOculta |
| CA-T005-7 | error | — | TestRecuperar_PanicResponde500Uniforme, TestRecuperar_ElServidorSigueAtendiendo |
| CA-T005-8 | error | — | TestCatalogo_CubreTodosLosErrores |

### T-011 — Pantallas de proyectos, integrantes y Product Backlog

- **Spec:** [specs/T-011-pantallas-proyectos-backlog.md](../specs/T-011-pantallas-proyectos-backlog.md) · **Issue:** #79 (tarea) · #140 (spec)
- **Escenarios BDD:** ninguno
- **Código:** `internal/app/`, `internal/adapters/memoria/`, `internal/adapters/postgres/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/`
- **Tests que la citan:** 25

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-T011-1 | normal | — | TestPantallas_Recorrido, TestConsultarProyecto_ConIntegrantesOrdenados |
| CA-T011-2 | alternativo | — | TestPantallas_RedireccionesConYSinHTMX, TestPantallas_PaginaOFragmento |
| CA-T011-3 | límite | — | — |
| CA-T011-4 | error | — | TestPantallas_ErroresDeValidacion, TestPantallas_EmailRepetido |
| CA-T011-5 | plataforma | — | TestListarProyectos |

### T-015 — Pantallas de sprints

- **Spec:** [specs/T-015-pantallas-sprints.md](../specs/T-015-pantallas-sprints.md) · **Issue:** #141 (tarea) · #142 (spec)
- **Escenarios BDD:** ninguno
- **Código:** `internal/app/`, `internal/adapters/memoria/`, `internal/adapters/postgres/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/`
- **Tests que la citan:** 0

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-T015-1 | normal | — | — |
| CA-T015-2 | alternativo | — | — |
| CA-T015-3 | límite | — | — |
| CA-T015-4 | error | — | — |
| CA-T015-5 | plataforma | — | — |
