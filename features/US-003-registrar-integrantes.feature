# language: es
# Usa SOLO frases de docs/diccionario-steps.md. Despues de "Dado" no va "que".
# Las dos frases con email son PROPUESTA de US-003 (se aprueban en el PR).
# Lo que el diccionario no puede expresar se cubre con tests unitarios:
#   - nombre de 100 y de 101 caracteres, contados en runas (CL-003-1, CL-003-2)
#   - los tres errores unidos en orden fijo (CL-003-11)
#   - proyecto inexistente con datos invalidos, sin unir errores (CL-003-13)
#   - dos altas simultaneas con el mismo email (CL-003-14)
@US-003
Característica: Registrar integrantes en un proyecto
  Como integrante del equipo
  quiero registrar a las personas que trabajan en un proyecto, con su nombre, email y rol
  para saber quién participa y poder asignarle trabajo y esfuerzo más adelante

  Especificacion: specs/US-003-registrar-integrantes.md

  Antecedentes:
    Dado un proyecto "Demo"

  # ---------------------------------------------------------------- caso normal
  @CA-003-1
  Escenario: Registrar un integrante con datos válidos
    Cuando se registra al integrante "Ana" con rol "product_builder"
    Entonces la operación es exitosa
    Y el proyecto "Demo" tiene 1 integrante

  # ------------------------------------------------------------ caso alternativo
  @CA-003-2
  Escenario: Varios integrantes con roles distintos quedan en el mismo proyecto
    Dado el integrante "Ana" con rol "product_architect"
    Y el integrante "Bruno" con rol "agile_enabler"
    Cuando se registra al integrante "Carla" con rol "product_builder"
    Entonces la operación es exitosa
    Y el proyecto "Demo" tiene 3 integrantes

  @CA-003-2 @limite
  Escenario: Dos integrantes con el mismo nombre y distinto email son válidos
    Dado el integrante "Ana" con email "ana.uno@ejemplo.test" y rol "product_builder"
    Cuando se registra al integrante "Ana" con email "ana.dos@ejemplo.test" y rol "product_builder"
    Entonces la operación es exitosa
    Y el proyecto "Demo" tiene 2 integrantes

  @CA-003-3
  Escenario: El mismo email puede registrarse en dos proyectos distintos
    Dado el integrante "Ana" con email "ana@ejemplo.test" y rol "product_builder"
    Y un proyecto "Otro"
    Cuando se registra al integrante "Ana" con email "ana@ejemplo.test" y rol "product_builder"
    Entonces la operación es exitosa
    Y el proyecto "Otro" tiene 1 integrante
    Y el proyecto "Demo" tiene 1 integrante

  # ----------------------------------------------------------------- caso limite
  @CA-003-4 @limite
  Escenario: Un email con espacios y mayúsculas se acepta
    Cuando se registra al integrante "Tomás" con email " Tomas@Ejemplo.Test " y rol "product_builder"
    Entonces la operación es exitosa
    Y el proyecto "Demo" tiene 1 integrante

  # ----------------------------------------------------------------------- error
  @CA-003-5 @error
  Escenario: Un nombre vacío se rechaza y no se registra nada
    Cuando se registra al integrante "" con rol "product_builder"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el nombre es obligatorio"
    Y el proyecto "Demo" tiene 0 integrantes

  @CA-003-5 @error
  Escenario: Un nombre con solo espacios se trata como vacío
    Cuando se registra al integrante "   " con rol "product_builder"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el nombre es obligatorio"
    Y el proyecto "Demo" tiene 0 integrantes

  @CA-003-6 @error
  Escenario: Un email vacío se rechaza
    Cuando se registra al integrante "Ana" con email "" y rol "product_builder"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el email es obligatorio"
    Y el proyecto "Demo" tiene 0 integrantes

  @CA-003-6 @error
  Escenario: Un email sin formato válido se rechaza
    Cuando se registra al integrante "Ana" con email "sin-arroba" y rol "product_builder"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el email no tiene un formato valido"
    Y el proyecto "Demo" tiene 0 integrantes

  @CA-003-6 @error
  Escenario: Un email con nombre visible se rechaza
    Cuando se registra al integrante "Tomás" con email "Tomas <tomas@ejemplo.test>" y rol "product_builder"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el email no tiene un formato valido"
    Y el proyecto "Demo" tiene 0 integrantes

  @CA-003-7 @error
  Escenario: Un rol fuera de los tres permitidos se rechaza
    Cuando se registra al integrante "Ana" con rol "jefe"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el rol tiene que ser product_architect, agile_enabler o product_builder"
    Y el proyecto "Demo" tiene 0 integrantes

  @CA-003-7 @error
  Escenario: El rol no se normaliza: en mayúsculas se rechaza
    Cuando se registra al integrante "Ana" con rol "PRODUCT_BUILDER"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el rol tiene que ser product_architect, agile_enabler o product_builder"
    Y el proyecto "Demo" tiene 0 integrantes

  @CA-003-8 @error
  Escenario: Un email ya registrado en el proyecto se rechaza
    Dado el integrante "Ana" con email "ana@ejemplo.test" y rol "product_builder"
    Cuando se registra al integrante "Otra Ana" con email "ana@ejemplo.test" y rol "agile_enabler"
    Entonces la operación se rechaza
    Y el mensaje de error indica "ya hay un integrante con ese email en el proyecto"
    Y el proyecto "Demo" tiene 1 integrante

  @CA-003-8 @error
  Escenario: El email repetido se rechaza sin importar mayúsculas ni espacios
    Dado el integrante "Tomás" con email "tomas@ejemplo.test" y rol "product_builder"
    Cuando se registra al integrante "Tomás Otro" con email " TOMAS@ejemplo.test " y rol "agile_enabler"
    Entonces la operación se rechaza
    Y el mensaje de error indica "ya hay un integrante con ese email en el proyecto"
    Y el proyecto "Demo" tiene 1 integrante

  @CA-003-9 @error
  Escenario: Registrar un integrante en un proyecto que no existe se rechaza
    Dado ningún proyecto
    Cuando se registra al integrante "Ana" con rol "product_builder"
    Entonces la operación se rechaza
    Y el mensaje de error indica "el proyecto no existe"
    Y el proyecto "Demo" tiene 0 integrantes