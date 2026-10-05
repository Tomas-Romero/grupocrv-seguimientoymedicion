-- Datos de ejemplo para que el proyecto se pueda evaluar sin cargar nada a mano.
-- Los carga la app al arrancar si CARGAR_DATOS_EJEMPLO=true (solo en el perfil full) y
-- tambien `make seed`. Es idempotente: se puede correr las veces que haga falta.
-- ON CONFLICT DO NOTHING va sin columna para cubrir cualquier restriccion unica, no
-- solo el id: por ejemplo, la de (proyecto_id, email) en integrantes.

INSERT INTO proyectos (id, nombre, descripcion, fecha_inicio, fecha_fin)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Software Metrics & Estimation',
    'TP Integrador de Ingenieria y Calidad de Software — UTN FRSR 2026. Este es nuestro propio proyecto cargado en la aplicacion.',
    '2026-09-14',
    '2026-11-02'
)
ON CONFLICT DO NOTHING;

INSERT INTO integrantes (id, proyecto_id, nombre, email, rol) VALUES
    ('00000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001', 'Tomas Romero', 'tomas@ejemplo.test',   'agile_enabler'),
    ('00000000-0000-0000-0000-000000000012', '00000000-0000-0000-0000-000000000001', 'Conforti',     'conforti@ejemplo.test', 'product_builder'),
    ('00000000-0000-0000-0000-000000000013', '00000000-0000-0000-0000-000000000001', 'Vergara',      'vergara@ejemplo.test',  'product_builder')
ON CONFLICT DO NOTHING;
