# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Frases nuevas de este .feature (se aprueban en el PR, ver docs/diccionario-steps.md):
# `el integrante "X" abre una nueva ronda en la sesión de la historia "H"` y
# `los votos de la ronda N de la sesión de la historia "H" son:` (tabla integrante | voto).
# Sin escenario todavia:
# - La sesion finalizada de CA-017-4 y la tercera ronda (CL-017-4): no hay frases de contexto
#   para una sesion finalizada ni para una ronda ya abierta. Las cubren los tests del dominio,
#   de la aplicacion, de memoria y de Postgres.
# - CA-017-5 (dos pedidos a la vez): lo cubre el test concurrente de Postgres, como en US-005.
@US-017
Característica: Abrir una nueva ronda de votación cuando hay divergencia
  Como integrante del equipo
  quiero abrir una nueva ronda de votación cuando los votos divergen
  para que el equipo discuta y vuelva a votar

  Especificación: specs/US-017-nueva-ronda.md

  # ---------------------------------------------------------------- caso normal
  @CA-017-1
  Escenario: Con divergencia, la nueva ronda deja la sesión votando en la ronda siguiente sin votos
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 3 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Beto" abre una nueva ronda en la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la sesión de la historia "Alta de proyectos" está "votando" en la ronda 2
    Y los participantes de la sesión de la historia "Alta de proyectos" son:
      | Ana  |
      | Beto |
    Y en la sesión de la historia "Alta de proyectos" faltan votar:
      | Ana  |
      | Beto |
    Y los votos de la sesión de la historia "Alta de proyectos" están ocultos

  @CA-017-1 @limite
  Escenario: Valores contiguos de la escala también abren una nueva ronda
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Ana" abre una nueva ronda en la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la sesión de la historia "Alta de proyectos" está "votando" en la ronda 2

  # ---------------------------------------------------------- casos alternativos
  @CA-017-2
  Escenario: Los votos de la ronda anterior se conservan y se ven, y los de la ronda nueva quedan ocultos
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y el integrante "Caro" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 3 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Beto" abre una nueva ronda en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" vota 5 en la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y los votos de la ronda 1 de la sesión de la historia "Alta de proyectos" son:
      | integrante | voto |
      | Ana        | 3    |
      | Beto       | 8    |
    Y en la sesión de la historia "Alta de proyectos" votaron:
      | Ana |
    Y los votos de la sesión de la historia "Alta de proyectos" están ocultos

  @CA-017-2
  Escenario: Revelar la ronda nueva muestra solo sus votos
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 3 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Beto" abre una nueva ronda en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" vota 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" vota 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la sesión de la historia "Alta de proyectos" está "revelada" en la ronda 2
    Y los votos de la sesión de la historia "Alta de proyectos" son:
      | integrante | voto |
      | Ana        | 5    |
      | Beto       | 5    |
    Y no hay divergencia en la sesión de la historia "Alta de proyectos"
    Y los votos de la ronda 1 de la sesión de la historia "Alta de proyectos" son:
      | integrante | voto |
      | Ana        | 3    |
      | Beto       | 8    |

  # ---------------------------------------------------------------- casos limite
  @CA-017-3 @limite
  Escenario: Si todos votaron lo mismo no se abre una nueva ronda
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Beto" abre una nueva ronda en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "no hay divergencia: registren la estimacion acordada"
    Y la sesión de la historia "Alta de proyectos" está "revelada" en la ronda 1

  @CA-017-3 @limite
  Escenario: Con un único voto emitido no hay divergencia y no se abre una nueva ronda
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 13 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Beto" abre una nueva ronda en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "no hay divergencia: registren la estimacion acordada"
    Y la sesión de la historia "Alta de proyectos" está "revelada" en la ronda 1

  # ------------------------------------------------------------------- errores
  @CA-017-4 @error
  Escenario: Con la sesión votando no se abre una nueva ronda
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 3 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Ana" abre una nueva ronda en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la nueva ronda se abre con los votos revelados y la sesion sin finalizar"
    Y la sesión de la historia "Alta de proyectos" está "votando" en la ronda 1
    Y los votos de la sesión de la historia "Alta de proyectos" están ocultos

  @CA-017-4 @error
  Escenario: Un integrante que no participa de la sesión no abre una nueva ronda
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 3 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Y el integrante "Caro" con rol "product_builder"
    Cuando el integrante "Caro" abre una nueva ronda en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el integrante no participa de esta sesion de planning poker"
    Y la sesión de la historia "Alta de proyectos" está "revelada" en la ronda 1

  @CA-017-4 @error
  Escenario: Abrir una nueva ronda en una historia sin sesión se rechaza
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando el integrante "Ana" abre una nueva ronda en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la sesion de planning poker no existe"

  @CA-017-4 @error
  Escenario: Un integrante que no existe no abre una nueva ronda
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 3 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Fantasma" abre una nueva ronda en la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el integrante no existe"
    Y la sesión de la historia "Alta de proyectos" está "revelada" en la ronda 1
