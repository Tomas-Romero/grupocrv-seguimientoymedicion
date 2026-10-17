-- +goose Up
-- US-003: el ID de un integrante lo genera la base (RN-003-10), igual que el de los
-- proyectos (US-001) y el de los items del backlog (US-005). Nunca se edita
-- 00001_init.sql: se agrega esta.

ALTER TABLE integrantes ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- +goose Down
ALTER TABLE integrantes ALTER COLUMN id DROP DEFAULT;