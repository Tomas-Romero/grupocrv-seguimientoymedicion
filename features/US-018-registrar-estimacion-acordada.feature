# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Una sesion finalizada ya no esta abierta (RN-014-7): las frases de la sesion leen la
# ultima sesion de la historia.
# Sin escenario todavia:
# - La parte de CA-018-5 de una sesion ya finalizada: no hay frase de contexto para una
#   sesion finalizada, y el escenario necesitaria dos registros. La cubren los tests del
#   dominio, de la aplicacion, de memoria y de Postgres.
@US-018
Característica: Registrar la estimación acordada como valor final de la historia
  Como integrante del equipo
  quiero registrar la estimación que acordamos en la sesión de Planning Poker
  para que quede como el valor final de la historia

  Especificación: specs/US-018-registrar-estimacion-acordada.md

  # ---------------------------------------------------------------- caso normal
  @CA-018-1
  Escenario: La sesión queda finalizada con el valor acordado y la historia, estimada
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Beto" registra la estimación acordada de 5 en la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la sesión de la historia "Alta de proyectos" está "finalizada" en la ronda 1
    Y la estimación acordada de la sesión de la historia "Alta de proyectos" es 5
    Y la historia "Alta de proyectos" tiene 5 story points

  # ---------------------------------------------------------- casos alternativos
  @CA-018-2
  Escenario: Se acuerda un valor que nadie votó, aunque haya divergencia
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 3 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Ana" registra la estimación acordada de 5 en la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la estimación acordada de la sesión de la historia "Alta de proyectos" es 5
    Y la historia "Alta de proyectos" tiene 5 story points

  @CA-018-3
  Escenario: Una historia ya estimada queda con el valor acordado
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must" y 3 story points
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Ana" registra la estimación acordada de 8 en la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" tiene 8 story points

  # ---------------------------------------------------------------- casos limite
  @CA-018-4 @limite
  Esquema del escenario: Solo se acuerda un valor de la escala
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Ana" registra la estimación acordada de <valor> en la sesión de la historia "Alta de proyectos"
    Entonces la operación <resultado>

    Ejemplos:
      | valor | resultado  |
      | 1     | es exitosa |
      | 13    | es exitosa |
      | 0     | se rechaza |
      | 4     | se rechaza |
      | 21    | se rechaza |

  @CA-018-4 @limite
  Escenario: Un valor fuera de la escala deja la sesión revelada y la historia sin estimar
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Ana" registra la estimación acordada de 4 en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la estimacion tiene que ser 1, 2, 3, 5, 8 o 13"
    Y la sesión de la historia "Alta de proyectos" está "revelada" en la ronda 1
    Y la historia "Alta de proyectos" está sin estimar

  # ------------------------------------------------------------------- errores
  @CA-018-5 @error
  Escenario: Con la sesión votando no se registra
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Ana" registra la estimación acordada de 5 en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la estimacion acordada se registra con los votos revelados"
    Y la sesión de la historia "Alta de proyectos" está "votando" en la ronda 1
    Y la historia "Alta de proyectos" está sin estimar

  @CA-018-6 @error
  Escenario: Un integrante que no participa de la sesión no registra
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Y el integrante "Caro" con rol "product_builder"
    Cuando el integrante "Caro" registra la estimación acordada de 5 en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el integrante no participa de esta sesion de planning poker"

  @CA-018-7 @error
  Escenario: Si la historia se completó, no se registra y la sesión sigue revelada
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Y la historia "Alta de proyectos" completada
    Cuando el integrante "Ana" registra la estimación acordada de 8 en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el item esta completado"
    Y la sesión de la historia "Alta de proyectos" está "revelada" en la ronda 1
    Y la historia "Alta de proyectos" tiene 5 story points

  @CA-018-8 @error
  Escenario: Registrar en una historia sin sesión se rechaza
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando el integrante "Ana" registra la estimación acordada de 5 en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la sesion de planning poker no existe"

  @CA-018-8 @error
  Escenario: Un integrante que no existe no registra
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Fantasma" registra la estimación acordada de 5 en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el integrante no existe"
