#!/usr/bin/env bash
# Verifica que todo commit de un PR lleve el ID de su historia, defecto o tarea en el asunto
# (T-014, accion de la retro del Sprint 3).
#
# Con "Rebase and merge" cada commit del PR entra a main tal cual: un commit sin ID queda asi
# para siempre y se pierde la trazabilidad commit -> historia.
#
# Lee de la entrada estandar un commit por linea, con el formato "SHA<TAB>ASUNTO", y sale con
# 1 si algun asunto no lleva [US-NNN], [DEF-NNN] o [T-NNN]. Los commits de merge no se piden:
# quien llama los saca antes (en el CI, con el filtro de la API de GitHub).
#
# Uso en el CI (workflow Convenciones del PR):
#   gh api --paginate repos/OWNER/REPO/pulls/N/commits \
#     --jq '.[] | select((.parents|length)==1) | "\(.sha[0:7])\t\(.commit.message|split("\n")[0])"' \
#     | ./scripts/verificar-ids-commits.sh
set -uo pipefail

PATRON='\[(US|DEF|T)-[0-9]{3}\]'

FALTAN=()
while IFS=$'\t' read -r sha asunto; do
  asunto="${asunto%$'\r'}"
  [[ -z "${sha// }" && -z "${asunto// }" ]] && continue
  # Por si se pasa la lista de git log sin filtrar: un merge no necesita ID.
  [[ "$asunto" == Merge\ * ]] && continue
  if [[ ! "$asunto" =~ $PATRON ]]; then
    FALTAN+=("  $sha  $asunto")
  fi
done

if [[ ${#FALTAN[@]} -gt 0 ]]; then
  echo "Estos commits no llevan el ID de su historia, defecto o tarea en el asunto:"
  echo
  printf '%s\n' "${FALTAN[@]}"
  echo
  echo "Con Rebase and merge cada commit entra a main tal cual, asi que el ID tiene que estar en todos."
  echo "Formato: tipo(alcance): descripcion [US-NNN]   (o [DEF-NNN] o [T-NNN])"
  echo "Ejemplo: docs(bdd): frases de esfuerzo en el diccionario [US-019]"
  echo
  echo "Para corregirlo: git rebase -i <base>, cambiar 'pick' por 'reword' en esos commits y"
  echo "git push --force-with-lease."
  exit 1
fi
echo "Todos los commits llevan el ID."
