# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
@US-NNN
Característica: Titulo de la historia
  Como <rol>
  quiero <accion>
  para <beneficio>

  Especificacion: specs/US-NNN-slug.md

  Antecedentes:
    Dado un proyecto "Demo" con fechas del "2026-09-14" al "2026-11-16"

  # ---------------------------------------------------------------- caso normal
  @CA-NNN-1
  Escenario: El camino feliz, descrito por lo que ve el usuario
    Dado ...
    Cuando ...
    Entonces ...

  # ------------------------------------------------------------ caso alternativo
  @CA-NNN-2
  Escenario: Una variante valida del camino principal
    Dado ...
    Cuando ...
    Entonces ...

  # ----------------------------------------------------------------- caso limite
  @CA-NNN-3 @limite
  Escenario: El borde que siempre rompe
    Dado ningún sprint cerrado
    Cuando se consulta la velocidad del equipo
    Entonces la operación es exitosa
    Y la velocidad del equipo es 0

  # ----------------------------------------------------------------------- error
  @CA-NNN-4 @error
  Escenario: Entrada invalida
    Dado ...
    Cuando se intenta ...
    Entonces se rechaza la operacion
    Y el mensaje indica "..."

  # ------------------------------------------------- varios casos con una tabla
  @CA-NNN-5
  Esquema del escenario: Validacion de rangos
    Cuando se registran <horas> horas
    Entonces el resultado es "<resultado>"

    Ejemplos:
      | horas | resultado |
      | 0     | rechazado |
      | 0.5   | aceptado  |
      | 24    | aceptado  |
      | 25    | rechazado |
      | -1    | rechazado |
