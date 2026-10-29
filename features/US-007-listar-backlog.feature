# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Las frases `se lista el backlog ...` y `la lista del backlog ...` son nuevas (spec US-007, seccion 5): se
# aprueban y se suman al diccionario en el PR de implementacion.
# Una historia `completada` se arma con las frases de US-009 y US-010 (historia estimada y con criterios,
# sprint, asignacion y marcado). Ninguna operacion deja una historia en `en_progreso` (US-004, hueco 3), asi que
# ese estado solo se prueba como filtro que no devuelve nada; con historias lo cubren los tests del caso de uso.
# Que con el proyecto inexistente se devuelva SOLO ese error (CA-007-6) no se puede afirmar aca (el mensaje se
# compara por contenido, no por exclusion): lo prueban los tests del caso de uso con errors.Is.
@US-007
Característica: Listar y filtrar el backlog de un proyecto por estado y prioridad
  Como integrante del equipo
  quiero listar el Product Backlog de un proyecto filtrando por el estado y la prioridad de las historias
  para encontrar enseguida las que me interesan sin recorrer el backlog entero

  Especificación: specs/US-007-listar-backlog.md

  # ---------------------------------------------------------------- caso normal
  @CA-007-1
  Escenario: Filtrar por un estado deja solo las historias en ese estado
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Editar proyecto" con prioridad "should"
    Y una historia "Listar proyectos" con prioridad "must"
    Y la historia "Crear proyecto" asignada al sprint "Sprint 1"
    Y la historia "Crear proyecto" completada
    Cuando se lista el backlog con estado "pendiente"
    Entonces la operación es exitosa
    Y la lista del backlog tiene 2 historias
    Y la lista del backlog incluye la historia "Editar proyecto"
    Y la lista del backlog incluye la historia "Listar proyectos"
    Y la lista del backlog no incluye la historia "Crear proyecto"

  @CA-007-1
  Escenario: Filtrar por completado deja solo las historias completadas
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Editar proyecto" con prioridad "should"
    Y la historia "Crear proyecto" asignada al sprint "Sprint 1"
    Y la historia "Crear proyecto" completada
    Cuando se lista el backlog con estado "completado"
    Entonces la operación es exitosa
    Y la lista del backlog tiene 1 historia
    Y la lista del backlog incluye la historia "Crear proyecto"
    Y la lista del backlog no incluye la historia "Editar proyecto"

  @CA-007-1
  Escenario: Filtrar por una prioridad deja solo las historias de esa prioridad, en cualquier estado
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Editar proyecto" con prioridad "should"
    Y una historia "Listar proyectos" con prioridad "must"
    Y la historia "Crear proyecto" asignada al sprint "Sprint 1"
    Y la historia "Crear proyecto" completada
    Cuando se lista el backlog con prioridad "must"
    Entonces la operación es exitosa
    Y la lista del backlog tiene 2 historias
    Y la lista del backlog incluye la historia "Crear proyecto"
    Y la lista del backlog incluye la historia "Listar proyectos"
    Y la lista del backlog no incluye la historia "Editar proyecto"

  @CA-007-1
  Escenario: El filtro no cambia el orden: primero por prioridad y dentro de cada una por número
    Dado un proyecto "Demo"
    Y una historia "Alta de sprints" con prioridad "could"
    Y una historia "Crear proyecto" con prioridad "should"
    Y una historia "Editar proyecto" con prioridad "must"
    Y una historia "Listar proyectos" con prioridad "should"
    Y una historia "Cerrar sprint" con prioridad "must"
    Cuando se lista el backlog con estado "pendiente"
    Entonces la operación es exitosa
    Y la lista del backlog muestra las historias en el orden:
      | Editar proyecto  |
      | Cerrar sprint    |
      | Crear proyecto   |
      | Listar proyectos |
      | Alta de sprints  |

  # ---------------------------------------------------------- casos alternativos
  @CA-007-2
  Escenario: Los dos filtros se combinan: salen solo las historias que cumplen ambos
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Editar proyecto" con prioridad "should"
    Y una historia "Listar proyectos" con prioridad "must"
    Y la historia "Crear proyecto" asignada al sprint "Sprint 1"
    Y la historia "Crear proyecto" completada
    Cuando se lista el backlog con estado "pendiente" y prioridad "must"
    Entonces la operación es exitosa
    Y la lista del backlog tiene 1 historia
    Y la lista del backlog incluye la historia "Listar proyectos"
    Y la lista del backlog no incluye la historia "Crear proyecto"
    Y la lista del backlog no incluye la historia "Editar proyecto"

  @CA-007-3
  Escenario: Sin filtros se devuelve el backlog completo, en cualquier estado y ordenado
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Editar proyecto" con prioridad "should"
    Y una historia "Listar proyectos" con prioridad "must"
    Y la historia "Crear proyecto" asignada al sprint "Sprint 1"
    Y la historia "Crear proyecto" completada
    Cuando se lista el backlog
    Entonces la operación es exitosa
    Y la lista del backlog tiene 3 historias
    Y la lista del backlog muestra las historias en el orden:
      | Crear proyecto   |
      | Listar proyectos |
      | Editar proyecto  |

  # ------------------------------------------------------------------ casos límite
  @CA-007-4 @limite
  Escenario: Un estado sin historias da una lista vacía, no un error
    Dado un proyecto "Demo"
    Y una historia "Crear proyecto" con prioridad "must"
    Y una historia "Editar proyecto" con prioridad "should"
    Cuando se lista el backlog con estado "completado"
    Entonces la operación es exitosa
    Y la lista del backlog tiene 0 historias

  @CA-007-4 @limite
  Escenario: Una prioridad sin historias da una lista vacía, no un error
    Dado un proyecto "Demo"
    Y una historia "Crear proyecto" con prioridad "must"
    Y una historia "Editar proyecto" con prioridad "should"
    Cuando se lista el backlog con prioridad "wont"
    Entonces la operación es exitosa
    Y la lista del backlog tiene 0 historias

  @CA-007-4 @limite
  Escenario: Dos filtros que tienen historias por separado pero no en común dan una lista vacía
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "wont", 5 story points y 2 criterios de aceptación
    Y una historia "Editar proyecto" con prioridad "must"
    Y la historia "Crear proyecto" asignada al sprint "Sprint 1"
    Y la historia "Crear proyecto" completada
    Cuando se lista el backlog con estado "pendiente" y prioridad "wont"
    Entonces la operación es exitosa
    Y la lista del backlog tiene 0 historias

  @CA-007-4 @limite
  Escenario: Un proyecto sin ninguna historia da una lista vacía, con o sin filtros
    Dado un proyecto "Demo"
    Cuando se lista el backlog con estado "pendiente" y prioridad "must"
    Entonces la operación es exitosa
    Y la lista del backlog tiene 0 historias

  @CA-007-4 @limite
  Escenario: En progreso es un estado válido, pero ninguna historia lo tiene hoy
    Dado un proyecto "Demo"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se lista el backlog con estado "en_progreso"
    Entonces la operación es exitosa
    Y la lista del backlog tiene 0 historias

  # ------------------------------------------------------------------ casos de error
  @CA-007-5 @error
  Esquema del escenario: Un estado que no es uno de los tres se rechaza, sin normalizarlo
    Dado un proyecto "Demo"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se lista el backlog con estado "<estado>"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el estado tiene que ser pendiente, en_progreso o completado"

    Ejemplos:
      | estado      |
      | terminado   |
      | Pendiente   |
      | COMPLETADO  |
      | en progreso |

  @CA-007-5 @error
  Esquema del escenario: Una prioridad que no es una de las cuatro se rechaza, sin normalizarla
    Dado un proyecto "Demo"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se lista el backlog con prioridad "<prioridad>"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la prioridad tiene que ser must, should, could o wont"

    Ejemplos:
      | prioridad |
      | urgente   |
      | Must      |
      | MUST      |

  @CA-007-5 @error
  Escenario: Con el estado y la prioridad inválidos a la vez se informan los dos
    Dado un proyecto "Demo"
    Y una historia "Crear proyecto" con prioridad "must"
    Cuando se lista el backlog con estado "terminado" y prioridad "urgente"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el estado tiene que ser pendiente, en_progreso o completado"
    Y el mensaje de error indica "la prioridad tiene que ser must, should, could o wont"

  @CA-007-6 @error
  Escenario: Un proyecto inexistente se rechaza, aunque los filtros también sean inválidos
    Dado ningún proyecto
    Cuando se lista el backlog con estado "terminado" y prioridad "urgente"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"
