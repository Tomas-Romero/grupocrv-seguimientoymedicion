# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# El nombre entre comillas ("Sprint 1") es el nombre ESPERADO: el sprint lo nombra la
# aplicacion y, si se crea con otro nombre, el step falla.
# CA-008-4 y CA-008-9 (empezar despues o antes del fin del ultimo sprint) no tienen
# escenario todavia: necesitan un sprint anterior cerrado y cerrar es US-011. Los
# cubren los tests del dominio y de Postgres; sus escenarios llegan con US-011.
@US-008
Característica: Crear un Sprint con Sprint Goal y fechas
  Como integrante del equipo
  quiero crear un Sprint con su Sprint Goal y sus fechas
  para planificar el trabajo de un período

  Especificación: specs/US-008-crear-sprint.md

  # ---------------------------------------------------------------- caso normal
  @CA-008-1
  Escenario: Un sprint válido queda abierto y con el número siguiente del proyecto
    Dado un proyecto "Demo"
    Cuando se crea el sprint "Sprint 1" con el objetivo "MVP navegable" del "2026-10-06" al "2026-10-19"
    Entonces la operación es exitosa
    Y el sprint "Sprint 1" está abierto

  # ---------------------------------------------------------- casos alternativos
  @CA-008-2
  Escenario: El primer sprint de cada proyecto es el Sprint 1
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y un proyecto "Otro"
    Cuando se crea el sprint "Sprint 1" con el objetivo "Primer objetivo del otro proyecto"
    Entonces la operación es exitosa
    Y el sprint "Sprint 1" está abierto

  # ---------------------------------------------------------------- casos limite
  @CA-008-3 @limite
  Esquema del escenario: Un sprint dura entre 1 y 28 días
    Dado un proyecto "Demo"
    Cuando se crea el sprint "Sprint 1" con el objetivo "Objetivo" del "2026-10-06" al "<fin>"
    Entonces la operación <resultado>

    Ejemplos:
      | fin        | resultado  |
      | 2026-10-06 | es exitosa |
      | 2026-11-02 | es exitosa |
      | 2026-11-03 | se rechaza |

  @CA-008-5 @limite
  Escenario: Un sprint que ocupa exactamente las fechas del proyecto es válido
    Dado un proyecto "Corto" con fechas del "2026-10-06" al "2026-11-02"
    Cuando se crea el sprint "Sprint 1" con el objetivo "Todo el proyecto" del "2026-10-06" al "2026-11-02"
    Entonces la operación es exitosa

  # ------------------------------------------------------------------- errores
  @CA-008-6 @error
  Escenario: Un Sprint Goal vacío se rechaza
    Dado un proyecto "Demo"
    Cuando se crea el sprint "Sprint 1" con el objetivo "" del "2026-10-06" al "2026-10-19"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el Sprint Goal es obligatorio"

  @CA-008-7 @error
  Escenario: La fecha de fin anterior a la de inicio se rechaza
    Dado un proyecto "Demo"
    Cuando se crea el sprint "Sprint 1" con el objetivo "MVP navegable" del "2026-10-19" al "2026-10-06"
    Entonces la operación se rechaza
    Y el mensaje de error indica "fecha de fin anterior a la de inicio"

  @CA-008-8 @error
  Escenario: Un sprint fuera de las fechas del proyecto se rechaza
    Dado un proyecto "Corto" con fechas del "2026-10-06" al "2026-11-02"
    Cuando se crea el sprint "Sprint 1" con el objetivo "MVP navegable" del "2026-10-05" al "2026-10-12"
    Entonces la operación se rechaza
    Y el mensaje de error indica "las fechas del sprint tienen que estar dentro de las del proyecto"

  @CA-008-8 @error
  Escenario: Con varios datos inválidos se informan todos juntos
    Dado un proyecto "Corto" con fechas del "2026-10-06" al "2026-11-02"
    Cuando se crea el sprint "Sprint 1" con el objetivo "" del "2026-10-10" al "2026-10-01"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el Sprint Goal es obligatorio"
    Y el mensaje de error indica "fecha de fin anterior a la de inicio"
    Y el mensaje de error indica "las fechas del sprint tienen que estar dentro de las del proyecto"

  @CA-008-10 @error
  Escenario: Crear un sprint en un proyecto que no existe se rechaza, sin mirar los demás datos
    Dado ningún proyecto
    Cuando se crea el sprint "Sprint 1" con el objetivo "" del "2026-10-19" al "2026-10-06"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"

  @CA-008-11 @error
  Escenario: Con el último sprint abierto no se puede crear otro
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se crea el sprint "Sprint 2" con el objetivo "Segundo objetivo"
    Entonces la operación se rechaza
    Y el mensaje de error indica "primero hay que cerrarlo"
