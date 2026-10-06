# language: es
@US-026
Característica: Velocidad del equipo
  Como Agile Enabler
  quiero conocer la velocidad del equipo sobre los sprints cerrados
  para estimar cuánto trabajo cabe en el próximo sprint

  Especificacion: specs/US-026-velocidad-equipo.md

  @CA-026-1
  Escenario: La velocidad es el promedio de los story points completados
    Dado el sprint cerrado "Sprint 1" con 20 story points completados de 25 planificados
    Y el sprint cerrado "Sprint 2" con 30 story points completados de 30 planificados
    Y el sprint cerrado "Sprint 3" con 25 story points completados de 30 planificados
    Cuando se consulta la velocidad del equipo
    Entonces la velocidad del equipo es 25

  @CA-026-2
  Escenario: Con un solo sprint cerrado la velocidad es lo que completó ese sprint
    Dado el sprint cerrado "Sprint 1" con 18 story points completados de 20 planificados
    Cuando se consulta la velocidad del equipo
    Entonces la velocidad del equipo es 18

  @CA-026-3 @limite
  Escenario: Sin sprints cerrados la velocidad es cero, no un error
    Dado ningún sprint cerrado
    Cuando se consulta la velocidad del equipo
    Entonces la operación es exitosa
    Y la velocidad del equipo es 0

  @limite
  Escenario: Sprints cerrados sin nada completado dan velocidad cero
    Dado el sprint cerrado "Sprint 1" con 0 story points completados de 10 planificados
    Y el sprint cerrado "Sprint 2" con 0 story points completados de 12 planificados
    Cuando se consulta la velocidad del equipo
    Entonces la operación es exitosa
    Y la velocidad del equipo es 0

  @limite
  Escenario: El promedio se redondea a un decimal
    Dado el sprint cerrado "Sprint 1" con 30 story points completados de 30 planificados
    Y el sprint cerrado "Sprint 2" con 33 story points completados de 33 planificados
    Cuando se consulta la velocidad del equipo
    Entonces la velocidad del equipo es 31,5

  @CA-026-4 @error
  Escenario: Un sprint con más completados que planificados es un dato corrupto
    Dado el sprint cerrado "Sprint 1" con 8 story points completados de 5 planificados
    Cuando se consulta la velocidad del equipo
    Entonces la operación se rechaza
    Y el mensaje de error indica "mas Story Points completados que planificados"

  @error
  Escenario: Un sprint cerrado sin nombre se rechaza
    Dado el sprint cerrado "" con 5 story points completados de 5 planificados
    Cuando se consulta la velocidad del equipo
    Entonces la operación se rechaza
    Y el mensaje de error indica "necesita un nombre"
