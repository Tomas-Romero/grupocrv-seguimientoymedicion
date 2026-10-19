# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Marcar y desmarcar operan sobre el ultimo sprint del proyecto de la historia, que es
# el unico que puede estar abierto (RN-008-17). Las historias que se asignan usan la
# frase con criterios, por la Definition of Ready de US-009.
# Sin escenario todavia:
# - CA-010-5 (con el sprint cerrado no se marca ni se desmarca) necesita un sprint
#   cerrado, y cerrar es US-011. Lo cubren los tests del dominio y de Postgres.
# - La fecha de completado (CA-010-1 y CA-010-3) no se verifica aca: el reloj de los
#   escenarios es fijo. La cubren los tests del dominio, la aplicacion y Postgres.
@US-010
Característica: Marcar historias como completadas dentro de un Sprint
  Como integrante del equipo
  quiero marcar como completadas las historias de un Sprint
  para registrar el avance del trabajo mientras el Sprint está en curso

  Especificación: specs/US-010-marcar-historias-completadas.md

  # ---------------------------------------------------------------- caso normal
  @CA-010-1
  Escenario: Una historia del sprint abierto queda completada
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Cuando se marca la historia "Alta de proyectos" como completada
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" tiene estado "completado"
    Y la historia "Alta de proyectos" está en el sprint "Sprint 1"

  # ---------------------------------------------------------- casos alternativos
  @CA-010-2
  Escenario: Una historia completada se desmarca y vuelve a pendiente
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Cuando se desmarca la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" tiene estado "pendiente"
    Y la historia "Alta de proyectos" está en el sprint "Sprint 1"

  # ---------------------------------------------------------------- casos limite
  @CA-010-3 @limite
  Escenario: Marcar una historia que ya está completada no es un error
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Cuando se marca la historia "Alta de proyectos" como completada
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" tiene estado "completado"

  # ------------------------------------------------------------------- errores
  @CA-010-4 @error
  Escenario: Una historia que no está en el sprint no se marca
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Cuando se marca la historia "Alta de proyectos" como completada
    Entonces la operación se rechaza
    Y el mensaje de error indica "la historia no esta asignada a este sprint"
    Y la historia "Alta de proyectos" tiene estado "pendiente"

  @CA-010-6 @error
  Escenario: Marcar una historia que no existe se rechaza
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se marca la historia "Historia fantasma" como completada
    Entonces la operación se rechaza
    Y el mensaje de error indica "el item no existe"

  @CA-010-6 @error
  Escenario: Marcar una historia de un proyecto sin sprints se rechaza y no cambia nada
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Cuando se marca la historia "Alta de proyectos" como completada
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint no existe"
    Y la historia "Alta de proyectos" tiene estado "pendiente"
