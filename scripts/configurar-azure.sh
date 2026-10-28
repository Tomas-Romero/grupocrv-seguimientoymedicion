#!/usr/bin/env bash
# Carga en GitHub todo lo que el despliegue continuo en Azure necesita (T-018): las variables y los secretos del
# Environment "azure". Lo corre quien tiene los datos que entrego la catedra, en su propia terminal.
#
#   bash scripts/configurar-azure.sh
#
# Los valores NO pasan por ningun chat, archivo ni historial: los secretos se piden sin mostrarlos en pantalla y
# se mandan directo a `gh secret set` por la entrada estandar. El script solo imprime los NOMBRES de lo que carga.
# Hace falta permiso de administrador del repositorio y `gh` iniciado. Se puede correr mas de una vez: lo que se
# vuelve a cargar reemplaza al valor anterior. Con ENTER se acepta el valor que sale entre corchetes.
set -euo pipefail

ENTORNO="azure"
CARPETA="$(cd "$(dirname "$0")" && pwd)"

command -v gh > /dev/null || { echo "Falta la CLI de GitHub (gh)."; exit 1; }
gh auth status > /dev/null 2>&1 || { echo "Inicia sesion primero: gh auth login"; exit 1; }
PYTHON=$(command -v python3 || command -v python || true)
[ -n "$PYTHON" ] || { echo "Falta Python (se usa para armar el JSON de AZURE_CREDENTIALS sin errores de comillas)."; exit 1; }

preguntar() { # $1 texto, $2 valor por defecto (opcional); deja la respuesta en $RESPUESTA
  local texto=$1 defecto=${2:-}
  if [ -n "$defecto" ]; then read -r -p "$texto [$defecto]: " RESPUESTA; RESPUESTA=${RESPUESTA:-$defecto}
  else read -r -p "$texto: " RESPUESTA; fi
}
preguntar_oculto() { # $1 texto; deja la respuesta en $RESPUESTA sin mostrarla
  read -r -s -p "$1 (no se ve lo que pegas): " RESPUESTA
  echo
  [ -n "$RESPUESTA" ] || { echo "No puede estar vacio."; exit 1; }
}
variable() { gh variable set "$1" --env "$ENTORNO" --body "$2" > /dev/null && echo "  variable  $1"; }
secreto() { printf '%s' "$2" | gh secret set "$1" --env "$ENTORNO" > /dev/null && echo "  secreto   $1"; }

echo "1/4 Environment ${ENTORNO} (limitado a la rama main)..."
bash "$CARPETA/crear-entorno-azure.sh" > /dev/null
echo "  listo"

echo
echo "2/4 Variables (no son secretas)"
preguntar "Nombre del registro (acr_name)" "ingcalidad2026"; ACR_NAME=$RESPUESTA
preguntar "URL del registro (acrname_url)" "${ACR_NAME}.azurecr.io"; ACR_URL=$RESPUESTA
preguntar "Grupo de recursos (AZURE_RESOURCE_GROUP)" "frsrutn"; GRUPO=$RESPUESTA
preguntar "Entorno de Container Apps (AZURE_CONTAINERAPPS_ENVIRONMENT)"; ENTORNO_APPS=$RESPUESTA
[ -n "$ENTORNO_APPS" ] || { echo "El entorno de Container Apps es obligatorio."; exit 1; }
preguntar "Nombre de la Container App (container_app_name, maximo 32 caracteres, empieza con grupo03-)" "grupo03-grupocrv-seguimiento"; APP=$RESPUESTA
if [ "${#APP}" -gt 32 ] || ! [[ "$APP" =~ ^grupo03-[a-z0-9-]*[a-z0-9]$ ]]; then
  echo "El nombre '$APP' tiene ${#APP} caracteres o un formato que Azure no acepta (2 a 32, minusculas, numeros y guiones, empieza con grupo03-)."
  exit 1
fi
preguntar "Imagen dentro del registro" "grupo03-grupocrv"; IMAGEN=$RESPUESTA
case "$IMAGEN" in grupo03-*) ;; *) echo "La imagen tiene que empezar con grupo03-."; exit 1 ;; esac
read -r -p "Dejar el despliegue automatico APAGADO (CD_ACTIVO=false) hasta que corras la verificacion (recomendado) [S/n]: " APAGADO

variable ACR_NAME "$ACR_NAME"
variable ACR_URL "$ACR_URL"
variable AZURE_RESOURCE_GROUP "$GRUPO"
variable AZURE_CONTAINERAPPS_ENVIRONMENT "$ENTORNO_APPS"
variable CONTAINER_APP_NAME "$APP"
variable IMAGE_NAME "$IMAGEN"
variable CARGAR_DATOS_EJEMPLO "true"
variable MIN_REPLICAS "0"
case "${APAGADO:-S}" in
  [nN]*) variable CD_ACTIVO "true" ;;
  *) variable CD_ACTIVO "false" ;;
esac

echo
echo "3/4 Secretos (se piden sin mostrarlos)"
preguntar_oculto "clientId"; CLIENT_ID=$RESPUESTA
preguntar_oculto "clientSecret"; CLIENT_SECRET=$RESPUESTA
preguntar_oculto "subscriptionId"; SUSCRIPCION=$RESPUESTA
preguntar_oculto "tenantId"; TENANT=$RESPUESTA
JSON=$(CLIENT_ID="$CLIENT_ID" CLIENT_SECRET="$CLIENT_SECRET" SUSCRIPCION="$SUSCRIPCION" TENANT="$TENANT" "$PYTHON" -c '
import json, os
print(json.dumps({"clientId": os.environ["CLIENT_ID"], "clientSecret": os.environ["CLIENT_SECRET"],
                  "subscriptionId": os.environ["SUSCRIPCION"], "tenantId": os.environ["TENANT"]}))')
secreto AZURE_CREDENTIALS "$JSON"
unset CLIENT_ID CLIENT_SECRET SUSCRIPCION TENANT JSON

preguntar_oculto "ACR_PASSWORD (la clave del registro)"; secreto ACR_PASSWORD "$RESPUESTA"

preguntar_oculto "DATABASE_URL (la cadena de conexion de la base de Neon)"
URL=$RESPUESTA
# Neon agrega channel_binding=require y nuestro driver (pgx) lo manda al servidor como una opcion propia: Postgres lo
# rechaza y la app no arranca. Se quita solo, igual que se hizo a mano en docs/ci-cd.md.
if [[ "$URL" == *channel_binding* ]]; then
  URL=$(printf '%s' "$URL" | sed -E 's/([?&])channel_binding=[^&]*&?/\1/; s/[?&]$//')
  echo "  (se quito channel_binding de la cadena)"
fi
[[ "$URL" == *sslmode=* ]] || echo "  AVISO: la cadena no tiene sslmode=require; Neon lo exige."
[[ "$URL" != *-pooler* ]] || echo "  AVISO: la cadena usa el host con pooler; conviene la conexion directa (sin -pooler)."
secreto DATABASE_URL "$URL"
unset URL RESPUESTA

echo
echo "4/4 Lo que quedo cargado en el Environment ${ENTORNO} (solo nombres):"
gh secret list --env "$ENTORNO"
gh variable list --env "$ENTORNO"

echo
echo "Siguiente paso: Actions > Verificar Azure (solo lectura) > Run workflow. No crea ni cambia nada en Azure y"
echo "deja en el resumen que se pudo comprobar. Cuando este todo bien:"
echo "  gh variable set CD_ACTIVO --env ${ENTORNO} --body true    # enciende el despliegue automatico"
echo "  gh workflow run 'CD (Azure)'                               # primer despliegue a mano"
