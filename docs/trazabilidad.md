# Matriz de trazabilidad

> Generada con `make trace` a partir de `specs/`, `features/` y los comentarios de los tests.
> No se edita a mano: se regenera. Enlaza Historia → Spec → Criterios de aceptación →
> Escenarios BDD → Tests → Código (la guía de la cátedra la pide como evidencia).

Total: 39 especificaciones, 267 criterios de aceptación, 190 con escenario BDD y 223 con test.

Un criterio sin escenario BDD o sin test se muestra con `—`: o falta la evidencia, o el criterio
se cubre de otra forma y la spec lo dice (por ejemplo, las tareas técnicas no llevan `.feature`).

## Resumen por historia

| Historia | Spec | Estado | Criterios | Con escenario | Con test | Código |
|---|---|---|---|---|---|---|
| US-001 | [specs/US-001-crear-proyecto.md](../specs/US-001-crear-proyecto.md) | borrador | 5 | 5 | 5 | `internal/domain/proyecto/`, `internal/app/`, `migraciones/` |
| US-002 | [specs/US-002-modificar-proyecto.md](../specs/US-002-modificar-proyecto.md) | borrador | 5 | 5 | 5 | `internal/domain/proyecto/`, `internal/app/` |
| US-003 | [specs/US-003-registrar-integrantes.md](../specs/US-003-registrar-integrantes.md) | borrador | 9 | 9 | 9 | `internal/domain/integrante/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00005_integrantes_id_por_defecto.sql` |
| US-004 | [specs/US-004-estado-proyecto.md](../specs/US-004-estado-proyecto.md) | borrador | 6 | 6 | 6 | `internal/domain/proyecto/`, `internal/app/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| US-005 | [specs/US-005-crear-item-backlog.md](../specs/US-005-crear-item-backlog.md) | borrador | 8 | 8 | 8 | `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/` |
| US-006 | [specs/US-006-editar-item-backlog.md](../specs/US-006-editar-item-backlog.md) | borrador | 7 | 5 | 6 | `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/` |
| US-008 | [specs/US-008-crear-sprint.md](../specs/US-008-crear-sprint.md) | borrador | 11 | 11 | 11 | `internal/domain/sprint/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/` |
| US-009 | [specs/US-009-asignar-historias-sprint.md](../specs/US-009-asignar-historias-sprint.md) | borrador | 8 | 8 | 8 | `internal/domain/sprint/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/` |
| US-010 | [specs/US-010-marcar-historias-completadas.md](../specs/US-010-marcar-historias-completadas.md) | borrador | 6 | 6 | 6 | `internal/domain/sprint/`, `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00008_items_completado_en.sql` |
| US-011 | [specs/US-011-cerrar-sprint.md](../specs/US-011-cerrar-sprint.md) | borrador | 6 | 6 | 5 | `internal/domain/sprint/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/` |
| US-012 | [specs/US-012-sprints-anteriores.md](../specs/US-012-sprints-anteriores.md) | borrador | 6 | 6 | 6 | `internal/app/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| US-013 | [specs/US-013-estimar-historia-fibonacci.md](../specs/US-013-estimar-historia-fibonacci.md) | borrador | 6 | 5 | 5 | `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/` |
| US-014 | [specs/US-014-iniciar-sesion-planning-poker.md](../specs/US-014-iniciar-sesion-planning-poker.md) | borrador | 7 | 7 | 7 | `internal/domain/planningpoker/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/00009_sesiones_planning_poker.sql` |
| US-015 | [specs/US-015-voto-individual-oculto.md](../specs/US-015-voto-individual-oculto.md) | borrador | 7 | 7 | 7 | `internal/domain/planningpoker/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/00010_votos_planning_poker.sql` |
| US-016 | [specs/US-016-revelar-estimaciones.md](../specs/US-016-revelar-estimaciones.md) | borrador | 8 | 8 | 8 | `internal/domain/planningpoker/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/` |
| US-017 | [specs/US-017-nueva-ronda.md](../specs/US-017-nueva-ronda.md) | borrador | 5 | 4 | 5 | `internal/domain/planningpoker/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/` |
| US-018 | [specs/US-018-registrar-estimacion-acordada.md](../specs/US-018-registrar-estimacion-acordada.md) | borrador | 8 | 8 | 8 | `internal/domain/planningpoker/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00011_sesiones_valor_acordado.sql` |
| US-019 | [specs/US-019-registrar-esfuerzo.md](../specs/US-019-registrar-esfuerzo.md) | borrador | 11 | 9 | 7 | `internal/domain/esfuerzo/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/` |
| US-020 | [specs/US-020-consultar-esfuerzo.md](../specs/US-020-consultar-esfuerzo.md) | borrador | 9 | 9 | 9 | `internal/domain/esfuerzo/`, `internal/app/consultar_esfuerzo.go` |
| US-021 | [specs/US-021-comparar-esfuerzo.md](../specs/US-021-comparar-esfuerzo.md) | borrador | 9 | 8 | 9 | `internal/domain/metricas/`, `internal/app/comparar_esfuerzo.go` |
| US-022 | [specs/US-022-registrar-defecto.md](../specs/US-022-registrar-defecto.md) | borrador | 12 | 12 | 12 | `internal/domain/defecto/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00012_defectos.sql` |
| US-023 | [specs/US-023-cambiar-estado-defecto.md](../specs/US-023-cambiar-estado-defecto.md) | borrador | 14 | 13 | 14 | `internal/domain/defecto/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00013_defectos_estado.sql` |
| US-024 | [specs/US-024-listar-defectos.md](../specs/US-024-listar-defectos.md) | borrador | 6 | 0 | 0 | `internal/app/`, `internal/domain/defecto/`, `internal/adapters/http/`, `web/templates/` |
| US-025 | [specs/US-025-story-points-sprint.md](../specs/US-025-story-points-sprint.md) | borrador | 4 | 4 | 1 | `internal/domain/metricas/` |
| US-026 | [specs/US-026-velocidad-equipo.md](../specs/US-026-velocidad-equipo.md) | borrador | 4 | 4 | 1 | `internal/domain/metricas/` |
| US-027 | [specs/US-027-horas-estimadas-reales.md](../specs/US-027-horas-estimadas-reales.md) | borrador | 6 | 6 | 6 | `internal/domain/metricas/` |
| US-028 | [specs/US-028-porcentaje-historias-completadas.md](../specs/US-028-porcentaje-historias-completadas.md) | borrador | 5 | 5 | 1 | `internal/domain/metricas/` |
| US-029 | [specs/US-029-defectos-por-sprint.md](../specs/US-029-defectos-por-sprint.md) | borrador | 6 | 6 | 6 | `internal/domain/metricas/` |
| US-030 | [specs/US-030-dashboard.md](../specs/US-030-dashboard.md) | borrador | 7 | 0 | 0 | `internal/app/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| US-033 | [specs/US-033-reporte-pantalla.md](../specs/US-033-reporte-pantalla.md) | borrador | 5 | 0 | 0 | `internal/app/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| US-034 | [specs/US-034-vista-imprimible.md](../specs/US-034-vista-imprimible.md) | borrador | 4 | 0 | 0 | `web/static/css/input.css`, `web/static/js/`, `web/templates/`, `internal/adapters/http/` |
| T-004 | [specs/T-004-layout-base.md](../specs/T-004-layout-base.md) | revision | 5 | 0 | 4 | `web/templates/`, `internal/adapters/http/`, `cmd/server/` |
| T-005 | [specs/T-005-validaciones-errores-uniformes.md](../specs/T-005-validaciones-errores-uniformes.md) | borrador | 8 | 0 | 7 | `internal/platform/errores/`, `internal/adapters/http/` |
| T-011 | [specs/T-011-pantallas-proyectos-backlog.md](../specs/T-011-pantallas-proyectos-backlog.md) | borrador | 5 | 0 | 4 | `internal/app/`, `internal/adapters/memoria/`, `internal/adapters/postgres/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| T-015 | [specs/T-015-pantallas-sprints.md](../specs/T-015-pantallas-sprints.md) | borrador | 5 | 0 | 4 | `internal/app/`, `internal/adapters/memoria/`, `internal/adapters/postgres/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| T-016 | [specs/T-016-cerrar-sprint.md](../specs/T-016-cerrar-sprint.md) | borrador | 5 | 0 | 4 | `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| T-019 | [specs/T-019-pantallas-esfuerzo.md](../specs/T-019-pantallas-esfuerzo.md) | borrador | 6 | 0 | 6 | `internal/adapters/http/`, `web/templates/`, `cmd/server/` |
| T-020 | [specs/T-020-pantallas-planning-poker.md](../specs/T-020-pantallas-planning-poker.md) | borrador | 7 | 0 | 7 | `internal/adapters/http/`, `web/templates/`, `internal/app/`, `cmd/server/` |
| T-021 | [specs/T-021-pantallas-defectos.md](../specs/T-021-pantallas-defectos.md) | borrador | 6 | 0 | 6 | `internal/app/`, `internal/adapters/memoria/`, `internal/adapters/postgres/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/` |

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

### US-004 — Consultar el estado general de un proyecto

- **Spec:** [specs/US-004-estado-proyecto.md](../specs/US-004-estado-proyecto.md) · **Issue:** #16
- **Escenarios BDD:** [features/US-004-estado-proyecto.feature](../features/US-004-estado-proyecto.feature)
- **Código:** `internal/domain/proyecto/`, `internal/app/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/`
- **Tests que la citan:** 18

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-004-1 | normal | Un proyecto en marcha está en curso y dice cuántos días faltan para el fin | TestPantallasEstado_Tarjeta, TestConsultarEstado_EstadoDeLasFechas, TestEstadoEn |
| CA-004-2 | alternativo | Un proyecto que todavía no empezó está no iniciado y dice en cuántos días empieza, Un proyecto que ya terminó está finalizado y dice hace cuántos días | TestPantallasEstado_Textos, TestConsultarEstado_EstadoDeLasFechas, TestEstadoEn |
| CA-004-3 | alternativo | El avance cuenta las historias por estado, los story points y el porcentaje, y muestra el sprint abierto | TestPantallasEstado_Tarjeta, TestConsultarEstado_Avance, TestConsultarEstado_PorcentajeConUnDecimal |
| CA-004-4 | límite | El día de inicio ya está en curso, El día de fin está en curso y no quedan días, Un proyecto de un solo día está en curso ese día, +1 más | TestPantallasEstado_Textos, TestEstadoEn |
| CA-004-5 | límite | Un proyecto sin historias ni sprints muestra todo en cero y no es un error, Con todos los sprints cerrados no hay sprint abierto | TestPantallasEstado_ProyectoVacio, TestConsultarEstado_ProyectoVacio |
| CA-004-6 | error | Un proyecto que no existe se rechaza | TestPantallasEstado_ProyectoInexistente, TestConsultarEstado_ProyectoInexistente |

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
| CA-008-4 | límite | El sprint siguiente empieza después del fin del último | TestRegistrarSprint_UltimoCerrado, TestVerificarUltimoSprint_Cerrado |
| CA-008-5 | límite | Un sprint que ocupa exactamente las fechas del proyecto es válido | TestNuevo_ExtremosDelProyectoEstanDentro |
| CA-008-6 | error | Un Sprint Goal vacío se rechaza | TestNuevo_SprintGoal |
| CA-008-7 | error | La fecha de fin anterior a la de inicio se rechaza | TestNuevo_Fechas |
| CA-008-8 | error | Un sprint fuera de las fechas del proyecto se rechaza, Con varios datos inválidos se informan todos juntos | TestCrearSprint_UsaLasFechasDelProyecto, TestNuevo_FueraDelProyecto |
| CA-008-9 | error | Empezar antes de que termine el último sprint se rechaza | TestRegistrarSprint_UltimoCerrado |
| CA-008-10 | error | Crear un sprint en un proyecto que no existe se rechaza, sin mirar los demás datos | TestBase_RegistrarSprint_ProyectoInexistente, TestRegistrarSprint_ProyectoInexistente, TestCrearSprint_ProyectoInexistente |
| CA-008-11 | error | Con el último sprint abierto no se puede crear otro | TestBase_RegistrarSprint_UltimoAbierto, TestRegistrarSprint_UltimoAbierto, TestVerificarUltimoSprint_Abierto |

### US-009 — Asignar historias del backlog a un Sprint

- **Spec:** [specs/US-009-asignar-historias-sprint.md](../specs/US-009-asignar-historias-sprint.md) · **Issue:** #10
- **Escenarios BDD:** [features/US-009-asignar-historias-sprint.feature](../features/US-009-asignar-historias-sprint.feature)
- **Código:** `internal/domain/sprint/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/`
- **Tests que la citan:** 38

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-009-1 | normal | Una historia estimada y con criterios queda en el sprint | TestBase_AsignarItemASprint, TestAsignarItemASprint_Asigna, TestVerificarAsignacion_Valida |
| CA-009-2 | alternativo | Se pueden asignar varias historias al mismo sprint | TestBase_AsignarItemASprint_Varios, TestAsignarItemASprint_VariosItemsAlMismoSprint |
| CA-009-3 | límite | Una historia con un solo criterio y un solo story point cumple la Definition of Ready | TestVerificarAsignacion_Valida |
| CA-009-4 | error | Una historia sin estimar se rechaza y sigue en el backlog | TestAsignarItemASprint_DefinitionOfReady, TestVerificarAsignacion_DefinitionOfReady |
| CA-009-5 | error | Una historia sin criterios de aceptación se rechaza y sigue en el backlog, Una historia sin estimar y sin criterios informa las dos faltas juntas | TestAsignarItemASprint_DefinitionOfReady, TestVerificarAsignacion_DefinitionOfReady |
| CA-009-6 | error | Asignar a un sprint cerrado se rechaza y la historia sigue en el backlog | TestAsignarItemASprint_SprintCerrado, TestVerificarAsignacion_SprintCerrado |
| CA-009-7 | error | Una historia que ya está en un sprint no se asigna otra vez | TestBase_AsignarItemASprint_YaAsignado, TestAsignarItemASprint_YaAsignado, TestVerificarAsignacion_ItemYaAsignado |
| CA-009-8 | error | Asignar a un sprint que no existe se rechaza, Asignar una historia que no existe se rechaza, Una historia de otro proyecto no se asigna a un sprint de este | TestAsignarItemASprint_LoQueNoExiste |

### US-010 — Marcar historias como completadas dentro de un Sprint

- **Spec:** [specs/US-010-marcar-historias-completadas.md](../specs/US-010-marcar-historias-completadas.md) · **Issue:** #18
- **Escenarios BDD:** [features/US-010-marcar-historias-completadas.feature](../features/US-010-marcar-historias-completadas.feature)
- **Código:** `internal/domain/sprint/`, `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00008_items_completado_en.sql`
- **Tests que la citan:** 36

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-010-1 | normal | Una historia del sprint abierto queda completada | TestBase_GuardarCambioDeEstado_Marca, TestGuardarCambioDeEstado_Marca, TestMarcarCompletado_Marca, TestCompletar, +1 más |
| CA-010-2 | alternativo | Una historia completada se desmarca y vuelve a pendiente | TestBase_GuardarCambioDeEstado_Desmarca, TestGuardarCambioDeEstado_Desmarca, TestDesmarcarCompletado_Desmarca, TestReabrir, +1 más |
| CA-010-3 | límite | Marcar una historia que ya está completada no es un error | TestBase_GuardarCambioDeEstado_SegundoMarcadoConservaLaFecha, TestGuardarCambioDeEstado_MarcarDosVeces, TestMarcarCompletado_YaCompletado, TestCompletar_YaCompletado |
| CA-010-4 | error | Una historia que no está en el sprint no se marca | TestGuardarCambioDeEstado_Reglas, TestMarcarYDesmarcar_ReglasDelDominio, TestVerificarMarcado_ItemFueraDelSprint |
| CA-010-5 | error | Con el sprint cerrado no se desmarca una historia completada, Con el sprint cerrado no se marca una historia que no se completó | TestGuardarCambioDeEstado_Reglas, TestMarcarYDesmarcar_ReglasDelDominio, TestVerificarMarcado_SprintCerrado |
| CA-010-6 | error | Marcar una historia que no existe se rechaza, Marcar una historia de un proyecto sin sprints se rechaza y no cambia nada | TestBase_GuardarCambioDeEstado_Inexistentes, TestGuardarCambioDeEstado_Inexistentes, TestMarcarYDesmarcar_LoQueNoExiste |

### US-011 — Cerrar un Sprint y devolver las historias no completadas al backlog

- **Spec:** [specs/US-011-cerrar-sprint.md](../specs/US-011-cerrar-sprint.md) · **Issue:** #19
- **Escenarios BDD:** [features/US-011-cerrar-sprint.feature](../features/US-011-cerrar-sprint.feature)
- **Código:** `internal/domain/sprint/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`
- **Tests que la citan:** 27

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-011-1 | normal | Al cerrar, las completadas quedan completadas y las demás vuelven al backlog | TestBase_GuardarCierre, TestGuardarCierre, TestCerrarSprint_Cierra, TestCerrar |
| CA-011-2 | alternativo | Después de cerrar se crea el sprint siguiente y una historia devuelta se asigna a él | TestBase_GuardarCierre_HistoriaDevueltaSeAsignaAlSiguiente, TestGuardarCierre_HistoriaDevueltaSeAsignaAlSiguiente |
| CA-011-3 | alternativo | Los planificados de un sprint cerrado incluyen las historias no completadas, La velocidad se calcula con lo completado en cada sprint cerrado | — |
| CA-011-4 | límite | Un sprint sin historias se cierra | TestBase_GuardarCierre_SinHistorias, TestGuardarCierre_SinHistorias, TestCerrarSprint_SinHistorias, TestCerrar_SinHistorias |
| CA-011-5 | error | Un sprint ya cerrado no se vuelve a cerrar | TestBase_GuardarCierre_Rechazos, TestGuardarCierre_Rechazos, TestCerrarSprint_SprintCerrado, TestCerrar_SprintCerrado |
| CA-011-6 | error | Cerrar un sprint que no existe se rechaza | TestBase_GuardarCierre_Rechazos, TestGuardarCierre_Rechazos, TestCerrarSprint_SprintInexistente |

### US-012 — Consultar los Sprints anteriores y su resultado

- **Spec:** [specs/US-012-sprints-anteriores.md](../specs/US-012-sprints-anteriores.md) · **Issue:** #20
- **Escenarios BDD:** [features/US-012-sprints-anteriores.feature](../features/US-012-sprints-anteriores.feature)
- **Código:** `internal/app/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/`
- **Tests que la citan:** 12

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-012-1 | normal | Los sprints cerrados salen del más reciente al más antiguo, cada uno con su resultado | TestPantallasAnteriores_Historial, TestConsultarSprintsAnteriores |
| CA-012-2 | alternativo | La velocidad es el promedio de los story points completados de los sprints cerrados | TestPantallasAnteriores_Historial, TestConsultarSprintsAnteriores |
| CA-012-3 | alternativo | El sprint abierto no aparece en el historial ni cuenta para la velocidad | TestPantallasAnteriores_SinSprintsCerrados, TestConsultarSprintsAnteriores |
| CA-012-4 | límite | Sin sprints cerrados no hay error: la lista está vacía y la velocidad es cero, Un sprint cerrado sin historias aparece con todo en cero | TestPantallasAnteriores_SinSprintsCerrados, TestConsultarSprintsAnteriores_CasosLimite |
| CA-012-5 | alternativo | Una historia que no se completó en un sprint y se completó después en otro sigue figurando como no completada en el primero | TestPantallasAnteriores_ElResultadoEsElDeLaFila, TestConsultarSprintsAnteriores_ElResultadoEsElDeLaFila |
| CA-012-6 | error | Un proyecto inexistente se rechaza | TestConsultarSprintsAnteriores_ProyectoInexistente |

### US-013 — Estimar una historia con Story Points en escala Fibonacci

- **Spec:** [specs/US-013-estimar-historia-fibonacci.md](../specs/US-013-estimar-historia-fibonacci.md) · **Issue:** #21
- **Escenarios BDD:** [features/US-013-estimar-historia-fibonacci.feature](../features/US-013-estimar-historia-fibonacci.feature)
- **Código:** `internal/domain/backlog/`, `internal/app/`, `internal/adapters/postgres/`
- **Tests que la citan:** 31

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-013-1 | normal | Una historia sin estimar queda con los story points elegidos | TestEstimarItem_Estima, TestNuevosStoryPoints_ValoresDeLaEscala, TestEstimar |
| CA-013-2 | alternativo | Se puede volver a estimar y la última estimación reemplaza a la anterior | TestEstimarItem_ReEstima |
| CA-013-3 | límite | La escala va de 1 a 13 y no incluye el 0 ni el 21 | — |
| CA-013-4 | error | Un valor fuera de la escala se rechaza y la historia conserva su estimación, Un valor fuera de la escala no deja la historia estimada | TestEstimarItem_Rechazos, TestNuevosStoryPoints_FueraDeLaEscala, TestEstimar_FueraDeEscala |
| CA-013-5 | error | — | TestEstimar_ItemCompletado |
| CA-013-6 | error | Estimar una historia que no existe se rechaza | TestBase_GuardarEstimacion_ItemInexistente, TestObtenerItem_Inexistente, TestGuardarEstimacion_ItemInexistente, TestEstimarItem_ItemInexistente |

### US-014 — Iniciar una sesión de Planning Poker sobre una historia

- **Spec:** [specs/US-014-iniciar-sesion-planning-poker.md](../specs/US-014-iniciar-sesion-planning-poker.md) · **Issue:** #25
- **Escenarios BDD:** [features/US-014-iniciar-sesion-planning-poker.feature](../features/US-014-iniciar-sesion-planning-poker.feature)
- **Código:** `internal/domain/planningpoker/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/00009_sesiones_planning_poker.sql`
- **Tests que la citan:** 26

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-014-1 | normal | La sesión queda votando, en la ronda 1, con todos los integrantes | TestBase_RegistrarSesion, TestRegistrarSesion, TestIniciarSesionPlanningPoker_Inicia, TestNuevaSesion |
| CA-014-2 | alternativo | Una historia ya estimada se vuelve a estimar en una sesión | TestIniciarSesionPlanningPoker_HistoriaEstimadaOEnProgreso |
| CA-014-3 | alternativo | Un integrante que se suma después de iniciar no participa | TestBase_RegistrarSesion_ParticipantesAlIniciar, TestRegistrarSesion_ParticipantesAlIniciar |
| CA-014-4 | límite | Con dos integrantes la sesión se inicia, Con un solo integrante la sesión se rechaza | TestIniciarSesionPlanningPoker_ReglasDelDominio, TestNuevaSesion_MinimoDeParticipantes |
| CA-014-5 | error | Una historia completada no se pone en sesión | TestIniciarSesionPlanningPoker_ReglasDelDominio, TestNuevaSesion_Rechazos |
| CA-014-6 | error | Una historia con una sesión abierta no admite otra | TestIniciarSesionPlanningPoker_ReglasDelDominio, TestNuevaSesion_Rechazos |
| CA-014-7 | error | Una historia que no existe se rechaza, Un proyecto que no existe se rechaza, sin mirar la historia | TestBase_RegistrarSesion_HistoriaInexistente, TestRegistrarSesion_HistoriaInexistente, TestIniciarSesionPlanningPoker_LoQueNoExiste |

### US-015 — Registrar el voto individual manteniéndolo oculto hasta el cierre

- **Spec:** [specs/US-015-voto-individual-oculto.md](../specs/US-015-voto-individual-oculto.md) · **Issue:** #26
- **Escenarios BDD:** [features/US-015-voto-individual-oculto.feature](../features/US-015-voto-individual-oculto.feature)
- **Código:** `internal/domain/planningpoker/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/00010_votos_planning_poker.sql`
- **Tests que la citan:** 32

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-015-1 | normal | Un participante vota y queda registrado que votó, sin que se vea el valor | TestBase_RegistrarVoto, TestRegistrarVoto, TestVotarEnSesion_Vota, TestVerificarVoto_Valido, +1 más |
| CA-015-2 | alternativo | Un participante cambia su voto y figura una sola vez | TestBase_RegistrarVoto_CambiarElVoto, TestRegistrarVoto_CambiarElVoto |
| CA-015-3 | alternativo | Aunque hayan votado todos, los valores siguen ocultos | TestVotacion_VotaronYFaltan, TestVotacion_NoHayFormaDeLeerLosValoresMientrasSeVota |
| CA-015-4 | límite | Solo se vota con un valor de la escala, Un valor fuera de la escala no cuenta como voto | TestVotarEnSesion_ReglasDelDominio, TestVerificarVoto_Valido, TestVerificarVoto_Rechazos |
| CA-015-5 | error | Un integrante que se sumó después de iniciar no vota | TestVotarEnSesion_ReglasDelDominio, TestVerificarVoto_Rechazos |
| CA-015-6 | error | Con la sesión revelada ya no se vota | TestVotarEnSesion_ReglasDelDominio, TestVerificarVoto_Rechazos |
| CA-015-7 | error | Votar en una historia sin sesión se rechaza, Un integrante que no existe no vota | TestBase_RegistrarVoto_SesionInexistente, TestRegistrarVoto_SesionInexistente, TestVotarEnSesion_LoQueNoExiste |

### US-016 — Revelar las estimaciones y detectar divergencias entre integrantes

- **Spec:** [specs/US-016-revelar-estimaciones.md](../specs/US-016-revelar-estimaciones.md) · **Issue:** #27
- **Escenarios BDD:** [features/US-016-revelar-estimaciones.feature](../features/US-016-revelar-estimaciones.feature)
- **Código:** `internal/domain/planningpoker/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`
- **Tests que la citan:** 25

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-016-1 | normal | Si todos votaron lo mismo, se revelan los votos y no hay divergencia | TestBase_GuardarRevelado, TestGuardarRevelado, TestRevelarVotos_Revela, TestRevelar |
| CA-016-2 | alternativo | Si votaron valores distintos, se revelan los votos y hay divergencia | TestBase_GuardarRevelado, TestRevelar |
| CA-016-3 | alternativo | Se puede revelar aunque falten votos, y se sabe quiénes no votaron, Un participante que cambió su voto antes de revelar figura una sola vez, con el último | TestBase_GuardarRevelado, TestGuardarRevelado, TestRevelarVotos_Revela, TestRevelar |
| CA-016-4 | límite | Valores contiguos de la escala también son una divergencia, Con un único voto emitido se revela y no hay divergencia | TestRevelar |
| CA-016-5 | error | Sin ningún voto no se revela y la sesión sigue votando | TestBase_GuardarRevelado_Rechazos, TestGuardarRevelado_Reglas, TestRevelarVotos_ReglasDelDominio, TestRevelar_Rechazos |
| CA-016-6 | error | Un integrante que no participa de la sesión no revela | TestBase_GuardarRevelado_Rechazos, TestGuardarRevelado_Reglas, TestRevelarVotos_ReglasDelDominio, TestRevelar_Rechazos |
| CA-016-7 | error | Una sesión ya revelada no se revela de nuevo | TestBase_GuardarRevelado_Rechazos, TestBase_GuardarRevelado_DespuesNoSeVota, TestGuardarRevelado_Reglas, TestGuardarRevelado_DespuesNoSeVota, +2 más |
| CA-016-8 | error | Revelar en una historia sin sesión se rechaza, Un integrante que no existe no revela | TestRevelarVotos_LoQueNoExiste |

### US-017 — Abrir una nueva ronda de votación cuando hay divergencia

- **Spec:** [specs/US-017-nueva-ronda.md](../specs/US-017-nueva-ronda.md) · **Issue:** #28
- **Escenarios BDD:** [features/US-017-nueva-ronda.feature](../features/US-017-nueva-ronda.feature)
- **Código:** `internal/domain/planningpoker/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`
- **Tests que la citan:** 24

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-017-1 | normal | Con divergencia, la nueva ronda deja la sesión votando en la ronda siguiente sin votos, Valores contiguos de la escala también abren una nueva ronda | TestBase_GuardarNuevaRonda, TestGuardarNuevaRonda, TestAbrirNuevaRonda_Abre, TestAbrirNuevaRonda |
| CA-017-2 | alternativo | Los votos de la ronda anterior se conservan y se ven, y los de la ronda nueva quedan ocultos, Revelar la ronda nueva muestra solo sus votos | TestBase_GuardarNuevaRonda, TestGuardarNuevaRonda, TestVotacionDeRonda |
| CA-017-3 | límite | Si todos votaron lo mismo no se abre una nueva ronda, Con un único voto emitido no hay divergencia y no se abre una nueva ronda | TestBase_GuardarNuevaRonda_Reglas, TestGuardarNuevaRonda_Reglas, TestAbrirNuevaRonda_ReglasDelDominio, TestAbrirNuevaRonda_Rechazos |
| CA-017-4 | error | Con la sesión votando no se abre una nueva ronda, Un integrante que no participa de la sesión no abre una nueva ronda, Abrir una nueva ronda en una historia sin sesión se rechaza, +1 más | TestBase_GuardarNuevaRonda_Reglas, TestGuardarNuevaRonda_Reglas, TestAbrirNuevaRonda_LoQueNoExiste, TestAbrirNuevaRonda_ReglasDelDominio, +1 más |
| CA-017-5 | error | — | TestBase_GuardarNuevaRonda_Reglas, TestGuardarNuevaRonda_DosPedidos |

### US-018 — Registrar la estimación acordada como valor final de la historia

- **Spec:** [specs/US-018-registrar-estimacion-acordada.md](../specs/US-018-registrar-estimacion-acordada.md) · **Issue:** #29
- **Escenarios BDD:** [features/US-018-registrar-estimacion-acordada.feature](../features/US-018-registrar-estimacion-acordada.feature)
- **Código:** `internal/domain/planningpoker/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00011_sesiones_valor_acordado.sql`
- **Tests que la citan:** 23

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-018-1 | normal | La sesión queda finalizada con el valor acordado y la historia, estimada | TestBase_GuardarAcuerdo, TestGuardarAcuerdo, TestRegistrarEstimacionAcordada_Registra, TestAcordar |
| CA-018-2 | alternativo | Se acuerda un valor que nadie votó, aunque haya divergencia | TestBase_GuardarAcuerdo, TestGuardarAcuerdo, TestAcordar |
| CA-018-3 | alternativo | Una historia ya estimada queda con el valor acordado | TestBase_GuardarAcuerdo_HistoriaYaEstimada, TestGuardarAcuerdo_HistoriaYaEstimada |
| CA-018-4 | límite | Solo se acuerda un valor de la escala, Un valor fuera de la escala deja la sesión revelada y la historia sin estimar | TestRegistrarEstimacionAcordada_ReglasDelDominio, TestAcordar_Rechazos |
| CA-018-5 | error | Con la sesión votando no se registra | TestAcordar_Rechazos |
| CA-018-6 | error | Un integrante que no participa de la sesión no registra | TestAcordar_Rechazos |
| CA-018-7 | error | Si la historia se completó, no se registra y la sesión sigue revelada | TestBase_GuardarAcuerdo_HistoriaCompletada, TestRegistrarEstimacionAcordada_ReglasDelDominio, TestAcordar_HistoriaCompletada |
| CA-018-8 | error | Registrar en una historia sin sesión se rechaza, Un integrante que no existe no registra | TestBase_GuardarAcuerdo_SesionInexistente, TestGuardarAcuerdo_SesionInexistente, TestRegistrarEstimacionAcordada_LoQueNoExiste |

### US-019 — Registrar esfuerzo con integrante, fecha, actividad y horas

- **Spec:** [specs/US-019-registrar-esfuerzo.md](../specs/US-019-registrar-esfuerzo.md) · **Issue:** #22
- **Escenarios BDD:** [features/US-019-registrar-esfuerzo.feature](../features/US-019-registrar-esfuerzo.feature)
- **Código:** `internal/domain/esfuerzo/`, `internal/app/`, `internal/adapters/postgres/`, `migraciones/`
- **Tests que la citan:** 34

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

### US-020 — Consultar el esfuerzo registrado por historia, integrante y sprint

- **Spec:** [specs/US-020-consultar-esfuerzo.md](../specs/US-020-consultar-esfuerzo.md) · **Issue:** #23
- **Escenarios BDD:** [features/US-020-consultar-esfuerzo.feature](../features/US-020-consultar-esfuerzo.feature)
- **Código:** `internal/domain/esfuerzo/`, `internal/app/consultar_esfuerzo.go`
- **Tests que la citan:** 29

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-020-1 | normal | Sin filtros se devuelven todos los registros, del más reciente al más antiguo | TestConsultarEsfuerzo_SinFiltros |
| CA-020-2 | normal | Filtrar por una historia deja solo sus registros, Los filtros se combinan: solo los registros que cumplen todos | TestConsultarEsfuerzo_FiltroPorHistoria |
| CA-020-3 | normal | Filtrar por un integrante deja sus registros de todas las historias | TestConsultarEsfuerzo_FiltroPorIntegrante |
| CA-020-4 | normal | Filtrar por un sprint deja los registros cuya fecha cae en él, con el primer y el último día incluidos, Las horas de una historia se reparten entre los sprints según la fecha del trabajo | TestConsultarEsfuerzo_FiltroPorSprint, TestSprintDeLaFecha_DentroDelRango |
| CA-020-5 | alternativo | Los totales por historia, por integrante y por sprint suman el total | TestConsultarEsfuerzo_TotalesPorGrupo |
| CA-020-6 | límite | Un registro cuya fecha no cae en ningún sprint se cuenta aparte | TestConsultarEsfuerzo_CadaRegistroLlevaSuDetalle |
| CA-020-7 | límite | Un proyecto sin registros da una consulta vacía y no es un error, Filtros sin coincidencias dan una consulta vacía y no es un error | TestConsultarEsfuerzo_SinRegistros |
| CA-020-8 | error | Una historia de otro proyecto como filtro se rechaza, Un integrante de otro proyecto como filtro se rechaza, Un sprint que no es del proyecto como filtro se rechaza | TestConsultarEsfuerzo_Inexistentes |
| CA-020-9 | error | Un proyecto inexistente se rechaza | TestConsultarEsfuerzo_Inexistentes |

### US-021 — Comparar el esfuerzo estimado contra el esfuerzo real

- **Spec:** [specs/US-021-comparar-esfuerzo.md](../specs/US-021-comparar-esfuerzo.md) · **Issue:** #24
- **Escenarios BDD:** [features/US-021-comparar-esfuerzo.feature](../features/US-021-comparar-esfuerzo.feature)
- **Código:** `internal/domain/metricas/`, `internal/app/comparar_esfuerzo.go`
- **Tests que la citan:** 23

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-021-1 | normal | Una historia completada se compara contra las horas que se registraron | TestCompararHistorias_UnaHistoria |
| CA-021-2 | normal | Se trabajó menos de lo estimado | TestCompararHistorias_UnaHistoria |
| CA-021-3 | alternativo | La historia donde más nos equivocamos sale primero y el total suma las dos | TestCompararEsfuerzo, TestCompararHistorias_Orden |
| CA-021-4 | límite | Una historia completada y estimada sin horas aparece con la desviación de sus estimadas | TestCompararHistorias_UnaHistoria |
| CA-021-5 | límite | Una historia sin estimar con horas aparece sin porcentaje y no suma al total | TestCompararHistorias_UnaHistoria, TestCompararHistorias_SoloLasCompletadasYEstimadasSumanAlTotal |
| CA-021-6 | límite | Una historia pendiente sin horas no aparece y una en curso con horas no suma al total | TestCompararEsfuerzo, TestCompararHistorias_QueAparece, TestCompararHistorias_SoloLasCompletadasYEstimadasSumanAlTotal |
| CA-021-7 | límite | Un proyecto sin historias da una comparación vacía y no es un error | TestCompararEsfuerzo_SinHistorias, TestCompararHistorias_SinHistorias |
| CA-021-8 | error | Un proyecto inexistente se rechaza | TestCompararEsfuerzo_ProyectoInexistente |
| CA-021-9 | error | — | TestCompararEsfuerzo_DatosInvalidos, TestCompararHistorias_DatosInvalidos |

### US-022 — Registrar un defecto con severidad, historia relacionada y sprint de detección

- **Spec:** [specs/US-022-registrar-defecto.md](../specs/US-022-registrar-defecto.md) · **Issue:** #30
- **Escenarios BDD:** [features/US-022-registrar-defecto.feature](../features/US-022-registrar-defecto.feature)
- **Código:** `internal/domain/defecto/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00012_defectos.sql`
- **Tests que la citan:** 35

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-022-1 | normal | Se registra un defecto abierto, sin sprint de resolución | TestBase_RegistrarDefecto, TestRegistrarDefecto_Guarda, TestRegistrarDefecto_Registra, TestNuevo_NaceAbiertoYSinResolver |
| CA-022-2 | alternativo | Se pueden registrar defectos de las cuatro severidades | TestRegistrarDefecto_CuatroSeveridades, TestNuevo_CuatroSeveridades |
| CA-022-3 | alternativo | El sprint de detección puede ser distinto del sprint de la historia, El sprint de detección puede estar cerrado | TestRegistrarDefecto_GuardaElIDDelSprintDeDeteccion |
| CA-022-4 | alternativo | Se puede registrar un defecto sobre una historia completada y la historia no cambia, Se puede registrar un defecto sobre una historia que no está en ningún sprint | TestRegistrarDefecto_SobreHistoriaEnCualquierEstado |
| CA-022-5 | alternativo | Se pueden registrar varios defectos sobre la misma historia, incluso idénticos | TestRegistrarDefecto_Repetidos, TestRegistrarDefecto_RepetidosSonValidos |
| CA-022-6 | límite | La descripción admite hasta 2000 caracteres, Una descripción de más de 2000 caracteres se rechaza y no se registra nada | TestNuevo_LargoDeLaDescripcion, TestNuevo_ElLargoSeMideRecortado |
| CA-022-7 | error | Una descripción vacía se rechaza y no se registra nada | TestRegistrarDefecto_DatosInvalidos, TestNuevo_DescripcionVacia |
| CA-022-8 | error | Una severidad fuera de critica, alta, media y baja se rechaza | TestRegistrarDefecto_DatosInvalidos, TestNuevo_SeveridadInvalida |
| CA-022-9 | error | No se registra un defecto sobre una historia que no existe, No se registra un defecto sobre una historia de otro proyecto | TestRegistrarDefecto_LoQueDesaparece, TestRegistrarDefecto_LoQueNoExiste |
| CA-022-10 | error | No se registra un defecto detectado en un sprint que no existe, No se registra un defecto detectado en un sprint de otro proyecto | TestRegistrarDefecto_LoQueNoExiste |
| CA-022-11 | error | Registrar un defecto en un proyecto que no existe se rechaza | TestRegistrarDefecto_LoQueDesaparece, TestRegistrarDefecto_LoQueNoExiste |
| CA-022-12 | error | Con la descripción vacía y la severidad inválida se informan los dos errores | TestRegistrarDefecto_DatosInvalidos, TestNuevo_TodosLosErroresJuntos |

### US-023 — Cambiar el estado de un defecto y registrar su sprint de resolución

- **Spec:** [specs/US-023-cambiar-estado-defecto.md](../specs/US-023-cambiar-estado-defecto.md) · **Issue:** #31
- **Escenarios BDD:** [features/US-023-cambiar-estado-defecto.feature](../features/US-023-cambiar-estado-defecto.feature)
- **Código:** `internal/domain/defecto/`, `internal/app/`, `internal/adapters/postgres/`, `internal/adapters/memoria/`, `migraciones/00013_defectos_estado.sql`
- **Tests que la citan:** 27

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-023-1 | normal | Un defecto abierto se resuelve en un sprint posterior al de detección | TestBase_ActualizarEstadoDefecto_Guarda, TestActualizarEstadoDefecto_Guarda, TestCambiarEstadoDefecto_Resuelve |
| CA-023-2 | alternativo | Un defecto abierto pasa a en_correccion y queda sin sprint de resolución, Un defecto en corrección se resuelve en un sprint, Un defecto abierto se descarta y queda sin sprint de resolución | TestCambiarEstadoDefecto_SoloCambiaEstadoYSprintDeResolucion, TestCambiarEstado_TransicionesPermitidas |
| CA-023-3 | alternativo | Un defecto resuelto se reabre y queda sin sprint de resolución, Un defecto descartado se reabre y queda sin sprint de resolución, Un defecto reabierto se puede resolver de nuevo en otro sprint | TestActualizarEstadoDefecto_Guarda, TestCambiarEstado_TransicionesPermitidas |
| CA-023-4 | alternativo | Un defecto de una historia completada se resuelve en un sprint cerrado y la historia no cambia | TestCambiarEstadoDefecto_SoloCambiaEstadoYSprintDeResolucion |
| CA-023-5 | límite | Un sprint de resolución igual al de detección se acepta, Un sprint de resolución anterior al de detección se rechaza | TestCambiarEstado_SprintDeResolucionNoAnteriorAlDeDeteccion |
| CA-023-6 | error | Una transición no permitida se rechaza y el defecto no cambia, Pedir el estado actual se rechaza y no es un cambio nulo | TestCambiarEstado_TransicionesProhibidas |
| CA-023-7 | error | Resolver sin sprint de resolución se rechaza, Informar un sprint de resolución con un estado distinto de resuelto se rechaza | TestCambiarEstado_SprintDeResolucionSegunEstado |
| CA-023-8 | error | Un sprint de resolución anterior al de detección indica los dos números | TestCambiarEstado_SprintDeResolucionNoAnteriorAlDeDeteccion |
| CA-023-9 | error | Un estado fuera de los cuatro se rechaza | TestCambiarEstado_EstadoInvalido |
| CA-023-10 | error | No se cambia el estado de un defecto que no existe | TestBase_DefectoEnProyecto, TestRepositorioDefectos_DefectoEnProyecto, TestCambiarEstadoDefecto_LoQueNoExiste |
| CA-023-11 | error | No se resuelve un defecto en un sprint que no existe, No se resuelve un defecto en un sprint de otro proyecto | TestBase_NumeroDeSprintEnProyecto, TestRepositorioDefectos_NumeroDeSprintEnProyecto, TestCambiarEstadoDefecto_LoQueNoExiste |
| CA-023-12 | error | Cambiar el estado de un defecto en un proyecto que no existe se rechaza | TestCambiarEstadoDefecto_LoQueNoExiste |
| CA-023-13 | error | Con una transición prohibida y sin sprint de resolución se informan los dos errores | TestCambiarEstadoDefecto_DatosInvalidos, TestCambiarEstado_UneLosErroresEnOrden |
| CA-023-14 | error | — | TestBase_ActualizarEstadoDefecto_RechazaElEstadoQueYaNoEsElGuardado, TestDefectos_LosCheckDelEstado, TestActualizarEstadoDefecto_RechazaUnEstadoAnteriorQueYaNoEsElGuardado, TestActualizarEstadoDefecto_CambiosSimultaneos, +1 más |

### US-024 — Listar los defectos de un proyecto filtrando por sprint de detección y severidad

- **Spec:** [specs/US-024-listar-defectos.md](../specs/US-024-listar-defectos.md) · **Issue:** #32 (historia) · issue de la spec: pendiente
- **Escenarios BDD:** ninguno
- **Código:** `internal/app/`, `internal/domain/defecto/`, `internal/adapters/http/`, `web/templates/`
- **Tests que la citan:** 0

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-024-1 | normal | — | — |
| CA-024-2 | alternativo | — | — |
| CA-024-3 | límite | — | — |
| CA-024-4 | error | — | — |
| CA-024-5 | alternativo | — | — |
| CA-024-6 | error | — | — |

### US-025 — Calcular Story Points planificados y completados de un Sprint

- **Spec:** [specs/US-025-story-points-sprint.md](../specs/US-025-story-points-sprint.md) · **Issue:** #11
- **Escenarios BDD:** [features/US-025-story-points-sprint.feature](../features/US-025-story-points-sprint.feature)
- **Código:** `internal/domain/metricas/`
- **Tests que la citan:** 3

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
- **Escenarios BDD:** [features/US-027-horas-estimadas-reales.feature](../features/US-027-horas-estimadas-reales.feature)
- **Código:** `internal/domain/metricas/`
- **Tests que la citan:** 4

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-027-1 | normal | Se trabajó un poco más de lo estimado | TestCalcularHoras |
| CA-027-2 | alternativo | Se trabajó menos de lo estimado | TestCalcularHoras |
| CA-027-3 | límite | Sin horas registradas la desviación es la de las estimadas en negativo | TestCalcularHoras |
| CA-027-4 | límite | Sin story points planificados no se puede calcular el porcentaje | TestCalcularHoras |
| CA-027-5 | error | Una hora en cero se rechaza indicando su posición, Una hora negativa se rechaza | TestCalcularHoras_Rechazos |
| CA-027-6 | error | Story points planificados negativos se rechazan | TestCalcularHoras_Rechazos |

### US-028 — Calcular el porcentaje de historias completadas

- **Spec:** [specs/US-028-porcentaje-historias-completadas.md](../specs/US-028-porcentaje-historias-completadas.md) · **Issue:** #34
- **Escenarios BDD:** [features/US-028-porcentaje-historias-completadas.feature](../features/US-028-porcentaje-historias-completadas.feature)
- **Código:** `internal/domain/metricas/`
- **Tests que la citan:** 4

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-028-1 | normal | El porcentaje es la parte de las historias que están completadas | — |
| CA-028-2 | alternativo | Se cuentan historias y no story points | — |
| CA-028-3 | límite | Un sprint sin historias da cero y no es un error | — |
| CA-028-4 | límite | El porcentaje se redondea a un decimal | — |
| CA-028-5 | error | Una historia con story points negativos se rechaza | TestCalcularPorcentajeCompletadas_StoryPointsNegativos |

### US-029 — Calcular defectos detectados y resueltos por sprint

- **Spec:** [specs/US-029-defectos-por-sprint.md](../specs/US-029-defectos-por-sprint.md) · **Issue:** #35
- **Escenarios BDD:** [features/US-029-defectos-por-sprint.feature](../features/US-029-defectos-por-sprint.feature)
- **Código:** `internal/domain/metricas/`
- **Tests que la citan:** 5

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-029-1 | normal | Se cuentan los detectados y los resueltos del sprint | TestCalcularDefectosDelSprint |
| CA-029-2 | alternativo | Un defecto se cuenta como detectado donde se detectó y como resuelto donde se resolvió | TestCalcularDefectosDelSprint |
| CA-029-3 | alternativo | Un defecto detectado y resuelto en el mismo sprint cuenta en las dos cifras | TestCalcularDefectosDelSprint |
| CA-029-4 | límite | Un sprint sin defectos da cero y no es un error, Un defecto sin resolver cuenta como detectado y nunca como resuelto | TestCalcularDefectosDelSprint |
| CA-029-5 | error | Un defecto resuelto antes de detectarse se rechaza | TestCalcularDefectosDelSprint_Rechazos |
| CA-029-6 | error | Un número de sprint menor que 1 se rechaza | TestCalcularDefectosDelSprint_Rechazos |

### US-030 — Ver un Dashboard con el estado del proyecto y sus métricas principales

- **Spec:** [specs/US-030-dashboard.md](../specs/US-030-dashboard.md) · **Issue:** #36
- **Escenarios BDD:** ninguno
- **Código:** `internal/app/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/`
- **Tests que la citan:** 0

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-030-1 | normal | — | — |
| CA-030-2 | normal | — | — |
| CA-030-3 | alternativo | — | — |
| CA-030-4 | límite | — | — |
| CA-030-5 | error | — | — |
| CA-030-6 | normal | — | — |
| CA-030-7 | límite | — | — |

### US-033 — Generar en pantalla el reporte completo de un proyecto o Sprint

- **Spec:** [specs/US-033-reporte-pantalla.md](../specs/US-033-reporte-pantalla.md) · **Issue:** #39
- **Escenarios BDD:** ninguno
- **Código:** `internal/app/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/`
- **Tests que la citan:** 0

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-033-1 | normal | — | — |
| CA-033-2 | alternativo | — | — |
| CA-033-3 | límite | — | — |
| CA-033-4 | error | — | — |
| CA-033-5 | plataforma | — | — |

### US-034 — Imprimir el reporte o guardarlo como PDF desde el navegador

- **Spec:** [specs/US-034-vista-imprimible.md](../specs/US-034-vista-imprimible.md) · **Issue:** #40
- **Escenarios BDD:** ninguno
- **Código:** `web/static/css/input.css`, `web/static/js/`, `web/templates/`, `internal/adapters/http/`
- **Tests que la citan:** 0

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-034-1 | normal | — | — |
| CA-034-2 | alternativo | — | — |
| CA-034-3 | límite | — | — |
| CA-034-4 | plataforma | — | — |

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
- **Tests que la citan:** 40

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
- **Tests que la citan:** 27

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
- **Tests que la citan:** 22

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-T015-1 | normal | — | TestPantallasSprints_Recorrido |
| CA-T015-2 | alternativo | — | TestPantallasSprints_RedireccionesConYSinHTMX |
| CA-T015-3 | límite | — | TestPantallasSprints_SprintCerrado |
| CA-T015-4 | error | — | TestPantallasSprints_ErroresAlCrear |
| CA-T015-5 | plataforma | — | — |

### T-016 — Cerrar un sprint desde el navegador

- **Spec:** [specs/T-016-cerrar-sprint.md](../specs/T-016-cerrar-sprint.md) · **Issue:** #157 (tarea) · #158 (spec)
- **Escenarios BDD:** ninguno
- **Código:** `internal/adapters/http/`, `web/templates/`, `cmd/server/`
- **Tests que la citan:** 6

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-T016-1 | normal | — | TestPantallasCierre_Recorrido |
| CA-T016-2 | alternativo | — | TestPantallasCierre_Recorrido, TestPantallasCierre_ConHTMX |
| CA-T016-3 | límite | — | TestPantallasCierre_SprintSinHistorias |
| CA-T016-4 | error | — | TestPantallasCierre_Rechazos |
| CA-T016-5 | plataforma | — | — |

### T-019 — Pantallas de esfuerzo: cargar horas, consultar y comparar

- **Spec:** [specs/T-019-pantallas-esfuerzo.md](../specs/T-019-pantallas-esfuerzo.md) · **Issue:** #189 (tarea) · #193 (spec)
- **Escenarios BDD:** ninguno
- **Código:** `internal/adapters/http/`, `web/templates/`, `cmd/server/`
- **Tests que la citan:** 15

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-T019-1 | normal | — | TestPantallasEsfuerzo_Recorrido |
| CA-T019-2 | alternativo | — | TestPantallasEsfuerzo_Filtros |
| CA-T019-3 | alternativo | — | TestPantallasEsfuerzo_Comparacion |
| CA-T019-4 | límite | — | TestPantallasEsfuerzo_SinCoincidencias |
| CA-T019-5 | error | — | TestPantallasEsfuerzo_Rechazos |
| CA-T019-6 | plataforma | — | TestPantallasEsfuerzo_ConYSinHTMX |

### T-020 — Pantallas de Planning Poker: iniciar, votar, revelar y acordar

- **Spec:** [specs/T-020-pantallas-planning-poker.md](../specs/T-020-pantallas-planning-poker.md) · **Issue:** #190
- **Escenarios BDD:** ninguno
- **Código:** `internal/adapters/http/`, `web/templates/`, `internal/app/`, `cmd/server/`
- **Tests que la citan:** 12

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-T020-1 | normal | — | TestPantallasPoker_Recorrido |
| CA-T020-2 | normal | — | TestConsultarSesionPlanningPoker_Votando |
| CA-T020-3 | alternativo | — | TestConsultarSesionPlanningPoker_Revelada |
| CA-T020-4 | alternativo | — | TestPantallasPoker_Recorrido |
| CA-T020-5 | alternativo | — | TestPantallasPoker_NuevaRonda, TestConsultarSesionPlanningPoker_Votando |
| CA-T020-6 | límite | — | TestPantallasPoker_SesionAbiertaYCompletada |
| CA-T020-7 | error | — | TestPantallasPoker_Errores |

### T-021 — Pantallas de defectos: registrar, listar y cambiar el estado

- **Spec:** [specs/T-021-pantallas-defectos.md](../specs/T-021-pantallas-defectos.md) · **Issue:** #191 (tarea) · issue de la spec: pendiente
- **Escenarios BDD:** ninguno
- **Código:** `internal/app/`, `internal/adapters/memoria/`, `internal/adapters/postgres/`, `internal/adapters/http/`, `web/templates/`, `cmd/server/`
- **Tests que la citan:** 37

| Criterio | Tipo | Escenarios BDD | Tests |
|---|---|---|---|
| CA-T021-1 | normal | — | TestPantallasDefectos_Registrar, TestPantallasDefectos_RegistrarConYSinHTMX |
| CA-T021-2 | alternativo | — | TestPantallasDefectos_Lista |
| CA-T021-3 | alternativo | — | TestPantallasDefectos_BotonesSegunElEstado, TestPantallasDefectos_CambiarElEstado |
| CA-T021-4 | límite | — | TestPantallasDefectos_ListasDelFormulario, TestPantallasDefectos_SinHistoriasOSinSprints, TestPantallasDefectos_SprintCerradoYRepetidos |
| CA-T021-5 | error | — | TestPantallasDefectos_ErroresAlRegistrar, TestPantallasDefectos_ResolverEnUnSprintAnterior, TestPantallasDefectos_PaginaDesactualizada, TestPantallasDefectos_EstadoMandadoAMano |
| CA-T021-6 | plataforma | — | TestPantallasDefectos_NavegacionYFragmento, TestBase_ListarDefectosDelProyecto, TestBase_ListarDefectosDelProyecto_CreadoEnCreciente, TestRepositorioDefectos_ListarDefectosDelProyecto, +1 más |
