#!/usr/bin/env bash
# =============================================================================
#  Protege la rama main.
#
#  IMPORTANTE: correr esto al CIERRE del Sprint 0, no al principio.
#  La proteccion exige que los checks de CI esten en verde; si la activas antes
#  de que exista codigo Go, los propios PRs del Sprint 0 quedan bloqueados.
#
#  Precondicion: `make check` pasa en local y el ultimo push a main dio CI verde.
#
#  Uso: ./scripts/proteger-main.sh
# =============================================================================
set -euo pipefail

OWNER="Tomas-Romero"
REPO="metrics-estimation"

command -v gh >/dev/null || { echo "Falta gh CLI"; exit 1; }

echo "==> Protegiendo main en $OWNER/$REPO"

gh api -X PUT "repos/$OWNER/$REPO/branches/main/protection" \
  -H "Accept: application/vnd.github+json" \
  --input - <<'JSON'
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["Lint", "Tests y cobertura", "BDD", "Titulo y rama"]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": true,
    "required_approving_review_count": 1
  },
  "restrictions": null,
  "required_linear_history": true,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "required_conversation_resolution": true
}
JSON

cat <<'FINAL'

  main protegida. Desde ahora:
    - No se puede pushear directo a main.
    - Todo cambio entra por Pull Request con 1 aprobacion.
    - Los 4 checks de CI tienen que estar en verde.
    - Solo squash merge, historial lineal.

  Si dio error 403 o "Upgrade to GitHub Pro": tu cuenta es Free y el repo es
  privado. Opciones:
    a) Pasar el repo a publico:  gh repo edit --visibility public --accept-visibility-change-consequences
    b) Dejarlo sin proteccion y documentarlo en la retro del Sprint 0.
       El flujo de PRs se sigue respetando por acuerdo del equipo; lo que se
       pierde es que GitHub lo haga cumplir solo.

FINAL
