# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Las historias que se asignan usan la frase con criterios, por la Definition of Ready de US-009.
# La descripcion de solo espacios no tiene frase (spec US-022, seccion 5): la cubren los tests
# unitarios del dominio.
@US-022
Característica: Registrar un defecto con severidad, historia relacionada y sprint de detección
  Como integrante del equipo
  quiero registrar un defecto con su severidad, la historia donde se introdujo y el sprint en el que lo detectamos
  para llevar la cuenta de los defectos y medir cómo evoluciona la calidad del producto de un sprint a otro

  Especificación: specs/US-022-registrar-defecto.md

  # ---------------------------------------------------------------- caso normal
  @CA-022-1
  Escenario: Se registra un defecto abierto, sin sprint de resolución
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se registra un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación es exitosa
    Y el proyecto "Demo" tiene 1 defecto
    Y el defecto de la historia "Crear proyecto" tiene severidad "alta"
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"
    Y el defecto de la historia "Crear proyecto" fue detectado en el sprint "Sprint 1"
    Y el defecto de la historia "Crear proyecto" está sin resolver

  # ---------------------------------------------------------- casos alternativos
  @CA-022-2
  Esquema del escenario: Se pueden registrar defectos de las cuatro severidades
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se registra un defecto de severidad "<severidad>" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" tiene severidad "<severidad>"
    Y el defecto de la historia "Crear proyecto" tiene estado "abierto"
    Y el defecto de la historia "Crear proyecto" está sin resolver

    Ejemplos:
      | severidad |
      | critica   |
      | alta      |
      | media     |
      | baja      |

  @CA-022-3
  Escenario: El sprint de detección puede ser distinto del sprint de la historia
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y la historia "Crear proyecto" asignada al sprint "Sprint 1"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Cuando se registra un defecto de severidad "media" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 2"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" fue detectado en el sprint "Sprint 2"

  @CA-022-3
  Escenario: El sprint de detección puede estar cerrado
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y la historia "Crear proyecto" asignada al sprint "Sprint 1"
    Y el sprint "Sprint 1" cerrado
    Cuando se registra un defecto de severidad "baja" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación es exitosa
    Y el defecto de la historia "Crear proyecto" fue detectado en el sprint "Sprint 1"

  @CA-022-4
  Escenario: Se puede registrar un defecto sobre una historia completada y la historia no cambia
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y la historia "Crear proyecto" asignada al sprint "Sprint 1"
    Y la historia "Crear proyecto" completada
    Cuando se registra un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación es exitosa
    Y el proyecto "Demo" tiene 1 defecto
    Y la historia "Crear proyecto" tiene estado "completado"
    Y la historia "Crear proyecto" está en el sprint "Sprint 1"

  @CA-022-4
  Escenario: Se puede registrar un defecto sobre una historia que no está en ningún sprint
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se registra un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación es exitosa
    Y el proyecto "Demo" tiene 1 defecto
    Y la historia "Crear proyecto" tiene estado "pendiente"
    Y la historia "Crear proyecto" está en el backlog

  @CA-022-5
  Escenario: Se pueden registrar varios defectos sobre la misma historia, incluso idénticos
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se registra un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación es exitosa
    Y el proyecto "Demo" tiene 2 defectos

  # ---------------------------------------------------------------- casos limite
  @CA-022-6 @limite
  Esquema del escenario: La descripción admite hasta 2000 caracteres
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se registra un defecto de severidad "alta" con una descripción de <largo> caracteres en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación <resultado>

    Ejemplos:
      | largo | resultado  |
      | 2000  | es exitosa |
      | 2001  | se rechaza |

  # ------------------------------------------------------------------- errores
  @CA-022-6 @error
  Escenario: Una descripción de más de 2000 caracteres se rechaza y no se registra nada
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se registra un defecto de severidad "alta" con una descripción de 2001 caracteres en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la descripcion no puede tener mas de 2000 caracteres"
    Y el proyecto "Demo" tiene 0 defectos

  @CA-022-7 @error
  Escenario: Una descripción vacía se rechaza y no se registra nada
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se registra un defecto de severidad "alta" con la descripción "" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la descripcion es obligatoria"
    Y el proyecto "Demo" tiene 0 defectos

  @CA-022-8 @error
  Esquema del escenario: Una severidad fuera de critica, alta, media y baja se rechaza
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se registra un defecto de severidad "<severidad>" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la severidad tiene que ser critica, alta, media o baja"
    Y el proyecto "Demo" tiene 0 defectos

    Ejemplos:
      | severidad |
      | urgente   |
      | Critica   |
      |           |

  @CA-022-9 @error
  Escenario: No se registra un defecto sobre una historia que no existe
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se registra un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el item no existe"
    Y el proyecto "Demo" tiene 0 defectos

  @CA-022-9 @error
  Escenario: No se registra un defecto sobre una historia de otro proyecto
    Dado un proyecto "Otro"
    Y una historia "Listar proyectos" con prioridad "must"
    Y un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se registra un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Listar proyectos" detectado en el sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el item no existe"
    Y el proyecto "Otro" tiene 0 defectos

  @CA-022-10 @error
  Escenario: No se registra un defecto detectado en un sprint que no existe
    Dado un proyecto "Demo"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se registra un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint no existe"
    Y el proyecto "Demo" tiene 0 defectos

  @CA-022-10 @error
  Escenario: No se registra un defecto detectado en un sprint de otro proyecto
    Dado un proyecto "Otro"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y un proyecto "Demo"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se registra un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 2"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint no existe"
    Y el proyecto "Demo" tiene 0 defectos

  @CA-022-11 @error
  Escenario: Registrar un defecto en un proyecto que no existe se rechaza
    Dado ningún proyecto
    Cuando se registra un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"

  @CA-022-12 @error
  Escenario: Con la descripción vacía y la severidad inválida se informan los dos errores
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se registra un defecto de severidad "urgente" con la descripción "" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la descripcion es obligatoria"
    Y el mensaje de error indica "la severidad tiene que ser critica, alta, media o baja"
    Y el proyecto "Demo" tiene 0 defectos
