#!/usr/bin/env bash
# =============================================================================
#  Carga el Product Backlog inicial como issues de GitHub.
#
#  Lee scripts/backlog.csv (separado por |) y por cada fila:
#    1. crea el issue con sus etiquetas de tipo, prioridad, area, SP y sprint
#    2. lo agrega al Project
#    3. completa los campos del Project: Story Points, Sprint, Priority, Epic,
#       Tipo y Status
#
#  Es idempotente: si ya existe un issue con el mismo ID en el titulo no lo
#  duplica, pero vuelve a completar sus campos en el Project.
#
#  Precondicion: el Project tiene los campos Status, Sprint (iteration),
#  Story Points, Priority, Epic y Tipo, con las opciones que usa el CSV.
#
#  Uso:
#    ./scripts/crear-issues.sh            # crea los issues
#    ./scripts/crear-issues.sh --dry-run  # solo muestra lo que haria
# =============================================================================
set -euo pipefail

OWNER="Tomas-Romero"
REPO="grupocrv-seguimientoymedicion"
PROYECTO_NUM=3
CSV="$(dirname "$0")/backlog.csv"
HECHOS=" T-001 T-002 T-003 "   # tareas ya terminadas al momento de la carga
DRY_RUN=0
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=1

command -v gh >/dev/null || { echo "Falta gh CLI"; exit 1; }
[[ -f "$CSV" ]] || { echo "No encuentro $CSV"; exit 1; }

declare -A CAMPO OPCION ITER
PROYECTO_ID=""

# Lee una sola vez los IDs del Project: campos, opciones e iteraciones.
# Las iteraciones cuya fecha ya paso figuran como "completadas": se leen igual.
cargar_ids() {
  local tipo clave valor
  while IFS='|' read -r tipo clave valor; do
    case "$tipo" in
      P) PROYECTO_ID="$clave" ;;
      F) CAMPO["$clave"]="$valor" ;;
      O) OPCION["$clave"]="$valor" ;;
      I) ITER["$clave"]="$valor" ;;
    esac
  done < <(gh api graphql \
    -f query='query($o:String!,$n:Int!){user(login:$o){projectV2(number:$n){id fields(first:30){nodes{
      ... on ProjectV2Field{id name}
      ... on ProjectV2SingleSelectField{id name options{id name}}
      ... on ProjectV2IterationField{id name configuration{iterations{id title} completedIterations{id title}}}}}}}}' \
    -f o="$OWNER" -F n="$PROYECTO_NUM" \
    --jq '.data.user.projectV2 | "P|\(.id)|", (.fields.nodes[] | select(.name) | .name as $f
      | "F|\($f)|\(.id)",
        ((.options // [])[] | "O|\($f)/\(.name)|\(.id)"),
        (((.configuration.iterations // []) + (.configuration.completedIterations // []))[]
          | "I|\($f)/\(.title)|\(.id)"))')
  [[ -n "$PROYECTO_ID" ]] || { echo "No pude leer el Project #$PROYECTO_NUM"; exit 1; }
}

fijar_opcion() { # item campo opcion
  local opcion="${OPCION["$2/$3"]:-}"
  [[ -n "$opcion" ]] || { echo "    (falta la opcion '$3' en el campo '$2')"; return 0; }
  gh project item-edit --id "$1" --project-id "$PROYECTO_ID" \
    --field-id "${CAMPO[$2]}" --single-select-option-id "$opcion" >/dev/null
}

if [[ "$DRY_RUN" -eq 0 ]]; then
  cargar_ids
  # etiquetas de sprint
  for n in 0 1 2 3 4; do
    gh label create "sprint:$n" --color "EDEDED" --description "Asignada tentativamente al Sprint $n" \
      --repo "$OWNER/$REPO" --force >/dev/null
  done
fi

CREADOS=0
while IFS='|' read -r id titulo epica sp prioridad sprint area tipo; do
  [[ "$id" == "id" ]] && continue      # cabecera
  [[ -z "${id// }" ]] && continue      # linea vacia

  LABELS="tipo:${tipo},prio:${prioridad},area:${area},sp:${sp},sprint:${sprint}"

  if [[ "$DRY_RUN" -eq 1 ]]; then
    printf '  [dry-run] %-8s %-70.70s  %s\n' "$id" "$titulo" "$LABELS"
    continue
  fi

  URL=$(gh issue list --repo "$OWNER/$REPO" --state all --search "\"[$id]\" in:title" \
    --json url --jq '.[0].url // empty')

  if [[ -z "$URL" ]]; then
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

> Antes de entrar a un Sprint tiene que cumplir el Definition of Ready (ver CONTRIBUTING.md).
BODY
)
    URL=$(gh issue create --repo "$OWNER/$REPO" --title "[$id] $titulo" --body "$CUERPO" --label "$LABELS")
    CREADOS=$((CREADOS+1))
  fi

  # item-add es idempotente: si el issue ya esta en el Project devuelve el mismo item
  ITEM=$(gh project item-add "$PROYECTO_NUM" --owner "$OWNER" --url "$URL" --format json --jq '.id')

  # campos del Project
  gh project item-edit --id "$ITEM" --project-id "$PROYECTO_ID" \
    --field-id "${CAMPO[Story Points]}" --number "$sp" >/dev/null
  gh project item-edit --id "$ITEM" --project-id "$PROYECTO_ID" \
    --field-id "${CAMPO[Sprint]}" --iteration-id "${ITER["Sprint/Sprint $sprint"]}" >/dev/null
  fijar_opcion "$ITEM" "Epic" "$epica"
  case "$prioridad" in
    must) p="Must" ;; should) p="Should" ;; could) p="Could" ;; *) p="Won't" ;;
  esac
  fijar_opcion "$ITEM" "Priority" "$p"
  case "$tipo" in
    historia) t="Historia" ;; defecto) t="Defecto" ;; tecnica) t="Tecnica" ;; *) t="Spike" ;;
  esac
  fijar_opcion "$ITEM" "Tipo" "$t"

  if [[ "$HECHOS" == *" $id "* ]]; then
    fijar_opcion "$ITEM" "Status" "Done"
    gh issue close "$URL" --reason completed >/dev/null 2>&1 || true
  else
    fijar_opcion "$ITEM" "Status" "Backlog"
  fi

  printf '  %-8s %s\n' "$id" "$URL"
  sleep 1   # no pasarse del rate limit
done < <(tr -d '\015' < "$CSV")   # el CSV puede venir con finales de linea de Windows

echo ""
echo "  Listo. Issues creados en esta corrida: $CREADOS"
