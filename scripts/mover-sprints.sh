#!/usr/bin/env bash
# =============================================================================
#  Reasigna el campo Sprint del Project segun el plan decidido en el Planning.
#
#  Reglas:
#    - Solo mueve items que NO estan en Done y que ya tienen sprint.
#    - Cada item se busca por su ID entre corchetes en el titulo: "[US-001] ...".
#    - Lee el estado una sola vez y calcula todos los destinos desde esa foto.
#    - Con --dry-run solo muestra lo que haria, sin tocar el Project.
#    - Con --sprint-actual "Sprint 2" tambien acomoda el Status para que el Sprint
#      Board muestre lo correcto: los items del sprint actual pasan a
#      "Sprint Backlog" y el resto a "Backlog". No toca In Progress, In Review ni
#      Done.
#
#  Uso:
#    ./scripts/mover-sprints.sh --dry-run --plan scripts/plan-sprints.txt --sprint-actual "Sprint 2"
#    ./scripts/mover-sprints.sh --plan scripts/plan-sprints.txt --sprint-actual "Sprint 2"
#    ./scripts/mover-sprints.sh "Sprint 1=Sprint 2"          # por nombre de sprint
#
#  --plan ARCHIVO    lineas "ID=Sprint N" o "ID=ninguno" (ver scripts/plan-sprints.txt)
#  Un par "Sprint N=Sprint M" mueve todo lo que esta en N a M (no se encadenan).
#  Las reglas por ID tienen prioridad sobre las reglas por sprint.
# =============================================================================
set -euo pipefail

OWNER="Tomas-Romero"
PROYECTO_NUM=3

DRY_RUN=0
PLAN=""
ACTUAL=""
PARES=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)       DRY_RUN=1 ;;
    --plan)          PLAN="${2:?falta el archivo de --plan}"; shift ;;
    --sprint-actual) ACTUAL="${2:?falta el nombre de --sprint-actual}"; shift ;;
    *)               PARES+=("$1") ;;
  esac
  shift
done

if [[ -z "$PLAN" && ${#PARES[@]} -eq 0 ]]; then
  cat <<'USO'
Uso: ./scripts/mover-sprints.sh [--dry-run] [--plan ARCHIVO] [--sprint-actual "Sprint N"] ["ORIGEN=DESTINO" ...]

Ejemplo:
  ./scripts/mover-sprints.sh --dry-run --plan scripts/plan-sprints.txt --sprint-actual "Sprint 2"
USO
  exit 1
fi

command -v gh >/dev/null || { echo "Falta gh CLI"; exit 1; }

declare -A POR_ITEM POR_SPRINT ITERACION ESTADO
for par in "${PARES[@]}"; do
  [[ "$par" == *=* ]] || { echo "Formato invalido: '$par' (se espera ORIGEN=DESTINO)"; exit 1; }
  POR_SPRINT["${par%%=*}"]="${par#*=}"
done
if [[ -n "$PLAN" ]]; then
  [[ -f "$PLAN" ]] || { echo "No encuentro $PLAN"; exit 1; }
  while IFS= read -r linea; do
    linea="${linea%$'\r'}"
    [[ -z "${linea// }" || "$linea" == \#* ]] && continue
    [[ "$linea" == *=* ]] || { echo "Linea invalida en $PLAN: '$linea'"; exit 1; }
    POR_ITEM["${linea%%=*}"]="${linea#*=}"
  done < "$PLAN"
fi

# IDs del Project: el proyecto, los campos Sprint y Status, y cada iteracion y
# opcion por su nombre. Las iteraciones cuya fecha ya paso figuran como
# "completadas": se leen igual.
PROYECTO_ID=""; CAMPO_SPRINT=""; CAMPO_STATUS=""
while IFS='|' read -r tipo clave valor; do
  case "$tipo" in
    P) PROYECTO_ID="$clave" ;;
    F) CAMPO_SPRINT="$clave" ;;
    S) CAMPO_STATUS="$clave" ;;
    I) ITERACION["$clave"]="$valor" ;;
    O) ESTADO["$clave"]="$valor" ;;
  esac
done < <(gh api graphql \
  -f query='query($o:String!,$n:Int!){user(login:$o){projectV2(number:$n){id
    sprint: field(name:"Sprint"){... on ProjectV2IterationField{id configuration{iterations{id title} completedIterations{id title}}}}
    status: field(name:"Status"){... on ProjectV2SingleSelectField{id options{id name}}}}}}' \
  -f o="$OWNER" -F n="$PROYECTO_NUM" \
  --jq '.data.user.projectV2 | "P|\(.id)|", "F|\(.sprint.id)|", "S|\(.status.id)|",
        ((.sprint.configuration.iterations + .sprint.configuration.completedIterations)[] | "I|\(.title)|\(.id)"),
        (.status.options[] | "O|\(.name)|\(.id)")')

[[ -n "$PROYECTO_ID" && -n "$CAMPO_SPRINT" ]] || { echo "No pude leer el Project #$PROYECTO_NUM"; exit 1; }

validar() { # destino
  [[ "$1" == "ninguno" || -n "${ITERACION[$1]:-}" ]] || { echo "No existe la iteracion destino '$1' en el Project"; exit 1; }
}
for destino in "${POR_ITEM[@]:-}" "${POR_SPRINT[@]:-}"; do
  [[ -n "$destino" ]] && validar "$destino"
done
if [[ -n "$ACTUAL" ]]; then
  validar "$ACTUAL"
  [[ -n "${ESTADO[Backlog]:-}" && -n "${ESTADO[Sprint Backlog]:-}" ]] || { echo "No encuentro los estados Backlog y Sprint Backlog"; exit 1; }
fi

declare -A CUENTA PUNTOS
MOVIDOS=0; REESTADOS=0

# Una sola lectura de los items: id|sprint|status|story points|titulo
while IFS='|' read -r item sprint estado puntos titulo; do
  id_historia=""
  [[ "$titulo" =~ ^\[([^]]+)\] ]] && id_historia="${BASH_REMATCH[1]}"

  destino="$sprint"
  if [[ -n "$id_historia" && -n "${POR_ITEM[$id_historia]:-}" ]]; then
    destino="${POR_ITEM[$id_historia]}"
  elif [[ -n "${POR_SPRINT[$sprint]:-}" ]]; then
    destino="${POR_SPRINT[$sprint]}"
  fi

  # status que corresponde segun el sprint al que queda
  nuevo_estado=""
  if [[ -n "$ACTUAL" && "$estado" != "In Progress" && "$estado" != "In Review" ]]; then
    if [[ "$destino" == "$ACTUAL" ]]; then nuevo_estado="Sprint Backlog"; else nuevo_estado="Backlog"; fi
    [[ "$nuevo_estado" == "$estado" ]] && nuevo_estado=""
  fi

  cambia_sprint=0
  [[ "$destino" != "$sprint" ]] && cambia_sprint=1
  [[ "$cambia_sprint" -eq 0 && -z "$nuevo_estado" ]] && continue

  nota=""
  [[ -n "$nuevo_estado" ]] && nota="  [status: $estado -> $nuevo_estado]"
  printf '  %-9s -> %-9s %3s SP  %s%s\n' "$sprint" "$destino" "$puntos" "${titulo:0:48}" "$nota"
  CUENTA["$destino"]=$(( ${CUENTA["$destino"]:-0} + 1 ))
  PUNTOS["$destino"]=$(( ${PUNTOS["$destino"]:-0} + puntos ))
  [[ "$cambia_sprint" -eq 1 ]] && MOVIDOS=$((MOVIDOS + 1))
  [[ -n "$nuevo_estado" ]] && REESTADOS=$((REESTADOS + 1))

  if [[ "$DRY_RUN" -eq 0 ]]; then
    if [[ "$cambia_sprint" -eq 1 ]]; then
      if [[ "$destino" == "ninguno" ]]; then
        gh project item-edit --id "$item" --project-id "$PROYECTO_ID" --field-id "$CAMPO_SPRINT" --clear >/dev/null
      else
        gh project item-edit --id "$item" --project-id "$PROYECTO_ID" \
          --field-id "$CAMPO_SPRINT" --iteration-id "${ITERACION[$destino]}" >/dev/null
      fi
    fi
    if [[ -n "$nuevo_estado" ]]; then
      gh project item-edit --id "$item" --project-id "$PROYECTO_ID" \
        --field-id "$CAMPO_STATUS" --single-select-option-id "${ESTADO[$nuevo_estado]}" >/dev/null
    fi
  fi
done < <(gh project item-list "$PROYECTO_NUM" --owner "$OWNER" --limit 200 --format json \
  --jq '.items[] | select(.status != "Done") | select(.sprint.title != null)
        | "\(.id)|\(.sprint.title)|\(.status)|\(.["story Points"] // 0)|\(.title)"')

echo
[[ "$DRY_RUN" -eq 1 ]] && echo "  (simulacion: no se modifico nada)"
echo "  Items con sprint nuevo: $MOVIDOS   Items con status nuevo: $REESTADOS"
# Se lee linea por linea: el nombre de una iteracion tiene un espacio ("Sprint 2")
# y un for sobre $(...) lo partiria en dos.
while IFS= read -r destino; do
  [[ -n "$destino" ]] || continue
  printf '    %-9s %2d items  %3d SP\n' "$destino" "${CUENTA[$destino]}" "${PUNTOS[$destino]}"
done < <(printf '%s\n' "${!CUENTA[@]}" | sort)
