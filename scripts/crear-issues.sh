#!/usr/bin/env bash
# =============================================================================
#  Carga el Product Backlog inicial como issues de GitHub.
#
#  Lee scripts/backlog.csv (separado por |) y crea un issue por fila con sus
#  etiquetas de tipo, prioridad, area, story points y sprint.
#
#  Uso:
#    ./scripts/crear-issues.sh            # crea los issues
#    ./scripts/crear-issues.sh --dry-run  # solo muestra lo que haria
#
#  Despues de correrlo: abri el Project y asigna el campo Sprint (iteration)
#  y Story Points a cada item. La etiqueta sprint:N te dice cual va en cada uno.
# =============================================================================
set -euo pipefail

OWNER="Tomas-Romero"
REPO="metrics-estimation"
CSV="$(dirname "$0")/backlog.csv"
PROYECTO="Metrics & Estimation"
DRY_RUN=0
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=1

command -v gh >/dev/null || { echo "Falta gh CLI"; exit 1; }
[[ -f "$CSV" ]] || { echo "No encuentro $CSV"; exit 1; }

# etiquetas de sprint
for n in 0 1 2 3 4; do
  gh label create "sprint:$n" --color "EDEDED" --description "Asignada tentativamente al Sprint $n" \
    --repo "$OWNER/$REPO" --force >/dev/null 2>&1 || true
done

CREADOS=0
while IFS='|' read -r id titulo epica sp prioridad sprint area tipo; do
  [[ "$id" == "id" ]] && continue      # cabecera
  [[ -z "${id// }" ]] && continue      # linea vacia

  LABELS="tipo:${tipo},prio:${prioridad},area:${area},sp:${sp},sprint:${sprint}"

  CUERPO=$(cat <<BODY
**Epica:** ${epica}
**Story Points (estimacion inicial):** ${sp}
**Sprint tentativo:** ${sprint}

---

### Criterios de aceptacion

- [ ] **CA-${id#*-}-1** _(completar en Refinement)_
- [ ] **CA-${id#*-}-2**

### Trazabilidad

| | |
|---|---|
| Spec SDD | \`specs/${id}-slug.md\` |
| Escenarios BDD | \`features/${id}-slug.feature\` |
| Rama | \`feat/${id}-slug\` |

> Issue creado automaticamente desde \`scripts/backlog.csv\` en el Sprint 0.
> Antes de entrar a un Sprint tiene que cumplir el Definition of Ready (ver CONTRIBUTING.md).
BODY
)

  if [[ "$DRY_RUN" -eq 1 ]]; then
    printf '  [dry-run] %-8s %-70.70s  %s\n' "$id" "$titulo" "$LABELS"
  else
    URL=$(gh issue create \
      --repo "$OWNER/$REPO" \
      --title "[$id] $titulo" \
      --body "$CUERPO" \
      --label "$LABELS")
    printf '  %-8s %s\n' "$id" "$URL"
    gh project item-add --owner "$OWNER" --url "$URL" \
      "$(gh project list --owner "$OWNER" --format json | python3 -c "import sys,json;print(next(p['number'] for p in json.load(sys.stdin)['projects'] if p['title']=='$PROYECTO'))" 2>/dev/null)" \
      >/dev/null 2>&1 || true
    CREADOS=$((CREADOS+1))
    sleep 1   # no pasarse del rate limit
  fi
done < "$CSV"

echo ""
echo "  Listo. Issues creados: $CREADOS"
echo "  Ahora entra al Project y carga los campos Sprint y Story Points."
