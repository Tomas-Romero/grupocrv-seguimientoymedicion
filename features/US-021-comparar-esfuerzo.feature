# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# En los escenarios "hoy" es el 2026-10-06 (el reloj fijo de ahoraEnEscenarios): el esfuerzo se
# registra en enero, dentro del Sprint 1, que arranca el primer dia del proyecto.
# CA-021-9 (datos invalidos guardados) no tiene escenario: ningun flujo de la aplicacion puede
# guardar horas negativas ni Story Points negativos, asi que lo cubren los tests del dominio.
# Frases nuevas de este .feature (se aprueban en el PR, ver docs/diccionario-steps.md):
# `se compara el esfuerzo del proyecto "P"`, `la comparación tiene N historias`,
# `la primera historia de la comparación es "H"`,
# `la historia "H" tiene N horas estimadas y M reales en la comparación`,
# `la desviación de la historia "H" en la comparación es N`,
# `la desviación porcentual de la historia "H" en la comparación es N` (o `no se puede calcular`),
# `el total comparable tiene N historias` y `las horas sin comparar son N`. Despues de comparar,
# `las horas estimadas son N`, `las horas reales son N`, `la desviación de horas es N` y
# `la desviación porcentual es N` se refieren al total de las historias comparables.
@US-021
Característica: Comparar el esfuerzo estimado contra el esfuerzo real
  Como Agile Enabler
  quiero comparar el esfuerzo estimado de cada historia contra las horas que realmente se trabajaron
  para ver dónde nos equivocamos al estimar y calibrar el próximo Planning

  Especificación: specs/US-021-comparar-esfuerzo.md

  Antecedentes:
    Dado un proyecto "Demo"
    Y el integrante "Ana" con rol "product_builder"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"

  # ---------------------------------------------------------------- caso normal
  @CA-021-1
  Escenario: Una historia completada se compara contra las horas que se registraron
    Dado una historia "Alta de proyectos" con prioridad "must", 5 story points y 1 criterio de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Y el integrante "Ana" ya registró 3 horas de "Análisis" en la historia "Alta de proyectos" el 2026-01-05
    Y el integrante "Ana" ya registró 5 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-01-06
    Cuando se compara el esfuerzo del proyecto "Demo"
    Entonces la operación es exitosa
    Y la comparación tiene 1 historia
    Y la historia "Alta de proyectos" tiene 7,5 horas estimadas y 8 reales en la comparación
    Y la desviación de la historia "Alta de proyectos" en la comparación es 0,5
    Y la desviación porcentual de la historia "Alta de proyectos" en la comparación es 6,7

  @CA-021-2
  Escenario: Se trabajó menos de lo estimado
    Dado una historia "Alta de proyectos" con prioridad "must", 8 story points y 1 criterio de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Y el integrante "Ana" ya registró 2 horas de "Análisis" en la historia "Alta de proyectos" el 2026-01-05
    Y el integrante "Ana" ya registró 3 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-01-06
    Y el integrante "Ana" ya registró 4 horas de "Pruebas" en la historia "Alta de proyectos" el 2026-01-07
    Cuando se compara el esfuerzo del proyecto "Demo"
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" tiene 12 horas estimadas y 9 reales en la comparación
    Y la desviación de la historia "Alta de proyectos" en la comparación es -3
    Y la desviación porcentual de la historia "Alta de proyectos" en la comparación es -25

  # ---------------------------------------------------------- casos alternativos
  @CA-021-3
  Escenario: La historia donde más nos equivocamos sale primero y el total suma las dos
    Dado una historia "Listar proyectos" con prioridad "should", 5 story points y 1 criterio de aceptación
    Y una historia "Alta de proyectos" con prioridad "must", 8 story points y 1 criterio de aceptación
    Y la historia "Listar proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Listar proyectos" completada
    Y la historia "Alta de proyectos" completada
    Y el integrante "Ana" ya registró 8 horas de "Desarrollo" en la historia "Listar proyectos" el 2026-01-05
    Y el integrante "Ana" ya registró 9 horas de "Desarrollo" en la historia "Alta de proyectos" el 2026-01-06
    Cuando se compara el esfuerzo del proyecto "Demo"
    Entonces la operación es exitosa
    Y la comparación tiene 2 historias
    Y la primera historia de la comparación es "Alta de proyectos"
    Y el total comparable tiene 2 historias
    Y las horas estimadas son 19,5
    Y las horas reales son 17
    Y la desviación de horas es -2,5
    Y la desviación porcentual es -12,8

  # --------------------------------------------------------------- casos límite
  @CA-021-4 @limite
  Escenario: Una historia completada y estimada sin horas aparece con la desviación de sus estimadas
    Dado una historia "Alta de proyectos" con prioridad "must", 8 story points y 1 criterio de aceptación
    Y la historia "Alta de proyectos" asignada al sprint "Sprint 1"
    Y la historia "Alta de proyectos" completada
    Cuando se compara el esfuerzo del proyecto "Demo"
    Entonces la operación es exitosa
    Y la historia "Alta de proyectos" tiene 12 horas estimadas y 0 reales en la comparación
    Y la desviación de la historia "Alta de proyectos" en la comparación es -12
    Y la desviación porcentual de la historia "Alta de proyectos" en la comparación es -100

  @CA-021-5 @limite
  Escenario: Una historia sin estimar con horas aparece sin porcentaje y no suma al total
    Dado una historia "Tarea suelta" con prioridad "could"
    Y el integrante "Ana" ya registró 4 horas de "Investigación" en la historia "Tarea suelta" el 2026-01-05
    Cuando se compara el esfuerzo del proyecto "Demo"
    Entonces la operación es exitosa
    Y la comparación tiene 1 historia
    Y la desviación de la historia "Tarea suelta" en la comparación es 4
    Y la desviación porcentual de la historia "Tarea suelta" en la comparación no se puede calcular
    Y el total comparable tiene 0 historias
    Y las horas sin comparar son 4

  @CA-021-6 @limite
  Escenario: Una historia pendiente sin horas no aparece y una en curso con horas no suma al total
    Dado una historia "Sin empezar" con prioridad "should", 3 story points y 1 criterio de aceptación
    Y una historia "En curso" con prioridad "must", 5 story points y 1 criterio de aceptación
    Y el integrante "Ana" ya registró 2 horas de "Desarrollo" en la historia "En curso" el 2026-01-05
    Cuando se compara el esfuerzo del proyecto "Demo"
    Entonces la operación es exitosa
    Y la comparación tiene 1 historia
    Y la primera historia de la comparación es "En curso"
    Y el total comparable tiene 0 historias
    Y las horas sin comparar son 2

  @CA-021-7 @limite
  Escenario: Un proyecto sin historias da una comparación vacía y no es un error
    Cuando se compara el esfuerzo del proyecto "Demo"
    Entonces la operación es exitosa
    Y la comparación tiene 0 historias
    Y las horas estimadas son 0
    Y las horas reales son 0
    Y la desviación porcentual no se puede calcular

  # ---------------------------------------------------------------- casos de error
  @CA-021-8 @error
  Escenario: Un proyecto inexistente se rechaza
    Cuando se compara el esfuerzo del proyecto "Fantasma"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"
