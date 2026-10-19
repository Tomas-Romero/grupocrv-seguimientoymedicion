-- +goose Up
-- US-010: la fecha en que se marco cada item como completado. Es un instante, como
-- creado_en, y el burndown (US-031) la agrupa por dia. Es nula mientras el item no esta
-- completado: desmarcarlo la vuelve a dejar nula.
--
-- Los items que ya estuvieran completados no tienen una fecha real: se les pone la del
-- momento de la migracion para que cumplan el CHECK.
--
-- El CHECK es la red de seguridad de la spec (seccion 5): solo un item completado tiene
-- fecha, y todo item completado la tiene. Las reglas de marcar y desmarcar las hacen
-- cumplir el dominio y el repositorio.

ALTER TABLE items_backlog ADD COLUMN completado_en TIMESTAMPTZ;

UPDATE items_backlog SET completado_en = now() WHERE estado = 'completado';

ALTER TABLE items_backlog
    ADD CONSTRAINT completado_con_fecha CHECK ((estado = 'completado') = (completado_en IS NOT NULL));

-- +goose Down
ALTER TABLE items_backlog DROP CONSTRAINT IF EXISTS completado_con_fecha;
ALTER TABLE items_backlog DROP COLUMN IF EXISTS completado_en;
