#!/usr/bin/env bash
# Crea el Environment "azure" del repositorio y lo limita a la rama main (T-018).
#
# Los secretos de un Environment solo los reciben los jobs que corren desde las ramas permitidas. Con
# secretos comunes del repositorio, cualquiera con permiso de escritura podria escribir un workflow en su
# propia rama y leerlos; con el Environment limitado a main no, porque main esta protegida.
#
# Hace falta permiso de administrador del repositorio y `gh` iniciado. No pide ni toca ningun secreto: los
# secretos se cargan despues con scripts/configurar-azure.sh. Se puede correr mas de una vez.
set -euo pipefail

ENTORNO="azure"

echo "Creando (o actualizando) el Environment ${ENTORNO}..."
gh api --method PUT "repos/{owner}/{repo}/environments/${ENTORNO}" \
  --input - > /dev/null <<'JSON'
{"deployment_branch_policy": {"protected_branches": false, "custom_branch_policies": true}}
JSON

existentes=$(gh api "repos/{owner}/{repo}/environments/${ENTORNO}/deployment-branch-policies" --jq '.branch_policies[].name')
if printf '%s\n' "$existentes" | grep -qx "main"; then
  echo "La rama main ya esta permitida."
else
  gh api --method POST "repos/{owner}/{repo}/environments/${ENTORNO}/deployment-branch-policies" \
    -f name=main -f type=branch > /dev/null
  echo "Rama main permitida."
fi

echo
echo "Politica actual del Environment ${ENTORNO}:"
gh api "repos/{owner}/{repo}/environments/${ENTORNO}" --jq '.deployment_branch_policy'
gh api "repos/{owner}/{repo}/environments/${ENTORNO}/deployment-branch-policies" --jq '.branch_policies[] | "  rama permitida: " + .name'
echo
echo "Siguiente paso: cargar las variables y los secretos (se piden sin mostrarlos):"
echo "  bash scripts/configurar-azure.sh"
