# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# La Definition of Ready (RN-009-7) pide una historia estimada y con al menos un criterio
# de aceptacion: por eso las historias que se asignan con exito usan la frase con criterios.
# CA-009-6 (asignar a un sprint cerrado) no tiene escenario todavia: necesita un sprint
# cerrado y cerrar es US-011. Lo cubren los tests del dominio y de Postgres.
@US-009
Característica: Asignar historias del backlog a un Sprint
  Como integrante del equipo
  quiero asignar historias del backlog a un Sprint
  para planificar qué trabajo se hace en ese Sprint

  Especificación: specs/US-009-asignar-historias-sprint.md

  # ---------------------------------------------------------------- caso normal
  @CA-009-1
  Escenario: Una historia estimada y con criterios queda en el sprint
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Cuando se asigna la historia "Alta de proyectos" al sprint "Sprint 1"
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" está en el sprint "Sprint 1"

  # ---------------------------------------------------------- casos alternativos
  @CA-009-2
  Escenario: Se pueden asignar varias historias al mismo sprint
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Listar proyectos" con prioridad "should", 3 story points y 1 criterio de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Cuando se asigna la historia "Listar proyectos" al sprint "Sprint 1"
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" está en el sprint "Sprint 1"
    Y la historia "Listar proyectos" está en el sprint "Sprint 1"

  # ---------------------------------------------------------------- casos limite
  @CA-009-3 @limite
  Escenario: Una historia con un solo criterio y un solo story point cumple la Definition of Ready
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "could", 1 story points y 1 criterio de aceptación
    Cuando se asigna la historia "Alta de proyectos" al sprint "Sprint 1"
    Entonces la operación es exitosa

  # ------------------------------------------------------------------- errores
  @CA-009-4 @error
  Escenario: Una historia sin estimar se rechaza y sigue en el backlog
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando se asigna la historia "Alta de proyectos" al sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la historia no esta estimada"
    Y la historia "Alta de proyectos" está en el backlog

  @CA-009-5 @error
  Escenario: Una historia sin criterios de aceptación se rechaza y sigue en el backlog
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must" y 5 story points
    Cuando se asigna la historia "Alta de proyectos" al sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la historia no tiene criterios de aceptacion"
    Y la historia "Alta de proyectos" está en el backlog

  @CA-009-5 @error
  Escenario: Una historia sin estimar y sin criterios informa las dos faltas juntas
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando se asigna la historia "Alta de proyectos" al sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la historia no esta estimada"
    Y el mensaje de error indica "la historia no tiene criterios de aceptacion"

  @CA-009-7 @error
  Escenario: Una historia que ya está en un sprint no se asigna otra vez
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Cuando se asigna la historia "Alta de proyectos" al sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la historia ya esta asignada a un sprint"

  @CA-009-8 @error
  Escenario: Asignar a un sprint que no existe se rechaza
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Cuando se asigna la historia "Alta de proyectos" al sprint "Sprint 9"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint no existe"
    Y la historia "Alta de proyectos" está en el backlog

  @CA-009-8 @error
  Escenario: Asignar una historia que no existe se rechaza
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se asigna la historia "Historia fantasma" al sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el item no existe"

  @CA-009-8 @error
  Escenario: Una historia de otro proyecto no se asigna a un sprint de este
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y un proyecto "Otro"
    Y un sprint "Sprint 1" con el objetivo "Objetivo del otro proyecto"
    Cuando se asigna la historia "Alta de proyectos" al sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la historia y el sprint son de proyectos distintos"
