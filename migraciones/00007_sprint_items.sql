-- +goose Up
-- US-009: la relacion entre un sprint y los items que se le asignan. Es historica
-- (RN-009-13): tiene una fila por cada vez que un item estuvo en un sprint, con el
-- resultado de ese paso, y ninguna se borra al cerrar el sprint. Asi los Story Points
-- planificados de un sprint cerrado (US-025) salen de todas sus filas, completadas o no.
--
-- La clave es (item_id, sprint_id): no tiene ID propio, asi que no hay nada que generar
-- con DEFAULT gen_random_uuid(). El resultado es 'pendiente' mientras el sprint esta
-- abierto; US-011 lo pasa a 'completado' o 'no_completado' al cerrarlo.
--
-- El indice unico parcial es la red de seguridad de RN-009-6: un item esta como mucho
-- en un sprint abierto, su sprint actual. La Definition of Ready, el estado del sprint y
-- el proyecto no se imponen aca: las hacen cumplir el dominio y el repositorio.

CREATE TABLE sprint_items (
    item_id      UUID        NOT NULL REFERENCES items_backlog(id) ON DELETE CASCADE,
    sprint_id    UUID        NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
    resultado    TEXT        NOT NULL DEFAULT 'pendiente'
                             CHECK (resultado IN ('pendiente', 'completado', 'no_completado')),
    asignado_en  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (item_id, sprint_id)
);

CREATE UNIQUE INDEX un_sprint_pendiente_por_item ON sprint_items (item_id) WHERE resultado = 'pendiente';
CREATE INDEX idx_sprint_items_sprint ON sprint_items (sprint_id);

-- +goose Down
DROP TABLE IF EXISTS sprint_items;
