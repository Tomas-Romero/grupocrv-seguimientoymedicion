#!/usr/bin/env bash
# Prueba scripts/verificar-ids-commits.sh con casos armados (T-014): no necesita red ni git.
#
# Uso: ./scripts/test-verificar-ids-commits.sh
set -uo pipefail

cd "$(dirname "$0")/.."
VERIFICADOR=./scripts/verificar-ids-commits.sh

FALLAS=0
TAB=$'\t'

# caso NOMBRE ESPERADO ENTRADA [TEXTO QUE TIENE QUE APARECER EN LA SALIDA]
caso() {
  local nombre="$1" esperado="$2" entrada="$3" texto="${4:-}"
  local salida codigo
  salida=$(printf '%b' "$entrada" | "$VERIFICADOR" 2>&1)
  codigo=$?
  if [[ "$codigo" -ne "$esperado" ]]; then
    echo "FALLA: $nombre: termino con $codigo, se esperaba $esperado"
    echo "$salida" | sed 's/^/    /'
    FALLAS=$((FALLAS + 1))
    return
  fi
  if [[ -n "$texto" && "$salida" != *"$texto"* ]]; then
    echo "FALLA: $nombre: la salida no tiene \"$texto\""
    echo "$salida" | sed 's/^/    /'
    FALLAS=$((FALLAS + 1))
    return
  fi
  echo "ok: $nombre"
}

if [[ ! -x "$VERIFICADOR" ]]; then
  echo "FALLA: no existe $VERIFICADOR (o no es ejecutable)"
  exit 1
fi

caso "todos los commits llevan el ID" 0 \
  "a1b2c3d${TAB}feat(app): GREEN lecturas [T-011]\ne4f5a6b${TAB}test(app): RED lecturas [T-011]\n"

caso "sin commits no hay nada que verificar" 0 ""

caso "un commit sin ID falla y se nombra" 1 \
  "a1b2c3d${TAB}feat(app): GREEN lecturas [T-011]\ne4f5a6b${TAB}docs(bdd): frases al dia\n" \
  "e4f5a6b"

caso "el mensaje de error da un ejemplo de como corregirlo" 1 \
  "e4f5a6b${TAB}docs(bdd): frases al dia\n" \
  "[US-NNN]"

caso "se informan todos los commits que faltan" 1 \
  "1111111${TAB}docs: uno\n2222222${TAB}docs: dos [US-001]\n3333333${TAB}docs: tres\n" \
  "3333333"

caso "los tres tipos de ID valen" 0 \
  "1111111${TAB}feat: historia [US-001]\n2222222${TAB}fix: defecto [DEF-004]\n3333333${TAB}ci: tarea [T-014]\n"

caso "el ID puede ir en cualquier lugar del asunto" 0 \
  "1111111${TAB}feat(app) [US-019]: registrar esfuerzo\n"

caso "las minusculas no valen" 1 "1111111${TAB}feat: algo [us-001]\n"
caso "con menos de tres digitos no vale" 1 "1111111${TAB}feat: algo [US-1]\n"
caso "con otro prefijo no vale" 1 "1111111${TAB}feat: algo [HU-001]\n"
caso "el ID sin corchetes no vale" 1 "1111111${TAB}feat: algo US-001\n"

caso "los commits de merge no se exigen" 0 \
  "1111111${TAB}Merge branch 'main' into feat/US-001-algo\n2222222${TAB}feat: algo [US-001]\n"

caso "los saltos de linea de Windows no molestan" 0 \
  "1111111${TAB}feat: algo [US-001]\r\n"

if [[ "$FALLAS" -ne 0 ]]; then
  echo
  echo "$FALLAS caso(s) fallaron"
  exit 1
fi
echo
echo "Todos los casos pasan."
