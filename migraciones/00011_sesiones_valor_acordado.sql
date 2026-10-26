-- +goose Up
-- US-018: el valor que el equipo acordo al finalizar una sesion de Planning Poker. Lo usa
-- el reporte de estimaciones: los Story Points de la historia solos no alcanzan, porque
-- pueden cambiar despues con US-013.
--
-- El CHECK es la red de seguridad de RN-018-11: solo una sesion finalizada tiene valor
-- acordado, y toda sesion finalizada lo tiene. Ninguna sesion se finalizo antes de esta
-- historia, asi que no hay filas que completar. El valor no repite la escala de US-013: la
-- valida el dominio con backlog.NuevosStoryPoints (RN-018-4).

ALTER TABLE sesiones_planning_poker ADD COLUMN valor_acordado INTEGER;

ALTER TABLE sesiones_planning_poker
    ADD CONSTRAINT finalizada_con_valor_acordado CHECK ((estado = 'finalizada') = (valor_acordado IS NOT NULL));

-- +goose Down
ALTER TABLE sesiones_planning_poker DROP CONSTRAINT IF EXISTS finalizada_con_valor_acordado;
ALTER TABLE sesiones_planning_poker DROP COLUMN IF EXISTS valor_acordado;
