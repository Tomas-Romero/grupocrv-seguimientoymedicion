# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Las historias que se asignan usan la frase con criterios, por la Definition of Ready
# de US-009. Los escenarios que otras historias dejaron para esta (CA-008-4, CA-008-9,
# CA-009-6 y CA-010-5) estan en el .feature de cada historia.
@US-011
Característica: Cerrar un Sprint y devolver las historias no completadas al backlog
  Como integrante del equipo
  quiero cerrar un Sprint
  para dar por terminado el período, dejar registrado qué se completó y devolver al backlog lo que no

  Especificación: specs/US-011-cerrar-sprint.md

  # ---------------------------------------------------------------- caso normal
  @CA-011-1
  Escenario: Al cerrar, las completadas quedan completadas y las demás vuelven al backlog
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Listar proyectos" con prioridad "should", 3 story points y 1 criterio de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Listar proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Cuando se cierra el sprint "Sprint 1"
    Entonces la operación es exitosa
    Y el sprint "Sprint 1" está cerrado
    Y la historia "Alta de proyectos" tiene estado "completado"
    Y la historia "Listar proyectos" está en el backlog
    Y la historia "Listar proyectos" tiene estado "pendiente"

  # ---------------------------------------------------------- casos alternativos
  @CA-011-2
  Escenario: Después de cerrar se crea el sprint siguiente y una historia devuelta se asigna a él
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Listar proyectos" con prioridad "should", 3 story points y 1 criterio de aceptación
    Y la historia "Listar proyectos" asignada al sprint "Sprint 1"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Cuando se asigna la historia "Listar proyectos" al sprint "Sprint 2"
    Entonces la operación es exitosa
    Y la historia "Listar proyectos" está en el sprint "Sprint 2"

  @CA-011-3
  Escenario: Los planificados de un sprint cerrado incluyen las historias no completadas
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Listar proyectos" con prioridad "should", 3 story points y 1 criterio de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Listar proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Y el sprint "Sprint 1" cerrado
    Cuando se consultan los story points planificados del sprint "Sprint 1"
    Entonces los story points planificados son 8

  @CA-011-3
  Escenario: La velocidad se calcula con lo completado en cada sprint cerrado
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Listar proyectos" con prioridad "should", 3 story points y 1 criterio de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Listar proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y la historia "Listar proyectos" asignada al sprint "Sprint 2"
    Y la historia "Listar proyectos" completada
    Y el sprint "Sprint 2" cerrado
    Cuando se consulta la velocidad del equipo
    Entonces la velocidad del equipo es 4

  # ---------------------------------------------------------------- casos limite
  @CA-011-4 @limite
  Escenario: Un sprint sin historias se cierra
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se cierra el sprint "Sprint 1"
    Entonces la operación es exitosa
    Y el sprint "Sprint 1" está cerrado

  # ------------------------------------------------------------------- errores
  @CA-011-5 @error
  Escenario: Un sprint ya cerrado no se vuelve a cerrar
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Cuando se cierra el sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint esta cerrado"

  @CA-011-6 @error
  Escenario: Cerrar un sprint que no existe se rechaza
    Dado un proyecto "Demo"
    Cuando se cierra el sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint no existe"
