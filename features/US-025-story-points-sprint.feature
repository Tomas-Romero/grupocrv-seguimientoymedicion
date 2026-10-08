# language: es
@US-025
Característica: Story Points planificados y completados de un sprint
  Como Agile Enabler
  quiero saber cuántos Story Points se planificaron y cuántos se completaron en un sprint
  para medir el avance del sprint y calcular la velocidad del equipo

  Especificacion: specs/US-025-story-points-sprint.md

  @CA-025-1
  Escenario: Los planificados suman todas las historias del sprint
    Dado el sprint "Sprint 1" con una historia de 5 story points completada
    Y el sprint "Sprint 1" con una historia de 3 story points sin completar
    Y el sprint "Sprint 1" con una historia de 8 story points completada
    Cuando se consultan los story points planificados del sprint "Sprint 1"
    Entonces los story points planificados son 16

  @CA-025-1
  Escenario: Los completados suman solo las historias completadas
    Dado el sprint "Sprint 1" con una historia de 5 story points completada
    Y el sprint "Sprint 1" con una historia de 3 story points sin completar
    Y el sprint "Sprint 1" con una historia de 8 story points completada
    Cuando se consultan los story points completados del sprint "Sprint 1"
    Entonces los story points completados son 13

  @CA-025-2
  Escenario: Con todas las historias completadas, los completados igualan a los planificados
    Dado el sprint "Sprint 1" con una historia de 2 story points completada
    Y el sprint "Sprint 1" con una historia de 5 story points completada
    Cuando se consultan los story points completados del sprint "Sprint 1"
    Entonces los story points completados son 7

  @CA-025-3 @limite
  Escenario: Un sprint sin historias no suma nada y no es un error
    Cuando se consultan los story points planificados del sprint "Sprint 1"
    Entonces la operación es exitosa
    Y los story points planificados son 0

  @limite
  Escenario: Una historia con cero story points no altera ninguna suma
    Dado el sprint "Sprint 1" con una historia de 0 story points completada
    Y el sprint "Sprint 1" con una historia de 0 story points sin completar
    Y el sprint "Sprint 1" con una historia de 3 story points completada
    Cuando se consultan los story points completados del sprint "Sprint 1"
    Entonces los story points completados son 3

  @CA-025-4 @error
  Escenario: Una historia con story points negativos se rechaza
    Dado el sprint "Sprint 1" con una historia de -3 story points sin completar
    Cuando se consultan los story points planificados del sprint "Sprint 1"
    Entonces la operación se rechaza
    Y el mensaje de error indica "no pueden ser negativos"
