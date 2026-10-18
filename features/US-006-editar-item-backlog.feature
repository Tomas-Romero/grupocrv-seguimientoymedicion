# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Sin escenario todavia:
# - CA-006-3 (editar una historia asignada a un sprint abierto) necesita crear un
#   sprint y asignarle la historia, y US-008 y US-009 no estan en main. Lo cubren
#   los tests del dominio y del repositorio; el escenario llega con esas historias.
# - CA-006-6 (una historia completada no se edita) necesita
#   `Dado la historia "X" completada`, que conecta US-010. Lo cubren los tests del
#   dominio y del repositorio.
# - La parte de CA-006-4 de la lista de criterios vacia: una tabla de Gherkin no
#   puede estar vacia, asi que no hay frase que la escriba. La cubren los tests.
@US-006
Característica: Editar y repriorizar ítems del backlog
  Como integrante del equipo
  quiero editar los datos de un ítem del backlog y cambiar su prioridad
  para mantener el backlog al día a medida que se refina el trabajo

  Especificación: specs/US-006-editar-item-backlog.md

  # ---------------------------------------------------------------- caso normal
  @CA-006-1
  Escenario: Una historia editada queda con los datos nuevos y conserva número, estado y estimación
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una historia "Listar proyectos" con prioridad "could" y 5 story points
    Cuando se edita la historia "Listar proyectos" con el título "Listar proyectos activos", la prioridad "should", la descripción "Solo los que no terminaron" y los criterios:
      | Se ven los proyectos activos |
      | Se ordenan por nombre        |
    Entonces la operación es exitosa
    Y la historia "Listar proyectos activos" tiene prioridad "should"
    Y la historia "Listar proyectos activos" tiene los criterios:
      | Se ven los proyectos activos |
      | Se ordenan por nombre        |
    Y la historia "Listar proyectos activos" tiene el número 2
    Y la historia "Listar proyectos activos" tiene estado "pendiente"
    Y la historia "Listar proyectos activos" tiene 5 story points

  # ---------------------------------------------------------- casos alternativos
  @CA-006-2
  Escenario: Repriorizar cambia solo la prioridad
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "could" y 3 story points
    Cuando se cambia la prioridad de la historia "Alta de proyectos" a "must"
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" tiene prioridad "must"
    Y la historia "Alta de proyectos" tiene 3 story points
    Y la historia "Alta de proyectos" tiene el número 1
    Y la historia "Alta de proyectos" tiene estado "pendiente"

  # ---------------------------------------------------------------- casos limite
  @CA-006-4 @limite
  Esquema del escenario: El título editado tiene como máximo 120 caracteres
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando se edita la historia "Alta de proyectos" con el título "<titulo>", la prioridad "must", la descripción "" y los criterios:
      | Se guarda |
    Entonces la operación <resultado>

    # Los títulos repiten "1234567890": 12 veces son 120 caracteres.
    Ejemplos:
      | titulo                                                                                                                    | resultado  |
      | 123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890  | es exitosa |
      | 1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901 | se rechaza |

  # ------------------------------------------------------------------- errores
  @CA-006-5 @error
  Esquema del escenario: Con datos inválidos se rechaza y la historia queda como estaba
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "should" y 5 story points
    Cuando se edita la historia "Alta de proyectos" con el título "<titulo>", la prioridad "<prioridad>", la descripción "Nueva" y los criterios:
      | <criterio> |
    Entonces la operación se rechaza
    Y el mensaje de error indica "<mensaje>"
    Y la historia "Alta de proyectos" tiene prioridad "should"
    Y la historia "Alta de proyectos" tiene 5 story points

    Ejemplos:
      | titulo            | prioridad | criterio  | mensaje                                             |
      |                   | must      | Se guarda | el titulo es obligatorio                            |
      | Alta de proyectos | urgente   | Se guarda | la prioridad tiene que ser must, should, could o wont |
      | Alta de proyectos | must      |           | el criterio de aceptacion de la posicion 1 esta vacio |

  @CA-006-7 @error
  Escenario: Editar una historia que no existe se rechaza y no cambia nada
    Dado un proyecto "Demo"
    Y una historia "Alta de proyectos" con prioridad "must"
    Cuando se edita la historia "Historia fantasma" con el título "", la prioridad "urgente", la descripción "" y los criterios:
      | Se guarda |
    Entonces la operación se rechaza
    Y el mensaje de error indica "el item no existe"
    Y el backlog del proyecto "Demo" tiene 1 historia
    Y la historia "Alta de proyectos" tiene prioridad "must"
