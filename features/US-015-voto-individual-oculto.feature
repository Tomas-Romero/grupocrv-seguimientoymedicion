# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Los escenarios no comprueban que valor quedo guardado: con la sesion votando no se
# puede ver (RN-015-7). Lo cubren los tests del repositorio, y los escenarios de US-016.
# Sin escenario todavia:
# - CA-015-6 con la sesion finalizada necesita US-018. Con la sesion revelada (US-016) si
#   hay escenario. La finalizada la cubren los tests del dominio, de la aplicacion y de
#   Postgres.
@US-015
Característica: Registrar el voto individual manteniéndolo oculto hasta el cierre
  Como integrante del equipo
  quiero registrar mi voto en una sesión de Planning Poker sin que los demás lo vean
  para estimar sin dejarme influir por los votos de los otros

  Especificación: specs/US-015-voto-individual-oculto.md

  # ---------------------------------------------------------------- caso normal
  @CA-015-1
  Escenario: Un participante vota y queda registrado que votó, sin que se vea el valor
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y el integrante "Caro" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Cuando el integrante "Ana" vota 5 en la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y en la sesión de la historia "Alta de proyectos" votaron:
      | Ana |
    Y en la sesión de la historia "Alta de proyectos" faltan votar:
      | Beto |
      | Caro |
    Y los votos de la sesión de la historia "Alta de proyectos" están ocultos

  # ---------------------------------------------------------- casos alternativos
  @CA-015-2
  Escenario: Un participante cambia su voto y figura una sola vez
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 3 en la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Ana" vota 8 en la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y en la sesión de la historia "Alta de proyectos" votaron:
      | Ana |
    Y en la sesión de la historia "Alta de proyectos" faltan votar:
      | Beto |

  @CA-015-3
  Escenario: Aunque hayan votado todos, los valores siguen ocultos
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Beto" vota 13 en la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y en la sesión de la historia "Alta de proyectos" votaron:
      | Ana  |
      | Beto |
    Y los votos de la sesión de la historia "Alta de proyectos" están ocultos

  # ---------------------------------------------------------------- casos limite
  @CA-015-4 @limite
  Esquema del escenario: Solo se vota con un valor de la escala
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Cuando el integrante "Ana" vota <valor> en la sesión de la historia "Alta de proyectos"
    Entonces la operación <resultado>

    Ejemplos:
      | valor | resultado  |
      | 1     | es exitosa |
      | 13    | es exitosa |
      | 0     | se rechaza |
      | 4     | se rechaza |
      | 21    | se rechaza |

  @CA-015-4 @limite
  Escenario: Un valor fuera de la escala no cuenta como voto
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Cuando el integrante "Ana" vota 4 en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la estimacion tiene que ser 1, 2, 3, 5, 8 o 13"
    Y en la sesión de la historia "Alta de proyectos" faltan votar:
      | Ana  |
      | Beto |

  # ------------------------------------------------------------------- errores
  @CA-015-5 @error
  Escenario: Un integrante que se sumó después de iniciar no vota
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Caro" con rol "product_builder"
    Cuando el integrante "Caro" vota 5 en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el integrante no participa de esta sesion de planning poker"

  @CA-015-7 @error
  Escenario: Votar en una historia sin sesión se rechaza
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando el integrante "Ana" vota 5 en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la sesion de planning poker no existe"

  @CA-015-7 @error
  Escenario: Un integrante que no existe no vota
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Cuando el integrante "Fantasma" vota 5 en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el integrante no existe"

  @CA-015-6 @error
  Escenario: Con la sesión revelada ya no se vota
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Beto" vota 8 en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la votacion de esta ronda esta cerrada"
    Y los votos de la sesión de la historia "Alta de proyectos" son:
      | integrante | voto |
      | Ana        | 5    |
