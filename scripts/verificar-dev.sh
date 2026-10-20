#!/usr/bin/env bash
# Verifica que `make dev` levante el servidor (DEF-004, issue #55).
#
# El defecto: con `templ generate --watch`, templ solo lanza el --cmd despues de generar
# codigo Go a partir de un .templ. Sin ningun .templ en el repositorio el servidor nunca
# arrancaba y `make dev` se quedaba en "Watching files". Desde T-004 hay plantillas, pero
# este script comprueba que siga funcionando, en particular si alguien las borra.
#
# Corre `make dev` en segundo plano, espera a que /health conteste 200 y lo apaga todo al
# terminar, pase lo que pase. Usa otro puerto que el de desarrollo para no chocar con un
# `make dev` que ya tengas abierto: el proxy de templ sigue apuntando al 8080, pero lo que
# se comprueba es que templ haya lanzado el servidor.
#
# Uso: ./scripts/verificar-dev.sh   (con DATABASE_URL exportada, la de desarrollo)
# Necesita make, templ, curl y la base levantada (make up). La primera vez descarga
# Tailwind, asi que la espera por defecto es larga; se cambia con VERIFICAR_DEV_ESPERA.
set -euo pipefail

: "${DATABASE_URL:?defini DATABASE_URL (por ejemplo: set -a; source .env; set +a)}"

for herramienta in make templ curl; do
  if ! command -v "$herramienta" >/dev/null 2>&1; then
    echo "ERROR: falta $herramienta en el PATH."
    exit 2
  fi
done

cd "$(dirname "$0")/.."

PUERTO="${VERIFICAR_DEV_PUERTO:-18080}"
ESPERA="${VERIFICAR_DEV_ESPERA:-180}"
REGISTRO="$(mktemp)"

# Con job control, el proceso en segundo plano tiene su propio grupo y se puede apagar entero.
set -m
HTTP_ADDR=":$PUERTO" make dev >"$REGISTRO" 2>&1 &
PID=$!
# En Windows (Git Bash) el PID de bash no es el del sistema: taskkill necesita el segundo.
PID_WINDOWS=$(ps -p "$PID" 2>/dev/null | awk 'NR==2{print $4}' || true)

apagar() {
  if command -v taskkill >/dev/null 2>&1 && [ -n "${PID_WINDOWS:-}" ]; then
    taskkill //F //T //PID "$PID_WINDOWS" >/dev/null 2>&1 || true
  else
    kill -- "-$PID" 2>/dev/null || kill "$PID" 2>/dev/null || true
  fi
  rm -f "$REGISTRO"
}
trap apagar EXIT

echo "==> make dev (esperando hasta ${ESPERA}s a http://localhost:${PUERTO}/health)"
for ((segundo = 1; segundo <= ESPERA; segundo++)); do
  if ! kill -0 "$PID" 2>/dev/null; then
    echo "FALLA: make dev termino antes de levantar el servidor. Ultimas lineas:"
    tail -n 15 "$REGISTRO"
    exit 1
  fi
  if curl -fsS -o /dev/null "http://localhost:${PUERTO}/health" 2>/dev/null; then
    echo "OK: make dev levanto el servidor en ${segundo}s."
    exit 0
  fi
  sleep 1
done

echo "FALLA: make dev no levanto el servidor en ${ESPERA}s. Ultimas lineas:"
tail -n 15 "$REGISTRO"
exit 1
