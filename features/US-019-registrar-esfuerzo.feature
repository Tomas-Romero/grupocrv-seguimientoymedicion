# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# En los escenarios "hoy" es el 2026-10-06 (el reloj fijo de ahoraEnEscenarios).
# CA-019-3 (registrar sobre una historia completada) no tiene escenario todavia:
# depende de `Dado la historia "X" completada`, del area Sprints, que se conecta con
# US-010. Mientras tanto lo cubren los tests de la base en memoria y de Postgres.
@US-019
Característica: Registrar esfuerzo
  Como integrante del equipo
  quiero registrar las horas que dediqué a una historia, con la fecha y la actividad
  para medir el esfuerzo real y compararlo contra lo estimado

  Especificación: specs/US-019-registrar-esfuerzo.md

  # ---------------------------------------------------------------- caso normal
  @CA-019-1
  Escenario: Un integrante registra sus horas de una historia
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando el integrante "Ana" registra 2,5 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-10-05
    Entonces la operación es exitosa
    Y el proyecto "Demo" tiene 1 registro de esfuerzo
    Y la historia "Alta de proyectos" tiene 2,5 horas registradas

  # ---------------------------------------------------------- casos alternativos
  @CA-019-2
  Escenario: Se pueden registrar varias veces horas del mismo integrante el mismo día
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y el integrante "Ana" ya registró 3 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-10-05
    Cuando el integrante "Ana" registra 2 horas de "Revisión" en la historia "Alta de proyectos" el 2026-10-05
    Entonces la operación es exitosa
    Y el proyecto "Demo" tiene 2 registros de esfuerzo
    Y la historia "Alta de proyectos" tiene 5 horas registradas

  @CA-019-2 @limite
  Escenario: Dos registros del mismo día que suman más de 24 horas se aceptan
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y el integrante "Ana" ya registró 20 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-10-05
    Cuando el integrante "Ana" registra 20 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-10-05
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" tiene 40 horas registradas

  # ---------------------------------------------------------------- casos limite
  @CA-019-4 @limite
  Esquema del escenario: Las horas van de 0,25 a 24 en múltiplos de 0,25
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando el integrante "Ana" registra <horas> horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-10-05
    Entonces la operación <resultado>

    Ejemplos:
      | horas | resultado  |
      | 0,25  | es exitosa |
      | 24    | es exitosa |
      | 0     | se rechaza |
      | 24,25 | se rechaza |

  @CA-019-5 @limite
  Esquema del escenario: La fecha no puede ser posterior a hoy
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando el integrante "Ana" registra 1 horas de "Desarrollo" en la historia "Alta de proyectos" el <fecha>
    Entonces la operación <resultado>

    Ejemplos:
      | fecha      | resultado  |
      | 2026-10-06 | es exitosa |
      | 2026-10-07 | se rechaza |

  # ------------------------------------------------------------------- errores
  @CA-019-6 @error
  Escenario: Horas negativas se rechazan y no se registra nada
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando el integrante "Ana" registra -1 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-10-05
    Entonces la operación se rechaza
    Y el mensaje de error indica "las horas tienen que ser mayores a 0"
    Y el proyecto "Demo" tiene 0 registros de esfuerzo

  @CA-019-7 @error
  Escenario: Horas que no son múltiplo de 0,25 se rechazan
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando el integrante "Ana" registra 1,3 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-10-05
    Entonces la operación se rechaza
    Y el mensaje de error indica "multiplos de 0,25"

  @CA-019-8 @error
  Escenario: Una actividad vacía se rechaza
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando el integrante "Ana" registra 2 horas de "" en la historia "Alta de proyectos" el 2026-10-05
    Entonces la operación se rechaza
    Y el mensaje de error indica "la actividad es obligatoria"

  @CA-019-8 @error
  Escenario: Con varios datos inválidos se informan todos juntos
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando el integrante "Ana" registra 0 horas de "" en la historia "Alta de proyectos" el 2026-10-07
    Entonces la operación se rechaza
    Y el mensaje de error indica "la fecha no puede ser posterior a hoy"
    Y el mensaje de error indica "la actividad es obligatoria"
    Y el mensaje de error indica "las horas tienen que ser mayores a 0"

  @CA-019-9 @error
  Escenario: Un integrante de otro proyecto no puede registrar esfuerzo
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y un proyecto "Otro"
    Y una historia "Listar proyectos" con prioridad "must"
    Cuando el integrante "Ana" registra 2 horas de "Desarrollo" en la historia "Listar proyectos" el 2026-10-05
    Entonces la operación se rechaza
    Y el mensaje de error indica "el integrante no existe"
    Y el proyecto "Otro" tiene 0 registros de esfuerzo

  @CA-019-9 @error
  Escenario: No se registra esfuerzo sobre una historia de otro proyecto
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y un proyecto "Otro"
    Y el integrante "Beto" con rol "product_builder"
    Cuando el integrante "Beto" registra 2 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-10-05
    Entonces la operación se rechaza
    Y el mensaje de error indica "el item no existe"

  @CA-019-10 @error
  Escenario: Registrar esfuerzo en un proyecto que no existe se rechaza
    Dado ningún proyecto
    Cuando el integrante "Ana" registra 2 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-10-05
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"
