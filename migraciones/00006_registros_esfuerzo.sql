-- +goose Up
-- US-019: registros de esfuerzo. Cada fila son las horas que un integrante dedico a
-- una historia un dia, en una actividad.
--
-- El ID lo genera la base (RN-019-10). El sprint no se guarda: se deduce de la fecha
-- contra las fechas de los sprints (RN-019-12). Las horas son NUMERIC(4,2) y no un
-- float, para que 0,25 o 23,75 se guarden exactos; el CHECK repite como ultima
-- defensa el rango del dominio (mayor a 0 y hasta 24, RN-019-1).
--
-- Los tres ON DELETE CASCADE evitan horas huerfanas si se borra el proyecto, la
-- historia o el integrante. Que la historia y el integrante sean del proyecto del
-- registro (RN-019-6) lo verifica el caso de uso, no la base.

CREATE TABLE registros_esfuerzo (
    id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    proyecto_id   UUID          NOT NULL REFERENCES proyectos(id)    ON DELETE CASCADE,
    item_id       UUID          NOT NULL REFERENCES items_backlog(id) ON DELETE CASCADE,
    integrante_id UUID          NOT NULL REFERENCES integrantes(id)   ON DELETE CASCADE,
    fecha         DATE          NOT NULL,
    actividad     TEXT          NOT NULL CHECK (char_length(btrim(actividad)) BETWEEN 1 AND 120),
    horas         NUMERIC(4, 2) NOT NULL CHECK (horas > 0 AND horas <= 24),
    creado_en     TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_registros_esfuerzo_item ON registros_esfuerzo(item_id);
CREATE INDEX idx_registros_esfuerzo_integrante ON registros_esfuerzo(integrante_id);

-- +goose Down
DROP TABLE IF EXISTS registros_esfuerzo;
