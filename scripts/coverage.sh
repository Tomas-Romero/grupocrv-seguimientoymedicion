#!/usr/bin/env bash
# Corre los tests con cobertura y falla si no se alcanzan los umbrales.
# Uso: ./scripts/coverage.sh    (o `make cover`)
set -euo pipefail

MIN_DOMAIN="${MIN_DOMAIN:-85}"
MIN_TOTAL="${MIN_TOTAL:-70}"

# El umbral total se mide sobre ./internal/..., no sobre ./... : cmd/server es
# cableado de dependencias y no se testea unitariamente, asi que incluirlo solo
# ensuciaria el numero.
echo "==> Cobertura de internal/..."
go test ./internal/... -race -covermode=atomic -coverprofile=coverage.out -count=1 >/dev/null
TOTAL=$(go tool cover -func=coverage.out | awk '/^total:/ {gsub("%","",$3); print $3}')

echo "==> Tests del resto del proyecto (sin medir cobertura)"
go test ./... -race -count=1 >/dev/null

echo "==> Cobertura del dominio (internal/domain/...)"
if go list ./internal/domain/... >/dev/null 2>&1 && [ -n "$(go list ./internal/domain/... 2>/dev/null)" ]; then
  go test ./internal/domain/... -covermode=atomic -coverprofile=domain.out -count=1 >/dev/null
  DOMAIN=$(go tool cover -func=domain.out | awk '/^total:/ {gsub("%","",$3); print $3}')
else
  echo "    (todavia no hay paquetes de dominio; se omite el umbral)"
  DOMAIN="$MIN_DOMAIN"
fi

go tool cover -html=coverage.out -o coverage.html

printf '\n  dominio : %s%%  (minimo %s%%)\n' "$DOMAIN" "$MIN_DOMAIN"
printf '  internal: %s%%  (minimo %s%%)\n\n' "$TOTAL" "$MIN_TOTAL"

FAIL=0
awk -v a="$DOMAIN" -v b="$MIN_DOMAIN" 'BEGIN{exit !(a+0 < b+0)}' && {
  echo "ERROR: la cobertura del dominio ($DOMAIN%) esta por debajo del minimo ($MIN_DOMAIN%)."
  echo "       El dominio son las reglas de negocio: ahi TDD no es opcional."
  FAIL=1
}
awk -v a="$TOTAL" -v b="$MIN_TOTAL" 'BEGIN{exit !(a+0 < b+0)}' && {
  echo "ERROR: la cobertura de internal/ ($TOTAL%) esta por debajo del minimo ($MIN_TOTAL%)."
  FAIL=1
}

[ "$FAIL" -eq 0 ] && echo "Cobertura OK. Reporte navegable en coverage.html"
exit "$FAIL"
