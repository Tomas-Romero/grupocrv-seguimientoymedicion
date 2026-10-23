-- +goose Up
-- US-014: las sesiones de Planning Poker y sus participantes.
--
-- Una sesion estima una historia. Nace votando y en la ronda 1; US-016 la revela, US-017
-- abre rondas nuevas y US-018 la finaliza (RN-014-8). El ID lo genera la base
-- (RN-014-12). iniciada_en es informativa: ninguna regla la usa.
--
-- El indice unico parcial es la red de seguridad de RN-014-7: una historia tiene como
-- mucho una sesion abierta, es decir, no finalizada. La regla la hacen cumplir el dominio
-- y el repositorio, con la historia bloqueada.
--
-- Los participantes son los integrantes del proyecto al iniciar (RN-014-4): una fila por
-- participante, que no cambia aunque se sumen integrantes despues.

CREATE TABLE sesiones_planning_poker (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id      UUID        NOT NULL REFERENCES items_backlog(id) ON DELETE CASCADE,
    estado       TEXT        NOT NULL DEFAULT 'votando'
                             CHECK (estado IN ('votando', 'revelada', 'finalizada')),
    ronda        INTEGER     NOT NULL DEFAULT 1 CHECK (ronda >= 1),
    iniciada_en  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX una_sesion_abierta_por_item ON sesiones_planning_poker (item_id) WHERE estado <> 'finalizada';

CREATE TABLE participantes_planning_poker (
    sesion_id      UUID NOT NULL REFERENCES sesiones_planning_poker(id) ON DELETE CASCADE,
    integrante_id  UUID NOT NULL REFERENCES integrantes(id) ON DELETE CASCADE,
    PRIMARY KEY (sesion_id, integrante_id)
);

-- +goose Down
DROP TABLE IF EXISTS participantes_planning_poker;
DROP TABLE IF EXISTS sesiones_planning_poker;
