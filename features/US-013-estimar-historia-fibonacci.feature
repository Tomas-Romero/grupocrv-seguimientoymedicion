# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# CA-013-5 (estimar una historia completada) no tiene escenario todavia: depende
# de `Dado la historia "X" completada`, que es del area Sprints y no esta
# conectada hasta US-010. Mientras tanto lo cubren los tests del dominio.
@US-013
Característica: Estimar una historia con Story Points en escala Fibonacci
  Como integrante del equipo
  quiero asignarle Story Points a una historia del backlog usando la escala Fibonacci
  para dimensionar el trabajo antes de planificar el sprint

  Especificación: specs/US-013-estimar-historia-fibonacci.md

  # ---------------------------------------------------------------- caso normal
  @CA-013-1
  Escenario: Una historia sin estimar queda con los story points elegidos
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando se estima la historia "Alta de proyectos" en 5 story points
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" tiene 5 story points

  # ---------------------------------------------------------- casos alternativos
  @CA-013-2
  Escenario: Se puede volver a estimar y la última estimación reemplaza a la anterior
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must" y 3 story points
    Cuando se estima la historia "Alta de proyectos" en 8 story points
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" tiene 8 story points

  # ---------------------------------------------------------------- casos limite
  @CA-013-3 @limite
  Esquema del escenario: La escala va de 1 a 13 y no incluye el 0 ni el 21
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando se estima la historia "Alta de proyectos" en <puntos> story points
    Entonces la operación <resultado>

    Ejemplos:
      | puntos | resultado  |
      | 1      | es exitosa |
      | 13     | es exitosa |
      | 0      | se rechaza |
      | 21     | se rechaza |

  # ------------------------------------------------------------------- errores
  @CA-013-4 @error
  Escenario: Un valor fuera de la escala se rechaza y la historia conserva su estimación
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must" y 3 story points
    Cuando se estima la historia "Alta de proyectos" en 4 story points
    Entonces la operación se rechaza
    Y el mensaje de error indica "la estimacion tiene que ser 1, 2, 3, 5, 8 o 13"
    Y la historia "Alta de proyectos" tiene 3 story points

  @CA-013-4 @error
  Escenario: Un valor fuera de la escala no deja la historia estimada
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando se estima la historia "Alta de proyectos" en 21 story points
    Entonces la operación se rechaza
    Y la historia "Alta de proyectos" está sin estimar

  @CA-013-6 @error
  Escenario: Estimar una historia que no existe se rechaza
    Dado un proyecto "Demo"
    Cuando se estima la historia "Historia fantasma" en 5 story points
    Entonces la operación se rechaza
    Y el mensaje de error indica "el item no existe"
