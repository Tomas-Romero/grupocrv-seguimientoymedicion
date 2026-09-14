## Qué hace

<!-- Dos o tres lineas. Que cambia desde el punto de vista del usuario. -->

Closes #

## Historia

<!-- Identificador y link a la spec -->

- Historia: `US-NNN`
- Spec SDD: `specs/US-NNN-slug.md`
- Escenarios BDD: `features/US-NNN-slug.feature`

## Evidencia de TDD

<!-- Pega los commits del ciclo. Si no hay RED antes del GREEN, el PR vuelve. -->

```
test(...): RED  ...
feat(...): GREEN ...
refactor(...): ...
```

## Definition of Done

- [ ] `make check` pasa en local (lint + cobertura + BDD)
- [ ] Hay un commit `test(...)` **anterior** al `feat`/`fix` correspondiente
- [ ] Los escenarios BDD de esta historia estan automatizados y en verde
- [ ] La cobertura del dominio sigue arriba del 85%
- [ ] Los casos limite y las condiciones de error de la spec estan cubiertos por tests
- [ ] Los errores se devuelven envueltos con contexto, no se ignoran ni se hace `panic`
- [ ] La spec SDD esta actualizada si la implementacion la contradijo
- [ ] El issue esta enlazado con `Closes #N` y movido en el tablero

## Uso de IA

<!-- Obligatorio. Esta permitido usar IA; no declararlo es lo que no. -->

- [ ] No use IA en este cambio
- [ ] Use IA para: <!-- que parte exactamente -->
- Revisado y entendido por: <!-- tu nombre -->

> Puedo explicar cualquier linea de este PR sin ayuda. Si no podes, no lo abras todavia.

## Como probarlo

1. 
2. 

## Capturas

<!-- Solo si el cambio toca la interfaz -->
