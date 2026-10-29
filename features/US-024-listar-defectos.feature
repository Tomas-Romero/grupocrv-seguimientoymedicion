# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Cada defecto se arma sobre una historia distinta, asi la lista se lee por sus historias.
# "Sprint de otro proyecto" y "proyecto inexistente" (CL-024-13 y CL-024-18) no se pueden escribir con estas
# frases, porque el sprint se busca por nombre dentro del proyecto actual: los prueban el caso de uso y la
# pantalla. Que con el sprint inexistente y la severidad invalida a la vez se devuelva SOLO el error del sprint
# (CA-024-6) tampoco se puede afirmar aca (el mensaje se compara por contenido, no por exclusion): lo prueban los
# tests del caso de uso con errors.Is. El escenario de abajo comprueba que el rechazo es por el sprint.
@US-024
Característica: Listar los defectos de un proyecto filtrando por sprint de detección y severidad
  Como integrante del equipo
  quiero listar los defectos de un proyecto filtrando por el sprint en que se detectaron y por su severidad
  para encontrar enseguida los que me interesan sin recorrer la lista entera

  Especificación: specs/US-024-listar-defectos.md

  # ---------------------------------------------------------------- caso normal
  @CA-024-1
  Escenario: Filtrar por un sprint deja solo los defectos detectados en él
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y una historia "Crear proyecto" con prioridad "must"
    Y una historia "Editar proyecto" con prioridad "must"
    Y una historia "Listar proyectos" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y un defecto de severidad "baja" con la descripción "Texto cortado" en la historia "Editar proyecto" detectado en el sprint "Sprint 1"
    Y un defecto de severidad "alta" con la descripción "Orden incorrecto" en la historia "Listar proyectos" detectado en el sprint "Sprint 2"
    Cuando se listan los defectos del sprint "Sprint 1"
    Entonces la operación es exitosa
    Y la lista de defectos tiene 2 defectos
    Y la lista de defectos incluye el defecto de la historia "Crear proyecto"
    Y la lista de defectos incluye el defecto de la historia "Editar proyecto"
    Y la lista de defectos no incluye el defecto de la historia "Listar proyectos"

  @CA-024-1
  Escenario: Filtrar por una severidad deja solo los defectos de esa severidad, de cualquier sprint
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y una historia "Crear proyecto" con prioridad "must"
    Y una historia "Editar proyecto" con prioridad "must"
    Y una historia "Listar proyectos" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y un defecto de severidad "baja" con la descripción "Texto cortado" en la historia "Editar proyecto" detectado en el sprint "Sprint 1"
    Y un defecto de severidad "alta" con la descripción "Orden incorrecto" en la historia "Listar proyectos" detectado en el sprint "Sprint 2"
    Cuando se listan los defectos de severidad "alta"
    Entonces la operación es exitosa
    Y la lista de defectos tiene 2 defectos
    Y la lista de defectos incluye el defecto de la historia "Crear proyecto"
    Y la lista de defectos incluye el defecto de la historia "Listar proyectos"
    Y la lista de defectos no incluye el defecto de la historia "Editar proyecto"

  @CA-024-1
  Escenario: El sprint de resolución no cuenta: el defecto sale con el sprint en que se detectó
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y una historia "Crear proyecto" con prioridad "must"
    Y una historia "Listar proyectos" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y un defecto de severidad "alta" con la descripción "Orden incorrecto" en la historia "Listar proyectos" detectado en el sprint "Sprint 2"
    Y el defecto de la historia "Crear proyecto" resuelto en el sprint "Sprint 2"
    Cuando se listan los defectos del sprint "Sprint 2"
    Entonces la operación es exitosa
    Y la lista de defectos tiene 1 defecto
    Y la lista de defectos incluye el defecto de la historia "Listar proyectos"
    Y la lista de defectos no incluye el defecto de la historia "Crear proyecto"

  # ---------------------------------------------------------- casos alternativos
  @CA-024-2
  Escenario: Los dos filtros se combinan: salen solo los defectos que cumplen ambos
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y una historia "Crear proyecto" con prioridad "must"
    Y una historia "Editar proyecto" con prioridad "must"
    Y una historia "Listar proyectos" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y un defecto de severidad "baja" con la descripción "Texto cortado" en la historia "Editar proyecto" detectado en el sprint "Sprint 1"
    Y un defecto de severidad "alta" con la descripción "Orden incorrecto" en la historia "Listar proyectos" detectado en el sprint "Sprint 2"
    Cuando se listan los defectos del sprint "Sprint 1" y de severidad "alta"
    Entonces la operación es exitosa
    Y la lista de defectos tiene 1 defecto
    Y la lista de defectos incluye el defecto de la historia "Crear proyecto"
    Y la lista de defectos no incluye el defecto de la historia "Editar proyecto"
    Y la lista de defectos no incluye el defecto de la historia "Listar proyectos"

  @CA-024-5
  Escenario: Sin filtros se devuelve la lista completa, en cualquier estado
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y una historia "Crear proyecto" con prioridad "must"
    Y una historia "Editar proyecto" con prioridad "must"
    Y una historia "Listar proyectos" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y un defecto de severidad "baja" con la descripción "Texto cortado" en la historia "Editar proyecto" detectado en el sprint "Sprint 1"
    Y un defecto de severidad "alta" con la descripción "Orden incorrecto" en la historia "Listar proyectos" detectado en el sprint "Sprint 2"
    Y el defecto de la historia "Editar proyecto" en estado "descartado"
    Y el defecto de la historia "Listar proyectos" resuelto en el sprint "Sprint 2"
    Cuando se listan los defectos
    Entonces la operación es exitosa
    Y la lista de defectos tiene 3 defectos
    Y la lista de defectos incluye el defecto de la historia "Crear proyecto"
    Y la lista de defectos incluye el defecto de la historia "Editar proyecto"
    Y la lista de defectos incluye el defecto de la historia "Listar proyectos"

  # ------------------------------------------------------------------ casos límite
  @CA-024-3
  Escenario: Un sprint del proyecto sin defectos da una lista vacía, no un error
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se listan los defectos del sprint "Sprint 2"
    Entonces la operación es exitosa
    Y la lista de defectos tiene 0 defectos

  @CA-024-3
  Escenario: Una severidad sin defectos da una lista vacía, no un error
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se listan los defectos de severidad "critica"
    Entonces la operación es exitosa
    Y la lista de defectos tiene 0 defectos

  @CA-024-3
  Escenario: Dos filtros que tienen defectos por separado pero no en común dan una lista vacía
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y el sprint "Sprint 1" cerrado
    Y un sprint "Sprint 2" con el objetivo "Listados"
    Y una historia "Crear proyecto" con prioridad "must"
    Y una historia "Listar proyectos" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Y un defecto de severidad "critica" con la descripción "Se cierra la sesión" en la historia "Listar proyectos" detectado en el sprint "Sprint 2"
    Cuando se listan los defectos del sprint "Sprint 1" y de severidad "critica"
    Entonces la operación es exitosa
    Y la lista de defectos tiene 0 defectos

  @CA-024-3
  Escenario: Un proyecto sin ningún defecto da una lista vacía, con o sin filtros
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se listan los defectos del sprint "Sprint 1" y de severidad "alta"
    Entonces la operación es exitosa
    Y la lista de defectos tiene 0 defectos

  # ------------------------------------------------------------------ casos de error
  @CA-024-4
  Escenario: Un sprint que no existe en el proyecto se rechaza
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se listan los defectos del sprint "Sprint 9"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint no existe"

  @CA-024-4
  Esquema del escenario: Una severidad que no es una de las cuatro se rechaza, sin normalizarla
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Y una historia "Crear proyecto" con prioridad "must"
    Y un defecto de severidad "alta" con la descripción "Falla al guardar" en la historia "Crear proyecto" detectado en el sprint "Sprint 1"
    Cuando se listan los defectos de severidad "<severidad>"
    Entonces la operación se rechaza
    Y el mensaje de error indica "la severidad tiene que ser critica, alta, media o baja"

    Ejemplos:
      | severidad |
      | urgente   |
      | Alta      |
      | CRITICA   |
      | crítica   |

  @CA-024-6
  Escenario: Con el sprint inexistente y la severidad inválida a la vez, se rechaza por el sprint
    Dado un proyecto "Demo"
    Y un sprint "Sprint 1" con el objetivo "MVP navegable"
    Cuando se listan los defectos del sprint "Sprint 9" y de severidad "urgente"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el sprint no existe"
