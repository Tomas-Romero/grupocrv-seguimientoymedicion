-- +goose Up
-- US-001: el ID de un proyecto lo genera la base (RN-001-6), igual que el de los
-- items del backlog (US-005). Nunca se edita 00001_init.sql: se agrega esta.

ALTER TABLE proyectos ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- +goose Down
ALTER TABLE proyectos ALTER COLUMN id DROP DEFAULT;