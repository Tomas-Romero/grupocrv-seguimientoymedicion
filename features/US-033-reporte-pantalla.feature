# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# En los escenarios "hoy" es el 2026-10-06 (el reloj fijo de ahoraEnEscenarios). El proyecto tiene las fechas por defecto (desde el
# 2026-01-01) y los sprints duran 14 dias seguidos: el Sprint 1 va del 2026-01-01 al 2026-01-14. Las horas se registran con fecha dentro
# de ese rango.
# Frases nuevas de este .feature (se aprueban en el PR, ver docs/diccionario-steps.md):
# `se consulta el reporte del proyecto "P"`, `se consulta el reporte del sprint "Sprint N" del proyecto "P"`,
# `el reporte tiene las secciones:` (con los titulos en una tabla, en orden), `la sección "S" del reporte está vacía`,
# `la sección "S" del reporte tiene N elementos`, `ninguna sección del reporte informa un error`,
# `el reporte del sprint muestra la historia "H" como "Completada"` (o `"No completada"`) y
# `el reporte del sprint lista N defectos detectados y M resueltos`. Despues de consultar el reporte de un sprint, `los story points
# planificados son N`, `los story points completados son N`, `el porcentaje de historias completadas es N`, `las horas estimadas son N`,
# `las horas reales son N` y `la desviación de horas es N` se refieren a ese sprint.
@US-033
Característica: Reporte en pantalla de un proyecto o de un sprint
  Como integrante del equipo
  quiero generar en pantalla el reporte completo de un proyecto o de un sprint
  para revisar todo lo que pasó en un solo lugar y poder imprimirlo

  Especificación: specs/US-033-reporte-pantalla.md

  Antecedentes:
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"

  # ---------------------------------------------------------------- caso normal
  @CA-033-1
  Escenario: El reporte del proyecto tiene sus siete secciones, siempre en el mismo orden
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Login" con prioridad "must", 5 story points y 1 criterio de aceptación
    Y una historia "Perfil" con prioridad "should", 3 story points y 1 criterio de aceptación
    Y la historia "Login" asignada al sprint "Sprint 1"
    Y la historia "Login" completada
    Y el integrante "Ana" ya registró 3 horas de "Desarrollo" en la historia "Login" el 2026-01-05
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Login" detectado en el sprint "Sprint 1"
    Cuando se consulta el reporte del proyecto "Demo"
    Entonces la operación es exitosa
    Y el reporte tiene las secciones:
      | Proyecto        |
      | Integrantes     |
      | Product Backlog |
      | Sprints         |
      | Esfuerzo        |
      | Defectos        |
      | Métricas        |
    Y la sección "Integrantes" del reporte tiene 1 elemento
    Y la sección "Product Backlog" del reporte tiene 2 elementos
    Y la sección "Sprints" del reporte tiene 1 elemento
    Y la sección "Defectos" del reporte tiene 1 elemento
    Y ninguna sección del reporte informa un error

  @CA-033-2
  Escenario: El reporte de un sprint muestra sus historias, sus métricas, su esfuerzo y sus defectos
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Búsqueda" con prioridad "must", 5 story points y 1 criterio de aceptación
    Y una historia "Filtros" con prioridad "should", 3 story points y 1 criterio de aceptación
    Y la historia "Búsqueda" asignada al sprint "Sprint 1"
    Y la historia "Filtros" asignada al sprint "Sprint 1"
    Y la historia "Búsqueda" completada
    Y el integrante "Ana" ya registró 3 horas de "Desarrollo" en la historia "Búsqueda" el 2026-01-05
    Y el integrante "Ana" ya registró 5 horas de "Pruebas" en la historia "Filtros" el 2026-01-06
    Y un defecto de severidad "media" con la descripción "Texto cortado" en la historia "Búsqueda" detectado en el sprint "Sprint 1"
    Cuando se consulta el reporte del sprint "Sprint 1" del proyecto "Demo"
    Entonces la operación es exitosa
    Y el reporte tiene las secciones:
      | Sprint              |
      | Historias           |
      | Métricas            |
      | Esfuerzo registrado |
      | Defectos            |
    Y el reporte del sprint muestra la historia "Búsqueda" como "Completada"
    Y el reporte del sprint muestra la historia "Filtros" como "No completada"
    Y los story points planificados son 8
    Y los story points completados son 5
    Y el porcentaje de historias completadas es 50
    Y las horas estimadas son 12
    Y las horas reales son 8
    Y la desviación de horas es -4
    Y la sección "Esfuerzo registrado" del reporte tiene 2 elementos
    Y el reporte del sprint lista 1 defecto detectado y 0 resueltos

  # ---------------------------------------------------------- casos alternativos
  @CA-033-2 @limite
  Esquema del escenario: Una historia se muestra con el resultado de su fila en cada sprint, no con su estado actual
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Login" con prioridad "must", 5 story points y 1 criterio de aceptación
    Y la historia "Login" asignada al sprint "Sprint 1"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Perfil"
    Y la historia "Login" asignada al sprint "Sprint 2"
    Y la historia "Login" completada
    Y el sprint "Sprint 2" cerrado
    Cuando se consulta el reporte del sprint "<sprint>" del proyecto "Demo"
    Entonces el reporte del sprint muestra la historia "Login" como "<resultado>"

    Ejemplos:
      | sprint   | resultado      |
      | Sprint 1 | No completada  |
      | Sprint 2 | Completada     |

  @CA-033-2 @limite
  Esquema del escenario: Un defecto aparece como detectado en el sprint donde se detectó y como resuelto en el sprint donde se resolvió
    Dado un sprint "Sprint 1" con el objetivo "Uno"
    Y una historia "Login" con prioridad "must", 5 story points y 1 criterio de aceptación
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Dos"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Login" detectado en el sprint "Sprint 2"
    Y el sprint "Sprint 2" cerrado
    Y un sprint "Sprint 3" con el objetivo "Tres"
    Y el sprint "Sprint 3" cerrado
    Y un sprint "Sprint 4" con el objetivo "Cuatro"
    Y el defecto de la historia "Login" resuelto en el sprint "Sprint 4"
    Cuando se consulta el reporte del sprint "<sprint>" del proyecto "Demo"
    Entonces el reporte del sprint lista <detectados> defectos detectados y <resueltos> resueltos

    Ejemplos:
      | sprint   | detectados | resueltos |
      | Sprint 1 | 0          | 0         |
      | Sprint 2 | 1          | 0         |
      | Sprint 3 | 0          | 0         |
      | Sprint 4 | 0          | 1         |

  @CA-033-2 @limite
  Escenario: Un defecto descartado está en la lista del proyecto pero no cuenta en los de ningún sprint
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Login" con prioridad "must", 5 story points y 1 criterio de aceptación
    Y un defecto de severidad "baja" con la descripción "Texto cortado" en la historia "Login" detectado en el sprint "Sprint 1"
    Y el defecto de la historia "Login" en estado "descartado"
    Cuando se consulta el reporte del sprint "Sprint 1" del proyecto "Demo"
    Entonces el reporte del sprint lista 0 defectos detectados y 0 resueltos

  @CA-033-1 @limite
  Escenario: El defecto descartado sí está en la sección de defectos del reporte del proyecto
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Login" con prioridad "must", 5 story points y 1 criterio de aceptación
    Y un defecto de severidad "baja" con la descripción "Texto cortado" en la historia "Login" detectado en el sprint "Sprint 1"
    Y el defecto de la historia "Login" en estado "descartado"
    Cuando se consulta el reporte del proyecto "Demo"
    Entonces la sección "Defectos" del reporte tiene 1 elemento

  # --------------------------------------------------------------------- límites
  @CA-033-3 @limite
  Escenario: Un proyecto recién creado muestra todas sus secciones, vacías y sin error
    Cuando se consulta el reporte del proyecto "Demo"
    Entonces la operación es exitosa
    Y el reporte tiene las secciones:
      | Proyecto        |
      | Integrantes     |
      | Product Backlog |
      | Sprints         |
      | Esfuerzo        |
      | Defectos        |
      | Métricas        |
    Y la sección "Product Backlog" del reporte está vacía
    Y la sección "Sprints" del reporte está vacía
    Y la sección "Esfuerzo" del reporte está vacía
    Y la sección "Defectos" del reporte está vacía
    Y la sección "Métricas" del reporte está vacía
    Y ninguna sección del reporte informa un error

  @CA-033-3 @limite
  Escenario: Un sprint sin historias ni horas muestra todas sus secciones vacías y el porcentaje en cero
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se consulta el reporte del sprint "Sprint 1" del proyecto "Demo"
    Entonces la operación es exitosa
    Y la sección "Historias" del reporte está vacía
    Y la sección "Esfuerzo registrado" del reporte está vacía
    Y la sección "Defectos" del reporte está vacía
    Y el porcentaje de historias completadas es 0
    Y las horas reales son 0
    Y ninguna sección del reporte informa un error

  # ----------------------------------------------------------------------- error
  @CA-033-4 @error
  Escenario: Un proyecto que no existe se rechaza
    Cuando se consulta el reporte del proyecto "Fantasma"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"

  @CA-033-4 @error
  Escenario: Un sprint que no existe se rechaza
    Cuando se consulta el reporte del sprint "Sprint 9" del proyecto "Demo"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint no existe"
