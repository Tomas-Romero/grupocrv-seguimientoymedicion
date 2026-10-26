-- +goose Up
-- US-022: defectos. Cada fila es una falla del producto con su severidad, la historia donde se
-- introdujo y el sprint donde se detecto.
--
-- El ID lo genera la base (RN-022-10). El sprint de deteccion es obligatorio y se guarda como
-- ID: su numero ya esta en sprints y no se copia (RN-022-6). sprint_resolucion_id y estado se
-- crean ahora para que US-023 cambie el estado con un UPDATE, sin otra migracion; queda NULL
-- hasta entonces (RN-022-3). estado lleva DEFAULT 'abierto' y no un CHECK con lista de
-- valores: los define US-023, y hasta entonces los protege el dominio.
--
-- Los CHECK de la descripcion y la severidad repiten las reglas del dominio como ultima
-- defensa (RN-022-1 y RN-022-2). Los cuatro ON DELETE CASCADE evitan defectos huerfanos, como
-- en registros_esfuerzo. Que la historia y el sprint sean del proyecto del defecto (RN-022-5)
-- lo verifica el caso de uso, no la base.

CREATE TABLE defectos (
    id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    proyecto_id          UUID        NOT NULL REFERENCES proyectos(id)     ON DELETE CASCADE,
    item_id              UUID        NOT NULL REFERENCES items_backlog(id) ON DELETE CASCADE,
    sprint_deteccion_id  UUID        NOT NULL REFERENCES sprints(id)       ON DELETE CASCADE,
    sprint_resolucion_id UUID                 REFERENCES sprints(id)       ON DELETE CASCADE,
    descripcion          TEXT        NOT NULL CHECK (char_length(btrim(descripcion)) BETWEEN 1 AND 2000),
    severidad            TEXT        NOT NULL CHECK (severidad IN ('critica', 'alta', 'media', 'baja')),
    estado               TEXT        NOT NULL DEFAULT 'abierto',
    creado_en            TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_defectos_proyecto ON defectos(proyecto_id);
CREATE INDEX idx_defectos_item ON defectos(item_id);
CREATE INDEX idx_defectos_sprint_deteccion ON defectos(sprint_deteccion_id);

-- +goose Down
DROP TABLE IF EXISTS defectos;
