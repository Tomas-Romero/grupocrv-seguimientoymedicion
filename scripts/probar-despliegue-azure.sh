#!/usr/bin/env bash
# Prueba el despliegue en Azure (ACI) de punta a punta, sin dejar nada encendido.
#
# Uso: bash scripts/probar-despliegue-azure.sh <region>        (por ejemplo: chilecentral)
#
# Carga la variable ACI_REGION del Environment azure, enciende CD_ACTIVO, corre "CD (Azure)" en main y espera el resultado.
# CD_ACTIVO se vuelve a apagar SIEMPRE al terminar, aunque el despliegue falle o se corte con Ctrl+C. No imprime secretos.
#
# Mientras corre NO hay que mergear nada a main: cada merge dispararia otro despliegue.
# Si falla con RequestDisallowedByAzure, la politica de la suscripcion no permite ACI en esa region (el rechazo ocurre antes de crear
# nada, asi que no gasta creditos): se corre de nuevo con otra. Hace falta tener `gh` iniciado con permiso de escritura en el repo.
set -uo pipefail
export GIT_PAGER=cat

REGION="${1:-}"
[ -n "$REGION" ] || { echo "Uso: bash scripts/probar-despliegue-azure.sh <region>   (por ejemplo: chilecentral)"; exit 1; }
[[ "$REGION" =~ ^[a-z0-9]+$ ]] || { echo "La region es el nombre de Azure en minusculas y sin espacios: $REGION no sirve."; exit 1; }

reintentar() { for i in 1 2 3 4 5; do "$@" && return 0; echo "fallo la red, reintento $i/5..." >&2; sleep 8; done; return 1; }

gh auth status > /dev/null 2>&1 || { echo "Inicia sesion con: gh auth login"; exit 1; }
REPO=$(gh repo view --json nameWithOwner --jq .nameWithOwner) || { echo "No pude saber en que repositorio estoy."; exit 1; }

abiertos=$(gh pr list --state open --json number --jq 'length')
[ "$abiertos" = "0" ] || echo "Aviso: hay $abiertos PR abierto(s); no los mergees hasta que termine la prueba."

apagar() {
  echo
  reintentar gh variable set CD_ACTIVO --env azure --body false > /dev/null && echo "CD_ACTIVO vuelto a apagar (false)."
}
trap apagar EXIT

reintentar gh variable set ACI_REGION --env azure --body "$REGION" > /dev/null && echo "ACI_REGION = $REGION"
reintentar gh variable set CD_ACTIVO --env azure --body true > /dev/null && echo "CD_ACTIVO = true (se apaga solo al terminar)"

previo=$(gh run list --workflow cd-azure.yml --limit 1 --json databaseId --jq '.[0].databaseId // 0')
reintentar gh workflow run cd-azure.yml --ref main
id=""
for _ in $(seq 1 15); do
  sleep 5
  id=$(gh run list --workflow cd-azure.yml --limit 1 --json databaseId --jq '.[0].databaseId // 0')
  [ "$id" != "$previo" ] && [ "$id" != "0" ] && break
  id=""
done
[ -n "$id" ] || { echo "No aparecio la corrida del CD."; exit 1; }
echo "Corrida: $id  (https://github.com/$REPO/actions/runs/$id)"

gh run watch "$id" --exit-status --interval 15 > /dev/null 2>&1
resultado=$(gh run view "$id" --json conclusion --jq .conclusion)
echo "Resultado del CD: $resultado"
gh run view "$id" --json jobs --jq '.jobs[]|"  job \(.name): \(.conclusion)"'
if [ "$resultado" != "success" ]; then
  echo
  echo "--- Lo que fallo (ultimas lineas relevantes):"
  gh run view "$id" --log-failed 2>&1 | grep -iE 'error|disallowed|failed|denied|not found|invalid|location' | sed 's/^[^\t]*\t[^\t]*\t[^ ]* //' | cut -c1-300 | tail -15
  echo
  echo "Si fue RequestDisallowedByAzure, la region no esta permitida: corre de nuevo con otra."
fi
[ "$resultado" = "success" ]
