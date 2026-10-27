-- +goose Up
-- US-023: cambiar el estado de un defecto y registrar su sprint de resolucion. No crea columnas:
-- estado y sprint_resolucion_id ya las crea la 00012 (que no se edita). Solo agrega los dos CHECK
-- que la 00012 dejo pendientes porque los valores del estado los define esta historia.
--
-- estado_de_defecto_valido repite RN-023-1: el estado es uno de los cuatro, exactos (sin recortar
-- ni pasar a minuscula, igual que el dominio). resolucion_segun_estado repite RN-023-3: hay sprint
-- de resolucion si y solo si el estado es 'resuelto'. Ambas son la ultima defensa (RN-023-12):
-- US-029 cuenta como resuelto todo defecto con sprint de resolucion, y un UPDATE a mano no puede
-- desacoplar el estado del sprint.
--
-- No hay CHECK de texto con funciones que difieran del dominio (como btrim contra TrimSpace en la
-- 00012): aca se compara con igualdad exacta. Lo que no cabe en un CHECK de una fila, como que el
-- sprint de resolucion no sea anterior al de deteccion (RN-023-4, compara numeros de otra tabla) o
-- que sea del mismo proyecto, lo hacen cumplir el dominio y el caso de uso.
--
-- Al aplicarse no hay defectos que las violen: la 00012 solo permite crearlos abiertos y sin sprint
-- de resolucion.

ALTER TABLE defectos
    ADD CONSTRAINT estado_de_defecto_valido
        CHECK (estado IN ('abierto', 'en_correccion', 'resuelto', 'descartado')),
    ADD CONSTRAINT resolucion_segun_estado
        CHECK ((estado = 'resuelto') = (sprint_resolucion_id IS NOT NULL));

-- +goose Down
ALTER TABLE defectos
    DROP CONSTRAINT IF EXISTS resolucion_segun_estado,
    DROP CONSTRAINT IF EXISTS estado_de_defecto_valido;
