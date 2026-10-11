# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
@US-001
Característica: Crear un proyecto
  Como integrante del equipo
  quiero crear un proyecto con nombre, descripción y fechas de inicio y fin
  para tener el punto de partida del backlog, los sprints y las métricas

  Especificación: specs/US-001-crear-proyecto.md

  # ---------------------------------------------------------------- caso normal
  @CA-001-1
  Escenario: Un proyecto con nombre válido y fechas coherentes se crea
    Cuando se crea un proyecto "Demo" con fechas del "2026-09-14" al "2026-11-18"
    Entonces la operación es exitosa
    Y el proyecto "Demo" existe con fechas del "2026-09-14" al "2026-11-18"

  # ------------------------------------------------------------ caso alternativo
  @CA-001-2
  Escenario: Crear un segundo proyecto no modifica el primero
    Dado un proyecto "Alfa" con fechas del "2026-01-05" al "2026-03-06"
    Cuando se crea un proyecto "Beta" con fechas del "2026-04-01" al "2026-06-30"
    Entonces la operación es exitosa
    Y el proyecto "Beta" existe con fechas del "2026-04-01" al "2026-06-30"
    Y el proyecto "Alfa" existe con fechas del "2026-01-05" al "2026-03-06"

  # ----------------------------------------------------------------- caso limite
  @CA-001-3 @limite
  Escenario: Un proyecto de un solo día es válido
    Cuando se crea un proyecto "Relámpago" con fechas del "2026-10-12" al "2026-10-12"
    Entonces la operación es exitosa
    Y el proyecto "Relámpago" existe con fechas del "2026-10-12" al "2026-10-12"

  # ----------------------------------------------------------------------- error
  @CA-001-4 @error
  Escenario: La fecha de fin anterior a la de inicio se rechaza
    Cuando se crea un proyecto "Invertido" con fechas del "2026-11-18" al "2026-09-14"
    Entonces la operación se rechaza
    Y el mensaje de error indica "fecha de fin anterior a la de inicio"

  @CA-001-5 @error
  Escenario: Un nombre vacío se rechaza
    Cuando se crea un proyecto "" con fechas del "2026-09-14" al "2026-11-18"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el nombre es obligatorio"