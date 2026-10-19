-- +goose Up
-- US-008: sprints de un proyecto. Cada sprint tiene su Sprint Goal, sus fechas y su
-- estado; el numero es correlativo dentro del proyecto y lo asigna el repositorio
-- (RN-008-11). El nombre ("Sprint N") no se guarda: se arma a partir del numero
-- (RN-008-12), asi no puede quedar distinto de el.
--
-- El ID lo genera la base (RN-008-16). Los CHECK y las restricciones repiten las
-- reglas del dominio como ultima defensa. La duracion, las fechas del proyecto y
-- RN-008-13 (empezar despues del ultimo) no se imponen aca: las hacen cumplir el
-- dominio y el repositorio.

CREATE TABLE sprints (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    proyecto_id   UUID        NOT NULL REFERENCES proyectos(id) ON DELETE CASCADE,
    numero        INTEGER     NOT NULL CHECK (numero > 0),
    sprint_goal   TEXT        NOT NULL CHECK (char_length(btrim(sprint_goal)) BETWEEN 1 AND 200),
    fecha_inicio  DATE        NOT NULL,
    fecha_fin     DATE        NOT NULL,
    estado        TEXT        NOT NULL DEFAULT 'abierto'
                              CHECK (estado IN ('abierto', 'cerrado')),
    creado_en     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fechas_del_sprint_coherentes CHECK (fecha_fin >= fecha_inicio),
    CONSTRAINT numero_de_sprint_unico_por_proyecto UNIQUE (proyecto_id, numero)
);

-- Red de seguridad de RN-008-17: como mucho un sprint abierto por proyecto, y es
-- siempre el ultimo.
CREATE UNIQUE INDEX un_sprint_abierto_por_proyecto ON sprints (proyecto_id) WHERE estado = 'abierto';

-- +goose Down
DROP TABLE IF EXISTS sprints;
