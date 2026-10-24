# language: es
@US-028
Característica: Porcentaje de historias completadas
  Como Agile Enabler
  quiero conocer el porcentaje de historias completadas
  para ver de un vistazo qué parte del trabajo planificado se terminó

  Especificación: specs/US-028-porcentaje-historias-completadas.md

  @CA-028-1
  Escenario: El porcentaje es la parte de las historias que están completadas
    Dado el sprint "Sprint 1" con una historia de 5 story points completada
    Y el sprint "Sprint 1" con una historia de 3 story points completada
    Y el sprint "Sprint 1" con una historia de 2 story points completada
    Y el sprint "Sprint 1" con una historia de 8 story points sin completar
    Cuando se consulta el porcentaje de historias completadas del sprint "Sprint 1"
    Entonces la operación es exitosa
    Y el porcentaje de historias completadas es 75

  @CA-028-2
  Escenario: Se cuentan historias y no story points
    Dado el sprint "Sprint 1" con una historia de 13 story points completada
    Y el sprint "Sprint 1" con una historia de 1 story points sin completar
    Y el sprint "Sprint 1" con una historia de 1 story points sin completar
    Y el sprint "Sprint 1" con una historia de 1 story points sin completar
    Cuando se consulta el porcentaje de historias completadas del sprint "Sprint 1"
    Entonces el porcentaje de historias completadas es 25

  @CA-028-3 @limite
  Escenario: Un sprint sin historias da cero y no es un error
    Cuando se consulta el porcentaje de historias completadas del sprint "Sprint 1"
    Entonces la operación es exitosa
    Y el porcentaje de historias completadas es 0

  @CA-028-4 @limite
  Escenario: El porcentaje se redondea a un decimal
    Dado el sprint "Sprint 1" con una historia de 3 story points completada
    Y el sprint "Sprint 1" con una historia de 5 story points completada
    Y el sprint "Sprint 1" con una historia de 2 story points sin completar
    Cuando se consulta el porcentaje de historias completadas del sprint "Sprint 1"
    Entonces el porcentaje de historias completadas es 66,7

  @CA-028-5 @error
  Escenario: Una historia con story points negativos se rechaza
    Dado el sprint "Sprint 1" con una historia de -3 story points sin completar
    Cuando se consulta el porcentaje de historias completadas del sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "no pueden ser negativos"
