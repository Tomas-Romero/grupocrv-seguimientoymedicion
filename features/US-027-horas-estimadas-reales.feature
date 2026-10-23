# language: es
# Frases nuevas de este .feature (se aprueban en el PR, ver docs/diccionario-steps.md):
# `el sprint "S" con N horas registradas`, `se consultan las horas del sprint "S"`,
# `las horas estimadas son N`, `las horas reales son N`, `la desviación de horas es N`,
# `la desviación porcentual es N` y `la desviación porcentual no se puede calcular`.
@US-027
Característica: Horas estimadas, horas reales y desviación de un sprint
  Como Agile Enabler
  quiero conocer las horas estimadas, las horas reales y la desviación entre ambas de un sprint
  para saber cuánto nos equivocamos al estimar y calibrar el próximo Planning

  Especificacion: specs/US-027-horas-estimadas-reales.md

  # La equivalencia es de 1,5 horas por story point (RN-027-1).

  @CA-027-1
  Escenario: Se trabajó un poco más de lo estimado
    Dado el sprint "Sprint 1" con una historia de 5 story points sin completar
    Y el sprint "Sprint 1" con 3 horas registradas
    Y el sprint "Sprint 1" con 5 horas registradas
    Cuando se consultan las horas del sprint "Sprint 1"
    Entonces las horas estimadas son 7,5
    Y las horas reales son 8
    Y la desviación de horas es 0,5
    Y la desviación porcentual es 6,7

  @CA-027-2
  Escenario: Se trabajó menos de lo estimado
    Dado el sprint "Sprint 1" con una historia de 5 story points completada
    Y el sprint "Sprint 1" con una historia de 3 story points sin completar
    Y el sprint "Sprint 1" con 2 horas registradas
    Y el sprint "Sprint 1" con 3 horas registradas
    Y el sprint "Sprint 1" con 4 horas registradas
    Cuando se consultan las horas del sprint "Sprint 1"
    Entonces las horas estimadas son 12
    Y las horas reales son 9
    Y la desviación de horas es -3
    Y la desviación porcentual es -25

  @CA-027-3 @limite
  Escenario: Sin horas registradas la desviación es la de las estimadas en negativo
    Dado el sprint "Sprint 1" con una historia de 8 story points sin completar
    Cuando se consultan las horas del sprint "Sprint 1"
    Entonces la operación es exitosa
    Y las horas reales son 0
    Y la desviación de horas es -12

  @CA-027-4 @limite
  Escenario: Sin story points planificados no se puede calcular el porcentaje
    Dado el sprint "Sprint 1" con 4 horas registradas
    Cuando se consultan las horas del sprint "Sprint 1"
    Entonces la operación es exitosa
    Y las horas estimadas son 0
    Y la desviación de horas es 4
    Y la desviación porcentual no se puede calcular

  @limite
  Escenario: Las horas de cuarto en cuarto se suman sin ruido
    Dado el sprint "Sprint 1" con una historia de 1 story points sin completar
    Y el sprint "Sprint 1" con 0,25 horas registradas
    Y el sprint "Sprint 1" con 0,25 horas registradas
    Y el sprint "Sprint 1" con 0,5 horas registradas
    Cuando se consultan las horas del sprint "Sprint 1"
    Entonces las horas reales son 1
    Y la desviación de horas es -0,5

  @CA-027-5 @error
  Escenario: Una hora en cero se rechaza indicando su posición
    Dado el sprint "Sprint 1" con una historia de 5 story points sin completar
    Y el sprint "Sprint 1" con 2 horas registradas
    Y el sprint "Sprint 1" con 0 horas registradas
    Cuando se consultan las horas del sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la hora de la posicion 2 no es un numero mayor a 0"

  @CA-027-5 @error
  Escenario: Una hora negativa se rechaza
    Dado el sprint "Sprint 1" con una historia de 5 story points sin completar
    Y el sprint "Sprint 1" con -2 horas registradas
    Cuando se consultan las horas del sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la hora de la posicion 1 no es un numero mayor a 0"

  @CA-027-6 @error
  Escenario: Story points planificados negativos se rechazan
    Dado el sprint "Sprint 1" con una historia de -1 story points sin completar
    Y el sprint "Sprint 1" con 2 horas registradas
    Cuando se consultan las horas del sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "no pueden ser negativos"
