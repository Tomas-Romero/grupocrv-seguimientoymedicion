# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# En los escenarios "hoy" es el 2026-10-06 (el reloj fijo de ahoraEnEscenarios). El proyecto tiene las fechas por defecto
# (desde el 2026-01-01) y los sprints duran 14 dias seguidos: el Sprint 1 va del 2026-01-01 al 2026-01-14. Las horas se
# registran con fecha dentro de ese rango.
# Frases nuevas de este .feature (se aprueban en el PR, ver docs/diccionario-steps.md):
# `se consulta el dashboard del proyecto "P"`, `la tarjeta del sprint es el "Sprint N"`,
# `la velocidad del dashboard es N`, `el dashboard no tiene sprint de referencia`,
# `el gráfico de Story Points tiene N grupos de barras`, `la barra de "completados" del "Sprint N" mide N`
# (o "planificados"), `el dashboard lista N sprints cerrados`,
# `el sprint cerrado "Sprint N" figura con N defectos detectados y M resueltos` y `el dashboard cuenta N defectos`.
# Despues de consultar, `los story points planificados son N`, `los story points completados son N`,
# `el porcentaje de historias completadas es N`, `las horas estimadas son N`, `las horas reales son N` y
# `la desviación de horas es N` se refieren a la tarjeta del sprint de referencia.
@US-030
Característica: Dashboard del proyecto
  Como Agile Enabler
  quiero ver en una sola pantalla las métricas del proyecto y sus gráficos
  para saber cómo viene el sprint y el equipo sin abrir cada pantalla

  Especificación: specs/US-030-dashboard.md

  Antecedentes:
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"

  # ---------------------------------------------------------------- caso normal
  @CA-030-1
  Escenario: La tarjeta del sprint abierto muestra sus Story Points, su porcentaje y sus horas
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Búsqueda" con prioridad "must", 5 story points y 1 criterio de aceptación
    Y una historia "Filtros" con prioridad "should", 3 story points y 1 criterio de aceptación
    Y una historia "Orden" con prioridad "could", 2 story points y 1 criterio de aceptación
    Y la historia "Búsqueda" asignada al sprint "Sprint 1"
    Y la historia "Filtros" asignada al sprint "Sprint 1"
    Y la historia "Orden" asignada al sprint "Sprint 1"
    Y la historia "Búsqueda" completada
    Y el integrante "Ana" ya registró 3 horas de "Desarrollo" en la historia "Búsqueda" el 2026-01-05
    Y el integrante "Ana" ya registró 5 horas de "Pruebas" en la historia "Filtros" el 2026-01-06
    Cuando se consulta el dashboard del proyecto "Demo"
    Entonces la operación es exitosa
    Y la tarjeta del sprint es el "Sprint 1"
    Y los story points planificados son 10
    Y los story points completados son 5
    Y el porcentaje de historias completadas es 33,3
    Y las horas estimadas son 15
    Y las horas reales son 8
    Y la desviación de horas es -7

  @CA-030-2
  Escenario: La velocidad y la tabla de sprints cerrados van del sprint más reciente al más antiguo
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Login" con prioridad "must", 8 story points y 1 criterio de aceptación
    Y la historia "Login" asignada al sprint "Sprint 1"
    Y la historia "Login" completada
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Perfil"
    Y una historia "Perfil" con prioridad "must", 5 story points y 1 criterio de aceptación
    Y la historia "Perfil" asignada al sprint "Sprint 2"
    Y la historia "Perfil" completada
    Y el sprint "Sprint 2" cerrado
    Cuando se consulta el dashboard del proyecto "Demo"
    Entonces la operación es exitosa
    Y la velocidad del dashboard es 6,5
    Y el dashboard lista 2 sprints cerrados
    Y la tarjeta del sprint es el "Sprint 2"

  # ---------------------------------------------------------- casos alternativos
  @CA-030-2
  Escenario: Cada sprint cerrado figura con sus defectos detectados y resueltos
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Login" con prioridad "must", 8 story points y 1 criterio de aceptación
    Y la historia "Login" asignada al sprint "Sprint 1"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Login" detectado en el sprint "Sprint 1"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Arreglos"
    Y el defecto de la historia "Login" resuelto en el sprint "Sprint 2"
    Y el sprint "Sprint 2" cerrado
    Cuando se consulta el dashboard del proyecto "Demo"
    Entonces el sprint cerrado "Sprint 1" figura con 1 defecto detectado y 0 resueltos
    Y el sprint cerrado "Sprint 2" figura con 0 defectos detectados y 1 resuelto

  @CA-030-3
  Escenario: Los defectos se cuentan todos, también los descartados
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Login" con prioridad "must", 8 story points y 1 criterio de aceptación
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Login" detectado en el sprint "Sprint 1"
    Y un defecto de severidad "baja" con la descripción "Texto cortado" en la historia "Login" detectado en el sprint "Sprint 1"
    Y el defecto de la historia "Login" en estado "descartado"
    Cuando se consulta el dashboard del proyecto "Demo"
    Entonces la operación es exitosa
    Y el dashboard cuenta 2 defectos

  @CA-030-3
  Escenario: Un defecto descartado no cuenta en los defectos de su sprint
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Login" con prioridad "must", 8 story points y 1 criterio de aceptación
    Y la historia "Login" asignada al sprint "Sprint 1"
    Y un defecto de severidad "baja" con la descripción "Texto cortado" en la historia "Login" detectado en el sprint "Sprint 1"
    Y el defecto de la historia "Login" en estado "descartado"
    Y el sprint "Sprint 1" cerrado
    Cuando se consulta el dashboard del proyecto "Demo"
    Entonces el sprint cerrado "Sprint 1" figura con 0 defectos detectados y 0 resueltos
    Y el dashboard cuenta 1 defecto

  @CA-030-6
  Escenario: El gráfico de Story Points tiene un grupo por sprint cerrado, con las alturas sobre la barra más alta
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Login" con prioridad "must", 8 story points y 1 criterio de aceptación
    Y una historia "Reporte" con prioridad "should", 2 story points y 1 criterio de aceptación
    Y la historia "Login" asignada al sprint "Sprint 1"
    Y la historia "Reporte" asignada al sprint "Sprint 1"
    Y la historia "Login" completada
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Perfil"
    Y una historia "Perfil" con prioridad "must", 5 story points y 1 criterio de aceptación
    Y la historia "Perfil" asignada al sprint "Sprint 2"
    Y la historia "Perfil" completada
    Y el sprint "Sprint 2" cerrado
    Cuando se consulta el dashboard del proyecto "Demo"
    Entonces el gráfico de Story Points tiene 2 grupos de barras
    Y la barra de "planificados" del "Sprint 1" mide 100
    Y la barra de "completados" del "Sprint 1" mide 80
    Y la barra de "planificados" del "Sprint 2" mide 50
    Y la barra de "completados" del "Sprint 2" mide 50

  # --------------------------------------------------------------------- límites
  @CA-030-4 @CA-030-7 @limite
  Escenario: Un proyecto sin sprints, horas ni defectos muestra todo en cero y sin gráficos
    Cuando se consulta el dashboard del proyecto "Demo"
    Entonces la operación es exitosa
    Y el dashboard no tiene sprint de referencia
    Y la velocidad del dashboard es 0
    Y el dashboard lista 0 sprints cerrados
    Y el gráfico de Story Points tiene 0 grupos de barras
    Y el dashboard cuenta 0 defectos

  @CA-030-1 @limite
  Escenario: Un sprint abierto sin historias ni horas muestra todo en cero
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se consulta el dashboard del proyecto "Demo"
    Entonces la operación es exitosa
    Y la tarjeta del sprint es el "Sprint 1"
    Y los story points planificados son 0
    Y los story points completados son 0
    Y el porcentaje de historias completadas es 0
    Y las horas estimadas son 0
    Y las horas reales son 0

  @CA-030-4 @limite
  Escenario: Con todos los sprints cerrados la tarjeta es la del sprint de número más alto
    Dado un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Perfil"
    Y el sprint "Sprint 2" cerrado
    Cuando se consulta el dashboard del proyecto "Demo"
    Entonces la tarjeta del sprint es el "Sprint 2"

  # ----------------------------------------------------------------------- error
  @CA-030-5 @error
  Escenario: Un proyecto que no existe se rechaza
    Cuando se consulta el dashboard del proyecto "Fantasma"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"
