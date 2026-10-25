# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Frases nuevas de este .feature (se aprueban en el PR, ver docs/diccionario-steps.md):
# `un defecto detectado en el sprint N y resuelto en el sprint M`, `un defecto detectado en el sprint N
# sin resolver`, `se consultan los defectos del sprint N`, `los defectos detectados son N` y
# `los defectos resueltos son N`. Los defectos se arman en el contexto, sin pasar por US-022.
@US-029
Característica: Defectos detectados y resueltos por sprint
  Como Agile Enabler
  quiero conocer cuántos defectos se detectaron y cuántos se resolvieron en un sprint
  para ver si la calidad del producto mejora o se deteriora de un sprint a otro

  Especificación: specs/US-029-defectos-por-sprint.md

  # ---------------------------------------------------------------- caso normal
  @CA-029-1
  Escenario: Se cuentan los detectados y los resueltos del sprint
    Dado un defecto detectado en el sprint 3 y resuelto en el sprint 3
    Y un defecto detectado en el sprint 3 y resuelto en el sprint 3
    Y un defecto detectado en el sprint 3 sin resolver
    Cuando se consultan los defectos del sprint 3
    Entonces la operación es exitosa
    Y los defectos detectados son 3
    Y los defectos resueltos son 2

  # ---------------------------------------------------------- casos alternativos
  @CA-029-2
  Esquema del escenario: Un defecto se cuenta como detectado donde se detectó y como resuelto donde se resolvió
    Dado un defecto detectado en el sprint 2 y resuelto en el sprint 4
    Cuando se consultan los defectos del sprint <sprint>
    Entonces los defectos detectados son <detectados>
    Y los defectos resueltos son <resueltos>

    Ejemplos:
      | sprint | detectados | resueltos |
      | 2      | 1          | 0         |
      | 3      | 0          | 0         |
      | 4      | 0          | 1         |

  @CA-029-3
  Escenario: Un defecto detectado y resuelto en el mismo sprint cuenta en las dos cifras
    Dado un defecto detectado en el sprint 5 y resuelto en el sprint 5
    Cuando se consultan los defectos del sprint 5
    Entonces los defectos detectados son 1
    Y los defectos resueltos son 1

  Escenario: Los defectos de otros sprints no cuentan
    Dado un defecto detectado en el sprint 1 y resuelto en el sprint 1
    Y un defecto detectado en el sprint 2 y resuelto en el sprint 2
    Y un defecto detectado en el sprint 3 sin resolver
    Cuando se consultan los defectos del sprint 2
    Entonces los defectos detectados son 1
    Y los defectos resueltos son 1

  # ---------------------------------------------------------------- casos limite
  @CA-029-4 @limite
  Escenario: Un sprint sin defectos da cero y no es un error
    Cuando se consultan los defectos del sprint 1
    Entonces la operación es exitosa
    Y los defectos detectados son 0
    Y los defectos resueltos son 0

  @CA-029-4 @limite
  Escenario: Un defecto sin resolver cuenta como detectado y nunca como resuelto
    Dado un defecto detectado en el sprint 1 sin resolver
    Cuando se consultan los defectos del sprint 1
    Entonces los defectos detectados son 1
    Y los defectos resueltos son 0

  @limite
  Escenario: Un sprint puede resolver más defectos de los que detectó
    Dado un defecto detectado en el sprint 4 sin resolver
    Y un defecto detectado en el sprint 1 y resuelto en el sprint 4
    Y un defecto detectado en el sprint 2 y resuelto en el sprint 4
    Y un defecto detectado en el sprint 3 y resuelto en el sprint 4
    Cuando se consultan los defectos del sprint 4
    Entonces la operación es exitosa
    Y los defectos detectados son 1
    Y los defectos resueltos son 3

  # ------------------------------------------------------------------- errores
  @CA-029-5 @error
  Escenario: Un defecto resuelto antes de detectarse se rechaza
    Dado un defecto detectado en el sprint 3 y resuelto en el sprint 2
    Cuando se consultan los defectos del sprint 3
    Entonces la operación se rechaza
    Y el mensaje de error indica "el defecto se resolvio antes de detectarse"
    Y el mensaje de error indica "defecto en la posicion 1"

  @CA-029-6 @error
  Esquema del escenario: Un número de sprint menor que 1 se rechaza
    Cuando se consultan los defectos del sprint <sprint>
    Entonces la operación se rechaza
    Y el mensaje de error indica "el numero de sprint tiene que ser 1 o mayor"

    Ejemplos:
      | sprint |
      | 0      |
      | -1     |

  @error
  Escenario: Un defecto detectado en un sprint menor que 1 se rechaza
    Dado un defecto detectado en el sprint 0 sin resolver
    Cuando se consultan los defectos del sprint 1
    Entonces la operación se rechaza
    Y el mensaje de error indica "el numero de sprint tiene que ser 1 o mayor"
