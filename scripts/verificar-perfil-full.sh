#!/usr/bin/env bash
# Verifica el camino del evaluador: `docker compose --profile full up` sobre un
# volumen vacio tiene que dejar la base migrada y con los datos de ejemplo.
# Es la prueba de DEF-002 (issue #53) y la corre el job "Perfil full" del CI.
#
# Uso: ./scripts/verificar-perfil-full.sh   (necesita Docker y Docker Compose)
#
# Usa un proyecto de compose propio para que el volumen arranque vacio sin tocar
# el de desarrollo. Como docker-compose.yml fija los nombres de los contenedores
# (metrics-db, metrics-app), antes hay que bajar la base de desarrollo con
# `make down`; sus datos quedan en su volumen.
set -euo pipefail

PROYECTO="verificar-perfil-full"

compose() {
  docker compose -p "$PROYECTO" --profile full "$@"
}

consulta() {
  compose exec -T db psql -U metrics -d metrics -v ON_ERROR_STOP=1 "$@"
}

limpiar() {
  compose down -v --remove-orphans >/dev/null 2>&1 || true
}

if docker ps -a --format '{{.Names}}' | grep -qx -e metrics-db -e metrics-app; then
  echo "ERROR: ya existen los contenedores metrics-db o metrics-app."
  echo "       Bajalos con 'make down' (los datos quedan en el volumen) y volve a correr."
  exit 2
fi

trap limpiar EXIT
limpiar # por si quedo algo de una corrida anterior: el volumen tiene que arrancar vacio

echo "==> Levantando el perfil full con un volumen vacio"
if ! compose up -d --build --wait; then
  echo "ERROR: el perfil full no termino de levantar."
  compose logs app
  exit 1
fi

echo "==> Health"
curl -fsS --retry 30 --retry-connrefused --retry-delay 1 http://localhost:8080/health
echo

echo "==> Tablas"
consulta -c '\dt'

FALLAS=0

# verificar TABLA ESPERADO: cuenta las filas de TABLA y las compara con ESPERADO.
verificar() {
  local cantidad
  if ! cantidad=$(consulta -tAc "SELECT count(*) FROM $1" 2>&1); then
    echo "FALLA: no se pudo contar $1: $cantidad"
    FALLAS=1
    return
  fi
  echo "$1: $cantidad (esperado $2)"
  if [ "$cantidad" != "$2" ]; then
    FALLAS=1
  fi
}

echo "==> Datos de ejemplo"
verificar proyectos 1
verificar integrantes 3

if [ "$FALLAS" -ne 0 ]; then
  echo
  echo "ERROR: el perfil full no deja la base migrada con los datos de ejemplo (DEF-002)."
  echo "--- logs de la app"
  compose logs app
  exit 1
fi

echo
echo "Perfil full OK: base migrada y con los datos de ejemplo."
