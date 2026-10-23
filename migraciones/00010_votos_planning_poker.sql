-- +goose Up
-- US-015: los votos de las sesiones de Planning Poker.
--
-- Hay como mucho un voto por sesion, ronda e integrante (RN-015-6): votar otra vez en la
-- misma ronda reemplaza el valor, y una ronda nueva (US-017) empieza sin votos y no toca
-- los anteriores. votado_en es informativa: ninguna regla la usa.
--
-- El valor no repite la escala de US-013 en un CHECK: la valida el dominio con
-- backlog.NuevosStoryPoints (RN-015-4), y asi la escala vive en un solo lugar.
--
-- La clave foranea (sesion_id, integrante_id) hacia los participantes es la red de
-- seguridad de RN-015-2: solo vota quien participa de la sesion.
--
-- Mientras la sesion esta votando, los valores no salen del dominio (RN-015-7): el
-- repositorio los lee para armar la votacion, y la votacion no los expone.

CREATE TABLE votos_planning_poker (
    sesion_id      UUID        NOT NULL,
    ronda          INTEGER     NOT NULL CHECK (ronda >= 1),
    integrante_id  UUID        NOT NULL,
    valor          INTEGER     NOT NULL,
    votado_en      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (sesion_id, ronda, integrante_id),
    FOREIGN KEY (sesion_id, integrante_id)
        REFERENCES participantes_planning_poker (sesion_id, integrante_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS votos_planning_poker;
