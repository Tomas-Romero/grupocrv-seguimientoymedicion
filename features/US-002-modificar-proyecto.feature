# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
@US-002
Característica: Modificar los datos de un proyecto existente
  Como integrante del equipo
  quiero corregir el nombre y las fechas de un proyecto ya creado
  para no tener que crearlo de nuevo ni perder lo que cuelga de él

  Especificacion: specs/US-002-modificar-proyecto.md

  Antecedentes:
    Dado un proyecto "Demo" con fechas del "2026-09-14" al "2026-11-16"

  # ---------------------------------------------------------------- caso normal
  @CA-002-1
  Escenario: Al modificar el nombre, el proyecto conserva sus fechas
    Cuando se modifica el nombre del proyecto "Demo" a "Demo v2"
    Entonces la operación es exitosa
    Y el proyecto "Demo v2" existe con fechas del "2026-09-14" al "2026-11-16"

  # ------------------------------------------------------------ caso alternativo
  @CA-002-2
  Escenario: Al modificar las fechas, el proyecto conserva su nombre
    Cuando se modifican las fechas del proyecto "Demo" del "2026-10-01" al "2026-12-15"
    Entonces la operación es exitosa
    Y el proyecto "Demo" existe con fechas del "2026-10-01" al "2026-12-15"

  # ----------------------------------------------------------------- caso limite
  @CA-002-3 @limite
  Escenario: Las fechas de inicio y fin pueden ser iguales
    Cuando se modifican las fechas del proyecto "Demo" del "2026-10-01" al "2026-10-01"
    Entonces la operación es exitosa
    Y el proyecto "Demo" existe con fechas del "2026-10-01" al "2026-10-01"

  # ----------------------------------------------------------------------- error
  @CA-002-4 @error
  Escenario: Un nombre nuevo vacío se rechaza y el proyecto no cambia
    Cuando se modifica el nombre del proyecto "Demo" a ""
    Entonces la operación se rechaza
    Y el mensaje de error indica "el nombre es obligatorio"
    Y el proyecto "Demo" existe con fechas del "2026-09-14" al "2026-11-16"

  @CA-002-5 @error
  Escenario: Modificar un proyecto que no existe se rechaza
    Cuando se modifica el nombre del proyecto "Fantasma" a "Otro"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"