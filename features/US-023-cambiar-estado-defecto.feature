# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Las historias que se asignan usan la frase con criterios, por la Definition of Ready de US-009.
# CA-023-14 (cambios simultaneos) no tiene escenario: exige dos cambios a la vez y un escenario
# tiene un solo "Cuando". Lo cubren los tests del repositorio y del caso de uso (spec US-023, CA-023-14).
# Tampoco hay escenario de un defecto "de otro proyecto" (CL-023-10): el area Defectos busca el
# defecto por su historia y los sprints en el proyecto de esa historia. Lo prueban el caso de uso
# y el repositorio.
@US-023
Característica: Cambiar el estado de un defecto y registrar su sprint de resolución
  Como integrante del equipo
  quiero cambiar el estado de un defecto y, cuando se resuelve, registrar el sprint en el que se resolvió
  para seguir su ciclo de vida y medir cuántos defectos resolvemos en cada sprint

  Especificación: specs/US-023-cambiar-estado-defecto.md

  # ---------------------------------------------------------------- caso normal
  @CA-023-1
  Escenario: Un defecto abierto se resuelve en un sprint posterior al de detección
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 2"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" tiene estado "resuelto"
    Y el defecto de la historia "Crear proyecto" fue resuelto en el sprint "Sprint 2"
    Y el defecto de la historia "Crear proyecto" fue detectado en el sprint "Sprint 1"

  # ---------------------------------------------------------- casos alternativos
  @CA-023-2
  Escenario: Un defecto abierto pasa a en_correccion y queda sin sprint de resolución
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "en_correccion"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" tiene estado "en_correccion"
    Y el defecto de la historia "Crear proyecto" está sin resolver

  @CA-023-2
  Escenario: Un defecto en corrección se resuelve en un sprint
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y el defecto de la historia "Crear proyecto" en estado "en_correccion"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 1"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" tiene estado "resuelto"
    Y el defecto de la historia "Crear proyecto" fue resuelto en el sprint "Sprint 1"

  @CA-023-2
  Escenario: Un defecto abierto se descarta y queda sin sprint de resolución
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "baja" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "descartado"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" tiene estado "descartado"
    Y el defecto de la historia "Crear proyecto" está sin resolver

  @CA-023-3
  Escenario: Un defecto resuelto se reabre y queda sin sprint de resolución
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y el defecto de la historia "Crear proyecto" resuelto en el sprint "Sprint 1"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "abierto"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"
    Y el defecto de la historia "Crear proyecto" está sin resolver

  @CA-023-3
  Escenario: Un defecto descartado se reabre y queda sin sprint de resolución
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "baja" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y el defecto de la historia "Crear proyecto" en estado "descartado"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "abierto"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"
    Y el defecto de la historia "Crear proyecto" está sin resolver

  @CA-023-3
  Escenario: Un defecto reabierto se puede resolver de nuevo en otro sprint
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y el sprint "Sprint 2" cerrado
    Y un sprint "Sprint 3" con el objetivo "Reportes"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y el defecto de la historia "Crear proyecto" resuelto en el sprint "Sprint 2"
    Y el defecto de la historia "Crear proyecto" en estado "abierto"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 3"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" tiene estado "resuelto"
    Y el defecto de la historia "Crear proyecto" fue resuelto en el sprint "Sprint 3"

  @CA-023-4
  Escenario: Un defecto de una historia completada se resuelve en un sprint cerrado y la historia no cambia
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y la historia "Crear proyecto" asignada al sprint "Sprint 1"
    Y la historia "Crear proyecto" completada
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y el sprint "Sprint 2" cerrado
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 2"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" fue resuelto en el sprint "Sprint 2"
    Y la historia "Crear proyecto" tiene estado "completado"

  # ---------------------------------------------------------------- casos limite
  @CA-023-5 @limite
  Escenario: Un sprint de resolución igual al de detección se acepta
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 2"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 2"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" fue resuelto en el sprint "Sprint 2"

  @CA-023-5 @limite @error
  Escenario: Un sprint de resolución anterior al de detección se rechaza
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 2"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 1"
    Entonces la operación se rechaza
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"
    Y el defecto de la historia "Crear proyecto" está sin resolver

  # ------------------------------------------------------------------- errores
  @CA-023-6 @error
  Esquema del escenario: Una transición no permitida se rechaza y el defecto no cambia
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y el defecto de la historia "Crear proyecto" resuelto en el sprint "Sprint 1"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "<pedido>"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el defecto no puede pasar a ese estado desde el actual"
    Y el mensaje de error indica "de resuelto a <pedido>"
    Y el defecto de la historia "Crear proyecto" tiene estado "resuelto"
    Y el defecto de la historia "Crear proyecto" fue resuelto en el sprint "Sprint 1"

    Ejemplos:
      | pedido        |
      | en_correccion |
      | descartado    |

  @CA-023-6 @error
  Escenario: Pedir el estado actual se rechaza y no es un cambio nulo
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "abierto"
    Entonces la operación se rechaza
    Y el mensaje de error indica "de abierto a abierto"
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"

  @CA-023-7 @error
  Escenario: Resolver sin sprint de resolución se rechaza
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto"
    Entonces la operación se rechaza
    Y el mensaje de error indica "un defecto resuelto necesita un sprint de resolucion"
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"

  @CA-023-7 @error
  Escenario: Informar un sprint de resolución con un estado distinto de resuelto se rechaza
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "en_correccion" con el sprint de resolución "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "solo un defecto resuelto tiene sprint de resolucion"
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"
    Y el defecto de la historia "Crear proyecto" está sin resolver

  @CA-023-8 @error
  Escenario: Un sprint de resolución anterior al de detección indica los dos números
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 2"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint de resolucion no puede ser anterior al de deteccion"
    Y el mensaje de error indica "resolucion en el sprint 1, deteccion en el sprint 2"
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"

  @CA-023-9 @error
  Esquema del escenario: Un estado fuera de los cuatro se rechaza
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "<estado>"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el estado tiene que ser abierto, en_correccion, resuelto o descartado"
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"

    Ejemplos:
      | estado   |
      | cerrado  |
      | Resuelto |
      |          |

  @CA-023-10 @error
  Escenario: No se cambia el estado de un defecto que no existe
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "en_correccion"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el defecto no existe"
    Y el proyecto "Demo" tiene 0 defectos

  @CA-023-11 @error
  Escenario: No se resuelve un defecto en un sprint que no existe
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 9"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint no existe"
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"

  @CA-023-11 @error
  Escenario: No se resuelve un defecto en un sprint de otro proyecto
    Dado un proyecto "Otro"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 2"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint no existe"
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"

  @CA-023-12 @error
  Escenario: Cambiar el estado de un defecto en un proyecto que no existe se rechaza
    Dado ningún proyecto
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto" con el sprint de resolución "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"

  @CA-023-13 @error
  Escenario: Con una transición prohibida y sin sprint de resolución se informan los dos errores
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "baja" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y el defecto de la historia "Crear proyecto" en estado "descartado"
    Cuando se cambia el estado del defecto de la historia "Crear proyecto" a "resuelto"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el defecto no puede pasar a ese estado desde el actual"
    Y el mensaje de error indica "un defecto resuelto necesita un sprint de resolucion"
    Y el defecto de la historia "Crear proyecto" tiene estado "descartado"
