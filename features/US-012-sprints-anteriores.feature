# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Frases nuevas de este .feature (se aprueban en el PR, ver docs/diccionario-steps.md):
# `se consultan los sprints anteriores`, `los sprints anteriores son:` (tabla sprint | planificados |
# completados | historias completadas | historias no completadas), `no hay sprints anteriores` y
# `la velocidad de los sprints anteriores es N`.
@US-012
Característica: Consultar los Sprints anteriores y su resultado
  Como integrante del equipo
  quiero consultar los Sprints anteriores y su resultado
  para aprender de lo que pasó y planificar mejor el próximo

  Especificación: specs/US-012-sprints-anteriores.md

  # ---------------------------------------------------------------- caso normal
  @CA-012-1
  Escenario: Los sprints cerrados salen del más reciente al más antiguo, cada uno con su resultado
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Listar proyectos" con prioridad "should", 3 story points y 1 criterio de aceptación
    Y una historia "Editar proyectos" con prioridad "should", 8 story points y 1 criterio de aceptación
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Listar proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y la historia "Editar proyectos" asignada al sprint "Sprint 2"
    Y la historia "Editar proyectos" completada
    Y el sprint "Sprint 2" cerrado
    Cuando se consultan los sprints anteriores
    Entonces la operación es exitosa
    Y los sprints anteriores son:
      | sprint   | planificados | completados | historias completadas | historias no completadas |
      | Sprint 2 | 8            | 8           | 1                     | 0                        |
      | Sprint 1 | 8            | 5           | 1                     | 1                        |

  # ---------------------------------------------------------- casos alternativos
  @CA-012-2
  Escenario: La velocidad es el promedio de los story points completados de los sprints cerrados
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Editar proyectos" con prioridad "should", 8 story points y 1 criterio de aceptación
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y la historia "Editar proyectos" asignada al sprint "Sprint 2"
    Y la historia "Editar proyectos" completada
    Y el sprint "Sprint 2" cerrado
    Cuando se consultan los sprints anteriores
    Entonces la operación es exitosa
    Y la velocidad de los sprints anteriores es 6,5

  @CA-012-3
  Escenario: El sprint abierto no aparece en el historial ni cuenta para la velocidad
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Editar proyectos" con prioridad "should", 8 story points y 1 criterio de aceptación
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y la historia "Editar proyectos" asignada al sprint "Sprint 2"
    Y la historia "Editar proyectos" completada
    Cuando se consultan los sprints anteriores
    Entonces los sprints anteriores son:
      | sprint   | planificados | completados | historias completadas | historias no completadas |
      | Sprint 1 | 5            | 5           | 1                     | 0                        |
    Y la velocidad de los sprints anteriores es 5

  @CA-012-5
  Escenario: Una historia que no se completó en un sprint y se completó después en otro sigue figurando como no completada en el primero
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Listar proyectos" con prioridad "should", 3 story points y 1 criterio de aceptación
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Listar proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y la historia "Listar proyectos" asignada al sprint "Sprint 2"
    Y la historia "Listar proyectos" completada
    Y el sprint "Sprint 2" cerrado
    Cuando se consultan los sprints anteriores
    Entonces los sprints anteriores son:
      | sprint   | planificados | completados | historias completadas | historias no completadas |
      | Sprint 2 | 3            | 3           | 1                     | 0                        |
      | Sprint 1 | 8            | 5           | 1                     | 1                        |

  # ---------------------------------------------------------------- casos limite
  @CA-012-4 @limite
  Escenario: Sin sprints cerrados no hay error: la lista está vacía y la velocidad es cero
    Dado un proyecto "Demo"
    Cuando se consultan los sprints anteriores
    Entonces la operación es exitosa
    Y no hay sprints anteriores
    Y la velocidad de los sprints anteriores es 0

  @CA-012-4 @limite
  Escenario: Un sprint cerrado sin historias aparece con todo en cero
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "Vacío"
    Y el sprint "Sprint 1" cerrado
    Cuando se consultan los sprints anteriores
    Entonces los sprints anteriores son:
      | sprint   | planificados | completados | historias completadas | historias no completadas |
      | Sprint 1 | 0            | 0           | 0                     | 0                        |
    Y la velocidad de los sprints anteriores es 0

  @limite
  Escenario: Si solo hay un sprint abierto, la lista está vacía
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se consultan los sprints anteriores
    Entonces la operación es exitosa
    Y no hay sprints anteriores

  # ------------------------------------------------------------------- errores
  @CA-012-6 @error
  Escenario: Un proyecto inexistente se rechaza
    Dado ningún proyecto
    Cuando se consultan los sprints anteriores
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"
