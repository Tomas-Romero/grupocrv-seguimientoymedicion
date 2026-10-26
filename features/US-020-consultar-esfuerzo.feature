# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# En los escenarios "hoy" es el 2026-10-06 (el reloj fijo de ahoraEnEscenarios), asi que el
# esfuerzo se registra en septiembre y los dos sprints son anteriores a hoy.
# Frases nuevas de este .feature (se aprueban en el PR, ver docs/diccionario-steps.md):
# `se consulta el esfuerzo del proyecto "P"` (con `de la historia "H"`, `del integrante "I"` y
# `del sprint "S"`, en ese orden y todos opcionales), `la consulta devuelve N registros`,
# `el total de horas de la consulta es N`, `la consulta devuelve primero el registro de "A"`,
# `las horas de la historia "H" en la consulta son N`,
# `las horas del integrante "I" en la consulta son N`,
# `las horas del sprint "S" en la consulta son N` y
# `las horas fuera de todo sprint en la consulta son N`.
@US-020
Característica: Consultar el esfuerzo registrado por historia, integrante y sprint
  Como integrante del equipo
  quiero consultar el esfuerzo registrado filtrando por historia, por integrante y por sprint
  para saber cuántas horas se dedicaron a cada cosa y quién las dedicó

  Especificación: specs/US-020-consultar-esfuerzo.md

  # El Sprint 1 va del 1 al 14 de septiembre y el Sprint 2, del 16 al 29. El 15 no es de ninguno.
  Antecedentes:
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y el integrante "Beto" con rol "product_builder"
    Y una historia "Alta de proyectos" con prioridad "must"
    Y una historia "Listar proyectos" con prioridad "should"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable" del "2026-09-01" al "2026-09-14"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados" del "2026-09-16" al "2026-09-29"
    Y el integrante "Ana" ya registró 2 horas de "Análisis" en la historia "Alta de proyectos" el 2026-09-01
    Y el integrante "Beto" ya registró 3 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-09-14
    Y el integrante "Beto" ya registró 1 horas de "Revisión" en la historia "Listar proyectos" el 2026-09-15
    Y el integrante "Ana" ya registró 1,5 horas de "Pruebas" en la historia "Listar proyectos" el 2026-09-17

  # ---------------------------------------------------------------- caso normal
  @CA-020-1
  Escenario: Sin filtros se devuelven todos los registros, del más reciente al más antiguo
    Cuando se consulta el esfuerzo del proyecto "Demo"
    Entonces la operación es exitosa
    Y la consulta devuelve 4 registros
    Y la consulta devuelve primero el registro de "Pruebas"
    Y el total de horas de la consulta es 7,5

  @CA-020-2
  Escenario: Filtrar por una historia deja solo sus registros
    Cuando se consulta el esfuerzo del proyecto "Demo" de la historia "Alta de proyectos"
    Entonces la operación es exitosa
    Y la consulta devuelve 2 registros
    Y el total de horas de la consulta es 5
    Y las horas del integrante "Beto" en la consulta son 3
    Y las horas del integrante "Ana" en la consulta son 2

  @CA-020-3
  Escenario: Filtrar por un integrante deja sus registros de todas las historias
    Cuando se consulta el esfuerzo del proyecto "Demo" del integrante "Ana"
    Entonces la operación es exitosa
    Y la consulta devuelve 2 registros
    Y el total de horas de la consulta es 3,5
    Y las horas de la historia "Alta de proyectos" en la consulta son 2
    Y las horas de la historia "Listar proyectos" en la consulta son 1,5

  @CA-020-4
  Escenario: Filtrar por un sprint deja los registros cuya fecha cae en él, con el primer y el último día incluidos
    Cuando se consulta el esfuerzo del proyecto "Demo" del sprint "Sprint 1"
    Entonces la operación es exitosa
    Y la consulta devuelve 2 registros
    Y el total de horas de la consulta es 5

  # ---------------------------------------------------------- casos alternativos
  @CA-020-5
  Escenario: Los totales por historia, por integrante y por sprint suman el total
    Cuando se consulta el esfuerzo del proyecto "Demo"
    Entonces las horas de la historia "Alta de proyectos" en la consulta son 5
    Y las horas de la historia "Listar proyectos" en la consulta son 2,5
    Y las horas del integrante "Ana" en la consulta son 3,5
    Y las horas del integrante "Beto" en la consulta son 4
    Y las horas del sprint "Sprint 1" en la consulta son 5
    Y las horas del sprint "Sprint 2" en la consulta son 1,5
    Y las horas fuera de todo sprint en la consulta son 1

  @CA-020-2
  Escenario: Los filtros se combinan: solo los registros que cumplen todos
    Cuando se consulta el esfuerzo del proyecto "Demo" de la historia "Alta de proyectos" del integrante "Beto" del sprint "Sprint 1"
    Entonces la operación es exitosa
    Y la consulta devuelve 1 registro
    Y la consulta devuelve primero el registro de "Desarrollo"
    Y el total de horas de la consulta es 3

  @CA-020-4
  Escenario: Las horas de una historia se reparten entre los sprints según la fecha del trabajo
    Cuando se consulta el esfuerzo del proyecto "Demo" de la historia "Listar proyectos" del sprint "Sprint 2"
    Entonces la operación es exitosa
    Y la consulta devuelve 1 registro
    Y el total de horas de la consulta es 1,5

  # --------------------------------------------------------------- casos límite
  @CA-020-6 @limite
  Escenario: Un registro cuya fecha no cae en ningún sprint se cuenta aparte
    Cuando se consulta el esfuerzo del proyecto "Demo" de la historia "Listar proyectos"
    Entonces la operación es exitosa
    Y las horas fuera de todo sprint en la consulta son 1
    Y las horas del sprint "Sprint 2" en la consulta son 1,5

  @CA-020-7 @limite
  Escenario: Un proyecto sin registros da una consulta vacía y no es un error
    Dado un proyecto "Vacio"
    Cuando se consulta el esfuerzo del proyecto "Vacio"
    Entonces la operación es exitosa
    Y la consulta devuelve 0 registros
    Y el total de horas de la consulta es 0

  @CA-020-7 @limite
  Escenario: Filtros sin coincidencias dan una consulta vacía y no es un error
    Cuando se consulta el esfuerzo del proyecto "Demo" de la historia "Alta de proyectos" del integrante "Ana" del sprint "Sprint 2"
    Entonces la operación es exitosa
    Y la consulta devuelve 0 registros
    Y el total de horas de la consulta es 0

  # ---------------------------------------------------------------- casos de error
  @CA-020-8 @error
  Escenario: Una historia de otro proyecto como filtro se rechaza
    Dado un proyecto "Otro"
    Y una historia "Ajena" con prioridad "must"
    Cuando se consulta el esfuerzo del proyecto "Demo" de la historia "Ajena"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el item no existe"

  @CA-020-8 @error
  Escenario: Un integrante de otro proyecto como filtro se rechaza
    Dado un proyecto "Otro"
    Y el integrante "Ajeno" con rol "product_builder"
    Cuando se consulta el esfuerzo del proyecto "Demo" del integrante "Ajeno"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el integrante no existe"

  @CA-020-8 @error
  Escenario: Un sprint que no es del proyecto como filtro se rechaza
    Cuando se consulta el esfuerzo del proyecto "Demo" del sprint "Sprint 3"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint no existe"

  @CA-020-9 @error
  Escenario: Un proyecto inexistente se rechaza
    Cuando se consulta el esfuerzo del proyecto "Fantasma"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"
