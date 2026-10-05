#!/usr/bin/env bash
# Verifica que `make seed` cargue los datos de ejemplo (DEF-003, issue #54).
#
# Crea una base temporal vacia, la migra, corre `make seed` contra ella y cuenta
# las filas. No cuenta sobre la base de desarrollo: esa ya tiene los datos y daria
# verde aunque el seed no funcione. La base temporal se borra al terminar, pase
# lo que pase.
#
# Uso: ./scripts/verificar-seed.sh   (con DATABASE_URL exportada, la de desarrollo)
# Necesita psql, la CLI de goose y make en el PATH, y un usuario con permiso para
# crear bases (el de desarrollo lo tiene).
set -euo pipefail

: "${DATABASE_URL:?defini DATABASE_URL (por ejemplo: set -a; source .env; set +a)}"

for herramienta in psql goose make; do
  if ! command -v "$herramienta" >/dev/null 2>&1; then
    echo "ERROR: falta $herramienta en el PATH."
    exit 2
  fi
done

cd "$(dirname "$0")/.."

# Las llamadas a psql de este script pasan la URL con -d antes de cualquier otra
# opcion: en Windows, psql ignora lo que va despues de un argumento suelto.
sql() {
  psql -X -q -v ON_ERROR_STOP=1 -d "$1" "${@:2}"
}

BASE="verificar_seed_$$_$RANDOM"
# La misma DATABASE_URL apuntando a la base temporal.
URL_TEMPORAL=$(printf '%s' "$DATABASE_URL" | sed -E "s#^(postgres(ql)?://[^/]+)/[^?]*#\1/$BASE#")

borrar() {
  sql "$DATABASE_URL" -c "DROP DATABASE IF EXISTS $BASE WITH (FORCE)" >/dev/null 2>&1 || true
}

echo "==> Creando la base temporal $BASE"
sql "$DATABASE_URL" -c "CREATE DATABASE $BASE"
trap borrar EXIT

echo "==> Migrando con la CLI de goose"
goose -dir migraciones postgres "$URL_TEMPORAL" up

echo "==> make seed"
# La entrada estandar va desde /dev/null: si psql ignora -f (DEF-003), se pondria
# a leer comandos de la terminal y el script quedaria colgado.
DATABASE_URL="$URL_TEMPORAL" make seed </dev/null

echo "==> Datos de ejemplo"
proyectos=$(sql "$URL_TEMPORAL" -tAc "SELECT count(*) FROM proyectos")
integrantes=$(sql "$URL_TEMPORAL" -tAc "SELECT count(*) FROM integrantes")
echo "proyectos: $proyectos (esperado 1)"
echo "integrantes: $integrantes (esperado 3)"

if [ "$proyectos" != "1" ] || [ "$integrantes" != "3" ]; then
  echo
  echo "ERROR: make seed no cargo los datos de ejemplo (DEF-003)."
  exit 1
fi

echo
echo "make seed OK: cargo 1 proyecto y 3 integrantes."
