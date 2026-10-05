# language: es
# T-006: escenario de cableado. Atraviesa runner, steps y servicios para probar
# que la integracion de godog funciona antes de que exista la primera historia.
# No pertenece a ninguna historia de negocio: se borra cuando entre el primer
# escenario real (US-001). Ver docs/adr/0002-integracion-bdd-godog.md.
@T-006
Característica: Cableado de godog
  Como equipo
  quiero un escenario que atraviese runner, steps y servicios
  para confirmar que el BDD funciona de punta a punta antes de la primera historia

  @T-006
  Escenario: Un proyecto se crea y se consulta de punta a punta
    Cuando se crea un proyecto "Demo" con fechas del "2026-09-14" al "2026-11-02"
    Entonces la operación es exitosa
    Y el proyecto "Demo" existe con fechas del "2026-09-14" al "2026-11-02"
