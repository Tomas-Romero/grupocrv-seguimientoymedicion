# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Frases nuevas de este .feature (se aprueban en el PR, ver docs/diccionario-steps.md):
# `el integrante "X" revela los votos de la sesión de la historia "H"` (y su variante de
# contexto, `el integrante "X" reveló ...`), `los votos de la sesión de la historia "H" son:`
# (tabla integrante | voto), `hay divergencia en la sesión de la historia "H"` y su negación.
# Revelar con la sesion finalizada necesita US-018 para llegar a ese estado: se prueba la
# revelada, y la finalizada la cubren los tests del dominio.
@US-016
Característica: Revelar las estimaciones y detectar divergencias entre integrantes
  Como integrante del equipo
  quiero revelar las estimaciones de todos a la vez y ver si hay divergencias
  para decidir si hace falta discutir y votar de nuevo

  Especificación: specs/US-016-revelar-estimaciones.md

  # ---------------------------------------------------------------- caso normal
  @CA-016-1
  Escenario: Si todos votaron lo mismo, se revelan los votos y no hay divergencia
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y el integrante "Caro" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Caro" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Ana" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la sesión de la historia "Alta de proyectos" está "revelada" en la ronda 1
    Y los votos de la sesión de la historia "Alta de proyectos" son:
      | integrante | voto |
      | Ana        | 5    |
      | Beto       | 5    |
      | Caro       | 5    |
    Y no hay divergencia en la sesión de la historia "Alta de proyectos"

  # ---------------------------------------------------------- casos alternativos
  @CA-016-2
  Escenario: Si votaron valores distintos, se revelan los votos y hay divergencia
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y el integrante "Caro" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 3 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Caro" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Beto" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y los votos de la sesión de la historia "Alta de proyectos" son:
      | integrante | voto |
      | Ana        | 3    |
      | Beto       | 5    |
      | Caro       | 8    |
    Y hay divergencia en la sesión de la historia "Alta de proyectos"

  @CA-016-3
  Escenario: Se puede revelar aunque falten votos, y se sabe quiénes no votaron
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y el integrante "Caro" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Caro" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la sesión de la historia "Alta de proyectos" está "revelada" en la ronda 1
    Y los votos de la sesión de la historia "Alta de proyectos" son:
      | integrante | voto |
      | Ana        | 5    |
      | Beto       | 5    |
    Y en la sesión de la historia "Alta de proyectos" faltan votar:
      | Caro |
    Y no hay divergencia en la sesión de la historia "Alta de proyectos"

  @CA-016-3
  Escenario: Un participante que cambió su voto antes de revelar figura una sola vez, con el último
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 3 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Ana" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y los votos de la sesión de la historia "Alta de proyectos" son:
      | integrante | voto |
      | Ana        | 8    |
      | Beto       | 8    |
    Y no hay divergencia en la sesión de la historia "Alta de proyectos"

  # ---------------------------------------------------------------- casos limite
  @CA-016-4 @limite
  Escenario: Valores contiguos de la escala también son una divergencia
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Beto" ya votó 8 en la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Ana" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y hay divergencia en la sesión de la historia "Alta de proyectos"

  @CA-016-4 @limite
  Escenario: Con un único voto emitido se revela y no hay divergencia
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 13 en la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Beto" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y los votos de la sesión de la historia "Alta de proyectos" son:
      | integrante | voto |
      | Ana        | 13   |
    Y no hay divergencia en la sesión de la historia "Alta de proyectos"

  # ------------------------------------------------------------------- errores
  @CA-016-5 @error
  Escenario: Sin ningún voto no se revela y la sesión sigue votando
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Cuando el integrante "Ana" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "no hay votos para revelar en esta ronda"
    Y la sesión de la historia "Alta de proyectos" está "votando" en la ronda 1
    Y los votos de la sesión de la historia "Alta de proyectos" están ocultos

  @CA-016-6 @error
  Escenario: Un integrante que no participa de la sesión no revela
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Caro" con rol "product_builder"
    Cuando el integrante "Caro" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el integrante no participa de esta sesion de planning poker"
    Y la sesión de la historia "Alta de proyectos" está "votando" en la ronda 1
    Y los votos de la sesión de la historia "Alta de proyectos" están ocultos

  @CA-016-7 @error
  Escenario: Una sesión ya revelada no se revela de nuevo
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Y el integrante "Ana" reveló los votos de la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Beto" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la votacion de esta ronda esta cerrada"
    Y la sesión de la historia "Alta de proyectos" está "revelada" en la ronda 1

  @CA-016-8 @error
  Escenario: Revelar en una historia sin sesión se rechaza
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando el integrante "Ana" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la sesion de planning poker no existe"

  @CA-016-8 @error
  Escenario: Un integrante que no existe no revela
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_architect"
    Y el integrante "Beto" con rol "agile_enabler"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una sesión de Planning Poker sobre la historia "Alta de proyectos"
    Y el integrante "Ana" ya votó 5 en la sesión de la historia "Alta de proyectos"
    Cuando el integrante "Fantasma" revela los votos de la sesión de la historia "Alta de proyectos"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el integrante no existe"
    Y la sesión de la historia "Alta de proyectos" está "votando" en la ronda 1
