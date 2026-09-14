#!/usr/bin/env bash
# =============================================================================
#  Configuracion completa del repositorio en GitHub — Sprint 0
#
#  Que hace:
#    1. Crea el repositorio y lo sube
#    2. Crea todas las etiquetas del proyecto
#    3. Da acceso de escritura a los dos Product Builders
#    4. Crea el GitHub Project
#
#  La proteccion de la rama main NO se hace aca: va en scripts/proteger-main.sh
#  y se corre al CIERRE del Sprint 0, cuando CI ya pasa en verde.
#
#  Requisitos:
#    - gh CLI instalado y autenticado:  gh auth login
#    - permisos de project:             gh auth refresh -s project,read:project
#
#  Uso:
#    ./scripts/setup-repo.sh
# =============================================================================
set -euo pipefail

# --- CONFIGURACION: ajustar estos valores antes de correr --------------------
OWNER="Tomas-Romero"
REPO="metrics-estimation"
DESC="Software Metrics & Estimation — TP Integrador de Ingenieria y Calidad de Software, UTN FRSR 2026"
BUILDER_1="ConfortiAngelo"
BUILDER_2="JuanVergara-9"
PROYECTO="Metrics & Estimation"
VISIBILIDAD="--public"   # cambiar a --public si la catedra lo pide
# -----------------------------------------------------------------------------

azul()  { printf '\n\033[36m==> %s\033[0m\n' "$1"; }
ok()    { printf '    \033[32m%s\033[0m\n' "$1"; }
aviso() { printf '    \033[33m%s\033[0m\n' "$1"; }

command -v gh >/dev/null || { echo "Falta gh CLI: https://cli.github.com"; exit 1; }
gh auth status >/dev/null 2>&1 || { echo "No estas autenticado. Corre: gh auth login"; exit 1; }

# =============================================================================
azul "1/4  Repositorio"
# =============================================================================
if gh repo view "$OWNER/$REPO" >/dev/null 2>&1; then
  aviso "El repositorio ya existe, se omite la creacion."
else
  gh repo create "$OWNER/$REPO" $VISIBILIDAD --description "$DESC" --source=. --remote=origin --push
  ok "Repositorio creado y subido."
fi

gh repo edit "$OWNER/$REPO" \
  --enable-issues \
  --enable-projects \
  --enable-discussions \
  --delete-branch-on-merge \
  --enable-squash-merge \
  --enable-merge-commit=false \
  --enable-rebase-merge=false
ok "Opciones del repositorio configuradas (solo squash merge, borra la rama al mergear)."

# =============================================================================
azul "2/4  Etiquetas"
# =============================================================================
crear_label() {
  gh label create "$1" --color "$2" --description "$3" --repo "$OWNER/$REPO" --force >/dev/null
  printf '    %s\n' "$1"
}

# tipo de trabajo
crear_label "tipo:historia"  "0E8A16" "Historia de usuario con valor para el usuario"
crear_label "tipo:defecto"   "B60205" "Comportamiento incorrecto del producto"
crear_label "tipo:tecnica"   "5319E7" "Trabajo tecnico sin valor directo de usuario"
crear_label "tipo:spec"      "1D76DB" "Escribir una especificacion SDD"
crear_label "tipo:spike"     "C2E0C6" "Investigacion acotada en tiempo"

# prioridad MoSCoW
crear_label "prio:must"      "D93F0B" "Imprescindible para la entrega"
crear_label "prio:should"    "FBCA04" "Importante pero no bloqueante"
crear_label "prio:could"     "C5DEF5" "Deseable si sobra capacidad"
crear_label "prio:wont"      "EEEEEE" "Fuera del alcance de esta entrega"

# area del codigo
crear_label "area:domain"    "006B75" "internal/domain — reglas de negocio"
crear_label "area:http"      "0052CC" "Handlers y router"
crear_label "area:ui"        "8B5CF6" "Plantillas templ, HTMX y estilos"
crear_label "area:db"        "5B21B6" "Persistencia y migraciones"
crear_label "area:ci"        "4B5563" "Pipeline y automatizacion"
crear_label "area:docs"      "9CA3AF" "Documentacion, specs y actas"

# estado del proceso
crear_label "sdd:pendiente"  "E99695" "Le falta la especificacion SDD"
crear_label "bdd:pendiente"  "F9D0C4" "Le faltan los escenarios BDD"
crear_label "bloqueada"      "000000" "No puede avanzar hasta resolver una dependencia"

# story points
for sp in 1 2 3 5 8 13; do
  crear_label "sp:$sp" "BFD4F2" "Estimada en $sp Story Points"
done
ok "Etiquetas creadas."

# =============================================================================
azul "3/4  Colaboradores"
# =============================================================================
for u in "$BUILDER_1" "$BUILDER_2"; do
  if [[ "$u" == usuario-github-de-* ]]; then
    aviso "Editá BUILDER_1 y BUILDER_2 en este script con los usuarios reales. Se omite '$u'."
    continue
  fi
  gh api -X PUT "repos/$OWNER/$REPO/collaborators/$u" -f permission=push >/dev/null
  ok "$u agregado con permiso de escritura."
done

# =============================================================================
azul "4/4  GitHub Project"
# =============================================================================
if gh project list --owner "$OWNER" --format json 2>/dev/null | grep -q "\"title\":\"$PROYECTO\""; then
  aviso "El proyecto '$PROYECTO' ya existe."
else
  gh project create --owner "$OWNER" --title "$PROYECTO" >/dev/null 2>&1 \
    && ok "Proyecto '$PROYECTO' creado." \
    || aviso "No se pudo crear el proyecto. Corre: gh auth refresh -s project,read:project"
fi

cat <<'FINAL'

  ---------------------------------------------------------------------------
  Lo que queda por hacer A MANO en el Project (la API v2 es incomoda para esto
  y son cinco minutos de clicks):

  Campos personalizados
    Status          (ya existe)  -> Backlog · Ready · Sprint Backlog ·
                                    In Progress · In Review · Done
    Sprint          Iteration    -> duracion 2 semanas, inicio 21/09/2026
                                    (el Sprint 0 se carga como iteracion de 1 semana
                                     desde el 14/09, o se deja sin iteracion)
    Story Points    Number
    Priority        Single select -> Must · Should · Could · Won't
    Epic            Single select -> E1 … E10
    Tipo            Single select -> Historia · Defecto · Tecnica · Spike

  Vistas
    1. "Product Backlog"  Table  · agrupar por Epic     · ordenar por Priority
    2. "Sprint Board"     Board  · columnas por Status  · filtro: iteration:@current
    3. "Burndown"         Insight· Story Points restantes por dia
    4. "Roadmap"          Roadmap· agrupado por Sprint

  Automatizaciones (Project > Workflows): activar
    - Item closed            -> Status: Done
    - Pull request merged    -> Status: Done
    - Item added to project  -> Status: Backlog

  Despues de eso:
    1. ./scripts/crear-issues.sh      carga las 43 historias como issues
    2. Activá "Auto-add to project" en el Project (Workflows) ANTES del paso 1
       si querés que los issues entren solos al tablero.
    3. Al CIERRE del Sprint 0, con make check en verde:
       ./scripts/proteger-main.sh     protege la rama main
  ---------------------------------------------------------------------------

FINAL
