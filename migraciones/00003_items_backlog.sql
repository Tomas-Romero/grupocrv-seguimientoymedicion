-- +goose Up
-- US-005: items del Product Backlog de un proyecto. La 00002 queda para US-001.
--
-- El ID lo genera la base (RN-005-15) y el numero es correlativo por proyecto
-- (RN-005-10): lo asigna el repositorio y UNIQUE (proyecto_id, numero) queda como
-- red de seguridad. story_points nulo significa "sin estimar", que no es lo mismo
-- que 0 (RN-005-7). Los criterios van en un TEXT[] porque no tienen identidad
-- propia: siempre se leen con su item y se reemplazan enteros al editarlo (US-006).
-- Los CHECK repiten las reglas del dominio como ultima defensa.

CREATE TABLE items_backlog (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    proyecto_id   UUID        NOT NULL REFERENCES proyectos(id) ON DELETE CASCADE,
    numero        INTEGER     NOT NULL CHECK (numero > 0),
    titulo        TEXT        NOT NULL CHECK (char_length(titulo) BETWEEN 1 AND 120),
    descripcion   TEXT        NOT NULL DEFAULT '',
    prioridad     TEXT        NOT NULL CHECK (prioridad IN ('must', 'should', 'could', 'wont')),
    estado        TEXT        NOT NULL DEFAULT 'pendiente'
                              CHECK (estado IN ('pendiente', 'en_progreso', 'completado')),
    story_points  INTEGER     CHECK (story_points >= 0),
    criterios     TEXT[]      NOT NULL DEFAULT '{}',
    creado_en     TIMESTAMPTZ NOT NULL,
    CONSTRAINT numero_unico_por_proyecto UNIQUE (proyecto_id, numero)
);

-- +goose Down
DROP TABLE IF EXISTS items_backlog;
