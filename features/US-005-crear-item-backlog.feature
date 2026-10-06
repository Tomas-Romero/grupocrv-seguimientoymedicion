# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Sin Antecedentes: cada escenario declara su proyecto, asi CA-005-8 puede
# arrancar sin ninguno.
@US-005
Característica: Crear un ítem de Product Backlog
  Como integrante del equipo
  quiero cargar un ítem en el Product Backlog del proyecto con título,
  descripción, prioridad y criterios de aceptación
  para tener registrado el trabajo pendiente antes de planificarlo

  Especificación: specs/US-005-crear-item-backlog.md

  # ---------------------------------------------------------------- caso normal
  @CA-005-1
  Escenario: Un ítem completo queda pendiente, sin estimar y con el número siguiente
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando se crea la historia "Listar proyectos" con prioridad "should", la descripción "Ver todos los proyectos" y los criterios:
      | Se ven todos los proyectos |
      | Se ordenan por nombre      |
    Entonces la operación es exitosa
    Y la historia "Listar proyectos" tiene el número 2
    Y la historia "Listar proyectos" tiene estado "pendiente"
    Y la historia "Listar proyectos" tiene prioridad "should"
    Y la historia "Listar proyectos" está sin estimar
    Y la historia "Listar proyectos" tiene los criterios:
      | Se ven todos los proyectos |
      | Se ordenan por nombre      |

  # ---------------------------------------------------------- casos alternativos
  @CA-005-2
  Escenario: Se puede crear sin descripción y sin criterios
    Dado un proyecto "Demo"
    Cuando se crea la historia "Alta de proyectos" con prioridad "could"
    Entonces la operación es exitosa
    Y el backlog del proyecto "Demo" tiene 1 historia

  @CA-005-3
  Escenario: La numeración es independiente por proyecto
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y un proyecto "Otro"
    Cuando se crea la historia "Alta de sprints" con prioridad "must"
    Entonces la operación es exitosa
    Y la historia "Alta de sprints" tiene el número 1
    Y el backlog del proyecto "Demo" tiene 1 historia

  # ---------------------------------------------------------------- casos limite
  @CA-005-4 @limite
  Esquema del escenario: El título tiene como máximo 120 caracteres
    Dado un proyecto "Demo"
    Cuando se crea una historia con un título de <largo> caracteres y prioridad "must"
    Entonces la operación <resultado>

    Ejemplos:
      | largo | resultado  |
      | 120   | es exitosa |
      | 121   | se rechaza |

  @CA-005-4 @limite
  Escenario: Los espacios de los extremos del título se recortan
    Dado un proyecto "Demo"
    Cuando se crea la historia "  Alta de proyectos  " con prioridad "must"
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" tiene el número 1

  # --------------------------------------------------------------------- errores
  @CA-005-5 @error
  Escenario: Un título vacío se rechaza y no se registra nada
    Dado un proyecto "Demo"
    Cuando se crea la historia "   " con prioridad "must"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el titulo es obligatorio"
    Y el backlog del proyecto "Demo" tiene 0 historias

  @CA-005-6 @error
  Escenario: Una prioridad fuera de MoSCoW se rechaza
    Dado un proyecto "Demo"
    Cuando se crea la historia "Alta de proyectos" con prioridad "urgente"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la prioridad tiene que ser must, should, could o wont"
    Y el backlog del proyecto "Demo" tiene 0 historias

  @CA-005-7 @error
  Escenario: Un criterio vacío se rechaza indicando su posición
    Dado un proyecto "Demo"
    Cuando se crea la historia "Alta de proyectos" con prioridad "must", la descripción "Alta" y los criterios:
      | Se guarda el proyecto |
      |                       |
    Entonces la operación se rechaza
    Y el mensaje de error indica "el criterio de aceptacion de la posicion 2 esta vacio"
    Y el backlog del proyecto "Demo" tiene 0 historias

  @CA-005-8 @error
  Escenario: Un proyecto inexistente se rechaza antes que los demás errores
    Dado ningún proyecto
    Cuando se crea la historia "   " con prioridad "urgente"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"
