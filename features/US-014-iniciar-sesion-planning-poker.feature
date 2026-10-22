# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# La sesion se nombra por su historia: una historia tiene como maximo una sesion abierta
# (RN-014-7). Los participantes son los integrantes del proyecto al iniciar, y se
# comparan por nombre, en orden de alta en el proyecto.
@US-014
Característica: Iniciar una sesión de Planning Poker sobre una historia
  Como integrante del equipo
  quiero iniciar una sesión de Planning Poker sobre una historia del backlog
  para que el equipo la estime votando

  Especificación: specs/US-014-iniciar-sesion-planning-poker.md

  # ---------------------------------------------------------------- caso normal
  @CA-014-1
  Escenario: La sesión queda votando, en la ronda 1, con todos los integrantes
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y el integrante "Caro" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando se inicia una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la sesión de la historia "Alta de proyectos" está "votando" en la ronda 1
    Y los participantes de la sesión de la historia "Alta de proyectos" son:
      | Ana  |
      | Beto |
      | Caro |

  # ---------------------------------------------------------- casos alternativos
  @CA-014-2
  Escenario: Una historia ya estimada se vuelve a estimar en una sesión
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must" y 5 story points
    Cuando se inicia una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la sesión de la historia "Alta de proyectos" está "votando" en la ronda 1

  @CA-014-3
  Escenario: Un integrante que se suma después de iniciar no participa
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Cuando se registra al integrante "Caro" con rol "product_builder"
    Entonces la operación es exitosa
    Y los participantes de la sesión de la historia "Alta de proyectos" son:
      | Ana  |
      | Beto |

  # ---------------------------------------------------------------- casos limite
  @CA-014-4 @limite
  Escenario: Con dos integrantes la sesión se inicia
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando se inicia una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Entonces la operación es exitosa

  @CA-014-4 @limite
  Escenario: Con un solo integrante la sesión se rechaza
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando se inicia una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "se necesitan al menos 2 integrantes"

  # ------------------------------------------------------------------- errores
  @CA-014-5 @error
  Escenario: Una historia completada no se pone en sesión
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Cuando se inicia una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el item esta completado"

  @CA-014-6 @error
  Escenario: Una historia con una sesión abierta no admite otra
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Cuando se inicia una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la historia ya tiene una sesion de planning poker abierta"

  @CA-014-7 @error
  Escenario: Una historia que no existe se rechaza
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Cuando se inicia una sesión de Planning Poker sobre la historia "Historia fantasma"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el item no existe"

  @CA-014-7 @error
  Escenario: Un proyecto que no existe se rechaza, sin mirar la historia
    Dado ningún proyecto
    Cuando se inicia una sesión de Planning Poker sobre la historia "Historia fantasma"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"
