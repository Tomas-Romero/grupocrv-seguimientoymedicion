# language: es
@US-004
Característica: Estado general de un proyecto
  Como integrante del equipo
  quiero consultar el estado general de un proyecto
  para saber de un vistazo si está por empezar, en marcha o terminado, y cuánto del trabajo está hecho

  Especificación: specs/US-004-estado-proyecto.md

  # El reloj de los escenarios es fijo: hoy es el 2026-10-06.

  @CA-004-1
  Escenario: Un proyecto en marcha está en curso y dice cuántos días faltan para el fin
    Dado un proyecto "Demo" con fechas del "2026-01-01" al "2026-12-31"
    Cuando se consulta el estado del proyecto "Demo"
    Entonces la operación es exitosa
    Y el estado del proyecto es "en_curso"
    Y los días del estado son 86

  @CA-004-2
  Escenario: Un proyecto que todavía no empezó está no iniciado y dice en cuántos días empieza
    Dado un proyecto "Demo" con fechas del "2026-10-18" al "2026-12-31"
    Cuando se consulta el estado del proyecto "Demo"
    Entonces el estado del proyecto es "no_iniciado"
    Y los días del estado son 12

  @CA-004-2
  Escenario: Un proyecto que ya terminó está finalizado y dice hace cuántos días
    Dado un proyecto "Demo" con fechas del "2026-08-01" al "2026-10-03"
    Cuando se consulta el estado del proyecto "Demo"
    Entonces el estado del proyecto es "finalizado"
    Y los días del estado son 3

  @CA-004-3
  Escenario: El avance cuenta las historias por estado, los story points y el porcentaje, y muestra el sprint abierto
    Dado un proyecto "Demo" con fechas del "2026-01-01" al "2026-12-31"
    Y una historia "Crear proyecto" con prioridad "must", 5 story points y 2 criterios de aceptación
    Y una historia "Crear sprint" con prioridad "must", 3 story points y 2 criterios de aceptación
    Y una historia "Listar sprints" con prioridad "should", 2 story points y 2 criterios de aceptación
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y la historia "Crear proyecto" asignada al sprint "Sprint 1"
    Y la historia "Crear proyecto" completada
    Cuando se consulta el estado del proyecto "Demo"
    Entonces el proyecto tiene 2 historias pendientes, 0 en progreso y 1 completadas
    Y los story points del proyecto son 10 y los completados 5
    Y el porcentaje de historias completadas del proyecto es 33,3
    Y el sprint abierto del proyecto es "Sprint 1"

  @CA-004-4 @limite
  Escenario: El día de inicio ya está en curso
    Dado un proyecto "Demo" con fechas del "2026-10-06" al "2026-10-20"
    Cuando se consulta el estado del proyecto "Demo"
    Entonces el estado del proyecto es "en_curso"
    Y los días del estado son 14

  @CA-004-4 @limite
  Escenario: El día de fin está en curso y no quedan días
    Dado un proyecto "Demo" con fechas del "2026-09-01" al "2026-10-06"
    Cuando se consulta el estado del proyecto "Demo"
    Entonces el estado del proyecto es "en_curso"
    Y los días del estado son 0

  @CA-004-4 @limite
  Escenario: Un proyecto de un solo día está en curso ese día
    Dado un proyecto "Demo" con fechas del "2026-10-06" al "2026-10-06"
    Cuando se consulta el estado del proyecto "Demo"
    Entonces el estado del proyecto es "en_curso"
    Y los días del estado son 0

  @CA-004-4 @limite
  Escenario: El día siguiente al fin ya está finalizado
    Dado un proyecto "Demo" con fechas del "2026-09-01" al "2026-10-05"
    Cuando se consulta el estado del proyecto "Demo"
    Entonces el estado del proyecto es "finalizado"
    Y los días del estado son 1

  @CA-004-5 @limite
  Escenario: Un proyecto sin historias ni sprints muestra todo en cero y no es un error
    Dado un proyecto "Demo" con fechas del "2026-01-01" al "2026-12-31"
    Cuando se consulta el estado del proyecto "Demo"
    Entonces la operación es exitosa
    Y el proyecto tiene 0 historias pendientes, 0 en progreso y 0 completadas
    Y los story points del proyecto son 0 y los completados 0
    Y el porcentaje de historias completadas del proyecto es 0
    Y el proyecto no tiene sprint abierto

  @CA-004-5 @limite
  Escenario: Con todos los sprints cerrados no hay sprint abierto
    Dado un proyecto "Demo" con fechas del "2026-01-01" al "2026-12-31"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Cuando se consulta el estado del proyecto "Demo"
    Entonces el proyecto no tiene sprint abierto

  @CA-004-6 @error
  Escenario: Un proyecto que no existe se rechaza
    Dado ningún proyecto
    Cuando se consulta el estado del proyecto "Fantasma"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"
