-- +goose Up
-- Esquema inicial. Se amplia sprint a sprint; nunca se edita una migracion ya
-- mergeada: se agrega una nueva.

CREATE TABLE proyectos (
    id              UUID PRIMARY KEY,
    nombre          TEXT        NOT NULL,
    descripcion     TEXT        NOT NULL DEFAULT '',
    fecha_inicio    DATE        NOT NULL,
    fecha_fin       DATE        NOT NULL,
    creado_en       TIMESTAMPTZ NOT NULL DEFAULT now(),
    actualizado_en  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fechas_coherentes CHECK (fecha_fin >= fecha_inicio),
    CONSTRAINT nombre_no_vacio   CHECK (length(btrim(nombre)) > 0)
);

CREATE TABLE integrantes (
    id           UUID PRIMARY KEY,
    proyecto_id  UUID NOT NULL REFERENCES proyectos(id) ON DELETE CASCADE,
    nombre       TEXT NOT NULL,
    email        TEXT NOT NULL,
    rol          TEXT NOT NULL CHECK (rol IN ('product_architect', 'agile_enabler', 'product_builder')),
    creado_en    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (proyecto_id, email)
);

CREATE INDEX idx_integrantes_proyecto ON integrantes(proyecto_id);

-- +goose Down
DROP TABLE IF EXISTS integrantes;
DROP TABLE IF EXISTS proyectos;
